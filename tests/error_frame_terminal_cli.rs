//! The stderr error frame under a real terminal (card #2804, design #2799 D2 §4 and §11).
//!
//! Each case runs the built binary with stderr on a pseudo terminal (stdout piped), so the frame
//! is drawn as a user would see it, and again with both streams piped, as a lane or a script runs
//! it. The pipe keeps today's bytes; the stdout JSON envelope is byte-identical in both. Every
//! run uses a temporary HOME. Moved here from the Python module `error_frame_tty_test.py`
//! (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::{env_with, piped, stderr_on_terminal, Env, Seen};
use common::screens::strip_colors;
use std::path::PathBuf;
use tempfile::TempDir;

/// What a command that reads its JSON from stdin says when stdin is empty (`claim-verify .` with
/// no input). It was `invalid typed JSON input` before #2936 S-ERR (F14), which now keeps that
/// wording only for a document that is there but malformed.
const EMPTY_STDIN: &str = "typed JSON input missing: this command reads its JSON from stdin and stdin was empty; pipe the plan's apply.stdin into it";

const RED_ERROR: &str = "\u{1b}[31merror:\u{1b}[0m ";

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = std::fs::canonicalize(tmp.path()).unwrap();
    Fx { _tmp: tmp, home }
}

impl Fx {
    fn env(&self, extra: &[(&str, &str)]) -> Env {
        env_with(
            &[
                ("PATH", "/usr/bin:/bin"),
                ("HOME", self.home.to_str().unwrap()),
                ("LANG", "en_US.UTF-8"),
                ("TERM", "xterm-256color"),
            ],
            extra,
        )
    }

    fn piped(&self, args: &[&str], extra: &[(&str, &str)]) -> Seen {
        piped(common::MOAI_CLI_BIN, &self.home, &self.env(extra), args)
    }

    fn framed(&self, args: &[&str], extra: &[(&str, &str)]) -> Seen {
        stderr_on_terminal(common::MOAI_CLI_BIN, &self.home, &self.env(extra), args)
    }
}

#[test]
fn an_operation_error_is_framed_on_a_terminal_and_unchanged_on_a_pipe() {
    let f = fx();
    let args = ["claim-verify", "."]; // empty stdin: EMPTY_STDIN (#2936 S-ERR)
    let p = f.piped(&args, &[]);
    assert_eq!(p.code, Some(1));
    assert_eq!(p.stderr, format!("{EMPTY_STDIN}\n"), "today's bytes");
    let t = f.framed(&args, &[]);
    assert_eq!(t.code, Some(1));
    assert_eq!(t.stdout, p.stdout, "the envelope is byte-identical");
    assert_eq!(t.stderr, format!("{RED_ERROR}{EMPTY_STDIN}\n"));
}

#[test]
fn plain_modes_draw_the_same_frame_without_escapes() {
    let f = fx();
    for extra in [
        vec![("NO_COLOR", "1")],
        vec![("TERM", "dumb")],
        vec![("COLUMNS", "60"), ("NO_COLOR", "1")],
    ] {
        let t = f.framed(&["claim-verify", "."], &extra);
        assert_eq!(t.stderr, format!("error: {EMPTY_STDIN}\n"), "{extra:?}");
    }
}

#[test]
fn a_missing_argument_keeps_its_name_and_ends_with_the_next_step() {
    let f = fx();
    let args = ["project-plan"];
    let p = f.piped(&args, &[]);
    assert_eq!(p.code, Some(2));
    assert_eq!(
        p.stderr,
        "error: the following required arguments were not provided:\n  <PROJECT>\n"
    );
    let t = f.framed(&args, &[]);
    assert_eq!(t.code, Some(2));
    assert_eq!(t.stdout, p.stdout, "the envelope is byte-identical");
    let plain = strip_colors(&t.stderr);
    let lines: Vec<&str> = plain.lines().collect();
    assert_eq!(
        lines[0],
        "error: the following required arguments were not provided:"
    );
    assert_eq!(lines[1], "  <PROJECT>");
    let last = lines[lines.len() - 1];
    assert!(last.starts_with("  next: run `"), "{plain}");
    assert!(
        last.contains(" project-plan --help` to see its arguments"),
        "{last}"
    );
    assert!(
        t.stderr.contains("\u{1b}[36mnext:\u{1b}[0m"),
        "{:?}",
        t.stderr
    );
}

#[test]
fn a_key_command_keeps_its_fixed_text_on_a_terminal() {
    // #2393: a key command's stray argument may be a key; nothing of it is echoed.
    let f = fx();
    let t = f.framed(&["key-status", "--sk-ant-pasted"], &[]);
    assert_eq!(t.code, Some(2));
    assert_eq!(t.stderr, "unexpected argument; see --help\n");
    assert!(!t.stdout.contains("sk-ant"), "{}", t.stdout);
}

#[test]
fn the_color_flag_comes_before_the_environment() {
    // #2799 ③ (D2 order): `--color=never|always` first, then NO_COLOR, TERM=dumb, the terminal.
    let f = fx();
    let auto = f.piped(&["claim-verify", "."], &[]);
    for flag in [vec!["--color", "never"], vec!["--color=never"]] {
        let mut args = flag.clone();
        args.extend(["claim-verify", "."]);
        let t = f.framed(&args, &[]);
        assert_eq!(t.code, Some(1), "{flag:?}");
        assert_eq!(t.stdout, auto.stdout, "the envelope is byte-identical");
        assert_eq!(t.stderr, format!("error: {EMPTY_STDIN}\n"), "{flag:?}");
    }
    for extra in [vec![("NO_COLOR", "1")], vec![("TERM", "dumb")]] {
        let t = f.framed(&["--color", "always", "claim-verify", "."], &extra);
        assert_eq!(t.stderr, format!("{RED_ERROR}{EMPTY_STDIN}\n"), "{extra:?}");
    }
    let t = f.framed(
        &["--color", "auto", "claim-verify", "."],
        &[("NO_COLOR", "1")],
    );
    assert_eq!(t.stderr, format!("error: {EMPTY_STDIN}\n"));
}

#[test]
fn the_color_flag_keeps_the_pipe_bytes() {
    let f = fx();
    let want = f.piped(&["claim-verify", "."], &[]);
    for flag in ["always", "never", "auto"] {
        let got = f.piped(&["--color", flag, "claim-verify", "."], &[]);
        assert_eq!(got, want, "{flag}");
    }
}

#[test]
fn a_flag_before_the_command_keeps_the_next_step_of_that_command() {
    let f = fx();
    let t = f.framed(&["--color", "never", "project-plan"], &[]);
    assert_eq!(t.code, Some(2));
    let last = t.stderr.lines().last().unwrap();
    assert!(
        last.contains(" project-plan --help` to see its arguments"),
        "{last}"
    );
}

#[test]
fn help_all_reads_the_line_without_a_leading_color_flag() {
    // Audit #2952 B2: `help --all` is answered before parsing from position 1.
    let f = fx();
    let want = f.piped(&["help", "--all"], &[]);
    assert_eq!(want.code, Some(0));
    for flag in [vec!["--color", "never"], vec!["--color=always"]] {
        let mut args = flag.clone();
        args.extend(["help", "--all"]);
        assert_eq!(f.piped(&args, &[]), want, "{flag:?}");
    }
}

#[test]
fn a_hook_word_after_a_leading_color_flag_is_still_a_hook() {
    // Audit #2952 B3: hook mode is judged from position 1 (`hook-receipt`), so a leading
    // `--color` must not turn a hook call into an ordinary usage error.
    let f = fx();
    let want = f.piped(&["hook-receipt"], &[]);
    let got = f.piped(&["--color", "never", "hook-receipt"], &[]);
    assert_eq!(got, want);
}

#[test]
fn an_unknown_color_value_is_a_usage_error() {
    let f = fx();
    let r = f.piped(&["--color", "bogus", "claim-verify", "."], &[]);
    assert_eq!(r.code, Some(2));
    assert_eq!(r.stdout, "");
    assert!(r.stderr.contains("bogus"), "{}", r.stderr);
    for value in ["auto", "always", "never"] {
        assert!(r.stderr.contains(value), "{value}: {}", r.stderr);
    }
}
