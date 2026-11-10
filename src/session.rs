//! Development launcher: starts a leader or agent session on a host CLI with the
//! model/effort/advisor options the operator decided (2026-11-08). The launcher composes
//! argv from documented flags only and never invents options.
//!
//! Hosts (00-F #2028; sources read 2026-11-08):
//! - `claude`: Claude Code. `--name`, `--model`, `--effort`, `--dangerously-skip-permissions`
//!   from https://code.claude.com/docs/en/cli-reference; `--advisor <model>` from
//!   https://code.claude.com/docs/en/advisor (alias `fable`, `opus`, `sonnet` or a full id,
//!   must rank at or above the main model, disabled when `CLAUDE_CODE_DISABLE_ADVISOR_TOOL=1`).
//! - `glm`: Claude Code pointed at z.ai through process environment only, per
//!   https://docs.z.ai/scenario-example/develop-tools/claude. The token is inherited from the
//!   caller and only its presence is checked; its value is never kept, printed or written.
//! - `codex`: Codex CLI `-s workspace-write -a on-request [-m <model>] [PROMPT]`, per
//!   https://learn.chatgpt.com/docs/developer-commands?surface=cli. Codex has no session-name
//!   flag, so the name travels in the first prompt.
//!
//! Without a role (`moai-cli cc|glm|codex` with neither `-l` nor `-a`, Issue #2053) the launcher
//! opens a plain host session: no role is guessed, no session name is added, and
//! `--dangerously-skip-permissions` is never added. `-l|-a` argv are unchanged.
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

const ROLES: [&str; 2] = ["leader", "agent"];
const HOSTS: [&str; 3] = ["claude", "glm", "codex"];
const EFFORTS: [&str; 6] = ["low", "medium", "high", "xhigh", "max", "ultracode"];
const ADVISOR_ALIASES: [&str; 3] = ["fable", "opus", "sonnet"];
const DEFAULT_MODEL: &str = "opus";
const DEFAULT_EFFORT: &str = "medium";
/// User decision 2026-11-08 (#2053 HOLD-L1 = A, #2062): a session without a role still gets the
/// launcher's Claude defaults (`--model opus --effort medium --advisor fable`), never a
/// permission skip. False would pass no model, effort or advisor flag so the user's own Claude
/// settings apply (option B, not chosen). This constant is the only place to change.
pub const ROLELESS_LAUNCHER_DEFAULTS: bool = true;
/// The environment that points Claude Code at z.ai GLM. The set and its values follow the Go
/// tool's `setGLMEnv` (user decision 2026-11-10): endpoint, timeout, the model each Claude
/// alias maps to, the 1M context window of those models (Claude Code assumes 200K for an id
/// it does not know) and the switch that strips Anthropic beta headers the proxy rejects.
pub const GLM_ENV: [(&str, &str); 9] = [
    ("ANTHROPIC_BASE_URL", "https://api.z.ai/api/anthropic"),
    ("API_TIMEOUT_MS", "3000000"),
    ("ANTHROPIC_DEFAULT_OPUS_MODEL", "glm-5.3"),
    ("ANTHROPIC_DEFAULT_SONNET_MODEL", "glm-5.3"),
    ("ANTHROPIC_DEFAULT_HAIKU_MODEL", "glm-5.3-flash"),
    ("ANTHROPIC_DEFAULT_FABLE_MODEL", "glm-5.3"),
    ("CLAUDE_CODE_AUTO_COMPACT_WINDOW", "1000000"),
    ("CLAUDE_CODE_MAX_CONTEXT_TOKENS", "1000000"),
    ("CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS", "1"),
];
/// The z.ai MCP server reads the same key under this name; the launcher passes it along.
pub const GLM_MCP_KEY_VAR: &str = "Z_AI_API_KEY";
/// Taken from the caller's environment, else from a key file (#2359). The value is passed to
/// the host process only; it is never printed, logged or written.
pub const GLM_TOKEN_VAR: &str = "ANTHROPIC_AUTH_TOKEN";
/// Key files under HOME, read in this order when the variable is absent. The first is the
/// global store shared with the Go tool (user decision 2026-11-10). Both are read, never
/// created or changed.
pub const GLM_KEY_FILES: [&str; 2] = [".moai/.env.glm", ".moai-cli/.env.glm"];
/// Names a key file may use for the token.
const GLM_KEY_NAMES: [&str; 2] = [GLM_TOKEN_VAR, "GLM_API_KEY"];
/// Largest key file that is read; a bigger one is skipped.
const GLM_KEY_FILE_MAX: u64 = 65536;

/// The token from the first key file that holds a non-empty one. Only a regular file within
/// the size limit is opened, so a FIFO or a device at the path cannot block the launcher.
fn glm_file_token() -> Option<String> {
    let home = std::path::PathBuf::from(std::env::var_os("HOME")?);
    GLM_KEY_FILES.iter().find_map(|file| {
        let path = home.join(file);
        let meta = std::fs::metadata(&path).ok()?;
        if !meta.is_file() || meta.len() > GLM_KEY_FILE_MAX {
            return None;
        }
        let text = std::fs::read_to_string(path).ok()?;
        text.lines().find_map(|line| {
            let line = line.trim();
            let line = line.strip_prefix("export ").unwrap_or(line);
            let (key, value) = line.split_once('=')?;
            if !GLM_KEY_NAMES.contains(&key.trim()) {
                return None;
            }
            let value = value.trim().trim_matches(|c| c == '"' || c == '\'');
            (!value.is_empty() && !value.chars().any(char::is_control)).then(|| value.to_string())
        })
    })
}
const MAX_PROMPT: usize = 4096;

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct SessionRequest {
    /// `leader` or `agent`; `None` is a plain host session (#2053).
    #[serde(default)]
    pub role: Option<String>,
    pub host: String,
    /// `None` adds no session name (no `--name`, no name prompt).
    #[serde(default)]
    pub name: Option<String>,
    /// `None` uses the host default: `opus` on Claude hosts, Codex's own default on codex.
    #[serde(default)]
    pub model: Option<String>,
    /// `None` uses `medium` on Claude hosts; Codex has no effort flag here.
    #[serde(default)]
    pub effort: Option<String>,
    #[serde(default)]
    pub advisor: Option<String>,
    /// Adds `--dangerously-skip-permissions` on Claude hosts (operator's interim default).
    #[serde(default)]
    pub skip_permissions: bool,
    /// Extra first-prompt text (codex only).
    #[serde(default)]
    pub prompt: Option<String>,
    #[serde(default = "default_claude_bin")]
    pub claude_bin: String,
    #[serde(default = "default_codex_bin")]
    pub codex_bin: String,
}

fn default_claude_bin() -> String {
    "claude".into()
}
fn default_codex_bin() -> String {
    "codex".into()
}

fn rank(model: &str) -> Option<u8> {
    let m = model.to_ascii_lowercase();
    ["haiku", "sonnet", "opus", "fable"]
        .iter()
        .position(|k| m.contains(k))
        .map(|p| p as u8)
}
fn clean(field: &str, value: &str) -> Result<(), String> {
    if value.is_empty() || value.len() > 128 || value.chars().any(|c| c.is_control() || c == ' ') {
        return Err(format!("invalid {field}"));
    }
    Ok(())
}

/// Whether the launcher's model/effort defaults apply: always with a role, and without one
/// while [`ROLELESS_LAUNCHER_DEFAULTS`] is true (user decision 2026-11-08: A).
pub fn launcher_defaults(role: Option<&str>) -> bool {
    role.is_some() || ROLELESS_LAUNCHER_DEFAULTS
}

/// What the project's `.claude/settings.json` says about a `permissions` key (#2244).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PermissionsKey {
    /// A top-level `permissions` key exists (any value): the project owns its permissions.
    Present,
    /// No file, or a JSON object without the key.
    Absent,
    /// The file exists but is not a bounded regular file holding a JSON object.
    Unreadable,
}

/// Reads `<dir>/.claude/settings.json` with the settings reader's bounds (symlink refused,
/// 256 KiB). Only `settings.json` is read; `settings.local.json` is not (#2244 gap).
pub fn permissions_key(dir: &std::path::Path) -> PermissionsKey {
    match crate::claude::read_existing(&dir.join(".claude/settings.json")) {
        Ok(None) => PermissionsKey::Absent,
        Ok(Some(bytes)) => match serde_json::from_slice::<Value>(&bytes) {
            Ok(Value::Object(m)) if m.contains_key("permissions") => PermissionsKey::Present,
            Ok(Value::Object(_)) => PermissionsKey::Absent,
            _ => PermissionsKey::Unreadable,
        },
        Err(_) => PermissionsKey::Unreadable,
    }
}

fn claude_argv(r: &SessionRequest, notes: &mut Vec<&'static str>) -> Result<Vec<String>, String> {
    let defaults = launcher_defaults(r.role.as_deref());
    let model = r.model.as_deref().or(defaults.then_some(DEFAULT_MODEL));
    let effort = r.effort.as_deref().or(defaults.then_some(DEFAULT_EFFORT));
    if let Some(e) = effort {
        if !EFFORTS.contains(&e) {
            return Err("effort must be one of low|medium|high|xhigh|max|ultracode".into());
        }
    }
    if let Some(m) = model {
        clean("model", m)?;
    }
    clean("claude_bin", &r.claude_bin)?;
    let mut argv = vec![r.claude_bin.clone()];
    if let Some(n) = &r.name {
        argv.push("--name".into());
        argv.push(n.clone());
    }
    if let Some(m) = model {
        argv.push("--model".into());
        argv.push(m.into());
    }
    if let Some(e) = effort {
        argv.push("--effort".into());
        argv.push(e.into());
    }
    if let Some(a) = &r.advisor {
        clean("advisor", a)?;
        if !ADVISOR_ALIASES.contains(&a.as_str()) && !a.starts_with("claude-") {
            return Err("advisor must be fable|opus|sonnet or a full claude-* model id".into());
        }
        match (model.and_then(rank), rank(a)) {
            (Some(m), Some(x)) if x < m => {
                return Err("advisor must rank at or above the main model".into())
            }
            (None, _) | (_, None) => notes.push("advisor/model ranking not checked: unknown id"),
            _ => {}
        }
        if std::env::var("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .ok()
            .as_deref()
            == Some("1")
        {
            notes.push("CLAUDE_CODE_DISABLE_ADVISOR_TOOL=1 is set: the host ignores --advisor");
        }
        notes.push("--advisor is documented but hidden from `claude --help`; experimental, Anthropic API only");
        argv.push("--advisor".into());
        argv.push(a.clone());
    }
    if r.prompt.is_some() {
        notes.push("PROMPT is used by codex only; it was not passed to the Claude host");
    }
    if r.skip_permissions {
        notes.push("--dangerously-skip-permissions: every tool call runs without a permission prompt (operator interim default 2026-11-08; disable with --permissions)");
        argv.push("--dangerously-skip-permissions".into());
    }
    Ok(argv)
}

fn codex_argv(r: &SessionRequest, notes: &mut Vec<&'static str>) -> Result<Vec<String>, String> {
    clean("codex_bin", &r.codex_bin)?;
    let mut argv = vec![
        r.codex_bin.clone(),
        "-s".into(),
        "workspace-write".into(),
        "-a".into(),
        "on-request".into(),
    ];
    if let Some(m) = &r.model {
        clean("model", m)?;
        argv.push("-m".into());
        argv.push(m.clone());
    }
    if r.effort.is_some() {
        notes.push("--effort is not passed to codex by this launcher");
    }
    if r.advisor.is_some() || r.skip_permissions {
        notes.push("advisor and permission-skip apply to Claude hosts only; codex runs with -s workspace-write -a on-request");
    }
    let mut parts: Vec<String> = r
        .name
        .iter()
        .map(|n| format!("MoAI session name: {n}"))
        .collect();
    if let Some(extra) = &r.prompt {
        if extra.len() > MAX_PROMPT
            || extra
                .chars()
                .any(|c| c.is_control() && c != '\n' && c != '\t')
        {
            return Err("invalid prompt".into());
        }
        parts.push(extra.clone());
    }
    if !parts.is_empty() {
        argv.push(parts.join("\n"));
    }
    if r.name.is_some() {
        notes.push("codex has no session-name flag: the name is carried in the first prompt");
    }
    notes.push("codex sessions have no cross-session channel in RC1 (no ListAgents/SendMessage)");
    Ok(argv)
}

/// Composes the host command without running it. Returns
/// `{host, role, name, argv, env_keys, notes, executed:false, host_loaded:false}`.
/// `env_keys` lists variable names only, never values.
pub fn compose(r: &SessionRequest) -> Result<Value, String> {
    if let Some(role) = &r.role {
        if !ROLES.contains(&role.as_str()) {
            return Err("role must be leader or agent".into());
        }
    }
    if !HOSTS.contains(&r.host.as_str()) {
        return Err(format!("unknown host {} (claude|glm|codex)", r.host));
    }
    if let Some(n) = &r.name {
        clean("name", n)?;
    }
    let mut notes = Vec::new();
    if r.role.is_none() {
        notes.push("no role: a plain host session; no role, session name or permission skip is added (#2053)");
        if r.host != "codex" {
            notes.push(if ROLELESS_LAUNCHER_DEFAULTS {
                "launcher defaults without a role: model opus, effort medium, advisor fable (user decision 2026-11-08: A)"
            } else {
                "no model, effort or advisor flag without a role; the user's Claude settings apply"
            });
        }
    }
    let mut env_keys: Vec<&str> = Vec::new();
    let argv = match r.host.as_str() {
        "codex" => codex_argv(r, &mut notes)?,
        "glm" => {
            if std::env::var_os(GLM_TOKEN_VAR).is_none() {
                if glm_file_token().is_none() {
                    return Err(format!(
                        "blocked: {GLM_TOKEN_VAR} is not set in this environment and no key file holds one (set it in your shell, or put GLM_API_KEY=<key> in ~/{}; ~/{} is read too; the launcher never prints or stores the value)",
                        GLM_KEY_FILES[0], GLM_KEY_FILES[1]
                    ));
                }
                notes.push("ANTHROPIC_AUTH_TOKEN comes from a key file under HOME; it is passed to the host process only");
            }
            let mut plain = r.clone();
            if plain.advisor.take().is_some() {
                notes.push("--advisor omitted on glm: the advisor is Anthropic API only");
            }
            notes.push("model aliases map to z.ai models for this session: opus, sonnet and fable to glm-5.3, haiku to glm-5.3-flash, with the 1M context window declared");
            env_keys.extend(GLM_ENV.iter().map(|(k, _)| *k));
            env_keys.push(GLM_TOKEN_VAR);
            env_keys.push(GLM_MCP_KEY_VAR);
            claude_argv(&plain, &mut notes)?
        }
        _ => claude_argv(r, &mut notes)?,
    };
    Ok(
        json!({"host": r.host, "role": r.role, "name": r.name, "argv": argv,
        "env_keys": env_keys, "notes": notes, "host_loaded": false, "executed": false}),
    )
}

/// Composes and, unless `dry_run`, starts the host and waits for it to exit.
pub fn launch(r: &SessionRequest, dry_run: bool) -> Result<Value, String> {
    let mut v = compose(r)?;
    if dry_run {
        return Ok(v);
    }
    let argv: Vec<String> = serde_json::from_value(v["argv"].clone()).map_err(|_| "argv")?;
    let mut command = std::process::Command::new(&argv[0]);
    command.args(&argv[1..]);
    if r.host == "glm" {
        command.envs(GLM_ENV);
        // The token is inherited unchanged from the caller's environment when it is there.
        let token = match std::env::var_os(GLM_TOKEN_VAR) {
            Some(token) => token,
            None => {
                let token = glm_file_token().ok_or("blocked: the key file changed")?;
                command.env(GLM_TOKEN_VAR, &token);
                token.into()
            }
        };
        command.env(GLM_MCP_KEY_VAR, token);
    }
    let status = command
        .status()
        .map_err(|e| format!("cannot start {}: {e}", argv[0]))?;
    v["executed"] = json!(true);
    v["exit_code"] = json!(status.code());
    Ok(v)
}
