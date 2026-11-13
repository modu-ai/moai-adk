//! #2723: `moai-cli home-state-apply` / `home-state-revert` on temporary HOME folders only.
//! The real HOME is never read or written: every case passes `--home <tempdir>`.
#![cfg(unix)]

use serde_json::{json, Value};
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");
const NOW: &str = "2026-11-13T19:00:00Z";
const LATER: &str = "2026-11-13T19:30:00Z";

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

fn go_home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    write(
        &home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(&home, ".moai/state/active-sessions.json", "[]");
    write(&home, ".moai/credentials.yaml", "token: PLACEHOLDER\n");
    write(&home, ".moai/config/participation.yaml", "x: 1\n");
    write(&home, ".moai/memories/a/note.md", "remember\n");
    write(&home, ".moai/logs/run.log", "log\n");
    (dir, home)
}

fn tree(root: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out = Vec::new();
    if !root.exists() {
        return out;
    }
    let mut stack = vec![root.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(root).unwrap().display().to_string();
            if fs::symlink_metadata(&p).unwrap().is_dir() {
                out.push((format!("{rel}/"), Vec::new()));
                stack.push(p);
            } else {
                out.push((rel, fs::read(&p).unwrap_or_default()));
            }
        }
    }
    out.sort();
    out
}

fn call(args: &[&str]) -> Value {
    let out = Command::new(BIN)
        .args(args)
        .env_remove("MOAI_HOME")
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan(home: &Path) -> Value {
    let v = call(&[
        "home-state-plan",
        "--home",
        home.to_str().unwrap(),
        "--now",
        NOW,
    ]);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn apply(home: &Path, digest: &str, perm: &str, copy: &[&str]) -> Value {
    let mut args = vec![
        "home-state-apply",
        "--home",
        home.to_str().unwrap(),
        "--now",
        NOW,
        "--plan-digest",
        digest,
        "--permission-reference",
        perm,
    ];
    for c in copy {
        args.extend(["--copy", c]);
    }
    call(&args)
}

fn revert(home: &Path, perm: &str) -> Value {
    call(&[
        "home-state-revert",
        "--home",
        home.to_str().unwrap(),
        "--now",
        LATER,
        "--permission-reference",
        perm,
    ])
}

#[test]
fn apply_moves_aside_copies_the_selection_and_rerun_is_none_then_revert_restores() {
    let (_d, home) = go_home();
    let original = tree(&home.join(".moai"));
    let p = plan(&home);
    assert_eq!(p["action"], json!("migrate"), "{p}");
    let a = apply(
        &home,
        p["plan_digest"].as_str().unwrap(),
        "user approved in the fixture",
        &["config", "memories"],
    );
    assert_eq!(a["ok"], json!(true), "{a}");
    assert_eq!(a["result"]["state"], json!("applied"));
    // The Go folder is intact under its backup name.
    assert_eq!(tree(&home.join(".moai-backup")), original);
    // The Rust folder holds the selection, the receipt and the marker, nothing else.
    let rust = tree(&home.join(".moai"));
    let names: Vec<&str> = rust.iter().map(|(n, _)| n.as_str()).collect();
    assert_eq!(
        names,
        [
            "config/",
            "config/participation.yaml",
            "home-state-receipt.json",
            "memories/",
            "memories/a/",
            "memories/a/note.md",
            "moai-cli.json"
        ]
    );
    let receipt: Value =
        serde_json::from_slice(&fs::read(home.join(".moai/home-state-receipt.json")).unwrap())
            .unwrap();
    assert_eq!(receipt["copied"], json!(["config", "memories"]));
    let again = plan(&home);
    assert_eq!(again["state"]["generation"], json!("rust"), "{again}");
    assert_eq!(again["action"], json!("none"));

    // Revert: the Rust folder moves aside, the Go folder comes back byte for byte.
    let r = revert(&home, "user approved the revert");
    assert_eq!(r["ok"], json!(true), "{r}");
    assert_eq!(tree(&home.join(".moai")), original);
    assert!(!home.join(".moai-backup").exists());
    let aside = home.join(".moai.rust-20261113T193000Z");
    assert!(aside.join("moai-cli.json").is_file());
    assert_eq!(plan(&home)["state"]["generation"], json!("go"));
    // A second revert finds no Rust folder and moves nothing.
    let r2 = revert(&home, "again");
    assert_eq!(r2["ok"], json!(false), "{r2}");
    assert_eq!(tree(&home.join(".moai")), original);
}

#[test]
fn every_refusal_writes_nothing() {
    let (_d, home) = go_home();
    let before = tree(&home);
    let p = plan(&home);
    let d = p["plan_digest"].as_str().unwrap().to_string();
    for (copy, perm, digest, why) in [
        (vec![], "", d.as_str(), "permission"),
        (vec![], "approved", "sha256:00", "stale or altered plan"),
        (vec!["credentials.yaml"], "approved", d.as_str(), "secrets"),
        (
            vec!["state"],
            "approved",
            d.as_str(),
            "marks the Go generation",
        ),
        (
            vec!["user-assets.json"],
            "approved",
            d.as_str(),
            "marks the Go generation",
        ),
        (vec!["nope"], "approved", d.as_str(), "not a copy candidate"),
    ] {
        let a = apply(&home, digest, perm, &copy);
        assert_eq!(a["ok"], json!(false), "{why}: {a}");
        assert!(a.to_string().contains(why), "{why}: {a}");
        assert_eq!(tree(&home), before, "{why}");
    }
    // A link inside the selection refuses the whole apply before anything moves.
    std::os::unix::fs::symlink("/etc", home.join(".moai/config/outside")).unwrap();
    let before = tree(&home);
    let p = plan(&home);
    let a = apply(
        &home,
        p["plan_digest"].as_str().unwrap(),
        "approved",
        &["config"],
    );
    assert!(a.to_string().contains("never copied"), "{a}");
    assert_eq!(tree(&home), before);
    fs::remove_file(home.join(".moai/config/outside")).unwrap();
    // A blocked preview is never applied.
    write(&home, ".moai/worktrees/x/.git", "gitdir: /elsewhere\n");
    let p = plan(&home);
    assert_eq!(p["action"], json!("blocked"));
    let before = tree(&home);
    let a = apply(&home, p["plan_digest"].as_str().unwrap(), "approved", &[]);
    assert!(a.to_string().contains("not `migrate`"), "{a}");
    assert_eq!(tree(&home), before);
}

#[test]
fn a_failure_after_the_rename_puts_the_go_folder_back() {
    let (_d, home) = go_home();
    let secret_log = home.join(".moai/logs/run.log");
    fs::set_permissions(&secret_log, fs::Permissions::from_mode(0o000)).unwrap();
    if fs::read(&secret_log).is_ok() {
        // Running as a user that reads anything (root): the failure cannot be staged.
        fs::set_permissions(&secret_log, fs::Permissions::from_mode(0o644)).unwrap();
        return;
    }
    let p = plan(&home);
    let a = apply(
        &home,
        p["plan_digest"].as_str().unwrap(),
        "approved",
        &["logs"],
    );
    fs::set_permissions(
        home.join(".moai/logs/run.log"),
        fs::Permissions::from_mode(0o644),
    )
    .unwrap();
    assert_eq!(a["ok"], json!(false), "{a}");
    assert!(a.to_string().contains("is back"), "{a}");
    assert_eq!(plan(&home)["state"]["generation"], json!("go"));
    assert!(!home.join(".moai-backup").exists());
    let partial = home.join(".moai.partial-20261113T190000Z");
    assert!(
        partial.is_dir(),
        "the partial Rust folder is kept, not deleted"
    );
    assert!(
        !partial.join("moai-cli.json").exists(),
        "never marked finished"
    );
}

#[test]
fn revert_refuses_without_a_receipt_backup() {
    let (_d, home) = go_home();
    let p = plan(&home);
    let a = apply(&home, p["plan_digest"].as_str().unwrap(), "approved", &[]);
    assert_eq!(a["ok"], json!(true), "{a}");
    // The aside name is taken: nothing moves.
    fs::create_dir(home.join(".moai.rust-20261113T193000Z")).unwrap();
    let before = tree(&home);
    let r = revert(&home, "approved");
    assert_eq!(r["ok"], json!(false), "{r}");
    assert_eq!(tree(&home), before);
    fs::remove_dir(home.join(".moai.rust-20261113T193000Z")).unwrap();
    // A receipt that points outside HOME is refused.
    let receipt = home.join(".moai/home-state-receipt.json");
    let mut v: Value = serde_json::from_slice(&fs::read(&receipt).unwrap()).unwrap();
    v["backup"] = json!("/tmp/.moai-backup");
    fs::write(&receipt, v.to_string()).unwrap();
    let before = tree(&home);
    let r = revert(&home, "approved");
    assert!(r.to_string().contains("beside ~/.moai"), "{r}");
    assert_eq!(tree(&home), before);
    // No permission reference: refused.
    assert_eq!(revert(&home, "")["ok"], json!(false));
}
