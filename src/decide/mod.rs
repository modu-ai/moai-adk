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
//! - [`request`] reads the request (design v2, slice V1): a short question, 2 to 8 options with
//!   required descriptions, the required `undo` line and `tier`, and evidence as 1 to 8
//!   named fields with byte bounds. Every bound refuses; nothing is cut.
//! - [`gate`] refuses a choice that is not the model's to make: not reversible, no category,
//!   an unknown one, `guarded`, one of the five HOLD categories, a default that is not an
//!   option, or a text holding one shape of the narrow deny list (known threats only, not a
//!   risk judge: a force push, `reset --hard`, `rm -rf`, dropping a table, making a
//!   repository public, `chmod 777`, a currency amount).
//! - [`outbound`] is the text that would leave the machine. A credential shape in it refuses
//!   the request (user decision 2026-11-16, D-secret-hit: refuse and fall back, never mask and
//!   send); the home folder and the user name are masked.
//! - [`provider`] builds the exact request bytes for `typesafe` (Jev) and `openai`, and reads
//!   their answers. The endpoints are compile-time constants; no setting, variable or request
//!   field can aim a request elsewhere.
//!
//! Every [`Refusal`] means the same thing to the caller: the model does not decide, the
//! leader does (`fallback: leader`). A [`Record`] carries the fields every decision record
//! names (SKILL.md), how the choice is undone, its tier and the names (never the values) of
//! the evidence fields, and who decided, `model:<provider>/<model>`, with `model_source` saying
//! whether the response reported that model or, when it named none, the request's alias
//! stands in; the model gives no written reason and none is recorded (leader default
//! D-reason).

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
    endpoint: &'static str,
    body: Vec<u8>,
    outbound: outbound::Outbound,
    reversible: bool,
    category: String,
    default: String,
    deadline: String,
    tier: u8,
}

impl Prepared {
    /// Where the bytes would go: one of the two compiled endpoints. Read-only, so nothing
    /// outside this module can aim a prepared request elsewhere.
    pub fn endpoint(&self) -> &'static str {
        self.endpoint
    }

    /// The request body, exactly as it would be sent. Read-only: the bytes that passed the
    /// gate and the secret screen are the bytes that go.
    pub fn body(&self) -> &[u8] {
        &self.body
    }

    /// The masked text the body was built from. Read-only: an answer is checked against
    /// the options that were sent, not against ones put here afterwards.
    pub fn outbound(&self) -> &outbound::Outbound {
        &self.outbound
    }
}

/// Largest request body, in bytes (256 KiB). A hard bound for now; the providers' token
/// limits come with the settings (S3).
pub const BODY_MAX: usize = 256 * 1024;

/// One model decision, as recorded. The fields are the ones every decision record names
/// (what, options, default, reversible, category, deadline, who decided), how the choice is
/// undone, its tier, the names of the evidence fields, plus the answer. The texts are the
/// masked ones that were sent (`undo` is masked and never sent). There is no reason field,
/// and no evidence value.
#[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct Record {
    /// 2 since design v2 (slice V1 added `undo`, `tier` and `evidence_fields`).
    pub schema_version: u32,
    pub what: String,
    pub options: Vec<String>,
    pub default: String,
    pub reversible: bool,
    pub category: String,
    pub deadline: String,
    pub undo: String,
    pub tier: u8,
    pub evidence_fields: Vec<String>,
    /// `model:<provider>/<model>`; the model is the one the response reported, and the alias
    /// the request asked for only when the response names none (`model_source`).
    pub decided_by: String,
    /// Where the model name in `decided_by` comes from: `response` when the provider's
    /// response named the model, `request` when it named none and the alias that was asked
    /// for stands in (leader decision 2026-11-17, audit B2).
    pub model_source: String,
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
    if body.len() > BODY_MAX {
        return Err(outbound::too_large());
    }
    Ok(Prepared {
        provider,
        endpoint: provider.endpoint(),
        body,
        outbound,
        reversible: request.reversible == Some(true),
        category: request.category.clone().unwrap_or_default(),
        default: request.default.clone(),
        deadline: request.deadline.clone(),
        // The parser admits 1 or 2 only.
        tier: request.tier.map_or(0, |tier| tier as u8),
    })
}

/// The record of the provider's `response` to `prepared`, or the refusal that sends the
/// choice back to the leader (a provider refusal, an answer that is not one of the options,
/// a response that cannot be read).
pub fn conclude(prepared: &Prepared, response: &[u8]) -> Result<Record, Refusal> {
    let answer = provider::parse_answer(prepared.provider, &prepared.outbound, response)?;
    Ok(Record {
        schema_version: 2,
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
        undo: prepared.outbound.undo().to_string(),
        tier: prepared.tier,
        evidence_fields: prepared.outbound.evidence_fields().to_vec(),
        decided_by: format!("model:{}/{}", prepared.provider.name(), answer.model),
        model_source: answer.model_source.to_string(),
        chosen: answer.choice,
        confidence: answer.confidence,
        probabilities: answer.probabilities,
        card: prepared.outbound.card().map(str::to_string),
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
            "undo": "set the runner line in ci.yml back to the other runner",
            "tier": 1,
            "evidence": {"ci": "CI runs nextest; the local docs name both runners."},
            "card": "C-37"
        });
        edit(&mut v);
        v.to_string()
    }

    /// Evidence of one field, `log`, holding `text`: for the tests that are about one text.
    pub fn evidence(text: impl serde::Serialize) -> Value {
        json!({ "log": text })
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
        assert_eq!(p.body(), b"body");
        assert_eq!(p.endpoint(), "https://api.openai.com/v1/decisions");
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

    /// Each deny-list shape refuses with its own code before the provider's builder runs, in
    /// whichever text it stands.
    #[test]
    fn each_deny_shape_is_refused_before_the_provider() {
        for (text, code) in [
            ("then git push --force", "deny_force_push"),
            ("git reset --hard HEAD~1", "deny_reset_hard"),
            ("rm -rf target", "deny_rm_rf"),
            ("DROP TABLE users", "deny_drop_table"),
            ("gh repo edit --visibility public", "deny_visibility_public"),
            ("chmod 777 bin", "deny_chmod_777"),
            ("about $20 a month", "deny_currency"),
        ] {
            let r = refused(&request_json(|v| v["what"] = json!(text)));
            assert_eq!(r.code, code, "{text}");
            let r = refused(&request_json(|v| v["undo"] = json!(text)));
            assert_eq!(r.code, code, "undo: {text}");
            let r = refused(&request_json(|v| v["evidence"]["log"] = json!(text)));
            assert_eq!(r.code, code, "evidence: {text}");
        }
        // The near misses reach the builder.
        for text in ["git push origin x", "echo $HOME", "rm -r tmp"] {
            assert!(
                prepare(
                    &request_json(|v| v["what"] = json!(text)),
                    Provider::Typesafe,
                    &who()
                )
                .is_ok(),
                "{text}"
            );
        }
    }

    /// The parser's refusals come before the provider's builder too, each with its own code.
    #[test]
    fn the_v2_request_rules_refuse_before_the_provider() {
        let nine: Value = (0..9)
            .map(|i| json!({"value": format!("o{i}"), "description": "d"}))
            .collect();
        for (code, edit) in [
            (
                "undo_missing",
                &(|v: &mut Value| {
                    v.as_object_mut().unwrap().remove("undo");
                }) as &dyn Fn(&mut Value),
            ),
            ("tier_missing", &|v: &mut Value| {
                v.as_object_mut().unwrap().remove("tier");
            }),
            ("tier_invalid", &|v: &mut Value| v["tier"] = json!(3)),
            ("options_count", &|v: &mut Value| {
                v["options"] = nine.clone();
                v["default"] = json!("o0");
            }),
            ("description_missing", &|v: &mut Value| {
                v["options"][1]["description"] = json!("")
            }),
            ("option_reserved", &|v: &mut Value| {
                v["options"][1]["value"] = json!("none_of_these")
            }),
            ("evidence_field_too_large", &|v: &mut Value| {
                v["evidence"] = json!({"a": "e".repeat(4_097)})
            }),
            ("evidence_count", &|v: &mut Value| {
                v["evidence"] = (0..9)
                    .map(|i| (format!("f{i}"), json!("x")))
                    .collect::<serde_json::Map<_, _>>()
                    .into()
            }),
        ] {
            assert_eq!(refused(&request_json(edit)).code, code);
        }
    }

    #[test]
    fn a_secret_shape_is_refused_before_the_provider_and_is_not_quoted() {
        // Test data: built at run time, shaped like a key and not one.
        let fake = format!("sk-{}", "a1".repeat(12));
        let r = refused(&request_json(|v| {
            v["evidence"] = testing::evidence(format!("the log shows {fake} in the header"))
        }));
        assert_eq!(r.code, "secret_shape");
        assert!(!r.reason.contains(&fake));
        assert!(!r.to_json().to_string().contains(&fake));
    }

    #[test]
    fn a_request_that_does_not_parse_is_refused_before_the_provider() {
        assert_eq!(refused("{").code, "invalid_request");
        // A request cannot name where it goes (no URL is written here: the scan of this
        // folder in tests/home_isolation_scan.rs allows the two endpoint constants only).
        let extra = refused(&request_json(|v| v["endpoint"] = json!("example.invalid")));
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

    /// A Jev answer that weighs its own choice highest, as a consistent answer does.
    fn jev_answer(choice: &str) -> Vec<u8> {
        let other = if choice == "nextest" {
            "cargo-test"
        } else {
            "nextest"
        };
        json!({
            "model": "jev-1.14.0",
            "answers": {"decision": {
                "type": "choice", "choice": choice, "confidence": 0.875,
                "probabilities": {choice: 0.875, other: 0.125}
            }},
            "usage": {"input_tokens": 61}
        })
        .to_string()
        .into_bytes()
    }

    /// The record names the fields of every decision record and who decided: the provider and
    /// the model the response reported, not the alias that was asked for. It names how the
    /// choice is undone, its tier and the evidence field names, never their values. No
    /// reason field.
    #[test]
    fn the_record_names_the_model_the_response_reported() {
        let text = request_json(|v| {
            v["tier"] = json!(2);
            v["undo"] = json!("revert the commit in /Users/testperson/proj");
            v["evidence"]["notes"] = json!("VALUE_ONLY_IN_THE_BODY");
        });
        let p = prepare(&text, Provider::Typesafe, &who()).unwrap();
        let record = conclude(&p, &jev_answer("cargo-test")).unwrap();
        assert_eq!(
            serde_json::to_value(&record).unwrap(),
            json!({
                "schema_version": 2,
                "what": "Which test runner does the new module use?",
                "options": ["nextest", "cargo-test"],
                "default": "nextest",
                "reversible": true,
                "category": "none",
                "deadline": "2026-11-18T00:00:00Z",
                "undo": "revert the commit in ~/proj",
                "tier": 2,
                "evidence_fields": ["ci", "notes"],
                "decided_by": "model:typesafe/jev-1.14.0",
                "model_source": "response",
                "chosen": "cargo-test",
                "confidence": 0.875,
                "probabilities": {"cargo-test": 0.875, "nextest": 0.125},
                "card": "C-37"
            })
        );
        assert!(String::from_utf8_lossy(p.body()).contains("VALUE_ONLY_IN_THE_BODY"));
        // The record reads back, and a field it does not know (a written reason) does not.
        let text = serde_json::to_string(&record).unwrap();
        assert!(!text.contains("VALUE_ONLY_IN_THE_BODY"), "{text}");
        assert_eq!(serde_json::from_str::<Record>(&text).unwrap(), record);
        let mut with_reason = serde_json::to_value(&record).unwrap();
        with_reason["reason"] = json!("because");
        assert!(serde_json::from_value::<Record>(with_reason).is_err());
    }

    /// Audit B2: the documented OpenAI response names no model. The record then names the
    /// alias the request asked for and says so (`model_source: request`); a response that
    /// does name one is believed (`response`).
    #[test]
    fn an_openai_record_says_where_its_model_name_comes_from() {
        let p = prepare(&request_json(|_| {}), Provider::Openai, &who()).unwrap();
        let documented = json!({"answers": [{
            "type": "choice", "name": "decision", "choice": "nextest",
            "probabilities": [
                {"value": "nextest", "probability": 0.95},
                {"value": "cargo-test", "probability": 0.05}
            ],
            "confidence": 0.93
        }]});
        let record = conclude(&p, documented.to_string().as_bytes()).unwrap();
        assert_eq!(record.decided_by, "model:openai/gpt-6-luna");
        assert_eq!(record.model_source, "request");
        assert_eq!(record.chosen, "nextest");
        assert_eq!(record.confidence, Some(0.93));
        assert_eq!(record.probabilities["cargo-test"], 0.05);
        let mut named = documented.clone();
        named["model"] = json!("gpt-6-luna-2026-10-01");
        let record = conclude(&p, named.to_string().as_bytes()).unwrap();
        assert_eq!(record.decided_by, "model:openai/gpt-6-luna-2026-10-01");
        assert_eq!(record.model_source, "response");
    }

    /// The card is recorded, never sent: masked like the sent fields, and a secret shape in
    /// it refuses the request before the provider (audit N5).
    #[test]
    fn the_card_is_masked_in_the_record_and_screened() {
        let text = request_json(|v| v["card"] = json!("testperson-C-37"));
        let p = prepare(&text, Provider::Typesafe, &who()).unwrap();
        assert!(!String::from_utf8_lossy(p.body()).contains("C-37"));
        let record = conclude(&p, &jev_answer("nextest")).unwrap();
        assert_eq!(record.card.as_deref(), Some("[user]-C-37"));
        let fake = format!("sk-{}", "a1".repeat(12));
        let r = refused(&request_json(|v| v["card"] = json!(fake.clone())));
        assert_eq!(r.code, "secret_shape");
    }

    /// The body size at its edge (audit N10): 262,144 bytes are sent, 262,145 are refused.
    /// The request bounds of design v2 keep a real body far below it, so the builder here
    /// stands in with a body of the size under test.
    #[test]
    fn a_body_over_256_kib_is_refused() {
        assert_eq!(BODY_MAX, 262_144);
        let with_body = |n: usize| {
            prepare_with(&request_json(|_| {}), Provider::Typesafe, &who(), |_, _| {
                vec![b'a'; n]
            })
        };
        assert_eq!(with_body(262_144).unwrap().body().len(), 262_144);
        assert_eq!(with_body(262_145).unwrap_err().code, "too_large");
    }

    /// The largest request the v2 bounds allow, with every text escaped to twice its size,
    /// stays under the body bound for both providers.
    #[test]
    fn the_largest_v2_request_fits_the_body_bound() {
        let quotes = |n: usize| "\"".repeat(n);
        let text = request_json(|v| {
            v["what"] = json!(quotes(500));
            v["options"] = (0..8)
                .map(|i| json!({"value": format!("o{i}{}", "v".repeat(62)), "description": quotes(300)}))
                .collect();
            v["default"] = json!(format!("o0{}", "v".repeat(62)));
            v["evidence"] = (0..4)
                .map(|i| (format!("f{i}"), json!(quotes(4_096))))
                .collect::<serde_json::Map<_, _>>()
                .into();
        });
        for provider in Provider::ALL {
            let p = prepare(&text, provider, &who()).unwrap();
            assert!(p.body().len() < BODY_MAX / 4, "{}", p.body().len());
        }
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
