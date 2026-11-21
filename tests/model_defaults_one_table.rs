//! SPEC-20261121-model-defaults-one-table: a session with a role that has no saved model setting
//! starts with the default of the one table in `model_effort_settings`, the same value the
//! question screen of `moai update` shows as "(default)". Every case uses `--dry-run` with a
//! synthetic HOME that holds no launcher settings file and a temporary git project; no host is
//! started and no real settings file is read.

#![cfg(unix)]

mod common;

use moai_cli::model_effort_settings::{default_choice, Choice};
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::process::{Command, Output};

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Fixture {
    fn new() -> Self {
        let project = tempfile::tempdir().unwrap();
        assert!(common::git::command()
            .args(["init", "-q"])
            .current_dir(project.path())
            .status()
            .unwrap()
            .success());
        let bin = tempfile::tempdir().unwrap();
        let stub = bin.path().join("omp");
        fs::write(&stub, "#!/bin/sh\nexit 0\n").unwrap();
        fs::set_permissions(&stub, fs::Permissions::from_mode(0o755)).unwrap();
        let home = tempfile::tempdir().unwrap();
        for dir in [".claude/skills/moai", ".agents/skills/moai"] {
            fs::create_dir_all(home.path().join(dir)).unwrap();
            fs::write(home.path().join(dir).join("SKILL.md"), "skill\n").unwrap();
        }
        Fixture { project, bin, home }
    }

    fn config(&self, text: &str) {
        let dir = self.home.path().join(".moai-cli/config");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("launcher.yaml"), text).unwrap();
    }

    fn run(&self, args: &[&str]) -> Output {
        Command::new(common::MOAI_BIN)
            .args(args)
            .arg("--dry-run")
            .current_dir(self.project.path())
            .env(
                "PATH",
                format!("{}:/usr/bin:/bin", self.bin.path().display()),
            )
            .env("HOME", self.home.path())
            .env("USERPROFILE", self.home.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env_remove("ZAI_API_KEY")
            .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
            .output()
            .unwrap()
    }

    fn dry(&self, args: &[&str]) -> Vec<String> {
        let out = self.run(args);
        assert!(
            out.status.success(),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        serde_json::from_value(v["argv"].clone()).unwrap()
    }
}

/// True when `flag` is followed directly by `value` in the argv.
fn has_pair(argv: &[String], flag: &str, value: &str) -> bool {
    argv.windows(2).any(|w| w[0] == flag && w[1] == value)
}

fn effort_arg(e: &str) -> String {
    format!("model_reasoning_effort=\"{e}\"")
}

/// The launcher word, the table key of its tool, the model flag and the effort flag of its host.
const HOSTS: [(&str, &str, &str, &str); 4] = [
    ("cc", "claude", "--model", "--effort"),
    ("codex", "codex", "-m", "-c"),
    ("omp", "omp", "--model", "--thinking"),
    ("glm", "glm", "--model", "--effort"),
];

/// The model the host takes: the glm host is Claude Code on z.ai and takes the bare name.
fn host_model(word: &str, choice: &Choice) -> String {
    if word == "glm" {
        choice
            .model
            .strip_prefix("zai/")
            .unwrap_or(&choice.model)
            .to_string()
    } else {
        choice.model.clone()
    }
}

fn host_effort(word: &str, effort: &str) -> String {
    if word == "codex" {
        effort_arg(effort)
    } else {
        effort.to_string()
    }
}

/// AC-2 (REQ-3): nothing saved, a role given: the argv carries the table default of the tool and
/// role in each host's own flags, for every host and both roles.
#[test]
fn an_unsaved_role_session_starts_with_the_table_default() {
    let f = Fixture::new();
    for (word, tool, model_flag, effort_flag) in HOSTS {
        for (role_flag, role) in [("-l", "leader"), ("-a", "agent")] {
            let choice = default_choice(tool, role).expect("the table has a row");
            let a = f.dry(&[word, role_flag]);
            assert!(
                has_pair(&a, model_flag, &host_model(word, &choice)),
                "{word} {role}: {a:?}"
            );
            assert!(
                has_pair(&a, effort_flag, &host_effort(word, &choice.effort)),
                "{word} {role}: {a:?}"
            );
        }
    }
}

/// AC-2 (REQ-3): a launcher settings file that exists but saves nothing for the tool and role
/// gives the same default (the file is not the only way to reach it).
#[test]
fn a_settings_file_without_models_gives_the_same_default() {
    let f = Fixture::new();
    f.config(r#"{"launcher":{"future_key":"kept"}}"#);
    let a = f.dry(&["cc", "-a"]);
    assert!(has_pair(&a, "--model", "claude-sonnet-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "high"), "{a:?}");
}

/// AC-2 (REQ-4): a flag the person gives replaces only its field; the other field is the default.
#[test]
fn a_flag_replaces_only_its_own_field() {
    let f = Fixture::new();
    let a = f.dry(&["cc", "-l", "--model", "claude-haiku-5-5"]);
    assert!(has_pair(&a, "--model", "claude-haiku-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "high"), "{a:?}");
    let a = f.dry(&["codex", "-a", "--effort", "low"]);
    assert!(has_pair(&a, "-m", "gpt-6.1-terra"), "{a:?}");
    assert!(has_pair(&a, "-c", &effort_arg("low")), "{a:?}");
}

/// AC-2 (REQ-6): a saved choice is used before the default.
#[test]
fn a_saved_choice_wins_over_the_default() {
    let f = Fixture::new();
    f.config(
        r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-sonnet-5-5","effort":"low"}}}}}"#,
    );
    let a = f.dry(&["cc", "-l"]);
    assert!(has_pair(&a, "--model", "claude-sonnet-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "low"), "{a:?}");
    // The agent of the same tool has nothing saved: its default.
    let a = f.dry(&["cc", "-a"]);
    assert!(has_pair(&a, "--model", "claude-sonnet-5-5"), "{a:?}");
    assert!(has_pair(&a, "--effort", "high"), "{a:?}");
}

/// AC-3 (REQ-5): a session without a role keeps the launcher's own Claude defaults.
#[test]
fn a_roleless_session_keeps_opus_and_medium() {
    let f = Fixture::new();
    let a = f.dry(&["cc"]);
    assert!(has_pair(&a, "--model", "opus"), "{a:?}");
    assert!(has_pair(&a, "--effort", "medium"), "{a:?}");
}
