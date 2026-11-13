//! `workflow_reference` trailing-slash normalization fixtures (Issue #2095).
//! New-file discipline: `src/agents.rs` and every existing fixture/test file are
//! untouched; the canonical form is the slash-terminated spelling that the recorded
//! configuration and the existing golden renders already carry.

mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

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

fn request_with(reference: &str) -> Value {
    json!({
        "protocol_version": 1,
        "language": "en",
        "workflow_reference": reference,
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

fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}

fn plan_for(root: &TempDir, reference: &str) -> Value {
    let out = cli(
        &["generation-plan", root.path().to_str().unwrap()],
        &request_with(reference),
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    ok_result(&out)
}

fn agents_entry(plan: &Value) -> &Value {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!("AGENTS.md"))
        .expect("AGENTS.md in the plan")
}

/// AC-1: the slash and no-slash spellings of one `workflow_reference` produce the
/// same generation-plan output — same plan digest, same canonical request field,
/// same AGENTS.md render sha. The canonical form keeps the trailing slash (the
/// recorded configuration and the golden renders carry it).
#[test]
fn generation_plan_reference_slash_is_canonical() {
    let root = generation_project();
    let with_slash = plan_for(&root, "skills/moai/workflows/");
    let without_slash = plan_for(&root, "skills/moai/workflows");
    let double_slash = plan_for(&root, "skills/moai/workflows//");

    assert_eq!(
        with_slash["plan_digest"], without_slash["plan_digest"],
        "one meaning, one plan digest"
    );
    assert_eq!(with_slash["plan_digest"], double_slash["plan_digest"]);
    for plan in [&with_slash, &without_slash, &double_slash] {
        assert_eq!(
            plan["request"]["workflow_reference"],
            json!("skills/moai/workflows/"),
            "the plan carries the canonical slash-terminated form"
        );
    }
    assert_eq!(
        agents_entry(&with_slash)["rendered_sha256"],
        agents_entry(&without_slash)["rendered_sha256"],
        "the AGENTS.md render is byte-identical"
    );
}

/// A plan body carrying the raw no-slash spelling together with the canonical digest
/// is the same plan after normalization: apply accepts it and renders the canonical
/// slash-terminated bytes. (A plan whose recorded digest was computed over the raw
/// spelling by a pre-#2095 binary fails the same re-derivation and is refused as
/// stale — covered by the existing stale-plan fixtures.)
#[test]
fn generation_apply_normalizes_raw_spelling_to_canonical() {
    let root = generation_project();
    let plan = plan_for(&root, "skills/moai/workflows/");
    let mut raw_body = plan.clone();
    raw_body["request"]["workflow_reference"] = json!("skills/moai/workflows");

    let out = cli(
        &[
            "generation-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &raw_body,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let agents_md = fs::read_to_string(root.path().join("AGENTS.md")).unwrap();
    assert!(
        agents_md.contains("Read only the workflow you need from `skills/moai/workflows/`."),
        "the managed block renders the canonical slash-terminated reference"
    );
    assert!(
        !agents_md.contains("from `skills/moai/workflows`."),
        "the raw spelling never reaches the render"
    );
}
