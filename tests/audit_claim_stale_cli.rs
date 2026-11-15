//! #3015 C7b: `audit-claim-stale` refuses a bad `--repo`, `--since-hours` or `--now` before it
//! asks GitHub for anything, and appears in the help. The listing is pinned in the unit tests
//! with a recorded forge; the binary runs with a cleared environment and an empty temporary HOME.
mod common;

use common::CleanEnv;
use serde_json::Value;
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;

fn run(args: &[&str]) -> (Output, Value) {
    let home = tempfile::tempdir().unwrap();
    let out = Command::new(BIN)
        .arg("audit-claim-stale")
        .args(args)
        .current_dir(home.path())
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out, v)
}

#[test]
fn a_repository_that_is_not_owner_slash_repo_is_refused() {
    for repo in ["acme", "acme/widgets/extra", "/widgets", "acme/"] {
        let (out, v) = run(&["--repo", repo]);
        assert_ne!(out.status.code(), Some(0), "{repo}: {v}");
        assert!(v.to_string().contains("expected OWNER/REPO"), "{repo}: {v}");
    }
}

#[test]
fn a_look_back_outside_one_to_720_hours_is_refused() {
    for hours in ["0", "721", "-5", "x"] {
        let (out, v) = run(&["--repo", "acme/widgets", "--since-hours", hours]);
        assert_ne!(out.status.code(), Some(0), "{hours}: {v}");
    }
}

#[test]
fn a_time_that_is_not_utc_text_is_refused() {
    let (out, v) = run(&["--repo", "acme/widgets", "--now", "tomorrow"]);
    assert_ne!(out.status.code(), Some(0), "{v}");
    assert!(v.to_string().contains("invalid --now"), "{v}");
}

#[test]
fn the_command_is_listed_in_the_help() {
    let out = Command::new(BIN).args(["help", "--all"]).output().unwrap();
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("audit-claim-stale"), "{text}");
}
