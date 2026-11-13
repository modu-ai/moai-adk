//! 06-A (#2005) claim journal and Issue comment claim. Library tests replay recorded
//! exchanges through `RecordedTransport` with an injected clock — no network, no sleeping
//! (the same discipline as `tests/forge_cli.rs`). The CLI tests point `claim-acquire` at a
//! stub `gh` and real git repositories in temporary directories. Real GitHub writes are
//! never performed (live not_run). The file lives beside `product_cli.rs` but stays a new
//! file per the lane dispatch; the AC-named test functions below carry the AC ids.
use moai_cli::claim::{
    acquire, claim_token, release, show, task_key, verify_claim_token, ClaimRequest, CommitCarrier,
    Journal, Owner, SpecRef, TaskRef, TokenCarriers, WorktreeId,
};
use moai_cli::contract;
use moai_cli::forge::policy::{format_utc, Clock};
use moai_cli::forge::transport::{Exchange, HttpRequest, RecordedTransport};
use moai_cli::forge::{Client, Config};
mod common;

use serde_json::{json, Value};
use std::{
    cell::{Cell, RefCell},
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    thread,
    time::{Duration, SystemTime, UNIX_EPOCH},
};

// ---- fixtures ----

const CLOCK0: i64 = 1_794_000_000;

struct FakeClock {
    now: Cell<i64>,
    slept: RefCell<Vec<u64>>,
}
impl FakeClock {
    fn at(now: i64) -> Self {
        FakeClock {
            now: Cell::new(now),
            slept: RefCell::new(Vec::new()),
        }
    }
    fn advance(&self, secs: i64) {
        self.now.set(self.now.get() + secs);
    }
}
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.now.get()
    }
    fn sleep(&self, secs: u64) {
        self.slept.borrow_mut().push(secs);
        self.now.set(self.now.get() + secs as i64);
    }
}

fn task_ref() -> TaskRef {
    TaskRef::new("github", "github.com", "acme", "widgets", 101, 7001, 7)
}

fn lib_req(session: &str, lease_until: &str) -> ClaimRequest {
    ClaimRequest::new(
        1,
        task_ref(),
        "06-A",
        Owner::new("session", session),
        "a".repeat(40),
        vec!["src/claim.rs".into()],
        SpecRef {
            id: "SPEC-RC1-06-A".into(),
            revision: "r1".into(),
            digest: None,
        },
    )
    .lease_until(lease_until)
    .projection_intents(vec![
        json!({"kind": "projection_intent", "target": "assignees", "action": "add"}),
    ])
}

fn worktree() -> WorktreeId {
    WorktreeId::new(
        "/tmp/repo/.git",
        "/tmp/repo/.git/worktrees/lane",
        "/tmp/repo",
        "b".repeat(40),
    )
}

use common::{exchange, json_response as resp};

fn comments_url() -> String {
    "https://api.github.com/repos/acme/widgets/issues/7/comments?per_page=100".into()
}

fn client<'a, T: moai_cli::forge::transport::Transport, C: Clock>(
    t: &'a T,
    c: &'a C,
) -> Client<'a, T, C> {
    Client::new(t, c, Config::default(), Default::default())
}

/// Caller-side waiting policy (REQ-06A-05 leaves waiting to the caller, never the product).
/// Under parallel test load macOS has been observed to answer WouldBlock for one poll after
/// a same-process drop of the previous handle; cross-process claimants are unaffected,
/// because a process exit closes every descriptor. The product itself never retries.
fn lock_journal(root: &Path) -> Journal {
    for _ in 0..200 {
        let journal = Journal::open(root).unwrap();
        if journal.try_lock().unwrap() {
            return journal;
        }
        drop(journal);
        thread::sleep(Duration::from_millis(10));
    }
    panic!("journal lock at {} not acquired within 2s", root.display());
}

/// Acquires against a fresh journal in `root` with the given recorded exchanges.
fn acquire_in(
    root: &Path,
    session: &str,
    lease_until: &str,
    exchanges: Vec<Exchange>,
    clock: &FakeClock,
) -> (moai_cli::claim::AcquireOutcome, RecordedTransport) {
    acquire_with(root, lib_req(session, lease_until), exchanges, clock)
}

/// Acquires with a fully caller-built request (#2169 schema tests set the new fields).
fn acquire_with(
    root: &Path,
    req: moai_cli::claim::ClaimRequest,
    exchanges: Vec<Exchange>,
    clock: &FakeClock,
) -> (moai_cli::claim::AcquireOutcome, RecordedTransport) {
    fs::create_dir_all(root).unwrap();
    let journal = lock_journal(root);
    let t = RecordedTransport::new(exchanges);
    let out = acquire(&journal, &req, worktree(), &client(&t, clock), clock).unwrap();
    (out, t)
}

fn claim_comment(id: i64, session: &str, lease_until: &str) -> Value {
    json!({"id": id, "body": format!("moai-claim: session={session} epoch=1 lease_until={lease_until}")})
}

/// The happy-path exchange script: pre-check GET (empty), the one POST, and the
/// post-POST re-read that shows our own comment as the smallest live id.
fn happy_exchanges(id: i64, session: &str, lease_until: &str) -> Vec<Exchange> {
    vec![
        exchange("GET", &comments_url(), resp(200, json!([]))),
        exchange(
            "POST",
            "https://api.github.com/repos/acme/widgets/issues/7/comments",
            resp(201, json!({"id": id})),
        ),
        exchange(
            "GET",
            &comments_url(),
            resp(200, json!([claim_comment(id, session, lease_until)])),
        ),
    ]
}

// ---- AC-06A-11: exactly one comment POST, the id recorded ----

#[test]
fn claim_comment_posted_once_and_recorded() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-x",
        &lease,
        vec![
            exchange("GET", &comments_url(), resp(200, json!([]))),
            exchange(
                "POST",
                "https://api.github.com/repos/acme/widgets/issues/7/comments",
                resp(201, json!({"id": 900})),
            ),
            // The post-POST re-read must see our own comment as the smallest live id.
            exchange(
                "GET",
                &comments_url(),
                resp(200, json!([claim_comment(900, "agent-x", &lease)])),
            ),
        ],
        &clock,
    );
    assert_eq!(out.outcome, "claimed");
    assert_eq!(t.remaining(), 0);
    let record = out.record.as_ref().unwrap();
    assert_eq!(record.comment_id, Some(900));
    assert_eq!(record.claim_token.as_deref(), Some("Moai-Claim: 7#900"));
    assert_eq!(record.epoch, 1);
    assert_eq!(record.state, "assigned");
    assert_eq!(record.last_observed_at, format_utc(CLOCK0));
    // The intents are recorded as intents (REQ-06A-08) plus the comment effect.
    assert!(record
        .effects
        .iter()
        .any(|e| e["kind"] == "projection_intent"));
    // Exactly one POST was sent, with the canonical claim line; the race is settled by
    // the re-read (GET), not by the POST itself.
    let sent = t.sent.borrow();
    let posts: Vec<&HttpRequest> = sent.iter().filter(|r| r.method == "POST").collect();
    assert_eq!(posts.len(), 1);
    assert_eq!(
        posts[0].body,
        Some(json!({"body": format!("moai-claim: session=agent-x epoch=1 lease_until={lease}")}))
    );
    let methods: Vec<&str> = sent.iter().map(|r| r.method.as_str()).collect();
    assert_eq!(methods, ["GET", "POST", "GET"]);
}

// ---- REQ-06A-13 at claim time: the re-read after the POST settles the race ----

/// A concurrent cross-host claimant can pass the same pre-check and POST between our GET
/// and our POST. The comment id, not the POST, decides: after posting we re-read and the
/// smallest live id wins. The loser keeps its posted comment (it is inert under the rule,
/// and deleting it is no authorized claim write) but records nothing.
#[test]
fn claim_loser_stops_after_post_reread() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-x",
        &lease,
        vec![
            exchange("GET", &comments_url(), resp(200, json!([]))),
            exchange(
                "POST",
                "https://api.github.com/repos/acme/widgets/issues/7/comments",
                resp(201, json!({"id": 900})),
            ),
            exchange(
                "GET",
                &comments_url(),
                resp(
                    200,
                    json!([
                        claim_comment(850, "agent-a", &lease),
                        claim_comment(900, "agent-x", &lease)
                    ]),
                ),
            ),
        ],
        &clock,
    );
    assert_eq!(out.outcome, "already_owned");
    assert_eq!(out.record, None, "the loser records nothing in the journal");
    let foreign = out.foreign_owner.unwrap();
    assert_eq!(foreign["session"], "agent-a");
    assert_eq!(foreign["comment_id"], 850);
    let sent = t.sent.borrow();
    assert_eq!(sent.iter().filter(|r| r.method == "POST").count(), 1);
    assert!(
        out.notes.iter().any(|n| n.contains("no journal record")),
        "{:?}",
        out.notes
    );
    assert!(
        !tmp.path().join("journal.log").exists(),
        "the loser's journal must stay unwritten"
    );
}

/// When the post-POST re-read fails (transport error) or does not show our own comment,
/// the winner is unknown: the outcome is `uncertain` and the journal is not written.
#[test]
fn claim_uncertain_when_reread_fails() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    // The script ends after the POST, so the re-read hits an exhausted transport.
    let (out, _) = acquire_in(
        tmp.path(),
        "agent-x",
        &lease,
        vec![
            exchange("GET", &comments_url(), resp(200, json!([]))),
            exchange(
                "POST",
                "https://api.github.com/repos/acme/widgets/issues/7/comments",
                resp(201, json!({"id": 900})),
            ),
        ],
        &clock,
    );
    assert_eq!(out.outcome, "uncertain");
    assert_eq!(out.record, None);
    assert_eq!(out.foreign_owner, None);
    assert!(
        !tmp.path().join("journal.log").exists(),
        "an uncertain claim records nothing"
    );

    // A successful re-read that does not show our own comment is equally uncertain:
    // the winner rule cannot be evaluated without read-your-writes visibility.
    let tmp = tempfile::tempdir().unwrap();
    let (out, _) = acquire_in(
        tmp.path(),
        "agent-x",
        &lease,
        vec![
            exchange("GET", &comments_url(), resp(200, json!([]))),
            exchange(
                "POST",
                "https://api.github.com/repos/acme/widgets/issues/7/comments",
                resp(201, json!({"id": 900})),
            ),
            exchange(
                "GET",
                &comments_url(),
                resp(200, json!([claim_comment(950, "agent-b", &lease)])),
            ),
        ],
        &clock,
    );
    assert_eq!(out.outcome, "uncertain");
    assert!(!tmp.path().join("journal.log").exists());
}

// ---- AC-06A-8: the smallest live comment id wins; the loser writes nothing ----

#[test]
fn claim_lowest_comment_id_wins() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let live = format_utc(CLOCK0 + 3600);
    let expired = format_utc(CLOCK0 - 10);
    // agent-x sees claims of agent-a (id 100, live), agent-z (id 50, expired) and its own
    // earlier comment (id 200, live).
    let page = resp(
        200,
        json!([
            claim_comment(100, "agent-a", &live),
            claim_comment(50, "agent-z", &expired),
            {"id": 200, "body": format!("leader relay\n\nmoai-claim: session=agent-x epoch=1 lease_until={live}")},
        ]),
    );
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-x",
        &live,
        vec![exchange("GET", &comments_url(), page)],
        &clock,
    );
    // The winner is the smallest live id (100); the expired 50 is ignored; the loser stops
    // without posting or touching the journal (REQ-06A-13).
    assert_eq!(out.outcome, "already_owned");
    assert_eq!(out.record, None);
    let foreign = out.foreign_owner.unwrap();
    assert_eq!(foreign["session"], "agent-a");
    assert_eq!(foreign["comment_id"], 100);
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
    assert!(t.remaining() == 0, "no POST was sent");
    assert!(
        !tmp.path().join("journal.log").exists(),
        "journal untouched"
    );

    // A live claim comment carrying the caller's session value is not adopted when this
    // journal has no record of posting it (#2165 REQ-A6-01; this block used to expect
    // adoption without proof).
    let tmp = tempfile::tempdir().unwrap();
    let page = resp(200, json!([claim_comment(300, "agent-x", &live)]));
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-x",
        &live,
        vec![exchange("GET", &comments_url(), page)],
        &clock,
    );
    assert_eq!(out.outcome, "uncertain");
    assert_eq!(out.record, None);
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"), "no POST");
    assert!(
        !tmp.path().join("journal.log").exists(),
        "journal untouched"
    );
}

// ---- #2165 A6 (REQ-A6-01, REQ-AM-04): same session value, no proof of authorship ----

#[test]
fn claim_same_session_value_without_proof_is_not_adopted() {
    // 2026-11-09: two live sessions were both named `agent-1`. The other one's comment
    // must not become this session's claim.
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let live = format_utc(CLOCK0 + 3600);
    let page = resp(200, json!([claim_comment(6729181667, "agent-1", &live)]));
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-1",
        &live,
        vec![exchange("GET", &comments_url(), page)],
        &clock,
    );
    assert_eq!(out.outcome, "uncertain");
    assert_eq!(out.record, None);
    let foreign = out.foreign_owner.expect("the unproven comment is reported");
    assert_eq!(foreign["session"], "agent-1");
    assert_eq!(foreign["comment_id"], 6729181667i64);
    assert!(out.notes.iter().any(|n| n.contains("not adopted")));
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"), "no POST");
    assert_eq!(t.remaining(), 0);
    assert!(
        !tmp.path().join("journal.log").exists(),
        "journal untouched"
    );
}

#[test]
fn claim_same_session_comment_recorded_in_journal_is_adopted() {
    // Proof: this journal recorded posting comment 900 (epoch 1, later released), so the
    // same live comment is this clone's own and is adopted without a second POST.
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let (first, _) = acquire_in(
        &root,
        "agent-x",
        &lease,
        happy_exchanges(900, "agent-x", &lease),
        &clock,
    );
    assert_eq!(first.outcome, "claimed");
    let journal = lock_journal(&root);
    release(&journal, &task_key(&task_ref()), 1, "handoff", &clock).unwrap();
    drop(journal);

    let page = resp(200, json!([claim_comment(900, "agent-x", &lease)]));
    let (out, t) = acquire_in(
        &root,
        "agent-x",
        &lease,
        vec![exchange("GET", &comments_url(), page)],
        &clock,
    );
    assert_eq!(out.outcome, "claimed");
    assert_eq!(out.record.as_ref().unwrap().comment_id, Some(900));
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"), "no POST");
}

// ---- AC-06A-13: a claim performs no assignee/label/issue-body write ----

#[test]
fn claim_no_other_github_writes() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-x",
        &lease,
        happy_exchanges(901, "agent-x", &lease),
        &clock,
    );
    assert_eq!(out.outcome, "claimed");
    let sent = t.sent.borrow();
    assert!(sent.iter().all(|r| r.method == "GET" || r.method == "POST"));
    assert!(sent.iter().all(|r| !r.url.contains("/labels")
        && !r.url.contains("/assignees")
        && !r.url.ends_with("/issues/7")));
    assert_eq!(
        sent.iter().filter(|r| r.method == "POST").count(),
        1,
        "the comment POST is the only write"
    );
}

// ---- AC-06A-12 / AC-06A-9: the token travels in all three carriers ----

#[test]
fn claim_token_carried_in_branch_trailer_pr() {
    let token = claim_token(2005, 6720575543);
    let carriers = TokenCarriers {
        branch: Some("issue/2005-06-A-c2005-6720575543".into()),
        commits: vec![CommitCarrier {
            sha: "d".repeat(40),
            moai_claim_trailer: Some(token.clone()),
        }],
        pr_body: format!("Refs #2005\nClaim: {token}\n"),
    };
    let v = verify_claim_token(2005, 6720575543, &carriers);
    assert!(v.pass, "{:?}", v.reasons);
}

#[test]
fn claim_token_missing_fails_verification() {
    let token = claim_token(2005, 6720575543);
    // A PR/commit without the token fails (AC-06A-9)...
    let missing = TokenCarriers {
        branch: Some("issue/2005-06-A".into()),
        commits: vec![CommitCarrier {
            sha: "d".repeat(40),
            moai_claim_trailer: None,
        }],
        pr_body: "Refs #2005".into(),
    };
    let v = verify_claim_token(2005, 6720575543, &missing);
    assert!(!v.pass);
    assert_eq!(v.reasons.len(), 3);
    // ...and so does a commit carrying a different (stale) fencing token.
    let stale = TokenCarriers {
        branch: Some("issue/2005-06-A-c2005-6720575543".into()),
        commits: vec![CommitCarrier {
            sha: "d".repeat(40),
            moai_claim_trailer: Some("Moai-Claim: 2005#42".into()),
        }],
        pr_body: token.clone(),
    };
    let v = verify_claim_token(2005, 6720575543, &stale);
    assert!(!v.pass);
    assert!(v.reasons[0].contains("different claim token"));
    assert_eq!(contract::classify(&v.reasons[0]), contract::Kind::Internal);
}

// ---- journal semantics on a real temp directory (no git needed at library level) ----

/// Acquire at t0, then exercise lease and epoch behaviour with the same journal.
fn journal_root(tmp: &Path) -> PathBuf {
    tmp.join("claims")
}

#[test]
fn claim_expired_requires_operator() {
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let exchanges = happy_exchanges(900, "agent-x", &lease);
    let (first, _) = acquire_in(&root, "agent-x", &lease, exchanges, &clock);
    assert_eq!(first.outcome, "claimed");

    // Still live: already_owned, record unchanged.
    let (second, t2) = acquire_in(&root, "agent-y", &lease, vec![], &clock);
    assert_eq!(second.outcome, "already_owned");
    assert_eq!(second.record.as_ref().unwrap().owner.id, "agent-x");
    assert_eq!(second.record, first.record);
    assert!(
        t2.sent.borrow().is_empty(),
        "no GitHub call for a journal-owned task"
    );

    // Expired: no automatic takeover — the operator (06-B) is named instead.
    clock.advance(7200);
    let (third, t3) = acquire_in(&root, "agent-y", &lease, vec![], &clock);
    assert_eq!(third.outcome, "expired");
    assert_eq!(third.record, first.record);
    assert!(t3.sent.borrow().is_empty());
    assert!(third.notes[0].contains("operator approval"));
    // The journal still holds exactly the original record.
    let journal = Journal::open_read(&root).unwrap();
    assert_eq!(journal.read_events().unwrap().len(), 1);
}

#[test]
fn claim_epoch_monotonic() {
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let post = |id: i64| happy_exchanges(id, "agent-x", &lease);
    let (first, _) = acquire_in(&root, "agent-x", &lease, post(900), &clock);
    assert_eq!(first.record.as_ref().unwrap().epoch, 1);

    // Release with the right epoch...
    let journal = lock_journal(&root);
    let rel = release(&journal, &task_key(&task_ref()), 1, "handoff", &clock).unwrap();
    assert_eq!(rel.outcome, "released");
    assert_eq!(rel.record.as_ref().unwrap().state, "cancelled");
    drop(journal);

    // ...then the next acquire is epoch 2 (AC-06A-2).
    let (second, _) = acquire_in(&root, "agent-y", &lease, post(901), &clock);
    assert_eq!(second.outcome, "claimed");
    assert_eq!(second.record.as_ref().unwrap().epoch, 2);
    assert_eq!(
        second.record.as_ref().unwrap().comment_id,
        Some(901),
        "the new claim posts its own comment"
    );
}

#[test]
fn claim_stale_epoch_rejected() {
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    // A stale expected_epoch on a fresh journal is rejected before any request is sent.
    let (out, t) = {
        fs::create_dir_all(&root).unwrap();
        let journal = lock_journal(&root);
        let mut req = lib_req("agent-x", &lease);
        req.expected_epoch = Some(9);
        let t = RecordedTransport::new(vec![]);
        let err = acquire(&journal, &req, worktree(), &client(&t, &clock), &clock).unwrap_err();
        assert!(err.contains("stale epoch"), "{err}");
        assert_eq!(contract::classify(&err), contract::Kind::Conflict, "{err}");
        assert!(t.sent.borrow().is_empty());
        (err, t)
    };
    let _ = out;
    let _ = t;

    // Releasing with an epoch that is not the current record's is a conflict too.
    let exchanges = happy_exchanges(900, "agent-x", &lease);
    let (first, _) = acquire_in(&root, "agent-x", &lease, exchanges, &clock);
    let journal = lock_journal(&root);
    let err = release(&journal, &task_key(&task_ref()), 5, "x", &clock).unwrap_err();
    assert!(err.contains("stale epoch 5"), "{err}");
    assert_eq!(contract::classify(&err), contract::Kind::Conflict);
    // The record is untouched by the refused release.
    let events = journal.read_events().unwrap();
    assert_eq!(events.len(), 1);
    assert_eq!(events[0].record, first.record.unwrap());
}

#[test]
fn claim_journal_torn_write_recovers() {
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let exchanges = happy_exchanges(900, "agent-x", &lease);
    let (first, _) = acquire_in(&root, "agent-x", &lease, exchanges, &clock);
    let log = root.join("journal.log");
    let before = fs::read(&log).unwrap();

    // Fault injection 1: a torn append (no trailing newline) is ignored — the last
    // complete record still reads back (AC-06A-3).
    let mut f = fs::OpenOptions::new().append(true).open(&log).unwrap();
    f.write_all(b"{\"event\":\"acquire\",\"record\":{\"schema_version\":1")
        .unwrap();
    drop(f);
    let journal = Journal::open_read(&root).unwrap();
    let events = journal.read_events().unwrap();
    assert_eq!(events.len(), 1);
    assert_eq!(
        events[0].record.claim_token,
        first.record.as_ref().unwrap().claim_token
    );
    let seen = show(&root, &task_key(&task_ref())).unwrap();
    assert_eq!(seen["outcome"], "held");
    assert_eq!(
        seen["record"]["claim_token"],
        first.record.as_ref().unwrap().claim_token.clone().unwrap()
    );

    // Fault injection 2: an orphaned temporary record file is ignored.
    fs::write(
        root.join(format!("{}.json.tmp-999", task_key(&task_ref()))),
        b"junk",
    )
    .unwrap();
    let events = Journal::open_read(&root).unwrap().read_events().unwrap();
    assert_eq!(events.len(), 1);

    // A corrupt COMPLETE line is a real error, not a silently dropped record.
    let mut restored = before.clone();
    restored.extend_from_slice(b"not json\n");
    fs::write(&log, &restored).unwrap();
    let err = Journal::open_read(&root)
        .unwrap()
        .read_events()
        .unwrap_err();
    assert!(err.contains("claim journal read failed at line 2"), "{err}");
}

#[test]
fn claim_symlink_journal_rejected() {
    let tmp = tempfile::tempdir().unwrap();
    let real = tmp.path().join("claims-real");
    fs::create_dir_all(&real).unwrap();
    // AC-06A-6: a symlinked journal path is refused, both for writing and reading.
    // (The symlink sits beside, not at, the root the second case uses.)
    #[cfg(unix)]
    {
        std::os::unix::fs::symlink(&real, tmp.path().join("claims-link")).unwrap();
    }
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let journal = Journal::open(&tmp.path().join("claims-link")).unwrap_err();
    assert_eq!(journal, "journal path is a symlink; refused");
    let seen = show(&tmp.path().join("claims-link"), "0123456789abcdef").unwrap_err();
    assert_eq!(seen, "journal path is a symlink; refused");

    // A symlinked journal.log inside an honest directory is refused too.
    let root = journal_root(tmp.path());
    fs::create_dir_all(&root).unwrap();
    #[cfg(unix)]
    std::os::unix::fs::symlink(real.join("elsewhere.log"), root.join("journal.log")).unwrap();
    let journal = lock_journal(&root);
    let err = acquire(
        &journal,
        &lib_req("agent-x", &lease),
        worktree(),
        &client(&RecordedTransport::new(vec![]), &clock),
        &clock,
    )
    .unwrap_err();
    assert_eq!(err, "journal path is a symlink; refused");
}

#[test]
fn claim_busy_never_waits() {
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let first = Journal::open(&root).unwrap();
    assert!(first.try_lock().unwrap());
    let second = Journal::open(&root).unwrap();
    // REQ-06A-05, asserted strictly: WouldBlock becomes `busy` at once while held.
    assert!(
        !second.try_lock().unwrap(),
        "a held lock must refuse another handle"
    );
    drop(first);
    // The release side is the one macOS answers late under parallel load (see
    // lock_journal); a bounded caller-side retry keeps the assertion about the
    // semantics — released once the holder is gone — without weakening the product.
    for _ in 0..200 {
        if second.try_lock().unwrap() {
            return;
        }
        thread::sleep(Duration::from_millis(10));
    }
    panic!("the lock is released when the holder drops (not observed within 2s)");
}

#[test]
fn claim_show_is_byte_identical() {
    // Library-level determinism: the same journal read twice serializes identically and
    // touches nothing (the CLI-level byte test is claim_show_idempotent below).
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let exchanges = happy_exchanges(900, "agent-x", &lease);
    acquire_in(&root, "agent-x", &lease, exchanges, &clock);
    let key = task_key(&task_ref());
    let a = serde_json::to_string(&show(&root, &key).unwrap()).unwrap();
    let b = serde_json::to_string(&show(&root, &key).unwrap()).unwrap();
    assert_eq!(a, b);
    assert!(a.contains("cross-host-eventual"));
    assert!(a.contains("\"cross_host_atomic\":false"));
    // An unknown key reads as unclaimed, single-host.
    let none = show(&root, "ffffffffffffffff").unwrap();
    assert_eq!(none["outcome"], "unclaimed");
    assert_eq!(none["coordination_scope"], "single-host");
}

// ---- CLI level: stub gh, real git, real processes ----

fn git(dir: &Path, args: &[&str]) {
    let status = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?} failed");
}

fn git_repo(dir: &Path) {
    fs::create_dir_all(dir).unwrap();
    git(dir, &["init", "-q"]);
    git(
        dir,
        &[
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
            "commit",
            "--allow-empty",
            "-q",
            "-m",
            "init",
        ],
    );
}

/// A stub `gh` modelling read-your-writes: comment-list GETs answer `CLAIM_STUB_JSON`
/// until a comment POST happens, after which they show the posted claim comment with
/// `CLAIM_STUB_ID` (the claim line body is JSON-safe by construction). Everything else is
/// the 201 comment creation.
fn stub_gh(dir: &Path) -> PathBuf {
    let script = "#!/bin/sh\n# claim test stub for gh api (stateful: read-your-writes)\ncase \"$*\" in\n  *per_page=100*)\n    if [ -n \"${CLAIM_STUB_JSON:-}\" ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$CLAIM_STUB_JSON\";\n    elif [ -f \"$CLAIM_STUB_DIR/posted\" ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n[{\"id\":%s,\"body\":\"%s\"}]' \"${CLAIM_STUB_ID:-4242}\" \"$CLAIM_STUB_BODY\";\n    else printf 'HTTP/2.0 200 OK\\r\\n\\r\\n[]'; fi;;\n  *)\n    : > \"$CLAIM_STUB_DIR/posted\"\n    printf 'HTTP/2.0 201 Created\\r\\n\\r\\n{\"id\":%s}' \"${CLAIM_STUB_ID:-4242}\";;\nesac\n";
    let path = dir.join("gh-stub");
    common::write_executable(&path, script);
    path
}

fn now_utc() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

fn cli_req(lease_until: &str, session: &str) -> String {
    json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com",
            "owner": "modu-ai", "repo": "moai-adk",
            "repository_id": 9001, "issue_id": 123456789, "issue_number": 2005,
        },
        "logical_key": "06-A",
        "owner": {"kind": "session", "id": session},
        "attempt": 1,
        "base_revision": "a".repeat(40),
        "write_scope": ["src/claim.rs"],
        "spec_ref": {"id": "SPEC-RC1-06-A", "revision": "r1", "digest": null},
        "lease_until": lease_until,
        "projection_intents": [],
    })
    .to_string()
}

fn claim_body(session: &str, lease_until: &str) -> String {
    format!("moai-claim: session={session} epoch=1 lease_until={lease_until}")
}

fn run_claim(
    repo: &Path,
    args: &[&str],
    stdin: Option<&str>,
    home: &Path,
    gh: &Path,
    stub_body: &str,
) -> std::process::Output {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .current_dir(repo)
        .env("HOME", home)
        .env("MOAI_FORGE_GH_BIN", gh)
        // The stateful stub keeps its posted marker next to itself and shows the claim
        // body it was given on every post-POST re-read (read-your-writes).
        .env("CLAIM_STUB_DIR", gh.parent().unwrap())
        .env("CLAIM_STUB_BODY", stub_body)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(if stdin.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    if let (Some(text), Some(mut sin)) = (stdin, child.stdin.take()) {
        // The child may exit before reading stdin; only a broken pipe is tolerated (#2024).
        match sin.write_all(text.as_bytes()) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
    }
    child.wait_with_output().unwrap()
}

fn envelope(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn journal_dir(repo: &Path) -> PathBuf {
    repo.join(".git").join(".moai-cli").join("claims")
}

#[test]
fn claim_show_idempotent() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    git_repo(&repo);
    let gh = stub_gh(tmp.path());
    let lease = format_utc(now_utc() + 3600);
    let out = run_claim(
        &repo,
        &["claim-acquire", "."],
        Some(&cli_req(&lease, "agent-6-742cb6")),
        &home,
        &gh,
        &claim_body("agent-6-742cb6", &lease),
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["operation"], json!("claim.acquire"));
    assert_eq!(v["result"]["outcome"], json!("claimed"));
    assert_eq!(
        v["result"]["coordination_scope"],
        json!("cross-host-eventual")
    );
    assert_eq!(v["result"]["cross_host_atomic"], json!(false));
    let key = v["result"]["record"]["task_key"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(
        v["result"]["record"]["claim_token"],
        json!("Moai-Claim: 2005#4242")
    );
    let log = journal_dir(&repo).join("journal.log");
    let before = (
        fs::read(&log).unwrap(),
        fs::metadata(&log).unwrap().modified().unwrap(),
    );

    // AC-06A-5: two shows are byte-identical; the journal is not touched.
    let mut outputs = Vec::new();
    for _ in 0..2 {
        let o = run_claim(
            &repo,
            &["claim-show", ".", "--task", &key],
            None,
            &home,
            &gh,
            &claim_body("agent-6-742cb6", &lease),
        );
        assert!(o.status.success());
        outputs.push(o.stdout.clone());
    }
    assert_eq!(outputs[0], outputs[1]);
    let v2: Value = serde_json::from_slice(&outputs[0]).unwrap();
    assert_eq!(v2["result"]["outcome"], json!("held"));
    assert_eq!(v2["result"]["record"]["epoch"], json!(1));
    let after = (
        fs::read(&log).unwrap(),
        fs::metadata(&log).unwrap().modified().unwrap(),
    );
    assert_eq!(before.0, after.0, "journal bytes unchanged");
    assert_eq!(before.1, after.1, "journal mtime unchanged");

    // Release with the recorded epoch; a stale epoch is refused with kind conflict.
    let o = run_claim(
        &repo,
        &["claim-release", ".", "--task", &key, "--epoch", "2"],
        None,
        &home,
        &gh,
        &claim_body("agent-6-742cb6", &lease),
    );
    assert!(!o.status.success());
    let v3 = envelope(&o);
    assert_eq!(v3["error"]["kind"], json!("conflict"));
    assert!(v3["error"]["message"]
        .as_str()
        .unwrap()
        .contains("stale epoch"));
    let o = run_claim(
        &repo,
        &["claim-release", ".", "--task", &key, "--epoch", "1"],
        None,
        &home,
        &gh,
        &claim_body("agent-6-742cb6", &lease),
    );
    assert!(o.status.success());
    let v4 = envelope(&o);
    assert_eq!(v4["result"]["outcome"], json!("released"));
    assert_eq!(v4["result"]["record"]["state"], json!("cancelled"));
}

#[test]
fn claim_concurrent_exclusive() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    git_repo(&repo);
    let gh = stub_gh(tmp.path());
    let lease = format_utc(now_utc() + 3600);
    let req = cli_req(&lease, "agent-x");

    // AC-06A-1: N >= 8 parallel claim attempts, each retrying the journal lock-busy
    // answer (the caller-side waiting policy — the product itself never waits,
    // REQ-06A-05). #2345: a held lock answers the failure envelope — kind `unavailable`,
    // retryable true, exit 1 — the same answer as every other journal-writing op.
    let n = 8u32;
    let outcomes: Vec<Value> = thread::scope(|s| {
        let handles: Vec<_> = (0..n)
            .map(|_| {
                s.spawn(|| {
                    for _ in 0..300 {
                        let out = run_claim(
                            &repo,
                            &["claim-acquire", "."],
                            Some(&req),
                            &home,
                            &gh,
                            &claim_body("agent-x", &lease),
                        );
                        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
                        if v["error"]["kind"] == json!("unavailable") {
                            assert_eq!(v["ok"], json!(false));
                            assert_eq!(v["error"]["retryable"], json!(true));
                            assert!(!out.status.success());
                            thread::sleep(Duration::from_millis(5));
                            continue;
                        }
                        assert!(
                            out.status.success(),
                            "{}",
                            String::from_utf8_lossy(&out.stderr)
                        );
                        return v;
                    }
                    panic!("the journal stayed busy for 300 attempts");
                })
            })
            .collect();
        handles
            .into_iter()
            .map(|h| h.join().unwrap())
            .collect::<Vec<_>>()
    });

    let claimed = outcomes
        .iter()
        .filter(|v| v["result"]["outcome"] == json!("claimed"))
        .count();
    let already = outcomes
        .iter()
        .filter(|v| v["result"]["outcome"] == json!("already_owned"))
        .count();
    assert_eq!(claimed, 1, "exactly one winner");
    assert_eq!(already, (n - 1) as usize, "every loser is already_owned");
    // The journal holds exactly one record; only one comment POST ever happened.
    let log = fs::read_to_string(journal_dir(&repo).join("journal.log")).unwrap();
    assert_eq!(log.lines().count(), 1);
    let record: Value = serde_json::from_str(log.trim()).unwrap();
    assert_eq!(record["event"], "acquire");
    assert_eq!(record["record"]["epoch"], 1);
}

#[test]
fn claim_go_paths_untouched() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    let home = tmp.path().join("home");
    // A fake Go-generation state tree in the fake home and the project (REQ-06A-09).
    fs::create_dir_all(home.join(".moai")).unwrap();
    fs::write(home.join(".moai/sentinel.txt"), b"go state").unwrap();
    git_repo(&repo);
    fs::create_dir_all(repo.join(".moai")).unwrap();
    fs::write(repo.join(".moai/sentinel.txt"), b"go state").unwrap();

    let hash_tree = |root: &Path| -> Vec<String> {
        let mut out = Vec::new();
        fn walk(root: &Path, dir: &Path, out: &mut Vec<String>) {
            for e in fs::read_dir(dir).unwrap() {
                let p = e.unwrap().path();
                let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
                if p.is_dir() {
                    walk(root, &p, out);
                } else {
                    let bytes = fs::read(&p).unwrap();
                    use moai_cli::digest;
                    out.push(format!("{rel} {}", digest(&bytes)));
                }
            }
        }
        walk(root, root, &mut out);
        out.sort();
        out
    };
    let home_before = hash_tree(&home);
    let repo_before = hash_tree(&repo);

    let gh = stub_gh(tmp.path());
    let lease = format_utc(now_utc() + 3600);
    let out = run_claim(
        &repo,
        &["claim-acquire", "."],
        Some(&cli_req(&lease, "agent-6-742cb6")),
        &home,
        &gh,
        &claim_body("agent-6-742cb6", &lease),
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );

    // AC-06A-4: `.moai/` and `~/.moai/` hashes are identical; nothing under `.claude/`.
    assert_eq!(hash_tree(&home), home_before, "fake ~/.moai untouched");
    let repo_after = hash_tree(&repo);
    let before: std::collections::HashSet<String> = repo_before.into_iter().collect();
    let after: std::collections::HashSet<String> = repo_after.into_iter().collect();
    let removed: Vec<_> = before.difference(&after).collect();
    let added: Vec<_> = after.difference(&before).collect();
    assert!(removed.is_empty(), "files vanished or changed: {removed:?}");
    assert!(!added.is_empty(), "no new files recorded");
    for f in &added {
        assert!(f.contains(".moai-cli"), "new file outside .moai-cli: {f}");
        assert!(!f.contains("/.moai/") && !f.contains(".claude"), "{f}");
    }
}

/// REQ-B2-01 (#2169): the record stores the four ownership fields separately — `owner`
/// (the execution owner) is the only field the journal consults for exclusivity; the
/// assignee, the provider session and the integration owner are recorded projections
/// and none of them is a lock. Records written before #2169 deserialize unchanged.
#[test]
fn claim_record_four_ownership_fields_are_separate_and_not_locks() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    let mut req = lib_req("agent-x", &lease);
    req.assignee = Some("octocat".into());
    req.provider_session = Some("sess-1234".into());
    let (out, _t) = acquire_with(
        tmp.path(),
        req,
        happy_exchanges(900, "agent-x", &lease),
        &clock,
    );
    assert_eq!(out.outcome, "claimed");

    // The four fields are distinct JSON keys on the record.
    let record = out.record.as_ref().unwrap();
    let v = serde_json::to_value(record).unwrap();
    assert_eq!(
        v["owner"]["id"],
        json!("agent-x"),
        "execution owner (the alias target)"
    );
    assert_eq!(v["assignee"], json!("octocat"));
    assert_eq!(v["provider_session"], json!("sess-1234"));
    assert!(
        v.get("integration_owner").is_none(),
        "acquire never sets the integration owner — the leader gate records it later"
    );

    // Exclusivity consults `owner` only: a second acquire whose ASSIGNEE matches but
    // whose owner differs is still already_owned by agent-x — the assignee granted
    // nothing and no second record appeared.
    let mut second = lib_req("agent-y", &lease);
    second.assignee = Some("octocat".into());
    let journal = lock_journal(tmp.path());
    let second_out = acquire(
        &journal,
        &second,
        worktree(),
        &client(&RecordedTransport::new(vec![]), &clock),
        &clock,
    )
    .unwrap();
    assert_eq!(second_out.outcome, "already_owned");
    assert_eq!(
        second_out.record.as_ref().unwrap().owner.id,
        "agent-x",
        "the assignee did not become a lock or an ownership claim"
    );
    drop(journal);

    // Old-record compatibility: a record JSON without the three additive fields
    // deserializes with them as None and the execution owner intact.
    let mut old = serde_json::to_value(out.record.as_ref().unwrap()).unwrap();
    let obj = old.as_object_mut().unwrap();
    obj.remove("assignee");
    obj.remove("provider_session");
    obj.remove("integration_owner");
    let parsed: moai_cli::claim::ClaimRecord = serde_json::from_value(old).unwrap();
    assert_eq!(parsed.owner.id, "agent-x");
    assert_eq!(parsed.assignee, None);
    assert_eq!(parsed.provider_session, None);
    assert_eq!(parsed.integration_owner, None);
}

// ---- #2166 (rc.2 A7): default lease, explicit lease wins, expiry shown ----

/// REQ-A7-01: an absent `lease_until` becomes now + 4h, in the posted comment and the
/// record alike.
#[test]
fn lease_defaults_to_four_hours() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let expected = format_utc(CLOCK0 + 4 * 60 * 60);
    assert_eq!(moai_cli::claim::DEFAULT_LEASE, 4 * 60 * 60);
    let (out, t) = acquire_in(
        tmp.path(),
        "agent-x",
        "",
        happy_exchanges(910, "agent-x", &expected),
        &clock,
    );
    assert_eq!(out.outcome, "claimed");
    let record = out.record.unwrap();
    assert_eq!(record.lease_until.as_deref(), Some(expected.as_str()));
    let sent = t.sent.borrow();
    let posted = sent.iter().find(|r| r.method == "POST").unwrap();
    assert!(
        posted.body.as_ref().unwrap().to_string().contains(&format!(
            "moai-claim: session=agent-x epoch=1 lease_until={expected}"
        )),
        "{:?}",
        posted.body
    );
}

/// The caller's explicit `lease_until` wins over the default.
#[test]
fn explicit_lease_wins_over_default() {
    let tmp = tempfile::tempdir().unwrap();
    let clock = FakeClock::at(CLOCK0);
    let explicit = format_utc(CLOCK0 + 600);
    let (out, _) = acquire_in(
        tmp.path(),
        "agent-x",
        &explicit,
        happy_exchanges(911, "agent-x", &explicit),
        &clock,
    );
    assert_eq!(out.outcome, "claimed");
    assert_eq!(
        out.record.unwrap().lease_until.as_deref(),
        Some(explicit.as_str())
    );
}

/// AC-2: `claim-show` marks an expired active record `expired: true` and moves nothing;
/// the same read twice at one instant is byte-identical.
#[test]
fn show_marks_expired_lease_without_takeover() {
    let tmp = tempfile::tempdir().unwrap();
    let root = journal_root(tmp.path());
    let clock = FakeClock::at(CLOCK0);
    let lease = format_utc(CLOCK0 + 3600);
    acquire_in(
        &root,
        "agent-x",
        &lease,
        happy_exchanges(912, "agent-x", &lease),
        &clock,
    );
    let key = task_key(&task_ref());
    let log = root.join("journal.log");
    let before = fs::read(&log).unwrap();

    let live = moai_cli::claim::show_at(&root, &key, CLOCK0 + 60).unwrap();
    assert_eq!(live["expired"], json!(false));
    let late = moai_cli::claim::show_at(&root, &key, CLOCK0 + 3601).unwrap();
    assert_eq!(late["expired"], json!(true));
    assert_eq!(late["outcome"], "held", "expiry moves no ownership");
    assert_eq!(late["record"]["epoch"], json!(1));
    assert_eq!(
        serde_json::to_string(&late).unwrap(),
        serde_json::to_string(&moai_cli::claim::show_at(&root, &key, CLOCK0 + 3601).unwrap())
            .unwrap()
    );
    assert_eq!(fs::read(&log).unwrap(), before, "journal untouched");
    let none = moai_cli::claim::show_at(&root, "ffffffffffffffff", CLOCK0).unwrap();
    assert_eq!(none["expired"], json!(false));
}
