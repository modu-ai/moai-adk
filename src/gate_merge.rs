//! `gate-merge` (Issue #2674 S2): the leader's merge, after the gate passes — what the end of
//! `target/leader/gate_merge.sh` did. Four steps, each only after the one before succeeded:
//!
//! 1. collect the snapshot and run every gate rule ([`crate::gate::check`]), then read the PR's
//!    base branch on GitHub: when it is not the local `--tip` (or the head moved), the gate was
//!    judged on a stale tip and nothing is written (audit 6856351941 F1);
//! 2. post the `moai/audit` success status on the head ([`crate::forge::audit_status::post`]);
//! 3. merge with `sha` set to the head, so GitHub refuses if the head moved; a refusal that may
//!    only mean the new status is not registered yet is retried, at most three attempts;
//! 4. post the merge record comment — only when the pull request is merged.
//!
//! A transport error after a write may hide its effect, so the PR is read back and the result
//! says what is known (`merged`, `merge_commit`, `record_posted: "uncertain"`) instead of a bare
//! error (audit F2). Reads, the merge and the record go through the operator's transport; only
//! the status goes through `status_transport`, which may carry the audit App token (audit F3).
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

/// `(merged, merge_commit_sha)` read back from the pull request; `None` when the read fails.
fn read_back<T: Transport>(transport: &T, url: &str) -> Option<(bool, Option<String>)> {
    let resp = transport.send(&HttpRequest::new("GET", url)).ok()?;
    let body = resp.body.filter(|_| resp.status == 200)?;
    Some((
        body["merged"].as_bool()?,
        body["merge_commit_sha"].as_str().map(String::from),
    ))
}

/// Runs the four steps for `target`. `transport` (the operator's) reads, merges and records;
/// `status_transport` posts the status, and `app` says whether it carries the audit App token.
pub fn merge<T: Transport, S: Transport, C: Clock>(
    target: &Target<'_>,
    permission: &str,
    app: bool,
    transport: &T,
    status_transport: &S,
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

    // The gate judged the local tip; GitHub merges onto the remote base. Read both back.
    let pr_url = format!("https://api.github.com/repos/{owner}/{repo}/pulls/{pr}");
    let resp = transport.send(&HttpRequest::new("GET", &pr_url))?;
    let pr_body = resp
        .body
        .filter(|_| resp.status == 200)
        .ok_or_else(|| format!("reading the pull request returned HTTP {}", resp.status))?;
    let base_ref = pr_body["base"]["ref"]
        .as_str()
        .ok_or("the pull request has no base ref")?;
    let resp = transport.send(&HttpRequest::new(
        "GET",
        format!("https://api.github.com/repos/{owner}/{repo}/git/ref/heads/{base_ref}"),
    ))?;
    let remote_tip = resp
        .body
        .filter(|_| resp.status == 200)
        .and_then(|b| b["object"]["sha"].as_str().map(String::from))
        .ok_or_else(|| format!("reading the base branch returned HTTP {}", resp.status))?;
    let remote_head = pr_body["head"]["sha"].as_str().unwrap_or_default();
    if remote_tip != tip || remote_head != head {
        return Ok(json!({
            "pr": pr, "head": head, "merged": false, "rule": "G6",
            "reason": format!(
                "the gate judged tip {} and head {}, but GitHub has base {base_ref} at {} and head {}; fetch and run again",
                short(&tip), short(&head), short(&remote_tip), short(remote_head)
            ),
            "writes": [],
        }));
    }

    // #2833 D1: a PASS carried by a mechanical delta is recorded on the PR before any other
    // write, once per `from → to`, so the audit trail names the head it now covers. A failed
    // record writes nothing else.
    let mut writes = Vec::new();
    if let Some(d) = gate::accepted_auto_delta(&snapshot) {
        let first = format!("AUTO-DELTA #{pr} {}→{}", d.from, d.to);
        let already = snapshot
            .comments
            .iter()
            .any(|c| c.body.lines().next() == Some(first.as_str()));
        if !already {
            let mut body = format!(
                "{first}\n\ngate-merge (#2833): {} merges only the base branch into {}; the PASS in comment {verdict_id} on {} counts for {}. CI on {}: success (G5).\n\n| file | patch-id |\n| --- | --- |\n",
                short(&d.to),
                short(&d.from),
                short(&d.from),
                short(&d.to),
                short(&d.to),
            );
            for f in &d.files {
                body.push_str(&format!("| `{}` | `{}` |\n", f.path, f.patch_id_from));
            }
            let sent = transport.send(
                &HttpRequest::new(
                    "POST",
                    format!("https://api.github.com/repos/{owner}/{repo}/issues/{pr}/comments"),
                )
                .with_body(json!({ "body": body })),
            );
            match sent {
                Ok(r) if r.status == 201 => writes.push(json!("auto_delta")),
                Ok(r) => {
                    return Ok(json!({
                        "pr": pr, "head": head, "merged": false, "rule": "G1",
                        "reason": format!("the AUTO-DELTA record returned HTTP {}; nothing else written", r.status),
                        "writes": [],
                    }))
                }
                Err(e) => {
                    return Ok(json!({
                        "pr": pr, "head": head, "merged": false, "rule": "G1",
                        "reason": format!("the AUTO-DELTA record failed ({e}); check the PR before running again"),
                        "writes": [],
                    }))
                }
            }
        }
    }

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
    audit_status::post(&status, permission, app, status_transport, clock).map_err(|e| e.message)?;
    writes.push(json!("status"));

    let url = format!("https://api.github.com/repos/{owner}/{repo}/pulls/{pr}/merge");
    let mut merged_sha = None;
    let mut last = 0;
    for attempt in 1..=MERGE_ATTEMPTS {
        let resp = match transport.send(
            &HttpRequest::new("PUT", &url).with_body(json!({"merge_method": "merge", "sha": head})),
        ) {
            Ok(resp) => resp,
            // The PUT may have merged before the error: read the effect back.
            Err(e) => match read_back(transport, &pr_url) {
                Some((true, sha)) => {
                    merged_sha = Some(sha.unwrap_or_default());
                    break;
                }
                Some((false, _)) => {
                    return Ok(json!({
                        "pr": pr, "head": head, "merged": false, "writes": writes,
                        "reason": format!("merge PUT failed ({e}); read back: not merged; no merge record posted"),
                    }))
                }
                None => {
                    return Ok(json!({
                        "pr": pr, "head": head, "merged": "uncertain", "writes": writes,
                        "reason": format!("merge PUT failed ({e}) and the pull request could not be read back; check it before anything else"),
                    }))
                }
            },
        };
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
    let sent = transport.send(
        &HttpRequest::new(
            "POST",
            format!("https://api.github.com/repos/{owner}/{repo}/issues/{pr}/comments"),
        )
        .with_body(json!({"body": record})),
    );
    // A transport error may still have posted the comment: say so instead of guessing.
    let (record_posted, record_error) = match sent {
        Ok(resp) if resp.status == 201 => (json!(true), None),
        Ok(resp) => (json!(false), Some(format!("HTTP {}", resp.status))),
        Err(e) => (json!("uncertain"), Some(e)),
    };
    if record_posted == json!(true) {
        writes.push(json!("record"));
    }
    let mut out = json!({
        "pr": pr, "head": head, "merged": true, "merge_commit": merge_sha,
        "verdict": verdict_id, "verdict_base": base, "overlap": overlap,
        "record_posted": record_posted, "writes": writes,
    });
    if let Some(e) = record_error {
        out["record_error"] = json!(e);
    }
    Ok(out)
}
