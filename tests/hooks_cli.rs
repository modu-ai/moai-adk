//! rc.2 I1 #2238: host hook generation and the hook-receipt handler. Every case runs in a
//! temporary project and a temporary HOME; no real home and no host session is touched.

mod common;

use moai_cli::contract::{classify, Kind};
use moai_cli::hooks;
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const PERMISSION: &str = "test-permission-2238";

fn project() -> tempfile::TempDir {
    tempfile::TempDir::new().unwrap()
}

fn read_json(path: &Path) -> Value {
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

fn plan_apply(root: &Path, host: &str) -> (Value, Value) {
    let plan = hooks::plan(root, host, None).unwrap();
    let applied = hooks::apply(
        root,
        host,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    (plan, applied)
}

/// AC-1 (Claude): plan writes nothing, apply writes `hooks` and the receipt, re-plan is a no-op.
#[test]
fn claude_plan_apply_replan_is_noop() {
    let p = project();
    let plan = hooks::plan(p.path(), "claude", None).unwrap();
    assert_eq!(plan["action"], json!("create"));
    assert_eq!(plan["writes"], json!(false));
    assert_eq!(plan["path"], json!(".claude/settings.json"));
    assert!(!p.path().join(".claude").exists(), "plan must not write");

    let applied = hooks::apply(
        p.path(),
        "claude",
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    assert_eq!(applied["state"], json!("applied"));
    assert_eq!(applied["host_loaded"], json!("not_run"));
    // #2877 Q1: Claude Code also gets the hook engine's preflight guard (Codex stays at 3).
    assert_eq!(applied["hooks_generated"].as_array().unwrap().len(), 4);
    let settings = read_json(&p.path().join(".claude/settings.json"));
    assert_eq!(
        settings["hooks"],
        Value::Object(hooks::hooks_object("claude").unwrap())
    );
    // REQ-I1-02: the receipt records the generated hooks MoAI owns.
    let receipt = read_json(&p.path().join(".moai-cli/manifests/claude-settings.json"));
    assert_eq!(receipt["owned"]["hooks"], settings["hooks"]);

    let again = hooks::plan(p.path(), "claude", None).unwrap();
    assert_eq!(again["action"], json!("unchanged"));
    let before = fs::read(p.path().join(".claude/settings.json")).unwrap();
    let noop = hooks::apply(
        p.path(),
        "claude",
        None,
        again["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    assert_eq!(noop["state"], json!("unchanged"));
    assert_eq!(
        fs::read(p.path().join(".claude/settings.json")).unwrap(),
        before
    );
}

/// AC-1 (Codex): the same cycle for `.codex/hooks.json`, with the list in its own receipt.
#[test]
fn codex_plan_apply_replan_is_noop() {
    let p = project();
    let plan = hooks::plan(p.path(), "codex", None).unwrap();
    assert_eq!(plan["action"], json!("create"));
    assert!(!p.path().join(".codex").exists(), "plan must not write");
    let (_, applied) = plan_apply(p.path(), "codex");
    assert_eq!(applied["state"], json!("applied"));
    assert_eq!(applied["files_created"], json!([".codex/hooks.json"]));
    let file = read_json(&p.path().join(".codex/hooks.json"));
    assert_eq!(
        file["hooks"],
        Value::Object(hooks::hooks_object("codex").unwrap())
    );
    assert!(file["description"].as_str().unwrap().contains("MoAI"));
    let receipt = read_json(&p.path().join(".moai-cli/manifests/codex-hooks.json"));
    assert_eq!(receipt["hooks"].as_array().unwrap().len(), 3);
    assert_eq!(receipt["sha256"], applied["hooks_sha256_after"]);
    assert!(receipt.get("pending_sha256").is_none(), "pending cleared");

    let again = hooks::plan(p.path(), "codex", None).unwrap();
    assert_eq!(again["action"], json!("unchanged"));
    let noop = hooks::apply(
        p.path(),
        "codex",
        None,
        again["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    assert_eq!(noop["state"], json!("unchanged"));
}

/// The hook set itself (#2936 R9 S1, selection table of Issue #2936 comment 6881107273): three
/// events and no `Stop`, an explicit timeout on every handler, and bare command names (no
/// absolute path). `SessionStart` answers the session context for the model, `PreToolUse` on
/// Bash runs the claim guard (#2255 AC-2) and `PostToolUse` on Bash is a receipt, the same
/// commands for every host.
#[test]
fn hook_set_has_explicit_timeouts_and_bare_commands() {
    for host in hooks::HOSTS {
        let obj = hooks::hooks_object(host).unwrap();
        let events: Vec<&String> = obj.keys().collect();
        assert_eq!(events, ["SessionStart", "PreToolUse", "PostToolUse"]);
        assert!(obj.get("Stop").is_none(), "{host}: Stop has no reader");
        assert!(obj["SessionStart"][0].get("matcher").is_none());
        assert_eq!(obj["PreToolUse"][0]["matcher"], json!("Bash"));
        assert_eq!(obj["PostToolUse"][0]["matcher"], json!("Bash"));
        for (event, groups) in &obj {
            let handler = &groups[0]["hooks"][0];
            assert_eq!(handler["type"], json!("command"));
            assert_eq!(handler["timeout"], json!(hooks::HOOK_TIMEOUT_SECS));
            let command = handler["command"].as_str().unwrap();
            let expected = match event.as_str() {
                "PreToolUse" => "moai claim-verify . --hook".to_string(),
                "SessionStart" => {
                    format!("moai hook-dispatch --host {host} --event session-check")
                }
                _ => format!("moai hook-receipt --host {host}"),
            };
            assert_eq!(command, expected);
            assert!(!command.contains('/'), "{command}");
        }
    }
}

/// #2936 R9 S1 AC-6: the OMP coverage report names `session_stop` as the turn-end event (OMP
/// 18.18.7, extensions.md `session_stop`) and says it is not wired, instead of claiming OMP has
/// no turn-end event.
#[test]
fn omp_coverage_names_session_stop_and_says_it_is_not_wired() {
    let caps = moai_cli::hook_engine::capabilities("omp");
    let tv = &caps["turn-validation"];
    assert_eq!(tv["supported"], json!(false), "{tv}");
    assert_eq!(tv["event"], json!("session_stop"), "{tv}");
    let reason = tv["reason"].as_str().unwrap();
    assert!(reason.contains("not wired"), "{reason}");
    assert!(!reason.contains("no turn-end event"), "{reason}");
}

/// #2255 AC-2 (wiring): the command `hooks-apply` writes for `PreToolUse` is the guard, and
/// running exactly that command line with a host's PreToolUse input blocks a tokenless commit
/// in a lane worktree (exit 2, reason on stderr) and lets a non-commit through (exit 0).
#[cfg(unix)]
#[test]
fn generated_pre_tool_use_command_blocks_an_unverified_commit() {
    let dir = project();
    let root = fs::canonicalize(dir.path()).unwrap();
    let git = |args: &[&str]| {
        let full = [
            ["-c", "user.name=t", "-c", "user.email=t@example.invalid"].as_slice(),
            args,
        ]
        .concat();
        common::git::git_ok(&root, &full);
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
    // A lane worktree's clone has a live lane session (SPEC-20261117-commit-guard-users).
    common::register_lane(&root);

    for host in hooks::HOSTS {
        let obj = hooks::hooks_object(host).unwrap();
        let command = obj["PreToolUse"][0]["hooks"][0]["command"]
            .as_str()
            .unwrap()
            .to_string();
        let mut words = command.split_whitespace();
        assert_eq!(words.next(), Some("moai"));
        let args: Vec<&str> = words.collect();
        let run = |tool_command: &str| {
            let mut child = Command::new(crate::common::MOAI_BIN)
                .args(&args)
                .current_dir(&root)
                .env_remove("MOAI_FORGE_GH_BIN")
                .env("GIT_CONFIG_GLOBAL", "/dev/null")
                .env("GIT_CONFIG_SYSTEM", "/dev/null")
                .stdin(Stdio::piped())
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .spawn()
                .unwrap();
            let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                               "tool_input": {"command": tool_command}, "cwd": root});
            common::feed_stdin(&mut child, input.to_string().as_bytes());
            child.wait_with_output().unwrap()
        };
        let blocked = run("git commit -m work");
        assert_eq!(blocked.status.code(), Some(2), "{host}");
        assert!(
            String::from_utf8_lossy(&blocked.stderr).contains("claim verify: uncertain"),
            "{host}"
        );
        assert_eq!(run("cargo test").status.code(), Some(0), "{host}");
    }
}

/// A `hooks` key the user already set is a conflict: nothing is written, the file stays.
#[test]
fn claude_user_hooks_conflict_and_nothing_is_written() {
    let p = project();
    fs::create_dir(p.path().join(".claude")).unwrap();
    let user = "{\n  \"hooks\": {\"Stop\": [{\"hooks\": [{\"type\": \"command\", \"command\": \"say done\"}]}]}\n}\n";
    fs::write(p.path().join(".claude/settings.json"), user).unwrap();
    let plan = hooks::plan(p.path(), "claude", None).unwrap();
    assert_eq!(plan["action"], json!("conflict"));
    assert_eq!(plan["conflicts"], json!(["hooks"]));
    let err = hooks::apply(
        p.path(),
        "claude",
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");
    assert_eq!(
        fs::read_to_string(p.path().join(".claude/settings.json")).unwrap(),
        user
    );
    assert!(!p.path().join(".moai-cli").exists());
}

/// Generating hooks keeps a settings key MoAI already owns (here `outputStyle`).
#[test]
fn claude_hooks_keep_other_owned_keys() {
    let p = project();
    let style = json!({"outputStyle": "MoAI"}).as_object().unwrap().clone();
    let plan = moai_cli::claude::settings_plan(p.path(), &style).unwrap();
    moai_cli::claude::settings_apply(
        p.path(),
        &style,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    plan_apply(p.path(), "claude");
    let settings = read_json(&p.path().join(".claude/settings.json"));
    assert_eq!(settings["outputStyle"], json!("MoAI"));
    assert!(settings["hooks"].is_object());
    let receipt = read_json(&p.path().join(".moai-cli/manifests/claude-settings.json"));
    assert_eq!(receipt["owned"]["outputStyle"], json!("MoAI"));
}

/// A `.codex/hooks.json` MoAI never wrote, or one edited after apply, is not overwritten.
#[test]
fn codex_unowned_or_edited_file_requires_review() {
    let p = project();
    fs::create_dir(p.path().join(".codex")).unwrap();
    fs::write(p.path().join(".codex/hooks.json"), "{\"hooks\": {}}\n").unwrap();
    let err = hooks::plan(p.path(), "codex", None).unwrap_err();
    assert_eq!(classify(&err), Kind::Conflict, "{err}");

    let q = project();
    plan_apply(q.path(), "codex");
    fs::write(q.path().join(".codex/hooks.json"), "{\"hooks\": {}}\n").unwrap();
    let err = hooks::plan(q.path(), "codex", None).unwrap_err();
    assert!(err.contains("requires preservation review"), "{err}");
    assert_eq!(classify(&err), Kind::Conflict);
}

/// An apply interrupted after the pending receipt but before the file write resolves on the
/// next plan: the receipt still owns the old sha256, so the cycle completes.
#[test]
fn codex_pending_receipt_resolves_on_replan() {
    let p = project();
    plan_apply(p.path(), "codex");
    let receipt_path = p.path().join(".moai-cli/manifests/codex-hooks.json");
    let mut receipt = read_json(&receipt_path);
    receipt["pending_sha256"] = json!(format!("sha256:{}", "0".repeat(64)));
    fs::write(&receipt_path, serde_json::to_vec_pretty(&receipt).unwrap()).unwrap();
    let plan = hooks::plan(p.path(), "codex", None).unwrap();
    assert_eq!(plan["action"], json!("unchanged"));
    let done = hooks::apply(
        p.path(),
        "codex",
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    assert_eq!(done["state"], json!("unchanged"));
    assert!(read_json(&receipt_path).get("pending_sha256").is_none());
}

/// REQ-I1-01 isolation: the home directory itself, or a project inside a host's home
/// configuration, is refused by plan and apply alike.
/// #2696 C4b (audit #2848 B1): `~/.moai` is refused only when it is this generation's HOME
/// state folder (marker present), and never under its worktree folder — where W2 creates
/// worktrees and the Go generation keeps its own. A Go `~/.moai` (no marker) is not ours.
#[test]
fn home_moai_is_refused_only_as_this_generations_state_outside_worktrees() {
    let home = project();
    let plan_at = |rel: &str| {
        let dir = home.path().join(rel);
        fs::create_dir_all(&dir).unwrap();
        hooks::plan(&dir, "claude", Some(home.path()))
    };
    // A Go `~/.moai` (no marker): nothing under it is refused.
    fs::create_dir_all(home.path().join(".moai/state")).unwrap();
    fs::write(home.path().join(".moai/state/active-sessions.json"), "{}").unwrap();
    for rel in [
        ".moai/worktrees/demo-1",
        ".moai/worktrees/projects/demo/agent-1",
        ".moai/other/p",
    ] {
        assert!(plan_at(rel).is_ok(), "go HOME, {rel}");
    }
    // This generation's `~/.moai`: refused, except under its worktree folder.
    fs::remove_dir_all(home.path().join(".moai/state")).unwrap();
    fs::write(
        home.path().join(".moai/moai-cli.json"),
        r#"{"schema_version":1,"generation":"moai-cli"}"#,
    )
    .unwrap();
    for rel in [
        ".moai/worktrees/demo-2",
        ".moai/worktrees/projects/demo/agent-2",
    ] {
        assert!(plan_at(rel).is_ok(), "marked HOME, {rel}");
    }
    for rel in [".moai/state/p", ".moai/agents/p"] {
        let err = plan_at(rel).unwrap_err();
        assert!(err.contains("home directory refused"), "{rel}: {err}");
    }
}

#[test]
fn home_directory_is_refused() {
    let home = project();
    for host in hooks::HOSTS {
        let err = hooks::plan(home.path(), host, Some(home.path())).unwrap_err();
        assert_eq!(classify(&err), Kind::InvalidInput, "{err}");
        let err = hooks::apply(home.path(), host, Some(home.path()), "x", PERMISSION).unwrap_err();
        assert!(err.contains("home directory refused"), "{err}");
    }
    for dir in [".claude", ".codex", ".omp", ".agents", ".moai-cli"] {
        let inside = home.path().join(dir).join("p");
        fs::create_dir_all(&inside).unwrap();
        let err = hooks::plan(&inside, "codex", Some(home.path())).unwrap_err();
        assert!(err.contains("home directory refused"), "{dir}: {err}");
    }
    // A project elsewhere under the home (not in a config directory) is allowed.
    let elsewhere = home.path().join("work");
    fs::create_dir(&elsewhere).unwrap();
    hooks::plan(&elsewhere, "codex", Some(home.path())).unwrap();
    assert!(!home.path().join(".codex/hooks.json").exists());
}

/// Refusal branches: unknown host, stale digest, missing permission reference.
#[test]
fn refusals_are_classified() {
    let p = project();
    let err = hooks::plan(p.path(), "gemini", None).unwrap_err();
    assert_eq!(classify(&err), Kind::InvalidInput, "{err}");
    for host in hooks::HOSTS {
        let err = hooks::apply(p.path(), host, None, "sha256:stale", PERMISSION).unwrap_err();
        assert_eq!(classify(&err), Kind::StalePlan, "{host}: {err}");
        let plan = hooks::plan(p.path(), host, None).unwrap();
        let err = hooks::apply(
            p.path(),
            host,
            None,
            plan["plan_digest"].as_str().unwrap(),
            " ",
        )
        .unwrap_err();
        assert_eq!(classify(&err), Kind::PermissionReferenceMissing, "{err}");
    }
    assert!(!p.path().join(".claude").exists());
    assert!(!p.path().join(".codex").exists());
}

/// hook-receipt writes nothing without `.moai-cli/`, and never fails.
#[test]
fn receipt_skips_without_state_dir() {
    let p = project();
    let v = hooks::receipt("claude", br#"{"hook_event_name":"Stop"}"#, p.path(), 1);
    assert_eq!(v["recorded"], json!(false));
    assert!(
        fs::read_dir(p.path()).unwrap().next().is_none(),
        "nothing created"
    );
    for bad in [&b"not json"[..], b"[1,2]", b""] {
        assert_eq!(
            hooks::receipt("claude", bad, p.path(), 1)["recorded"],
            json!(false)
        );
    }
    assert_eq!(
        hooks::receipt("gemini", br#"{}"#, p.path(), 1)["recorded"],
        json!(false)
    );
}

/// REQ-I1-02 / secrets 0: only allow-listed fields reach the log; tool input, responses and
/// prompts never do, and a secret-shaped allow-listed value is dropped.
#[test]
fn receipt_keeps_only_allow_listed_fields() {
    let p = project();
    fs::create_dir(p.path().join(".moai-cli")).unwrap();
    let input = json!({
        "hook_event_name": "PostToolUse",
        "tool_name": "Bash",
        "session_id": "s-1",
        "cwd": "/somewhere",
        "tool_input": {"command": "echo GITHUB_TOKEN=ghp_secretvalue"},
        "tool_response": {"stdout": "api_key=abc"},
        "prompt": "my password is hunter2",
    });
    let v = hooks::receipt("codex", input.to_string().as_bytes(), p.path(), 42);
    assert_eq!(v["recorded"], json!(true), "{v}");
    let log = fs::read_to_string(p.path().join(".moai-cli/receipts/hooks.jsonl")).unwrap();
    let line: Value = serde_json::from_str(log.trim()).unwrap();
    assert_eq!(
        line,
        json!({"schema_version": 1, "host": "codex", "unix_ms": 42,
               "hook_event_name": "PostToolUse", "tool_name": "Bash"})
    );
    for leaked in [
        "ghp_",
        "api_key",
        "hunter2",
        "/somewhere",
        "s-1",
        "tool_input",
    ] {
        assert!(!log.contains(leaked), "{leaked} leaked: {log}");
    }

    let secret = json!({"hook_event_name": "Stop", "tool_name": "token=ghp_x"});
    let v = hooks::receipt("claude", secret.to_string().as_bytes(), p.path(), 43);
    assert_eq!(v["dropped"], json!(["tool_name"]));
    let log = fs::read_to_string(p.path().join(".moai-cli/receipts/hooks.jsonl")).unwrap();
    assert_eq!(log.lines().count(), 2);
    assert!(!log.contains("ghp_x"));
}

/// The enforced limits, at their boundaries: a 128-character field is kept and 129 is dropped;
/// a log at 4 MiB takes no further line and is never truncated.
#[test]
fn receipt_limits_hold_at_their_boundaries() {
    let p = project();
    fs::create_dir(p.path().join(".moai-cli")).unwrap();
    let at = json!({"hook_event_name": "Stop", "tool_name": "a".repeat(128)});
    let v = hooks::receipt("claude", at.to_string().as_bytes(), p.path(), 1);
    assert_eq!(v["fields"]["tool_name"], json!("a".repeat(128)));
    let over = json!({"hook_event_name": "Stop", "tool_name": "a".repeat(129)});
    let v = hooks::receipt("claude", over.to_string().as_bytes(), p.path(), 2);
    assert_eq!(v["dropped"], json!(["tool_name"]));
    let ctl = json!({"hook_event_name": "Stop\nx"});
    let v = hooks::receipt("claude", ctl.to_string().as_bytes(), p.path(), 3);
    assert_eq!(v["dropped"], json!(["hook_event_name"]));

    let log = p.path().join(".moai-cli/receipts/hooks.jsonl");
    let f = fs::OpenOptions::new().write(true).open(&log).unwrap();
    f.set_len(4 * 1024 * 1024 - 1).unwrap();
    let v = hooks::receipt("claude", br#"{"hook_event_name":"Stop"}"#, p.path(), 4);
    assert_eq!(v["recorded"], json!(true), "below the limit still appends");
    let full = fs::metadata(&log).unwrap().len();
    assert!(full >= 4 * 1024 * 1024);
    let v = hooks::receipt("claude", br#"{"hook_event_name":"Stop"}"#, p.path(), 5);
    assert_eq!(v["recorded"], json!(false));
    assert_eq!(v["reason"], json!("receipt log full"));
    assert_eq!(
        fs::metadata(&log).unwrap().len(),
        full,
        "never truncated or grown"
    );
}

#[cfg(unix)]
#[test]
fn receipt_refuses_symlinked_log() {
    let p = project();
    fs::create_dir_all(p.path().join(".moai-cli/receipts")).unwrap();
    let outside = p.path().join("outside.txt");
    fs::write(&outside, "").unwrap();
    std::os::unix::fs::symlink(&outside, p.path().join(".moai-cli/receipts/hooks.jsonl")).unwrap();
    let v = hooks::receipt("claude", br#"{"hook_event_name":"Stop"}"#, p.path(), 1);
    assert_eq!(v["recorded"], json!(false));
    assert_eq!(fs::read_to_string(&outside).unwrap(), "");
}

fn cli(args: &[&str], cwd: &Path, home: &Path, stdin: &[u8]) -> (i32, Value) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        // A PATH with no `moai`: the gate (#3004 P4-2) reads the PATH, and the machine's own
        // `moai` must not decide this test.
        .env("PATH", "/usr/bin:/bin")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin);
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        serde_json::from_slice(&out.stdout).unwrap(),
    )
}

/// The CLI path: hooks-plan / hooks-apply envelopes, the HOME refusal through the binary, and
/// hook-receipt exiting 0 even on input it cannot use.
#[test]
fn cli_round_trip_in_isolated_home() {
    let home = project();
    let p = project();
    let root = p.path().to_str().unwrap();
    let (code, plan) = cli(
        &["hooks-plan", root, "--host", "codex"],
        p.path(),
        home.path(),
        b"",
    );
    assert_eq!(code, 0, "{plan}");
    assert_eq!(plan["operation"], json!("hooks.plan"));
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let (code, applied) = cli(
        &[
            "hooks-apply",
            root,
            "--host",
            "codex",
            "--plan-digest",
            &digest,
            "--permission-reference",
            PERMISSION,
        ],
        p.path(),
        home.path(),
        b"",
    );
    assert_eq!(code, 0, "{applied}");
    assert_eq!(applied["result"]["state"], json!("applied"));
    assert!(!home.path().join(".codex").exists(), "home untouched");

    let home_root = home.path().to_str().unwrap();
    let (code, refused) = cli(
        &["hooks-plan", home_root, "--host", "claude"],
        p.path(),
        home.path(),
        b"",
    );
    assert_ne!(code, 0);
    assert_eq!(
        refused["error"]["kind"],
        json!("invalid_input"),
        "{refused}"
    );

    // Hook mode prints nothing (#2348); the receipt log is the record.
    fs::create_dir(p.path().join(".moai-cli/receipts")).ok();
    let log = p.path().join(".moai-cli/receipts/hooks.jsonl");
    let lines = || {
        fs::read_to_string(&log)
            .map(|s| s.lines().count())
            .unwrap_or(0)
    };
    let (code, stdout) = cli_raw(
        &["hook-receipt", "--host", "codex"],
        p.path(),
        home.path(),
        br#"{"hook_event_name":"SessionStart","source":"startup"}"#,
    );
    assert_eq!((code, stdout.as_str()), (0, ""));
    assert_eq!(lines(), 1);
    let (code, stdout) = cli_raw(
        &["hook-receipt", "--host", "codex"],
        p.path(),
        home.path(),
        b"{oops",
    );
    assert_eq!((code, stdout.as_str()), (0, ""));
    assert_eq!(lines(), 1, "a cut-off object records nothing");
}

fn cli_raw(args: &[&str], cwd: &Path, home: &Path, stdin: &[u8]) -> (i32, String) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin);
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// #2239 × #2238: the user-facing settings op never writes `hooks` itself, and changing
/// `outputStyle` after hooks-apply keeps the generated hook set instead of removing it.
#[test]
fn claude_settings_op_neither_writes_nor_drops_hooks() {
    let p = project();
    let raw = json!({"hooks": {}}).as_object().unwrap().clone();
    let err = moai_cli::claude::user_settings_desired(p.path(), &raw).unwrap_err();
    assert_eq!(classify(&err), Kind::InvalidInput, "{err}");

    plan_apply(p.path(), "claude");
    let style = json!({"outputStyle": "MoAI"}).as_object().unwrap().clone();
    let desired = moai_cli::claude::user_settings_desired(p.path(), &style).unwrap();
    let plan = moai_cli::claude::settings_plan(p.path(), &desired).unwrap();
    moai_cli::claude::settings_apply(
        p.path(),
        &desired,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    let settings = read_json(&p.path().join(".claude/settings.json"));
    assert_eq!(settings["outputStyle"], json!("MoAI"));
    assert_eq!(
        settings["hooks"],
        Value::Object(hooks::hooks_object("claude").unwrap())
    );
    assert_eq!(
        hooks::plan(p.path(), "claude", None).unwrap()["action"],
        json!("unchanged")
    );
}

// ---- #2633: OMP — the hooks file is an extension module MoAI owns ----

#[test]
fn omp_plan_apply_writes_the_module_and_replan_is_noop() {
    let p = project();
    let plan = hooks::plan(p.path(), "omp", None).unwrap();
    assert_eq!(plan["action"], json!("create"));
    assert_eq!(plan["path"], json!(".omp/extensions/moai-hooks.ts"));
    assert!(!p.path().join(".omp").exists(), "plan must not write");
    let (_, applied) = plan_apply(p.path(), "omp");
    assert_eq!(applied["state"], json!("applied"));
    assert_eq!(
        applied["files_created"],
        json!([".omp/extensions/moai-hooks.ts"])
    );
    let module = fs::read_to_string(p.path().join(".omp/extensions/moai-hooks.ts")).unwrap();
    assert_eq!(module, hooks::omp_module("moai"));
    let receipt = read_json(&p.path().join(".moai-cli/manifests/omp-hooks.json"));
    assert_eq!(receipt["hooks"].as_array().unwrap().len(), 3);
    assert_eq!(receipt["sha256"], applied["hooks_sha256_after"]);
    assert!(applied["trust_required"].as_str().unwrap().contains("OMP"));
    let again = hooks::plan(p.path(), "omp", None).unwrap();
    assert_eq!(again["action"], json!("unchanged"));
    assert_eq!(again["engine_capabilities"]["host"], json!("omp"));
    let st = hooks::status(p.path(), "omp", None, "").unwrap();
    assert_eq!(st["configured"], json!("yes"), "{st}");
    // A user's edit is never overwritten and makes status drifted.
    fs::write(p.path().join(".omp/extensions/moai-hooks.ts"), "// mine\n").unwrap();
    let err = hooks::plan(p.path(), "omp", None).unwrap_err();
    assert!(
        err.contains(".omp/extensions/moai-hooks.ts requires preservation review"),
        "{err}"
    );
    assert_eq!(classify(&err), Kind::Conflict);
    let st = hooks::status(p.path(), "omp", None, "").unwrap();
    assert_eq!(st["configured"], json!("drifted"), "{st}");
}

/// The module calls the engine with the bare binary, guards only `bash`, appends to the
/// system prompt instead of replacing it, and never blocks on its own failure.
#[test]
fn omp_module_wiring_is_pinned() {
    let module = hooks::omp_module("moai");
    let m = module.as_str();
    assert!(m.contains(r#"Bun.spawn(["moai", "hook-dispatch", "--host", "omp", "--event", event]"#));
    // Both the preflight and the observation handler stop at a non-bash tool (mutation H5).
    for handler in ["tool_call", "tool_result"] {
        let start = m.find(&format!(r#"pi.on("{handler}""#)).unwrap();
        let body = &m[start..];
        let first_line = body.lines().nth(1).unwrap().trim();
        assert_eq!(
            first_line, r#"if (event.toolName !== "bash") return;"#,
            "{handler}"
        );
    }
    assert!(m.contains("systemPrompt: [...event.systemPrompt, answer.context]"));
    assert!(m.contains("if (code !== 0 || out.trim() === \"\") return null;"));
    assert!(m.contains("catch {\n\t\treturn null;"));
    assert!(m.contains("const TIMEOUT_MS = 10000;"));
    assert!(
        !m.contains("turn-validation") && !m.contains("session_stop"),
        "the module does not subscribe to OMP's session_stop (the turn-validation rules are empty)"
    );
}

#[cfg(unix)]
#[test]
fn omp_symlinked_folders_are_refused() {
    for linked in [".omp", ".omp/extensions"] {
        let p = project();
        let outside = project();
        let link = p.path().join(linked);
        fs::create_dir_all(link.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(outside.path(), &link).unwrap();
        assert!(hooks::plan(p.path(), "omp", None).is_err(), "{linked}");
        assert!(
            fs::read_dir(outside.path()).unwrap().next().is_none(),
            "{linked}: wrote through the link"
        );
    }
}

// ---- SPEC-20261119-enforce-merge-confirm: a direct pull request merge asks first ----

/// Commands that merge a pull request outside the gate.
const DIRECT_MERGES: &[&str] = &[
    "gh pr merge 123",
    "gh pr merge 123 --auto --merge --match-head-commit abc",
    "gh pr merge --squash",
    "gh pr merge 5 -R o/r",
    "gh api -X PUT repos/o/r/pulls/123/merge",
    "gh api -XPUT repos/o/r/pulls/9/merge -f merge_method=merge",
    "gh api --method PUT /repos/o/r/pulls/9/merge",
    "gh api --method=put repos/o/r/pulls/9/merge",
    "gh api graphql -f query='mutation { mergePullRequest(input: {pullRequestId: \"x\"}) { clientMutationId } }'",
    "gh api graphql -f query='mutation { enablePullRequestAutoMerge(input: {pullRequestId: \"x\"}) { clientMutationId } }'",
    "echo hi && gh pr merge 5",
    "FOO=1 gh pr merge 5",
    "sh -c 'gh pr merge 5'",
    // Compound statements: the command after a keyword is still a command.
    "if true; then gh pr merge 5; fi",
    "if true; then echo a; else gh pr merge 5; fi",
    "for i in 1 2; do gh pr merge 5; done",
    "while true; do gh pr merge 5; done",
    "until false; do gh pr merge 5; done",
    "! gh pr merge 5",
    "{ gh pr merge 5; }",
    // Two keywords in a row: both are peeled.
    "if true; then ! gh pr merge 5; fi",
    "while true; do { gh pr merge 5; }; done",
];

/// A temporary work tree (a hand-made `.git` folder) with or without a `.moai` folder.
fn merge_repo(moai: bool) -> (tempfile::TempDir, std::path::PathBuf) {
    let tmp = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    fs::create_dir_all(root.join(".git")).unwrap();
    fs::write(root.join(".git/HEAD"), "ref: refs/heads/main\n").unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    // The folder the engine log is written under.
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    if moai {
        fs::create_dir_all(root.join(".moai")).unwrap();
    }
    (tmp, root)
}

/// The HOME hook config of the temporary project, with `text` as its content.
fn merge_config(root: &Path, text: &str) {
    let path = moai_cli::home_root::path(&root.join("home")).join("hook-config.toml");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// The rule's mode set to `enforce` in the HOME hook config.
fn merge_enforce(root: &Path) {
    merge_config(root, "[hooks.modes]\nmerge_confirm = \"enforce\"\n");
}

/// The lines of the engine log of the temporary project.
fn merge_log(root: &Path) -> Vec<Value> {
    fs::read_to_string(root.join(".moai-cli/receipts/hook-engine.jsonl"))
        .unwrap_or_default()
        .lines()
        .map(|l| serde_json::from_str(l).unwrap())
        .collect()
}

/// The answer to one Bash call, as a host would get it: (exit code, decision, reason); the
/// decision is empty for an allowed call.
fn merge_preflight(host: &str, root: &Path, command: &str) -> (i32, String, String) {
    merge_preflight_as(host, root, command, "s")
}

fn merge_preflight_as(
    host: &str,
    root: &Path,
    command: &str,
    session: &str,
) -> (i32, String, String) {
    let input = json!({"session_id": session, "hook_event_name": "PreToolUse",
        "tool_name": "Bash", "tool_use_id": "t1",
        "tool_input": {"command": command}, "cwd": root});
    let mut child = Command::new(common::MOAI_BIN)
        .args(["hook-dispatch", "--host", host, "--event", "tool-preflight"])
        .current_dir(root)
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
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    let v: Value = serde_json::from_str(&stdout).unwrap_or(Value::Null);
    let h = &v["hookSpecificOutput"];
    (
        out.status.code().unwrap(),
        h["permissionDecision"].as_str().unwrap_or("").to_string(),
        h["permissionDecisionReason"]
            .as_str()
            .unwrap_or("")
            .to_string(),
    )
}

/// REQ-1, REQ-4: in a MoAI project every direct merge asks on Claude, and is refused on Codex,
/// which cannot ask.
#[test]
fn merge_confirm_asks_for_a_direct_merge_in_a_moai_project() {
    let (_tmp, root) = merge_repo(true);
    merge_enforce(&root);
    for (n, command) in DIRECT_MERGES.iter().enumerate() {
        // One session per command: the third refusal in one session is released.
        let session = format!("s{n}");
        let (code, decision, reason) = merge_preflight_as("claude", &root, command, &session);
        assert_eq!((code, decision.as_str()), (0, "ask"), "{command}: {reason}");
        assert!(reason.contains("moai gate-merge"), "{command}: {reason}");
        let (code, decision, reason) = merge_preflight_as("codex", &root, command, &session);
        assert_eq!(
            (code, decision.as_str()),
            (0, "deny"),
            "{command}: {reason}"
        );
        assert!(reason.contains("Codex cannot ask"), "{command}: {reason}");
    }
}

/// REQ-2: the gate itself, reads, comments and anything that only mentions a merge go through.
#[test]
fn merge_confirm_leaves_the_gate_and_reads_alone() {
    let (_tmp, root) = merge_repo(true);
    merge_enforce(&root);
    for command in [
        "moai gate-merge --pr 5 --permission-reference x",
        "moai gate-loop --fetch",
        "gh pr view 5",
        "gh pr list --state open",
        "gh pr comment 5 --body ok",
        "gh pr checks 5",
        "gh pr create --title t --body b",
        "gh api repos/o/r/pulls/5/merge",
        "gh api -X GET repos/o/r/pulls/5/merge",
        "gh api repos/o/r/pulls/5",
        // A PUT that does not merge a pull request.
        "gh api -X PUT repos/o/r/pulls/5/labels",
        "gh api -X PUT repos/o/r/branches/merge",
        "gh api graphql -f query='query { viewer { login } }'",
        "git merge feature",
        "echo gh pr merge 5",
        "gh issue comment 5 -b 'gh pr merge 5'",
        "grep -n 'gh pr merge' README.md",
        // Keywords of compound statements do not turn other commands into merges.
        "if true; then echo gh pr merge 5; fi",
        "if gh pr view 5; then gh pr comment 5 --body ok; fi",
        "for i in 1 2; do gh pr checks $i; done",
        "! gh pr view 5",
        "{ git merge feature; }",
    ] {
        let (code, decision, reason) = merge_preflight("claude", &root, command);
        assert_eq!(
            (code, decision.as_str(), reason.as_str()),
            (0, "", ""),
            "{command}"
        );
    }
}

/// REQ-3: a work tree without a `.moai` folder is not a MoAI project; nothing is asked there.
#[test]
fn merge_confirm_is_off_outside_a_moai_project() {
    let (_tmp, root) = merge_repo(false);
    merge_enforce(&root);
    for command in DIRECT_MERGES {
        for host in ["claude", "codex"] {
            let (code, decision, reason) = merge_preflight(host, &root, command);
            assert_eq!(
                (code, decision.as_str(), reason.as_str()),
                (0, "", ""),
                "{host} {command}"
            );
        }
    }
    // Nothing was written either: the rule did not apply.
    assert_eq!(merge_log(&root), Vec::<Value>::new());
}

/// REQ-6, REQ-7, REQ-8: no setting, or a setting that is not one of the three words, is
/// observe: the command goes through and one `would_block` line (not a `deny`) says why. Off
/// goes through and writes nothing.
#[test]
fn merge_confirm_mode_defaults_to_observe_and_off_is_silent() {
    for config in [
        None,
        Some(""),
        Some("[hooks.modes]\n"),
        Some("[hooks.modes]\nmerge_confirm = \"yes\"\n"),
        Some("[hooks.modes]\nmerge_confirm = 3\n"),
        Some("[hooks.modes]\nother_rule = \"enforce\"\n"),
        Some("this is = not [valid toml"),
        Some("[hooks.modes]\nmerge_confirm = \"observe\"\n"),
    ] {
        let (_tmp, root) = merge_repo(true);
        if let Some(text) = config {
            merge_config(&root, text);
        }
        let (code, decision, reason) = merge_preflight("claude", &root, "gh pr merge 987654");
        assert_eq!(
            (code, decision.as_str(), reason.as_str()),
            (0, "", ""),
            "{config:?}"
        );
        let lines = merge_log(&root);
        let blocks: Vec<&Value> = lines
            .iter()
            .filter(|l| l["kind"] == json!("would_block"))
            .collect();
        assert_eq!(blocks.len(), 1, "{config:?}: {lines:?}");
        assert_eq!(blocks[0]["rule"], json!("merge_confirm"));
        assert!(blocks[0]["reason"]
            .as_str()
            .unwrap()
            .contains("moai gate-merge"));
        // The string fields only: a time or a count may hold any digits by chance.
        let text: String = lines
            .iter()
            .flat_map(|l| l.as_object().unwrap().values())
            .filter_map(|v| v.as_str())
            .collect::<Vec<_>>()
            .join("\n");
        for leak in ["gh pr merge", "987654", root.to_str().unwrap()] {
            assert!(!text.contains(leak), "{leak} leaked into {text}");
        }
        assert!(
            lines.iter().all(|l| l["kind"] != json!("deny")),
            "an observed call is not a deny: {lines:?}"
        );
    }
    let (_tmp, root) = merge_repo(true);
    merge_config(&root, "[hooks.modes]\nmerge_confirm = \"off\"\n");
    for host in ["claude", "codex"] {
        let (code, decision, reason) = merge_preflight(host, &root, "gh pr merge 987654");
        assert_eq!(
            (code, decision.as_str(), reason.as_str()),
            (0, "", ""),
            "{host}"
        );
    }
    assert_eq!(merge_log(&root), Vec::<Value>::new());
}

/// REQ-9, REQ-10, REQ-11: where enforce refuses (Codex cannot ask), the third refusal of the
/// same rule in the same session goes through and says so; another session starts again; the
/// ask on Claude is never counted; and a refusal that cannot be written down goes through.
#[test]
fn merge_confirm_strikes_release_the_third_refusal() {
    let (_tmp, root) = merge_repo(true);
    merge_enforce(&root);
    let merge = "gh pr merge 123";
    for n in 1..=2 {
        let (code, decision, _) = merge_preflight_as("codex", &root, merge, "one");
        assert_eq!((code, decision.as_str()), (0, "deny"), "refusal {n}");
    }
    let lines = merge_log(&root);
    let kinds = |kind: &str| lines.iter().filter(|l| l["kind"] == json!(kind)).count();
    assert_eq!(
        (kinds("rule_block"), kinds("rule_release")),
        (2, 0),
        "{lines:?}"
    );
    assert!(lines
        .iter()
        .filter(|l| l["kind"] == json!("rule_block"))
        .all(|l| l["rule"] == json!("merge_confirm")));
    // The third and the fourth go through, each leaving a release line with the reason.
    for _ in 0..2 {
        let (code, decision, reason) = merge_preflight_as("codex", &root, merge, "one");
        assert_eq!(
            (code, decision.as_str(), reason.as_str()),
            (0, "", ""),
            "released"
        );
    }
    let lines = merge_log(&root);
    let releases: Vec<&Value> = lines
        .iter()
        .filter(|l| l["kind"] == json!("rule_release"))
        .collect();
    assert_eq!(releases.len(), 2, "{lines:?}");
    assert_eq!(releases[0]["rule"], json!("merge_confirm"));
    assert!(releases[0]["reason"]
        .as_str()
        .unwrap()
        .contains("moai gate-merge"));
    // Another session is counted on its own.
    let (code, decision, _) = merge_preflight_as("codex", &root, merge, "two");
    assert_eq!((code, decision.as_str()), (0, "deny"));
    // The ask on Claude (a person answers it) is not a refusal and is never counted.
    for _ in 0..4 {
        let (code, decision, _) = merge_preflight_as("claude", &root, merge, "three");
        assert_eq!((code, decision.as_str()), (0, "ask"));
    }
    let lines = merge_log(&root);
    let claude_blocks = lines
        .iter()
        .filter(|l| l["kind"] == json!("rule_block") && l["host"] == json!("claude"))
        .count();
    assert_eq!(claude_blocks, 0, "{lines:?}");

    // No folder to write the line into: the refusal could not be counted, so it goes through.
    fs::remove_dir_all(root.join(".moai-cli")).unwrap();
    let (code, decision, reason) = merge_preflight_as("codex", &root, merge, "four");
    assert_eq!(
        (code, decision.as_str(), reason.as_str()),
        (0, "", ""),
        "unwritable"
    );
}

/// REQ-9: only this rule's `rule_block` lines of this session count: another rule's block and
/// this rule's own release line are not strikes.
#[test]
fn merge_confirm_counts_only_its_own_blocks() {
    use std::io::Write;
    let (_tmp, root) = merge_repo(true);
    merge_enforce(&root);
    let merge = "gh pr merge 123";
    let (_, decision, _) = merge_preflight_as("codex", &root, merge, "one");
    assert_eq!(decision, "deny", "first refusal");
    let first = merge_log(&root).remove(0);
    let session = first["session"].clone();
    let mut log = fs::OpenOptions::new()
        .append(true)
        .open(root.join(".moai-cli/receipts/hook-engine.jsonl"))
        .unwrap();
    for (kind, rule) in [
        ("rule_block", "other_rule"),
        ("rule_release", "merge_confirm"),
    ] {
        let line = json!({"schema_version": 1, "kind": kind, "rule": rule, "reason": "x",
            "host": "codex", "session": session, "call": null, "unix_ms": 1});
        writeln!(log, "{line}").unwrap();
    }
    // One strike of this rule so far: the next refusal still refuses, the one after is released.
    let (_, decision, _) = merge_preflight_as("codex", &root, merge, "one");
    assert_eq!(decision, "deny", "second refusal");
    let (code, decision, _) = merge_preflight_as("codex", &root, merge, "one");
    assert_eq!((code, decision.as_str()), (0, ""), "third is released");
}

/// One Bash call to the preflight guard as `host` sends it; `session` is the `session_id` field,
/// or none (OMP sends none). Returns the exit code and the raw answer.
fn merge_raw(host: &str, root: &Path, command: &str, session: Option<&str>) -> (i32, String) {
    let mut input = json!({"hook_event_name": "PreToolUse",
        "tool_name": "Bash", "tool_use_id": "t1",
        "tool_input": {"command": command}, "cwd": root});
    if let Some(s) = session {
        input["session_id"] = json!(s);
    }
    let mut child = Command::new(common::MOAI_BIN)
        .args(["hook-dispatch", "--host", host, "--event", "tool-preflight"])
        .current_dir(root)
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
    (
        out.status.code().unwrap(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// True when a raw answer refuses the call: Codex's `deny`, OMP's `block`.
fn merge_refused(answer: &str) -> bool {
    answer.contains("\"permissionDecision\":\"deny\"") || answer.contains("\"block\":true")
}

/// #3216-style audit of S4, B1: a host that sends no session id (OMP) or an empty one (Codex)
/// has no session to count refusals in, so the refusal is never released: every refusal stays
/// a refusal, in every later call, and each is still written down.
#[test]
fn merge_confirm_without_a_session_id_is_never_released() {
    for host in ["omp", "codex"] {
        for session in [None, Some("")] {
            let (_tmp, root) = merge_repo(true);
            merge_enforce(&root);
            for n in 1..=5 {
                let (code, answer) = merge_raw(host, &root, "gh pr merge 123", session);
                assert_eq!(code, 0, "{host} {session:?} call {n}");
                assert!(
                    merge_refused(&answer),
                    "{host} {session:?} call {n} was released: {answer}"
                );
            }
            let lines = merge_log(&root);
            let kinds = |kind: &str| lines.iter().filter(|l| l["kind"] == json!(kind)).count();
            assert_eq!(
                (kinds("rule_block"), kinds("rule_release")),
                (5, 0),
                "{host} {session:?}: {lines:?}"
            );
        }
    }
}

/// B2, mutant M4: `would_block` lines (observe) are not refusals and are not counted: after
/// three observed calls the first enforced call is refused, the second too, the third released.
#[test]
fn merge_confirm_observed_calls_are_not_strikes() {
    let (_tmp, root) = merge_repo(true);
    let merge = "gh pr merge 123";
    for _ in 0..3 {
        let (code, answer) = merge_raw("codex", &root, merge, Some("one"));
        assert_eq!((code, answer.as_str()), (0, ""), "observed");
    }
    merge_enforce(&root);
    for n in 1..=2 {
        let (_, answer) = merge_raw("codex", &root, merge, Some("one"));
        assert!(merge_refused(&answer), "enforced refusal {n}: {answer}");
    }
    let (_, answer) = merge_raw("codex", &root, merge, Some("one"));
    assert_eq!(answer, "", "the third enforced refusal is released");
}

/// B2, mutant M12: OMP cannot ask, so its refusals are counted like Codex's: with a session id
/// the third is released, and another session starts again.
#[test]
fn merge_confirm_counts_omp_refusals_per_session() {
    let (_tmp, root) = merge_repo(true);
    merge_enforce(&root);
    let merge = "gh pr merge 123";
    for n in 1..=2 {
        let (_, answer) = merge_raw("omp", &root, merge, Some("o1"));
        assert!(merge_refused(&answer), "omp refusal {n}: {answer}");
    }
    let (_, answer) = merge_raw("omp", &root, merge, Some("o1"));
    assert_eq!(answer, "", "released on the third");
    let (_, answer) = merge_raw("omp", &root, merge, Some("o2"));
    assert!(merge_refused(&answer), "another session: {answer}");
    let lines = merge_log(&root);
    let release = lines
        .iter()
        .filter(|l| l["kind"] == json!("rule_release"))
        .count();
    assert_eq!(release, 1, "{lines:?}");
}

/// N2 of the audit: a comment that only says `pr merge`, an option before the subcommand, and a
/// query string after the endpoint.
#[test]
fn merge_confirm_reads_the_subcommand_not_the_words_after_it() {
    let (_tmp, root) = merge_repo(true);
    merge_enforce(&root);
    for (n, command) in [
        "gh -R owner/repo pr merge 5",
        "gh --repo owner/repo pr merge 5 --squash",
        "gh api -X put 'repos/o/r/pulls/5/merge?x=1'",
        "gh --hostname github.com api -X PUT repos/o/r/pulls/5/merge",
    ]
    .iter()
    .enumerate()
    {
        let session = format!("s{n}");
        let (_, decision, reason) = merge_preflight_as("claude", &root, command, &session);
        assert_eq!(decision, "ask", "{command}: {reason}");
    }
    for command in [
        "gh pr comment 5 --body pr merge",
        "gh issue comment 5 --body 'pr merge soon'",
        "gh pr edit 5 --title 'pr merge'",
    ] {
        let (code, decision, reason) = merge_preflight_as("claude", &root, command, "q");
        assert_eq!(
            (code, decision.as_str(), reason.as_str()),
            (0, "", ""),
            "{command}"
        );
    }
}

/// REQ-5: every host's confirmation list names the direct merge.
#[test]
fn merge_confirm_capabilities_name_the_direct_merge() {
    for host in ["claude", "codex", "omp"] {
        let caps = moai_cli::hook_engine::capabilities(host);
        let commands = caps["tool-preflight"]["confirm"]["commands"].to_string();
        assert!(
            commands.contains("gh pr merge") && commands.contains("moai gate-merge"),
            "{host}: {commands}"
        );
        let modes = caps["tool-preflight"]["confirm"]["modes"].to_string();
        assert!(
            [
                "off",
                "observe",
                "enforce",
                "[hooks.modes]",
                "merge_confirm"
            ]
            .iter()
            .all(|w| modes.contains(w)),
            "{host}: {modes}"
        );
        // The limits of the release and of the log, named (audit of S4, B1 and N1).
        for limit in [
            "naming no session",
            "refused every time",
            "hook cwd in a subfolder",
            "size limit",
            ".moai-cli",
        ] {
            assert!(modes.contains(limit), "{host} {limit}: {modes}");
        }
    }
}

/// #2877 Q1: the Claude set carries the hook engine's preflight guard after the claim guard, and
/// the generated command, run as the host runs it, denies a dangerous Bash call and a write
/// into `.git`, and lets an ordinary call through. Codex keeps its three hooks and has no
/// preflight guard, only the session-check context.
#[test]
fn claude_runs_the_preflight_guard_and_codex_does_not() {
    let claude = hooks::hooks_object("claude").unwrap();
    let pre = claude["PreToolUse"].as_array().unwrap();
    assert_eq!(pre.len(), 2, "{pre:?}");
    assert_eq!(pre[0]["matcher"], json!("Bash"));
    assert!(pre[0]["hooks"][0]["command"]
        .as_str()
        .unwrap()
        .contains("claim-verify"));
    // The tools the engine guards (hook_engine::capabilities("claude")).
    assert_eq!(
        pre[1]["matcher"],
        json!("Bash|Write|Edit|MultiEdit|NotebookEdit")
    );
    let command = pre[1]["hooks"][0]["command"].as_str().unwrap();
    assert_eq!(
        command,
        "moai hook-dispatch --host claude --event tool-preflight"
    );
    let codex = hooks::hooks_object("codex").unwrap();
    assert_eq!(codex["PreToolUse"].as_array().unwrap().len(), 1);
    assert!(!codex
        .values()
        .any(|v| v.to_string().contains("tool-preflight")));
    assert_eq!(
        codex["SessionStart"][0]["hooks"][0]["command"],
        json!("moai hook-dispatch --host codex --event session-check")
    );

    // Run it the way Claude Code would: the command's words, our binary for `moai`.
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    let args: Vec<&str> = command.split_whitespace().skip(1).collect();
    let run = |input: Value| {
        let mut child = Command::new(common::MOAI_BIN)
            .args(&args)
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
        (
            out.status.code().unwrap(),
            String::from_utf8_lossy(&out.stdout).into_owned(),
        )
    };
    let tool = |name: &str, input: Value| {
        json!({"session_id": "s", "hook_event_name": "PreToolUse", "tool_name": name,
               "tool_use_id": "t1", "tool_input": input, "cwd": root})
    };
    for denied in [
        tool("Bash", json!({"command": "rm -rf /"})),
        tool(
            "Write",
            json!({"file_path": root.join(".git/config"), "content": "x"}),
        ),
    ] {
        let (code, stdout) = run(denied.clone());
        assert_eq!(code, 0, "{denied}");
        let v: Value = serde_json::from_str(&stdout).unwrap_or(Value::Null);
        assert_eq!(
            v["hookSpecificOutput"]["permissionDecision"],
            json!("deny"),
            "{denied} -> {stdout}"
        );
    }
    let (code, stdout) = run(tool("Bash", json!({"command": "ls"})));
    assert_eq!(
        (code, stdout.as_str()),
        (0, ""),
        "an ordinary call goes through"
    );
}

// ---- SPEC-20261119-enforce-spec-hooks S1: the GREEN lock against Bash, the deny audit line ----

const RED_COMMIT: &str = "0123456789abcdef0123456789abcdef01234567";

/// A temporary work tree in the GREEN phase of a TDD cycle: a `.git` folder (written by hand,
/// no `git` is started), the phase record that locks `tests/locked_test.rs`, an unlocked test
/// file, a source file and a `.moai-cli` folder for the engine's audit log.
struct GreenRepo {
    _tmp: tempfile::TempDir,
    root: std::path::PathBuf,
}

impl GreenRepo {
    fn new(phase: &str) -> Self {
        Self::build(phase, false)
    }

    /// The work tree three folders below the temporary folder, so that its parent is deep
    /// enough not to be a top-level directory on any platform (`/tmp` is one on Linux).
    fn nested(phase: &str) -> Self {
        Self::build(phase, true)
    }

    fn build(phase: &str, nested: bool) -> Self {
        let tmp = tempfile::TempDir::new().unwrap();
        let mut root = fs::canonicalize(tmp.path()).unwrap();
        if nested {
            root = root.join("outer/inner/repo");
        }
        fs::create_dir_all(root.join(".git")).unwrap();
        fs::write(root.join(".git/HEAD"), "ref: refs/heads/main\n").unwrap();
        for dir in ["home", "src", "tests", ".moai-cli"] {
            fs::create_dir_all(root.join(dir)).unwrap();
        }
        for file in [
            "src/a.rs",
            "tests/locked_test.rs",
            "tests/other_test.rs",
            "Cargo.toml",
        ] {
            fs::write(root.join(file), "x\n").unwrap();
        }
        let repo = GreenRepo { _tmp: tmp, root };
        repo.write_record(phase);
        repo
    }

    fn write_record(&self, phase: &str) {
        fs::create_dir_all(self.root.join(".git/moai")).unwrap();
        fs::write(
            self.root.join(".git/moai/tdd.json"),
            json!({"schema_version": 1, "spec": "SPEC-20261119-fixture", "phase": phase,
                   "red_commit": RED_COMMIT, "locked": ["tests/locked_test.rs"],
                   "checks": ["cargo test"], "at_unix": 1})
            .to_string(),
        )
        .unwrap();
    }

    /// Creates `files` (relative paths) and makes them the whole locked list of the record.
    fn lock_only(&self, files: &[&str]) {
        for file in files {
            let path = self.root.join(file);
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            fs::write(path, "x\n").unwrap();
        }
        fs::write(
            self.root.join(".git/moai/tdd.json"),
            json!({"schema_version": 1, "spec": "SPEC-20261119-fixture", "phase": "green",
                   "red_commit": RED_COMMIT, "locked": files,
                   "checks": ["cargo test"], "at_unix": 1})
            .to_string(),
        )
        .unwrap();
    }

    /// The Bash call as Claude Code sends it, run from the work tree; `cwd` is the folder the
    /// hook input names.
    fn bash(&self, command: &str, cwd: &Path) -> (i32, String) {
        let input = json!({"session_id": "s", "hook_event_name": "PreToolUse",
            "tool_name": "Bash", "tool_use_id": "t1",
            "tool_input": {"command": command}, "cwd": cwd});
        let mut child = Command::new(common::MOAI_BIN)
            .args([
                "hook-dispatch",
                "--host",
                "claude",
                "--event",
                "tool-preflight",
            ])
            .current_dir(&self.root)
            .env("HOME", self.root.join("home"))
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, input.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap(),
            String::from_utf8_lossy(&out.stdout).into_owned(),
        )
    }

    fn run(&self, command: &str) -> (i32, String) {
        self.bash(command, &self.root)
    }

    fn deny_reason(&self, command: &str) -> String {
        let (code, stdout) = self.run(command);
        let v: Value = serde_json::from_str(&stdout).unwrap_or(Value::Null);
        assert_eq!(code, 0, "{command}: {stdout}");
        assert_eq!(
            v["hookSpecificOutput"]["permissionDecision"],
            json!("deny"),
            "{command} -> {stdout}"
        );
        v["hookSpecificOutput"]["permissionDecisionReason"]
            .as_str()
            .unwrap()
            .to_string()
    }

    fn assert_allowed(&self, command: &str) {
        let (code, stdout) = self.run(command);
        assert_eq!((code, stdout.as_str()), (0, ""), "{command}");
    }

    fn audit_lines(&self) -> Vec<Value> {
        fs::read_to_string(self.root.join(".moai-cli/receipts/hook-engine.jsonl"))
            .unwrap_or_default()
            .lines()
            .map(|l| serde_json::from_str(l).unwrap())
            .collect()
    }
}

/// REQ-1, REQ-4: while the cycle is GREEN or REFACTOR, removing or moving a locked test with
/// Bash is refused with the lock's own reason, however the path is spelled.
#[test]
fn green_lock_bash_refuses_removing_or_moving_a_locked_test() {
    for phase in ["green", "refactor"] {
        let repo = GreenRepo::new(phase);
        let absolute = repo.root.join("tests/locked_test.rs");
        let absolute = absolute.to_str().unwrap();
        for command in [
            "rm tests/locked_test.rs".to_string(),
            "rm -f tests/locked_test.rs".to_string(),
            "rm -- tests/locked_test.rs".to_string(),
            "unlink tests/locked_test.rs".to_string(),
            "git rm tests/locked_test.rs".to_string(),
            "git rm -f -- tests/locked_test.rs".to_string(),
            "git mv tests/locked_test.rs old.rs".to_string(),
            "mv tests/locked_test.rs src/moved.rs".to_string(),
            // A locked path as the destination, and a new test file as the destination.
            "mv src/a.rs tests/locked_test.rs".to_string(),
            "mv src/a.rs tests/brand_new_test.rs".to_string(),
            "rm tests/other_test.rs tests/locked_test.rs".to_string(),
            "rm tests/../tests/locked_test.rs".to_string(),
            "rm ./tests/locked_test.rs".to_string(),
            format!("rm {absolute}"),
            "echo hi && rm tests/locked_test.rs".to_string(),
            "sh -c 'rm tests/locked_test.rs'".to_string(),
            "FOO=1 rm tests/locked_test.rs".to_string(),
        ] {
            let reason = repo.deny_reason(&command);
            assert!(
                reason.contains("GREEN or REFACTOR"),
                "{phase} {command}: {reason}"
            );
        }
    }
}

/// REQ-2, REQ-3: everything else goes through unchanged: an unlocked file, a read of a locked
/// one, a removal of nothing, and a cycle that is not in GREEN.
#[test]
fn green_lock_bash_allows_everything_else() {
    let repo = GreenRepo::new("green");
    for command in [
        "rm src/a.rs",
        "rm tests/other_test.rs",
        "mv src/a.rs src/b.rs",
        "git rm src/a.rs",
        "git mv src/a.rs src/c.rs",
        "unlink src/a.rs",
        // Nothing to remove: the lock's rule against new test files is not meant for it.
        "rm -f tests/gone_test.rs",
        "unlink tests/gone_test.rs",
        "git rm --cached tests/gone_test.rs",
        "cat tests/locked_test.rs",
        "ls tests/locked_test.rs",
        "cp tests/locked_test.rs /tmp/copy",
        "git status",
        "rm",
        "echo rm tests/locked_test.rs",
        "grep -n rm tests/locked_test.rs",
    ] {
        repo.assert_allowed(command);
    }
}

/// REQ-3, REQ-4: no record, an ended cycle, a repository without a work tree, and a relative
/// path read against the hook input's `cwd`.
#[test]
fn green_lock_bash_follows_the_record_and_the_input_cwd() {
    let repo = GreenRepo::new("green");
    let sub = repo.root.join("src");
    // The path is relative to the folder the hook input names, not the process folder.
    let (code, stdout) = repo.bash("rm ../tests/locked_test.rs", &sub);
    assert_eq!(code, 0);
    assert!(
        stdout.contains("\"permissionDecision\":\"deny\""),
        "{stdout}"
    );
    let (code, stdout) = repo.bash("rm tests/locked_test.rs", &sub);
    assert_eq!(
        (code, stdout.as_str()),
        (0, ""),
        "src/tests/ does not exist"
    );

    // The cycle is over (no record), or the record is not one this build reads.
    fs::remove_file(repo.root.join(".git/moai/tdd.json")).unwrap();
    repo.assert_allowed("rm tests/locked_test.rs");
    repo.write_record("red");
    repo.assert_allowed("rm tests/locked_test.rs");
    repo.write_record("green");
    repo.deny_reason("rm tests/locked_test.rs");

    // Not a work tree at all: nothing to read, nothing refused.
    fs::remove_dir_all(repo.root.join(".git")).unwrap();
    repo.assert_allowed("rm tests/locked_test.rs");
}

/// REQ-8: removing or moving the folder that holds a locked test, and a glob or a brace group
/// that names one, are refused like the file itself; a folder or a pattern that holds or
/// names no locked test is not.
#[test]
fn green_lock_bash_covers_folders_globs_and_braces() {
    let repo = GreenRepo::new("green");
    let absolute = repo.root.join("tests");
    let absolute = absolute.to_str().unwrap();
    for command in [
        "rm -rf tests".to_string(),
        "rm -r tests/".to_string(),
        "rm -r ./tests".to_string(),
        "rm -rf src tests".to_string(),
        "git rm -r tests".to_string(),
        "git rm -rf --cached tests".to_string(),
        "mv tests tests_old".to_string(),
        "git mv tests tests_old".to_string(),
        format!("rm -rf {absolute}"),
        "rm tests/locked_*.rs".to_string(),
        "rm tests/locked_test.r?".to_string(),
        "rm tests/locked_tes[t].rs".to_string(),
        "rm tests/[a-m]ocked_test.rs".to_string(),
        "rm tests/[!x]ocked_test.rs".to_string(),
        "rm tests/[^x]ocked_test.rs".to_string(),
        "rm tests/*".to_string(),
        "rm te*/locked_test.rs".to_string(),
        // A glob that stops at the folder names the folder.
        "rm -rf t*".to_string(),
        "rm -r te?ts".to_string(),
        "rm -rf [t]ests".to_string(),
        "mv te* /tmp/".to_string(),
        "git rm -r tes*".to_string(),
        "mv tests/locked_* src/".to_string(),
        "rm tests/{locked_test.rs,other_test.rs}".to_string(),
        "rm tests/{other,locked}_test.rs".to_string(),
        "rm tests/{x,{locked_test.rs,y}}".to_string(),
        "FOO=1 rm -rf tests".to_string(),
        "echo hi && rm -rf tests".to_string(),
    ] {
        let reason = repo.deny_reason(&command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    for command in [
        "rm -rf src",
        "rm -r target",
        "mv src src2",
        "git mv src src2",
        "rm tests/other_*.rs",
        "rm tests/zz*.rs",
        "rm tests/[a-m]ther_test.rs",
        "rm tests/[!l]ocked_test.rs",
        "rm tests/{other,zz}_test.rs",
        "rm src/*.rs",
        "rm -r s*",
        "rm zz*",
        "rm -r [a-s]rc",
        "ls tests",
        "cat tests/*",
        "cp -r tests /tmp/tests_copy",
        "echo ${HOME}/x",
    ] {
        repo.assert_allowed(command);
    }
}

/// N1 (audit of #3216, mutant H7): after `--` a word that starts with a dash is a path; before
/// it, it is an option.
#[test]
fn green_lock_bash_reads_dash_names_after_double_dash() {
    let repo = GreenRepo::new("green");
    fs::write(repo.root.join("tests/-odd_test.rs"), "x\n").unwrap();
    fs::write(
        repo.root.join(".git/moai/tdd.json"),
        json!({"schema_version": 1, "spec": "SPEC-20261119-fixture", "phase": "green",
               "red_commit": RED_COMMIT, "locked": ["tests/-odd_test.rs"],
               "checks": ["cargo test"], "at_unix": 1})
        .to_string(),
    )
    .unwrap();
    let tests = repo.root.join("tests");
    let (code, stdout) = repo.bash("rm -- -odd_test.rs", &tests);
    assert!(
        code == 0 && stdout.contains("\"permissionDecision\":\"deny\""),
        "{code} {stdout}"
    );
    let (code, stdout) = repo.bash("rm -f -- -odd_test.rs", &tests);
    assert!(
        stdout.contains("\"permissionDecision\":\"deny\""),
        "{code} {stdout}"
    );
    // Without `--` the same word is an option, and nothing is named.
    let (code, stdout) = repo.bash("rm -odd_test.rs", &tests);
    assert_eq!((code, stdout.as_str()), (0, ""));
}

/// #3216 re-audit B1: a locked test whose real name holds `[..]`, `*`, `?` or a brace group is
/// still protected when the word is written as it is: the word is read literally as well as
/// matched as a pattern.
#[test]
fn green_lock_bash_reads_locked_names_with_pattern_characters() {
    let repo = GreenRepo::new("green");
    repo.lock_only(&[
        "app/[id]/page.test.tsx",
        "tests/[x]_test.rs",
        "tests/a{b,c}_test.rs",
        "tests/star*_test.rs",
        "tests/what?_test.rs",
    ]);
    for command in [
        "rm 'app/[id]/page.test.tsx'",
        "rm app/[id]/page.test.tsx",
        "rm app/\\[id\\]/page.test.tsx",
        "rm \"app/[id]/page.test.tsx\"",
        "git rm 'app/[id]/page.test.tsx'",
        "mv 'app/[id]/page.test.tsx' /tmp/x",
        "unlink 'app/[id]/page.test.tsx'",
        "rm 'tests/[x]_test.rs'",
        "rm 'tests/a{b,c}_test.rs'",
        "rm 'tests/star*_test.rs'",
        "rm 'tests/what?_test.rs'",
        "mv src/a.rs 'tests/new_[1]_test.rs'",
        "mv src/a.rs 'tests/new_{1,2}_test.rs'",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    // A name that matches no locked test and is not itself one stays open.
    for command in [
        "rm 'app/[id]/other.tsx'",
        "rm 'tests/[y]_test.rs'",
        "git rm 'app/[ab]/page.test.tsx'",
    ] {
        repo.assert_allowed(command);
    }
}

/// #3216 re-audit N1: the shapes the first repair let through without naming them. `!`, a `{ …; }`
/// group, `if`/`while`/`for` bodies, backticks, `busybox`, `git -C dir` and a folder above the
/// work tree are read; `case`, variables, `~`, extended globs and the brace limits are named in
/// `capabilities` and stay open (`green_lock_bash_named_limits_are_not_guarded`).
#[test]
fn green_lock_bash_reads_wrapped_grouped_and_rebased_commands() {
    let repo = GreenRepo::new("green");
    for command in [
        "! rm tests/locked_test.rs".to_string(),
        "{ rm tests/locked_test.rs; }".to_string(),
        "{ true; rm tests/locked_test.rs; }".to_string(),
        "( rm tests/locked_test.rs )".to_string(),
        "if true; then rm tests/locked_test.rs; fi".to_string(),
        "if rm tests/locked_test.rs; then echo x; fi".to_string(),
        "while true; do mv tests/locked_test.rs /tmp/x; done".to_string(),
        "for f in a b; do git rm tests/locked_test.rs; done".to_string(),
        "f() { rm tests/locked_test.rs; }; f".to_string(),
        "echo `rm tests/locked_test.rs`".to_string(),
        "`rm tests/locked_test.rs`".to_string(),
        "busybox rm tests/locked_test.rs".to_string(),
        "busybox mv tests/locked_test.rs /tmp/x".to_string(),
        "git -C tests rm locked_test.rs".to_string(),
        "git -C tests mv locked_test.rs /tmp/x".to_string(),
        "git -C . rm tests/locked_test.rs".to_string(),
        "git -C tests -C .. rm tests/locked_test.rs".to_string(),
        "git -C tests rm -r .".to_string(),
    ] {
        let reason = repo.deny_reason(&command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    // A folder above the work tree holds the whole tree. The tree sits three folders down, so
    // the parent is no top-level directory and the static `rm` guard does not answer first.
    let deep = GreenRepo::nested("green");
    let parent = deep.root.parent().unwrap().to_str().unwrap().to_string();
    for command in [
        format!("rm -rf {parent}"),
        format!("rm -rf {parent}/"),
        format!("mv {parent} /tmp/moved"),
    ] {
        let reason = deep.deny_reason(&command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    // The static rm guard already refuses these, with its own sentence.
    for command in [
        "rm -rf ..",
        "rm -rf ../",
        "rm -rf /",
        "rm -rf .",
        "rm -rf ./",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("rm aimed at"), "{command}: {reason}");
    }
    for command in [
        "echo '{ rm tests/locked_test.rs; }'",
        "echo \"then rm tests/locked_test.rs\"",
        "echo `date`",
        "busybox ls tests",
        "busybox cat tests/locked_test.rs",
        "git -C tests status",
        "git -C src log --oneline",
        "git -C tests rm other_test.rs",
        "git -C src rm a.rs",
        "! ls tests",
        "if true; then ls tests; fi",
        "{ ls tests; }",
        "rm -rf src/../src",
    ] {
        repo.assert_allowed(command);
    }
}

/// #3216 re-audit round 3, B1: a structure word is a word wherever a space, `;`, `&`, `|`, `(`,
/// `)`, a quote or a backtick ends it, not only with spaces on both sides.
#[test]
fn green_lock_bash_reads_structure_words_at_every_boundary() {
    let repo = GreenRepo::new("green");
    for command in [
        "if true;then rm tests/locked_test.rs;fi",
        "for f in a b;do rm tests/locked_test.rs;done",
        "while true;do mv tests/locked_test.rs /tmp/x;done",
        "until false;do git rm tests/locked_test.rs;done",
        "if false;then :;else rm tests/locked_test.rs;fi",
        "if false;then :;elif true;then rm tests/locked_test.rs;fi",
        "if false;then :;elif rm tests/locked_test.rs;then :;fi",
        "(if true;then rm tests/locked_test.rs;fi)",
        "(! rm tests/locked_test.rs)",
        "true&&{ rm tests/locked_test.rs;}",
        "true||{ rm tests/locked_test.rs; }",
        "sh -c '! rm tests/locked_test.rs'",
        "sh -c '{ rm tests/locked_test.rs; }'",
        "bash -c \"if true;then rm tests/locked_test.rs;fi\"",
        "eval '{ rm tests/locked_test.rs; }'",
        "eval 'if true;then mv tests/locked_test.rs /tmp/x;fi'",
        "`if true;then rm tests/locked_test.rs;fi`",
        // Named false denials: a structure word is read as one wherever it stands as a whole
        // word, so a command line that only mentions one still refuses (#3216 N2).
        "echo do rm tests/locked_test.rs",
        "echo then mv tests/locked_test.rs x",
        "grep -c if rm tests/locked_test.rs",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    for command in [
        "echo 'if true;then rm tests/locked_test.rs;fi'",
        "echo \"for f in a;do rm tests/locked_test.rs;done\"",
        "if true;then ls tests;fi",
        "for f in a b;do ls tests;done",
        "sh -c '! ls tests'",
        "sh -c '{ ls tests; }'",
        "true&&{ ls tests;}",
        "echo ifthen;do",
        // A backslash-escaped boundary is part of its word: this only echoes.
        "echo a\\;do rm tests/locked_test.rs",
    ] {
        repo.assert_allowed(command);
    }
}

/// #3216 re-audit round 3, B2: ANSI-C (`$'…'`) and locale (`$"…"`) quoting is read as plain
/// quoting; the escape sequences inside `$'…'` are not decoded and stay named in `capabilities`.
#[test]
fn green_lock_bash_reads_ansi_c_and_locale_quotes() {
    let repo = GreenRepo::new("green");
    for command in [
        "rm $'tests/locked_test.rs'",
        "mv $'tests/locked_test.rs' /tmp/x",
        "git rm $'tests/locked_test.rs'",
        "rm $\"tests/locked_test.rs\"",
        "rm tests/$'locked_test.rs'",
        "rm -- $'tests/locked_test.rs'",
        "busybox rm $'tests/locked_test.rs'",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    for command in [
        "rm $'tests/other_test.rs'",
        "echo $'tests/locked_test.rs'",
        "echo '$tests/locked_test.rs'",
        // The escape sequences are not decoded (named in `capabilities`).
        "rm $'\\x74ests/locked_test.rs'",
        "rm $'tests/locked_test\\x2ers'",
    ] {
        repo.assert_allowed(command);
    }
}

/// #3216 re-audit N3: edges the first repair left untested.
#[test]
fn green_lock_bash_edges_of_folders_globs_and_moves() {
    let repo = GreenRepo::new("green");
    // The work-tree root as a folder word, with and without a slash.
    for command in ["git rm -r .", "git rm -rf ./", "git mv . ../moved"] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    // A class range includes both ends; `*` may match nothing; `?` needs one character.
    for command in [
        "rm tests/[a-l]ocked_test.rs",
        "rm tests/[l-z]ocked_test.rs",
        "rm tests/locked_test*.rs",
        "rm tests/locked_test.r?",
        "rm tests/*locked_test.rs",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    for command in [
        "rm tests/[a-k]ocked_test.rs",
        "rm tests/[m-z]ocked_test.rs",
        "rm tests/locked_test?.rs",
    ] {
        repo.assert_allowed(command);
    }
    // A moved file lands on a new test path: the lock refuses a new test file too.
    for command in [
        "git mv src/a.rs tests/brand_new_test.rs",
        "mv src/a.rs tests/brand_new_test.rs",
        "git mv tests/other_test.rs tests/brand_new_test.rs",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
    // A symlink as the glob's folder part: the real folder is the one that is read.
    std::os::unix::fs::symlink(repo.root.join("tests"), repo.root.join("link")).unwrap();
    for command in [
        "rm link/*",
        "rm link/locked_*.rs",
        "rm -rf link",
        "rm -rf link/",
    ] {
        let reason = repo.deny_reason(command);
        assert!(reason.contains("GREEN or REFACTOR"), "{command}: {reason}");
    }
}

/// The brace limits named in `capabilities`: 256 alternatives, 4 levels of nesting.
#[test]
fn green_lock_bash_brace_limits_are_the_named_ones() {
    let repo = GreenRepo::new("green");
    let list = |n: usize| {
        (0..n)
            .map(|i| format!("z{i}"))
            .chain(["locked_test".to_string()])
            .collect::<Vec<_>>()
            .join(",")
    };
    let reason = repo.deny_reason(&format!("rm tests/{{{}}}.rs", list(100)));
    assert!(reason.contains("GREEN or REFACTOR"), "{reason}");
    repo.assert_allowed(&format!("rm tests/{{{}}}.rs", list(300)));
    // The edge itself (#3216 N3): `list(n)` holds n + 1 alternatives, 256 read and 257 not.
    let reason = repo.deny_reason(&format!("rm tests/{{{}}}.rs", list(255)));
    assert!(reason.contains("GREEN or REFACTOR"), "{reason}");
    repo.assert_allowed(&format!("rm tests/{{{}}}.rs", list(256)));
    let reason = repo.deny_reason("rm tests/{{{{locked_test,b},c},d},e}.rs");
    assert!(reason.contains("GREEN or REFACTOR"), "{reason}");
    repo.assert_allowed("rm tests/{{{{{locked_test,b},c},d},e},f}.rs");
}

/// The forms `capabilities` names as not guarded stay allowed, so that its words and the
/// guard cannot drift apart: when one of these is closed, the test and the named limit change
/// together.
#[test]
fn green_lock_bash_named_limits_are_not_guarded() {
    let repo = GreenRepo::new("green");
    for command in [
        "F=tests/locked_test.rs; rm $F",
        "rm \"$LOCKED\"",
        "echo tests/locked_test.rs | xargs rm",
        "xargs rm < list.txt",
        "find tests -name locked_test.rs -exec rm {} +",
        "find tests -name locked_test.rs -delete",
        "for f in tests/locked_test.rs; do rm $f; done",
        "rm tests/@(locked_test).rs",
        "rm tests/+(locked_test).rs",
        "cd tests && rm locked_test.rs",
        "echo x > tests/locked_test.rs",
        "sed -i s/a/b/ tests/locked_test.rs",
        "tee tests/locked_test.rs < /dev/null",
        "bash cleanup.sh",
        "python3 -c \"import os; os.remove('tests/locked_test.rs')\"",
        "rm ~/../tests/locked_test.rs",
        // #3216 re-audit round 3 N1, N6: named, not guarded.
        "coproc rm tests/locked_test.rs",
        "cp /dev/null tests/locked_test.rs",
        "truncate -s 0 tests/locked_test.rs",
        "dd if=/dev/null of=tests/locked_test.rs",
        "install /dev/null tests/locked_test.rs",
        "ln -f other.rs tests/locked_test.rs",
        "rsync -a --delete empty/ tests/",
        "git checkout -- tests/locked_test.rs",
        "git restore tests/locked_test.rs",
        "git stash",
        "git apply fix.patch",
    ] {
        repo.assert_allowed(command);
    }
}

/// REQ-5: a denied call leaves one line with the reason text, and no command, path or content.
#[test]
fn deny_audit_line_carries_the_reason_and_no_command_or_path() {
    let repo = GreenRepo::new("green");
    let locked = repo.deny_reason("rm tests/locked_test.rs");
    let danger = repo.deny_reason("rm -rf /");
    assert_ne!(locked, danger);
    repo.assert_allowed("ls");

    let lines = repo.audit_lines();
    let denies: Vec<&Value> = lines
        .iter()
        .filter(|l| l["kind"] == json!("deny"))
        .collect();
    assert_eq!(denies.len(), 2, "{lines:?}");
    assert_eq!(denies[0]["reason"], json!(locked));
    assert_eq!(denies[1]["reason"], json!(danger));
    let text = serde_json::to_string(&lines).unwrap();
    for leak in ["locked_test", "rm -rf", repo.root.to_str().unwrap()] {
        assert!(!text.contains(leak), "{leak} leaked into {text}");
    }
}

/// REQ-7: a refusal whose sentence is built at run time leaves `reason: null` in the line.
#[test]
fn deny_audit_line_is_null_when_the_sentence_is_built_at_run_time() {
    let repo = GreenRepo::new("green");
    let reason = repo.deny_reason("git -C src commit -m a && git -C tests commit -m b");
    assert!(reason.contains("more than one folder"), "{reason}");
    let lines = repo.audit_lines();
    let denies: Vec<&Value> = lines
        .iter()
        .filter(|l| l["kind"] == json!("deny"))
        .collect();
    assert_eq!(denies.len(), 1, "{lines:?}");
    assert_eq!(denies[0]["reason"], Value::Null, "{lines:?}");
}

/// REQ-6: every host's answer says what the Bash lock covers and what it does not.
#[test]
fn capabilities_name_the_bash_lock_limits() {
    for host in ["claude", "codex", "omp"] {
        let caps = moai_cli::hook_engine::capabilities(host);
        let pre = &caps["tool-preflight"];
        let guards = pre["guards"].to_string();
        for covered in [
            "rm, unlink, git rm, mv or git mv of a test path the GREEN record locks",
            "of a folder holding one or above the work tree",
            "glob (*, ?, [..])",
            "brace ({a,b}) pattern",
            "read literally",
            "behind the structure words ! { } if then elif else fi while until do done",
            "a space, ; & | ( ), a quote or a backtick",
            "backticks, busybox and git -C",
            "ANSI-C ($'...') and locale ($ before a double quote) quoting read as plain quotes",
            "a structure word as an argument refuses too (echo do rm tests/x)",
        ] {
            assert!(guards.contains(covered), "{host} {covered}: {guards}");
        }
        let open = pre["not_guarded"].to_string();
        // Each form that the guard lets through (`green_lock_bash_named_limits_are_not_guarded`,
        // `green_lock_bash_brace_limits_are_the_named_ones`) is named, not only the redirections.
        for limit in [
            "redirections",
            "tee",
            "sed -i",
            "scripts",
            "variables (rm $F)",
            "xargs",
            "find -exec",
            "find -delete",
            "extended globs (@(..), +(..))",
            "~ expansion",
            "more than 256 brace alternatives",
            "braces nested deeper than 4",
            "any command run after a cd",
            "escape sequences inside $'...'",
            "coproc",
            "cp, truncate, dd of=, install, ln -f, rsync --delete",
            "git checkout --, git restore, git stash and git apply",
        ] {
            assert!(
                open.contains("Bash edits of a locked test") && open.contains(limit),
                "{host} {limit}: {open}"
            );
        }
    }
}
