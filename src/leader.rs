//! Leader supervision tick: `leader-tick` (rc.2 A3 #2162; SPEC
//! `docs/design-review/autonomy-modes-spec.md` §5, `AGENTS.local.md` § Assignment record and
//! supervision).
//!
//! One call reads every open pull request into the base branch and, per PR, its merge state,
//! the `moai/audit` status and the CI check run (`ci`, or `rc-ci` before #2563) on the current head, and the claim and
//! audit-verdict comments of the Issue it `Refs`. It compares that with the previous tick
//! (`<project>/.moai-cli/forge/leader-tick.json`) and reports only the PRs whose observation
//! changed (REQ-A3-01; an unchanged PR is not reported, so polling does not multiply reports).
//!
//! It is GET-only on GitHub. For a PR that needs a repair (a `FAIL` verdict on the current
//! head, a failed CI check, or a `dirty` merge state) it drafts an `assignee-hint` comment for
//! the first caller-supplied free lane that is not the PR's author (REQ-A3-02; ownership still
//! comes only from a claim comment) and a one-line repair Issue title. Posting either is the
//! leader's act through `effect-intend` → `gh` → `effect-confirm` (leader decision 2026-11-09).
//!
//! The interval comes from the profile key `leader.pr_check_interval_secs` (default 300, a
//! user-chosen value, not a measured one). The tick never sleeps: scheduling is the caller's.
//! Not done here: regenerating `assignments.json` (a follow-up), live GitHub runs (fixtures only).
use crate::claim::{parse_claim_line, resolve_winner, ClaimComment};
use crate::forge::policy::{next_link, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

const API: &str = "https://api.github.com";
/// User-chosen default (D-2, 2026-11-09); not measured.
pub const DEFAULT_INTERVAL_SECS: u64 = 300;
const STATE: &str = "forge/leader-tick.json";
const MAX_PAGES: u32 = 50;

/// `leader-tick` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct TickRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub base_branch: String,
    /// Free lanes in preference order (the leader reads them from its session list).
    #[serde(default)]
    pub lanes: Vec<String>,
}

/// What one tick saw for one PR; equality decides whether the PR is reported.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct Observation {
    pub head: String,
    pub issue: Option<i64>,
    pub mergeable_state: String,
    pub audit: String,
    pub ci: String,
    pub verdict: String,
    pub author: Option<String>,
}

#[derive(Debug, Default, Serialize, Deserialize, PartialEq, Eq)]
pub struct TickState {
    pub schema_version: u32,
    pub prs: BTreeMap<i64, Observation>,
}

fn get<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    url: &str,
) -> Result<(Value, Option<String>)> {
    let resp = client.get(url, None).map_err(String::from)?;
    let next = next_link(resp.header("link"));
    Ok((resp.body.unwrap_or(Value::Null), next))
}

fn pages<T: Transport, C: Clock>(client: &Client<'_, T, C>, first: String) -> Result<Vec<Value>> {
    let (mut url, mut out, mut n) = (Some(first), Vec::new(), 0);
    while let Some(u) = url {
        n += 1;
        if n > MAX_PAGES {
            return Err("leader tick page limit exceeded (50)".into());
        }
        let (body, next) = get(client, &u)?;
        out.extend(body.as_array().cloned().unwrap_or_default());
        url = next;
    }
    Ok(out)
}

/// The first `Refs #<n>` in a PR body.
pub(crate) fn refs_issue(body: &str) -> Option<i64> {
    let i = body.find("Refs #")?;
    let digits: String = body[i + 6..]
        .chars()
        .take_while(char::is_ascii_digit)
        .collect();
    digits.parse().ok()
}

/// The last `AUDIT … head <sha> … verdict: PASS|FAIL` line that names `head`.
fn verdict_for(comments: &[Value], head: &str) -> String {
    let mut found = "none".to_string();
    for c in comments {
        for line in c["body"].as_str().unwrap_or_default().lines() {
            if line.contains("AUDIT #") && line.contains(head) {
                if let Some(v) = line.split("verdict:").nth(1) {
                    let v = v.trim();
                    if v.starts_with("PASS") {
                        found = "PASS".into();
                    } else if v.starts_with("FAIL") {
                        found = "FAIL".into();
                    }
                }
            }
        }
    }
    found
}

/// The session of the smallest live claim comment (the owner, as 06-A decides it).
fn author(comments: &[Value], now: i64) -> Option<String> {
    let live: Vec<ClaimComment> = comments
        .iter()
        .filter_map(|c| {
            let mut line = parse_claim_line(c["body"].as_str()?)?;
            line.comment_id = c["id"].as_i64()?;
            parse_utc(&line.lease_until).filter(|t| *t > now)?;
            Some(line)
        })
        .collect();
    resolve_winner(&live).map(|w| w.session)
}

fn observe<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    req: &TickRequest,
    pr: &Value,
    now: i64,
) -> Result<Option<(i64, Observation)>> {
    let (Some(number), Some(head)) = (pr["number"].as_i64(), pr["head"]["sha"].as_str()) else {
        return Ok(None);
    };
    let repo = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let (detail, _) = get(client, &format!("{repo}/pulls/{number}"))?;
    let (status, _) = get(client, &format!("{repo}/commits/{head}/status"))?;
    let (runs, _) = get(client, &format!("{repo}/commits/{head}/check-runs"))?;
    let audit = status["statuses"]
        .as_array()
        .and_then(|a| a.iter().find(|s| s["context"] == "moai/audit"))
        .and_then(|s| s["state"].as_str())
        .unwrap_or("none")
        .to_string();
    let ci = runs["check_runs"]
        .as_array()
        // #2563: either spelling of the CI check; the newest run (largest id) decides,
        // so the answer does not depend on the order the API lists the runs in.
        .and_then(|a| {
            a.iter()
                .filter(|r| {
                    r["name"]
                        .as_str()
                        .is_some_and(|actual| crate::flow::check_name_matches("ci", actual))
                })
                .max_by_key(|r| r["id"].as_i64().unwrap_or(0))
        })
        .map(|r| {
            r["conclusion"]
                .as_str()
                .or(r["status"].as_str())
                .unwrap_or("none")
        })
        .unwrap_or("none")
        .to_string();
    let issue = refs_issue(pr["body"].as_str().unwrap_or_default());
    let (verdict, author) = match issue {
        Some(i) => {
            let comments = pages(client, format!("{repo}/issues/{i}/comments?per_page=100"))?;
            (verdict_for(&comments, head), author(&comments, now))
        }
        None => ("none".into(), None),
    };
    Ok(Some((
        number,
        Observation {
            head: head.into(),
            issue,
            mergeable_state: detail["mergeable_state"]
                .as_str()
                .unwrap_or("unknown")
                .into(),
            audit,
            ci,
            verdict,
            author,
        },
    )))
}

fn repair_reasons(o: &Observation) -> Vec<&'static str> {
    let mut r = Vec::new();
    if o.verdict == "FAIL" {
        r.push("audit verdict FAIL on the current head");
    }
    if ["failure", "timed_out"].contains(&o.ci.as_str()) {
        r.push("ci failed");
    }
    if o.mergeable_state == "dirty" {
        r.push("merge conflict with the base");
    }
    r
}

/// The hint draft for a PR that needs a repair: the first free lane that is not the author.
pub fn hint(pr: i64, o: &Observation, lanes: &[String]) -> Option<Value> {
    let reasons = repair_reasons(o);
    if reasons.is_empty() {
        return None;
    }
    let to = lanes
        .iter()
        .find(|l| Some(l.as_str()) != o.author.as_deref())
        .cloned();
    let body = to.as_ref().map(|lane| {
        format!(
            "assignee-hint: {lane}\n\nPR #{pr} needs a repair ({}). The author lane is not \
             assigned the repair. A hint grants nothing: ownership still comes only from a \
             claim comment.",
            reasons.join("; ")
        )
    });
    Some(json!({
        "pr": pr,
        "issue": o.issue,
        "author": o.author,
        "reasons": reasons,
        "hint_to": to,
        "comment_body": body,
        "repair_issue_title": o.issue.map(|i| format!("[repair] PR #{pr} (Refs #{i})")),
        "note": if to.is_none() { Some("no free lane other than the author") } else { None },
    }))
}

fn state_path(project: &Path) -> PathBuf {
    project.join(crate::PROJECT_STATE_DIR).join(STATE)
}

/// Loads the previous tick; an absent or foreign file is a first tick.
pub fn load_state(project: &Path) -> TickState {
    fs::read(state_path(project))
        .ok()
        .and_then(|b| serde_json::from_slice::<TickState>(&b).ok())
        .filter(|s| s.schema_version == 1)
        .unwrap_or_default()
}

fn save_state(project: &Path, state: &TickState) -> Result<()> {
    let path = state_path(project);
    let dir = path
        .parent()
        .ok_or("leader tick state directory unavailable")?;
    fs::create_dir_all(dir).map_err(|_| "leader tick state directory unavailable")?;
    let tmp = path.with_extension(format!("json.tmp-{}", std::process::id()));
    let bytes = serde_json::to_vec_pretty(state).map_err(|_| "encode failed")?;
    let mut f = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| "leader tick state write failed")?;
    f.write_all(&bytes)
        .and_then(|_| f.sync_all())
        .map_err(|_| "leader tick state write failed")?;
    fs::rename(&tmp, &path).map_err(|_| {
        let _ = fs::remove_file(&tmp);
        "leader tick state replace failed".into()
    })
}

/// The interval and where it came from: the profile key, or the default when no profile
/// or no key exists.
pub fn interval(project: &Path) -> Result<(u64, &'static str)> {
    match crate::read_profile(project) {
        Ok((p, _)) => Ok(match p.leader {
            Some(l) => (l.pr_check_interval_secs, "profile"),
            None => (DEFAULT_INTERVAL_SECS, "default"),
        }),
        Err(e) if e == "project profile missing" => Ok((DEFAULT_INTERVAL_SECS, "default")),
        Err(e) => Err(e),
    }
}

/// One tick: observe, compare with `prev`, and return the report and the new state.
pub fn tick<T: Transport, C: Clock>(
    req: &TickRequest,
    prev: &TickState,
    client: &Client<'_, T, C>,
    clock: &C,
) -> Result<(Value, TickState)> {
    if req.schema_version != 1 {
        return Err("unsupported request schema_version".into());
    }
    if req.base_branch.is_empty() || req.base_branch.contains(['&', '?', '#', ' ']) {
        return Err("leader tick base_branch must be a plain branch name".into());
    }
    let now = clock.now();
    let list = pages(
        client,
        format!(
            "{API}/repos/{}/{}/pulls?state=open&base={}&per_page=100",
            req.owner, req.repo, req.base_branch
        ),
    )?;
    let mut next = TickState {
        schema_version: 1,
        prs: BTreeMap::new(),
    };
    for pr in &list {
        if let Some((n, mut o)) = observe(client, req, pr, now)? {
            // GitHub answers `unknown` while it is still computing; that is not a change.
            if o.mergeable_state == "unknown" {
                if let Some(p) = prev.prs.get(&n).filter(|p| p.head == o.head) {
                    o.mergeable_state = p.mergeable_state.clone();
                }
            }
            next.prs.insert(n, o);
        }
    }
    let mut changes = Vec::new();
    let mut hints = Vec::new();
    for (n, o) in &next.prs {
        if prev.prs.get(n) != Some(o) {
            changes.push(json!({"pr": n, "now": o, "before": prev.prs.get(n)}));
            hints.extend(hint(*n, o, &req.lanes));
        }
    }
    for n in prev.prs.keys().filter(|n| !next.prs.contains_key(n)) {
        changes.push(json!({"pr": n, "closed": true}));
    }
    // Issue #2678: the count and "no change" are said outright, not left to an empty list.
    let summary = if changes.is_empty() {
        format!(
            "{} open pull request(s) read; no change since the previous tick",
            next.prs.len()
        )
    } else {
        format!(
            "{} open pull request(s) read; {} changed since the previous tick",
            next.prs.len(),
            changes.len()
        )
    };
    let report = json!({
        "summary": summary,
        "observed": next.prs.len(),
        "changes": changes,
        "hints": hints,
        "writes": "none on GitHub; post hints through effect-intend, gh, effect-confirm",
    });
    Ok((report, next))
}

/// CLI: interval from the profile, previous state, one tick, state replaced on success.
pub fn tick_cli<T: Transport, C: Clock>(
    project: &Path,
    req: TickRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    let (secs, source) = interval(project)?;
    let prev = load_state(project);
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let (mut report, state) = tick(&req, &prev, &client, clock)?;
    save_state(project, &state)?;
    report["interval_secs"] = json!(secs);
    report["interval_source"] = json!(source);
    report["interval_note"] = json!("user-chosen value (D-2), not measured; the tick never sleeps");
    Ok(report)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn obs(verdict: &str, ci: &str, state: &str, author: Option<&str>) -> Observation {
        Observation {
            head: "h".into(),
            issue: Some(7),
            mergeable_state: state.into(),
            audit: "none".into(),
            ci: ci.into(),
            verdict: verdict.into(),
            author: author.map(str::to_string),
        }
    }

    #[test]
    fn verdict_and_refs_are_read_for_the_current_head() {
        let c = vec![
            json!({"body": "AUDIT #7 (A1) head aaa base b — verdict: FAIL"}),
            json!({"body": "AUDIT #7 (A1) head bbb base b — verdict: PASS"}),
        ];
        assert_eq!(verdict_for(&c, "aaa"), "FAIL");
        assert_eq!(verdict_for(&c, "bbb"), "PASS");
        assert_eq!(verdict_for(&c, "ccc"), "none");
        assert_eq!(refs_issue("x\nRefs #2162\n"), Some(2162));
        assert_eq!(refs_issue("no ref"), None);
    }

    #[test]
    fn hint_skips_the_author_and_healthy_prs() {
        let lanes = vec!["agent-1".to_string(), "agent-2".to_string()];
        let h = hint(5, &obs("FAIL", "success", "clean", Some("agent-1")), &lanes).unwrap();
        assert_eq!(h["hint_to"], json!("agent-2"));
        assert!(h["comment_body"]
            .as_str()
            .unwrap()
            .starts_with("assignee-hint: agent-2"));
        let only_author = vec!["agent-1".to_string()];
        let h = hint(
            5,
            &obs("none", "failure", "clean", Some("agent-1")),
            &only_author,
        )
        .unwrap();
        assert_eq!(h["hint_to"], Value::Null);
        assert!(hint(5, &obs("PASS", "success", "clean", Some("agent-1")), &lanes).is_none());
    }
}
