//! The work on a card after it is taken (card C-53 S1): which card comes next, notes, the
//! state of an acceptance criterion, the pull request link, `active -> review`, `review -> done`
//! and the write-boundary predicate. The reservation rules stay in [`super::reservation`]; this
//! module is the SQL of the finish path and nothing else.
//!
//! - [`next`] only reads: the cards a holder holds, the `ready` cards it could take (priority,
//!   then number) and every card it cannot take, with the reason.
//! - [`add_note`] appends one note. Notes are not fenced: a decision is answered by someone who
//!   does not hold the card. Every note gets a `ref` that is unique on its card.
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
}

impl SkipReason {
    pub fn code(&self) -> &'static str {
        match self {
            SkipReason::Held { .. } => "held",
            SkipReason::ScopeOverlap { .. } => "scope_overlap",
            SkipReason::EmptyScope => "empty_scope",
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
        if state != "ready" {
            continue;
        }
        // A scope that is not a list of strings is no scope: `take` could not use it either.
        let scope: Vec<String> = serde_json::from_str(&write_scope).unwrap_or_default();
        let reason = if scope.is_empty() {
            Some(SkipReason::EmptyScope)
        } else {
            reservation::find_overlap(conn, card_id, &scope)?.map(|(other_card, mine, theirs)| {
                SkipReason::ScopeOverlap {
                    other_card,
                    mine,
                    theirs,
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
/// card; without one a random one is made (`note:<card>:<32 hex digits>`).
pub fn add_note(
    w: &mut WriteLedger,
    card_id: i64,
    kind: &str,
    author: &str,
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
    if body.trim().is_empty() {
        return Err(invalid("the note text is empty".into()));
    }
    if body.contains('\0') {
        return Err(invalid("the note text contains a NUL character".into()));
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
            &json!({"note_id": note_id, "kind": kind, "ref": reference, "bytes": body.len()}),
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
/// current holder at `generation` (a live lease is not needed).
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
        let existing: Option<i64> = tx
            .query_row(
                "SELECT link_id FROM card_link WHERE from_card = ?1 AND kind = 'pr' \
                     AND target_host = ?2 AND target_project = ?3 AND target_number = ?4",
                params![card_id, t.host, t.project, t.number],
                |r| r.get(0),
            )
            .optional()?;
        let link_id = match existing {
            Some(link_id) => {
                tx.execute(
                    "UPDATE card_link SET pr_state = ?1, merge_ref = ?2, state = 'active', \
                         updated_at = ?3 WHERE link_id = ?4",
                    params![t.pr_state, merge_ref, now, link_id],
                )?;
                link_id
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
        Ok(PrLink {
            link_id,
            card_id,
            host: t.host.to_string(),
            project: t.project.to_string(),
            number: t.number,
            pr_state: t.pr_state.to_string(),
            merge_ref: merge_ref.map(str::to_string),
            created: existing.is_none(),
        })
    })
    .map_err(WorkflowError::from)
}

// ---- review and done ---------------------------------------------------------------------------

/// A card after [`review`] or [`done`].
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
    } else {
        Ok(())
    }
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
