//! `card fold` (SPEC-20261124-card-fold-stale): a duplicate card moves to `hold` with one
//! `card.fold` event and one `relates` link to the card it folds into; its notes and criteria stay
//! readable, `card next` never offers it and `card take` refuses it naming the target. A fold that
//! cannot happen writes nothing, and a fold that ends a lapsed reservation is the leader's.
//!
//! The fold is written through the ledger function the `card fold` action calls; every case runs
//! in a temporary HOME and the real home state is not read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger::close::{self, CloseError};
use moai_cli::ledger::{self, WriteLedger};
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

/// A fold that must be refused: the card, its target and what the refusal must look like.
type FoldCase = (i64, i64, fn(&CloseError) -> bool);

struct World {
    home: TempHome,
    p: String,
}

impl World {
    /// A registered project with `n` ready cards, C-1 to C-n, each with its own scope.
    fn new(n: usize) -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let w = World {
            p: project.to_str().unwrap().to_string(),
            home,
        };
        for i in 1..=n {
            let scope = format!("apps/card{i}/");
            let added = w.card(
                None,
                &[
                    "add", "--title", "fixture", "--kind", "task", "--scope", &scope, "--ac", "one",
                ],
            );
            let name = added["name"].as_str().unwrap().to_string();
            w.card(None, &["ready", &name]);
        }
        w
    }

    fn output(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.stdin(Stdio::null());
        cmd.output().expect("run moai")
    }

    fn card(&self, session: Option<&str>, args: &[&str]) -> Value {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.p]);
        let out = self.output(session, &all);
        let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
            panic!(
                "stdout is not JSON ({e}): {} {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(out.status.code(), Some(0), "{v}");
        v["result"].clone()
    }

    fn note(&self, card: &str, text: &str) {
        let mut cmd = self.home.command(&[
            "card",
            "note",
            card,
            "--kind",
            "comment",
            "--project",
            &self.p,
        ]);
        cmd.stdin(Stdio::piped()).stdout(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        child
            .stdin
            .take()
            .unwrap()
            .write_all(text.as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        assert_eq!(out.status.code(), Some(0));
    }

    fn write(&self) -> WriteLedger {
        ledger::open_write(self.home.path(), "demo").unwrap()
    }

    fn fold(&self, card: i64, into: i64, by: &str) -> Result<(), CloseError> {
        close::fold(
            &mut self.write(),
            card,
            into,
            "same work as the target",
            by,
            None,
        )
    }

    fn sql(&self, statement: &str) {
        self.write()
            .with_write_tx(|tx, _| {
                tx.execute_batch(statement)?;
                Ok(())
            })
            .unwrap();
    }

    fn rows(&self, sql: &str) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut st = l.conn().prepare(sql).unwrap();
        st.query_map([], |r| r.get::<_, String>(0))
            .unwrap()
            .collect::<Result<_, _>>()
            .unwrap()
    }

    /// The parts of the ledger a fold or a withdraw writes, as text.
    fn dump(&self) -> Vec<String> {
        let mut all = Vec::new();
        for sql in [
            "SELECT card_id || ':' || state || ':' || rev FROM card ORDER BY card_id",
            "SELECT event_id || ':' || kind || ':' || ifnull(card_id, '') || ':' || payload \
             FROM event ORDER BY event_id",
            "SELECT from_card || '>' || ifnull(to_card, '') || ':' || kind || ':' || state \
             FROM card_link ORDER BY link_id",
            "SELECT card_id || ':' || generation || ':' || status || ':' || lease_until \
             FROM reservation ORDER BY reservation_id",
        ] {
            all.extend(self.rows(sql));
        }
        all
    }
}

/// AC-1: the folded card is `hold` with one event and one link; its notes and criteria stay
/// readable; `next` offers the target and not the folded card; `take` refuses naming the target.
#[test]
fn a_folded_card_is_hold_with_an_event_and_a_link_and_is_never_offered() {
    let w = World::new(2);
    w.note("C-1", "kept after the fold");
    w.fold(1, 2, "lane").unwrap();

    assert_eq!(w.rows("SELECT state FROM card WHERE card_id = 1"), ["hold"]);
    let events = w.rows("SELECT payload FROM event WHERE card_id = 1 AND kind = 'card.fold'");
    assert_eq!(events.len(), 1, "{events:?}");
    let payload: Value = serde_json::from_str(&events[0]).unwrap();
    assert_eq!(payload["into"], json!("C-2"), "{payload}");
    assert_eq!(payload["reason"], json!("same work as the target"));
    assert_eq!(payload["by"], json!("lane"));
    assert_eq!(
        w.rows(
            "SELECT CAST(count(*) AS TEXT) FROM card_link \
             WHERE from_card = 1 AND to_card = 2 AND kind = 'relates' AND state = 'active'"
        ),
        ["1"]
    );

    let shown = w.card(None, &["show", "C-1"]);
    assert_eq!(shown["state"], json!("hold"), "{shown}");
    assert_eq!(shown["criteria"].as_array().unwrap().len(), 1, "{shown}");
    assert!(
        shown["notes"].as_array().unwrap().iter().any(|n| n["body"]
            .as_str()
            .is_some_and(|b| b.contains("kept after the fold"))),
        "{shown}"
    );

    let out = w.output(Some(S_A), &["card", "next", &w.p]);
    assert_eq!(out.status.code(), Some(0));
    let listing: Value = serde_json::from_slice::<Value>(&out.stdout).unwrap()["result"].clone();
    let named = |key: &str| -> Vec<String> {
        listing[key]
            .as_array()
            .unwrap()
            .iter()
            .map(|e| e["name"].as_str().unwrap().to_string())
            .collect()
    };
    assert_eq!(named("ready"), ["C-2"], "{listing}");
    assert!(!named("skipped").contains(&"C-1".to_string()), "{listing}");

    let out = w.output(Some(S_A), &["card", "take", "C-1", "--project", &w.p]);
    assert_ne!(out.status.code(), Some(0));
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(
        text.contains("folded into C-2"),
        "take names the target: {text}"
    );
}

/// SPEC-20261124-card-help-gates AC-6 (REQ-4): `take` of a folded card says `folded into C-2`
/// and never calls the card withdrawn (it was folded, not withdrawn; AUDIT #444 N3).
#[test]
fn take_of_a_folded_card_does_not_say_withdrawn() {
    let w = World::new(2);
    w.fold(1, 2, "lane").unwrap();
    let out = w.output(Some(S_A), &["card", "take", "C-1", "--project", &w.p]);
    assert_ne!(out.status.code(), Some(0));
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(text.contains("folded into C-2"), "{text}");
    assert!(!text.contains("withdrawn"), "{text}");
}

/// AC-2: every fold that cannot happen is refused and the ledger is the same before and after;
/// the same card folded twice is refused the second time.
#[test]
fn a_fold_that_cannot_happen_writes_nothing() {
    let w = World::new(4);
    // C-5 is `done`; the insert is the only way to make one without a landing.
    w.sql(
        "INSERT INTO card (title, state, created_by, created_at, updated_at, closed_at) \
         VALUES ('done card', 'done', 'fixture', 1, 1, 1)",
    );
    // C-3 is withdrawn: `hold`.
    close::withdraw(&mut w.write(), 3, "not needed", "lane", None).unwrap();
    // C-4 is held by a live lease.
    w.card(Some(S_A), &["take", "C-4"]);

    let cases: [FoldCase; 7] = [
        (1, 1, |e| matches!(e, CloseError::BadTarget { .. })),
        (1, 99, |e| matches!(e, CloseError::BadTarget { .. })),
        (99, 2, |e| matches!(e, CloseError::CardNotFound(99))),
        (5, 2, |e| matches!(e, CloseError::WrongState { .. })),
        (1, 3, |e| matches!(e, CloseError::BadTarget { .. })),
        (3, 2, |e| matches!(e, CloseError::WrongState { .. })),
        (4, 2, |e| matches!(e, CloseError::Held { .. })),
    ];
    for (card, into, expected) in cases {
        let before = w.dump();
        let err = w.fold(card, into, "lane").unwrap_err();
        assert!(expected(&err), "C-{card} into C-{into}: {err:?}");
        assert!(w.dump() == before, "C-{card} into C-{into} wrote something");
    }

    w.fold(1, 2, "lane").unwrap();
    let before = w.dump();
    let err = w.fold(1, 2, "lane").unwrap_err();
    assert!(matches!(err, CloseError::WrongState { .. }), "{err:?}");
    assert!(w.dump() == before, "the second fold wrote something");

    let long = "x".repeat(501);
    for reason in ["", "   ", long.as_str(), "two\nlines"] {
        let before = w.dump();
        let err = close::fold(&mut w.write(), 2, 1, reason, "lane", None).unwrap_err();
        assert!(
            matches!(err, CloseError::InvalidInput(_)),
            "{reason:?}: {err:?}"
        );
        assert!(w.dump() == before, "{reason:?} wrote something");
    }
}

/// REQ-8/9: a fold of a card whose lease ended ends that reservation, and only for the leader.
#[test]
fn a_fold_of_a_card_with_an_ended_lease_is_the_leaders_and_ends_the_reservation() {
    let w = World::new(2);
    w.card(Some(S_A), &["take", "C-1"]);
    w.sql("UPDATE reservation SET lease_until = unixepoch() - 5 WHERE card_id = 1");

    let before = w.dump();
    let err = w.fold(1, 2, "lane").unwrap_err();
    assert!(matches!(err, CloseError::LeaderOnly { .. }), "{err:?}");
    assert!(w.dump() == before, "a non-leader fold wrote something");

    w.fold(1, 2, "leader").unwrap();
    assert_eq!(
        w.rows("SELECT status FROM reservation WHERE card_id = 1"),
        ["released"]
    );
    assert_eq!(w.rows("SELECT state FROM card WHERE card_id = 1"), ["hold"]);
    let events = w.rows("SELECT payload FROM event WHERE card_id = 1 AND kind = 'card.fold'");
    let payload: Value = serde_json::from_str(&events[0]).unwrap();
    assert_eq!(payload["released"]["holder"], json!(S_A), "{payload}");
    assert_eq!(payload["released"]["generation"], json!(1), "{payload}");
    assert!(payload["released"]["lease_until"].is_i64(), "{payload}");
}

/// AC-1 through the wired action: `card fold` answers the folded card and its target, a second
/// call is refused, a target that is `hold` is refused, and `card ready` takes a folded card back.
#[test]
fn the_wired_fold_action_folds_once_and_ready_takes_the_card_back() {
    let w = World::new(3);
    let fold = |card: &str, into: &str| {
        w.output(
            None,
            &[
                "card",
                "fold",
                card,
                "--into",
                into,
                "--reason",
                "same work",
                "--by",
                "lane",
                &w.p,
            ],
        )
    };
    let out = fold("C-1", "C-2");
    assert_eq!(out.status.code(), Some(0));
    let v: Value = serde_json::from_slice::<Value>(&out.stdout).unwrap()["result"].clone();
    assert_eq!(v["card"], json!("C-1"), "{v}");
    assert_eq!(v["into"], json!("C-2"), "{v}");
    assert_eq!(v["state"], json!("hold"), "{v}");

    let before = w.dump();
    let out = fold("C-1", "C-2");
    assert_eq!(out.status.code(), Some(1), "the second fold is refused");
    assert!(
        String::from_utf8_lossy(&out.stdout).contains("card fold refused"),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let out = fold("C-3", "C-1");
    assert_eq!(
        out.status.code(),
        Some(1),
        "a target that is hold is refused"
    );
    assert!(w.dump() == before, "a refused fold wrote something");

    w.card(None, &["ready", "C-1"]);
    assert_eq!(
        w.rows("SELECT state FROM card WHERE card_id = 1"),
        ["ready"]
    );
}

/// AC-1: the default `card list` hides a folded card as it hides a withdrawn one, and `--all`
/// shows it.
#[test]
fn the_default_list_hides_a_folded_card_until_all_is_asked() {
    let w = World::new(2);
    w.fold(1, 2, "lane").unwrap();
    let names = |extra: &[&str]| -> Vec<String> {
        let mut args = vec!["card", "list"];
        args.extend_from_slice(extra);
        args.push(&w.p);
        let out = w.output(None, &args);
        assert_eq!(out.status.code(), Some(0));
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        v["result"]["cards"]
            .as_array()
            .unwrap()
            .iter()
            .map(|c| c["name"].as_str().unwrap().to_string())
            .collect()
    };
    assert_eq!(names(&[]), ["C-2"]);
    assert_eq!(names(&["--all"]), ["C-1", "C-2"]);
}
