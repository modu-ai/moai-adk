//! The read-only tidy plan (SPEC-20261124-home-tidy-slice2, REQ-5 to REQ-8). It names each
//! candidate with an action, the rule that made it one and the reason, offers only what the
//! rules reach, and gives the digest of its candidate list, which `tidy-apply` confirms.
//!
//! What needs git proof (worktrees, build folders, evidence landed, receipt trees) is kept with
//! `proof_pending`; what no one is known to write (`logs`) is kept as unknown; another product
//! and every protected entry are kept as listed; the Go-era names of `rules::KEPT_GO_DATA` are
//! kept. The only `delete` is an empty folder of the retired-tool class.

use super::inventory::{Entry, Item, KindRow, Scan};
use super::rules::Thresholds;
use serde_json::{json, Value};

struct Planned {
    path: String,
    kind: String,
    action: &'static str,
    rule: &'static str,
    reason: String,
    bytes: u64,
    newest: u64,
}

impl Planned {
    fn json(&self) -> Value {
        json!({
            "path": self.path,
            "kind": self.kind,
            "action": self.action,
            "rule": self.rule,
            "reason": self.reason,
            "bytes": self.bytes,
            "newest": self.newest,
        })
    }
}

fn keep(
    path: &str,
    kind: &str,
    rule: &'static str,
    reason: String,
    bytes: u64,
    newest: u64,
) -> Planned {
    Planned {
        path: path.to_string(),
        kind: kind.to_string(),
        action: "keep",
        rule,
        reason,
        bytes,
        newest,
    }
}

/// The plan over a scan. `thresholds` are the ones the scan's kind rows were computed with.
pub fn build(scan: &Scan, thresholds: &Thresholds) -> (Vec<Value>, String) {
    let mut top: Vec<Planned> = Vec::new();
    let mut by_kind: Vec<Planned> = Vec::new();
    for e in &scan.entries {
        match e.kind {
            Some(kind) if e.file_type == "dir" && !e.unreadable && kind != "logs" => {
                let row = scan.kinds.iter().find(|r| r.kind == kind);
                kind_items(e, kind, row, thresholds, scan.now, &mut by_kind);
            }
            _ => top.push(entry_item(e)),
        }
    }
    // Oldest first inside each kind; the kinds keep the table order.
    by_kind.sort_by(|a, b| {
        (kind_order(&a.kind), a.newest, &a.path).cmp(&(kind_order(&b.kind), b.newest, &b.path))
    });
    let mut all = top;
    all.extend(by_kind);
    let mut text = String::new();
    for p in all.iter().filter(|p| p.action != "keep") {
        text.push_str(&format!(
            "{}\t{}\t{}\t{}\t{}\n",
            p.path, p.action, p.rule, p.bytes, p.newest
        ));
    }
    let digest = crate::digest(text.as_bytes());
    (all.iter().map(Planned::json).collect(), digest)
}

fn kind_order(kind: &str) -> usize {
    super::rules::KINDS
        .iter()
        .position(|k| *k == kind)
        .unwrap_or(usize::MAX)
}

/// A top-level entry that is not a kind folder with items of its own.
fn entry_item(e: &Entry) -> Planned {
    let (bytes, newest) = (e.stats.bytes, e.stats.newest);
    let name = e.name.to_ascii_lowercase();
    if e.class != "protected" && super::rules::KEPT_GO_DATA.contains(&name.as_str()) {
        return keep(
            &e.name,
            &e.class,
            "keep_go_run_data",
            "go_run_data_kept: the update cleanup never archives this Go-era data either".into(),
            bytes,
            newest,
        );
    }
    match e.class.as_str() {
        "retired_tool" if e.empty_dir => Planned {
            path: e.name.clone(),
            kind: "retired_tool".into(),
            action: "delete",
            rule: "retired_tool_empty",
            reason: "an empty folder of a retired tool; nothing in it is lost".into(),
            bytes,
            newest,
        },
        "retired_tool" if !e.unreadable && !e.over_limit => Planned {
            path: e.name.clone(),
            kind: "retired_tool".into(),
            action: "archive",
            rule: "retired_tool_archive",
            reason: format!(
                "left by a retired tool ({}); archived before any delete",
                e.reason
            ),
            bytes,
            newest,
        },
        "retired_tool" => keep(
            &e.name,
            "retired_tool",
            "keep_unknown",
            format!(
                "unknown: the entry could not be read in full ({})",
                e.reason
            ),
            bytes,
            newest,
        ),
        "other_product" => keep(
            &e.name,
            "other_product",
            "keep_other_product",
            "other_product: listed, never changed".into(),
            bytes,
            newest,
        ),
        "protected" => keep(
            &e.name,
            "protected",
            "keep_protected",
            format!("protected: never opened, followed or hashed ({})", e.reason),
            bytes,
            newest,
        ),
        "unknown" => keep(
            &e.name,
            "unknown",
            "keep_unknown",
            format!("unknown: {}", e.reason),
            bytes,
            newest,
        ),
        _ => keep(
            &e.name,
            "owned",
            "keep_no_rule",
            "no tidy rule for this folder in this slice".into(),
            bytes,
            newest,
        ),
    }
}

/// The items of one folder kind: only `reports` has an age rule here; the rest wait for proof.
fn kind_items(
    e: &Entry,
    kind: &str,
    row: Option<&KindRow>,
    thresholds: &Thresholds,
    now: u64,
    out: &mut Vec<Planned>,
) {
    let over = row.is_some_and(|r| !r.passed.is_empty());
    for it in &e.items {
        out.push(match kind {
            "reports" => report_item(it, over, thresholds, now),
            _ => keep(
                &it.path,
                kind,
                if over { "keep_proof_pending" } else { "keep_under_threshold" },
                if over {
                    "proof_pending: the proof that this is safe to remove comes with the worktree and landing rules".into()
                } else {
                    "under_threshold: the kind is under every threshold".into()
                },
                it.bytes,
                it.newest,
            ),
        });
    }
}

fn report_item(it: &Item, over: bool, thresholds: &Thresholds, now: u64) -> Planned {
    let limit = thresholds.get("reports", "archive_after_days");
    let age_days = now.saturating_sub(it.newest) / 86_400;
    if !over {
        return keep(
            &it.path,
            "reports",
            "keep_under_threshold",
            "under_threshold: reports are under every threshold".into(),
            it.bytes,
            it.newest,
        );
    }
    if age_days as f64 > limit {
        Planned {
            path: it.path.clone(),
            kind: "reports".into(),
            action: "archive",
            rule: "reports_age",
            reason: format!(
                "{age_days} days old, past {limit} days, while reports are over a threshold"
            ),
            bytes: it.bytes,
            newest: it.newest,
        }
    } else {
        keep(
            &it.path,
            "reports",
            "keep_not_old_enough",
            format!("not_old_enough: {age_days} days old, the rule needs more than {limit}"),
            it.bytes,
            it.newest,
        )
    }
}
