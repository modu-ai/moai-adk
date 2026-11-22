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
    // SPEC-20261122-next-flow-hooks: the Claude set is the NeXT-Flow set — the four sources
    // matcher on SessionStart, the dangerous duty and the ordinary duty on the shell tools,
    // the ordinary duty again on the write tools, every entry in exec form with its timeout,
    // the dangerous entry alone with onFailure block. The other hosts keep the Issue-era base.
    let claude = hooks::hooks_object("claude").unwrap();
    let events: Vec<&String> = claude.keys().collect();
    assert_eq!(events, ["SessionStart", "PreToolUse"], "{claude:?}");
    assert!(claude.get("Stop").is_none(), "claude: Stop has no reader");
    assert!(claude.get("PostToolUse").is_none(), "claude: no receipt");
    assert_eq!(
        claude["SessionStart"][0]["matcher"],
        json!(hooks::CLAUDE_SESSION_MATCHER)
    );
    assert_eq!(
        claude["PreToolUse"][0]["matcher"],
        json!(hooks::CLAUDE_SHELL_MATCHER)
    );
    assert_eq!(
        claude["PreToolUse"][0]["hooks"][0]["onFailure"],
        json!("block"),
        "the dangerous duty alone blocks on its own failure"
    );
    assert!(
        claude["PreToolUse"][1]["hooks"][0]
            .get("onFailure")
            .is_none(),
        "the ordinary duty stays fail-open"
    );
    assert_eq!(
        claude["PreToolUse"][2]["matcher"],
        json!(hooks::CLAUDE_WRITE_MATCHER)
    );
    for (_, groups) in &claude {
        for group in groups.as_array().unwrap() {
            for handler in group["hooks"].as_array().unwrap() {
                assert_eq!(handler["type"], json!("command"), "{handler}");
                assert!(handler["timeout"].is_u64(), "{handler}");
                assert!(handler["args"].is_array(), "exec form: {handler}");
                assert!(
                    handler["command"].as_str().unwrap().starts_with('/'),
                    "absolute binary path: {handler}"
                );
            }
        }
    }
    for host in ["codex", "omp"] {
        let obj = hooks::hooks_object(host).unwrap();
        assert_eq!(
            obj.keys().collect::<Vec<_>>(),
            ["SessionStart", "PreToolUse", "PostToolUse"],
            "{host}"
        );
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
    // SPEC-20261122-next-flow-hooks REQ-1: the Claude set has no claim-verify entry; its
    // entries are exec form, so the "bare moai + args" premise is gone with it. What stays:
    // running the args of a PreToolUse entry as the host would denies a dangerous call and
    // allows a plain one, on every host whose set carries the preflight duty.
    let dir = project();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join(".moai")).unwrap();
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

    {
        // Only Claude carries the preflight duty (#2877 Q1; codex keeps its base set).
        let obj = hooks::hooks_object("claude").unwrap();
        let entry = obj["PreToolUse"]
            .as_array()
            .unwrap()
            .iter()
            .flat_map(|g| g["hooks"].as_array().unwrap().iter().cloned())
            .find(|h| {
                h["args"]
                    .as_array()
                    .map(|a| a.iter().any(|w| w == &json!("tool-preflight")))
                    .unwrap_or_else(|| {
                        h["command"]
                            .as_str()
                            .is_some_and(|c| c.contains("tool-preflight"))
                    })
            })
            .unwrap();
        let args: Vec<String> = entry["args"]
            .as_array()
            .map(|a| {
                a.iter()
                    .filter_map(|w| w.as_str().map(str::to_string))
                    .collect()
            })
            .unwrap_or_else(|| {
                entry["command"]
                    .as_str()
                    .unwrap()
                    .split_whitespace()
                    .skip(1)
                    .map(str::to_string)
                    .collect()
            });
        let run = |tool_command: &str| {
            let input = json!({
                "hook_event_name": "PreToolUse", "tool_name": "Bash",
                "tool_input": {"command": tool_command}, "cwd": root,
                "session_id": "s",
            });
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
            common::feed_stdin(&mut child, input.to_string().as_bytes());
            let out = child.wait_with_output().unwrap();
            (
                out.status.code().unwrap(),
                String::from_utf8_lossy(&out.stdout).into_owned(),
            )
        };
        let (code, stdout) = run("rm -rf /");
        assert!(
            code == 2 || stdout.contains("\"deny\""),
            "{} {stdout}",
            code
        );
        let (code, stdout) = run("git status");
        assert_eq!((code, stdout.as_str()), (0, ""));
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
    // Named false confirmation (the closed rule of `open_shell_structure`): a structure word
    // standing as an argument still counts as one, so this only echoes and still asks.
    "echo do gh pr merge 5",
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
    // The log answers for either state-folder name (SPEC-20261122-next-flow-hooks REQ-12), so
    // both must be gone.
    fs::remove_dir_all(root.join(".moai-cli")).unwrap();
    let _ = fs::remove_dir_all(root.join(".moai"));
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
    // SPEC-20261122-next-flow-hooks: the Claude PreToolUse groups are the dangerous duty and
    // the ordinary duty on the shell tools plus the ordinary duty on the write tools — no
    // claim-verify anywhere — and the generated command, run as the host runs it, denies a
    // dangerous Bash call and a write into `.git`, and lets an ordinary call through. Codex
    // keeps its three base hooks.
    let claude = hooks::hooks_object("claude").unwrap();
    let pre = claude["PreToolUse"].as_array().unwrap();
    assert_eq!(pre.len(), 3, "{pre:?}");
    let text = serde_json::to_string(&claude).unwrap();
    assert!(!text.contains("claim-verify"), "{text}");
    assert!(!text.contains("hook-receipt"), "{text}");
    for matcher in [hooks::CLAUDE_SHELL_MATCHER, hooks::CLAUDE_WRITE_MATCHER] {
        assert!(
            pre.iter().any(|g| g["matcher"] == json!(matcher)),
            "{pre:?}"
        );
    }
    let codex = hooks::hooks_object("codex").unwrap();
    assert_eq!(
        codex.keys().collect::<Vec<_>>(),
        ["SessionStart", "PreToolUse", "PostToolUse"]
    );
    let codex_text = serde_json::to_string(&codex).unwrap();
    assert!(codex_text.contains("claim-verify"), "codex keeps its base");
    // Run as the host runs it: the same argv the entry spells. Inside a test binary
    // `moai_absolute` names this test executable, so the run uses the built moai binary with
    // the entry's own args vector (the shape assertions above pin what is written).
    let entry = &pre[1]["hooks"][0];
    let program = crate::common::MOAI_BIN;
    let args: Vec<String> = entry["args"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(|a| a.as_str().map(str::to_string))
        .collect();
    let root = tempfile::TempDir::new().unwrap();
    std::fs::create_dir_all(root.path().join(".moai")).unwrap();
    let run = |tool: &str, tool_input: Value| {
        let input = json!({"hook_event_name": "PreToolUse", "tool_name": tool,
                           "tool_input": tool_input, "cwd": root.path(),
                           "session_id": "s"});
        let mut child = std::process::Command::new(program)
            .args(&args)
            .current_dir(root.path())
            .env("HOME", root.path().join("home"))
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(std::process::Stdio::piped())
            .stdout(std::process::Stdio::piped())
            .stderr(std::process::Stdio::null())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, input.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap(),
            String::from_utf8_lossy(&out.stdout).into_owned(),
        )
    };
    let (code, stdout) = run("Bash", json!({"command": "rm -rf /"}));
    assert_eq!(code, 2, "{stdout}");
    assert!(stdout.contains("\"deny\""), "{stdout}");
    let (code, stdout) = run(
        "Write",
        json!({"file_path": ".git/hooks/pre-commit", "content": "x"}),
    );
    assert_eq!(code, 2, "{stdout}");
    let (code, stdout) = run("Bash", json!({"command": "git status"}));
    assert_eq!(code, 0, "{stdout}");
    assert!(stdout.is_empty(), "{stdout}");
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
        // A refusal is exit code 2 with the documented deny object beside it (REQ-8).
        assert_eq!(code, 2, "{command}: {stdout}");
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
    assert_eq!(code, 2, "a refusal is exit 2 (REQ-8)");
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
        code == 2 && stdout.contains("\"permissionDecision\":\"deny\""),
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
    // Unix only: `std::os::unix` does not exist on Windows (SPEC-20261119-windows-unix-gate).
    #[cfg(unix)]
    {
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

// ---- SPEC-20261119-enforce-ac-verify-recorded -------------------------------------------------

const AC_SPEC: &str = "SPEC-20261119-ac-demo";
const AC_BRANCH: &str = "spec/20261119-ac-demo";
const AC_HEAD: &str = "1111111111111111111111111111111111111111";
const AC_HEAD_2: &str = "2222222222222222222222222222222222222222";
const AC_HEAD_PART: &str = "---\nid: SPEC-20261119-ac-demo\ntitle: \"demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-19\n---\n\n# demo\n\n## Acceptance\n\n";
/// Six criteria: two with one command, one with two, and three the rule must leave out.
const AC_BODY: &str = "- AC-1 (REQ-1): one · verify: `cargo test --test one`\n\
- AC-2 (REQ-2): two · verify: `cargo test --test two -- --exact x`\n\
- AC-3 (REQ-3): three · manual: `read the page`\n\
- AC-4 (REQ-4): four · verify: `<command that checks it>`\n\
- AC-5 (REQ-5): five · unknown: `no way yet`\n\
- AC-6 (REQ-6): six · verify: `cargo test --test six_a`; `cargo test --test six_b`\n";
/// Only criteria the rule leaves out.
const AC_BODY_LEFT_OUT: &str = "- AC-3 (REQ-3): three · manual: `read the page`\n\
- AC-4 (REQ-4): four · verify: `<command that checks it>`\n\
- AC-5 (REQ-5): five · unknown: `no way yet`\n";

/// A temporary work tree on `branch` (a hand-made `.git`: `HEAD` and the branch's ref file with
/// `head`, when there is one) holding the demo SPEC `body` when `spec` is given.
fn ac_repo(
    branch: &str,
    head: Option<&str>,
    spec: Option<&str>,
) -> (tempfile::TempDir, std::path::PathBuf) {
    let (tmp, root) = merge_repo(false);
    fs::write(
        root.join(".git/HEAD"),
        format!("ref: refs/heads/{branch}\n"),
    )
    .unwrap();
    if let Some(sha) = head {
        let r = root.join(".git/refs/heads").join(branch);
        fs::create_dir_all(r.parent().unwrap()).unwrap();
        fs::write(r, format!("{sha}\n")).unwrap();
    }
    if let Some(body) = spec {
        let d = root.join(".moai/specs").join(AC_SPEC);
        fs::create_dir_all(&d).unwrap();
        fs::write(d.join("spec.md"), format!("{AC_HEAD_PART}{body}")).unwrap();
    }
    (tmp, root)
}

fn ac_dispatch(event: &str, host: &str, root: &Path, input: &Value) -> (i32, String) {
    let mut child = Command::new(common::MOAI_BIN)
        .args(["hook-dispatch", "--host", host, "--event", event])
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

/// One finished (or, with `failed`, failed) Bash call as the host reports it to tool-observe.
fn ac_bash(root: &Path, session: &str, command: &str, failed: bool) -> (i32, String) {
    let event = if failed {
        "PostToolUseFailure"
    } else {
        "PostToolUse"
    };
    let input = json!({"session_id": session, "hook_event_name": event,
        "tool_name": "Bash", "tool_use_id": "t", "tool_input": {"command": command},
        "tool_response": {}, "cwd": root});
    ac_dispatch("tool-observe", "claude", root, &input)
}

/// One file write of the session.
fn ac_write(root: &Path, session: &str) -> (i32, String) {
    let input = json!({"session_id": session, "hook_event_name": "PostToolUse",
        "tool_name": "Write", "tool_use_id": "t",
        "tool_input": {"file_path": root.join("src/x.rs"), "content": "x"},
        "tool_response": {}, "cwd": root});
    ac_dispatch("tool-observe", "claude", root, &input)
}

fn ac_stop(root: &Path, session: &str, active: bool) -> (i32, String) {
    let input = json!({"session_id": session, "hook_event_name": "Stop",
        "stop_hook_active": active, "cwd": root});
    ac_dispatch("turn-validation", "claude", root, &input)
}

fn ac_runs(root: &Path) -> Vec<Value> {
    merge_log(root)
        .into_iter()
        .filter(|l| l["kind"] == json!("ac_run"))
        .collect()
}

/// The rule's mode in the HOME hook config of the temporary project.
fn ac_mode(root: &Path, mode: &str) {
    merge_config(
        root,
        &format!("[hooks.modes]\nac_verify_recorded = \"{mode}\"\n"),
    );
}

/// Whether Claude's Stop answer keeps the session working; the reason when it does.
fn ac_reason(answer: &(i32, String)) -> Option<String> {
    assert_eq!(answer.0, 0, "{answer:?}");
    let v: Value = serde_json::from_str(&answer.1).ok()?;
    (v["decision"] == json!("block")).then(|| v["reason"].as_str().unwrap_or("").to_string())
}

/// Lets a later log line carry a later millisecond than the one before it.
fn ac_pause() {
    std::thread::sleep(std::time::Duration::from_millis(5));
}

/// REQ-1: a successful Bash call equal to a verify command (white space collapsed) leaves one
/// `ac_run` line with the SPEC, the AC, the command's position, the full commit, the session
/// and the time; the observe line is still written and no command text is kept.
#[test]
fn ac_verify_run_is_recorded_for_the_verify_command() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    for (command, ac, position) in [
        ("cargo test --test one", "AC-1", 0),
        ("  cargo   test\t--test one  ", "AC-1", 0),
        ("cargo test --test two -- --exact x", "AC-2", 0),
        ("cargo test --test six_a", "AC-6", 0),
        ("cargo test --test six_b", "AC-6", 1),
    ] {
        let before = ac_runs(&root).len();
        let (code, out) = ac_bash(&root, "s1", command, false);
        assert_eq!((code, out.as_str()), (0, ""), "{command}");
        let runs = ac_runs(&root);
        assert_eq!(runs.len(), before + 1, "{command}: {runs:?}");
        let run = runs.last().unwrap();
        assert_eq!(run["spec"], json!(AC_SPEC), "{command}");
        assert_eq!(run["ac"], json!(ac), "{command}");
        assert_eq!(run["cmd"], json!(position), "{command}");
        assert_eq!(run["head"], json!(AC_HEAD), "{command}");
        assert_eq!(run["session"].as_str().map(str::len), Some(16), "{command}");
        assert!(run["unix_ms"].as_u64().unwrap() > 0, "{command}");
    }
    let log = merge_log(&root);
    let observes = log.iter().filter(|l| l["kind"] == json!("observe")).count();
    assert_eq!(observes, 5, "{log:?}");
    let text = fs::read_to_string(root.join(".moai-cli/receipts/hook-engine.jsonl")).unwrap();
    assert!(!text.contains("cargo test"), "a command leaked: {text}");
}

/// REQ-1, REQ-2: what is not exactly a verify command, a failed call, a call outside a SPEC
/// work branch and a SPEC, commit or state folder that cannot be read leave no `ac_run` line,
/// and every one of those calls goes through.
#[test]
fn ac_verify_run_ignores_what_is_not_the_verify_command() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    for command in [
        "cargo test --test one extra",
        "cargo test --test one | tail -3",
        "FOO=1 cargo test --test one",
        "cargo test --test",
        "cargo test --test one; true",
        "read the page",
        "no way yet",
        "",
    ] {
        let (code, out) = ac_bash(&root, "s1", command, false);
        assert_eq!((code, out.as_str()), (0, ""), "{command:?}");
    }
    let (code, out) = ac_bash(&root, "s1", "cargo test --test one", true);
    assert_eq!((code, out.as_str()), (0, ""), "a failed call");
    assert_eq!(ac_runs(&root), Vec::<Value>::new());

    for (branch, head, spec) in [
        ("main", Some(AC_HEAD), Some(AC_BODY)),
        (AC_BRANCH, Some(AC_HEAD), None),
        (AC_BRANCH, None, Some(AC_BODY)),
        (AC_BRANCH, Some("not-a-commit"), Some(AC_BODY)),
        ("spec/20261119-other", Some(AC_HEAD), Some(AC_BODY)),
    ] {
        let (_tmp, root) = ac_repo(branch, head, spec);
        let (code, out) = ac_bash(&root, "s1", "cargo test --test one", false);
        assert_eq!((code, out.as_str()), (0, ""), "{branch} {head:?}");
        assert_eq!(
            ac_runs(&root),
            Vec::<Value>::new(),
            "{branch} {head:?} {spec:?}"
        );
    }

    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    fs::remove_dir_all(root.join(".moai-cli")).unwrap();
    let (code, out) = ac_bash(&root, "s1", "cargo test --test one", false);
    assert_eq!((code, out.as_str()), (0, ""), "no state folder");
    assert!(!root.join(".moai-cli").exists(), "nothing is created");
}

/// REQ-1 (tier 2 audit B1): a Bash call that did not run to its end is not a successful run of
/// the verify command: one started in the background (`run_in_background`, or a
/// `backgroundTaskId` in its response), one the host reports as interrupted, and an input that
/// does not say which event it is leave no `ac_run` line. The same call as a finished
/// `PostToolUse` is recorded, and every one of these calls still leaves its observe line.
#[test]
fn ac_verify_run_ignores_calls_that_did_not_finish() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    let command = "cargo test --test one";
    let call = |event: Option<&str>, tool_input: Value, tool_response: Value| {
        let mut input = json!({"session_id": "s1", "tool_name": "Bash", "tool_use_id": "t",
            "tool_input": tool_input, "tool_response": tool_response, "cwd": root});
        if let Some(event) = event {
            input["hook_event_name"] = json!(event);
        }
        ac_dispatch("tool-observe", "claude", &root, &input)
    };
    let not_finished = [
        (
            "run_in_background",
            Some("PostToolUse"),
            json!({"command": command, "run_in_background": true}),
            json!({"backgroundTaskId": "b1"}),
        ),
        (
            "run_in_background without a task id",
            Some("PostToolUse"),
            json!({"command": command, "run_in_background": true}),
            json!({}),
        ),
        (
            "a backgroundTaskId alone",
            Some("PostToolUse"),
            json!({"command": command}),
            json!({"backgroundTaskId": "b1"}),
        ),
        (
            "interrupted",
            Some("PostToolUse"),
            json!({"command": command}),
            json!({"interrupted": true}),
        ),
        (
            "no event name",
            None,
            json!({"command": command}),
            json!({}),
        ),
    ];
    for (what, event, tool_input, tool_response) in not_finished {
        let (code, out) = call(event, tool_input, tool_response);
        assert_eq!((code, out.as_str()), (0, ""), "{what}");
        assert_eq!(ac_runs(&root), Vec::<Value>::new(), "{what}");
    }
    let observes = merge_log(&root)
        .iter()
        .filter(|l| l["kind"] == json!("observe"))
        .count();
    assert_eq!(observes, 5, "every call is still observed");
    // The same call, finished: one line. `false` flags do not hide it.
    let (code, out) = call(
        Some("PostToolUse"),
        json!({"command": command, "run_in_background": false}),
        json!({"interrupted": false, "stdout": "ok"}),
    );
    assert_eq!((code, out.as_str()), (0, ""));
    assert_eq!(ac_runs(&root).len(), 1, "{:?}", ac_runs(&root));
}

/// REQ-3, REQ-6: after the session wrote a file, the Stop answer names the verify commands
/// (by AC) that have no `ac_run` line at the current commit, written at or after that write;
/// criteria with `manual:`, `unknown:` or a placeholder are not named; a new write or a new
/// commit makes the older lines not count; a line of another session counts.
#[test]
fn ac_verify_rule_stop_asks_for_unrecorded_commands() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    ac_mode(&root, "enforce");
    assert_eq!(
        ac_reason(&ac_stop(&root, "s1", false)),
        None,
        "no write yet"
    );
    ac_write(&root, "s1");
    ac_pause();
    let reason = ac_reason(&ac_stop(&root, "s1", false)).expect("asks");
    assert!(reason.contains(AC_SPEC), "{reason}");
    for ac in ["AC-1", "AC-2", "AC-6"] {
        assert!(reason.contains(ac), "{ac}: {reason}");
    }
    for ac in ["AC-3", "AC-4", "AC-5"] {
        assert!(!reason.contains(ac), "{ac}: {reason}");
    }

    ac_bash(&root, "s1", "cargo test --test one", false);
    let reason = ac_reason(&ac_stop(&root, "s1", false)).expect("still asks");
    assert!(!reason.contains("AC-1"), "{reason}");
    assert!(
        reason.contains("AC-2") && reason.contains("AC-6"),
        "{reason}"
    );

    ac_bash(&root, "s1", "cargo test --test two -- --exact x", false);
    ac_bash(&root, "s1", "cargo test --test six_a", false);
    let reason = ac_reason(&ac_stop(&root, "s1", false)).expect("one command of AC-6 is left");
    assert!(
        reason.contains("AC-6") && !reason.contains("AC-2"),
        "{reason}"
    );

    ac_bash(&root, "s1", "cargo test --test six_b", false);
    assert_eq!(
        ac_reason(&ac_stop(&root, "s1", false)),
        None,
        "all recorded"
    );

    // A write after the runs: the earlier lines do not count.
    ac_pause();
    ac_write(&root, "s1");
    ac_pause();
    let reason = ac_reason(&ac_stop(&root, "s1", false)).expect("a later write");
    for ac in ["AC-1", "AC-2", "AC-6"] {
        assert!(reason.contains(ac), "{ac}: {reason}");
    }
    for command in [
        "cargo test --test one",
        "cargo test --test two -- --exact x",
        "cargo test --test six_a",
        "cargo test --test six_b",
    ] {
        ac_bash(&root, "s1", command, false);
    }
    assert_eq!(ac_reason(&ac_stop(&root, "s1", false)), None);

    // A new commit: the lines of the old one do not count.
    fs::write(
        root.join(".git/refs/heads").join(AC_BRANCH),
        format!("{AC_HEAD_2}\n"),
    )
    .unwrap();
    assert!(
        ac_reason(&ac_stop(&root, "s1", false)).is_some(),
        "new commit"
    );
    // Another session that wrote nothing is not asked; its runs at the new commit count for
    // the first session.
    assert_eq!(
        ac_reason(&ac_stop(&root, "s2", false)),
        None,
        "wrote nothing"
    );
    for command in [
        "cargo test --test one",
        "cargo test --test two -- --exact x",
        "cargo test --test six_a",
        "cargo test --test six_b",
    ] {
        ac_bash(&root, "s2", command, false);
    }
    assert_eq!(
        ac_reason(&ac_stop(&root, "s1", false)),
        None,
        "recorded by s2"
    );
    let runs = ac_runs(&root);
    assert_eq!(runs.last().unwrap()["head"], json!(AC_HEAD_2));
}

/// REQ-4, REQ-6: the turn ends, although the session wrote a file, when the branch is not a SPEC
/// work branch, the SPEC or the commit cannot be read, or no criterion has a runnable verify
/// command.
#[test]
fn ac_verify_rule_stop_lets_the_turn_end() {
    for (branch, head, spec) in [
        ("main", Some(AC_HEAD), Some(AC_BODY)),
        (AC_BRANCH, Some(AC_HEAD), None),
        (AC_BRANCH, None, Some(AC_BODY)),
        (AC_BRANCH, Some("not-a-commit"), Some(AC_BODY)),
        (AC_BRANCH, Some(AC_HEAD), Some(AC_BODY_LEFT_OUT)),
        (AC_BRANCH, Some(AC_HEAD), Some("no acceptance here\n")),
        ("spec/20261119-other", Some(AC_HEAD), Some(AC_BODY)),
    ] {
        let (_tmp, root) = ac_repo(branch, head, spec);
        ac_mode(&root, "enforce");
        ac_write(&root, "s1");
        ac_pause();
        let answer = ac_stop(&root, "s1", false);
        assert_eq!(
            (answer.0, ac_reason(&answer)),
            (0, None),
            "{branch} {head:?} {spec:?}"
        );
        assert!(
            merge_log(&root)
                .iter()
                .all(|l| l["kind"] != json!("stop_block") && l["kind"] != json!("would_block")),
            "{branch} {head:?} {spec:?}"
        );
    }
}

/// REQ-5: `off` writes nothing, the default and `observe` write one `would_block` line and let
/// the turn end, `enforce` keeps the session working twice in a row and lets the third turn
/// end; a turn the host already continues ends at once.
#[test]
fn ac_verify_rule_modes_off_observe_enforce() {
    for config in [
        None,
        Some(""),
        Some("[hooks.modes]\n"),
        Some("[hooks.modes]\nac_verify_recorded = \"observe\"\n"),
        Some("[hooks.modes]\nac_verify_recorded = \"yes\"\n"),
        Some("[hooks.modes]\nmerge_confirm = \"enforce\"\n"),
        Some("this is = not [valid toml"),
    ] {
        let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
        if let Some(text) = config {
            merge_config(&root, text);
        }
        ac_write(&root, "s1");
        ac_pause();
        let answer = ac_stop(&root, "s1", false);
        assert_eq!((answer.0, ac_reason(&answer)), (0, None), "{config:?}");
        assert_eq!(answer.1, "", "{config:?}");
        let log = merge_log(&root);
        let would: Vec<&Value> = log
            .iter()
            .filter(|l| l["kind"] == json!("would_block"))
            .collect();
        assert_eq!(would.len(), 1, "{config:?}: {log:?}");
        assert_eq!(would[0]["rule"], json!("ac_verify_recorded"), "{config:?}");
        assert!(
            would[0]["reason"].as_str().unwrap().contains("AC-1"),
            "{config:?}"
        );
        assert!(
            log.iter().all(|l| l["kind"] != json!("stop_block")),
            "{config:?}: {log:?}"
        );
    }

    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    ac_mode(&root, "off");
    ac_write(&root, "s1");
    ac_pause();
    assert_eq!(ac_stop(&root, "s1", false), (0, String::new()));
    assert!(
        merge_log(&root)
            .iter()
            .all(|l| l["kind"] == json!("observe")),
        "off writes nothing but the observe line"
    );

    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    ac_mode(&root, "enforce");
    ac_write(&root, "s1");
    ac_pause();
    assert!(
        ac_reason(&ac_stop(&root, "s1", true)).is_none(),
        "already continued"
    );
    assert!(ac_reason(&ac_stop(&root, "s1", false)).is_some(), "first");
    assert!(ac_reason(&ac_stop(&root, "s1", false)).is_some(), "second");
    assert!(
        ac_reason(&ac_stop(&root, "s1", false)).is_none(),
        "third ends"
    );
    let blocks = merge_log(&root)
        .iter()
        .filter(|l| l["kind"] == json!("stop_block") && l["rule"] == json!("ac_verify_recorded"))
        .count();
    assert_eq!(blocks, 2);
}

/// REQ-7: Claude's and Codex's capabilities name the rule, its modes and its named limits;
/// OMP still says it is not wired, without the old "STOP_RULES is empty" reason.
#[test]
fn ac_verify_rule_capabilities_name_the_rule_and_its_limits() {
    for host in ["claude", "codex"] {
        let caps = moai_cli::hook_engine::capabilities(host);
        let tv = &caps["turn-validation"];
        assert_eq!(tv["rules"], json!(1), "{host}: {tv}");
        assert_eq!(
            tv["rule_ids"],
            json!(["ac_verify_recorded"]),
            "{host}: {tv}"
        );
        let modes = tv["modes"].to_string();
        for word in [
            "off",
            "observe",
            "enforce",
            "[hooks.modes]",
            "ac_verify_recorded",
        ] {
            assert!(modes.contains(word), "{host} {word}: {modes}");
        }
        let limits = tv["limits"].to_string();
        for limit in [
            "do not call tool-observe or turn-validation",
            "a commit changes HEAD",
            "shell writes (sed -i) are not writes",
            "must equal the verify command",
            "plain file",
        ] {
            assert!(limits.contains(limit), "{host} {limit}: {limits}");
        }
    }
    let omp = moai_cli::hook_engine::capabilities("omp");
    let reason = omp["turn-validation"]["reason"].as_str().unwrap();
    assert!(reason.contains("not wired"), "{reason}");
    assert!(!reason.contains("STOP_RULES is empty"), "{reason}");
}

/// SPEC-20261119-ac-verify-recorded-limits REQ-1 (card C-130, audit N2 N5 N8 and the subfolder
/// decision D-2): the named limits also say that Stop reads the process folder while the
/// observation reads the input's cwd, that a verify command over 4096 bytes is never recorded,
/// that a detached HEAD, sha256 objects, a reftable and a symbolic ref switch the rule off
/// silently, and that a run from a subfolder counts.
#[test]
fn the_ac_verify_limits_name_the_process_folder_the_command_size_the_unreadable_heads_and_the_subfolder(
) {
    for host in ["claude", "codex"] {
        let caps = moai_cli::hook_engine::capabilities(host);
        let limits = caps["turn-validation"]["limits"].to_string();
        for limit in [
            "the process folder",
            "over 4096 bytes",
            "a detached HEAD",
            "sha256",
            "reftable",
            "symbolic ref",
            "from a subfolder",
        ] {
            assert!(limits.contains(limit), "{host} {limit}: {limits}");
        }
    }
}

/// Three blocks of one session in a row: the first two keep it working, the third lets the turn
/// end and (REQ-2) writes one `rule_release` line that names the rule, its reason and the
/// session. The release is not a block: the count of `stop_block` lines stays two.
#[test]
fn the_stop_release_after_the_cap_writes_one_rule_release_line() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    ac_mode(&root, "enforce");
    ac_write(&root, "s1");
    ac_pause();
    assert!(ac_reason(&ac_stop(&root, "s1", false)).is_some(), "first");
    assert!(ac_reason(&ac_stop(&root, "s1", false)).is_some(), "second");
    let released = |root: &Path| {
        merge_log(root)
            .into_iter()
            .filter(|l| l["kind"] == json!("rule_release"))
            .collect::<Vec<Value>>()
    };
    assert_eq!(
        released(&root),
        Vec::<Value>::new(),
        "no release before the cap"
    );
    assert!(
        ac_reason(&ac_stop(&root, "s1", false)).is_none(),
        "third ends"
    );
    let lines = released(&root);
    assert_eq!(lines.len(), 1, "{lines:?}");
    assert_eq!(lines[0]["rule"], json!("ac_verify_recorded"), "{lines:?}");
    // The log keeps a key of the session id, the same one its `stop_block` lines carry.
    let block_session = merge_log(&root)
        .into_iter()
        .find(|l| l["kind"] == json!("stop_block"))
        .map(|l| l["session"].clone())
        .unwrap();
    assert_eq!(lines[0]["session"], block_session, "{lines:?}");
    assert!(
        lines[0]["reason"].as_str().unwrap().contains("AC-1"),
        "{lines:?}"
    );
    assert!(lines[0]["unix_ms"].as_u64().is_some(), "{lines:?}");
    let blocks = merge_log(&root)
        .iter()
        .filter(|l| l["kind"] == json!("stop_block"))
        .count();
    assert_eq!(blocks, 2, "the release is not a block");
    // Another session is not released by this session's count.
    ac_write(&root, "s2");
    ac_pause();
    assert!(
        ac_reason(&ac_stop(&root, "s2", false)).is_some(),
        "s2 first"
    );
    assert_eq!(released(&root).len(), 1, "s2 has no release");
}

/// REQ-2's other side: a session at the cap whose rule would not ask any more (the SPEC changed
/// to criteria the rule leaves out, with no tool call in between) is let go without a
/// `rule_release` line: there was nothing to release.
#[test]
fn no_rule_release_line_is_written_when_no_rule_would_block() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    ac_mode(&root, "enforce");
    ac_write(&root, "s1");
    ac_pause();
    assert!(ac_reason(&ac_stop(&root, "s1", false)).is_some(), "first");
    assert!(ac_reason(&ac_stop(&root, "s1", false)).is_some(), "second");
    fs::write(
        root.join(".moai/specs").join(AC_SPEC).join("spec.md"),
        format!("{AC_HEAD_PART}{AC_BODY_LEFT_OUT}"),
    )
    .unwrap();
    assert!(
        ac_reason(&ac_stop(&root, "s1", false)).is_none(),
        "nothing to ask"
    );
    assert!(
        merge_log(&root)
            .iter()
            .all(|l| l["kind"] != json!("rule_release")),
        "no release without a refusal to release"
    );
}

/// REQ-8 (audit N6): the engine log has a size limit. Up to the limit the rule reads the log and
/// asks; at exactly the limit it reads, asks, and cannot write the block down, so it lets the
/// turn end (an unrecorded block could not be bounded). Past the limit it reads no line and lets
/// the turn end too. The last two cases look the same from outside, so this test pins the
/// write limit and the limit-1 read; the read limit itself is pinned only on the oversized-side
/// lib test (`an_oversized_log_on_one_side_does_not_hide_the_other_side`), and a read limit of
/// `<` instead of `<=` survives (audit note on C-130, finding 1).
#[test]
fn the_ac_verify_rule_lets_the_turn_end_when_the_log_is_full() {
    let limit = moai_cli::hooks::RECEIPT_LOG_LIMIT;
    for (size, asks) in [(limit - 1, true), (limit, false), (limit + 1, false)] {
        let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
        ac_mode(&root, "enforce");
        ac_write(&root, "s1");
        ac_pause();
        // Grow the log to `size` bytes with one long valid JSON line.
        let log = root.join(".moai-cli/receipts/hook-engine.jsonl");
        let now = fs::metadata(&log).unwrap().len();
        let pad = (size - now) as usize;
        let filler = "{\"kind\":\"pad\",\"x\":\"".len() + "\"}\n".len();
        let line = format!(
            "{{\"kind\":\"pad\",\"x\":\"{}\"}}\n",
            "a".repeat(pad - filler)
        );
        let mut file = fs::OpenOptions::new().append(true).open(&log).unwrap();
        std::io::Write::write_all(&mut file, line.as_bytes()).unwrap();
        drop(file);
        assert_eq!(fs::metadata(&log).unwrap().len(), size, "size {size}");
        let before = fs::metadata(&log).unwrap().len();
        let answer = ac_stop(&root, "s1", false);
        assert_eq!(
            ac_reason(&answer).is_some(),
            asks,
            "log of {size} bytes (limit {limit}): {answer:?}"
        );
        if !asks {
            assert_eq!(
                fs::metadata(&log).unwrap().len(),
                before,
                "a full log is never grown (size {size})"
            );
        }
    }
}

/// Decision D-2 (audit N4), pinned: the hook process runs in the project's folder and the call
/// names its own cwd. A verify command run from a subfolder of the work tree (the call's cwd is
/// below the top) leaves its `ac_run` line in the project's log, because the SPEC is found from
/// the top of the work tree the call's cwd is in. A call whose cwd is another work tree is read
/// with that work tree's branch, not the project's.
#[test]
fn a_run_from_a_subfolder_of_the_work_tree_is_recorded() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    let sub = root.join("src/deep");
    fs::create_dir_all(&sub).unwrap();
    let input = json!({"session_id": "s1", "hook_event_name": "PostToolUse",
        "tool_name": "Bash", "tool_use_id": "t",
        "tool_input": {"command": "cargo test --test one"}, "tool_response": {}, "cwd": sub});
    let (code, out) = ac_dispatch("tool-observe", "claude", &root, &input);
    assert_eq!((code, out.as_str()), (0, ""));
    let runs = ac_runs(&root);
    assert_eq!(runs.len(), 1, "{runs:?}");
    assert_eq!(runs[0]["ac"], json!("AC-1"), "{runs:?}");
    assert_eq!(runs[0]["head"], json!(AC_HEAD), "{runs:?}");
    // A different work tree (a branch that is no SPEC branch) records nothing.
    let (_other_tmp, other) = ac_repo("main", Some(AC_HEAD), Some(AC_BODY));
    let input = json!({"session_id": "s1", "hook_event_name": "PostToolUse",
        "tool_name": "Bash", "tool_use_id": "t",
        "tool_input": {"command": "cargo test --test one"}, "tool_response": {}, "cwd": other});
    let (code, _) = ac_dispatch("tool-observe", "claude", &root, &input);
    assert_eq!(code, 0);
    assert_eq!(ac_runs(&root).len(), 1, "the other work tree adds no run");
}

/// REQ-7 (audit N-r1; the host documents read 2026-11-19 by research-c130-host-shapes):
/// Claude Code documents `run_in_background` as a boolean, `backgroundTaskId` as a string only
/// for background commands, `interrupted` as a boolean, and sets `timedOutAfterMs` or
/// `backgroundedByUser` when it moved the command to the background itself. A recorder that must
/// not count an unfinished call reads every value that is not the documented plain "no" as
/// unfinished: the string "true", the number 1, an empty task id, a task id in snake_case, a
/// key that is present with another type. A `null` or an absent key, and the plain `false`, are
/// a finished call and are recorded.
#[test]
fn ac_verify_run_ignores_a_call_marked_unfinished_in_every_documented_spelling() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    let command = "cargo test --test one";
    let call = |tool_input: Value, tool_response: Value| {
        let mut input_tool = json!({"command": command});
        for (k, v) in tool_input.as_object().unwrap() {
            input_tool[k] = v.clone();
        }
        let input = json!({"session_id": "s1", "hook_event_name": "PostToolUse",
            "tool_name": "Bash", "tool_use_id": "t",
            "tool_input": input_tool, "tool_response": tool_response, "cwd": root});
        let (code, out) = ac_dispatch("tool-observe", "claude", &root, &input);
        assert_eq!((code, out.as_str()), (0, ""));
    };
    let unfinished: [(&str, Value, Value); 15] = [
        (
            "run_in_background string",
            json!({"run_in_background": "true"}),
            json!({}),
        ),
        (
            "run_in_background one",
            json!({"run_in_background": 1}),
            json!({}),
        ),
        (
            "run_in_background false string",
            json!({"run_in_background": "false"}),
            json!({}),
        ),
        (
            "runInBackground",
            json!({"runInBackground": true}),
            json!({}),
        ),
        ("empty task id", json!({}), json!({"backgroundTaskId": ""})),
        ("numeric task id", json!({}), json!({"backgroundTaskId": 7})),
        (
            "snake_case task id",
            json!({}),
            json!({"background_task_id": "b1"}),
        ),
        (
            "interrupted string",
            json!({}),
            json!({"interrupted": "true"}),
        ),
        ("interrupted one", json!({}), json!({"interrupted": 1})),
        (
            "interrupted false string",
            json!({}),
            json!({"interrupted": "false"}),
        ),
        ("interrupted zero", json!({}), json!({"interrupted": 0})),
        (
            "timed out into the background",
            json!({}),
            json!({"timedOutAfterMs": 120000}),
        ),
        (
            "moved by the user",
            json!({}),
            json!({"backgroundedByUser": true}),
        ),
        (
            "moved by the user string",
            json!({}),
            json!({"backgroundedByUser": "true"}),
        ),
        ("response is a string", json!({}), json!("done")),
    ];
    for (what, tool_input, tool_response) in unfinished {
        call(tool_input, tool_response);
        assert_eq!(ac_runs(&root), Vec::<Value>::new(), "{what}");
    }
    // The finished shapes: the documented plain values, `null` and absent keys.
    for (what, tool_input, tool_response) in [
        (
            "plain false flags",
            json!({"run_in_background": false}),
            json!({"interrupted": false, "backgroundedByUser": false, "stdout": "ok"}),
        ),
        (
            "null markers",
            json!({}),
            json!({"backgroundTaskId": null, "timedOutAfterMs": null, "interrupted": false}),
        ),
    ] {
        let before = ac_runs(&root).len();
        call(tool_input, tool_response);
        assert_eq!(ac_runs(&root).len(), before + 1, "{what}");
    }
}

/// REQ-3 (audit N3): the verify command's white space is folded as the SPEC folds it only
/// where the shell folds it. A call with a newline or a no-break space inside the command is a
/// different command and is not recorded as the verify command; two spaces or a tab are the
/// same command.
#[test]
fn a_call_with_a_newline_or_a_no_break_space_inside_is_not_recorded_as_the_verify_command() {
    let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
    for different in [
        "cargo test\n--test one",
        "cargo test\u{a0}--test one",
        "cargo test --test\none",
    ] {
        let (code, out) = ac_bash(&root, "s1", different, false);
        assert_eq!((code, out.as_str()), (0, ""), "{different:?}");
        assert_eq!(ac_runs(&root), Vec::<Value>::new(), "{different:?}");
    }
    for same in [
        "cargo  test --test one",
        "cargo test\t--test one",
        " cargo test --test one ",
        "cargo test --test one\n",
    ] {
        let (code, out) = ac_bash(&root, "s1", same, false);
        assert_eq!((code, out.as_str()), (0, ""), "{same:?}");
    }
    assert_eq!(ac_runs(&root).len(), 4, "{:?}", ac_runs(&root));
}

/// SPEC-20261119-ac-verify-recorded-limits REQ-10 (card C-130 criterion 5, leader decision of
/// note 3387 on audit N10): the rule counts a write of the session only when its outcome is not a
/// failure (a success and an unknown outcome both count: dropping the unknown would weaken the
/// rule quietly) and its path is not outside the project folder; a line with no path counts,
/// conservatively. After a write that does not count the Stop rule has no write to ask about.
#[test]
fn ac_verify_rule_counts_only_writes_that_did_not_fail_and_are_inside_the_project() {
    let project_file = |root: &Path| json!({"file_path": root.join("src/x.rs"), "content": "x"});
    // (what, host, event, tool, tool_input builder, counts)
    type WriteCase<'a> = (
        &'a str,
        &'a str,
        &'a str,
        &'a str,
        Box<dyn Fn(&Path) -> Value>,
        bool,
    );
    let cases: Vec<WriteCase> = vec![
        (
            "a project write that succeeded",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(project_file),
            true,
        ),
        (
            "a failed Edit",
            "claude",
            "PostToolUseFailure",
            "Edit",
            Box::new(project_file),
            false,
        ),
        (
            "a write outside the project folder",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|_| json!({"file_path": "/tmp/moai-c130-scratch.txt", "content": "x"})),
            false,
        ),
        (
            "a write that climbs out of the folder",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|_| json!({"file_path": "../elsewhere.txt", "content": "x"})),
            false,
        ),
        (
            "an unknown outcome (Codex apply_patch)",
            "codex",
            "PostToolUse",
            "apply_patch",
            Box::new(project_file),
            true,
        ),
        (
            "a write line with no path",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|_| json!({"content": "x"})),
            true,
        ),
        // Audit of 0ca2f8892b, findings 2 and 3: a path that only passes through `..` and stays
        // inside is a project write; a path the log cannot keep (too long, a control character)
        // is `unrecorded` and counts, as the rule counts what it cannot place.
        (
            "a relative path that goes through .. and stays inside",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|_| json!({"file_path": "src/../x.rs", "content": "x"})),
            true,
        ),
        (
            "an absolute path that goes through .. and stays inside",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|root| json!({"file_path": root.join("src/../x.rs"), "content": "x"})),
            true,
        ),
        (
            "an absolute path that climbs out of the folder from inside it",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|root| json!({"file_path": root.join("src/../../x.rs"), "content": "x"})),
            false,
        ),
        (
            "a path longer than the log keeps",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(
                |_| json!({"file_path": format!("src/{}.rs", "a".repeat(1100)), "content": "x"}),
            ),
            true,
        ),
        (
            "a path with a control character",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|_| json!({"file_path": "src/x\u{7}.rs", "content": "x"})),
            true,
        ),
        (
            "an empty path",
            "claude",
            "PostToolUse",
            "Write",
            Box::new(|_| json!({"file_path": "", "content": "x"})),
            true,
        ),
    ];
    for (what, host, event, tool, tool_input, counts) in cases {
        let (_tmp, root) = ac_repo(AC_BRANCH, Some(AC_HEAD), Some(AC_BODY));
        ac_mode(&root, "enforce");
        let input = json!({"session_id": "s1", "hook_event_name": event,
            "tool_name": tool, "tool_use_id": "t",
            "tool_input": tool_input(&root), "tool_response": {}, "cwd": root});
        let (code, _) = ac_dispatch("tool-observe", host, &root, &input);
        assert_eq!(code, 0, "{what}");
        // Every one of these calls is a write line of the session in the log.
        let writes = merge_log(&root)
            .iter()
            .filter(|l| l["kind"] == json!("observe") && l["class"] == json!("write"))
            .count();
        assert_eq!(writes, 1, "{what}: the call is observed as a write");
        ac_pause();
        let answer = ac_stop(&root, "s1", false);
        assert_eq!(
            ac_reason(&answer).is_some(),
            counts,
            "{what}: {answer:?} {:?}",
            merge_log(&root)
        );
    }
}
