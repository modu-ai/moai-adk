//! Fixed worktree slots (SPEC-20261123-worktree-slots slice 1; user decision 2026-11-22,
//! ledger notes 6628 and 6646).
//!
//! Worktrees are the fixed slots `.claude/worktrees/lane-01..10`, `audit-1..3` and `seed` of
//! the primary checkout; only subagents use one-off worktrees. [`plan`] reads the slots and the
//! caps and writes nothing. [`ensure`] adds a missing slot detached at a commit with
//! `git worktree add --detach` (never forced) and, when the slot has no `target/` while the
//! seed has one, warms it with a copy of the seed's `target/`: an APFS clone (`cp -c -R -p`) on
//! macOS, which costs almost no disk and keeps modification times so Cargo reports the
//! dependencies `Fresh` and recompiles only the workspace crate; a plain `cp -R -p` elsewhere.
//! A present slot is left as it is, its own `target/` included.
//!
//! Caps, checked before a slot is added: at most [`MAX_WORKTREES`] linked worktrees and at
//! least [`MIN_FREE_PERCENT`] of the disk free (from `df`, which, unlike `du`, sees the blocks
//! clones share). The 300 GB cap on `.claude/worktrees` is reported `measured: false`: `du`
//! counts every clone in full, so it waits for the cleanup slice (SPEC D-4).

use crate::identity::scrub_git_env;
use crate::worktree_report::{parse_worktree_list, WorktreeEntry};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The fixed slots, in the order the plan lists them.
pub const SLOTS: [&str; 14] = [
    "lane-01", "lane-02", "lane-03", "lane-04", "lane-05", "lane-06", "lane-07", "lane-08",
    "lane-09", "lane-10", "audit-1", "audit-2", "audit-3", "seed",
];
/// The slots' folder, relative to the primary checkout (Claude Code's own worktree place).
pub const SLOT_DIR: &str = ".claude/worktrees";
/// The slot whose `target/` warms the others.
pub const SEED: &str = "seed";
/// Linked worktrees allowed on the repository before a slot is refused.
pub const MAX_WORKTREES: usize = 25;
/// Free share of the disk, in percent, below which a slot is refused.
pub const MIN_FREE_PERCENT: u64 = 15;
/// The size cap of `.claude/worktrees`, reported but not yet enforced (SPEC D-4).
pub const SLOTS_GB: u64 = 300;

/// The limits a slot is checked against; the defaults are the user decision's.
#[derive(Clone, Copy, Debug)]
pub struct Caps {
    pub max_worktrees: usize,
    pub min_free_percent: u64,
}

impl Default for Caps {
    fn default() -> Self {
        Caps {
            max_worktrees: MAX_WORKTREES,
            min_free_percent: MIN_FREE_PERCENT,
        }
    }
}

/// Whether `name` is one of the fixed slots.
pub fn is_slot(name: &str) -> bool {
    SLOTS.contains(&name)
}

fn git(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    cmd.args(args)
        .current_dir(dir)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    scrub_git_env(&mut cmd);
    let out = cmd
        .output()
        .map_err(|e| format!("git did not start: {e}"))?;
    if !out.status.success() {
        let err = String::from_utf8_lossy(&out.stderr);
        let first = err.lines().next().unwrap_or("").trim();
        return Err(format!(
            "git {} failed: {first}",
            args.first().copied().unwrap_or("")
        ));
    }
    Ok(String::from_utf8_lossy(&out.stdout).trim().to_string())
}

/// The primary checkout of the repository that holds `project`.
fn primary(project: &Path) -> Result<PathBuf> {
    let common = git(
        project,
        &["rev-parse", "--path-format=absolute", "--git-common-dir"],
    )?;
    let common = PathBuf::from(common);
    let root = common
        .parent()
        .ok_or("the repository has no primary checkout")?;
    Ok(fs::canonicalize(root).unwrap_or_else(|_| root.to_path_buf()))
}

fn same(a: &Path, b: &Path) -> bool {
    let canon = |p: &Path| fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf());
    canon(a) == canon(b)
}

/// The linked worktrees (the primary checkout left out).
fn linked(primary: &Path) -> Result<Vec<WorktreeEntry>> {
    let text = git(primary, &["worktree", "list", "--porcelain"])?;
    Ok(parse_worktree_list(&text)
        .into_iter()
        .filter(|e| !same(&e.path, primary))
        .collect())
}

/// The free share of the disk that holds `path`, in whole percent, from `df -P -k`.
fn free_percent(path: &Path) -> Result<u64> {
    let out = Command::new("df")
        .args(["-P", "-k"])
        .arg(path)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("df did not start: {e}"))?;
    let text = String::from_utf8_lossy(&out.stdout);
    let line = text.lines().nth(1).ok_or("df printed no disk line")?;
    let cols: Vec<&str> = line.split_whitespace().collect();
    let num = |i: usize| -> Result<u64> {
        cols.get(i)
            .and_then(|c| c.parse::<u64>().ok())
            .ok_or_else(|| "df printed no block counts".to_string())
    };
    let (total, avail) = (num(1)?, num(3)?);
    if total == 0 {
        return Err("df printed an empty disk".into());
    }
    Ok(avail.saturating_mul(100) / total)
}

fn caps_json(count: usize, free: u64, caps: Caps) -> Value {
    json!({
        "worktrees": {"limit": caps.max_worktrees, "count": count,
                      "ok": count < caps.max_worktrees},
        "free_percent": {"limit": caps.min_free_percent, "value": free,
                         "ok": free >= caps.min_free_percent},
        "slots_gb": {"limit": SLOTS_GB, "measured": false,
                     "why": "du counts every APFS clone in full; measured by the cleanup slice"},
    })
}

fn slot_json(name: &str, path: &Path, entry: Option<&WorktreeEntry>) -> Value {
    let rel = format!("{SLOT_DIR}/{name}");
    match entry {
        Some(e) => json!({
            "name": name, "path": rel, "exists": true,
            "head": e.head,
            "branch": e.branch.as_deref()
                .map(|b| b.strip_prefix("refs/heads/").unwrap_or(b).to_string())
                .unwrap_or_else(|| "detached".into()),
            "locked": e.locked,
            "target": path.join("target").is_dir(),
        }),
        None => json!({
            "name": name, "path": rel, "exists": false,
            "folder": path.exists(),
        }),
    }
}

/// The fixed slots and the caps; writes nothing.
pub fn plan(project: &Path, caps: Caps) -> Result<Value> {
    let primary = primary(project)?;
    let linked = linked(&primary)?;
    let slots: Vec<Value> = SLOTS
        .iter()
        .map(|name| {
            let path = primary.join(SLOT_DIR).join(name);
            let entry = linked.iter().find(|e| same(&e.path, &path));
            slot_json(name, &path, entry)
        })
        .collect();
    let free = free_percent(&primary)?;
    Ok(json!({
        "slots": slots,
        "caps": caps_json(linked.len(), free, caps),
        "written": false,
    }))
}

/// Copies the seed's `target/` into a slot that has none: an APFS clone on macOS.
fn warm(seed_target: &Path, slot_target: &Path) -> Result<&'static str> {
    let (method, flags): (&str, &[&str]) = if cfg!(target_os = "macos") {
        ("clone", &["-c", "-R", "-p"])
    } else {
        ("copy", &["-R", "-p"])
    };
    let status = Command::new("cp")
        .args(flags)
        .arg(seed_target)
        .arg(slot_target)
        .stdin(Stdio::null())
        .status()
        .map_err(|e| format!("cp did not start: {e}"))?;
    if !status.success() {
        return Err(format!("cp of the seed's target/ ended with {status}"));
    }
    Ok(method)
}

/// Adds the slot `name` detached at `at` when it is missing, and warms it from the seed.
pub fn ensure(project: &Path, name: &str, at: &str, caps: Caps) -> Result<Value> {
    if !is_slot(name) {
        return Err(
            "invalid slot name: one of lane-01 to lane-10, audit-1 to audit-3 or seed".into(),
        );
    }
    let primary = primary(project)?;
    let path = primary.join(SLOT_DIR).join(name);
    let rel = format!("{SLOT_DIR}/{name}");
    let linked = linked(&primary)?;
    if linked.iter().any(|e| same(&e.path, &path)) {
        return Ok(json!({
            "slot": name, "path": rel, "added": false, "present": true, "warmed": "none",
        }));
    }
    if path.exists() {
        return Err(format!(
            "{rel} exists but is not a worktree of this repository; nothing was added"
        ));
    }
    if linked.len() >= caps.max_worktrees {
        return Err(format!(
            "worktree cap reached: limit {}, count {}; nothing was added",
            caps.max_worktrees,
            linked.len()
        ));
    }
    let free = free_percent(&primary)?;
    if free < caps.min_free_percent {
        return Err(format!(
            "free disk under the floor: limit {} percent, free {free} percent; nothing was added",
            caps.min_free_percent
        ));
    }
    let commit = git(
        &primary,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            &format!("{at}^{{commit}}"),
        ],
    )
    .map_err(|_| format!("{at:?} names no commit; nothing was added"))?;
    let path_arg = path.to_string_lossy().to_string();
    git(
        &primary,
        &["worktree", "add", "--detach", &path_arg, &commit],
    )?;
    let seed_target = primary.join(SLOT_DIR).join(SEED).join("target");
    let slot_target = path.join("target");
    let (warmed, warm_error) = if name != SEED && seed_target.is_dir() && !slot_target.exists() {
        match warm(&seed_target, &slot_target) {
            Ok(m) => (m, None),
            Err(e) => ("failed", Some(e)),
        }
    } else {
        ("none", None)
    };
    Ok(json!({
        "slot": name, "path": rel, "at": commit, "added": true, "present": false,
        "warmed": warmed, "warm_error": warm_error,
    }))
}
