//! #2359 (user decision 2026-11-10, comment 6760976883): a role session (`-l`, `-a`) starts
//! with full access on every host — `--dangerously-skip-permissions` on `cc`/`glm` whatever the
//! project's `permissions` key says (this replaces #2244), `-s danger-full-access -a never` on
//! `codex` — and `--permissions` keeps the prompts. A roleless session gets no permission flag
//! at all, so the host's own settings choose its mode. Every case is `--dry-run` in a temporary
//! git project (the `-a` name preview reads that project's counter); no host is started.
mod common;
use serde_json::Value;
use std::{fs, path::Path, process::Command};

const SKIP: &str = "--dangerously-skip-permissions";
const CODEX_FULL: [&str; 4] = ["-s", "danger-full-access", "-a", "never"];
const CODEX_ASK: [&str; 4] = ["-s", "workspace-write", "-a", "on-request"];

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
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(p.path())
        .status()
        .unwrap();
    assert!(init.success());
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
const HOSTS: [&str; 3] = ["cc", "glm", "codex"];
const ROLES: [&[&str]; 3] = [&["-l"], &["-a"], &["-a", "--name", "agent-7"]];

/// Any permission-related flag the launcher could add, on either host family.
fn has_permission_flag(argv: &[String]) -> bool {
    argv.iter()
        .any(|a| a == SKIP || a == "-s" || a == "--sandbox" || a == "--ask-for-approval")
        // `-a` is codex's approval flag; on the host argv (not the launcher's) it only
        // appears in the codex pair.
        || (argv.first().map(String::as_str) == Some("codex") && argv.iter().any(|a| a == "-a"))
}
fn full_access(host: &str, argv: &[String]) -> bool {
    if host == "codex" {
        argv.len() >= 5 && argv[1..5] == CODEX_FULL
    } else {
        argv.iter().any(|a| a == SKIP)
    }
}
fn asks(host: &str, argv: &[String]) -> bool {
    if host == "codex" {
        argv.len() >= 5 && argv[1..5] == CODEX_ASK
    } else {
        !argv.iter().any(|a| a == SKIP)
    }
}

#[test]
fn a_role_session_has_full_access_whatever_the_project_settings() {
    for settings in SETTINGS {
        let p = project(settings);
        for host in HOSTS {
            for role in ROLES {
                let args: Vec<&str> = std::iter::once(host).chain(role.iter().copied()).collect();
                let (code, argv, notes, stderr) = dry(p.path(), &args);
                assert_eq!(code, 0, "{args:?} {settings:?}: {stderr}");
                assert!(full_access(host, &argv), "{args:?} {settings:?}: {argv:?}");
                assert!(notes.contains("role session"), "{args:?}: {notes}");
                // The project file is not consulted, so nothing reports on it.
                assert!(!stderr.contains("permissions key"), "{args:?}: {stderr}");
                assert!(!notes.contains("permissions key"), "{args:?}: {notes}");
            }
        }
    }
}

#[test]
fn a_roleless_session_gets_no_permission_flag_at_all() {
    for settings in SETTINGS {
        let p = project(settings);
        for host in HOSTS {
            for extra in [&[][..], &["--permissions"][..], &["--name", "mine"][..]] {
                let args: Vec<&str> = std::iter::once(host).chain(extra.iter().copied()).collect();
                let (code, argv, notes, stderr) = dry(p.path(), &args);
                assert_eq!(code, 0, "{args:?}: {stderr}");
                assert!(
                    !has_permission_flag(&argv),
                    "{args:?} {settings:?}: {argv:?}"
                );
                assert!(
                    notes.contains("no permission flag")
                        || notes.contains("no sandbox or approval flag"),
                    "{args:?}: {notes}"
                );
            }
        }
        // Byte-exact for codex: nothing between the binary and the prompt.
        let (_, argv, _, _) = dry(p.path(), &["codex"]);
        assert_eq!(argv, ["codex"]);
        let (_, argv, _, _) = dry(p.path(), &["codex", "review the diff"]);
        assert_eq!(argv, ["codex", "review the diff"]);
    }
}

#[test]
fn the_permissions_flag_keeps_the_prompts_on_a_role_session() {
    for settings in [None, SETTINGS[1]] {
        let p = project(settings);
        for host in HOSTS {
            for role in ROLES {
                let args: Vec<&str> = std::iter::once(host)
                    .chain(role.iter().copied())
                    .chain(["--permissions"])
                    .collect();
                let (code, argv, notes, stderr) = dry(p.path(), &args);
                assert_eq!(code, 0, "{args:?}: {stderr}");
                assert!(asks(host, &argv), "{args:?} {settings:?}: {argv:?}");
                assert!(
                    !notes.contains("role session, user decision"),
                    "{args:?}: {notes}"
                );
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
    for word in HOSTS {
        let (code, argv, _, stderr) = dry(p.path(), &["-a", word]);
        assert_eq!(code, 0, "{word}: {stderr}");
        assert!(full_access(word, &argv), "{word}: {argv:?}");
        let (_, argv, _, _) = dry(p.path(), &["-a", word, "--permissions"]);
        assert!(asks(word, &argv), "{word}: {argv:?}");
    }
}
