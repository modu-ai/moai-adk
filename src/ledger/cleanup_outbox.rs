//! The cleanup outbox (SPEC-20261123-cleanup-outbox): the ledger rows a closed card leaves for
//! the executor in [`crate::card::cleanup`].
//!
//! The outbox is made of `event` rows, not a table of its own: an event is append-only and has a
//! unique key, and a new table would need a schema version bump and a migration of every ledger
//! (the version check is an exact match). Five kinds, all with the card's id:
//!
//! - `cleanup.requested`: written by the transaction that closes the card ([`record_request`]),
//!   naming the worktrees and branches the card's reservations named and the head it landed;
//! - `cleanup.started`: a claim of one request with a lease ([`claim`]);
//! - `cleanup.step`: what one attempt did to one target;
//! - `cleanup.released`: an attempt that ended without a result gives its claim back;
//! - `cleanup.result`: the request is over. A request is pending while no result names it.
//!
//! Every write after the request is one transaction of its own that first reads what it depends
//! on, so two executors cannot both claim a request and a result is written once.

use super::open::WriteLedger;
use super::reservation;
use super::LedgerError;
use rusqlite::{params, Connection, OptionalExtension, Transaction};
use serde_json::{json, Value};

pub const REQUESTED: &str = "cleanup.requested";
pub const STARTED: &str = "cleanup.started";
pub const STEP: &str = "cleanup.step";
pub const RELEASED: &str = "cleanup.released";
pub const RESULT: &str = "cleanup.result";

/// A request nobody has finished.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Pending {
    /// The event id of the `cleanup.requested` row: the request's name.
    pub request: i64,
    pub card_id: i64,
    /// Which command closed the card: `done` or `finish`.
    pub via: String,
    /// The worktrees the card's reservations named, as stored.
    pub worktrees: Vec<String>,
    /// The branches the card's reservations named, as short names.
    pub branches: Vec<String>,
    /// The head the card landed, when it landed one.
    pub head: Option<String>,
}

/// The branch name without `refs/heads/`.
fn short(branch: &str) -> String {
    branch
        .strip_prefix("refs/heads/")
        .unwrap_or(branch)
        .to_string()
}

/// Writes the request of a card that closes now, in the closing transaction `tx`. What it names
/// is read from the reservation rows of the card, so a target no reservation named is never in a
/// request. The head is the newest `card.landed` head of the card, when one names a head.
pub(super) fn record_request(
    tx: &Connection,
    now: i64,
    actor: &str,
    card_id: i64,
    via: &str,
) -> Result<(), LedgerError> {
    let worktrees = reservation::worktrees_of_card(tx, card_id)?;
    let mut st = tx.prepare(
        "SELECT DISTINCT branch FROM reservation \
         WHERE card_id = ?1 AND branch IS NOT NULL ORDER BY branch",
    )?;
    let mut branches: Vec<String> = st
        .query_map([card_id], |r| r.get::<_, String>(0))?
        .collect::<Result<_, _>>()?;
    drop(st);
    for b in &mut branches {
        *b = short(b);
    }
    branches.sort();
    branches.dedup();
    let head: Option<String> = tx
        .query_row(
            "SELECT json_extract(payload, '$.head') FROM event \
             WHERE card_id = ?1 AND kind = 'card.landed' AND json_type(payload, '$.head') = 'text' \
             ORDER BY event_id DESC LIMIT 1",
            [card_id],
            |r| r.get(0),
        )
        .optional()?;
    reservation::add_event(
        tx,
        now,
        actor,
        card_id,
        REQUESTED,
        &json!({"via": via, "worktrees": worktrees, "branches": branches, "head": head}),
    )
}

/// Every request that has no `cleanup.result`, oldest first. Read-only.
pub fn pending(conn: &Connection) -> Result<Vec<Pending>, LedgerError> {
    let mut st = conn.prepare(
        "SELECT e.event_id, e.card_id, e.payload FROM event e \
         WHERE e.kind = 'cleanup.requested' AND e.card_id IS NOT NULL \
           AND NOT EXISTS (SELECT 1 FROM event r WHERE r.kind = 'cleanup.result' \
                            AND r.card_id = e.card_id \
                            AND json_extract(r.payload, '$.request') = e.event_id) \
         ORDER BY e.event_id",
    )?;
    let rows: Vec<(i64, i64, String)> = st
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)))?
        .collect::<Result<_, _>>()?;
    let strings = |v: &Value, key: &str| -> Vec<String> {
        v[key]
            .as_array()
            .map(|a| {
                a.iter()
                    .filter_map(|x| x.as_str().map(str::to_string))
                    .collect()
            })
            .unwrap_or_default()
    };
    Ok(rows
        .into_iter()
        .map(|(request, card_id, payload)| {
            let v: Value = serde_json::from_str(&payload).unwrap_or(Value::Null);
            Pending {
                request,
                card_id,
                via: v["via"].as_str().unwrap_or("").to_string(),
                worktrees: strings(&v, "worktrees"),
                branches: strings(&v, "branches"),
                head: v["head"].as_str().map(str::to_string),
            }
        })
        .collect())
}

/// What a claim attempt found.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Claim {
    /// This caller holds the request now, as attempt `attempt`. `abandoned` is the earlier attempt
    /// whose lease ended with no release and no result, when this claim found one.
    Taken {
        attempt: i64,
        abandoned: Option<i64>,
    },
    /// Another attempt holds a lease that has not ended.
    Busy,
    /// The request already has a result.
    Over,
}

fn result_exists(tx: &Transaction<'_>, request: i64, card_id: i64) -> Result<bool, LedgerError> {
    Ok(tx.query_row(
        "SELECT EXISTS (SELECT 1 FROM event WHERE kind = 'cleanup.result' AND card_id = ?1 \
                         AND json_extract(payload, '$.request') = ?2)",
        params![card_id, request],
        |r| r.get(0),
    )?)
}

/// The newest attempt of the request: its number, its holder and the end of its lease.
fn newest_attempt(
    tx: &Transaction<'_>,
    request: i64,
    card_id: i64,
) -> Result<Option<(i64, String, i64)>, LedgerError> {
    Ok(tx
        .query_row(
            "SELECT json_extract(payload, '$.attempt'), json_extract(payload, '$.holder'), \
                    json_extract(payload, '$.lease_until') \
             FROM event WHERE kind = 'cleanup.started' AND card_id = ?1 \
               AND json_extract(payload, '$.request') = ?2 \
             ORDER BY json_extract(payload, '$.attempt') DESC LIMIT 1",
            params![card_id, request],
            |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?)),
        )
        .optional()?)
}

/// True when the attempt gave its claim back.
fn released(
    tx: &Transaction<'_>,
    request: i64,
    card_id: i64,
    attempt: i64,
) -> Result<bool, LedgerError> {
    Ok(tx.query_row(
        "SELECT EXISTS (SELECT 1 FROM event WHERE kind = 'cleanup.released' \
                         AND card_id = ?1 AND json_extract(payload, '$.request') = ?2 \
                         AND json_extract(payload, '$.attempt') = ?3)",
        params![card_id, request, attempt],
        |r| r.get(0),
    )?)
}

/// Inserts one event with a natural key; a second insert of the same key adds nothing.
fn put(
    tx: &Transaction<'_>,
    now: i64,
    actor: &str,
    card_id: i64,
    kind: &str,
    key: &str,
    payload: &Value,
) -> Result<bool, LedgerError> {
    Ok(tx.execute(
        "INSERT OR IGNORE INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
        params![key, now, actor, card_id, kind, payload.to_string()],
    )? > 0)
}

/// Takes the request for `holder` for `lease_secs` seconds of the database clock, in one
/// `BEGIN IMMEDIATE` transaction: the result and the newest attempt are read inside it, so of two
/// callers that arrive together one gets the claim and the other reads it.
pub fn claim(
    w: &mut WriteLedger,
    request: i64,
    card_id: i64,
    holder: &str,
    lease_secs: i64,
) -> Result<Claim, LedgerError> {
    w.with_write_tx(|tx, now| {
        if result_exists(tx, request, card_id)? {
            return Ok(Claim::Over);
        }
        // The newest attempt: its number, its lease and whether it gave the claim back.
        let newest = newest_attempt(tx, request, card_id)?;
        let mut abandoned = None;
        let attempt = match newest {
            None => 1,
            Some((attempt, _, lease_until)) => {
                if !released(tx, request, card_id, attempt)? {
                    if lease_until > now {
                        return Ok(Claim::Busy);
                    }
                    // Its lease is over and it wrote no end: a run that died. The dead run could
                    // not leave a line, so this claim does, once (D-2 of the follow-up SPEC).
                    abandoned = Some(attempt);
                }
                attempt + 1
            }
        };
        if let (Some(dead), Some((_, holder_of_dead, lease_until))) = (abandoned, newest) {
            put(
                tx,
                now,
                holder,
                card_id,
                STEP,
                &format!("{STEP}:{request}:{dead}:attempt:abandoned"),
                &json!({
                    "request": request, "attempt": dead, "kind": "attempt",
                    "target": format!("attempt {dead}"), "status": "abandoned",
                    "reasons": ["lease_ended"],
                    "detail": format!(
                        "held by {holder_of_dead} until {lease_until} (database clock) and wrote no end"
                    ),
                }),
            )?;
        }
        put(
            tx,
            now,
            holder,
            card_id,
            STARTED,
            &format!("{STARTED}:{request}:{attempt}"),
            &json!({
                "request": request, "attempt": attempt, "holder": holder,
                "lease_until": now + lease_secs,
            }),
        )?;
        Ok(Claim::Taken { attempt, abandoned })
    })
}

/// What one attempt did to one target.
#[derive(Debug, Clone, PartialEq)]
pub struct Step {
    /// `worktree`, `branch`, `queue` (a ref of the drained reference-transaction queue) or
    /// `attempt` (the line a claim leaves for an earlier attempt that died).
    pub kind: &'static str,
    pub target: String,
    pub status: String,
    pub reasons: Vec<String>,
    pub detail: Option<String>,
}

/// How an attempt ends.
#[derive(Debug, Clone, PartialEq)]
pub enum Outcome {
    /// The request is over, with this status (`settled`, `settled_with_kept` or `skipped`).
    Result(String),
    /// The attempt gives its claim back and the request stays pending.
    Release,
}

/// What [`finish_attempt`] did.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Finish {
    /// The steps and the end of the attempt were written.
    Written,
    /// The request already has a result: nothing was written.
    Over,
    /// The attempt is not live (its lease ended, a newer attempt started or it gave the claim
    /// back): nothing was written.
    LeaseEnded,
}

/// Writes the steps of an attempt and its end in one transaction. A result is written once: when
/// the request already has one, nothing is written and [`Finish::Over`] comes back.
pub fn finish_attempt(
    w: &mut WriteLedger,
    request: i64,
    card_id: i64,
    attempt: i64,
    holder: &str,
    steps: &[Step],
    outcome: &Outcome,
) -> Result<Finish, LedgerError> {
    w.with_write_tx(|tx, now| {
        if result_exists(tx, request, card_id)? {
            return Ok(Finish::Over);
        }
        // Only the newest attempt, still inside its lease and still holding the claim, writes.
        let live = match newest_attempt(tx, request, card_id)? {
            Some((newest, owner, lease_until)) => {
                newest == attempt
                    && owner == holder
                    && lease_until > now
                    && !released(tx, request, card_id, attempt)?
            }
            None => false,
        };
        if !live {
            return Ok(Finish::LeaseEnded);
        }
        for step in steps {
            put(
                tx,
                now,
                holder,
                card_id,
                STEP,
                &format!("{STEP}:{request}:{attempt}:{}:{}", step.kind, step.target),
                &json!({
                    "request": request, "attempt": attempt, "kind": step.kind,
                    "target": step.target, "status": step.status, "reasons": step.reasons,
                    "detail": step.detail,
                }),
            )?;
        }
        match outcome {
            Outcome::Result(status) => {
                put(
                    tx,
                    now,
                    holder,
                    card_id,
                    RESULT,
                    &format!("{RESULT}:{request}"),
                    &json!({"request": request, "attempt": attempt, "status": status}),
                )?;
            }
            Outcome::Release => {
                put(
                    tx,
                    now,
                    holder,
                    card_id,
                    RELEASED,
                    &format!("{RELEASED}:{request}:{attempt}"),
                    &json!({"request": request, "attempt": attempt}),
                )?;
            }
        }
        Ok(Finish::Written)
    })
}
