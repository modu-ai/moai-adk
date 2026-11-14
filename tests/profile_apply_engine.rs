//! Issue #2071: the project profile writes through the 01-B apply engine only, and its
//! receipt digests, plan `expected_sha256` and recovery agree with the files on disk.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

fn cli(args: &[&str], input: &Value) -> Value {
    let mut child = Command::new(crate::common::MOAI_BIN)
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
    // An earlier-root project: its state stays in `.moai-cli/` (#2696 C2b gives a project
    // with no state `.moai/`; that layout is pinned in tests/state_root_cli.rs).
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
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

/// #2590 AC-1: a hand edit of `project.yaml` after an apply shows as `drift` in the next
/// plan, even when the input equals the baseline and the action stays `unchanged` (the
/// edit is kept). A file that still matches the baseline carries no `drift` key.
#[test]
fn a_hand_edit_after_apply_shows_as_drift() {
    let root = TempDir::new().unwrap();
    // An earlier-root project: its state stays in `.moai-cli/` (#2696 C2b gives a project
    // with no state `.moai/`; that layout is pinned in tests/state_root_cli.rs).
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let p = root.path().to_str().unwrap();
    apply(root.path(), "first");
    let clean = cli(&["project-plan", p], &profile("first"));
    assert_eq!(clean["action"], json!("unchanged"), "{clean}");
    assert!(clean.get("drift").is_none(), "{clean}");

    let path = root.path().join(".moai-cli/project.yaml");
    let mut file: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    file["purpose"] = json!("edited by hand");
    file["tools"] = json!(["Rust", "Python"]);
    file["my_own_key"] = json!("not managed");
    fs::write(&path, serde_json::to_vec_pretty(&file).unwrap()).unwrap();

    let drifted = cli(&["project-plan", p], &profile("first"));
    assert_eq!(drifted["action"], json!("unchanged"), "{drifted}");
    assert_eq!(drifted["drift"], json!(["purpose", "tools"]), "{drifted}");
    assert_ne!(drifted["plan_digest"], clean["plan_digest"]);
    // The hand edit is never overwritten by applying that plan.
    cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            drifted["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &drifted,
    );
    let after: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(after["purpose"], json!("edited by hand"));
}

/// #2590 AC-2: in this repository the apply receipts stay out of git and the profile
/// baseline is tracked next to `project.yaml` (docs/rust-product-path.md, "Per-file
/// receipts and recovery").
#[test]
fn receipts_are_ignored_and_the_baseline_is_tracked_here() {
    let ignore = fs::read_to_string(concat!(env!("CARGO_MANIFEST_DIR"), "/.gitignore")).unwrap();
    let rules: Vec<&str> = ignore
        .lines()
        .map(str::trim)
        .filter(|l| !l.is_empty() && !l.starts_with('#'))
        .collect();
    assert!(
        rules.contains(&"/.moai-cli/receipts/"),
        "receipts rule missing"
    );
    for rule in &rules {
        let r = rule.trim_start_matches('/').trim_end_matches('/');
        assert!(
            !(r == ".moai-cli"
                || r == ".moai-cli/manifests"
                || r.ends_with("profile-baseline.json")
                || r == "*.json"),
            "rule {rule} would hide the profile baseline"
        );
    }
}
