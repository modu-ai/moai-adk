//! #2799 D3 decision 1 (a), card #2804: a command line whose first word names no command is a
//! usage error like any other — stdout carries the `usage` envelope, stderr and the exit code
//! stay clap's own. Only that case changes: an option, `help <word>`, a launcher flag before a
//! word and a hook-mode line keep their output.

mod common;

use serde_json::{json, Value};
use std::process::{Command, Output, Stdio};
use tempfile::TempDir;

fn run(args: &[&str]) -> Output {
    let home = TempDir::new().unwrap();
    Command::new(common::MOAI_BIN)
        .args(args)
        .env("HOME", home.path())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .unwrap()
}

fn envelope() -> Value {
    json!({
        "protocol_version": 1,
        "operation": "cli",
        "ok": false,
        "error": {"kind": "usage", "message": "unknown command; see --help", "retryable": false}
    })
}

#[test]
fn an_unknown_first_word_prints_the_usage_envelope_and_keeps_clap_stderr() {
    // A near miss keeps clap's suggestion on stderr; the envelope never echoes the word.
    for word in ["no-such-command", "projectplan", "sk-ant-pasted-by-mistake"] {
        let out = run(&[word]);
        assert_eq!(out.status.code(), Some(2), "{word}");
        let stdout = String::from_utf8(out.stdout).unwrap();
        assert_eq!(stdout.lines().count(), 1, "{word}: one envelope line");
        let v: Value = serde_json::from_str(&stdout).unwrap();
        assert_eq!(v, envelope(), "{word}");
        assert!(
            !stdout.contains(word),
            "{word}: the envelope echoes no argument"
        );
        let stderr = String::from_utf8(out.stderr).unwrap();
        assert!(
            stderr.starts_with(&format!("error: unrecognized subcommand '{word}'")),
            "{word}: {stderr}"
        );
        assert!(
            stderr.contains("For more information, try '--help'."),
            "{stderr}"
        );
    }
    let near = String::from_utf8(run(&["projectplan"]).stderr).unwrap();
    assert!(
        near.contains("tip: some similar subcommands exist"),
        "{near}"
    );
}

/// The boundary: each of these is refused by clap too, but its first word is not an unknown
/// command, so stdout stays empty as before.
#[test]
fn other_command_line_errors_keep_their_output() {
    for args in [
        vec!["help", "nosuch"],
        vec!["-l", "nosuch"],
        vec!["--bogus"],
        vec!["no-such-command", "--hook"],
    ] {
        let out = run(&args);
        assert_ne!(out.status.code(), Some(0), "{args:?}");
        assert!(
            out.stdout.is_empty(),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stdout)
        );
    }
    // A known command with a missing argument keeps its own operation's envelope.
    let out = run(&["project-plan"]);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["operation"], json!("project.plan"));
    assert_eq!(v["error"]["kind"], json!("usage"));
}
