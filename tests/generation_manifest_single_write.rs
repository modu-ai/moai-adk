//! Interruption between the `agents_apply` manifest write and the generation
//! manifest write (Issue #2114). New-file discipline: existing fixtures and
//! `src/agents.rs` behavior stay untouched.
//!
//! `generation_apply` delegates the agents portion to `agents_apply`, which rewrites
//! `generation.json` with the linkage fields (`workflow_id`, `interview_revision`)
//! reset, and only afterwards restores them in its own manifest write. A failure in
//! between must not leave a manifest that lost what the previous successful
//! generation apply recorded.

#![cfg(unix)]

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    os::unix::fs::PermissionsExt,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const MANIFEST: &str = ".moai-cli/manifests/generation.json";
const WORKFLOW_DIR: &str = ".moai-cli/workflows/review-flow";

fn generation_project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), &bytes).unwrap();
    root
}

fn request(language: &str) -> Value {
    json!({
        "protocol_version": 1,
        "language": language,
        "workflow_reference": "skills/moai/workflows/",
        "local_file": "tracked",
        "workflow_id": "review-flow",
        "native_roles": [],
        "skill_projections": []
    })
}

fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(EXE)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let data = input.to_string();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(data.as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn plan(root: &TempDir, language: &str) -> Value {
    let out = cli(
        &["generation-plan", root.path().to_str().unwrap()],
        &request(language),
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
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

fn manifest(root: &TempDir) -> Value {
    serde_json::from_slice(&fs::read(root.path().join(MANIFEST)).unwrap()).unwrap()
}

fn generation_paths(m: &Value) -> Vec<String> {
    let mut paths: Vec<String> = m["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["section_id"]
                .as_str()
                .is_some_and(|s| s.starts_with("generation:"))
        })
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect();
    paths.sort();
    paths
}

/// #2109's bare `kind` rides the single manifest publish: the hook's value is
/// re-decoded through the typed `OwnedFile`, which must keep it, and an interrupted
/// apply leaves the previous `kind` entries in place.
#[test]
fn single_publish_keeps_owned_file_kind() {
    let root = generation_project();
    let out = apply(&root, &plan(&root, "en"));
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let kinds = |m: &Value| -> Vec<(String, Value)> {
        let mut v: Vec<(String, Value)> = m["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| {
                f["section_id"]
                    .as_str()
                    .is_some_and(|s| s.starts_with("generation:"))
            })
            .map(|f| (f["path"].as_str().unwrap().to_string(), f["kind"].clone()))
            .collect();
        v.sort_by(|a, b| a.0.cmp(&b.0));
        v
    };
    let before = kinds(&manifest(&root));
    assert!(!before.is_empty());
    assert!(
        before.iter().all(|(_, k)| *k == json!("workflow")),
        "every generation entry carries kind after the typed re-decode: {before:?}"
    );

    let second = plan(&root, "ko");
    let dir = root.path().join(WORKFLOW_DIR);
    let original = fs::metadata(&dir).unwrap().permissions();
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o555)).unwrap();
    let out = apply(&root, &second);
    fs::set_permissions(&dir, original).unwrap();
    assert!(!out.status.success());
    assert_eq!(
        kinds(&manifest(&root)),
        before,
        "kind survives the interruption"
    );
}

/// AC-1/AC-2: a generation file publish fails after `agents_apply` has already
/// rewritten the manifest. The manifest must still carry the generation baselines
/// and the linkage fields the previous successful apply recorded.
#[test]
fn interrupted_generation_apply_keeps_manifest_generation_state() {
    let root = generation_project();
    let first = plan(&root, "en");
    let out = apply(&root, &first);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let before = manifest(&root);
    assert_eq!(before["workflow_id"], json!("review-flow"));
    assert_eq!(before["interview_revision"], json!("rc1-fixture-1"));
    let before_paths = generation_paths(&before);
    assert!(
        !before_paths.is_empty(),
        "first apply recorded generation files"
    );

    // en -> ko re-renders AGENTS.md (agents portion) and the workflow files
    // (generation portion, `merge`), so both writers run.
    let second = plan(&root, "ko");
    assert!(
        second["files"].as_array().unwrap().iter().any(|f| f["kind"]
            .as_str()
            .is_some_and(|k| k.starts_with("generation:workflow:"))
            && f["action"] == json!("merge")),
        "the second plan merges a workflow file: {second}"
    );

    // Inject the interruption: the workflow directory refuses the staged temp file.
    let dir = root.path().join(WORKFLOW_DIR);
    let original = fs::metadata(&dir).unwrap().permissions();
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o555)).unwrap();
    let out = apply(&root, &second);
    fs::set_permissions(&dir, original).unwrap();

    assert!(
        !out.status.success(),
        "injected failure must fail the apply"
    );
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(
        text.contains("partial apply") && text.contains(WORKFLOW_DIR),
        "the failure is at the generation file publish, after the agents portion: {text}"
    );

    let after = manifest(&root);
    assert_eq!(
        generation_paths(&after),
        before_paths,
        "generation baselines survive the interruption"
    );
    assert_eq!(
        after["workflow_id"],
        json!("review-flow"),
        "workflow_id survives the interruption: {after}"
    );
    assert_eq!(
        after["interview_revision"],
        json!("rc1-fixture-1"),
        "interview_revision survives the interruption: {after}"
    );
}
