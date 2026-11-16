//! `card-times` (#3067 S6): when each stage of a card and of its pull requests happened, read
//! only from the server.
//!
//! Every stage is a `{state, at, at_epoch, source, comment_id?}` value whose `state` is one of
//! [`OBSERVED`], [`ABSENT`] and [`UNKNOWN`]:
//!
//! - `observed`: a server time was read (a comment's `createdAt`, a pull request's `createdAt` or
//!   `mergedAt`, a timeline event, a check suite). A commit date is also shown, under the source
//!   `commit_date`, because the repair rebuild needs the next head commit.
//! - `absent`: the read finished and holds no such record (not merged, not started, on hold).
//! - `unknown`: the read was incomplete (comments or commits cut at the page size, no request
//!   left in the budget), or the value cannot be trusted (a time in the future, a malformed time).
//!
//! A span length is shown only when both ends are `observed`; an end earlier than its start makes
//! the span `unknown`. The time a `pr-wait:` line states in its own text is kept apart as
//! `stated_at` and never enters a span. Nothing here writes to GitHub or to disk.

use crate::forge::policy::format_utc;
use crate::lane_next::{audit_claim, body_claim, strict_utc, verdict, Comment};
use serde_json::{json, Map, Value};

/// The three stage words. One place: #3068 reads them later, and the leader may rename them.
pub const OBSERVED: &str = "observed";
pub const ABSENT: &str = "absent";
pub const UNKNOWN: &str = "unknown";

/// One stage of a card or a pull request.
#[derive(Clone, Debug, PartialEq)]
pub struct Stage {
    pub state: &'static str,
    /// Seconds since the epoch; `Some` only when `state` is [`OBSERVED`].
    pub at: Option<i64>,
    pub source: String,
    pub comment_id: Option<u64>,
    pub reason: Option<String>,
    pub extra: Map<String, Value>,
}

impl Stage {
    /// A server time read as `at`. A time in the future of `now`, or one that is not
    /// `YYYY-MM-DDTHH:MM:SSZ`, is [`UNKNOWN`]: a clock or a format that cannot be trusted.
    pub fn observed(source: &str, at: &str, now: i64, comment_id: Option<u64>) -> Stage {
        match strict_utc(at) {
            None => Stage::unknown(source, "time is not YYYY-MM-DDTHH:MM:SSZ"),
            Some(t) if t > now => Stage::unknown(source, "time is in the future of this read"),
            Some(t) => Stage {
                state: OBSERVED,
                at: Some(t),
                source: source.to_string(),
                comment_id,
                reason: None,
                extra: Map::new(),
            },
        }
        .with_comment(comment_id)
    }

    pub fn absent(source: &str, reason: &str) -> Stage {
        Stage::bare(ABSENT, source, reason)
    }

    pub fn unknown(source: &str, reason: &str) -> Stage {
        Stage::bare(UNKNOWN, source, reason)
    }

    fn bare(state: &'static str, source: &str, reason: &str) -> Stage {
        Stage {
            state,
            at: None,
            source: source.to_string(),
            comment_id: None,
            reason: Some(reason.to_string()),
            extra: Map::new(),
        }
    }

    /// Keeps the comment id on a stage that turned unknown, so a person can find the comment.
    fn with_comment(mut self, id: Option<u64>) -> Stage {
        self.comment_id = id;
        self
    }

    /// Adds one named field to the stage's JSON (`result`, `head`, `stated_at`).
    pub fn with(mut self, key: &str, value: Value) -> Stage {
        self.extra.insert(key.to_string(), value);
        self
    }

    pub fn is_observed(&self) -> bool {
        self.state == OBSERVED
    }

    pub fn to_json(&self) -> Value {
        let mut v = json!({
            "state": self.state,
            "at": self.at.map(format_utc),
            "at_epoch": self.at,
            "source": self.source,
        });
        if let Some(id) = self.comment_id {
            v["comment_id"] = json!(id);
        }
        if let Some(reason) = &self.reason {
            v["reason"] = json!(reason);
        }
        for (k, x) in &self.extra {
            v[k] = x.clone();
        }
        v
    }
}

/// The length between two stages. `secs` appears only when both ends are observed and the end is
/// not before the start; any unknown end, or a negative length, makes it unknown; otherwise an
/// absent end makes it absent.
pub fn span(from: &Stage, to: &Stage) -> Value {
    match (from.at, to.at) {
        (Some(a), Some(b)) if b >= a => json!({"state": OBSERVED, "secs": b - a}),
        (Some(_), Some(_)) => {
            json!({"state": UNKNOWN, "reason": "the end is earlier than the start"})
        }
        _ if from.state == UNKNOWN || to.state == UNKNOWN => {
            json!({"state": UNKNOWN, "reason": "an end of the span is unknown"})
        }
        _ => json!({"state": ABSENT, "reason": "an end of the span is absent"}),
    }
}

/// One pull request as read.
#[derive(Clone, Debug, Default)]
pub struct PrRead {
    pub number: u64,
    /// `OPEN`, `CLOSED` or `MERGED`.
    pub state: String,
    pub draft: bool,
    pub merged: bool,
    pub created_at: String,
    pub merged_at: Option<String>,
    pub head: String,
    pub branch: String,
    pub body: String,
    /// Comments a repository member wrote, oldest first.
    pub comments: Vec<Comment>,
    pub comments_complete: bool,
    pub ready_events: Vec<String>,
    pub convert_events: Vec<String>,
    pub events_complete: bool,
    /// `(oid, committedDate)`, oldest first.
    pub commits: Vec<(String, String)>,
    pub commits_complete: bool,
    pub check_suites: Vec<String>,
    pub check_runs: Vec<String>,
    pub checks_complete: bool,
}

/// One card as read.
#[derive(Clone, Debug, Default)]
pub struct CardRead {
    pub number: u64,
    /// Comments a repository member wrote, oldest first.
    pub comments: Vec<Comment>,
    pub comments_complete: bool,
}

/// Does `pr` belong to card `card`: its body names `Moai-Claim: <card>#…`, or its branch is
/// `issue/<card>-…`.
pub fn claims_pr(card: u64, pr: &PrRead) -> bool {
    body_claim(&pr.body).is_some_and(|(issue, _)| issue == card)
        || pr.branch.starts_with(&format!("issue/{card}-"))
}

/// The earliest claim of the card. A card whose newest comments only were read cannot say which
/// claim is earliest, so it is unknown.
pub fn claim_stage(card: &CardRead, now: i64) -> Stage {
    const SOURCE: &str = "issue_comment";
    if !card.comments_complete {
        return Stage::unknown(
            SOURCE,
            "older comments were not read: the earliest claim may be there",
        );
    }
    match card
        .comments
        .iter()
        .find(|c| c.body.lines().any(|l| l.trim().starts_with("moai-claim:")))
    {
        Some(c) => Stage::observed(SOURCE, &c.at, now, Some(c.id)),
        None => Stage::absent(SOURCE, "no claim comment by a member"),
    }
}

/// The card's `pr-wait:` lines: their server times, and the card-level word (`observed`,
/// `absent` when the read finished with none, `unknown` when comments were cut).
pub fn pr_wait(card: &CardRead, now: i64) -> (&'static str, Vec<Value>) {
    let list: Vec<Value> = card
        .comments
        .iter()
        .filter(|c| {
            c.body
                .lines()
                .next()
                .is_some_and(|l| l.trim().starts_with("pr-wait:"))
        })
        .map(|c| {
            let stated = c
                .body
                .lines()
                .next()
                .unwrap_or("")
                .split_whitespace()
                .find(|w| strict_utc(w).is_some());
            Stage::observed("issue_comment", &c.at, now, Some(c.id))
                .with("stated_at", json!(stated))
                .to_json()
        })
        .collect();
    let state = if !list.is_empty() {
        OBSERVED
    } else if card.comments_complete {
        ABSENT
    } else {
        UNKNOWN
    };
    (state, list)
}

/// The earliest of `times`, or the word for none: absent when the read finished, unknown when it
/// was cut or a time is malformed.
fn earliest(source: &str, times: &[String], complete: bool, none: &str, now: i64) -> Stage {
    if !complete {
        return Stage::unknown(source, "the list was cut: the earliest may be missing");
    }
    let mut best: Option<&String> = None;
    for t in times {
        let Some(e) = strict_utc(t) else {
            return Stage::unknown(source, "a time is not YYYY-MM-DDTHH:MM:SSZ");
        };
        if best.is_none_or(|b| strict_utc(b).is_some_and(|x| e < x)) {
            best = Some(t);
        }
    }
    match best {
        Some(t) => Stage::observed(source, t, now, None),
        None => Stage::absent(source, none),
    }
}

fn ready_stage(pr: &PrRead, now: i64) -> Stage {
    const SOURCE: &str = "ready_for_review_event";
    if !pr.ready_events.is_empty() {
        return earliest(SOURCE, &pr.ready_events, true, "", now);
    }
    if !pr.events_complete {
        return Stage::unknown(SOURCE, "timeline events were cut");
    }
    if pr.draft {
        return Stage::absent(SOURCE, "still a draft");
    }
    if pr.convert_events.is_empty() {
        // Opened ready: it was ready when it was created.
        return Stage::observed("pr_created_at", &pr.created_at, now, None);
    }
    Stage::unknown(SOURCE, "converted to draft and no ready event was read")
}

fn merged_stage(pr: &PrRead, now: i64) -> Stage {
    const SOURCE: &str = "pr_merged_at";
    if !pr.merged {
        return Stage::absent(SOURCE, "not merged");
    }
    match &pr.merged_at {
        Some(at) => Stage::observed(SOURCE, at, now, None),
        None => Stage::unknown(SOURCE, "merged, but no mergedAt was read"),
    }
}

fn audit_reserved_stage(pr: &PrRead, now: i64) -> Stage {
    const SOURCE: &str = "pr_comment";
    if !pr.comments_complete {
        return Stage::unknown(
            SOURCE,
            "older comments were not read: the earliest reservation may be there",
        );
    }
    match pr
        .comments
        .iter()
        .find(|c| audit_claim(c).is_some_and(|(_, head)| head == pr.head))
    {
        Some(c) => Stage::observed(SOURCE, &c.at, now, Some(c.id)).with("head", json!(pr.head)),
        None => Stage::absent(SOURCE, "no audit reservation for the current head"),
    }
}

/// A verdict is read from the newest comments, so a found verdict stands even when older
/// comments were not read; only "none found" depends on the read being complete.
fn verdict_stage(pr: &PrRead, now: i64) -> Stage {
    const SOURCE: &str = "pr_comment";
    let hit = pr.comments.iter().rev().find_map(|c| {
        let (head, result) = verdict(c)?;
        (head == pr.head).then_some((c, result))
    });
    match hit {
        Some((c, result)) => Stage::observed(SOURCE, &c.at, now, Some(c.id))
            .with("result", json!(result))
            .with("head", json!(pr.head)),
        None if pr.comments_complete => Stage::absent(SOURCE, "no verdict for the current head"),
        None => Stage::unknown(SOURCE, "comments were cut"),
    }
}

/// All stages and spans of one pull request.
pub fn pr_json(pr: &PrRead, now: i64) -> Value {
    let opened = Stage::observed("pr_created_at", &pr.created_at, now, None);
    let ready = ready_stage(pr, now);
    let ci_queued = earliest(
        "check_suite_created_at",
        &pr.check_suites,
        pr.checks_complete,
        "no check suite on the head commit",
        now,
    );
    let ci_running = earliest(
        "check_run_started_at",
        &pr.check_runs,
        pr.checks_complete,
        "no check run started on the head commit",
        now,
    );
    let reserved = audit_reserved_stage(pr, now);
    let verdict = verdict_stage(pr, now);
    let merged = merged_stage(pr, now);
    json!({
        "number": pr.number,
        "state": pr.state.to_ascii_lowercase(),
        "draft": pr.draft,
        "head": pr.head,
        "opened": opened.to_json(),
        "ready": ready.to_json(),
        "ci_queued": ci_queued.to_json(),
        "ci_running": ci_running.to_json(),
        "audit_reserved": reserved.to_json(),
        "verdict": verdict.to_json(),
        "repairs": repairs(pr, now),
        "repairs_complete": pr.comments_complete && pr.commits_complete,
        "merged": merged.to_json(),
        "spans": {
            "opened_to_ready": span(&opened, &ready),
            "ready_to_audit_reserved": span(&ready, &reserved),
            "audit_reserved_to_verdict": span(&reserved, &verdict),
            "verdict_to_merged": span(&verdict, &merged),
            "opened_to_merged": span(&opened, &merged),
        },
    })
}

/// Repairs rebuilt from the thread: no `REPAIR` line exists, so each `AUDIT … FAIL` of a head is
/// followed to the next head commit and then to the next verdict of another head.
pub fn repairs(pr: &PrRead, now: i64) -> Vec<Value> {
    let mut out = Vec::new();
    for (i, c) in pr.comments.iter().enumerate() {
        let Some((failed, "FAIL")) = verdict(c) else {
            continue;
        };
        let fail =
            Stage::observed("pr_comment", &c.at, now, Some(c.id)).with("head", json!(failed));
        let commit = match pr.commits.iter().position(|(oid, _)| *oid == failed) {
            None => Stage::unknown(
                "commit_date",
                "the failed head is not among the commits read",
            ),
            Some(p) => match pr.commits.get(p + 1) {
                Some((_, date)) => Stage::observed("commit_date", date, now, None),
                None if pr.commits_complete => {
                    Stage::absent("commit_date", "no commit after the failed head")
                }
                None => Stage::unknown("commit_date", "commits were cut"),
            },
        };
        let next = pr.comments[i + 1..].iter().find_map(|d| {
            let (head, result) = verdict(d)?;
            (head != failed).then_some((d, head, result))
        });
        let next_verdict = match next {
            Some((d, head, result)) => Stage::observed("pr_comment", &d.at, now, Some(d.id))
                .with("result", json!(result))
                .with("head", json!(head)),
            None if pr.comments_complete => {
                Stage::absent("pr_comment", "no later verdict of another head")
            }
            None => Stage::unknown("pr_comment", "comments were cut"),
        };
        out.push(json!({
            "failed_head": failed,
            "fail": fail.to_json(),
            "commit": commit.to_json(),
            "next_verdict": next_verdict.to_json(),
            "spans": {
                "fail_to_commit": span(&fail, &commit),
                "commit_to_next_verdict": span(&commit, &next_verdict),
                "fail_to_next_verdict": span(&fail, &next_verdict),
            },
        }));
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::forge::policy::parse_utc;

    const NOW: &str = "2026-11-16T12:00:00Z";

    fn now() -> i64 {
        parse_utc(NOW).unwrap()
    }

    fn head(c: char) -> String {
        c.to_string().repeat(40)
    }

    fn comment(id: u64, at: &str, body: &str) -> Comment {
        Comment {
            id,
            body: body.to_string(),
            at: at.to_string(),
        }
    }

    fn pr() -> PrRead {
        PrRead {
            number: 7,
            state: "OPEN".into(),
            created_at: "2026-11-16T08:00:00Z".into(),
            head: head('b'),
            branch: "issue/3067-x".into(),
            comments_complete: true,
            events_complete: true,
            commits_complete: true,
            checks_complete: true,
            ..PrRead::default()
        }
    }

    fn j(v: &Value, path: &[&str]) -> Value {
        path.iter().fold(v.clone(), |v, k| v[*k].clone())
    }

    #[test]
    fn an_observed_stage_carries_the_server_time_and_source() {
        let s = Stage::observed("issue_comment", "2026-11-16T08:00:00Z", now(), Some(5));
        let v = s.to_json();
        assert_eq!(v["state"], OBSERVED);
        assert_eq!(v["at"], "2026-11-16T08:00:00Z");
        assert_eq!(v["at_epoch"], parse_utc("2026-11-16T08:00:00Z").unwrap());
        assert_eq!(v["source"], "issue_comment");
        assert_eq!(v["comment_id"], 5);
    }

    #[test]
    fn a_time_after_now_is_unknown_and_now_itself_is_observed() {
        assert_eq!(Stage::observed("s", NOW, now(), None).state, OBSERVED);
        let later = format_utc(now() + 1);
        let s = Stage::observed("s", &later, now(), None);
        assert_eq!(s.state, UNKNOWN);
        assert_eq!(s.at, None);
        assert!(s.reason.as_deref().unwrap().contains("future"), "{s:?}");
    }

    #[test]
    fn a_malformed_time_is_unknown() {
        for bad in [
            "",
            "2026-11-16",
            "2026-11-16 08:00:00",
            "2026-11-16T08:00:00+09:00",
        ] {
            let s = Stage::observed("s", bad, now(), None);
            assert_eq!(s.state, UNKNOWN, "{bad:?}");
            assert_eq!(s.to_json()["at_epoch"], Value::Null, "{bad:?}");
        }
    }

    #[test]
    fn absent_and_unknown_are_kept_apart_in_the_json() {
        let a = Stage::absent("s", "no record").to_json();
        let u = Stage::unknown("s", "cut").to_json();
        assert_eq!(
            (a["state"].as_str(), a["reason"].as_str()),
            (Some(ABSENT), Some("no record"))
        );
        assert_eq!(
            (u["state"].as_str(), u["reason"].as_str()),
            (Some(UNKNOWN), Some("cut"))
        );
        assert_eq!(a["at_epoch"], Value::Null);
    }

    #[test]
    fn a_span_has_secs_only_when_both_ends_are_observed() {
        let from = Stage::observed("s", "2026-11-16T08:00:00Z", now(), None);
        let to = Stage::observed("s", "2026-11-16T08:05:00Z", now(), None);
        assert_eq!(j(&span(&from, &to), &["secs"]), 300);
        assert_eq!(j(&span(&from, &to), &["state"]), OBSERVED);
        let gone = Stage::absent("s", "none");
        let cut = Stage::unknown("s", "cut");
        assert_eq!(span(&from, &gone)["state"], ABSENT);
        assert_eq!(span(&gone, &to)["state"], ABSENT);
        assert_eq!(span(&from, &cut)["state"], UNKNOWN);
        assert_eq!(
            span(&cut, &gone)["state"],
            UNKNOWN,
            "unknown wins over absent"
        );
        for v in [span(&from, &gone), span(&from, &cut)] {
            assert!(v.get("secs").is_none(), "{v}");
        }
    }

    #[test]
    fn a_span_boundary_zero_is_observed_and_minus_one_is_unknown() {
        let t = Stage::observed("s", "2026-11-16T08:00:00Z", now(), None);
        let same = span(&t, &t);
        assert_eq!(
            (same["state"].as_str(), same["secs"].as_i64()),
            (Some(OBSERVED), Some(0))
        );
        let earlier = Stage::observed("s", "2026-11-16T07:59:59Z", now(), None);
        let neg = span(&t, &earlier);
        assert_eq!(neg["state"], UNKNOWN);
        assert!(neg.get("secs").is_none(), "{neg}");
    }

    #[test]
    fn merged_is_absent_unmerged_unknown_when_untrustworthy() {
        let mut p = pr();
        assert_eq!(pr_json(&p, now())["merged"]["state"], ABSENT);
        p.merged = true;
        p.state = "MERGED".into();
        p.merged_at = Some("2026-11-16T09:00:00Z".into());
        let v = pr_json(&p, now());
        assert_eq!(v["merged"]["state"], OBSERVED);
        assert_eq!(v["merged"]["source"], "pr_merged_at");
        p.merged_at = Some(format_utc(now() + 3600));
        assert_eq!(
            pr_json(&p, now())["merged"]["state"],
            UNKNOWN,
            "future mergedAt"
        );
        p.merged_at = None;
        assert_eq!(
            pr_json(&p, now())["merged"]["state"],
            UNKNOWN,
            "merged without a time"
        );
    }

    #[test]
    fn ready_comes_from_the_event_or_from_a_pr_opened_ready() {
        let mut p = pr();
        assert_eq!(pr_json(&p, now())["ready"]["source"], "pr_created_at");
        p.draft = true;
        assert_eq!(
            pr_json(&p, now())["ready"]["state"],
            ABSENT,
            "still a draft"
        );
        p.ready_events = vec!["2026-11-16T09:00:00Z".into()];
        p.draft = false;
        let v = pr_json(&p, now());
        assert_eq!(v["ready"]["source"], "ready_for_review_event");
        assert_eq!(v["ready"]["at"], "2026-11-16T09:00:00Z");
        p.ready_events.clear();
        p.convert_events = vec!["2026-11-16T08:30:00Z".into()];
        assert_eq!(
            pr_json(&p, now())["ready"]["state"],
            UNKNOWN,
            "converted, no ready event"
        );
        p.events_complete = false;
        p.convert_events.clear();
        assert_eq!(pr_json(&p, now())["ready"]["state"], UNKNOWN, "events cut");
    }

    #[test]
    fn ci_stages_are_the_earliest_suite_and_run() {
        let mut p = pr();
        assert_eq!(pr_json(&p, now())["ci_queued"]["state"], ABSENT);
        assert_eq!(pr_json(&p, now())["ci_running"]["state"], ABSENT);
        p.check_suites = vec!["2026-11-16T08:10:00Z".into(), "2026-11-16T08:02:00Z".into()];
        p.check_runs = vec!["2026-11-16T08:12:00Z".into(), "2026-11-16T08:03:00Z".into()];
        let v = pr_json(&p, now());
        assert_eq!(v["ci_queued"]["at"], "2026-11-16T08:02:00Z");
        assert_eq!(v["ci_running"]["at"], "2026-11-16T08:03:00Z");
        p.checks_complete = false;
        p.check_suites.clear();
        assert_eq!(pr_json(&p, now())["ci_queued"]["state"], UNKNOWN);
    }

    #[test]
    fn audit_stages_follow_the_current_head_only() {
        let mut p = pr();
        let old = head('a');
        p.comments = vec![
            comment(
                1,
                "2026-11-16T08:10:00Z",
                &format!("moai-audit-claim: lane=agent-2 head={old}"),
            ),
            comment(
                2,
                "2026-11-16T08:20:00Z",
                &format!("AUDIT #7 head {old} — verdict: FAIL"),
            ),
            comment(
                3,
                "2026-11-16T09:10:00Z",
                &format!("moai-audit-claim: lane=agent-2 head={}", p.head),
            ),
            comment(
                4,
                "2026-11-16T09:30:00Z",
                &format!("AUDIT #7 head {} — verdict: PASS", p.head),
            ),
        ];
        let v = pr_json(&p, now());
        assert_eq!(v["audit_reserved"]["comment_id"], 3);
        assert_eq!(v["audit_reserved"]["at"], "2026-11-16T09:10:00Z");
        assert_eq!(v["verdict"]["comment_id"], 4);
        assert_eq!(v["verdict"]["result"], "PASS");
        assert_eq!(v["spans"]["audit_reserved_to_verdict"]["secs"], 1200);
        p.comments.truncate(2);
        let v = pr_json(&p, now());
        assert_eq!(
            v["audit_reserved"]["state"], ABSENT,
            "the old head's lines do not count"
        );
        assert_eq!(v["verdict"]["state"], ABSENT);
    }

    #[test]
    fn cut_pr_comments_make_the_earliest_stage_unknown_but_a_found_verdict_stands() {
        let mut p = pr();
        p.comments_complete = false;
        let v = pr_json(&p, now());
        assert_eq!(v["audit_reserved"]["state"], UNKNOWN);
        assert_eq!(v["verdict"]["state"], UNKNOWN);
        p.comments = vec![comment(
            9,
            "2026-11-16T09:30:00Z",
            &format!("AUDIT #7 head {} — verdict: FAIL", p.head),
        )];
        let v = pr_json(&p, now());
        assert_eq!(
            v["verdict"]["state"], OBSERVED,
            "the newest comments hold the newest verdict"
        );
        assert_eq!(v["verdict"]["result"], "FAIL");
        assert_eq!(
            v["audit_reserved"]["state"], UNKNOWN,
            "the earliest one may be older"
        );
    }

    fn repair_pr() -> PrRead {
        let mut p = pr();
        p.head = head('c');
        p.commits = vec![
            (head('a'), "2026-11-16T07:00:00Z".into()),
            (head('b'), "2026-11-16T09:00:00Z".into()),
            (head('c'), "2026-11-16T10:00:00Z".into()),
        ];
        p.comments = vec![
            comment(
                1,
                "2026-11-16T08:30:00Z",
                &format!("AUDIT #7 head {} — verdict: FAIL", head('a')),
            ),
            comment(
                2,
                "2026-11-16T09:30:00Z",
                &format!("AUDIT #7 head {} — verdict: FAIL", head('b')),
            ),
            comment(
                3,
                "2026-11-16T10:30:00Z",
                &format!("AUDIT #7 head {} — verdict: PASS", head('c')),
            ),
        ];
        p
    }

    #[test]
    fn a_repair_is_rebuilt_from_fail_then_next_commit_then_next_verdict() {
        let r = repairs(&repair_pr(), now());
        assert_eq!(r.len(), 2);
        assert_eq!(r[0]["failed_head"], head('a'));
        assert_eq!(r[0]["fail"]["comment_id"], 1);
        assert_eq!(r[0]["commit"]["source"], "commit_date");
        assert_eq!(r[0]["commit"]["at"], "2026-11-16T09:00:00Z");
        assert_eq!(r[0]["next_verdict"]["comment_id"], 2);
        assert_eq!(r[0]["next_verdict"]["result"], "FAIL");
        assert_eq!(r[0]["spans"]["fail_to_next_verdict"]["secs"], 3600);
        assert_eq!(r[1]["failed_head"], head('b'));
        assert_eq!(r[1]["next_verdict"]["result"], "PASS");
        assert_eq!(r[1]["spans"]["commit_to_next_verdict"]["secs"], 1800);
        assert_eq!(r[1]["spans"]["fail_to_commit"]["secs"], 1800);
    }

    #[test]
    fn a_repair_with_a_commit_before_the_fail_has_an_unknown_span_not_a_negative_one() {
        let mut p = repair_pr();
        p.commits[1].1 = "2026-11-16T08:00:00Z".into();
        let r = repairs(&p, now());
        assert_eq!(r[0]["commit"]["state"], OBSERVED);
        assert_eq!(r[0]["spans"]["fail_to_commit"]["state"], UNKNOWN);
        assert!(r[0]["spans"]["fail_to_commit"].get("secs").is_none());
    }

    #[test]
    fn a_repair_without_a_next_commit_or_verdict_is_absent_and_a_cut_read_is_unknown() {
        let mut p = repair_pr();
        p.comments.truncate(2);
        p.commits.truncate(2);
        let r = repairs(&p, now());
        assert_eq!(r.len(), 2);
        assert_eq!(r[1]["failed_head"], head('b'));
        assert_eq!(
            r[1]["commit"]["state"], ABSENT,
            "no commit after the failed head"
        );
        assert_eq!(r[1]["next_verdict"]["state"], ABSENT);
        assert_eq!(r[1]["spans"]["fail_to_next_verdict"]["state"], ABSENT);
        p.commits = vec![(head('z'), "2026-11-16T07:00:00Z".into())];
        let r = repairs(&p, now());
        assert_eq!(
            r[0]["commit"]["state"], UNKNOWN,
            "the failed head is not among the commits read"
        );
        p.comments_complete = false;
        assert_eq!(repairs(&p, now())[1]["next_verdict"]["state"], UNKNOWN);
    }

    #[test]
    fn a_pass_is_no_repair() {
        let mut p = pr();
        p.comments = vec![comment(
            1,
            "2026-11-16T08:30:00Z",
            &format!("AUDIT #7 head {} — verdict: PASS", p.head),
        )];
        assert!(repairs(&p, now()).is_empty());
    }

    #[test]
    fn claim_is_the_earliest_member_claim_and_unknown_when_comments_are_cut() {
        let mut card = CardRead {
            number: 3067,
            comments_complete: true,
            ..CardRead::default()
        };
        assert_eq!(claim_stage(&card, now()).state, ABSENT);
        card.comments = vec![
            comment(
                5,
                "2026-11-16T08:00:00Z",
                "moai-claim: session=agent-10 epoch=1",
            ),
            comment(
                6,
                "2026-11-16T08:30:00Z",
                "moai-claim: session=agent-11 epoch=1",
            ),
            comment(7, "2026-11-16T08:40:00Z", "moai-claim-release: comment=5"),
        ];
        let s = claim_stage(&card, now());
        assert_eq!((s.state, s.comment_id), (OBSERVED, Some(5)));
        card.comments_complete = false;
        assert_eq!(claim_stage(&card, now()).state, UNKNOWN);
    }

    #[test]
    fn pr_wait_keeps_the_stated_time_apart_from_the_server_time() {
        let mut card = CardRead {
            number: 3067,
            comments_complete: true,
            ..CardRead::default()
        };
        assert_eq!(pr_wait(&card, now()), (ABSENT, vec![]));
        card.comments = vec![
            comment(
                8,
                "2026-11-16T09:00:00Z",
                "pr-wait: 2026-11-16T01:00:00Z waits for the open-PR cap",
            ),
            comment(9, "2026-11-16T09:30:00Z", "pr-wait: waits for CI"),
        ];
        let (state, list) = pr_wait(&card, now());
        assert_eq!(state, OBSERVED);
        assert_eq!(list.len(), 2);
        assert_eq!(
            list[0]["at"], "2026-11-16T09:00:00Z",
            "the comment's server time"
        );
        assert_eq!(list[0]["stated_at"], "2026-11-16T01:00:00Z");
        assert_eq!(list[1]["stated_at"], Value::Null);
        assert_eq!(list[0]["source"], "issue_comment");
        card.comments_complete = false;
        assert_eq!(
            pr_wait(&card, now()).0,
            OBSERVED,
            "what was read is still shown"
        );
        card.comments.clear();
        assert_eq!(pr_wait(&card, now()).0, UNKNOWN);
    }

    #[test]
    fn a_pr_belongs_to_the_card_by_its_claim_line_or_its_branch() {
        let mut p = pr();
        p.branch = "main".into();
        assert!(!claims_pr(3067, &p));
        p.body = "Refs #3067\nMoai-Claim: 3067#6897022999".into();
        assert!(claims_pr(3067, &p));
        assert!(!claims_pr(3068, &p));
        p.body = String::new();
        p.branch = "issue/3067-card-times".into();
        assert!(claims_pr(3067, &p));
        assert!(!claims_pr(306, &p), "issue/306- is not issue/3067-");
        p.branch = "issue/30670-x".into();
        assert!(!claims_pr(3067, &p));
    }

    #[test]
    fn opened_and_verdict_to_merged_spans_use_server_times() {
        let mut p = pr();
        p.merged = true;
        p.state = "MERGED".into();
        p.merged_at = Some("2026-11-16T11:00:00Z".into());
        p.comments = vec![comment(
            4,
            "2026-11-16T10:00:00Z",
            &format!("AUDIT #7 head {} — verdict: PASS", p.head),
        )];
        let v = pr_json(&p, now());
        assert_eq!(v["spans"]["opened_to_merged"]["secs"], 3 * 3600);
        assert_eq!(v["spans"]["verdict_to_merged"]["secs"], 3600);
        assert_eq!(v["spans"]["opened_to_ready"]["secs"], 0);
        assert_eq!(v["state"], "merged");
        assert_eq!(v["number"], 7);
    }
}
