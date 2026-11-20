//! `card withdraw` (no schema change): a withdrawn card is refused by `card take` with its
//! withdrawal reason, and the default `card list` omits it. The withdrawn mark is the free
//! event kind `card.withdraw` beside the existing `hold` state, written through the ledger
//! function the CLI action calls (its wiring lands with this slice). Every case runs in a
//! temporary HOME; the real `~/.moai` is not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, WriteLedger};
use serde_json::{json, Value};
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, S_A)
        .output()
        .expect("run moai")
}

/// Asserts exit 0 and a success envelope, and unwraps the result.
fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// Asserts exit 1 and returns the whole refusal envelope.
fn refusal(out: &Output) -> Value {
    assert_eq!(out.status.code(), Some(1));
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false), "{v}");
    v
}

/// A registered project with two ready cards: C-1 withdrawn through the ledger function the
/// CLI action calls, C-2 still plain.
fn world() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    let p = project.to_str().unwrap().to_string();
    for title in ["withdraw me", "plain card"] {
        let added = result(&run(
            &home,
            &[
                "card",
                "add",
                "--title",
                title,
                "--scope",
                "apps/demo/",
                "--project",
                &p,
            ],
        ));
        let name = added["name"].as_str().unwrap().to_string();
        result(&run(&home, &["card", "ready", &name, "--project", &p]));
    }
    let mut w: WriteLedger = ledger::open_write(home.path(), "demo").unwrap();
    moai_cli::ledger::workflow::withdraw(
        &mut w,
        1,
        "superseded by another card",
        "leader",
        Some(S_A),
    )
    .unwrap();
    drop(w);
    (home, p)
}

#[test]
fn take_on_a_withdrawn_card_names_the_reason_and_keeps_the_kind() {
    let (home, p) = world();
    let v = refusal(&run(&home, &["card", "take", "C-1", "--project", &p]));
    assert_eq!(v["error"]["kind"], json!("conflict"), "{v}");
    let message = v["error"]["message"].as_str().unwrap();
    assert!(
        message.contains("superseded by another card"),
        "the refusal names the withdrawal reason: {message}"
    );
}

#[test]
fn list_omits_the_withdrawn_card_until_all_is_asked() {
    let (home, p) = world();
    let before = home.snapshot();
    let listed = result(&run(&home, &["card", "list", &p]));
    assert_eq!(listed["count"], json!(1), "{listed}");
    assert_eq!(listed["cards"][0]["name"], json!("C-2"), "{listed}");
    let all = result(&run(&home, &["card", "list", "--all", &p]));
    assert_eq!(all["count"], json!(2), "{all}");
    assert_eq!(home.snapshot(), before, "reading the list wrote nothing");
}

#[test]
fn a_hold_card_without_a_withdraw_keeps_the_plain_refusal() {
    let (home, p) = world();
    {
        let mut w: WriteLedger = ledger::open_write(home.path(), "demo").unwrap();
        w.with_write_tx(|tx, now| {
            Ok(tx.execute(
                "UPDATE card SET state = 'hold', updated_at = ?2 WHERE card_id = 2",
                rusqlite::params![2, now],
            )?)
        })
        .unwrap();
    }
    let v = refusal(&run(&home, &["card", "take", "C-2", "--project", &p]));
    assert_eq!(v["error"]["kind"], json!("conflict"), "{v}");
    let message = v["error"]["message"].as_str().unwrap();
    assert!(
        !message.contains("withdrawn"),
        "a plain hold card keeps the plain wording: {message}"
    );
}
