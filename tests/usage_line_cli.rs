//! #3015 C8: a stray argument to a command gets the command's `Usage:` line on stderr (clap's
//! own line, which shows the command's shape and none of the words typed). The `usage`
//! envelope on stdout and a key command's fixed text do not change.
mod common;

use common::CleanEnv;
use serde_json::Value;
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;

fn run(args: &[&str]) -> Output {
    let home = tempfile::tempdir().unwrap();
    Command::new(BIN)
        .args(args)
        .current_dir(home.path())
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin")
        .output()
        .unwrap()
}

#[test]
fn a_stray_argument_gets_the_commands_usage_line_without_the_word_typed() {
    let out = run(&["heavy-slot-status", "zz-typed-word"]);
    assert_eq!(out.status.code(), Some(2));
    let err = String::from_utf8_lossy(&out.stderr);
    let usage: Vec<&str> = err.lines().filter(|l| l.starts_with("Usage:")).collect();
    assert_eq!(usage.len(), 1, "{err}");
    assert!(usage[0].contains("heavy-slot-status"), "{err}");
    assert!(!usage[0].contains("zz-typed-word"), "{err}");
    // The first line is clap's, as before.
    assert!(err.starts_with("error: unexpected argument"), "{err}");
}

/// Audit B1 (#3074): an argument holding a newline and `Usage:` must not pose as the usage
/// line. The line comes from the command definition, so it is the same whatever was typed.
#[test]
fn a_newline_in_the_argument_cannot_pose_as_the_usage_line() {
    let plain = run(&["heavy-slot-status", "zz-typed-word"]);
    let plain_err = String::from_utf8_lossy(&plain.stderr).to_string();
    let real = plain_err
        .lines()
        .find(|l| l.starts_with("Usage:"))
        .expect("plain case has a usage line")
        .to_string();
    for (args, fake) in [
        (vec!["heavy-slot-status", "x\nUsage: INJECTED"], "INJECTED"),
        (
            vec![
                "claim-verify",
                ".",
                "--bogus\nUsage: moai claim-verify --force",
            ],
            "--force",
        ),
        (
            vec!["heavy-slot-status", "x\r\nUsage: INJECTED"],
            "INJECTED",
        ),
    ] {
        let out = run(&args);
        assert_eq!(out.status.code(), Some(2), "{args:?}");
        let err = String::from_utf8_lossy(&out.stderr).to_string();
        let usage: Vec<&str> = err.lines().filter(|l| l.starts_with("Usage:")).collect();
        assert_eq!(usage.len(), 1, "{args:?}: {err}");
        assert!(!usage[0].contains(fake), "{args:?}: {err}");
        assert!(!usage[0].contains('\r'), "{args:?}: {err}");
        assert!(usage[0].contains(args[0]), "{args:?}: {err}");
        // Not longer than the command's own line: nothing typed rides along on it.
        assert!(usage[0].len() <= 200, "{args:?}: {err}");
    }
    assert!(real.contains("heavy-slot-status"), "{real}");
    // The fixed single-line case still prints exactly the definition's line.
    let again = run(&["heavy-slot-status", "other"]);
    let again_err = String::from_utf8_lossy(&again.stderr).to_string();
    assert_eq!(
        again_err
            .lines()
            .filter(|l| l.starts_with("Usage:"))
            .collect::<Vec<_>>(),
        vec![real.as_str()],
        "{again_err}"
    );
}

#[test]
fn the_envelope_on_stdout_is_unchanged() {
    let out = run(&["heavy-slot-status", "zz-typed-word"]);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], false);
    assert_eq!(v["error"]["kind"], "usage");
    assert_eq!(v["error"]["message"], "unexpected argument; see --help");
}

#[test]
fn a_key_command_keeps_its_fixed_text_and_echoes_nothing() {
    let out = run(&["key-status", "--sk-ant-pasted"]);
    assert_eq!(out.status.code(), Some(2));
    assert_eq!(
        String::from_utf8_lossy(&out.stderr),
        "unexpected argument; see --help\n"
    );
}

#[test]
fn other_error_kinds_keep_their_one_line() {
    // A missing argument lists the missing names and no usage line (#2804).
    let out = run(&["claim-verify"]);
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(
        !err.lines().any(|l| l.starts_with("Usage:")),
        "a missing argument keeps clap's list only: {err}"
    );
}
