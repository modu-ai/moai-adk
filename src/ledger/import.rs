//! The ledger side of a host import (#3071 H3a): the SQLite implementation of
//! [`ImportLedger`], the port `forge::import_apply` writes through.
//!
//! One call writes a whole [`ImportBatch`] in one write transaction ([`WriteLedger::with_write_tx`]):
//! every card the ledger does not have yet, with its criteria, its notes and one `import` event,
//! or nothing at all when any statement fails. A card whose
//! `(source_host, source_project, source_number)` is already there is left exactly as it is, notes
//! included, and is reported as [`CardOutcome::AlreadyThere`]; that source key is also `UNIQUE`
//! in the schema, so two processes importing at once cannot make two cards.
//!
//! What this writes is text. A card is `draft` (the column default); there is no reservation,
//! no state change, no link and no closing. Numbers come from the table, times from the DB clock
//! of the transaction, and the actor is [`IMPORT_ACTOR`]. Each criterion row is numbered
//! `1..` in the order of the batch, status `open`. The card's `rev` goes up by one for every
//! criterion row (the schema's own trigger), so a fresh import has `rev = 1 + criteria`.
//!
//! Nothing here reads a clock, the network, `git` or the HOME; the caller opens the ledger.
use super::open::WriteLedger;
use super::LedgerError;
use crate::forge::import_apply::{
    BatchCard, CardOutcome, ImportBatch, ImportCommit, ImportFault, ImportLedger,
};
use rusqlite::{params, OptionalExtension, Transaction};

/// `card.created_by` of an imported card and `event.actor` of its `import` event.
pub const IMPORT_ACTOR: &str = "import";

/// The `event.kind` of the one event an imported card gets.
pub const IMPORT_EVENT: &str = "import";

/// [`ImportLedger`] over an open write connection.
pub struct SqliteImport<'a> {
    ledger: &'a mut WriteLedger,
}

impl<'a> SqliteImport<'a> {
    pub fn new(ledger: &'a mut WriteLedger) -> Self {
        Self { ledger }
    }
}

impl ImportLedger for SqliteImport<'_> {
    fn commit(&mut self, batch: &ImportBatch) -> Result<ImportCommit, ImportFault> {
        self.ledger
            .with_write_tx(|tx, now| {
                let mut outcomes = Vec::with_capacity(batch.cards.len());
                for card in &batch.cards {
                    outcomes.push(write_card(tx, now, card)?);
                }
                Ok(ImportCommit { outcomes })
            })
            .map_err(|error: LedgerError| ImportFault(error.to_string()))
    }
}

fn write_card(
    tx: &Transaction<'_>,
    now: i64,
    card: &BatchCard,
) -> Result<CardOutcome, LedgerError> {
    let number = i64::try_from(card.source_number)
        .map_err(|_| LedgerError::Constraint("source number out of range".into()))?;
    let there: Option<i64> = tx
        .query_row(
            "SELECT card_id FROM card \
             WHERE source_host = ?1 AND source_project = ?2 AND source_number = ?3",
            params![card.source_host, card.source_project, number],
            |row| row.get(0),
        )
        .optional()?;
    if there.is_some() {
        return Ok(CardOutcome::AlreadyThere);
    }
    tx.execute(
        "INSERT INTO card (title, body, kind, created_by, created_at, updated_at, \
                           source_host, source_project, source_number) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?5, ?6, ?7, ?8)",
        params![
            card.title,
            card.body,
            card.kind,
            IMPORT_ACTOR,
            now,
            card.source_host,
            card.source_project,
            number
        ],
    )?;
    let card_id = tx.last_insert_rowid();
    for (index, ac) in card.acs.iter().enumerate() {
        tx.execute(
            "INSERT INTO card_ac (card_id, seq, text, verify, updated_at) \
             VALUES (?1, ?2, ?3, ?4, ?5)",
            params![card_id, index as i64 + 1, ac.text, ac.verify, now],
        )?;
    }
    for note in &card.notes {
        tx.execute(
            "INSERT INTO card_note (card_id, kind, author, origin, body, ref, created_at) \
             VALUES (?1, ?2, ?3, 'import', ?4, ?5, ?6)",
            params![
                card_id,
                note.kind,
                note.author,
                note.body,
                note.reference,
                now
            ],
        )?;
    }
    tx.execute(
        "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
        params![
            card.event_key,
            now,
            IMPORT_ACTOR,
            card_id,
            IMPORT_EVENT,
            card.event_payload
        ],
    )?;
    Ok(CardOutcome::Created)
}
