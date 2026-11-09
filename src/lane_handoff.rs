//! Lane handoff check (A10 #2264): the read-only gate a lane runs before safely
//! exiting its worktree (design §25 HANDOFF-01 / EXIT-01).
//!
//! Checks, in order:
//! 1. `status_clean` — no uncommitted changes (`git status --porcelain` empty).
//! 2. `head_remote_match` — local HEAD equals the remote head the lane observed
//!    via `git ls-remote` (injected observation record, not re-measured here).
//! 3. `pr_exists` — a PR record is present and its head equals the local HEAD.
//!    PR creation is *delivered*, not done: separation fields keep issue-done
//!    and merge explicitly unobserved (design §25).
//! 4. `trailers` — every commit (merges included) in `base..HEAD` carries the
//!    claim trailer `Moai-Claim: <issue>#<comment_id>`.
//! 5. `leader_report` — a durable report record file exists, is non-empty, and
//!    mentions the issue. The record's existence is enforced; the sending itself
//!    is attested (AC-3: 권고/강제 구분 — enforced rows are the mechanically
//!    verified paths only).
//! 6. `no_uncertain_effects` — no forge outbox entry is left in `publish_unknown`
//!    (EXIT-01: uncertain external effects block exit).
//!
//! Any failed check blocks exit: the operation surfaces
//! `lane handoff blocked: <failed names>` (kind `blocked`, exit != 0) with
//! per-check rows on stderr and a preservation note. Nothing here writes.

use serde::Deserialize;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};

type Result<T> = std::result::Result<T, String>;

pub const PROTOCOL_VERSION: u32 = 1;

#[derive(Deserialize)]
pub struct LaneHandoffRequest {
    pub protocol_version: u32,
    pub project: PathBuf,
    pub branch: String,
    pub base_branch: String,
    pub issue: u64,
    pub claim_comment_id: u64,
    /// The remote head the lane observed via `git ls-remote origin <branch>`
    /// (injected observation record, 40-hex).
    pub remote_head: String,
    /// The PR record the lane observed via its host client; null = no PR.
    pub pr: Option<PrRecord>,
    /// Durable record of the leader report (path to the saved report text).
    pub leader_report_path: Option<String>,
}

#[derive(Clone, Deserialize)]
pub struct PrRecord {
    pub url: String,
    pub number: u64,
    pub head: String,
}

fn git(project: &Path, args: &[&str]) -> Result<String> {
    let out = Command::new("git")
        .arg("-C")
        .arg(project)
        .args(args)
        .output()
        .map_err(|_| "git unavailable".to_string())?;
    if !out.status.success() {
        return Err(String::from_utf8_lossy(&out.stderr).trim().to_string());
    }
    Ok(String::from_utf8_lossy(&out.stdout).trim().to_string())
}

fn hex40(s: &str) -> bool {
    s.len() == 40 && s.bytes().all(|b| b.is_ascii_hexdigit())
}

struct Check {
    name: &'static str,
    enforcement: &'static str,
    ok: bool,
    detail: String,
}

fn check(name: &'static str, enforcement: &'static str, ok: bool, detail: String) -> Check {
    Check {
        name,
        enforcement,
        ok,
        detail,
    }
}

fn uncertain_effects(project: &Path) -> Result<u64> {
    let dir = project.join(".moai-cli").join("forge").join("outbox");
    let entries = match std::fs::read_dir(&dir) {
        Ok(e) => e,
        Err(_) => return Ok(0), // no outbox yet = no unresolved effects
    };
    let mut uncertain = 0u64;
    for entry in entries.flatten() {
        let path = entry.path();
        if path.extension().and_then(|e| e.to_str()) != Some("json") {
            continue;
        }
        let bytes = match fs::read(&path) {
            Ok(b) => b,
            Err(_) => continue,
        };
        let value: Value = match serde_json::from_slice(&bytes) {
            Ok(v) => v,
            Err(_) => continue,
        };
        if value["state"] == json!("publish_unknown") {
            uncertain += 1;
        }
    }
    Ok(uncertain)
}

/// Read-only lane handoff gate (REQ-A4-01 pattern applied to the loop's exit).
/// Returns the full report; the caller turns a `failed` status into
/// `lane handoff blocked: …` (kind `blocked`, exit != 0) with the per-check
/// rows on stderr. Nothing writes.
pub fn handoff_check(req: &LaneHandoffRequest) -> Result<Value> {
    if req.protocol_version != PROTOCOL_VERSION {
        return Err(format!(
            "unsupported protocol_version; supported: {}",
            PROTOCOL_VERSION
        ));
    }
    if !hex40(&req.remote_head) {
        return Err(format!(
            "remote_head must be a 40-hex commit, found {:?}",
            req.remote_head
        ));
    }
    let project = std::fs::canonicalize(&req.project)
        .map_err(|_| "project directory unavailable".to_string())?;
    let mut checks: Vec<Check> = Vec::new();

    // 1. status_clean
    let porcelain = git(&project, &["status", "--porcelain"]).unwrap_or_default();
    let dirty: Vec<&str> = porcelain.lines().filter(|l| !l.is_empty()).collect();
    checks.push(check(
        "status_clean",
        "enforced",
        dirty.is_empty(),
        if dirty.is_empty() {
            "no uncommitted changes".into()
        } else {
            format!(
                "{} uncommitted changes; commit or preserve them before exit",
                dirty.len()
            )
        },
    ));

    // 2. head_remote_match
    let head = git(&project, &["rev-parse", "HEAD"]).unwrap_or_default();
    let head_ok = hex40(&head) && head == req.remote_head;
    let short = |s: &str| s[..8.min(s.len())].to_string();
    checks.push(check(
        "head_remote_match",
        "enforced",
        head_ok,
        if head_ok {
            format!(
                "HEAD {} == observed remote {}",
                short(&head),
                short(&req.remote_head)
            )
        } else {
            format!(
                "HEAD {} != observed remote {}; push or re-observe ls-remote",
                short(&head),
                short(&req.remote_head)
            )
        },
    ));

    // 3. pr_exists (injected observation record)
    let pr_ok = req.pr.is_some();
    checks.push(check(
        "pr_exists",
        "enforced",
        pr_ok,
        if pr_ok {
            "PR record present".into()
        } else {
            "no PR record; open the PR before exiting".into()
        },
    ));
    // 4. trailers — every commit in base..HEAD carries the claim trailer.
    let trailer = format!("Moai-Claim: {}#{}", req.issue, req.claim_comment_id);
    let revs = git(
        &project,
        &["rev-list", &format!("{}..HEAD", req.base_branch)],
    )
    .unwrap_or_default();
    let mut trailer_commits = 0usize;
    let mut missing_trailers: Vec<String> = Vec::new();
    if !revs.is_empty() {
        for sha in revs.lines() {
            let body = git(&project, &["show", "-s", "--format=%B", sha]).unwrap_or_default();
            trailer_commits += 1;
            if !body.contains(&trailer) {
                missing_trailers.push(short(sha));
            }
        }
    }
    let pr_head_ok = req.pr.as_ref().map(|p| p.head == head).unwrap_or(false);
    checks.push(check(
        "trailers",
        "enforced",
        !revs.is_empty() && missing_trailers.is_empty() && pr_head_ok,
        if revs.is_empty() {
            "no commits in base..HEAD".into()
        } else if !missing_trailers.is_empty() {
            format!(
                "commits without the claim trailer: {}",
                missing_trailers.join(", ")
            )
        } else if !pr_head_ok {
            format!(
                "PR head does not match local HEAD ({} commits checked)",
                trailer_commits
            )
        } else {
            format!("all {} commits carry the claim trailer", trailer_commits)
        },
    ));

    // 5. leader_report — durable record file (enforced existence; the sending
    //    itself is attested, AC-3).
    let report_ok = match &req.leader_report_path {
        Some(p) => {
            fs::metadata(p)
                .map(|m| m.is_file() && m.len() > 0)
                .unwrap_or(false)
                && fs::read_to_string(p)
                    .map(|c| c.contains(&format!("#{}", req.issue)))
                    .unwrap_or(false)
        }
        None => false,
    };
    checks.push(check(
        "leader_report",
        "attested",
        report_ok,
        if report_ok {
            "leader report record present".into()
        } else {
            "leader report record missing; send the report and save the record".into()
        },
    ));

    // 6. no_uncertain_effects (EXIT-01)
    let uncertain = uncertain_effects(&project).unwrap_or(0);
    checks.push(check(
        "no_uncertain_effects",
        "enforced",
        uncertain == 0,
        if uncertain == 0 {
            "no publish_unknown effects in the outbox".into()
        } else {
            format!("{} publish_unknown effects need preservation", uncertain)
        },
    ));

    let failed: Vec<&str> = checks.iter().filter(|c| !c.ok).map(|c| c.name).collect();
    let status = if failed.is_empty() {
        "passed"
    } else {
        "failed"
    };

    let rows: Vec<Value> = checks
        .iter()
        .map(|c| {
            json!({
                "name": c.name,
                "enforcement": c.enforcement,
                "ok": c.ok,
                "detail": c.detail,
            })
        })
        .collect();

    let pr_json = match &req.pr {
        Some(p) => json!({
            "exists": true,
            "url": p.url,
            "number": p.number,
            "head": p.head,
            "head_matches": pr_head_ok,
        }),
        None => json!({"exists": false}),
    };

    let preservation = if status == "failed" {
        "exit refused; changes and records are preserved - resolve the failed \
         checks and re-run this check before exiting"
            .to_string()
    } else {
        "safe to exit with the worktree preserved (keep); the PR stays open and \
         integration is the leader's queue"
            .to_string()
    };

    Ok(json!({
        "status": status,
        "branch": req.branch,
        "base_branch": req.base_branch,
        "head": head,
        "remote_head": req.remote_head,
        "checks": rows,
        "failed_checks": failed,
        "pr": pr_json,
        // PR 제출은 delivered이지 issue done·merge가 아니다 (design §25).
        "separation": {
            "pr_created_is_not_done": true,
            "issue_done": null,
            "merged": false,
        },
        "trailer": trailer,
        "trailer_commits": trailer_commits,
        "uncertain_effects": uncertain,
        "preservation": preservation,
        "writes": false,
    }))
}

/// Summarizes a failed report into the envelope message. `report` must carry
/// `status` and `failed_checks` (produced by [`handoff_check`]). None = passed.
pub fn blocked_message(report: &Value) -> Option<String> {
    if report["status"] == json!("passed") {
        return None;
    }
    let names: Vec<&str> = report["failed_checks"]
        .as_array()?
        .iter()
        .filter_map(|n| n.as_str())
        .collect();
    Some(format!("lane handoff blocked: {}", names.join("; ")))
}
