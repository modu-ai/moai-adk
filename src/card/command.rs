//! What `moai card list | show | add` do (Issue #3070 C1, plan comment 6905829690). The
//! `src/ops/card.rs` file only parses arguments and calls these.
//!
//! `show` also prints the reservation that holds the card now (card C-53 S1), or `null`; `show`
//! and `list` print the SPEC id the card names (`spec`, SDD-4), or `null`.
//!
//! `list` and `show` open the ledger read-only (`open_read_only`: no folder, no file, no journal).
//! `add` is the only writer here: one `BEGIN IMMEDIATE` transaction that stores the card, its
//! criteria and one `card.add` event, with every time taken from the transaction's database clock.
//! `add` never creates the ledger file: a project without one gets `ledger not found` and the
//! two commands that create it (`ledger-register-plan`, `ledger-register-apply`).
//!
//! `export` (#3070 K2a) reads the ledger the same read-only way and writes the K0 export folder
//! into the directory it is given, nowhere else.
//!
//! Errors use the ledger's own message prefixes, so `contract::classify` keeps its kinds:
//! `ledger not found` is `not_found`, `ledger blocked` is `blocked`, and a refused input is
//! `card add refused: …` (`invalid_input`).

use super::export_db;
use super::export_dir::{self, DirError};
use super::export_format;
use super::input::{self, AddInput, CardRef};
use crate::ledger::{self, LedgerError};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The actor of a card made without `--by`.
pub const DEFAULT_ACTOR: &str = "user";
/// The longest `--by`, in characters.
pub const ACTOR_MAX_CHARS: usize = 80;

pub(super) const ABSENT_HINT: &str = "ledger not found: no ledger is registered for this project; \
run `moai ledger-register-plan`, then `moai ledger-register-apply`";

fn project_id(project: &Path) -> Result<String> {
    ledger::command::project_id_of(project)
}

/// The project id and a write connection to its existing ledger. An absent ledger is the
/// `ledger not found` answer with the two register commands; nothing is created.
pub fn open_for_write(project: &Path, home: &Path) -> Result<(String, ledger::WriteLedger)> {
    let id = project_id(project)?;
    let ledger = ledger::open_write(home, &id).map_err(|e| match e {
        LedgerError::Absent => ABSENT_HINT.to_string(),
        other => other.to_string(),
    })?;
    Ok((id, ledger))
}

fn card_json(conn: &Connection, card_id: i64) -> std::result::Result<Option<Value>, LedgerError> {
    let row = conn
        .query_row(
            "SELECT card_id, title, body, kind, state, priority, write_scope, rev, created_by, \
                    created_at, updated_at, closed_at, source_host, source_project, \
                    source_number, source_url \
               FROM card WHERE card_id = ?1",
            [card_id],
            |r| {
                let host: Option<String> = r.get(12)?;
                let proj: Option<String> = r.get(13)?;
                let number: Option<i64> = r.get(14)?;
                let id: i64 = r.get(0)?;
                let scope: String = r.get(6)?;
                Ok(json!({
                    "name": input::card_name(id),
                    "card_id": id,
                    "title": r.get::<_, String>(1)?,
                    "body": r.get::<_, String>(2)?,
                    "kind": r.get::<_, String>(3)?,
                    "state": r.get::<_, String>(4)?,
                    "priority": r.get::<_, String>(5)?,
                    "write_scope": serde_json::from_str::<Value>(&scope).unwrap_or(Value::Null),
                    "rev": r.get::<_, i64>(7)?,
                    "created_by": r.get::<_, String>(8)?,
                    "created_at": r.get::<_, i64>(9)?,
                    "updated_at": r.get::<_, i64>(10)?,
                    "closed_at": r.get::<_, Option<i64>>(11)?,
                    "source": match (host, proj, number) {
                        (Some(h), Some(p), Some(n)) => json!(format!("{h}:{p}#{n}")),
                        _ => Value::Null,
                    },
                    "source_url": r.get::<_, Option<String>>(15)?,
                }))
            },
        )
        .optional()?;
    Ok(row)
}

/// `card list [--state S]`: every card, oldest first. An absent ledger is "no cards" (exit 0).
pub fn list(project: &Path, home: &Path, state: Option<&str>) -> Result<Value> {
    list_filtered(project, home, state, None, false)
}

/// A `--title-prefix` typed for `card list` (card C-99, SPEC-20261119-flaky-followups): not
/// empty and without a control character. It is taken as typed, not trimmed: `flaky: ` ends in
/// the space that tells the rule's title from `flakiness`.
fn title_prefix(given: &str) -> std::result::Result<&str, &'static str> {
    if given.is_empty() {
        return Err("the title prefix is empty");
    }
    if given.chars().any(char::is_control) {
        return Err("the title prefix has a control character");
    }
    Ok(given)
}

/// `card list [--state S] [--title-prefix P]`: the cards in that state whose title starts with
/// `P` (case-sensitive), oldest first, each card in the shape `list` gives. Both options are
/// checked before the ledger is read.
pub fn list_filtered(
    project: &Path,
    home: &Path,
    state: Option<&str>,
    title_prefix_given: Option<&str>,
    all: bool,
) -> Result<Value> {
    let state = state
        .map(|s| input::state(s).map_err(|e| format!("card list refused: {e}")))
        .transpose()?;
    let prefix = title_prefix_given
        .map(|p| title_prefix(p).map_err(|e| format!("card list refused: {e}")))
        .transpose()?;
    let id = project_id(project)?;
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Ok(json!({
                "action": "list", "project_id": id, "ledger": "absent",
                "count": 0, "cards": [], "writes": false,
            }));
        }
        Err(e) => return Err(e.to_string()),
    };
    let conn = ledger.conn();
    let mut stmt = conn
        .prepare(&format!(
            "SELECT card_id, title, kind, state, priority, rev, updated_at, \
                    source_host, source_project, source_number, {} \
               FROM card WHERE (?1 IS NULL OR state = ?1) \
                           AND (?2 IS NULL OR substr(title, 1, length(?2)) = ?2) \
                           AND (?3 OR NOT EXISTS (SELECT 1 FROM event \
                             WHERE event.card_id = card.card_id \
                               AND event.kind = 'card.withdraw')) \
              ORDER BY card_id",
            super::spec_link::LINKED_COLUMN
        ))
        .map_err(|e| LedgerError::from(e).to_string())?;
    let rows = stmt
        .query_map(params![state, prefix, all], |r| {
            let host: Option<String> = r.get(7)?;
            let proj: Option<String> = r.get(8)?;
            let number: Option<i64> = r.get(9)?;
            let id: i64 = r.get(0)?;
            Ok(json!({
                "name": input::card_name(id),
                "card_id": id,
                "title": r.get::<_, String>(1)?,
                "kind": r.get::<_, String>(2)?,
                "state": r.get::<_, String>(3)?,
                "priority": r.get::<_, String>(4)?,
                "rev": r.get::<_, i64>(5)?,
                "updated_at": r.get::<_, i64>(6)?,
                "source": match (host, proj, number) {
                    (Some(h), Some(p), Some(n)) => json!(format!("{h}:{p}#{n}")),
                    _ => Value::Null,
                },
                "spec": r.get::<_, Option<String>>(10)?,
            }))
        })
        .map_err(|e| LedgerError::from(e).to_string())?;
    let mut cards = Vec::new();
    for row in rows {
        cards.push(row.map_err(|e| LedgerError::from(e).to_string())?);
    }
    Ok(json!({
        "action": "list", "project_id": id, "ledger": "ready",
        "count": cards.len(), "cards": cards, "writes": false,
    }))
}

/// The notes of a card in note order. It never names `event`: the session of a note comes from
/// [`NOTE_EVENTS_SQL`], read once.
const NOTES_SQL: &str = "SELECT note_id, kind, author, origin, body, ref, created_at \
       FROM card_note WHERE card_id = ?1 ORDER BY note_id";

/// The `card.note` events of one card in event order: the note each names and the session it
/// records. One statement with no subquery, so `event` is read once whatever the number of notes.
const NOTE_EVENTS_SQL: &str = "SELECT json_extract(payload, '$.note_id'), \
            json_extract(payload, '$.session') \
       FROM event WHERE card_id = ?1 AND kind = 'card.note' ORDER BY event_id";

/// The session of each note of `card_id`, from its `card.note` events: the first event (lowest
/// event id) that names a note decides, even when its session is `null`; a note no event names
/// is not in the map. A note id that is not an integer, or a session that is neither text nor
/// `null`, is skipped (the tool writes neither).
fn note_sessions(
    conn: &Connection,
    card_id: i64,
) -> std::result::Result<HashMap<i64, Option<String>>, rusqlite::Error> {
    use rusqlite::types::Value;
    let mut stmt = conn.prepare(NOTE_EVENTS_SQL)?;
    let rows = stmt.query_map([card_id], |r| {
        Ok((r.get::<_, Value>(0)?, r.get::<_, Value>(1)?))
    })?;
    let mut sessions = HashMap::new();
    for row in rows {
        if let (Value::Integer(note_id), session) = row? {
            let session = match session {
                Value::Text(text) => Some(text),
                Value::Null => None,
                _ => continue,
            };
            sessions.entry(note_id).or_insert(session);
        }
    }
    Ok(sessions)
}

/// `card show C-n | #number`: the card with its criteria, notes, links and the reservation that
/// holds it now.
/// Writes the ledger. Withdraw a draft or ready card: the card moves to `hold` and one
/// `card.withdraw` event records the reason and the by, with the session that wrote it. The
/// ledger schema is unchanged (the six state values are pinned; the marker is the free event
/// kind). `card take` then refuses the card naming the reason, and the default `card list`
/// hides it.
pub fn withdraw(
    project: &Path,
    home: &Path,
    reference: &str,
    reason: &str,
    by: &str,
) -> Result<Value> {
    let which = input::parse_ref(reference).map_err(|e| format!("card withdraw refused: {e}"))?;
    let id = project_id(project)?;
    let mut ledger = ledger::open_write(home, &id).map_err(|e| match e {
        LedgerError::Absent => ABSENT_HINT.to_string(),
        other => other.to_string(),
    })?;
    let card_id = match which {
        CardRef::Id(n) => n,
        CardRef::SourceNumber(n) => {
            let mut stmt = ledger
                .conn()
                .prepare("SELECT card_id FROM card WHERE source_number = ?1 ORDER BY card_id")
                .map_err(|e| LedgerError::from(e).to_string())?;
            let ids: Vec<i64> = stmt
                .query_map([n], |r| r.get(0))
                .map_err(|e| LedgerError::from(e).to_string())?
                .collect::<std::result::Result<Vec<_>, _>>()
                .map_err(|e| LedgerError::from(e).to_string())?;
            match ids.as_slice() {
                [one] => *one,
                _ => {
                    return Err(format!(
                        "card withdraw refused: #{n} names {} cards; name the card instead",
                        ids.len()
                    ))
                }
            }
        }
    };
    let session = std::env::var("CLAUDE_CODE_SESSION_ID")
        .or_else(|_| std::env::var("CODEX_THREAD_ID"))
        .ok();
    ledger::workflow::withdraw(&mut ledger, card_id, reason, by, session.as_deref())
        .map_err(|e| format!("card withdraw refused: {e}"))?;
    Ok(json!({
        "action": "withdraw", "project_id": id,
        "card": input::card_name(card_id), "state": "hold",
        "writes": true,
    }))
}

pub fn show(project: &Path, home: &Path, reference: &str) -> Result<Value> {
    let which = input::parse_ref(reference).map_err(|e| format!("card show refused: {e}"))?;
    let id = project_id(project)?;
    let ledger = ledger::open_read_only(home, &id).map_err(|e| match e {
        LedgerError::Absent => ABSENT_HINT.to_string(),
        other => other.to_string(),
    })?;
    let conn = ledger.conn();
    let sql_err = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let card_id = match which {
        CardRef::Id(n) => n,
        CardRef::SourceNumber(n) => {
            let mut stmt = conn
                .prepare("SELECT card_id FROM card WHERE source_number = ?1 ORDER BY card_id")
                .map_err(sql_err)?;
            let ids = stmt
                .query_map([n], |r| r.get::<_, i64>(0))
                .map_err(sql_err)?
                .collect::<std::result::Result<Vec<_>, _>>()
                .map_err(sql_err)?;
            match ids.as_slice() {
                [one] => *one,
                [] => return Err(format!("card not found: no card came from #{n}")),
                many => {
                    let names: Vec<String> = many.iter().map(|c| input::card_name(*c)).collect();
                    return Err(format!(
                        "card show refused: #{n} matches {} cards ({}); name one by its C-number",
                        many.len(),
                        names.join(", ")
                    ));
                }
            }
        }
    };
    let mut card = card_json(conn, card_id)
        .map_err(|e| e.to_string())?
        .ok_or_else(|| format!("card not found: there is no {}", input::card_name(card_id)))?;

    let mut stmt = conn
        .prepare(
            "SELECT seq, text, verify, status, evidence_ref, withdrawn_reason, updated_at \
               FROM card_ac WHERE card_id = ?1 ORDER BY seq",
        )
        .map_err(sql_err)?;
    let criteria = stmt
        .query_map([card_id], |r| {
            Ok(json!({
                "seq": r.get::<_, i64>(0)?, "text": r.get::<_, String>(1)?,
                "verify": r.get::<_, String>(2)?, "status": r.get::<_, String>(3)?,
                "evidence_ref": r.get::<_, Option<String>>(4)?,
                "withdrawn_reason": r.get::<_, Option<String>>(5)?,
                "updated_at": r.get::<_, i64>(6)?,
            }))
        })
        .map_err(sql_err)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(sql_err)?;

    let mut stmt = conn.prepare(NOTES_SQL).map_err(sql_err)?;
    // `session`: the session that wrote the note, from its `card.note` event (`null` for a note
    // written without one, and for an imported note). The events are read once for the card,
    // not once per note: `event` has no index on `card_id`, so a per-note lookup scanned the
    // whole table for every note and slowed other sessions' writes while a big card was read.
    let sessions = note_sessions(conn, card_id).map_err(sql_err)?;
    let notes = stmt
        .query_map([card_id], |r| {
            let note_id = r.get::<_, i64>(0)?;
            Ok(json!({
                "note_id": note_id, "kind": r.get::<_, String>(1)?,
                "author": r.get::<_, String>(2)?, "origin": r.get::<_, String>(3)?,
                "body": r.get::<_, String>(4)?, "ref": r.get::<_, Option<String>>(5)?,
                "created_at": r.get::<_, i64>(6)?,
                "session": sessions.get(&note_id).cloned().flatten(),
            }))
        })
        .map_err(sql_err)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(sql_err)?;

    let mut stmt = conn
        .prepare(
            "SELECT link_id, kind, to_card, target_host, target_project, target_number, \
                    pr_state, merge_ref, state \
               FROM card_link WHERE from_card = ?1 ORDER BY link_id",
        )
        .map_err(sql_err)?;
    let links = stmt
        .query_map([card_id], |r| {
            let to: Option<i64> = r.get(2)?;
            let host: Option<String> = r.get(3)?;
            let proj: Option<String> = r.get(4)?;
            let number: Option<i64> = r.get(5)?;
            Ok(json!({
                "link_id": r.get::<_, i64>(0)?, "kind": r.get::<_, String>(1)?,
                "to": match (to, host, proj, number) {
                    (Some(c), _, _, _) => json!(input::card_name(c)),
                    (None, Some(h), Some(p), Some(n)) => json!(format!("{h}:{p}#{n}")),
                    _ => Value::Null,
                },
                "pr_state": r.get::<_, Option<String>>(6)?,
                "merge_ref": r.get::<_, Option<String>>(7)?,
                "state": r.get::<_, String>(8)?,
            }))
        })
        .map_err(sql_err)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(sql_err)?;

    card["action"] = json!("show");
    card["project_id"] = json!(id);
    card["criteria"] = json!(criteria);
    card["notes"] = json!(notes);
    card["links"] = json!(links);
    card["spec"] = json!(super::spec_link::kind_and_spec(conn, card_id)
        .map_err(sql_err)?
        .and_then(|(_, spec)| spec));
    card["reservation"] = ledger::reservation::current(conn, card_id)
        .map_err(|e| e.to_string())?
        .map_or(Value::Null, |r| super::dispatch::reservation_json(&r));
    card["writes"] = json!(false);
    Ok(card)
}

/// What `card add` was given.
pub struct AddRequest<'a> {
    pub title: &'a str,
    pub body: &'a str,
    pub kind: Option<&'a str>,
    pub priority: Option<&'a str>,
    pub scope: &'a [String],
    pub criteria: &'a [String],
    pub by: Option<&'a str>,
}

/// `--by`: one line of 1 to [`ACTOR_MAX_CHARS`] characters, `user` when not given.
pub fn actor(given: Option<&str>, action: &str) -> Result<String> {
    let by = given.unwrap_or(DEFAULT_ACTOR);
    let chars = by.chars().count();
    if by.trim().is_empty() || chars > ACTOR_MAX_CHARS || by.chars().any(char::is_control) {
        return Err(format!(
            "card {action} refused: --by must be one line of 1 to {ACTOR_MAX_CHARS} characters"
        ));
    }
    Ok(by.to_string())
}

/// `card add`: a new `draft` card with its criteria and one `card.add` event, in one transaction.
pub fn add(project: &Path, home: &Path, req: &AddRequest<'_>) -> Result<Value> {
    let card = input::validate_add(&AddInput {
        title: req.title,
        body: req.body,
        kind: req.kind,
        priority: req.priority,
        scope: req.scope,
        criteria: req.criteria,
    })
    .map_err(|e| format!("card add refused: {e}"))?;
    // The reservation limits too (64 entries, 512 characters each), so the card can be taken
    // with its own scope (card C-53 audit N4).
    super::dispatch::checked_scope("add", req.scope)?;
    let by = actor(req.by, "add")?;
    let (id, mut ledger) = open_for_write(project, home)?;
    let (card_id, rev) = ledger
        .with_write_tx(|tx, now| {
            tx.execute(
                "INSERT INTO card (title, body, kind, priority, write_scope, created_by, \
                                   created_at, updated_at) \
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?7)",
                params![
                    card.title,
                    card.body,
                    card.kind,
                    card.priority,
                    card.write_scope,
                    by,
                    now
                ],
            )?;
            let card_id = tx.last_insert_rowid();
            for ac in &card.criteria {
                tx.execute(
                    "INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES (?1, ?2, ?3, ?4)",
                    params![card_id, ac.seq, ac.text, now],
                )?;
            }
            let key = ledger::new_event_key(tx, "card.add", Some(card_id))?;
            let payload = json!({"title": card.title, "criteria": card.criteria.len()});
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, 'card.add', ?5)",
                params![key, now, by, card_id, payload.to_string()],
            )?;
            let rev: i64 =
                tx.query_row("SELECT rev FROM card WHERE card_id = ?1", [card_id], |r| {
                    r.get(0)
                })?;
            Ok((card_id, rev))
        })
        .map_err(|e| e.to_string())?;
    Ok(json!({
        "action": "add", "project_id": id,
        "name": input::card_name(card_id), "card_id": card_id,
        "state": "draft", "kind": card.kind, "priority": card.priority,
        "criteria": card.criteria.len(), "rev": rev, "created_by": by,
        "events": 1, "writes": true,
    }))
}

/// `card export --out DIR` (#3070 K2a): the five exported tables of the ledger as an export
/// folder. The ledger is opened read-only and only read; the files go into `out` and nowhere else
/// (an empty or new folder, nothing replaced: `export_dir::write`), private to the owner (0700
/// folder, 0600 files) and never inside the state folder `<home>/.moai`. Links in the parents of
/// `out` are followed and the real path is the one written and reported (`out`; `out_given` is
/// what was typed). A project without a ledger gets `ledger not found` and nothing is made,
/// `out` included.
pub fn export(project: &Path, home: &Path, out: &Path) -> Result<Value> {
    let id = project_id(project)?;
    let ledger = ledger::open_read_only(home, &id).map_err(|e| match e {
        LedgerError::Absent => ABSENT_HINT.to_string(),
        other => other.to_string(),
    })?;
    let snapshot =
        export_db::read_snapshot(ledger.conn(), ledger.project_id(), ledger.schema_version())
            .map_err(|e| e.to_string())?;
    let rendered = export_format::render(&snapshot)
        .map_err(|e| format!("card export refused: the ledger's rows cannot be exported: {e}"))?;
    // Where the folder really is: links in its parents are followed, and a folder inside the
    // state folder (`<home>/.moai`, where the ledgers are) is refused. This path is written to.
    let real_out = export_dir::resolve_out(out, &crate::home_workdir::root(home))
        .map_err(|e| format!("card export refused: {e}"))?;
    export_dir::write(&real_out, &rendered).map_err(|e| match e {
        DirError::Io { .. } => format!(
            "card export write failed: {e}. Files already written stay in {} (it has no \
             index.json, so it is not an export); nothing was removed",
            real_out.display()
        ),
        other => format!("card export refused: {other}"),
    })?;
    Ok(json!({
        "action": "export", "project_id": id,
        "out": real_out.display().to_string(), "out_given": out.display().to_string(),
        "cards": snapshot.cards.len(), "acs": snapshot.acs.len(),
        "notes": snapshot.notes.len(), "links": snapshot.links.len(),
        "events": snapshot.events.len(),
        "files": rendered.cards.len() + 2,
        "index_sha256": format!("{:x}", Sha256::digest(rendered.index.as_bytes())),
        "writes": false, "writes_files": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ledger::schema;

    fn ledger_with_events(events: &[(&str, i64, &str)]) -> Connection {
        let mut conn = Connection::open_in_memory().unwrap();
        schema::create_v1(&mut conn, "unit", "test").unwrap();
        conn.execute(
            "INSERT INTO card (card_id, title, created_by, created_at, updated_at) \
             VALUES (1, 't', 'x', 1, 1), (2, 'u', 'x', 1, 1)",
            [],
        )
        .unwrap();
        for (key, card, payload) in events {
            conn.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, 1, 'x', ?2, 'card.note', ?3)",
                params![key, card, payload],
            )
            .unwrap();
        }
        conn
    }

    fn plan(conn: &Connection, sql: &str) -> Vec<String> {
        let mut stmt = conn.prepare(&format!("EXPLAIN QUERY PLAN {sql}")).unwrap();
        stmt.query_map([1], |r| r.get::<_, String>(3))
            .unwrap()
            .map(|row| row.unwrap())
            .collect()
    }

    /// REQ-1, REQ-3: the first event of the card that names a note decides, even with a `null`
    /// session; another card's event and a note no event names add nothing; and the plans show
    /// `event` read once by the events query and not at all by the notes query, with no
    /// subquery in either.
    #[test]
    fn the_note_events_are_read_once() {
        let conn = ledger_with_events(&[
            ("k1", 1, r#"{"note_id":1,"session":"s-one"}"#),
            ("k2", 1, r#"{"note_id":1,"session":"s-later"}"#),
            ("k3", 1, r#"{"note_id":2,"session":null}"#),
            ("k4", 1, r#"{"note_id":2,"session":"s-after-null"}"#),
            ("k5", 2, r#"{"note_id":3,"session":"other-card"}"#),
            ("k6", 1, r#"{"note_id":"7","session":"text-id"}"#),
            ("k7", 1, r#"{"note_id":4,"session":7}"#),
        ]);
        conn.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES ('k8', 1, 'x', 1, 'card.scope', '{\"note_id\":5,\"session\":\"wrong-kind\"}')",
            [],
        )
        .unwrap();
        let sessions = note_sessions(&conn, 1).unwrap();
        assert_eq!(sessions.get(&1), Some(&Some("s-one".to_string())));
        assert_eq!(
            sessions.get(&2),
            Some(&None),
            "a null first session stays null"
        );
        assert_eq!(sessions.get(&3), None, "another card's event is not read");
        assert_eq!(
            sessions.get(&5),
            None,
            "an event of another kind is not read"
        );
        assert_eq!(sessions.get(&4), None, "a numeric session is skipped");
        assert_eq!(sessions.len(), 2, "{sessions:?}");

        let events = plan(&conn, NOTE_EVENTS_SQL);
        assert!(
            events.iter().all(|l| !l.contains("SUBQUERY")),
            "no subquery: {events:?}"
        );
        assert_eq!(
            events.iter().filter(|l| l.contains("event")).count(),
            1,
            "event is read once: {events:?}"
        );
        let notes = plan(&conn, NOTES_SQL);
        assert!(
            notes
                .iter()
                .all(|l| !l.contains("SUBQUERY") && !l.contains("event")),
            "the notes query leaves event alone: {notes:?}"
        );
    }
}
