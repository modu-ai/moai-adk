//! `optimize.plan` (Issue #2724, parent #2380): a read-only preview of what space the
//! repository's leftovers hold and which of them could go — linked worktrees after their
//! merge, cargo build folders outside the worktrees (`/tmp/moai-rc-target-*` and the like,
//! only under roots the caller names), and reports. One table: size, last change, retention,
//! the reasons a row is kept, and how to get it back.
//!
//! Nothing is deleted or written. The worktree rows reuse `worktree.gc.plan` with its
//! squash-merge test turned on ([`crate::worktree_gc::plan_with`]); on top of that a row is
//! a `remove` candidate only with independent-audit evidence for its exact HEAD, and any
//! non-`skip` row needs the caller's observation of live claims and HOLD Issues (the
//! #2669 rule: what the local repository cannot tell comes from an observations file with
//! its source, and a part that was not observed keeps the row). Time alone never makes a
//! worktree a candidate: retention is shown, not acted on.
//!
//! Applying stays where it was (#2389 comment 6774899518 item 2: the user confirms first).
//! `worktree.gc.apply` recomputes its own plan, where only an ancestor of the integration tip
//! counts as merged, so a squash-merged worktree stays `skip` there; folders outside the
//! worktrees and reports have no apply step in #2724.

use crate::worktree_gc;
use crate::worktree_report::{bytes_under, parse_worktree_list};
use serde::Deserialize;
use serde_json::{json, Value};
use std::collections::BTreeSet;
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

type Result<T> = std::result::Result<T, String>;

/// Default retention of a build folder outside the worktrees, in days.
pub const DEFAULT_RETENTION_DAYS: u64 = 7;

/// The observations file (typed JSON, unknown keys rejected). Every part is optional; a part
/// that is missing keeps the rows that need it.
#[derive(Clone, Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Observations {
    pub schema_version: u32,
    #[serde(default)]
    pub audits: Option<Audits>,
    #[serde(default)]
    pub claims: Option<Claims>,
    #[serde(default)]
    pub holds: Option<Holds>,
}

/// Independent-audit verdicts read from the forge.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Audits {
    pub source: String,
    pub items: Vec<AuditItem>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AuditItem {
    pub pr: u64,
    /// The 40-hex head the verdict names.
    pub head: String,
    pub verdict_comment: u64,
    /// `PASS` or `FAIL`.
    pub verdict: String,
}

/// Live claims (any session) read from the forge.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Claims {
    pub source: String,
    pub items: Vec<ClaimItem>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ClaimItem {
    pub issue: u64,
    pub claim_comment: u64,
}

/// Issues on HOLD or `needs:decision`, read from the forge.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Holds {
    pub source: String,
    pub items: Vec<HoldItem>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct HoldItem {
    pub issue: u64,
}

/// The inputs of one preview.
pub struct Request<'a> {
    pub project: &'a Path,
    pub integration_ref: &'a str,
    pub recent: Duration,
    pub budget: Duration,
    pub now_unix: Option<u64>,
    pub observations: &'a Observations,
    pub build_roots: &'a [PathBuf],
    pub retention_days: u64,
}

/// `issue/<n>-…` → `n`.
pub fn branch_issue(branch: &str) -> Option<u64> {
    let rest = branch.strip_prefix("issue/")?;
    let digits: String = rest.chars().take_while(|c| c.is_ascii_digit()).collect();
    (!digits.is_empty() && rest[digits.len()..].starts_with('-'))
        .then(|| digits.parse().ok())
        .flatten()
}

fn is_hex40(s: &str) -> bool {
    s.len() == 40
        && s.bytes()
            .all(|b| b.is_ascii_hexdigit() && !b.is_ascii_uppercase())
}

fn validate(obs: &Observations) -> Result<()> {
    if obs.schema_version != 1 {
        return Err("invalid observations: schema_version must be 1".into());
    }
    let sources = [
        obs.audits.as_ref().map(|a| a.source.as_str()),
        obs.claims.as_ref().map(|c| c.source.as_str()),
        obs.holds.as_ref().map(|h| h.source.as_str()),
    ];
    if sources.iter().flatten().any(|s| s.trim().is_empty()) {
        return Err("invalid observations: every part must name its source".into());
    }
    for a in obs.audits.iter().flat_map(|a| &a.items) {
        if !is_hex40(&a.head) {
            return Err(format!(
                "invalid observations: audit head `{}` must be 40 lowercase hex",
                a.head
            ));
        }
        if a.verdict != "PASS" && a.verdict != "FAIL" {
            return Err(format!(
                "invalid observations: audit verdict `{}` must be PASS or FAIL",
                a.verdict
            ));
        }
    }
    Ok(())
}

fn unix_secs(t: SystemTime) -> Option<u64> {
    t.duration_since(UNIX_EPOCH).ok().map(|d| d.as_secs())
}

fn mtime(p: &Path) -> Option<u64> {
    fs::symlink_metadata(p)
        .ok()
        .and_then(|m| m.modified().ok())
        .and_then(unix_secs)
}

/// The extra checks on one worktree row of the gc plan. Returns the new action, the added
/// reasons and the audit evidence that held.
fn judge_worktree(
    action: &str,
    branch: Option<&str>,
    head: Option<&str>,
    has_trim: bool,
    obs: &Observations,
) -> (String, Vec<&'static str>, Option<Value>) {
    let mut added = Vec::new();
    if action == "skip" {
        return (action.into(), added, None);
    }
    let issue = branch.and_then(branch_issue);
    match &obs.claims {
        None => added.push("claims_unobserved"),
        Some(c) if issue.is_some_and(|n| c.items.iter().any(|i| i.issue == n)) => {
            added.push("live_claim")
        }
        Some(_) => {}
    }
    match &obs.holds {
        None => added.push("holds_unobserved"),
        Some(h) if issue.is_some_and(|n| h.items.iter().any(|i| i.issue == n)) => {
            added.push("hold")
        }
        Some(_) => {}
    }
    if !added.is_empty() {
        return ("skip".into(), added, None);
    }
    if action != "remove" {
        return (action.into(), added, None);
    }
    let evidence = obs.audits.as_ref().and_then(|a| {
        a.items
            .iter()
            .find(|i| Some(i.head.as_str()) == head && i.verdict == "PASS")
            .map(|i| {
                json!({"source": a.source, "pr": i.pr, "head": i.head, "verdict_comment": i.verdict_comment})
            })
    });
    if evidence.is_some() {
        return ("remove".into(), added, evidence);
    }
    added.push(if obs.audits.is_none() {
        "audits_unobserved"
    } else {
        "audit_unconfirmed"
    });
    // Without audit evidence the worktree stays; its cargo output is regenerable.
    let next = if has_trim { "trim" } else { "skip" };
    (next.into(), added, None)
}

/// Newest modification time among a build folder, its cargo entries and their `deps` and
/// `incremental` folders.
fn newest_build_change(folder: &Path, cargo: &[PathBuf]) -> Option<u64> {
    let mut paths = vec![folder.to_path_buf()];
    for c in cargo {
        paths.push(c.clone());
        paths.push(c.join("deps"));
        paths.push(c.join("incremental"));
    }
    paths.iter().filter_map(|p| mtime(p)).max()
}

/// Reasons a build folder outside the worktrees is kept; empty means it is a candidate.
pub fn decide_build_folder(
    has_cargo: bool,
    other_entries: usize,
    in_use: Option<bool>,
    age_secs: Option<u64>,
    recent_secs: u64,
    retention_secs: u64,
) -> Vec<&'static str> {
    let mut r = Vec::new();
    if !has_cargo {
        r.push("not_cargo_output");
    }
    if other_entries > 0 {
        r.push("other_files_in_folder");
    }
    match in_use {
        Some(false) => {}
        Some(true) => r.push("process_in_use"),
        None => r.push("process_check_unavailable"),
    }
    match age_secs {
        None => r.push("change_age_unknown"),
        Some(a) if a < recent_secs => r.push("recent_change"),
        Some(a) if a < retention_secs => r.push("within_retention"),
        Some(_) => {}
    }
    r
}

fn build_folder_rows(
    root: &Path,
    now: u64,
    recent_secs: u64,
    retention_secs: u64,
    cwds: &Option<Vec<PathBuf>>,
    deadline: Instant,
) -> Result<Vec<Value>> {
    if !root.is_absolute() {
        return Err(format!(
            "invalid --build-root `{}`: must be an absolute path",
            root.display()
        ));
    }
    match fs::symlink_metadata(root) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => {
            return Err(format!(
                "invalid --build-root `{}`: not a real folder (a symlink or a file is refused)",
                root.display()
            ))
        }
        Err(e) => return Err(format!("build root unavailable: {}: {e}", root.display())),
    }
    let mut rows = Vec::new();
    let rd = fs::read_dir(root)
        .map_err(|e| format!("build root unreadable: {}: {e}", root.display()))?;
    let mut entries: Vec<PathBuf> = rd.flatten().map(|e| e.path()).collect();
    entries.sort();
    for path in entries {
        let meta = match fs::symlink_metadata(&path) {
            Ok(m) => m,
            Err(_) => continue,
        };
        if !meta.is_dir() {
            continue;
        }
        let (cargo, other) = worktree_gc::split_target(&path);
        let modified = newest_build_change(&path, &cargo);
        let age = modified.map(|t| now.saturating_sub(t));
        let in_use = cwds.as_ref().map(|list| {
            let canon = fs::canonicalize(&path).unwrap_or_else(|_| path.clone());
            list.iter().any(|c| c.starts_with(&canon))
        });
        let reasons = decide_build_folder(
            !cargo.is_empty(),
            other.len(),
            in_use,
            age,
            recent_secs,
            retention_secs,
        );
        let action = if reasons.is_empty() { "trim" } else { "skip" };
        let (bytes, partial) = bytes_under(std::slice::from_ref(&path), deadline);
        rows.push(json!({
            "kind": "build_folder",
            "path": path.to_string_lossy(),
            "action": action,
            "reasons": reasons,
            "trim_paths": cargo.iter().map(|c| c.file_name().map(|n| n.to_string_lossy().into_owned()).unwrap_or_default()).collect::<Vec<_>>(),
            "preserved": other,
            "bytes": bytes,
            "bytes_partial": partial,
            "modified_unix": modified,
            "age_secs": age,
            "retention_until_unix": modified.map(|t| t + retention_secs),
            "restore": "cargo rebuilds it: run the build again with CARGO_TARGET_DIR set to this folder",
            "apply": "none in #2724 (preview only)",
        }));
    }
    Ok(rows)
}

/// Builds the preview. Reads only.
pub fn plan(req: &Request) -> Result<Value> {
    validate(req.observations)?;
    if req.retention_days == 0 {
        return Err("invalid --retention-days: must be at least 1".into());
    }
    let deadline = Instant::now() + req.budget;
    let gc = worktree_gc::plan_with(
        req.project,
        req.integration_ref,
        req.recent,
        req.budget,
        req.now_unix,
        true,
    )?;
    let now = gc["now_unix"].as_u64().ok_or("gc plan without now_unix")?;
    let recent_secs = req.recent.as_secs();
    let retention_secs = req.retention_days.saturating_mul(86_400);
    let listing = worktree_gc::git_out(req.project, &["worktree", "list", "--porcelain"])
        .ok_or("git failed: worktree list")?;
    let primary = parse_worktree_list(&listing)
        .first()
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .ok_or("git failed: no worktree listed")?;

    let mut rows = Vec::new();
    for c in gc["candidates"].as_array().into_iter().flatten() {
        let rel = c["path"].as_str().unwrap_or_default();
        let abs = if rel == "." {
            primary.clone()
        } else if Path::new(rel).is_absolute() {
            PathBuf::from(rel)
        } else {
            primary.join(rel)
        };
        let trim_paths = c["trim_paths"].as_array().cloned().unwrap_or_default();
        let (action, added, evidence) = judge_worktree(
            c["action"].as_str().unwrap_or("skip"),
            c["branch"].as_str(),
            c["head"].as_str(),
            !trim_paths.is_empty(),
            req.observations,
        );
        let mut reasons: Vec<Value> = c["reasons"].as_array().cloned().unwrap_or_default();
        reasons.extend(added.iter().map(|r| json!(r)));
        let (bytes, partial) = bytes_under(std::slice::from_ref(&abs), deadline);
        let modified = c["change_age_secs"].as_u64().map(|a| now.saturating_sub(a));
        let restore = match action.as_str() {
            "remove" => json!(format!(
                "git worktree add {} {} — the branch is kept; ignored files and build output are not restored",
                abs.display(),
                c["branch"].as_str().unwrap_or("<branch>")
            )),
            "trim" => json!("cargo rebuilds target/ on the next build"),
            _ => Value::Null,
        };
        rows.push(json!({
            "kind": "worktree",
            "path": rel,
            "branch": c["branch"],
            "head": c["head"],
            "merged_by": c["merged_by"],
            "action": action,
            "reasons": reasons,
            "audit_evidence": evidence,
            "trim_paths": trim_paths,
            "preserved": c["preserved_in_target"],
            "bytes": bytes,
            "bytes_partial": partial,
            "trim_bytes": if action == "skip" { json!(0) } else { c["trim_bytes"].clone() },
            "modified_unix": modified,
            "age_secs": c["change_age_secs"],
            "retention_until_unix": modified.map(|t| t + retention_secs),
            "restore": restore,
            "apply": "worktree-gc-apply with its own plan digest (ancestor merges only), after the user confirms",
        }));
    }

    let cwds = worktree_gc::process_cwds().map(|v| {
        v.into_iter()
            .map(|p| fs::canonicalize(&p).unwrap_or(p))
            .collect::<Vec<_>>()
    });
    let mut seen = BTreeSet::new();
    for root in req.build_roots {
        if !seen.insert(root.clone()) {
            continue;
        }
        rows.extend(build_folder_rows(
            root,
            now,
            recent_secs,
            retention_secs,
            &cwds,
            deadline,
        )?);
    }

    let reports = primary.join(".moai-cli/reports");
    if fs::symlink_metadata(&reports).is_ok_and(|m| m.is_dir()) {
        let (bytes, partial) = bytes_under(std::slice::from_ref(&reports), deadline);
        let modified = mtime(&reports);
        rows.push(json!({
            "kind": "report",
            "path": ".moai-cli/reports",
            "action": "keep",
            "reasons": ["reports_kept"],
            "trim_paths": [],
            "bytes": bytes,
            "bytes_partial": partial,
            "modified_unix": modified,
            "age_secs": modified.map(|t| now.saturating_sub(t)),
            "retention_until_unix": Value::Null,
            "restore": "not regenerable — reports are kept",
            "apply": "none",
        }));
    }

    let bound: Vec<Value> = rows
        .iter()
        .map(|r| json!({"kind": r["kind"], "path": r["path"], "head": r["head"], "action": r["action"], "reasons": r["reasons"], "trim_paths": r["trim_paths"]}))
        .collect();
    let digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    let count = |a: &str| rows.iter().filter(|r| r["action"] == a).count();
    let sum = |f: &dyn Fn(&Value) -> u64| rows.iter().map(f).sum::<u64>();
    let freeable = sum(&|r: &Value| match r["action"].as_str() {
        Some("remove") => r["bytes"].as_u64().unwrap_or(0),
        Some("trim") if r["kind"] == "worktree" => r["trim_bytes"].as_u64().unwrap_or(0),
        Some("trim") => r["bytes"].as_u64().unwrap_or(0),
        _ => 0,
    });
    Ok(json!({
        "schema_version": 1,
        "integration_ref": req.integration_ref,
        "integration_tip": gc["integration_tip"],
        "recent_minutes": recent_secs / 60,
        "retention_days": req.retention_days,
        "now_unix": now,
        "observations": {
            "audits": req.observations.audits.as_ref().map(|a| &a.source),
            "claims": req.observations.claims.as_ref().map(|c| &c.source),
            "holds": req.observations.holds.as_ref().map(|h| &h.source),
        },
        "totals": {
            "rows": rows.len(),
            "remove": count("remove"),
            "trim": count("trim"),
            "skip": count("skip"),
            "keep": count("keep"),
            "bytes": sum(&|r: &Value| r["bytes"].as_u64().unwrap_or(0)),
            "freeable_bytes": freeable,
        },
        "rows": rows,
        "preview_digest": digest,
        "apply": "optimize-plan applies nothing. Worktree rows: worktree-gc-apply recomputes its own plan, where only an ancestor of the integration tip counts as merged, and runs only after the user confirms (#2389 comment 6774899518 item 2). Build folders outside the worktrees and reports: no apply step in #2724.",
        "network": "disabled",
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn obs(claims: &[u64], holds: &[u64], audits: Option<Vec<(&str, &str)>>) -> Observations {
        Observations {
            schema_version: 1,
            audits: audits.map(|v| Audits {
                source: "test".into(),
                items: v
                    .into_iter()
                    .map(|(h, verdict)| AuditItem {
                        pr: 1,
                        head: h.into(),
                        verdict_comment: 2,
                        verdict: verdict.into(),
                    })
                    .collect(),
            }),
            claims: Some(Claims {
                source: "test".into(),
                items: claims
                    .iter()
                    .map(|&issue| ClaimItem {
                        issue,
                        claim_comment: 3,
                    })
                    .collect(),
            }),
            holds: Some(Holds {
                source: "test".into(),
                items: holds.iter().map(|&issue| HoldItem { issue }).collect(),
            }),
        }
    }

    const H: &str = "0123456789abcdef0123456789abcdef01234567";

    #[test]
    fn remove_needs_a_pass_for_the_exact_head() {
        let b = Some("issue/7-x");
        let pass = obs(&[], &[], Some(vec![(H, "PASS")]));
        assert_eq!(
            judge_worktree("remove", b, Some(H), true, &pass).0,
            "remove"
        );
        let fail = obs(&[], &[], Some(vec![(H, "FAIL")]));
        let (a, r, e) = judge_worktree("remove", b, Some(H), true, &fail);
        assert_eq!(
            (a.as_str(), r, e),
            ("trim", vec!["audit_unconfirmed"], None)
        );
        let other = obs(&[], &[], Some(vec![("f".repeat(40).leak(), "PASS")]));
        assert_eq!(
            judge_worktree("remove", b, Some(H), false, &other).0,
            "skip"
        );
        let none = obs(&[], &[], None);
        assert_eq!(
            judge_worktree("remove", b, Some(H), true, &none).1,
            vec!["audits_unobserved"]
        );
    }

    #[test]
    fn a_live_claim_a_hold_or_a_missing_observation_keeps_the_row() {
        let b = Some("issue/7-x");
        let p = Some(vec![(H, "PASS")]);
        assert_eq!(
            judge_worktree("trim", b, Some(H), true, &obs(&[7], &[], p.clone())).1,
            vec!["live_claim"]
        );
        assert_eq!(
            judge_worktree("remove", b, Some(H), true, &obs(&[], &[7], p.clone())).0,
            "skip"
        );
        let mut o = obs(&[], &[], p);
        o.claims = None;
        o.holds = None;
        assert_eq!(
            judge_worktree("trim", b, Some(H), true, &o).1,
            vec!["claims_unobserved", "holds_unobserved"]
        );
    }

    #[test]
    fn build_folder_retention_and_recent_edges() {
        let (recent, keep) = (7200, 7 * 86_400);
        let d = |age| decide_build_folder(true, 0, Some(false), Some(age), recent, keep);
        assert_eq!(d(recent - 1), vec!["recent_change"]);
        assert_eq!(d(keep - 1), vec!["within_retention"]);
        assert!(d(keep).is_empty());
        assert_eq!(
            decide_build_folder(false, 1, None, None, recent, keep),
            vec![
                "not_cargo_output",
                "other_files_in_folder",
                "process_check_unavailable",
                "change_age_unknown"
            ]
        );
    }

    #[test]
    fn observations_are_validated() {
        let mut o = obs(&[], &[], Some(vec![("ABC", "PASS")]));
        assert!(validate(&o).unwrap_err().contains("40 lowercase hex"));
        o.audits = Some(Audits {
            source: " ".into(),
            items: vec![],
        });
        assert!(validate(&o).unwrap_err().contains("source"));
        o.audits = None;
        o.schema_version = 2;
        assert!(validate(&o).is_err());
    }

    #[test]
    fn branch_issue_reads_only_issue_branches() {
        assert_eq!(branch_issue("issue/2724-optimize"), Some(2724));
        assert_eq!(branch_issue("issue/2724"), None);
        assert_eq!(branch_issue("feature/2724-x"), None);
    }
}
