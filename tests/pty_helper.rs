//! The terminal helper of `tests/common/pty.rs`: what a command sees and what the test gets
//! back. The commands are plain `sh` scripts, so these cases run the same on Linux and macOS and
//! never start the product binary.

#![cfg(unix)]

mod common;
use common::pty::{output_within_limit, piped_with, Terminal};
use std::panic::{catch_unwind, AssertUnwindSafe};
use std::path::Path;
use std::process::Command;
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
fn stdin_text_is_read_by_the_command_and_then_ends() {
    let s = sh("read a; read b; echo \"[$a][$b]\"; if read c; then echo more; else echo ended; fi; test -t 0 || echo stdin-pipe; test -t 2 && echo stderr-terminal")
        .stdin_text("one\ntwo\n")
        .run();
    assert!(s.text.contains("[one][two]"), "{:?}", s.text);
    assert!(
        s.text.contains("ended") && !s.text.contains("more"),
        "{:?}",
        s.text
    );
    assert!(s.text.contains("stdin-pipe"), "{:?}", s.text);
    assert!(s.text.contains("stderr-terminal"), "{:?}", s.text);
}

#[test]
fn a_piped_command_reads_its_input_and_returns_both_streams() {
    let env = common::pty::env_with(&[("PATH", "/usr/bin:/bin")], &[]);
    let dir = std::env::temp_dir();
    let seen = common::pty::piped_with(
        "/bin/sh",
        &dir,
        &env,
        &["-c", "read a; echo \"got-$a\"; echo oops >&2; exit 4"],
        Some("hello\n"),
    );
    assert_eq!(seen.code, Some(4));
    assert_eq!(seen.stdout, "got-hello\n");
    assert_eq!(seen.stderr, "oops\n");
}

#[test]
#[should_panic(expected = "still running after")]
fn a_command_that_does_not_end_is_killed_with_the_screen_so_far() {
    let _ = sh("echo started; sleep 30")
        .limit(Duration::from_secs(1))
        .run();
}

/// The text a panic carried.
fn panic_text(err: Box<dyn std::any::Any + Send>) -> String {
    err.downcast_ref::<String>()
        .cloned()
        .or_else(|| err.downcast_ref::<&str>().map(|s| s.to_string()))
        .unwrap_or_default()
}

/// The pid a script wrote to `file` (it writes it before anything else the test waits for).
fn pid_in(file: &Path) -> libc::pid_t {
    for _ in 0..100 {
        if let Some(pid) = std::fs::read_to_string(file)
            .ok()
            .and_then(|t| t.trim().parse().ok())
        {
            return pid;
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    panic!("no pid in {}", file.display());
}

/// Whether `pid` is gone within a few seconds (it can take a moment to be reaped).
fn gone(pid: libc::pid_t) -> bool {
    (0..100).any(|_| {
        let ended = unsafe { libc::kill(pid, 0) } != 0;
        if !ended {
            std::thread::sleep(Duration::from_millis(50));
        }
        ended
    })
}

/// A script that starts a long-lived child of its own, notes its pid, and waits for it.
fn with_grandchild(file: &Path, then: &str) -> String {
    format!("sleep 25 & echo $! > '{}'; {then}; wait", file.display())
}

#[test]
fn the_limit_kills_the_command_and_the_children_it_started() {
    let dir = tempfile::tempdir().unwrap();
    let pidfile = dir.path().join("pid");
    let script = with_grandchild(&pidfile, "echo started");
    let err = catch_unwind(AssertUnwindSafe(|| {
        sh(&script).limit(Duration::from_secs(1)).run();
    }))
    .unwrap_err();
    assert!(panic_text(err).contains("still running after"));
    assert!(gone(pid_in(&pidfile)), "the child outlived the limit");
}

#[test]
fn a_panic_in_the_test_kills_the_command_and_its_children() {
    let dir = tempfile::tempdir().unwrap();
    let pidfile = dir.path().join("pid");
    let script = with_grandchild(&pidfile, "printf 'go on? [y/N] '");
    let err = catch_unwind(AssertUnwindSafe(|| {
        sh(&script)
            .before_answer(|| panic!("the hook failed"))
            .run();
    }))
    .unwrap_err();
    assert_eq!(panic_text(err), "the hook failed");
    assert!(gone(pid_in(&pidfile)), "the child outlived the panic");
}

#[test]
fn the_limit_counts_again_from_the_answer() {
    // 1.3 s before the question and 1.3 s after the answer: each stretch is under the 2 s limit,
    // the sum is not.
    let s = sh("sleep 1.3; printf 'go on? [y/N] '; read a; sleep 1.3; echo done-$a")
        .answer("y\n")
        .limit(Duration::from_secs(2))
        .run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    assert!(s.text.contains("done-y"), "{}", s.text);
}

#[test]
fn a_command_that_closed_the_terminal_and_runs_on_is_killed() {
    let err = catch_unwind(AssertUnwindSafe(|| {
        sh("exec 0<&- 1>&- 2>&-; sleep 25")
            .limit(Duration::from_secs(1))
            .run();
    }))
    .unwrap_err();
    assert!(panic_text(err).contains("killed"));
}

#[test]
fn a_big_input_and_a_big_output_do_not_wait_on_each_other() {
    // `cat` writes what it reads: with the input written before anything reads the output, both
    // pipes fill and nobody moves.
    let big = "x".repeat(1 << 20);
    let s = Terminal::new("/bin/cat")
        .env("PATH", "/usr/bin:/bin")
        .stdin_text(&big)
        .stdout_not_a_terminal()
        .limit(Duration::from_secs(20))
        .run();
    assert_eq!(s.stdout.len(), big.len());
    let env = common::pty::env_with(&[("PATH", "/usr/bin:/bin")], &[]);
    let seen = piped_with("/bin/cat", &std::env::temp_dir(), &env, &[], Some(&big));
    assert_eq!(seen.stdout.len(), big.len());
}

#[test]
fn a_piped_command_that_does_not_end_is_killed_with_its_children() {
    let dir = tempfile::tempdir().unwrap();
    let pidfile = dir.path().join("pid");
    let mut cmd = Command::new("/bin/sh");
    cmd.args(["-c", &with_grandchild(&pidfile, "echo started")])
        .env("PATH", "/usr/bin:/bin");
    let err = catch_unwind(AssertUnwindSafe(|| {
        output_within_limit(&mut cmd, None, Duration::from_secs(1));
    }))
    .unwrap_err();
    assert!(panic_text(err).contains("still running after"));
    assert!(gone(pid_in(&pidfile)), "the child outlived the limit");
}
