//! Card roles (SPEC-20261122-card-roles): who wrote a card, who audits it and who landed it, read
//! from ledger rows, and the two typed rows that record an audit assignment and a verdict.
//!
//! The rule the instructions already state, now written in the ledger: the agent that implemented
//! a change never audits it. An **author** is any session that holds or ever held a reservation of
//! the card, in any generation, released or not. The **auditor** is the identity the leader
//! assigned with `card assign` (a `card.audit_assigned` event); the **lander** is the caller and
//! head of the newest `card.landed` event. Nothing here changes the schema: both new kinds are
//! free-text event kinds with a JSON payload, the writing session as the actor.
//!
//! A verdict (`card verdict`) is the typed row of an audit: the head it judges, `pass` or `fail`
//! and an optional evidence reference. It is refused for a card that is not in review, from an
//! author, and, when an auditor was assigned, from any other session. It does not move the card
//! and no landing reads it yet: that gate is the next slice.

use super::command::open_for_write;
use super::input::{self, CardRef};
use crate::ledger::workflow::LEADER_BY;
use crate::ledger::{self, LedgerError};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

const VERDICT: &str = "card.verdict";
const ASSIGNED: &str = "card.audit_assigned";

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

/// A full commit id: 40 lower-case hex digits.
fn full_head_ok(head: &str) -> bool {
    head.len() == 40 && head.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f'))
}

/// What `card verdict` was given.
pub struct VerdictRequest<'a> {
    /// `--head`: the commit the verdict judges, a full 40-digit id.
    pub head: &'a str,
    /// `--result`: `pass` or `fail`.
    pub result: &'a str,
    /// `--evidence`: a reference to what the audit read (a note id, a report path), optional.
    pub evidence: Option<&'a str>,
    /// This session's id, read from the host's environment by the caller.
    pub session: &'a str,
}

/// `card verdict C-n --head SHA --result pass|fail [--evidence REF]`: one `card.verdict` row for a
/// card in review, written by this session. See the module comment for what refuses.
pub fn verdict(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &VerdictRequest<'_>,
) -> Result<Value> {
    let card_id = card_id_of("verdict", reference)?;
    let name = input::card_name(card_id);
    if !full_head_ok(req.head) {
        return Err(
            "card verdict refused: --head is a full 40-digit commit id (lower-case hex); \
             nothing was written"
                .to_string(),
        );
    }
    if req.result != "pass" && req.result != "fail" {
        return Err(
            "card verdict refused: --result is pass or fail; nothing was written".to_string(),
        );
    }
    if let Some(e) = req.evidence {
        if e.is_empty() || e.chars().count() > 200 || e.chars().any(char::is_control) {
            return Err(
                "card verdict refused: --evidence is one line of at most 200 characters; \
                 nothing was written"
                    .to_string(),
            );
        }
    }
    let (id, mut writer) = open_for_write(project, home)?;
    let outcome = writer
        .with_write_tx(|tx, now| {
            let state = match state_of(tx, "verdict", card_id) {
                Ok(s) => s,
                Err(e) => return Ok(Err(e)),
            };
            if state != "review" {
                return Ok(Err(format!(
                    "card conflict in verdict: {name} is {state}, and a verdict is written for a \
                     card in review; nothing was written"
                )));
            }
            let held = match authors(tx, card_id) {
                Ok(h) => h,
                Err(e) => return Ok(Err(e)),
            };
            if held.iter().any(|h| h == req.session) {
                return Ok(Err(format!(
                    "card conflict in verdict: this session held {name}, and the agent that \
                     implemented a change never writes its verdict; nothing was written"
                )));
            }
            match auditor_of(tx, card_id) {
                Ok(Some(assigned)) if assigned != req.session => {
                    return Ok(Err(format!(
                        "card conflict in verdict: the auditor assigned to {name} is {assigned}, \
                         not this session; nothing was written"
                    )));
                }
                Ok(_) => {}
                Err(e) => return Ok(Err(e)),
            }
            let repeated: Option<i64> = tx
                .query_row(
                    "SELECT event_id FROM event WHERE card_id = ?1 AND kind = ?2 AND actor = ?3 \
                       AND json_extract(payload, '$.head') = ?4 \
                       AND json_extract(payload, '$.result') = ?5 \
                     ORDER BY event_id DESC LIMIT 1",
                    params![card_id, VERDICT, req.session, req.head, req.result],
                    |r| r.get(0),
                )
                .optional()?;
            if let Some(event_id) = repeated {
                return Ok(Ok((false, event_id)));
            }
            let key = ledger::new_event_key(tx, VERDICT, Some(card_id))?;
            tx.execute(
                "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
                params![
                    key,
                    now,
                    req.session,
                    card_id,
                    VERDICT,
                    json!({"head": req.head, "result": req.result, "evidence": req.evidence})
                        .to_string()
                ],
            )?;
            let event_id = tx.last_insert_rowid();
            Ok(Ok((true, event_id)))
        })
        .map_err(|e| e.to_string())??;
    let (written, event_id) = outcome;
    Ok(json!({
        "action": "verdict", "project_id": id, "name": name, "head": req.head,
        "result": req.result, "evidence": req.evidence, "actor": req.session,
        "event_id": event_id, "written": written, "writes": written,
    }))
}

/// `card assign C-n --auditor ID --by leader`: one `card.audit_assigned` row for a card in review.
/// Only the leader word assigns, and never an identity that holds or held the card.
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
            "card assign refused: --auditor is one word of 1 to 100 characters; nothing was \
             written"
                .to_string(),
        );
    }
    let (id, mut writer) = open_for_write(project, home)?;
    let event_id = writer
        .with_write_tx(|tx, now| {
            let state = match state_of(tx, "assign", card_id) {
                Ok(s) => s,
                Err(e) => return Ok(Err(e)),
            };
            if state != "review" {
                return Ok(Err(format!(
                    "card conflict in assign: {name} is {state}, and an auditor is assigned to a \
                     card in review; nothing was written"
                )));
            }
            let held = match authors(tx, card_id) {
                Ok(h) => h,
                Err(e) => return Ok(Err(e)),
            };
            if held.iter().any(|h| h == auditor) {
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
    fn heads_and_identities_are_checked_by_shape() {
        assert!(full_head_ok(&"a".repeat(40)));
        assert!(!full_head_ok(&"a".repeat(39)));
        assert!(!full_head_ok(&"A".repeat(40)));
        assert!(!full_head_ok(&"g".repeat(40)));
        assert!(identity_ok("lane-3"));
        assert!(!identity_ok(""));
        assert!(!identity_ok("two words"));
        assert!(!identity_ok(&"x".repeat(101)));
        assert!(identity_ok(&"x".repeat(100)));
    }
}
