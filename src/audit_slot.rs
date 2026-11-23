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
//! run beside an `ensure` or each other on one repository. Before the first write and again
//! under the lock they check that no part of the slot path is a link and that the slot is its
//! own worktree of the repository ([`worktree_slot::own_worktree`]); otherwise git run in it
//! would act on another checkout.
//!
//! A slot has one holder at a time. `prep` records it (name, session, optional pid, head and a
//! lease) in the git common dir under the lock; another holder is refused while the lease runs
//! and its pid, when it gave one, is alive; a dead holder is taken over with a line in the
//! answer; `done` releases the slot. `prep auto` takes the first slot without a live holder.
//!
//! After a build is placed, every tracked file of the slot gets the current modification time,
//! so Cargo compiles the workspace crate once from the audited sources and reuses only the
//! dependencies: a lane's `target/` may hold a mutant build that Cargo would call `Fresh`.

use crate::spec_take::git;
use crate::worktree_slot::{self, Caps, SLOT_DIR, TEMP_PREFIXES};
use serde_json::{json, Value};
use std::fs;
use std::path::{Component, Path, PathBuf};
use std::process::{Command, Stdio};

type Result<T> = std::result::Result<T, String>;

/// The slots an audit may take.
pub const AUDIT_SLOTS: [&str; 3] = ["audit-1", "audit-2", "audit-3"];
/// The `prep` slot name that takes the first slot without a live holder.
pub const AUTO: &str = "auto";
/// The folder in the git common dir that holds one holder record per slot.
const HOLDERS_DIR: &str = "moai-audit-slots";
/// How long a holder keeps a slot when `--lease-secs` is not given, and the longest lease.
pub const DEFAULT_LEASE_SECS: u64 = 3 * 3600;
pub const MAX_LEASE_SECS: u64 = 24 * 3600;

/// Who takes or finishes a slot.
pub struct Who<'a> {
    /// A lane or auditor name (`lane_registry` rules).
    pub holder: &'a str,
    /// A process whose end ends the hold before the lease does.
    pub pid: Option<i64>,
    pub lease_secs: u64,
}

fn check_name(name: &str) -> Result<()> {
    if AUDIT_SLOTS.contains(&name) {
        Ok(())
    } else {
        Err(format!(
            "audit-slot refused {name:?}: invalid slot name: one of audit-1, audit-2 or audit-3"
        ))
    }
}

/// Refuses a slot path when it or a folder above it inside the repository (`.claude`,
/// `.claude/worktrees`) is a symbolic link: every write of `prep` and `done` would land in
/// whatever it points at (another worktree, a shared folder). Checked before the first write and
/// again under the slot lock.
fn no_links(primary: &Path, slot: &Path, rel: &str) -> Result<()> {
    let mut at = primary.to_path_buf();
    let below = slot.strip_prefix(primary).unwrap_or(slot);
    for part in below.components() {
        at.push(part);
        if fs::symlink_metadata(&at).is_ok_and(|m| m.file_type().is_symlink()) {
            let shown = at
                .strip_prefix(primary)
                .unwrap_or(&at)
                .display()
                .to_string();
            return Err(if shown == rel {
                format!("audit-slot refused {rel}: it is a link, not a slot folder; nothing was changed")
            } else {
                format!("audit-slot refused {rel}: {shown} above it is a link; nothing was changed")
            });
        }
    }
    Ok(())
}

/// Refuses a slot that is not its own worktree of this repository: git in a plain folder walks
/// up to the primary checkout, and a borrowed `.git` acts on another worktree's index.
fn own(slot: &Path, common: &Path, rel: &str, changed: &str) -> Result<()> {
    worktree_slot::own_worktree(slot, common).map_err(|e| {
        format!("audit-slot refused {rel}: it is not its own worktree of this repository ({e}); {changed}")
    })
}

/// `--head` must be a full commit id: a short id, a branch or a tag can name another commit by
/// the time the audit reads it.
fn full_hex(head: &str) -> Result<()> {
    if head.len() == 40 && head.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f')) {
        Ok(())
    } else {
        Err(format!(
            "audit-slot refused --head {head:?}: give the full 40-character commit id; nothing was changed"
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
        // A copy that a killed prep left is one folder of many files: name the folder once.
        let first = path.split('/').next().unwrap_or(path);
        let path = if !tracked && TEMP_PREFIXES.iter().any(|p| first.starts_with(p)) {
            if out.iter().any(|c: &Change| c.path == first) {
                continue;
            }
            first
        } else {
            path
        };
        out.push(Change {
            path: path.to_string(),
            tracked,
        });
    }
    // A copy that git ignores (an `info/exclude` or `.gitignore` line such as `.target-*`) is not
    // in `git status` at all: read the slot folder, so it is named, refused and discarded too.
    for name in worktree_slot::leftovers(slot)? {
        if !out.iter().any(|c| c.path == name) {
            out.push(Change {
                path: name,
                tracked: false,
            });
        }
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

fn now() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |d| d.as_secs() as i64)
}

/// The calling session, when the host names one.
fn session() -> Option<String> {
    crate::session_current::SOURCES
        .iter()
        .find_map(|k| std::env::var(k).ok())
        .filter(|id| crate::session_current::well_formed(id))
}

fn check_who(who: &Who) -> Result<()> {
    if !crate::lane_registry::valid_name(who.holder) {
        return Err(format!(
            "audit-slot refused --holder {:?}: 1 to 64 letters, digits, `.`, `_` or `-`; nothing was changed",
            who.holder
        ));
    }
    if !(1..=MAX_LEASE_SECS).contains(&who.lease_secs) {
        return Err(format!(
            "audit-slot refused --lease-secs {}: 1 to {MAX_LEASE_SECS}; nothing was changed",
            who.lease_secs
        ));
    }
    if who
        .pid
        .is_some_and(|p| !(1..=i64::from(i32::MAX)).contains(&p))
    {
        return Err("audit-slot refused --pid: a process id from 1 up; nothing was changed".into());
    }
    Ok(())
}

fn holder_file(common: &Path, name: &str) -> PathBuf {
    common.join(HOLDERS_DIR).join(format!("{name}.json"))
}

/// The slot's holder record; a record that does not read as an object counts as a dead holder.
fn read_holder(common: &Path, name: &str) -> Option<Value> {
    let bytes = fs::read(holder_file(common, name)).ok()?;
    Some(
        serde_json::from_slice::<Value>(&bytes)
            .ok()
            .filter(Value::is_object)
            .unwrap_or_else(|| json!({})),
    )
}

/// A holder is alive while its lease runs and its pid, when it gave one, is not known dead.
fn alive(rec: &Value, at: i64) -> bool {
    let leased = rec["lease_until"].as_i64().is_some_and(|t| t > at);
    let pid_ok = match rec["pid"].as_i64() {
        Some(pid) => crate::home_state::pid_alive(pid) != Some(false),
        None => true,
    };
    leased && pid_ok
}

/// The same holder: the same name, and the same session when both name one.
fn same_holder(rec: &Value, who: &Who, session: Option<&str>) -> bool {
    rec["holder"].as_str() == Some(who.holder)
        && match (rec["session"].as_str(), session) {
            (Some(a), Some(b)) => a == b,
            _ => true,
        }
}

/// How the slot stands for `who`.
enum Hold {
    /// No record.
    Free,
    /// `who` holds it already.
    Mine,
    /// A holder whose lease ended or whose process ended.
    Dead(Value),
    /// Another live holder.
    Held(Value),
}

fn hold_of(common: &Path, name: &str, who: &Who, session: Option<&str>) -> Hold {
    match read_holder(common, name) {
        None => Hold::Free,
        Some(rec) if same_holder(&rec, who, session) => Hold::Mine,
        Some(rec) if alive(&rec, now()) => Hold::Held(rec),
        Some(rec) => Hold::Dead(rec),
    }
}

fn held_by(rel: &str, rec: &Value) -> String {
    format!(
        "audit-slot refused {rel}: it is held by {} (session {}, pid {}, head {}) until {}; nothing was changed",
        rec["holder"], rec["session"], rec["pid"], rec["head"], rec["lease_until"]
    )
}

/// Writes the holder record through a temporary file and a rename.
fn write_holder(common: &Path, name: &str, rec: &Value) -> Result<()> {
    let dir = common.join(HOLDERS_DIR);
    fs::create_dir_all(&dir).map_err(|e| format!("the holder folder could not be made: {e}"))?;
    let temp = dir.join(format!(".{name}.{}.tmp", std::process::id()));
    fs::write(&temp, format!("{rec}\n"))
        .and_then(|()| fs::rename(&temp, holder_file(common, name)))
        .map_err(|e| {
            let _ = fs::remove_file(&temp);
            format!("the holder record could not be written: {e}")
        })
}

/// Puts the record that was there before back, or removes the new one.
fn restore_holder(common: &Path, name: &str, before: Option<&Value>) {
    match before {
        Some(rec) => {
            let _ = write_holder(common, name, rec);
        }
        None => {
            let _ = fs::remove_file(holder_file(common, name));
        }
    }
}

/// Gives every tracked file of the slot the current modification time, so Cargo compiles the
/// workspace crate from these sources instead of trusting a build made from others (a lane's
/// mutant, an earlier audit's). Dependencies outside the slot keep their times and stay `Fresh`.
fn touch_sources(slot: &Path) -> Result<usize> {
    let list = git_step(slot, "ls-files", &["ls-files", "-z"])?;
    let at = std::time::SystemTime::now();
    let mut count = 0;
    for rel in list.split('\0').filter(|r| !r.is_empty()) {
        let p = inside(slot, rel)?;
        match fs::symlink_metadata(&p) {
            Ok(m) if m.file_type().is_file() => {
                fs::File::open(&p)
                    .and_then(|f| f.set_modified(at))
                    .map_err(|e| format!("{rel} could not be touched: {e}"))?;
                count += 1;
            }
            _ => {}
        }
    }
    Ok(count)
}

/// What an `ensure` did, for a refusal that comes after it.
fn changed_by(ensured: &Value) -> String {
    if ensured["added"] == json!(true) {
        "this run added the slot (worktree-slot ensure) and changed nothing else".into()
    } else if ensured["warmed"] != json!("none") {
        "this run warmed the slot's target/ from the seed and changed nothing else".into()
    } else {
        "nothing was changed".into()
    }
}

/// The audit slot whose holder record names `holder`, whether its lease still runs or not (the
/// same name is the same holder, as [`prep_with`] reads it). The card-addressed commands find a
/// card's slot with this.
pub fn slot_held_by(project: &Path, holder: &str) -> Result<Option<String>> {
    let common = worktree_slot::common_dir(project)?;
    Ok(AUDIT_SLOTS
        .iter()
        .find(|name| {
            read_holder(&common, name).is_some_and(|rec| rec["holder"].as_str() == Some(holder))
        })
        .map(|name| (*name).to_string()))
}

/// Takes the audit slot `name` (or, with [`AUTO`], the first one without a live holder) for
/// `head` as `who`: adds it when missing, refuses it when another live holder has it or it holds
/// changes, records the holder, switches it detached to the head, places a build and touches the
/// sources (see the module notes).
pub fn prep(
    project: &Path,
    name: &str,
    head: &str,
    lane: Option<&Path>,
    caps: Caps,
    who: &Who,
) -> Result<Value> {
    prep_with(project, name, head, lane, caps, who, session().as_deref())
}

/// [`prep`] with the session the holder record names given by the caller. `None` records none,
/// and a holder with no session is the same holder in every session: the card-addressed
/// `audit-prep` (SPEC-20261124-audit-prep D-6) uses that.
pub fn prep_with(
    project: &Path,
    name: &str,
    head: &str,
    lane: Option<&Path>,
    caps: Caps,
    who: &Who,
    session: Option<&str>,
) -> Result<Value> {
    if name != AUTO {
        check_name(name)?;
    }
    check_who(who)?;
    full_hex(head)?;
    let common = worktree_slot::common_dir(project)?;
    let primary = worktree_slot::primary_of(&common)?;
    let sha = commit_of(&primary, head)?;
    let from = match lane {
        Some(lane) => Some(lane_target(lane, &primary, &sha)?),
        None => None,
    };
    let names: Vec<&str> = if name == AUTO {
        AUDIT_SLOTS.to_vec()
    } else {
        vec![name]
    };
    let mut busy = Vec::new();
    for slot_name in names {
        let step = prep_slot(
            project,
            &common,
            &primary,
            slot_name,
            &sha,
            lane.zip(from.as_deref()),
            caps,
            who,
            session,
        )?;
        match step {
            Ok(answer) => return Ok(answer),
            Err(rec) if name == AUTO => busy.push(json!({"slot": slot_name, "holder": rec})),
            Err(rec) => return Err(held_by(&format!("{SLOT_DIR}/{slot_name}"), &rec)),
        }
    }
    Err(format!(
        "audit-slot refused auto: all audit slots are held ({}); nothing was changed",
        Value::Array(busy)
    ))
}

/// One slot of [`prep`]: `Ok(Err(record))` when another live holder has it.
#[allow(clippy::too_many_arguments)]
fn prep_slot(
    project: &Path,
    common: &Path,
    primary: &Path,
    name: &str,
    sha: &str,
    lane: Option<(&Path, &Path)>,
    caps: Caps,
    who: &Who,
    session: Option<&str>,
) -> Result<std::result::Result<Value, Value>> {
    let slot = primary.join(SLOT_DIR).join(name);
    let rel = format!("{SLOT_DIR}/{name}");
    {
        // Before `ensure` writes anything: a held slot, a link or a slot that is not its own
        // worktree is refused with nothing changed.
        let _lock = worktree_slot::lock(common)?;
        no_links(primary, &slot, &rel)?;
        if let Hold::Held(rec) = hold_of(common, name, who, session) {
            return Ok(Err(rec));
        }
        if fs::symlink_metadata(&slot).is_ok() {
            own(&slot, common, &rel, "nothing was changed")?;
        }
    }
    let ensured = worktree_slot::ensure(project, name, sha, caps).map_err(|e| {
        if e.contains("cap reached") || e.contains("under the floor") {
            format!("audit-slot refused {rel}: {e}")
        } else {
            e
        }
    })?;
    let changed = changed_by(&ensured);
    let added = ensured["added"] == json!(true);
    let seeded = ensured["warmed"] != json!("none");
    let _lock = worktree_slot::lock(common)?;
    no_links(primary, &slot, &rel).map_err(|e| e.replace("nothing was changed", &changed))?;
    own(&slot, common, &rel, &changed)?;
    let before = read_holder(common, name);
    let took_over = match hold_of(common, name, who, session) {
        Hold::Held(rec) => return Ok(Err(rec)),
        Hold::Dead(rec) => Some(rec),
        Hold::Free | Hold::Mine => None,
    };
    let dirty = changes(&slot)?;
    if !dirty.is_empty() {
        return Err(format!(
            "audit-slot refused {rel}: it holds changes of an earlier audit ({}); run `done {name} --discard` first; {changed}",
            names(&dirty)
        ));
    }
    let at = now();
    let record = json!({
        "holder": who.holder, "session": session, "pid": who.pid, "head": sha,
        "at": at, "lease_until": at + who.lease_secs as i64,
    });
    write_holder(common, name, &record)?;
    let placed = (|| -> Result<Value> {
        git_step(&slot, "switch", &["switch", "--quiet", "--detach", sha])?;
        let mut build = match lane {
            Some((lane, from)) => {
                let method = inherit(from, lane, sha, &slot).map_err(|e| {
                    format!("{rel} is at {sha} but inheriting the lane build failed: {e}")
                })?;
                json!({"source": "lane", "method": method, "from": lane.display().to_string()})
            }
            None if seeded => json!({"source": "seed", "method": ensured["warmed"]}),
            None if slot.join("target").is_dir() => json!({"source": "kept", "method": null}),
            None => json!({"source": "none", "method": null}),
        };
        build["sources_touched"] = if slot.join("target").is_dir() {
            json!(touch_sources(&slot)?)
        } else {
            json!(0)
        };
        Ok(build)
    })();
    let build = match placed {
        Ok(b) => b,
        Err(e) => {
            restore_holder(common, name, before.as_ref());
            return Err(e);
        }
    };
    let head = git_step(&slot, "head", &["rev-parse", "HEAD"])?;
    let clean = changes(&slot)?.is_empty();
    Ok(Ok(json!({
        "slot": name, "path": rel, "head": head, "added": added, "build": build, "clean": clean,
        "holder": record, "took_over": took_over,
    })))
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

/// Finishes an audit in the slot `name` as `who`: a clean slot is done; one with changes is
/// refused unless `discard`, which reverts tracked changes, removes untracked files outside
/// `target/` and keeps `target/`. Another live holder's slot is refused; a dead holder's is taken
/// over. A slot left clean is released.
pub fn done(project: &Path, name: &str, discard: bool, who: &Who) -> Result<Value> {
    done_with(project, name, discard, who, session().as_deref())
}

/// [`done`] with the calling session given by the caller (see [`prep_with`]).
pub fn done_with(
    project: &Path,
    name: &str,
    discard: bool,
    who: &Who,
    session: Option<&str>,
) -> Result<Value> {
    check_name(name)?;
    check_who(who)?;
    let common = worktree_slot::common_dir(project)?;
    let primary = worktree_slot::primary_of(&common)?;
    let slot = primary.join(SLOT_DIR).join(name);
    let rel = format!("{SLOT_DIR}/{name}");
    let _lock = worktree_slot::lock(&common)?;
    no_links(&primary, &slot, &rel)?;
    let registered = worktree_slot::linked(&primary)?
        .iter()
        .any(|e| worktree_slot::same(&e.path, &slot));
    if !registered || !slot.is_dir() {
        return Err(format!(
            "audit-slot refused {rel}: not a worktree of this repository (prep adds it)"
        ));
    }
    own(&slot, &common, &rel, "nothing was changed")?;
    let took_over = match hold_of(&common, name, who, session) {
        Hold::Held(rec) => return Err(held_by(&rel, &rec)),
        Hold::Dead(rec) => Some(rec),
        Hold::Free | Hold::Mine => None,
    };
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
    if !left.is_empty() {
        return Err(format!(
            "audit-slot: {rel} still holds changes after the discard ({}); the slot stays held",
            names(&left)
        ));
    }
    let released = fs::remove_file(holder_file(&common, name)).is_ok();
    let head = git_step(&slot, "head", &["rev-parse", "HEAD"])?;
    Ok(json!({
        "slot": name, "path": rel, "head": head,
        "reverted": tracked.iter().map(|c| c.path.clone()).collect::<Vec<_>>(),
        "removed": untracked.iter().map(|c| c.path.clone()).collect::<Vec<_>>(),
        "clean": true,
        "target_kept": slot.join("target").is_dir(),
        "released": released, "took_over": took_over,
    }))
}
