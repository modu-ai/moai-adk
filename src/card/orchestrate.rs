//! What `moai card orchestrate plan` does (SPEC-20261124-orchestrate-plan): the coordination
//! window's round as one read-only plan. It reads the ledger, the lane registry and, when a file
//! is given, the host state in the shape of `claude agents --json`, and prints what a round would
//! do: the lanes and whether each is free, the dispatches, the clears, the audit requests, the
//! landing queue and what waits for the user. It never sends, starts, audits or lands anything,
//! and it starts no host program.
//!
//! Every part is a function of rows the ledger and the registry already hold. Nothing in the
//! answer depends on the clock, so two plans of an unchanged state are byte for byte equal.

use super::{inbox, input, spec_link, work};
use crate::lane_registry::{self, SessionEntry};
use crate::ledger::workflow::{self, VerdictEvent};
use crate::ledger::{self, LedgerError};
use crate::trailers::same_identity;
use rusqlite::Connection;
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The largest host file read, in bytes.
const MAX_HOST_BYTES: u64 = 4 * 1024 * 1024;
/// How much of a note's text is read to find its head and base.
const NOTE_READ_CHARS: usize = 2000;
/// Reservation statuses that hold a card.
const HOLDING: [&str; 3] = ["held", "held_for_review", "expired_candidate"];
/// Why a `clear` has no channel on the native host.
const NO_CLEAR_CHANNEL: &str = "a message never runs /clear: it reaches the lane as plain text";
/// The rule an audit request carries about its runner.
const RUNNER_RULE: &str = "the runner session differs from every author session of the card";
/// Where a decision or a block waits for its answer.
const NOTE_ANSWERED_AT: &str =
    "a card note by the leader, or the user's selection in the session the user talks to";

/// A card as the plan reads it.
struct CardRow {
    id: i64,
    state: String,
    priority: String,
}

/// A reservation row, any status.
struct Reservation {
    id: i64,
    card_id: i64,
    generation: i64,
    holder: String,
    status: String,
}

/// One session of the host file.
struct HostSession {
    id: Option<String>,
    session: Option<String>,
    name: Option<String>,
    state: Option<String>,
    status: Option<String>,
    waiting_for: Option<String>,
}

/// What the host file gave.
enum Host {
    NotRead,
    Read(Vec<HostSession>),
    Unread(String),
}

fn text_of(v: &Value, key: &str) -> Option<String> {
    v.get(key)
        .and_then(Value::as_str)
        .filter(|s| !s.is_empty())
        .map(str::to_string)
}

/// The host file in the shape of `claude agents --json`: a JSON array of objects. Nothing of the
/// path or of the file's text is carried into the reason.
fn read_host(path: Option<&Path>) -> Host {
    let Some(path) = path else {
        return Host::NotRead;
    };
    let meta = match std::fs::metadata(path) {
        Ok(m) => m,
        Err(e) => return Host::Unread(format!("the host file cannot be read ({:?})", e.kind())),
    };
    if !meta.is_file() {
        return Host::Unread("the host file is not a regular file".into());
    }
    if meta.len() > MAX_HOST_BYTES {
        return Host::Unread(format!(
            "the host file is larger than {MAX_HOST_BYTES} bytes"
        ));
    }
    let bytes = match std::fs::read(path) {
        Ok(b) => b,
        Err(e) => return Host::Unread(format!("the host file cannot be read ({:?})", e.kind())),
    };
    let value: Value = match serde_json::from_slice(&bytes) {
        Ok(v) => v,
        Err(e) => return Host::Unread(format!("the host file is not JSON ({e})")),
    };
    let Some(list) = value.as_array() else {
        return Host::Unread("the host file is not a JSON array".into());
    };
    Host::Read(
        list.iter()
            .filter(|item| item.is_object())
            .map(|item| HostSession {
                id: text_of(item, "id"),
                session: text_of(item, "sessionId"),
                name: text_of(item, "name"),
                state: text_of(item, "state"),
                status: text_of(item, "status"),
                waiting_for: text_of(item, "waitingFor"),
            })
            .collect(),
    )
}

/// The sort key of a lane name: `agent-2` before `agent-10`.
fn lane_key(name: &str) -> (String, u64, String) {
    if let Some((prefix, number)) = name.rsplit_once('-') {
        if let Ok(n) = number.parse::<u64>() {
            return (prefix.to_string(), n, name.to_string());
        }
    }
    (name.to_string(), 0, name.to_string())
}

/// 40 lowercase hex characters following `key` in `text`.
fn hex40_after(text: &str, key: &str) -> Option<String> {
    let at = text.find(key)? + key.len();
    let candidate: String = text[at..].chars().take(40).collect();
    (candidate.len() == 40
        && candidate
            .bytes()
            .all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f')))
    .then_some(candidate)
}

fn held_by(sessions: &[String], holder: &str) -> bool {
    sessions.iter().any(|s| same_identity(s, holder))
}

fn all_cards(conn: &Connection) -> Result<BTreeMap<i64, CardRow>> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare("SELECT card_id, state, priority FROM card")
        .map_err(db)?;
    let rows = stmt
        .query_map([], |r| {
            Ok(CardRow {
                id: r.get(0)?,
                state: r.get(1)?,
                priority: r.get(2)?,
            })
        })
        .map_err(db)?;
    let mut out = BTreeMap::new();
    for row in rows {
        let row = row.map_err(db)?;
        out.insert(row.id, row);
    }
    Ok(out)
}

fn all_reservations(conn: &Connection) -> Result<Vec<Reservation>> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT reservation_id, card_id, generation, holder, status FROM reservation \
             ORDER BY reservation_id",
        )
        .map_err(db)?;
    let rows = stmt
        .query_map([], |r| {
            Ok(Reservation {
                id: r.get(0)?,
                card_id: r.get(1)?,
                generation: r.get(2)?,
                holder: r.get(3)?,
                status: r.get(4)?,
            })
        })
        .map_err(db)?;
    rows.collect::<std::result::Result<Vec<_>, _>>().map_err(db)
}

/// The newest `LANE-READY-INTEGRATION` note of each card that is not done: its head and base.
fn ready_heads(conn: &Connection) -> Result<BTreeMap<i64, (String, Option<String>)>> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT n.card_id, substr(n.body, 1, ?1) FROM card_note n \
               JOIN card c ON c.card_id = n.card_id \
              WHERE n.origin = 'local' AND c.state <> 'done' \
                AND substr(ltrim(n.body), 1, 22) = 'LANE-READY-INTEGRATION' \
              ORDER BY n.note_id",
        )
        .map_err(db)?;
    let rows = stmt
        .query_map([NOTE_READ_CHARS as i64], |r| {
            Ok((r.get::<_, i64>(0)?, r.get::<_, String>(1)?))
        })
        .map_err(db)?;
    let mut out = BTreeMap::new();
    for row in rows {
        let (card_id, body) = row.map_err(db)?;
        // The newest note wins; a note with no head names nothing and hides no older one.
        if let Some(head) = hex40_after(&body, "head=") {
            out.insert(card_id, (head, hex40_after(&body, "base=")));
        }
    }
    Ok(out)
}

/// The verdict events of every card that has any, oldest first.
fn all_verdicts(conn: &Connection) -> Result<BTreeMap<i64, Vec<VerdictEvent>>> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT DISTINCT card_id FROM event \
              WHERE kind = ?1 AND card_id IS NOT NULL ORDER BY card_id",
        )
        .map_err(db)?;
    let ids = stmt
        .query_map([workflow::VERDICT_KIND], |r| r.get::<_, i64>(0))
        .map_err(db)?
        .collect::<std::result::Result<Vec<_>, _>>()
        .map_err(db)?;
    let mut out = BTreeMap::new();
    for id in ids {
        out.insert(
            id,
            workflow::verdict_events(conn, id).map_err(|e| e.to_string())?,
        );
    }
    Ok(out)
}

/// The window part: who holds the leader entry of the registry, read without writing.
fn window(entries: &std::result::Result<Vec<SessionEntry>, String>, now: i64) -> Value {
    let Ok(entries) = entries else {
        return json!({"status": "unread", "reason": "the session registry cannot be read"});
    };
    let held = entries
        .iter()
        .find(|e| e.name == "leader" && e.mode.as_deref() == Some("leader") && e.lease_until > now);
    let Some(entry) = held else {
        return json!({"status": "none"});
    };
    let running = lane_registry::running_session();
    let mine = matches!(
        (&running, &entry.session),
        (Some(a), Some(b)) if same_identity(a, b)
    );
    let mut out = json!({
        "status": if mine { "registered" } else { "other" },
        "lease_until": crate::forge::policy::format_utc(entry.lease_until),
    });
    if let Some(s) = &entry.session {
        out["session"] = json!(s);
    }
    out
}

/// One lane as the plan reads it.
struct Lane {
    name: String,
    session: Option<String>,
    sessions: Vec<String>,
}

/// `card orchestrate plan [--agents-json FILE]`. Reads only; a project without a ledger has an
/// empty plan (exit 0) and nothing is created.
pub fn plan(project: &Path, home: &Path, agents_json: Option<&Path>) -> Result<Value> {
    let id = ledger::command::project_id_of(project)?;
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Ok(json!({
                "action": "orchestrate", "project_id": id, "ledger": "absent", "writes": false,
                "window": {"status": "none"}, "lanes": [], "actions": [], "land_queue": [],
                "attention": [], "host": {"status": "not_read"},
            }));
        }
        Err(e) => return Err(e.to_string()),
    };
    let conn = ledger.conn();
    let now: i64 = conn
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .map_err(|e| LedgerError::from(e).to_string())?;
    let cards = all_cards(conn)?;
    let reservations = all_reservations(conn)?;
    let heads = ready_heads(conn)?;
    let verdicts = all_verdicts(conn)?;
    let entries = lane_registry::session_entries(project);
    let host = read_host(agents_json);

    let mut attention: Vec<Value> = Vec::new();
    let mut actions: Vec<Value> = Vec::new();

    // ---- lanes ----------------------------------------------------------------------------
    let mut lanes: Vec<Lane> = Vec::new();
    match &entries {
        Ok(list) => {
            for e in list {
                if e.mode.as_deref() == Some("agent") && e.lease_until > now {
                    let mut sessions: Vec<String> = e.session.iter().cloned().collect();
                    sessions.extend(e.previous_sessions.iter().cloned());
                    lanes.push(Lane {
                        name: e.name.clone(),
                        session: e.session.clone(),
                        sessions,
                    });
                }
            }
        }
        Err(_) => attention.push(json!({
            "kind": "registry", "status": "unread",
            "reason": "the session registry cannot be read", "answer": null,
            "answered_at": "the person who owns the registry folder",
            "key": "attention:registry:unread",
        })),
    }
    lanes.sort_by_key(|l| lane_key(&l.name));

    let mut lane_json: Vec<Value> = Vec::new();
    let mut free: Vec<&Lane> = Vec::new();
    let mut clears: Vec<Value> = Vec::new();
    for lane in &lanes {
        let mut v = json!({"name": lane.name, "session": lane.session});
        if lane.session.is_none() {
            v["state"] = json!("unknown");
            v["reason"] = json!("the registry entry carries no host session id");
        } else {
            let mine: Vec<&Reservation> = reservations
                .iter()
                .filter(|r| held_by(&lane.sessions, &r.holder))
                .collect();
            let holding: BTreeSet<i64> = mine
                .iter()
                .filter(|r| HOLDING.contains(&r.status.as_str()))
                .map(|r| r.card_id)
                .collect();
            let last = mine.iter().max_by_key(|r| r.id);
            v["last_card"] = json!(last.map(|r| input::card_name(r.card_id)));
            if holding.is_empty() {
                v["state"] = json!("free");
                free.push(lane);
                if let Some(r) = last {
                    let done = cards.get(&r.card_id).is_some_and(|c| c.state == "done");
                    if done || r.status == "released" {
                        let card = input::card_name(r.card_id);
                        clears.push(json!({
                            "kind": "clear", "lane": lane.name, "card": card,
                            "channel": "none", "reason": NO_CLEAR_CHANNEL,
                            "key": format!("clear:{}:{}.g{}", lane.name, card, r.generation),
                        }));
                    }
                }
            } else {
                v["state"] = json!("busy");
                v["cards"] = json!(holding
                    .iter()
                    .map(|c| input::card_name(*c))
                    .collect::<Vec<_>>());
            }
        }
        if let Host::Read(sessions) = &host {
            let found = sessions.iter().find(|s| {
                s.session
                    .as_deref()
                    .is_some_and(|sid| held_by(&lane.sessions, sid))
                    || s.name.as_deref() == Some(lane.name.as_str())
            });
            if let Some(s) = found {
                v["host"] = json!({
                    "state": s.state, "status": s.status, "waitingFor": s.waiting_for,
                });
            }
        }
        lane_json.push(v);
    }
    actions.extend(clears);

    // ---- dispatches and blocked cards -------------------------------------------------------
    let next = work::next(project, home, None)?;
    let mut candidates: Vec<i64> = Vec::new();
    let mut blocked: BTreeMap<i64, String> = BTreeMap::new();
    for c in next["ready"].as_array().into_iter().flatten() {
        let Some(card_id) = c["card_id"].as_i64() else {
            continue;
        };
        let kind = c["kind"].as_str().unwrap_or("task");
        if spec_link::changes_repository(kind) && c["spec"]["approved"] != json!(true) {
            blocked.insert(card_id, "spec_not_approved".into());
        } else {
            candidates.push(card_id);
        }
    }
    for s in next["skipped"].as_array().into_iter().flatten() {
        let (Some(card_id), Some(reason)) = (s["card_id"].as_i64(), s["reason"].as_str()) else {
            continue;
        };
        // A card a registered lane holds is that lane's work, not a block; `already_landed`
        // is for `card finish`.
        let lane_holds = reason == "held"
            && s["holder"]
                .as_str()
                .is_some_and(|h| lanes.iter().any(|l| held_by(&l.sessions, h)));
        if reason != "already_landed" && !lane_holds {
            blocked.insert(card_id, reason.to_string());
        }
    }
    for (lane, card_id) in free.iter().zip(candidates.iter()) {
        let card = input::card_name(*card_id);
        actions.push(json!({
            "kind": "dispatch", "lane": lane.name,
            "key": format!("dispatch:{}:{}", lane.name, card), "card": card,
        }));
    }

    // ---- audit requests ---------------------------------------------------------------------
    for (card_id, card) in &cards {
        if card.state != "review" {
            continue;
        }
        let Some((head, base)) = heads.get(card_id) else {
            continue;
        };
        let judged = verdicts
            .get(card_id)
            .is_some_and(|events| events.iter().any(|e| &e.head == head));
        if judged {
            continue;
        }
        let authors: BTreeSet<&str> = reservations
            .iter()
            .filter(|r| r.card_id == *card_id)
            .map(|r| r.holder.as_str())
            .collect();
        let name = input::card_name(*card_id);
        actions.push(json!({
            "kind": "audit", "card": name, "head": head, "base": base,
            "author_sessions": authors, "runner_rule": RUNNER_RULE,
            "key": format!("audit:{name}:{head}"),
        }));
    }

    // ---- blocked cards ----------------------------------------------------------------------
    for (card_id, reason) in &blocked {
        let name = input::card_name(*card_id);
        actions.push(json!({
            "kind": "blocked", "card": name, "reason": reason,
            "key": format!("blocked:{name}:{reason}"),
        }));
    }

    // ---- landing queue ----------------------------------------------------------------------
    let mut queue: Vec<(String, i64, i64, Value)> = Vec::new();
    for (card_id, events) in &verdicts {
        let Some(card) = cards.get(card_id).filter(|c| c.state != "done") else {
            continue;
        };
        let authors: Vec<String> = reservations
            .iter()
            .filter(|r| r.card_id == *card_id)
            .map(|r| r.holder.clone())
            .collect();
        let mut by_head: BTreeMap<&str, Vec<&VerdictEvent>> = BTreeMap::new();
        for e in events {
            by_head.entry(e.head.as_str()).or_default().push(e);
        }
        let newest_head = events.last().map(|e| e.head.as_str());
        let ready_head = heads.get(card_id).map(|(h, _)| h.as_str());
        for (head, list) in by_head {
            let Some(newest) = list.last() else {
                continue;
            };
            if !newest.pass || held_by(&authors, &newest.reviewer) {
                continue;
            }
            let name = input::card_name(*card_id);
            let current = ready_head.map_or(newest_head == Some(head), |h| h == head);
            queue.push((
                card.priority.clone(),
                *card_id,
                newest.event_id,
                json!({
                    "card": name, "head": head, "reviewer": newest.reviewer,
                    "tier": newest.tier, "event": newest.event_id, "current": current,
                    "command": format!("moai card land {name} --head {head} --by leader"),
                    "key": format!("land:{name}:{head}"),
                }),
            ));
        }
    }
    queue.sort_by(|a, b| (&a.0, a.1, a.2).cmp(&(&b.0, b.1, b.2)));
    let land_queue: Vec<Value> = queue.into_iter().map(|q| q.3).collect();

    // ---- what waits for the user ------------------------------------------------------------
    let waiting = inbox::inbox(project, home, None)?;
    for item in waiting["items"].as_array().into_iter().flatten() {
        let family = item["family"].as_str().unwrap_or_default();
        if family != "blocked" && family != "decision_question" {
            continue;
        }
        let (card, note) = (
            item["card"].as_str().unwrap_or_default(),
            item["note_id"].as_i64().unwrap_or_default(),
        );
        attention.push(json!({
            "kind": "note", "card": card, "note_id": note, "family": family,
            "line": item["line"], "answer": null, "answered_at": NOTE_ANSWERED_AT,
            "key": format!("attention:{card}:{note}"),
        }));
    }
    let host_json = match &host {
        Host::NotRead => json!({"status": "not_read"}),
        Host::Unread(reason) => {
            attention.push(json!({
                "kind": "host", "status": "unread", "reason": reason, "answer": null,
                "answered_at": "the person who owns the host file", "key": "attention:host:unread",
            }));
            json!({"status": "unread", "reason": reason})
        }
        Host::Read(sessions) => {
            for s in sessions {
                let Some(waiting_for) = &s.waiting_for else {
                    continue;
                };
                let handle =
                    s.id.as_deref()
                        .or(s.session.as_deref())
                        .unwrap_or("unknown");
                attention.push(json!({
                    "kind": "session", "session": s.session, "name": s.name, "id": s.id,
                    "waitingFor": waiting_for, "answer": null,
                    "answered_at": format!("claude attach {handle}"),
                    "key": format!("attention:{handle}:{}", waiting_for.replace(' ', "_")),
                }));
            }
            json!({"status": "read", "sessions": sessions.len()})
        }
    };

    Ok(json!({
        "action": "orchestrate", "project_id": id, "ledger": "ready", "writes": false,
        "window": window(&entries, now), "lanes": lane_json, "actions": actions,
        "land_queue": land_queue, "attention": attention, "host": host_json,
    }))
}
