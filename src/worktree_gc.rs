//! `worktree.gc.plan` (WT-4, Issue #2380; design `docs/worktree-strategy-2026-11-10.md`
//! §D5, §D8). Read-only: for every linked worktree it lists whether its cargo build output
//! may be trimmed or the whole worktree removed, with the reasons, and a digest of the list.
//! The plan deletes nothing. `worktree.gc.apply` ([`apply`], #2498) acts on a plan only when
//! the recomputed digest is the same; under the user decision of 2026-11-10 (#2389 comment
//! 6774899518, item 2) running it on the user's own machine is the leader's step after the
//! user confirms.
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
//! 6811382861 B2). `worktree.gc.apply` (#2498) re-checks open files right before it deletes;
//! claim leases still need the network and stay unchecked by both:
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
/// Smallest accepted window (#2498 N2): half the design default. A smaller value would turn
/// the recent-change protection off for work that pauses between builds; it is refused, not
/// clamped, so the caller sees it.
pub const MIN_RECENT_MINUTES: u64 = 60;
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
        // #2498 N3: a worktree with nothing in it but its checkout — never built, no files
        // of its own — looks exactly like one a lane has just created and not started; it
        // frees almost nothing, so it is kept.
        (Some(0), Some(0)) if !f.has_build_output => {
            reasons.push("empty_worktree");
            return ("skip", reasons);
        }
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

/// Splits the children of a `target/` folder into cargo-owned entries and others. A `target`
/// that is not a real folder — a symlink (possibly to a shared build folder outside the
/// worktree) or a file — has no cargo-owned entries and counts as one file outside the build
/// output, so the plan never trims through it (audit B1 on PR #2557).
pub fn split_target(target: &Path) -> (Vec<PathBuf>, Vec<String>) {
    let (mut cargo, mut other) = (Vec::new(), Vec::new());
    match fs::symlink_metadata(target) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => {
            other.push("target".to_string());
            return (cargo, other);
        }
        Err(_) => return (cargo, other),
    }
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

pub(crate) fn git_out(dir: &Path, args: &[&str]) -> Option<String> {
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

/// Squash-merge test (#2724), network-free: `Some(true)` when every path the branch changed
/// since its merge base with `tip` has, at `tip`, exactly the content it has at `head` (a
/// path the branch deleted is absent at `tip` too). `Some(false)` when one differs or the
/// branch changed nothing (an empty branch is not "merged by squash"); `None` when git
/// failed. Renames are off so a renamed file lists both its old and its new path.
pub fn squash_merged(primary: &Path, head: &str, tip: &str) -> Option<bool> {
    let base = git_out(primary, &["merge-base", head, tip])?;
    let base = base.trim();
    let changed = git_out(
        primary,
        &[
            "diff",
            "--no-renames",
            "--no-ext-diff",
            "--no-textconv",
            "--name-only",
            "-z",
            base,
            head,
        ],
    )?;
    let paths: Vec<String> = changed
        .split('\0')
        .filter(|p| !p.is_empty())
        .map(|p| format!(":(literal){p}"))
        .collect();
    if paths.is_empty() {
        return Some(false);
    }
    let mut cmd = Command::new("git");
    scrub_git_env(&mut cmd);
    let st = cmd
        .arg("--no-optional-locks")
        .arg("-C")
        .arg(primary)
        .args([
            "diff",
            "--quiet",
            "--no-renames",
            "--no-ext-diff",
            "--no-textconv",
            head,
            tip,
            "--",
        ])
        .args(&paths)
        .status()
        .ok()?;
    match st.code() {
        Some(0) => Some(true),
        Some(1) => Some(false),
        _ => None,
    }
}

/// Working directories of the processes this user can see; `None` without a source.
pub(crate) fn process_cwds() -> Option<Vec<PathBuf>> {
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

/// Paths of the files the processes this user can see hold open (`/proc/*/fd` on Linux,
/// `lsof -Fn` elsewhere), canonicalised; `None` without a source. Only `apply` reads this,
/// right before it deletes (#2498; the plan reads working directories only).
pub(crate) fn open_file_paths() -> Option<Vec<PathBuf>> {
    let mut out = Vec::new();
    if Path::new("/proc/self/fd").exists() {
        for e in fs::read_dir("/proc").ok()?.flatten() {
            if let Ok(fds) = fs::read_dir(e.path().join("fd")) {
                out.extend(fds.flatten().filter_map(|fd| fs::read_link(fd.path()).ok()));
            }
        }
    } else {
        let o = Command::new("lsof")
            .args(["-w", "-Fn"])
            .env("LC_ALL", "C")
            .output()
            .ok()?;
        if o.stdout.is_empty() {
            return None;
        }
        out.extend(
            String::from_utf8_lossy(&o.stdout)
                .lines()
                .filter_map(|l| l.strip_prefix('n'))
                .filter(|l| l.starts_with('/'))
                .map(PathBuf::from),
        );
    }
    Some(
        out.into_iter()
            .map(|p| fs::canonicalize(&p).unwrap_or(p))
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
pub(crate) fn census(wt: &Path) -> Option<(usize, usize, bool)> {
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

/// Ignored files a merged worktree may lose (#2879 D2) — the one table. Regenerable caches,
/// a lane's temporary files and byte-identical copies of the primary checkout's same path
/// (what `.worktreeinclude` copied). Anything else keeps the worktree: a `.moai/` report that
/// exists only in the worktree, or differs from the primary's copy, is user data (measured
/// 2026-11-14 on 41 merged worktrees: 34 held a differing file, 6 a file the primary lacks).
const REGENERABLE_DIRS: &[&str] = &["__pycache__", ".ruff_cache", ".pytest_cache", ".mypy_cache"];
const REGENERABLE_EXTS: &[&str] = &["pyc", "pyo"];
const LANE_TEMP_EXTS: &[&str] = &["log"];
/// More ignored files than this in one worktree and the census is unknown (fail closed).
const IGNORED_FILE_LIMIT: usize = 20_000;

/// A `target/` entry that is a lane's own temporary folder: `agent-9`, `agent9`, `lane-3`.
pub(crate) fn lane_temp_dir(name: &str) -> bool {
    ["agent", "lane"].iter().any(|p| {
        name.strip_prefix(p).is_some_and(|rest| {
            let digits = rest.strip_prefix('-').unwrap_or(rest);
            !digits.is_empty() && digits.bytes().all(|b| b.is_ascii_digit())
        })
    })
}

fn same_bytes(a: &Path, b: &Path) -> bool {
    let (Ok(ma), Ok(mb)) = (fs::symlink_metadata(a), fs::symlink_metadata(b)) else {
        return false;
    };
    if !ma.is_file() || !mb.is_file() || ma.len() != mb.len() {
        return false;
    }
    matches!((fs::read(a), fs::read(b)), (Ok(x), Ok(y)) if x == y)
}

/// One ignored file (path relative to the worktree): safe to lose, or kept.
fn ignored_file_is_safe(wt: &Path, primary: &Path, rel: &Path) -> bool {
    let path = wt.join(rel);
    if fs::symlink_metadata(&path).map_or(true, |m| !m.is_file()) {
        return false; // links and special files are never counted safe
    }
    let regenerable = rel.components().any(|c| {
        REGENERABLE_DIRS
            .iter()
            .any(|d| c.as_os_str() == std::ffi::OsStr::new(d))
    });
    let ext = rel.extension().and_then(|e| e.to_str()).unwrap_or("");
    regenerable
        || REGENERABLE_EXTS.contains(&ext)
        || LANE_TEMP_EXTS.contains(&ext)
        || same_bytes(&path, &primary.join(rel))
}

/// The ignored files of `wt` outside its cargo build output, split into a safe count and the
/// kept paths (#2879 D2). `preserved` are the non-cargo entries of `target/`; a lane's
/// temporary folder there is safe. `None` when git fails or there are too many files.
pub(crate) fn ignored_split(
    wt: &Path,
    primary: &Path,
    preserved: &[String],
) -> Option<(usize, Vec<String>)> {
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
    let (mut safe, mut kept, mut seen) = (0usize, Vec::new(), 0usize);
    for rec in s.split('\0').filter(|r| r.len() > 3) {
        let (code, entry) = rec.split_at(3);
        if code != "!! " || entry == "target/" {
            continue;
        }
        let mut files = Vec::new();
        let top = wt.join(entry.trim_end_matches('/'));
        if fs::symlink_metadata(&top).is_ok_and(|m| m.is_dir()) {
            let mut stack = vec![top];
            while let Some(d) = stack.pop() {
                for e in fs::read_dir(&d).ok()?.flatten() {
                    let p = e.path();
                    if fs::symlink_metadata(&p).is_ok_and(|m| m.is_dir()) {
                        stack.push(p);
                    } else {
                        files.push(p);
                    }
                    seen += 1;
                    if seen > IGNORED_FILE_LIMIT {
                        return None;
                    }
                }
            }
        } else {
            files.push(top);
        }
        for f in files {
            let rel = f.strip_prefix(wt).ok()?.to_path_buf();
            if ignored_file_is_safe(wt, primary, &rel) {
                safe += 1;
            } else {
                kept.push(rel.to_string_lossy().into_owned());
            }
        }
    }
    for name in preserved {
        if lane_temp_dir(name)
            && fs::symlink_metadata(wt.join("target").join(name)).is_ok_and(|m| m.is_dir())
        {
            safe += 1;
        } else {
            kept.push(format!("target/{name}"));
        }
    }
    kept.sort();
    Some((safe, kept))
}

/// How far one plan goes (#2879 D2): at most `max_checked` worktrees are examined, starting
/// after `cursor` (a candidate `path` of an earlier plan) and wrapping around. The default
/// examines all of them, as before.
#[derive(Clone, Debug, Default)]
pub struct Pace {
    pub max_checked: Option<usize>,
    pub cursor: Option<String>,
}

/// The home worktree folder and the Go generation's folder (#2811), each as written and as
/// resolved; empty when no home is known.
fn work_roots() -> (Vec<PathBuf>, Vec<PathBuf>) {
    let Ok(home) = crate::home_workdir::home() else {
        return (Vec::new(), Vec::new());
    };
    let both = |p: PathBuf| {
        let mut v = vec![p.clone()];
        if let Ok(real) = fs::canonicalize(&p) {
            if real != p {
                v.push(real);
            }
        }
        v
    };
    (
        both(crate::home_workdir::worktrees_root(&home)),
        both(crate::home_workdir::go_worktrees_root(&home)),
    )
}

fn shown(path: &Path, root: &Path) -> String {
    match path.strip_prefix(root) {
        Ok(rel) if rel.as_os_str().is_empty() => ".".into(),
        Ok(rel) => rel.to_string_lossy().into_owned(),
        Err(_) => path.to_string_lossy().into_owned(),
    }
}

/// Builds the plan. `now_unix` overrides the clock (tests); `None` uses the system time.
/// Only an ancestor of the integration tip counts as merged here: `worktree.gc.apply`
/// recomputes this plan, and a squash merge alone never makes it remove (#2724).
pub fn plan(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
) -> Result<Value> {
    plan_with(project, integration_ref, recent, budget, now_unix, false)
}

/// [`plan`] with the squash-merge test (#2724) on or off. With `squash` on, a worktree whose
/// HEAD is not an ancestor of the tip but passes [`squash_merged`] counts as merged; every
/// candidate names `merged_by` (`ancestor`, `squash` or null) either way. Only the read-only
/// `optimize.plan` turns it on.
pub fn plan_with(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
    squash: bool,
) -> Result<Value> {
    plan_paced(
        project,
        integration_ref,
        recent,
        budget,
        now_unix,
        squash,
        &Pace::default(),
    )
}

/// [`plan_with`] that examines at most `pace.max_checked` worktrees, starting after
/// `pace.cursor`; the answer says how many were left (`remaining`) and where the next run
/// goes on (`next_cursor`) — a plan cut short by its time budget says the same (#2879 D2).
pub fn plan_paced(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
    squash: bool,
    pace: &Pace,
) -> Result<Value> {
    if recent.as_secs() < MIN_RECENT_MINUTES * 60 {
        return Err(format!(
            "--recent-minutes must be at least {MIN_RECENT_MINUTES} (#2498); got {}",
            recent.as_secs() / 60
        ));
    }
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
    // #2811: the home work folders, as written and resolved, so a linked `~/.moai` still
    // names its worktrees. The Go generation's `worktrees/projects` is never touched (D8a).
    let (home_roots, go_roots) = work_roots();

    let mut candidates = Vec::new();
    // #2879 D2: the primary first, then the linked worktrees starting after the cursor.
    let shown_of = |e: &crate::worktree_report::WorktreeEntry| {
        shown(
            &fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()),
            &primary,
        )
    };
    let mut order: Vec<usize> = (1..entries.len()).collect();
    if let Some(c) = pace.cursor.as_deref() {
        if let Some(at) = order.iter().position(|&i| shown_of(&entries[i]) == c) {
            order.rotate_left(at + 1);
        }
    }
    if !entries.is_empty() {
        order.insert(0, 0);
    }
    let (mut checked, mut remaining, mut last_checked) = (0usize, 0usize, None::<String>);
    for i in order {
        let e = &entries[i];
        if e.bare {
            continue;
        }
        let path = fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone());
        let target = path.join("target");
        let (cargo, preserved) = split_target(&target);
        let location = if i == 0 {
            "primary"
        } else {
            crate::home_workdir::location_in(&path, &home_roots, &go_roots, &primary)
        };
        let mut f = Facts {
            primary: i == 0,
            // Without a known home the Go folder's shape alone still keeps it (fail-safe).
            preserved_location: location == "go"
                || path
                    .to_string_lossy()
                    .contains("/.moai/worktrees/projects/")
                || (path.components().any(|c| c.as_os_str() == ".claude")
                    && path.to_string_lossy().contains("/.claude/worktrees/")),
            current_directory: cwd.as_ref().is_some_and(|c| c.starts_with(&path)) && i != 0,
            locked: e.locked,
            prunable: e.prunable,
            has_build_output: !cargo.is_empty(),
            ..Default::default()
        };
        let mut merged_by: Option<&str> = None;
        let mut ignored_safe: Option<usize> = None;
        let mut ignored_kept: Vec<String> = Vec::new();
        let over_count = i != 0 && pace.max_checked.is_some_and(|m| checked >= m);
        if Instant::now() >= deadline || over_count {
            f.budget_spent = true;
            if i != 0 {
                remaining += 1;
            }
        } else if !e.prunable {
            if i != 0 {
                checked += 1;
                last_checked = Some(shown(&path, &primary));
            }
            f.merged = e.head.as_deref().and_then(|h| merged(&primary, h, &tip));
            if f.merged == Some(true) {
                merged_by = Some("ancestor");
            } else if squash && f.merged == Some(false) {
                let sq = e
                    .head
                    .as_deref()
                    .and_then(|h| squash_merged(&primary, h, &tip));
                if sq == Some(true) {
                    merged_by = Some("squash");
                }
                // a squash test that could not run leaves the worktree unknown, not unmerged
                f.merged = sq;
            }
            f.tracked_dirty = git_out(&path, &["status", "--porcelain", "--untracked-files=no"])
                .map(|s| !s.trim().is_empty());
            if i == 0 {
                // The primary is never removed; its ignored folders (other worktrees, build
                // output) are not walked file by file.
                if let Some((u, ign, _)) = census(&path) {
                    f.untracked = Some(u);
                    f.ignored_outside_build = Some(ign + preserved.len());
                }
            } else if let (Some((u, _, _)), Some((safe, kept))) =
                (census(&path), ignored_split(&path, &primary, &preserved))
            {
                f.untracked = Some(u);
                f.ignored_outside_build = Some(kept.len());
                ignored_safe = Some(safe);
                ignored_kept = kept;
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
            "location": location,
            "branch": e.branch,
            "head": e.head,
            "action": action,
            "reasons": reasons,
            "merged_by": merged_by,
            "trim_paths": cargo.iter().map(|c| shown(c, &path)).collect::<Vec<_>>(),
            "preserved_in_target": preserved,
            "untracked": f.untracked,
            "ignored_outside_build": f.ignored_outside_build,
            "ignored_safe": ignored_safe,
            "ignored_kept": ignored_kept.iter().take(5).collect::<Vec<_>>(),
            "change_age_secs": f.change_age_secs,
            "trim_bytes": trim_bytes,
            "bytes_partial": partial,
        }));
    }
    candidates.sort_by(|a, b| a["path"].as_str().cmp(&b["path"].as_str()));
    // The digest binds what a later apply acts on: path, head, action, reasons and the exact
    // trim paths (#2498 N1 — a cargo folder that appears after the plan changes the digest).
    let bound: Vec<Value> = candidates
        .iter()
        .map(|c| json!({"path": c["path"], "head": c["head"], "action": c["action"], "reasons": c["reasons"], "trim_paths": c["trim_paths"]}))
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
        "remaining": remaining,
        "next_cursor": if remaining > 0 { json!(last_checked) } else { Value::Null },
        "plan_digest": plan_digest,
        "apply": "worktree-gc-apply with this plan_digest and a permission reference (#2498); on this machine the leader runs it only after the user confirms (#2389 comment 6774899518 item 2)",
        "network": "disabled",
        "writes": false,
    }))
}

fn validate_permission(permission: &str) -> Result<()> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    Ok(())
}

/// Why `wt` must not be touched right now, re-checked just before each deletion: a process
/// working inside it, a tracked change, or (for a removal) any untracked file or ignored file
/// outside the cargo build output. An unknown fact refuses too.
/// `primary` is the primary checkout, for the copy check of the ignored files when removing.
pub(crate) fn busy_now(wt: &Path, removing: bool, primary: Option<&Path>) -> Option<&'static str> {
    let cwds = match process_cwds() {
        Some(v) => v,
        None => return Some("process_check_unavailable"),
    };
    if cwds
        .into_iter()
        .map(|p| fs::canonicalize(&p).unwrap_or(p))
        .any(|c| c.starts_with(wt))
    {
        return Some("process_in_use");
    }
    match open_file_paths() {
        Some(paths) if paths.iter().any(|p| p.starts_with(wt)) => return Some("file_open_in_use"),
        Some(_) => {}
        None => return Some("open_file_check_unavailable"),
    }
    match git_out(wt, &["status", "--porcelain", "--untracked-files=no"]) {
        Some(s) if s.trim().is_empty() => {}
        Some(_) => return Some("tracked_changes"),
        None => return Some("status_unknown"),
    }
    if removing {
        // The same split the plan used (#2879 D2): only safe ignored files may go with it.
        let (_, preserved) = split_target(&wt.join("target"));
        let split = primary.and_then(|p| ignored_split(wt, p, &preserved));
        match (census(wt), split) {
            (Some((0, _, _)), Some((_, kept))) if kept.is_empty() => {}
            (Some(_), Some(_)) => return Some("files_outside_build_output"),
            _ => return Some("file_census_unknown"),
        }
    }
    None
}

/// Deletes the cargo-owned trim paths of `wt` that the plan listed. Each path must still sit
/// directly in `wt/target`, not be a symlink, and still be cargo-owned.
fn trim(wt: &Path, trim_paths: &[String]) -> std::result::Result<(), String> {
    let target = wt.join("target");
    // Audit B1 on PR #2557: never delete through a linked `target` — it may point outside
    // the worktree (a shared build folder). `wt` itself is the canonical worktree path.
    let target_meta = fs::symlink_metadata(&target).map_err(|_| "target/ unavailable")?;
    if !target_meta.is_dir() {
        return Err("target/ is a symlink or not a folder; refused".into());
    }
    let (owned, _) = split_target(&target);
    for rel in trim_paths {
        let path = wt.join(rel);
        if path.parent() != Some(target.as_path()) || !owned.contains(&path) {
            return Err(format!("{rel} is no longer a cargo-owned entry of target/"));
        }
        let meta = fs::symlink_metadata(&path).map_err(|_| format!("{rel} unavailable"))?;
        if meta.file_type().is_symlink() {
            return Err(format!("{rel} is a symlink; refused"));
        }
        let removed = if meta.is_dir() {
            fs::remove_dir_all(&path)
        } else {
            fs::remove_file(&path)
        };
        removed.map_err(|_| format!("{rel} could not be removed"))?;
    }
    Ok(())
}

/// `worktree.gc.apply` (#2498, WT-9): recomputes the plan with the same inputs and refuses
/// before any write unless its digest equals `plan_digest`. Then, in the plan's order, it
/// acts only on the `trim` and `remove` entries — `trim` deletes the listed cargo-owned
/// paths, `remove` runs `git worktree remove` without `--force` (git refuses a dirty tree;
/// the branch is kept) — after re-checking each worktree right before touching it
/// ([`busy_now`]). A refused or failed entry is reported and the rest go on; nothing is
/// retried. Running this on the user's own machine is a user-confirmed step the leader takes.
#[allow(clippy::too_many_arguments)]
pub fn apply(
    project: &Path,
    integration_ref: &str,
    recent: Duration,
    budget: Duration,
    now_unix: Option<u64>,
    pace: &Pace,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value> {
    validate_permission(permission_reference)?;
    let fresh = plan_paced(
        project,
        integration_ref,
        recent,
        budget,
        now_unix,
        false,
        pace,
    )?;
    if fresh["plan_digest"].as_str() != Some(plan_digest) {
        return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
    }
    let top = git_out(
        project,
        &["rev-parse", "--path-format=absolute", "--show-toplevel"],
    )
    .map(|s| PathBuf::from(s.trim()))
    .ok_or("not inside a git work tree")?;
    let listing =
        git_out(&top, &["worktree", "list", "--porcelain"]).ok_or("git worktree list failed")?;
    let entries = parse_worktree_list(&listing);
    let worktrees: Vec<PathBuf> = entries
        .iter()
        .filter(|e| !e.bare)
        .map(|e| fs::canonicalize(&e.path).unwrap_or_else(|_| e.path.clone()))
        .collect();
    let primary = worktrees.first().cloned().ok_or("no primary worktree")?;

    let mut results = Vec::new();
    let mut not_done = 0usize;
    for c in fresh["candidates"].as_array().into_iter().flatten() {
        let action = c["action"].as_str().unwrap_or("skip");
        if action != "trim" && action != "remove" {
            continue;
        }
        let shown_path = c["path"].as_str().unwrap_or_default().to_string();
        let wt = fs::canonicalize(primary.join(&shown_path))
            .unwrap_or_else(|_| primary.join(&shown_path));
        let outcome = if wt == primary || !worktrees.contains(&wt) {
            Err("not a linked worktree of this repository".to_string())
        } else if let Some(why) = busy_now(&wt, action == "remove", Some(&primary)) {
            Err(format!("refused: {why}"))
        } else if action == "trim" {
            let paths: Vec<String> = c["trim_paths"]
                .as_array()
                .into_iter()
                .flatten()
                .filter_map(|p| p.as_str().map(str::to_string))
                .collect();
            trim(&wt, &paths).map(|()| "trimmed")
        } else {
            let mut cmd = Command::new("git");
            scrub_git_env(&mut cmd);
            match cmd
                .arg("-C")
                .arg(&primary)
                .args(["worktree", "remove"])
                .arg(&wt)
                .output()
            {
                Ok(o) if o.status.success() => Ok("removed"),
                Ok(o) => Err(format!(
                    "git worktree remove refused: {}",
                    String::from_utf8_lossy(&o.stderr).trim()
                )),
                Err(_) => Err("git worktree remove could not run".to_string()),
            }
        };
        let (status, detail) = match outcome {
            Ok(s) => (s, Value::Null),
            Err(e) => {
                not_done += 1;
                (
                    if e.starts_with("refused") {
                        "refused"
                    } else {
                        "failed"
                    },
                    json!(e),
                )
            }
        };
        results.push(
            json!({"path": shown_path, "action": action, "status": status, "detail": detail}),
        );
    }
    Ok(json!({
        "schema_version": 1,
        "status": if results.is_empty() { "unchanged" } else if not_done == 0 { "applied" } else { "partial" },
        "plan_digest": plan_digest,
        "results": results,
        "retry": "none: a refused or failed entry is reported, never retried",
        "network": "disabled",
        "writes": !results.is_empty(),
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

    /// #2498: `trim` re-checks each listed path right before deleting it. A path that is no
    /// longer cargo-owned (its `.cargo-lock` marker gone, e.g. a lane folder put in its place
    /// after the plan) is refused and kept, and so is one outside `target/`.
    #[test]
    fn trim_refuses_a_path_that_is_no_longer_cargo_owned() {
        let dir = tempfile::TempDir::new().unwrap();
        let wt = dir.path();
        fs::create_dir_all(wt.join("target/debug")).unwrap();
        fs::write(wt.join("target/debug/note.md"), "lane note\n").unwrap();
        let err = trim(wt, &["target/debug".to_string()]).unwrap_err();
        assert!(err.contains("no longer a cargo-owned"), "{err}");
        assert!(wt.join("target/debug/note.md").exists());
        fs::write(wt.join("keep.txt"), "x\n").unwrap();
        assert!(trim(wt, &["keep.txt".to_string()]).is_err());
        assert!(wt.join("keep.txt").exists());
        // Control: with the marker back it is cargo-owned and trimmed.
        fs::write(wt.join("target/debug/.cargo-lock"), "").unwrap();
        trim(wt, &["target/debug".to_string()]).unwrap();
        assert!(!wt.join("target/debug").exists());
    }

    /// Audit B1 on PR #2557: `trim` itself refuses when `target` is a symlink, even if a
    /// caller passes paths that look planned — the delete never follows the link out.
    #[cfg(unix)]
    #[test]
    fn trim_refuses_a_symlinked_target_folder() {
        let dir = tempfile::TempDir::new().unwrap();
        let wt = dir.path().join("wt");
        let outside = dir.path().join("outside");
        fs::create_dir_all(outside.join("debug")).unwrap();
        fs::write(outside.join("debug/.cargo-lock"), "").unwrap();
        fs::write(outside.join("debug/other-lane.rlib"), "x").unwrap();
        fs::create_dir_all(&wt).unwrap();
        std::os::unix::fs::symlink(&outside, wt.join("target")).unwrap();
        let err = trim(&wt, &["target/debug".to_string()]).unwrap_err();
        assert!(err.contains("symlink"), "{err}");
        assert!(outside.join("debug/other-lane.rlib").exists());
        let (owned, other) = split_target(&wt.join("target"));
        assert!(owned.is_empty(), "{owned:?}");
        assert!(
            !other.is_empty(),
            "a linked target counts as files outside the build"
        );
    }

    /// #2498 N3: clean and merged but never built and holding nothing — kept, not removed.
    #[test]
    fn an_empty_worktree_is_skipped_not_removed() {
        let mut f = eligible();
        f.has_build_output = false;
        assert_eq!(decide(&f, 7_200), ("skip", vec!["empty_worktree"]));
        f.has_build_output = true;
        assert_eq!(decide(&f, 7_200), ("remove", vec![]));
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

    /// #2879 D2: only a lane's own folder names are temporary; anything else in `target/` is
    /// a note kept with the worktree.
    #[test]
    fn lane_temp_dirs_are_named_for_a_lane() {
        for ok in ["agent-9", "agent29", "lane-3", "lane12"] {
            assert!(lane_temp_dir(ok), "{ok}");
        }
        for no in [
            "agent", "agent-", "agents-9", "notes", "agent-9x", "lane-a", "pr-12",
        ] {
            assert!(!lane_temp_dir(no), "{no}");
        }
    }
}
