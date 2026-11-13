//! Read-only merge train projection (#2388 AC-1c). Order comes exclusively from a
//! caller-pinned remote ledger comment, never from PR age or a new local store.
use crate::forge::{policy::Clock, transport::Transport, Client};
use crate::lane_status::{self, IssueData, LaneStatusRequest, PrData, Snapshot};
use serde::Deserialize;
use serde_json::{json, Value};
use std::{collections::BTreeSet, path::Path};

type Result<T> = std::result::Result<T, String>;
const API: &str = "https://api.github.com";
const MAX_ROWS: usize = 50;
const MAX_PAGES: usize = 50;

#[derive(Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct QueueQueryRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub base_branch: String,
    pub ledger_issue: i64,
    pub ledger_comment_id: i64,
}

fn validate(r: &QueueQueryRequest) -> Result<()> {
    let segment = |s: &str| {
        !s.is_empty()
            && s.bytes()
                .all(|c| c.is_ascii_alphanumeric() || b"-_.".contains(&c))
            && s != "."
            && s != ".."
    };
    if r.schema_version != 1
        || r.ledger_issue <= 0
        || r.ledger_comment_id <= 0
        || !segment(&r.owner)
        || !segment(&r.repo)
        || r.base_branch.is_empty()
        || r.base_branch.chars().any(char::is_control)
    {
        return Err("invalid queue query request".into());
    }
    Ok(())
}

/// The existing leader ledger's Markdown table: `순서 | PR` (or `position | PR`).
/// Refuse ambiguity, duplicate PRs, gaps and malformed rows instead of inventing order.
pub fn parse_order(body: &str) -> Result<Vec<i64>> {
    let mut order = Vec::new();
    let (mut found, mut active, mut separator) = (false, false, false);
    let mut fence: Option<(u8, usize)> = None;
    for line in body.lines() {
        let line = line.trim();
        let marker = line.as_bytes().first().copied();
        let count = marker.map_or(0, |m| line.bytes().take_while(|b| *b == m).count());
        if let Some((opened, length)) = fence {
            // A different marker, shorter run or trailing info is code content.
            if marker == Some(opened) && count >= length && line[count..].is_empty() {
                fence = None;
            }
            continue;
        }
        if matches!(marker, Some(b'`' | b'~'))
            && count >= 3
            && (marker != Some(b'`') || !line[count..].contains('`'))
        {
            fence = marker.map(|m| (m, count));
            continue;
        }
        if !line.starts_with('|') {
            active = false;
            continue;
        }
        let cells: Vec<&str> = line.trim_matches('|').split('|').map(str::trim).collect();
        if cells.len() >= 2 && matches!(cells[0], "순서" | "position") && cells[1] == "PR" {
            if found {
                return Err("queue ledger contains multiple order tables".into());
            }
            found = true;
            active = true;
            separator = true;
            continue;
        }
        if !active {
            continue;
        }
        if separator {
            if cells.len() < 2
                || !cells.iter().all(|c| {
                    c.trim_matches(':').len() >= 3 && c.trim_matches(':').bytes().all(|b| b == b'-')
                })
            {
                return Err("queue ledger table separator is invalid".into());
            }
            separator = false;
            continue;
        }
        let pos = cells.first().and_then(|s| s.parse::<usize>().ok());
        let pr = cells
            .get(1)
            .and_then(|s| s.split_whitespace().next())
            .and_then(|s| s.strip_prefix('#'))
            .and_then(|s| s.parse::<i64>().ok());
        let Some(pr) = pr.filter(|n| *n > 0) else {
            return Err("queue ledger PR is invalid".into());
        };
        if pos != Some(order.len() + 1) || order.contains(&pr) || order.len() == MAX_ROWS {
            return Err("queue ledger needs unique PRs in positions 1..50 without gaps".into());
        }
        order.push(pr);
    }
    if !found || separator {
        return Err("queue ledger order table is missing or incomplete".into());
    }
    Ok(order)
}

fn one<T: Transport, C: Clock>(client: &Client<'_, T, C>, url: &str) -> Result<Value> {
    let (v, _) = crate::queue::get_json(client, url)?;
    if !v.is_object() {
        return Err("queue query expected an API object".into());
    }
    Ok(v)
}

fn pages<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    url: String,
    key: Option<&str>,
) -> Result<Vec<Value>> {
    let endpoint = url.split('?').next().unwrap_or_default().to_string();
    let mut next = Some(url);
    let mut out = Vec::new();
    let mut seen = BTreeSet::new();
    while let Some(url) = next {
        if url.split('?').next() != Some(endpoint.as_str()) {
            return Err("queue pagination changed endpoint".into());
        }
        if seen.len() == MAX_PAGES || !seen.insert(url.clone()) {
            return Err("queue query pagination limit or cycle".into());
        }
        let (v, link) = crate::queue::get_json(client, &url)?;
        let a = key
            .map_or(&v, |k| &v[k])
            .as_array()
            .ok_or("queue query expected an API array")?;
        out.extend(a.iter().cloned());
        next = link;
    }
    Ok(out)
}

fn sha(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// Only the exact first-line PR, head and base verdict grammar counts. Reuse lane-status's
/// projection after filtering; incidental quotes and PASS-with-conditions cannot become PASS.
fn exact_comments(raw: Vec<Value>, pr: i64, head: &str, base: &str) -> Vec<lane_status::Comment> {
    lane_status::read_comments(raw)
        .into_iter()
        .filter(|c| {
            let words: Vec<_> = c
                .body
                .lines()
                .next()
                .unwrap_or_default()
                .split_whitespace()
                .collect();
            words.len() == 10
                && words[0] == "AUDIT"
                && words[1] == format!("#{pr}")
                && words[2].starts_with('(')
                && words[2].ends_with(')')
                && words[3] == "head"
                && words[4] == head
                && words[5] == "base"
                && words[6] == base
                && words[7] == "—"
                && words[8] == "verdict:"
                && matches!(words[9], "PASS" | "FAIL")
        })
        .map(|mut c| {
            // The legacy projection reads the first `verdict:` occurrence. Replace the
            // opaque key so only the already validated terminal verdict can be read.
            let verdict = c
                .body
                .lines()
                .next()
                .unwrap()
                .split_whitespace()
                .last()
                .unwrap();
            c.body =
                format!("AUDIT #{pr} (queue-query) head {head} base {base} — verdict: {verdict}");
            c
        })
        .collect()
}

/// GET-only query. `project` is used solely for the existing read-only worktree listing.
pub fn query_cli<T: Transport, C: Clock>(
    project: &Path,
    req: QueueQueryRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    validate(&req)?;
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let repo = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let ledger = one(
        &client,
        &format!("{repo}/issues/comments/{}", req.ledger_comment_id),
    )?;
    if ledger["id"].as_i64() != Some(req.ledger_comment_id)
        || ledger["issue_url"] != format!("{repo}/issues/{}", req.ledger_issue)
    {
        return Err("queue ledger comment does not belong to the requested Issue".into());
    }
    let body = ledger["body"]
        .as_str()
        .ok_or("queue ledger comment has no body")?;
    let order = parse_order(body)?;
    let worktrees = crate::worktree::worktrees(project)?;
    let mut rows = Vec::new();
    let mut preceding = Vec::new();
    for (index, number) in order.into_iter().enumerate() {
        let pr = one(&client, &format!("{repo}/pulls/{number}"))?;
        let head = pr["head"]["sha"]
            .as_str()
            .filter(|s| sha(s))
            .ok_or("queue PR has no valid head")?;
        let base = pr["base"]["sha"]
            .as_str()
            .filter(|s| sha(s))
            .ok_or("queue PR has no valid base")?;
        if pr["number"].as_i64() != Some(number)
            || pr["base"]["ref"] != req.base_branch
            || !matches!(pr["state"].as_str(), Some("open" | "closed"))
        {
            return Err("queue PR number, state or base does not match".into());
        }
        let comments = pages(
            &client,
            format!("{repo}/issues/{number}/comments?per_page=100"),
            None,
        )?;
        let issue_number = crate::leader::refs_issue(pr["body"].as_str().unwrap_or_default());
        let mut issue = IssueData {
            number: issue_number.unwrap_or(0),
            open: true,
            body: String::new(),
            comments: Vec::new(),
        };
        if let Some(n) = issue_number {
            let data = one(&client, &format!("{repo}/issues/{n}"))?;
            if !matches!(data["state"].as_str(), Some("open" | "closed")) {
                return Err("queue Issue has no valid state".into());
            }
            issue.open = data["state"] == "open";
            issue.body = data["body"].as_str().unwrap_or_default().into();
            issue.comments = lane_status::read_comments(pages(
                &client,
                format!("{repo}/issues/{n}/comments?per_page=100"),
                None,
            )?);
        }
        let blockers: BTreeSet<_> = crate::queue::refs_on(&issue.body, "Blocked-By:")
            .into_iter()
            .filter(|n| *n > 0 && *n != issue.number)
            .collect();
        if blockers.len() > MAX_ROWS {
            return Err("queue blocker limit exceeded (50)".into());
        }
        let mut open_blockers = Vec::new();
        for n in blockers {
            let data = one(&client, &format!("{repo}/issues/{n}"))?;
            match data["state"].as_str() {
                Some("open") => open_blockers.push(n),
                Some("closed") => {}
                _ => return Err("queue blocker has no valid state".into()),
            }
        }
        // The lane-status ownership resolver keeps expired claims visible without taking over.
        let status_req = LaneStatusRequest {
            schema_version: 1,
            owner: req.owner.clone(),
            repo: req.repo.clone(),
            base_branch: req.base_branch.clone(),
            issues: issue_number.into_iter().collect(),
            lanes: vec![],
            pending_decisions: vec![],
        };
        let snap = Snapshot {
            issues: vec![issue.clone()],
            other_open: Default::default(),
            prs: vec![],
            worktrees: Ok(vec![]),
            now: clock.now(),
        };
        let status = lane_status::build(&status_req, &snap);
        let holder = &status["rows"][0];
        let mut audit_issue = issue.clone();
        audit_issue.comments = exact_comments(
            issue
                .comments
                .iter()
                .map(|c| json!({"id": c.id, "body": c.body}))
                .collect(),
            number,
            head,
            base,
        );
        let audit_pr = PrData {
            number,
            head: head.into(),
            branch: pr["head"]["ref"].as_str().unwrap_or_default().into(),
            issue: issue_number,
            commits: vec![],
            claim_ids: vec![],
            comments: exact_comments(comments, number, head, base),
        };
        let audit = lane_status::audit_value(&audit_issue, &audit_pr);
        let checks = pages(
            &client,
            format!("{repo}/commits/{head}/check-runs?per_page=100&filter=all"),
            Some("check_runs"),
        )?;
        let mut ci = serde_json::Map::new();
        for name in ["ci", "windows-build"] {
            let newest = checks
                .iter()
                .filter(|c| {
                    // #2563: `ci` also reads runs under its former name `rc-ci`; the
                    // newest run under either spelling decides.
                    c["name"]
                        .as_str()
                        .is_some_and(|actual| crate::flow::check_name_matches(name, actual))
                        && c["head_sha"] == head
                        && c["app"]["slug"] == "github-actions"
                })
                .max_by_key(|c| c["id"].as_i64().unwrap_or(0));
            let state = newest
                .map(|c| {
                    if c["status"] == "completed" {
                        c["conclusion"].as_str().unwrap_or("unknown")
                    } else {
                        c["status"].as_str().unwrap_or("unknown")
                    }
                })
                .unwrap_or("not_run");
            ci.insert(
                name.into(),
                json!({"state": state, "check_id": newest.map(|c| &c["id"]), "head": head}),
            );
        }
        let mut blocked = Vec::new();
        let open = pr["state"] == "open";
        if !open {
            blocked.push("pr_closed");
        }
        if !preceding.is_empty() {
            blocked.push("waiting_predecessor");
        }
        if !open_blockers.is_empty() {
            blocked.push("waiting_prerequisite");
        }
        if pr["draft"] != false {
            blocked.push("draft_or_unknown");
        }
        if holder["session"].is_null() {
            blocked.push("ownership_unknown");
        } else if holder["ownership"]["live"] != true {
            blocked.push("claim_expired_or_handoff");
        }
        match audit["verdict_on_head"]["verdict"].as_str() {
            Some("PASS") => {}
            Some("FAIL") => blocked.push("audit_failed"),
            _ => blocked.push("audit_missing_or_stale"),
        }
        // #2554 (user decision U10): on a PR `windows-build` does not run (it runs after
        // merge). Its job-level `if` leaves a check run with conclusion `skipped` (or none at
        // all, `not_run`); neither blocks. Any other state that is not `success` (failure,
        // cancelled, still running, ...) still does. `ci` must be `success` on the head.
        for (name, reason) in [
            ("ci", "ci_not_success"),
            ("windows-build", "windows_build_not_success"),
        ] {
            let state = &ci[name]["state"];
            let absent_ok = name == "windows-build" && (state == "not_run" || state == "skipped");
            if state != "success" && !absent_ok {
                blocked.push(reason);
            }
        }
        match pr["mergeable_state"].as_str() {
            Some("clean" | "unstable" | "has_hooks") => {}
            Some("dirty") => blocked.push("merge_conflict"),
            Some("behind") => blocked.push("base_behind"),
            Some("blocked") => blocked.push("forge_gate_blocked"),
            _ => blocked.push("mergeability_unknown"),
        }
        // Detect a moving head/base during a multi-request observation. Do not publish a
        // mixed snapshot as current; the caller may explicitly query again.
        let after = one(&client, &format!("{repo}/pulls/{number}"))?;
        if after["number"] != pr["number"]
            || after["base"]["ref"] != req.base_branch
            || after["head"]["ref"] != pr["head"]["ref"]
            || after["head"]["sha"] != head
            || after["base"]["sha"] != base
            || after["state"] != pr["state"]
        {
            return Err("queue PR changed during query; query again".into());
        }
        rows.push(json!({"position": index+1, "pr": number, "issue": issue_number, "session": holder["session"], "ownership": holder["ownership"], "head": head, "base": base, "verdict": audit["verdict_on_head"], "ci": ci, "mergeable_state": pr["mergeable_state"], "blocked_reasons": blocked, "preceding_prs": preceding, "open_blockers": open_blockers}));
        if open {
            preceding.push(number);
        }
    }
    let ledger_after = one(
        &client,
        &format!("{repo}/issues/comments/{}", req.ledger_comment_id),
    )?;
    if ledger_after["body"] != ledger["body"] || ledger_after["updated_at"] != ledger["updated_at"]
    {
        return Err("queue ledger changed during query; query again".into());
    }
    Ok(
        json!({"rows": rows, "ledger": {"issue": req.ledger_issue, "comment_id": req.ledger_comment_id, "body_digest": crate::digest(body.as_bytes()), "updated_at": ledger["updated_at"]}, "observed_at": clock.now(), "local_worktrees": worktrees.len(), "writes": "none", "note": "read-only observation; not an integration authorization; ledger freshness is caller-selected"}),
    )
}
