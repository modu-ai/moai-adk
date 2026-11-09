//! Claim ownership re-verification (rc.2 A9 #2255): the write boundary a lane crosses before
//! every commit and push.
//!
//! The claim contract is unchanged (AGENTS.md, 06-A): one `moai-claim:` comment per claimant,
//! the lowest comment id among open claims holds the Issue, a lease is 4 h, and an expired
//! lease never hands the Issue to someone else — a handoff comment or the operator does. The
//! comment is not an atomic lock; this module only re-reads every comment and decides, so a
//! competing claim that became visible late is judged on the re-read.
//!
//! What it adds:
//! - [`classify_comment`]: one parser for the claim line (the RC1 form plus the appended
//!   fields `render_claim_line` writes) and for two machine lines that close a claim,
//!   `moai-claim-release: comment=<id>` and `moai-claim-handoff: comment=<id> to=<session>`.
//! - [`evaluate`]: a pure verdict over all comments — `owner`, `not_owner`, `stale`, `closed`,
//!   `no_claim` or `uncertain` — with every ignored comment reported (`malformed`,
//!   `duplicate`, `author_mismatch`; REQ-A9-02).
//! - `claim-verify`: GET-only; succeeds only for `owner`, otherwise refuses with a classified
//!   message (REQ-A9-01). `--hook` reads a host `PreToolUse` input and refuses only a Bash
//!   `git commit` / `git push` run inside a worktree that has an active claim record.
//!
//! Limits: every lane posts through one shared GitHub account, so "author mismatch" means an
//! author outside `allowed_authors`; a forged session value from an allowed author cannot be
//! told apart from a real one.

use crate::claim::{parse_claim_line, ClaimRecord};
use crate::forge::policy::{next_link, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use crate::Result;
use serde::Deserialize;
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

const MAX_PAGES: u32 = 50;
const RELEASE_PREFIX: &str = "moai-claim-release:";
const HANDOFF_PREFIX: &str = "moai-claim-handoff:";

/// One Issue comment as the evaluator sees it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct IssueComment {
    pub id: i64,
    pub author: String,
    pub body: String,
}

/// What one comment says about claims.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum CommentKind {
    Claim {
        session: String,
        epoch: u64,
        lease_until: String,
    },
    Release {
        target: i64,
    },
    Handoff {
        target: i64,
        to: String,
    },
    /// A line that starts like a claim line but cannot be read; the reason says why.
    Malformed(&'static str),
    /// No claim line at all.
    Other,
}

fn token<'a>(rest: &'a str, key: &str) -> Option<&'a str> {
    rest.split_whitespace()
        .find_map(|t| t.strip_prefix(key).and_then(|v| v.strip_prefix('=')))
}

/// Classifies a comment by its first claim-family line.
pub fn classify_comment(body: &str) -> CommentKind {
    for line in body.lines().map(str::trim) {
        if line.starts_with("moai-claim:") {
            return match parse_claim_line(line) {
                Some(c) if !c.session.is_empty() && parse_utc(&c.lease_until).is_some() => {
                    CommentKind::Claim {
                        session: c.session,
                        epoch: c.epoch,
                        lease_until: c.lease_until,
                    }
                }
                Some(_) => CommentKind::Malformed("claim lease_until is not YYYY-MM-DDTHH:MM:SSZ"),
                None => CommentKind::Malformed("claim line lacks session, epoch or lease_until"),
            };
        }
        if let Some(rest) = line.strip_prefix(RELEASE_PREFIX) {
            return match token(rest, "comment").and_then(|v| v.parse().ok()) {
                Some(target) => CommentKind::Release { target },
                None => CommentKind::Malformed("release line lacks comment=<id>"),
            };
        }
        if let Some(rest) = line.strip_prefix(HANDOFF_PREFIX) {
            let target = token(rest, "comment").and_then(|v| v.parse().ok());
            let to = token(rest, "to").filter(|v| !v.is_empty());
            return match (target, to) {
                (Some(target), Some(to)) => CommentKind::Handoff {
                    target,
                    to: to.to_string(),
                },
                _ => CommentKind::Malformed("handoff line lacks comment=<id> or to=<session>"),
            };
        }
    }
    CommentKind::Other
}

/// The release line a leader or operator posts to close a claim (no POST here).
pub fn render_release_line(claim_comment_id: i64, by: &str) -> String {
    format!("{RELEASE_PREFIX} comment={claim_comment_id} by={by}")
}

/// The handoff line that closes a claim and names the session expected to claim next.
pub fn render_handoff_line(claim_comment_id: i64, to: &str, by: &str) -> String {
    format!("{HANDOFF_PREFIX} comment={claim_comment_id} to={to} by={by}")
}

/// Who is verifying: the session, and the claim comment id when the journal knows it.
#[derive(Clone, Debug, Default)]
pub struct Verifier<'a> {
    pub session: &'a str,
    pub comment_id: Option<i64>,
    /// Logins allowed to post claim lines; empty means not checked.
    pub allowed_authors: &'a [String],
    pub now: i64,
}

#[derive(Clone, Debug)]
struct ValidClaim {
    id: i64,
    session: String,
    epoch: u64,
    live: bool,
}

/// The verdict over all comments of one Issue. `status` is one of `owner`, `not_owner`,
/// `stale`, `closed`, `no_claim`, `uncertain`; only `owner` may write.
pub fn evaluate(comments: &[IssueComment], who: &Verifier<'_>) -> Value {
    let mut sorted: Vec<&IssueComment> = comments.iter().collect();
    sorted.sort_by_key(|c| c.id);
    let mut claims: Vec<ValidClaim> = Vec::new();
    let mut invalid = Vec::new();
    let mut closed = Vec::new();
    let mut seen_claim_lines = 0;
    for c in sorted {
        let kind = classify_comment(&c.body);
        if kind == CommentKind::Other {
            continue;
        }
        seen_claim_lines += 1;
        if !who.allowed_authors.is_empty() && !who.allowed_authors.contains(&c.author) {
            invalid.push(json!({"comment_id": c.id, "reason": "author_mismatch"}));
            continue;
        }
        match kind {
            CommentKind::Claim {
                session,
                epoch,
                lease_until,
            } => {
                // The same session and epoch claimed again: the first stands.
                if claims
                    .iter()
                    .any(|v| v.session == session && v.epoch == epoch)
                {
                    invalid.push(json!({"comment_id": c.id, "reason": "duplicate"}));
                    continue;
                }
                let live = parse_utc(&lease_until).is_some_and(|t| t > who.now);
                claims.push(ValidClaim {
                    id: c.id,
                    session,
                    epoch,
                    live,
                });
            }
            CommentKind::Release { target } | CommentKind::Handoff { target, .. } => {
                // Only an earlier valid claim can be closed.
                if !claims.iter().any(|v| v.id == target) {
                    invalid.push(json!({"comment_id": c.id, "reason": "malformed",
                                        "detail": "closes no earlier claim comment"}));
                    continue;
                }
                let (by, to) = match &kind {
                    CommentKind::Handoff { to, .. } => ("handoff", Some(to.clone())),
                    _ => ("release", None),
                };
                closed.push(
                    json!({"comment_id": target, "by": by, "to": to, "closing_comment_id": c.id}),
                );
            }
            CommentKind::Malformed(detail) => {
                invalid.push(json!({"comment_id": c.id, "reason": "malformed", "detail": detail}));
            }
            CommentKind::Other => unreachable!(),
        }
    }
    let is_closed = |id: i64| closed.iter().any(|c| c["comment_id"] == json!(id));
    let mine = match who.comment_id {
        Some(id) => claims.iter().find(|v| v.id == id),
        None => claims.iter().find(|v| v.session == who.session),
    };
    let holder = claims.iter().find(|v| !is_closed(v.id));
    let (status, lost_to) = if let (Some(id), None) = (who.comment_id, mine) {
        let visible = comments.iter().any(|c| c.id == id);
        (if visible { "no_claim" } else { "uncertain" }, None)
    } else if mine.is_some_and(|m| is_closed(m.id)) {
        ("closed", None)
    } else {
        match holder {
            None => ("no_claim", None),
            // An expired holder stays the holder: no takeover by time alone (REQ-A9-03).
            Some(h) if !h.live => ("stale", None),
            Some(h) if mine.is_some_and(|m| m.id == h.id) => ("owner", None),
            Some(h) => ("not_owner", Some(h.id)),
        }
    };
    json!({
        "status": status,
        "holder": holder.map(|h| h.id),
        "holder_live": holder.map(|h| h.live),
        "mine": mine.map(|m| m.id),
        "lost_to": lost_to,
        "invalid": invalid,
        "closed": closed,
        "claim_lines": seen_claim_lines,
        "auto_assigned": false,
    })
}

/// The refusal message for a non-owner verdict; `None` for `owner`.
pub fn refusal(verdict: &Value) -> Option<String> {
    let id = |k: &str| {
        verdict[k]
            .as_i64()
            .map_or("none".to_string(), |v| v.to_string())
    };
    Some(match verdict["status"].as_str()? {
        "owner" => return None,
        "not_owner" => format!(
            "claim verify: not owner; the lowest open claim is comment {} (REQ-A9-01), so no commit or push",
            id("lost_to")
        ),
        "stale" => format!(
            "claim verify: stale; claim comment {} lease expired without a handoff, and the operator decides (REQ-A9-03)",
            id("holder")
        ),
        "closed" => format!(
            "claim verify: closed; claim comment {} was released or handed off",
            id("mine")
        ),
        "uncertain" => "claim verify: uncertain; this session's claim comment is not visible on the issue".into(),
        _ => "claim verify: no claim of this session on the issue".into(),
    })
}

/// `claim-verify` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct VerifyRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub issue_number: i64,
    pub session: String,
    #[serde(default)]
    pub comment_id: Option<i64>,
    #[serde(default)]
    pub allowed_authors: Vec<String>,
}

/// Every comment of the Issue (GET only, `Link` pagination).
pub fn read_comments<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    issue: i64,
) -> Result<Vec<IssueComment>> {
    let mut url = Some(format!(
        "https://api.github.com/repos/{owner}/{repo}/issues/{issue}/comments?per_page=100"
    ));
    let mut out = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err("claim comment page limit exceeded (50)".into());
        }
        let resp = client.get(&u, None).map_err(String::from)?;
        for c in resp
            .body
            .as_ref()
            .and_then(Value::as_array)
            .into_iter()
            .flatten()
        {
            if let (Some(id), Some(body)) = (c["id"].as_i64(), c["body"].as_str()) {
                out.push(IssueComment {
                    id,
                    author: c["user"]["login"].as_str().unwrap_or("").to_string(),
                    body: body.to_string(),
                });
            }
        }
        url = resp.header("link").and_then(|l| next_link(Some(l)));
    }
    Ok(out)
}

fn check_request(req: &VerifyRequest) -> Result<()> {
    if req.schema_version != 1 {
        return Err("unsupported request schema_version; supported: 1".into());
    }
    let name = |s: &str| {
        !s.is_empty()
            && s.len() <= 100
            && s.chars()
                .all(|c| c.is_ascii_alphanumeric() || "-_.".contains(c))
    };
    if !name(&req.owner) || !name(&req.repo) || req.issue_number <= 0 || req.session.is_empty() {
        return Err(
            "invalid claim verify request: owner, repo, issue_number and session are required"
                .into(),
        );
    }
    Ok(())
}

/// `claim-verify` with a typed request: the verdict when this session owns the Issue,
/// otherwise the refusal as an error (exit 1, classified).
pub fn verify_cli<T: Transport, C: Clock>(
    req: &VerifyRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    check_request(req)?;
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let comments = read_comments(&client, &req.owner, &req.repo, req.issue_number)?;
    let verdict = evaluate(
        &comments,
        &Verifier {
            session: &req.session,
            comment_id: req.comment_id,
            allowed_authors: &req.allowed_authors,
            now: clock.now(),
        },
    );
    match refusal(&verdict) {
        None => Ok(verdict),
        Some(message) => Err(message),
    }
}

/// True when a shell command line runs `git commit` or `git push` in any of its segments.
pub fn is_commit_or_push(command: &str) -> bool {
    command
        .split([';', '&', '|', '\n', '(', ')'])
        .any(|segment| {
            let mut words = segment.split_whitespace();
            let found_git = words.any(|w| w == "git" || w.ends_with("/git"));
            found_git && words.any(|w| w == "commit" || w == "push")
        })
}

/// The active claim record whose worktree contains `cwd`.
fn record_for(records: Vec<ClaimRecord>, cwd: &Path) -> Option<ClaimRecord> {
    records
        .into_iter()
        .find(|r| std::fs::canonicalize(&r.worktree.path).is_ok_and(|p| cwd.starts_with(p)))
}

/// `claim-verify --hook`: the host `PreToolUse` guard. Allows everything except a Bash
/// `git commit` / `git push` inside a worktree with an active claim record, which it
/// verifies; a refusal is an error the CLI turns into exit 2 so the host blocks the call.
pub fn hook_cli<T: Transport, C: Clock>(
    project: &Path,
    stdin: &[u8],
    transport: &T,
    clock: &C,
) -> Result<Value> {
    let allow = |reason: &str| Ok(json!({"allowed": true, "reason": reason}));
    let input: Value = serde_json::from_slice(stdin).unwrap_or(Value::Null);
    let command = input["tool_input"]["command"].as_str().unwrap_or("");
    if input["tool_name"] != json!("Bash") || !is_commit_or_push(command) {
        return allow("not a git commit or push");
    }
    let cwd = input["cwd"]
        .as_str()
        .map(PathBuf::from)
        .or_else(|| std::env::current_dir().ok())
        .and_then(|p| std::fs::canonicalize(p).ok());
    let Some(cwd) = cwd else {
        return allow("working directory unavailable; no claimed worktree to guard");
    };
    let records = crate::worktree::active_records(project).unwrap_or_default();
    let Some(record) = record_for(records, &cwd) else {
        return allow("not a claimed worktree");
    };
    let req = VerifyRequest {
        schema_version: 1,
        owner: record.task_ref.owner.clone(),
        repo: record.task_ref.repo.clone(),
        issue_number: record.task_ref.issue_number,
        session: record.owner.id.clone(),
        comment_id: record.comment_id,
        allowed_authors: Vec::new(),
    };
    let mut verdict = verify_cli(&req, transport, clock)?;
    if is_push(command) {
        if let Some(token) = &record.claim_token {
            verdict["trailer_checked"] =
                json!(missing_trailers(&cwd, &record.base_revision, token)?);
        }
    }
    verdict["allowed"] = json!(true);
    verdict["issue"] = json!(record.task_ref.issue_number);
    Ok(verdict)
}

/// True when a command line runs `git push` in any segment.
fn is_push(command: &str) -> bool {
    command
        .split([';', '&', '|', '\n', '(', ')'])
        .any(|segment| {
            let mut words = segment.split_whitespace();
            words.any(|w| w == "git" || w.ends_with("/git")) && words.any(|w| w == "push")
        })
}

/// Before a push, every commit on the branch's own first-parent line since the claim's base —
/// merge commits included, the merged-in commits of other cards excluded — must carry the
/// claim token (the #2226 failure: merge commits without the trailer). Returns the number of
/// commits checked; refuses with the short SHAs that lack it.
fn missing_trailers(cwd: &Path, base: &str, token: &str) -> Result<usize> {
    if base.len() != 40 || !base.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err(
            "claim verify: trailer check unavailable; the claim base is not a commit".into(),
        );
    }
    let out = std::process::Command::new("git")
        .arg("-C")
        .arg(cwd)
        .args(["log", "--first-parent", "--format=%H%x00%B%x1e"])
        .arg(format!("{base}..HEAD"))
        .output()
        .map_err(|_| "claim verify: trailer check unavailable; git could not run")?;
    if !out.status.success() {
        return Err(
            "claim verify: trailer check unavailable; git log of the claim base failed".into(),
        );
    }
    let text = String::from_utf8_lossy(&out.stdout);
    let mut checked = 0;
    let mut lacking = Vec::new();
    for entry in text
        .split('\u{1e}')
        .map(str::trim)
        .filter(|e| !e.is_empty())
    {
        let (sha, body) = entry.split_once('\0').unwrap_or((entry, ""));
        checked += 1;
        if !body.lines().any(|l| l.trim() == token) {
            lacking.push(sha.get(..9).unwrap_or(sha).to_string());
        }
    }
    if lacking.is_empty() {
        Ok(checked)
    } else {
        Err(format!(
            "claim verify: trailer missing; commits {} lack `{token}`, so no push",
            lacking.join(", ")
        ))
    }
}
