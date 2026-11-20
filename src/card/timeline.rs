//! What `moai card timeline` does (SPEC-20261120-card-timeline, the first layer of the
//! collab-board design): one card's events, read without writing and without making a writer wait.
//!
//! The one way it reads: open the ledger read-only with a short wait ([`WAIT`]), run the three
//! statements below inside one read transaction, close both, and only then build the answer. An
//! open read transaction holds a writer's commit for as long as it lives (measured: eight seconds
//! of it made a write fail after five), so nothing is rendered, sorted or printed while it is
//! open, and no statement runs once per row: `event` has no `card_id` index, so each of the
//! three reads that table once and the notes are joined by their primary key.
//!
//! The answer is [`render`], a function of plain rows, so its edges are tested without a ledger.

use super::command::ABSENT_HINT;
use super::input::{self, CardRef};
use crate::ledger::{self, LedgerError};
use serde_json::{json, Map, Value};
use std::collections::HashMap;
use std::path::Path;
use std::time::Duration;

type Result<T> = std::result::Result<T, String>;

/// How long the read connection waits for a writer before it answers `ledger busy`.
pub const WAIT: Duration = Duration::from_millis(250);
/// Events per call when `--limit` is not given, and the most `--limit` may ask for.
pub const DEFAULT_LIMIT: i64 = 50;
pub const MAX_LIMIT: i64 = 500;
/// The kinds a default call leaves out (`--all` shows them). A rejected call (`fence_reject`)
/// is never in this list. The same words are in [`EVENTS_SQL`].
pub const BOOKKEEPING: [&str; 1] = ["renew"];
/// The most note events read to learn which lane a session writes as.
const LANE_ROWS: i64 = 2000;
/// A cursor is `tl1:<card number>:<event number>`; callers hand it back and never read it.
const CURSOR_TAG: &str = "tl1";
/// A payload value or a note line is cut to this many characters, and a payload shows at most
/// this many fields.
const TEXT_MAX_CHARS: usize = 200;
const DETAIL_MAX_FIELDS: usize = 24;

/// The events of one card after a cursor: `?1` card, `?2` cursor, `?3` 1 to show bookkeeping
/// events, `?4` row limit. A note event brings its note by primary key (first line only: a note
/// can be 1 MiB).
pub const EVENTS_SQL: &str = "SELECT e.event_id, e.at, e.actor, e.kind, e.payload, \
        n.author, n.origin, \
        substr(CASE WHEN instr(n.body, char(10)) > 0 \
                    THEN substr(n.body, 1, instr(n.body, char(10)) - 1) \
                    ELSE n.body END, 1, 2000) \
   FROM event e \
   LEFT JOIN card_note n ON e.kind = 'card.note' AND n.card_id = e.card_id \
                        AND n.note_id = json_extract(e.payload, '$.note_id') \
  WHERE e.card_id = ?1 AND e.event_id > ?2 AND (?3 = 1 OR e.kind <> 'renew') \
  ORDER BY e.event_id LIMIT ?4";

/// The newest notes' sessions and authors of one card: `?1` card, `?2` row limit. A session's
/// lane is the author of the newest note it wrote here.
pub const LANES_SQL: &str = "SELECT json_extract(e.payload, '$.session'), \
        json_extract(e.payload, '$.author') \
   FROM event e \
  WHERE e.card_id = ?1 AND e.kind = 'card.note' \
    AND json_extract(e.payload, '$.session') IS NOT NULL \
  ORDER BY e.event_id DESC LIMIT ?2";

/// The newest event number of one card: `?1` card.
pub const END_SQL: &str = "SELECT max(e.event_id) FROM event e WHERE e.card_id = ?1";

/// The cursor that comes after event `event_id` of card `card_id`.
pub fn cursor(card_id: i64, event_id: i64) -> String {
    format!("{CURSOR_TAG}:{card_id}:{event_id}")
}

/// The event number a cursor of card `card_id` stands for. Anything else (another card's cursor,
/// a different tag, a number that is not plain digits) is refused.
pub fn parse_cursor(card_id: i64, text: &str) -> Result<i64> {
    let refused = || {
        "card timeline refused: --after is not a cursor of this card (hand back the `next` or \
         `end` of a timeline answer for the same card)"
            .to_string()
    };
    let mut parts = text.split(':');
    let (Some(tag), Some(card), Some(event), None) =
        (parts.next(), parts.next(), parts.next(), parts.next())
    else {
        return Err(refused());
    };
    let digits = |s: &str| !s.is_empty() && s.len() <= 18 && s.bytes().all(|b| b.is_ascii_digit());
    if tag != CURSOR_TAG || !digits(card) || !digits(event) {
        return Err(refused());
    }
    if card.parse::<i64>().ok() != Some(card_id) {
        return Err(refused());
    }
    event.parse().map_err(|_| refused())
}

/// `--limit`: 1 to [`MAX_LIMIT`], [`DEFAULT_LIMIT`] when absent.
pub fn limit_of(given: Option<i64>) -> Result<i64> {
    let n = given.unwrap_or(DEFAULT_LIMIT);
    if (1..=MAX_LIMIT).contains(&n) {
        Ok(n)
    } else {
        Err(format!(
            "card timeline refused: --limit must be 1 to {MAX_LIMIT}"
        ))
    }
}

fn cut(text: &str) -> String {
    text.trim_end().chars().take(TEXT_MAX_CHARS).collect()
}

/// The fields of an event payload a reader sees: numbers, booleans, nulls and strings (cut to
/// [`TEXT_MAX_CHARS`] characters) as they are, and each list or object as `<name>_count`, at most
/// [`DETAIL_MAX_FIELDS`] fields. A payload that is not an object shows nothing.
pub fn detail(payload: &str) -> Value {
    let mut out = Map::new();
    if let Ok(Value::Object(map)) = serde_json::from_str::<Value>(payload) {
        for (name, value) in map.iter().take(DETAIL_MAX_FIELDS) {
            match value {
                Value::String(s) => {
                    out.insert(name.clone(), json!(cut(s)));
                }
                Value::Array(items) => {
                    out.insert(format!("{name}_count"), json!(items.len()));
                }
                Value::Object(fields) => {
                    out.insert(format!("{name}_count"), json!(fields.len()));
                }
                scalar => {
                    out.insert(name.clone(), scalar.clone());
                }
            }
        }
    }
    Value::Object(out)
}

/// One event row as the query reads it.
#[derive(Clone, Debug, PartialEq)]
pub struct EventRow {
    pub event_id: i64,
    pub at: i64,
    pub actor: String,
    pub kind: String,
    pub payload: String,
    /// `(author, origin, first line)` of the note a `card.note` event stands for.
    pub note: Option<(String, String, String)>,
}

/// The session an event belongs to: the one its payload names, else its actor when the actor is
/// a session id (the reservation events are written as the holder). Otherwise none.
fn session_of(row: &EventRow, payload: &Value) -> Option<String> {
    if let Some(s) = payload["session"].as_str() {
        return Some(s.to_string());
    }
    crate::session_current::well_formed(&row.actor).then(|| row.actor.clone())
}

/// The events as JSON, oldest first. `lanes` maps a session to the lane it writes as.
pub fn render(rows: &[EventRow], lanes: &HashMap<String, String>) -> Vec<Value> {
    rows.iter()
        .map(|row| {
            let payload: Value = serde_json::from_str(&row.payload).unwrap_or(Value::Null);
            let session = session_of(row, &payload);
            let lane = session
                .as_deref()
                .and_then(|s| lanes.get(s).cloned())
                .or_else(|| row.note.as_ref().map(|(author, _, _)| author.clone()));
            let mut event = json!({
                "at": row.at, "kind": row.kind, "actor": row.actor, "session": session,
                "lane": lane, "detail": detail(&row.payload),
            });
            if let Some((author, origin, line)) = &row.note {
                event["note"] = json!({"author": author, "origin": origin, "line": cut(line)});
            }
            event
        })
        .collect()
}

/// Everything read inside the transaction, as plain values.
struct Page {
    now: i64,
    title: String,
    state: String,
    rows: Vec<EventRow>,
    more: bool,
    end: Option<i64>,
    lanes: HashMap<String, String>,
    owner: Option<(String, i64, String, i64)>,
}

/// `card timeline C-n [--after CURSOR] [--limit N] [--all]`. Reads only; the ledger is open for
/// the length of [`read_page`] and not while the answer is built.
pub fn timeline(
    project: &Path,
    home: &Path,
    session: Option<&str>,
    reference: &str,
    after: Option<&str>,
    limit: Option<i64>,
    all: bool,
) -> Result<Value> {
    let card_id =
        match input::parse_ref(reference).map_err(|e| format!("card timeline refused: {e}"))? {
            CardRef::Id(n) => n,
            CardRef::SourceNumber(_) => {
                return Err("card timeline refused: name the card by its C-number".to_string())
            }
        };
    let limit = limit_of(limit)?;
    let after_id = match after {
        Some(text) => parse_cursor(card_id, text)?,
        None => 0,
    };
    let id = ledger::command::project_id_of(project)?;
    let page = {
        let ledger = ledger::open::open_read_only_within(home, &id, WAIT).map_err(|e| match e {
            LedgerError::Absent => ABSENT_HINT.to_string(),
            other => other.to_string(),
        })?;
        read_page(&ledger, card_id, after_id, limit, all)?
    };
    let events = render(&page.rows, &page.lanes);
    let next = match page.rows.last() {
        Some(last) => cursor(card_id, last.event_id),
        None => cursor(card_id, after_id),
    };
    let owner = page.owner.map(|(holder, generation, status, lease_until)| {
        json!({
            "session": holder, "lane": page.lanes.get(&holder), "generation": generation,
            "status": status, "lease_until": lease_until,
        })
    });
    Ok(json!({
        "action": "timeline", "ledger": "ready",
        "header": {"project_id": id, "session": session, "read_at": page.now},
        "card": input::card_name(card_id), "card_id": card_id,
        "title": page.title, "state": page.state, "owner": owner,
        "all": all, "limit": limit, "after": after,
        "count": events.len(), "events": events, "more": page.more,
        "next": next, "end": page.end.map(|e| cursor(card_id, e)),
        "writes": false,
    }))
}

/// The three statements, one read transaction, a bounded number of rows.
fn read_page(
    ledger: &ledger::ReadLedger,
    card_id: i64,
    after_id: i64,
    limit: i64,
    all: bool,
) -> Result<Page> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let tx = ledger.conn().unchecked_transaction().map_err(db)?;
    let now: i64 = tx
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .map_err(db)?;
    let card = tx
        .query_row(
            "SELECT title, state FROM card WHERE card_id = ?1",
            [card_id],
            |r| Ok((r.get::<_, String>(0)?, r.get::<_, String>(1)?)),
        )
        .map(Some)
        .or_else(|e| match e {
            rusqlite::Error::QueryReturnedNoRows => Ok(None),
            other => Err(other),
        })
        .map_err(db)?;
    let Some((title, state)) = card else {
        return Err(format!(
            "card not found: there is no {}",
            input::card_name(card_id)
        ));
    };

    // One row more than asked for says whether there is a next page.
    let mut stmt = tx.prepare(EVENTS_SQL).map_err(db)?;
    let mut rows: Vec<EventRow> = stmt
        .query_map(
            rusqlite::params![card_id, after_id, i64::from(all), limit + 1],
            |r| {
                let author: Option<String> = r.get(5)?;
                let origin: Option<String> = r.get(6)?;
                let line: Option<String> = r.get(7)?;
                Ok(EventRow {
                    event_id: r.get(0)?,
                    at: r.get(1)?,
                    actor: r.get(2)?,
                    kind: r.get(3)?,
                    payload: r.get(4)?,
                    note: match (author, origin) {
                        (Some(a), Some(o)) => Some((a, o, line.unwrap_or_default())),
                        _ => None,
                    },
                })
            },
        )
        .map_err(db)?
        .collect::<std::result::Result<_, _>>()
        .map_err(db)?;
    drop(stmt);
    let more = rows.len() as i64 > limit;
    rows.truncate(limit as usize);

    // A session's lane: the author of the newest note it wrote on this card.
    let mut lanes: HashMap<String, String> = HashMap::new();
    let mut stmt = tx.prepare(LANES_SQL).map_err(db)?;
    let pairs = stmt
        .query_map(rusqlite::params![card_id, LANE_ROWS], |r| {
            Ok((r.get::<_, String>(0)?, r.get::<_, Option<String>>(1)?))
        })
        .map_err(db)?;
    for pair in pairs {
        let (session, author) = pair.map_err(db)?;
        if let Some(author) = author {
            lanes.entry(session).or_insert(author);
        }
    }
    drop(stmt);

    let end: Option<i64> = tx.query_row(END_SQL, [card_id], |r| r.get(0)).map_err(db)?;
    let owner = ledger::reservation::current(&tx, card_id)
        .map_err(|e| e.to_string())?
        .map(|r| (r.holder, r.generation, r.status, r.lease_until));
    tx.finish().map_err(db)?;
    Ok(Page {
        now,
        title,
        state,
        rows,
        more,
        end,
        lanes,
        owner,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    fn row(event_id: i64, kind: &str, actor: &str, payload: &str) -> EventRow {
        EventRow {
            event_id,
            at: 100 + event_id,
            actor: actor.to_string(),
            kind: kind.to_string(),
            payload: payload.to_string(),
            note: None,
        }
    }

    #[test]
    fn cursor_round_trips_and_refuses_what_is_not_this_cards() {
        assert_eq!(cursor(7, 42), "tl1:7:42");
        assert_eq!(parse_cursor(7, "tl1:7:42"), Ok(42));
        assert_eq!(parse_cursor(7, "tl1:7:0"), Ok(0));
        for bad in [
            "",
            "tl1",
            "tl1:7",
            "tl1:7:",
            "tl1:7:-1",
            "tl1:7:+1",
            "tl1:7:4 2",
            "tl1:7:42:9",
            "tl2:7:42",
            "tl1:8:42",
            "tl1:07x:42",
            "tl1:7:9999999999999999999",
        ] {
            let e = parse_cursor(7, bad).unwrap_err();
            assert!(e.starts_with("card timeline refused:"), "{bad}: {e}");
        }
    }

    #[test]
    fn the_limit_is_one_to_five_hundred() {
        assert_eq!(limit_of(None), Ok(DEFAULT_LIMIT));
        assert_eq!(limit_of(Some(1)), Ok(1));
        assert_eq!(limit_of(Some(MAX_LIMIT)), Ok(MAX_LIMIT));
        for bad in [0, -1, MAX_LIMIT + 1] {
            assert!(limit_of(Some(bad)).is_err(), "{bad}");
        }
    }

    #[test]
    fn detail_keeps_scalars_cuts_long_strings_and_counts_lists() {
        let long = "x".repeat(300);
        let payload = json!({
            "generation": 2, "ok": true, "gone": null, "text": long,
            "write_scope": ["a", "b", "c"], "nested": {"k": 1},
        })
        .to_string();
        let d = detail(&payload);
        assert_eq!(d["generation"], json!(2));
        assert_eq!(d["ok"], json!(true));
        assert_eq!(d["gone"], Value::Null);
        assert_eq!(d["text"].as_str().unwrap().chars().count(), TEXT_MAX_CHARS);
        assert_eq!(d["write_scope_count"], json!(3));
        assert_eq!(d["nested_count"], json!(1));
        assert!(d.get("write_scope").is_none() && d.get("nested").is_none());
        assert_eq!(detail("not json"), json!({}));
        assert_eq!(detail("[1,2]"), json!({}));
        let many: Map<String, Value> = (0..40).map(|i| (format!("f{i:02}"), json!(i))).collect();
        assert_eq!(
            detail(&Value::Object(many).to_string())
                .as_object()
                .unwrap()
                .len(),
            DETAIL_MAX_FIELDS
        );
    }

    #[test]
    fn a_session_is_the_payloads_or_the_actors_and_its_lane_is_the_newest_note_author() {
        let a = "aaaaaaaa-0000-4000-8000-00000000000a";
        let mut lanes = HashMap::new();
        lanes.insert(a.to_string(), "agent-a".to_string());
        let rows = [
            row(1, "reserve", a, "{}"),
            row(2, "card.add", "user", "{}"),
            row(
                3,
                "card.note",
                "agent-z",
                &json!({"session": a}).to_string(),
            ),
        ];
        let events = render(&rows, &lanes);
        assert_eq!(events[0]["session"], json!(a));
        assert_eq!(events[0]["lane"], json!("agent-a"));
        assert_eq!(events[1]["session"], Value::Null);
        assert_eq!(events[1]["lane"], Value::Null);
        assert_eq!(events[2]["session"], json!(a));
        assert_eq!(events[2]["lane"], json!("agent-a"));
        // A note event of a session with no entry still names its author.
        let mut with_note = row(4, "card.note", "agent-q", "{}");
        with_note.note = Some(("agent-q".into(), "local".into(), "line".into()));
        let events = render(&[with_note], &HashMap::new());
        assert_eq!(events[0]["lane"], json!("agent-q"));
        assert_eq!(events[0]["note"]["line"], json!("line"));
    }

    #[test]
    fn the_events_statement_leaves_out_exactly_the_bookkeeping_list() {
        for kind in BOOKKEEPING {
            assert!(
                EVENTS_SQL.contains(&format!("e.kind <> '{kind}'")),
                "{kind}"
            );
        }
        assert!(!EVENTS_SQL.contains("fence_reject"));
    }
}
