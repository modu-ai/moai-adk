//! The cleanup executor (SPEC-20261123-cleanup-outbox): it takes the requests a closed card left
//! in the ledger ([`crate::ledger::cleanup_outbox`]) and removes the card's worktrees and
//! branches after a fresh re-check.
//!
//! One request, in order:
//!
//! 1. claim it with a lease (another live claim: leave it; a result already there: nothing to do);
//! 2. re-read: the card is `done` in the ledger, and the head the request names is inside the
//!    integration tip. A card that is no longer done ends the request as `skipped`; a head that
//!    is not inside the tip keeps everything and says so;
//! 3. the worktrees: [`worktree_cleanup::run`] (the cleanup plan and its apply on the plan's own
//!    digest, `git worktree remove` without `--force`). It adds no way to remove a path: the
//!    safe criteria, the preserved `.claude/worktrees/` folder and the fixed slots are the plan's;
//! 4. the branches, only when no worktree of the card was kept: [`branch_gc::plan`] is read again
//!    for the exact integration ref and acts on the row of the one branch (`git branch -d` only);
//! 5. one `cleanup.step` row per target and one `cleanup.result` row, or, when something passes by
//!    itself or failed, the claim is given back and the request stays pending for a later run.
//!
//! No remote is read or written. A problem of the executor is the `error` of the answer and never
//! undoes the close that started it.

use super::command::open_for_write;
use super::landed;
use crate::ledger::cleanup_outbox::{self, Claim, Finish, Outcome, Pending, Step};
use crate::ledger::WriteLedger;
use crate::{branch_gc, worktree_cleanup, worktree_gc};
use serde_json::{json, Value};
use std::path::Path;
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// Branch reasons that pass by themselves: another card's reservation of the branch ends.
const BRANCH_TEMPORARY: &[&str] = &["card_reserved", "reservation_check_unavailable"];

/// Seams of one executor run (SPEC D-9): the library function takes them, the commands pass the
/// default.
pub struct Options<'a> {
    /// Lease length of the claim, in seconds of the database clock.
    pub lease_secs: i64,
    /// Stop after the worktree step and write nothing more, as a killed run would.
    pub stop_after_worktrees: bool,
    /// Called just before the branch step.
    pub before_branch: Option<&'a dyn Fn()>,
}

impl Default for Options<'_> {
    fn default() -> Self {
        Options {
            lease_secs: 300,
            stop_after_worktrees: false,
            before_branch: None,
        }
    }
}

/// Runs every pending cleanup request of the project and answers what each one did. Never fails:
/// a problem is the `error` of the answer.
pub fn run_pending(project: &Path, home: &Path, session: &str, opts: &Options<'_>) -> Value {
    match run(project, home, session, opts) {
        Ok(answer) => answer,
        Err(error) => json!({"requests": [], "error": error}),
    }
}

/// The request of `card_id` in an answer of [`run_pending`], when it has one.
pub fn entry_of(answer: &Value, card_id: i64) -> Option<&Value> {
    answer["requests"]
        .as_array()?
        .iter()
        .rev()
        .find(|r| r["card_id"] == json!(card_id))
}

/// The worktree cleanup answer of the card's request, the shape `card land` has always put under
/// its `cleanup` key: the cleanup's own answer when the request ran it, else an answer that
/// carries the `error` that stopped it, or `null` when the card has no request in the answer.
pub fn cleanup_of(answer: &Value, card_id: i64) -> Value {
    if let Some(entry) = entry_of(answer, card_id) {
        if entry["cleanup"].is_object() {
            return entry["cleanup"].clone();
        }
        if let Some(why) = entry["error"].as_str() {
            return json!({
                "trigger": "landing", "applied": false, "worktrees": [], "error": why,
            });
        }
        return Value::Null;
    }
    match answer["error"].as_str() {
        Some(why) => json!({
            "trigger": "landing", "applied": false, "worktrees": [], "error": why,
        }),
        None => Value::Null,
    }
}

/// True when the answer says the executor removed or deleted something.
pub fn wrote(answer: &Value) -> bool {
    answer["requests"].as_array().is_some_and(|list| {
        list.iter().any(|r| {
            ["worktrees", "branches"].iter().any(|key| {
                r[key].as_array().is_some_and(|items| {
                    items
                        .iter()
                        .any(|i| matches!(i["status"].as_str(), Some("removed" | "deleted")))
                })
            })
        })
    })
}

fn run(project: &Path, home: &Path, session: &str, opts: &Options<'_>) -> Result<Value> {
    let (_id, mut ledger) = open_for_write(project, home)?;
    let pending = cleanup_outbox::pending(ledger.conn()).map_err(|e| e.to_string())?;
    if pending.is_empty() {
        return Ok(json!({"requests": [], "error": null}));
    }
    let integration = landed::integration_branch(project).and_then(|b| landed::branch_ref(&b));
    let mut requests = Vec::new();
    let mut error: Option<String> = None;
    for request in &pending {
        let entry = match one(
            project,
            home,
            &mut ledger,
            session,
            opts,
            integration.as_deref().map_err(String::as_str),
            request,
        ) {
            Ok(entry) => entry,
            Err(why) => {
                error.get_or_insert(why.clone());
                json!({
                    "request": request.request, "card_id": request.card_id,
                    "status": "error", "error": why, "worktrees": [], "branches": [],
                })
            }
        };
        if let Some(why) = entry["error"].as_str() {
            error.get_or_insert(why.to_string());
        }
        requests.push(entry);
    }
    Ok(json!({"requests": requests, "error": error}))
}

/// The outcome of one target before it is written as a step.
struct Target {
    kind: &'static str,
    name: String,
    status: &'static str,
    reasons: Vec<String>,
    detail: Option<String>,
}

impl Target {
    fn json(&self) -> Value {
        let key = match self.kind {
            "worktree" => "path",
            "queue" => "ref",
            _ => "branch",
        };
        json!({
            key: self.name, "status": self.status, "reasons": self.reasons, "detail": self.detail,
        })
    }

    fn step(&self) -> Step {
        Step {
            kind: self.kind,
            target: self.name.clone(),
            status: self.status.to_string(),
            reasons: self.reasons.clone(),
            detail: self.detail.clone(),
        }
    }

    /// A target whose reason passes by itself or whose step failed: the request stays pending.
    fn open(&self) -> bool {
        matches!(self.status, "deferred" | "failed")
    }

    /// A target that was acted on or is already gone.
    fn clean(&self) -> bool {
        matches!(
            self.status,
            "removed" | "deleted" | "absent" | "gone" | "still_there"
        )
    }
}

fn state_of(ledger: &WriteLedger, card_id: i64) -> Result<String> {
    ledger
        .conn()
        .query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .map_err(|e| e.to_string())
}

/// True when `head` is an ancestor of (or equal to) the commit `reference` names.
fn inside(project: &Path, head: &str, reference: &str) -> Result<bool> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let status = cmd
        .arg("-C")
        .arg(project)
        .args(["merge-base", "--is-ancestor", head, reference])
        .status()
        .map_err(|e| format!("git could not run: {e}"))?;
    match status.code() {
        Some(0) => Ok(true),
        Some(1) => Ok(false),
        _ => Err(format!("git merge-base --is-ancestor {head} failed")),
    }
}

#[allow(clippy::too_many_arguments)]
fn one(
    project: &Path,
    home: &Path,
    ledger: &mut WriteLedger,
    session: &str,
    opts: &Options<'_>,
    integration: std::result::Result<&str, &str>,
    request: &Pending,
) -> Result<Value> {
    let base = |status: &str| {
        json!({
            "request": request.request, "card_id": request.card_id, "status": status,
            "worktrees": [], "branches": [],
        })
    };
    // No integration ref: nothing can be measured, so nothing is claimed and nothing is written.
    let integration: &str = integration.map_err(str::to_string)?;
    let (attempt, abandoned) = match cleanup_outbox::claim(
        ledger,
        request.request,
        request.card_id,
        session,
        opts.lease_secs,
    )
    .map_err(|e| e.to_string())?
    {
        Claim::Taken { attempt, abandoned } => (attempt, abandoned),
        Claim::Busy => return Ok(base("busy")),
        Claim::Over => return Ok(base("over")),
    };
    // The earlier attempt this claim found dead, when there was one, is named in every answer.
    let name = |status: &str| {
        let mut answer = base(status);
        if let Some(dead) = abandoned {
            answer["abandoned_attempt"] = json!(dead);
        }
        answer
    };
    // Whatever goes wrong from here, the claim is given back and the request stays pending. An
    // attempt whose lease ended writes nothing (SPEC-20261123-cleanup-outbox-followups REQ-1): the
    // answer is `lease_lost` and the next run takes the request again.
    let end =
        |ledger: &mut WriteLedger, steps: &[Step], outcome: &Outcome| -> Result<Option<Value>> {
            let finished = cleanup_outbox::finish_attempt(
                ledger,
                request.request,
                request.card_id,
                attempt,
                session,
                steps,
                outcome,
            )
            .map_err(|e| e.to_string())?;
            if finished != Finish::LeaseEnded {
                return Ok(None);
            }
            let mut lost = name("lease_lost");
            lost["attempt"] = json!(attempt);
            lost["error"] = json!(format!(
                "the lease of attempt {attempt} of request {} ended before the run finished, so \
             nothing of it was recorded; the next run takes the request again",
                request.request
            ));
            Ok(Some(lost))
        };

    // The card must still be done: a reopened card keeps its worktree and its branch.
    if state_of(ledger, request.card_id)? != "done" {
        let kept = Target {
            kind: "worktree",
            name: "(card)".into(),
            status: "kept",
            reasons: vec!["card_not_done".into()],
            detail: None,
        };
        if let Some(lost) = end(
            ledger,
            &[kept.step()],
            &Outcome::Result("skipped".to_string()),
        )? {
            return Ok(lost);
        }
        return Ok(name("skipped"));
    }
    if let Some(head) = &request.head {
        let proved = inside(project, head, integration);
        match proved {
            Ok(true) => {}
            Ok(false) => {
                let kept = Target {
                    kind: "worktree",
                    name: "(landing)".into(),
                    status: "kept",
                    reasons: vec!["landing_not_in_integration".into()],
                    detail: Some(format!(
                        "the landed head {head} is not inside {integration}"
                    )),
                };
                if let Some(lost) = end(
                    ledger,
                    &[kept.step()],
                    &Outcome::Result("settled_with_kept".to_string()),
                )? {
                    return Ok(lost);
                }
                let mut answer = name("settled_with_kept");
                answer["worktrees"] = json!([kept.json()]);
                return Ok(answer);
            }
            Err(why) => {
                // The git failure is the answer; a lease that ended meanwhile adds nothing to it.
                end(ledger, &[], &Outcome::Release)?;
                return Err(why);
            }
        }
    }

    // The worktrees: the cleanup plan and its apply, on the digest of the plan just read.
    let cleaned = worktree_cleanup::run(
        project,
        home,
        &worktree_cleanup::Request {
            card_id: request.card_id,
            integration_ref: integration,
            trigger: worktree_cleanup::Trigger::Landing,
        },
    );
    if let Some(why) = cleaned["error"].as_str() {
        if let Some(lost) = end(ledger, &[], &Outcome::Release)? {
            return Ok(lost);
        }
        let mut answer = name("error");
        answer["error"] = json!(why);
        answer["cleanup"] = cleaned.clone();
        return Ok(answer);
    }
    let worktrees = worktree_targets(project, &cleaned, request);
    if opts.stop_after_worktrees {
        let mut answer = name("interrupted");
        answer["worktrees"] = Value::Array(worktrees.iter().map(Target::json).collect());
        return Ok(answer);
    }
    if let Some(before) = opts.before_branch {
        before();
    }

    // The branches: only when every worktree of the card is gone or never was a worktree.
    let all_gone = worktrees.iter().all(Target::clean);
    let branches = branch_targets(project, home, integration, request, &worktrees, all_gone);

    // The queue the worktree step drained: one step per ref, so the ledger says which lines the
    // run consumed (SPEC-20261123-cleanup-outbox-followups REQ-5).
    let queue = queue_targets(&cleaned);
    let targets: Vec<&Target> = worktrees
        .iter()
        .chain(branches.iter())
        .chain(queue.iter())
        .collect();
    let steps: Vec<Step> = targets.iter().map(|t| t.step()).collect();
    let (status, outcome) = if targets.iter().any(|t| t.open()) {
        ("pending", Outcome::Release)
    } else if targets.iter().all(|t| t.clean()) {
        ("settled", Outcome::Result("settled".to_string()))
    } else {
        (
            "settled_with_kept",
            Outcome::Result("settled_with_kept".to_string()),
        )
    };
    // The run has acted by now. An attempt whose lease ended writes no line, but its answer still
    // names what it removed and deleted, so the caller (and `wrote`) is not told "nothing"
    // (SPEC-20261123-cleanup-outbox-followups D-5).
    let mut answer = match end(ledger, &steps, &outcome)? {
        Some(lost) => lost,
        None => {
            let mut done = name(status);
            done["attempt"] = json!(attempt);
            done
        }
    };
    answer["worktrees"] = Value::Array(worktrees.iter().map(Target::json).collect());
    answer["branches"] = Value::Array(branches.iter().map(Target::json).collect());
    answer["queue"] = Value::Array(queue.iter().map(Target::json).collect());
    answer["cleanup"] = cleaned;
    Ok(answer)
}

/// One target per ref of the queue the worktree step took, in ref order: `gone` for a ref that no
/// longer exists, `still_there` for one that does (its line was a pack or a move and is dropped).
fn queue_targets(cleaned: &Value) -> Vec<Target> {
    let mut refs: Vec<(&str, &'static str)> = Vec::new();
    for (key, status) in [("gone", "gone"), ("still_there", "still_there")] {
        for full in cleaned["queue"][key].as_array().into_iter().flatten() {
            if let Some(full) = full.as_str() {
                refs.push((full, status));
            }
        }
    }
    refs.sort();
    refs.dedup();
    refs.into_iter()
        .map(|(full, status)| Target {
            kind: "queue",
            name: full.to_string(),
            status,
            reasons: Vec::new(),
            detail: None,
        })
        .collect()
}

/// What the worktree cleanup did to each worktree the request names.
fn worktree_targets(project: &Path, cleaned: &Value, request: &Pending) -> Vec<Target> {
    let primary = crate::worktree_slot::common_dir(project)
        .and_then(|common| crate::worktree_slot::primary_of(&common))
        .ok()
        .map(|p| std::fs::canonicalize(&p).unwrap_or(p));
    // What git lists as a worktree now, read after the cleanup ran: a worktree is gone only when
    // it is no longer in this list, whatever the cleanup's own row said.
    let listed: Option<Result<Vec<std::path::PathBuf>>> = primary.as_deref().map(|root| {
        crate::worktree_slot::linked(root).map(|l| l.into_iter().map(|e| e.path).collect())
    });
    let unregistered: Vec<&str> = cleaned["not_a_worktree"]
        .as_array()
        .map(|a| a.iter().filter_map(Value::as_str).collect())
        .unwrap_or_default();
    request
        .worktrees
        .iter()
        .map(|w| {
            let mut target = Target {
                kind: "worktree",
                name: w.clone(),
                status: "absent",
                reasons: Vec::new(),
                detail: None,
            };
            // The reservation's relative path is the project's, never the folder of this run.
            let named = worktree_cleanup::reserved_path(primary.as_deref().unwrap_or(project), w);
            // A path whose last element is a link is never touched: it stands for what it points
            // at, which may be a worktree the card does not own.
            if worktree_cleanup::last_element_is_link(&named) {
                target.status = "kept";
                target.reasons = vec!["link_path".into()];
                return target;
            }
            if unregistered.contains(&w.as_str()) {
                return target;
            }
            let wanted = canonical(&named);
            let shown = primary.as_deref().map_or_else(
                || wanted.to_string_lossy().into_owned(),
                |root| worktree_gc::shown(&wanted, root),
            );
            let Some(row) = cleaned["worktrees"]
                .as_array()
                .and_then(|rows| rows.iter().find(|r| r["path"] == json!(shown)))
            else {
                // Registered but not listed: the plan did not look at it, so nothing was done.
                target.status = "kept";
                target.reasons = vec!["not_listed_by_the_plan".into()];
                return target;
            };
            let reasons: Vec<String> = row["reasons"]
                .as_array()
                .map(|a| {
                    a.iter()
                        .filter_map(|x| x.as_str().map(str::to_string))
                        .collect()
                })
                .unwrap_or_default();
            let detail = row["detail"].as_str().map(str::to_string);
            match row["status"].as_str() {
                Some("removed") => {
                    // Removed only when git no longer lists it; a list that cannot be read
                    // proves nothing, so the step fails and the request stays pending.
                    match &listed {
                        Some(Ok(still)) => {
                            if still.iter().any(|p| crate::worktree_slot::same(p, &wanted)) {
                                target.status = "kept";
                                target.reasons = vec!["still_registered".into()];
                            } else {
                                target.status = "removed";
                            }
                        }
                        _ => {
                            target.status = "failed";
                            target.detail = Some("the worktree list could not be read".into());
                        }
                    }
                }
                // The plan trimmed the build output and left the worktree standing: it is kept,
                // and so is its branch.
                Some("trimmed") => {
                    target.status = "kept";
                    target.reasons = vec!["trimmed_only".into()];
                    target.detail = detail;
                }
                _ => {
                    let passes = (!reasons.is_empty()
                        && reasons.iter().all(|r| worktree_cleanup::is_temporary(r)))
                        || detail
                            .as_deref()
                            .is_some_and(worktree_cleanup::is_temporary);
                    target.status = if passes { "deferred" } else { "kept" };
                    target.reasons = reasons;
                    target.detail = detail;
                }
            }
            target
        })
        .collect()
}

/// The canonical form of a worktree path that may be gone: the folder itself when it exists, else
/// its canonical parent and its name, so a removed worktree still names the row the plan listed.
fn canonical(given: &Path) -> std::path::PathBuf {
    if let Ok(real) = std::fs::canonicalize(given) {
        return real;
    }
    match (given.parent().map(std::fs::canonicalize), given.file_name()) {
        (Some(Ok(parent)), Some(name)) => parent.join(name),
        _ => given.to_path_buf(),
    }
}

/// What the branch step did to each branch the request names.
fn branch_targets(
    project: &Path,
    home: &Path,
    integration: &str,
    request: &Pending,
    worktrees: &[Target],
    all_gone: bool,
) -> Vec<Target> {
    let kept_by_worktree = |branch: &str| -> Target {
        let deferred = worktrees.iter().any(|t| t.status == "deferred");
        Target {
            kind: "branch",
            name: branch.to_string(),
            status: if deferred { "deferred" } else { "kept" },
            reasons: vec!["worktree_kept".into()],
            detail: None,
        }
    };
    if request.branches.is_empty() {
        return Vec::new();
    }
    if !all_gone {
        return request
            .branches
            .iter()
            .map(|b| kept_by_worktree(b))
            .collect();
    }
    let plan = match branch_gc::plan(project, integration, Some(home)) {
        Ok(plan) => plan,
        Err(why) => {
            return request
                .branches
                .iter()
                .map(|b| Target {
                    kind: "branch",
                    name: b.clone(),
                    status: "failed",
                    reasons: Vec::new(),
                    detail: Some(why.clone()),
                })
                .collect()
        }
    };
    let rows = plan["candidates"].as_array().cloned().unwrap_or_default();
    request
        .branches
        .iter()
        .map(|b| {
            let mut target = Target {
                kind: "branch",
                name: b.clone(),
                status: "absent",
                reasons: Vec::new(),
                detail: None,
            };
            let Some(row) = rows.iter().find(|r| r["branch"] == json!(b)) else {
                return target;
            };
            if row["action"] == json!("delete") {
                let tip = row["tip"].as_str().unwrap_or_default();
                match branch_gc::delete_planned(project, b, tip) {
                    Ok(()) => target.status = "deleted",
                    Err((_, detail)) => {
                        target.status = "failed";
                        target.detail = Some(detail);
                    }
                }
                return target;
            }
            let reasons: Vec<String> = row["reasons"]
                .as_array()
                .map(|a| {
                    a.iter()
                        .filter_map(|x| x.as_str().map(str::to_string))
                        .collect()
                })
                .unwrap_or_default();
            let passes = !reasons.is_empty()
                && reasons
                    .iter()
                    .all(|r| BRANCH_TEMPORARY.contains(&r.as_str()));
            target.status = if passes { "deferred" } else { "kept" };
            target.reasons = reasons;
            target
        })
        .collect()
}
