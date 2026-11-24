//! SPEC-20261121-model-defaults-one-table, as narrowed by
//! SPEC-20261124-launcher-retire-role-flags: the launcher no longer starts role sessions, so the
//! role rows of the default table are no longer reachable through `moai cc|glm`; they stay the
//! "(default)" of the question screen. A launcher session starts with the defaults of its host
//! (`tests/launcher_retire_role_flags.rs` pins them per host). Every case uses `--dry-run` with
//! a synthetic HOME and a temporary git project; no host is started and no real settings file
//! is read.

#![cfg(unix)]

mod common;

use moai_cli::model_effort_settings::{default_choice, ROLES, TOOLS};
use serde_json::Value;
use std::fs;
use std::process::{Command, Output};

struct Fixture {
    project: tempfile::TempDir,
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
        let home = tempfile::tempdir().unwrap();
        // The skill copies of the host, so a launch prints no missing-skill warning.
        for dir in [".claude/skills/moai", ".agents/skills/moai"] {
            fs::create_dir_all(home.path().join(dir)).unwrap();
            fs::write(home.path().join(dir).join("SKILL.md"), "skill\n").unwrap();
        }
        Fixture { project, home }
    }

    /// Writes the user launcher settings file under the synthetic HOME.
    fn config(&self, text: &str) {
        let dir = moai_cli::home_root::path(self.home.path()).join("config");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("launcher.yaml"), text).unwrap();
    }

    fn run(&self, args: &[&str]) -> Output {
        Command::new(crate::common::MOAI_BIN)
            .args(args)
            .arg("--dry-run")
            .current_dir(self.project.path())
            .env("PATH", "/usr/bin:/bin")
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

    /// The launcher's argv with `--dry-run`; the command must have exited 0.
    fn argv(&self, args: &[&str]) -> Vec<String> {
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

/// The value after `flag`, `None` when the flag is not in the argv.
fn value_of<'a>(argv: &'a [String], flag: &str) -> Option<&'a str> {
    argv.windows(2)
        .find(|w| w[0] == flag)
        .map(|w| w[1].as_str())
}

/// The default table keeps its four rows (claude and glm by two roles): the question screen of
/// `moai update` shows them as "(default)".
#[test]
fn the_default_table_keeps_its_rows_for_the_question_screen() {
    let mut rows = 0;
    for tool in TOOLS {
        for role in ROLES {
            assert!(
                default_choice(tool, role).is_some(),
                "{tool} {role}: the table has no default"
            );
            rows += 1;
        }
    }
    assert_eq!(rows, 4);
}

/// A launch takes no row of the table: with or without a name, and with a settings file that
/// still holds an old role entry, a Claude launch starts with `opus` and `medium`.
#[test]
fn a_launch_takes_no_role_row_of_the_table() {
    let f = Fixture::new();
    for args in [&["cc"][..], &["cc", "--name", "lane-1"][..]] {
        let a = f.argv(args);
        assert_eq!(value_of(&a, "--model"), Some("opus"), "{a:?}");
        assert_eq!(value_of(&a, "--effort"), Some("medium"), "{a:?}");
    }
    f.config(
        r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-sonnet-5-5","effort":"low"}}}}}"#,
    );
    for args in [&["cc"][..], &["cc", "--name", "lane-1"][..]] {
        let a = f.argv(args);
        assert_eq!(value_of(&a, "--model"), Some("opus"), "{a:?}");
        assert_eq!(value_of(&a, "--effort"), Some("medium"), "{a:?}");
    }
}

/// The help of `--model` and `--effort` does not say that a role picks the default.
#[test]
fn the_model_and_effort_help_names_no_role() {
    let f = Fixture::new();
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["cc", "--help"])
        .env("HOME", f.home.path())
        .output()
        .unwrap();
    assert!(out.status.success());
    let help = String::from_utf8_lossy(&out.stdout).to_string();
    let section = |flag: &str| -> String {
        let start = help
            .find(flag)
            .unwrap_or_else(|| panic!("no {flag} in the help"));
        help[start..].lines().take(3).collect::<Vec<_>>().join(" ")
    };
    for flag in ["--model <MODEL>", "--effort <EFFORT>"] {
        let text = section(flag);
        assert!(text.contains("default"), "{flag}: {text}");
        assert!(!text.contains("role"), "{flag}: {text}");
    }
}
