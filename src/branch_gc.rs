//! `branch.gc.plan` and `branch.gc.apply` (SPEC-20261123-branch-gc): the local branches that
//! are merged into the integration ref and unused, and their deletion after a digest match.
//!
//! The plan is read-only. A branch is a `delete` row only when all four hold: its tip is an
//! ancestor of the integration tip, no worktree has it checked out, no active card reservation
//! names it, and it is not a preserved ref. Every other branch is a `keep` row with its reasons.
//! A fact that cannot be established (the profile or the ledger exists but cannot be read) keeps
//! the branch: it never becomes a `delete` row on a guess.
//!
//! The apply deletes with `git branch -d` only. `-d` is git's own second check (merged into the
//! branch's upstream, or into HEAD without one) and the reason `-D` is never used. Nothing here
//! runs a remote command or reads a remote-tracking ref.

use crate::identity::scrub_git_env;
use crate::worktree_gc::git_out;
use crate::worktree_report::parse_worktree_list;
use serde_json::{json, Value};
use std::collections::{BTreeMap, HashSet};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// Branch names that are never deleted (REQ-2), beside the `archive/` prefix and the local
/// branch the integration ref names.
const PRESERVED_NAMES: [&str; 4] = ["main", "master", "v2026.11.0", "2026.11.0-rc.1"];
const PRESERVED_PREFIX: &str = "archive/";

/// Active reservation statuses; the same set `src/ledger/reservation.rs` calls active.
const ACTIVE_STATUSES: &str = "('held','expired_candidate','held_for_review')";

struct Branch {
    name: String,
    tip: String,
    symbolic: bool,
}

/// Every local branch with its tip and whether it is a symbolic ref (`%(symref)` is set).
fn local_branches(top: &Path) -> Result<Vec<Branch>> {
    let text = git_out(
        top,
        &[
            "for-each-ref",
            "--format=%(refname)%09%(objectname)%09%(symref)",
            "refs/heads",
        ],
    )
    .ok_or("git for-each-ref failed")?;
    let mut out = Vec::new();
    for line in text.lines() {
        let mut parts = line.split('\t');
        let (Some(full), Some(tip), Some(symref)) = (parts.next(), parts.next(), parts.next())
        else {
            return Err(format!("unreadable branch line: {line}"));
        };
        let Some(name) = full.strip_prefix("refs/heads/") else {
            continue;
        };
        out.push(Branch {
            name: name.to_string(),
            tip: tip.to_string(),
            symbolic: !symref.is_empty(),
        });
    }
    Ok(out)
}

/// The branches whose tip is an ancestor of (or equal to) `tip`, in one git call.
fn merged_into(top: &Path, tip: &str) -> Result<HashSet<String>> {
    let text = git_out(
        top,
        &[
            "for-each-ref",
            &format!("--merged={tip}"),
            "--format=%(refname)",
            "refs/heads",
        ],
    )
    .ok_or("git for-each-ref --merged failed")?;
    Ok(text
        .lines()
        .filter_map(|l| l.strip_prefix("refs/heads/"))
        .map(str::to_string)
        .collect())
}

/// The branches any worktree (the primary one, linked ones, prunable ones) has checked out.
fn checked_out(top: &Path) -> Result<HashSet<String>> {
    let listing =
        git_out(top, &["worktree", "list", "--porcelain"]).ok_or("git worktree list failed")?;
    Ok(parse_worktree_list(&listing)
        .into_iter()
        .filter_map(|e| e.branch)
        .collect())
}

/// The branch names active card reservations name. A project without a profile or without a
/// ledger has none. A profile or ledger that exists but cannot be read is `Err(())`.
fn reserved(top: &Path, home: Option<&Path>) -> std::result::Result<HashSet<String>, ()> {
    let real = fs::canonicalize(top).map_err(|_| ())?;
    let Some(id) = crate::project_id::read(&real).map_err(|_| ())? else {
        return Ok(HashSet::new());
    };
    let home = crate::ledger::command::resolve_home(home.map(Path::to_path_buf)).map_err(|_| ())?;
    let file = crate::ledger::ledger_path(&home, &id).map_err(|_| ())?;
    if !file.is_file() {
        return Ok(HashSet::new());
    }
    let read = crate::ledger::open_read_only(&home, &id).map_err(|_| ())?;
    let mut stmt = read
        .conn()
        .prepare(&format!(
            "SELECT branch FROM reservation WHERE status IN {ACTIVE_STATUSES} AND branch IS NOT NULL"
        ))
        .map_err(|_| ())?;
    let rows = stmt
        .query_map([], |r| r.get::<_, String>(0))
        .map_err(|_| ())?
        .collect::<std::result::Result<Vec<String>, _>>()
        .map_err(|_| ())?;
    Ok(rows
        .into_iter()
        .map(|b| {
            b.strip_prefix("refs/heads/")
                .map(str::to_string)
                .unwrap_or(b)
        })
        .collect())
}

fn top_of(project: &Path) -> Result<PathBuf> {
    git_out(
        project,
        &["rev-parse", "--path-format=absolute", "--show-toplevel"],
    )
    .map(|s| PathBuf::from(s.trim()))
    .ok_or_else(|| "not inside a git work tree".to_string())
}

/// The commit `integration_ref` names, and the local branch it names when it is one.
fn integration(top: &Path, integration_ref: &str) -> Result<(String, Option<String>)> {
    // `git rev-parse --symbolic-full-name` echoes `--end-of-options` back as a name, so the
    // value is kept from looking like an option instead.
    if integration_ref.starts_with('-') {
        return Err(format!(
            "integration ref `{integration_ref}` must not start with `-`"
        ));
    }
    let tip = git_out(
        top,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            "--end-of-options",
            &format!("{integration_ref}^{{commit}}"),
        ],
    )
    .map(|s| s.trim().to_string())
    .filter(|s| !s.is_empty())
    .ok_or_else(|| format!("integration ref `{integration_ref}` is not a commit here"))?;
    let local = git_out(top, &["rev-parse", "--symbolic-full-name", integration_ref])
        .and_then(|s| s.trim().strip_prefix("refs/heads/").map(str::to_string));
    Ok((tip, local))
}

fn preserved(name: &str, integration_branch: Option<&str>) -> bool {
    PRESERVED_NAMES.contains(&name)
        || name.starts_with(PRESERVED_PREFIX)
        || integration_branch == Some(name)
}

/// Builds the plan (`branch.gc.plan`). `home` is the HOME that holds the ledger (`None`: the
/// environment's); tests pass a temporary one.
pub fn plan(project: &Path, integration_ref: &str, home: Option<&Path>) -> Result<Value> {
    let top = top_of(project)?;
    let (tip, integration_branch) = integration(&top, integration_ref)?;
    let branches = local_branches(&top)?;
    let merged = merged_into(&top, &tip)?;
    let held = checked_out(&top)?;
    let reservations = reserved(&top, home);

    let mut rows: BTreeMap<String, Value> = BTreeMap::new();
    let mut delete: Vec<(String, String)> = Vec::new();
    for b in &branches {
        let mut reasons: Vec<&str> = Vec::new();
        if preserved(&b.name, integration_branch.as_deref()) {
            reasons.push("preserved");
        }
        if held.contains(&b.name) {
            reasons.push("checked_out");
        }
        if reservations
            .as_ref()
            .is_ok_and(|names| names.contains(&b.name))
        {
            reasons.push("card_reserved");
        }
        if !merged.contains(&b.name) {
            reasons.push("not_merged");
        }
        if b.symbolic {
            reasons.push("symbolic_ref");
        }
        if reasons.is_empty() && reservations.is_err() {
            reasons.push("reservation_check_unavailable");
        }
        let action = if reasons.is_empty() { "delete" } else { "keep" };
        if action == "delete" {
            delete.push((b.name.clone(), b.tip.clone()));
        }
        rows.insert(
            b.name.clone(),
            json!({"branch": b.name, "tip": b.tip, "action": action, "reasons": reasons}),
        );
    }
    delete.sort();
    let bound = json!({"integration_tip": tip, "delete": delete});
    let plan_digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    let candidates: Vec<Value> = rows.into_values().collect();
    Ok(json!({
        "schema_version": 1,
        "integration_ref": integration_ref,
        "integration_tip": tip,
        "totals": {
            "branches": candidates.len(),
            "delete": delete.len(),
            "keep": candidates.len() - delete.len(),
        },
        "candidates": candidates,
        "plan_digest": plan_digest,
        "apply": "branch-gc-apply with this plan_digest and a permission reference; on this machine the leader runs it only after the user confirms",
        "network": "disabled",
        "writes": false,
    }))
}

fn tip_of(top: &Path, branch: &str) -> Option<String> {
    git_out(
        top,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            "--end-of-options",
            &format!("refs/heads/{branch}"),
        ],
    )
    .map(|s| s.trim().to_string())
    .filter(|s| !s.is_empty())
}

/// Deletes one planned branch: the tip and the checkouts are read again, then `git branch -d`
/// runs, then the ref is read back. `Err((status, detail))` carries what the answer reports.
fn delete_one(
    top: &Path,
    branch: &str,
    planned_tip: &str,
) -> std::result::Result<(), (&'static str, String)> {
    match tip_of(top, branch) {
        Some(now) if now == planned_tip => {}
        Some(_) => return Err(("refused", "the branch tip moved after the plan".into())),
        None => return Err(("refused", "the branch is gone".into())),
    }
    match checked_out(top) {
        Ok(held) if held.contains(branch) => {
            return Err(("refused", "a worktree has the branch checked out".into()))
        }
        Ok(_) => {}
        Err(e) => return Err(("refused", e)),
    }
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(top)
        .args(["branch", "-d", "--"])
        .arg(branch)
        .output()
        .map_err(|e| ("failed", format!("git branch could not run: {e}")))?;
    if !out.status.success() {
        return Err((
            "refused",
            String::from_utf8_lossy(&out.stderr).trim().to_string(),
        ));
    }
    if tip_of(top, branch).is_some() {
        return Err((
            "failed",
            "the ref is still there after git branch -d".into(),
        ));
    }
    Ok(())
}

/// Acts on a plan (`branch.gc.apply`): recomputes it with the same input, refuses before any
/// write unless its digest equals `plan_digest`, then deletes each `delete` branch. A refused or
/// failed branch is reported and the rest go on; nothing is retried. `before_delete` is called
/// with each branch name just before it is re-checked (a test seam; the command passes a no-op).
pub fn apply(
    project: &Path,
    integration_ref: &str,
    plan_digest: &str,
    permission_reference: &str,
    home: Option<&Path>,
    before_delete: &dyn Fn(&str),
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let fresh = plan(project, integration_ref, home)?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    let top = top_of(project)?;
    let mut results = Vec::new();
    let mut not_done = 0usize;
    for c in fresh["candidates"].as_array().into_iter().flatten() {
        if c["action"] != "delete" {
            continue;
        }
        let branch = c["branch"].as_str().unwrap_or_default();
        let tip = c["tip"].as_str().unwrap_or_default();
        before_delete(branch);
        let (status, detail) = match delete_one(&top, branch, tip) {
            Ok(()) => ("deleted", Value::Null),
            Err((status, detail)) => {
                not_done += 1;
                (status, json!(detail))
            }
        };
        results.push(json!({"branch": branch, "tip": tip, "status": status, "detail": detail}));
    }
    Ok(json!({
        "schema_version": 1,
        "status": if results.is_empty() { "unchanged" } else if not_done == 0 { "applied" } else { "partial" },
        "plan_digest": plan_digest,
        "results": results,
        "restore": "a deleted branch comes back with `git branch <name> <tip>` from this list",
        "retry": "none: a refused or failed branch is reported, never retried",
        "network": "disabled",
        "writes": !results.is_empty(),
    }))
}
