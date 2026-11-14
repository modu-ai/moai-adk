//! The claim acquire closures and `queue-free` (#2675) over recorded exchanges
//! (`RecordedTransport` with an injected clock — no network, no sleeping), in the fixture style
//! of `tests/claim_cli.rs`. `RecordedTransport` asserts method and URL in order, so an
//! unexpected request fails the test. `queue-next` (rc.2 A2 #2161) is retired (#2901):
//! `lane-next` names the next startable card and the lane claims it with `claim-acquire`.
mod common;

use common::{exchange, json_response as resp};
use moai_cli::claim::{
    acquire, task_key, ClaimRequest, Journal, Owner, SpecRef, TaskRef, WorktreeId,
};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, RecordedTransport};
use moai_cli::forge::{Client, Config};
use moai_cli::queue::{free, FreeRequest};
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

/// Exercise acquire directly: queue prefiltering would hide invalid closure targets
/// before the claim consumer could demonstrate its own live-owner safety boundary.
#[test]
fn acquire_honors_literal_release_and_handoff_closures() {
    let req = ClaimRequest::new(
        1,
        TaskRef::new("github", "github.com", "acme", "widgets", 101, 7000, 7),
        "A1".to_string(),
        Owner::new("session", "recipient"),
        "a".repeat(40),
        vec!["src/queue.rs".into()],
        SpecRef {
            id: "SPEC-RC2-A2".into(),
            revision: "r1".into(),
            digest: None,
        },
    )
    .lease_until(LEASE.to_string());
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

/// #2901 AC8: `queue-next` is retired — `lane-next` names the next startable card (read only)
/// and the lane claims it with `claim-acquire`. The word is a usage error with the `cli`
/// envelope, never the old operation's, and the full command list does not name it.
#[test]
fn queue_next_is_not_a_command() {
    let home = tempfile::tempdir().unwrap();
    let run = |args: &[&str]| {
        let out = std::process::Command::new(common::MOAI_BIN)
            .args(args)
            .current_dir(home.path())
            .env("HOME", home.path())
            .stdin(std::process::Stdio::null())
            .output()
            .unwrap();
        (
            out.status.code(),
            String::from_utf8_lossy(&out.stdout).into_owned(),
            String::from_utf8_lossy(&out.stderr).into_owned(),
        )
    };
    let (code, stdout, stderr) = run(&["queue-next", "."]);
    assert_eq!(code, Some(2), "{stdout}{stderr}");
    assert_eq!(
        stdout,
        "{\"protocol_version\":1,\"operation\":\"cli\",\"ok\":false,\"error\":{\"kind\":\"usage\",\"message\":\"unknown command; see --help\",\"retryable\":false}}\n"
    );
    assert!(stderr.contains("unrecognized subcommand"), "{stderr}");
    let (code, help, stderr) = run(&["help", "--all"]);
    assert_eq!(code, Some(0), "{stderr}");
    assert!(help.contains("queue-free"), "{help}");
    assert!(help.contains("lane-next"), "{help}");
    assert!(!help.contains("queue-next"), "{help}");
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
