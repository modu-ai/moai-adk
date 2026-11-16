//! The ledger side of a host import (#3071 H3a): the SQLite implementation of
//! [`ImportLedger`], the port that `forge::import_apply` writes through. The port's types are
//! defined in this file and `forge::import_apply` re-exports them, so the ledger library depends
//! on nothing of the host side.
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
use rusqlite::{params, OptionalExtension, Transaction};

/// One acceptance criterion. The ledger numbers the rows `AC1`… in this order.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct BatchAc {
    pub text: String,
    pub verify: String,
}

/// One note, always `origin = 'import'`. `reference` is the host's comment id as text.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct BatchNote {
    pub reference: String,
    /// `comment` or `observation`.
    pub kind: &'static str,
    pub author: String,
    pub body: String,
}

/// One card to create unless its source is already in the ledger.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct BatchCard {
    pub source_host: &'static str,
    pub source_project: String,
    pub source_number: u64,
    pub title: String,
    pub kind: &'static str,
    /// The issue body exactly as written.
    pub body: String,
    pub acs: Vec<BatchAc>,
    pub notes: Vec<BatchNote>,
    /// The `event.event_key` of the `import` event: `import:<host>:<project>:<number>:<hash>`.
    pub event_key: String,
    /// The `event.payload` JSON.
    pub event_payload: String,
}

/// Everything one import writes.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ImportBatch {
    /// The digest the plan had, which the user was shown.
    pub digest: String,
    pub cards: Vec<BatchCard>,
}

/// What the ledger did with one card.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum CardOutcome {
    Created,
    /// A card with this source was there; nothing of it was written, notes included.
    AlreadyThere,
}

/// The ledger's answer: one outcome per card, in the order of the batch.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ImportCommit {
    pub outcomes: Vec<CardOutcome>,
}

/// The ledger could not write the batch; it wrote none of it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ImportFault(pub String);

/// The ledger side of an import. An implementation writes all or none of a batch.
pub trait ImportLedger {
    fn commit(&mut self, batch: &ImportBatch) -> Result<ImportCommit, ImportFault>;
}

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
