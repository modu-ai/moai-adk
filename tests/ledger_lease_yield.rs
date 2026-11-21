//! An ended lease no longer blocks a new take or scope widening, and the holder of an ended
//! lease can renew it (SPEC-20261121-card-lease-yield). Every test runs in its own temporary
//! HOME; none touches the real `~/.moai`. A lease is ended or live by setting `lease_until`
//! to `unixepoch()` (the lease-end second itself no longer holds) or to `unixepoch() + 300`;
//! the one-second edge is pinned on the pure rule `blocks_new_claims`, so no test depends on
//! the clock standing still.

mod common;

use common::home::TempHome;
use moai_cli::ledger::reservation::{self, FenceError, ReserveError, WidenError};
use moai_cli::ledger::workflow::{self, SkipReason};
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

fn add_card(w: &mut WriteLedger, write_scope: &str) -> i64 {
    w.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, priority, write_scope, created_by, created_at, updated_at) \
             VALUES ('card', 'ready', 'p2', ?2, 'test', ?1, ?1)",
            rusqlite::params![now, write_scope],
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

fn paths(items: &[&str]) -> Vec<String> {
    items.iter().map(|s| s.to_string()).collect()
}

/// A card taken by `holder` with a one-hour lease.
fn taken(w: &mut WriteLedger, holder: &str, scope: &[&str]) -> i64 {
    let card = add_card(w, "[]");
    reservation::reserve(w, card, holder, &paths(scope), 3600, None, None).unwrap();
    card
}

fn sql(w: &mut WriteLedger, statement: &str) {
    w.with_write_tx(|tx, _| {
        tx.execute(statement, [])?;
        Ok(())
    })
    .unwrap();
}

/// The lease of `card` ended (the lease-end second) or live (300 seconds left).
fn end_lease(w: &mut WriteLedger, card: i64) {
    sql(
        w,
        &format!("UPDATE reservation SET lease_until = unixepoch() WHERE card_id = {card}"),
    );
}

fn live_lease(w: &mut WriteLedger, card: i64) {
    sql(
        w,
        &format!("UPDATE reservation SET lease_until = unixepoch() + 300 WHERE card_id = {card}"),
    );
}

fn set_status(w: &mut WriteLedger, card: i64, status: &str) {
    sql(
        w,
        &format!("UPDATE reservation SET status = '{status}' WHERE card_id = {card}"),
    );
}

fn scope_of(w: &WriteLedger, card: i64) -> Vec<String> {
    reservation::current(w.conn(), card)
        .unwrap()
        .unwrap()
        .write_scope
}

fn events(w: &WriteLedger, kind: &str) -> Vec<(i64, serde_json::Value)> {
    let mut st = w
        .conn()
        .prepare("SELECT card_id, payload FROM event WHERE kind = ?1 ORDER BY event_id")
        .unwrap();
    st.query_map([kind], |r| {
        Ok((r.get::<_, i64>(0)?, r.get::<_, String>(1)?))
    })
    .unwrap()
    .map(|row| {
        let (card, payload) = row.unwrap();
        (card, serde_json::from_str(&payload).unwrap())
    })
    .collect()
}

fn try_take(
    w: &mut WriteLedger,
    card: i64,
    holder: &str,
    scope: &[&str],
) -> Result<(), ReserveError> {
    reservation::reserve(w, card, holder, &paths(scope), 3600, None, None).map(|_| ())
}

fn widen(
    w: &mut WriteLedger,
    card: i64,
    holder: &str,
    new_scope: &[&str],
) -> Result<(), WidenError> {
    let new = paths(new_scope);
    w.with_write_tx(|tx, now| reservation::widen_in_tx(tx, now, card, &new, holder, "needs it"))
        .unwrap()
        .map(|_| ())
}

/// The first second of the lease end as `utc_from_epoch` writes it, read from the row.
fn lease_end_of(w: &WriteLedger, card: i64) -> i64 {
    reservation::current(w.conn(), card)
        .unwrap()
        .unwrap()
        .lease_until
}

// ---- take_ (REQ-1, REQ-2, REQ-9) ----------------------------------------------------------

/// The lease edge: only a `held` reservation whose lease has not ended blocks a new claim, one
/// second before the end, at the end and one second after it.
#[test]
fn take_lease_edge_is_one_second_before_at_and_after_the_end() {
    assert!(reservation::blocks_new_claims("held", 100, 99), "N-1: live");
    assert!(
        !reservation::blocks_new_claims("held", 100, 100),
        "N: ended"
    );
    assert!(
        !reservation::blocks_new_claims("held", 100, 101),
        "N+1: ended"
    );
    for status in ["held_for_review", "expired_candidate"] {
        for now in [99, 100, 101] {
            assert!(
                reservation::blocks_new_claims(status, 100, now),
                "{status} keeps blocking at {now}"
            );
        }
    }
}

#[test]
fn take_yields_the_paths_of_an_ended_lease() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**", "docs/**"]);
    end_lease(&mut w, old);
    let new = add_card(&mut w, "[]");
    try_take(&mut w, new, B, &["src/a/x.rs"]).unwrap();
    assert_eq!(
        scope_of(&w, old),
        ["docs/**"],
        "only the overlapping pattern left"
    );
    assert_eq!(scope_of(&w, new), ["src/a/x.rs"]);
    // The old card keeps its holder, generation and state.
    let kept = reservation::current(w.conn(), old).unwrap().unwrap();
    assert_eq!(
        (kept.holder.as_str(), kept.generation, kept.status.as_str()),
        (A, 1, "held")
    );
    let state: String = w
        .conn()
        .query_row("SELECT state FROM card WHERE card_id = ?1", [old], |r| {
            r.get(0)
        })
        .unwrap();
    assert_eq!(state, "active");
}

#[test]
fn take_refuses_a_live_lease_and_names_the_card_the_holder_and_the_lease_end() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    live_lease(&mut w, old);
    let end = lease_end_of(&w, old);
    let new = add_card(&mut w, "[]");
    let err = try_take(&mut w, new, B, &["src/a/x.rs"]).unwrap_err();
    assert_eq!(err.code(), "scope_overlap", "{err}");
    let text = err.to_string();
    assert!(
        text.starts_with("reservation conflict in write scope:"),
        "the first words stay: {text}"
    );
    for part in [
        format!("card C-{old}"),
        format!("held by {A}"),
        format!("lease until {end}"),
    ] {
        assert!(text.contains(&part), "{part:?} missing in: {text}");
    }
    assert_eq!(scope_of(&w, old), ["src/a/**"], "nothing was removed");
    assert!(events(&w, "card.yield").is_empty(), "nothing was written");
    assert!(reservation::current(w.conn(), new).unwrap().is_none());
}

#[test]
fn take_refuses_a_review_reservation_even_when_its_lease_ended() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, old);
    set_status(&mut w, old, "held_for_review");
    let new = add_card(&mut w, "[]");
    let err = try_take(&mut w, new, B, &["src/a/x.rs"]).unwrap_err();
    assert_eq!(err.code(), "scope_overlap", "{err}");
    assert_eq!(scope_of(&w, old), ["src/a/**"]);
    assert!(events(&w, "card.yield").is_empty());
}

#[test]
fn take_refuses_an_expired_candidate_even_when_its_lease_ended() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, old);
    set_status(&mut w, old, "expired_candidate");
    let new = add_card(&mut w, "[]");
    let err = try_take(&mut w, new, B, &["src/a/x.rs"]).unwrap_err();
    assert_eq!(err.code(), "scope_overlap", "{err}");
    assert_eq!(scope_of(&w, old), ["src/a/**"]);
}

#[test]
fn take_with_a_live_and_an_ended_overlap_refuses_and_yields_nothing() {
    let (_home, mut w) = setup();
    let ended = taken(&mut w, A, &["src/a/**"]);
    let live = taken(&mut w, A, &["src/b/**"]);
    end_lease(&mut w, ended);
    live_lease(&mut w, live);
    let new = add_card(&mut w, "[]");
    let err = try_take(&mut w, new, B, &["src/a/x.rs", "src/b/y.rs"]).unwrap_err();
    assert_eq!(err.code(), "scope_overlap", "{err}");
    assert_eq!(
        scope_of(&w, ended),
        ["src/a/**"],
        "the ended one is untouched too"
    );
    assert!(events(&w, "card.yield").is_empty());
}

// ---- widen_ (REQ-3, REQ-4, REQ-9) ---------------------------------------------------------

#[test]
fn widen_yields_the_paths_of_an_ended_lease() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**", "docs/**"]);
    end_lease(&mut w, old);
    let mine = taken(&mut w, B, &["tests/b/**"]);
    widen(&mut w, mine, B, &["tests/b/**", "src/a/x.rs"]).unwrap();
    assert_eq!(scope_of(&w, old), ["docs/**"]);
    let mut now_mine = scope_of(&w, mine);
    now_mine.sort();
    assert_eq!(now_mine, ["src/a/x.rs", "tests/b/**"]);
}

#[test]
fn widen_refuses_a_live_lease_and_names_the_card_the_holder_and_the_lease_end() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    live_lease(&mut w, old);
    let end = lease_end_of(&w, old);
    let mine = taken(&mut w, B, &["tests/b/**"]);
    let err = widen(&mut w, mine, B, &["tests/b/**", "src/a/x.rs"]).unwrap_err();
    let text = err.to_string();
    assert!(
        text.starts_with("reservation conflict in write scope:"),
        "the first words stay: {text}"
    );
    for part in [
        format!("card C-{old}"),
        format!("held by {A}"),
        format!("lease until {end}"),
    ] {
        assert!(text.contains(&part), "{part:?} missing in: {text}");
    }
    assert_eq!(scope_of(&w, old), ["src/a/**"]);
    assert_eq!(scope_of(&w, mine), ["tests/b/**"]);
    assert!(events(&w, "card.yield").is_empty());
}

#[test]
fn widen_refuses_a_review_reservation_even_when_its_lease_ended() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, old);
    set_status(&mut w, old, "held_for_review");
    let mine = taken(&mut w, B, &["tests/b/**"]);
    let err = widen(&mut w, mine, B, &["tests/b/**", "src/a/x.rs"]).unwrap_err();
    assert!(matches!(err, WidenError::ScopeOverlap { .. }), "{err}");
    assert_eq!(scope_of(&w, old), ["src/a/**"]);
    assert_eq!(scope_of(&w, mine), ["tests/b/**"]);
}

// ---- event_ (REQ-5) -----------------------------------------------------------------------

#[test]
fn event_yield_names_the_paths_both_scopes_the_holder_the_lease_end_and_the_taker() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**", "docs/**"]);
    end_lease(&mut w, old);
    let end = lease_end_of(&w, old);
    let new = add_card(&mut w, "[]");
    try_take(&mut w, new, B, &["src/a/x.rs"]).unwrap();
    let all = events(&w, "card.yield");
    assert_eq!(all.len(), 1, "{all:?}");
    let (card, p) = &all[0];
    assert_eq!(*card, old, "the event sits on the old card");
    assert_eq!(p["yielded"], serde_json::json!(["src/a/**"]));
    assert_eq!(p["from"], serde_json::json!(["docs/**", "src/a/**"]));
    assert_eq!(p["to"], serde_json::json!(["docs/**"]));
    assert_eq!(p["holder"], serde_json::json!(A));
    assert_eq!(p["generation"], serde_json::json!(1));
    assert_eq!(p["lease_until"], serde_json::json!(end));
    assert_eq!(p["by"], serde_json::json!(B));
    assert_eq!(p["taken_by_card"], serde_json::json!(new));
    let reason = p["reason"].as_str().unwrap();
    assert!(reason.contains("lease had ended"), "{reason}");
}

#[test]
fn event_yield_is_written_for_each_ended_reservation_that_overlaps() {
    let (_home, mut w) = setup();
    let one = taken(&mut w, A, &["src/a/**"]);
    let two = taken(&mut w, A, &["src/b/**"]);
    end_lease(&mut w, one);
    end_lease(&mut w, two);
    let new = add_card(&mut w, "[]");
    try_take(&mut w, new, B, &["src/a/x.rs", "src/b/y.rs"]).unwrap();
    let mut cards: Vec<i64> = events(&w, "card.yield").iter().map(|e| e.0).collect();
    cards.sort();
    assert_eq!(cards, [one, two]);
    assert_eq!(scope_of(&w, one), Vec::<String>::new());
}

#[test]
fn event_yield_from_a_widening_names_the_widening_card() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, old);
    let mine = taken(&mut w, B, &["tests/b/**"]);
    widen(&mut w, mine, B, &["tests/b/**", "src/a/x.rs"]).unwrap();
    let all = events(&w, "card.yield");
    assert_eq!(all.len(), 1, "{all:?}");
    assert_eq!(all[0].1["taken_by_card"], serde_json::json!(mine));
    assert_eq!(all[0].1["by"], serde_json::json!(B));
    assert_eq!(all[0].1["yielded"], serde_json::json!(["src/a/**"]));
}

// ---- renew_ (REQ-6, REQ-7) ----------------------------------------------------------------

#[test]
fn renew_extends_an_ended_lease_at_the_same_generation() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, card);
    let r = reservation::renew(&mut w, card, A, 1, 600).unwrap();
    assert_eq!(
        (r.generation, r.holder.as_str(), r.status.as_str()),
        (1, A, "held")
    );
    let now: i64 = w
        .conn()
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .unwrap();
    assert!(r.lease_until >= now + 599, "{} vs {now}", r.lease_until);
    let renewed = events(&w, "renew");
    assert_eq!(renewed.len(), 1);
    assert_eq!(renewed[0].1["generation"], serde_json::json!(1));
}

#[test]
fn renew_still_extends_a_live_lease_and_never_shortens_it() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, A, &["src/a/**"]);
    let before = lease_end_of(&w, card);
    let r = reservation::renew(&mut w, card, A, 1, 60).unwrap();
    assert!(r.lease_until >= before, "{} < {before}", r.lease_until);
}

#[test]
fn renew_refuses_a_wrong_generation_and_another_session_when_the_lease_ended() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, card);
    let wrong_gen = reservation::renew(&mut w, card, A, 5, 600)
        .map(|_| ())
        .unwrap_err();
    assert!(
        matches!(wrong_gen, FenceError::UnknownGeneration { .. }),
        "{wrong_gen:?}"
    );
    let other = reservation::renew(&mut w, card, B, 1, 600)
        .map(|_| ())
        .unwrap_err();
    assert!(matches!(other, FenceError::WrongHolder { .. }), "{other:?}");
    // The lease did not move.
    let still = reservation::current(w.conn(), card).unwrap().unwrap();
    let now: i64 = w
        .conn()
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .unwrap();
    assert!(
        still.lease_until <= now,
        "refused renews left the lease alone"
    );
}

#[test]
fn renew_after_a_yield_keeps_the_rest_of_the_scope_and_not_the_yielded_path() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**", "docs/**"]);
    end_lease(&mut w, old);
    let new = add_card(&mut w, "[]");
    try_take(&mut w, new, B, &["src/a/x.rs"]).unwrap();
    reservation::renew(&mut w, old, A, 1, 600).unwrap();
    assert_eq!(scope_of(&w, old), ["docs/**"]);
    assert_eq!(scope_of(&w, new), ["src/a/x.rs"]);
}

// ---- next_ (REQ-8, REQ-9) -----------------------------------------------------------------

fn overlap_skips(w: &WriteLedger, card: i64) -> bool {
    let n = workflow::next(w.conn(), None).unwrap();
    assert!(
        n.ready.iter().any(|c| c.card_id == card) != n.skipped.iter().any(|s| s.card_id == card),
        "the card is either ready or skipped"
    );
    n.skipped
        .iter()
        .any(|s| s.card_id == card && matches!(s.reason, SkipReason::ScopeOverlap { .. }))
}

#[test]
fn next_offers_a_ready_card_whose_scope_only_an_ended_lease_overlaps() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    end_lease(&mut w, old);
    let ready = add_card(&mut w, "[\"src/a/x.rs\"]");
    assert!(!overlap_skips(&w, ready));
}

#[test]
fn next_skips_a_ready_card_that_a_live_lease_overlaps() {
    let (_home, mut w) = setup();
    let old = taken(&mut w, A, &["src/a/**"]);
    live_lease(&mut w, old);
    let ready = add_card(&mut w, "[\"src/a/x.rs\"]");
    assert!(overlap_skips(&w, ready));
}

#[test]
fn next_skips_a_ready_card_that_a_review_or_an_expired_candidate_overlaps() {
    for status in ["held_for_review", "expired_candidate"] {
        let (_home, mut w) = setup();
        let old = taken(&mut w, A, &["src/a/**"]);
        end_lease(&mut w, old);
        set_status(&mut w, old, status);
        let ready = add_card(&mut w, "[\"src/a/x.rs\"]");
        assert!(overlap_skips(&w, ready), "{status}");
    }
}
