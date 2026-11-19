//! `moai card show` and the session each note shows (SPEC-20261120-card-show-notes-once).
//!
//! A note's `session` is the session that wrote it, read from its `card.note` event. The answer
//! must not change when the notes are matched to their events once instead of once per note:
//! these tests pin what `session` is for a card with notes from several sessions, a note without
//! a session, an imported note that has no event, a second event that names the same note, and an
//! event of another card, and they pin that the read writes nothing.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use rusqlite::{params, Connection};
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Output, Stdio};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

struct World {
    home: TempHome,
    project: String,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project: project.to_str().unwrap().to_string(),
    }
}

impl World {
    fn card(&self, session: Option<&str>, args: &[&str], stdin: &str) -> Output {
        let mut all = vec!["card"];
        all.extend_from_slice(args);
        all.extend_from_slice(&["--project", &self.project]);
        let mut cmd = self.home.command(&all);
        if let Some(id) = session {
            cmd.env(SESSION_VAR, id);
        }
        let mut child = cmd
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .expect("run moai");
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stdin.as_bytes())
            .unwrap();
        child.wait_with_output().unwrap()
    }

    fn add(&self, title: &str) -> String {
        result(&self.card(None, &["add", "--title", title, "--ac", "one"], ""))["name"]
            .as_str()
            .unwrap()
            .to_string()
    }

    fn note(&self, card: &str, session: Option<&str>, body: &str) -> i64 {
        let out = self.card(
            session,
            &["note", card, "--kind", "report", "--by", "agent-9"],
            body,
        );
        result(&out)["note_id"].as_i64().unwrap()
    }

    fn show(&self, card: &str) -> Output {
        self.card(None, &["show", card], "")
    }

    fn notes(&self, card: &str) -> Vec<Value> {
        result(&self.show(card))["notes"]
            .as_array()
            .unwrap()
            .clone()
    }

    /// A read-write connection for rows the CLI cannot make (an imported note, a stray event).
    fn raw(&self) -> Connection {
        Connection::open(ledger::ledger_path(self.home.path(), "demo").unwrap()).unwrap()
    }

    fn card_id(&self, name: &str) -> i64 {
        name.trim_start_matches("C-").parse().unwrap()
    }

    fn event_row(&self, card_id: i64, note_id: i64, session: Value, key: &str) {
        self.raw()
            .execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, 1, 'test', ?2, 'card.note', ?3)",
                params![
                    key,
                    card_id,
                    json!({"note_id": note_id, "session": session}).to_string()
                ],
            )
            .unwrap();
    }

    /// Every row of the ledger tables that `card show` could change, as text.
    fn dump(&self) -> Vec<String> {
        let l = ledger::open_read_only(self.home.path(), "demo").unwrap();
        let mut rows = Vec::new();
        for table in [
            "card",
            "card_ac",
            "card_note",
            "card_link",
            "reservation",
            "event",
        ] {
            let mut st = l
                .conn()
                .prepare(&format!("SELECT * FROM {table} ORDER BY 1"))
                .unwrap();
            let n = st.column_count();
            let got: Vec<String> = st
                .query_map([], |r| {
                    Ok((0..n)
                        .map(|i| format!("{:?}", r.get_ref(i).unwrap()))
                        .collect::<Vec<_>>()
                        .join("|"))
                })
                .unwrap()
                .map(Result::unwrap)
                .collect();
            rows.push(format!("{table}: {}", got.join(";")));
        }
        rows
    }
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn sessions(notes: &[Value]) -> Vec<Value> {
    notes.iter().map(|n| n["session"].clone()).collect()
}

/// REQ-1: each note shows the session that wrote it; a note written without one shows `null`.
#[test]
fn each_note_shows_the_session_that_wrote_it() {
    let w = world();
    let name = w.add("notes from three writers");
    let one = w.note(&name, Some(S_A), "first, from session A");
    let two = w.note(&name, Some(S_B), "second, from session B");
    let three = w.note(&name, None, "third, from no session");
    let four = w.note(&name, Some(S_A), "fourth, from session A again");
    let notes = w.notes(&name);
    assert_eq!(
        notes.iter().map(|n| n["note_id"].clone()).collect::<Vec<_>>(),
        vec![json!(one), json!(two), json!(three), json!(four)],
        "notes come in note order: {notes:?}"
    );
    assert_eq!(
        sessions(&notes),
        vec![json!(S_A), json!(S_B), Value::Null, json!(S_A)],
        "{notes:?}"
    );
    for n in &notes {
        assert_eq!(n["author"], json!("agent-9"), "author is what the note is signed with: {n}");
    }
}

/// REQ-1: a note without a `card.note` event (an imported one) shows `session: null`, and its
/// neighbours keep theirs.
#[test]
fn a_note_without_an_event_shows_a_null_session() {
    let w = world();
    let name = w.add("one note without its event");
    let id = w.card_id(&name);
    w.note(&name, Some(S_A), "written through the command");
    w.raw()
        .execute(
            "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
             VALUES (?1, 'report', 'import', 'import', 'imported, no event', 5)",
            params![id],
        )
        .unwrap();
    w.note(&name, Some(S_B), "written after the import");
    let notes = w.notes(&name);
    assert_eq!(notes.len(), 3, "{notes:?}");
    assert_eq!(
        sessions(&notes),
        vec![json!(S_A), Value::Null, json!(S_B)],
        "{notes:?}"
    );
}

/// REQ-1: when two `card.note` events of the card name one note, the first (lowest event id)
/// decides; and an event of another card that names the same note id is never read for it.
#[test]
fn the_first_event_of_this_card_decides_and_another_card_is_not_read() {
    let w = world();
    let name = w.add("the card under test");
    let other = w.add("another card");
    let id = w.card_id(&name);
    let other_id = w.card_id(&other);
    let note = w.note(&name, Some(S_A), "the note");
    w.event_row(id, note, json!(S_B), "stray-second-event");
    w.event_row(other_id, note, json!(S_B), "stray-other-card");
    let notes = w.notes(&name);
    assert_eq!(sessions(&notes), vec![json!(S_A)], "{notes:?}");
    assert_eq!(w.notes(&other), Vec::<Value>::new());

    // A note whose own event came second still takes the first event that names it.
    let late = w.raw();
    late.execute(
        "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
         VALUES (?1, 'report', 'import', 'import', 'imported first', 6)",
        params![id],
    )
    .unwrap();
    let imported = late.last_insert_rowid();
    drop(late);
    w.event_row(id, imported, json!(S_B), "late-event-one");
    w.event_row(id, imported, json!(S_A), "late-event-two");
    let notes = w.notes(&name);
    assert_eq!(
        sessions(&notes),
        vec![json!(S_A), json!(S_B)],
        "{notes:?}"
    );
}

/// A card with many notes shows every one, in order, each with its own session; the read
/// writes nothing (every row of the ledger is the same afterwards) and two reads give the same
/// bytes.
#[test]
fn a_large_card_is_read_whole_and_nothing_is_written() {
    let w = world();
    let name = w.add("a card with many notes");
    let id = w.card_id(&name);
    {
        let mut c = w.raw();
        let tx = c.transaction().unwrap();
        for n in 0..300i64 {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, 'report', 'agent-9', 'local', ?2, ?3)",
                params![id, format!("note {n}"), 10 + n],
            )
            .unwrap();
            let note_id = tx.last_insert_rowid();
            let session = match n % 3 {
                0 => json!(S_A),
                1 => json!(S_B),
                _ => Value::Null,
            };
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, 'agent-9', ?3, 'card.note', ?4)",
                params![
                    format!("bulk-{n}"),
                    10 + n,
                    id,
                    json!({"note_id": note_id, "kind": "report", "ref": null,
                           "bytes": 6, "author": "agent-9", "session": session})
                    .to_string()
                ],
            )
            .unwrap();
        }
        tx.commit().unwrap();
    }
    let before = w.dump();
    let first = w.show(&name);
    let second = w.show(&name);
    assert_eq!(first.stdout, second.stdout, "two reads, two answers");
    let notes = result(&first)["notes"].as_array().unwrap().clone();
    assert_eq!(notes.len(), 300);
    for (n, note) in notes.iter().enumerate() {
        let want = match n % 3 {
            0 => json!(S_A),
            1 => json!(S_B),
            _ => Value::Null,
        };
        assert_eq!(note["session"], want, "note {n}: {note}");
        assert_eq!(note["body"], json!(format!("note {n}")), "note {n}");
    }
    assert_eq!(w.dump(), before, "card show wrote a row");
}
