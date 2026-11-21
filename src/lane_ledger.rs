//! The per-lane ledger summary (leader-round slice 1): what each registered lane's session
//! holds, read-only, so the leader's round answer comes from the card ledger instead of
//! Issue-era signals. The join is the registry entry's `session` against the ledger
//! reservation's `holder`; a session id that changed after a clear simply holds nothing.
//! Nothing here writes: the ledger opens read-only and the registry entries are only read.

use serde_json::{json, Value};
use std::{collections::BTreeMap, fs, path::Path};

use crate::ledger::{self, LedgerError};

/// The note-body markers a lane's newest note can open with; the first match names it.
const MARKERS: [&str; 6] = [
    "MERGE-READY",
    "READY",
    "SWITCH",
    "RESUME",
    "LANE-BLOCKED",
    "LANE-HOLD",
];

/// One held or released card of a lane's session, as the round reports it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CardLine {
    pub card: String,
    pub state: String,
    /// Seconds of lease left; negative on an ended lease.
    pub lease_left: i64,
    /// The marker word the newest note opens with, or the note's kind.
    pub newest_note_kind: String,
    /// The card carries a landed record and is not closed yet.
    pub landed_waiting_finish: bool,
    pub status: String,
}

/// One lane's line: its registry entry joined with its session's ledger rows.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct LaneLine {
    pub lane: String,
    pub session: String,
    pub cards: Vec<CardLine>,
    pub idle: bool,
}

/// The lanes section of the round answer, or one not_run line.
#[derive(Debug, PartialEq, Eq)]
pub enum Lanes {
    NotRun(String),
    Run(Vec<LaneLine>),
}

/// Reads one registry entry file: `name`, the `session` it registered with, its lease.
fn read_entry(path: &Path) -> Option<(String, String, i64)> {
    let text = fs::read_to_string(path).ok()?;
    let v: Value = serde_json::from_str(&text).ok()?;
    let name = v.get("name")?.as_str()?.to_string();
    let session = v.get("session")?.as_str()?.to_string();
    let lease = v.get("lease_until")?.as_i64()?;
    Some((name, session, lease))
}

/// The live registry entries: `(name, session)` pairs whose lease has not ended.
pub fn live_entries(entries_dir: &Path, now: i64) -> Vec<(String, String)> {
    let mut out = Vec::new();
    let Ok(entries) = fs::read_dir(entries_dir) else {
        return out;
    };
    for e in entries.filter_map(|e| e.ok()) {
        let path = e.path();
        if path.extension().is_some_and(|x| x == "json") {
            if let Some((name, session, lease)) = read_entry(&path) {
                if lease > now {
                    out.push((name, session));
                }
            }
        }
    }
    out.sort();
    out
}

/// The per-lane summary. `entries_dir` holds the session-registry entry files; the ledger is
/// the project's, opened read-only. Any read failure is one not_run line.
pub fn lane_summaries(entries_dir: &Path, home: &Path, project: &Path, now: i64) -> Value {
    let Ok(id) = crate::ledger::command::project_id_of(project) else {
        return json!({"lanes": {"state": "not_run", "reason": "the project id could not be read"}});
    };
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return json!({"lanes": {"state": "not_run", "reason": "no card ledger for this project"}});
        }
        Err(e) => {
            return json!({"lanes": {"state": "not_run", "reason": e.to_string()}});
        }
    };
    let conn = ledger.conn();
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();

    // Every reservation of every session the registry names, held or released.
    let mut by_session: BTreeMap<String, Vec<CardLine>> = BTreeMap::new();
    let sessions: Vec<String> = live_entries(entries_dir, now)
        .into_iter()
        .map(|(_, s)| s)
        .collect();
    for session in &sessions {
        let mut stmt = match conn
            .prepare(
                "SELECT r.card_id, c.title, c.state, r.status, r.lease_until \
                 FROM reservation r JOIN card c ON c.card_id = r.card_id \
                 WHERE r.holder = ?1 ORDER BY r.card_id",
            )
            .map_err(|e| db(e))
        {
            Ok(s) => s,
            Err(reason) => return json!({"lanes": {"state": "not_run", "reason": reason}}),
        };
        let rows = match stmt.query_map(rusqlite::params![session], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, String>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, String>(3)?,
                r.get::<_, i64>(4)?,
            ))
        }) {
            Ok(r) => r,
            Err(e) => return json!({"lanes": {"state": "not_run", "reason": db(e)}}),
        };
        for row in rows.flatten() {
            let (card_id, _title, state, status, lease_until) = row;
            let newest = newest_note(&conn, card_id, now);
            let landed = conn
                .query_row(
                    "SELECT 1 FROM event WHERE card_id = ?1 AND kind = 'card.landed' LIMIT 1",
                    rusqlite::params![card_id],
                    |_| Ok(()),
                )
                .is_ok();
            let waiting = landed && state != "done";
            by_session
                .entry(session.clone())
                .or_default()
                .push(CardLine {
                    card: format!("C-{card_id}"),
                    state,
                    lease_left: lease_until - now,
                    newest_note_kind: newest,
                    landed_waiting_finish: waiting,
                    status,
                });
        }
    }

    let mut lanes: Vec<LaneLine> = Vec::new();
    for (name, session) in live_entries(entries_dir, now) {
        let cards = by_session.remove(&session).unwrap_or_default();
        let idle = cards.iter().all(|c| c.status != "held")
            && cards
                .iter()
                .all(|c| !MARKERS.contains(&c.newest_note_kind.as_str()));
        lanes.push(LaneLine {
            lane: name,
            session,
            cards,
            idle,
        });
    }
    let lanes = json!(lanes
        .iter()
        .map(|l| json!({
            "lane": l.lane,
            "session": l.session,
            "idle": l.idle,
            "cards": l.cards.iter().map(|c| json!({
                "card": c.card,
                "state": c.state,
                "status": c.status,
                "lease_left": c.lease_left,
                "newest_note_kind": c.newest_note_kind,
                "landed_waiting_finish": c.landed_waiting_finish,
            })).collect::<Vec<_>>(),
        }))
        .collect::<Vec<_>>());
    json!({"lanes": {"state": "run", "items": lanes}})
}

/// The marker word or kind of a card's newest note, judged the same way every time.
fn newest_note(conn: &rusqlite::Connection, card_id: i64, _now: i64) -> String {
    let row: Option<(String, String)> = conn
        .query_row(
            "SELECT kind, body FROM card_note WHERE card_id = ?1 \
             ORDER BY note_id DESC LIMIT 1",
            rusqlite::params![card_id],
            |r| Ok((r.get(0)?, r.get(1)?)),
        )
        .ok();
    let Some((kind, body)) = row else {
        return "none".into();
    };
    let first = body.trim_start();
    for m in MARKERS {
        if first.starts_with(m) {
            return m.to_string();
        }
    }
    kind
}
