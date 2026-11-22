//! What `moai card wait` does (SPEC-20261122-card-wait): wait, with a timeout that must be given,
//! for the next note a lane cares about, and print it once.
//!
//! The wait fixes a start mark (`--since`, else the largest note id in the ledger), then polls the
//! ledger read-only every `--every` seconds. A poll keeps the local notes above the mark that are
//! on a named card, or whose first line holds `all lanes` in any case on a card that is not done,
//! and leaves out the notes of a `--not-author` name. The first poll that keeps any ends the wait
//! with all of them (`outcome: "note"`); the timeout ends it with none (`outcome: "timeout"`).
//! Nothing is printed while it waits, and nothing is written. The judgement is [`matching`], a
//! function of plain rows, so its edges are tested without a ledger; [`wait`] reads the rows.

use super::input::{self, CardRef};
use crate::ledger::{self, LedgerError};
use rusqlite::Connection;
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::path::Path;
use std::time::{Duration, Instant};

type Result<T> = std::result::Result<T, String>;

/// The longest `--timeout`, in seconds: 12 hours.
pub const TIMEOUT_MAX_SECS: u64 = 12 * 60 * 60;
/// `--every` when it is not given, in seconds.
pub const EVERY_DEFAULT_SECS: u64 = 15;
/// The longest `--every`, in seconds.
pub const EVERY_MAX_SECS: u64 = 300;
/// The longest `--not-author`, in characters.
pub const AUTHOR_MAX_CHARS: usize = 80;
/// The longest first line printed, in characters (as `card inbox` prints it).
const LINE_MAX_CHARS: usize = 200;
/// How long one poll waits for a writer before it counts as a poll that found nothing.
const POLL_BUSY: Duration = Duration::from_millis(250);
/// The text whose presence in a first line, in any case, makes a note an announcement.
const ANNOUNCEMENT: &str = "all lanes";

/// `<n>`, `<n>s`, `<n>m` or `<n>h` in seconds: ASCII digits, no sign, no space, no fraction.
fn seconds(text: &str) -> Option<u64> {
    let (digits, unit) = match text.as_bytes().last()? {
        b's' => (&text[..text.len() - 1], 1),
        b'm' => (&text[..text.len() - 1], 60),
        b'h' => (&text[..text.len() - 1], 60 * 60),
        _ => (text, 1),
    };
    if digits.is_empty() || !digits.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    digits.parse::<u64>().ok()?.checked_mul(unit)
}

fn within(text: &str, max: u64, limit: &str) -> Result<u64> {
    match seconds(text) {
        Some(n) if (1..=max).contains(&n) => Ok(n),
        _ => Err(format!(
            "write <n>, <n>s, <n>m or <n>h between 1 second and {limit}"
        )),
    }
}

/// `--timeout`: 1 second to [`TIMEOUT_MAX_SECS`]. The command line uses it as its parser, so a
/// value it refuses is a usage error before the ledger is read.
pub fn timeout_arg(text: &str) -> Result<u64> {
    within(text, TIMEOUT_MAX_SECS, "12 hours")
}

/// `--every`: 1 to [`EVERY_MAX_SECS`] seconds, a usage error otherwise (as [`timeout_arg`]).
pub fn every_arg(text: &str) -> Result<u64> {
    within(text, EVERY_MAX_SECS, "300 seconds")
}

/// A local note above the mark, as a poll reads it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Row {
    pub card_id: i64,
    pub note_id: i64,
    pub kind: String,
    pub author: String,
    /// The note's first line.
    pub line: String,
    pub created_at: i64,
    /// The note's card is `done`.
    pub done: bool,
}

/// Whether a first line is an announcement to every lane: it holds `all lanes` in any case.
fn announces(line: &str) -> bool {
    line.to_lowercase().contains(ANNOUNCEMENT)
}

/// The notes that end a wait, by note id: above `mark`, by none of `not_authors`, and on one of
/// `cards` or announcing to all lanes on a card that is not done. `rows` come in any order.
pub fn matching<'a>(
    rows: &'a [Row],
    cards: &BTreeSet<i64>,
    not_authors: &[String],
    mark: i64,
) -> Vec<&'a Row> {
    let mut found: Vec<&Row> = rows
        .iter()
        .filter(|row| row.note_id > mark)
        .filter(|row| !not_authors.contains(&row.author))
        .filter(|row| cards.contains(&row.card_id) || (!row.done && announces(&row.line)))
        .collect();
    found.sort_by_key(|row| row.note_id);
    found
}

/// What `card wait` was asked, after the command line read it.
pub struct WaitRequest<'a> {
    /// The card names, `C-<n>` or `#<source number>`; at least one.
    pub cards: &'a [String],
    pub timeout_secs: u64,
    pub every_secs: u64,
    /// The start mark; `None` takes the largest note id in the ledger.
    pub since: Option<i64>,
    pub not_authors: &'a [String],
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card wait refused: {why}")
}

fn db(e: rusqlite::Error) -> String {
    LedgerError::from(e).to_string()
}

/// The card ids of the names, each checked against the ledger.
fn resolve(conn: &Connection, refs: &[CardRef]) -> Result<BTreeSet<i64>> {
    let mut ids = BTreeSet::new();
    for which in refs {
        match *which {
            CardRef::Id(n) => {
                let found: bool = conn
                    .query_row(
                        "SELECT EXISTS(SELECT 1 FROM card WHERE card_id = ?1)",
                        [n],
                        |r| r.get(0),
                    )
                    .map_err(db)?;
                if !found {
                    return Err(format!(
                        "card not found: there is no {}",
                        input::card_name(n)
                    ));
                }
                ids.insert(n);
            }
            CardRef::SourceNumber(n) => {
                let mut stmt = conn
                    .prepare("SELECT card_id FROM card WHERE source_number = ?1 ORDER BY card_id")
                    .map_err(db)?;
                let found = stmt
                    .query_map([n], |r| r.get::<_, i64>(0))
                    .map_err(db)?
                    .collect::<std::result::Result<Vec<_>, _>>()
                    .map_err(db)?;
                match found.as_slice() {
                    [one] => {
                        ids.insert(*one);
                    }
                    [] => return Err(format!("card not found: no card came from #{n}")),
                    many => {
                        let names: Vec<String> =
                            many.iter().map(|c| input::card_name(*c)).collect();
                        return Err(refused(format!(
                            "#{n} matches {} cards ({}); name one by its C-number",
                            many.len(),
                            names.join(", ")
                        )));
                    }
                }
            }
        }
    }
    Ok(ids)
}

/// One poll: the local notes above `mark`, with only the first line of each leaving the
/// database (a note can be 1 MiB). A ledger that another process holds past [`POLL_BUSY`] is a
/// poll that found nothing; the next one reads it again.
fn poll(home: &Path, project_id: &str, mark: i64) -> Result<Vec<Row>> {
    let read = || -> std::result::Result<Vec<Row>, LedgerError> {
        let ledger = ledger::open::open_read_only_within(home, project_id, POLL_BUSY)?;
        let mut stmt = ledger.conn().prepare(
            "SELECT n.card_id, n.note_id, n.kind, n.author, n.created_at, \
                    substr(CASE WHEN instr(n.body, char(10)) > 0 \
                                THEN substr(n.body, 1, instr(n.body, char(10)) - 1) \
                                ELSE n.body END, 1, 2000), \
                    c.state = 'done' \
               FROM card_note n JOIN card c ON c.card_id = n.card_id \
              WHERE n.origin = 'local' AND n.note_id > ?1 \
              ORDER BY n.note_id",
        )?;
        let rows = stmt
            .query_map([mark], |r| {
                Ok(Row {
                    card_id: r.get(0)?,
                    note_id: r.get(1)?,
                    kind: r.get(2)?,
                    author: r.get(3)?,
                    created_at: r.get(4)?,
                    line: r.get(5)?,
                    done: r.get(6)?,
                })
            })?
            .collect::<std::result::Result<Vec<_>, _>>()?;
        Ok(rows)
    };
    match read() {
        Ok(rows) => Ok(rows),
        Err(LedgerError::Busy) => Ok(Vec::new()),
        Err(e) => Err(e.to_string()),
    }
}

/// Polls until `found` returns something or `deadline` passes: the first poll at once, then one
/// every `every`, the last one at the deadline. It never runs past the deadline plus one poll.
fn until<T>(
    deadline: Instant,
    every: Duration,
    mut found: impl FnMut() -> Result<Vec<T>>,
) -> Result<Vec<T>> {
    loop {
        let got = found()?;
        if !got.is_empty() {
            return Ok(got);
        }
        let now = Instant::now();
        if now >= deadline {
            return Ok(got);
        }
        std::thread::sleep(every.min(deadline - now));
    }
}

/// The first line of a note, cut to [`LINE_MAX_CHARS`] characters.
fn shown(line: &str) -> String {
    line.trim_end().chars().take(LINE_MAX_CHARS).collect()
}

/// `card wait --card NAME… --timeout T [--every S] [--since N] [--not-author NAME…]`. Reads only.
/// An unknown card, an absent ledger or one the ledger refuses is refused before the wait, and
/// nothing is created.
pub fn wait(project: &Path, home: &Path, request: &WaitRequest) -> Result<Value> {
    if !(1..=TIMEOUT_MAX_SECS).contains(&request.timeout_secs) {
        return Err(refused("--timeout must be 1 second to 12 hours"));
    }
    if !(1..=EVERY_MAX_SECS).contains(&request.every_secs) {
        return Err(refused("--every must be 1 to 300 seconds"));
    }
    if request.cards.is_empty() {
        return Err(refused("name at least one card with --card"));
    }
    let refs = request
        .cards
        .iter()
        .map(|name| input::parse_ref(name.trim()).map_err(refused))
        .collect::<Result<Vec<CardRef>>>()?;
    for name in request.not_authors {
        if name.trim().is_empty()
            || name.chars().count() > AUTHOR_MAX_CHARS
            || name.chars().any(char::is_control)
        {
            return Err(refused(format!(
                "--not-author must be one line of 1 to {AUTHOR_MAX_CHARS} characters"
            )));
        }
    }
    if request.since.is_some_and(|n| n < 0) {
        return Err(refused("--since must be 0 or more"));
    }
    let id = ledger::command::project_id_of(project)?;
    // The names and the mark come from one read-only connection, closed before the wait.
    let (cards, mark) = {
        let ledger = ledger::open_read_only(home, &id).map_err(|e| match e {
            LedgerError::Absent => super::command::ABSENT_HINT.to_string(),
            other => other.to_string(),
        })?;
        let conn = ledger.conn();
        let cards = resolve(conn, &refs)?;
        let mark = match request.since {
            Some(n) => n,
            None => conn
                .query_row("SELECT COALESCE(MAX(note_id), 0) FROM card_note", [], |r| {
                    r.get(0)
                })
                .map_err(db)?,
        };
        (cards, mark)
    };
    let started = Instant::now();
    let deadline = started + Duration::from_secs(request.timeout_secs);
    let rows = until(deadline, Duration::from_secs(request.every_secs), || {
        let rows = poll(home, &id, mark)?;
        let kept: Vec<Row> = matching(&rows, &cards, request.not_authors, mark)
            .into_iter()
            .cloned()
            .collect();
        Ok(kept)
    })?;
    let waited = started.elapsed().as_secs();
    let items: Vec<Value> = rows
        .iter()
        .map(|row| {
            json!({
                "card": input::card_name(row.card_id), "card_id": row.card_id,
                "note_id": row.note_id, "kind": row.kind, "author": row.author,
                "created_at": row.created_at, "line": shown(&row.line),
            })
        })
        .collect();
    let names: Vec<String> = cards.iter().map(|c| input::card_name(*c)).collect();
    Ok(json!({
        "action": "wait", "project_id": id,
        "outcome": if items.is_empty() { "timeout" } else { "note" },
        "cards": names, "not_authors": request.not_authors, "mark": mark,
        "waited_secs": waited, "count": items.len(), "items": items, "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn row(card_id: i64, note_id: i64, author: &str, line: &str, done: bool) -> Row {
        Row {
            card_id,
            note_id,
            kind: "comment".to_string(),
            author: author.to_string(),
            line: line.to_string(),
            created_at: 0,
            done,
        }
    }

    fn ids(rows: &[Row], cards: &[i64], not_authors: &[&str], mark: i64) -> Vec<i64> {
        let cards: BTreeSet<i64> = cards.iter().copied().collect();
        let not_authors: Vec<String> = not_authors.iter().map(|s| s.to_string()).collect();
        matching(rows, &cards, &not_authors, mark)
            .into_iter()
            .map(|row| row.note_id)
            .collect()
    }

    #[test]
    fn a_note_on_a_named_card_is_kept_and_one_on_another_card_is_not() {
        let rows = [
            row(1, 10, "leader", "answer", false),
            row(2, 11, "leader", "answer to someone else", false),
            row(3, 12, "leader", "on a named card that is done", true),
        ];
        assert_eq!(ids(&rows, &[1, 3], &[], 0), vec![10, 12]);
        assert_eq!(ids(&rows, &[2], &[], 0), vec![11]);
        assert_eq!(ids(&rows, &[], &[], 0), Vec::<i64>::new());
    }

    #[test]
    fn all_lanes_in_the_first_line_in_any_case_is_kept_on_any_card_that_is_not_done() {
        let rows = [
            row(5, 1, "leader", "ALL LANES: stop pushing", false),
            row(6, 2, "leader", "notice to all lanes", false),
            row(7, 3, "leader", "All Lanes rebase now", false),
            row(8, 4, "leader", "all lanes on a done card", true),
            row(9, 5, "leader", "all-lanes is not the phrase", false),
            row(9, 6, "leader", "alllanes neither", false),
        ];
        assert_eq!(ids(&rows, &[1], &[], 0), vec![1, 2, 3]);
        assert!(announces("ALL LANES"));
        assert!(!announces("all  lanes"));
        assert!(!announces(""));
    }

    #[test]
    fn a_not_author_name_is_skipped_on_a_named_card_and_in_an_announcement() {
        let rows = [
            row(1, 1, "agent-1", "LANE-NOTE agent-1 progress", false),
            row(1, 2, "leader", "answer", false),
            row(2, 3, "agent-1", "all lanes: from agent-1", false),
            row(2, 4, "agent-2", "all lanes: from agent-2", false),
        ];
        assert_eq!(ids(&rows, &[1], &["agent-1"], 0), vec![2, 4]);
        assert_eq!(ids(&rows, &[1], &["agent-1", "agent-2"], 0), vec![2]);
        assert_eq!(ids(&rows, &[1], &[], 0), vec![1, 2, 3, 4]);
        // The name is compared whole and as typed.
        assert_eq!(ids(&rows, &[1], &["agent"], 0), vec![1, 2, 3, 4]);
        assert_eq!(ids(&rows, &[1], &["Agent-1"], 0), vec![1, 2, 3, 4]);
    }

    #[test]
    fn the_mark_is_a_strict_bound_and_rows_come_out_in_note_order() {
        let rows = [
            row(1, 12, "leader", "c", false),
            row(1, 10, "leader", "a", false),
            row(1, 11, "leader", "b", false),
        ];
        assert_eq!(ids(&rows, &[1], &[], 9), vec![10, 11, 12]);
        assert_eq!(ids(&rows, &[1], &[], 10), vec![11, 12]);
        assert_eq!(ids(&rows, &[1], &[], 11), vec![12]);
        assert_eq!(ids(&rows, &[1], &[], 12), Vec::<i64>::new());
    }

    #[test]
    fn a_duration_is_read_with_its_unit_and_its_limits() {
        assert_eq!(seconds("7"), Some(7));
        assert_eq!(seconds("7s"), Some(7));
        assert_eq!(seconds("7m"), Some(420));
        assert_eq!(seconds("7h"), Some(25_200));
        for bad in [
            "", "s", "m", "-1", "+1", "1.5s", "1 s", " 1", "5x", "1S", "1d", "١",
        ] {
            assert_eq!(seconds(bad), None, "{bad:?}");
        }
        assert_eq!(seconds(&format!("{}h", u64::MAX)), None, "overflow");

        // --timeout: 1 second to 12 hours, at each edge.
        for ok in ["1", "1s", "43200", "43200s", "720m", "12h"] {
            assert!(timeout_arg(ok).is_ok(), "{ok}");
        }
        for bad in ["0", "0s", "0h", "43201", "721m", "13h", "5x", ""] {
            assert!(timeout_arg(bad).is_err(), "{bad}");
        }
        assert_eq!(timeout_arg("12h"), Ok(TIMEOUT_MAX_SECS));
        // --every: 1 to 300 seconds.
        for ok in ["1", "300", "300s", "5m"] {
            assert!(every_arg(ok).is_ok(), "{ok}");
        }
        for bad in ["0", "301", "6m", "1h"] {
            assert!(every_arg(bad).is_err(), "{bad}");
        }
    }

    #[test]
    fn the_wait_polls_until_something_is_found_or_the_deadline_passes() {
        // Found on the third poll: the wait ends there.
        let mut polls = 0;
        let got = until(
            Instant::now() + Duration::from_secs(30),
            Duration::from_millis(1),
            || {
                polls += 1;
                Ok(if polls == 3 { vec![polls] } else { vec![] })
            },
        )
        .unwrap();
        assert_eq!((got, polls), (vec![3], 3));
        // A deadline already passed: one poll, then the timeout.
        let mut polls = 0;
        let got: Vec<i32> = until(Instant::now(), Duration::from_secs(300), || {
            polls += 1;
            Ok(vec![])
        })
        .unwrap();
        assert_eq!((got, polls), (vec![], 1));
        // A refused poll ends the wait with its error.
        let err = until::<i32>(
            Instant::now() + Duration::from_secs(30),
            Duration::from_millis(1),
            || Err("ledger blocked: x".to_string()),
        )
        .unwrap_err();
        assert_eq!(err, "ledger blocked: x");
    }

    #[test]
    fn a_printed_line_is_cut_at_the_limit() {
        assert_eq!(
            shown(&"x".repeat(LINE_MAX_CHARS + 1)).chars().count(),
            LINE_MAX_CHARS
        );
        assert_eq!(shown("answer  "), "answer");
    }
}
