//! A card read as the short text a person sees on the host (#3071 H4). Reads only, through a
//! [`ReadLedger`]; nothing here writes a row, reads a clock, or touches the network. The text is
//! built from the card's own columns, so it carries no secret the ledger does not already hold,
//! and `forge::card_export` still passes it through the connector's redaction before it is posted.
//!
//! The card's body and notes are left out on purpose: the summary is a status line for people
//! (title, kind, state, where the card came from, each criterion and whether it is met), not a
//! copy of the card.
use super::open::ReadLedger;
use super::LedgerError;
use rusqlite::{params, OptionalExtension};

/// Where a card came from on a host.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CardSource {
    pub host: String,
    /// Lower case `owner/repo`.
    pub project: String,
    pub number: u64,
}

/// One acceptance criterion with its state.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CriterionLine {
    pub seq: i64,
    pub text: String,
    /// `open`, `met` or `withdrawn`.
    pub status: String,
    pub evidence_ref: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CardSummary {
    pub card_id: i64,
    pub title: String,
    pub kind: String,
    pub state: String,
    pub source: Option<CardSource>,
    pub criteria: Vec<CriterionLine>,
}

fn sql(error: rusqlite::Error) -> LedgerError {
    LedgerError::Sql(error.to_string())
}

/// The card with this id, or `None` when the ledger has none. The ledger numbers cards; a card
/// is written `C-<id>` for people.
pub fn read_card_summary(
    ledger: &ReadLedger,
    card_id: i64,
) -> Result<Option<CardSummary>, LedgerError> {
    let conn = ledger.conn();
    let head = conn
        .query_row(
            "SELECT title, kind, state, source_host, source_project, source_number
             FROM card WHERE card_id = ?1",
            params![card_id],
            |row| {
                Ok((
                    row.get::<_, String>(0)?,
                    row.get::<_, String>(1)?,
                    row.get::<_, String>(2)?,
                    row.get::<_, Option<String>>(3)?,
                    row.get::<_, Option<String>>(4)?,
                    row.get::<_, Option<i64>>(5)?,
                ))
            },
        )
        .optional()
        .map_err(sql)?;
    let Some((title, kind, state, host, project, number)) = head else {
        return Ok(None);
    };
    let source = match (host, project, number) {
        (Some(host), Some(project), Some(number)) => {
            u64::try_from(number).ok().map(|number| CardSource {
                host,
                project,
                number,
            })
        }
        _ => None,
    };
    let mut statement = conn
        .prepare(
            "SELECT seq, text, status, evidence_ref FROM card_ac
             WHERE card_id = ?1 ORDER BY seq",
        )
        .map_err(sql)?;
    let criteria = statement
        .query_map(params![card_id], |row| {
            Ok(CriterionLine {
                seq: row.get(0)?,
                text: row.get(1)?,
                status: row.get(2)?,
                evidence_ref: row.get(3)?,
            })
        })
        .map_err(sql)?
        .collect::<Result<Vec<_>, _>>()
        .map_err(sql)?;
    Ok(Some(CardSummary {
        card_id,
        title,
        kind,
        state,
        source,
        criteria,
    }))
}

/// The marker that says "this card's summary is already there". One per card, so the second run
/// posts nothing even when the card has moved on since.
pub fn summary_marker(card_id: i64) -> String {
    format!("<!-- moai-card-summary C-{card_id} -->")
}

impl CardSummary {
    /// The text posted to the host. A status line, then one line per criterion. Criterion text is
    /// quoted with a bullet and cut to one line so it cannot start a heading, a table or a quote
    /// of its own.
    pub fn render(&self) -> String {
        let mut out = format!(
            "**MoAI card C-{}** · {} · {}\n\n{}\n",
            self.card_id,
            self.kind,
            self.state,
            one_line(&self.title, 300)
        );
        if self.criteria.is_empty() {
            out.push_str("\n_No acceptance criteria are recorded on this card._\n");
        } else {
            out.push_str("\nAcceptance criteria:\n");
            for line in &self.criteria {
                let mark = match line.status.as_str() {
                    "met" => "met",
                    "withdrawn" => "withdrawn",
                    _ => "open",
                };
                out.push_str(&format!(
                    "- AC{} [{mark}] {}\n",
                    line.seq,
                    one_line(&line.text, 300)
                ));
            }
        }
        out
    }
}

/// The first line of `text`, cut to `max` characters.
fn one_line(text: &str, max: usize) -> String {
    let first = text.lines().next().unwrap_or("").trim();
    let mut cut: String = first.chars().take(max).collect();
    if first.chars().count() > max {
        cut.push('…');
    }
    cut
}
