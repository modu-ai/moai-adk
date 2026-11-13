//! The stdin payload Claude Code passes to a `statusLine` command (official contract,
//! docs/design-review/statusline-rust-stage1-2026-11-10.md §1) plus the extra fields the
//! Go statusline reads (`tasks[]`, `cost.total_usd`, `context_window.used/total`).
//!
//! Decoding follows Go `encoding/json` into the Go `StdinData` shape (#2350): duplicate and
//! case-folded keys, `null` and numbers are resolved by `godecode` first (Go's rules there);
//! a value of the wrong type fails the whole payload, which then degrades to "no input"
//! rather than an error; only the first JSON value of the stream is read. Known difference:
//! a task's raw `effort` deeper than 128 levels fails the payload (Go keeps it).

use std::borrow::Cow;

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
    /// #2756 AC-8: the main conversation's prompt cache (Claude Code 2.1.251+). Read loosely:
    /// an unexpected value here leaves the cache timer out instead of failing the payload.
    pub prompt_cache: Option<Value>,
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
///
/// Like Go's `json.Decoder` into a struct (#2350 audit F1): a top-level `null` is an empty
/// payload, not an error; invalid UTF-8 bytes and unpaired `\uD800`–`\uDFFF` escapes
/// become U+FFFD instead of failing the whole payload.
pub fn parse(bytes: &[u8]) -> Option<Input> {
    let text = go_lossy_utf8(bytes);
    let text = replace_unpaired_surrogates(&text);
    // Go's decoder semantics (duplicate keys, case-folded keys, `null`, number parsing) are
    // applied first; the serde types then read a value with each field set at most once.
    serde_json::from_value(super::godecode::decode(&text)?).ok()
}

/// UTF-8 decoding as Go's `encoding/json` does it inside strings: every byte of an invalid
/// sequence becomes one U+FFFD (`utf8.DecodeRune` returns width 1 for it).
fn go_lossy_utf8(bytes: &[u8]) -> Cow<'_, str> {
    if let Ok(s) = std::str::from_utf8(bytes) {
        return Cow::Borrowed(s);
    }
    let mut out = String::with_capacity(bytes.len() + 8);
    let mut rest = bytes;
    loop {
        match std::str::from_utf8(rest) {
            Ok(s) => {
                out.push_str(s);
                return Cow::Owned(out);
            }
            Err(e) => {
                let (valid, after) = rest.split_at(e.valid_up_to());
                out.push_str(std::str::from_utf8(valid).unwrap_or_default());
                let bad = e.error_len().unwrap_or(after.len());
                out.extend(std::iter::repeat_n('\u{FFFD}', bad));
                rest = &after[bad..];
            }
        }
    }
}

/// Rewrites every `\uXXXX` escape inside a JSON string that is a surrogate without its
/// partner to `\uFFFD`, as Go's `unquote` does; serde_json would reject the payload.
fn replace_unpaired_surrogates(text: &str) -> Cow<'_, str> {
    if !text.contains("\\u") {
        return Cow::Borrowed(text);
    }
    let b = text.as_bytes();
    let hex4 = |at: usize| {
        let digits = b.get(at..at + 4)?;
        u32::from_str_radix(std::str::from_utf8(digits).ok()?, 16).ok()
    };
    let escape_at = |at: usize| (b.get(at..at + 2) == Some(b"\\u")).then(|| hex4(at + 2))?;
    let mut out = String::new();
    let (mut i, mut copied, mut in_string) = (0, 0, false);
    while i < b.len() {
        match (in_string, b[i]) {
            (false, b'"') | (true, b'"') => {
                in_string = !in_string;
                i += 1;
            }
            (true, b'\\') => match escape_at(i) {
                // A high surrogate followed by a low one is a valid pair.
                Some(0xD800..=0xDBFF) if matches!(escape_at(i + 6), Some(0xDC00..=0xDFFF)) => {
                    i += 12
                }
                Some(0xD800..=0xDFFF) => {
                    out.push_str(&text[copied..i]);
                    out.push_str("\\uFFFD");
                    i += 6;
                    copied = i;
                }
                Some(_) => i += 6,
                None => i += 2,
            },
            _ => i += 1,
        }
    }
    if copied == 0 {
        return Cow::Borrowed(text);
    }
    out.push_str(&text[copied..]);
    Cow::Owned(out)
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

    /// #2350 audit F1: valid JSON Go renders must not blank the Rust status line.
    #[test]
    fn top_level_null_is_an_empty_payload() {
        let input = parse(b" null\n").unwrap();
        assert!(input.model.is_none() && input.session_id.is_none());
        assert!(parse(b"nul").is_none());
        assert!(parse(b"nope").is_none());
    }

    #[test]
    fn unpaired_surrogate_escapes_become_replacement_characters() {
        let name = |json: &[u8]| parse(json).unwrap().model.unwrap().display_name;
        assert_eq!(
            name(br#"{"model":{"display_name":"Op\ud800us"}}"#),
            "Op\u{FFFD}us"
        );
        assert_eq!(
            name(br#"{"model":{"display_name":"a\udc00b"}}"#),
            "a\u{FFFD}b"
        );
        // A high surrogate whose next escape is not a low one: only the high is replaced.
        assert_eq!(
            name(br#"{"model":{"display_name":"\ud800\u0041"}}"#),
            "\u{FFFD}A"
        );
        // A valid pair and an escaped backslash stay as they are.
        assert_eq!(
            name(br#"{"model":{"display_name":"\ud83d\ude00"}}"#),
            "\u{1F600}"
        );
        assert_eq!(name(br#"{"model":{"display_name":"\\ud800"}}"#), "\\ud800");
        // Range edges: D7FF and E000 are ordinary characters; D800, DBFF, DC00 and DFFF
        // alone are surrogates; DBFF+DFFF is the last valid pair (U+10FFFF).
        assert_eq!(name(br#"{"model":"\ud7ff\ue000"}"#), "\u{D7FF}\u{E000}");
        assert_eq!(name(br#"{"model":"\ud800\udbff"}"#), "\u{FFFD}\u{FFFD}");
        assert_eq!(name(br#"{"model":"\udc00\udfff"}"#), "\u{FFFD}\u{FFFD}");
        assert_eq!(name(br#"{"model":"\udbff\udfff"}"#), "\u{10FFFF}");
        assert_eq!(name(br#"{"model":"\ud800\ue000"}"#), "\u{FFFD}\u{E000}");
        // Pair start edge: D800+DC00 is the first valid pair (U+10000); D7FF is not a
        // high surrogate, so the DC00 after it stands alone.
        assert_eq!(name(br#"{"model":"\ud800\udc00"}"#), "\u{10000}");
        assert_eq!(name(br#"{"model":"\ud7ff\udc00"}"#), "\u{D7FF}\u{FFFD}");
        let s = parse(br#"{"session_name":"my-se\udc00ssion","session_id":"x"}"#).unwrap();
        assert_eq!(s.session_name.as_deref(), Some("my-se\u{FFFD}ssion"));
        assert_eq!(s.session_id.as_deref(), Some("x"));
    }

    #[test]
    fn invalid_utf8_bytes_become_one_replacement_character_each() {
        let name = |json: &[u8]| parse(json).unwrap().model.unwrap().display_name;
        assert_eq!(name(b"{\"model\":\"a\xff\xfeb\"}"), "a\u{FFFD}\u{FFFD}b");
        // A truncated 3-byte sequence: Go replaces each of its 2 bytes.
        assert_eq!(name(b"{\"model\":\"a\xe2\x82b\"}"), "a\u{FFFD}\u{FFFD}b");
        // Outside a string it is still a syntax error.
        assert!(parse(b"{\xff}").is_none());
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
        let s = parse(br#"{"model": "claude-sonnet-5-5"}"#)
            .unwrap()
            .model
            .unwrap();
        assert_eq!(s.display_name, "claude-sonnet-5-5");
        assert_eq!(s.id, "claude-sonnet-5-5");
        let o = parse(br#"{"model": {"id": "x", "display_name": "Opus"}}"#)
            .unwrap()
            .model
            .unwrap();
        assert_eq!(
            (o.id.as_str(), o.display_name.as_str(), o.name.as_str()),
            ("x", "Opus", "")
        );
        assert!(parse(br#"{"model": 3}"#).is_none());
    }

    #[test]
    fn tasks_present_but_empty_differs_from_absent() {
        assert_eq!(
            parse(br#"{"tasks": []}"#).unwrap().tasks.map(|t| t.len()),
            Some(0)
        );
        assert!(parse(br#"{"tasks": null}"#).unwrap().tasks.is_none());
        assert!(parse(br#"{}"#).unwrap().tasks.is_none());
    }
}
