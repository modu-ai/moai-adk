//! SPEC-20261120-lane-screen-followups: tests for the follow-ups of the lane-screen audits.
//!
//! - The classifier half runs `moai lane-screen` on a screen that otherwise reads as working and
//!   carries `[1] Yes` / `1: Yes` / `a. Yes` choice rows, the bracket question `Overwrite? (y/N/a)`
//!   or a choice-shaped row under a known question; it must never answer `working` or `idle`.
//! - The send half runs `moai lane-send` against the same screens through a fake Orca and checks
//!   that no `send` call happens.
//! - `scan_followups` runs `moai lane-screen --scan` with a total time cap, a terminal-count cap,
//!   a list larger than the output limit and a long program path.
//!
//! The Orca program is a fake shell script (`MOAI_ORCA_BIN`) that records every call, so "sent
//! nothing" is "no `send` call". Every case runs in empty temporary folders; the real `~/.moai`
//! is not read or written.

mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
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

/// The choice shapes of the follow-up, each under a working marker (an `esc to interrupt` row
/// and six non-blank rows), so the gap is visible: all four read as `working` before the change.
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

// ---- AC-1, classifier half ----

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
    assert_eq!(state, "working", "the regression guard of this change");
}

/// A screen with a known question above `rows`, over a working marker.
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

// ---- AC-4: the shape criterion (one table test per direction) ----

/// Under a known question these shapes read as working before the change; under the shape
/// criterion they must refuse.
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

/// The twelve refusals stay refused under the shape criterion (the guard of the rewrite), while
/// question-adjacent prose that is not choice-shaped is freed.
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

// ---- the fake Orca program ----

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

    /// The same fake program, but at a path of about 300 characters.
    fn with_long_program_path(list: &str) -> Fake {
        let dir = tempfile::tempdir().unwrap();
        let mut folder = dir.path().to_path_buf();
        for part in ["a", "b", "c"] {
            folder.push(part.repeat(90));
        }
        fs::create_dir_all(&folder).unwrap();
        let bin = folder.join("orca");
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
        self.run_with_env(args, &[])
    }

    fn run_with_env(&self, args: &[&str], env: &[(&str, &str)]) -> Run {
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
            .envs(env.iter().copied())
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

// ---- AC-1, send half ----

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

// ---- AC-2 and AC-3: the --scan follow-ups ----

mod scan_followups {
    use super::*;

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
        let terminals = v["result"]["terminals"].as_array().unwrap();
        assert_eq!(terminals.len(), 3, "every listed terminal has a row: {v}");
        assert_eq!(terminals[0]["state"], json!("working"), "{v}");
        assert_eq!(terminals[2]["state"], json!("unreachable"), "{v}");
    }

    #[test]
    fn a_list_over_the_terminal_count_cap_fails_naming_the_size() {
        let fake = Fake::new(&list(
            &[term("t1", "one"), term("t2", "two"), term("t3", "three")],
            3,
        ));
        let r = fake.run(&["lane-screen", "--scan", "--max-terminals", "2"]);
        assert_eq!(r.code, 1, "{} / {}", r.stdout, r.stderr);
        let text = format!("{} {}", r.stdout, r.stderr);
        assert!(
            text.contains('3'),
            "the reason names the reported size 3: {text}"
        );
        assert!(
            text.contains("max-terminals"),
            "the reason names the cap: {text}"
        );
        assert!(
            !fake.dir.path().join("calls.txt").exists()
                || !fs::read_to_string(fake.dir.path().join("calls.txt"))
                    .unwrap()
                    .contains("read"),
            "an over-large list is never partially scanned"
        );
    }

    /// D-7: a list answer longer than the output limit (1 MiB) fails with a reason that names
    /// the size, not with the wrong reason that the answer is not an Orca answer.
    #[test]
    fn a_list_output_over_the_output_limit_fails_naming_the_size() {
        let mut big = term("t1", "huge");
        big["preview"] = json!("x".repeat(1_200_000));
        let fake = Fake::new(&list(&[big], 1));
        let r = fake.run(&["lane-screen", "--scan"]);
        assert_eq!(r.code, 1, "{} / {}", r.stdout.len(), r.stderr);
        let text = format!("{} {}", r.stdout, r.stderr);
        assert!(
            text.contains("1 MiB"),
            "the reason names the size limit: {text}"
        );
        assert!(
            !text.contains("is not an Orca answer"),
            "the wrong reason is gone: {text}"
        );
    }

    #[test]
    fn a_long_program_path_does_not_hide_the_timeout_reason() {
        let t = json!({
            "handle": "t1", "title": "slow lane", "branch": "main", "connected": true,
            "writable": true, "worktreePath": "/example/work", "preview": "",
        });
        let fake = Fake::with_long_program_path(&list(&[t], 1));
        fake.screen("t1", &working_screen_text());
        fake.flag("late-t1");
        let r = fake.run_with_env(
            &["lane-screen", "--scan"],
            &[("MOAI_ORCA_TIMEOUT_SECS", "1")],
        );
        assert_eq!(r.code, 0, "{} / {}", r.stdout, r.stderr);
        assert!(
            fake.bin.to_string_lossy().len() > 250,
            "the case needs a long program path"
        );
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
}
