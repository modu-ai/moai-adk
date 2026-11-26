//! `moai card adopt` (SPEC-20261123-lane-continues-card): after a `/clear` or a restart the same
//! lane — the same registered name, the same git worktree, a new host session id — takes its own
//! card over at the next generation, with no wait for the lease and none of the proofs of an
//! early reclaim.
//!
//! The proofs are the session registry's own record and are read before anything is written
//! (REQ-10): the running session's entry lists the holder session in `previous_sessions`
//! (`not_same_lane`), no entry still names the holder (`holder_registered`), and the worktree
//! the reservation records, the worktree of the running folder and the worktree the entry
//! records are one (`other_worktree`). The card's state, the worktree's changes and processes
//! and the `.live` lock are never read (REQ-12): the same lane in the same tree is the one who
//! has the work, and the write fence stops a mistaken adoption from writing. The write is one
//! ledger transaction (`ledger::adopt`) that re-checks the holder the proofs were read for.

use super::command::open_for_write;
use super::dispatch::{reservation_json, DEFAULT_LEASE_SECS};
use super::input::{self, CardRef};
use crate::lane_registry::{self, SessionEntry};
use crate::ledger::adopt::{self as ledger_adopt, AdoptProof};
use crate::ledger::reservation::ReclaimError;
use crate::ledger::{self};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// What `card adopt` was given.
pub struct AdoptRequest<'a> {
    /// The new holder: this session's own id.
    pub to_holder: &'a str,
    /// The generation the caller saw as the card's current one.
    pub generation: i64,
    /// Lease of the new reservation in seconds (default [`DEFAULT_LEASE_SECS`]).
    pub lease_secs: Option<i64>,
    /// The running folder, whose git worktree is the one the lane continues in.
    pub running_dir: &'a Path,
}

/// A refusal that names the proof that failed. The `conflict in` wording keeps
/// `contract::classify` honest: it is a `conflict`.
fn refuse(code: &str, why: impl std::fmt::Display) -> String {
    format!("card conflict in {code}: {why}")
}

fn now_secs() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64)
}

/// `card adopt C-n --generation G`: see the module text. A card that cannot be taken over for
/// its state is refused with the answer `card reclaim` gives for the same state; every refusal
/// writes nothing.
pub fn adopt(
    project: &Path,
    home: &Path,
    reference: &str,
    req: &AdoptRequest<'_>,
) -> Result<Value> {
    let card_id =
        match input::parse_ref(reference).map_err(|e| format!("card adopt refused: {e}"))? {
            CardRef::Id(n) => n,
            CardRef::SourceNumber(_) => {
                return Err(
                    "card adopt refused: name the card by its C-number (`card show #N` finds it)"
                        .to_string(),
                )
            }
        };
    // The state first, on a connection that cannot write: the holder and worktree the proofs
    // are read for come from the reservation as it is now.
    let id = crate::ledger::command::project_id_of(project)
        .map_err(|e| format!("card adopt refused: {e}"))?;
    let old = match ledger::open_read_only(home, &id) {
        Ok(ro) => {
            ledger_adopt::peek(ro.conn(), card_id, req.generation).map_err(|e| e.to_string())?
        }
        Err(_) => {
            // The ledger's own refusal (absent, foreign, newer) is the answer.
            open_for_write(project, home)?;
            return Err("card adopt refused: the ledger cannot be read".to_string());
        }
    };
    let holder = old.holder.clone();

    let proofs = Proofs::read(project, req, &holder, old.worktree.as_deref())?;

    let (id, mut ledger) = open_for_write(project, home)?;
    let lease_secs = req.lease_secs.unwrap_or(DEFAULT_LEASE_SECS);
    let adopted = ledger_adopt::adopt(
        &mut ledger,
        card_id,
        req.to_holder,
        req.generation,
        lease_secs,
        &AdoptProof {
            holder: &holder,
            lane: &proofs.lane,
            worktree: &proofs.worktree,
            previous_sessions: &proofs.previous_sessions,
        },
    )
    .map_err(|e| match e {
        ReclaimError::HolderChanged { proven, current } => refuse(
            "holder_changed",
            format!(
                "the proofs were read while session {proven} held the card, but session \
                 {current} holds it now"
            ),
        ),
        other => other.to_string(),
    })?;
    Ok(json!({
        "action": "adopt", "project_id": id,
        "reservation": reservation_json(&adopted.reservation),
        "previous": reservation_json(&adopted.previous),
        "lane": proofs.lane, "worktree": proofs.worktree,
        "previous_sessions": proofs.previous_sessions,
        "state": adopted.card_state,
        "events": 1, "writes": true,
    }))
}

/// What the proofs found: the lane's name, the canonical worktree and the ids the entry lists.
struct Proofs {
    lane: String,
    worktree: String,
    previous_sessions: Vec<String>,
}

impl Proofs {
    /// REQ-5, REQ-6, REQ-7 in this order; each is read only.
    fn read(
        project: &Path,
        req: &AdoptRequest<'_>,
        holder: &str,
        reserved_worktree: Option<&str>,
    ) -> Result<Proofs> {
        let entries = lane_registry::session_entries(project).map_err(|why| {
            refuse(
                "not_same_lane",
                format!("the session registry cannot be read: {why}"),
            )
        })?;
        let now = now_secs();
        // REQ-5: this session's entry lists the holder among its previous sessions.
        let mine: &SessionEntry = entries
            .iter()
            .filter(|e| e.session.as_deref() == Some(req.to_holder))
            .max_by_key(|e| e.lease_until)
            .ok_or_else(|| {
                refuse(
                    "not_same_lane",
                    "this session has no entry in the session registry; run \
                     `session-register --continue` for the lane first",
                )
            })?;
        if mine.lease_until <= now {
            return Err(refuse(
                "not_same_lane",
                format!(
                    "the registry entry `{}` of this session has ended its lease; run \
                     `session-register --continue` again",
                    mine.name
                ),
            ));
        }
        if !mine.previous_sessions.iter().any(|s| s == holder) {
            return Err(refuse(
                "not_same_lane",
                format!(
                    "the registry entry `{}` of this session does not list the holder session \
                     of the card among its previous sessions",
                    mine.name
                ),
            ));
        }
        // REQ-6: nothing still registers the holder, under any lane name or lease state.
        if let Some(other) = entries
            .iter()
            .find(|e| e.session.as_deref() == Some(holder))
        {
            return Err(refuse(
                "holder_registered",
                format!(
                    "the lane `{}` still registers the holder session of the card; the holder \
                     is not gone",
                    other.name
                ),
            ));
        }
        // REQ-7: the reservation, the running folder and the entry name one worktree.
        let Some(reserved) = reserved_worktree.filter(|w| !w.is_empty()) else {
            return Err(refuse(
                "other_worktree",
                "the reservation records no worktree",
            ));
        };
        let reserved = std::fs::canonicalize(reserved).map_err(|e| {
            refuse(
                "other_worktree",
                format!("the worktree the reservation records cannot be resolved: {e}"),
            )
        })?;
        let Some(running) = lane_registry::worktree_of(req.running_dir) else {
            return Err(refuse(
                "other_worktree",
                "the running folder is not inside a git worktree",
            ));
        };
        match mine.worktree.as_deref() {
            None => {
                return Err(refuse(
                    "other_worktree",
                    format!("the registry entry `{}` records no worktree", mine.name),
                ))
            }
            Some(w) if w != running => {
                return Err(refuse(
                    "other_worktree",
                    format!(
                        "the registry entry `{}` records another worktree than the running one",
                        mine.name
                    ),
                ))
            }
            Some(_) => {}
        }
        if reserved.to_str() != Some(running.as_str()) {
            return Err(refuse(
                "other_worktree",
                "the reservation records another worktree than the running one",
            ));
        }
        Ok(Proofs {
            lane: mine.name.clone(),
            worktree: running,
            previous_sessions: mine.previous_sessions.clone(),
        })
    }
}
