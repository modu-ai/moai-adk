//! #2359 (user decision 2026-11-10): every session the launcher opens — roleless, `-l` and
//! `-a`, on `cc`, `glm` and `codex` — starts with full access and no permission prompt, and
//! `--permissions` keeps the prompts. This replaces #2053 (no skip without a role) and #2244
//! (a project `permissions` key drops the skip): the project's `.claude/settings.json` no
//! longer changes the argv. Every case is `--dry-run` in a temporary project; no host is
//! started and no real settings file is read or written.
use serde_json::Value;
use std::{fs, path::Path, process::Command};

const SKIP: &str = "--dangerously-skip-permissions";
const CODEX_FULL: [&str; 5] = ["codex", "-s", "danger-full-access", "-a", "never"];
const CODEX_ASK: [&str; 5] = ["codex", "-s", "workspace-write", "-a", "on-request"];

fn dry(dir: &Path, args: &[&str]) -> (i32, Vec<String>, String, String) {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .arg("--dry-run")
        .current_dir(dir)
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        // A dummy token so `glm` composes; no host is started in this file.
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    let argv: Vec<String> = serde_json::from_value(v["argv"].clone()).unwrap_or_default();
    (
        out.status.code().unwrap_or(-1),
        argv,
        v["notes"].to_string(),
        String::from_utf8_lossy(&out.stderr).to_string(),
    )
}

fn project(settings: Option<&str>) -> tempfile::TempDir {
    let p = tempfile::tempdir().unwrap();
    if let Some(text) = settings {
        fs::create_dir(p.path().join(".claude")).unwrap();
        fs::write(p.path().join(".claude/settings.json"), text).unwrap();
    }
    p
}

/// No file, a file with a `permissions` key (the #2244 case) and an unreadable file.
const SETTINGS: [Option<&str>; 3] = [
    None,
    Some(r#"{"permissions":{"allow":["Bash(cargo test *)"]}}"#),
    Some("{ not json"),
];
/// Roleless, leader and agent, in the host-word form and the old top-level form.
const FORMS: [&[&str]; 4] = [&[], &["-l"], &["-a"], &["-a", "--name", "agent-7"]];

fn full_access(host: &str, argv: &[String]) -> bool {
    if host == "codex" {
        argv.len() >= 5 && argv[..5] == CODEX_FULL
    } else {
        argv.iter().any(|a| a == SKIP)
    }
}
fn asks(host: &str, argv: &[String]) -> bool {
    if host == "codex" {
        argv.len() >= 5 && argv[..5] == CODEX_ASK && !argv.iter().any(|a| a == SKIP)
    } else {
        !argv.iter().any(|a| a == SKIP)
    }
}

#[test]
fn every_launched_session_has_full_access_whatever_the_project_settings() {
    for settings in SETTINGS {
        let p = project(settings);
        for host in ["cc", "glm", "codex"] {
            for form in FORMS {
                let args: Vec<&str> = std::iter::once(host).chain(form.iter().copied()).collect();
                let (code, argv, notes, stderr) = dry(p.path(), &args);
                assert_eq!(code, 0, "{args:?} {settings:?}: {stderr}");
                assert!(full_access(host, &argv), "{args:?} {settings:?}: {argv:?}");
                assert!(
                    notes.contains("full-access default"),
                    "{args:?}: the note names the default: {notes}"
                );
                // The project file is not consulted any more, so nothing reports on it.
                assert!(!stderr.contains("permissions key"), "{args:?}: {stderr}");
                assert!(!notes.contains("permissions key"), "{args:?}: {notes}");
                assert!(!notes.contains("unreadable"), "{args:?}: {notes}");
            }
        }
    }
}

#[test]
fn the_permissions_flag_keeps_the_prompts_on_every_host_and_form() {
    for settings in [None, SETTINGS[1]] {
        let p = project(settings);
        for host in ["cc", "glm", "codex"] {
            for form in FORMS {
                let args: Vec<&str> = std::iter::once(host)
                    .chain(form.iter().copied())
                    .chain(["--permissions"])
                    .collect();
                let (code, argv, notes, stderr) = dry(p.path(), &args);
                assert_eq!(code, 0, "{args:?}: {stderr}");
                assert!(asks(host, &argv), "{args:?} {settings:?}: {argv:?}");
                assert!(!notes.contains("full-access default"), "{args:?}: {notes}");
            }
        }
    }
}

#[test]
fn the_old_top_level_role_form_has_full_access_too() {
    let p = project(None);
    let (code, argv, _, stderr) = dry(p.path(), &["-l"]);
    assert_eq!(code, 0, "{stderr}");
    assert!(full_access("cc", &argv), "{argv:?}");
    for (host, word) in [("cc", "cc"), ("glm", "glm"), ("codex", "codex")] {
        let (code, argv, _, stderr) = dry(p.path(), &["-a", word]);
        assert_eq!(code, 0, "{word}: {stderr}");
        assert!(full_access(host, &argv), "{word}: {argv:?}");
        let (_, argv, _, _) = dry(p.path(), &["-a", word, "--permissions"]);
        assert!(asks(host, &argv), "{word}: {argv:?}");
    }
}
