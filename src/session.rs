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
//! - `codex`: Codex CLI `-s danger-full-access -a never [-m <model>] [PROMPT]`
//!   (`-s workspace-write -a on-request` with `--permissions`), per
//!   https://learn.chatgpt.com/docs/developer-commands?surface=cli. Codex has no session-name
//!   flag, so the name travels in the first prompt.
//!
//! Without a role (`moai-cli cc|glm|codex` with neither `-l` nor `-a`, Issue #2053) the launcher
//! opens a plain host session: no role is guessed, no session name is added and no permission
//! flag is passed, so the host starts in the mode its own settings choose. A role session
//! (`-l` named `leader`, `-a` named `agent-N`) starts with full access unless `--permissions`
//! is given ([`FULL_ACCESS_DEFAULT`], user decision 2026-11-10, #2359).
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::fs::{self, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

const ROLES: [&str; 2] = ["leader", "agent"];
const HOSTS: [&str; 3] = ["claude", "glm", "codex"];
const EFFORTS: [&str; 6] = ["low", "medium", "high", "xhigh", "max", "ultracode"];
const ADVISOR_ALIASES: [&str; 3] = ["fable", "opus", "sonnet"];
const DEFAULT_MODEL: &str = "opus";
const DEFAULT_EFFORT: &str = "medium";
/// User decision 2026-11-08 (#2053 HOLD-L1 = A, #2062): a session without a role still gets the
/// launcher's Claude defaults (`--model opus --effort medium --advisor fable`). False would pass no model, effort or advisor flag so the user's own Claude
/// settings apply (option B, not chosen). This constant is the only place to change.
pub const ROLELESS_LAUNCHER_DEFAULTS: bool = true;
/// User decision 2026-11-10 (#2359, comment 6760976883): a role session (`-l`, `-a`) on every
/// host starts with full access and no permission prompt (`--dangerously-skip-permissions` on
/// Claude hosts whatever the project's `permissions` key says, `-s danger-full-access -a never`
/// on codex); `--permissions` keeps the prompts. A roleless session gets no permission flag at
/// all. This replaces #2244 (a project `permissions` key dropped the skip). Choosing this
/// default from a settings key is a recorded follow-up; until then this constant is the only
/// place to change.
pub const FULL_ACCESS_DEFAULT: bool = true;
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
/// Legacy key files under HOME, read in this order when the variable is absent and the vault
/// holds no entry (#2393 S2). The first is the global store shared with the Go tool (user
/// decision 2026-11-10). Both are read, never created or changed. One list with the import.
pub const GLM_KEY_FILES: [&str; 2] = crate::key_vault::LEGACY_GLM_FILES;
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
        // One parser for the launcher and `key-import-*` (#2393 S3).
        crate::key_vault::legacy_token_in(&text)
    })
}

/// The one rule for a token value from any source: surrounding whitespace is dropped, and an
/// empty, whitespace-only or control-character value counts as no token (#2359 AC4).
fn usable_token(value: &str) -> Option<String> {
    let value = value.trim();
    (!value.is_empty() && !value.chars().any(char::is_control)).then(|| value.to_string())
}

/// Where the GLM token came from, by name only (never the value).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum KeySource {
    Env,
    Vault(crate::key_vault::Backend),
    LegacyFile,
}

impl KeySource {
    fn name(self) -> &'static str {
        match self {
            KeySource::Env => "env:ANTHROPIC_AUTH_TOKEN",
            KeySource::Vault(_) => "vault:glm",
            KeySource::LegacyFile => "legacy_key_file",
        }
    }
}

/// The token and its source, in this order (#2393 S2, leader decision on the Issue):
/// 1. the variable, when it holds a usable value (an empty or blank one does not hide the rest);
/// 2. the vault's `glm` entry through the selected backend (`~/.moai-cli/config`, file by
///    default) — a vault that is present but unsafe, or a backend that cannot answer, stops
///    the launch with its code instead of falling through, so a broken store is not bypassed;
/// 3. only when the vault holds no entry, the legacy key files (read only, as before #2393),
///    so a user who has not moved the key yet keeps working and the Go tool's file is kept.
fn glm_token() -> Result<Option<(String, KeySource)>, crate::key_vault::Blocked> {
    let from_env = std::env::var(GLM_TOKEN_VAR)
        .ok()
        .and_then(|v| usable_token(&v));
    if let Some(token) = from_env {
        return Ok(Some((token, KeySource::Env)));
    }
    if let Ok(home) = crate::key_vault::home() {
        if let Some(token) = crate::key_vault::read(&home, "glm")? {
            let backend =
                crate::key_vault::backend(&home).unwrap_or(crate::key_vault::Backend::File);
            return Ok(Some((token, KeySource::Vault(backend))));
        }
    }
    Ok(glm_file_token().map(|token| (token, KeySource::LegacyFile)))
}

/// The launcher's one-line refusal for a vault that cannot give the key: code, reason and the
/// next action, never the value.
fn blocked_line(b: &crate::key_vault::Blocked) -> String {
    format!("blocked: {}: {} (next: {})", b.code, b.reason, b.next)
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
    /// Full access: `--dangerously-skip-permissions` on Claude hosts, `-s danger-full-access
    /// -a never` on codex. The launcher sets it from [`full_access`].
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

/// Whether a launched session runs with full access: only a role session, and then
/// [`FULL_ACCESS_DEFAULT`] unless `--permissions` asks to keep the prompts. Host and project
/// settings do not matter.
pub fn full_access(has_role: bool, keep_prompts: bool) -> bool {
    has_role && FULL_ACCESS_DEFAULT && !keep_prompts
}

// ---- agent-N numbering (#2359 decision 6760976883; resolves #2361) ----

/// Largest store that is read: `u64::MAX` is 20 digits, plus the newline.
const AGENT_SEQ_MAX_BYTES: u64 = 21;
/// Test-only: milliseconds to hold the counter lock between reading and writing.
pub const AGENT_SEQ_TEST_HOLD_VAR: &str = "MOAI_CLI_TEST_AGENT_SEQ_HOLD_MS";

/// Where `-a` session numbers are kept: `<clone common dir>/.moai-cli/sessions/`, beside the
/// claim journal (`claim::journal_root`). Every worktree of one clone resolves the same common
/// dir, so they share one counter across hosts, and git never lists the files. `agent-seq`
/// holds the last number handed out (decimal, one line); `.lock` is the `File::lock` target.
pub fn agent_seq_dir(project: &Path) -> Result<PathBuf, String> {
    let id = crate::identity::identity(project)
        .map_err(|e| format!("agent numbering: {e}; pass --name"))?;
    if id["identity_status"].as_str() != Some("ok") {
        return Err(format!(
            "agent numbering needs a git repository (identity_status {}); run from the project or pass --name",
            id["identity_status"].as_str().unwrap_or("unknown")
        ));
    }
    let common = id["clone_anchor"]
        .as_str()
        .ok_or("agent numbering needs a git repository; pass --name")?;
    Ok(Path::new(common).join(".moai-cli").join("sessions"))
}

fn refuse_symlink(path: &Path) -> Result<(), String> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "agent number store {} is a symlink; refused (pass --name)",
            path.display()
        )),
        _ => Ok(()),
    }
}

/// The last number handed out; 0 when none was. Only a regular file is read (a FIFO at the
/// path would block a plain read). Anything but one decimal number is an error, never 0.
fn read_last(file: &Path) -> Result<u64, String> {
    let broken = || {
        format!(
            "agent number store {} is unreadable or not a number; fix or remove it, or pass --name",
            file.display()
        )
    };
    match fs::symlink_metadata(file) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(0),
        Err(_) => return Err(broken()),
        Ok(m) if !m.file_type().is_file() || m.len() > AGENT_SEQ_MAX_BYTES => return Err(broken()),
        Ok(_) => {}
    }
    let text = fs::read_to_string(file).map_err(|_| broken())?;
    let digits = text.strip_suffix('\n').unwrap_or(&text);
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return Err(broken());
    }
    digits.parse().map_err(|_| broken())
}

fn next_of(last: u64) -> Result<u64, String> {
    last.checked_add(1)
        .ok_or_else(|| "agent number store is exhausted; pass --name".into())
}

/// The number the next `-a` launch would take, read without taking it (`--dry-run`).
pub fn peek_agent_number(project: &Path) -> Result<u64, String> {
    let dir = agent_seq_dir(project)?;
    refuse_symlink(&dir)?;
    next_of(read_last(&dir.join("agent-seq"))?)
}

// @MX:WARN: [AUTO] cross-process critical section; a second launch blocks on `.lock` until the
//   first has renamed its new number into place.
// @MX:REASON: two launches that read the same last number would both open `agent-<n+1>`.
/// Takes the next agent number. Under an exclusive `File::lock` on `.lock` (a concurrent launch
/// waits, it does not fail) it reads the last number, writes last + 1 (temp file, fsync,
/// rename) and returns it. Numbers only increase and are never handed out twice; every failure
/// is an error for the caller to show, never a silent `agent-1`.
pub fn take_agent_number(project: &Path) -> Result<u64, String> {
    let dir = agent_seq_dir(project)?;
    refuse_symlink(&dir)?;
    fs::create_dir_all(&dir).map_err(|_| {
        format!(
            "agent number store {} cannot be created; pass --name",
            dir.display()
        )
    })?;
    let lock_path = dir.join(".lock");
    refuse_symlink(&lock_path)?;
    let lock = OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(&lock_path)
        .and_then(|f| f.lock().map(|()| f))
        .map_err(|_| {
            format!(
                "agent number lock {} unavailable; pass --name",
                lock_path.display()
            )
        })?;
    let file = dir.join("agent-seq");
    refuse_symlink(&file)?;
    let n = next_of(read_last(&file)?)?;
    // Test seam: widens the read-to-write window (at most 2 s) so the concurrency test can
    // show the lock is what keeps numbers apart. Unset in normal use.
    if let Some(ms) = std::env::var(AGENT_SEQ_TEST_HOLD_VAR)
        .ok()
        .and_then(|v| v.parse::<u64>().ok())
    {
        std::thread::sleep(std::time::Duration::from_millis(ms.min(2000)));
    }
    let tmp = dir.join(format!("agent-seq.tmp-{}", std::process::id()));
    let written = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .and_then(|mut f| {
            f.write_all(format!("{n}\n").as_bytes())?;
            f.sync_all()
        })
        .and_then(|()| fs::rename(&tmp, &file));
    if written.is_err() {
        let _ = fs::remove_file(&tmp);
        return Err(format!(
            "agent number store {} cannot be written; pass --name",
            file.display()
        ));
    }
    drop(lock);
    Ok(n)
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
        notes.push("--dangerously-skip-permissions: every tool call runs without a permission prompt (role session, user decision 2026-11-10 #2359; keep prompts with --permissions)");
        argv.push("--dangerously-skip-permissions".into());
    } else if r.role.is_none() {
        notes.push("no permission flag: Claude Code starts in the mode the user's and project's settings choose (roleless session, #2359)");
    }
    Ok(argv)
}

fn codex_argv(r: &SessionRequest, notes: &mut Vec<&'static str>) -> Result<Vec<String>, String> {
    clean("codex_bin", &r.codex_bin)?;
    // User decision 2026-11-10 (#2359): a role session gets full access (no sandbox, never
    // ask), or the sandboxed ask-first pair with `--permissions`; a roleless session gets no
    // flag, so `~/.codex/config.toml` decides (Codex ranks CLI flags above it, so any flag here
    // would override the user's choice).
    let pair = if r.skip_permissions {
        notes.push("codex full access: -s danger-full-access -a never, every command runs without a sandbox or an approval prompt (role session, user decision 2026-11-10 #2359; keep prompts with --permissions)");
        Some(("danger-full-access", "never"))
    } else if r.role.is_some() {
        Some(("workspace-write", "on-request"))
    } else {
        notes.push("no sandbox or approval flag: codex takes them from ~/.codex/config.toml (roleless session, #2359)");
        None
    };
    let mut argv = vec![r.codex_bin.clone()];
    if let Some((sandbox, approval)) = pair {
        argv.extend(["-s".into(), sandbox.into(), "-a".into(), approval.into()]);
    }
    if let Some(m) = &r.model {
        clean("model", m)?;
        argv.push("-m".into());
        argv.push(m.clone());
    }
    if r.effort.is_some() {
        notes.push("--effort is not passed to codex by this launcher");
    }
    if r.advisor.is_some() {
        notes.push("the advisor applies to Claude hosts only; it was not passed to codex");
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
        notes.push("no role: a plain host session; no role or session name is added (#2053)");
        if r.host != "codex" {
            notes.push(if ROLELESS_LAUNCHER_DEFAULTS {
                "launcher defaults without a role: model opus, effort medium, advisor fable (user decision 2026-11-08: A)"
            } else {
                "no model, effort or advisor flag without a role; the user's Claude settings apply"
            });
        }
    }
    let mut env_keys: Vec<&str> = Vec::new();
    let mut key_source: Option<&str> = None;
    let argv = match r.host.as_str() {
        "codex" => codex_argv(r, &mut notes)?,
        "glm" => {
            match glm_token() {
                Err(b) => return Err(blocked_line(&b)),
                Ok(None) => {
                    return Err(format!(
                        "blocked: key_absent: {GLM_TOKEN_VAR} is not set or is empty in this environment, the vault holds no glm key and no key file holds one (store it with `moai-cli key-set glm` reading the value from stdin, or set the variable; the legacy ~/{} and ~/{} are still read; the launcher never prints or stores the value)",
                        GLM_KEY_FILES[0], GLM_KEY_FILES[1]
                    ));
                }
                Ok(Some((_, source))) => {
                    key_source = Some(source.name());
                    match source {
                        KeySource::LegacyFile => notes.push("ANTHROPIC_AUTH_TOKEN comes from a legacy key file under HOME (read only); `moai-cli key-set glm` moves it into the vault; it is passed to the host process only"),
                        KeySource::Vault(_) => notes.push("ANTHROPIC_AUTH_TOKEN comes from the moai-cli key vault; it is passed to the host process only"),
                        KeySource::Env => {}
                    }
                }
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
    let mut out = json!({"host": r.host, "role": r.role, "name": r.name, "argv": argv,
        "env_keys": env_keys, "notes": notes, "host_loaded": false, "executed": false});
    // Only a launch that reads a key names its source (#2393 S2); the other hosts' answer
    // keeps its recorded protocol shape.
    if let Some(source) = key_source {
        out["key_source"] = json!(source);
    }
    Ok(out)
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
        // The same rule as `compose`: both names carry the trimmed token, so an empty variable
        // is replaced by the key file value and never reaches the host.
        let (token, _) = glm_token()
            .map_err(|b| blocked_line(&b))?
            .ok_or("blocked: the key changed")?;
        command.env(GLM_TOKEN_VAR, &token);
        command.env(GLM_MCP_KEY_VAR, &token);
    }
    let status = command
        .status()
        .map_err(|e| format!("cannot start {}: {e}", argv[0]))?;
    v["executed"] = json!(true);
    v["exit_code"] = json!(status.code());
    Ok(v)
}
