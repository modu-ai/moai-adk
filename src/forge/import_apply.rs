//! Applying an import plan (#3071 H2b): the half that writes, behind one port.
//!
//! [`apply_import`] reads the host again, plans the import again and writes only when the digest
//! of what it read is the digest the user was shown. Anything that differs, or a listing that
//! cannot be read to its end, writes nothing. The writes go through [`ImportLedger`], one call
//! per import: the ledger's implementation puts the whole batch in one transaction, skips a card
//! whose `(source_host, source_project, source_number)` is already there, and gives numbers,
//! times and the actor itself. This module reads no clock, file, environment or HOME.
//!
//! What a batch holds is text. A card is `draft`; the batch has no reservation, no state and no
//! link, so importing cannot hold, start or finish anything. Notes keep the host's comment id as
//! their `ref` so a second import adds nothing.
use super::connector::{ConnectorError, HostConnector, ProjectRef};
use super::import_plan::{field, hash, plan_import, AcReading, ImportPlan, PlannedCard};
use serde_json::json;
use sha2::{Digest, Sha256};
use std::collections::BTreeMap;
use std::fmt;

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

#[derive(Debug)]
pub enum ApplyError {
    /// The host could not be read, to its end, as asked.
    Host(ConnectorError),
    /// The host no longer gives the issue set the user was shown; nothing was written.
    HostChanged { expected: String, found: String },
    /// The ledger refused the batch and wrote none of it.
    Ledger(ImportFault),
    /// The ledger's answer does not match the batch, so what it wrote is not known.
    LedgerAnswer { cards: usize, outcomes: usize },
}

impl fmt::Display for ApplyError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ApplyError::Host(e) => write!(f, "{e}"),
            ApplyError::HostChanged { expected, found } => write!(
                f,
                "the open issues changed since the preview (preview {expected}, now {found}); nothing was written"
            ),
            ApplyError::Ledger(e) => write!(f, "the ledger wrote nothing: {}", e.0),
            ApplyError::LedgerAnswer { cards, outcomes } => write!(
                f,
                "the ledger answered for {outcomes} cards but the batch had {cards}; check the ledger before importing again"
            ),
        }
    }
}

impl std::error::Error for ApplyError {}

/// What an import did, by issue number.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ImportReport {
    pub digest: String,
    pub created: Vec<u64>,
    pub already_there: Vec<u64>,
    /// Numbers that came more than once from the host and got no card.
    pub duplicate_numbers: Vec<u64>,
    /// Created cards that have no criteria (the table could not be read, or had no rows).
    pub created_without_criteria: Vec<u64>,
    /// Notes written for created cards, by kind.
    pub notes_written: BTreeMap<&'static str, usize>,
    pub pull_requests_skipped: usize,
}

/// A hash of what a card would be made of, for the event key: the same card text gives the same
/// key, and any change to the title, kind, body, criteria or notes gives another.
pub fn card_content_hash(card: &PlannedCard) -> String {
    let mut out = b"moai-import-card-v1\n".to_vec();
    field(&mut out, &card.title);
    field(&mut out, card.kind);
    field(&mut out, &hash(&card.body));
    if let AcReading::Read { rows, .. } = &card.ac {
        field(&mut out, &rows.len().to_string());
        for row in rows {
            field(&mut out, &row.label);
            field(&mut out, &row.text);
            field(&mut out, &row.verify);
        }
    } else {
        field(&mut out, "none");
    }
    field(&mut out, &card.notes.len().to_string());
    for note in &card.notes {
        field(&mut out, &note.comment_id.to_string());
        field(&mut out, note.kind.as_str());
        field(&mut out, &note.author);
        field(&mut out, &hash(&note.body));
    }
    format!("{:x}", Sha256::digest(&out))
}

fn batch_card(plan: &ImportPlan, card: &PlannedCard) -> BatchCard {
    let acs: Vec<BatchAc> = match &card.ac {
        AcReading::Read { rows, .. } => rows
            .iter()
            .map(|r| BatchAc {
                text: r.text.clone(),
                verify: r.verify.clone(),
            })
            .collect(),
        AcReading::Unreadable(_) => Vec::new(),
    };
    let notes: Vec<BatchNote> = card
        .notes
        .iter()
        .map(|n| BatchNote {
            reference: n.comment_id.to_string(),
            kind: n.kind.as_str(),
            author: n.author.clone(),
            body: n.body.clone(),
        })
        .collect();
    let event_payload = json!({
        "digest": plan.digest,
        "criteria": acs.len(),
        "notes": notes.len(),
        "kind_source": format!("{:?}", card.kind_source).to_lowercase(),
    })
    .to_string();
    BatchCard {
        source_host: card.source_host,
        source_project: card.source_project.clone(),
        source_number: card.source_number,
        title: card.title.clone(),
        kind: card.kind,
        body: card.body.clone(),
        acs,
        notes,
        event_key: format!(
            "import:{}:{}:{}:{}",
            card.source_host,
            card.source_project,
            card.source_number,
            card_content_hash(card)
        ),
        event_payload,
    }
}

/// What [`apply_import`] would hand the ledger for `plan`.
pub fn build_batch(plan: &ImportPlan) -> ImportBatch {
    ImportBatch {
        digest: plan.digest.clone(),
        cards: plan.cards.iter().map(|c| batch_card(plan, c)).collect(),
    }
}

/// Reads `project`'s open issues again and, when they give `expected_digest`, writes the import
/// through `ledger` in one call. A different digest, an unreadable-to-the-end listing or a host
/// that is not the project's writes nothing. With no card to write the ledger is not called.
pub fn apply_import(
    connector: &dyn HostConnector,
    ledger: &mut dyn ImportLedger,
    project: &ProjectRef,
    expected_digest: &str,
) -> Result<ImportReport, ApplyError> {
    if connector.host() != project.host() {
        return Err(ApplyError::Host(ConnectorError::Invalid(
            "the connector is for another host than the project".into(),
        )));
    }
    let issues = connector.open_issues(project).map_err(ApplyError::Host)?;
    let plan = plan_import(&issues);
    if plan.digest != expected_digest {
        return Err(ApplyError::HostChanged {
            expected: expected_digest.to_string(),
            found: plan.digest,
        });
    }
    let batch = build_batch(&plan);
    let mut report = ImportReport {
        digest: plan.digest.clone(),
        created: Vec::new(),
        already_there: Vec::new(),
        duplicate_numbers: plan.duplicate_numbers.clone(),
        created_without_criteria: Vec::new(),
        notes_written: BTreeMap::new(),
        pull_requests_skipped: plan.pull_requests_skipped,
    };
    if batch.cards.is_empty() {
        return Ok(report);
    }
    let commit = ledger.commit(&batch).map_err(ApplyError::Ledger)?;
    if commit.outcomes.len() != batch.cards.len() {
        return Err(ApplyError::LedgerAnswer {
            cards: batch.cards.len(),
            outcomes: commit.outcomes.len(),
        });
    }
    for (card, outcome) in batch.cards.iter().zip(&commit.outcomes) {
        match outcome {
            CardOutcome::AlreadyThere => report.already_there.push(card.source_number),
            CardOutcome::Created => {
                report.created.push(card.source_number);
                if card.acs.is_empty() {
                    report.created_without_criteria.push(card.source_number);
                }
                for note in &card.notes {
                    *report.notes_written.entry(note.kind).or_default() += 1;
                }
            }
        }
    }
    Ok(report)
}
