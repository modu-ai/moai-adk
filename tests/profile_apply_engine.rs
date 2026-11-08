//! Issue #2071: the project profile writes through the 01-B apply engine only, and its
//! receipt digests, plan `expected_sha256` and recovery agree with the files on disk.
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn cli(args: &[&str], input: &Value) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
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
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}
fn profile(purpose: &str) -> Value {
    json!({"schema_version":2,"project_id":"demo","purpose":purpose,"tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"})
}
fn apply(root: &Path, purpose: &str) -> (Value, Value) {
    let p = root.to_str().unwrap();
    let plan = cli(&["project-plan", p], &profile(purpose));
    let applied = cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    (plan, applied)
}
fn sha(path: &Path) -> Option<String> {
    let bytes = fs::read(path).ok()?;
    Some(moai_cli::digest(&bytes))
}

/// AC-2 (behaviour): for both profile files, the plan's expected digest and the receipt's
/// before/after digests are the bytes actually on disk before and after the apply, and
/// recovery puts the before bytes back.
#[test]
fn profile_receipt_digests_match_the_files() {
    let root = TempDir::new().unwrap();
    let profile_path = root.path().join(".moai-cli/project.yaml");
    let baseline_path = root
        .path()
        .join(".moai-cli/manifests/profile-baseline.json");
    apply(root.path(), "first");
    let (before_p, before_b) = (sha(&profile_path), sha(&baseline_path));
    assert!(before_p.is_some() && before_b.is_some());
    let (plan, applied) = apply(root.path(), "second");
    assert_eq!(plan["action"], json!("update"));
    assert_eq!(plan["expected_sha256"], json!(before_p));
    let receipt_id = applied["receipt"].as_str().unwrap();
    let receipt: Value = serde_json::from_slice(
        &fs::read(
            root.path()
                .join(format!(".moai-cli/receipts/{receipt_id}.json")),
        )
        .unwrap(),
    )
    .unwrap();
    let entry = |rel: &str| {
        receipt["entries"]
            .as_array()
            .unwrap()
            .iter()
            .find(|e| e["path"] == json!(rel))
            .cloned()
            .unwrap_or_else(|| panic!("no entry for {rel}"))
    };
    for (rel, path, before) in [
        (".moai-cli/project.yaml", &profile_path, &before_p),
        (
            ".moai-cli/manifests/profile-baseline.json",
            &baseline_path,
            &before_b,
        ),
    ] {
        let e = entry(rel);
        assert_eq!(e["published"], json!(true), "{rel}");
        assert_eq!(e["before_sha256"], json!(before), "{rel}");
        assert_eq!(e["after_sha256"], json!(sha(path)), "{rel}");
    }
    let recovered = cli(
        &[
            "apply-recover",
            root.path().to_str().unwrap(),
            "--receipt",
            receipt_id,
            "--permission-reference",
            "fixture-approved",
        ],
        &json!({}),
    );
    assert_eq!(recovered["state"], json!("reverted"), "{recovered}");
    assert_eq!(
        (sha(&profile_path), sha(&baseline_path)),
        (before_p, before_b)
    );
}

/// AC-2 (single path): profile.rs publishes nothing itself; every write goes through
/// `apply::publish_replace` / `apply::publish_exclusive`.
#[test]
fn profile_writes_only_through_the_apply_engine() {
    let src = fs::read_to_string(concat!(env!("CARGO_MANIFEST_DIR"), "/src/profile.rs")).unwrap();
    for own in [
        "fs::rename(",
        "fs::hard_link(",
        ".moai-apply-",
        "OpenOptions::new()",
    ] {
        assert!(
            !src.contains(own),
            "profile.rs still publishes on its own: {own}"
        );
    }
    assert!(src.contains("apply::publish_replace") && src.contains("apply::publish_exclusive"));
}
