//! Lane status table: `lane-status` (rc.2 A11 #2265; research record
//! `docs/design-review/lane-status-a11-2026-11-10.md`).
//!
//! One read-only call builds a row per (session, Issue) for the Issues the caller names: the
//! claim holder and its lease (`claim_verify::evaluate`, A9), the open PR that `Refs` the
//! Issue with its head and branch, the local lane worktree (`git worktree list`), and a phase
//! with the record it was read from. It classifies why each caller-named lane is idle, from
//! records only (`unknown` when none supports a cause), and judges duplicate, out-of-order and
//! unreported heads by the PR's own commit order.
//!
//! GitHub is read with GET only and nothing is written anywhere, so a repeated call over the
//! same remote gives the same table. Pane states and pending user decisions are not observable
//! here: the caller passes them with a source and a time, and they are cited as given.
//! Not done here: host session events (#A12), posting the assignment drafts, live runs.
use crate::claim_verify::{classify_comment, evaluate, CommentKind, IssueComment, Verifier};
use crate::forge::policy::{next_link, parse_utc, Clock};
use crate::forge::transport::{HttpRequest, HttpResponse, Transport};
use crate::forge::Client;
use crate::message::Message;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

const API: &str = "https://api.github.com";
/// Most Issues one call inspects (a conservative default, not a measured value).
pub const MAX_ISSUES: usize = 50;
/// Most lanes one call classifies (a conservative default, not a measured value).
pub const MAX_LANES: usize = 50;
/// Most pages one list read follows, as `leader-tick`.
const MAX_PAGES: u32 = 50;

/// `lane-status` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct LaneStatusRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    pub base_branch: String,
    /// The Issues to inspect (the leader's active cards).
    pub issues: Vec<i64>,
    /// The lanes to classify, with the lane name the leader uses for each session.
    #[serde(default)]
    pub lanes: Vec<LaneInput>,
    /// User decisions the leader is waiting on, from its own question record.
    #[serde(default)]
    pub pending_decisions: Vec<PendingDecision>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct LaneInput {
    pub session: String,
    #[serde(default)]
    pub name: Option<String>,
    #[serde(default)]
    pub pane: Option<PaneObservation>,
}

/// What the caller saw on the lane's pane; never inferred here.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct PaneObservation {
    pub state: PaneState,
    pub source: String,
    pub observed_at: String,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum PaneState {
    Working,
    PermissionPrompt,
    InputPending,
    Unreachable,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct PendingDecision {
    pub issue: i64,
    pub source: String,
    pub recorded_at: String,
}

/// Where a card stands, read from a record (`phase_evidence` names it).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Phase {
    Claimed,
    Implementing,
    PrOpen,
    Handoff,
    Waiting,
}

impl Phase {
    pub const ALL: [Phase; 5] = [
        Phase::Claimed,
        Phase::Implementing,
        Phase::PrOpen,
        Phase::Handoff,
        Phase::Waiting,
    ];
    pub fn as_str(self) -> &'static str {
        match self {
            Phase::Claimed => "claimed",
            Phase::Implementing => "implementing",
            Phase::PrOpen => "pr_open",
            Phase::Handoff => "handoff",
            Phase::Waiting => "waiting",
        }
    }
    /// The lane is expected to be working on the card itself.
    fn active(self) -> bool {
        matches!(self, Phase::Claimed | Phase::Implementing | Phase::Handoff)
    }
}

/// Why a lane is idle; `Unknown` when no record supports a cause.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum IdleCause {
    Unreachable,
    PermissionPrompt,
    InputPending,
    UserDecision,
    WaitingPrerequisite,
    WaitingAudit,
    WaitingIntegration,
    NoReadyCard,
    ReadyCardUnassigned,
    Unknown,
}

impl IdleCause {
    pub const ALL: [IdleCause; 10] = [
        IdleCause::Unreachable,
        IdleCause::PermissionPrompt,
        IdleCause::InputPending,
        IdleCause::UserDecision,
        IdleCause::WaitingPrerequisite,
        IdleCause::WaitingAudit,
        IdleCause::WaitingIntegration,
        IdleCause::NoReadyCard,
        IdleCause::ReadyCardUnassigned,
        IdleCause::Unknown,
    ];
    pub fn as_str(self) -> &'static str {
        match self {
            IdleCause::Unreachable => "unreachable",
            IdleCause::PermissionPrompt => "permission_prompt",
            IdleCause::InputPending => "input_pending",
            IdleCause::UserDecision => "user_decision",
            IdleCause::WaitingPrerequisite => "waiting_prerequisite",
            IdleCause::WaitingAudit => "waiting_audit",
            IdleCause::WaitingIntegration => "waiting_integration",
            IdleCause::NoReadyCard => "no_ready_card",
            IdleCause::ReadyCardUnassigned => "ready_card_unassigned",
            IdleCause::Unknown => "unknown",
        }
    }
}

/// One Issue comment as read from the API.
#[derive(Clone, Debug)]
pub struct Comment {
    pub id: i64,
    pub created_at: String,
    pub body: String,
}

#[derive(Clone, Debug)]
pub struct IssueData {
    pub number: i64,
    pub open: bool,
    pub body: String,
    pub comments: Vec<Comment>,
}

#[derive(Clone, Debug)]
pub struct PrData {
    pub number: i64,
    pub head: String,
    pub branch: String,
    pub issue: Option<i64>,
    /// The PR's commits in the API's order (oldest first, at most 250).
    pub commits: Vec<String>,
    /// Claim comment ids from `Moai-Claim: <issue>#<id>` lines in the PR's commit messages,
    /// for this PR's own Issue only, in commit order without repeats (#2403 item 3).
    pub claim_ids: Vec<i64>,
    /// The PR's own conversation comments (audit reservations and verdicts; #2403 item 3).
    pub comments: Vec<Comment>,
}

#[derive(Clone, Debug)]
pub struct WorktreeData {
    pub path: String,
    pub head: String,
    pub branch: Option<String>,
}

/// Everything one call read; [`build`] turns it into the table without further I/O.
#[derive(Clone, Debug)]
pub struct Snapshot {
    pub issues: Vec<IssueData>,
    /// Open state of each `Blocked-By:` Issue that was not itself inspected.
    pub other_open: BTreeMap<i64, bool>,
    pub prs: Vec<PrData>,
    pub worktrees: std::result::Result<Vec<WorktreeData>, String>,
    pub now: i64,
}

fn validate(req: &LaneStatusRequest) -> Result<()> {
    if req.schema_version != 1 {
        return Err("unsupported request schema_version".into());
    }
    if req.base_branch.is_empty() || req.base_branch.contains(['&', '?', '#', ' ']) {
        return Err("lane status base_branch must be a plain branch name".into());
    }
    if req.issues.is_empty() || req.issues.len() > MAX_ISSUES || req.issues.iter().any(|n| *n <= 0)
    {
        return Err("lane status issues must be 1 to 50 Issue numbers".into());
    }
    if req.lanes.len() > MAX_LANES {
        return Err("lane status lanes must be at most 50".into());
    }
    let blank = |s: &str| s.trim().is_empty();
    let pane_blank = req
        .lanes
        .iter()
        .filter_map(|l| l.pane.as_ref())
        .any(|p| blank(&p.source) || blank(&p.observed_at));
    let decision_blank = req
        .pending_decisions
        .iter()
        .any(|d| blank(&d.source) || blank(&d.recorded_at));
    if pane_blank || decision_blank || req.lanes.iter().any(|l| blank(&l.session)) {
        return Err(
            "lane status evidence must be given with a session, a source and a time".into(),
        );
    }
    Ok(())
}

pub(crate) fn pages<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    first: String,
) -> Result<Vec<Value>> {
    // Shared readers follow GitHub's Link verbatim, including numeric repository URLs.
    pages_with_evidence(client, first, None).map(|(values, _)| values)
}

fn pages_with_evidence<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    first: String,
    mut source_urls: Option<&mut Vec<String>>,
) -> Result<(Vec<Value>, bool)> {
    let mut complete = true;
    let (mut url, mut out, mut n) = (Some(first), Vec::new(), 0);
    while let Some(u) = url {
        n += 1;
        if n > MAX_PAGES {
            return Err("lane status page limit exceeded (50)".into());
        }
        let resp = client.get(&u, None).map_err(String::from)?;
        if let Some(urls) = &mut source_urls {
            urls.push(u);
        }
        let link = resp.header("link");
        url = next_link(link);
        if link.is_some_and(|link| link.contains("rel=\"next\"")) && url.is_none() {
            complete = false;
        }
        match resp.body {
            Some(Value::Array(values)) => out.extend(values),
            _ => complete = false,
        }
    }
    Ok((out, complete))
}

pub(crate) fn one<T: Transport, C: Clock>(client: &Client<'_, T, C>, url: &str) -> Result<Value> {
    let resp = client.get(url, None).map_err(String::from)?;
    Ok(resp.body.unwrap_or(Value::Null))
}

/// Issue or PR comments from the API, sorted by id without repeats.
pub(crate) fn read_comments(raw: Vec<Value>) -> Vec<Comment> {
    comments_from(&raw)
}

fn comments_from(raw: &[Value]) -> Vec<Comment> {
    let mut comments: Vec<Comment> = raw
        .iter()
        .filter_map(|c| {
            Some(Comment {
                id: c["id"].as_i64()?,
                created_at: c["created_at"].as_str().unwrap_or_default().into(),
                body: c["body"].as_str().unwrap_or_default().into(),
            })
        })
        .collect();
    comments.sort_by_key(|c| c.id);
    comments.dedup_by_key(|c| c.id);
    comments
}

/// Claim comment ids from whole-line `Moai-Claim: <issue>#<id>` trailers naming `issue`.
pub(crate) fn claim_ids_in(message: &str, issue: i64) -> Vec<i64> {
    message
        .lines()
        .filter_map(|l| l.trim().strip_prefix("Moai-Claim:"))
        .filter_map(|t| t.trim().split_once('#'))
        .filter(|(n, _)| n.parse::<i64>().ok() == Some(issue))
        .filter_map(|(_, id)| id.parse::<i64>().ok())
        .collect()
}

/// A whole-line `moai-audit-claim: lane=<lane> head=<40 lowercase hex>` reservation.
pub(crate) fn reservation_in(body: &str) -> Option<(String, String)> {
    body.lines().find_map(|l| {
        let rest = l.trim().strip_prefix("moai-audit-claim:")?;
        let mut words = rest.split_whitespace();
        let lane = words.next()?.strip_prefix("lane=")?;
        let head = words.next()?.strip_prefix("head=")?;
        let hex40 =
            head.len() == 40 && head.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'));
        (words.next().is_none() && !lane.is_empty() && hex40)
            .then(|| (lane.to_string(), head.to_string()))
    })
}

/// The audit columns of a PR row (#2403 item 3): reservations with the current-head winner,
/// the verdict on the current head (PR comments first, then the Issue's), and the sessions
/// whose claim tokens appear in the PR's commits — the authors, who may not audit it.
pub(crate) fn audit_value(issue: &IssueData, pr: &PrData) -> Value {
    let reservations: Vec<Value> = pr
        .comments
        .iter()
        .filter_map(|c| {
            let (lane, head) = reservation_in(&c.body)?;
            let current = head == pr.head;
            let state = if current { "current" } else { "stale" };
            Some(json!({"lane": lane, "comment_id": c.id, "head": head, "current": current, "state": state}))
        })
        .collect();
    // The same holder rule as `audit-claim` (#2499): lowest current-head reservation, moved by
    // the last leader handoff line that names it.
    let holder = crate::audit_claim::holder_of(&pr.comments, &pr.head);
    let winner = holder
        .as_ref()
        .map(|h| json!({"lane": h.lane, "comment_id": h.comment_id}));
    let verdict = verdict_on(&pr.comments, &pr.head)
        .map(|(v, id)| json!({"verdict": v, "comment_id": id, "source": "pr"}))
        .or_else(|| {
            verdict_on(&issue.comments, &pr.head)
                .map(|(v, id)| json!({"verdict": v, "comment_id": id, "source": "issue"}))
        });
    let mut authors = Vec::new();
    let mut unresolved = Vec::new();
    for id in &pr.claim_ids {
        let session = issue.comments.iter().find(|c| c.id == *id).and_then(|c| {
            match classify_comment(&c.body) {
                CommentKind::Claim { session, .. } => Some(session),
                _ => None,
            }
        });
        match session {
            Some(s) => authors.push(json!({"session": s, "claim_comment_id": id})),
            None => unresolved.push(*id),
        }
    }
    let winner_is_author = holder.as_ref().map(|h| {
        authors.iter().any(|a| {
            crate::audit_claim::same_lane(a["session"].as_str().unwrap_or_default(), &h.lane)
        })
    });
    json!({
        "head": pr.head,
        "reservations": reservations,
        "winner_on_head": winner,
        "winner_via": holder.as_ref().map(|h| if h.via_handoff { "handoff" } else { "reservation" }),
        "winner_reservation_id": holder.as_ref().map(|h| h.reservation_id),
        "winner_is_author": winner_is_author,
        "lane_name_rule": crate::audit_claim::NAME_RULE,
        "verdict_on_head": verdict,
        "author_sessions": authors,
        "unresolved_claim_ids": unresolved,
    })
}

/// Reads everything for `req` (GET only) and the local worktree list of `project`.
pub fn collect<T: Transport, C: Clock>(
    project: &Path,
    req: &LaneStatusRequest,
    client: &Client<'_, T, C>,
    clock: &C,
) -> Result<Snapshot> {
    collect_with_raw(project, req, client, clock).map(|(snapshot, _)| snapshot)
}

struct ReadIssue {
    number: i64,
    issue: Value,
    comments: Value,
    complete: bool,
    source_urls: Vec<String>,
}

fn collect_with_raw<T: Transport, C: Clock>(
    project: &Path,
    req: &LaneStatusRequest,
    client: &Client<'_, T, C>,
    clock: &C,
) -> Result<(Snapshot, Vec<ReadIssue>)> {
    validate(req)?;
    let repo = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let wanted: BTreeSet<i64> = req.issues.iter().copied().collect();
    let pr_list = pages(
        client,
        format!(
            "{repo}/pulls?state=open&base={}&per_page=100",
            req.base_branch
        ),
    )?;
    let mut issues = Vec::new();
    let mut raw = Vec::new();
    for n in &wanted {
        let issue = one(client, &format!("{repo}/issues/{n}"))?;
        let mut source_urls = Vec::new();
        let (raw_comments, complete) = pages_with_evidence(
            client,
            format!("{repo}/issues/{n}/comments?per_page=100"),
            Some(&mut source_urls),
        )?;
        let comments = comments_from(&raw_comments);
        issues.push(IssueData {
            number: *n,
            open: issue["state"].as_str() != Some("closed"),
            body: issue["body"].as_str().unwrap_or_default().into(),
            comments,
        });
        raw.push(ReadIssue {
            number: *n,
            issue,
            comments: Value::Array(raw_comments),
            complete,
            source_urls,
        });
    }
    let others: BTreeSet<i64> = issues
        .iter()
        .flat_map(|i| crate::queue::refs_on(&i.body, "Blocked-By:"))
        .filter(|b| !wanted.contains(b))
        .collect();
    let mut other_open = BTreeMap::new();
    for b in others {
        let issue = one(client, &format!("{repo}/issues/{b}"))?;
        other_open.insert(b, issue["state"].as_str() != Some("closed"));
    }
    let mut prs: Vec<PrData> = pr_list
        .iter()
        .filter_map(|p| {
            Some(PrData {
                number: p["number"].as_i64()?,
                head: p["head"]["sha"].as_str()?.into(),
                branch: p["head"]["ref"].as_str().unwrap_or_default().into(),
                issue: crate::leader::refs_issue(p["body"].as_str().unwrap_or_default()),
                commits: Vec::new(),
                claim_ids: Vec::new(),
                comments: Vec::new(),
            })
        })
        .filter(|p| p.issue.is_some_and(|i| wanted.contains(&i)))
        .collect();
    prs.sort_by_key(|p| p.number);
    prs.dedup_by_key(|p| p.number);
    for pr in prs.iter_mut() {
        let commits = pages(
            client,
            format!("{repo}/pulls/{}/commits?per_page=100", pr.number),
        )?;
        pr.commits = commits
            .iter()
            .filter_map(|c| c["sha"].as_str().map(str::to_string))
            .collect();
        let issue = pr.issue.unwrap_or_default();
        for c in &commits {
            for id in claim_ids_in(c["commit"]["message"].as_str().unwrap_or_default(), issue) {
                if !pr.claim_ids.contains(&id) {
                    pr.claim_ids.push(id);
                }
            }
        }
        pr.comments = read_comments(pages(
            client,
            format!("{repo}/issues/{}/comments?per_page=100", pr.number),
        )?);
    }
    let worktrees = crate::worktree::worktrees(project).map(|list| {
        list.into_iter()
            .map(|e| WorktreeData {
                path: e.path,
                head: e.head,
                branch: e
                    .branch
                    .map(|b| b.strip_prefix("refs/heads/").unwrap_or(&b).to_string()),
            })
            .collect()
    });
    Ok((
        Snapshot {
            issues,
            other_open,
            prs,
            worktrees,
            now: clock.now(),
        },
        raw,
    ))
}

/// Every 40-hex token in `text`, in order of appearance, without repeats.
fn full_shas(text: &str) -> Vec<String> {
    let mut out: Vec<String> = Vec::new();
    for word in text.split(|c: char| !c.is_ascii_hexdigit()) {
        if word.len() == 40
            && word
                .bytes()
                .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
            && !out.iter().any(|s| s == word)
        {
            out.push(word.to_string());
        }
    }
    out
}

/// The last `AUDIT … <head> … verdict: PASS|FAIL` line naming `head`, with its comment id.
fn verdict_on(comments: &[Comment], head: &str) -> Option<(&'static str, i64)> {
    let mut found = None;
    for c in comments {
        for line in c.body.lines() {
            if !(line.contains("AUDIT #") && line.contains(head)) {
                continue;
            }
            match line.split("verdict:").nth(1).map(str::trim) {
                Some(v) if v.starts_with("PASS") => found = Some(("PASS", c.id)),
                Some(v) if v.starts_with("FAIL") => found = Some(("FAIL", c.id)),
                _ => {}
            }
        }
    }
    found
}

/// The newest owned report (`forge-report` marker) after comment `after`, and its `상태:`.
fn report_after(comments: &[Comment], after: i64) -> Option<(String, i64)> {
    comments
        .iter()
        .filter(|c| c.id > after && c.body.contains("<!-- moai:report run="))
        .filter_map(|c| {
            let status = c
                .body
                .lines()
                .find_map(|l| l.trim().strip_prefix("상태:"))?
                .trim()
                .to_string();
            Some((status, c.id))
        })
        .next_back()
}

/// Duplicate, out-of-order and unreported heads of one PR, judged by its commit order.
fn events(comments: &[Comment], pr: &PrData) -> Value {
    let index: BTreeMap<&str, usize> = pr
        .commits
        .iter()
        .enumerate()
        .map(|(i, s)| (s.as_str(), i))
        .collect();
    let mut reported: BTreeMap<String, i64> = BTreeMap::new();
    let mut newest: Option<(usize, String, i64)> = None;
    let (mut duplicate, mut out_of_order, mut ignored) = (Vec::new(), Vec::new(), 0usize);
    for c in comments {
        let shas = full_shas(&c.body);
        let mut in_pr: Vec<(usize, String)> = Vec::new();
        for s in shas {
            match index.get(s.as_str()) {
                Some(i) => in_pr.push((*i, s)),
                None => ignored += 1,
            }
        }
        // A comment that names several of the PR's commits reports the newest of them.
        let Some((i, head)) = in_pr.into_iter().max() else {
            continue;
        };
        // An older head after a newer one arrived out of order, whether or not it was
        // reported before; otherwise a head reported again is a duplicate.
        match &newest {
            Some((ni, nh, nc)) if i < *ni => {
                out_of_order.push(json!({
                    "comment_id": c.id, "head": head, "newer_head": nh, "newer_reported_by": nc
                }));
                reported.entry(head).or_insert(c.id);
            }
            _ => match reported.get(&head) {
                Some(first) => duplicate
                    .push(json!({"comment_id": c.id, "head": head, "first_reported_by": first})),
                None => {
                    reported.insert(head.clone(), c.id);
                    newest = Some((i, head, c.id));
                }
            },
        }
    }
    let unreported = (!reported.contains_key(&pr.head)).then(
        || json!({"head": pr.head, "last_reported_head": newest.as_ref().map(|n| n.1.clone())}),
    );
    json!({
        "duplicate": duplicate,
        "out_of_order": out_of_order,
        "unreported_head": unreported,
        "ignored_shas": ignored,
        "order_source": format!("pulls/{}/commits", pr.number),
    })
}

/// One (session, Issue) row before it becomes JSON.
struct Row {
    session: String,
    issue: i64,
    phase: Phase,
    value: Value,
    open_blockers: Vec<i64>,
}

fn rows_for(issue: &IssueData, snap: &Snapshot, names: &BTreeMap<&str, Option<&str>>) -> Vec<Row> {
    let comments: Vec<IssueComment> = issue
        .comments
        .iter()
        .map(|c| IssueComment {
            id: c.id,
            author: String::new(),
            body: c.body.clone(),
        })
        .collect();
    let verdict = evaluate(
        &comments,
        &Verifier {
            session: "",
            comment_id: None,
            allowed_authors: &[],
            now: snap.now,
        },
    );
    let kind_of = |id: i64| {
        issue
            .comments
            .iter()
            .find(|c| c.id == id)
            .map(|c| classify_comment(&c.body))
    };
    // The holder, or the session a handoff named when it has not claimed yet.
    let (session, ownership, handoff) = match verdict["holder"].as_i64() {
        Some(h) => match kind_of(h) {
            Some(CommentKind::Claim {
                session,
                epoch,
                lease_until,
            }) => (
                session,
                json!({"claim_comment_id": h, "epoch": epoch, "lease_until": lease_until,
                       "live": verdict["holder_live"]}),
                None,
            ),
            _ => return Vec::new(),
        },
        None => {
            let last = verdict["closed"]
                .as_array()
                .and_then(|a| a.iter().max_by_key(|c| c["closing_comment_id"].as_i64()))
                .cloned();
            match last {
                Some(c) if c["by"] == json!("handoff") => (
                    c["to"].as_str().unwrap_or_default().to_string(),
                    json!({"claim_comment_id": null, "handed_off_from": c["comment_id"]}),
                    c["closing_comment_id"].as_i64(),
                ),
                _ => return Vec::new(),
            }
        }
    };
    let pr = snap.prs.iter().find(|p| p.issue == Some(issue.number));
    let worktree = snap.worktrees.as_ref().ok().and_then(|list| {
        list.iter().find(|w| {
            w.branch
                .as_deref()
                .and_then(crate::claim_verify::issue_from_branch)
                == Some(issue.number)
        })
    });
    let holder_id = ownership["claim_comment_id"].as_i64().unwrap_or(0);
    let (phase, evidence) = if let Some(h) = handoff {
        (
            Phase::Handoff,
            json!({"handoff_comment_id": h, "claim_by_recipient": null}),
        )
    } else if let Some(pr) = pr {
        match verdict_on(&issue.comments, &pr.head) {
            Some(("PASS", id)) => (
                Phase::Waiting,
                json!({"pr": pr.number, "head": pr.head, "verdict": "PASS", "verdict_comment_id": id}),
            ),
            Some((v, id)) => (
                Phase::Implementing,
                json!({"pr": pr.number, "head": pr.head, "verdict": v, "verdict_comment_id": id}),
            ),
            None => (
                Phase::PrOpen,
                json!({"pr": pr.number, "head": pr.head, "verdict": null}),
            ),
        }
    } else if let Some(w) = worktree {
        (
            Phase::Implementing,
            json!({"worktree": w.path, "worktree_head": w.head}),
        )
    } else if let Some((status, id)) =
        report_after(&issue.comments, holder_id).filter(|(s, _)| s == "running")
    {
        (
            Phase::Implementing,
            json!({"report_comment_id": id, "report_status": status}),
        )
    } else {
        (Phase::Claimed, json!({"claim_comment_id": holder_id}))
    };
    let open_blockers: Vec<i64> = crate::queue::refs_on(&issue.body, "Blocked-By:")
        .into_iter()
        .filter(|b| *b != issue.number && blocker_open(*b, snap))
        .collect();
    let last_event_at = issue
        .comments
        .iter()
        .filter(|c| {
            classify_comment(&c.body) != CommentKind::Other
                || c.body.contains("AUDIT #")
                || c.body.contains("<!-- moai:report run=")
                || pr.is_some_and(|p| full_shas(&c.body).iter().any(|s| p.commits.contains(s)))
        })
        .map(|c| c.created_at.clone())
        .max();
    let value = json!({
        "session": session,
        "lane": names.get(session.as_str()).copied().flatten(),
        "issue": issue.number,
        "branch": pr.map(|p| p.branch.clone()).or_else(|| worktree.and_then(|w| w.branch.clone())),
        "worktree": worktree.map(|w| w.path.clone()),
        "pr": pr.map(|p| p.number),
        "head": pr.map(|p| p.head.clone()).or_else(|| worktree.map(|w| w.head.clone())),
        "phase": phase.as_str(),
        "phase_evidence": evidence,
        "ownership": ownership,
        "open_blockers": open_blockers,
        "last_event_at": last_event_at,
        "events": pr.map(|p| events(&issue.comments, p)),
        "audit": pr.map(|p| audit_value(issue, p)),
    });
    vec![Row {
        session,
        issue: issue.number,
        phase,
        value,
        open_blockers,
    }]
}

fn blocker_open(n: i64, snap: &Snapshot) -> bool {
    match snap.issues.iter().find(|i| i.number == n) {
        Some(i) => i.open,
        None => snap.other_open.get(&n).copied().unwrap_or(false),
    }
}

/// Issues with no claim-family line and no open blocker (the `queue-next` rule).
fn ready(snap: &Snapshot) -> Vec<i64> {
    snap.issues
        .iter()
        .filter(|i| i.open)
        .filter(|i| {
            i.comments
                .iter()
                .all(|c| classify_comment(&c.body) == CommentKind::Other)
        })
        .filter(|i| {
            crate::queue::refs_on(&i.body, "Blocked-By:")
                .into_iter()
                .all(|b| b == i.number || !blocker_open(b, snap))
        })
        .map(|i| i.number)
        .collect()
}

fn idle_entry(
    lane: &LaneInput,
    rows: &[&Row],
    req: &LaneStatusRequest,
    ready: &[i64],
) -> (IdleCause, Option<bool>, Value) {
    if let Some(p) = &lane.pane {
        let seen = json!({"source": p.source, "observed_at": p.observed_at});
        match p.state {
            PaneState::Unreachable => return (IdleCause::Unreachable, Some(true), seen),
            PaneState::PermissionPrompt => return (IdleCause::PermissionPrompt, Some(true), seen),
            PaneState::InputPending => return (IdleCause::InputPending, Some(true), seen),
            PaneState::Working => {}
        }
    }
    if rows.is_empty() {
        if lane
            .pane
            .as_ref()
            .is_some_and(|p| p.state == PaneState::Working)
        {
            return (
                IdleCause::Unknown,
                Some(false),
                json!({"pane": "working", "rows": []}),
            );
        }
        return if ready.is_empty() {
            (
                IdleCause::NoReadyCard,
                Some(true),
                json!({"inspected": req.issues.iter().collect::<BTreeSet<_>>(), "ready": []}),
            )
        } else {
            (
                IdleCause::ReadyCardUnassigned,
                Some(true),
                json!({"ready": ready}),
            )
        };
    }
    // A card held up by a record outranks the lane's activity.
    for r in rows {
        if let Some(d) = req.pending_decisions.iter().find(|d| d.issue == r.issue) {
            return (
                IdleCause::UserDecision,
                Some(true),
                json!({"issue": r.issue, "source": d.source, "recorded_at": d.recorded_at}),
            );
        }
        if !r.open_blockers.is_empty() {
            return (
                IdleCause::WaitingPrerequisite,
                Some(true),
                json!({"issue": r.issue, "blocked_by": r.open_blockers}),
            );
        }
    }
    if lane
        .pane
        .as_ref()
        .is_some_and(|p| p.state == PaneState::Working)
    {
        let p = lane
            .pane
            .as_ref()
            .map(|p| json!({"source": p.source, "observed_at": p.observed_at}));
        return (
            IdleCause::Unknown,
            Some(false),
            json!({"pane": "working", "observed": p}),
        );
    }
    let active: Vec<i64> = rows
        .iter()
        .filter(|r| r.phase.active())
        .map(|r| r.issue)
        .collect();
    if !active.is_empty() {
        return (
            IdleCause::Unknown,
            None,
            json!({"active_issues": active, "pane": null,
                   "note": "an active card and no pane observation; idleness is not observable"}),
        );
    }
    let r = rows[0];
    let ev = r.value["phase_evidence"].clone();
    match r.phase {
        Phase::PrOpen => (IdleCause::WaitingAudit, Some(true), ev),
        Phase::Waiting => (IdleCause::WaitingIntegration, Some(true), ev),
        _ => (IdleCause::Unknown, None, ev),
    }
}

/// The table, idle classification and assignment drafts for one snapshot (no I/O).
pub fn build(req: &LaneStatusRequest, snap: &Snapshot) -> Value {
    let names: BTreeMap<&str, Option<&str>> = req
        .lanes
        .iter()
        .map(|l| (l.session.as_str(), l.name.as_deref()))
        .collect();
    let mut rows: Vec<Row> = snap
        .issues
        .iter()
        .flat_map(|i| rows_for(i, snap, &names))
        .collect();
    rows.sort_by(|a, b| (&a.session, a.issue).cmp(&(&b.session, b.issue)));
    let ready = ready(snap);
    let mut idle = Vec::new();
    let mut free = Vec::new();
    for lane in &req.lanes {
        let mine: Vec<&Row> = rows.iter().filter(|r| r.session == lane.session).collect();
        let (cause, is_idle, evidence) = idle_entry(lane, &mine, req, &ready);
        if matches!(cause, IdleCause::ReadyCardUnassigned) {
            free.push(lane.session.clone());
        }
        idle.push(json!({
            "session": lane.session,
            "lane": lane.name,
            "idle": is_idle,
            "cause": if is_idle == Some(false) { Value::Null } else { json!(cause.as_str()) },
            "evidence": evidence,
        }));
    }
    let drafts: Vec<Value> = free
        .iter()
        .zip(ready.iter())
        .map(|(lane, issue)| json!({"lane": lane, "issue": issue, "posted": false}))
        .collect();
    let worktrees = match &snap.worktrees {
        Ok(list) => json!({"available": true, "count": list.len()}),
        Err(e) => json!({"available": false, "reason": e}),
    };
    // Issue #2678: say the row count in words, and say "no result" outright, so a reader never
    // takes an empty list for a count it did not make.
    let summary = if rows.is_empty() {
        "0 rows: no claim, pull request or worktree was found for the named Issues".to_string()
    } else {
        format!(
            "{} row(s); {} idle lane(s); {} ready Issue(s)",
            rows.len(),
            idle.len(),
            ready.len()
        )
    };
    json!({
        "summary": summary,
        "rows": rows.into_iter().map(|r| r.value).collect::<Vec<_>>(),
        "idle": idle,
        "ready": ready,
        "assignment_drafts": drafts,
        "worktrees": worktrees,
        "writes": "none",
        "note": "pane states and pending decisions are the caller's records, cited as given",
    })
}

// S6 consumes the same complete Issue snapshot as the original table. It never opens
// the message journal: reading status must not create a lock, receipt or directory.
#[derive(Deserialize, Serialize, PartialEq)]
#[serde(deny_unknown_fields)]
struct MessageWire {
    schema_version: u32,
    kind: String,
    message: Message,
    payload_digest: String,
}

struct ReportObservation {
    message: Message,
    record: Option<(u64, i64)>,
    ack: Option<(u64, i64)>,
}

#[derive(Default, Serialize)]
struct CommunicationLane {
    session: String,
    sent_reports: Vec<Value>,
    received_acks: Vec<Value>,
    unacknowledged_report_ids: Vec<String>,
    dispatches: Vec<Value>,
}

struct ObservationClock<'a, C> {
    now: i64,
    source: &'a C,
}

impl<C: Clock> Clock for ObservationClock<'_, C> {
    fn now(&self) -> i64 {
        self.now
    }

    fn sleep(&self, secs: u64) {
        self.source.sleep(secs);
    }
}

// S2 retains ownership of dispatch revision/decision validation. Its GET-only checker
// receives this frozen, already-paginated snapshot, not a second live network read.
struct MessageSnapshot<'a> {
    base: String,
    repository: &'a Value,
    issue: &'a ReadIssue,
}

impl Transport for MessageSnapshot<'_> {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse> {
        if req.method != "GET" || req.body.is_some() {
            return Err("lane status snapshot permits GET only".into());
        }
        let body = if req.url == self.base {
            self.repository
        } else if req.url == format!("{}/issues/{}", self.base, self.issue.number) {
            &self.issue.issue
        } else if req.url
            == format!(
                "{}/issues/{}/comments?per_page=100",
                self.base, self.issue.number
            )
        {
            &self.issue.comments
        } else {
            return Err("lane status snapshot target unavailable".into());
        };
        Ok(HttpResponse::new(200).with_body(body.clone()))
    }
}

fn has_message(comment: &Value) -> bool {
    comment["body"].as_str().is_some_and(|body| {
        body.split('\n')
            .next()
            .unwrap_or_default()
            .split("<!--")
            .skip(1)
            .any(|inner| {
                let inner = inner.trim_start();
                inner
                    .strip_prefix("moai:op ")
                    .or_else(|| inner.strip_prefix("moai-op "))
                    .is_some_and(|operation| {
                        let operation = operation.trim_start();
                        operation.starts_with("message-record:")
                            || operation.starts_with("message-ack:")
                    })
            })
    })
}

fn message_observations(
    req: &LaneStatusRequest,
    issue: &ReadIssue,
    repository: &Value,
    now: i64,
) -> Result<Vec<ReportObservation>> {
    let invalid = || "message publication evidence unavailable".to_string();
    let mut reports: BTreeMap<String, ReportObservation> = BTreeMap::new();
    let mut comments: BTreeMap<u64, (MessageWire, i64)> = BTreeMap::new();
    for comment in issue.comments.as_array().ok_or_else(invalid)? {
        if !has_message(comment) {
            continue;
        }
        let body = comment["body"].as_str().ok_or_else(invalid)?;
        let mut wire: MessageWire = body
            .split('\n')
            .nth(1)
            .and_then(|line| serde_json::from_str(line).ok())
            .ok_or_else(invalid)?;
        let digest = crate::message::validate(&mut wire.message).map_err(|_| invalid())?;
        let ack = wire.kind == "ack";
        let m = &wire.message;
        let task = &m.task;
        let operation = if ack { "message-ack" } else { "message-record" };
        let marker = crate::forge::outbox::op_marker(operation, &m.project_id, &m.report_id, 1);
        let rendered = format!(
            "{marker}\n{}\n<!-- moai:message:end -->",
            serde_json::to_string(&wire).map_err(|_| invalid())?
        );
        if wire.schema_version != 1
            || !matches!(wire.kind.as_str(), "record" | "ack")
            || wire.payload_digest != digest
            || body != rendered
            || task.owner != req.owner
            || task.repo != req.repo
            || repository["id"].as_u64() != Some(task.repository_id)
            || task.issue_number != issue.number as u64
            || issue.issue["number"].as_u64() != Some(task.issue_number)
            || issue.issue["id"].as_u64() != Some(task.issue_id)
            || issue.issue["node_id"].as_str() != Some(task.issue_node_id.as_str())
            || issue
                .issue
                .get("pull_request")
                .is_some_and(|v| !v.is_null())
            || comment["user"]["id"].as_u64()
                != Some(if ack {
                    m.receiver.actor_id
                } else {
                    m.sender.actor_id
                })
        {
            return Err(invalid());
        }
        let id = comment["id"]
            .as_u64()
            .filter(|id| *id > 0)
            .ok_or_else(invalid)?;
        let at = comment["created_at"]
            .as_str()
            .and_then(parse_utc)
            .filter(|at| now.checked_sub(*at).is_some_and(|age| age >= 0))
            .ok_or_else(invalid)?;
        if let Some((previous, previous_at)) = comments.get(&id) {
            if previous != &wire || *previous_at != at {
                return Err(invalid());
            }
            continue;
        }
        let report = reports
            .entry(m.report_id.clone())
            .or_insert_with(|| ReportObservation {
                message: m.clone(),
                record: None,
                ack: None,
            });
        if report.message != *m {
            return Err(invalid());
        }
        let slot = if ack {
            &mut report.ack
        } else {
            &mut report.record
        };
        if slot.is_none_or(|(old, _)| id < old) {
            *slot = Some((id, at));
        }
        comments.insert(id, (wire, at));
    }
    for report in reports.values() {
        let (_, record_at) = report.record.ok_or_else(invalid)?;
        if report.ack.is_some_and(|(_, at)| at < record_at) {
            return Err(invalid());
        }
    }
    Ok(reports.into_values().collect())
}

fn communication_lane<'a>(
    lanes: &'a mut BTreeMap<String, CommunicationLane>,
    session: &str,
) -> &'a mut CommunicationLane {
    lanes
        .entry(session.to_string())
        .or_insert_with(|| CommunicationLane {
            session: session.to_string(),
            ..Default::default()
        })
}

// Collection identity belongs to the S6 observation, never to shared pagination.
// Numeric aliases are accepted only for the repository whose metadata was verified.
fn comment_collection_matches(issue: &ReadIssue, repository_id: Option<u64>) -> bool {
    let Some(first) = issue.source_urls.first() else {
        return false;
    };
    let collection = first.split('?').next().unwrap_or_default();
    let same_slug = |url: &str| {
        url.strip_prefix(collection)
            .is_some_and(|suffix| suffix.starts_with('?'))
    };
    if issue.source_urls.iter().all(|url| same_slug(url)) {
        return true;
    }
    let Some(repository_id) = repository_id.filter(|id| *id > 0) else {
        return false;
    };
    let numeric = format!(
        "{API}/repositories/{repository_id}/issues/{}/comments?",
        issue.number
    );
    issue
        .source_urls
        .iter()
        .all(|url| same_slug(url) || url.starts_with(&numeric))
}

fn communication<T: Transport, C: Clock>(
    req: &LaneStatusRequest,
    raw: &[ReadIssue],
    client: &Client<'_, T, C>,
    clock: &C,
    now: i64,
) -> Value {
    let mut lanes = BTreeMap::new();
    for lane in &req.lanes {
        communication_lane(&mut lanes, &lane.session);
    }
    let mut errors: Vec<Value> = raw
        .iter()
        .filter(|issue| !issue.complete)
        .map(|issue| {
            json!({"issue":issue.number,
            "reason":"message comments snapshot incomplete"})
        })
        .collect();
    let has_packets = raw.iter().any(|i| {
        i.comments
            .as_array()
            .is_some_and(|comments| comments.iter().any(has_message))
    });
    if !has_packets
        && raw
            .iter()
            .all(|issue| comment_collection_matches(issue, None))
    {
        return json!({"available":errors.is_empty(), "observed_at":now,
            "lanes":lanes.into_values().collect::<Vec<_>>(), "errors":errors});
    }
    let base = format!("{API}/repos/{}/{}", req.owner, req.repo);
    let repository = match one(client, &base) {
        Ok(repository) => repository,
        Err(_) => {
            return json!({"available":false, "observed_at":now,
            "lanes":lanes.into_values().collect::<Vec<_>>(),
            "errors":[{"issue":null,"reason":"message repository evidence unavailable"}]})
        }
    };
    let observation_clock = ObservationClock { now, source: clock };
    let mut report_ids = BTreeSet::new();
    for issue in raw {
        if !issue.complete {
            continue;
        }
        let observations = (|| {
            if !comment_collection_matches(issue, repository["id"].as_u64()) {
                return Err("message comments collection mismatch".to_string());
            }
            let reports = message_observations(req, issue, &repository, now)?;
            // A report cannot change its task by moving to another inspected Issue.
            if reports
                .iter()
                .any(|report| report_ids.contains(&report.message.report_id))
            {
                return Err("message immutable identity conflict".to_string());
            }
            let frozen = MessageSnapshot {
                base: base.clone(),
                repository: &repository,
                issue,
            };
            let mut dispatches = BTreeMap::new();
            for report in &reports {
                if report.message.event == "dispatch" {
                    let view = crate::dispatch::check_cli(
                        crate::dispatch::CheckRequest {
                            dispatch: report.message.clone(),
                        },
                        &frozen,
                        &observation_clock,
                    )
                    .map_err(|_| "dispatch observation unavailable".to_string())?;
                    dispatches.insert(report.message.report_id.clone(), view);
                }
            }
            // Route every referenced decision into the same S2 task group for validation.
            for report in &reports {
                if report.message.event == "dispatch-decision"
                    && !reports.iter().any(|r| {
                        r.message.event == "dispatch"
                            && r.message.project_id == report.message.project_id
                            && r.message.task == report.message.task
                            && r.message.report_id == report.message.payload["dispatch_report_id"]
                    })
                {
                    return Err("dispatch observation unavailable".to_string());
                }
            }
            Ok::<_, String>((reports, dispatches))
        })();
        let (reports, mut dispatches) = match observations {
            Ok(observations) => observations,
            Err(reason) => {
                errors.push(json!({"issue":issue.number,"reason":reason}));
                continue;
            }
        };
        for report in reports {
            let m = report.message;
            report_ids.insert(m.report_id.clone());
            let (record_id, at) = report.record.expect("validated record");
            let elapsed = now.checked_sub(at);
            let alert = if report.ack.is_some() {
                "received"
            } else if elapsed.is_some_and(|n| n >= 300) {
                "unacknowledged"
            } else {
                "waiting"
            };
            communication_lane(&mut lanes, &m.receiver.session_id);
            let lane = communication_lane(&mut lanes, &m.sender.session_id);
            lane.sent_reports.push(json!({
                "report_id":m.report_id,"issue":issue.number,"event":m.event,
                "to":m.receiver.session_id,"record_comment_id":record_id,
                "ack_comment_id":report.ack.map(|(id,_)|id),"received_ack":report.ack.is_some(),
                "ack_alert":alert,"elapsed_secs":elapsed,
            }));
            if let Some((id, _)) = report.ack {
                lane.received_acks.push(json!({"report_id":m.report_id,
                    "issue":issue.number,"ack_comment_id":id}));
            }
            if alert == "unacknowledged" {
                lane.unacknowledged_report_ids.push(m.report_id.clone());
            }
            if let Some(mut dispatch) = dispatches.remove(&m.report_id) {
                dispatch["issue"] = json!(issue.number);
                dispatch["from"] = json!(m.sender.session_id);
                dispatch["to"] = json!(m.receiver.session_id);
                // A self-addressed packet appears once, not twice in the same lane.
                if m.sender.session_id != m.receiver.session_id {
                    communication_lane(&mut lanes, &m.receiver.session_id)
                        .dispatches
                        .push(dispatch.clone());
                }
                communication_lane(&mut lanes, &m.sender.session_id)
                    .dispatches
                    .push(dispatch);
            }
        }
    }
    if !errors.is_empty() {
        for lane in lanes.values_mut() {
            lane.sent_reports.clear();
            lane.received_acks.clear();
            lane.unacknowledged_report_ids.clear();
            lane.dispatches.clear();
        }
    }
    json!({"available":errors.is_empty(),"observed_at":now,
        "lanes":lanes.into_values().collect::<Vec<_>>(),"errors":errors})
}

/// CLI: validate, read (GET only), build. Writes nothing.
pub fn status_cli<T: Transport, C: Clock>(
    project: &Path,
    req: LaneStatusRequest,
    transport: &T,
    clock: &C,
) -> Result<Value> {
    validate(&req)?;
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let (snap, raw) = collect_with_raw(project, &req, &client, clock)?;
    let mut table = build(&req, &snap);
    table["communication"] = communication(&req, &raw, &client, clock, snap.now);
    Ok(table)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn full_shas_take_only_40_lowercase_hex_words() {
        let a = "a".repeat(40);
        let text = format!(
            "head `{a}`, short {} upper {} long {}",
            &a[..10],
            a.to_uppercase(),
            "b".repeat(41)
        );
        assert_eq!(full_shas(&text), vec![a.clone()]);
        assert_eq!(full_shas(&format!("{a} {a}")), vec![a]);
    }

    #[test]
    fn verdict_on_reads_the_last_line_for_the_head() {
        let c = |id, body: &str| Comment {
            id,
            created_at: String::new(),
            body: body.into(),
        };
        let list = vec![
            c(1, "AUDIT #7 (k) head hh base b — verdict: FAIL"),
            c(2, "AUDIT #7 (k) head hh base b — verdict: PASS"),
            c(3, "AUDIT #7 (k) head other base b — verdict: FAIL"),
        ];
        assert_eq!(verdict_on(&list, "hh"), Some(("PASS", 2)));
        assert_eq!(verdict_on(&list, "zz"), None);
    }

    #[test]
    fn report_after_needs_the_marker_and_a_later_id() {
        let c = |id, body: &str| Comment {
            id,
            created_at: String::new(),
            body: body.into(),
        };
        let list = vec![
            c(5, "<!-- moai:report run=r1 -->\n상태: running"),
            c(9, "plain prose 상태: running"),
            c(12, "<!-- moai:report run=r2 -->\n상태: blocked"),
        ];
        assert_eq!(report_after(&list, 4), Some(("blocked".into(), 12)));
        assert_eq!(report_after(&list, 12), None);
    }

    /// #2678: an empty table says "0 rows" and why, never just `[]`.
    #[test]
    fn an_empty_table_says_so_in_words() {
        let req = LaneStatusRequest {
            schema_version: 1,
            owner: "o".into(),
            repo: "r".into(),
            base_branch: "main".into(),
            issues: vec![],
            lanes: vec![],
            pending_decisions: vec![],
        };
        let snap = Snapshot {
            issues: vec![],
            other_open: BTreeMap::new(),
            prs: vec![],
            worktrees: Ok(vec![]),
            now: 0,
        };
        let out = build(&req, &snap);
        assert_eq!(out["rows"], json!([]));
        assert_eq!(
            out["summary"],
            json!("0 rows: no claim, pull request or worktree was found for the named Issues")
        );
    }
}
