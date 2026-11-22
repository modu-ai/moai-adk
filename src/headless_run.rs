//! Headless Claude runs (SPEC-20261122-headless-run-contract): every `claude -p` command moai
//! starts is built here, in one function every caller goes through. The builder pins the
//! permission mode, the permission prompts, the allowed-tool rules, the MCP flags, the model,
//! the agent, the output format with its schema, the turn and budget caps, the setting sources
//! and the one `--settings` object; it refuses an audit run requested by the change's author
//! session (the author-never-audits rule), keeps credential values out of the child argv and
//! environment, warns once when `ANTHROPIC_API_KEY` is present in the parent environment, and
//! resolves `claude` from the child PATH so a test stand-in can take its place (SPEC D-5).

use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The credential-carrying variables scrubbed from the child environment (SPEC D-10, audit
/// N-5: not only the Anthropic three — git, GitHub and the cloud providers' tokens stay out
/// too); a caller that needs one for a gateway run passes it as `gateway_credential` and the
/// runner sets it again after the scrub.
const CREDENTIAL_ENVS: [&str; 12] = [
    "ANTHROPIC_API_KEY",
    "ANTHROPIC_AUTH_TOKEN",
    "CLAUDE_CODE_OAUTH_TOKEN",
    "GH_TOKEN",
    "GITHUB_TOKEN",
    "AWS_ACCESS_KEY_ID",
    "AWS_SECRET_ACCESS_KEY",
    "AWS_SESSION_TOKEN",
    "AWS_BEARER_TOKEN_BEDROCK",
    "AWS_WEB_IDENTITY_TOKEN_FILE",
    "GOOGLE_APPLICATION_CREDENTIALS",
    "GOOGLE_OAUTH_ACCESS_TOKEN",
];

/// The session-identity variables of the *starting* session, removed so the child is its own
/// session and never a nested one; the requesting session is recorded on the built run instead.
const SESSION_ENVS: [&str; 3] = [
    "CLAUDECODE",
    "CLAUDE_CODE_CHILD_SESSION",
    "CLAUDE_CODE_SESSION_ID",
];

/// `BASH_DEFAULT_TIMEOUT_MS` the child gets: at least ten minutes, or the parent's own value
/// when it asked for longer (SPEC D-10; the default is 120000 and headless.md raises it because
/// a `-p` run can exit before a long subagent command finishes).
const BASH_TIMEOUT_FLOOR_MS: u64 = 600_000;
const BASH_TIMEOUT_DEFAULT_MS: u64 = 120_000;

/// What the run is for (SPEC D-1): an audit names the change's author session; a worker answers
/// wake messages; a lane run is work in a lane and is the only role that turns prompt-denying
/// on (the user's pick 3, through `--settings`, never a settings file).
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum RunRole {
    Audit { author_session: String },
    Worker,
    Lane,
}

/// One headless run a caller asks for, with every value the builder pins (SPEC REQ-1).
#[derive(Debug, Clone)]
pub struct HeadlessRunSpec {
    pub role: RunRole,
    pub model: String,
    pub agent: Option<String>,
    pub schema: Option<Value>,
    pub prompt: String,
    pub max_turns: u32,
    pub max_budget_usd: String,
    pub allowed_tools: Vec<String>,
    pub foreign_tree: bool,
    pub gateway_credential: Option<(String, String)>,
}

/// The session asking for the run and, for tests, a platform root standing in for the real
/// managed-MCP path.
#[derive(Debug, Clone)]
pub struct RunContext {
    pub requesting_session: String,
    pub platform_root: Option<PathBuf>,
}

/// A built but not yet started run: the resolved `claude`, the exact argv after it, the exact
/// child environment, and what the run records.
#[derive(Debug, Clone)]
pub struct BuiltRun {
    pub program: PathBuf,
    pub args: Vec<String>,
    pub env: BTreeMap<String, String>,
    pub permission_mode: String,
    pub mode_reason: String,
    pub requesting_session: String,
    pub managed_mcp_deployed: bool,
    pub scrubbed: Vec<String>,
    pub credential_warning: Option<String>,
}

/// The raw outcome of one started run (parsed in the next slice).
#[derive(Debug, Clone)]
pub struct RawRun {
    pub success: bool,
    pub code: Option<i32>,
    pub stdout: String,
    pub stderr: String,
}

/// Builds the pinned `claude -p` command (SPEC REQ-1 to REQ-10). Every refusal names its reason
/// and nothing is spawned.
pub fn build(
    spec: &HeadlessRunSpec,
    ctx: &RunContext,
    parent_env: &BTreeMap<String, String>,
) -> Result<BuiltRun> {
    // N-8: the requesting session comes from the caller or, when blank, from the
    // environment's CLAUDE_CODE_SESSION_ID; the author comparison ignores case.
    let requesting = match ctx.requesting_session.trim() {
        "" => parent_env
            .get("CLAUDE_CODE_SESSION_ID")
            .map(|s| s.trim().to_string())
            .unwrap_or_default(),
        given => given.to_string(),
    };
    if requesting.is_empty() {
        return Err("headless run without a requesting session".to_string());
    }

    // The author-never-audits rule (SPEC REQ-1, D-9): fail closed on both ends.
    let (mode, mode_reason) = match &spec.role {
        RunRole::Audit { author_session } => {
            let author = author_session.trim();
            if author.is_empty() {
                return Err("audit run names no author session".to_string());
            }
            if author.eq_ignore_ascii_case(&requesting) {
                return Err(format!(
                    "audit run requested by the change's author session {author}: the author never audits"
                ));
            }
            ("dontAsk", "audit: read-only, pre-approved tools only")
        }
        RunRole::Worker => ("dontAsk", "worker: pre-approved tools only"),
        RunRole::Lane => ("auto", "lane work: classifier-reviewed"),
    };

    if spec.model.trim().is_empty() {
        return Err("headless run without a pinned model".to_string());
    }
    if matches!(spec.role, RunRole::Audit { .. }) {
        if spec.schema.is_none() {
            return Err("audit run without a --json-schema verdict schema".to_string());
        }
        if spec.agent.as_deref().unwrap_or("").trim().is_empty() {
            return Err("audit run without a pinned --agent".to_string());
        }
    }
    for tool in &spec.allowed_tools {
        narrow_rule(tool)?;
    }
    let budget = spec
        .max_budget_usd
        .trim()
        .parse::<f64>()
        .map_err(|_| format!("unusable --max-budget-usd {:?}", spec.max_budget_usd))?;
    if budget <= 0.0 || budget.is_nan() {
        return Err(format!(
            "--max-budget-usd must be positive, got {:?}",
            spec.max_budget_usd
        ));
    }
    if spec.max_turns == 0 {
        return Err("--max-turns must be positive".to_string());
    }

    // The managed-MCP adaptation (SPEC REQ-3, D-2): with the platform file deployed both MCP
    // flags make Claude Code exit at startup, so the runner passes neither; detection is by
    // existence, which fails safe when a file is half-deployed.
    let deployed = managed_mcp_deployed(ctx.platform_root.as_deref());

    let mut args: Vec<String> = vec!["-p".into()];
    args.push("--permission-mode".into());
    args.push(mode.into());
    args.push("--permission-prompts".into());
    args.push("none".into());
    if !spec.allowed_tools.is_empty() {
        args.push("--allowedTools".into());
        args.push(spec.allowed_tools.join(","));
    }
    if !deployed {
        args.push("--strict-mcp-config".into());
    }
    args.push("--model".into());
    args.push(spec.model.trim().to_string());
    if let Some(agent) = spec.agent.as_deref().filter(|a| !a.trim().is_empty()) {
        args.push("--agent".into());
        args.push(agent.trim().to_string());
    }
    args.push("--output-format".into());
    args.push("json".into());
    if let Some(schema) = &spec.schema {
        args.push("--json-schema".into());
        args.push(schema.to_string());
    }
    args.push("--max-turns".into());
    args.push(spec.max_turns.to_string());
    args.push("--max-budget-usd".into());
    args.push(spec.max_budget_usd.trim().to_string());
    args.push("--setting-sources".into());
    args.push(
        if spec.foreign_tree {
            "user"
        } else {
            "user,project,local"
        }
        .into(),
    );

    // One `--settings` object, role-scoped (SPEC REQ-8, REQ-9, D-3): every worker accepts wake
    // messages; only a lane run denies-instead-of-prompting, and never through a settings file.
    let mut settings = json!({"crossSessionInbound": "accept"});
    if matches!(spec.role, RunRole::Lane) {
        settings["permissions"] = json!({"denyInsteadOfPrompting": true});
    }
    args.push("--settings".into());
    args.push(settings.to_string());

    // The child environment (SPEC REQ-10, D-10): inherit the parent, scrub credentials and the
    // starting session's identity, set the unattended-run variables, then set a caller-named
    // gateway credential back.
    let mut env = parent_env.clone();
    let mut scrubbed = Vec::new();
    for name in CREDENTIAL_ENVS {
        if env.remove(name).is_some() {
            scrubbed.push(name.to_string());
        }
    }
    for name in SESSION_ENVS {
        env.remove(name);
    }
    env.insert("CLAUDE_CODE_RETRY_WATCHDOG".into(), "1".into());
    let bash_ms = parent_env
        .get("BASH_DEFAULT_TIMEOUT_MS")
        .and_then(|v| v.trim().parse::<u64>().ok())
        .unwrap_or(BASH_TIMEOUT_DEFAULT_MS)
        .max(BASH_TIMEOUT_FLOOR_MS);
    env.insert("BASH_DEFAULT_TIMEOUT_MS".into(), bash_ms.to_string());
    if let Some((name, value)) = &spec.gateway_credential {
        env.insert(name.clone(), value.clone());
    }
    let credential_warning = parent_env.contains_key("ANTHROPIC_API_KEY").then(|| {
        // N-6: a caller-named gateway credential may carry the key back in — the warning then
        // says so instead of claiming the child is free of it.
        if matches!(&spec.gateway_credential, Some((name, _)) if name == "ANTHROPIC_API_KEY") {
            "moai: ANTHROPIC_API_KEY is set in the parent environment and the caller-named \
             gateway credential carries it into this child; interactive sessions still bill \
             to that key"
        } else {
            "moai: ANTHROPIC_API_KEY is set in the parent environment; this headless run keeps \
             it out of the child, and interactive sessions still bill to that key"
        }
        .to_string()
    });

    // The prompt is the positional argument, last, behind `--` so a dash-leading prompt is
    // never read as a flag (N-7).
    args.push("--".into());
    args.push(spec.prompt.clone());

    let program = resolve_program(&env)?;
    Ok(BuiltRun {
        program,
        args,
        env,
        permission_mode: mode.to_string(),
        mode_reason: mode_reason.to_string(),
        requesting_session: requesting,
        managed_mcp_deployed: deployed,
        scrubbed,
        credential_warning,
    })
}

/// Starts a built run with the default wall-clock limit of thirty minutes (N-2).
pub fn run(built: &BuiltRun, cwd: &Path) -> Result<RawRun> {
    run_with_limit(built, cwd, std::time::Duration::from_secs(1800))
}

/// Starts a built run under a wall-clock limit (N-2): prints the credential warning once,
/// spawns the resolved program with the exact argv and environment, and kills the child at
/// the deadline, reporting no success.
pub fn run_with_limit(built: &BuiltRun, cwd: &Path, limit: std::time::Duration) -> Result<RawRun> {
    use std::process::Stdio;
    if let Some(warning) = &built.credential_warning {
        eprintln!("{warning}");
    }
    let mut child = std::process::Command::new(&built.program)
        .args(&built.args)
        .env_clear()
        .envs(&built.env)
        .current_dir(cwd)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("starting {}: {e}", built.program.display()))?;
    let deadline = std::time::Instant::now() + limit;
    let mut timed_out = false;
    loop {
        match child.try_wait() {
            Ok(Some(_)) => break,
            Ok(None) => {
                if std::time::Instant::now() >= deadline {
                    let _ = child.kill();
                    timed_out = true;
                    break;
                }
                std::thread::sleep(std::time::Duration::from_millis(50));
            }
            Err(e) => return Err(format!("waiting for {}: {e}", built.program.display())),
        }
    }
    let output = child
        .wait_with_output()
        .map_err(|e| format!("collecting {}: {e}", built.program.display()))?;
    let stderr = if timed_out {
        format!(
            "{}\nmoai: wall-clock limit of {} seconds exceeded; the child was killed",
            String::from_utf8_lossy(&output.stderr),
            limit.as_secs()
        )
    } else {
        String::from_utf8_lossy(&output.stderr).into_owned()
    };
    Ok(RawRun {
        success: output.status.success() && !timed_out,
        code: output.status.code(),
        stdout: String::from_utf8_lossy(&output.stdout).into_owned(),
        stderr,
    })
}

/// What one run's result is, checked before it is trusted (SPEC REQ-11): a `success` subtype
/// alone is not an answer.
#[derive(Debug, Clone, PartialEq)]
pub enum RunOutcome {
    Answered(Value),
    NoVerdict { reason: String },
}

/// The ledger event kind of one recorded run (SPEC REQ-12, D-7).
pub const RUN_EVENT: &str = "claude.run";

/// Checks one raw outcome before it is trusted (SPEC REQ-11): the process must have exited
/// zero, the JSON result must carry `is_error` false and the `structured_output` field; a
/// refusal `stop_reason` and a missing field are both no verdict.
pub fn parse_result(raw: &RawRun) -> RunOutcome {
    let no = |reason: &str| RunOutcome::NoVerdict {
        reason: reason.to_string(),
    };
    if !raw.success {
        return no(&format!(
            "exit code {}",
            raw.code
                .map(|c| c.to_string())
                .unwrap_or_else(|| "unknown".into())
        ));
    }
    let Ok(value) = serde_json::from_str::<Value>(&raw.stdout) else {
        return no("stdout is not one JSON object");
    };
    // N-4: only the success subtype can carry an answer.
    if value.get("subtype").and_then(Value::as_str) != Some("success") {
        return no("the result subtype is not success");
    }
    if value.get("is_error").and_then(Value::as_bool) != Some(false) {
        return no("the result carries is_error true or no is_error field");
    }
    if value.get("stop_reason").and_then(Value::as_str) == Some("refusal") {
        return no("the model refused the request");
    }
    match value.get("structured_output") {
        Some(answer) if !answer.is_null() => RunOutcome::Answered(answer.clone()),
        Some(_) => no("structured_output is null"),
        None => no("a success subtype without structured_output"),
    }
}

/// Checks the verdict field of an answered run against the schema that asked for it (N-1):
/// the field must exist, carry the schema's type and, when the schema names an enum, be one
/// of its words. Returns the verdict string; the message never carries other values.
pub fn check_verdict(answer: &Value, schema: &Value) -> Result<String> {
    let verdict = answer
        .get("verdict")
        .ok_or_else(|| "the answer carries no verdict field".to_string())?;
    let field = schema
        .get("properties")
        .and_then(|p| p.get("verdict"))
        .ok_or_else(|| "the schema names no verdict property".to_string())?;
    if let Some(words) = field.get("enum").and_then(Value::as_array) {
        let held = verdict.as_str().unwrap_or("");
        if !words.iter().any(|w| w.as_str() == Some(held)) {
            return Err("the verdict is not one of the schema's words".to_string());
        }
    }
    match (field.get("type").and_then(Value::as_str), verdict) {
        (Some("string"), Value::String(s)) => Ok(s.clone()),
        (Some("string"), _) => Err("the verdict field must be a string".to_string()),
        (_, Value::String(s)) => Ok(s.clone()),
        (_, other) => Err(format!(
            "the verdict field must be a scalar, not {}",
            json_kind(other)
        )),
    }
}

fn json_kind(value: &Value) -> &'static str {
    match value {
        Value::Null => "null",
        Value::Bool(_) => "a boolean",
        Value::Number(_) => "a number",
        Value::String(_) => "a string",
        Value::Array(_) => "an array",
        Value::Object(_) => "an object",
    }
}

/// One `claude.run` ledger payload (SPEC REQ-12): the verdict, the session id, the cost as a
/// marked estimate, the turns, the terminal reason and the permission denials.
pub fn ledger_row(built: &BuiltRun, raw: &RawRun, outcome: &RunOutcome) -> Value {
    let result: Value = serde_json::from_str(&raw.stdout).unwrap_or(Value::Null);
    let verdict = match outcome {
        RunOutcome::Answered(answer) => answer
            .get("verdict")
            .cloned()
            .unwrap_or_else(|| Value::String("answered".into())),
        RunOutcome::NoVerdict { reason } => json!({"no_verdict": reason}),
    };
    let terminal_reason = json!({
        "subtype": result.get("subtype").cloned().unwrap_or(Value::Null),
        "stop_reason": result.get("stop_reason").cloned().unwrap_or(Value::Null),
        "exit_code": raw.code,
        "success": raw.success,
    });
    json!({
        "verdict": verdict,
        "session_id": result.get("session_id").cloned().unwrap_or(Value::Null),
        "total_cost_usd": result.get("total_cost_usd").cloned().unwrap_or(Value::Null),
        "cost_is_estimate": true,
        "num_turns": result.get("num_turns").cloned().unwrap_or(Value::Null),
        "terminal_reason": terminal_reason,
        "permission_denials": result
            .get("permission_denials")
            .cloned()
            .unwrap_or_else(|| json!([])),
        "requesting_session": built.requesting_session,
        "permission_mode": built.permission_mode,
        "mode_reason": built.mode_reason,
        "managed_mcp_deployed": built.managed_mcp_deployed,
        "scrubbed_credentials": built.scrubbed,
    })
}

/// Records one run as a `claude.run` event row in one transaction (SPEC REQ-12): the row
/// fits the existing free-text event table with no schema change.
pub fn record_run(
    conn: &mut rusqlite::Connection,
    card_id: Option<i64>,
    row: &Value,
    by: &str,
) -> Result<String> {
    let tx = conn
        .transaction()
        .map_err(|e| format!("record_run transaction: {e}"))?;
    let key = crate::ledger::open::new_event_key(&tx, RUN_EVENT, card_id)
        .map_err(|e| format!("record_run key: {e}"))?;
    tx.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, unixepoch(), ?2, ?3, ?4, ?5)",
        rusqlite::params![key, by, card_id, RUN_EVENT, row.to_string()],
    )
    .map_err(|e| format!("record_run insert: {e}"))?;
    tx.commit().map_err(|e| format!("record_run commit: {e}"))?;
    Ok(key)
}

/// A tool rule the builder accepts: never a blanket `Bash`·`PowerShell`·`Agent`·`Monitor`,
/// never a wildcard-only scope in any of its spellings (`*`, `:*`, empty), never a
/// wildcarded interpreter or wrapper (`python *`, `sh -c *`, `timeout *`) — every shape auto
/// mode drops (SPEC REQ-2, audit N-3).
fn narrow_rule(tool: &str) -> Result<()> {
    let blanket = matches!(tool, "Bash" | "PowerShell" | "Agent" | "Monitor");
    let scope = tool
        .strip_prefix("Bash(")
        .or_else(|| tool.strip_prefix("PowerShell("))
        .and_then(|inner| inner.strip_suffix(')'));
    let wildcard_only = scope.is_some_and(|inner| {
        let inner = inner.trim();
        inner.is_empty() || inner.starts_with('*') || inner == ":*"
    });
    let wildcarded_word = scope.is_some_and(|inner| {
        if !inner.contains('*') {
            return false;
        }
        inner
            .split_whitespace()
            .next()
            .is_some_and(broad_first_word)
    });
    if blanket || wildcard_only || wildcarded_word {
        return Err(format!(
            "blanket tool rule {tool:?} is not allowed; use a narrow rule"
        ));
    }
    Ok(())
}

/// Whether a command rule's first word turns a wildcarded rule into a blanket one: the
/// interpreters and wrappers that run what follows them, in every spelling — a versioned
/// interpreter (python3.12) and the `:*` colon suffix included (permission-modes.md's
/// dropped-rules list, audit R-1 and R-2).
fn broad_first_word(word: &str) -> bool {
    let word = word.split(':').next().unwrap_or(word);
    if INTERPRETERS_AND_WRAPPERS.contains(&word) {
        return true;
    }
    // a versioned python: python3, python3.12, pypy-style digits and dots after the name
    word.strip_prefix("python")
        .is_some_and(|rest| rest.chars().all(|c| c.is_ascii_digit() || c == '.'))
}

/// The first words that run what follows them (permission-modes.md's dropped-rules list plus
/// the env/xargs/sudo/npx escapes of audit R-2).
const INTERPRETERS_AND_WRAPPERS: [&str; 19] = [
    "python",
    "python3",
    "node",
    "deno",
    "ruby",
    "perl",
    "sh",
    "bash",
    "zsh",
    "fish",
    "osascript",
    "timeout",
    "time",
    "nice",
    "nohup",
    "env",
    "xargs",
    "sudo",
    "npx",
];

/// Whether a platform `managed-mcp.json` is deployed (SPEC REQ-3, D-2). A test passes its own
/// platform root; production reads the real per-OS directory.
fn managed_mcp_deployed(root: Option<&Path>) -> bool {
    let dir = match root {
        Some(dir) => dir.to_path_buf(),
        None => platform_dir(),
    };
    dir.join("managed-mcp.json").exists()
}

/// The per-OS directory a deployed `managed-mcp.json` lives in (managed-mcp.md).
fn platform_dir() -> PathBuf {
    if cfg!(target_os = "macos") {
        PathBuf::from("/Library/Application Support/ClaudeCode")
    } else if cfg!(target_os = "windows") {
        PathBuf::from(r"C:\Program Files\ClaudeCode")
    } else {
        PathBuf::from("/etc/claude-code")
    }
}

/// Resolves `claude` from the PATH of the child environment, so a stand-in first on that PATH
/// takes the program's place (SPEC D-5) and the spawn never depends on the parent's PATH. On
/// Windows the PATH separator is `;` and the program carries an `.exe` suffix (N-10).
fn resolve_program(env: &BTreeMap<String, String>) -> Result<PathBuf> {
    let path = env
        .get("PATH")
        .ok_or_else(|| "child environment has no PATH".to_string())?;
    let program = if cfg!(windows) {
        "claude.exe"
    } else {
        "claude"
    };
    let separator = if cfg!(windows) { ';' } else { ':' };
    for dir in path.split(separator) {
        if dir.is_empty() {
            continue;
        }
        let candidate = Path::new(dir).join(program);
        if is_executable(&candidate) {
            return Ok(candidate);
        }
    }
    Err("no executable claude on the child PATH".to_string())
}

#[cfg(unix)]
fn is_executable(path: &Path) -> bool {
    use std::os::unix::fs::PermissionsExt;
    path.is_file()
        && std::fs::metadata(path)
            .map(|m| m.permissions().mode() & 0o111 != 0)
            .unwrap_or(false)
}

#[cfg(not(unix))]
fn is_executable(path: &Path) -> bool {
    path.is_file()
}
