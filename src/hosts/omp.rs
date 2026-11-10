//! OMP host adapter (#2362): composes the `omp` (oh-my-pi) argv for the common launcher
//! without running it. Wiring into `session` and the CLI waits for PR #2360; until then this
//! module is reachable only from its own tests.
//!
//! Sources (read 2026-11-10):
//! - `omp --help` of the installed omp 18.18.6 (npm `@oh-my-pi/pi-coding-agent` 18.18.6, the
//!   registry's `latest`): `--model=<value>` (fuzzy match or `provider/id`), `--thinking=<value>`
//!   with `off|minimal|low|medium|high|xhigh|max|auto`, positional MESSAGES (a leading `@`
//!   includes a file), `--auto-approve`, `--approval-mode=<value>`, `--plan-yolo`. No
//!   session-name flag is listed.
//! - Installed package source 18.18.6: `src/tools/settings.ts` declares `tools.approvalMode`
//!   with default `yolo` ("auto-approves every tier"); `src/slash-commands/builtin-lifecycle.ts`
//!   declares `/rename [title]`. Whether a startup positional `/rename` runs as a command is
//!   not verified.
//! - https://github.com/can1357/oh-my-pi README at commit f0200abf (install methods).
//!
//! Undecided on #2362 and therefore not decided here: how the session name reaches omp,
//! launcher defaults for model and thinking, and whether omp opens with a role. The adapter
//! never adds `--auto-approve`, `--approval-mode` or `--plan-yolo`, and offers no field that
//! would.
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

/// `--thinking` values listed by `omp --help` (18.18.6).
pub const THINKING: [&str; 8] = [
    "off", "minimal", "low", "medium", "high", "xhigh", "max", "auto",
];
const MAX_FIELD: usize = 128;
const MAX_PROMPT: usize = 4096;

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct OmpRequest {
    /// Validated and reported; not passed to omp while name delivery is undecided.
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
/// `{host, name, argv, env_keys, notes, host_loaded:false, executed:false}`, the shape of
/// `session::compose`; `env_keys` is always empty (the adapter reads and sets no variable).
pub fn compose(r: &OmpRequest) -> Result<Value, String> {
    clean("omp_bin", &r.omp_bin)?;
    if let Some(n) = &r.name {
        clean("name", n)?;
    }
    let mut notes: Vec<&'static str> = vec![
        "approval follows the user's OMP settings: the launcher passes no --auto-approve, --approval-mode or --plan-yolo; the shipped default of tools.approvalMode in omp 18.18.6 is yolo (auto-approves every tier), so tools may run without a prompt (HOLD on #2362)",
        "this adapter takes no role; whether omp opens as leader or agent is undecided (#2362)",
        "omp cross-session messaging is unverified and not claimed",
    ];
    let mut argv = vec![r.omp_bin.clone()];
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
        notes.push("omp has no session-name flag and name delivery is undecided (#2362): the name is validated and reported, not passed to omp");
    }
    if let Some(p) = &r.prompt {
        clean_prompt(p)?;
        argv.push(p.clone());
    }
    Ok(
        json!({"host": "omp", "name": r.name, "argv": argv, "env_keys": [],
        "notes": notes, "host_loaded": false, "executed": false}),
    )
}
