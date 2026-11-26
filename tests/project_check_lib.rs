//! The generic local check of a project (SPEC-20261122-project-ci-design REQ-5, slice D): it
//! runs the commands the project profile declares, in the declared order, in the project folder.
//! The library function behind the later operation is driven here in temporary folders; no
//! binary runs, so no HOME is involved.
//!
//! A project with no declared commands gets a plain explanation, not a green result; every
//! step runs even when an earlier one failed; a command that hangs is stopped at the limit.

use moai_cli::project_check::{compare, declared, run, workflow_run_steps, Options};
use moai_cli::project_ci::CheckEntry;
use serde_json::{json, Value};
use std::fs;
use tempfile::TempDir;

fn profile(checks: Value) -> Value {
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "checks": checks
    })
}

fn project(checks: Value) -> TempDir {
    let root = TempDir::new().unwrap();
    let path = root.path().join(".moai-cli/project.yaml");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, serde_json::to_vec_pretty(&profile(checks)).unwrap()).unwrap();
    root
}

fn entry(id: &str, command: &str) -> CheckEntry {
    CheckEntry {
        id: id.into(),
        command: command.into(),
    }
}

fn options() -> Options {
    Options { timeout_secs: 20 }
}

fn exits(report: &Value) -> Vec<(String, Option<i64>)> {
    report["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| (s["id"].as_str().unwrap().to_string(), s["exit"].as_i64()))
        .collect()
}

/// REQ-5: the declared commands run in the declared order, each in the project folder, and the
/// answer names every step with its exit status and ends with one summary.
#[test]
fn the_declared_commands_run_in_the_declared_order_in_the_project_folder() {
    let root = project(json!([
        {"id": "one", "command": "printf 1 >> order.txt && pwd > where.txt"},
        {"id": "two", "command": "printf 2 >> order.txt"},
        {"id": "three", "command": "printf 3 >> order.txt"}
    ]));
    let report = run(root.path(), &options()).unwrap();
    assert_eq!(report["result"], "PASS", "{report}");
    assert_eq!(
        exits(&report),
        vec![
            ("one".to_string(), Some(0)),
            ("two".to_string(), Some(0)),
            ("three".to_string(), Some(0))
        ]
    );
    assert_eq!(
        fs::read_to_string(root.path().join("order.txt")).unwrap(),
        "123"
    );
    let where_ = fs::read_to_string(root.path().join("where.txt")).unwrap();
    assert_eq!(
        fs::canonicalize(where_.trim()).unwrap(),
        fs::canonicalize(root.path()).unwrap()
    );
    let summary = report["summary"].as_str().unwrap();
    assert!(summary.starts_with("project-check "), "{summary}");
    assert!(summary.contains("one=0 two=0 three=0"), "{summary}");
    assert!(summary.ends_with("result=PASS"), "{summary}");
}

/// A failing step does not stop the rest, and the result is FAIL with the failing status shown.
#[test]
fn a_failing_step_does_not_stop_the_rest_and_the_result_is_fail() {
    let root = project(json!([
        {"id": "build", "command": "exit 0"},
        {"id": "lint", "command": "echo broken >&2; exit 3"},
        {"id": "test", "command": "touch ran-after"}
    ]));
    let report = run(root.path(), &options()).unwrap();
    assert_eq!(report["result"], "FAIL", "{report}");
    assert_eq!(
        exits(&report),
        vec![
            ("build".to_string(), Some(0)),
            ("lint".to_string(), Some(3)),
            ("test".to_string(), Some(0))
        ]
    );
    assert!(root.path().join("ran-after").exists());
    assert!(report["steps"][1]["output_tail"]
        .as_str()
        .unwrap()
        .contains("broken"));
    assert!(report["summary"].as_str().unwrap().ends_with("result=FAIL"));
}

/// A project with no declared commands, or none saved at all, is explained, never reported
/// green.
#[test]
fn a_project_without_declared_commands_is_explained_not_passed() {
    // No list saved at all: the plain explanation.
    let root = project(Value::Null);
    let err = run(root.path(), &options()).unwrap_err();
    assert!(err.contains("no declared checks"), "{err}");
    assert!(declared(root.path()).is_err());
    // An empty list is refused when the profile is read, so it never runs either.
    let root = project(json!([]));
    assert!(run(root.path(), &options()).is_err());
    assert!(declared(root.path()).is_err());
    let empty = TempDir::new().unwrap();
    assert!(run(empty.path(), &options()).is_err());
}

/// A command that never ends is stopped at the limit and counted as a failure, and the steps
/// after it still run.
#[test]
fn a_command_that_hangs_is_stopped_at_the_limit() {
    let root = project(json!([
        {"id": "hang", "command": "sleep 30"},
        {"id": "after", "command": "touch after"}
    ]));
    let started = std::time::Instant::now();
    let report = run(root.path(), &Options { timeout_secs: 1 }).unwrap();
    assert!(
        started.elapsed().as_secs() < 20,
        "the limit did not stop it"
    );
    assert_eq!(report["result"], "FAIL", "{report}");
    assert_eq!(report["steps"][0]["timed_out"], true, "{report}");
    assert_eq!(report["steps"][0]["exit"], Value::Null, "{report}");
    assert!(root.path().join("after").exists());
}

/// The run does not read git's repository variables from the caller's environment: a check
/// that runs `git` sees the project's own repository.
#[test]
fn the_steps_see_the_project_repository_not_the_callers() {
    let root = project(
        json!([{"id": "where", "command": "printf '%s' \"${GIT_DIR-unset}\" > git_dir.txt"}]),
    );
    std::env::set_var("GIT_DIR", "/nonexistent");
    let report = run(root.path(), &options()).unwrap();
    std::env::remove_var("GIT_DIR");
    assert_eq!(report["result"], "PASS", "{report}");
    assert_eq!(
        fs::read_to_string(root.path().join("git_dir.txt")).unwrap(),
        "unset"
    );
}

/// `declared` is the profile's list as saved, in order.
#[test]
fn declared_returns_the_profile_list_in_order() {
    let root = project(json!([
        {"id": "b", "command": "cargo build"},
        {"id": "a", "command": "cargo test"}
    ]));
    assert_eq!(
        declared(root.path()).unwrap(),
        vec![entry("b", "cargo build"), entry("a", "cargo test")]
    );
}

const WORKFLOW: &str = "\
# a
# b
# c

name: \"MoAI check\"
on:
  push:
jobs:
  check:
    name: \"check\"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@abc # v1
        with:
          persist-credentials: false
      - name: \"build\"
        run: |
          cargo build
      - name: \"test\"
        run: |
          cargo test
  build-windows:
    name: \"build-windows\"
    runs-on: windows-latest
    steps:
      - name: \"build\"
        run: |
          cargo build
";

/// REQ-4: the run steps of the `check` job are read as the commands in order; another job's
/// steps are not part of the list.
#[test]
fn the_run_steps_of_the_check_job_are_read_in_order() {
    assert_eq!(
        workflow_run_steps(WORKFLOW),
        vec!["cargo build".to_string(), "cargo test".to_string()]
    );
    assert_eq!(workflow_run_steps("name: x\n"), Vec::<String>::new());
}

/// REQ-4: a declared command the workflow lacks, an undeclared command in a run step and a
/// different order are each a finding; an equal list is none.
#[test]
fn the_comparison_names_a_missing_an_extra_and_a_reordered_command() {
    let both = [entry("build", "cargo build"), entry("test", "cargo test")];
    assert_eq!(compare(&both, WORKFLOW), Vec::<String>::new());
    let more = [
        entry("build", "cargo build"),
        entry("test", "cargo test"),
        entry("lint", "cargo clippy"),
    ];
    let findings = compare(&more, WORKFLOW);
    assert!(
        findings
            .iter()
            .any(|f| f.contains("cargo clippy") && f.contains("missing")),
        "{findings:?}"
    );
    let fewer = [entry("build", "cargo build")];
    let findings = compare(&fewer, WORKFLOW);
    assert!(
        findings
            .iter()
            .any(|f| f.contains("cargo test") && f.contains("not declared")),
        "{findings:?}"
    );
    let swapped = [entry("test", "cargo test"), entry("build", "cargo build")];
    let findings = compare(&swapped, WORKFLOW);
    assert!(findings.iter().any(|f| f.contains("order")), "{findings:?}");
    // A file with no run step at all, against declared commands, misses every one.
    let findings = compare(&both, "name: x\n");
    assert_eq!(findings.len(), 2, "{findings:?}");
}

#[test]
fn a_run_that_cannot_start_reports_a_failure_not_a_pass() {
    let root = project(json!([{"id": "gone", "command": "/nonexistent/binary-xyz"}]));
    let report = run(root.path(), &options()).unwrap();
    assert_eq!(report["result"], "FAIL", "{report}");
    assert!(
        report["steps"][0]["exit"].as_i64().unwrap() != 0,
        "{report}"
    );
}
