//! Development launcher: starts a leader or agent session on a host CLI with the
//! model/effort/advisor options the operator decided (2026-11-08). Only the Claude Code
//! host is wired in this slice; the launcher composes argv from documented flags and
//! never invents options. Advisor rules follow https://code.claude.com/docs/en/advisor
//! (read 2026-11-08): `--advisor <model>` takes an alias (`fable`, `opus`, `sonnet`) or a
//! full id, must rank at or above the main model, and is disabled entirely when
//! `CLAUDE_CODE_DISABLE_ADVISOR_TOOL=1` is set.
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

const ROLES: [&str; 2] = ["leader", "agent"];
const EFFORTS: [&str; 6] = ["low", "medium", "high", "xhigh", "max", "ultracode"];
const ADVISOR_ALIASES: [&str; 3] = ["fable", "opus", "sonnet"];

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct SessionRequest {
    pub role: String,
    pub host: String,
    pub name: String,
    pub model: String,
    pub effort: String,
    pub advisor: Option<String>,
    pub claude_bin: String,
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

/// Composes the host command without running it. Returns `{host, argv, notes}`.
pub fn compose(r: &SessionRequest) -> Result<Value, String> {
    if !ROLES.contains(&r.role.as_str()) {
        return Err("role must be leader or agent".into());
    }
    if r.host != "claude" {
        return Err(format!("host {} not implemented in this slice", r.host));
    }
    if !EFFORTS.contains(&r.effort.as_str()) {
        return Err("effort must be one of low|medium|high|xhigh|max|ultracode".into());
    }
    clean("name", &r.name)?;
    clean("model", &r.model)?;
    clean("claude_bin", &r.claude_bin)?;
    let mut argv = vec![
        r.claude_bin.clone(),
        "--name".into(),
        r.name.clone(),
        "--model".into(),
        r.model.clone(),
        "--effort".into(),
        r.effort.clone(),
    ];
    let mut notes = Vec::new();
    if let Some(a) = &r.advisor {
        clean("advisor", a)?;
        if !ADVISOR_ALIASES.contains(&a.as_str()) && !a.starts_with("claude-") {
            return Err("advisor must be fable|opus|sonnet or a full claude-* model id".into());
        }
        match (rank(&r.model), rank(a)) {
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
    Ok(
        json!({"host": r.host, "role": r.role, "argv": argv, "notes": notes,
        "host_loaded": false, "executed": false}),
    )
}

/// Composes and, unless `dry_run`, replaces nothing: it spawns the host and waits.
pub fn launch(r: &SessionRequest, dry_run: bool) -> Result<Value, String> {
    let mut v = compose(r)?;
    if dry_run {
        return Ok(v);
    }
    let argv: Vec<String> = serde_json::from_value(v["argv"].clone()).map_err(|_| "argv")?;
    let status = std::process::Command::new(&argv[0])
        .args(&argv[1..])
        .status()
        .map_err(|e| format!("cannot start {}: {e}", argv[0]))?;
    v["executed"] = json!(true);
    v["exit_code"] = json!(status.code());
    Ok(v)
}
