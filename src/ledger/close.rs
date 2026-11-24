//! Closing a card that nobody works on (SPEC-20261124-card-fold-stale): `fold` merges a duplicate
//! card into another, `withdraw` also takes a card whose lease ended, and `stale_facts` reads what
//! `card next` needs to list the cards that want closing.
//!
//! Nothing here adds a state or a table: a folded or withdrawn card is `hold`, the marker is the
//! free event kind (`card.fold`, `card.withdraw`), and a fold adds one `relates` link.

use super::{LedgerError, WriteLedger};
use rusqlite::Connection;

/// Why a fold or a withdraw did not happen.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum CloseError {
    CardNotFound(i64),
    /// The card is in a state the call does not close.
    WrongState { card_id: i64, state: String },
    /// A live lease holds the card.
    Held {
        card_id: i64,
        holder: String,
        generation: i64,
    },
    /// The reservation is not `held` (`expired_candidate`, `held_for_review`): `recover` decides it.
    ReservationNotHeld { card_id: i64, status: String },
    /// Ending a reservation is the leader's: `--by leader`.
    LeaderOnly { card_id: i64, holder: String },
    /// A fold target that is the card itself, missing or `hold`.
    BadTarget { card_id: i64, into: i64, why: String },
    InvalidInput(String),
    Ledger(LedgerError),
}

impl std::fmt::Display for CloseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "not implemented")
    }
}

impl std::error::Error for CloseError {}

impl From<LedgerError> for CloseError {
    fn from(e: LedgerError) -> Self {
        CloseError::Ledger(e)
    }
}

/// Withdraws a card: `hold`, one `card.withdraw` event. See the SPEC for what may be withdrawn.
pub fn withdraw(
    _w: &mut WriteLedger,
    _card_id: i64,
    _reason: &str,
    _by: &str,
    _session: Option<&str>,
) -> Result<(), CloseError> {
    Err(CloseError::InvalidInput("not implemented".into()))
}

/// Folds `card_id` into `into`: `hold`, one `card.fold` event, one `relates` link.
pub fn fold(
    _w: &mut WriteLedger,
    _card_id: i64,
    _into: i64,
    _reason: &str,
    _by: &str,
    _session: Option<&str>,
) -> Result<(), CloseError> {
    Err(CloseError::InvalidInput("not implemented".into()))
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
}

/// A card of the state the kind names, with what its row needs.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Plain {
    pub card_id: i64,
    pub title: String,
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
    /// Every `draft` card with its last update.
    pub drafts: Vec<Plain>,
    /// Every `active` or `review` card: the ones a landing proof is asked for.
    pub open: Vec<(i64, String)>,
}

/// Reads the facts. `now` is the DB clock of the read. Read-only.
pub fn stale_facts(_conn: &Connection, _now: i64) -> Result<StaleFacts, LedgerError> {
    Ok(StaleFacts::default())
}
