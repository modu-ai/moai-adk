//! The `stale` key of `moai card next` (SPEC-20261124-card-fold-stale): the cards that want
//! closing, one row each, with the command that closes it. Read-only: the rows only say what to
//! run.
//!
//! A card has one row, under the first kind that fits: `landed_open`, `lease_ended`,
//! `review_no_holder`, `old_draft`. A landed card whose lease ended is closed by `card finish`, not
//! by a withdraw, so the landing wins.

use super::input;
use super::landed::Proof;
use crate::ledger::close::StaleFacts;
use crate::ledger::sweep::sweep_via;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::path::Path;

/// How many days a `draft` card may sit before it is listed, when the profile names none.
pub const DEFAULT_DRAFT_DAYS: u64 = 14;

const DAY_SECS: i64 = 86_400;

/// Whether a draft last updated at `updated_at` is at least `days` old at `now`.
pub fn is_old_draft(updated_at: i64, now: i64, days: u64) -> bool {
    let window = i64::try_from(days)
        .unwrap_or(i64::MAX)
        .saturating_mul(DAY_SECS);
    now.saturating_sub(updated_at) >= window
}

/// The stale-draft days of the project: the profile key `stale_draft_days`, else the default. A
/// project with no profile has the default; a profile that cannot be read is an error the caller
/// reports, never a guess.
pub fn draft_days(project: &Path) -> Result<u64, String> {
    if crate::profile::chosen_profile(project)
        .map_err(|e| e.to_string())?
        .is_none()
    {
        return Ok(DEFAULT_DRAFT_DAYS);
    }
    let (profile, _) = crate::profile::read_profile(project).map_err(|e| e.to_string())?;
    Ok(profile
        .stale_draft_days
        .filter(|d| *d > 0)
        .unwrap_or(DEFAULT_DRAFT_DAYS))
}

/// The command that withdraws `name`; the reason is the caller's to fill.
fn withdraw_command(name: &str) -> String {
    format!("moai card withdraw {name} --reason <why> --by leader")
}

fn row(
    card_id: i64,
    title: &str,
    state: &str,
    priority: &str,
    kind: &str,
    command: String,
) -> Value {
    json!({
        "name": input::card_name(card_id), "card_id": card_id, "title": title,
        "state": state, "priority": priority, "kind": kind, "command": command,
    })
}

/// The command that returns a card whose lease ended long ago to `ready`
/// (SPEC-20261124-ended-lease-sweep).
fn sweep_command(name: &str) -> String {
    format!("moai card sweep {name} --by leader")
}

/// The rows of `stale`, ordered by card. `proofs` are the landing proofs of `landed::find` and
/// `now` is the DB clock of the read; a draft is listed when [`is_old_draft`] says so for `days`.
/// A card whose lease ended at least `grace_secs` ago, or whose holder has written nothing for the
/// silence window, is closed by a sweep, not by a withdraw: its work is not abandoned, only its
/// holder is gone. The row shows the holder's heartbeat (`last_seen_at`, `silent_secs`).
pub fn rows(
    facts: &StaleFacts,
    proofs: &BTreeMap<i64, Proof>,
    now: i64,
    days: u64,
    grace_secs: i64,
) -> Vec<Value> {
    let mut out: BTreeMap<i64, Value> = BTreeMap::new();
    for c in &facts.open {
        let Some(proof) = proofs.get(&c.card_id) else {
            continue;
        };
        let name = input::card_name(c.card_id);
        let (command, extra) = match proof {
            Proof::Head { head } => (
                format!("moai card finish {name} --head {head} --by leader"),
                json!({"via": "landed_head", "head": head}),
            ),
            Proof::MergedPr { pr, merge_ref } => (
                format!("moai card finish {name} --by leader"),
                json!({"via": "merged_pull_request", "pr": pr, "merge_ref": merge_ref}),
            ),
        };
        let mut v = row(
            c.card_id,
            &c.title,
            &c.state,
            &c.priority,
            "landed_open",
            command,
        );
        merge(&mut v, &extra);
        out.entry(c.card_id).or_insert(v);
    }
    for c in &facts.lease_ended {
        let name = input::card_name(c.card_id);
        // Only an `active` card is swept; a card in `review` keeps the withdraw it had. The rule is
        // the sweep's own (`sweep_via`): the grace ran out, or the holder is silent
        // (SPEC-20261119-live-lease-takeover REQ-11).
        let via = if c.state == "active" {
            sweep_via(c.lease_until, c.last_seen_at, grace_secs, now)
        } else {
            None
        };
        let command = if via.is_some() {
            sweep_command(&name)
        } else {
            withdraw_command(&name)
        };
        let mut v = row(
            c.card_id,
            &c.title,
            &c.state,
            &c.priority,
            "lease_ended",
            command,
        );
        merge(
            &mut v,
            &json!({
                "holder": c.holder, "generation": c.generation, "lease_until": c.lease_until,
                "last_seen_at": c.last_seen_at,
                "silent_secs": now.saturating_sub(c.last_seen_at),
            }),
        );
        if let Some(via) = via {
            merge(&mut v, &json!({"via": via.code()}));
        }
        out.entry(c.card_id).or_insert(v);
    }
    for c in &facts.review_no_holder {
        let v = row(
            c.card_id,
            &c.title,
            &c.state,
            &c.priority,
            "review_no_holder",
            withdraw_command(&input::card_name(c.card_id)),
        );
        out.entry(c.card_id).or_insert(v);
    }
    for c in facts
        .drafts
        .iter()
        .filter(|c| is_old_draft(c.updated_at, now, days))
    {
        let mut v = row(
            c.card_id,
            &c.title,
            &c.state,
            &c.priority,
            "old_draft",
            withdraw_command(&input::card_name(c.card_id)),
        );
        merge(
            &mut v,
            &json!({"age_days": now.saturating_sub(c.updated_at) / DAY_SECS, "threshold_days": days}),
        );
        out.entry(c.card_id).or_insert(v);
    }
    out.into_values().collect()
}

/// Copies the keys of the object `extra` into the object `v`.
fn merge(v: &mut Value, extra: &Value) {
    if let (Some(v), Some(extra)) = (v.as_object_mut(), extra.as_object()) {
        for (k, x) in extra {
            v.insert(k.clone(), x.clone());
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_draft_is_old_at_exactly_the_threshold_and_not_one_second_before() {
        let now = 10_000_000;
        let n = 14;
        assert!(is_old_draft(now - n * DAY_SECS, now, n as u64));
        assert!(!is_old_draft(now - n * DAY_SECS + 1, now, n as u64));
        assert!(is_old_draft(now - n * DAY_SECS - 1, now, n as u64));
    }

    #[test]
    fn a_huge_threshold_never_overflows() {
        assert!(!is_old_draft(0, 1_000, u64::MAX));
    }
}
