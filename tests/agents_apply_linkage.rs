//! `agents-apply` run on its own after `generation-apply` (Issue #2127). New-file
//! discipline: existing fixtures and goldens stay untouched.
//!
//! The generation linkage fields in `generation.json` (`workflow_id`,
//! `interview_revision`) belong to the last generation apply; re-applying only the
//! AGENTS.md machinery must not reset them to null.

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const MANIFEST: &str = ".moai-cli/manifests/generation.json";

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

fn ok_result(out: &std::process::Output) -> Value {
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn plan_then_apply(root: &TempDir, op: &str, request: &Value) {
    let plan = ok_result(&cli(
        &[&format!("{op}-plan"), root.path().to_str().unwrap()],
        request,
    ));
    ok_result(&cli(
        &[
            &format!("{op}-apply"),
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    ));
}

fn manifest(root: &TempDir) -> Value {
    serde_json::from_slice(&fs::read(root.path().join(MANIFEST)).unwrap()).unwrap()
}

/// AC-1/AC-2: generation-apply records the linkage; a later agents-apply on its own
/// (same and changed language) keeps it.
#[test]
fn agents_apply_alone_keeps_generation_linkage() {
    let root = generation_project();
    plan_then_apply(
        &root,
        "generation",
        &json!({
            "protocol_version": 1,
            "language": "en",
            "workflow_reference": "skills/moai/workflows/",
            "local_file": "tracked",
            "workflow_id": "review-flow",
            "native_roles": [],
            "skill_projections": []
        }),
    );
    let before = manifest(&root);
    assert_eq!(before["workflow_id"], json!("review-flow"));
    assert_eq!(before["interview_revision"], json!("rc1-fixture-1"));

    for language in ["en", "ko"] {
        plan_then_apply(
            &root,
            "agents",
            &json!({
                "protocol_version": 1,
                "language": language,
                "workflow_reference": "skills/moai/workflows/",
                "local_file": "tracked"
            }),
        );
        let after = manifest(&root);
        assert_eq!(
            after["workflow_id"],
            json!("review-flow"),
            "agents-apply ({language}) keeps workflow_id: {after}"
        );
        assert_eq!(
            after["interview_revision"],
            json!("rc1-fixture-1"),
            "agents-apply ({language}) keeps interview_revision: {after}"
        );
        assert_eq!(
            after["profile_sha256"], before["profile_sha256"],
            "profile unchanged, digest unchanged"
        );
    }
}
