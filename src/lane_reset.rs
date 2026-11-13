//! `lane-safe-to-reset` (Issue #2669): may a lane's session be cleared (`/clear`) or its work
//! dispatched to another lane without losing anything? Read-only.
//!
//! The leader must not clear or redispatch a lane before checking, for that lane: which
//! worktrees it works in, whether any of them holds uncommitted, unpushed or stashed work,
//! whether its screen still runs a background job, and whether every open claim has a resume
//! note. This module gathers what the local repository can tell and takes the rest — the
//! lane's open claims with their resume notes, its session folder, its screen — from an
//! observations file the caller fills from the forge and the terminal, each with its source.
//!
//! The answer is `safe` only when every check ran and found nothing; any finding makes it
//! `not_safe`; a check that could not run (no observation, git failed, an observation too old)
//! makes it `unknown` with the reason. Nothing is written: git runs with
//! `--no-optional-locks` (a plain `status` may rewrite another lane's index) and a scrubbed
//! environment, and no fetch is made — "unpushed" means "not on any local remote-tracking
//! ref", which a stale clone can overstate but never understate.

use crate::identity::scrub_git_env;
use crate::worktree_report::{parse_worktree_list, WorktreeEntry};
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::{Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// Longest age, in seconds, of a screen observation the verdict may rely on.
pub const SCREEN_MAX_AGE_SECS: i64 = 600;

/// The observations file (typed JSON, unknown keys rejected). Every part is optional; a
/// missing part leaves its check `unknown`.
#[derive(Clone, Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Observations {
    pub schema_version: u32,
    #[serde(default)]
    pub claims: Option<ClaimsObservation>,
    #[serde(default)]
    pub session_cwd: Option<CwdObservation>,
    #[serde(default)]
    pub screen: Option<ScreenObservation>,
}

/// The lane's open claims as read from the forge.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ClaimsObservation {
    pub source: String,
    pub items: Vec<OpenClaim>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct OpenClaim {
    pub issue: u64,
    pub claim_comment: u64,
    /// The comment holding the resume note for this claim, if one exists.
    #[serde(default)]
    pub resume_note: Option<u64>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CwdObservation {
    pub source: String,
    pub path: PathBuf,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ScreenObservation {
    pub source: String,
    /// `YYYY-MM-DDTHH:MM:SSZ`.
    pub observed_at: String,
    pub background_jobs: u64,
}

/// A lane name as the fleet writes it: 1 to 64 of ASCII letters, digits, `-`, `_`, `.`.
pub fn valid_lane(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= 64
        && name
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
}

fn git(dir: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).into_owned())
}

/// Whether a worktree folder or branch name carries the lane name as a whole token
/// (`agent-26-audit-2237` carries `agent-26`, not `agent-2`). The fleet writes the same lane
/// with and without its dashes (`2636-adapters-agent22` for `agent-22`; audit FAIL 6850218012
/// on PR #2689), so the dashless spelling counts too.
pub fn name_carries_lane(name: &str, lane: &str) -> bool {
    let dashless = lane.replace('-', "");
    let bytes = name.as_bytes();
    let carries = |l: &str| {
        !l.is_empty()
            && name.match_indices(l).any(|(i, _)| {
                let end = i + l.len();
                let before = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
                let after = end == bytes.len() || !bytes[end].is_ascii_alphanumeric();
                before && after
            })
    };
    carries(lane) || carries(&dashless)
}

/// The `issue/<n>-…` number of a branch name.
fn branch_issue(branch: &str) -> Option<u64> {
    let rest = branch.strip_prefix("issue/")?;
    let digits: String = rest.chars().take_while(|c| c.is_ascii_digit()).collect();
    (!digits.is_empty() && rest[digits.len()..].starts_with('-'))
        .then(|| digits.parse().ok())
        .flatten()
}

/// Why each worktree belongs to the lane; empty means it does not.
fn matched_by(
    entry: &WorktreeEntry,
    lane: &str,
    claimed: &BTreeSet<u64>,
    cwd_worktree: Option<&Path>,
) -> Vec<String> {
    let mut why = Vec::new();
    if let Some(n) = entry.branch.as_deref().and_then(branch_issue) {
        if claimed.contains(&n) {
            why.push(format!("claim #{n}"));
        }
    }
    if cwd_worktree == Some(entry.path.as_path()) {
        why.push("session_cwd".into());
    }
    if entry
        .path
        .file_name()
        .is_some_and(|n| name_carries_lane(&n.to_string_lossy(), lane))
    {
        why.push("folder name".into());
    }
    // A branch named for the lane (`issue/2670-worktreeinclude-secrets-agent22`).
    if entry
        .branch
        .as_deref()
        .is_some_and(|b| name_carries_lane(b, lane))
    {
        why.push("branch name".into());
    }
    why
}

/// Stash entries per branch name (`WIP on <b>:` / `On <b>:`); `(no branch)` for detached.
pub(crate) fn stash_counts(project: &Path) -> Option<Vec<String>> {
    let text = git(project, &["stash", "list", "--format=%gs"])?;
    Some(
        text.lines()
            .filter_map(|l| {
                let rest = l
                    .strip_prefix("WIP on ")
                    .or_else(|| l.strip_prefix("On "))?;
                Some(rest.split_once(':').map_or(rest, |(b, _)| b).to_string())
            })
            .collect(),
    )
}

fn count_lines(text: &str) -> u64 {
    text.lines().filter(|l| !l.is_empty()).count() as u64
}

/// The worktree part of the answer; findings and unknowns are appended to the lists.
fn check_worktree(
    entry: &WorktreeEntry,
    why: Vec<String>,
    stashes: Option<&[String]>,
    attributed: bool,
    findings: &mut Vec<String>,
    unknowns: &mut Vec<String>,
) -> Value {
    let shown = entry.path.display().to_string();
    let uncommitted = git(
        &entry.path,
        &["status", "--porcelain=v1", "--untracked-files=normal"],
    )
    .map(|t| count_lines(&t));
    // Commits reachable from HEAD and from no remote-tracking ref (no fetch is made).
    let unpushed = git(
        &entry.path,
        &["rev-list", "--count", "HEAD", "--not", "--remotes"],
    )
    .and_then(|t| t.trim().parse::<u64>().ok());
    let label = entry.branch.clone().unwrap_or_else(|| "(no branch)".into());
    let stash = stashes.map(|s| s.iter().filter(|b| **b == label).count() as u64);
    for (what, value) in [
        ("uncommitted change", uncommitted),
        ("unpushed commit", unpushed),
        ("stash entry", stash),
    ] {
        match value {
            None => unknowns.push(format!("{shown}: {what}s could not be read")),
            Some(0) => {}
            Some(n) => findings.push(format!("{shown}: {n} {what}(s)")),
        }
    }
    json!({
        "attributed": attributed,
        "path": shown,
        "branch": entry.branch,
        "head": entry.head,
        "matched_by": why,
        "uncommitted": uncommitted,
        "unpushed": unpushed,
        "stash": stash,
    })
}

/// Builds the answer for `lane` in the repository holding `project`, with `now` as Unix
/// seconds (injected so the screen-age check does not depend on the wall clock in tests).
pub fn check(project: &Path, lane: &str, obs: &Observations, now: i64) -> Result<Value> {
    if !valid_lane(lane) {
        return Err("invalid lane name: 1 to 64 of letters, digits, `-`, `_`, `.`".into());
    }
    if obs.schema_version != 1 {
        return Err("invalid observations: schema_version must be 1".into());
    }
    let list = git(project, &["worktree", "list", "--porcelain"])
        .ok_or("git worktree list unavailable for the project")?;
    let entries = parse_worktree_list(&list);
    let (mut findings, mut unknowns) = (Vec::new(), Vec::new());

    // (4) every open claim has a resume note.
    let claimed: BTreeSet<u64> = obs
        .claims
        .as_ref()
        .map(|c| c.items.iter().map(|i| i.issue).collect())
        .unwrap_or_default();
    let claims = match &obs.claims {
        None => {
            unknowns.push("no claims observation: open claims and resume notes unchecked".into());
            Value::Null
        }
        Some(c) => {
            for i in c.items.iter().filter(|i| i.resume_note.is_none()) {
                findings.push(format!(
                    "open claim #{} (comment {}) has no resume note",
                    i.issue, i.claim_comment
                ));
            }
            json!({
                "source": c.source,
                "items": c.items.iter().map(|i| json!({
                    "issue": i.issue,
                    "claim_comment": i.claim_comment,
                    "resume_note": i.resume_note,
                })).collect::<Vec<_>>(),
            })
        }
    };

    // (1) the lane's worktrees, (2) their uncommitted, unpushed and stashed work.
    // git lists real paths (`/private/var/…` on macOS), so a symlinked spelling is resolved.
    let cwd_real = obs
        .session_cwd
        .as_ref()
        .map(|c| c.path.canonicalize().unwrap_or_else(|_| c.path.clone()));
    let cwd = cwd_real.as_deref();
    if obs.session_cwd.is_none() {
        unknowns.push("no session_cwd observation: the session's own folder unchecked".into());
    }
    let stashes = stash_counts(project);
    if stashes.is_none() {
        unknowns.push("git stash list unavailable".into());
    }
    // Linked worktrees usually live inside the primary checkout, so the session folder belongs
    // to the deepest worktree that contains it, not to every one.
    let cwd_worktree = cwd.and_then(|c| {
        entries
            .iter()
            .filter(|e| !e.bare && c.starts_with(&e.path))
            .max_by_key(|e| e.path.components().count())
            .map(|e| e.path.as_path())
    });
    // The primary checkout is shared: lanes start there and work in linked worktrees, so its
    // changes belong to the person and the leader, not to a lane whose only tie to it is the
    // session folder (leader's first use, #2669 comment 6851881955). Its counts are reported
    // under `unattributed` and do not decide the verdict.
    // Only a non-bare first entry is a primary checkout; a bare repository has none, and its
    // first linked worktree is a lane's like any other (Codex review on PR #2715).
    let primary = entries
        .first()
        .filter(|e| !e.bare)
        .map(|e| e.path.as_path());
    let mut worktrees = Vec::new();
    let mut unattributed = Vec::new();
    for entry in entries.iter().filter(|e| !e.bare) {
        let why = matched_by(entry, lane, &claimed, cwd_worktree);
        if why.is_empty() {
            continue;
        }
        let only_cwd = why.iter().all(|w| w == "session_cwd");
        let attributed = !(only_cwd && primary == Some(entry.path.as_path()));
        let (mut f, mut u) = (Vec::new(), Vec::new());
        worktrees.push(check_worktree(
            entry,
            why,
            stashes.as_deref(),
            attributed,
            &mut f,
            &mut u,
        ));
        if attributed {
            findings.extend(f);
            unknowns.extend(u);
        } else {
            unattributed.extend(f.into_iter().chain(u));
        }
    }
    if let Some(c) = cwd {
        if cwd_worktree.is_none() {
            unknowns.push(format!(
                "session_cwd {} is in no worktree of this repository",
                c.display()
            ));
        }
    }
    for n in &claimed {
        let has = entries
            .iter()
            .any(|e| e.branch.as_deref().and_then(branch_issue) == Some(*n));
        if !has {
            unknowns.push(format!(
                "open claim #{n} has no local `issue/{n}-…` worktree to check"
            ));
        }
    }

    // (3) background jobs on the lane's screen.
    let screen = match &obs.screen {
        None => {
            unknowns.push("no screen observation: background jobs unchecked".into());
            Value::Null
        }
        Some(s) => {
            match crate::forge::policy::parse_utc(&s.observed_at) {
                None => unknowns.push(format!(
                    "screen observed_at `{}` is not YYYY-MM-DDTHH:MM:SSZ",
                    s.observed_at
                )),
                Some(at) if now - at > SCREEN_MAX_AGE_SECS || at > now => unknowns.push(format!(
                    "screen observation is {} s old; the limit is {SCREEN_MAX_AGE_SECS} s",
                    now - at
                )),
                Some(_) if s.background_jobs > 0 => findings.push(format!(
                    "{} background job(s) still running on the lane's screen",
                    s.background_jobs
                )),
                Some(_) => {}
            }
            json!({"source": s.source, "observed_at": s.observed_at, "background_jobs": s.background_jobs})
        }
    };

    let verdict = if !findings.is_empty() {
        "not_safe"
    } else if !unknowns.is_empty() {
        "unknown"
    } else {
        "safe"
    };
    Ok(json!({
        "lane": lane,
        "verdict": verdict,
        "findings": findings,
        "unknowns": unknowns,
        "unattributed": unattributed,
        "worktrees": worktrees,
        "claims": claims,
        "session_cwd": obs.session_cwd.as_ref().map(|c| json!({"source": c.source, "path": c.path})),
        "screen": screen,
        "unpushed_basis": "local remote-tracking refs; no fetch",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn lane_names_match_as_whole_tokens() {
        assert!(name_carries_lane("agent-26-audit-2237", "agent-26"));
        assert!(name_carries_lane("audit-2491-agent-28", "agent-28"));
        assert!(name_carries_lane("agent-27", "agent-27"));
        assert!(!name_carries_lane("agent-26-audit-2237", "agent-2"));
        assert!(!name_carries_lane("xagent-27", "agent-27"));
        assert!(!name_carries_lane("2669-lane-safe-to-reset", "agent-27"));
        // The fleet also drops the dash (audit FAIL 6850218012 on PR #2689).
        assert!(name_carries_lane("2636-adapters-agent22", "agent-22"));
        assert!(name_carries_lane(
            "issue/2670-worktreeinclude-secrets-agent22",
            "agent-22"
        ));
        assert!(name_carries_lane("audit-2486-agent29", "agent-29"));
        assert!(!name_carries_lane("2636-adapters-agent220", "agent-22"));
        assert!(!name_carries_lane("2636-adapters-agent2", "agent-22"));
    }

    #[test]
    fn branch_numbers_need_the_issue_prefix_and_a_dash() {
        assert_eq!(branch_issue("issue/2669-lane-safe"), Some(2669));
        assert_eq!(branch_issue("issue/2669"), None);
        assert_eq!(branch_issue("feature/2669-x"), None);
        assert_eq!(branch_issue("issue/x-1"), None);
    }

    #[test]
    fn lane_name_limits() {
        assert!(valid_lane(&"a".repeat(64)));
        assert!(!valid_lane(&"a".repeat(65)));
        assert!(!valid_lane(""));
        assert!(!valid_lane("agent 27"));
        assert!(!valid_lane("../x"));
    }
}
