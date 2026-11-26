//! The offer and accept transactions (SPEC-20261126-card-offer-accept): the holder offers a card
//! to one named session, nothing moves until that session accepts, and every situation in which
//! the offer is no longer good is refused with nothing written.
//!
//! An offer is one `card.offer` event and moves no reservation row. A cancel is one
//! `card.offer_cancel` event. An accept is the handoff's reservation rows plus one `card.accept`
//! event, in one transaction that re-reads the state. The newest event of the offer family
//! (`card.offer` or `card.offer_cancel`) of the card is the only one that counts, so a new offer
//! replaces the older one and a cancel voids the newest. The ledger schema is not touched: the
//! kinds are free strings of `event.kind`. No `card.reclaim` event is written and the allowance
//! of three reclaims is neither read nor used.

use super::adopt;
use super::reservation::{
    add_event, fenced, load_reservation, FenceError, ReclaimError, Reservation, MAX_LEASE_SECS,
    MIN_LEASE_SECS,
};
use super::{LedgerError, WriteLedger};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};

/// The event kind of an offer.
pub const KIND_OFFER: &str = "card.offer";
/// The event kind of a cancelled offer.
pub const KIND_CANCEL: &str = "card.offer_cancel";
/// The event kind of an accepted offer.
pub const KIND_ACCEPT: &str = "card.accept";

/// Seconds an offer stays valid when `--valid-secs` is not given (SPEC decision D-1).
pub const DEFAULT_VALID_SECS: i64 = 1_800;

/// What [`offer`] wrote.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Offered {
    pub generation: i64,
    pub to: String,
    pub valid_secs: i64,
    pub valid_until: i64,
    pub handover: Option<String>,
}

/// What [`cancel`] voided.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Cancelled {
    pub generation: i64,
    pub to: String,
}

/// What [`accept`] did.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Accepted {
    /// The reservation that ended, as it is now (`status: released`).
    pub previous: Reservation,
    /// The new `held` reservation of the accepting session, one generation up.
    pub reservation: Reservation,
    /// The card's state, which an accept does not change.
    pub card_state: String,
    /// The `event_id` of the offer that was accepted.
    pub offer_event: i64,
}

/// Why an accept or a cancel was refused. Every refusal writes nothing.
#[derive(Debug)]
pub enum OfferError {
    /// The card has no offer at this generation (or nothing open to cancel).
    NoOffer { generation: i64 },
    /// The newest offer of this generation was cancelled.
    OfferCancelled,
    /// The accepting session is not the one the newest offer names.
    NotTarget,
    /// The offer's validity ended at `valid_until` (ledger clock `now`).
    OfferExpired { valid_until: i64, now: i64 },
    /// The holder's lease ended at `lease_until` (ledger clock `now`): the reclaim road is open.
    LeaseExpired { lease_until: i64, now: i64 },
    /// The state of the card or reservation, in the words `card reclaim` uses for it.
    State(ReclaimError),
}

impl std::fmt::Display for OfferError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            OfferError::NoOffer { generation } => write!(
                f,
                "card conflict in no_offer: the card has no open offer at generation {generation}"
            ),
            OfferError::OfferCancelled => f.write_str(
                "card conflict in offer_cancelled: the holder cancelled the newest offer",
            ),
            OfferError::NotTarget => f.write_str(
                "card conflict in not_target: this session is not the session the offer names",
            ),
            OfferError::OfferExpired { valid_until, now } => write!(
                f,
                "card conflict in offer_expired: the offer was valid until {valid_until} (now {now})"
            ),
            OfferError::LeaseExpired { lease_until, now } => write!(
                f,
                "card conflict in lease_expired: the holder's lease ended at {lease_until} \
                 (now {now}); the card goes the reclaim way"
            ),
            OfferError::State(e) => e.fmt(f),
        }
    }
}

impl std::error::Error for OfferError {}

impl From<LedgerError> for OfferError {
    fn from(e: LedgerError) -> Self {
        OfferError::State(ReclaimError::Ledger(e))
    }
}

/// The newest event of the offer family of one card, read from its payload.
struct Family {
    event_id: i64,
    kind: String,
    generation: Option<i64>,
    to: Option<String>,
    valid_until: Option<i64>,
}

fn newest_family_event(conn: &Connection, card_id: i64) -> Result<Option<Family>, LedgerError> {
    let row: Option<(i64, String, String)> = conn
        .query_row(
            "SELECT event_id, kind, payload FROM event \
             WHERE card_id = ?1 AND kind IN (?2, ?3) ORDER BY event_id DESC LIMIT 1",
            params![card_id, KIND_OFFER, KIND_CANCEL],
            |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)),
        )
        .optional()?;
    Ok(row.map(|(event_id, kind, payload)| {
        let v: Value = serde_json::from_str(&payload).unwrap_or(Value::Null);
        Family {
            event_id,
            kind,
            generation: v["generation"].as_i64(),
            to: v["to"].as_str().map(str::to_string),
            valid_until: v["valid_until"].as_i64(),
        }
    }))
}

fn invalid(what: String) -> FenceError {
    FenceError::InvalidInput(what)
}

/// Records an offer of the card the holder holds at `generation` to the session `to`. Fenced
/// like a handoff: the caller must be the holder at the current generation with a live lease,
/// and a refused holder call is recorded as one `fence_reject` event. The reservation, the
/// generation and the card's state stay as they are.
pub fn offer(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    to: &str,
    valid_secs: i64,
    handover: Option<&str>,
) -> Result<Offered, FenceError> {
    if to == holder {
        return Err(invalid("to is the holder itself".into()));
    }
    if to.is_empty() || to.chars().any(char::is_control) || to != to.trim() {
        return Err(invalid("to is not a session id".into()));
    }
    if !(MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&valid_secs) {
        return Err(invalid(format!(
            "valid_secs {valid_secs} is outside {MIN_LEASE_SECS}..={MAX_LEASE_SECS}"
        )));
    }
    fenced(
        w,
        "offer",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| {
            let valid_until = now + valid_secs;
            add_event(
                tx,
                now,
                holder,
                card_id,
                KIND_OFFER,
                &json!({
                    "holder": holder,
                    "generation": generation,
                    "to": to,
                    "valid_secs": valid_secs,
                    "valid_until": valid_until,
                    "handover": handover,
                    "lease_until": row.lease_until,
                }),
            )?;
            Ok(Offered {
                generation,
                to: to.to_string(),
                valid_secs,
                valid_until,
                handover: handover.map(str::to_string),
            })
        },
    )
}

/// Voids the newest offer of the card the holder holds at `generation`. Fenced like [`offer`].
/// Only a card whose newest event of the offer family is an offer at this generation can be
/// cancelled; anything else is [`OfferError::NoOffer`] and nothing is written.
pub fn cancel(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
) -> Result<Result<Cancelled, OfferError>, FenceError> {
    fenced(
        w,
        "offer-cancel",
        card_id,
        holder,
        generation,
        true,
        |tx, now, _row| {
            let open = newest_family_event(tx, card_id)?.filter(|e| {
                e.kind == KIND_OFFER && e.generation == Some(generation) && e.to.is_some()
            });
            let Some(offer) = open else {
                return Ok(Err(OfferError::NoOffer { generation }));
            };
            let to = offer.to.unwrap_or_default();
            add_event(
                tx,
                now,
                holder,
                card_id,
                KIND_CANCEL,
                &json!({
                    "holder": holder,
                    "generation": generation,
                    "to": to,
                    "offer_event": offer.event_id,
                }),
            )?;
            Ok(Ok(Cancelled { generation, to }))
        },
    )
}

/// Gives the reservation at `generation` to `accepter` at generation + 1 when the newest offer
/// of the card names `accepter` at that generation, is still valid, and the holder's lease is
/// alive (see the module text). The checks run in one transaction in this order, and the first
/// that fails is the answer: the state of the card as `card reclaim` words it, no offer, a
/// cancelled offer, a session that is not the target, an offer that ran out, a lease that
/// ended. Every refusal writes nothing, not even a `fence_reject` event: the caller is not the
/// holder the fence checks.
pub fn accept(
    w: &mut WriteLedger,
    card_id: i64,
    accepter: &str,
    generation: i64,
    lease_secs: i64,
) -> Result<Accepted, OfferError> {
    let bad = |why: String| OfferError::State(ReclaimError::InvalidInput(why));
    if accepter.is_empty() || accepter.chars().any(char::is_control) || accepter != accepter.trim()
    {
        return Err(bad("accepter is not a session id".into()));
    }
    if !(MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&lease_secs) {
        return Err(bad(format!(
            "lease_secs {lease_secs} is outside {MIN_LEASE_SECS}..={MAX_LEASE_SECS}"
        )));
    }
    w.with_write_tx(|tx, now| {
        let old = match adopt::peek(tx, card_id, generation) {
            Ok(old) => old,
            Err(ReclaimError::Ledger(e)) => return Err(e),
            Err(refusal) => return Ok(Err(OfferError::State(refusal))),
        };
        let newest = newest_family_event(tx, card_id)?;
        let Some(newest) = newest.filter(|e| e.generation == Some(generation)) else {
            return Ok(Err(OfferError::NoOffer { generation }));
        };
        if newest.kind == KIND_CANCEL {
            return Ok(Err(OfferError::OfferCancelled));
        }
        if newest.to.as_deref() != Some(accepter) {
            return Ok(Err(OfferError::NotTarget));
        }
        let valid_until = newest.valid_until.unwrap_or(0);
        if now >= valid_until {
            return Ok(Err(OfferError::OfferExpired { valid_until, now }));
        }
        if old.lease_until <= now {
            return Ok(Err(OfferError::LeaseExpired {
                lease_until: old.lease_until,
                now,
            }));
        }
        let state: String = tx.query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )?;
        tx.execute(
            "UPDATE reservation SET status = 'released', updated_at = ?1 WHERE reservation_id = ?2",
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
                accepter,
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
        add_event(
            tx,
            now,
            accepter,
            card_id,
            KIND_ACCEPT,
            &json!({
                "from_holder": old.holder,
                "from_generation": generation,
                "from_lease_until": old.lease_until,
                "holder": accepter,
                "generation": next,
                "lease_until": now + lease_secs,
                "offer_event": newest.event_id,
                "worktree": old.worktree,
                "branch": old.branch,
                "card_state": state,
            }),
        )?;
        let previous = load_reservation(tx, card_id, generation)?
            .ok_or_else(|| LedgerError::Sql("the offered reservation cannot be read".into()))?;
        let reservation = load_reservation(tx, card_id, next)?
            .ok_or_else(|| LedgerError::Sql("the new reservation cannot be read".into()))?;
        Ok(Ok(Accepted {
            previous,
            reservation,
            card_state: state,
            offer_event: newest.event_id,
        }))
    })?
}
