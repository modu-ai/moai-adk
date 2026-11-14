//! Product CLI tests of the Internal group (#2834 R2a): one file per command group, so a pull
//! request that adds or changes a command touches its own group file, not
//! `tests/product_cli.rs`. These pin the command-line binding of commands whose behaviour is
//! tested through library calls only (`project-skill-*`, `codex-projection-*`): the operation
//! name in the envelope, the error kind and message, and the exit code.

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
fn project_skill_commands_read_a_typed_request_and_name_their_operation() {
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    for (word, operation) in [
        ("project-skill-plan", "project_skill.plan"),
        ("project-skill-apply", "project_skill.apply"),
        ("project-skill-revert", "project_skill.revert"),
    ] {
        // `{}` lacks the required request fields; a JSON string is the wrong shape.
        for input in [json!({}), json!("nope")] {
            let out = cli(&[word, p], &input);
            refusal(&out, operation, "invalid_input", "invalid typed JSON input");
        }
    }
}

#[test]
fn codex_projection_commands_name_their_operation() {
    let root = TempDir::new().unwrap();
    let pkg = root.path().join("pkg");
    let target = root.path().join("moai");
    let (pkg, target) = (pkg.to_str().unwrap(), target.to_str().unwrap());
    let out = cli(
        &["codex-projection-plan", pkg, "--target", target],
        &Value::Null,
    );
    refusal(
        &out,
        "codex.projection_plan",
        "not_found",
        "package unavailable",
    );
    let out = cli(
        &[
            "codex-projection-apply",
            pkg,
            "--target",
            target,
            "--plan-digest",
            "x",
            "--permission-reference",
            "r",
        ],
        &Value::Null,
    );
    refusal(
        &out,
        "codex.projection_apply",
        "not_found",
        "package unavailable",
    );
}
