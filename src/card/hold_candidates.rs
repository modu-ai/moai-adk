//! `moai card hold-candidates`: the open cards whose body or notes say HOLD in free text and that
//! have no structured HOLD in force (SPEC-20261121-card-hold-gate REQ-11). The list is read-only
//! and a hint: no code moves a free-text HOLD into the ledger, and `card next` and `card take`
//! read the `card.hold` event only. A person looks at a candidate and, if it is a real HOLD,
//! runs `card hold` for it.

use crate::ledger::{self, workflow, LedgerError};
use serde_json::{json, Value};
use std::path::Path;

/// The longest quote of a matching line, in characters.
const QUOTE_CHARS: usize = 160;

/// Whether `text` has the word `HOLD` in capitals: not inside a longer word, `_` counting as a
/// word character, so `HOLDS` and `NO_HOLD` are not matches.
fn has_hold_word(text: &str) -> bool {
    text.match_indices("HOLD").any(|(i, word)| {
        let is_word = |c: char| c.is_ascii_alphanumeric() || c == '_';
        let before = text[..i].chars().next_back().is_some_and(is_word);
        let after = text[i + word.len()..].chars().next().is_some_and(is_word);
        !before && !after
    })
}

/// The first line of `text` that has the word, trimmed and cut to [`QUOTE_CHARS`] characters.
fn quote(text: &str) -> String {
    let line = text
        .lines()
        .find(|l| has_hold_word(l))
        .unwrap_or_default()
        .trim();
    line.chars().take(QUOTE_CHARS).collect()
}

/// The candidates: one entry per open card (`draft`, `ready`, `active` or `review`) with no
/// `card.hold` in force, naming the newest note that says HOLD, else the card body. A project
/// without a ledger has none. Nothing is written.
pub fn candidates(project: &Path, home: &Path) -> Result<Value, String> {
    let id = ledger::command::project_id_of(project)?;
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Ok(json!({
                "action": "hold-candidates", "project_id": id, "ledger": "absent",
                "candidates": [], "count": 0, "writes": false,
            }));
        }
        Err(e) => return Err(e.to_string()),
    };
    let conn = ledger.conn();
    let sql_err = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT card_id, title, state, body FROM card \
             WHERE state IN ('draft','ready','active','review') ORDER BY card_id",
        )
        .map_err(sql_err)?;
    let cards: Vec<(i64, String, String, String)> = stmt
        .query_map([], |r| Ok((r.get(0)?, r.get(1)?, r.get(2)?, r.get(3)?)))
        .map_err(sql_err)?
        .collect::<Result<_, _>>()
        .map_err(sql_err)?;
    let mut found: Vec<Value> = Vec::new();
    for (card_id, title, state, body) in cards {
        if workflow::current_hold(conn, card_id)
            .map_err(sql_err)?
            .is_some()
        {
            continue;
        }
        let mut notes = conn
            .prepare("SELECT note_id, body FROM card_note WHERE card_id = ?1 ORDER BY note_id DESC")
            .map_err(sql_err)?;
        let from_note: Option<(i64, String)> = notes
            .query_map([card_id], |r| Ok((r.get(0)?, r.get(1)?)))
            .map_err(sql_err)?
            .collect::<Result<Vec<(i64, String)>, _>>()
            .map_err(sql_err)?
            .into_iter()
            .find(|(_, text)| has_hold_word(text));
        let (source, note_id, line) = match from_note {
            Some((note_id, text)) => ("note", json!(note_id), quote(&text)),
            None if has_hold_word(&body) => ("body", Value::Null, quote(&body)),
            None => continue,
        };
        found.push(json!({
            "name": crate::card::input::card_name(card_id), "card_id": card_id,
            "title": title, "state": state, "source": source, "note_id": note_id,
            "quote": line,
        }));
    }
    Ok(json!({
        "action": "hold-candidates", "project_id": id, "ledger": "ready",
        "count": found.len(), "candidates": found, "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_word_is_matched_whole_and_in_capitals() {
        assert!(has_hold_word("HOLD: wait"));
        assert!(has_hold_word("a card on HOLD."));
        assert!(has_hold_word("(HOLD)"));
        assert!(!has_hold_word("HOLDS"));
        assert!(!has_hold_word("NO_HOLD"));
        assert!(!has_hold_word("holdout and hold"));
    }

    #[test]
    fn the_quote_is_the_first_matching_line_cut_short() {
        assert_eq!(quote("first\n  HOLD: money  \nHOLD again"), "HOLD: money");
        let long = format!("HOLD {}", "x".repeat(300));
        assert_eq!(quote(&long).chars().count(), QUOTE_CHARS);
    }
}
