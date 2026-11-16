//! The two decision providers of `moai decide` (card C-37 S2): the request bytes each one
//! takes and the answer each one gives. Pure: bytes in, bytes out, no transport.
//!
//! | name | label | request | answer |
//! | --- | --- | --- | --- |
//! | `typesafe` | Jev | `{model, state, questions}`; `questions` is an object keyed by the question id, each `{type, instructions, criteria}` with `criteria` an object `{option: description}` | `{model, answers: {<id>: {type, choice, confidence, probabilities}}}` with `probabilities` an object `{option: probability}` |
//! | `openai` | OpenAI | `{model, input, questions}`; `questions` is a list of `{type, name, instructions, choices: [{value, description}]}` | `{answers: [{type, name, choice, probabilities, confidence}]}` with `probabilities` a list of `{value, probability}` and no top-level `model`, or an answer of `type: "refusal"` |
//!
//! Sources: the plan of cards C-37 and C-54 (provider facts read 2026-11-16 from the vendors'
//! documentation and from the Go generation's Jev client). Nothing here was checked against
//! a live endpoint.
//!
//! The endpoints and the model aliases are compile-time constants. No setting, environment
//! variable or request field reaches them, so a request cannot be aimed at another host.
//! The model a record names is the one the response reports. The documented OpenAI response
//! reports none: there the alias that was sent stands in and the answer says so
//! (`model_source: "request"`, leader decision 2026-11-17 on audit finding B2); a top-level
//! `model` in an OpenAI response is used when it is there. Jev always reports its model, and
//! a Jev response without one is no answer.

use super::outbound::Outbound;
use super::Refusal;
use serde_json::{json, Map, Value};
use std::collections::BTreeMap;

pub const TYPESAFE_ENDPOINT: &str = "https://api.typesafe.ai/v1/systemone";
pub const TYPESAFE_MODEL: &str = "jev-latest";
pub const OPENAI_ENDPOINT: &str = "https://api.openai.com/v1/decisions";
pub const OPENAI_MODEL: &str = "gpt-6-luna";
/// The id of the one question a request asks, and the key its answer comes back under.
pub const QUESTION_ID: &str = "decision";
/// Longest model name accepted from a response, in bytes.
pub const MODEL_NAME_MAX: usize = 64;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Provider {
    Typesafe,
    Openai,
}

impl Provider {
    pub const ALL: [Provider; 2] = [Provider::Typesafe, Provider::Openai];

    /// The provider of a vault and settings name, exactly as written.
    pub fn from_name(name: &str) -> Option<Self> {
        Self::ALL.into_iter().find(|p| p.name() == name)
    }

    /// The name used for the vault entry, the settings and `decided_by`.
    pub fn name(self) -> &'static str {
        match self {
            Provider::Typesafe => "typesafe",
            Provider::Openai => "openai",
        }
    }

    /// The name a person reads.
    pub fn label(self) -> &'static str {
        match self {
            Provider::Typesafe => "Jev",
            Provider::Openai => "OpenAI",
        }
    }

    pub fn endpoint(self) -> &'static str {
        match self {
            Provider::Typesafe => TYPESAFE_ENDPOINT,
            Provider::Openai => OPENAI_ENDPOINT,
        }
    }

    /// The model alias a request asks for.
    pub fn model(self) -> &'static str {
        match self {
            Provider::Typesafe => TYPESAFE_MODEL,
            Provider::Openai => OPENAI_MODEL,
        }
    }
}

/// What a provider answered: one of the options, and the model that chose it.
#[derive(Debug, Clone, PartialEq)]
pub struct Answer {
    pub choice: String,
    pub confidence: Option<f64>,
    pub probabilities: BTreeMap<String, f64>,
    /// As the response reported it, or the alias the request asked for when the response
    /// names none (OpenAI); `model_source` says which.
    pub model: String,
    /// `response` or `request`.
    pub model_source: &'static str,
}

/// The request body for `provider`, exactly as it would be sent: compact JSON, keys in the
/// order the provider documents them, options in the request's order.
pub fn request_body(provider: Provider, outbound: &Outbound) -> Vec<u8> {
    // serde_json keeps the order keys are written in (`preserve_order`, Cargo.toml); the
    // golden bodies in the tests pin it.
    let body = match provider {
        Provider::Typesafe => {
            let criteria: Map<String, Value> = outbound
                .options()
                .iter()
                .map(|(value, description)| (value.clone(), json!(description)))
                .collect();
            json!({
                "model": TYPESAFE_MODEL,
                "state": outbound.evidence(),
                "questions": {
                    QUESTION_ID: {
                        "type": "choice",
                        "instructions": outbound.what(),
                        "criteria": criteria,
                    }
                },
            })
        }
        Provider::Openai => {
            let choices: Vec<Value> = outbound
                .options()
                .iter()
                .map(|(value, description)| json!({"value": value, "description": description}))
                .collect();
            json!({
                "model": OPENAI_MODEL,
                "input": outbound.evidence(),
                "questions": [{
                    "type": "choice",
                    "name": QUESTION_ID,
                    "instructions": outbound.what(),
                    "choices": choices,
                }],
            })
        }
    };
    serde_json::to_vec(&body).expect("a JSON value serializes")
}

fn no_answer(rule: &str) -> Refusal {
    Refusal::new(
        "answer_invalid",
        format!("the provider's response is not an answer: {rule}"),
    )
}

/// A number from 0 to 1, as a confidence or a probability is.
fn unit(value: &Value) -> Option<f64> {
    value.as_f64().filter(|n| (0.0..=1.0).contains(n))
}

/// The answer in a provider's `response` body. A refusal by the provider, a choice that is not
/// one of the options and a body that is not the documented shape all send the choice back to
/// the leader; none of them quotes the response.
pub fn parse_answer(
    provider: Provider,
    outbound: &Outbound,
    response: &[u8],
) -> Result<Answer, Refusal> {
    let body: Value = serde_json::from_slice(response).map_err(|_| no_answer("not JSON"))?;
    let answer = match provider {
        Provider::Typesafe => body["answers"].as_object().and_then(|a| a.get(QUESTION_ID)),
        Provider::Openai => body["answers"]
            .as_array()
            .and_then(|a| a.iter().find(|one| one["name"] == QUESTION_ID)),
    }
    .ok_or_else(|| no_answer("no answer to the question that was asked"))?;
    match answer["type"].as_str() {
        Some("choice") => {}
        Some("refusal") => {
            return Err(Refusal::new(
                "provider_refused",
                format!(
                    "{} declined to answer: the leader decides",
                    provider.label()
                ),
            ))
        }
        _ => return Err(no_answer("the answer is not of type `choice`")),
    }
    let plain = |m: &str| {
        (1..=MODEL_NAME_MAX).contains(&m.len())
            && m.bytes()
                .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'-' | b':'))
    };
    let (model, model_source) = match (&body["model"], provider) {
        (Value::String(m), _) if plain(m) => (m.as_str(), "response"),
        // The documented OpenAI response names no model: the alias that was sent stands in.
        (Value::Null, Provider::Openai) => (OPENAI_MODEL, "request"),
        _ => return Err(no_answer("no plain `model` name of 1 to 64 bytes")),
    };
    let is_option = |value: &str| outbound.options().iter().any(|(v, _)| v == value);
    let choice = answer["choice"]
        .as_str()
        .ok_or_else(|| no_answer("`choice` is not a string"))?;
    if !is_option(choice) {
        return Err(Refusal::new(
            "unknown_choice",
            "the answer names none of the options: the leader decides",
        ));
    }
    let confidence = match &answer["confidence"] {
        Value::Null => None,
        value => Some(unit(value).ok_or_else(|| no_answer("`confidence` is not 0 to 1"))?),
    };
    let mut probabilities = BTreeMap::new();
    match (provider, &answer["probabilities"]) {
        (_, Value::Null) => {}
        (Provider::Typesafe, Value::Object(map)) => {
            for (option, value) in map {
                let p = unit(value)
                    .filter(|_| is_option(option))
                    .ok_or_else(|| no_answer("`probabilities` is not {option: 0 to 1}"))?;
                probabilities.insert(option.clone(), p);
            }
        }
        (Provider::Openai, Value::Array(list)) => {
            for item in list {
                let (option, p) = item["value"]
                    .as_str()
                    .filter(|option| is_option(option))
                    .zip(unit(&item["probability"]))
                    .ok_or_else(|| {
                        no_answer("`probabilities` is not [{value: option, probability: 0 to 1}]")
                    })?;
                probabilities.insert(option.to_string(), p);
            }
        }
        _ => return Err(no_answer("`probabilities` is not in the provider's shape")),
    }
    Ok(Answer {
        choice: choice.to_string(),
        confidence,
        probabilities,
        model: model.to_string(),
        model_source,
    })
}

#[cfg(test)]
mod tests {
    use super::super::testing::{request_json, who};
    use super::super::{prepare, Prepared};
    use super::*;

    fn prepared(provider: Provider) -> Prepared {
        prepare(&request_json(|_| {}), provider, &who()).unwrap()
    }

    fn answer(provider: Provider, response: &Value) -> Result<Answer, Refusal> {
        let p = prepared(provider);
        parse_answer(provider, &p.outbound, response.to_string().as_bytes())
    }

    fn code(provider: Provider, response: &Value) -> &'static str {
        answer(provider, response).unwrap_err().code
    }

    /// The endpoints and aliases, literally: a change here is a change of where text goes.
    #[test]
    fn the_endpoints_and_models_are_the_compiled_ones() {
        let t = Provider::Typesafe;
        let o = Provider::Openai;
        assert_eq!(
            (t.name(), t.label(), t.endpoint(), t.model()),
            (
                "typesafe",
                "Jev",
                "https://api.typesafe.ai/v1/systemone",
                "jev-latest"
            )
        );
        assert_eq!(
            (o.name(), o.label(), o.endpoint(), o.model()),
            (
                "openai",
                "OpenAI",
                "https://api.openai.com/v1/decisions",
                "gpt-6-luna"
            )
        );
        assert_eq!(Provider::from_name("typesafe"), Some(t));
        assert_eq!(Provider::from_name("openai"), Some(o));
        for other in ["", "jev", "OpenAI", "typesafe ", "glm"] {
            assert_eq!(Provider::from_name(other), None, "{other:?}");
        }
        assert_eq!(prepared(t).endpoint(), TYPESAFE_ENDPOINT);
        assert_eq!(prepared(o).endpoint(), OPENAI_ENDPOINT);
    }

    /// Golden body, Jev: `questions` is an object keyed by the question id and `criteria` an
    /// object keyed by option (a list is answered with HTTP 422).
    #[test]
    fn the_typesafe_body_is_the_documented_bytes() {
        let want = concat!(
            r#"{"model":"jev-latest","#,
            r#""state":"CI runs nextest; the local docs name both runners.","#,
            r#""questions":{"decision":{"type":"choice","#,
            r#""instructions":"Which test runner does the new module use?","#,
            r#""criteria":{"nextest":"cargo nextest, as CI runs it","#,
            r#""cargo-test":"plain cargo test"}}}}"#
        );
        assert_eq!(
            String::from_utf8(prepared(Provider::Typesafe).body().to_vec()).unwrap(),
            want
        );
    }

    /// Golden body, OpenAI: `questions` is a list, the question is named, and the options are
    /// a list of `{value, description}`.
    #[test]
    fn the_openai_body_is_the_documented_bytes() {
        let want = concat!(
            r#"{"model":"gpt-6-luna","#,
            r#""input":"CI runs nextest; the local docs name both runners.","#,
            r#""questions":[{"type":"choice","name":"decision","#,
            r#""instructions":"Which test runner does the new module use?","#,
            r#""choices":[{"value":"nextest","description":"cargo nextest, as CI runs it"},"#,
            r#"{"value":"cargo-test","description":"plain cargo test"}]}]}"#
        );
        assert_eq!(
            String::from_utf8(prepared(Provider::Openai).body().to_vec()).unwrap(),
            want
        );
    }

    /// The body holds the sent fields only: not the card, the default, the deadline or the
    /// category, and the masked text, not the original.
    #[test]
    fn the_body_carries_only_the_masked_sent_fields() {
        for provider in Provider::ALL {
            let text = request_json(|v| {
                v["evidence"] = json!("/Users/testperson/proj/ci.yml runs nextest");
                v["card"] = json!("C-9999");
                v["deadline"] = json!("2031-01-02T03:04:05Z");
            });
            let prepared = prepare(&text, provider, &who()).unwrap();
            let body = String::from_utf8(prepared.body().to_vec()).unwrap();
            assert!(body.contains("~/proj/ci.yml runs nextest"), "{body}");
            for absent in [
                "testperson",
                "C-9999",
                "2031-01-02",
                "reversible",
                "category",
                "default",
            ] {
                assert!(!body.contains(absent), "{absent}: {body}");
            }
            // Text is JSON-escaped, so it cannot add a field.
            let quoted = request_json(|v| v["what"] = json!("a \"quoted\"\nline, \"model\":\"x\""));
            let prepared = prepare(&quoted, provider, &who()).unwrap();
            let v: Value = serde_json::from_slice(prepared.body()).unwrap();
            assert_eq!(v["model"], provider.model());
            assert_eq!(v.as_object().unwrap().len(), 3);
        }
    }

    fn jev(answer: Value) -> Value {
        json!({"model": "jev-1.14.0", "answers": {"decision": answer}, "usage": {"input_tokens": 61}})
    }

    /// The documented OpenAI response: a list of answers and no top-level `model`.
    fn openai(answer: Value) -> Value {
        json!({"answers": [answer]})
    }

    #[test]
    fn a_documented_typesafe_answer_is_read() {
        let a = answer(
            Provider::Typesafe,
            &jev(json!({
                "type": "choice", "choice": "cargo-test", "confidence": 0.75,
                "probabilities": {"nextest": 0.25, "cargo-test": 0.75}, "legend": {}
            })),
        )
        .unwrap();
        assert_eq!(a.choice, "cargo-test");
        assert_eq!(a.confidence, Some(0.75));
        assert_eq!(
            a.probabilities,
            BTreeMap::from([
                ("cargo-test".to_string(), 0.75),
                ("nextest".to_string(), 0.25)
            ])
        );
        assert_eq!(
            (a.model.as_str(), a.model_source),
            ("jev-1.14.0", "response")
        );
    }

    #[test]
    fn a_documented_openai_answer_is_read() {
        let a = answer(
            Provider::Openai,
            &openai(json!({
                "type": "choice", "name": "decision", "choice": "nextest",
                "probabilities": [
                    {"value": "nextest", "probability": 0.95},
                    {"value": "cargo-test", "probability": 0.05}
                ],
                "confidence": 0.93
            })),
        )
        .unwrap();
        assert_eq!(a.choice, "nextest");
        assert_eq!(a.confidence, Some(0.93));
        assert_eq!(
            a.probabilities,
            BTreeMap::from([
                ("cargo-test".to_string(), 0.05),
                ("nextest".to_string(), 0.95)
            ])
        );
        // The documented response names no model: the alias that was asked for stands in,
        // and the answer says so.
        assert_eq!(
            (a.model.as_str(), a.model_source),
            ("gpt-6-luna", "request")
        );
        // Confidence and probabilities may be absent; the choice may not.
        let bare = answer(
            Provider::Openai,
            &openai(json!({"type": "choice", "name": "decision", "choice": "nextest"})),
        )
        .unwrap();
        assert_eq!((bare.confidence, bare.probabilities.len()), (None, 0));
    }

    #[test]
    fn a_provider_refusal_falls_back() {
        let refusal = json!({"type": "refusal", "name": "decision", "refusal": "cannot decide"});
        let r = answer(Provider::Openai, &openai(refusal.clone())).unwrap_err();
        assert_eq!(r.code, "provider_refused");
        assert!(!r.reason.contains("cannot decide"), "{}", r.reason);
        assert_eq!(code(Provider::Typesafe, &jev(refusal)), "provider_refused");
    }

    #[test]
    fn a_choice_that_is_not_an_option_falls_back() {
        for choice in ["pytest", "Nextest", "nextest ", ""] {
            let a = json!({"type": "choice", "name": "decision", "choice": choice});
            assert_eq!(
                code(Provider::Openai, &openai(a.clone())),
                "unknown_choice",
                "{choice:?}"
            );
            assert_eq!(
                code(Provider::Typesafe, &jev(a)),
                "unknown_choice",
                "{choice:?}"
            );
        }
    }

    /// Each provider's `probabilities` in its own documented shape only: a list of
    /// `{value, probability}` for OpenAI, an object `{option: probability}` for Jev.
    #[test]
    fn probabilities_have_each_providers_shape() {
        let good = json!({"type": "choice", "name": "decision", "choice": "nextest"});
        let with = |value: Value| {
            let mut a = good.clone();
            a["probabilities"] = value;
            a
        };
        let list = json!([{"value": "nextest", "probability": 1.0}]);
        let map = json!({"nextest": 1.0});
        assert!(answer(Provider::Openai, &openai(with(list.clone()))).is_ok());
        assert!(answer(Provider::Typesafe, &jev(with(map.clone()))).is_ok());
        assert_eq!(code(Provider::Openai, &openai(with(map))), "answer_invalid");
        assert_eq!(code(Provider::Typesafe, &jev(with(list))), "answer_invalid");
        for bad in [
            json!([{"value": "pytest", "probability": 0.5}]),
            json!([{"value": "nextest", "probability": 2}]),
            json!([{"value": "nextest", "probability": "high"}]),
            json!([{"value": "nextest"}]),
            json!([{"probability": 0.5}]),
            json!([0.5, 0.5]),
            json!("nextest"),
        ] {
            assert_eq!(
                code(Provider::Openai, &openai(with(bad.clone()))),
                "answer_invalid",
                "{bad}"
            );
        }
    }

    /// OpenAI: a top-level `model` is believed when it is a plain name, refused when it is
    /// there and is not one, and replaced by the request's alias only when it is absent.
    #[test]
    fn an_openai_model_name_is_the_responses_or_the_requests() {
        let good = json!({"type": "choice", "name": "decision", "choice": "nextest"});
        let with_model = |model: Value| {
            let mut v = openai(good.clone());
            v["model"] = model;
            v
        };
        let named = answer(
            Provider::Openai,
            &with_model(json!("gpt-6-luna-2026-10-01")),
        )
        .unwrap();
        assert_eq!(
            (named.model.as_str(), named.model_source),
            ("gpt-6-luna-2026-10-01", "response")
        );
        let absent = answer(Provider::Openai, &openai(good.clone())).unwrap();
        assert_eq!(
            (absent.model.as_str(), absent.model_source),
            ("gpt-6-luna", "request")
        );
        for bad in [
            json!(""),
            json!("two words"),
            json!(7),
            json!("m".repeat(65)),
        ] {
            assert_eq!(
                code(Provider::Openai, &with_model(bad.clone())),
                "answer_invalid",
                "{bad}"
            );
        }
    }

    /// Anything that is not the documented shape is no answer.
    #[test]
    fn a_response_of_another_shape_falls_back() {
        let good = json!({"type": "choice", "name": "decision", "choice": "nextest"});
        let with = |key: &str, value: Value| {
            let mut a = good.clone();
            a[key] = value;
            a
        };
        for bad in [
            with("type", json!("score")),
            with("choice", json!(1)),
            with("confidence", json!("high")),
            with("confidence", json!(1.5)),
            with("confidence", json!(-0.1)),
            with("probabilities", json!([0.5, 0.5])),
            with("probabilities", json!({"nextest": 2})),
            with("probabilities", json!({"pytest": 0.5})),
        ] {
            assert_eq!(
                code(Provider::Openai, &openai(bad.clone())),
                "answer_invalid",
                "{bad}"
            );
            assert_eq!(
                code(Provider::Typesafe, &jev(bad.clone())),
                "answer_invalid",
                "{bad}"
            );
        }
        // The answer is missing, or under another question's id or name.
        assert_eq!(
            code(Provider::Typesafe, &json!({"model": "jev-1.14.0"})),
            "answer_invalid"
        );
        assert_eq!(
            code(
                Provider::Typesafe,
                &json!({"model": "jev-1.14.0", "answers": {"other": good}})
            ),
            "answer_invalid"
        );
        assert_eq!(
            code(Provider::Openai, &openai(with("name", json!("other")))),
            "answer_invalid"
        );
        // One provider's shape is not the other's.
        assert_eq!(code(Provider::Openai, &jev(good.clone())), "answer_invalid");
        assert_eq!(
            code(Provider::Typesafe, &openai(good.clone())),
            "answer_invalid"
        );
        // Not JSON at all, and an error body.
        let p = prepared(Provider::Typesafe);
        for body in [
            &b"<html>502</html>"[..],
            b"",
            b"{\"detail\":[{\"msg\":\"bad\"}]}",
        ] {
            let r = parse_answer(Provider::Typesafe, &p.outbound, body).unwrap_err();
            assert_eq!(r.code, "answer_invalid");
            assert!(!r.reason.contains("502") && !r.reason.contains("bad"));
        }
    }

    /// `decided_by` needs the model the response names: without a plain one there is no
    /// answer. The name length at its edge: 64 bytes read, 65 refused.
    #[test]
    fn the_model_comes_from_the_response_or_there_is_no_answer() {
        let good = json!({"type": "choice", "name": "decision", "choice": "nextest"});
        let with_model = |model: Value| {
            let mut v = jev(good.clone());
            v["model"] = model;
            v
        };
        assert_eq!(
            answer(Provider::Typesafe, &with_model(json!("m".repeat(64))))
                .unwrap()
                .model
                .len(),
            64
        );
        for bad in [
            json!("m".repeat(65)),
            json!(""),
            json!("two words"),
            json!("line\nbreak"),
            json!(7),
            Value::Null,
        ] {
            assert_eq!(
                code(Provider::Typesafe, &with_model(bad.clone())),
                "answer_invalid",
                "{bad}"
            );
        }
        let mut missing = jev(good);
        missing.as_object_mut().unwrap().remove("model");
        assert_eq!(code(Provider::Typesafe, &missing), "answer_invalid");
    }
}
