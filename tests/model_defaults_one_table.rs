//! SPEC-20261121-model-defaults-one-table: a launcher session with a role and nothing saved for its
//! tool and role starts with the model and effort of the default table, the table the question
//! screen shows as "(default)". Every case uses `--dry-run` with a synthetic HOME and a temporary
//! git project; no host is started and no real settings file is read.

#![cfg(unix)]

mod common;

use moai_cli::model_effort_settings::{default_choice, ROLES, TOOLS};
use serde_json::Value;
use std::fs;
use std::process::{Command, Output};

/// The launcher word and the role flag of a tool and role.
fn launcher_args(tool: &str, role: &str) -> [&'static str; 2] {
    let word = match tool {
        "claude" => "cc",
        "glm" => "glm",
        other => panic!("unexpected tool {other}"),
    };
    let flag = match role {
        "leader" => "-l",
        "agent" => "-a",
        other => panic!("unexpected role {other}"),
    };
    [word, flag]
}

/// The model the host gets: the glm host takes the id without the `zai/` provider prefix.
fn host_model(tool: &str, model: &str) -> String {
    if tool == "glm" {
        model.strip_prefix("zai/").unwrap_or(model).to_string()
    } else {
        model.to_string()
    }
}

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
        let dir = self.home.path().join(".moai-cli/config");
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

/// Every tool and role of the table, with the table's model (as the host takes it) and effort.
fn table() -> Vec<(&'static str, &'static str, String, String)> {
    let mut rows = Vec::new();
    for tool in TOOLS {
        for role in ROLES {
            let choice = default_choice(tool, role)
                .unwrap_or_else(|| panic!("{tool} {role}: the table has no default"));
            rows.push((tool, role, host_model(tool, &choice.model), choice.effort));
        }
    }
    rows
}

/// AC-2 (REQ-2, REQ-3, REQ-4): with no settings file at all, every role session of both hosts starts
/// with the table's model and effort: the launcher's unsaved default equals the table.
#[test]
fn without_a_settings_file_a_role_session_starts_with_the_table_default() {
    let rows = table();
    assert_eq!(rows.len(), 4, "the table is claude and glm by two roles");
    for (tool, role, model, effort) in rows {
        let f = Fixture::new();
        let a = f.argv(&launcher_args(tool, role));
        assert_eq!(
            value_of(&a, "--model"),
            Some(model.as_str()),
            "{tool} {role}: {a:?}"
        );
        assert_eq!(
            value_of(&a, "--effort"),
            Some(effort.as_str()),
            "{tool} {role}: {a:?}"
        );
    }
}

/// AC-2 (REQ-2, REQ-4): a settings file that saves nothing for the tool and role gives the same
/// launch, so the table is the fallback of the saved value and not of the file.
#[test]
fn a_settings_file_without_models_gives_the_table_default() {
    for (tool, role, model, effort) in table() {
        let f = Fixture::new();
        f.config(r#"{"launcher":{"future_key":"SENTINEL-SAVED-TEXT"}}"#);
        let a = f.argv(&launcher_args(tool, role));
        assert_eq!(
            value_of(&a, "--model"),
            Some(model.as_str()),
            "{tool} {role}: {a:?}"
        );
        assert_eq!(
            value_of(&a, "--effort"),
            Some(effort.as_str()),
            "{tool} {role}: {a:?}"
        );
    }
}

/// AC-3 (REQ-6): a model given on the command line stays and only the effort comes from the table;
/// an effort given on the command line stays and only the model comes from the table.
#[test]
fn a_flag_given_keeps_that_field_and_the_table_fills_the_other() {
    for (tool, role, model, effort) in table() {
        let f = Fixture::new();
        let [word, flag] = launcher_args(tool, role);
        let a = f.argv(&[word, flag, "--model", "claude-haiku-5-5"]);
        assert_eq!(
            value_of(&a, "--model"),
            Some("claude-haiku-5-5"),
            "{tool} {role}: {a:?}"
        );
        assert_eq!(
            value_of(&a, "--effort"),
            Some(effort.as_str()),
            "{tool} {role}: {a:?}"
        );
        let a = f.argv(&[word, flag, "--effort", "low"]);
        assert_eq!(
            value_of(&a, "--model"),
            Some(model.as_str()),
            "{tool} {role}: {a:?}"
        );
        assert_eq!(
            value_of(&a, "--effort"),
            Some("low"),
            "{tool} {role}: {a:?}"
        );
    }
}

/// AC-3 (REQ-5): a saved choice for the tool and role wins over the table, and a role the file does
/// not save still gets the table.
#[test]
fn a_saved_choice_wins_over_the_table_and_the_other_role_keeps_the_table() {
    let f = Fixture::new();
    f.config(
        r#"{"launcher":{"models":{
            "claude":{"leader":{"model":"claude-sonnet-5-5","effort":"medium"}},
            "glm":{"agent":{"model":"zai/glm-5.3","effort":"low"}}}}}"#,
    );
    let a = f.argv(&["cc", "-l"]);
    assert_eq!(value_of(&a, "--model"), Some("claude-sonnet-5-5"), "{a:?}");
    assert_eq!(value_of(&a, "--effort"), Some("medium"), "{a:?}");
    let a = f.argv(&["glm", "-a"]);
    assert_eq!(value_of(&a, "--model"), Some("glm-5.3"), "{a:?}");
    assert_eq!(value_of(&a, "--effort"), Some("low"), "{a:?}");
    // Not saved: the claude agent and the glm leader fall back to the table.
    for (tool, role, model, effort) in table() {
        if (tool, role) == ("claude", "agent") || (tool, role) == ("glm", "leader") {
            let a = f.argv(&launcher_args(tool, role));
            assert_eq!(
                value_of(&a, "--model"),
                Some(model.as_str()),
                "{tool} {role}: {a:?}"
            );
            assert_eq!(
                value_of(&a, "--effort"),
                Some(effort.as_str()),
                "{tool} {role}: {a:?}"
            );
        }
    }
}

/// AC-4 (REQ-7): a session without a role keeps the model `opus` and the effort `medium`, with
/// or without a settings file (user decision 2026-11-08).
#[test]
fn a_session_without_a_role_keeps_opus_and_medium() {
    let f = Fixture::new();
    let a = f.argv(&["cc"]);
    assert_eq!(value_of(&a, "--model"), Some("opus"), "{a:?}");
    assert_eq!(value_of(&a, "--effort"), Some("medium"), "{a:?}");
    f.config(
        r#"{"launcher":{"models":{"claude":{"leader":{"model":"claude-sonnet-5-5","effort":"low"}}}}}"#,
    );
    let a = f.argv(&["cc"]);
    assert_eq!(value_of(&a, "--model"), Some("opus"), "{a:?}");
    assert_eq!(value_of(&a, "--effort"), Some("medium"), "{a:?}");
}

/// AC-6 (REQ-9): the help of `--model` and `--effort` no longer says that a Claude host defaults to
/// `opus` and `medium`, which holds only for a session without a role.
#[test]
fn the_model_and_effort_help_names_the_default_table() {
    let f = Fixture::new();
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["cc", "--help"])
        .env("HOME", f.home.path())
        .output()
        .unwrap();
    assert!(out.status.success());
    let help = String::from_utf8_lossy(&out.stdout).to_string();
    for stale in ["Claude hosts default `opus`", "Claude default `medium`"] {
        assert!(!help.contains(stale), "stale help text {stale:?}:\n{help}");
    }
    let section = |flag: &str| -> String {
        let start = help
            .find(flag)
            .unwrap_or_else(|| panic!("no {flag} in the help"));
        help[start..].lines().take(3).collect::<Vec<_>>().join(" ")
    };
    for flag in ["--model <MODEL>", "--effort <EFFORT>"] {
        let text = section(flag);
        assert!(text.contains("default table"), "{flag}: {text}");
        assert!(text.contains("role"), "{flag}: {text}");
    }
}
