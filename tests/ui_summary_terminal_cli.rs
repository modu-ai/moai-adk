//! The plan/apply summary on a terminal's stderr (Issue #2805, design #2799 D2 §1, §6, §11).
//!
//! stderr goes to a pseudo terminal and stdout to a pipe, so the summary branch runs as a user
//! sees it while the envelope is captured byte for byte. Every run uses a temporary HOME. Moved
//! here from the Python module `ui_summary_tty_test.py`
//! (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::{output_within, Terminal};
use common::screens::{has_row, strip_colors};
use common::CleanEnv;
use std::{
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const PACKAGE: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/skills/moai");

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = std::fs::canonicalize(tmp.path()).unwrap();
    Fx { _tmp: tmp, home }
}

fn env_of(home: &Path, extra: &[(&str, &str)]) -> Vec<(String, String)> {
    let mut env: Vec<(String, String)> = vec![
        ("PATH".into(), "/usr/bin:/bin".into()),
        ("HOME".into(), home.display().to_string()),
        ("LANG".into(), "en_US.UTF-8".into()),
        ("TERM".into(), "xterm-256color".into()),
        ("COLUMNS".into(), "100".into()),
    ];
    for (k, v) in extra {
        env.retain(|(name, _)| name != k);
        env.push((k.to_string(), v.to_string()));
    }
    env
}

struct Run {
    stdout: String,
    stderr: String,
    code: Option<i32>,
}

/// stdout is always a pipe; stderr is the terminal when `stderr_tty`, else a pipe.
fn run(cwd: &Path, home: &Path, args: &[&str], extra: &[(&str, &str)], stderr_tty: bool) -> Run {
    if !stderr_tty {
        let mut cmd = Command::new(common::MOAI_CLI_BIN);
        cmd.args(args)
            .clean_env()
            .current_dir(cwd)
            .stdin(Stdio::null());
        for (k, v) in env_of(home, extra) {
            cmd.env(k, v);
        }
        let out = output_within(&mut cmd, None);
        return Run {
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
            code: out.status.code(),
        };
    }
    let mut term = Terminal::new(common::MOAI_CLI_BIN)
        .args(args)
        .cwd(cwd)
        .stdin_not_a_terminal()
        .stdout_not_a_terminal();
    for (k, v) in env_of(home, extra) {
        term = term.env(k, v);
    }
    let s = term.run();
    Run {
        stdout: s.stdout,
        stderr: s.text,
        code: s.code,
    }
}

fn routing(f: &Fx, stderr_tty: bool, extra: &[(&str, &str)]) -> Run {
    let home = f.home.display().to_string();
    run(
        &f.home,
        &f.home,
        &["home-routing-plan", &home, PACKAGE],
        extra,
        stderr_tty,
    )
}

#[test]
fn a_terminal_gets_the_table_and_the_envelope_stays_the_same() {
    let f = fx();
    let piped = routing(&f, false, &[]);
    let tty = routing(&f, true, &[]);
    assert_eq!(tty.stdout, piped.stdout);
    assert_eq!(tty.code, piped.code);
    assert_eq!(piped.stderr, "", "no summary when stderr is a pipe");
    let plain = strip_colors(&tty.stderr);
    assert!(plain.contains("home-routing-plan"), "{plain}");
    assert!(has_row(&plain, &["ACTION", "HOST"]), "{plain}");
    assert!(has_row(&plain, &["create", "claude"]), "{plain}");
    assert!(!plain.contains("codex"), "{plain}");
    assert!(plain.contains("writes nothing"), "{plain}");
    assert!(plain.contains("next:"), "{plain}");
    assert!(plain.contains("home-routing-apply"), "{plain}");
    assert!(
        tty.stderr.contains("\u{1b}["),
        "color on a 256-color terminal"
    );
    assert!(!tty.stderr.contains("38;5;"), "{:?}", tty.stderr);
    assert!(!tty.stderr.contains("38;2;"), "{:?}", tty.stderr);
}

#[test]
fn quiet_turns_the_summary_off() {
    let f = fx();
    let r = routing(&f, true, &[("MOAI_OUTPUT", "quiet")]);
    assert_eq!(r.stderr, "");
}

#[test]
fn plain_and_dumb_terminals_get_ascii_without_escapes() {
    let f = fx();
    for extra in [
        vec![("MOAI_OUTPUT", "plain")],
        vec![("TERM", "dumb")],
        vec![("NO_COLOR", "1"), ("LANG", "C")],
    ] {
        let r = routing(&f, true, &extra);
        assert!(!r.stderr.is_empty(), "{extra:?}");
        assert!(!r.stderr.contains('\u{1b}'), "{extra:?}");
        assert!(r.stderr.is_ascii(), "{extra:?}: {:?}", r.stderr);
    }
}

#[test]
fn narrow_terminals_keep_every_line_inside_the_width() {
    let f = fx();
    let r = routing(&f, true, &[("COLUMNS", "60")]);
    let plain = strip_colors(&r.stderr);
    let lines: Vec<&str> = plain.lines().collect();
    assert!(
        lines.iter().any(|l| l.contains("--plan-digest")),
        "{lines:?}"
    );
    for line in lines {
        if line.trim_start().starts_with("--plan-digest") {
            continue; // the digest is copied whole, never cut
        }
        assert!(line.chars().count() <= 60, "{line}");
    }
}

#[test]
fn a_result_without_rows_is_one_line() {
    let f = fx();
    let home = f.home.display().to_string();
    let r = run(
        &f.home,
        &f.home,
        &[
            "home-state-plan",
            "--home",
            &home,
            "--now",
            "2026-11-13T19:00:00Z",
        ],
        &[],
        true,
    );
    assert_eq!(r.code, Some(0), "{}", r.stdout);
    let plain = strip_colors(&r.stderr);
    let lines: Vec<&str> = plain.lines().collect();
    assert_eq!(lines.len(), 4, "{lines:?}"); // title, totals, next, digest
    assert!(lines[1].contains("writes nothing"), "{lines:?}");
    assert!(lines[2].contains("home-state-apply"), "{lines:?}");
    assert!(
        lines[3].trim_start().starts_with("--plan-digest sha256:"),
        "{}",
        lines[3]
    );
}

// #2818 audit B1 (CWE-150): a HOME folder named with a window-title OSC and a clear-screen CSI;
// the summary shows them as text, the envelope escapes them as JSON.
#[test]
fn control_characters_in_a_folder_name_never_reach_the_terminal() {
    let f = fx();
    let evil = f.home.join("h\u{1b}]0;PWNED-TITLE\u{7}\u{1b}[2Jz");
    std::fs::create_dir_all(&evil).unwrap();
    let evil_arg = evil.display().to_string();
    let args = [
        "home-state-plan",
        "--home",
        evil_arg.as_str(),
        "--now",
        "2026-11-13T19:00:00Z",
    ];
    let piped = run(&f.home, &evil, &args, &[], false);
    let r = run(&f.home, &evil, &args, &[], true);
    assert_eq!(r.code, Some(0), "{}", r.stdout);
    assert_eq!(r.stdout, piped.stdout);
    assert!(
        r.stdout.contains("\\u001b]0;PWNED-TITLE"),
        "the envelope escapes it as JSON: {}",
        r.stdout
    );
    assert!(!r.stderr.contains("\u{1b}]"), "{:?}", r.stderr);
    assert!(!r.stderr.contains("\u{1b}[2J"), "{:?}", r.stderr);
    assert!(!r.stderr.contains('\u{7}'), "{:?}", r.stderr);
    assert!(
        strip_colors(&r.stderr).contains("\\x1b]0;PWNED-TITLE\\x07\\x1b[2Jz"),
        "{:?}",
        r.stderr
    );
}

#[test]
fn other_operations_print_no_summary() {
    let f = fx();
    let home = f.home.display().to_string();
    let r = run(&f.home, &f.home, &["identity", &home], &[], true);
    assert_eq!(r.code, Some(0), "{}", r.stdout);
    assert_eq!(r.stderr, "");
}
