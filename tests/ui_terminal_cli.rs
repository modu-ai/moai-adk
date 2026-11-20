//! Human output under a real terminal (Issue #2803, design #2799 D2 §3 and §11).
//!
//! Each case runs the built binary with stdout on a pseudo terminal, so the terminal branches run
//! as a user would see them: the banner's shape and color depth by `TERM`, `COLORTERM`,
//! `NO_COLOR` and `COLUMNS`, and the stdout JSON envelope staying byte-identical between a pipe
//! and a terminal. Every run uses a temporary HOME. Moved here from the Python module
//! `ui_tty_test.py` (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::Terminal;
use common::CleanEnv;
use std::{
    path::PathBuf,
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// The block letters of the logo.
const LOGO_CHAR: char = '█';

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = std::fs::canonicalize(tmp.path()).unwrap();
    Fx { _tmp: tmp, home }
}

fn env_of(fx: &Fx, extra: &[(&str, &str)]) -> Vec<(String, String)> {
    let mut env: Vec<(String, String)> = vec![
        ("PATH".into(), "/usr/bin:/bin".into()),
        ("HOME".into(), fx.home.display().to_string()),
        ("LANG".into(), "en_US.UTF-8".into()),
    ];
    for (k, v) in extra {
        env.retain(|(name, _)| name != k);
        env.push((k.to_string(), v.to_string()));
    }
    env
}

/// stdout on a terminal, stderr a pipe: `(stdout text, exit)`.
fn on_a_terminal(fx: &Fx, args: &[&str], extra: &[(&str, &str)]) -> (String, Option<i32>) {
    let mut term = Terminal::new(common::MOAI_CLI_BIN)
        .args(args)
        .cwd(&fx.home)
        .stdin_not_a_terminal()
        .stderr_not_a_terminal();
    for (k, v) in env_of(fx, extra) {
        term = term.env(k, v);
    }
    let s = term.run();
    (s.text, s.code)
}

/// All three streams pipes: `(stdout, exit)`.
fn on_a_pipe(fx: &Fx, args: &[&str], extra: &[(&str, &str)]) -> (String, Option<i32>) {
    let mut cmd = Command::new(common::MOAI_CLI_BIN);
    cmd.args(args)
        .clean_env()
        .current_dir(&fx.home)
        .stdin(Stdio::null());
    for (k, v) in env_of(fx, extra) {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    (
        String::from_utf8_lossy(&out.stdout).into_owned(),
        out.status.code(),
    )
}

fn help_on_a_terminal(fx: &Fx, extra: &[(&str, &str)]) -> String {
    let (out, code) = on_a_terminal(fx, &["--help"], extra);
    assert_eq!(code, Some(0), "{out}");
    out
}

#[test]
fn term_dumb_prints_one_ascii_line_and_no_escape() {
    let f = fx();
    let out = help_on_a_terminal(&f, &[("TERM", "dumb"), ("COLUMNS", "120")]);
    assert!(!out.contains('\u{1b}'), "{out:?}");
    assert!(!out.contains(LOGO_CHAR), "{out:?}");
    assert!(
        out.starts_with("MoAI-ADK "),
        "{:?}",
        &out[..out.len().min(80)]
    );
}

#[test]
fn a_narrow_terminal_gets_the_line_not_the_logo() {
    let f = fx();
    let narrow = help_on_a_terminal(&f, &[("TERM", "xterm-256color"), ("COLUMNS", "61")]);
    assert!(!narrow.contains(LOGO_CHAR), "{narrow:?}");
    assert!(
        narrow.starts_with("MoAI-ADK "),
        "{:?}",
        &narrow[..narrow.len().min(80)]
    );
    let wide = help_on_a_terminal(&f, &[("TERM", "xterm-256color"), ("COLUMNS", "62")]);
    assert!(wide.contains(LOGO_CHAR), "{wide:?}");
}

#[test]
fn a_256_color_terminal_gets_no_24_bit_sequence() {
    let f = fx();
    let out = help_on_a_terminal(&f, &[("TERM", "xterm-256color"), ("COLUMNS", "120")]);
    assert!(out.contains("\u{1b}[38;5;"), "{out:?}");
    assert!(!out.contains("38;2;"), "{out:?}");
}

#[test]
fn a_truecolor_terminal_keeps_the_24_bit_gradient() {
    let f = fx();
    let out = help_on_a_terminal(
        &f,
        &[
            ("TERM", "xterm-256color"),
            ("COLORTERM", "truecolor"),
            ("COLUMNS", "120"),
        ],
    );
    assert!(out.contains("\u{1b}[38;2;217;119;87m"), "{out:?}");
}

#[test]
fn a_sixteen_color_terminal_gets_the_four_bit_accent() {
    let f = fx();
    let out = help_on_a_terminal(&f, &[("TERM", "xterm"), ("COLUMNS", "120")]);
    assert!(out.contains(LOGO_CHAR), "{out:?}");
    assert!(!out.contains("38;5;"), "{out:?}");
    assert!(!out.contains("38;2;"), "{out:?}");
}

#[test]
fn no_color_keeps_the_logo_without_escapes() {
    let f = fx();
    let out = help_on_a_terminal(
        &f,
        &[
            ("TERM", "xterm-256color"),
            ("NO_COLOR", "1"),
            ("COLUMNS", "120"),
        ],
    );
    assert!(out.contains(LOGO_CHAR), "{out:?}");
    assert!(!out.contains('\u{1b}'), "{out:?}");
}

#[test]
fn a_bare_run_follows_the_same_shape() {
    let f = fx();
    let (out, code) = on_a_terminal(&f, &[], &[("TERM", "dumb"), ("COLUMNS", "120")]);
    assert_eq!(code, Some(1), "{out}");
    assert!(!out.contains("\u{1b}["), "{out:?}");
    assert!(
        out.starts_with("MoAI-ADK "),
        "{:?}",
        &out[..out.len().min(80)]
    );
}

/// The stdout envelope does not depend on whether stdout is a terminal (D2 §1, §11).
fn assert_same(f: &Fx, args: &[&str]) {
    for extra in [
        vec![("TERM", "xterm-256color"), ("COLORTERM", "truecolor")],
        vec![("TERM", "dumb")],
    ] {
        let mut env = vec![("COLUMNS", "120")];
        env.extend(extra.iter().copied());
        let (piped, pcode) = on_a_pipe(f, args, &env);
        let (shown, tcode) = on_a_terminal(f, args, &env);
        assert_eq!(shown, piped, "{args:?} {extra:?}");
        assert_eq!(tcode, pcode, "{args:?} {extra:?}");
        assert!(
            piped.starts_with("{\"protocol_version\":1,"),
            "{:?}",
            &piped[..piped.len().min(80)]
        );
    }
}

#[test]
fn the_envelope_of_a_read_only_operation_is_the_same_bytes_on_a_terminal() {
    let f = fx();
    let home = f.home.display().to_string();
    assert_same(&f, &["identity", &home]);
}

#[test]
fn the_envelope_of_a_usage_error_is_the_same_bytes_on_a_terminal() {
    let f = fx();
    assert_same(&f, &["claim-verify"]);
}
