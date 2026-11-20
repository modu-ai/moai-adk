//! SPEC-20261120-lane-limit-resume: `moai lane-resume` reads lane screens, names the reset
//! instant in UTC of every lane a usage limit stalls, and sends the given resume line once to
//! each stalled lane whose reset instant has passed — never into a confirmation prompt.
//!
//! The Orca program is a fake shell script (`MOAI_ORCA_BIN`) in the shape of
//! `lane_send_cli.rs`: a `terminal read` prints the next prepared screen read, a `terminal
//! list` prints `list.json`, and a `terminal send` appends its arguments to `sent.txt` in the
//! fake's folder. "Nothing was sent" is therefore "`sent.txt` does not exist". The screen rows
//! are synthetic or read from `tests/fixtures/lane_screens/`. Nothing here touches a real
//! terminal.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;
const NOW_TEXT: &str = "2026-11-20T08:00:00Z";
const TERMINAL: &str = "term_lane9";

/// The glm usage-limit row: the reset time is shown in UTC+8, so 18:23 there is 10:23 UTC.
const LIMIT_ROW: &str =
    "Usage limit reached for 5 hour. Your limit will reset at 2026-11-20 18:23 (UTC+8)";
const LIMIT_ROW_PAST: &str =
    "Usage limit reached for 5 hour. Your limit will reset at 2026-11-20 16:23 (UTC+8)";
const LIMIT_ROW_FAR: &str =
    "Usage limit reached for 5 hour. Your limit will reset at 2026-11-21 16:23 (UTC+8)";
const LIMIT_ROW_UNKNOWABLE: &str = "Usage limit reached. Your limit will reset soon.";
const RESUME_TEXT: &str = "/moai --mode-agent 계속해 주세요";

/// The fake `orca`. `read-N.json` is the N-th read (the last one repeats), `list.json` is the
/// answer of `terminal list`, `read-fail-N` makes the N-th read exit 3, `send-fail` makes a
/// send exit 4.
const FAKE_ORCA: &str = r#"#!/bin/sh
d="$FAKE_ORCA_DIR"
case "$2" in
read)
  n=$(cat "$d/count" 2>/dev/null || echo 0)
  n=$((n+1))
  echo "$n" > "$d/count"
  if [ -f "$d/read-fail-$n" ]; then echo "terminal not found" >&2; exit 3; fi
  f="$d/read-$n.json"
  [ -f "$f" ] || f=$(ls "$d"/read-*.json 2>/dev/null | sort | tail -1)
  cat "$f"
  ;;
list)
  cat "$d/list.json"
  ;;
send)
  printf '%s\n' "$*" >> "$d/sent.txt"
  if [ -f "$d/send-fail" ]; then echo "send refused" >&2; exit 4; fi
  echo '{"id":"x","ok":true,"result":{"send":{"accepted":true}}}'
  ;;
*)
  exit 9
  ;;
esac
"#;

fn fixture(name: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/lane_screens")
        .join(name);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

/// An Orca `terminal read --screen --json` answer holding the rows of `text`.
fn orca_read(rows: &[&str]) -> String {
    json!({
        "id": "fixture", "ok": true,
        "result": {"terminal": {
            "handle": TERMINAL, "status": "running", "tail": rows,
            "truncated": false, "limited": false,
            "returnedLineCount": rows.len(), "source": "screen",
        }}
    })
    .to_string()
}

fn limit_screen(row: &str) -> String {
    orca_read(&["● Working (5m · esc to interrupt)", "", row, "", "> "])
}

struct Fake {
    dir: tempfile::TempDir,
    bin: PathBuf,
}

impl Fake {
    /// A fake whose N-th read is `reads[N-1]` (each an Orca answer as text).
    fn new(reads: &[String]) -> Fake {
        let dir = tempfile::tempdir().unwrap();
        let bin = dir.path().join("orca");
        common::write_executable(&bin, FAKE_ORCA);
        for (i, read) in reads.iter().enumerate() {
            fs::write(dir.path().join(format!("read-{}.json", i + 1)), read).unwrap();
        }
        Fake { dir, bin }
    }

    fn one(read: String) -> Fake {
        Fake::new(&[read])
    }

    /// What was sent, one line per `orca terminal send`; `None` when nothing was.
    fn sent(&self) -> Option<Vec<String>> {
        fs::read_to_string(self.dir.path().join("sent.txt"))
            .ok()
            .map(|t| t.lines().map(String::from).collect())
    }

    /// How many times `terminal read` ran.
    fn reads(&self) -> usize {
        fs::read_to_string(self.dir.path().join("count"))
            .map(|t| t.trim().parse().unwrap())
            .unwrap_or(0)
    }
}

struct Run {
    code: i32,
    envelope: Value,
    stderr: String,
}

impl Run {
    fn ok(&self) -> &Value {
        assert_eq!(self.code, 0, "{} / {}", self.envelope, self.stderr);
        assert_eq!(self.envelope["ok"], json!(true), "{}", self.envelope);
        assert_eq!(self.envelope["operation"], json!("lane.resume"));
        &self.envelope["result"]
    }

    /// The refusal or failure message: a non-zero code and an error envelope.
    fn error(&self) -> &str {
        assert_ne!(self.code, 0, "expected a non-zero code: {}", self.envelope);
        assert_eq!(self.envelope["ok"], json!(false), "{}", self.envelope);
        self.envelope["error"]["message"].as_str().unwrap()
    }
}

/// `moai lane-resume --text TEXT [extra]` with `bin` as the Orca program; `terminal` of `Some`
/// reads one screen, `None` scans the list.
fn run_with(
    bin: &Path,
    fake_dir: &Path,
    terminal: Option<&str>,
    text: &str,
    extra: &[&str],
) -> Run {
    let home = tempfile::tempdir().unwrap();
    let default_now: &[&str] = if extra.contains(&"--now") {
        &[]
    } else {
        &["--now", NOW_TEXT]
    };
    let mut cmd = Command::new(BIN);
    cmd.arg("lane-resume");
    if let Some(t) = terminal {
        cmd.arg(format!("--terminal={t}"));
    }
    let out = cmd
        .arg(format!("--text={text}"))
        .args(default_now)
        .args(extra)
        .current_dir(home.path())
        .env("HOME", home.path())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("MOAI_ORCA_BIN", bin)
        .env("FAKE_ORCA_DIR", fake_dir)
        .env("MOAI_ORCA_TIMEOUT_SECS", "2")
        .output()
        .unwrap();
    Run {
        code: out.status.code().unwrap(),
        envelope: serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "an envelope on stdout; stdout {:?} stderr {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        }),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

fn run(fake: &Fake, terminal: Option<&str>, text: &str, extra: &[&str]) -> Run {
    run_with(&fake.bin, fake.dir.path(), terminal, text, extra)
}

/// AC-1 (REQ-1): the 429 row (UTC+8) yields the reset instant in UTC, named as waiting; nothing
/// is sent.
#[test]
fn a_limit_screen_names_the_reset_instant_utc() {
    let fake = Fake::one(limit_screen(LIMIT_ROW));
    let run = run(&fake, Some(TERMINAL), RESUME_TEXT, &[]);
    let message = run.error();
    assert!(
        message.contains("2026-11-20T10:23:00Z"),
        "the reset instant in UTC: {message}"
    );
    assert!(message.contains("waiting"), "named as waiting: {message}");
    assert!(message.contains(TERMINAL), "the lane by handle: {message}");
    assert!(fake.sent().is_none(), "nothing was sent");
    assert_eq!(fake.reads(), 1, "one read decides the waiting answer");
}

/// AC-2 (REQ-2): past the reset instant the resume line goes out, exactly once, and the answer
/// names the state, the evidence row and the reset instant.
#[test]
fn resume_sends_once_after_the_reset() {
    let fake = Fake::one(limit_screen(LIMIT_ROW_PAST));
    let run = run(
        &fake,
        Some(TERMINAL),
        RESUME_TEXT,
        &["--now", "2026-11-20T10:30:00Z", "--enter"],
    );
    let result = run.ok();
    assert_eq!(result["sent"], json!(true), "{}", result);
    let line = &result["lines"][0];
    assert_eq!(line["terminal"], json!(TERMINAL), "{}", result);
    assert_eq!(line["enter"], json!(true), "{}", result);
    let sent = fake.sent().unwrap();
    assert_eq!(sent.len(), 1, "one line: {sent:?}");
    assert!(sent[0].contains(TERMINAL), "to the lane: {:?}", sent[0]);
    assert!(
        sent[0].contains(RESUME_TEXT),
        "the resume text: {:?}",
        sent[0]
    );
    assert!(
        sent[0].contains("--enter"),
        "enter passed on: {:?}",
        sent[0]
    );
    let named = result.to_string();
    assert!(
        named.contains("2026-11-20T08:23:00Z"),
        "the reset instant in UTC: {named}"
    );
    assert!(named.contains("usage_limit"), "the state: {named}");
}

/// AC-3 (REQ-3): a confirmation prompt never receives the text.
#[test]
fn a_prompt_never_receives_the_text() {
    let fake = Fake::one(orca_read(
        &fixture("synthetic-permission-prompt.txt")
            .lines()
            .collect::<Vec<_>>(),
    ));
    let run = run(&fake, Some(TERMINAL), RESUME_TEXT, &[]);
    let message = run.error();
    assert!(
        message.contains("permission_prompt"),
        "the refusal names the state: {message}"
    );
    assert!(fake.sent().is_none(), "nothing was sent");
    assert_eq!(fake.reads(), 1, "the refusal needs no second read");
}

/// AC-4 (REQ-4): the screen is read again just before the send; a lane that recovered gets
/// nothing and its new state is named.
#[test]
fn the_screen_is_read_again_just_before_the_send() {
    let idle = orca_read(&fixture("synthetic-idle.txt").lines().collect::<Vec<_>>());
    let fake = Fake::new(&[limit_screen(LIMIT_ROW_PAST), idle]);
    let run = run(
        &fake,
        Some(TERMINAL),
        RESUME_TEXT,
        &["--now", "2026-11-20T10:30:00Z"],
    );
    let message = run.error();
    assert!(message.contains("idle"), "the new state: {message}");
    assert!(fake.sent().is_none(), "nothing was sent");
    assert_eq!(fake.reads(), 2, "the screen is read twice");
}

/// AC-5 (REQ-5, REQ-12): the text is checked before any `orca` call.
#[test]
fn the_text_is_checked_before_orca_runs() {
    let fake = Fake::one(limit_screen(LIMIT_ROW_PAST));
    let run = run(&fake, Some(TERMINAL), "-resume-now", &[]);
    let message = run.error();
    assert!(message.contains("refused"), "the refusal: {message}");
    assert!(fake.sent().is_none(), "nothing was sent");
    assert_eq!(fake.reads(), 0, "no orca call before the text check");
}

/// AC-6 (REQ-8): `--wait` with a reset beyond the sleep cap sleeps for no time, sends nothing
/// and names the lane as waiting.
#[test]
fn wait_beyond_the_cap_is_refused_without_sleeping() {
    let fake = Fake::one(limit_screen(LIMIT_ROW_FAR));
    let run = run(&fake, Some(TERMINAL), RESUME_TEXT, &["--wait"]);
    let message = run.error();
    assert!(
        message.contains("2026-11-21T08:23:00Z"),
        "the instant in UTC: {message}"
    );
    assert!(message.contains("waiting"), "named as waiting: {message}");
    assert!(fake.sent().is_none(), "nothing was sent");
    assert_eq!(fake.reads(), 1, "no re-read without a sleep");
}

/// AC-7 (REQ-9, REQ-10): `--terminal` reads only that terminal even though the list holds two.
#[test]
fn terminal_narrow_reads_only_that_terminal() {
    let list = json!({
        "id": "fixture", "ok": true,
        "result": {"terminals": [
            {"handle": "term_other", "title": "other", "branch": "main"},
            {"handle": TERMINAL, "title": "mine", "branch": "main"}
        ]}
    })
    .to_string();
    let dir = tempfile::tempdir().unwrap();
    let bin = dir.path().join("orca");
    common::write_executable(&bin, FAKE_ORCA);
    fs::write(dir.path().join("list.json"), list).unwrap();
    fs::write(dir.path().join("read-1.json"), limit_screen(LIMIT_ROW_PAST)).unwrap();
    let fake = Fake { dir, bin };
    let run = run(
        &fake,
        Some(TERMINAL),
        RESUME_TEXT,
        &["--now", "2026-11-20T10:30:00Z"],
    );
    run.ok();
    assert_eq!(
        fake.reads(),
        2,
        "only that terminal, read twice: before and just before the send"
    );
    assert_eq!(fake.sent().map(|s| s.len()), Some(1));
}

/// The scan: the due lane gets its one line, a prompt lane is named refused, and one send
/// keeps the exit at 0.
#[test]
fn scan_resumes_the_due_lane_and_names_a_prompt_lane() {
    let list = json!({
        "id": "fixture", "ok": true,
        "result": {"terminals": [
            {"handle": "term_ask", "title": "ask", "branch": "main"},
            {"handle": "term_due", "title": "due", "branch": "main"}
        ]}
    })
    .to_string();
    let prompt = orca_read(
        &fixture("synthetic-permission-prompt.txt")
            .lines()
            .collect::<Vec<_>>(),
    );
    let dir = tempfile::tempdir().unwrap();
    let bin = dir.path().join("orca");
    common::write_executable(&bin, FAKE_ORCA);
    fs::write(dir.path().join("list.json"), list).unwrap();
    // The scan reads in list order: term_ask shows the prompt, term_due the past reset. The
    // pre-send look of term_due is the third read, and the fake repeats the last prepared
    // screen — term_due's own — so the limit row still holds and the line goes out.
    fs::write(dir.path().join("read-1.json"), prompt).unwrap();
    fs::write(dir.path().join("read-2.json"), limit_screen(LIMIT_ROW_PAST)).unwrap();
    let fake = Fake { dir, bin };
    let run = run(&fake, None, RESUME_TEXT, &["--now", "2026-11-20T10:30:00Z"]);
    let result = run.ok();
    let named = result.to_string();
    assert!(named.contains("term_due"), "the due lane: {named}");
    assert!(named.contains("term_ask"), "the prompt lane: {named}");
    assert!(
        named.contains("refused"),
        "the prompt lane refused: {named}"
    );
    assert_eq!(fake.sent().map(|s| s.len()), Some(1), "one line went out");
}

/// AC-7 (REQ-10): a lane whose reset instant cannot be told is skipped with its reason.
#[test]
fn a_lane_with_an_unknowable_reset_is_skipped() {
    let fake = Fake::one(limit_screen(LIMIT_ROW_UNKNOWABLE));
    let run = run(&fake, Some(TERMINAL), RESUME_TEXT, &[]);
    let message = run.error();
    assert!(
        message.contains(TERMINAL) && message.contains("reset") && message.contains("unknown"),
        "the lane is named with its unknown reset: {message}"
    );
    assert!(fake.sent().is_none(), "nothing was sent");
    assert_eq!(fake.reads(), 1, "one read decides the answer");
}

/// REQ-4's other side: the second read still shows the stale limit row, so the one line goes
/// out after the reset — the two reads agree and nothing else is sent.
#[test]
fn a_stale_limit_row_after_the_reset_still_sends_once() {
    let fake = Fake::one(limit_screen(LIMIT_ROW_PAST));
    let run = run(
        &fake,
        Some(TERMINAL),
        RESUME_TEXT,
        &["--now", "2026-11-20T10:30:00Z"],
    );
    let result = run.ok();
    assert_eq!(result["sent"], json!(true), "{}", result);
    assert_eq!(fake.reads(), 2, "the pre-send look");
    assert_eq!(fake.sent().map(|s| s.len()), Some(1));
}

/// REQ-10 on the scan path: a lane that cannot be read is skipped with its reason and the
/// others go on — the due lane still gets its one line.
#[test]
fn an_unreadable_lane_is_skipped_on_the_scan() {
    let list = json!({
        "id": "fixture", "ok": true,
        "result": {"terminals": [
            {"handle": "term_dead", "title": "dead", "branch": "main"},
            {"handle": "term_due", "title": "due", "branch": "main"}
        ]}
    })
    .to_string();
    let dir = tempfile::tempdir().unwrap();
    let bin = dir.path().join("orca");
    common::write_executable(&bin, FAKE_ORCA);
    fs::write(dir.path().join("list.json"), list).unwrap();
    // The scan reads in list order: the first read (term_dead) fails, the second (term_due)
    // shows a past reset and gets the one line.
    fs::write(dir.path().join("read-fail-1"), "").unwrap();
    fs::write(dir.path().join("read-2.json"), limit_screen(LIMIT_ROW_PAST)).unwrap();
    let fake = Fake { dir, bin };
    let run = run(&fake, None, RESUME_TEXT, &["--now", "2026-11-20T10:30:00Z"]);
    let result = run.ok();
    let named = result.to_string();
    assert!(named.contains("term_dead"), "the skipped lane: {named}");
    assert_eq!(
        fake.sent().map(|s| s.len()),
        Some(1),
        "the other lane went on"
    );
}
