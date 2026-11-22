//! `moai` headless runs (SPEC-20261122-headless-run-contract): every `claude -p` command moai
//! starts is built in one runner function that pins the permission mode, the MCP flags, the
//! model, the output schema and the budgets, refuses an audit run requested by the change's
//! author session, keeps credential values out of the child argv and environment, and warns once
//! when `ANTHROPIC_API_KEY` is present. Every run goes through a stand-in `claude` script that
//! records its argv and environment; no test starts a real `claude`, reaches the network or
//! spends money.
#![cfg(unix)]

mod common;

use moai_cli::headless_run::{build, run, HeadlessRunSpec, RunContext, RunRole};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};

const REQUESTING: &str = "11111111-0000-4000-8000-000000000001";
const AUTHOR_OTHER: &str = "22222222-0000-4000-8000-000000000002";
const SECRET: &str = "super-secret-value-XYZ";

fn audit_spec() -> HeadlessRunSpec {
    HeadlessRunSpec {
        role: RunRole::Audit {
            author_session: AUTHOR_OTHER.to_string(),
        },
        model: "sonnet".to_string(),
        agent: Some("moai-audit".to_string()),
        schema: Some(
            json!({"type": "object", "properties": {"verdict": {"type": "string"}}, "required": ["verdict"]}),
        ),
        prompt: "audit the head".to_string(),
        max_turns: 40,
        max_budget_usd: "5.00".to_string(),
        allowed_tools: vec!["Read".to_string(), "Bash(git diff *)".to_string()],
        foreign_tree: false,
        gateway_credential: None,
    }
}

fn ctx(platform_root: &Path) -> RunContext {
    RunContext {
        requesting_session: REQUESTING.to_string(),
        platform_root: Some(platform_root.to_path_buf()),
    }
}

/// The parent environment the runner inherits: a stand-in `claude` first on PATH, a planted
/// credential, a nested-session marker, and the record path the stand-in writes to.
fn parent_env(fixtures: &Path, record: &Path) -> BTreeMap<String, String> {
    BTreeMap::from([
        (
            "PATH".to_string(),
            format!("{}:/usr/bin:/bin", fixtures.display()),
        ),
        ("HOME".to_string(), "/tmp".to_string()),
        ("ANTHROPIC_API_KEY".to_string(), SECRET.to_string()),
        ("CLAUDECODE".to_string(), "1".to_string()),
        (
            "MOAI_STANDIN_RECORD".to_string(),
            record.display().to_string(),
        ),
    ])
}

/// A stand-in `claude` (SPEC D-5): records its argv and environment to
/// `$MOAI_STANDIN_RECORD` in the form `ARGV`, one arg per line, then `ENV`, sorted `env`
/// lines, then answers one JSON result and exits zero.
fn standin_claude(fixtures: &Path) {
    let script = "#!/bin/sh\n\
r=\"$MOAI_STANDIN_RECORD\"\n\
{ printf 'ARGV\\n'; for a in \"$@\"; do printf '%s\\n' \"$a\"; done; printf 'ENV\\n'; env | LC_ALL=C sort; } > \"$r\"\n\
printf '%s' '{\"type\":\"result\",\"subtype\":\"success\",\"is_error\":false,\"session_id\":\"standin\",\"total_cost_usd\":0.01,\"num_turns\":1}'\n";
    common::write_executable(&fixtures.join("claude"), script);
}

/// A minimal parent environment whose PATH carries the stand-in `claude`, for tests that
/// assert on the built argv without starting the run.
fn standin_env(fixtures: &Path) -> BTreeMap<String, String> {
    BTreeMap::from([
        (
            "PATH".to_string(),
            format!("{}:/usr/bin:/bin", fixtures.display()),
        ),
        ("HOME".to_string(), "/tmp".to_string()),
    ])
}

/// `tmp/bin` with a stand-in `claude` in it.
fn standin_dir(tmp: &Path) -> PathBuf {
    let fixtures = tmp.join("bin");
    fs::create_dir_all(&fixtures).unwrap();
    standin_claude(&fixtures);
    fixtures
}

/// Reads the stand-in's record as `(argv lines, raw text)`.
fn read_record(record: &Path) -> (Vec<String>, String) {
    let text = fs::read_to_string(record).unwrap();
    let argv: Vec<String> = text
        .lines()
        .skip(1)
        .take_while(|l| *l != "ENV")
        .map(str::to_string)
        .collect();
    (argv, text)
}

/// The value of the argument that follows `flag` in the recorded argv.
fn after<'a>(argv: &'a [String], flag: &str) -> &'a str {
    let i = argv.iter().position(|a| a == flag).expect(flag);
    &argv[i + 1]
}

fn count(argv: &[String], flag: &str) -> usize {
    argv.iter().filter(|a| *a == flag).count()
}

fn settings_json(argv: &[String]) -> Value {
    serde_json::from_str(after(argv, "--settings")).unwrap()
}

/// AC-1: the stand-in records the exact argv and environment, every pinned flag of REQ-2 to
/// REQ-10 is present, and neither the argv nor the environment dump carries the planted
/// credential value.
#[test]
fn standin_records_every_pinned_flag() {
    let tmp = tempfile::tempdir().unwrap();
    let fixtures = tmp.path().join("bin");
    fs::create_dir_all(&fixtures).unwrap();
    let record = tmp.path().join("record.txt");
    standin_claude(&fixtures);
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();

    let built = build(&audit_spec(), &ctx(&root), &parent_env(&fixtures, &record)).unwrap();

    assert_eq!(built.args.first().map(String::as_str), Some("-p"));
    assert_eq!(after(&built.args, "--permission-mode"), "dontAsk");
    assert_eq!(after(&built.args, "--permission-prompts"), "none");
    assert_eq!(
        after(&built.args, "--allowedTools"),
        "Read,Bash(git diff *)"
    );
    assert!(built.args.contains(&"--strict-mcp-config".to_string()));
    assert!(!built.args.iter().any(|a| a == "--mcp-config"));
    assert_eq!(after(&built.args, "--model"), "sonnet");
    assert_eq!(after(&built.args, "--agent"), "moai-audit");
    assert_eq!(after(&built.args, "--output-format"), "json");
    let schema: Value = serde_json::from_str(after(&built.args, "--json-schema")).unwrap();
    assert_eq!(schema["required"][0], "verdict");
    assert_eq!(after(&built.args, "--max-turns"), "40");
    assert_eq!(after(&built.args, "--max-budget-usd"), "5.00");
    assert_eq!(
        after(&built.args, "--setting-sources"),
        "user,project,local"
    );
    assert_eq!(count(&built.args, "--settings"), 1);
    assert_eq!(
        built.args.last().map(String::as_str),
        Some("audit the head")
    );
    assert_eq!(
        built
            .env
            .get("CLAUDE_CODE_RETRY_WATCHDOG")
            .map(String::as_str),
        Some("1")
    );
    assert!(
        built
            .env
            .get("BASH_DEFAULT_TIMEOUT_MS")
            .unwrap()
            .parse::<u64>()
            .unwrap()
            >= 600000
    );
    assert!(!built.env.contains_key("ANTHROPIC_API_KEY"));
    assert!(!built.env.contains_key("CLAUDECODE"));
    assert!(built.scrubbed.contains(&"ANTHROPIC_API_KEY".to_string()));
    assert!(built
        .env
        .get("PATH")
        .unwrap()
        .starts_with(fixtures.display().to_string().as_str()));

    let raw = run(&built, tmp.path()).unwrap();
    assert!(
        raw.success,
        "stand-in stdout: {} stderr: {}",
        raw.stdout, raw.stderr
    );
    let (argv, text) = read_record(&record);
    assert_eq!(argv, built.args);
    assert!(text.contains("CLAUDE_CODE_RETRY_WATCHDOG=1"));
    assert!(
        !text.contains(SECRET),
        "credential value leaked into the stand-in record"
    );
}

/// AC-2: a deployed platform `managed-mcp.json` removes both MCP flags; no file keeps
/// `--strict-mcp-config` and never `--mcp-config`.
#[test]
fn managed_mcp_adapts_the_flags() {
    let tmp = tempfile::tempdir().unwrap();
    let fixtures = tmp.path().join("bin");
    fs::create_dir_all(&fixtures).unwrap();
    let record = tmp.path().join("record.txt");
    standin_claude(&fixtures);
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let env = parent_env(&fixtures, &record);

    let deployed = build(&audit_spec(), &ctx(&root), &env).unwrap();
    assert!(deployed.args.contains(&"--strict-mcp-config".to_string()));
    assert!(!deployed.args.iter().any(|a| a == "--mcp-config"));
    assert!(!deployed.managed_mcp_deployed);

    fs::write(root.join("managed-mcp.json"), "{\"mcpServers\": {}}").unwrap();
    let bare = build(&audit_spec(), &ctx(&root), &env).unwrap();
    assert!(!bare.args.contains(&"--strict-mcp-config".to_string()));
    assert!(!bare.args.iter().any(|a| a == "--mcp-config"));
    assert!(bare.managed_mcp_deployed);
}

/// AC-3: one `--settings` object, role-scoped — the lane run adds
/// `permissions.denyInsteadOfPrompting` on top of the inbound accept, the plain worker run
/// carries only the inbound key, and no run writes the key into a settings file.
#[test]
fn settings_object_is_one_and_role_scoped() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let env = standin_env(&standin_dir(tmp.path()));

    let mut lane = audit_spec();
    lane.role = RunRole::Lane;
    let built = build(&lane, &ctx(&root), &env).unwrap();
    assert_eq!(count(&built.args, "--settings"), 1);
    let settings = settings_json(&built.args);
    assert_eq!(settings["crossSessionInbound"], "accept");
    assert_eq!(settings["permissions"]["denyInsteadOfPrompting"], true);
    assert_eq!(after(&built.args, "--permission-mode"), "auto");

    let mut worker = audit_spec();
    worker.role = RunRole::Worker;
    worker.agent = None;
    let built = build(&worker, &ctx(&root), &env).unwrap();
    assert_eq!(count(&built.args, "--settings"), 1);
    let settings = settings_json(&built.args);
    assert_eq!(settings["crossSessionInbound"], "accept");
    assert!(settings.get("permissions").is_none());
}

/// AC-4: a budget of `0` or an unusable budget or turn count refuses to build before any
/// process starts.
#[test]
fn zero_or_missing_budget_refuses_to_build() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let env = standin_env(&standin_dir(tmp.path()));

    let mut spec = audit_spec();
    spec.max_budget_usd = "0".to_string();
    assert!(build(&spec, &ctx(&root), &env).is_err());

    spec.max_budget_usd = "0.00".to_string();
    assert!(build(&spec, &ctx(&root), &env).is_err());

    spec.max_budget_usd = "five".to_string();
    assert!(build(&spec, &ctx(&root), &env).is_err());

    spec.max_budget_usd = "5.00".to_string();
    spec.max_turns = 0;
    assert!(build(&spec, &ctx(&root), &env).is_err());
}

/// AC-9: an audit run whose requesting session is the author session, or that names no
/// author, refuses to build; a different author builds, and non-audit roles need no author.
#[test]
fn author_session_audit_refuses_to_build() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let env = standin_env(&standin_dir(tmp.path()));

    let mut self_audit = audit_spec();
    self_audit.role = RunRole::Audit {
        author_session: REQUESTING.to_string(),
    };
    let refused = build(&self_audit, &ctx(&root), &env).unwrap_err();
    assert!(refused.to_string().contains("author"), "{refused}");

    let mut no_author = audit_spec();
    no_author.role = RunRole::Audit {
        author_session: String::new(),
    };
    assert!(build(&no_author, &ctx(&root), &env).is_err());

    let ok = build(&audit_spec(), &ctx(&root), &env).unwrap();
    assert_eq!(ok.requesting_session, REQUESTING);

    let mut worker = audit_spec();
    worker.role = RunRole::Worker;
    assert!(build(&worker, &ctx(&root), &env).is_ok());
}

/// AC-7: the warning names the key's presence, never its value, and appears only when the
/// key is present.
#[test]
fn api_key_warning_names_presence_only() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let fixtures = standin_dir(tmp.path());
    let mut with_key = standin_env(&fixtures);
    with_key.insert("ANTHROPIC_API_KEY".to_string(), SECRET.to_string());
    let built = build(&audit_spec(), &ctx(&root), &with_key).unwrap();
    let warning = built.credential_warning.expect("warning present");
    assert!(warning.contains("ANTHROPIC_API_KEY"));
    assert!(!warning.contains(SECRET));

    let no_key = standin_env(&fixtures);
    let built = build(&audit_spec(), &ctx(&root), &no_key).unwrap();
    assert!(built.credential_warning.is_none());
}

/// REQ-2: blanket tool rules refuse to build.
#[test]
fn blanket_tool_rules_refuse_to_build() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let env = standin_env(&standin_dir(tmp.path()));
    for blanket in ["Bash", "Bash(*)", "PowerShell"] {
        let mut spec = audit_spec();
        spec.allowed_tools = vec![blanket.to_string()];
        assert!(
            build(&spec, &ctx(&root), &env).is_err(),
            "blanket rule {blanket} must refuse"
        );
    }
}

/// REQ-5 + REQ-1: an audit run needs a schema and an agent, and the foreign-tree audit run
/// replaces the setting sources with `user` only (REQ-9).
#[test]
fn audit_needs_schema_and_foreign_tree_narrows_sources() {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();
    let env = standin_env(&standin_dir(tmp.path()));

    let mut no_schema = audit_spec();
    no_schema.schema = None;
    assert!(build(&no_schema, &ctx(&root), &env).is_err());

    let mut no_agent = audit_spec();
    no_agent.agent = None;
    assert!(build(&no_agent, &ctx(&root), &env).is_err());

    let mut foreign = audit_spec();
    foreign.foreign_tree = true;
    let built = build(&foreign, &ctx(&root), &env).unwrap();
    assert_eq!(after(&built.args, "--setting-sources"), "user");
}

/// AC-5: a `success` subtype with no `structured_output` is no verdict, and so are a
/// refusal `stop_reason`, `is_error` true, a non-zero exit and unparseable output.
#[test]
fn success_without_structured_output_is_no_verdict() {
    use moai_cli::headless_run::{parse_result, RawRun, RunOutcome};

    fn raw(success: bool, code: i32, stdout: &str) -> RawRun {
        RawRun {
            success,
            code: Some(code),
            stdout: stdout.to_string(),
            stderr: String::new(),
        }
    }

    let cases = [
        raw(
            true,
            0,
            r#"{"type":"result","subtype":"success","is_error":false,"session_id":"s","total_cost_usd":0.1,"num_turns":2}"#,
        ),
        raw(
            true,
            0,
            r#"{"type":"result","subtype":"success","is_error":false,"stop_reason":"refusal","structured_output":{"verdict":"pass"}}"#,
        ),
        raw(
            true,
            0,
            r#"{"type":"result","subtype":"success","is_error":true,"structured_output":{"verdict":"pass"}}"#,
        ),
        raw(false, 2, ""),
        raw(true, 0, "not json"),
    ];
    for (i, case) in cases.iter().enumerate() {
        assert!(
            matches!(parse_result(case), RunOutcome::NoVerdict { .. }),
            "case {i} must be no verdict"
        );
    }

    let answered = raw(
        true,
        0,
        r#"{"type":"result","subtype":"success","is_error":false,"structured_output":{"verdict":"fail","blocking":["B-1"]},"session_id":"s","total_cost_usd":0.1,"num_turns":3}"#,
    );
    match parse_result(&answered) {
        RunOutcome::Answered(v) => assert_eq!(v["verdict"], "fail"),
        other => panic!("expected Answered, got {other:?}"),
    }
}

/// AC-6: one finished run records exactly one `claude.run` event row whose payload names
/// the verdict, the session id, the cost estimate marker, the turns, the terminal reason
/// and the permission denials.
#[test]
fn one_run_one_ledger_row() {
    use moai_cli::headless_run::{ledger_row, parse_result, record_run, RUN_EVENT};
    use moai_cli::ledger::schema;
    use rusqlite::Connection;

    let tmp = tempfile::tempdir().unwrap();
    let fixtures = tmp.path().join("bin");
    fs::create_dir_all(&fixtures).unwrap();
    let record = tmp.path().join("record.txt");
    standin_claude(&fixtures);
    let root = tmp.path().join("platform");
    fs::create_dir_all(&root).unwrap();

    let built = build(&audit_spec(), &ctx(&root), &parent_env(&fixtures, &record)).unwrap();
    let raw = run(&built, tmp.path()).unwrap();
    let outcome = parse_result(&raw);
    let row = ledger_row(&built, &raw, &outcome);

    assert_eq!(row["session_id"], "standin");
    assert_eq!(row["cost_is_estimate"], true);
    assert!(row["total_cost_usd"].is_number());
    assert!(row["num_turns"].is_number());
    assert!(row["terminal_reason"].is_object());
    assert!(row.get("permission_denials").is_some());
    assert_eq!(row["requesting_session"], REQUESTING);
    assert!(row["scrubbed_credentials"].is_array());

    let mut conn = Connection::open_in_memory().unwrap();
    schema::create_v1(&mut conn, "demo", "test").unwrap();
    let key = record_run(&mut conn, None, &row, "tester").unwrap();
    assert!(key.starts_with("claude.run:"), "{key}");
    let (count, payload): (i64, String) = conn
        .query_row(
            "SELECT COUNT(*), payload FROM event WHERE kind = ?1",
            [RUN_EVENT],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    assert_eq!(count, 1);
    let payload: serde_json::Value = serde_json::from_str(&payload).unwrap();
    assert_eq!(payload["session_id"], "standin");
}
