//! `audit-queue` (SPEC-20261120-audit-request-notes): the open audit requests of a ledger, read
//! from card notes of four fixed first lines, with their age, state and the time of each stage.
//!
//! The judgement is a function of plain rows ([`judge`]), so its edges are tested without a
//! ledger; [`run`] only reads the rows and prints them. Times are the ledger's own note times,
//! never a time written in a note. Nothing is written.

use moai_cli::card::input::card_name;
use moai_cli::ledger::{self, LedgerError};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::PathBuf;

/// List the open audit requests of the ledger, with age, state and stage times
///
/// Reads the local notes of every card that is not done and lists the audit requests that wait.
/// A note counts by its first line, in one of four shapes: `AUDIT-REQUEST <card> head=<40-hex>
/// tier=<0|1|2> scope=<text>` asks for an audit of that head; `AUDIT-CLAIM <card> head=<40-hex>
/// by=<name>` says an auditor took it; `AUDIT-VERDICT <card> head=<40-hex> verdict=<PASS|FAIL>`
/// closes it; `REPAIR <card> head=<40-hex> of=<40-hex>` names the head that repairs a FAIL and is
/// the request for its re-audit. Each listed request has one state: `waiting` (no claim yet),
/// `running` (claimed, no verdict), `repair` (a FAIL no repair answers yet) or `stale` (waiting
/// or running for longer than `--stale-secs`, with `was` naming the state it left). The auditor
/// of a card is the `by` of its first claim, so a repair head goes to the same auditor. A
/// request for a head that already has a verdict is `reuse`: the audit of that head is not run
/// again. `times` gives the seconds from request to claim, claim to verdict, verdict to repair
/// and repair to the next claim, computed from the ledger's note times; a span without an end
/// note is null. A first line that starts with one of the four words but breaks its shape is
/// listed under `malformed` and not counted. `--all` also lists the heads that are finished.
/// Writes nothing; a project without a ledger has nothing waiting (exit 0).
#[derive(clap::Args)]
pub struct Args {
    /// The project folder (its profile holds the project id).
    #[arg(long, default_value = ".")]
    project: PathBuf,
    /// The HOME holding the ledger (default `$HOME`). Tests pass a temporary folder.
    #[arg(long)]
    home: Option<PathBuf>,
    /// Seconds in `waiting` or `running` after which a request is `stale` (1 to 604800).
    #[arg(long, default_value_t = 1800, value_parser = clap::value_parser!(i64).range(1..=604800))]
    stale_secs: i64,
    /// Also list the heads that have a verdict (`passed`, `repaired`) with their times.
    #[arg(long)]
    all: bool,
}

pub const OPERATION: &str = "audit.queue";

/// The four first-line words.
const WORDS: [&str; 4] = ["AUDIT-REQUEST", "AUDIT-CLAIM", "AUDIT-VERDICT", "REPAIR"];

/// What one note says.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Event {
    Request {
        head: String,
        tier: u8,
        scope: String,
    },
    Claim {
        head: String,
        by: String,
    },
    Verdict {
        head: String,
        pass: bool,
    },
    Repair {
        head: String,
        of: String,
    },
}

/// A note's first line read against the four shapes.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Parsed {
    /// Not one of the four words: ordinary text.
    Other,
    /// One of the four words with a broken shape.
    Malformed,
    Event(Event),
}

fn is_head(text: &str) -> bool {
    text.len() == 40
        && text
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

fn value<'a>(token: Option<&'a str>, key: &str) -> Option<&'a str> {
    token?.strip_prefix(key)?.strip_prefix('=')
}

/// Reads the first line of a note of card `card` (its name, `C-1`).
pub fn parse(line: &str, card: &str) -> Parsed {
    let line = line.trim();
    let Some(word) = WORDS.iter().find(|w| {
        line.strip_prefix(**w)
            .is_some_and(|rest| rest.is_empty() || rest.starts_with(char::is_whitespace))
    }) else {
        return Parsed::Other;
    };
    let rest = line[word.len()..].trim_start();
    let mut tokens = rest.split_whitespace();
    if tokens.next() != Some(card) {
        return Parsed::Malformed;
    }
    let head = match value(tokens.next(), "head") {
        Some(h) if is_head(h) => h.to_string(),
        _ => return Parsed::Malformed,
    };
    let event = match *word {
        "AUDIT-REQUEST" => {
            let tier = match value(tokens.next(), "tier") {
                Some("0") => 0,
                Some("1") => 1,
                Some("2") => 2,
                _ => return Parsed::Malformed,
            };
            let Some(at) = rest.find("scope=") else {
                return Parsed::Malformed;
            };
            let scope = rest[at + "scope=".len()..].trim();
            if scope.is_empty() {
                return Parsed::Malformed;
            }
            Event::Request {
                head,
                tier,
                scope: scope.to_string(),
            }
        }
        "AUDIT-CLAIM" => match value(tokens.next(), "by") {
            Some(by) if !by.is_empty() && tokens.next().is_none() => Event::Claim {
                head,
                by: by.to_string(),
            },
            _ => return Parsed::Malformed,
        },
        "AUDIT-VERDICT" => match value(tokens.next(), "verdict") {
            Some(v @ ("PASS" | "FAIL")) if tokens.next().is_none() => Event::Verdict {
                head,
                pass: v == "PASS",
            },
            _ => return Parsed::Malformed,
        },
        _ => match value(tokens.next(), "of") {
            Some(of) if is_head(of) && tokens.next().is_none() => Event::Repair {
                head,
                of: of.to_string(),
            },
            _ => return Parsed::Malformed,
        },
    };
    Parsed::Event(event)
}

/// A card note that carries an [`Event`].
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Row {
    pub card_id: i64,
    pub note_id: i64,
    pub author: String,
    pub at: i64,
    pub event: Event,
}

/// One audit of one head, as the notes of its card tell it.
#[derive(Clone, Debug, PartialEq)]
pub struct Item {
    pub card_id: i64,
    pub head: String,
    pub tier: Option<u8>,
    pub scope: Option<String>,
    pub of: Option<String>,
    pub state: &'static str,
    pub was: Option<&'static str>,
    pub requested_by: String,
    pub request_note_id: i64,
    pub request_at: i64,
    pub state_since: i64,
    pub auditor: Option<String>,
    pub claim_by: Option<String>,
    pub verdict: Option<&'static str>,
    pub times: [Option<i64>; 4],
}

fn span(from: Option<i64>, to: Option<i64>) -> Option<i64> {
    match (from, to) {
        (Some(a), Some(b)) if b >= a => Some(b - a),
        _ => None,
    }
}

/// Every audit the rows of one ledger describe, in request order: the main entry of each head
/// and, after a verdict, one `reuse` entry for the newest later request of the same head.
/// `now` and `stale_secs` decide `stale`. Rows come in note order.
pub fn judge(rows: &[Row], now: i64, stale_secs: i64) -> Vec<Item> {
    #[derive(Default)]
    struct Head {
        card_id: i64,
        request: Option<(i64, i64, String)>,
        tier: Option<u8>,
        scope: Option<String>,
        of: Option<String>,
        claim: Option<(i64, i64, String)>,
        verdict: Option<(i64, i64, bool)>,
        repair: Option<i64>,
        later_request: Option<(i64, i64, String)>,
    }
    let mut first_claimer: BTreeMap<i64, String> = BTreeMap::new();
    for r in rows {
        if let Event::Claim { by, .. } = &r.event {
            first_claimer.entry(r.card_id).or_insert_with(|| by.clone());
        }
    }
    let mut heads: BTreeMap<(i64, String), Head> = BTreeMap::new();
    // Requests first: a head exists from its request or from the repair that names it.
    for r in rows {
        match &r.event {
            Event::Request { head, tier, scope } => {
                let h = heads.entry((r.card_id, head.clone())).or_default();
                h.card_id = r.card_id;
                if h.request.is_none() {
                    h.request = Some((r.note_id, r.at, r.author.clone()));
                    h.tier = Some(*tier);
                    h.scope = Some(scope.clone());
                }
            }
            Event::Repair { head, of } => {
                let h = heads.entry((r.card_id, head.clone())).or_default();
                h.card_id = r.card_id;
                if h.request.is_none() {
                    h.request = Some((r.note_id, r.at, r.author.clone()));
                    h.of = Some(of.clone());
                }
                if let Some(old) = heads.get_mut(&(r.card_id, of.clone())) {
                    old.repair.get_or_insert(r.at);
                }
            }
            _ => {}
        }
    }
    // A repair head takes the tier and scope of the head it repairs.
    let repair_heads: Vec<(i64, String)> = heads
        .iter()
        .filter(|(_, h)| h.tier.is_none() && h.of.is_some())
        .map(|(key, _)| key.clone())
        .collect();
    for key in repair_heads {
        let of = heads[&key].of.clone().unwrap_or_default();
        let (tier, scope) = heads
            .get(&(key.0, of))
            .map_or((None, None), |old| (old.tier, old.scope.clone()));
        if let Some(h) = heads.get_mut(&key) {
            h.tier = tier;
            h.scope = scope;
        }
    }
    for r in rows {
        match &r.event {
            Event::Claim { head, by } => {
                if let Some(h) = heads.get_mut(&(r.card_id, head.clone())) {
                    let after = h.request.as_ref().is_some_and(|q| r.note_id > q.0);
                    if after && h.claim.is_none() {
                        h.claim = Some((r.note_id, r.at, by.clone()));
                    }
                }
            }
            Event::Verdict { head, pass } => {
                if let Some(h) = heads.get_mut(&(r.card_id, head.clone())) {
                    let floor = h
                        .claim
                        .as_ref()
                        .map(|c| c.0)
                        .or(h.request.as_ref().map(|q| q.0));
                    if floor.is_some_and(|f| r.note_id > f) && h.verdict.is_none() {
                        h.verdict = Some((r.note_id, r.at, *pass));
                    }
                }
            }
            _ => {}
        }
    }
    for r in rows {
        if let Event::Request { head, .. } = &r.event {
            if let Some(h) = heads.get_mut(&(r.card_id, head.clone())) {
                if h.verdict.as_ref().is_some_and(|v| r.note_id > v.0) {
                    h.later_request = Some((r.note_id, r.at, r.author.clone()));
                }
            }
        }
    }
    // The claim of a repair head, for the span from the repair to the next claim.
    let claim_at_of: BTreeMap<(i64, String), i64> = heads
        .iter()
        .filter_map(|(k, h)| Some((k.clone(), h.claim.as_ref()?.1)))
        .collect();
    let repaired_by: BTreeMap<(i64, String), String> = heads
        .iter()
        .filter_map(|((card, head), h)| Some(((*card, h.of.clone()?), head.clone())))
        .collect();
    let mut items: Vec<Item> = Vec::new();
    for ((card_id, head), h) in &heads {
        let Some((req_note, req_at, req_by)) = h.request.clone() else {
            continue;
        };
        let claim_at = h.claim.as_ref().map(|c| c.1);
        let verdict_at = h.verdict.as_ref().map(|v| v.1);
        let repair_next_claim = repaired_by
            .get(&(*card_id, head.clone()))
            .and_then(|new| claim_at_of.get(&(*card_id, new.clone())).copied());
        let times = [
            span(Some(req_at), claim_at),
            span(claim_at, verdict_at),
            span(verdict_at, h.repair),
            span(h.repair, repair_next_claim),
        ];
        let (mut state, since): (&'static str, i64) = match (&h.claim, &h.verdict) {
            (_, Some((_, at, true))) => ("passed", *at),
            (_, Some((_, at, false))) if h.repair.is_some() => ("repaired", *at),
            (_, Some((_, at, false))) => ("repair", *at),
            (Some((_, at, _)), None) => ("running", *at),
            (None, None) => ("waiting", req_at),
        };
        let mut was = None;
        if matches!(state, "waiting" | "running") && now - since > stale_secs {
            was = Some(state);
            state = "stale";
        }
        let verdict = h
            .verdict
            .as_ref()
            .map(|v| if v.2 { "PASS" } else { "FAIL" });
        items.push(Item {
            card_id: *card_id,
            head: head.clone(),
            tier: h.tier,
            scope: h.scope.clone(),
            of: h.of.clone(),
            state,
            was,
            requested_by: req_by,
            request_note_id: req_note,
            request_at: req_at,
            state_since: since,
            auditor: first_claimer.get(card_id).cloned(),
            claim_by: h.claim.as_ref().map(|c| c.2.clone()),
            verdict,
            times,
        });
        if let (Some((note, at, by)), Some(v)) = (&h.later_request, verdict) {
            items.push(Item {
                card_id: *card_id,
                head: head.clone(),
                tier: h.tier,
                scope: h.scope.clone(),
                of: h.of.clone(),
                state: "reuse",
                was: None,
                requested_by: by.clone(),
                request_note_id: *note,
                request_at: *at,
                state_since: *at,
                auditor: first_claimer.get(card_id).cloned(),
                claim_by: h.claim.as_ref().map(|c| c.2.clone()),
                verdict: Some(v),
                times: [None; 4],
            });
        }
    }
    items.sort_by_key(|i| (i.request_note_id, i.card_id));
    items
}

fn shown(item: &Item, now: i64) -> Value {
    let [req_claim, claim_verdict, verdict_repair, repair_claim] = item.times;
    json!({
        "card": card_name(item.card_id), "card_id": item.card_id, "head": item.head,
        "tier": item.tier, "scope": item.scope, "of": item.of,
        "state": item.state, "was": item.was,
        "requested_by": item.requested_by, "request_note_id": item.request_note_id,
        "age_secs": (now - item.request_at).max(0),
        "state_age_secs": (now - item.state_since).max(0),
        "auditor": item.auditor, "claim_by": item.claim_by, "verdict": item.verdict,
        "times": {
            "request_to_claim_secs": req_claim,
            "claim_to_verdict_secs": claim_verdict,
            "verdict_to_repair_secs": verdict_repair,
            "repair_to_claim_secs": repair_claim,
        },
    })
}

pub fn run(args: Args) -> Result<Value, String> {
    let home = ledger::command::resolve_home(args.home)?;
    let id = ledger::command::project_id_of(&args.project)?;
    let ledger = match ledger::open_read_only(&home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Ok(json!({
                "action": "audit-queue", "project_id": id, "ledger": "absent", "now": null,
                "count": 0, "requests": [], "malformed": [], "writes": false,
            }));
        }
        Err(e) => return Err(e.to_string()),
    };
    let conn = ledger.conn();
    let db = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let now: i64 = conn
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .map_err(db)?;
    // Only the first line of each note leaves the database: a note can be 1 MiB.
    let mut stmt = conn
        .prepare(
            "SELECT n.card_id, n.note_id, n.author, n.created_at, \
                    substr(CASE WHEN instr(n.body, char(10)) > 0 \
                                THEN substr(n.body, 1, instr(n.body, char(10)) - 1) \
                                ELSE n.body END, 1, 2000) \
               FROM card_note n JOIN card c ON c.card_id = n.card_id \
              WHERE c.state <> 'done' AND n.origin = 'local' \
              ORDER BY n.note_id",
        )
        .map_err(db)?;
    let mut rows: Vec<Row> = Vec::new();
    let mut malformed: Vec<Value> = Vec::new();
    let mapped = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, i64>(0)?,
                r.get::<_, i64>(1)?,
                r.get::<_, String>(2)?,
                r.get::<_, i64>(3)?,
                r.get::<_, String>(4)?,
            ))
        })
        .map_err(db)?;
    for item in mapped {
        let (card_id, note_id, author, at, line) = item.map_err(db)?;
        match parse(&line, &card_name(card_id)) {
            Parsed::Other => {}
            Parsed::Malformed => malformed.push(json!({
                "card": card_name(card_id), "card_id": card_id, "note_id": note_id,
                "line": line.trim_end().chars().take(200).collect::<String>(),
            })),
            Parsed::Event(event) => rows.push(Row {
                card_id,
                note_id,
                author,
                at,
                event,
            }),
        }
    }
    let requests: Vec<Value> = judge(&rows, now, args.stale_secs)
        .iter()
        .filter(|i| args.all || !matches!(i.state, "passed" | "repaired"))
        .map(|i| shown(i, now))
        .collect();
    Ok(json!({
        "action": "audit-queue", "project_id": id, "ledger": "ready", "now": now,
        "stale_secs": args.stale_secs, "count": requests.len(), "requests": requests,
        "malformed": malformed, "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const H1: &str = "1111111111111111111111111111111111111111";
    const H2: &str = "2222222222222222222222222222222222222222";

    fn row(note_id: i64, at: i64, event: Event) -> Row {
        Row {
            card_id: 1,
            note_id,
            author: "a".into(),
            at,
            event,
        }
    }
    fn request(head: &str) -> Event {
        Event::Request {
            head: head.into(),
            tier: 1,
            scope: "x".into(),
        }
    }
    fn claim(head: &str) -> Event {
        Event::Claim {
            head: head.into(),
            by: "aud".into(),
        }
    }

    #[test]
    fn the_four_shapes_are_read_and_a_broken_one_is_malformed() {
        let line = format!("AUDIT-REQUEST C-1 head={H1} tier=2 scope=a b c");
        assert_eq!(
            parse(&line, "C-1"),
            Parsed::Event(Event::Request {
                head: H1.into(),
                tier: 2,
                scope: "a b c".into()
            })
        );
        assert_eq!(
            parse(&format!("AUDIT-CLAIM C-1 head={H1} by=x"), "C-1"),
            Parsed::Event(Event::Claim {
                head: H1.into(),
                by: "x".into()
            })
        );
        assert_eq!(
            parse(&format!("REPAIR C-1 head={H2} of={H1}"), "C-1"),
            Parsed::Event(Event::Repair {
                head: H2.into(),
                of: H1.into()
            })
        );
        // The card token must be the note's own card; tier 3, a short head, a trailing word and
        // an empty scope are broken shapes; a word glued to a colon or in lower case is text.
        for bad in [
            format!("AUDIT-REQUEST C-2 head={H1} tier=1 scope=x"),
            format!("AUDIT-REQUEST C-1 head={H1} tier=3 scope=x"),
            format!("AUDIT-REQUEST C-1 head={H1} tier=1 scope="),
            format!("AUDIT-CLAIM C-1 head={H1} by=x extra"),
            format!("AUDIT-VERDICT C-1 head={H1} verdict=pass"),
            "REPAIR C-1 head=abc of=def".to_string(),
            "AUDIT-VERDICT".to_string(),
        ] {
            assert_eq!(parse(&bad, "C-1"), Parsed::Malformed, "{bad}");
        }
        assert_eq!(parse("REPAIR: later", "C-1"), Parsed::Other);
        assert_eq!(parse("audit-request C-1", "C-1"), Parsed::Other);
        assert_eq!(parse("REPAIRED it", "C-1"), Parsed::Other);
    }

    #[test]
    fn stale_starts_strictly_after_the_limit() {
        let rows = [row(1, 1000, request(H1))];
        assert_eq!(judge(&rows, 1100, 100)[0].state, "waiting");
        assert_eq!(judge(&rows, 1101, 100)[0].state, "stale");
        assert_eq!(judge(&rows, 1101, 100)[0].was, Some("waiting"));
        let rows = [row(1, 1000, request(H1)), row(2, 1050, claim(H1))];
        assert_eq!(judge(&rows, 1150, 100)[0].state, "running");
        assert_eq!(judge(&rows, 1151, 100)[0].state, "stale");
    }

    #[test]
    fn a_claim_before_its_request_does_not_count() {
        let rows = [row(1, 900, claim(H1)), row(2, 1000, request(H1))];
        let items = judge(&rows, 1010, 100);
        assert_eq!(items[0].state, "waiting");
        assert_eq!(items[0].claim_by, None);
    }

    #[test]
    fn a_span_with_an_end_before_its_start_is_null() {
        assert_eq!(span(Some(10), Some(5)), None);
        assert_eq!(span(Some(10), Some(10)), Some(0));
        assert_eq!(span(None, Some(10)), None);
    }
}
