//! Flow control (A13 #2279): WIP caps, the integration-gate verdict, and the read
//! side of flow metrics. Everything here is GET-only — status posting, merging and
//! ruleset changes stay with the leader flow (decision 06:02, 2026-11-09).
//!
//! GitHub is read through the injected forge transport; tests replay recorded JSON.
//! gh call count is minimised: the open-PR list, the PR, the commit list, the two
//! compare calls and the head's check runs are the only reads a verdict performs.
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// The config key names (the numbers themselves are user-specified defaults, not
/// measured baselines — the issue requires them marked as such).
pub const CAP_LANE: &str = "lane_open_pr_cap";
pub const CAP_TOTAL: &str = "total_open_pr_cap";
/// Defaults are the user-approved values (2026-11-09 06:02Z decision: 1 audit-
/// pending PR per lane, 8 total open) — not measured baselines.
pub const DEFAULT_LANE_CAP: u64 = 1;
pub const DEFAULT_TOTAL_CAP: u64 = 8;

fn api(owner: &str, repo: &str, tail: &str) -> String {
    format!("https://api.github.com/repos/{owner}/{repo}/{tail}")
}

/// One open pull request as the WIP check sees it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OpenPr {
    pub number: i64,
    pub head_ref: String,
    pub author: String,
    pub draft: bool,
}

/// One commit of a PR as the gate verdict sees it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PrCommit {
    pub sha: String,
    pub parents: usize,
    pub message: String,
}

/// Marker forms that count as a claim line (06:02: lane-authored commits and the PR
/// body must carry `Moai-Claim: <issue>#<comment-id>`).
fn has_claim_line(text: &str) -> bool {
    text.contains("Moai-Claim: ")
}

fn fetch_open_pulls<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
) -> Result<Vec<OpenPr>> {
    let mut out = Vec::new();
    let mut url = Some(api(owner, repo, "pulls?state=open&per_page=100"));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 10 {
            return Err("flow open pull list page limit exceeded (10)".to_string());
        }
        let resp = client.get(&u, None)?;
        let arr = resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default();
        for p in &arr {
            out.push(OpenPr {
                number: p["number"].as_i64().unwrap_or(0),
                head_ref: p["head"]["ref"].as_str().unwrap_or_default().into(),
                author: p["user"]["login"].as_str().unwrap_or_default().into(),
                draft: p["draft"].as_bool().unwrap_or(false),
            });
        }
        url = resp
            .header("link")
            .and_then(|l| crate::forge::policy::next_link(Some(l)));
    }
    Ok(out)
}

pub(crate) fn fetch_pr<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    number: i64,
) -> Result<Value> {
    let resp = client.get(&api(owner, repo, &format!("pulls/{number}")), None)?;
    resp.body
        .clone()
        .ok_or_else(|| "pr response carried no body".to_string())
}

fn fetch_base_tip<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    base_ref: &str,
) -> Result<String> {
    let resp = client.get(&api(owner, repo, &format!("commits/{base_ref}")), None)?;
    resp.body
        .as_ref()
        .and_then(|b| b["sha"].as_str().map(str::to_string))
        .ok_or_else(|| "base branch response carried no sha".to_string())
}

/// ahead_by: how many commits `head` is ahead of `base` (0 when base is not an
/// ancestor — the compare endpoint reports diverged with both counts).
fn fetch_compare<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    base: &str,
    head: &str,
) -> Result<(u64, u64, Vec<String>)> {
    let resp = client.get(&api(owner, repo, &format!("compare/{base}...{head}")), None)?;
    let body = resp.body.clone().unwrap_or(Value::Null);
    let ahead = body["ahead_by"].as_u64().unwrap_or(u64::MAX);
    let behind = body["behind_by"].as_u64().unwrap_or(u64::MAX);
    let files: Vec<String> = body["files"]
        .as_array()
        .cloned()
        .unwrap_or_default()
        .iter()
        .filter_map(|f| f["filename"].as_str().map(str::to_string))
        .collect();
    Ok((ahead, behind, files))
}

fn fetch_pr_commits<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    number: i64,
) -> Result<Vec<PrCommit>> {
    let mut out = Vec::new();
    let mut url = Some(api(
        owner,
        repo,
        &format!("pulls/{number}/commits?per_page=100"),
    ));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 10 {
            return Err("flow pr commit page limit exceeded (10)".to_string());
        }
        let resp = client.get(&u, None)?;
        let arr = resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .cloned()
            .unwrap_or_default();
        for c in &arr {
            out.push(PrCommit {
                sha: c["sha"].as_str().unwrap_or_default().to_string(),
                parents: c["parents"].as_array().map(|a| a.len()).unwrap_or(0),
                message: c["commit"]["message"]
                    .as_str()
                    .unwrap_or_default()
                    .to_string(),
            });
        }
        url = resp
            .header("link")
            .and_then(|l| crate::forge::policy::next_link(Some(l)));
    }
    Ok(out)
}

/// The head's check runs (`commits/<sha>/check-runs`), paginated with the same 10-page
/// cap as the other list reads. This is the CI judgment source (#2335): rc-ci reports
/// as a check run and never lands in the combined commit status, whose `state` also
/// carries unrelated contexts (a Vercel deployment failure poisoned PR #2333's
/// combined status while rc-ci was green). The combined status endpoint is no longer
/// read at all.
fn fetch_check_runs<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    sha: &str,
) -> Result<Vec<Value>> {
    let mut out = Vec::new();
    let mut url = Some(api(
        owner,
        repo,
        &format!("commits/{sha}/check-runs?per_page=100"),
    ));
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > 10 {
            return Err("flow check runs page limit exceeded (10)".to_string());
        }
        let resp = client.get(&u, None)?;
        let arr = resp
            .body
            .as_ref()
            .and_then(|b| b["check_runs"].as_array())
            .cloned()
            .unwrap_or_default();
        out.extend(arr);
        url = resp
            .header("link")
            .and_then(|l| crate::forge::policy::next_link(Some(l)));
    }
    Ok(out)
}

/// Every required name must exist, and all returned runs with that name must have
/// concluded `success`. A failed or pending duplicate cannot be hidden by list order.
/// Unrelated names are ignored. An empty policy blocks.
fn required_ci_judgment(runs: &[Value], required: &[String]) -> (bool, String) {
    if required.is_empty() {
        return (false, "no required checks configured".to_string());
    }
    for name in required {
        let mut found = false;
        for r in runs
            .iter()
            .filter(|r| r["name"].as_str() == Some(name.as_str()))
        {
            found = true;
            let status = r["status"].as_str().unwrap_or_default();
            let conclusion = r["conclusion"].as_str().unwrap_or_default();
            if conclusion == "success" {
                continue;
            }
            if conclusion.is_empty() {
                return (false, format!("{name}: {status}, pending"));
            }
            return (false, format!("{name}: {conclusion}"));
        }
        if !found {
            return (false, format!("{name} check run not found"));
        }
    }
    (true, format!("{}: success", required.join(", ")))
}

// ---- WIP cap (assignment-refusal signal; the refusal itself is the leader's act) ----

fn count_by_author(prs: &[OpenPr], author: &str) -> u64 {
    prs.iter().filter(|p| p.author == author).count() as u64
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct WipVerdict {
    pub lane: String,
    pub lane_open: u64,
    pub total_open: u64,
    pub lane_cap: u64,
    pub total_cap: u64,
    pub exceeded: bool,
}

/// WIP cap check (read-only): one open-PR list call, then the lane count and the
/// total are compared with the caps. `exceeded` is the assignment-refusal signal.
pub fn wip_check<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    lane: &str,
    lane_cap: u64,
    total_cap: u64,
) -> Result<WipVerdict> {
    let prs = fetch_open_pulls(client, owner, repo)?;
    let lane_open = count_by_author(&prs, lane);
    let total_open = prs.len() as u64;
    Ok(WipVerdict {
        lane: lane.to_string(),
        lane_open,
        total_open,
        lane_cap,
        total_cap,
        exceeded: lane_open >= lane_cap || total_open >= total_cap,
    })
}

pub fn wip_check_json<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    lane: &str,
    lane_cap: u64,
    total_cap: u64,
) -> Result<Value> {
    let v = wip_check(client, owner, repo, lane, lane_cap, total_cap)?;
    Ok(json!({
        "schema_version": SCHEMA_VERSION,
        "lane": v.lane,
        "lane_open": v.lane_open,
        "total_open": v.total_open,
        "lane_cap": v.lane_cap,
        "total_cap": v.total_cap,
        "exceeded": v.exceeded,
        "verdict": if v.exceeded { "refuse_assignment" } else { "allow" },
        "writes": false,
    }))
}

// ---- integration gate verdict (06:02 rules, read-only) ----

/// One gate check with its outcome.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct GateCheck {
    pub check: &'static str,
    pub pass: bool,
    pub detail: String,
}

/// The integration gate verdict (verdict only — status posting and merging stay on
/// the leader path). Checks:
/// tip containment (behind_by 0 — a diverged branch fails even with ahead_by > 0),
/// files changed since the audited head (advisory count), claim lines on lane-authored
/// non-merge commits, the PR body claim line, the required check runs' conclusions
/// (#2335: the combined commit status is not consulted), and a head re-read.
pub fn gate_verdict<T: Transport, C: crate::forge::policy::Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    number: i64,
    audit_head: &str,
    base_ref: &str,
    required_checks: &[String],
) -> Result<Value> {
    let pr = fetch_pr(client, owner, repo, number)?;
    let head = pr["head"]["sha"]
        .as_str()
        .ok_or_else(|| "pr response carried no head sha".to_string())?
        .to_string();
    let body = pr["body"].as_str().unwrap_or_default().to_string();

    let tip = fetch_base_tip(client, owner, repo, base_ref)?;
    // #2335: containment requires the tip to be fully merged — behind_by > 0 means the
    // branch diverged and the tip's commits are missing, whatever ahead_by says.
    let (ahead_by, behind_by, _files) = fetch_compare(client, owner, repo, &tip, &head)?;
    let tip_contained = behind_by == 0 && (ahead_by > 0 || tip == head);
    let checks0 = GateCheck {
        check: "tip_contained",
        pass: tip_contained,
        detail: format!("tip {tip} ahead_by {ahead_by} behind_by {behind_by}"),
    };

    let changed_since_audit = head != audit_head;
    let (audit_ahead, _b2, changed_files) = if changed_since_audit {
        fetch_compare(client, owner, repo, audit_head, &head)?
    } else {
        (0u64, 0u64, Vec::new())
    };
    let checks1 = GateCheck {
        check: "changed_since_audit",
        pass: !changed_since_audit || changed_files.is_empty(),
        detail: format!(
            "audit {audit_head} vs head: {} file(s) differ",
            changed_files.len()
        ),
    };
    let _ = audit_ahead;

    // Claim lines: lane-authored (non-merge) commits must each carry one; merge
    // commits are records and are exempt (06:02).
    let commits = fetch_pr_commits(client, owner, repo, number)?;
    let no_claim: Vec<String> = commits
        .iter()
        .filter(|c| c.parents <= 1 && !has_claim_line(&c.message))
        .map(|c| c.sha.clone())
        .collect();
    let checks2 = GateCheck {
        check: "commit_claim_lines",
        pass: no_claim.is_empty(),
        detail: format!(
            "{} lane commit(s) checked; missing claim: {}",
            commits.len(),
            if no_claim.is_empty() {
                "none".into()
            } else {
                no_claim.join(", ")
            }
        ),
    };

    let body_claim = has_claim_line(&body);
    let checks3 = GateCheck {
        check: "pr_body_claim",
        pass: body_claim,
        detail: if body_claim {
            "claim line present".into()
        } else {
            "PR body carries no Moai-Claim line".into()
        },
    };

    // #2335: CI is judged from the required check runs' conclusions — the combined
    // commit status is neither consulted nor able to mask or poison the result. A read
    // failure of the check-runs endpoint itself is `unknown`, never a pass.
    let (ci_ok, ci_detail) = match fetch_check_runs(client, owner, repo, &head) {
        Ok(runs) => required_ci_judgment(&runs, required_checks),
        Err(_e) => (false, "check runs unavailable".to_string()),
    };
    let checks4 = GateCheck {
        check: "ci_status",
        pass: ci_ok,
        detail: format!("required check runs: {ci_detail}"),
    };

    let pr_again = fetch_pr(client, owner, repo, number)?;
    let head_stable = pr_again["head"]["sha"].as_str() == Some(head.as_str());
    let checks5 = GateCheck {
        check: "head_recheck",
        pass: head_stable,
        detail: format!("head {head} re-read"),
    };

    let all = [checks0, checks1, checks2, checks3, checks4, checks5];
    let pass = all.iter().all(|c| c.pass);
    let verdict = if pass { "ok" } else { "blocked" };
    let reasons: Vec<String> = all
        .iter()
        .filter(|c| !c.pass)
        .map(|c| format!("{}: {}", c.check, c.detail))
        .collect();
    Ok(json!({
        "schema_version": SCHEMA_VERSION,
        "pr": number,
        "head": head,
        "audit_head": audit_head,
        "tip": tip,
        "changed_since_audit": changed_since_audit,
        "changed_files": changed_files,
        "checks": all.iter().map(|c| json!({
            "check": c.check, "pass": c.pass, "detail": c.detail,
        })).collect::<Vec<_>>(),
        "verdict": verdict,
        "reasons": reasons,
        "writes": false,
    }))
}
