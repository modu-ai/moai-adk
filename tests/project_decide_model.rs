//! SPEC-20261118-project-decide-model added a decision model question to the project interview;
//! SPEC-20261120-drop-decision-question removed it (user decision 2026-11-21: the decision model
//! is no longer used). Pinned here: the question bank and the PROJECT workflow ask no decision
//! model question, no skill text that asks the user anything passes a key through the question
//! tool or argv, and a profile from before the removal that still carries a `decision_model`
//! decision is still accepted. Nothing here calls an endpoint or touches a key.

mod common;
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Command, Stdio};
use tempfile::TempDir;

fn skill_text(rel: &str) -> String {
    std::fs::read_to_string(format!("{}/skills/moai/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// AC-1 (REQ-1): the question bank has no `decision_model` row and no paragraph about one.
#[test]
fn question_bank_has_no_decision_model_row_or_paragraph() {
    let bank = skill_text("references/internal/interview.md");
    let rows: Vec<&str> = bank
        .lines()
        .filter(|l| l.starts_with("| `decision_model` |"))
        .collect();
    assert!(rows.is_empty(), "the bank still has the row: {rows:?}");
    assert!(
        !bank.contains("decision_model"),
        "the bank still mentions `decision_model`"
    );
    assert!(
        !bank.to_lowercase().contains("decision model"),
        "the bank still describes a decision model"
    );
}

/// AC-1 (REQ-2): workflows/project.md has no `Decision model` step 3 item and the final report
/// table names no decision model.
#[test]
fn project_workflow_names_no_decision_model() {
    let text = skill_text("workflows/project.md");
    let items: Vec<&str> = text
        .lines()
        .filter(|l| l.trim_start().starts_with("- Decision model"))
        .collect();
    assert!(items.is_empty(), "project.md still has the item: {items:?}");
    let report = text
        .lines()
        .find(|l| l.starts_with("Final report (card C-87): "))
        .expect("project.md has the final report paragraph");
    assert!(
        !report.to_lowercase().contains("decision model"),
        "the final report table still names a decision model: {report}"
    );
    assert!(
        !text.contains("decision_model") && !text.to_lowercase().contains("decision model"),
        "project.md still mentions a decision model"
    );
}

/// The skill files of the routes that ask the user anything, named one by one (the scan reads
/// these files, not the tree): SKILL.md, the three routes that ask and their two references.
const ASKING_FILES: [&str; 6] = [
    "SKILL.md",
    "workflows/project.md",
    "workflows/setup.md",
    "workflows/update.md",
    "references/internal/interview.md",
    "references/internal/questions.md",
];

fn skill_files() -> Vec<(String, String)> {
    ASKING_FILES
        .iter()
        .map(|rel| (rel.to_string(), skill_text(rel)))
        .collect()
}

/// Where a skill text that asks the user anything shows a `moai key-set` command, it is
/// `moai key-set <provider>` with `--prompt` (and at most `--shared`), so the value is typed
/// at the hidden prompt: no value placeholder, no `--value`, no `$VAR`, no pipe and no
/// redirect. Zero such commands is allowed: the decision model step that showed one is gone.
#[test]
fn no_skill_text_passes_a_key_through_the_question_tool_or_argv() {
    for (rel, text) in skill_files() {
        for span in text.split('`').skip(1).step_by(2) {
            let Some(rest) = span.trim().strip_prefix("moai key-set") else {
                continue;
            };
            let tokens: Vec<&str> = rest.split_whitespace().collect();
            assert!(
                tokens.contains(&"--prompt"),
                "{rel}: `{span}` does not use the hidden prompt"
            );
            assert!(
                tokens
                    .iter()
                    .all(|t| ["--prompt", "--shared"].contains(t) || *t == "<provider>"),
                "{rel}: `{span}` has an argument that could carry the value: {tokens:?}"
            );
            assert!(
                !span.contains(['|', '<', '>']) || span.matches("<provider>").count() == 1,
                "{rel}: `{span}` pipes or redirects"
            );
        }
    }
}

fn plan(decision_value: &str) -> Value {
    let root = TempDir::new().unwrap();
    // The binary never sees the real HOME: it gets a temporary one of its own.
    let home = TempDir::new().unwrap();
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "r1",
        "decisions": [{"id": "decision_model", "topic": "decision_model",
            "value": decision_value, "source": "user", "scope": "project"}]
    });
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["project-plan", root.path().to_str().unwrap()])
        .env("HOME", home.path())
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

/// AC-2 (REQ-3): backward compatibility. A profile written before the question was removed
/// still carries the `decisions` entry `decision_model` (`off`, `typesafe` or `openai`), and
/// `project-plan` still accepts it.
#[test]
fn old_profile_with_a_decision_model_entry_still_plans_for_backward_compatibility() {
    for value in ["off", "typesafe", "openai"] {
        let planned = plan(value);
        assert_eq!(planned["ok"], json!(true), "{value}: {planned}");
    }
}
