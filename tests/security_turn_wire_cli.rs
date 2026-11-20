//! SPEC-20261121-security-turn-wire: the generated Claude Code hook set ends with one `Stop` group
//! that runs `hook-dispatch --event security-turn`, the hosts codex and omp get no such group (the
//! Codex lane Stop group stays the only Stop group of a Codex set), the legacy set has none,
//! `hooks-plan` brings a project that holds the untouched set of the release before this change to
//! the new one as an `update` (and never overwrites a hand-edited one), the capabilities answer
//! lists the event, the two documents name both groups, and the generated command runs the scan.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use moai_cli::{hook_engine, hooks};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::{Command, Stdio};

const PERMISSION: &str = "test-permission-turn-wire";
const TURN_COMMAND: &str = "moai hook-dispatch --host claude --event security-turn";

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

fn doc(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs")
            .join(name),
    )
    .unwrap()
}

/// AC-1 (REQ-1, REQ-2, REQ-3, REQ-5): the Stop event comes last with one group that has no matcher,
/// the exact command and the hook timeout; the other events keep their groups.
#[test]
fn claude_ends_the_hooks_object_with_one_stop_group() {
    let obj = hooks::hooks_object("claude").unwrap();
    let events: Vec<&String> = obj.keys().collect();
    assert_eq!(
        events,
        ["SessionStart", "PreToolUse", "PostToolUse", "Stop"]
    );
    let stop = obj["Stop"].as_array().unwrap();
    assert_eq!(stop.len(), 1, "{stop:?}");
    assert!(
        stop[0].get("matcher").is_none(),
        "a Stop group has no matcher: {:?}",
        stop[0]
    );
    let handler = &stop[0]["hooks"][0];
    assert_eq!(handler["type"], json!("command"));
    assert_eq!(handler["command"], json!(TURN_COMMAND));
    assert_eq!(handler["timeout"], json!(hooks::HOOK_TIMEOUT_SECS));
    assert_eq!(stop[0]["hooks"].as_array().unwrap().len(), 1);
    let post = commands(&obj, "PostToolUse");
    assert_eq!(
        post,
        [
            "moai hook-receipt --host claude",
            "moai hook-dispatch --host claude --event security-scan"
        ]
    );
    assert_eq!(obj["PreToolUse"].as_array().unwrap().len(), 2);
    assert_eq!(obj["SessionStart"].as_array().unwrap().len(), 1);
}

/// AC-1 (REQ-3), through the binary with `[hooks.edit_diagnostics]` enabled: the edit-diagnostics
/// group joins PostToolUse and the Stop event is still the last one of the written settings.
#[test]
fn claude_keeps_the_stop_group_last_while_edit_diagnostics_is_enabled() {
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
    // `hooks-apply` writes only where the first `moai` on PATH is the Rust build: this one.
    let bin = root.join("bin");
    fs::create_dir_all(&bin).unwrap();
    std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
    let path = format!(
        "{}:{}",
        bin.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    let run = |args: &[&str]| -> Value {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", &path)
            .env("HOME", &home)
            .current_dir(&proj)
            .args(args)
            .output()
            .unwrap();
        assert!(out.status.success(), "{args:?}: {out:?}");
        serde_json::from_slice(&out.stdout).unwrap()
    };
    let project = proj.to_str().unwrap();
    let plan = run(&["hooks-plan", project, "--host", "claude"]);
    let digest = plan["result"]["plan_digest"].as_str().unwrap().to_string();
    let applied = run(&[
        "hooks-apply",
        project,
        "--host",
        "claude",
        "--plan-digest",
        &digest,
        "--permission-reference",
        PERMISSION,
    ]);
    assert_eq!(applied["result"]["state"], json!("applied"), "{applied}");
    // The written file is the order a host reads: the edit-diagnostics group joins PostToolUse
    // and the Stop event is still the last one.
    let text = fs::read_to_string(proj.join(".claude/settings.json")).unwrap();
    let settings: Value = serde_json::from_str(&text).unwrap();
    let object = settings["hooks"].as_object().unwrap();
    let events: Vec<&String> = object.keys().collect();
    assert_eq!(
        events,
        ["SessionStart", "PreToolUse", "PostToolUse", "Stop"]
    );
    let post = commands(object, "PostToolUse");
    assert_eq!(post.len(), 3, "{post:?}");
    assert!(post[2].contains("--event edit-diagnostics"), "{post:?}");
    let stop = object["Stop"].as_array().unwrap();
    assert_eq!(stop.len(), 1, "{stop:?}");
    assert_eq!(stop[0]["hooks"][0]["command"], json!(TURN_COMMAND));
}

/// AC-2 (REQ-4, REQ-5): codex and omp run no security-turn group and keep their other groups; the
/// Codex set planned with the lane loop holds one Stop group, the lane's.
#[test]
fn other_hosts_get_no_security_turn_group() {
    for host in ["codex", "omp"] {
        let obj = hooks::hooks_object(host).unwrap();
        for event in obj.keys() {
            for command in commands(&obj, event) {
                assert!(
                    !command.contains("security-turn"),
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
        assert!(obj.get("Stop").is_none(), "{host}");
    }
    let lane = hooks::current_hooks("codex", Some(("lane-1", "o/r"))).unwrap();
    let stop = lane["Stop"].as_array().unwrap();
    assert_eq!(stop.len(), 1, "{stop:?}");
    let command = stop[0]["hooks"][0]["command"].as_str().unwrap();
    assert!(command.contains("lane-stop-hook"), "{command}");
    assert!(!command.contains("security-turn"), "{command}");
}

/// AC-4 (REQ-9): the legacy set (the old binary name) holds no Stop group, so the files of the
/// releases before the rename are still recognised; the same binary name in the current set does.
#[test]
fn legacy_set_holds_no_stop_group() {
    let legacy = hooks::legacy_hooks("claude", None).unwrap();
    assert!(legacy.get("Stop").is_none(), "{legacy:?}");
    for event in legacy.keys() {
        for command in commands(&legacy, event) {
            assert!(command.starts_with("moai-cli "), "{command}");
            assert!(!command.contains("security-turn"), "{event}: {command}");
        }
    }
    let current = hooks::current_hooks("claude", None).unwrap();
    assert_eq!(current["Stop"].as_array().unwrap().len(), 1);
}

/// Writes the settings and the ownership receipt of a project as if the release before this
/// change had applied its set: the current hooks object without the Stop group, next to one key
/// MoAI does not own.
fn install_previous_release_set(root: &Path) -> Value {
    let mut previous = hooks::hooks_object("claude").unwrap();
    assert!(previous.shift_remove("Stop").is_some());
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
    assert!(listed.contains("--event security-turn"), "{listed}");
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

/// AC-5 (REQ-10, REQ-11): the capabilities answer lists the event for claude with the host event
/// Stop and the default mode observe, lists it as unsupported for codex and omp, and no longer says
/// that security-scan is not generated into settings.
#[test]
fn capabilities_list_the_event_for_claude_and_not_for_the_others() {
    let claude = hook_engine::capabilities("claude");
    let turn = &claude["security-turn"];
    assert_eq!(turn["event"], json!("Stop"), "{turn}");
    assert_eq!(turn["default_mode"], json!("observe"), "{turn}");
    assert_eq!(turn["mode_key"], json!("security_turn"), "{turn}");
    assert_eq!(turn["blocks"], json!(false), "{turn}");
    assert_eq!(turn["answer"], json!("systemMessage"), "{turn}");
    for (name, entry) in [
        ("security-turn", turn),
        ("security-scan", &claude["security-scan"]),
    ] {
        let generated = entry["generated"].as_str().unwrap_or("");
        assert!(
            !generated.is_empty() && !generated.contains("not wired"),
            "{name}: {generated}"
        );
    }
    for host in ["codex", "omp"] {
        let answer = hook_engine::capabilities(host);
        assert_eq!(
            answer["security-turn"]["supported"],
            json!(false),
            "{host}: {answer}"
        );
        assert!(
            answer["security-turn"]["reason"].as_str().is_some(),
            "{host}"
        );
    }
    let dir = tempfile::TempDir::new().unwrap();
    let plan = hooks::plan(dir.path(), "claude", None).unwrap();
    assert_eq!(plan["engine_capabilities"], claude);
}

/// AC-6 (REQ-12): both documents name the security-scan group, the Stop group and the command of
/// each.
#[test]
fn docs_name_both_groups_of_the_claude_set() {
    for name in ["hooks-generation.md", "claude-adapter-contract.md"] {
        let text = doc(name);
        for needle in ["--event security-scan", "--event security-turn", "Stop"] {
            assert!(text.contains(needle), "{name} lacks `{needle}`");
        }
    }
}

/// AC-7: the command the generated Stop group holds, run as a host would run it, notes the
/// finding in `security-scan.jsonl` in the default observe mode and prints nothing, and in enforce
/// mode answers with a `systemMessage`.
#[test]
fn end_to_end_the_generated_stop_command_runs_the_scan() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (proj, home) = (root.join("proj"), root.join("home"));
    fs::create_dir_all(proj.join(".moai-cli")).unwrap();
    fs::create_dir_all(proj.join("src")).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    let git = |args: &[&str]| {
        let out = Command::new("git")
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
            .args([
                "-c",
                "commit.gpgsign=false",
                "-c",
                "init.defaultBranch=main",
            ])
            .args(args)
            .current_dir(&proj)
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}: {out:?}");
    };
    fs::write(proj.join("src/a.py"), "x = 1\n").unwrap();
    git(&["init", "-q"]);
    git(&["add", "-A"]);
    git(&["commit", "-q", "-m", "base"]);
    fs::write(proj.join("src/new.py"), "y = eval(z)\n").unwrap();

    let object = hooks::hooks_object("claude").unwrap();
    let command = commands(&object, "Stop").remove(0);
    let args: Vec<&str> = command.split(' ').skip(1).collect();
    assert_eq!(args[0], "hook-dispatch", "{command}");
    let run = |session: &str| {
        let mut child = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .current_dir(&proj)
            .args(&args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        let input = json!({"session_id": session, "hook_event_name": "Stop",
                           "stop_hook_active": false, "cwd": proj});
        common::feed_stdin(&mut child, input.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap_or(-1),
            String::from_utf8_lossy(&out.stdout).trim().to_string(),
            String::from_utf8_lossy(&out.stderr).to_string(),
        )
    };
    let lines = || -> Vec<Value> {
        fs::read_to_string(proj.join(".moai-cli/receipts/security-scan.jsonl"))
            .unwrap_or_default()
            .lines()
            .filter_map(|l| serde_json::from_str(l).ok())
            .collect()
    };

    let (code, out, err) = run("s1");
    assert_eq!((code, out.as_str(), err.as_str()), (0, "", ""));
    let seen = lines();
    assert_eq!(seen.len(), 1, "{seen:?}");
    assert_eq!(seen[0]["kind"], json!("would_notify"));
    assert_eq!(seen[0]["rule"], json!("security_turn"));

    fs::write(
        home.join(".moai-cli/hook-config.toml"),
        "[hooks.modes]\nsecurity_turn = \"enforce\"\n",
    )
    .unwrap();
    let (code, out, err) = run("s2");
    assert_eq!((code, err.as_str()), (0, ""), "{out}");
    let answer: Value = serde_json::from_str(&out).unwrap();
    assert!(answer["systemMessage"].as_str().is_some(), "{out}");
    assert_eq!(answer.as_object().unwrap().len(), 1, "{out}");
}
