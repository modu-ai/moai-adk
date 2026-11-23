//! Where `card take` gets the write scope it reserves, and keeping that scope on the card
//! (SPEC-20261123-take-scope-stored).
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

/// What a take fills an empty-scope card with, and where it came from.
#[derive(Debug)]
pub struct Fill {
    pub scope: Vec<String>,
    pub source: Source,
    /// Paths a SPEC list named more than once; each is in `scope` once.
    pub merged: Vec<String>,
}

/// Whether a `write:` entry of a SPEC has the shape of a path or a glob: no white space, no comma,
/// no leading `-`, and a `/`, a glob character or a file extension. A SPEC's list is prose the
/// reader scanned for paths, so a span that is a word, an option or code (`the`, `--reason`,
/// `pub mod build;`) must not become the card's scope.
fn path_shaped(entry: &str) -> bool {
    if entry.is_empty()
        || entry.starts_with('-')
        || entry.contains(|c: char| c.is_whitespace() || c == ',')
    {
        return false;
    }
    if entry.contains(['/', '*', '?', '[', '{']) {
        return true;
    }
    // A bare file name: an extension of 1 to 10 letters or digits after the last dot.
    entry.rsplit_once('.').is_some_and(|(_, ext)| {
        (1..=10).contains(&ext.len()) && ext.bytes().all(|b| b.is_ascii_alphanumeric())
    })
}

/// `list` with each entry once (the first place kept), and the entries that were there again.
fn merge_repeats(list: &[String]) -> (Vec<String>, Vec<String>) {
    let mut kept: Vec<String> = Vec::with_capacity(list.len());
    let mut repeated: Vec<String> = Vec::new();
    for entry in list {
        let entry = entry.trim();
        if kept.iter().any(|k| k == entry) {
            if !repeated.iter().any(|r| r == entry) {
                repeated.push(entry.to_string());
            }
        } else {
            kept.push(entry.to_string());
        }
    }
    (kept, repeated)
}

/// The refusal for a SPEC list that cannot fill the scope: the file, the reason, the two ways out.
fn spec_refusal(card: &str, file: &str, why: &str) -> String {
    format!(
        "card take refused: the `write:` list of {file} cannot fill the scope of {card} ({why}); \
         correct that list, or name the paths with `card take {card} --scope <path>` (repeat it \
         per path), or set them first with `card scope {card} --scope <path>`"
    )
}

/// What a take would fill an empty-scope card with: the `write:` list of the SPEC the card names,
/// else the scope of its last reservation, each checked as a typed `--scope` is (the same limits
/// as `reserve`). `Ok(None)` when no source gives a path; `Err` when a source is there but cannot
/// be used (over a limit, or an entry that is not a path), so a SPEC that names the work in full
/// is never silently replaced by an older scope and a word is never stored as a path. A path a
/// SPEC list names twice is merged into one (decision D-4). `take` and `card next` share this
/// choice, so `next` offers exactly what `take` would accept.
pub fn choose(
    project: &Path,
    conn: &Connection,
    card_id: i64,
    spec: Option<&str>,
) -> Result<Option<Fill>, String> {
    let from_spec = spec_list(project, spec);
    if let (false, Some(id)) = (from_spec.is_empty(), spec) {
        let card = super::input::card_name(card_id);
        let file = format!("{}/{id}/spec.md", spec_doc::SPECS_DIR);
        let (list, merged) = merge_repeats(&from_spec);
        if let Some(i) = list.iter().position(|e| !path_shaped(e)) {
            let why = format!(
                "entry {} `{}` is not a path: a path has no space or comma, does not start with \
                 `-`, and has a `/`, a glob character or a file extension",
                i + 1,
                list[i]
            );
            return Err(spec_refusal(&card, &file, &why));
        }
        let scope = super::dispatch::checked_scope("take", &list).map_err(|e| {
            let why = e.strip_prefix("card take refused: ").unwrap_or(&e);
            spec_refusal(&card, &file, why)
        })?;
        return Ok(Some(Fill {
            scope,
            source: Source::Spec,
            merged,
        }));
    }
    let last =
        last_reservation(conn, card_id).map_err(|e| ledger::LedgerError::from(e).to_string())?;
    if last.is_empty() {
        return Ok(None);
    }
    let scope = super::dispatch::checked_scope("take", &last)?;
    Ok(Some(Fill {
        scope,
        source: Source::LastReservation,
        merged: Vec::new(),
    }))
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
