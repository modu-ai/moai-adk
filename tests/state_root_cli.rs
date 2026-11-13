//! #2725 S2b: where the project state is written follows `state_root` — `.moai/` once its
//! marker says this generation owns it, otherwise the earlier `.moai-cli/` (decision D7:
//! both are read, one writes). Synthetic projects in temporary folders only.
mod common;
use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const MARKER: &str = r#"{"schema_version":1,"generation":"moai-cli","created_by":"moai-cli test"}"#;

fn profile() -> Value {
    json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"})
}

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
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "not JSON: {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// `project-plan` then `project-apply` with the plan's result, as PROJECT does.
fn create(root: &Path) -> Value {
    let p = root.to_str().unwrap();
    let plan = cli(&["project-plan", p], &profile());
    cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    )
}

fn tree(root: &Path) -> BTreeMap<String, Vec<u8>> {
    let mut out = BTreeMap::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in fs::read_dir(&dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p);
            } else {
                let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
                out.insert(rel, fs::read(&p).unwrap());
            }
        }
    }
    out
}

/// #2696 C2b: a project with no state at all gets `.moai/` and its marker from its first
/// profile; the earlier root is not created.
#[test]
fn a_project_with_no_state_starts_on_moai() {
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    let plan = cli(&["project-plan", p], &profile());
    assert_eq!(plan["action"], json!("create"), "{plan}");
    assert!(
        !root.path().join(".moai").exists(),
        "a preview writes nothing"
    );
    create(root.path());
    let marker: Value =
        serde_json::from_slice(&fs::read(root.path().join(".moai/moai-cli.json")).unwrap())
            .unwrap();
    assert_eq!(marker["generation"], json!("moai-cli"), "{marker}");
    assert_eq!(marker["schema_version"], json!(1), "{marker}");
    assert!(root.path().join(".moai/project.yaml").is_file());
    assert!(root
        .path()
        .join(".moai/manifests/profile-baseline.json")
        .is_file());
    assert!(!root.path().join(".moai-cli").exists());
    let detect = cli(&["project-detect", p], &json!({}));
    assert_eq!(
        detect["state_root"]["generation"],
        json!("rust"),
        "{detect}"
    );
    // A second run plans `unchanged` against the profile under `.moai/`.
    let again = cli(&["project-plan", p], &profile());
    assert_eq!(again["action"], json!("unchanged"), "{again}");
}

/// A project that already keeps its state in the earlier `.moai-cli/` (no marker) keeps
/// writing there; nothing appears under `.moai/`.
#[test]
fn a_project_without_the_marker_keeps_the_earlier_root() {
    let root = TempDir::new().unwrap();
    fs::create_dir(root.path().join(".moai-cli")).unwrap();
    create(root.path());
    assert!(root.path().join(".moai-cli/project.yaml").is_file());
    assert!(root
        .path()
        .join(".moai-cli/manifests/profile-baseline.json")
        .is_file());
    assert!(!root.path().join(".moai").exists(), "nothing under .moai/");
}

#[test]
fn a_project_with_the_marker_keeps_its_state_in_moai() {
    let root = TempDir::new().unwrap();
    fs::create_dir(root.path().join(".moai")).unwrap();
    fs::write(root.path().join(".moai/moai-cli.json"), MARKER).unwrap();
    create(root.path());
    assert!(root.path().join(".moai/project.yaml").is_file());
    assert!(root
        .path()
        .join(".moai/manifests/profile-baseline.json")
        .is_file());
    assert!(
        !root.path().join(".moai-cli").exists(),
        "the earlier root is not created"
    );
    // A second run plans `unchanged` against the profile it finds under `.moai/`.
    let again = cli(&["project-plan", root.path().to_str().unwrap()], &profile());
    assert_eq!(again["action"], json!("unchanged"), "{again}");
    // project-detect reads the same profile and says whose `.moai/` it is.
    let detect = cli(
        &["project-detect", root.path().to_str().unwrap()],
        &json!({}),
    );
    assert_eq!(
        detect["state_root"]["generation"],
        json!("rust"),
        "{detect}"
    );
    assert_eq!(detect["profile"]["project_id"], json!("demo"), "{detect}");
}

#[test]
fn a_mixed_moai_folder_is_left_alone_and_the_earlier_root_is_used() {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join(".moai/config")).unwrap();
    fs::write(root.path().join(".moai/moai-cli.json"), MARKER).unwrap();
    fs::write(root.path().join(".moai/config/config.yaml"), "x: 1\n").unwrap();
    let moai_before = tree(&root.path().join(".moai"));
    create(root.path());
    assert_eq!(
        tree(&root.path().join(".moai")),
        moai_before,
        "a mixed .moai/ is not written"
    );
    assert!(root.path().join(".moai-cli/project.yaml").is_file());
}

/// #2696 C2 (audit #2776 N2): what a migrated project is told names `.moai/`, not the earlier
/// root — the missing-profile step, the generation-manifest step after `project-apply`, and
/// a project without the marker still hears the earlier root.
#[test]
fn guidance_names_the_state_folder_the_project_uses() {
    let root = TempDir::new().unwrap();
    fs::create_dir(root.path().join(".moai")).unwrap();
    fs::write(root.path().join(".moai/moai-cli.json"), MARKER).unwrap();
    let p = root.path().to_str().unwrap();
    let detect = cli(&["project-detect", p], &json!({}));
    let actions = detect["next_actions"].to_string();
    assert!(actions.contains("no .moai/project.yaml yet"), "{actions}");
    assert!(!actions.contains(".moai-cli"), "{actions}");
    let applied = create(root.path());
    let next = applied["next_actions"].to_string();
    assert!(next.contains(".moai/manifests/generation.json"), "{next}");
    assert!(!next.contains(".moai-cli"), "{next}");

    // #2696 C2b: an empty folder's profile will be created under `.moai/`; a folder that
    // already has the earlier root hears that one.
    let plain = TempDir::new().unwrap();
    let detect = cli(
        &["project-detect", plain.path().to_str().unwrap()],
        &json!({}),
    );
    assert!(
        detect["next_actions"]
            .to_string()
            .contains("no .moai/project.yaml yet"),
        "{detect}"
    );
    fs::create_dir(plain.path().join(".moai-cli")).unwrap();
    let detect = cli(
        &["project-detect", plain.path().to_str().unwrap()],
        &json!({}),
    );
    assert!(
        detect["next_actions"]
            .to_string()
            .contains("no .moai-cli/project.yaml yet"),
        "{detect}"
    );
}
