//! `work-flow-plan` (SPEC-20261120-mflow-integration-branch-choice): the evidence for the
//! first-setup question — which branch integrates this project's work. Reads only: it
//! prints the branch candidates (a pull-request candidate ref is never canonical), one
//! recommended branch with its reason — the recorded `work_flow.integration_branch` on a
//! re-run, else the remote's default branch, else the current branch — the recorded
//! work-flow state, and, when the recorded target is missing or occupied by another
//! worktree, the state with the safe choices. It creates no branch, worktree, file or ref.

use serde_json::{json, Value};
use std::path::Path;
use std::process::{Command, Stdio};

/// Remote-ref first components that name pull-request candidates, never integration
/// targets (`refs/remotes/origin/pr/<n>` and friends).
const CANDIDATE_NAMESPACES: &[&str] = &["pr", "pull", "patch", "review", "mr"];

/// Runs one read-only git command in `project`; `Err` on a failing command.
fn git_out(project: &Path, args: &[&str]) -> Result<String, String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    // Its own process group, so nothing it starts outlives this process (the
    // plan_work_flow_remote shape; the commands here are local reads).
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let out = cmd
        .arg("-C")
        .arg(project)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    if out.status.success() {
        Ok(String::from_utf8_lossy(&out.stdout).trim().to_string())
    } else {
        Err(format!("git {args:?} failed"))
    }
}

/// The candidate entry for one git ref: `main`-style local names and `origin/<name>`
/// remote-tracking names are canonical; a pull-request candidate namespace is not.
fn candidate(refname: &str) -> Option<Value> {
    let name = refname
        .strip_prefix("refs/heads/")
        .map(|n| (n.to_string(), true, "local"))
        .or_else(|| {
            refname.strip_prefix("refs/remotes/").map(|n| {
                let rest = n.split_once('/').map(|(_, r)| r).unwrap_or(n);
                let first = rest.split('/').next().unwrap_or("");
                if CANDIDATE_NAMESPACES.contains(&first) {
                    (n.to_string(), false, "candidate")
                } else {
                    (n.to_string(), true, "remote-tracking")
                }
            })
        })?;
    let (name, canonical, kind) = name;
    let mut v = json!({"name": name, "canonical": canonical, "kind": kind});
    if !canonical {
        v["label"] = json!("candidate: not the integration target");
    }
    Some(v)
}

/// The branch another worktree has checked out, if the recorded branch is occupied by a
/// worktree other than the project folder itself.
fn occupied_worktree(project: &Path, recorded: &str) -> Option<String> {
    let text = git_out(project, &["worktree", "list", "--porcelain"]).ok()?;
    for block in text.split("\n\n") {
        let mut path: Option<&str> = None;
        let mut branch: Option<&str> = None;
        for line in block.lines() {
            if let Some(p) = line.strip_prefix("worktree ") {
                path = Some(p);
            } else if let Some(b) = line.strip_prefix("branch refs/heads/") {
                branch = Some(b);
            }
        }
        if let (Some(p), Some(b)) = (path, branch) {
            if b == recorded && Path::new(p) != project {
                return Some(p.to_string());
            }
        }
    }
    None
}

/// Builds the whole answer: candidates, recommendation, profile, state.
pub fn plan(project: &Path) -> Result<Value, String> {
    let project = project
        .canonicalize()
        .map_err(|e| format!("cannot read {project:?}: {e}"))?;
    // A profile that is present but cannot be read is its own state; a missing profile is
    // just an unrecorded one, and the discovery fallbacks carry on (leader decision on the
    // slice-1 audit: profile-unreadable, note 4115).
    let profile_read = crate::profile::read_profile(&project);
    let profile_unreadable = match &profile_read {
        Err(e) if e != "project profile missing" => Some(e.to_string()),
        _ => None,
    };
    let work_flow = profile_read.ok().and_then(|(p, _)| p.work_flow);
    let recorded: Option<String> = work_flow
        .as_ref()
        .and_then(|w| w.integration_branch.clone())
        .filter(|b| !b.trim().is_empty());

    let refs = git_out(
        &project,
        &[
            "for-each-ref",
            "--format=%(refname)",
            "refs/heads/",
            "refs/remotes/",
        ],
    )?;
    let candidates: Vec<Value> = refs.lines().filter_map(candidate).collect();
    let has_local = |b: &str| refs.lines().any(|l| l == format!("refs/heads/{b}"));
    let has_origin = |b: &str| {
        refs.lines()
            .any(|l| l == format!("refs/remotes/origin/{b}"))
    };
    let current = git_out(&project, &["branch", "--show-current"]).unwrap_or_default();

    // A release-prefixed branch is for release pull requests (user decision: remote pull
    // requests only at a formal release), so the discovery rules never name one; the
    // recorded value stays the user's own answer (leader decision, note 4115).
    let usable_remote_default = if let Ok(head) = git_out(
        &project,
        &["symbolic-ref", "-q", "refs/remotes/origin/HEAD"],
    ) {
        let named = head
            .strip_prefix("refs/remotes/origin/")
            .unwrap_or(&head)
            .to_string();
        // origin/HEAD that was set by something else — a pull-request candidate namespace,
        // or a ref no fetch ever brought — names no integration target: skipped.
        let namespaced = named
            .split('/')
            .next()
            .is_some_and(|first| CANDIDATE_NAMESPACES.contains(&first));
        let exists = refs.lines().any(|l| l == head);
        (!namespaced && exists && !named.starts_with("release/")).then_some((
            named.clone(),
            format!("the remote's default branch (origin/HEAD names {named})"),
        ))
    } else {
        None
    };
    let current_pick = (!current.is_empty() && !current.starts_with("release/"))
        .then(|| (current.clone(), "the current branch".to_string()));

    let (branch, reason) = if let Some(b) = recorded.as_ref() {
        (
            b.clone(),
            "the value recorded in the project profile (work_flow.integration_branch)".to_string(),
        )
    } else if let Some((b, r)) = usable_remote_default {
        (b, r)
    } else if let Some((b, r)) = current_pick {
        (b, r)
    } else {
        (
            String::new(),
            "the checkout is detached; pick a branch from the candidates".to_string(),
        )
    };

    let recommendation = if branch.is_empty() {
        Value::Null
    } else {
        json!({"branch": branch, "reason": reason})
    };

    let occupied = recorded
        .as_deref()
        .and_then(|b| occupied_worktree(&project, b));
    let state = if let Some(e) = profile_unreadable {
        json!({
            "kind": "profile-unreadable",
            "message": format!(
                "the project profile cannot be read ({e}); the recommendation comes from the repository instead — this tool creates nothing on its own"
            ),
        })
    } else {
        match recorded.as_deref() {
            None => json!({"kind": "ok"}),
            Some(b) => {
                if let Some(wt) = occupied.as_deref() {
                    json!({
                        "kind": "occupied-worktree",
                        "worktree": wt,
                        "message": format!(
                            "the recorded integration branch {b} is checked out in another worktree ({wt}); close it or pick another branch — this tool creates nothing on its own"
                        ),
                    })
                } else if !has_local(b) && !has_origin(b) {
                    json!({
                        "kind": "missing-target",
                        "message": format!(
                            "the recorded integration branch {b} is missing; pick an existing branch from the candidates or confirm creating one with the user — this tool creates nothing on its own"
                        ),
                    })
                } else if !has_local(b) {
                    json!({
                        "kind": "missing-target",
                        "message": format!(
                            "the recorded integration branch {b} is missing locally (it exists on origin); fetch and create an existing branch from it or pick another existing branch — this tool creates nothing on its own"
                        ),
                    })
                } else {
                    json!({"kind": "ok"})
                }
            }
        }
    };

    Ok(json!({
        "recommendation": recommendation,
        "candidates": candidates,
        "profile": work_flow.map(|w| json!({"flow": w.flow, "integration_branch": w.integration_branch})),
        "state": state,
    }))
}

#[cfg(test)]
mod tests {
    use super::candidate;

    /// A pull-request candidate namespace is never canonical, in any remote.
    #[test]
    fn candidate_namespaces_are_never_canonical() {
        for ns in ["pr", "pull", "patch", "review", "mr"] {
            let v = candidate(&format!("refs/remotes/origin/{ns}/42")).unwrap();
            assert_eq!(v["canonical"], false, "{ns}");
            assert!(v["label"].as_str().unwrap().contains("candidate"));
        }
    }

    /// Local and remote-tracking branch refs stay canonical, with no label.
    #[test]
    fn branch_refs_stay_canonical() {
        let local = candidate("refs/heads/main").unwrap();
        assert_eq!(local["canonical"], true);
        assert_eq!(local["kind"], "local");
        let remote = candidate("refs/remotes/origin/trunk").unwrap();
        assert_eq!(remote["canonical"], true);
        assert_eq!(remote["kind"], "remote-tracking");
        assert!(remote.get("label").is_none());
    }
}
