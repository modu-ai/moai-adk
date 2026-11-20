//! The terminal helper of `tests/common/pty.rs`: what a command sees and what the test gets
//! back. The commands are plain `sh` scripts, so these cases run the same on Linux and macOS and
//! never start the product binary.

#![cfg(unix)]

mod common;
use common::pty::Terminal;
use std::time::{Duration, Instant};

fn sh(script: &str) -> Terminal {
    Terminal::new("/bin/sh")
        .args(["-c", script])
        .env("PATH", "/usr/bin:/bin")
}

#[test]
fn stdin_stdout_and_stderr_are_one_terminal() {
    let s = sh("test -t 0 && test -t 1 && test -t 2 && echo all-terminals").run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(s.text.contains("all-terminals"), "{}", s.text);
}

#[test]
fn a_command_that_asked_for_no_terminal_stdin_gets_none() {
    let s = sh("if test -t 0; then echo stdin-terminal; else echo stdin-not-terminal; fi; test -t 1 && echo stdout-terminal")
        .stdin_not_a_terminal()
        .run();
    assert!(s.text.contains("stdin-not-terminal"), "{}", s.text);
    assert!(s.text.contains("stdout-terminal"), "{}", s.text);
}

#[test]
fn the_answer_is_typed_after_the_question_not_before() {
    let s = sh("printf 'go on? [y/N] '; read a; echo got-$a")
        .answer("y\n")
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    // The echo of the typed `y` follows the question; typed first it would come before it.
    assert!(s.text.starts_with("go on? [y/N] y\n"), "{:?}", s.text);
    assert!(s.text.contains("got-y"), "{:?}", s.text);
}

#[test]
fn without_an_answer_a_question_gets_end_of_input_and_stays_on_screen() {
    let started = Instant::now();
    let s = sh("printf 'go on? [y/N] '; if read a; then echo got-$a; else echo no-input; fi").run();
    assert!(s.text.contains("go on? [y/N] "), "{:?}", s.text);
    assert!(s.text.contains("no-input"), "{:?}", s.text);
    assert!(started.elapsed() < Duration::from_secs(20));
}

#[test]
fn the_exit_code_comes_back_and_a_line_end_is_one_newline() {
    let s = sh("echo one; echo two; exit 3").run();
    assert_eq!(s.code, Some(3));
    assert_eq!(s.text, "one\ntwo\n");
}

#[test]
fn before_the_answer_the_hook_runs_while_the_question_is_shown() {
    let dir = tempfile::tempdir().unwrap();
    let marker = dir.path().join("hook-ran");
    let seen = marker.clone();
    let script = format!(
        "printf 'go on? [y/N] '; read a; if test -e '{}'; then echo hook-first; else echo hook-late; fi",
        marker.display()
    );
    let s = sh(&script)
        .answer("y\n")
        .before_answer(move || std::fs::write(&seen, "x").unwrap())
        .run();
    assert!(s.text.contains("hook-first"), "{:?}", s.text);
}

#[test]
fn the_environment_is_clean_plus_what_the_case_sets() {
    let s = sh(r#"echo "[$HOME][$CASE_VAR]""#)
        .env("CASE_VAR", "set")
        .run();
    assert!(s.text.contains("[][set]"), "{:?}", s.text);
}

#[test]
fn a_piped_stdout_is_captured_and_stderr_stays_on_the_terminal() {
    let s = sh(
        "echo out; echo err >&2; if test -t 1; then echo out-terminal; else echo out-pipe >&2; fi",
    )
    .stdout_not_a_terminal()
    .run();
    assert_eq!(s.stdout, "out\n");
    assert!(
        s.text.contains("err\n") && s.text.contains("out-pipe"),
        "{:?}",
        s.text
    );
    assert!(!s.text.contains("out\nout"), "{:?}", s.text);
    assert_eq!(s.stderr, "");
}

#[test]
fn a_piped_stderr_is_captured_and_stdout_stays_on_the_terminal() {
    let s = sh("echo out; echo err >&2; test -t 2 || echo err-pipe")
        .stderr_not_a_terminal()
        .run();
    assert_eq!(s.stderr, "err\n");
    assert!(
        s.text.contains("out\n") && s.text.contains("err-pipe"),
        "{:?}",
        s.text
    );
    assert_eq!(s.stdout, "");
}

#[test]
fn a_big_piped_output_does_not_stop_the_command() {
    let s = sh("head -c 300000 /dev/zero | tr '\\0' x")
        .stdout_not_a_terminal()
        .run();
    assert_eq!(s.code, Some(0));
    assert_eq!(s.stdout.len(), 300_000);
}

#[test]
#[should_panic(expected = "still running after")]
fn a_command_that_does_not_end_is_killed_with_the_screen_so_far() {
    let _ = sh("echo started; sleep 30")
        .limit(Duration::from_secs(1))
        .run();
}
