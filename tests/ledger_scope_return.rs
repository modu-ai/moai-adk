//! Giving paths back from a held reservation in the ledger library
//! (SPEC-20261120-card-shared-line-lease): `give_back`, `review_keeping` and the two steps of a
//! line lease. Every test runs in its own temporary HOME; none touches the real `~/.moai`.

mod common;

use common::home::TempHome;
use moai_cli::ledger::reservation::{self, FenceError};
use moai_cli::ledger::scope_return::{
    self, GiveBackError, LeaseError, ReviewKeepError, ScopeReturnError,
};
use moai_cli::ledger::workflow::WorkflowError;
use moai_cli::ledger::{self, WriteLedger};

const PROJECT: &str = "demo";
const A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

fn setup() -> (TempHome, WriteLedger) {
    let home = TempHome::new();
    ledger::register(home.path(), PROJECT).unwrap();
    let w = ledger::open_write(home.path(), PROJECT).unwrap();
    (home, w)
}

fn add_card(w: &mut WriteLedger) -> i64 {
    w.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, priority, write_scope, created_by, created_at, updated_at) \
             VALUES ('card', 'ready', 'p2', '[]', 'test', ?1, ?1)",
            rusqlite::params![now],
        )?;
        let id = tx.last_insert_rowid();
        tx.execute(
            "INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES (?1, 1, 'ac', ?2)",
            rusqlite::params![id, now],
        )?;
        Ok(id)
    })
    .unwrap()
}

fn taken(w: &mut WriteLedger, holder: &str, scope: &[&str]) -> (i64, i64) {
    let card = add_card(w);
    let scope: Vec<String> = scope.iter().map(|s| s.to_string()).collect();
    let r = reservation::reserve(w, card, holder, &scope, 3600, None, None).unwrap();
    (card, r.generation)
}

fn paths(items: &[&str]) -> Vec<String> {
    items.iter().map(|s| s.to_string()).collect()
}

fn scope_of(w: &WriteLedger, card: i64) -> Vec<String> {
    reservation::current(w.conn(), card)
        .unwrap()
        .unwrap()
        .write_scope
}

fn events(w: &WriteLedger, kind: &str) -> Vec<serde_json::Value> {
    let mut st = w
        .conn()
        .prepare("SELECT payload FROM event WHERE kind = ?1 ORDER BY event_id")
        .unwrap();
    st.query_map([kind], |r| r.get::<_, String>(0))
        .unwrap()
        .map(|p| serde_json::from_str(&p.unwrap()).unwrap())
        .collect()
}

#[test]
fn give_back_removes_the_paths_and_a_second_card_takes_them_at_once() {
    let (_home, mut w) = setup();
    let (a, g) = taken(&mut w, A, &["src/a.rs", "tests/a.rs"]);
    let done =
        scope_return::give_back(&mut w, a, A, g, &paths(&["src/a.rs"]), "give-back").unwrap();
    assert_eq!(done.given_back, ["src/a.rs"]);
    assert_eq!(done.from, ["src/a.rs", "tests/a.rs"]);
    assert_eq!(scope_of(&w, a), ["tests/a.rs"]);
    let (b, _) = taken(&mut w, B, &["src/a.rs"]);
    assert_eq!(scope_of(&w, b), ["src/a.rs"]);
    let ev = events(&w, "card.scope");
    assert_eq!(
        ev.last().unwrap()["given_back"],
        serde_json::json!(["src/a.rs"])
    );
}

#[test]
fn give_back_refuses_an_unknown_path_a_stale_generation_and_another_holder() {
    let (_home, mut w) = setup();
    let (a, g) = taken(&mut w, A, &["src/a.rs"]);
    match scope_return::give_back(&mut w, a, A, g, &paths(&["src/never.rs"]), "give-back") {
        Err(GiveBackError::Scope(ScopeReturnError::NotInReservation { paths })) => {
            assert_eq!(paths, ["src/never.rs"]);
        }
        other => panic!("{other:?}"),
    }
    assert!(matches!(
        scope_return::give_back(&mut w, a, A, g + 1, &paths(&["src/a.rs"]), "give-back"),
        Err(GiveBackError::Fence(FenceError::UnknownGeneration { .. }))
    ));
    assert!(matches!(
        scope_return::give_back(&mut w, a, B, g, &paths(&["src/a.rs"]), "give-back"),
        Err(GiveBackError::Fence(FenceError::WrongHolder { .. }))
    ));
    assert!(matches!(
        scope_return::give_back(&mut w, a, A, g, &[], "give-back"),
        Err(GiveBackError::Scope(ScopeReturnError::InvalidInput(_)))
    ));
    assert_eq!(scope_of(&w, a), ["src/a.rs"], "refusals write nothing");
}

#[test]
fn review_keeping_gives_back_all_but_the_kept_paths_in_one_event() {
    let (_home, mut w) = setup();
    let (a, g) = taken(&mut w, A, &["src/a.rs", "tests/a.rs", "spec/x.md"]);
    let (moved, given, kept) =
        scope_return::review_keeping(&mut w, a, A, g, &paths(&["spec/x.md"])).unwrap();
    assert_eq!(moved.state, "review");
    assert_eq!(given, ["src/a.rs", "tests/a.rs"]);
    assert_eq!(kept, ["spec/x.md"]);
    assert_eq!(scope_of(&w, a), ["spec/x.md"]);
    assert_eq!(moved.reservation.status, "held");
    let ev = events(&w, "card.review");
    assert_eq!(ev.len(), 1);
    assert_eq!(
        ev[0]["given_back"],
        serde_json::json!(["src/a.rs", "tests/a.rs"])
    );
    assert_eq!(ev[0]["kept"], serde_json::json!(["spec/x.md"]));
}

#[test]
fn review_keeping_refuses_a_keep_path_not_held_and_a_card_not_active() {
    let (_home, mut w) = setup();
    let (a, g) = taken(&mut w, A, &["src/a.rs"]);
    match scope_return::review_keeping(&mut w, a, A, g, &paths(&["src/other.rs"])) {
        Err(ReviewKeepError::Scope(ScopeReturnError::NotInReservation { paths })) => {
            assert_eq!(paths, ["src/other.rs"]);
        }
        other => panic!("{other:?}"),
    }
    let state: String = w
        .conn()
        .query_row("SELECT state FROM card WHERE card_id = ?1", [a], |r| {
            r.get(0)
        })
        .unwrap();
    assert_eq!(state, "active", "a refused review writes nothing");
    assert_eq!(scope_of(&w, a), ["src/a.rs"]);
    scope_return::review_keeping(&mut w, a, A, g, &[]).unwrap();
    assert!(matches!(
        scope_return::review_keeping(&mut w, a, A, g, &[]),
        Err(ReviewKeepError::Workflow(WorkflowError::WrongState { .. }))
    ));
}

#[test]
fn a_line_lease_adds_a_path_and_gives_back_exactly_what_it_added() {
    let (_home, mut w) = setup();
    let (a, g) = taken(&mut w, A, &["src/a.rs"]);
    let leased = scope_return::lease_take(&mut w, a, A, g, &paths(&["src/lib.rs"])).unwrap();
    assert_eq!(leased.added, ["src/lib.rs"]);
    assert_eq!(scope_of(&w, a), ["src/a.rs", "src/lib.rs"]);
    let released = scope_return::lease_release(&mut w, a, A, g, &leased.added)
        .unwrap()
        .unwrap();
    assert_eq!(released.given_back, ["src/lib.rs"]);
    assert_eq!(scope_of(&w, a), ["src/a.rs"]);
    // A path the reservation already holds is neither added nor given back.
    let held = scope_return::lease_take(&mut w, a, A, g, &paths(&["src/a.rs"])).unwrap();
    assert!(held.added.is_empty());
    assert!(scope_return::lease_release(&mut w, a, A, g, &held.added)
        .unwrap()
        .is_none());
    assert_eq!(scope_of(&w, a), ["src/a.rs"]);
}

#[test]
fn a_line_lease_on_a_path_another_card_holds_is_an_overlap_that_names_that_card() {
    let (_home, mut w) = setup();
    let (owner, _) = taken(&mut w, B, &["src/lib.rs"]);
    let (a, g) = taken(&mut w, A, &["src/a.rs"]);
    match scope_return::lease_take(&mut w, a, A, g, &paths(&["src/lib.rs"])) {
        Err(LeaseError::Overlap(text)) => {
            assert!(text.contains(&format!("C-{owner}")), "{text}");
        }
        other => panic!("{other:?}"),
    }
    assert_eq!(
        scope_of(&w, a),
        ["src/a.rs"],
        "a refused lease writes nothing"
    );
    // Once the owner gives it back, the lease goes through.
    let og = reservation::current(w.conn(), owner)
        .unwrap()
        .unwrap()
        .generation;
    scope_return::give_back(&mut w, owner, B, og, &paths(&["src/lib.rs"]), "give-back").unwrap();
    let leased = scope_return::lease_take(&mut w, a, A, g, &paths(&["src/lib.rs"])).unwrap();
    assert_eq!(leased.added, ["src/lib.rs"]);
}

#[test]
fn a_mixed_list_of_known_and_unknown_paths_removes_nothing() {
    let (_home, mut w) = setup();
    let (a, g) = taken(&mut w, A, &["src/a.rs", "docs/a.md"]);
    for list in [
        &["src/a.rs", "src/never.rs"][..],
        &["src/never.rs", "src/a.rs"][..],
    ] {
        assert!(matches!(
            scope_return::give_back(&mut w, a, A, g, &paths(list), "give-back"),
            Err(GiveBackError::Scope(
                ScopeReturnError::NotInReservation { .. }
            ))
        ));
    }
    let mut scope = scope_of(&w, a);
    scope.sort();
    assert_eq!(scope, ["docs/a.md", "src/a.rs"], "nothing was removed");
}
