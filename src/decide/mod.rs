//! `moai decide`, the pure half (card C-37 S2): a reversible, category-`none` choice may be
//! put to a decision model instead of the leader. Nothing here opens a socket, starts a
//! process or reads a key; the round trip is a later slice (S4) and uses only what this
//! module hands it.
//!
//! The order is fixed, and the types keep it: a [`request::Request`] becomes a
//! [`gate::Cleared`] only through [`gate::check`], a `Cleared` becomes an
//! [`outbound::Outbound`] only through [`outbound::build`], and the provider code
//! ([`provider::request_body`]) takes nothing but an `Outbound`. So a request the gate or the
//! secret screen refuses never reaches a line that knows a provider.
//!
//! - [`gate`] refuses a choice that is not the model's to make: not reversible, no category,
//!   an unknown one, `guarded`, one of the five HOLD categories, a default that is not an
//!   option, or text that reads as one of those classes while `none` is declared.
//! - [`outbound`] is the text that would leave the machine. A credential shape in it refuses
//!   the request (user decision 2026-11-16, D-secret-hit: refuse and fall back, never mask and
//!   send); the home folder and the user name are masked.
//! - [`provider`] builds the exact request bytes for `typesafe` (Jev) and `openai`, and reads
//!   their answers. The endpoints are compile-time constants; no setting, variable or request
//!   field can aim a request elsewhere.
//!
//! Every [`Refusal`] means the same thing to the caller: the model does not decide, the
//! leader does (`fallback: leader`). A [`Record`] carries the fields every decision record
//! names (SKILL.md) and who decided, `model:<provider>/<model the response reported>`; the
//! model gives no written reason and none is recorded (leader default D-reason).

pub mod gate;
pub mod outbound;
pub mod provider;
pub mod request;

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::BTreeMap;

/// Why the model does not decide this request. `code` is fixed and machine-readable; `reason`
/// is a sentence built from fixed words and rule names, never from the request's own text.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Refusal {
    pub code: &'static str,
    pub reason: String,
}

impl Refusal {
    pub(crate) fn new(code: &'static str, reason: impl Into<String>) -> Self {
        Refusal {
            code,
            reason: reason.into(),
        }
    }

    /// The answer a caller gives instead of a decision: the leader decides.
    pub fn to_json(&self) -> Value {
        json!({"code": self.code, "reason": self.reason, "fallback": "leader"})
    }
}

/// A request that passed the gate and the secret screen, with the bytes that would be sent.
#[derive(Debug)]
pub struct Prepared {
    pub provider: provider::Provider,
    /// Where the bytes would go: one of the two compiled endpoints.
    pub endpoint: &'static str,
    /// The request body, exactly as it would be sent.
    pub body: Vec<u8>,
    pub outbound: outbound::Outbound,
    reversible: bool,
    category: String,
    default: String,
    deadline: String,
    card: Option<String>,
}

/// One model decision, as recorded. The fields are the ones every decision record names
/// (what, options, default, reversible, category, deadline, who decided) plus the answer.
/// The texts are the masked ones that were sent. There is no reason field.
#[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Record {
    pub schema_version: u32,
    pub what: String,
    pub options: Vec<String>,
    pub default: String,
    pub reversible: bool,
    pub category: String,
    pub deadline: String,
    /// `model:<provider>/<model>`; the model is the one the response reported, never the
    /// alias the request asked for.
    pub decided_by: String,
    pub chosen: String,
    #[serde(default)]
    pub confidence: Option<f64>,
    #[serde(default)]
    pub probabilities: BTreeMap<String, f64>,
    #[serde(default)]
    pub card: Option<String>,
}

/// The request text `text` (JSON), checked and turned into the bytes for `provider`. Any
/// refusal comes before the provider code runs.
pub fn prepare(
    text: &str,
    provider: provider::Provider,
    who: &outbound::Who,
) -> Result<Prepared, Refusal> {
    prepare_with(text, provider, who, provider::request_body)
}

/// [`prepare`] with the provider's body builder as an argument, so a test can show the
/// builder was never called for a refused request.
fn prepare_with(
    text: &str,
    provider: provider::Provider,
    who: &outbound::Who,
    build: impl FnOnce(provider::Provider, &outbound::Outbound) -> Vec<u8>,
) -> Result<Prepared, Refusal> {
    let request = request::parse(text)?;
    let cleared = gate::check(&request)?;
    let outbound = outbound::build(&cleared, who)?;
    let body = build(provider, &outbound);
    Ok(Prepared {
        provider,
        endpoint: provider.endpoint(),
        body,
        outbound,
        reversible: request.reversible == Some(true),
        category: request.category.clone().unwrap_or_default(),
        default: request.default.clone(),
        deadline: request.deadline.clone(),
        card: request.card.clone(),
    })
}

/// The record of the provider's `response` to `prepared`, or the refusal that sends the
/// choice back to the leader (a provider refusal, an answer that is not one of the options,
/// a response that cannot be read).
pub fn conclude(prepared: &Prepared, response: &[u8]) -> Result<Record, Refusal> {
    let answer = provider::parse_answer(prepared.provider, &prepared.outbound, response)?;
    Ok(Record {
        schema_version: 1,
        what: prepared.outbound.what().to_string(),
        options: prepared
            .outbound
            .options()
            .iter()
            .map(|(value, _)| value.clone())
            .collect(),
        default: prepared.default.clone(),
        reversible: prepared.reversible,
        category: prepared.category.clone(),
        deadline: prepared.deadline.clone(),
        decided_by: format!("model:{}/{}", prepared.provider.name(), answer.model),
        chosen: answer.choice,
        confidence: answer.confidence,
        probabilities: answer.probabilities,
        card: prepared.card.clone(),
    })
}

#[cfg(test)]
pub(crate) mod testing {
    use super::*;

    /// A request that passes every check; `edit` changes it before it is serialized.
    pub fn request_json(edit: impl FnOnce(&mut Value)) -> String {
        let mut v = json!({
            "what": "Which test runner does the new module use?",
            "options": [
                {"value": "nextest", "description": "cargo nextest, as CI runs it"},
                {"value": "cargo-test", "description": "plain cargo test"}
            ],
            "default": "nextest",
            "reversible": true,
            "category": "none",
            "deadline": "2026-11-18T00:00:00Z",
            "evidence": "CI runs nextest; the local docs name both runners.",
            "card": "C-37"
        });
        edit(&mut v);
        v.to_string()
    }

    /// Test data: a made-up person, never the one running the tests.
    pub fn who() -> outbound::Who {
        outbound::Who {
            home: Some("/Users/testperson".into()),
            user: Some("testperson".into()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::testing::{request_json, who};
    use super::*;
    use provider::Provider;
    use std::cell::Cell;

    /// The refusal of `text`, after showing the provider's builder was not called.
    fn refused(text: &str) -> Refusal {
        for provider in [Provider::Typesafe, Provider::Openai] {
            let reached = Cell::new(false);
            let result = prepare_with(text, provider, &who(), |_, _| {
                reached.set(true);
                Vec::new()
            });
            assert!(result.is_err(), "not refused: {text}");
            assert!(!reached.get(), "the provider builder ran for: {text}");
        }
        prepare(text, Provider::Typesafe, &who()).unwrap_err()
    }

    #[test]
    fn a_good_request_reaches_the_builder_once_per_call() {
        let calls = Cell::new(0);
        let p = prepare_with(&request_json(|_| {}), Provider::Openai, &who(), |_, _| {
            calls.set(calls.get() + 1);
            b"body".to_vec()
        })
        .unwrap();
        assert_eq!(calls.get(), 1);
        assert_eq!(p.body, b"body");
        assert_eq!(p.endpoint, "https://api.openai.com/v1/decisions");
    }

    /// One case per category, literal: only `none` passes. Each refusal comes before the
    /// provider's builder.
    #[test]
    fn every_category_but_none_is_refused_before_the_provider() {
        for (category, code) in [
            ("guarded", "category_guarded"),
            ("money", "category_hold"),
            ("legal", "category_hold"),
            ("external_contract", "category_hold"),
            ("secret", "category_hold"),
            ("real_data", "category_hold"),
        ] {
            let r = refused(&request_json(|v| v["category"] = json!(category)));
            assert_eq!(r.code, code, "{category}");
            assert!(r.reason.contains(category), "{}", r.reason);
        }
        // The seven literals above and `none` are the whole shared list.
        assert_eq!(crate::decision_file::CATEGORIES.len(), 7);
        assert!(prepare(&request_json(|_| {}), Provider::Typesafe, &who()).is_ok());
    }

    #[test]
    fn a_missing_or_unknown_category_is_refused_before_the_provider() {
        let missing = refused(&request_json(|v| {
            v.as_object_mut().unwrap().remove("category");
        }));
        assert_eq!(missing.code, "category_missing");
        let null = refused(&request_json(|v| v["category"] = Value::Null));
        assert_eq!(null.code, "category_missing");
        for unknown in ["", "None", "trivial", "none ", "external transmission"] {
            let r = refused(&request_json(|v| v["category"] = json!(unknown)));
            assert_eq!(r.code, "category_unknown", "{unknown:?}");
        }
    }

    #[test]
    fn a_choice_that_is_not_reversible_is_refused_before_the_provider() {
        let no = refused(&request_json(|v| v["reversible"] = json!(false)));
        assert_eq!(no.code, "irreversible");
        let missing = refused(&request_json(|v| {
            v.as_object_mut().unwrap().remove("reversible");
        }));
        assert_eq!(missing.code, "irreversible");
        // Anything but the JSON value `true` is not a yes.
        for other in [json!("true"), json!(1), json!("yes")] {
            let r = refused(&request_json(|v| v["reversible"] = other.clone()));
            assert_eq!(r.code, "invalid_request", "{other}");
        }
    }

    #[test]
    fn a_default_that_is_not_an_option_is_refused_before_the_provider() {
        let r = refused(&request_json(|v| v["default"] = json!("pytest")));
        assert_eq!(r.code, "default_not_an_option");
    }

    #[test]
    fn text_that_reads_as_a_hold_class_under_none_is_refused_before_the_provider() {
        let r = refused(&request_json(|v| {
            v["what"] = json!("Which paid plan do we take for the CI minutes?")
        }));
        assert_eq!(r.code, "class_mismatch");
        assert!(r.reason.contains("money"), "{}", r.reason);
    }

    #[test]
    fn a_secret_shape_is_refused_before_the_provider_and_is_not_quoted() {
        // Test data: built at run time, shaped like a key and not one.
        let fake = format!("sk-{}", "a1".repeat(12));
        let r = refused(&request_json(|v| {
            v["evidence"] = json!(format!("the log shows {fake} in the header"))
        }));
        assert_eq!(r.code, "secret_shape");
        assert!(!r.reason.contains(&fake));
        assert!(!r.to_json().to_string().contains(&fake));
    }

    #[test]
    fn a_request_that_does_not_parse_is_refused_before_the_provider() {
        assert_eq!(refused("{").code, "invalid_request");
        let extra = refused(&request_json(|v| {
            v["endpoint"] = json!("https://example.invalid")
        }));
        assert_eq!(extra.code, "invalid_request");
    }

    #[test]
    fn every_refusal_falls_back_to_the_leader() {
        let r = refused(&request_json(|v| v["category"] = json!("money")));
        assert_eq!(
            r.to_json(),
            json!({"code": "category_hold", "reason": r.reason, "fallback": "leader"})
        );
    }

    fn jev_answer(choice: &str) -> Vec<u8> {
        json!({
            "model": "jev-1.14.0",
            "answers": {"decision": {
                "type": "choice", "choice": choice, "confidence": 0.875,
                "probabilities": {"nextest": 0.875, "cargo-test": 0.125}
            }},
            "usage": {"input_tokens": 61}
        })
        .to_string()
        .into_bytes()
    }

    /// The record names the fields of every decision record and who decided: the provider and
    /// the model the response reported, not the alias that was asked for. No reason field.
    #[test]
    fn the_record_names_the_model_the_response_reported() {
        let p = prepare(&request_json(|_| {}), Provider::Typesafe, &who()).unwrap();
        let record = conclude(&p, &jev_answer("cargo-test")).unwrap();
        assert_eq!(
            serde_json::to_value(&record).unwrap(),
            json!({
                "schema_version": 1,
                "what": "Which test runner does the new module use?",
                "options": ["nextest", "cargo-test"],
                "default": "nextest",
                "reversible": true,
                "category": "none",
                "deadline": "2026-11-18T00:00:00Z",
                "decided_by": "model:typesafe/jev-1.14.0",
                "chosen": "cargo-test",
                "confidence": 0.875,
                "probabilities": {"cargo-test": 0.125, "nextest": 0.875},
                "card": "C-37"
            })
        );
        // The record reads back, and a field it does not know (a written reason) does not.
        let text = serde_json::to_string(&record).unwrap();
        assert_eq!(serde_json::from_str::<Record>(&text).unwrap(), record);
        let mut with_reason = serde_json::to_value(&record).unwrap();
        with_reason["reason"] = json!("because");
        assert!(serde_json::from_value::<Record>(with_reason).is_err());
    }

    #[test]
    fn an_answer_outside_the_options_makes_no_record() {
        let p = prepare(&request_json(|_| {}), Provider::Typesafe, &who()).unwrap();
        assert_eq!(
            conclude(&p, &jev_answer("pytest")).unwrap_err().code,
            "unknown_choice"
        );
    }
}
