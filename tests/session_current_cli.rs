//! #3015 C4: `session-current` reads this session's own id from the host's environment and
//! never from the per-project record that the latest session start overwrites. The binary runs
//! with a cleared environment, an empty temporary HOME and a project folder whose record names
//! a different session.
mod common;

use serde_json::Value;
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;
const OWN: &str = "4d5550e0-88b3-40ca-8c46-07425eedc812";
const LEADER: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";

/// Run in a project whose `.moai/state/current-session-id.txt` holds the leader's id.
fn run(vars: &[(&str, &str)]) -> (Output, Value) {
    let home = tempfile::tempdir().unwrap();
    let project = tempfile::tempdir().unwrap();
    let state = project.path().join(".moai/state");
    std::fs::create_dir_all(&state).unwrap();
    std::fs::write(state.join("current-session-id.txt"), format!("{LEADER}\n")).unwrap();
    let mut cmd = Command::new(BIN);
    cmd.arg("session-current")
        .current_dir(project.path())
        .env_clear()
        .env("HOME", home.path())
        .env("PATH", std::env::var("PATH").unwrap_or_default());
    for (k, v) in vars {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out, v)
}

#[test]
fn a_lane_gets_its_own_id_and_never_the_project_records() {
    let (out, v) = run(&[("CLAUDE_CODE_SESSION_ID", OWN)]);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["operation"], "session.current");
    assert_eq!(v["result"]["source_session_id"], OWN);
    assert_eq!(v["result"]["source"], "env:CLAUDE_CODE_SESSION_ID");
    assert!(!String::from_utf8_lossy(&out.stdout).contains(LEADER));
}

#[test]
fn without_a_host_variable_the_project_record_is_not_used_as_the_answer() {
    let (out, v) = run(&[]);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert!(v["result"]["source_session_id"].is_null(), "{v}");
    assert!(
        v["result"]["fallback"]
            .as_str()
            .unwrap()
            .starts_with("source_session_id: <not-available"),
        "{v}"
    );
    assert!(!String::from_utf8_lossy(&out.stdout).contains(LEADER));
}

#[test]
fn a_codex_thread_id_is_used_when_the_claude_variable_is_absent() {
    let (_, v) = run(&[("CODEX_THREAD_ID", OWN)]);
    assert_eq!(v["result"]["source"], "env:CODEX_THREAD_ID");
    assert_eq!(v["result"]["source_session_id"], OWN);
}

#[test]
fn a_malformed_value_is_named_and_not_echoed() {
    let (_, v) = run(&[("CLAUDE_CODE_SESSION_ID", "$(touch /tmp/x)")]);
    assert!(v["result"]["source_session_id"].is_null());
    assert_eq!(
        v["result"]["ignored"],
        serde_json::json!(["CLAUDE_CODE_SESSION_ID"])
    );
    assert!(!v.to_string().contains("touch"));
}

#[test]
fn the_command_writes_nothing_into_the_project() {
    let home = tempfile::tempdir().unwrap();
    let project = tempfile::tempdir().unwrap();
    let out = Command::new(BIN)
        .arg("session-current")
        .current_dir(project.path())
        .env_clear()
        .env("HOME", home.path())
        .env("CLAUDE_CODE_SESSION_ID", OWN)
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(0));
    assert_eq!(std::fs::read_dir(project.path()).unwrap().count(), 0);
    assert_eq!(std::fs::read_dir(home.path()).unwrap().count(), 0);
}
