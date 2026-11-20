//! #2936 R-fix R-C (PLAN 6880040242 R7): the writers that remove a project skill leave no empty
//! `.agents/skills` (or `.claude/skills`) behind. `generation-apply` removing the last project
//! skill and `project-skill-revert` of a copy it created prune the folders their removal
//! emptied, up to the project root and no further; a folder that still holds anything stays.
//!
//! Every case builds a temporary project; the child runs with an empty temporary HOME and a
//! cleared environment, so nothing of the developer's machine is read.

#![cfg(unix)]

mod common;

use common::CleanEnv;
use moai_cli::project_skills::{
    project_skill_apply, project_skill_revert, ProjectSkillRequest, ProjectSkillRevertRequest,
};
use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn cli(args: &[&str], input: &Value) -> std::process::Output {
    // An empty temporary HOME for the child; it lives until the child has exited.
    let home = TempDir::new().unwrap();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn generation_project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let profile = json!({"schema_version": 2, "project_id": "demo",
        "purpose": "Build a small editor", "tools": ["Rust"], "development_methods": ["TDD"],
        "hosts": ["claude", "codex"], "interview_revision": "r1", "decisions": [
            {"id": "test_command", "topic": "test_command", "value": "cargo test --locked",
             "source": "user", "scope": "project"}]});
    fs::write(
        root.path().join(".moai-cli/project.yaml"),
        serde_json::to_vec_pretty(&profile).unwrap(),
    )
    .unwrap();
    root
}

fn request(skills: Value) -> Value {
    let mut r = json!({"protocol_version": 1, "language": "en",
        "workflow_reference": "skills/moai/workflows", "local_file": "none",
        "workflow_id": "review-flow", "native_roles": [], "skill_projections": []});
    if !skills.is_null() {
        r["project_skills"] = skills;
    }
    r
}

fn plan(root: &Path, req: &Value) -> Value {
    result(&cli(&["generation-plan", root.to_str().unwrap()], req))
}

fn apply(root: &Path, p: &Value) -> Value {
    result(&cli(
        &[
            "generation-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        p,
    ))
}

fn skills(ids: &[&str]) -> Value {
    Value::Array(
        ids.iter()
            .map(|id| json!({"id": id, "purpose": "a project skill"}))
            .collect(),
    )
}

const CODEX_LAST: &str = ".agents/skills/moai-prj-checks/SKILL.md";

// ---- AC-9: generation-apply removing the last project skill ----

#[test]
fn removing_the_last_project_skill_leaves_no_empty_skills_folder() {
    let root = generation_project();
    let r = root.path();
    apply(r, &plan(r, &request(skills(&["prj-checks"]))));
    assert!(r.join(CODEX_LAST).exists());

    let p = plan(r, &request(Value::Null));
    apply(r, &p);
    assert!(!r.join(CODEX_LAST).exists());
    for gone in [
        ".agents/skills/moai-prj-checks",
        ".agents/skills",
        ".agents",
        ".claude/skills",
    ] {
        assert!(!r.join(gone).exists(), "{gone} was left behind");
    }
    // The project root itself, and what the apply never wrote, stay.
    assert!(r.join("AGENTS.md").exists());
}

/// The miss direction: a sibling keeps the folders above it, a file of the user's keeps them
/// too.
#[test]
fn a_folder_that_still_holds_something_stays() {
    let root = generation_project();
    let r = root.path();
    apply(r, &plan(r, &request(skills(&["prj-checks", "prj-extra"]))));
    let p = plan(r, &request(skills(&["prj-checks"])));
    apply(r, &p);
    assert!(!r.join(".agents/skills/moai-prj-extra").exists());
    assert!(r.join(".agents/skills/moai-prj-checks/SKILL.md").exists());

    // Now the last one goes, but the user has a skill of their own beside it.
    fs::create_dir_all(r.join(".agents/skills/mine")).unwrap();
    fs::write(r.join(".agents/skills/mine/SKILL.md"), "mine\n").unwrap();
    apply(r, &plan(r, &request(Value::Null)));
    assert!(!r.join(".agents/skills/moai-prj-checks").exists());
    assert!(r.join(".agents/skills/mine/SKILL.md").exists(), "{r:?}");
    assert!(r.join(".agents").exists());
}

/// `apply-recover` puts the last skill back, folders and all.
#[test]
fn the_last_project_skill_can_be_recovered_after_its_folders_went() {
    let root = generation_project();
    let r = root.path();
    apply(r, &plan(r, &request(skills(&["prj-checks"]))));
    let original = fs::read(r.join(CODEX_LAST)).unwrap();
    let applied = apply(r, &plan(r, &request(Value::Null)));
    assert!(!r.join(".agents").exists());
    let receipt = applied["receipt"].as_str().unwrap().to_string();
    result(&cli(
        &[
            "apply-recover",
            r.to_str().unwrap(),
            "--receipt",
            &receipt,
            "--permission-reference",
            "fixture-approved",
        ],
        &json!({}),
    ));
    assert_eq!(fs::read(r.join(CODEX_LAST)).unwrap(), original);
}

// ---- AC-10: project-skill-revert ----

fn write_source(dir: &Path) {
    fs::create_dir_all(dir.join("references")).unwrap();
    fs::write(
        dir.join("SKILL.md"),
        b"---\nname: moai\ndescription: fixture\n---\n",
    )
    .unwrap();
    fs::write(dir.join("references/a.md"), b"a").unwrap();
}

fn skill_apply(project: &Path, source: &Path, hosts: &[&str]) -> String {
    project_skill_apply(&ProjectSkillRequest {
        protocol_version: 1,
        opt_in: true,
        project_root: project.to_path_buf(),
        home: None,
        hosts: hosts.iter().map(|h| (*h).to_string()).collect(),
        source_dir: source.to_path_buf(),
    })
    .unwrap()
    .request_id
}

fn skill_revert(project: &Path, request_id: &str) -> Value {
    project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: 1,
        project_root: project.to_path_buf(),
        request_id: request_id.to_string(),
    })
    .unwrap()
}

#[test]
fn reverting_a_created_copy_leaves_no_empty_skills_folder() {
    let project = TempDir::new().unwrap();
    let source = TempDir::new().unwrap();
    write_source(source.path());
    let id = skill_apply(project.path(), source.path(), &["codex", "claude"]);
    assert!(project.path().join(".agents/skills/moai/SKILL.md").exists());
    skill_revert(project.path(), &id);
    for gone in [
        ".agents/skills/moai",
        ".agents/skills",
        ".agents",
        ".claude/skills/moai",
        ".claude/skills",
        ".claude",
    ] {
        assert!(
            !project.path().join(gone).exists(),
            "{gone} was left behind"
        );
    }
}

#[test]
fn a_revert_keeps_the_folders_that_hold_other_files() {
    let project = TempDir::new().unwrap();
    let source = TempDir::new().unwrap();
    write_source(source.path());
    // The user already has a skill of their own under both hosts' folders.
    for host_dir in [".agents/skills/mine", ".claude/skills/mine"] {
        fs::create_dir_all(project.path().join(host_dir)).unwrap();
        fs::write(project.path().join(host_dir).join("SKILL.md"), "mine\n").unwrap();
    }
    let id = skill_apply(project.path(), source.path(), &["codex", "claude"]);
    skill_revert(project.path(), &id);
    assert!(!project.path().join(".agents/skills/moai").exists());
    assert!(!project.path().join(".claude/skills/moai").exists());
    for kept in [
        ".agents/skills/mine/SKILL.md",
        ".claude/skills/mine/SKILL.md",
    ] {
        assert!(project.path().join(kept).exists(), "{kept} was removed");
    }
}
