//! `moai card sweep [C-n] [--dry-run] --by leader` and the grace period it uses
//! (SPEC-20261124-ended-lease-sweep): the cards whose lease ended long ago go back to `ready`
//! in one command, with one note on each card. The rules are the ledger's (`ledger::sweep`);
//! this module reads what the user typed and the project profile, opens the ledger and prints
//! what the ledger returned.

use super::command::open_for_write;
use super::input::{self, CardRef};
use crate::ledger::sweep::{self, Outcome, SkipReason, Swept};
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// How many hours a lease may be over before a sweep returns its card, when the profile names none.
pub const DEFAULT_GRACE_HOURS: u64 = 24;

const HOUR_SECS: i64 = 3_600;

/// The grace period of the project in seconds: the profile key `lease_sweep_hours`, else the
/// default. A project with no profile has the default; a profile that cannot be read is an error
/// the caller reports, never a guess.
pub fn grace_secs(project: &Path) -> Result<i64> {
    Ok(grace_hours(project)?
        .try_into()
        .unwrap_or(i64::MAX / HOUR_SECS)
        .saturating_mul(HOUR_SECS))
}

fn grace_hours(project: &Path) -> Result<u64> {
    if crate::profile::chosen_profile(project)
        .map_err(|e| e.to_string())?
        .is_none()
    {
        return Ok(DEFAULT_GRACE_HOURS);
    }
    let (profile, _) = crate::profile::read_profile(project).map_err(|e| e.to_string())?;
    Ok(profile
        .lease_sweep_hours
        .filter(|h| *h > 0)
        .unwrap_or(DEFAULT_GRACE_HOURS))
}

/// The grace period for a command that must not fail on a profile it cannot read (`card take`,
/// `card next`): the default then.
pub fn grace_secs_or_default(project: &Path) -> i64 {
    grace_secs_with_source(project).0
}

/// The grace period and where it came from: `profile` (the key is set), `default` (no profile, or
/// none that sets the key) or `default_unreadable_profile` (a profile that cannot be read has no
/// key to read, so the default holds, and the answer says so).
pub fn grace_secs_with_source(project: &Path) -> (i64, &'static str) {
    let default = DEFAULT_GRACE_HOURS as i64 * HOUR_SECS;
    match crate::profile::chosen_profile(project) {
        Ok(None) => return (default, "default"),
        Err(_) => return (default, "default_unreadable_profile"),
        Ok(Some(_)) => {}
    }
    match crate::profile::read_profile(project) {
        Ok((profile, _)) => match profile.lease_sweep_hours.filter(|h| *h > 0) {
            Some(hours) => (
                i64::try_from(hours)
                    .unwrap_or(i64::MAX / HOUR_SECS)
                    .saturating_mul(HOUR_SECS),
                "profile",
            ),
            None => (default, "default"),
        },
        Err(_) => (default, "default_unreadable_profile"),
    }
}

fn caller_session() -> Option<String> {
    std::env::var("CLAUDE_CODE_SESSION_ID")
        .or_else(|_| std::env::var("CODEX_THREAD_ID"))
        .ok()
}

fn swept_json(s: &Swept) -> Value {
    json!({
        "name": input::card_name(s.card_id), "card_id": s.card_id, "holder": s.holder,
        "generation": s.generation, "lease_until": s.lease_until,
        "worktree": s.worktree, "branch": s.branch,
    })
}

fn skipped_json(card_id: i64, reason: &SkipReason) -> Value {
    let mut v = json!({
        "name": input::card_name(card_id), "card_id": card_id, "reason": reason.code(),
    });
    match reason {
        SkipReason::WithinGrace {
            lease_until,
            sweepable_at,
        } => {
            v["lease_until"] = json!(lease_until);
            v["sweepable_at"] = json!(sweepable_at);
        }
        SkipReason::NotHeld { status } => v["status"] = json!(status),
        SkipReason::InReview | SkipReason::Changed | SkipReason::NoEndedLease => {}
    }
    v
}

fn outcome_json(id: &str, grace: i64, out: &Outcome) -> Value {
    json!({
        "action": "sweep", "project_id": id, "dry_run": out.dry_run,
        "grace_hours": grace / HOUR_SECS,
        "swept": out.swept.iter().map(swept_json).collect::<Vec<_>>(),
        "skipped": out.skipped.iter().map(|s| skipped_json(s.card_id, &s.reason)).collect::<Vec<_>>(),
        "writes": !out.dry_run && !out.swept.is_empty(),
    })
}

/// `card sweep [C-n] [--dry-run] --by leader`: every card (or the one named) whose `held` lease
/// ended at least the grace period ago goes back to `ready`. Without `--by leader` the command
/// is refused and nothing is written; a dry run only lists.
pub fn sweep(
    project: &Path,
    home: &Path,
    reference: Option<&str>,
    by: &str,
    dry_run: bool,
) -> Result<Value> {
    let card =
        match reference {
            Some(r) => match input::parse_ref(r).map_err(|e| format!("card sweep refused: {e}"))? {
                CardRef::Id(n) => Some(n),
                CardRef::SourceNumber(_) => return Err(
                    "card sweep refused: name the card by its C-number (`card show #N` finds it)"
                        .to_string(),
                ),
            },
            None => None,
        };
    let (grace, grace_source) = grace_secs_with_source(project);
    let (id, mut ledger) = open_for_write(project, home)?;
    let session = caller_session();
    let out = match card {
        None => sweep::sweep_all(&mut ledger, grace, by, session.as_deref(), dry_run)
            .map_err(|e| format!("card sweep refused: {e}"))?,
        Some(card_id) => {
            let mut out = Outcome {
                swept: Vec::new(),
                skipped: Vec::new(),
                dry_run,
            };
            // One card: list it as `sweep_all` would, so a dry run and a real run agree.
            let all = sweep::sweep_all(&mut ledger, grace, by, session.as_deref(), true)
                .map_err(|e| format!("card sweep refused: {e}"))?;
            out.swept = all
                .swept
                .into_iter()
                .filter(|s| s.card_id == card_id)
                .collect();
            out.skipped = all
                .skipped
                .into_iter()
                .filter(|s| s.card_id == card_id)
                .collect();
            if out.swept.is_empty() && out.skipped.is_empty() {
                out.skipped.push(sweep::Skipped {
                    card_id,
                    reason: SkipReason::NoEndedLease,
                });
            }
            if !dry_run {
                for s in std::mem::take(&mut out.swept) {
                    match sweep::sweep_card(
                        &mut ledger,
                        s.card_id,
                        Some(s.lease_until),
                        grace,
                        by,
                        session.as_deref(),
                    )
                    .map_err(|e| format!("card sweep refused: {e}"))?
                    {
                        Ok(done) => out.swept.push(done),
                        Err(reason) => out.skipped.push(sweep::Skipped {
                            card_id: s.card_id,
                            reason,
                        }),
                    }
                }
            }
            out
        }
    };
    let mut answer = outcome_json(&id, grace, &out);
    answer["grace_source"] = json!(grace_source);
    Ok(answer)
}
