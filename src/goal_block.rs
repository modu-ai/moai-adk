//! Deterministic `/goal` paste block (A4 #2163, REQ-A4-01/02), moved from
//! `tools/gen_goal_block.py` (#2708) so the shipped mode workflows need only `moai-cli`.
//!
//! From the state a session read — `milestone`, `issues` and `prs` (each `{number, state}`)
//! and `keep_set` (each `{number, reason}`) — it builds the leader's milestone-wide condition,
//! the agent's lane-wide `loop_line` (no open PR and no startable Issue, #2515) with the
//! card-scoped `goal_lines`, and the Codex line with its caveat. Open entries are sorted by
//! number, so any input order gives the same block. The model cannot type `/goal` itself; the
//! user pastes the line (`model_can_input: false`). Reads and writes nothing else.
use serde_json::{json, Map, Value};

const CODEX_NOTE: &str = "Codex /goal의 진행 강제 여부는 미확인 — 효과를 보장하지 않는다(docs/host-support-matrix.md goal 행, 2026-11-09 확인)";

fn invalid(reason: &str) -> String {
    format!("invalid fixture: {reason}")
}

/// A positive JSON integer (a boolean or a float is refused, as in the Python tool).
fn number(item: &Map<String, Value>) -> Option<u64> {
    item.get("number")?.as_u64().filter(|n| *n > 0)
}

fn group<'a>(
    fixture: &'a Map<String, Value>,
    name: &str,
) -> Result<Vec<&'a Map<String, Value>>, String> {
    let Some(value) = fixture.get(name) else {
        return Ok(Vec::new());
    };
    let items = value
        .as_array()
        .ok_or_else(|| invalid(&format!("{name} must be a list")))?;
    items
        .iter()
        .map(|item| {
            let obj = item
                .as_object()
                .ok_or_else(|| invalid(&format!("{name} entries must be objects")))?;
            if number(obj).is_none() {
                return Err(invalid(&format!(
                    "{name} entries require a positive integer number"
                )));
            }
            if name == "keep_set" && !obj.get("reason").is_some_and(Value::is_string) {
                return Err(invalid("keep_set entries require a string reason"));
            }
            Ok(obj)
        })
        .collect()
}

/// The block for one fixture, or `invalid fixture: …`.
pub fn build(fixture: &Value) -> Result<Value, String> {
    let fixture = fixture
        .as_object()
        .ok_or_else(|| invalid("expected an object"))?;
    let milestone = match fixture.get("milestone") {
        None => "",
        Some(Value::String(s)) => s.as_str(),
        Some(_) => return Err(invalid("milestone must be a string")),
    };
    let issues_all = group(fixture, "issues")?;
    let prs_all = group(fixture, "prs")?;
    let mut keep_set = group(fixture, "keep_set")?;
    keep_set.sort_by_key(|k| number(k));
    let excluded: Vec<u64> = keep_set.iter().filter_map(|k| number(k)).collect();
    let open = |items: &[&Map<String, Value>]| -> Vec<u64> {
        let mut out: Vec<u64> = items
            .iter()
            .filter(|i| i.get("state").and_then(Value::as_str) == Some("open"))
            .filter_map(|i| number(i))
            .filter(|n| !excluded.contains(n))
            .collect();
        out.sort();
        out
    };
    let (issues, prs) = (open(&issues_all), open(&prs_all));

    let subject = match (issues.is_empty(), prs.is_empty()) {
        (false, true) => "Issue",
        (true, false) => "PR",
        _ => "Issue·PR",
    };
    let any_open = !issues.is_empty() || !prs.is_empty();
    let mut condition = if any_open {
        let particle = if subject == "Issue" { "가" } else { "이" };
        format!("milestone {milestone}의 열린 {subject}{particle} 모두 병합되거나 종료됨")
    } else {
        format!("milestone {milestone}의 열린 항목 없음")
    };
    let exclusions = keep_set
        .iter()
        .map(|k| {
            format!(
                "#{}({})",
                number(k).unwrap_or(0),
                k["reason"].as_str().unwrap_or("")
            )
        })
        .collect::<Vec<_>>()
        .join(", ");
    if !keep_set.is_empty() {
        condition.push_str(&format!(" — 다음은 제외: {exclusions}"));
    }
    let leader_line = any_open.then(|| format!("/goal {condition}"));
    let loop_line = any_open.then(|| {
        let mut line = format!(
            "/goal 저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(milestone {milestone}, HOLD·needs:decision 제외)"
        );
        if !keep_set.is_empty() {
            line.push_str(&format!(" — 다음은 제외: {exclusions}"));
        }
        line
    });
    let goal_lines: Option<Vec<String>> = (!issues.is_empty()).then(|| {
        issues
            .iter()
            .map(|n| format!("/goal #{n}의 PR이 병합되거나 blocked로 보고됨"))
            .collect()
    });
    Ok(json!({
        "leader": {"goal_line": leader_line, "condition": condition},
        "agent": {"loop_line": loop_line, "goal_lines": goal_lines},
        "codex": {"goal_line": leader_line, "note": CODEX_NOTE},
        "counts": {
            "open_issues": issues.len(),
            "open_prs": prs.len(),
            "keep_set": keep_set.len(),
        },
        "keep_set_excluded": keep_set
            .iter()
            .map(|k| json!({"number": number(k), "reason": k["reason"]}))
            .collect::<Vec<_>>(),
        "model_can_input": false,
        "writes": false,
    }))
}
