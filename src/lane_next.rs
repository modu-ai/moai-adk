//! `lane-next` (#2831 C1): the one thing a lane should do next, and why, without a leader.
//!
//! Lanes learned their next task only from the leader reading their screens: a delta re-audit
//! whose head had moved went unnoticed (#2821), a decision already posted was still awaited
//! (#2827 6868439066), a re-audit request was missed (#2820). This reads GitHub (open pull
//! requests and their comments, the lane's claim threads, a few open Issues) and the tip's CI,
//! and picks one task in the order of the user's priority list (#1975 6855190942):
//!
//! 1. P0 the tip is red (`gate_collect::tip_ci`, the D8(a) judgment the merge gate uses);
//! 2. P1 the lane's own pull request: an audit FAIL on its head, a conflict, a failed CI;
//! 3. P1 a pull request the lane reserved to audit: not judged yet, its head moved past the
//!    lane's verdict (an update-branch too) or a re-audit was asked for — with the head now;
//! 4. P1 another lane's pull request no auditor reserved;
//! 5. a `LANE-BLOCKED kind=decision` of the lane that a `LEADER-DECISION` now answers;
//! 6. P2 the next step of a card the lane claims (no open pull request of its own for it);
//! 7. P2–P4 an open Issue nobody claims.
//!
//! Reads only (leader decision D1): reservations and comments stay with the existing commands.
//! One collection sends at most [`API_BUDGET`] GitHub requests (D4), taking the higher
//! priorities first; what the budget left unread is named in `partial`, never guessed.

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
/// Open Issues whose comments are read to find one nobody claims.
pub const FREE_ISSUE_SCAN: usize = 5;
/// Lanes assumed to poll at once with one GitHub token (they all read as one account).
pub const SHARED_LANES: u64 = 10;
/// The part of GitHub's 5,000 requests per hour per token the lanes' polling may use: half,
/// so the leader loop and the merge gate keep the other half (#2856 audit B1).
pub const HOURLY_SHARE: u64 = 2_500;
// The share is at most half of GitHub's 5,000 per hour, checked when the crate compiles.
const _: () = assert!(HOURLY_SHARE * 2 <= 5_000);
/// Seconds between two collections while `--wait` waits, derived from the budget so that
/// `SHARED_LANES` lanes each sending `API_BUDGET` requests per collection stay within
/// `HOURLY_SHARE` per hour: 3600 × 30 × 10 / 2500 = 432.
pub const POLL_SECS: u64 = (3600 * API_BUDGET as u64 * SHARED_LANES).div_ceil(HOURLY_SHARE);
/// The longest `--wait` accepted (seconds).
pub const MAX_WAIT_SECS: u64 = 3600;

/// One comment as read: its id and body (oldest first within a thread).
#[derive(Clone, Debug, Default, Serialize, Deserialize, PartialEq)]
pub struct Comment {
    pub id: u64,
    pub body: String,
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
    #[serde(default)]
    pub comments: Vec<Comment>,
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
    #[serde(default)]
    pub tip: Option<Tip>,
    #[serde(default)]
    pub prs: Vec<Pr>,
    /// Issue threads: the lane's claim threads and threads holding its decision requests.
    #[serde(default)]
    pub threads: Vec<Thread>,
    /// Open Issues read as candidates for unclaimed work, oldest first.
    #[serde(default)]
    pub open_issues: Vec<Thread>,
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

fn has_audit_claim(pr: &Pr) -> bool {
    pr.comments.iter().any(|c| audit_claim(c).is_some())
}

/// Decides the next task of `lane` from `snap`. Pure: the same snapshot gives the same bytes.
pub fn decide(lane: &str, snap: &Snapshot) -> Value {
    let mut waiting = Vec::new();
    let next = (|| {
        // 1. P0: a red tip stops everything else.
        if let Some(tip) = snap.tip.as_ref().filter(|t| t.state == "red") {
            return Some(task(
                "P0",
                "tip_red",
                json!({"type": "tip", "head": tip.tip}),
                "repair the red tip (a PR with a `moai-tip-repair:` line)".into(),
                json!({"decided_by": tip.decided_by, "detail": tip.detail}),
            ));
        }
        let mut prs: Vec<&Pr> = snap.prs.iter().collect();
        prs.sort_by_key(|p| p.number);
        // 2. P1: the lane's own pull requests.
        for pr in prs
            .iter()
            .filter(|p| p.author_lane.as_deref() == Some(lane))
        {
            if let Some(t) = own_pr_task(pr) {
                return Some(t);
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
        // 4. P1: another lane's pull request with no auditor.
        if let Some(pr) = prs
            .iter()
            .find(|p| !p.draft && p.author_lane.as_deref() != Some(lane) && !has_audit_claim(p))
        {
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
            match comments[bi + 1..].iter().find(|c| leader_decision(c)) {
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
            let decision_pending = waiting.iter().any(|w| w["number"] == json!(t.number));
            if decision_pending {
                continue;
            }
            // A pushed branch held back by the open-PR limit (`pr-wait:` after the claim, no pull
            // request of the lane for the card): open it as a draft now instead of waiting for
            // a slot — the merge gate marks a draft ready once it has its PASS (leader decision
            // #1975 6870150891, #2856 audit B2).
            let claim_at = t.comments.iter().position(|c| c.id == claim);
            if let Some(w) = claim_at.map(|i| &t.comments[i + 1..]).and_then(|after| {
                after
                    .iter()
                    .rev()
                    .find(|c| first_line(c).starts_with("pr-wait:"))
            }) {
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
            return Some(task(
                "P2",
                "continue_card",
                json!({"type": "issue", "number": t.number}),
                format!("continue your claimed card #{} ({})", t.number, t.title),
                json!({"claim_comment": claim}),
            ));
        }
        // 7. P2–P4: an open Issue nobody claims (no claim line, not on hold).
        snap.open_issues
            .iter()
            .find(|t| {
                !t.labels.iter().any(|l| l.eq_ignore_ascii_case("hold"))
                    && !t.comments.iter().any(|c| claim_session(c).is_some())
            })
            .map(|t| {
                task(
                    "P2",
                    "pick_issue",
                    json!({"type": "issue", "number": t.number}),
                    format!(
                        "claim open Issue #{} ({}) after reading its newest comments",
                        t.number, t.title
                    ),
                    json!({"labels": t.labels}),
                )
            })
    })();
    json!({
        "lane": lane,
        "next": next,
        "waiting": waiting,
        "partial": snap.partial,
        "api_calls": snap.api_calls,
        "writes": false,
    })
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
            "lane-next: nothing to do now ({} waiting on the leader)",
            v["waiting"].as_array().map_or(0, Vec::len)
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

fn comments_of<T: Transport>(t: &T, base: &str, number: u64) -> Result<(Vec<Comment>, bool)> {
    let (v, next) =
        crate::gate_collect::get_json(t, &format!("{base}/issues/{number}/comments?per_page=100"))?;
    let list = v
        .as_array()
        .map(|a| {
            a.iter()
                .map(|c| Comment {
                    id: c["id"].as_u64().unwrap_or(0),
                    body: c["body"].as_str().unwrap_or("").to_string(),
                })
                .collect()
        })
        .unwrap_or_default();
    Ok((list, next.is_some()))
}

/// `Moai-Claim: <issue>#<comment>` in a body.
fn body_claim(body: &str) -> Option<(u64, u64)> {
    body.lines().find_map(|l| {
        let rest = l.trim().strip_prefix("Moai-Claim:")?.trim();
        let (issue, comment) = rest.split_once('#')?;
        Some((issue.trim().parse().ok()?, comment.trim().parse().ok()?))
    })
}

fn search<T: Transport>(t: &T, owner: &str, repo: &str, text: &str) -> Result<Vec<(u64, String)>> {
    let q = format!("repo:{owner}/{repo} is:open in:comments \"{text}\"");
    let enc: String = q
        .bytes()
        .map(|b| match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' => (b as char).to_string(),
            b' ' => "+".into(),
            _ => format!("%{b:02X}"),
        })
        .collect();
    let (v, _) = crate::gate_collect::get_json(
        t,
        &format!("https://api.github.com/search/issues?q={enc}&per_page=30"),
    )?;
    Ok(v["items"]
        .as_array()
        .map(|a| {
            a.iter()
                .filter_map(|i| {
                    let title = i["title"].as_str().unwrap_or("").to_string();
                    Some((i["number"].as_u64()?, title))
                })
                .collect()
        })
        .unwrap_or_default())
}

/// The tip's state as #2831 states P0 (D8(a)): the aggregate check (`ci`/`rc-ci`) is not
/// judged on its own — a newer push cancels the older run, and the aggregate then reads
/// `failure` with every real job passed or cancelled — and a cancelled job says nothing about
/// the code. The remaining jobs are judged with `flow::tip_ci_judgment`: `red` only when a job
/// itself failed with no success on the same commit.
pub fn d8a_judgment(runs: &[Value]) -> Value {
    let kept: Vec<Value> = runs
        .iter()
        .filter(|r| {
            let name = r["name"].as_str().unwrap_or("");
            !crate::flow::CI_CHECK_NAMES.contains(&name)
                && r["conclusion"].as_str() != Some("cancelled")
        })
        .cloned()
        .collect();
    let mut names: Vec<String> = kept
        .iter()
        .filter_map(|r| r["name"].as_str().map(str::to_string))
        .collect();
    names.sort();
    names.dedup();
    if names.is_empty() {
        return json!({"state": "pending"});
    }
    crate::flow::tip_ci_judgment(&kept, &names)
}

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

/// Reads one snapshot for `target`, higher priorities first, within [`API_BUDGET`] requests.
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
    // P1: open pull requests, their comments and authors.
    let pulls =
        match crate::gate_collect::get_json(&t, &format!("{base}/pulls?state=open&per_page=100")) {
            Ok((v, _)) => v.as_array().cloned().unwrap_or_default(),
            Err(e) => {
                miss("open pull requests".into(), e, &mut snap);
                Vec::new()
            }
        };
    // P0: the tip — the named branch, else the base most open pull requests target.
    let mut bases: BTreeMap<&str, usize> = BTreeMap::new();
    for p in &pulls {
        if let Some(b) = p["base"]["ref"].as_str() {
            *bases.entry(b).or_default() += 1;
        }
    }
    let tip_branch = target.tip_branch.map(str::to_string).or_else(|| {
        bases
            .iter()
            .max_by_key(|(_, n)| **n)
            .map(|(b, _)| (*b).to_string())
    });
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
    let mut claim_lane: BTreeMap<u64, Option<String>> = BTreeMap::new();
    for p in &pulls {
        let number = p["number"].as_u64().unwrap_or(0);
        let mut pr = Pr {
            number,
            title: p["title"].as_str().unwrap_or("").to_string(),
            head: p["head"]["sha"].as_str().unwrap_or("").to_string(),
            draft: p["draft"].as_bool().unwrap_or(false),
            ..Pr::default()
        };
        if let Some((issue, comment)) = body_claim(p["body"].as_str().unwrap_or("")) {
            pr.claim_issue = Some(issue);
            if let std::collections::btree_map::Entry::Vacant(slot) = claim_lane.entry(comment) {
                let lane = match crate::gate_collect::get_json(
                    &t,
                    &format!("{base}/issues/comments/{comment}"),
                ) {
                    Ok((c, _)) => claim_session(&Comment {
                        id: comment,
                        body: c["body"].as_str().unwrap_or("").to_string(),
                    }),
                    Err(e) => {
                        miss(
                            format!("claim comment {comment} of PR #{number}"),
                            e,
                            &mut snap,
                        );
                        None
                    }
                };
                slot.insert(lane);
            }
            pr.author_lane = claim_lane[&comment].clone();
        }
        match comments_of(&t, &base, number) {
            Ok((c, more)) => {
                pr.comments = c;
                if more {
                    snap.partial
                        .push(format!("PR #{number}: only the first 100 comments read"));
                }
            }
            Err(e) => miss(format!("comments of PR #{number}"), e, &mut snap),
        }
        if pr.author_lane.as_deref() == Some(target.lane) {
            match crate::gate_collect::get_json(&t, &format!("{base}/pulls/{number}")) {
                Ok((v, _)) => {
                    pr.conflicting = match (v["mergeable"].as_bool(), v["mergeable_state"].as_str())
                    {
                        (Some(false), _) | (_, Some("dirty")) => Some(true),
                        (Some(true), _) => Some(false),
                        _ => None,
                    }
                }
                Err(e) => miss(format!("mergeable of PR #{number}"), e, &mut snap),
            }
            match crate::gate_collect::tip_runs(&t, &base, &pr.head) {
                Ok(runs) => {
                    let j = crate::gate_collect::completed_judgment(&runs);
                    pr.ci = j["state"].as_str().map(str::to_string);
                    pr.ci_detail = j.get("failed").map(Value::to_string);
                }
                Err(e) => miss(format!("CI of PR #{number}"), e, &mut snap),
            }
        }
        snap.prs.push(pr);
    }
    // 5–6: the lane's claim threads and decision requests.
    let mut numbers = Vec::new();
    for text in [
        format!("session={}", target.lane),
        format!("LANE-BLOCKED {}", target.lane),
    ] {
        match search(&t, target.owner, target.repo, &text) {
            Ok(found) => numbers.extend(found),
            Err(e) => miss(format!("search {text:?}"), e, &mut snap),
        }
    }
    numbers.sort_unstable();
    numbers.dedup_by_key(|(n, _)| *n);
    let pr_numbers: Vec<u64> = snap.prs.iter().map(|p| p.number).collect();
    for (number, title) in numbers.into_iter().filter(|(n, _)| !pr_numbers.contains(n)) {
        match comments_of(&t, &base, number) {
            Ok((comments, _)) => snap.threads.push(Thread {
                number,
                title,
                comments,
                ..Thread::default()
            }),
            Err(e) => miss(format!("comments of #{number}"), e, &mut snap),
        }
    }
    // 7: a few open Issues, oldest first.
    match crate::gate_collect::get_json(
        &t,
        &format!("{base}/issues?state=open&sort=created&direction=asc&per_page=30"),
    ) {
        Ok((v, _)) => {
            let issues: Vec<&Value> = v
                .as_array()
                .map(|a| {
                    a.iter()
                        .filter(|i| i.get("pull_request").is_none())
                        .collect()
                })
                .unwrap_or_default();
            for i in issues.into_iter().take(FREE_ISSUE_SCAN) {
                let number = i["number"].as_u64().unwrap_or(0);
                let labels = i["labels"]
                    .as_array()
                    .map(|a| {
                        a.iter()
                            .filter_map(|l| l["name"].as_str().map(str::to_string))
                            .collect()
                    })
                    .unwrap_or_default();
                match comments_of(&t, &base, number) {
                    Ok((comments, _)) => snap.open_issues.push(Thread {
                        number,
                        title: i["title"].as_str().unwrap_or("").to_string(),
                        labels,
                        comments,
                    }),
                    Err(e) => {
                        miss(format!("comments of open Issue #{number}"), e, &mut snap);
                        break;
                    }
                }
            }
        }
        Err(e) => miss("open Issues".into(), e, &mut snap),
    }
    snap.api_calls = t.used.get();
    snap
}

/// Collects and decides; with `wait_secs` > 0, collects again every [`POLL_SECS`] until a task
/// appears or the wait is over (at most [`MAX_WAIT_SECS`]).
pub fn run<T: Transport, C: Clock>(
    target: &Target<'_>,
    transport: &T,
    clock: &C,
    wait_secs: u64,
) -> Result<Value> {
    if wait_secs > MAX_WAIT_SECS {
        return Err(format!("--wait accepts at most {MAX_WAIT_SECS} seconds"));
    }
    let deadline = clock.now() + wait_secs as i64;
    let mut rounds = 0u32;
    loop {
        rounds += 1;
        let snap = collect(target, transport);
        let mut v = decide(target.lane, &snap);
        let done = !v["next"].is_null() || clock.now() + POLL_SECS as i64 > deadline;
        if done {
            v["rounds"] = json!(rounds);
            return Ok(v);
        }
        clock.sleep(POLL_SECS);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn c(id: u64, body: &str) -> Comment {
        Comment {
            id,
            body: body.into(),
        }
    }

    /// Answers every GET: 40 open pull requests, empty lists elsewhere; counts requests.
    struct Busy {
        sent: Cell<u32>,
    }

    impl Transport for Busy {
        fn send(&self, req: &HttpRequest) -> Result<crate::forge::transport::HttpResponse> {
            self.sent.set(self.sent.get() + 1);
            let body = if req.url.contains("/pulls?state=open") {
                json!((1..=40)
                    .map(|n| json!({"number": n, "head": {"sha": "a".repeat(40)}, "base": {"ref": "main"}, "body": ""}))
                    .collect::<Vec<_>>())
            } else if req.url.contains("/branches/") {
                json!({"commit": {"sha": "b".repeat(40)}})
            } else {
                json!([])
            };
            Ok(crate::forge::transport::HttpResponse::new(200).with_body(body))
        }
    }

    /// D4: one collection sends at most API_BUDGET requests, reads the higher priorities first
    /// and names what it left unread instead of guessing.
    #[test]
    fn a_collection_stops_at_its_request_budget() {
        let t = Busy { sent: Cell::new(0) };
        let dir = tempfile::TempDir::new().unwrap();
        let target = Target {
            owner: "o",
            repo: "r",
            lane: "agent-10",
            tip_branch: None,
            dir: dir.path(),
        };
        let snap = collect(&target, &t);
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
        assert!(snap.tip.is_some() || snap.partial.iter().any(|p| p.starts_with("tip")));
        assert!(!snap.prs.is_empty());
        assert_eq!(API_BUDGET, 30, "the leader's number (#2831 D4)");
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
        let v = run(&target, &Quiet, &clock, 900).unwrap();
        assert_eq!(v["next"], Value::Null);
        assert_eq!(v["rounds"], json!(3));
        assert_eq!(clock.slept.get(), 2 * POLL_SECS);
        let once = run(&target, &Quiet, &clock, 0).unwrap();
        assert_eq!(once["rounds"], json!(1));
        assert!(run(&target, &Quiet, &clock, MAX_WAIT_SECS + 1).is_err());
        assert_eq!((POLL_SECS, MAX_WAIT_SECS), (432, 3600));
    }

    /// #2856 audit B1: ten lanes polling with one token stay within half of GitHub's 5,000
    /// requests per hour (literals, so a changed number fails here).
    #[test]
    fn polling_stays_within_the_shared_hourly_budget() {
        assert_eq!((API_BUDGET, SHARED_LANES, HOURLY_SHARE), (30, 10, 2_500));
        let per_hour = SHARED_LANES * API_BUDGET as u64 * 3600 / POLL_SECS;
        assert!(per_hour <= HOURLY_SHARE, "{per_hour} requests per hour");
        // One second less would go over.
        assert!(SHARED_LANES * API_BUDGET as u64 * 3600 / (POLL_SECS - 1) > HOURLY_SHARE);
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
}
