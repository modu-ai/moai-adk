//! Typed judgment backend route and outcome policy (Issue #2205, rc.2 F7; design §20 F20,
//! research `docs/judge-backend-research-2026-11-09.md`, user decision 01:1x(11-09)).
//!
//! REQ-F7-01: native judgment stays the default and a backend result never earns
//! authority — `authority` is exactly `none` on every surface here, and an outcome that
//! carries an authority field does not parse (`deny_unknown_fields`). An external backend
//! (`decisions`, `jev`) routes only when it is enabled AND carries a cost ceiling; the
//! ceiling unset or zero keeps external calls at 0 (decision 01:1x: "비용 상한 미설정 =
//! 외부 backend 호출 0", inheriting the Go `workflow.jev.enabled: false` default). One
//! judgment uses one backend path — never both externals. No HTTP client ships in this
//! card: `transport_wired` stays false and both operations are pure stdin/stdout, so the
//! live-call work keeps the seam (`endpoint`, `model_alias` are caller configuration,
//! never code constants). Policy values (pricing, credit eligibility) are not hardened as
//! constants; provenance is the research document plus its verification date (research §5).

use crate::contract;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// The research document and its verification date, recorded — not policy values in code.
const POLICY_RESEARCH: &str = "docs/judge-backend-research-2026-11-09.md";
const POLICY_VERIFIED_AT: &str = "2026-11-09";

const QUESTION_LIMIT: usize = 64;

#[derive(Serialize, Deserialize, Debug, PartialEq, Clone, Copy)]
#[serde(rename_all = "snake_case")]
pub enum Backend {
    /// The coding agent running the workflow. The default and the only call-free path.
    Native,
    /// OpenAI Decisions (`POST /v1/decisions`); opt-in.
    Decisions,
    /// TypeSafe Jev System One (`POST /v1/systemone`); opt-in, inheriting the Go
    /// `workflow.jev.enabled: false` default.
    Jev,
}

#[derive(Serialize, Deserialize, Debug, PartialEq, Clone, Copy)]
#[serde(rename_all = "snake_case")]
pub enum QuestionKind {
    /// Probability that a proposition is true. Never read as a Decisions predicate
    /// without an explicit conversion (design §20).
    Noul,
    /// One of a fixed candidate set.
    Choice,
    /// A bounded numeric answer.
    Score,
}

#[derive(Serialize, Deserialize, Debug, PartialEq, Clone)]
#[serde(deny_unknown_fields)]
pub struct JudgmentQuestion {
    pub id: String,
    pub kind: QuestionKind,
    #[serde(default)]
    pub min: Option<f64>,
    #[serde(default)]
    pub max: Option<f64>,
}

#[derive(Serialize, Deserialize, Debug, PartialEq, Clone)]
#[serde(deny_unknown_fields)]
pub struct BackendConfig {
    #[serde(default)]
    pub enabled: bool,
    /// Cost ceiling in micro-USD (millionths); absent or zero keeps external calls at 0.
    #[serde(default)]
    pub limit_usd_micros: Option<u64>,
    /// Caller configuration, never a code constant; nothing here drives a network call.
    #[serde(default)]
    pub endpoint: Option<String>,
    /// The model alias the request names (e.g. `jev-latest`); kept apart from the
    /// response's own model id, like the Go baseline (`internal/jev/jev.go`).
    #[serde(default)]
    pub model_alias: Option<String>,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct RouteRequest {
    pub protocol_version: u32,
    #[serde(default)]
    pub questions: Vec<JudgmentQuestion>,
    /// Omitted means native (REQ-F7-01): the mechanical default.
    #[serde(default)]
    pub selected: Option<Backend>,
    #[serde(default)]
    pub decisions: Option<BackendConfig>,
    #[serde(default)]
    pub jev: Option<BackendConfig>,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct JudgmentOutcome {
    pub backend: Backend,
    pub kind: QuestionKind,
    #[serde(default)]
    pub probability: Option<f64>,
    #[serde(default)]
    pub chosen: Option<String>,
    #[serde(default)]
    pub score: Option<f64>,
    /// The model id the response itself names — attribution, separate from the request
    /// alias; a mismatch is a change notice, never an error.
    #[serde(default)]
    pub model_id: Option<String>,
    #[serde(default)]
    pub rationale_ref: Option<String>,
    // No authority field: an outcome claiming one fails to parse (REQ-F7-01).
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct EvaluationRequest {
    pub protocol_version: u32,
    pub question: JudgmentQuestion,
    /// The alias the outgoing request named, for the alias/model-id check.
    #[serde(default)]
    pub model_alias: Option<String>,
    pub outcome: JudgmentOutcome,
}

fn plain(s: &str, max: usize) -> bool {
    !s.trim().is_empty() && s.len() <= max && !s.chars().any(char::is_control)
}

fn name_of(backend: Backend) -> &'static str {
    match backend {
        Backend::Native => "native",
        Backend::Decisions => "decisions",
        Backend::Jev => "jev",
    }
}

fn kind_name(kind: QuestionKind) -> &'static str {
    match kind {
        QuestionKind::Noul => "noul",
        QuestionKind::Choice => "choice",
        QuestionKind::Score => "score",
    }
}

/// Structural checks serde cannot express. The message names the failed rule, never a value.
fn validate_questions(questions: &[JudgmentQuestion]) -> Result<()> {
    if questions.is_empty() || questions.len() > QUESTION_LIMIT {
        return Err("judgment questions must be 1 to 64 typed questions".into());
    }
    for q in questions {
        if !plain(&q.id, 128) {
            return Err("question id must be a single-line string".into());
        }
        match q.kind {
            QuestionKind::Score => {
                let (Some(min), Some(max)) = (q.min, q.max) else {
                    return Err("score bounds are required on a score question".into());
                };
                if !min.is_finite() || !max.is_finite() || min > max {
                    return Err("score bounds must be finite with min <= max".into());
                }
            }
            _ => {
                if q.min.is_some() || q.max.is_some() {
                    return Err("score bounds are forbidden on a non-score question".into());
                }
            }
        }
    }
    Ok(())
}

/// Resolves the one backend path for this judgment and its external-call gate. Pure: no
/// filesystem, no network, no process spawn — the project path names the caller's context
/// and is not read.
pub fn route(request: &RouteRequest) -> Result<Value> {
    if request.protocol_version != contract::PROTOCOL_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    validate_questions(&request.questions)?;
    let selected = request.selected.unwrap_or(Backend::Native);
    let config_of = |backend| match backend {
        Backend::Decisions => request.decisions.as_ref(),
        Backend::Jev => request.jev.as_ref(),
        Backend::Native => None,
    };
    if selected != Backend::Native && config_of(selected).is_none() {
        return Err("judgment config missing for the selected backend".into());
    }
    let (gate_open, reason, external_allowed) = match selected {
        Backend::Native => (false, "native judgment needs no external call", false),
        other => {
            let cfg = config_of(other).expect("config presence checked above");
            if !cfg.enabled {
                (
                    false,
                    "backend disabled; the default keeps external calls at 0",
                    false,
                )
            } else {
                match cfg.limit_usd_micros {
                    None => (false, "cost ceiling unset; external calls stay 0", false),
                    Some(0) => (false, "cost ceiling is zero; external calls stay 0", false),
                    Some(_) => (true, "backend enabled with a cost ceiling", true),
                }
            }
        }
    };
    // One judgment, one path: other enabled backends are named, never selected.
    let mut unselected_enabled: Vec<&str> = Vec::new();
    for (backend, config) in [
        (Backend::Decisions, request.decisions.as_ref()),
        (Backend::Jev, request.jev.as_ref()),
    ] {
        if backend != selected && config.is_some_and(|c| c.enabled) {
            unselected_enabled.push(name_of(backend));
        }
    }
    Ok(json!({
        "selected_backend": name_of(selected),
        "external_calls_allowed": external_allowed,
        "gate": {"open": gate_open, "reason": reason},
        "transport_wired": false,
        "unselected_enabled": unselected_enabled,
        "policy_provenance": {"research": POLICY_RESEARCH, "verified_at": POLICY_VERIFIED_AT},
        "authority": "none",
        "network": "disabled",
    }))
}

/// Evaluates one backend outcome against the response policy. Pure; the outcome is a
/// fixture or, later, a certified transport's response — nothing here reaches the network.
pub fn evaluate(request: &EvaluationRequest) -> Result<Value> {
    if request.protocol_version != contract::PROTOCOL_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    validate_questions(std::slice::from_ref(&request.question))?;
    let outcome = &request.outcome;
    if outcome.kind != request.question.kind {
        return Err("outcome kind must be the question's kind (no implicit conversion)".into());
    }
    match outcome.kind {
        QuestionKind::Noul => {
            let Some(p) = outcome.probability else {
                return Err("noul probability must be present and within 0 to 1".into());
            };
            if !p.is_finite() || !(0.0..=1.0).contains(&p) {
                return Err("noul probability must be present and within 0 to 1".into());
            }
            if outcome.chosen.is_some() || outcome.score.is_some() {
                return Err("chosen is forbidden on a noul outcome".into());
            }
        }
        QuestionKind::Choice => {
            let Some(chosen) = outcome.chosen.as_deref() else {
                return Err("chosen must be present for a choice outcome".into());
            };
            if !plain(chosen, 512) {
                return Err("chosen must be a single-line string".into());
            }
            if outcome.probability.is_some() || outcome.score.is_some() {
                return Err("probability and score are forbidden on a choice outcome".into());
            }
        }
        QuestionKind::Score => {
            let Some(s) = outcome.score else {
                return Err("score must be present and within the question bounds".into());
            };
            let min = request.question.min.expect("score bounds validated above");
            let max = request.question.max.expect("score bounds validated above");
            if !s.is_finite() || s < min || s > max {
                return Err("score must be present and within the question bounds".into());
            }
            if outcome.probability.is_some() || outcome.chosen.is_some() {
                return Err("probability and chosen are forbidden on a score outcome".into());
            }
        }
    }
    if outcome
        .model_id
        .as_deref()
        .is_some_and(|id| !plain(id, 128))
    {
        return Err("model id must be a plain string".into());
    }
    if outcome
        .rationale_ref
        .as_deref()
        .is_some_and(|r| !plain(r, 512))
    {
        return Err("rationale ref must be a single-line string".into());
    }
    let model_match = match (request.model_alias.as_deref(), outcome.model_id.as_deref()) {
        (Some(alias), Some(model_id)) => Some(alias == model_id),
        _ => None,
    };
    Ok(json!({
        "question_id": request.question.id,
        "backend": name_of(outcome.backend),
        "kind": kind_name(outcome.kind),
        "model_match": model_match,
        "change_notice": model_match == Some(false),
        "authority": "none",
        "network": "disabled",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn question(kind: QuestionKind) -> JudgmentQuestion {
        JudgmentQuestion {
            id: "q1".into(),
            kind,
            min: None,
            max: None,
        }
    }

    fn route_request(
        selected: Option<Backend>,
        decisions: Option<BackendConfig>,
        jev: Option<BackendConfig>,
    ) -> RouteRequest {
        RouteRequest {
            protocol_version: 1,
            questions: vec![question(QuestionKind::Noul)],
            selected,
            decisions,
            jev,
        }
    }

    fn config(enabled: bool, limit: Option<u64>) -> BackendConfig {
        BackendConfig {
            enabled,
            limit_usd_micros: limit,
            endpoint: None,
            model_alias: None,
        }
    }

    #[test]
    fn route_defaults_to_native() {
        let r = route(&route_request(None, None, None)).expect("route");
        assert_eq!(r["selected_backend"], json!("native"));
        assert_eq!(r["external_calls_allowed"], json!(false));
        assert_eq!(r["authority"], json!("none"));
    }

    #[test]
    fn gate_matrix_keeps_calls_at_zero_without_ceiling() {
        let open = route(&route_request(
            Some(Backend::Decisions),
            Some(config(true, Some(5_000))),
            None,
        ))
        .expect("route");
        assert_eq!(open["gate"]["open"], json!(true));

        for limit in [None, Some(0)] {
            let r = route(&route_request(
                Some(Backend::Decisions),
                Some(config(true, limit)),
                None,
            ))
            .expect("route");
            assert_eq!(r["gate"]["open"], json!(false), "{limit:?}");
            assert_eq!(r["external_calls_allowed"], json!(false));
        }

        let disabled = route(&route_request(
            Some(Backend::Jev),
            None,
            Some(config(false, Some(1))),
        ))
        .expect("route");
        assert_eq!(disabled["gate"]["open"], json!(false));
    }

    #[test]
    fn route_names_other_enabled_backend_once() {
        let r = route(&route_request(
            Some(Backend::Decisions),
            Some(config(true, Some(1))),
            Some(config(true, Some(1))),
        ))
        .expect("route");
        assert_eq!(r["unselected_enabled"], json!(["jev"]));
    }

    #[test]
    fn route_refuses_selection_without_config() {
        let err = route(&route_request(Some(Backend::Decisions), None, None)).unwrap_err();
        assert!(err.contains("missing"), "{err}");
        assert_eq!(contract::classify(&err), contract::Kind::InvalidInput);
    }

    #[test]
    fn evaluate_enforces_kind_match_and_noul_range() {
        let ok = evaluate(&EvaluationRequest {
            protocol_version: 1,
            question: question(QuestionKind::Noul),
            model_alias: Some("jev-latest".into()),
            outcome: JudgmentOutcome {
                backend: Backend::Jev,
                kind: QuestionKind::Noul,
                probability: Some(0.82),
                chosen: None,
                score: None,
                model_id: Some("jev-1.14.0".into()),
                rationale_ref: None,
            },
        })
        .expect("evaluate");
        assert_eq!(ok["model_match"], json!(false));
        assert_eq!(ok["change_notice"], json!(true));
        assert_eq!(ok["authority"], json!("none"));

        let mismatch = evaluate(&EvaluationRequest {
            protocol_version: 1,
            question: question(QuestionKind::Noul),
            model_alias: None,
            outcome: JudgmentOutcome {
                backend: Backend::Decisions,
                kind: QuestionKind::Choice,
                probability: None,
                chosen: Some("a".into()),
                score: None,
                model_id: None,
                rationale_ref: None,
            },
        })
        .unwrap_err();
        assert!(mismatch.contains("must be"), "{mismatch}");

        let out_of_range = evaluate(&EvaluationRequest {
            protocol_version: 1,
            question: question(QuestionKind::Noul),
            model_alias: None,
            outcome: JudgmentOutcome {
                backend: Backend::Jev,
                kind: QuestionKind::Noul,
                probability: Some(1.5),
                chosen: None,
                score: None,
                model_id: None,
                rationale_ref: None,
            },
        })
        .unwrap_err();
        assert_eq!(
            contract::classify(&out_of_range),
            contract::Kind::InvalidInput
        );
    }

    #[test]
    fn evaluate_score_requires_bounds_within_range() {
        let mut q = question(QuestionKind::Score);
        q.min = Some(0.0);
        q.max = Some(1.0);
        let ok = evaluate(&EvaluationRequest {
            protocol_version: 1,
            question: q.clone(),
            model_alias: None,
            outcome: JudgmentOutcome {
                backend: Backend::Native,
                kind: QuestionKind::Score,
                probability: None,
                chosen: None,
                score: Some(0.5),
                model_id: None,
                rationale_ref: None,
            },
        })
        .expect("evaluate");
        assert_eq!(ok["authority"], json!("none"));

        let missing_bounds = evaluate(&EvaluationRequest {
            protocol_version: 1,
            question: question(QuestionKind::Score),
            model_alias: None,
            outcome: JudgmentOutcome {
                backend: Backend::Native,
                kind: QuestionKind::Score,
                probability: None,
                chosen: None,
                score: Some(0.5),
                model_id: None,
                rationale_ref: None,
            },
        })
        .unwrap_err();
        assert!(missing_bounds.contains("required"), "{missing_bounds}");
    }

    #[test]
    fn every_new_message_keeps_its_deliberate_kind() {
        let cases = [
            (
                "judgment questions must be 1 to 64 typed questions",
                contract::Kind::InvalidInput,
            ),
            (
                "question id must be a single-line string",
                contract::Kind::InvalidInput,
            ),
            (
                "score bounds are required on a score question",
                contract::Kind::InvalidInput,
            ),
            (
                "score bounds are forbidden on a non-score question",
                contract::Kind::InvalidInput,
            ),
            (
                "score bounds must be finite with min <= max",
                contract::Kind::InvalidInput,
            ),
            (
                "outcome kind must be the question's kind (no implicit conversion)",
                contract::Kind::InvalidInput,
            ),
            (
                "noul probability must be present and within 0 to 1",
                contract::Kind::InvalidInput,
            ),
            (
                "chosen is forbidden on a noul outcome",
                contract::Kind::InvalidInput,
            ),
            (
                "chosen must be present for a choice outcome",
                contract::Kind::InvalidInput,
            ),
            (
                "chosen must be a single-line string",
                contract::Kind::InvalidInput,
            ),
            (
                "probability and score are forbidden on a choice outcome",
                contract::Kind::InvalidInput,
            ),
            (
                "score must be present and within the question bounds",
                contract::Kind::InvalidInput,
            ),
            (
                "probability and chosen are forbidden on a score outcome",
                contract::Kind::InvalidInput,
            ),
            (
                "model id must be a plain string",
                contract::Kind::InvalidInput,
            ),
            (
                "rationale ref must be a single-line string",
                contract::Kind::InvalidInput,
            ),
            (
                "judgment config missing for the selected backend",
                contract::Kind::InvalidInput,
            ),
        ];
        for (message, kind) in cases {
            assert_eq!(contract::classify(message), kind, "{message}");
        }
    }
}
