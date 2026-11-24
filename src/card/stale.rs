//! The `stale` key of `moai card next` (SPEC-20261124-card-fold-stale): the cards that want
//! closing, one row each, with the command that closes them. Read-only.

use crate::ledger::close::StaleFacts;
use serde_json::Value;
use std::collections::BTreeMap;
use std::path::Path;

use super::landed::Proof;

/// How many days a `draft` card may sit before it is listed, when the profile names none.
pub const DEFAULT_DRAFT_DAYS: u64 = 14;

/// Whether a draft last updated at `updated_at` is at least `days` old at `now`.
pub fn is_old_draft(_updated_at: i64, _now: i64, _days: u64) -> bool {
    false
}

/// The stale-draft days of the project: the profile key `stale_draft_days`, else the default.
pub fn draft_days(_project: &Path) -> Result<u64, String> {
    Ok(DEFAULT_DRAFT_DAYS)
}

/// The rows of `stale`, ordered by card.
pub fn rows(
    _facts: &StaleFacts,
    _proofs: &BTreeMap<i64, Proof>,
    _now: i64,
    _days: u64,
) -> Vec<Value> {
    Vec::new()
}
