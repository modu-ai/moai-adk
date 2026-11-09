//! Issue #2205 (rc.2 F7): typed judgment backend route and outcome policy.
//! Every case is pure stdin/stdout: no network, no filesystem writes, fixture outcomes only.
use serde_json::{json, Value};
use std::{
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// Runs an internal operation; returns the exit code and the parsed envelope.
fn run(args: &[&str], stdin: &Value) -> (Option<i32>, Value) {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out.status.code(), v)
}

/// `judge-route PROJECT` carries the project path as its positional argument.
fn run_route(dir: &Path, stdin: &Value) -> (Option<i32>, Value) {
    let args = vec!["judge-route", dir.to_str().expect("temp path")];
    run(&args, stdin)
}

fn route_request(selected: Value, decisions: Value, jev: Value) -> Value {
    let mut v = json!({
        "protocol_version": 1,
        "questions": [{"id": "q1", "kind": "noul"}],
    });
    if !selected.is_null() {
        v["selected"] = selected;
    }
    if !decisions.is_null() {
        v["decisions"] = decisions;
    }
    if !jev.is_null() {
        v["jev"] = jev;
    }
    v
}

fn eval_request(question: Value, alias: Value, outcome: Value) -> Value {
    let mut v = json!({
        "protocol_version": 1,
        "question": question,
        "outcome": outcome,
    });
    if !alias.is_null() {
        v["model_alias"] = alias;
    }
    v
}

fn kind_of(v: &Value) -> &str {
    v["error"]["kind"].as_str().expect("error kind")
}

#[test]
fn route_defaults_to_native_with_no_external_call() {
    let dir = TempDir::new().unwrap();
    let (code, v) = run_route(
        dir.path(),
        &route_request(Value::Null, Value::Null, Value::Null),
    );
    assert_eq!(code, Some(0), "exit 0, got {v}");
    assert_eq!(v["ok"], json!(true));
    let r = &v["result"];
    assert_eq!(r["selected_backend"], json!("native"));
    assert_eq!(r["external_calls_allowed"], json!(false));
    assert_eq!(r["gate"]["open"], json!(false));
    assert_eq!(r["transport_wired"], json!(false));
    assert_eq!(r["unselected_enabled"], json!([]));
    assert_eq!(r["authority"], json!("none"));
    assert_eq!(r["network"], json!("disabled"));
    assert_eq!(
        r["policy_provenance"]["research"],
        json!("docs/judge-backend-research-2026-11-09.md")
    );
    assert_eq!(r["policy_provenance"]["verified_at"], json!("2026-11-09"));
}

#[test]
fn route_enabled_external_with_ceiling_opens_the_gate() {
    let dir = TempDir::new().unwrap();
    let req = route_request(
        json!("decisions"),
        json!({"enabled": true, "limit_usd_micros": 5000}),
        Value::Null,
    );
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["selected_backend"], json!("decisions"));
    assert_eq!(r["external_calls_allowed"], json!(true));
    assert_eq!(r["gate"]["open"], json!(true));
    // Routing only: no certified transport ships in this card, so no call is possible.
    assert_eq!(r["transport_wired"], json!(false));
    assert_eq!(r["network"], json!("disabled"));
    assert_eq!(r["authority"], json!("none"));
}

#[test]
fn route_external_without_ceiling_keeps_calls_at_zero() {
    let dir = TempDir::new().unwrap();
    let req = route_request(json!("decisions"), json!({"enabled": true}), Value::Null);
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["external_calls_allowed"], json!(false));
    assert_eq!(r["gate"]["open"], json!(false));
    assert!(
        r["gate"]["reason"]
            .as_str()
            .unwrap()
            .contains("cost ceiling unset"),
        "reason names the unset ceiling: {}",
        r["gate"]["reason"]
    );
}

#[test]
fn route_disabled_backend_keeps_calls_at_zero() {
    let dir = TempDir::new().unwrap();
    let req = route_request(json!("jev"), Value::Null, json!({"enabled": false}));
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["selected_backend"], json!("jev"));
    assert_eq!(r["external_calls_allowed"], json!(false));
    assert!(
        r["gate"]["reason"]
            .as_str()
            .unwrap()
            .contains("backend disabled"),
        "reason names the disabled backend: {}",
        r["gate"]["reason"]
    );
}

#[test]
fn route_zero_ceiling_keeps_calls_at_zero() {
    let dir = TempDir::new().unwrap();
    let req = route_request(
        json!("jev"),
        Value::Null,
        json!({"enabled": true, "limit_usd_micros": 0}),
    );
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["external_calls_allowed"], json!(false));
    assert!(
        r["gate"]["reason"]
            .as_str()
            .unwrap()
            .contains("cost ceiling is zero"),
        "reason names the zero ceiling: {}",
        r["gate"]["reason"]
    );
}

#[test]
fn route_selects_exactly_one_path_and_names_other_enabled() {
    let dir = TempDir::new().unwrap();
    let req = route_request(
        json!("decisions"),
        json!({"enabled": true, "limit_usd_micros": 5000}),
        json!({"enabled": true, "limit_usd_micros": 5000}),
    );
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["selected_backend"], json!("decisions"));
    assert_eq!(r["unselected_enabled"], json!(["jev"]));
    assert_eq!(r["authority"], json!("none"));
}

#[test]
fn route_refuses_an_unknown_backend() {
    let dir = TempDir::new().unwrap();
    let req = route_request(json!("openai"), Value::Null, Value::Null);
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(v["ok"], json!(false));
    assert_eq!(kind_of(&v), "invalid_input");
}

#[test]
fn route_refuses_selection_without_its_config() {
    let dir = TempDir::new().unwrap();
    let req = route_request(json!("decisions"), Value::Null, Value::Null);
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
}

#[test]
fn route_rejects_unsupported_protocol_version() {
    let dir = TempDir::new().unwrap();
    let mut req = route_request(Value::Null, Value::Null, Value::Null);
    req["protocol_version"] = json!(2);
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "unsupported_protocol_version");
}

#[test]
fn route_rejects_empty_questions() {
    let dir = TempDir::new().unwrap();
    let mut req = route_request(Value::Null, Value::Null, Value::Null);
    req["questions"] = json!([]);
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
}

#[test]
fn route_rejects_too_many_questions() {
    let dir = TempDir::new().unwrap();
    let mut req = route_request(Value::Null, Value::Null, Value::Null);
    req["questions"] = json!((0..65)
        .map(|i| json!({"id": format!("q{i}"), "kind": "noul"}))
        .collect::<Vec<_>>());
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
}

#[test]
fn route_accepts_exactly_the_question_limit() {
    let dir = TempDir::new().unwrap();
    let req = route_request(Value::Null, Value::Null, Value::Null);
    let mut req = req;
    req["questions"] = json!((0..64)
        .map(|i| json!({"id": format!("q{i}"), "kind": "noul"}))
        .collect::<Vec<_>>());
    let (code, v) = run_route(dir.path(), &req);
    assert_eq!(code, Some(0), "64 questions stay inside the limit, got {v}");
    assert_eq!(v["ok"], json!(true));
}

#[test]
fn evaluate_reports_authority_none_and_alias_model_separation() {
    let req = eval_request(
        json!({"id": "q1", "kind": "noul"}),
        json!("jev-latest"),
        json!({
            "backend": "jev",
            "kind": "noul",
            "probability": 0.82,
            "model_id": "jev-1.14.0"
        }),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    assert_eq!(v["ok"], json!(true));
    let r = &v["result"];
    assert_eq!(r["authority"], json!("none"));
    assert_eq!(r["network"], json!("disabled"));
    // Request alias `jev-latest` vs the response's actual model id: a notice, never an error.
    assert_eq!(r["model_match"], json!(false));
    assert_eq!(r["change_notice"], json!(true));
}

#[test]
fn evaluate_matching_model_is_no_notice() {
    let req = eval_request(
        json!({"id": "q1", "kind": "noul"}),
        json!("jev-1.14.0"),
        json!({
            "backend": "jev",
            "kind": "noul",
            "probability": 0.82,
            "model_id": "jev-1.14.0"
        }),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["model_match"], json!(true));
    assert_eq!(r["change_notice"], json!(false));
}

#[test]
fn evaluate_refuses_a_kind_mismatch_without_conversion() {
    let req = eval_request(
        json!({"id": "q1", "kind": "noul"}),
        Value::Null,
        json!({"backend": "decisions", "kind": "choice", "chosen": "a"}),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
    assert!(
        v["error"]["message"].as_str().unwrap().contains("must be"),
        "message names the kind rule: {}",
        v["error"]["message"]
    );
}

#[test]
fn evaluate_refuses_probability_out_of_range() {
    let req = eval_request(
        json!({"id": "q1", "kind": "noul"}),
        Value::Null,
        json!({"backend": "jev", "kind": "noul", "probability": 1.5}),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
}

#[test]
fn evaluate_accepts_probability_at_both_bounds() {
    for p in [0.0, 1.0] {
        let req = eval_request(
            json!({"id": "q1", "kind": "noul"}),
            Value::Null,
            json!({"backend": "jev", "kind": "noul", "probability": p}),
        );
        let (code, v) = run(&["judge-evaluate"], &req);
        assert_eq!(code, Some(0), "probability {p} is inside 0..=1, got {v}");
        assert_eq!(v["result"]["authority"], json!("none"));
    }
}

#[test]
fn evaluate_refuses_a_choice_field_on_noul() {
    let req = eval_request(
        json!({"id": "q1", "kind": "noul"}),
        Value::Null,
        json!({"backend": "jev", "kind": "noul", "probability": 0.5, "chosen": "a"}),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("forbidden"),
        "message names the forbidden field: {}",
        v["error"]["message"]
    );
}

#[test]
fn evaluate_choice_outcome_is_accepted() {
    let req = eval_request(
        json!({"id": "q2", "kind": "choice"}),
        Value::Null,
        json!({"backend": "decisions", "kind": "choice", "chosen": "b"}),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    let r = &v["result"];
    assert_eq!(r["authority"], json!("none"));
    assert_eq!(r["kind"], json!("choice"));
}

#[test]
fn evaluate_score_checks_the_question_bounds() {
    let in_bounds = eval_request(
        json!({"id": "q3", "kind": "score", "min": 0, "max": 1}),
        Value::Null,
        json!({"backend": "native", "kind": "score", "score": 0.5}),
    );
    let (code, v) = run(&["judge-evaluate"], &in_bounds);
    assert_eq!(code, Some(0), "exit 0, got {v}");
    assert_eq!(v["result"]["authority"], json!("none"));

    let out_of_bounds = eval_request(
        json!({"id": "q3", "kind": "score", "min": 0, "max": 1}),
        Value::Null,
        json!({"backend": "native", "kind": "score", "score": 2.0}),
    );
    let (code, v) = run(&["judge-evaluate"], &out_of_bounds);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");

    // The bound itself is inside: a score of exactly max passes.
    let at_bound = eval_request(
        json!({"id": "q3", "kind": "score", "min": 0, "max": 1}),
        Value::Null,
        json!({"backend": "native", "kind": "score", "score": 1}),
    );
    let (code, v) = run(&["judge-evaluate"], &at_bound);
    assert_eq!(
        code,
        Some(0),
        "score == max stays inside the bounds, got {v}"
    );
    assert_eq!(v["result"]["authority"], json!("none"));
}

#[test]
fn evaluate_refuses_a_score_question_without_bounds() {
    let req = eval_request(
        json!({"id": "q3", "kind": "score"}),
        Value::Null,
        json!({"backend": "native", "kind": "score", "score": 0.5}),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
}

#[test]
fn evaluate_refuses_an_outcome_claiming_authority() {
    let req = eval_request(
        json!({"id": "q1", "kind": "noul"}),
        Value::Null,
        json!({
            "backend": "jev",
            "kind": "noul",
            "probability": 0.9,
            "authority": "merge"
        }),
    );
    let (code, v) = run(&["judge-evaluate"], &req);
    assert_eq!(code, Some(1), "exit 1, got {v}");
    assert_eq!(kind_of(&v), "invalid_input");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("invalid typed JSON input"),
        "unknown fields stay refused: {}",
        v["error"]["message"]
    );
}
