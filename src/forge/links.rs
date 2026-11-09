//! Sub-issue and issue-dependency capability probe with a body-link fallback (#2172, B5).
//!
//! `forge-links-probe` asks GitHub, for one Issue, whether the sub-issues list
//! (`GET /repos/{o}/{r}/issues/{n}/sub_issues`) and the blocked-by list
//! (`GET /repos/{o}/{r}/issues/{n}/dependencies/blocked_by`) answer, with one `per_page=1`
//! GET each. Each feature is judged on its own:
//!
//! - `available` (2xx): the link is planned as API write intents (the POST endpoints). They are
//!   intents only — nothing is executed here; a write is a separate permissioned effect.
//! - `unavailable` (404, 410) or `blocked` (401/403, an exhausted rate limit, no response):
//!   the link is planned as Issue body lines `Parent: #<n>` / `Blocked-By: #<a>, #<b>`, and no
//!   API write is planned at all (REQ-B5-01).
//!
//! A probe failure is a result, not an error: only invalid input fails the operation. The
//! parser [`parse_body_links`] also reads this repository's own spellings `상위:` for the
//! parent and `없음` for an empty list.
//!
//! Limits (read 2026-11-09, not enforced here): up to 100 sub-issues per parent and eight
//! nesting levels (docs.github.com, "Adding sub-issues"); up to 50 issues for each dependency
//! relationship type (github.blog changelog 2025-08-21; not stated in the REST reference).
use super::policy::Clock;
use super::transport::{HttpRequest, Transport};
use super::{Client, Config, ForgeError, API, SCHEMA_VERSION};
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::Path;

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct LinksRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub issue_number: i64,
    #[serde(default)]
    pub parent: Option<i64>,
    #[serde(default)]
    pub blocked_by: Vec<i64>,
}

fn valid_name(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 100
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b"-_.".contains(&b))
        && s != "."
        && s != ".."
}

fn validate(req: &LinksRequest) -> Result<(), ForgeError> {
    if req.schema_version != SCHEMA_VERSION {
        return Err(ForgeError::of(
            "unsupported_protocol_version",
            "unsupported request schema_version",
        ));
    }
    if !valid_name(&req.owner) || !valid_name(&req.repo) {
        return Err(ForgeError::invalid("owner and repo must be GitHub names"));
    }
    let n = req.issue_number;
    if n <= 0 || req.parent.is_some_and(|p| p <= 0) || req.blocked_by.iter().any(|b| *b <= 0) {
        return Err(ForgeError::invalid("issue numbers must be positive"));
    }
    if req.parent == Some(n) || req.blocked_by.contains(&n) {
        return Err(ForgeError::invalid("issue links must be to other issues"));
    }
    if req.blocked_by.len() > 50 {
        return Err(ForgeError::invalid(
            "blocked_by must be at most 50 issues (one relationship type)",
        ));
    }
    let mut seen = req.blocked_by.clone();
    seen.sort_unstable();
    seen.dedup();
    if seen.len() != req.blocked_by.len() {
        return Err(ForgeError::invalid(
            "blocked_by must be free of repeated issues",
        ));
    }
    Ok(())
}

/// One feature's probe: `(status, http_status, error kind)`.
fn probe_one<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    url: String,
) -> (&'static str, Option<u16>, Option<&'static str>) {
    let req = HttpRequest {
        method: "GET".into(),
        url,
        headers: BTreeMap::new(),
        body: None,
    };
    match client.send_with_backoff(&req) {
        Ok(resp) if (200..300).contains(&resp.status) => ("available", Some(resp.status), None),
        Ok(resp) if resp.status == 404 || resp.status == 410 => {
            ("unavailable", Some(resp.status), Some("not_found"))
        }
        Ok(resp) => ("blocked", Some(resp.status), Some("transport")),
        Err(e) => ("blocked", None, Some(e.kind)),
    }
}

/// `Parent: #<n>` for a parent, `Blocked-By: #<a>, #<b>` for blockers.
pub fn render_body_links(parent: Option<i64>, blocked_by: &[i64]) -> (Vec<String>, Vec<String>) {
    let parent = parent
        .map(|p| format!("Parent: #{p}"))
        .into_iter()
        .collect();
    let blocked = if blocked_by.is_empty() {
        Vec::new()
    } else {
        let list: Vec<String> = blocked_by.iter().map(|b| format!("#{b}")).collect();
        vec![format!("Blocked-By: {}", list.join(", "))]
    };
    (parent, blocked)
}

/// Reads `Parent:` (or `상위:`) and `Blocked-By:` lines from an Issue body. `없음`/`none`, or
/// no `#<n>`, is an empty list. Only the first number of the parent line is taken.
pub fn parse_body_links(body: &str) -> (Option<i64>, Vec<i64>) {
    let numbers = |s: &str| -> Vec<i64> {
        s.split('#')
            .skip(1)
            .filter_map(|t| {
                let d: String = t.chars().take_while(|c| c.is_ascii_digit()).collect();
                d.parse().ok()
            })
            .collect()
    };
    let mut parent = None;
    let mut blocked = Vec::new();
    for line in body.lines() {
        let line = line.trim();
        if let Some(rest) = line
            .strip_prefix("Parent:")
            .or_else(|| line.strip_prefix("상위:"))
        {
            parent = parent.or_else(|| numbers(rest).first().copied());
        } else if let Some(rest) = line.strip_prefix("Blocked-By:") {
            blocked.extend(numbers(rest));
        }
    }
    (parent, blocked)
}

/// Probes both features and plans the links. GET only; nothing is written.
pub fn probe<T: Transport, C: Clock>(
    req: &LinksRequest,
    transport: &T,
    clock: &C,
) -> Result<Value, ForgeError> {
    validate(req)?;
    let client = Client::new(transport, clock, Config::default(), BTreeMap::new());
    let (o, r, n) = (&req.owner, &req.repo, req.issue_number);
    let base = format!("{API}/repos/{o}/{r}/issues");
    let (parent_lines, blocked_lines) = render_body_links(req.parent, &req.blocked_by);

    let feature = |url: String, intents: Vec<Value>, body_lines: Vec<String>| {
        let (status, http_status, kind) = probe_one(&client, url);
        let api = status == "available";
        json!({
            "status": status,
            "http_status": http_status,
            "error_kind": kind,
            "mode": if api { "api" } else { "body_fallback" },
            "api_write_intents": if api { intents } else { Vec::new() },
            "body_lines": if api { Vec::new() } else { body_lines },
        })
    };
    let sub_intents: Vec<Value> = req
        .parent
        .map(|p| {
            json!({"method": "POST", "path": format!("repos/{o}/{r}/issues/{p}/sub_issues"),
                   "body": {"sub_issue_id": null}, "sub_issue_number": n,
                   "needs": format!("the database id of #{n} as sub_issue_id")})
        })
        .into_iter()
        .collect();
    let dep_intents: Vec<Value> = req
        .blocked_by
        .iter()
        .map(|b| {
            json!({"method": "POST", "path": format!("repos/{o}/{r}/issues/{n}/dependencies/blocked_by"),
                   "body": {"issue_id": null}, "blocked_by_number": b,
                   "needs": format!("the database id of #{b} as issue_id")})
        })
        .collect();
    let sub_issues = feature(
        format!("{base}/{n}/sub_issues?per_page=1"),
        sub_intents,
        parent_lines,
    );
    let dependencies = feature(
        format!("{base}/{n}/dependencies/blocked_by?per_page=1"),
        dep_intents,
        blocked_lines,
    );
    Ok(json!({
        "schema_version": SCHEMA_VERSION,
        "issue": {"owner": o, "repo": r, "number": n},
        "sub_issues": sub_issues,
        "dependencies": dependencies,
        "executed": false,
        "notes": [
            "probe sends GET only; api_write_intents are plans, never executed here",
            "a feature that is unavailable or blocked plans body lines and no API write (REQ-B5-01)"
        ],
    }))
}

/// CLI entry (`forge.links_probe`). The project directory is not written.
pub fn links_probe_cli<T: Transport, C: Clock>(
    project: &Path,
    req: LinksRequest,
    transport: &T,
    clock: &C,
) -> Result<Value, String> {
    if !project.is_dir() {
        return Err(ForgeError::invalid("project must be a directory").into());
    }
    Ok(probe(&req, transport, clock)?)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn body_links_round_trip_and_accept_the_repository_spellings() {
        let (p, b) = render_body_links(Some(1978), &[1985, 1987]);
        let body = format!("intro\n{}\n{}\n", p[0], b[0]);
        assert_eq!(parse_body_links(&body), (Some(1978), vec![1985, 1987]));
        assert_eq!(
            parse_body_links("상위: #2154 (EP-B)\nBlocked-By: 없음\n"),
            (Some(2154), vec![])
        );
        assert_eq!(render_body_links(None, &[]), (vec![], vec![]));
    }
}
