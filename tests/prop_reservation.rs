//! Reservation state-machine property tests (C-100, SPEC-20261119-property-tests): under
//! random interleavings of the state machine in a temporary ledger, no card left in `review`
//! is without a held reservation, and no two live reservations cover one path (C-68 N-A/N-B).
//! The cases share one temporary HOME — each case works only on its own three cards and ages
//! every lease out at its end, so earlier cases can never stay `live` — and nothing touches
//! the real HOME.

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

/// The current generation, read through the world's one reader connection.
fn generation(world: &World, card: i64) -> i64 {
    world
        .reader
        .conn()
        .query_row(
            "SELECT generation FROM card WHERE card_id = ?1",
            [card],
            |r| r.get(0),
        )
        .unwrap()
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

/// REQ-7, REQ-8: the two invariants for one case's cards, read-only.
fn assert_invariants(world: &mut World, cards: &[i64]) {
    let conn = world.reader.conn();
    let now = db_now(&mut world.w);
    let mut live: Vec<reservation::Reservation> = Vec::new();
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
}

/// REQ-7: random interleavings; errors are legal refusals and are simply dropped. The world
/// is one temporary ledger for the whole property; each case takes three fresh cards and ages
/// every lease out at its end, so the next case starts overlap-free.
#[test]
fn random_interleavings_keep_the_two_invariants() {
    let mut world = setup_world();
    run_property("reservation interleave", MIN_CASES, |rng| {
        let cards = fresh_cards(&mut world.w);
        let scopes: Vec<Vec<String>> = cards
            .iter()
            .map(|_| vec![(*rng.pick(SCOPES)).to_string()])
            .collect();
        for _ in 0..12 {
            let at = rng.below(cards.len() as u64) as usize;
            let card = cards[at];
            let session = *rng.pick(SESSIONS);
            let scope = scopes[at].clone();
            let gen = generation(&world, card);
            match rng.below(9) {
                0 => {
                    let _ =
                        reservation::reserve(&mut world.w, card, session, &scope, 60, None, None);
                }
                1 => {
                    let _ = reservation::renew(&mut world.w, card, session, gen, 60);
                }
                2 => {
                    let _ = reservation::release(&mut world.w, card, session, gen);
                }
                3 => {
                    let _ = workflow::review(&mut world.w, card, session, gen);
                }
                4 => {
                    let _ = workflow::reopen(&mut world.w, card, session, gen);
                }
                5 => {
                    let _ = workflow::done(&mut world.w, card, session, gen);
                }
                6 => {
                    let to = *rng.pick(SESSIONS);
                    let _ = reservation::handoff(&mut world.w, card, session, gen, to);
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
                }
                _ => {
                    let _ = reservation::reclaim(
                        &mut world.w,
                        card,
                        session,
                        gen,
                        "prop",
                        "property case",
                        60,
                    );
                }
            }
        }
        assert_invariants(&mut world, &cards);
        // This case's reservations stop being live: the next case starts overlap-free.
        let _ = world.w.with_write_tx(|tx, _| {
            tx.execute(
                "UPDATE reservation SET lease_until = lease_until - 100_000",
                [],
            )?;
            Ok(())
        });
    });
}
