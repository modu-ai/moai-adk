//! Issue #2666 F2 (audit 6861988023 B1): the `flake-record` CLI checks its input before any
//! request — the 16 MiB limit at its edge, JSON, and `--repo` — so these runs never reach gh.

mod common;
use serde_json::Value;
use std::fs;
use std::process::Command;
use tempfile::TempDir;

const LIMIT: u64 = 16 * 1024 * 1024;

fn run(file: &std::path::Path, repo: &str) -> Value {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["flake-record", "--classified"])
        .arg(file)
        .args([
            "--run",
            "1",
            "--permission-reference",
            "fixture-approved",
            "--repo",
            repo,
        ])
        // A gh that is never called; any call would fail loudly.
        .env("MOAI_FORGE_GH_BIN", "/nonexistent/gh")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap()
}

fn sized(dir: &TempDir, name: &str, len: u64) -> std::path::PathBuf {
    let p = dir.path().join(name);
    let f = fs::File::create(&p).unwrap();
    f.set_len(len).unwrap();
    p
}

#[test]
fn the_input_limit_is_sixteen_mib() {
    let dir = TempDir::new().unwrap();
    let over = run(&sized(&dir, "over.json", LIMIT + 1), "o/r");
    assert_eq!(over["ok"], Value::Bool(false));
    assert!(over.to_string().contains("up to 16 MiB"), "{over}");
    // Exactly the limit passes the size check and stops at JSON.
    let at = run(&sized(&dir, "at.json", LIMIT), "o/r");
    assert_eq!(at["ok"], Value::Bool(false));
    assert!(at.to_string().contains("not JSON"), "{at}");
    assert!(!at.to_string().contains("16 MiB"), "{at}");
}

#[test]
fn a_repo_without_a_slash_is_refused() {
    let dir = TempDir::new().unwrap();
    let p = dir.path().join("c.json");
    fs::write(&p, r#"{"flaky": []}"#).unwrap();
    let v = run(&p, "no-slash");
    assert_eq!(v["ok"], Value::Bool(false));
    assert!(v.to_string().contains("OWNER/REPO"), "{v}");
}
