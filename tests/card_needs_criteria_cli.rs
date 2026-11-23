//! `moai card ready` and `moai card take` refuse a card that has no acceptance criterion
//! (SPEC-20261123-card-needs-criteria): such a card can never be closed, because the done rule of
//! the ledger needs a met criterion. A criterion counts when its status is `open` or `met`; a
//! `withdrawn` one does not. Every command runs against a temporary HOME; the real `~/.moai` is
//! not read or written.

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().expect("run moai")
}

fn as_session(home: &TempHome, session: &str, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, session)
        .output()
        .expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn registered() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    (home, project.to_str().unwrap().to_string())
}

/// Adds a card with a scope and the given criteria; returns its name.
fn add_card(home: &TempHome, project: &str, criteria: &[&str]) -> String {
    let mut args = vec![
        "card",
        "add",
        "--title",
        "t",
        "--scope",
        "src/a.rs",
        "--project",
        project,
    ];
    for ac in criteria {
        args.extend_from_slice(&["--ac", ac]);
    }
    result(&run(home, &args))["name"]
        .as_str()
        .unwrap()
        .to_string()
}

fn sql(home: &TempHome, statement: &str) {
    let mut w = ledger::open_write(home.path(), "demo").unwrap();
    w.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn count(home: &TempHome, query: &str) -> i64 {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    l.conn().query_row(query, [], |r| r.get(0)).unwrap()
}

fn state_of(home: &TempHome, id: i64) -> String {
    let l = ledger::open_read_only(home.path(), "demo").unwrap();
    l.conn()
        .query_row("SELECT state FROM card WHERE card_id = ?1", [id], |r| {
            r.get(0)
        })
        .unwrap()
}

fn ready(home: &TempHome, project: &str, card: &str) -> Output {
    run(home, &["card", "ready", card, "--project", project])
}

fn take(home: &TempHome, project: &str, card: &str) -> Output {
    as_session(home, S_A, &["card", "take", card, "--project", project])
}

fn criterion_add(home: &TempHome, project: &str, card: &str) {
    result(&run(
        home,
        &[
            "card",
            "criterion-add",
            card,
            "--ac",
            "it works",
            "--project",
            project,
        ],
    ));
}

/// The refusal a card without a counted criterion gets, from either command.
fn assert_names_the_way_out(kind: &str, message: &str, action: &str, card: &str) {
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(
        message.starts_with(&format!("card {action} refused: ")),
        "{message}"
    );
    assert!(
        message.contains(&format!("moai card criterion-add {card} --ac")),
        "{message}"
    );
}

#[test]
fn ready_refuses_a_card_without_a_criterion_and_writes_nothing() {
    let (home, project) = registered();
    let card = add_card(&home, &project, &[]);
    let (rev, events) = (
        count(&home, "SELECT rev FROM card WHERE card_id = 1"),
        count(&home, "SELECT count(*) FROM event"),
    );

    let (kind, message) = refusal(&ready(&home, &project, &card));
    assert_names_the_way_out(&kind, &message, "ready", &card);

    assert_eq!(state_of(&home, 1), "draft");
    assert_eq!(count(&home, "SELECT rev FROM card WHERE card_id = 1"), rev);
    assert_eq!(count(&home, "SELECT count(*) FROM event"), events);

    // One criterion later the same command goes through.
    criterion_add(&home, &project, &card);
    let answer = result(&ready(&home, &project, &card));
    assert_eq!(answer["state"], json!("ready"), "{answer}");
}

#[test]
fn ready_refuses_a_card_whose_only_criterion_is_withdrawn() {
    let (home, project) = registered();
    let card = add_card(&home, &project, &["first"]);
    sql(
        &home,
        "UPDATE card_ac SET status = 'withdrawn', withdrawn_reason = 'no longer applies' \
         WHERE card_id = 1",
    );

    let (kind, message) = refusal(&ready(&home, &project, &card));
    assert_names_the_way_out(&kind, &message, "ready", &card);
    assert_eq!(state_of(&home, 1), "draft");

    criterion_add(&home, &project, &card);
    result(&ready(&home, &project, &card));
    assert_eq!(state_of(&home, 1), "ready");
}

#[test]
fn take_refuses_a_ready_card_without_a_criterion_and_reserves_nothing() {
    let (home, project) = registered();
    let card = add_card(&home, &project, &[]);
    // A card that is `ready` without a criterion (made before the rule) is only reachable by SQL.
    sql(&home, "UPDATE card SET state = 'ready' WHERE card_id = 1");
    let events = count(&home, "SELECT count(*) FROM event");

    let (kind, message) = refusal(&take(&home, &project, &card));
    assert_names_the_way_out(&kind, &message, "take", &card);

    assert_eq!(state_of(&home, 1), "ready");
    assert_eq!(count(&home, "SELECT count(*) FROM reservation"), 0);
    assert_eq!(count(&home, "SELECT count(*) FROM event"), events);

    criterion_add(&home, &project, &card);
    let answer = result(&take(&home, &project, &card));
    assert_eq!(answer["state"], json!("active"), "{answer}");
    assert_eq!(count(&home, "SELECT count(*) FROM reservation"), 1);
}

#[test]
fn take_refuses_a_ready_card_whose_only_criterion_is_withdrawn() {
    let (home, project) = registered();
    let card = add_card(&home, &project, &["first"]);
    result(&ready(&home, &project, &card));
    sql(
        &home,
        "UPDATE card_ac SET status = 'withdrawn', withdrawn_reason = 'no longer applies' \
         WHERE card_id = 1",
    );

    let (kind, message) = refusal(&take(&home, &project, &card));
    assert_names_the_way_out(&kind, &message, "take", &card);
    assert_eq!(state_of(&home, 1), "ready");
    assert_eq!(count(&home, "SELECT count(*) FROM reservation"), 0);
}

#[test]
fn a_met_criterion_counts_and_an_earlier_refusal_keeps_its_words() {
    let (home, project) = registered();
    // A met criterion is a counted criterion: the card readies and is taken.
    let met = add_card(&home, &project, &["first"]);
    sql(
        &home,
        "UPDATE card_ac SET status = 'met', evidence_ref = 'e' WHERE card_id = 1",
    );
    result(&ready(&home, &project, &met));
    result(&take(&home, &project, &met));

    // A card without criteria in a state the commands refuse keeps that refusal, not the new one.
    let bare = add_card(&home, &project, &[]);
    let (_, not_ready) = refusal(&take(&home, &project, &bare));
    assert!(!not_ready.contains("criterion-add"), "{not_ready}");
    sql(&home, "UPDATE card SET state = 'ready' WHERE card_id = 2");
    let (_, ready_again) = refusal(&ready(&home, &project, &bare));
    assert!(
        ready_again.contains("only a `draft` or `hold` card becomes `ready`"),
        "{ready_again}"
    );
    assert!(!ready_again.contains("criterion-add"), "{ready_again}");
    let (kind, _) = refusal(&take(&home, &project, "C-9"));
    assert_eq!(kind, "not_found");
}
