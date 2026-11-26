//! Card roles (SPEC-20261122-card-roles): who wrote a card, who audits it and who landed it, read
//! from ledger rows, and the typed row that records an audit assignment.
//!
//! The rule the instructions already state, now written in the ledger: the agent that implemented
//! a change never audits it. An **author** is any session that holds or ever held a reservation of
//! the card, in any generation, released or not. The **auditor** is the identity the leader
//! assigned with `card assign` (a `card.audit_assigned` event); the **lander** is the caller and
//! head of the newest `card.landed` event. Nothing here changes the schema: both new kinds are
//! free-text event kinds with a JSON payload, the writing session as the actor.
//!
//! The verdict row itself is written by `card verdict` (`super::verdict`, SPEC-20261123-landing-
//! reads-verdict); this module only reads the roles that command and `card assign` rely on.

use super::command::open_for_write;
use super::input::{self, CardRef};
use crate::ledger::workflow::{AUDIT_ASSIGNED_KIND, LEADER_BY};
use crate::ledger::{self, LedgerError};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

const ASSIGNED: &str = AUDIT_ASSIGNED_KIND;

fn sql_err(e: rusqlite::Error) -> String {
    LedgerError::from(e).to_string()
}

/// The card id a reference names; a source-number reference is not accepted here.
fn card_id_of(action: &str, reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(|e| format!("card {action} refused: {e}"))? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(format!(
            "card {action} refused: name the card by its C-number, not by a source number"
        )),
    }
}

/// The holders of every reservation generation of the card, oldest generation first, once each.
fn authors(conn: &Connection, card_id: i64) -> Result<Vec<String>> {
    let mut stmt = conn
        .prepare("SELECT holder FROM reservation WHERE card_id = ?1 ORDER BY generation")
        .map_err(sql_err)?;
    let rows = stmt
        .query_map([card_id], |r| r.get::<_, String>(0))
        .map_err(sql_err)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(sql_err)?;
    let mut out: Vec<String> = Vec::new();
    for holder in rows {
        if !out.contains(&holder) {
            out.push(holder);
        }
    }
    Ok(out)
}

/// The newest event of `kind` on the card: its payload and actor.
fn newest(conn: &Connection, card_id: i64, kind: &str) -> Result<Option<(Value, String)>> {
    let row: Option<(String, String)> = conn
        .query_row(
            "SELECT payload, actor FROM event WHERE card_id = ?1 AND kind = ?2 \
             ORDER BY event_id DESC LIMIT 1",
            params![card_id, kind],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .optional()
        .map_err(sql_err)?;
    Ok(row.map(|(payload, actor)| (serde_json::from_str(&payload).unwrap_or(Value::Null), actor)))
}

/// The assigned auditor of the card, if any.
fn auditor_of(conn: &Connection, card_id: i64) -> Result<Option<String>> {
    Ok(newest(conn, card_id, ASSIGNED)?
        .and_then(|(p, _)| p["auditor"].as_str().map(str::to_string)))
}

/// The roles of a card as the `roles` object of `card show`: `authors` (a list, empty when the
/// card was never held), `auditor` (a string or null) and `lander` (`{by, head}` or null).
pub fn roles_json(conn: &Connection, card_id: i64) -> Result<Value> {
    let lander = newest(conn, card_id, "card.landed")?
        .map(|(p, _)| json!({"by": p["by"], "head": p["head"]}));
    Ok(json!({
        "authors": authors(conn, card_id)?,
        "auditor": auditor_of(conn, card_id)?,
        "lander": lander,
    }))
}

/// The `roles` object of the card a `card show` answer describes (its `card_id`), read from the
/// ledger again; `show` calls this after it built its answer.
pub fn roles_of_shown(project: &Path, home: &Path, shown: &Value) -> Result<Value> {
    let card_id = shown["card_id"]
        .as_i64()
        .ok_or("card show: the answer has no card id")?;
    let id = ledger::command::project_id_of(project)?;
    let reader = ledger::open_read_only(home, &id).map_err(|e| match e {
        LedgerError::Absent => super::command::ABSENT_HINT.to_string(),
        other => other.to_string(),
    })?;
    roles_json(reader.conn(), card_id)
}

/// The state of the card, or the refusal for a card that is not there.
fn state_of(conn: &Connection, action: &str, card_id: i64) -> Result<String> {
    conn.query_row(
        "SELECT state FROM card WHERE card_id = ?1",
        [card_id],
        |r| r.get::<_, String>(0),
    )
    .optional()
    .map_err(sql_err)?
    .ok_or_else(|| {
        format!(
            "card not found: there is no {} (card {action})",
            input::card_name(card_id)
        )
    })
}

/// An identity word: 1 to 100 printable characters without whitespace.
fn identity_ok(word: &str) -> bool {
    (1..=100).contains(&word.chars().count())
        && word.chars().all(|c| !c.is_whitespace() && !c.is_control())
}

/// Whether the word has the shape of a session identity: a session id with an optional `:<12hex>`
/// worktree key, ASCII case ignored. A lane name never does, and the verdict writer (which compares
/// through `trailers::same_identity`) could never match it (SPEC-20261122-card-roles REQ-8).
fn session_identity_ok(word: &str) -> bool {
    crate::trailers::value_shape(&word.to_ascii_lowercase()) != crate::trailers::Shape::Bad
}

/// `card assign C-n --auditor ID --by leader`: one `card.audit_assigned` row for a card in review,
/// or for a `ready` card whose PASS verdict the auditor wrote. Only the leader word assigns, and
/// never an identity that holds or held the card.
pub fn assign(
    project: &Path,
    home: &Path,
    reference: &str,
    auditor: &str,
    by: &str,
) -> Result<Value> {
    let card_id = card_id_of("assign", reference)?;
    let name = input::card_name(card_id);
    if by != LEADER_BY {
        return Err(format!(
            "card assign refused: only the leader assigns an auditor: give --by {LEADER_BY}; \
             nothing was written"
        ));
    }
    if !identity_ok(auditor) {
        return Err(
            "card assign refused: --auditor is one word of 1 to 100 characters without \
             whitespace, and it is a session id (hex digits and dashes, 8 to 64 characters), \
             optionally followed by `:` and the 12 hex digits of a worktree key; nothing was \
             written"
                .to_string(),
        );
    }
    if !session_identity_ok(auditor) {
        return Err(format!(
            "card assign refused: --auditor {auditor} is not a session identity: give a session \
             id (hex digits and dashes, 8 to 64 characters), optionally followed by `:` and the \
             12 hex digits of a worktree key; a lane name never matches the session that writes \
             a verdict; nothing was written"
        ));
    }
    let (id, mut writer) = open_for_write(project, home)?;
    let event_id = writer
        .with_write_tx(|tx, now| {
            let state = match state_of(tx, "assign", card_id) {
                Ok(s) => s,
                Err(e) => return Ok(Err(e)),
            };
            // A `review` card takes an assignment as it always did. A `ready` card (sent back
            // after its audit) takes it only for a reviewer of a PASS verdict of the card
            // (SPEC-20261125-card-ready-pass-landing-path), so the record names who audited.
            match state.as_str() {
                "review" => {}
                "ready" => {
                    let passed = match ledger::workflow::verdict_events(tx, card_id) {
                        Ok(events) => events.iter().any(|e| {
                            e.pass && crate::trailers::same_identity(&e.reviewer, auditor)
                        }),
                        Err(e) => return Ok(Err(e.to_string())),
                    };
                    if !passed {
                        return Ok(Err(format!(
                            "card conflict in assign: {name} is `ready`, and no PASS verdict of it \
                             names {auditor} as its reviewer; nothing was written"
                        )));
                    }
                }
                _ => {
                    return Ok(Err(format!(
                        "card conflict in assign: {name} is {state}; an auditor is assigned to a \
                         card in `review`, or in `ready` with a PASS verdict of its own; nothing \
                         was written"
                    )));
                }
            }
            let held = match authors(tx, card_id) {
                Ok(h) => h,
                Err(e) => return Ok(Err(e)),
            };
            if held
                .iter()
                .any(|h| crate::trailers::same_identity(h, auditor))
            {
                return Ok(Err(format!(
                    "card conflict in assign: {auditor} held {name}, and an auditor is never one \
                     of its authors; nothing was written"
                )));
            }
            let key = ledger::new_event_key(tx, ASSIGNED, Some(card_id))?;
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
                params![
                    key,
                    now,
                    by,
                    card_id,
                    ASSIGNED,
                    json!({"auditor": auditor, "by": by}).to_string()
                ],
            )?;
            Ok(Ok(tx.last_insert_rowid()))
        })
        .map_err(|e| e.to_string())??;
    Ok(json!({
        "action": "assign", "project_id": id, "name": name, "auditor": auditor, "by": by,
        "event_id": event_id, "written": true, "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn identities_are_checked_by_shape() {
        assert!(identity_ok("lane-3"));
        assert!(!identity_ok(""));
        assert!(!identity_ok("two words"));
        assert!(!identity_ok(&"x".repeat(101)));
        assert!(identity_ok(&"x".repeat(100)));
    }

    #[test]
    fn session_identities_are_session_ids_with_an_optional_worktree_key() {
        let id = "aaaaaaaa-0000-4000-8000-00000000000a";
        assert!(session_identity_ok(id));
        assert!(session_identity_ok(&id.to_uppercase()));
        assert!(session_identity_ok(&format!("{id}:0123456789ab")));
        assert!(session_identity_ok(&format!("{id}:0123456789AB")));
        assert!(!session_identity_ok("agent-4"));
        assert!(!session_identity_ok("leader"));
        assert!(!session_identity_ok(&format!("{id}:xyz")));
        assert!(!session_identity_ok(&format!("{id}:0123456789a")));
    }
}
