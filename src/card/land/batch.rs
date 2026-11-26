//! `moai card land-batch` (SPEC-20261124-card-select-batch-land): 2 to 4 audited heads are joined
//! one after another onto the integration tip, the wide check runs once on the last join in a
//! private worktree, and each head lands through [`super::land`] in the given order.
//!
//! The loop, so that one wide run serves as many heads as it can and no result outlives its tip:
//!
//! 1. the tip is read and the pending heads are joined onto it in order (a head the tip already
//!    holds or stands on needs no join; a head that cannot be joined is dropped);
//! 2. the check runs on the last join. On PASS the heads land in order, each by its own
//!    compare-and-swap; the tip has moved, so what is still pending is joined and checked again.
//!    A tip that moved between the run and the landing sends the same set through step 1 again;
//! 3. on FAIL a set of one head drops that head; a larger set is cut to its first half and that
//!    half runs on its own. A passing half lands; the rest is checked again on the new tip.
//!
//! Each head is judged by the landing itself (verdict, scans, stale tip): a refusal stops the
//! batch at that head and the heads after it are reported as not run. A join candidate is landed
//! with its audited head named, so the verdict of the audited head covers it.
//!
//! Who may call it, what a run covers and when it gives up (SPEC revision 3):
//!
//! - the calling session must be the session of the live `leader` entry of the lane registry
//!   ([`confirm_leader`]); `--by leader` alone is a word anyone can type;
//! - a wide run that passed on the last join leaves a summary line in its private worktree; a
//!   full-selection PASS line of that join covers the scans the merge tree needs, so the heads land
//!   in enforce mode without a scan receipt of their own ([`receipt_covers`]);
//! - a tip that moves after a check passed sends the heads through the join and the check again,
//!   at most [`MAX_TIP_MOVES`] times.

use super::LandRequest;
use crate::card::input::{self, CardRef};
use crate::card::landed::{self, Containment, Standing};
use crate::ledger::workflow::LEADER_BY;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The most heads one batch lands (the card says 2 to 4).
pub const MAX_HEADS: usize = 4;

/// The most times a passing check is thrown away because the integration branch moved before the
/// heads landed. Every re-run is a cold build, so a branch that moves after each check ends the
/// batch (`tip_kept_moving`) instead of looping.
pub const MAX_TIP_MOVES: usize = 3;

/// What `card land-batch` was given.
pub struct BatchRequest<'a> {
    pub session: Option<&'a str>,
    pub by: &'a str,
    /// Each `C-n=<head>`: a card and its audited head (40 lowercase hex).
    pub pairs: &'a [String],
    /// The check program: it gets `<tip>..<join>` as its one argument, runs in a private worktree
    /// at the join and passes with exit 0. `None` runs the wide `check affected` in this process.
    /// The command line has no flag for it: the tests set it through the environment.
    pub check_cmd: Option<&'a Path>,
}

#[derive(Clone, Debug)]
struct Item {
    name: String,
    head: String,
    /// The SPEC the card names: the join commit of its head is named for it.
    spec: String,
}

/// One head of a set after the join: the commit that lands for it.
struct Candidate {
    item: Item,
    commit: String,
    joined: bool,
}

fn refused(why: impl std::fmt::Display) -> String {
    refusal_of("land-batch", why)
}

/// The refusal text of the command `what`: `card <what> refused: <why>; nothing was changed`.
fn refusal_of(what: &str, why: impl std::fmt::Display) -> String {
    format!("card {what} refused: {why}; nothing was changed")
}

fn parse_pairs(pairs: &[String]) -> Result<Vec<Item>> {
    if !(2..=MAX_HEADS).contains(&pairs.len()) {
        return Err(refused(format!(
            "give 2 to {MAX_HEADS} `C-n=<head>` pairs, got {}",
            pairs.len()
        )));
    }
    let mut items: Vec<Item> = Vec::new();
    for pair in pairs {
        let Some((card, head)) = pair.split_once('=') else {
            return Err(refused(format!("`{pair}` is not `C-n=<40 hex head>`")));
        };
        if !crate::gate::is_full_sha(head) {
            return Err(refused(format!(
                "the head of `{pair}` must be 40 lowercase hex"
            )));
        }
        let name = match input::parse_ref(card).map_err(refused)? {
            CardRef::Id(n) => input::card_name(n),
            CardRef::SourceNumber(_) => {
                return Err(refused("name each card by its C-number"));
            }
        };
        if items.iter().any(|i| i.name == name || i.head == head) {
            return Err(refused(format!("`{pair}` repeats a card or a head")));
        }
        items.push(Item {
            name,
            head: head.to_string(),
            spec: String::new(),
        });
    }
    Ok(items)
}

/// Fill each item's SPEC from the ledger. A card that names none is dropped with the reason: the
/// join commit of its head is named for the SPEC and carries it as `Moai-Spec`, never a card number.
fn attach_specs(project: &Path, home: &Path, items: Vec<Item>) -> Result<(Vec<Item>, Vec<Value>)> {
    let (_, ledger) = crate::card::command::open_for_write(project, home).map_err(refused)?;
    let mut kept = Vec::new();
    let mut dropped = Vec::new();
    for mut item in items {
        let CardRef::Id(id) = input::parse_ref(&item.name).map_err(refused)? else {
            return Err(refused("name each card by its C-number"));
        };
        let spec = crate::card::spec_link::kind_and_spec(ledger.conn(), id)
            .map_err(|e| refused(e.to_string()))?
            .and_then(|(_, spec)| spec);
        match spec {
            Some(spec) => {
                item.spec = spec;
                kept.push(item);
            }
            None => dropped.push(entry(
                &item,
                Some("no_spec: the card names no SPEC, so its join commit cannot name one"),
            )),
        }
    }
    Ok((kept, dropped))
}

pub(super) fn short(sha: &str) -> &str {
    sha.get(..12).unwrap_or(sha)
}

/// The session confirms itself as the leader (SPEC D-12; the rule of `internal/loop.md`, "The
/// confirmed leader"): the lane registry holds a live entry named `leader` in mode `leader`, and
/// its `session` is the session that made this call. The name of a session, its folder or the word
/// `--by leader` prove nothing; an unreadable registry is a refusal, never an answer of "yes".
pub(crate) fn confirm_leader(project: &Path, session: Option<&str>, now: i64) -> Result<()> {
    leader_reason(project, session, now, "land-batch")
}

/// The one leader gate of the card commands that move the integration branch for more than one
/// card: [`confirm_leader`] for `land-batch` and `land-sync`. `what` is the command's name
/// (`land-batch`, `land-sync`); it starts the refusal (`card <what> refused: ...`) and picks what
/// only the leader may do there. The four causes and their words are the same for both.
pub(super) fn leader_reason(
    project: &Path,
    session: Option<&str>,
    now: i64,
    what: &str,
) -> Result<()> {
    let refused = |why: String| refusal_of(what, why);
    let does = match what {
        "land-batch" => "lands a batch",
        "land-sync" => "syncs a branch",
        _ => "runs this command",
    };
    let Some(me) = session else {
        return Err(refused(format!(
            "this call has no session id, and only the registered leader session {does}"
        )));
    };
    let entries = crate::lane_registry::session_entries(project)
        .map_err(|why| refused(format!("the lane registry cannot be read ({why})")))?;
    let live = entries
        .iter()
        .find(|e| e.name == "leader" && e.mode.as_deref() == Some("leader") && e.lease_until > now);
    match live {
        None => Err(refused(
            "no live leader entry is registered: the leader window registers itself with \
             `moai session-register . --name leader --mode leader` (`/moai leader`)"
                .to_string(),
        )),
        Some(entry) if entry.session.as_deref() == Some(me) => Ok(()),
        Some(_) => Err(refused(
            "this session is not the session of the registered leader entry".to_string(),
        )),
    }
}

/// Whether `line` is the summary of a full wide run of `last`: a `check-affected` line that names
/// the head `last` (12 hex digits), ends in `result=PASS` and carries no `selection=` word (a
/// `--card` or `--lane` run checked a part). Such a run ran every listed scan on the tree that
/// lands, so its receipt covers the scans of the heads it holds (SPEC D-11).
pub(super) fn receipt_covers(line: &str, last: &str) -> bool {
    let line = line.strip_suffix('\n').unwrap_or(line);
    if line.contains('\n') {
        return false;
    }
    let words: Vec<&str> = line.split(' ').collect();
    words.first() == Some(&"check-affected")
        && words
            .iter()
            .any(|w| w.strip_prefix("head=") == Some(short(last)))
        && words.last() == Some(&"result=PASS")
        && !words.iter().any(|w| w.starts_with("selection="))
}

/// Writes the summary line of an in-process run where `moai check affected` writes it for its
/// caller (`<out>/summary.txt`), so that [`read_run_receipt`] finds it. The command line does this
/// after `check_affected::run` returns; a direct call of `run` does not.
/// The options of the wide check of a batch: the full selection, and no earlier run is read —
/// the wide check is the safety net (SPEC-20261125-check-affected-rerun REQ-11).
pub(super) fn wide_options() -> crate::check_affected::RunOptions {
    crate::check_affected::RunOptions {
        slot_dir: None,
        max: 4,
        min_free_gb: 50,
        out: None,
        pick: crate::check_affected::Pick::Full,
        integration_ref: None,
        full: true,
    }
}

pub(super) fn record_summary(answer: &Value) {
    if let (Some(out), Some(summary)) = (answer["out"].as_str(), answer["summary"].as_str()) {
        let _ = crate::card::ready_receipt::write_summary(Path::new(out), summary);
    }
}

/// The summary line a wide run left in its private worktree `work` for the join `last`, or
/// `None` when there is none or it is not a small regular file.
pub(super) fn read_run_receipt(work: &Path, last: &str) -> Option<String> {
    let file = work
        .join("target")
        .join(format!("check-affected-{}", short(last)))
        .join("summary.txt");
    let meta = std::fs::symlink_metadata(&file).ok()?;
    if !meta.is_file() || meta.len() > 4096 {
        return None;
    }
    std::fs::read_to_string(&file).ok()
}

/// The trailers of a join commit: its SPEC, the lander's session and no verification kind.
pub(super) fn trailers(project: &Path, session: Option<&str>, spec: &str) -> String {
    let mut text = format!("Moai-Spec: {spec}\n");
    if let (Some(id), Some(key)) = (session, crate::trailers::worktree_key(project)) {
        text.push_str(&format!("Moai-Session: {id}:{key}\n"));
    }
    text.push_str("Moai-Verified: none\n");
    text
}

/// The heads that joined and the heads that did not, each with its reason.
type Joined = (Vec<Candidate>, Vec<(Item, String)>);

/// Join `items` in order onto `tip`. A head the cursor already holds or stands on lands as it is;
/// any other is merged with `git merge-tree --write-tree` and a merge commit that has the cursor
/// as its first parent. The heads that cannot be joined come back with the reason.
fn join(
    project: &Path,
    session: Option<&str>,
    branch: &str,
    tip: &str,
    items: &[Item],
) -> Result<Joined> {
    let mut cursor = tip.to_string();
    let mut out = Vec::new();
    let mut bad = Vec::new();
    for item in items {
        match landed::contained(project, &item.head, &cursor)? {
            Containment::NotACommit => {
                bad.push((
                    item.clone(),
                    format!(
                        "not_a_commit: {} is not a commit of this repository",
                        item.head
                    ),
                ));
                continue;
            }
            Containment::Inside => {
                out.push(Candidate {
                    item: item.clone(),
                    commit: item.head.clone(),
                    joined: false,
                });
                continue;
            }
            Containment::Outside => {}
        }
        if landed::standing(project, &item.head, &cursor)? == Standing::OnTip {
            cursor = item.head.clone();
            out.push(Candidate {
                item: item.clone(),
                commit: item.head.clone(),
                joined: false,
            });
            continue;
        }
        let tree = match super::check::git_text(
            project,
            &[
                "merge-tree",
                "--write-tree",
                "--no-messages",
                &cursor,
                &item.head,
            ],
        ) {
            Ok(text) => text.lines().next().unwrap_or("").trim().to_string(),
            Err(why) => {
                bad.push((item.clone(), format!("join_failed: {why}")));
                continue;
            }
        };
        let message = format!(
            "merge(next-flow): {} into {branch}, head {}\n\n{}",
            item.spec,
            short(&item.head),
            trailers(project, session, &item.spec)
        );
        let commit = super::check::git_text(
            project,
            &[
                "commit-tree",
                &tree,
                "-p",
                &cursor,
                "-p",
                &item.head,
                "-m",
                &message,
            ],
        )
        .map_err(|why| {
            format!(
                "join_failed: the merge commit of {} cannot be made: {why}",
                item.name
            )
        })?
        .trim()
        .to_string();
        cursor = commit.clone();
        out.push(Candidate {
            item: item.clone(),
            commit,
            joined: true,
        });
    }
    Ok((out, bad))
}

/// What one check run gave.
pub(super) struct Run {
    pub(super) exit: i32,
    pub(super) log: PathBuf,
    /// The run passed and left a full-selection PASS summary of the join: its scans are covered.
    pub(super) scans_covered: bool,
}

/// Run the check on `last` in a private worktree under `<project>/target/land-batch/`, which the
/// run removes again. `Err` means the check could not start; an exit code other than 0 is a FAIL.
pub(super) fn run_check(
    project: &Path,
    check: Option<&Path>,
    tip: &str,
    last: &str,
    number: usize,
) -> Result<Run> {
    let base = project.join("target").join("land-batch");
    std::fs::create_dir_all(&base)
        .map_err(|e| format!("check_not_started: cannot create {}: {e}", base.display()))?;
    let work = base.join(format!("{}-{number}", short(last)));
    let log = base.join(format!("{}-{number}.log", short(last)));
    let work_text = work.to_string_lossy().to_string();
    super::check::git_text(project, &["worktree", "add", "--detach", &work_text, last])
        .map_err(|why| format!("check_not_started: the private worktree cannot be made: {why}"))?;
    let range = format!("{tip}..{last}");
    let result = match check {
        Some(program) => {
            let file = std::fs::File::create(&log)
                .map_err(|e| format!("check_not_started: cannot write {}: {e}", log.display()));
            match file {
                Err(e) => Err(e),
                Ok(file) => {
                    let again = file
                        .try_clone()
                        .map_err(|e| format!("check_not_started: {e}"));
                    match again {
                        Err(e) => Err(e),
                        Ok(again) => Command::new(program)
                            .arg(&range)
                            .current_dir(&work)
                            .stdin(Stdio::null())
                            .stdout(file)
                            .stderr(again)
                            .status()
                            .map(|s| s.code().unwrap_or(1))
                            .map_err(|e| {
                                format!("check_not_started: {} cannot run: {e}", program.display())
                            }),
                    }
                }
            }
        }
        None => {
            let opts = wide_options();
            match crate::check_affected::run(&work, &range, &opts) {
                Ok(answer) => {
                    let summary = answer["summary"].as_str().unwrap_or_default();
                    let _ = std::fs::write(&log, format!("{summary}\n"));
                    record_summary(&answer);
                    Ok(i32::from(answer["result"] != "PASS"))
                }
                Err(why) => Err(format!("check_not_started: {why}")),
            }
        }
    };
    // The receipt of the run is read before the worktree goes (SPEC D-11). Only this program's own
    // wide run covers the scans (REQ-20): a replaced check program may write any line.
    let scans_covered = check.is_none()
        && matches!(result, Ok(0))
        && read_run_receipt(&work, last).is_some_and(|line| receipt_covers(&line, last));
    // The worktree is the run's own: git removes it with its checkout.
    let _ = super::check::git_text(project, &["worktree", "remove", "--force", &work_text]);
    let _ = super::check::git_text(project, &["worktree", "prune"]);
    result.map(|exit| Run {
        exit,
        log,
        scans_covered,
    })
}

fn entry(item: &Item, why: Option<&str>) -> Value {
    match why {
        Some(why) => json!({"card": item.name, "head": item.head, "reason": why}),
        None => json!({"card": item.name, "head": item.head}),
    }
}

/// `card land-batch`: see the module comment.
pub fn land_batch(project: &Path, home: &Path, req: &BatchRequest<'_>) -> Result<Value> {
    if req.by != LEADER_BY {
        return Err(refused(format!(
            "only the leader lands a card until the typed verdict exists: give --by {LEADER_BY}"
        )));
    }
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64);
    confirm_leader(project, req.session, now)?;
    let parsed = parse_pairs(req.pairs)?;
    let branch = landed::integration_branch(project).map_err(|why| {
        refused(format!(
            "a landing is made on the integration branch, and the profile gives none ({why})"
        ))
    })?;
    let (mut pending, mut dropped) = attach_specs(project, home, parsed)?;

    let mut landed_cards: Vec<Value> = Vec::new();
    let mut runs: Vec<Value> = Vec::new();
    let mut stopped: Value = Value::Null;
    let mut not_run: Vec<Value> = Vec::new();
    let mut moved_tips = 0usize;

    'outer: while !pending.is_empty() {
        let mut take = pending.len();
        loop {
            let tip = landed::tip_of(project, &branch).map_err(refused)?;
            let set: Vec<Item> = pending[..take].to_vec();
            let (cands, bad) = join(project, req.session, &branch, &tip, &set)?;
            if !bad.is_empty() {
                for (item, why) in &bad {
                    dropped.push(entry(item, Some(why)));
                    pending.retain(|p| p.name != item.name);
                }
                continue 'outer;
            }
            let last = cands.last().map(|c| c.commit.clone()).unwrap_or_default();
            // Heads the tip already holds need no check; a set of them lands as it is.
            let moves = cands.iter().any(|c| landed_moves(project, &c.commit, &tip));
            let mut scans_covered = false;
            let exit = if moves {
                match run_check(project, req.check_cmd, &tip, &last, runs.len() + 1) {
                    Ok(run) => {
                        runs.push(json!({
                            "cards": set.iter().map(|i| i.name.clone()).collect::<Vec<_>>(),
                            "range": format!("{tip}..{last}"),
                            "exit": run.exit,
                            "log": run.log.to_string_lossy(),
                            "scans_covered": run.scans_covered,
                        }));
                        scans_covered = run.scans_covered;
                        run.exit
                    }
                    Err(why) => {
                        stopped = json!({"code": "check_not_started", "why": why});
                        break 'outer;
                    }
                }
            } else {
                0
            };
            if exit != 0 {
                if take == 1 {
                    let item = pending.remove(0);
                    dropped.push(entry(
                        &item,
                        Some(&format!(
                            "check_failed: the check exited {exit} on the join of this head alone"
                        )),
                    ));
                    continue 'outer;
                }
                take = take.div_ceil(2);
                continue;
            }
            // A tip that moved since the run has no say in this result: join and check again.
            if landed::tip_of(project, &branch).map_err(refused)? != tip {
                moved_tips += 1;
                if moved_tips >= MAX_TIP_MOVES {
                    stopped = json!({
                        "code": "tip_kept_moving",
                        "why": format!(
                            "the integration branch moved {moved_tips} times after a check \
                             passed, so no result could be used; nothing landed from those runs: \
                             run the batch again when the branch is quiet"
                        ),
                    });
                    break 'outer;
                }
                continue;
            }
            // Every head is judged before the first one lands: a head that the landing would
            // refuse is dropped now, while the branch is still where it was, and the rest is
            // joined and checked again. Otherwise the branch would stop at a join of the first
            // heads that no check ever ran on. The heads then land in order, each on the way to
            // the join that was checked.
            let mut parent = tip.clone();
            let mut refused_head: Option<(Item, String)> = None;
            for cand in &cands {
                let one = LandRequest {
                    session: req.session,
                    head: &cand.commit,
                    met: &[],
                    generation: None,
                    by: req.by,
                    keep_open: false,
                    audited: cand.joined.then_some(cand.item.head.as_str()),
                };
                match super::land_preflight(
                    project,
                    home,
                    &cand.item.name,
                    &one,
                    &parent,
                    scans_covered,
                ) {
                    Ok(_) => parent = cand.commit.clone(),
                    Err(why) => {
                        refused_head = Some((cand.item.clone(), why));
                        break;
                    }
                }
            }
            if let Some((item, why)) = refused_head {
                dropped.push(entry(&item, Some(&format!("land_refused: {why}"))));
                pending.retain(|p| p.name != item.name);
                continue 'outer;
            }
            for cand in &cands {
                let answer = super::land_in_batch(
                    project,
                    home,
                    &cand.item.name,
                    &LandRequest {
                        session: req.session,
                        head: &cand.commit,
                        met: &[],
                        generation: None,
                        by: req.by,
                        keep_open: false,
                        audited: cand.joined.then_some(cand.item.head.as_str()),
                    },
                    scans_covered,
                );
                match answer {
                    Ok(done) => {
                        let mut e = entry(&cand.item, None);
                        e["candidate"] = json!(cand.commit);
                        e["moved"] = done["moved"].clone();
                        e["closed"] = done["closed"].clone();
                        // A held head stays open: the batch says which verdicts hold it.
                        if let Some(held) = done.get("held") {
                            e["held"] = held.clone();
                            e["open_criteria"] = done["open_criteria"].clone();
                        }
                        landed_cards.push(e);
                        pending.retain(|p| p.name != cand.item.name);
                    }
                    Err(why) => {
                        stopped = json!({
                            "code": "land_refused", "card": cand.item.name, "why": why,
                        });
                        break 'outer;
                    }
                }
            }
            continue 'outer;
        }
    }
    for item in &pending {
        not_run.push(entry(
            item,
            Some("not run: the batch stopped before this head"),
        ));
    }
    Ok(json!({
        "action": "land-batch", "branch": branch,
        "landed": landed_cards, "dropped": dropped, "not_run": not_run,
        "runs": runs, "checks_run": runs.len(), "stopped": stopped,
        "writes": !landed_cards.is_empty(),
    }))
}

/// Whether landing `commit` on `tip` changes the branch (it is not inside the tip).
fn landed_moves(project: &Path, commit: &str, tip: &str) -> bool {
    !matches!(
        landed::contained(project, commit, tip),
        Ok(Containment::Inside)
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    const LAST: &str = "da585ae4d507bec047ab5a6446bda6e367649282";

    /// SPEC-20261125-check-affected-rerun AC-6: the wide check of a batch reads no earlier run.
    #[test]
    fn the_wide_check_of_a_batch_runs_full_and_reads_no_earlier_run() {
        let opts = wide_options();
        assert!(opts.full, "land-batch must not narrow its wide check");
        assert!(matches!(opts.pick, crate::check_affected::Pick::Full));
    }

    /// SPEC-20261124-card-select-batch-land REQ-24: the `Verify` section of `CLAUDE.md` and the
    /// flow reference say that the READY receipt is the card set, that the wide set runs once per
    /// batch in `card land-batch`, and that the full run every five landings stays. The texts are
    /// compiled in, so the check reads no file at run time.
    #[test]
    fn the_verify_section_and_the_flow_reference_name_the_new_order() {
        let docs = [
            ("CLAUDE.md", include_str!("../../../CLAUDE.md")),
            (
                "next-flow.md",
                include_str!("../../../skills/moai/references/flows/next-flow.md"),
            ),
        ];
        for (file, text) in docs {
            for needle in [
                "check affected --card",
                "card land-batch",
                "every five landings",
            ] {
                assert!(text.contains(needle), "{file} does not say `{needle}`");
            }
        }
    }

    const OTHER: &str = "0123456789abcdef0123456789abcdef01234567";

    /// The summary line of a finished run of `head`, made by the writer of `check affected` itself
    /// (`summary_line`), so the format the batch reads cannot drift from the format that is written.
    fn made(pick: crate::check_affected::Pick, head: &str, pass: bool) -> String {
        let results: Vec<(String, i32)> = ["diff_check", "fmt", "clippy", "test"]
            .iter()
            .map(|label| {
                (
                    label.to_string(),
                    i32::from(!pass && *label == "test") * 100,
                )
            })
            .collect();
        crate::check_affected::summary_line(&crate::check_affected::SummaryParts {
            base: "b0bffbbd48cc3aae17ac595f49ce10d6e3bfb011",
            head,
            targets: 150,
            scans: "listed:135",
            pick,
            threads: "5",
            results: &results,
            stopped_at: None,
            pass,
        })
    }

    #[test]
    fn a_full_pass_summary_of_the_join_covers_the_scans() {
        use crate::check_affected::Pick;
        let full = made(Pick::Full, LAST, true);
        assert!(!full.contains("selection="), "{full}");
        assert!(receipt_covers(&full, LAST), "{full}");
        assert!(receipt_covers(&format!("{full}\n"), LAST));
    }

    #[test]
    fn a_part_run_a_failure_another_head_or_a_planted_line_covers_nothing() {
        use crate::check_affected::Pick;
        for (why, text) in [
            ("a --card run", made(Pick::Card, LAST, true)),
            ("a --lane run", made(Pick::Lane, LAST, true)),
            ("a failure", made(Pick::Full, LAST, false)),
            ("another head", made(Pick::Full, OTHER, true)),
            (
                "two lines",
                format!("{}\n{}", made(Pick::Full, LAST, true), "x"),
            ),
            ("not a summary", "result=PASS".to_string()),
        ] {
            assert!(!receipt_covers(&text, LAST), "{why}");
        }
    }

    /// Audit B1 of the follow-up: the in-process run of the batch returns its answer without
    /// writing `summary.txt` (the command line does), so the batch records the line itself, and
    /// what it reads back from the private worktree is the line the real writer made.
    #[test]
    fn the_summary_of_the_in_process_run_is_recorded_and_read_back_as_the_cover() {
        use crate::check_affected::Pick;
        let work = tempfile::tempdir().unwrap();
        let out = work
            .path()
            .join("target")
            .join(format!("check-affected-{}", short(LAST)));
        std::fs::create_dir_all(&out).unwrap();
        assert!(read_run_receipt(work.path(), LAST).is_none());
        let answer =
            |pick, pass| json!({"out": out.to_string_lossy(), "summary": made(pick, LAST, pass)});
        record_summary(&answer(Pick::Full, true));
        let read = read_run_receipt(work.path(), LAST).expect("the summary was recorded");
        assert!(receipt_covers(&read, LAST), "{read}");
        // A later part run or failure overwrites it and covers nothing.
        for (pick, pass) in [(Pick::Card, true), (Pick::Full, false)] {
            record_summary(&answer(pick, pass));
            let read = read_run_receipt(work.path(), LAST).unwrap();
            assert!(!receipt_covers(&read, LAST), "{read}");
        }
        // An answer with no `out` or no summary records nothing and does not fail.
        record_summary(&json!({"summary": "x"}));
        record_summary(&json!({"out": out.to_string_lossy()}));
    }

    /// The default check of the batch runs `check_affected::run` in process and records its
    /// summary right after (a real run needs a Rust project and the whole test set, which no unit
    /// test starts), so the call is pinned in the text: the line after the run is the record.
    #[test]
    fn the_default_check_records_the_summary_of_its_run() {
        let source = include_str!("batch.rs");
        let lines: Vec<&str> = source.lines().collect();
        let run = lines
            .iter()
            .position(|l| l.contains("crate::check_affected::run(&work, &range, &opts)"))
            .expect("the default check runs check affected in process");
        let after = lines[run + 1..run + 8].join("\n");
        assert!(
            after.contains("record_summary(&answer);"),
            "the summary of the run is not recorded:\n{after}"
        );
    }

    /// Audit B1 of the follow-up: the test seam for the check program is read only by a debug
    /// build. The line that reads `MOAI_TEST_LAND_BATCH_CHECK` stands under `cfg(debug_assertions)`
    /// and a release build takes `None` (the text is compiled in, so no file is read at run time).
    #[test]
    fn the_check_program_seam_is_read_by_debug_builds_only() {
        let source = include_str!("../../ops/card.rs");
        let reads: Vec<usize> = source
            .lines()
            .enumerate()
            .filter(|(_, l)| l.contains("env(\"MOAI_TEST_LAND_BATCH_CHECK\")"))
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

    #[test]
    fn no_leader_entry_and_no_session_id_confirm_nothing() {
        let dir = tempfile::tempdir().unwrap();
        let project = dir.path();
        let git = |args: &[&str]| {
            let mut cmd = std::process::Command::new("git");
            cmd.arg("-C").arg(project).args(args);
            crate::identity::scrub_git_env(&mut cmd);
            let out = cmd.output().unwrap();
            assert!(out.status.success(), "git {args:?}");
        };
        git(&["init", "-q", "-b", "main"]);
        let me = "dddddddd-0000-4000-8000-00000000000d";
        // No entry, no session id: refused.
        assert!(confirm_leader(project, Some(me), 100).is_err());
        assert!(confirm_leader(project, None, 100).is_err());
    }

    /// A throwaway repository with a lane registry that holds `entries` (file name and JSON text).
    fn registry(entries: &[(&str, String)]) -> tempfile::TempDir {
        let dir = tempfile::tempdir().unwrap();
        let mut cmd = std::process::Command::new("git");
        cmd.arg("-C")
            .arg(dir.path())
            .args(["init", "-q", "-b", "main"]);
        crate::identity::scrub_git_env(&mut cmd);
        assert!(cmd.output().unwrap().status.success());
        let sessions = crate::clone_root::path(&dir.path().join(".git")).join("sessions");
        std::fs::create_dir_all(&sessions).unwrap();
        for (file, text) in entries {
            std::fs::write(sessions.join(file), text).unwrap();
        }
        dir
    }

    /// The text of a lane registry entry of the session `session` with the lease `lease`.
    fn entry_text(name: &str, mode: &str, session: &str, lease: i64) -> String {
        json!({
            "schema_version": 1, "name": name, "instance": "1".repeat(32),
            "lease_until": lease, "registered_at": 0, "mode": mode, "session": session,
        })
        .to_string()
    }

    const ME: &str = "dddddddd-0000-4000-8000-00000000000d";
    const LANE: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

    /// The four causes of the gate, as the text of a refusal of the command `what` (`prefix` is
    /// what the command's refusals start with, `words` what the leader alone may do).
    fn four_causes(prefix: &str, words: &str, call: impl Fn(&Path, Option<&str>) -> Result<()>) {
        let tail = "; nothing was changed";
        let live = entry_text("leader", "leader", ME, 1_000);
        let ok = registry(&[("leader.json", live.clone())]);
        assert!(call(ok.path(), Some(ME)).is_ok(), "the live leader passes");
        assert_eq!(
            call(ok.path(), None).unwrap_err(),
            format!(
                "{prefix}this call has no session id, and only the registered leader session \
                 {words}{tail}"
            )
        );
        assert_eq!(
            call(ok.path(), Some(LANE)).unwrap_err(),
            format!("{prefix}this session is not the session of the registered leader entry{tail}")
        );
        let none_live = "no live leader entry is registered: the leader window registers itself \
                         with `moai session-register . --name leader --mode leader` \
                         (`/moai leader`)";
        let nobody = registry(&[]);
        assert_eq!(
            call(nobody.path(), Some(ME)).unwrap_err(),
            format!("{prefix}{none_live}{tail}")
        );
        let ended = registry(&[("leader.json", entry_text("leader", "leader", ME, 100))]);
        assert_eq!(
            call(ended.path(), Some(ME)).unwrap_err(),
            format!("{prefix}{none_live}{tail}")
        );
        let lane_mode = registry(&[("leader.json", entry_text("leader", "agent", ME, 1_000))]);
        assert_eq!(
            call(lane_mode.path(), Some(ME)).unwrap_err(),
            format!("{prefix}{none_live}{tail}")
        );
        // A registry that is a file, not a folder, cannot be read: a refusal, never a pass.
        let broken = registry(&[]);
        let sessions = crate::clone_root::path(&broken.path().join(".git")).join("sessions");
        std::fs::remove_dir(&sessions).unwrap();
        std::fs::write(&sessions, "x").unwrap();
        let text = call(broken.path(), Some(ME)).unwrap_err();
        let (head, rest) = (
            format!("{prefix}the lane registry cannot be read ("),
            format!("){tail}"),
        );
        assert!(
            text.starts_with(&head) && text.ends_with(&rest),
            "unreadable registry: {text}"
        );
    }

    /// The words of `card land-batch` stay as they were, through the wrapper and through the new
    /// function (SPEC-20261125-land-sync-merge REQ-1: the batch wording is unchanged).
    #[test]
    fn the_four_refusals_of_the_batch_keep_their_words() {
        let prefix = "card land-batch refused: ";
        four_causes(prefix, "lands a batch", |p, s| confirm_leader(p, s, 500));
        four_causes(prefix, "lands a batch", |p, s| {
            leader_reason(p, s, 500, "land-batch")
        });
    }

    /// The same four causes name `card land-sync` and what only the leader does there; the text
    /// never says `land-batch`.
    #[test]
    fn the_four_refusals_of_the_sync_name_the_sync_command() {
        let prefix = "card land-sync refused: ";
        four_causes(prefix, "syncs a branch", |p, s| {
            leader_reason(p, s, 500, "land-sync")
        });
        let dir = registry(&[]);
        let text = leader_reason(dir.path(), None, 500, "land-sync").unwrap_err();
        assert!(!text.contains("land-batch"), "{text}");
    }
}
