//! SPEC-20261127-doctor-moai-layer: `moai doctor`, the read-only diagnosis of the moai layer of a
//! Claude Code setup (first checks: the Claude config folder, the Claude Code version against
//! the minimum-version table, the login source, the API key warning, and the pointers to Claude
//! Code's own checks). Every case runs against a temporary HOME and a stand-in `claude` that
//! records its calls; the real Claude config folder, the real home state folder and `claude` are never touched.
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
    /// The managed settings folder the command is told to read (`--managed-dir`), so no case
    /// reads the platform folder of this machine (REQ-15).
    managed: PathBuf,
    /// Holds the call record, outside the HOME tree the read-only test compares.
    _record: tempfile::TempDir,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        let bin = home.path().join("fake-bin");
        fs::create_dir_all(&bin).unwrap();
        let managed = home.path().join("managed-claude-code");
        fs::create_dir_all(&managed).unwrap();
        let record = tempfile::tempdir().unwrap();
        let calls = record.path().join("claude-calls.log");
        World {
            home,
            project,
            bin,
            calls,
            managed,
            _record: record,
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
        let managed = self.managed.to_str().unwrap();
        let mut args = vec!["doctor", "--project", project, "--managed-dir", managed];
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

    /// A file of the managed settings folder: `managed-settings.json` or a
    /// `managed-settings.d/<name>.json`.
    fn managed_file(&self, name: &str, json: &str) {
        let path = self.managed.join(name);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, json).unwrap();
    }

    /// `.claude/<name>` of the project (`settings.json` or `settings.local.json`).
    fn project_file(&self, name: &str, json: &str) {
        let dir = self.project.join(".claude");
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join(name), json).unwrap();
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
    // 2.1.30 is below the oldest row (`claude auth status`, 2.1.41).
    w.claude("2.1.30 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let c = check(&result, "claude_version");
    assert_eq!(c["state"], "warn", "{c}");
    assert_eq!(c["evidence"]["installed"], "2.1.30", "{c}");
    let rows = c["evidence"]["rows"].as_array().unwrap();
    let row = rows
        .iter()
        .find(|r| r["feature"] == "worktree.baseRef")
        .expect("the worktree.baseRef row");
    assert_eq!(row["min"], "2.1.133");
    assert_eq!(row["state"], "below");
    assert!(
        c["summary"].as_str().unwrap().contains("worktree.baseRef"),
        "{c}"
    );
    // A row of another platform is not judged; every other row is below.
    assert!(
        rows.iter()
            .filter(|r| r["state"] != "not_applicable")
            .all(|r| r["state"] == "below"),
        "2.1.30 is below every applicable row: {rows:?}"
    );
    let windows = rows
        .iter()
        .find(|r| r["feature"] == "cross-session messaging (native Windows)")
        .expect("the Windows row");
    assert_eq!(windows["state"], "not_applicable", "{windows}");
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
        .all(|r| r["state"] == "ok" || r["state"] == "not_applicable"));
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

// ---- locks_ (REQ-9, REQ-10; AC-6) -----------------------------------------------------------

const MODS_ONLY_MANAGED: &str = r#"{"pluginConfigs":{"cc-plugin-sec-default@builtin":{"options":{"allowManagedModsOnly":true}}}}"#;

/// One key of the `managed_locks` check.
fn lock<'a>(result: &'a Value, key: &str) -> &'a Value {
    let keys = &check(result, "managed_locks")["evidence"]["keys"];
    assert!(keys.get(key).is_some(), "no key {key} in {keys}");
    &keys[key]
}

fn strings(v: &Value) -> Vec<&str> {
    v.as_array()
        .unwrap_or_else(|| panic!("not a list: {v}"))
        .iter()
        .map(|s| s.as_str().unwrap())
        .collect()
}

#[test]
fn locks_managed_hooks_only_counts_in_the_managed_file_and_is_ignored_elsewhere() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let unset = w.result(&[], &[]);
    let k = lock(&unset, "allowManagedHooksOnly");
    assert_eq!(k["state"], "unset", "{k}");
    assert_eq!(state(&unset, "managed_locks"), "ok");

    // The user and the project carry it: the docs say it has no effect there.
    w.user_settings(r#"{"allowManagedHooksOnly": true}"#);
    w.project_file("settings.json", r#"{"allowManagedHooksOnly": true}"#);
    let ignored = w.result(&[], &[]);
    let k = lock(&ignored, "allowManagedHooksOnly");
    assert_eq!(k["state"], "ignored", "{k}");
    assert_eq!(
        strings(&k["ignored_in"]),
        ["user settings", "project settings"],
        "{k}"
    );
    assert_eq!(state(&ignored, "managed_locks"), "ok");

    // The managed file carries it: effective, with the file and the surfaces it locks.
    w.managed_file(
        "managed-settings.json",
        r#"{"allowManagedHooksOnly": true}"#,
    );
    let effective = w.result(&[], &[]);
    let k = lock(&effective, "allowManagedHooksOnly");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(k["source"], "managed-settings.json", "{k}");
    let surfaces = strings(&k["surfaces"]);
    assert!(
        surfaces.contains(&"hooks") && surfaces.contains(&"mods"),
        "{k}"
    );
    assert_eq!(state(&effective, "managed_locks"), "warn");
}

#[test]
fn locks_a_later_managed_file_overrides_an_earlier_one_in_alphabetical_order() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.managed_file(
        "managed-settings.json",
        r#"{"allowManagedHooksOnly": false}"#,
    );
    w.managed_file(
        "managed-settings.d/10-first.json",
        r#"{"allowManagedHooksOnly": true}"#,
    );
    let first = w.result(&[], &[]);
    let k = lock(&first, "allowManagedHooksOnly");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(k["source"], "managed-settings.d/10-first.json", "{k}");

    w.managed_file(
        "managed-settings.d/20-second.json",
        r#"{"allowManagedHooksOnly": false}"#,
    );
    let second = w.result(&[], &[]);
    assert_eq!(
        lock(&second, "allowManagedHooksOnly")["state"],
        "unset",
        "{second}"
    );
}

#[test]
fn locks_strict_plugin_only_locks_the_surfaces_the_managed_value_names() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.managed_file(
        "managed-settings.json",
        r#"{"strictPluginOnlyCustomization": true}"#,
    );
    let all = w.result(&[], &[]);
    let k = lock(&all, "strictPluginOnlyCustomization");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(
        strings(&k["surfaces"]),
        ["skills", "agents", "hooks", "mcp"],
        "{k}"
    );

    w.managed_file(
        "managed-settings.json",
        r#"{"strictPluginOnlyCustomization": ["hooks", "mcp"]}"#,
    );
    let some = w.result(&[], &[]);
    let k = lock(&some, "strictPluginOnlyCustomization");
    assert_eq!(strings(&k["surfaces"]), ["hooks", "mcp"], "{k}");

    // A wrong form locks all four surfaces (Claude Code 2.1.282 and later).
    w.managed_file(
        "managed-settings.json",
        r#"{"strictPluginOnlyCustomization": "yes"}"#,
    );
    let wrong = w.result(&[], &[]);
    let k = lock(&wrong, "strictPluginOnlyCustomization");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(
        strings(&k["surfaces"]),
        ["skills", "agents", "hooks", "mcp"],
        "{k}"
    );

    // In the user settings the key has no effect.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"strictPluginOnlyCustomization": true}"#);
    let user = w.result(&[], &[]);
    let k = lock(&user, "strictPluginOnlyCustomization");
    assert_eq!(k["state"], "ignored", "{k}");
}

#[test]
fn locks_mods_only_counts_only_under_the_managed_plugin_config() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.managed_file("managed-settings.json", MODS_ONLY_MANAGED);
    let managed = w.result(&[], &[]);
    let k = lock(&managed, "allowManagedModsOnly");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(k["source"], "managed-settings.json", "{k}");
    assert_eq!(strings(&k["surfaces"]), ["mods"], "{k}");

    // The same value in the user settings: Claude Code ignores the option there.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(MODS_ONLY_MANAGED);
    let user = w.result(&[], &[]);
    let k = lock(&user, "allowManagedModsOnly");
    assert_eq!(k["state"], "ignored", "{k}");
    assert_eq!(strings(&k["ignored_in"]), ["user settings"], "{k}");

    // A top-level key of the managed file is not that option.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.managed_file("managed-settings.json", r#"{"allowManagedModsOnly": true}"#);
    let top = w.result(&[], &[]);
    assert_eq!(lock(&top, "allowManagedModsOnly")["state"], "unset");
}

#[test]
fn locks_disable_all_hooks_counts_from_any_file_by_the_documented_precedence() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"disableAllHooks": true}"#);
    let user = w.result(&[], &[]);
    let k = lock(&user, "disableAllHooks");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(k["source"], "user settings", "{k}");
    // A non-managed file does not stop the managed hooks.
    assert!(!strings(&k["surfaces"]).contains(&"managed_hooks"), "{k}");

    // The project beats the user: false there turns it off again.
    w.project_file("settings.json", r#"{"disableAllHooks": false}"#);
    let project = w.result(&[], &[]);
    assert_eq!(lock(&project, "disableAllHooks")["state"], "unset");

    // The local file beats the project.
    w.project_file("settings.local.json", r#"{"disableAllHooks": true}"#);
    let local = w.result(&[], &[]);
    let k = lock(&local, "disableAllHooks");
    assert_eq!(k["state"], "effective", "{k}");
    assert_eq!(k["source"], "local settings", "{k}");

    // The managed file also stops the managed hooks.
    w.managed_file("managed-settings.json", r#"{"disableAllHooks": true}"#);
    let managed = w.result(&[], &[]);
    let k = lock(&managed, "disableAllHooks");
    assert_eq!(k["source"], "managed-settings.json", "{k}");
    assert!(strings(&k["surfaces"]).contains(&"managed_hooks"), "{k}");

    // A wrong managed value is discarded with a warning of Claude Code, so it locks nothing.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.managed_file("managed-settings.json", r#"{"disableAllHooks": "yes"}"#);
    let wrong = w.result(&[], &[]);
    let k = lock(&wrong, "disableAllHooks");
    assert_eq!(k["state"], "unset", "{k}");
    assert_eq!(strings(&k["discarded"]), ["managed-settings.json"], "{k}");
}

#[test]
fn locks_an_unreadable_file_makes_the_keys_it_could_set_unknown_and_names_it() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.managed_file(
        "managed-settings.json",
        r#"{"allowManagedHooksOnly": true}"#,
    );
    w.managed_file("managed-settings.d/20-broken.json", "{ this is not json");
    let broken = w.result(&[], &[]);
    for key in [
        "allowManagedHooksOnly",
        "strictPluginOnlyCustomization",
        "allowManagedModsOnly",
        "disableAllHooks",
    ] {
        let k = lock(&broken, key);
        assert_eq!(k["state"], "unknown", "{key}: {k}");
        assert_eq!(
            strings(&k["unreadable"]),
            ["managed-settings.d/20-broken.json"],
            "{key}: {k}"
        );
    }
    assert_eq!(state(&broken, "managed_locks"), "unknown");

    // An unreadable user file matters only for the key every file can set.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings("{ this is not json");
    let user = w.result(&[], &[]);
    let k = lock(&user, "disableAllHooks");
    assert_eq!(k["state"], "unknown", "{k}");
    assert_eq!(strings(&k["unreadable"]), ["user settings"], "{k}");
    assert_eq!(lock(&user, "allowManagedHooksOnly")["state"], "unset");
}

#[test]
fn locks_name_the_sources_that_are_never_read_as_files() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let not_read = strings(&check(&result, "managed_locks")["evidence"]["not_read"]).join(" | ");
    for name in ["policyHelper", "server-managed settings", "--settings"] {
        assert!(not_read.contains(name), "{name} in {not_read}");
    }
}

// ---- inbound_ (REQ-11; AC-7) ----------------------------------------------------------------

fn inbound(result: &Value) -> &Value {
    check(result, "cross_session_inbound")
}

#[test]
fn inbound_unset_warns_because_a_bypassed_lane_holds_messages_then() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&[], &[]);
    let c = inbound(&result);
    assert_eq!(c["state"], "warn", "{c}");
    assert_eq!(c["evidence"]["value"], "unset", "{c}");
    assert_eq!(c["evidence"]["source"], Value::Null, "{c}");
    assert!(c["summary"].as_str().unwrap().contains("bypass"), "{c}");
}

#[test]
fn inbound_accept_in_the_user_settings_is_ok_and_names_the_file() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"crossSessionInbound": "accept"}"#);
    let result = w.result(&[], &[]);
    let c = inbound(&result);
    assert_eq!(c["state"], "ok", "{c}");
    assert_eq!(c["evidence"]["value"], "accept", "{c}");
    assert_eq!(c["evidence"]["source"], "user settings", "{c}");
}

#[test]
fn inbound_hold_and_refuse_warn_and_the_managed_file_wins_over_the_user() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"crossSessionInbound": "accept"}"#);
    w.managed_file(
        "managed-settings.json",
        r#"{"crossSessionInbound": "hold"}"#,
    );
    let held = w.result(&[], &[]);
    let c = inbound(&held);
    assert_eq!(c["state"], "warn", "{c}");
    assert_eq!(c["evidence"]["value"], "hold", "{c}");
    assert_eq!(c["evidence"]["source"], "managed-settings.json", "{c}");

    w.managed_file(
        "managed-settings.d/50-late.json",
        r#"{"crossSessionInbound": "refuse"}"#,
    );
    let refused = w.result(&[], &[]);
    let c = inbound(&refused);
    assert_eq!(c["state"], "warn", "{c}");
    assert_eq!(c["evidence"]["value"], "refuse", "{c}");
    assert_eq!(
        c["evidence"]["source"], "managed-settings.d/50-late.json",
        "{c}"
    );
}

#[test]
fn inbound_a_project_or_local_value_counts_only_when_stricter_or_nothing_trusted_sets_one() {
    // The user accepts, the project refuses: stricter, so it counts.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"crossSessionInbound": "accept"}"#);
    w.project_file("settings.json", r#"{"crossSessionInbound": "refuse"}"#);
    let stricter = w.result(&[], &[]);
    let c = inbound(&stricter);
    assert_eq!(c["evidence"]["value"], "refuse", "{c}");
    assert_eq!(c["evidence"]["source"], "project settings", "{c}");

    // The user holds, the project accepts: not stricter, so the user value stays.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"crossSessionInbound": "hold"}"#);
    w.project_file("settings.json", r#"{"crossSessionInbound": "accept"}"#);
    let looser = w.result(&[], &[]);
    let c = inbound(&looser);
    assert_eq!(c["evidence"]["value"], "hold", "{c}");
    assert_eq!(c["evidence"]["source"], "user settings", "{c}");

    // No trusted source: the project's accept does not count, its hold does.
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.project_file("settings.json", r#"{"crossSessionInbound": "accept"}"#);
    let accept_only = w.result(&[], &[]);
    let c = inbound(&accept_only);
    assert_eq!(c["evidence"]["value"], "unset", "{c}");
    assert_eq!(c["state"], "warn", "{c}");
    w.project_file("settings.local.json", r#"{"crossSessionInbound": "hold"}"#);
    let local_hold = w.result(&[], &[]);
    let c = inbound(&local_hold);
    assert_eq!(c["evidence"]["value"], "hold", "{c}");
    assert_eq!(c["evidence"]["source"], "local settings", "{c}");
}

#[test]
fn inbound_a_wrong_value_is_hold_in_the_user_settings_and_refuse_in_the_managed_file() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings(r#"{"crossSessionInbound": "maybe"}"#);
    let user = w.result(&[], &[]);
    let c = inbound(&user);
    assert_eq!(c["evidence"]["value"], "hold", "{c}");
    assert_eq!(c["evidence"]["source"], "user settings", "{c}");

    w.managed_file(
        "managed-settings.json",
        r#"{"crossSessionInbound": "maybe"}"#,
    );
    let managed = w.result(&[], &[]);
    let c = inbound(&managed);
    assert_eq!(c["evidence"]["value"], "refuse", "{c}");
    assert_eq!(c["evidence"]["source"], "managed-settings.json", "{c}");
}

#[test]
fn inbound_an_unreadable_file_makes_the_check_unknown_and_sources_never_read_are_named() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    w.user_settings("{ this is not json");
    let result = w.result(&[], &[]);
    let c = inbound(&result);
    assert_eq!(c["state"], "unknown", "{c}");
    assert_eq!(
        strings(&c["evidence"]["unreadable"]),
        ["user settings"],
        "{c}"
    );
    let not_read = strings(&c["evidence"]["not_read"]).join(" | ");
    for name in ["server-managed settings", "--settings"] {
        assert!(not_read.contains(name), "{name} in {not_read}");
    }
}

// ---- table_ (REQ-12; AC-8) ------------------------------------------------------------------

const COLUMNS: [&str; 5] = ["cli", "gateway", "desktop", "vscode", "cloud"];

fn cell_text(cell: &Value, field: &str) -> String {
    cell[field]
        .as_str()
        .filter(|s| !s.is_empty())
        .unwrap_or_else(|| panic!("cell lacks {field}: {cell}"))
        .to_string()
}

#[test]
fn table_is_printed_only_with_the_option_with_five_columns_and_eight_rows() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let bare = w.result(&[], &[]);
    assert!(bare.get("capability_table").is_none(), "{bare}");

    let with = w.result(&["--table"], &[]);
    let table = &with["capability_table"];
    assert_eq!(strings(&table["columns"]), COLUMNS, "{table}");
    let rows = table["rows"].as_array().unwrap();
    assert_eq!(rows.len(), 8, "{table}");
    for row in rows {
        assert!(row["feature"].as_str().is_some_and(|f| !f.is_empty()));
        for column in COLUMNS {
            let cell = &row["cells"][column];
            let value = cell["value"].as_str().unwrap_or_default();
            assert!(
                ["yes", "no", "partial", "unverified"].contains(&value),
                "{row}: {column}"
            );
            cell_text(cell, "docs");
            cell_text(cell, "line");
            let read_at = cell_text(cell, "read_at");
            assert_eq!(read_at.len(), 10, "{read_at}");
        }
    }
}

#[test]
fn table_never_prints_an_unverified_cell_as_yes() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let result = w.result(&["--table"], &[]);
    let rows = result["capability_table"]["rows"].as_array().unwrap();
    let cell = |feature: &str, column: &str| {
        rows.iter()
            .find(|r| r["feature"] == feature)
            .unwrap_or_else(|| panic!("no row {feature}"))["cells"][column]
            .clone()
    };
    // The cells the SPEC's research could not read from the docs, or only inferred.
    for (feature, column) in [
        ("cross_session_same_machine", "gateway"),
        ("cross_session_same_machine", "cloud"),
        ("cross_session_other_machine", "desktop"),
        ("cross_session_other_machine", "vscode"),
        ("output_styles", "gateway"),
        ("rate_limit_status_line", "desktop"),
        ("artifacts", "vscode"),
    ] {
        assert_eq!(
            cell(feature, column)["value"],
            "unverified",
            "{feature}/{column}"
        );
    }
    // A cell the docs state outright keeps its value.
    assert_eq!(cell("remote_control", "cli")["value"], "yes");
    assert_eq!(cell("remote_control", "gateway")["value"], "no");
    for row in rows {
        for column in COLUMNS {
            let c = &row["cells"][column];
            if c["value"] == "yes" {
                assert!(
                    !cell_text(c, "line").to_lowercase().contains("unverified"),
                    "{row}"
                );
            }
        }
    }
}

#[test]
fn table_names_the_host_it_detects_from_the_base_url_and_never_prints_the_url() {
    let w = World::new();
    w.claude("2.1.332 (Claude Code)", "{}", 0);
    let host = |env: &[(&str, &str)]| -> (String, String) {
        let out = w.run(&[], env);
        let shown = String::from_utf8_lossy(&out.stdout).to_string();
        let v: Value = serde_json::from_str(&shown).unwrap();
        let caps = check(&v["result"], "host_capabilities").clone();
        (
            caps["evidence"]["host"].as_str().unwrap().to_string(),
            shown,
        )
    };
    assert_eq!(host(&[]).0, "cli");
    assert_eq!(
        host(&[("ANTHROPIC_BASE_URL", "https://api.anthropic.com")]).0,
        "cli"
    );
    let (other, shown) = host(&[("ANTHROPIC_BASE_URL", "https://gateway.example.invalid/v1")]);
    assert_eq!(other, "gateway");
    assert!(!shown.contains("gateway.example.invalid"), "{shown}");
    // Credentials in the URL are not read as a host and never printed.
    let (with_user, shown) = host(&[(
        "ANTHROPIC_BASE_URL",
        "https://user:PWSECRET@api.anthropic.com/",
    )]);
    assert_eq!(with_user, "cli");
    assert!(!shown.contains("PWSECRET"), "{shown}");

    // The detected host's cells are in the check: one per row, the column of that host.
    let out = w.result(
        &[],
        &[("ANTHROPIC_BASE_URL", "https://gateway.example.invalid")],
    );
    let cells = &check(&out, "host_capabilities")["evidence"]["cells"];
    assert_eq!(cells.as_object().unwrap().len(), 8, "{cells}");
    assert_eq!(cells["remote_control"]["value"], "no", "{cells}");
}

// ---- auth status time-out (note 10861 of C-338) ----------------------------------------------

/// Whether the process `pid` is still alive, by the system `kill -0`.
fn alive(pid: &str) -> bool {
    std::process::Command::new("/bin/kill")
        .args(["-0", pid])
        .stderr(std::process::Stdio::null())
        .status()
        .is_ok_and(|s| s.success())
}

#[test]
fn login_a_hanging_auth_status_is_stopped_with_its_whole_process_group() {
    let w = World::new();
    let pid_file = w._record.path().join("grandchild.pid");
    let script = format!(
        "#!/bin/sh\n\
         case \"$1\" in\n\
         --version) echo '2.1.332 (Claude Code)' ;;\n\
         auth) sleep 60 &\n\
               echo $! > '{pid}'\n\
               wait ;;\n\
         esac\n",
        pid = pid_file.display()
    );
    common::write_executable(&w.bin.join("claude"), &script);
    let result = w.result(&["--auth-status"], &[]);
    let auth = &check(&result, "login")["evidence"]["auth_status"];
    assert_eq!(auth["state"], "unknown", "{auth}");
    assert_eq!(auth["reason"], "timeout", "{auth}");

    let pid = text_of(&pid_file).trim().to_string();
    assert!(!pid.is_empty(), "the stand-in wrote no grandchild pid");
    // The grandchild has had the time the group signal needs; it must be gone.
    let mut gone = false;
    for _ in 0..30 {
        if !alive(&pid) {
            gone = true;
            break;
        }
        std::thread::sleep(std::time::Duration::from_millis(100));
    }
    // Never leave the stand-in's `sleep` behind, whatever the answer was.
    let _ = std::process::Command::new("/bin/kill")
        .args(["-KILL", &pid])
        .stderr(std::process::Stdio::null())
        .status();
    assert!(
        gone,
        "the grandchild {pid} outlived the time-out of auth status"
    );
}
