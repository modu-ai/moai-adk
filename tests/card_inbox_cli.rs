//! `moai card inbox` (card C-116, SPEC-20261119-leader-inbox): the lane notes that still wait for
//! the leader, read from the ledger without writing anything.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! notes are inserted with SQL so their author, origin and time are exactly what a case needs:
//! `card note` signs a note with the session and stamps it with the clock.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::process::Output;

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
    /// `moai card inbox <project> [args]`.
    fn inbox(&self, extra: &[&str]) -> Output {
        let mut args = vec!["card", "inbox", self.project.as_str()];
        args.extend_from_slice(extra);
        self.home.command(&args).output().expect("run moai")
    }

    /// `moai card add --title T` for each title; the cards are `C-1`, `C-2`, … in order.
    fn cards(&self, titles: &[&str]) {
        for title in titles {
            let out = self
                .home
                .command(&["card", "add", "--title", title, "--project", &self.project])
                .output()
                .unwrap();
            assert_eq!(out.status.code(), Some(0), "{}", text(&out));
        }
    }

    /// A card that is `done`, inserted whole (an insert has no state check, and `done` needs a
    /// merged pull request link to be reached by the commands).
    fn done_card(&self, id: i64, title: &str) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "INSERT INTO card (card_id, title, state, created_by, created_at, updated_at, closed_at) \
                 VALUES (?1, ?2, 'done', 'tester', 1, 2, 2)",
                rusqlite::params![id, title],
            )
            .unwrap();
    }

    /// One note, in the order of the calls (the note id grows with each).
    fn note(&self, card: i64, kind: &str, author: &str, origin: &str, body: &str, at: i64) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
                rusqlite::params![card, kind, author, origin, body, at],
            )
            .unwrap();
    }

    fn lane(&self, card: i64, body: &str, at: i64) {
        self.note(card, "comment", "agent-1", "local", body, at);
    }

    fn leader(&self, card: i64, body: &str, at: i64) {
        self.note(card, "decision", "leader", "local", body, at);
    }
}

fn text(out: &Output) -> String {
    format!(
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn result(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(0), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("card"), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(out.status.code(), Some(1), "{}", text(out));
    let v: Value =
        serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {}", text(out)));
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

/// `(card, family, note_id)` of every item, in the order printed.
fn rows(r: &Value) -> Vec<(String, String, i64)> {
    r["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| {
            (
                i["card"].as_str().unwrap().to_string(),
                i["family"].as_str().unwrap().to_string(),
                i["note_id"].as_i64().unwrap(),
            )
        })
        .collect()
}

fn row(card: &str, family: &str, id: i64) -> (String, String, i64) {
    (card.to_string(), family.to_string(), id)
}

#[test]
fn the_inbox_lists_each_request_family_in_note_order() {
    let w = world();
    w.cards(&["one", "two", "three"]);
    w.done_card(4, "closed");
    // C-1: one note of each family, with lines that look like a request and are not.
    w.lane(
        1,
        "감사 요청 tier 2 head=0123456789abcdef0123456789abcdef01234567\n\nbody",
        1000,
    ); // 1
    w.lane(1, "LANE-DONE agent-1 C-1 PR #7 head abc — done", 1001); // 2: no `audit requested`
    w.lane(
        1,
        "LANE-NOTE agent-1 C-1 question? default if you stay silent",
        1002,
    ); // 3
    w.lane(1, "see LANE-BLOCKED in the middle of a line", 1003); // 4
    w.lane(
        1,
        "LANE-BLOCKED agent-1 C-1 slice 2 · kind=decision owner=leader",
        1004,
    ); // 5
    w.lane(
        1,
        "LANE-READY-INTEGRATION C-1 head=abc spec=SPEC-20261119-x",
        1005,
    ); // 6
    w.note(
        1,
        "decision",
        "agent-1",
        "local",
        "DECISION-QUESTION agent-1 C-1 (reversible): a or b",
        1006,
    ); // 7
       // C-2: the hand-off form of an audit request.
    w.lane(
        2,
        "LANE-DONE agent-2 C-2 PR #8 head abc — audit requested",
        1010,
    ); // 8
       // C-3: nothing asked. C-4 is done: its request is not the leader's work any more.
    w.lane(3, "LANE-DONE agent-3 C-3 PR #9 head abc — done", 1020); // 9
    w.lane(4, "LANE-BLOCKED agent-4 C-4 old", 1030); // 10

    let r = result(&w.inbox(&[]));
    assert_eq!(r["action"], json!("inbox"));
    assert_eq!(r["ledger"], json!("ready"));
    assert_eq!(r["leader"], json!("leader"));
    assert_eq!(r["writes"], json!(false));
    assert_eq!(
        rows(&r),
        vec![
            row("C-1", "audit_request", 1),
            row("C-1", "blocked", 5),
            row("C-1", "ready_for_integration", 6),
            row("C-1", "decision_question", 7),
            row("C-2", "audit_request", 8),
        ]
    );
    assert_eq!(r["count"], json!(5));
    let now = r["now"].as_i64().unwrap();
    assert!(now > 1_700_000_000, "{now}");
    for item in r["items"].as_array().unwrap() {
        let at = item["created_at"].as_i64().unwrap();
        assert_eq!(item["age_secs"].as_i64().unwrap(), now - at, "{item}");
        // Every request of this case was written by agent-1 (C-2's hand-off line by agent-1 too).
        assert_eq!(item["author"], json!("agent-1"), "{item}");
    }
    // The first line is what the leader reads; the rest of the note stays in `card show`.
    let first = &r["items"][0];
    assert_eq!(first["title"], json!("one"));
    assert_eq!(first["state"], json!("draft"));
    assert_eq!(
        first["line"],
        json!("감사 요청 tier 2 head=0123456789abcdef0123456789abcdef01234567")
    );
}

#[test]
fn a_later_leader_note_answers_and_a_newer_note_supersedes() {
    let w = world();
    w.cards(&[
        "answered",
        "asked after",
        "superseded",
        "newest unanswered",
        "two families",
    ]);
    // C-1: the leader answers after the request: gone.
    w.lane(1, "LANE-BLOCKED agent-1 C-1 first", 100); // 1
    w.leader(1, "LEADER answer", 101); // 2
                                       // C-2: the leader wrote before the request (the id just below): still waiting.
    w.leader(2, "LEADER earlier note", 110); // 3
    w.lane(2, "LANE-BLOCKED agent-1 C-2 later", 111); // 4
                                                      // C-3: the older request was answered, a newer one of the same family came after: the new one.
    w.lane(3, "LANE-BLOCKED agent-1 C-3 old", 120); // 5
    w.leader(3, "LEADER answer to the old one", 121); // 6
    w.lane(3, "LANE-BLOCKED agent-1 C-3 new", 122); // 7
                                                    // C-4: two unanswered requests of one family: only the newest is listed.
    w.lane(4, "LANE-BLOCKED agent-1 C-4 old", 130); // 8
    w.lane(4, "LANE-BLOCKED agent-1 C-4 new", 131); // 9
                                                    // C-5: the leader answers after both; a later request of the other family waits alone.
    w.lane(5, "LANE-BLOCKED agent-1 C-5", 140); // 10
    w.lane(5, "감사 요청 tier 2 head=abc", 141); // 11
    w.leader(5, "LEADER answer to both", 142); // 12
    w.lane(5, "LANE-READY-INTEGRATION C-5 head=abc", 143); // 13

    let r = result(&w.inbox(&[]));
    assert_eq!(
        rows(&r),
        vec![
            row("C-2", "blocked", 4),
            row("C-3", "blocked", 7),
            row("C-4", "blocked", 9),
            row("C-5", "ready_for_integration", 13),
        ]
    );
}

#[test]
fn the_leader_name_and_imported_notes_decide_what_counts() {
    let w = world();
    w.cards(&["names", "imported", "leader's own line"]);
    // C-1: `boss` answers; the author `leader` is then not the leader.
    w.lane(1, "LANE-BLOCKED agent-1 C-1", 100); // 1
    w.note(1, "decision", "boss", "local", "BOSS answer", 101); // 2
    w.lane(2, "LANE-BLOCKED agent-1 C-2 a real request", 100); // 3
                                                               // A newer imported GitHub comment must not replace the real request as the newest one.
    w.note(
        2,
        "comment",
        "GoosLab",
        "import",
        "LANE-BLOCKED imported comment",
        101,
    ); // 4
       // C-3: the leader's own `LANE-BLOCKED` line is not a request to the leader.
    w.note(
        3,
        "comment",
        "leader",
        "local",
        "LANE-BLOCKED leader 'note to self'",
        100,
    ); // 5

    let default = result(&w.inbox(&[]));
    assert_eq!(
        rows(&default),
        vec![row("C-1", "blocked", 1), row("C-2", "blocked", 3)],
        "with the author `leader`, `boss` answers nothing and the import is no request"
    );
    let boss = result(&w.inbox(&["--leader", "boss"]));
    assert_eq!(boss["leader"], json!("boss"));
    assert_eq!(
        rows(&boss),
        vec![row("C-2", "blocked", 3), row("C-3", "blocked", 5)],
        "with `boss`, C-1 is answered and a line of `leader` is a lane's line again"
    );

    // The edges of `--leader`: 80 characters pass, 81, empty and a control character do not.
    let longest = "a".repeat(80);
    assert_eq!(
        result(&w.inbox(&["--leader", &longest]))["leader"],
        json!(longest)
    );
    let before = w.home.snapshot();
    for bad in ["a".repeat(81), String::new(), "bo\u{7}ss".to_string()] {
        let (kind, message) = refusal(&w.inbox(&["--leader", &bad]));
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(message.contains("card inbox refused"), "{message}");
    }
    assert_eq!(w.home.snapshot(), before, "a refusal writes nothing");
}

#[test]
fn the_reader_commands_accept_the_project_flag() {
    let w = world();
    w.cards(&["one", "two"]);
    // `card next` skips a card without a scope, so the card it names carries one and is ready.
    let added = result(
        &w.home
            .command(&[
                "card",
                "add",
                "--title",
                "start here",
                "--scope",
                "apps/demo/",
                "--project",
                &w.project,
            ])
            .output()
            .unwrap(),
    );
    let name = added["name"].as_str().unwrap().to_string();
    result(
        &w.home
            .command(&["card", "ready", &name, "--project", &w.project])
            .output()
            .unwrap(),
    );
    w.lane(1, "LANE-BLOCKED agent-1 C-1", 100);
    // `card list` with `--project` only: the answer carries the cards array.
    let list = result(
        &w.home
            .command(&["card", "list", "--project", &w.project])
            .output()
            .unwrap(),
    );
    assert!(list["cards"].is_array(), "{list}");
    // `card next` with `--project` only: the scoped ready card.
    let next = result(
        &w.home
            .command(&["card", "next", "--project", &w.project])
            .output()
            .unwrap(),
    );
    assert_eq!(next["next"], json!(name), "{next}");
    // `card withdraw` with `--project` only.
    let mut wd = w.home.command(&[
        "card",
        "withdraw",
        "C-2",
        "--reason",
        "not needed",
        "--by",
        "tester",
        "--project",
        &w.project,
    ]);
    wd.env(
        "CLAUDE_CODE_SESSION_ID",
        "aaaaaaaa-0000-4000-8000-00000000000a",
    );
    result(&wd.output().unwrap());
    // `card inbox` with `--project` only: the withdrawn card holds no request.
    let inbox = result(
        &w.home
            .command(&["card", "inbox", "--project", &w.project])
            .output()
            .unwrap(),
    );
    assert_eq!(inbox["count"], json!(1), "{inbox}");
}

#[test]
fn the_inbox_reads_only_and_an_absent_ledger_is_empty() {
    let w = world();
    w.cards(&["one"]);
    w.lane(1, "LANE-BLOCKED agent-1 C-1", 100);
    let before = w.home.snapshot();
    let r = result(&w.inbox(&[]));
    assert_eq!(r["count"], json!(1));
    assert_eq!(
        w.home.snapshot(),
        before,
        "no file, folder, size or time changed"
    );

    // A project with no ledger: nothing to list, nothing created.
    let bare = TempHome::new();
    let project = bare.project("demo").to_str().unwrap().to_string();
    let snapshot = bare.snapshot();
    let r = result(&bare.command(&["card", "inbox", &project]).output().unwrap());
    assert_eq!(r["ledger"], json!("absent"));
    assert_eq!(r["count"], json!(0));
    assert_eq!(r["items"], json!([]));
    assert_eq!(bare.snapshot(), snapshot);

    // A file that is not the ledger: the ledger's own refusal, and the file is left alone.
    let path = ledger::ledger_path(w.home.path(), "demo").unwrap();
    std::fs::write(&path, b"this is not a database").unwrap();
    let before = w.home.snapshot();
    let (kind, message) = refusal(&w.inbox(&[]));
    assert_eq!(kind, "blocked", "{message}");
    assert!(message.contains("ledger_unavailable"), "{message}");
    assert_eq!(w.home.snapshot(), before);

    // The first sentence of `--help`.
    let help = w
        .home
        .command(&["card", "inbox", "--help"])
        .output()
        .unwrap();
    assert!(help.status.success());
    let head = String::from_utf8_lossy(&help.stdout).into_owned();
    let first = head.lines().find(|l| !l.trim().is_empty()).unwrap_or("");
    assert!(first.starts_with("Reads the ledger only."), "{first}");
}
