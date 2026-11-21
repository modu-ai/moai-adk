//! Where `card take` gets the write scope it reserves, and keeping that scope on the card
//! (SPEC-20261121-take-scope-stored).
//!
//! A card added without a scope is skipped by `card next` as `empty_scope`. Before this module a
//! `take --scope` kept the scope on the reservation only, so a released card was empty again and
//! the next taker had to run `card scope` first. The order of sources is the given `--scope`, the
//! card's own scope, the `write:` list of the SPEC the card names, and the scope of the card's
//! most recent reservation.

use crate::ledger::{self, WriteLedger};
use crate::spec_doc;
use rusqlite::{params, Connection, OptionalExtension};
use std::path::Path;

/// Where the scope of a take came from; the word is `scope_source` in the answer.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Source {
    /// `--scope` was given.
    Given,
    /// The card's own scope.
    Card,
    /// The `write:` list of the SPEC the card names.
    Spec,
    /// The scope of the card's most recent reservation.
    LastReservation,
}

impl Source {
    pub fn word(self) -> &'static str {
        match self {
            Source::Given => "given",
            Source::Card => "card",
            Source::Spec => "spec",
            Source::LastReservation => "last_reservation",
        }
    }
}

/// The card's own write scope, or `None` for a card that is not there.
pub fn own_scope(conn: &Connection, card_id: i64) -> rusqlite::Result<Option<Vec<String>>> {
    let text: Option<String> = conn
        .query_row(
            "SELECT write_scope FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .optional()?;
    Ok(text.map(|t| serde_json::from_str(&t).unwrap_or_default()))
}

/// The `write:` list of the SPEC `spec` names; empty when the card names none, the SPEC cannot
/// be read, or its list is empty. A SPEC that lives only on another branch is not seen.
pub fn spec_list(project: &Path, spec: Option<&str>) -> Vec<String> {
    let Some(id) = spec else {
        return Vec::new();
    };
    match spec_doc::find(project, id) {
        Ok(doc) => doc.scope.write,
        Err(_) => Vec::new(),
    }
}

/// The scope of the card's most recent reservation (the highest generation, released or not),
/// empty when the card was never reserved.
pub fn last_reservation(conn: &Connection, card_id: i64) -> rusqlite::Result<Vec<String>> {
    let mut stmt = conn.prepare(
        "SELECT pattern FROM reservation_scope WHERE reservation_id = \
         (SELECT reservation_id FROM reservation WHERE card_id = ?1 \
          ORDER BY generation DESC LIMIT 1) ORDER BY pattern",
    )?;
    let rows = stmt.query_map([card_id], |r| r.get::<_, String>(0))?;
    rows.collect()
}

/// The refusal for a card that has no scope and nothing to fill it from: both ways out.
pub fn no_source_message(card: &str) -> String {
    format!(
        "card take refused: {card} has no write scope and nothing to fill it from (no scope on \
         the card, no `write:` list in a linked SPEC, no earlier reservation); name the paths \
         with `card take {card} --scope <path>` (repeat it per path), or set them first with \
         `card scope {card} --scope <path>`"
    )
}

/// Keeps `scope` on the card as its own write scope, with one `card.scope` event whose actor is
/// `holder`. Only a card whose own scope is still empty is written, so a scope another session
/// set a moment ago is never replaced. Returns whether the card was written.
pub fn store(
    ledger: &mut WriteLedger,
    card_id: i64,
    scope: &[String],
    holder: &str,
    source: Source,
) -> Result<bool, String> {
    let json = serde_json::to_string(scope).map_err(|e| e.to_string())?;
    ledger
        .with_write_tx(|tx, now| {
            let changed = tx.execute(
                "UPDATE card SET write_scope = ?2, updated_at = ?3 \
                 WHERE card_id = ?1 AND write_scope = '[]'",
                params![card_id, json, now],
            )?;
            if changed == 0 {
                return Ok(false);
            }
            let key = ledger::new_event_key(tx, "card.scope", Some(card_id))?;
            let payload = serde_json::json!({
                "from": [], "to": scope, "source": source.word(), "by": "take",
            });
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, 'card.scope', ?5)",
                params![key, now, holder, card_id, payload.to_string()],
            )?;
            Ok(true)
        })
        .map_err(|e| e.to_string())
}
