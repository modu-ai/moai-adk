//! `lane-next` (#2831 C1): the one thing a lane should do next, and why, without a leader.
//!
//! Lanes learned their next task only from the leader reading their screens: a delta re-audit
//! whose head had moved went unnoticed (#2821), a decision already posted was still awaited
//! (#2827 6868439066), a re-audit request was missed (#2820). This reads GitHub (open pull
//! requests and their comments, the lane's claim threads, a few open Issues) and the tip's CI,
//! and picks one task in the order of the user's priority list (#1975 6855190942):
//!
//! 1. P0 the tip is red (`gate_collect::tip_ci`, the D8(a) judgment the merge gate uses);
//! 2. P1 the lane's own pull request: an audit FAIL on its head, a conflict, a failed CI; then
//!    a claimed card whose pull request merged with no close table posted (SYNC, #2880 W3);
//! 3. P1 a pull request the lane reserved to audit: not judged yet, its head moved past the
//!    lane's verdict (an update-branch too) or a re-audit was asked for — with the head now;
//! 4. P1 another lane's pull request no auditor reserved;
//! 5. a `LANE-BLOCKED kind=decision` of the lane that a `LEADER-DECISION` now answers;
//! 6. P2 the next step of a card the lane claims (no open pull request of its own for it);
//! 7. P2–P4 an open Issue nobody claims.
//!
//! Steps 6 and 7 take only a startable card ([`crate::startable::judge`], the judgment
//! `goal-block` and `queue-free` share, #2901): one on hold or waiting for a decision, left out
//! by the active goal, an epic or blocked is named under `skipped` with its reason, and the
//! list goes on. `end_condition` is the lane-wide end: no open pull request and no startable
//! open Issue.
//!
//! Reads only (leader decision D1): reservations and comments stay with the existing commands.
//! One collection sends at most [`API_BUDGET`] GitHub requests (D4), taking the higher
//! priorities first; what the budget left unread is named in `partial`, never guessed. The open
//! pull requests with their comments, the lane's threads and the open Issues come from a few
//! GraphQL requests, not one request per pull request (#2873 N1), so a long list no longer
//! spends the budget before the lower priorities are read.

use crate::forge::policy::Clock;
use crate::forge::transport::{HttpRequest, Transport};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::cell::Cell;
use std::collections::BTreeMap;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// GitHub requests one collection may send (#2831 D4): ten lanes polling stay far below the
/// 5,000-per-hour limit of one token.
pub const API_BUDGET: u32 = 30;
/// Startable open Issues (the oldest) whose comments are read to find one nobody claims.
pub const FREE_ISSUE_SCAN: usize = 5;
/// Lanes assumed to poll at once with one GitHub token (they all read as one account), unless
/// `--shared-lanes` names the number (#2873 N9).
pub const SHARED_LANES: u64 = 10;
/// The most lanes `--shared-lanes` accepts: the poll interval then stays below
/// [`MAX_WAIT_SECS`] (3600 × 30 × 50 / 2500 = 2160).
pub const MAX_SHARED_LANES: u64 = 50;
/// The part of GitHub's 5,000 requests per hour per token the lanes' polling may use: half,
/// so the leader loop and the merge gate keep the other half (#2856 audit B1).
pub const HOURLY_SHARE: u64 = 2_500;
// The share is at most half of GitHub's 5,000 per hour, checked when the crate compiles.
const _: () = assert!(HOURLY_SHARE * 2 <= 5_000);
/// Seconds between two collections while `--wait` waits, derived from the budget so that
/// `shared_lanes` lanes each sending `API_BUDGET` requests per collection stay within
/// `HOURLY_SHARE` per hour: ⌈3600 × 30 × lanes / 2500⌉.
pub const fn poll_secs(shared_lanes: u64) -> u64 {
    (3600 * API_BUDGET as u64 * shared_lanes).div_ceil(HOURLY_SHARE)
}
/// The interval for the default [`SHARED_LANES`]: 3600 × 30 × 10 / 2500 = 432.
pub const POLL_SECS: u64 = poll_secs(SHARED_LANES);
/// The longest `--wait` accepted (seconds).
pub const MAX_WAIT_SECS: u64 = 3600;

/// One comment as read: its id and body (oldest first within a thread).
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Comment {
    pub id: u64,
    pub body: String,
    /// When it was posted (ISO 8601 UTC); empty when unknown.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub at: String,
}

/// One open pull request as read.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Pr {
    pub number: u64,
    #[serde(default)]
    pub title: String,
    /// The head commit now, 40 hex.
    pub head: String,
    #[serde(default)]
    pub draft: bool,
    /// The branch it targets (`baseRefName`); empty when not read.
    #[serde(default)]
    pub base: String,
    /// GitHub's `authorAssociation` of its author (`OWNER`, `MEMBER`, `COLLABORATOR`,
    /// `CONTRIBUTOR`, …); empty when not read.
    #[serde(default)]
    pub author_association: String,
    /// The lane whose claim the body's `Moai-Claim:` names; `None` when unknown.
    #[serde(default)]
    pub author_lane: Option<String>,
    /// The Issue the body's `Moai-Claim:` names.
    #[serde(default)]
    pub claim_issue: Option<u64>,
    /// `true` when GitHub reports a merge conflict (`mergeable: false` / state `dirty`); read
    /// only for the lane's own pull requests.
    #[serde(default)]
    pub conflicting: Option<bool>,
    /// The head's check runs: `green`, `red`, `pending` or `unknown`; read only for the lane's
    /// own pull requests.
    #[serde(default)]
    pub ci: Option<String>,
    /// The failed checks with their run ids, when `ci` is `red`.
    #[serde(default)]
    pub ci_detail: Option<String>,
    /// The body carries a `moai-tip-repair:` line: the pull request that repairs a red tip.
    #[serde(default)]
    pub tip_repair: bool,
    #[serde(default)]
    pub comments: Vec<Comment>,
}

/// A merged pull request that refers to a thread, with the claim its body's `Moai-Claim:`
/// names (#2873 N8).
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct MergedPr {
    pub number: u64,
    #[serde(default)]
    pub claim: Option<u64>,
    /// When it was merged (ISO 8601 UTC); empty when unknown.
    #[serde(default)]
    pub merged_at: String,
}

/// An Issue thread read for the lane's claims and decision requests.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Thread {
    pub number: u64,
    #[serde(default)]
    pub title: String,
    #[serde(default)]
    pub labels: Vec<String>,
    #[serde(default)]
    pub comments: Vec<Comment>,
    /// Merged pull requests that refer to this thread.
    #[serde(default)]
    pub merged: Vec<MergedPr>,
    /// The body, for its `Blocked-By:` line (#2901).
    #[serde(default)]
    pub body: String,
    /// Sub-issues GitHub reports still open (#2901).
    #[serde(default)]
    pub open_children: u64,
    /// Open Issues GitHub reports as blocking this one (#2901).
    #[serde(default)]
    pub blocked_by: Vec<u64>,
}

/// What [`crate::startable::judge`] reads of a thread.
fn card(t: &Thread) -> crate::startable::Card<'_> {
    crate::startable::Card {
        number: t.number,
        title: &t.title,
        labels: &t.labels,
        body: &t.body,
        open_children: t.open_children,
        blocked_by: &t.blocked_by,
    }
}

/// The highest Issue number of the open list when it was read only in part (oldest first,
/// `cards_truncated`); `None` when the list is whole or was not read. A `Blocked-By:`
/// reference above it is of unknown state (PR #2916 review).
fn read_through(snap: &Snapshot) -> Option<u64> {
    if !snap.cards_truncated {
        return None;
    }
    Some(
        snap.cards
            .iter()
            .flatten()
            .map(|t| t.number)
            .max()
            .unwrap_or(0),
    )
}

/// The tip's CI as the merge gate judges it (`gate::TipCi` fields).
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Tip {
    pub state: String,
    #[serde(default)]
    pub tip: String,
    #[serde(default)]
    pub decided_by: Option<String>,
    #[serde(default)]
    pub detail: String,
}

/// Everything one decision reads. Tests and `--snapshot` give it as JSON; `collect` reads it.
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Snapshot {
    /// When it was read (`YYYY-MM-DDTHH:MM:SSZ`): given by a saved snapshot, or set from the
    /// clock by [`run`]; echoed as `now` (#2933 S1).
    #[serde(default)]
    pub now: Option<String>,
    #[serde(default)]
    pub tip: Option<Tip>,
    /// The integration branch the tip was read from; `None` when not known.
    #[serde(default)]
    pub tip_branch: Option<String>,
    #[serde(default)]
    pub prs: Vec<Pr>,
    /// Issue threads: the lane's claim threads and threads holding its decision requests.
    #[serde(default)]
    pub threads: Vec<Thread>,
    /// Open Issues read as candidates for unclaimed work, oldest first: the oldest startable
    /// ones of `cards` (#2901).
    #[serde(default)]
    pub open_issues: Vec<Thread>,
    /// Every open Issue the list request returned, without comments, oldest first — what the
    /// startable judgment and the end condition read (#2901); `None` when the list was not read
    /// (a snapshot saved before #2901 too).
    #[serde(default)]
    pub cards: Option<Vec<Thread>>,
    /// More open Issues exist than `cards` holds.
    #[serde(default)]
    pub cards_truncated: bool,
    /// The active goal's exclusions (`goal-show`, read from the local clone).
    #[serde(default)]
    pub keep_set: Vec<crate::goal::Exclusion>,
    /// What the request budget or a failed read left unread.
    #[serde(default)]
    pub partial: Vec<String>,
    #[serde(default)]
    pub api_calls: u32,
}

// --- comment lines -----------------------------------------------------------------------

fn lines(c: &Comment) -> impl Iterator<Item = &str> {
    c.body.lines().map(str::trim)
}

fn first_line(c: &Comment) -> &str {
    c.body.lines().next().unwrap_or("").trim()
}

fn field<'a>(line: &'a str, key: &str) -> Option<&'a str> {
    line.split_whitespace()
        .find_map(|w| w.strip_prefix(key)?.strip_prefix('='))
}

fn is_hex40(s: &str) -> bool {
    s.len() == 40 && s.bytes().all(|b| b.is_ascii_hexdigit())
}

/// `moai-audit-claim: lane=<lane> head=<40hex>` lines of a comment.
fn audit_claim(c: &Comment) -> Option<(String, String)> {
    lines(c).find_map(|l| {
        let rest = l.strip_prefix("moai-audit-claim:")?;
        Some((
            field(rest, "lane")?.to_string(),
            field(rest, "head")?.to_string(),
        ))
    })
}

/// A verdict comment: first line `AUDIT #<n> … head <40hex> … verdict: PASS|FAIL`.
fn verdict(c: &Comment) -> Option<(String, &'static str)> {
    let l = first_line(c);
    if !l.starts_with("AUDIT #") {
        return None;
    }
    let words: Vec<&str> = l.split_whitespace().collect();
    let head = words
        .windows(2)
        .find(|w| w[0] == "head" && is_hex40(w[1]))
        .map(|w| w[1].to_string())?;
    let v = if l.contains("verdict: PASS") {
        "PASS"
    } else if l.contains("verdict: FAIL") {
        "FAIL"
    } else {
        return None;
    };
    Some((head, v))
}

/// A comment asking the auditor to look again (a repair note or a re-audit request).
fn reaudit_request(c: &Comment) -> bool {
    let l = first_line(c);
    l.starts_with("REPAIR")
        || l.contains("재감사 요청")
        || l.to_lowercase().contains("re-audit request")
}

/// `moai-claim: session=<lane> …` — the claiming lane.
fn claim_session(c: &Comment) -> Option<String> {
    lines(c).find_map(|l| Some(field(l.strip_prefix("moai-claim:")?, "session")?.to_string()))
}

/// The claim comment ids a thread releases or hands off.
fn closed_claims(thread: &[Comment]) -> Vec<u64> {
    thread
        .iter()
        .flat_map(lines)
        .filter_map(|l| {
            let rest = l
                .strip_prefix("moai-claim-release:")
                .or_else(|| l.strip_prefix("moai-claim-handoff:"))?;
            field(rest, "comment")?.parse().ok()
        })
        .collect()
}

/// The lane's live claim in a thread: its comment id, when one is not released or handed off.
fn live_claim(thread: &[Comment], lane: &str) -> Option<u64> {
    let closed = closed_claims(thread);
    thread
        .iter()
        .rev()
        .find(|c| claim_session(c).as_deref() == Some(lane) && !closed.contains(&c.id))
        .map(|c| c.id)
}

/// A `LANE-BLOCKED` line of this lane: it names the lane, or it sits in a thread the lane
/// claims (lines written before the lane name was required).
fn blocked_line<'a>(c: &'a Comment, lane: &str, claimed: bool) -> Option<&'a str> {
    lines(c).find(|l| {
        l.starts_with("LANE-BLOCKED")
            && l.contains("kind=decision")
            && (claimed
                || l.split_whitespace()
                    .any(|w| w == lane || w == format!("lane={lane}")))
    })
}

fn leader_decision(c: &Comment) -> bool {
    first_line(c).starts_with("LEADER-DECISION")
}

/// A `DECISION-APPLIED <decision comment id>` line: the lane has acted on that decision, so it
/// is no longer a task (#2873 N11).
fn decision_applied(c: &Comment, decision: u64) -> bool {
    lines(c).any(|l| {
        l.strip_prefix("DECISION-APPLIED")
            .and_then(|rest| rest.split_whitespace().next())
            .and_then(|id| id.parse::<u64>().ok())
            == Some(decision)
    })
}

/// The body line that marks a tip-repair pull request.
fn is_tip_repair(body: &str) -> bool {
    body.lines()
        .any(|l| l.trim_start().starts_with("moai-tip-repair:"))
}

/// An external pull request (#2664 E1): its author is not the repository's `OWNER`, a
/// `MEMBER` or a `COLLABORATOR`, or it targets another branch than the integration branch.
/// A value that was not read (empty, or no tip branch known) never makes it external.
fn is_external(pr: &Pr, tip_branch: Option<&str>) -> bool {
    let outsider = !pr.author_association.is_empty()
        && !matches!(
            pr.author_association.as_str(),
            "OWNER" | "MEMBER" | "COLLABORATOR"
        );
    let elsewhere = match tip_branch {
        Some(tip) => !pr.base.is_empty() && pr.base != tip,
        None => false,
    };
    outsider || elsewhere
}

// --- decision ----------------------------------------------------------------------------

fn task(priority: &str, kind: &str, target: Value, action: String, evidence: Value) -> Value {
    json!({"priority": priority, "kind": kind, "target": target, "action": action, "evidence": evidence})
}

fn pr_target(pr: &Pr) -> Value {
    json!({"type": "pull_request", "number": pr.number, "head": pr.head})
}

/// The lane's own pull request: what blocks it, if anything (audit FAIL on this head,
/// conflict, failed CI — in that order).
fn own_pr_task(pr: &Pr) -> Option<Value> {
    let fail = pr.comments.iter().rev().find_map(|c| {
        let (head, v) = verdict(c)?;
        (head == pr.head).then_some((c.id, v))
    });
    if let Some((id, "FAIL")) = fail {
        return Some(task(
            "P1",
            "repair_audit_fail",
            pr_target(pr),
            format!("repair the audit findings on PR #{} and push", pr.number),
            json!({"verdict_comment": id}),
        ));
    }
    if pr.conflicting == Some(true) {
        return Some(task(
            "P1",
            "resolve_conflict",
            pr_target(pr),
            format!(
                "merge the tip into PR #{} and resolve the conflict",
                pr.number
            ),
            json!({"mergeable": false}),
        ));
    }
    if pr.ci.as_deref() == Some("red") {
        return Some(task(
            "P1",
            "repair_ci",
            pr_target(pr),
            format!("repair the failed CI on PR #{}", pr.number),
            json!({"ci": pr.ci_detail}),
        ));
    }
    None
}

/// A pull request this lane reserved to audit: whether it owes a verdict now.
fn audit_task(pr: &Pr, lane: &str) -> Option<Value> {
    let (claim_idx, claim_head) = pr.comments.iter().enumerate().rev().find_map(|(i, c)| {
        let (l, h) = audit_claim(c)?;
        (l == lane).then_some((i, h))
    })?;
    let after = &pr.comments[claim_idx + 1..];
    // The latest verdict after the reservation: a re-audit's verdict replaces the first one.
    let mine = after
        .iter()
        .enumerate()
        .rev()
        .find_map(|(i, c)| verdict(c).map(|v| (i, c.id, v)));
    let Some((vi, vid, (vhead, v))) = mine else {
        let kind = if claim_head == pr.head {
            "audit"
        } else {
            "audit_new_head"
        };
        return Some(task(
            "P1",
            kind,
            pr_target(pr),
            format!(
                "audit PR #{} at head {}{}",
                pr.number,
                pr.head,
                if kind == "audit" {
                    ""
                } else {
                    " (post a new moai-audit-claim for this head first)"
                }
            ),
            json!({"audit_claim_head": claim_head}),
        ));
    };
    if vhead != pr.head {
        return Some(task(
            "P1",
            "reaudit_head_moved",
            pr_target(pr),
            format!(
                "delta re-audit PR #{}: your {v} verdict was on {vhead}, the head is now {} (post a new moai-audit-claim for it)",
                pr.number, pr.head
            ),
            json!({"verdict_comment": vid, "verdict_head": vhead}),
        ));
    }
    let request = after[vi + 1..].iter().find(|c| reaudit_request(c));
    request.map(|c| {
        task(
            "P1",
            "reaudit_requested",
            pr_target(pr),
            format!(
                "re-audit PR #{} at head {}: asked in comment {}",
                pr.number, pr.head, c.id
            ),
            json!({"request_comment": c.id, "verdict_comment": vid}),
        )
    })
}

/// The newest audit reservation of a pull request: its comment id and lane.
fn latest_audit_claim(pr: &Pr) -> Option<(u64, String)> {
    pr.comments
        .iter()
        .rev()
        .find_map(|c| audit_claim(c).map(|(lane, _)| (c.id, lane)))
}

/// What a claimed card's thread says after the claim: the newest merged pull request whose
/// body names the claim (#2873 N8), a `pr-wait:` line still waiting, and whether a close table
/// is posted (#2880 W3).
struct CardState<'a> {
    merged: Option<&'a MergedPr>,
    wait: Option<&'a Comment>,
    closed: bool,
}

fn card_state(t: &Thread, claim: u64) -> CardState<'_> {
    let merged = t
        .merged
        .iter()
        .filter(|m| m.claim == Some(claim))
        .max_by(|a, b| a.merged_at.cmp(&b.merged_at));
    let after = t
        .comments
        .iter()
        .position(|c| c.id == claim)
        .map_or(&[][..], |i| &t.comments[i + 1..]);
    // A branch whose pull request merged after the `pr-wait` line is not waiting any more; with
    // a time missing on either side the merge is taken as later, so no second pull request is
    // suggested for work already merged.
    let wait = after
        .iter()
        .rev()
        .find(|c| first_line(c).starts_with("pr-wait:"))
        .filter(|w| {
            merged.is_none_or(|m| !w.at.is_empty() && !m.merged_at.is_empty() && m.merged_at < w.at)
        });
    let closed = after
        .iter()
        .any(|c| first_line(c).starts_with("CLOSE-TABLE"));
    CardState {
        merged,
        wait,
        closed,
    }
}

/// Decides the next task of `lane` from `snap`. Pure: the same snapshot gives the same bytes.
pub fn decide(lane: &str, snap: &Snapshot) -> Value {
    let mut waiting = Vec::new();
    // Pull requests passed over in steps 2 and 4, each with its reason (#2933 S1).
    let mut pr_skipped: Vec<Value> = Vec::new();
    // Cards passed over in steps 6 and 7, each with its reason (#2901), once per Issue.
    let mut skipped: Vec<Value> = Vec::new();
    let mut skip = |entry: Value| {
        if !skipped.iter().any(|s| s["issue"] == entry["issue"]) {
            skipped.push(entry);
        }
    };
    // One judgment for both steps and the end condition (#2901): a `Blocked-By:` reference is
    // open when the open Issue list holds it.
    let open: std::collections::BTreeSet<u64> = snap
        .cards
        .iter()
        .flatten()
        .chain(&snap.open_issues)
        .map(|t| t.number)
        .collect();
    let read_through = read_through(snap);
    let judge =
        |t: &Thread| crate::startable::judge_within(&card(t), &open, read_through, &snap.keep_set);
    let next = (|| {
        // 1. P0: a red tip stops everything else — until a tip-repair pull request is open.
        // Then one lane repairs: its author goes on with it as its own pull request (2.), the
        // others are told it is open and go on down the list (#2873 N2).
        if let Some(tip) = snap.tip.as_ref().filter(|t| t.state == "red") {
            let repairs: Vec<&Pr> = snap.prs.iter().filter(|p| p.tip_repair).collect();
            if repairs.is_empty() {
                return Some(task(
                    "P0",
                    "tip_red",
                    json!({"type": "tip", "head": tip.tip}),
                    "repair the red tip (a PR with a `moai-tip-repair:` line)".into(),
                    json!({"decided_by": tip.decided_by, "detail": tip.detail}),
                ));
            }
            for p in repairs {
                waiting.push(json!({"number": p.number, "reason": "tip_repair_open",
                    "tip": tip.tip, "author_lane": p.author_lane}));
            }
        }
        let mut prs: Vec<&Pr> = snap.prs.iter().collect();
        prs.sort_by_key(|p| p.number);
        // An external pull request is the leader's to report to the user, never lane work
        // (#2664 E1): listed once, then left out of every step below.
        prs.retain(|p| {
            let external = is_external(p, snap.tip_branch.as_deref());
            if external {
                waiting.push(json!({"number": p.number, "reason": "external_pr",
                    "author_association": p.author_association, "base": p.base}));
            }
            !external
        });
        // 2. P1: the lane's own pull requests.
        for pr in prs
            .iter()
            .filter(|p| p.author_lane.as_deref() == Some(lane))
        {
            if let Some(t) = own_pr_task(pr) {
                return Some(t);
            }
            let reason = if pr.draft { "draft" } else { "own_no_blocker" };
            pr_skipped.push(json!({"pr": pr.number, "reason": reason}));
        }
        // 2b. P1: SYNC — a claimed card whose pull request merged and whose close table is not
        // posted yet: close it before new work (#2880 W3). A card with another open pull
        // request of the lane is still in progress.
        let mut claimed: Vec<&Thread> = snap.threads.iter().collect();
        claimed.sort_by_key(|t| t.number);
        for t in claimed {
            let Some(claim) = live_claim(&t.comments, lane) else {
                continue;
            };
            let open_pr = prs
                .iter()
                .any(|p| p.author_lane.as_deref() == Some(lane) && p.claim_issue == Some(t.number));
            let card = card_state(t, claim);
            if let (Some(m), None, false, false) = (card.merged, card.wait, card.closed, open_pr) {
                return Some(task(
                    "P1",
                    "card_pr_merged",
                    json!({"type": "issue", "number": t.number}),
                    format!(
                        "PR #{} for your claim on #{} is merged: post the close table (first line CLOSE-TABLE #{}) and release the claim (moai-claim-release: comment={claim} by={lane}) before new work; if the card's criteria are not all met yet, say so there and continue the card",
                        m.number, t.number, t.number
                    ),
                    json!({"claim_comment": claim, "merged_pr": m.number}),
                ));
            }
        }
        // 3. P1: audits the lane reserved.
        for pr in prs
            .iter()
            .filter(|p| p.author_lane.as_deref() != Some(lane))
        {
            if let Some(t) = audit_task(pr, lane) {
                return Some(t);
            }
        }
        // 4. P1: another lane's pull request with no auditor; a draft or a reserved one is
        // named with its reason and the list goes on.
        for pr in prs
            .iter()
            .filter(|p| p.author_lane.as_deref() != Some(lane))
        {
            if pr.draft {
                pr_skipped.push(json!({"pr": pr.number, "reason": "draft"}));
                continue;
            }
            if let Some((comment, by)) = latest_audit_claim(pr) {
                pr_skipped.push(json!({"pr": pr.number, "reason": "audit_reserved",
                    "by": by, "comment": comment}));
                continue;
            }
            return Some(task(
                "P1",
                "reserve_audit",
                pr_target(pr),
                format!(
                    "reserve the audit of PR #{} (moai-audit-claim: lane={lane} head={}) after checking you wrote no commit in it",
                    pr.number, pr.head
                ),
                json!({"author_lane": pr.author_lane}),
            ));
        }
        // 5. A decision request of the lane that a leader decision now answers.
        let threads = snap
            .threads
            .iter()
            .map(|t| {
                (
                    t.number,
                    &t.comments,
                    live_claim(&t.comments, lane).is_some(),
                )
            })
            .chain(prs.iter().map(|p| {
                (
                    p.number,
                    &p.comments,
                    p.author_lane.as_deref() == Some(lane),
                )
            }));
        for (number, comments, claimed) in threads {
            let Some(bi) = comments
                .iter()
                .rposition(|c| blocked_line(c, lane, claimed).is_some())
            else {
                continue;
            };
            let blocked = &comments[bi];
            let after = &comments[bi + 1..];
            match after.iter().find(|c| leader_decision(c)) {
                Some(d) if after.iter().any(|c| decision_applied(c, d.id)) => {}
                Some(d) => {
                    return Some(task(
                        "P1",
                        "resume_decision",
                        json!({"type": "thread", "number": number}),
                        format!(
                            "resume #{number} with the leader decision in comment {}",
                            d.id
                        ),
                        json!({"blocked_comment": blocked.id, "decision_comment": d.id,
                               "decision": first_line(d)}),
                    ))
                }
                None => waiting.push(json!({"number": number, "blocked_comment": blocked.id,
                    "line": blocked_line(blocked, lane, claimed)})),
            }
        }
        // 6. P2: the next step of a card the lane claims.
        let own_pr_issues: Vec<u64> = prs
            .iter()
            .filter(|p| p.author_lane.as_deref() == Some(lane))
            .filter_map(|p| p.claim_issue)
            .collect();
        let mut threads: Vec<&Thread> = snap.threads.iter().collect();
        threads.sort_by_key(|t| t.number);
        for t in &threads {
            let Some(claim) = live_claim(&t.comments, lane) else {
                continue;
            };
            if own_pr_issues.contains(&t.number) {
                continue;
            }
            let decision_pending = waiting
                .iter()
                .any(|w| w["number"] == json!(t.number) && w.get("blocked_comment").is_some());
            if decision_pending {
                continue;
            }
            // A claimed card on hold, left out by the goal, an epic or blocked is not the next
            // step (#2901 AC1, the #2345 case): it is named and the list goes on.
            if let Some(s) = judge(t) {
                skip(s.entry(t.number));
                continue;
            }
            let card = card_state(t, claim);
            // A pushed branch held back by the open-PR limit (`pr-wait:` after the claim, no pull
            // request of the lane for the card): open it as a draft now instead of waiting for
            // a slot — the merge gate marks a draft ready once it has its PASS (leader decision
            // #1975 6870150891, #2856 audit B2).
            if let Some(w) = card.wait {
                return Some(task(
                    "P2",
                    "open_draft_pr",
                    json!({"type": "issue", "number": t.number}),
                    format!(
                        "open the pushed branch of #{} as a draft pull request (gh pr create --draft); the open-PR limit does not hold a draft",
                        t.number
                    ),
                    json!({"claim_comment": claim, "pr_wait_comment": w.id}),
                ));
            }
            // Merged: step 2b named it, or its close table is posted (the release follows).
            if card.merged.is_some() {
                continue;
            }
            return Some(task(
                "P2",
                "continue_card",
                json!({"type": "issue", "number": t.number}),
                format!("continue your claimed card #{} ({})", t.number, t.title),
                json!({"claim_comment": claim}),
            ));
        }
        // 7. P2–P4: an open Issue that is startable and nobody claims. The open Issues the
        // collection did not read in detail because they are not startable are named first.
        for t in snap.cards.iter().flatten() {
            if let Some(s) = judge(t) {
                skip(s.entry(t.number));
            }
        }
        for t in &snap.open_issues {
            if let Some(s) = judge(t) {
                skip(s.entry(t.number));
                continue;
            }
            if let Some(c) = t.comments.iter().find(|c| claim_session(c).is_some()) {
                skip(json!({"issue": t.number, "reason": "claimed", "claim_comment": c.id}));
                continue;
            }
            return Some(task(
                "P2",
                "pick_issue",
                json!({"type": "issue", "number": t.number}),
                format!(
                    "claim open Issue #{} ({}) after reading its newest comments",
                    t.number, t.title
                ),
                json!({"labels": t.labels}),
            ));
        }
        None
    })();
    // The lane-wide end (#2901, #2515): no open pull request and no startable open Issue, a
    // claimed one included — a claimed card is still work. Unknown while the list was not read
    // or read in part.
    let startable = snap.cards.as_ref().and_then(|cards| {
        let n = cards.iter().filter(|t| judge(t).is_none()).count() as u64;
        (n > 0 || !snap.cards_truncated).then_some(n)
    });
    let open_prs = (snap.cards.is_some() || !snap.prs.is_empty()).then_some(snap.prs.len() as u64);
    let mut v = json!({
        "lane": lane,
        "next": next,
        "waiting": waiting,
        "skipped": skipped,
        "pr_skipped": pr_skipped,
        "end_condition": crate::startable::end_condition(open_prs, startable),
        "partial": snap.partial,
        "api_calls": snap.api_calls,
        "writes": false,
    });
    if let Some(now) = &snap.now {
        v["now"] = json!(now);
    }
    v
}

/// `--explain` (#2933 S1): one line for each pull request passed over, one for each Issue
/// skipped, and a last line for the end condition.
pub fn explain_lines(v: &Value) -> Vec<String> {
    let none = Vec::new();
    let prs = v["pr_skipped"].as_array().unwrap_or(&none).iter().map(|p| {
        let by = match (p["by"].as_str(), p["comment"].as_u64()) {
            (Some(by), Some(c)) => format!(" by {by} (comment {c})"),
            _ => String::new(),
        };
        format!(
            "PR #{} passed over: {}{by}",
            p["pr"],
            p["reason"].as_str().unwrap_or("?")
        )
    });
    let issues = v["skipped"].as_array().unwrap_or(&none).iter().map(|s| {
        format!(
            "Issue #{} skipped: {}",
            s["issue"],
            s["reason"].as_str().unwrap_or("?")
        )
    });
    let end = &v["end_condition"];
    prs.chain(issues)
        .chain(std::iter::once(format!(
            "end_condition: open_prs={} startable={} met={}",
            end["open_prs"], end["startable"], end["met"]
        )))
        .collect()
}

/// The one-line summary for a terminal (D2): stderr only. It carries no text read from GitHub
/// (titles, comment lines) — only fixed words, numbers and a hex head — so nothing a comment
/// says can reach the terminal as a control sequence (#2799 6867586011).
pub fn summary_line(v: &Value) -> String {
    match v["next"].as_object() {
        Some(n) => {
            let number = n["target"]["number"]
                .as_u64()
                .map_or(String::new(), |x| format!(" #{x}"));
            let head = n["target"]["head"]
                .as_str()
                .filter(|h| is_hex40(h))
                .map_or(String::new(), |h| format!(" head {}", &h[..10]));
            format!(
                "lane-next: {} {}{number}{head}",
                n["priority"].as_str().unwrap_or("?"),
                n["kind"].as_str().unwrap_or("?"),
            )
        }
        None => format!(
            "lane-next: nothing to do now ({} waiting, {} skipped, {} PRs passed)",
            v["waiting"].as_array().map_or(0, Vec::len),
            v["skipped"].as_array().map_or(0, Vec::len),
            v["pr_skipped"].as_array().map_or(0, Vec::len)
        ),
    }
}

// --- collection --------------------------------------------------------------------------

/// The repository and lane a collection reads for.
pub struct Target<'a> {
    pub owner: &'a str,
    pub repo: &'a str,
    pub lane: &'a str,
    /// The integration branch whose CI is P0; `None` takes the base most open pull requests
    /// target (the integration branch while one is in use).
    pub tip_branch: Option<&'a str>,
    /// The local clone `git rev-list` walks for the tip judgment.
    pub dir: &'a Path,
}

/// Sends at most [`API_BUDGET`] requests; the next one is refused with `budget`.
struct Budgeted<'t, T: Transport> {
    inner: &'t T,
    used: Cell<u32>,
}

impl<T: Transport> Transport for Budgeted<'_, T> {
    fn send(&self, req: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
        if self.used.get() >= API_BUDGET {
            return Err("budget".into());
        }
        self.used.set(self.used.get() + 1);
        self.inner.send(req)
    }
}

/// GitHub's GraphQL endpoint: one request reads many lists, and it is limited by its own
/// points per hour, apart from the REST requests (GitHub Docs, "Rate limits and query limits
/// for the GraphQL API").
const GRAPHQL_URL: &str = "https://api.github.com/graphql";
/// Issues read in one GraphQL request.
pub const ISSUES_PER_QUERY: usize = 25;
/// Newest comments read per thread (GraphQL's largest page).
const COMMENTS_READ: u64 = 100;

/// What the startable judgment reads of an Issue (#2901): body, labels, open sub-issues and
/// the Issues GitHub reports as blocking it.
macro_rules! card_fields {
    () => {
        "number title body labels(first:20){nodes{name}} \
subIssuesSummary{total completed} blockedBy(first:20){nodes{number state}}"
    };
}

/// Open pull requests with their newest comments, the lane's two searches and the oldest
/// hundred open Issues with what the startable judgment reads (#2901), in one request.
const LISTS_QUERY: &str = concat!(
    "query($owner:String!,$repo:String!,$claims:String!,$blocked:String!){\
repository(owner:$owner,name:$repo){\
pullRequests(states:OPEN,first:100,orderBy:{field:CREATED_AT,direction:ASC}){totalCount nodes{\
number title body headRefOid isDraft mergeable baseRefName authorAssociation \
comments(last:100){totalCount nodes{databaseId body createdAt}}}}\
issues(states:OPEN,first:100,orderBy:{field:CREATED_AT,direction:ASC}){totalCount nodes{",
    card_fields!(),
    "}}}\
claims:search(query:$claims,type:ISSUE,first:30){nodes{...on Issue{number} ...on PullRequest{number}}}\
blocked:search(query:$blocked,type:ISSUE,first:30){nodes{...on Issue{number} ...on PullRequest{number}}}}"
);

/// One Issue's startable fields, newest comments and the merged pull requests that refer to
/// it.
fn issue_fields(n: u64) -> String {
    format!(
        "i{n}:issue(number:{n}){{{} \
comments(last:{COMMENTS_READ}){{totalCount nodes{{databaseId body createdAt}}}} \
timelineItems(itemTypes:[CROSS_REFERENCED_EVENT],last:50){{nodes{{...on CrossReferencedEvent{{\
source{{...on PullRequest{{number merged mergedAt body}}}}}}}}}}}}",
        card_fields!()
    )
}

/// Sends one GraphQL query; the `data` and, when GitHub also answered errors next to the data,
/// their messages (a missing Issue, for example).
fn graphql<T: Transport>(t: &T, query: &str, variables: Value) -> Result<(Value, Option<String>)> {
    let req = HttpRequest::new("POST", GRAPHQL_URL)
        .with_body(json!({"query": query, "variables": variables}));
    let resp = t.send(&req)?;
    if resp.status != 200 {
        return Err(format!("GraphQL returned HTTP {}", resp.status));
    }
    let body = resp.body.unwrap_or(Value::Null);
    let errors = body["errors"]
        .as_array()
        .filter(|e| !e.is_empty())
        .map(|e| {
            e.iter()
                .map(|x| x["message"].as_str().unwrap_or("error"))
                .collect::<Vec<_>>()
                .join("; ")
        });
    if !body["data"].is_object() {
        return Err(errors.unwrap_or_else(|| "GraphQL answered without data".into()));
    }
    Ok((body["data"].clone(), errors))
}

/// The comments of a GraphQL connection, oldest first, and whether older ones were left out.
fn gql_comments(conn: &Value) -> (Vec<Comment>, bool) {
    let list: Vec<Comment> = conn["nodes"]
        .as_array()
        .map(|a| {
            a.iter()
                .map(|c| Comment {
                    id: c["databaseId"].as_u64().unwrap_or(0),
                    body: c["body"].as_str().unwrap_or("").to_string(),
                    at: c["createdAt"].as_str().unwrap_or("").to_string(),
                })
                .collect()
        })
        .unwrap_or_default();
    let more = conn["totalCount"].as_u64().unwrap_or(0) > list.len() as u64;
    (list, more)
}

/// An Issue read by [`issue_fields`] as a thread.
fn gql_thread(i: &Value) -> (Thread, bool) {
    let (comments, more) = gql_comments(&i["comments"]);
    let merged = i["timelineItems"]["nodes"]
        .as_array()
        .map(|a| {
            a.iter()
                .map(|e| &e["source"])
                .filter(|pr| pr["merged"].as_bool() == Some(true))
                .filter_map(|pr| {
                    Some(MergedPr {
                        number: pr["number"].as_u64()?,
                        claim: body_claim(pr["body"].as_str().unwrap_or("")).map(|(_, c)| c),
                        merged_at: pr["mergedAt"].as_str().unwrap_or("").to_string(),
                    })
                })
                .collect()
        })
        .unwrap_or_default();
    let labels = i["labels"]["nodes"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|l| l["name"].as_str().map(str::to_string))
                .collect()
        })
        .unwrap_or_default();
    let children = &i["subIssuesSummary"];
    let blocked_by = i["blockedBy"]["nodes"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter(|b| b["state"].as_str() == Some("OPEN"))
                .filter_map(|b| b["number"].as_u64())
                .collect()
        })
        .unwrap_or_default();
    (
        Thread {
            number: i["number"].as_u64().unwrap_or(0),
            title: i["title"].as_str().unwrap_or("").to_string(),
            labels,
            comments,
            merged,
            body: i["body"].as_str().unwrap_or("").to_string(),
            open_children: children["total"]
                .as_u64()
                .unwrap_or(0)
                .saturating_sub(children["completed"].as_u64().unwrap_or(0)),
            blocked_by,
        },
        more,
    )
}

/// `Moai-Claim: <issue>#<comment>` in a body.
/// A bare line wins; without one, the first inline `Moai-Claim: <digits>#<digits>` (in a list
/// item or code span) counts, the way the merge gate's G11 reads the marker anywhere in the
/// body (`gate_collect::has_body_claim`; WF-OBS 6874554271, #2401).
fn body_claim(body: &str) -> Option<(u64, u64)> {
    let bare = body.lines().find_map(|l| {
        let rest = l.trim().strip_prefix("Moai-Claim:")?.trim();
        let (issue, comment) = rest.split_once('#')?;
        Some((issue.trim().parse().ok()?, comment.trim().parse().ok()?))
    });
    bare.or_else(|| {
        body.match_indices("Moai-Claim:").find_map(|(at, m)| {
            let rest = body[at + m.len()..].trim_start();
            let digits = |s: &str| s.find(|c: char| !c.is_ascii_digit()).unwrap_or(s.len());
            let n = digits(rest);
            let tail = rest[n..].strip_prefix('#')?;
            let k = digits(tail);
            Some((rest[..n].parse().ok()?, tail[..k].parse().ok()?))
        })
    })
}

/// The tip's state as #2831 states P0 (D8(a)) — the merge gate's own judgment, one copy
/// (#2856 D3, merged as `gate_collect::d8a_judgment` by #2869).
pub use crate::gate_collect::d8a_judgment;

/// P0 input: the tip commit judged by [`d8a_judgment`]; while it reads `pending` (running, or
/// only cancelled), the earlier first-parent tip commits are read in turn, at most
/// `gate_collect::TIP_WALK`, and the first with a completed judgment decides. When the local
/// clone lacks the newest tip (no fetch here: reads only), the tip commit alone is judged and
/// the returned note says so.
fn tip_d8a<T: Transport>(
    t: &T,
    base: &str,
    dir: &Path,
    tip: &str,
) -> Result<(Tip, Option<String>)> {
    let walk = crate::audit_delta::command(dir)
        .args([
            "rev-list",
            "--first-parent",
            &format!("--max-count={}", crate::gate_collect::TIP_WALK),
            tip,
        ])
        .output()
        .ok()
        .filter(|o| o.status.success())
        .map(|o| String::from_utf8_lossy(&o.stdout).into_owned());
    let (shas, note): (Vec<String>, Option<String>) = match &walk {
        Some(out) => (out.lines().map(str::to_string).collect(), None),
        None => (
            vec![tip.to_string()],
            Some(format!(
                "tip CI: {tip} is not in the local clone; judged the tip commit alone, without earlier commits"
            )),
        ),
    };
    for sha in &shas {
        let j = d8a_judgment(&crate::gate_collect::tip_runs(t, base, sha)?);
        let state = j["state"].as_str().unwrap_or("unknown").to_string();
        if state == "pending" {
            continue;
        }
        return Ok((
            Tip {
                state,
                tip: tip.to_string(),
                decided_by: Some(sha.clone()),
                detail: j.get("failed").map(Value::to_string).unwrap_or_default(),
            },
            note,
        ));
    }
    Ok((
        Tip {
            state: "unknown".into(),
            tip: tip.to_string(),
            decided_by: None,
            detail: format!("no completed job on the last {} tip commit(s)", shas.len()),
        },
        note,
    ))
}

/// Reads one snapshot for `target`, higher priorities first, within [`API_BUDGET`] requests:
/// one GraphQL request for the lists, the tip's CI (REST), one GraphQL request per
/// [`ISSUES_PER_QUERY`] Issues, then the CI of the lane's own pull requests (REST).
pub fn collect<T: Transport>(target: &Target<'_>, transport: &T) -> Snapshot {
    let t = Budgeted {
        inner: transport,
        used: Cell::new(0),
    };
    let base = format!(
        "https://api.github.com/repos/{}/{}",
        target.owner, target.repo
    );
    let mut snap = Snapshot::default();
    let miss = |what: String, e: String, snap: &mut Snapshot| {
        snap.partial.push(if e == "budget" {
            format!("{what}: request budget ({API_BUDGET}) spent")
        } else {
            format!("{what}: {e}")
        });
    };
    // P1 and 5–7: the lists, in one request.
    let search = |text: String| {
        format!(
            "repo:{}/{} is:open in:comments \"{text}\"",
            target.owner, target.repo
        )
    };
    let vars = json!({"owner": target.owner, "repo": target.repo,
        "claims": search(format!("session={}", target.lane)),
        "blocked": search(format!("LANE-BLOCKED {}", target.lane))});
    let lists = match graphql(&t, LISTS_QUERY, vars) {
        Ok((data, errors)) => {
            if let Some(e) = errors {
                snap.partial.push(format!("lists: {e}"));
            }
            data
        }
        Err(e) => {
            miss("open pull requests and Issues".into(), e, &mut snap);
            Value::Null
        }
    };
    let repo = &lists["repository"];
    let pulls = repo["pullRequests"]["nodes"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    if repo["pullRequests"]["totalCount"].as_u64().unwrap_or(0) > pulls.len() as u64 {
        snap.partial.push(format!(
            "only the oldest {} open pull requests read",
            pulls.len()
        ));
    }
    // P0: the tip — the named branch, else the base most open pull requests target.
    let mut bases: BTreeMap<&str, usize> = BTreeMap::new();
    for p in &pulls {
        if let Some(b) = p["baseRefName"].as_str() {
            *bases.entry(b).or_default() += 1;
        }
    }
    let tip_branch = target.tip_branch.map(str::to_string).or_else(|| {
        bases
            .iter()
            .max_by_key(|(_, n)| **n)
            .map(|(b, _)| (*b).to_string())
    });
    snap.tip_branch = tip_branch.clone();
    match tip_branch {
        None => snap
            .partial
            .push("tip: no --tip and no open pull request to take the base from".into()),
        Some(tip_branch) => {
            match crate::gate_collect::get_json(&t, &format!("{base}/branches/{tip_branch}")) {
                Ok((b, _)) => {
                    let sha = b["commit"]["sha"].as_str().unwrap_or("").to_string();
                    match tip_d8a(&t, &base, target.dir, &sha) {
                        Ok((tip, note)) => {
                            snap.tip = Some(tip);
                            if let Some(n) = note {
                                snap.partial.push(n);
                            }
                        }
                        Err(e) => miss("tip CI".into(), e, &mut snap),
                    }
                }
                Err(e) => miss("tip branch".into(), e, &mut snap),
            }
        }
    }
    let mut claims: Vec<(u64, u64)> = Vec::new();
    for p in &pulls {
        let number = p["number"].as_u64().unwrap_or(0);
        let body = p["body"].as_str().unwrap_or("");
        let (comments, more) = gql_comments(&p["comments"]);
        if more {
            snap.partial.push(format!(
                "PR #{number}: only the newest {COMMENTS_READ} comments read"
            ));
        }
        let claim = body_claim(body);
        if let Some(c) = claim {
            claims.push(c);
        }
        snap.prs.push(Pr {
            number,
            title: p["title"].as_str().unwrap_or("").to_string(),
            head: p["headRefOid"].as_str().unwrap_or("").to_string(),
            draft: p["isDraft"].as_bool().unwrap_or(false),
            base: p["baseRefName"].as_str().unwrap_or("").to_string(),
            author_association: p["authorAssociation"].as_str().unwrap_or("").to_string(),
            claim_issue: claim.map(|(i, _)| i),
            conflicting: match p["mergeable"].as_str() {
                Some("CONFLICTING") => Some(true),
                Some("MERGEABLE") => Some(false),
                _ => None,
            },
            tip_repair: is_tip_repair(body),
            comments,
            ..Pr::default()
        });
    }
    // 5–7: the Issues to read — the claim Issues of the pull requests (for their authors), the
    // lane's threads and the oldest open Issues.
    let pr_numbers: Vec<u64> = snap.prs.iter().map(|p| p.number).collect();
    let mut own: Vec<u64> = ["claims", "blocked"]
        .iter()
        .flat_map(|k| lists[*k]["nodes"].as_array().cloned().unwrap_or_default())
        .filter_map(|n| n["number"].as_u64())
        .filter(|n| !pr_numbers.contains(n))
        .collect();
    own.sort_unstable();
    own.dedup();
    // #2901: every open Issue of the list is judged with what the list already carries (no
    // request more), and only the oldest startable ones are read in detail — five held ones
    // at the front no longer hide the sixth.
    if let Some(nodes) = repo["issues"]["nodes"].as_array() {
        let cards: Vec<Thread> = nodes.iter().map(|i| gql_thread(i).0).collect();
        if repo["issues"]["totalCount"].as_u64().unwrap_or(0) > cards.len() as u64 {
            snap.cards_truncated = true;
            snap.partial
                .push(format!("only the oldest {} open Issues read", cards.len()));
        }
        snap.cards = Some(cards);
    }
    // The active goal's exclusions, from the local clone (no request).
    match crate::goal::show(target.dir) {
        Ok(g) if g["state"] == "active" => {
            match serde_json::from_value(g["goal"]["exclusions"].clone()) {
                Ok(keep) => snap.keep_set = keep,
                Err(e) => snap.partial.push(format!("goal exclusions: {e}")),
            }
        }
        Ok(_) => {}
        Err(e) => snap.partial.push(format!("goal exclusions: {e}")),
    }
    let open: std::collections::BTreeSet<u64> =
        snap.cards.iter().flatten().map(|t| t.number).collect();
    let through = read_through(&snap);
    let free: Vec<u64> = snap
        .cards
        .iter()
        .flatten()
        .filter(|t| {
            crate::startable::judge_within(&card(t), &open, through, &snap.keep_set).is_none()
        })
        .map(|t| t.number)
        .take(FREE_ISSUE_SCAN)
        .collect();
    let mut wanted: Vec<u64> = claims.iter().map(|(i, _)| *i).collect();
    wanted.extend(&own);
    wanted.extend(&free);
    let mut seen = std::collections::BTreeSet::new();
    wanted.retain(|n| seen.insert(*n));
    let mut issues: BTreeMap<u64, (Thread, bool)> = BTreeMap::new();
    for chunk in wanted.chunks(ISSUES_PER_QUERY) {
        let fields: String = chunk.iter().map(|n| issue_fields(*n)).collect();
        let query = format!(
            "query($owner:String!,$repo:String!){{repository(owner:$owner,name:$repo){{{fields}}}}}"
        );
        match graphql(
            &t,
            &query,
            json!({"owner": target.owner, "repo": target.repo}),
        ) {
            Ok((data, errors)) => {
                if let Some(e) = errors {
                    snap.partial.push(format!("Issues: {e}"));
                }
                for n in chunk {
                    let i = &data["repository"][format!("i{n}")];
                    if i.is_object() {
                        issues.insert(*n, gql_thread(i));
                    }
                }
            }
            Err(e) => miss(
                format!(
                    "Issues {}",
                    chunk
                        .iter()
                        .map(|n| format!("#{n}"))
                        .collect::<Vec<_>>()
                        .join(" ")
                ),
                e,
                &mut snap,
            ),
        }
    }
    // Who wrote each pull request: the lane of the claim its body names, read in the claim
    // Issue; a claim older than the comments read there is read by itself (REST).
    let mut claim_lane: BTreeMap<u64, Option<String>> = BTreeMap::new();
    for (idx, p) in pulls.iter().enumerate() {
        let number = snap.prs[idx].number;
        let Some((issue, comment)) = body_claim(p["body"].as_str().unwrap_or("")) else {
            continue;
        };
        if let std::collections::btree_map::Entry::Vacant(slot) = claim_lane.entry(comment) {
            let read = issues
                .get(&issue)
                .and_then(|(th, _)| th.comments.iter().find(|c| c.id == comment))
                .map(claim_session);
            let lane = match read {
                Some(lane) => lane,
                None => match crate::gate_collect::get_json(
                    &t,
                    &format!("{base}/issues/comments/{comment}"),
                ) {
                    Ok((c, _)) => claim_session(&Comment {
                        id: comment,
                        body: c["body"].as_str().unwrap_or("").to_string(),
                        ..Comment::default()
                    }),
                    Err(e) => {
                        miss(
                            format!("claim comment {comment} of PR #{number}"),
                            e,
                            &mut snap,
                        );
                        None
                    }
                },
            };
            slot.insert(lane);
        }
        snap.prs[idx].author_lane = claim_lane[&comment].clone();
    }
    let take = |n: u64, snap: &mut Snapshot| -> Option<Thread> {
        match issues.get(&n) {
            Some((th, more)) => {
                let note = format!("#{n}: only the newest {COMMENTS_READ} comments read");
                if *more && !snap.partial.contains(&note) {
                    snap.partial.push(note);
                }
                Some(th.clone())
            }
            None => {
                if !snap.partial.iter().any(|p| p.contains(&format!("#{n}"))) {
                    snap.partial.push(format!("#{n}: not read"));
                }
                None
            }
        }
    };
    for n in own {
        if let Some(th) = take(n, &mut snap) {
            snap.threads.push(th);
        }
    }
    for n in free {
        if let Some(th) = take(n, &mut snap) {
            snap.open_issues.push(th);
        }
    }
    // The CI of the lane's own pull requests.
    for pr in snap.prs.iter_mut() {
        if pr.author_lane.as_deref() != Some(target.lane) {
            continue;
        }
        match crate::gate_collect::tip_runs(&t, &base, &pr.head) {
            Ok(runs) => {
                let j = crate::gate_collect::completed_judgment(&runs);
                pr.ci = j["state"].as_str().map(str::to_string);
                pr.ci_detail = j.get("failed").map(Value::to_string);
            }
            Err(e) => snap.partial.push(if e == "budget" {
                format!(
                    "CI of PR #{}: request budget ({API_BUDGET}) spent",
                    pr.number
                )
            } else {
                format!("CI of PR #{}: {e}", pr.number)
            }),
        }
    }
    snap.api_calls = t.used.get();
    snap
}

/// Collects and decides; with `wait_secs` > 0, collects again every [`poll_secs`]
/// (`shared_lanes`) seconds until a task appears or the wait is over (at most
/// [`MAX_WAIT_SECS`]).
pub fn run<T: Transport, C: Clock>(
    target: &Target<'_>,
    transport: &T,
    clock: &C,
    wait_secs: u64,
    shared_lanes: u64,
) -> Result<Value> {
    if wait_secs > MAX_WAIT_SECS {
        return Err(format!("--wait accepts at most {MAX_WAIT_SECS} seconds"));
    }
    if !(1..=MAX_SHARED_LANES).contains(&shared_lanes) {
        return Err(format!(
            "--shared-lanes accepts 1 to {MAX_SHARED_LANES} lanes"
        ));
    }
    let poll = poll_secs(shared_lanes);
    let deadline = clock.now() + wait_secs as i64;
    let mut rounds = 0u32;
    loop {
        rounds += 1;
        let mut snap = collect(target, transport);
        snap.now
            .get_or_insert_with(|| crate::forge::policy::format_utc(clock.now()));
        let mut v = decide(target.lane, &snap);
        let done = !v["next"].is_null() || clock.now() + poll as i64 > deadline;
        if done {
            v["rounds"] = json!(rounds);
            v["poll_secs"] = json!(poll);
            return Ok(v);
        }
        clock.sleep(poll);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn c(id: u64, body: &str) -> Comment {
        Comment {
            id,
            body: body.into(),
            ..Comment::default()
        }
    }

    /// GitHub as GraphQL and REST answer it: `prs` open pull requests, each naming claim
    /// `n` on Issue `1000 + n`; Issue 3000 found by the lane's search, with a merged and an
    /// unmerged pull request referring to it; ten open Issues. With `claims_in_issue` false
    /// the claim comments are older than the comments an Issue read returns, so each is read
    /// by itself. Counts requests.
    struct Hub {
        prs: u64,
        claims_in_issue: bool,
        sent: Cell<u32>,
    }

    impl Transport for Hub {
        fn send(&self, req: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
            self.sent.set(self.sent.get() + 1);
            let ok = |b: Value| Ok(crate::forge::transport::HttpResponse::new(200).with_body(b));
            if req.url == GRAPHQL_URL {
                assert_eq!(req.method, "POST");
                let q = req.body.as_ref().unwrap()["query"]
                    .as_str()
                    .unwrap()
                    .to_string();
                if q.contains("pullRequests") {
                    let prs: Vec<Value> = (1..=self.prs)
                        .map(|n| {
                            let tip = if n == 1 { "\nmoai-tip-repair: tip red" } else { "" };
                            json!({"number": n, "title": format!("pr {n}"),
                                "body": format!("Moai-Claim: {}#{n}{tip}", 1000 + n),
                                "headRefOid": "a".repeat(40), "isDraft": false,
                                "mergeable": if n == 2 { "CONFLICTING" } else { "MERGEABLE" },
                                "baseRefName": "main",
                                "authorAssociation": if n == 2 { "CONTRIBUTOR" } else { "MEMBER" },
                                "comments": {"totalCount": 1, "nodes": [
                                    {"databaseId": 5000 + n, "body": "note", "createdAt": "2026-11-14T00:00:00Z"}]}})
                        })
                        .collect();
                    return ok(json!({"data": {
                        "repository": {"pullRequests": {"totalCount": self.prs, "nodes": prs},
                            "issues": {"nodes": (0..10).map(|k| json!({"number": 2000 + k})).collect::<Vec<_>>()}},
                        "claims": {"nodes": [{"number": 3000}, {"number": 1}]},
                        "blocked": {"nodes": []}}}));
                }
                let mut repo = serde_json::Map::new();
                for part in q.split(":issue(number:").skip(1) {
                    let digits: String = part.chars().take_while(char::is_ascii_digit).collect();
                    let n: u64 = digits.parse().unwrap();
                    assert!(q.contains(&format!("i{n}:issue(number:{n})")));
                    let comments = if n == 3000 {
                        vec![
                            json!({"databaseId": 77, "body": "moai-claim: session=agent-10 epoch=1"}),
                        ]
                    } else if (1001..2000).contains(&n) && self.claims_in_issue {
                        vec![
                            json!({"databaseId": n - 1000, "body": "moai-claim: session=agent-x epoch=1"}),
                        ]
                    } else {
                        vec![]
                    };
                    let timeline = if n == 3000 {
                        json!([
                            {"source": {"number": 90, "merged": true, "mergedAt": "2026-11-14T03:00:00Z",
                                "body": "x\nMoai-Claim: 3000#77"}},
                            {"source": {"number": 91, "merged": false, "body": "Moai-Claim: 3000#77"}},
                            {}
                        ])
                    } else {
                        json!([])
                    };
                    repo.insert(
                        format!("i{n}"),
                        json!({"number": n, "title": format!("issue {n}"),
                        "labels": {"nodes": [{"name": "P3"}]},
                        "comments": {"totalCount": comments.len(), "nodes": comments},
                        "timelineItems": {"nodes": timeline}}),
                    );
                }
                return ok(json!({"data": {"repository": repo}}));
            }
            assert_eq!(req.method, "GET");
            if req.url.contains("/branches/") {
                ok(json!({"commit": {"sha": "b".repeat(40)}}))
            } else if req.url.contains("/issues/comments/") {
                ok(json!({"body": "moai-claim: session=agent-x epoch=1"}))
            } else {
                ok(json!({"check_runs": []}))
            }
        }
    }

    fn target(dir: &Path) -> Target<'_> {
        Target {
            owner: "o",
            repo: "r",
            lane: "agent-10",
            tip_branch: None,
            dir,
        }
    }

    /// #2873 N1: forty open pull requests with their comments, their authors, the lane's thread
    /// and the open Issues come from a fixed handful of requests — nothing left unread.
    #[test]
    fn forty_pull_requests_are_read_in_a_few_requests() {
        let dir = tempfile::TempDir::new().unwrap();
        for prs in [1u64, 40] {
            let t = Hub {
                prs,
                claims_in_issue: true,
                sent: Cell::new(0),
            };
            let snap = collect(&target(dir.path()), &t);
            // lists + tip branch + tip check runs + Issue reads (ISSUES_PER_QUERY each).
            let issue_reads = (prs as usize + 1 + FREE_ISSUE_SCAN).div_ceil(ISSUES_PER_QUERY);
            assert_eq!(snap.api_calls, 3 + issue_reads as u32, "{:?}", snap.partial);
            assert_eq!(t.sent.get(), snap.api_calls);
            assert!(
                !snap
                    .partial
                    .iter()
                    .any(|p| p.contains("budget") || p.contains("not read")),
                "{:?}",
                snap.partial
            );
            assert_eq!(snap.prs.len() as u64, prs);
            assert!(snap
                .prs
                .iter()
                .all(|p| p.author_lane.as_deref() == Some("agent-x") && p.comments.len() == 1));
            assert_eq!(snap.open_issues.len(), FREE_ISSUE_SCAN);
            assert_eq!(snap.open_issues[0].labels, vec!["P3".to_string()]);
            // The search hit that is an open pull request is not read again as a thread.
            assert_eq!(
                snap.threads.iter().map(|t| t.number).collect::<Vec<_>>(),
                vec![3000]
            );
            // N8: only the merged pull request, with the claim its body names.
            assert_eq!(
                snap.threads[0].merged,
                vec![MergedPr {
                    number: 90,
                    claim: Some(77),
                    merged_at: "2026-11-14T03:00:00Z".into()
                }]
            );
            // N2: the tip-repair line and the conflict come from the same read.
            assert!(snap.prs[0].tip_repair);
            if prs > 1 {
                assert_eq!(snap.prs[1].conflicting, Some(true));
                assert!(!snap.prs[1].tip_repair);
            }
            assert_eq!(snap.prs[0].conflicting, Some(false));
            // #2664 E1: the base, the author association and the tip branch come from the
            // same read.
            assert_eq!(snap.tip_branch.as_deref(), Some("main"));
            assert_eq!(snap.prs[0].base, "main");
            assert_eq!(snap.prs[0].author_association, "MEMBER");
            if prs > 1 {
                assert_eq!(snap.prs[1].author_association, "CONTRIBUTOR");
            }
        }
        assert!(LISTS_QUERY.contains(" authorAssociation "));
        assert_eq!(ISSUES_PER_QUERY, 25);
    }

    /// D4: one collection still sends at most API_BUDGET requests — claims older than the
    /// comments read are read one by one until the budget is spent — and names what it left
    /// unread instead of guessing.
    #[test]
    fn a_collection_stops_at_its_request_budget() {
        let t = Hub {
            prs: 40,
            claims_in_issue: false,
            sent: Cell::new(0),
        };
        let dir = tempfile::TempDir::new().unwrap();
        let snap = collect(&target(dir.path()), &t);
        assert_eq!(t.sent.get(), API_BUDGET, "{snap:?}");
        assert_eq!(snap.api_calls, API_BUDGET);
        assert!(
            snap.partial
                .iter()
                .any(|p| p.contains("request budget (30) spent")),
            "{:?}",
            snap.partial
        );
        // The tip and the pull request list came first.
        assert!(snap.tip.is_some());
        assert_eq!(snap.prs.len(), 40);
        assert_eq!(API_BUDGET, 30, "the leader's number (#2831 D4)");
    }

    /// No open pull request; `held` of the `issues` oldest open Issues carry `HOLD`, the rest
    /// none; `total` open Issues exist. Records the Issues each detail read asks for.
    struct HeldOldest {
        issues: u64,
        held: u64,
        total: u64,
        sent: Cell<u32>,
        read: std::cell::RefCell<Vec<u64>>,
    }

    impl Transport for HeldOldest {
        fn send(&self, req: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
            self.sent.set(self.sent.get() + 1);
            let ok = |b: Value| Ok(crate::forge::transport::HttpResponse::new(200).with_body(b));
            assert_eq!(
                req.url, GRAPHQL_URL,
                "no REST request without a pull request"
            );
            let q = req.body.as_ref().unwrap()["query"].as_str().unwrap();
            let labels = |n: u64| {
                if n - 2000 <= self.held {
                    json!({"nodes": [{"name": "HOLD"}]})
                } else {
                    json!({"nodes": []})
                }
            };
            if q.contains("pullRequests") {
                let nodes: Vec<Value> = (2001..=2000 + self.issues)
                    .map(|n| {
                        json!({"number": n, "title": format!("issue {n}"), "body": "",
                        "labels": labels(n)})
                    })
                    .collect();
                return ok(json!({"data": {
                    "repository": {"pullRequests": {"totalCount": 0, "nodes": []},
                        "issues": {"totalCount": self.total, "nodes": nodes}},
                    "claims": {"nodes": []}, "blocked": {"nodes": []}}}));
            }
            let mut repo = serde_json::Map::new();
            for part in q.split(":issue(number:").skip(1) {
                let digits: String = part.chars().take_while(char::is_ascii_digit).collect();
                let n: u64 = digits.parse().unwrap();
                self.read.borrow_mut().push(n);
                repo.insert(
                    format!("i{n}"),
                    json!({"number": n, "title": format!("issue {n}"), "labels": labels(n),
                        "comments": {"totalCount": 0, "nodes": []}, "timelineItems": {"nodes": []}}),
                );
            }
            ok(json!({"data": {"repository": repo}}))
        }
    }

    /// #2901 AC3/AC4: the five oldest open Issues all on hold do not hide the sixth — the
    /// comments read are the oldest *startable* ones, the held ones are named under `skipped`,
    /// and the request count stays the same (one list request, one Issue read), held or not.
    #[test]
    fn held_oldest_issues_do_not_hide_a_free_one() {
        let dir = tempfile::TempDir::new().unwrap();
        for held in [5u64, 0] {
            let t = HeldOldest {
                issues: 6,
                held,
                total: 6,
                sent: Cell::new(0),
                read: Default::default(),
            };
            let snap = collect(&target(dir.path()), &t);
            assert_eq!(snap.api_calls, 2, "held {held}: {:?}", snap.partial);
            assert_eq!(t.sent.get(), 2);
            let want: Vec<u64> = (2001 + held..=2006).take(FREE_ISSUE_SCAN).collect();
            assert_eq!(*t.read.borrow(), want, "held {held}");
            let v = decide("agent-10", &snap);
            assert_eq!(v["next"]["kind"], json!("pick_issue"), "{v}");
            assert_eq!(v["next"]["target"]["number"], json!(2001 + held));
            let skipped: Vec<u64> = v["skipped"]
                .as_array()
                .unwrap()
                .iter()
                .map(|s| s["issue"].as_u64().unwrap())
                .collect();
            assert_eq!(skipped, (2001..2001 + held).collect::<Vec<_>>(), "{v}");
            assert_eq!(
                v["end_condition"],
                json!({"open_prs": 0, "startable": 6 - held, "met": false})
            );
        }
        // Every open Issue on hold: the loop's end condition is met — unless more open Issues
        // exist than the list read, which leaves it open.
        for (total, met) in [(6u64, json!(true)), (7, Value::Null)] {
            let t = HeldOldest {
                issues: 6,
                held: 6,
                total,
                sent: Cell::new(0),
                read: Default::default(),
            };
            let snap = collect(&target(dir.path()), &t);
            assert_eq!(
                snap.api_calls, 1,
                "nothing startable, nothing read in detail"
            );
            let v = decide("agent-10", &snap);
            assert_eq!(v["next"], Value::Null, "{v}");
            assert_eq!(v["end_condition"]["met"], met, "{v}");
        }
    }

    /// Answers every GET with an empty list: never a task.
    struct Quiet;
    impl Transport for Quiet {
        fn send(&self, _: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
            Ok(crate::forge::transport::HttpResponse::new(200).with_body(json!([])))
        }
    }

    struct FakeClock {
        now: Cell<i64>,
        slept: Cell<u64>,
    }
    impl Clock for FakeClock {
        fn now(&self) -> i64 {
            self.now.get()
        }
        fn sleep(&self, secs: u64) {
            self.slept.set(self.slept.get() + secs);
            self.now.set(self.now.get() + secs as i64);
        }
    }

    /// `--wait`: with nothing to do it reads again every POLL_SECS until the wait is over, and
    /// a wait above MAX_WAIT_SECS is refused.
    #[test]
    fn wait_reads_again_until_the_wait_is_over() {
        let dir = tempfile::TempDir::new().unwrap();
        let target = Target {
            owner: "o",
            repo: "r",
            lane: "agent-10",
            tip_branch: Some("main"),
            dir: dir.path(),
        };
        let clock = FakeClock {
            now: Cell::new(0),
            slept: Cell::new(0),
        };
        let v = run(&target, &Quiet, &clock, 900, SHARED_LANES).unwrap();
        assert_eq!(v["next"], Value::Null);
        assert_eq!(v["rounds"], json!(3));
        assert_eq!(v["poll_secs"], json!(432));
        assert_eq!(clock.slept.get(), 2 * POLL_SECS);
        let once = run(&target, &Quiet, &clock, 0, SHARED_LANES).unwrap();
        assert_eq!(once["rounds"], json!(1));
        assert!(run(&target, &Quiet, &clock, MAX_WAIT_SECS + 1, SHARED_LANES).is_err());
        assert_eq!((POLL_SECS, MAX_WAIT_SECS), (432, 3600));
        // #2873 N9: twenty lanes on the token poll half as often.
        clock.slept.set(0);
        let v = run(&target, &Quiet, &clock, 900, 20).unwrap();
        assert_eq!(
            (v["rounds"].clone(), v["poll_secs"].clone()),
            (json!(2), json!(864))
        );
        assert_eq!(clock.slept.get(), 864);
    }

    /// #2873 N9: `--shared-lanes` accepts 1 to 50 — 0 and 51 are refused.
    #[test]
    fn shared_lanes_are_bounded() {
        let dir = tempfile::TempDir::new().unwrap();
        let clock = FakeClock {
            now: Cell::new(0),
            slept: Cell::new(0),
        };
        let t = Target {
            tip_branch: Some("main"),
            ..target(dir.path())
        };
        for (lanes, ok) in [(0, false), (1, true), (50, true), (51, false)] {
            assert_eq!(run(&t, &Quiet, &clock, 0, lanes).is_ok(), ok, "{lanes}");
        }
        assert_eq!(MAX_SHARED_LANES, 50);
        assert_eq!((poll_secs(1), poll_secs(50)), (44, 2160));
    }

    /// #2933 S1: the time a decision was made comes from the snapshot, or — when reading
    /// GitHub — from the clock.
    #[test]
    fn now_comes_from_the_snapshot_or_the_clock() {
        let snap = Snapshot {
            now: Some("2026-11-14T12:00:00Z".into()),
            ..Snapshot::default()
        };
        assert_eq!(
            decide("agent-10", &snap)["now"],
            json!("2026-11-14T12:00:00Z")
        );
        assert!(decide("agent-10", &Snapshot::default())
            .get("now")
            .is_none());
        let dir = tempfile::TempDir::new().unwrap();
        let clock = FakeClock {
            now: Cell::new(crate::forge::policy::parse_utc("2026-11-14T12:34:56Z").unwrap()),
            slept: Cell::new(0),
        };
        let t = Target {
            tip_branch: Some("main"),
            ..target(dir.path())
        };
        let v = run(&t, &Quiet, &clock, 0, SHARED_LANES).unwrap();
        assert_eq!(v["now"], json!("2026-11-14T12:34:56Z"), "{v}");
    }

    /// #2933 S1: with nothing to do the terminal line counts what was waited on, skipped and
    /// passed over — fixed words and numbers only.
    #[test]
    fn the_summary_counts_what_was_passed_over() {
        let v = json!({"next": null, "waiting": [],
            "skipped": [{"issue": 7, "reason": "held"}],
            "pr_skipped": [{"pr": 5, "reason": "draft"}, {"pr": 6, "reason": "draft"}]});
        assert_eq!(
            summary_line(&v),
            "lane-next: nothing to do now (0 waiting, 1 skipped, 2 PRs passed)"
        );
    }

    /// #2856 audit B1: ten lanes polling with one token stay within half of GitHub's 5,000
    /// requests per hour (literals, so a changed number fails here).
    #[test]
    fn polling_stays_within_the_shared_hourly_budget() {
        assert_eq!((API_BUDGET, SHARED_LANES, HOURLY_SHARE), (30, 10, 2_500));
        for lanes in 1..=MAX_SHARED_LANES {
            let poll = poll_secs(lanes);
            let per_hour = lanes * API_BUDGET as u64 * 3600 / poll;
            assert!(
                per_hour <= HOURLY_SHARE,
                "{lanes} lanes: {per_hour} per hour"
            );
            // One second less would go over (or reach the share exactly).
            assert!(
                lanes * API_BUDGET as u64 * 3600 / (poll - 1) >= HOURLY_SHARE,
                "{lanes}"
            );
            assert!(poll <= MAX_WAIT_SECS);
        }
    }

    /// The tip branch, and a failed check run on the tip commit; nothing else.
    struct RedTip;
    impl Transport for RedTip {
        fn send(&self, req: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
            let body = if req.url.contains("/branches/") {
                json!({"commit": {"sha": "c".repeat(40)}})
            } else if req.url.contains("/check-runs") {
                json!({"check_runs": [{"name": "ci / macos-arm64", "status": "completed",
                    "conclusion": "failure",
                    "details_url": "https://github.com/o/r/actions/runs/9/job/1"}]})
            } else {
                json!([])
            };
            Ok(crate::forge::transport::HttpResponse::new(200).with_body(body))
        }
    }

    /// The local clone lacks the newest tip (no fetch: reads only): the tip commit's own check
    /// runs still decide, `partial` says the walk was skipped, and a red tip is P0.
    #[test]
    fn a_tip_the_clone_lacks_is_judged_from_its_own_check_runs() {
        let dir = tempfile::TempDir::new().unwrap();
        let target = Target {
            owner: "o",
            repo: "r",
            lane: "agent-10",
            tip_branch: Some("main"),
            dir: dir.path(),
        };
        let snap = collect(&target, &RedTip);
        let tip = snap.tip.clone().expect("judged");
        assert_eq!(tip.state, "red", "{snap:?}");
        assert_eq!(tip.tip, "c".repeat(40));
        assert!(
            snap.partial
                .iter()
                .any(|p| p.contains("judged the tip commit alone")),
            "{:?}",
            snap.partial
        );
        assert_eq!(decide("agent-10", &snap)["next"]["kind"], json!("tip_red"));
    }

    fn check(name: &str, conclusion: &str) -> Value {
        json!({"name": name, "status": "completed", "conclusion": conclusion,
               "details_url": "https://github.com/o/r/actions/runs/7/job/1"})
    }

    /// #2831 D8(a): a newer push cancels the macOS jobs and the aggregate `ci` then reads
    /// `failure`; that is not red. Only a job that itself failed is (run 43550300926 shape).
    #[test]
    fn the_aggregate_and_cancelled_jobs_do_not_make_the_tip_red() {
        let cancelled_by_newer_push = [
            check("ci", "failure"),
            check("ci / linux-x86_64", "success"),
            check("ci / linux-arm64", "success"),
            check("windows-build", "success"),
            check("ci / macos-x86_64", "cancelled"),
            check("ci / macos-arm64", "cancelled"),
        ];
        assert_eq!(
            d8a_judgment(&cancelled_by_newer_push)["state"],
            json!("green")
        );
        let job_failed = [
            check("ci", "failure"),
            check("ci / linux-x86_64", "success"),
            check("ci / macos-arm64", "failure"),
        ];
        let j = d8a_judgment(&job_failed);
        assert_eq!(j["state"], json!("red"), "{j}");
        assert!(j["failed"].to_string().contains("ci / macos-arm64"));
        // Only cancelled or only the aggregate: nothing judged yet.
        assert_eq!(
            d8a_judgment(&[
                check("ci", "failure"),
                check("ci / macos-arm64", "cancelled")
            ])["state"],
            json!("pending")
        );
        // A failure a re-run turned green on the same commit is flaky, not red.
        assert_eq!(
            d8a_judgment(&[
                check("ci / macos-arm64", "failure"),
                check("ci / macos-arm64", "success")
            ])["state"],
            json!("flaky")
        );
        assert_eq!(
            d8a_judgment(&[
                check("rc-ci", "failure"),
                check("ci / linux-x86_64", "success")
            ])["state"],
            json!("green")
        );
    }

    /// Answers `commits/<sha>/check-runs` from a table; every other commit has no runs.
    struct WalkRuns(std::collections::HashMap<String, Vec<Value>>);
    impl Transport for WalkRuns {
        fn send(&self, req: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
            let sha = req
                .url
                .split("/commits/")
                .nth(1)
                .and_then(|r| r.split('/').next())
                .unwrap_or("");
            let runs = self.0.get(sha).cloned().unwrap_or_default();
            Ok(crate::forge::transport::HttpResponse::new(200)
                .with_body(json!({"check_runs": runs})))
        }
    }

    /// A temporary repository with `n` empty commits; their SHAs, newest first.
    fn commits(dir: &Path, n: usize) -> Vec<String> {
        let git = |args: &[&str]| {
            let out = crate::audit_delta::command(dir)
                .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
                .args(args)
                .output()
                .unwrap();
            assert!(out.status.success(), "{args:?}: {out:?}");
            String::from_utf8(out.stdout).unwrap()
        };
        git(&["init", "-q", "-b", "main"]);
        for i in 0..n {
            git(&["commit", "-q", "--allow-empty", "-m", &format!("c{i}")]);
        }
        git(&["rev-list", "--first-parent", "HEAD"])
            .lines()
            .map(str::to_string)
            .collect()
    }

    /// #2856 delta re-audit B1: while the tip commit reads `pending` (only cancelled, or still
    /// running), the earlier first-parent commits decide — a real failure one commit back is
    /// red, decided by that commit.
    #[test]
    fn a_pending_tip_walks_back_to_the_commit_that_decides() {
        let d = tempfile::TempDir::new().unwrap();
        let shas = commits(d.path(), 3);
        let runs = std::collections::HashMap::from([
            (
                shas[0].clone(),
                vec![
                    check("ci / macos-arm64", "cancelled"),
                    check("ci", "failure"),
                ],
            ),
            (
                shas[1].clone(),
                vec![
                    check("ci / linux-x86_64", "success"),
                    check("ci / macos-arm64", "failure"),
                ],
            ),
            (shas[2].clone(), vec![check("ci / linux-x86_64", "success")]),
        ]);
        let (tip, note) = tip_d8a(
            &WalkRuns(runs),
            "https://api.github.com/repos/o/r",
            d.path(),
            &shas[0],
        )
        .unwrap();
        assert_eq!(tip.state, "red", "{tip:?}");
        assert_eq!(tip.tip, shas[0]);
        assert_eq!(tip.decided_by.as_deref(), Some(shas[1].as_str()));
        assert_eq!(note, None, "the walk ran in the local clone");
    }

    /// The walk reads at most `TIP_WALK` (10) commits: a decision on the 10th is used, one on
    /// the 11th is beyond the walk and the tip stays `unknown`.
    #[test]
    fn the_walk_stops_at_ten_commits() {
        assert_eq!(crate::gate_collect::TIP_WALK, 10);
        let d = tempfile::TempDir::new().unwrap();
        let shas = commits(d.path(), 11);
        for (decider, want) in [(9usize, "red"), (10, "unknown")] {
            let runs = std::collections::HashMap::from([(
                shas[decider].clone(),
                vec![check("ci / macos-arm64", "failure")],
            )]);
            let (tip, _) = tip_d8a(
                &WalkRuns(runs),
                "https://api.github.com/repos/o/r",
                d.path(),
                &shas[0],
            )
            .unwrap();
            assert_eq!(
                tip.state,
                want,
                "decision on commit {}: {tip:?}",
                decider + 1
            );
            if want == "red" {
                assert_eq!(tip.decided_by.as_deref(), Some(shas[decider].as_str()));
            }
        }
    }

    #[test]
    fn comment_lines_are_read_strictly() {
        let h = "a".repeat(40);
        assert_eq!(
            audit_claim(&c(1, &format!("moai-audit-claim: lane=agent-10 head={h}"))),
            Some(("agent-10".into(), h.clone()))
        );
        assert_eq!(
            verdict(&c(
                2,
                &format!("AUDIT #5 (x) head {h} base {h} — verdict: FAIL\nbody")
            )),
            Some((h.clone(), "FAIL"))
        );
        assert_eq!(verdict(&c(3, "AUDIT #5 head short — verdict: PASS")), None);
        assert_eq!(
            verdict(&c(
                4,
                &format!("not first\nAUDIT #5 head {h} verdict: PASS")
            )),
            None
        );
        assert_eq!(
            claim_session(&c(5, "moai-claim: session=agent-10 epoch=1 lease_until=x")),
            Some("agent-10".into())
        );
        assert_eq!(
            body_claim("x\nMoai-Claim: 2817#6867074561\n"),
            Some((2817, 6867074561))
        );
    }

    /// WF-OBS 6874554271 (#2401): a body that names its claim only inside a list item or code
    /// span still names its author lane, as the merge gate's G11 (`has_body_claim`) already
    /// reads it anywhere; a bare line wins over an inline mention, and a placeholder is no claim.
    #[test]
    fn body_claim_reads_an_inline_token_when_no_bare_line_exists() {
        assert_eq!(
            body_claim("## 요약\n- claim 토큰: `Moai-Claim: 2825#6873551363`\n"),
            Some((2825, 6873551363))
        );
        assert_eq!(
            body_claim("see `Moai-Claim: 1#2` above\nMoai-Claim: 2825#6873551363\n"),
            Some((2825, 6873551363))
        );
        assert_eq!(body_claim("- token `Moai-Claim: <n>#<id>`\n"), None);
        assert_eq!(body_claim("- token `Moai-Claim: 2825#`\n"), None);
        assert_eq!(body_claim("- token `Moai-Claim: #6873551363`\n"), None);
        assert_eq!(
            body_claim("- `Moai-Claim: <n>#<id>` then `Moai-Claim:2825#7`\n"),
            Some((2825, 7))
        );
    }
}
