//! #2348: hook-mode output and exit codes. A generated hook runs `moai-cli hook-receipt` or
//! `moai-cli claim-verify . --hook`; both hosts read the handler's stdout as hook output and
//! read exit 2 as "block". In hook mode stdout stays empty and only a guard refusal exits 2;
//! usage errors and receipt errors exit 1 (a non-blocking hook error) with the reason on
//! stderr, and a command the guard does not check never reads the transport setting. A
//! guarded command the guard cannot verify is a refusal (fail-closed). Every case runs the
//! real binary in a temporary project and HOME.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Output, Stdio};

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

fn run(args: &[&str], cwd: &Path, stdin: &[u8], env: &[(&str, &str)]) -> Run {
    let mut cmd = Command::new(crate::common::MOAI_BIN);
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
    let mut child = cmd.spawn().unwrap();
    // A usage error exits before reading stdin; a broken pipe here is not a failure.
    common::feed_stdin(&mut child, stdin);
    let out: Output = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

fn project() -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    fs::create_dir_all(dir.path().join("home")).unwrap();
    dir
}

fn pre_tool_use(command: &str, cwd: &Path) -> Vec<u8> {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
           "tool_input": {"command": command}, "cwd": cwd})
    .to_string()
    .into_bytes()
}

/// AC-1: a successful hook run prints nothing — allow, skip and malformed stdin alike — while
/// the receipt itself is still written.
#[test]
fn hook_mode_success_prints_nothing() {
    let p = project();
    let root = fs::canonicalize(p.path()).unwrap();
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    let session = br#"{"hook_event_name":"SessionStart","source":"startup"}"#;
    for host in ["claude", "codex"] {
        let r = run(&["hook-receipt", "--host", host], &root, session, &[]);
        assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{host}: {}", r.stderr);
        let r = run(&["hook-receipt", "--host", host], &root, b"{oops", &[]);
        assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{host} malformed");
    }
    let log = fs::read_to_string(root.join(".moai-cli/receipts/hooks.jsonl")).unwrap();
    assert_eq!(
        log.lines().count(),
        2,
        "two well-formed inputs, two receipt lines"
    );
    // A project without .moai-cli/ is a skip — still nothing on stdout.
    let bare = project();
    let r = run(
        &["hook-receipt", "--host", "claude"],
        bare.path(),
        session,
        &[],
    );
    assert_eq!((r.code, r.stdout.as_str()), (0, ""));

    for input in [
        pre_tool_use("ls -la", &root),
        b"not json".to_vec(),
        Vec::new(),
    ] {
        let r = run(&["claim-verify", ".", "--hook"], &root, &input, &[]);
        assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{}", r.stderr);
    }
}

/// AC-1: a call outside hook mode keeps its envelope.
#[test]
fn non_hook_calls_keep_the_envelope() {
    let p = project();
    let r = run(&["claim-verify", "."], p.path(), b"not json", &[]);
    assert_eq!(r.code, 1);
    let env: Value = serde_json::from_str(&r.stdout).unwrap();
    assert_eq!(env["operation"], "claim.verify");
    assert_eq!(env["ok"], false);
    // A usage error outside hook mode is still the usage envelope with exit 2.
    let r = run(
        &["hooks-plan", ".", "--host", "claude", "--bogus"],
        p.path(),
        b"",
        &[],
    );
    assert_eq!(r.code, 2);
    let env: Value = serde_json::from_str(&r.stdout).unwrap();
    assert_eq!(env["error"]["kind"], "usage");
}

/// AC-2: in hook mode a usage error is a non-blocking error — exit 1, nothing on stdout, the
/// reason on stderr — never exit 2.
#[test]
fn hook_mode_usage_errors_never_block() {
    let p = project();
    let input = pre_tool_use("ls", p.path());
    for args in [
        &["hook-receipt", "--host", "claude", "--bogus"][..],
        &["hook-receipt"][..],
        &["claim-verify", ".", "--hook", "--bogus"][..],
        &["claim-verify", "--hook"][..],
        &["claim-verifyx", ".", "--hook"][..],
    ] {
        let r = run(args, p.path(), &input, &[]);
        assert_eq!(r.code, 1, "{args:?}: {}", r.stderr);
        assert_eq!(r.stdout, "", "{args:?}");
        assert!(
            !r.stderr.trim().is_empty(),
            "{args:?}: the reason goes to stderr"
        );
    }
}

/// AC-2: a bad transport setting does not touch a call the guard does not check, and a
/// command the guard does check still refuses with exit 2 (the guard's own refusal).
#[test]
fn hook_mode_transport_setting_only_matters_to_a_guarded_command() {
    let p = project();
    let root = fs::canonicalize(p.path()).unwrap();
    let bad = [("MOAI_FORGE_GH_BIN", "gh")];
    let r = run(
        &["claim-verify", ".", "--hook"],
        &root,
        &pre_tool_use("ls", &root),
        &bad,
    );
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "{}", r.stderr);

    // A lane worktree: a commit is guarded, and the refusal still blocks.
    // Preserve the fixture's explicit identity while applying the shared git helper's guards.
    let git = |args: &[&str]| {
        let out = Command::new("git")
            .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
            .args(args)
            .current_dir(&root)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("GIT_AUTHOR_NAME", "t")
            .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
            .env("GIT_COMMITTER_NAME", "t")
            .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}");
    };
    git(&["init", "-q", "-b", "main"]);
    git(&["commit", "-q", "--allow-empty", "-m", "base"]);
    git(&[
        "remote",
        "add",
        "origin",
        "https://github.com/acme/widgets.git",
    ]);
    git(&["checkout", "-q", "-b", "issue/7-widget"]);
    let r = run(
        &["claim-verify", ".", "--hook"],
        &root,
        &pre_tool_use("git commit -m work", &root),
        &[],
    );
    assert_eq!(r.code, 2, "{}", r.stderr);
    assert_eq!(r.stdout, "");
    assert!(r.stderr.contains("claim verify: uncertain"), "{}", r.stderr);

    // A guarded command with a bad transport setting cannot be verified, so the guard refuses
    // it (fail-closed, exit 2) rather than letting an unverified commit through. Whether a
    // configuration error on a guarded command should block is the guard policy decision (D5,
    // HOLD); this pins the current behaviour so a change to it is deliberate.
    let r = run(
        &["claim-verify", ".", "--hook"],
        &root,
        &pre_tool_use("git commit -m work", &root),
        &bad,
    );
    assert_eq!(r.code, 2, "{}", r.stderr);
    assert_eq!(r.stdout, "");
    assert!(r.stderr.contains("MOAI_FORGE_GH_BIN"), "{}", r.stderr);
}

/// AC-3: hooks-plan and hooks-apply report whether the generated command resolves on the PATH
/// they see; an unresolved binary is its own state and the set is never reported active.
#[test]
fn hooks_plan_and_apply_report_an_unresolved_binary() {
    let p = project();
    let root = fs::canonicalize(p.path()).unwrap();
    let bin = root.join("bin");
    fs::create_dir_all(&bin).unwrap();
    let empty = root.join("empty");
    fs::create_dir_all(&empty).unwrap();
    let no_binary = empty.to_string_lossy().to_string();
    let r = run(
        &["hooks-plan", ".", "--host", "codex"],
        &root,
        b"",
        &[("PATH", &no_binary)],
    );
    let plan: Value = serde_json::from_str(&r.stdout).unwrap();
    assert_eq!(plan["ok"], true, "{plan}");
    assert_eq!(plan["result"]["hook_binary"]["state"], "binary_unresolved");
    assert_eq!(plan["result"]["hook_binary"]["name"], "moai");
    assert_eq!(plan["result"]["hooks_active"], false);

    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let r = run(
        &[
            "hooks-apply",
            ".",
            "--host",
            "codex",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-2348",
        ],
        &root,
        b"",
        &[("PATH", &no_binary)],
    );
    let applied: Value = serde_json::from_str(&r.stdout).unwrap();
    assert_eq!(applied["ok"], true, "{applied}");
    assert_eq!(
        applied["result"]["hook_binary"]["state"],
        "binary_unresolved"
    );
    assert_eq!(applied["result"]["hooks_active"], false);

    // Only the alias `moai-cli` on that PATH: the generated hooks call `moai` (#2812 W3), so
    // they still do not resolve.
    common::write_executable(&bin.join("moai-cli"), "#!/bin/sh\nexit 0\n");
    let with_binary = bin.to_string_lossy().to_string();
    let r = run(
        &["hooks-plan", ".", "--host", "claude"],
        &root,
        b"",
        &[("PATH", &with_binary)],
    );
    let plan: Value = serde_json::from_str(&r.stdout).unwrap();
    assert_eq!(plan["result"]["hook_binary"]["name"], "moai", "{plan}");
    assert_eq!(
        plan["result"]["hook_binary"]["state"], "binary_unresolved",
        "{plan}"
    );

    // With an executable `moai` on that PATH the binary resolves; whether the host runs
    // the hooks is still unknown until a host receipt says so.
    common::write_executable(&bin.join("moai"), "#!/bin/sh\nexit 0\n");
    let r = run(
        &["hooks-plan", ".", "--host", "claude"],
        &root,
        b"",
        &[("PATH", &with_binary)],
    );
    let plan: Value = serde_json::from_str(&r.stdout).unwrap();
    assert_eq!(plan["result"]["hook_binary"]["state"], "resolved", "{plan}");
    assert_eq!(plan["result"]["hooks_active"], "not_run");
}
