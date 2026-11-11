//! User-only launcher access configuration; all homes and credentials are synthetic.
mod common;
use serde_json::Value;
use std::{
    fs,
    path::Path,
    process::{Command, Output},
};

const HOSTS: [&str; 3] = ["cc", "glm", "codex"];
const ROLES: [&[&str]; 3] = [&["-l"], &["-a"], &[]];
const INVALID: [&str; 17] = [
    r#"{"launcher":{"role_access":"SECRET-invalid-access"}}"#,
    r#"{"launcher":{"role_access":null}}"#,
    r#"{"launcher":{"role_access":7}}"#,
    r#"{"launcher":{"role_access":true}}"#,
    r#"{"launcher":{"role_access":[]}}"#,
    r#"{"launcher":{"role_access":{}}}"#,
    r#"{"launcher":null}"#,
    r#"{"launcher":[]}"#,
    r#"{"launcher":"SECRET-invalid-structure"}"#,
    "null",
    "[]",
    "7",
    "{ SECRET-malformed",
    r#"{"launcher":false}"#,
    r#"{"launcher":7}"#,
    r#""SECRET-invalid-root""#,
    "true",
];

fn project() -> tempfile::TempDir {
    let p = tempfile::tempdir().unwrap();
    assert!(common::git::command()
        .args(["init", "-q"])
        .current_dir(p.path())
        .status()
        .unwrap()
        .success());
    p
}

fn config(home: &Path, text: &str) {
    fs::create_dir_all(home.join(".moai-cli/config")).unwrap();
    fs::write(home.join(".moai-cli/config/launcher.yaml"), text).unwrap();
}

fn command(project: &Path, home: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    c.args(args)
        .current_dir(project)
        .env("HOME", home)
        .env("USERPROFILE", home)
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real");
    c
}

fn dry(project: &Path, home: &Path, args: &[&str]) -> Value {
    let out = command(project, home, args)
        .arg("--dry-run")
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice(&out.stdout).unwrap()
}

// Literal complete host argv snapshots: order and defaults are part of the contract.
fn expected(host: &str, role: usize, prompt: bool) -> Vec<&'static str> {
    match (host, role, prompt) {
        ("cc", 0, false) => vec![
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ],
        ("cc", 1, false) => vec![
            "claude",
            "--name",
            "agent-1",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ],
        ("cc", 0, true) => vec![
            "claude", "--name", "leader", "--model", "opus", "--effort", "medium",
        ],
        ("cc", 1, true) => vec![
            "claude", "--name", "agent-1", "--model", "opus", "--effort", "medium",
        ],
        ("cc", 2, _) => vec!["claude", "--model", "opus", "--effort", "medium"],
        ("glm", 0, false) => vec![
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ],
        ("glm", 1, false) => vec![
            "claude",
            "--name",
            "agent-1",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions",
        ],
        ("glm", 0, true) => vec![
            "claude", "--name", "leader", "--model", "opus", "--effort", "medium",
        ],
        ("glm", 1, true) => vec![
            "claude", "--name", "agent-1", "--model", "opus", "--effort", "medium",
        ],
        ("glm", 2, _) => vec!["claude", "--model", "opus", "--effort", "medium"],
        ("codex", 0, false) => vec![
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: leader",
        ],
        ("codex", 1, false) => vec![
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: agent-1",
        ],
        ("codex", 0, true) => vec![
            "codex",
            "-s",
            "workspace-write",
            "-a",
            "on-request",
            "MoAI session name: leader",
        ],
        ("codex", 1, true) => vec![
            "codex",
            "-s",
            "workspace-write",
            "-a",
            "on-request",
            "MoAI session name: agent-1",
        ],
        ("codex", 2, _) => vec!["codex"],
        _ => unreachable!(),
    }
}

#[test]
fn access_matrix_and_cli_precedence_preserve_complete_host_argv() {
    for setting in [None, Some("full"), Some("prompt")] {
        let home = tempfile::tempdir().unwrap();
        if let Some(value) = setting {
            config(
                home.path(),
                &format!(r#"{{"launcher":{{"role_access":"{value}"}}}}"#),
            );
        }
        let p = project();
        for host in HOSTS {
            for (role, flags) in ROLES.iter().enumerate() {
                for cli in [false, true] {
                    let mut args = vec![host];
                    args.extend_from_slice(flags);
                    if cli {
                        args.push("--permissions");
                    }
                    let v = dry(p.path(), home.path(), &args);
                    assert_eq!(
                        v["argv"],
                        serde_json::json!(expected(host, role, cli || setting == Some("prompt"))),
                        "{args:?} {setting:?}"
                    );
                }
            }
        }
    }
}

#[test]
fn absent_key_retains_byte_identical_output_and_project_config_is_ignored() {
    let p = project();
    let home = tempfile::tempdir().unwrap();
    for host in HOSTS {
        for flags in ROLES {
            let mut args = vec![host];
            args.extend_from_slice(flags);
            let baseline = command(p.path(), home.path(), &args)
                .arg("--dry-run")
                .output()
                .unwrap();
            assert!(baseline.status.success());
            for text in [
                "{}",
                r#"{"launcher":{}}"#,
                r#"{"unrelated":"SECRET-ignored"}"#,
            ] {
                let alternate = tempfile::tempdir().unwrap();
                config(alternate.path(), text);
                fs::create_dir_all(p.path().join(".moai-cli")).unwrap();
                fs::write(
                    p.path().join(".moai-cli/project.yaml"),
                    r#"{"launcher":{"role_access":"prompt"}}"#,
                )
                .unwrap();
                let out = command(p.path(), alternate.path(), &args)
                    .arg("--dry-run")
                    .output()
                    .unwrap();
                assert!(out.status.success());
                assert_eq!(out.stdout, baseline.stdout);
                assert_eq!(out.stderr, baseline.stderr);
            }
        }
    }
}

#[test]
fn project_setting_cannot_override_either_user_setting() {
    for (user, project_value) in [("full", "prompt"), ("prompt", "full")] {
        let home = tempfile::tempdir().unwrap();
        config(
            home.path(),
            &format!(r#"{{"launcher":{{"role_access":"{user}"}}}}"#),
        );
        for claude_settings in [
            None,
            Some(r#"{"permissions":{"allow":["Bash(cargo test *)"]}}"#),
            Some("{ not json"),
        ] {
            let p = project();
            fs::create_dir_all(p.path().join(".moai-cli")).unwrap();
            fs::write(
                p.path().join(".moai-cli/project.yaml"),
                format!(r#"{{"launcher":{{"role_access":"{project_value}"}}}}"#),
            )
            .unwrap();
            if let Some(text) = claude_settings {
                fs::create_dir(p.path().join(".claude")).unwrap();
                fs::write(p.path().join(".claude/settings.json"), text).unwrap();
            }
            for host in HOSTS {
                for (role, flags) in ROLES.iter().enumerate() {
                    let mut args = vec![host];
                    args.extend_from_slice(flags);
                    assert_eq!(
                        dry(p.path(), home.path(), &args)["argv"],
                        serde_json::json!(expected(host, role, user == "prompt"))
                    );
                }
            }
        }
    }
}

fn rejected(out: &Output) {
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
    assert!(!String::from_utf8_lossy(&out.stderr).contains("SECRET"));
}

#[test]
fn invalid_config_is_refused_even_for_roleless_and_cli_override() {
    for bad in INVALID {
        let p = project();
        let home = tempfile::tempdir().unwrap();
        config(home.path(), bad);
        for host in HOSTS {
            for flags in ROLES {
                for cli in [false, true] {
                    let mut args = vec![host];
                    args.extend_from_slice(flags);
                    if cli {
                        args.push("--permissions");
                    }
                    rejected(
                        &command(p.path(), home.path(), &args)
                            .arg("--dry-run")
                            .output()
                            .unwrap(),
                    );
                }
            }
        }
    }
}

#[test]
fn nonregular_user_config_is_refused() {
    let p = project();
    let home = tempfile::tempdir().unwrap();
    fs::create_dir_all(home.path().join(".moai-cli/config/launcher.yaml")).unwrap();
    for host in HOSTS {
        for args in [vec![host, "-a"], vec![host, "--permissions"]] {
            rejected(
                &command(p.path(), home.path(), &args)
                    .arg("--dry-run")
                    .output()
                    .unwrap(),
            );
        }
    }
}

#[test]
fn legacy_top_level_roles_use_the_same_access_resolution() {
    let p = project();
    let home = tempfile::tempdir().unwrap();
    for setting in ["full", "prompt"] {
        config(
            home.path(),
            &format!(r#"{{"launcher":{{"role_access":"{setting}"}}}}"#),
        );
        assert_eq!(
            dry(p.path(), home.path(), &["-l"])["argv"],
            serde_json::json!(expected("cc", 0, setting == "prompt"))
        );
        for host in HOSTS {
            assert_eq!(
                dry(p.path(), home.path(), &["-a", host])["argv"],
                serde_json::json!(expected(host, 1, setting == "prompt"))
            );
        }
    }
}
