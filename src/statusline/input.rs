//! The stdin payload Claude Code passes to a `statusLine` command (official contract,
//! docs/design-review/statusline-rust-stage1-2026-11-10.md §1) plus the extra fields the
//! Go statusline reads (`tasks[]`, `cost.total_usd`, `context_window.used/total`).
//!
//! Decoding follows Go `encoding/json` into the Go `StdinData` shape (#2350): every field
//! is optional and `null` reads as absent; a value of the wrong type fails the whole
//! payload, which then degrades to "no input" rather than an error; only the first JSON
//! value of the stream is read. Known difference: Go matches keys case-insensitively,
//! this decoder matches them exactly.

use serde::{Deserialize, Deserializer};
use serde_json::Value;

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Input {
    pub hook_event_name: Option<String>,
    pub session_id: Option<String>,
    pub session_name: Option<String>,
    pub transcript_path: Option<String>,
    pub cwd: Option<String>,
    #[serde(deserialize_with = "model_info")]
    pub model: Option<ModelInfo>,
    pub workspace: Option<Workspace>,
    pub cost: Option<Cost>,
    pub context_window: Option<ContextWindow>,
    pub output_style: Option<Named>,
    pub rate_limits: Option<RateLimits>,
    pub effort: Option<Effort>,
    pub thinking: Option<Thinking>,
    pub version: Option<String>,
    pub pr: Option<Pr>,
    pub agent: Option<Named>,
    pub worktree: Option<Worktree>,
    pub exceeds_200k_tokens: Option<bool>,
    /// Subagent status-line rows. Present (even `[]`) switches the output to JSONL rows;
    /// absent or `null` keeps the multi-line bar.
    pub tasks: Option<Vec<SubagentTask>>,
}

#[derive(Debug, Default, Clone, PartialEq, Eq)]
pub struct ModelInfo {
    pub id: String,
    pub display_name: String,
    pub name: String,
}

/// `model` is an object, or a bare string that fills all three names (Go `UnmarshalJSON`).
fn model_info<'de, D: Deserializer<'de>>(d: D) -> Result<Option<ModelInfo>, D::Error> {
    #[derive(Deserialize)]
    #[serde(default)]
    #[derive(Default)]
    struct Object {
        id: Option<String>,
        display_name: Option<String>,
        name: Option<String>,
    }
    #[derive(Deserialize)]
    #[serde(untagged)]
    enum Shape {
        Text(String),
        Object(Object),
    }
    Ok(Option::<Shape>::deserialize(d)?.map(|shape| match shape {
        Shape::Text(s) => ModelInfo {
            id: s.clone(),
            display_name: s.clone(),
            name: s,
        },
        Shape::Object(o) => ModelInfo {
            id: o.id.unwrap_or_default(),
            display_name: o.display_name.unwrap_or_default(),
            name: o.name.unwrap_or_default(),
        },
    }))
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Workspace {
    pub current_dir: Option<String>,
    pub project_dir: Option<String>,
    pub git_worktree: Option<String>,
    pub repo: Option<Repo>,
}

#[derive(Debug, Default, Clone, Deserialize)]
#[serde(default)]
pub struct Repo {
    pub host: Option<String>,
    pub owner: Option<String>,
    pub name: Option<String>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Cost {
    pub total_usd: Option<f64>,
    pub total_cost_usd: Option<f64>,
    pub input_tokens: Option<i64>,
    pub output_tokens: Option<i64>,
    pub total_duration_ms: Option<i64>,
    pub total_lines_added: Option<i64>,
    pub total_lines_removed: Option<i64>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct ContextWindow {
    pub used_percentage: Option<f64>,
    pub remaining_percentage: Option<f64>,
    pub context_window_size: Option<i64>,
    pub total_input_tokens: Option<i64>,
    pub total_output_tokens: Option<i64>,
    pub used: Option<i64>,
    pub total: Option<i64>,
    pub current_usage: Option<CurrentUsage>,
}

#[derive(Debug, Default, Clone, Deserialize)]
#[serde(default)]
pub struct CurrentUsage {
    pub input_tokens: Option<i64>,
    pub cache_creation_input_tokens: Option<i64>,
    pub cache_read_input_tokens: Option<i64>,
    pub output_tokens: Option<i64>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Named {
    pub name: Option<String>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct RateLimits {
    pub five_hour: Option<RateWindow>,
    pub seven_day: Option<RateWindow>,
}

#[derive(Debug, Default, Clone, Copy, Deserialize)]
#[serde(default)]
pub struct RateWindow {
    pub used_percentage: Option<f64>,
    /// Unix epoch seconds.
    pub resets_at: Option<i64>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Effort {
    pub level: Option<String>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Thinking {
    pub enabled: Option<bool>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Pr {
    pub number: Option<i64>,
    pub url: Option<String>,
    pub review_state: Option<String>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct Worktree {
    pub name: Option<String>,
    pub path: Option<String>,
    pub branch: Option<String>,
    pub original_cwd: Option<String>,
    pub original_branch: Option<String>,
}

/// One `tasks[]` row. Fields Go types strictly are typed here too, so a malformed row
/// fails the payload the same way.
#[derive(Debug, Default, Deserialize)]
#[serde(default)]
pub struct SubagentTask {
    pub id: Option<String>,
    pub name: Option<String>,
    #[serde(rename = "type")]
    pub kind: Option<String>,
    #[serde(rename = "agentType")]
    pub agent_type: Option<String>,
    pub status: Option<String>,
    pub description: Option<String>,
    pub label: Option<String>,
    #[serde(rename = "startTime")]
    pub start_time: Option<i64>,
    pub model: Option<String>,
    pub effort: Option<Value>,
    #[serde(rename = "contextWindowSize")]
    pub context_window_size: Option<i64>,
    #[serde(rename = "tokenCount")]
    pub token_count: Option<i64>,
    #[serde(rename = "tokenSamples")]
    pub token_samples: Option<Vec<i64>>,
    pub cwd: Option<String>,
}

/// The first JSON value of `bytes`, or `None` for empty, malformed or mistyped input.
pub fn parse(bytes: &[u8]) -> Option<Input> {
    serde_json::Deserializer::from_slice(bytes)
        .into_iter::<Input>()
        .next()?
        .ok()
}

/// A string field, empty when absent.
pub fn text(v: &Option<String>) -> &str {
    v.as_deref().unwrap_or("")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn invalid_or_empty_input_is_none() {
        assert!(parse(b"{ this is not json\n").is_none());
        assert!(parse(b"").is_none());
        assert!(parse(b"[]").is_none());
    }

    #[test]
    fn a_mistyped_field_fails_the_whole_payload() {
        assert!(parse(br#"{"session_id": 7}"#).is_none());
        assert!(parse(br#"{"cost": {"total_duration_ms": 1.5}}"#).is_none());
        assert!(parse(br#"{"session_id": "s", "cost": {"total_duration_ms": 1000}}"#).is_some());
    }

    #[test]
    fn null_reads_as_absent_and_only_the_first_value_counts() {
        let input = parse(br#"{"session_id": null, "context_window": null} trailing"#).unwrap();
        assert!(input.session_id.is_none());
        assert!(input.context_window.is_none());
    }

    #[test]
    fn model_accepts_a_string_or_an_object() {
        let s = parse(br#"{"model": "claude-sonnet-5-5"}"#).unwrap().model.unwrap();
        assert_eq!(s.display_name, "claude-sonnet-5-5");
        assert_eq!(s.id, "claude-sonnet-5-5");
        let o = parse(br#"{"model": {"id": "x", "display_name": "Opus"}}"#)
            .unwrap()
            .model
            .unwrap();
        assert_eq!((o.id.as_str(), o.display_name.as_str(), o.name.as_str()), ("x", "Opus", ""));
        assert!(parse(br#"{"model": 3}"#).is_none());
    }

    #[test]
    fn tasks_present_but_empty_differs_from_absent() {
        assert_eq!(parse(br#"{"tasks": []}"#).unwrap().tasks.map(|t| t.len()), Some(0));
        assert!(parse(br#"{"tasks": null}"#).unwrap().tasks.is_none());
        assert!(parse(br#"{}"#).unwrap().tasks.is_none());
    }
}
