//! #2391 stage 1: `moai-cli hook-dispatch --host claude --event session-check|tool-preflight`.
//! Every case runs the real binary in a temporary git repository with a temporary HOME; no
//! user or project host configuration is read or written.
//!
//! Host semantics the simulated host applies (https://code.claude.com/docs/en/hooks, read
//! 2026-11-10): exit 0 with `hookSpecificOutput.permissionDecision: "deny"` prevents the tool
//! call; exit 2 blocks; any other exit code, or a command that fails to start, is a
//! non-blocking error and the tool call goes ahead. The simulation proves the engine's answer
//! under those rules; a real Claude Code session firing the hook is not run here.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::io::{ErrorKind, Write as _};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// Runs `program args` in `cwd` with `stdin`. The child may exit before reading all of its
/// stdin; a BrokenPipe on the write is that race, not a failure (#2348).
fn run_program(
    program: &Path,
    args: &[&str],
    cwd: &Path,
    stdin: &[u8],
    env: &[(&str, &str)],
) -> std::io::Result<Run> {
    let mut cmd = Command::new(program);
    cmd.args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    for (k, v) in env {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn()?;
    if let Err(e) = child.stdin.take().unwrap().write_all(stdin) {
        assert_eq!(e.kind(), ErrorKind::BrokenPipe, "stdin write: {e}");
    }
    let out = child.wait_with_output()?;
    Ok(Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    })
}

fn dispatch(event: &str, cwd: &Path, stdin: &[u8], env: &[(&str, &str)]) -> Run {
    run_program(
        Path::new(BIN),
        &["hook-dispatch", "--host", "claude", "--event", event],
        cwd,
        stdin,
        env,
    )
    .unwrap()
}

fn git(cwd: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        .args(args)
        .current_dir(cwd)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).trim().to_string()
}

/// A temporary repository on `branch` with one commit and its own HOME.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", branch]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    fs::write(root.join("a.txt"), "a\n").unwrap();
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    git(&root, &["add", "a.txt", ".gitignore"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    (dir, root)
}

fn pre_tool(command: &str, cwd: &Path) -> Vec<u8> {
    json!({"session_id": "s", "hook_event_name": "PreToolUse", "tool_name": "Bash",
           "tool_use_id": "t1", "tool_input": {"command": command}, "cwd": cwd})
    .to_string()
    .into_bytes()
}

/// The documented PreToolUse deny object, and nothing else, on stdout.
fn assert_deny(r: &Run) -> String {
    assert_eq!(r.code, 0, "deny exits 0: {}", r.stderr);
    let v: Value = serde_json::from_str(r.stdout.trim()).expect("stdout is one JSON object");
    let out = v.as_object().unwrap();
    assert_eq!(out.keys().collect::<Vec<_>>(), vec!["hookSpecificOutput"]);
    let h = out["hookSpecificOutput"].as_object().unwrap();
    let mut keys: Vec<&String> = h.keys().collect();
    keys.sort();
    assert_eq!(
        keys,
        vec![
            "hookEventName",
            "permissionDecision",
            "permissionDecisionReason"
        ]
    );
    assert_eq!(h["hookEventName"], "PreToolUse");
    assert_eq!(h["permissionDecision"], "deny");
    h["permissionDecisionReason"].as_str().unwrap().to_string()
}

fn assert_allow(r: &Run) {
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "allow: {}", r.stderr);
}

fn assert_engine_error(r: &Run) {
    assert_eq!(
        (r.code, r.stdout.as_str()),
        (1, ""),
        "engine error exits 1, empty stdout"
    );
    assert!(r.stderr.contains("engine error"), "{}", r.stderr);
}

#[test]
fn commit_without_claim_token_on_lane_branch_is_denied_without_echoing_the_command() {
    let (_d, root) = repo("issue/77-fixture-key");
    let r = dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git commit -m 'SECRETWORD change'", &root),
        &[],
    );
    let reason = assert_deny(&r);
    assert!(reason.contains("Moai-Claim: 77#"), "{reason}");
    assert!(
        !reason.contains("SECRETWORD") && !reason.contains("fixture-key"),
        "{reason}"
    );
}

#[test]
fn commit_naming_another_issue_is_denied() {
    let (_d, root) = repo("issue/77-k");
    let r = dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git commit -m 'x\n\nMoai-Claim: 78#5'", &root),
        &[],
    );
    assert!(assert_deny(&r).contains("another Issue"));
}

#[test]
fn commit_with_the_branch_token_is_allowed() {
    let (_d, root) = repo("issue/77-k");
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git commit -m 'x\n\nMoai-Claim: 77#5'", &root),
        &[],
    ));
}

#[test]
fn non_lane_branch_non_commit_and_non_bash_are_allowed() {
    let (_d, root) = repo("main");
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git commit -m x", &root),
        &[],
    ));
    let (_d2, lane) = repo("issue/77-k");
    assert_allow(&dispatch(
        "tool-preflight",
        &lane,
        &pre_tool("git status", &lane),
        &[],
    ));
    assert_allow(&dispatch(
        "tool-preflight",
        &lane,
        &pre_tool("echo git commit -m x", &lane),
        &[],
    ));
    let write = json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
                       "tool_input": {"file_path": "a.txt", "content": "git commit"}, "cwd": lane});
    assert_allow(&dispatch(
        "tool-preflight",
        &lane,
        write.to_string().as_bytes(),
        &[],
    ));
}

#[test]
fn unknown_input_fields_are_ignored() {
    let (_d, root) = repo("issue/77-k");
    let mut v: Value = serde_json::from_slice(&pre_tool("git commit -m x", &root)).unwrap();
    v["permission_mode"] = json!("default");
    v["future_field"] = json!({"nested": [1, 2]});
    assert_deny(&dispatch(
        "tool-preflight",
        &root,
        v.to_string().as_bytes(),
        &[],
    ));
}

#[test]
fn engine_errors_exit_1_with_empty_stdout() {
    let (_d, root) = repo("issue/77-k");
    let good = pre_tool("git commit -m x", &root);
    assert_engine_error(&dispatch("tool-preflight", &root, b"{oops", &[]));
    assert_engine_error(&dispatch("tool-preflight", &root, b"[1,2]", &[]));
    assert_engine_error(&dispatch("tool-preflight", &root, b"", &[]));
    // The event name in the input must match the requested duty.
    let stop = json!({"hook_event_name": "Stop", "tool_name": "Bash",
                      "tool_input": {"command": "git commit -m x"}});
    assert_engine_error(&dispatch(
        "tool-preflight",
        &root,
        stop.to_string().as_bytes(),
        &[],
    ));
    assert_engine_error(&dispatch("tool-observe", &root, &good, &[]));
    assert_engine_error(&dispatch("turn-validation", &root, &good, &[]));
    assert_engine_error(&dispatch("no-such-duty", &root, &good, &[]));
    for host in ["codex", "other"] {
        let r = run_program(
            Path::new(BIN),
            &["hook-dispatch", "--host", host, "--event", "tool-preflight"],
            &root,
            &good,
            &[],
        )
        .unwrap();
        assert_engine_error(&r);
    }
}

#[test]
fn a_bad_command_line_is_a_non_blocking_error_never_exit_2() {
    let (_d, root) = repo("issue/77-k");
    for args in [
        &["hook-dispatch", "--host", "claude"][..],
        &["hook-dispatch", "--bogus"][..],
    ] {
        let r = run_program(Path::new(BIN), args, &root, b"{}", &[]).unwrap();
        assert_eq!(
            (r.code, r.stdout.as_str()),
            (1, ""),
            "{args:?}: {}",
            r.stderr
        );
    }
}

#[test]
fn session_check_adds_fixed_vocabulary_context() {
    let (_d, root) = repo("issue/77-secret-key-name");
    let start = br#"{"hook_event_name":"SessionStart","source":"startup","session_id":"s"}"#;
    let r = dispatch("session-check", &root, start, &[]);
    assert_eq!(r.code, 0, "{}", r.stderr);
    let v: Value = serde_json::from_str(r.stdout.trim()).unwrap();
    assert_eq!(v["hookSpecificOutput"]["hookEventName"], "SessionStart");
    let ctx = v["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .unwrap();
    assert!(ctx.contains("lane branch: Issue #77"), "{ctx}");
    assert!(ctx.contains("state directory .moai-cli: absent"), "{ctx}");
    assert!(ctx.contains("local claim record: absent"), "{ctx}");
    assert!(!ctx.contains("secret-key-name"), "{ctx}");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let r = dispatch("session-check", &root, start, &[]);
    assert!(
        r.stdout.contains("state directory .moai-cli: present"),
        "{}",
        r.stdout
    );
}

/// No path reaches the forge: a fake `gh` that would leave a marker is never run.
#[test]
fn no_network_on_any_path() {
    let (_d, root) = repo("issue/77-k");
    let marker = root.join("home/gh-called");
    let fake = root.join("home/fake-gh");
    common::write_executable(
        &fake,
        &format!("#!/bin/sh\ntouch '{}'\nexit 0\n", marker.display()),
    );
    let env = [("MOAI_FORGE_GH_BIN", fake.to_str().unwrap())];
    let start = br#"{"hook_event_name":"SessionStart","source":"startup"}"#;
    dispatch("session-check", &root, start, &env);
    for cmd in [
        "git commit -m x",
        "git commit -m 'x Moai-Claim: 77#5'",
        "git push",
        "ls",
    ] {
        dispatch("tool-preflight", &root, &pre_tool(cmd, &root), &env);
    }
    dispatch("tool-preflight", &root, b"{oops", &env);
    assert!(!marker.exists(), "hook-dispatch called the forge");
}

/// A host following the documented rules: runs the hook, then runs the tool command only when
/// the hook did not block. Returns whether the tool ran.
fn host_runs_tool(hook: &Path, root: &Path, command: &str) -> bool {
    let proceed = match run_program(
        hook,
        &[
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ],
        root,
        &pre_tool(command, root),
        &[],
    ) {
        // Failed to start: a non-blocking error.
        Err(_) => true,
        Ok(r) if r.code == 2 => false,
        Ok(r) if r.code == 0 => serde_json::from_str::<Value>(r.stdout.trim())
            .map(|v| v["hookSpecificOutput"]["permissionDecision"] != "deny")
            .unwrap_or(true),
        Ok(_) => true,
    };
    if proceed {
        let status = Command::new("sh")
            .args(["-c", command])
            .current_dir(root)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("HOME", root.join("home"))
            .status()
            .unwrap();
        assert!(status.success(), "tool command failed: {command}");
    }
    proceed
}

fn commit_count(root: &Path) -> usize {
    git(root, &["rev-list", "--count", "HEAD"]).parse().unwrap()
}

#[test]
fn denied_commit_leaves_no_commit_and_the_allowed_control_does() {
    let (_d, root) = repo("issue/77-k");
    fs::write(root.join("a.txt"), "b\n").unwrap();
    git(&root, &["add", "a.txt"]);
    let before = commit_count(&root);
    assert!(!host_runs_tool(
        Path::new(BIN),
        &root,
        "git commit -q -m 'no token'"
    ));
    assert_eq!(commit_count(&root), before, "side effect despite deny");
    assert!(host_runs_tool(
        Path::new(BIN),
        &root,
        "git commit -q -m 'ok\n\nMoai-Claim: 77#5'"
    ));
    assert_eq!(
        commit_count(&root),
        before + 1,
        "allowed control must commit"
    );
}

#[test]
fn a_missing_binary_is_non_blocking_by_the_host_rules() {
    // Recorded host behaviour, not a guarantee of the engine: with the binary gone the guard
    // does not run and the tool call goes ahead. Blocking here needs the host's
    // `onFailure: "block"` (Claude Code v2.1.295+), which is a user decision (#2383).
    let (_d, root) = repo("issue/77-k");
    fs::write(root.join("a.txt"), "c\n").unwrap();
    git(&root, &["add", "a.txt"]);
    let before = commit_count(&root);
    assert!(host_runs_tool(
        &root.join("home/no-such-moai-cli"),
        &root,
        "git commit -q -m 'no token'"
    ));
    assert_eq!(commit_count(&root), before + 1);
}
