//! Claim verdict cache (rc.2 #2391 stage 4, decision D4 — Issue #2391 comment 6771356748).
//!
//! NOT WIRED. `claim-verify --hook` still asks the forge on every guarded commit and push.
//! Replacing that live check with a local file changes the gate (the guarded session can write
//! the file), so turning the cache on is a user decision; until then this module is a pure
//! library and [`ENABLED_BY_DEFAULT`] is `false`. The values are the most conservative ones the
//! leader set for D4:
//! - an entry lives [`TTL_MS`] (300 s) and never past the claim's `lease_until`;
//! - a miss (no entry, an expired one, another claim) means "verify live", as today;
//! - only a commit may hit; a push always verifies live;
//! - only an `owner` verdict from a live verification is ever recorded.
//!
//! Entries are JSON lines, append-only; the last entry for a claim wins. Every function takes
//! the time as an argument, so no decision reads the clock.

use serde_json::{json, Value};

/// Off until the user decides D4.
pub const ENABLED_BY_DEFAULT: bool = false;
/// How long a cached `owner` verdict stands (milliseconds), capped by the claim's lease.
pub const TTL_MS: u64 = 300_000;

/// The guarded git action.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Action {
    Commit,
    Push,
}

/// What the cache answers.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Lookup {
    /// A live `owner` verdict for this claim stands until `valid_until_ms`.
    Hit { valid_until_ms: u64 },
    /// Verify live; the reason is for the report only.
    Miss(&'static str),
}

/// The line to append after a live verification, or `None` when nothing may be cached: only
/// an `owner` verdict is recorded. Valid until the earlier of `now + TTL` and the lease.
pub fn entry(
    issue: i64,
    comment_id: i64,
    verdict: &Value,
    lease_until_ms: Option<u64>,
    now_ms: u64,
) -> Option<Value> {
    if verdict["status"] != json!("owner") {
        return None;
    }
    let ttl_end = now_ms.saturating_add(TTL_MS);
    let valid_until_ms = lease_until_ms.map_or(ttl_end, |lease| lease.min(ttl_end));
    if valid_until_ms <= now_ms {
        return None;
    }
    Some(json!({
        "schema_version": 1,
        "issue": issue,
        "comment_id": comment_id,
        "verified_ms": now_ms,
        "valid_until_ms": valid_until_ms,
    }))
}

/// Looks `issue`/`comment_id` up in the cache `lines` at `now_ms`. A push always misses; the
/// last entry for the claim decides; an entry stands while `now_ms < valid_until_ms`.
pub fn lookup(lines: &str, action: Action, issue: i64, comment_id: i64, now_ms: u64) -> Lookup {
    if action == Action::Push {
        return Lookup::Miss("a push always verifies live");
    }
    let last = lines
        .lines()
        .rev()
        .filter_map(|l| serde_json::from_str::<Value>(l).ok())
        .find(|v| {
            v["schema_version"] == json!(1)
                && v["issue"] == json!(issue)
                && v["comment_id"] == json!(comment_id)
        });
    let Some(entry) = last else {
        return Lookup::Miss("no cached verdict for this claim");
    };
    match entry["valid_until_ms"].as_u64() {
        Some(until) if now_ms < until => Lookup::Hit {
            valid_until_ms: until,
        },
        Some(_) => Lookup::Miss("cached verdict expired"),
        None => Lookup::Miss("cached entry unreadable"),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn owner() -> Value {
        json!({"status": "owner", "mine": 5})
    }

    #[test]
    #[allow(clippy::assertions_on_constants)] // pins the decided D4 defaults
    fn disabled_by_default_and_ttl_is_300_seconds() {
        assert!(!ENABLED_BY_DEFAULT);
        assert_eq!(TTL_MS, 300_000);
    }

    #[test]
    fn only_an_owner_verdict_is_recorded() {
        for status in ["uncertain", "not_owner", "stale", "closed", "no_claim"] {
            assert_eq!(
                entry(7, 5, &json!({"status": status}), None, 1_000),
                None,
                "{status}"
            );
        }
        let e = entry(7, 5, &owner(), None, 1_000).unwrap();
        assert_eq!(e["valid_until_ms"], json!(301_000));
        assert_eq!(e["verified_ms"], json!(1_000));
    }

    #[test]
    fn the_lease_caps_the_ttl_and_a_past_lease_records_nothing() {
        assert_eq!(
            entry(7, 5, &owner(), Some(200_000), 1_000).unwrap()["valid_until_ms"],
            json!(200_000)
        );
        assert_eq!(
            entry(7, 5, &owner(), Some(400_000), 1_000).unwrap()["valid_until_ms"],
            json!(301_000)
        );
        assert_eq!(
            entry(7, 5, &owner(), Some(301_000), 1_000).unwrap()["valid_until_ms"],
            json!(301_000)
        );
        assert_eq!(entry(7, 5, &owner(), Some(1_000), 1_000), None);
        assert_eq!(entry(7, 5, &owner(), Some(999), 1_000), None);
    }

    #[test]
    fn a_commit_hits_until_one_ms_before_expiry() {
        let line = entry(7, 5, &owner(), None, 1_000).unwrap().to_string();
        assert_eq!(
            lookup(&line, Action::Commit, 7, 5, 300_999),
            Lookup::Hit {
                valid_until_ms: 301_000
            }
        );
        assert_eq!(
            lookup(&line, Action::Commit, 7, 5, 301_000),
            Lookup::Miss("cached verdict expired")
        );
        assert_eq!(
            lookup(&line, Action::Commit, 7, 5, 301_001),
            Lookup::Miss("cached verdict expired")
        );
    }

    #[test]
    fn a_push_always_misses() {
        let line = entry(7, 5, &owner(), None, 1_000).unwrap().to_string();
        assert_eq!(
            lookup(&line, Action::Push, 7, 5, 1_001),
            Lookup::Miss("a push always verifies live")
        );
    }

    #[test]
    fn another_claim_misses_and_the_last_entry_wins() {
        let old = entry(7, 5, &owner(), None, 1_000).unwrap().to_string();
        let newer = entry(7, 5, &owner(), Some(50_000), 40_000)
            .unwrap()
            .to_string();
        let lines = format!("{old}\nnot json\n{newer}\n");
        assert!(matches!(
            lookup(&lines, Action::Commit, 7, 6, 2_000),
            Lookup::Miss(_)
        ));
        assert!(matches!(
            lookup(&lines, Action::Commit, 8, 5, 2_000),
            Lookup::Miss(_)
        ));
        // The newer entry (valid to 50,000) decides even though the older one ran to 301,000.
        assert_eq!(
            lookup(&lines, Action::Commit, 7, 5, 49_999),
            Lookup::Hit {
                valid_until_ms: 50_000
            }
        );
        assert_eq!(
            lookup(&lines, Action::Commit, 7, 5, 50_000),
            Lookup::Miss("cached verdict expired")
        );
        assert_eq!(
            lookup("", Action::Commit, 7, 5, 1),
            Lookup::Miss("no cached verdict for this claim")
        );
    }
}
