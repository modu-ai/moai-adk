//! `gate-loop --once` (Issue #2674 S4): one pass of what `target/leader/auto_loop.sh` did every
//! 90 seconds, for the leader (or a tool's scheduler) to call instead of a shell loop.
//!
//! For every open pull request into the integration branch it runs [`crate::gate_merge::merge`]
//! (which writes nothing unless the gate passes), then acts once per head on two stops:
//!
//! - G6 "overlap in non-doc file": update the branch (`expected_head_sha` pinned) so CI runs on
//!   the merged result, and leave one note for the auditor;
//! - G2 (the verdict's author holds no allowed claim): leave one note asking the claim holder.
//!
//! "Once per head" is kept in the pull request itself: each note carries a hidden marker
//! `<!-- moai:gate-loop <kind> head=<sha> -->`, and a marker already present means done. No
//! local state file is written. A head that is not in the local repository is skipped.
//!
//! Like gate-merge, everything goes through the operator's transport `t` except the status,
//! which goes through `status_t` (the audit App token when set; #2735 audit F3).

use crate::forge::policy::{next_link, Clock};
use crate::forge::transport::{HttpRequest, Transport};
use crate::gate_collect::Target;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Where one pass reads and writes.
pub struct LoopTarget<'a> {
    pub dir: &'a Path,
    pub owner: &'a str,
    pub repo: &'a str,
    /// Integration branch name (pull requests into it are gated).
    pub base_branch: &'a str,
    /// Local ref or SHA of its tip.
    pub tip: &'a str,
    pub review_confirmed_prs: Vec<u64>,
}

fn marker(kind: &str, head: &str) -> String {
    format!("<!-- moai:gate-loop {kind} head={head} -->")
}

fn get<T: Transport>(t: &T, url: &str) -> Result<(Value, Option<String>)> {
    let resp = t.send(&HttpRequest::new("GET", url))?;
    if resp.status != 200 {
        return Err(format!("GET {url} returned HTTP {}", resp.status));
    }
    let next = next_link(resp.header("link"));
    Ok((resp.body.unwrap_or(Value::Null), next))
}

/// Whether any comment on the pull request already carries `mark`.
fn marked<T: Transport>(t: &T, api: &str, pr: u64, mark: &str) -> Result<bool> {
    let mut url = Some(format!("{api}/issues/{pr}/comments?per_page=100"));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 30 {
            return Err("more than 30 comment pages; refused".into());
        }
        let (page, next) = get(t, &u)?;
        if page.as_array().is_some_and(|a| {
            a.iter()
                .any(|c| c["body"].as_str().is_some_and(|b| b.contains(mark)))
        }) {
            return Ok(true);
        }
        url = next;
    }
    Ok(false)
}

fn comment<T: Transport>(t: &T, api: &str, pr: u64, body: String) -> Result<()> {
    let resp = t.send(
        &HttpRequest::new("POST", format!("{api}/issues/{pr}/comments"))
            .with_body(json!({"body": body})),
    )?;
    if resp.status != 201 {
        return Err(format!("comment POST returned HTTP {}", resp.status));
    }
    Ok(())
}

/// One pass over the open pull requests. Returns one row per pull request.
pub fn once<T: Transport, S: Transport, C: Clock>(
    target: &LoopTarget<'_>,
    permission: &str,
    app: bool,
    t: &T,
    status_t: &S,
    clock: &C,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("a permission reference is required; nothing was read or written".into());
    }
    let api = format!(
        "https://api.github.com/repos/{}/{}",
        target.owner, target.repo
    );
    let (list, _) = get(
        t,
        &format!(
            "{api}/pulls?state=open&base={}&per_page=100",
            target.base_branch
        ),
    )?;
    let mut rows = Vec::new();
    for p in list.as_array().ok_or("pull request list is not a list")? {
        if p["draft"].as_bool() == Some(true) {
            continue;
        }
        let pr = p["number"].as_u64().ok_or("pull request without number")?;
        let head = p["head"]["sha"].as_str().unwrap_or("").to_string();
        let one = Target {
            dir: target.dir,
            owner: target.owner,
            repo: target.repo,
            pr,
            tip: target.tip,
            review_confirmed_prs: target.review_confirmed_prs.clone(),
        };
        let out = match crate::gate_merge::merge(&one, permission, app, t, status_t, clock) {
            Ok(out) => out,
            Err(e) => {
                rows.push(json!({"pr": pr, "outcome": "skipped", "reason": e}));
                continue;
            }
        };
        if out["merged"] == json!(true) {
            rows.push(json!({"pr": pr, "outcome": "merged", "merge_commit": out["merge_commit"]}));
            continue;
        }
        // A merge whose effect could not be read back: report it, act on nothing else.
        if out["merged"] == json!("uncertain") {
            rows.push(json!({"pr": pr, "outcome": "uncertain", "reason": out["reason"]}));
            continue;
        }
        let rule = out["rule"].as_str().unwrap_or("");
        let reason = out["reason"].as_str().unwrap_or("").to_string();
        let action = if rule == "G6" && reason.starts_with("overlap in non-doc file") {
            let mark = marker("update-branch", &head);
            if marked(t, &api, pr, &mark)? {
                "already_updated"
            } else {
                let resp = t.send(
                    &HttpRequest::new("PUT", format!("{api}/pulls/{pr}/update-branch"))
                        .with_body(json!({"expected_head_sha": head})),
                )?;
                if resp.status != 202 {
                    return Err(format!(
                        "update-branch returned HTTP {} for #{pr}",
                        resp.status
                    ));
                }
                comment(t, &api, pr, format!(
                    "리더(`moai-cli gate-loop`) — 게이트 중단 → **update-branch 실행**. 감사 PASS·CI 는 갖췄고 충돌도 없지만, 판정 base 이후 tip 이 이 PR 과 같은 코드 파일을 바꿨다({reason}). 합친 head 에서 CI 를 다시 돈다. 저자가 할 일은 없다. 감사자(이 PR 의 예약 보유 레인): 병합 커밋의 겹친 파일만 delta 재감사하고, 판정에 자기 예약 id 를 적는다.\n\n{mark}\n\n🗿 MoAI"
                ))?;
                "updated_branch"
            }
        } else if rule == "G2" {
            let mark = marker("auditor", &head);
            if marked(t, &api, pr, &mark)? {
                "already_asked"
            } else {
                comment(t, &api, pr, format!(
                    "리더(`moai-cli gate-loop`) — 게이트 중단: 마지막 판정을 쓴 레인이 이 PR 의 허용된 감사 예약을 가진 레인이 아니다({reason}). 허용되는 감사자는 이 PR 의 가장 낮은 `moai-audit-claim` 을 가진 레인, 또는 지금 head 를 적은 가장 낮은 예약을 가진 레인이다. 그 레인이 판정을 내고, 판정 첫 줄 다음에 자기 레인 이름과 예약 id 를 적어 달라.\n\n{mark}\n\n🗿 MoAI"
                ))?;
                "asked_auditor"
            }
        } else {
            "stopped"
        };
        rows.push(json!({"pr": pr, "outcome": action, "rule": rule, "reason": reason}));
    }
    Ok(json!({"pull_requests": rows}))
}
