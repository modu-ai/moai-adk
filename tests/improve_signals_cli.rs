//! `moai improve-signals` (SPEC-20261121-self-improvement-loop, slice S1): the observed problems of
//! a project, grouped into classes with a count, read from the card ledger and the project's
//! `improve/notes.jsonl` without writing anything.
//!
//! Every command runs against a temporary HOME; the real `~/.moai` is not read or written. The
//! notes are inserted with SQL so their kind, author and time are exactly what a case needs.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::time::{Duration, Instant};

/// 2026-11-01T00:00:00Z, 2026-11-02T00:00:00Z.
const DAY1: i64 = 1_793_491_200;
const DAY2: i64 = 1_793_577_600;

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
    fn command(&self, extra: &[&str]) -> Command {
        let mut args = vec!["improve-signals", self.project.as_str()];
        args.extend_from_slice(extra);
        self.home.command(&args)
    }

    /// `moai improve-signals <project>`, ended by a deadline so a read that blocks fails the
    /// test instead of hanging it.
    fn signals(&self) -> Output {
        within(self.command(&[]), Duration::from_secs(30))
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

    /// One note, in the order of the calls (the note id grows with each).
    fn note(&self, card: i64, kind: &str, body: &str, at: i64) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "INSERT INTO card_note (card_id, kind, author, origin, body, created_at) \
                 VALUES (?1, ?2, 'tester', 'local', ?3, ?4)",
                rusqlite::params![card, kind, body, at],
            )
            .unwrap();
    }

    /// The time a card was created, so a case does not depend on the clock of the run.
    fn card_at(&self, card: i64, at: i64) {
        let w = ledger::open_write(self.home.path(), "demo").unwrap();
        w.conn()
            .execute(
                "UPDATE card SET created_at = ?1 WHERE card_id = ?2",
                rusqlite::params![at, card],
            )
            .unwrap();
    }

    /// An audit FAIL note of the shape the leader writes on a card.
    fn audit_fail(&self, card: i64, body: &str, at: i64) {
        let line = format!(
            "AUDIT C-{card} head {} base {} — verdict: FAIL",
            "a".repeat(40),
            "b".repeat(40)
        );
        self.note(card, "decision", &format!("{line}\n{body}"), at);
    }

    fn blocked(&self, card: i64, kind: &str, at: i64) {
        self.note(
            card,
            "comment",
            &format!("LANE-BLOCKED agent-1 C-{card} waits · kind={kind} owner=leader next=check"),
            at,
        );
    }

    /// The project's `improve/notes.jsonl` (the state folder is named by the product, never
    /// spelled here).
    fn notes_path(&self) -> PathBuf {
        moai_cli::state_root::path(Path::new(&self.project), "improve/notes.jsonl")
    }

    fn write_notes(&self, lines: &[Value]) {
        let path = self.notes_path();
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        let mut body = String::new();
        for l in lines {
            body.push_str(&l.to_string());
            body.push('\n');
        }
        std::fs::write(path, body).unwrap();
    }
}

fn within(mut cmd: Command, limit: Duration) -> Output {
    let mut child = cmd
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("run moai");
    let start = Instant::now();
    loop {
        if child.try_wait().unwrap().is_some() {
            return child.wait_with_output().unwrap();
        }
        if start.elapsed() > limit {
            let _ = child.kill();
            let _ = child.wait();
            panic!("moai did not finish within {limit:?}: a read may be blocked");
        }
        std::thread::sleep(Duration::from_millis(20));
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
    assert_eq!(v["operation"], json!("improve.signals"), "{v}");
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

/// The class named `key`, if the result has one.
fn class<'a>(r: &'a Value, key: &str) -> Option<&'a Value> {
    r["classes"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["key"] == json!(key))
}

fn keys(r: &Value) -> Vec<String> {
    r["classes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|c| c["key"].as_str().unwrap().to_string())
        .collect()
}

fn improve(kind: &str, key: &str, at: &str, text: &str) -> Value {
    json!({"v": 1, "at": at, "kind": kind, "key": key, "text": text, "by": "tester"})
}

fn fifo(path: &Path) {
    std::fs::create_dir_all(path.parent().unwrap()).unwrap();
    let c = std::ffi::CString::new(path.to_str().unwrap()).unwrap();
    // SAFETY: mkfifo with a valid NUL-terminated path.
    assert_eq!(unsafe { libc::mkfifo(c.as_ptr(), 0o600) }, 0, "{path:?}");
}

#[test]
fn every_source_becomes_a_class_with_count_times_and_examples() {
    let w = world();
    w.cards(&["work", "flaky: some_test"]);
    w.audit_fail(
        1,
        "src/a.rs:12 has no test; src/b.rs is out of scope",
        DAY1 + 10,
    );
    w.audit_fail(1, "src/a.rs again", DAY1 + 20);
    w.blocked(1, "ci", DAY1 + 30);
    w.blocked(1, "ci", DAY1 + 40);
    w.write_notes(&[
        improve(
            "correction",
            "naming",
            "2026-11-01T00:00:00Z",
            "the user asked for Korean terms",
        ),
        improve(
            "correction",
            "naming",
            "2026-11-02T00:00:00Z",
            "asked again",
        ),
    ]);
    let r = result(&w.signals());

    assert_eq!(r["project_id"], json!("demo"));
    assert_eq!(r["ledger"], json!("ready"));
    assert_eq!(r["writes"], json!(false));
    assert_eq!(r["host_memory"], json!("off"));
    assert_eq!(r["truncated"], json!(false));

    let a = class(&r, "audit_fail:src/a.rs").expect("a.rs class");
    assert_eq!(a["kind"], json!("audit_fail"));
    assert_eq!(a["count"], json!(2));
    assert_eq!(a["status"], json!("candidate"));
    assert_eq!(a["first_seen"], json!(DAY1 + 10));
    assert_eq!(a["last_seen"], json!(DAY1 + 20));
    assert_eq!(class(&r, "audit_fail:src/b.rs").unwrap()["count"], json!(1));
    assert_eq!(class(&r, "block:ci").unwrap()["count"], json!(2));
    assert_eq!(class(&r, "flake:some_test").unwrap()["count"], json!(1));
    let c = class(&r, "correction:naming").unwrap();
    assert_eq!(c["count"], json!(2));
    assert_eq!(c["first_seen"], json!(DAY1));
    assert_eq!(c["last_seen"], json!(DAY2));
    // The fixtures hold no other class.
    assert_eq!(keys(&r).len(), 5, "{:?}", keys(&r));
    // An example names where the signal came from and carries a short text.
    let first = &a["signals"][0];
    assert!(
        first["ref"].as_str().unwrap().starts_with("C-1#"),
        "{first}"
    );
    assert!(
        first["text"].as_str().unwrap().starts_with("AUDIT C-1"),
        "{first}"
    );
}

#[test]
fn a_class_shows_at_most_five_examples_and_the_newest_ones() {
    let w = world();
    w.cards(&["work"]);
    for i in 0..7 {
        w.audit_fail(1, "src/many.rs", DAY1 + i);
    }
    let r = result(&w.signals());
    let c = class(&r, "audit_fail:src/many.rs").unwrap();
    assert_eq!(c["count"], json!(7));
    let signals = c["signals"].as_array().unwrap();
    assert_eq!(signals.len(), 5);
    let times: Vec<i64> = signals.iter().map(|s| s["at"].as_i64().unwrap()).collect();
    assert_eq!(
        times,
        vec![DAY1 + 2, DAY1 + 3, DAY1 + 4, DAY1 + 5, DAY1 + 6]
    );
}

#[test]
fn one_signal_is_observed_once_and_two_are_candidates_for_every_source() {
    let w = world();
    w.cards(&["work", "flaky: edge_test"]);
    w.audit_fail(1, "src/edge.rs", DAY1);
    w.blocked(1, "audit", DAY1);
    w.write_notes(&[improve(
        "failure",
        "cargo-test",
        "2026-11-01T00:00:00Z",
        "red",
    )]);
    let r = result(&w.signals());
    for key in [
        "audit_fail:src/edge.rs",
        "block:audit",
        "flake:edge_test",
        "failure:cargo-test",
    ] {
        let c = class(&r, key).unwrap_or_else(|| panic!("{key}: {:?}", keys(&r)));
        assert_eq!(c["count"], json!(1), "{key}");
        assert_eq!(c["status"], json!("observed_once"), "{key}");
    }

    // The second signal of each source: a second audit note, a second blocked note, a note on
    // the flaky card (the card plus one recurrence), a second line in the notes file.
    w.audit_fail(1, "src/edge.rs", DAY2);
    w.blocked(1, "audit", DAY2);
    w.note(2, "observation", "failed again on the same head", DAY2);
    w.write_notes(&[
        improve("failure", "cargo-test", "2026-11-01T00:00:00Z", "red"),
        improve("failure", "cargo-test", "2026-11-02T00:00:00Z", "red again"),
    ]);
    let r = result(&w.signals());
    for key in [
        "audit_fail:src/edge.rs",
        "block:audit",
        "flake:edge_test",
        "failure:cargo-test",
    ] {
        let c = class(&r, key).unwrap();
        assert_eq!(c["count"], json!(2), "{key}");
        assert_eq!(c["status"], json!("candidate"), "{key}");
    }
}

#[test]
fn an_agents_own_claim_is_never_a_signal() {
    let w = world();
    w.cards(&["work"]);
    for (kind, body) in [
        ("comment", "LANE-READY agent-1 C-1 fixed src/a.rs"),
        ("comment", "LANE-DONE agent-1 C-1 src/a.rs is repaired"),
        (
            "report",
            "CANDIDATE C-1 merge=cccccccccccccccccccccccccccccccccccccccc src/a.rs",
        ),
        (
            "report",
            "MERGE-READY C-1 base=dddddddddddddddddddddddddddddddddddddddd src/a.rs",
        ),
        ("report", "MERGED-LOCAL C-1 src/a.rs"),
        ("comment", "fixed src/a.rs, all tests pass"),
        (
            "comment",
            "verdict: FAIL is what the audit said about src/a.rs",
        ),
    ] {
        w.note(1, kind, body, DAY1);
    }
    w.write_notes(&[improve(
        "false_claim",
        "ready-without-tests",
        "2026-11-01T00:00:00Z",
        "said passed, did not run",
    )]);
    let r = result(&w.signals());
    assert_eq!(
        keys(&r),
        vec!["false_claim:ready-without-tests".to_string()],
        "{r}"
    );
    assert_eq!(r["skipped"]["no_rule"], json!(7));
    assert_eq!(r["skipped"]["bad_line"], json!(0));
}

#[test]
fn an_audit_note_cites_each_path_once_and_a_note_without_a_path_is_uncited() {
    let w = world();
    w.cards(&["work"]);
    w.audit_fail(
        1,
        "src/a.rs:3 and src/a.rs:9 and docs/x.md, see also tests/t.py.",
        DAY1,
    );
    w.audit_fail(1, "no path is named here", DAY1 + 1);
    let r = result(&w.signals());
    assert_eq!(class(&r, "audit_fail:src/a.rs").unwrap()["count"], json!(1));
    assert_eq!(
        class(&r, "audit_fail:docs/x.md").unwrap()["count"],
        json!(1)
    );
    assert_eq!(
        class(&r, "audit_fail:tests/t.py").unwrap()["count"],
        json!(1)
    );
    assert_eq!(class(&r, "audit_fail:uncited").unwrap()["count"], json!(1));
    // A PASS verdict and a blocked note without `kind=` are read by their own rules.
    w.note(
        1,
        "decision",
        "AUDIT C-1 head x base y — verdict: PASS\nsrc/a.rs fine",
        DAY1 + 2,
    );
    w.note(
        1,
        "comment",
        "LANE-BLOCKED agent-1 C-1 no kind word here",
        DAY1 + 3,
    );
    let r = result(&w.signals());
    assert_eq!(class(&r, "audit_fail:src/a.rs").unwrap()["count"], json!(1));
    assert_eq!(class(&r, "block:unknown").unwrap()["count"], json!(1));
}

#[test]
fn bad_lines_of_the_notes_file_are_counted_and_skipped() {
    let w = world();
    w.cards(&["work"]);
    let path = w.notes_path();
    std::fs::create_dir_all(path.parent().unwrap()).unwrap();
    let good = improve("correction", "tone", "2026-11-01T00:00:00Z", "ok").to_string();
    let unknown_kind = improve("praise", "tone", "2026-11-01T00:00:00Z", "ok").to_string();
    let bad_time = improve("correction", "tone", "yesterday", "ok").to_string();
    std::fs::write(
        &path,
        format!("{good}\nnot json at all\n{unknown_kind}\n{bad_time}\n\n{good}\n"),
    )
    .unwrap();
    let r = result(&w.signals());
    assert_eq!(class(&r, "correction:tone").unwrap()["count"], json!(2));
    assert_eq!(r["skipped"]["bad_line"], json!(3));
}

#[test]
fn nothing_is_written_with_or_without_a_ledger_and_an_absent_source_is_reported() {
    let w = world();
    w.cards(&["work", "flaky: t"]);
    w.audit_fail(1, "src/a.rs", DAY1);
    w.write_notes(&[improve("correction", "tone", "2026-11-01T00:00:00Z", "ok")]);
    let before = w.home.snapshot();
    let r = result(&w.signals());
    assert!(!keys(&r).is_empty());
    assert_eq!(
        w.home.snapshot(),
        before,
        "no file, folder, size or time changed"
    );

    // A project with no ledger and no notes file: both are reported, nothing is created.
    let bare = TempHome::new();
    let project = bare.project("demo").to_str().unwrap().to_string();
    let snapshot = bare.snapshot();
    let r = result(&within(
        bare.command(&["improve-signals", &project]),
        Duration::from_secs(30),
    ));
    assert_eq!(r["ledger"], json!("absent"));
    assert_eq!(r["notes_file"], json!("absent"));
    assert_eq!(r["classes"], json!([]));
    assert_eq!(r["writes"], json!(false));
    assert_eq!(
        bare.snapshot(),
        snapshot,
        "no improve folder and no ledger file appears"
    );
}

#[test]
fn a_broken_ledger_and_an_unsafe_notes_file_are_refused() {
    let w = world();
    w.cards(&["work"]);

    // A symbolic link where the notes file should be.
    let path = w.notes_path();
    std::fs::create_dir_all(path.parent().unwrap()).unwrap();
    let target = w.home.path().join("elsewhere.jsonl");
    std::fs::write(&target, "{}\n").unwrap();
    symlink(&target, &path).unwrap();
    let before = w.home.snapshot();
    let (kind, message) = refusal(&w.signals());
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("improve-signals refused"), "{message}");
    assert_eq!(w.home.snapshot(), before, "a refusal writes nothing");
    std::fs::remove_file(&path).unwrap();

    // A named pipe where the notes file should be: it is not read (a read would block).
    fifo(&path);
    let (kind, message) = refusal(&w.signals());
    assert_eq!(kind, "invalid_input", "{message}");
    std::fs::remove_file(&path).unwrap();

    // A file that is not the ledger: the ledger's own refusal, and the file is left alone.
    let ledger_file = ledger::ledger_path(w.home.path(), "demo").unwrap();
    std::fs::write(&ledger_file, b"this is not a database").unwrap();
    let before = w.home.snapshot();
    let (kind, message) = refusal(&w.signals());
    assert_eq!(kind, "blocked", "{message}");
    assert!(message.contains("ledger_unavailable"), "{message}");
    assert_eq!(w.home.snapshot(), before);
}

#[test]
fn host_memory_is_never_opened() {
    let w = world();
    w.cards(&["work"]);
    w.audit_fail(1, "src/a.rs", DAY1);
    // Where Claude auto memory and Codex memories would be: a read of either blocks forever.
    fifo(&w.home.path().join(".claude/projects/demo/memory/MEMORY.md"));
    fifo(&w.home.path().join(".codex/memories/m.md"));
    let r = result(&w.signals());
    assert_eq!(r["host_memory"], json!("off"));
    assert_eq!(class(&r, "audit_fail:src/a.rs").unwrap()["count"], json!(1));
}

#[test]
fn texts_are_masked_and_cut_to_120_characters() {
    let w = world();
    w.cards(&["work"]);
    // A secret-shaped word in a card note's first line and in an improve note.
    w.note(
        1,
        "decision",
        "AUDIT C-1 head x base y — verdict: FAIL password=hunter2-secret src/a.rs",
        DAY1,
    );
    w.audit_fail(1, &"y".repeat(10), DAY1 + 1);
    w.write_notes(&[improve(
        "correction",
        "keys",
        "2026-11-01T00:00:00Z",
        "the token=abcd1234 was pasted",
    )]);
    let out = w.signals();
    let all = text(&out);
    assert!(!all.contains("hunter2-secret"), "{all}");
    assert!(!all.contains("abcd1234"), "{all}");
    assert!(all.contains("[redacted]"), "{all}");

    // A first line over 120 characters is cut at 120.
    let w = world();
    w.cards(&["work"]);
    w.note(
        1,
        "decision",
        &format!("AUDIT C-1 — verdict: FAIL {}", "x".repeat(300)),
        DAY1,
    );
    let r = result(&w.signals());
    let c = class(&r, "audit_fail:uncited").unwrap();
    let shown = c["signals"][0]["text"].as_str().unwrap();
    assert_eq!(shown.chars().count(), 120, "{shown}");
}

#[test]
fn a_flaky_card_counts_the_card_and_its_recurrence_notes_only() {
    let w = world();
    w.cards(&["flaky: slow_one (src/x.rs:9) fails under load", "work"]);
    w.card_at(1, DAY1);
    // The test name is the first word after the prefix; the rest of the title is not part of it.
    w.note(1, "observation", "second failure, run 2", DAY1 + 5);
    w.note(1, "report", "third failure, run 3", DAY1 + 6);
    w.note(1, "comment", "a remark that is not a recurrence", DAY1 + 7);
    w.note(
        1,
        "decision",
        "AUDIT C-1 — verdict: FAIL src/x.rs",
        DAY1 + 8,
    );
    let r = result(&w.signals());
    let c = class(&r, "flake:slow_one").unwrap();
    assert_eq!(c["count"], json!(3));
    assert_eq!(c["last_seen"], json!(DAY1 + 6));
    // The audit note on the flaky card is an audit signal, as it is on any card.
    assert_eq!(class(&r, "audit_fail:src/x.rs").unwrap()["count"], json!(1));
}
