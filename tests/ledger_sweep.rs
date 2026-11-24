//! A lease that ended long ago returns its card to `ready` (SPEC-20261124-ended-lease-sweep).
//! Every test runs in its own temporary HOME; none touches the real `~/.moai`. A lease ended
//! `n` hours ago is set with `lease_until = unixepoch() - n * 3600`; the one-second edge of the
//! grace period is pinned on the pure rule `past_grace`, so no test depends on the clock standing
//! still.

mod common;

use common::home::TempHome;
use moai_cli::ledger::reservation;
use moai_cli::ledger::sweep::{self, SkipReason, SweepError};
use moai_cli::ledger::{self, WriteLedger};
use serde_json::Value;
use std::fs;

const PROJECT: &str = "demo";
const A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const HOUR: i64 = 3600;
const GRACE: i64 = 24 * HOUR;

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

/// A card taken by `A` with a one-hour lease, a worktree and a branch.
fn taken(w: &mut WriteLedger, scope: &[&str]) -> i64 {
    let card = add_card(w, "[]");
    reservation::reserve(
        w,
        card,
        A,
        &paths(scope),
        HOUR,
        Some("/work/tree"),
        Some("card/x"),
    )
    .unwrap();
    card
}

fn sql(w: &mut WriteLedger, statement: &str) {
    w.with_write_tx(|tx, _| {
        tx.execute(statement, [])?;
        Ok(())
    })
    .unwrap();
}

/// The lease of the card's reservation ended `hours` hours ago.
fn end_lease(w: &mut WriteLedger, card: i64, hours: i64) {
    sql(
        w,
        &format!(
            "UPDATE reservation SET lease_until = unixepoch() - {} WHERE card_id = {card}",
            hours * HOUR
        ),
    );
}

fn state_of(w: &WriteLedger, card: i64) -> String {
    w.conn()
        .query_row("SELECT state FROM card WHERE card_id = ?1", [card], |r| {
            r.get(0)
        })
        .unwrap()
}

fn status_of(w: &WriteLedger, card: i64) -> String {
    w.conn()
        .query_row(
            "SELECT status FROM reservation WHERE card_id = ?1 ORDER BY generation DESC LIMIT 1",
            [card],
            |r| r.get(0),
        )
        .unwrap()
}

fn count(w: &WriteLedger, sql: &str) -> i64 {
    w.conn().query_row(sql, [], |r| r.get(0)).unwrap()
}

fn event_and_note_rows(w: &WriteLedger) -> i64 {
    count(w, "SELECT count(*) FROM event") + count(w, "SELECT count(*) FROM card_note")
}

fn events(w: &WriteLedger, kind: &str) -> Vec<Value> {
    let mut st = w
        .conn()
        .prepare("SELECT payload FROM event WHERE kind = ?1 ORDER BY event_id")
        .unwrap();
    st.query_map([kind], |r| r.get::<_, String>(0))
        .unwrap()
        .map(|p| serde_json::from_str(&p.unwrap()).unwrap())
        .collect()
}

fn leader() -> &'static str {
    ledger::workflow::LEADER_BY
}

// ---- AC-1: which cards are swept --------------------------------------------------------------

#[test]
fn sweep_returns_a_card_past_the_grace_to_ready_and_ends_its_reservation() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let out = sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    assert_eq!(out.swept.len(), 1, "{out:?}");
    assert_eq!(out.swept[0].card_id, card);
    assert_eq!(out.swept[0].holder, A);
    assert_eq!(out.swept[0].generation, 1);
    assert!(out.skipped.is_empty(), "{out:?}");
    assert_eq!(state_of(&w, card), "ready");
    assert_eq!(status_of(&w, card), "released");
    let sweeps = events(&w, "card.sweep");
    assert_eq!(sweeps.len(), 1, "one card.sweep event");
    assert_eq!(sweeps[0]["holder"], A);
    assert_eq!(sweeps[0]["generation"], 1);
    assert_eq!(sweeps[0]["by"], leader());
}

#[test]
fn sweep_skips_a_card_inside_the_grace_period_and_changes_nothing() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 23);
    let events_before = count(&w, "SELECT count(*) FROM event");
    let out = sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    assert!(out.swept.is_empty(), "{out:?}");
    assert_eq!(out.skipped.len(), 1, "{out:?}");
    assert_eq!(out.skipped[0].card_id, card);
    assert!(
        matches!(out.skipped[0].reason, SkipReason::WithinGrace { .. }),
        "{out:?}"
    );
    assert_eq!(state_of(&w, card), "active");
    assert_eq!(status_of(&w, card), "held");
    assert_eq!(count(&w, "SELECT count(*) FROM event"), events_before);
}

#[test]
fn sweep_skips_a_live_lease_without_listing_it_as_past_grace() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    let out = sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    assert!(out.swept.is_empty(), "{out:?}");
    assert_eq!(state_of(&w, card), "active");
    assert_eq!(status_of(&w, card), "held");
}

#[test]
fn sweep_skips_a_card_in_review() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    sql(
        &mut w,
        &format!("UPDATE card SET state = 'review' WHERE card_id = {card}"),
    );
    end_lease(&mut w, card, 30);
    let out = sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    assert!(out.swept.is_empty(), "{out:?}");
    assert!(
        matches!(out.skipped[0].reason, SkipReason::InReview),
        "{out:?}"
    );
    assert_eq!(state_of(&w, card), "review");
    assert_eq!(status_of(&w, card), "held");
}

#[test]
fn sweep_skips_a_recovered_reservation_and_an_expired_candidate() {
    let (_home, mut w) = setup();
    let recovered = taken(&mut w, &["src/a.rs"]);
    let candidate = taken(&mut w, &["src/b.rs"]);
    sql(
        &mut w,
        &format!("UPDATE reservation SET status = 'held_for_review' WHERE card_id = {recovered}"),
    );
    sql(
        &mut w,
        &format!("UPDATE reservation SET status = 'expired_candidate' WHERE card_id = {candidate}"),
    );
    end_lease(&mut w, recovered, 30);
    end_lease(&mut w, candidate, 30);
    let out = sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    assert!(out.swept.is_empty(), "{out:?}");
    assert_eq!(out.skipped.len(), 2, "{out:?}");
    for skipped in &out.skipped {
        assert!(
            matches!(skipped.reason, SkipReason::NotHeld { .. }),
            "{out:?}"
        );
    }
    assert_eq!(status_of(&w, recovered), "held_for_review");
    assert_eq!(status_of(&w, candidate), "expired_candidate");
    assert_eq!(state_of(&w, recovered), "active");
}

#[test]
fn sweep_skips_a_lease_renewed_after_the_command_read_it() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let read: i64 = w
        .conn()
        .query_row(
            "SELECT lease_until FROM reservation WHERE card_id = ?1",
            [card],
            |r| r.get(0),
        )
        .unwrap();
    // The holder renews between the read and the write.
    reservation::renew(&mut w, card, A, 1, HOUR).unwrap();
    let out = sweep::sweep_card(&mut w, card, Some(read), GRACE, leader(), None).unwrap();
    assert!(
        matches!(out, Err(SkipReason::Changed)),
        "a renewed lease is skipped: {out:?}"
    );
    assert_eq!(state_of(&w, card), "active");
    assert_eq!(status_of(&w, card), "held");
    assert!(events(&w, "card.sweep").is_empty());
}

/// The lease end a caller listed, read the way `sweep_all` reads it.
fn lease_until_of(w: &WriteLedger, card: i64) -> i64 {
    w.conn()
        .query_row(
            "SELECT lease_until FROM reservation WHERE card_id = ?1 ORDER BY generation DESC LIMIT 1",
            [card],
            |r| r.get(0),
        )
        .unwrap()
}

#[test]
fn sweep_card_skips_a_card_that_went_to_review_after_the_listing() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let read = lease_until_of(&w, card);
    sql(
        &mut w,
        &format!("UPDATE card SET state = 'review' WHERE card_id = {card}"),
    );
    let out = sweep::sweep_card(&mut w, card, Some(read), GRACE, leader(), None).unwrap();
    assert!(matches!(out, Err(SkipReason::InReview)), "{out:?}");
    assert_eq!(state_of(&w, card), "review");
    assert_eq!(status_of(&w, card), "held");
    assert!(events(&w, "card.sweep").is_empty());
}

#[test]
fn sweep_card_skips_a_reservation_that_is_no_longer_held_after_the_listing() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let read = lease_until_of(&w, card);
    sql(
        &mut w,
        &format!("UPDATE reservation SET status = 'held_for_review' WHERE card_id = {card}"),
    );
    let out = sweep::sweep_card(&mut w, card, Some(read), GRACE, leader(), None).unwrap();
    assert!(
        matches!(out, Err(SkipReason::NotHeld { ref status }) if status == "held_for_review"),
        "{out:?}"
    );
    assert_eq!(state_of(&w, card), "active");
    assert_eq!(status_of(&w, card), "held_for_review");
    assert!(events(&w, "card.sweep").is_empty());
}

#[test]
fn sweep_card_skips_a_card_inside_the_grace_period_whatever_the_caller_listed() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 23);
    let read = lease_until_of(&w, card);
    let out = sweep::sweep_card(&mut w, card, Some(read), GRACE, leader(), None).unwrap();
    assert!(
        matches!(out, Err(SkipReason::WithinGrace { lease_until, sweepable_at })
            if lease_until == read && sweepable_at == read + GRACE),
        "{out:?}"
    );
    assert_eq!(state_of(&w, card), "active");
    assert_eq!(status_of(&w, card), "held");
    assert!(events(&w, "card.sweep").is_empty());
}

#[test]
fn sweep_card_skips_a_card_that_is_not_active_and_one_that_does_not_exist() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let read = lease_until_of(&w, card);
    reservation::release(&mut w, card, A, 1).unwrap();
    assert_eq!(state_of(&w, card), "ready");
    let out = sweep::sweep_card(&mut w, card, Some(read), GRACE, leader(), None).unwrap();
    assert!(matches!(out, Err(SkipReason::Changed)), "{out:?}");
    let out = sweep::sweep_card(&mut w, 9_999, None, GRACE, leader(), None).unwrap();
    assert!(matches!(out, Err(SkipReason::Changed)), "{out:?}");
    assert!(events(&w, "card.sweep").is_empty());
}

#[test]
fn sweep_dry_run_lists_the_cards_and_writes_nothing() {
    let (_home, mut w) = setup();
    let ended = taken(&mut w, &["src/a.rs"]);
    let inside = taken(&mut w, &["src/b.rs"]);
    end_lease(&mut w, ended, 25);
    end_lease(&mut w, inside, 23);
    let rows_before = event_and_note_rows(&w);
    let out = sweep::sweep_all(&mut w, GRACE, leader(), None, true).unwrap();
    assert!(out.dry_run);
    assert_eq!(out.swept.len(), 1, "{out:?}");
    assert_eq!(out.swept[0].card_id, ended);
    assert_eq!(out.skipped.len(), 1, "{out:?}");
    assert_eq!(state_of(&w, ended), "active");
    assert_eq!(status_of(&w, ended), "held");
    assert_eq!(event_and_note_rows(&w), rows_before);
}

#[test]
fn sweep_is_the_leaders_and_a_stranger_writes_nothing() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let err = sweep::sweep_all(&mut w, GRACE, "someone", None, false).unwrap_err();
    assert!(matches!(err, SweepError::LeaderOnly), "{err:?}");
    assert_eq!(state_of(&w, card), "active");
    assert!(events(&w, "card.sweep").is_empty());
}

#[test]
fn sweep_the_grace_edge_is_the_second_the_grace_has_run_out() {
    let lease_until = 1_000_000;
    assert!(!sweep::past_grace(
        lease_until,
        GRACE,
        lease_until + GRACE - 1
    ));
    assert!(sweep::past_grace(lease_until, GRACE, lease_until + GRACE));
    assert!(sweep::past_grace(
        lease_until,
        GRACE,
        lease_until + GRACE + 1
    ));
    assert!(
        !sweep::past_grace(i64::MAX, GRACE, i64::MAX - 1),
        "a huge lease never overflows"
    );
}

// ---- AC-2: what a sweep keeps ------------------------------------------------------------------

#[test]
fn kept_the_generation_branch_worktree_notes_and_criteria_stay() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    ledger::workflow::add_note(
        &mut w,
        card,
        "report",
        "lane",
        None,
        "RESUME C-1 gen=1",
        None,
    )
    .unwrap();
    end_lease(&mut w, card, 25);
    let notes_before = count(&w, "SELECT count(*) FROM card_note");
    let criteria_before = count(&w, "SELECT count(*) FROM card_ac");
    sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    let generation: i64 = w
        .conn()
        .query_row(
            "SELECT generation FROM card WHERE card_id = ?1",
            [card],
            |r| r.get(0),
        )
        .unwrap();
    assert_eq!(generation, 1, "the generation is not advanced");
    let (worktree, branch): (String, String) = w
        .conn()
        .query_row(
            "SELECT worktree, branch FROM reservation WHERE card_id = ?1",
            [card],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .unwrap();
    assert_eq!(
        (worktree.as_str(), branch.as_str()),
        ("/work/tree", "card/x")
    );
    assert_eq!(
        count(&w, "SELECT count(*) FROM card_note"),
        notes_before + 1,
        "one note is added and none is changed"
    );
    assert_eq!(count(&w, "SELECT count(*) FROM card_ac"), criteria_before);
    let kept: String = w
        .conn()
        .query_row(
            "SELECT body FROM card_note WHERE card_id = ?1 AND body LIKE 'RESUME%'",
            [card],
            |r| r.get(0),
        )
        .unwrap();
    assert_eq!(kept, "RESUME C-1 gen=1");
}

#[test]
fn kept_the_note_names_the_holder_generation_lease_end_worktree_and_branch() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    end_lease(&mut w, card, 25);
    let lease_until: i64 = w
        .conn()
        .query_row(
            "SELECT lease_until FROM reservation WHERE card_id = ?1",
            [card],
            |r| r.get(0),
        )
        .unwrap();
    sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    let body: String = w
        .conn()
        .query_row(
            "SELECT body FROM card_note WHERE card_id = ?1 ORDER BY note_id DESC LIMIT 1",
            [card],
            |r| r.get(0),
        )
        .unwrap();
    for fact in [
        A.to_string(),
        "generation 1".to_string(),
        lease_until.to_string(),
        "/work/tree".to_string(),
        "card/x".to_string(),
        "Read the notes".to_string(),
    ] {
        assert!(body.contains(&fact), "the note lacks `{fact}`: {body}");
    }
}

#[test]
fn kept_the_reclaim_allowance_is_not_used() {
    let (_home, mut w) = setup();
    let card = taken(&mut w, &["src/a.rs"]);
    for _ in 0..2 {
        sql(
            &mut w,
            &format!(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES ('test-' || lower(hex(randomblob(8))), unixepoch(), 'test', {card}, \
                 'card.reclaim', '{{}}')"
            ),
        );
    }
    end_lease(&mut w, card, 25);
    sweep::sweep_all(&mut w, GRACE, leader(), None, false).unwrap();
    assert_eq!(
        count(
            &w,
            &format!("SELECT count(*) FROM event WHERE card_id = {card} AND kind = 'card.reclaim'")
        ),
        2,
        "a sweep writes no card.reclaim event, so the allowance of three stays at one more"
    );
    // The card is taken again at the next generation, and a reclaim of that reservation is
    // still allowed once its lease ended.
    reservation::reserve(&mut w, card, A, &paths(&["src/a.rs"]), HOUR, None, None).unwrap();
    end_lease(&mut w, card, 1);
    reservation::reclaim(&mut w, card, A, 2, "leader", "after a sweep", HOUR).unwrap();
}

// ---- AC-3: the grace period of the profile -----------------------------------------------------

fn write_profile(project: &std::path::Path, extra: &str) {
    let cfg = moai_cli::state_root::path(project, "project.yaml");
    fs::create_dir_all(cfg.parent().unwrap()).unwrap();
    fs::write(
        &cfg,
        format!(
            "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]{extra}}}"
        ),
    )
    .unwrap();
}

#[test]
fn grace_comes_from_the_profile_key_and_defaults_to_twenty_four_hours() {
    let home = TempHome::new();
    let project = home.project("demo");
    write_profile(&project, "");
    assert_eq!(
        moai_cli::card::sweep::grace_secs(&project).unwrap(),
        24 * HOUR
    );
    write_profile(&project, ", \"lease_sweep_hours\": 48");
    assert_eq!(
        moai_cli::card::sweep::grace_secs(&project).unwrap(),
        48 * HOUR
    );
}

#[test]
fn grace_a_profile_with_zero_hours_is_refused() {
    let home = TempHome::new();
    let project = home.project("demo");
    write_profile(&project, ", \"lease_sweep_hours\": 0");
    assert!(
        moai_cli::profile::read_profile(&project).is_err(),
        "lease_sweep_hours: 0 must not read"
    );
    write_profile(&project, ", \"lease_sweep_hours\": 1");
    assert!(moai_cli::profile::read_profile(&project).is_ok());
}

#[test]
fn grace_sweeps_at_forty_nine_hours_and_skips_at_forty_seven_for_a_forty_eight_hour_key() {
    let (_home, mut w) = setup();
    let late = taken(&mut w, &["src/a.rs"]);
    let early = taken(&mut w, &["src/b.rs"]);
    end_lease(&mut w, late, 49);
    end_lease(&mut w, early, 47);
    let out = sweep::sweep_all(&mut w, 48 * HOUR, leader(), None, false).unwrap();
    assert_eq!(out.swept.len(), 1, "{out:?}");
    assert_eq!(out.swept[0].card_id, late);
    assert_eq!(out.skipped[0].card_id, early);
}
