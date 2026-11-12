//! `worktree.gc.plan` (WT-4, Issue #2380; design `docs/worktree-strategy-2026-11-10.md`
//! §D5, §D8). Read-only: for every linked worktree it lists whether its cargo build output
//! may be trimmed or the whole worktree removed, with the reasons, and a digest of the list.
//! Nothing is deleted here. Under the user decision of 2026-11-10 (#2389 comment 6774899518,
//! item 2) the leader deletes after checking this list; `worktree.gc.apply` is a later card.
//!
//! Fail closed: a fact that cannot be established (git failed, no process table, the time
//! budget spent) makes the worktree `skip` with a named reason, never a candidate.
//!
//! Two data-loss guards found while designing this card:
//! - lanes keep notes, PR bodies and audit logs under the ignored `target/<name>/`, so a
//!   trim covers only cargo-owned entries of `target/` (profile folders marked by
//!   `.cargo-lock`, target-triple folders marked by their own `CACHEDIR.TAG`, plus
//!   `CACHEDIR.TAG`, `.rustc_info.json` and `tmp`); everything else is listed as preserved;
//! - `git worktree remove` without `--force` deletes ignored files of a clean worktree, so a
//!   worktree is a `remove` candidate only with no untracked file and no ignored file outside
//!   the cargo build output.
//!
//! "In use" evidence: a process whose working directory is inside the worktree (`lsof` cwd
//! list, `/proc/*/cwd` on Linux) and a recent change. Most lanes claim on Issues and do not
//! lock worktrees, so locks only add a reason. The change age is a proxy: the newest
//! modification time among the worktree's git admin files (`index`, `HEAD`, `logs/HEAD`),
//! its top folder, its `target/` and the cargo folders inside — commits, checkouts, builds
//! and new files show there; an uncommitted edit shows as a tracked change instead.
//!
//! What the plan does not check, against design §D5 and the WT-4 card row (audit FAIL
//! 6811382861 B2). `worktree.gc.apply` must re-check the first two right before it deletes:
//! - claim leases: §D5 counts a live claim lease as "in use", but the plan uses no network and
//!   cannot read Issue comments, so a worktree with a live lease and no other sign can still
//!   be a candidate;
//! - open files: §D5 also counts a process holding a file open inside the worktree, but only
//!   working directories are read; a process working elsewhere with a file open inside (the
//!   audit's `tail -f` on `target/debug/deps/blob`) is not seen;
//! - "pushed": the card's push condition is replaced by "HEAD is contained in the local
//!   integration ref" (`merged`), which needs no network and implies the commits already sit
//!   on the integration branch; a pushed but unmerged worktree is `skip` with `not_merged`.

use crate::identity::scrub_git_env;
use crate::worktree_report::{bytes_under, parse_worktree_list};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

type Result<T> = std::result::Result<T, String>;

/// Default "recent change" window: 120 minutes (design §D5).
pub const DEFAULT_RECENT_MINUTES: u64 = 120;
/// Default time budget for the whole plan.
pub const DEFAULT_BUDGET_MS: u64 = 120_000;

/// What was observed about one worktree. `None` means the fact could not be established.
#[derive(Clone, Debug, Default)]
pub struct Facts {
    pub primary: bool,
    pub preserved_location: bool,
    pub current_directory: bool,
    pub locked: bool,
    pub prunable: bool,
    pub budget_spent: bool,
    pub merged: Option<bool>,
    pub tracked_dirty: Option<bool>,
    pub untracked: Option<usize>,
    pub ignored_outside_build: Option<usize>,
    pub in_use: Option<bool>,
    pub change_age_secs: Option<u64>,
    pub has_build_output: bool,
}

/// True when a change `age_secs` old falls inside the window.
pub fn is_recent(age_secs: u64, window_secs: u64) -> bool {
    age_secs < window_secs
}

/// The decision for one worktree: `remove`, `trim` or `skip`, with every reason that held.
pub fn decide(f: &Facts, window_secs: u64) -> (&'static str, Vec<&'static str>) {
    let mut reasons = Vec::new();
    for (hit, why) in [
        (f.primary, "primary"),
        (f.preserved_location, "preserved_location"),
        (f.current_directory, "current_directory"),
        (f.locked, "locked"),
        (f.prunable, "prunable"),
        (f.budget_spent, "budget_spent"),
    ] {
        if hit {
            reasons.push(why);
        }
    }
    match f.merged {
        Some(true) => {}
        Some(false) => reasons.push("not_merged"),
        None => reasons.push("merged_unknown"),
    }
    match f.tracked_dirty {
        Some(false) => {}
        Some(true) => reasons.push("tracked_changes"),
        None => reasons.push("status_unknown"),
    }
    match f.in_use {
        Some(false) => {}
        Some(true) => reasons.push("process_in_use"),
        None => reasons.push("process_check_unavailable"),
    }
    match f.change_age_secs {
        Some(age) if is_recent(age, window_secs) => reasons.push("recent_change"),
        Some(_) => {}
        None => reasons.push("change_age_unknown"),
    }
    if !reasons.is_empty() {
        return ("skip", reasons);
    }
    match (f.untracked, f.ignored_outside_build) {
        (Some(0), Some(0)) => return ("remove", reasons),
        (u, i) => {
            if u.is_none() || i.is_none() {
                reasons.push("file_census_unknown");
            }
            if u.is_some_and(|n| n > 0) {
                reasons.push("untracked_files");
            }
            if i.is_some_and(|n| n > 0) {
                reasons.push("ignored_files_outside_build_output");
            }
        }
    }
    if f.has_build_output {
        ("trim", reasons)
    } else {
        reasons.push("no_build_output");
        ("skip", reasons)
    }
}

/// Splits the children of a `target/` folder into cargo-owned entries and others.
pub fn split_target(target: &Path) -> (Vec<PathBuf>, Vec<String>) {
    let (mut cargo, mut other) = (Vec::new(), Vec::new());
    let Ok(rd) = fs::read_dir(target) else {
        return (cargo, other);
    };
    for e in rd.flatten() {
        let name = e.file_name().to_string_lossy().into_owned();
        let path = e.path();
        let Ok(meta) = fs::symlink_metadata(&path) else {
            continue;
        };
        let owned = if meta.is_dir() {
            name == "tmp"
                || path.join(".cargo-lock").is_file()
                || path.join("CACHEDIR.TAG").is_file()
        } else {
            meta.is_file() && (name == "CACHEDIR.TAG" || name == ".rustc_info.json")
        };
        if owned {
            cargo.push(path);
        } else {
            other.push(name);
        }
    }
    cargo.sort();
    other.sort();
    (cargo, other)
}

fn git_out(dir: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let out = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).into_owned())
}

/// `Some(true)` when `head` is an ancestor of (or equal to) `tip`, `Some(false)` when not.
fn merged(primary: &Path, head: &str, tip: &str) -> Option<bool> {
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let st = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(primary)
        .args(["merge-base", "--is-ancestor", head, tip])
        .status()
        .ok()?;
    match st.code() {
        Some(0) => Some(true),
        Some(1) => Some(false),
        _ => None,
    }
}

/// Working directories of the processes this user can see; `None` without a source.
fn process_cwds() -> Option<Vec<PathBuf>> {
    if Path::new("/proc/self/cwd").exists() {
        let mut out = Vec::new();
        for e in fs::read_dir("/proc").ok()?.flatten() {
            if let Ok(p) = fs::read_link(e.path().join("cwd")) {
                out.push(p);
            }
        }
        return Some(out);
    }
    let o = Command::new("lsof")
        .args(["-w", "-a", "-d", "cwd", "-Fn"])
        .env("LC_ALL", "C")
        .output()
        .ok()?;
    // lsof exits 1 when some processes cannot be read; its listing is still usable.
    if o.stdout.is_empty() {
        return None;
    }
    Some(
        String::from_utf8_lossy(&o.stdout)
            .lines()
            .filter_map(|l| l.strip_prefix('n'))
            .map(PathBuf::from)
            .collect(),
    )
}

fn unix_secs(t: SystemTime) -> Option<u64> {
    t.duration_since(UNIX_EPOCH).ok().map(|d| d.as_secs())
}

/// Newest modification time (Unix seconds) among the change-age proxy paths.
fn newest_change(wt: &Path, target: &Path, cargo: &[PathBuf]) -> Option<u64> {
    let gitdir = git_out(wt, &["rev-parse", "--absolute-git-dir"])?;
    let gitdir = PathBuf::from(gitdir.trim());
    let mut paths = vec![
        gitdir.join("index"),
        gitdir.join("HEAD"),
        gitdir.join("logs/HEAD"),
        wt.to_path_buf(),
        target.to_path_buf(),
    ];
    for c in cargo {
        paths.push(c.clone());
        paths.push(c.join("deps"));
        paths.push(c.join("incremental"));
    }
    paths
        .iter()
        .filter_map(|p| fs::symlink_metadata(p).ok())
        .filter_map(|m| m.modified().ok())
        .filter_map(unix_secs)
        .max()
}

/// `(untracked, ignored entries outside target/)` from one porcelain status.
fn census(wt: &Path) -> Option<(usize, usize, bool)> {
    let s = git_out(
        wt,
        &[
            "status",
            "--porcelain=v1",
            "-z",
            "--untracked-files=normal",
            "--ignored=traditional",
        ],
    )?;
    let (mut untracked, mut ignored, mut target_ignored) = (0, 0, false);
    for rec in s.split('\0').filter(|r| r.len() > 3) {
        let (code, path) = rec.split_at(3);
        match code {
            "?? " => untracked += 1,
            "!! " if path == "target/" => target_ignored = true,
            "!! " => ignored += 1,
            _ => {}
        }
    }
    Some((untracked, ignored, target_ignored))
}

fn shown(path: &Path, root: &Path) -> String {
    match path.strip_prefix(root) {
        Ok(rel) if rel.as_os_str().is_empty() => ".".into(),
        Ok(rel) => rel.to_string_lossy().into_owned(),
        Err(_) => path.to_string_lossy().into_owned(),
    }
}

/// Builds the plan. `now_unix` overrides the clock (tests); `None` uses the system time.
pub fn plan(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
) -> Result<Value> {
    let deadline = Instant::now() + budget;
    let now = now_unix
        .or_else(|| unix_secs(SystemTime::now()))
        .ok_or("system clock unavailable")?;
    let top = git_out(
        project,
        &["rev-parse", "--path-format=absolute", "--show-toplevel"],
    )
    .map(|s| PathBuf::from(s.trim()))
    .ok_or("not inside a git work tree")?;
    let spec = format!("{integration_ref}^{{commit}}");
    let tip = git_out(&top, &["rev-parse", "--verify", "--end-of-options", &spec])
        .map(|s| s.trim().to_string())
        .ok_or("integration ref not found; give a local ref such as origin/<branch>")?;
    let listing =
        git_out(&top, &["worktree", "list", "--porcelain"]).ok_or("git worktree list failed")?;
    let entries = parse_worktree_list(&listing);
    let primary = entries
        .first()
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .unwrap_or_else(|| top.clone());
    let cwd = std::env::current_dir()
        .ok()
        .and_then(|p| fs::canonicalize(p).ok());
    let cwds = process_cwds().map(|v| {
        v.into_iter()
            .map(|p| fs::canonicalize(&p).unwrap_or(p))
            .collect::<Vec<_>>()
    });
    let window = recent.as_secs();

    let mut candidates = Vec::new();
    for (i, e) in entries.iter().enumerate() {
        if e.bare {
            continue;
        }
        let path = fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone());
        let target = path.join("target");
        let (cargo, preserved) = split_target(&target);
        let mut f = Facts {
            primary: i == 0,
            preserved_location: path.components().any(|c| c.as_os_str() == ".claude")
                && path.to_string_lossy().contains("/.claude/worktrees/"),
            current_directory: cwd.as_ref().is_some_and(|c| c.starts_with(&path)) && i != 0,
            locked: e.locked,
            prunable: e.prunable,
            has_build_output: !cargo.is_empty(),
            ..Default::default()
        };
        if Instant::now() >= deadline {
            f.budget_spent = true;
        } else if !e.prunable {
            f.merged = e.head.as_deref().and_then(|h| merged(&primary, h, &tip));
            f.tracked_dirty = git_out(&path, &["status", "--porcelain", "--untracked-files=no"])
                .map(|s| !s.trim().is_empty());
            if let Some((u, ign, _)) = census(&path) {
                f.untracked = Some(u);
                f.ignored_outside_build = Some(ign + preserved.len());
            }
            f.in_use = cwds
                .as_ref()
                .map(|list| list.iter().any(|c| c.starts_with(&path)));
            f.change_age_secs =
                newest_change(&path, &target, &cargo).map(|t| now.saturating_sub(t));
        }
        let (action, reasons) = decide(&f, window);
        let (trim_bytes, partial) = if action == "skip" {
            (0, false)
        } else {
            bytes_under(&cargo, deadline)
        };
        candidates.push(json!({
            "path": shown(&path, &primary),
            "branch": e.branch,
            "head": e.head,
            "action": action,
            "reasons": reasons,
            "trim_paths": cargo.iter().map(|c| shown(c, &path)).collect::<Vec<_>>(),
            "preserved_in_target": preserved,
            "untracked": f.untracked,
            "ignored_outside_build": f.ignored_outside_build,
            "change_age_secs": f.change_age_secs,
            "trim_bytes": trim_bytes,
            "bytes_partial": partial,
        }));
    }
    candidates.sort_by(|a, b| a["path"].as_str().cmp(&b["path"].as_str()));
    // The digest binds what a later apply would act on: path, head, action and reasons.
    let bound: Vec<Value> = candidates
        .iter()
        .map(|c| json!({"path": c["path"], "head": c["head"], "action": c["action"], "reasons": c["reasons"]}))
        .collect();
    let plan_digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    let count = |a: &str| candidates.iter().filter(|c| c["action"] == a).count();
    let trim_total: u64 = candidates
        .iter()
        .filter_map(|c| c["trim_bytes"].as_u64())
        .sum();
    Ok(json!({
        "schema_version": 1,
        "integration_ref": integration_ref,
        "integration_tip": tip,
        "recent_minutes": window / 60,
        "now_unix": now,
        "totals": {
            "remove": count("remove"),
            "trim": count("trim"),
            "skip": count("skip"),
            "trim_bytes": trim_total,
        },
        "candidates": candidates,
        "plan_digest": plan_digest,
        "apply": "not implemented here: the leader deletes after checking this list (#2389 comment 6774899518 item 2); worktree.gc.apply is a later card",
        "network": "disabled",
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn eligible() -> Facts {
        Facts {
            merged: Some(true),
            tracked_dirty: Some(false),
            untracked: Some(0),
            ignored_outside_build: Some(0),
            in_use: Some(false),
            change_age_secs: Some(7_200),
            has_build_output: true,
            ..Default::default()
        }
    }

    #[test]
    fn recency_boundary_is_exactly_the_window() {
        assert!(is_recent(7_199, 7_200));
        assert!(!is_recent(7_200, 7_200));
        let mut f = eligible();
        f.change_age_secs = Some(7_199);
        assert_eq!(decide(&f, 7_200), ("skip", vec!["recent_change"]));
        f.change_age_secs = Some(7_200);
        assert_eq!(decide(&f, 7_200).0, "remove");
    }

    #[test]
    fn every_unknown_or_blocking_fact_skips() {
        type Edit = fn(&mut Facts);
        let cases: [(Edit, &str); 14] = [
            (|f| f.primary = true, "primary"),
            (|f| f.preserved_location = true, "preserved_location"),
            (|f| f.current_directory = true, "current_directory"),
            (|f| f.locked = true, "locked"),
            (|f| f.prunable = true, "prunable"),
            (|f| f.budget_spent = true, "budget_spent"),
            (|f| f.merged = Some(false), "not_merged"),
            (|f| f.merged = None, "merged_unknown"),
            (|f| f.tracked_dirty = Some(true), "tracked_changes"),
            (|f| f.tracked_dirty = None, "status_unknown"),
            (|f| f.in_use = Some(true), "process_in_use"),
            (|f| f.in_use = None, "process_check_unavailable"),
            (|f| f.change_age_secs = None, "change_age_unknown"),
            (|f| f.change_age_secs = Some(0), "recent_change"),
        ];
        for (edit, why) in cases {
            let mut f = eligible();
            edit(&mut f);
            assert_eq!(decide(&f, 7_200), ("skip", vec![why]), "{why}");
        }
    }

    #[test]
    fn leftovers_downgrade_remove_to_trim_or_skip() {
        assert_eq!(decide(&eligible(), 7_200), ("remove", vec![]));
        let mut f = eligible();
        f.untracked = Some(1);
        assert_eq!(decide(&f, 7_200), ("trim", vec!["untracked_files"]));
        let mut f = eligible();
        f.ignored_outside_build = Some(1);
        assert_eq!(
            decide(&f, 7_200),
            ("trim", vec!["ignored_files_outside_build_output"])
        );
        let mut f = eligible();
        f.untracked = None;
        assert_eq!(decide(&f, 7_200), ("trim", vec!["file_census_unknown"]));
        f.has_build_output = false;
        assert_eq!(
            decide(&f, 7_200),
            ("skip", vec!["file_census_unknown", "no_build_output"])
        );
    }
}
