//! SPEC-20261118-project-decide-model: the project interview asks for the autonomous decision
//! model (not used / Jev / OpenAI) and, for a picked provider, the key — and the key never
//! reaches the session: it is typed at `moai key-set <provider> --prompt` in the user's own
//! terminal and checked with `key-status`. Nothing here calls an endpoint or touches a key.

mod common;
use serde_json::{json, Value};
use std::io::Write;
use std::process::{Command, Stdio};
use tempfile::TempDir;

fn skill_text(rel: &str) -> String {
    std::fs::read_to_string(format!("{}/skills/moai/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn bank_row(id: &str) -> String {
    skill_text("references/internal/interview.md")
        .lines()
        .find(|l| l.starts_with(&format!("| `{id}` |")))
        .unwrap_or_else(|| panic!("the question bank has no `{id}` row"))
        .to_string()
}

/// The workflow bullet that starts with `- Decision model`.
fn decision_model_step() -> String {
    let text = skill_text("workflows/project.md");
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("   - Decision model") {
            inside = true;
        } else if inside && (line.starts_with("   - ") || !line.starts_with(' ')) {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "project.md has no `- Decision model` step");
    out
}

fn position(text: &str, needle: &str) -> usize {
    text.find(needle)
        .unwrap_or_else(|| panic!("{needle:?} is missing from: {text}"))
}

/// AC-1 (card criterion 1): the row offers not used (recommended, default), Jev and OpenAI,
/// in that order, and says that a provider pick sends text to an external API.
#[test]
fn decision_model_row_offers_off_jev_and_openai_with_the_transmission_note() {
    let row = bank_row("decision_model");
    let cells: Vec<&str> = row.split(" | ").collect();
    assert_eq!(cells.len(), 5, "five cells: {row}");
    let options = cells[3];
    assert!(
        position(options, "Not used (recommended") < position(options, "Jev")
            && position(options, "Jev") < position(options, "OpenAI"),
        "options order: {options}"
    );
    assert!(
        options.contains("default"),
        "the default is named: {options}"
    );
    let question = cells[2];
    assert!(
        question.contains("sends the text of those decision requests to an external API"),
        "the external transmission is stated in the question: {question}"
    );
    assert!(
        cells[4].starts_with("decisions"),
        "goes to decisions: {row}"
    );
}

/// AC-2 (card criterion 3): the key question is asked for the picked provider only; the
/// existing key is offered only when `key-status` reads present; a new key is typed by the
/// user at the hidden prompt of their own terminal and checked with `key-status`.
#[test]
fn project_workflow_asks_existing_or_new_key_for_the_picked_provider_only() {
    let step = decision_model_step();
    for text in [
        "`decision_model`",
        "only for the provider that was picked",
        "use the existing key",
        "enter a new key",
        "only when `moai key-status <provider>` reads present",
        "`moai key-set <provider> --prompt`",
        "in their own terminal",
        "`moai key-status <provider>`",
        "Jev is `typesafe`",
        "OpenAI is `openai`",
        // REQ-8 and REQ-9: no answer, or a key that is still absent, is not a pick.
        "a cancelled, empty or unanswered question is not a pick",
        "reports the key as absent",
        "`decision_model` stays `off`",
    ] {
        assert!(step.contains(text), "the step lacks {text:?}: {step}");
    }
    // The existing key is the first option only when it is present; it is not listed otherwise.
    assert!(
        position(&step, "use the existing key") < position(&step, "enter a new key"),
        "existing key first: {step}"
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

/// AC-3 (card criterion 4): where a skill text shows a `moai key-set` command it is
/// `moai key-set <provider>` with `--prompt` (and at most `--shared`), so the value is typed
/// at the hidden prompt: no value placeholder, no `--value`, no `$VAR`, no pipe and no
/// redirect. The workflow also says the key never goes through the question tool.
#[test]
fn no_skill_text_passes_a_key_through_the_question_tool_or_argv() {
    let mut seen = 0;
    for (rel, text) in skill_files() {
        for span in text.split('`').skip(1).step_by(2) {
            let Some(rest) = span.trim().strip_prefix("moai key-set") else {
                continue;
            };
            seen += 1;
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
    assert!(seen >= 1, "no `moai key-set` command in any skill text");
    let step = decision_model_step();
    for text in [
        "never goes through the question tool",
        "a command argument",
        "this session's output",
    ] {
        assert!(step.contains(text), "the step lacks {text:?}: {step}");
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

/// AC-4 (card criterion 5): the pick is recorded as the `decision_model` entry of
/// `decisions`; `off` is the default and means no network call; the profile takes the entry.
#[test]
fn decision_model_pick_is_recorded_as_a_decisions_entry_that_defaults_to_off() {
    let step = decision_model_step();
    for text in [
        "the `decisions` entry `decision_model`",
        "`off`",
        "`typesafe`",
        "`openai`",
        "makes no network call",
        "the default",
    ] {
        assert!(step.contains(text), "the step lacks {text:?}: {step}");
    }
    for value in ["off", "typesafe", "openai"] {
        let planned = plan(value);
        assert_eq!(planned["ok"], json!(true), "{value}: {planned}");
    }
}
