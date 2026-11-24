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

/// The shared files of the project profile (`shared_files`, SPEC-20261124-shared-files-profile):
/// files many cards edit by a line, which a reservation never holds whole. Read from the bytes of
/// the profile a reader opens, with no check of the other keys, so a profile that is short or
/// has a key this build does not know still gives its list. Empty when the key is absent, is not
/// a list or the profile cannot be read; a list entry that is not a string is skipped.
pub fn shared_files(project: &Path) -> Vec<String> {
    let Ok(Some(bytes)) = crate::profile::chosen_profile(project) else {
        return Vec::new();
    };
    let Ok(value) = serde_json::from_slice::<serde_json::Value>(&bytes) else {
        return Vec::new();
    };
    value
        .get("shared_files")
        .and_then(serde_json::Value::as_array)
        .map(|list| {
            list.iter()
                .filter_map(|v| v.as_str())
                .map(str::to_string)
                .collect()
        })
        .unwrap_or_default()
}

/// The shared entry that `entry` overlaps (the one rule reservations use, `scope::overlap`), in
/// either direction: an entry that covers a shared file, a path inside a shared glob, the same
/// entry.
pub fn shared_overlap<'a>(shared: &'a [String], entry: &str) -> Option<&'a str> {
    shared
        .iter()
        .map(String::as_str)
        .find(|s| crate::scope::overlap(entry, s))
}

/// The command that writes a shared file in short: `card line-lease`.
pub fn line_lease_form(card: &str, generation: &str, path: &str) -> String {
    format!("card line-lease {card} --generation {generation} --path {path} -- <command>")
}

/// The refusal of an entry that overlaps a shared file (`action` is `take` or `scope`): the
/// entry, the shared entry and the way to write the file.
pub fn shared_refusal(action: &str, card: &str, entry: &str, shared: &str) -> String {
    format!(
        "card {action} refused: `{entry}` overlaps the shared file `{shared}` (`shared_files` of \
         the project profile); a shared file is never reserved whole, so leave it out of the \
         scope and write it in short with `{}`",
        line_lease_form(card, "<g>", entry)
    )
}

/// The first of `entries` that overlaps a shared entry, as the refusal of `action` for `card`.
pub fn refuse_shared(
    action: &str,
    card: &str,
    shared: &[String],
    entries: &[String],
) -> Result<(), String> {
    for entry in entries {
        if let Some(hit) = shared_overlap(shared, entry) {
            return Err(shared_refusal(action, card, entry, hit));
        }
    }
    Ok(())
}

/// What a take fills an empty-scope card with, and where it came from.
#[derive(Debug)]
pub struct Fill {
    pub scope: Vec<String>,
    pub source: Source,
    /// Paths a SPEC list named more than once; each is in `scope` once.
    pub merged: Vec<String>,
    /// Entries left out of `scope` because they overlap a shared file: the entry and the shared
    /// entry it overlaps.
    pub left_out: Vec<(String, String)>,
}

/// `scope` without the entries that overlap a shared file, and those entries. A fill that has
/// nothing left refuses, naming each entry and the way to write the file.
fn leave_out_shared(
    project: &Path,
    card: &str,
    scope: Vec<String>,
) -> Result<(Vec<String>, Vec<(String, String)>), String> {
    let shared = shared_files(project);
    if shared.is_empty() {
        return Ok((scope, Vec::new()));
    }
    let mut kept = Vec::with_capacity(scope.len());
    let mut left_out = Vec::new();
    for entry in scope {
        match shared_overlap(&shared, &entry) {
            Some(hit) => left_out.push((entry, hit.to_string())),
            None => kept.push(entry),
        }
    }
    if kept.is_empty() {
        if let Some((entry, hit)) = left_out.first() {
            return Err(format!(
                "{} (every path of the fill is a shared file, nothing is left to reserve; name \
                 the other paths with `card take {card} --scope <path>`)",
                shared_refusal("take", card, entry, hit)
            ));
        }
    }
    Ok((kept, left_out))
}

/// The warning that names a shared entry a fill left out.
pub fn left_out_warning(card: &str, entry: &str, shared: &str) -> String {
    format!(
        "{card}: `{entry}` overlaps the shared file `{shared}` and was left out of the scope; \
         write it in short with `{}`",
        line_lease_form(card, "<g>", entry)
    )
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
        let (scope, left_out) = leave_out_shared(project, &card, scope)?;
        return Ok(Some(Fill {
            scope,
            source: Source::Spec,
            merged,
            left_out,
        }));
    }
    let last =
        last_reservation(conn, card_id).map_err(|e| ledger::LedgerError::from(e).to_string())?;
    if last.is_empty() {
        return Ok(None);
    }
    let scope = super::dispatch::checked_scope("take", &last)?;
    let (scope, left_out) = leave_out_shared(project, &super::input::card_name(card_id), scope)?;
    Ok(Some(Fill {
        scope,
        source: Source::LastReservation,
        merged: Vec::new(),
        left_out,
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

#[cfg(test)]
mod tests {
    use super::*;

    /// AC-8 (SPEC-20261124-shared-files-profile REQ-9): the profile of this repository lists the
    /// files the leader registered, and each entry overlaps itself under the reservation rule.
    #[test]
    fn this_repository_lists_its_shared_files() {
        let list = shared_files(Path::new(env!("CARGO_MANIFEST_DIR")));
        for want in [
            "src/lib.rs",
            "src/help.rs",
            "src/ops/card.rs",
            "tests/areas.toml",
            "tools/repo_scans.json",
            "skills/moai/references/source-manifest.json",
            "docs/design-review/appendix/operations/card.md",
            "tests/fixtures/help/*.txt",
            // Leader note 8070 added the ninth: every card adds a key to the profile.
            "src/profile.rs",
        ] {
            assert!(
                list.iter().any(|e| e == want),
                "`{want}` is not listed: {list:?}"
            );
            assert_eq!(shared_overlap(&list, want), Some(want), "{want}");
        }
    }

    #[test]
    fn an_entry_overlaps_a_shared_file_either_way_round() {
        let list = vec![
            "src/lib.rs".to_string(),
            "tests/fixtures/help/*.txt".to_string(),
        ];
        assert_eq!(shared_overlap(&list, "src/**"), Some("src/lib.rs"));
        assert_eq!(
            shared_overlap(&list, "tests/fixtures/help/a.txt"),
            Some("tests/fixtures/help/*.txt")
        );
        assert_eq!(shared_overlap(&list, "src/other.rs"), None);
        assert_eq!(shared_overlap(&[], "src/lib.rs"), None);
    }
}
