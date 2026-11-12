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

use moai_cli::claim::{ClaimRecord, Journal, Owner, SpecRef, TaskRef, WorktreeId};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

/// Runs `program args` in `cwd` with `stdin`, fed through `common::feed_stdin` (#2348: a
/// child that exits before reading its stdin is tolerated, any other write error fails).
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
    common::feed_stdin(&mut child, stdin);
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

/// Fixture git through the shared runner (#2352: gc/maintenance off, no global config), with
/// the fixture's own HOME.
fn git(cwd: &Path, args: &[&str]) -> String {
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
    // Codex has an adapter since stage 3 (tests below); any other host is an engine error.
    for host in ["other", "Claude", ""] {
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

/// A host that cancels a hook at `timeout` (https://code.claude.com/docs/en/hooks, read
/// 2026-11-12: "Claude Code cancels a `command` … hook that reaches its timeout, discarding the
/// hook's output"; "On `PreToolUse`, a timed-out … hook doesn't block the tool call"); `None`
/// waits for the hook. The hook runs with `path_prefix` first on PATH; the tool command runs
/// with the normal PATH. Returns (tool ran, hook was cancelled).
fn host_runs_tool_with_timeout(
    root: &Path,
    path_prefix: &Path,
    command: &str,
    timeout: Option<std::time::Duration>,
) -> (bool, bool) {
    let path = format!(
        "{}:{}",
        path_prefix.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    let mut child = Command::new(BIN)
        .args([
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ])
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("PATH", path)
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        // A cancelled hook's children may outlive it; no pipe of theirs is waited on.
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, &pre_tool(command, root));
    let start = std::time::Instant::now();
    let cancelled = loop {
        if child.try_wait().unwrap().is_some() {
            break false;
        }
        if timeout.is_some_and(|t| start.elapsed() >= t) {
            child.kill().unwrap();
            child.wait().unwrap();
            break true;
        }
        std::thread::sleep(std::time::Duration::from_millis(20));
    };
    let proceed = if cancelled {
        // Output discarded; a timeout is a non-blocking error on PreToolUse.
        true
    } else {
        let mut stdout = String::new();
        use std::io::Read as _;
        child
            .stdout
            .take()
            .unwrap()
            .read_to_string(&mut stdout)
            .unwrap();
        serde_json::from_str::<Value>(stdout.trim())
            .map(|v| v["hookSpecificOutput"]["permissionDecision"] != "deny")
            .unwrap_or(true)
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
    (proceed, cancelled)
}

/// Writes an executable `git` into `dir` running `body` (a POSIX sh script body).
fn git_shim(dir: &Path, body: &str) {
    use std::os::unix::fs::PermissionsExt;
    fs::create_dir_all(dir).unwrap();
    let path = dir.join("git");
    fs::write(&path, format!("#!/bin/sh\n{body}\n")).unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
}

/// R1 timeout (#2391): the engine has no deadline of its own; the host's timeout bounds it
/// and, by decision D1 (fail-open, `docs/decisions-2026-11-08.md` 17:52Z(11-10)), a timed-out
/// tool-preflight lets the tool run. The engine is held past the host timeout by a `git` on
/// the hook's PATH that sleeps (a lane commit makes the engine read the branch with `git`).
/// Control: a shim that only forwards to the real git, waited for without a timeout (a loaded
/// machine took 5–7 s here, so no wall-clock bound is put on it), denies with no side effect.
/// The sleeping shim never answers, so the host timeout decides whatever its value.
#[test]
fn a_timed_out_preflight_is_non_blocking_by_the_host_rules() {
    let host_timeout = std::time::Duration::from_secs(3);
    let (_d, root) = repo("issue/77-k");
    fs::write(root.join("a.txt"), "c\n").unwrap();
    git(&root, &["add", "a.txt"]);
    let real_git = String::from_utf8(
        Command::new("sh")
            .args(["-c", "command -v git"])
            .output()
            .unwrap()
            .stdout,
    )
    .unwrap();
    let real_git = real_git.trim();
    assert!(real_git.starts_with('/'), "{real_git}");
    let fast = root.join("home/fast-bin");
    git_shim(&fast, &format!("exec '{real_git}' \"$@\""));
    let slow = root.join("home/slow-bin");
    git_shim(&slow, "exec sleep 120");
    let before = commit_count(&root);

    let (ran, cancelled) =
        host_runs_tool_with_timeout(&root, &fast, "git commit -q -m 'no token'", None);
    assert!(!cancelled);
    assert!(
        !ran,
        "the engine reached git through the shim PATH and denied"
    );
    assert_eq!(commit_count(&root), before, "no side effect despite deny");

    let start = std::time::Instant::now();
    let (ran, cancelled) = host_runs_tool_with_timeout(
        &root,
        &slow,
        "git commit -q -m 'no token'",
        Some(host_timeout),
    );
    assert!(
        cancelled,
        "the sleeping git holds the engine past the host timeout"
    );
    assert!(
        ran,
        "a timed-out PreToolUse hook does not block (D1 fail-open)"
    );
    assert_eq!(commit_count(&root), before + 1, "the tool ran");
    assert!(
        start.elapsed() < std::time::Duration::from_secs(90),
        "the host timeout, not the engine, ended the hook"
    );
}

// ---- audit #2399 (6770112093) B1-B4: guards pinned by tests ----

/// A PreToolUse input for a lane commit without a claim token, padded to exactly `total`
/// bytes (the padding is a field the engine ignores).
fn sized_commit_input(root: &Path, total: usize) -> Vec<u8> {
    let make = |pad: &str| {
        json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
               "tool_input": {"command": "git commit -m x", "pad": pad}, "cwd": root})
        .to_string()
        .into_bytes()
    };
    let base = make("").len();
    let bytes = make(&"y".repeat(total - base));
    assert_eq!(bytes.len(), total);
    bytes
}

/// B1: the stdin bound is 4,194,304 bytes. An input of exactly that size is read whole and
/// decided (deny); one byte more is cut off, no longer parses and is an engine error — exit 1,
/// which the host treats as non-blocking, so such a commit goes ahead. That oversized-input
/// bypass is the same question as fail-closed (`onFailure: "block"`, #2383 / Issue #2391 D1).
#[test]
fn stdin_bound_is_4194304_bytes_and_one_more_is_a_non_blocking_engine_error() {
    let (_d, root) = repo("issue/77-k");
    let r = dispatch(
        "tool-preflight",
        &root,
        &sized_commit_input(&root, 4_194_304),
        &[],
    );
    assert_deny(&r);
    let r = dispatch(
        "tool-preflight",
        &root,
        &sized_commit_input(&root, 4_194_305),
        &[],
    );
    assert_engine_error(&r);
}

/// B2: the branch is read in the directory the input's `cwd` names, not the hook process's.
#[test]
fn input_cwd_decides_over_the_process_cwd() {
    let (_a, plain) = repo("main");
    let (_b, lane) = repo("issue/77-k");
    // Process in a non-lane repo, command in the lane repo: denied.
    assert_deny(&dispatch(
        "tool-preflight",
        &plain,
        &pre_tool("git commit -m x", &lane),
        &[],
    ));
    // Process in the lane repo, command in the non-lane repo: allowed.
    assert_allow(&dispatch(
        "tool-preflight",
        &lane,
        &pre_tool("git commit -m x", &plain),
        &[],
    ));
}

/// #2399 review: read commands that merely mention `commit` are not commits (they were denied
/// before the subcommand-position fix); a real commit behind global options still is.
#[test]
fn read_commands_mentioning_commit_are_allowed_and_real_commits_are_not() {
    let (_d, root) = repo("issue/77-k");
    for cmd in [
        "git log --grep commit",
        "git log -- commit",
        "git show commit-ish",
    ] {
        assert_allow(&dispatch(
            "tool-preflight",
            &root,
            &pre_tool(cmd, &root),
            &[],
        ));
    }
    // #2399 re-audit B5: a value-taking global option does not hide the commit.
    assert_deny(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git --attr-source HEAD commit -m x", &root),
        &[],
    ));
    let dir_commit = format!("git -C {} commit -m x", root.display());
    assert_deny(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool(&dir_commit, &root),
        &[],
    ));
}

/// B3: stage 1 guards commits only; a push stays with `claim-verify --hook`.
#[test]
fn lane_push_without_token_is_allowed_by_stage_one() {
    let (_d, root) = repo("issue/77-k");
    for cmd in [
        "git push",
        "git push origin HEAD",
        "git push -u origin issue/77-k",
    ] {
        assert_allow(&dispatch(
            "tool-preflight",
            &root,
            &pre_tool(cmd, &root),
            &[],
        ));
    }
}

/// B4: with an active local claim record for the worktree, the engine allows and leaves the
/// forge check to `claim-verify --hook`.
#[test]
fn a_local_claim_record_hands_the_commit_to_the_claim_guard() {
    let (_d, root) = repo("issue/77-k");
    let base = git(&root, &["rev-parse", "HEAD"]);
    let record = ClaimRecord {
        schema_version: 1,
        task_ref: TaskRef::new("github", "github.com", "acme", "widgets", 1, 70, 77),
        task_key: "0123456789abcdef".into(),
        logical_key: "B4".into(),
        owner: Owner::new("session", "agent-fixture"),
        assignee: None,
        provider_session: None,
        integration_owner: None,
        attempt: None,
        epoch: 1,
        base_revision: base,
        worktree: WorktreeId::new(
            root.join(".git").display().to_string(),
            root.join(".git").display().to_string(),
            root.display().to_string(),
            "b".repeat(40),
        ),
        write_scope: vec!["a.txt".into()],
        spec_ref: SpecRef {
            id: "SPEC-B4".into(),
            revision: "r1".into(),
            digest: None,
        },
        comment_id: Some(5),
        claim_token: Some("Moai-Claim: 77#5".into()),
        lease_until: Some("2026-11-10T23:00:00Z".into()),
        state: "assigned".into(),
        effects: vec![],
        last_observed_at: "2026-11-10T12:00:00Z".into(),
    };
    let journal = Journal::open(&moai_cli::claim::journal_root(&root).unwrap()).unwrap();
    assert!(journal.try_lock().unwrap());
    journal.commit("acquire", &record).unwrap();
    drop(journal);
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git commit -m x", &root),
        &[],
    ));
    // Control: the same commit in an identical repo without the record is denied.
    let (_e, bare) = repo("issue/77-k");
    assert_deny(&dispatch(
        "tool-preflight",
        &bare,
        &pre_tool("git commit -m x", &bare),
        &[],
    ));
}

// ---- stage 2: tool-observe and turn-validation ----

fn engine_log(root: &Path) -> PathBuf {
    root.join(".moai-cli/receipts/hook-engine.jsonl")
}

fn post_tool(event: &str, tool: &str, file_path: &str, response: &str) -> Vec<u8> {
    json!({"session_id": "raw-session-id-123", "hook_event_name": event, "tool_name": tool,
           "tool_use_id": "t1", "tool_input": {"file_path": file_path, "content": "CONTENTMARK"},
           "tool_response": {"output": response}})
    .to_string()
    .into_bytes()
}

#[test]
fn tool_observe_records_a_bounded_line_and_never_answers() {
    let (_d, root) = repo("issue/77-k");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let head = git(&root, &["rev-parse", "HEAD"]);
    let abs = root.join("src/x.rs");
    let r = dispatch(
        "tool-observe",
        &root,
        &post_tool(
            "PostToolUse",
            "Write",
            abs.to_str().unwrap(),
            "RESPONSEMARK",
        ),
        &[],
    );
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    let r = dispatch(
        "tool-observe",
        &root,
        &post_tool("PostToolUseFailure", "Edit", "/etc/hosts", "x"),
        &[],
    );
    assert_eq!((r.code, r.stdout.as_str()), (0, ""));
    let text = fs::read_to_string(engine_log(&root)).unwrap();
    let lines: Vec<Value> = text
        .lines()
        .map(|l| serde_json::from_str(l).unwrap())
        .collect();
    assert_eq!(lines.len(), 2, "{text}");
    assert_eq!(lines[0]["kind"], "observe");
    assert_eq!(lines[0]["tool"], "Write");
    assert_eq!(lines[0]["class"], "write");
    assert_eq!(lines[0]["path"], "src/x.rs");
    assert_eq!(lines[0]["outcome"], "success");
    assert_eq!(lines[0]["head"], json!(head));
    assert_eq!(lines[1]["path"], "outside");
    assert_eq!(lines[1]["outcome"], "failure");
    for mark in [
        "RESPONSEMARK",
        "CONTENTMARK",
        "raw-session-id-123",
        "/etc/hosts",
    ] {
        assert!(!text.contains(mark), "{mark} leaked: {text}");
    }
}

#[test]
fn tool_observe_without_state_dir_or_with_bad_input_is_a_quiet_skip() {
    let (_d, root) = repo("main");
    let ok = post_tool("PostToolUse", "Read", "a.txt", "x");
    let r = dispatch("tool-observe", &root, &ok, &[]);
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    assert!(
        !root.join(".moai-cli").exists(),
        "never creates the state dir"
    );
    fs::create_dir(root.join(".moai-cli")).unwrap();
    // A tool result past the 4 MiB stdin bound is cut off: quiet skip, nothing recorded.
    let big = post_tool("PostToolUse", "Bash", "a.txt", &"y".repeat(5 * 1024 * 1024));
    let r = dispatch("tool-observe", &root, &big, &[]);
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    let r = dispatch("tool-observe", &root, b"{oops", &[]);
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    assert!(!engine_log(&root).exists());
    // A wrong event name is a configuration error: engine error, still non-blocking.
    assert_engine_error(&dispatch(
        "tool-observe",
        &root,
        &post_tool("Stop", "Read", "a.txt", "x"),
        &[],
    ));
}

#[test]
fn tool_observe_refuses_a_symlinked_receipts_dir() {
    let (_d, root) = repo("main");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let elsewhere = root.join("home/elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, root.join(".moai-cli/receipts")).unwrap();
    let r = dispatch(
        "tool-observe",
        &root,
        &post_tool("PostToolUse", "Read", "a.txt", "x"),
        &[],
    );
    assert_eq!((r.code, r.stdout.as_str()), (0, ""));
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
}

#[test]
fn turn_validation_ships_with_no_rule_and_never_blocks() {
    let (_d, root) = repo("issue/77-k");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    for active in [false, true] {
        let stop = json!({"hook_event_name": "Stop", "session_id": "s",
                          "stop_hook_active": active, "last_assistant_message": "MSGMARK"});
        let r = dispatch("turn-validation", &root, stop.to_string().as_bytes(), &[]);
        assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    }
    assert!(!engine_log(&root).exists(), "no rule, nothing written");
    assert_engine_error(&dispatch("turn-validation", &root, b"{oops", &[]));
    let wrong = json!({"hook_event_name": "PreToolUse"});
    assert_engine_error(&dispatch(
        "turn-validation",
        &root,
        wrong.to_string().as_bytes(),
        &[],
    ));
}

#[test]
fn stage_two_duties_never_call_the_forge() {
    let (_d, root) = repo("issue/77-k");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let marker = root.join("home/gh-called");
    let fake = root.join("home/fake-gh");
    common::write_executable(
        &fake,
        &format!("#!/bin/sh\ntouch '{}'\nexit 0\n", marker.display()),
    );
    let env = [("MOAI_FORGE_GH_BIN", fake.to_str().unwrap())];
    dispatch(
        "tool-observe",
        &root,
        &post_tool("PostToolUse", "Bash", "a.txt", "x"),
        &env,
    );
    let stop = br#"{"hook_event_name":"Stop","session_id":"s","stop_hook_active":false}"#;
    dispatch("turn-validation", &root, stop, &env);
    assert!(!marker.exists());
}

// ---- stage 3: the Codex adapter (https://learn.chatgpt.com/docs/hooks, read 2026-11-10) ----

fn codex(event: &str, cwd: &Path, stdin: &[u8]) -> Run {
    run_program(
        Path::new(BIN),
        &["hook-dispatch", "--host", "codex", "--event", event],
        cwd,
        stdin,
        &[],
    )
    .unwrap()
}

/// Fields the Codex page says are parsed but not supported (the hook run fails and the tool
/// call continues) — never a way to block.
fn assert_no_unsupported_codex_fields(stdout: &str) {
    for field in [
        "\"ask\"",
        "\"approve\"",
        "\"continue\"",
        "\"stopReason\"",
        "\"updatedInput\"",
        "\"allow\"",
    ] {
        assert!(!stdout.contains(field), "{field} in {stdout}");
    }
}

#[test]
fn codex_tool_preflight_denies_with_the_documented_codex_shape() {
    let (_d, root) = repo("issue/77-k");
    let input = json!({"session_id": "s", "turn_id": "t", "hook_event_name": "PreToolUse",
                       "tool_name": "Bash", "tool_use_id": "u", "model": "m",
                       "tool_input": {"command": "git commit -m x"}, "cwd": root});
    let r = codex("tool-preflight", &root, input.to_string().as_bytes());
    let reason = assert_deny(&r);
    assert!(reason.contains("Moai-Claim: 77#"), "{reason}");
    assert_no_unsupported_codex_fields(&r.stdout);
    let ok = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                    "tool_input": {"command": "git commit -m 'x Moai-Claim: 77#5'"}, "cwd": root});
    assert_allow(&codex("tool-preflight", &root, ok.to_string().as_bytes()));
    // apply_patch cannot commit; a patch touching no protected path is allowed (stage 5 guards
    // only .git internals, credential paths and secret content in patches).
    let patch = json!({"hook_event_name": "PreToolUse", "tool_name": "apply_patch",
                       "tool_input": {"command": "*** Begin Patch\n*** git commit\n*** End Patch"},
                       "cwd": root});
    assert_allow(&codex(
        "tool-preflight",
        &root,
        patch.to_string().as_bytes(),
    ));
}

#[test]
fn codex_denied_commit_leaves_no_commit() {
    let (_d, root) = repo("issue/77-k");
    fs::write(root.join("a.txt"), "z\n").unwrap();
    git(&root, &["add", "a.txt"]);
    let before = commit_count(&root);
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                       "tool_input": {"command": "git commit -q -m 'no token'"}, "cwd": root});
    let r = codex("tool-preflight", &root, input.to_string().as_bytes());
    // Codex rules: exit 2, or the PreToolUse deny object, blocks the call.
    let blocked = r.code == 2
        || serde_json::from_str::<Value>(r.stdout.trim())
            .map(|v| v["hookSpecificOutput"]["permissionDecision"] == "deny")
            .unwrap_or(false);
    assert!(blocked, "{}", r.stdout);
    assert_eq!(commit_count(&root), before, "a blocked call never runs");
}

#[test]
fn codex_session_check_and_turn_validation_answers() {
    let (_d, root) = repo("issue/77-k");
    let start =
        br#"{"hook_event_name":"SessionStart","source":"startup","session_id":"s","model":"m"}"#;
    let r = codex("session-check", &root, start);
    assert_eq!(r.code, 0, "{}", r.stderr);
    let v: Value = serde_json::from_str(r.stdout.trim()).unwrap();
    assert_eq!(v.as_object().unwrap().len(), 1);
    assert_eq!(v["hookSpecificOutput"]["hookEventName"], "SessionStart");
    assert!(v["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .unwrap()
        .contains("lane branch: Issue #77"));
    assert_no_unsupported_codex_fields(&r.stdout);
    let stop = br#"{"hook_event_name":"Stop","session_id":"s","turn_id":"t","stop_hook_active":false,"last_assistant_message":null}"#;
    let r = codex("turn-validation", &root, stop);
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
}

#[test]
fn codex_tool_observe_records_apply_patch_and_has_no_failure_event() {
    let (_d, root) = repo("main");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let post = json!({"session_id": "s", "turn_id": "t", "hook_event_name": "PostToolUse",
                      "tool_name": "apply_patch", "tool_use_id": "u",
                      "tool_input": {"command": "*** Begin Patch PATCHMARK"},
                      "tool_response": "RESPONSEMARK"});
    let r = codex("tool-observe", &root, post.to_string().as_bytes());
    assert_eq!((r.code, r.stdout.as_str(), r.stderr.as_str()), (0, "", ""));
    let text = fs::read_to_string(engine_log(&root)).unwrap();
    let line: Value = serde_json::from_str(text.lines().next().unwrap()).unwrap();
    assert_eq!(line["tool"], "apply_patch");
    assert_eq!(line["class"], "write");
    assert_eq!(line["path"], Value::Null);
    assert!(
        !text.contains("PATCHMARK") && !text.contains("RESPONSEMARK"),
        "{text}"
    );
    // Codex has no PostToolUseFailure: that name is a configuration error there.
    let failure = json!({"hook_event_name": "PostToolUseFailure", "tool_name": "Bash"});
    assert_engine_error(&codex(
        "tool-observe",
        &root,
        failure.to_string().as_bytes(),
    ));
}

#[test]
fn hooks_plan_reports_per_host_engine_capabilities() {
    let (_d, root) = repo("main");
    let mut seen = Vec::new();
    for host in ["claude", "codex"] {
        let r = run_program(
            Path::new(BIN),
            &["hooks-plan", root.to_str().unwrap(), "--host", host],
            &root,
            b"",
            &[],
        )
        .unwrap();
        assert_eq!(r.code, 0, "{}", r.stdout);
        let v: Value = serde_json::from_str(r.stdout.trim()).unwrap();
        let caps = v["result"]["engine_capabilities"].clone();
        assert_eq!(caps["host"], json!(host));
        seen.push(caps);
    }
    let guards = |c: &Value| c["tool-preflight"]["guards"].to_string();
    let not_guarded = |c: &Value| c["tool-preflight"]["not_guarded"].to_string();
    // Codex file edits arrive as apply_patch, Claude's as Write/Edit/MultiEdit/NotebookEdit.
    assert!(guards(&seen[1]).contains("apply_patch"));
    assert!(!guards(&seen[0]).contains("apply_patch"));
    assert!(guards(&seen[0]).contains("NotebookEdit"));
    for caps in &seen {
        assert!(not_guarded(caps).contains("outside the project"));
        assert!(not_guarded(caps).contains("MCP tools"));
    }
    assert_eq!(seen[1]["tool-observe"]["events"], json!(["PostToolUse"]));
    assert_eq!(
        seen[0]["tool-observe"]["events"],
        json!(["PostToolUse", "PostToolUseFailure"])
    );
    assert!(seen[1]["trust"].as_str().unwrap().contains("trust"));
    for never in ["ask", "continue=false", "stopReason"] {
        assert!(seen[1]["tool-preflight"]["never_uses"]
            .to_string()
            .contains(never));
    }
    assert!(!root.join(".codex").exists() && !root.join(".claude").exists());
}

// ---- audit #2400 (6770591310) B1-B2 and the canonical-cwd question ----

/// A receipt log of exactly `len` bytes (one padding line, newline-terminated).
fn log_of_len(path: &Path, len: usize) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    let mut bytes = vec![b'x'; len - 1];
    bytes.push(b'\n');
    fs::write(path, bytes).unwrap();
    assert_eq!(fs::metadata(path).unwrap().len(), len as u64);
}

/// B1: the shared append path takes a line while the log is below 4,194,304 bytes and
/// refuses at exactly 4,194,304 — for the engine log and for `hook-receipt`'s log alike.
#[test]
fn receipt_logs_append_below_4194304_bytes_and_refuse_at_exactly_4194304() {
    let (_d, root) = repo("main");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let engine = engine_log(&root);
    let post = post_tool("PostToolUse", "Read", "a.txt", "x");
    log_of_len(&engine, 4_194_303);
    dispatch("tool-observe", &root, &post, &[]);
    assert!(
        fs::metadata(&engine).unwrap().len() > 4_194_303,
        "4,194,303 bytes: appended"
    );
    log_of_len(&engine, 4_194_304);
    dispatch("tool-observe", &root, &post, &[]);
    assert_eq!(
        fs::metadata(&engine).unwrap().len(),
        4_194_304,
        "4,194,304 bytes: refused"
    );
    let receipts = root.join(".moai-cli/receipts/hooks.jsonl");
    let start = br#"{"hook_event_name":"SessionStart","source":"startup"}"#;
    let receipt = |root: &Path| {
        run_program(
            Path::new(BIN),
            &["hook-receipt", "--host", "claude"],
            root,
            start,
            &[],
        )
        .unwrap()
    };
    log_of_len(&receipts, 4_194_303);
    assert_eq!(receipt(&root).code, 0);
    assert!(fs::metadata(&receipts).unwrap().len() > 4_194_303);
    log_of_len(&receipts, 4_194_304);
    assert_eq!(receipt(&root).code, 0);
    assert_eq!(fs::metadata(&receipts).unwrap().len(), 4_194_304);
}

/// B2: a secret-shaped target path is never logged, and a relative path that climbs out of
/// the session directory is logged as `outside`, not as written.
#[test]
fn tool_observe_never_logs_secret_shaped_or_escaping_paths() {
    let (_d, root) = repo("main");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let cases = [
        ("notes/ghp_SECRETVALUE123.txt", "unrecorded"),
        ("cfg/token=SECRETVALUE456", "unrecorded"),
        ("../x", "outside"),
        ("a/../../x", "outside"),
        ("./src/ok.rs", "./src/ok.rs"),
    ];
    for (path, _) in cases {
        let r = dispatch(
            "tool-observe",
            &root,
            &post_tool("PostToolUse", "Edit", path, "x"),
            &[],
        );
        assert_eq!((r.code, r.stdout.as_str()), (0, ""));
    }
    let text = fs::read_to_string(engine_log(&root)).unwrap();
    let logged: Vec<Value> = text
        .lines()
        .map(|l| serde_json::from_str::<Value>(l).unwrap()["path"].clone())
        .collect();
    let expected: Vec<Value> = cases.iter().map(|(_, want)| json!(want)).collect();
    assert_eq!(logged, expected);
    for leak in ["SECRETVALUE", "../x"] {
        assert!(!text.contains(leak), "{leak} leaked: {text}");
    }
}

/// The engine canonicalizes the hook process's working directory before any duty (stage 2,
/// applied to the stage-1 duties too, deliberately): a session started through a symlinked
/// path still finds its local claim record, as `claim-verify --hook` does.
#[test]
fn session_check_through_a_symlinked_cwd_still_finds_the_local_record() {
    let (_d, root) = repo("issue/77-k");
    let base = git(&root, &["rev-parse", "HEAD"]);
    let record = ClaimRecord {
        schema_version: 1,
        task_ref: TaskRef::new("github", "github.com", "acme", "widgets", 1, 70, 77),
        task_key: "0123456789abcdef".into(),
        logical_key: "CW".into(),
        owner: Owner::new("session", "agent-fixture"),
        assignee: None,
        provider_session: None,
        integration_owner: None,
        attempt: None,
        epoch: 1,
        base_revision: base,
        worktree: WorktreeId::new(
            root.join(".git").display().to_string(),
            root.join(".git").display().to_string(),
            root.display().to_string(),
            "b".repeat(40),
        ),
        write_scope: vec!["a.txt".into()],
        spec_ref: SpecRef {
            id: "SPEC-CW".into(),
            revision: "r1".into(),
            digest: None,
        },
        comment_id: Some(5),
        claim_token: Some("Moai-Claim: 77#5".into()),
        lease_until: Some("2026-11-10T23:00:00Z".into()),
        state: "assigned".into(),
        effects: vec![],
        last_observed_at: "2026-11-10T12:00:00Z".into(),
    };
    let journal = Journal::open(&moai_cli::claim::journal_root(&root).unwrap()).unwrap();
    assert!(journal.try_lock().unwrap());
    journal.commit("acquire", &record).unwrap();
    drop(journal);
    let link_dir = tempfile::TempDir::new().unwrap();
    let link = link_dir.path().join("via-link");
    std::os::unix::fs::symlink(&root, &link).unwrap();
    fs::create_dir_all(link.join("home")).unwrap();
    let start = br#"{"hook_event_name":"SessionStart","source":"startup"}"#;
    let r = dispatch("session-check", &link, start, &[]);
    assert!(
        r.stdout.contains("local claim record: present"),
        "{}",
        r.stdout
    );
}

/// Audit #2402 H3-B1: Codex sends `PostToolUse` after a failed Bash command too, so the event
/// alone never records success. Literal cases: exit 0 → success, exit 17 → failure, no
/// `exit_code` / a string / another tool → unknown. Claude keeps its event-based outcome.
#[test]
fn codex_observe_outcome_comes_from_the_tool_response_or_is_unknown() {
    let (_d, root) = repo("main");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let observe = |host: &str, event: &str, tool: &str, response: Value| {
        let input = json!({"hook_event_name": event, "session_id": "s", "tool_name": tool,
                           "tool_input": {"command": "x"}, "tool_response": response});
        let r = run_program(
            Path::new(BIN),
            &["hook-dispatch", "--host", host, "--event", "tool-observe"],
            &root,
            input.to_string().as_bytes(),
            &[],
        )
        .unwrap();
        assert_eq!((r.code, r.stdout.as_str()), (0, ""));
        let text = fs::read_to_string(engine_log(&root)).unwrap();
        let last: Value = serde_json::from_str(text.lines().last().unwrap()).unwrap();
        (
            last["outcome"].as_str().unwrap().to_string(),
            last["outcome_source"].as_str().unwrap().to_string(),
        )
    };
    let codex = |tool: &str, response: Value| observe("codex", "PostToolUse", tool, response);
    let (o, src) = codex("Bash", json!({"exit_code": 0, "aggregated_output": ""}));
    assert_eq!(o, "success");
    assert!(src.contains("undocumented"), "{src}");
    assert_eq!(
        codex("Bash", json!({"exit_code": 17, "aggregated_output": ""})).0,
        "failure"
    );
    assert_eq!(codex("Bash", json!({"exit_code": -1})).0, "failure");
    assert_eq!(codex("Bash", json!({"output": "done"})).0, "unknown");
    assert_eq!(codex("Bash", json!("plain text output")).0, "unknown");
    assert_eq!(codex("Bash", json!({"exit_code": "0"})).0, "unknown");
    assert_eq!(codex("apply_patch", json!({"exit_code": 0})).0, "unknown");
    assert_eq!(codex("mcp__fs__read", json!({"content": []})).0, "unknown");
    // Claude: the event decides, as before.
    assert_eq!(
        observe("claude", "PostToolUse", "Bash", json!({"exit_code": 17})),
        ("success".to_string(), "event".to_string())
    );
    assert_eq!(
        observe("claude", "PostToolUseFailure", "Bash", json!({})).0,
        "failure"
    );
}

// ---- stage 4: cost ----

/// Audit #2400 N3: tool-observe spawns `git` only when it will record. A fake `git` first on
/// PATH leaves a marker; without `.moai-cli/` there is no marker, with it there is one (the
/// control proves the fake is the `git` the engine would run).
#[test]
fn tool_observe_spawns_no_git_without_a_state_dir() {
    let (_d, root) = repo("main");
    let bin = root.join("home/fakebin");
    fs::create_dir_all(&bin).unwrap();
    let marker = root.join("home/git-called");
    common::write_executable(
        &bin.join("git"),
        &format!(
            "#!/bin/sh\ntouch '{}'\necho 0123456789abcdef0123456789abcdef01234567\n",
            marker.display()
        ),
    );
    let path = format!(
        "{}:{}",
        bin.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    let env = [("PATH", path.as_str())];
    let post = post_tool("PostToolUse", "Read", "a.txt", "x");
    dispatch("tool-observe", &root, &post, &env);
    assert!(!marker.exists(), "git ran although nothing is recorded");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    dispatch("tool-observe", &root, &post, &env);
    assert!(marker.exists(), "control: the fake git is the one run");
    let text = fs::read_to_string(engine_log(&root)).unwrap();
    assert!(text.contains("0123456789abcdef0123456789abcdef01234567"));
}

/// Cost (#2391 stage 4, measured): a lane commit that carries its token is decided after one
/// `git` call (the branch); only a commit without it pays the two `rev-parse` calls of the
/// local-record lookup. Counted, not timed: a logging `git` wrapper execs the real one.
#[test]
fn a_commit_with_its_token_spawns_one_git_and_one_without_spawns_three() {
    let (_d, root) = repo("issue/77-k");
    let real = String::from_utf8(
        Command::new("sh")
            .args(["-c", "command -v git"])
            .output()
            .unwrap()
            .stdout,
    )
    .unwrap()
    .trim()
    .to_string();
    let bin = root.join("home/wrapbin");
    fs::create_dir_all(&bin).unwrap();
    let log = root.join("home/git-calls");
    common::write_executable(
        &bin.join("git"),
        &format!(
            "#!/bin/sh\necho \"$*\" >> '{}'\nexec '{}' \"$@\"\n",
            log.display(),
            real
        ),
    );
    let path = format!(
        "{}:{}",
        bin.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    let env = [("PATH", path.as_str())];
    let calls = |cmd: &str| {
        let _ = fs::remove_file(&log);
        let r = dispatch("tool-preflight", &root, &pre_tool(cmd, &root), &env);
        let n = fs::read_to_string(&log).map_or(0, |t| t.lines().count());
        (n, r)
    };
    let (n, r) = calls("git commit -m 'x\n\nMoai-Claim: 77#5'");
    assert_allow(&r);
    assert_eq!(n, 1, "token present: the branch lookup only");
    let (n, r) = calls("git commit -m x");
    assert_deny(&r);
    assert_eq!(n, 3, "no token: branch + two rev-parse for the journal");
    let (n, r) = calls("ls");
    assert_allow(&r);
    assert_eq!(n, 0, "not a commit: no git at all");
}

// ---- stage 5: minimal guard bundle (A1–A4) ----

/// The refused removal leaves the fixture's `.git` in place; the allowed control really runs.
/// Both commands stay inside the temporary fixture repository.
#[test]
fn a_refused_removal_runs_nothing_and_an_allowed_one_runs() {
    let (_d, root) = repo("main");
    fs::create_dir_all(root.join("build")).unwrap();
    assert!(!host_runs_tool(Path::new(BIN), &root, "rm -rf .git"));
    assert!(root.join(".git").is_dir(), "refused: .git must survive");
    assert!(host_runs_tool(Path::new(BIN), &root, "rm -rf build"));
    assert!(!root.join("build").exists(), "allowed control must run");
}

#[test]
fn stage_five_reasons_are_fixed_text_for_both_hosts() {
    let (_d, root) = repo("main");
    let marker = "UNIQUEMARK7";
    let cmd = format!("git push --force origin main # {marker}");
    for host in ["claude", "codex"] {
        let r = run_program(
            Path::new(BIN),
            &["hook-dispatch", "--host", host, "--event", "tool-preflight"],
            &root,
            &pre_tool(&cmd, &root),
            &[],
        )
        .unwrap();
        let reason = assert_deny(&r);
        assert!(reason.contains("force push"), "{reason}");
        assert!(!reason.contains(marker), "{reason}");
    }
}

#[test]
fn file_writes_into_git_internals_are_refused_for_both_hosts() {
    let (_d, root) = repo("main");
    let write = json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
                       "tool_input": {"file_path": ".git/hooks/pre-commit", "content": "x"}, "cwd": root});
    assert_deny(&dispatch(
        "tool-preflight",
        &root,
        write.to_string().as_bytes(),
        &[],
    ));
    let patch = json!({"hook_event_name": "PreToolUse", "tool_name": "apply_patch",
                       "tool_input": {"command": "*** Begin Patch\n*** Update File: .git/config\n@@\n-a\n+b\n*** End Patch"},
                       "cwd": root});
    assert_deny(&codex(
        "tool-preflight",
        &root,
        patch.to_string().as_bytes(),
    ));
    let fine = json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
                      "tool_input": {"file_path": "src/ok.rs", "content": "x"}, "cwd": root});
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        fine.to_string().as_bytes(),
        &[],
    ));
}

// ---- stage 6: the confirmation class (B2 = A) ----

/// The PreToolUse object's decision and reason.
fn decision(r: &Run) -> (String, String) {
    assert_eq!(r.code, 0, "{}", r.stderr);
    let v: Value = serde_json::from_str(r.stdout.trim()).expect("one JSON object");
    let h = &v["hookSpecificOutput"];
    assert_eq!(h["hookEventName"], "PreToolUse");
    (
        h["permissionDecision"].as_str().unwrap().to_string(),
        h["permissionDecisionReason"].as_str().unwrap().to_string(),
    )
}

/// Decision B2 = A (Issue #2389 c6774899518 row 25): Claude asks, Codex refuses with a fixed
/// suffix; nothing is recorded as a deny for an ask. Strings only, nothing is executed.
#[test]
fn a_confirmation_asks_on_claude_and_is_refused_on_codex() {
    let (_d, root) = repo("main");
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    let input = pre_tool("git reset --hard HEAD~1", &root);
    let (d, reason) = decision(&dispatch("tool-preflight", &root, &input, &[]));
    assert_eq!(d, "ask");
    assert!(reason.contains("discards uncommitted changes"), "{reason}");
    let (d, reason) = decision(&codex("tool-preflight", &root, &input));
    assert_eq!(d, "deny");
    assert!(reason.contains("Codex cannot ask"), "{reason}");
    // An ask leaves no deny line (the user's answer is unknown).
    let log = root.join(".moai-cli/receipts/hook-engine.jsonl");
    let text = fs::read_to_string(&log).unwrap_or_default();
    assert!(!text.contains("\"kind\":\"deny\""), "{text}");
}

/// A deny anywhere on the line wins over a confirmation (Claude's order deny > ask).
#[test]
fn a_deny_on_the_same_line_wins_over_a_confirmation() {
    let (_d, root) = repo("issue/77-x");
    let r = dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git reset --hard && git push --force origin main", &root),
        &[],
    );
    assert_eq!(decision(&r).0, "deny");
    let r = dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git reset --hard; git commit -m no-token", &root),
        &[],
    );
    assert_eq!(decision(&r).0, "deny");
    let r = dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git reset --hard", &root),
        &[],
    );
    assert_eq!(decision(&r).0, "ask");
}

// ---- stage 7: opt-in rules from ~/.moai-cli/config (B3, B9) ----

fn config(home: &Path, text: &str) {
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    fs::write(home.join(".moai-cli/config"), text).unwrap();
}

fn write_input(path: &Path, cwd: &Path) -> Vec<u8> {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
           "tool_input": {"file_path": path, "content": "x"}, "cwd": cwd})
    .to_string()
    .into_bytes()
}

/// B9 is off without a config (and with a broken one), and refuses a write outside the
/// project once `[hooks.outside_project_writes]` turns it on; an `allow` folder passes.
#[test]
fn outside_project_writes_follow_the_config() {
    let (_d, root) = repo("main");
    let home = root.join("home");
    let outside = root
        .parent()
        .unwrap()
        .join(format!("moai-b9-{}", std::process::id()));
    let input = write_input(&outside.join("x.txt"), &root);
    let env = [("HOME", home.to_str().unwrap())];
    assert_allow(&dispatch("tool-preflight", &root, &input, &env));
    config(
        &home,
        "[hooks.outside_project_writes]\nenabled = true\nallow = [\n",
    );
    assert_allow(&dispatch("tool-preflight", &root, &input, &env));
    config(
        &home,
        "[hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\"]\n",
    );
    let reason = assert_deny(&dispatch("tool-preflight", &root, &input, &env));
    assert!(reason.contains("outside the project"), "{reason}");
    let reason = assert_deny(&codex("tool-preflight", &root, &input));
    assert!(reason.is_empty() || reason.contains("MoAI"), "{reason}");
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &write_input(&root.join("src.rs"), &root),
        &env,
    ));
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &write_input(&home.join(".claude/settings.json"), &root),
        &env,
    ));
    assert!(!outside.exists(), "a hook never writes the file");
}

/// B3 is off unless the repo is listed; then a switch in a primary checkout with a linked
/// worktree is refused, and an `allow_commands` line passes.
#[test]
fn a_listed_shared_primary_checkout_is_protected() {
    let (_d, root) = repo("main");
    let home = root.join("home");
    let lane_dir = format!("../{}-b3-lane", root.file_name().unwrap().to_string_lossy());
    git(&root, &["worktree", "add", "-q", "-b", "lane", &lane_dir]);
    let env = [("HOME", home.to_str().unwrap())];
    let switch = pre_tool("git switch lane", &root);
    assert_allow(&dispatch("tool-preflight", &root, &switch, &env));
    config(
        &home,
        &format!(
            "[hooks.primary_checkout]\nenabled = true\nrepos = [\"{}\"]\nallow_commands = [\"git switch main\"]\n",
            root.display()
        ),
    );
    let reason = assert_deny(&dispatch("tool-preflight", &root, &switch, &env));
    assert!(reason.contains("primary checkout"), "{reason}");
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git switch main", &root),
        &env,
    ));
    assert_allow(&dispatch(
        "tool-preflight",
        &root,
        &pre_tool("git status", &root),
        &env,
    ));
    // The linked worktree itself is not the primary checkout.
    let lane = std::fs::canonicalize(root.join(&lane_dir)).unwrap();
    assert_allow(&dispatch(
        "tool-preflight",
        &lane,
        &pre_tool("git switch main", &lane),
        &env,
    ));
    git(&root, &["worktree", "remove", "--force", &lane_dir]);
}

/// A B9 config that turns the rule on, padded with an ASCII comment to exactly `len` bytes.
fn padded_b9_config(len: usize) -> String {
    let rule = "[hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\"]\n";
    let pad = len - rule.len() - 2;
    format!("{rule}#{}\n", "x".repeat(pad))
}

/// The config read stops at 65,536 bytes (pre-audit FAIL 6780344159 B1): a valid config of
/// exactly 65,536 bytes turns B9 on (deny), the same config one byte longer is off (allow).
#[test]
fn a_config_of_65536_bytes_is_read_and_65537_bytes_is_off() {
    let (_d, root) = repo("main");
    let home = root.join("home");
    let outside = root
        .parent()
        .unwrap()
        .join(format!("moai-b9-size-{}", std::process::id()));
    let input = write_input(&outside.join("x.txt"), &root);
    let env = [("HOME", home.to_str().unwrap())];
    let path = home.join(".moai-cli/config");
    config(&home, &padded_b9_config(65_536));
    assert_eq!(fs::metadata(&path).unwrap().len(), 65_536);
    let reason = assert_deny(&dispatch("tool-preflight", &root, &input, &env));
    assert!(reason.contains("outside the project"), "{reason}");
    config(&home, &padded_b9_config(65_537));
    assert_eq!(fs::metadata(&path).unwrap().len(), 65_537);
    assert_allow(&dispatch("tool-preflight", &root, &input, &env));
}

/// A config that is a symlink is not followed (pre-audit FAIL 6780344159 B1): the same text
/// denies as a regular file and is off when `config` only points at it.
#[test]
fn a_symlinked_config_is_off() {
    let (_d, root) = repo("main");
    let home = root.join("home");
    let outside = root
        .parent()
        .unwrap()
        .join(format!("moai-b9-link-{}", std::process::id()));
    let input = write_input(&outside.join("x.txt"), &root);
    let env = [("HOME", home.to_str().unwrap())];
    let text = "[hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\"]\n";
    config(&home, text);
    assert_deny(&dispatch("tool-preflight", &root, &input, &env));
    let path = home.join(".moai-cli/config");
    let target = home.join("elsewhere.toml");
    fs::write(&target, text).unwrap();
    fs::remove_file(&path).unwrap();
    std::os::unix::fs::symlink(&target, &path).unwrap();
    assert!(fs::symlink_metadata(&path)
        .unwrap()
        .file_type()
        .is_symlink());
    assert_allow(&dispatch("tool-preflight", &root, &input, &env));
}

/// The `config_error` lines of the engine log, in order.
fn config_errors(root: &Path) -> Vec<Value> {
    fs::read_to_string(engine_log(root))
        .unwrap_or_default()
        .lines()
        .map(|l| serde_json::from_str::<Value>(l).unwrap())
        .filter(|l| l["kind"] == json!("config_error"))
        .collect()
}

/// Whether this process can read a file it made unreadable (root can).
fn reads_through_mode_000() -> bool {
    let out = Command::new("id").arg("-u").output().unwrap();
    String::from_utf8_lossy(&out.stdout).trim() == "0"
}

/// N1 (pre-audit FAIL 6780344159, design 6777574227): an unusable config turns both rules
/// off and leaves exactly one `config_error` line with a fixed reason in the engine log — no
/// path, no config text. An absent config and a usable one leave none.
#[test]
fn an_unusable_config_is_off_and_logged_once_with_its_reason() {
    let (_d, root) = repo("main");
    fs::create_dir(root.join(".moai-cli")).unwrap();
    let home = root.join("home");
    let outside = root
        .parent()
        .unwrap()
        .join(format!("moai-b9-n1-{}", std::process::id()));
    let input = write_input(&outside.join("x.txt"), &root);
    let env = [("HOME", home.to_str().unwrap())];
    let path = home.join(".moai-cli/config");
    let text = "[hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\"]\n";

    assert_allow(&dispatch("tool-preflight", &root, &input, &env));
    config(&home, text);
    assert_deny(&dispatch("tool-preflight", &root, &input, &env));
    assert!(
        config_errors(&root).is_empty(),
        "absent and usable: no line"
    );

    let mut cases: Vec<(&str, Box<dyn Fn() + '_>)> = vec![
        (
            "invalid_toml",
            Box::new(|| {
                config(
                    &home,
                    "[hooks.outside_project_writes]\nenabled = true\nallow = [\n",
                )
            }),
        ),
        (
            "invalid_toml",
            Box::new(|| {
                fs::write(
                    &path,
                    b"[hooks.outside_project_writes]\nenabled = true # \xff\n",
                )
                .unwrap()
            }),
        ),
        (
            "too_large",
            Box::new(|| config(&home, &padded_b9_config(65_537))),
        ),
        (
            "not_regular_file",
            Box::new(|| {
                let target = home.join("elsewhere.toml");
                fs::write(&target, text).unwrap();
                fs::remove_file(&path).unwrap();
                std::os::unix::fs::symlink(&target, &path).unwrap();
            }),
        ),
        (
            "not_regular_file",
            Box::new(|| {
                fs::remove_file(&path).unwrap();
                fs::create_dir(&path).unwrap();
            }),
        ),
    ];
    if !reads_through_mode_000() {
        cases.push((
            "unreadable",
            Box::new(|| {
                use std::os::unix::fs::PermissionsExt;
                fs::remove_dir(&path).unwrap();
                config(&home, text);
                fs::set_permissions(&path, fs::Permissions::from_mode(0o000)).unwrap();
            }),
        ));
    }
    for (n, (reason, make)) in cases.iter().enumerate() {
        make();
        assert_allow(&dispatch("tool-preflight", &root, &input, &env));
        let lines = config_errors(&root);
        assert_eq!(lines.len(), n + 1, "{reason}: one line per run");
        let line = &lines[n];
        assert_eq!(line["reason"], json!(reason), "{line}");
        assert_eq!(line["host"], json!("claude"), "{line}");
        let mut keys: Vec<&String> = line.as_object().unwrap().keys().collect();
        keys.sort();
        assert_eq!(
            keys,
            [
                "host",
                "kind",
                "reason",
                "schema_version",
                "session",
                "unix_ms"
            ],
            "{line}"
        );
        let raw = line.to_string();
        assert!(!raw.contains(home.to_str().unwrap()), "no path: {raw}");
        assert!(!raw.contains("outside_project_writes"), "no text: {raw}");
    }
    assert!(!outside.exists(), "a hook never writes the file");
}
