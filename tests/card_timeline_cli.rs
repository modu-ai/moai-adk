//! `moai card timeline` (SPEC-20261120-card-timeline): one card's events, read without writing
//! and without making a writer wait.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! holder is the session id in the environment (`CLAUDE_CODE_SESSION_ID`). Events are made by the
//! real commands where a case needs the real payload, and inserted with SQL where it needs a
//! number or a time the commands do not give.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::card::timeline;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::process::{Output, Stdio};
use std::time::{Duration, Instant};

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

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

    /// `card note` with `text` on stdin.
    fn note(&self, session: &str, card: &str, by: &str, text: &str) {
        let mut cmd = self.home.command(&[
            "card",
            "note",
            card,
            "--kind",
            "comment",
            "--by",
            by,
            "--project",
            &self.project,
        ]);
        cmd.env(SESSION_VAR, session)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let mut child = cmd.spawn().expect("run moai");
        common::feed_stdin(&mut child, text.as_bytes());
        let out = child.wait_with_output().unwrap();
        assert_eq!(out.status.code(), Some(0), "{}", text_of(&out));
    }

    /// `card timeline C-1 [extra]`, as `session`.
    fn timeline(&self, session: Option<&str>, extra: &[&str]) -> Output {
        let mut args = vec!["timeline", "C-1"];
        args.extend_from_slice(extra);
        self.card(session, &args)
    }

    /// A card `C-1`: added, readied, taken by A (generation 1), a note by each of A and B, a
    /// renewal, and a handoff to B (generation 2). Returns nothing; the events are in the ledger.
    fn story(&self) {
        result(&self.card(
            None,
            &[
                "add", "--title", "story", "--scope", "src/x.rs", "--ac", "it works",
            ],
        ));
        result(&self.card(None, &["ready", "C-1"]));
        result(&self.card(Some(S_A), &["take", "C-1"]));
        self.note(S_A, "C-1", "agent-a", "hello from a\nsecond line");
        result(&self.card(Some(S_A), &["renew", "C-1", "--generation", "1"]));
        self.note(S_B, "C-1", "agent-b", "from b");
        result(&self.card(
            Some(S_A),
            &["handoff", "C-1", "--generation", "1", "--to", S_B],
        ));
    }

    /// One event row inserted with SQL (a time and a kind the commands do not give).
    fn event(&self, card: i64, kind: &str, at: i64, payload: &Value) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        let key = w
            .conn()
            .query_row(
                "SELECT ?1 || ':' || lower(hex(randomblob(8)))",
                [format!("test.{kind}")],
                |r| r.get::<_, String>(0),
            )
            .unwrap();
        w.conn()
            .execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, 'tester', ?3, ?4, ?5)",
                rusqlite::params![key, at, card, kind, payload.to_string()],
            )
            .unwrap();
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

fn refusal(out: &Output) -> (String, String, bool) {
    assert_eq!(out.status.code(), Some(1), "{}", text_of(out));
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
        v["error"]["retryable"].as_bool().unwrap_or(false),
    )
}

fn kinds(r: &Value) -> Vec<String> {
    r["events"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["kind"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn the_timeline_lists_the_cards_events_oldest_first_with_the_contract_fields() {
    let w = world();
    w.story();
    let r = result(&w.timeline(Some(S_B), &[]));
    assert_eq!(r["action"], json!("timeline"));
    assert_eq!(r["writes"], json!(false));
    assert_eq!(r["card"], json!("C-1"));

    // The header: the project, this session and the time of the read.
    assert_eq!(r["header"]["project_id"], json!("demo"));
    assert_eq!(r["header"]["session"], json!(S_B));
    assert!(r["header"]["read_at"].as_i64().unwrap() > 1_700_000_000);
    let no_session = result(&w.timeline(None, &[]));
    assert_eq!(no_session["header"]["session"], Value::Null);

    // Oldest first; the renewal is bookkeeping and is not shown by default.
    assert_eq!(
        kinds(&r),
        [
            "card.add",
            "card.ready",
            "reserve",
            "card.note",
            "card.note",
            "handoff"
        ]
    );
    let events = r["events"].as_array().unwrap();
    for pair in events.windows(2) {
        assert!(pair[0]["at"].as_i64() <= pair[1]["at"].as_i64(), "{pair:?}");
    }
    for e in events {
        for key in ["at", "kind", "session", "lane", "detail"] {
            assert!(e.get(key).is_some(), "{key} missing in {e}");
        }
    }

    // A note line carries the author and its first line; the lane is the author of the newest
    // note the event's session wrote on this card.
    let first_note = &events[3];
    assert_eq!(first_note["detail"]["note_id"], json!(1));
    assert_eq!(first_note["note"]["author"], json!("agent-a"));
    assert_eq!(first_note["note"]["origin"], json!("local"));
    assert_eq!(first_note["note"]["line"], json!("hello from a"));
    assert_eq!(first_note["session"], json!(S_A));
    assert_eq!(first_note["lane"], json!("agent-a"));
    assert_eq!(events[2]["session"], json!(S_A), "the reserve event");
    assert_eq!(events[2]["lane"], json!("agent-a"));

    // The current owner after the handoff: B, generation 2, named by the lane B wrote as.
    assert_eq!(r["owner"]["session"], json!(S_B));
    assert_eq!(r["owner"]["lane"], json!("agent-b"));
    assert_eq!(r["owner"]["generation"], json!(2));
    assert_eq!(r["owner"]["status"], json!("held"));
    assert!(r["owner"]["lease_until"].as_i64().unwrap() > 0);
    assert_eq!(r["count"], json!(6));
    assert_eq!(r["more"], json!(false));
}

#[test]
fn the_timeline_reads_and_writes_nothing() {
    let w = world();
    w.story();
    let before = w.home.snapshot();
    result(&w.timeline(Some(S_A), &[]));
    result(&w.timeline(Some(S_A), &["--all", "--limit", "2"]));
    assert_eq!(
        w.home.snapshot(),
        before,
        "no file, folder, byte, size or time changed"
    );

    // A project with no ledger: refused like `card show`, and nothing is created.
    let bare = TempHome::new();
    let project = bare.project("demo").to_str().unwrap().to_string();
    let snapshot = bare.snapshot();
    let out = bare
        .command(&["card", "timeline", "C-1", "--project", &project])
        .output()
        .unwrap();
    let (kind, message, _) = refusal(&out);
    assert_eq!(kind, "not_found", "{message}");
    assert!(message.contains("ledger-register-plan"), "{message}");
    assert_eq!(bare.snapshot(), snapshot);
}

#[test]
fn one_timeline_call_never_rescans_event_per_row() {
    let w = world();
    result(&w.card(None, &["add", "--title", "big"]));
    {
        let l = ledger::open_write(w.home.path(), "demo").unwrap();
        let tx = l.conn().unchecked_transaction().unwrap();
        for i in 1..=900 {
            tx.execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (1, 'comment', 'agent-a', 'local', ?1, ?2)",
                rusqlite::params![format!("note {i}\nbody"), 1000 + i],
            )
            .unwrap();
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, 'tester', 1, 'card.note', ?3)",
                rusqlite::params![
                    format!("test.note:{i}"),
                    1000 + i,
                    json!({"note_id": i, "kind": "comment", "session": S_A, "bytes": 10})
                        .to_string()
                ],
            )
            .unwrap();
        }
        tx.commit().unwrap();
    }
    // The three statements the command runs, with their plans: no correlated sub-query, and
    // each reads `event` once.
    let l = ledger::open_read_only(w.home.path(), "demo").unwrap();
    for (name, sql, params) in [
        ("events", timeline::EVENTS_SQL, vec![1_i64, 0, 0, 50]),
        ("lanes", timeline::LANES_SQL, vec![1, 2000]),
        ("end", timeline::END_SQL, vec![1]),
    ] {
        let mut stmt = l
            .conn()
            .prepare(&format!("EXPLAIN QUERY PLAN {sql}"))
            .unwrap();
        let plan: Vec<String> = stmt
            .query_map(rusqlite::params_from_iter(params), |r| {
                r.get::<_, String>(3)
            })
            .unwrap()
            .collect::<Result<_, _>>()
            .unwrap();
        let joined = plan.join(" | ");
        assert!(!joined.contains("CORRELATED"), "{name}: {joined}");
        let on_event = plan
            .iter()
            .filter(|d| d.starts_with("SCAN e") || d.starts_with("SEARCH e"))
            .count();
        assert_eq!(on_event, 1, "{name} reads event once: {joined}");
    }
    // And the call itself: a page of the 900, bounded by the limit.
    let r = result(&w.timeline(Some(S_A), &["--limit", "500"]));
    assert_eq!(r["count"], json!(500));
    assert_eq!(r["more"], json!(true));
    // The first event is the card's own `card.add`; the notes follow, first line only.
    assert_eq!(r["events"][0]["kind"], json!("card.add"));
    assert_eq!(r["events"][1]["note"]["line"], json!("note 1"));
    assert_eq!(r["events"][2]["note"]["line"], json!("note 2"));
}

#[test]
fn the_cursor_continues_with_no_gap_and_no_overlap() {
    let w = world();
    result(&w.card(None, &["add", "--title", "cursor"]));
    for i in 1..=10 {
        w.event(1, "card.pr", 5000 + i, &json!({"number": i}));
    }
    let all = result(&w.timeline(None, &["--limit", "500"]));
    let want: Vec<Value> = all["events"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| json!([e["kind"], e["at"], e["detail"]]))
        .collect();
    assert_eq!(want.len(), 11, "the add event and ten rows");
    assert_eq!(all["more"], json!(false));

    // The edge: exactly as many events as the limit is the last page; one fewer is not.
    let exact = result(&w.timeline(None, &["--limit", "11"]));
    assert_eq!(
        (exact["count"].clone(), exact["more"].clone()),
        (json!(11), json!(false))
    );
    let short = result(&w.timeline(None, &["--limit", "10"]));
    assert_eq!(
        (short["count"].clone(), short["more"].clone()),
        (json!(10), json!(true))
    );

    // Three at a time, handing `next` back.
    let mut got: Vec<Value> = Vec::new();
    let mut cursor: Option<String> = None;
    for _ in 0..10 {
        let mut extra = vec!["--limit", "3"];
        if let Some(c) = cursor.as_deref() {
            extra.extend_from_slice(&["--after", c]);
        }
        let page = result(&w.timeline(None, &extra));
        for e in page["events"].as_array().unwrap() {
            got.push(json!([e["kind"], e["at"], e["detail"]]));
        }
        cursor = Some(page["next"].as_str().unwrap().to_string());
        if page["more"] == json!(false) {
            break;
        }
    }
    assert_eq!(got, want, "no gap, no overlap, same order");
    assert_eq!(
        all["end"],
        json!(cursor.clone().unwrap()),
        "end is the newest"
    );

    // Nothing new: nothing returned, and the cursor stays; a new event comes alone.
    let cursor = cursor.unwrap();
    let again = result(&w.timeline(None, &["--after", &cursor]));
    assert_eq!(again["count"], json!(0));
    assert_eq!(again["next"], json!(cursor));
    w.event(1, "card.pr", 6000, &json!({"number": 99}));
    let one = result(&w.timeline(None, &["--after", &cursor]));
    assert_eq!(one["count"], json!(1));
    assert_eq!(one["events"][0]["detail"]["number"], json!(99));

    // Another card's cursor and a malformed one are refused.
    result(&w.card(None, &["add", "--title", "other"]));
    let other = format!("tl1:2:{}", 1);
    for bad in [
        other.as_str(),
        "garbage",
        "tl1:1:-5",
        "tl1:1:",
        "tl2:1:3",
        "",
    ] {
        let (kind, message, retryable) = refusal(&w.timeline(None, &["--after", bad]));
        assert_eq!(kind, "invalid_input", "{bad}: {message}");
        assert!(message.contains("card timeline refused"), "{message}");
        assert!(!retryable);
    }
}

#[test]
fn bookkeeping_events_are_hidden_unless_all_and_a_fence_reject_always_shows() {
    let w = world();
    result(&w.card(
        None,
        &[
            "add", "--title", "fence", "--scope", "src/y.rs", "--ac", "ok",
        ],
    ));
    result(&w.card(None, &["ready", "C-1"]));
    result(&w.card(Some(S_A), &["take", "C-1"]));
    result(&w.card(Some(S_A), &["renew", "C-1", "--generation", "1"]));
    // A rejected call leaves a `fence_reject` event behind.
    let rejected = w.card(Some(S_A), &["renew", "C-1", "--generation", "9"]);
    assert_eq!(rejected.status.code(), Some(1), "{}", text_of(&rejected));

    let plain = kinds(&result(&w.timeline(None, &[])));
    assert!(!plain.contains(&"renew".to_string()), "{plain:?}");
    assert!(plain.contains(&"fence_reject".to_string()), "{plain:?}");
    let all = kinds(&result(&w.timeline(None, &["--all"])));
    assert!(all.contains(&"renew".to_string()), "{all:?}");
    assert!(all.contains(&"fence_reject".to_string()), "{all:?}");
    assert_eq!(all.len(), plain.len() + 1, "only the renewal is extra");
}

#[test]
fn a_busy_ledger_a_missing_card_and_bad_input_use_the_existing_envelope() {
    let w = world();
    w.story();

    // Another process holds the ledger: the call gives up quickly with the busy answer.
    let path = ledger::ledger_path(w.home.path(), "demo").unwrap();
    let holder = rusqlite::Connection::open(&path).unwrap();
    holder.execute_batch("BEGIN EXCLUSIVE").unwrap();
    let started = Instant::now();
    let out = w.timeline(Some(S_A), &[]);
    let waited = started.elapsed();
    let (kind, message, retryable) = refusal(&out);
    assert_eq!(kind, "io", "{message}");
    assert!(message.contains("ledger busy"), "{message}");
    assert!(retryable, "busy is retryable");
    assert!(
        waited < Duration::from_secs(3),
        "gave up after the short wait, not the five seconds of a write: {waited:?}"
    );
    holder.execute_batch("COMMIT").unwrap();
    drop(holder);
    result(&w.timeline(Some(S_A), &[]));

    // A card that is not there.
    let missing = w.card(None, &["timeline", "C-99"]);
    let (kind, message, _) = refusal(&missing);
    assert_eq!(kind, "not_found", "{message}");
    // Bad numbers and names.
    for extra in [
        vec!["--limit", "0"],
        vec!["--limit", "501"],
        vec!["--limit", "-3"],
    ] {
        let (kind, message, _) = refusal(&w.timeline(None, &extra));
        assert_eq!(kind, "invalid_input", "{extra:?}: {message}");
        assert!(message.contains("card timeline refused"), "{message}");
    }
    let (kind, message, _) = refusal(&w.card(None, &["timeline", "#12"]));
    assert_eq!(kind, "invalid_input", "{message}");
    // An unknown flag is the usual usage error, exit 2.
    let usage = w.timeline(None, &["--follow"]);
    assert_eq!(usage.status.code(), Some(2), "{}", text_of(&usage));
}

#[test]
fn the_timeline_help_and_the_operations_table_say_it_reads_the_ledger_only() {
    let w = world();
    let help = w
        .home
        .command(&["card", "timeline", "--help"])
        .output()
        .unwrap();
    assert!(help.status.success(), "{}", text_of(&help));
    let head = String::from_utf8_lossy(&help.stdout).into_owned();
    let first = head.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
    assert!(first.starts_with("Reads the ledger only."), "{first}");

    let table = std::fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/card.md"),
    )
    .unwrap();
    let row = table
        .lines()
        .find(|l| l.starts_with("| `card timeline"))
        .expect("a row for card timeline in the operations table");
    assert!(row.contains("읽기만"), "{row}");
}
