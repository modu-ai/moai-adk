//! Reservations, generations, write scopes and the fence (Issue #3068 P3; contract comments
//! 6896940106, 6897392517 and 6904689241).
//!
//! A card is worked by one holder at a time. [`reserve`] takes a `ready` card, issues the next
//! generation (`card.generation + 1`), records the holder, the lease and the write scope, and
//! moves the card `ready -> active`, all in one `BEGIN IMMEDIATE` transaction. The unique index
//! `reservation_one_active` makes a second holder of the same card impossible, and the scope
//! check ([`crate::scope::overlap`], the one predicate for "could both write this path") runs in
//! the same transaction against every other card's active reservation.
//!
//! The generation is the fence: [`renew`], [`release`], [`handoff`] and [`fence`] accept only
//! the current generation, held by the offered holder, with a lease that has not ended (a
//! release does not need a live lease). A lower generation, another holder, a reservation that
//! is no longer `held` or an ended lease is rejected with a [`FenceError`], and the rejection is
//! written to `event` (kind `fence_reject`) in the same transaction that finds it, which then
//! commits: the record survives the rejection.
//!
//! Every time is the DB clock (`unixepoch()` read after the write lock was won); no caller
//! passes a time. Nothing here creates a file, calls `git` or reads the network.
//!
//! What a rejected call leaves behind is in `docs/design-review/appendix/ledger_reservation.md`.

use super::{new_event_key, LedgerError, ReadLedger, WriteLedger};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};

/// Shortest lease a caller may ask for, in seconds.
pub const MIN_LEASE_SECS: i64 = 60;
/// Longest lease a caller may ask for, in seconds (24 hours).
pub const MAX_LEASE_SECS: i64 = 86_400;
/// Most write-scope patterns one reservation may hold.
pub const MAX_SCOPE_PATTERNS: usize = 64;
/// Longest write-scope pattern, in characters.
pub const MAX_PATTERN_LEN: usize = 512;
/// Longest holder name, in characters (shortest is 1).
pub const MAX_HOLDER_LEN: usize = 200;
/// Longest worktree path or branch name, in characters.
pub const MAX_FIELD_LEN: usize = 512;

/// The `event.kind` of a rejected fence-checked call.
pub const EVENT_FENCE_REJECT: &str = "fence_reject";

/// The statuses that count as "someone holds the card" (the partial unique index).
pub(super) const ACTIVE_STATUSES: &str = "('held','expired_candidate','held_for_review')";

/// One row of `reservation` with its write scope.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Reservation {
    pub reservation_id: i64,
    pub card_id: i64,
    pub generation: i64,
    pub holder: String,
    /// `held`, `released`, `expired_candidate`, `reclaimed` or `held_for_review`.
    pub status: String,
    pub acquired_at: i64,
    /// First second at which the lease no longer holds (DB clock).
    pub lease_until: i64,
    pub updated_at: i64,
    pub worktree: Option<String>,
    pub branch: Option<String>,
    /// The write-scope patterns, sorted and without duplicates.
    pub write_scope: Vec<String>,
}

/// What [`release`] did.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Released {
    /// The reservation as it is now (`status: released`).
    pub reservation: Reservation,
    /// The card's state after the call: `ready` when it was `active`, otherwise unchanged.
    pub card_state: String,
}

// ---- errors ------------------------------------------------------------------------------------

/// Why [`reserve`] did not make a reservation. [`ReserveError::code`] is the stable word.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ReserveError {
    /// A value is out of its limits or malformed. Nothing was read or written.
    InvalidInput(String),
    /// The write scope is empty: an overlap with other cards cannot be shown for no pattern.
    EmptyScope,
    CardNotFound(i64),
    /// The card is not `ready`.
    CardNotReady {
        state: String,
    },
    /// The card already has an active reservation (another holder, or this one again).
    AlreadyHeld {
        holder: String,
        generation: i64,
    },
    /// A pattern of the scope could be written under a pattern of another card's active
    /// reservation.
    ScopeOverlap {
        other_card: i64,
        mine: String,
        theirs: String,
    },
    /// The ledger itself failed (busy, unavailable, constraint ...).
    Ledger(LedgerError),
}

impl ReserveError {
    pub fn code(&self) -> &'static str {
        match self {
            ReserveError::InvalidInput(_) => "invalid_input",
            ReserveError::EmptyScope => "empty_scope",
            ReserveError::CardNotFound(_) => "card_not_found",
            ReserveError::CardNotReady { .. } => "card_not_ready",
            ReserveError::AlreadyHeld { .. } => "already_held",
            ReserveError::ScopeOverlap { .. } => "scope_overlap",
            ReserveError::Ledger(e) => e.code(),
        }
    }
}

impl std::fmt::Display for ReserveError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ReserveError::InvalidInput(why) => write!(f, "invalid reservation input: {why}"),
            ReserveError::EmptyScope => f.write_str(
                "invalid reservation input: the write scope is empty; name at least one path",
            ),
            ReserveError::CardNotFound(id) => write!(f, "card not found: C-{id}"),
            ReserveError::CardNotReady { state } => write!(
                f,
                "reservation conflict in card: the card is `{state}`, only a `ready` card can be reserved"
            ),
            ReserveError::AlreadyHeld { holder, generation } => write!(
                f,
                "reservation conflict in card: already held by {holder:?} at generation {generation}"
            ),
            ReserveError::ScopeOverlap {
                other_card,
                mine,
                theirs,
            } => write!(
                f,
                "reservation conflict in write scope: {mine:?} could be written under {theirs:?} reserved for card C-{other_card}"
            ),
            ReserveError::Ledger(e) => e.fmt(f),
        }
    }
}

impl std::error::Error for ReserveError {}

impl From<LedgerError> for ReserveError {
    fn from(e: LedgerError) -> Self {
        ReserveError::Ledger(e)
    }
}

/// Why [`renew`], [`release`], [`handoff`] or [`fence`] refused. Every variant that names a
/// reservation problem (all but `InvalidInput`, `CardNotFound` and `Ledger`) is also written to
/// `event` as `fence_reject`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum FenceError {
    /// A value is out of its limits or malformed. Nothing was read or written.
    InvalidInput(String),
    CardNotFound(i64),
    /// The card has never had a reservation.
    NoReservation,
    /// The offered generation is lower than the current one: a newer holder exists or existed.
    StaleGeneration {
        offered: i64,
        current: i64,
    },
    /// The offered generation was never issued.
    UnknownGeneration {
        offered: i64,
        current: i64,
    },
    /// The current reservation is not `held` (released, a reclaim candidate, reclaimed ...).
    NotHeld {
        status: String,
    },
    /// The current reservation belongs to another holder.
    WrongHolder {
        held_by: String,
    },
    /// The current reservation's lease ended at `lease_until` (DB clock `now`).
    LeaseExpired {
        lease_until: i64,
        now: i64,
    },
    /// The ledger itself failed (busy, unavailable, constraint ...).
    Ledger(LedgerError),
}

impl FenceError {
    pub fn code(&self) -> &'static str {
        match self {
            FenceError::InvalidInput(_) => "invalid_input",
            FenceError::CardNotFound(_) => "card_not_found",
            FenceError::NoReservation => "no_reservation",
            FenceError::StaleGeneration { .. } => "stale_generation",
            FenceError::UnknownGeneration { .. } => "unknown_generation",
            FenceError::NotHeld { .. } => "not_held",
            FenceError::WrongHolder { .. } => "wrong_holder",
            FenceError::LeaseExpired { .. } => "lease_expired",
            FenceError::Ledger(e) => e.code(),
        }
    }

    /// Whether this refusal is a reservation problem that gets an `event` row.
    fn is_recorded(&self) -> bool {
        !matches!(
            self,
            FenceError::InvalidInput(_) | FenceError::CardNotFound(_) | FenceError::Ledger(_)
        )
    }
}

impl std::fmt::Display for FenceError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            FenceError::InvalidInput(why) => write!(f, "invalid reservation input: {why}"),
            FenceError::CardNotFound(id) => write!(f, "card not found: C-{id}"),
            FenceError::NoReservation => {
                f.write_str("reservation conflict in card: the card has no reservation")
            }
            FenceError::StaleGeneration { offered, current } => write!(
                f,
                "reservation conflict in card: generation {offered} is stale, the current generation is {current}"
            ),
            FenceError::UnknownGeneration { offered, current } => write!(
                f,
                "reservation conflict in card: generation {offered} was never issued, the current generation is {current}"
            ),
            FenceError::NotHeld { status } => write!(
                f,
                "reservation conflict in card: the current reservation is `{status}`, not `held`"
            ),
            FenceError::WrongHolder { held_by } => write!(
                f,
                "reservation conflict in card: the current reservation is held by {held_by:?}"
            ),
            FenceError::LeaseExpired { lease_until, now } => write!(
                f,
                "reservation conflict in card: the lease ended at {lease_until} (now {now})"
            ),
            FenceError::Ledger(e) => e.fmt(f),
        }
    }
}

impl std::error::Error for FenceError {}

impl From<LedgerError> for FenceError {
    fn from(e: LedgerError) -> Self {
        FenceError::Ledger(e)
    }
}

// ---- input checks ------------------------------------------------------------------------------

fn check_text(what: &str, value: &str, max: usize) -> Result<(), String> {
    let n = value.chars().count();
    if n == 0 {
        Err(format!("{what} is empty"))
    } else if n > max {
        Err(format!("{what} is longer than {max} characters ({n})"))
    } else if value.chars().any(char::is_control) {
        Err(format!("{what} contains a control character"))
    } else {
        Ok(())
    }
}

fn check_holder(holder: &str) -> Result<(), String> {
    check_text("holder", holder, MAX_HOLDER_LEN)?;
    if holder != holder.trim() {
        return Err("holder starts or ends with a space".into());
    }
    Ok(())
}

fn check_lease(lease_secs: i64) -> Result<(), String> {
    if (MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&lease_secs) {
        Ok(())
    } else {
        Err(format!(
            "lease_secs {lease_secs} is outside {MIN_LEASE_SECS}..={MAX_LEASE_SECS}"
        ))
    }
}

fn check_optional(what: &str, value: Option<&str>) -> Result<(), String> {
    match value {
        Some(v) => check_text(what, v, MAX_FIELD_LEN),
        None => Ok(()),
    }
}

/// The scope as stored: not empty, at most [`MAX_SCOPE_PATTERNS`] patterns of 1 to
/// [`MAX_PATTERN_LEN`] characters, sorted, duplicates dropped (the count limit applies to the
/// list as given).
fn check_scope(write_scope: &[String]) -> Result<Vec<String>, ReserveError> {
    if write_scope.is_empty() {
        return Err(ReserveError::EmptyScope);
    }
    if write_scope.len() > MAX_SCOPE_PATTERNS {
        return Err(ReserveError::InvalidInput(format!(
            "write scope has {} patterns, at most {MAX_SCOPE_PATTERNS}",
            write_scope.len()
        )));
    }
    let mut out: Vec<String> = Vec::with_capacity(write_scope.len());
    for pattern in write_scope {
        check_text("write-scope pattern", pattern, MAX_PATTERN_LEN)
            .map_err(ReserveError::InvalidInput)?;
        out.push(pattern.clone());
    }
    out.sort();
    out.dedup();
    Ok(out)
}

// ---- reads -------------------------------------------------------------------------------------

fn card_row(conn: &Connection, card_id: i64) -> Result<Option<(String, i64)>, LedgerError> {
    Ok(conn
        .query_row(
            "SELECT state, generation FROM card WHERE card_id = ?1",
            [card_id],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .optional()?)
}

pub(super) fn load_reservation(
    conn: &Connection,
    card_id: i64,
    generation: i64,
) -> Result<Option<Reservation>, LedgerError> {
    load_where(
        conn,
        "card_id = ?1 AND generation = ?2",
        params![card_id, generation],
    )
}

fn load_where(
    conn: &Connection,
    clause: &str,
    args: impl rusqlite::Params,
) -> Result<Option<Reservation>, LedgerError> {
    let row = conn
        .query_row(
            &format!(
                "SELECT reservation_id, card_id, generation, holder, status, acquired_at, \
                 lease_until, updated_at, worktree, branch FROM reservation WHERE {clause} \
                 ORDER BY generation DESC LIMIT 1"
            ),
            args,
            |r| {
                Ok(Reservation {
                    reservation_id: r.get(0)?,
                    card_id: r.get(1)?,
                    generation: r.get(2)?,
                    holder: r.get(3)?,
                    status: r.get(4)?,
                    acquired_at: r.get(5)?,
                    lease_until: r.get(6)?,
                    updated_at: r.get(7)?,
                    worktree: r.get(8)?,
                    branch: r.get(9)?,
                    write_scope: Vec::new(),
                })
            },
        )
        .optional()?;
    let Some(mut reservation) = row else {
        return Ok(None);
    };
    let mut stmt = conn.prepare(
        "SELECT pattern FROM reservation_scope WHERE reservation_id = ?1 ORDER BY pattern",
    )?;
    reservation.write_scope = stmt
        .query_map([reservation.reservation_id], |r| r.get(0))?
        .collect::<Result<_, _>>()?;
    Ok(Some(reservation))
}

/// The reservation that holds the card now (`held`, `expired_candidate` or `held_for_review`),
/// if any. Read-only: it works on a [`ReadLedger::conn`] too.
pub fn current(conn: &Connection, card_id: i64) -> Result<Option<Reservation>, LedgerError> {
    load_where(
        conn,
        &format!("card_id = ?1 AND status IN {ACTIVE_STATUSES}"),
        [card_id],
    )
}

/// Every reservation of the card, oldest generation first. Read-only.
pub fn history(conn: &Connection, card_id: i64) -> Result<Vec<Reservation>, LedgerError> {
    let mut stmt =
        conn.prepare("SELECT generation FROM reservation WHERE card_id = ?1 ORDER BY generation")?;
    let generations: Vec<i64> = stmt
        .query_map([card_id], |r| r.get(0))?
        .collect::<Result<_, _>>()?;
    let mut out = Vec::with_capacity(generations.len());
    for g in generations {
        out.extend(load_reservation(conn, card_id, g)?);
    }
    Ok(out)
}

// ---- writes ------------------------------------------------------------------------------------

pub(super) fn add_event(
    conn: &Connection,
    now: i64,
    actor: &str,
    card_id: i64,
    kind: &str,
    payload: &Value,
) -> Result<(), LedgerError> {
    let key = new_event_key(conn, kind, Some(card_id))?;
    conn.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
        params![key, now, actor, card_id, kind, payload.to_string()],
    )?;
    Ok(())
}

/// The first pair of a pattern in `mine` and a pattern of another card's active reservation
/// that [`crate::scope::overlap`] says could be written under both, in card and pattern order.
pub(super) fn find_overlap(
    conn: &Connection,
    card_id: i64,
    mine: &[String],
) -> Result<Option<(i64, String, String)>, LedgerError> {
    let mut stmt = conn.prepare(&format!(
        "SELECT r.card_id, s.pattern FROM reservation r \
         JOIN reservation_scope s ON s.reservation_id = r.reservation_id \
         WHERE r.status IN {ACTIVE_STATUSES} AND r.card_id <> ?1 \
         ORDER BY r.card_id, s.pattern"
    ))?;
    let theirs: Vec<(i64, String)> = stmt
        .query_map([card_id], |r| Ok((r.get(0)?, r.get(1)?)))?
        .collect::<Result<_, _>>()?;
    for (other, pattern) in theirs {
        for m in mine {
            if crate::scope::overlap(m, &pattern) {
                return Ok(Some((other, m.clone(), pattern)));
            }
        }
    }
    Ok(None)
}

/// Takes a `ready` card: next generation, a `held` reservation for `holder` with a lease of
/// `lease_secs` from the DB clock, the (deduplicated) `write_scope`, the card moved
/// `ready -> active`, and a `reserve` event, in one transaction.
///
/// Rejected, with nothing written: bad input, an empty scope, an unknown card, a card that is
/// not `ready`, a card that already has an active reservation (checked first, so the losers of
/// a race see [`ReserveError::AlreadyHeld`]) and a scope that overlaps another card's active
/// reservation. The generation a rolled-back call would have used is issued again.
pub fn reserve(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    write_scope: &[String],
    lease_secs: i64,
    worktree: Option<&str>,
    branch: Option<&str>,
) -> Result<Reservation, ReserveError> {
    check_holder(holder).map_err(ReserveError::InvalidInput)?;
    check_lease(lease_secs).map_err(ReserveError::InvalidInput)?;
    check_optional("worktree", worktree).map_err(ReserveError::InvalidInput)?;
    check_optional("branch", branch).map_err(ReserveError::InvalidInput)?;
    let scope = check_scope(write_scope)?;
    w.with_write_tx(|tx, now| {
        let Some((state, generation)) = card_row(tx, card_id)? else {
            return Ok(Err(ReserveError::CardNotFound(card_id)));
        };
        if let Some(active) = current(tx, card_id)? {
            return Ok(Err(ReserveError::AlreadyHeld {
                holder: active.holder,
                generation: active.generation,
            }));
        }
        if state != "ready" {
            return Ok(Err(ReserveError::CardNotReady { state }));
        }
        if let Some((other_card, mine, theirs)) = find_overlap(tx, card_id, &scope)? {
            return Ok(Err(ReserveError::ScopeOverlap {
                other_card,
                mine,
                theirs,
            }));
        }
        let next = generation + 1;
        tx.execute(
            "UPDATE card SET generation = ?1, state = 'active' WHERE card_id = ?2",
            params![next, card_id],
        )?;
        tx.execute(
            "INSERT INTO reservation (card_id, generation, holder, status, acquired_at, \
             lease_until, updated_at, worktree, branch) \
             VALUES (?1, ?2, ?3, 'held', ?4, ?5, ?4, ?6, ?7)",
            params![
                card_id,
                next,
                holder,
                now,
                now + lease_secs,
                worktree,
                branch
            ],
        )?;
        let reservation_id = tx.last_insert_rowid();
        for pattern in &scope {
            tx.execute(
                "INSERT INTO reservation_scope (reservation_id, pattern) VALUES (?1, ?2)",
                params![reservation_id, pattern],
            )?;
        }
        add_event(
            tx,
            now,
            holder,
            card_id,
            "reserve",
            &json!({
                "generation": next,
                "holder": holder,
                "lease_until": now + lease_secs,
                "write_scope": scope,
                "worktree": worktree,
                "branch": branch,
            }),
        )?;
        let made = load_reservation(tx, card_id, next)?.ok_or_else(|| {
            LedgerError::Sql("the reservation just inserted cannot be read".into())
        })?;
        Ok(Ok(made))
    })?
}

/// The shared check of renew, release, handoff and fence: the offered `(holder, generation)`
/// must be the current reservation, `held`, and (when `need_live`) with a lease that has not
/// ended. Returns the reservation, or the refusal; a SQL failure is the outer `Err`.
fn evaluate(
    conn: &Connection,
    card_id: i64,
    holder: &str,
    offered: i64,
    now: i64,
    need_live: bool,
) -> Result<Result<Reservation, FenceError>, LedgerError> {
    let Some((_, current)) = card_row(conn, card_id)? else {
        return Ok(Err(FenceError::CardNotFound(card_id)));
    };
    if current == 0 {
        return Ok(Err(FenceError::NoReservation));
    }
    if offered < current {
        return Ok(Err(FenceError::StaleGeneration { offered, current }));
    }
    if offered > current {
        return Ok(Err(FenceError::UnknownGeneration { offered, current }));
    }
    let Some(row) = load_reservation(conn, card_id, current)? else {
        return Ok(Err(FenceError::NoReservation));
    };
    if row.status != "held" {
        return Ok(Err(FenceError::NotHeld { status: row.status }));
    }
    if row.holder != holder {
        return Ok(Err(FenceError::WrongHolder {
            held_by: row.holder,
        }));
    }
    if need_live && row.lease_until <= now {
        return Ok(Err(FenceError::LeaseExpired {
            lease_until: row.lease_until,
            now,
        }));
    }
    Ok(Ok(row))
}

/// Writes the `fence_reject` event of a refusal into the transaction that found it.
fn record_rejection(
    conn: &Connection,
    now: i64,
    op: &str,
    card_id: i64,
    holder: &str,
    offered: i64,
    error: &FenceError,
) -> Result<(), LedgerError> {
    let current = card_row(conn, card_id)?.map_or(0, |(_, g)| g);
    add_event(
        conn,
        now,
        holder,
        card_id,
        EVENT_FENCE_REJECT,
        &json!({
            "op": op,
            "card_id": card_id,
            "holder": holder,
            "offered_generation": offered,
            "current_generation": current,
            "reason": error.code(),
        }),
    )
}

/// Runs `body` after `evaluate` in one write transaction. A refusal is recorded and the
/// transaction commits (only the event is new); a SQL failure rolls everything back.
pub(super) fn fenced<T>(
    w: &mut WriteLedger,
    op: &str,
    card_id: i64,
    holder: &str,
    offered: i64,
    need_live: bool,
    body: impl FnOnce(&Connection, i64, Reservation) -> Result<T, LedgerError>,
) -> Result<T, FenceError> {
    w.with_write_tx(
        |tx, now| match evaluate(tx, card_id, holder, offered, now, need_live)? {
            Ok(row) => Ok(Ok(body(tx, now, row)?)),
            Err(e) => {
                if e.is_recorded() {
                    record_rejection(tx, now, op, card_id, holder, offered, &e)?;
                }
                Ok(Err(e))
            }
        },
    )?
}

/// Extends the lease of the current `held` reservation of `holder` at `generation`: the new
/// `lease_until` is the DB clock plus `lease_secs`, but never earlier than the lease it
/// already has. Needs a live lease.
pub fn renew(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    lease_secs: i64,
) -> Result<Reservation, FenceError> {
    check_holder(holder).map_err(FenceError::InvalidInput)?;
    check_lease(lease_secs).map_err(FenceError::InvalidInput)?;
    fenced(
        w,
        "renew",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| {
            let until = row.lease_until.max(now + lease_secs);
            tx.execute(
            "UPDATE reservation SET lease_until = ?1, updated_at = ?2 WHERE reservation_id = ?3",
            params![until, now, row.reservation_id],
        )?;
            add_event(
                tx,
                now,
                holder,
                card_id,
                "renew",
                &json!({"generation": generation, "holder": holder, "lease_until": until}),
            )?;
            load_reservation(tx, card_id, generation)?
                .ok_or_else(|| LedgerError::Sql("the renewed reservation cannot be read".into()))
        },
    )
}

/// Ends the current reservation of `holder` at `generation` (`released`). A live lease is not
/// needed: a holder whose lease ran out can still give the card back while nobody reclaimed
/// it. The card goes `active -> ready`; a card that has meanwhile moved on (`review`, `hold`,
/// `done` ...) keeps its state.
pub fn release(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
) -> Result<Released, FenceError> {
    check_holder(holder).map_err(FenceError::InvalidInput)?;
    fenced(
        w,
        "release",
        card_id,
        holder,
        generation,
        false,
        |tx, now, row| {
            tx.execute(
            "UPDATE reservation SET status = 'released', updated_at = ?1 WHERE reservation_id = ?2",
            params![now, row.reservation_id],
        )?;
            let (state, _) = card_row(tx, card_id)?.ok_or_else(|| {
                LedgerError::Sql("the card vanished inside the transaction".into())
            })?;
            let card_state = if state == "active" {
                tx.execute(
                    "UPDATE card SET state = 'ready' WHERE card_id = ?1",
                    [card_id],
                )?;
                "ready".to_string()
            } else {
                state
            };
            add_event(
                tx,
                now,
                holder,
                card_id,
                "release",
                &json!({"generation": generation, "holder": holder, "card_state": card_state}),
            )?;
            let reservation = load_reservation(tx, card_id, generation)?.ok_or_else(|| {
                LedgerError::Sql("the released reservation cannot be read".into())
            })?;
            Ok(Released {
                reservation,
                card_state,
            })
        },
    )
}

/// Passes the card from `holder` to `to_holder` in one step: the old reservation becomes
/// `released` and a new `held` one with generation + 1, the same scope, worktree, branch and
/// lease end is inserted for `to_holder`. The card stays as it is (`active`). From then on the
/// old generation is stale for every call. Needs a live lease.
pub fn handoff(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    to_holder: &str,
) -> Result<Reservation, FenceError> {
    check_holder(holder).map_err(FenceError::InvalidInput)?;
    check_holder(to_holder).map_err(|why| FenceError::InvalidInput(format!("to_holder: {why}")))?;
    if to_holder == holder {
        return Err(FenceError::InvalidInput(
            "to_holder is the holder itself".into(),
        ));
    }
    fenced(
        w,
        "handoff",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| {
            tx.execute(
            "UPDATE reservation SET status = 'released', updated_at = ?1 WHERE reservation_id = ?2",
            params![now, row.reservation_id],
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
                    row.lease_until,
                    row.worktree,
                    row.branch
                ],
            )?;
            let new_id = tx.last_insert_rowid();
            tx.execute(
                "INSERT INTO reservation_scope (reservation_id, pattern) \
             SELECT ?1, pattern FROM reservation_scope WHERE reservation_id = ?2",
                params![new_id, row.reservation_id],
            )?;
            add_event(
                tx,
                now,
                holder,
                card_id,
                "handoff",
                &json!({
                    "from": holder,
                    "to": to_holder,
                    "from_generation": generation,
                    "generation": next,
                    "lease_until": row.lease_until,
                }),
            )?;
            load_reservation(tx, card_id, next)?
                .ok_or_else(|| LedgerError::Sql("the new reservation cannot be read".into()))
        },
    )
}

/// The check before a push or a write: `Ok` only when `(holder, generation)` is the current
/// `held` reservation of the card and its lease has not ended (DB clock). Any other answer is a
/// [`FenceError`], and a reservation problem is recorded as a `fence_reject` event that stays
/// even though the call is refused. A successful check writes nothing.
pub fn fence(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
) -> Result<(), FenceError> {
    check_holder(holder).map_err(FenceError::InvalidInput)?;
    fenced(w, "fence", card_id, holder, generation, true, |_, _, _| {
        Ok(())
    })
}

/// The same check as [`fence`] on a read-only connection: nothing is recorded and nothing is
/// created. For a status display; a guard that refuses writes should call [`fence`].
pub fn fence_check(
    r: &ReadLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
) -> Result<(), FenceError> {
    check_holder(holder).map_err(FenceError::InvalidInput)?;
    let now: i64 = r
        .conn()
        .query_row("SELECT unixepoch()", [], |row| row.get(0))
        .map_err(LedgerError::from)?;
    evaluate(r.conn(), card_id, holder, generation, now, true)??;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::contract::{classify, Kind};

    fn scope(n: usize) -> Vec<String> {
        (0..n).map(|i| format!("src/p{i}/")).collect()
    }

    #[test]
    fn lease_bounds_accept_n_and_refuse_n_plus_one() {
        assert!(check_lease(MIN_LEASE_SECS).is_ok());
        assert!(check_lease(MIN_LEASE_SECS - 1).is_err());
        assert!(check_lease(MAX_LEASE_SECS).is_ok());
        assert!(check_lease(MAX_LEASE_SECS + 1).is_err());
        assert!(check_lease(0).is_err());
        assert!(check_lease(-1).is_err());
        assert!(check_lease(i64::MIN).is_err());
        assert!(check_lease(i64::MAX).is_err());
    }

    #[test]
    fn holder_length_accepts_1_and_200_and_refuses_0_and_201() {
        assert!(check_holder("a").is_ok());
        assert!(check_holder(&"a".repeat(MAX_HOLDER_LEN)).is_ok());
        assert!(check_holder("").is_err());
        assert!(check_holder(&"a".repeat(MAX_HOLDER_LEN + 1)).is_err());
        // Characters, not bytes: 200 multi-byte characters are inside the limit.
        assert!(check_holder(&"가".repeat(MAX_HOLDER_LEN)).is_ok());
        assert!(check_holder(&"가".repeat(MAX_HOLDER_LEN + 1)).is_err());
        assert!(check_holder(" lane").is_err());
        assert!(check_holder("lane ").is_err());
        assert!(check_holder("la\nne").is_err());
        assert!(check_holder("la\0ne").is_err());
    }

    #[test]
    fn pattern_count_and_length_accept_n_and_refuse_n_plus_one() {
        assert_eq!(check_scope(&scope(MAX_SCOPE_PATTERNS)).unwrap().len(), 64);
        assert!(matches!(
            check_scope(&scope(MAX_SCOPE_PATTERNS + 1)),
            Err(ReserveError::InvalidInput(_))
        ));
        assert_eq!(check_scope(&[]), Err(ReserveError::EmptyScope));
        let long = "a".repeat(MAX_PATTERN_LEN);
        assert!(check_scope(std::slice::from_ref(&long)).is_ok());
        assert!(matches!(
            check_scope(&[format!("{long}a")]),
            Err(ReserveError::InvalidInput(_))
        ));
        assert!(matches!(
            check_scope(&[String::new()]),
            Err(ReserveError::InvalidInput(_))
        ));
        assert!(matches!(
            check_scope(&["a\nb".to_string()]),
            Err(ReserveError::InvalidInput(_))
        ));
    }

    #[test]
    fn duplicate_patterns_are_dropped_and_count_before_dropping() {
        let dup = vec![
            "src/a".to_string(),
            "src/b".to_string(),
            "src/a".to_string(),
        ];
        assert_eq!(check_scope(&dup).unwrap(), vec!["src/a", "src/b"]);
        // 65 entries that are all the same still exceed the count limit.
        let many = vec!["src/a".to_string(); MAX_SCOPE_PATTERNS + 1];
        assert!(check_scope(&many).is_err());
    }

    #[test]
    fn worktree_and_branch_length_accept_n_and_refuse_n_plus_one() {
        let at = "a".repeat(MAX_FIELD_LEN);
        assert!(check_optional("worktree", Some(&at)).is_ok());
        assert!(check_optional("worktree", Some(&format!("{at}a"))).is_err());
        assert!(check_optional("branch", None).is_ok());
        assert!(check_optional("branch", Some("")).is_err());
    }

    #[test]
    fn every_error_has_a_distinct_code_and_a_kind_the_envelope_can_use() {
        let reserve: Vec<(ReserveError, &str, Kind)> = vec![
            (
                ReserveError::InvalidInput("x".into()),
                "invalid_input",
                Kind::InvalidInput,
            ),
            (ReserveError::EmptyScope, "empty_scope", Kind::InvalidInput),
            (
                ReserveError::CardNotFound(3),
                "card_not_found",
                Kind::NotFound,
            ),
            (
                ReserveError::CardNotReady {
                    state: "draft".into(),
                },
                "card_not_ready",
                Kind::Conflict,
            ),
            (
                ReserveError::AlreadyHeld {
                    holder: "a".into(),
                    generation: 1,
                },
                "already_held",
                Kind::Conflict,
            ),
            (
                ReserveError::ScopeOverlap {
                    other_card: 2,
                    mine: "src/".into(),
                    theirs: "src/a".into(),
                },
                "scope_overlap",
                Kind::Conflict,
            ),
            (ReserveError::Ledger(LedgerError::Busy), "busy", Kind::Io),
        ];
        let mut seen = std::collections::HashSet::new();
        for (error, code, kind) in reserve {
            assert_eq!(error.code(), code);
            assert!(seen.insert(code), "{code} is used twice");
            assert_eq!(classify(&error.to_string()), kind, "{error}");
        }
        let fence: Vec<(FenceError, &str, Kind)> = vec![
            (
                FenceError::InvalidInput("x".into()),
                "invalid_input",
                Kind::InvalidInput,
            ),
            (
                FenceError::CardNotFound(3),
                "card_not_found",
                Kind::NotFound,
            ),
            (FenceError::NoReservation, "no_reservation", Kind::Conflict),
            (
                FenceError::StaleGeneration {
                    offered: 1,
                    current: 2,
                },
                "stale_generation",
                Kind::Conflict,
            ),
            (
                FenceError::UnknownGeneration {
                    offered: 3,
                    current: 2,
                },
                "unknown_generation",
                Kind::Conflict,
            ),
            (
                FenceError::NotHeld {
                    status: "released".into(),
                },
                "not_held",
                Kind::Conflict,
            ),
            (
                FenceError::WrongHolder {
                    held_by: "b".into(),
                },
                "wrong_holder",
                Kind::Conflict,
            ),
            (
                FenceError::LeaseExpired {
                    lease_until: 1,
                    now: 2,
                },
                "lease_expired",
                Kind::Conflict,
            ),
            (FenceError::Ledger(LedgerError::Busy), "busy", Kind::Io),
        ];
        let mut seen = std::collections::HashSet::new();
        for (error, code, kind) in fence {
            assert_eq!(error.code(), code);
            assert!(seen.insert(code), "{code} is used twice");
            assert_eq!(classify(&error.to_string()), kind, "{error}");
        }
    }

    #[test]
    fn only_reservation_problems_are_recorded() {
        assert!(!FenceError::InvalidInput("x".into()).is_recorded());
        assert!(!FenceError::CardNotFound(1).is_recorded());
        assert!(!FenceError::Ledger(LedgerError::Busy).is_recorded());
        assert!(FenceError::NoReservation.is_recorded());
        assert!(FenceError::StaleGeneration {
            offered: 1,
            current: 2
        }
        .is_recorded());
        assert!(FenceError::LeaseExpired {
            lease_until: 1,
            now: 2
        }
        .is_recorded());
    }
}
