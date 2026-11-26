//! `project-check` (SPEC-20261122-project-ci-design REQ-5): the local check of a project, run
//! through the binary in temporary folders with a temporary HOME of its own. The library is
//! tested in `project_check_lib.rs`; here the answer's shape and the exit behaviour are.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{fs, process::Output};
use tempfile::TempDir;

fn project(checks: Value) -> TempDir {
    let root = TempDir::new().unwrap();
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "checks": checks
    });
    let path = root.path().join(".moai-cli/project.yaml");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, serde_json::to_vec_pretty(&profile).unwrap()).unwrap();
    root
}

fn run(args: &[&str]) -> Output {
    TempHome::new().command(args).output().unwrap()
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

/// The answer names every step, ends with one summary line and says PASS.
#[test]
fn the_check_runs_the_declared_commands_and_answers_in_one_summary() {
    let root = project(json!([
        {"id": "build", "command": "echo one >> out.txt"},
        {"id": "test", "command": "echo two >> out.txt"}
    ]));
    let out = run(&["project-check", root.path().to_str().unwrap()]);
    let v = envelope(&out);
    assert_eq!(v["operation"], "project.check", "{v}");
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["result"], "PASS", "{v}");
    assert_eq!(
        v["result"]["summary"], "project-check build=0 test=0 result=PASS",
        "{v}"
    );
    assert_eq!(
        fs::read_to_string(root.path().join("out.txt")).unwrap(),
        "one\ntwo\n"
    );
}

/// A failing step makes the answer FAIL, as `check affected` answers: the envelope is ok and the
/// summary line carries the result; every step still ran.
#[test]
fn a_failing_step_is_a_fail_answer_and_every_step_still_ran() {
    let root = project(json!([
        {"id": "lint", "command": "exit 4"},
        {"id": "test", "command": "touch ran"}
    ]));
    let out = run(&["project-check", root.path().to_str().unwrap()]);
    let v = envelope(&out);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["result"]["result"], "FAIL", "{v}");
    assert_eq!(
        v["result"]["summary"], "project-check lint=4 test=0 result=FAIL",
        "{v}"
    );
    assert!(root.path().join("ran").exists());
}

/// A project with no declared checks is refused with the plain explanation and the operation
/// named; nothing is run.
#[test]
fn a_project_without_declared_checks_is_refused_in_plain_words() {
    let root = project(Value::Null);
    let out = run(&["project-check", root.path().to_str().unwrap()]);
    let v = envelope(&out);
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(v["operation"], "project.check", "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("no declared checks"),
        "{v}"
    );
}

/// The per-command limit is an option, and a command that hangs is stopped by it.
#[test]
fn the_limit_option_stops_a_hanging_command() {
    let root = project(json!([{"id": "hang", "command": "sleep 30"}]));
    let started = std::time::Instant::now();
    let out = run(&[
        "project-check",
        root.path().to_str().unwrap(),
        "--timeout-secs",
        "1",
    ]);
    assert!(started.elapsed().as_secs() < 20);
    let v = envelope(&out);
    assert_eq!(v["result"]["steps"][0]["timed_out"], true, "{v}");
    assert_eq!(v["result"]["result"], "FAIL", "{v}");
}
