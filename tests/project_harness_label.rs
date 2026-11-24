//! SPEC-20261124-harness-step-label: the `generation` step of `moai project-update-plan` is shown
//! to people as Project harness (프로젝트 맞춤 하네스 in Korean) while its id, the commands, the
//! manifest path and the JSON field names stay. The two plan cases run the built binary in a
//! temporary project with an empty temporary HOME and read what it prints; the text cases read
//! the two workflow files and the two source files the identifiers live in. Nothing else runs
//! and the real HOME is never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const UPDATE: &str = "skills/moai/workflows/update.md";
const PROJECT: &str = "skills/moai/workflows/project.md";
const WORKFLOWS: [&str; 2] = [UPDATE, PROJECT];

/// The one sentence each workflow text carries, in its five parts: the id it names, the two
/// labels, the clause that translates the name, the clause that keeps the id for tools, and the
/// clause that the id is never the name said to the user.
const NAMING_SENTENCE: &str = "Call the step whose id is `generation` \"Project harness\" in \
     English and \"프로젝트 맞춤 하네스\" in Korean, and translate that name for any other \
     conversation language; the id stays `generation` for the commands and the JSON, never as \
     the name you say to the user.";
const ENGLISH_AND_KOREAN: &str =
    "\"Project harness\" in English and \"프로젝트 맞춤 하네스\" in Korean";
const TRANSLATE: &str = "translate that name for any other conversation language";
const ID_FOR_TOOLS: &str = "the id stays `generation` for the commands and the JSON";
const NEVER_THE_NAME: &str = ", never as the name you say to the user";

fn repo_file(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The first line of `text` that starts with `head`.
fn line<'a>(text: &'a str, head: &str) -> &'a str {
    text.lines()
        .find(|l| l.starts_with(head))
        .unwrap_or_else(|| panic!("no line starts with {head:?}"))
}

/// `moai project-update-plan` of an empty temporary project: its `result`.
fn plan() -> Value {
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let (home, root): (PathBuf, PathBuf) = (base.join("home"), base.join("project"));
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&root).unwrap();
    let out = Command::new(EXE)
        .args(["project-update-plan", root.to_str().unwrap()])
        .clean_env()
        .env("HOME", &home)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "no envelope: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn phase<'a>(plan: &'a Value, name: &str) -> &'a Value {
    plan["phases"]
        .as_array()
        .expect("phases")
        .iter()
        .find(|p| p["phase"] == name)
        .unwrap_or_else(|| panic!("no {name} phase: {plan}"))
}

#[test]
fn install_reason_names_the_step_project_harness() {
    let reason = phase(&plan(), "install")["reason"]
        .as_str()
        .unwrap()
        .to_string();
    let label = "profile, project harness (CLAUDE.md managed section and generated files), ";
    assert!(reason.starts_with(label), "the install reason: {reason}");
    assert!(
        !reason.contains("generation"),
        "the install reason still names the step by its id: {reason}"
    );
    let rest = reason.trim_start_matches(label);
    assert!(
        rest.starts_with("Claude Code settings, Claude Code hooks, GitHub templates;"),
        "the other install steps keep their order: {reason}"
    );
}

#[test]
fn identifiers_stay_as_they_were() {
    // The plan still lists the step under its id.
    let plan = plan();
    let install = phase(&plan, "install");
    assert!(
        install["steps"]
            .as_array()
            .expect("steps")
            .iter()
            .any(|s| s["step"] == "generation"),
        "{install}"
    );
    // The source still prints the apply command and reads the manifest path; the help still
    // lists both commands.
    for (rel, needles) in [
        (
            "src/project_update.rs",
            [
                "name: \"generation\"",
                "argv(\"generation-apply\"",
                "manifests/generation.json",
            ]
            .as_slice(),
        ),
        (
            "src/help.rs",
            ["(\"generation-plan\"", "(\"generation-apply\""].as_slice(),
        ),
    ] {
        let text = repo_file(rel);
        for needle in needles {
            assert!(text.contains(needle), "{rel} lacks {needle}");
        }
    }
}

#[test]
fn workflow_text_names_the_step_in_english_and_korean() {
    for rel in WORKFLOWS {
        let text = repo_file(rel);
        assert_eq!(
            text.matches(NAMING_SENTENCE).count(),
            1,
            "{rel} carries the naming sentence exactly once"
        );
        assert!(
            text.contains(ENGLISH_AND_KOREAN),
            "{rel}: {ENGLISH_AND_KOREAN}"
        );
    }
}

#[test]
fn workflow_text_tells_to_translate_and_keeps_the_id_for_tools() {
    for rel in WORKFLOWS {
        let text = repo_file(rel);
        for clause in [TRANSLATE, ID_FOR_TOOLS, NEVER_THE_NAME] {
            assert!(text.contains(clause), "{rel} lacks {clause}");
        }
    }
}

#[test]
fn workflow_text_names_the_step_where_it_lists_it() {
    let update = repo_file(UPDATE);
    let step = line(&update, "1. **Preview everything once.**");
    assert!(
        step.contains("the project harness `files`"),
        "the update workflow lists the step's preview by its display name: {step}"
    );
    assert!(
        !step.contains("the generation `files`"),
        "the update workflow still lists the step by its id: {step}"
    );

    let project = repo_file(PROJECT);
    line(&project, "   - Project harness (from SETUP):");
    assert!(
        !project
            .lines()
            .any(|l| l.starts_with("   - Generation (from SETUP):")),
        "the harness step's item still carries the id as its title"
    );
    let report = line(&project, "Final report (");
    assert!(
        report.contains("hosts, the project harness, the official plugins"),
        "the final report table names the harness row by its display name: {report}"
    );
    assert!(
        !report.contains("hosts, project agents,"),
        "the final report table still names the row `project agents`: {report}"
    );
}

#[test]
fn refusal_notes_call_the_step_project_harness() {
    let undo = plan()["undo_order"].clone();
    let text = undo.to_string();
    for phrase in [
        "<agents_receipt of the project harness step>",
        "<receipt of the project harness step>",
        "refused while the project harness step's",
    ] {
        assert!(text.contains(phrase), "undo_order lacks {phrase}: {undo}");
    }
    assert!(
        !text.contains("generation step"),
        "undo_order still calls the step `the generation step`: {undo}"
    );
    // The step is still listed under its id.
    assert!(
        undo["steps"]
            .as_array()
            .expect("steps")
            .iter()
            .any(|s| s["step"] == "generation"),
        "{undo}"
    );
}
