//! `moai leader brief` (SPEC-20261124-leader-brief): the short state a leader reads when it
//! wakes. It shrinks the answer of `card orchestrate plan` to the landing queue, the blocks and
//! the decisions, each with its note number, in at most 4096 bytes.
//!
//! Every command runs against a temporary HOME and a temporary git project; the user's real
//! state folder and registry are not read or written. Ledger rows are inserted with SQL so their
//! author, state and order are exactly what a case needs. No `claude` or `orca` process starts.
#![cfg(unix)]

mod common;

use common::git::git;
use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

const S_TWO: &str = "22222222-0000-4000-8000-000000000002";
const S_AUDITOR: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const BASE: &str = "ffffffffffffffffffffffffffffffffffffffff";
/// The cap of the whole printed envelope, in bytes (REQ-1).
const LIMIT: usize = 4096;

struct World {
    home: TempHome,
    project: PathBuf,
    events: std::cell::Cell<i64>,
}

fn world() -> World {
    let home = TempHome::new();
    let project = home.project("demo");
    git(&project, &["init", "-q", "-b", "main"]);
    ledger::register(home.path(), "demo").unwrap();
    World {
        home,
        project,
        events: std::cell::Cell::new(0),
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
    v["result"].clone()
}

/// The 40-hex head of the card with number `n`.
fn head_of(n: i64) -> String {
    format!("{n:040x}")
}

impl World {
    fn project_str(&self) -> String {
        self.project.to_str().unwrap().to_string()
    }

    fn run(&self, args: &[&str]) -> Output {
        let mut cmd = self.home.command(args);
        cmd.env_remove("CLAUDE_CODE_SESSION_ID")
            .env_remove("CODEX_THREAD_ID");
        cmd.output().expect("run moai")
    }

    fn brief_out(&self) -> Output {
        let project = self.project_str();
        self.run(&["leader", "brief", project.as_str()])
    }

    fn brief(&self) -> Value {
        result(&self.brief_out())
    }

    /// The bytes of a read-only command's whole stdout.
    fn bytes_of(&self, args: &[&str]) -> usize {
        let out = self.run(args);
        assert_eq!(out.status.code(), Some(0), "{}", text(&out));
        out.stdout.len()
    }

    fn sql(&self, sql: &str, params: impl rusqlite::Params) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn().execute(sql, params).unwrap();
    }

    fn card(&self, id: i64, state: &str) {
        self.sql(
            "INSERT INTO card (card_id, title, kind, state, priority, write_scope, created_by, \
             created_at, updated_at, closed_at) \
             VALUES (?1, ?2, 'research', ?3, 'p1', '[\"docs/a.md\"]', 'tester', 1, 2, \
             CASE WHEN ?3 = 'done' THEN 3 END)",
            rusqlite::params![id, format!("card {id}"), state],
        );
        self.sql(
            "INSERT INTO card_ac (card_id, seq, text, status, updated_at) \
             VALUES (?1, 1, 'one criterion', 'open', 2)",
            rusqlite::params![id],
        );
    }

    fn reserve_released(&self, card: i64, holder: &str) {
        self.sql(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at, worktree, branch) \
             VALUES (?1, 1, ?2, 'released', 100, 99999999999, 100, NULL, NULL)",
            rusqlite::params![card, holder],
        );
    }

    /// Adds a local note and answers its id.
    fn note(&self, card: i64, author: &str, body: &str) -> i64 {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, 'report', ?2, 'local', ?3, 1000)",
                rusqlite::params![card, author, body],
            )
            .unwrap();
        w.conn().last_insert_rowid()
    }

    /// A card in `ready` whose hand-back head carries a `pass` verdict event of another session.
    fn audited_by_event(&self, card: i64) -> String {
        let head = head_of(card);
        self.card(card, "ready");
        self.reserve_released(card, S_TWO);
        self.note(
            card,
            "agent-1",
            &format!("LANE-READY-INTEGRATION C-{card} head={head}\nbase={BASE}\nchecks passed"),
        );
        let n = self.events.get() + 1;
        self.events.set(n);
        let payload = json!({
            "head": head, "base": BASE, "verdict": "pass", "tier": 1,
            "evidence": "read and ran", "receipt": null, "reviewer": S_AUDITOR,
        });
        self.sql(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, 500, ?2, ?3, 'card.verdict', ?4)",
            rusqlite::params![
                format!("test-verdict-{n}"),
                S_AUDITOR,
                card,
                payload.to_string()
            ],
        );
        head
    }

    /// A card in `ready` whose hand-back head carries a first-line `AUDIT … PASS` note.
    fn audited_by_note(&self, card: i64) -> (String, i64) {
        let head = head_of(card);
        self.card(card, "ready");
        self.reserve_released(card, S_TWO);
        self.note(
            card,
            "agent-1",
            &format!("LANE-READY-INTEGRATION C-{card} head={head}\nbase={BASE}\nchecks passed"),
        );
        let id = self.note(
            card,
            S_AUDITOR,
            &format!("AUDIT C-{card} (x) head {head} base {BASE} — verdict: PASS\nread and ran"),
        );
        (head, id)
    }

    /// The profile whose integration branch is `integ`, with a commit on it, so that git can
    /// answer whether a head is landed.
    fn integration_branch(&self) {
        let cfg = moai_cli::state_root::path(&self.project, "project.yaml");
        fs::create_dir_all(cfg.parent().unwrap()).unwrap();
        fs::write(
            &cfg,
            br#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"], "work_flow": {"flow": "m-flow", "integration_branch": "integ"}}"#,
        )
        .unwrap();
        git(
            &self.project,
            &[
                "-c",
                "user.name=tester",
                "-c",
                "user.email=t@example.invalid",
                "commit",
                "--allow-empty",
                "-q",
                "-m",
                "tip",
            ],
        );
        git(&self.project, &["branch", "integ"]);
    }
}

fn items<'a>(brief: &'a Value, part: &str) -> &'a Vec<Value> {
    brief[part]
        .as_array()
        .unwrap_or_else(|| panic!("no list {part} in {brief}"))
}

fn cards(list: &[Value]) -> Vec<String> {
    list.iter()
        .map(|i| i["card"].as_str().unwrap().to_string())
        .collect()
}

// ---- shape_ (AC-1: REQ-1, REQ-3, REQ-5, REQ-6, REQ-8) -----------------------------------------

#[test]
fn shape_prints_the_parts_with_numbers_and_writes_nothing() {
    let w = world();
    w.integration_branch();
    let by_event = w.audited_by_event(1);
    let (by_note, audit_note) = w.audited_by_note(2);
    w.card(3, "active");
    let blocked_note = w.note(
        3,
        "agent-2",
        "LANE-BLOCKED agent-2 C-3 waits for the holder of src/help.rs · kind=claim owner=lane",
    );
    w.card(4, "active");
    let decision_note = w.note(
        4,
        "agent-3",
        "LANE-BLOCKED agent-3 C-4 pick one of two names · kind=decision owner=leader",
    );
    let before = w.home.snapshot();
    let out = w.brief_out();
    let brief = result(&out);
    assert!(out.stdout.len() <= LIMIT, "{} bytes", out.stdout.len());
    for part in [
        "counts",
        "queue",
        "blocked",
        "decisions",
        "omitted",
        "unread",
    ] {
        assert!(brief.get(part).is_some(), "missing {part}: {brief}");
    }
    assert_eq!(brief["writes"], json!(false));
    assert_eq!(brief["ledger"], json!("ready"));
    assert_eq!(
        brief["counts"],
        json!({"queue": 2, "audit": 0, "blocked": 1, "decisions": 1}),
        "{brief}"
    );
    let queue = items(&brief, "queue");
    assert_eq!(cards(queue), ["C-1", "C-2"], "{brief}");
    assert_eq!(queue[0]["head"], json!(by_event));
    assert!(
        queue[0]["ref"].as_str().unwrap().starts_with("event:"),
        "{brief}"
    );
    assert_eq!(queue[1]["head"], json!(by_note));
    assert_eq!(queue[1]["ref"], json!(format!("note:{audit_note}")));
    let blocked = items(&brief, "blocked");
    assert_eq!(cards(blocked), ["C-3"]);
    assert_eq!(blocked[0]["note"], json!(blocked_note));
    let text = blocked[0]["text"].as_str().unwrap();
    assert!(text.contains("kind=claim"), "{text}");
    assert!(!text.starts_with("LANE-BLOCKED"), "{text}");
    let decisions = items(&brief, "decisions");
    assert_eq!(cards(decisions), ["C-4"]);
    assert_eq!(decisions[0]["note"], json!(decision_note));
    assert_eq!(brief["omitted"], json!({"queue": 0, "blocked": 0, "decisions": 0}));
    assert_eq!(brief["unread"], json!([]), "{brief}");
    // REQ-5 and REQ-8: nothing written, and the same ledger prints the same bytes.
    assert_eq!(
        before,
        w.home.snapshot(),
        "the ledger file, the registry folder and the state folder stay as they were"
    );
    assert_eq!(out.stdout, w.brief_out().stdout);
}

#[test]
fn shape_a_project_without_a_ledger_is_empty_and_exits_zero() {
    let home = TempHome::new();
    let project = home.project("bare");
    git(&project, &["init", "-q", "-b", "main"]);
    let before = home.snapshot();
    let out = home
        .command(&["leader", "brief", project.to_str().unwrap()])
        .output()
        .unwrap();
    let brief = result(&out);
    assert_eq!(brief["ledger"], json!("absent"));
    for part in ["queue", "blocked", "decisions"] {
        assert_eq!(brief[part], json!([]), "{part}: {brief}");
    }
    assert_eq!(
        brief["counts"],
        json!({"queue": 0, "audit": 0, "blocked": 0, "decisions": 0})
    );
    assert_eq!(brief["writes"], json!(false));
    assert_eq!(before, home.snapshot());
}

// ---- size_ (AC-2: REQ-2) ----------------------------------------------------------------------

/// 60 cards with a blocked note of long Korean text and 40 audited heads.
fn crowded() -> World {
    let w = world();
    w.integration_branch();
    for n in 1..=60 {
        w.card(100 + n, "active");
        w.note(
            100 + n,
            "agent-1",
            &format!(
                "LANE-BLOCKED agent-1 C-{} 이 카드는 다른 카드가 쥔 파일을 기다려요, 파일 이름과 이유를 길게 적은 줄이에요 · kind=claim owner=lane next=보유 카드가 파일을 돌려줄 때까지 기다려요",
                100 + n
            ),
        );
    }
    for n in 1..=40 {
        w.audited_by_event(n);
    }
    w
}

#[test]
fn size_the_whole_answer_stays_within_the_cap_and_keeps_the_newest() {
    let w = crowded();
    let out = w.brief_out();
    assert!(
        out.stdout.len() <= LIMIT,
        "{} bytes: {}",
        out.stdout.len(),
        text(&out)
    );
    let brief = result(&out);
    assert_eq!(brief["counts"]["blocked"], json!(60), "{brief}");
    assert_eq!(brief["counts"]["queue"], json!(40), "{brief}");
    for part in ["queue", "blocked", "decisions"] {
        let shown = items(&brief, part).len() as i64;
        let omitted = brief["omitted"][part].as_i64().unwrap();
        assert_eq!(
            shown + omitted,
            brief["counts"][part].as_i64().unwrap(),
            "{part}: {brief}"
        );
    }
    assert!(
        brief["omitted"]["blocked"].as_i64().unwrap() > 0
            || brief["omitted"]["queue"].as_i64().unwrap() > 0,
        "the fixture must not fit: {brief}"
    );
    // The newest notes are kept: the blocked list reads newest first and starts at C-160.
    let blocked = items(&brief, "blocked");
    assert_eq!(blocked[0]["card"], json!("C-160"), "{brief}");
    let notes: Vec<i64> = blocked
        .iter()
        .map(|i| i["note"].as_i64().unwrap())
        .collect();
    assert!(notes.windows(2).all(|p| p[0] > p[1]), "{notes:?}");
    // A cut first line is cut on a character boundary at 96 bytes.
    for item in blocked {
        assert!(item["text"].as_str().unwrap().len() <= 96, "{item}");
    }
}

// ---- split_ (AC-3: REQ-4) ---------------------------------------------------------------------

#[test]
fn split_a_decision_is_told_from_a_block_and_answered_or_done_notes_vanish() {
    let w = world();
    w.integration_branch();
    // C-1 a block, C-2 a decision, C-3 a decision question, C-4 a word that only looks like
    // a decision, C-5 an answered block, C-6 a block of a done card, C-7 a progress line.
    for (id, state) in [
        (1, "active"),
        (2, "active"),
        (3, "active"),
        (4, "active"),
        (5, "active"),
        (6, "done"),
        (7, "active"),
    ] {
        w.card(id, state);
    }
    w.note(1, "agent-1", "LANE-BLOCKED agent-1 C-1 claim · kind=claim owner=lane");
    w.note(
        2,
        "agent-1",
        "LANE-BLOCKED agent-1 C-2 pick a name · kind=decision owner=leader",
    );
    w.note(3, "agent-1", "DECISION-QUESTION agent-1 (C-3) which file");
    w.note(
        4,
        "agent-1",
        "LANE-BLOCKED agent-1 C-4 later · kind=decisions owner=leader",
    );
    w.note(5, "agent-1", "LANE-BLOCKED agent-1 C-5 old · kind=decision owner=leader");
    w.note(5, "leader", "answered: take the first name");
    w.note(6, "agent-1", "LANE-BLOCKED agent-1 C-6 done · kind=decision owner=leader");
    w.note(7, "agent-1", "LANE-NOTE agent-1 C-7 still working");
    let brief = w.brief();
    let mut decisions = cards(items(&brief, "decisions"));
    decisions.sort();
    assert_eq!(decisions, ["C-2", "C-3"], "{brief}");
    let mut blocked = cards(items(&brief, "blocked"));
    blocked.sort();
    assert_eq!(blocked, ["C-1", "C-4"], "{brief}");
    assert_eq!(
        brief["counts"],
        json!({"queue": 0, "audit": 0, "blocked": 2, "decisions": 2})
    );
}

// ---- unread_ (AC-4: REQ-7) --------------------------------------------------------------------

#[test]
fn unread_an_unreadable_integration_branch_is_listed_and_the_rest_still_prints() {
    // No profile: git cannot name an integration branch.
    let w = world();
    w.card(1, "ready");
    w.reserve_released(1, S_TWO);
    w.note(
        1,
        "agent-1",
        &format!("LANE-READY-INTEGRATION C-1 head={}\nbase={BASE}", head_of(1)),
    );
    w.card(2, "active");
    w.note(
        2,
        "agent-2",
        "LANE-BLOCKED agent-2 C-2 pick · kind=decision owner=leader",
    );
    let brief = w.brief();
    let unread: Vec<&str> = items(&brief, "unread")
        .iter()
        .map(|u| u.as_str().unwrap())
        .collect();
    assert_eq!(unread, ["integration"], "{brief}");
    assert_eq!(cards(items(&brief, "decisions")), ["C-2"], "{brief}");
    assert_eq!(brief["counts"]["audit"], json!(1), "{brief}");
}

#[test]
fn unread_an_unreadable_lane_registry_is_listed() {
    let w = world();
    w.integration_branch();
    // The registry folder is a file: it cannot be read as a folder.
    let dir = moai_cli::clone_root::path(&w.project.join(".git")).join("sessions");
    fs::create_dir_all(dir.parent().unwrap()).unwrap();
    fs::write(&dir, b"not a folder").unwrap();
    w.card(1, "active");
    w.note(
        1,
        "agent-1",
        "LANE-BLOCKED agent-1 C-1 claim · kind=claim owner=lane",
    );
    let brief = w.brief();
    let unread: Vec<&str> = items(&brief, "unread")
        .iter()
        .map(|u| u.as_str().unwrap())
        .collect();
    assert_eq!(unread, ["registry"], "{brief}");
    assert_eq!(cards(items(&brief, "blocked")), ["C-1"], "{brief}");
}

// ---- measure_ (AC-5: card criterion 3) --------------------------------------------------------

#[test]
fn measure_the_brief_is_smaller_than_the_two_reads_it_replaces() {
    let w = crowded();
    let project = w.project_str();
    let inbox = w.bytes_of(&["card", "inbox", project.as_str()]);
    let plan = w.bytes_of(&["card", "orchestrate", "plan", project.as_str()]);
    let out = w.brief_out();
    result(&out);
    let brief = out.stdout.len();
    assert!(brief <= LIMIT, "{brief} bytes");
    assert!(
        brief * 4 < inbox + plan,
        "brief {brief} bytes against inbox {inbox} + plan {plan} bytes"
    );
}
