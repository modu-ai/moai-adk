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

/// `review -> done`, as the current holder at `generation` (a live lease is not needed). Needs
/// what `card_done_guard` needs: a met criterion, no open criterion and an active `pr` link
/// whose state is `merged`. The reservation ends (`released`) in the same transaction.
pub fn done(
    w: &mut WriteLedger,
    card_id: i64,
    holder: &str,
    generation: i64,
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
            } else if count(
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
                &json!({"from": from, "generation": generation, "closed_at": now}),
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
