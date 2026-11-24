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

/// A card that was (or, for a dry run, would be) swept.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Swept {
    pub card_id: i64,
    pub holder: String,
    pub generation: i64,
    pub lease_until: i64,
    pub worktree: Option<String>,
    pub branch: Option<String>,
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
fn note_body(card_id: i64, held: &reservation::Reservation, grace_secs: i64, by: &str) -> String {
    let ended = crate::baseline::utc_from_epoch(held.lease_until);
    let worktree = held.worktree.as_deref().unwrap_or("none");
    let branch = held.branch.as_deref().unwrap_or("none");
    format!(
        "SWEPT C-{card_id} generation {generation}: the lease of {holder} ended {ended} \
         (lease until {lease_until}), {hours} hour(s) of grace had run out, and {by} returned the \
         card to ready. Branch {branch} and worktree {worktree} were not touched. Read the notes \
         (a RESUME note first) before you take this card.",
        generation = held.generation,
        holder = held.holder,
        lease_until = held.lease_until,
        hours = grace_secs / 3600,
    )
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
    if !past_grace(held.lease_until, grace_secs, now) {
        return Ok(Err(SkipReason::WithinGrace {
            lease_until: held.lease_until,
            sweepable_at: held.lease_until.saturating_add(grace_secs),
        }));
    }
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
            "by": signed.by,
            "worktree": held.worktree,
            "branch": held.branch,
        }),
    )?;
    let body = note_body(card_id, &held, grace_secs, signed.by);
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
    worktree: Option<String>,
    branch: Option<String>,
}

/// The `active` and `review` cards whose current reservation has an ended lease at `now`, in
/// card order. A live lease is never listed.
fn list_ended(conn: &Connection, now: i64) -> Result<Vec<Listed>, LedgerError> {
    let mut stmt = conn.prepare(
        "SELECT c.card_id, c.state, r.status, r.holder, r.generation, r.lease_until, \
                r.worktree, r.branch \
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
                worktree: r.get(6)?,
                branch: r.get(7)?,
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
        if !past_grace(row.lease_until, grace_secs, now) {
            out.skipped.push(skip(SkipReason::WithinGrace {
                lease_until: row.lease_until,
                sweepable_at: row.lease_until.saturating_add(grace_secs),
            }));
            continue;
        }
        if dry_run {
            out.swept.push(Swept {
                card_id: row.card_id,
                holder: row.holder,
                generation: row.generation,
                lease_until: row.lease_until,
                worktree: row.worktree,
                branch: row.branch,
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
