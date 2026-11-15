//! `session-current` (#3015 C4): this session's own id, read from the environment only.
//!
//! The Go build's `session current` falls back to a per-project file that every session start
//! overwrites, so a lane that asks for its own id gets the leader's. This reads only the
//! variables a host sets for the session it starts, and never that file.

use serde_json::{json, Value};

/// Variables that carry the session's own id, strongest first. The first is set by Claude
/// Code. `CODEX_THREAD_ID` is what Codex puts in the terminal environment of a thread; it is
/// not in Codex's documentation, so a Codex lane has to measure it before it is relied on.
pub const SOURCES: [&str; 2] = ["CLAUDE_CODE_SESSION_ID", "CODEX_THREAD_ID"];

/// The line the Go build prints when it has no id; kept word for word so a reader that
/// already matches it keeps working.
pub const FALLBACK: &str = "source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>";

/// A usable id is UUID-shaped: hex digits and `-`, 8 to 64 characters.
pub fn well_formed(id: &str) -> bool {
    (8..=64).contains(&id.len()) && id.bytes().all(|b| b.is_ascii_hexdigit() || b == b'-')
}

/// Resolve from `get`, which reads one variable (the process environment in the command).
pub fn resolve(get: impl Fn(&str) -> Option<String>) -> Value {
    let mut ignored = Vec::new();
    for name in SOURCES {
        let Some(raw) = get(name) else { continue };
        let value = raw.trim();
        if value.is_empty() {
            continue;
        }
        if !well_formed(value) {
            ignored.push(name);
            continue;
        }
        return json!({
            "source_session_id": value,
            "source": format!("env:{name}"),
            "ignored": ignored,
        });
    }
    json!({
        "source_session_id": null,
        "source": "none",
        "ignored": ignored,
        "fallback": FALLBACK,
        "hint": "No host variable carries this session's id. Use the id the session-start hook printed; the per-project record is not this session's id.",
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    const A: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";
    const B: &str = "4d5550e0-88b3-40ca-8c46-07425eedc812";

    fn env(pairs: &[(&str, &str)]) -> impl Fn(&str) -> Option<String> {
        let m: HashMap<String, String> = pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_string()))
            .collect();
        move |k| m.get(k).cloned()
    }

    #[test]
    fn the_claude_variable_is_the_session_id() {
        let v = resolve(env(&[("CLAUDE_CODE_SESSION_ID", A)]));
        assert_eq!(v["source_session_id"], A);
        assert_eq!(v["source"], "env:CLAUDE_CODE_SESSION_ID");
        assert!(v.get("fallback").is_none());
    }

    #[test]
    fn the_codex_variable_is_used_when_the_claude_one_is_absent() {
        let v = resolve(env(&[("CODEX_THREAD_ID", B)]));
        assert_eq!(v["source_session_id"], B);
        assert_eq!(v["source"], "env:CODEX_THREAD_ID");
    }

    #[test]
    fn the_claude_variable_wins_over_the_codex_one() {
        let v = resolve(env(&[
            ("CODEX_THREAD_ID", B),
            ("CLAUDE_CODE_SESSION_ID", A),
        ]));
        assert_eq!(v["source_session_id"], A);
    }

    #[test]
    fn nothing_set_gives_null_the_fallback_line_and_a_hint() {
        let v = resolve(env(&[]));
        assert!(v["source_session_id"].is_null());
        assert_eq!(v["source"], "none");
        assert_eq!(v["fallback"], FALLBACK);
        assert!(v["hint"].as_str().unwrap().contains("session-start hook"));
        assert_eq!(v["ignored"], serde_json::json!([]));
    }

    #[test]
    fn a_value_that_is_not_an_id_is_ignored_by_name_and_never_echoed() {
        let v = resolve(env(&[
            ("CLAUDE_CODE_SESSION_ID", "not an id; rm -rf /"),
            ("CODEX_THREAD_ID", B),
        ]));
        assert_eq!(v["source_session_id"], B);
        assert_eq!(v["ignored"], serde_json::json!(["CLAUDE_CODE_SESSION_ID"]));
        assert!(!v.to_string().contains("rm -rf"));
    }

    #[test]
    fn an_empty_or_blank_value_is_absent_not_ignored() {
        let v = resolve(env(&[
            ("CLAUDE_CODE_SESSION_ID", "  "),
            ("CODEX_THREAD_ID", ""),
        ]));
        assert!(v["source_session_id"].is_null());
        assert_eq!(v["ignored"], serde_json::json!([]));
    }

    #[test]
    fn surrounding_whitespace_is_trimmed() {
        let v = resolve(env(&[("CLAUDE_CODE_SESSION_ID", &format!("  {A}\n"))]));
        assert_eq!(v["source_session_id"], A);
    }

    #[test]
    fn the_id_shape_has_exact_length_edges() {
        assert!(!well_formed(&"a".repeat(7)));
        assert!(well_formed(&"a".repeat(8)));
        assert!(well_formed(&"a".repeat(64)));
        assert!(!well_formed(&"a".repeat(65)));
        assert!(!well_formed("zzzzzzzz"));
        assert!(well_formed("0189abcd-ef01-7234-8abc-0123456789ab"));
    }
}
