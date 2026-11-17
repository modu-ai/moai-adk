//! SPEC-20261117-commit-guard-users: the `issue/<n>` commit guard is a lane safety. Both
//! installed guards — `hook-dispatch --event tool-preflight` and `claim-verify . --hook` — refuse
//! a commit without `Moai-Claim` on an `issue/<n>` branch only in a clone with a live MoAI lane
//! session (its session registry, `<clone common dir>/.moai-cli/sessions/`). An ordinary user's
//! commit there is allowed; `spec/` branches are not lane branches either way. Every case runs
//! in a temporary repository with its own HOME.
#![cfg(unix)]
mod common;

use serde_json::json;
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};

const BIN: &str = crate::common::MOAI_BIN;

fn git(root: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(root)
        .env("HOME", root.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A repository on `branch` with one commit, its own HOME and a GitHub origin.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", branch]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    git(
        &root,
        &[
            "remote",
            "add",
            "origin",
            "https://github.com/acme/widgets.git",
        ],
    );
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    git(&root, &["add", ".gitignore"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    (dir, root)
}

fn moai(root: &Path, args: &[&str], stdin: &[u8]) -> Output {
    let mut child = Command::new(BIN)
        .args(args)
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("MOAI_FORGE_GH_BIN")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child.stdin.take().unwrap().write_all(stdin).unwrap();
    child.wait_with_output().unwrap()
}

/// A lane session of this clone, registered the way MODE-AGENT does at start (`--now` moves the
/// lease into the past for the expired case).
fn register_lane(root: &Path, now: Option<&str>) {
    let mut args = vec![
        "session-register",
        root.to_str().unwrap(),
        "--name",
        "agent-1",
    ];
    if let Some(now) = now {
        args.extend(["--now", now]);
    }
    let out = moai(root, &args, b"");
    assert!(
        out.status.success(),
        "session-register: {}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
}

fn input(command: &str, root: &Path) -> Vec<u8> {
    json!({"session_id": "s", "hook_event_name": "PreToolUse", "tool_name": "Bash",
           "tool_use_id": "t1", "tool_input": {"command": command}, "cwd": root})
    .to_string()
    .into_bytes()
}

/// The two installed guards' answers to one Bash command: (preflight denied, claim guard exit).
fn guards(root: &Path, command: &str) -> (bool, Option<i32>) {
    let pre = moai(
        root,
        &[
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ],
        &input(command, root),
    );
    assert_eq!(
        pre.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&pre.stderr)
    );
    let stdout = String::from_utf8_lossy(&pre.stdout);
    let denied = stdout.contains("\"permissionDecision\":\"deny\"");
    assert!(denied || stdout.trim().is_empty(), "{stdout}");
    let claim = moai(
        root,
        &["claim-verify", ".", "--hook"],
        &input(command, root),
    );
    (denied, claim.status.code())
}

const PLAIN: &str = "git commit -m work";

#[test]
fn an_ordinary_user_commits_on_an_issue_branch() {
    let (_d, root) = repo("issue/12-fix-typo");
    assert_eq!(guards(&root, PLAIN), (false, Some(0)));
    // The commit itself goes through as well.
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", "a.txt"]);
    git(&root, &["commit", "-q", "-m", "work"]);
}

#[test]
fn a_lane_session_without_a_claim_is_refused() {
    let (_d, root) = repo("issue/12-fix-typo");
    register_lane(&root, None);
    assert_eq!(guards(&root, PLAIN), (true, Some(2)));
}

#[test]
fn a_lane_session_with_its_claim_token_passes_the_preflight() {
    let (_d, root) = repo("issue/12-fix-typo");
    register_lane(&root, None);
    let (denied, _) = guards(&root, "git commit -m \"work\n\nMoai-Claim: 12#100\"");
    // The claim guard then verifies the token against GitHub (not reached offline here); the
    // library test `hook_guards_a_lane_branch_without_a_claim_record` covers that path.
    assert!(!denied);
}

/// The lease is the boundary: an entry whose lease has ended is no live lane.
#[test]
fn an_ended_lane_lease_leaves_ordinary_commits_alone() {
    let (_d, root) = repo("issue/12-fix-typo");
    register_lane(&root, Some("2020-01-01T00:00:00Z"));
    assert_eq!(guards(&root, PLAIN), (false, Some(0)));
}

/// REQ-3: a registry that cannot be read keeps the guard on. A symlinked `sessions` folder is
/// refused by the registry reader (never followed), so the clone counts as running a lane.
#[test]
fn an_unreadable_registry_keeps_the_guard_on() {
    let (_d, root) = repo("issue/12-fix-typo");
    let elsewhere = root.join("home/elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    let clone_state = root.join(".git/.moai-cli");
    fs::create_dir_all(&clone_state).unwrap();
    std::os::unix::fs::symlink(&elsewhere, clone_state.join("sessions")).unwrap();
    assert_eq!(guards(&root, PLAIN), (true, Some(2)));
}

#[test]
fn spec_branches_are_not_guarded_with_or_without_a_lane() {
    let (_d, root) = repo("spec/20261117-commit-guard-users");
    assert_eq!(guards(&root, PLAIN), (false, Some(0)));
    register_lane(&root, None);
    assert_eq!(guards(&root, PLAIN), (false, Some(0)));
}
