//! The two operations of the CI files (SPEC-20261122-project-ci-design, REQ-6 to REQ-8):
//! `project-ci-plan` previews, `project-ci-apply` publishes exactly the plan the user picked.
//! The library is tested in `project_ci_plan_lib.rs`; here the binary is driven in temporary
//! folders with a temporary HOME of its own, so no test reads or writes the real one.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Output, Stdio},
};
use tempfile::TempDir;

const CHECK: &str = ".github/workflows/moai-check.yml";
const NIGHTLY: &str = ".github/workflows/moai-nightly.yml";

fn run(args: &[&str], stdin: &Value) -> Output {
    let home = TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(stdin.to_string().as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v
}

fn profile(flow: &str) -> Value {
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "work_flow": {"flow": flow, "integration_branch": "trunk"},
        "detected": {"project_kind": "existing", "manifests": ["Cargo.toml"],
                     "languages": ["rust"], "lsp": [], "mcp_servers": [], "databases": []},
        "ci": {"kind": "cli", "systems": ["linux"], "deploy": {"enabled": false},
               "visibility": "public", "strict": "report"},
        "checks": [
            {"id": "build", "command": "cargo build"},
            {"id": "test", "command": "cargo test"}
        ]
    })
}

fn project(flow: &str) -> TempDir {
    let root = TempDir::new().unwrap();
    set_profile(root.path(), flow);
    root
}

fn set_profile(root: &Path, flow: &str) {
    let path = root.join(".moai-cli/project.yaml");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, serde_json::to_vec_pretty(&profile(flow)).unwrap()).unwrap();
}

fn plan(root: &TempDir, request: Value) -> Value {
    ok(&run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &request,
    ))
}

fn apply(root: &TempDir, plan: &Value, digest: &str, permission: &str) -> Output {
    run(
        &[
            "project-ci-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ],
        plan,
    )
}

fn digest(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
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

/// REQ-6: the plan prints the files, the estimate and the digest, and writes nothing.
#[test]
fn the_plan_previews_and_writes_nothing() {
    let root = project("m-flow");
    let before = fs::read_dir(root.path()).unwrap().count();
    let out = run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &json!({"protocol_version": 1}),
    );
    let v = envelope(&out);
    assert_eq!(v["operation"], "project.ci_plan", "{v}");
    let p = ok(&out);
    assert_eq!(
        actions(&p),
        vec![
            (CHECK.to_string(), "create".to_string()),
            (NIGHTLY.to_string(), "create".to_string())
        ],
        "{p}"
    );
    assert!(p["estimate"]["free"].as_bool().unwrap(), "{p}");
    assert!(digest(&p).starts_with("sha256:"), "{p}");
    assert!(!root.path().join(".github").exists());
    assert_eq!(fs::read_dir(root.path()).unwrap().count(), before);
}

/// REQ-7: the pick is the digest plus a permission reference; the files exist only after it,
/// and a second apply changes nothing.
#[test]
fn the_pick_writes_exactly_the_plan_and_a_second_apply_changes_nothing() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    let out = apply(&root, &p, &digest(&p), "fixture-approved");
    let v = envelope(&out);
    assert_eq!(v["operation"], "project.ci_apply", "{v}");
    let done = ok(&out);
    assert_eq!(done["written"], json!([CHECK, NIGHTLY]), "{done}");
    assert!(root.path().join(CHECK).is_file());
    assert!(fs::read_to_string(root.path().join(CHECK))
        .unwrap()
        .contains("persist-credentials: false"));

    let again = plan(&root, json!({"protocol_version": 1}));
    assert_eq!(
        actions(&again),
        vec![
            (CHECK.to_string(), "unchanged".to_string()),
            (NIGHTLY.to_string(), "unchanged".to_string())
        ]
    );
    let second = ok(&apply(&root, &again, &digest(&again), "fixture-approved"));
    assert_eq!(second["written"], json!([]), "{second}");
    assert!(second["request_id"].is_null(), "{second}");
}

/// REQ-7: no permission reference, a wrong digest, a plan changed by hand and a plan the
/// profile has since outgrown each write nothing.
#[test]
fn a_missing_wrong_or_stale_pick_writes_nothing() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    for (d, permission) in [
        (digest(&p), ""),
        (digest(&p), "null"),
        ("sha256:not-the-digest".to_string(), "fixture-approved"),
    ] {
        refused(&apply(&root, &p, &d, permission));
    }
    let mut altered = p.clone();
    altered["files"][0]["rendered_sha256"] = json!("sha256:00");
    refused(&apply(&root, &altered, &digest(&p), "fixture-approved"));
    set_profile(root.path(), "github-flow");
    refused(&apply(&root, &p, &digest(&p), "fixture-approved"));
    assert!(!root.path().join(".github").exists());
}

/// A request that is not the typed shape is refused by the plan with the operation named.
#[test]
fn a_wrong_request_is_refused_with_its_operation_named() {
    let root = project("m-flow");
    for request in [json!({}), json!({"protocol_version": 2}), json!([])] {
        let v = refused(&run(
            &["project-ci-plan", root.path().to_str().unwrap()],
            &request,
        ));
        assert_eq!(v["operation"], "project.ci_plan", "{v}");
    }
}

/// B2-N2 through the binary: the removal of a file the plan stopped owning is a pick in the
/// request, previewed as `remove`, applied with the digest.
#[test]
fn a_file_the_plan_stopped_owning_is_removed_only_when_picked() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    ok(&apply(&root, &p, &digest(&p), "fixture-approved"));
    set_profile(root.path(), "github-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    assert!(actions(&p).contains(&(NIGHTLY.to_string(), "stale".to_string())));
    let picked = plan(&root, json!({"protocol_version": 1, "remove": [NIGHTLY]}));
    assert!(actions(&picked).contains(&(NIGHTLY.to_string(), "remove".to_string())));
    assert!(root.path().join(NIGHTLY).exists(), "the plan removed it");
    let done = ok(&apply(&root, &picked, &digest(&picked), "fixture-approved"));
    assert_eq!(done["removed"], json!([NIGHTLY]), "{done}");
    assert!(!root.path().join(NIGHTLY).exists());
    // A pick for a path the plan does not own is refused.
    refused(&run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &json!({"protocol_version": 1, "remove": [".github/workflows/ci.yml"]}),
    ));
}
