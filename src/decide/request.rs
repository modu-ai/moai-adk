//! The typed request of `moai decide` (card C-37, design v2 slice V1): one choice with its
//! options, in the fields every decision record names (SKILL.md), how the choice is undone,
//! its tier, and the evidence the model reads.
//!
//! Unknown fields are refused, so a request cannot carry an endpoint, a model name or a key.
//! `reversible` and `category` may be absent here: the gate, not the parser, refuses them, so
//! the refusal says which rule it was. Every limit refuses; nothing is cut, since a cut text
//! is another question.
//!
//! | Field | Rule | Code when broken |
//! | --- | --- | --- |
//! | `what` | text, at most 500 bytes | `invalid_request`, `what_too_long` |
//! | `options` | 2 to 8 | `options_count` |
//! | option `value` | a name of 1 to 64 bytes on one line with at least one letter; not `none_of_these` (any case), which the provider slice adds itself | `option_value`, `option_reserved` |
//! | option `description` | required text, at most 300 bytes | `description_missing`, `description_too_long` |
//! | `undo` | required, one line, at most 200 bytes: how the choice is reversed | `undo_missing`, `undo_invalid` |
//! | `tier` | required, 1 (undone inside the same pull request) or 2 (seen by others before it is undone) | `tier_missing`, `tier_invalid` |
//! | `evidence` | an object of 1 to 8 named fields, names `[a-z][a-z0-9_]{0,31}`, each value text of at most 4 KiB, all values together at most 16 KiB; a name given twice is refused | `evidence_count`, `evidence_name`, `evidence_value`, `evidence_field_too_large`, `evidence_too_large` |

use super::Refusal;
use serde::{Deserialize, Deserializer};
use std::collections::BTreeMap;

/// Fewest and most options of one choice (design v2: 2 to 8).
pub const OPTIONS_MIN: usize = 2;
pub const OPTIONS_MAX: usize = 8;
/// Longest option value, in bytes: a short name, not a sentence.
pub const VALUE_MAX: usize = 64;
/// Longest question, in bytes.
pub const WHAT_MAX: usize = 500;
/// Longest option description, in bytes.
pub const DESCRIPTION_MAX: usize = 300;
/// Longest `undo` line, in bytes.
pub const UNDO_MAX: usize = 200;
/// Most evidence fields.
pub const EVIDENCE_FIELDS_MAX: usize = 8;
/// Longest evidence field name, in bytes.
pub const EVIDENCE_NAME_MAX: usize = 32;
/// Largest evidence value, in bytes (4 KiB).
pub const EVIDENCE_FIELD_MAX: usize = 4 * 1024;
/// Largest sum of the evidence values, in bytes (16 KiB).
pub const EVIDENCE_MAX: usize = 16 * 1024;
/// The option value the provider slice appends as the way out; a request may not use it.
pub const RESERVED_VALUE: &str = "none_of_these";

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
    /// How the choice is reversed, in one line. Optional here so its absence has its own code.
    #[serde(default)]
    pub undo: Option<String>,
    /// 1 or 2. Optional here so its absence has its own code.
    #[serde(default)]
    pub tier: Option<u64>,
    /// What the model reads to decide: named fields, in name order.
    #[serde(deserialize_with = "distinct_fields")]
    pub evidence: BTreeMap<String, String>,
    /// The card the choice belongs to; recorded, never sent.
    #[serde(default)]
    pub card: Option<String>,
}

/// The evidence object, refusing a name given twice: the last one would silently replace
/// the first, and a dropped field is a cut.
fn distinct_fields<'de, D: Deserializer<'de>>(d: D) -> Result<BTreeMap<String, String>, D::Error> {
    struct Fields;
    impl<'de> serde::de::Visitor<'de> for Fields {
        type Value = BTreeMap<String, String>;
        fn expecting(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
            f.write_str("an object of named text fields")
        }
        fn visit_map<A: serde::de::MapAccess<'de>>(
            self,
            mut map: A,
        ) -> Result<Self::Value, A::Error> {
            let mut out = BTreeMap::new();
            while let Some((name, value)) = map.next_entry::<String, String>()? {
                if out.insert(name, value).is_some() {
                    return Err(serde::de::Error::custom("an evidence field is named twice"));
                }
            }
            Ok(out)
        }
    }
    d.deserialize_map(Fields)
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

/// An evidence field name: `[a-z][a-z0-9_]{0,31}`.
fn is_field_name(s: &str) -> bool {
    let mut chars = s.chars();
    chars.next().is_some_and(|c| c.is_ascii_lowercase())
        && s.len() <= EVIDENCE_NAME_MAX
        && chars.all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_')
}

/// The first rule of design v2 that `request` breaks, as its own code and the rule's words.
fn v2_rule(request: &Request) -> Option<(&'static str, &'static str)> {
    let options = &request.options;
    let evidence = &request.evidence;
    let rule = if request.what.len() > WHAT_MAX {
        ("what_too_long", "`what` is at most 500 bytes")
    } else if !(OPTIONS_MIN..=OPTIONS_MAX).contains(&options.len()) {
        ("options_count", "`options` holds 2 to 8 choices")
    } else if !options
        .iter()
        .all(|o| is_name(&o.value, VALUE_MAX) && o.value.chars().any(char::is_alphabetic))
    {
        (
            "option_value",
            "an option value is a meaningful name: 1 to 64 bytes on one line, without spaces \
             around it, with at least one letter",
        )
    } else if options
        .iter()
        .any(|o| o.value.eq_ignore_ascii_case(RESERVED_VALUE))
    {
        (
            "option_reserved",
            "`none_of_these` is reserved for the provider's way out and is not an option value",
        )
    } else if options.iter().any(|o| o.description.trim().is_empty()) {
        ("description_missing", "every option has a description")
    } else if options
        .iter()
        .any(|o| o.description.len() > DESCRIPTION_MAX)
    {
        (
            "description_too_long",
            "an option's description is at most 300 bytes",
        )
    } else if request.undo.is_none() {
        (
            "undo_missing",
            "`undo` is required: one line saying how the choice is reversed",
        )
    } else if !request
        .undo
        .as_deref()
        .is_some_and(|u| is_text(u) && !u.contains('\n') && u.len() <= UNDO_MAX)
    {
        (
            "undo_invalid",
            "`undo` is one line of text, not empty, at most 200 bytes",
        )
    } else if request.tier.is_none() {
        ("tier_missing", "`tier` is required: 1 or 2")
    } else if !matches!(request.tier, Some(1 | 2)) {
        (
            "tier_invalid",
            "`tier` is 1 (undone inside the same pull request) or 2 (seen by others before \
             it is undone)",
        )
    } else if !(1..=EVIDENCE_FIELDS_MAX).contains(&evidence.len()) {
        ("evidence_count", "`evidence` holds 1 to 8 named fields")
    } else if !evidence.keys().all(|name| is_field_name(name)) {
        (
            "evidence_name",
            "an evidence field name is [a-z][a-z0-9_]{0,31}",
        )
    } else if !evidence.values().all(|value| is_text(value)) {
        (
            "evidence_value",
            "an evidence value is text, not empty and without control characters",
        )
    } else if evidence
        .values()
        .any(|value| value.len() > EVIDENCE_FIELD_MAX)
    {
        (
            "evidence_field_too_large",
            "an evidence value is at most 4 KiB: refused, never cut",
        )
    } else if evidence.values().map(String::len).sum::<usize>() > EVIDENCE_MAX {
        (
            "evidence_too_large",
            "the evidence values together are at most 16 KiB: refused, never cut",
        )
    } else {
        return None;
    };
    Some(rule)
}

/// The request in `text` (one JSON object). The refusal names the rule that failed, never a
/// value of the request.
pub fn parse(text: &str) -> Result<Request, Refusal> {
    let request: Request = serde_json::from_str(text).map_err(|_| {
        invalid(
            "one JSON object with what, options, default, reversible, category, deadline, \
             undo, tier, evidence (an object of named text fields) and optionally card; no \
             other field",
        )
    })?;
    let rule = if !is_text(&request.what) {
        "`what` is text, not empty and without control characters"
    } else if let Some((code, rule)) = v2_rule(&request) {
        return Err(Refusal::new(code, format!("invalid request: {rule}")));
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
    use super::super::testing::{evidence, request_json};
    use super::*;
    use serde_json::{json, Value};

    fn rule(text: &str) -> String {
        let r = parse(text).unwrap_err();
        assert_eq!(r.code, "invalid_request");
        r.reason
    }

    /// The code of the refusal of `edit`, or `parsed`.
    fn code(edit: impl FnOnce(&mut Value)) -> &'static str {
        match parse(&request_json(edit)) {
            Ok(_) => "parsed",
            Err(r) => r.code,
        }
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
        assert_eq!(
            r.undo.as_deref(),
            Some("set the runner line in ci.yml back to the other runner")
        );
        assert_eq!(r.tier, Some(1));
        assert_eq!(
            r.evidence.into_iter().collect::<Vec<_>>(),
            [(
                "ci".to_string(),
                "CI runs nextest; the local docs name both runners.".to_string()
            )]
        );
        let without = parse(&request_json(|v| {
            v.as_object_mut().unwrap().remove("card");
        }))
        .unwrap();
        assert_eq!(without.card, None);
    }

    /// No field beyond the ten: a request cannot name where it goes or which model reads it.
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

    /// The option count at both edges: 1 refused, 2 and 8 accepted, 9 refused.
    #[test]
    fn a_choice_has_two_to_eight_options() {
        assert_eq!((OPTIONS_MIN, OPTIONS_MAX), (2, 8));
        let with = |n: usize| {
            request_json(|v| {
                v["options"] = options(n);
                v["default"] = json!("o0");
            })
        };
        for n in [0, 1, 9, 255] {
            let r = parse(&with(n)).unwrap_err();
            assert_eq!(r.code, "options_count", "{n}");
            assert!(r.reason.contains("2 to 8"), "{}", r.reason);
        }
        assert_eq!(parse(&with(2)).unwrap().options.len(), 2);
        assert_eq!(parse(&with(8)).unwrap().options.len(), 8);
    }

    #[test]
    fn option_values_are_distinct_meaningful_names() {
        let set = |value: Value| {
            request_json(|v| {
                v["options"][1] = json!({"value": value, "description": "d"});
            })
        };
        assert!(rule(&set(json!("nextest"))).contains("distinct"));
        for bad in [
            "",
            " padded",
            "two\nlines",
            "tab\there",
            "1",
            "2",
            "--",
            "#3",
        ] {
            let r = parse(&set(json!(bad))).unwrap_err();
            assert_eq!(r.code, "option_value", "{bad:?}");
        }
        // A letter in any script is a letter.
        assert!(parse(&set(json!("예"))).is_ok());
        assert!(parse(&set(json!("v2"))).is_ok());
        // The value length at its edge: 64 bytes accepted, 65 refused.
        assert!(parse(&set(json!("v".repeat(64)))).is_ok());
        assert_eq!(
            parse(&set(json!("v".repeat(65)))).unwrap_err().code,
            "option_value"
        );
    }

    /// `none_of_these` is the provider's way out; as an option of the request it is refused,
    /// in any letter case. A value that only contains it is a value.
    #[test]
    fn the_reserved_value_is_refused() {
        assert_eq!(RESERVED_VALUE, "none_of_these");
        for reserved in ["none_of_these", "None_Of_These", "NONE_OF_THESE"] {
            assert_eq!(
                code(|v| v["options"][1]["value"] = json!(reserved)),
                "option_reserved",
                "{reserved}"
            );
        }
        for near in ["none_of_these2", "none-of-these", "none_of_this"] {
            assert_eq!(
                code(|v| v["options"][1]["value"] = json!(near)),
                "parsed",
                "{near}"
            );
        }
    }

    /// Slice V2: the reserved value is compared after the normaliser, so a character without
    /// width, full-width letters or the long s do not hide it; a bidi character in a value is
    /// refused with its own code.
    #[test]
    fn the_reserved_value_is_compared_after_the_normaliser() {
        let wide: String = "none"
            .chars()
            .map(|c| char::from_u32(c as u32 + 0xfee0).unwrap())
            .collect();
        for hidden in [
            "none\u{200b}_of_these".to_string(),
            "\u{feff}none_of_these".to_string(),
            format!("{wide}_of_these"),
            "none_of_the\u{17f}e".to_string(),
            "none_of_these\u{fe0f}".to_string(),
        ] {
            assert_eq!(
                code(|v| v["options"][1]["value"] = json!(hidden)),
                "option_reserved",
                "{hidden}"
            );
        }
        assert_eq!(
            code(|v| v["options"][1]["value"] = json!("none\u{202e}_of_these")),
            "bidi_character"
        );
    }

    /// A description is required, at most 300 bytes: 300 accepted, 301 refused.
    #[test]
    fn a_description_is_required_and_short() {
        assert_eq!(DESCRIPTION_MAX, 300);
        for empty in ["", "  ", "\n"] {
            assert_eq!(
                code(|v| v["options"][1]["description"] = json!(empty)),
                "description_missing",
                "{empty:?}"
            );
        }
        // V1 audit N10: an absent description has the same code as an empty one.
        assert_eq!(
            code(|v| {
                v["options"][1]
                    .as_object_mut()
                    .unwrap()
                    .remove("description");
            }),
            "description_missing"
        );
        assert_eq!(
            code(|v| v["options"][1]["description"] = Value::Null),
            "invalid_request"
        );
        assert_eq!(
            code(|v| v["options"][1]["description"] = json!("d".repeat(300))),
            "parsed"
        );
        assert_eq!(
            code(|v| v["options"][1]["description"] = json!("d".repeat(301))),
            "description_too_long"
        );
        // A control character is still the old text rule.
        assert_eq!(
            code(|v| v["options"][1]["description"] = json!("bell\u{7}")),
            "invalid_request"
        );
    }

    /// `what` at its edge: 500 bytes accepted, 501 refused.
    #[test]
    fn the_question_is_at_most_500_bytes() {
        assert_eq!(WHAT_MAX, 500);
        assert_eq!(code(|v| v["what"] = json!("w".repeat(500))), "parsed");
        assert_eq!(
            code(|v| v["what"] = json!("w".repeat(501))),
            "what_too_long"
        );
    }

    /// `undo`: required, one line of text, at most 200 bytes. V1 audit N7: each broken rule
    /// has its own code; every Unicode line terminator counts as a second line.
    #[test]
    fn undo_is_one_required_line() {
        assert_eq!(UNDO_MAX, 200);
        let missing = code(|v| {
            v.as_object_mut().unwrap().remove("undo");
        });
        assert_eq!(missing, "undo_missing");
        assert_eq!(code(|v| v["undo"] = Value::Null), "undo_missing");
        for blank in ["", "   ", "\t"] {
            assert_eq!(
                code(|v| v["undo"] = json!(blank)),
                "undo_missing",
                "{blank:?}"
            );
        }
        for two in [
            "two\nlines",
            "cr\rhere",
            "crlf\r\nhere",
            "vt\u{b}here",
            "ff\u{c}here",
            "nel\u{85}here",
            "ls\u{2028}here",
            "ps\u{2029}here",
            "end\n",
        ] {
            assert_eq!(
                code(|v| v["undo"] = json!(two)),
                "undo_multiline",
                "{two:?}"
            );
        }
        assert_eq!(
            code(|v| v["undo"] = json!("esc\u{1b}[0m")),
            "undo_invalid"
        );
        assert_eq!(code(|v| v["undo"] = json!("u".repeat(200))), "parsed");
        assert_eq!(
            code(|v| v["undo"] = json!("u".repeat(201))),
            "undo_too_long"
        );
        // A long text on two lines: the line rule answers first.
        assert_eq!(
            code(|v| v["undo"] = json!(format!("{}\n", "u".repeat(201)))),
            "undo_multiline"
        );
        assert_eq!(code(|v| v["undo"] = json!(1)), "invalid_request");
    }

    /// `tier`: required, 1 or 2.
    #[test]
    fn tier_is_one_or_two() {
        assert_eq!(code(|v| v["tier"] = json!(1)), "parsed");
        assert_eq!(code(|v| v["tier"] = json!(2)), "parsed");
        let missing = code(|v| {
            v.as_object_mut().unwrap().remove("tier");
        });
        assert_eq!(missing, "tier_missing");
        assert_eq!(code(|v| v["tier"] = Value::Null), "tier_missing");
        for bad in [0, 3, 9] {
            assert_eq!(code(|v| v["tier"] = json!(bad)), "tier_invalid", "{bad}");
        }
        for bad in [json!("1"), json!(1.5), json!(-1), json!(true)] {
            assert_eq!(
                code(|v| v["tier"] = bad.clone()),
                "invalid_request",
                "{bad}"
            );
        }
    }

    /// `evidence` is an object of 1 to 8 named fields: 0 and 9 refused, 1 and 8 accepted.
    /// A plain string (the first design) is no longer a request.
    #[test]
    fn evidence_is_one_to_eight_named_fields() {
        assert_eq!(EVIDENCE_FIELDS_MAX, 8);
        let fields = |n: usize| -> Value {
            (0..n)
                .map(|i| (format!("f{i}"), json!("text")))
                .collect::<serde_json::Map<_, _>>()
                .into()
        };
        assert_eq!(code(|v| v["evidence"] = fields(0)), "evidence_count");
        assert_eq!(code(|v| v["evidence"] = fields(1)), "parsed");
        assert_eq!(code(|v| v["evidence"] = fields(8)), "parsed");
        assert_eq!(code(|v| v["evidence"] = fields(9)), "evidence_count");
        assert_eq!(
            code(|v| v["evidence"] = json!("plain text")),
            "invalid_request"
        );
        assert_eq!(code(|v| v["evidence"] = json!({"a": 1})), "invalid_request");
        assert_eq!(code(|v| v["evidence"] = json!(["a"])), "invalid_request");
        // A name given twice would drop a field: refused with its own code (V1 audit N6),
        // and the reason quotes neither the name nor a value.
        let text = request_json(|_| {}).replace(
            r#""evidence":{"ci":"#,
            r#""evidence":{"ci":"MARKER_FIRST","ci":"#,
        );
        assert!(text.contains(r#""ci":"MARKER_FIRST","ci":"#), "{text}");
        let r = parse(&text).unwrap_err();
        assert_eq!(r.code, "evidence_duplicate");
        assert!(!r.reason.contains("MARKER") && !r.reason.contains("\"ci\""));
        // A field named twice elsewhere in the text is not this rule.
        let elsewhere = request_json(|_| {}).replace(r#""tier":1"#, r#""tier":1,"tier":1"#);
        assert_eq!(parse(&elsewhere).unwrap_err().code, "invalid_request");
    }

    #[test]
    fn evidence_names_are_short_lower_case_identifiers() {
        assert_eq!(EVIDENCE_NAME_MAX, 32);
        for good in ["a", "ci", "ci_log", "f9", &format!("a{}", "b".repeat(31))] {
            assert_eq!(
                code(|v| v["evidence"] = json!({ good: "x" })),
                "parsed",
                "{good}"
            );
        }
        for bad in [
            "",
            "Ci",
            "9f",
            "_a",
            "ci-log",
            "ci log",
            "é",
            &format!("a{}", "b".repeat(32)),
        ] {
            assert_eq!(
                code(|v| v["evidence"] = json!({ bad: "x" })),
                "evidence_name",
                "{bad:?}"
            );
        }
    }

    #[test]
    fn evidence_values_are_text() {
        for bad in ["", " \n ", "esc\u{1b}[0m"] {
            assert_eq!(
                code(|v| v["evidence"] = evidence(bad)),
                "evidence_value",
                "{bad:?}"
            );
        }
        assert_eq!(code(|v| v["evidence"] = evidence("one\n\ttwo")), "parsed");
    }

    /// Each value at most 4 KiB (4,096 accepted, 4,097 refused); all values together at most
    /// 16 KiB (16,384 accepted, 16,385 refused). Refused, never cut.
    #[test]
    fn evidence_has_a_field_bound_and_a_total_bound() {
        assert_eq!((EVIDENCE_FIELD_MAX, EVIDENCE_MAX), (4_096, 16_384));
        assert_eq!(
            code(|v| v["evidence"] = evidence("e".repeat(4_096))),
            "parsed"
        );
        assert_eq!(
            code(|v| v["evidence"] = evidence("e".repeat(4_097))),
            "evidence_field_too_large"
        );
        // A multi-byte character counts by its bytes: 1,366 three-byte letters are 4,098.
        assert_eq!(
            code(|v| v["evidence"] = evidence("가".repeat(1_366))),
            "evidence_field_too_large"
        );
        let four = |last: usize| {
            json!({
                "a": "e".repeat(4_096),
                "b": "e".repeat(4_096),
                "c": "e".repeat(4_096),
                "d": "e".repeat(last),
            })
        };
        assert_eq!(code(|v| v["evidence"] = four(4_096)), "parsed");
        let mut five = four(4_096);
        five["e"] = json!("e");
        assert_eq!(code(|v| v["evidence"] = five), "evidence_too_large");
        let mut at = four(4_095);
        at["e"] = json!("e");
        assert_eq!(code(|v| v["evidence"] = at), "parsed");
    }

    #[test]
    fn the_texts_and_the_deadline_have_a_shape() {
        for bad in ["", " \n ", "esc\u{1b}[0m"] {
            assert!(
                rule(&request_json(|v| v["what"] = json!(bad))).contains("what"),
                "{bad:?}"
            );
        }
        // More than one line is fine.
        assert!(parse(&request_json(|v| v["what"] = json!("one\n\ttwo"))).is_ok());
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
            request_json(|v| v["undo"] = json!(format!("{marker}\n"))),
            request_json(|v| v["evidence"] = json!({ marker: "x" })),
            request_json(|v| v["evidence"] = evidence(marker.repeat(200))),
            format!("{{\"what\": \"{marker}\""),
        ] {
            let r = parse(&text).unwrap_err();
            assert!(!r.reason.contains(marker), "{text}");
        }
    }
}
