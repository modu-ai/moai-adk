//! Reservation state-machine property tests (C-100, SPEC-20261119-property-tests): under
//! random interleavings of the state machine in a temporary ledger, no card left in `review`
//! is without a held reservation, and no two live reservations cover one path (C-68 N-A/N-B).
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

/// Scope patterns drawn from a small alphabet, so collisions (and their refusals) happen:
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

/// REQ-7, REQ-8: the two invariants for one case's cards, read-only. Returns how many review
/// cards the first invariant judged and how many live pairs the second compared — the audit
/// of the first version (B1) showed a checker that never runs is green and worthless.
fn assert_invariants(world: &mut World, cards: &[i64]) -> (u32, u32) {
    let conn = world.reader.conn();
    let now = db_now(&mut world.w);
    let mut live: Vec<reservation::Reservation> = Vec::new();
    let mut review_checks = 0;
    for card in cards {
        let (state, _): (String, i64) = conn
            .query_row(
                "SELECT state, generation FROM card WHERE card_id = ?1",
                [card],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .unwrap();
        let current = reservation::current(conn, *card).unwrap();
        if state == "review" {
            review_checks += 1;
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
    let mut live_pairs = 0;
    for (i, a) in live.iter().enumerate() {
        for b in live.iter().skip(i + 1) {
            if a.card_id == b.card_id {
                continue;
            }
            live_pairs += 1;
            for m in &a.write_scope {
                for t in &b.write_scope {
                    assert!(
                        !scope::overlap(m, t),
                        "live reservations of cards {} and {} both cover a path ({m} vs {t})",
                        a.card_id,
                        b.card_id
                    );
                }
            }
        }
    }
    (review_checks, live_pairs)
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
        let (review_checks, live_pairs) = assert_invariants(&mut world, &cards);
        coverage.review_checks += review_checks;
        coverage.live_pairs += live_pairs;
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
    }
}
