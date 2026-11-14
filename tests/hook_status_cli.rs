//! #2391 stage 4: the hook state vocabulary (`installed`, `configured`, `trusted`, `fired`,
//! `action_blocked`, `verified`) from local evidence only, and the engine's `deny` line that
//! `action_blocked` reads. Every case uses a temporary project and HOME; no host runs and no
//! real `~/.claude` or `~/.codex` is read or written.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

fn project() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    fs::create_dir_all(root.join(".moai-cli/receipts")).unwrap();
    (dir, root)
}

fn status(root: &Path, host: &str, path_env: &str) -> Value {
    moai_cli::hooks::status(root, host, Some(&root.join("home")), path_env).unwrap()
}

/// Writes the generated hook set for `host` through the real plan/apply pair.
fn configure(root: &Path, host: &str) {
    let home = root.join("home");
    let plan = moai_cli::hooks::plan(root, host, Some(&home)).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    moai_cli::hooks::apply(root, host, Some(&home), &digest, "test-permission").unwrap();
}

fn config_ms(root: &Path, rel: &str) -> u64 {
    fs::metadata(root.join(rel))
        .unwrap()
        .modified()
        .unwrap()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_millis() as u64
}

fn append(root: &Path, file: &str, line: Value) {
    let path = root.join(".moai-cli/receipts").join(file);
    let mut text = fs::read_to_string(&path).unwrap_or_default();
    text.push_str(&line.to_string());
    text.push('\n');
    fs::write(path, text).unwrap();
}

#[test]
fn an_empty_project_has_no_state_and_trust_is_never_claimed() {
    let (_d, root) = project();
    for host in ["claude", "codex"] {
        let s = status(&root, host, "");
        assert_eq!(s["installed"], json!(false));
        assert_eq!(s["configured"], json!("no"));
        assert_eq!(s["trusted"], json!("unknown"));
        assert_eq!(s["fired"], json!(false));
        assert_eq!(s["action_blocked"], json!(false));
        assert_eq!(s["verified"], json!(false));
    }
}

#[test]
fn installed_follows_the_given_path_only() {
    let (_d, root) = project();
    let bin = root.join("home/bin");
    fs::create_dir_all(&bin).unwrap();
    common::write_executable(&bin.join("moai-cli"), "#!/bin/sh\nexit 0\n");
    assert_eq!(
        status(&root, "codex", bin.to_str().unwrap())["installed"],
        json!(true)
    );
    assert_eq!(
        status(&root, "codex", "/nonexistent-dir")["installed"],
        json!(false)
    );
}

#[test]
fn configured_is_yes_after_apply_and_drifted_after_an_edit() {
    let (_d, root) = project();
    for (host, file) in [
        ("codex", ".codex/hooks.json"),
        ("claude", ".claude/settings.json"),
    ] {
        configure(&root, host);
        assert_eq!(
            status(&root, host, "")["configured"],
            json!("yes"),
            "{host}"
        );
        let mut v: Value =
            serde_json::from_str(&fs::read_to_string(root.join(file)).unwrap()).unwrap();
        v["hooks"]["Stop"] = json!([]);
        fs::write(root.join(file), v.to_string()).unwrap();
        assert_eq!(
            status(&root, host, "")["configured"],
            json!("drifted"),
            "{host}"
        );
    }
}

#[test]
fn fired_counts_only_this_hosts_lines_from_the_config_time_on() {
    let (_d, root) = project();
    configure(&root, "codex");
    let t = config_ms(&root, ".codex/hooks.json");
    append(
        &root,
        "hooks.jsonl",
        json!({"schema_version": 1, "host": "codex", "unix_ms": t - 1}),
    );
    append(
        &root,
        "hooks.jsonl",
        json!({"schema_version": 1, "host": "claude", "unix_ms": t + 5}),
    );
    assert_eq!(
        status(&root, "codex", "")["fired"],
        json!(false),
        "one ms early, other host"
    );
    append(
        &root,
        "hooks.jsonl",
        json!({"schema_version": 1, "host": "codex", "unix_ms": t}),
    );
    assert_eq!(
        status(&root, "codex", "")["fired"],
        json!(true),
        "exactly at the config time"
    );
}

#[test]
fn action_blocked_needs_a_deny_without_a_later_observation_of_the_same_call() {
    let (_d, root) = project();
    configure(&root, "codex");
    let t = config_ms(&root, ".codex/hooks.json");
    let deny = |session: &str, call: Value, ms: u64| json!({"kind": "deny", "host": "codex", "session": session, "call": call, "unix_ms": ms});
    append(&root, "hook-engine.jsonl", deny("s", json!("aaaa"), t - 1));
    append(&root, "hook-engine.jsonl", deny("s", Value::Null, t + 1));
    append(
        &root,
        "hook-engine.jsonl",
        deny("unknown", json!("cccc"), t + 1),
    );
    assert_eq!(
        status(&root, "codex", "")["action_blocked"],
        json!(false),
        "early, keyless, or without a session key"
    );
    append(&root, "hook-engine.jsonl", deny("s", json!("bbbb"), t));
    let s = status(&root, "codex", "");
    assert_eq!(s["action_blocked"], json!(true));
    assert_eq!(
        s["fired"],
        json!(true),
        "the deny line is itself a fired line"
    );
    append(
        &root,
        "hook-engine.jsonl",
        json!({"kind": "observe", "host": "codex", "session": "s", "call": "bbbb", "unix_ms": t}),
    );
    assert_eq!(
        status(&root, "codex", "")["action_blocked"],
        json!(false),
        "the denied call ran after all (same host, session, call, same ms)"
    );
}

/// Audit H4-B1: only an observation of the same host, session and call at or after the deny
/// cancels it. Literal counter-examples from the audit, each against a fresh deny at `t + 10`.
#[test]
fn unrelated_or_earlier_observations_never_cancel_a_deny() {
    for (label, host, session, offset, cancels) in [
        ("other host", "claude", "s", 2i64, false),
        ("earlier", "codex", "s", -1, false),
        ("other session", "codex", "other", 2, false),
        ("same call later", "codex", "s", 2, true),
    ] {
        let (_d, root) = project();
        configure(&root, "codex");
        let t = config_ms(&root, ".codex/hooks.json") + 10;
        append(
            &root,
            "hook-engine.jsonl",
            json!({"kind": "deny", "host": "codex", "session": "s", "call": "same-call", "unix_ms": t}),
        );
        append(
            &root,
            "hook-engine.jsonl",
            json!({"kind": "observe", "host": host, "session": session, "call": "same-call",
                   "unix_ms": (t as i64 + offset) as u64}),
        );
        assert_eq!(
            status(&root, "codex", "")["action_blocked"],
            json!(!cancels),
            "{label}"
        );
    }
}

/// Audit H4-B2: a deny produced by calling the engine directly (no host) leaves the same lines a
/// host run would, so local evidence never sets `verified`.
#[test]
fn a_direct_engine_call_never_makes_the_hooks_verified() {
    let (_d, root) = project();
    configure(&root, "codex");
    let input = json!({"hook_event_name": "PreToolUse", "session_id": "direct-cli-no-host",
                       "tool_use_id": "fixture-direct-call", "tool_name": "Bash",
                       "tool_input": {"command": "git commit -m no-token"}, "cwd": root});
    common::git::command()
        .args(["init", "-q", "-b", "issue/77-x"])
        .current_dir(&root)
        .env("HOME", root.join("home"))
        .status()
        .unwrap();
    let mut child = Command::new(BIN)
        .args([
            "hook-dispatch",
            "--host",
            "codex",
            "--event",
            "tool-preflight",
        ])
        .current_dir(&root)
        .env("HOME", root.join("home"))
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    use std::io::Write;
    child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(String::from_utf8_lossy(&out.stdout).contains("\"deny\""));
    let s = status(&root, "codex", "");
    assert_eq!(s["fired"], json!(true), "local line exists");
    assert_eq!(
        s["action_blocked"],
        json!(true),
        "local deny, no later observation"
    );
    assert_eq!(s["verified"], json!(false), "no host ran");
    assert!(s["evidence"]
        .as_str()
        .unwrap()
        .contains("direct hook-dispatch call"));
}

#[test]
fn engine_deny_leaves_a_keyed_line_without_the_raw_call_id() {
    let (_d, root) = project();
    let git = |args: &[&str]| {
        // Fixture git through the shared runner (#2352 guards), with the fixture HOME.
        let ok = common::git::command()
            .args(args)
            .current_dir(&root)
            .env("HOME", root.join("home"))
            .status()
            .unwrap()
            .success();
        assert!(ok, "git {args:?}");
    };
    git(&["init", "-q", "-b", "issue/77-k"]);
    git(&[
        "-c",
        "user.name=t",
        "-c",
        "user.email=t@example.invalid",
        "commit",
        "-q",
        "--allow-empty",
        "-m",
        "i",
    ]);
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash", "session_id": "s",
                       "tool_use_id": "RAWCALLID-123", "tool_input": {"command": "git commit -m x"},
                       "cwd": root});
    let mut child = Command::new(BIN)
        .args([
            "hook-dispatch",
            "--host",
            "codex",
            "--event",
            "tool-preflight",
        ])
        .current_dir(&root)
        .env("HOME", root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    assert!(String::from_utf8_lossy(&out.stdout).contains("\"deny\""));
    let text = fs::read_to_string(root.join(".moai-cli/receipts/hook-engine.jsonl")).unwrap();
    let line: Value = serde_json::from_str(text.lines().next().unwrap()).unwrap();
    assert_eq!(line["kind"], json!("deny"));
    assert_eq!(line["host"], json!("codex"));
    assert_eq!(line["call"].as_str().unwrap().len(), 16);
    assert!(!text.contains("RAWCALLID"), "{text}");
}

#[test]
fn hooks_plan_reports_hook_status_and_writes_nothing() {
    let (_d, root) = project();
    let out = Command::new(BIN)
        .args(["hooks-plan", root.to_str().unwrap(), "--host", "codex"])
        .current_dir(&root)
        .env("HOME", root.join("home"))
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(0));
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["result"]["hook_status"]["configured"], json!("no"));
    assert_eq!(v["result"]["hook_status"]["trusted"], json!("unknown"));
    assert!(!root.join(".codex").exists());
}
