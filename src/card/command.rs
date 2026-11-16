//! What `moai card list | show | add` do (Issue #3070 C1, plan comment 6905829690). The
//! `src/ops/card.rs` file only parses arguments and calls these.
//!
//! `show` also prints the reservation that holds the card now (card C-53 S1), or `null`.
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
    let state = state
        .map(|s| input::state(s).map_err(|e| format!("card list refused: {e}")))
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
        .prepare(
            "SELECT card_id, title, kind, state, priority, rev, updated_at, \
                    source_host, source_project, source_number \
               FROM card WHERE (?1 IS NULL OR state = ?1) ORDER BY card_id",
        )
        .map_err(|e| LedgerError::from(e).to_string())?;
    let rows = stmt
        .query_map([state], |r| {
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

/// `card show C-n | #number`: the card with its criteria, notes, links and the reservation that
/// holds it now.
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

    let mut stmt = conn
        .prepare(
            // `session`: the session that wrote the note, from its `card.note` event (`null`
            // for a note written without one, and for an imported note).
            "SELECT note_id, kind, author, origin, body, ref, created_at, \
                    (SELECT json_extract(e.payload, '$.session') FROM event e \
                      WHERE e.card_id = card_note.card_id AND e.kind = 'card.note' \
                        AND json_extract(e.payload, '$.note_id') = card_note.note_id \
                      ORDER BY e.event_id LIMIT 1) \
               FROM card_note WHERE card_id = ?1 ORDER BY note_id",
        )
        .map_err(sql_err)?;
    let notes = stmt
        .query_map([card_id], |r| {
            Ok(json!({
                "note_id": r.get::<_, i64>(0)?, "kind": r.get::<_, String>(1)?,
                "author": r.get::<_, String>(2)?, "origin": r.get::<_, String>(3)?,
                "body": r.get::<_, String>(4)?, "ref": r.get::<_, Option<String>>(5)?,
                "created_at": r.get::<_, i64>(6)?,
                "session": r.get::<_, Option<String>>(7)?,
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
