//! generation-apply receipt ordering (Issue #2119). New-file discipline: existing
//! fixtures and goldens stay untouched.
//!
//! The `generation.apply` receipt must not read `complete` when the one manifest
//! publish that follows the generation files failed.

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
const MANIFEST_DIR: &str = ".moai-cli/manifests";
const RECEIPTS_DIR: &str = ".moai-cli/receipts";

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

/// Every `generation.apply` receipt on disk, as (status, published paths).
fn generation_receipts(root: &TempDir) -> Vec<(String, Vec<String>)> {
    let mut out = Vec::new();
    for entry in fs::read_dir(root.path().join(RECEIPTS_DIR)).unwrap() {
        let path = entry.unwrap().path();
        if path.extension().and_then(|e| e.to_str()) != Some("json") {
            continue;
        }
        let v: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
        if v["operation"] != json!("generation.apply") {
            continue;
        }
        let published = v["entries"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|e| e["published"] == json!(true))
            .map(|e| e["path"].as_str().unwrap().to_string())
            .collect();
        out.push((v["status"].as_str().unwrap().to_string(), published));
    }
    out
}

/// AC-1 path 2: the manifest publish fails after the generation files landed. The
/// generation receipt of that run must not say `complete`.
#[test]
fn manifest_publish_failure_leaves_generation_receipt_partial() {
    let root = generation_project();
    let out = apply(&root, &plan(&root, "en"));
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert_eq!(
        generation_receipts(&root)
            .iter()
            .filter(|(s, _)| s == "complete")
            .count(),
        1,
        "the first, successful apply is complete"
    );

    // en -> ko: the workflow files merge (generation portion) and the manifest
    // changes; the manifest directory refuses the staged temp file.
    let second = plan(&root, "ko");
    let dir = root.path().join(MANIFEST_DIR);
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
        text.contains("partial apply") && text.contains("generation.json"),
        "the failure is at the manifest publish: {text}"
    );

    let receipts = generation_receipts(&root);
    assert_eq!(
        receipts.len(),
        2,
        "one generation receipt per run: {receipts:?}"
    );
    assert_eq!(
        receipts.iter().filter(|(s, _)| s == "complete").count(),
        1,
        "the failed run's generation receipt must not read complete: {receipts:?}"
    );
    assert!(
        receipts.iter().any(|(s, published)| s == "partial"
            && published
                .iter()
                .any(|p| p.starts_with(".moai-cli/workflows/review-flow/"))),
        "the failed run's receipt reads partial and lists the files it published: {receipts:?}"
    );
}
