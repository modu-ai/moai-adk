//! What `moai card ready | scope | take | release | handoff` do (Issue #3070 K4a, plan comment
//! 6906545561; `scope` from card C-53): move a card to `ready`, set the write scope of a card
//! nobody holds, and take, give back and hand over its reservation.
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
use super::early_reclaim;
use super::input::{self, CardRef};
use super::spec_link;
use super::take_scope;
use crate::ledger::reservation::{self, Reservation};
use crate::ledger::{self};
use crate::session_current;
use rusqlite::{params, OptionalExtension};
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
pub(super) fn card_id(action: &str, reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(|e| format!("card {action} refused: {e}"))? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(format!(
            "card {action} refused: name the card by its C-number (`card show #N` finds it)"
        )),
    }
}

/// The `overlaps` answer of a take and a widening: each pair of a pattern of this card and a
/// pattern of another card's active reservation, with that card and its holder
/// (SPEC-20261125-card-no-file-locks). Nothing refuses on it.
pub(super) fn overlaps_json(list: &[reservation::Overlap]) -> Value {
    Value::Array(
        list.iter()
            .map(|o| {
                json!({
                    "card": input::card_name(o.other_card),
                    "holder": o.holder,
                    "mine": o.mine,
                    "theirs": o.theirs,
                    "lease_until": o.lease_until,
                    "status": o.status,
                })
            })
            .collect(),
    )
}

pub(super) fn reservation_json(r: &Reservation) -> Value {
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

/// Names what a fill of an empty reservation reserved (`scope_filled`) and left out
/// (`scope_left_out`, each with its reason) in an answer; nothing is added when no fill was made
/// (SPEC-20261125-review-lease-hold REQ-5, REQ-6).
pub(super) fn add_scope_fill(answer: &mut Value, fill: Option<&reservation::ScopeFill>) {
    let Some(fill) = fill else {
        return;
    };
    answer["scope_filled"] = json!(fill.filled);
    answer["scope_left_out"] = json!(fill
        .left_out
        .iter()
        .map(|l| json!({"path": l.path, "reason": l.reason}))
        .collect::<Vec<_>>());
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
            // The state is checked first, so a taken (`active`) card is refused for its state; only
            // a `draft` or `hold` card can still carry a live reservation worth naming.
            if from != "draft" && from != "hold" {
                return Ok(Err(format!(
                    "card ready refused: C-{card_id} is `{from}`; only a `draft` or `hold` card becomes `ready`"
                )));
            }
            let live: i64 = tx.query_row(
                "SELECT count(*) FROM reservation WHERE card_id = ?1 \
                 AND status IN ('held','expired_candidate','held_for_review')",
                [card_id],
                |r| r.get(0),
            )?;
            if live > 0 {
                return Ok(Err(format!(
                    "card ready refused: C-{card_id} has a live reservation; release or hand it over first"
                )));
            }
            if let Some(why) = no_criteria_refusal(tx, "ready", card_id)? {
                return Ok(Err(why));
            }
            // A withdrawn or folded card left its SPEC room for another open card; reviving it
            // without a relation must not make a second one (SPEC-20261124-card-help-gates REQ-8).
            if from == "hold" {
                if let Some(why) = spec_link::ready_conflict(tx, card_id)? {
                    return Ok(Err(why));
                }
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

/// The refusal for a card that has no counted criterion: no `open` or `met` row (a `withdrawn` row
/// does not count), or `None` when it has one. A card without one can never be closed, because
/// `card_done_guard` and `card finish` both need a met criterion. A card that does not exist has
/// none either, so the callers run this after the checks that own that case. `action` is `ready`
/// or `take`.
fn no_criteria_refusal(
    conn: &rusqlite::Connection,
    action: &str,
    card_id: i64,
) -> rusqlite::Result<Option<String>> {
    if crate::ledger::workflow::counted_criteria(conn, card_id)? > 0 {
        return Ok(None);
    }
    let name = input::card_name(card_id);
    Ok(Some(format!(
        "card {action} refused: {name} has no acceptance criterion that is not withdrawn, so it \
         could never be closed; add one with `moai card criterion-add {name} --ac <text>`; nothing \
         was changed"
    )))
}

/// The reservation states that hold a card (the ones `card ready` also refuses).
const LIVE_RESERVATION: &str = "status IN ('held','expired_candidate','held_for_review')";

/// A write scope a user typed, checked as `add --scope` checks it, and within what
/// `reservation::reserve` stores, so a card that gets it can also be taken with it. `add`,
/// `scope` and `take --scope` all use it.
pub(super) fn checked_scope(action: &str, entries: &[String]) -> Result<Vec<String>> {
    // On the raw text: `trim` strips U+2028 and U+2029 at the ends, so after it only an inner
    // one would be seen (SPEC-20261119-card-scope-widen-followup REQ-1).
    if let Some(i) = entries.iter().position(|e| has_line_separator(e)) {
        return Err(format!(
            "card {action} refused: scope entry {} has a line separator (U+2028 or U+2029)",
            i + 1
        ));
    }
    // One `--scope` names one path and a glob here has no braces, so a comma is never part of a
    // path: it is a list typed into one value, which would be stored as a single pattern that
    // matches nothing (SPEC-20261123-take-scope-stored REQ-12).
    if let Some(i) = entries.iter().position(|e| e.contains(',')) {
        // The option the caller typed: give-back and line-lease take `--path`, review `--keep`.
        let option = match action {
            "give-back" | "line-lease" => "--path",
            "review" => "--keep",
            _ => "--scope",
        };
        return Err(format!(
            "card {action} refused: scope entry {} has a comma; one `{option}` per path (repeat the \
             option), a comma does not separate paths",
            i + 1
        ));
    }
    let scope =
        input::validate_scope(entries).map_err(|e| format!("card {action} refused: {e}"))?;
    if scope.len() > reservation::MAX_SCOPE_PATTERNS {
        return Err(format!(
            "card {action} refused: the write scope has {} entries, at most {}",
            scope.len(),
            reservation::MAX_SCOPE_PATTERNS
        ));
    }
    if let Some(i) = scope
        .iter()
        .position(|s| s.chars().count() > reservation::MAX_PATTERN_LEN)
    {
        return Err(format!(
            "card {action} refused: scope entry {} is longer than {} characters",
            i + 1,
            reservation::MAX_PATTERN_LEN
        ));
    }
    Ok(scope)
}

/// `card scope C-n --scope P…`: sets the write scope of a card.
///
/// A `draft` or `ready` card that nobody holds gets its scope replaced, with one `card.scope`
/// event that keeps the old scope, the new one and `--by` (and the `--reason` when one is given).
/// A card that has an active reservation (`active`, or `review`) is widened instead (card C-94):
/// the new list must keep every pattern the reservation has, the added patterns are checked
/// against every other card's active reservation, and they are inserted into the same
/// reservation under the same generation, with a `card.scope` event that also names the
/// generation, the holder and the `--reason` (required then). The card's own `write_scope` is
/// not touched by a widening. `generation`, when given, is the generation the caller read: the
/// card's live reservation must be at it, else a `conflict` that names the generation the card
/// has (or that it is not held), checked in the same write transaction (card C-129, SPEC
/// SPEC-20261119-card-scope-widen-followup REQ-4 and REQ-5). A card added without `--scope` is
/// skipped by `card next` (`empty_scope`) until it has a scope: this command sets it, and so does
/// `take`, which keeps a given scope or fills an empty one from the card's SPEC or its last
/// reservation (SPEC-20261123-take-scope-stored). Any other state is a conflict that names it;
/// nothing is written then.
pub fn scope(
    project: &Path,
    home: &Path,
    reference: &str,
    entries: &[String],
    by: &str,
    reason: Option<&str>,
    generation: Option<i64>,
) -> Result<Value> {
    let card_id = card_id("scope", reference)?;
    let new = checked_scope("scope", entries)?;
    if new.is_empty() {
        return Err("card scope refused: give at least one --scope".to_string());
    }
    let by = actor(Some(by), "scope")?;
    let shared = take_scope::shared_files(project);
    let reason = reason.map(|r| checked_reason("scope", r)).transpose()?;
    let new_json = serde_json::to_string(&new).map_err(|e| e.to_string())?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let outcome = ledger
        .with_write_tx(|tx, now| {
            let row: Option<(String, String)> = tx
                .query_row(
                    "SELECT state, write_scope FROM card WHERE card_id = ?1",
                    [card_id],
                    |r| Ok((r.get(0)?, r.get(1)?)),
                )
                .optional()?;
            let Some((state, old)) = row else {
                return Ok(Err(format!("card not found: there is no C-{card_id}")));
            };
            let holder: Option<(String, i64)> = tx
                .query_row(
                    &format!(
                        "SELECT holder, generation FROM reservation WHERE card_id = ?1 \
                         AND {LIVE_RESERVATION} ORDER BY generation DESC LIMIT 1"
                    ),
                    [card_id],
                    |r| Ok((r.get(0)?, r.get(1)?)),
                )
                .optional()?;
            let held = holder.as_ref().map_or(String::new(), |(h, g)| {
                format!(", held by {h} at generation {g}")
            });
            if let Some(want) = generation {
                match &holder {
                    Some((_, have)) if *have == want => {}
                    Some((_, have)) => {
                        return Ok(Err(format!(
                            "card conflict in generation: C-{card_id} is at generation {have}, \
                             not {want}; read the card again (`card show`) before widening it"
                        )));
                    }
                    None => {
                        return Ok(Err(format!(
                            "card conflict in generation: C-{card_id} is `{state}` and not held, \
                             so there is no generation {want} to widen at"
                        )));
                    }
                }
            }
            // A shared file of the profile is never added to a scope whole; an entry the card's
            // current scope already holds passes, so a held card is still widened by other paths
            // (SPEC-20261124-shared-files-profile).
            if !shared.is_empty() {
                let current: Vec<String> =
                    if holder.is_some() && (state == "active" || state == "review") {
                        let mut stmt = tx.prepare(&format!(
                            "SELECT pattern FROM reservation_scope WHERE reservation_id = \
                         (SELECT reservation_id FROM reservation WHERE card_id = ?1 \
                          AND {LIVE_RESERVATION} ORDER BY generation DESC LIMIT 1)"
                        ))?;
                        let rows = stmt.query_map([card_id], |r| r.get::<_, String>(0))?;
                        rows.collect::<rusqlite::Result<Vec<String>>>()?
                    } else {
                        serde_json::from_str(&old).unwrap_or_default()
                    };
                let added: Vec<String> = new
                    .iter()
                    .filter(|entry| !current.contains(entry))
                    .cloned()
                    .collect();
                if let Err(refused) =
                    take_scope::refuse_shared("scope", &input::card_name(card_id), &shared, &added)
                {
                    return Ok(Err(refused));
                }
            }
            if holder.is_some() && (state == "active" || state == "review") {
                let Some(reason) = reason.as_deref() else {
                    return Ok(Err(format!(
                        "card scope refused: C-{card_id} is `{state}`{held}; widening a held \
                         card needs --reason (one line: who asked for the wider scope and why)"
                    )));
                };
                let widened = match reservation::widen_in_tx(tx, now, card_id, &new, &by, reason)? {
                    Ok(widened) => widened,
                    Err(refused) => return Ok(Err(refused.to_string())),
                };
                touch_card(tx, card_id)?;
                let rev = card_rev(tx, card_id)?;
                let overlaps = reservation::overlaps(tx, now, card_id, &widened.added)?;
                return Ok(Ok(json!({
                    "state": state, "held": true,
                    "generation": widened.reservation.generation,
                    "holder": widened.reservation.holder,
                    "from": widened.from, "write_scope": widened.reservation.write_scope,
                    "added": widened.added, "reason": reason, "rev": rev,
                    "overlaps": overlaps_json(&overlaps),
                })));
            }
            if state != "draft" && state != "ready" {
                return Ok(Err(format!(
                    "card conflict in state: C-{card_id} is `{state}`{held}; only a `draft` or \
                     `ready` card that nobody holds gets a new scope, and a held card only \
                     grows; the scope is set before the card is taken"
                )));
            }
            if holder.is_some() {
                return Ok(Err(format!(
                    "card conflict in state: C-{card_id} is `{state}`{held}; release or hand it \
                     over first"
                )));
            }
            tx.execute(
                "UPDATE card SET write_scope = ?2, updated_at = ?3 WHERE card_id = ?1",
                params![card_id, new_json, now],
            )?;
            let old: Value = serde_json::from_str(&old).unwrap_or(Value::Null);
            let key = ledger::new_event_key(tx, "card.scope", Some(card_id))?;
            let mut payload = json!({"from": old, "to": new});
            if let Some(reason) = reason.as_deref() {
                payload["reason"] = json!(reason);
            }
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, 'card.scope', ?5)",
                params![key, now, by, card_id, payload.to_string()],
            )?;
            let rev = card_rev(tx, card_id)?;
            let mut answer = json!({
                "state": state, "from": old, "write_scope": new, "rev": rev,
            });
            if let Some(reason) = reason.as_deref() {
                answer["reason"] = json!(reason);
            }
            Ok(Ok(answer))
        })
        .map_err(|e| e.to_string())??;
    let mut answer = json!({
        "action": "scope", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "by": by, "events": 1, "writes": true,
    });
    if let (Some(base), Some(extra)) = (answer.as_object_mut(), outcome.as_object()) {
        for (key, value) in extra {
            base.insert(key.clone(), value.clone());
        }
    }
    Ok(answer)
}

/// The `rev` of a card (it moves on every update of the card row).
fn card_rev(tx: &rusqlite::Connection, card_id: i64) -> rusqlite::Result<i64> {
    tx.query_row("SELECT rev FROM card WHERE card_id = ?1", [card_id], |r| {
        r.get(0)
    })
}

/// A widening changes the reservation, not the card row; touching the card moves `updated_at`
/// and `rev` so `card show` readers see that the card changed.
fn touch_card(tx: &rusqlite::Connection, card_id: i64) -> rusqlite::Result<()> {
    tx.execute(
        "UPDATE card SET updated_at = unixepoch() WHERE card_id = ?1",
        [card_id],
    )
    .map(|_| ())
}

/// What `card take` was given.
pub struct TakeRequest<'a> {
    pub holder: &'a str,
    pub lease_secs: Option<i64>,
    pub scope: &'a [String],
    pub worktree: Option<&'a str>,
    pub branch: Option<&'a str>,
    /// Refuse a card that changes the repository unless the SPEC it names is approved.
    pub require_approved_spec: bool,
    /// The folder the command runs in: without `worktree` the reservation records its git
    /// worktree (SPEC-20261125-card-take-records-worktree REQ-1).
    pub running_dir: Option<&'a Path>,
}

/// `card take C-n`: reserves a `ready` card for the holder. Without `--scope` the card's own
/// write scope is used, read inside the reserving transaction (`reserve_own_scope`), so a
/// `card scope` at the same moment cannot leave the reservation with the old scope.
///
/// The answer carries the SPEC the card names and its status (`spec`) and a warning when a card
/// that changes the repository has no approved SPEC (`warnings`, SDD-4). With
/// `require_approved_spec` that warning is a refusal instead, before anything is reserved. The
/// link is read before the reserving transaction: this is a warning gate, not the merge gate.
pub fn take(project: &Path, home: &Path, reference: &str, req: &TakeRequest<'_>) -> Result<Value> {
    let card_id = card_id("take", reference)?;
    // Checked as `add --scope` checks it: no absolute path, no `.` or `..` (card C-68 N-e).
    let given = if req.scope.is_empty() {
        None
    } else {
        Some(checked_scope("take", req.scope)?)
    };
    // A shared file of the profile is never reserved whole (SPEC-20261124-shared-files-profile):
    // a given scope is refused before the ledger is opened for write.
    let shared = take_scope::shared_files(project);
    if let Some(given) = &given {
        take_scope::refuse_shared("take", &input::card_name(card_id), &shared, given)?;
    }
    let (id, mut ledger) = open_for_write(project, home)?;
    // A card that is not there has no SPEC; `reserve` answers `card not found` for it.
    let linked = spec_link::kind_and_spec(ledger.conn(), card_id)
        .map_err(|e| ledger::LedgerError::from(e).to_string())?;
    let (spec, warning) = match &linked {
        Some((kind, spec)) => {
            spec_link::status(project, &input::card_name(card_id), kind, spec.as_deref())
        }
        None => (Value::Null, None),
    };
    if let (true, Some(why)) = (req.require_approved_spec, &warning) {
        return Err(format!(
            "card take refused: --require-approved-spec: {why}; nothing was taken"
        ));
    }
    // A `ready` card without a counted criterion is refused here; a missing card, a card in
    // another state and a held card keep the refusal `reserve` gives them.
    let is_ready: bool = ledger
        .conn()
        .query_row(
            "SELECT state = 'ready' FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| e.to_string())?
        .unwrap_or(false);
    if is_ready {
        // A HOLD or an unfinished prerequisite is refused before any reservation is written
        // (SPEC-20261121-card-hold-gate); `card next` asks the same two functions.
        if let Some(why) = crate::ledger::workflow::hold_gate_refusal(ledger.conn(), card_id)
            .map_err(|e| e.to_string())?
        {
            return Err(why);
        }
        if let Some(why) =
            no_criteria_refusal(ledger.conn(), "take", card_id).map_err(|e| e.to_string())?
        {
            return Err(why);
        }
        // At the limit of cards in work and in review a session helps with a check or a landing
        // instead of taking one more (SPEC-20261125-landing-moves-checkout REQ-1).
        if let Some(why) = super::wip_limit::refusal(project, ledger.conn(), "take")? {
            return Err(why);
        }
    }
    let lease = req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS);
    // The worktree is always recorded when it can be known: the given path as it is, else the
    // git worktree of the running folder (SPEC-20261125-card-take-records-worktree REQ-1, REQ-2).
    let running_worktree: Option<String> = match req.worktree {
        Some(_) => None,
        None => req.running_dir.and_then(crate::lane_registry::worktree_of),
    };
    let worktree: Option<&str> = req.worktree.or(running_worktree.as_deref());
    // Where the scope comes from (SPEC-20261123-take-scope-stored): the given one, the card's
    // own, the linked SPEC's `write:` list, the last reservation's scope. A scope that is not the
    // card's own is kept on the card once the reservation is made, so a released card is not
    // empty again; `reserve` runs first and a refusal stores nothing.
    let own = take_scope::own_scope(ledger.conn(), card_id)
        .map_err(|e| ledger::LedgerError::from(e).to_string())?;
    let own_is_empty = own.as_ref().is_some_and(Vec::is_empty);
    let mut merged: Vec<String> = Vec::new();
    let mut left_out = take_scope::LeftOut::new();
    let (scope, source) = match given {
        Some(scope) => (Some(scope), take_scope::Source::Given),
        None if own_is_empty => {
            let spec_id = linked.as_ref().and_then(|(_, spec)| spec.as_deref());
            match take_scope::choose(project, ledger.conn(), card_id, spec_id)? {
                Some(fill) => {
                    merged = fill.merged;
                    left_out = fill.left_out;
                    (Some(fill.scope), fill.source)
                }
                None => return Err(take_scope::no_source_message(&input::card_name(card_id))),
            }
        }
        None => {
            // The card's own scope: a shared file in it is refused like a given one.
            if let Some(own) = &own {
                take_scope::refuse_shared("take", &input::card_name(card_id), &shared, own)?;
            }
            (None, take_scope::Source::Card)
        }
    };
    // A card whose lease ended at least the grace period ago is swept and taken in this one
    // transaction (SPEC-20261124-ended-lease-sweep REQ-12); a profile that cannot be read gives
    // the default grace period and never stops a take.
    let grace = super::sweep::grace_secs_or_default(project);
    let r = match scope {
        Some(scope) => reservation::reserve_sweeping(
            &mut ledger,
            card_id,
            req.holder,
            &scope,
            lease,
            worktree,
            req.branch,
            grace,
        ),
        None => reservation::reserve_own_scope_sweeping(
            &mut ledger,
            card_id,
            req.holder,
            lease,
            worktree,
            req.branch,
            grace,
        ),
    }
    .map_err(|e| e.to_string())?;
    let mut warnings: Vec<String> = warning.into_iter().collect();
    if worktree.is_none() {
        warnings.push(format!(
            "{}: the reservation records no worktree (no `--worktree` and the running folder is \
             not inside a git worktree); `card adopt` needs one, so pass `--worktree`",
            input::card_name(card_id)
        ));
    }
    if !merged.is_empty() {
        let names: Vec<String> = merged.iter().map(|p| format!("`{p}`")).collect();
        warnings.push(format!(
            "{}: the `write:` list of the SPEC names {} more than once; each path is reserved \
             once (merged)",
            input::card_name(card_id),
            names.join(", ")
        ));
    }
    // The worktree's hook settings may point at files it does not have (SPEC-20261125-card-
    // worktree-hook-config); a read problem never blocks or warns.
    if let Some(line) = worktree.and_then(|w| {
        crate::go_hooks::missing_scripts_warning(&input::card_name(card_id), Path::new(w))
    }) {
        warnings.push(line);
    }
    for (entry, hit) in &left_out {
        warnings.push(take_scope::left_out_warning(
            &input::card_name(card_id),
            entry,
            hit,
        ));
    }
    if own_is_empty && source != take_scope::Source::Card {
        // The take already succeeded; a scope that could not be kept is a warning, never a
        // failure that would leave the card taken and reported refused.
        if let Err(why) =
            take_scope::store(&mut ledger, card_id, &r.write_scope, req.holder, source)
        {
            warnings.push(format!(
                "{}: the scope could not be kept on the card ({why}); `card scope` sets it",
                input::card_name(card_id)
            ));
        }
    }
    // The take is made whatever other cards reserve; the overlaps are only reported.
    let overlaps = reservation::overlaps(ledger.conn(), r.acquired_at, card_id, &r.write_scope)
        .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "take", "project_id": id, "reservation": reservation_json(&r),
        "state": "active", "spec": spec, "scope_source": source.word(), "warnings": warnings,
        "overlaps": overlaps_json(&overlaps), "writes": true,
    }))
}

/// `card release C-n`: gives the card back; an `active` card becomes `ready`. A `review` card
/// is refused by `reservation::release`, inside its transaction and after the fence.
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
    let mut answer = json!({
        "action": "release", "project_id": id, "reservation": reservation_json(&out.reservation),
        "state": out.card_state, "writes": true,
    });
    // A card whose landing is recorded is not plain work again: say so and name the way to close
    // it. The ledger is the only source (no git command), and nothing else of the answer changes.
    if has_recorded_landing(ledger.conn(), card_id).map_err(|e| e.to_string())? {
        answer["landed"] = json!(true);
        answer["next_step"] = json!(format!(
            "{name} already landed: close it with `card finish {name} --by <you>` instead of working it again",
            name = input::card_name(card_id)
        ));
    }
    Ok(answer)
}

/// The leader's `card release C-n --by leader --reason <why>`
/// (SPEC-20261125-card-take-records-worktree REQ-8 to REQ-14): the reservation of ANOTHER session
/// ends as `released` and the card returns to `ready`, so any lane takes it again with `card take`.
/// It is no reclaim: no `card.reclaim` event is written and the three-reclaim allowance is not
/// used, which is the way out of a card that reached the limit.
///
/// Only the session of the live registered leader entry may call it (the proof `card land-batch`
/// asks for, `confirm_leader`): the word `--by leader` proves nothing, and a call without a
/// session id reaches the same refusal (`not_leader`). A reservation whose
/// lease ended is released as it is; a live one needs the same proofs as an early reclaim
/// ([`early_reclaim::assess`]) and the first missing proof is named in the `lease_live` refusal.
/// A card in `review` with a `pass` verdict is refused (`verdict_pass`): it is landed instead.
/// Every refusal writes nothing.
pub fn release_by_leader(
    project: &Path,
    home: &Path,
    reference: &str,
    generation: i64,
    session: Option<&str>,
    by: &str,
    reason: &str,
) -> Result<Value> {
    if by != ledger::workflow::LEADER_BY {
        return Err(format!(
            "card release refused: --by must be `{}`; nothing was changed",
            ledger::workflow::LEADER_BY
        ));
    }
    let card_id = card_id("release", reference)?;
    let reason = checked_reason("release", reason)?;
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64);
    if let Err(inner) = super::land::batch::confirm_leader(project, session, now) {
        // The land-batch wording of the shared check is replaced by this command's own; the
        // `conflict in` words make `contract::classify` answer `conflict`.
        let inner = inner
            .strip_prefix("card land-batch refused: ")
            .unwrap_or(&inner);
        let inner = inner.strip_suffix("; nothing was changed").unwrap_or(inner);
        return Err(format!(
            "card release refused: conflict in leader: not_leader: {inner}; nothing was changed"
        ));
    }
    let early = early_reclaim::assess(project, home, card_id, session.unwrap_or_default());
    let proof = match &early {
        early_reclaim::Early::Proven(proof) => Some(reservation::EarlyProof {
            holder: &proof.holder,
            detail: &proof.detail,
        }),
        _ => None,
    };
    let (id, mut ledger) = open_for_write(project, home)?;
    let out = reservation::release_by_leader(
        &mut ledger,
        card_id,
        generation,
        by,
        &reason,
        proof.as_ref(),
    )
    .map_err(|e| match (&e, &early) {
        (
            reservation::ReclaimError::LeaseLive { .. }
            | reservation::ReclaimError::LeaseHeldForAudit { .. },
            early_reclaim::Early::Unproven(why),
        ) => format!(
            "{e}; lease_live: a release of a live reservation needs the early proof that \
                 the holder is gone, and {why}"
        ),
        (
            reservation::ReclaimError::LeaseLive { .. }
            | reservation::ReclaimError::LeaseHeldForAudit { .. },
            _,
        ) => format!(
            "{e}; lease_live: a release of a live reservation needs the early proof that \
                 the holder is gone"
        ),
        (reservation::ReclaimError::HolderChanged { .. }, _) => format!(
            "{e}; holder_changed: the proof was read for another holder, read the card again"
        ),
        _ => e.to_string(),
    })?;
    let mut answer = json!({
        "action": "release", "project_id": id,
        "by": by, "reason": reason,
        "reservation": reservation_json(&out.previous),
        "state": out.card_state_after,
        "card_state_before": out.card_state_before,
        "from_holder": out.previous.holder,
        "events": 1, "writes": true,
    });
    if let Some(p) = &proof {
        answer["early_proof"] = p.detail.clone();
    }
    Ok(answer)
}

/// Whether the ledger records a landing of the card: a `card.landed` event or a merged pull
/// request link.
fn has_recorded_landing(conn: &rusqlite::Connection, card_id: i64) -> rusqlite::Result<bool> {
    conn.query_row(
        "SELECT EXISTS(SELECT 1 FROM event WHERE card_id = ?1 AND kind = 'card.landed') \
         OR EXISTS(SELECT 1 FROM card_link WHERE from_card = ?1 AND kind = 'pr' \
                   AND state = 'active' AND pr_state = 'merged')",
        [card_id],
        |r| r.get(0),
    )
}

/// `card handoff C-n`: the holder's reservation ends and `to` gets the next generation. The
/// card keeps its state (`active` or `review`), which the answer reports.
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
        "from_holder": holder,
        "state": card_state(&ledger, card_id)?.unwrap_or_default(),
        "writes": true,
    }))
}

/// `--reason` of `reclaim`: one line of 1 to [`reservation::MAX_REASON_LEN`] characters with
/// something visible in it and no word shaped like a secret (the filter `card note` uses). It is
/// what the reclaimer found out about the old holder.
/// A `--reason` that is kept for good: one line of 1 to [`reservation::MAX_REASON_LEN`]
/// characters, passing the one card text screen ([`crate::card::input::secret_words`], the same
/// check `card note` uses; `card export` writes it out). The text is never repeated. The
/// text is kept without its leading and trailing blanks; a line separator (U+2028, U+2029) or a
/// control character anywhere in what was typed is refused.
pub(super) fn checked_reason(action: &str, reason: &str) -> Result<String> {
    let kept = reason.trim();
    if kept.is_empty()
        || kept.chars().count() > reservation::MAX_REASON_LEN
        || reason.chars().any(char::is_control)
        || has_line_separator(reason)
    {
        return Err(format!(
            "card {action} refused: --reason must be one line of 1 to {} characters",
            reservation::MAX_REASON_LEN
        ));
    }
    // Follow-ups REQ-5 (SPEC-20261120-card-secret-followups): the same one screen as the note
    // and the criterion fields, not the shared shape filter alone.
    let words = crate::card::input::secret_words(kept);
    if words > 0 {
        return Err(format!(
            "card {action} refused: --reason has {}",
            crate::card::input::secret_refusal(words)
        ));
    }
    Ok(kept.to_string())
}

/// U+2028 (line separator) and U+2029 (paragraph separator): not control characters, so
/// `char::is_control` lets them through, but they end a line in many readers.
fn has_line_separator(text: &str) -> bool {
    text.chars().any(|c| matches!(c, '\u{2028}' | '\u{2029}'))
}

fn reclaim_reason(reason: &str) -> Result<String> {
    checked_reason("reclaim", reason)
}

/// The reason of an early reclaim: the caller's reason, then the proof in brackets, as one line
/// that still fits the reason limit (REQ-11). A reason that would not fit is refused before any
/// write, naming how much the proof adds.
fn early_reason(reason: &str, proof: &str) -> Result<String> {
    let combined = format!("{reason} [{proof}]");
    let (own, all) = (reason.chars().count(), combined.chars().count());
    if all > reservation::MAX_REASON_LEN {
        return Err(format!(
            "card reclaim refused: --reason is {own} characters and the early-reclaim proof adds {}, \
             over the limit of {}; shorten the reason",
            all - own,
            reservation::MAX_REASON_LEN
        ));
    }
    checked_reason("reclaim", &combined)
}

/// What `card reclaim` was given.
pub struct ReclaimRequest<'a> {
    /// The new holder: this session.
    pub to_holder: &'a str,
    /// The generation the caller saw as the card's current one.
    pub generation: i64,
    /// Who reclaims (the audit name; the actor of the event).
    pub by: &'a str,
    /// What the caller found out about the old holder.
    pub reason: &'a str,
    /// Lease of the new reservation in seconds (default [`DEFAULT_LEASE_SECS`]).
    pub lease_secs: Option<i64>,
    /// The folder the command runs in: a reservation that records no worktree takes its git
    /// worktree (SPEC-20261125-card-take-records-worktree REQ-3).
    pub running_dir: Option<&'a Path>,
}

/// `card reclaim C-n`: the reservation at `generation` must be `held` with a lease that has
/// ended; it becomes `reclaimed` and this session gets the next generation with the same scope,
/// worktree and branch (`reservation::reclaim`). The card keeps its state, notes, criteria and
/// links; one `card.reclaim` event names the old holder, the new one, `--by` and `--reason`. A
/// live lease, another generation, a `done` card or a card that is not held is refused and
/// nothing is written.
///
/// One exception to "a live lease is refused" (SPEC-20261120-reclaim-dead-holder, user decision
/// 2026-11-20): when the holder session is proven gone ([`super::early_reclaim`]: its launcher
/// has exited, its worktree is gone or clean and no process works in it) the reservation is
/// reclaimed while its lease still runs, and the proof goes into the reason and the event. A
/// missing or unreadable proof leaves the refusal as it was, with the missing part named.
pub fn reclaim(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &ReclaimRequest<'_>,
) -> Result<Value> {
    let card_id = card_id("reclaim", reference)?;
    let by = actor(Some(req.by), "reclaim")?;
    let reason = reclaim_reason(req.reason)?;
    let kept = takeover_gates(project, home, card_id, req.to_holder)?;
    let early = early_reclaim::assess(project, home, card_id, req.to_holder);
    let reason = match &early {
        early_reclaim::Early::Proven(proof) => early_reason(&reason, &proof.text)?,
        _ => reason,
    };
    let (id, mut ledger) = open_for_write(project, home)?;
    let lease_secs = req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS);
    let proof = match &early {
        early_reclaim::Early::Proven(proof) => Some(reservation::EarlyProof {
            holder: &proof.holder,
            detail: &proof.detail,
        }),
        _ => None,
    };
    // The running folder's worktree fills a worktree the old reservation never recorded.
    let running_worktree = req.running_dir.and_then(crate::lane_registry::worktree_of);
    // A reservation that comes back without paths is filled from the card's own scope, leaving
    // out the shared files of the profile (SPEC-20261125-review-lease-hold REQ-5, REQ-6).
    let shared = take_scope::shared_files(project);
    let taken = reservation::reclaim_recording(
        &mut ledger,
        card_id,
        req.to_holder,
        req.generation,
        &by,
        &reason,
        lease_secs,
        proof.as_ref(),
        kept.as_ref(),
        running_worktree.as_deref(),
        &shared,
    );
    let out = taken.map_err(|e| match (&e, &early) {
        // The lease is live and the proof is missing: say which part.
        (
            reservation::ReclaimError::LeaseLive { .. }
            | reservation::ReclaimError::LeaseHeldForAudit { .. },
            early_reclaim::Early::Unproven(why),
        ) => {
            format!("{e}; an early reclaim needs proof that the holder is gone, and {why}")
        }
        _ => e.to_string(),
    })?;
    let mut answer = json!({
        "action": "reclaim", "project_id": id,
        "reservation": reservation_json(&out.reservation),
        "previous": reservation_json(&out.previous),
        "by": by, "reason": reason, "state": out.card_state,
        "counted": out.counted,
        "reclaims_counted": out.reclaims_counted,
        "reclaim_allowance": reservation::RECLAIM_ALLOWANCE,
        "events": 1, "writes": true,
    });
    add_scope_fill(&mut answer, out.scope_fill.as_ref());
    if let Some(kept) = kept {
        answer["kept_commits"] = kept;
    }
    // The worktree the reservation now records may have hook settings pointing at files it
    // lacks (SPEC-20261125-card-worktree-hook-config).
    if let Some(line) = out.reservation.worktree.as_deref().and_then(|w| {
        crate::go_hooks::missing_scripts_warning(&input::card_name(card_id), Path::new(w))
    }) {
        answer["warnings"] = json!([line]);
    }
    Ok(answer)
}

/// The card's state now, `None` when there is no such card.
fn card_state(ledger: &ledger::WriteLedger, card_id: i64) -> Result<Option<String>> {
    ledger
        .conn()
        .query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .optional()
        .map_err(|e| ledger::LedgerError::from(e).to_string())
}

// ---- takeover gates (SPEC-20261119-card-reclaim-guard) ----

/// The takeover gates run before the ledger write: a self-reclaim is a renewal, not a
/// takeover, and skips them (REQ-4); otherwise the recorded worktree must not hold work a
/// reclaim would orphan. The holder-liveness gate has no mechanical join yet — the session
/// registry records no host session id — so it stays a follow-up (card C-43 note 2836).
/// Refusals keep the `conflict in` wording so `contract::classify` maps them to `conflict`.
///
/// The answer is the commits the takeover keeps (`kept_commits`, SPEC-20261124-reclaim-keeps-
/// worktree): in the takeover of an ended lease only an uncommitted change stops it, unpushed
/// commits stay in the same worktree for the new holder. While the lease is live the strict
/// [`worktree_gate`] judges, as it always did. `None` when no worktree was read or kept.
fn takeover_gates(
    project: &Path,
    home: &Path,
    card_id: i64,
    to_holder: &str,
) -> Result<Option<Value>> {
    let id = crate::ledger::command::project_id_of(project)
        .map_err(|e| format!("card reclaim refused: {e}"))?;
    // Read-only peek: an absent ledger reaches its own refusal at `open_for_write`, and the
    // peek never creates anything (`open_read_only`).
    let Ok(ro) = ledger::open_read_only(home, &id) else {
        return Ok(None);
    };
    let current = reservation::current(ro.conn(), card_id).map_err(|e| e.to_string())?;
    let Some(cur) = current else {
        return Ok(None); // the ledger refuses with its own `no reservation`
    };
    if cur.holder == to_holder {
        return Ok(None); // a self-reclaim renews its own reservation
    }
    // Only the takeover of an ended lease keeps unpushed commits (SPEC-20261124-reclaim-keeps-
    // worktree REQ-6, D-6). A live lease, a reservation that is not `held` and a clock that
    // cannot be read all go through the strict gate, whose refusal text is the one the command
    // always had; the ledger then refuses a live lease with its own reason after it.
    let clock = ro
        .conn()
        .query_row("SELECT unixepoch()", [], |r| r.get::<_, i64>(0));
    if is_ended_lease_takeover(&cur.status, cur.lease_until, clock) {
        takeover_worktree_gate(project, cur.worktree.as_deref())
    } else {
        worktree_gate(project, cur.worktree.as_deref()).map(|()| None)
    }
}

/// Whether a reclaim is the takeover of an ended lease: the reservation is `held`, the clock
/// reading succeeded and the lease has ended at that reading. Every other case (a reservation in
/// another status, a clock that cannot be read, a live lease) is judged by the strict gate. The
/// clock is a parameter so that a test can hand over an unreadable one
/// (SPEC-20261124-card-help-gates REQ-11).
fn is_ended_lease_takeover(status: &str, lease_until: i64, clock: rusqlite::Result<i64>) -> bool {
    status == "held" && clock.is_ok_and(|now| reservation::lease_ended(lease_until, now))
}

/// REQ-1..3: refuse a worktree that exists and holds uncommitted or unpushed work, go on
/// only when the path is known to be gone (nothing there can be orphaned), refuse as
/// unverifiable when the path cannot be told apart from "gone" or git cannot be read.
pub(super) fn worktree_gate(project: &Path, worktree: Option<&str>) -> Result<()> {
    let Some(read) = read_worktree(project, worktree)? else {
        return Ok(());
    };
    if read.uncommitted > 0 || read.unpushed > 0 {
        return Err(format!(
            "card conflict in worktree dirty: {} holds {} uncommitted change(s) \
             and {} unpushed commit(s); a reclaim would orphan them",
            read.wt, read.uncommitted, read.unpushed
        ));
    }
    Ok(())
}

/// What the takeover gate keeps: the unpushed commits of the recorded worktree. The new holder
/// continues in the same worktree on the same branch, so these commits are not lost, and a lane
/// never pushes, so "unpushed" is the normal state of an open card
/// (SPEC-20261124-reclaim-keeps-worktree REQ-1, REQ-3).
const KEPT_COMMIT_IDS: usize = 20;

/// REQ-1, REQ-2, REQ-4 of SPEC-20261124-reclaim-keeps-worktree: the gate of a takeover of an
/// ended lease. It refuses what a takeover really loses, an uncommitted change, whether or not
/// commits stand beside it, and refuses as unreadable when git cannot count the unpushed commits.
/// Unpushed commits pass and are returned as `kept_commits` (the count and up to twenty ids,
/// newest first). `None` when no worktree is recorded or its path is known to be gone.
fn takeover_worktree_gate(project: &Path, worktree: Option<&str>) -> Result<Option<Value>> {
    let Some(read) = read_worktree(project, worktree)? else {
        return Ok(None);
    };
    if read.uncommitted > 0 {
        return Err(format!(
            "card conflict in worktree dirty: {} holds {} uncommitted change(s) \
             and {} unpushed commit(s); a reclaim would orphan the uncommitted changes",
            read.wt, read.uncommitted, read.unpushed
        ));
    }
    let ids = unpushed_ids(
        Path::new(&read.wt),
        "HEAD",
        read.integration.as_deref(),
        KEPT_COMMIT_IDS,
    )
    .ok_or_else(|| {
        format!(
            "card conflict in worktree unknown: the worktree {} could not be read: git could not \
             list its unpushed commits",
            read.wt
        )
    })?;
    Ok(Some(json!({ "count": read.unpushed, "ids": ids })))
}

/// What one read of a recorded worktree found.
struct WorktreeRead {
    wt: String,
    uncommitted: u64,
    unpushed: u64,
    integration: Option<String>,
}

/// The one read both gates share: `Ok(None)` when no worktree is recorded or its path is known
/// to be gone (nothing there can be orphaned), a refusal when the path cannot be told apart from
/// "gone" or git cannot be read, else the counts of uncommitted changes and unpushed commits.
fn read_worktree(project: &Path, worktree: Option<&str>) -> Result<Option<WorktreeRead>> {
    let Some(wt) = worktree.filter(|w| !w.is_empty()) else {
        return Ok(None);
    };
    // The reason git or the file system gave comes with the refusal (SPEC-20261122-test-defects-
    // session-reclaim REQ-18).
    let unreadable = |why: &str| {
        format!("card conflict in worktree unknown: the worktree {wt} could not be read: {why}")
    };
    let path = Path::new(wt);
    // A relative path means a different folder under another working directory, so "it does
    // not exist here" proves nothing (audit-c43 B1).
    if !path.is_absolute() {
        return Err(format!(
            "card conflict in worktree unknown: the recorded worktree {wt} is not an absolute \
             path, so it cannot be checked from here"
        ));
    }
    // `Path::exists` turns a permission error into "gone"; only a clean "not found" is gone.
    match path.try_exists() {
        Ok(false) => return Ok(None),
        Ok(true) => {}
        Err(e) => return Err(unreadable(&e.to_string())),
    }
    // A folder that lost its own `.git` and lies inside another checkout would be judged by that
    // checkout, which may ignore it: only a folder git calls the top of a checkout is judged
    // (REQ-22, REQ-23).
    checkout_top_is(path).map_err(|why| unreadable(&why))?;
    // One changed or untracked path per entry, every untracked file on its own (a folder git
    // would show as one entry could hide real work beside a tool cache), names unquoted (`-z`).
    // An untracked file under the runtime state folder is tool state, not work
    // (SPEC-20261120-reclaim-dead-holder REQ-12, widened from one file to the folder by
    // SPEC-20261122-test-defects-session-reclaim REQ-3). A folder git cannot open is only a
    // warning to git, so any warning makes the worktree unreadable (REQ-9, REQ-10).
    let state = super::early_reclaim::state_prefix(path);
    let uncommitted = git_entries(
        path,
        &["status", "--porcelain=v1", "-z", "--untracked-files=all"],
    )
    .map_err(|why| unreadable(&why))?
    .iter()
    .filter(|e| !super::early_reclaim::is_tool_cache_entry(&state, e))
    .count() as u64;
    let integration = integration_ref(project, path);
    let unpushed = unpushed_commits(path, "HEAD", integration.as_deref())
        .ok_or_else(|| unreadable("git could not count its unpushed commits"))?;
    Ok(Some(WorktreeRead {
        wt: wt.to_string(),
        uncommitted,
        unpushed,
        integration,
    }))
}

/// The integration branch the NeXT-Flow work flow of the profile in `dir` names, if any.
fn mflow_integration_branch(dir: &Path) -> Option<String> {
    let (profile, _) = crate::profile::read_profile(dir).ok()?;
    profile
        .work_flow
        .filter(|w| crate::profile::is_local_flow(&w.flow))?
        .integration_branch
}

/// The local integration branch of a NeXT-Flow project, as a full ref: the profile of the
/// project folder names the NeXT-Flow work flow with an integration branch, else the profile of
/// the recorded worktree does (a lane worktree has no state-root marker, so its own profile is
/// the earlier one without a work flow), and that branch exists locally. The ref is looked up
/// in the worktree, whose refs are the clone's. `None` for every other case (no profile, one
/// that cannot be read, another flow, no branch named, a branch this clone lacks), which leaves
/// the count of unpushed commits to the remote-tracking refs alone
/// (SPEC-20261120-reclaim-dead-holder REQ-13).
pub(super) fn integration_ref(project: &Path, worktree: &Path) -> Option<String> {
    let branch =
        mflow_integration_branch(project).or_else(|| mflow_integration_branch(worktree))?;
    let full = format!("refs/heads/{branch}");
    git_output(worktree, &["rev-parse", "--verify", "--quiet", &full])?;
    Some(full)
}

/// How many commits of `tip` are on no remote-tracking ref and, when `integration` names a
/// ref, are not in it either. `None` when git cannot answer.
pub(super) fn unpushed_commits(dir: &Path, tip: &str, integration: Option<&str>) -> Option<u64> {
    let mut args = vec!["rev-list", "--count", tip, "--not", "--remotes"];
    args.extend(integration);
    git_output(dir, &args)?.trim().parse().ok()
}

/// The ids of the first `max` commits of `tip` that are on no remote-tracking ref and, when
/// `integration` names a ref, not in it either, newest first. `None` when git cannot answer.
fn unpushed_ids(
    dir: &Path,
    tip: &str,
    integration: Option<&str>,
    max: usize,
) -> Option<Vec<String>> {
    let limit = format!("--max-count={max}");
    let mut args = vec!["rev-list", limit.as_str(), tip, "--not", "--remotes"];
    args.extend(integration);
    let out = git_output(dir, &args)?;
    Some(out.lines().map(str::to_string).collect())
}

/// The standard output of one read-only git command in `dir`: the environment is scrubbed like
/// every spawn (`identity::scrub_git_env`) and the meaning matches `lane_reset`'s probes — no
/// fetch, untracked files counted, `None` on any failure.
pub(super) fn git_output(dir: &Path, args: &[&str]) -> Option<String> {
    git_run(dir, args)
        .ok()
        .map(|out| String::from_utf8_lossy(&out.stdout).into_owned())
}

/// The entries of one git file list read with `-z` (`status` or `ls-files`), names as written.
/// `Err` with the reason on a failure and also when git wrote a line on its standard error that
/// is no noise line: `git status` exits 0 with only a warning for a folder it cannot open and
/// lists nothing for it, so a warning means the list is incomplete
/// (SPEC-20261122-test-defects-session-reclaim REQ-9, REQ-18).
pub(super) fn git_entries(dir: &Path, args: &[&str]) -> std::result::Result<Vec<String>, String> {
    let out = git_run(dir, args)?;
    if let Some(line) = first_unexpected_line(&out.stderr) {
        return Err(format!("git said: {}", capped(&line)));
    }
    let status = args.first() == Some(&"status");
    Ok(split_entries(&String::from_utf8_lossy(&out.stdout), status))
}

/// One line of git's words, short enough to carry in a refusal.
fn capped(line: &str) -> String {
    line.chars().take(200).collect()
}

/// The first line git wrote on its standard error, which is no noise line (REQ-18).
fn first_unexpected_line(stderr: &[u8]) -> Option<String> {
    String::from_utf8_lossy(stderr)
        .lines()
        .map(str::trim)
        .filter(|l| !l.is_empty())
        .find(|l| !is_noise_line(l))
        .map(str::to_string)
}

/// A line git writes when it carried on and the list is complete (REQ-19, D-10): a name in an
/// attributes file that is not a valid attribute name, with the file and the line it stands on.
/// Every other line, one that only looks like it included, is a warning. The line is trimmed
/// before it gets here, so the space that opens the separator already means a name stands before it.
fn is_noise_line(line: &str) -> bool {
    line.split_once(" is not a valid attribute name: ")
        .is_some_and(|(_, at)| {
            at.rsplit_once(':').is_some_and(|(file, n)| {
                !file.is_empty() && !n.is_empty() && n.bytes().all(|b| b.is_ascii_digit())
            })
        })
}

/// `Ok` when git places the top of a checkout exactly at `dir`; the reason otherwise.
fn checkout_top_is(dir: &Path) -> std::result::Result<(), String> {
    let out = git_run(dir, &["rev-parse", "--show-toplevel"])?;
    // Only git's closing line break goes: a folder name may itself end in a space.
    let top = String::from_utf8_lossy(&out.stdout)
        .trim_end_matches('\n')
        .to_string();
    match (std::fs::canonicalize(&top), std::fs::canonicalize(dir)) {
        (Ok(a), Ok(b)) if a == b => Ok(()),
        _ => Err(format!(
            "it is not the top of a git checkout of its own (git places it in the checkout at \
             {top})"
        )),
    }
}

/// Splits `-z` output into entries. In `git status` output a renamed or copied entry (`R` or
/// `C` in either status column) is followed by its origin path as a field of its own, which is
/// part of that entry and not one more.
fn split_entries(raw: &str, status: bool) -> Vec<String> {
    let mut out = Vec::new();
    let mut fields = raw.split('\0').filter(|f| !f.is_empty());
    while let Some(entry) = fields.next() {
        let code = entry.get(..2).unwrap_or("");
        if status && code.contains(['R', 'C']) {
            fields.next();
        }
        out.push(entry.to_string());
    }
    out
}

/// One read-only git command in `dir`, its output when it exits 0, else the reason: how git
/// ended and the first line it wrote (REQ-18). The file monitor is switched off for every read:
/// a read gains nothing from a watcher, and a hook that does not exist would make git write two
/// `fatal:` lines on every read (REQ-20, D-11).
fn git_run(dir: &Path, args: &[&str]) -> std::result::Result<std::process::Output, String> {
    // `MOAI_GIT_BIN` names the git program for a runner whose child carries no `PATH`
    // (the same seam as `MOAI_FORGE_GH_BIN`); everyone else spawns `git`.
    let program = std::env::var("MOAI_GIT_BIN").unwrap_or_else(|_| "git".to_string());
    let mut cmd = std::process::Command::new(program);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .args(["-c", "core.fsmonitor=false"])
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(std::process::Stdio::null())
        .output()
        .map_err(|e| format!("git could not start: {e}"))?;
    if out.status.success() {
        return Ok(out);
    }
    let how = out
        .status
        .code()
        .map_or("was stopped by a signal".to_string(), |c| {
            format!("exited with {c}")
        });
    Err(match first_unexpected_line(&out.stderr) {
        Some(line) => format!("git {how} and said: {}", capped(&line)),
        None => format!("git {how}"),
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    /// SPEC-20261124-card-help-gates AC-13 (REQ-6, REQ-11): only a `held` reservation whose lease
    /// has ended at a readable clock is the takeover of an ended lease; a reservation in another
    /// status, an unreadable clock and a lease that has not ended each go to the strict gate.
    #[test]
    fn the_takeover_gate_choice_follows_status_clock_and_lease() {
        let unreadable = || Err(rusqlite::Error::QueryReturnedNoRows);
        // lease_until <= now is an ended lease, so 100 at 100 and at 101 are ended, at 99 is not.
        assert!(is_ended_lease_takeover("held", 100, Ok(101)));
        assert!(is_ended_lease_takeover("held", 100, Ok(100)));
        assert!(!is_ended_lease_takeover("held", 100, Ok(99)));
        assert!(!is_ended_lease_takeover("held_for_review", 100, Ok(101)));
        assert!(!is_ended_lease_takeover("expired_candidate", 100, Ok(101)));
        assert!(!is_ended_lease_takeover("held", 100, unreadable()));
    }

    /// SPEC-20261122-test-defects-session-reclaim AC-9: `-z` entries keep a space as written; in
    /// `git status` output a rename or copy carries its origin as one more field, which is part
    /// of that entry, while an `ls-files` list has no status columns to read.
    #[test]
    fn a_z_list_keeps_names_and_folds_a_rename_origin() {
        let status = "?? my notes.md\0R  new name.rs\0old name.rs\0 C a.rs\0b.rs\0 M c.rs\0";
        assert_eq!(
            split_entries(status, true),
            ["?? my notes.md", "R  new name.rs", " C a.rs", " M c.rs"]
        );
        assert_eq!(
            split_entries("R  x.rs\0C  y.rs\0", false),
            ["R  x.rs", "C  y.rs"]
        );
        assert!(split_entries("", true).is_empty());
    }

    /// SPEC-20261122-test-defects-session-reclaim AC-14 (REQ-18): the line a refusal carries is
    /// the first one git wrote that is no noise line, whatever stands before it.
    #[test]
    fn the_refusal_names_the_first_line_git_wrote() {
        let warn = "warning: could not open directory 'a/': Permission denied";
        assert_eq!(first_unexpected_line(b""), None);
        assert_eq!(first_unexpected_line(b"\n  \n"), None);
        let two = format!("{warn}\nwarning: another\n");
        assert_eq!(first_unexpected_line(two.as_bytes()).as_deref(), Some(warn));
        let noise_first =
            format!("we!rd is not a valid attribute name: .gitattributes:1\n{warn}\n");
        assert_eq!(
            first_unexpected_line(noise_first.as_bytes()).as_deref(),
            Some(warn)
        );
    }

    /// AC-15 (REQ-19): the one noise shape is a name that is not a valid attribute name, with the
    /// file and the line it stands on; a line that only looks like it is still a warning.
    #[test]
    fn a_noise_line_is_not_a_warning() {
        for noise in [
            "we!rd is not a valid attribute name: .gitattributes:1\n",
            "x is not a valid attribute name: sub dir/.gitattributes:12\n",
        ] {
            assert_eq!(first_unexpected_line(noise.as_bytes()), None, "{noise}");
        }
        for look_alike in [
            "is not a valid attribute name\n",
            "we!rd is not a valid attribute name: .gitattributes\n",
            "we!rd is not a valid attribute name: .gitattributes:x\n",
            "we!rd is not a valid attribute name: :3\n",
            "we!rd is not a valid attribute name: .gitattributes:\n",
            "warning: could not open directory 'x/': Permission denied\n",
        ] {
            assert!(
                first_unexpected_line(look_alike.as_bytes()).is_some(),
                "{look_alike}"
            );
        }
    }

    /// One git command in `dir` with a fixture identity; the test fails when git does.
    fn fixture_git(dir: &Path, args: &[&str]) {
        let mut cmd = std::process::Command::new("git");
        crate::identity::scrub_git_env(&mut cmd);
        let out = cmd
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
            ])
            .args([
                "-c",
                "commit.gpgsign=false",
                "-c",
                "core.excludesFile=/dev/null",
            ])
            .arg("-C")
            .arg(dir)
            .args(args)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
    }

    /// AC-17 (REQ-22, REQ-23): a recorded folder that lost its `.git` inside a checkout that
    /// ignores it is not clean because that checkout is: the gate refuses it as unreadable and
    /// names the checkout git finds. A checkout of its own is still judged by itself.
    #[test]
    fn a_folder_that_lost_its_git_is_not_judged_by_the_checkout_around_it() {
        let tmp = tempfile::tempdir().unwrap();
        let base = tmp.path().canonicalize().unwrap();
        let (bare, outer) = (base.join("remote.git"), base.join("outer"));
        for d in [&bare, &outer] {
            std::fs::create_dir_all(d).unwrap();
        }
        fixture_git(&bare, &["init", "-q", "--bare", "-b", "main"]);
        fixture_git(&outer, &["init", "-q", "-b", "main"]);
        std::fs::write(outer.join(".gitignore"), "lost/\n").unwrap();
        fixture_git(&outer, &["add", ".gitignore"]);
        fixture_git(&outer, &["commit", "-q", "-m", "ignore the folder"]);
        fixture_git(&outer, &["remote", "add", "origin", bare.to_str().unwrap()]);
        fixture_git(&outer, &["push", "-q", "origin", "main"]);
        let lost = outer.join("lost");
        std::fs::create_dir_all(&lost).unwrap();
        std::fs::write(lost.join("work.txt"), "x\n").unwrap();

        let err = worktree_gate(&outer, Some(lost.to_str().unwrap())).unwrap_err();
        assert!(err.contains("worktree unknown"), "{err}");
        assert!(err.contains("not the top of a git checkout"), "{err}");
        assert!(err.contains(outer.to_str().unwrap()), "{err}");

        fixture_git(&base, &["clone", "-q", bare.to_str().unwrap(), "own"]);
        let own = base.join("own");
        assert_eq!(worktree_gate(&outer, Some(own.to_str().unwrap())), Ok(()));
    }

    /// AC-17 (REQ-22): what a lane really records is judged by itself and not refused: a linked
    /// worktree made by `git worktree add` beside the checkout and one inside it (the layout of
    /// `.claude/worktrees`), and a link to a checkout, whose path git spells through the real one.
    #[test]
    fn a_linked_worktree_and_a_link_to_a_checkout_are_judged_by_themselves() {
        let tmp = tempfile::tempdir().unwrap();
        let base = tmp.path().canonicalize().unwrap();
        let (bare, outer) = (base.join("remote.git"), base.join("outer"));
        for d in [&bare, &outer] {
            std::fs::create_dir_all(d).unwrap();
        }
        fixture_git(&bare, &["init", "-q", "--bare", "-b", "main"]);
        fixture_git(&outer, &["init", "-q", "-b", "main"]);
        std::fs::write(outer.join("a.txt"), "a\n").unwrap();
        fixture_git(&outer, &["add", "a.txt"]);
        fixture_git(&outer, &["commit", "-q", "-m", "first"]);
        fixture_git(&outer, &["remote", "add", "origin", bare.to_str().unwrap()]);
        fixture_git(&outer, &["push", "-q", "origin", "main"]);

        let beside = base.join("beside");
        let inside = outer.join("inside");
        for (path, branch) in [(&beside, "lane-beside"), (&inside, "lane-inside")] {
            fixture_git(
                &outer,
                &[
                    "worktree",
                    "add",
                    "-q",
                    "-b",
                    branch,
                    path.to_str().unwrap(),
                ],
            );
            assert_eq!(worktree_gate(&outer, Some(path.to_str().unwrap())), Ok(()));
        }

        fixture_git(&base, &["clone", "-q", bare.to_str().unwrap(), "own"]);
        let link = base.join("link");
        std::os::unix::fs::symlink(base.join("own"), &link).unwrap();
        assert_eq!(worktree_gate(&outer, Some(link.to_str().unwrap())), Ok(()));
    }

    /// AC-17 (REQ-22): the checkout top git prints is cut only at its line break, so a real
    /// linked worktree whose folder name ends in a space is judged by itself.
    #[test]
    fn a_worktree_whose_name_ends_in_a_space_is_judged_by_itself() {
        let tmp = tempfile::tempdir().unwrap();
        let base = tmp.path().canonicalize().unwrap();
        let (bare, outer) = (base.join("remote.git"), base.join("outer"));
        for d in [&bare, &outer] {
            std::fs::create_dir_all(d).unwrap();
        }
        fixture_git(&bare, &["init", "-q", "--bare", "-b", "main"]);
        fixture_git(&outer, &["init", "-q", "-b", "main"]);
        std::fs::write(outer.join("a.txt"), "a\n").unwrap();
        fixture_git(&outer, &["add", "a.txt"]);
        fixture_git(&outer, &["commit", "-q", "-m", "first"]);
        fixture_git(&outer, &["remote", "add", "origin", bare.to_str().unwrap()]);
        fixture_git(&outer, &["push", "-q", "origin", "main"]);

        let spaced = base.join("lane ");
        fixture_git(
            &outer,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane-spaced",
                spaced.to_str().unwrap(),
            ],
        );
        assert_eq!(
            worktree_gate(&outer, Some(spaced.to_str().unwrap())),
            Ok(())
        );
    }
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
