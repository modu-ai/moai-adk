//! The adopt transaction (SPEC-20261123-lane-continues-card): the same lane, after a `/clear`,
//! takes over its own card's reservation at the next generation.
//!
//! It is the reclaim transaction without the lease rule and without the reclaim allowance: the
//! reservation at `generation` must be the current one and `held`, and it must still be held by
//! the holder the caller's proofs were read for (the proofs themselves — the registry entry,
//! the worktree — are gathered by `card::adopt`, outside the ledger). The old reservation becomes
//! `reclaimed`, the new one is `held` with the same write scope, worktree and branch and a lease
//! from the ledger clock, and one `card.adopt` event keeps the proof. The card's state, notes,
//! criteria and links are not touched, and no `card.reclaim` event is written, so the card's
//! lifetime allowance of three reclaims stays whole.

use super::reservation::{
    add_event, load_reservation, ReclaimError, Reservation, MAX_HOLDER_LEN, MAX_LEASE_SECS,
    MIN_LEASE_SECS,
};
use super::{LedgerError, WriteLedger};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::json;

/// What the caller proved before the write, and what goes into the event.
pub struct AdoptProof<'a> {
    /// The holder session the proofs were read for; the reservation must still be held by it.
    pub holder: &'a str,
    /// The lane name of the registry entry of the new holder (the actor of the event).
    pub lane: &'a str,
    /// The canonical worktree the proofs were made for.
    pub worktree: &'a str,
    /// The `previous_sessions` of the new holder's entry: the proof that it continues the holder.
    pub previous_sessions: &'a [String],
}

/// What [`adopt`] did.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Adopted {
    /// The reservation that ended, as it is now (`status: reclaimed`).
    pub previous: Reservation,
    /// The new `held` reservation of the adopting session, one generation up.
    pub reservation: Reservation,
    /// The card's state, which an adoption does not change.
    pub card_state: String,
}

/// The state checks of a takeover, in the order `reclaim` makes them and with the answers it
/// gives: the card exists and is not `done`, it has a reservation, `generation` is the current
/// one, and that reservation is `held`. Reads only.
fn state_of(
    conn: &Connection,
    card_id: i64,
    generation: i64,
) -> Result<Result<(String, Reservation), ReclaimError>, LedgerError> {
    let row: Option<(String, i64)> = conn
        .query_row(
            "SELECT state, generation FROM card WHERE card_id = ?1",
            [card_id],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .optional()?;
    let Some((state, current)) = row else {
        return Ok(Err(ReclaimError::CardNotFound(card_id)));
    };
    if state == "done" {
        return Ok(Err(ReclaimError::CardDone(card_id)));
    }
    if current == 0 {
        return Ok(Err(ReclaimError::NoReservation));
    }
    if generation < current {
        return Ok(Err(ReclaimError::StaleGeneration {
            offered: generation,
            current,
        }));
    }
    if generation > current {
        return Ok(Err(ReclaimError::UnknownGeneration {
            offered: generation,
            current,
        }));
    }
    let Some(old) = load_reservation(conn, card_id, current)? else {
        return Ok(Err(ReclaimError::NoReservation));
    };
    if old.status != "held" {
        return Ok(Err(ReclaimError::NotHeld { status: old.status }));
    }
    Ok(Ok((state, old)))
}

/// The reservation `adopt` would take over, read on any connection: the answer `reclaim` gives
/// for the same state when the card cannot be taken over. Nothing is written.
pub fn peek(conn: &Connection, card_id: i64, generation: i64) -> Result<Reservation, ReclaimError> {
    Ok(state_of(conn, card_id, generation)??.1)
}

fn check_text(what: &str, value: &str, max: usize) -> Result<(), ReclaimError> {
    let n = value.chars().count();
    if n == 0 || n > max || value.chars().any(char::is_control) || value != value.trim() {
        return Err(ReclaimError::InvalidInput(format!(
            "{what} must be one line of 1 to {max} characters without a leading or trailing space"
        )));
    }
    Ok(())
}

/// Gives the reservation at `generation` to `to_holder` at generation + 1 (see the module text).
/// Rejected, with nothing written: bad input, an unknown or `done` card, a card with no
/// reservation, a generation that is not the current one, a reservation that is not `held`, and
/// a reservation now held by another session than `proof.holder`
/// ([`ReclaimError::HolderChanged`]).
pub fn adopt(
    w: &mut WriteLedger,
    card_id: i64,
    to_holder: &str,
    generation: i64,
    lease_secs: i64,
    proof: &AdoptProof<'_>,
) -> Result<Adopted, ReclaimError> {
    check_text("to_holder", to_holder, MAX_HOLDER_LEN)?;
    check_text("lane", proof.lane, 80)?;
    if !(MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&lease_secs) {
        return Err(ReclaimError::InvalidInput(format!(
            "lease_secs {lease_secs} is outside {MIN_LEASE_SECS}..={MAX_LEASE_SECS}"
        )));
    }
    w.with_write_tx(|tx, now| {
        let (state, old) = match state_of(tx, card_id, generation)? {
            Ok(found) => found,
            Err(refusal) => return Ok(Err(refusal)),
        };
        // The proofs were made for one holder; any other holder at the write voids them.
        if old.holder != proof.holder {
            return Ok(Err(ReclaimError::HolderChanged {
                proven: proof.holder.to_string(),
                current: old.holder,
            }));
        }
        tx.execute(
            "UPDATE reservation SET status = 'reclaimed', updated_at = ?1 WHERE reservation_id = ?2",
            params![now, old.reservation_id],
        )?;
        let next = generation + 1;
        tx.execute(
            "UPDATE card SET generation = ?1 WHERE card_id = ?2",
            params![next, card_id],
        )?;
        tx.execute(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at, worktree, branch) \
             VALUES (?1, ?2, ?3, 'held', ?4, ?5, ?4, ?6, ?7)",
            params![
                card_id,
                next,
                to_holder,
                now,
                now + lease_secs,
                old.worktree,
                old.branch
            ],
        )?;
        let new_id = tx.last_insert_rowid();
        tx.execute(
            "INSERT INTO reservation_scope (reservation_id, pattern) \
             SELECT ?1, pattern FROM reservation_scope WHERE reservation_id = ?2",
            params![new_id, old.reservation_id],
        )?;
        let payload = json!({
            "from_holder": old.holder,
            "from_generation": generation,
            "from_lease_until": old.lease_until,
            "holder": to_holder,
            "generation": next,
            "lease_until": now + lease_secs,
            "lane": proof.lane,
            "worktree": proof.worktree,
            "branch": old.branch,
            "previous_sessions": proof.previous_sessions,
            "card_state": state,
        });
        add_event(tx, now, proof.lane, card_id, "card.adopt", &payload)?;
        let previous = load_reservation(tx, card_id, generation)?
            .ok_or_else(|| LedgerError::Sql("the adopted reservation cannot be read".into()))?;
        let reservation = load_reservation(tx, card_id, next)?
            .ok_or_else(|| LedgerError::Sql("the new reservation cannot be read".into()))?;
        Ok(Ok(Adopted {
            previous,
            reservation,
            card_state: state,
        }))
    })?
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ledger::{self, reservation};

    /// A temporary ledger with one `ready` card taken by `holder`.
    fn taken(holder: &str) -> (tempfile::TempDir, WriteLedger, i64) {
        let home = tempfile::tempdir().unwrap();
        ledger::register(home.path(), "demo").unwrap();
        let mut w = ledger::open_write(home.path(), "demo").unwrap();
        let card = w
            .with_write_tx(|tx, now| {
                tx.execute(
                    "INSERT INTO card (title, state, created_by, created_at, updated_at) \
                     VALUES ('card', 'ready', 'test', ?1, ?1)",
                    [now],
                )?;
                Ok(tx.last_insert_rowid())
            })
            .unwrap();
        reservation::reserve(
            &mut w,
            card,
            holder,
            &["src/lane/".to_string()],
            3600,
            Some("/work/tree"),
            Some("lane-a"),
        )
        .unwrap();
        (home, w, card)
    }

    fn proof<'a>(holder: &'a str, previous: &'a [String]) -> AdoptProof<'a> {
        AdoptProof {
            holder,
            lane: "agent-1",
            worktree: "/work/tree",
            previous_sessions: previous,
        }
    }

    fn adopt_events(w: &WriteLedger) -> i64 {
        w.conn()
            .query_row(
                "SELECT COUNT(*) FROM event WHERE kind = 'card.adopt'",
                [],
                |r| r.get(0),
            )
            .unwrap()
    }

    /// AC-8 (REQ-9): a reservation that changed holder between the proof and the write is
    /// refused with `holder_changed` and the ledger shows no `card.adopt` event; the same call
    /// for the holder it was proven for goes through.
    #[test]
    fn ac8_a_holder_that_changed_after_the_proof_is_refused() {
        let (_home, mut w, card) = taken("aaaaaaaa-0000-4000-8000-000000000001");
        let previous = ["aaaaaaaa-0000-4000-8000-000000000001".to_string()];
        // The proofs were read while "…0001" held the card; another session holds the same
        // generation by the time of the write (a handoff would raise the generation and end in a
        // stale generation instead, so the row is changed directly).
        w.with_write_tx(|tx, _| {
            tx.execute(
                "UPDATE reservation SET holder = 'cccccccc-0000-4000-8000-000000000003' \
                 WHERE card_id = ?1 AND generation = 1",
                [card],
            )?;
            Ok(())
        })
        .unwrap();
        let err = adopt(
            &mut w,
            card,
            "bbbbbbbb-0000-4000-8000-000000000002",
            1,
            14_400,
            &proof("aaaaaaaa-0000-4000-8000-000000000001", &previous),
        )
        .unwrap_err();
        assert_eq!(err.code(), "holder_changed", "{err}");
        assert_eq!(adopt_events(&w), 0, "no event for a refusal");
        let now_held = reservation::current(w.conn(), card).unwrap().unwrap();
        assert_eq!(now_held.holder, "cccccccc-0000-4000-8000-000000000003");
        assert_eq!(now_held.generation, 1, "nothing moved");

        // The same call for the holder the proof names goes through.
        let got = adopt(
            &mut w,
            card,
            "bbbbbbbb-0000-4000-8000-000000000002",
            1,
            14_400,
            &proof("cccccccc-0000-4000-8000-000000000003", &previous),
        )
        .unwrap();
        assert_eq!(got.reservation.generation, 2);
        assert_eq!(got.reservation.write_scope, vec!["src/lane/".to_string()]);
        assert_eq!(got.previous.status, "reclaimed");
        assert_eq!(adopt_events(&w), 1);
    }

    /// REQ-8: the state answers are `reclaim`'s own, and `peek` reads them without a write.
    #[test]
    fn peek_answers_what_reclaim_answers() {
        let (_home, w, card) = taken("aaaaaaaa-0000-4000-8000-000000000001");
        assert_eq!(peek(w.conn(), card, 1).unwrap().generation, 1);
        assert_eq!(
            peek(w.conn(), card, 2).unwrap_err().code(),
            "unknown_generation"
        );
        assert_eq!(
            peek(w.conn(), card + 9, 1).unwrap_err().code(),
            "card_not_found"
        );
    }
}
