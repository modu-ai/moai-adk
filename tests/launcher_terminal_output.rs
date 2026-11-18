//! #3004 P9: after `moai cc|glm|codex -l|-a` the host exits and the launcher used to print one
//! JSON line (`executed:true`) even on a terminal. On a terminal without `--json` it now prints
//! nothing after a normal exit, a line or two on stderr after a failed one, and a readable
//! summary for `--dry-run`; a pipe, a file and `--json` keep the envelope byte for byte.
//!
//! A stub `claude` on PATH stands in for the host, in a temporary git project with a temporary
//! HOME; no real host starts and no real key is read. The terminal tests run the launcher under
//! BSD `script` (a pseudo terminal, the way `tests/update_command_cli.rs` does), so they are
//! macOS only; the decision itself is unit tested for Linux in `src/launcher_terminal.rs`.
#![cfg(unix)]
mod common;
use std::collections::BTreeSet;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    process::{Command, Output, Stdio},
};

/// A value that stands for the z.ai key; it must never reach any output of the launcher.
const SECRET: &str = "moai-test-secret-token-do-not-print";
/// What the stub host prints, so the test can tell the host's output from the launcher's.
const MARK: &str = "STUB-HOST-OUTPUT";

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
}

impl Fixture {
    fn new() -> Self {
        let project = tempfile::tempdir().unwrap();
        let ok = common::git::command()
            .args(["init", "-q"])
            .current_dir(project.path())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .unwrap()
            .success();
        assert!(ok);
        let bin = tempfile::tempdir().unwrap();
        let stub = bin.path().join("claude");
        fs::write(
            &stub,
            format!("#!/bin/sh\necho {MARK}\nexit \"${{STUB_EXIT:-0}}\"\n"),
        )
        .unwrap();
        fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
        Fixture { project, bin }
    }

    fn command(&self, program: &str, lead: &[&str], args: &[&str], stub_exit: i32) -> Command {
        let mut c = Command::new(program);
        c.args(lead)
            .args(args)
            .current_dir(self.project.path())
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.bin.path().display()),
            )
            .env("HOME", self.project.path().join("no-such-home"))
            .env("ANTHROPIC_AUTH_TOKEN", SECRET)
            .env("STUB_EXIT", stub_exit.to_string())
            .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR)
            .stdin(Stdio::null());
        c
    }

    /// Stdout is a pipe, so the launcher sees a non-terminal.
    fn pipe(&self, args: &[&str], stub_exit: i32) -> Output {
        self.command(common::MOAI_BIN, &[], args, stub_exit)
            .output()
            .unwrap()
    }

    /// Stdout is a pseudo terminal (BSD `script`): the launcher's and the host's output both
    /// arrive here, in order, with `\r\n` line ends turned back into `\n`. Returns the exit
    /// status `script` reports (the launcher's) and the text.
    #[cfg(target_os = "macos")]
    fn tty(&self, args: &[&str], stub_exit: i32) -> (Option<i32>, String) {
        let out = self
            .command(
                "/usr/bin/script",
                &["-q", "/dev/null", common::MOAI_BIN],
                args,
                stub_exit,
            )
            .stdout(Stdio::piped())
            .stderr(Stdio::null())
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout).replace("\r\n", "\n");
        // `script` echoes its own end-of-input character ("^D" and two backspaces) first.
        let text = text
            .strip_prefix("^D\u{8}\u{8}")
            .unwrap_or(&text)
            .to_string();
        (out.status.code(), text)
    }
}

/// The key set of the envelope as the launcher printed it before this change.
fn envelope_keys(with_key_source: bool) -> BTreeSet<String> {
    let mut keys: BTreeSet<String> = [
        "host",
        "role",
        "name",
        "argv",
        "env_keys",
        "notes",
        "host_loaded",
        "executed",
        "exit_code",
    ]
    .iter()
    .map(|k| k.to_string())
    .collect();
    if with_key_source {
        keys.insert("key_source".into());
    }
    keys
}

fn keys_of(v: &serde_json::Value) -> BTreeSet<String> {
    v.as_object().unwrap().keys().cloned().collect()
}

/// The last line of `text` that is a JSON object, parsed.
fn last_envelope(text: &str) -> serde_json::Value {
    let line = text
        .lines()
        .rev()
        .find(|l| l.trim_start().starts_with('{'))
        .unwrap_or_else(|| panic!("no envelope line in:\n{text}"));
    serde_json::from_str(line.trim()).unwrap()
}

#[cfg(target_os = "macos")]
fn has_json_line(text: &str) -> bool {
    text.lines().any(|l| l.trim_start().starts_with('{'))
}

// ---- a pipe or a file: the envelope, as before ----

#[test]
fn a_pipe_gets_the_envelope_after_a_normal_exit() {
    let f = Fixture::new();
    let out = f.pipe(&["cc", "-a"], 0);
    assert_eq!(out.status.code(), Some(0));
    let stdout = String::from_utf8_lossy(&out.stdout);
    let lines: Vec<&str> = stdout.lines().collect();
    assert_eq!(lines.len(), 2, "{stdout}");
    assert_eq!(lines[0], MARK);
    let v: serde_json::Value = serde_json::from_str(lines[1]).unwrap();
    assert_eq!(keys_of(&v), envelope_keys(false), "{v}");
    assert_eq!(v["executed"], serde_json::json!(true));
    assert_eq!(v["exit_code"], serde_json::json!(0));
    assert_eq!(v["host_loaded"], serde_json::json!(false));
    assert_eq!(v["role"], "agent");
    assert_eq!(v["name"], "agent-1");
    assert!(v["notes"].as_array().unwrap().len() >= 2, "{v}");
}

#[test]
fn a_pipe_gets_the_envelope_with_the_host_exit_code_and_the_process_exits_as_before() {
    let f = Fixture::new();
    let out = f.pipe(&["cc", "-a"], 3);
    // The launcher has always exited 0 once the host ran; the host's code is in the envelope.
    assert_eq!(out.status.code(), Some(0));
    let v = last_envelope(&String::from_utf8_lossy(&out.stdout));
    assert_eq!(v["executed"], serde_json::json!(true));
    assert_eq!(v["exit_code"], serde_json::json!(3));
}

#[test]
fn a_pipe_dry_run_and_the_glm_key_source_keep_their_shape() {
    let f = Fixture::new();
    let out = f.pipe(&["glm", "-a", "--dry-run"], 0);
    assert_eq!(out.status.code(), Some(0));
    let stdout = String::from_utf8_lossy(&out.stdout);
    assert!(!stdout.contains(SECRET), "{stdout}");
    let v: serde_json::Value = serde_json::from_str(stdout.trim()).unwrap();
    let mut keys = envelope_keys(true);
    keys.remove("exit_code");
    assert_eq!(keys_of(&v), keys, "{v}");
    assert_eq!(v["executed"], serde_json::json!(false));
    assert!(v["env_keys"]
        .as_array()
        .unwrap()
        .contains(&serde_json::json!("ANTHROPIC_AUTH_TOKEN")));
    let ran = f.pipe(&["glm", "-a"], 0);
    let stdout = String::from_utf8_lossy(&ran.stdout);
    assert!(!stdout.contains(SECRET), "{stdout}");
    let v = last_envelope(&stdout);
    assert_eq!(keys_of(&v), envelope_keys(true), "{v}");
    assert_eq!(v["key_source"], "env:ANTHROPIC_AUTH_TOKEN");
}

#[test]
fn json_is_accepted_on_a_pipe_and_changes_nothing() {
    let f = Fixture::new();
    let plain = f.pipe(&["cc", "-a"], 0);
    let json = f.pipe(&["cc", "--json", "-a"], 0);
    assert_eq!(json.status.code(), Some(0));
    // A second launch in the same project takes the next number; compare the shape.
    let a = last_envelope(&String::from_utf8_lossy(&plain.stdout));
    let b = last_envelope(&String::from_utf8_lossy(&json.stdout));
    assert_eq!(keys_of(&a), keys_of(&b));
    assert_eq!(b["executed"], serde_json::json!(true));
}

// ---- a terminal: plain text, no envelope ----

#[cfg(target_os = "macos")]
#[test]
fn a_terminal_normal_exit_prints_nothing_after_the_host() {
    let f = Fixture::new();
    let (code, out) = f.tty(&["cc", "-a"], 0);
    assert_eq!(code, Some(0), "{out}");
    assert!(!has_json_line(&out), "{out}");
    assert!(!out.contains(SECRET), "{out}");
    let at = out
        .find(MARK)
        .unwrap_or_else(|| panic!("no host output in:\n{out}"));
    assert_eq!(out[at + MARK.len()..].trim(), "", "after the host: {out:?}");
    // What the user needs to know is said once, before the host starts.
    let before = &out[..at];
    assert_eq!(before.matches("agent number 1 taken").count(), 1, "{out}");
    assert_eq!(
        before.matches(moai_cli::session::AGENT_ONE_WARNING).count(),
        1,
        "{out}"
    );
    assert_eq!(before.matches("moai: initialized").count(), 1, "{out}");
}

#[cfg(target_os = "macos")]
#[test]
fn a_terminal_failed_exit_prints_one_or_two_plain_lines_and_keeps_the_exit_code() {
    let f = Fixture::new();
    let (code, out) = f.tty(&["cc", "-a"], 3);
    // The process exit code is what it was before (and what a pipe gives): 0 once the host ran.
    assert_eq!(code, f.pipe(&["cc", "-a"], 3).status.code(), "{out}");
    assert!(!has_json_line(&out), "{out}");
    let at = out
        .find(MARK)
        .unwrap_or_else(|| panic!("no host output in:\n{out}"));
    let after: Vec<&str> = out[at + MARK.len()..]
        .lines()
        .filter(|l| !l.trim().is_empty())
        .collect();
    assert!((1..=2).contains(&after.len()), "after the host: {after:?}");
    let after = after.join("\n");
    assert!(after.contains("claude exited with code 3"), "{after}");
    assert!(after.contains("`moai cc -a --dry-run`"), "{after}");
    assert!(!out.contains(SECRET), "{out}");
}

#[cfg(target_os = "macos")]
#[test]
fn a_terminal_dry_run_is_a_readable_summary() {
    let f = Fixture::new();
    let (code, out) = f.tty(&["glm", "-a", "--dry-run"], 0);
    assert_eq!(code, Some(0), "{out}");
    assert!(!has_json_line(&out) && !out.contains('"'), "{out}");
    for want in [
        "nothing was started",
        "glm",
        "agent-1",
        "claude --name agent-1",
        "ANTHROPIC_AUTH_TOKEN",
        "env:ANTHROPIC_AUTH_TOKEN",
        "model aliases map to z.ai models",
    ] {
        assert!(out.contains(want), "{want}\n{out}");
    }
    assert!(!out.contains(SECRET), "{out}");
    assert!(!out.contains(MARK), "the host must not start:\n{out}");
}

#[cfg(target_os = "macos")]
#[test]
fn json_on_a_terminal_brings_the_envelope_back() {
    let f = Fixture::new();
    let (code, out) = f.tty(&["cc", "--json", "-a"], 3);
    assert_eq!(code, Some(0), "{out}");
    let v = last_envelope(&out);
    assert_eq!(keys_of(&v), envelope_keys(false), "{v}");
    assert_eq!(v["executed"], serde_json::json!(true));
    assert_eq!(v["exit_code"], serde_json::json!(3));
    assert!(!out.contains("exited with code"), "{out}");
    let (_, dry) = f.tty(&["glm", "--json", "-a", "--dry-run"], 0);
    assert_eq!(last_envelope(&dry)["executed"], serde_json::json!(false));
}

#[cfg(target_os = "macos")]
#[test]
fn the_key_never_appears_on_a_terminal_in_any_form() {
    let f = Fixture::new();
    for args in [
        &["glm", "-a"][..],
        &["glm", "-a", "--dry-run"],
        &["glm", "--json", "-a"],
        &["glm", "-a", "--json", "--dry-run"],
    ] {
        for exit in [0, 3] {
            let (_, out) = f.tty(args, exit);
            assert!(!out.contains(SECRET), "{args:?} exit {exit}:\n{out}");
        }
    }
}

#[cfg(target_os = "macos")]
#[test]
fn an_option_the_launcher_drops_is_said_once_before_the_host_starts() {
    let f = Fixture::new();
    let (_, out) = f.tty(&["glm", "-a", "--advisor", "fable"], 0);
    let at = out
        .find(MARK)
        .unwrap_or_else(|| panic!("no host output in:\n{out}"));
    assert_eq!(
        out[..at]
            .matches(moai_cli::session::NOTE_ADVISOR_OMITTED_GLM)
            .count(),
        1,
        "{out}"
    );
    assert_eq!(
        out.matches(moai_cli::session::NOTE_ADVISOR_OMITTED_GLM)
            .count(),
        1,
        "{out}"
    );
    // Informational notes stay out of a normal launch.
    assert!(!out.contains("model aliases map"), "{out}");
}

/// OMP takes no advisor and no `ultracode` effort: both are said once on the terminal before the
/// host starts (Codex review on #3211).
#[cfg(target_os = "macos")]
#[test]
fn an_omp_option_the_launcher_drops_is_said_once_before_the_host_starts() {
    let f = Fixture::new();
    let stub = f.bin.path().join("omp");
    fs::write(&stub, format!("#!/bin/sh\necho {MARK}\nexit 0\n")).unwrap();
    fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
    let (_, out) = f.tty(
        &["omp", "-a", "--effort", "ultracode", "--advisor", "fable"],
        0,
    );
    let at = out
        .find(MARK)
        .unwrap_or_else(|| panic!("no host output in:\n{out}"));
    for note in [
        moai_cli::session::NOTE_EFFORT_NOT_PASSED_OMP,
        moai_cli::session::NOTE_ADVISOR_NOT_PASSED_OMP,
    ] {
        assert_eq!(out[..at].matches(note).count(), 1, "{note}\n{out}");
        assert_eq!(out.matches(note).count(), 1, "{note}\n{out}");
    }
}
