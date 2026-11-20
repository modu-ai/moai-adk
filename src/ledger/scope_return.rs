//! Giving paths back from a held reservation (SPEC-20261120-card-shared-line-lease).
//!
//! [`reservation::widen_in_tx`] makes a held scope grow. Three calls here make it shrink, each in
//! one write transaction behind the same fence (`held`, the offered holder and generation):
//!
//! - [`give_back`] removes the named paths and writes one `card.scope` event that names them;
//! - [`review_keeping`] is `card review`: `active -> review` and every path that `keep` does not
//!   name is given back, both lists in the `card.review` event;
//! - [`lease_take`] adds the paths of a line lease that the reservation does not hold yet, and
//!   [`lease_release`] gives exactly those back. The caller runs its one command between them.
//!
//! The rows of `reservation_scope` are deleted (no trigger forbids it; the rows of `reservation`,
//! `card`, notes and events are never deleted). The history of a scope stays in the events: every
//! change writes `from`, `to` and the paths given back. A given-back path is free for any card the
//! moment the transaction commits, because [`reservation::find_overlap`] reads those rows.
//!
//! Nothing here creates a file, calls `git` or reads the network.

use super::reservation::{
    self, add_event, FenceError, Reservation, WidenError, MAX_PATTERN_LEN, MAX_SCOPE_PATTERNS,
};
use super::workflow::{Moved, WorkflowError};
use super::{LedgerError, WriteLedger};
use rusqlite::{params, Connection};
use serde_json::json;

/// What a give-back did.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct GivenBack {
    /// The reservation as it is now, still `held`.
    pub reservation: Reservation,
    /// The scope before the call, sorted.
    pub from: Vec<String>,
    /// The paths removed, sorted.
    pub given_back: Vec<String>,
}

/// Why a path list was refused. Nothing is written.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ScopeReturnError {
    /// A list is empty, too long or has a path out of its limits.
    InvalidInput(String),
    /// The card has no active reservation.
    NotHeld(i64),
    /// These paths are not in the reservation's scope (an exact text match).
    NotInReservation { paths: Vec<String> },
}

impl std::fmt::Display for ScopeReturnError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::InvalidInput(why) => write!(f, "invalid reservation input: {why}"),
            Self::NotHeld(card_id) => write!(
                f,
                "card conflict in state: C-{card_id} has no active reservation to give paths back from"
            ),
            Self::NotInReservation { paths } => write!(
                f,
                "card conflict in scope: {} not in the reservation's scope; `card show` lists what \
                 it holds; nothing was changed",
                list(paths)
            ),
        }
    }
}

impl std::error::Error for ScopeReturnError {}

/// A [`give_back`] refused.
#[derive(Debug)]
pub enum GiveBackError {
    Fence(FenceError),
    Scope(ScopeReturnError),
}

impl std::fmt::Display for GiveBackError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Fence(e) => write!(f, "{e}"),
            Self::Scope(e) => write!(f, "{e}"),
        }
    }
}

impl From<FenceError> for GiveBackError {
    fn from(e: FenceError) -> Self {
        Self::Fence(e)
    }
}

/// A [`review_keeping`] refused.
#[derive(Debug)]
pub enum ReviewKeepError {
    Workflow(WorkflowError),
    Scope(ScopeReturnError),
}

impl std::fmt::Display for ReviewKeepError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Workflow(e) => write!(f, "{e}"),
            Self::Scope(e) => write!(f, "{e}"),
        }
    }
}

impl From<FenceError> for ReviewKeepError {
    fn from(e: FenceError) -> Self {
        Self::Workflow(WorkflowError::from(e))
    }
}

impl From<LedgerError> for ReviewKeepError {
    fn from(e: LedgerError) -> Self {
        Self::Workflow(WorkflowError::Ledger(e))
    }
}

/// A [`lease_take`] that did not get the paths.
#[derive(Debug)]
pub enum LeaseError {
    Fence(FenceError),
    /// Another card's reservation covers a path: the text names that card. A caller may retry.
    Overlap(String),
    /// Anything else the widening refused (limits, letter case, ...). A retry does not help.
    Refused(String),
}

impl std::fmt::Display for LeaseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Fence(e) => write!(f, "{e}"),
            Self::Overlap(text) | Self::Refused(text) => f.write_str(text),
        }
    }
}

impl From<FenceError> for LeaseError {
    fn from(e: FenceError) -> Self {
        Self::Fence(e)
    }
}

/// The paths a [`lease_take`] added (`added` is empty when the reservation held them all).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Leased {
    pub added: Vec<String>,
    pub reservation: Reservation,
}

fn list(paths: &[String]) -> String {
    let quoted: Vec<String> = paths.iter().map(|p| format!("{p:?}")).collect();
    match quoted.len() {
        1 => format!("{} is", quoted[0]),
        _ => format!("{} are", quoted.join(", ")),
    }
}

/// A path list as stored: not empty, at most [`MAX_SCOPE_PATTERNS`] entries of 1 to
/// [`MAX_PATTERN_LEN`] characters, sorted, duplicates dropped.
fn clean(paths: &[String]) -> Result<Vec<String>, ScopeReturnError> {
    if paths.is_empty() {
        return Err(ScopeReturnError::InvalidInput(
            "give at least one path".to_string(),
        ));
    }
    if paths.len() > MAX_SCOPE_PATTERNS {
        return Err(ScopeReturnError::InvalidInput(format!(
            "{} paths given, at most {MAX_SCOPE_PATTERNS}",
            paths.len()
        )));
    }
    for p in paths {
        let n = p.chars().count();
        if n == 0 || n > MAX_PATTERN_LEN || p.contains(['\0', '\n', '\r']) {
            return Err(ScopeReturnError::InvalidInput(format!(
                "a path is 1 to {MAX_PATTERN_LEN} characters on one line: {p:?}"
            )));
        }
    }
    let mut out = paths.to_vec();
    out.sort();
    out.dedup();
    Ok(out)
}

/// Removes `paths` from the active reservation of `card_id` inside the caller's write
/// transaction. `Ok(Err(..))` is a refusal that wrote nothing. Writes one `card.scope` event.
pub fn give_back_in_tx(
    tx: &Connection,
    now: i64,
    card_id: i64,
    by: &str,
    paths: &[String],
    reason: &str,
) -> Result<Result<GivenBack, ScopeReturnError>, LedgerError> {
    let Some(active) = reservation::current(tx, card_id)? else {
        return Ok(Err(ScopeReturnError::NotHeld(card_id)));
    };
    let from = active.write_scope.clone();
    let missing: Vec<String> = paths
        .iter()
        .filter(|p| !from.contains(p))
        .cloned()
        .collect();
    if !missing.is_empty() {
        return Ok(Err(ScopeReturnError::NotInReservation { paths: missing }));
    }
    for path in paths {
        tx.execute(
            "DELETE FROM reservation_scope WHERE reservation_id = ?1 AND pattern = ?2",
            params![active.reservation_id, path],
        )?;
    }
    tx.execute(
        "UPDATE reservation SET updated_at = ?1 WHERE reservation_id = ?2",
        params![now, active.reservation_id],
    )?;
    let to: Vec<String> = from
        .iter()
        .filter(|p| !paths.contains(p))
        .cloned()
        .collect();
    add_event(
        tx,
        now,
        by,
        card_id,
        "card.scope",
        &json!({
            "from": from,
            "to": to,
            "given_back": paths,
            "held": true,
            "generation": active.generation,
            "holder": active.holder,
            "reason": reason,
        }),
    )?;
    let reservation = reservation::current(tx, card_id)?
        .ok_or_else(|| LedgerError::Sql("the reservation cannot be read".into()))?;
    Ok(Ok(GivenBack {
        reservation,
        from,
        given_back: paths.to_vec(),
    }))
}

/// `card give-back`: the holder at `generation` gives `paths` back. The lease need not be live:
/// a holder whose lease ran out may still free its files while nobody reclaimed the card.
pub fn give_back(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    paths: &[String],
    reason: &str,
) -> Result<GivenBack, GiveBackError> {
    let paths = clean(paths).map_err(GiveBackError::Scope)?;
    let inner = reservation::fenced(
        w,
        "give-back",
        card_id,
        holder,
        generation,
        false,
        |tx, now, _| give_back_in_tx(tx, now, card_id, holder, &paths, reason),
    )?;
    inner.map_err(GiveBackError::Scope)
}

/// `card review`: `active -> review`, and every path that `keep` does not name is given back,
/// in one transaction. A `keep` path the reservation does not hold refuses the whole call.
/// The `card.review` event names `given_back` and `kept`. Returns the move and both lists.
pub fn review_keeping(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    keep: &[String],
) -> Result<(Moved, Vec<String>, Vec<String>), ReviewKeepError> {
    let mut keep = keep.to_vec();
    keep.sort();
    keep.dedup();
    let inner = reservation::fenced(
        w,
        "review",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| -> Result<Result<(Moved, Vec<String>, Vec<String>), ReviewKeepError>, LedgerError> {
            let from: String = tx.query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [card_id],
                |r| r.get(0),
            )?;
            if from != "active" {
                return Ok(Err(ReviewKeepError::Workflow(WorkflowError::WrongState {
                    card_id,
                    state: from,
                    needs: "active",
                    action: "sent to review",
                })));
            }
            let held = row.write_scope.clone();
            let missing: Vec<String> =
                keep.iter().filter(|p| !held.contains(p)).cloned().collect();
            if !missing.is_empty() {
                return Ok(Err(ReviewKeepError::Scope(
                    ScopeReturnError::NotInReservation { paths: missing },
                )));
            }
            let given_back: Vec<String> =
                held.iter().filter(|p| !keep.contains(p)).cloned().collect();
            for path in &given_back {
                tx.execute(
                    "DELETE FROM reservation_scope WHERE reservation_id = ?1 AND pattern = ?2",
                    params![row.reservation_id, path],
                )?;
            }
            tx.execute(
                "UPDATE reservation SET updated_at = ?1 WHERE reservation_id = ?2",
                params![now, row.reservation_id],
            )?;
            tx.execute(
                "UPDATE card SET state = 'review', updated_at = ?2 WHERE card_id = ?1",
                params![card_id, now],
            )?;
            add_event(
                tx,
                now,
                holder,
                card_id,
                "card.review",
                &json!({
                    "from": from,
                    "generation": generation,
                    "given_back": given_back,
                    "kept": keep,
                }),
            )?;
            let rev: i64 =
                tx.query_row("SELECT rev FROM card WHERE card_id = ?1", [card_id], |r| {
                    r.get(0)
                })?;
            let reservation = reservation::current(tx, card_id)?
                .ok_or_else(|| LedgerError::Sql("the reservation cannot be read".into()))?;
            Ok(Ok((
                Moved {
                    card_id,
                    from,
                    state: "review".to_string(),
                    rev,
                    closed_at: None,
                    reservation,
                },
                given_back,
                keep.clone(),
            )))
        },
    )?;
    inner
}

/// The first step of a line lease: adds the paths that the reservation does not hold yet, by the
/// same check as a widening (another card's reservation of the same path refuses it with
/// [`LeaseError::Overlap`], which a caller may retry). A path the reservation already holds is
/// left alone and is not part of `added`, so [`lease_release`] never gives it back.
pub fn lease_take(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    paths: &[String],
) -> Result<Leased, LeaseError> {
    let paths = clean(paths).map_err(|e| LeaseError::Refused(e.to_string()))?;
    let inner = reservation::fenced(
        w,
        "line-lease",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| -> Result<Result<Leased, LeaseError>, LedgerError> {
            let added: Vec<String> = paths
                .iter()
                .filter(|p| !row.write_scope.iter().any(|h| h.eq_ignore_ascii_case(p)))
                .cloned()
                .collect();
            if added.is_empty() {
                return Ok(Ok(Leased {
                    added,
                    reservation: row,
                }));
            }
            let mut new_scope = row.write_scope.clone();
            new_scope.extend(added.iter().cloned());
            match reservation::widen_in_tx(tx, now, card_id, &new_scope, holder, "line lease")? {
                Ok(widened) => Ok(Ok(Leased {
                    added: widened.added,
                    reservation: widened.reservation,
                })),
                Err(e @ WidenError::ScopeOverlap { .. }) => {
                    Ok(Err(LeaseError::Overlap(e.to_string())))
                }
                Err(other) => Ok(Err(LeaseError::Refused(other.to_string()))),
            }
        },
    )?;
    inner
}

/// The last step of a line lease: gives the paths [`lease_take`] added back.
pub fn lease_release(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    added: &[String],
) -> Result<Option<GivenBack>, GiveBackError> {
    if added.is_empty() {
        return Ok(None);
    }
    give_back(w, card_id, holder, generation, added, "line lease ended").map(Some)
}
