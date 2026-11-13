//! The repository's own `REVIEW.md` (#2614 scope 5) is read whole as the root review file by
//! `review-package`: within the byte and link limits, nothing skipped. A later edit past 64
//! counted links or 150,000 bytes would otherwise drop it from every review without a word.

mod common;

use common::git::{command, git, git_ok};
use serde_json::{json, Value};
use std::{fs, path::Path};

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

#[test]
fn this_repository_review_file_is_read_whole() {
    // The file as it is in this checkout, committed into a temporary repository, so the test
    // guards an edit before it is committed and never depends on this repository's history.
    let text = fs::read(Path::new(env!("CARGO_MANIFEST_DIR")).join("REVIEW.md")).unwrap();
    let tmp = tempfile::TempDir::new().unwrap();
    let root = tmp.path();
    git_ok(root, &["init", "-q", "-b", "main"]);
    fs::write(root.join("REVIEW.md"), text).unwrap();
    git_ok(root, &["add", "-A"]);
    let mut commit = command();
    commit
        .args([
            "-c",
            "user.name=F",
            "-c",
            "user.email=f@example.invalid",
            "commit",
            "-q",
            "-m",
            "x",
        ])
        .current_dir(root);
    assert!(commit.status().unwrap().success());
    let head = git(root, &["rev-parse", "HEAD"]);
    let out = std::process::Command::new(EXE)
        .arg("review-package")
        .arg(root)
        .args(["--base", &head, "--head", &head])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    let r = &v["result"];
    let root_files: Vec<(&str, &str)> = r["review"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["role"] != json!("subdirectory"))
        .map(|f| (f["path"].as_str().unwrap(), f["role"].as_str().unwrap()))
        .collect();
    assert_eq!(root_files, [("REVIEW.md", "root")], "{}", r["skipped"]);
    let skipped = r["skipped"].as_array().unwrap();
    assert!(
        !skipped.iter().any(|s| s["path"] == json!("REVIEW.md")),
        "{}",
        r["skipped"]
    );
}
