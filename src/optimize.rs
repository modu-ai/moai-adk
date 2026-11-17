//! `optimize.plan` (Issue #2724, parent #2380): a read-only preview of what space the
//! repository's leftovers hold and which of them could go — linked worktrees after their
//! merge, cargo build folders outside the worktrees (`/tmp/moai-rc-target-*` and the like,
//! only under roots the caller names), the agent build folders under `<HOME>/.moai/build/`
//! (C-82: a candidate only when its owner record proves the worktree gone or merged), and
//! reports. One table: size, last change, retention, the reasons a row is kept, and how to get
//! it back.
//!
//! Nothing is deleted or written. `worktrees.count` is the repository's worktree count, and
//! `warnings` names it when it is above [`WORKTREE_WARN_OVER`] (#2930 item 6). The worktree rows reuse `worktree.gc.plan` with its
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

/// The retention the design gives the hook log (#3009 §6, "7 days or 64 MiB"). The preview shows
/// whether a log is past it ([`hook_log_row`]); nothing is rotated, trimmed or deleted by it.
pub const HOOK_LOG_RETENTION_DAYS: u64 = 7;
pub const HOOK_LOG_RETENTION_BYTES: u64 = 64 * 1024 * 1024;

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
    /// The user's HOME, where a project with a fixed id keeps its hook log (#3009 S4b); `None`
    /// lists only the project's own log.
    pub home: Option<&'a Path>,
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

/// C-82: the owner record `tools/build_folder.py env` writes into an agent build folder
/// (`<HOME>/.moai/build/<project>-<worktree>/`): `schema_version` 1, the repository's
/// `git_common_dir` and the `worktree` the folder builds, both absolute.
pub const AGENT_BUILD_RECORD: &str = ".moai-build.json";

/// The longest owner record read; a longer file counts as no record.
const AGENT_BUILD_RECORD_MAX: u64 = 4096;

/// The owner of an agent build folder, from its record, or why it is unknown.
fn agent_build_owner(dir: &Path) -> Option<(PathBuf, PathBuf)> {
    let rec = dir.join(AGENT_BUILD_RECORD);
    let meta = fs::symlink_metadata(&rec).ok()?;
    if !meta.is_file() || meta.len() > AGENT_BUILD_RECORD_MAX {
        return None;
    }
    let v: Value = serde_json::from_slice(&fs::read(&rec).ok()?).ok()?;
    if v["schema_version"] != 1 {
        return None;
    }
    let common = PathBuf::from(v["git_common_dir"].as_str()?);
    let worktree = PathBuf::from(v["worktree"].as_str()?);
    (common.is_absolute() && worktree.is_absolute()).then_some((common, worktree))
}

/// C-82 (user decision 2026-11-17, "빌드 폴더를 한 곳으로 모으기"): rows for the agent build
/// folders under `<home>/.moai/build/`. A folder is a `trim` candidate only when its owner record
/// names this repository and a worktree that is gone (not in `git worktree list`) or whose branch
/// the gc plan counts as merged, and the folder holds only cargo output and the record. A link is
/// never followed. Returns the rows and the warning lines.
#[allow(clippy::too_many_arguments)]
fn agent_build_rows(
    home: Option<&Path>,
    common: Option<&Path>,
    listed: &BTreeSet<PathBuf>,
    merged: &BTreeSet<PathBuf>,
    now: u64,
    cwds: &Option<Vec<PathBuf>>,
    deadline: Instant,
) -> (Vec<Value>, Vec<String>) {
    let (mut rows, mut warnings) = (Vec::new(), Vec::new());
    let Some(home) = home else {
        return (rows, warnings);
    };
    // Audit B1: every step below HOME is read without following a link, `.moai` included.
    let moai = home.join(".moai");
    match fs::symlink_metadata(&moai) {
        Ok(m) if m.is_dir() => {}
        Ok(m) if m.file_type().is_symlink() => {
            warnings.push(format!(
                "{} is a link; agent build folders under it are not listed (C-82)",
                moai.display()
            ));
            return (rows, warnings);
        }
        Ok(_) => {
            warnings.push(format!(
                "{} is not a folder; no agent build folder is listed (C-82)",
                moai.display()
            ));
            return (rows, warnings);
        }
        Err(_) => return (rows, warnings),
    }
    let root = moai.join("build");
    match fs::symlink_metadata(&root) {
        Ok(m) if m.is_dir() => {}
        Ok(m) if m.file_type().is_symlink() => {
            warnings.push(format!(
                "{} is a link; agent build folders under it are not listed (C-82)",
                root.display()
            ));
            return (rows, warnings);
        }
        Ok(_) => {
            warnings.push(format!(
                "{} is not a folder; no agent build folder is listed (C-82)",
                root.display()
            ));
            return (rows, warnings);
        }
        Err(_) => return (rows, warnings),
    }
    let Ok(rd) = fs::read_dir(&root) else {
        warnings.push(format!("{} is unreadable (C-82)", root.display()));
        return (rows, warnings);
    };
    let mut entries: Vec<PathBuf> = rd.flatten().map(|e| e.path()).collect();
    entries.sort();
    let canon = |p: &Path| fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf());
    for path in entries {
        let Ok(meta) = fs::symlink_metadata(&path) else {
            continue;
        };
        let mut reasons: Vec<&str> = Vec::new();
        let mut worktree = Value::Null;
        let why = if meta.file_type().is_symlink() {
            reasons.push("symlink");
            "symlink"
        } else if !meta.is_dir() {
            continue;
        } else {
            match agent_build_owner(&path) {
                None => {
                    reasons.push("no_owner_record");
                    "no_owner_record"
                }
                Some((rec_common, rec_worktree)) => {
                    worktree = json!(rec_worktree.to_string_lossy());
                    if common.map(|c| canon(&rec_common) != c).unwrap_or(true) {
                        reasons.push("other_repository");
                        "other_repository"
                    } else if !listed.contains(&canon(&rec_worktree)) {
                        "worktree_gone"
                    } else if merged.contains(&canon(&rec_worktree)) {
                        "branch_merged"
                    } else {
                        reasons.push("worktree_live");
                        "worktree_live"
                    }
                }
            }
        };
        let (cargo, other) = if why == "symlink" {
            (Vec::new(), Vec::new())
        } else {
            let (cargo, other) = worktree_gc::split_target(&path);
            let other: Vec<String> = other
                .into_iter()
                .filter(|n| n != AGENT_BUILD_RECORD)
                .collect();
            (cargo, other)
        };
        if why != "symlink" {
            if cargo.is_empty() {
                reasons.push("not_cargo_output");
            }
            if !other.is_empty() {
                reasons.push("other_files_in_folder");
            }
            match cwds.as_ref().map(|list| {
                let c = canon(&path);
                list.iter().any(|p| p.starts_with(&c))
            }) {
                Some(false) => {}
                Some(true) => reasons.push("process_in_use"),
                None => reasons.push("process_check_unavailable"),
            }
        }
        let action = if reasons.is_empty() { "trim" } else { "skip" };
        let (bytes, partial) = if why == "symlink" {
            (0, false)
        } else {
            bytes_under(std::slice::from_ref(&path), deadline)
        };
        let modified = newest_build_change(&path, &cargo);
        rows.push(json!({
            "kind": "agent_build_folder",
            "path": path.to_string_lossy(),
            "action": action,
            "why": why,
            "reasons": reasons,
            "worktree": worktree,
            "trim_paths": cargo.iter().map(|c| c.file_name().map(|n| n.to_string_lossy().into_owned()).unwrap_or_default()).collect::<Vec<_>>(),
            "preserved": other,
            "bytes": bytes,
            "bytes_partial": partial,
            "modified_unix": modified,
            "age_secs": modified.map(|t| now.saturating_sub(t)),
            "retention_until_unix": Value::Null,
            "restore": "cargo rebuilds it on the next build with CARGO_TARGET_DIR set to this folder",
            "apply": "only after the user picks this row: delete its cargo output (internal/build-folder; in the moai-adk repository `tools/build_folder.py remove`)",
        }));
    }
    (rows, warnings)
}

/// #2930 item 6: a repository with more worktrees than this (the primary one included) gets a
/// warning line in the preview; exactly this many does not.
pub const WORKTREE_WARN_OVER: usize = 200;

/// The warning for a repository with `count` worktrees, or `None` at [`WORKTREE_WARN_OVER`] or fewer.
pub fn worktree_count_warning(count: usize) -> Option<String> {
    (count > WORKTREE_WARN_OVER).then(|| {
        format!(
            "this repository has {count} worktrees, more than {WORKTREE_WARN_OVER}; `worktree-gc-plan` lists the merged ones that could go"
        )
    })
}

/// `unix_ms` of the first line of a log (the oldest, the log is only appended to), read from at
/// most its first 8 KiB; `None` when the line is longer, empty or not a receipt line.
fn first_line_unix_ms(path: &Path) -> Option<u64> {
    use std::io::Read;
    let mut buf = vec![0u8; 8192];
    let n = fs::File::open(path).ok()?.read(&mut buf).ok()?;
    let text = String::from_utf8_lossy(&buf[..n]);
    let line = text.lines().next()?;
    serde_json::from_str::<Value>(line).ok()?["unix_ms"].as_u64()
}

/// One `hook_log` row for the hook log at `path` (#3009 S4d), or `None` when there is no such
/// file. Always `keep`: the row shows the size, the oldest line, whether the log is full (it
/// stops taking lines at [`crate::hooks::RECEIPT_LOG_LIMIT`]) and whether it is past the design's
/// retention, so the user can decide; nothing is rotated, trimmed or deleted. A link or a
/// folder in the place of the log is listed as such and is not read through.
fn hook_log_row(side: &str, shown: &str, path: &Path, now: u64) -> Option<Value> {
    let meta = fs::symlink_metadata(path).ok()?;
    let regular = meta.is_file();
    let bytes = if regular { meta.len() } else { 0 };
    let oldest = if regular {
        first_line_unix_ms(path).map(|ms| ms / 1000)
    } else {
        None
    };
    let past_days =
        oldest.is_some_and(|t| now.saturating_sub(t) > HOOK_LOG_RETENTION_DAYS * 86_400);
    let past_bytes = bytes > HOOK_LOG_RETENTION_BYTES;
    let full = bytes >= crate::hooks::RECEIPT_LOG_LIMIT;
    let mut reasons = vec!["hook_log_kept"];
    if !regular {
        reasons.push("hook_log_not_a_regular_file");
    }
    if full {
        reasons.push("hook_log_full");
    }
    if past_days || past_bytes {
        reasons.push("past_design_retention");
    }
    let modified = if regular { mtime(path) } else { None };
    Some(json!({
        "kind": "hook_log",
        "side": side,
        "path": shown,
        "action": "keep",
        "reasons": reasons,
        "trim_paths": [],
        "bytes": bytes,
        "bytes_partial": false,
        "modified_unix": modified,
        "age_secs": modified.map(|t| now.saturating_sub(t)),
        "oldest_line_unix": oldest,
        "full": full,
        "limit_bytes": crate::hooks::RECEIPT_LOG_LIMIT,
        "design_retention": {
            "days": HOOK_LOG_RETENTION_DAYS,
            "bytes": HOOK_LOG_RETENTION_BYTES,
            "past_days": past_days,
            "past_bytes": past_bytes,
        },
        "retention_until_unix": Value::Null,
        "restore": "not regenerable — the hook log is evidence and is kept",
        "apply": "none — the design's retention is shown, not applied; nothing is rotated, trimmed or deleted",
    }))
}

/// The hook log rows of the project at `primary`: the project's own log, then the one in the
/// HOME ([`crate::hooks::home_area_dir`]: only when the project has a fixed id and the HOME
/// state root is a real folder, the same place the hooks write and read).
fn hook_log_rows(home: Option<&Path>, primary: &Path, now: u64) -> Vec<Value> {
    let mut rows = Vec::new();
    let own = crate::state_root::path(primary, "receipts").join("hooks.jsonl");
    let shown = own
        .strip_prefix(primary)
        .map_or_else(|_| own.display().to_string(), |p| p.display().to_string());
    rows.extend(hook_log_row("project", &shown, &own, now));
    if let Some(dir) = crate::hooks::home_area_dir(home, primary, "receipts") {
        let log = dir.join("hooks.jsonl");
        rows.extend(hook_log_row("home", &log.display().to_string(), &log, now));
    }
    rows
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
    let listed = parse_worktree_list(&listing);
    let worktree_count = listed.len();
    let primary = listed
        .first()
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .ok_or("git failed: no worktree listed")?;

    let canon = |p: &Path| fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf());
    let listed_paths: BTreeSet<PathBuf> = listed.iter().map(|e| canon(&e.path)).collect();
    let mut merged_paths = BTreeSet::new();
    let common_dir =
        worktree_gc::git_out(req.project, &["rev-parse", "--git-common-dir"]).map(|s| {
            let p = PathBuf::from(s.trim());
            canon(&if p.is_absolute() {
                p
            } else {
                req.project.join(p)
            })
        });

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
        // Audit N2: a branch counts as merged for its build folder only when the gc plan would
        // act on that worktree; a fresh worktree with no commit of its own is an ancestor of the
        // tip too, but the plan keeps it (`recent_change`, a dirty tree, a process in it).
        if !c["merged_by"].is_null() && c["action"] != "skip" {
            merged_paths.insert(canon(&abs));
        }
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
            "lifecycle": c["lifecycle"],
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
    let (agent_rows, agent_warnings) = agent_build_rows(
        req.home,
        common_dir.as_deref(),
        &listed_paths,
        &merged_paths,
        now,
        &cwds,
        deadline,
    );
    rows.extend(agent_rows);

    // Reports are kept, never trimmed. A migration leaves them in the earlier root (#2725 D5),
    // and a project on `.moai/` may write new ones there (#2696 C4d): both are listed.
    let mut report_dirs = vec![crate::state_root::rel(&primary, "reports")];
    if !report_dirs
        .iter()
        .any(|d| d == crate::fixed_state!("reports"))
    {
        report_dirs.push(crate::fixed_state!("reports").to_string());
    }
    for rel in report_dirs {
        let reports = primary.join(&rel);
        if !fs::symlink_metadata(&reports).is_ok_and(|m| m.is_dir()) {
            continue;
        }
        let (bytes, partial) = bytes_under(std::slice::from_ref(&reports), deadline);
        let modified = mtime(&reports);
        rows.push(json!({
            "kind": "report",
            "path": rel,
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

    rows.extend(hook_log_rows(req.home, &primary, now));

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
        "before_free_bytes": gc["before_free_bytes"],
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
        "worktrees": {"count": worktree_count, "warn_over": WORKTREE_WARN_OVER},
        "warnings": worktree_count_warning(worktree_count).into_iter().chain(agent_warnings).collect::<Vec<_>>(),
        "rows": rows,
        "preview_digest": digest,
        "apply": "optimize-plan applies nothing. Worktree rows: worktree-gc-apply recomputes its own plan, where only an ancestor of the integration tip counts as merged, and runs only after the user confirms (#2389 comment 6774899518 item 2). Build folders outside the worktrees and reports: no apply step in #2724. Agent build folders under <HOME>/.moai/build: only the rows the user picks, after this preview (C-82).",
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

    /// #2930 item 6: strictly more than 200 worktrees warns; 200 does not.
    #[test]
    fn the_worktree_count_warning_starts_above_200() {
        assert_eq!(WORKTREE_WARN_OVER, 200);
        for (count, warns) in [
            (0, false),
            (1, false),
            (199, false),
            (200, false),
            (201, true),
        ] {
            assert_eq!(worktree_count_warning(count).is_some(), warns, "{count}");
        }
        assert_eq!(
            worktree_count_warning(201).unwrap(),
            "this repository has 201 worktrees, more than 200; `worktree-gc-plan` lists the merged ones that could go"
        );
    }

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
