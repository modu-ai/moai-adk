//! Fixed worktree slots (SPEC-20261123-worktree-slots slice 1; user decision 2026-11-22).
//!
//! Worktrees are the fixed slots `.claude/worktrees/lane-01..10`, `audit-1..3` and `seed` of
//! the primary checkout; only subagents use one-off worktrees. [`plan`] reads the slots and the
//! caps and writes nothing. [`ensure`] adds a missing slot detached at a commit with
//! `git worktree add --detach` (never forced) and, when the slot has no `target/` while the
//! seed has one, warms it with a copy of the seed's `target/`: an APFS clone (`cp -c -R -p`) on
//! macOS, which costs almost no disk and keeps modification times so Cargo reports the
//! dependencies `Fresh` and recompiles only the workspace crate; a plain `cp -R -p` elsewhere.
//! A slot's own `target/` is never copied into or over.
//!
//! One `ensure` runs at a time per repository: an exclusive lock on a file in the git common
//! directory covers the check, the add and the warm, so a second run for the same slot finds
//! it present. The copy goes to a temporary name beside the slot's `target/` and is renamed
//! into place only when `cp` succeeded, so a failed copy never looks like a warm build.
//!
//! Caps, checked before a slot is added: at most [`MAX_WORKTREES`] linked worktrees and at
//! least [`MIN_FREE_PERCENT`] of the disk free (from `df -P -k`, which, unlike `du`, sees the
//! blocks clones share). The 300 GB cap on `.claude/worktrees` is reported `measured: false`:
//! `du` counts every clone in full, so it waits for the cleanup slice (SPEC D-4).

use crate::identity::scrub_git_env;
use crate::worktree_report::{parse_worktree_list, WorktreeEntry};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

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
/// The lock file, in the git common directory, that serializes `ensure`.
pub const LOCK_FILE: &str = "moai-worktree-slot.lock";
/// How long `ensure` waits for another run's lock before it refuses.
pub const LOCK_WAIT: Duration = Duration::from_secs(300);
/// Shortens [`LOCK_WAIT`] to this many seconds, for tests.
pub const LOCK_WAIT_ENV: &str = "MOAI_WORKTREE_SLOT_LOCK_WAIT_SECS";
/// The temporary folders a warm or an inherit leaves only when it was killed mid-copy.
pub(crate) const TEMP_PREFIXES: [&str; 3] =
    [".target-warming-", ".target-inherit-", ".target-old-"];

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

/// The line of git's stderr that says why it failed: the first `fatal:` or `error:` line,
/// else the last non-empty one (the first is often a progress line such as
/// `Preparing worktree`).
fn git_reason(stderr: &str) -> &str {
    let lines: Vec<&str> = stderr
        .lines()
        .map(str::trim)
        .filter(|l| !l.is_empty())
        .collect();
    lines
        .iter()
        .find(|l| l.starts_with("fatal:") || l.starts_with("error:"))
        .or(lines.last())
        .copied()
        .unwrap_or("")
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
        return Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            git_reason(&err)
        ));
    }
    Ok(String::from_utf8_lossy(&out.stdout).trim().to_string())
}

/// The git common directory of the repository that holds `project`.
pub(crate) fn common_dir(project: &Path) -> Result<PathBuf> {
    git(
        project,
        &["rev-parse", "--path-format=absolute", "--git-common-dir"],
    )
    .map(PathBuf::from)
}

/// The primary checkout of the repository whose common directory is `common`.
pub(crate) fn primary_of(common: &Path) -> Result<PathBuf> {
    let root = common
        .parent()
        .ok_or("the repository has no primary checkout")?;
    Ok(fs::canonicalize(root).unwrap_or_else(|_| root.to_path_buf()))
}

pub(crate) fn same(a: &Path, b: &Path) -> bool {
    let canon = |p: &Path| fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf());
    canon(a) == canon(b)
}

/// The linked worktrees (the primary checkout left out).
pub(crate) fn linked(primary: &Path) -> Result<Vec<WorktreeEntry>> {
    let text = git(primary, &["worktree", "list", "--porcelain"])?;
    Ok(parse_worktree_list(&text)
        .into_iter()
        .filter(|e| !same(&e.path, primary))
        .collect())
}

/// The free share of the disk that holds `path`, in whole percent, from `df -P -k`. A `df`
/// that fails or prints no block counts refuses.
fn free_percent(path: &Path) -> Result<u64> {
    let out = Command::new("df")
        .args(["-P", "-k"])
        .arg(path)
        .stdin(Stdio::null())
        .output()
        .map_err(|e| format!("df did not start: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "df failed ({}); the free disk share is unknown",
            out.status
        ));
    }
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
                         "ok": free >= caps.min_free_percent, "source": "df -P -k"},
        "slots_gb": {"limit": SLOTS_GB, "measured": false,
                     "why": "du counts every APFS clone in full; measured by the cleanup slice"},
    })
}

fn slot_json(name: &str, path: &Path, entry: Option<&WorktreeEntry>) -> Value {
    let rel = format!("{SLOT_DIR}/{name}");
    match entry {
        Some(e) => json!({
            "name": name, "path": rel, "exists": true,
            // Registered, but its folder is gone: `git worktree prune` removes the record.
            "missing": !path.is_dir(),
            "prunable": e.prunable,
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
    let primary = primary_of(&common_dir(project)?)?;
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

/// Whether `path` is its own linked worktree of the repository whose git common dir is
/// `common`: git run in it names `path` as the top level and `common` as the common dir, and
/// its private git dir points back at `path/.git`. A plain folder (git walks up to the primary
/// checkout), another repository, or a `.git` file that borrows another worktree's git dir
/// (git would use that worktree's index) each fail, with the reason.
pub(crate) fn own_worktree(path: &Path, common: &Path) -> Result<()> {
    let canon =
        |p: &Path| fs::canonicalize(p).map_err(|e| format!("{} unreadable: {e}", p.display()));
    let me = canon(path)?;
    let top = git(path, &["rev-parse", "--show-toplevel"])
        .map_err(|e| format!("git cannot read it: {e}"))?;
    if canon(Path::new(&top))? != me {
        return Err(format!("git in it works on {top}"));
    }
    let shared = git(
        path,
        &["rev-parse", "--path-format=absolute", "--git-common-dir"],
    )
    .map_err(|e| format!("git cannot read it: {e}"))?;
    if canon(Path::new(&shared))? != canon(common)? {
        return Err(format!("it belongs to the repository at {shared}"));
    }
    let own = git(path, &["rev-parse", "--path-format=absolute", "--git-dir"])
        .map_err(|e| format!("git cannot read it: {e}"))?;
    let back = fs::read_to_string(Path::new(&own).join("gitdir"))
        .map_err(|_| format!("its git dir {own} names no linked worktree"))?;
    if canon(Path::new(back.trim())).ok() != Some(me.join(".git")) {
        return Err(format!(
            "its git dir {own} belongs to the worktree at {}",
            back.trim()
        ));
    }
    Ok(())
}

/// Takes the repository's exclusive `ensure` lock, waiting at most [`LOCK_WAIT`].
pub(crate) fn lock(common: &Path) -> Result<fs::File> {
    let file = fs::OpenOptions::new()
        .create(true)
        .truncate(false)
        .write(true)
        .open(common.join(LOCK_FILE))
        .map_err(|e| format!("the slot lock could not be opened: {e}"))?;
    let wait = std::env::var(LOCK_WAIT_ENV)
        .ok()
        .and_then(|v| v.parse::<u64>().ok())
        .map_or(LOCK_WAIT, Duration::from_secs);
    let deadline = Instant::now() + wait;
    loop {
        match file.try_lock() {
            Ok(()) => return Ok(file),
            Err(fs::TryLockError::WouldBlock) if Instant::now() < deadline => {
                std::thread::sleep(Duration::from_millis(100));
            }
            Err(fs::TryLockError::WouldBlock) => {
                return Err(format!(
                    "another worktree-slot or audit-slot run held the slot lock for {} seconds; nothing was changed",
                    wait.as_secs()
                ))
            }
            Err(fs::TryLockError::Error(e)) => {
                return Err(format!("the slot lock could not be taken: {e}"))
            }
        }
    }
}

/// The seed's `target/` to warm from, or `None` when there is no seed build. A seed folder that
/// is not a worktree of this repository, or a seed `target` that is not a real folder (a link
/// to a shared build folder, a file), refuses: copying it would share a build folder.
fn seed_target(primary: &Path, linked: &[WorktreeEntry]) -> Result<Option<PathBuf>> {
    let seed = primary.join(SLOT_DIR).join(SEED);
    if fs::symlink_metadata(&seed).is_err() {
        return Ok(None);
    }
    if !linked.iter().any(|e| same(&e.path, &seed)) {
        return Err(format!(
            "{SLOT_DIR}/{SEED} exists but is not a worktree of this repository; nothing was added"
        ));
    }
    let target = seed.join("target");
    match fs::symlink_metadata(&target) {
        Err(_) => Ok(None),
        Ok(m) if m.file_type().is_dir() => Ok(Some(target)),
        Ok(_) => Err(format!(
            "{SLOT_DIR}/{SEED}/target is not a real folder (a link or a file); nothing was added"
        )),
    }
}

/// Copies the seed's `target/` to a temporary name beside the slot's `target/`, then renames
/// it into place. Returns the method; on failure the temporary copy is removed and the slot
/// has no `target/`.
fn warm(seed_target: &Path, slot: &Path) -> Result<&'static str> {
    let (method, flags): (&str, &[&str]) = if cfg!(target_os = "macos") {
        ("clone", &["-c", "-R", "-p"])
    } else {
        ("copy", &["-R", "-p"])
    };
    let temp = slot.join(format!(".target-warming-{}", std::process::id()));
    let status = Command::new("cp")
        .args(flags)
        .arg(seed_target)
        .arg(&temp)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .map_err(|e| format!("cp did not start: {e}"))?;
    let renamed = if status.success() {
        fs::rename(&temp, slot.join("target"))
            .map_err(|e| format!("the warm copy could not be put in place: {e}"))
    } else {
        Err(format!("cp of the seed's target/ ended with {status}"))
    };
    if renamed.is_err() {
        // Only this run's own temporary copy inside the slot.
        let _ = fs::remove_dir_all(&temp);
    }
    renamed.map(|()| method)
}

/// Removes the temporary copies a killed warm or inherit left in a present slot. The caller holds
/// the slot lock, so no other run is copying into the slot now. Returns the names removed.
fn clean_leftovers(slot: &Path) -> Result<Vec<String>> {
    let mut removed = Vec::new();
    let entries = fs::read_dir(slot).map_err(|e| format!("{} unreadable: {e}", slot.display()))?;
    for e in entries.flatten() {
        let name = e.file_name().to_string_lossy().to_string();
        if TEMP_PREFIXES.iter().any(|p| name.starts_with(p)) {
            let p = e.path();
            let gone = match fs::symlink_metadata(&p) {
                Ok(m) if m.file_type().is_dir() => fs::remove_dir_all(&p),
                _ => fs::remove_file(&p),
            };
            gone.map_err(|e| format!("{} could not be removed: {e}", p.display()))?;
            removed.push(name);
        }
    }
    removed.sort();
    Ok(removed)
}

/// Adds the slot `name` detached at `at` when it is missing, and warms a slot without
/// `target/` from the seed. A present slot keeps its checkout; one without `target/` is warmed.
pub fn ensure(project: &Path, name: &str, at: &str, caps: Caps) -> Result<Value> {
    if !is_slot(name) {
        return Err(
            "invalid slot name: one of lane-01 to lane-10, audit-1 to audit-3 or seed".into(),
        );
    }
    let common = common_dir(project)?;
    let primary = primary_of(&common)?;
    let _lock = lock(&common)?;
    let path = primary.join(SLOT_DIR).join(name);
    let rel = format!("{SLOT_DIR}/{name}");
    if fs::symlink_metadata(&path).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err(format!(
            "{rel} is a link, not a slot folder; refused, nothing was changed"
        ));
    }
    let linked = linked(&primary)?;
    let present = linked.iter().any(|e| same(&e.path, &path));
    if present && !path.is_dir() {
        return Err(format!(
            "{rel} is registered but its folder is missing; `git worktree prune` removes the record; nothing was added"
        ));
    }
    if !present && fs::symlink_metadata(&path).is_ok() {
        return Err(format!(
            "{rel} exists but is not a worktree of this repository; nothing was added"
        ));
    }
    if present {
        // A registered path holding a plain folder, another repository or a borrowed `.git`
        // would take the seed copy, and git in it would act on another checkout.
        own_worktree(&path, &common).map_err(|e| {
            format!("{rel} is registered but is not its own worktree ({e}); refused, nothing was changed")
        })?;
    }
    let seed = if name == SEED {
        None
    } else {
        seed_target(&primary, &linked)?
    };
    // An audit slot's leftover copy is audit-slot's to name and to discard with `done --discard`,
    // so `ensure` leaves it there for `audit_slot` to refuse on.
    let cleaned = if present && !name.starts_with("audit-") {
        clean_leftovers(&path)?
    } else {
        Vec::new()
    };
    let needs_warm = seed.is_some() && fs::symlink_metadata(path.join("target")).is_err();
    let mut commit = Value::Null;
    if present && needs_warm {
        // A warm copy takes disk like an added slot does, so it meets the same floor.
        let free = free_percent(&primary)?;
        if free < caps.min_free_percent {
            return Err(format!(
                "free disk under the floor: limit {} percent, free {free} percent; {rel} was not warmed",
                caps.min_free_percent
            ));
        }
    }
    if !present {
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
        let sha = git(
            &primary,
            &[
                "rev-parse",
                "--verify",
                "--quiet",
                "--end-of-options",
                &format!("{at}^{{commit}}"),
            ],
        )
        .map_err(|_| format!("{at:?} names no commit; nothing was added"))?;
        let path_arg = path.to_string_lossy().to_string();
        git(&primary, &["worktree", "add", "--detach", &path_arg, &sha])?;
        commit = json!(sha);
    }
    // A slot's own `target/` (a warm build, or one its checkout tracks) is never copied into.
    let warmed = match seed {
        Some(seed) if fs::symlink_metadata(path.join("target")).is_err() => warm(&seed, &path)
            .map_err(|e| {
                let state = if present { "present" } else { "added" };
                format!("{rel} is {state} but warming failed: {e}; it has no target/")
            })?,
        _ => "none",
    };
    Ok(json!({
        "slot": name, "path": rel, "at": commit, "added": !present, "present": present,
        "warmed": warmed, "cleaned": cleaned,
    }))
}

#[cfg(test)]
mod tests {
    use super::git_reason;

    #[test]
    fn git_reason_keeps_the_fatal_line_over_the_progress_line() {
        let err = "Preparing worktree (detached HEAD 1234567)\nfatal: '/x' already exists\n";
        assert_eq!(git_reason(err), "fatal: '/x' already exists");
        assert_eq!(git_reason("error: bad\n"), "error: bad");
        assert_eq!(git_reason("one\n\ntwo\n\n"), "two");
        assert_eq!(git_reason(""), "");
    }
}
