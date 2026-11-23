//! What starts a worktree cleanup (SPEC-20261123-worktree-slots REQ-11, REQ-12;
//! SPEC-20261123-landing-cleanup-real-clock).
//!
//! The cleanup itself is [`crate::worktree_gc`]: this module only names which worktrees to
//! hand it and runs its plan, then its apply on the plan's own digest. It never widens what the
//! plan may remove, so the safe criteria, the preserved `.claude/worktrees/` folder and the
//! fixed slots keep their effect.
//!
//! Two things start it:
//!
//! - a landing: [`run`] with [`Trigger::Landing`], called by `card land` after the card closed,
//!   for the worktrees the card's reservations named. The card is closed, so every reservation of
//!   it has ended: for those worktrees the plan drops the recent-change window and the
//!   empty-worktree guard (a lane lands within hours of its last change, and a closed card never
//!   comes back to this trigger). A claim of another card, a process inside, a dirty or unmerged
//!   tree and the other reasons still keep a worktree;
//! - a verdict: [`run`] with [`Trigger::Verdict`], the entry the verdict step calls for the
//!   audited card (it is not wired to a command yet).
//!
//! Both also read the queue the `reference-transaction` hook fills ([`QUEUE_FILE`]): one line
//! per deleted branch. A line is a hint. The ref is read again, a ref that exists is dropped,
//! and for a ref that is gone the worktrees of the cards that held that branch are added. Those
//! worktrees keep the normal window. A line whose worktree was skipped for a temporary reason
//! (`TEMPORARY`) is written back, so a later run sees it again; every other line is used up.
//!
//! The queue is taken by a rename, so a hook that appends meanwhile writes a new queue. A taken
//! file that a run left behind (it died before it finished) is read by the next run. A line a
//! hook wrote into the taken file after it was read goes back to the queue.
//!
//! A cleanup never fails the call that started it: [`run`] answers a value, and a problem is its
//! `error` key.

use crate::ledger::reservation;
use crate::worktree_gc::{self, Pace};
use crate::worktree_slot;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::time::Duration;

type Result<T> = std::result::Result<T, String>;

/// The queue the hook appends to, relative to the common git folder.
pub const QUEUE_FILE: &str = "moai/worktree-cleanup-queue";
/// Test seam: the current time in Unix seconds instead of the clock.
pub const NOW_SEAM: &str = "MOAI_WORKTREE_CLEANUP_NOW_UNIX";
/// Name prefix of the file a run takes the queue into.
const DRAINING_PREFIX: &str = "worktree-cleanup-queue.draining-";
/// Reasons that pass by themselves: time, a process that ends, a claim that is released. A queue
/// line whose worktree was skipped only for reasons of this list goes back to the queue; a dirty
/// or unmerged worktree, a preserved or fixed one and the rest are lasting, and their line is
/// used up, also when a temporary reason holds beside them.
const TEMPORARY: &[&str] = &[
    "recent_change",
    "empty_worktree",
    "process_in_use",
    "process_check_unavailable",
    "file_open_in_use",
    "open_file_check_unavailable",
    "live_claim",
    "claim_check_unavailable",
    "budget_spent",
];

/// What started the cleanup.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Trigger {
    Landing,
    Verdict,
}

impl Trigger {
    fn word(self) -> &'static str {
        match self {
            Trigger::Landing => "landing",
            Trigger::Verdict => "verdict",
        }
    }
}

/// One cleanup: the card whose worktrees to clean and the integration ref the plan measures
/// "merged" against.
pub struct Request<'a> {
    pub card_id: i64,
    pub integration_ref: &'a str,
    pub trigger: Trigger,
}

/// Runs the cleanup for the card and the queue. Never fails: the answer says what was planned
/// and what was done, or carries an `error`.
pub fn run(project: &Path, home: &Path, req: &Request<'_>) -> Value {
    match try_run(project, home, req) {
        Ok(v) => v,
        Err(error) => json!({
            "trigger": req.trigger.word(), "applied": false, "worktrees": [], "error": error,
        }),
    }
}

/// True when the answer of [`run`] says it removed or trimmed something.
pub fn wrote(answer: &Value) -> bool {
    answer["worktrees"].as_array().is_some_and(|list| {
        list.iter()
            .any(|w| matches!(w["status"].as_str(), Some("removed" | "trimmed")))
    })
}

fn now_seam() -> Option<u64> {
    std::env::var(NOW_SEAM).ok()?.trim().parse().ok()
}

/// The queue taken out of the way, so a hook appending now writes a new queue and loses no line.
struct Taken {
    queue: PathBuf,
    /// Every taken file, with the number of bytes of it that were read.
    files: Vec<(PathBuf, usize)>,
    /// What was read, kept as it was written, for a run that does not finish.
    text: String,
    refs: Vec<String>,
}

fn lines_of(bytes: &[u8]) -> String {
    let mut text = String::from_utf8_lossy(bytes).into_owned();
    if !text.is_empty() && !text.ends_with('\n') {
        text.push('\n');
    }
    text
}

/// Puts `text` at the end of the queue file, creating it. It never replaces the file, so a queue a
/// hook created in the meantime keeps its lines.
fn append_to(queue: &Path, text: &str) -> std::io::Result<()> {
    use std::io::Write;
    if let Some(dir) = queue.parent() {
        fs::create_dir_all(dir)?;
    }
    fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(queue)?
        .write_all(text.as_bytes())
}

/// A taken file that could not be read goes back under the queue name, unless a queue exists
/// again: a link cannot replace it. Then the file stays as it is, and the next run reads it.
fn restore_unread(taken: &Path, queue: &Path) {
    if fs::hard_link(taken, queue).is_ok() {
        let _ = fs::remove_file(taken);
    }
}

fn take_queue(common: &Path) -> Result<Option<Taken>> {
    let queue = common.join(QUEUE_FILE);
    let dir = queue.parent().unwrap_or(common).to_path_buf();
    let mut files: Vec<(PathBuf, usize)> = Vec::new();
    let mut text = String::new();

    // Files a run left behind: it died between taking the queue and finishing.
    let mut left: Vec<PathBuf> = fs::read_dir(&dir)
        .map(|rd| {
            rd.flatten()
                .filter(|e| e.file_name().to_string_lossy().starts_with(DRAINING_PREFIX))
                .map(|e| e.path())
                .filter(|p| fs::symlink_metadata(p).is_ok_and(|m| m.is_file()))
                .collect()
        })
        .unwrap_or_default();
    left.sort();
    for path in left {
        let bytes = fs::read(&path).map_err(|e| {
            format!("a cleanup queue file left by an earlier run could not be read: {e}")
        })?;
        text.push_str(&lines_of(&bytes));
        files.push((path, bytes.len()));
    }

    // A name no other run uses, so a taken file is never replaced.
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_nanos());
    let own = dir.join(format!("{DRAINING_PREFIX}{}-{nanos}", std::process::id()));
    match fs::rename(&queue, &own) {
        Ok(()) => match fs::read(&own) {
            Ok(bytes) => {
                text.push_str(&lines_of(&bytes));
                files.push((own, bytes.len()));
            }
            Err(e) => {
                restore_unread(&own, &queue);
                return Err(format!("the cleanup queue could not be read: {e}"));
            }
        },
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(e) => return Err(format!("the cleanup queue could not be taken: {e}")),
    }
    if files.is_empty() {
        return Ok(None);
    }
    let mut refs: Vec<String> = text
        .lines()
        .map(str::trim)
        .filter(|l| l.starts_with("refs/heads/") && l.len() > "refs/heads/".len())
        .map(str::to_string)
        .collect();
    refs.sort();
    refs.dedup();
    Ok(Some(Taken {
        queue,
        files,
        text,
        refs,
    }))
}

impl Taken {
    /// The run is over. `keep` names the lines to put back (`None`: the run did not finish, so
    /// every line goes back); a line a hook wrote into a taken file after it was read goes back
    /// too. The taken files are removed only after the queue holds what goes back; if that write
    /// fails they stay, and the next run reads them.
    fn finish(self, keep: Option<&[String]>) {
        let mut back = match keep {
            None => self.text.clone(),
            Some(lines) => lines.iter().map(|l| format!("{l}\n")).collect(),
        };
        for (path, read) in &self.files {
            if let Ok(now) = fs::read(path) {
                if now.len() > *read {
                    back.push_str(&lines_of(&now[*read..]));
                }
            }
        }
        if !back.is_empty() && append_to(&self.queue, &back).is_err() {
            return;
        }
        for (path, _) in &self.files {
            let _ = fs::remove_file(path);
        }
    }
}

fn try_run(project: &Path, home: &Path, req: &Request<'_>) -> Result<Value> {
    let common = worktree_slot::common_dir(project)?;
    let primary = worktree_slot::primary_of(&common)?;
    let taken = take_queue(&common)?;
    let outcome = clean(
        project,
        home,
        req,
        &primary,
        taken.as_ref().map(|t| &t.refs[..]),
    );
    match (taken, outcome) {
        (Some(t), Ok((answer, keep))) => {
            t.finish(Some(&keep));
            Ok(answer)
        }
        (Some(t), Err(e)) => {
            t.finish(None);
            Err(e)
        }
        (None, outcome) => outcome.map(|(answer, _)| answer),
    }
}

/// The answer and the queue lines to put back.
fn clean(
    project: &Path,
    home: &Path,
    req: &Request<'_>,
    primary: &Path,
    queued: Option<&[String]>,
) -> Result<(Value, Vec<String>)> {
    let id = crate::project_id::read(project)?.ok_or("the project has no profile, so no ledger")?;
    let read = crate::ledger::open_read_only(home, &id).map_err(|e| e.to_string())?;
    let conn = read.conn();
    let own: Vec<String> =
        reservation::worktrees_of_card(conn, req.card_id).map_err(|e| e.to_string())?;
    let mut wanted: Vec<String> = own.clone();
    // The queue: a ref that is still there was only a pack or a move, so its line is dropped.
    let mut gone = Vec::new();
    let mut still_there = Vec::new();
    // For each ref that is gone, the worktrees its cards held.
    let mut held_by: Vec<(String, Vec<String>)> = Vec::new();
    for full in queued.unwrap_or_default() {
        let exists = worktree_gc::git_out(
            project,
            &["rev-parse", "--verify", "--quiet", "--end-of-options", full],
        )
        .is_some();
        if exists {
            still_there.push(full.clone());
            continue;
        }
        gone.push(full.clone());
        let short = full.strip_prefix("refs/heads/").unwrap_or(full);
        let mut worktrees = Vec::new();
        for card in reservation::cards_of_branch(conn, short).map_err(|e| e.to_string())? {
            worktrees
                .extend(reservation::worktrees_of_card(conn, card).map_err(|e| e.to_string())?);
        }
        wanted.extend(worktrees.iter().cloned());
        held_by.push((full.clone(), worktrees));
    }
    wanted.sort();
    wanted.dedup();

    // Only a registered linked worktree of this clone goes to the plan; the rest are named.
    let linked = worktree_slot::linked(primary)?;
    let registered = |w: &str| -> Option<PathBuf> {
        let p = PathBuf::from(w);
        linked
            .iter()
            .find(|e| worktree_slot::same(&e.path, &p))
            .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
    };
    let mut paths = Vec::new();
    let mut unregistered = Vec::new();
    for w in &wanted {
        match registered(w) {
            Some(p) => paths.push(p),
            None => unregistered.push(w.clone()),
        }
    }
    paths.sort();
    paths.dedup();
    // A landing closed the card: its own worktrees are settled (no reservation of it is left).
    let mut settled: Vec<PathBuf> = if req.trigger == Trigger::Landing {
        own.iter().filter_map(|w| registered(w)).collect()
    } else {
        Vec::new()
    };
    settled.sort();
    settled.dedup();

    let mut answer = json!({
        "trigger": req.trigger.word(),
        "integration_ref": req.integration_ref,
        "queue": {"gone": gone, "still_there": still_there},
        "not_a_worktree": unregistered,
        "worktrees": [],
        "plan_digest": Value::Null,
        "applied": false,
        "error": Value::Null,
    });
    if paths.is_empty() {
        return Ok((answer, Vec::new()));
    }

    let recent = Duration::from_secs(worktree_gc::DEFAULT_RECENT_MINUTES * 60);
    let budget = Duration::from_millis(worktree_gc::DEFAULT_BUDGET_MS);
    let now = now_seam();
    let pace = Pace {
        only: paths.clone(),
        ..Pace::default()
    };
    let plan = worktree_gc::plan_settled(
        project,
        req.integration_ref,
        recent,
        budget,
        now,
        false,
        &pace,
        &settled,
    )?;
    let digest = plan["plan_digest"]
        .as_str()
        .ok_or("the plan carries no digest")?
        .to_string();
    let candidates = plan["candidates"].as_array().cloned().unwrap_or_default();
    let acts = candidates
        .iter()
        .any(|c| matches!(c["action"].as_str(), Some("remove" | "trim")));
    let mut results: Vec<Value> = Vec::new();
    if acts {
        // The apply recomputes the plan and refuses unless its digest is the one just read.
        let applied = worktree_gc::apply_settled(
            project,
            req.integration_ref,
            recent,
            budget,
            now,
            &pace,
            &settled,
            &digest,
            &format!("worktree-cleanup {} C-{}", req.trigger.word(), req.card_id),
        )?;
        results = applied["results"].as_array().cloned().unwrap_or_default();
    }
    answer["worktrees"] = Value::Array(
        candidates
            .iter()
            .map(|c| {
                let path = c["path"].clone();
                let done = results.iter().find(|r| r["path"] == path);
                json!({
                    "path": path,
                    "action": c["action"],
                    "reasons": c["reasons"],
                    "status": done.map_or(json!("not_acted"), |r| r["status"].clone()),
                    "detail": done.map_or(Value::Null, |r| r["detail"].clone()),
                })
            })
            .collect(),
    );
    answer["plan_digest"] = json!(digest);
    answer["applied"] = json!(acts);

    // A queue line goes back when one of its worktrees was skipped for a temporary reason.
    let primary_real = fs::canonicalize(primary).unwrap_or_else(|_| primary.to_path_buf());
    let temporary = |path: &Path| -> bool {
        let shown = json!(worktree_gc::shown(path, &primary_real));
        let skipped = candidates.iter().any(|c| {
            c["path"] == shown
                && c["action"] == "skip"
                // One lasting reason keeps the worktree whatever else holds, so every reason
                // must be a temporary one.
                && c["reasons"].as_array().is_some_and(|r| {
                    !r.is_empty() && r.iter().all(|x| TEMPORARY.iter().any(|t| x == t))
                })
        });
        let refused = results.iter().any(|r| {
            r["path"] == shown
                && r["status"] != "removed"
                && r["status"] != "trimmed"
                && r["detail"]
                    .as_str()
                    .is_some_and(|d| TEMPORARY.iter().any(|t| d.contains(t)))
        });
        skipped || refused
    };
    let keep: Vec<String> = held_by
        .iter()
        .filter(|(_, worktrees)| {
            worktrees
                .iter()
                .filter_map(|w| registered(w))
                .any(|p| temporary(&p))
        })
        .map(|(full, _)| full.clone())
        .collect();
    Ok((answer, keep))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn common() -> (tempfile::TempDir, PathBuf) {
        let dir = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(dir.path()).unwrap();
        fs::create_dir_all(root.join("moai")).unwrap();
        (dir, root)
    }

    fn names(common: &Path) -> Vec<String> {
        let mut v: Vec<String> = fs::read_dir(common.join("moai"))
            .unwrap()
            .flatten()
            .map(|e| e.file_name().to_string_lossy().to_string())
            .collect();
        v.sort();
        v
    }

    /// A line a hook wrote into the taken file after it was read goes back to the queue, and the
    /// taken file goes.
    #[test]
    fn a_line_written_after_the_read_goes_back_to_the_queue() {
        let (_keep, root) = common();
        let queue = root.join(QUEUE_FILE);
        fs::write(&queue, "refs/heads/a\n").unwrap();
        let taken = take_queue(&root).unwrap().unwrap();
        assert_eq!(taken.refs, vec!["refs/heads/a".to_string()]);
        // The hook had the file open before the rename and writes now.
        let (file, _) = &taken.files[0];
        let mut f = fs::OpenOptions::new().append(true).open(file).unwrap();
        std::io::Write::write_all(&mut f, b"refs/heads/late\n").unwrap();
        taken.finish(Some(&[]));
        assert_eq!(fs::read_to_string(&queue).unwrap(), "refs/heads/late\n");
        assert_eq!(names(&root), vec!["worktree-cleanup-queue"]);
    }

    /// A run that did not finish puts back every line, the late one too, and leaves no taken file.
    #[test]
    fn a_run_that_did_not_finish_puts_back_every_line() {
        let (_keep, root) = common();
        let queue = root.join(QUEUE_FILE);
        fs::write(&queue, "refs/heads/a\nrefs/heads/b\n").unwrap();
        let taken = take_queue(&root).unwrap().unwrap();
        let (file, _) = &taken.files[0];
        let mut f = fs::OpenOptions::new().append(true).open(file).unwrap();
        std::io::Write::write_all(&mut f, b"refs/heads/late\n").unwrap();
        taken.finish(None);
        assert_eq!(
            fs::read_to_string(&queue).unwrap(),
            "refs/heads/a\nrefs/heads/b\nrefs/heads/late\n"
        );
        assert_eq!(names(&root), vec!["worktree-cleanup-queue"]);
    }

    /// The lines put back are appended: a queue a hook created since keeps its own lines.
    #[test]
    fn lines_put_back_never_replace_a_new_queue() {
        let (_keep, root) = common();
        let queue = root.join(QUEUE_FILE);
        fs::write(&queue, "refs/heads/a\n").unwrap();
        let taken = take_queue(&root).unwrap().unwrap();
        fs::write(&queue, "refs/heads/new\n").unwrap();
        taken.finish(None);
        assert_eq!(
            fs::read_to_string(&queue).unwrap(),
            "refs/heads/new\nrefs/heads/a\n"
        );
    }

    /// A taken file an earlier run left is read with the queue and removed after it.
    #[test]
    fn a_leftover_taken_file_is_read_and_removed() {
        let (_keep, root) = common();
        let left = root.join(format!("{QUEUE_FILE}.draining-7"));
        fs::write(&left, "refs/heads/old\n").unwrap();
        fs::write(root.join(QUEUE_FILE), "refs/heads/a\n").unwrap();
        let taken = take_queue(&root).unwrap().unwrap();
        assert_eq!(
            taken.refs,
            vec!["refs/heads/a".to_string(), "refs/heads/old".to_string()]
        );
        taken.finish(Some(&[]));
        assert!(names(&root).is_empty(), "{:?}", names(&root));
    }

    /// With only a leftover file and no queue, the leftover alone is read.
    #[test]
    fn a_leftover_taken_file_without_a_queue_is_read() {
        let (_keep, root) = common();
        fs::write(
            root.join(format!("{QUEUE_FILE}.draining-7")),
            "refs/heads/old\n",
        )
        .unwrap();
        let taken = take_queue(&root).unwrap().unwrap();
        assert_eq!(taken.refs, vec!["refs/heads/old".to_string()]);
        taken.finish(Some(&["refs/heads/old".to_string()]));
        assert_eq!(
            fs::read_to_string(root.join(QUEUE_FILE)).unwrap(),
            "refs/heads/old\n"
        );
        assert_eq!(names(&root), vec!["worktree-cleanup-queue"]);
    }

    /// A taken file that cannot be restored under the queue name because a queue exists again
    /// stays, and the queue is not touched.
    #[test]
    fn an_unread_file_does_not_replace_a_new_queue() {
        let (_keep, root) = common();
        let queue = root.join(QUEUE_FILE);
        let taken = root.join(format!("{QUEUE_FILE}.draining-7"));
        fs::write(&taken, "refs/heads/a\n").unwrap();
        fs::write(&queue, "refs/heads/new\n").unwrap();
        restore_unread(&taken, &queue);
        assert_eq!(fs::read_to_string(&queue).unwrap(), "refs/heads/new\n");
        assert_eq!(fs::read_to_string(&taken).unwrap(), "refs/heads/a\n");
    }

    /// With no queue the taken file goes back under the queue name.
    #[test]
    fn an_unread_file_goes_back_when_no_queue_exists() {
        let (_keep, root) = common();
        let queue = root.join(QUEUE_FILE);
        let taken = root.join(format!("{QUEUE_FILE}.draining-7"));
        fs::write(&taken, "refs/heads/a\n").unwrap();
        restore_unread(&taken, &queue);
        assert_eq!(fs::read_to_string(&queue).unwrap(), "refs/heads/a\n");
        assert_eq!(names(&root), vec!["worktree-cleanup-queue"]);
    }

    /// A queue nobody can read is restored under its name, with nothing lost.
    #[cfg(unix)]
    #[test]
    fn an_unreadable_queue_is_restored_under_its_name() {
        use std::os::unix::fs::PermissionsExt;
        let (_keep, root) = common();
        let queue = root.join(QUEUE_FILE);
        fs::write(&queue, "refs/heads/a\n").unwrap();
        fs::set_permissions(&queue, fs::Permissions::from_mode(0o000)).unwrap();
        let got = take_queue(&root);
        fs::set_permissions(&queue, fs::Permissions::from_mode(0o600)).unwrap();
        assert!(got.is_err());
        assert_eq!(fs::read_to_string(&queue).unwrap(), "refs/heads/a\n");
        assert_eq!(names(&root), vec!["worktree-cleanup-queue"]);
    }

    #[test]
    fn wrote_counts_a_removal_or_a_trim_and_nothing_else() {
        let none = json!({"worktrees": [{"status": "not_acted"}, {"status": "refused"}]});
        assert!(!wrote(&none));
        assert!(!wrote(&json!({"worktrees": []})));
        assert!(wrote(
            &json!({"worktrees": [{"status": "not_acted"}, {"status": "removed"}]})
        ));
        assert!(wrote(&json!({"worktrees": [{"status": "trimmed"}]})));
    }
}
