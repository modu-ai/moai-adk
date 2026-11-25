//! `moai flow-measure` (SPEC-20261125-flow-measure): how long the stages of the card flow took,
//! read from the ledger and never written.
//!
//! The definitions are those of the workflow audit report. A submit is the newest hand-back note
//! (`LANE-READY-INTEGRATION`, `REPAIRED` or `READY `, judged by [`marker_of_line`], not written by
//! the leader and not a decision) at or before the verdict; a verdict is a `card.verdict` event
//! that says `pass`, its actor being the auditor session; done is the `card.finish` event; the
//! return is the moment the reservation left `held`.
//!
//! The one way it reads: open the ledger read-only with a short wait, run the statements below
//! inside one read transaction, close it, and only then compute. An open read transaction holds a
//! writer's commit for as long as it lives, so nothing is sorted or rendered while it is open.
//! [`compute`] is a function of plain rows, so its edges are tested without a ledger.

use crate::card::input::{self, CardRef};
use crate::card::ready_receipt::marker_of_line;
use crate::ledger::{self, LedgerError};
use serde_json::{json, Map, Value};
use std::collections::{BTreeMap, HashMap};
use std::path::Path;
use std::time::Duration;

type Result<T> = std::result::Result<T, String>;

/// How long the read connection waits for a writer before it answers `ledger busy`.
pub const WAIT: Duration = Duration::from_millis(250);
/// The longest window, in hours (ten years).
pub const MAX_HOURS: i64 = 24 * 365 * 10;

/// The events the measure reads. `event` has no `card_id` index, so one read filters by kind; only
/// a verdict's payload is read.
const EVENTS_SQL: &str = "SELECT event_id, at, actor, card_id, kind, \
        CASE WHEN kind = 'card.verdict' THEN payload END \
    FROM event \
    WHERE kind IN ('card.add', 'card.landed', 'card.finish', 'card.verdict') \
    ORDER BY event_id";

/// The notes the measure reads: hand-back notes and blocks, by the start of the body (a superset:
/// the Rust side judges the first line again). Only the first line leaves the database.
const NOTES_SQL: &str = "SELECT card_id, note_id, kind, author, created_at, \
        substr(CASE WHEN instr(body, char(10)) > 0 \
                    THEN substr(body, 1, instr(body, char(10)) - 1) \
                    ELSE body END, 1, 2000) \
    FROM card_note \
    WHERE body LIKE 'LANE-READY-INTEGRATION%' OR body LIKE 'REPAIRED%' \
       OR body LIKE 'READY %' OR body LIKE 'LANE-BLOCKED%' \
    ORDER BY note_id";

const RESERVATIONS_SQL: &str = "SELECT card_id, generation, status, acquired_at, updated_at \
    FROM reservation ORDER BY card_id, generation";

/// What the command was asked for.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Request {
    pub cards: Vec<i64>,
    pub hours: Option<i64>,
    pub now: Option<i64>,
}

/// Reads the arguments: card names (`C-<number>`), an optional window and an optional end time.
pub fn parse_request(cards: &[String], hours: Option<&str>, now: Option<i64>) -> Result<Request> {
    let mut named: Vec<i64> = Vec::new();
    for text in cards {
        match input::parse_ref(text) {
            Ok(CardRef::Id(n)) => {
                if !named.contains(&n) {
                    named.push(n);
                }
            }
            _ => {
                return Err(format!(
                "flow-measure refused: the card name `{text}` is not valid; write it as C-<number>"
            ))
            }
        }
    }
    let hours = match hours {
        None => None,
        Some(text) => {
            let n = text
                .parse::<i64>()
                .ok()
                .filter(|n| (1..=MAX_HOURS).contains(n));
            match n {
                Some(n) => Some(n),
                None => {
                    return Err(format!(
                    "flow-measure refused: --hours must be a whole number from 1 to {MAX_HOURS}, \
                         not `{text}`"
                ))
                }
            }
        }
    };
    if named.is_empty() && hours.is_none() {
        return Err(
            "flow-measure refused: give a card name (C-<number>) or --hours <N>, or both"
                .to_string(),
        );
    }
    Ok(Request {
        cards: named,
        hours,
        now,
    })
}

#[derive(Debug, Clone)]
struct Verdict {
    id: i64,
    at: i64,
    actor: String,
    card: i64,
    pass: bool,
    head: Option<String>,
}

#[derive(Debug, Clone)]
struct Stamp {
    card: i64,
    at: i64,
}

#[derive(Debug, Clone)]
struct Submit {
    card: i64,
    note_id: i64,
    at: i64,
}

#[derive(Debug, Clone)]
struct Block {
    at: i64,
    kind: String,
}

#[derive(Debug, Clone)]
struct Reservation {
    card: i64,
    generation: i64,
    status: String,
    acquired: i64,
    updated: i64,
}

/// The rows the measure works from.
#[derive(Debug, Default)]
struct Data {
    added: Vec<i64>,
    landed: Vec<Stamp>,
    finished: Vec<Stamp>,
    verdicts: Vec<Verdict>,
    unreadable_verdicts: usize,
    submits: Vec<Submit>,
    blocks: Vec<Block>,
    reservations: Vec<Reservation>,
    cards: BTreeMap<i64, String>,
}

/// `flow-measure`. Reads only; a project without a ledger answers `ledger: "absent"` with empty
/// lists (exit 0) and nothing is created.
pub fn measure(project: &Path, home: &Path, request: &Request) -> Result<Value> {
    let id = ledger::command::project_id_of(project)?;
    let (clock, data) = match ledger::open::open_read_only_within(home, &id, WAIT) {
        Ok(ledger) => {
            let (clock, data) = read(&ledger)?;
            (clock, Some(data))
        }
        Err(LedgerError::Absent) => (system_clock(), None),
        Err(other) => return Err(other.to_string()),
    };
    let now = request.now.unwrap_or(clock);
    let present = data.is_some();
    let data = data.unwrap_or_default();
    if present {
        for card in &request.cards {
            if !data.cards.contains_key(card) {
                return Err(format!(
                    "flow-measure refused: card not found: there is no {}",
                    input::card_name(*card)
                ));
            }
        }
    }
    let mut answer = compute(&data, request, now)?;
    answer["project_id"] = json!(id);
    answer["ledger"] = json!(if present { "present" } else { "absent" });
    Ok(answer)
}

fn system_clock() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| i64::try_from(d.as_secs()).unwrap_or(i64::MAX))
}

/// The statements, one read transaction, no rendering while it is open.
fn read(ledger: &ledger::ReadLedger) -> Result<(i64, Data)> {
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let tx = ledger.conn().unchecked_transaction().map_err(db)?;
    let now: i64 = tx
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .map_err(db)?;
    let mut data = Data::default();

    let mut stmt = tx.prepare("SELECT card_id, state FROM card").map_err(db)?;
    let rows = stmt
        .query_map([], |r| Ok((r.get::<_, i64>(0)?, r.get::<_, String>(1)?)))
        .map_err(db)?;
    for row in rows {
        let (id, state) = row.map_err(db)?;
        data.cards.insert(id, state);
    }
    drop(stmt);

    let mut stmt = tx.prepare(EVENTS_SQL).map_err(db)?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, i64>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, Option<i64>>(3)?,
                r.get::<_, String>(4)?,
                r.get::<_, Option<String>>(5)?,
            ))
        })
        .map_err(db)?;
    for row in rows {
        let (id, at, actor, card, kind, payload) = row.map_err(db)?;
        match (kind.as_str(), card) {
            ("card.add", _) => data.added.push(at),
            ("card.landed", Some(card)) => data.landed.push(Stamp { card, at }),
            ("card.finish", Some(card)) => data.finished.push(Stamp { card, at }),
            ("card.verdict", Some(card)) => match verdict_of(payload.as_deref()) {
                Some((pass, head)) => data.verdicts.push(Verdict {
                    id,
                    at,
                    actor,
                    card,
                    pass,
                    head,
                }),
                None => data.unreadable_verdicts += 1,
            },
            _ => {}
        }
    }
    drop(stmt);

    let mut stmt = tx.prepare(NOTES_SQL).map_err(db)?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, i64>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, String>(3)?,
                r.get::<_, i64>(4)?,
                r.get::<_, String>(5)?,
            ))
        })
        .map_err(db)?;
    for row in rows {
        let (card, note_id, kind, author, at, line) = row.map_err(db)?;
        if line.starts_with("LANE-BLOCKED") {
            data.blocks.push(Block {
                at,
                kind: block_kind(&line),
            });
        } else if marker_of_line(&line).is_some()
            && kind != "decision"
            && !author.starts_with("leader")
        {
            data.submits.push(Submit { card, note_id, at });
        }
    }
    drop(stmt);

    let mut stmt = tx.prepare(RESERVATIONS_SQL).map_err(db)?;
    let rows = stmt
        .query_map([], |r| {
            Ok(Reservation {
                card: r.get(0)?,
                generation: r.get(1)?,
                status: r.get(2)?,
                acquired: r.get(3)?,
                updated: r.get(4)?,
            })
        })
        .map_err(db)?;
    for row in rows {
        data.reservations.push(row.map_err(db)?);
    }
    drop(stmt);
    drop(tx);
    Ok((now, data))
}

/// `(pass, head)` of a verdict payload; `None` when it does not read.
fn verdict_of(payload: Option<&str>) -> Option<(bool, Option<String>)> {
    let v: Value = serde_json::from_str(payload?).ok()?;
    let pass = match v.get("verdict")?.as_str()? {
        "pass" => true,
        "fail" => false,
        _ => return None,
    };
    let head = v.get("head").and_then(Value::as_str).map(str::to_string);
    Some((pass, head))
}

/// The `kind=<word>` of a `LANE-BLOCKED` line, or `unspecified`.
fn block_kind(line: &str) -> String {
    let word = line
        .split_whitespace()
        .find_map(|w| w.strip_prefix("kind="))
        .map(|w| {
            w.chars()
                .take_while(|c| c.is_ascii_alphanumeric() || *c == '_' || *c == '-')
                .collect::<String>()
        })
        .unwrap_or_default();
    if word.is_empty() {
        "unspecified".to_string()
    } else {
        word
    }
}

/// The value at nearest rank `ceil(p * n / 100)` of an ascending list; `None` for no sample.
fn nearest_rank(sorted: &[i64], percent: usize) -> Option<i64> {
    let n = sorted.len();
    if n == 0 {
        return None;
    }
    let rank = (percent * n).div_ceil(100).max(1);
    sorted.get(rank - 1).copied()
}

/// `{n, p50_sec, p90_sec}` and, when `raw`, the ascending raw values.
fn summary(values: &[i64], raw: bool) -> Value {
    let mut sorted = values.to_vec();
    sorted.sort_unstable();
    let mut out = json!({
        "n": sorted.len(),
        "p50_sec": nearest_rank(&sorted, 50),
        "p90_sec": nearest_rank(&sorted, 90),
    });
    if raw {
        out["values_sec"] = json!(sorted);
    }
    out
}

/// The rows grouped by card, each group in time order.
struct Index<'a> {
    submits: HashMap<i64, Vec<&'a Submit>>,
    verdicts: HashMap<i64, Vec<&'a Verdict>>,
    passes: HashMap<i64, Vec<&'a Verdict>>,
    finishes: HashMap<i64, Vec<&'a Stamp>>,
    reservations: HashMap<i64, Vec<&'a Reservation>>,
}

impl<'a> Index<'a> {
    fn new(data: &'a Data) -> Self {
        let mut submits: Vec<&Submit> = data.submits.iter().collect();
        submits.sort_by_key(|s| (s.at, s.note_id));
        let mut verdicts: Vec<&Verdict> = data.verdicts.iter().collect();
        verdicts.sort_by_key(|v| (v.at, v.id));
        let mut finishes: Vec<&Stamp> = data.finished.iter().collect();
        finishes.sort_by_key(|s| s.at);
        let mut index = Index {
            submits: HashMap::new(),
            verdicts: HashMap::new(),
            passes: HashMap::new(),
            finishes: HashMap::new(),
            reservations: HashMap::new(),
        };
        for s in submits {
            index.submits.entry(s.card).or_default().push(s);
        }
        for v in verdicts {
            index.verdicts.entry(v.card).or_default().push(v);
            if v.pass {
                index.passes.entry(v.card).or_default().push(v);
            }
        }
        for f in finishes {
            index.finishes.entry(f.card).or_default().push(f);
        }
        for r in &data.reservations {
            index.reservations.entry(r.card).or_default().push(r);
        }
        for list in index.reservations.values_mut() {
            list.sort_by_key(|r| r.generation);
        }
        index
    }

    /// The newest hand-back note of a card at or before `at`.
    fn submit_before(&self, card: i64, at: i64) -> Option<&'a Submit> {
        newest_at_or_before(self.submits.get(&card)?, at, |s| s.at)
    }

    /// The newest `pass` verdict of a card at or before `at`.
    fn pass_before(&self, card: i64, at: i64) -> Option<&'a Verdict> {
        newest_at_or_before(self.passes.get(&card)?, at, |v| v.at)
    }
}

fn newest_at_or_before<'a, T>(
    sorted: &[&'a T],
    at: i64,
    at_of: impl Fn(&T) -> i64,
) -> Option<&'a T> {
    let upto = sorted.partition_point(|x| at_of(x) <= at);
    upto.checked_sub(1).map(|i| sorted[i])
}

/// The seconds between two ends, when both are observed.
fn span(from: Option<i64>, to: Option<i64>) -> Option<i64> {
    match (from, to) {
        (Some(a), Some(b)) => Some(b - a),
        _ => None,
    }
}

/// One card, by the definitions of the module.
fn card_report(data: &Data, index: &Index<'_>, card: i64) -> Value {
    let pass = index.passes.get(&card).and_then(|p| p.last()).copied();
    let submit = pass.and_then(|p| index.submit_before(card, p.at));
    let done = pass.and_then(|p| {
        index
            .finishes
            .get(&card)?
            .iter()
            .find(|f| f.at >= p.at)
            .copied()
    });
    let fail_count = index
        .verdicts
        .get(&card)
        .map_or(0, |v| v.iter().filter(|v| !v.pass).count());
    let reservation = index
        .reservations
        .get(&card)
        .and_then(|r| r.last())
        .map(|r| {
            let left = matches!(r.status.as_str(), "released" | "reclaimed").then_some(r.updated);
            json!({"generation": r.generation, "status": r.status, "left_held_at": left})
        });
    json!({
        "card": input::card_name(card),
        "state": data.cards.get(&card),
        "submit_note": submit.map(|s| s.note_id),
        "submit_at": submit.map(|s| s.at),
        "verdict_at": pass.map(|p| p.at),
        "done_at": done.map(|d| d.at),
        "submit_to_verdict_sec": span(submit.map(|s| s.at), pass.map(|p| p.at)),
        "pass_to_done_sec": span(pass.map(|p| p.at), done.map(|d| d.at)),
        "fail_count": fail_count,
        "auditor": pass.map(|p| p.actor.as_str()),
        "head": pass.and_then(|p| p.head.as_deref()),
        "reservation": reservation,
    })
}

/// The window: events counted apart, the stage summaries, the first verdicts and the blocks.
fn window_report(data: &Data, index: &Index<'_>, hours: i64, now: i64) -> Value {
    let since = now.saturating_sub(hours.saturating_mul(3600));
    let inside = |at: i64| at >= since && at <= now;

    let mut take_to_submit: Vec<i64> = Vec::new();
    for list in index.reservations.values() {
        for (i, r) in list.iter().enumerate() {
            let until = list.get(i + 1).map_or(i64::MAX, |next| next.acquired);
            let first = index
                .submits
                .get(&r.card)
                .and_then(|s| s.iter().find(|s| s.at >= r.acquired && s.at < until));
            if let Some(s) = first.filter(|s| inside(s.at)) {
                take_to_submit.push(s.at - r.acquired);
            }
        }
    }
    let submit_to_verdict: Vec<i64> = data
        .verdicts
        .iter()
        .filter(|v| inside(v.at))
        .filter_map(|v| index.submit_before(v.card, v.at).map(|s| v.at - s.at))
        .collect();
    let pass_to_landed: Vec<i64> = data
        .landed
        .iter()
        .filter(|l| inside(l.at))
        .filter_map(|l| index.pass_before(l.card, l.at).map(|p| l.at - p.at))
        .collect();
    let pass_to_done: Vec<i64> = data
        .finished
        .iter()
        .filter(|f| inside(f.at))
        .filter_map(|f| index.pass_before(f.card, f.at).map(|p| f.at - p.at))
        .collect();

    let mut first_n = 0usize;
    let mut first_fail = 0usize;
    for list in index.verdicts.values() {
        if let Some(first) = list.first().filter(|v| inside(v.at)) {
            first_n += 1;
            if !first.pass {
                first_fail += 1;
            }
        }
    }
    let fail_rate = (first_n > 0).then(|| {
        let share = first_fail as f64 / first_n as f64;
        (share * 100.0).round() / 100.0
    });

    let mut kinds: BTreeMap<&str, usize> = BTreeMap::new();
    let mut total = 0usize;
    for b in data.blocks.iter().filter(|b| inside(b.at)) {
        *kinds.entry(b.kind.as_str()).or_default() += 1;
        total += 1;
    }
    let kinds: Map<String, Value> = kinds
        .into_iter()
        .map(|(k, n)| (k.to_string(), json!(n)))
        .collect();

    json!({
        "hours": hours, "since": since, "until": now,
        "issued": data.added.iter().filter(|a| inside(**a)).count(),
        "landed": data.landed.iter().filter(|l| inside(l.at)).count(),
        "finished": data.finished.iter().filter(|f| inside(f.at)).count(),
        "stages": {
            "take_to_submit": summary(&take_to_submit, false),
            "submit_to_verdict": summary(&submit_to_verdict, false),
            "pass_to_landed": summary(&pass_to_landed, false),
            "pass_to_done": summary(&pass_to_done, false),
        },
        "first_verdict": {"n": first_n, "fail": first_fail, "fail_rate": fail_rate},
        "lane_blocked": {"total": total, "kinds": kinds},
    })
}

/// The answer for plain rows: the named cards and the window, and the table lines of both.
fn compute(data: &Data, request: &Request, now: i64) -> Result<Value> {
    let index = Index::new(data);
    let mut table: Vec<String> = Vec::new();
    let mut answer = json!({
        "action": "measure",
        "now": now,
        "writes": false,
        "unreadable_verdicts": data.unreadable_verdicts,
        "cards": [],
        "pass_to_done": Value::Null,
        "window": Value::Null,
    });
    if !request.cards.is_empty() && !data.cards.is_empty() {
        let cards: Vec<Value> = request
            .cards
            .iter()
            .map(|c| card_report(data, &index, *c))
            .collect();
        let samples: Vec<i64> = cards
            .iter()
            .filter_map(|c| c["pass_to_done_sec"].as_i64())
            .collect();
        let overall = summary(&samples, true);
        table.extend(card_table(&cards, &overall));
        answer["cards"] = json!(cards);
        answer["pass_to_done"] = overall;
    }
    if let Some(hours) = request.hours {
        let window = window_report(data, &index, hours, now);
        table.extend(window_table(&window));
        answer["window"] = window;
    }
    answer["table"] = json!(table);
    Ok(answer)
}

/// Seconds as minutes with one decimal, or `-` for no value.
fn minutes(sec: &Value) -> String {
    match sec.as_i64() {
        Some(s) => format!("{:.1} min", s as f64 / 60.0),
        None => "-".to_string(),
    }
}

fn short(text: &Value, width: usize) -> String {
    text.as_str()
        .map_or_else(|| "-".to_string(), |t| t.chars().take(width).collect())
}

fn summary_line(name: &str, s: &Value) -> String {
    format!(
        "{name}  n={}  P50 {}  P90 {}",
        s["n"],
        minutes(&s["p50_sec"]),
        minutes(&s["p90_sec"])
    )
}

fn card_table(cards: &[Value], overall: &Value) -> Vec<String> {
    let mut lines = vec![
        "card  state  submit→verdict  pass→done  fails  auditor  head  reservation".to_string(),
    ];
    for c in cards {
        lines.push(format!(
            "{}  {}  {}  {}  {}  {}  {}  {}",
            short(&c["card"], 12),
            short(&c["state"], 8),
            minutes(&c["submit_to_verdict_sec"]),
            minutes(&c["pass_to_done_sec"]),
            c["fail_count"],
            short(&c["auditor"], 8),
            short(&c["head"], 12),
            short(&c["reservation"]["status"], 12),
        ));
    }
    lines.push(summary_line("pass_to_done", overall));
    lines
}

fn window_table(w: &Value) -> Vec<String> {
    let mut lines = vec![
        format!(
            "window {} h (unix seconds {} to {})",
            w["hours"], w["since"], w["until"]
        ),
        format!(
            "issued {} · landed {} · finished {}",
            w["issued"], w["landed"], w["finished"]
        ),
    ];
    for stage in [
        "take_to_submit",
        "submit_to_verdict",
        "pass_to_landed",
        "pass_to_done",
    ] {
        lines.push(summary_line(stage, &w["stages"][stage]));
    }
    let fv = &w["first_verdict"];
    lines.push(match fv["fail_rate"].as_f64() {
        Some(rate) => format!(
            "first_verdict  fail {} of {} ({rate:.2})",
            fv["fail"], fv["n"]
        ),
        None => "first_verdict  no verdict in the window".to_string(),
    });
    let lb = &w["lane_blocked"];
    let kinds: Vec<String> = lb["kinds"]
        .as_object()
        .map(|o| o.iter().map(|(k, n)| format!("{k} {n}")).collect())
        .unwrap_or_default();
    lines.push(format!(
        "LANE-BLOCKED {}: {}",
        lb["total"],
        kinds.join(", ")
    ));
    lines
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn nearest_rank_picks_ceil_of_p_times_n() {
        let v = [10, 20, 30, 40, 50];
        assert_eq!(nearest_rank(&v, 50), Some(30));
        assert_eq!(nearest_rank(&v, 90), Some(50));
        assert_eq!(nearest_rank(&v[..4], 50), Some(20));
        assert_eq!(nearest_rank(&v[..4], 90), Some(40));
        assert_eq!(nearest_rank(&v[..1], 50), Some(10));
        assert_eq!(nearest_rank(&[], 50), None);
    }

    #[test]
    fn a_block_kind_is_the_word_after_kind_equals() {
        assert_eq!(block_kind("LANE-BLOCKED kind=decision C-1 x"), "decision");
        assert_eq!(
            block_kind("LANE-BLOCKED kind=permission: C-1"),
            "permission"
        );
        assert_eq!(block_kind("LANE-BLOCKED C-1 no word"), "unspecified");
        assert_eq!(block_kind("LANE-BLOCKED kind= C-1"), "unspecified");
    }

    #[test]
    fn the_arguments_are_checked_by_name() {
        let cards = |c: &[&str]| c.iter().map(|s| s.to_string()).collect::<Vec<_>>();
        assert!(parse_request(&[], None, None)
            .unwrap_err()
            .contains("--hours"));
        assert!(parse_request(&cards(&["C-x"]), None, None)
            .unwrap_err()
            .contains("C-x"));
        assert!(parse_request(&cards(&["#3"]), None, None)
            .unwrap_err()
            .contains("#3"));
        assert!(parse_request(&[], Some("0"), None)
            .unwrap_err()
            .contains("--hours"));
        assert!(parse_request(&[], Some("abc"), None)
            .unwrap_err()
            .contains("--hours"));
        let ok = parse_request(&cards(&["C-3", "C-3", "C-4"]), Some("12"), Some(5)).unwrap();
        assert_eq!(ok.cards, vec![3, 4]);
        assert_eq!((ok.hours, ok.now), (Some(12), Some(5)));
    }
}
