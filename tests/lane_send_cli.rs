//! SPEC-20261119-lane-send-guard (card C-120, slice B): `moai lane-send` reads a lane terminal's
//! rendered screen first and sends the text only when the lane is `working` or `idle`.
//!
//! The Orca program is a fake shell script (`MOAI_ORCA_BIN`): a `terminal read` prints the next
//! prepared screen read, a `terminal send` writes its arguments to `sent.txt` in the fake's
//! folder. "Nothing was sent" is therefore "`sent.txt` does not exist". The screens are the
//! synthetic ones in `tests/fixtures/lane_screens/`. Nothing here touches a real terminal.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, Instant};

const BIN: &str = crate::common::MOAI_BIN;
const NOW_TEXT: &str = "2026-11-19T00:00:00Z";
const TERMINAL: &str = "term_lane7";

/// The fake `orca`. `read-N.json` is the N-th read (the last one repeats), `read-fail` makes a
/// read exit 3, `send-fail` makes a send exit 4, `read-hang` makes a read sleep; `send-ok-false`,
/// `send-garbage`, `send-no-ok` and `send-empty` make a send exit 0 with an answer that is not
/// `ok: true` (an `ok: false` answer, plain text, JSON without `ok`, nothing).
const FAKE_ORCA: &str = r#"#!/bin/sh
d="$FAKE_ORCA_DIR"
case "$2" in
read)
  n=$(cat "$d/count" 2>/dev/null || echo 0)
  n=$((n+1))
  echo "$n" > "$d/count"
  if [ -f "$d/read-hang" ]; then exec sleep 6; fi
  if [ -f "$d/read-fail" ]; then echo "terminal not found" >&2; exit 3; fi
  f="$d/read-$n.json"
  [ -f "$f" ] || f=$(ls "$d"/read-*.json | sort | tail -1)
  cat "$f"
  ;;
send)
  printf '%s\n' "$*" >> "$d/sent.txt"
  if [ -f "$d/send-fail" ]; then echo "send refused" >&2; exit 4; fi
  if [ -f "$d/send-child-holds" ]; then sleep 6 & echo '{"id":"x","ok":true}'; exit 0; fi
  if [ -f "$d/send-ok-false" ]; then echo '{"id":"x","ok":false,"error":{"message":"no such terminal"}}'; exit 0; fi
  if [ -f "$d/send-garbage" ]; then echo 'sent it, thanks'; exit 0; fi
  if [ -f "$d/send-no-ok" ]; then echo '{"id":"x","result":{}}'; exit 0; fi
  if [ -f "$d/send-empty" ]; then exit 0; fi
  echo '{"id":"x","ok":true,"result":{"send":{"handle":"term_lane7","accepted":true}}}'
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
fn orca_read(text: &str, source: &str) -> String {
    let rows: Vec<&str> = text.lines().collect();
    json!({
        "id": "fixture", "ok": true,
        "result": {"terminal": {
            "handle": TERMINAL, "status": "running", "tail": rows,
            "truncated": false, "limited": false, "returnedLineCount": rows.len(),
            "source": source,
        }}
    })
    .to_string()
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

    /// A fake that reads `fixture` as a rendered screen every time.
    fn screen(name: &str) -> Fake {
        Fake::new(&[orca_read(&fixture(name), "screen")])
    }

    fn flag(&self, name: &str) {
        fs::write(self.dir.path().join(name), "").unwrap();
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
        assert_eq!(self.envelope["operation"], json!("lane.send"));
        &self.envelope["result"]
    }

    /// The refusal or failure message: a non-zero code and an error envelope.
    fn error(&self) -> &str {
        assert_ne!(self.code, 0, "expected a non-zero code: {}", self.envelope);
        assert_eq!(self.envelope["ok"], json!(false), "{}", self.envelope);
        self.envelope["error"]["message"].as_str().unwrap()
    }
}

/// `moai lane-send --terminal TERMINAL --text TEXT ARGS` with `bin` as the Orca program.
fn run_with(bin: &Path, fake_dir: &Path, terminal: &str, text: &str, extra: &[&str]) -> Run {
    let home = tempfile::tempdir().unwrap();
    let default_now: &[&str] = if extra.contains(&"--now") {
        &[]
    } else {
        &["--now", NOW_TEXT]
    };
    let out = Command::new(BIN)
        .arg("lane-send")
        .arg(format!("--terminal={terminal}"))
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

fn run(fake: &Fake, text: &str, extra: &[&str]) -> Run {
    run_with(&fake.bin, fake.dir.path(), TERMINAL, text, extra)
}

// ---- AC-1: refusals ----

#[test]
fn lane_send_refuses_a_yes_no_confirmation_and_sends_nothing() {
    // The accident of 2026-11-19: a text and Enter went into a lane showing this screen.
    for name in [
        "synthetic-permission-prompt.txt",
        "synthetic-permission-prompt-while-working.txt",
    ] {
        let fake = Fake::screen(name);
        let r = run(&fake, "do the next card", &["--enter"]);
        let message = r.error();
        assert!(message.contains("permission_prompt"), "{name}: {message}");
        assert!(message.contains("Do you want"), "{name}: {message}");
        assert_eq!(fake.sent(), None, "{name}");
    }
}

/// Audit of slice A (note 3212, B1): these three numbered Yes/No dialogs once read as `working`
/// or `idle`, so a text with Enter would have picked Yes.
#[test]
fn lane_send_refuses_a_numbered_yes_no_dialog_in_any_dress_and_sends_nothing() {
    for name in [
        "synthetic-ascii-cursor-dialog.txt",
        "synthetic-bordered-dialog.txt",
        "synthetic-no-cursor-dialog.txt",
    ] {
        let fake = Fake::screen(name);
        let r = run(&fake, "do the next card", &["--enter"]);
        let message = r.error();
        assert!(message.contains("permission_prompt"), "{name}: {message}");
        assert!(message.contains("refused"), "{name}: {message}");
        assert_eq!(fake.sent(), None, "{name}");
    }
}

/// R-N1 (leader, note 3256) as narrowed by S-N2 (note 3362): a `Do you want …` question near the
/// bottom with `Yes` and `No` rows under it that cannot be read as choices is `unknown`, so a
/// `working` or `idle` footer never lets a text reach the dialog.
#[test]
fn lane_send_refuses_a_question_whose_choices_cannot_be_read_and_sends_nothing() {
    let fake = Fake::screen("synthetic-unreadable-choice-idle.txt");
    let r = run(&fake, "do the next card", &["--enter"]);
    let message = r.error();
    assert!(message.contains("unknown"), "{message}");
    assert!(message.contains("refused"), "{message}");
    assert_eq!(fake.sent(), None);
}

/// Audit of slice B (note 3362, B1): a numbered `Yes` row with a numbered `No` row is a
/// confirmation whatever stands before the numbers. Before the repair each of these sent.
#[test]
fn lane_send_refuses_numbered_yes_no_rows_after_any_prefix_and_sends_nothing() {
    for name in [
        "synthetic-foreign-cursor-no-question-working.txt",
        "synthetic-foreign-cursor-far-question-working.txt",
        "synthetic-unreadable-choice-working.txt",
    ] {
        let fake = Fake::screen(name);
        let r = run(&fake, "do the next card", &["--enter"]);
        let message = r.error();
        assert!(message.contains("permission_prompt"), "{name}: {message}");
        assert!(message.contains("refused"), "{name}: {message}");
        assert_eq!(fake.sent(), None, "{name}");
    }
}

/// Re-audit of slice B (note 3433, R-B1 and R-N1): a known question with its choices on its own
/// row, or with two numbered rows or a choice word (`yes no allow deny y n`) below it, is
/// `unknown`, so nothing is sent. Before the repair each of these five screens sent.
#[test]
fn lane_send_refuses_a_question_with_its_choices_on_the_row_or_below_and_sends_nothing() {
    for name in [
        "synthetic-question-same-line-working.txt",
        "synthetic-fingerprint-question-working.txt",
        "synthetic-question-allow-deny-working.txt",
        "synthetic-question-allow-once-always-deny-working.txt",
        "synthetic-question-numbered-y-n-working.txt",
    ] {
        let fake = Fake::screen(name);
        let r = run(&fake, "do the next card", &["--enter"]);
        let message = r.error();
        assert!(message.contains("unknown"), "{name}: {message}");
        assert!(message.contains("refused"), "{name}: {message}");
        assert_eq!(fake.sent(), None, "{name}");
    }
}

/// S-N2 (note 3362): a lane that ended its turn with a question and waits at its input line is
/// `idle`, and the text goes in.
#[test]
fn lane_send_sends_to_a_lane_that_ended_its_turn_with_a_question() {
    let fake = Fake::screen("synthetic-question-then-input-idle.txt");
    let r = run(&fake, "yes, run the lint step", &["--enter"]);
    let result = r.ok();
    assert_eq!(result["sent"], json!(true));
    assert_eq!(result["state"], json!("idle"));
    assert_eq!(fake.sent().map(|s| s.len()), Some(1));
}

/// Audit of slice B (note 3362, B2): a send is reported as sent only when Orca's answer says
/// `ok: true`. An answer with `ok: false` (exit 0), an answer that is not JSON and an answer
/// without `ok` are failures, and the output never says `sent`.
#[test]
fn lane_send_does_not_report_a_send_whose_answer_is_not_ok_true() {
    for (flag, expect) in [
        ("send-ok-false", "ok:false"),
        ("send-garbage", "not an Orca answer"),
        ("send-no-ok", "not an Orca answer"),
        ("send-empty", "not an Orca answer"),
    ] {
        let fake = Fake::screen("synthetic-idle.txt");
        fake.flag(flag);
        let r = run(&fake, "hello", &[]);
        let message = r.error();
        assert!(message.contains("lane send failed"), "{flag}: {message}");
        assert!(message.contains(expect), "{flag}: {message}");
        assert!(!r.envelope.to_string().contains("\"sent\":true"), "{flag}");
        // The call itself happened, once: the text may have reached the lane.
        assert_eq!(fake.sent().map(|s| s.len()), Some(1), "{flag}");
    }
}

#[test]
fn lane_send_refuses_every_state_that_is_not_working_or_idle() {
    for (name, state) in [
        ("synthetic-choice-dialog.txt", "input_pending"),
        ("synthetic-usage-limit-glm.txt", "usage_limit"),
        ("synthetic-usage-limit-claude.txt", "usage_limit"),
        ("synthetic-shell.txt", "unknown"),
        ("synthetic-narrow-tail.txt", "unknown"),
    ] {
        let fake = Fake::screen(name);
        let r = run(&fake, "hello", &[]);
        let message = r.error();
        assert!(message.contains(state), "{name}: {message}");
        assert!(message.contains("refused"), "{name}: {message}");
        assert_eq!(fake.sent(), None, "{name}");
    }
}

#[test]
fn lane_send_refuses_a_read_that_is_not_the_rendered_screen() {
    let idle = fixture("synthetic-idle.txt");
    for source in ["stream", "screen-unavailable"] {
        let fake = Fake::new(&[orca_read(&idle, source)]);
        let r = run(&fake, "hello", &[]);
        assert!(r.error().contains("unknown"), "{source}: {}", r.error());
        assert_eq!(fake.sent(), None, "{source}");
    }
    // Plain text on Orca's stdout is not an Orca answer: an idle-looking screen is not trusted.
    let fake = Fake::new(&[idle]);
    let r = run(&fake, "hello", &[]);
    assert!(r.error().contains("unreachable"), "{}", r.error());
    assert_eq!(fake.sent(), None);
}

#[test]
fn lane_send_refuses_when_the_second_read_differs_from_the_first() {
    // Idle on the first read, a confirmation on the second: the screen changed in between.
    let fake = Fake::new(&[
        orca_read(&fixture("synthetic-idle.txt"), "screen"),
        orca_read(&fixture("synthetic-permission-prompt.txt"), "screen"),
    ]);
    let r = run(&fake, "hello", &["--enter"]);
    let message = r.error();
    assert!(message.contains("changed"), "{message}");
    assert!(
        message.contains("idle") && message.contains("permission_prompt"),
        "{message}"
    );
    assert_eq!(fake.sent(), None);
    assert_eq!(fake.reads(), 2);
    // Any difference counts, also working then idle.
    let fake = Fake::new(&[
        orca_read(&fixture("synthetic-working.txt"), "screen"),
        orca_read(&fixture("synthetic-idle.txt"), "screen"),
    ]);
    assert!(run(&fake, "hello", &[]).error().contains("changed"));
    assert_eq!(fake.sent(), None);
}

#[test]
fn lane_send_refuses_a_usage_limit_with_its_reset_time_in_utc() {
    // Nov 20 7pm in Asia/Seoul is 10:00 UTC; `--now` is 2026-11-19T00:00:00Z.
    let fake = Fake::screen("synthetic-usage-limit-claude.txt");
    let message = run(&fake, "hello", &[]).error().to_string();
    assert!(message.contains("2026-11-20T10:00:00Z"), "{message}");
    // The glm screen shows no time zone: the refusal says the reset time is unknown.
    let fake = Fake::screen("synthetic-usage-limit-glm.txt");
    let message = run(&fake, "hello", &[]).error().to_string();
    assert!(message.contains("unknown"), "{message}");
    assert!(!message.contains("2026-11-19T17:30:55Z"), "{message}");
    assert_eq!(fake.sent(), None);
}

// ---- AC-2: sends ----

#[test]
fn lane_send_sends_the_text_once_to_an_idle_lane() {
    let fake = Fake::screen("synthetic-idle.txt");
    let r = run(&fake, "do the next card, please", &[]);
    let result = r.ok();
    assert_eq!(result["sent"], json!(true));
    assert_eq!(result["state"], json!("idle"));
    assert_eq!(result["terminal"], json!(TERMINAL));
    assert_eq!(result["enter"], json!(false));
    // Both reads ran before the one send; the text went unchanged and without Enter.
    assert_eq!(fake.reads(), 2);
    let sent = fake.sent().expect("a send");
    assert_eq!(sent.len(), 1, "{sent:?}");
    assert!(
        sent[0].starts_with("terminal send --terminal term_lane7 --text"),
        "{sent:?}"
    );
    assert!(sent[0].contains("do the next card, please"), "{sent:?}");
    assert!(!sent[0].contains("--enter"), "{sent:?}");
}

#[test]
fn lane_send_sends_to_a_working_lane_and_enter_only_when_asked() {
    let fake = Fake::screen("synthetic-working.txt");
    let r = run(&fake, "한 줄 지시 ✓", &["--enter"]);
    let result = r.ok();
    assert_eq!(result["state"], json!("working"));
    assert_eq!(result["enter"], json!(true));
    let sent = fake.sent().expect("a send");
    assert_eq!(sent.len(), 1);
    assert!(sent[0].contains("--enter"), "{sent:?}");
    assert!(sent[0].contains("한 줄 지시 ✓"), "{sent:?}");
}

#[test]
fn lane_send_sends_failure_is_never_reported_as_sent() {
    let fake = Fake::screen("synthetic-idle.txt");
    fake.flag("send-fail");
    let r = run(&fake, "hello", &[]);
    let message = r.error().to_string();
    assert!(message.contains("send failed"), "{message}");
    assert!(!message.to_lowercase().contains("sent"), "{message}");
    // The send was attempted once and not repeated.
    assert_eq!(fake.sent().map(|s| s.len()), Some(1));
}

/// Re-audit of the scan slice (delta note 3648, N1): a send that has happened, whose `orca` then
/// leaves a child holding its output, fails as "did not finish its output" and must say the text
/// may have reached the lane, as a send that did not answer does: otherwise the reader sends
/// again and the text goes in twice.
#[test]
fn lane_send_a_send_whose_output_a_child_holds_says_the_text_may_have_reached_the_lane() {
    let fake = Fake::screen("synthetic-idle.txt");
    fake.flag("send-child-holds");
    let started = Instant::now();
    let r = run(&fake, "hello", &[]);
    let message = r.error().to_string();
    assert!(message.contains("did not finish its output"), "{message}");
    assert!(
        message.contains("the text may have reached the lane"),
        "{message}"
    );
    assert!(!message.to_lowercase().contains("sent"), "{message}");
    assert!(
        started.elapsed() < Duration::from_secs(5),
        "cut off after the 2 second limit, took {:?}",
        started.elapsed()
    );
    // The send happened once and was not repeated.
    assert_eq!(fake.sent().map(|s| s.len()), Some(1));
}

// ---- AC-3: unsafe text, unreachable lane ----

#[test]
fn lane_send_unsafe_text_is_refused_before_any_orca_call() {
    let long = "x".repeat(2001);
    for text in [
        "",
        "-x",
        "--interrupt",
        "line one\nline two",
        "with\ttab",
        "esc \u{1b}[31m red",
        "bidi \u{202e} text",
        "separator \u{2028} text",
        long.as_str(),
    ] {
        let fake = Fake::screen("synthetic-idle.txt");
        let r = run(&fake, text, &[]);
        let message = r.error();
        assert!(message.contains("refused"), "{text:?}: {message}");
        assert_eq!(fake.sent(), None, "{text:?}");
        assert_eq!(
            fake.reads(),
            0,
            "the text is checked before any read: {text:?}"
        );
        // The refused text is not echoed back.
        if !text.is_empty() && text.chars().count() < 100 {
            assert!(!message.contains(text), "{text:?}: {message}");
        }
    }
    // The longest allowed text passes the text check.
    let fake = Fake::screen("synthetic-idle.txt");
    let ok = "x".repeat(2000);
    assert_eq!(run(&fake, &ok, &[]).ok()["sent"], json!(true));
}

#[test]
fn lane_send_unsafe_terminal_handles_are_refused() {
    for handle in ["", "-x", "term lane", "term;rm", "a".repeat(129).as_str()] {
        let fake = Fake::screen("synthetic-idle.txt");
        let r = run_with(&fake.bin, fake.dir.path(), handle, "hello", &[]);
        assert!(r.error().contains("refused"), "{handle:?}: {}", r.error());
        assert_eq!(fake.reads(), 0, "{handle:?}");
        assert_eq!(fake.sent(), None, "{handle:?}");
    }
}

#[test]
fn lane_send_unsafe_an_absent_orca_is_an_unreachable_lane() {
    let dir = tempfile::tempdir().unwrap();
    let missing = dir.path().join("no-such-orca");
    let r = run_with(&missing, dir.path(), TERMINAL, "hello", &[]);
    assert!(r.error().contains("unreachable"), "{}", r.error());
    assert!(!dir.path().join("sent.txt").exists());
}

#[test]
fn lane_send_unsafe_a_failed_read_or_an_exited_terminal_is_unreachable() {
    let fake = Fake::screen("synthetic-idle.txt");
    fake.flag("read-fail");
    let r = run(&fake, "hello", &[]);
    assert!(r.error().contains("unreachable"), "{}", r.error());
    assert_eq!(fake.sent(), None);
    let exited = json!({"id": "x", "ok": true,
        "result": {"terminal": {"handle": TERMINAL, "status": "exited", "tail": [], "source": "screen"}}});
    let fake = Fake::new(&[exited.to_string()]);
    let r = run(&fake, "hello", &[]);
    assert!(r.error().contains("unreachable"), "{}", r.error());
    assert_eq!(fake.sent(), None);
}

#[test]
fn lane_send_unsafe_a_hanging_orca_is_cut_off_and_nothing_is_sent() {
    let fake = Fake::screen("synthetic-idle.txt");
    fake.flag("read-hang");
    let started = Instant::now();
    let r = run(&fake, "hello", &[]);
    assert!(r.error().contains("unreachable"), "{}", r.error());
    assert!(
        started.elapsed() < Duration::from_secs(5),
        "cut off after the 2 second limit, took {:?}",
        started.elapsed()
    );
    assert_eq!(fake.sent(), None);
}

#[test]
fn lane_send_unsafe_a_bad_now_is_a_usage_refusal() {
    let fake = Fake::screen("synthetic-idle.txt");
    let r = run(&fake, "hello", &["--now", "yesterday"]);
    assert!(r.error().contains("--now"), "{}", r.error());
    assert_eq!(fake.sent(), None);
}
