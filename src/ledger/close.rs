//! Closing a card that nobody works on (SPEC-20261124-card-fold-stale): `fold` merges a duplicate
//! card into another, `withdraw` also takes a card whose lease ended, and `stale_facts` reads what
//! `card next` needs to list the cards that want closing.
//!
//! Nothing here adds a state or a table: a folded or withdrawn card is `hold`, the marker is the
//! free event kind (`card.fold`, `card.withdraw`), and a fold adds one `relates` link.
//!
//! What may be closed, for both calls:
//! - a `draft`, `ready`, `active` or `review` card that no reservation holds;
//! - an `active` or `review` card whose only reservation is `held` with an ended lease, and only
//!   for the caller `--by leader` (the same word `card land` asks for). That call ends the
//!   reservation as `released` and records who held it. The worktree and the branch are not
//!   touched.
//!
//! A reservation that is not `held` (`expired_candidate`, `held_for_review`) belongs to `recover`
//! and is refused.

use super::reservation::{self, Reservation, ACTIVE_STATUSES};
use super::workflow::{LEADER_BY, MAX_AUTHOR_CHARS, MAX_LINE_CHARS};
use super::{LedgerError, WriteLedger};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};

/// Why a fold or a withdraw did not happen. Every refusal writes nothing.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum CloseError {
    CardNotFound(i64),
    /// The card is in a state the call does not close.
    WrongState {
        card_id: i64,
        state: String,
    },
    /// A live lease holds the card.
    Held {
        card_id: i64,
        holder: String,
        generation: i64,
    },
    /// The reservation is not `held` (`expired_candidate`, `held_for_review`): `recover` decides it.
    ReservationNotHeld {
        card_id: i64,
        status: String,
    },
    /// Ending a reservation is the leader's: `--by leader`.
    LeaderOnly {
        card_id: i64,
        holder: String,
        generation: i64,
    },
    /// A fold target that is the card itself, missing or `hold`.
    BadTarget {
        card_id: i64,
        into: i64,
        why: String,
    },
    InvalidInput(String),
    Ledger(LedgerError),
}

impl std::fmt::Display for CloseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CloseError::CardNotFound(id) => write!(f, "card not found: C-{id}"),
            CloseError::WrongState { card_id, state } => write!(
                f,
                "the card C-{card_id} is `{state}`; only a `draft`, `ready`, `active` or `review` card can be closed this way"
            ),
            CloseError::Held {
                card_id,
                holder,
                generation,
            } => write!(
                f,
                "the card C-{card_id} is held by {holder:?} at generation {generation} and the lease has not ended"
            ),
            CloseError::ReservationNotHeld { card_id, status } => write!(
                f,
                "the reservation of C-{card_id} is `{status}`, not `held`; `recover` decides it"
            ),
            CloseError::LeaderOnly {
                card_id,
                holder,
                generation,
            } => write!(
                f,
                "the lease of C-{card_id} ({holder:?}, generation {generation}) has ended; only the leader ends a reservation: give --by {LEADER_BY}"
            ),
            CloseError::BadTarget { card_id, into, why } => {
                write!(f, "C-{card_id} cannot fold into C-{into}: {why}")
            }
            CloseError::InvalidInput(why) => write!(f, "{why}"),
            CloseError::Ledger(e) => write!(f, "{e}"),
        }
    }
}

impl std::error::Error for CloseError {}

impl From<LedgerError> for CloseError {
    fn from(e: LedgerError) -> Self {
        CloseError::Ledger(e)
    }
}

/// One line of 1 to `max` characters that is not only blanks and holds no control or directional
/// character, as `workflow::withdraw` asks of its text.
fn one_line(what: &str, value: &str, max: usize) -> Result<(), CloseError> {
    let n = value.chars().count();
    let bad = if value.trim().is_empty() {
        Some(format!("{what} is empty"))
    } else if n > max {
        Some(format!("{what} is longer than {max} characters ({n})"))
    } else if value.chars().any(char::is_control) {
        Some(format!("{what} is not one line"))
    } else if value
        .chars()
        .any(|c| matches!(c, '\u{2028}'..='\u{202e}' | '\u{2066}'..='\u{2069}'))
    {
        Some(format!(
            "{what} holds a directional or formatting character (U+2028–U+202E, U+2066–U+2069)"
        ))
    } else {
        None
    };
    bad.map_or(Ok(()), |why| Err(CloseError::InvalidInput(why)))
}

fn check_text(reason: &str, by: &str, session: Option<&str>) -> Result<(), CloseError> {
    one_line("the reason", reason, MAX_LINE_CHARS)?;
    one_line("the by", by, MAX_AUTHOR_CHARS)?;
    if let Some(session) = session {
        one_line("the session id", session, MAX_AUTHOR_CHARS)?;
    }
    Ok(())
}

/// The reservation a close call would end, or `None` when nothing holds the card; an `Err`
/// when the card may not be closed now. Every refusal comes before any write.
fn closable(
    tx: &Connection,
    now: i64,
    card_id: i64,
    by: &str,
) -> Result<Result<Option<Reservation>, CloseError>, LedgerError> {
    let state: Option<String> = tx
        .query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .optional()?;
    let Some(state) = state else {
        return Ok(Err(CloseError::CardNotFound(card_id)));
    };
    let wrong_state = |state: String| Err(CloseError::WrongState { card_id, state });
    match reservation::current(tx, card_id)? {
        Some(held) => {
            if held.status != "held" {
                return Ok(Err(CloseError::ReservationNotHeld {
                    card_id,
                    status: held.status,
                }));
            }
            if !reservation::lease_ended(held.lease_until, now) {
                return Ok(Err(CloseError::Held {
                    card_id,
                    holder: held.holder,
                    generation: held.generation,
                }));
            }
            if by != LEADER_BY {
                return Ok(Err(CloseError::LeaderOnly {
                    card_id,
                    holder: held.holder,
                    generation: held.generation,
                }));
            }
            if state != "active" && state != "review" {
                return Ok(wrong_state(state));
            }
            Ok(Ok(Some(held)))
        }
        None => {
            if matches!(state.as_str(), "draft" | "ready" | "active" | "review") {
                Ok(Ok(None))
            } else {
                Ok(wrong_state(state))
            }
        }
    }
}

/// Ends `held` as `released` and moves the card to `hold`; returns the part of the event that
/// names the former holder (`Null` when nothing held the card).
fn close_card(
    tx: &Connection,
    now: i64,
    card_id: i64,
    held: Option<&Reservation>,
) -> Result<Value, LedgerError> {
    let mut released = Value::Null;
    if let Some(r) = held {
        tx.execute(
            "UPDATE reservation SET status = 'released', updated_at = ?1 WHERE reservation_id = ?2",
            params![now, r.reservation_id],
        )?;
        released = json!({
            "holder": r.holder, "generation": r.generation, "lease_until": r.lease_until,
        });
    }
    tx.execute(
        "UPDATE card SET state = 'hold', updated_at = ?2 WHERE card_id = ?1",
        params![card_id, now],
    )?;
    Ok(released)
}

/// Withdraws a card: it moves to `hold` and one `card.withdraw` event records the reason and the
/// by (and the former holder, when a lapsed reservation was ended). `card take` then refuses it
/// naming the reason and the default `card list` hides it.
pub fn withdraw(
    w: &mut WriteLedger,
    card_id: i64,
    reason: &str,
    by: &str,
    session: Option<&str>,
) -> Result<(), CloseError> {
    check_text(reason, by, session)?;
    w.with_write_tx(|tx, now| {
        let held = match closable(tx, now, card_id, by)? {
            Ok(held) => held,
            Err(e) => return Ok(Err(e)),
        };
        let released = close_card(tx, now, card_id, held.as_ref())?;
        let mut payload = json!({"reason": reason, "by": by});
        if !released.is_null() {
            payload["released"] = released;
        }
        reservation::add_event(
            tx,
            now,
            session.unwrap_or("user"),
            card_id,
            "card.withdraw",
            &payload,
        )?;
        Ok(Ok(()))
    })
    .map_err(CloseError::from)?
}

/// Folds `card_id` into `into`: the card moves to `hold`, one `card.fold` event names the target,
/// the reason and the by, and one active `relates` link runs from the card to the target. The
/// card's notes, criteria and links stay as they are. The target must exist and must not be
/// `hold`.
pub fn fold(
    w: &mut WriteLedger,
    card_id: i64,
    into: i64,
    reason: &str,
    by: &str,
    session: Option<&str>,
) -> Result<(), CloseError> {
    check_text(reason, by, session)?;
    w.with_write_tx(|tx, now| {
        let held = match closable(tx, now, card_id, by)? {
            Ok(held) => held,
            Err(e) => return Ok(Err(e)),
        };
        let bad_target = |why: &str| {
            Err(CloseError::BadTarget {
                card_id,
                into,
                why: why.to_string(),
            })
        };
        if into == card_id {
            return Ok(bad_target("a card cannot fold into itself"));
        }
        let target: Option<String> = tx
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [into],
                |r| r.get(0),
            )
            .optional()?;
        match target.as_deref() {
            None => return Ok(bad_target("the target card does not exist")),
            Some("hold") => {
                return Ok(bad_target(
                    "the target card is `hold` (withdrawn or folded); fold into the card that took its place",
                ))
            }
            Some(_) => {}
        }
        let released = close_card(tx, now, card_id, held.as_ref())?;
        let link: Option<(i64, String)> = tx
            .query_row(
                "SELECT link_id, state FROM card_link \
                 WHERE from_card = ?1 AND kind = 'relates' AND to_card = ?2",
                params![card_id, into],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .optional()?;
        match link {
            None => {
                tx.execute(
                    "INSERT INTO card_link (from_card, kind, to_card, created_at, updated_at) \
                     VALUES (?1, 'relates', ?2, ?3, ?3)",
                    params![card_id, into, now],
                )?;
            }
            Some((link_id, state)) if state != "active" => {
                tx.execute(
                    "UPDATE card_link SET state = 'active', updated_at = ?2 WHERE link_id = ?1",
                    params![link_id, now],
                )?;
            }
            Some(_) => {}
        }
        let mut payload = json!({"into": format!("C-{into}"), "reason": reason, "by": by});
        if !released.is_null() {
            payload["released"] = released;
        }
        reservation::add_event(
            tx,
            now,
            session.unwrap_or("user"),
            card_id,
            "card.fold",
            &payload,
        )?;
        Ok(Ok(()))
    })
    .map_err(CloseError::from)?
}

/// A card whose `held` reservation has an ended lease.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LeaseEnded {
    pub card_id: i64,
    pub title: String,
    pub state: String,
    pub priority: String,
    pub holder: String,
    pub generation: i64,
    pub lease_until: i64,
    /// The holder's heartbeat (SPEC-20261119-live-lease-takeover REQ-1): the value the sweep reads.
    pub last_seen_at: i64,
}

/// A card with what a row of `stale` needs; `updated_at` is its last update (DB clock).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Plain {
    pub card_id: i64,
    pub title: String,
    pub state: String,
    pub priority: String,
    pub updated_at: i64,
}

/// What the ledger says about the cards that may want closing.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct StaleFacts {
    /// `active` or `review` cards whose `held` lease ended.
    pub lease_ended: Vec<LeaseEnded>,
    /// `review` cards no reservation holds.
    pub review_no_holder: Vec<Plain>,
    /// Every `draft` card with its last update: the age filter is the caller's.
    pub drafts: Vec<Plain>,
    /// Every `active` or `review` card: the ones a landing proof is asked for.
    pub open: Vec<Plain>,
}

fn plain(sql: &str, conn: &Connection) -> Result<Vec<Plain>, LedgerError> {
    let mut stmt = conn.prepare(sql)?;
    let rows = stmt.query_map([], |r| {
        Ok(Plain {
            card_id: r.get(0)?,
            title: r.get(1)?,
            state: r.get(2)?,
            priority: r.get(3)?,
            updated_at: r.get(4)?,
        })
    })?;
    Ok(rows.collect::<Result<_, _>>()?)
}

/// Reads the facts at `now` (the DB clock of the read). Read-only: it works on a
/// [`super::ReadLedger::conn`].
pub fn stale_facts(conn: &Connection, now: i64) -> Result<StaleFacts, LedgerError> {
    let mut stmt = conn.prepare(
        "SELECT c.card_id, c.title, c.state, c.priority, r.holder, r.generation, r.lease_until, \
                r.updated_at \
         FROM reservation r JOIN card c ON c.card_id = r.card_id \
         WHERE r.status = 'held' AND r.lease_until <= ?1 AND c.state IN ('active', 'review') \
         ORDER BY c.card_id",
    )?;
    // The heartbeat is read with the function the sweep uses, so a row and a sweep never read
    // two different values (SPEC-20261119-live-lease-takeover D-7).
    let ended = stmt
        .query_map([now], |r| {
            Ok((
                LeaseEnded {
                    card_id: r.get(0)?,
                    title: r.get(1)?,
                    state: r.get(2)?,
                    priority: r.get(3)?,
                    holder: r.get(4)?,
                    generation: r.get(5)?,
                    lease_until: r.get(6)?,
                    last_seen_at: 0,
                },
                r.get::<_, i64>(7)?,
            ))
        })?
        .collect::<Result<Vec<_>, _>>()?;
    let mut lease_ended = Vec::with_capacity(ended.len());
    for (mut c, updated_at) in ended {
        c.last_seen_at = reservation::heartbeat_at(conn, c.card_id, &c.holder, updated_at)?;
        lease_ended.push(c);
    }
    Ok(StaleFacts {
        lease_ended,
        review_no_holder: plain(
            &format!(
                "SELECT card_id, title, state, priority, updated_at FROM card c \
                 WHERE state = 'review' AND NOT EXISTS ( \
                   SELECT 1 FROM reservation r \
                   WHERE r.card_id = c.card_id AND r.status IN {ACTIVE_STATUSES}) \
                 ORDER BY card_id"
            ),
            conn,
        )?,
        drafts: plain(
            "SELECT card_id, title, state, priority, updated_at FROM card \
             WHERE state = 'draft' ORDER BY card_id",
            conn,
        )?,
        open: plain(
            "SELECT card_id, title, state, priority, updated_at FROM card \
             WHERE state IN ('active', 'review') ORDER BY card_id",
            conn,
        )?,
    })
}
