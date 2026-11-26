//! What `moai card orchestrate plan` does (SPEC-20261124-orchestrate-plan): the coordination
//! window's round as one read-only plan. It reads the ledger, the lane registry and, when a file
//! is given, the host state in the shape of `claude agents --json`, and prints what a round would
//! do: the lanes and whether each is free, the dispatches, the clears, the audit requests, the
//! landing queue, apart from it the passed heads that are no longer their card's current head
//! (`superseded`, SPEC-20261126-orchestrate-superseded-heads) and what waits for the user. It
//! never sends, starts, audits or lands anything, and it starts no host program.
//!
//! Every part is a function of rows the ledger and the registry already hold, and of one question
//! to git (is a handed-back head inside the integration tip), asked with read-only commands.
//! Nothing in the answer depends on the clock, so two plans of an unchanged state are byte for
//! byte equal.

use super::{inbox, input, landed, spec_link, work};
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

/// What a note that hands a head back starts with. The same table as `MARKERS` of
/// `ready_receipt::marker_of_line`, which lives on the integration tip and not on the base this
/// slice was cut from; `READY` needs its space so that a longer word is no marker. The landing
/// side swaps this for a call of that function (SPEC-20261124-orchestrate-plan, D-10).
const HAND_BACK_MARKERS: [&str; 3] = ["LANE-READY-INTEGRATION", "REPAIRED", "READY "];

/// A note hands a head back when its first line, blanks before it skipped, starts with a marker.
fn hands_back(body: &str) -> bool {
    let first = body.lines().next().unwrap_or("").trim_start();
    HAND_BACK_MARKERS.iter().any(|m| first.starts_with(m))
}

/// The newest hand-back note (`LANE-READY-INTEGRATION`, `REPAIRED` or `READY`) of each card that
/// is not done and names a head: that head and its base.
fn ready_heads(conn: &Connection) -> Result<BTreeMap<i64, (String, Option<String>)>> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    // The SQL only narrows the rows; `hands_back` decides.
    let mut stmt = conn
        .prepare(
            "SELECT n.card_id, substr(n.body, 1, ?1) FROM card_note n \
               JOIN card c ON c.card_id = n.card_id \
              WHERE n.origin = 'local' AND c.state <> 'done' \
                AND (n.body LIKE '%READY%' OR n.body LIKE '%REPAIRED%') \
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
        if !hands_back(&body) {
            continue;
        }
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

/// A verdict a card note carries: its first line reads `AUDIT … head <40 hex> base <40 hex> —
/// verdict: <word>` (`gate::parse_verdict`).
struct NoteVerdict {
    note_id: i64,
    author: String,
    head: String,
    pass: bool,
}

/// The AUDIT notes of every card that is not done, oldest first. Only the first line of a note
/// is a verdict: a line quoted further down, a line that is not in the verdict shape and a
/// note that does not start with `AUDIT ` judge nothing.
fn audit_notes(conn: &Connection) -> Result<BTreeMap<i64, Vec<NoteVerdict>>> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let mut stmt = conn
        .prepare(
            "SELECT n.card_id, n.note_id, n.author, substr(n.body, 1, ?1) FROM card_note n \
               JOIN card c ON c.card_id = n.card_id \
              WHERE n.origin = 'local' AND c.state <> 'done' AND n.body LIKE 'AUDIT %' \
              ORDER BY n.note_id",
        )
        .map_err(db)?;
    let rows = stmt
        .query_map([NOTE_READ_CHARS as i64], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, i64>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, String>(3)?,
            ))
        })
        .map_err(db)?;
    let mut out: BTreeMap<i64, Vec<NoteVerdict>> = BTreeMap::new();
    for row in rows {
        let (card_id, note_id, author, body) = row.map_err(db)?;
        let line = body.lines().next().unwrap_or("");
        if !line.starts_with("AUDIT ") {
            continue;
        }
        let Some((head, _base, word)) = crate::gate::parse_verdict(line) else {
            continue;
        };
        out.entry(card_id).or_default().push(NoteVerdict {
            note_id,
            author,
            head,
            // Conditions make a verdict FAIL: only the word PASS passes.
            pass: word == "PASS",
        });
    }
    Ok(out)
}

/// What judges a head (SPEC-20261124-orchestrate-plan, revision 3, "Judgement of a head").
struct Judgement {
    pass: bool,
    reviewer: String,
    /// `event` or `note`.
    source: &'static str,
    event: Option<i64>,
    note: Option<i64>,
    tier: Option<u8>,
}

/// The newest `card.verdict` event of `head` on the card; with none, the newest AUDIT note of the
/// card that names the head. An event always wins over a note.
fn judgement(
    verdicts: &BTreeMap<i64, Vec<VerdictEvent>>,
    notes: &BTreeMap<i64, Vec<NoteVerdict>>,
    card_id: i64,
    head: &str,
) -> Option<Judgement> {
    let event = verdicts
        .get(&card_id)
        .and_then(|events| events.iter().rev().find(|e| e.head == head));
    if let Some(e) = event {
        return Some(Judgement {
            pass: e.pass,
            reviewer: e.reviewer.clone(),
            source: "event",
            event: Some(e.event_id),
            note: None,
            tier: Some(e.tier),
        });
    }
    let note = notes
        .get(&card_id)
        .and_then(|list| list.iter().rev().find(|n| n.head == head))?;
    Some(Judgement {
        pass: note.pass,
        reviewer: note.author.clone(),
        source: "note",
        event: None,
        note: Some(note.note_id),
        tier: None,
    })
}

/// Whether a head is landed: inside the tip of the integration branch, read with the functions
/// `landed::find` is built from. The tip is read once and each head asked once; a branch git
/// cannot read leaves every head not landed and the reason in `unread`.
struct Landing<'a> {
    project: &'a Path,
    tip: Option<Option<String>>,
    asked: BTreeMap<String, bool>,
    unread: Option<String>,
}

impl<'a> Landing<'a> {
    fn new(project: &'a Path) -> Self {
        Landing {
            project,
            tip: None,
            asked: BTreeMap::new(),
            unread: None,
        }
    }

    fn tip(&mut self) -> Option<String> {
        if self.tip.is_none() {
            let read = landed::integration_branch(self.project)
                .and_then(|branch| landed::tip_of(self.project, &branch));
            self.tip = Some(match read {
                Ok(tip) => Some(tip),
                Err(reason) => {
                    self.unread = Some(reason);
                    None
                }
            });
        }
        self.tip.clone().flatten()
    }

    fn is_landed(&mut self, head: &str) -> bool {
        if let Some(known) = self.asked.get(head) {
            return *known;
        }
        let inside = match self.tip() {
            None => false,
            Some(tip) => match landed::contained(self.project, head, &tip) {
                Ok(c) => c == landed::Containment::Inside,
                Err(_) => {
                    self.unread.get_or_insert(
                        "git could not tell whether a head is inside the tip".into(),
                    );
                    false
                }
            },
        };
        self.asked.insert(head.to_string(), inside);
        inside
    }
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
                "superseded": [], "attention": [], "host": {"status": "not_read"},
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
    let notes = audit_notes(conn)?;
    let mut landing = Landing::new(project);
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
    // A card whose newest hand-back head is landed, waits for an audit (no judgement) or waits
    // for its landing (newest judgement `pass`) is not work for a new lane; a `fail` sends it
    // back to repair. A lane that hands a card back gives its reservation back, so such a card
    // is `ready` or `active`, whatever `card next` ranks.
    let mut audits: Vec<Value> = Vec::new();
    let mut waiting: BTreeSet<i64> = BTreeSet::new();
    for (card_id, (head, base)) in &heads {
        if landing.is_landed(head) {
            waiting.insert(*card_id);
            continue;
        }
        match judgement(&verdicts, &notes, *card_id, head) {
            Some(verdict) => {
                if verdict.pass {
                    waiting.insert(*card_id);
                }
            }
            None => {
                waiting.insert(*card_id);
                let authors: BTreeSet<&str> = reservations
                    .iter()
                    .filter(|r| r.card_id == *card_id)
                    .map(|r| r.holder.as_str())
                    .collect();
                let name = input::card_name(*card_id);
                audits.push(json!({
                    "kind": "audit", "card": name, "head": head, "base": base,
                    "author_sessions": authors, "runner_rule": RUNNER_RULE,
                    "key": format!("audit:{name}:{head}"),
                }));
            }
        }
    }
    candidates.retain(|card_id| !waiting.contains(card_id));
    for (lane, card_id) in free.iter().zip(candidates.iter()) {
        let card = input::card_name(*card_id);
        actions.push(json!({
            "kind": "dispatch", "lane": lane.name,
            "key": format!("dispatch:{}:{}", lane.name, card), "card": card,
        }));
    }
    actions.extend(audits);

    // ---- blocked cards ----------------------------------------------------------------------
    for (card_id, reason) in &blocked {
        let name = input::card_name(*card_id);
        actions.push(json!({
            "kind": "blocked", "card": name, "reason": reason,
            "key": format!("blocked:{name}:{reason}"),
        }));
    }

    // ---- landing queue and superseded heads --------------------------------------------------
    let mut queue: Vec<(String, i64, i64, Value)> = Vec::new();
    let judged_cards: BTreeSet<i64> = verdicts.keys().chain(notes.keys()).copied().collect();
    for card_id in &judged_cards {
        let Some(card) = cards.get(card_id).filter(|c| c.state != "done") else {
            continue;
        };
        let authors: Vec<String> = reservations
            .iter()
            .filter(|r| r.card_id == *card_id)
            .map(|r| r.holder.clone())
            .collect();
        let judged_heads: BTreeSet<&str> = verdicts
            .get(card_id)
            .into_iter()
            .flatten()
            .map(|e| e.head.as_str())
            .chain(
                notes
                    .get(card_id)
                    .into_iter()
                    .flatten()
                    .map(|n| n.head.as_str()),
            )
            .collect();
        let newest_head = verdicts
            .get(card_id)
            .and_then(|events| events.last())
            .map(|e| e.head.as_str())
            .or_else(|| {
                notes
                    .get(card_id)
                    .and_then(|list| list.last())
                    .map(|n| n.head.as_str())
            });
        let ready_head = heads.get(card_id).map(|(h, _)| h.as_str());
        for head in judged_heads {
            let Some(j) = judgement(&verdicts, &notes, *card_id, head) else {
                continue;
            };
            if !j.pass || held_by(&authors, &j.reviewer) || landing.is_landed(head) {
                continue;
            }
            let name = input::card_name(*card_id);
            let current = ready_head.map_or(newest_head == Some(head), |h| h == head);
            let mut entry = json!({
                "card": name, "head": head, "reviewer": j.reviewer, "source": j.source,
                "tier": j.tier, "event": j.event, "note": j.note, "current": current,
            });
            // Only the current head of a card carries a line that lands it.
            if current {
                entry["command"] =
                    json!(format!("moai card land {name} --head {head} --by leader"));
            }
            let list = if current { "land" } else { "superseded" };
            entry["key"] = json!(format!("{list}:{name}:{head}"));
            queue.push((
                card.priority.clone(),
                *card_id,
                j.event.or(j.note).unwrap_or_default(),
                entry,
            ));
        }
    }
    queue.sort_by(|a, b| (&a.0, a.1, a.2).cmp(&(&b.0, b.1, b.2)));
    let (land_queue, superseded): (Vec<Value>, Vec<Value>) = queue
        .into_iter()
        .map(|q| q.3)
        .partition(|entry| entry["current"] == json!(true));
    if let Some(reason) = &landing.unread {
        attention.push(json!({
            "kind": "integration", "status": "unread", "reason": reason, "answer": null,
            "answered_at": "the person who owns the repository",
            "key": "attention:integration:unread",
        }));
    }

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
        "land_queue": land_queue, "superseded": superseded, "attention": attention,
        "host": host_json,
    }))
}
