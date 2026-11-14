//! Product CLI tests of the Project group (#2834 R4a): one file per command group, so a pull
//! request that adds or changes a Project command touches this file, not `tests/product_cli.rs`.
//! These pin the command-line binding that no other test read when the commands moved to
//! `src/ops/`: the operation name in the envelope, the refusal kind and message, and the exit code.

mod common;

use common::cli::cli;
use serde_json::{json, Value};
use tempfile::TempDir;

/// Asserts a refusal: the envelope names `operation`, carries `kind` and `message`, and the
/// command exits 1.
fn refusal(out: &std::process::Output, operation: &str, kind: &str, message: &str) {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["operation"], json!(operation), "{v}");
    assert_eq!(v["ok"], json!(false), "{v}");
    assert_eq!(v["error"]["kind"], json!(kind), "{v}");
    assert_eq!(v["error"]["message"], json!(message), "{v}");
    assert_eq!(out.status.code(), Some(1), "{v}");
}

#[test]
fn project_commands_name_their_operation_on_a_refusal() {
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    let missing = root.path().join("none");
    refusal(
        &cli(&["project-detect", missing.to_str().unwrap()], &json!(null)),
        "project.detect",
        "not_found",
        "project directory unavailable",
    );
    // `{}` lacks the required request fields.
    for (word, operation) in [
        ("generation-plan", "generation.plan"),
        ("github-plan", "github.plan"),
    ] {
        refusal(
            &cli(&[word, p], &json!({})),
            operation,
            "invalid_input",
            "invalid typed JSON input",
        );
    }
}

#[test]
fn language_setup_answers_must_be_an_object() {
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    refusal(
        &cli(&["language-setup", p, "--answers"], &json!([])),
        "language.setup",
        "invalid_input",
        "answers must be a JSON object",
    );
}

/// The empty permission reference is refused before stdin is read: with a stdin that is not
/// JSON, the answer is still the permission refusal, not a parse error (the library checks
/// the reference too, but only after reading the plan).
#[test]
fn labels_apply_refuses_an_empty_permission_reference_before_reading_the_plan() {
    use std::io::Write as _;
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    let mut child = std::process::Command::new(common::MOAI_BIN)
        .args([
            "labels-apply",
            p,
            "--plan-digest",
            "d",
            "--permission-reference",
            "",
        ])
        .env("MOAI_FORGE_GH_BIN", "/usr/bin/false")
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .unwrap();
    match child.stdin.take().unwrap().write_all(b"not json") {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    refusal(
        &child.wait_with_output().unwrap(),
        "labels.apply",
        "permission_reference_missing",
        "permission reference required",
    );
}
