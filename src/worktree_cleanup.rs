//! What starts a worktree cleanup (SPEC-20261123-worktree-slots REQ-11, REQ-12).
//!
//! The cleanup itself is [`crate::worktree_gc`]: this module only names which worktrees to
//! hand it and runs its plan, then its apply on the plan's own digest. It never widens what the
//! plan may remove, so the safe criteria, the preserved `.claude/worktrees/` folder and the
//! fixed slots keep their effect.
//!
//! Two things start it:
//!
//! - a landing: [`run`] with [`Trigger::Landing`], called by `card land` after the card closed,
//!   for the worktrees the card's reservations named;
//! - a verdict: [`run`] with [`Trigger::Verdict`], the entry the verdict step calls for the
//!   audited card (it is not wired to a command yet).
//!
//! Both also read the queue the `reference-transaction` hook fills ([`QUEUE_FILE`]): one line
//! per deleted branch. A line is a hint. The ref is read again, a ref that exists is dropped,
//! and for a ref that is gone the worktrees of the cards that held that branch are added.
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

fn now_seam() -> Option<u64> {
    std::env::var(NOW_SEAM).ok()?.trim().parse().ok()
}

/// The queue taken out of the way, so a hook appending now writes a new queue and loses no line.
struct Taken {
    draining: PathBuf,
    queue: PathBuf,
    refs: Vec<String>,
}

fn take_queue(common: &Path) -> Result<Option<Taken>> {
    let queue = common.join(QUEUE_FILE);
    let draining = common.join(format!("{QUEUE_FILE}.draining-{}", std::process::id()));
    match fs::rename(&queue, &draining) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(e) => return Err(format!("the cleanup queue could not be taken: {e}")),
    }
    let text = match fs::read_to_string(&draining) {
        Ok(t) => t,
        Err(e) => {
            // Put it back where the hook writes, so no line is lost.
            let _ = fs::rename(&draining, &queue);
            return Err(format!("the cleanup queue could not be read: {e}"));
        }
    };
    let mut refs: Vec<String> = text
        .lines()
        .map(str::trim)
        .filter(|l| l.starts_with("refs/heads/") && l.len() > "refs/heads/".len())
        .map(str::to_string)
        .collect();
    refs.sort();
    refs.dedup();
    Ok(Some(Taken {
        draining,
        queue,
        refs,
    }))
}

impl Taken {
    /// The lines are used up: the taken file goes.
    fn done(self) {
        let _ = fs::remove_file(&self.draining);
    }

    /// The cleanup did not finish: the taken file is appended back to the queue, so nothing is
    /// lost, and goes.
    fn give_back(self) {
        use std::io::Write;
        if let Ok(text) = fs::read(&self.draining) {
            if let Ok(mut f) = fs::OpenOptions::new()
                .create(true)
                .append(true)
                .open(&self.queue)
            {
                let _ = f.write_all(&text);
            }
        }
        let _ = fs::remove_file(&self.draining);
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
    match (taken, &outcome) {
        (Some(t), Ok(_)) => t.done(),
        (Some(t), Err(_)) => t.give_back(),
        (None, _) => {}
    }
    outcome
}

fn clean(
    project: &Path,
    home: &Path,
    req: &Request<'_>,
    primary: &Path,
    queued: Option<&[String]>,
) -> Result<Value> {
    let id = crate::project_id::read(project)?.ok_or("the project has no profile, so no ledger")?;
    let read = crate::ledger::open_read_only(home, &id).map_err(|e| e.to_string())?;
    let conn = read.conn();
    let mut wanted: Vec<String> =
        reservation::worktrees_of_card(conn, req.card_id).map_err(|e| e.to_string())?;
    // The queue: a ref that is still there was only a pack or a move, so its line is dropped.
    let mut gone = Vec::new();
    let mut still_there = Vec::new();
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
        for card in reservation::cards_of_branch(conn, short).map_err(|e| e.to_string())? {
            wanted.extend(reservation::worktrees_of_card(conn, card).map_err(|e| e.to_string())?);
        }
    }
    wanted.sort();
    wanted.dedup();

    // Only a registered linked worktree of this clone goes to the plan; the rest are named.
    let linked = worktree_slot::linked(primary)?;
    let mut paths = Vec::new();
    let mut unregistered = Vec::new();
    for w in &wanted {
        let p = PathBuf::from(w);
        match linked.iter().find(|e| worktree_slot::same(&e.path, &p)) {
            Some(e) => paths.push(fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone())),
            None => unregistered.push(w.clone()),
        }
    }
    paths.sort();
    paths.dedup();

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
        return Ok(answer);
    }

    let recent = Duration::from_secs(worktree_gc::DEFAULT_RECENT_MINUTES * 60);
    let budget = Duration::from_millis(worktree_gc::DEFAULT_BUDGET_MS);
    let now = now_seam();
    let pace = Pace {
        only: paths,
        ..Pace::default()
    };
    let plan = worktree_gc::plan_paced(
        project,
        req.integration_ref,
        recent,
        budget,
        now,
        false,
        &pace,
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
        let applied = worktree_gc::apply(
            project,
            req.integration_ref,
            recent,
            budget,
            now,
            &pace,
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
    Ok(answer)
}
