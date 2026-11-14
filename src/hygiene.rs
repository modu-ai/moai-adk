//! `hygiene-plan` (#2934 S4a, #2930 R-4): which leftovers of lane work may move to the holding
//! folder `<HOME>/.moai/holding/<UTC date>/`, from where a later purge deletes them after
//! [`HOLDING_KEEP_SECS`]. The plan is read-only; `hygiene-apply` (#2934 S4b, [`apply`]) moves
//! what a matching plan holds, and the purge is a separate command.
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

/// The first line of every `manifest.tsv`.
pub const MANIFEST_HEADER: &str = "original\tholding\tkind\tbytes\treason\n";

/// Refuses a path that is a symbolic link (an absent path passes).
fn not_a_link(p: &Path) -> Result<()> {
    match fs::symlink_metadata(p) {
        Ok(m) if m.file_type().is_symlink() => Err(format!(
            "{} is a symlink; refused, nothing was moved",
            p.display()
        )),
        _ => Ok(()),
    }
}

/// Opens `p` without following a link at its last component.
fn open_nofollow(o: &mut fs::OpenOptions, p: &Path) -> std::io::Result<fs::File> {
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        o.custom_flags(libc::O_NOFOLLOW);
    }
    o.open(p)
}

/// Creates `manifest` with its header, or accepts an existing regular file as it is.
fn start_manifest(manifest: &Path) -> Result<()> {
    match open_nofollow(
        fs::OpenOptions::new().write(true).create_new(true),
        manifest,
    ) {
        Ok(mut f) => {
            use std::io::Write;
            f.write_all(MANIFEST_HEADER.as_bytes())
                .and_then(|()| f.sync_data())
                .map_err(|e| format!("manifest could not be written: {e}"))
        }
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {
            match fs::symlink_metadata(manifest) {
                Ok(m) if m.is_file() => Ok(()),
                _ => Err(format!(
                    "{} is a symlink or not a file; refused, nothing was moved",
                    manifest.display()
                )),
            }
        }
        Err(e) => Err(format!("manifest could not be created: {e}")),
    }
}

/// Appends and flushes one manifest line; returns the file and its length before the line.
fn record(manifest: &Path, line: &str) -> std::io::Result<(fs::File, u64)> {
    use std::io::Write;
    let mut f = open_nofollow(fs::OpenOptions::new().append(true), manifest)?;
    let len = f.metadata()?.len();
    f.write_all(line.as_bytes())?;
    f.sync_data()?;
    Ok((f, len))
}

/// Why `dir` must not move right now, checked just before the move: a process working in it
/// or naming it ([`in_use`]), or one holding a file open inside it. An unknown fact refuses.
fn busy_now(dir: &Path) -> Option<&'static str> {
    match in_use(dir, &crate::worktree_gc::process_cwds(), &process_text()) {
        None => return Some("in_use_unknown"),
        Some(true) => return Some("in_use"),
        Some(false) => {}
    }
    let canon = fs::canonicalize(dir).unwrap_or_else(|_| dir.to_path_buf());
    match crate::worktree_gc::open_file_paths() {
        None => Some("open_file_check_unavailable"),
        Some(open) if open.iter().any(|p| p.starts_with(&canon)) => Some("file_open_in_use"),
        Some(_) => None,
    }
}

/// `hygiene.apply` (#2934 S4b): recomputes the plan with the same inputs and refuses before
/// any write unless its digest equals `plan_digest` or the holding folder or its manifest is a
/// link. Then each `hold` item, re-checked right before ([`busy_now`]), gets its manifest line
/// written and flushed first and moves by one same-file-system `rename` into
/// `<holding>/<n>-<name>` — never a copy; a failed rename (another file system included) skips
/// the item and takes its line back. Moved worktrees are pruned from git's list.
pub fn apply(inp: &Inputs, plan_digest: &str, permission_reference: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let fresh = plan(inp)?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    let root = holding_root(inp.home);
    let holding = root.join(utc_date(inp.now));
    let manifest = holding.join("manifest.tsv");
    for p in [&root, &holding, &manifest] {
        not_a_link(p)?;
    }
    let items: Vec<&Value> = fresh["items"]
        .as_array()
        .into_iter()
        .flatten()
        .filter(|i| i["action"] == "hold")
        .collect();
    let (mut moved, mut skipped) = (Vec::new(), Vec::new());
    let mut moved_worktree = false;
    if !items.is_empty() {
        fs::create_dir_all(&holding).map_err(|e| format!("holding folder: {e}"))?;
        for p in [&root, &holding] {
            not_a_link(p)?;
        }
        start_manifest(&manifest)?;
    }
    let mut n = fs::read_dir(&holding)
        .map(|rd| {
            rd.flatten()
                .filter(|e| e.file_name() != "manifest.tsv")
                .count()
        })
        .unwrap_or(0);
    let deadline = Instant::now() + SIZE_BUDGET;
    for it in &items {
        let path = PathBuf::from(it["path"].as_str().unwrap_or_default());
        let kind = it["kind"].as_str().unwrap_or_default();
        let reason = it["reason"].as_str().unwrap_or_default();
        let skip =
            |why: String| json!({"path": path.to_string_lossy(), "kind": kind, "reason": why});
        let name = path.file_name().map(|s| s.to_string_lossy().into_owned());
        let shown = path.to_string_lossy();
        if !fs::symlink_metadata(&path).is_ok_and(|m| m.is_dir()) || name.is_none() {
            skipped.push(skip("not_a_folder".into()));
            continue;
        }
        if shown.contains(['\t', '\n']) {
            skipped.push(skip("path_not_tsv_safe".into()));
            continue;
        }
        if let Some(why) = busy_now(&path) {
            skipped.push(skip(why.into()));
            continue;
        }
        let dest = loop {
            n += 1;
            let d = holding.join(format!("{n}-{}", name.as_deref().unwrap_or_default()));
            if fs::symlink_metadata(&d).is_err() {
                break d;
            }
        };
        let bytes = it["bytes"]
            .as_u64()
            .unwrap_or_else(|| bytes_under(std::slice::from_ref(&path), deadline).0);
        let line = format!("{shown}\t{}\t{kind}\t{bytes}\t{reason}\n", dest.display());
        let staged = record(&manifest, &line).ok();
        if staged.is_none() {
            skipped.push(skip("manifest_write_failed".into()));
            continue;
        }
        if let Err(e) = fs::rename(&path, &dest) {
            if let Some((f, len)) = staged {
                let _ = f.set_len(len).and_then(|()| f.sync_data());
            }
            let why = if e.kind() == std::io::ErrorKind::CrossesDevices {
                "other_file_system".to_string()
            } else {
                format!("rename_failed: {e}")
            };
            skipped.push(skip(why));
            continue;
        }
        moved_worktree |= kind == "worktree";
        moved.push(json!({
            "original": shown,
            "holding": dest.to_string_lossy(),
            "kind": kind,
            "bytes": bytes,
            "reason": reason,
        }));
    }
    let pruned = moved_worktree
        && crate::worktree_gc::git_out(inp.project, &["worktree", "prune"]).is_some();
    Ok(json!({
        "schema_version": 1,
        "status": if moved.is_empty() && skipped.is_empty() { "unchanged" } else if skipped.is_empty() { "applied" } else { "partial" },
        "plan_digest": plan_digest,
        "holding": holding.to_string_lossy(),
        "manifest": manifest.to_string_lossy(),
        "moved": moved,
        "skipped": skipped,
        "pruned": pruned,
        "retry": "none: a skipped item is reported, never retried",
        "network": "disabled",
        "writes": !items.is_empty(),
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
