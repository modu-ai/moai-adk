//! rc.2 A2 (#2161) `queue-next`: self-claim over recorded exchanges (`RecordedTransport`
//! with an injected clock — no network, no sleeping), in the fixture style of
//! `tests/claim_cli.rs`. `RecordedTransport` asserts method and URL in order, so an
//! unexpected request (for example a second POST to a lost card) fails the test.
mod common;

use common::{exchange, json_response as resp};
use moai_cli::claim::{
    acquire, task_key, ClaimRequest, Journal, Owner, SpecRef, TaskRef, WorktreeId,
};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, RecordedTransport};
use moai_cli::forge::{Client, Config};
use moai_cli::queue::{free, next, ClaimTemplate, FreeRequest, QueueRequest};
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
            owner: Owner::new("session", session),
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
    WorktreeId::new(
        "/tmp/repo/.git",
        "/tmp/repo/.git/worktrees/lane",
        "/tmp/repo",
        "b".repeat(40),
    )
}

fn card(number: i64, key: &str, body: &str) -> Value {
    json!({"number": number, "id": number * 1000, "title": format!("[rc.2][{key}] card {number}"),
           "body": body, "milestone": {"number": 3}})
}

fn claim(id: i64, session: &str) -> Value {
    json!({"id": id, "body": format!("moai-claim: session={session} epoch=1 lease_until={LEASE}")})
}

fn handoff(id: i64, target: i64, to: &str) -> Value {
    json!({"id": id, "body": moai_cli::claim_verify::render_handoff_line(target, to, "leader")})
}

fn release(id: i64, target: i64) -> Value {
    json!({"id": id, "body": moai_cli::claim_verify::render_release_line(target, "leader")})
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
    attempt_with_history(n, posted, vec![], reread)
}

fn attempt_with_history(
    n: i64,
    posted: i64,
    history: Vec<Value>,
    reread: Vec<Value>,
) -> Vec<Exchange> {
    vec![
        exchange("GET", &comments(n), resp(200, json!(history))),
        exchange("GET", &comments(n), resp(200, json!(history))),
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
    let key = task_key(&TaskRef::new(
        "github",
        "github.com",
        "acme",
        "widgets",
        101,
        issue * 1000,
        issue,
    ));
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

/// #2344 AC-1/AC-3(a): once the lane has handed its card to the gate (`pr_ready`, written
/// by pr-confirm), the next queue-next skips that card even though its Issue is still open
/// until the merge, and claims the free one. The handed-off record is left exactly as it
/// was: still `pr_ready` (so handoff, worktree and scope-check keep treating it as active)
/// and still visible to claim-show, never `cancelled`.
#[test]
fn own_pr_ready_card_is_skipped_and_the_free_card_is_claimed() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");

    // Round 1: claim #7.
    let mut ex = reads(vec![card(7, "A1", "상위: #2"), card(8, "A2", "상위: #2")]);
    ex.extend(attempt(7, 500, vec![claim(500, "session-a")]));
    let (first, t1) = run(&root, "session-a", ex);
    assert_eq!(first["outcome"], json!("claimed"), "{first}");
    assert_eq!(first["issue"], json!(7));
    assert_eq!(t1.remaining(), 0);

    // Hand-off: the same change pr-confirm makes (pr_link.rs: `r.state = "pr_ready"`).
    {
        let journal = locked(&root);
        let mut rec = journal
            .read_events()
            .unwrap()
            .last()
            .unwrap()
            .record
            .clone();
        rec.state = "pr_ready".into();
        journal.commit("pr_ready", &rec).unwrap();
    }

    // Round 2: #7 is still open (not merged yet), #8 is free. Every request the run makes
    // is recorded, so a read or POST for #7 beyond the open-issue list fails the test.
    let mut ex = reads(vec![card(7, "A1", "상위: #2"), card(8, "A2", "상위: #2")]);
    ex.extend(attempt(8, 600, vec![claim(600, "session-a")]));
    let (second, t2) = run(&root, "session-a", ex);
    assert_eq!(second["outcome"], json!("claimed"), "{second}");
    assert_eq!(second["issue"], json!(8), "{second}");
    assert_eq!(
        second["skipped"],
        json!([{"issue": 7, "reason": "handed_off", "state": "pr_ready", "epoch": 1}]),
        "{second}"
    );
    assert_eq!(t2.remaining(), 0);

    // AC-3(a): the #7 record is unchanged and still shown.
    let key7 = task_key(&TaskRef::new(
        "github",
        "github.com",
        "acme",
        "widgets",
        101,
        7000,
        7,
    ));
    let shown = moai_cli::claim::show(&root, &key7).unwrap();
    assert_eq!(shown["outcome"], json!("held"), "{shown}");
    assert_eq!(shown["record"]["state"], json!("pr_ready"), "{shown}");
    assert_eq!(shown["record"]["comment_id"], json!(500), "{shown}");
}

/// #2344 control for the skip above: an own card that is NOT handed off yet (`assigned`) is
/// still returned as `already_owned`, so a restarted lane resumes its unfinished card
/// instead of claiming another one. No request is made for the free #8.
#[test]
fn own_unfinished_card_is_resumed_not_skipped() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let mut ex = reads(vec![card(7, "A1", "상위: #2"), card(8, "A2", "상위: #2")]);
    ex.extend(attempt(7, 500, vec![claim(500, "session-a")]));
    let (first, _) = run(&root, "session-a", ex);
    assert_eq!(first["record"]["state"], json!("assigned"), "{first}");

    let mut ex = reads(vec![card(7, "A1", "상위: #2"), card(8, "A2", "상위: #2")]);
    ex.push(exchange(
        "GET",
        &comments(7),
        resp(200, json!([claim(500, "session-a")])),
    ));
    let (again, t) = run(&root, "session-a", ex);
    assert_eq!(again["outcome"], json!("already_owned"), "{again}");
    assert_eq!(again["issue"], json!(7), "{again}");
    assert_eq!(again["skipped"], json!([]), "{again}");
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

/// A retargeted handoff is a literal reservation, not a prefix/substring match or a
/// claim. The near-named session skips without writing, while the recipient must POST
/// its own claim and read it back before it owns the card.
#[test]
fn handoff_reserves_for_exact_recipient_and_recipient_acquires() {
    let tmp = tempfile::tempdir().unwrap();
    let history = vec![
        claim(100, "source"),
        handoff(150, 100, "session-target-extra"),
        handoff(200, 100, "session-target"),
    ];
    for session in ["session-target-extra", "session-targe"] {
        let foreign_root = tmp.path().join(session);
        let mut ex = reads(vec![card(7, "A1", ""), card(8, "A2", "")]);
        ex.push(exchange("GET", &comments(7), resp(200, json!(history))));
        ex.extend(attempt(8, 600, vec![claim(600, session)]));
        let (out, t) = run(&foreign_root, session, ex);
        assert_eq!(out["outcome"], json!("claimed"), "{out}");
        assert_eq!(out["issue"], json!(8));
        assert_eq!(
            out["skipped"],
            json!([{
                "issue": 7, "reason": "handed_to_other", "to": "session-target",
                "comment_id": 200, "target": 100,
            }])
        );
        assert_eq!(posts(&t), vec![format!("{API}/issues/8/comments")]);
        assert!(!record_file(&foreign_root, 7).exists());
        assert_eq!(
            fs::read_to_string(foreign_root.join("journal.log"))
                .unwrap()
                .lines()
                .count(),
            1
        );
        assert_eq!(t.remaining(), 0);
    }

    let recipient_root = tmp.path().join("recipient");
    let mut ex = reads(vec![card(7, "A1", "")]);
    let mut reread = history.clone();
    reread.push(claim(500, "session-target"));
    ex.extend(attempt_with_history(7, 500, history, reread));
    let (out, t) = run(&recipient_root, "session-target", ex);
    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(7));
    assert_eq!(out["record"]["comment_id"], json!(500));
    assert_eq!(out["record"]["owner"]["id"], json!("session-target"));
    assert_eq!(out["skipped"], json!([]));
    assert_eq!(out["cross_host_atomic"], json!(false));
    assert_eq!(posts(&t), vec![format!("{API}/issues/7/comments")]);
    assert!(record_file(&recipient_root, 7).exists());
    assert_eq!(t.remaining(), 0);
}

/// Releasing the source cancels its reservation and opens the card to an unrelated
/// session. Release is not an implicit transfer to the previous handoff recipient.
#[test]
fn release_reopens_handed_card_to_any_session() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let history = vec![
        claim(100, "source"),
        handoff(150, 100, "recipient"),
        release(200, 100),
    ];
    let mut ex = reads(vec![card(7, "A1", "")]);
    let mut reread = history.clone();
    reread.push(claim(500, "unrelated"));
    ex.extend(attempt_with_history(7, 500, history, reread));
    let (out, t) = run(&root, "unrelated", ex);
    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(7));
    assert_eq!(out["skipped"], json!([]));
    assert_eq!(posts(&t), vec![format!("{API}/issues/7/comments")]);
    assert_eq!(t.remaining(), 0);
}

/// Only a reference to an earlier actual claim creates a reservation. Invalid source
/// references must neither reserve an unclaimed card nor close a future claim.
#[test]
fn handoff_requires_an_earlier_actual_claim() {
    for (name, history, held) in [
        ("missing", vec![handoff(200, 99, "other")], false),
        (
            "nonclaim",
            vec![
                json!({"id": 100, "body": "ordinary comment"}),
                handoff(200, 100, "other"),
            ],
            false,
        ),
        (
            "forward",
            vec![handoff(100, 200, "other"), claim(200, "source")],
            true,
        ),
    ] {
        let tmp = tempfile::tempdir().unwrap();
        let root = tmp.path().join(name);
        let mut ex = reads(vec![card(7, "A1", "")]);
        if held {
            ex.push(exchange("GET", &comments(7), resp(200, json!(history))));
        } else {
            let mut reread = history.clone();
            reread.push(claim(500, "session-a"));
            ex.extend(attempt_with_history(7, 500, history, reread));
        }
        let (out, t) = run(&root, "session-a", ex);
        if held {
            assert_eq!(out["outcome"], json!("exhausted"), "{name}: {out}");
            assert_eq!(out["skipped"][0]["reason"], json!("claimed"));
            assert_eq!(out["skipped"][0]["comment_id"], json!(200));
            assert!(posts(&t).is_empty());
            assert!(!record_file(&root, 7).exists());
        } else {
            assert_eq!(out["outcome"], json!("claimed"), "{name}: {out}");
            assert_eq!(out["skipped"], json!([]));
            assert_eq!(posts(&t), vec![format!("{API}/issues/7/comments")]);
        }
        assert_eq!(t.remaining(), 0, "{name}");
    }
}

/// A recipient's later claim consumes the reservation. Releasing that claim leaves
/// the card open, rather than resurrecting the original source's historical handoff.
#[test]
fn recipient_claim_and_release_do_not_leave_a_historical_reservation() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let history = vec![
        claim(100, "source"),
        handoff(150, 100, "recipient"),
        claim(200, "recipient"),
        release(250, 200),
    ];
    let mut ex = reads(vec![card(7, "A1", "")]);
    let mut reread = history.clone();
    reread.push(claim(500, "unrelated"));
    ex.extend(attempt_with_history(7, 500, history, reread));
    let (out, t) = run(&root, "unrelated", ex);
    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["skipped"], json!([]));
    assert_eq!(out["record"]["comment_id"], json!(500));
    assert_eq!(posts(&t), vec![format!("{API}/issues/7/comments")]);
    assert_eq!(t.remaining(), 0);
}

/// The reservation can live on a later comment page than its source. Read through
/// Link pagination before deciding that an expired claim merely blocks the candidate.
#[test]
fn handoff_on_later_comment_page_is_respected() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let page2 = format!("{}&page=2", comments(7));
    let mut ex = reads(vec![card(7, "A1", ""), card(8, "A2", "")]);
    ex.push(exchange(
        "GET",
        &comments(7),
        resp(
            200,
            json!([claim_until(100, "source", "2026-01-01T00:00:00Z")]),
        )
        .with_header("link", format!("<{page2}>; rel=\"next\"")),
    ));
    ex.push(exchange(
        "GET",
        &page2,
        resp(200, json!([handoff(200, 100, "recipient")])),
    ));
    ex.extend(attempt(8, 600, vec![claim(600, "session-a")]));
    let (out, t) = run(&root, "session-a", ex);
    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(8));
    assert_eq!(out["skipped"][0]["reason"], json!("handed_to_other"));
    assert_eq!(out["skipped"][0]["to"], json!("recipient"));
    assert_eq!(out["skipped"][0]["comment_id"], json!(200));
    assert_eq!(posts(&t), vec![format!("{API}/issues/8/comments")]);
    assert!(!record_file(&root, 7).exists());
    assert_eq!(t.remaining(), 0);
}

/// Remote reservation wins over a stale assigned journal record. Do not mutate that
/// record or return already_owned; the lane can claim another free card instead.
#[test]
fn stale_own_active_record_cannot_bypass_foreign_handoff() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let mut ex = reads(vec![card(7, "A1", "")]);
    ex.extend(attempt(7, 100, vec![claim(100, "session-a")]));
    let (first, _) = run(&root, "session-a", ex);
    assert_eq!(first["outcome"], json!("claimed"));
    let original = fs::read(record_file(&root, 7)).unwrap();
    let mut ex = reads(vec![card(7, "A1", ""), card(8, "A2", "")]);
    ex.push(exchange(
        "GET",
        &comments(7),
        resp(
            200,
            json!([claim(100, "session-a"), handoff(200, 100, "recipient")]),
        ),
    ));
    ex.extend(attempt(8, 600, vec![claim(600, "session-a")]));
    let (out, t) = run(&root, "session-a", ex);
    assert_eq!(out["outcome"], json!("claimed"), "{out}");
    assert_eq!(out["issue"], json!(8));
    assert_eq!(out["skipped"][0]["reason"], json!("handed_to_other"));
    assert_eq!(out["skipped"][0]["to"], json!("recipient"));
    assert_eq!(posts(&t), vec![format!("{API}/issues/8/comments")]);
    assert_eq!(fs::read(record_file(&root, 7)).unwrap(), original);
    assert_eq!(
        fs::read_to_string(root.join("journal.log"))
            .unwrap()
            .lines()
            .count(),
        2
    );
    assert_eq!(t.remaining(), 0);
}

/// A reservation does not bypass acquire's uncertain readback rule or allow the lane
/// to touch the free card after posting a claim whose visibility is unresolved.
#[test]
fn recipient_with_unseen_claim_readback_stops_uncertain() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let history = vec![claim(100, "source"), handoff(200, 100, "recipient")];
    let mut ex = reads(vec![card(7, "A1", ""), card(8, "A2", "")]);
    ex.extend(attempt_with_history(7, 500, history.clone(), history));
    let (out, t) = run(&root, "recipient", ex);
    assert_eq!(out["outcome"], json!("uncertain"), "{out}");
    assert_eq!(out["issue"], json!(7));
    assert_eq!(out["record"], Value::Null);
    assert_eq!(out["cross_host_atomic"], json!(false));
    assert_eq!(posts(&t), vec![format!("{API}/issues/7/comments")]);
    assert!(!record_file(&root, 7).exists());
    assert!(!record_file(&root, 8).exists());
    assert!(!root.join("journal.log").exists());
    assert_eq!(t.remaining(), 0);
}

/// A syntactically complete prior claim is not free just because its lease is invalid.
/// Preserve the queue's no-automatic-takeover behavior without relying on parser wording.
#[test]
fn prior_claim_with_invalid_lease_blocks_without_writing() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("claims");
    let mut ex = reads(vec![card(7, "A1", "")]);
    ex.push(exchange(
        "GET",
        &comments(7),
        resp(200, json!([claim_until(100, "source", "invalid-lease")])),
    ));
    let (out, t) = run(&root, "session-a", ex);
    assert_eq!(out["outcome"], json!("exhausted"), "{out}");
    assert_eq!(out["skipped"][0]["reason"], json!("claimed"));
    assert_eq!(out["skipped"][0]["claimed_by"], json!("source"));
    assert_eq!(out["skipped"][0]["comment_id"], json!(100));
    assert!(posts(&t).is_empty());
    assert!(!record_file(&root, 7).exists());
    assert!(!root.join("journal.log").exists());
    assert_eq!(t.remaining(), 0);
}

/// Exercise acquire directly: queue prefiltering would hide invalid closure targets
/// before the claim consumer could demonstrate its own live-owner safety boundary.
#[test]
fn acquire_honors_literal_release_and_handoff_closures() {
    let template = request("recipient").claim;
    let req = ClaimRequest::new(
        1,
        TaskRef::new("github", "github.com", "acme", "widgets", 101, 7000, 7),
        "A1".to_string(),
        template.owner,
        template.base_revision,
        template.write_scope,
        template.spec_ref,
    )
    .lease_until(template.lease_until);
    for kind in ["release", "handoff"] {
        for (case, source, target, closing_id, closes) in [
            ("exact", 100, 100, 2000, true),
            ("shorter_target", 1000, 100, 2000, false),
            ("longer_target", 100, 1000, 2000, false),
            ("before_source", 100, 100, 50, false),
            ("near_marker", 100, 100, 2000, false),
        ] {
            let tmp = tempfile::tempdir().unwrap();
            let root = tmp.path().join("claims");
            let journal = locked(&root);
            let mut closing = match kind {
                "release" => release(closing_id, target),
                _ => handoff(closing_id, target, "recipient"),
            };
            if case == "near_marker" {
                closing["body"] = json!(format!(
                    "moai-claim-{kind}-extra: comment={target} to=recipient"
                ));
            }
            let history = vec![claim(source, "source"), closing];
            let mut ex = vec![exchange("GET", &comments(7), resp(200, json!(history)))];
            if closes {
                let mut reread = history;
                reread.push(claim(3000, "recipient"));
                ex.push(exchange(
                    "POST",
                    &format!("{API}/issues/7/comments"),
                    resp(201, json!({"id": 3000})),
                ));
                ex.push(exchange("GET", &comments(7), resp(200, json!(reread))));
            }
            let t = RecordedTransport::new(ex);
            let clock = FakeClock(Cell::new(CLOCK0));
            let client = Client::new(&t, &clock, Config::default(), Default::default());
            let out = acquire(&journal, &req, worktree(), &client, &clock).unwrap();
            if closes {
                assert_eq!(out.outcome, "claimed", "{kind}/{case}: {out:?}");
                let record = out.record.as_ref().unwrap();
                assert_eq!(record.owner.id, "recipient");
                assert_eq!(record.comment_id, Some(3000));
                assert!(out.foreign_owner.is_none());
                assert_eq!(posts(&t), vec![format!("{API}/issues/7/comments")]);
                assert!(record_file(&root, 7).exists());
                assert_eq!(journal.read_events().unwrap().len(), 1);
            } else {
                assert_eq!(out.outcome, "already_owned", "{kind}/{case}: {out:?}");
                assert!(out.record.is_none());
                let foreign = out.foreign_owner.as_ref().unwrap();
                assert_eq!(foreign["session"], json!("source"));
                assert_eq!(foreign["comment_id"], json!(source));
                assert!(posts(&t).is_empty());
                assert!(!record_file(&root, 7).exists());
                assert!(!root.join("journal.log").exists());
                assert!(journal.read_events().unwrap().is_empty());
            }
            assert_eq!(t.remaining(), 0, "{kind}/{case}");
        }
    }
}

// ---- queue-free (#2675) ----

const EXPIRED: &str = "2026-11-01T00:00:00Z";

fn free_run(exchanges: Vec<Exchange>) -> (Result<Value, String>, RecordedTransport) {
    let t = RecordedTransport::new(exchanges);
    let clock = FakeClock(Cell::new(CLOCK0));
    let client = Client::new(&t, &clock, Config::default(), Default::default());
    let req = FreeRequest {
        schema_version: 1,
        owner: "acme".into(),
        repo: "widgets".into(),
    };
    let out = free(&req, &client, &clock);
    (out, t)
}

fn open_issues(issues: Vec<Value>) -> Exchange {
    exchange(
        "GET",
        &format!("{API}/issues?state=open&per_page=100"),
        resp(200, json!(issues)),
    )
}

fn numbers(v: &Value, key: &str) -> Vec<i64> {
    v[key]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["issue"].as_i64().unwrap())
        .collect()
}

/// Every open Issue lands in exactly one bucket; held cards, epics and blocked cards cost no
/// comment read (the recorded transport fails on any request it does not expect), and the
/// run writes nothing.
#[test]
fn queue_free_sorts_every_card_into_one_bucket() {
    let mut held = card(1, "H", "");
    held["labels"] = json!([{"name": "type:feature"}, {"name": "needs:decision"}]);
    let epic = json!({"number": 2, "id": 2000, "title": "[rc.2][EP-X] epic", "body": ""});
    let blocked = card(3, "B", "Blocked-By: #1");
    let pr = json!({"number": 4, "id": 4000, "title": "pr", "body": "",
                    "pull_request": {"url": "x"}});
    let (out, t) = free_run(vec![
        open_issues(vec![
            held,
            epic,
            blocked,
            pr,
            card(5, "U", ""),
            card(6, "R", ""),
            card(7, "E", ""),
            card(8, "C", ""),
            card(9, "V", ""),
            card(10, "D", "Blocked-By: #999"),
        ]),
        exchange("GET", &comments(5), resp(200, json!([]))),
        exchange(
            "GET",
            &comments(6),
            resp(200, json!([claim(60, "agent-6"), release(61, 60)])),
        ),
        exchange(
            "GET",
            &comments(7),
            resp(200, json!([claim_until(70, "agent-7", EXPIRED)])),
        ),
        exchange(
            "GET",
            &comments(8),
            resp(200, json!([claim(80, "agent-8")])),
        ),
        exchange(
            "GET",
            &comments(9),
            resp(
                200,
                json!([claim(90, "agent-1"), handoff(91, 90, "agent-9")]),
            ),
        ),
        exchange("GET", &comments(10), resp(200, json!([]))),
    ]);
    let out = out.unwrap();
    assert_eq!(t.remaining(), 0);
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
    assert_eq!(out["writes"], json!(false));
    assert_eq!(out["inspected"], json!(9), "{out}");
    assert_eq!(numbers(&out, "free"), vec![5, 6, 7, 10]);
    let why: Vec<&str> = out["free"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["why"].as_str().unwrap())
        .collect();
    assert_eq!(why, vec!["unclaimed", "released", "expired", "unclaimed"]);
    assert_eq!(out["free"][2]["claimed_by"], json!("agent-7"));
    assert_eq!(out["free"][2]["lease_until"], json!(EXPIRED));
    assert_eq!(out["free_count"], json!(4));
    assert_eq!(numbers(&out, "claimed"), vec![8]);
    assert_eq!(out["claimed"][0]["claimed_by"], json!("agent-8"));
    assert_eq!(numbers(&out, "reserved"), vec![9]);
    assert_eq!(out["reserved"][0]["to"], json!("agent-9"));
    assert_eq!(numbers(&out, "held"), vec![1]);
    assert_eq!(out["held"][0]["labels"], json!(["needs:decision"]));
    assert_eq!(numbers(&out, "epics"), vec![2]);
    assert_eq!(numbers(&out, "blocked"), vec![3]);
    assert_eq!(out["blocked"][0]["blocked_by"], json!([1]));
    assert_eq!(
        out["summary"],
        json!("free 4건 (inspected 9, claimed 1, reserved 1, held 1, epics 1, blocked 1)")
    );
}

/// An empty answer states its count (#2675 AC-2): `free_count: 0`, an empty list, and a
/// summary that begins `free 0건`.
#[test]
fn queue_free_says_zero_when_nothing_is_free() {
    let mut held = card(1, "H", "");
    held["labels"] = json!([{"name": "HOLD"}]);
    let (out, t) = free_run(vec![
        open_issues(vec![held, card(2, "C", "")]),
        exchange(
            "GET",
            &comments(2),
            resp(200, json!([claim(20, "agent-2")])),
        ),
    ]);
    let out = out.unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(out["free_count"], json!(0));
    assert_eq!(out["free"], json!([]));
    assert!(
        out["summary"].as_str().unwrap().starts_with("free 0건 "),
        "{out}"
    );
}

/// A failed read is an error, never an empty list (#2675: an empty output was once read as
/// "0 cards").
#[test]
fn queue_free_read_failure_is_an_error_not_an_empty_list() {
    let (out, _t) = free_run(vec![
        open_issues(vec![card(5, "U", "")]),
        exchange(
            "GET",
            &comments(5),
            resp(404, json!({"message": "Not Found"})),
        ),
    ]);
    assert!(out.is_err(), "{out:?}");
}

/// Request shape: only schema version 1 and GitHub owner/repo names are read.
#[test]
fn queue_free_refuses_a_bad_request_before_any_read() {
    let t = RecordedTransport::new(vec![]);
    let clock = FakeClock(Cell::new(CLOCK0));
    let client = Client::new(&t, &clock, Config::default(), Default::default());
    for (v, owner) in [(2, "acme"), (1, "../acme"), (1, "")] {
        let req = FreeRequest {
            schema_version: v,
            owner: owner.into(),
            repo: "widgets".into(),
        };
        assert!(free(&req, &client, &clock).is_err(), "{v} {owner}");
    }
    assert!(t.sent.borrow().is_empty());
}

/// #2675 audit N4: only whole label names hold a card. `threshold` or `holder` labels leave the
/// card free; `HOLD`, `status:hold` and `needs:decision` hold it.
#[test]
fn queue_free_holds_only_on_whole_label_names() {
    let labelled = |n: i64, names: &[&str]| {
        let mut c = card(n, "L", "");
        c["labels"] = json!(names.iter().map(|l| json!({"name": l})).collect::<Vec<_>>());
        c
    };
    let (out, t) = free_run(vec![
        open_issues(vec![
            labelled(1, &["threshold"]),
            labelled(2, &["area:holder"]),
            labelled(3, &["HOLD"]),
            labelled(4, &["status:hold"]),
            labelled(5, &["Needs:Decision"]),
        ]),
        exchange("GET", &comments(1), resp(200, json!([]))),
        exchange("GET", &comments(2), resp(200, json!([]))),
    ]);
    let out = out.unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(numbers(&out, "free"), vec![1, 2]);
    assert_eq!(numbers(&out, "held"), vec![3, 4, 5]);
}

/// #2675 audit N1–N3: the edges of `queue-free` — a card naming itself in `Blocked-By:` is not
/// blocked; a lease that ends exactly now is expired (live only when it ends later, the same
/// `t > now` as `queue-next`); with several live claims the lowest comment id is the holder.
#[test]
fn queue_free_edges_self_block_lease_now_and_lowest_live_claim() {
    const NOW: &str = "2026-11-06T21:20:00Z"; // CLOCK0
    const NOW_PLUS_1: &str = "2026-11-06T21:20:01Z";
    let (out, t) = free_run(vec![
        open_issues(vec![
            card(1, "S", "Blocked-By: #1"),
            card(2, "N", ""),
            card(3, "P", ""),
            card(4, "M", ""),
        ]),
        exchange("GET", &comments(1), resp(200, json!([]))),
        exchange(
            "GET",
            &comments(2),
            resp(200, json!([claim_until(20, "agent-2", NOW)])),
        ),
        exchange(
            "GET",
            &comments(3),
            resp(200, json!([claim_until(30, "agent-3", NOW_PLUS_1)])),
        ),
        exchange(
            "GET",
            &comments(4),
            resp(
                200,
                json!([
                    claim_until(41, "agent-b", LEASE),
                    claim_until(40, "agent-a", LEASE)
                ]),
            ),
        ),
    ]);
    let out = out.unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(
        numbers(&out, "blocked"),
        Vec::<i64>::new(),
        "self reference is not a blocker"
    );
    assert_eq!(numbers(&out, "free"), vec![1, 2]);
    assert_eq!(
        out["free"][1]["why"],
        json!("expired"),
        "a lease ending now is expired"
    );
    assert_eq!(numbers(&out, "claimed"), vec![3, 4]);
    assert_eq!(out["claimed"][1]["claimed_by"], json!("agent-a"));
    assert_eq!(out["claimed"][1]["comment_id"], json!(40));
}
