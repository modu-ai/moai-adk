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
