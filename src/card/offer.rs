//! `moai card offer` and `moai card accept` (SPEC-20261126-card-offer-accept): the command layer
//! over `ledger::offer`. It names the card, takes this session's own id from the environment
//! (never typed), checks the text the caller typed, and words the refusals as
//! `card conflict in <code>: …` so `contract::classify` reads them as a `conflict`.
//!
//! The target of an offer is told by a cross-session message from the offering session: the
//! ledger is the truth and the message is a notice (the SPEC's Non-goals).

use super::command::open_for_write;
use super::dispatch::{reservation_json, DEFAULT_LEASE_SECS};
use super::input::{self, CardRef};
use crate::ledger::offer::{self as ledger_offer, DEFAULT_VALID_SECS};
use crate::ledger::reservation::{MAX_LEASE_SECS, MIN_LEASE_SECS};
use crate::session_current;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The longest handover text a holder can attach to an offer.
const MAX_HANDOVER_LEN: usize = 500;

/// What `card offer` was given.
pub struct OfferRequest<'a> {
    /// The holder: this session's own id.
    pub holder: &'a str,
    /// The generation the holder holds.
    pub generation: i64,
    /// The session offered the card; none with `cancel`.
    pub to: Option<&'a str>,
    /// Void the newest offer instead of making one.
    pub cancel: bool,
    /// Seconds the offer stays valid (default [`DEFAULT_VALID_SECS`]).
    pub valid_secs: Option<i64>,
    /// What the receiver should know: the preserved work, the next step, open approvals and
    /// remote requests. Free text kept in the ledger for good, so it passes the card text screen.
    pub handover: Option<&'a str>,
}

/// What `card accept` was given.
pub struct AcceptRequest<'a> {
    /// The accepting session: this session's own id.
    pub accepter: &'a str,
    /// The generation the offer was made at.
    pub generation: i64,
    /// Lease of the new reservation in seconds (default [`DEFAULT_LEASE_SECS`]).
    pub lease_secs: Option<i64>,
}

fn card_id(action: &str, reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(|e| format!("card {action} refused: {e}"))? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(format!(
            "card {action} refused: name the card by its C-number (`card show #N` finds it)"
        )),
    }
}

fn usage(why: &str) -> String {
    format!("card offer refused: {why}")
}

/// A handover text: one line of 1 to [`MAX_HANDOVER_LEN`] characters without control characters
/// or line separators, passing the one card text screen. It is never repeated in a refusal.
fn checked_handover(text: &str) -> Result<String> {
    let kept = text.trim();
    if kept.is_empty()
        || kept.chars().count() > MAX_HANDOVER_LEN
        || text.chars().any(char::is_control)
        || text.chars().any(|c| matches!(c, '\u{2028}' | '\u{2029}'))
    {
        return Err(usage(&format!(
            "--handover must be one line of 1 to {MAX_HANDOVER_LEN} characters"
        )));
    }
    let words = input::secret_words(kept);
    if words > 0 {
        return Err(usage(&format!(
            "--handover has {}",
            input::secret_refusal(words)
        )));
    }
    Ok(kept.to_string())
}

/// `card offer C-n --generation G --to <session>` (or `--cancel`): see the module text.
pub fn offer(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &OfferRequest<'_>,
) -> Result<Value> {
    let card_id = card_id("offer", reference)?;
    if req.cancel {
        if req.to.is_some() || req.valid_secs.is_some() || req.handover.is_some() {
            return Err(usage(
                "--cancel takes no --to, --valid-secs or --handover: it voids the newest offer",
            ));
        }
        let (id, mut ledger) = open_for_write(project, home)?;
        let cancelled = ledger_offer::cancel(&mut ledger, card_id, req.holder, req.generation)
            .map_err(|e| e.to_string())?
            .map_err(|e| e.to_string())?;
        return Ok(json!({
            "action": "offer-cancel", "project_id": id, "card": format!("C-{card_id}"),
            "generation": cancelled.generation, "to": cancelled.to,
            "events": 1, "writes": true,
        }));
    }
    let Some(to) = req.to else {
        return Err(usage("name the session that is offered the card with --to"));
    };
    if !session_current::well_formed(to) {
        return Err(usage(
            "--to must be a session id (hex digits and `-`, 8 to 64 characters)",
        ));
    }
    let valid_secs = req.valid_secs.unwrap_or(DEFAULT_VALID_SECS);
    if !(MIN_LEASE_SECS..=MAX_LEASE_SECS).contains(&valid_secs) {
        return Err(usage(&format!(
            "--valid-secs must be {MIN_LEASE_SECS} to {MAX_LEASE_SECS} seconds"
        )));
    }
    let handover = req.handover.map(checked_handover).transpose()?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let offered = ledger_offer::offer(
        &mut ledger,
        card_id,
        req.holder,
        req.generation,
        to,
        valid_secs,
        handover.as_deref(),
    )
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "offer", "project_id": id, "card": format!("C-{card_id}"),
        "generation": offered.generation, "to": offered.to,
        "valid_secs": offered.valid_secs, "valid_until": offered.valid_until,
        "handover": offered.handover,
        "events": 1, "writes": true,
    }))
}

/// `card accept C-n --generation G`: see the module text. Every refusal writes nothing.
pub fn accept(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &AcceptRequest<'_>,
) -> Result<Value> {
    let card_id = card_id("accept", reference)?;
    let lease_secs = req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS);
    let (id, mut ledger) = open_for_write(project, home)?;
    let accepted = ledger_offer::accept(
        &mut ledger,
        card_id,
        req.accepter,
        req.generation,
        lease_secs,
    )
    .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "accept", "project_id": id,
        "reservation": reservation_json(&accepted.reservation),
        "previous": reservation_json(&accepted.previous),
        "state": accepted.card_state, "offer_event": accepted.offer_event,
        "events": 1, "writes": true,
    }))
}
