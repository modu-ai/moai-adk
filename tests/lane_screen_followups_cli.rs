//! C-134 (SPEC-20261120-lane-screen-followups): RED tests for the five C-120 audit follow-ups.
//!
//! - `state_shapes` runs against the merged classifier today and is RED now: a screen that
//!   otherwise reads as working, carrying `[1] Yes` / `1: Yes` / `a. Yes` choice rows or the
//!   bracket question `Overwrite? (y/N/a)`, is currently answered `working`.
//! - `send_guard` and `scan_followups` target the `lane-send` / `lane-screen --scan` surface of
//!   C-120's slices B·C, which are not in the integration branch while this RED is written: they
//!   fail on the missing surface until those land, and on the missing behaviour after that, until
//!   the implementation slice widens this card's scope.
//!
//! The Orca program is a fake shell script (`MOAI_ORCA_BIN`) that records every call, so "sent
//! nothing" is "no `send` call". Every case runs in empty temporary folders; the real `~/.moai`
//! is not read or written.

mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = common::MOAI_BIN;
const NOW_TEXT: &str = "2026-11-20T04:00:00Z";

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

impl Run {
    fn envelope(&self) -> Value {
        serde_json::from_str(&self.stdout).unwrap_or(Value::Null)
    }
}

/// `moai lane-screen ARGS` reading the screen JSON `input` from standard input.
fn run_screen(input: &str, extra: &[&str]) -> Run {
    let dir = tempfile::tempdir().unwrap();
    let mut child = Command::new(BIN)
        .arg("lane-screen")
        .arg("--now")
        .arg(NOW_TEXT)
        .args(extra)
        .current_dir(dir.path())
        .env("HOME", dir.path())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap_or(-1),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// An Orca `terminal read --screen --json` answer holding `rows`.
fn orca(rows: &[&str], source: &str) -> String {
    json!({
        "id": "fixture", "ok": true,
        "result": {"terminal": {
            "handle": "term_fixture", "status": "running", "tail": rows,
            "truncated": false, "limited": false, "returnedLineCount": rows.len(),
            "source": source,
        }}
    })
    .to_string()
}

fn state_of(input: &str) -> String {
    let r = run_screen(input, &[]);
    assert_eq!(r.code, 0, "{} / {}", r.stdout, r.stderr);
    let v = r.envelope();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"]["state"].as_str().unwrap().to_string()
}

/// The C-134 choice shapes, each under a working marker (an `esc to interrupt` row and six
/// non-blank rows), so the gap is visible: today all four read as `working`.
fn shape_screens() -> Vec<(&'static str, Vec<&'static str>)> {
    [
        ("bracket-number", vec!["[1] Yes", "[2] No"]),
        ("colon-number", vec!["1: Yes", "2: No"]),
        ("letter-dot", vec!["a. Yes", "b. No"]),
        ("bracket-question", vec!["Overwrite? (y/N/a)"]),
    ]
    .to_vec()
}

fn working_with(rows: Vec<&str>) -> String {
    let mut screen: Vec<&str> = vec![
        "repo main clean",
        "agent idle",
        "nothing to do",
        "waiting",
        "esc to interrupt",
    ];
    screen.extend(rows);
    orca(&screen, "screen")
}

// ---- criterion 1, classifier half ----

#[test]
fn other_choice_shapes_and_bracket_questions_are_never_working() {
    for (name, rows) in shape_screens() {
        let state = state_of(&working_with(rows));
        assert!(
            state == "permission_prompt" || state == "unknown",
            "{name}: expected permission_prompt or unknown, got {state}"
        );
    }
}

#[test]
fn the_plain_working_screen_still_reads_working() {
    let state = state_of(&working_with(vec!["cargo build running"]));
    assert_eq!(state, "working", "the regression guard of this slice");
}

/// A screen with a known question (the C-120 question set) above `rows`, over a working marker.
fn question_with(rows: &[&str]) -> String {
    let mut screen: Vec<&str> = vec![
        "repo main clean",
        "agent working",
        "cargo build running",
        "waiting",
        "esc to interrupt",
        "Do you want to continue?",
    ];
    screen.extend_from_slice(rows);
    orca(&screen, "screen")
}

// ---- criterion 4 (leader scope addition: the re-audit's N1 and N2, one table test) ----

/// The re-audit's N2: under a known question these shapes are sent today (working); under the
/// shape criterion they must refuse.
#[test]
fn n2_shapes_under_a_known_question_are_never_working() {
    for rows in [
        &["Approve / Reject"][..],
        &["Accept / Decline"][..],
        &["OK / Cancel"][..],
        &["Continue / Abort"][..],
        &["[ Proceed ] [ Cancel ]"][..],
        &["[1] Approve / [2] Reject"][..],
        &["1. Proceed"][..],
    ] {
        let state = state_of(&question_with(rows));
        assert!(
            state == "permission_prompt" || state == "unknown",
            "N2 shape {rows:?}: expected permission_prompt or unknown, got {state}"
        );
    }
}

/// The re-audit's N1: the twelve refusals stay refused under the shape criterion (the guard of
/// the rewrite), while question-adjacent prose that is not choice-shaped is freed.
#[test]
fn n1_table_keeps_the_twelve_refusals_and_frees_plain_prose() {
    for rows in [
        &["proceed? (yes/no)"][..],
        &["continue connecting (yes/no/[fingerprint])?"][..],
        &["Save changes? (Yes/No)"][..],
        &["choose: yes or no"][..],
        &["Overwrite? (y/N/a)"][..],
        &["[Y]es [N]o"][..],
        &["Allow", "Deny"][..],
        &["1. Allow once", "2. Allow always", "3. Deny"][..],
        &["1. Y", "2. N"][..],
        &["1. Proceed", "2. Cancel"][..],
        &["[1] Allow", "[2] Deny"][..],
        &["1. run the tests", "2. open the pull request"][..],
    ] {
        let state = state_of(&question_with(rows));
        assert!(
            state == "permission_prompt" || state == "unknown",
            "the twelve refusals must stay refused, {rows:?} got {state}"
        );
    }
    for prose in [
        "no changes yet",
        "allowed 3 lanes so far",
        "yesterday's build is green",
    ] {
        let state = state_of(&question_with(&[prose]));
        assert!(
            state == "working" || state == "idle",
            "plain prose under a question is not a choice: {prose} got {state}"
        );
    }
}

// ---- the fake Orca program (pattern of tests/lane_screen_cli.rs, slice C) ----

/// Records every call in `calls.txt`; `list.json` is what `terminal list` prints, `read-<handle>`
/// what `terminal read` prints; `slow-<handle>` makes the read sleep 30 s (longer than any
/// timeout), `late-<handle>` makes it sleep 2 s; a `send` call lands in `sent.txt`.
const FAKE_ORCA: &str = r#"#!/bin/sh
d="$FAKE_ORCA_DIR"
printf '%s\n' "$*" >> "$d/calls.txt"
case "$2" in
list)
  cat "$d/list.json"
  ;;
read)
  if [ -f "$d/slow-$4" ]; then sleep 30; fi
  if [ -f "$d/late-$4" ]; then sleep 2; fi
  cat "$d/read-$4.json"
  ;;
send)
  printf '%s\n' "$*" >> "$d/sent.txt"
  echo '{"ok":true}'
  ;;
*)
  exit 9
  ;;
esac
"#;

struct Fake {
    dir: tempfile::TempDir,
    bin: PathBuf,
}

impl Fake {
    fn new(list: &str) -> Fake {
        let dir = tempfile::tempdir().unwrap();
        let bin = dir.path().join("orca");
        common::write_executable(&bin, FAKE_ORCA);
        fs::write(dir.path().join("list.json"), list).unwrap();
        Fake { dir, bin }
    }

    /// `handle` reads as the rendered screen of `text`.
    fn screen(&self, handle: &str, text: &str) {
        let rows: Vec<&str> = text.lines().collect();
        fs::write(
            self.dir.path().join(format!("read-{handle}.json")),
            orca(&rows, "screen"),
        )
        .unwrap();
    }

    fn flag(&self, name: &str) {
        fs::write(self.dir.path().join(name), "").unwrap();
    }

    fn sent(&self) -> bool {
        self.dir.path().join("sent.txt").exists()
    }

    fn run(&self, args: &[&str]) -> Run {
        let home = tempfile::tempdir().unwrap();
        let out = Command::new(BIN)
            .args(args)
            .arg("--now")
            .arg(NOW_TEXT)
            .current_dir(self.dir.path())
            .env("HOME", home.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("MOAI_ORCA_BIN", &self.bin)
            .env("FAKE_ORCA_DIR", self.dir.path())
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .output()
            .unwrap();
        Run {
            code: out.status.code().unwrap_or(-1),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }
}

/// An `orca terminal list --json` answer.
fn list(terminals: &[Value], total: usize) -> String {
    json!({
        "id": "fixture", "ok": true,
        "result": {
            "terminals": terminals, "hostScope": "local", "topologyRevisions": {},
            "totalCount": total, "truncated": false,
        }
    })
    .to_string()
}

fn term(handle: &str, title: &str) -> Value {
    json!({
        "handle": handle, "title": title, "branch": "main", "connected": true,
        "writable": true, "worktreePath": "/example/work", "preview": "",
    })
}

fn working_screen_text() -> String {
    "repo main clean\nagent idle\nnothing to do\nwaiting\nesc to interrupt\ncargo build running\n"
        .to_string()
}

// ---- criterion 1, send half (lane-send guard of C-120 slice B + this slice's shapes) ----

#[test]
fn lane_send_refuses_the_new_choice_shapes_and_sends_nothing() {
    for (name, rows) in shape_screens() {
        let fake = Fake::new(&list(&[term("t1", "lane one")], 1));
        let mut screen = working_screen_text();
        for row in rows {
            screen.push_str(row);
            screen.push('\n');
        }
        fake.screen("t1", &screen);
        let r = fake.run(&["lane-send", "--terminal", "t1", "--text", "hello"]);
        assert_eq!(
            r.code, 1,
            "{name}: lane-send must refuse, stdout {} stderr {}",
            r.stdout, r.stderr
        );
        assert!(!fake.sent(), "{name}: no orca send call may happen");
    }
}

// ---- criteria 2 and 3 (--scan follow-ups) ----

#[test]
fn scan_stops_at_the_total_time_cap_and_reports_the_rest_unreachable() {
    let fake = Fake::new(&list(
        &[
            term("t1", "fast"),
            term("t2", "late"),
            term("t3", "never read"),
        ],
        3,
    ));
    fake.screen("t1", &working_screen_text());
    fake.screen("t2", &working_screen_text());
    fake.screen("t3", &working_screen_text());
    fake.flag("late-t2");
    let r = fake.run(&["lane-screen", "--scan", "--max-seconds", "1"]);
    assert_eq!(r.code, 0, "{} / {}", r.stdout, r.stderr);
    let v = r.envelope();
    assert_eq!(v["ok"], json!(true), "{v}");
    let text = r.stdout;
    assert!(
        text.contains("unreachable"),
        "the terminals the cap left unread are reported unreachable: {text}"
    );
    assert!(
        text.contains("max-seconds 1") || text.contains("1 s"),
        "the reason names the cap: {text}"
    );
}

#[test]
fn a_list_over_the_size_cap_fails_naming_the_size() {
    let fake = Fake::new(&list(
        &[term("t1", "one"), term("t2", "two"), term("t3", "three")],
        3,
    ));
    let r = fake.run(&["lane-screen", "--scan", "--max-terminals", "2"]);
    assert_eq!(r.code, 1, "{} / {}", r.stdout, r.stderr);
    assert!(
        r.stdout.contains('3') || r.stderr.contains('3'),
        "the reason names the reported size 3: {} {}",
        r.stdout,
        r.stderr
    );
    assert!(
        r.stdout.contains("max-terminals") || r.stderr.contains("max-terminals"),
        "the reason names the cap: {} {}",
        r.stdout,
        r.stderr
    );
}

#[test]
fn a_long_program_path_does_not_hide_the_timeout_reason() {
    let long_path = format!("/example/{}", "work/".repeat(60));
    let t = json!({
        "handle": "t1", "title": "slow lane", "branch": "main", "connected": true,
        "writable": true, "worktreePath": long_path, "preview": "",
    });
    let fake = Fake::new(&list(&[t], 1));
    fake.screen("t1", &working_screen_text());
    fake.flag("slow-t1");
    let r = fake.run(&["lane-screen", "--scan", "--max-seconds", "30"]);
    assert_eq!(r.code, 0, "{} / {}", r.stdout, r.stderr);
    assert!(
        r.stdout.contains("did not answer within"),
        "the timeout reason stays visible: {}",
        r.stdout
    );
    assert!(
        r.stdout.contains("slow lane"),
        "the terminal is still named: {}",
        r.stdout
    );
}
