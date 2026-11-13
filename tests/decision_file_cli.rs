//! Issue #2619 AC-2/AC-3: `decision-file-check` refuses a malformed decision bundle, and
//! `decision-due` shows an unanswered reversible decision taking its default once its
//! deadline passes while a one-way, `guarded` or HOLD-category one keeps waiting.

mod common;
use serde_json::Value;
use std::path::Path;
use std::process::{Command, Stdio};

const BUNDLE: &str = concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/tests/fixtures/decisions/bundle.md"
);

fn cli(args: &[&str]) -> (Value, i32) {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stdout)));
    (v, out.status.code().unwrap_or(-1))
}

fn ids(v: &Value, key: &str) -> Vec<String> {
    v["result"][key]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r["id"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn the_fixture_bundle_passes_the_check() {
    let (v, code) = cli(&["decision-file-check", BUNDLE]);
    assert_eq!((code, &v["ok"]), (0, &Value::Bool(true)), "{v}");
    assert_eq!(v["operation"], "decision.file_check");
    assert_eq!(v["result"]["rows"].as_array().unwrap().len(), 5);
}

#[test]
fn a_malformed_bundle_is_refused_with_its_line() {
    let tmp = tempfile::tempdir().unwrap();
    let bad = tmp.path().join("bad.md");
    let text = std::fs::read_to_string(BUNDLE)
        .unwrap()
        .replace("| D3 | publish the docs site | A: now; B: later | B | no | external_contract | 2026-11-14T00:00:00Z | user |", "| D3 | publish the docs site | A: now; B: later | B | no | external_contract | 2026-11-14T00:00:00Z | leader |");
    std::fs::write(&bad, text).unwrap();
    let (v, code) = cli(&["decision-file-check", bad.to_str().unwrap()]);
    assert_eq!((code, &v["ok"]), (1, &Value::Bool(false)), "{v}");
    let message = v["error"]["message"].as_str().unwrap();
    assert!(
        message.contains("line 15 D3: a one-way or HOLD-category decision has `decider` user"),
        "{message}"
    );
    let (due, code) = cli(&[
        "decision-due",
        bad.to_str().unwrap(),
        "--now",
        "2026-11-15T00:00:00Z",
    ]);
    assert_eq!((code, &due["ok"]), (1, &Value::Bool(false)), "{due}");
    let (missing, code) = cli(&[
        "decision-file-check",
        &Path::new(BUNDLE).with_extension("none").to_string_lossy(),
    ]);
    assert_eq!(
        (code, &missing["ok"]),
        (1, &Value::Bool(false)),
        "{missing}"
    );
}

#[test]
fn past_the_deadline_only_the_reversible_none_user_decision_takes_its_default() {
    // One second before D2..D4's deadline: everything unanswered is still open.
    let (before, code) = cli(&["decision-due", BUNDLE, "--now", "2026-11-13T23:59:59Z"]);
    assert_eq!(code, 0, "{before}");
    assert_eq!(ids(&before, "open"), ["D2", "D3", "D4", "D5"]);
    assert!(ids(&before, "default_applies").is_empty());
    assert_eq!(ids(&before, "answered"), ["D1"]);
    // At the deadline: D2 takes its default A; D3 (one-way) and D4 (guarded) stay on hold;
    // D5's later deadline keeps it open.
    let (at, code) = cli(&["decision-due", BUNDLE, "--now", "2026-11-14T00:00:00Z"]);
    assert_eq!(code, 0, "{at}");
    assert_eq!(ids(&at, "default_applies"), ["D2"]);
    assert_eq!(at["result"]["default_applies"][0]["default"], "A");
    assert_eq!(ids(&at, "hold"), ["D3", "D4"]);
    assert_eq!(ids(&at, "open"), ["D5"]);
    // A year on, the one-way and guarded decisions are still the user's.
    let (later, _) = cli(&["decision-due", BUNDLE, "--now", "2027-11-14T00:00:00Z"]);
    assert_eq!(ids(&later, "hold"), ["D3", "D4"]);
    assert_eq!(ids(&later, "default_applies"), ["D2", "D5"]);
    // A malformed clock is refused, not read as the epoch.
    let (bad_now, code) = cli(&["decision-due", BUNDLE, "--now", "2026-02-31T00:00:00Z"]);
    assert_eq!(
        (code, &bad_now["ok"]),
        (1, &Value::Bool(false)),
        "{bad_now}"
    );
}
