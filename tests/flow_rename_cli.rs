//! The flow rename (SPEC-20261122-next-flow), through the binary in temporary folders only:
//! the value `next-flow` and the old value `m-flow` select the same flow (AC-1), and a new
//! choice records `next-flow` while a stored old value is never rewritten by a reader (AC-2).

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

/// Runs the binary with a temporary HOME of its own, so no test reads or writes the real one.
fn run(args: &[&str], stdin: &[u8]) -> Output {
    let home = TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(stdin) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

fn ok(out: &Output) -> Value {
    assert!(out.status.success(), "{}", text(out));
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn p(root: &TempDir) -> &str {
    root.path().to_str().unwrap()
}

fn base() -> Value {
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]
    })
}

fn with(flow: &str) -> Value {
    let mut profile = base();
    profile["work_flow"] = json!({"flow": flow, "integration_branch": "v2026.11.0"});
    profile
}

fn plan(root: &TempDir, profile: &Value) -> Output {
    run(&["project-plan", p(root)], profile.to_string().as_bytes())
}

/// `work_flow_remote` (and the rename's `flow_notice`) are preview-only attachments of
/// `project-plan`'s answer: the apply input type never carries them, so the round trip
/// removes them the way every preview-only field leaves a plan before it is applied.
fn applyable(plan: &Value) -> Value {
    let mut plan = plan.clone();
    if let Some(object) = plan.as_object_mut() {
        object.remove("work_flow_remote");
        object.remove("flow_notice");
    }
    plan
}

fn apply(root: &TempDir, plan: &Value) -> Output {
    let plan = applyable(plan);
    run(
        &[
            "project-apply",
            p(root),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    )
}

/// The project profile the apply wrote, so a stored value is read from disk, not echoed.
fn stored_profile(root: &Path) -> String {
    let yaml = root.join(".moai").join("project.yaml");
    let text = fs::read_to_string(&yaml).unwrap_or_default();
    assert!(!text.is_empty(), "the apply wrote no project.yaml");
    text
}

/// AC-1 (SPEC REQ-1, REQ-2): both values select the same flow, the old value is
/// named once on stderr (the plan object stays applyable), and an unknown value
/// is still refused.
#[test]
fn ac1_both_values_select_the_same_flow_and_the_old_is_named_once() {
    let root = TempDir::new().unwrap();
    let new_out = plan(&root, &with("next-flow"));
    let new = ok(&new_out);
    let old_out = plan(&root, &with("m-flow"));
    let old = ok(&old_out);
    assert_eq!(
        new["work_flow_remote"], old["work_flow_remote"],
        "the two values must select the same flow"
    );
    let notice = "m-flow is the old name of the NeXT-Flow flow; both values select the same flow";
    let stderr = String::from_utf8_lossy(&old_out.stderr);
    assert_eq!(
        stderr.lines().filter(|line| line.contains(notice)).count(),
        1,
        "the old value is answered once with the one old-name line"
    );
    assert!(
        !String::from_utf8_lossy(&new_out.stderr).contains("old name"),
        "the new value is not the old name"
    );
    assert!(
        old.get("flow_notice").is_none() && new.get("flow_notice").is_none(),
        "the notice never enters the plan object, which project-apply reads back"
    );
    let refused = plan(&root, &with("gitlab-flow"));
    assert!(!refused.status.success(), "an unknown value is refused");
}

/// AC-2 (SPEC REQ-5): a new choice records `next-flow`, and a stored old value
/// survives a read untouched.
#[test]
fn ac2_a_new_choice_records_next_flow_and_a_stored_old_value_survives() {
    let root = TempDir::new().unwrap();
    let planned = ok(&plan(&root, &with("next-flow")));
    ok(&apply(&root, &planned));
    assert!(
        stored_profile(root.path()).contains("next-flow"),
        "a new choice records next-flow"
    );

    let old_root = TempDir::new().unwrap();
    let old_planned = ok(&plan(&old_root, &with("m-flow")));
    ok(&apply(&old_root, &old_planned));
    let before = stored_profile(old_root.path());
    assert!(before.contains("m-flow"));
    // A reader (the preview) must not rewrite the stored value.
    ok(&plan(&old_root, &with("m-flow")));
    let after = stored_profile(old_root.path());
    assert_eq!(
        before, after,
        "reading a profile never rewrites its stored flow value"
    );
}
