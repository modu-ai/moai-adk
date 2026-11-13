//! Audit 6846850239 B1 on PR #2654: the `hosts` row of the interview question bank offers
//! exactly the hosts the profile accepts. Each offered answer passes `project-plan`, and a
//! host the row says is not recorded yet is refused, so the bank and the profile cannot drift.

use serde_json::{json, Value};
use std::io::Write;
use std::process::{Command, Stdio};
use tempfile::TempDir;

/// The tool names the row offers and the profile value each one records.
const OFFERED: [(&str, &str); 2] = [("Claude Code", "claude"), ("Codex CLI", "codex")];
const NOT_YET: [(&str, &str); 3] = [("OMP", "omp"), ("Kiro", "kiro"), ("Cursor CLI", "cursor")];

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
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
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
fn hosts_question_offers_exactly_the_hosts_the_profile_accepts() {
    let row = hosts_row();
    let question = row.split(" | ").nth(2).unwrap();
    let offered = question.split('?').next().unwrap();
    for (name, value) in OFFERED {
        assert!(offered.contains(name), "{name} not offered: {offered}");
        assert!(row.contains(&format!("`{value}`")), "{value}");
        assert_eq!(plan(&[value])["ok"], json!(true), "{value}");
    }
    assert_eq!(plan(&["claude", "codex"])["ok"], json!(true));
    for (name, value) in NOT_YET {
        assert!(
            !offered.contains(name),
            "{name} offered but refused by the profile"
        );
        assert!(row.contains(name), "{name} is named as not recorded yet");
        assert_eq!(plan(&[value])["ok"], json!(false), "{value}");
    }
}
