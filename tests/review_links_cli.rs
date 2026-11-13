//! #2614 AC-4 record 6841710561: a REVIEW.md past 64 counted links is skipped whole however its
//! links are spelled — a space before the destination (found by Codex CLI in the two-tool
//! comparison), a destination in `<…>`, or reference definitions. Temporary repositories only.

mod common;

use common::git::{command, git, git_ok};
use serde_json::{json, Value};
use std::{fs, path::Path};

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn package_with_review(text: &str) -> Value {
    let tmp = tempfile::TempDir::new().unwrap();
    let repo = tmp.path();
    git_ok(repo, &["init", "-q", "-b", "main"]);
    fs::write(repo.join("REVIEW.md"), text).unwrap();
    fs::write(repo.join("a.txt"), "a\n").unwrap();
    let commit = |m: &str| {
        git_ok(repo, &["add", "-A"]);
        let mut c = command();
        c.args([
            "-c",
            "user.name=F",
            "-c",
            "user.email=f@example.invalid",
            "commit",
            "-q",
            "-m",
            m,
        ])
        .current_dir(repo);
        assert!(c.status().unwrap().success());
        git(repo, &["rev-parse", "HEAD"])
    };
    let base = commit("base");
    fs::write(repo.join("a.txt"), "b\n").unwrap();
    let head = commit("head");
    let out = std::process::Command::new(EXE)
        .arg("review-package")
        .arg(repo as &Path)
        .args(["--base", &base, "--head", &head])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn root_read(r: &Value) -> bool {
    r["review"]
        .as_array()
        .unwrap()
        .iter()
        .any(|f| f["path"] == json!("REVIEW.md"))
}

#[test]
fn every_link_spelling_counts_toward_the_limit() {
    for (name, link) in [
        ("space", "[l{i}]( sub/l{i}.md)\n"),
        ("angle", "[l{i}](<sub/l{i}.md>)\n"),
        ("reference", "[l{i}]: sub/l{i}.md\n"),
    ] {
        let at: String = (0..64)
            .map(|i| link.replace("{i}", &i.to_string()))
            .collect();
        let over: String = (0..65)
            .map(|i| link.replace("{i}", &i.to_string()))
            .collect();
        assert!(
            root_read(&package_with_review(&at)),
            "{name}: 64 links must be read"
        );
        let r = package_with_review(&over);
        assert!(
            !root_read(&r),
            "{name}: 65 links must skip the root file: {r}"
        );
        assert_eq!(
            r["skipped"],
            json!([{"path": "REVIEW.md", "reason": "too-many-links"}]),
            "{name}"
        );
    }
}
