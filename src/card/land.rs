//! `moai card land` (SPEC-20261121-card-land): the leader moves the integration branch to a
//! candidate commit by one compare-and-swap and closes the card in the one transaction
//! `card finish` uses.
//!
//! Order of a call, so that nothing is written until everything that can be known is known:
//!
//! 1. the caller word (only the leader lands until the typed verdict exists), the shape of the
//!    met options and of the head;
//! 2. the card: a card that closes on a report is `card finish`'s; a card already `done` with the
//!    same landed head answers done and writes nothing, with another head it refuses;
//! 3. git: the integration branch is read, the candidate is a commit, and where it stands
//!    against the tip decides the move (inside the tip: no move; the tip on its first-parent
//!    line: move; anything else is `stale_base`); a move is refused while any worktree has the
//!    branch checked out;
//! 4. the SPEC the card names passes at the candidate, and every check `card finish` would make
//!    passes ([`workflow::finish_preflight`], the same code the finish transaction runs first);
//! 5. the move: `git update-ref <branch> <candidate> <tip read>` is git's own compare-and-swap, so
//!    a branch that moved in between is refused as `tip_moved`;
//! 6. the close: [`workflow::finish`] with the candidate as the landed head. A close that is
//!    refused after the move leaves the branch moved and says so; the same call run again finds
//!    the candidate inside the tip, skips the move and closes (or repeats the refusal). Git is
//!    the record of the move and the ledger the record of the close, so no row is written
//!    between them.

use super::command::open_for_write;
use super::finish::{parse_met, previous_json};
use super::input::{self, CardRef};
use super::landed::{self, Containment, RefMove, Standing};
use super::spec_link;
use crate::ledger::workflow::{self, FinishEvidence, WorkflowError, LEADER_BY};
use crate::ledger::LedgerError;
use rusqlite::OptionalExtension;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// What `card land` was given.
pub struct LandRequest<'a> {
    /// This session's id, when it has one.
    pub session: Option<&'a str>,
    /// `--head`: the candidate commit (40 lowercase hex).
    pub head: &'a str,
    /// Each `--met N=REF`.
    pub met: &'a [String],
    /// `--generation`: the generation of the card's newest reservation the caller expects.
    pub generation: Option<i64>,
    /// `--by`: who lands; only the leader word is accepted.
    pub by: &'a str,
}

fn refused(why: impl std::fmt::Display) -> String {
    format!("card land refused: {why}")
}

/// A refusal that comes before any write.
fn unchanged(why: impl std::fmt::Display) -> String {
    refused(format!("{why}; nothing was changed"))
}

/// A refusal about the state of the repository (a conflict, not a mistake in the call).
fn conflict(why: impl std::fmt::Display) -> String {
    format!("card conflict in landing: {why}; nothing was changed")
}

fn card_id(reference: &str) -> Result<i64> {
    match input::parse_ref(reference).map_err(refused)? {
        CardRef::Id(n) => Ok(n),
        CardRef::SourceNumber(_) => Err(refused(
            "name the card by its C-number (`card show #N` finds it)",
        )),
    }
}

/// The sentence a refusal that wrote nothing ends with. Finish's own refusals end with it too, so
/// it is added once.
const NOTHING_CHANGED: &str = "nothing was changed";

/// A refusal of the finish checks, or of a finish that comes before the branch moved: said as
/// `card finish` words it for an input mistake, and with `nothing was changed` once for a check
/// that refused before any write.
fn finish_refusal(e: WorkflowError) -> String {
    match e {
        WorkflowError::InvalidInput(why) => refused(why),
        WorkflowError::FinishRefused(why) if why.ends_with(NOTHING_CHANGED) => refused(why),
        WorkflowError::FinishRefused(why) => unchanged(why),
        other => other.to_string(),
    }
}

/// A refusal of the close after the branch moved. The branch was changed, so the refusal never
/// says that nothing was: it says the branch moved and the call can be run again.
fn refusal_after_move(e: WorkflowError, branch: &str, head: &str) -> String {
    let why = match e {
        WorkflowError::InvalidInput(why) | WorkflowError::FinishRefused(why) => why,
        other => other.to_string(),
    };
    let why = why
        .strip_suffix(&format!("; {NOTHING_CHANGED}"))
        .unwrap_or(&why);
    format!(
        "card land refused after the move: {why}; the integration branch {branch} already moved \
         to {head} and the card is still open: run the same call again after the cause is repaired"
    )
}

/// `card land C-n`: see the module comment.
pub fn land(project: &Path, home: &Path, reference: &str, req: &LandRequest<'_>) -> Result<Value> {
    let card_id = card_id(reference)?;
    if req.by != LEADER_BY {
        return Err(unchanged(format!(
            "only the leader lands a card until the typed verdict exists: give --by {LEADER_BY}"
        )));
    }
    let met = parse_met(req.met).map_err(refused)?;
    if !crate::gate::is_full_sha(req.head) {
        return Err(refused("the head must be 40 lowercase hex"));
    }
    let sql_err = |e: rusqlite::Error| LedgerError::from(e).to_string();
    let (id, mut ledger) = open_for_write(project, home)?;
    let (kind, spec) = spec_link::kind_and_spec(ledger.conn(), card_id)
        .map_err(sql_err)?
        .ok_or_else(|| WorkflowError::CardNotFound(card_id).to_string())?;
    if !spec_link::changes_repository(&kind) {
        return Err(unchanged(format!(
            "a `{kind}` card closes on a report, not on a landing: use `card finish`"
        )));
    }
    let name = input::card_name(card_id);

    // A card that is already done: the same head answers done, another head refuses.
    let state: String = ledger
        .conn()
        .query_row(
            "SELECT state FROM card WHERE card_id = ?1",
            [card_id],
            |r| r.get(0),
        )
        .map_err(sql_err)?;
    if state == "done" {
        let recorded: Option<String> = ledger
            .conn()
            .query_row(
                "SELECT json_extract(payload, '$.head') FROM event \
                 WHERE card_id = ?1 AND kind = 'card.landed' \
                   AND json_type(payload, '$.head') = 'text' \
                 ORDER BY event_id DESC LIMIT 1",
                [card_id],
                |r| r.get(0),
            )
            .optional()
            .map_err(sql_err)?;
        return match recorded {
            Some(head) if head == req.head => Ok(json!({
                "action": "land", "project_id": id, "name": name, "card_id": card_id,
                "state": "done", "closed": true, "moved": false, "already_done": true,
                "head": req.head, "by": req.by, "writes": false,
            })),
            Some(head) => Err(unchanged(format!(
                "{name} is done with the landed head {head}; the head given is {}",
                req.head
            ))),
            None => Err(unchanged(format!(
                "{name} is done and records no landed head; the head given is {}",
                req.head
            ))),
        };
    }

    // Git: the branch, the tip, the candidate and where it stands.
    let branch = landed::integration_branch(project).map_err(|why| {
        unchanged(format!(
            "a landing is made on the integration branch, and the profile gives none ({why})"
        ))
    })?;
    let full = landed::branch_ref(&branch).map_err(unchanged)?;
    let tip = landed::tip_of(project, &branch).map_err(unchanged)?;
    match landed::contained(project, req.head, &tip).map_err(unchanged)? {
        Containment::NotACommit => {
            return Err(unchanged(format!(
                "the head {} is not a commit of this repository",
                req.head
            )))
        }
        Containment::Inside | Containment::Outside => {}
    }
    let standing = landed::standing(project, req.head, &tip).map_err(unchanged)?;
    match standing {
        Standing::Inside => {}
        Standing::OnTip => {
            let trees = landed::checked_out_at(project, &full).map_err(unchanged)?;
            if !trees.is_empty() {
                return Err(conflict(format!(
                    "the integration branch {branch} is checked out in {}: a move under it would \
                     leave that worktree behind the reference",
                    trees.join(", ")
                )));
            }
        }
        Standing::Stale => {
            return Err(conflict(format!(
                "stale_base: the integration tip {tip} is not on the first-parent line of {}: \
                 merge the tip into the head and land the merge",
                req.head
            )))
        }
    }
    let to_tip = if standing == Standing::OnTip {
        req.head.to_string()
    } else {
        tip.clone()
    };

    // Every check that can be made without a write: the SPEC at the candidate, then finish's own.
    if let Some(why) = spec_link::done_refusal_at_tip(project, &name, spec.as_deref(), req.head) {
        return Err(unchanged(why));
    }
    let evidence = FinishEvidence::Landing {
        head: req.head,
        tip: &to_tip,
    };
    workflow::finish_preflight(
        ledger.conn(),
        card_id,
        req.session,
        req.by,
        &evidence,
        &met,
        req.generation,
    )
    .map_err(finish_refusal)?;

    // The move: git's compare-and-swap on the tip that was read.
    let mut moved = false;
    if standing == Standing::OnTip {
        let message = format!("moai card land {}", &req.head[..12]);
        match landed::move_branch(project, &full, req.head, &tip, &message).map_err(unchanged)? {
            RefMove::Moved => moved = true,
            RefMove::Lost(why) => {
                // Read where the branch is now: another call may have landed this very head a
                // moment ago (nothing to move), the branch may have moved to something else
                // (`tip_moved`), or it did not move at all and git refused for another reason (a
                // lock file, a read-only store): that one is git's reason, not a moved tip.
                let now = landed::tip_of(project, &branch).map_err(unchanged)?;
                if now == tip {
                    return Err(unchanged(format!(
                        "git refused to move the integration branch {branch} to {}: {why}",
                        req.head
                    )));
                }
                if now != req.head {
                    return Err(conflict(format!(
                        "tip_moved: the integration branch {branch} no longer points at {tip} \
                         (now {now}; git said: {why}); read the tip again and run the call again"
                    )));
                }
            }
        }
    }

    // The close. A refusal here comes after the move and says so.
    let done = workflow::finish(
        &mut ledger,
        card_id,
        req.session,
        req.by,
        &evidence,
        &met,
        req.generation,
    )
    .map_err(|e| {
        if moved {
            refusal_after_move(e, &branch, req.head)
        } else {
            finish_refusal(e)
        }
    })?;
    let applied: &[(i64, String)] = if done.closed { &met } else { &[] };
    Ok(json!({
        "action": "land", "project_id": id, "name": name, "card_id": card_id,
        "branch": branch, "moved": moved, "from_tip": tip, "to_tip": to_tip, "head": req.head,
        "from": done.from, "state": done.state, "rev": done.rev, "closed_at": done.closed_at,
        "closed": done.closed, "open_criteria": done.open_criteria,
        "met": applied.iter().map(|(seq, text)| json!({"seq": seq, "evidence": text})).collect::<Vec<_>>(),
        "by": req.by, "session": req.session,
        "previous": previous_json(&done.previous), "released": done.released,
        "events": done.events, "writes": done.events > 0 || moved,
    }))
}
