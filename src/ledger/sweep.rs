//! The sweep of ended leases (SPEC-20261124-ended-lease-sweep): an `active` card whose `held`
//! reservation ended its lease at least a grace period ago goes back to `ready`, and its
//! reservation ends as `released`.
//!
//! A sweep is not a reclaim. It makes no new holder and no new lease, it does not count against
//! the three reclaims a card may have, and it reads no worktree: the card keeps its generation,
//! and the branch and the worktree of the old reservation stay where they are. One
//! `card.sweep` event and one note on the card name the old holder, generation, lease end,
//! worktree and branch, so the lane that takes the card next reads them first.
//!
//! Only a `held` reservation is swept. A `held_for_review` or `expired_candidate` one, nobody has
//! confirmed whose it is; a card in `review` has nobody who could close it. Both are listed as
//! skipped, never changed. Each card is its own write transaction that reads the reservation
//! again, so a lease that was renewed after the listing is left alone.
//!
//! A second qualification (SPEC-20261119-live-lease-takeover): a card whose lease ended and whose
//! holder wrote nothing to the ledger for [`SILENCE_SECS`] is swept before the grace runs out. A
//! lease that has not ended is never swept, whatever its holder did or did not write.

use super::reservation::{self, add_event};
use super::workflow::LEADER_BY;
use super::{LedgerError, WriteLedger};
use rusqlite::{params, Connection, OptionalExtension, Transaction};
use serde_json::json;

/// Whether a lease that ended at `lease_until` has been over for `grace_secs` at `now`: the
/// second the grace runs out counts as past it, as the lease-end second counts as ended.
pub fn past_grace(lease_until: i64, grace_secs: i64, now: i64) -> bool {
    lease_until.saturating_add(grace_secs) <= now
}

/// How long a holder with an ended lease may write nothing to the ledger before its card is swept,
/// whatever the grace period says (SPEC-20261119-live-lease-takeover REQ-7).
pub const SILENCE_SECS: i64 = 2 * 3600;

/// Whether a holder whose heartbeat was at `last_seen_at` has been silent for [`SILENCE_SECS`] at
/// `now`: the second the window runs out counts as silent.
pub fn silent(last_seen_at: i64, now: i64) -> bool {
    now.saturating_sub(last_seen_at) >= SILENCE_SECS
}

/// What made a card sweepable.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Via {
    /// The lease ended at least the grace period ago.
    Grace,
    /// The lease ended and the holder's heartbeat is [`SILENCE_SECS`] old.
    Heartbeat,
}

impl Via {
    /// The word an event and an answer carry.
    pub fn code(self) -> &'static str {
        match self {
            Via::Grace => "grace",
            Via::Heartbeat => "heartbeat",
        }
    }
}

/// Why a reservation may be swept at `now`, or `None`: only a lease that has ended is ever
/// swept, and then either the grace ran out (whatever the heartbeat) or the holder is silent.
pub fn sweep_via(lease_until: i64, last_seen_at: i64, grace_secs: i64, now: i64) -> Option<Via> {
    if !reservation::lease_ended(lease_until, now) {
        return None;
    }
    if past_grace(lease_until, grace_secs, now) {
        return Some(Via::Grace);
    }
    silent(last_seen_at, now).then_some(Via::Heartbeat)
}

/// A card that was (or, for a dry run, would be) swept.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Swept {
    pub card_id: i64,
    pub holder: String,
    pub generation: i64,
    pub lease_until: i64,
    pub worktree: Option<String>,
    pub branch: Option<String>,
    pub via: Via,
    /// The holder's heartbeat (SPEC-20261119-live-lease-takeover REQ-1) and the seconds since.
    pub last_seen_at: i64,
    pub silent_secs: i64,
}

/// Why a card with an ended lease was left as it is.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SkipReason {
    /// The lease ended less than the grace period ago; `sweepable_at` is the second it runs out.
    WithinGrace { lease_until: i64, sweepable_at: i64 },
    /// The card is in `review`: nobody could close it once it went back to `ready`.
    InReview,
    /// The reservation is not `held` (`held_for_review`, `expired_candidate`, ...).
    NotHeld { status: String },
    /// The card or its reservation changed after the listing: renewed, released, reclaimed,
    /// taken again or closed.
    Changed,
    /// A named card that is not listed: it has no ended lease (a live lease, no reservation, or
    /// a state other than `active` and `review`).
    NoEndedLease,
}

impl SkipReason {
    /// The word an answer carries.
    pub fn code(&self) -> &'static str {
        match self {
            SkipReason::WithinGrace { .. } => "within_grace",
            SkipReason::InReview => "in_review",
            SkipReason::NotHeld { .. } => "not_held",
            SkipReason::Changed => "changed",
            SkipReason::NoEndedLease => "no_ended_lease",
        }
    }
}

/// A card with an ended lease that a sweep left alone.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Skipped {
    pub card_id: i64,
    pub reason: SkipReason,
}

/// What [`sweep_all`] did, or for a dry run would do.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Outcome {
    pub swept: Vec<Swept>,
    pub skipped: Vec<Skipped>,
    pub dry_run: bool,
}

/// Why a sweep was refused; nothing was written.
#[derive(Debug)]
pub enum SweepError {
    /// Only `--by leader` ends reservations of other sessions, as `card withdraw` does.
    LeaderOnly,
    InvalidInput(String),
    Ledger(LedgerError),
}

impl SweepError {
    pub fn code(&self) -> &'static str {
        match self {
            SweepError::LeaderOnly => "leader_only",
            SweepError::InvalidInput(_) => "invalid_input",
            SweepError::Ledger(e) => e.code(),
        }
    }
}

impl std::fmt::Display for SweepError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SweepError::LeaderOnly => write!(
                f,
                "a sweep ends reservations of other sessions: give --by {LEADER_BY}"
            ),
            SweepError::InvalidInput(why) => write!(f, "{why}"),
            SweepError::Ledger(e) => write!(f, "{e}"),
        }
    }
}

impl std::error::Error for SweepError {}

impl From<LedgerError> for SweepError {
    fn from(e: LedgerError) -> Self {
        SweepError::Ledger(e)
    }
}

/// Who signs a sweep: `by` goes into the event, `actor` is the session or name the event is
/// filed under, and `note_author` signs the note.
pub(super) struct Signed<'a> {
    pub by: &'a str,
    pub actor: &'a str,
    pub note_author: &'a str,
}

fn check_grace(grace_secs: i64) -> Result<(), SweepError> {
    if grace_secs <= 0 {
        return Err(SweepError::InvalidInput(format!(
            "the grace period is {grace_secs} seconds, at least 1"
        )));
    }
    Ok(())
}

fn check_by(by: &str) -> Result<(), SweepError> {
    if by != LEADER_BY {
        return Err(SweepError::LeaderOnly);
    }
    Ok(())
}

/// The note a sweep leaves on the card.
fn note_body(card_id: i64, held: &reservation::Reservation, why: &str, by: &str) -> String {
    let ended = crate::baseline::utc_from_epoch(held.lease_until);
    let worktree = held.worktree.as_deref().unwrap_or("none");
    let branch = held.branch.as_deref().unwrap_or("none");
    format!(
        "SWEPT C-{card_id} generation {generation}: the lease of {holder} ended {ended} \
         (lease until {lease_until}), {why}, and {by} returned the card to ready. Branch \
         {branch} and worktree {worktree} were not touched. Read the notes (a RESUME note \
         first) before you take this card.",
        generation = held.generation,
        holder = held.holder,
        lease_until = held.lease_until,
    )
}

/// The clause of the note that says why the card was swept.
fn why_text(via: Via, grace_secs: i64, silent_secs: i64) -> String {
    match via {
        Via::Grace => format!("{} hour(s) of grace had run out", grace_secs / 3600),
        Via::Heartbeat => format!(
            "the holder wrote nothing to the ledger for {} minute(s)",
            silent_secs / 60
        ),
    }
}

/// The sweep of one card inside the caller's write transaction: `Ok(Ok)` when the reservation
/// ended and the card went back to `ready`, `Ok(Err(reason))` when it was left as it is, an
/// `Err` when the ledger failed. `expect_lease_until` is the lease end the caller listed: a
/// reservation whose lease end is another one was renewed since and is left alone.
pub(super) fn sweep_in_tx(
    tx: &Transaction<'_>,
    now: i64,
    card_id: i64,
    expect_lease_until: Option<i64>,
    grace_secs: i64,
    signed: &Signed<'_>,
) -> Result<Result<Swept, SkipReason>, LedgerError> {
    let state: Option<String> = tx
        .query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .optional()?;
    let Some(state) = state else {
        return Ok(Err(SkipReason::Changed));
    };
    if state == "review" {
        return Ok(Err(SkipReason::InReview));
    }
    if state != "active" {
        return Ok(Err(SkipReason::Changed));
    }
    let Some(held) = reservation::current(tx, card_id)? else {
        return Ok(Err(SkipReason::Changed));
    };
    if held.status != "held" {
        return Ok(Err(SkipReason::NotHeld {
            status: held.status,
        }));
    }
    if expect_lease_until.is_some_and(|read| read != held.lease_until) {
        return Ok(Err(SkipReason::Changed));
    }
    let last_seen_at = reservation::heartbeat_at(tx, card_id, &held.holder, held.updated_at)?;
    let Some(via) = sweep_via(held.lease_until, last_seen_at, grace_secs, now) else {
        return Ok(Err(SkipReason::WithinGrace {
            lease_until: held.lease_until,
            sweepable_at: held.lease_until.saturating_add(grace_secs),
        }));
    };
    let silent_secs = now.saturating_sub(last_seen_at);
    tx.execute(
        "UPDATE reservation SET status = 'released', updated_at = ?1 WHERE reservation_id = ?2",
        params![now, held.reservation_id],
    )?;
    tx.execute(
        "UPDATE card SET state = 'ready', updated_at = ?2 WHERE card_id = ?1",
        params![card_id, now],
    )?;
    add_event(
        tx,
        now,
        signed.actor,
        card_id,
        "card.sweep",
        &json!({
            "holder": held.holder,
            "generation": held.generation,
            "lease_until": held.lease_until,
            "grace_secs": grace_secs,
            "via": via.code(),
            "last_seen_at": last_seen_at,
            "silent_secs": silent_secs,
            "by": signed.by,
            "worktree": held.worktree,
            "branch": held.branch,
        }),
    )?;
    let body = note_body(
        card_id,
        &held,
        &why_text(via, grace_secs, silent_secs),
        signed.by,
    );
    let reference: String = tx.query_row(
        "SELECT 'note:' || ?1 || ':' || lower(hex(randomblob(16)))",
        [card_id],
        |r| r.get(0),
    )?;
    tx.execute(
        "INSERT INTO card_note (card_id, kind, author, body, ref, created_at) \
         VALUES (?1, 'report', ?2, ?3, ?4, ?5)",
        params![card_id, signed.note_author, body, reference, now],
    )?;
    let note_id = tx.last_insert_rowid();
    add_event(
        tx,
        now,
        signed.actor,
        card_id,
        "card.note",
        &json!({
            "note_id": note_id, "kind": "report", "ref": reference, "bytes": body.len(),
            "author": signed.note_author, "session": signed.actor,
        }),
    )?;
    Ok(Ok(Swept {
        card_id,
        holder: held.holder,
        generation: held.generation,
        lease_until: held.lease_until,
        worktree: held.worktree,
        branch: held.branch,
        via,
        last_seen_at,
        silent_secs,
    }))
}

/// One row of the listing: a card whose `held` or recovered reservation has an ended lease.
struct Listed {
    card_id: i64,
    state: String,
    status: String,
    holder: String,
    generation: i64,
    lease_until: i64,
    updated_at: i64,
    worktree: Option<String>,
    branch: Option<String>,
}

/// The `active` and `review` cards whose current reservation has an ended lease at `now`, in
/// card order. A live lease is never listed.
fn list_ended(conn: &Connection, now: i64) -> Result<Vec<Listed>, LedgerError> {
    let mut stmt = conn.prepare(
        "SELECT c.card_id, c.state, r.status, r.holder, r.generation, r.lease_until, \
                r.updated_at, r.worktree, r.branch \
         FROM card c JOIN reservation r ON r.card_id = c.card_id AND r.generation = c.generation \
         WHERE c.state IN ('active', 'review') \
           AND r.status IN ('held', 'expired_candidate', 'held_for_review') \
           AND r.lease_until <= ?1 \
         ORDER BY c.card_id",
    )?;
    let rows = stmt
        .query_map([now], |r| {
            Ok(Listed {
                card_id: r.get(0)?,
                state: r.get(1)?,
                status: r.get(2)?,
                holder: r.get(3)?,
                generation: r.get(4)?,
                lease_until: r.get(5)?,
                updated_at: r.get(6)?,
                worktree: r.get(7)?,
                branch: r.get(8)?,
            })
        })?
        .collect::<Result<Vec<_>, _>>()?;
    Ok(rows)
}

/// The sweep of every card with an ended lease past `grace_secs`: one write transaction per card,
/// so one failing card does not hold back the rest. With `dry_run` nothing is written and the
/// answer lists the cards that would be swept. `by` must be `leader`.
pub fn sweep_all(
    w: &mut WriteLedger,
    grace_secs: i64,
    by: &str,
    session: Option<&str>,
    dry_run: bool,
) -> Result<Outcome, SweepError> {
    check_by(by)?;
    check_grace(grace_secs)?;
    let now: i64 = w
        .conn()
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .map_err(LedgerError::from)?;
    let listed = list_ended(w.conn(), now)?;
    let mut out = Outcome {
        swept: Vec::new(),
        skipped: Vec::new(),
        dry_run,
    };
    let signed = Signed {
        by,
        actor: session.unwrap_or(by),
        note_author: by,
    };
    for row in listed {
        let skip = |reason| Skipped {
            card_id: row.card_id,
            reason,
        };
        if row.state == "review" {
            out.skipped.push(skip(SkipReason::InReview));
            continue;
        }
        if row.status != "held" {
            out.skipped.push(skip(SkipReason::NotHeld {
                status: row.status.clone(),
            }));
            continue;
        }
        let last_seen_at =
            reservation::heartbeat_at(w.conn(), row.card_id, &row.holder, row.updated_at)?;
        let Some(via) = sweep_via(row.lease_until, last_seen_at, grace_secs, now) else {
            out.skipped.push(skip(SkipReason::WithinGrace {
                lease_until: row.lease_until,
                sweepable_at: row.lease_until.saturating_add(grace_secs),
            }));
            continue;
        };
        if dry_run {
            out.swept.push(Swept {
                card_id: row.card_id,
                holder: row.holder,
                generation: row.generation,
                lease_until: row.lease_until,
                worktree: row.worktree,
                branch: row.branch,
                via,
                last_seen_at,
                silent_secs: now.saturating_sub(last_seen_at),
            });
            continue;
        }
        let one = w
            .with_write_tx(|tx, now| {
                sweep_in_tx(
                    tx,
                    now,
                    row.card_id,
                    Some(row.lease_until),
                    grace_secs,
                    &signed,
                )
            })
            .map_err(SweepError::from)?;
        match one {
            Ok(swept) => out.swept.push(swept),
            Err(reason) => out.skipped.push(skip(reason)),
        }
    }
    Ok(out)
}

/// The sweep of one card: `Ok(Err(reason))` when it was left as it is. `expect_lease_until` is
/// the lease end the caller read; a reservation renewed since is skipped as
/// [`SkipReason::Changed`]. `by` must be `leader`.
pub fn sweep_card(
    w: &mut WriteLedger,
    card_id: i64,
    expect_lease_until: Option<i64>,
    grace_secs: i64,
    by: &str,
    session: Option<&str>,
) -> Result<Result<Swept, SkipReason>, SweepError> {
    check_by(by)?;
    check_grace(grace_secs)?;
    let signed = Signed {
        by,
        actor: session.unwrap_or(by),
        note_author: by,
    };
    w.with_write_tx(|tx, now| {
        sweep_in_tx(tx, now, card_id, expect_lease_until, grace_secs, &signed)
    })
    .map_err(SweepError::from)
}

#[cfg(test)]
mod tests {
    use super::*;

    const NOW: i64 = 1_000_000;
    const GRACE: i64 = 24 * 3600;

    #[test]
    fn heartbeat_window_is_silent_at_two_hours_and_not_one_second_before() {
        assert!(!silent(NOW - SILENCE_SECS + 1, NOW), "one second short");
        assert!(silent(NOW - SILENCE_SECS, NOW), "exactly the window");
        assert!(silent(NOW - SILENCE_SECS - 1, NOW), "one second over");
    }

    #[test]
    fn heartbeat_window_never_overflows_and_a_future_heartbeat_is_not_silent() {
        assert!(!silent(i64::MAX, NOW));
        assert!(silent(i64::MIN, NOW));
    }

    #[test]
    fn heartbeat_window_a_live_lease_is_never_sweepable_whatever_the_heartbeat() {
        let ancient = NOW - 100 * 3600;
        assert_eq!(sweep_via(NOW + 1, ancient, GRACE, NOW), None);
        assert_eq!(sweep_via(NOW + 3600, i64::MIN, GRACE, NOW), None);
    }

    #[test]
    fn heartbeat_window_the_lease_end_second_counts_as_ended() {
        let ancient = NOW - 100 * 3600;
        assert_eq!(sweep_via(NOW, ancient, GRACE, NOW), Some(Via::Heartbeat));
        assert_eq!(sweep_via(NOW + 1, ancient, GRACE, NOW), None);
    }

    #[test]
    fn heartbeat_window_a_fresh_heartbeat_keeps_an_ended_lease_inside_the_grace() {
        let ended = NOW - 3 * 3600;
        assert_eq!(sweep_via(ended, NOW - 60, GRACE, NOW), None);
        assert_eq!(
            sweep_via(ended, NOW - SILENCE_SECS, GRACE, NOW),
            Some(Via::Heartbeat)
        );
        assert_eq!(sweep_via(ended, NOW - SILENCE_SECS + 1, GRACE, NOW), None);
    }

    #[test]
    fn heartbeat_window_the_grace_still_sweeps_whatever_the_heartbeat_says() {
        let ended = NOW - GRACE;
        assert_eq!(sweep_via(ended, NOW, GRACE, NOW), Some(Via::Grace));
        assert_eq!(sweep_via(ended, NOW - 1, GRACE, NOW), Some(Via::Grace));
        assert_eq!(sweep_via(ended + 1, NOW, GRACE, NOW), None);
    }
}
