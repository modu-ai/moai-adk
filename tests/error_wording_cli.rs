//! #2936 S-ERR (F12, F14): an unopenable `<project>` says it must be a folder path and is
//! `not_found`; a command that reads JSON from stdin says how to give it when stdin is empty,
//! and keeps "invalid typed JSON input" for a document that is there but malformed.
mod common;
use serde_json::Value;
use std::{
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn run(args: &[&str], stdin: &[u8], home: &std::path::Path) -> (Option<i32>, Value) {
    let mut child = Command::new(EXE)
        .args(args)
        .env("HOME", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(stdin);
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    let v = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.code(), v)
}

fn error_of(v: &Value) -> (&str, &str, bool) {
    let e = &v["error"];
    (
        e["kind"].as_str().unwrap(),
        e["message"].as_str().unwrap(),
        e["retryable"].as_bool().unwrap(),
    )
}

#[test]
fn an_unopenable_project_is_not_found_and_says_it_is_a_folder_path() {
    let home = TempDir::new().unwrap();
    let gone = home.path().join("no-such-project");
    // An id-looking word and an absolute path that does not exist: the same answer.
    for target in [gone.to_str().unwrap(), "my-project-id"] {
        for command in ["project-update-plan", "go-state-plan", "agents-merge-plan"] {
            let (code, v) = run(&[command, target], b"", home.path());
            assert_eq!(code, Some(1), "{command} {target}: {v}");
            let (kind, message, retryable) = error_of(&v);
            assert_eq!(kind, "not_found", "{command} {target}: {message}");
            assert!(!retryable, "{command}: {message}");
            assert!(message.contains("folder path"), "{command}: {message}");
            assert!(!message.contains("project folder unavailable"), "{message}");
        }
    }
}

#[test]
fn an_empty_stdin_says_how_to_give_the_json_and_a_broken_one_keeps_its_wording() {
    let home = TempDir::new().unwrap();
    let project = TempDir::new().unwrap();
    let p = project.path().to_str().unwrap();
    for empty in [&b""[..], b"  \n"] {
        let (code, v) = run(&["claude-settings-plan", p], empty, home.path());
        assert_eq!(code, Some(1), "{v}");
        let (kind, message, retryable) = error_of(&v);
        assert_eq!(kind, "invalid_input", "{message}");
        assert!(!retryable);
        assert!(
            message.contains("stdin") && message.contains("apply.stdin"),
            "{message}"
        );
    }
    let (code, v) = run(&["claude-settings-plan", p], b"{", home.path());
    assert_eq!(code, Some(1), "{v}");
    let (kind, message, _) = error_of(&v);
    assert_eq!(
        (kind, message),
        ("invalid_input", "invalid typed JSON input")
    );
    assert!(!project.path().join(".claude").exists(), "nothing written");
}
