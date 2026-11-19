//! `lane-send` (card C-120 slice B, SPEC-20261119-lane-send-guard): the guard in front of
//! `orca terminal send`.
//!
//! On 2026-11-19 a text and Enter went into a lane whose screen showed a Yes/No confirmation,
//! and the highlighted `Yes` ran `git reset --hard`. `orca terminal send` types into whatever
//! terminal it is given without looking. This module looks first: it reads the terminal's
//! rendered screen, classifies it with [`crate::lane_status::lane_screen`] and hands the text on
//! only when the lane is `working` or `idle`. The screen is read once more just before the send;
//! a different state sends nothing. The text must also pass the input-safety rule of
//! `lane-duty`, plus three rules of its own (not empty, not starting with `-`, at most
//! [`MAX_TEXT_CHARS`] characters).
//!
//! The transport is the [`Orca`] trait: [`ProcessOrca`] runs the `orca` program with a time
//! limit per call, and the unit tests use a fake.

use crate::forge::policy::format_utc;
use crate::lane_status::lane_screen::{self, Reset, Screen, Seen, State};
use serde_json::{json, Value};
use std::io::Read;
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// Longest text, in characters.
pub const MAX_TEXT_CHARS: usize = 2000;
/// Longest terminal handle, in characters.
pub const MAX_HANDLE_CHARS: usize = 128;
/// Seconds one `orca` call may take.
pub const DEFAULT_TIMEOUT_SECS: u64 = 15;
/// Most bytes read from one `orca` output stream.
const OUTPUT_LIMIT: u64 = 1 << 20;

/// The text to send is refused before Orca is called. The text itself is never echoed.
pub fn check_text(text: &str) -> Result<(), String> {
    let refuse = |why: &str| Err(format!("lane send refused: the text {why}"));
    if text.is_empty() {
        return refuse("is empty");
    }
    if text.starts_with('-') {
        return refuse("begins with `-`, which Orca could read as an option");
    }
    if text.chars().count() > MAX_TEXT_CHARS {
        return refuse(&format!("is longer than {MAX_TEXT_CHARS} characters"));
    }
    if text.chars().any(crate::lane_duty::unsafe_char) {
        return refuse(
            "carries a control, line-separator or bidirectional character (it is typed into another terminal)",
        );
    }
    Ok(())
}

/// A terminal handle is a short word of letters, digits, `_`, `-` and `.`, starting with a letter
/// or a digit.
pub fn check_handle(handle: &str) -> Result<(), String> {
    let valid = !handle.is_empty()
        && handle.chars().count() <= MAX_HANDLE_CHARS
        && handle.starts_with(|c: char| c.is_ascii_alphanumeric())
        && handle
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '_' | '-' | '.'));
    if valid {
        Ok(())
    } else {
        Err(format!(
            "lane send refused: the terminal handle is invalid: use 1 to {MAX_HANDLE_CHARS} of A-Z a-z 0-9 . _ -, starting with a letter or digit"
        ))
    }
}

/// The two calls the guard makes to Orca. An `Err` carries one line of detail.
pub trait Orca {
    /// `orca terminal read --terminal <handle> --screen --json`: its standard output.
    fn read(&self, terminal: &str) -> Result<String, String>;
    /// `orca terminal send --terminal <handle> --text <text> [--enter] --json`: its standard
    /// output.
    fn send(&self, terminal: &str, text: &str, enter: bool) -> Result<String, String>;
    /// `orca terminal list --json`: its standard output (`lane-screen --scan`, slice C).
    fn list(&self) -> Result<String, String>;
}

/// Reads the screen of `terminal` for a scan: the handle is checked first, so a handle Orca
/// could read as an option never reaches the program. The detail has no `lane send` wording.
pub fn read_checked(orca: &dyn Orca, terminal: &str) -> Result<String, String> {
    check_handle(terminal).map_err(|d| d.trim_start_matches("lane send refused: ").to_string())?;
    orca.read(terminal)
}

fn allowed(state: State) -> bool {
    matches!(state, State::Working | State::Idle)
}

/// Reads and classifies the terminal's screen. Anything that is not an Orca answer is an
/// unreachable lane: plain text is never taken for a screen here.
fn look(orca: &dyn Orca, terminal: &str) -> Result<(Screen, Seen), String> {
    let out = orca
        .read(terminal)
        .map_err(|d| format!("lane send refused: lane unreachable: {d}"))?;
    let (screen, seen) = lane_screen::observe(&out);
    if screen.source != "orca:screen" {
        return Err(
            "lane send refused: lane unreachable: the answer of `orca terminal read` is not an Orca answer"
                .into(),
        );
    }
    Ok((screen, seen))
}

/// The refusal for a screen that does not allow a send: the state, the reason, the one screen
/// row it was read from and, for a usage limit, the reset time in UTC.
fn refusal(terminal: &str, screen: &Screen, seen: &Seen, now: i64) -> String {
    let mut message = format!(
        "lane send refused: the screen of {terminal} is {} ({})",
        seen.state.as_str(),
        seen.reason
    );
    if let Some(row) = &seen.evidence {
        message.push_str(&format!("; row: {row}"));
    }
    if seen.state == State::UsageLimit {
        match lane_screen::reset_of(&screen.rows, now) {
            Some(Reset::Utc(t)) => {
                message.push_str(&format!("; the limit resets at {}", format_utc(t)))
            }
            Some(Reset::Unknown { why, .. }) => {
                message.push_str(&format!("; the reset time is unknown ({why})"))
            }
            None => message.push_str("; the reset time is unknown (no reset row)"),
        }
    }
    message
}

/// Sends `text` to `terminal` when its screen is `working` or `idle` both before and just before
/// the send. `now` is Unix seconds, for the reset time of a usage limit.
pub fn send(
    orca: &dyn Orca,
    terminal: &str,
    text: &str,
    enter: bool,
    now: i64,
) -> Result<Value, String> {
    check_handle(terminal)?;
    check_text(text)?;
    let (screen, seen) = look(orca, terminal)?;
    if !allowed(seen.state) {
        return Err(refusal(terminal, &screen, &seen, now));
    }
    let (again, seen_again) = look(orca, terminal)?;
    if seen_again.state != seen.state {
        let mut message = format!(
            "lane send refused: the screen of {terminal} changed between two reads (first {}, then {})",
            seen.state.as_str(),
            seen_again.state.as_str()
        );
        if !allowed(seen_again.state) {
            // The newer screen is what the lane shows now: say why it does not allow a send.
            let now_shows = refusal(terminal, &again, &seen_again, now);
            message.push_str(&format!(
                "; {}",
                now_shows.trim_start_matches("lane send refused: ")
            ));
        }
        return Err(message);
    }
    if !allowed(seen_again.state) {
        return Err(refusal(terminal, &again, &seen_again, now));
    }
    let answer = orca
        .send(terminal, text, enter)
        .map_err(|d| format!("lane send failed: {d}"))?;
    Ok(json!({
        "sent": true,
        "terminal": terminal,
        "state": seen.state.as_str(),
        "evidence": seen.evidence,
        "enter": enter,
        "orca": serde_json::from_str::<Value>(&answer).unwrap_or(Value::Null),
    }))
}

/// Runs the `orca` program, each call under a time limit.
pub struct ProcessOrca {
    pub program: std::ffi::OsString,
    pub timeout: Duration,
}

/// One finished call.
struct Called {
    code: Option<i32>,
    stdout: String,
    stderr: String,
}

impl ProcessOrca {
    fn call(&self, args: &[&str]) -> Result<Called, String> {
        let shown = self.program.to_string_lossy().into_owned();
        let mut child = Command::new(&self.program)
            .args(args)
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| format!("cannot run `{shown}`: {}", e.kind()))?;
        let drain = |stream: Option<Box<dyn Read + Send>>| {
            std::thread::spawn(move || {
                let mut bytes = Vec::new();
                if let Some(s) = stream {
                    let _ = s.take(OUTPUT_LIMIT).read_to_end(&mut bytes);
                }
                String::from_utf8_lossy(&bytes).into_owned()
            })
        };
        let out = drain(
            child
                .stdout
                .take()
                .map(|s| Box::new(s) as Box<dyn Read + Send>),
        );
        let err = drain(
            child
                .stderr
                .take()
                .map(|s| Box::new(s) as Box<dyn Read + Send>),
        );
        let started = Instant::now();
        let status = loop {
            match child.try_wait() {
                Ok(Some(status)) => break status,
                Ok(None) if started.elapsed() >= self.timeout => {
                    let _ = child.kill();
                    let _ = child.wait();
                    return Err(format!(
                        "`{shown}` did not answer within {} seconds",
                        self.timeout.as_secs()
                    ));
                }
                Ok(None) => std::thread::sleep(Duration::from_millis(20)),
                Err(e) => return Err(format!("cannot wait for `{shown}`: {}", e.kind())),
            }
        };
        Ok(Called {
            code: status.code(),
            stdout: out.join().unwrap_or_default(),
            stderr: err.join().unwrap_or_default(),
        })
    }
}

/// The first non-blank line of a stream, cleaned and cut, for an error message.
fn first_line(text: &str) -> String {
    let line = text
        .lines()
        .map(str::trim)
        .find(|l| !l.is_empty())
        .unwrap_or("");
    let clean: String = line.chars().filter(|c| !c.is_control()).take(160).collect();
    if clean.is_empty() {
        String::new()
    } else {
        format!(": {clean}")
    }
}

impl ProcessOrca {
    /// The `orca` program of this machine: `MOAI_ORCA_BIN` overrides the program name and
    /// `MOAI_ORCA_TIMEOUT_SECS` (1 to 300) the [`DEFAULT_TIMEOUT_SECS`] limit per call.
    pub fn from_env() -> ProcessOrca {
        let secs = std::env::var("MOAI_ORCA_TIMEOUT_SECS")
            .ok()
            .and_then(|v| v.parse::<u64>().ok())
            .filter(|s| (1..=300).contains(s))
            .unwrap_or(DEFAULT_TIMEOUT_SECS);
        ProcessOrca {
            program: std::env::var_os("MOAI_ORCA_BIN").unwrap_or_else(|| "orca".into()),
            timeout: Duration::from_secs(secs),
        }
    }
}

impl Orca for ProcessOrca {
    fn list(&self) -> Result<String, String> {
        let called = self.call(&["terminal", "list", "--json"])?;
        match called.code {
            Some(0) => Ok(called.stdout),
            code => Err(format!(
                "`orca terminal list` exited with {}{}",
                code.map_or("a signal".to_string(), |c| format!("code {c}")),
                first_line(&called.stderr)
            )),
        }
    }

    fn read(&self, terminal: &str) -> Result<String, String> {
        let called = self.call(&[
            "terminal",
            "read",
            "--terminal",
            terminal,
            "--screen",
            "--json",
        ])?;
        match called.code {
            Some(0) => Ok(called.stdout),
            code => Err(format!(
                "`orca terminal read` exited with {}{}",
                code.map_or("a signal".to_string(), |c| format!("code {c}")),
                first_line(&called.stderr)
            )),
        }
    }

    fn send(&self, terminal: &str, text: &str, enter: bool) -> Result<String, String> {
        let mut args = vec!["terminal", "send", "--terminal", terminal, "--text", text];
        if enter {
            args.push("--enter");
        }
        args.push("--json");
        let called = self.call(&args).map_err(|d| {
            if d.contains("did not answer") {
                format!("{d}; the text may have reached the lane")
            } else {
                d
            }
        })?;
        match called.code {
            Some(0) => Ok(called.stdout),
            code => Err(format!(
                "`orca terminal send` exited with {}{}",
                code.map_or("a signal".to_string(), |c| format!("code {c}")),
                first_line(&called.stderr)
            )),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::cell::RefCell;

    const NOW: i64 = 1_795_046_400;

    /// A fake Orca: the N-th read answers `reads[N-1]` (the last one repeats); sends are kept.
    struct Fake {
        reads: Vec<String>,
        read_calls: RefCell<usize>,
        sent: RefCell<Vec<(String, bool)>>,
        send_error: Option<String>,
    }

    fn fake(reads: &[String]) -> Fake {
        Fake {
            reads: reads.to_vec(),
            read_calls: RefCell::new(0),
            sent: RefCell::new(Vec::new()),
            send_error: None,
        }
    }

    impl Orca for Fake {
        fn list(&self) -> Result<String, String> {
            Err("the fake lists nothing".into())
        }
        fn read(&self, _: &str) -> Result<String, String> {
            let n = *self.read_calls.borrow();
            *self.read_calls.borrow_mut() += 1;
            Ok(self.reads[n.min(self.reads.len() - 1)].clone())
        }
        fn send(&self, _: &str, text: &str, enter: bool) -> Result<String, String> {
            self.sent.borrow_mut().push((text.to_string(), enter));
            match &self.send_error {
                Some(e) => Err(e.clone()),
                None => Ok("{\"ok\":true}".into()),
            }
        }
    }

    #[test]
    fn read_checked_never_hands_a_bad_handle_to_orca() {
        let orca = fake(&[idle()]);
        for bad in ["-x", "", "a b", "é", &"a".repeat(MAX_HANDLE_CHARS + 1)] {
            let err = read_checked(&orca, bad).unwrap_err();
            assert!(err.contains("handle"), "{bad}: {err}");
            assert!(!err.contains("lane send"), "{bad}: {err}");
        }
        assert_eq!(*orca.read_calls.borrow(), 0);
        assert!(read_checked(&orca, &"a".repeat(MAX_HANDLE_CHARS)).is_ok());
        assert_eq!(*orca.read_calls.borrow(), 1);
    }

    fn screen(rows: &[&str]) -> String {
        json!({"ok": true, "result": {"terminal": {"status": "running", "tail": rows, "source": "screen"}}})
            .to_string()
    }

    fn idle() -> String {
        screen(&[
            "a",
            "b",
            "c",
            "d",
            "❯ ",
            "  ⏵⏵ bypass permissions on (shift+tab to cycle)",
        ])
    }

    fn prompt() -> String {
        screen(&["Do you want to proceed?", " ❯ 1. Yes", "   2. No"])
    }

    #[test]
    fn an_idle_lane_gets_the_text_once_after_two_reads() {
        let f = fake(&[idle()]);
        let out = send(&f, "term_1", "hello", false, NOW).unwrap();
        assert_eq!(out["sent"], json!(true));
        assert_eq!(out["state"], json!("idle"));
        assert_eq!(*f.read_calls.borrow(), 2);
        assert_eq!(*f.sent.borrow(), vec![("hello".to_string(), false)]);
    }

    #[test]
    fn enter_goes_only_when_asked() {
        let f = fake(&[idle()]);
        send(&f, "term_1", "hello", true, NOW).unwrap();
        assert_eq!(*f.sent.borrow(), vec![("hello".to_string(), true)]);
    }

    #[test]
    fn a_confirmation_sends_nothing_and_names_its_row() {
        let f = fake(&[prompt()]);
        let e = send(&f, "term_1", "hello", true, NOW).unwrap_err();
        assert!(
            e.contains("permission_prompt") && e.contains("Do you want"),
            "{e}"
        );
        assert!(f.sent.borrow().is_empty());
        assert_eq!(*f.read_calls.borrow(), 1);
    }

    #[test]
    fn a_change_between_the_reads_sends_nothing() {
        let f = fake(&[idle(), prompt()]);
        let e = send(&f, "term_1", "hello", false, NOW).unwrap_err();
        assert!(e.contains("permission_prompt"), "{e}");
        assert!(f.sent.borrow().is_empty());
        let working = screen(&["a", "b", "c", "d", "✳ Frosting… (esc to interrupt)", "❯ "]);
        let f = fake(&[idle(), working]);
        let e = send(&f, "term_1", "hello", false, NOW).unwrap_err();
        assert!(
            e.contains("changed") && e.contains("idle") && e.contains("working"),
            "{e}"
        );
        assert!(f.sent.borrow().is_empty());
    }

    #[test]
    fn a_failed_send_is_not_reported_as_sent() {
        let mut f = fake(&[idle()]);
        f.send_error = Some("`orca terminal send` exited with code 4".into());
        let e = send(&f, "term_1", "hello", false, NOW).unwrap_err();
        assert!(e.starts_with("lane send failed"), "{e}");
        assert_eq!(f.sent.borrow().len(), 1);
    }

    #[test]
    fn plain_text_from_orca_is_not_a_screen() {
        let f = fake(&["a\nb\nc\nd\n❯\n⏵⏵ bypass permissions on (shift+tab to cycle)".into()]);
        let e = send(&f, "term_1", "hello", false, NOW).unwrap_err();
        assert!(e.contains("unreachable"), "{e}");
        assert!(f.sent.borrow().is_empty());
    }

    #[test]
    fn the_text_boundaries() {
        assert!(check_text(&"x".repeat(MAX_TEXT_CHARS)).is_ok());
        assert!(check_text(&"x".repeat(MAX_TEXT_CHARS + 1)).is_err());
        assert!(check_text("").is_err());
        assert!(check_text("-a").is_err());
        assert!(check_text("a-").is_ok());
        for bad in ["a\nb", "a\tb", "a\u{1b}b", "a\u{202e}b", "a\u{2028}b"] {
            assert!(check_text(bad).is_err(), "{bad:?}");
        }
        // A refusal never echoes the text.
        assert!(!check_text("secret\nline").unwrap_err().contains("secret"));
    }

    #[test]
    fn the_handle_boundaries() {
        assert!(check_handle("term_123").is_ok());
        assert!(check_handle("a.b-c_d").is_ok());
        assert!(check_handle(&"a".repeat(MAX_HANDLE_CHARS)).is_ok());
        assert!(check_handle(&"a".repeat(MAX_HANDLE_CHARS + 1)).is_err());
        for bad in ["", "-x", "_x", "a b", "a;b", "a/b", "한글"] {
            assert!(check_handle(bad).is_err(), "{bad:?}");
        }
    }

    #[test]
    fn an_unsafe_text_is_refused_before_any_read() {
        let f = fake(&[idle()]);
        assert!(send(&f, "term_1", "a\nb", false, NOW).is_err());
        assert_eq!(*f.read_calls.borrow(), 0);
    }

    #[test]
    fn a_usage_limit_names_its_reset_time_in_utc() {
        let claude = screen(&[
            "x",
            "y",
            "  5-hour limit reached ∙ resets Nov 20 at 7pm (Asia/Seoul)",
            " ❯ 1. Stop and wait for limit to reset",
            "   2. Upgrade your plan",
        ]);
        let f = fake(&[claude]);
        let e = send(&f, "term_1", "hello", false, NOW).unwrap_err();
        assert!(
            e.contains("usage_limit") && e.contains("2026-11-20T10:00:00Z"),
            "{e}"
        );
        let glm = screen(&[
            "x",
            "y",
            "  ⎿  Your limit will reset at 2026-11-19 17:30:55",
            "a",
            "b",
            "c",
        ]);
        let f = fake(&[glm]);
        let e = send(&f, "term_1", "hello", false, NOW).unwrap_err();
        assert!(
            e.contains("reset time is unknown") && e.contains("time zone"),
            "{e}"
        );
    }
}
