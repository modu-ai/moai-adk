//! User-only launcher access configuration; all homes and credentials are synthetic.
//! SPEC-20261123-launcher-auto-mode: the default is auto mode; `full` (the bypass) is only the
//! user's own choice.
mod common;
use serde_json::Value;
use std::{
    fs,
    path::Path,
    process::{Command, Output},
};

const HOSTS: [&str; 2] = ["cc", "glm"];
/// An unnamed launch and a launch named with `--name`.
const NAMES: [&[&str]; 2] = [&[], &["--name", "lane-1"]];
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

/// The three access modes a launch can end in.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Mode {
    /// `--permission-mode auto` (the default).
    Auto,
    /// `--dangerously-skip-permissions` (only `launcher.role_access: full`).
    Full,
    /// No permission flag (`--permissions` or `launcher.role_access: prompt`).
    Prompt,
}

/// The mode a launch resolves to: `--permissions` > the user key > auto.
fn mode_of(setting: Option<&str>, cli_permissions: bool) -> Mode {
    if cli_permissions {
        return Mode::Prompt;
    }
    match setting {
        Some("full") => Mode::Full,
        Some("prompt") => Mode::Prompt,
        _ => Mode::Auto,
    }
}

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

/// The skill copies of both hosts, so a launch prints no missing-skill warning: that line names
/// the HOME folder, and the outputs compared below come from different HOMEs
/// (SPEC-20261117-installer-skill-agents REQ-10).
fn with_skill(home: &Path) {
    for dir in [".claude/skills/moai", ".agents/skills/moai"] {
        fs::create_dir_all(home.join(dir)).unwrap();
        fs::write(home.join(dir).join("SKILL.md"), "skill\n").unwrap();
    }
}

fn command(project: &Path, home: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(crate::common::MOAI_BIN);
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

// Literal complete host argv snapshots: order and defaults are part of the contract. Both hosts
// start the same Claude Code argv shape (glm differs by its environment). `named` is 0 for an
// unnamed launch and 1 for `--name lane-1`.
fn expected(_host: &str, named: usize, mode: Mode) -> Vec<&'static str> {
    let mut argv = vec!["claude"];
    if named == 1 {
        argv.extend(["--name", "lane-1"]);
    }
    argv.extend(["--model", "opus", "--effort", "medium"]);
    match mode {
        Mode::Auto => argv.extend(["--permission-mode", "auto"]),
        Mode::Full => argv.push("--dangerously-skip-permissions"),
        Mode::Prompt => {}
    }
    argv
}

#[test]
fn access_matrix_and_cli_precedence_preserve_complete_host_argv() {
    for setting in [None, Some("auto"), Some("full"), Some("prompt")] {
        let home = tempfile::tempdir().unwrap();
        if let Some(value) = setting {
            config(
                home.path(),
                &format!(r#"{{"launcher":{{"role_access":"{value}"}}}}"#),
            );
        }
        let p = project();
        for host in HOSTS {
            for (role, flags) in NAMES.iter().enumerate() {
                for cli in [false, true] {
                    let mut args = vec![host];
                    args.extend_from_slice(flags);
                    if cli {
                        args.push("--permissions");
                    }
                    let v = dry(p.path(), home.path(), &args);
                    assert_eq!(
                        v["argv"],
                        serde_json::json!(expected(host, role, mode_of(setting, cli))),
                        "{args:?} {setting:?}"
                    );
                }
            }
        }
    }
}

/// REQ-1: with no key, no `--permissions` and a HOME without a launcher file, every launch form
/// of both hosts, starts in auto mode and never in the bypass.
#[test]
fn every_launch_form_defaults_to_auto_mode_and_never_to_the_bypass() {
    let p = project();
    let home = tempfile::tempdir().unwrap();
    for host in HOSTS {
        for flags in NAMES {
            let mut args = vec![host];
            args.extend_from_slice(flags);
            let argv = dry(p.path(), home.path(), &args)["argv"].clone();
            let list: Vec<&str> = argv
                .as_array()
                .unwrap()
                .iter()
                .map(|a| a.as_str().unwrap())
                .collect();
            assert!(
                list.windows(2).any(|w| w == ["--permission-mode", "auto"]),
                "{args:?}: {list:?}"
            );
            assert!(
                !list.contains(&"--dangerously-skip-permissions"),
                "{args:?}: {list:?}"
            );
        }
    }
}

#[test]
fn absent_key_retains_byte_identical_output_and_project_config_is_ignored() {
    let p = project();
    // #2481: the baseline runs with a settings file too; without one the dry run reports the
    // first-launch initialization it would do, and the outputs would differ for that alone.
    fs::create_dir_all(p.path().join(".moai-cli")).unwrap();
    fs::write(p.path().join(".moai-cli/project.yaml"), "{}").unwrap();
    let home = tempfile::tempdir().unwrap();
    with_skill(home.path());
    for host in HOSTS {
        for flags in NAMES {
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
                with_skill(alternate.path());
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
    for (user, project_value) in [
        ("full", "prompt"),
        ("prompt", "full"),
        ("auto", "full"),
        ("auto", "prompt"),
    ] {
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
                for (role, flags) in NAMES.iter().enumerate() {
                    let mut args = vec![host];
                    args.extend_from_slice(flags);
                    assert_eq!(
                        dry(p.path(), home.path(), &args)["argv"],
                        serde_json::json!(expected(host, role, mode_of(Some(user), false)))
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
fn invalid_config_is_refused_even_for_unnamed_and_cli_override() {
    for bad in INVALID {
        let p = project();
        let home = tempfile::tempdir().unwrap();
        config(home.path(), bad);
        for host in HOSTS {
            for flags in NAMES {
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

/// REQ-2: the refusal names the three values and never quotes the bad one.
#[test]
fn the_refusal_names_auto_full_and_prompt() {
    let p = project();
    let home = tempfile::tempdir().unwrap();
    config(
        home.path(),
        r#"{"launcher":{"role_access":"SECRET-invalid-access"}}"#,
    );
    let out = command(p.path(), home.path(), &["cc", "--dry-run"])
        .output()
        .unwrap();
    rejected(&out);
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(err.contains("auto, full or prompt"), "{err}");
}

#[test]
fn nonregular_user_config_is_refused() {
    let p = project();
    let home = tempfile::tempdir().unwrap();
    fs::create_dir_all(home.path().join(".moai-cli/config/launcher.yaml")).unwrap();
    for host in HOSTS {
        for args in [vec![host, "--name", "lane-1"], vec![host, "--permissions"]] {
            rejected(
                &command(p.path(), home.path(), &args)
                    .arg("--dry-run")
                    .output()
                    .unwrap(),
            );
        }
    }
}

/// REQ-7 and REQ-8: the notes say what the launch did; the bypass is never called the default,
/// and only an auto-mode glm launch points at the troubleshooting notice.
#[test]
fn notes_name_the_mode_and_only_glm_points_at_the_classifier_notice() {
    let p = project();
    for (host, expect_pointer) in [("cc", false), ("glm", true)] {
        // A fresh HOME per host: the `full` setting written below must not reach the next host.
        let home = tempfile::tempdir().unwrap();
        for flags in NAMES {
            let mut args = vec![host];
            args.extend_from_slice(flags);
            let notes = dry(p.path(), home.path(), &args)["notes"].to_string();
            assert!(
                notes.contains("--permission-mode auto"),
                "{args:?}: {notes}"
            );
            assert!(!notes.contains("dangerously"), "{args:?}: {notes}");
            assert_eq!(
                notes.contains("docs/troubleshooting.md"),
                expect_pointer,
                "{args:?}: {notes}"
            );
        }
        // The bypass chosen by the user is named as the user's own choice; `--permissions` and
        // `prompt` add no pointer either.
        config(home.path(), r#"{"launcher":{"role_access":"full"}}"#);
        let notes = dry(p.path(), home.path(), &[host, "--name", "lane-1"])["notes"].to_string();
        assert!(notes.contains("launcher.role_access"), "{notes}");
        assert!(!notes.contains("--permission-mode"), "{notes}");
        assert!(!notes.contains("docs/troubleshooting.md"), "{notes}");
        let bare = tempfile::tempdir().unwrap();
        let notes = dry(
            p.path(),
            bare.path(),
            &[host, "--name", "lane-1", "--permissions"],
        )["notes"]
            .to_string();
        assert!(!notes.contains("--permission-mode"), "{notes}");
        assert!(!notes.contains("docs/troubleshooting.md"), "{notes}");
    }
}
