//! `moai-cli goal-block`: full-output snapshots and failure paths (A4 #2163, #2515, #2901, #2708).
//!
//! Every case feeds a fixture to the built binary on stdin and reads the envelope: on success the
//! `result` object, on failure the exit code and the error text. Every run uses a temporary HOME.
//! Moved here from the Python module `goal_block_test.py`
//! (SPEC-20261120-python-binary-tests-to-rust); the snapshot texts are the ones that module held.

mod common;
use common::pty::{env_with, piped_with};
use serde_json::{json, Value};

const LOOP: &str =
    "/goal 저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(milestone rc.2, HOLD·needs:decision 제외)";

fn issues() -> Value {
    json!({"milestone": "rc.2",
        "issues": [{"number": 204, "state": "open"}, {"number": 202, "state": "open"}],
        "prs": [], "keep_set": []})
}

fn prs() -> Value {
    json!({"milestone": "rc.2", "issues": [{"number": 210, "state": "closed"}],
        "prs": [{"number": 2200, "state": "open"}, {"number": 2199, "state": "open"},
                {"number": 2198, "state": "merged"}],
        "keep_set": []})
}

fn kept() -> Value {
    json!({"milestone": "rc.2",
        "issues": [{"number": 2016, "state": "open"}, {"number": 202, "state": "open"}],
        "prs": [{"number": 2199, "state": "open"}, {"number": 2017, "state": "open"}],
        "keep_set": [{"number": 2017, "reason": "운영자 보유"},
                     {"number": 2016, "reason": "운영자 보유"}]})
}

/// A failed run: its exit code and what it wrote (the envelope and the error line).
struct Failed {
    code: Option<i32>,
    text: String,
}

/// `moai-cli goal-block` on `fixture`: the envelope's `result` on success.
fn invoke(fixture: &Value) -> Result<Value, Failed> {
    let home = tempfile::tempdir().unwrap();
    let env = env_with(
        &[
            ("PATH", "/usr/bin:/bin"),
            ("HOME", home.path().to_str().unwrap()),
            ("LANG", "en_US.UTF-8"),
        ],
        &[],
    );
    let seen = piped_with(
        common::MOAI_CLI_BIN,
        home.path(),
        &env,
        &["goal-block"],
        Some(&fixture.to_string()),
    );
    if seen.code == Some(0) {
        let envelope: Value = serde_json::from_str(&seen.stdout)
            .unwrap_or_else(|e| panic!("exit 0 without JSON ({e}): {:?}", seen.stdout));
        assert_eq!(
            envelope["ok"], true,
            "goal-block exit 0 without ok: {:?}",
            seen.stdout
        );
        return Ok(envelope["result"].clone());
    }
    Err(Failed {
        code: seen.code,
        text: format!("{}{}", seen.stdout, seen.stderr),
    })
}

fn run_generator(fixture: &Value) -> Value {
    match invoke(fixture) {
        Ok(v) => v,
        Err(f) => panic!("generator failed: {}", f.text),
    }
}

#[test]
fn snapshot_issues_only() {
    assert_eq!(
        run_generator(&issues()),
        json!({
            "leader": {
                "goal_line": "/goal milestone rc.2의 열린 Issue가 모두 병합되거나 종료됨",
                "condition": "milestone rc.2의 열린 Issue가 모두 병합되거나 종료됨"},
            "agent": {"loop_line": LOOP, "goal_lines": [
                "/goal #202의 PR이 병합되거나 blocked로 보고됨",
                "/goal #204의 PR이 병합되거나 blocked로 보고됨"]},
            "counts": {"open_issues": 2, "open_prs": 0, "keep_set": 0},
            "keep_set_excluded": [], "model_can_input": false, "writes": false,
            "startable": [202, 204], "held": [],
            "end_condition": {"open_prs": 0, "startable": 2, "met": false},
        })
    );
}

#[test]
fn snapshot_prs_only() {
    assert_eq!(
        run_generator(&prs()),
        json!({
            "leader": {
                "goal_line": "/goal milestone rc.2의 열린 PR이 모두 병합되거나 종료됨",
                "condition": "milestone rc.2의 열린 PR이 모두 병합되거나 종료됨"},
            "agent": {"loop_line": LOOP, "goal_lines": null},
            "counts": {"open_issues": 0, "open_prs": 2, "keep_set": 0},
            "keep_set_excluded": [], "model_can_input": false, "writes": false,
            "startable": [], "held": [],
            "end_condition": {"open_prs": 2, "startable": 0, "met": false},
        })
    );
}

#[test]
fn snapshot_keep_set_included() {
    let goal = "/goal milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨 — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)";
    let condition = "milestone rc.2의 열린 Issue·PR이 모두 병합되거나 종료됨 — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)";
    assert_eq!(
        run_generator(&kept()),
        json!({
            "leader": {"goal_line": goal, "condition": condition},
            "agent": {"loop_line": format!("{LOOP} — 다음은 제외: #2016(운영자 보유), #2017(운영자 보유)"),
                      "goal_lines": ["/goal #202의 PR이 병합되거나 blocked로 보고됨"]},
            "counts": {"open_issues": 1, "open_prs": 1, "keep_set": 2},
            "keep_set_excluded": [{"number": 2016, "reason": "운영자 보유"},
                                  {"number": 2017, "reason": "운영자 보유"}],
            "model_can_input": false, "writes": false,
            "startable": [202],
            "held": [{"issue": 2016, "reason": "keep_set", "keep_reason": "운영자 보유"}],
            "end_condition": {"open_prs": 1, "startable": 1, "met": false},
        })
    );
}

#[test]
fn no_open_items_does_not_offer_an_already_satisfied_goal() {
    for fixture in [
        json!({"milestone": "rc.2"}),
        json!({"milestone": "rc.2", "issues": [{"number": 1, "state": "open"}],
               "prs": [{"number": 2, "state": "open"}],
               "keep_set": [{"number": 1, "reason": "보유"}, {"number": 2, "reason": "보유"}]}),
    ] {
        let out = run_generator(&fixture);
        assert!(out["leader"]["goal_line"].is_null(), "{fixture}");
        assert!(out.get("codex").is_none(), "{fixture}");
        assert!(out["agent"]["goal_lines"].is_null(), "{fixture}");
        assert!(out["agent"]["loop_line"].is_null(), "{fixture}");
        assert!(
            out["leader"]["condition"]
                .as_str()
                .unwrap()
                .contains("열린 항목 없음"),
            "{fixture}"
        );
        assert_eq!(out["counts"]["open_issues"], 0, "{fixture}");
        assert_eq!(out["counts"]["open_prs"], 0, "{fixture}");
    }
}

// #2515: a card-scoped goal ends the session when that card ends; the loop line holds until no
// open PR and no startable Issue remain, whatever card is held.
#[test]
fn the_agent_default_is_the_lane_wide_round_not_one_card() {
    let out = run_generator(&prs());
    assert!(out["agent"]["goal_lines"].is_null());
    assert_eq!(out["agent"]["loop_line"], LOOP);
    let head = out["agent"]["loop_line"]
        .as_str()
        .unwrap()
        .split('—')
        .next()
        .unwrap();
    assert!(!head.contains('#'), "{head}");
}

// #2901 AC5: open Issues that are all on hold or waiting for a decision, and no open PR: the
// loop's end condition is met, so no loop line is offered.
#[test]
fn held_cards_alone_meet_the_end_condition() {
    let out = run_generator(&json!({"milestone": "rc.2", "issues": [
        {"number": 2345, "state": "open", "labels": ["needs:decision"]},
        {"number": 2346, "state": "open", "labels": ["HOLD", "P1"]}]}));
    assert_eq!(
        out["end_condition"],
        json!({"open_prs": 0, "startable": 0, "met": true})
    );
    assert_eq!(out["startable"], json!([]));
    assert_eq!(
        out["held"],
        json!([
            {"issue": 2345, "reason": "held", "labels": ["needs:decision"]},
            {"issue": 2346, "reason": "held", "labels": ["HOLD"]}])
    );
    assert!(out["agent"]["loop_line"].is_null());
}

// #2901 AC5: one open PR, or one startable Issue next to held ones: not met.
#[test]
fn an_open_pr_or_a_startable_card_keeps_the_loop() {
    let held = json!({"number": 2345, "state": "open", "labels": ["needs:decision"]});
    for (fixture, want) in [
        (
            json!({"milestone": "rc.2", "issues": [held.clone()],
                   "prs": [{"number": 9, "state": "open"}]}),
            json!({"open_prs": 1, "startable": 0, "met": false}),
        ),
        (
            json!({"milestone": "rc.2", "issues": [held.clone(),
                   {"number": 7, "state": "open", "labels": ["threshold"]}]}),
            json!({"open_prs": 0, "startable": 1, "met": false}),
        ),
    ] {
        let out = run_generator(&fixture);
        assert_eq!(out["end_condition"], want, "{fixture}");
        assert_eq!(out["agent"]["loop_line"], LOOP, "{fixture}");
    }
}

// #2901: title (epic), open sub-issues, a body Blocked-By line naming an open Issue, a blocker
// GitHub reports; a Blocked-By line naming a closed Issue blocks nothing.
#[test]
fn each_issue_field_is_judged() {
    let out = run_generator(&json!({"milestone": "rc.2", "issues": [
        {"number": 1, "state": "open", "title": "[rc.2][EP-A] epic"},
        {"number": 2, "state": "open", "open_children": 2},
        {"number": 3, "state": "open", "body": "Blocked-By: #5"},
        {"number": 4, "state": "open", "blocked_by": [5]},
        {"number": 5, "state": "open", "body": "Blocked-By: #6"},
        {"number": 6, "state": "closed"}]}));
    assert_eq!(
        out["held"],
        json!([
            {"issue": 1, "reason": "epic", "open_children": 0},
            {"issue": 2, "reason": "epic", "open_children": 2},
            {"issue": 3, "reason": "blocked_by", "blocked_by": [5]},
            {"issue": 4, "reason": "blocked_by", "blocked_by": [5]}])
    );
    assert_eq!(out["startable"], json!([5]));
    for bad in [
        json!({"labels": "HOLD"}),
        json!({"labels": [1]}),
        json!({"title": 3}),
        json!({"body": null}),
        json!({"open_children": -1}),
        json!({"blocked_by": [0]}),
        json!({"blocked_by": "5"}),
    ] {
        let mut issue = json!({"number": 1, "state": "open"});
        for (k, v) in bad.as_object().unwrap() {
            issue[k] = v.clone();
        }
        let failed = invoke(&json!({"issues": [issue]}))
            .err()
            .unwrap_or_else(|| {
                panic!("{bad} was accepted");
            });
        assert_eq!(failed.code, Some(1), "{bad}");
        assert!(
            failed.text.contains("invalid fixture"),
            "{bad}: {}",
            failed.text
        );
    }
}

#[test]
fn input_permutations_produce_identical_output() {
    for fixture in [issues(), prs(), kept()] {
        let mut reversed = fixture.clone();
        for key in ["issues", "prs", "keep_set"] {
            reversed[key].as_array_mut().unwrap().reverse();
        }
        assert_eq!(
            run_generator(&fixture),
            run_generator(&reversed),
            "{fixture}"
        );
    }
}

#[test]
fn bad_items_are_rejected_without_traceback() {
    let mut fixtures = vec![
        json!([]),
        json!({"milestone": 1}),
        json!({"issues": null}),
        json!({"issues": [null]}),
        json!({"prs": [{}]}),
        json!({"issues": [{"state": "open"}]}),
        json!({"keep_set": [{"number": 1}]}),
        json!({"keep_set": [{"number": 1, "reason": 3}]}),
    ];
    for number in [json!(true), json!(0), json!(-1), json!("2016"), json!(1.5)] {
        fixtures.push(json!({"issues": [{"number": number, "state": "open"}]}));
    }
    for fixture in fixtures {
        let failed = invoke(&fixture)
            .err()
            .unwrap_or_else(|| panic!("{fixture} was accepted"));
        assert_eq!(failed.code, Some(1), "{fixture}");
        assert!(
            failed.text.contains("invalid fixture"),
            "{fixture}: {}",
            failed.text
        );
        assert!(
            !failed.text.contains("Traceback"),
            "{fixture}: {}",
            failed.text
        );
    }
}
