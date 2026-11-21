//! SPEC-20261121-codex-loop-ledger: `moai lane-stop-hook` — the Codex lane Stop hook's
//! decision, read from the project's card ledger only (no GitHub request), so installing the
//! hook cannot stop a lane on a missing token or a slow read. A lane with a held card is
//! blocked with its next step; a lane with nothing held but a `ready` card is blocked with
//! the take; a `hold` card is never offered; a lane with nothing to work on is allowed to
//! stop. Every command runs against a temporary HOME and a temporary project.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::path::PathBuf;
use std::process::Output;

const S: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const T: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

fn as_session(home: &TempHome, session: &str, args: &[&str]) -> Output {
    home.command(args)
        .env("CLAUDE_CODE_SESSION_ID", session)
        .output()
        .expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// `(exit code, envelope)` of `moai card …` (envelope commands).
fn card(home: &TempHome, session: &str, args: &[&str], _project: &str) -> Value {
    let out = as_session(home, session, args);
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// The Stop hook's raw answer: `(exit code, stdout)` — a block is the decision JSON on
/// stdout; an allow is an empty stdout (the hook prints nothing when the lane may stop).
fn hook(home: &TempHome, session: &str, project: &str) -> (i32, Value) {
    let out = as_session(home, session, &["lane-stop-hook", "--project", project]);
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    let v: Value = serde_json::from_str(stdout.trim()).unwrap_or_else(|e| {
        panic!(
            "stdout is not a decision ({e}): {stdout:?} {}",
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code().unwrap_or(-1), v)
}

/// `(exit code, stdout)` of the hook when the answer may be empty (an allow).
fn hook_raw(home: &TempHome, session: &str, project: &str) -> (i32, String) {
    let out = as_session(home, session, &["lane-stop-hook", "--project", project]);
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

fn sql(home: &TempHome, statement: &str) {
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    w.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn project_with_a_ready_card(home: &TempHome) -> (PathBuf, String) {
    let p = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    card(
        home,
        "leader",
        &[
            "card",
            "add",
            "--title",
            "The next card of the lane",
            "--priority",
            "p1",
            "--scope",
            "src/a.rs",
            "--project",
            p.to_str().unwrap(),
        ],
        p.to_str().unwrap(),
    );
    card(
        home,
        "leader",
        &["card", "ready", "C-1", "--project", p.to_str().unwrap()],
        p.to_str().unwrap(),
    );
    (p, "C-1".to_string())
}

#[test]
fn a_lane_with_a_held_card_is_blocked_with_its_next_step() {
    let home = TempHome::new();
    let (p, name) = project_with_a_ready_card(&home);
    card(
        &home,
        S,
        &["card", "take", &name, "--project", p.to_str().unwrap()],
        p.to_str().unwrap(),
    );
    let (code, v) = hook(&home, S, p.to_str().unwrap());
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["decision"], json!("block"), "{v}");
    let reason = v["reason"].as_str().unwrap();
    assert!(reason.contains("continue C-1 (active)"), "{reason}");
    assert!(reason.contains("the lane holds this card"), "{reason}");
}

#[test]
fn a_lane_with_a_ready_card_is_blocked_with_the_take() {
    let home = TempHome::new();
    let (p, name) = project_with_a_ready_card(&home);
    let (code, v) = hook(&home, T, p.to_str().unwrap());
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["decision"], json!("block"), "{v}");
    let reason = v["reason"].as_str().unwrap();
    assert!(
        reason.starts_with(&format!("take {name} (p1): ")),
        "{reason}"
    );
    assert!(reason.contains("The next card of the lane"), "{reason}");
}

#[test]
fn an_ended_lease_on_the_held_card_asks_for_a_reclaim() {
    let home = TempHome::new();
    let (p, name) = project_with_a_ready_card(&home);
    card(
        &home,
        S,
        &["card", "take", &name, "--project", p.to_str().unwrap()],
        p.to_str().unwrap(),
    );
    sql(
        &home,
        "UPDATE reservation SET lease_until = unixepoch() - 5 WHERE status = 'held'",
    );
    let (code, v) = hook(&home, S, p.to_str().unwrap());
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["decision"], json!("block"), "{v}");
    let reason = v["reason"].as_str().unwrap();
    assert!(reason.contains("re-claim C-1"), "{reason}");
    assert!(reason.contains("generation + 1"), "{reason}");
}

#[test]
fn a_hold_card_is_never_offered() {
    let home = TempHome::new();
    let (p, _name) = project_with_a_ready_card(&home);
    sql(
        &home,
        "UPDATE card SET state = 'hold' WHERE title = 'The next card of the lane'",
    );
    // The hold card is not a `ready` card, so the lane may stop: the hook prints nothing.
    let (code, stdout) = hook_raw(&home, T, p.to_str().unwrap());
    assert_eq!(code, 0, "{stdout}");
    assert_eq!(stdout, "", "an allow prints nothing: {stdout:?}");
}

#[test]
fn a_project_without_a_ledger_allows_the_stop() {
    let home = TempHome::new();
    let p = home.project("demo");
    let (code, stdout) = hook_raw(&home, T, p.to_str().unwrap());
    assert_eq!(code, 0, "{stdout}");
    assert_eq!(stdout, "", "an allow prints nothing: {stdout:?}");
}
