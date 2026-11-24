//! `moai card land` (SPEC-20261121-card-land): the leader moves the integration branch to a
//! candidate commit by one compare-and-swap and closes the card in the one transaction
//! `card finish` uses.
//!
//! Order of a call, so that nothing is written until everything that can be known is known:
//!
//! 1. the caller word (only the leader lands until the typed verdict exists), the shape of the
//!    met options and of the head;
//! 2. the card: a card that closes on a report is `card finish`'s; a card already `done` with the
//!    same landed head answers done and writes nothing, with another head it refuses;
//! 3. git: the integration branch is read, the candidate is a commit, and where it stands
//!    against the tip decides the move (inside the tip: no move; the tip on its first-parent
//!    line: move; anything else is `stale_base`); a move is refused while any worktree has the
//!    branch checked out;
//! 4. the SPEC the card names passes at the candidate, and every check `card finish` would make
//!    passes ([`workflow::finish_preflight`], the same code the finish transaction runs first);
//! 5. a move starts with one ledger row ([`workflow::record_land_started`], run through the same
//!    checks as the close, once per card, head and tip) and the intent record the guard of the
//!    integration branch reads (SPEC-20261121-ref-guard); the record is removed when the update
//!    is over, whatever came of it;
//! 6. the move: `git update-ref <branch> <candidate> <tip read>` is git's own compare-and-swap, so
//!    a branch that moved in between is refused as `tip_moved`, and a reference update git
//!    refuses for another reason (the guard, a lock file) is reported with git's reason;
//! 7. the close: [`workflow::finish`] with the candidate as the landed head. A close that is
//!    refused after the move leaves the branch moved and the start row in the ledger, and says
//!    so; the same call run again finds the candidate inside the tip, skips the move and closes
//!    (or repeats the refusal).
//!
//! The answer names `variants` (branches that differ from the integration branch's name only in
//! letter case) and `profile_disagrees` (the profile's branch when it is not the recorded one).

pub mod batch;
pub mod check;
pub mod scans;

use super::command::open_for_write;
use super::drift;
use super::finish::{parse_met, previous_json};
use super::guard;
use super::input::{self, CardRef};
use super::landed::{self, Containment, RefMove, Standing};
use super::spec_link;
use crate::ledger::workflow::{self, FinishEvidence, WorkflowError, LEADER_BY};
use crate::ledger::LedgerError;
use rusqlite::OptionalExtension;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// What `card land` was given.
pub struct LandRequest<'a> {
    /// This session's id, when it has one.
    pub session: Option<&'a str>,
    /// `--head`: the candidate commit (40 lowercase hex).
    pub head: &'a str,
    /// Each `--met N=REF`.
    pub met: &'a [String],
    /// `--generation`: the generation of the card's newest reservation the caller expects.
    pub generation: Option<i64>,
    /// `--by`: who lands; only the leader word is accepted.
    pub by: &'a str,
    /// `--keep-open`: land a slice, record the landing and close nothing
    /// (SPEC-20261124-land-closes-card).
    pub keep_open: bool,
    /// The audited head a join candidate was made of (SPEC-20261124-card-select-batch-land D-5):
    /// the verdict check reads the verdicts of this head instead of the candidate's. Only a
    /// two-parent candidate whose second parent is this head and whose tree is the clean merge of
    /// its parents qualifies.
    pub audited: Option<&'a str>,
}

/// The refusal of an `audited` head that the candidate is not a clean join of.
fn audited_refusal(project: &Path, candidate: &str, audited: &str) -> Option<String> {
    if !crate::gate::is_full_sha(audited) {
        return Some("the audited head must be 40 lowercase hex".to_string());
    }
    let line = match check::git_text(project, &["rev-list", "--parents", "-n", "1", candidate]) {
        Ok(line) => line,
        Err(why) => return Some(format!("the candidate's parents cannot be read: {why}")),
    };
    let ids: Vec<&str> = line.split_whitespace().collect();
    if ids.len() != 3 || ids[2] != audited {
        return Some(format!(
            "the audited head {audited} is not the second parent of a two-parent candidate"
        ));
    }
    if !check::is_clean_join(project, candidate, ids[1], ids[2]) {
        return Some(
            "the candidate is not a clean join of its parents (its tree differs from the merge of \
             the two parents)"
                .to_string(),
        );
    }
    None
}

/// The refusal of a join candidate whose first parent is not `want`: the branch moves only from
/// the commit the join was made on, so a first parent beyond the tip would carry commits nobody
/// landed.
fn first_parent_refusal(project: &Path, candidate: &str, want: &str) -> Option<String> {
    let line = match check::git_text(project, &["rev-list", "--parents", "-n", "1", candidate]) {
        Ok(line) => line,
        Err(why) => return Some(format!("the candidate's parents cannot be read: {why}")),
    };
    let first = line.split_whitespace().nth(1);
    (first != Some(want)).then(|| {
        format!(
            "the first parent of the candidate {candidate} is {}, not the integration tip {want}: \
             join the head onto the tip again",
            first.unwrap_or("missing")
        )
    })
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card land refused: {why}")
}

/// A refusal that comes before any write.
fn unchanged(why: impl std::fmt::Display) -> String {
    refused(format!("{why}; nothing was changed"))
}

/// A refusal about the state of the repository (a conflict, not a mistake in the call).
fn conflict(why: impl std::fmt::Display) -> String {
    format!("card conflict in landing: {why}; nothing was changed")
}

fn card_id(reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(refused)? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(refused(
            "name the card by its C-number (`card show #N` finds it)",
        )),
    }
}

/// What a refusal after the start row says in place of `nothing was changed`: the one row that
/// was written and that the branch did not move.
const START_ROW_KEPT: &str =
    "only the start row of this landing was written to the ledger and the branch did not move";

/// A refusal of the move after the start row was written.
fn after_start_row(why: impl std::fmt::Display) -> String {
    refused(format!("{why}; {START_ROW_KEPT}"))
}

/// A refusal about the state of the repository after the start row was written.
fn conflict_after_start_row(why: impl std::fmt::Display) -> String {
    format!("card conflict in landing: {why}; {START_ROW_KEPT}")
}

/// The sentence a refusal that wrote nothing ends with. Finish's own refusals end with it too, so
/// it is added once.
const NOTHING_CHANGED: &str = "nothing was changed";

/// A refusal of the finish checks, or of a finish that comes before the branch moved: said as
/// `card finish` words it for an input mistake, and with `nothing was changed` once for a check
/// that refused before any write.
fn finish_refusal(e: WorkflowError) -> String {
    match e {
        WorkflowError::InvalidInput(why) => refused(why),
        WorkflowError::FinishRefused(why) if why.ends_with(NOTHING_CHANGED) => refused(why),
        WorkflowError::FinishRefused(why) => unchanged(why),
        other => other.to_string(),
    }
}

/// A refusal of the close after the branch moved. The branch was changed, so the refusal never
/// says that nothing was: it says the branch moved and the call can be run again.
fn refusal_after_move(e: WorkflowError, branch: &str, head: &str) -> String {
    let why = match e {
        WorkflowError::InvalidInput(why) | WorkflowError::FinishRefused(why) => why,
        other => other.to_string(),
    };
    let why = why
        .strip_suffix(&format!("; {NOTHING_CHANGED}"))
        .unwrap_or(&why);
    format!(
        "card land refused after the move: {why}; the integration branch {branch} already moved \
         to {head} and the card is still open: run the same call again after the cause is repaired"
    )
}

/// The reason of the `no_pass` refusal: the code `no_pass` when the check found no verdict for
/// the head, the check's own code and reason for any other finding (a FAIL, an unreadable git).
fn no_pass_reason(would_refuse: Option<&Value>) -> String {
    let code = would_refuse.and_then(|w| w["code"].as_str());
    let why = would_refuse.and_then(|w| w["why"].as_str()).unwrap_or("");
    match code {
        None | Some("no_verdict") | Some("verdict_other_head") => {
            format!(
            "no_pass: a landing that closes the card needs an independent PASS of this exact head \
             recorded by `card verdict` ({}); land a slice with --keep-open instead",
            if why.is_empty() { "none was found" } else { why }
        )
        }
        Some(code) => format!("{code}: {why}"),
    }
}

/// `card land C-n`: see the module comment.
pub fn land(project: &Path, home: &Path, reference: &str, req: &LandRequest<'_>) -> Result<Value> {
    land_inner(project, home, reference, req, None)
}

/// Every refusal of [`land`] that comes before the first write, and nothing written: the batch asks
/// it for each head before the first of them lands, so a head that would be refused is dropped
/// while the tip is still the one the check ran on. `parent` is the commit the candidate's first
/// parent must be: the tip for the first head, the commit of the head before it for the next.
pub fn land_preflight(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &LandRequest<'_>,
    parent: &str,
) -> Result<Value> {
    land_inner(project, home, reference, req, Some(parent))
}

fn land_inner(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &LandRequest<'_>,
    preflight: Option<&str>,
) -> Result<Value> {
    let card_id = card_id(reference)?;
    if req.by != LEADER_BY {
        return Err(unchanged(format!(
            "only the leader lands a card until the typed verdict exists: give --by {LEADER_BY}"
        )));
    }
    let met = parse_met(req.met).map_err(refused)?;
    if req.keep_open && !met.is_empty() {
        return Err(unchanged(
            "--keep-open closes nothing, so it takes no --met option: leave --met out, or land \
             without --keep-open",
        ));
    }
    if !crate::gate::is_full_sha(req.head) {
        return Err(refused("the head must be 40 lowercase hex"));
    }
    if let Some(audited) = req.audited {
        if let Some(why) = audited_refusal(project, req.head, audited) {
            return Err(unchanged(why));
        }
    }
    let verdict_head = req.audited.unwrap_or(req.head);
    let sql_err = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let (id, mut ledger) = open_for_write(project, home)?;
    let (kind, spec) = spec_link::kind_and_spec(ledger.conn(), card_id)
        .map_err(sql_err)?
        .ok_or_else(|| WorkflowError::CardNotFound(card_id).to_string())?;
    if !spec_link::changes_repository(&kind) {
        return Err(unchanged(format!(
            "a `{kind}` card closes on a report, not on a landing: use `card finish`"
        )));
    }
    let name = input::card_name(card_id);

    // A card that is already done: the same head answers done, another head refuses.
    let state: String = ledger
        .conn()
        .query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .map_err(sql_err)?;
    if state == "done" {
        let recorded: Option<String> = ledger
            .conn()
            .query_row(
                "SELECT json_extract(payload, '$.head') FROM event \
                 WHERE card_id = ?1 AND kind = 'card.landed' \
                   AND json_type(payload, '$.head') = 'text' \
                 ORDER BY event_id DESC LIMIT 1",
                [card_id],
                |r| r.get(0),
            )
            .optional()
            .map_err(sql_err)?;
        return match recorded {
            Some(head) if head == req.head => Ok(json!({
                "action": "land", "project_id": id, "name": name, "card_id": card_id,
                "state": "done", "closed": true, "moved": false, "already_done": true,
                "head": req.head, "by": req.by, "writes": false,
            })),
            Some(head) => Err(unchanged(format!(
                "{name} is done with the landed head {head}; the head given is {}",
                req.head
            ))),
            None => Err(unchanged(format!(
                "{name} is done and records no landed head; the head given is {}",
                req.head
            ))),
        };
    }

    // The criteria still open: a slice needs one (`--keep-open` of a card with nothing open would
    // close it), and a closing landing fills each one the caller did not name.
    let open: Vec<i64> = {
        let mut stmt = ledger
            .conn()
            .prepare("SELECT seq FROM card_ac WHERE card_id = ?1 AND status = 'open' ORDER BY seq")
            .map_err(sql_err)?;
        let rows = stmt.query_map([card_id], |r| r.get(0)).map_err(sql_err)?;
        rows.collect::<std::result::Result<_, _>>()
            .map_err(sql_err)?
    };
    if req.keep_open && open.is_empty() {
        return Err(unchanged(format!(
            "{name} has no open criterion, so --keep-open has nothing to keep open: land without it"
        )));
    }

    // Git: the branch, the tip, the candidate and where it stands.
    let branch = landed::integration_branch(project).map_err(|why| {
        unchanged(format!(
            "a landing is made on the integration branch, and the profile gives none ({why})"
        ))
    })?;
    let full = landed::branch_ref(&branch).map_err(unchanged)?;
    let tip = landed::tip_of(project, &branch).map_err(unchanged)?;
    match landed::contained(project, req.head, &tip).map_err(unchanged)? {
        Containment::NotACommit => {
            return Err(unchanged(format!(
                "the head {} is not a commit of this repository",
                req.head
            )))
        }
        Containment::Inside | Containment::Outside => {}
    }
    let standing = landed::standing(project, req.head, &tip).map_err(unchanged)?;
    match standing {
        Standing::Inside => {}
        Standing::OnTip => {
            let trees = landed::checked_out_at(project, &full).map_err(unchanged)?;
            if !trees.is_empty() {
                return Err(conflict(format!(
                    "the integration branch {branch} is checked out in {}: a move under it would \
                     leave that worktree behind the reference",
                    trees.join(", ")
                )));
            }
        }
        Standing::Stale => {
            return Err(conflict(format!(
                "stale_base: the integration tip {tip} is not on the first-parent line of {}: \
                 merge the tip into the head and land the merge",
                req.head
            )))
        }
    }
    // A join candidate is landed from the commit it was made on: its first parent is the tip read
    // here (or, in the batch's preflight, the commit of the head before it).
    if req.audited.is_some() && standing == Standing::OnTip {
        if let Some(why) = first_parent_refusal(project, req.head, preflight.unwrap_or(&tip)) {
            return Err(unchanged(why));
        }
    }
    if let Some(why) = drift::refusal(ledger.conn(), &branch, &tip).map_err(unchanged)? {
        return Err(conflict(why));
    }
    let to_tip = if standing == Standing::OnTip {
        req.head.to_string()
    } else {
        tip.clone()
    };

    // Every check that can be made without a write: the SPEC at the candidate, then finish's own.
    if let Some(why) = spec_link::done_refusal_at_tip(project, &name, spec.as_deref(), req.head) {
        return Err(unchanged(why));
    }
    let evidence = FinishEvidence::Landing {
        head: req.head,
        tip: &to_tip,
    };
    workflow::finish_preflight(
        ledger.conn(),
        card_id,
        req.session,
        req.by,
        &evidence,
        &met,
        req.generation,
    )
    .map_err(finish_refusal)?;

    // The audit verdict (SPEC-20261123-landing-reads-verdict): read before the first write. A head
    // the branch already holds is closed on the verdict its start row names; a head that still
    // stands on the tip has moved nothing, so its verdicts are judged again (REQ-15, REQ-24). In
    // observe mode a refusal is only reported, and the landing goes on as it did before the check.
    let mode = check::mode(project).map_err(unchanged)?;
    let input = check::Input {
        project,
        conn: ledger.conn(),
        card_id,
        head: req.head,
        verdict_head,
        tip: &tip,
        on_tip: standing == Standing::OnTip,
        lander: req.session,
    };
    let check::Reading {
        relied,
        would_refuse,
    } = check::read(&input, mode).map_err(unchanged)?;

    // A landing that closes the card needs the PASS the check relies on, in every mode
    // (SPEC-20261124-land-closes-card REQ-4): `no_pass` for a head with no usable verdict, the
    // check's own code and reason for a FAIL or anything else it found. A slice (`--keep-open`)
    // goes on as the mode says. Nothing has been written yet.
    let met = if req.keep_open {
        met
    } else {
        let Some(relied) = relied.as_ref() else {
            return Err(unchanged(no_pass_reason(would_refuse.as_ref())));
        };
        let Some(event) = relied["event"].as_i64() else {
            return Err(unchanged(
                "no_pass: the verdict the landing relies on names no event, so it cannot be \
                 cited as evidence",
            ));
        };
        let line = format!("verdict event {event} PASS head {}", &verdict_head[..12]);
        let mut filled = met;
        for seq in &open {
            if !filled.iter().any(|(named, _)| named == seq) {
                filled.push((*seq, line.clone()));
            }
        }
        // The same checks again with the criteria filled, still before the first write.
        workflow::finish_preflight(
            ledger.conn(),
            card_id,
            req.session,
            req.by,
            &evidence,
            &filled,
            req.generation,
        )
        .map_err(finish_refusal)?;
        filled
    };

    // The scans the merge tree needs (SPEC-20261123-land-repo-scans): only a landing that moves the
    // branch asks for them; a head the branch already holds was judged when it moved (REQ-9).
    let scans = if standing == Standing::OnTip {
        scans::read(&input, mode).map_err(unchanged)?
    } else {
        Value::Null
    };

    // The batch's preflight stops here: every check above is read only.
    if preflight.is_some() {
        return Ok(json!({
            "action": "land", "preflight": true, "project_id": id, "name": name,
            "card_id": card_id, "head": req.head, "writes": false,
        }));
    }

    // The one value for the whole repository: the record seeds from the profile once, and a
    // profile that disagrees with it is reported. Case variants that already exist are reported.
    let guarded = guard::protected(project, true).map_err(unchanged)?;
    let profile_disagrees = guarded.and_then(|g| g.profile_disagrees);
    let variants = guard::variants(project, &full).map_err(unchanged)?;

    // The move: the ledger row first, then the intent record and git's compare-and-swap on the
    // tip that was read.
    let mut moved = false;
    if standing == Standing::OnTip {
        let common = guard::common_dir_of(project).map_err(unchanged)?;
        workflow::record_land_started(
            &mut ledger,
            card_id,
            req.session,
            req.by,
            &evidence,
            &met,
            req.generation,
            &branch,
            &tip,
            relied.as_ref(),
        )
        .map_err(finish_refusal)?;
        guard::write_intent(&common, &full, &tip, req.head).map_err(after_start_row)?;
        let message = format!("moai card land {}", &req.head[..12]);
        let attempt = landed::move_branch(project, &full, req.head, &tip, &message);
        guard::remove_intent(&common);
        match attempt.map_err(after_start_row)? {
            RefMove::Moved => moved = true,
            RefMove::Lost(why) => {
                // Read where the branch is now: another call may have landed this very head a
                // moment ago (nothing to move), the branch may have moved to something else
                // (`tip_moved`), or it did not move at all and git refused for another reason (a
                // lock file, a read-only store): that one is git's reason, not a moved tip.
                let now = landed::tip_of(project, &branch).map_err(after_start_row)?;
                if now == tip {
                    return Err(after_start_row(format!(
                        "git refused to move the integration branch {branch} to {}: {why}",
                        req.head
                    )));
                }
                if now != req.head {
                    return Err(conflict_after_start_row(format!(
                        "tip_moved: the integration branch {branch} no longer points at {tip} \
                         (now {now}; git said: {why}); read the tip again and run the call again"
                    )));
                }
            }
        }
    }

    // The close. A refusal here comes after the move and says so.
    // The landing event names the verdict this call relied on, not the one an earlier call's start
    // row names: a call that stopped before the move and judged again has a newer one (REQ-13).
    let done = workflow::finish_naming(
        &mut ledger,
        card_id,
        req.session,
        req.by,
        &evidence,
        &met,
        req.generation,
        workflow::LandedVerdict::Relied(relied.as_ref()),
        req.keep_open,
    )
    .map_err(|e| {
        if moved {
            refusal_after_move(e, &branch, req.head)
        } else {
            finish_refusal(e)
        }
    })?;
    let applied: &[(i64, String)] = if done.closed { &met } else { &[] };
    // A closed card's worktrees and branches are cleaned by the cleanup executor
    // (SPEC-20261123-cleanup-outbox): the close wrote this card's request, and the executor takes
    // it with every older pending one. It runs after the close and never undoes it: a problem is
    // the `error` of `cleanup_outbox`. The `cleanup` key keeps the worktree answer of this card.
    // A card that stays open keeps its worktrees and branches.
    let outbox = if done.closed {
        super::cleanup::run_pending(
            project,
            home,
            req.session.unwrap_or(req.by),
            &super::cleanup::Options::default(),
        )
    } else {
        Value::Null
    };
    let cleanup = if done.closed {
        super::cleanup::cleanup_of(&outbox, card_id)
    } else {
        Value::Null
    };
    let mut answer = json!({
        "action": "land", "project_id": id, "name": name, "card_id": card_id,
        "branch": branch, "moved": moved, "from_tip": tip, "to_tip": to_tip, "head": req.head,
        "variants": variants, "profile_disagrees": profile_disagrees,
        "from": done.from, "state": done.state, "rev": done.rev, "closed_at": done.closed_at,
        "closed": done.closed, "open_criteria": done.open_criteria,
        "met": applied.iter().map(|(seq, text)| json!({"seq": seq, "evidence": text})).collect::<Vec<_>>(),
        "by": req.by, "session": req.session,
        "previous": previous_json(&done.previous), "released": done.released,
        // A worktree the cleanup removed or trimmed is a write too.
        "events": done.events,
        "writes": done.events > 0 || moved || super::cleanup::wrote(&outbox),
        "cleanup": cleanup,
        "scans": scans,
        "cleanup_outbox": outbox,
    });
    if let Some(verdict) = &relied {
        answer["verdict"] = verdict.clone();
        if let Some(after) = verdict["event"].as_i64() {
            if let Some(late) =
                check::late_fail(ledger.conn(), card_id, verdict_head, after).map_err(unchanged)?
            {
                answer["late_fail"] = late;
            }
        }
    }
    if let Some(refusal) = would_refuse {
        answer["would_refuse"] = refusal;
    }
    Ok(answer)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn g(dir: &Path, args: &[&str]) -> String {
        let mut all = vec!["-c", "user.name=t", "-c", "user.email=t@example.invalid"];
        all.extend_from_slice(args);
        check::git_text(dir, &all).unwrap().trim().to_string()
    }

    /// SPEC-20261124-card-select-batch-land D-5: an `audited` head is taken only for a clean join
    /// of the tip and that head, with the head as the second parent.
    #[test]
    fn audited_must_be_the_second_parent_of_a_clean_join() {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        g(p, &["init", "-q", "-b", "main"]);
        std::fs::write(p.join("a.txt"), "a\n").unwrap();
        g(p, &["add", "a.txt"]);
        g(p, &["commit", "-q", "-m", "tip"]);
        let tip = g(p, &["rev-parse", "HEAD"]);
        g(p, &["checkout", "-q", "-b", "lane"]);
        std::fs::write(p.join("b.txt"), "b\n").unwrap();
        g(p, &["add", "b.txt"]);
        g(p, &["commit", "-q", "-m", "lane"]);
        let lane = g(p, &["rev-parse", "HEAD"]);
        g(p, &["checkout", "-q", "main"]);
        let merged = g(p, &["merge-tree", "--write-tree", &tip, &lane]);
        let tree = merged.lines().next().unwrap().to_string();
        let join = g(
            p,
            &["commit-tree", &tree, "-p", &tip, "-p", &lane, "-m", "join"],
        );
        assert_eq!(audited_refusal(p, &join, &lane), None);
        // The tip is the first parent, not the audited head; a short id is no head.
        assert!(audited_refusal(p, &join, &tip).is_some());
        assert!(audited_refusal(p, &join, "abc").is_some());
        // A commit with one parent is no join.
        assert!(audited_refusal(p, &lane, &tip).is_some());
        // A join whose tree is not the merge of its parents counts as authored work.
        std::fs::write(p.join("evil.txt"), "e\n").unwrap();
        g(p, &["add", "evil.txt"]);
        let evil = g(p, &["write-tree"]);
        let edited = g(
            p,
            &[
                "commit-tree",
                &evil,
                "-p",
                &tip,
                "-p",
                &lane,
                "-m",
                "edited",
            ],
        );
        assert!(audited_refusal(p, &edited, &lane).is_some());
    }
}
