//! Agent self-claim queue: `queue-next` (rc.2 A2 #2161; SPEC
//! `docs/design-review/autonomy-modes-spec.md` §3 steps 1-6).
//!
//! One call re-reads the open Issues of the repository (the remote is the source of truth,
//! user decision D-4), keeps the cards of one milestone, drops pull requests, epics (title
//! `][EP-`) and cards whose `Blocked-By:` line names an Issue that is still open, orders the
//! rest by parent epic (`상위: #N`) and number, and claims the first one it can through
//! 06-A's [`crate::claim::acquire`] under one journal lock for the whole run.
//!
//! - `already_owned` against a live foreign claim → the card is recorded in `lost_to` and the
//!   next candidate is tried. The loser writes nothing to the journal (`claim.rs`, unchanged).
//! - `claimed` → stop. `uncertain` → stop: the comment is posted and this session may be the
//!   winner (SPEC §4 item 4), so no further candidate is touched. `expired`, or an active
//!   record of this session, → stop and return it.
//! - Projection (SPEC §3 step 6) is recorded as intents only. RC1 does not apply state labels
//!   (`.github/moai/labels.yml`, "state lives in the report comment"), and the owned report
//!   comment belongs to `forge-report` with its own gates, so both are `not_run` here.
//!
//! - A candidate whose comments already carry any `moai-claim:` line — live or expired —
//!   is left out before acquire and recorded in `skipped` with `claimed_by` and `expired`
//!   (REQ-A2-01 "only Issues without a claim comment", REQ-A2-02 / REQ-A7-02 "no automatic
//!   takeover"): nothing is posted to it. The one exception is a card this journal already
//!   holds an active record for, which acquire returns unchanged.
//!
//! Not handled here: a local assignments file is not read (the remote is authoritative);
//! live GitHub runs are not part of the tests (fixtures only).
use crate::claim::{
    acquire, is_active, journal_root, parse_claim_line, task_key, worktree_id, ClaimComment,
    ClaimRequest, Journal, Owner, SpecRef, TaskRef,
};
use crate::forge::policy::{next_link, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use crate::identity;
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The record state pr-confirm writes once the PR is handed to the gate (#2344).
const HANDED_OFF_STATE: &str = "pr_ready";

const API: &str = "https://api.github.com";
/// Same safety bound as the claim comment pagination.
const MAX_PAGES: u32 = 50;

/// The caller-supplied part of every claim request; `queue-next` fills `task_ref` and
/// `logical_key` per candidate and defaults nothing else.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ClaimTemplate {
    pub owner: Owner,
    pub base_revision: String,
    pub write_scope: Vec<String>,
    pub spec_ref: SpecRef,
    pub lease_until: String,
}

/// `queue-next` stdin (typed JSON, unknown keys rejected).
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct QueueRequest {
    pub schema_version: u32,
    pub owner: String,
    pub repo: String,
    /// Milestone number whose cards are candidates.
    pub milestone: i64,
    pub claim: ClaimTemplate,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Candidate {
    pub number: i64,
    pub id: i64,
    pub logical_key: String,
    pub parent: Option<i64>,
}

/// Issue numbers named on lines that start with `prefix` (`#123` tokens only).
pub(crate) fn refs_on(body: &str, prefix: &str) -> Vec<i64> {
    body.lines()
        .filter_map(|l| l.trim().strip_prefix(prefix))
        .flat_map(|rest| {
            rest.split('#')
                .skip(1)
                .filter_map(|t| {
                    let digits: String = t.chars().take_while(char::is_ascii_digit).collect();
                    digits.parse().ok()
                })
                .collect::<Vec<i64>>()
        })
        .collect()
}

/// The second bracket of `[rc.2][A2] …` → `A2`; otherwise `issue-<n>`.
fn logical_key(title: &str, number: i64) -> String {
    let brackets: Vec<&str> = title
        .split('[')
        .skip(1)
        .filter_map(|s| s.split_once(']').map(|(k, _)| k.trim()))
        .collect();
    match brackets.get(1) {
        Some(k) if !k.is_empty() && k.len() <= 100 && !k.chars().any(char::is_control) => {
            (*k).to_string()
        }
        _ => format!("issue-{number}"),
    }
}

/// Splits the open Issues into ordered candidates and the cards left out (with a reason).
/// `issues` are GitHub issue objects of every open Issue of the repository, so a blocker in
/// another milestone still counts as open.
pub fn select(issues: &[Value], milestone: i64) -> (Vec<Candidate>, Vec<Value>) {
    let open: BTreeSet<i64> = issues.iter().filter_map(|i| i["number"].as_i64()).collect();
    let mut picked = Vec::new();
    let mut skipped = Vec::new();
    for i in issues {
        let (Some(number), Some(id)) = (i["number"].as_i64(), i["id"].as_i64()) else {
            continue;
        };
        if i["milestone"]["number"].as_i64() != Some(milestone) {
            continue;
        }
        let title = i["title"].as_str().unwrap_or_default();
        let body = i["body"].as_str().unwrap_or_default();
        let reason = if i.get("pull_request").is_some_and(|p| !p.is_null()) {
            Some(json!("pull_request"))
        } else if title.contains("][EP-") {
            Some(json!("epic"))
        } else {
            let open_blockers: Vec<i64> = refs_on(body, "Blocked-By:")
                .into_iter()
                .filter(|n| open.contains(n) && *n != number)
                .collect();
            (!open_blockers.is_empty()).then(|| json!({"blocked_by": open_blockers}))
        };
        match reason {
            Some(r) => skipped.push(json!({"issue": number, "reason": r})),
            None => picked.push(Candidate {
                number,
                id,
                logical_key: logical_key(title, number),
                parent: refs_on(body, "상위:").first().copied(),
            }),
        }
    }
    picked.sort_by_key(|c| (c.parent.unwrap_or(i64::MAX), c.number));
    (picked, skipped)
}

fn get_json<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    url: &str,
) -> Result<(Value, Option<String>)> {
    let resp = client.get(url, None).map_err(String::from)?;
    let next = next_link(resp.header("link"));
    Ok((resp.body.unwrap_or(Value::Null), next))
}

/// Every open Issue object of the repository (GET only, `Link` pagination).
fn read_open_issues<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
) -> Result<Vec<Value>> {
    let mut url = Some(format!(
        "{API}/repos/{owner}/{repo}/issues?state=open&per_page=100"
    ));
    let mut out = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err("queue issue page limit exceeded (50)".into());
        }
        let (body, next) = get_json(client, &u)?;
        out.extend(body.as_array().cloned().unwrap_or_default());
        url = next;
    }
    Ok(out)
}

/// Every `moai-claim:` comment of one Issue (GET only, `Link` pagination).
fn read_claim_comments<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    owner: &str,
    repo: &str,
    number: i64,
) -> Result<Vec<ClaimComment>> {
    let mut url = Some(format!(
        "{API}/repos/{owner}/{repo}/issues/{number}/comments?per_page=100"
    ));
    let mut out = Vec::new();
    let mut closed = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err("queue claim comment page limit exceeded (50)".into());
        }
        let (body, next) = get_json(client, &u)?;
        for c in body.as_array().cloned().unwrap_or_default() {
            let (Some(id), Some(text)) = (c["id"].as_i64(), c["body"].as_str()) else {
                continue;
            };
            if let Some(mut line) = parse_claim_line(text) {
                line.comment_id = id;
                out.push(line);
            }
            // #2255: a claim closed by a release or handoff line no longer holds the card.
            match crate::claim_verify::classify_comment(text) {
                crate::claim_verify::CommentKind::Release { target }
                | crate::claim_verify::CommentKind::Handoff { target, .. } => closed.push(target),
                _ => {}
            }
        }
        url = next;
    }
    out.retain(|c| !closed.contains(&c.comment_id));
    Ok(out)
}

/// The `skipped` entry of a card that already carries claim comments: the smallest live
/// comment id if one is live, otherwise the smallest comment id (expired).
fn claimed_entry(number: i64, claims: &[ClaimComment], now: i64) -> Option<Value> {
    let live = |c: &ClaimComment| parse_utc(&c.lease_until).is_some_and(|t| t > now);
    let first = claims
        .iter()
        .filter(|c| live(c))
        .min_by_key(|c| c.comment_id)
        .or_else(|| claims.iter().min_by_key(|c| c.comment_id))?;
    Some(json!({
        "issue": number,
        "reason": "claimed",
        "claimed_by": first.session,
        "comment_id": first.comment_id,
        "lease_until": first.lease_until,
        "expired": !live(first),
    }))
}

fn validate(req: &QueueRequest) -> Result<()> {
    if req.schema_version != 1 {
        return Err("unsupported request schema_version".into());
    }
    if req.milestone <= 0 {
        return Err("queue milestone must be a positive number".into());
    }
    Ok(())
}

/// The run: read, select, then claim the first candidate that can be claimed. The caller
/// holds the journal lock for the whole run.
pub fn next<T: Transport, C: Clock>(
    journal: &Journal,
    req: &QueueRequest,
    worktree: crate::claim::WorktreeId,
    client: &Client<'_, T, C>,
    clock: &C,
) -> Result<Value> {
    validate(req)?;
    let (repo, _) = get_json(client, &format!("{API}/repos/{}/{}", req.owner, req.repo))?;
    let repository_id = repo["id"]
        .as_i64()
        .ok_or("gh transport: repository response carried no id")?;
    let issues = read_open_issues(client, &req.owner, &req.repo)?;
    let (candidates, mut skipped) = select(&issues, req.milestone);
    let events = journal.read_events()?;
    let projection = json!({
        "label": {"state": "not_run", "reason": "RC1 does not apply state labels (.github/moai/labels.yml)"},
        "status_line": {"state": "not_run", "reason": "the owned report comment is forge-report's, behind its own gates"},
    });
    let mut lost_to = Vec::new();
    for c in &candidates {
        // #2169 display projections (assignee, provider_session) stay empty: queue-next's
        // request carries neither, and they never decide ownership (REQ-B2-01). The
        // constructor leaves them and expected_epoch/attempt unset (#2256).
        let claim = ClaimRequest::new(
            1,
            TaskRef::new(
                "github",
                "github.com",
                req.owner.clone(),
                req.repo.clone(),
                repository_id,
                c.id,
                c.number,
            ),
            c.logical_key.clone(),
            req.claim.owner.clone(),
            req.claim.base_revision.clone(),
            req.claim.write_scope.clone(),
            req.claim.spec_ref.clone(),
        )
        .lease_until(req.claim.lease_until.clone())
        .projection_intents(vec![
            json!({"kind": "projection_intent", "target": "label", "action": "not_run"}),
            json!({"kind": "projection_intent", "target": "status_line", "action": "not_run"}),
        ]);
        // REQ-A2-01/02: a card that already carries a claim comment (live or expired) is
        // not a candidate, unless this journal holds an active record for it.
        let key = task_key(&claim.task_ref);
        let own = events.iter().rev().find(|e| e.record.task_key == key);
        // #2344: a card this lane already handed to the gate (`pr_ready`, written by
        // pr-confirm) stays open until the merge, but it is not work to pick up again. Skip
        // it here only; the record itself stays active so a repair handoff, worktree ops and
        // scope-check keep treating it as owned.
        if let Some(e) = own.filter(|e| e.record.state == HANDED_OFF_STATE) {
            skipped.push(json!({
                "issue": c.number,
                "reason": "handed_off",
                "state": e.record.state,
                "epoch": e.record.epoch,
            }));
            continue;
        }
        let own_active = own.is_some_and(|e| is_active(&e.record.state));
        if !own_active {
            let claims = read_claim_comments(client, &req.owner, &req.repo, c.number)?;
            if let Some(entry) = claimed_entry(c.number, &claims, clock.now()) {
                skipped.push(entry);
                continue;
            }
        }
        let out = acquire(journal, &claim, worktree.clone(), client, clock)?;
        let foreign = out.foreign_owner.clone();
        if out.outcome == "already_owned" {
            if let Some(f) = foreign {
                // A live foreign claim wins (smallest comment id); nothing was written here.
                lost_to.push(json!({"issue": c.number, "comment_id": f["comment_id"]}));
                continue;
            }
        }
        let mut v = out.to_json();
        v["issue"] = json!(c.number);
        v["lost_to"] = json!(lost_to);
        v["skipped"] = json!(skipped);
        v["projection"] = if out.outcome == "claimed" {
            projection
        } else {
            Value::Null
        };
        return Ok(v);
    }
    let outcome = if candidates.is_empty() {
        "idle"
    } else {
        "exhausted"
    };
    Ok(json!({
        "outcome": outcome,
        "issue": Value::Null,
        "record": Value::Null,
        "foreign_owner": Value::Null,
        "coordination_scope": "cross-host-eventual",
        "cross_host_atomic": false,
        "lost_to": lost_to,
        "skipped": skipped,
        "projection": Value::Null,
        "notes": [if candidates.is_empty() {
            "no candidate: record the idle line once and stop (SPEC §3 step 2)"
        } else {
            "every candidate already carries a claim comment or is owned by a live foreign claim"
        }],
    }))
}

/// CLI: identity + worktree, journal open + one lock for the whole run (busy never waits).
pub fn next_cli<T: Transport, C: Clock>(
    project: &Path,
    req: QueueRequest,
    transport: &T,
    clock: &C,
) -> std::result::Result<Value, crate::contract::OpError> {
    let root = journal_root(project)?;
    let id = identity::identity(project)?;
    let worktree = worktree_id(&id)?;
    let journal = Journal::open(&root)?;
    if !journal.try_lock()? {
        // #2345: the same journal lock-busy answer as every other journal-writing op,
        // with the kind set at the raise site.
        return Err(crate::contract::OpError::new(
            crate::contract::Kind::Unavailable,
            crate::claim::JOURNAL_LOCK_BUSY,
        ));
    }
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let v = next(&journal, &req, worktree, &client, clock)?;
    // HOLD-00D-1: the picked card's acquire refusal (expired, uncertain) is a failure, as
    // in claim-acquire. Foreign owners were skipped into `lost_to` above.
    match crate::claim::acquire_refusal(&v, None) {
        Some(refused) => Err(refused),
        None => Ok(v),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn issue(number: i64, title: &str, body: &str, milestone: i64) -> Value {
        json!({"number": number, "id": number * 10, "title": title, "body": body,
               "milestone": {"number": milestone}})
    }

    #[test]
    fn selection_drops_prs_epics_blocked_and_orders_by_parent() {
        let issues = vec![
            issue(10, "[rc.2][EP-A] epic", "", 3),
            issue(12, "[rc.2][A2] two", "상위: #10\nBlocked-By: #11", 3),
            issue(11, "[rc.2][A1] one", "상위: #10\nBlocked-By: 없음", 3),
            issue(21, "[rc.2][B1] b", "상위: #20\nBlocked-By: #999", 3),
            issue(13, "[rc.2][A3] three", "상위: #10", 3),
            issue(30, "[rc.2][X] other milestone", "", 4),
            json!({"number": 14, "id": 140, "title": "pr", "body": "", "milestone": {"number": 3},
                   "pull_request": {"url": "x"}}),
        ];
        let (picked, skipped) = select(&issues, 3);
        let order: Vec<i64> = picked.iter().map(|c| c.number).collect();
        // #999 is not open, so #21 is not blocked; parent #10 sorts before #20.
        assert_eq!(order, vec![11, 13, 21]);
        assert_eq!(picked[0].logical_key, "A1");
        assert_eq!(picked[0].parent, Some(10));
        let reasons: Vec<(i64, Value)> = skipped
            .iter()
            .map(|s| (s["issue"].as_i64().unwrap(), s["reason"].clone()))
            .collect();
        assert!(reasons.contains(&(10, json!("epic"))));
        assert!(reasons.contains(&(12, json!({"blocked_by": [11]}))));
        assert!(reasons.contains(&(14, json!("pull_request"))));
        assert!(
            !reasons.iter().any(|(n, _)| *n == 30),
            "other milestones are not listed"
        );
    }

    #[test]
    fn logical_key_falls_back_to_issue_number() {
        assert_eq!(logical_key("[rc.2][A2] x", 7), "A2");
        assert_eq!(logical_key("plain title", 7), "issue-7");
        assert_eq!(logical_key("[only] one", 7), "issue-7");
    }

    /// #2255: a claim closed by a release or handoff line no longer counts, so the card can
    /// be a candidate again; an open claim still does.
    #[test]
    fn released_claims_do_not_hold_the_card() {
        use crate::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
        struct Fixed;
        impl Clock for Fixed {
            fn now(&self) -> i64 {
                0
            }
            fn sleep(&self, _: u64) {}
        }
        let lease = "2026-11-09T09:00:00Z";
        let page = |body: Value| Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/repos/acme/widgets/issues/7/comments?per_page=100"),
            ),
            response: HttpResponse::new(200).with_body(body),
        };
        let claim = |id: i64, s: &str| json!({"id": id, "body": format!("moai-claim: session={s} epoch=1 lease_until={lease}")});
        let t = RecordedTransport::new(vec![
            page(json!([
                claim(100, "a"),
                {"id": 150, "body": crate::claim_verify::render_release_line(100, "leader")},
            ])),
            page(json!([
                claim(100, "a"),
                {"id": 150, "body": crate::claim_verify::render_handoff_line(100, "b", "leader")},
                claim(200, "b"),
            ])),
        ]);
        let client = Client::new(&t, &Fixed, Default::default(), Default::default());
        assert!(read_claim_comments(&client, "acme", "widgets", 7)
            .unwrap()
            .is_empty());
        let open = read_claim_comments(&client, "acme", "widgets", 7).unwrap();
        assert_eq!(open.iter().map(|c| c.comment_id).collect::<Vec<_>>(), [200]);
    }
}
