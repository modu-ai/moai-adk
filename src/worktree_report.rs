//! Read-only worktree space report (WT-1, Issue #2380 stage 2; design
//! `docs/worktree-strategy-2026-11-10.md` §D7, §D8 `worktree.report`).
//!
//! The report states facts only: how many worktrees, build folders, temporary build
//! folders and audit clones exist and how much space they hold, plus the disk's free
//! space. It does not judge what may be removed. The design's "in use" conditions (§D5:
//! live claim lease, processes using a path, the remote head) need the network, process
//! tables and a user decision (Q4), so candidates are left to `worktree.gc.plan` (WT-4)
//! and the report says `candidates.status = "not_evaluated"`.
//!
//! Read-only: git is read with `--no-optional-locks` (a plain `status` may refresh and
//! rewrite another lane's index), symlinks are never followed, nothing is created. The
//! size walk and the per-worktree `status` calls share one time budget; past it a class
//! is marked `partial` and its bytes are a lower bound. Other lanes create and delete
//! folders while the report runs, so an entry that vanishes mid-walk is skipped.
//!
//! Sizes follow `du`: allocated blocks, each inode counted once (cargo `target/` holds
//! many hard links). On platforms without block and inode metadata the apparent length is
//! used and `size_basis` says so.

use crate::identity::scrub_git_env;
use serde_json::{json, Value};
use std::collections::HashSet;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, Instant};

type Result<T> = std::result::Result<T, String>;

/// Default time budget for the size walk and per-worktree status calls.
pub const DEFAULT_BUDGET_MS: u64 = 30_000;

/// Prefixes of the per-session cargo build folders under the temporary root, checked before
/// any audit pattern (`moai-target-agent-2-t61-audit-2326` is a build folder). Lanes name new
/// folders `moai-target-<label>` (#2563); sessions that already build in the former
/// `moai-rc-target-<label>` keep it, so both prefixes count — a cleanup or a total that
/// skipped the old folders would miss live build output.
const TARGET_PREFIXES: [&str; 2] = ["moai-target-", "moai-rc-target-"];

/// The temporary folders the design counted as empty leftovers (§1.1).
const EMPTY_PREFIXES: [&str; 4] = [
    "moai-cli-home-",
    "moai-cli-profiles-",
    "moai-cli-gitconfig-",
    "moai-audit-receipts-",
];

/// What a temporary `moai-*` folder is, by name and shape.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum TmpClass {
    Target,
    AuditClone,
    EmptyDir,
    Other,
}

/// Classifies one folder directly under the temporary root. `is_empty` and `has_git` are
/// the folder's observed shape. The build-folder prefix wins over any audit pattern.
pub fn classify_tmp(name: &str, is_empty: bool, has_git: bool) -> Option<TmpClass> {
    if !name.starts_with("moai-") {
        return None;
    }
    if TARGET_PREFIXES.iter().any(|p| name.starts_with(p)) {
        return Some(TmpClass::Target);
    }
    if is_empty && EMPTY_PREFIXES.iter().any(|p| name.starts_with(p)) {
        return Some(TmpClass::EmptyDir);
    }
    if name.contains("audit") && has_git {
        return Some(TmpClass::AuditClone);
    }
    Some(TmpClass::Other)
}

/// Bytes counted under one root, and whether the walk stopped early.
#[derive(Default)]
struct Walk {
    bytes: u64,
    partial: bool,
}

struct Sizer {
    deadline: Instant,
    seen: HashSet<(u64, u64)>,
}

impl Sizer {
    fn expired(&self) -> bool {
        Instant::now() >= self.deadline
    }

    /// Allocated size of one entry; `None` when this inode was already counted.
    fn entry_bytes(&mut self, meta: &fs::Metadata) -> Option<u64> {
        #[cfg(unix)]
        {
            use std::os::unix::fs::MetadataExt;
            if meta.nlink() > 1 && !meta.is_dir() && !self.seen.insert((meta.dev(), meta.ino())) {
                return None;
            }
            Some(meta.blocks() * 512)
        }
        #[cfg(not(unix))]
        {
            let _ = &self.seen;
            Some(if meta.is_dir() { 0 } else { meta.len() })
        }
    }

    /// Walks `root` without following symlinks. Paths in `skip` (and below) are not
    /// entered. Vanished or unreadable entries are skipped; an unreadable directory marks
    /// the walk partial because its bytes are unknown.
    fn walk(&mut self, root: &Path, skip: &[PathBuf]) -> Walk {
        let mut out = Walk::default();
        let mut stack = vec![root.to_path_buf()];
        while let Some(path) = stack.pop() {
            if self.expired() {
                out.partial = true;
                return out;
            }
            if skip.iter().any(|s| s == &path) {
                continue;
            }
            let meta = match fs::symlink_metadata(&path) {
                Ok(m) => m,
                Err(_) => continue,
            };
            if let Some(b) = self.entry_bytes(&meta) {
                out.bytes += b;
            }
            if meta.is_dir() {
                match fs::read_dir(&path) {
                    Ok(entries) => {
                        for e in entries.flatten() {
                            stack.push(e.path());
                        }
                    }
                    Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
                    Err(_) => out.partial = true,
                }
            }
        }
        out
    }
}

/// Allocated bytes under `paths` (symlinks not followed, each inode once) and whether the
/// walk stopped at `deadline` (then the total is a lower bound). Shared with WT-4.
pub(crate) fn bytes_under(paths: &[PathBuf], deadline: Instant) -> (u64, bool) {
    let mut sizer = Sizer {
        deadline,
        seen: HashSet::new(),
    };
    let (mut bytes, mut partial) = (0, false);
    for p in paths {
        let w = sizer.walk(p, &[]);
        bytes += w.bytes;
        partial |= w.partial;
    }
    (bytes, partial)
}

fn git(dir: &Path, args: &[&str]) -> Option<String> {
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

/// One entry of `git worktree list --porcelain`.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct WorktreeEntry {
    pub path: PathBuf,
    pub head: Option<String>,
    pub branch: Option<String>,
    pub detached: bool,
    pub bare: bool,
    pub locked: bool,
    pub lock_reason: Option<String>,
    pub prunable: bool,
}

/// Parses `git worktree list --porcelain` output (records separated by blank lines).
pub fn parse_worktree_list(text: &str) -> Vec<WorktreeEntry> {
    let mut out = Vec::new();
    let mut cur: Option<WorktreeEntry> = None;
    for line in text.lines() {
        if line.is_empty() {
            out.extend(cur.take());
            continue;
        }
        let (key, val) = line.split_once(' ').unwrap_or((line, ""));
        if key == "worktree" {
            out.extend(cur.take());
            cur = Some(WorktreeEntry {
                path: PathBuf::from(val),
                ..Default::default()
            });
            continue;
        }
        let Some(e) = cur.as_mut() else { continue };
        match key {
            "HEAD" => e.head = Some(val.to_string()),
            "branch" => e.branch = Some(val.trim_start_matches("refs/heads/").to_string()),
            "detached" => e.detached = true,
            "bare" => e.bare = true,
            "locked" => {
                e.locked = true;
                e.lock_reason = (!val.is_empty()).then(|| val.to_string());
            }
            "prunable" => e.prunable = true,
            _ => {}
        }
    }
    out.extend(cur);
    out
}

/// `(free, total)` bytes of the file system holding `path`, from POSIX `df -Pk`; `None` off
/// Unix or when `df` cannot be read.
pub fn disk_bytes(path: &Path) -> Option<(u64, u64)> {
    if !cfg!(unix) {
        return None;
    }
    Command::new("df")
        .arg("-Pk")
        .arg(path)
        .env("LC_ALL", "C")
        .output()
        .ok()
        .filter(|o| o.status.success())
        .and_then(|o| {
            let text = String::from_utf8_lossy(&o.stdout).into_owned();
            let line = text.lines().nth(1)?.to_string();
            let f: Vec<&str> = line.split_whitespace().collect();
            let total = f.get(1)?.parse::<u64>().ok()?;
            let avail = f.get(3)?.parse::<u64>().ok()?;
            Some((avail * 1024, total * 1024))
        })
}

/// Bytes held by the repository's worktrees (their files plus their `target/`), and whether
/// the bounded walk was cut short (then the total is a lower bound). Used by the WT-5
/// warning of `worktree-apply`. `None` when the report cannot run.
pub fn project_bytes(project: &Path, budget: Duration) -> Option<(u64, bool)> {
    // No temporary root is scanned for this total.
    let r = report(project, Path::new("/nonexistent-moai-tmp-root"), budget).ok()?;
    let c = &r["classes"];
    let bytes = c["worktree_checkout"]["bytes"].as_u64().unwrap_or(0)
        + c["worktree_target"]["bytes"].as_u64().unwrap_or(0);
    let partial =
        c["worktree_checkout"]["partial"] == true || c["worktree_target"]["partial"] == true;
    Some((bytes, partial))
}

/// Free and total bytes of the file system holding `path`, as the report's `disk` object.
fn disk(path: &Path) -> Value {
    if !cfg!(unix) {
        return json!({"free_bytes": null, "total_bytes": null, "reason": "not measured on this platform"});
    }
    match disk_bytes(path) {
        Some((free, total)) => json!({"free_bytes": free, "total_bytes": total, "reason": null}),
        None => {
            json!({"free_bytes": null, "total_bytes": null, "reason": "df -Pk unavailable or unreadable"})
        }
    }
}

fn class(count: u64, bytes: u64, partial: bool) -> Value {
    json!({"count": count, "bytes": bytes, "partial": partial})
}

/// A worktree path as shown in the report: relative to the project root when inside it.
fn shown(path: &Path, root: &Path) -> (String, bool) {
    match path.strip_prefix(root) {
        Ok(rel) if rel.as_os_str().is_empty() => (".".into(), true),
        Ok(rel) => (rel.to_string_lossy().into_owned(), true),
        Err(_) => (path.to_string_lossy().into_owned(), false),
    }
}

/// Builds the report for the repository containing `project`. `tmp_root` is the folder
/// whose `moai-*` children are classified (the CLI default is `/tmp`). Writes nothing.
pub fn report(project: &Path, tmp_root: &Path, budget: Duration) -> Result<Value> {
    let started = Instant::now();
    let deadline = started + budget;
    let top = git(
        project,
        &["rev-parse", "--path-format=absolute", "--show-toplevel"],
    )
    .map(|s| PathBuf::from(s.trim()))
    .ok_or("not inside a git work tree")?;
    let listing =
        git(&top, &["worktree", "list", "--porcelain"]).ok_or("git worktree list failed")?;
    let entries = parse_worktree_list(&listing);
    let primary = entries
        .first()
        .map(|e| e.path.clone())
        .unwrap_or_else(|| top.clone());
    let primary = fs::canonicalize(&primary).unwrap_or(primary);
    let mut sizer = Sizer {
        deadline,
        seen: HashSet::new(),
    };
    let wt_paths: Vec<PathBuf> = entries
        .iter()
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .collect();

    let (mut co_count, mut co_bytes, mut co_partial) = (0u64, 0u64, false);
    let (mut tg_count, mut tg_bytes, mut tg_partial) = (0u64, 0u64, false);
    let mut worktrees = Vec::new();
    for (e, path) in entries.iter().zip(&wt_paths) {
        if e.bare {
            continue;
        }
        let target = path.join("target");
        let has_target = fs::symlink_metadata(&target)
            .map(|m| m.is_dir())
            .unwrap_or(false);
        // The checkout excludes its own build folder, its git directory and any other
        // worktree nested inside it (the primary holds `.moai-cli/worktrees/*`).
        let mut skip = vec![target.clone(), path.join(".git")];
        skip.extend(
            wt_paths
                .iter()
                .filter(|p| *p != path && p.starts_with(path))
                .cloned(),
        );
        let co = sizer.walk(path, &skip);
        let tg = if has_target {
            sizer.walk(&target, &[])
        } else {
            Walk::default()
        };
        let dirty = if sizer.expired() || e.prunable {
            Value::Null
        } else {
            match git(path, &["status", "--porcelain"]) {
                Some(s) => json!(!s.trim().is_empty()),
                None => Value::Null,
            }
        };
        co_count += 1;
        co_bytes += co.bytes;
        co_partial |= co.partial;
        if has_target {
            tg_count += 1;
            tg_bytes += tg.bytes;
            tg_partial |= tg.partial;
        }
        let (shown_path, inside) = shown(path, &primary);
        worktrees.push(json!({
            "path": shown_path,
            "inside_project": inside,
            "branch": e.branch,
            "head": e.head,
            "detached": e.detached,
            "locked": e.locked,
            "lock_reason": e.lock_reason,
            "prunable": e.prunable,
            "dirty": dirty,
            "has_target": has_target,
            "target_cachedir_tag": has_target && target.join("CACHEDIR.TAG").is_file(),
            "checkout_bytes": co.bytes,
            "target_bytes": tg.bytes,
            "partial": co.partial || tg.partial,
        }));
    }

    let mut tmp = [(0u64, 0u64, false); 4];
    let mut tmp_listed = true;
    match fs::read_dir(tmp_root) {
        Ok(rd) => {
            for ent in rd.flatten() {
                let name = ent.file_name().to_string_lossy().into_owned();
                let path = ent.path();
                let Ok(meta) = fs::symlink_metadata(&path) else {
                    continue;
                };
                if !meta.is_dir() {
                    continue;
                }
                let is_empty = fs::read_dir(&path)
                    .map(|mut r| r.next().is_none())
                    .unwrap_or(false);
                let has_git = fs::symlink_metadata(path.join(".git")).is_ok();
                let Some(cls) = classify_tmp(&name, is_empty, has_git) else {
                    continue;
                };
                let i = cls as usize;
                let w = if is_empty {
                    Walk::default()
                } else {
                    sizer.walk(&path, &[])
                };
                tmp[i].0 += 1;
                tmp[i].1 += w.bytes;
                tmp[i].2 |= w.partial;
            }
        }
        Err(_) => tmp_listed = false,
    }
    if !tmp_listed {
        for t in tmp.iter_mut() {
            t.2 = true;
        }
    }

    let refs = git(
        &top,
        &[
            "for-each-ref",
            "--format=%(refname)",
            "refs/heads",
            "refs/remotes/pr",
        ],
    )
    .unwrap_or_default();
    let branches = refs
        .lines()
        .filter(|l| l.starts_with("refs/heads/"))
        .count();
    let pr_refs = refs
        .lines()
        .filter(|l| l.starts_with("refs/remotes/pr/"))
        .count();

    let classes = json!({
        "worktree_checkout": class(co_count, co_bytes, co_partial),
        "worktree_target": class(tg_count, tg_bytes, tg_partial),
        "tmp_target": class(tmp[0].0, tmp[0].1, tmp[0].2),
        "tmp_audit_clone": class(tmp[1].0, tmp[1].1, tmp[1].2),
        "tmp_empty_dir": class(tmp[2].0, tmp[2].1, tmp[2].2),
        "tmp_other": class(tmp[3].0, tmp[3].1, tmp[3].2),
        "local_refs": {"branches": branches, "remote_pr_refs": pr_refs},
    });
    let partial_classes: Vec<&str> = [
        ("worktree_checkout", co_partial),
        ("worktree_target", tg_partial),
        ("tmp_target", tmp[0].2),
        ("tmp_audit_clone", tmp[1].2),
        ("tmp_empty_dir", tmp[2].2),
        ("tmp_other", tmp[3].2),
    ]
    .iter()
    .filter(|(_, p)| *p)
    .map(|(n, _)| *n)
    .collect();
    Ok(json!({
        "schema_version": 1,
        "tmp_root": tmp_root.to_string_lossy(),
        "budget_ms": budget.as_millis() as u64,
        "partial": !partial_classes.is_empty(),
        "partial_classes": partial_classes,
        "size_basis": if cfg!(unix) { "allocated_blocks_inode_once" } else { "apparent_length" },
        "disk": disk(&primary),
        "classes": classes,
        "worktrees": worktrees,
        "candidates": {
            "status": "not_evaluated",
            "reason": "removal candidates need the in-use checks of worktree.gc.plan (WT-4) and the user decision Q4",
        },
        "network": "disabled",
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn build_folder_prefix_wins_over_the_audit_pattern() {
        // #2563: the new `moai-target-` and the former `moai-rc-target-` are both build
        // folders; a name that only resembles them is not.
        for name in [
            "moai-target-agent-2-t61-audit-2326",
            "moai-rc-target-agent-2-t61-audit-2326",
        ] {
            assert_eq!(
                classify_tmp(name, false, true),
                Some(TmpClass::Target),
                "{name}"
            );
        }
        for name in [
            "moai-targets-x",
            "moai-product-target",
            "moai-rc-product-target",
        ] {
            assert_eq!(
                classify_tmp(name, false, false),
                Some(TmpClass::Other),
                "{name}"
            );
        }
        assert_eq!(
            classify_tmp("moai-audit-2369-x", false, true),
            Some(TmpClass::AuditClone)
        );
        // Without a git directory an audit-named folder is not counted as a clone.
        assert_eq!(
            classify_tmp("moai-audit-2369-x", false, false),
            Some(TmpClass::Other)
        );
        assert_eq!(
            classify_tmp("moai-cli-home-abc", true, false),
            Some(TmpClass::EmptyDir)
        );
        // A non-empty leftover of an empty-folder prefix is not empty.
        assert_eq!(
            classify_tmp("moai-cli-home-abc", false, false),
            Some(TmpClass::Other)
        );
        assert_eq!(classify_tmp("other-tool", false, false), None);
    }

    #[test]
    fn porcelain_list_is_parsed_record_by_record() {
        let text = "worktree /r\nHEAD aaa\nbranch refs/heads/main\n\nworktree /r/w1\nHEAD bbb\ndetached\nlocked moai claim x e1\n\nworktree /r/w2\nHEAD ccc\nbranch refs/heads/issue/1-k\nprunable gitdir file points to non-existent location\n";
        let e = parse_worktree_list(text);
        assert_eq!(e.len(), 3);
        assert_eq!(e[0].branch.as_deref(), Some("main"));
        assert!(e[1].detached && e[1].locked);
        assert_eq!(e[1].lock_reason.as_deref(), Some("moai claim x e1"));
        assert!(e[2].prunable && !e[2].locked);
    }
}
