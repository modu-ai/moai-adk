//! write_scope overlap between active claims (Issue #2007, 06-C; REQ-06C-05/06).
//!
//! Two claims conflict when some path could fall under both of their write_scope entries.
//! A conflict marks the new claim `blocked(reason=scope_conflict)` unless the leader names an
//! integration order (`--integration-order <task_ref>`), in which case it is allowed and the
//! order is recorded. [`check`] computes the decision and returns the record change as
//! `record_update_request`; [`check_and_record`] also commits it to the claim journal through
//! 06-B's `Journal::commit` when the record came from the journal (`--task`).
//!
//! The overlap test is the interface 05-B's private `plan::paths_overlap` is meant to move to
//! (decision 2026-11-08 19:1x); 06-C does not edit `src/plan.rs`.
use crate::claim::{ClaimRecord, ACTIVE_STATES};
use serde_json::{json, Value};

/// Paths every claim shares by default (REQ-06C-06): two active claims that both reach one of
/// these conflict exactly like any other overlap.
pub const SHARED_INTERFACE_PATHS: &[&str] = &[
    "Cargo.lock",
    "Cargo.toml",
    ".github/**",
    ".moai-cli/manifests/**",
    "AGENTS.md",
];

fn segments(path: &str) -> Vec<&str> {
    let mut p = path.trim();
    while let Some(rest) = p.strip_prefix("./") {
        p = rest;
    }
    p.split('/')
        .filter(|s| !s.is_empty() && *s != ".")
        .collect()
}

/// `*` inside one segment matches any run of characters within that segment.
fn segment_match(pattern: &str, text: &str) -> bool {
    let parts: Vec<&str> = pattern.split('*').collect();
    if parts.len() == 1 {
        return pattern == text;
    }
    let (first, last) = (parts[0], parts[parts.len() - 1]);
    if !text.starts_with(first) || text.len() < first.len() + last.len() || !text.ends_with(last) {
        return false;
    }
    let mut rest = &text[first.len()..text.len() - last.len()];
    for mid in &parts[1..parts.len() - 1] {
        match rest.find(mid) {
            Some(i) => rest = &rest[i + mid.len()..],
            None => return false,
        }
    }
    true
}

fn segments_compatible(a: &str, b: &str) -> bool {
    match (a.contains('*'), b.contains('*')) {
        (false, false) => a == b,
        (true, false) => segment_match(a, b),
        (false, true) => segment_match(b, a),
        // Two wildcard segments: assume some name matches both (conservative).
        (true, true) => true,
    }
}

// @MX:ANCHOR: [AUTO] write_scope overlap interface owned by 06-C #2007
// @MX:REASON: scope-check, the shared interface list and (later) 05-B plan admission all decide
//   ownership conflicts through this one predicate; changing it changes who may write where
// @MX:DEBT: [AUTO] std-only matching: exact path, directory prefix, `**` (anything below) and
//   `*` within one segment; no character classes, braces or escapes (Cargo.toml additions are
//   forbidden for this Issue, so no glob crate)
// @MX:CEILING: repository-relative paths; symlinks and case-insensitive file systems are not seen
// @MX:UPGRADE: adopt a standard glob matcher once a dependency decision allows one
/// True when some path could be written under both `a` and `b`. A plain path also covers
/// everything below it (as 05-B's prefix rule does), and `**` covers everything below.
pub fn overlap(a: &str, b: &str) -> bool {
    let (sa, sb) = (segments(a), segments(b));
    for (x, y) in sa.iter().zip(sb.iter()) {
        if *x == "**" || *y == "**" {
            return true;
        }
        if !segments_compatible(x, y) {
            return false;
        }
    }
    true
}

fn is_shared(path: &str) -> bool {
    SHARED_INTERFACE_PATHS.iter().any(|s| overlap(s, path))
}

/// Every pair of overlapping entries, in input order.
pub fn overlaps(mine: &[String], theirs: &[String]) -> Vec<Value> {
    let mut out = Vec::new();
    for m in mine {
        for t in theirs {
            if overlap(m, t) {
                out.push(json!({
                    "mine": m,
                    "theirs": t,
                    "shared_interface": is_shared(m) && is_shared(t),
                }));
            }
        }
    }
    out
}

/// The scope decision for `mine` against every other active claim in `records`.
/// `integration_order` is the task the leader put first; a conflict with it is allowed.
pub fn check(
    mine: &ClaimRecord,
    records: &[ClaimRecord],
    integration_order: Option<&str>,
) -> Value {
    let mut conflicts = Vec::new();
    for other in records {
        if other.task_key == mine.task_key || !ACTIVE_STATES.contains(&other.state.as_str()) {
            continue;
        }
        let overlap = overlaps(&mine.write_scope, &other.write_scope);
        if overlap.is_empty() {
            continue;
        }
        let ordered = integration_order.is_some_and(|o| {
            o == other.task_key
                || o == other.logical_key
                || o == format!("#{}", other.task_ref.issue_number)
        });
        conflicts.push(json!({
            "scope_conflict": {
                "other_task": other.task_key,
                "other_logical_key": other.logical_key,
                "other_issue_number": other.task_ref.issue_number,
                "overlap": overlap,
            },
            "allowed_by_integration_order": ordered,
        }));
    }
    let blocked = conflicts
        .iter()
        .any(|c| c["allowed_by_integration_order"] == false);
    let decision = if blocked {
        "blocked"
    } else if conflicts.is_empty() {
        "clear"
    } else {
        "allowed_with_order"
    };
    let request = if blocked {
        json!({
            "task_key": mine.task_key,
            "epoch": mine.epoch,
            "state": "blocked",
            "reason": "scope_conflict",
        })
    } else if conflicts.is_empty() {
        Value::Null
    } else {
        json!({
            "task_key": mine.task_key,
            "epoch": mine.epoch,
            "integration_order": integration_order,
        })
    };
    json!({
        "decision": decision,
        "task_key": mine.task_key,
        "conflicts": conflicts,
        "shared_interface_paths": SHARED_INTERFACE_PATHS,
        "record_update_request": request,
        "record_update_owner": crate::worktree::RECORD_UPDATE_OWNER,
    })
}

/// [`check`] plus the record write: `blocked` with a `scope_conflict` effect, or an
/// `integration_order` effect when the leader's order allowed the overlap. A repeated
/// check with the same outcome adds no event.
pub fn check_and_record<C: crate::forge::policy::Clock>(
    project: &std::path::Path,
    mine: &ClaimRecord,
    from_journal: bool,
    records: &[ClaimRecord],
    integration_order: Option<&str>,
    clock: &C,
) -> std::result::Result<Value, crate::contract::OpError> {
    use crate::worktree::{commit_update, not_written, open_locked_journal};
    let mut out = check(mine, records, integration_order);
    let effect = match out["decision"].as_str() {
        Some("blocked") => Some(json!({"kind": "scope_conflict", "conflicts": out["conflicts"]})),
        Some("allowed_with_order") => {
            Some(json!({"kind": "integration_order", "order": integration_order}))
        }
        _ => None,
    };
    out["record_update"] = match (effect, from_journal) {
        (None, _) => json!({"applied": false, "reason": "no scope conflict"}),
        (Some(_), false) => not_written(),
        (Some(effect), true) => {
            let blocked = out["decision"] == "blocked";
            let journal = open_locked_journal(project)?;
            commit_update(
                &journal,
                &mine.task_key,
                mine.epoch,
                "scope_check",
                clock,
                |r| {
                    if blocked {
                        r.state = "blocked".into();
                    }
                    if r.effects.last() != Some(&effect) {
                        r.effects.push(effect);
                    }
                },
            )?
        }
    };
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::overlap;

    #[test]
    fn overlap_rules() {
        assert!(overlap("src/lib.rs", "src/lib.rs"));
        assert!(overlap("src", "src/lib.rs"));
        assert!(overlap("./src/", "src/a.rs"));
        assert!(overlap(".github/**", ".github/workflows/rc-ci.yml"));
        assert!(overlap("src/*.rs", "src/lib.rs"));
        assert!(overlap("docs/**", "docs/a/*.md"));
        assert!(!overlap("src/lib.rs", "src/main.rs"));
        assert!(!overlap("src/*.rs", "src/lib.md"));
        assert!(!overlap("docs/a.md", "src/a.md"));
        assert!(!overlap("src/lib.rs", "src/lib.rs.bak"));
    }
}
