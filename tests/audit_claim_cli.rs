//! rc.2 #2403 item 1: `audit-claim` over an in-memory GitHub (no network). The fake serves
//! the PR, its commits, the referenced Issue's comments and the PR's comments, appends a
//! POSTed comment with the next id, and records every request, so the tests count POSTs and
//! the GETs after a POST. Two overlapping calls meet at a barrier between their pre-read and
//! their POST, so both pass the pre-check before either line exists.
use moai_cli::audit_claim::{render_line, reserve_cli, AuditClaimRequest};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{HttpRequest, HttpResponse, Transport};
use serde_json::{json, Value};
use std::cell::Cell;
use std::sync::{Barrier, Mutex};

const API: &str = "https://api.github.com/repos/acme/widgets";
const PR: i64 = 7;
const ISSUE: i64 = 3;

struct FakeClock(Cell<i64>);
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.0.get()
    }
    fn sleep(&self, secs: u64) {
        self.0.set(self.0.get() + secs as i64);
    }
}

fn clock() -> FakeClock {
    FakeClock(Cell::new(1_794_000_000))
}

fn sha(c: char) -> String {
    std::iter::repeat_n(c, 40).collect()
}

struct Forge {
    head: String,
    state: &'static str,
    pr_body: String,
    commits: Vec<String>,
    issue_comments: Vec<(i64, String)>,
    pr_comments: Mutex<Vec<(i64, String)>>,
    next_id: Mutex<i64>,
    log: Mutex<Vec<(String, String)>>,
    /// When set, each POST waits here first (two overlapping callers).
    barrier: Option<Barrier>,
}

impl Forge {
    fn new(head: &str) -> Self {
        Forge {
            head: head.into(),
            state: "open",
            pr_body: format!("change\n\nRefs #{ISSUE}\n"),
            commits: vec!["feat: x\n\nMoai-Claim: 3#301".into()],
            issue_comments: vec![(
                301,
                "moai-claim: session=author-1 epoch=1 lease_until=2030-01-01T00:00:00Z".into(),
            )],
            pr_comments: Mutex::new(Vec::new()),
            next_id: Mutex::new(1000),
            log: Mutex::new(Vec::new()),
            barrier: None,
        }
    }
    fn with_pr_comments(self, list: &[(i64, &str)]) -> Self {
        *self.pr_comments.lock().unwrap() = list.iter().map(|(i, b)| (*i, b.to_string())).collect();
        self
    }
    fn comments_json(list: &[(i64, String)]) -> Value {
        Value::Array(
            list.iter()
                .map(|(id, body)| json!({"id": id, "created_at": "2026-11-11T00:00:00Z", "body": body}))
                .collect(),
        )
    }
    fn methods(&self) -> Vec<String> {
        self.log
            .lock()
            .unwrap()
            .iter()
            .map(|(m, _)| m.clone())
            .collect()
    }
    fn posts(&self) -> usize {
        self.methods().iter().filter(|m| *m == "POST").count()
    }
    /// GETs sent after the first POST.
    fn gets_after_post(&self) -> usize {
        let m = self.methods();
        match m.iter().position(|x| x == "POST") {
            Some(i) => m[i + 1..].iter().filter(|x| *x == "GET").count(),
            None => 0,
        }
    }
}

impl Transport for Forge {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse, String> {
        self.log
            .lock()
            .unwrap()
            .push((req.method.clone(), req.url.clone()));
        let ok = |v: Value| Ok(HttpResponse::new(200).with_body(v));
        let pr_comments = format!("{API}/issues/{PR}/comments");
        match (req.method.as_str(), req.url.as_str()) {
            ("GET", u) if u == format!("{API}/pulls/{PR}") => ok(json!({
                "number": PR, "state": self.state, "body": self.pr_body,
                "head": {"sha": self.head, "ref": "issue/3-x"}
            })),
            ("GET", u) if u == format!("{API}/pulls/{PR}/commits?per_page=100") => {
                ok(Value::Array(
                    self.commits
                        .iter()
                        .enumerate()
                        .map(|(i, m)| json!({"sha": format!("{i:040}"), "commit": {"message": m}}))
                        .collect(),
                ))
            }
            ("GET", u) if u == format!("{API}/issues/{ISSUE}/comments?per_page=100") => {
                ok(Self::comments_json(&self.issue_comments))
            }
            ("GET", u) if u == format!("{pr_comments}?per_page=100") => {
                ok(Self::comments_json(&self.pr_comments.lock().unwrap()))
            }
            ("POST", u) if u == pr_comments => {
                if let Some(b) = &self.barrier {
                    b.wait();
                }
                let body = req.body.as_ref().unwrap()["body"]
                    .as_str()
                    .unwrap()
                    .to_string();
                let mut next = self.next_id.lock().unwrap();
                *next += 1;
                let id = *next;
                self.pr_comments.lock().unwrap().push((id, body));
                Ok(HttpResponse::new(201).with_body(json!({"id": id})))
            }
            (m, u) => Err(format!("unexpected request {m} {u}")),
        }
    }
}

fn req(lane: &str, head: &str) -> AuditClaimRequest {
    AuditClaimRequest {
        schema_version: 1,
        owner: "acme".into(),
        repo: "widgets".into(),
        pr: PR,
        head: head.into(),
        lane: lane.into(),
    }
}

#[test]
fn an_unreserved_head_is_posted_once_and_re_read_once() {
    let head = sha('a');
    let forge = Forge::new(&head);
    let out = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "reserved");
    assert_eq!(out["posted"], true);
    assert_eq!(out["comment_id"], 1001);
    assert_eq!(out["holder"]["lane"], "auditor-1");
    assert_eq!(out["holder"]["via"], "reservation");
    assert_eq!(out["writes"], "one comment");
    assert_eq!(out["author_rule"], moai_cli::audit_claim::AUTHOR_RULE);
    assert!(out["author_rule"]
        .as_str()
        .unwrap()
        .starts_with("exact match"));
    assert_eq!(
        out["authors"],
        json!([{"session": "author-1", "claim_comment_id": 301}])
    );
    assert_eq!(forge.posts(), 1);
    assert_eq!(forge.gets_after_post(), 1);
    let posted = forge.pr_comments.lock().unwrap();
    assert_eq!(posted[0].1, render_line("auditor-1", &head));
    assert_eq!(
        posted[0].1,
        format!("moai-audit-claim: lane=auditor-1 head={head}")
    );
}

#[test]
fn a_head_held_by_another_lane_is_not_written() {
    let head = sha('a');
    let line = render_line("auditor-2", &head);
    let forge = Forge::new(&head).with_pr_comments(&[(500, &line)]);
    let out = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "held_by");
    assert_eq!(out["posted"], false);
    assert_eq!(out["writes"], "none");
    assert_eq!(out["holder"]["lane"], "auditor-2");
    assert_eq!(out["holder"]["comment_id"], 500);
    assert_eq!(forge.posts(), 0);
}

#[test]
fn the_same_lane_again_returns_its_line_without_posting() {
    let head = sha('a');
    let mine = render_line("auditor-1", &head);
    let late = render_line("auditor-2", &head);
    let forge = Forge::new(&head).with_pr_comments(&[(500, &mine), (501, &late)]);
    let out = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "reserved");
    assert_eq!(out["comment_id"], 500);
    assert_eq!(forge.posts(), 0);
    // The later lane sees it lost, again without a write.
    let out = reserve_cli(req("auditor-2", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "lost");
    assert_eq!(out["comment_id"], 501);
    assert_eq!(out["holder"]["comment_id"], 500);
    assert_eq!(forge.posts(), 0);
}

#[test]
fn a_reservation_on_another_head_is_separate() {
    let (old, head) = (sha('b'), sha('a'));
    let stale = render_line("auditor-2", &old);
    let forge = Forge::new(&head).with_pr_comments(&[(500, &stale)]);
    let out = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "reserved");
    assert_eq!(
        out["reservations"],
        json!([{"lane": "auditor-1", "comment_id": 1001}])
    );
    assert_eq!(forge.posts(), 1);
}

#[test]
fn an_author_lane_is_refused_without_a_write() {
    let head = sha('a');
    let forge = Forge::new(&head);
    let err = reserve_cli(req("author-1", &head), &forge, &clock()).unwrap_err();
    assert!(
        err.starts_with(
            "audit claim conflict: lane author-1 is an author of PR #7 (claim comment 301)"
        ),
        "{err}"
    );
    assert_eq!(forge.posts(), 0);
}

#[test]
fn a_repair_handoff_makes_both_claim_sessions_authors() {
    let head = sha('a');
    let mut forge = Forge::new(&head);
    forge.commits.push("fix: y\n\nMoai-Claim: 3#302".into());
    forge.issue_comments.push((
        302,
        "moai-claim: session=repairer-1 epoch=2 lease_until=2030-01-01T00:00:00Z".into(),
    ));
    let err = reserve_cli(req("repairer-1", &head), &forge, &clock()).unwrap_err();
    assert!(
        err.contains("is an author of PR #7 (claim comment 302)"),
        "{err}"
    );
    let out = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap();
    assert_eq!(out["authors"].as_array().unwrap().len(), 2);
}

#[test]
fn an_unreadable_trailer_claim_refuses_the_call() {
    let head = sha('a');
    let mut forge = Forge::new(&head);
    forge.commits.push("fix: y\n\nMoai-Claim: 3#999".into());
    let err = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap_err();
    assert!(
        err.starts_with("audit claim blocked: claim comment 999"),
        "{err}"
    );
    assert_eq!(forge.posts(), 0);
}

#[test]
fn a_stale_head_or_a_closed_pr_is_refused_without_a_write() {
    let head = sha('a');
    let forge = Forge::new(&head);
    let err = reserve_cli(req("auditor-1", &sha('c')), &forge, &clock()).unwrap_err();
    assert!(
        err.starts_with("audit claim conflict: head is not the current head of PR #7"),
        "{err}"
    );
    let mut closed = Forge::new(&head);
    closed.state = "closed";
    let err = reserve_cli(req("auditor-1", &head), &closed, &clock()).unwrap_err();
    assert_eq!(err, "audit claim conflict: PR #7 is not open");
    assert_eq!(forge.posts() + closed.posts(), 0);
}

#[test]
fn the_commit_limit_blocks_at_250_and_passes_at_249() {
    let head = sha('a');
    let mut at_limit = Forge::new(&head);
    at_limit.commits = vec!["chore: x".into(); 250];
    let err = reserve_cli(req("auditor-1", &head), &at_limit, &clock()).unwrap_err();
    assert_eq!(
        err,
        "audit claim blocked: PR #7 lists 250 commits, the API limit; authors may be missing"
    );
    assert_eq!(at_limit.posts(), 0);
    let mut below = Forge::new(&head);
    below.commits = vec!["chore: x".into(); 249];
    let out = reserve_cli(req("auditor-1", &head), &below, &clock()).unwrap();
    assert_eq!(out["outcome"], "reserved");
}

#[test]
fn a_lane_name_of_100_bytes_passes_and_101_is_refused() {
    let head = sha('a');
    let forge = Forge::new(&head);
    let long = "l".repeat(101);
    let err = reserve_cli(req(&long, &head), &forge, &clock()).unwrap_err();
    assert_eq!(
        err,
        "audit claim refused: lane must be a plain name (letters, digits, - _ .)"
    );
    assert!(forge.methods().is_empty(), "refused before any request");
    let ok = "l".repeat(100);
    let out = reserve_cli(req(&ok, &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "reserved");
}

#[test]
fn a_pr_without_refs_is_blocked() {
    let head = sha('a');
    let mut forge = Forge::new(&head);
    forge.pr_body = "no reference".into();
    let err = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap_err();
    assert!(
        err.starts_with("audit claim blocked: PR #7 names no Refs"),
        "{err}"
    );
    assert_eq!(forge.posts(), 0);
}

#[test]
fn a_leader_handoff_moves_the_holder() {
    let head = sha('a');
    let first = render_line("auditor-2", &head);
    let forge = Forge::new(&head).with_pr_comments(&[
        (500, &first),
        (
            510,
            "moai-claim-handoff: comment=500 to=auditor-1 by=leader",
        ),
        // Only the leader moves a reservation: a later line by a lane is not read.
        (
            520,
            "moai-claim-handoff: comment=500 to=auditor-3 by=auditor-2",
        ),
    ]);
    let out = reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "reserved");
    assert_eq!(out["comment_id"], 510);
    assert_eq!(out["holder"]["via"], "handoff");
    let out = reserve_cli(req("auditor-2", &head), &forge, &clock()).unwrap();
    assert_eq!(out["outcome"], "lost");
    assert_eq!(forge.posts(), 0);
}

#[test]
fn two_overlapping_calls_have_exactly_one_winner_the_lowest_id() {
    let head = sha('a');
    let mut forge = Forge::new(&head);
    forge.barrier = Some(Barrier::new(2));
    let (a, b) = std::thread::scope(|s| {
        let ta = s.spawn(|| reserve_cli(req("auditor-1", &head), &forge, &clock()).unwrap());
        let tb = s.spawn(|| reserve_cli(req("auditor-2", &head), &forge, &clock()).unwrap());
        (ta.join().unwrap(), tb.join().unwrap())
    });
    assert_eq!(
        forge.posts(),
        2,
        "both passed the pre-check before either line existed"
    );
    let mut outcomes = [a["outcome"].clone(), b["outcome"].clone()];
    outcomes.sort_by_key(|v| v.to_string());
    assert_eq!(outcomes, [json!("lost"), json!("reserved")]);
    let winner = if a["outcome"] == "reserved" { &a } else { &b };
    let loser = if a["outcome"] == "reserved" { &b } else { &a };
    assert_eq!(winner["comment_id"], 1001, "the lowest id wins");
    assert_eq!(loser["comment_id"], 1002);
    assert_eq!(loser["holder"]["comment_id"], 1001);
    assert_eq!(loser["holder"]["lane"], winner["lane"]);
}
