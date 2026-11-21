//! The work on a card after it is taken (card C-53 S1): which card comes next, notes, the
//! state of an acceptance criterion, the pull request link, `active -> review`, `review -> done`
//! and the write-boundary predicate. The reservation rules stay in [`super::reservation`]; this
//! module is the SQL of the finish path and nothing else.
//!
//! - [`next`] only reads: the cards a holder holds, the `ready` cards it could take (priority,
//!   then number) and every card it cannot take, with the reason.
//! - [`add_note`] appends one note. Notes are not fenced: a decision is answered by someone who
//!   does not hold the card. Every note gets a `ref` that is unique on its card, and its event
//!   records the session that wrote it next to the author it is signed with.
//! - [`add_criteria_as`] (and [`add_criteria`], the call without a session) appends criteria to a
//!   card that exists (card C-108). It is not fenced either: a criterion changes nothing a lane may
//!   write, and a card nobody holds (a migrated Issue card) must be able to get one. A `done` card
//!   is refused. Its event records the session that added the criteria next to the author it is
//!   signed with (`null` without one, card C-125), and the total it answers is the card's row
//!   count.
//! - [`reopen`] is `review -> active` for the holder, the way back after a failed review.
//! - [`set_criterion`], [`link_pr`], [`review`] and [`done`] are fenced writers: they accept only
//!   the current generation of the offered holder, and a refusal is recorded as a
//!   `fence_reject` event like every other fenced call. [`set_criterion`] and [`review`] need a
//!   live lease; [`link_pr`] and [`done`] do not, like `release`: a pull request can be merged
//!   after the lease ended, and nothing reclaims a card yet.
//! - [`done`] checks what the schema's `card_done_guard` checks, in the same order and with the
//!   same sentences, before it writes; the trigger stays the last word.
//! - [`finish`] closes a `ready`, `active` or `review` card in one write transaction that is not
//!   fenced: every check comes before its first write, so a refusal leaves every table as it was
//!   (a fenced refusal commits a `fence_reject` event). [`finish_gate`] is its reservation rule
//!   as a pure function; [`landing_facts`] reads what the ledger records about landings.
//!
//! The schema is not changed here. The state of a pull request is a value the caller passes in:
//! nothing in this module reads a code host.

use super::reservation::{self, FenceError, Reservation};
use super::{LedgerError, WriteLedger};
use rusqlite::{params, Connection, OptionalExtension};
use serde_json::json;

/// The kinds a note can have (the schema's list).
pub const NOTE_KINDS: [&str; 4] = ["decision", "report", "observation", "comment"];
/// Largest note body, in bytes (1 MiB, the limit of a card body).
pub const MAX_NOTE_BYTES: usize = 1_048_576;
/// Longest note `ref`, evidence reference or withdrawal reason, in characters.
pub const MAX_LINE_CHARS: usize = 500;
/// Longest note author, in characters.
pub const MAX_AUTHOR_CHARS: usize = 200;
/// Longest report path [`finish`] stores (the resolved form of a path typed in at most
/// [`MAX_LINE_CHARS`] characters), in characters.
pub const MAX_STORED_PATH_CHARS: usize = 4096;

/// Why a call of this module did not do what was asked. [`WorkflowError::code`] is the stable
/// word; `Display` keeps the prefixes `contract::classify` reads.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum WorkflowError {
    /// A value is out of its limits or malformed. Nothing was read or written.
    InvalidInput(String),
    CardNotFound(i64),
    /// The caller is not the current holder at the current generation (or the lease ended).
    Fence(FenceError),
    /// The card has no criterion with this number.
    CriterionNotFound {
        card_id: i64,
        seq: i64,
    },
    /// The card already has a note with this `ref`.
    DuplicateRef {
        card_id: i64,
        reference: String,
    },
    /// The card is not in the state the action starts from.
    WrongState {
        card_id: i64,
        state: String,
        needs: &'static str,
        action: &'static str,
    },
    /// `done` was refused by one of the three closing rules; the text is the schema's own.
    NotClosable(String),
    /// The card is `done`: a closed card keeps the criteria it had.
    Closed {
        card_id: i64,
    },
    /// The card's link to this pull request was retired; it is not made live again.
    LinkRetired {
        card_id: i64,
        target: String,
    },
    /// `finish` was refused by one of its own rules (criteria, evidence); the text names it.
    FinishRefused(String),
    /// `finish` found another session's reservation whose lease has not ended.
    FinishHeld {
        card_id: i64,
        holder: String,
        generation: i64,
        lease_until: i64,
        now: i64,
    },
    /// The ledger itself failed (busy, unavailable, constraint ...).
    Ledger(LedgerError),
}

impl WorkflowError {
    pub fn code(&self) -> &'static str {
        match self {
            WorkflowError::InvalidInput(_) => "invalid_input",
            WorkflowError::CardNotFound(_) => "card_not_found",
            WorkflowError::Fence(e) => e.code(),
            WorkflowError::CriterionNotFound { .. } => "criterion_not_found",
            WorkflowError::DuplicateRef { .. } => "duplicate_ref",
            WorkflowError::WrongState { .. } => "wrong_state",
            WorkflowError::NotClosable(_) => "not_closable",
            WorkflowError::Closed { .. } => "closed",
            WorkflowError::LinkRetired { .. } => "link_retired",
            WorkflowError::FinishRefused(_) => "finish_refused",
            WorkflowError::FinishHeld { .. } => "finish_held",
            WorkflowError::Ledger(e) => e.code(),
        }
    }
}

impl std::fmt::Display for WorkflowError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            WorkflowError::InvalidInput(why) => write!(f, "invalid card input: {why}"),
            WorkflowError::CardNotFound(id) => write!(f, "card not found: C-{id}"),
            WorkflowError::Fence(e) => e.fmt(f),
            WorkflowError::CriterionNotFound { card_id, seq } => {
                write!(f, "criterion not found: C-{card_id} has no criterion {seq}")
            }
            WorkflowError::DuplicateRef { card_id, reference } => write!(
                f,
                "card conflict in note ref: C-{card_id} already has a note with ref {reference:?}"
            ),
            WorkflowError::WrongState {
                card_id,
                state,
                needs,
                action,
            } => write!(
                f,
                "card conflict in state: C-{card_id} is `{state}`; only a `{needs}` card can be {action}"
            ),
            WorkflowError::NotClosable(why) => write!(f, "card done refused: {why}"),
            WorkflowError::Closed { card_id } => write!(
                f,
                "card conflict in state: C-{card_id} is `done`; a closed card keeps the criteria it had"
            ),
            WorkflowError::LinkRetired { card_id, target } => write!(
                f,
                "card conflict in pr link: C-{card_id}'s link to {target} is retired; a retired link is not linked again"
            ),
            WorkflowError::FinishRefused(why) => write!(f, "card finish refused: {why}"),
            WorkflowError::FinishHeld {
                card_id,
                holder,
                generation,
                lease_until,
                now,
            } => write!(
                f,
                "reservation conflict in card: C-{card_id} is held by {holder} at generation {generation} until {lease_until} (now {now}); finish needs your own or an ended lease"
            ),
            WorkflowError::Ledger(e) => e.fmt(f),
        }
    }
}

impl std::error::Error for WorkflowError {}

impl From<LedgerError> for WorkflowError {
    fn from(e: LedgerError) -> Self {
        WorkflowError::Ledger(e)
    }
}

impl From<FenceError> for WorkflowError {
    fn from(e: FenceError) -> Self {
        match e {
            FenceError::Ledger(inner) => WorkflowError::Ledger(inner),
            FenceError::CardNotFound(id) => WorkflowError::CardNotFound(id),
            other => WorkflowError::Fence(other),
        }
    }
}

// ---- next --------------------------------------------------------------------------------------

/// A card the holder holds now.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Mine {
    pub reservation: Reservation,
    pub title: String,
    pub state: String,
    pub priority: String,
    /// Whether the lease has not ended (DB clock).
    pub lease_live: bool,
}

/// A `ready` card the holder could take now.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Candidate {
    pub card_id: i64,
    pub title: String,
    pub kind: String,
    pub priority: String,
    pub write_scope: Vec<String>,
}

/// Why a card is not offered.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SkipReason {
    /// Another holder has a reservation on it.
    Held {
        holder: String,
        generation: i64,
        lease_until: i64,
    },
    /// Its write scope could be written under another card's active reservation.
    ScopeOverlap {
        other_card: i64,
        mine: String,
        theirs: String,
    },
    /// It has no write scope, and a reservation needs one.
    EmptyScope,
    /// It is in `review` and nobody holds it, so nobody can close or reopen it.
    ReviewNoHolder,
}

impl SkipReason {
    pub fn code(&self) -> &'static str {
        match self {
            SkipReason::Held { .. } => "held",
            SkipReason::ScopeOverlap { .. } => "scope_overlap",
            SkipReason::EmptyScope => "empty_scope",
            SkipReason::ReviewNoHolder => "review_no_holder",
        }
    }
}

/// A card that is not offered, with the reason.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Skipped {
    pub card_id: i64,
    pub title: String,
    pub state: String,
    pub priority: String,
    pub reason: SkipReason,
}

/// What [`next`] found.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Next {
    /// The DB clock of the read.
    pub now: i64,
    pub mine: Vec<Mine>,
    pub ready: Vec<Candidate>,
    pub skipped: Vec<Skipped>,
}

/// The cards `holder` holds, then the `ready` cards it could take (priority, then number), then
/// the cards it cannot take. Read-only; `holder` is `None` for a caller without a session id,
/// who holds nothing.
pub fn next(conn: &Connection, holder: Option<&str>) -> Result<Next, LedgerError> {
    let now: i64 = conn.query_row("SELECT unixepoch()", [], |r| r.get(0))?;
    let mut stmt = conn.prepare(
        "SELECT card_id, title, kind, state, priority, write_scope FROM card \
         WHERE state <> 'done' ORDER BY priority, card_id",
    )?;
    let cards: Vec<(i64, String, String, String, String, String)> = stmt
        .query_map([], |r| {
            Ok((
                r.get(0)?,
                r.get(1)?,
                r.get(2)?,
                r.get(3)?,
                r.get(4)?,
                r.get(5)?,
            ))
        })?
        .collect::<Result<_, _>>()?;
    let mut out = Next {
        now,
        mine: Vec::new(),
        ready: Vec::new(),
        skipped: Vec::new(),
    };
    for (card_id, title, kind, state, priority, write_scope) in cards {
        if let Some(held) = reservation::current(conn, card_id)? {
            if holder == Some(held.holder.as_str()) {
                out.mine.push(Mine {
                    lease_live: held.lease_until > now,
                    reservation: held,
                    title,
                    state,
                    priority,
                });
            } else {
                out.skipped.push(Skipped {
                    card_id,
                    title,
                    state,
                    priority,
                    reason: SkipReason::Held {
                        holder: held.holder,
                        generation: held.generation,
                        lease_until: held.lease_until,
                    },
                });
            }
            continue;
        }
        if state == "review" {
            // Nobody can close or reopen it: `review`, `done` and `reopen` all need the holder.
            out.skipped.push(Skipped {
                card_id,
                title,
                state,
                priority,
                reason: SkipReason::ReviewNoHolder,
            });
            continue;
        }
        if state != "ready" {
            continue;
        }
        // A scope that is not a list of strings is no scope: `take` could not use it either.
        let scope: Vec<String> = serde_json::from_str(&write_scope).unwrap_or_default();
        let reason = if scope.is_empty() {
            Some(SkipReason::EmptyScope)
        } else {
            reservation::find_overlap(conn, now, card_id, &scope)?.map(|o| {
                SkipReason::ScopeOverlap {
                    other_card: o.other_card,
                    mine: o.mine,
                    theirs: o.theirs,
                }
            })
        };
        match reason {
            Some(reason) => out.skipped.push(Skipped {
                card_id,
                title,
                state,
                priority,
                reason,
            }),
            None => out.ready.push(Candidate {
                card_id,
                title,
                kind,
                priority,
                write_scope: scope,
            }),
        }
    }
    Ok(out)
}

// ---- notes -------------------------------------------------------------------------------------

/// One stored note.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Note {
    pub note_id: i64,
    pub card_id: i64,
    pub kind: String,
    pub author: String,
    pub reference: String,
    pub created_at: i64,
    pub bytes: usize,
}

/// Appends one note to the card, with a `card.note` event. `reference` names the note on its
/// card; without one a random one is made (`note:<card>:<32 hex digits>`). `author` is the name
/// the note is signed with and `session` the id of the session that writes it (`None` for a
/// caller without one): the event keeps both, so a note signed `leader` still says which
/// session wrote it.
pub fn add_note(
    w: &mut WriteLedger,
    card_id: i64,
    kind: &str,
    author: &str,
    session: Option<&str>,
    body: &str,
    reference: Option<&str>,
) -> Result<Note, WorkflowError> {
    let invalid = WorkflowError::InvalidInput;
    if !NOTE_KINDS.contains(&kind) {
        return Err(invalid(format!(
            "the note kind is one of {}",
            NOTE_KINDS.join(", ")
        )));
    }
    one_line("the note author", author, MAX_AUTHOR_CHARS).map_err(invalid)?;
    if let Some(session) = session {
        one_line("the session id", session, MAX_AUTHOR_CHARS).map_err(invalid)?;
    }
    if body.trim().is_empty() {
        return Err(invalid("the note text is empty".into()));
    }
    if body.contains('\0') {
        return Err(invalid("the note text contains a NUL character".into()));
    }
    // C-68 AC-3: Windows line endings become line feeds, so a note written on one system is
    // stored one way; a carriage return that is not part of a line ending is judged below.
    let body = &body.replace("\r\n", "\n");
    // A note is shown on terminals and written out by `card export`: an escape sequence or a
    // carriage return could repaint what a reader sees. Tabs and line breaks are text.
    if body.contains('\r') {
        return Err(invalid(
            "the note text contains a control character: a carriage return outside a line \
             ending; normalise Windows line endings"
                .into(),
        ));
    }
    if body
        .chars()
        .any(|c| c.is_control() && c != '\t' && c != '\n')
    {
        return Err(invalid(
            "the note text contains a control character; only tabs and line breaks are kept".into(),
        ));
    }
    if body.chars().any(is_directional) {
        return Err(invalid(
            "the note text contains a directional or formatting character \
             (U+2028–U+202E, U+2066–U+2069); take it out"
                .into(),
        ));
    }
    if body.len() > MAX_NOTE_BYTES {
        return Err(invalid(format!(
            "the note text is {} bytes, at most {MAX_NOTE_BYTES}",
            body.len()
        )));
    }
    if let Some(r) = reference {
        one_line("the note ref", r, MAX_LINE_CHARS).map_err(invalid)?;
    }
    w.with_write_tx(|tx, now| {
        if !card_exists(tx, card_id)? {
            return Ok(Err(WorkflowError::CardNotFound(card_id)));
        }
        let reference: String = match reference {
            Some(r) => {
                let taken: i64 = tx.query_row(
                    "SELECT count(*) FROM card_note WHERE card_id = ?1 AND ref = ?2",
                    params![card_id, r],
                    |row| row.get(0),
                )?;
                if taken > 0 {
                    return Ok(Err(WorkflowError::DuplicateRef {
                        card_id,
                        reference: r.to_string(),
                    }));
                }
                r.to_string()
            }
            None => tx.query_row(
                "SELECT 'note:' || ?1 || ':' || lower(hex(randomblob(16)))",
                [card_id],
                |row| row.get(0),
            )?,
        };
        tx.execute(
            "INSERT INTO card_note (card_id, kind, author, body, ref, created_at) \
             VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
            params![card_id, kind, author, body, reference, now],
        )?;
        let note_id = tx.last_insert_rowid();
        reservation::add_event(
            tx,
            now,
            author,
            card_id,
            "card.note",
            // `session` is who wrote the note, whatever `author` claims; `null` without one.
            &json!({
                "note_id": note_id, "kind": kind, "ref": reference, "bytes": body.len(),
                "author": author, "session": session,
            }),
        )?;
        Ok(Ok(Note {
            note_id,
            card_id,
            kind: kind.to_string(),
            author: author.to_string(),
            reference,
            created_at: now,
            bytes: body.len(),
        }))
    })?
}

// ---- criteria ----------------------------------------------------------------------------------

/// What to do with a criterion.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CriterionChange<'a> {
    Met { evidence: &'a str },
    Withdrawn { reason: &'a str },
}

/// A criterion after [`set_criterion`].
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Criterion {
    pub card_id: i64,
    pub seq: i64,
    /// The status before the call.
    pub from: String,
    pub status: String,
    pub evidence_ref: Option<String>,
    pub withdrawn_reason: Option<String>,
}

/// Marks criterion `seq` of the card met (with its evidence) or withdrawn (with the reason), as
/// the current holder at `generation` with a live lease.
pub fn set_criterion(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    seq: i64,
    change: CriterionChange<'_>,
) -> Result<Criterion, WorkflowError> {
    if seq < 1 {
        return Err(WorkflowError::InvalidInput(
            "a criterion number is 1 or more".into(),
        ));
    }
    match change {
        CriterionChange::Met { evidence } => one_line("the evidence", evidence, MAX_LINE_CHARS),
        CriterionChange::Withdrawn { reason } => one_line("the reason", reason, MAX_LINE_CHARS),
    }
    .map_err(WorkflowError::InvalidInput)?;
    reservation::fenced(
        w,
        "criterion",
        card_id,
        holder,
        generation,
        true,
        |tx, now, _| {
            let Some(from) = tx
                .query_row(
                    "SELECT status FROM card_ac WHERE card_id = ?1 AND seq = ?2",
                    params![card_id, seq],
                    |r| r.get::<_, String>(0),
                )
                .optional()?
            else {
                return Ok(Err(WorkflowError::CriterionNotFound { card_id, seq }));
            };
            // A met criterion drops an old withdrawal reason; a withdrawn one keeps the
            // evidence it had, so the history of the claim stays readable.
            let status = match change {
                CriterionChange::Met { evidence } => {
                    tx.execute(
                        "UPDATE card_ac SET status = 'met', evidence_ref = ?1, \
                         withdrawn_reason = NULL, updated_at = ?2 WHERE card_id = ?3 AND seq = ?4",
                        params![evidence, now, card_id, seq],
                    )?;
                    "met"
                }
                CriterionChange::Withdrawn { reason } => {
                    tx.execute(
                        "UPDATE card_ac SET status = 'withdrawn', withdrawn_reason = ?1, \
                         updated_at = ?2 WHERE card_id = ?3 AND seq = ?4",
                        params![reason, now, card_id, seq],
                    )?;
                    "withdrawn"
                }
            };
            let (evidence_ref, withdrawn_reason): (Option<String>, Option<String>) = tx.query_row(
                "SELECT evidence_ref, withdrawn_reason FROM card_ac \
                     WHERE card_id = ?1 AND seq = ?2",
                params![card_id, seq],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )?;
            reservation::add_event(
                tx,
                now,
                holder,
                card_id,
                "card.criterion",
                &json!({
                    "seq": seq, "from": from, "status": status, "generation": generation,
                    "evidence_ref": evidence_ref, "withdrawn_reason": withdrawn_reason,
                }),
            )?;
            Ok(Ok(Criterion {
                card_id,
                seq,
                from,
                status: status.to_string(),
                evidence_ref,
                withdrawn_reason,
            }))
        },
    )?
}

/// What [`add_criteria`] stored.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct CriteriaAdded {
    pub card_id: i64,
    /// The numbers the new criteria got, in the order given.
    pub added: Vec<i64>,
    /// How many criteria the card has now (withdrawn ones included).
    pub total: i64,
    pub rev: i64,
}

/// Appends `texts` to the card as `open` criteria, numbered after the card's highest criterion
/// number (a withdrawn criterion keeps its number), and writes one `card.criterion_add` event
/// with who added them, the first number, the count and `session` as `null` (this call has no
/// session id; [`add_criteria_as`] records one). Nobody has to hold the card: the
/// criteria change nothing a lane may write, so there is no fence. A `done` card is refused.
/// The texts are one line each, checked by the caller (`card::input`); an empty list is refused.
/// The call without a session (card C-125): [`add_criteria_as`] with `None`.
pub fn add_criteria(
    w: &mut WriteLedger,
    card_id: i64,
    by: &str,
    texts: &[String],
) -> Result<CriteriaAdded, WorkflowError> {
    add_criteria_as(w, card_id, by, None, texts)
}

/// [`add_criteria`] that also records the session that adds the criteria (card C-125, C-108 audit
/// N1): the event payload carries `session` next to `first` and `count`, `null` when the caller
/// has no session id, as `card.note` records its session. `by` stays the author the event is signed
/// with (free text); `session` is the id the host gave, one line within the author limit, checked
/// before anything is written. `total` is the number of criterion rows the card has after the add
/// (card C-125, audit N3), not its highest number.
pub fn add_criteria_as(
    w: &mut WriteLedger,
    card_id: i64,
    by: &str,
    session: Option<&str>,
    texts: &[String],
) -> Result<CriteriaAdded, WorkflowError> {
    one_line("the author", by, MAX_AUTHOR_CHARS).map_err(WorkflowError::InvalidInput)?;
    if let Some(session) = session {
        one_line("the session id", session, MAX_AUTHOR_CHARS)
            .map_err(WorkflowError::InvalidInput)?;
    }
    if texts.is_empty() {
        return Err(WorkflowError::InvalidInput(
            "there is no criterion to add".into(),
        ));
    }
    w.with_write_tx(|tx, now| {
        if !card_exists(tx, card_id)? {
            return Ok(Err(WorkflowError::CardNotFound(card_id)));
        }
        if card_state(tx, card_id)? == "done" {
            return Ok(Err(WorkflowError::Closed { card_id }));
        }
        // The highest number over every row, withdrawn included: rows are never deleted, so a
        // number is never given twice.
        let highest: i64 = tx.query_row(
            "SELECT COALESCE(MAX(seq), 0) FROM card_ac WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )?;
        let mut added = Vec::with_capacity(texts.len());
        for (offset, text) in texts.iter().enumerate() {
            let seq = highest + 1 + offset as i64;
            tx.execute(
                "INSERT INTO card_ac (card_id, seq, text, updated_at) VALUES (?1, ?2, ?3, ?4)",
                params![card_id, seq, text, now],
            )?;
            added.push(seq);
        }
        reservation::add_event(
            tx,
            now,
            by,
            card_id,
            "card.criterion_add",
            &json!({"first": added[0], "count": added.len(), "session": session}),
        )?;
        // The rows the card has now, read after the inserts in this transaction: the highest
        // number says how many rows there are only while the numbers have no gap.
        let total: i64 = tx.query_row(
            "SELECT COUNT(*) FROM card_ac WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )?;
        Ok(Ok(CriteriaAdded {
            card_id,
            total,
            added,
            rev: card_rev(tx, card_id)?,
        }))
    })?
}

// ---- withdraw ----------------------------------------------------------------------------------

/// Why a card withdraw did not happen. Its own type: nothing outside this module matches it
/// exhaustively, so the variants can grow without touching the callers' matches.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum WithdrawError {
    CardNotFound(i64),
    /// The card is past the withdrawable states (`draft`, `ready`).
    WrongState {
        card_id: i64,
        state: String,
    },
    /// A live reservation holds the card.
    Held {
        card_id: i64,
        holder: String,
        generation: i64,
    },
    InvalidInput(String),
    Ledger(LedgerError),
}

impl std::fmt::Display for WithdrawError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            WithdrawError::CardNotFound(id) => write!(f, "card not found: C-{id}"),
            WithdrawError::WrongState { card_id, state } => write!(
                f,
                "the card C-{card_id} is `{state}`, only a `draft` or `ready` card can be withdrawn"
            ),
            WithdrawError::Held {
                card_id,
                holder,
                generation,
            } => write!(
                f,
                "the card C-{card_id} is held by {holder:?} at generation {generation}"
            ),
            WithdrawError::InvalidInput(why) => write!(f, "{why}"),
            WithdrawError::Ledger(e) => write!(f, "{e}"),
        }
    }
}

impl std::error::Error for WithdrawError {}

impl From<LedgerError> for WithdrawError {
    fn from(e: LedgerError) -> Self {
        WithdrawError::Ledger(e)
    }
}

/// Withdraws a `draft` or `ready` card that nobody holds: the card moves to `hold` and one
/// `card.withdraw` event records the reason and the by, with the session that wrote it
/// (`None` without one) as the event's session. There is no `withdrawn` state value: the
/// schema CHECK pins the six values the card table accepts, so a withdrawn card reuses
/// `hold` and the marker is the event — `card take` refuses it naming the reason, and the
/// default `card list` hides it. Nothing is deleted; input bounds are checked before
/// anything is read.
pub fn withdraw(
    w: &mut WriteLedger,
    card_id: i64,
    reason: &str,
    by: &str,
    session: Option<&str>,
) -> Result<(), WithdrawError> {
    one_line("the reason", reason, MAX_LINE_CHARS).map_err(WithdrawError::InvalidInput)?;
    one_line("the by", by, MAX_AUTHOR_CHARS).map_err(WithdrawError::InvalidInput)?;
    if let Some(session) = session {
        one_line("the session id", session, MAX_AUTHOR_CHARS)
            .map_err(WithdrawError::InvalidInput)?;
    }
    w.with_write_tx(|tx, now| {
        let state = match tx
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [card_id],
                |r| r.get(0),
            )
            .optional()?
        {
            Some(state) => state,
            None => return Ok(Err(WithdrawError::CardNotFound(card_id))),
        };
        if let Some(active) = reservation::current(tx, card_id)? {
            return Ok(Err(WithdrawError::Held {
                card_id,
                holder: active.holder,
                generation: active.generation,
            }));
        }
        if state != "draft" && state != "ready" {
            return Ok(Err(WithdrawError::WrongState { card_id, state }));
        }
        tx.execute(
            "UPDATE card SET state = 'hold', updated_at = ?2 WHERE card_id = ?1",
            params![card_id, now],
        )?;
        reservation::add_event(
            tx,
            now,
            session.unwrap_or("user"),
            card_id,
            "card.withdraw",
            &json!({"reason": reason, "by": by}),
        )?;
        Ok(Ok(()))
    })
    .map_err(WithdrawError::from)?
}

// ---- pull request link -------------------------------------------------------------------------

/// The pull request a card is linked to, as the caller read it from the host.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct PrTarget<'a> {
    /// `github` or `gitlab`.
    pub host: &'a str,
    /// Lower case, as `card.source_project` is stored.
    pub project: &'a str,
    pub number: i64,
    /// `open`, `merged` or `closed`.
    pub pr_state: &'a str,
    /// The merge commit; required when `pr_state` is `merged`.
    pub merge_ref: Option<&'a str>,
}

/// A `pr` link after [`link_pr`].
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PrLink {
    pub link_id: i64,
    pub card_id: i64,
    pub host: String,
    pub project: String,
    pub number: i64,
    pub pr_state: String,
    pub merge_ref: Option<String>,
    /// Whether the link is new (`false`: an existing link got the new state).
    pub created: bool,
}

/// The fence check alone, for a caller that must ask a code host between the check and the
/// write: `Ok` only for the current holder at `generation`. A refusal is recorded as
/// `fence_reject` with `op`; a pass writes nothing.
pub fn holder_check(
    w: &mut WriteLedger,
    op: &str,
    card_id: i64,
    holder: &str,
    generation: i64,
    need_live: bool,
) -> Result<(), WorkflowError> {
    Ok(reservation::fenced(
        w,
        op,
        card_id,
        holder,
        generation,
        need_live,
        |_, _, _| Ok(()),
    )?)
}

/// Links the card to a pull request, or stores the new state of a link it already has, as the
/// current holder at `generation` (a live lease is not needed). A link that was retired stays
/// retired: the call is refused and nothing changes.
pub fn link_pr(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    target: &PrTarget<'_>,
) -> Result<PrLink, WorkflowError> {
    let invalid = |why: &str| Err(WorkflowError::InvalidInput(why.to_string()));
    let t = *target;
    if !["github", "gitlab"].contains(&t.host) {
        return invalid("the host is `github` or `gitlab`");
    }
    if t.project.is_empty()
        || t.project != t.project.to_lowercase()
        || t.project.chars().any(|c| c == ' ' || c.is_control())
        || t.project.chars().count() > MAX_LINE_CHARS
    {
        return invalid("the host project is a lower-case name without spaces");
    }
    if t.number < 1 {
        return invalid("a pull request number is 1 or more");
    }
    if !["open", "merged", "closed"].contains(&t.pr_state) {
        return invalid("the pull request state is `open`, `merged` or `closed`");
    }
    if let Some(merge_ref) = t.merge_ref {
        one_line("the merge commit", merge_ref, MAX_LINE_CHARS)
            .map_err(WorkflowError::InvalidInput)?;
    }
    if t.pr_state == "merged" && t.merge_ref.is_none() {
        return invalid("a merged pull request needs its merge commit");
    }
    // Only a merged pull request has a merge commit worth keeping.
    let merge_ref = t.merge_ref.filter(|_| t.pr_state == "merged");
    reservation::fenced(w, "pr", card_id, holder, generation, false, |tx, now, _| {
        let existing: Option<(i64, String)> = tx
            .query_row(
                "SELECT link_id, state FROM card_link WHERE from_card = ?1 AND kind = 'pr' \
                     AND target_host = ?2 AND target_project = ?3 AND target_number = ?4",
                params![card_id, t.host, t.project, t.number],
                |r| Ok((r.get(0)?, r.get(1)?)),
            )
            .optional()?;
        let link_id = match &existing {
            // Retiring a link is a decision; reading the host again does not undo it.
            Some((_, state)) if state != "active" => {
                return Ok(Err(WorkflowError::LinkRetired {
                    card_id,
                    target: format!("{}:{}#{}", t.host, t.project, t.number),
                }));
            }
            Some((link_id, _)) => {
                tx.execute(
                    "UPDATE card_link SET pr_state = ?1, merge_ref = ?2, updated_at = ?3 \
                     WHERE link_id = ?4",
                    params![t.pr_state, merge_ref, now, link_id],
                )?;
                *link_id
            }
            None => {
                tx.execute(
                    "INSERT INTO card_link (from_card, kind, target_host, target_project, \
                         target_number, pr_state, merge_ref, created_at, updated_at) \
                         VALUES (?1, 'pr', ?2, ?3, ?4, ?5, ?6, ?7, ?7)",
                    params![card_id, t.host, t.project, t.number, t.pr_state, merge_ref, now],
                )?;
                tx.last_insert_rowid()
            }
        };
        reservation::add_event(
            tx,
            now,
            holder,
            card_id,
            "card.pr",
            &json!({
                "link_id": link_id, "host": t.host, "project": t.project,
                "number": t.number, "pr_state": t.pr_state, "merge_ref": merge_ref,
                "generation": generation, "created": existing.is_none(),
            }),
        )?;
        Ok(Ok(PrLink {
            link_id,
            card_id,
            host: t.host.to_string(),
            project: t.project.to_string(),
            number: t.number,
            pr_state: t.pr_state.to_string(),
            merge_ref: merge_ref.map(str::to_string),
            created: existing.is_none(),
        }))
    })?
}

// ---- review and done ---------------------------------------------------------------------------

/// A card after [`review`], [`reopen`] or [`done`].
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Moved {
    pub card_id: i64,
    pub from: String,
    pub state: String,
    pub rev: i64,
    /// Set by [`done`].
    pub closed_at: Option<i64>,
    /// The holder's reservation after the call: still `held` after [`review`], `released`
    /// after [`done`].
    pub reservation: Reservation,
}

/// `active -> review`, as the current holder at `generation` with a live lease. The reservation
/// stays `held`.
pub fn review(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
) -> Result<Moved, WorkflowError> {
    reservation::fenced(
        w,
        "review",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| {
            let from = card_state(tx, card_id)?;
            if from != "active" {
                return Ok(Err(WorkflowError::WrongState {
                    card_id,
                    state: from,
                    needs: "active",
                    action: "sent to review",
                }));
            }
            tx.execute(
                "UPDATE card SET state = 'review', updated_at = ?2 WHERE card_id = ?1",
                params![card_id, now],
            )?;
            reservation::add_event(
                tx,
                now,
                holder,
                card_id,
                "card.review",
                &json!({"from": from, "generation": generation}),
            )?;
            Ok(Ok(Moved {
                card_id,
                from,
                state: "review".to_string(),
                rev: card_rev(tx, card_id)?,
                closed_at: None,
                reservation: row,
            }))
        },
    )?
}

/// `review -> active`, as the current holder at `generation` with a live lease: the way back
/// after a review that found something to repair. The reservation stays `held`.
pub fn reopen(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
) -> Result<Moved, WorkflowError> {
    reservation::fenced(
        w,
        "reopen",
        card_id,
        holder,
        generation,
        true,
        |tx, now, row| {
            let from = card_state(tx, card_id)?;
            if from != "review" {
                return Ok(Err(WorkflowError::WrongState {
                    card_id,
                    state: from,
                    needs: "review",
                    action: "reopened",
                }));
            }
            tx.execute(
                "UPDATE card SET state = 'active', updated_at = ?2 WHERE card_id = ?1",
                params![card_id, now],
            )?;
            reservation::add_event(
                tx,
                now,
                holder,
                card_id,
                "card.reopen",
                &json!({"from": from, "generation": generation}),
            )?;
            Ok(Ok(Moved {
                card_id,
                from,
                state: "active".to_string(),
                rev: card_rev(tx, card_id)?,
                closed_at: None,
                reservation: row,
            }))
        },
    )?
}

/// The local landing evidence a M-Flow project closes on (SPEC-20261120-card-done-mflow):
/// the landed head and the integration branch tip that contains it, checked at done time.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Landing {
    pub head: String,
    pub tip: String,
}

/// `card landed`: record that this card's head landed on the local integration branch
/// (M-Flow). The card must be in `review`; one free-form `card.landed` event carries the head
/// and the by, with the writing session as the actor.
pub fn land(
    w: &mut WriteLedger,
    card_id: i64,
    head: &str,
    by: &str,
    session: Option<&str>,
) -> Result<(), WorkflowError> {
    if !crate::gate::is_full_sha(head) {
        return Err(WorkflowError::InvalidInput(
            "the head must be 40 lowercase hex".to_string(),
        ));
    }
    one_line("the by", by, MAX_AUTHOR_CHARS).map_err(WorkflowError::InvalidInput)?;
    w.with_write_tx(|tx, now| {
        let from: Option<String> = tx
            .query_row(
                "SELECT state FROM card WHERE card_id = ?1",
                [card_id],
                |r| r.get(0),
            )
            .optional()?;
        let from = match from {
            Some(state) => state,
            None => return Ok(Err(WorkflowError::CardNotFound(card_id))),
        };
        if from != "review" {
            return Ok(Err(WorkflowError::WrongState {
                card_id,
                state: from,
                needs: "review",
                action: "landed",
            }));
        }
        reservation::add_event(
            tx,
            now,
            session.unwrap_or("user"),
            card_id,
            "card.landed",
            &json!({"head": head, "by": by}),
        )?;
        Ok(Ok(()))
    })?
}

/// `review -> done`, as the current holder at `generation` (a live lease is not needed). Needs
/// what `card_done_guard` needs: a met criterion, no open criterion and — unless the caller
/// passes local [`Landing`] evidence (M-Flow) — an active `pr` link whose state is `merged`.
/// The reservation ends (`released`) in the same transaction.
pub fn done(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
    landing: Option<Landing>,
) -> Result<Moved, WorkflowError> {
    reservation::fenced(
        w,
        "done",
        card_id,
        holder,
        generation,
        false,
        |tx, now, _| {
            let from = card_state(tx, card_id)?;
            if from != "review" {
                return Ok(Err(WorkflowError::WrongState {
                    card_id,
                    state: from,
                    needs: "review",
                    action: "closed",
                }));
            }
            // The three rules of `card_done_guard`, in its order and with its sentences.
            let count = |sql: &str| -> Result<i64, LedgerError> {
                Ok(tx.query_row(sql, [card_id], |r| r.get(0))?)
            };
            let refusal = if count(
                "SELECT count(*) FROM card_ac WHERE card_id = ?1 AND status = 'met'",
            )? == 0
            {
                Some("cannot close: the card has no acceptance criterion that is met")
            } else if count("SELECT count(*) FROM card_ac WHERE card_id = ?1 AND status = 'open'")?
                > 0
            {
                Some("cannot close: an acceptance criterion is still open")
            } else if landing.is_none()
                && count(
                    "SELECT count(*) FROM card_link WHERE from_card = ?1 AND kind = 'pr' \
                     AND state = 'active' AND pr_state = 'merged'",
                )? == 0
            {
                Some("cannot close: no merged pull request is linked")
            } else {
                None
            };
            if let Some(why) = refusal {
                return Ok(Err(WorkflowError::NotClosable(why.to_string())));
            }
            tx.execute(
                "UPDATE card SET state = 'done', closed_at = ?2, updated_at = ?2 WHERE card_id = ?1",
                params![card_id, now],
            )?;
            tx.execute(
                "UPDATE reservation SET status = 'released', updated_at = ?1 \
                 WHERE card_id = ?2 AND generation = ?3",
                params![now, card_id, generation],
            )?;
            reservation::add_event(
                tx,
                now,
                holder,
                card_id,
                "card.done",
                &json!({
                    "from": from, "generation": generation, "closed_at": now,
                    "mflow": landing.as_ref().map(|l| json!({"head": l.head, "tip": l.tip})),
                }),
            )?;
            let reservation =
                reservation::load_reservation(tx, card_id, generation)?.ok_or_else(|| {
                    LedgerError::Sql("the closed card's reservation cannot be read".into())
                })?;
            Ok(Ok(Moved {
                card_id,
                from,
                state: "done".to_string(),
                rev: card_rev(tx, card_id)?,
                closed_at: Some(now),
                reservation,
            }))
        },
    )?
}

// ---- finish ------------------------------------------------------------------------------------

/// What a finish closes on. The git proof of a landing is made by the caller before the call:
/// the ledger never spawns a command.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum FinishEvidence<'a> {
    /// A landed head the caller proved to be inside the integration branch tip.
    Landing { head: &'a str, tip: &'a str },
    /// The card has a merged pull request link, which is the evidence the schema accepts.
    MergedPr,
    /// A report file the caller checked (research and audit cards).
    Report { path: &'a str },
    /// A report or decision note of the card, by its `ref` (research and audit cards).
    Note { reference: &'a str },
}

/// A card after [`finish`].
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Finished {
    pub card_id: i64,
    /// The state the card had before the call.
    pub from: String,
    /// Whether the call closed the card. `false` for a landing-only call: an open criterion had
    /// no met option, so only the landing was recorded and [`Finished::open_criteria`] names
    /// what is still open.
    pub closed: bool,
    /// The state the card has after the call (`done` when it closed, `from` otherwise).
    pub state: String,
    pub rev: i64,
    /// When the card closed; `None` for a landing-only call.
    pub closed_at: Option<i64>,
    /// The numbers of the criteria that are still open after the call (empty when it closed).
    pub open_criteria: Vec<i64>,
    /// The reservation the card had before the call, when it had one.
    pub previous: Option<Reservation>,
    /// Whether that reservation was marked `released` by the call.
    pub released: bool,
    /// The events the call wrote: the landing event (unless the same head is already recorded
    /// or a merged pull request is the evidence) and, when the card closed, the finish event.
    pub events: i64,
}

/// The reservation rule of [`finish`] as a pure function: `finish` passes a reservation the
/// caller holds and one whose lease has ended at `now` (the lease-end second itself counts as
/// ended, as in [`reservation::lease_ended`]); another session's live lease is refused. A caller
/// without a session id holds nothing, so only an ended lease passes.
pub fn finish_gate(
    held: &Reservation,
    session: Option<&str>,
    now: i64,
) -> Result<(), WorkflowError> {
    if session == Some(held.holder.as_str()) || reservation::lease_ended(held.lease_until, now) {
        Ok(())
    } else {
        Err(WorkflowError::FinishHeld {
            card_id: held.card_id,
            holder: held.holder.clone(),
            generation: held.generation,
            lease_until: held.lease_until,
            now,
        })
    }
}

/// The `--by` of the leader: only it, or the holder of the card, may record a landing that
/// closes nothing (`card finish` with an open criterion and no met option for it).
pub const LEADER_BY: &str = "leader";

/// What the checks of a finish decided: everything the writes need, and nothing written.
#[derive(Debug, Clone)]
struct FinishPlan {
    from: String,
    newest_generation: i64,
    previous: Option<Reservation>,
    uncovered: Vec<i64>,
    landing_only: bool,
    record_landing: bool,
}

/// The checks of [`finish`] that need no ledger: the shape of `by`, of each met option and of the
/// evidence.
fn finish_input_checks(
    by: &str,
    met: &[(i64, String)],
    evidence: &FinishEvidence<'_>,
) -> Result<(), WorkflowError> {
    one_line("the by", by, MAX_AUTHOR_CHARS).map_err(WorkflowError::InvalidInput)?;
    for (n, (seq, text)) in met.iter().enumerate() {
        if *seq < 1 {
            return Err(WorkflowError::InvalidInput(
                "a criterion number is 1 or more".into(),
            ));
        }
        one_line("the evidence", text, MAX_LINE_CHARS).map_err(WorkflowError::InvalidInput)?;
        if met[..n].iter().any(|(earlier, _)| earlier == seq) {
            return Err(WorkflowError::FinishRefused(format!(
                "criterion {seq} is named twice by --met; each open criterion takes one"
            )));
        }
    }
    let screened = match evidence {
        // The path is stored in its resolved form, which a link or a symlinked parent (`/var` on
        // macOS) can make longer than the 500 characters the caller was typed; the caller
        // applies that limit to what was typed.
        FinishEvidence::Report { path: given } => {
            one_line("the report path", given, MAX_STORED_PATH_CHARS)
        }
        FinishEvidence::Note { reference } => one_line("the note ref", reference, MAX_LINE_CHARS),
        FinishEvidence::Landing { .. } | FinishEvidence::MergedPr => Ok(()),
    };
    screened.map_err(WorkflowError::InvalidInput)?;
    Ok(())
}

/// Every ledger check of [`finish`], in one place and with no write: the write transaction runs
/// it first, and [`finish_preflight`] runs it with nothing held, so the two cannot drift.
#[allow(clippy::too_many_arguments)]
fn finish_plan(
    conn: &Connection,
    card_id: i64,
    session: Option<&str>,
    by: &str,
    evidence: &FinishEvidence<'_>,
    met: &[(i64, String)],
    generation: Option<i64>,
    now: i64,
) -> Result<Result<FinishPlan, WorkflowError>, LedgerError> {
    let refused = |why: String| Ok(Err(WorkflowError::FinishRefused(why)));
    if !card_exists(conn, card_id)? {
        return Ok(Err(WorkflowError::CardNotFound(card_id)));
    }
    let from = card_state(conn, card_id)?;
    if !matches!(from.as_str(), "ready" | "active" | "review") {
        return Ok(Err(WorkflowError::WrongState {
            card_id,
            state: from,
            needs: "ready, active or review",
            action: "finished",
        }));
    }
    // The generation of the newest reservation, released and expired ones included (a ready
    // card keeps the number of the reservation it had; a card that never had one is at 0).
    let newest_generation: i64 = conn.query_row(
        "SELECT COALESCE(MAX(generation), 0) FROM reservation WHERE card_id = ?1",
        [card_id],
        |r| r.get(0),
    )?;
    if let Some(given) = generation {
        if given != newest_generation {
            return refused(format!(
                "--generation {given} does not match generation {newest_generation}, the \
                 card's newest reservation (0 when it never had one); nothing was changed"
            ));
        }
    }
    let previous = reservation::current(conn, card_id)?;
    if let Some(held) = &previous {
        // A lane holds its review card with a live lease right after it reported the merge
        // ready, and that is when the leader closes it: a landing git proved passes for the
        // leader word or the holder session, and the holder's reservation is released below
        // like any other.
        let owner = by == LEADER_BY || session == Some(held.holder.as_str());
        let proven_review =
            from == "review" && owner && matches!(evidence, FinishEvidence::Landing { .. });
        if !proven_review {
            if let Err(e) = finish_gate(held, session, now) {
                return Ok(Err(e));
            }
        }
    }
    // Every open criterion is named by exactly one met option, and only open ones are.
    let open: Vec<i64> = {
        let mut stmt = conn.prepare(
            "SELECT seq FROM card_ac WHERE card_id = ?1 AND status = 'open' ORDER BY seq",
        )?;
        let rows = stmt.query_map([card_id], |r| r.get(0))?;
        rows.collect::<Result<_, _>>()?
    };
    if let Some((seq, _)) = met.iter().find(|(seq, _)| !open.contains(seq)) {
        return refused(format!(
            "criterion {seq} is not an open criterion of C-{card_id}; --met names open criteria only"
        ));
    }
    // An open criterion with no met option: a card that closes on a landing records the
    // landing and says what is open; a report or a verdict closes the card or nothing.
    let uncovered: Vec<i64> = open
        .iter()
        .copied()
        .filter(|seq| !met.iter().any(|(named, _)| named == seq))
        .collect();
    let landing_only = !uncovered.is_empty()
        && matches!(
            evidence,
            FinishEvidence::Landing { .. } | FinishEvidence::MergedPr
        );
    if landing_only {
        let holder = previous
            .as_ref()
            .is_some_and(|held| session == Some(held.holder.as_str()));
        if by != LEADER_BY && !holder {
            return refused(format!(
                "a call that records a landing and closes nothing is for the leader or the \
                 holder of the card: give --by {LEADER_BY}, or call with the holder's session; \
                 nothing was changed"
            ));
        }
    }
    if let (false, Some(seq)) = (landing_only, uncovered.first()) {
        return refused(format!(
            "criterion {seq} is an open criterion and no --met names it; every open criterion needs one --met N=REF"
        ));
    }
    if !landing_only {
        let already_met: i64 = conn.query_row(
            "SELECT count(*) FROM card_ac WHERE card_id = ?1 AND status = 'met'",
            [card_id],
            |r| r.get(0),
        )?;
        if already_met + (met.len() as i64) == 0 {
            return refused("the card has no acceptance criterion that is met".into());
        }
    }
    // The evidence the card layer could not check against the ledger.
    let mut record_landing = true;
    match evidence {
        FinishEvidence::Landing { head, .. } => {
            let recorded: i64 = conn.query_row(
                "SELECT count(*) FROM event WHERE card_id = ?1 AND kind = 'card.landed' \
                 AND json_extract(payload, '$.head') = ?2",
                params![card_id, head],
                |r| r.get(0),
            )?;
            record_landing = recorded == 0;
        }
        FinishEvidence::MergedPr => {
            let linked: i64 = conn.query_row(
                "SELECT count(*) FROM card_link WHERE from_card = ?1 AND kind = 'pr' \
                 AND state = 'active' AND pr_state = 'merged'",
                [card_id],
                |r| r.get(0),
            )?;
            if linked == 0 {
                return refused("no merged pull request is linked to the card".into());
            }
            record_landing = false;
        }
        FinishEvidence::Note { reference } => {
            let kind: Option<String> = conn
                .query_row(
                    "SELECT kind FROM card_note WHERE card_id = ?1 AND ref = ?2",
                    params![card_id, reference],
                    |r| r.get(0),
                )
                .optional()?;
            match kind.as_deref() {
                None => {
                    return refused(format!(
                        "the card has no note with ref {reference:?}; --evidence-note names a report or decision note of this card"
                    ))
                }
                Some("report" | "decision") => {}
                Some(other) => {
                    return refused(format!(
                        "note {reference:?} is a {other} note; --evidence-note needs a report or decision note of this card"
                    ))
                }
            }
        }
        FinishEvidence::Report { .. } => {}
    }
    Ok(Ok(FinishPlan {
        from,
        newest_generation,
        previous,
        uncovered,
        landing_only,
        record_landing,
    }))
}

/// The ledger row `card land` writes before it moves the integration branch (SPEC-20261121-ref-guard,
/// D-8): the free event kind `card.land_started` with the branch, the tip that was read, the
/// candidate head and the caller. It is written in one transaction that runs every check of
/// [`finish`], so it is refused for the same reasons and a refused call leaves no row; one row
/// exists per card, head and tip (a rerun finds it and writes none). Returns whether a row was
/// written. A close refused after the move leaves the row as the ledger's record of the move.
#[allow(clippy::too_many_arguments)]
pub fn record_land_started(
    w: &mut WriteLedger,
    card_id: i64,
    session: Option<&str>,
    by: &str,
    evidence: &FinishEvidence<'_>,
    met: &[(i64, String)],
    generation: Option<i64>,
    branch: &str,
    from_tip: &str,
) -> Result<bool, WorkflowError> {
    finish_input_checks(by, met, evidence)?;
    w.with_write_tx(|tx, now| {
        if let Err(e) = finish_plan(tx, card_id, session, by, evidence, met, generation, now)? {
            return Ok(Err(e));
        }
        let FinishEvidence::Landing { head, .. } = evidence else {
            return Ok(Ok(false));
        };
        let exists: bool = tx.query_row(
            "SELECT EXISTS (SELECT 1 FROM event WHERE card_id = ?1 AND kind = 'card.land_started' \
             AND json_extract(payload, '$.head') = ?2 AND json_extract(payload, '$.tip') = ?3)",
            params![card_id, head, from_tip],
            |r| r.get(0),
        )?;
        if exists {
            return Ok(Ok(false));
        }
        reservation::add_event(
            tx,
            now,
            session.unwrap_or("user"),
            card_id,
            "card.land_started",
            &json!({"branch": branch, "tip": from_tip, "head": head, "by": by}),
        )?;
        Ok(Ok(true))
    })?
}

/// What [`finish`] would refuse before its first write, asked without writing anything: the same
/// checks, run against the ledger as it is now. `Ok(())` is no promise: a call that gets in
/// between can change the ledger, and the write transaction checks again. `card land` uses it so
/// that a refusal comes before the integration branch moves.
pub fn finish_preflight(
    conn: &Connection,
    card_id: i64,
    session: Option<&str>,
    by: &str,
    evidence: &FinishEvidence<'_>,
    met: &[(i64, String)],
    generation: Option<i64>,
) -> Result<(), WorkflowError> {
    finish_input_checks(by, met, evidence)?;
    let now: i64 = conn
        .query_row("SELECT unixepoch()", [], |r| r.get(0))
        .map_err(LedgerError::from)?;
    finish_plan(conn, card_id, session, by, evidence, met, generation, now)??;
    Ok(())
}

/// Closes a `ready`, `active` or `review` card in one write transaction: each open criterion is
/// marked met with the reference `met` gives for it, the evidence is recorded, the card walks
/// the rows of the transition table to `done`, a held reservation is marked `released`, and two
/// events are written (`card.landed` and `card.finish`; only the finish event when the same head
/// is already recorded or a merged pull request is the evidence). The landing event records the
/// head, the tip, the generation of the card's newest reservation and `by`.
///
/// When an open criterion has no met option and the evidence is a landing (a `Landing` head or
/// a `MergedPr` link: a card that changes the repository), the call is a landing-only call: it
/// writes the landing event once (nothing when the same head is already recorded), changes
/// nothing else and answers `closed: false` with the open criteria. Report and note evidence
/// keep the refusal. A met option that names a criterion that is not open, or one twice, is
/// refused in both cases.
///
/// `generation`, when given, must equal the generation of the card's newest reservation (0 for a
/// card that never had one); another number is refused before any write.
///
/// This is not a fenced write: a fence refusal commits a `fence_reject` event and a finish
/// refusal must write nothing, so every check comes before the first write, inside the one
/// write transaction that takes the write lock first (`BEGIN IMMEDIATE`): of two calls that
/// reach one card at the same time the second reads the first one's result and is refused. A
/// `ready` card has no reservation, so none is made and no generation is spent.
pub fn finish(
    w: &mut WriteLedger,
    card_id: i64,
    session: Option<&str>,
    by: &str,
    evidence: &FinishEvidence<'_>,
    met: &[(i64, String)],
    generation: Option<i64>,
) -> Result<Finished, WorkflowError> {
    finish_input_checks(by, met, evidence)?;
    w.with_write_tx(|tx, now| {
        let FinishPlan {
            from,
            newest_generation,
            previous,
            uncovered,
            landing_only,
            record_landing,
        } = match finish_plan(tx, card_id, session, by, evidence, met, generation, now)? {
            Ok(plan) => plan,
            Err(e) => return Ok(Err(e)),
        };

        // The writes. A landing-only call writes the landing event and nothing else.
        let actor = session.unwrap_or("user");
        let (via, head, tip, report, note) = match evidence {
            FinishEvidence::Landing { head, tip } => {
                ("landing", Some(*head), Some(*tip), None, None)
            }
            FinishEvidence::MergedPr => ("merged_pr", None, None, None, None),
            FinishEvidence::Report { path } => ("report", None, None, Some(*path), None),
            FinishEvidence::Note { reference } => ("note", None, None, None, Some(*reference)),
        };
        let mut events = 0;
        if record_landing {
            reservation::add_event(
                tx,
                now,
                actor,
                card_id,
                "card.landed",
                &json!({
                    "head": head, "tip": tip, "generation": newest_generation, "by": by,
                    "report": report, "note": note,
                    "via": "finish", "no_code": head.is_none(),
                }),
            )?;
            events += 1;
        }
        if landing_only {
            return Ok(Ok(Finished {
                card_id,
                state: from.clone(),
                from,
                closed: false,
                rev: card_rev(tx, card_id)?,
                closed_at: None,
                open_criteria: uncovered,
                previous,
                released: false,
                events,
            }));
        }
        for (seq, text) in met {
            tx.execute(
                "UPDATE card_ac SET status = 'met', evidence_ref = ?1, withdrawn_reason = NULL, \
                 updated_at = ?2 WHERE card_id = ?3 AND seq = ?4 AND status = 'open'",
                params![text, now, card_id, seq],
            )?;
        }
        let steps: &[&str] = match from.as_str() {
            "ready" => &["active", "review", "done"],
            "active" => &["review", "done"],
            _ => &["done"],
        };
        for step in steps {
            if *step == "done" {
                tx.execute(
                    "UPDATE card SET state = 'done', closed_at = ?2, updated_at = ?2 \
                     WHERE card_id = ?1",
                    params![card_id, now],
                )?;
            } else {
                tx.execute(
                    "UPDATE card SET state = ?2, updated_at = ?3 WHERE card_id = ?1",
                    params![card_id, step, now],
                )?;
            }
        }
        let released = previous.is_some();
        if let Some(held) = &previous {
            tx.execute(
                "UPDATE reservation SET status = 'released', updated_at = ?1 \
                 WHERE card_id = ?2 AND generation = ?3",
                params![now, card_id, held.generation],
            )?;
        }
        reservation::add_event(
            tx,
            now,
            actor,
            card_id,
            "card.finish",
            &json!({
                "from": from, "by": by, "session": session, "via": via,
                "head": head, "tip": tip, "report": report, "note": note,
                "met": met.iter().map(|(seq, text)| json!({"seq": seq, "evidence": text})).collect::<Vec<_>>(),
                "previous": previous.as_ref().map(|p| json!({
                    "holder": p.holder, "generation": p.generation, "lease_until": p.lease_until,
                })),
                "released": released, "closed_at": now,
            }),
        )?;
        events += 1;
        Ok(Ok(Finished {
            card_id,
            from,
            closed: true,
            state: "done".to_string(),
            rev: card_rev(tx, card_id)?,
            closed_at: Some(now),
            open_criteria: Vec::new(),
            previous,
            released,
            events,
        }))
    })?
}

/// What the ledger records about a landing of one card that is not `done`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LandingFacts {
    pub card_id: i64,
    pub state: String,
    /// The head of the newest `card.landed` event that names one. An event with no head (a
    /// report close) is never a head.
    pub head: Option<String>,
    /// The newest active merged pull request link: `<host>:<project>#<number>` and its merge ref.
    pub merged_pr: Option<(String, String)>,
}

/// The landing facts of every card that is not `done`, in card order. Read-only: it names no
/// SPEC and runs nothing outside the ledger.
pub fn landing_facts(conn: &Connection) -> Result<Vec<LandingFacts>, LedgerError> {
    let mut stmt = conn.prepare(
        "SELECT card_id, state, \
           (SELECT json_extract(e.payload, '$.head') FROM event e \
             WHERE e.card_id = card.card_id AND e.kind = 'card.landed' \
               AND json_type(e.payload, '$.head') = 'text' \
             ORDER BY e.event_id DESC LIMIT 1), \
           (SELECT l.target_host || ':' || l.target_project || '#' || l.target_number \
              FROM card_link l WHERE l.from_card = card.card_id AND l.kind = 'pr' \
               AND l.state = 'active' AND l.pr_state = 'merged' ORDER BY l.link_id DESC LIMIT 1), \
           (SELECT l.merge_ref FROM card_link l WHERE l.from_card = card.card_id \
               AND l.kind = 'pr' AND l.state = 'active' AND l.pr_state = 'merged' \
             ORDER BY l.link_id DESC LIMIT 1) \
         FROM card WHERE state <> 'done' ORDER BY card_id",
    )?;
    let rows = stmt.query_map([], |r| {
        let target: Option<String> = r.get(3)?;
        let merge_ref: Option<String> = r.get(4)?;
        Ok(LandingFacts {
            card_id: r.get(0)?,
            state: r.get(1)?,
            head: r.get(2)?,
            merged_pr: target.zip(merge_ref),
        })
    })?;
    Ok(rows.collect::<Result<_, _>>()?)
}

// ---- write boundary ----------------------------------------------------------------------------

/// True when the repository path `path` lies under the write-scope pattern `pattern`: the
/// pattern itself, anything below a plain pattern, `*` inside one segment, `**` for any number
/// of segments. Both are compared lower-cased with `.` and `..` resolved, as
/// [`crate::scope::overlap`] does; a path or pattern that climbs out of the repository is never
/// covered, and neither is a folder above the pattern.
pub fn covers(pattern: &str, path: &str) -> bool {
    let (Some(pattern), Some(path)) = (segments(pattern), segments(path)) else {
        return false;
    };
    if pattern.is_empty() || path.is_empty() {
        return false;
    }
    covers_from(&pattern, &path)
}

/// The segments of a path, lower-cased, with `.` and `..` resolved; `None` when a `..` climbs
/// out of the repository.
fn segments(path: &str) -> Option<Vec<String>> {
    let mut out: Vec<String> = Vec::new();
    for s in path.trim().split('/') {
        match s {
            "" | "." => {}
            ".." => {
                out.pop()?;
            }
            _ => out.push(s.to_lowercase()),
        }
    }
    Some(out)
}

fn covers_from(pattern: &[String], path: &[String]) -> bool {
    match pattern.split_first() {
        // The pattern is used up: it covers itself and everything below it.
        None => true,
        Some((first, rest)) if first == "**" => {
            (0..=path.len()).any(|skip| covers_from(rest, &path[skip..]))
        }
        Some((first, rest)) => match path.split_first() {
            // The path ends above the pattern: a folder that holds more than the pattern.
            None => false,
            Some((name, below)) => segment_matches(first, name) && covers_from(rest, below),
        },
    }
}

/// `*` in a pattern segment stands for any run of characters of one path segment.
fn segment_matches(pattern: &str, name: &str) -> bool {
    let parts: Vec<&str> = pattern.split('*').collect();
    if parts.len() == 1 {
        return pattern == name;
    }
    let (first, last) = (parts[0], parts[parts.len() - 1]);
    if name.len() < first.len() + last.len() || !name.starts_with(first) || !name.ends_with(last) {
        return false;
    }
    let mut rest = &name[first.len()..name.len() - last.len()];
    for mid in &parts[1..parts.len() - 1] {
        match rest.find(mid) {
            Some(i) => rest = &rest[i + mid.len()..],
            None => return false,
        }
    }
    true
}

/// The check `card check` makes for one path: the first pattern of the reservation's scope that
/// [`covers`] it.
pub fn covering_pattern<'a>(reservation: &'a Reservation, path: &str) -> Option<&'a str> {
    reservation
        .write_scope
        .iter()
        .map(String::as_str)
        .find(|p| covers(p, path))
}

// ---- shared ------------------------------------------------------------------------------------

/// One line of 1 to `max` characters that is not only blanks.
fn one_line(what: &str, value: &str, max: usize) -> Result<(), String> {
    let n = value.chars().count();
    if value.trim().is_empty() {
        Err(format!("{what} is empty"))
    } else if n > max {
        Err(format!("{what} is longer than {max} characters ({n})"))
    } else if value.chars().any(char::is_control) {
        Err(format!("{what} is not one line"))
    } else if value.chars().any(is_directional) {
        Err(format!(
            "{what} holds a directional or formatting character (U+2028–U+202E, U+2066–U+2069)"
        ))
    } else {
        Ok(())
    }
}

/// A bidi or directional character (C-68 AC-3): one of these changes how the text around it
/// is drawn, whatever it claims to be, the same set `decide`'s normaliser refuses.
fn is_directional(c: char) -> bool {
    matches!(c, '\u{2028}'..='\u{202e}' | '\u{2066}'..='\u{2069}')
}

fn card_exists(conn: &Connection, card_id: i64) -> Result<bool, LedgerError> {
    Ok(conn
        .query_row("SELECT 1 FROM card WHERE card_id = ?1", [card_id], |_| {
            Ok(())
        })
        .optional()?
        .is_some())
}

fn card_state(conn: &Connection, card_id: i64) -> Result<String, LedgerError> {
    Ok(conn.query_row(
        "SELECT state FROM card WHERE card_id = ?1",
        [card_id],
        |r| r.get(0),
    )?)
}

fn card_rev(conn: &Connection, card_id: i64) -> Result<i64, LedgerError> {
    Ok(
        conn.query_row("SELECT rev FROM card WHERE card_id = ?1", [card_id], |r| {
            r.get(0)
        })?,
    )
}
