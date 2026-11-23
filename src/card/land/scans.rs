//! The scan check of `card land` (SPEC-20261123-land-repo-scans): before the landing moves the
//! integration branch it works out, from `tools/repo_scans.json`, which repo-wide scan tests the
//! merge tree needs ([`crate::repo_scans::needed`]), and asks for their receipt: a step
//! `scan:<id>` at exit 0 in the receipt of a PASS verdict that names the landing head.
//!
//! The mode is the one of the verdict check (`work_flow.land_verdict`, [`super::check::mode`]):
//! observe reports the refusal as `scans.would_refuse` and lands, enforce refuses. The check reads
//! the table, git and the verdict events and writes nothing.

use super::check::{Input, Mode};
use crate::ledger::workflow;
use crate::repo_scans::{self, Needed};
use serde_json::{json, Value};

type Result<T> = std::result::Result<T, String>;

/// The code of a refusal for scans that no receipt covers.
pub const MISSING: &str = "scans_missing";
/// The code of a refusal when the table or git could not be read.
pub const UNREADABLE: &str = "scans_not_read";

/// The receipt step that covers the scan `id`.
pub fn step_name(id: &str) -> String {
    format!("scan:{id}")
}

/// Whether a check line of the table is a cargo command: tests are Rust only, so the landing asks
/// for no other kind of check (a Python line stays in the table for `moai repo-scans`).
fn is_cargo_check(check: &str) -> bool {
    check.split_whitespace().next() == Some("cargo")
}

/// The scans the landing asks for, from the scans the merge tree needs: each keeps its cargo
/// checks only, and a scan with none is not asked for (SPEC-20261123-land-repo-scans REQ-15).
fn cargo_only(needed: Vec<Needed>) -> Vec<Needed> {
    needed
        .into_iter()
        .map(|mut need| {
            need.checks.retain(|c| is_cargo_check(c));
            need
        })
        .filter(|need| !need.checks.is_empty())
        .collect()
}

/// The one line that tells what to run and what to record for the scans not covered.
fn missing_line(missing: &[&Needed]) -> String {
    let parts: Vec<String> = missing
        .iter()
        .map(|n| {
            format!(
                "{} (run `{}` on the merge tree, then record the step `{}` with its exit code)",
                n.id,
                n.checks.join(" && "),
                step_name(&n.id)
            )
        })
        .collect();
    format!(
        "the merge tree needs these repo-wide scan tests and no PASS receipt of this head covers \
         them: {}",
        parts.join("; ")
    )
}

/// The answer's `scans` object and, in enforce mode, the refusal text.
///
/// The object is `{required: [{id, checks, step, covered}], missing: [id], would_refuse}`;
/// `would_refuse` is `{code, why}` or null. Enforce mode answers `Err("<code>: <why>")` instead of
/// a would-refuse.
pub fn read(input: &Input<'_>, mode: Mode) -> Result<Value> {
    let needed = match repo_scans::needed(input.project, input.tip, input.head) {
        Ok(needed) => cargo_only(needed),
        Err(why) => {
            return refuse(
                mode,
                UNREADABLE,
                format!("the scan table could not be read: {why}"),
                json!([]),
                json!([]),
            )
        }
    };
    let covered_steps: Vec<String> = match workflow::verdict_events(input.conn, input.card_id) {
        Ok(events) => events
            .iter()
            .filter(|e| e.pass && e.head == input.head)
            .filter_map(|e| e.receipt.as_ref())
            .filter(|r| r.head == input.head)
            .flat_map(|r| r.steps.iter())
            .filter(|s| s.exit == 0)
            .map(|s| s.name.clone())
            .collect(),
        Err(e) => {
            return refuse(
                mode,
                UNREADABLE,
                format!("the verdict events could not be read: {e}"),
                json!([]),
                json!([]),
            )
        }
    };
    let mut required = Vec::new();
    let mut missing: Vec<&Needed> = Vec::new();
    for need in &needed {
        let step = step_name(&need.id);
        let covered = covered_steps.contains(&step);
        if !covered {
            missing.push(need);
        }
        required.push(json!({
            "id": need.id, "checks": need.checks, "step": step, "covered": covered,
        }));
    }
    let missing_ids: Vec<&str> = missing.iter().map(|n| n.id.as_str()).collect();
    if missing.is_empty() {
        return Ok(json!({"required": required, "missing": [], "would_refuse": null}));
    }
    refuse(
        mode,
        MISSING,
        missing_line(&missing),
        Value::Array(required),
        json!(missing_ids),
    )
}

fn refuse(mode: Mode, code: &str, why: String, required: Value, missing: Value) -> Result<Value> {
    match mode {
        Mode::Enforce => Err(format!("{code}: {why}")),
        Mode::Observe => Ok(json!({
            "required": required, "missing": missing,
            "would_refuse": {"code": code, "why": why},
        })),
    }
}
