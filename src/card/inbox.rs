//! What `moai card inbox` does (card C-116, SPEC-20261119-leader-inbox): the lane notes that
//! still wait for the leader, in one read-only query instead of a regex over the last two notes
//! of every card.
//!
//! A *request* is a note whose first line starts with one of the prefixes in [`family_of`]. It
//! is *answered* when its card holds a later note (a larger note id) written by the leader, whose
//! name is the note's `author`. On each card only the newest request of each family is listed,
//! and only when it is not answered. Cards that are `done` and notes imported from a code host
//! are left out. The judgement is [`pending`], a function of plain rows, so its edges are tested
//! without a ledger; [`inbox`] only reads the rows and prints them.

use super::input;
use crate::ledger::{self, LedgerError};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The author of the leader's notes when `--leader` is not given.
pub const DEFAULT_LEADER: &str = "leader";
/// The longest `--leader`, in characters.
pub const LEADER_MAX_CHARS: usize = 80;
/// The longest first line printed, in characters.
const LINE_MAX_CHARS: usize = 200;

/// The kinds of request a lane makes of the leader.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub enum Family {
    AuditRequest,
    Blocked,
    ReadyForIntegration,
    DecisionQuestion,
    /// A lane's `READY …` note (the M-Flow bundle flow).
    Ready,
    /// A lane's `MERGE-READY …` note.
    MergeReady,
    /// A lane's `CANDIDATE …` note (the lane-merge pair lines).
    Candidate,
}

impl Family {
    /// The name printed as `family`.
    pub fn code(self) -> &'static str {
        match self {
            Family::AuditRequest => "audit_request",
            Family::Blocked => "blocked",
            Family::ReadyForIntegration => "ready_for_integration",
            Family::DecisionQuestion => "decision_question",
            Family::Ready => "ready",
            Family::MergeReady => "merge_ready",
            Family::Candidate => "candidate",
        }
    }
}

/// The family of a note's first line, when it is a request (SPEC REQ-2). Prefixes are measured
/// from the lanes' notes of 2026-11-19; a new spelling is one more line here.
pub fn family_of(first_line: &str) -> Option<Family> {
    let line = first_line.trim_start();
    if line.starts_with("감사 요청")
        || (line.starts_with("LANE-DONE") && line.contains("audit requested"))
    {
        Some(Family::AuditRequest)
    } else if line.starts_with("LANE-BLOCKED") {
        Some(Family::Blocked)
    } else if line.starts_with("LANE-READY-INTEGRATION") {
        Some(Family::ReadyForIntegration)
    } else if line.starts_with("DECISION-QUESTION") {
        Some(Family::DecisionQuestion)
    } else if line.starts_with("READY") {
        Some(Family::Ready)
    } else if line.starts_with("MERGE-READY") {
        Some(Family::MergeReady)
    } else if line.starts_with("CANDIDATE") {
        Some(Family::Candidate)
    } else {
        None
    }
}

/// A local note of a card that is not done, as the query reads it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Row {
    pub card_id: i64,
    pub note_id: i64,
    pub author: String,
    /// The note's first line.
    pub line: String,
    pub created_at: i64,
}

/// The requests that wait: per card, the newest request of each family, when no note by
/// `leader` with a larger note id is on the card. Oldest note first. `rows` come in any order.
pub fn pending<'a>(rows: &'a [Row], leader: &str) -> Vec<(Family, &'a Row)> {
    let mut last_leader: BTreeMap<i64, i64> = BTreeMap::new();
    let mut newest: BTreeMap<(i64, Family), &Row> = BTreeMap::new();
    for row in rows {
        if row.author == leader {
            let at = last_leader.entry(row.card_id).or_insert(row.note_id);
            *at = (*at).max(row.note_id);
        } else if let Some(family) = family_of(&row.line) {
            let slot = newest.entry((row.card_id, family)).or_insert(row);
            if row.note_id > slot.note_id {
                *slot = row;
            }
        }
    }
    let mut waiting: Vec<(Family, &Row)> = newest
        .into_iter()
        .filter(|((card_id, _), row)| {
            last_leader
                .get(card_id)
                .is_none_or(|answered| row.note_id > *answered)
        })
        .map(|((_, family), row)| (family, row))
        .collect();
    waiting.sort_by_key(|(_, row)| row.note_id);
    waiting
}

/// `--leader`: one line of 1 to [`LEADER_MAX_CHARS`] characters.
pub fn leader_name(given: Option<&str>) -> Result<String> {
    let name = given.unwrap_or(DEFAULT_LEADER);
    if name.trim().is_empty()
        || name.chars().count() > LEADER_MAX_CHARS
        || name.chars().any(char::is_control)
    {
        return Err(format!(
            "card inbox refused: --leader must be one line of 1 to {LEADER_MAX_CHARS} characters"
        ));
    }
    Ok(name.to_string())
}

/// The first line of a note, cut to [`LINE_MAX_CHARS`] characters.
fn shown(line: &str) -> String {
    line.trim_end().chars().take(LINE_MAX_CHARS).collect()
}

/// `card inbox [--leader NAME]`. Reads only; a project without a ledger has nothing waiting
/// (exit 0) and nothing is created.
pub fn inbox(project: &Path, home: &Path, leader: Option<&str>) -> Result<Value> {
    let leader = leader_name(leader)?;
    let id = ledger::command::project_id_of(project)?;
    let ledger = match ledger::open_read_only(home, &id) {
        Ok(l) => l,
        Err(LedgerError::Absent) => {
            return Ok(json!({
                "action": "inbox", "project_id": id, "ledger": "absent", "leader": leader,
                "now": null, "count": 0, "items": [], "writes": false,
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
                                ELSE n.body END, 1, 2000), \
                    c.title, c.state \
               FROM card_note n JOIN card c ON c.card_id = n.card_id \
              WHERE c.state <> 'done' AND n.origin = 'local' \
              ORDER BY n.note_id",
        )
        .map_err(db)?;
    let mut rows: Vec<Row> = Vec::new();
    let mut cards: BTreeMap<i64, (String, String)> = BTreeMap::new();
    let mapped = stmt
        .query_map([], |r| {
            Ok((
                Row {
                    card_id: r.get(0)?,
                    note_id: r.get(1)?,
                    author: r.get(2)?,
                    created_at: r.get(3)?,
                    line: r.get(4)?,
                },
                r.get::<_, String>(5)?,
                r.get::<_, String>(6)?,
            ))
        })
        .map_err(db)?;
    for item in mapped {
        let (row, title, state) = item.map_err(db)?;
        cards.entry(row.card_id).or_insert((title, state));
        rows.push(row);
    }
    let items: Vec<Value> = pending(&rows, &leader)
        .into_iter()
        .map(|(family, row)| {
            let (title, state) = &cards[&row.card_id];
            json!({
                "card": input::card_name(row.card_id), "card_id": row.card_id,
                "title": title, "state": state, "family": family.code(),
                "note_id": row.note_id, "author": row.author, "created_at": row.created_at,
                "age_secs": (now - row.created_at).max(0), "line": shown(&row.line),
            })
        })
        .collect();
    Ok(json!({
        "action": "inbox", "project_id": id, "ledger": "ready", "leader": leader,
        "now": now, "count": items.len(), "items": items, "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn row(card_id: i64, note_id: i64, author: &str, line: &str) -> Row {
        Row {
            card_id,
            note_id,
            author: author.to_string(),
            line: line.to_string(),
            created_at: 0,
        }
    }

    fn ids(rows: &[Row], leader: &str) -> Vec<(i64, Family)> {
        pending(rows, leader)
            .into_iter()
            .map(|(family, row)| (row.note_id, family))
            .collect()
    }

    #[test]
    fn family_is_read_from_the_start_of_the_first_line_only() {
        for (line, want) in [
            ("감사 요청 tier 2 head=abc", Some(Family::AuditRequest)),
            ("  감사 요청: PR #3111", Some(Family::AuditRequest)),
            (
                "LANE-DONE agent-1 C-1 PR #7 head abc — audit requested",
                Some(Family::AuditRequest),
            ),
            ("LANE-DONE agent-1 C-1 PR #7 head abc — done", None),
            ("LANE-BLOCKED agent-1 C-1 x", Some(Family::Blocked)),
            (
                "LANE-READY-INTEGRATION C-1 head=abc",
                Some(Family::ReadyForIntegration),
            ),
            (
                "DECISION-QUESTION agent-8 (C-61)",
                Some(Family::DecisionQuestion),
            ),
            ("LANE-NOTE agent-1 C-33 question?", None),
            ("see LANE-BLOCKED in the middle", None),
            ("lane-blocked in lower case", None),
            // The M-Flow families the flow report asked for (READY, MERGE-READY, CANDIDATE).
            (
                "READY C-7 head=abc six light checks ok",
                Some(Family::Ready),
            ),
            (
                "MERGE-READY C-7 head=abc base=def",
                Some(Family::MergeReady),
            ),
            ("CANDIDATE pair head=abc base=def", Some(Family::Candidate)),
            ("ready in lower case is not a request", None),
            ("see MERGE-READY later in the line", None),
            ("", None),
        ] {
            assert_eq!(family_of(line), want, "{line:?}");
        }
    }

    #[test]
    fn a_leader_note_answers_only_a_request_with_a_smaller_note_id() {
        let blocked = "LANE-BLOCKED a";
        // Answered by the next id, and still open when the leader wrote the id before.
        let after = [row(1, 10, "agent-1", blocked), row(1, 11, "leader", "x")];
        assert_eq!(ids(&after, "leader"), vec![]);
        let before = [row(1, 10, "leader", "x"), row(1, 11, "agent-1", blocked)];
        assert_eq!(ids(&before, "leader"), vec![(11, Family::Blocked)]);
        // The leader's note on another card answers nothing here.
        let other = [row(1, 10, "agent-1", blocked), row(2, 11, "leader", "x")];
        assert_eq!(ids(&other, "leader"), vec![(10, Family::Blocked)]);
    }

    #[test]
    fn the_new_families_follow_the_answer_and_newest_rules() {
        // A leader note written after a READY note answers it, whatever its family.
        let answered = [
            row(1, 10, "agent-1", "READY C-1 head=abc"),
            row(1, 11, "leader", "MERGED-LOCAL C-1"),
        ];
        assert_eq!(ids(&answered, "leader"), vec![]);
        // A READY note written after the last leader note waits.
        let open = [
            row(1, 10, "leader", "x"),
            row(1, 11, "agent-1", "MERGE-READY C-1 head=abc"),
        ];
        assert_eq!(ids(&open, "leader"), vec![(11, Family::MergeReady)]);
        // Only the newest CANDIDATE of a card is listed.
        let superseded = [
            row(1, 10, "agent-1", "CANDIDATE old head=abc"),
            row(1, 11, "agent-1", "CANDIDATE new head=def"),
        ];
        assert_eq!(ids(&superseded, "leader"), vec![(11, Family::Candidate)]);
    }

    #[test]
    fn the_newest_leader_note_of_a_card_decides_whatever_order_the_rows_come_in() {
        let blocked = "LANE-BLOCKED a";
        // leader 10 → request 11 → leader 12: the later leader note answers the request, so
        // choosing the smallest leader id (10 < 11) would keep it open by mistake.
        let rows = [
            row(1, 10, "leader", "x"),
            row(1, 11, "agent-1", blocked),
            row(1, 12, "leader", "y"),
        ];
        assert_eq!(ids(&rows, "leader"), vec![]);
        let reversed = [rows[2].clone(), rows[1].clone(), rows[0].clone()];
        assert_eq!(ids(&reversed, "leader"), vec![]);
        // A request written after the last leader note stays open.
        let open = [
            row(1, 10, "leader", "x"),
            row(1, 12, "leader", "y"),
            row(1, 13, "agent-1", blocked),
        ];
        assert_eq!(ids(&open, "leader"), vec![(13, Family::Blocked)]);
    }

    #[test]
    fn only_the_newest_request_of_a_family_is_listed_and_rows_come_in_any_order() {
        let rows = [
            row(1, 12, "agent-1", "LANE-BLOCKED new"),
            row(1, 10, "agent-1", "LANE-BLOCKED old"),
            row(1, 11, "agent-1", "감사 요청 head=abc"),
            row(2, 13, "agent-2", "LANE-BLOCKED other card"),
        ];
        assert_eq!(
            ids(&rows, "leader"),
            vec![
                (11, Family::AuditRequest),
                (12, Family::Blocked),
                (13, Family::Blocked)
            ]
        );
        // An old request that was answered does not hide a newer one.
        let rows = [
            row(1, 1, "agent-1", "LANE-BLOCKED old"),
            row(1, 2, "leader", "answer"),
            row(1, 3, "agent-1", "LANE-BLOCKED new"),
        ];
        assert_eq!(ids(&rows, "leader"), vec![(3, Family::Blocked)]);
    }

    #[test]
    fn the_leader_never_asks_itself_and_a_renamed_leader_changes_who_answers() {
        // Two families, so neither note hides the other.
        let rows = [
            row(1, 1, "leader", "LANE-BLOCKED note to self"),
            row(1, 2, "agent-1", "감사 요청 head=abc"),
        ];
        // `leader` wrote note 1, which is no request to itself, and agent-1's note 2 waits.
        assert_eq!(ids(&rows, "leader"), vec![(2, Family::AuditRequest)]);
        // With `boss` as the leader, `leader` is a lane: its note 1 is a request that waits too.
        assert_eq!(
            ids(&rows, "boss"),
            vec![(1, Family::Blocked), (2, Family::AuditRequest)]
        );
    }

    #[test]
    fn the_leader_name_has_the_edges_of_a_by_name() {
        assert_eq!(leader_name(None).unwrap(), "leader");
        assert!(leader_name(Some(&"a".repeat(80))).is_ok());
        for bad in [
            "".to_string(),
            "   ".to_string(),
            "a".repeat(81),
            "a\nb".to_string(),
        ] {
            let e = leader_name(Some(&bad)).unwrap_err();
            assert!(e.starts_with("card inbox refused: --leader"), "{e}");
        }
    }

    #[test]
    fn a_printed_line_is_cut_at_the_limit() {
        assert_eq!(
            shown(&"x".repeat(LINE_MAX_CHARS)).chars().count(),
            LINE_MAX_CHARS
        );
        assert_eq!(
            shown(&"x".repeat(LINE_MAX_CHARS + 1)).chars().count(),
            LINE_MAX_CHARS
        );
        assert_eq!(shown("감사 요청  "), "감사 요청");
    }
}
