//! Reservation state-machine property tests (SPEC-20261119-property-tests): under
//! random interleavings of the state machine in a temporary ledger, no card left in `review`
//! is without a held reservation, a card has at most one live reservation, and a path that two
//! cards' live reservations can both write is reported as an overlap, never refused
//! (SPEC-20261125-card-no-file-locks REQ-1: "the card ledger shall make the reservation and
//! answer each overlap as `overlaps`").
//! The cases share one temporary HOME — each case works only on its own three cards and, at
//! its end, walks them back through the system's own transitions (`reopen` then `release`,
//! an ended lease through `reclaim`), so earlier cases never block a later take — and nothing
//! touches the real HOME.

#[path = "common/prop.rs"]
mod prop;

mod common;

use common::home::TempHome;
use moai_cli::ledger::{self, WriteLedger};
use moai_cli::ledger::{reservation, workflow};
use moai_cli::scope;
use prop::{run_property, MIN_CASES};
use rusqlite::params;

const PROJECT: &str = "prop";

/// Scope patterns drawn from a small alphabet, so collisions (and their overlap reports) happen:
/// `zone-a/**` overlaps `zone-a/sub/**`, and two cards may draw the same zone.
const SCOPES: &[&str] = &[
    "zone-a/**",
    "zone-a/sub/**",
    "zone-b/**",
    "shared/**",
    "src/**",
    "docs/*.md",
];

const SESSIONS: &[&str] = &["s0", "s1", "s2"];

struct World {
    /// Kept alive (never read again): the temporary HOME the ledger lives in.
    _home: TempHome,
    w: WriteLedger,
    reader: ledger::ReadLedger,
}

fn setup_world() -> World {
    let home = TempHome::new();
    ledger::register(home.path(), PROJECT).unwrap();
    let w = ledger::open_write(home.path(), PROJECT).unwrap();
    let reader = ledger::open_read_only(home.path(), PROJECT).unwrap();
    World {
        _home: home,
        w,
        reader,
    }
}

fn db_now(w: &mut WriteLedger) -> i64 {
    w.with_write_tx(|_, now| Ok(now)).unwrap()
}

/// Three fresh `ready` cards for one case, in one write transaction.
fn fresh_cards(w: &mut WriteLedger) -> Vec<i64> {
    w.with_write_tx(|tx, now| {
        let mut ids = Vec::new();
        for _ in 0..3 {
            tx.execute(
                "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                 VALUES ('prop', 'ready', 'test', ?1, ?1)",
                params![now],
            )?;
            ids.push(tx.last_insert_rowid());
        }
        Ok(ids)
    })
    .unwrap()
}

/// What one case's invariant check judged: review cards, live pairs of cards, and live pairs
/// whose patterns overlap (each of those must be reported by `reservation::overlaps`).
#[derive(Default, Clone, Copy)]
struct Judged {
    review_checks: u32,
    live_pairs: u32,
    overlap_pairs: u32,
}

/// REQ-7, REQ-8: the invariants for one case's cards, read-only: a review card keeps a held
/// reservation; a card has at most one live reservation (SPEC-20261125-card-no-file-locks REQ-6
/// keeps "a card is held by one reservation" while REQ-1 and REQ-2 allow overlapping scopes);
/// and for every pair of live reservations of two cards whose patterns can both write a path,
/// `reservation::overlaps` names the other card, both patterns and its holder (the `overlaps`
/// warning of a take). Returns what it judged — the audit of the first version (B1) showed a
/// checker that never runs is green and worthless.
fn assert_invariants(world: &mut World, cards: &[i64]) -> Judged {
    let conn = world.reader.conn();
    let now = db_now(&mut world.w);
    let mut live: Vec<reservation::Reservation> = Vec::new();
    let mut judged = Judged::default();
    for card in cards {
        let active: i64 = conn
            .query_row(
                "SELECT COUNT(*) FROM reservation WHERE card_id = ?1 \
                 AND status IN ('held', 'expired_candidate', 'held_for_review')",
                [card],
                |r| r.get(0),
            )
            .unwrap();
        assert!(
            active <= 1,
            "card {card} holds {active} active reservations"
        );
        let (state, _): (String, i64) = conn
            .query_row(
                "SELECT state, generation FROM card WHERE card_id = ?1",
                [card],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .unwrap();
        let current = reservation::current(conn, *card).unwrap();
        if state == "review" {
            judged.review_checks += 1;
            let _held = current
                .as_ref()
                .filter(|r| r.status == "held" || r.status == "held_for_review")
                .unwrap_or_else(|| {
                    panic!("card {card} is left in review without a held reservation")
                });
        }
        if let Some(r) = current.filter(|r| r.lease_until > now) {
            live.push(r);
        }
    }
    for (i, a) in live.iter().enumerate() {
        for b in live.iter().skip(i + 1) {
            if a.card_id == b.card_id {
                continue;
            }
            judged.live_pairs += 1;
            // What a take of `a`'s card would be told about `b`'s card, asked now.
            let told = reservation::overlaps(conn, now, a.card_id, &a.write_scope).unwrap();
            for m in &a.write_scope {
                for t in &b.write_scope {
                    if !scope::card_overlap(m, t) {
                        continue;
                    }
                    judged.overlap_pairs += 1;
                    assert!(
                        told.iter().any(|o| o.other_card == b.card_id
                            && &o.mine == m
                            && &o.theirs == t
                            && o.holder == b.holder),
                        "live reservations of cards {} and {} both cover a path ({m} vs {t}) \
                         and the overlap is not reported: {told:?}",
                        a.card_id,
                        b.card_id
                    );
                }
            }
        }
    }
    judged
}

/// How many of each transition the run actually achieved — the audit of the first version
/// (B1) showed a driver can be green while exercising almost nothing, so the counts are
/// asserted against floors the run must clear.
#[derive(Default, Debug)]
struct Coverage {
    reserve: u32,
    renew: u32,
    release: u32,
    review: u32,
    reopen: u32,
    handoff: u32,
    reclaim: u32,
    aged: u32,
    review_checks: u32,
    live_pairs: u32,
    overlap_pairs: u32,
}

/// REQ-7: random interleavings; errors are legal refusals and are simply dropped. The world
/// is one temporary ledger for the whole property; each case takes three fresh cards and, at
/// its end, walks them back through the system's own transitions (`reopen` then `release`,
/// an ended lease through `reclaim`), so the next case starts overlap-free without the
/// harness manufacturing a state the state machine itself cannot reach.
#[test]
fn random_interleavings_keep_the_two_invariants() {
    let mut world = setup_world();
    let mut coverage = Coverage::default();
    run_property("reservation interleave", MIN_CASES, |rng| {
        let cards = fresh_cards(&mut world.w);
        let scopes: Vec<Vec<String>> = cards
            .iter()
            .map(|_| vec![(*rng.pick(SCOPES)).to_string()])
            .collect();
        for _ in 0..10 {
            let at = rng.below(cards.len() as u64) as usize;
            let card = cards[at];
            let session = *rng.pick(SESSIONS);
            let scope = scopes[at].clone();
            // The fenced calls answer to the CURRENT reservation's holder and generation
            // (the card table can lag a handoff), the way a lane that holds the card knows them.
            let cur = reservation::current(world.reader.conn(), card).unwrap();
            let (holder, gen) = match &cur {
                Some(r) => (r.holder.as_str(), r.generation),
                None => ("", 0),
            };
            match rng.below(9) {
                0 => {
                    if reservation::reserve(&mut world.w, card, session, &scope, 60, None, None)
                        .is_ok()
                    {
                        coverage.reserve += 1;
                    }
                }
                1 => {
                    if reservation::renew(&mut world.w, card, holder, gen, 60).is_ok() {
                        coverage.renew += 1;
                    }
                }
                2 => {
                    if reservation::release(&mut world.w, card, holder, gen).is_ok() {
                        coverage.release += 1;
                    }
                }
                3 => {
                    if workflow::review(&mut world.w, card, holder, gen).is_ok() {
                        coverage.review += 1;
                    }
                }
                4 => {
                    if workflow::reopen(&mut world.w, card, holder, gen).is_ok() {
                        coverage.reopen += 1;
                    }
                }
                5 => {
                    let _ = workflow::done(&mut world.w, card, holder, gen, None);
                }
                6 => {
                    let to = *rng.pick(SESSIONS);
                    if reservation::handoff(&mut world.w, card, holder, gen, to).is_ok() {
                        coverage.handoff += 1;
                    }
                }
                7 => {
                    // Time passes: the lease ages out (the reclaim path's precondition).
                    let _ = world.w.with_write_tx(|tx, _| {
                        tx.execute(
                            "UPDATE reservation SET lease_until = lease_until - 600 \
                             WHERE card_id = ?1",
                            params![card],
                        )?;
                        Ok(())
                    });
                    coverage.aged += 1;
                }
                _ => {
                    if reservation::reclaim(
                        &mut world.w,
                        card,
                        session,
                        gen,
                        "prop",
                        "property case",
                        60,
                    )
                    .is_ok()
                    {
                        coverage.reclaim += 1;
                    }
                }
            }
        }
        let judged = assert_invariants(&mut world, &cards);
        coverage.review_checks += judged.review_checks;
        coverage.live_pairs += judged.live_pairs;
        coverage.overlap_pairs += judged.overlap_pairs;
        // This case's cards go back to `ready` through the system's own transitions, so the
        // next case's takes are not blocked by rows this case left behind. The holder and the
        // generation come from the reservation itself; a review card whose lease has ended is
        // reclaimed first (the system's own way back for it), then released.
        for card in &cards {
            for _ in 0..3 {
                let Some(r) = reservation::current(world.reader.conn(), *card).unwrap() else {
                    break;
                };
                let _ = workflow::reopen(&mut world.w, *card, &r.holder, r.generation);
                if reservation::release(&mut world.w, *card, &r.holder, r.generation).is_ok() {
                    break;
                }
                let reclaimer = *rng.pick(SESSIONS);
                if reservation::reclaim(
                    &mut world.w,
                    *card,
                    reclaimer,
                    r.generation,
                    "prop",
                    "case cleanup",
                    60,
                )
                .is_err()
                {
                    break;
                }
            }
        }
    });
    // The driver must actually exercise the state machine (audit B1): the floors sit far
    // below the observed counts of a healthy run. A `PROP_ONLY` replay runs one case on
    // purpose, so the floors stand down for it (re-audit N7).
    if std::env::var("PROP_ONLY").is_err() {
        assert!(coverage.reserve >= 500, "{coverage:?}");
        assert!(coverage.renew >= 50, "{coverage:?}");
        assert!(coverage.release >= 50, "{coverage:?}");
        assert!(coverage.review >= 50, "{coverage:?}");
        assert!(coverage.reopen >= 5, "{coverage:?}");
        assert!(coverage.handoff >= 20, "{coverage:?}");
        assert!(coverage.reclaim >= 5, "{coverage:?}");
        assert!(coverage.aged >= 500, "{coverage:?}");
        assert!(coverage.review_checks >= 50, "{coverage:?}");
        assert!(coverage.live_pairs >= 100, "{coverage:?}");
        assert!(coverage.overlap_pairs >= 20, "{coverage:?}");
    }
}
