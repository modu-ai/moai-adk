//! `hygiene-plan` (#2934 S4a, #2930 R-4): which leftovers of lane work may move to the holding
//! folder `<HOME>/.moai/holding/<UTC date>/`, from where a later purge deletes them after
//! [`HOLDING_KEEP_SECS`]. Read-only; the apply and the purge are separate commands.
//!
//! Two kinds, and the rule for each is not new here:
//! - `worktree`: a linked worktree that `worktree-gc-plan` would remove (merged into the
//!   integration tip, no uncommitted or unsafe ignored file, no process working in it, no
//!   recent change — `crate::worktree_gc`). Hygiene only holds it for a while instead of
//!   deleting it at once.
//! - `tmp_build`: a cargo build output directly in the system temporary folder (`/tmp`
//!   `moai-rc-target-*`, `moai-audit*` …: 277 GB on 2026-11-14, #2930), recognised by the
//!   `CACHEDIR.TAG` cargo writes at its top, left alone for [`TMP_MIN_AGE_SECS`], and in use by
//!   no process — neither as its working folder nor named in its command line (and, on Linux,
//!   its environment: `CARGO_TARGET_DIR`; see [`process_text`]).
//!
//! Fail closed: a fact that cannot be read keeps the item (`keep`) with a named reason.

use crate::worktree_report::bytes_under;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, Instant, UNIX_EPOCH};

type Result<T> = std::result::Result<T, String>;

/// A temporary build output untouched for less than this stays (2 days, the leader's bundle A
/// rule of 2026-11-14, #2930).
pub const TMP_MIN_AGE_SECS: u64 = 2 * 86_400;
/// Held items are deleted only after this long (3 days, #2930 item 4).
pub const HOLDING_KEEP_SECS: u64 = 3 * 86_400;
/// Time spent measuring sizes, in total; a size cut short says so (`bytes_partial`).
pub const SIZE_BUDGET: Duration = Duration::from_secs(20);

/// `<HOME>/.moai/holding`, where the leader's bundle B put worktrees on 2026-11-14.
pub fn holding_root(home: &Path) -> PathBuf {
    home.join(".moai").join("holding")
}

/// `YYYY-MM-DD` (UTC) of a Unix time.
pub fn utc_date(now: u64) -> String {
    crate::forge::policy::format_utc(now as i64)[..10].to_string()
}

/// What the plan reads besides the file system.
pub struct Inputs<'a> {
    pub project: &'a Path,
    pub integration_ref: &'a str,
    /// The folder whose direct children are looked at for `tmp_build` items.
    pub tmp_root: &'a Path,
    pub home: &'a Path,
    pub now: u64,
    /// Working folders of running processes; `None` when they cannot be read.
    pub cwds: Option<Vec<PathBuf>>,
    /// Command lines and environments of running processes; `None` when unreadable.
    pub process_text: Option<String>,
}

/// What running processes name: on Linux every readable `/proc/<pid>/cmdline` and `environ`
/// (so `CARGO_TARGET_DIR` counts); elsewhere the command lines from `ps -ww -A -o command`.
/// macOS shows no other process's environment there (measured 2026-11-14 on Darwin 27: `ps eww`
/// printed `/bin/sleep 31` for a process started with `CARGO_TARGET_DIR=…`), so on macOS a
/// build output named only in an environment is protected by its age alone — a build writes
/// into it and keeps it recent — and the apply re-checks open files before it moves anything.
pub fn process_text() -> Option<String> {
    if Path::new("/proc/self/cmdline").exists() {
        let mut out = String::new();
        for e in fs::read_dir("/proc").ok()?.flatten() {
            for part in ["cmdline", "environ"] {
                if let Ok(b) = fs::read(e.path().join(part)) {
                    out.push_str(&String::from_utf8_lossy(&b).replace('\0', " "));
                    out.push('\n');
                }
            }
        }
        return Some(out);
    }
    let o = Command::new("ps")
        .args(["-ww", "-A", "-o", "command"])
        .env("LC_ALL", "C")
        .output()
        .ok()?;
    if !o.status.success() || o.stdout.is_empty() {
        return None;
    }
    Some(String::from_utf8_lossy(&o.stdout).into_owned())
}

fn mtime(p: &Path) -> Option<u64> {
    fs::symlink_metadata(p)
        .ok()?
        .modified()
        .ok()?
        .duration_since(UNIX_EPOCH)
        .ok()
        .map(|d| d.as_secs())
}

/// The newest change at the top of a build output: the folder itself and its direct children
/// (cargo touches `debug/`, `.rustc_info.json` and the tag on every build).
fn newest_top(dir: &Path) -> Option<u64> {
    let mut newest = mtime(dir)?;
    for e in fs::read_dir(dir).ok()?.flatten() {
        newest = newest.max(mtime(&e.path())?);
    }
    Some(newest)
}

/// Whether a process works in `dir` or names it; `None` when a source is unreadable.
fn in_use(dir: &Path, cwds: &Option<Vec<PathBuf>>, text: &Option<String>) -> Option<bool> {
    let (cwds, text) = (cwds.as_ref()?, text.as_ref()?);
    let canon = fs::canonicalize(dir).unwrap_or_else(|_| dir.to_path_buf());
    let by_cwd = cwds
        .iter()
        .any(|c| c.starts_with(&canon) || c.starts_with(dir));
    let named = [dir, canon.as_path()]
        .iter()
        .any(|p| text.contains(&*p.to_string_lossy()));
    Some(by_cwd || named)
}

/// The `tmp_build` items: every directory directly in `tmp_root` (not a link) whose top holds
/// a regular `CACHEDIR.TAG`. Anything else there is not listed.
fn tmp_builds(inp: &Inputs, deadline: Instant) -> Vec<Value> {
    let mut out = Vec::new();
    let Ok(rd) = fs::read_dir(inp.tmp_root) else {
        return out;
    };
    let mut dirs: Vec<PathBuf> = rd
        .flatten()
        .map(|e| e.path())
        .filter(|p| fs::symlink_metadata(p).is_ok_and(|m| m.is_dir()))
        .filter(|p| fs::symlink_metadata(p.join("CACHEDIR.TAG")).is_ok_and(|m| m.is_file()))
        .collect();
    dirs.sort();
    for dir in dirs {
        let age = newest_top(&dir).map(|t| inp.now.saturating_sub(t));
        let (action, reason) = match (age, in_use(&dir, &inp.cwds, &inp.process_text)) {
            (None, _) => ("keep", "age_unknown"),
            (_, None) => ("keep", "in_use_unknown"),
            (_, Some(true)) => ("keep", "in_use"),
            (Some(a), _) if a < TMP_MIN_AGE_SECS => ("keep", "recent"),
            _ => ("hold", "idle_build_output"),
        };
        let (bytes, partial) = if action == "hold" {
            bytes_under(std::slice::from_ref(&dir), deadline)
        } else {
            (0, false)
        };
        out.push(json!({
            "kind": "tmp_build",
            "path": dir.to_string_lossy(),
            "action": action,
            "reason": reason,
            "age_secs": age,
            "bytes": bytes,
            "bytes_partial": partial,
        }));
    }
    out
}

/// The `worktree` items: `worktree-gc-plan`'s `remove` candidates, as absolute paths; its
/// other candidates are kept with the gc's first reason.
fn worktrees(inp: &Inputs) -> Result<(Vec<Value>, String)> {
    let gc = crate::worktree_gc::plan(
        inp.project,
        inp.integration_ref,
        Duration::from_secs(crate::worktree_gc::DEFAULT_RECENT_MINUTES * 60),
        SIZE_BUDGET,
        Some(inp.now),
    )?;
    // The gc names a worktree relative to the primary checkout when it lies inside it.
    let primary = crate::worktree::worktrees(inp.project)?
        .into_iter()
        .next()
        .map(|e| PathBuf::from(e.path))
        .ok_or("project root unavailable")?;
    let mut out = Vec::new();
    for c in gc["candidates"].as_array().into_iter().flatten() {
        let shown = c["path"].as_str().unwrap_or("");
        let path = if Path::new(shown).is_absolute() {
            PathBuf::from(shown)
        } else if shown == "." {
            primary.clone()
        } else {
            primary.join(shown)
        };
        let hold = c["action"] == "remove";
        out.push(json!({
            "kind": "worktree",
            "path": path.to_string_lossy(),
            "action": if hold { "hold" } else { "keep" },
            "reason": if hold {
                json!("merged_clean_idle")
            } else {
                c["reasons"].get(0).cloned().unwrap_or(json!(c["action"]))
            },
            "head": c["head"],
        }));
    }
    Ok((
        out,
        gc["integration_tip"].as_str().unwrap_or("").to_string(),
    ))
}

/// The plan: items, totals, the holding folder for today and a digest of what an apply moves.
pub fn plan(inp: &Inputs) -> Result<Value> {
    let deadline = Instant::now() + SIZE_BUDGET;
    let (mut items, tip) = worktrees(inp)?;
    items.extend(tmp_builds(inp, deadline));
    items.sort_by(|a, b| a["path"].as_str().cmp(&b["path"].as_str()));
    let bound: Vec<Value> = items
        .iter()
        .filter(|i| i["action"] == "hold")
        .map(|i| json!({"kind": i["kind"], "path": i["path"], "head": i["head"]}))
        .collect();
    let plan_digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    let count = |a: &str| items.iter().filter(|i| i["action"] == a).count();
    let hold_bytes: u64 = items.iter().filter_map(|i| i["bytes"].as_u64()).sum();
    Ok(json!({
        "schema_version": 1,
        "integration_ref": inp.integration_ref,
        "integration_tip": tip,
        "now_unix": inp.now,
        "holding": holding_root(inp.home).join(utc_date(inp.now)).to_string_lossy(),
        "tmp_min_age_secs": TMP_MIN_AGE_SECS,
        "holding_keep_secs": HOLDING_KEEP_SECS,
        "totals": {"hold": count("hold"), "keep": count("keep"), "hold_bytes": hold_bytes},
        "items": items,
        "plan_digest": plan_digest,
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_holding_folder_is_named_by_the_utc_date() {
        assert_eq!(utc_date(0), "1970-01-01");
        assert_eq!(utc_date(1_794_643_200), "2026-11-14");
        assert_eq!(
            holding_root(Path::new("/h")).join(utc_date(1_794_643_200)),
            PathBuf::from("/h/.moai/holding/2026-11-14")
        );
    }

    #[test]
    fn unreadable_process_sources_are_not_a_free_pass() {
        let d = Path::new("/x/build");
        assert_eq!(in_use(d, &None, &Some(String::new())), None);
        assert_eq!(in_use(d, &Some(vec![]), &None), None);
        assert_eq!(in_use(d, &Some(vec![]), &Some(String::new())), Some(false));
        assert_eq!(
            in_use(
                d,
                &Some(vec![PathBuf::from("/x/build/debug")]),
                &Some(String::new())
            ),
            Some(true)
        );
        assert_eq!(
            in_use(
                d,
                &Some(vec![]),
                &Some("cargo CARGO_TARGET_DIR=/x/build".into())
            ),
            Some(true)
        );
    }

    #[test]
    fn an_idle_build_output_is_kept_while_process_sources_are_unreadable() {
        let tmp = tempfile::tempdir().unwrap();
        let build = tmp.path().join("old-target");
        fs::create_dir(&build).unwrap();
        fs::write(
            build.join("CACHEDIR.TAG"),
            "Signature: 8a477f597d28d172789f06886806bc55",
        )
        .unwrap();
        let newest = newest_top(&build).unwrap();
        let mk = |cwds: Option<Vec<PathBuf>>, text: Option<String>| Inputs {
            project: tmp.path(),
            integration_ref: "origin/main",
            tmp_root: tmp.path(),
            home: tmp.path(),
            now: newest + TMP_MIN_AGE_SECS,
            cwds,
            process_text: text,
        };
        let one = |inp: &Inputs| {
            let items = tmp_builds(inp, Instant::now() + SIZE_BUDGET);
            assert_eq!(items.len(), 1, "{items:?}");
            (items[0]["action"].clone(), items[0]["reason"].clone())
        };
        let free = (Some(vec![]), Some(String::new()));
        assert_eq!(
            one(&mk(free.0, free.1)),
            (json!("hold"), json!("idle_build_output"))
        );
        for (cwds, text) in [
            (None, Some(String::new())),
            (Some(vec![]), None),
            (None, None),
        ] {
            assert_eq!(
                one(&mk(cwds, text)),
                (json!("keep"), json!("in_use_unknown"))
            );
        }
    }
}
