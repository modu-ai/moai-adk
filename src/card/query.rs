//! What `moai card query` does (SPEC-20261126-card-read-queries): many cards, the reservations
//! that hold a card and notes named by number, read in one call that writes nothing.
//!
//! Every input is checked before the ledger is opened. The ledger is then opened read-only with
//! the short wait of `card timeline` ([`WAIT`]), every value is read inside one read transaction,
//! and the connection is closed before the answer is handed back to be printed. A card or a note
//! that does not exist refuses the whole call, so there is never a partial answer.

use super::command::{criteria_of, note_sessions, ABSENT_HINT};
use super::input::{self, CardRef};
use super::timeline::WAIT;
use super::wip_limit::current_head;
use crate::ledger::workflow::{self, VerdictEvent};
use crate::ledger::{self, reservation, LedgerError};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use std::collections::HashMap;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The most cards one call may name; a card named twice counts once.
pub const MAX_CARDS: usize = 50;
/// The most note numbers one call may name; a number given twice counts once.
pub const MAX_NOTE_IDS: usize = 10;
/// The most note headers `--notes` may ask for of each card.
pub const MAX_NOTE_HEADS: i64 = 20;
/// A note header's first line is cut to this many characters.
const LINE_MAX_CHARS: usize = 200;

/// The newest notes of one card without their bodies: `?1` card, `?2` row limit. The first line
/// is cut in the statement too, because a body can be 1 MiB.
const NOTE_HEADS_SQL: &str = "SELECT note_id, kind, author, origin, created_at, \
        length(CAST(body AS BLOB)), \
        substr(CASE WHEN instr(body, char(10)) > 0 \
                    THEN substr(body, 1, instr(body, char(10)) - 1) \
                    ELSE body END, 1, 2000) \
   FROM card_note WHERE card_id = ?1 ORDER BY note_id DESC LIMIT ?2";

/// One note by its number: `?1` note.
const NOTE_SQL: &str = "SELECT card_id, kind, author, origin, body, ref, created_at \
   FROM card_note WHERE note_id = ?1";

/// What `card query` was given. At least one of `cards`, `reservations` and `note_ids` is asked
/// for, and `notes` only with `cards`: the argument parser sees to both.
pub struct QueryRequest<'a> {
    /// The cards, each `C-<n>`.
    pub cards: &'a [String],
    /// Whether to list the reservations that hold a card.
    pub reservations: bool,
    /// The numbers of the notes to show whole.
    pub note_ids: &'a [i64],
    /// How many of each card's newest note headers to show (none when absent).
    pub notes: Option<i64>,
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card query refused: {why}")
}

fn db(e: rusqlite::Error) -> String {
    LedgerError::from(e).to_string()
}

/// `items` without the repeats, in the order given.
fn distinct(items: impl IntoIterator<Item = i64>) -> Vec<i64> {
    let mut out = Vec::new();
    for item in items {
        if !out.contains(&item) {
            out.push(item);
        }
    }
    out
}

/// The card numbers of `--card`.
fn card_ids(given: &[String]) -> Result<Vec<i64>> {
    let ids = given
        .iter()
        .map(
            |reference| match input::parse_ref(reference).map_err(refused)? {
                CardRef::Id(n) => Ok(n),
                CardRef::SourceNumber(_) => Err(refused("name the card by its C-number")),
            },
        )
        .collect::<Result<Vec<_>>>()?;
    let ids = distinct(ids);
    if ids.len() > MAX_CARDS {
        return Err(refused(format!(
            "--card takes at most {MAX_CARDS} cards in one call"
        )));
    }
    Ok(ids)
}

/// The title and state of a card, or the refusal of a card that is not there.
fn card_row(conn: &Connection, card_id: i64) -> Result<(String, String)> {
    conn.query_row(
        "SELECT title, state FROM card WHERE card_id = ?1",
        [card_id],
        |r| Ok((r.get(0)?, r.get(1)?)),
    )
    .optional()
    .map_err(db)?
    .ok_or_else(|| format!("card not found: there is no {}", input::card_name(card_id)))
}

fn verdict_json(v: &VerdictEvent) -> Value {
    let receipt = v.receipt.as_ref().map(|r| {
        let steps: Vec<Value> = r
            .steps
            .iter()
            .map(|s| json!({"name": s.name, "exit": s.exit}))
            .collect();
        json!({"head": r.head, "steps": steps})
    });
    json!({
        "event": v.event_id, "verdict": if v.pass { "pass" } else { "fail" },
        "head": v.head, "base": v.base, "tier": v.tier, "reviewer": v.reviewer,
        "evidence": v.evidence, "receipt": receipt,
    })
}

/// The card's verdicts, oldest first, and the number of its verdict rows that do not read. With
/// any such row no verdict is given: an older one must not pass for the newest.
fn verdicts_of(conn: &Connection, card_id: i64) -> Result<(Vec<VerdictEvent>, usize)> {
    let rows: i64 = conn
        .query_row(
            "SELECT count(*) FROM event WHERE card_id = ?1 AND kind = ?2",
            params![card_id, workflow::VERDICT_KIND],
            |r| r.get(0),
        )
        .map_err(db)?;
    let read = workflow::verdict_events(conn, card_id).map_err(|e| e.to_string())?;
    // `verdict_events` leaves unreadable rows out without a word, so the counts are compared.
    let unreadable = (rows as usize).saturating_sub(read.len());
    if unreadable > 0 {
        return Ok((Vec::new(), unreadable));
    }
    Ok((read, 0))
}

/// The headers of the card's newest `limit` notes, newest first.
fn note_heads(conn: &Connection, card_id: i64, limit: i64) -> Result<Vec<Value>> {
    let sessions = note_sessions(conn, card_id).map_err(db)?;
    let mut stmt = conn.prepare(NOTE_HEADS_SQL).map_err(db)?;
    let rows = stmt
        .query_map([card_id, limit], |r| {
            let note_id: i64 = r.get(0)?;
            let line: String = r.get(6)?;
            let line: String = line.trim_end().chars().take(LINE_MAX_CHARS).collect();
            Ok(json!({
                "note_id": note_id, "kind": r.get::<_, String>(1)?,
                "author": r.get::<_, String>(2)?, "origin": r.get::<_, String>(3)?,
                "session": sessions.get(&note_id).cloned().flatten(),
                "created_at": r.get::<_, i64>(4)?, "bytes": r.get::<_, i64>(5)?,
                "line": line,
            }))
        })
        .map_err(db)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(db)?;
    Ok(rows)
}

/// One entry of `cards`; `notes` is the number of note headers it carries (0: no list at all).
fn card_entry(conn: &Connection, card_id: i64, notes: i64) -> Result<Value> {
    // The first read of a card: a missing card is refused before anything else is read of it.
    let (title, state) = card_row(conn, card_id)?;
    let head = current_head(conn, card_id)?;
    let (verdicts, unreadable) = verdicts_of(conn, card_id)?;
    let on_head = head
        .as_deref()
        .and_then(|head| verdicts.iter().rev().find(|v| v.head == head));
    let criteria = criteria_of(conn, card_id).map_err(db)?;
    let count = |status: &str| criteria.iter().filter(|c| c["status"] == status).count();
    let open: Vec<Value> = criteria
        .iter()
        .filter(|c| c["status"] == "open")
        .map(|c| json!({"seq": c["seq"], "text": c["text"], "verify": c["verify"]}))
        .collect();
    let mut entry = json!({
        "card": input::card_name(card_id), "card_id": card_id,
        "title": title, "state": state, "head": head,
        "verdict": verdicts.last().map(verdict_json),
        "head_verdict": on_head.map(verdict_json),
        "verdicts_unreadable": unreadable,
        "criteria": {
            "open": count("open"), "met": count("met"), "withdrawn": count("withdrawn"),
        },
        "open_criteria": open,
    });
    if notes > 0 {
        entry["notes"] = json!(note_heads(conn, card_id, notes)?);
    }
    Ok(entry)
}

/// Every reservation that holds a card, by card number. `now` is the ledger's clock.
fn reservation_entries(conn: &Connection, now: i64) -> Result<Vec<Value>> {
    reservation::current_all(conn)
        .map_err(|e| e.to_string())?
        .into_iter()
        .map(|r| {
            let (title, state) = card_row(conn, r.card_id)?;
            Ok(json!({
                "card": input::card_name(r.card_id), "card_id": r.card_id,
                "title": title, "state": state, "status": r.status, "session": r.holder,
                "generation": r.generation, "acquired_at": r.acquired_at,
                "lease_until": r.lease_until,
                "lease_ended": reservation::lease_ended(r.lease_until, now),
                "worktree": r.worktree, "branch": r.branch,
            }))
        })
        .collect()
}

/// The notes `ids` name, whole, in that order.
fn whole_notes(conn: &Connection, ids: &[i64]) -> Result<Vec<Value>> {
    // The sessions of a card's notes are read once for the card, not once per note.
    let mut sessions: HashMap<i64, HashMap<i64, Option<String>>> = HashMap::new();
    let mut out = Vec::with_capacity(ids.len());
    for &note_id in ids {
        let (card_id, mut note) = conn
            .query_row(NOTE_SQL, [note_id], |r| {
                let card_id: i64 = r.get(0)?;
                let body: String = r.get(4)?;
                Ok((
                    card_id,
                    json!({
                        "note_id": note_id, "card": input::card_name(card_id),
                        "card_id": card_id, "kind": r.get::<_, String>(1)?,
                        "author": r.get::<_, String>(2)?, "origin": r.get::<_, String>(3)?,
                        "ref": r.get::<_, Option<String>>(5)?,
                        "created_at": r.get::<_, i64>(6)?,
                        "bytes": body.len(), "body": body,
                    }),
                ))
            })
            .optional()
            .map_err(db)?
            .ok_or_else(|| format!("note not found: there is no note {note_id}"))?;
        let of_card = match sessions.entry(card_id) {
            std::collections::hash_map::Entry::Occupied(e) => e.into_mut(),
            std::collections::hash_map::Entry::Vacant(e) => {
                e.insert(note_sessions(conn, card_id).map_err(db)?)
            }
        };
        note["session"] = json!(of_card.get(&note_id).cloned().flatten());
        out.push(note);
    }
    Ok(out)
}

/// `card query [--card C-n,…] [--reservations] [--note ID,…] [--notes N]`. Reads only. The
/// answer has a `cards`, `reservations` and `notes` list for each of the three asked for.
pub fn query(project: &Path, home: &Path, request: &QueryRequest<'_>) -> Result<Value> {
    let cards = card_ids(request.cards)?;
    let note_ids = distinct(request.note_ids.iter().copied());
    if note_ids.len() > MAX_NOTE_IDS {
        return Err(refused(format!(
            "--note takes at most {MAX_NOTE_IDS} note numbers in one call"
        )));
    }
    let heads = request.notes.unwrap_or(0);
    if !(0..=MAX_NOTE_HEADS).contains(&heads) {
        return Err(refused(format!("--notes must be 0 to {MAX_NOTE_HEADS}")));
    }
    let id = ledger::command::project_id_of(project)?;
    let mut answer = json!({
        "action": "query", "project_id": id, "ledger": "ready", "writes": false,
    });
    {
        let ledger = ledger::open::open_read_only_within(home, &id, WAIT).map_err(|e| match e {
            LedgerError::Absent => ABSENT_HINT.to_string(),
            other => other.to_string(),
        })?;
        let tx = ledger.conn().unchecked_transaction().map_err(db)?;
        let now: i64 = tx
            .query_row("SELECT unixepoch()", [], |r| r.get(0))
            .map_err(db)?;
        answer["read_at"] = json!(now);
        if !cards.is_empty() {
            let entries = cards
                .iter()
                .map(|&card_id| card_entry(&tx, card_id, heads))
                .collect::<Result<Vec<_>>>()?;
            answer["cards"] = json!(entries);
        }
        if request.reservations {
            answer["reservations"] = json!(reservation_entries(&tx, now)?);
        }
        if !note_ids.is_empty() {
            answer["notes"] = json!(whole_notes(&tx, &note_ids)?);
        }
        tx.finish().map_err(db)?;
    }
    Ok(answer)
}
