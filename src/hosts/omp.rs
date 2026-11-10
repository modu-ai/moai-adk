//! OMP host adapter (#2362): composes the `omp` (oh-my-pi) argv for the common launcher
//! without running it. Wiring into `session` and the CLI belongs to #2359; until then this
//! module is reachable only from its own tests.
//!
//! Launcher rule (user decision relayed on #2359, comment 6760976883): `-l` opens a session
//! named `leader` and `-a` one named `agent-N`, both with full access; a session without a
//! role gets no permission flag at all, so the user's own OMP settings apply. For OMP full
//! access is `--auto-approve`. The agent-N counter is #2359's; this adapter only receives the
//! name. A role session asking to keep prompts (`--permissions`) is refused here until the
//! OMP confirm-mode value is decided (HOLD, #2362 comment 6761185710).
//!
//! Sources (read 2026-11-10):
//! - `omp --help` of the installed omp 18.18.6 (npm `@oh-my-pi/pi-coding-agent` 18.18.6):
//!   `--model=<value>`, `--thinking=<value>` with `off|minimal|low|medium|high|xhigh|max|auto`,
//!   positional MESSAGES (a leading `@` includes a file), `--auto-approve` ("Auto-approve all
//!   tool calls"), `--approval-mode=<value>`, `--plan-yolo`. No session-name flag is listed.
//!   Installed source: `src/cli/args.ts:401` (`--auto-approve` / `--yolo`), `src/main.ts:1993`
//!   (sets `tools.approvalMode` to `yolo` for the session); `src/tools/settings.ts` ships
//!   `tools.approvalMode` default `yolo`.
//! - Official docs, https://github.com/can1357/oh-my-pi at f0200abf, fetched 2026-11-10T06:58Z:
//!   `docs/cli-reference.md` (sha256 d55ddf65...1e23e) "`--auto-approve`, `--yolo`: Force yolo
//!   tier approval; explicit tool/user policies and provider safety checks still apply";
//!   `docs/approval-mode.md` (sha256 8d50e639...34931) `yolo` is the default mode,
//!   `tools.restrictedFolders` still prompts under `--auto-approve`.
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

/// `--thinking` values listed by `omp --help` (18.18.6).
pub const THINKING: [&str; 8] = [
    "off", "minimal", "low", "medium", "high", "xhigh", "max", "auto",
];
const MAX_FIELD: usize = 128;
const MAX_PROMPT: usize = 4096;

/// Launcher role. `-l` is `Leader`, `-a` is `Agent`; no flag is `None` on the request.
#[derive(Clone, Copy, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum Role {
    Leader,
    Agent,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct OmpRequest {
    /// `Some` opens a role session with full access (`--auto-approve`); `None` adds no
    /// permission flag.
    #[serde(default)]
    pub role: Option<Role>,
    /// `--permissions`: keep approval prompts in a role session. Refused for now (HOLD).
    #[serde(default)]
    pub keep_prompts: bool,
    /// Session name (`leader`, `agent-N` or `--name`), required for a role session. Validated
    /// and reported; not passed to omp, which lists no session-name flag.
    #[serde(default)]
    pub name: Option<String>,
    /// `None` passes no `--model`; the user's OMP settings choose.
    #[serde(default)]
    pub model: Option<String>,
    /// `None` passes no `--thinking`; the user's OMP settings choose.
    #[serde(default)]
    pub thinking: Option<String>,
    /// First message, passed as the single positional argument.
    #[serde(default)]
    pub prompt: Option<String>,
    #[serde(default = "default_omp_bin")]
    pub omp_bin: String,
}

impl Default for OmpRequest {
    fn default() -> Self {
        Self {
            role: None,
            keep_prompts: false,
            name: None,
            model: None,
            thinking: None,
            prompt: None,
            omp_bin: default_omp_bin(),
        }
    }
}

fn default_omp_bin() -> String {
    "omp".into()
}

/// Rejects empty, oversized, control-bearing, space-bearing and flag-shaped values.
fn clean(field: &str, value: &str) -> Result<(), String> {
    if value.is_empty()
        || value.len() > MAX_FIELD
        || value.starts_with('-')
        || value.chars().any(|c| c.is_control() || c == ' ')
    {
        return Err(format!("invalid {field}"));
    }
    Ok(())
}

/// The positional message is a bare argv token: a leading `-` would be parsed by omp as a flag
/// (for example `--auto-approve`) and a leading `@` as a file include, so both are refused.
fn clean_prompt(value: &str) -> Result<(), String> {
    let head = value.trim_start();
    if head.is_empty()
        || value.len() > MAX_PROMPT
        || head.starts_with('-')
        || head.starts_with('@')
        || value
            .chars()
            .any(|c| c.is_control() && c != '\n' && c != '\t')
    {
        return Err("invalid prompt".into());
    }
    Ok(())
}

/// Composes the omp command without running it. Returns
/// `{host, role, name, argv, env_keys, notes, host_loaded:false, executed:false}`, the shape of
/// `session::compose`; `env_keys` is always empty (the adapter reads and sets no variable).
pub fn compose(r: &OmpRequest) -> Result<Value, String> {
    clean("omp_bin", &r.omp_bin)?;
    if let Some(n) = &r.name {
        clean("name", n)?;
    }
    let mut argv = vec![r.omp_bin.clone()];
    let mut notes: Vec<&'static str> = Vec::new();
    match r.role {
        Some(_) if r.keep_prompts => {
            return Err("an omp role session with --permissions is not decided: omp's default approval mode is yolo, so passing nothing is not a confirm mode (HOLD, #2362 comment 6761185710)".into())
        }
        Some(_) if r.name.is_none() => {
            return Err("an omp role session needs a session name (leader or agent-N)".into())
        }
        Some(_) => {
            argv.push("--auto-approve".into());
            notes.push("role session: full access via --auto-approve (forces yolo tier approval for this session; tools.approval deny/prompt policies, tools.restrictedFolders and provider safety checks still apply)");
        }
        None => notes.push("no role: the launcher passes no permission flag, so the user's OMP settings apply (the shipped tools.approvalMode default in omp 18.18.6 is yolo)"),
    }
    notes.push("omp cross-session messaging is unverified and not claimed");
    match &r.model {
        Some(m) => {
            clean("model", m)?;
            argv.push(format!("--model={m}"));
        }
        None => notes.push("no --model: the user's OMP settings choose the model (launcher default undecided, #2362)"),
    }
    match &r.thinking {
        Some(t) if THINKING.contains(&t.as_str()) => argv.push(format!("--thinking={t}")),
        Some(_) => {
            return Err(
                "thinking must be one of off|minimal|low|medium|high|xhigh|max|auto".into(),
            )
        }
        None => notes.push("no --thinking: the user's OMP settings choose the level (launcher default undecided, #2362)"),
    }
    if r.name.is_some() {
        notes.push("omp lists no session-name flag and name delivery to omp is unverified (#2362): the name is validated and reported, not passed to omp");
    }
    if let Some(p) = &r.prompt {
        clean_prompt(p)?;
        argv.push(p.clone());
    }
    Ok(
        json!({"host": "omp", "role": r.role, "name": r.name, "argv": argv, "env_keys": [],
        "notes": notes, "host_loaded": false, "executed": false}),
    )
}
