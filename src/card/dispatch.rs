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
fn card_id(action: &str, reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(|e| format!("card {action} refused: {e}"))? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(format!(
            "card {action} refused: name the card by its C-number (`card show #N` finds it)"
        )),
    }
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
/// skipped by `card next`
/// (`empty_scope`), and nothing else sets a card's own scope (card C-53). Any other state is a
/// conflict that names it; nothing is written then.
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
                return Ok(Ok(json!({
                    "state": state, "held": true,
                    "generation": widened.reservation.generation,
                    "holder": widened.reservation.holder,
                    "from": widened.from, "write_scope": widened.reservation.write_scope,
                    "added": widened.added, "reason": reason, "rev": rev,
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
    let (id, mut ledger) = open_for_write(project, home)?;
    // A card that is not there has no SPEC; `reserve` answers `card not found` for it.
    let (spec, warning) = match spec_link::kind_and_spec(ledger.conn(), card_id)
        .map_err(|e| ledger::LedgerError::from(e).to_string())?
    {
        Some((kind, spec)) => {
            spec_link::status(project, &input::card_name(card_id), &kind, spec.as_deref())
        }
        None => (Value::Null, None),
    };
    if let (true, Some(why)) = (req.require_approved_spec, &warning) {
        return Err(format!(
            "card take refused: --require-approved-spec: {why}; nothing was taken"
        ));
    }
    let lease = req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS);
    let r = match given {
        Some(scope) => reservation::reserve(
            &mut ledger,
            card_id,
            req.holder,
            &scope,
            lease,
            req.worktree,
            req.branch,
        ),
        None => reservation::reserve_own_scope(
            &mut ledger,
            card_id,
            req.holder,
            lease,
            req.worktree,
            req.branch,
        ),
    }
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "take", "project_id": id, "reservation": reservation_json(&r),
        "state": "active", "spec": spec, "warnings": warning.into_iter().collect::<Vec<_>>(),
        "writes": true,
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
    Ok(json!({
        "action": "release", "project_id": id, "reservation": reservation_json(&out.reservation),
        "state": out.card_state, "writes": true,
    }))
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
fn checked_reason(action: &str, reason: &str) -> Result<String> {
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
    takeover_gates(project, home, card_id, req.to_holder)?;
    let early = early_reclaim::assess(project, home, card_id, req.to_holder);
    let reason = match &early {
        early_reclaim::Early::Proven(proof) => early_reason(&reason, &proof.text)?,
        _ => reason,
    };
    let (id, mut ledger) = open_for_write(project, home)?;
    let lease_secs = req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS);
    let taken = match &early {
        early_reclaim::Early::Proven(proof) => reservation::reclaim_early(
            &mut ledger,
            card_id,
            req.to_holder,
            req.generation,
            &by,
            &reason,
            lease_secs,
            &reservation::EarlyProof {
                holder: &proof.holder,
                detail: &proof.detail,
            },
        ),
        _ => reservation::reclaim(
            &mut ledger,
            card_id,
            req.to_holder,
            req.generation,
            &by,
            &reason,
            lease_secs,
        ),
    };
    let out = taken.map_err(|e| match (&e, &early) {
        // The lease is live and the proof is missing: say which part.
        (reservation::ReclaimError::LeaseLive { .. }, early_reclaim::Early::Unproven(why)) => {
            format!("{e}; an early reclaim needs proof that the holder is gone, and {why}")
        }
        _ => e.to_string(),
    })?;
    Ok(json!({
        "action": "reclaim", "project_id": id,
        "reservation": reservation_json(&out.reservation),
        "previous": reservation_json(&out.previous),
        "by": by, "reason": reason, "state": out.card_state,
        "events": 1, "writes": true,
    }))
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
fn takeover_gates(project: &Path, home: &Path, card_id: i64, to_holder: &str) -> Result<()> {
    let id = crate::ledger::command::project_id_of(project)
        .map_err(|e| format!("card reclaim refused: {e}"))?;
    // Read-only peek: an absent ledger reaches its own refusal at `open_for_write`, and the
    // peek never creates anything (`open_read_only`).
    let Ok(ro) = ledger::open_read_only(home, &id) else {
        return Ok(());
    };
    let current = reservation::current(ro.conn(), card_id).map_err(|e| e.to_string())?;
    let Some(cur) = current else {
        return Ok(()); // the ledger refuses with its own `no reservation`
    };
    if cur.holder == to_holder {
        return Ok(()); // a self-reclaim renews its own reservation
    }
    worktree_gate(cur.worktree.as_deref())
}

/// REQ-1..3: refuse a worktree that exists and holds uncommitted or unpushed work, go on
/// only when the path is known to be gone (nothing there can be orphaned), refuse as
/// unverifiable when the path cannot be told apart from "gone" or git cannot be read.
pub(super) fn worktree_gate(worktree: Option<&str>) -> Result<()> {
    let Some(wt) = worktree.filter(|w| !w.is_empty()) else {
        return Ok(());
    };
    let unreadable =
        || format!("card conflict in worktree unknown: the worktree {wt} could not be read");
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
        Ok(false) => return Ok(()),
        Ok(true) => {}
        Err(_) => return Err(unreadable()),
    }
    // One changed or untracked path per line, every untracked file on its own line (a folder
    // git would show as one entry could hide real work beside a tool cache). A listed tool
    // cache file is not work (SPEC-20261120-reclaim-dead-holder REQ-12).
    let uncommitted = git_output(path, &["status", "--porcelain=v1", "--untracked-files=all"])
        .ok_or_else(unreadable)?
        .lines()
        .filter(|l| !l.trim().is_empty() && !super::early_reclaim::is_tool_cache_entry(l))
        .count() as u64;
    let unpushed = git_output(path, &["rev-list", "--count", "HEAD", "--not", "--remotes"])
        .and_then(|n| n.trim().parse::<u64>().ok())
        .ok_or_else(unreadable)?;
    if uncommitted > 0 || unpushed > 0 {
        return Err(format!(
            "card conflict in worktree dirty: {wt} holds {uncommitted} uncommitted change(s) \
             and {unpushed} unpushed commit(s); a reclaim would orphan them"
        ));
    }
    Ok(())
}

/// The standard output of one read-only git command in `dir`: the environment is scrubbed like
/// every spawn (`identity::scrub_git_env`) and the meaning matches `lane_reset`'s probes — no
/// fetch, untracked files counted, `None` on any failure.
fn git_output(dir: &Path, args: &[&str]) -> Option<String> {
    // `MOAI_GIT_BIN` names the git program for a runner whose child carries no `PATH`
    // (the same seam as `MOAI_FORGE_GH_BIN`); everyone else spawns `git`.
    let program = std::env::var("MOAI_GIT_BIN").unwrap_or_else(|_| "git".to_string());
    let mut cmd = std::process::Command::new(program);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .stdin(std::process::Stdio::null())
        .output()
        .ok()?;
    if !out.status.success() {
        return None;
    }
    Some(String::from_utf8_lossy(&out.stdout).into_owned())
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
