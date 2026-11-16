//! Reading the ledger's five exported tables into the rows of an export (Issue #3070 K2a).
//!
//! `read_snapshot` takes an open connection and returns a [`Snapshot`]: `card`, `card_ac`,
//! `card_note`, `card_link` and `event`, each in id order, read inside one read transaction so
//! the five tables are one moment of the ledger even if another process writes meanwhile. It only
//! reads. Turning the rows into bytes is `export_format::render`, putting the bytes in a folder is
//! `export_dir::write`.

use super::export_format::{Ac, Card, Event, Link, Note, Snapshot};
use crate::ledger::LedgerError;
use rusqlite::{Connection, Params, Row, Transaction};

type Result<T> = std::result::Result<T, LedgerError>;

/// All rows of `sql`, mapped one by one.
fn rows<T, P: Params>(
    tx: &Transaction<'_>,
    sql: &str,
    params: P,
    map: impl Fn(&Row<'_>) -> rusqlite::Result<T>,
) -> Result<Vec<T>> {
    let mut stmt = tx.prepare(sql)?;
    let mapped = stmt.query_map(params, |r| map(r))?;
    let mut out = Vec::new();
    for row in mapped {
        out.push(row?);
    }
    Ok(out)
}

/// The five exported tables of the ledger behind `conn`, with the project id and schema version
/// the caller already checked when it opened the ledger.
pub fn read_snapshot(conn: &Connection, project_id: &str, schema_version: i64) -> Result<Snapshot> {
    // A deferred transaction on a read-only connection: it reads, then rolls back on drop.
    let tx = conn.unchecked_transaction()?;
    let cards = rows(
        &tx,
        "SELECT card_id, title, body, kind, state, priority, write_scope, generation, rev, \
                created_by, created_at, updated_at, closed_at, source_host, source_project, \
                source_number, source_url \
           FROM card ORDER BY card_id",
        [],
        |r| {
            Ok(Card {
                card_id: r.get(0)?,
                title: r.get(1)?,
                body: r.get(2)?,
                kind: r.get(3)?,
                state: r.get(4)?,
                priority: r.get(5)?,
                write_scope: r.get(6)?,
                generation: r.get(7)?,
                rev: r.get(8)?,
                created_by: r.get(9)?,
                created_at: r.get(10)?,
                updated_at: r.get(11)?,
                closed_at: r.get(12)?,
                source_host: r.get(13)?,
                source_project: r.get(14)?,
                source_number: r.get(15)?,
                source_url: r.get(16)?,
            })
        },
    )?;
    let acs = rows(
        &tx,
        "SELECT ac_id, card_id, seq, text, verify, status, evidence_ref, withdrawn_reason, \
                updated_at \
           FROM card_ac ORDER BY ac_id",
        [],
        |r| {
            Ok(Ac {
                ac_id: r.get(0)?,
                card_id: r.get(1)?,
                seq: r.get(2)?,
                text: r.get(3)?,
                verify: r.get(4)?,
                status: r.get(5)?,
                evidence_ref: r.get(6)?,
                withdrawn_reason: r.get(7)?,
                updated_at: r.get(8)?,
            })
        },
    )?;
    let notes = rows(
        &tx,
        "SELECT note_id, card_id, kind, author, origin, body, ref, created_at \
           FROM card_note ORDER BY note_id",
        [],
        |r| {
            Ok(Note {
                note_id: r.get(0)?,
                card_id: r.get(1)?,
                kind: r.get(2)?,
                author: r.get(3)?,
                origin: r.get(4)?,
                body: r.get(5)?,
                reference: r.get(6)?,
                created_at: r.get(7)?,
            })
        },
    )?;
    let links = rows(
        &tx,
        "SELECT link_id, from_card, kind, to_card, target_host, target_project, target_number, \
                pr_state, merge_ref, state, created_at, updated_at \
           FROM card_link ORDER BY link_id",
        [],
        |r| {
            Ok(Link {
                link_id: r.get(0)?,
                from_card: r.get(1)?,
                kind: r.get(2)?,
                to_card: r.get(3)?,
                target_host: r.get(4)?,
                target_project: r.get(5)?,
                target_number: r.get(6)?,
                pr_state: r.get(7)?,
                merge_ref: r.get(8)?,
                state: r.get(9)?,
                created_at: r.get(10)?,
                updated_at: r.get(11)?,
            })
        },
    )?;
    let events = rows(
        &tx,
        "SELECT event_id, event_key, at, actor, card_id, kind, payload \
           FROM event ORDER BY event_id",
        [],
        |r| {
            Ok(Event {
                event_id: r.get(0)?,
                event_key: r.get(1)?,
                at: r.get(2)?,
                actor: r.get(3)?,
                card_id: r.get(4)?,
                kind: r.get(5)?,
                payload: r.get(6)?,
            })
        },
    )?;
    tx.rollback()?;
    Ok(Snapshot {
        project_id: project_id.to_string(),
        schema_version,
        cards,
        acs,
        notes,
        links,
        events,
    })
}
