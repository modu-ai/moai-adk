//! Issue #2054: `baseline-inventory` reads PROJECT even when the caller's environment points
//! git at another repository (`GIT_DIR`, `GIT_WORK_TREE`, injected config).
mod common;
use serde_json::Value;
use std::{fs, path::Path, process::Command};

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .arg("-C")
        .arg(dir)
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .output()
        .unwrap();
    assert!(out.status.success(), "git {args:?}");
    String::from_utf8(out.stdout).unwrap().trim().to_string()
}

/// A repository with one commit whose content is `name`; returns its HEAD.
fn repo(dir: &Path, name: &str) -> String {
    git(dir, &["init", "-q", "-b", "main"]);
    fs::write(dir.join("file.txt"), name).unwrap();
    git(dir, &["add", "file.txt"]);
    git(dir, &["commit", "-q", "-m", name]);
    git(dir, &["rev-parse", "HEAD"])
}

fn inventory(project: &Path, env: &[(&str, &Path)]) -> (std::process::Output, Value) {
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.arg("baseline-inventory").arg(project);
    for (k, v) in env {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    let v = serde_json::from_slice::<Value>(&out.stdout)
        .map(|v| v["result"].clone())
        .unwrap_or(Value::Null);
    (out, v)
}

#[test]
fn baseline_reads_project_despite_injected_git_dir() {
    let project = tempfile::tempdir().unwrap();
    let other = tempfile::tempdir().unwrap();
    let project_head = repo(project.path(), "project");
    let other_head = repo(other.path(), "other");
    assert_ne!(project_head, other_head);

    let other_git = other.path().join(".git");
    let (out, v) = inventory(
        project.path(),
        &[("GIT_DIR", &other_git), ("GIT_WORK_TREE", other.path())],
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert_eq!(v["baseline"]["head_sha"], Value::from(project_head));
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(!text.contains(&other_head), "read the injected repository");
}

fn error_kind(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["error"]["kind"].clone()
}

#[test]
fn baseline_git_failures_follow_the_identity_rule() {
    // A git failure other than "not a git repository" is `git_failed` (#2033 rule).
    let bare = tempfile::tempdir().unwrap();
    git(bare.path(), &["init", "-q", "--bare"]);
    let (out, _) = inventory(bare.path(), &[]);
    assert!(!out.status.success());
    assert_eq!(error_kind(&out), Value::from("git_failed"));
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        !text.contains(bare.path().to_str().unwrap()),
        "path reached the output"
    );

    // Not a repository keeps its existing kind.
    let plain = tempfile::tempdir().unwrap();
    let (out, _) = inventory(plain.path(), &[]);
    assert!(!out.status.success());
    assert_eq!(error_kind(&out), Value::from("not_found"));
}

/// #2811 AC-3 (D4): a worktree in the new home place sits outside the project, so the
/// baseline names it `<outside-project>` — never its absolute path — and the old-place one
/// keeps its relative path. Temporary HOME only.
#[test]
fn a_home_place_worktree_reads_outside_project() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let project = root.join("repo");
    fs::create_dir_all(&project).unwrap();
    repo(&project, "p");
    fs::write(project.join(".gitignore"), ".moai-cli/\n").unwrap();
    let home = root.join("home");
    let new_wt = home.join(".moai/worktrees/repo-1-x");
    let old_wt = project.join(".moai-cli/worktrees/1-y");
    git(
        &project,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1-x",
            new_wt.to_str().unwrap(),
        ],
    );
    git(
        &project,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1-y",
            old_wt.to_str().unwrap(),
        ],
    );
    let (out, v) = inventory(&project, &[("HOME", &home)]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let paths: Vec<&str> = v["baseline"]["worktrees"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w["path"].as_str().unwrap())
        .collect();
    assert_eq!(paths, [".", ".moai-cli/worktrees/1-y", "<outside-project>"]);
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        !text.contains(home.to_str().unwrap()),
        "absolute home path leaked"
    );
}
