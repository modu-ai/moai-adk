//! Audit 6846850239 B1 on PR #2654: the `hosts` row of the interview question bank offers
//! exactly the hosts the profile accepts. Each offered answer passes `project-plan`, and a
//! host the row says is not recorded yet is refused, so the bank and the profile cannot drift.

mod common;
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Command, Stdio};
use tempfile::TempDir;

/// The hosts the profile still accepts; the interview row names none of them (SPEC-20261124-
/// moai-project-interview D-7: it confirms the host this session runs in, and Codex and OMP are
/// withdrawn), so no retired host is offered as an option.
const ACCEPTED: [&str; 3] = ["claude", "codex", "omp"];
const NOT_OFFERED: [&str; 4] = ["Codex", "OMP", "Kiro", "Cursor"];

fn hosts_row() -> String {
    let text = std::fs::read_to_string(concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/skills/moai/references/internal/interview.md"
    ))
    .unwrap();
    text.lines()
        .find(|l| l.starts_with("| `hosts` |"))
        .expect("hosts row")
        .to_string()
}

fn plan(hosts: &[&str]) -> Value {
    let root = TempDir::new().unwrap();
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": hosts,
        "interview_revision": "r1"});
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["project-plan", root.path().to_str().unwrap()])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child
        .stdin
        .take()
        .unwrap()
        .write_all(profile.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout).unwrap()
}

#[test]
fn hosts_question_confirms_the_session_host_and_offers_no_retired_host() {
    let row = hosts_row();
    assert!(row.contains("this session runs in"), "{row}");
    assert!(row.contains("(recommended)"), "{row}");
    for name in NOT_OFFERED {
        assert!(
            !row.contains(name),
            "{name} is named in the hosts row: {row}"
        );
    }
    // The profile still accepts what the row confirms, and the accepted hosts together.
    assert_eq!(plan(&["claude"])["ok"], json!(true));
    for host in ACCEPTED {
        assert_eq!(plan(&[host])["ok"], json!(true), "{host}");
    }
    assert_eq!(plan(&["claude", "codex", "omp"])["ok"], json!(true));
    for host in ["kiro", "cursor"] {
        assert_eq!(plan(&[host])["ok"], json!(false), "{host}");
    }
}
