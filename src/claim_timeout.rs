//! Claim timeout takeover check (#2906 S1): whether a lane may close another lane's claim with
//! `moai-claim-handoff: comment=<id> to=<me> by=timeout` (decision D2, #1975 comment
//! 6840449133). `claim-verify` keeps judging only the lease (decision Q2); the other two
//! conditions are read here, so the lane's record is a command's output rather than prose.
//!
//! The three conditions, each reported with its evidence:
//! - lease expired: `now >= lease_until`, the edge of `is_live = lease_until > now`;
//! - request unanswered: this lane's request names the holder's claim id, `now >= request_at +
//!   1800`, and after the request there is no comment of the holder lane and no release or
//!   handoff of that claim;
//! - no heartbeat: `now >= max(latest activity, claim written) + 7200`, where activity is the
//!   claim comment's `updated_at` (lease renewal), a comment of the holder lane on this Issue
//!   or anywhere in the repository in the last two hours, and a commit on the claim's
//!   `branch=` carrying `Moai-Claim: <issue>#<claim>` (decision Q4; a branch that is gone is
//!   zero commits). An activity time after `now` is never silence.
//!
//! A holder-lane comment is one that names the holder's lane or session as a whole word on any
//! line (`auditor: x`, `lane=x`, `LANE-… x` included): every lane posts through one account, so
//! the text is all there is. Over-counting only blocks a takeover, the safe direction.
//!
//! The request is this lane's only when its first non-empty line names `--lane` as a whole
//! word, it is not a holder-lane comment, and its id is above the claim's (written after it);
//! otherwise `request_not_mine` or `request_before_claim`. A holder-lane comment is never
//! taken out of the holder's activity.
//!
//! Reads are GET only, at most [`MAX_PAGES`] pages per list; a failed read or a page cap
//! overflow is `uncertain`, with no line.

use crate::claim::{claim_token, parse_claim_line};
use crate::claim_verify::{
    classify_comment, evaluate, render_handoff_line, CommentKind, IssueComment, Verifier,
};
use crate::forge::policy::{format_utc, next_link, parse_utc, Clock};
use crate::forge::transport::Transport;
use crate::forge::Client;
use serde_json::{json, Value};
use std::collections::BTreeMap;

/// How long a request naming the holder's claim waits for an answer.
pub const REQUEST_WAIT_SECS: i64 = 1800;
/// How long the holder must show no activity.
pub const HEARTBEAT_SECS: i64 = 7200;
/// Pages read per list at most; more is `uncertain`.
pub const MAX_PAGES: u32 = 10;

/// One comment with its times (`YYYY-MM-DDTHH:MM:SSZ`), as GitHub returns it.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct TimedComment {
    pub id: i64,
    pub author: String,
    pub body: String,
    pub created_at: String,
    pub updated_at: String,
}

/// One commit of the claim branch: its message and committer date.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct BranchCommit {
    pub sha: String,
    pub message: String,
    pub date: String,
}

/// Everything the judgement reads, gathered by [`check`] or built by a test.
#[derive(Clone, Debug, Default)]
pub struct Input {
    pub issue: i64,
    /// The holder's claim comment id.
    pub claim: i64,
    /// This lane's request comment id.
    pub request: i64,
    /// This lane, the `to=` of the line.
    pub me: String,
    /// Every comment of the Issue.
    pub issue_comments: Vec<TimedComment>,
    /// The repository's comments updated since `repo_since`.
    pub repo_comments: Vec<TimedComment>,
    pub repo_since: Option<String>,
    /// False when the claim branch is gone (404); its commits are then empty.
    pub branch_found: bool,
    /// The claim branch's commits since the claim was written.
    pub commits: Vec<BranchCommit>,
}

/// Whether `word` stands in `text` with no letter, digit, `-` or `_` right before or after it.
fn whole_word(text: &str, word: &str) -> bool {
    let joins = |c: char| c.is_ascii_alphanumeric() || c == '-' || c == '_';
    !word.is_empty()
        && text.match_indices(word).any(|(i, _)| {
            !text[..i].chars().next_back().is_some_and(joins)
                && !text[i + word.len()..].chars().next().is_some_and(joins)
        })
}

/// A `key=value` token of the comment's claim line.
fn claim_field<'a>(body: &'a str, key: &str) -> Option<&'a str> {
    let line = body
        .lines()
        .map(str::trim)
        .find(|l| l.starts_with("moai-claim:"))?;
    line.split_whitespace()
        .find_map(|t| t.strip_prefix(key).and_then(|v| v.strip_prefix('=')))
        .filter(|v| !v.is_empty())
}

fn uncertain(input: &Input, reason: String) -> Value {
    json!({
        "met": false,
        "state": "uncertain",
        "reason": reason,
        "issue": input.issue,
        "claim": input.claim,
        "request": input.request,
        "me": input.me,
        "line": null,
        "body": null,
    })
}

/// The pure judgement at `now` (Unix seconds); see the module comment.
pub fn judge(input: &Input, now: i64) -> Value {
    let Some(claim) = input.issue_comments.iter().find(|c| c.id == input.claim) else {
        return json!({"met": false, "state": "not_met", "reasons": ["claim_not_found"],
                      "issue": input.issue, "claim": input.claim, "request": input.request,
                      "me": input.me, "line": null, "body": null});
    };
    let Some(parsed) = parse_claim_line(&claim.body) else {
        return json!({"met": false, "state": "not_met", "reasons": ["claim_line_unreadable"],
                      "issue": input.issue, "claim": input.claim, "request": input.request,
                      "me": input.me, "line": null, "body": null});
    };
    let time = |what: &str, id: &str, t: &str| {
        parse_utc(t).ok_or_else(|| format!("unreadable time `{t}` on {what} {id}"))
    };
    let claim_id = claim.id.to_string();
    let (lease, written, renewed) = match (
        time("claim lease_until", &claim_id, &parsed.lease_until),
        time("claim comment", &claim_id, &claim.created_at),
        time("claim comment", &claim_id, &claim.updated_at),
    ) {
        (Ok(a), Ok(b), Ok(c)) => (a, b, c),
        (Err(e), _, _) | (_, Err(e), _) | (_, _, Err(e)) => return uncertain(input, e),
    };
    let lane = claim_field(&claim.body, "lane")
        .unwrap_or(&parsed.session)
        .to_string();
    let branch = claim_field(&claim.body, "branch");

    // The holder as claim-verify reads it (lowest open claim, renewals, closes).
    let plain: Vec<IssueComment> = input
        .issue_comments
        .iter()
        .map(|c| IssueComment {
            id: c.id,
            author: c.author.clone(),
            body: c.body.clone(),
        })
        .collect();
    let verdict = evaluate(
        &plain,
        &Verifier {
            session: "",
            comment_id: Some(input.claim),
            allowed_authors: &[],
            now,
        },
    );
    let holder = verdict["holder"].as_i64();

    // Every comment read once, Issue and repository lists merged by id.
    let mut all: BTreeMap<i64, &TimedComment> = BTreeMap::new();
    for c in input.issue_comments.iter().chain(&input.repo_comments) {
        all.entry(c.id).or_insert(c);
    }
    // Anywhere in the body: lanes sign on any line (`auditor: x`, `lane=x`, `LANE-… x`).
    let of_lane =
        |c: &TimedComment| whole_word(&c.body, &lane) || whole_word(&c.body, &parsed.session);
    let closes_claim = |c: &TimedComment| match classify_comment(&c.body) {
        CommentKind::Release { target } | CommentKind::Handoff { target, .. } => {
            target == input.claim
        }
        _ => false,
    };

    // 1. Lease.
    let lease_met = now >= lease;

    // 2. Request.
    let answered_by = all
        .values()
        .find(|c| c.id > input.request && (of_lane(c) || closes_claim(c)))
        .map(|c| c.id);
    let request = input.issue_comments.iter().find(|c| c.id == input.request);
    let (request_met, request_json) = match request {
        None => (
            false,
            json!({"met": false, "reason": "request_not_found", "answered_by": answered_by}),
        ),
        Some(r) => {
            let at = match time("request comment", &r.id.to_string(), &r.created_at) {
                Ok(t) => t,
                Err(e) => return uncertain(input, e),
            };
            let names = whole_word(&r.body, &claim_id);
            // This lane's: its first non-empty line names `--lane`, and it is not the holder
            // lane's comment (which stays the holder's activity).
            let mine = r
                .body
                .lines()
                .map(str::trim)
                .find(|l| !l.is_empty())
                .is_some_and(|first| whole_word(first, &input.me))
                && !of_lane(r);
            let after_claim = r.id > input.claim;
            let waited = now >= at + REQUEST_WAIT_SECS;
            let met = names && mine && after_claim && answered_by.is_none() && waited;
            let reason = if !names {
                "request_names_no_claim"
            } else if !mine {
                "request_not_mine"
            } else if !after_claim {
                "request_before_claim"
            } else if answered_by.is_some() {
                "answered"
            } else if !waited {
                "waiting"
            } else {
                "unanswered"
            };
            (
                met,
                json!({"met": met, "reason": reason, "request_at": r.created_at,
                       "wait_until": format_utc(at + REQUEST_WAIT_SECS),
                       "names_claim": names, "answered_by": answered_by}),
            )
        }
    };

    // 3. Heartbeat.
    let mut activity: Vec<(&str, Value, i64)> = vec![
        ("claim_written", json!(claim.id), written),
        ("claim_renewal", json!(claim.id), renewed),
    ];
    for c in all.values().filter(|c| c.id != claim.id && of_lane(c)) {
        match time("comment", &c.id.to_string(), &c.updated_at) {
            Ok(t) => activity.push(("lane_comment", json!(c.id), t)),
            Err(e) => return uncertain(input, e),
        }
    }
    let token = claim_token(input.issue, input.claim);
    let token_commits: Vec<&BranchCommit> = input
        .commits
        .iter()
        .filter(|m| m.message.lines().any(|l| l.trim() == token))
        .collect();
    for m in &token_commits {
        match time("commit", &m.sha, &m.date) {
            Ok(t) => activity.push(("commit", json!(m.sha), t)),
            Err(e) => return uncertain(input, e),
        }
    }
    let latest = activity
        .iter()
        .max_by_key(|a| a.2)
        .expect("claim times are always present");
    let future = latest.2 > now;
    let silent = now >= latest.2 + HEARTBEAT_SECS;
    let heartbeat_met = !future && silent;
    let heartbeat_json = json!({
        "met": heartbeat_met,
        "reason": if future { "future_activity" } else if silent { "silent" } else { "recent_activity" },
        "latest": {"kind": latest.0, "ref": latest.1, "at": format_utc(latest.2)},
        "heartbeat_until": format_utc(latest.2 + HEARTBEAT_SECS),
    });

    let holder_ok = holder == Some(input.claim);
    let mut reasons = Vec::new();
    if !holder_ok {
        reasons.push("claim_not_holder");
    }
    for (met, name) in [
        (lease_met, "lease_live"),
        (request_met, "request_not_ready"),
        (heartbeat_met, "heartbeat_seen"),
    ] {
        if !met {
            reasons.push(name);
        }
    }
    let met = reasons.is_empty();
    let since = input.repo_since.as_deref().unwrap_or("-");
    let branch_text = branch.unwrap_or("-");
    let read = format!(
        "activity read: issue comments {}, repository comments {} since {since}, commits {} on {branch_text} ({} with `{token}`)",
        input.issue_comments.len(),
        input.repo_comments.len(),
        input.commits.len(),
        token_commits.len()
    );
    let line = met.then(|| render_handoff_line(input.claim, &input.me, "timeout"));
    let body = line.as_ref().map(|line| {
        format!(
            "{line}\n\nclaim-timeout-check (#2906, decision D2): claim comment {claim} (lane {lane}) on #{issue}, judged at {now}.\n\
             - lease: lease_until {lease_until} has passed\n\
             - request comment {request}: no answer from {lane} and no release or handoff of claim {claim} for at least {REQUEST_WAIT_SECS} s\n\
             - heartbeat: none for at least {HEARTBEAT_SECS} s; latest {kind} {reference} at {latest_at}\n\
             - {read}\n",
            claim = input.claim,
            issue = input.issue,
            now = format_utc(now),
            lease_until = parsed.lease_until,
            request = input.request,
            kind = latest.0,
            reference = latest.1,
            latest_at = format_utc(latest.2),
        )
    });
    json!({
        "met": met,
        "state": if met { "met" } else { "not_met" },
        "reasons": reasons,
        "issue": input.issue,
        "claim": input.claim,
        "request": input.request,
        "me": input.me,
        "holder": holder,
        "holder_lane": lane,
        "branch": branch,
        "now": format_utc(now),
        "conditions": {
            "lease_expired": {"met": lease_met, "lease_until": parsed.lease_until},
            "request_unanswered": request_json,
            "no_heartbeat": heartbeat_json,
        },
        "activity_read": {
            "issue_comments": input.issue_comments.len(),
            "repo_comments": input.repo_comments.len(),
            "repo_since": input.repo_since,
            "branch": branch,
            "branch_found": input.branch_found,
            "commits": input.commits.len(),
            "token_commits": token_commits.len(),
        },
        "line": line,
        "body": body,
    })
}

/// What [`check`] reads: the repository, the Issue, the holder's claim, this lane's request.
pub struct Target<'a> {
    pub owner: &'a str,
    pub repo: &'a str,
    pub issue: i64,
    pub claim: i64,
    pub request: i64,
    pub me: &'a str,
}

/// A query value with everything but unreserved characters and `/` percent-encoded.
fn query_value(s: &str) -> String {
    s.bytes()
        .map(|b| {
            if b.is_ascii_alphanumeric() || b"-._~/".contains(&b) {
                (b as char).to_string()
            } else {
                format!("%{b:02X}")
            }
        })
        .collect()
}

/// Every item of a list, following `Link: rel="next"` for at most [`MAX_PAGES`] pages. The
/// error carries the forge error kind (`not_found` for a 404).
fn read_list<T: Transport, C: Clock>(
    client: &Client<'_, T, C>,
    first: String,
) -> std::result::Result<Vec<Value>, (&'static str, String)> {
    let mut url = Some(first);
    let mut items = Vec::new();
    let mut pages = 0;
    while let Some(u) = url {
        pages += 1;
        if pages > MAX_PAGES {
            return Err(("page_cap", format!("page limit exceeded ({MAX_PAGES})")));
        }
        let resp = client.get(&u, None).map_err(|e| (e.kind, e.message))?;
        match resp.body.as_ref().and_then(Value::as_array) {
            Some(list) => items.extend(list.iter().cloned()),
            None => return Err(("invalid", "response is not a list".into())),
        }
        url = resp.header("link").and_then(|l| next_link(Some(l)));
    }
    Ok(items)
}

fn timed(v: &Value) -> Option<TimedComment> {
    let s = |k: &str| v[k].as_str().unwrap_or("").to_string();
    Some(TimedComment {
        id: v["id"].as_i64()?,
        author: v["user"]["login"].as_str().unwrap_or("").to_string(),
        body: v["body"].as_str()?.to_string(),
        created_at: s("created_at"),
        updated_at: s("updated_at"),
    })
}

/// Reads the Issue's comments, the repository's comments of the last [`HEARTBEAT_SECS`] and
/// the claim branch's commits since the claim (GET only), then [`judge`]s at `now`.
pub fn check<T: Transport, C: Clock>(t: &Target<'_>, transport: &T, clock: &C, now: i64) -> Value {
    let client = Client::new(transport, clock, Default::default(), Default::default());
    let base = format!("https://api.github.com/repos/{}/{}", t.owner, t.repo);
    let mut input = Input {
        issue: t.issue,
        claim: t.claim,
        request: t.request,
        me: t.me.to_string(),
        ..Default::default()
    };
    let issue_url = format!("{base}/issues/{}/comments?per_page=100", t.issue);
    match read_list(&client, issue_url) {
        Ok(items) => input.issue_comments = items.iter().filter_map(timed).collect(),
        Err((_, e)) => return uncertain(&input, format!("issue comments unreadable: {e}")),
    }
    let since = format_utc(now - HEARTBEAT_SECS);
    let repo_url = format!("{base}/issues/comments?since={since}&per_page=100");
    match read_list(&client, repo_url) {
        Ok(items) => input.repo_comments = items.iter().filter_map(timed).collect(),
        Err((_, e)) => return uncertain(&input, format!("repository comments unreadable: {e}")),
    }
    input.repo_since = Some(since);
    let branch = input
        .issue_comments
        .iter()
        .find(|c| c.id == t.claim)
        .and_then(|c| {
            let branch = claim_field(&c.body, "branch")?.to_string();
            Some((branch, c.created_at.clone()))
        });
    if let Some((branch, created_at)) = branch {
        let Some(written) = parse_utc(&created_at).map(format_utc) else {
            return uncertain(
                &input,
                format!("unreadable time `{created_at}` on claim comment"),
            );
        };
        let url = format!(
            "{base}/commits?sha={}&since={written}&per_page=100",
            query_value(&branch)
        );
        match read_list(&client, url) {
            Ok(items) => {
                input.branch_found = true;
                input.commits = items
                    .iter()
                    .map(|v| BranchCommit {
                        sha: v["sha"].as_str().unwrap_or("").to_string(),
                        message: v["commit"]["message"].as_str().unwrap_or("").to_string(),
                        date: v["commit"]["committer"]["date"]
                            .as_str()
                            .unwrap_or("")
                            .to_string(),
                    })
                    .collect();
            }
            // Decision Q4: a branch that is gone has no commits.
            Err(("not_found", _)) => input.branch_found = false,
            Err((_, e)) => return uncertain(&input, format!("branch commits unreadable: {e}")),
        }
    }
    judge(&input, now)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::policy::{format_utc, parse_utc};

    const NOW: &str = "2026-11-14T12:00:00Z";
    const TOKEN: &str = "Moai-Claim: 2906#100";
    const LINE: &str = "moai-claim-handoff: comment=100 to=agent-25 by=timeout";

    fn now() -> i64 {
        parse_utc(NOW).unwrap()
    }

    fn at(offset: i64) -> String {
        format_utc(now() + offset)
    }

    fn comment(id: i64, body: &str, created: i64, updated: i64) -> TimedComment {
        TimedComment {
            id,
            author: "GoosLab".into(),
            body: body.into(),
            created_at: at(created),
            updated_at: at(updated),
        }
    }

    fn claim_body(lease: i64) -> String {
        format!(
            "moai-claim: session=agent-7-abc epoch=1 lease_until={} lane=agent-7 issue=2906 branch=issue/2906-x",
            at(lease)
        )
    }

    /// The met case: lease expired 10 s ago, the request is an hour old, the claim was written
    /// three hours ago and nothing happened since.
    fn base() -> Input {
        Input {
            issue: 2906,
            claim: 100,
            request: 200,
            me: "agent-25".into(),
            issue_comments: vec![
                comment(100, &claim_body(-10), -3 * 3600, -3 * 3600),
                comment(
                    200,
                    "agent-25: the holder of claim 100 shows no activity; timeout takeover after 1800 s",
                    -3600,
                    -3600,
                ),
            ],
            repo_comments: vec![],
            repo_since: Some(at(-7200)),
            branch_found: true,
            commits: vec![],
        }
    }

    fn cond(v: &Value, name: &str) -> bool {
        v["conditions"][name]["met"]
            .as_bool()
            .unwrap_or_else(|| panic!("{name} missing: {v}"))
    }

    fn token_commit(sha: &str, token: &str, offset: i64) -> BranchCommit {
        BranchCommit {
            sha: sha.into(),
            message: format!("fix(x): y (#2906)\n\n{token}\n🗿 MoAI"),
            date: at(offset),
        }
    }

    /// AC8: `now >= lease_until`, the same edge as `is_live = lease_until > now`.
    #[test]
    fn lease_expiry_boundary() {
        for (lease, want) in [(1, false), (0, true), (-1, true)] {
            let mut input = base();
            input.issue_comments[0].body = claim_body(lease);
            let v = judge(&input, now());
            assert_eq!(cond(&v, "lease_expired"), want, "lease {lease}: {v}");
            assert_eq!(v["met"], json!(want), "lease {lease}: {v}");
        }
    }

    /// AC9: the request waits 1800 s.
    #[test]
    fn request_wait_boundary() {
        for (age, want) in [(1799, false), (1800, true), (1801, true)] {
            let mut input = base();
            input.issue_comments[1].created_at = at(-age);
            input.issue_comments[1].updated_at = at(-age);
            let v = judge(&input, now());
            assert_eq!(cond(&v, "request_unanswered"), want, "age {age}: {v}");
            assert_eq!(v["met"], json!(want), "age {age}: {v}");
        }
    }

    /// AC10: a lease renewal (the claim comment's `updated_at`) is activity.
    #[test]
    fn heartbeat_boundary_claim_renewal() {
        for (age, want) in [(7199, false), (7200, true), (7201, true)] {
            let mut input = base();
            input.issue_comments[0].updated_at = at(-age);
            let v = judge(&input, now());
            assert_eq!(cond(&v, "no_heartbeat"), want, "age {age}: {v}");
            assert_eq!(v["met"], json!(want), "age {age}: {v}");
        }
    }

    /// AC10: a commit on the claim branch carrying the claim token is activity; a commit
    /// without it, or with another claim's token, is not.
    #[test]
    fn heartbeat_boundary_commit() {
        for (age, want) in [(7199, false), (7200, true), (7201, true)] {
            let mut input = base();
            input.commits = vec![
                token_commit("a1", TOKEN, -age),
                token_commit("a2", "Moai-Claim: 2906#1001", -5),
                BranchCommit {
                    sha: "a3".into(),
                    message: "merge tip".into(),
                    date: at(-5),
                },
            ];
            let v = judge(&input, now());
            assert_eq!(cond(&v, "no_heartbeat"), want, "age {age}: {v}");
            assert_eq!(v["met"], json!(want), "age {age}: {v}");
            assert_eq!(v["activity_read"]["token_commits"], json!(1), "{v}");
        }
    }

    /// AC10: a comment of the holder lane anywhere in the repository is activity; another
    /// lane whose name only starts the same (`agent-70`) is not.
    #[test]
    fn heartbeat_boundary_comment() {
        for (age, want) in [(7199, false), (7200, true), (7201, true)] {
            let mut input = base();
            input.repo_comments = vec![
                comment(150, "agent-7: working on #2901 tests", -age, -age),
                comment(160, "agent-70: started #2902", -5, -5),
            ];
            let v = judge(&input, now());
            assert_eq!(cond(&v, "no_heartbeat"), want, "age {age}: {v}");
            assert_eq!(v["met"], json!(want), "age {age}: {v}");
        }
    }

    /// AC11: a holder-lane comment, or a release closing the claim, after the request answers
    /// it; the same comment before the request does not.
    #[test]
    fn an_answer_after_the_request_blocks() {
        let mut before = base();
        before
            .repo_comments
            .push(comment(150, "agent-7: still here", -7300, -7300));
        let v = judge(&before, now());
        assert!(cond(&v, "request_unanswered"), "{v}");
        assert_eq!(v["met"], json!(true), "{v}");

        let mut answered = base();
        answered.repo_comments.push(comment(
            300,
            "agent-7: still here, pushing soon",
            -1000,
            -1000,
        ));
        let v = judge(&answered, now());
        assert!(!cond(&v, "request_unanswered"), "{v}");
        assert_eq!(
            v["conditions"]["request_unanswered"]["answered_by"],
            json!(300)
        );
        assert_eq!(v["met"], json!(false), "{v}");
        assert!(v["line"].is_null(), "{v}");

        let mut released = base();
        released.issue_comments.push(comment(
            301,
            "moai-claim-release: comment=100 by=leader",
            -1000,
            -1000,
        ));
        let v = judge(&released, now());
        assert_eq!(
            v["conditions"]["request_unanswered"]["answered_by"],
            json!(301),
            "{v}"
        );
        assert_eq!(v["met"], json!(false), "{v}");
    }

    /// AC11: a request that does not name the holder's claim id never starts the wait.
    #[test]
    fn a_request_naming_no_claim_blocks() {
        for body in [
            "agent-25: agent-7 seems gone",
            "agent-25: claim 1000 seems gone",
        ] {
            let mut input = base();
            input.issue_comments[1].body = body.into();
            let v = judge(&input, now());
            assert!(!cond(&v, "request_unanswered"), "{body}: {v}");
            assert_eq!(
                v["conditions"]["request_unanswered"]["reason"],
                json!("request_names_no_claim"),
                "{body}"
            );
            assert_eq!(v["met"], json!(false), "{body}");
        }
    }

    fn request_reason(v: &Value) -> &Value {
        &v["conditions"]["request_unanswered"]["reason"]
    }

    /// B1 (audit P1): the holder's own note passed as `--request` is not this lane's request,
    /// and it stays the holder's activity.
    #[test]
    fn the_holders_note_is_not_a_request() {
        for body in [
            "resume note - agent-7 (claim `2906#100`, will renew)",
            "agent-25 to agent-7: claim 100 shows no activity; timeout takeover after 1800 s",
        ] {
            let mut input = base();
            input.issue_comments[1].body = body.into();
            let v = judge(&input, now());
            assert_eq!(v["met"], json!(false), "{body}: {v}");
            assert_eq!(
                request_reason(&v),
                &json!("request_not_mine"),
                "{body}: {v}"
            );
            assert!(!cond(&v, "no_heartbeat"), "{body}: {v}");
            assert_eq!(
                v["conditions"]["no_heartbeat"]["latest"]["ref"],
                json!(200),
                "{body}: {v}"
            );
        }
    }

    /// B1 (audit P3): a request another lane wrote is not this lane's; `agent-250` is not
    /// `agent-25`.
    #[test]
    fn another_lanes_request_is_not_mine() {
        for body in [
            "agent-30: claim 100 shows no activity; timeout takeover after 1800 s",
            "agent-250: claim 100 shows no activity; timeout takeover after 1800 s",
            "claim 100 shows no activity\nagent-25",
        ] {
            let mut input = base();
            input.issue_comments[1].body = body.into();
            let v = judge(&input, now());
            assert_eq!(v["met"], json!(false), "{body}: {v}");
            assert_eq!(
                request_reason(&v),
                &json!("request_not_mine"),
                "{body}: {v}"
            );
        }
    }

    /// B1: a request written before the claim (a lower id) is not a request about it.
    #[test]
    fn a_request_before_the_claim_blocks() {
        let mut input = base();
        input.issue_comments[1].id = 50;
        input.request = 50;
        let v = judge(&input, now());
        assert_eq!(v["met"], json!(false), "{v}");
        assert_eq!(request_reason(&v), &json!("request_before_claim"), "{v}");
    }

    /// Holder-lane comments in the forms lanes write, the lane not on the first line.
    const HOLDER_FORMS: [&str; 6] = [
        "AUDIT #2884 (x) head a1 base b2 — verdict: PASS\n\nauditor: agent-7",
        "CLOSE-TABLE #2876 — all rows closed\nlane=agent-7",
        "DECISION-APPLIED #2880\n\n감사: agent-7 (claim 6873011928)",
        "pr-wait: #2890 waits for CI\nsession=agent-7-abc",
        "LANE-DONE agent-7 #2901",
        "moai-audit-claim: lane=agent-7 pr=#2903",
    ];

    /// B2 (audit P2): a holder-lane comment ten minutes old is a heartbeat wherever in the body
    /// the lane or session stands.
    #[test]
    fn a_holder_comment_anywhere_in_the_body_is_a_heartbeat() {
        for body in HOLDER_FORMS {
            let mut input = base();
            input.repo_comments = vec![comment(150, body, -600, -600)];
            let v = judge(&input, now());
            assert!(!cond(&v, "no_heartbeat"), "{body}: {v}");
            assert_eq!(
                v["conditions"]["no_heartbeat"]["latest"]["ref"],
                json!(150),
                "{body}: {v}"
            );
            assert_eq!(v["met"], json!(false), "{body}: {v}");
        }
    }

    /// B2: the same forms after the request answer it.
    #[test]
    fn a_holder_comment_anywhere_in_the_body_answers() {
        for body in HOLDER_FORMS {
            let mut input = base();
            input.repo_comments = vec![comment(300, body, -3000, -3000)];
            let v = judge(&input, now());
            assert_eq!(
                v["conditions"]["request_unanswered"]["answered_by"],
                json!(300),
                "{body}: {v}"
            );
            assert_eq!(request_reason(&v), &json!("answered"), "{body}: {v}");
            assert_eq!(v["met"], json!(false), "{body}: {v}");
        }
    }

    /// N1: a claim already closed before the request (release, handoff, D1 renewal by a later
    /// claim of the same session) is no longer the holder; only `holder_ok` refuses it.
    #[test]
    fn a_claim_closed_before_the_request_blocks() {
        for (closer, lease) in [
            ("moai-claim-release: comment=100 by=leader".to_string(), 0),
            (
                "moai-claim-handoff: comment=100 to=agent-9-x".to_string(),
                0,
            ),
            (
                "moai-claim: session=agent-7-abc epoch=2 lease_until={} lane=agent-7 issue=2906"
                    .to_string(),
                -5,
            ),
        ] {
            let closer = closer.replace("{}", &at(lease));
            let mut input = base();
            input
                .issue_comments
                .insert(1, comment(150, &closer, -7300, -7300));
            let v = judge(&input, now());
            assert_eq!(v["met"], json!(false), "{closer}: {v}");
            assert_eq!(v["reasons"], json!(["claim_not_holder"]), "{closer}: {v}");
            assert!(v["line"].is_null(), "{closer}: {v}");
        }
    }

    /// AC11: an activity time after `now` (a skewed clock) never counts as silence.
    #[test]
    fn a_future_activity_blocks() {
        let mut commit = base();
        commit.commits = vec![token_commit("a1", TOKEN, 60)];
        let mut renewal = base();
        renewal.issue_comments[0].updated_at = at(60);
        for input in [commit, renewal] {
            let v = judge(&input, now());
            assert!(!cond(&v, "no_heartbeat"), "{v}");
            assert_eq!(
                v["conditions"]["no_heartbeat"]["reason"],
                json!("future_activity"),
                "{v}"
            );
            assert_eq!(v["met"], json!(false), "{v}");
        }
    }

    /// AC13: only when all three hold, the exact `by=timeout` line and an evidence body that
    /// names the request and the activity read.
    #[test]
    fn met_renders_the_timeout_line() {
        let v = judge(&base(), now());
        assert_eq!(v["met"], json!(true), "{v}");
        assert_eq!(v["state"], json!("met"));
        assert_eq!(v["line"], json!(LINE));
        let body = v["body"].as_str().unwrap();
        assert!(body.starts_with(&format!("{LINE}\n")), "{body}");
        assert!(body.contains("request comment 200"), "{body}");
        assert!(
            body.contains("activity read: issue comments 2, repository comments 0"),
            "{body}"
        );
        assert_eq!(v["holder_lane"], json!("agent-7"));
        assert_eq!(v["branch"], json!("issue/2906-x"));
        // The rendered line closes this claim the way claim-verify reads it.
        assert_eq!(
            crate::claim_verify::classify_comment(LINE),
            crate::claim_verify::CommentKind::Handoff {
                target: 100,
                to: "agent-25".into()
            }
        );
    }
}
