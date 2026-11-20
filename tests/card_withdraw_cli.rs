//! `card withdraw` (no schema change): a withdrawn card is refused by `card take` with its
//! withdrawal reason, and the default `card list` omits it while an explicit state filter
//! still answers as before. The withdrawn mark is the free event kind `card.withdraw` beside
//! the existing `hold` state, written through the wired `card withdraw` CLI action or the
//! ledger function it calls. Every case runs in a temporary HOME; the real `~/.moai` is not
//! read or written.

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

/// A registered project with two ready cards: C-1 and C-2, both plain.
fn two_ready_cards() -> (TempHome, String) {
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
    (home, p)
}

/// [`two_ready_cards`] with C-1 withdrawn through the ledger function the CLI action calls.
fn world() -> (TempHome, String) {
    let (home, p) = two_ready_cards();
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

/// The wired `card withdraw` action end to end: the card moves to hold, take refuses naming
/// the reason, the default list hides it, and an explicit state filter still answers as
/// before (pins audit findings B-1 and B-2).
#[test]
fn the_wired_withdraw_action_refuses_take_and_hides_from_the_default_list() {
    let (home, p) = two_ready_cards();
    // `project` is the positional argument of `card withdraw` (default ".").
    let v = result(&run(
        &home,
        &[
            "card",
            "withdraw",
            "C-1",
            "--reason",
            "superseded by a newer card",
            "--by",
            "leader",
            &p,
        ],
    ));
    assert_eq!(v["state"], json!("hold"), "{v}");
    let r = refusal(&run(&home, &["card", "take", "C-1", "--project", &p]));
    assert_eq!(r["error"]["kind"], json!("conflict"), "{r}");
    assert!(
        r["error"]["message"]
            .as_str()
            .unwrap()
            .contains("superseded by a newer card"),
        "{r}"
    );
    let listed = result(&run(&home, &["card", "list", &p]));
    assert_eq!(listed["count"], json!(1), "{listed}");
    // An explicit state filter behaves as before: the withdrawn card is shown when asked.
    let holds = result(&run(&home, &["card", "list", "--state", "hold", &p]));
    assert_eq!(holds["count"], json!(1), "{holds}");
    assert_eq!(holds["cards"][0]["name"], json!("C-1"), "{holds}");
}
