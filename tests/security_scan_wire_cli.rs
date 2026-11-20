//! SPEC-20261120-security-scan-wire: the generated Claude Code hook set runs `hook-dispatch
//! --event security-scan` after each write tool (PostToolUse), the hosts codex and omp get no such
//! group, `hooks-plan` brings a project that holds the untouched set of the release before this
//! change to the new one as an `update` (and never overwrites a hand-edited one), and the
//! `would_notify` lines of the event have a log of their own whose size limit cannot stop the
//! deny lines of the engine log.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use moai_cli::{hook_engine, hooks};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const PERMISSION: &str = "test-permission-wire";
const SCAN_COMMAND: &str = "moai hook-dispatch --host claude --event security-scan";

/// The tool names the engine lists for the event, joined the way a host matcher is written.
fn capabilities_matcher() -> String {
    hook_engine::capabilities("claude")["security-scan"]["tools"]
        .as_array()
        .expect("the capabilities of claude list the tools of security-scan")
        .iter()
        .map(|t| t.as_str().unwrap())
        .collect::<Vec<_>>()
        .join("|")
}

/// The commands of the groups of one event of a hooks object, in order.
fn commands(obj: &serde_json::Map<String, Value>, event: &str) -> Vec<String> {
    obj[event]
        .as_array()
        .unwrap()
        .iter()
        .map(|g| g["hooks"][0]["command"].as_str().unwrap().to_string())
        .collect()
}

fn read_json(path: &Path) -> Value {
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

/// AC-1 (REQ-1, REQ-2, REQ-4, REQ-5): after the Bash receipt group comes the security-scan group,
/// with the matcher of the capabilities answer, the hook timeout and a bare command; the other
/// events and groups are as they were.
#[test]
fn claude_runs_security_scan_after_the_receipt_group() {
    let obj = hooks::hooks_object("claude").unwrap();
    let events: Vec<&String> = obj.keys().collect();
    // The turn-end scan adds `Stop` last (SPEC-20261121-security-turn-wire).
    assert_eq!(
        events,
        ["SessionStart", "PreToolUse", "PostToolUse", "Stop"]
    );
    let post = obj["PostToolUse"].as_array().unwrap();
    assert_eq!(post.len(), 2, "{post:?}");
    assert_eq!(post[0]["matcher"], json!("Bash"));
    assert_eq!(
        post[0]["hooks"][0]["command"],
        json!("moai hook-receipt --host claude")
    );
    assert_eq!(post[1]["matcher"], json!(capabilities_matcher()));
    assert_eq!(
        post[1]["matcher"],
        json!("Write|Edit|MultiEdit|NotebookEdit")
    );
    let handler = &post[1]["hooks"][0];
    assert_eq!(handler["type"], json!("command"));
    assert_eq!(handler["command"], json!(SCAN_COMMAND));
    assert_eq!(handler["timeout"], json!(hooks::HOOK_TIMEOUT_SECS));
    assert_eq!(obj["PreToolUse"].as_array().unwrap().len(), 2);
    assert_eq!(obj["SessionStart"].as_array().unwrap().len(), 1);
}

/// AC-1 (REQ-4), the generated order while `[hooks.edit_diagnostics]` is enabled: receipt,
/// security-scan, edit-diagnostics. Run through the binary with a temporary HOME.
#[test]
fn claude_places_security_scan_before_edit_diagnostics_when_that_rule_is_enabled() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (proj, home) = (root.join("proj"), root.join("home"));
    fs::create_dir_all(&proj).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    fs::write(
        home.join(".moai-cli/hook-config.toml"),
        "[hooks.edit_diagnostics]\nenabled = true\ncheckers = { rs = [\"x\"] }\n",
    )
    .unwrap();
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", &home)
        .current_dir(&proj)
        .args(["hooks-plan", proj.to_str().unwrap(), "--host", "claude"])
        .output()
        .unwrap();
    assert!(out.status.success(), "{out:?}");
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let listed: Vec<String> = v["result"]["hooks"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|h| h["event"] == json!("PostToolUse"))
        .map(|h| h["command"].as_str().unwrap_or("").to_string())
        .collect();
    let at = |needle: &str| listed.iter().position(|c| c.contains(needle));
    let (receipt, scan, diagnostics) = (
        at("hook-receipt"),
        at("--event security-scan"),
        at("--event edit-diagnostics"),
    );
    assert!(
        receipt.is_some() && scan.is_some() && diagnostics.is_some(),
        "{listed:?}"
    );
    assert!(receipt < scan && scan < diagnostics, "{listed:?}");
}

/// AC-2 (REQ-3): codex and omp run no security-scan group and keep their other groups.
#[test]
fn other_hosts_get_no_security_scan_group() {
    for host in ["codex", "omp"] {
        let obj = hooks::hooks_object(host).unwrap();
        for event in obj.keys() {
            for command in commands(&obj, event) {
                assert!(
                    !command.contains("security-scan"),
                    "{host}: {event}: {command}"
                );
            }
        }
        let events: Vec<&String> = obj.keys().collect();
        assert_eq!(
            events,
            ["SessionStart", "PreToolUse", "PostToolUse"],
            "{host}"
        );
        assert_eq!(obj["PostToolUse"].as_array().unwrap().len(), 1, "{host}");
        assert_eq!(obj["PreToolUse"].as_array().unwrap().len(), 1, "{host}");
    }
}

/// AC-5 (REQ-12): the legacy set (the old binary name) holds no security-scan group, so the files
/// of the releases before the rename are still recognised; the same binary name in the current set
/// does hold it.
#[test]
fn legacy_set_holds_no_security_scan_group() {
    let legacy = hooks::legacy_hooks("claude", None).unwrap();
    for event in legacy.keys() {
        for command in commands(&legacy, event) {
            assert!(command.starts_with("moai-cli "), "{command}");
            assert!(!command.contains("security-scan"), "{event}: {command}");
        }
    }
    assert_eq!(legacy["PostToolUse"].as_array().unwrap().len(), 1);
    let current = hooks::current_hooks("claude", None).unwrap();
    assert_eq!(current["PostToolUse"].as_array().unwrap().len(), 2);
}

/// Writes the settings and the ownership receipt of a project as if the release before this
/// change had applied its set: the current hooks object without the security-scan group, next to
/// one key MoAI does not own.
fn install_previous_release_set(root: &Path) -> Value {
    let mut previous = hooks::hooks_object("claude").unwrap();
    let post = previous
        .get_mut("PostToolUse")
        .and_then(Value::as_array_mut)
        .unwrap();
    post.retain(|g| {
        !g["hooks"][0]["command"]
            .as_str()
            .unwrap()
            .contains("security-scan")
    });
    assert_eq!(post.len(), 1, "the previous set has the receipt group only");
    let settings = json!({"model": "kept", "hooks": Value::Object(previous.clone())});
    let mut bytes = serde_json::to_vec_pretty(&settings).unwrap();
    bytes.push(b'\n');
    fs::create_dir_all(root.join(".claude")).unwrap();
    fs::write(root.join(".claude/settings.json"), &bytes).unwrap();
    fs::create_dir_all(root.join(".moai-cli/manifests")).unwrap();
    fs::write(
        root.join(".moai-cli/manifests/claude-settings.json"),
        serde_json::to_vec_pretty(&json!({
            "schema_version": 1,
            "settings_path": ".claude/settings.json",
            "owned": {"hooks": Value::Object(previous)},
            "settings_sha256": moai_cli::digest(&bytes),
        }))
        .unwrap(),
    )
    .unwrap();
    settings
}

/// AC-3 (REQ-6, REQ-7): the untouched set of the release before is an `update` that lists the new
/// group and keeps the other key; after `hooks-apply` the next plan is unchanged.
#[test]
fn plan_updates_the_previous_release_set_and_then_is_unchanged() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    install_previous_release_set(root);
    let plan = hooks::plan(root, "claude", None).unwrap();
    assert_eq!(plan["action"], json!("update"), "{plan}");
    assert_eq!(plan["writes"], json!(false));
    let listed = plan["hooks"].to_string();
    assert!(listed.contains("--event security-scan"), "{listed}");
    let applied = hooks::apply(
        root,
        "claude",
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    assert_eq!(applied["state"], json!("applied"), "{applied}");
    let settings = read_json(&root.join(".claude/settings.json"));
    assert_eq!(settings["model"], json!("kept"));
    assert_eq!(
        settings["hooks"],
        Value::Object(hooks::hooks_object("claude").unwrap())
    );
    let again = hooks::plan(root, "claude", None).unwrap();
    assert_eq!(again["action"], json!("unchanged"), "{again}");
}

/// AC-3 (REQ-8): a hand-edited set stays a conflict and the plan leaves its bytes as they were.
#[test]
fn plan_keeps_a_hand_edited_set_as_a_conflict() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    let mut settings = install_previous_release_set(root);
    settings["hooks"]["PostToolUse"][0]["hooks"][0]["command"] = json!("my own receipt tool");
    let mut edited = serde_json::to_vec_pretty(&settings).unwrap();
    edited.push(b'\n');
    fs::write(root.join(".claude/settings.json"), &edited).unwrap();
    let plan = hooks::plan(root, "claude", None).unwrap();
    assert_eq!(plan["action"], json!("conflict"), "{plan}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        edited
    );
}

struct Fx {
    _dir: tempfile::TempDir,
    proj: PathBuf,
    home: PathBuf,
}

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (proj, home) = (root.join("proj"), root.join("home"));
    fs::create_dir_all(proj.join("src")).unwrap();
    fs::create_dir_all(proj.join(".moai-cli/receipts")).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    Fx {
        _dir: dir,
        proj,
        home,
    }
}

impl Fx {
    fn dispatch(&self, event: &str, stdin: Value) -> (i32, String) {
        let mut child = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .current_dir(&self.proj)
            .args(["hook-dispatch", "--host", "claude", "--event", event])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, stdin.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap_or(-1),
            String::from_utf8_lossy(&out.stdout).trim().to_string(),
        )
    }

    fn receipts(&self, file: &str) -> PathBuf {
        self.proj.join(".moai-cli/receipts").join(file)
    }

    fn lines(&self, file: &str) -> Vec<Value> {
        fs::read_to_string(self.receipts(file))
            .unwrap_or_default()
            .lines()
            .filter_map(|l| serde_json::from_str(l).ok())
            .collect()
    }
}

/// AC-4 (REQ-10, REQ-11): with the log of the event at its size limit security-scan writes no
/// line, prints nothing and exits 0, and a deny of the preflight guard is still appended to the
/// engine log.
#[test]
fn a_full_security_scan_log_does_not_stop_the_engine_log() {
    let f = fx();
    let full = fs::File::create(f.receipts("security-scan.jsonl")).unwrap();
    full.set_len(hooks::RECEIPT_LOG_LIMIT).unwrap();
    drop(full);
    let write = json!({"session_id": "s1", "hook_event_name": "PostToolUse", "tool_name": "Write",
        "tool_use_id": "t1", "tool_response": {"success": true}, "cwd": f.proj,
        "tool_input": {"file_path": f.proj.join("src/a.py"), "content": "x = eval(y)\n"}});
    let (code, out) = f.dispatch("security-scan", write);
    assert_eq!((code, out.as_str()), (0, ""));
    assert_eq!(
        fs::metadata(f.receipts("security-scan.jsonl"))
            .unwrap()
            .len(),
        hooks::RECEIPT_LOG_LIMIT,
        "a full log is never appended to"
    );
    let danger = json!({"session_id": "s1", "hook_event_name": "PreToolUse", "tool_name": "Bash",
        "tool_use_id": "t2", "tool_input": {"command": "rm -rf /"}, "cwd": f.proj});
    let (code, out) = f.dispatch("tool-preflight", danger);
    assert_eq!(code, 0, "{out}");
    assert!(out.contains("deny"), "{out}");
    let engine = f.lines("hook-engine.jsonl");
    assert!(
        engine.iter().any(|l| l["kind"] == json!("deny")),
        "the deny line is written while the log of the event is full: {engine:?}"
    );
}

/// AC-4 (REQ-9, REQ-10): an observed finding lands in the log of the event and leaves the engine
/// log without a line (this fails when the line goes to `hook-engine.jsonl`), and the size limit
/// of the log of the event is exact: one byte below it the log still takes a line, at it not.
#[test]
fn an_observed_finding_stays_out_of_the_engine_log_and_the_limit_edge_is_exact() {
    let f = fx();
    let log = f.receipts("security-scan.jsonl");
    let write = |f: &Fx| {
        f.dispatch(
            "security-scan",
            json!({"session_id": "s1", "hook_event_name": "PostToolUse", "tool_name": "Write",
                "tool_use_id": "t1", "tool_response": {"success": true}, "cwd": f.proj,
                "tool_input": {"file_path": f.proj.join("src/a.py"), "content": "x = eval(y)\n"}}),
        )
    };
    assert_eq!(write(&f), (0, String::new()));
    let own = f.lines("security-scan.jsonl");
    assert_eq!(own.len(), 1, "{own:?}");
    assert_eq!(own[0]["kind"], json!("would_notify"));
    assert!(
        f.lines("hook-engine.jsonl").is_empty(),
        "the would_notify line must not reach the engine log"
    );
    let set_len = |n: u64| {
        fs::OpenOptions::new()
            .write(true)
            .open(&log)
            .unwrap()
            .set_len(n)
            .unwrap()
    };
    set_len(hooks::RECEIPT_LOG_LIMIT - 1);
    assert_eq!(write(&f), (0, String::new()));
    assert!(
        fs::metadata(&log).unwrap().len() > hooks::RECEIPT_LOG_LIMIT - 1,
        "one byte below the limit the log still takes a line"
    );
    set_len(hooks::RECEIPT_LOG_LIMIT);
    assert_eq!(write(&f), (0, String::new()));
    assert_eq!(
        fs::metadata(&log).unwrap().len(),
        hooks::RECEIPT_LOG_LIMIT,
        "at the limit it does not"
    );
}
