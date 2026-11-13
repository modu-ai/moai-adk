//! `gate-merge` (Issue #2674 S2): the leader's merge, after the gate passes — what the end of
//! `target/leader/gate_merge.sh` did. Four steps, each only after the one before succeeded:
//!
//! 1. collect the snapshot and run every gate rule ([`crate::gate::check`]); a stop writes nothing;
//! 2. post the `moai/audit` success status on the head ([`crate::forge::audit_status::post`]);
//! 3. merge with `sha` set to the head, so GitHub refuses if the head moved; a refusal that may
//!    only mean the new status is not registered yet is retried, at most three attempts;
//! 4. post the merge record comment — only when the pull request is merged.
//!
//! Every write needs a permission reference; authorization stays with the operator.

use crate::forge::audit_status::{self, AuditStatusRequest};
use crate::forge::policy::Clock;
use crate::forge::transport::{HttpRequest, Transport};
use crate::gate;
use crate::gate_collect::{self, Target};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// Merge attempts at most; the pause between them lets a fresh status register.
pub const MERGE_ATTEMPTS: u32 = 3;
const RETRY_PAUSE_S: u64 = 5;

fn short(s: &str) -> &str {
    s.get(..10).unwrap_or(s)
}

/// Runs the four steps for `target`. `app` says whether the audit App token posts the status.
pub fn merge<T: Transport, C: Clock>(
    target: &Target<'_>,
    permission: &str,
    app: bool,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was read or written".into());
    }
    let snapshot = gate_collect::collect(target, transport)?;
    let (base, overlap) = match gate::check(&snapshot) {
        Ok(pass) => pass,
        Err(stop) => {
            return Ok(json!({
                "pr": target.pr, "head": snapshot.head, "merged": false,
                "rule": stop.rule, "reason": stop.reason, "writes": [],
            }))
        }
    };
    let head = snapshot.head.clone();
    let verdict_id = snapshot
        .comments
        .iter()
        .filter(|c| {
            c.body
                .lines()
                .next()
                .is_some_and(|l| l.starts_with("AUDIT #"))
        })
        .map(|c| c.id)
        .max()
        .ok_or("no verdict comment")?;
    let (owner, repo, pr) = (target.owner, target.repo, target.pr);
    let tip = crate::audit_delta::command(target.dir)
        .args([
            "rev-parse",
            "--verify",
            &format!("{}^{{commit}}", target.tip),
        ])
        .output()
        .map_err(|e| format!("cannot execute git: {e}"))
        .map(|o| String::from_utf8_lossy(&o.stdout).trim().to_string())?;

    let status = AuditStatusRequest {
        schema_version: 1,
        owner: owner.into(),
        repo: repo.into(),
        sha: head.clone(),
        state: "success".into(),
        target_url: format!(
            "https://github.com/{owner}/{repo}/pull/{pr}#issuecomment-{verdict_id}"
        ),
        description: format!(
            "AUDIT PASS head {}; tip {} clean, overlap {}",
            short(&head),
            short(&tip),
            overlap.len()
        ),
    };
    audit_status::post(&status, permission, app, transport, clock).map_err(|e| e.message)?;
    let mut writes = vec![json!("status")];

    let url = format!("https://api.github.com/repos/{owner}/{repo}/pulls/{pr}/merge");
    let mut merged_sha = None;
    let mut last = 0;
    for attempt in 1..=MERGE_ATTEMPTS {
        let resp = transport.send(
            &HttpRequest::new("PUT", &url).with_body(json!({"merge_method": "merge", "sha": head})),
        )?;
        last = resp.status;
        match resp.status {
            200 => {
                merged_sha = resp
                    .body
                    .as_ref()
                    .and_then(|b| b["sha"].as_str())
                    .map(String::from);
                break;
            }
            // 405 not mergeable yet (status not registered), 409 head moved: GitHub's answers.
            405 if attempt < MERGE_ATTEMPTS => clock.sleep(RETRY_PAUSE_S),
            405 | 409 => break,
            s => return Err(format!("merge PUT returned HTTP {s}")),
        }
    }
    let Some(merge_sha) = merged_sha else {
        return Ok(json!({
            "pr": pr, "head": head, "merged": false, "writes": writes,
            "reason": format!("not merged after the attempts (last HTTP {last}); no merge record posted"),
        }));
    };
    writes.push(json!("merge"));

    let checks = snapshot
        .checks
        .iter()
        .filter(|c| c.conclusion.as_deref() == Some("success"))
        .count();
    let note = if overlap.is_empty() {
        String::new()
    } else {
        format!(" 겹친 파일: {}.", overlap.join(", "))
    };
    let record = format!(
        "리더 병합 기록 (`moai-cli gate-merge`, #2674): 판정 {verdict_id} PASS(head {}, base {}), check run success {checks}, Self-Audit checker 통과. 병합 시점 tip {} 과 merge-tree 충돌 0, 판정 base 이후 tip 변경 파일과 이 PR 변경 파일 겹침 {}.{note} 병합 커밋 {}.\n\n🗿 MoAI",
        short(&head),
        short(&base),
        short(&tip),
        overlap.len(),
        short(&merge_sha),
    );
    let resp = transport.send(
        &HttpRequest::new(
            "POST",
            format!("https://api.github.com/repos/{owner}/{repo}/issues/{pr}/comments"),
        )
        .with_body(json!({"body": record})),
    )?;
    if resp.status == 201 {
        writes.push(json!("record"));
    }
    Ok(json!({
        "pr": pr, "head": head, "merged": true, "merge_commit": merge_sha,
        "verdict": verdict_id, "verdict_base": base, "overlap": overlap,
        "record_posted": resp.status == 201, "writes": writes,
    }))
}
