//! The stdin payload Claude Code passes to a `statusLine` command (official statusline page,
//! "Available data") plus the `tasks[]` rows of the subagent status line and the
//! `context_window.used/total` pair an older host sent.
//!
//! The tolerant reader (SPEC-20261122-statusline-limits REQ-10, D-7): the stdin is parsed once
//! into a `serde_json::Value` and every field a shown item needs is taken with its own typed
//! getter, so a field of the wrong type reads as absent and changes only itself, and an unknown
//! field is never looked at. A payload that is not a JSON object (an array, a string, a number,
//! `null`, a boolean, broken JSON, no bytes) reads as the empty object. Keys match exactly, with
//! no case folding. A key given twice keeps its last value whole: serde_json's map keeps the
//! last insert, so two objects under one key are not merged. Only the first JSON value of the
//! stream is read. Invalid UTF-8 bytes and unpaired `\uD800`–`\uDFFF` escapes become U+FFFD
//! before the parse, so one bad string never empties the payload; a payload nested deeper than
//! serde_json's recursion limit (128) does not parse and reads as empty.
//!
//! REQ-11: only what a shown item needs is read. The cost, the session duration, the output
//! style, the agent name, the transcript path and the prompt cache are not.

use std::borrow::Cow;

use serde_json::Value;

#[derive(Debug, Default)]
pub struct Input {
    pub session_id: Option<String>,
    pub cwd: Option<String>,
    pub model: Option<ModelInfo>,
    pub workspace: Option<Workspace>,
    pub context_window: Option<ContextWindow>,
    pub rate_limits: Option<RateLimits>,
    /// `Some` when `effort` is an object, even without a level.
    pub effort: Option<Effort>,
    pub thinking: Option<Thinking>,
    pub version: Option<String>,
    pub pr: Option<Pr>,
    pub worktree: Option<Worktree>,
    /// Subagent status-line rows. An array (even `[]`) switches the output to JSONL rows;
    /// anything else keeps the multi-line bar.
    pub tasks: Option<Vec<SubagentTask>>,
}

#[derive(Debug, Default, Clone, PartialEq, Eq)]
pub struct ModelInfo {
    pub id: String,
    pub display_name: String,
    pub name: String,
}

#[derive(Debug, Default)]
pub struct Workspace {
    pub current_dir: Option<String>,
    pub project_dir: Option<String>,
    pub git_worktree: Option<String>,
    pub repo: Option<Repo>,
}

#[derive(Debug, Default, Clone)]
pub struct Repo {
    pub owner: Option<String>,
    pub name: Option<String>,
}

#[derive(Debug, Default)]
pub struct ContextWindow {
    pub used_percentage: Option<f64>,
    pub context_window_size: Option<i64>,
    pub used: Option<i64>,
    pub total: Option<i64>,
    pub current_usage: Option<CurrentUsage>,
}

/// The input token counts of `current_usage`, which the `CW` number falls back to.
#[derive(Debug, Default, Clone)]
pub struct CurrentUsage {
    pub input_tokens: Option<i64>,
    pub cache_creation_input_tokens: Option<i64>,
    pub cache_read_input_tokens: Option<i64>,
}

#[derive(Debug, Default)]
pub struct RateLimits {
    pub five_hour: Option<RateWindow>,
    pub seven_day: Option<RateWindow>,
}

#[derive(Debug, Default, Clone, Copy)]
pub struct RateWindow {
    pub used_percentage: Option<f64>,
    /// Unix epoch seconds.
    pub resets_at: Option<i64>,
}

#[derive(Debug, Default)]
pub struct Effort {
    pub level: Option<String>,
}

#[derive(Debug, Default)]
pub struct Thinking {
    pub enabled: Option<bool>,
}

#[derive(Debug, Default)]
pub struct Pr {
    pub number: Option<i64>,
    pub url: Option<String>,
    pub review_state: Option<String>,
    /// `mr` when the request is a GitLab merge request, absent for a GitHub pull request
    /// (official statusline page).
    pub kind: Option<String>,
}

#[derive(Debug, Default)]
pub struct Worktree {
    pub name: Option<String>,
    pub original_cwd: Option<String>,
}

/// One `tasks[]` row: the fields its row shows.
#[derive(Debug, Default)]
pub struct SubagentTask {
    pub id: Option<String>,
    pub name: Option<String>,
    pub agent_type: Option<String>,
    pub status: Option<String>,
    pub description: Option<String>,
    pub label: Option<String>,
    pub token_count: Option<i64>,
}

/// The payload of `bytes`, field by field; the empty payload when the first JSON value of
/// `bytes` is not an object.
pub fn parse(bytes: &[u8]) -> Input {
    let text = lossy_utf8(bytes);
    let text = replace_unpaired_surrogates(&text);
    let value = serde_json::Deserializer::from_str(&text)
        .into_iter::<Value>()
        .next()
        .and_then(Result::ok)
        .unwrap_or(Value::Null);
    read(&value)
}

/// Each field with its own getter: a wrong type is an absent field, never a failed payload.
fn read(v: &Value) -> Input {
    if !v.is_object() {
        return Input::default();
    }
    Input {
        session_id: get_str(v, "session_id"),
        cwd: get_str(v, "cwd"),
        model: match v.get("model") {
            Some(Value::String(s)) => Some(ModelInfo {
                id: s.clone(),
                display_name: s.clone(),
                name: s.clone(),
            }),
            Some(m @ Value::Object(_)) => Some(ModelInfo {
                id: get_str(m, "id").unwrap_or_default(),
                display_name: get_str(m, "display_name").unwrap_or_default(),
                name: get_str(m, "name").unwrap_or_default(),
            }),
            _ => None,
        },
        workspace: get_object(v, "workspace").map(|w| Workspace {
            current_dir: get_str(w, "current_dir"),
            project_dir: get_str(w, "project_dir"),
            git_worktree: get_str(w, "git_worktree"),
            repo: get_object(w, "repo").map(|r| Repo {
                owner: get_str(r, "owner"),
                name: get_str(r, "name"),
            }),
        }),
        context_window: get_object(v, "context_window").map(|cw| ContextWindow {
            used_percentage: get_f64(cw, "used_percentage"),
            context_window_size: get_i64(cw, "context_window_size"),
            used: get_i64(cw, "used"),
            total: get_i64(cw, "total"),
            current_usage: get_object(cw, "current_usage").map(|u| CurrentUsage {
                input_tokens: get_i64(u, "input_tokens"),
                cache_creation_input_tokens: get_i64(u, "cache_creation_input_tokens"),
                cache_read_input_tokens: get_i64(u, "cache_read_input_tokens"),
            }),
        }),
        rate_limits: get_object(v, "rate_limits").map(|r| RateLimits {
            five_hour: rate_window(r, "five_hour"),
            seven_day: rate_window(r, "seven_day"),
        }),
        effort: get_object(v, "effort").map(|e| Effort {
            level: get_str(e, "level"),
        }),
        thinking: get_object(v, "thinking").map(|t| Thinking {
            enabled: get_bool(t, "enabled"),
        }),
        version: get_str(v, "version"),
        pr: get_object(v, "pr").map(|p| Pr {
            number: get_i64(p, "number"),
            url: get_str(p, "url"),
            review_state: get_str(p, "review_state"),
            kind: get_str(p, "kind"),
        }),
        worktree: get_object(v, "worktree").map(|w| Worktree {
            name: get_str(w, "name"),
            original_cwd: get_str(w, "original_cwd"),
        }),
        tasks: v.get("tasks").and_then(Value::as_array).map(|rows| {
            rows.iter()
                .map(|t| SubagentTask {
                    id: get_str(t, "id"),
                    name: get_str(t, "name"),
                    agent_type: get_str(t, "agentType"),
                    status: get_str(t, "status"),
                    description: get_str(t, "description"),
                    label: get_str(t, "label"),
                    token_count: get_i64(t, "tokenCount"),
                })
                .collect()
        }),
    }
}

fn rate_window(v: &Value, key: &str) -> Option<RateWindow> {
    get_object(v, key).map(|w| RateWindow {
        used_percentage: get_f64(w, "used_percentage"),
        resets_at: get_i64(w, "resets_at"),
    })
}

/// The object under `key`; `None` when absent or of another type.
fn get_object<'a>(v: &'a Value, key: &str) -> Option<&'a Value> {
    v.get(key).filter(|o| o.is_object())
}

fn get_str(v: &Value, key: &str) -> Option<String> {
    v.get(key)?.as_str().map(str::to_string)
}

/// An integer literal that fits `i64`; a fraction or an exponent is another type.
fn get_i64(v: &Value, key: &str) -> Option<i64> {
    v.get(key)?.as_i64()
}

/// Any number.
fn get_f64(v: &Value, key: &str) -> Option<f64> {
    v.get(key)?.as_f64()
}

fn get_bool(v: &Value, key: &str) -> Option<bool> {
    v.get(key)?.as_bool()
}

/// UTF-8 decoding with every byte of an invalid sequence replaced by one U+FFFD, so one bad
/// string never empties the payload.
fn lossy_utf8(bytes: &[u8]) -> Cow<'_, str> {
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
    fn unpaired_surrogate_escapes_become_replacement_characters() {
        let name = |json: &[u8]| parse(json).model.unwrap().display_name;
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
        let s = parse(br#"{"cwd":"my-se\udc00ssion","session_id":"x"}"#);
        assert_eq!(s.cwd.as_deref(), Some("my-se\u{FFFD}ssion"));
        assert_eq!(s.session_id.as_deref(), Some("x"));
    }

    #[test]
    fn invalid_utf8_bytes_become_one_replacement_character_each() {
        let name = |json: &[u8]| parse(json).model.unwrap().display_name;
        assert_eq!(name(b"{\"model\":\"a\xff\xfeb\"}"), "a\u{FFFD}\u{FFFD}b");
        // A truncated 3-byte sequence: each of its 2 bytes is replaced.
        assert_eq!(name(b"{\"model\":\"a\xe2\x82b\"}"), "a\u{FFFD}\u{FFFD}b");
        // Outside a string it is still a syntax error, so the payload reads as empty.
        assert!(parse(b"{\xff}").model.is_none());
    }

    /// Numbers by their getter: an integer field takes an integer literal only, a percentage
    /// any number; another kind is an absent field and the rest of the object stays.
    #[test]
    fn a_number_of_another_kind_is_an_absent_field() {
        let cw = parse(
            br#"{"context_window": {"used_percentage": 40, "context_window_size": 1.5e5, "used": 1.0, "total": 9}}"#,
        )
        .context_window
        .unwrap();
        assert_eq!(cw.used_percentage, Some(40.0));
        assert_eq!(
            (cw.context_window_size, cw.used, cw.total),
            (None, None, Some(9))
        );
        let week = parse(
            br#"{"rate_limits": {"seven_day": {"used_percentage": 41.2, "resets_at": "soon"}}}"#,
        )
        .rate_limits
        .unwrap()
        .seven_day
        .unwrap();
        assert_eq!((week.used_percentage, week.resets_at), (Some(41.2), None));
    }

    #[test]
    fn null_reads_as_absent_and_only_the_first_value_counts() {
        let input = parse(br#"{"session_id": null, "context_window": null} trailing"#);
        assert!(input.session_id.is_none());
        assert!(input.context_window.is_none());
        assert_eq!(
            parse(br#"{"version": "2.1.330"} {"version": "9"}"#)
                .version
                .as_deref(),
            Some("2.1.330")
        );
    }

    #[test]
    fn model_accepts_a_string_or_an_object() {
        let s = parse(br#"{"model": "claude-sonnet-5-5"}"#).model.unwrap();
        assert_eq!(s.display_name, "claude-sonnet-5-5");
        assert_eq!(s.id, "claude-sonnet-5-5");
        let o = parse(br#"{"model": {"id": "x", "display_name": "Opus"}}"#)
            .model
            .unwrap();
        assert_eq!(
            (o.id.as_str(), o.display_name.as_str(), o.name.as_str()),
            ("x", "Opus", "")
        );
        // Another type is no model; a mistyped name inside the object is an empty name.
        assert!(parse(br#"{"model": 3}"#).model.is_none());
        let typed = parse(br#"{"model": {"id": "y", "display_name": 4}}"#)
            .model
            .unwrap();
        assert_eq!((typed.id.as_str(), typed.display_name.as_str()), ("y", ""));
    }

    #[test]
    fn tasks_present_but_empty_differs_from_absent() {
        assert_eq!(parse(br#"{"tasks": []}"#).tasks.map(|t| t.len()), Some(0));
        assert!(parse(br#"{"tasks": null}"#).tasks.is_none());
        assert!(parse(br#"{}"#).tasks.is_none());
        // A row that is not an object reads as an empty row (its row is then skipped for its
        // empty id); `tasks` of another type keeps the bar.
        let rows = parse(br#"{"tasks": [7, {"id": "t1", "tokenCount": "many"}]}"#)
            .tasks
            .unwrap();
        assert_eq!(rows.len(), 2);
        assert!(rows[0].id.is_none());
        assert_eq!(
            (rows[1].id.as_deref(), rows[1].token_count),
            (Some("t1"), None)
        );
        assert!(parse(br#"{"tasks": {"id": "t1"}}"#).tasks.is_none());
    }
}

/// AC-8 of SPEC-20261122-statusline-limits (REQ-10, REQ-11): the stdin is read field by field,
/// so a field of the wrong type or an unknown field changes only itself, a payload that is not a
/// JSON object reads as empty, and the values no shown item needs are not read.
#[cfg(test)]
mod tolerant_reader_tests {
    use super::*;
    use crate::statusline::state::{self, Context};
    use serde_json::json;
    use std::path::Path;

    /// The render clock of these tests (`--now-unix` seam).
    const NOW: i64 = 1_738_429_000;

    fn read(bytes: &[u8]) -> Input {
        parse(bytes)
    }

    /// A context with a temporary HOME, no working directory and plain output, at [`NOW`].
    fn ctx(home: &Path) -> Context {
        let vars = [("HOME", home.to_str().unwrap()), ("NO_COLOR", "1")];
        let mut ctx = Context::for_test(&vars, None);
        ctx.now = NOW;
        ctx
    }

    fn render(payload: &[u8], home: &Path) -> String {
        crate::statusline::render(payload, &ctx(home))
    }

    fn render_value(payload: &Value, home: &Path) -> String {
        render(payload.to_string().as_bytes(), home)
    }

    /// A payload with every item the lines show.
    fn full() -> Value {
        json!({
            "session_id": "tolerant-session",
            "model": {"id": "claude-opus-5-5", "display_name": "Opus"},
            "effort": {"level": "high"},
            "thinking": {"enabled": true},
            "version": "2.1.330",
            "workspace": {"repo": {"host": "github.com", "owner": "owner", "name": "repo"}},
            "worktree": {"name": "wt-name"},
            "pr": {"number": 42, "url": "https://github.com/owner/repo/pull/42", "review_state": "pending"},
            "context_window": {"used_percentage": 25, "context_window_size": 200000},
            "rate_limits": {"seven_day": {"used_percentage": 21, "resets_at": NOW + 3_600}}
        })
    }

    fn with(key: &str, value: Value) -> Value {
        let mut payload = full();
        payload[key] = value;
        payload
    }

    fn without(key: &str) -> Value {
        let mut payload = full();
        payload.as_object_mut().unwrap().remove(key);
        payload
    }

    #[test]
    fn a_mistyped_field_changes_only_itself() {
        let home = tempfile::tempdir().unwrap();
        let h = home.path();
        let base = render_value(&full(), h);
        for item in [
            "🧠 Opus high·t",
            "📡 owner/repo",
            "🌳 wt-name",
            "🔗 #42 pending",
            "7D 21%",
        ] {
            assert!(base.contains(item), "{item}: {base}");
        }
        assert!(base.contains("🤖 v2.1.330"), "{base}");
        // A shown field of the wrong type: its item goes and every other item stays.
        let no_model = render_value(&with("model", json!(5)), h);
        assert_eq!(no_model, render_value(&without("model"), h));
        assert!(!no_model.contains("🧠"), "{no_model}");
        for item in ["📡 owner/repo", "🔗 #42 pending", "7D 21%", "🤖 v2.1.330"] {
            assert!(no_model.contains(item), "{item}: {no_model}");
        }
        // A field no item shows changes nothing.
        assert_eq!(render_value(&with("cost", json!("x")), h), base);
        // Inside an object only the mistyped field goes: a text number drops the request, a text
        // percentage says `n/a`.
        let mut pr = full();
        pr["pr"]["number"] = json!("42");
        let out = render_value(&pr, h);
        assert!(!out.contains("🔗"), "{out}");
        assert!(
            out.contains("🧠 Opus high·t") && out.contains("7D 21%"),
            "{out}"
        );
        let mut week = full();
        week["rate_limits"]["seven_day"]["used_percentage"] = json!("21");
        let out = render_value(&week, h);
        assert!(
            out.contains("7D n/a") && out.contains("🔗 #42 pending"),
            "{out}"
        );
    }

    /// A guard: fields the reader does not know never change the line.
    #[test]
    fn an_unknown_field_changes_nothing() {
        let home = tempfile::tempdir().unwrap();
        let mut payload = full();
        payload["future_field"] = json!({"nested": [1, {"x": null}], "flag": true});
        payload["prompt_id"] = json!("550e8400-e29b-41d4-a716-446655440000");
        payload["rate_limits"]["spend_limit"] =
            json!({"used_percentage": 62.8, "resets_at": NOW, "period": "monthly"});
        payload["pr"]["badge"] = json!(7);
        payload["model"]["future"] = json!([]);
        assert_eq!(
            render_value(&payload, home.path()),
            render_value(&full(), home.path())
        );
    }

    /// A key given twice keeps its last value whole, as serde_json's map keeps its last insert:
    /// two objects under one key are not merged. Keys match exactly.
    #[test]
    fn a_duplicated_key_keeps_its_last_value_whole() {
        let home = tempfile::tempdir().unwrap();
        let h = home.path();
        let out = render(
            br#"{"version":"1.0.0","model":{"display_name":"First"},"version":"2.1.330","model":{"id":"claude-sonnet-5-5"}}"#,
            h,
        );
        assert!(
            out.contains("🤖 v2.1.330") && !out.contains("1.0.0"),
            "{out}"
        );
        assert!(!out.contains("First"), "{out}");
        let twice = br#"{"context_window":{"used_percentage":40},"context_window":{"context_window_size":200000}}"#;
        let cw = read(twice).context_window.unwrap();
        assert_eq!(
            (cw.used_percentage, cw.context_window_size),
            (None, Some(200_000))
        );
        let out = render(twice, h);
        assert!(out.contains("CW n/a"), "{out}");
        assert!(read(br#"{"Model":{"display_name":"Folded"}}"#)
            .model
            .is_none());
        assert!(!render(br#"{"Model":{"display_name":"Folded"}}"#, h).contains("Folded"));
    }

    /// A guard: an array, a string, a number, `null`, a boolean, broken JSON and no bytes all
    /// read as the empty object, and the context window says `n/a`.
    #[test]
    fn a_payload_that_is_not_an_object_reads_as_empty() {
        let home = tempfile::tempdir().unwrap();
        let empty = render(b"{}", home.path());
        assert!(empty.contains("CW n/a"), "{empty}");
        for payload in [
            &b"[]"[..],
            br#"[{"model":"x"}]"#,
            br#""text""#,
            b"5",
            b"null",
            b"true",
            b"{ not json",
            b"",
        ] {
            let shown = String::from_utf8_lossy(payload);
            assert_eq!(render(payload, home.path()), empty, "{shown}");
            assert_eq!(
                format!("{:?}", read(payload)),
                format!("{:?}", Input::default()),
                "{shown}"
            );
        }
    }

    /// REQ-11: the transcript path, the output style, the agent name, the session duration, the
    /// cache tokens as a hit ratio and the prompt cache are not collected.
    #[test]
    fn the_removed_fields_are_not_read() {
        let home = tempfile::tempdir().unwrap();
        let payload = json!({
            "model": {"display_name": "Opus"},
            "transcript_path": "/TRANSCRIPT_SENTINEL.jsonl",
            "output_style": {"name": "STYLE_SENTINEL"},
            "agent": {"name": "AGENT_SENTINEL"},
            "cost": {"total_duration_ms": 4_242_421},
            "context_window": {
                "used_percentage": 25,
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 1,
                    "cache_creation_input_tokens": 6_060_601,
                    "cache_read_input_tokens": 7_070_701
                }
            },
            "prompt_cache": {"warm": true, "expires_at": NOW + 4_343 * 60, "ttl": "CACHE_SENTINEL"}
        });
        let input = read(payload.to_string().as_bytes());
        let status = state::collect(Some(&input), &ctx(home.path()));
        let (input, status) = (format!("{input:?}"), format!("{status:?}"));
        for value in [
            "TRANSCRIPT_SENTINEL",
            "STYLE_SENTINEL",
            "AGENT_SENTINEL",
            "CACHE_SENTINEL",
            "4242421",
        ] {
            assert!(!input.contains(value), "{value} read: {input}");
        }
        for value in ["STYLE_SENTINEL", "4242421", "6060601", "7070701", "4343"] {
            assert!(!status.contains(value), "{value} collected: {status}");
        }
        assert!(status.contains("Opus"), "{status}");
    }

    /// REQ-11: the goal file of the session is not read: planting an armed one changes nothing
    /// that is collected.
    #[test]
    fn a_planted_goal_file_changes_nothing() {
        let home = tempfile::tempdir().unwrap();
        let project = home.path().join("project");
        // Where the removed reader looked: the Go layout's goal folder, by the named constant.
        let goal = project.join(crate::state_root::MOAI_DIR).join("state/goal");
        std::fs::create_dir_all(&goal).unwrap();
        let payload = json!({
            "session_id": "goal-session",
            "model": {"display_name": "Opus"},
            "workspace": {"project_dir": project, "current_dir": project}
        });
        let input = read(payload.to_string().as_bytes());
        let collected = || format!("{:?}", state::collect(Some(&input), &ctx(home.path())));
        let before = collected();
        std::fs::write(goal.join("goal-session.json"), r#"{"status":"armed"}"#).unwrap();
        assert_eq!(collected(), before);
    }

    /// A field that is no longer read changes nothing even when it has the wrong type.
    #[test]
    fn a_mistyped_removed_field_changes_nothing() {
        let home = tempfile::tempdir().unwrap();
        let base = render_value(&full(), home.path());
        for (key, value) in [
            ("agent", json!({"name": 123})),
            ("output_style", json!(5)),
            ("transcript_path", json!(7)),
            ("cost", json!({"total_duration_ms": "long"})),
            ("prompt_cache", json!("warm")),
        ] {
            assert_eq!(render_value(&with(key, value), home.path()), base, "{key}");
        }
    }
}
