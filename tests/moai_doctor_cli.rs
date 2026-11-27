//! SPEC-20261127-doctor-moai-layer: `moai doctor`, the read-only diagnosis of the moai layer of a
//! Claude Code setup (first checks: the Claude config folder, the Claude Code version against
//! the minimum-version table, the login source, the API key warning, and the pointers to Claude
//! Code's own checks). Every case runs against a temporary HOME and a stand-in `claude` that
//! records its calls; the real `~/.claude`, `~/.moai` and `claude` are never touched.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

/// The names the login check reports, in the order of the SPEC (REQ-6).
const LOGIN_NAMES: [&str; 7] = [
    "ANTHROPIC_API_KEY",
    "ANTHROPIC_AUTH_TOKEN",
    "CLAUDE_CODE_OAUTH_TOKEN",
    "ANTHROPIC_PROFILE",
    "CLAUDE_CODE_USE_BEDROCK",
    "CLAUDE_CODE_USE_VERTEX",
    "CLAUDE_CODE_USE_FOUNDRY",
];

/// A temporary HOME, a project inside it and a PATH folder that may hold a stand-in `claude`.
struct World {
    home: TempHome,
    project: PathBuf,
    bin: PathBuf,
    calls: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        let bin = home.path().join("fake-bin");
        fs::create_dir_all(&bin).unwrap();
        let calls = home.path().join("claude-calls.log");
        World {
            home,
            project,
            bin,
            calls,
        }
    }

    /// A stand-in `claude` that records its arguments and answers `--version` with
    /// `version_line` and `auth status` with `auth_json` and `auth_exit`.
    fn claude(&self, version_line: &str, auth_json: &str, auth_exit: i32) {
        let script = format!(
            "#!/bin/sh\necho \"$*\" >> '{calls}'\n\
             case \"$1\" in\n\
             --version) echo '{version_line}' ;;\n\
             auth) echo '{auth_json}'; exit {auth_exit} ;;\n\
             *) exit 9 ;;\n\
             esac\n",
            calls = self.calls.display()
        );
        common::write_executable(&self.bin.join("claude"), &script);
    }

    /// `moai doctor --project <project> <extra…>` with the given extra environment.
    fn run(&self, extra: &[&str], env: &[(&str, &str)]) -> Output {
        let project = self.project.to_str().unwrap();
        let mut args = vec!["doctor", "--project", project];
        args.extend_from_slice(extra);
        let mut cmd = self.home.command(&args);
        cmd.env("PATH", format!("{}:/usr/bin:/bin", self.bin.display()))
            .env("MOAI_ACCOUNT_HOME", self.home.path());
        for (k, v) in env {
            cmd.env(k, v);
        }
        cmd.output().unwrap()
    }

    fn result(&self, extra: &[&str], env: &[(&str, &str)]) -> Value {
        let out = self.run(extra, env);
        assert_eq!(
            out.status.code(),
            Some(0),
            "{} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let v: Value = serde_json::from_slice(&out.stdout).expect("stdout is the JSON envelope");
        assert_eq!(v["ok"], Value::Bool(true), "{v}");
        v["result"].clone()
    }

    /// What the stand-in `claude` was called with, one line per call.
    fn recorded_calls(&self) -> Vec<String> {
        fs::read_to_string(&self.calls)
            .map(|t| t.lines().map(str::to_string).collect())
            .unwrap_or_default()
    }

    /// The user settings file of the Claude config folder (the default `~/.claude`).
    fn user_settings(&self, json: &str) {
        let dir = self.home.path().join(".claude");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("settings.json"), json).unwrap();
    }
}

fn check<'a>(result: &'a Value, id: &str) -> &'a Value {
    result["checks"]
        .as_array()
        .expect("checks is a list")
        .iter()
        .find(|c| c["id"] == id)
        .unwrap_or_else(|| panic!("no check {id} in {result}"))
}

fn state(result: &Value, id: &str) -> String {
    check(result, id)["state"].as_str().unwrap().to_string()
}

fn text_of(path: &Path) -> String {
    fs::read_to_string(path).unwrap_or_default()
}

// ---- read_only_ (REQ-1, REQ-2, REQ-14; AC-1) ------------------------------------------------

#[test]
fn read_only_answer_has_the_fields_writes_nothing_and_prints_no_secret() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"apiKeyHelper": "echo HELPEROUTPUT789"}"#);
    let creds = w.home.path().join(".claude").join(".credentials.json");
    fs::write(&creds, "CREDSECRET000").unwrap();
    let before = w.home.snapshot();
    let creds_before = text_of(&creds);

    let out = w.run(
        &[],
        &[
            ("ANTHROPIC_API_KEY", "sk-ant-SECRETVALUE123"),
            ("ANTHROPIC_AUTH_TOKEN", "TOKENVALUE456"),
            ("CLAUDE_CODE_OAUTH_TOKEN", "OAUTHVALUE789"),
        ],
    );
    assert_eq!(out.status.code(), Some(0));
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    for secret in [
        "SECRETVALUE123",
        "TOKENVALUE456",
        "OAUTHVALUE789",
        "HELPEROUTPUT789",
        "CREDSECRET000",
    ] {
        assert!(!stdout.contains(secret), "{secret} leaked: {stdout}");
    }
    let v: Value = serde_json::from_str(&stdout).unwrap();
    let result = &v["result"];
    assert_eq!(v["operation"], "doctor", "{v}");
    assert_eq!(result["writes"], Value::Bool(false), "{result}");
    for id in ["config_folder", "claude_version", "login", "api_key"] {
        let c = check(result, id);
        let s = c["state"].as_str().unwrap();
        assert!(
            ["ok", "warn", "fail", "unknown", "not_applicable"].contains(&s),
            "{id}: {s}"
        );
        for field in ["summary", "evidence", "next"] {
            assert!(c.get(field).is_some(), "{id} lacks {field}: {c}");
        }
    }
    assert!(result["pointers"].is_array(), "{result}");

    assert_eq!(w.home.snapshot(), before, "the HOME tree is unchanged");
    assert_eq!(
        text_of(&creds),
        creds_before,
        "the credentials are untouched"
    );
    // The only call the stand-in `claude` saw is the version probe.
    assert_eq!(w.recorded_calls(), vec!["--version".to_string()]);
}

// ---- config_folder_ (REQ-3; AC-2) -----------------------------------------------------------

#[test]
fn config_folder_relative_variable_fails_naming_the_variable_and_not_its_value() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[("CLAUDE_CONFIG_DIR", "some/relative/dir")]);
    let c = check(&result, "config_folder");
    assert_eq!(c["state"], "fail", "{c}");
    let shown = c.to_string();
    assert!(shown.contains("CLAUDE_CONFIG_DIR"), "{shown}");
    assert!(!shown.contains("some/relative/dir"), "{shown}");
    // A check that needs the folder says so, and the version check still runs.
    assert_eq!(state(&result, "login"), "unknown", "{result}");
    assert_ne!(state(&result, "claude_version"), "unknown", "{result}");
}

#[test]
fn config_folder_empty_variable_fails_the_same_way() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[("CLAUDE_CONFIG_DIR", "")]);
    assert_eq!(state(&result, "config_folder"), "fail");
}

#[test]
fn config_folder_default_is_ok() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    assert_eq!(state(&result, "config_folder"), "ok", "{result}");
}

// ---- versions_ (REQ-4, REQ-5; AC-3) ---------------------------------------------------------

#[test]
fn versions_below_the_table_name_the_rows_and_their_versions() {
    let w = World::new();
    w.claude("2.1.100 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let c = check(&result, "claude_version");
    assert_eq!(c["state"], "warn", "{c}");
    assert_eq!(c["evidence"]["installed"], "2.1.100", "{c}");
    let rows = c["evidence"]["rows"].as_array().unwrap();
    let row = rows
        .iter()
        .find(|r| r["feature"] == "worktree.baseRef")
        .expect("the worktree.baseRef row");
    assert_eq!(row["min"], "2.1.133");
    assert_eq!(row["state"], "below");
    assert!(
        rows.iter().all(|r| r["state"] == "below"),
        "2.1.100 is below every row: {rows:?}"
    );
}

#[test]
fn versions_at_or_above_every_row_are_ok() {
    let w = World::new();
    w.claude("2.1.999 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let c = check(&result, "claude_version");
    assert_eq!(c["state"], "ok", "{c}");
    assert!(c["evidence"]["rows"]
        .as_array()
        .unwrap()
        .iter()
        .all(|r| r["state"] == "ok"));
}

#[test]
fn versions_judge_each_row_against_its_own_minimum() {
    let w = World::new();
    // Exactly the minimum of `notify_when_idle` (2.1.236): that row is ok, a later one is below.
    w.claude("2.1.236 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let rows = check(&result, "claude_version")["evidence"]["rows"]
        .as_array()
        .unwrap()
        .clone();
    let find = |name: &str| rows.iter().find(|r| r["feature"] == name).unwrap().clone();
    assert_eq!(find("notify_when_idle")["state"], "ok");
    assert_eq!(find("hook onFailure")["state"], "below");
}

#[test]
fn versions_unreadable_output_is_unknown() {
    let w = World::new();
    w.claude("not a version at all", "{}", 0);
    let result = w.result(&[], &[]);
    let c = check(&result, "claude_version");
    assert_eq!(c["state"], "unknown", "{c}");
    assert!(c["evidence"]["rows"]
        .as_array()
        .is_none_or(|r| r.is_empty()));
}

#[test]
fn versions_missing_claude_is_unknown_with_the_reason_not_found() {
    let w = World::new();
    let result = w.result(&[], &[]);
    let c = check(&result, "claude_version");
    assert_eq!(c["state"], "unknown", "{c}");
    assert_eq!(c["evidence"]["reason"], "not_found", "{c}");
    // The rest of the report is there.
    assert_eq!(state(&result, "config_folder"), "ok");
}

#[test]
fn versions_a_hanging_claude_times_out() {
    let w = World::new();
    common::write_executable(&w.bin.join("claude"), "#!/bin/sh\nsleep 30\n");
    let result = w.result(&[], &[]);
    let c = check(&result, "claude_version");
    assert_eq!(c["state"], "unknown", "{c}");
    assert_eq!(c["evidence"]["reason"], "timeout", "{c}");
}

// ---- login_ (REQ-6, REQ-7; AC-4) ------------------------------------------------------------

#[test]
fn login_reports_every_name_as_set_or_unset_and_the_credential_that_would_win() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let bare = w.result(&[], &[]);
    let login = check(&bare, "login");
    for name in LOGIN_NAMES {
        assert_eq!(
            login["evidence"]["variables"][name], "unset",
            "{name}: {login}"
        );
    }
    assert_eq!(login["evidence"]["would_pick"], "none_visible", "{login}");

    let with_key = w.result(&[], &[("ANTHROPIC_API_KEY", "k")]);
    let login = check(&with_key, "login");
    assert_eq!(login["evidence"]["variables"]["ANTHROPIC_API_KEY"], "set");
    assert_eq!(login["evidence"]["would_pick"], "ANTHROPIC_API_KEY");

    // The documented precedence: a cloud provider variable beats a token, a token beats a key.
    let both = w.result(
        &[],
        &[
            ("ANTHROPIC_API_KEY", "k"),
            ("ANTHROPIC_AUTH_TOKEN", "t"),
            ("CLAUDE_CODE_USE_BEDROCK", "1"),
        ],
    );
    assert_eq!(
        check(&both, "login")["evidence"]["would_pick"],
        "cloud_provider"
    );
    let token = w.result(
        &[],
        &[("ANTHROPIC_API_KEY", "k"), ("ANTHROPIC_AUTH_TOKEN", "t")],
    );
    assert_eq!(
        check(&token, "login")["evidence"]["would_pick"],
        "ANTHROPIC_AUTH_TOKEN"
    );
}

#[test]
fn login_names_the_credentials_file_by_existence_only() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let absent = w.result(&[], &[]);
    assert_eq!(
        check(&absent, "login")["evidence"]["credentials_file"],
        "absent"
    );
    let dir = w.home.path().join(".claude");
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join(".credentials.json"), "CREDSECRET000").unwrap();
    let present = w.run(&[], &[]);
    let shown = String::from_utf8_lossy(&present.stdout).to_string();
    assert!(!shown.contains("CREDSECRET000"), "{shown}");
    let v: Value = serde_json::from_str(&shown).unwrap();
    let login = check(&v["result"], "login");
    assert_eq!(login["evidence"]["credentials_file"], "present", "{login}");
    assert_eq!(login["evidence"]["would_pick"], "login_file", "{login}");
}

#[test]
fn login_expiry_is_unknown_without_the_option_and_never_runs_auth_status() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let login = check(&result, "login");
    assert_eq!(login["evidence"]["expiry"], "unknown", "{login}");
    assert!(
        login["next"]
            .as_str()
            .unwrap()
            .contains("claude auth status --text"),
        "{login}"
    );
    assert!(
        !w.recorded_calls().iter().any(|c| c.starts_with("auth")),
        "{:?}",
        w.recorded_calls()
    );
}

#[test]
fn login_with_the_option_keeps_three_fields_and_the_exit_code_only() {
    let w = World::new();
    w.claude(
        "2.1.332 (Claude Code)",
        r#"{"authMethod":"claude.ai","configDirectory":"/cfg","loginExpiresAt":"2026-12-01T00:00:00.000Z","email":"leak@example.invalid","organization":"LeakOrg"}"#,
        0,
    );
    let out = w.run(&["--auth-status"], &[]);
    let shown = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(!shown.contains("leak@example.invalid"), "{shown}");
    assert!(!shown.contains("LeakOrg"), "{shown}");
    let v: Value = serde_json::from_str(&shown).unwrap();
    let auth = &check(&v["result"], "login")["evidence"]["auth_status"];
    assert_eq!(auth["exit"], 0, "{auth}");
    assert_eq!(auth["authMethod"], "claude.ai", "{auth}");
    assert_eq!(auth["configDirectory"], "/cfg", "{auth}");
    assert_eq!(auth["loginExpiresAt"], "2026-12-01T00:00:00.000Z", "{auth}");
    assert_eq!(
        w.recorded_calls(),
        vec!["--version".to_string(), "auth status".to_string()]
    );
}

#[test]
fn login_with_the_option_records_a_failing_exit_and_a_missing_command() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", r#"{"authMethod":"none"}"#, 1);
    let result = w.result(&["--auth-status"], &[]);
    let auth = &check(&result, "login")["evidence"]["auth_status"];
    assert_eq!(auth["exit"], 1, "{auth}");
    assert_eq!(auth["authMethod"], "none", "{auth}");

    let w = World::new();
    let result = w.result(&["--auth-status"], &[]);
    let auth = &check(&result, "login")["evidence"]["auth_status"];
    assert_eq!(auth["state"], "unknown", "{auth}");
    assert_eq!(auth["reason"], "not_found", "{auth}");
}

// ---- api_key_ (REQ-8; AC-5) -----------------------------------------------------------------

#[test]
fn api_key_is_ok_when_nothing_is_set() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let c = check(&result, "api_key");
    assert_eq!(c["state"], "ok", "{c}");
    assert_eq!(c["evidence"]["present"], serde_json::json!([]), "{c}");
}

#[test]
fn api_key_warns_naming_each_variable_and_what_it_switches_off() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(
        &[],
        &[("ANTHROPIC_API_KEY", "k"), ("ANTHROPIC_AUTH_TOKEN", "t")],
    );
    let c = check(&result, "api_key");
    assert_eq!(c["state"], "warn", "{c}");
    let present = c["evidence"]["present"].to_string();
    assert!(present.contains("ANTHROPIC_API_KEY"), "{present}");
    assert!(present.contains("ANTHROPIC_AUTH_TOKEN"), "{present}");
    let summary = c["summary"].as_str().unwrap();
    assert!(summary.contains("instead of"), "{summary}");
    assert!(summary.contains("Remote Control"), "{summary}");
}

#[test]
fn api_key_helper_in_the_user_or_project_settings_counts_by_name() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"apiKeyHelper": "echo HELPEROUTPUT789"}"#);
    let result = w.result(&[], &[]);
    let c = check(&result, "api_key");
    assert_eq!(c["state"], "warn", "{c}");
    assert!(c["evidence"]["present"]
        .to_string()
        .contains("apiKeyHelper (user settings)"));
    assert_eq!(
        check(&result, "login")["evidence"]["would_pick"],
        "api_key_helper"
    );

    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let dir = w.project.join(".claude");
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join("settings.json"), r#"{"apiKeyHelper": "x"}"#).unwrap();
    let result = w.result(&[], &[]);
    assert!(check(&result, "api_key")["evidence"]["present"]
        .to_string()
        .contains("apiKeyHelper (project settings)"));
}

#[test]
fn api_key_settings_that_are_not_json_do_not_stop_the_report() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings("{ this is not json");
    let result = w.result(&[], &[]);
    assert_eq!(state(&result, "api_key"), "ok", "{result}");
    assert_eq!(state(&result, "claude_version"), "ok", "{result}");
}

// ---- pointers_ (REQ-13; AC-9) ---------------------------------------------------------------

#[test]
fn pointers_name_the_claude_code_checks_and_mark_each_against_the_docs() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let pointers = result["pointers"].as_array().unwrap();
    assert!(!pointers.is_empty());
    let commands: Vec<&str> = pointers
        .iter()
        .map(|p| p["command"].as_str().unwrap())
        .collect();
    for wanted in ["claude doctor", "/status", "/context"] {
        assert!(commands.contains(&wanted), "{wanted} in {commands:?}");
    }
    for p in pointers {
        let docs = p["docs"].as_str().unwrap();
        assert!(["verified", "unverified"].contains(&docs), "{p}");
    }
    // `claude doctor` and `/status` are named by the official pages the SPEC read.
    let doc_of =
        |name: &str| pointers.iter().find(|p| p["command"] == name).unwrap()["docs"].clone();
    assert_eq!(doc_of("claude doctor"), "verified");
    assert_eq!(doc_of("/status"), "verified");
    // None of them is run: the stand-in saw the version probe only.
    assert_eq!(w.recorded_calls(), vec!["--version".to_string()]);
}
