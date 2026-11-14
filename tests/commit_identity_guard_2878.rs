//! #2878 E3 (leader decision 6872992410): the Go generation's commit-identity guard in Rust —
//! opt-in `[hooks.commit_identity]`, refusing a commit-writing git command whose author or
//! committer email is one of this repository's test addresses or a `deny_emails` entry.
//! Temporary repositories only; git runs through the shared fixture runner.
mod common;

use moai_cli::hook_config::IdentityRule;
use moai_cli::hook_guards::commit_identity_guard;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Stdio;

fn git(cwd: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A repository whose own config names `email` as the identity.
fn repo(email: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    git(&root, &["config", "user.name", "someone"]);
    git(&root, &["config", "user.email", email]);
    (dir, root)
}

fn rule(extra: &[&str]) -> IdentityRule {
    IdentityRule {
        deny_emails: extra.iter().map(|e| e.to_string()).collect(),
    }
}

/// The guard probes with the test process's environment; keep the developer's own git config
/// out of it so only the repository's config speaks.
fn quiet_git_env() {
    std::env::set_var("GIT_CONFIG_GLOBAL", "/dev/null");
    std::env::set_var("GIT_CONFIG_SYSTEM", "/dev/null");
}

#[test]
fn a_test_identity_in_the_repository_config_refuses_every_committing_verb() {
    quiet_git_env();
    let (_d, root) = repo("fixture@example.invalid");
    for cmd in [
        "git commit -m x",
        "git merge --no-ff lane",
        "git cherry-pick abc123",
        "git revert abc123",
        "git rebase main",
        "git am patch.mbox",
        "git pull",
        "git commit-tree HEAD^{tree}",
        "cd . && git commit -m x",
    ] {
        let reason = commit_identity_guard(cmd, &root, &rule(&[])).unwrap_or_default();
        assert!(
            reason.contains("<fixture@example.invalid> is a test address"),
            "{cmd}: {reason:?}"
        );
    }
    for cmd in [
        "git status",
        "git log -1",
        "git merge-base HEAD main",
        "echo git commit",
    ] {
        assert_eq!(commit_identity_guard(cmd, &root, &rule(&[])), None, "{cmd}");
    }
}

#[test]
fn the_command_s_own_identity_is_judged_first() {
    quiet_git_env();
    let (_d, real) = repo("someone@moai.dev");
    assert_eq!(
        commit_identity_guard("git commit -m x", &real, &rule(&[])),
        None
    );
    for (cmd, role) in [
        ("git -c user.email=t@example.invalid commit -m x", "author"),
        ("git -c USER.EMAIL=t@example.invalid commit -m x", "author"),
        ("git -c user.email=me@moai.dev commit -m x", ""),
        (
            "GIT_AUTHOR_EMAIL=test@example.invalid git commit -m x",
            "author",
        ),
        ("GIT_COMMITTER_EMAIL=T@T git commit -m x", "committer"),
        (
            "git commit --author 'Fix Ture <fixture@example.invalid>' -m x",
            "author",
        ),
        (
            "git commit --author='Fix Ture <fixture@example.invalid>' -m x",
            "author",
        ),
    ] {
        let got = commit_identity_guard(cmd, &real, &rule(&[]));
        if role.is_empty() {
            // A real identity on the command line passes.
            assert_eq!(got, None, "{cmd}");
        } else {
            let reason = got.unwrap_or_default();
            assert!(
                reason.contains(&format!("the {role} identity")),
                "{cmd}: {reason}"
            );
        }
    }
    // A real author with a test committer is still refused, and both overridden passes.
    let (_d2, fixture) = repo("fixture@example.invalid");
    let reason = commit_identity_guard(
        "git commit --author 'Me <me@moai.dev>' -m x",
        &fixture,
        &rule(&[]),
    )
    .unwrap_or_default();
    assert!(
        reason.contains("the committer identity <fixture@example.invalid>"),
        "{reason}"
    );
    assert_eq!(
        commit_identity_guard(
            "GIT_AUTHOR_EMAIL=me@moai.dev GIT_COMMITTER_EMAIL=me@moai.dev git commit -m x",
            &fixture,
            &rule(&[])
        ),
        None
    );
}

#[test]
fn deny_emails_add_to_the_list_without_case() {
    quiet_git_env();
    let (_d, root) = repo("CI@Corp.Example");
    assert_eq!(
        commit_identity_guard("git commit -m x", &root, &rule(&[])),
        None
    );
    let reason = commit_identity_guard("git commit -m x", &root, &rule(&["ci@corp.example"]))
        .unwrap_or_default();
    assert!(reason.contains("<ci@corp.example>"), "{reason}");
}

/// Fail-open (the Go guard's rule): a folder git cannot read gives no identity, so nothing is
/// refused.
#[test]
fn an_unreadable_identity_lets_the_command_through() {
    quiet_git_env();
    let missing = std::env::temp_dir().join(format!("moai-2878-none-{}", std::process::id()));
    assert_eq!(
        commit_identity_guard("git commit -m x", &missing, &rule(&[])),
        None
    );
}

fn dispatch(cwd: &Path, command: &str) -> (i32, String) {
    let input = json!({"session_id": "s", "hook_event_name": "PreToolUse", "tool_name": "Bash",
                       "tool_use_id": "t1", "tool_input": {"command": command}, "cwd": cwd})
    .to_string();
    let mut child = std::process::Command::new(common::MOAI_BIN)
        .args([
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ])
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.as_bytes());
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// Off unless `~/.moai-cli/config` turns it on; then the hook answers with the documented
/// PreToolUse deny object.
#[test]
fn the_hook_follows_the_config() {
    let (_d, root) = repo("fixture@example.invalid");
    let home = root.join("home");
    assert_eq!(dispatch(&root, "git merge lane"), (0, String::new()));
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    fs::write(
        home.join(".moai-cli/config"),
        "[hooks.commit_identity]\nenabled = true\n",
    )
    .unwrap();
    let (code, stdout) = dispatch(&root, "git merge lane");
    assert_eq!(code, 0);
    let v: Value = serde_json::from_str(stdout.trim()).expect("one JSON object");
    assert_eq!(v["hookSpecificOutput"]["permissionDecision"], "deny");
    let reason = v["hookSpecificOutput"]["permissionDecisionReason"]
        .as_str()
        .unwrap();
    assert!(
        reason.contains("<fixture@example.invalid> is a test address"),
        "{reason}"
    );
    assert_eq!(dispatch(&root, "git status"), (0, String::new()));
}
