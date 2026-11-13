//! Native decision records (Issue #2011, 07-D, REQ-NATIVE-001..003).
//!
//! A record says which next step a workflow judged best, from which options and on what
//! input. It never grants anything: `authority` has exactly one value, `none`, so a record
//! that claims more does not parse. The judge is the running coding agent (`native`); a
//! separate judging backend is P10 HOLD and gets its own variant here only once decided.
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const DECISIONS_DIR: &str = "decisions";

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum DecisionSource {
    /// The coding agent running the workflow. P10 backends would be added beside it.
    Native,
}

#[derive(Serialize, Deserialize, Debug, PartialEq, Clone, Copy)]
#[serde(rename_all = "snake_case")]
pub enum Host {
    Claude,
    Codex,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum Confidence {
    /// No confidence was stated; nothing is inferred.
    Unknown,
    /// The judge stated one in its rationale; it is not a measured probability.
    Stated,
}

/// A decision is a candidate next step, never permission, a passed test or done.
#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum Authority {
    None,
}

/// Where a host reported context usage (REQ-CTX-007). Each host's own source only; values
/// from different hosts are never compared or added.
#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum UsageField {
    /// Claude Code statusline `context_window` (input tokens of the latest API response).
    ClaudeStatuslineContextWindow,
    /// Codex `model_context_window`.
    CodexModelContextWindow,
    /// Codex `/status`.
    CodexStatus,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ContextUsageSource {
    pub host: Host,
    pub field: UsageField,
    #[serde(default)]
    pub value: Option<f64>,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct DecisionRecord {
    pub schema_version: u32,
    pub decision_source: DecisionSource,
    pub host: Host,
    pub input_digest: String,
    pub options: Vec<String>,
    pub chosen: String,
    #[serde(default)]
    pub rationale_ref: Option<String>,
    pub confidence: Confidence,
    pub authority: Authority,
    pub recorded_at: String,
    #[serde(default)]
    pub context_usage_source: Option<ContextUsageSource>,
}

fn plain(s: &str, max: usize) -> bool {
    !s.trim().is_empty() && s.len() <= max && !s.chars().any(char::is_control)
}

/// Structural checks serde cannot express. The message names the failed rule, never a value.
pub fn validate(r: &DecisionRecord) -> Result<()> {
    let digest_ok = r.input_digest.len() == 71
        && r.input_digest.starts_with("sha256:")
        && r.input_digest[7..]
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b));
    let usage_ok = r.context_usage_source.as_ref().is_none_or(|u| {
        u.host == r.host
            && match u.host {
                Host::Claude => u.field == UsageField::ClaudeStatuslineContextWindow,
                Host::Codex => u.field != UsageField::ClaudeStatuslineContextWindow,
            }
            && u.value.is_none_or(|v| v.is_finite() && v >= 0.0)
    });
    let rule = if r.schema_version != 1 {
        "schema_version must be 1"
    } else if !digest_ok {
        "input_digest must be sha256:<64 hex>"
    } else if r.options.is_empty()
        || r.options.len() > 32
        || !r.options.iter().all(|o| plain(o, 512))
    {
        "options must be 1 to 32 single-line strings"
    } else if !r.options.contains(&r.chosen) {
        "chosen must be one of options"
    } else if r.rationale_ref.as_deref().is_some_and(|s| !plain(s, 512)) {
        "rationale_ref must be a single-line string"
    } else if !plain(&r.recorded_at, 64) {
        "recorded_at must be a timestamp string"
    } else if !usage_ok {
        "context_usage_source must be the record host's own source"
    } else {
        return Ok(());
    };
    Err(format!("invalid decision record: {rule}"))
}

fn run_id_ok(run: &str) -> bool {
    !run.is_empty()
        && run.len() <= 64
        && !run.starts_with('.')
        && run
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'.'))
}

fn refuse_symlink(path: &Path) -> Result<()> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err("symlink target refused".into()),
        _ => Ok(()),
    }
}

/// Appends `record` as one JSON line to `PROJECT/.moai-cli/decisions/<run>.jsonl`. The only
/// write is that append; the permission reference is caller metadata, not a verifier.
pub fn append(
    project: &Path,
    run: &str,
    record: &DecisionRecord,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    if !run_id_ok(run) {
        return Err("invalid run id; use 1 to 64 of A-Z a-z 0-9 - _ .".into());
    }
    validate(record)?;
    let root = fs::canonicalize(project).map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    let state = crate::state_root::path(&root, "");
    let dir = state.join(DECISIONS_DIR);
    let file: PathBuf = dir.join(format!("{run}.jsonl"));
    for p in [&state, &dir, &file] {
        refuse_symlink(p)?;
    }
    fs::create_dir_all(&dir).map_err(|_| "decisions directory unavailable")?;
    let created = fs::symlink_metadata(&file).is_err();
    let mut line = serde_json::to_vec(record).expect("typed record");
    line.push(b'\n');
    let mut f = fs::OpenOptions::new()
        .append(true)
        .create(true)
        .open(&file)
        .map_err(|_| "decision write failed")?;
    f.write_all(&line)
        .and_then(|_| f.sync_all())
        .map_err(|_| "decision write failed")?;
    let rel = crate::state_root::rel(&root, &format!("{DECISIONS_DIR}/{run}.jsonl"));
    Ok(json!({
        "path": rel,
        "line_sha256": crate::digest(&line),
        "authority": "none",
        "files_created": if created { vec![rel.clone()] } else { vec![] },
        "network": "disabled",
    }))
}
