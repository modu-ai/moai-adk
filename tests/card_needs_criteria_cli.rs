//! `moai card ready` and `moai card take` refuse a card that has no acceptance criterion
//! (SPEC-20261123-card-needs-criteria): such a card can never be closed, because the done rule of
//! the ledger needs a met criterion. A criterion counts when its status is `open` or `met`; a
//! `withdrawn` one does not. Every command runs against a temporary HOME; the real
//! home state is not read or written.

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

// ---- `card next` skips what `card take` refuses (SPEC-20261124-card-next-skips-no-criteria) ----

/// Adds a card with `scope` and the given criteria; returns its name.
fn add_card_in(home: &TempHome, project: &str, scope: &str, criteria: &[&str]) -> String {
    let mut args = vec![
        "card",
        "add",
        "--title",
        "t",
        "--scope",
        scope,
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

/// `moai card next` as `S_A`.
fn next(home: &TempHome, project: &str) -> Value {
    result(&as_session(home, S_A, &["card", "next", project]))
}

fn names(list: &Value) -> Vec<String> {
    list.as_array()
        .unwrap()
        .iter()
        .map(|c| c["name"].as_str().unwrap().to_string())
        .collect()
}

/// The reason `next` gives for skipping `card`, or `None` when it is not skipped.
fn skip_reason(next: &Value, card: &str) -> Option<String> {
    next["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["name"] == json!(card))
        .map(|s| s["reason"].as_str().unwrap().to_string())
}

fn card_id(card: &str) -> i64 {
    card.trim_start_matches("C-").parse().unwrap()
}

/// A `ready` card with no criterion (made before the rule, so only reachable by SQL).
fn ready_without_criterion(home: &TempHome, project: &str, scope: &str) -> String {
    let card = add_card_in(home, project, scope, &[]);
    sql(
        home,
        &format!(
            "UPDATE card SET state = 'ready' WHERE card_id = {}",
            card_id(&card)
        ),
    );
    card
}

fn withdraw_every_criterion(home: &TempHome, card: &str) {
    sql(
        home,
        &format!(
            "UPDATE card_ac SET status = 'withdrawn', withdrawn_reason = 'no longer applies' \
             WHERE card_id = {}",
            card_id(card)
        ),
    );
}

#[test]
fn next_skips_a_ready_card_with_no_criterion_and_take_agrees() {
    let (home, project) = registered();
    let bare = ready_without_criterion(&home, &project, "src/a.rs");

    let before = home.snapshot();
    let answer = next(&home, &project);
    assert_eq!(home.snapshot(), before, "next writes nothing");
    assert_eq!(answer["ready"], json!([]), "{answer}");
    assert_eq!(answer["next"], Value::Null, "{answer}");
    assert_eq!(answer["next_is"], Value::Null, "{answer}");
    assert_eq!(skip_reason(&answer, &bare).as_deref(), Some("no_criteria"));
    assert_eq!(answer["skipped"][0]["state"], json!("ready"), "{answer}");

    let (kind, message) = refusal(&take(&home, &project, &bare));
    assert_names_the_way_out(&kind, &message, "take", &bare);
}

#[test]
fn next_skips_a_ready_card_whose_only_criterion_is_withdrawn_and_take_agrees() {
    let (home, project) = registered();
    let card = add_card_in(&home, &project, "src/a.rs", &["first"]);
    result(&ready(&home, &project, &card));
    withdraw_every_criterion(&home, &card);

    let answer = next(&home, &project);
    assert_eq!(answer["ready"], json!([]), "{answer}");
    assert_eq!(answer["next"], Value::Null, "{answer}");
    assert_eq!(skip_reason(&answer, &card).as_deref(), Some("no_criteria"));

    let (kind, message) = refusal(&take(&home, &project, &card));
    assert_names_the_way_out(&kind, &message, "take", &card);
}

#[test]
fn next_offers_a_ready_card_with_one_open_or_one_met_criterion_and_take_takes_it() {
    let (home, project) = registered();
    let open = add_card_in(&home, &project, "src/a.rs", &["first"]);
    let met = add_card_in(&home, &project, "src/b.rs", &["first"]);
    sql(
        &home,
        "UPDATE card_ac SET status = 'met', evidence_ref = 'e' WHERE card_id = 2",
    );
    result(&ready(&home, &project, &open));
    result(&ready(&home, &project, &met));

    let answer = next(&home, &project);
    assert_eq!(names(&answer["ready"]), [open.clone(), met.clone()]);
    assert_eq!(answer["next"], json!(open), "{answer}");
    assert_eq!(answer["next_is"], json!("ready"), "{answer}");
    assert_eq!(answer["skipped"], json!([]), "{answer}");

    result(&take(&home, &project, &open));
    result(&take(&home, &project, &met));
}

#[test]
fn a_criterion_added_later_moves_a_skipped_card_to_ready() {
    let (home, project) = registered();
    let bare = ready_without_criterion(&home, &project, "src/a.rs");
    assert_eq!(
        skip_reason(&next(&home, &project), &bare).as_deref(),
        Some("no_criteria")
    );

    criterion_add(&home, &project, &bare);
    let answer = next(&home, &project);
    assert_eq!(
        names(&answer["ready"]),
        std::slice::from_ref(&bare),
        "{answer}"
    );
    assert_eq!(answer["next"], json!(bare), "{answer}");
    assert_eq!(skip_reason(&answer, &bare), None, "{answer}");
    result(&take(&home, &project, &bare));
}

#[test]
fn next_keeps_the_other_skip_reasons_whether_or_not_the_card_has_a_criterion() {
    let (home, project) = registered();
    let other = "bbbbbbbb-0000-4000-8000-00000000000b";

    // `held`: another session holds the card, then its only criterion is withdrawn.
    let held = add_card_in(&home, &project, "src/held.rs", &["first"]);
    result(&ready(&home, &project, &held));
    result(&as_session(
        &home,
        other,
        &["card", "take", &held, "--project", &project],
    ));
    withdraw_every_criterion(&home, &held);

    // `empty_scope`: a ready card with neither scope nor criterion.
    let scopeless = result(&run(
        &home,
        &["card", "add", "--title", "t", "--project", &project],
    ))["name"]
        .as_str()
        .unwrap()
        .to_string();
    sql(
        &home,
        &format!(
            "UPDATE card SET state = 'ready' WHERE card_id = {}",
            card_id(&scopeless)
        ),
    );

    // `review_no_holder`: a card in review that nobody holds, its only criterion withdrawn.
    let review = add_card_in(&home, &project, "src/review.rs", &["first"]);
    result(&ready(&home, &project, &review));
    result(&take(&home, &project, &review));
    sql(
        &home,
        &format!(
            "UPDATE card SET state = 'review' WHERE card_id = {id}; \
             UPDATE reservation SET status = 'released' WHERE card_id = {id}",
            id = card_id(&review)
        ),
    );
    withdraw_every_criterion(&home, &review);

    let answer = next(&home, &project);
    assert_eq!(answer["ready"], json!([]), "{answer}");
    assert_eq!(skip_reason(&answer, &held).as_deref(), Some("held"));
    assert_eq!(
        skip_reason(&answer, &scopeless).as_deref(),
        Some("empty_scope")
    );
    assert_eq!(
        skip_reason(&answer, &review).as_deref(),
        Some("review_no_holder")
    );
}
