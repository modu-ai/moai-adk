//! `moai card land-sync` (SPEC-20261125-land-sync-merge): the leader merges another branch into
//! the integration branch as one merge commit and moves the branch by one compare-and-swap.
//!
//! The order of a call, so that nothing is written until everything that can be known is known
//! (every refusal before step 10 ends with `nothing was changed`):
//!
//! 1. `--by leader`, then the registered leader session ([`batch::leader_reason`]);
//! 2. the source gate: a short local branch name that `work_flow.sync_sources` lists exactly and
//!    that is not the integration branch, and the approved SPEC `--spec` as the integration tip
//!    holds it;
//! 3. the tip is read once and every later step belongs to that tip: the drift check, the guard's
//!    record, and where the source stands against it (a source already inside the tip answers
//!    `already_synced`, and closes a sync that was interrupted after its move);
//! 4. the merge commit is made with git plumbing and no checkout (`merge-tree --write-tree`,
//!    `commit-tree -p <tip> -p <source tip>`), then the worktree that has the branch checked out
//!    is planned for ([`follow::plan`]), then the wide check of the merge runs in a private
//!    worktree and must be a full-selection PASS of the command's own run;
//! 5. the tip is read again (a different one is `stale_base`), and only then the writes: the
//!    ledger row `card.sync_started`, the intent record the guard of the integration branch reads,
//!    git's own compare-and-swap, the removal of the record whatever came of the move, the
//!    worktree follow and the row `card.synced`.
//!
//! A call that stopped after the move is closed by the same call run again: the source is inside
//! the tip then, the newest start row of the branch names the merge the tip stands on and no
//! `card.synced` row follows it, so the worktree follow and the closing row are made then.

use super::batch;
use crate::card::command::open_for_write;
use crate::card::drift::{self, SyncRow};
use crate::card::follow;
use crate::card::guard;
use crate::card::landed::{self, Containment, RefMove};
use crate::ledger::workflow::LEADER_BY;
use crate::ledger::WriteLedger;
use crate::spec_doc;
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// What `card land-sync` was given.
pub struct SyncRequest<'a> {
    /// This session's id, when it has one.
    pub session: Option<&'a str>,
    /// `--by`: who syncs; only the leader word is accepted.
    pub by: &'a str,
    /// The local branch to merge in.
    pub from: &'a str,
    /// `--spec`: the approved SPEC that decided the sync.
    pub spec: &'a str,
    /// The tests' seam for the check program (a debug build reads it from the environment).
    pub check_cmd: Option<&'a Path>,
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card land-sync refused: {why}")
}

/// A refusal that comes before any write.
fn unchanged(why: impl std::fmt::Display) -> String {
    refused(format!("{why}; nothing was changed"))
}

/// What a refusal after the start row says in place of `nothing was changed`.
const START_ROW_KEPT: &str =
    "only the start row of this sync was written to the ledger and the branch did not move";

/// A refusal of the move after the start row was written.
fn after_start_row(why: impl std::fmt::Display) -> String {
    refused(format!("{why}; {START_ROW_KEPT}"))
}

/// A refusal after the branch moved: the branch is changed, so it never says that nothing was.
fn after_move(why: impl std::fmt::Display, branch: &str, merge: &str) -> String {
    refused(format!(
        "{why}; the integration branch {branch} already moved to {merge}: run the same call again \
         after the cause is repaired, and it closes the sync"
    ))
}

// ---- the source gate -----------------------------------------------------------------------

/// The branches `work_flow.sync_sources` of the project profile lists (empty when the key is
/// absent).
fn listed_sources(project: &Path) -> Result<Vec<String>> {
    let (profile, _) = crate::profile::read_profile(project)
        .map_err(|why| format!("the project profile cannot be read ({why})"))?;
    Ok(profile
        .work_flow
        .map(|w| w.sync_sources().to_vec())
        .unwrap_or_default())
}

/// Why `from` is no source of a sync into `integration`, or `None`: a short and valid branch name,
/// not the integration branch, and listed exactly in `work_flow.sync_sources`.
fn source_refusal(from: &str, integration: &str, listed: &[String]) -> Option<String> {
    if from.starts_with("refs/") {
        return Some(format!(
            "`{from}` is not a short local branch name: give the branch name without `refs/`"
        ));
    }
    if !crate::profile::branch_name(from) {
        return Some(format!("`{from}` is not a valid branch name"));
    }
    let integration = integration
        .strip_prefix("refs/heads/")
        .unwrap_or(integration);
    if from == integration {
        return Some(format!(
            "`{from}` is the integration branch: a sync merges another branch into it"
        ));
    }
    if !listed.iter().any(|name| name == from) {
        return Some(if listed.is_empty() {
            format!(
                "`{from}` is no source: `work_flow.sync_sources` of the project profile is empty \
                 or missing, so no branch can be synced"
            )
        } else {
            format!("`{from}` is not listed in `work_flow.sync_sources` of the project profile")
        });
    }
    None
}

/// Why the SPEC `id`, as the text `text` holds it, is not an approved SPEC, or `None`: the rule of
/// `card take` and `card next` (`spec_link::status`): this generation's id, no error finding, an
/// approved or completed status and a recorded approval.
fn spec_problem(id: &str, text: &str) -> Option<String> {
    let opts = spec_doc::Options {
        folder: Some(id.to_string()),
        closure: false,
    };
    let doc = spec_doc::parse(text, &opts);
    if doc.id_form != Some(spec_doc::IdForm::New) || doc.status_legacy {
        return Some(format!(
            "the SPEC {id} is a Go-generation SPEC: give an approved SPEC-<yyyymmdd>-<slug> SPEC"
        ));
    }
    if doc.errors() > 0 {
        return Some(format!(
            "the SPEC {id} has {} error(s); `moai spec check {id}` lists them",
            doc.errors()
        ));
    }
    if !matches!(doc.status.as_deref(), Some("approved" | "completed")) {
        return Some(format!(
            "the SPEC {id} is `{}`, not approved",
            doc.status.as_deref().unwrap_or("unknown")
        ));
    }
    let recorded = doc.approval.as_ref().is_some_and(|a| {
        a.by.is_some() && a.at.is_some() && a.basis.is_some() && a.digest.is_some()
    });
    (!recorded).then(|| format!("the SPEC {id} has no recorded approval (by, at, basis, digest)"))
}

/// [`spec_problem`] of the SPEC `id` as the integration tip `tip` holds it.
fn spec_refusal(project: &Path, tip: &str, id: &str) -> Option<String> {
    if spec_doc::id_form(id).is_none() {
        return Some(format!("`{id}` is not a SPEC id (SPEC-<yyyymmdd>-<slug>)"));
    }
    let path = format!("{}/{id}/spec.md", spec_doc::SPECS_DIR);
    match landed::show_at(project, tip, &path) {
        Err(why) => Some(format!(
            "the SPEC {id} cannot be read from the tip {tip}: {why}"
        )),
        Ok(None) => Some(format!(
            "the SPEC {id} is not in the integration branch tip {tip} ({path} is absent there)"
        )),
        Ok(Some(text)) => spec_problem(id, &text),
    }
}

// ---- the merge commit ----------------------------------------------------------------------

/// The message of the merge commit: its subject names the SPEC, the source and the branch, and the
/// trailer block `trailers` ends it.
fn commit_message(spec: &str, from: &str, branch: &str, from_tip: &str, trailers: &str) -> String {
    format!(
        "merge(next-flow): {spec} sync {from} into {branch}, head {}\n\n{trailers}",
        batch::short(from_tip)
    )
}

/// The merge commit of `from_tip` into `tip`: the tree is git's own `merge-tree --write-tree` of
/// the two, the first parent is `tip` and the second `from_tip`. A merge that conflicts is
/// `conflicts`, any other failure of git (no common history included) is `join_failed`.
fn make_merge(
    project: &Path,
    req: &SyncRequest<'_>,
    branch: &str,
    tip: &str,
    from_tip: &str,
) -> Result<String> {
    let out = landed::git(
        project,
        &["merge-tree", "--write-tree", "--no-messages", tip, from_tip],
    )
    .map_err(|why| unchanged(format!("join_failed: {why}")))?;
    match out.status.code() {
        Some(0) => {}
        Some(1) => {
            return Err(unchanged(format!(
                "conflicts: {} does not merge cleanly into {branch} at {tip}; the command never \
                 resolves a conflict: merge it by hand on a lane and land that",
                req.from
            )))
        }
        _ => {
            return Err(unchanged(format!(
                "join_failed: git merge-tree failed: {}",
                String::from_utf8_lossy(&out.stderr).trim()
            )))
        }
    }
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    let tree = stdout.lines().next().map_or("", str::trim);
    if tree.is_empty() {
        return Err(unchanged("join_failed: git merge-tree named no tree"));
    }
    let message = commit_message(
        req.spec,
        req.from,
        branch,
        from_tip,
        &batch::trailers(project, req.session, req.spec),
    );
    let made = super::check::git_text(
        project,
        &[
            "commit-tree",
            tree,
            "-p",
            tip,
            "-p",
            from_tip,
            "-m",
            &message,
        ],
    )
    .map_err(|why| {
        unchanged(format!(
            "join_failed: the merge commit cannot be made: {why}"
        ))
    })?;
    Ok(made.trim().to_string())
}

// ---- the ledger of a sync ------------------------------------------------------------------

/// The newest sync of `branch` that moved the branch to `tip` and has no closing row: its start
/// row, when it was a sync of `from` at `from_tip`. A start row that names another source, another
/// tip or a merge the branch does not stand on is some other sync's, and answers `None`.
fn interrupted_sync(
    conn: &Connection,
    branch: &str,
    from: &str,
    from_tip: &str,
    tip: &str,
) -> Result<Option<SyncRow>> {
    let newest: Option<String> = conn
        .query_row(
            "SELECT payload FROM event WHERE kind = ?1 \
               AND json_extract(payload, '$.branch') = ?2 \
             ORDER BY event_id DESC LIMIT 1",
            params![drift::SYNC_STARTED, branch],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| e.to_string())?;
    let Some(payload) = newest else {
        return Ok(None);
    };
    let Ok(payload) = serde_json::from_str::<Value>(&payload) else {
        return Ok(None);
    };
    let field = |key: &str| payload.get(key).and_then(Value::as_str).map(str::to_string);
    let (Some(merge), Some(row_from), Some(row_tip), Some(old_tip), Some(spec)) = (
        field("merge"),
        field("from"),
        field("from_tip"),
        field("old_tip"),
        field("spec"),
    ) else {
        return Ok(None);
    };
    if row_from != from || row_tip != from_tip || merge != tip {
        return Ok(None);
    }
    let closed: bool = conn
        .query_row(
            "SELECT EXISTS (SELECT 1 FROM event WHERE kind = ?1 \
               AND json_extract(payload, '$.branch') = ?2 \
               AND json_extract(payload, '$.merge') = ?3 \
               AND json_extract(payload, '$.old_tip') = ?4)",
            params![drift::SYNCED, branch, merge, old_tip],
            |r| r.get(0),
        )
        .map_err(|e| e.to_string())?;
    Ok((!closed).then(|| SyncRow {
        branch: branch.to_string(),
        merge,
        from: row_from,
        from_tip: row_tip,
        old_tip,
        spec,
    }))
}

// ---- the answer ----------------------------------------------------------------------------

/// What the answer of the command reports.
struct Facts<'a> {
    branch: &'a str,
    from: &'a str,
    from_tip: &'a str,
    spec: &'a str,
    moved: bool,
    already_synced: bool,
    closed_previous: bool,
    merge: Option<&'a str>,
    to_tip: &'a str,
    rev_list_count: Option<u64>,
    variants: &'a [String],
    profile_disagrees: Option<&'a str>,
    worktree_follow: Value,
    runs: Vec<Value>,
    writes: bool,
}

/// The one shape of the answer: a sync that moved, a source already inside the tip and a sync
/// that an earlier call left open.
fn answer(f: &Facts<'_>) -> Value {
    json!({
        "action": "land-sync", "branch": f.branch, "from": f.from, "from_tip": f.from_tip,
        "spec": f.spec, "moved": f.moved, "already_synced": f.already_synced,
        "closed_previous": f.closed_previous, "merge": f.merge, "to_tip": f.to_tip,
        "rev_list_count": f.rev_list_count, "variants": f.variants,
        "profile_disagrees": f.profile_disagrees, "worktree_follow": f.worktree_follow,
        "runs": f.runs, "writes": f.writes,
    })
}

/// `git rev-list --count <branch>..<from>` read now (`None` when git cannot say).
fn behind_count(project: &Path, full: &str, from: &str) -> Option<u64> {
    let range = format!("{full}..refs/heads/{from}");
    let out = landed::git(project, &["rev-list", "--count", &range]).ok()?;
    if !out.status.success() {
        return None;
    }
    String::from_utf8_lossy(&out.stdout).trim().parse().ok()
}

/// What the steps after the source gate share.
struct Ctx<'a> {
    project: &'a Path,
    branch: &'a str,
    full: &'a str,
    from: &'a str,
    from_tip: &'a str,
    spec: &'a str,
    tip: &'a str,
    variants: &'a [String],
    profile_disagrees: Option<&'a str>,
}

/// The source is inside the tip: nothing to merge. A sync an earlier call moved and did not close
/// is closed now (the worktree follow, then the closing row); any other call writes nothing.
fn inside_answer(ctx: &Ctx<'_>, ledger: &mut WriteLedger) -> Result<Value> {
    let open = interrupted_sync(ledger.conn(), ctx.branch, ctx.from, ctx.from_tip, ctx.tip)
        .map_err(unchanged)?;
    let mut facts = Facts {
        branch: ctx.branch,
        from: ctx.from,
        from_tip: ctx.from_tip,
        spec: ctx.spec,
        moved: false,
        already_synced: true,
        closed_previous: false,
        merge: None,
        to_tip: ctx.tip,
        rev_list_count: behind_count(ctx.project, ctx.full, ctx.from),
        variants: ctx.variants,
        profile_disagrees: ctx.profile_disagrees,
        worktree_follow: follow::none(),
        runs: Vec::new(),
        writes: false,
    };
    let Some(row) = open else {
        return Ok(answer(&facts));
    };
    let stopped = |why: String| after_move(why, ctx.branch, &row.merge);
    match follow::rerun_from(ctx.project, &row.old_tip, ctx.full, ctx.branch, &row.merge)
        .map_err(stopped)?
    {
        follow::Rerun::Needed(plan) => {
            follow::apply(&plan).map_err(stopped)?;
            facts.worktree_follow =
                follow::answer(Some(&plan.tree), "finished", &plan.from, &plan.to, None);
        }
        follow::Rerun::Nothing(seen) => facts.worktree_follow = seen,
    }
    drift::record_synced(ledger, &row).map_err(stopped)?;
    facts.closed_previous = true;
    facts.merge = Some(&row.merge);
    facts.writes = true;
    Ok(answer(&facts))
}

/// `card land-sync`: see the module comment.
pub fn land_sync(project: &Path, home: &Path, req: &SyncRequest<'_>) -> Result<Value> {
    if req.by != LEADER_BY {
        return Err(unchanged(format!(
            "only the leader syncs a branch: give --by {LEADER_BY}"
        )));
    }
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64);
    batch::leader_reason(project, req.session, now, "land-sync")?;

    let branch = landed::integration_branch(project).map_err(|why| {
        unchanged(format!(
            "a sync is made on the integration branch, and the profile gives none ({why})"
        ))
    })?;
    let full = landed::branch_ref(&branch).map_err(unchanged)?;
    let listed = listed_sources(project).map_err(unchanged)?;
    let source_problem = source_refusal(req.from, &branch, &listed);
    if let Some(why) = source_problem {
        return Err(unchanged(why));
    }
    let from_tip = landed::tip_of(project, req.from).map_err(|_| {
        unchanged(format!(
            "`{}` is not a local branch (refs/heads/{} names no commit): a tag or a \
             remote-tracking name is no source",
            req.from, req.from
        ))
    })?;

    // The tip is read once; the merge is made on it and the move is made from it.
    let tip = landed::tip_of(project, &branch).map_err(unchanged)?;
    let spec_why = spec_refusal(project, &tip, req.spec);
    if let Some(why) = spec_why {
        return Err(unchanged(why));
    }
    let (_, mut ledger) = open_for_write(project, home).map_err(unchanged)?;
    let drifted = drift::refusal(ledger.conn(), &branch, &tip).map_err(unchanged)?;
    if let Some(why) = drifted {
        return Err(unchanged(why));
    }
    // A record that is not there yet is written right before the move, not here.
    let guarded = guard::protected(project, false).map_err(unchanged)?;
    let variants = guard::variants(project, &full).map_err(unchanged)?;
    let ctx = Ctx {
        project,
        branch: &branch,
        full: &full,
        from: req.from,
        from_tip: &from_tip,
        spec: req.spec,
        tip: &tip,
        variants: &variants,
        profile_disagrees: guarded
            .as_ref()
            .and_then(|g| g.profile_disagrees.as_deref()),
    };

    let held = landed::contained(project, &from_tip, &tip).map_err(unchanged)?;
    if held == Containment::NotACommit {
        return Err(unchanged(format!(
            "the tip {from_tip} of `{}` is not a commit of this repository",
            req.from
        )));
    }
    if held == Containment::Inside {
        return inside_answer(&ctx, &mut ledger);
    }

    let merge = make_merge(project, req, &branch, &tip, &from_tip)?;
    // A worktree that has the branch checked out follows the move: every refusal of that comes
    // before the check and before the first write.
    let follow_plan = follow::plan(project, &full, &branch, &tip, &merge).map_err(unchanged)?;

    let run = batch::run_check(project, req.check_cmd, &tip, &merge, 1).map_err(unchanged)?;
    if run.exit != 0 {
        return Err(unchanged(format!(
            "check_failed: the wide check exited {} on the merge {merge}; its log is {}",
            run.exit,
            run.log.display()
        )));
    }
    if !run.scans_covered {
        return Err(unchanged(format!(
            "not_full_pass: the check of the merge {merge} exited 0 but left no full-selection \
             PASS of this command's own wide run (a replaced check program never counts); its \
             log is {}",
            run.log.display()
        )));
    }
    let runs = vec![json!({
        "range": format!("{tip}..{merge}"), "exit": run.exit,
        "log": run.log.to_string_lossy(), "scans_covered": run.scans_covered,
    })];

    // A tip that moved while the check ran has no say in this result: the merge was made on the
    // old one. There is no second try; the leader calls again.
    let again = landed::tip_of(project, &branch).map_err(unchanged)?;
    if again != tip {
        return Err(unchanged(format!(
            "stale_base: the integration branch {branch} moved from {tip} to {again} while the \
             check ran; the merge was made on {tip}: run the call again"
        )));
    }

    // The writes: the ledger row first, then the intent record and git's own compare-and-swap on
    // the tip that was read.
    let guarded = guard::protected(project, true).map_err(unchanged)?;
    let profile_disagrees = guarded.and_then(|g| g.profile_disagrees);
    let common = guard::common_dir_of(project).map_err(unchanged)?;
    let row = SyncRow {
        branch: branch.clone(),
        merge: merge.clone(),
        from: req.from.to_string(),
        from_tip: from_tip.clone(),
        old_tip: tip.clone(),
        spec: req.spec.to_string(),
    };
    drift::record_sync_started(&mut ledger, &row).map_err(unchanged)?;
    guard::write_intent(&common, &full, &tip, &merge).map_err(after_start_row)?;
    let message = format!("moai card land-sync {}", batch::short(&merge));
    let attempt = landed::move_branch(project, &full, &merge, &tip, &message);
    guard::remove_intent(&common);
    if let RefMove::Lost(why) = attempt.map_err(after_start_row)? {
        // Where is the branch now: where it was (git refused for another reason: a lock file, a
        // read-only store), on the merge (another call moved it) or somewhere else.
        let now_tip = landed::tip_of(project, &branch).map_err(after_start_row)?;
        if now_tip == tip {
            return Err(after_start_row(format!(
                "git refused to move the integration branch {branch} to {merge}: {why}"
            )));
        }
        if now_tip == merge {
            return Err(refused(format!(
                "the integration branch {branch} already points at the merge {merge}: another \
                 call moved it, not this one (git said: {why}); run the same call again to close \
                 the sync"
            )));
        }
        return Err(after_start_row(format!(
            "stale_base: the integration branch {branch} no longer points at {tip} (now \
             {now_tip}; git said: {why}) and did not move through this call; read the tip again \
             and run the call again"
        )));
    }

    let mut worktree_follow = follow::none();
    if let Some(plan) = &follow_plan {
        follow::apply(plan).map_err(|why| after_move(why, &branch, &merge))?;
        worktree_follow = follow::answer(Some(&plan.tree), "followed", &plan.from, &plan.to, None);
    }
    drift::record_synced(&mut ledger, &row).map_err(|why| after_move(why, &branch, &merge))?;

    Ok(answer(&Facts {
        branch: &branch,
        from: req.from,
        from_tip: &from_tip,
        spec: req.spec,
        moved: true,
        already_synced: false,
        closed_previous: false,
        merge: Some(&merge),
        to_tip: &merge,
        rev_list_count: behind_count(project, &full, req.from),
        variants: &variants,
        profile_disagrees: profile_disagrees.as_deref(),
        worktree_follow,
        runs,
        writes: true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const TIP: &str = "da585ae4d507bec047ab5a6446bda6e367649282";

    /// SPEC-20261125-land-sync-merge REQ-2: the seam of the check program is read by a debug build
    /// only. The one line that reads `MOAI_TEST_LAND_SYNC_CHECK` stands right under
    /// `cfg(debug_assertions)` and a release build takes `None` (the text is compiled in, so no
    /// file is read at run time).
    #[test]
    fn the_check_program_seam_is_read_by_debug_builds_only() {
        let source = include_str!("../../ops/card.rs");
        let reads: Vec<usize> = source
            .lines()
            .enumerate()
            .filter(|(_, l)| l.contains("env(\"MOAI_TEST_LAND_SYNC_CHECK\")"))
            .map(|(i, _)| i)
            .collect();
        assert_eq!(reads.len(), 1, "the seam is read at one place: {reads:?}");
        let lines: Vec<&str> = source.lines().collect();
        let before = lines[..reads[0]]
            .iter()
            .rev()
            .map(|l| l.trim())
            .find(|l| !l.is_empty() && !l.starts_with("//"));
        assert_eq!(before, Some("#[cfg(debug_assertions)]"));
        let after = lines[reads[0] + 1..]
            .iter()
            .map(|l| l.trim())
            .find(|l| !l.is_empty() && !l.starts_with("//"));
        assert_eq!(after, Some("#[cfg(not(debug_assertions))]"));
        assert!(
            lines[reads[0] + 2].contains("let check_cmd: Option<PathBuf> = None;"),
            "a release build takes no check program: {}",
            lines[reads[0] + 2]
        );
    }

    /// The subject names the SPEC, the source, the branch and the first 12 digits of the source
    /// tip; the trailer block comes last, after one empty line, and `git` reads it as trailers.
    #[test]
    fn the_commit_message_names_the_sync_and_ends_with_the_trailers() {
        let trailers = "Moai-Spec: SPEC-20261125-x\nMoai-Verified: none\n";
        let text = commit_message("SPEC-20261125-x", "main", "NeXT", TIP, trailers);
        assert_eq!(
            text,
            "merge(next-flow): SPEC-20261125-x sync main into NeXT, head da585ae4d507\n\n\
             Moai-Spec: SPEC-20261125-x\nMoai-Verified: none\n"
        );
        assert!(!text.contains(TIP), "only 12 digits of the tip: {text}");
        assert!(text.ends_with(trailers));
        // A short source tip is not cut further than it is.
        let short = commit_message("SPEC-20261125-x", "main", "NeXT", "abc", trailers);
        assert!(short.contains("head abc\n"), "{short}");
    }

    fn list(names: &[&str]) -> Vec<String> {
        names.iter().map(|s| s.to_string()).collect()
    }

    /// REQ-2: only a short, valid branch name that the list holds exactly and that is not the
    /// integration branch is a source; every other reason names the key or the branch.
    #[test]
    fn a_source_is_a_listed_short_branch_name_that_is_not_the_integration_branch() {
        let listed = list(&["main", "release/next"]);
        assert_eq!(source_refusal("main", "NeXT", &listed), None);
        assert_eq!(source_refusal("release/next", "NeXT", &listed), None);
        // The integration branch, spelled either way, is no source even when listed.
        let with_integration = list(&["NeXT"]);
        for integration in ["NeXT", "refs/heads/NeXT"] {
            let why = source_refusal("NeXT", integration, &with_integration).unwrap();
            assert!(why.contains("integration branch"), "{why}");
        }
        // Unlisted, an empty list and a name that differs in case: the key is named.
        for (from, names) in [
            ("hotfix", list(&["main"])),
            ("main", Vec::new()),
            ("Main", list(&["main"])),
            ("mai", list(&["main"])),
        ] {
            let why = source_refusal(from, "NeXT", &names).unwrap();
            assert!(why.contains("work_flow.sync_sources"), "{why}");
            assert!(why.contains(from), "{why}");
        }
        // A full ref, a malformed name and the empty name are no short branch names.
        for from in [
            "refs/heads/main",
            "refs/tags/v1",
            "bad..name",
            "-x",
            "",
            "a b",
        ] {
            let why = source_refusal(from, "NeXT", &list(&[from])).unwrap();
            assert!(!why.is_empty(), "{from}");
        }
    }

    fn spec_text(status: &str, approval: &str) -> String {
        format!(
            "---\nid: SPEC-20261125-x\ntitle: \"x\"\nstatus: {status}\nrevision: 1\n\
             created: 2026-11-17\n{approval}---\n\n# x\n\n## Intent\n\nWhy.\n\n## Requirements\n\n\
             - REQ-1: When a user sends a valid password, the server shall accept the login. · \
             failure: none\n\n## Acceptance\n\n- AC-1 (REQ-1): it logs in · verify: `cargo test \
             --test login`\n\n## Scope\n\n- write: `src/login.rs`\n\n## Plan\n\n1. Fix it.\n"
        )
    }

    /// REQ-2: a draft, a bad id and a missing approval are each refused with the SPEC's name; an
    /// approved SPEC with its recorded approval passes.
    #[test]
    fn only_an_approved_spec_with_a_recorded_approval_passes() {
        let id = "SPEC-20261125-x";
        let draft = spec_text("draft", "");
        let why = spec_problem(id, &draft).unwrap();
        assert!(why.contains(id) && why.contains("not approved"), "{why}");
        // Approved by its word alone, with no approval record: refused (the digest also errors).
        let bare = spec_text("approved", "");
        let why = spec_problem(id, &bare).unwrap();
        assert!(why.contains(id), "{why}");
        // A text whose folder name differs from its id has an error finding.
        let why = spec_problem("SPEC-20261125-other", &draft).unwrap();
        assert!(why.contains("SPEC-20261125-other"), "{why}");
        // The approval record with the digest of the body passes.
        let digest = spec_doc::parse(
            &spec_text("draft", ""),
            &spec_doc::Options {
                folder: Some(id.to_string()),
                closure: false,
            },
        )
        .digest
        .expect("a GEARS draft has a digest");
        let approved = spec_text(
            "approved",
            &format!(
                "approval:\n  by: user\n  at: 2026-11-17\n  basis: test\n  digest: {digest}\n"
            ),
        );
        assert_eq!(spec_problem(id, &approved), None);
    }

    /// The answer has the same keys whether the branch moved, the source was inside the tip or an
    /// open sync was closed; a value that is not known is `null`.
    #[test]
    fn the_answer_has_one_shape_for_a_move_and_for_no_move() {
        let variants = vec!["refs/heads/next".to_string()];
        let moved = answer(&Facts {
            branch: "NeXT",
            from: "main",
            from_tip: TIP,
            spec: "SPEC-20261125-x",
            moved: true,
            already_synced: false,
            closed_previous: false,
            merge: Some("m"),
            to_tip: "m",
            rev_list_count: Some(0),
            variants: &variants,
            profile_disagrees: Some("other"),
            worktree_follow: follow::none(),
            runs: vec![json!({"exit": 0})],
            writes: true,
        });
        let idle = answer(&Facts {
            branch: "NeXT",
            from: "main",
            from_tip: TIP,
            spec: "SPEC-20261125-x",
            moved: false,
            already_synced: true,
            closed_previous: false,
            merge: None,
            to_tip: TIP,
            rev_list_count: None,
            variants: &[],
            profile_disagrees: None,
            worktree_follow: follow::none(),
            runs: Vec::new(),
            writes: false,
        });
        let keys = |v: &Value| -> Vec<String> { v.as_object().unwrap().keys().cloned().collect() };
        assert_eq!(keys(&moved), keys(&idle));
        for key in [
            "action",
            "branch",
            "from",
            "from_tip",
            "spec",
            "moved",
            "already_synced",
            "closed_previous",
            "merge",
            "to_tip",
            "rev_list_count",
            "variants",
            "profile_disagrees",
            "worktree_follow",
            "runs",
            "writes",
        ] {
            assert!(moved.get(key).is_some(), "{key} is missing");
        }
        assert_eq!(moved["action"], "land-sync");
        assert_eq!(moved["rev_list_count"], 0);
        assert_eq!(moved["profile_disagrees"], "other");
        assert_eq!(idle["merge"], Value::Null);
        assert_eq!(idle["rev_list_count"], Value::Null);
        assert_eq!(idle["profile_disagrees"], Value::Null);
        assert_eq!(idle["writes"], false);
        assert_eq!(idle["variants"], json!([]));
        assert_eq!(idle["runs"], json!([]));
    }

    /// A ledger with the one table the reader of an interrupted sync reads.
    fn events(rows: &[(&str, Value)]) -> Connection {
        let conn = Connection::open_in_memory().unwrap();
        conn.execute_batch(
            "CREATE TABLE event (event_id INTEGER PRIMARY KEY, kind TEXT NOT NULL, \
             payload TEXT NOT NULL)",
        )
        .unwrap();
        for (kind, payload) in rows {
            conn.execute(
                "INSERT INTO event (kind, payload) VALUES (?1, ?2)",
                params![kind, payload.to_string()],
            )
            .unwrap();
        }
        conn
    }

    fn sync_payload(merge: &str, old_tip: &str) -> Value {
        json!({
            "branch": "NeXT", "merge": merge, "from": "main", "from_tip": TIP,
            "old_tip": old_tip, "spec": "SPEC-20261125-x", "by": "leader",
        })
    }

    /// REQ-6: a start row with no closing row, whose merge the branch stands on and whose source
    /// is the one asked for, is the sync to close; a closing row, another source, another tip and
    /// a start row of another branch are not.
    #[test]
    fn only_the_newest_open_start_row_of_the_same_sync_is_read_as_interrupted() {
        let (m1, m2, old) = ("1".repeat(40), "2".repeat(40), "0".repeat(40));
        let start = |m: &str| (drift::SYNC_STARTED, sync_payload(m, &old));
        let close = |m: &str| (drift::SYNCED, sync_payload(m, &old));
        let read = |conn: &Connection, from: &str, from_tip: &str, tip: &str| {
            interrupted_sync(conn, "NeXT", from, from_tip, tip).unwrap()
        };
        let open = events(&[start(&m1)]);
        let row = read(&open, "main", TIP, &m1).expect("an open sync");
        assert_eq!(
            (row.merge.as_str(), row.old_tip.as_str(), row.spec.as_str()),
            (m1.as_str(), old.as_str(), "SPEC-20261125-x")
        );
        assert_eq!((row.branch.as_str(), row.from.as_str()), ("NeXT", "main"));
        // The branch does not stand on the merge, another source, another source tip.
        assert!(read(&open, "main", TIP, &m2).is_none());
        assert!(read(&open, "other", TIP, &m1).is_none());
        assert!(read(&open, "main", &"3".repeat(40), &m1).is_none());
        // A closing row of the same sync closes it; one of another merge does not.
        assert!(read(&events(&[start(&m1), close(&m1)]), "main", TIP, &m1).is_none());
        assert!(read(&events(&[start(&m1), close(&m2)]), "main", TIP, &m1).is_some());
        // The newest start row decides.
        assert!(read(&events(&[start(&m1), start(&m2)]), "main", TIP, &m1).is_none());
        assert!(read(&events(&[start(&m1), start(&m2)]), "main", TIP, &m2).is_some());
        // Nothing, a start row of another branch and a payload with a missing key.
        assert!(read(&events(&[]), "main", TIP, &m1).is_none());
        let mut other_branch = sync_payload(&m1, &old);
        other_branch["branch"] = json!("other");
        assert!(read(
            &events(&[(drift::SYNC_STARTED, other_branch)]),
            "main",
            TIP,
            &m1
        )
        .is_none());
        let mut broken = sync_payload(&m1, &old);
        broken.as_object_mut().unwrap().remove("old_tip");
        assert!(read(&events(&[(drift::SYNC_STARTED, broken)]), "main", TIP, &m1).is_none());
    }

    /// The refusals say what they changed: the helpers' words are fixed.
    #[test]
    fn the_refusal_words_say_what_was_changed() {
        assert_eq!(
            unchanged("why"),
            "card land-sync refused: why; nothing was changed"
        );
        let after = after_start_row("why");
        assert!(after.starts_with("card land-sync refused: why; only the start row"));
        assert!(!after.contains("nothing was changed"), "{after}");
        let moved = after_move("why", "NeXT", "abc");
        assert!(moved.contains("already moved to abc"), "{moved}");
        assert!(!moved.contains("nothing was changed"), "{moved}");
    }
}
