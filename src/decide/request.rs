//! The typed request of `moai decide` (card C-37 S2): one choice with its options, in the
//! fields every decision record names (SKILL.md), plus the evidence the model reads.
//!
//! Unknown fields are refused, so a request cannot carry an endpoint, a model name or a key.
//! `reversible` and `category` may be absent here: the gate, not the parser, refuses them, so
//! the refusal says which rule it was.

use super::Refusal;
use serde::Deserialize;

/// Fewest and most options of one choice (both providers document 2 to 255).
pub const OPTIONS_MIN: usize = 2;
pub const OPTIONS_MAX: usize = 255;
/// Longest option value, in bytes: a short name, not a sentence.
pub const VALUE_MAX: usize = 64;

/// One option: the short `value` the answer names and what it means.
#[derive(Deserialize, Debug, Clone, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Choice {
    pub value: String,
    pub description: String,
}

#[derive(Deserialize, Debug, Clone, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Request {
    /// The question, one choice.
    pub what: String,
    pub options: Vec<Choice>,
    /// The `value` that applies when nobody decides.
    pub default: String,
    #[serde(default)]
    pub reversible: Option<bool>,
    #[serde(default)]
    pub category: Option<String>,
    /// `YYYY-MM-DDTHH:MM:SSZ`.
    pub deadline: String,
    /// What the model reads to decide.
    pub evidence: String,
    /// The card the choice belongs to; recorded, never sent.
    #[serde(default)]
    pub card: Option<String>,
}

fn invalid(rule: &str) -> Refusal {
    Refusal::new("invalid_request", format!("invalid request: {rule}"))
}

/// Text a person wrote: something in it, and no control character but a line break or a tab.
fn is_text(s: &str) -> bool {
    !s.trim().is_empty() && !s.chars().any(|c| c.is_control() && c != '\n' && c != '\t')
}

/// A short name on one line.
fn is_name(s: &str, max: usize) -> bool {
    !s.is_empty() && s.len() <= max && s.trim() == s && !s.chars().any(char::is_control)
}

/// The request in `text` (one JSON object). The refusal names the rule that failed, never a
/// value of the request.
pub fn parse(text: &str) -> Result<Request, Refusal> {
    let request: Request = serde_json::from_str(text).map_err(|_| {
        invalid(
            "one JSON object with what, options, default, reversible, category, deadline, \
             evidence and optionally card; no other field",
        )
    })?;
    let rule = if !is_text(&request.what) {
        "`what` is text, not empty and without control characters"
    } else if !is_text(&request.evidence) {
        "`evidence` is text, not empty and without control characters"
    } else if !(OPTIONS_MIN..=OPTIONS_MAX).contains(&request.options.len()) {
        "`options` holds 2 to 255 choices"
    } else if !request.options.iter().all(|o| is_name(&o.value, VALUE_MAX)) {
        "an option value is a name of 1 to 64 bytes on one line, without spaces around it"
    } else if !request.options.iter().all(|o| is_text(&o.description)) {
        "an option's description is text, not empty and without control characters"
    } else if request
        .options
        .iter()
        .enumerate()
        .any(|(i, o)| request.options[..i].iter().any(|p| p.value == o.value))
    {
        "the option values are distinct"
    } else if !is_name(&request.default, VALUE_MAX) {
        "`default` is the value of an option"
    } else if crate::decision_file::deadline_secs(&request.deadline).is_none() {
        "`deadline` is YYYY-MM-DDTHH:MM:SSZ"
    } else if request
        .card
        .as_deref()
        .is_some_and(|c| !is_name(c, VALUE_MAX))
    {
        "`card` is a name of 1 to 64 bytes on one line"
    } else {
        return Ok(request);
    };
    Err(invalid(rule))
}

#[cfg(test)]
mod tests {
    use super::super::testing::request_json;
    use super::*;
    use serde_json::{json, Value};

    fn rule(text: &str) -> String {
        let r = parse(text).unwrap_err();
        assert_eq!(r.code, "invalid_request");
        r.reason
    }

    fn options(n: usize) -> Value {
        (0..n)
            .map(|i| json!({"value": format!("o{i}"), "description": format!("option {i}")}))
            .collect()
    }

    #[test]
    fn a_full_request_parses_and_card_is_optional() {
        let r = parse(&request_json(|_| {})).unwrap();
        assert_eq!(r.options.len(), 2);
        assert_eq!(r.options[1].value, "cargo-test");
        assert_eq!(r.reversible, Some(true));
        assert_eq!(r.category.as_deref(), Some("none"));
        assert_eq!(r.card.as_deref(), Some("C-37"));
        let without = parse(&request_json(|v| {
            v.as_object_mut().unwrap().remove("card");
        }))
        .unwrap();
        assert_eq!(without.card, None);
    }

    /// No field beyond the eight: a request cannot name where it goes or which model reads it.
    #[test]
    fn unknown_fields_are_refused_at_every_level() {
        for field in ["endpoint", "model", "provider", "url", "key", "reason"] {
            rule(&request_json(|v| v[field] = json!("x")));
        }
        rule(&request_json(|v| v["options"][0]["url"] = json!("x")));
        rule("[]");
        rule("");
        rule("{\"what\": \"only\"}");
    }

    /// The option count at both edges: 1 refused, 2 and 255 accepted, 256 refused.
    #[test]
    fn a_choice_has_two_to_255_options() {
        let with = |n: usize| {
            request_json(|v| {
                v["options"] = options(n);
                v["default"] = json!("o0");
            })
        };
        assert!(rule(&with(0)).contains("2 to 255"));
        assert!(rule(&with(1)).contains("2 to 255"));
        assert_eq!(parse(&with(2)).unwrap().options.len(), 2);
        assert_eq!(parse(&with(255)).unwrap().options.len(), 255);
        assert!(rule(&with(256)).contains("2 to 255"));
    }

    #[test]
    fn option_values_are_distinct_short_names_with_a_description() {
        let set = |value: Value, description: Value| {
            request_json(|v| {
                v["options"][1] = json!({"value": value, "description": description});
            })
        };
        assert!(rule(&set(json!("nextest"), json!("again"))).contains("distinct"));
        for bad in ["", " padded", "two\nlines", "tab\there"] {
            assert!(
                rule(&set(json!(bad), json!("d"))).contains("option value"),
                "{bad:?}"
            );
        }
        // The value length at its edge: 64 bytes accepted, 65 refused.
        assert!(parse(&set(json!("v".repeat(64)), json!("d"))).is_ok());
        assert!(rule(&set(json!("v".repeat(65)), json!("d"))).contains("option value"));
        for bad in ["", "  ", "bell\u{7}"] {
            assert!(
                rule(&set(json!("b"), json!(bad))).contains("description"),
                "{bad:?}"
            );
        }
    }

    #[test]
    fn the_texts_and_the_deadline_have_a_shape() {
        for field in ["what", "evidence"] {
            for bad in ["", " \n ", "esc\u{1b}[0m"] {
                assert!(
                    rule(&request_json(|v| v[field] = json!(bad))).contains(field),
                    "{field} {bad:?}"
                );
            }
            // More than one line is fine.
            assert!(parse(&request_json(|v| v[field] = json!("one\n\ttwo"))).is_ok());
        }
        for bad in [
            "",
            "tomorrow",
            "2026-11-18",
            "2026-02-31T00:00:00Z",
            "2026-11-18 00:00:00",
        ] {
            assert!(
                rule(&request_json(|v| v["deadline"] = json!(bad))).contains("deadline"),
                "{bad:?}"
            );
        }
        for bad in ["", "two\nlines", " C-37"] {
            assert!(
                rule(&request_json(|v| v["card"] = json!(bad))).contains("card"),
                "{bad:?}"
            );
        }
        assert!(rule(&request_json(|v| v["default"] = json!(""))).contains("default"));
    }

    /// The refusal names the rule and never a value of the request.
    #[test]
    fn a_refusal_quotes_nothing_from_the_request() {
        let marker = "MARKER_ONLY_NOT_A_SECRET";
        for text in [
            request_json(|v| v[marker] = json!(marker)),
            request_json(|v| v["deadline"] = json!(marker)),
            request_json(|v| v["options"][0]["value"] = json!(format!("{marker}\n"))),
            format!("{{\"what\": \"{marker}\""),
        ] {
            assert!(!rule(&text).contains(marker), "{text}");
        }
    }
}
