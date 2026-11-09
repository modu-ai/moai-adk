//! rc.2 I1 #2238: host hook generation and the hook-receipt handler. Every case runs in a
//! temporary project and a temporary HOME; no real home and no host session is touched.

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
    assert_eq!(applied["hooks_generated"].as_array().unwrap().len(), 3);
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

/// The hook set itself: the three events, an explicit timeout on every handler, and a bare
/// command name (no absolute path) that is the same op for both hosts.
#[test]
fn hook_set_has_explicit_timeouts_and_bare_commands() {
    for host in hooks::HOSTS {
        let obj = hooks::hooks_object(host).unwrap();
        let events: Vec<&String> = obj.keys().collect();
        assert_eq!(events, ["SessionStart", "PostToolUse", "Stop"]);
        assert_eq!(obj["PostToolUse"][0]["matcher"], json!("Bash"));
        assert!(obj["Stop"][0].get("matcher").is_none());
        for groups in obj.values() {
            let handler = &groups[0]["hooks"][0];
            assert_eq!(handler["type"], json!("command"));
            assert_eq!(handler["timeout"], json!(hooks::HOOK_TIMEOUT_SECS));
            let command = handler["command"].as_str().unwrap();
            assert_eq!(command, format!("moai-cli hook-receipt --host {host}"));
            assert!(!command.contains('/'), "{command}");
        }
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
#[test]
fn home_directory_is_refused() {
    let home = project();
    for host in hooks::HOSTS {
        let err = hooks::plan(home.path(), host, Some(home.path())).unwrap_err();
        assert_eq!(classify(&err), Kind::InvalidInput, "{err}");
        let err = hooks::apply(home.path(), host, Some(home.path()), "x", PERMISSION).unwrap_err();
        assert!(err.contains("home directory refused"), "{err}");
    }
    for dir in [".claude", ".codex", ".agents", ".moai-cli"] {
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
    use std::io::Write as _;
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child.stdin.take().unwrap().write_all(stdin).unwrap();
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

    fs::create_dir(p.path().join(".moai-cli/receipts")).ok();
    let (code, rec) = cli(
        &["hook-receipt", "--host", "codex"],
        p.path(),
        home.path(),
        br#"{"hook_event_name":"SessionStart","source":"startup"}"#,
    );
    assert_eq!(code, 0);
    assert_eq!(rec["operation"], json!("hook.receipt"));
    assert_eq!(rec["result"]["recorded"], json!(true), "{rec}");
    let (code, rec) = cli(
        &["hook-receipt", "--host", "codex"],
        p.path(),
        home.path(),
        b"{oops",
    );
    assert_eq!(code, 0);
    assert_eq!(rec["result"]["recorded"], json!(false));
}
