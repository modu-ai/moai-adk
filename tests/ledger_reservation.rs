//! Reservations, generations, write scopes and the fence (Issue #3068 P3; AC1, AC2, AC3 and the
//! first half of AC5). Every test runs in its own temporary HOME; none touches the real
//! `~/.moai`. The multi-process tests start this same test binary again (`--exact` on a child
//! test that does nothing unless its environment variable is set).

mod common;

use common::home::TempHome;
use common::CleanEnv;
use moai_cli::ledger::reservation::{
    self, FenceError, ReserveError, MAX_FIELD_LEN, MAX_HOLDER_LEN, MAX_LEASE_SECS, MAX_PATTERN_LEN,
    MAX_SCOPE_PATTERNS, MIN_LEASE_SECS,
};
use moai_cli::ledger::workflow::{self, WorkflowError};
use moai_cli::ledger::{self, LedgerError, WriteLedger};
use moai_cli::scope;
use rusqlite::Connection;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant};

const PROJECT: &str = "demo";

fn setup() -> (TempHome, WriteLedger) {
    let home = TempHome::new();
    ledger::register(home.path(), PROJECT).unwrap();
    let w = ledger::open_write(home.path(), PROJECT).unwrap();
    (home, w)
}

/// Inserts a card in `state` directly (an `INSERT` has no state check) and returns its id.
fn add_card(w: &mut WriteLedger, state: &str) -> i64 {
    w.with_write_tx(|tx, now| {
        tx.execute(
            "INSERT INTO card (title, state, created_by, created_at, updated_at) \
             VALUES ('card', ?1, 'test', ?2, ?2)",
            rusqlite::params![state, now],
        )?;
        Ok(tx.last_insert_rowid())
    })
    .unwrap()
}

fn ready_card(w: &mut WriteLedger) -> i64 {
    add_card(w, "ready")
}

fn s(patterns: &[&str]) -> Vec<String> {
    patterns.iter().map(|p| p.to_string()).collect()
}

/// A scope no other card of the test uses.
fn zone(card: i64) -> Vec<String> {
    vec![format!("zone{card}/")]
}

fn take(
    w: &mut WriteLedger,
    card: i64,
    holder: &str,
    scope: &[String],
) -> Result<reservation::Reservation, ReserveError> {
    reservation::reserve(w, card, holder, scope, 3600, None, None)
}

fn sql(w: &mut WriteLedger, statement: &str) {
    w.with_write_tx(|tx, _| {
        tx.execute_batch(statement)?;
        Ok(())
    })
    .unwrap();
}

fn card_state(conn: &Connection, card: i64) -> (String, i64) {
    conn.query_row(
        "SELECT state, generation FROM card WHERE card_id = ?1",
        [card],
        |r| Ok((r.get(0)?, r.get(1)?)),
    )
    .unwrap()
}

/// `(generation, holder, status)` of every reservation of the card, oldest first.
fn rows(conn: &Connection, card: i64) -> Vec<(i64, String, String)> {
    conn.prepare(
        "SELECT generation, holder, status FROM reservation WHERE card_id = ?1 ORDER BY generation",
    )
    .unwrap()
    .query_map([card], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)))
    .unwrap()
    .map(Result::unwrap)
    .collect()
}

/// `(card_id, actor, payload)` of every event of `kind`, oldest first.
fn events(conn: &Connection, kind: &str) -> Vec<(Option<i64>, String, Value)> {
    conn.prepare("SELECT card_id, actor, payload FROM event WHERE kind = ?1 ORDER BY event_id")
        .unwrap()
        .query_map([kind], |r| {
            let payload: String = r.get(2)?;
            Ok((
                r.get(0)?,
                r.get(1)?,
                serde_json::from_str(&payload).unwrap(),
            ))
        })
        .unwrap()
        .map(Result::unwrap)
        .collect()
}

fn count(conn: &Connection, table: &str) -> i64 {
    conn.query_row(&format!("SELECT count(*) FROM {table}"), [], |r| r.get(0))
        .unwrap()
}

fn db_now(conn: &Connection) -> i64 {
    conn.query_row("SELECT unixepoch()", [], |r| r.get(0))
        .unwrap()
}

fn reader(home: &TempHome) -> ledger::ReadLedger {
    ledger::open_read_only(home.path(), PROJECT).unwrap()
}

fn rejection_code<T: std::fmt::Debug>(r: Result<T, ReserveError>) -> &'static str {
    r.unwrap_err().code()
}

// ---- reserve -----------------------------------------------------------------------------------

#[test]
fn reserve_takes_a_ready_card_and_records_everything_on_the_db_clock() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    let r = reservation::reserve(
        &mut w,
        card,
        "lane-a",
        &s(&["src/b.rs", "src/a.rs", "src/b.rs"]),
        900,
        Some("/work/tree"),
        Some("issue/1-x"),
    )
    .unwrap();
    assert_eq!(r.card_id, card);
    assert_eq!(r.generation, 1);
    assert_eq!(r.holder, "lane-a");
    assert_eq!(r.status, "held");
    assert_eq!(r.lease_until - r.acquired_at, 900);
    assert_eq!(r.updated_at, r.acquired_at);
    assert_eq!(r.worktree.as_deref(), Some("/work/tree"));
    assert_eq!(r.branch.as_deref(), Some("issue/1-x"));
    assert_eq!(
        r.write_scope,
        s(&["src/a.rs", "src/b.rs"]),
        "sorted, no duplicate"
    );
    let system = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64;
    assert!(
        (system - r.acquired_at).abs() < 120,
        "the DB clock, {system} vs {}",
        r.acquired_at
    );

    let conn = reader(&home);
    let conn = conn.conn();
    assert_eq!(card_state(conn, card), ("active".to_string(), 1));
    assert_eq!(
        rows(conn, card),
        vec![(1, "lane-a".to_string(), "held".to_string())]
    );
    assert_eq!(reservation::current(conn, card).unwrap(), Some(r.clone()));
    let ev = events(conn, "reserve");
    assert_eq!(ev.len(), 1);
    assert_eq!(ev[0].0, Some(card));
    assert_eq!(ev[0].1, "lane-a");
    assert_eq!(ev[0].2["generation"], 1);
    assert_eq!(
        ev[0].2["write_scope"],
        serde_json::json!(["src/a.rs", "src/b.rs"])
    );
    let key: String = conn
        .query_row(
            "SELECT event_key FROM event WHERE kind = 'reserve'",
            [],
            |r| r.get(0),
        )
        .unwrap();
    assert!(key.starts_with(&format!("reserve:{card}:")), "{key}");
}

#[test]
fn reserve_refusals_leave_the_ledger_byte_for_byte_as_it_was() {
    let (home, mut w) = setup();
    let draft = add_card(&mut w, "draft");
    let hold = add_card(&mut w, "hold");
    let active_without_row = add_card(&mut w, "active");
    let taken = ready_card(&mut w);
    take(&mut w, taken, "first", &s(&["src/taken/"])).unwrap();
    let free = ready_card(&mut w);
    let before = home.snapshot();

    assert_eq!(
        rejection_code(take(&mut w, 9999, "a", &zone(1))),
        "card_not_found"
    );
    for card in [draft, hold, active_without_row] {
        assert_eq!(
            rejection_code(take(&mut w, card, "a", &zone(card))),
            "card_not_ready"
        );
    }
    // The same card again: by another holder and by the holder itself.
    for holder in ["second", "first"] {
        match take(&mut w, taken, holder, &s(&["elsewhere/"])).unwrap_err() {
            ReserveError::AlreadyHeld { holder, generation } => {
                assert_eq!((holder.as_str(), generation), ("first", 1))
            }
            other => panic!("{other:?}"),
        }
    }
    // Overlap with the other card's active reservation.
    match take(&mut w, free, "a", &s(&["docs/", "src/taken/inner.rs"])).unwrap_err() {
        ReserveError::ScopeOverlap {
            other_card,
            mine,
            theirs,
        } => assert_eq!(
            (other_card, mine.as_str(), theirs.as_str()),
            (taken, "src/taken/inner.rs", "src/taken/")
        ),
        other => panic!("{other:?}"),
    }
    assert_eq!(rejection_code(take(&mut w, free, "a", &[])), "empty_scope");
    assert_eq!(
        reservation::reserve(&mut w, free, "a", &zone(free), 3600, None, Some(""))
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(home.snapshot(), before, "no event, no row, no state change");
    assert_eq!(
        card_state(reader(&home).conn(), free),
        ("ready".to_string(), 0)
    );
}

#[test]
fn a_released_card_can_be_reserved_again_with_the_next_generation() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    let first = take(&mut w, card, "a", &s(&["src/x/"])).unwrap();
    let released = reservation::release(&mut w, card, "a", 1).unwrap();
    assert_eq!(released.card_state, "ready");
    assert_eq!(released.reservation.status, "released");
    let second = take(&mut w, card, "b", &s(&["src/y/"])).unwrap();
    assert_eq!(second.generation, 2);
    assert_ne!(first.reservation_id, second.reservation_id);
    let all = reservation::history(reader(&home).conn(), card).unwrap();
    assert_eq!(
        all.iter()
            .map(|r| (r.generation, r.holder.as_str(), r.status.as_str()))
            .collect::<Vec<_>>(),
        vec![(1, "a", "released"), (2, "b", "held")]
    );
    assert_eq!(
        card_state(reader(&home).conn(), card),
        ("active".to_string(), 2)
    );
}

// ---- AC2: write scopes ---------------------------------------------------------------------------

/// `(card A's pattern, card B's pattern, whether `scope::overlap` says they overlap)`.
const SCOPE_CASES: &[(&str, &str, bool)] = &[
    ("src/a.rs", "src/a.rs", true),
    ("src/a", "src/a/b.rs", true),
    ("src/a/b.rs", "src/a", true),
    ("src/a", "src/b", false),
    ("src/ab", "src/a", false),
    ("src/*.rs", "src/main.rs", true),
    ("src/main.rs", "src/*.rs", true),
    ("src/*.rs", "notes/x.txt", false),
    ("src/a/**", "src/a/x/y", true),
    ("src/a/**", "src/b/x", false),
    ("**", "any/where/at/all", true),
    ("SRC/A", "src/a/b", true),
    ("src/a/../b", "src/b/x", true),
    ("src/a/../b", "src/a/x", false),
    ("../outside", "src/a", true),
];

#[test]
fn scope_overlap_decides_between_two_cards_exactly_as_the_shared_predicate() {
    for (a, b, expected) in SCOPE_CASES {
        assert_eq!(
            scope::overlap(a, b),
            *expected,
            "the predicate for {a:?} / {b:?}"
        );
        let (home, mut w) = setup();
        let (one, two) = (ready_card(&mut w), ready_card(&mut w));
        take(&mut w, one, "a", &s(&[a])).unwrap();
        let second = take(&mut w, two, "b", &s(&[b]));
        match (*expected, second) {
            (
                true,
                Err(ReserveError::ScopeOverlap {
                    other_card,
                    mine,
                    theirs,
                }),
            ) => {
                assert_eq!(
                    (other_card, mine.as_str(), theirs.as_str()),
                    (one, *b, *a),
                    "{a:?} / {b:?}"
                )
            }
            (false, Ok(r)) => assert_eq!(r.generation, 1, "{a:?} / {b:?}"),
            (_, other) => panic!("{a:?} / {b:?}: expected overlap {expected}, got {other:?}"),
        }
        let still = card_state(reader(&home).conn(), two);
        assert_eq!(
            still.0,
            if *expected { "ready" } else { "active" },
            "{a:?} / {b:?}"
        );
    }
}

#[test]
fn only_other_cards_active_reservations_count_for_the_overlap() {
    let (home, mut w) = setup();
    let (one, two, three) = (ready_card(&mut w), ready_card(&mut w), ready_card(&mut w));
    take(&mut w, one, "a", &s(&["src/shared/"])).unwrap();
    // One pattern of several is enough.
    assert_eq!(
        rejection_code(take(&mut w, two, "b", &s(&["docs/", "src/shared/x"]))),
        "scope_overlap"
    );
    // A handoff keeps the scope reserved.
    reservation::handoff(&mut w, one, "a", 1, "c").unwrap();
    assert_eq!(
        rejection_code(take(&mut w, two, "b", &s(&["src/shared/"]))),
        "scope_overlap"
    );
    // A released reservation no longer blocks.
    reservation::release(&mut w, one, "c", 2).unwrap();
    take(&mut w, two, "b", &s(&["src/shared/"])).unwrap();
    // The card that holds the scope again blocks the third.
    assert_eq!(
        rejection_code(take(&mut w, three, "d", &s(&["src/shared/y"]))),
        "scope_overlap"
    );
    // The reserved card can take its own scope again later (no self-conflict across generations).
    reservation::release(&mut w, two, "b", 1).unwrap();
    take(&mut w, one, "a", &s(&["src/shared/"])).unwrap();
    assert_eq!(rows(reader(&home).conn(), one).len(), 3);
}

#[test]
fn renew_and_handoff_do_not_conflict_with_the_cards_own_scope() {
    let (home, mut w) = setup();
    let (one, two) = (ready_card(&mut w), ready_card(&mut w));
    take(&mut w, one, "a", &s(&["src/own/", "src/own/deeper.rs"])).unwrap();
    take(&mut w, two, "b", &s(&["docs/"])).unwrap();
    let renewed = reservation::renew(&mut w, one, "a", 1, 600).unwrap();
    assert_eq!(renewed.write_scope, s(&["src/own/", "src/own/deeper.rs"]));
    let moved = reservation::handoff(&mut w, one, "a", 1, "c").unwrap();
    assert_eq!((moved.generation, moved.holder.as_str()), (2, "c"));
    assert_eq!(moved.write_scope, s(&["src/own/", "src/own/deeper.rs"]));
    reservation::renew(&mut w, one, "c", 2, 600).unwrap();
    assert_eq!(
        count(reader(&home).conn(), "event"),
        2 + 3,
        "two reserves, renew, handoff, renew"
    );
}

#[test]
fn an_empty_write_scope_is_refused_because_no_overlap_can_be_shown() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    let before = home.snapshot();
    assert_eq!(
        take(&mut w, card, "a", &[]).unwrap_err(),
        ReserveError::EmptyScope
    );
    assert_eq!(home.snapshot(), before);
    assert_eq!(
        card_state(reader(&home).conn(), card),
        ("ready".to_string(), 0)
    );
}

// ---- limits (N accepted, N + 1 refused) --------------------------------------------------------------

#[test]
fn reserve_limits_hold_at_n_and_fail_at_n_plus_one() {
    let (home, mut w) = setup();
    // Lease seconds.
    for (secs, ok) in [
        (MIN_LEASE_SECS - 1, false),
        (MIN_LEASE_SECS, true),
        (MAX_LEASE_SECS, true),
        (MAX_LEASE_SECS + 1, false),
        (0, false),
        (-5, false),
    ] {
        let card = ready_card(&mut w);
        let before = home.snapshot();
        let r = reservation::reserve(&mut w, card, "a", &zone(card), secs, None, None);
        match (ok, r) {
            (true, Ok(r)) => assert_eq!(r.lease_until - r.acquired_at, secs),
            (false, Err(ReserveError::InvalidInput(_))) => assert_eq!(home.snapshot(), before),
            (_, other) => panic!("lease {secs}: {other:?}"),
        }
    }
    // Holder length: 1 and 200 accepted, 0 and 201 refused.
    for (holder, ok) in [
        ("h".to_string(), true),
        ("h".repeat(MAX_HOLDER_LEN), true),
        (String::new(), false),
        ("h".repeat(MAX_HOLDER_LEN + 1), false),
    ] {
        let card = ready_card(&mut w);
        let before = home.snapshot();
        let r = take(&mut w, card, &holder, &zone(card));
        assert_eq!(
            r.is_ok(),
            ok,
            "holder of {} characters",
            holder.chars().count()
        );
        if !ok {
            assert_eq!(r.unwrap_err().code(), "invalid_input");
            assert_eq!(home.snapshot(), before);
        }
    }
    // Number of patterns: 64 accepted, 65 refused.
    for (n, ok) in [(MAX_SCOPE_PATTERNS, true), (MAX_SCOPE_PATTERNS + 1, false)] {
        let card = ready_card(&mut w);
        let patterns: Vec<String> = (0..n).map(|i| format!("many{card}/{i}/")).collect();
        let before = home.snapshot();
        let r = take(&mut w, card, "a", &patterns);
        assert_eq!(r.is_ok(), ok, "{n} patterns");
        if ok {
            assert_eq!(r.unwrap().write_scope.len(), n);
        } else {
            assert_eq!(r.unwrap_err().code(), "invalid_input");
            assert_eq!(home.snapshot(), before);
        }
    }
    // Pattern length: 512 accepted, 513 refused.
    for (n, ok) in [(MAX_PATTERN_LEN, true), (MAX_PATTERN_LEN + 1, false)] {
        let card = ready_card(&mut w);
        let pattern = format!("len{card}/{}", "x".repeat(n - format!("len{card}/").len()));
        assert_eq!(pattern.chars().count(), n);
        let before = home.snapshot();
        let r = take(&mut w, card, "a", &[pattern]);
        assert_eq!(r.is_ok(), ok, "pattern of {n} characters");
        if !ok {
            assert_eq!(home.snapshot(), before);
        }
    }
    // Worktree and branch length: 512 accepted, 513 refused.
    for (n, ok) in [(MAX_FIELD_LEN, true), (MAX_FIELD_LEN + 1, false)] {
        for which in ["worktree", "branch"] {
            let card = ready_card(&mut w);
            let long = "w".repeat(n);
            let (tree, branch) = if which == "worktree" {
                (Some(long.as_str()), None)
            } else {
                (None, Some(long.as_str()))
            };
            let before = home.snapshot();
            let r = reservation::reserve(&mut w, card, "a", &zone(card), 3600, tree, branch);
            assert_eq!(r.is_ok(), ok, "{which} of {n} characters");
            if !ok {
                assert_eq!(home.snapshot(), before);
            }
        }
    }
}

#[test]
fn renew_release_handoff_and_fence_check_the_same_input_limits() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();
    let before = home.snapshot();
    let long = "h".repeat(MAX_HOLDER_LEN + 1);
    for secs in [MIN_LEASE_SECS - 1, MAX_LEASE_SECS + 1] {
        let e = reservation::renew(&mut w, card, "a", 1, secs).unwrap_err();
        assert_eq!(e.code(), "invalid_input", "lease {secs}");
    }
    assert_eq!(
        reservation::renew(&mut w, card, "", 1, 600)
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::renew(&mut w, card, &long, 1, 600)
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::release(&mut w, card, &long, 1)
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::fence(&mut w, card, &long, 1)
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::handoff(&mut w, card, &long, 1, "b")
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::handoff(&mut w, card, "a", 1, &long)
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::handoff(&mut w, card, "a", 1, "")
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::handoff(&mut w, card, "a", 1, "a")
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(home.snapshot(), before, "bad input records nothing");

    // Just inside: the longest holder, the shortest and the longest lease.
    reservation::renew(&mut w, card, "a", 1, MIN_LEASE_SECS).unwrap();
    reservation::renew(&mut w, card, "a", 1, MAX_LEASE_SECS).unwrap();
    let max_name = "h".repeat(MAX_HOLDER_LEN);
    let moved = reservation::handoff(&mut w, card, "a", 1, &max_name).unwrap();
    assert_eq!(moved.holder, max_name);
    reservation::fence(&mut w, card, &max_name, 2).unwrap();
    let back = reservation::handoff(&mut w, card, &max_name, 2, "b").unwrap();
    assert_eq!(back.holder, "b");
}

// ---- AC3: generations and the fence -----------------------------------------------------------------

#[test]
fn handoff_ends_the_old_reservation_and_starts_the_next_generation_in_one_step() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    let first = reservation::reserve(
        &mut w,
        card,
        "a",
        &s(&["src/m/", "notes/m.txt"]),
        900,
        Some("/wt"),
        Some("br"),
    )
    .unwrap();
    let second = reservation::handoff(&mut w, card, "a", 1, "b").unwrap();
    assert_eq!(second.generation, 2);
    assert_eq!(second.holder, "b");
    assert_eq!(second.status, "held");
    assert_eq!(second.write_scope, first.write_scope);
    assert_eq!(
        (second.worktree.as_deref(), second.branch.as_deref()),
        (Some("/wt"), Some("br"))
    );
    assert_eq!(
        second.lease_until, first.lease_until,
        "the receiver gets the remaining lease"
    );
    let conn = reader(&home);
    let conn = conn.conn();
    assert_eq!(
        rows(conn, card),
        vec![
            (1, "a".to_string(), "released".to_string()),
            (2, "b".to_string(), "held".to_string())
        ]
    );
    assert_eq!(card_state(conn, card), ("active".to_string(), 2));
    assert_eq!(
        count(conn, "reservation_scope"),
        4,
        "two patterns per generation"
    );
    let ev = events(conn, "handoff");
    assert_eq!(ev.len(), 1);
    assert_eq!(
        (ev[0].2["from"].as_str(), ev[0].2["to"].as_str()),
        (Some("a"), Some("b"))
    );
    assert_eq!(ev[0].2["generation"], 2);
    reservation::fence(&mut w, card, "b", 2).unwrap();
}

#[test]
fn after_a_handoff_every_call_of_the_old_generation_is_refused_and_recorded() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();
    reservation::handoff(&mut w, card, "a", 1, "b").unwrap();
    let rows_before = rows(reader(&home).conn(), card);
    let reserve_events_before = count(reader(&home).conn(), "event");
    assert!(events(reader(&home).conn(), "fence_reject").is_empty());

    let stale = FenceError::StaleGeneration {
        offered: 1,
        current: 2,
    };
    assert_eq!(
        reservation::renew(&mut w, card, "a", 1, 600).unwrap_err(),
        stale
    );
    assert_eq!(
        reservation::release(&mut w, card, "a", 1).unwrap_err(),
        stale
    );
    assert_eq!(
        reservation::handoff(&mut w, card, "a", 1, "c").unwrap_err(),
        stale
    );
    assert_eq!(reservation::fence(&mut w, card, "a", 1).unwrap_err(), stale);
    // The new holder at the old generation is stale too: a lower generation never wins.
    assert_eq!(reservation::fence(&mut w, card, "b", 1).unwrap_err(), stale);
    assert_eq!(
        reservation::fence(&mut w, card, "b", 0).unwrap_err(),
        FenceError::StaleGeneration {
            offered: 0,
            current: 2
        }
    );
    assert_eq!(
        reservation::fence(&mut w, card, "b", -3)
            .unwrap_err()
            .code(),
        "stale_generation"
    );

    // The rejections are on disk, seen by a second connection, and nothing else moved.
    let r = reader(&home);
    let log = events(r.conn(), "fence_reject");
    assert_eq!(log.len(), 7);
    let ops: Vec<&str> = log.iter().map(|e| e.2["op"].as_str().unwrap()).collect();
    assert_eq!(
        ops,
        ["renew", "release", "handoff", "fence", "fence", "fence", "fence"]
    );
    for (card_id, actor, payload) in &log[..4] {
        assert_eq!(*card_id, Some(card));
        assert_eq!(actor, "a");
        assert_eq!(payload["offered_generation"], 1);
        assert_eq!(payload["current_generation"], 2);
        assert_eq!(payload["reason"], "stale_generation");
        assert_eq!(payload["card_id"], card);
    }
    assert_eq!(count(r.conn(), "event"), reserve_events_before + 7);
    assert_eq!(rows(r.conn(), card), rows_before);
    assert_eq!(card_state(r.conn(), card), ("active".to_string(), 2));
    // The current holder is unharmed and every key is distinct.
    reservation::renew(&mut w, card, "b", 2, 600).unwrap();
    let distinct: i64 = r
        .conn()
        .query_row("SELECT count(DISTINCT event_key) FROM event", [], |row| {
            row.get(0)
        })
        .unwrap();
    assert_eq!(distinct, count(r.conn(), "event"));
}

#[test]
fn every_other_reason_to_refuse_is_typed_and_recorded() {
    let (home, mut w) = setup();
    let never = ready_card(&mut w);
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();

    assert_eq!(
        reservation::fence(&mut w, 9999, "a", 1).unwrap_err(),
        FenceError::CardNotFound(9999)
    );
    assert_eq!(
        reservation::fence(&mut w, never, "a", 1).unwrap_err(),
        FenceError::NoReservation
    );
    assert_eq!(
        reservation::fence(&mut w, card, "a", 7).unwrap_err(),
        FenceError::UnknownGeneration {
            offered: 7,
            current: 1
        }
    );
    assert_eq!(
        reservation::fence(&mut w, card, "b", 1).unwrap_err(),
        FenceError::WrongHolder {
            held_by: "a".to_string()
        }
    );
    // A holder who is not the holder cannot do any of the operations.
    assert_eq!(
        reservation::renew(&mut w, card, "b", 1, 600)
            .unwrap_err()
            .code(),
        "wrong_holder"
    );
    assert_eq!(
        reservation::release(&mut w, card, "b", 1)
            .unwrap_err()
            .code(),
        "wrong_holder"
    );
    assert_eq!(
        reservation::handoff(&mut w, card, "b", 1, "c")
            .unwrap_err()
            .code(),
        "wrong_holder"
    );
    reservation::release(&mut w, card, "a", 1).unwrap();
    assert_eq!(
        reservation::fence(&mut w, card, "a", 1).unwrap_err(),
        FenceError::NotHeld {
            status: "released".to_string()
        }
    );
    assert_eq!(
        reservation::release(&mut w, card, "a", 1)
            .unwrap_err()
            .code(),
        "not_held"
    );
    assert_eq!(
        reservation::renew(&mut w, card, "a", 1, 600)
            .unwrap_err()
            .code(),
        "not_held"
    );

    let r = reader(&home);
    let log = events(r.conn(), "fence_reject");
    let reasons: Vec<&str> = log
        .iter()
        .map(|e| e.2["reason"].as_str().unwrap())
        .collect();
    assert_eq!(
        reasons,
        [
            "no_reservation",
            "unknown_generation",
            "wrong_holder",
            "wrong_holder",
            "wrong_holder",
            "wrong_holder",
            "not_held",
            "not_held",
            "not_held"
        ],
        "an unknown card has no row to attach an event to and is not recorded"
    );
    assert_eq!(log[0].0, Some(never));
    assert_eq!(log[0].2["current_generation"], 0);
    assert_eq!(log[1].2["current_generation"], 1);
    assert_eq!(log[1].2["offered_generation"], 7);
}

#[test]
fn an_ended_lease_is_refused_by_everything_except_a_release() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();
    // `lease_until = unixepoch()`: the lease-end second itself no longer holds.
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() WHERE card_id = 1",
    );
    for e in [
        reservation::fence(&mut w, card, "a", 1).unwrap_err(),
        reservation::renew(&mut w, card, "a", 1, 600)
            .map(|_| ())
            .unwrap_err(),
        reservation::handoff(&mut w, card, "a", 1, "b")
            .map(|_| ())
            .unwrap_err(),
    ] {
        assert!(matches!(e, FenceError::LeaseExpired { .. }), "{e:?}");
    }
    let r = reader(&home);
    assert_eq!(
        rows(r.conn(), card),
        vec![(1, "a".to_string(), "held".to_string())],
        "nothing took over"
    );
    let log = events(r.conn(), "fence_reject");
    assert_eq!(log.len(), 3);
    assert!(log.iter().all(|e| e.2["reason"] == "lease_expired"));
    // A live lease passes the same check.
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() + 300 WHERE card_id = 1",
    );
    reservation::fence(&mut w, card, "a", 1).unwrap();
    // The holder can still give a card back whose lease ended.
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() - 5 WHERE card_id = 1",
    );
    let released = reservation::release(&mut w, card, "a", 1).unwrap();
    assert_eq!(
        (
            released.reservation.status.as_str(),
            released.card_state.as_str()
        ),
        ("released", "ready")
    );
}

#[test]
fn renew_extends_from_the_db_clock_and_never_shortens() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    let first = reservation::reserve(&mut w, card, "a", &zone(card), 600, None, None).unwrap();
    // A shorter request than what is left changes nothing about the end of the lease.
    let same = reservation::renew(&mut w, card, "a", 1, MIN_LEASE_SECS).unwrap();
    assert_eq!(same.lease_until, first.lease_until);
    // A longer one counts from the DB clock now.
    sql(
        &mut w,
        "UPDATE reservation SET lease_until = unixepoch() + 10 WHERE card_id = 1",
    );
    let before = db_now(reader(&home).conn());
    let longer = reservation::renew(&mut w, card, "a", 1, 7200).unwrap();
    let after = db_now(reader(&home).conn());
    assert!(
        (before + 7200..=after + 7200).contains(&longer.lease_until),
        "{before} {after} {}",
        longer.lease_until
    );
    assert!(longer.updated_at >= before && longer.updated_at <= after);
    assert_eq!(
        (
            longer.generation,
            longer.holder.as_str(),
            longer.status.as_str()
        ),
        (1, "a", "held")
    );
    assert_eq!(events(reader(&home).conn(), "renew").len(), 2);
}

#[test]
fn release_gives_the_card_back_unless_it_moved_on() {
    let (home, mut w) = setup();
    let plain = ready_card(&mut w);
    take(&mut w, plain, "a", &zone(plain)).unwrap();
    let released = reservation::release(&mut w, plain, "a", 1).unwrap();
    assert_eq!(released.card_state, "ready");
    assert_eq!(
        card_state(reader(&home).conn(), plain),
        ("ready".to_string(), 1)
    );
    assert_eq!(
        reservation::current(reader(&home).conn(), plain).unwrap(),
        None
    );

    // `hold` stands for "moved on" here; `review` is the one state a release refuses (below).
    let moved = ready_card(&mut w);
    take(&mut w, moved, "a", &zone(moved)).unwrap();
    sql(
        &mut w,
        &format!("UPDATE card SET state = 'hold' WHERE card_id = {moved}"),
    );
    let released = reservation::release(&mut w, moved, "a", 1).unwrap();
    assert_eq!(
        released.card_state, "hold",
        "a card that moved on keeps its state"
    );
    assert_eq!(
        card_state(reader(&home).conn(), moved),
        ("hold".to_string(), 1)
    );
    assert_eq!(
        rows(reader(&home).conn(), moved),
        vec![(1, "a".to_string(), "released".to_string())]
    );
    let ev = events(reader(&home).conn(), "release");
    assert_eq!(ev.len(), 2);
    assert_eq!(ev[1].2["card_state"], "hold");
}

// ---- release and review (card C-68 N-a) -----------------------------------------------------------

fn event_count(home: &TempHome) -> i64 {
    reader(home)
        .conn()
        .query_row("SELECT count(*) FROM event", [], |r| r.get(0))
        .unwrap()
}

/// The invariant of C-68 N-a: a `review` card has a `held` reservation.
fn assert_review_is_held(home: &TempHome, card: i64, what: &str) {
    let r = reader(home);
    let (state, _) = card_state(r.conn(), card);
    if state == "review" {
        let current = reservation::current(r.conn(), card).unwrap();
        assert_eq!(
            current.as_ref().map(|c| c.status.as_str()),
            Some("held"),
            "{what}: C-{card} is `review` with no live reservation"
        );
    }
}

#[test]
fn release_refuses_a_review_card_and_writes_nothing() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();
    workflow::review(&mut w, card, "a", 1).unwrap();
    let before = event_count(&home);

    // Review first, release second: the holder is refused, with the way out in the sentence.
    let err = reservation::release(&mut w, card, "a", 1).unwrap_err();
    assert_eq!(
        err,
        FenceError::CardInReview {
            card_id: card,
            generation: 1
        }
    );
    assert_eq!(err.code(), "card_in_review");
    let text = err.to_string();
    assert!(text.starts_with("card conflict in state: "), "{text}");
    assert!(
        text.contains(&format!("card reopen C-{card} --generation 1")),
        "{text}"
    );
    assert_eq!(event_count(&home), before, "nothing was written");
    assert_eq!(
        card_state(reader(&home).conn(), card),
        ("review".to_string(), 1)
    );
    assert_eq!(
        rows(reader(&home).conn(), card),
        vec![(1, "a".to_string(), "held".to_string())]
    );

    // Someone who is not the holder gets the fence refusal and its event, not the state
    // sentence: the fence is checked first.
    assert!(matches!(
        reservation::release(&mut w, card, "b", 1).unwrap_err(),
        FenceError::WrongHolder { .. }
    ));
    assert!(matches!(
        reservation::release(&mut w, card, "a", 2).unwrap_err(),
        FenceError::UnknownGeneration { .. }
    ));
    let rejects = events(reader(&home).conn(), "fence_reject");
    assert_eq!(rejects.len(), 2);
    assert_eq!(rejects[0].2["op"], "release");
    assert_eq!(rejects[0].2["reason"], "wrong_holder");
    assert_eq!(rejects[1].2["reason"], "unknown_generation");
    assert_eq!(event_count(&home), before + 2);

    // A lease that ended does not open the way either: the card would still be stranded.
    sql(
        &mut w,
        &format!("UPDATE reservation SET lease_until = 1 WHERE card_id = {card}"),
    );
    assert!(matches!(
        reservation::release(&mut w, card, "a", 1).unwrap_err(),
        FenceError::CardInReview { .. }
    ));
    sql(
        &mut w,
        &format!("UPDATE reservation SET lease_until = unixepoch() + 3600 WHERE card_id = {card}"),
    );

    // The way out: reopen, then release.
    workflow::reopen(&mut w, card, "a", 1).unwrap();
    let released = reservation::release(&mut w, card, "a", 1).unwrap();
    assert_eq!(released.card_state, "ready");
    assert_eq!(
        rows(reader(&home).conn(), card),
        vec![(1, "a".to_string(), "released".to_string())]
    );
}

#[test]
fn review_after_a_release_is_a_fence_refusal() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();
    // Release first, review second: the card is `ready` and the review is refused.
    reservation::release(&mut w, card, "a", 1).unwrap();
    assert!(matches!(
        workflow::review(&mut w, card, "a", 1),
        Err(WorkflowError::Fence(FenceError::NotHeld { .. }))
    ));
    assert_eq!(
        card_state(reader(&home).conn(), card),
        ("ready".to_string(), 1)
    );
}

/// The holder runs `release` and `review` at the same moment, 100 times, each call on its own
/// connection. Whichever wins the write lock, exactly one of the two succeeds and a `review`
/// card always keeps its `held` reservation.
#[test]
fn a_hundred_release_and_review_races_never_strand_a_review_card() {
    const ROUNDS: usize = 100;
    let (home, mut w) = setup();
    let cards: Vec<i64> = (0..ROUNDS)
        .map(|_| {
            let card = ready_card(&mut w);
            take(&mut w, card, "a", &zone(card)).unwrap();
            card
        })
        .collect();
    drop(w);
    let (mut release_won, mut review_won) = (0, 0);
    for card in cards {
        let barrier = std::sync::Barrier::new(2);
        let (released, reviewed) = std::thread::scope(|s| {
            let releasing = s.spawn(|| {
                let mut w = ledger::open_write(home.path(), PROJECT).unwrap();
                barrier.wait();
                reservation::release(&mut w, card, "a", 1)
            });
            let reviewing = s.spawn(|| {
                let mut w = ledger::open_write(home.path(), PROJECT).unwrap();
                barrier.wait();
                workflow::review(&mut w, card, "a", 1)
            });
            (releasing.join().unwrap(), reviewing.join().unwrap())
        });
        assert_review_is_held(&home, card, "after the race");
        let (state, _) = card_state(reader(&home).conn(), card);
        match (&released, &reviewed) {
            (Ok(_), Err(WorkflowError::Fence(FenceError::NotHeld { .. }))) => {
                assert_eq!(state, "ready", "C-{card}");
                release_won += 1;
            }
            (Err(FenceError::CardInReview { .. }), Ok(_)) => {
                assert_eq!(state, "review", "C-{card}");
                review_won += 1;
            }
            other => panic!("C-{card}: exactly one of the two succeeds, got {other:?}"),
        }
    }
    assert_eq!(release_won + review_won, ROUNDS);
}

// ---- read-only side -----------------------------------------------------------------------------------

#[test]
fn the_read_only_calls_create_nothing_and_record_nothing() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    take(&mut w, card, "a", &zone(card)).unwrap();
    reservation::handoff(&mut w, card, "a", 1, "b").unwrap();
    drop(w);
    let before = home.snapshot();
    let r = reader(&home);
    reservation::fence_check(&r, card, "b", 2).unwrap();
    assert_eq!(
        reservation::fence_check(&r, card, "a", 1).unwrap_err(),
        FenceError::StaleGeneration {
            offered: 1,
            current: 2
        }
    );
    assert_eq!(
        reservation::fence_check(&r, 9999, "a", 1)
            .unwrap_err()
            .code(),
        "card_not_found"
    );
    assert_eq!(
        reservation::fence_check(&r, card, "", 2)
            .unwrap_err()
            .code(),
        "invalid_input"
    );
    assert_eq!(
        reservation::current(r.conn(), card)
            .unwrap()
            .unwrap()
            .holder,
        "b"
    );
    assert_eq!(reservation::history(r.conn(), card).unwrap().len(), 2);
    assert_eq!(reservation::history(r.conn(), 9999).unwrap(), vec![]);
    assert!(
        events(r.conn(), "fence_reject").is_empty(),
        "the read-only check records nothing"
    );
    drop(r);
    assert_eq!(home.snapshot(), before, "no file, no journal, no change");
}

// ---- processes ----------------------------------------------------------------------------------------

fn wait_for(path: &Path, what: &str) {
    let deadline = Instant::now() + Duration::from_secs(60);
    while !path.exists() {
        assert!(Instant::now() < deadline, "timed out waiting for {what}");
        std::thread::sleep(Duration::from_millis(1));
    }
}

fn spawn_self(
    test: &str,
    home: &TempHome,
    env: &[(&str, &Path)],
    extra: &[(&str, String)],
) -> std::process::Child {
    let mut cmd = std::process::Command::new(std::env::current_exe().unwrap());
    cmd.args(["--exact", test, "--nocapture", "--test-threads=1"])
        .clean_env()
        .env("HOME", home.path())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped());
    for (k, v) in env {
        cmd.env(k, v);
    }
    for (k, v) in extra {
        cmd.env(k, v);
    }
    cmd.spawn().unwrap()
}

const RACE_ENV: &str = "MOAI_RESERVE_RACE_DIR";
const KILL_ENV: &str = "MOAI_RESERVE_KILL_DIR";
const RACERS: usize = 8;
const ROUNDS: usize = 8;

/// The child half of the AC1 race: for every round, announce, wait for the shared start file,
/// then try to reserve that round's card. Does nothing unless the parent set `RACE_ENV`.
#[test]
fn reserve_race_child() {
    let Some(dir) = std::env::var_os(RACE_ENV).map(PathBuf::from) else {
        return;
    };
    let role = std::env::var("MOAI_RESERVE_RACE_ROLE").unwrap();
    let home = PathBuf::from(std::env::var_os("HOME").unwrap());
    let mut w = ledger::open_write(&home, PROJECT).unwrap();
    for round in 0..ROUNDS {
        fs::write(dir.join(format!("r{round}.{role}.ready")), "").unwrap();
        wait_for(&dir.join(format!("r{round}.go")), "the start file");
        let card = round as i64 + 1;
        let mut tries = 0;
        let outcome = loop {
            let r = reservation::reserve(
                &mut w,
                card,
                &format!("holder-{role}"),
                &[format!("round{round}/")],
                3600,
                None,
                None,
            );
            tries += 1;
            // `Busy` means the write lock was held past the busy timeout: ask again.
            if matches!(r, Err(ReserveError::Ledger(LedgerError::Busy))) && tries < 20 {
                continue;
            }
            break r;
        };
        let text = match outcome {
            Ok(r) => format!("ok {}", r.generation),
            Err(e) => e.code().to_string(),
        };
        fs::write(dir.join(format!("r{round}.{role}.result")), text).unwrap();
    }
}

#[test]
fn eight_processes_reserving_one_card_produce_exactly_one_holder_per_round() {
    let (home, mut w) = setup();
    for _ in 0..ROUNDS {
        ready_card(&mut w);
    }
    drop(w);
    let signals = tempfile::tempdir().unwrap();
    let mut children: Vec<(usize, std::process::Child)> = (0..RACERS)
        .map(|i| {
            let child = spawn_self(
                "reserve_race_child",
                &home,
                &[(RACE_ENV, signals.path())],
                &[("MOAI_RESERVE_RACE_ROLE", i.to_string())],
            );
            (i, child)
        })
        .collect();
    for round in 0..ROUNDS {
        for i in 0..RACERS {
            wait_for(
                &signals.path().join(format!("r{round}.{i}.ready")),
                "a racer to be ready",
            );
        }
        // The next round starts only when everyone finished this one.
        fs::write(signals.path().join(format!("r{round}.go")), "").unwrap();
        for i in 0..RACERS {
            wait_for(
                &signals.path().join(format!("r{round}.{i}.result")),
                "a racer's result",
            );
        }
    }
    for (i, child) in children.drain(..) {
        let out = child.wait_with_output().unwrap();
        assert!(
            out.status.success(),
            "racer {i} failed:\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
    }
    for round in 0..ROUNDS {
        let results: Vec<String> = (0..RACERS)
            .map(|i| {
                fs::read_to_string(signals.path().join(format!("r{round}.{i}.result"))).unwrap()
            })
            .collect();
        let winners = results.iter().filter(|r| r.as_str() == "ok 1").count();
        let losers = results
            .iter()
            .filter(|r| r.as_str() == "already_held")
            .count();
        assert_eq!(
            (winners, losers),
            (1, RACERS - 1),
            "round {round}: {results:?}"
        );
    }
    let r = reader(&home);
    for round in 0..ROUNDS {
        let card = round as i64 + 1;
        let held: Vec<_> = rows(r.conn(), card);
        assert_eq!(
            held.len(),
            1,
            "round {round}: exactly one reservation row, {held:?}"
        );
        assert_eq!((held[0].0, held[0].2.as_str()), (1, "held"));
        assert_eq!(card_state(r.conn(), card), ("active".to_string(), 1));
    }
    assert_eq!(count(r.conn(), "reservation"), ROUNDS as i64);
    assert_eq!(count(r.conn(), "reservation_scope"), ROUNDS as i64);
    assert_eq!(events(r.conn(), "reserve").len(), ROUNDS);
    assert_eq!(
        TempHome::names_in(
            ledger::ledger_path(home.path(), PROJECT)
                .unwrap()
                .parent()
                .unwrap()
        ),
        vec!["ledger.db"]
    );
}

/// The child of the kill test: opens its own write connection, starts the transaction a reserve
/// would start, writes what a reserve writes (plus enough filler, with a tiny page cache, that
/// the changes reach the database file before any commit), and then waits to be killed.
#[test]
fn reserve_kill_child() {
    let Some(dir) = std::env::var_os(KILL_ENV).map(PathBuf::from) else {
        return;
    };
    let home = PathBuf::from(std::env::var_os("HOME").unwrap());
    let mut w = ledger::open_write(&home, PROJECT).unwrap();
    let _: Result<(), LedgerError> = w.with_write_tx(|tx, now| {
        tx.execute_batch("PRAGMA cache_size = 10;")?;
        tx.execute(
            "UPDATE card SET generation = 1, state = 'active' WHERE card_id = 1",
            [],
        )?;
        tx.execute(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at) VALUES (1, 1, 'doomed', 'held', ?1, ?1 + 3600, ?1)",
            [now],
        )?;
        tx.execute(
            "INSERT INTO reservation_scope (reservation_id, pattern) VALUES (last_insert_rowid(), 'src/doomed/')",
            [],
        )?;
        tx.execute_batch(
            "WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < 4000) \
             INSERT INTO event (event_key, at, actor, kind, payload) \
             SELECT 'filler:' || i, 1, 'doomed', 'filler', json_object('pad', printf('%0400d', i)) FROM n;",
        )?;
        fs::write(dir.join("in_tx"), "").unwrap();
        loop {
            std::thread::sleep(Duration::from_millis(50));
        }
    });
}

#[test]
fn a_reserve_killed_mid_transaction_leaves_no_trace_and_its_generation_is_reused() {
    let (home, mut w) = setup();
    let card = ready_card(&mut w);
    assert_eq!(card, 1);
    drop(w);
    let signals = tempfile::tempdir().unwrap();
    let mut child = spawn_self(
        "reserve_kill_child",
        &home,
        &[(KILL_ENV, signals.path())],
        &[],
    );
    wait_for(
        &signals.path().join("in_tx"),
        "the child inside its transaction",
    );
    child.kill().unwrap();
    let status = child.wait().unwrap();
    assert!(!status.success(), "the child was killed, not finished");

    // The next writer rolls back whatever the dead one left.
    let mut w = ledger::open_write(home.path(), PROJECT).unwrap();
    let check: String = w
        .conn()
        .query_row("PRAGMA integrity_check", [], |r| r.get(0))
        .unwrap();
    assert_eq!(check, "ok");
    assert_eq!(
        count(w.conn(), "reservation"),
        0,
        "no held reservation, no row at all"
    );
    assert_eq!(count(w.conn(), "reservation_scope"), 0);
    assert_eq!(
        count(w.conn(), "event"),
        0,
        "the filler rolled back with it"
    );
    assert_eq!(card_state(w.conn(), card), ("ready".to_string(), 0));
    assert_eq!(reservation::current(w.conn(), card).unwrap(), None);

    let again = take(&mut w, card, "survivor", &s(&["src/doomed/"])).unwrap();
    assert_eq!(
        again.generation, 1,
        "the rolled-back generation is issued again"
    );
    assert_eq!(again.holder, "survivor");
    drop(w);
    let r = reader(&home);
    assert_eq!(
        rows(r.conn(), card),
        vec![(1, "survivor".to_string(), "held".to_string())]
    );
    let check: String = r
        .conn()
        .query_row("PRAGMA integrity_check", [], |row| row.get(0))
        .unwrap();
    assert_eq!(check, "ok");
    drop(r);
    assert_eq!(
        TempHome::names_in(
            ledger::ledger_path(home.path(), PROJECT)
                .unwrap()
                .parent()
                .unwrap()
        ),
        vec!["ledger.db"],
        "no leftover journal or temporary file"
    );
}
