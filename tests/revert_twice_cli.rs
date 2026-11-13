//! #2727: reusing a receipt that was already reverted is refused as `conflict` — a refusal on
//! the target's state that automation can tell apart from a real `internal` error — with exit
//! 1 and nothing changed. Temporary HOME only; the real HOME is never read or written.

mod common;
use common::CleanEnv;
use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const PERM: &str = "fixture-approved";

fn call(args: &[&str]) -> (i32, Value) {
    let out = Command::new(EXE)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{args:?}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code().unwrap(), v)
}

fn home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap().join("home");
    fs::create_dir_all(&path).unwrap();
    (dir, path)
}

fn bytes(paths: &[PathBuf]) -> Vec<Option<Vec<u8>>> {
    paths.iter().map(|p| fs::read(p).ok()).collect()
}

fn assert_conflict(code: i32, v: &Value, message: &str) {
    assert_eq!(code, 1, "{v}");
    assert_eq!(v["ok"], Value::Bool(false), "{v}");
    assert_eq!(v["error"]["kind"], "conflict", "{v}");
    assert_eq!(v["error"]["retryable"], Value::Bool(false), "{v}");
    assert_eq!(v["error"]["message"], message, "{v}");
}

/// The routing section: apply, revert, then the same receipt again.
#[test]
fn a_second_routing_revert_is_a_conflict_and_changes_nothing() {
    let (_t, home) = home();
    let h = home.to_str().unwrap();
    let package = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai");
    let pkg = package.to_str().unwrap();
    let (code, plan) = call(&["home-routing-plan", h, pkg]);
    assert_eq!(code, 0, "{plan}");
    let digest = plan["result"]["plan_digest"].as_str().unwrap();
    let (code, applied) = call(&[
        "home-routing-apply",
        h,
        pkg,
        "--plan-digest",
        digest,
        "--permission-reference",
        PERM,
    ]);
    assert_eq!(code, 0, "{applied}");
    let id = applied["result"]["receipt"].as_str().unwrap().to_string();
    let revert = [
        "home-routing-revert",
        h,
        "--receipt",
        &id,
        "--permission-reference",
        PERM,
    ];
    let (code, first) = call(&revert);
    assert_eq!(code, 0, "{first}");
    let watched = [
        home.join(".claude/CLAUDE.md"),
        home.join(".codex/AGENTS.md"),
        home.join(format!(".moai-cli/home-routing/receipts/{id}/receipt.json")),
    ];
    assert!(watched[2].is_file(), "receipt path");
    let before = bytes(&watched);
    let (code, second) = call(&revert);
    assert_conflict(code, &second, "routing receipt already reverted");
    assert_eq!(bytes(&watched), before);
}

/// The base agents: a receipt in the `reverted` state is refused the same way. The receipt is
/// written in place so the test does not depend on the install arguments.
#[test]
fn a_second_agents_revert_is_a_conflict_and_changes_nothing() {
    let (_t, home) = home();
    let id = "agents-1-0";
    let dir = home.join(".moai-cli/agents/receipts").join(id);
    fs::create_dir_all(&dir).unwrap();
    let receipt = dir.join("receipt.json");
    fs::write(
        &receipt,
        serde_json::json!({
            "schema_version": 1,
            "id": id,
            "plan_digest": "sha256:0",
            "permission_reference": PERM,
            "status": "reverted",
            "entries": []
        })
        .to_string(),
    )
    .unwrap();
    let before = bytes(std::slice::from_ref(&receipt));
    let (code, v) = call(&[
        "agents-home-revert",
        home.to_str().unwrap(),
        "--receipt",
        id,
        "--permission-reference",
        PERM,
    ]);
    assert_conflict(code, &v, "agent receipt already reverted");
    assert_eq!(bytes(&[receipt]), before);
}
