//! What `moai card ready | take | release | handoff` do (Issue #3070 K4a, plan comment
//! 6906545561): move a card to `ready`, and take, give back and hand over its reservation.
//!
//! The reservation rules (generations, leases, write scopes, the fence) are
//! `ledger::reservation` (#3068 P3); this module only checks what a user typed, opens the ledger
//! and prints what P3 returned. Every action writes the ledger, none creates it, and every time
//! comes from the database clock inside P3's or this module's one transaction.
//!
//! P3's refusals keep their wording, so `contract::classify` gives them their kinds: `card not
//! found` is `not_found`, `reservation conflict in …` is `conflict`, `invalid reservation input`
//! is `invalid_input`.

use super::command::{actor, open_for_write};
use super::input::{self, CardRef};
use crate::ledger::reservation::{self, Reservation};
use crate::ledger::{self};
use rusqlite::params;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Seconds a reservation lasts when `--lease-secs` is not given (four hours).
pub const DEFAULT_LEASE_SECS: i64 = 14_400;

/// The card's number; a host number (`#12`) needs a lookup, so the writing actions refuse it.
fn card_id(action: &str, reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(|e| format!("card {action} refused: {e}"))? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(format!(
            "card {action} refused: name the card by its C-number (`card show #N` finds it)"
        )),
    }
}

fn reservation_json(r: &Reservation) -> Value {
    json!({
        "name": input::card_name(r.card_id),
        "card_id": r.card_id,
        "generation": r.generation,
        "holder": r.holder,
        "status": r.status,
        "acquired_at": r.acquired_at,
        "lease_until": r.lease_until,
        "worktree": r.worktree,
        "branch": r.branch,
        "write_scope": r.write_scope,
    })
}

/// `card ready C-n`: `draft` or `hold` becomes `ready`, with one `card.ready` event.
pub fn ready(project: &Path, home: &Path, reference: &str, by: Option<&str>) -> Result<Value> {
    let card_id = card_id("ready", reference)?;
    let by = actor(by, "ready")?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let (from, rev) = ledger
        .with_write_tx(|tx, now| {
            let state: Option<String> = tx
                .query_row("SELECT state FROM card WHERE card_id = ?1", [card_id], |r| {
                    r.get(0)
                })
                .map(Some)
                .or_else(|e| match e {
                    rusqlite::Error::QueryReturnedNoRows => Ok(None),
                    other => Err(other),
                })?;
            let Some(from) = state else {
                return Ok(Err(format!("card not found: there is no C-{card_id}")));
            };
            if from != "draft" && from != "hold" {
                return Ok(Err(format!(
                    "card ready refused: C-{card_id} is `{from}`; only a `draft` or `hold` card becomes `ready`"
                )));
            }
            tx.execute(
                "UPDATE card SET state = 'ready', updated_at = ?2 WHERE card_id = ?1",
                params![card_id, now],
            )?;
            let key = ledger::new_event_key(tx, "card.ready", Some(card_id))?;
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, 'card.ready', ?5)",
                params![key, now, by, card_id, json!({"from": from}).to_string()],
            )?;
            let rev: i64 =
                tx.query_row("SELECT rev FROM card WHERE card_id = ?1", [card_id], |r| {
                    r.get(0)
                })?;
            Ok(Ok((from, rev)))
        })
        .map_err(|e| e.to_string())??;
    Ok(json!({
        "action": "ready", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "from": from, "state": "ready", "rev": rev, "events": 1, "writes": true,
    }))
}

/// What `card take` was given.
pub struct TakeRequest<'a> {
    pub holder: &'a str,
    pub lease_secs: Option<i64>,
    pub scope: &'a [String],
    pub worktree: Option<&'a str>,
    pub branch: Option<&'a str>,
}

/// `card take C-n`: reserves a `ready` card for the holder. Without `--scope` the card's own
/// write scope is used.
pub fn take(project: &Path, home: &Path, reference: &str, req: &TakeRequest<'_>) -> Result<Value> {
    let card_id = card_id("take", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let scope: Vec<String> = if req.scope.is_empty() {
        let own: Option<String> = ledger
            .conn()
            .query_row(
                "SELECT write_scope FROM card WHERE card_id = ?1",
                [card_id],
                |r| r.get(0),
            )
            .ok();
        own.and_then(|s| serde_json::from_str(&s).ok())
            .unwrap_or_default()
    } else {
        req.scope.to_vec()
    };
    let r = reservation::reserve(
        &mut ledger,
        card_id,
        req.holder,
        &scope,
        req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS),
        req.worktree,
        req.branch,
    )
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "take", "project_id": id, "reservation": reservation_json(&r),
        "state": "active", "writes": true,
    }))
}

/// `card release C-n`: gives the card back; an `active` card becomes `ready`.
pub fn release(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
) -> Result<Value> {
    let card_id = card_id("release", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let out = reservation::release(&mut ledger, card_id, holder, generation)
        .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "release", "project_id": id, "reservation": reservation_json(&out.reservation),
        "state": out.card_state, "writes": true,
    }))
}

/// `card handoff C-n`: the holder's reservation ends and `to` gets the next generation.
pub fn handoff(
    project: &Path,
    home: &Path,
    reference: &str,
    holder: &str,
    generation: i64,
    to: &str,
) -> Result<Value> {
    let card_id = card_id("handoff", reference)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let r = reservation::handoff(&mut ledger, card_id, holder, generation, to)
        .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "handoff", "project_id": id, "reservation": reservation_json(&r),
        "from_holder": holder, "state": "active", "writes": true,
    }))
}
