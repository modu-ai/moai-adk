//! Cards that landed and still have open criteria: a "partial landing". The ledger already keeps
//! the facts (a `card.landed` event or a merged pull request link, and the criteria rows), and
//! this module only reads them together, so `card list`, `card next` and the status line can show
//! such a card apart from a plain `ready` one with the number of criteria that are left. It
//! writes nothing and runs nothing outside the ledger.

use super::LedgerError;
use rusqlite::Connection;

/// One card that is landed and not finished.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PartialLanding {
    pub card_id: i64,
    pub title: String,
    /// The criteria still `open` (a withdrawn criterion is not counted).
    pub open_criteria: i64,
    /// The head of the newest `card.landed` event that names one; none for a close by a merged
    /// pull request link or a report.
    pub head: Option<String>,
}

/// The SQL condition of a partial landing on the table `card`, written once for this module and
/// for the status line: the card is `ready`, `active` or `review`, it was landed (a
/// `card.landed` event or an active merged pull request link, the same two proofs `card list`
/// reports as `landed`) and at least one criterion is `open`. The landed events are read as one
/// list: `event` has no index on `card_id`, so a subquery per card would scan the whole table
/// for each of them, on every status line render.
macro_rules! partial_landing_where {
    () => {
        "card.state IN ('ready', 'active', 'review') \
         AND (card.card_id IN (SELECT e.card_id FROM event e \
                                WHERE e.kind = 'card.landed' AND e.card_id IS NOT NULL) \
              OR EXISTS (SELECT 1 FROM card_link l \
                          WHERE l.from_card = card.card_id AND l.kind = 'pr' \
                            AND l.state = 'active' AND l.pr_state = 'merged')) \
         AND EXISTS (SELECT 1 FROM card_ac a \
                      WHERE a.card_id = card.card_id AND a.status = 'open')"
    };
}
pub(crate) use partial_landing_where;

/// The cards that are not `done`, were landed and keep at least one open criterion, in card
/// order (see [`partial_landing_where`]). A `draft` and a `hold` card is never listed: the first
/// is not worked on and the second waits for a person's pick.
pub fn partial_landings(conn: &Connection) -> Result<Vec<PartialLanding>, LedgerError> {
    // The heads come from one pass over the landed events: with a single `max()` SQLite takes
    // the other columns of the group from the row that holds it, here the newest event.
    let mut stmt = conn.prepare(concat!(
        "SELECT card.card_id, card.title, \
                (SELECT count(*) FROM card_ac a \
                  WHERE a.card_id = card.card_id AND a.status = 'open'), \
                landed.head \
           FROM card \
           LEFT JOIN (SELECT e.card_id, json_extract(e.payload, '$.head') AS head, \
                             max(e.event_id) \
                        FROM event e \
                       WHERE e.kind = 'card.landed' AND json_type(e.payload, '$.head') = 'text' \
                       GROUP BY e.card_id) AS landed ON landed.card_id = card.card_id \
          WHERE ",
        partial_landing_where!(),
        " ORDER BY card.card_id"
    ))?;
    let rows = stmt.query_map([], |r| {
        Ok(PartialLanding {
            card_id: r.get(0)?,
            title: r.get(1)?,
            open_criteria: r.get(2)?,
            head: r.get(3)?,
        })
    })?;
    Ok(rows.collect::<Result<_, _>>()?)
}

/// An open criterion that a passing audit note of the landed head names by its number. This is a
/// pointer for a person, never a proof: the note's words are matched, and the criterion is
/// marked met only by `card finish --met N=REF` or `card criterion --met`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct MetCandidate {
    pub criterion: i64,
    /// The `ref` of the audit note that names the criterion.
    pub note: String,
}

/// The most criteria one note can name (a guard against a note that spells out a huge range).
const MAX_NAMED: i64 = 200;

/// The open criteria of the card that a note of the form
/// `AUDIT <card> … head <head> … — verdict: PASS` (first line, the head the card was landed at)
/// names by number, in criterion order, each with the newest such note. A note whose first line
/// is not an audit line of this head, or whose verdict is not `PASS`, names nothing.
pub fn met_candidates(
    conn: &Connection,
    card_id: i64,
    head: &str,
    open: &[i64],
) -> Result<Vec<MetCandidate>, LedgerError> {
    if open.is_empty() {
        return Ok(Vec::new());
    }
    let mut stmt =
        conn.prepare("SELECT ref, body FROM card_note WHERE card_id = ?1 ORDER BY note_id DESC")?;
    let notes: Vec<(String, String)> = stmt
        .query_map([card_id], |r| Ok((r.get(0)?, r.get(1)?)))?
        .collect::<Result<_, _>>()?;
    let mut found: Vec<MetCandidate> = Vec::new();
    for (reference, body) in notes {
        let first = body.lines().next().unwrap_or("");
        if !is_pass_audit_line(first, head) {
            continue;
        }
        let named = named_criteria(&body);
        for n in open {
            if named.contains(n) && !found.iter().any(|c| c.criterion == *n) {
                found.push(MetCandidate {
                    criterion: *n,
                    note: reference.clone(),
                });
            }
        }
    }
    found.sort_by_key(|c| c.criterion);
    Ok(found)
}

/// Whether `line` is the first line of a passing audit note of `head`: it starts with `AUDIT `,
/// carries `head <head>` as a word pair and ends with `verdict: PASS`.
fn is_pass_audit_line(line: &str, head: &str) -> bool {
    let line = line.trim();
    if !line.starts_with("AUDIT ") || !line.ends_with("verdict: PASS") {
        return false;
    }
    let words: Vec<&str> = line.split_whitespace().collect();
    words.windows(2).any(|w| w[0] == "head" && w[1] == head)
}

/// The criterion numbers a note names after the words `criterion`, `criteria`, `AC` or `기준`:
/// one number (`기준 3`), a range (`기준 1~4`, `criteria 2-3`) or a list (`기준 1, 2 and 4`).
fn named_criteria(body: &str) -> Vec<i64> {
    let chars: Vec<char> = body.chars().collect();
    // Lower-cased one character at a time, so `lower` and `chars` always have the same indices.
    let lower: Vec<char> = chars
        .iter()
        .map(|c| c.to_lowercase().next().unwrap_or(*c))
        .collect();
    let mut named: Vec<i64> = Vec::new();
    for key in ["criteria", "criterion", "ac", "기준"] {
        let key: Vec<char> = key.chars().collect();
        let mut at = 0;
        while at + key.len() <= lower.len() {
            if lower[at..at + key.len()] == key[..] && word_start(&lower, at, key[0]) {
                let mut i = at + key.len();
                let mut first = true;
                loop {
                    // A separator before the next number: spaces, `-` after `ac`, `#`, `,`, `·`
                    // and the words `and`, `와`, `과`.
                    let mut j = i;
                    while j < chars.len() && matches!(chars[j], ' ' | '-' | '#' | ',' | '·') {
                        j += 1;
                    }
                    if !first {
                        for word in ["and ", "와 ", "과 "] {
                            let w: Vec<char> = word.chars().collect();
                            if lower[j..].starts_with(&w[..]) {
                                j += w.len();
                            }
                        }
                    }
                    let Some((from, end)) = number_at(&chars, j) else {
                        break;
                    };
                    let mut to = from;
                    let mut end = end;
                    if end < chars.len() && matches!(chars[end], '~' | '-' | '–') {
                        if let Some((upper, after)) = number_at(&chars, end + 1) {
                            to = upper;
                            end = after;
                        }
                    }
                    if to >= from && to - from < MAX_NAMED {
                        named.extend(from..=to);
                    }
                    first = false;
                    i = end;
                }
            }
            at += 1;
        }
    }
    named.sort_unstable();
    named.dedup();
    named
}

/// A key word starts a word: the character before it is not a letter or digit (Korean text
/// has no space before a particle, but the key `기준` is matched at its start only).
fn word_start(lower: &[char], at: usize, first: char) -> bool {
    if first == '기' {
        return true;
    }
    at == 0 || !lower[at - 1].is_alphanumeric()
}

/// The number written at `at` (digits only, at most 4), and the index after it.
fn number_at(chars: &[char], at: usize) -> Option<(i64, usize)> {
    let mut end = at;
    while end < chars.len() && chars[end].is_ascii_digit() && end - at < 4 {
        end += 1;
    }
    if end == at {
        return None;
    }
    let text: String = chars[at..end].iter().collect();
    text.parse::<i64>()
        .ok()
        .filter(|n| *n >= 1)
        .map(|n| (n, end))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn named_criteria_reads_a_number_a_range_and_a_list() {
        assert_eq!(
            named_criteria("카드: 기준 1~4 가 원장에서 met"),
            vec![1, 2, 3, 4]
        );
        assert_eq!(named_criteria("criterion 3 is met"), vec![3]);
        assert_eq!(named_criteria("criteria 2-3 pass"), vec![2, 3]);
        assert_eq!(named_criteria("기준 1, 2 and 4 are covered"), vec![1, 2, 4]);
        assert_eq!(named_criteria("AC-2 holds, AC 5 too"), vec![2, 5]);
        assert_eq!(named_criteria("기준 1과 3"), vec![1, 3]);
    }

    #[test]
    fn named_criteria_ignores_words_that_only_contain_a_key() {
        assert_eq!(named_criteria("a trace 4 and a space 7"), Vec::<i64>::new());
        assert_eq!(
            named_criteria("no number after the criteria"),
            Vec::<i64>::new()
        );
        assert_eq!(named_criteria("기준이 없어요"), Vec::<i64>::new());
    }

    #[test]
    fn a_partial_landing_carries_the_newest_head_a_landed_event_names() {
        let conn = Connection::open_in_memory().unwrap();
        conn.execute_batch(
            "CREATE TABLE card (card_id INTEGER PRIMARY KEY, title TEXT, state TEXT);
             CREATE TABLE card_ac (card_id INTEGER, status TEXT);
             CREATE TABLE card_link (from_card INTEGER, kind TEXT, state TEXT, pr_state TEXT);
             CREATE TABLE event (event_id INTEGER PRIMARY KEY, card_id INTEGER, kind TEXT,
                                 payload TEXT);
             INSERT INTO card VALUES (1, 'two heads', 'ready'), (2, 'merged link', 'review'),
                                     (3, 'no head', 'active'), (4, 'not landed', 'ready'),
                                     (5, 'on hold', 'hold');
             INSERT INTO card_ac VALUES (1, 'open'), (1, 'met'), (2, 'open'), (2, 'open'),
                                        (3, 'open'), (4, 'open'), (5, 'open');
             INSERT INTO card_link VALUES (2, 'pr', 'active', 'merged');
             INSERT INTO event VALUES
                 (7, 1, 'card.landed', '{\"head\":\"newest\"}'),
                 (3, 1, 'card.landed', '{\"head\":\"older\"}'),
                 (9, 1, 'card.landed', '{}'),
                 (10, 1, 'card.note', '{\"head\":\"other kind\"}'),
                 (11, 3, 'card.landed', '{\"head\":null}'),
                 (12, 5, 'card.landed', '{\"head\":\"held\"}'),
                 (13, NULL, 'card.landed', '{\"head\":\"no card\"}');",
        )
        .unwrap();
        let found: Vec<(i64, i64, Option<String>)> = partial_landings(&conn)
            .unwrap()
            .into_iter()
            .map(|p| (p.card_id, p.open_criteria, p.head))
            .collect();
        assert_eq!(
            found,
            [
                (1, 1, Some("newest".to_string())),
                (2, 2, None),
                (3, 1, None),
            ]
        );
    }

    #[test]
    fn a_huge_range_names_nothing() {
        assert_eq!(named_criteria("기준 1~9999"), Vec::<i64>::new());
    }

    #[test]
    fn only_a_passing_audit_line_of_the_head_counts() {
        let head = "a".repeat(40);
        let good = format!("AUDIT C-1 (scope) head {head} base {head} — verdict: PASS");
        assert!(is_pass_audit_line(&good, &head));
        assert!(!is_pass_audit_line(&good.replace("PASS", "FAIL"), &head));
        assert!(!is_pass_audit_line(&good, &"b".repeat(40)));
        assert!(!is_pass_audit_line("REPORT head x — verdict: PASS", &head));
    }
}
