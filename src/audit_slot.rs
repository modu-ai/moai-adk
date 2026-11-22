//! Audit slots that inherit the lane's build (SPEC-20261123-audit-slot; user decision 2026-11-22).
//!
//! An audit runs in one of the fixed slots `.claude/worktrees/audit-1..3`, never in a new
//! isolated worktree and never in a shared build folder. [`prep`] adds a missing slot through
//! [`crate::worktree_slot::ensure`], switches it detached to the audited head and places a
//! build in it: an APFS clone (`cp -c -R -p`, a plain `cp -R -p` elsewhere) of the lane's
//! `target/` when the lane's worktree is at that head with no tracked change, else the slot's
//! own build or the seed's. A clone keeps modification times, so Cargo reports the copied
//! dependencies `Fresh` and an audit's mutant recompiles only the workspace crate. [`done`]
//! reverts what the audit changed in the slot and keeps its `target/` for the next audit.
//!
//! Both hold the slot lock of [`crate::worktree_slot`] while they change a slot, so they never
//! run beside an `ensure` or each other on one repository.

use crate::spec_take::git;
use crate::worktree_slot::{self, Caps, SLOT_DIR};
use serde_json::{json, Value};
use std::fs;
use std::path::{Component, Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The slots an audit may take.
pub const AUDIT_SLOTS: [&str; 3] = ["audit-1", "audit-2", "audit-3"];

fn check_name(name: &str) -> Result<()> {
    if AUDIT_SLOTS.contains(&name) {
        Ok(())
    } else {
        Err(format!(
            "audit-slot refused {name:?}: invalid slot name: one of audit-1, audit-2 or audit-3"
        ))
    }
}

/// Runs git and names the step on failure (`git failed` reads as an I/O error).
fn git_step(dir: &Path, step: &str, args: &[&str]) -> Result<String> {
    git(dir, args).map_err(|e| format!("git failed ({step}): {e}"))
}

/// The full commit `head` names, or a refusal.
fn commit_of(dir: &Path, head: &str) -> Result<String> {
    git(
        dir,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            "--end-of-options",
            &format!("{head}^{{commit}}"),
        ],
    )
    .map_err(|_| format!("audit-slot refused: {head:?} names no commit; nothing was changed"))
}

/// One changed path of `git status`: tracked (staged or not) or untracked.
struct Change {
    path: String,
    tracked: bool,
}

/// The slot's changes, `target/` left out (the slot's build is never a change). The format is
/// `--porcelain=v2 -z`: every entry starts with its type letter, so the trimmed output of
/// [`git`] loses nothing (a v1 entry can start with a space).
fn changes(slot: &Path) -> Result<Vec<Change>> {
    let text = git_step(
        slot,
        "status",
        &["status", "--porcelain=v2", "-z", "--untracked-files=all"],
    )?;
    let mut out = Vec::new();
    let mut fields = text.split('\0').filter(|f| !f.is_empty());
    while let Some(f) = fields.next() {
        // The path is the last of a fixed number of space-separated fields per entry type.
        let (path, tracked) = match f.as_bytes().first() {
            Some(b'1') => (f.splitn(9, ' ').nth(8), true),
            Some(b'2') => {
                // A rename or copy is followed by its source path, not a change of its own.
                fields.next();
                (f.splitn(10, ' ').nth(9), true)
            }
            Some(b'u') => (f.splitn(11, ' ').nth(10), true),
            Some(b'?') => (f.get(2..), false),
            _ => (None, false),
        };
        let Some(path) = path else { continue };
        if path == "target" || path.starts_with("target/") {
            continue;
        }
        out.push(Change {
            path: path.to_string(),
            tracked,
        });
    }
    Ok(out)
}

fn names(list: &[Change]) -> String {
    list.iter()
        .map(|c| c.path.as_str())
        .collect::<Vec<_>>()
        .join(", ")
}

/// The lane's `target/` to inherit: the lane must be a worktree of this repository at `sha`
/// with no tracked change, and its `target` a real folder (not a link to a shared one).
fn lane_target(lane: &Path, primary: &Path, sha: &str) -> Result<PathBuf> {
    let shown = lane.display();
    let registered = std::iter::once(primary.to_path_buf())
        .chain(worktree_slot::linked(primary)?.into_iter().map(|e| e.path))
        .any(|p| worktree_slot::same(&p, lane));
    if !registered {
        return Err(format!(
            "audit-slot refused the lane {shown}: not a worktree of this repository; nothing was changed"
        ));
    }
    let at = git_step(lane, "lane head", &["rev-parse", "HEAD"])?;
    if at != sha {
        return Err(format!(
            "audit-slot refused the lane {shown}: it is at {at}, not at the audited head {sha}; nothing was changed"
        ));
    }
    let dirty: Vec<Change> = changes(lane)?.into_iter().filter(|c| c.tracked).collect();
    if !dirty.is_empty() {
        return Err(format!(
            "audit-slot refused the lane {shown}: a tracked change ({}) means its build is not the head's; nothing was changed",
            names(&dirty)
        ));
    }
    let target = lane.join("target");
    match fs::symlink_metadata(&target) {
        Err(_) => Err(format!(
            "audit-slot refused the lane {shown}: it has no target/ to inherit; nothing was changed"
        )),
        Ok(m) if m.file_type().is_dir() => Ok(target),
        Ok(_) => Err(format!(
            "audit-slot refused the lane {shown}: its target is not a real folder (a link or a file); nothing was changed"
        )),
    }
}

/// Clones `from` to a temporary name in `slot`, re-reads the lane's head, then swaps it in for
/// the slot's `target/`. A failure leaves the slot's old `target/` as it was.
fn inherit(from: &Path, lane: &Path, sha: &str, slot: &Path) -> Result<&'static str> {
    let (method, flags): (&str, &[&str]) = if cfg!(target_os = "macos") {
        ("clone", &["-c", "-R", "-p"])
    } else {
        ("copy", &["-R", "-p"])
    };
    let pid = std::process::id();
    let temp = slot.join(format!(".target-inherit-{pid}"));
    let status = Command::new("cp")
        .args(flags)
        .arg(from)
        .arg(&temp)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .map_err(|e| format!("cp did not start: {e}"))?;
    let checked = if !status.success() {
        Err(format!("cp of the lane's target/ ended with {status}"))
    } else {
        match git_step(lane, "lane head", &["rev-parse", "HEAD"]) {
            Ok(at) if at == sha => Ok(()),
            Ok(at) => Err(format!(
                "audit-slot refused: the lane moved to {at} while its build was copied"
            )),
            Err(e) => Err(e),
        }
    };
    if let Err(e) = checked {
        // Only this run's own temporary copy inside the slot.
        let _ = fs::remove_dir_all(&temp);
        return Err(e);
    }
    let target = slot.join("target");
    let old = slot.join(format!(".target-old-{pid}"));
    let had = fs::symlink_metadata(&target).is_ok();
    if had {
        if let Err(e) = fs::rename(&target, &old) {
            let _ = fs::remove_dir_all(&temp);
            return Err(format!("the slot's target/ could not be set aside: {e}"));
        }
    }
    if let Err(e) = fs::rename(&temp, &target) {
        if had {
            let _ = fs::rename(&old, &target);
        }
        let _ = fs::remove_dir_all(&temp);
        return Err(format!(
            "the inherited build could not be put in place: {e}"
        ));
    }
    if had {
        fs::remove_dir_all(&old)
            .map_err(|e| format!("the slot's old target/ could not be removed: {e}"))?;
    }
    Ok(method)
}

/// Takes the audit slot `name` for `head`: adds it when missing, refuses it when it holds
/// changes, switches it detached to the head and places a build (see the module notes).
pub fn prep(
    project: &Path,
    name: &str,
    head: &str,
    lane: Option<&Path>,
    caps: Caps,
) -> Result<Value> {
    check_name(name)?;
    let common = worktree_slot::common_dir(project)?;
    let primary = worktree_slot::primary_of(&common)?;
    let sha = commit_of(&primary, head)?;
    let from = match lane {
        Some(lane) => Some(lane_target(lane, &primary, &sha)?),
        None => None,
    };
    let ensured = worktree_slot::ensure(project, name, &sha, caps)?;
    let added = ensured["added"] == json!(true);
    let seeded = ensured["warmed"] != json!("none");
    let _lock = worktree_slot::lock(&common)?;
    let slot = primary.join(SLOT_DIR).join(name);
    let rel = format!("{SLOT_DIR}/{name}");
    let dirty = changes(&slot)?;
    if !dirty.is_empty() {
        return Err(format!(
            "audit-slot refused {rel}: it holds changes of an earlier audit ({}); run `done {name} --discard` first; nothing was changed",
            names(&dirty)
        ));
    }
    git_step(&slot, "switch", &["switch", "--quiet", "--detach", &sha])?;
    let build = match (from, lane) {
        (Some(from), Some(lane)) => {
            let method = inherit(&from, lane, &sha, &slot).map_err(|e| {
                format!("{rel} is at {sha} but inheriting the lane build failed: {e}")
            })?;
            json!({"source": "lane", "method": method, "from": lane.display().to_string()})
        }
        _ if seeded => json!({"source": "seed", "method": ensured["warmed"]}),
        _ if slot.join("target").is_dir() => json!({"source": "kept", "method": null}),
        _ => json!({"source": "none", "method": null}),
    };
    let at = git_step(&slot, "head", &["rev-parse", "HEAD"])?;
    let clean = changes(&slot)?.is_empty();
    Ok(json!({
        "slot": name, "path": rel, "head": at, "added": added, "build": build, "clean": clean,
    }))
}

/// A path git printed, checked to stay inside the slot.
fn inside(slot: &Path, rel: &str) -> Result<PathBuf> {
    let p = Path::new(rel);
    if p.components().all(|c| matches!(c, Component::Normal(_))) {
        Ok(slot.join(p))
    } else {
        Err(format!(
            "audit-slot refused: git named a path outside the slot ({rel:?})"
        ))
    }
}

/// Finishes an audit in the slot `name`: a clean slot is done; one with changes is refused
/// unless `discard`, which reverts tracked changes, removes untracked files outside `target/`
/// and keeps `target/`.
pub fn done(project: &Path, name: &str, discard: bool) -> Result<Value> {
    check_name(name)?;
    let common = worktree_slot::common_dir(project)?;
    let primary = worktree_slot::primary_of(&common)?;
    let slot = primary.join(SLOT_DIR).join(name);
    let rel = format!("{SLOT_DIR}/{name}");
    let registered = worktree_slot::linked(&primary)?
        .iter()
        .any(|e| worktree_slot::same(&e.path, &slot));
    if !registered || !slot.is_dir() {
        return Err(format!(
            "audit-slot refused {rel}: not a worktree of this repository (prep adds it)"
        ));
    }
    let _lock = worktree_slot::lock(&common)?;
    let found = changes(&slot)?;
    if !found.is_empty() && !discard {
        return Err(format!(
            "audit-slot refused {rel}: it holds changes ({}); `done {name} --discard` reverts and removes them; nothing was changed",
            names(&found)
        ));
    }
    let (tracked, untracked): (Vec<Change>, Vec<Change>) =
        found.into_iter().partition(|c| c.tracked);
    let paths: Vec<PathBuf> = untracked
        .iter()
        .map(|c| inside(&slot, &c.path))
        .collect::<Result<_>>()?;
    if !tracked.is_empty() {
        git_step(&slot, "reset", &["reset", "--quiet", "--hard", "HEAD"])?;
    }
    for p in &paths {
        let removed = match fs::symlink_metadata(p) {
            Ok(m) if m.file_type().is_dir() => fs::remove_dir_all(p),
            Ok(_) => fs::remove_file(p),
            Err(_) => Ok(()),
        };
        removed.map_err(|e| format!("{} could not be removed: {e}", p.display()))?;
    }
    let left = changes(&slot)?;
    let head = git_step(&slot, "head", &["rev-parse", "HEAD"])?;
    Ok(json!({
        "slot": name, "path": rel, "head": head,
        "reverted": tracked.iter().map(|c| c.path.clone()).collect::<Vec<_>>(),
        "removed": untracked.iter().map(|c| c.path.clone()).collect::<Vec<_>>(),
        "clean": left.is_empty(),
        "target_kept": slot.join("target").is_dir(),
    }))
}
