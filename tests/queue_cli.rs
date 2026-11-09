//! rc.2 A2 (#2161) `queue-next`: self-claim over recorded exchanges (`RecordedTransport`
//! with an injected clock — no network, no sleeping), in the fixture style of
//! `tests/claim_cli.rs`. `RecordedTransport` asserts method and URL in order, so an
//! unexpected request (for example a second POST to a lost card) fails the test.
use moai_cli::claim::{task_key, Journal, Owner, SpecRef, TaskRef, WorktreeId};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::forge::{Client, Config};
use moai_cli::queue::{next, ClaimTemplate, QueueRequest};
use serde_json::{json, Value};
use std::{cell::Cell, fs, path::Path, thread, time::Duration};

const CLOCK0: i64 = 1_794_000_000;
const LEASE: &str = "2026-11-09T12:00:00Z";
const API: &str = "https://api.github.com/repos/acme/widgets";

struct FakeClock(Cell<i64>);
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.0.get()
    }
    fn sleep(&self, secs: u64) {
        self.0.set(self.0.get() + secs as i64);
    }
}

fn request(session: &str) -> QueueRequest {
    QueueRequest {
        schema_version: 1,
        owner: "acme".into(),
        repo: "widgets".into(),
        milestone: 3,
        claim: ClaimTemplate {
            owner: Owner {
                kind: "session".into(),
                id: session.into(),
                host_id: None,
                pid: None,
                process_start: None,
            },
            base_revision: "a".repeat(40),
            write_scope: vec!["src/queue.rs".into()],
            spec_ref: SpecRef {
                id: "SPEC-RC2-A2".into(),
                revision: "r1".into(),
                digest: None,
            },
            lease_until: LEASE.into(),
        },
    }
}

fn worktree() -> WorktreeId {
    WorktreeId {
        common_dir: "/tmp/repo/.git".into(),
        git_dir: "/tmp/repo/.git/worktrees/lane".into(),
        path: "/tmp/repo".into(),
        head: "b".repeat(40),
        branch: None,
    }
}

fn resp(status: u16, body: Value) -> HttpResponse {
    HttpResponse {
        status,
        headers: Default::default(),
        body: Some(body),
    }
}

fn exchange(method: &str, url: &str, response: HttpResponse) -> Exchange {
    Exchange {
        request: HttpRequest {
            method: method.into(),
            url: url.into(),
            headers: Default::default(),
            body: None,
        },
        response,
    }
}

fn card(number: i64, key: &str, body: &str) -> Value {
    json!({"number": number, "id": number * 1000, "title": format!("[rc.2][{key}] card {number}"),
           "body": body, "milestone": {"number": 3}})
}

fn claim(id: i64, session: &str) -> Value {
    json!({"id": id, "body": format!("moai-claim: session={session} epoch=1 lease_until={LEASE}")})
}

fn comments(n: i64) -> String {
    format!("{API}/issues/{n}/comments?per_page=100")
}

/// Repository and open-issue reads every run starts with.
fn reads(issues: Vec<Value>) -> Vec<Exchange> {
    vec![
        exchange("GET", API, resp(200, json!({"id": 101}))),
        exchange(
            "GET",
            &format!("{API}/issues?state=open&per_page=100"),
            resp(200, json!(issues)),
        ),
    ]
}

/// Queue candidate check (empty), acquire pre-check (empty), the one POST, and the re-read
/// the caller supplies.
fn attempt(n: i64, posted: i64, reread: Vec<Value>) -> Vec<Exchange> {
    vec![
        exchange("GET", &comments(n), resp(200, json!([]))),
        exchange("GET", &comments(n), resp(200, json!([]))),
        exchange(
            "POST",
            &format!("{API}/issues/{n}/comments"),
            resp(201, json!({"id": posted})),
        ),
        exchange("GET", &comments(n), resp(200, json!(reread))),
    ]
}

fn locked(root: &Path) -> Journal {
    fs::create_dir_all(root).unwrap();
    for _ in 0..200 {
        let j = Journal::open(root).unwrap();
        if j.try_lock().unwrap() {
            return j;
        }
        drop(j);
        thread::sleep(Duration::from_millis(10));
    }
    panic!("journal lock not acquired");
}

fn run(root: &Path, session: &str, exchanges: Vec<Exchange>) -> (Value, RecordedTransport) {
    let journal = locked(root);
    let t = RecordedTransport::new(exchanges);
    let clock = FakeClock(Cell::new(CLOCK0));
    let client = Client::new(&t, &clock, Config::default(), Default::default());
    let out = next(&journal, &request(session), worktree(), &client, &clock).unwrap();
    (out, t)
}

fn record_file(root: &Path, issue: i64) -> std::path::PathBuf {
    let key = task_key(&TaskRef {
        provider: "github".into(),
        host: "github.com".into(),
        owner: "acme".into(),
        repo: "widgets".into(),
        repository_id: 101,
        issue_id: issue * 1000,
        issue_number: issue,
    });
    root.join(format!("{key}.json"))
}

/// AC-1 and AC-3: a concurrent session wins #7 with the smaller live comment id; this
/// session records nothing for #7, moves on, and claims #8. One POST per card, no takeover.
#[test]
fn contest_loser_moves_to_next_candidate_and_records_nothing() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let mut ex = reads(vec![card(7, "A1", "상위: #2"), card(8, "A2", "상위: #2")]);
    ex.extend(attempt(
        7,
        500,
        vec![claim(400, "session-b"), claim(500, "session-a")],
    ));
    ex.extend(attempt(8, 600, vec![claim(600, "session-a")]));
    let (out, t) = run(&root, "session-a", ex);

    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(8));
    assert_eq!(out["lost_to"], json!([{"issue": 7, "comment_id": 400}]));
    assert_eq!(out["record"]["comment_id"], json!(600));
    let posts: Vec<String> = t
        .sent
        .borrow()
        .iter()
        .filter(|r| r.method == "POST")
        .map(|r| r.url.clone())
        .collect();
    assert_eq!(
        posts,
        vec![
            format!("{API}/issues/7/comments"),
            format!("{API}/issues/8/comments")
        ],
        "exactly one POST per card, none repeated on the lost card"
    );
    assert_eq!(t.remaining(), 0);
    // AC-3: the loser keeps no journal record for #7; the winner's record exists for #8.
    assert!(
        !record_file(&root, 7).exists(),
        "lost card must leave no record"
    );
    assert!(record_file(&root, 8).exists());
    let journal = fs::read_to_string(root.join("journal.log")).unwrap();
    assert_eq!(
        journal.lines().count(),
        1,
        "only the #8 acquire is journaled"
    );
    // Projection is intent only (RC1 applies no state label; the report comment is forge-report's).
    assert_eq!(out["projection"]["label"]["state"], json!("not_run"));
    assert_eq!(out["projection"]["status_line"]["state"], json!("not_run"));
}

/// AC-2: a card whose `Blocked-By:` names an open Issue is not a candidate and is never
/// read for claims; a blocker that is not open does not block.
#[test]
fn blocked_card_is_excluded_from_candidates() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let blocker = json!({"number": 9, "id": 9000, "title": "[rc.2][A0] open blocker",
                         "body": "Blocked-By: #9", "milestone": {"number": 4}});
    let mut ex = reads(vec![
        card(7, "A1", "상위: #2\nBlocked-By: #9"),
        card(8, "A2", "상위: #2\nBlocked-By: #555"),
        blocker,
    ]);
    ex.extend(attempt(8, 700, vec![claim(700, "session-a")]));
    let (out, t) = run(&root, "session-a", ex);

    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(8));
    assert_eq!(
        out["skipped"],
        json!([{"issue": 7, "reason": {"blocked_by": [9]}}])
    );
    assert!(t
        .sent
        .borrow()
        .iter()
        .all(|r| !r.url.contains("/issues/7/")));
    assert_eq!(t.remaining(), 0);
}

fn claim_until(id: i64, session: &str, lease: &str) -> Value {
    json!({"id": id, "body": format!("moai-claim: session={session} epoch=1 lease_until={lease}")})
}

fn posts(t: &RecordedTransport) -> Vec<String> {
    t.sent
        .borrow()
        .iter()
        .filter(|r| r.method == "POST")
        .map(|r| r.url.clone())
        .collect()
}

/// agent-2 FAIL probe on PR #2218, fixed as a test (REQ-A2-01, REQ-A2-02, REQ-A7-02): #7
/// carries only an expired foreign claim. It is not a candidate: no POST, no record; the
/// run moves to #8 and `skipped` names the claimant with `expired: true`.
#[test]
fn expired_foreign_claim_is_skipped_not_taken_over() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let mut ex = reads(vec![card(7, "A1", "상위: #2"), card(8, "A2", "상위: #2")]);
    ex.push(exchange(
        "GET",
        &comments(7),
        resp(
            200,
            json!([claim_until(300, "session-b", "2026-01-01T00:00:00Z")]),
        ),
    ));
    ex.extend(attempt(8, 600, vec![claim(600, "session-a")]));
    let (out, t) = run(&root, "session-a", ex);

    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(8));
    assert_eq!(
        out["skipped"],
        json!([{"issue": 7, "reason": "claimed", "claimed_by": "session-b",
                "comment_id": 300, "lease_until": "2026-01-01T00:00:00Z", "expired": true}])
    );
    assert_eq!(posts(&t), vec![format!("{API}/issues/8/comments")]);
    assert_eq!(t.remaining(), 0);
    assert!(
        !record_file(&root, 7).exists(),
        "expired card must leave no record"
    );
}

/// A live foreign claim comment is also skipped before acquire, with `expired: false`;
/// with no other candidate the run ends `exhausted` and posts nothing.
#[test]
fn live_foreign_claim_is_skipped_and_posts_nothing() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let mut ex = reads(vec![card(7, "A1", "상위: #2")]);
    ex.push(exchange(
        "GET",
        &comments(7),
        resp(200, json!([claim(400, "session-b")])),
    ));
    let (out, t) = run(&root, "session-a", ex);

    assert_eq!(out["outcome"], json!("exhausted"), "{out}");
    assert_eq!(out["skipped"][0]["claimed_by"], json!("session-b"));
    assert_eq!(out["skipped"][0]["expired"], json!(false));
    assert!(posts(&t).is_empty());
    assert_eq!(t.remaining(), 0);
}

/// No candidate left: `idle`, nothing posted.
#[test]
fn no_candidate_is_idle_and_posts_nothing() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let ex = reads(vec![
        json!({"number": 2, "id": 2000, "title": "[rc.2][EP-A] epic",
                               "body": "", "milestone": {"number": 3}}),
    ]);
    let (out, t) = run(&root, "session-a", ex);
    assert_eq!(out["outcome"], json!("idle"), "{out}");
    assert_eq!(out["skipped"], json!([{"issue": 2, "reason": "epic"}]));
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
}
