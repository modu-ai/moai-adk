//! What `moai card ready | take | release | handoff` do (Issue #3070 K4a, plan comment
//! 6906545561): move a card to `ready`, and take, give back and hand over its reservation.
//!
//! The holder is never typed: `take`, `release` and `handoff` act as **this session**, whose id
//! is read from the host's environment ([`session_current`](crate::session_current): the
//! variable Claude Code or Codex sets for the session it starts). A caller that only knows
//! another holder's name cannot release or hand over that holder's card (#3115 verdict N1).
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
use crate::session_current;
use rusqlite::params;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Seconds a reservation lasts when `--lease-secs` is not given (four hours).
pub const DEFAULT_LEASE_SECS: i64 = 14_400;

/// This session's id, the holder of every reservation it takes. `get` reads one environment
/// variable. A session without an id (no host variable, or a malformed one) holds nothing.
pub fn session_holder(action: &str, get: impl Fn(&str) -> Option<String>) -> Result<String> {
    let found = session_current::resolve(get);
    match found["source_session_id"].as_str() {
        Some(id) => Ok(id.to_string()),
        None => Err(format!(
            "card {action} refused: this session has no id, and a holder is always the session's own id; \
a host sets {} or {}",
            session_current::SOURCES[0],
            session_current::SOURCES[1]
        )),
    }
}

/// `--to` of `handoff`: another session's id, in the shape [`session_current`] accepts.
pub fn check_receiver(to: &str) -> Result<()> {
    if session_current::well_formed(to) {
        Ok(())
    } else {
        Err("card handoff refused: --to must be a session id (hex digits and `-`, 8 to 64 characters)"
            .to_string())
    }
}

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
    check_receiver(to)?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let r = reservation::handoff(&mut ledger, card_id, holder, generation, to)
        .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "handoff", "project_id": id, "reservation": reservation_json(&r),
        "from_holder": holder, "state": "active", "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    const A: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";
    const B: &str = "4d5550e0-88b3-40ca-8c46-07425eedc812";

    fn env(pairs: &[(&str, &str)]) -> impl Fn(&str) -> Option<String> {
        let m: HashMap<String, String> = pairs
            .iter()
            .map(|(k, v)| (k.to_string(), v.to_string()))
            .collect();
        move |k| m.get(k).cloned()
    }

    #[test]
    fn the_holder_is_the_session_id_the_host_set() {
        assert_eq!(
            session_holder("take", env(&[("CLAUDE_CODE_SESSION_ID", A)])).unwrap(),
            A
        );
        // The Claude variable wins over the Codex one.
        assert_eq!(
            session_holder(
                "take",
                env(&[("CODEX_THREAD_ID", B), ("CLAUDE_CODE_SESSION_ID", A)])
            )
            .unwrap(),
            A
        );
        assert_eq!(
            session_holder("take", env(&[("CODEX_THREAD_ID", B)])).unwrap(),
            B
        );
    }

    #[test]
    fn a_session_without_a_usable_id_holds_nothing() {
        for pairs in [
            vec![],
            vec![("CLAUDE_CODE_SESSION_ID", "")],
            vec![("CLAUDE_CODE_SESSION_ID", "agent-7")],
            vec![("CLAUDE_CODE_SESSION_ID", "short")],
        ] {
            let err = session_holder("take", env(&pairs)).unwrap_err();
            assert!(err.starts_with("card take refused: "), "{err}");
            assert!(err.contains("CLAUDE_CODE_SESSION_ID"), "{err}");
        }
    }

    #[test]
    fn the_receiver_is_a_session_id_shape_with_hard_edges() {
        assert!(check_receiver(A).is_ok());
        assert!(check_receiver(&"a".repeat(8)).is_ok());
        assert!(check_receiver(&"a".repeat(64)).is_ok());
        for bad in [
            "",
            "agent-b",
            &"a".repeat(7),
            &"a".repeat(65),
            "abcdefgh ",
            "abcdefg\n",
        ] {
            assert!(check_receiver(bad).is_err(), "{bad:?}");
        }
    }
}
