//! SPEC-20261121-base-agents-user-folder (REQ-8, REQ-9, REQ-14): a project's Claude Code roles
//! are written at `.claude/agents/moai/<role id>.md`, no base agent name is ever planned in a
//! project, and a flat role file that an earlier generation recorded moves into the folder when
//! nobody edited it and stays (as a conflict) when the user did. Every case runs the built
//! binary on a temporary project; the real HOME is never read or written.
#![cfg(unix)]

mod common;
use common::cli::ok_result;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
    sync::OnceLock,
};
use tempfile::TempDir;

const PURPOSE: &str = "Fix lint and test failures";

fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    // A project of the current state folder (`PROJECT_STATE_DIR`), as the fixture always was.
    let state = root.path().join(moai_cli::PROJECT_STATE_DIR);
    fs::create_dir_all(&state).unwrap();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"],"interview_revision":"rc1-fixture-1"});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(state.join("project.yaml"), &bytes).unwrap();
    root
}

fn build_request(roles: &[&str]) -> Value {
    let roles: Vec<Value> = roles
        .iter()
        .map(|id| json!({"role_id": id, "host": "claude", "purpose": PURPOSE, "tools": ["Bash", "Edit"]}))
        .collect();
    json!({
        "protocol_version": 1,
        "language": "en",
        "workflow_reference": "skills/moai/workflows",
        "local_file": "none",
        "workflow_id": "review-flow",
        "native_roles": roles,
        "skill_projections": []
    })
}

/// One temporary HOME for every case of this file: the binary never sees the real one.
fn home() -> &'static Path {
    static HOME: OnceLock<TempDir> = OnceLock::new();
    HOME.get_or_init(|| TempDir::new().unwrap()).path()
}

/// Runs `moai ARGS` with `input` (JSON) on stdin and the temporary HOME.
fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(common::MOAI_BIN)
        .args(args)
        .env("HOME", home())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
    {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    child.wait_with_output().unwrap()
}

fn plan(root: &TempDir, request: &Value) -> std::process::Output {
    cli(&["generation-plan", root.path().to_str().unwrap()], request)
}

fn apply(root: &TempDir, plan: &Value) -> std::process::Output {
    cli(
        &[
            "generation-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    )
}

fn actions(plan: &Value) -> Vec<(String, String)> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn action_of(plan: &Value, path: &str) -> Option<String> {
    actions(plan)
        .into_iter()
        .find(|(p, _)| p == path)
        .map(|(_, a)| a)
}

fn manifest_path(root: &Path) -> std::path::PathBuf {
    moai_cli::state_root::path(root, "manifests/generation.json")
}

/// The earlier layout: the role file back at `.claude/agents/<id>.md` and the manifest naming
/// that path, as the generation of an earlier version left them.
fn earlier_layout(root: &TempDir, id: &str) {
    let nested = format!(".claude/agents/moai/{id}.md");
    let flat = format!(".claude/agents/{id}.md");
    fs::rename(root.path().join(&nested), root.path().join(&flat)).unwrap();
    let path = manifest_path(root.path());
    let mut manifest: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    let mut changed = 0;
    for entry in manifest["owned_files"].as_array_mut().unwrap() {
        if entry["path"] == json!(nested) {
            entry["path"] = json!(flat);
            changed += 1;
        }
    }
    assert_eq!(
        changed, 1,
        "the manifest does not name {nested}: {manifest}"
    );
    fs::write(&path, serde_json::to_vec_pretty(&manifest).unwrap()).unwrap();
    let moai = root.path().join(".claude/agents/moai");
    if moai.is_dir() && fs::read_dir(&moai).unwrap().next().is_none() {
        fs::remove_dir(&moai).unwrap();
    }
}

/// REQ-8: a Claude Code role lands inside `agents/moai/`, with its name and tools, and nothing
/// the plan lists carries a base agent name.
#[test]
fn a_claude_role_is_written_inside_the_moai_folder() {
    let root = project();
    let request = build_request(&["prj-fixer", "prj-tester"]);
    let p = ok_result(&plan(&root, &request));
    for (path, _) in actions(&p) {
        let name = path.rsplit('/').next().unwrap();
        assert!(
            !name.starts_with("moai-"),
            "a base agent name is planned: {path}"
        );
    }
    ok_result(&apply(&root, &p));
    let mut inside: Vec<String> = fs::read_dir(root.path().join(".claude/agents/moai"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    inside.sort();
    assert_eq!(inside, ["prj-fixer.md", "prj-tester.md"]);
    let role = fs::read_to_string(root.path().join(".claude/agents/moai/prj-fixer.md")).unwrap();
    assert!(role.contains("name: prj-fixer"), "{role}");
    assert!(role.contains("tools: Bash, Edit"), "{role}");
    let top: Vec<String> = fs::read_dir(root.path().join(".claude/agents"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(top, ["moai"], "the top level of .claude/agents/");
    // A role named like a base agent is refused, and a second plan changes nothing.
    let refused = plan(&root, &build_request(&["moai-plan"]));
    assert!(!refused.status.success(), "{refused:?}");
    let again = ok_result(&plan(&root, &request));
    for (path, action) in actions(&again) {
        if path.contains(".claude/agents/") {
            assert_eq!(action, "unchanged", "{path}");
        }
    }
}

/// REQ-9: a flat role file that a generation recorded and nobody edited moves into the folder.
#[test]
fn an_unedited_flat_role_moves_into_the_moai_folder() {
    let root = project();
    let request = build_request(&["prj-fixer"]);
    ok_result(&apply(&root, &ok_result(&plan(&root, &request))));
    earlier_layout(&root, "prj-fixer");
    let before = fs::read(root.path().join(".claude/agents/prj-fixer.md")).unwrap();

    let p = ok_result(&plan(&root, &request));
    assert_eq!(
        action_of(&p, ".claude/agents/prj-fixer.md").as_deref(),
        Some("remove"),
        "{p}"
    );
    assert_eq!(
        action_of(&p, ".claude/agents/moai/prj-fixer.md").as_deref(),
        Some("create"),
        "{p}"
    );
    ok_result(&apply(&root, &p));
    assert!(!root.path().join(".claude/agents/prj-fixer.md").exists());
    assert_eq!(
        fs::read(root.path().join(".claude/agents/moai/prj-fixer.md")).unwrap(),
        before
    );
    let again = ok_result(&plan(&root, &request));
    for (path, action) in actions(&again) {
        if path.contains(".claude/agents/") {
            assert_eq!(action, "unchanged", "{path}");
        }
    }
}

/// REQ-14: a flat role file the user edited stays, the plan lists it as a conflict, and the
/// nested copy is not written (the apply of a plan with a conflict writes nothing).
#[test]
fn an_edited_flat_role_stays_and_blocks_the_nested_copy() {
    let root = project();
    let request = build_request(&["prj-fixer"]);
    ok_result(&apply(&root, &ok_result(&plan(&root, &request))));
    earlier_layout(&root, "prj-fixer");
    let flat = root.path().join(".claude/agents/prj-fixer.md");
    fs::write(
        &flat,
        "---\nname: prj-fixer\ndescription: mine\n---\n\nmine\n",
    )
    .unwrap();

    let p = ok_result(&plan(&root, &request));
    assert_eq!(
        action_of(&p, ".claude/agents/prj-fixer.md").as_deref(),
        Some("conflict"),
        "{p}"
    );
    let out = apply(&root, &p);
    assert!(!out.status.success(), "{out:?}");
    assert_eq!(
        fs::read_to_string(&flat).unwrap(),
        "---\nname: prj-fixer\ndescription: mine\n---\n\nmine\n"
    );
    assert!(!root
        .path()
        .join(".claude/agents/moai/prj-fixer.md")
        .exists());
}
