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
//! local state file is written. A head that is not in the local repository is skipped;
//! with `fetch` (#2906 S3) the pass first fetches the base branch and every open head.
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
    /// Fetch the base branch and every open pull request's head from `origin` before the
    /// pass (#2906 S3); off by default.
    pub fetch: bool,
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

/// What a dry run decides for one pull request: the gate's verdict and the action the pass
/// would take, from reads only (#2860 D2). It does not re-read the remote base and head the way
/// [`crate::gate_merge::merge`] does just before its first write, so a `would_merge` row still
/// meets that check in a real pass.
fn dry_decision<T: Transport>(
    one: &Target<'_>,
    t: &T,
    api: &str,
    head: &str,
) -> Result<(String, String, String)> {
    let snapshot = crate::gate_collect::collect(one, t)?;
    match crate::gate::check(&snapshot) {
        Ok(_) => {
            let action = if crate::gate::accepted_auto_delta(&snapshot).is_some() {
                "would_merge_auto_delta"
            } else {
                "would_merge"
            };
            Ok((action.into(), String::new(), String::new()))
        }
        Err(stop) => {
            let action = if stop.rule == "G6" && stop.reason.starts_with("overlap in non-doc file")
            {
                if marked(t, api, one.pr, &marker("update-branch", head))? {
                    "already_updated"
                } else {
                    "would_update_branch"
                }
            } else if stop.rule == "G2" {
                if marked(t, api, one.pr, &marker("auditor", head))? {
                    "already_asked"
                } else {
                    "would_ask_auditor"
                }
            } else {
                "stopped"
            };
            Ok((action.into(), stop.rule.to_string(), stop.reason))
        }
    }
}

/// One pass over the open pull requests, oldest number first (#2860 rule 5). Returns one row
/// per pull request. With `dry_run` it only reads and reports what it would do (#2860 D2).
pub fn once<T: Transport, S: Transport, C: Clock>(
    target: &LoopTarget<'_>,
    permission: &str,
    app: bool,
    dry_run: bool,
    t: &T,
    status_t: &S,
    clock: &C,
) -> Result<Value> {
    // A dry run writes nothing, so it needs no permission reference.
    if !dry_run && !crate::contract::permission_reference_ok(permission) {
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
    let mut open: Vec<&Value> = list
        .as_array()
        .ok_or("pull request list is not a list")?
        .iter()
        .collect();
    // #2860 rule 5: the oldest pull request first, so a long-waiting one is not passed over by
    // newer ones every pass (GitHub lists newest first by default).
    open.sort_by_key(|p| p["number"].as_u64().unwrap_or(u64::MAX));
    let total = open.len();
    // #2906 S3: `--fetch` brings the base branch and every open head in first, the way the
    // shell loop fetched per pull request. A failed fetch is reported and the pass goes on
    // with what is local (a head still missing is skipped below).
    let fetched = target.fetch.then(|| {
        let prs: Vec<u64> = open.iter().filter_map(|p| p["number"].as_u64()).collect();
        let refspecs = fetch_refspecs(target.base_branch, &prs);
        match fetch(target.dir, &refspecs) {
            Ok(()) => json!({"ok": true, "refspecs": refspecs}),
            Err(e) => json!({"ok": false, "refspecs": refspecs, "error": e}),
        }
    });
    for (seen, p) in open.into_iter().enumerate() {
        // #2860 rule 4 (leader decision 6870150891): drafts are gated too; a passing draft is
        // marked ready by gate-merge just before its merge.
        let pr = p["number"].as_u64().ok_or("pull request without number")?;
        // One static progress line per pull request (#2806); silent without a stderr sink.
        crate::progress::note_msg(
            crate::ui_text::Msg::CheckingPullRequest,
            &[
                ("pr", &pr.to_string()),
                ("at", &(seen + 1).to_string()),
                ("total", &total.to_string()),
            ],
        );
        let head = p["head"]["sha"].as_str().unwrap_or("").to_string();
        let one = Target {
            dir: target.dir,
            owner: target.owner,
            repo: target.repo,
            pr,
            tip: target.tip,
            review_confirmed_prs: target.review_confirmed_prs.clone(),
        };
        if dry_run {
            let row = match dry_decision(&one, t, &api, &head) {
                Ok((action, rule, reason)) => {
                    json!({"pr": pr, "head": head, "outcome": action, "rule": rule, "reason": reason})
                }
                Err(e) => json!({"pr": pr, "head": head, "outcome": "skipped", "reason": e}),
            };
            rows.push(row);
            continue;
        }
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
    let mut out = json!({"dry_run": dry_run, "pull_requests": rows});
    if let Some(f) = fetched {
        out["fetch"] = f;
    }
    Ok(out)
}

/// Explicit refspecs into named remote refs, so nothing depends on FETCH_HEAD (#2906 S3).
fn fetch_refspecs(base: &str, prs: &[u64]) -> Vec<String> {
    let mut specs = vec![format!("+refs/heads/{base}:refs/remotes/origin/{base}")];
    specs.extend(
        prs.iter()
            .map(|n| format!("+refs/pull/{n}/head:refs/remotes/origin/pull/{n}")),
    );
    specs
}

/// One `git fetch origin` with `refspecs`; FETCH_HEAD is not written. The operator's git
/// configuration stays in effect (credentials, URL rewrites), as in the shell loop.
fn fetch(dir: &Path, refspecs: &[String]) -> Result<()> {
    let mut cmd = std::process::Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .env("GIT_TERMINAL_PROMPT", "0")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(["fetch", "--quiet", "--no-write-fetch-head", "origin"])
        .args(refspecs)
        .current_dir(dir)
        .stdin(std::process::Stdio::null())
        .output()
        .map_err(|e| format!("cannot execute git: {e}"))?;
    if out.status.success() {
        Ok(())
    } else {
        Err(format!(
            "git fetch failed: {}",
            crate::identity::failure_summary(&out.stderr)
        ))
    }
}
