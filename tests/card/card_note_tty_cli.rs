//! `card note` with standard input on a terminal: it refuses at once with a usage error instead
//! of waiting for text a person never types (SPEC-20261126-card-note-tty-stdin, from mo.ai.kr
//! addendum 8 Q).
//!
//! The cases run the built binary with stdin on a real pseudo terminal (`common::pty`), the way
//! a bare `moai card note --kind comment .` would at a prompt. Every run uses a temporary HOME
//! and a temporary project folder; no ledger exists, so nothing here can write a note.

#![cfg(unix)]

use crate::common;
use common::pty::Terminal;
use std::time::Duration;
use tempfile::TempDir;

/// The words the refusal must carry, so a person knows what went wrong.
const REFUSAL: &str = "card note refused: standard input is a terminal";

/// A run that waited for terminal input would reach this limit and the helper would fail the
/// test; the refusal answers in well under it.
const LIMIT: Duration = Duration::from_secs(20);

fn run(stdin_on_terminal: bool, args: &[&str]) -> common::pty::Screen {
    let tmp = TempDir::new().unwrap();
    let home = std::fs::canonicalize(tmp.path()).unwrap();
    let folder = home.to_str().unwrap();
    let mut term = Terminal::new(common::MOAI_CLI_BIN)
        .args(args)
        .args(["--project", folder, "--home", folder])
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", folder)
        .env("LANG", "en_US.UTF-8")
        .cwd(&home)
        .stdout_not_a_terminal()
        .stderr_not_a_terminal()
        .limit(LIMIT);
    if !stdin_on_terminal {
        term = term.stdin_text("a note from a pipe\n");
    }
    term.run()
}

/// AC-1: stdin on a terminal ends `card note` with a non-zero exit code and the usage words,
/// within the limit.
#[test]
fn a_terminal_on_stdin_is_refused_with_a_usage_error_and_no_wait() {
    let screen = run(true, &["card", "note", "C-1", "--kind", "comment"]);
    assert_ne!(screen.code, Some(0), "stderr: {}", screen.stderr);
    assert!(
        screen.stderr.contains(REFUSAL),
        "stderr: {} stdout: {}",
        screen.stderr,
        screen.stdout
    );
    assert!(
        screen.stderr.contains("< note.md"),
        "the refusal shows how to give the text: {}",
        screen.stderr
    );
}

/// AC-2: the refusal depends on stdin alone; a pipe carrying text goes past it to the ordinary
/// handling (here a project with no ledger answers its own error).
#[test]
fn a_pipe_on_stdin_is_not_refused_as_a_terminal() {
    let screen = run(false, &["card", "note", "C-1", "--kind", "comment"]);
    assert!(
        !screen.stderr.contains(REFUSAL),
        "a pipe is read, not refused: {}",
        screen.stderr
    );
}
