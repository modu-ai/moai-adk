//! `moai card remote-copy` and the copy tail of `moai card land` (SPEC-20261122-lander-push, the
//! parent SPEC-20261122-remote-copy-slim REQ-1): after a landing closes its transaction the
//! integration branch is copied to its remote with one fast-forward-only push, and a copy that
//! failed, was refused or lost its answer is retried by the next landing command or on demand —
//! no resident loop, no new card.
//!
//! The state is derived, never stored as a latch: every run reads the branch and its tip, asks
//! the remote where its branch stands (`ls-remote`), and pushes only when the remote is behind or
//! absent — git's own PUSH RULES make a branch push fast-forward-only, and no force flag is ever
//! passed. A rejection and a remote failure are told apart by git's documented push output
//! (git-push(1), OUTPUT): the porcelain status line of the one ref is `!` with `rejected` for a
//! non-fast-forward, `!` with `remote failure` when the remote did not report the outcome, and an
//! unreachable remote prints no status line at all (measured with git 2.54.0, 2026-11-22).
//!
//! Every failure folds into the answer ([`Outcome`]), so no caller can fail on a copy: a landing
//! that could not copy answers `pending`, and the record of the state is one JSON file in the
//! common git folder, `moai/remote-copy`, written by a temporary file and a rename — the landing
//! itself stays recorded exactly once by the close.
//!
//! The git commands run through the git environment scrub and take their program from
//! `MOAI_GIT_BIN` when that is set, like the other card commands (the private helper of
//! `landed.rs` is repeated here because that module's helper is its own).

use super::landed;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::time::{SystemTime, UNIX_EPOCH};

/// The one record of the copy state, inside the common git folder.
pub const RECORD: &str = "moai/remote-copy";

/// How a copy ended.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CopyState {
    /// The remote holds the local tip (it was pushed, already held it, or holds a descendant).
    Done,
    /// The copy could not finish; the next landing command or the on-demand command retries it.
    Pending,
    /// The histories diverged (or the remote refused): remote and local stay as they are.
    Hold,
}

impl CopyState {
    fn as_str(self) -> &'static str {
        match self {
            CopyState::Done => "done",
            CopyState::Pending => "pending",
            CopyState::Hold => "hold",
        }
    }
}

/// What one copy run came to. `why` names the reason of a `pending` or `hold` state.
#[derive(Debug, Clone)]
pub struct Outcome {
    pub branch: String,
    pub full_ref: String,
    pub remote: String,
    pub tip: String,
    pub remote_tip: Option<String>,
    pub state: CopyState,
    pub why: Option<String>,
    pub pushed: bool,
    pub at: i64,
    pub by: String,
}

impl Outcome {
    /// The answer object of the copy (without an `action` key: the callers add theirs).
    pub fn json(&self) -> Value {
        json!({
            "branch": self.branch, "ref": self.full_ref, "remote": self.remote,
            "tip": self.tip, "remote_tip": self.remote_tip,
            "state": self.state.as_str(), "why": self.why, "pushed": self.pushed,
            "at": self.at, "by": self.by,
        })
    }
}

/// One git command in `project`, run the way the card commands run git.
fn git(project: &Path, args: &[&str]) -> Result<Output, String> {
    let program = std::env::var("MOAI_GIT_BIN").unwrap_or_else(|_| "git".to_string());
    let mut cmd = Command::new(program);
    crate::identity::scrub_git_env(&mut cmd);
    cmd.arg("--no-optional-locks")
        .arg("-C")
        .arg(project)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("git could not run: {e}"))
}

fn now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or_default()
}

/// The porcelain status line of the one pushed ref: the first char is git's flag (` ` a pushed
/// fast-forward, `*` a created ref, `=` up to date, `!` rejected or failed) and the text after
/// the last tab is the summary (`[rejected] (non-fast-forward)` and its kind, git-push(1)).
/// `None` means no status line was printed: the documented mark of a remote never reached.
fn porcelain_status(stdout: &str) -> Option<(char, String)> {
    for line in stdout.lines() {
        let mut chars = line.chars();
        if let (Some(flag), Some('\t')) = (chars.next(), chars.next()) {
            if " *+=-!".contains(flag) {
                let summary = line.rsplit_once('\t').map(|(_, s)| s.trim()).unwrap_or("");
                return Some((flag, summary.to_string()));
            }
        }
    }
    None
}

/// The remote a push of `branch` goes to, by git's own push order: `branch.<name>.pushRemote`,
/// then `remote.pushDefault`, then `branch.<name>.remote`, then `origin` (git-push(1)). A key git
/// cannot answer names nothing and the chain moves on.
fn push_remote(project: &Path, branch: &str) -> String {
    for key in [
        format!("branch.{branch}.pushRemote"),
        "remote.pushDefault".to_string(),
        format!("branch.{branch}.remote"),
    ] {
        if let Ok(out) = git(project, &["config", "--get", &key]) {
            let value = String::from_utf8_lossy(&out.stdout).trim().to_string();
            if out.status.success() && !value.is_empty() {
                return value;
            }
        }
    }
    "origin".to_string()
}

/// Where the remote's branch stands: `Ok(None)` when the remote has no such ref, `Err` when the
/// remote could not be reached at all.
fn ls_remote(project: &Path, remote: &str, full_ref: &str) -> Result<Option<String>, String> {
    let out = git(project, &["ls-remote", remote, full_ref])?;
    if !out.status.success() {
        let why = String::from_utf8_lossy(&out.stderr);
        let why = why.lines().find(|l| !l.trim().is_empty()).unwrap_or(&why);
        return Err(format!(
            "the remote {remote} could not be read: {}",
            why.trim()
        ));
    }
    // `<sha>\t<ref>` per line; the sha of the one ref asked for.
    Ok(String::from_utf8_lossy(&out.stdout)
        .lines()
        .find_map(|l| l.split_once('\t'))
        .filter(|(r, _)| *r == full_ref)
        .map(|(sha, _)| sha.trim().to_string())
        .filter(|sha| crate::gate::is_full_sha(sha)))
}

/// Whether `a` is an ancestor of `b`; `Err` when git cannot tell (a commit neither side has).
fn is_ancestor(project: &Path, a: &str, b: &str) -> Result<bool, String> {
    let out = git(project, &["merge-base", "--is-ancestor", a, b])?;
    match out.status.code() {
        Some(0) => Ok(true),
        Some(1) => Ok(false),
        _ => Err(format!(
            "git could not tell whether {a} is inside {b}: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        )),
    }
}

/// One fast-forward push of the one ref (no force flag anywhere); the outcome is read from git's
/// documented porcelain output, and `Err` means only that git could not run at all.
fn push(project: &Path, remote: &str, full_ref: &str) -> Result<Option<(char, String)>, String> {
    let spec = format!("{full_ref}:{full_ref}");
    let out = git(project, &["push", "--porcelain", remote, &spec])?;
    let stdout = String::from_utf8_lossy(&out.stdout);
    if let Some(status) = porcelain_status(&stdout) {
        return Ok(Some(status));
    }
    if out.status.success() {
        // A success that printed no status line (git prints one for every pushed ref): read as
        // the pushed fast-forward it reported nowhere.
        return Ok(Some((' ', String::new())));
    }
    // No status line: the remote was never reached; git's first stderr line is the reason.
    let stderr = String::from_utf8_lossy(&out.stderr);
    let why = stderr
        .lines()
        .find(|l| !l.trim().is_empty())
        .unwrap_or(&stderr);
    Ok(Some(('!', why.trim().to_string())))
}

/// The common git folder of `project`, absolute (git prints a relative path for the main worktree).
fn common_dir(project: &Path) -> Result<PathBuf, String> {
    let out = git(
        project,
        &["rev-parse", "--path-format=absolute", "--git-common-dir"],
    )?;
    if !out.status.success() {
        return Err(format!(
            "git could not name its common folder: {}",
            String::from_utf8_lossy(&out.stderr).trim()
        ));
    }
    Ok(PathBuf::from(String::from_utf8_lossy(&out.stdout).trim()))
}

/// Writes the one record by a temporary file and a rename; a record that cannot be written
/// changes nothing else (the answer carries the state with or without it).
fn write_record(project: &Path, outcome: &Outcome) -> Option<String> {
    let name = RECORD.rsplit_once('/')?.1;
    let dir = common_dir(project).ok()?.join("moai");
    fs::create_dir_all(&dir).ok()?;
    let path = dir.join(name);
    let temp = dir.join(format!(".{name}.tmp"));
    let body = serde_json::to_vec_pretty(&outcome.json()).ok()?;
    fs::write(&temp, body).ok()?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let _ = fs::set_permissions(&temp, fs::Permissions::from_mode(0o600));
    }
    match fs::rename(&temp, &path) {
        Ok(()) => Some(path.to_string_lossy().into_owned()),
        Err(_) => {
            let _ = fs::remove_file(&temp);
            None
        }
    }
}

/// What the classification decided this run does.
enum Step {
    /// The remote is behind or absent: one fast-forward push.
    Push,
    /// Nothing to push: the outcome is already settled.
    Settled,
}

/// One copy run. Never fails: every refusal, network trouble and divergence is an [`Outcome`] the
/// caller answers, so a landing that could not copy still stands. `by` names the caller (`land`
/// or `remote-copy`); the record of the state is written beside the answer.
pub fn copy(project: &Path, by: &str) -> Outcome {
    let mut outcome = match classify(project) {
        Ok(classified) => match classified.step {
            Step::Push => push_once(project, classified.outcome),
            Step::Settled => classified.outcome,
        },
        Err(pending) => pending,
    };
    outcome.at = now();
    outcome.by = by.to_string();
    write_record(project, &outcome);
    outcome
}

/// The answer of `card remote-copy`: the copy run by the on-demand command itself.
pub fn entry(project: &Path) -> Value {
    let outcome = copy(project, "remote-copy");
    let mut answer = outcome.json();
    answer["action"] = json!("remote-copy");
    // No ledger row is written; the copy writes only its one record in the git folder.
    answer["writes"] = json!(false);
    answer
}

struct Classified {
    outcome: Outcome,
    step: Step,
}

/// Reads the branch, the tip and the remote position and decides what this run does; the error is
/// already a `pending` outcome with the reason.
fn classify(project: &Path) -> Result<Classified, Outcome> {
    let branch = landed::integration_branch(project).map_err(pending_no_branch)?;
    let full_ref = landed::branch_ref(&branch).map_err(pending_no_branch)?;
    let tip = landed::tip_of(project, &branch).map_err(pending_no_branch)?;
    let remote = push_remote(project, &branch);
    let outcome = Outcome {
        branch: branch.clone(),
        full_ref: full_ref.clone(),
        remote: remote.clone(),
        tip: tip.clone(),
        remote_tip: None,
        state: CopyState::Pending,
        why: None,
        pushed: false,
        at: 0,
        by: String::new(),
    };
    let remote_tip = match ls_remote(project, &remote, &full_ref) {
        Ok(found) => found,
        Err(why) => {
            return Err(Outcome {
                why: Some(why),
                ..outcome
            })
        }
    };
    let outcome = Outcome {
        remote_tip: remote_tip.clone(),
        ..outcome
    };
    match remote_tip.as_deref() {
        // The remote has no such branch: the copy is a creation that destroys nothing.
        None => Ok(Classified {
            outcome,
            step: Step::Push,
        }),
        Some(r) if r == tip => Ok(Classified {
            outcome: Outcome {
                state: CopyState::Done,
                why: Some("the remote already holds this commit".to_string()),
                ..outcome
            },
            step: Step::Settled,
        }),
        Some(r) => match is_ancestor(project, r, &tip) {
            // The remote is behind: one fast-forward push.
            Ok(true) => Ok(Classified {
                outcome,
                step: Step::Push,
            }),
            // The remote holds a descendant of the tip: only the answer was lost; settled.
            Ok(false) => match is_ancestor(project, &tip, r) {
                Ok(true) => Ok(Classified {
                    outcome: Outcome {
                        state: CopyState::Done,
                        why: Some(
                            "the remote already holds this commit or a descendant".to_string(),
                        ),
                        ..outcome
                    },
                    step: Step::Settled,
                }),
                Ok(false) => Err(hold(
                    outcome,
                    format!(
                        "the remote tip {r} shares no ancestry with the local tip {tip}: nothing \
                         was pushed and nothing was forced"
                    ),
                )),
                Err(why) => Err(hold(outcome, why)),
            },
            Err(why) => Err(hold(outcome, why)),
        },
    }
}

/// A `pending` outcome for a project whose branch or tip could not even be read.
fn pending_no_branch(why: String) -> Outcome {
    Outcome {
        why: Some(why),
        ..empty_outcome()
    }
}

fn empty_outcome() -> Outcome {
    Outcome {
        branch: String::new(),
        full_ref: String::new(),
        remote: String::new(),
        tip: String::new(),
        remote_tip: None,
        state: CopyState::Pending,
        why: None,
        pushed: false,
        at: 0,
        by: String::new(),
    }
}

/// The classification's hold: remote and local stay as they are, and the reason is recorded.
fn hold(outcome: Outcome, why: String) -> Outcome {
    Outcome {
        state: CopyState::Hold,
        why: Some(why),
        ..outcome
    }
}

/// One fast-forward push, read by git's documented porcelain output; a rejection re-reads the
/// remote once (the remote may have moved to our tip while the answer was in flight).
fn push_once(project: &Path, outcome: Outcome) -> Outcome {
    let mut after = outcome.clone();
    match push(project, &outcome.remote, &outcome.full_ref) {
        // A pushed fast-forward, a created ref or up to date (the flags ` `, `*`, `=`).
        Ok(Some((flag, _))) if flag != '!' => {
            after.state = CopyState::Done;
            after.pushed = true;
            after.why = None;
        }
        Ok(Some(('!', summary))) => {
            if summary.contains("rejected") {
                settle_after_rejection(project, &mut after, &summary);
            } else if summary.contains("remote failure") {
                after.why = Some(format!("the remote did not report the outcome: {summary}"));
            } else {
                after.state = CopyState::Hold;
                after.why = Some(summary);
            }
        }
        Ok(None) => {}
        Err(why) => after.why = Some(why),
    }
    after
}

/// After a rejection: read the remote once more and settle as success when it now holds our tip
/// or a descendant; anything else is a hold with git's own words.
fn settle_after_rejection(project: &Path, after: &mut Outcome, summary: &str) {
    match ls_remote(project, &after.remote, &after.full_ref) {
        Ok(Some(r)) if r == after.tip => {
            after.state = CopyState::Done;
            after.why =
                Some("the remote took the commit while the answer was in flight".to_string());
        }
        Ok(Some(r)) => {
            let remote_holds_it = is_ancestor(project, &after.tip, &r).unwrap_or(false);
            if remote_holds_it {
                after.state = CopyState::Done;
                after.why =
                    Some("the remote already holds this commit or a descendant".to_string());
            } else {
                after.state = CopyState::Hold;
                after.why = Some(format!(
                    "the push was rejected ({summary}) and the remote tip {r} is not behind the \
                     local tip {}: nothing was pushed and nothing was forced",
                    after.tip
                ));
            }
        }
        _ => {
            after.state = CopyState::Hold;
            after.why = Some(format!("the push was rejected ({summary})"));
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The measured porcelain lines of git 2.54.0 (git-push(1), OUTPUT), read as one status each.
    #[test]
    fn the_porcelain_reader_takes_the_one_status_line_of_each_measured_case() {
        let rejected = "To ../remote.git\n!\trefs/heads/integ:refs/heads/integ\t[rejected] (non-fast-forward)\nDone\n";
        assert_eq!(
            porcelain_status(rejected),
            Some(('!', "[rejected] (non-fast-forward)".to_string()))
        );
        let created =
            "To ../remote.git\n*\trefs/heads/integ:refs/heads/integ\t[new branch]\nDone\n";
        assert_eq!(
            porcelain_status(created),
            Some(('*', "[new branch]".to_string()))
        );
        let forwarded =
            "To ../remote.git\n \trefs/heads/integ:refs/heads/integ\t54c2aa4..9a592e2\nDone\n";
        assert_eq!(
            porcelain_status(forwarded),
            Some((' ', "54c2aa4..9a592e2".to_string()))
        );
        let same = "To ../remote.git\n=\trefs/heads/integ:refs/heads/integ\t[up to date]\nDone\n";
        assert_eq!(
            porcelain_status(same),
            Some(('=', "[up to date]".to_string()))
        );
    }

    /// An unreachable remote prints no status line at all (measured: only `fatal:` on stderr).
    #[test]
    fn an_output_without_a_status_line_reads_as_none() {
        assert_eq!(porcelain_status("To ../gone\nDone\n"), None);
        assert_eq!(porcelain_status(""), None);
    }
}
