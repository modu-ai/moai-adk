//! `moai card query` (SPEC-20261126-card-read-queries): the leader's reads of many cards, the
//! live reservations and notes by number, in one call that writes nothing.
//!
//! Every command runs against a temporary HOME; the real home state is not read or written
//! (SPEC-20261126-card-read-queries AC-8). Cards and reservations are made by the real commands;
//! notes and verdicts by the ledger's own writers; SQL sets only what no command gives (a
//! reservation status, an unreadable verdict row, an imported note).
#![cfg(unix)]

use crate::common;

use common::home::TempHome;
use moai_cli::ledger;
use moai_cli::ledger::workflow::{self, NewVerdict, VerdictReceipt, VerdictStep};
use serde_json::{json, Value};
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
/// Sessions that never hold a card, so they may write verdicts.
const AUDITOR: &str = "cccccccc-0000-4000-8000-00000000000c";
const AUDITOR_2: &str = "dddddddd-0000-4000-8000-00000000000d";
const EVIDENCE: &str = "read the receipt";

fn sha(digit: char) -> String {
    digit.to_string().repeat(40)
}

struct World {
    home: TempHome,
    project: String,
}

/// A HOME with a project and a registered ledger.
fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo").to_str().unwrap().to_string();
    ledger::register(home.path(), "demo").unwrap();
    World { home, project }
}

impl World {
    /// `moai card <args> --project <project>`, as `session`.
    fn card(&self, session: Option<&str>, args: &[&str]) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        cmd.output().expect("run moai")
    }

    /// `card query <extra>`, with no session: a read needs none.
    fn query(&self, extra: &[&str]) -> Output {
        let mut args = vec!["query"];
        args.extend_from_slice(extra);
        self.card(None, &args)
    }

    /// The result of a `card query` call that must succeed.
    fn read(&self, extra: &[&str]) -> Value {
        result(&self.query(extra))
    }

    fn show(&self, card: &str) -> Value {
        result(&self.card(None, &["show", card]))
    }

    /// A `draft` card with its criteria and a write scope of its own; the new card's number.
    fn add(&self, title: &str, criteria: &[&str]) -> i64 {
        let scope = format!("src/{title}.rs");
        let mut args = vec!["add", "--title", title, "--scope", &scope];
        for text in criteria {
            args.extend_from_slice(&["--ac", text]);
        }
        result(&self.card(None, &args))["card_id"].as_i64().unwrap()
    }

    /// `card ready` and `card take` by `session`.
    fn take(&self, card: i64, session: &str) {
        let name = format!("C-{card}");
        result(&self.card(None, &["ready", &name]));
        result(&self.card(Some(session), &["take", &name]));
    }

    /// One note, written by the ledger's own writer; its number.
    fn note(&self, card: i64, author: &str, session: Option<&str>, body: &str) -> i64 {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        workflow::add_note(&mut w, card, "report", author, session, body, None)
            .unwrap()
            .note_id
    }

    /// The hand-back note that makes `head` the card's current head.
    fn hand_back(&self, card: i64, head: &str) -> i64 {
        let body = format!(
            "LANE-READY-INTEGRATION C-{card} head={head} base={}",
            sha('0')
        );
        self.note(card, "agent-a", Some(S_A), &body)
    }

    /// One verdict, written by the ledger's own writer; its event number.
    fn verdict(
        &self,
        card: i64,
        reviewer: &str,
        head: &str,
        pass: bool,
        receipt: Option<&VerdictReceipt>,
    ) -> i64 {
        let mut w = ledger::open_write(self.home.path(), "demo").unwrap();
        let base = sha('0');
        workflow::record_verdict(
            &mut w,
            card,
            &NewVerdict {
                head,
                base: &base,
                pass,
                tier: 2,
                evidence: EVIDENCE,
                receipt,
                reviewer,
                open: &[],
            },
        )
        .unwrap()
        .event_id
    }

    /// One statement on the ledger, for a state no command gives.
    fn sql(&self, statement: &str, params: impl rusqlite::Params) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn().execute(statement, params).unwrap();
    }

    /// A `card.verdict` row that does not read as a verdict.
    fn unreadable_verdict(&self, card: i64, mark: &str) {
        self.sql(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, 1000, 'tester', ?2, 'card.verdict', ?3)",
            rusqlite::params![
                format!("test.verdict:{mark}"),
                card,
                json!({"verdict": "maybe", "head": sha('a')}).to_string()
            ],
        );
    }

    /// `count` cards made through the library, for the limits.
    fn bulk_cards(&self, count: usize) {
        for i in 1..=count {
            moai_cli::card::command::add(
                std::path::Path::new(&self.project),
                self.home.path(),
                &moai_cli::card::command::AddRequest {
                    title: &format!("bulk {i}"),
                    body: "",
                    kind: None,
                    priority: None,
                    scope: &[],
                    criteria: &[],
                    by: None,
                },
            )
            .unwrap();
        }
    }
}

fn text_of(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text_of(out)))
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text_of(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

/// A refused call: exit 1, an error and no part of an answer. The error kind and message.
fn refusal(out: &Output) -> (String, String) {
    assert_eq!(out.status.code(), Some(1), "{}", text_of(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.get("result").is_none(), "no partial answer: {v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

/// A call the argument parser refuses: exit 2 and no part of an answer.
fn usage(out: &Output) {
    assert_eq!(out.status.code(), Some(2), "{}", text_of(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!("usage"), "{v}");
    assert!(v.get("result").is_none(), "no partial answer: {v}");
}

/// The values of `field` in the list `section` of an answer.
fn column(answer: &Value, section: &str, field: &str) -> Vec<Value> {
    answer[section]
        .as_array()
        .unwrap_or_else(|| panic!("{section} is not a list: {answer}"))
        .iter()
        .map(|row| row[field].clone())
        .collect()
}

/// The note `note_id` as `card show` prints it.
fn shown_note(shown: &Value, note_id: i64) -> Value {
    shown["notes"]
        .as_array()
        .unwrap()
        .iter()
        .find(|n| n["note_id"] == json!(note_id))
        .unwrap_or_else(|| panic!("card show has no note {note_id}"))
        .clone()
}

/// What `--reservations` should say of a card, from what `card show` says of it.
fn reservation_of(shown: &Value, lease_ended: bool) -> Value {
    let r = &shown["reservation"];
    json!({
        "card": shown["name"], "card_id": shown["card_id"], "title": shown["title"],
        "state": shown["state"], "status": r["status"], "session": r["holder"],
        "generation": r["generation"], "acquired_at": r["acquired_at"],
        "lease_until": r["lease_until"], "lease_ended": lease_ended,
        "worktree": r["worktree"], "branch": r["branch"],
    })
}

// ---- AC-1: one call answers all of it ----------------------------------------------------------

#[test]
fn one_query_answers_two_cards_the_reservations_and_a_note() {
    let w = world();
    w.add("one", &["first works", "second works", "third works"]);
    w.add("two", &["it works"]);
    w.take(1, S_A);
    w.sql(
        "UPDATE card_ac SET verify = 'cargo test one' WHERE card_id = 1 AND seq = 1",
        [],
    );
    w.sql(
        "UPDATE card_ac SET status = 'met', evidence_ref = 'run 1' WHERE card_id = 1 AND seq = 2",
        [],
    );
    w.sql(
        "UPDATE card_ac SET status = 'withdrawn', withdrawn_reason = 'not wanted' \
          WHERE card_id = 1 AND seq = 3",
        [],
    );
    w.sql(
        "UPDATE reservation SET worktree = '/work/one', branch = 'spec/one' WHERE card_id = 1",
        [],
    );
    let head = sha('a');
    let note_id = w.hand_back(1, &head);
    let receipt = VerdictReceipt {
        head: head.clone(),
        steps: vec![VerdictStep {
            name: "check".to_string(),
            exit: 0,
        }],
    };
    let event = w.verdict(1, AUDITOR, &head, true, Some(&receipt));

    let r = w.read(&[
        "--card",
        "C-1,C-2",
        "--reservations",
        "--note",
        &note_id.to_string(),
    ]);
    assert_eq!(r["action"], json!("query"));
    assert_eq!(r["project_id"], json!("demo"));
    assert_eq!(r["ledger"], json!("ready"));
    assert_eq!(r["writes"], json!(false));
    assert!(r["read_at"].as_i64().unwrap() > 1_700_000_000, "{r}");

    let verdict = json!({
        "event": event, "verdict": "pass", "head": head, "base": sha('0'), "tier": 2,
        "reviewer": AUDITOR, "evidence": EVIDENCE,
        "receipt": {"head": head, "steps": [{"name": "check", "exit": 0}]},
    });
    assert_eq!(
        r["cards"][0],
        json!({
            "card": "C-1", "card_id": 1, "title": "one", "state": "active", "head": head,
            "verdict": verdict, "head_verdict": verdict, "verdicts_unreadable": 0,
            "criteria": {"open": 1, "met": 1, "withdrawn": 1},
            "open_criteria": [{"seq": 1, "text": "first works", "verify": "cargo test one"}],
        })
    );
    assert_eq!(
        r["cards"][1],
        json!({
            "card": "C-2", "card_id": 2, "title": "two", "state": "draft", "head": null,
            "verdict": null, "head_verdict": null, "verdicts_unreadable": 0,
            "criteria": {"open": 1, "met": 0, "withdrawn": 0},
            "open_criteria": [{"seq": 1, "text": "it works", "verify": ""}],
        })
    );
    assert_eq!(r["cards"].as_array().unwrap().len(), 2);

    let shown = w.show("C-1");
    assert_eq!(shown["reservation"]["worktree"], json!("/work/one"));
    assert_eq!(r["reservations"], json!([reservation_of(&shown, false)]));
    assert_eq!(column(&r, "notes", "note_id"), [json!(note_id)]);
    assert_eq!(r["notes"][0]["body"], shown_note(&shown, note_id)["body"]);
}

#[test]
fn one_query_gives_a_section_only_when_it_is_asked_for() {
    let w = world();
    w.add("one", &["it works"]);
    let note_id = w.note(1, "agent-a", Some(S_A), "hello").to_string();
    for (args, present) in [
        (vec!["--card", "C-1"], "cards"),
        (vec!["--reservations"], "reservations"),
        (vec!["--note", note_id.as_str()], "notes"),
    ] {
        let r = w.read(&args);
        for section in ["cards", "reservations", "notes"] {
            assert_eq!(
                r.get(section).is_some(),
                section == present,
                "{args:?}: {section}: {r}"
            );
        }
    }
}

#[test]
fn one_query_drops_duplicates_and_keeps_the_order_given() {
    let w = world();
    w.add("one", &[]);
    w.add("two", &[]);
    w.add("three", &[]);
    let first = w.note(1, "agent-a", None, "first");
    let second = w.note(2, "agent-b", None, "second");
    let notes = format!("{second},{first},{second}");
    let r = w.read(&[
        "--card",
        "C-2,C-1,C-2",
        "--card",
        "c-3,C-1",
        "--note",
        &notes,
    ]);
    assert_eq!(
        column(&r, "cards", "card"),
        [json!("C-2"), json!("C-1"), json!("C-3")]
    );
    assert_eq!(
        column(&r, "notes", "note_id"),
        [json!(second), json!(first)]
    );
    assert_eq!(column(&r, "notes", "card"), [json!("C-2"), json!("C-1")]);
}

// ---- AC-2: the verdicts ------------------------------------------------------------------------

#[test]
fn verdict_a_late_fail_on_an_old_head_differs_from_the_head_verdict() {
    let w = world();
    w.add("one", &[]);
    let (old, new) = (sha('a'), sha('b'));
    w.hand_back(1, &old);
    w.verdict(1, AUDITOR, &old, true, None);
    w.hand_back(1, &new);
    let on_new = w.verdict(1, AUDITOR, &new, true, None);
    let late = w.verdict(1, AUDITOR_2, &old, false, None);

    let card = w.read(&["--card", "C-1"])["cards"][0].clone();
    assert_eq!(card["head"], json!(new));
    assert_eq!(card["verdict"]["event"], json!(late));
    assert_eq!(card["verdict"]["verdict"], json!("fail"));
    assert_eq!(card["verdict"]["head"], json!(old));
    assert_eq!(card["verdict"]["receipt"], Value::Null);
    assert_eq!(card["head_verdict"]["event"], json!(on_new));
    assert_eq!(card["head_verdict"]["verdict"], json!("pass"));
    assert_eq!(card["head_verdict"]["head"], json!(new));
    assert_eq!(card["verdicts_unreadable"], json!(0));
}

#[test]
fn verdict_the_newest_of_two_on_the_head_is_the_head_verdict() {
    let w = world();
    w.add("one", &[]);
    let head = sha('a');
    w.hand_back(1, &head);
    w.verdict(1, AUDITOR, &head, true, None);
    let newest = w.verdict(1, AUDITOR_2, &head, false, None);
    let card = w.read(&["--card", "C-1"])["cards"][0].clone();
    assert_eq!(card["head_verdict"]["event"], json!(newest));
    assert_eq!(card["head_verdict"]["verdict"], json!("fail"));
    assert_eq!(card["verdict"], card["head_verdict"]);
}

#[test]
fn verdict_none_names_the_head_or_there_is_no_head() {
    let w = world();
    w.add("moved", &[]);
    w.add("nohead", &[]);
    w.add("noverdict", &[]);
    w.hand_back(1, &sha('a'));
    let old = w.verdict(1, AUDITOR, &sha('a'), true, None);
    w.hand_back(1, &sha('b'));
    let loose = w.verdict(2, AUDITOR, &sha('a'), true, None);
    w.hand_back(3, &sha('c'));

    let r = w.read(&["--card", "C-1,C-2,C-3"]);
    let cards = &r["cards"];
    assert_eq!(cards[0]["head"], json!(sha('b')));
    assert_eq!(cards[0]["verdict"]["event"], json!(old));
    assert_eq!(cards[0]["head_verdict"], Value::Null);
    assert_eq!(cards[1]["head"], Value::Null);
    assert_eq!(cards[1]["verdict"]["event"], json!(loose));
    assert_eq!(cards[1]["head_verdict"], Value::Null);
    assert_eq!(cards[2]["head"], json!(sha('c')));
    assert_eq!(cards[2]["verdict"], Value::Null);
    assert_eq!(cards[2]["head_verdict"], Value::Null);
}

/// The head is the one of the card's present work: a take after the hand-back leaves none.
#[test]
fn verdict_a_hand_back_older_than_the_present_work_gives_no_head() {
    let w = world();
    w.add("one", &["it works"]);
    let head = sha('a');
    w.hand_back(1, &head);
    assert_eq!(
        w.read(&["--card", "C-1"])["cards"][0]["head"],
        json!(head),
        "before the take the hand-back is the head"
    );
    w.take(1, S_B);
    assert_eq!(w.read(&["--card", "C-1"])["cards"][0]["head"], Value::Null);
}

#[test]
fn verdict_an_unreadable_row_leaves_both_null_and_is_counted() {
    let w = world();
    w.add("broken", &[]);
    w.add("sound", &[]);
    let head = sha('a');
    for card in [1, 2] {
        w.hand_back(card, &head);
        w.verdict(card, AUDITOR, &head, true, None);
    }
    w.unreadable_verdict(1, "one");
    let r = w.read(&["--card", "C-1,C-2"]);
    let broken = &r["cards"][0];
    assert_eq!(broken["verdict"], Value::Null, "{broken}");
    assert_eq!(broken["head_verdict"], Value::Null, "{broken}");
    assert_eq!(broken["verdicts_unreadable"], json!(1));
    assert_eq!(broken["head"], json!(head), "the head is still read");
    let sound = &r["cards"][1];
    assert_eq!(sound["verdict"]["verdict"], json!("pass"));
    assert_eq!(sound["head_verdict"]["verdict"], json!("pass"));
    assert_eq!(sound["verdicts_unreadable"], json!(0));

    w.unreadable_verdict(1, "two");
    let again = w.read(&["--card", "C-1"]);
    assert_eq!(again["cards"][0]["verdicts_unreadable"], json!(2));
}

// ---- AC-3: the live reservations ---------------------------------------------------------------

#[test]
fn reservations_lists_the_three_live_statuses_by_card_number() {
    let w = world();
    for title in [
        "held",
        "candidate",
        "review",
        "released",
        "reclaimed",
        "late",
    ] {
        w.add(title, &["it works"]);
    }
    // Taken newest card first, so the order of the answer is not the order of the takes.
    for card in (1..=6_i64).rev() {
        w.take(card, &format!("aaaaaaaa-0000-4000-8000-{card:012}"));
    }
    let status = |card: i64, status: &str| {
        w.sql(
            "UPDATE reservation SET status = ?1 WHERE card_id = ?2",
            rusqlite::params![status, card],
        );
    };
    status(2, "expired_candidate");
    status(3, "held_for_review");
    status(4, "released");
    status(5, "reclaimed");
    w.sql(
        "UPDATE reservation SET lease_until = 1 WHERE card_id IN (2, 6)",
        [],
    );
    w.sql(
        "UPDATE reservation SET worktree = '/work/review', branch = 'spec/review' \
          WHERE card_id = 3",
        [],
    );

    let r = w.read(&["--reservations"]);
    assert_eq!(
        column(&r, "reservations", "card"),
        [json!("C-1"), json!("C-2"), json!("C-3"), json!("C-6")]
    );
    assert_eq!(
        column(&r, "reservations", "status"),
        [
            json!("held"),
            json!("expired_candidate"),
            json!("held_for_review"),
            json!("held")
        ]
    );
    assert_eq!(
        column(&r, "reservations", "lease_ended"),
        [json!(false), json!(true), json!(false), json!(true)]
    );
    for (at, (card, ended)) in [(1, false), (2, true), (3, false), (6, true)]
        .into_iter()
        .enumerate()
    {
        let shown = w.show(&format!("C-{card}"));
        assert_eq!(r["reservations"][at], reservation_of(&shown, ended));
    }
    assert_eq!(r["reservations"][2]["worktree"], json!("/work/review"));
    assert_eq!(r["reservations"][2]["branch"], json!("spec/review"));
}

#[test]
fn reservations_names_the_present_generation_once() {
    let w = world();
    w.add("again", &["it works"]);
    w.take(1, S_A);
    result(&w.card(Some(S_A), &["release", "C-1", "--generation", "1"]));
    result(&w.card(Some(S_B), &["take", "C-1"]));
    let r = w.read(&["--reservations"]);
    assert_eq!(column(&r, "reservations", "generation"), [json!(2)]);
    assert_eq!(column(&r, "reservations", "session"), [json!(S_B)]);
}

#[test]
fn reservations_is_an_empty_list_when_nobody_holds_a_card() {
    let w = world();
    assert_eq!(w.read(&["--reservations"])["reservations"], json!([]));
    w.add("one", &["it works"]);
    w.take(1, S_A);
    result(&w.card(Some(S_A), &["release", "C-1", "--generation", "1"]));
    assert_eq!(w.read(&["--reservations"])["reservations"], json!([]));
}

// ---- AC-4: notes -------------------------------------------------------------------------------

#[test]
fn note_by_number_gives_the_whole_body_byte_for_byte() {
    let w = world();
    w.add("one", &[]);
    w.add("two", &[]);
    let long = format!(
        "첫 줄\t탭\n둘째 줄 \"따옴표\" \\ 역슬래시\n{}\n\n끝 줄\n",
        "긴".repeat(5000)
    );
    let first = w.note(1, "agent-a", Some(S_A), &long);
    let second = w.note(2, "agent-b", None, "no session");
    // An imported note: no `card.note` event, and line endings kept as they came.
    w.sql(
        "INSERT INTO card_note (card_id, kind, author, origin, body, ref, created_at) \
         VALUES (2, 'comment', 'importer', 'import', ?1, 'import:1', 1000)",
        ["imported\r\nsecond\r\n"],
    );
    let third = second + 1;

    let ids = format!("{first},{second},{third}");
    let r = w.read(&["--note", &ids]);
    let shown = [w.show("C-1"), w.show("C-2"), w.show("C-2")];
    for (at, (note_id, card)) in [(first, 1), (second, 2), (third, 2)]
        .into_iter()
        .enumerate()
    {
        let mut want = shown_note(&shown[at], note_id);
        let bytes = want["body"].as_str().unwrap().len();
        want["card"] = json!(format!("C-{card}"));
        want["card_id"] = json!(card);
        want["bytes"] = json!(bytes);
        assert_eq!(r["notes"][at], want);
    }
    assert_eq!(r["notes"][0]["body"].as_str().unwrap(), long);
    assert_eq!(r["notes"][0]["session"], json!(S_A));
    assert_eq!(r["notes"][1]["session"], Value::Null);
    assert_eq!(r["notes"][2]["origin"], json!("import"));
    assert_eq!(r["notes"][2]["ref"], json!("import:1"));
    assert_eq!(
        r["notes"][2]["body"],
        json!("imported\r\nsecond\r\n"),
        "the stored bytes"
    );
}

#[test]
fn note_heads_are_the_newest_first_without_the_body() {
    let w = world();
    w.add("one", &[]);
    w.add("silent", &[]);
    w.add("other", &[]);
    w.note(1, "agent-a", Some(S_A), "oldest\nbody");
    let middle = w.note(1, "agent-b", None, "middle line\nsecond line");
    w.note(3, "agent-c", None, "another card");
    let newest = w.note(1, "agent-a", Some(S_A), "newest 첫 줄\n둘째");

    let r = w.read(&["--card", "C-1,C-2", "--notes", "2"]);
    let shown = w.show("C-1");
    let head = |note_id: i64, line: &str| {
        let note = shown_note(&shown, note_id);
        json!({
            "note_id": note_id, "kind": note["kind"], "author": note["author"],
            "origin": note["origin"], "session": note["session"],
            "created_at": note["created_at"], "bytes": note["body"].as_str().unwrap().len(),
            "line": line,
        })
    };
    assert_eq!(
        r["cards"][0]["notes"],
        json!([head(newest, "newest 첫 줄"), head(middle, "middle line")])
    );
    assert_eq!(r["cards"][1]["notes"], json!([]), "a card with no note");
    assert!(
        r.get("notes").is_none(),
        "no note was asked for by number: {r}"
    );

    // Without `--notes`, and with 0, a card carries no note list.
    for args in [vec!["--card", "C-1"], vec!["--card", "C-1", "--notes", "0"]] {
        let card = w.read(&args)["cards"][0].clone();
        assert!(card.get("notes").is_none(), "{args:?}: {card}");
    }
}

#[test]
fn note_heads_stop_at_the_number_asked_for() {
    let w = world();
    w.add("many", &[]);
    let ids: Vec<i64> = (1..=21)
        .map(|i| w.note(1, "agent-a", None, &format!("note {i}")))
        .collect();
    for (asked, given) in [("1", 1_usize), ("20", 20)] {
        let r = w.read(&["--card", "C-1", "--notes", asked]);
        let want: Vec<Value> = ids.iter().rev().take(given).map(|i| json!(i)).collect();
        assert_eq!(column(&r["cards"][0], "notes", "note_id"), want, "{asked}");
    }
}

#[test]
fn note_head_line_is_cut_at_two_hundred_characters() {
    let w = world();
    w.add("long", &[]);
    let (fits, over) = ("가".repeat(200), "나".repeat(201));
    w.note(1, "agent-a", None, &format!("{fits}\nrest"));
    w.note(1, "agent-a", None, &format!("{over}\nrest"));
    let r = w.read(&["--card", "C-1", "--notes", "2"]);
    assert_eq!(
        column(&r["cards"][0], "notes", "line"),
        [json!("나".repeat(200)), json!(fits)]
    );
}

// ---- AC-5: refusals ----------------------------------------------------------------------------

/// A ledger with one card and one note, and the proof that a good call is answered, so a refusal
/// below is a refusal of that input and not of the command.
fn refusable() -> World {
    let w = world();
    w.add("one", &["it works"]);
    w.note(1, "agent-a", None, "hello");
    let r = w.read(&["--card", "C-1", "--note", "1", "--notes", "1"]);
    assert_eq!(column(&r, "cards", "card"), [json!("C-1")]);
    w
}

#[test]
fn refusal_a_call_the_parser_cannot_take_is_a_usage_error() {
    let w = refusable();
    for args in [
        vec![],
        vec!["--notes", "1"],
        vec!["--reservations", "--notes", "1"],
        vec!["--note", "1", "--notes", "1"],
        vec!["--card", "C-1", "--notes", "two"],
        vec!["--card", "C-1", "--notes", "99999999999999999999"],
        vec!["--note", "one"],
        vec!["--card"],
    ] {
        usage(&w.query(&args));
    }
}

#[test]
fn refusal_an_unknown_card_or_note_refuses_the_whole_call() {
    let w = refusable();
    for (args, named) in [
        (vec!["--card", "C-99"], "there is no C-99"),
        (vec!["--card", "C-1,C-99"], "there is no C-99"),
        (vec!["--reservations", "--card", "C-99"], "there is no C-99"),
        (vec!["--note", "999"], "there is no note 999"),
        (vec!["--note", "1,999"], "there is no note 999"),
        (
            vec!["--card", "C-1", "--note", "999"],
            "there is no note 999",
        ),
        (vec!["--note", "0"], "there is no note 0"),
        (vec!["--note=-1"], "there is no note -1"),
    ] {
        let (kind, message) = refusal(&w.query(&args));
        assert_eq!(kind, "not_found", "{args:?}: {message}");
        assert!(message.contains(named), "{args:?}: {message}");
    }
}

#[test]
fn refusal_a_card_is_named_by_its_c_number() {
    let w = refusable();
    for name in ["#1", "C-1,#12", "banana", "C-0", "1", ""] {
        let (kind, message) = refusal(&w.query(&["--card", name]));
        assert_eq!(kind, "invalid_input", "{name:?}: {message}");
        assert!(
            message.starts_with("card query refused:"),
            "{name:?}: {message}"
        );
    }
}

#[test]
fn refusal_more_than_fifty_cards() {
    let w = world();
    w.bulk_cards(51);
    let names: Vec<String> = (1..=51).map(|i| format!("C-{i}")).collect();
    let fifty = names[..50].join(",");
    let r = w.read(&["--card", &fifty]);
    assert_eq!(r["cards"].as_array().unwrap().len(), 50);
    // A name given twice counts once.
    let repeated = format!("{fifty},C-1");
    let r = w.read(&["--card", &repeated]);
    assert_eq!(r["cards"].as_array().unwrap().len(), 50);
    let (kind, message) = refusal(&w.query(&["--card", &names.join(",")]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.starts_with("card query refused:"), "{message}");
    assert!(message.contains("50"), "{message}");
}

#[test]
fn refusal_more_than_ten_note_numbers() {
    let w = world();
    w.add("one", &[]);
    let ids: Vec<String> = (1..=11)
        .map(|i| w.note(1, "agent-a", None, &format!("note {i}")).to_string())
        .collect();
    let ten = ids[..10].join(",");
    let r = w.read(&["--note", &ten]);
    assert_eq!(r["notes"].as_array().unwrap().len(), 10);
    let repeated = format!("{ten},{}", ids[0]);
    let r = w.read(&["--note", &repeated]);
    assert_eq!(r["notes"].as_array().unwrap().len(), 10);
    let (kind, message) = refusal(&w.query(&["--note", &ids.join(",")]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.starts_with("card query refused:"), "{message}");
    assert!(message.contains("10"), "{message}");
}

#[test]
fn refusal_notes_outside_zero_to_twenty() {
    let w = refusable();
    for good in ["0", "20"] {
        w.read(&["--card", "C-1", "--notes", good]);
    }
    for bad in [
        vec!["--card", "C-1", "--notes", "21"],
        vec!["--card", "C-1", "--notes", "-1"],
        vec!["--card", "C-1", "--notes=-1"],
    ] {
        let (kind, message) = refusal(&w.query(&bad));
        assert_eq!(kind, "invalid_input", "{bad:?}: {message}");
        assert!(message.starts_with("card query refused:"), "{message}");
        assert!(message.contains("--notes"), "{message}");
    }
}

// ---- AC-6: nothing is written ------------------------------------------------------------------

#[test]
fn readonly_a_query_leaves_the_home_as_it_was() {
    let w = world();
    w.add("one", &["it works"]);
    w.take(1, S_A);
    w.hand_back(1, &sha('a'));
    w.verdict(1, AUDITOR, &sha('a'), true, None);
    let before = w.home.snapshot();
    w.read(&[
        "--card",
        "C-1",
        "--notes",
        "5",
        "--reservations",
        "--note",
        "1",
    ]);
    refusal(&w.query(&["--card", "C-9"]));
    refusal(&w.query(&["--note", "9"]));
    assert_eq!(
        w.home.snapshot(),
        before,
        "no file, folder, byte, size or time changed"
    );
}

#[test]
fn readonly_an_absent_ledger_is_refused_like_card_show_and_nothing_is_created() {
    let bare = TempHome::new();
    let project = bare.project("demo").to_str().unwrap().to_string();
    let snapshot = bare.snapshot();
    let run = |args: &[&str]| {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &project]);
        bare.command(&all).output().expect("run moai")
    };
    let like_show = refusal(&run(&["show", "C-1"]));
    assert!(
        like_show.1.contains("ledger-register-plan"),
        "{like_show:?}"
    );
    for args in [
        vec!["query", "--card", "C-1"],
        vec!["query", "--reservations"],
        vec!["query", "--note", "1"],
    ] {
        assert_eq!(refusal(&run(&args)), like_show, "{args:?}");
    }
    assert_eq!(bare.snapshot(), snapshot);
}

#[test]
fn readonly_the_help_says_it_reads_the_ledger_only() {
    let w = world();
    let help = w
        .home
        .command(&["card", "query", "--help"])
        .output()
        .unwrap();
    assert!(help.status.success(), "{}", text_of(&help));
    let text = String::from_utf8_lossy(&help.stdout).into_owned();
    let first = text.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
    assert!(first.starts_with("Reads the ledger only."), "{first}");
    for option in ["--card", "--reservations", "--note", "--notes"] {
        assert!(text.contains(option), "{option}: {text}");
    }
}
