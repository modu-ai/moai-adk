//! Per-item move journal (#2871 U1, survey 6876590509 R9–R15, steps ③–⑧): the library the
//! later HOME steps (U2 Go retire, U3 Rust install) use to move whole entries by rename.
//! No command calls it yet.
//!
//! Invariants:
//!
//! - Nothing is deleted and nothing is copied: each item is one rename, on one file system
//!   (the source and the destination folder must have the same device id; another file
//!   system is refused, never copied).
//! - Links are never followed: a source that is a link, or a destination folder that is a
//!   link, is refused.
//! - The destination must be free: an existing name, or one that differs from an existing
//!   entry of the destination folder only by letter case, is a `conflict` and nothing is
//!   overwritten. The rename itself is exclusive too (`renamex_np(RENAME_EXCL)` on macOS,
//!   `renameat2(RENAME_NOREPLACE)` on Linux); other platforms refuse.
//! - Only metadata is read (types, device ids, folder listings), never file contents.
//! - The journal is written and fsynced before the first rename: state `pending`, every item
//!   `planned`. Each item is marked `moving` before its rename and `moved` after it; the end
//!   writes `committed`. A crash leaves a journal [`resume`] finishes: a `planned` item whose
//!   source is still there and whose destination is free is moved; an unfinished item whose
//!   rename may or may not have happened (a `moving` one, or a `planned` one whose source is
//!   gone while something is at its destination) becomes `needs_review`.
//! - [`revert`] replays the `moved` items in reverse order, each only when the original path
//!   is free and the moved entry is still at its destination; otherwise that item is a
//!   `conflict` and stays where it is.
//! - An entry the caller's `in_use` check reports open (or cannot judge) is skipped (R2).
//!
//! The caller chooses the journal folder; this module never decides HOME paths.

use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
const PREFIX: &str = "move-journal-";

/// One planned move: absolute `from` to absolute `to`.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PlannedMove {
    pub from: PathBuf,
    pub to: PathBuf,
}

/// The caller's answer to "is this entry open by a running program?" (R2).
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum InUse {
    No,
    Yes(String),
    Unknown,
}

/// The no-op `in_use` check (tests, and callers that checked beforehand).
pub fn never_in_use(_: &Path) -> InUse {
    InUse::No
}

/// The device check used unless the caller supplies another: the device id of the source
/// (not followed) against that of the destination folder. `None` where device ids are not
/// available, which refuses the item.
pub fn same_device(from: &Path, to_parent: &Path) -> Option<bool> {
    #[cfg(unix)]
    {
        use std::os::unix::fs::MetadataExt;
        let a = fs::symlink_metadata(from).ok()?.dev();
        let b = fs::symlink_metadata(to_parent).ok()?.dev();
        Some(a == b)
    }
    #[cfg(not(unix))]
    {
        let _ = (from, to_parent);
        None
    }
}

/// What an apply, resume or revert consults besides the file system.
pub struct MoveEnv<'a> {
    pub in_use: &'a dyn Fn(&Path) -> InUse,
    pub same_device: &'a dyn Fn(&Path, &Path) -> Option<bool>,
    /// Test seam: stop with an error at this point, leaving everything as it is, the way a
    /// crash would (`after_journal`, `after_mark_moving`, `after_move:<n>`, `before_commit`).
    /// Consulted in debug builds only.
    pub fault: Option<&'a str>,
}

impl<'a> MoveEnv<'a> {
    pub fn new(in_use: &'a dyn Fn(&Path) -> InUse) -> Self {
        MoveEnv {
            in_use,
            same_device: &same_device,
            fault: None,
        }
    }

    fn fault(&self, point: &str) -> Result<()> {
        #[cfg(debug_assertions)]
        if self.fault == Some(point) {
            return Err(format!("injected crash at {point}"));
        }
        let _ = point;
        Ok(())
    }
}

#[derive(Clone, Copy, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum JournalState {
    Pending,
    Committed,
    Reverted,
    /// A revert left at least one item where it was (a conflict); a later revert retries.
    RevertPartial,
}

#[derive(Clone, Copy, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum ItemStatus {
    Planned,
    Moving,
    Moved,
    Skipped,
    Conflict,
    Refused,
    NeedsReview,
    Reverted,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct JournalItem {
    pub from: String,
    pub to: String,
    pub status: ItemStatus,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub reason: Option<String>,
    /// `file` or `folder`, recorded at the move so a revert only moves back the same kind.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub entry: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Journal {
    pub schema_version: u32,
    pub kind: String,
    pub stamp: String,
    pub state: JournalState,
    /// Old install id to new one (R5), recorded with the moves in one transaction.
    pub id_aliases: BTreeMap<String, String>,
    pub items: Vec<JournalItem>,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct ItemReport {
    pub from: String,
    pub to: String,
    /// `moved`, `skipped:<reason>`, `conflict`, `refused:<reason>`, `needs_review`,
    /// `reverted`, or `planned`/`moving` for an item a stopped run did not finish.
    pub outcome: String,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct Report {
    pub journal: String,
    pub state: JournalState,
    /// Whether this call renamed anything.
    pub changed: bool,
    pub items: Vec<ItemReport>,
}

/// The journal file of `stamp` (`YYYYMMDDTHHMMSSZ`) in `dir`.
pub fn journal_path(dir: &Path, stamp: &str) -> PathBuf {
    dir.join(format!("{PREFIX}{stamp}.json"))
}

/// Write the pending journal for `items` into `dir`, then move each item. `now` is
/// `YYYY-MM-DDTHH:MM:SSZ` and names the journal. When that journal already exists for the
/// same plan, this resumes it (a committed one is left unchanged); for another plan it fails.
pub fn apply(
    dir: &Path,
    kind: &str,
    items: &[PlannedMove],
    id_aliases: &BTreeMap<String, String>,
    now: &str,
    env: &MoveEnv,
) -> Result<Report> {
    let stamp = crate::home_state::stamp(now)?;
    let journal = Journal {
        schema_version: SCHEMA_VERSION,
        kind: kind.to_string(),
        stamp: stamp.clone(),
        state: JournalState::Pending,
        id_aliases: id_aliases.clone(),
        items: items.iter().map(planned).collect::<Result<_>>()?,
    };
    let path = journal_path(dir, &stamp);
    if fs::symlink_metadata(&path).is_ok() {
        let existing = read_journal(&path)?;
        let same_plan = existing.kind == journal.kind
            && existing.id_aliases == journal.id_aliases
            && existing.items.len() == journal.items.len()
            && existing
                .items
                .iter()
                .zip(&journal.items)
                .all(|(a, b)| a.from == b.from && a.to == b.to);
        if !same_plan {
            return Err(format!(
                "{}: a journal for another plan already has this stamp",
                path.display()
            ));
        }
        return run(&path, existing, env);
    }
    publish_new(&path, &journal)?;
    env.fault("after_journal")?;
    run(&path, journal, env)
}

/// Continue the journal at `path` after a stop (see the module comment). A committed or
/// reverted journal is left unchanged.
pub fn resume(path: &Path, env: &MoveEnv) -> Result<Report> {
    let journal = read_journal(path)?;
    run(path, journal, env)
}

/// Move the `moved` items of the journal at `path` back, last first (see the module comment).
pub fn revert(path: &Path) -> Result<Report> {
    let mut j = read_journal(path)?;
    if j.state == JournalState::Reverted {
        return Ok(report(path, &j, false));
    }
    let mut changed = false;
    let mut conflict = false;
    let mut outcomes = vec![None; j.items.len()];
    for i in (0..j.items.len()).rev() {
        if j.items[i].status != ItemStatus::Moved {
            continue;
        }
        let from = PathBuf::from(&j.items[i].from);
        let to = PathBuf::from(&j.items[i].to);
        let back = match (entry_kind(&to), &j.items[i].entry) {
            (Some(k), Some(rec)) if &k == rec => match destination_free(&from) {
                Ok(()) => rename_exclusive(&to, &from),
                Err(e) => Err(e),
            },
            _ => Err((
                ItemStatus::Conflict,
                "moved_entry_not_at_destination".into(),
            )),
        };
        match back {
            Ok(()) => {
                j.items[i].status = ItemStatus::Reverted;
                j.items[i].reason = None;
                changed = true;
                outcomes[i] = Some("reverted".to_string());
            }
            Err((_, why)) => {
                j.items[i].reason = Some(format!("revert_conflict:{why}"));
                conflict = true;
                outcomes[i] = Some("conflict".to_string());
            }
        }
        save(path, &j)?;
    }
    j.state = if conflict {
        JournalState::RevertPartial
    } else {
        JournalState::Reverted
    };
    save(path, &j)?;
    let mut r = report(path, &j, changed);
    for (item, outcome) in r.items.iter_mut().zip(outcomes) {
        if let Some(o) = outcome {
            item.outcome = o;
        }
    }
    Ok(r)
}

pub fn read_journal(path: &Path) -> Result<Journal> {
    let name = path
        .file_name()
        .and_then(|n| n.to_str())
        .and_then(|n| n.strip_prefix(PREFIX))
        .and_then(|n| n.strip_suffix(".json"));
    if !name.is_some_and(crate::state_root::utc_stamp) {
        return Err(format!("{}: not a move journal name", path.display()));
    }
    match fs::symlink_metadata(path) {
        Ok(m) if m.is_file() => {}
        _ => return Err(format!("{}: not a plain file", path.display())),
    }
    let bytes = fs::read(path).map_err(|e| format!("{}: {e}", path.display()))?;
    let j: Journal =
        serde_json::from_slice(&bytes).map_err(|e| format!("{}: {e}", path.display()))?;
    if j.schema_version != SCHEMA_VERSION {
        return Err(format!(
            "{}: schema_version {} is not {SCHEMA_VERSION}",
            path.display(),
            j.schema_version
        ));
    }
    Ok(j)
}

fn planned(m: &PlannedMove) -> Result<JournalItem> {
    let text = |p: &Path| -> Result<String> {
        if !p.is_absolute() {
            return Err(format!("{}: not an absolute path", p.display()));
        }
        p.to_str()
            .map(str::to_string)
            .ok_or_else(|| format!("{}: not UTF-8", p.display()))
    };
    let (from, to) = (text(&m.from)?, text(&m.to)?);
    if m.to.starts_with(&m.from) {
        return Err(format!("{to}: the destination is inside the source"));
    }
    Ok(JournalItem {
        from,
        to,
        status: ItemStatus::Planned,
        reason: None,
        entry: None,
    })
}

fn run(path: &Path, mut j: Journal, env: &MoveEnv) -> Result<Report> {
    if j.state != JournalState::Pending {
        return Ok(report(path, &j, false));
    }
    let mut moves = 0;
    for i in 0..j.items.len() {
        match j.items[i].status {
            ItemStatus::Planned => {}
            ItemStatus::Moving => {
                // A stop between the mark and the record: whether the rename happened is
                // not guessed.
                j.items[i].status = ItemStatus::NeedsReview;
                j.items[i].reason = Some("interrupted_during_rename".into());
                save(path, &j)?;
                continue;
            }
            _ => continue,
        }
        let from = PathBuf::from(&j.items[i].from);
        let to = PathBuf::from(&j.items[i].to);
        let entry = match check(&from, &to, env) {
            Ok(entry) => entry,
            Err((status, reason)) => {
                j.items[i].status = status;
                j.items[i].reason = Some(reason);
                save(path, &j)?;
                continue;
            }
        };
        j.items[i].status = ItemStatus::Moving;
        j.items[i].entry = Some(entry.to_string());
        save(path, &j)?;
        env.fault("after_mark_moving")?;
        match rename_exclusive(&from, &to) {
            Ok(()) => {
                j.items[i].status = ItemStatus::Moved;
                save(path, &j)?;
                moves += 1;
                env.fault(&format!("after_move:{moves}"))?;
            }
            Err((status, reason)) => {
                j.items[i].status = status;
                j.items[i].reason = Some(reason);
                save(path, &j)?;
            }
        }
    }
    env.fault("before_commit")?;
    j.state = JournalState::Committed;
    save(path, &j)?;
    Ok(report(path, &j, moves > 0))
}

type Verdict = (ItemStatus, String);

fn refused(reason: &str) -> Verdict {
    (ItemStatus::Refused, reason.to_string())
}

/// Every check before one rename; the entry kind on success.
fn check(from: &Path, to: &Path, env: &MoveEnv) -> std::result::Result<&'static str, Verdict> {
    let src = match fs::symlink_metadata(from) {
        Ok(m) => m,
        Err(_) if fs::symlink_metadata(to).is_ok() => {
            // The source is gone and something is at the destination: maybe an earlier run
            // moved it, maybe not; not guessed.
            return Err((
                ItemStatus::NeedsReview,
                "source_missing_destination_present".into(),
            ));
        }
        Err(_) => return Err(refused("source_missing")),
    };
    if src.file_type().is_symlink() {
        return Err(refused("source_is_link"));
    }
    let entry = if src.is_dir() {
        "folder"
    } else if src.is_file() {
        "file"
    } else {
        return Err(refused("source_not_file_or_folder"));
    };
    let parent = to
        .parent()
        .ok_or_else(|| refused("destination_has_no_folder"))?;
    match fs::symlink_metadata(parent) {
        Ok(m) if m.file_type().is_symlink() => return Err(refused("destination_folder_is_link")),
        Ok(m) if m.is_dir() => {}
        Ok(_) => return Err(refused("destination_folder_not_folder")),
        Err(_) => return Err(refused("destination_folder_missing")),
    }
    match (env.same_device)(from, parent) {
        Some(true) => {}
        Some(false) => return Err(refused("cross_device")),
        None => return Err(refused("device_unknown")),
    }
    destination_free(to)?;
    match (env.in_use)(from) {
        InUse::No => Ok(entry),
        InUse::Yes(reason) => Err((ItemStatus::Skipped, reason)),
        InUse::Unknown => Err((ItemStatus::Skipped, "in_use_unknown".into())),
    }
}

/// `file` or `folder` for a real entry at `p` (links and special files are `None`).
fn entry_kind(p: &Path) -> Option<String> {
    let m = fs::symlink_metadata(p).ok()?;
    if m.file_type().is_symlink() {
        None
    } else if m.is_dir() {
        Some("folder".into())
    } else if m.is_file() {
        Some("file".into())
    } else {
        None
    }
}

/// The destination name is free: nothing has that name, and no entry of its folder differs
/// from it only by letter case (a case-insensitive file system would treat them as one).
fn destination_free(to: &Path) -> std::result::Result<(), Verdict> {
    if fs::symlink_metadata(to).is_ok() {
        return Err((ItemStatus::Conflict, "destination_exists".into()));
    }
    let (Some(parent), Some(name)) = (to.parent(), to.file_name()) else {
        return Err(refused("destination_has_no_folder"));
    };
    let listing = fs::read_dir(parent).map_err(|_| refused("destination_folder_unreadable"))?;
    for e in listing {
        let other = e
            .map_err(|_| refused("destination_folder_unreadable"))?
            .file_name();
        if other != name && same_ignoring_case(&other, name) {
            return Err((ItemStatus::Conflict, "case_only_clash".into()));
        }
    }
    Ok(())
}

fn same_ignoring_case(a: &std::ffi::OsStr, b: &std::ffi::OsStr) -> bool {
    match (a.to_str(), b.to_str()) {
        (Some(a), Some(b)) => a.to_lowercase() == b.to_lowercase(),
        _ => a
            .as_encoded_bytes()
            .eq_ignore_ascii_case(b.as_encoded_bytes()),
    }
}

/// One rename that never replaces an existing name.
fn rename_exclusive(from: &Path, to: &Path) -> std::result::Result<(), Verdict> {
    rename_noreplace(from, to).map_err(|e| match e.raw_os_error() {
        _ if e.kind() == std::io::ErrorKind::AlreadyExists => {
            (ItemStatus::Conflict, "destination_exists".to_string())
        }
        // EXDEV on Linux and macOS.
        Some(18) => refused("cross_device"),
        _ if e.kind() == std::io::ErrorKind::Unsupported => refused("exclusive_rename_unsupported"),
        Some(code) => (ItemStatus::Refused, format!("rename_failed_os_{code}")),
        None => refused("rename_failed"),
    })
}

#[cfg(any(target_os = "macos", target_os = "linux"))]
fn rename_noreplace(from: &Path, to: &Path) -> std::io::Result<()> {
    use std::ffi::CString;
    use std::os::unix::ffi::OsStrExt;
    let f = CString::new(from.as_os_str().as_bytes()).map_err(std::io::Error::other)?;
    let t = CString::new(to.as_os_str().as_bytes()).map_err(std::io::Error::other)?;
    #[cfg(target_os = "macos")]
    // SAFETY: both arguments are NUL-terminated strings that outlive the call.
    let rc = unsafe { libc::renamex_np(f.as_ptr(), t.as_ptr(), libc::RENAME_EXCL) };
    #[cfg(target_os = "linux")]
    // SAFETY: both arguments are NUL-terminated strings that outlive the call.
    let rc = unsafe {
        libc::renameat2(
            libc::AT_FDCWD,
            f.as_ptr(),
            libc::AT_FDCWD,
            t.as_ptr(),
            libc::RENAME_NOREPLACE,
        )
    };
    if rc == 0 {
        Ok(())
    } else {
        let e = std::io::Error::last_os_error();
        // A file system without the exclusive flag answers EINVAL or ENOTSUP.
        if matches!(e.raw_os_error(), Some(libc::EINVAL) | Some(libc::ENOTSUP)) {
            return Err(std::io::ErrorKind::Unsupported.into());
        }
        Err(e)
    }
}

#[cfg(not(any(target_os = "macos", target_os = "linux")))]
fn rename_noreplace(_from: &Path, _to: &Path) -> std::io::Result<()> {
    Err(std::io::ErrorKind::Unsupported.into())
}

fn outcome(item: &JournalItem) -> String {
    let reason = item.reason.as_deref().unwrap_or("");
    match item.status {
        ItemStatus::Planned => "planned".into(),
        ItemStatus::Moving => "moving".into(),
        ItemStatus::Moved => "moved".into(),
        ItemStatus::Skipped => format!("skipped:{reason}"),
        ItemStatus::Conflict => "conflict".into(),
        ItemStatus::Refused => format!("refused:{reason}"),
        ItemStatus::NeedsReview => "needs_review".into(),
        ItemStatus::Reverted => "reverted".into(),
    }
}

fn report(path: &Path, j: &Journal, changed: bool) -> Report {
    Report {
        journal: path.display().to_string(),
        state: j.state,
        changed,
        items: j
            .items
            .iter()
            .map(|i| ItemReport {
                from: i.from.clone(),
                to: i.to.clone(),
                outcome: outcome(i),
            })
            .collect(),
    }
}

fn tmp_path(path: &Path) -> PathBuf {
    path.with_extension("json.tmp")
}

/// The journal bytes into `<path>.tmp`, fsynced. A link at the temp name is refused.
fn write_tmp(path: &Path, j: &Journal) -> Result<PathBuf> {
    let tmp = tmp_path(path);
    if fs::symlink_metadata(&tmp).is_ok_and(|m| !m.is_file()) {
        return Err(format!("{}: not a plain file", tmp.display()));
    }
    let bytes = serde_json::to_vec_pretty(j).map_err(|e| e.to_string())?;
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create(true)
        .truncate(true)
        .open(&tmp)
        .map_err(|e| format!("{}: {e}", tmp.display()))?;
    f.write_all(&bytes)
        .and_then(|()| f.sync_all())
        .map_err(|e| format!("{}: {e}", tmp.display()))?;
    Ok(tmp)
}

fn sync_dir(path: &Path) {
    if let Some(dir) = path.parent().and_then(|d| fs::File::open(d).ok()) {
        let _ = dir.sync_all();
    }
}

/// The first journal write: exclusive, so an existing journal is never replaced.
fn publish_new(path: &Path, j: &Journal) -> Result<()> {
    let tmp = write_tmp(path, j)?;
    rename_noreplace(&tmp, path).map_err(|e| format!("{}: {e}", path.display()))?;
    sync_dir(path);
    Ok(())
}

/// Every later journal write: temp, fsync, rename over, folder fsync.
fn save(path: &Path, j: &Journal) -> Result<()> {
    let tmp = write_tmp(path, j)?;
    fs::rename(&tmp, path).map_err(|e| format!("{}: {e}", path.display()))?;
    sync_dir(path);
    Ok(())
}

#[cfg(all(test, unix))]
mod tests {
    use super::*;
    use sha2::{Digest, Sha256};
    use std::os::unix::fs::{symlink, PermissionsExt};

    const NOW: &str = "2026-11-14T12:00:00Z";
    const STAMP: &str = "20261114T120000Z";

    struct Fixture {
        _tmp: tempfile::TempDir,
        tree: PathBuf,
        journals: PathBuf,
    }

    fn fixture() -> Fixture {
        let tmp = tempfile::tempdir().unwrap();
        let tree = tmp.path().join("tree");
        let journals = tmp.path().join("journals");
        fs::create_dir_all(tree.join("old")).unwrap();
        fs::create_dir_all(tree.join("new")).unwrap();
        fs::create_dir(&journals).unwrap();
        Fixture {
            _tmp: tmp,
            tree,
            journals,
        }
    }

    fn file(p: &Path, bytes: &[u8], mode: u32) {
        fs::write(p, bytes).unwrap();
        fs::set_permissions(p, fs::Permissions::from_mode(mode)).unwrap();
    }

    /// Names, sizes, modes and file bytes of the temp tree (links by target, not followed).
    fn digest_tree(root: &Path) -> String {
        fn walk(dir: &Path, rel: &str, h: &mut Sha256) {
            let mut names: Vec<_> = fs::read_dir(dir)
                .unwrap()
                .map(|e| e.unwrap().file_name())
                .collect();
            names.sort();
            for n in names {
                let p = dir.join(&n);
                let r = format!("{rel}/{}", n.to_string_lossy());
                let m = fs::symlink_metadata(&p).unwrap();
                h.update(format!("{r}\0{:o}\0{}\0", m.permissions().mode(), m.len()));
                if m.file_type().is_symlink() {
                    h.update(fs::read_link(&p).unwrap().to_string_lossy().as_bytes());
                } else if m.is_dir() {
                    walk(&p, &r, h);
                } else {
                    h.update(fs::read(&p).unwrap());
                }
            }
        }
        let mut h = Sha256::new();
        walk(root, "", &mut h);
        format!("{:x}", h.finalize())
    }

    /// Three movable items (a folder with a file inside, a file, a file named with spaces
    /// and a trailing dot) and a fourth whose destination is taken.
    fn plan(f: &Fixture) -> Vec<PlannedMove> {
        let (o, n) = (f.tree.join("old"), f.tree.join("new"));
        fs::create_dir(o.join("versions")).unwrap();
        file(&o.join("versions/a.bin"), b"alpha", 0o644);
        file(&o.join("path"), b"#!/bin/sh\n", 0o755);
        file(&o.join("my notes."), b"spaces and a dot", 0o600);
        file(&o.join("taken"), b"source", 0o644);
        file(&n.join("taken"), b"already here", 0o644);
        ["versions", "path", "my notes.", "taken"]
            .iter()
            .map(|name| PlannedMove {
                from: o.join(name),
                to: n.join(name),
            })
            .collect()
    }

    fn aliases() -> BTreeMap<String, String> {
        BTreeMap::from([("old-id".to_string(), "new-id".to_string())])
    }

    fn outcomes(r: &Report) -> Vec<&str> {
        r.items.iter().map(|i| i.outcome.as_str()).collect()
    }

    fn statuses(path: &Path) -> (JournalState, Vec<ItemStatus>) {
        let j = read_journal(path).unwrap();
        (j.state, j.items.iter().map(|i| i.status).collect())
    }

    fn env_with<'a>(in_use: &'a dyn Fn(&Path) -> InUse, fault: Option<&'a str>) -> MoveEnv<'a> {
        let mut env = MoveEnv::new(in_use);
        env.fault = fault;
        env
    }

    #[test]
    fn move_journal_happy_path_commits_then_reverts_byte_identical() {
        let f = fixture();
        let items = plan(&f);
        let before = digest_tree(&f.tree);
        let env = MoveEnv::new(&never_in_use);
        let r = apply(&f.journals, "test", &items, &aliases(), NOW, &env).unwrap();
        assert_eq!(r.state, JournalState::Committed);
        assert!(r.changed);
        assert_eq!(outcomes(&r), ["moved", "moved", "moved", "conflict"]);
        let path = journal_path(&f.journals, STAMP);
        let j = read_journal(&path).unwrap();
        assert_eq!(j.id_aliases, aliases());
        assert_eq!(j.kind, "test");
        assert_eq!(
            fs::read(f.tree.join("new/versions/a.bin")).unwrap(),
            b"alpha"
        );
        assert_eq!(
            fs::read(f.tree.join("new/my notes.")).unwrap(),
            b"spaces and a dot"
        );
        // The conflict overwrote nothing and moved nothing.
        assert_eq!(fs::read(f.tree.join("new/taken")).unwrap(), b"already here");
        assert_eq!(fs::read(f.tree.join("old/taken")).unwrap(), b"source");
        assert!(!f.tree.join("old/versions").exists());

        // Second apply of the committed journal: nothing changes (2회차 unchanged).
        let moved = digest_tree(&f.tree);
        let again = apply(&f.journals, "test", &items, &aliases(), NOW, &env).unwrap();
        assert!(!again.changed);
        assert_eq!(again.state, JournalState::Committed);
        assert_eq!(digest_tree(&f.tree), moved);
        let resumed = resume(&path, &env).unwrap();
        assert!(!resumed.changed);
        assert_eq!(digest_tree(&f.tree), moved);

        let back = revert(&path).unwrap();
        assert_eq!(back.state, JournalState::Reverted);
        assert_eq!(
            outcomes(&back),
            ["reverted", "reverted", "reverted", "conflict"]
        );
        assert_eq!(digest_tree(&f.tree), before);
        // A second revert is a no-op.
        assert!(!revert(&path).unwrap().changed);
        assert_eq!(digest_tree(&f.tree), before);
    }

    #[test]
    fn move_journal_another_plan_with_the_same_stamp_is_refused() {
        let f = fixture();
        let items = plan(&f);
        let env = MoveEnv::new(&never_in_use);
        apply(&f.journals, "test", &items, &aliases(), NOW, &env).unwrap();
        let err = apply(&f.journals, "test", &items[..1], &aliases(), NOW, &env).unwrap_err();
        assert!(err.contains("another plan"), "{err}");
    }

    /// The state sequence on disk at each stop point, then resume and revert.
    #[test]
    fn move_journal_crash_points_resume_then_revert_restore_the_tree() {
        use ItemStatus::*;
        let cases: [(&str, Vec<ItemStatus>, bool); 4] = [
            (
                "after_journal",
                vec![Planned, Planned, Planned, Planned],
                false,
            ),
            ("after_move:1", vec![Moved, Planned, Planned, Planned], true),
            ("after_move:3", vec![Moved, Moved, Moved, Planned], true),
            ("before_commit", vec![Moved, Moved, Moved, Conflict], true),
        ];
        for (point, expected, moved_first) in cases {
            let f = fixture();
            let items = plan(&f);
            let before = digest_tree(&f.tree);
            let crash = env_with(&never_in_use, Some(point));
            let err = apply(&f.journals, "test", &items, &aliases(), NOW, &crash).unwrap_err();
            assert!(err.contains(point), "{point}: {err}");
            let path = journal_path(&f.journals, STAMP);
            assert_eq!(
                statuses(&path),
                (JournalState::Pending, expected),
                "{point}: journal on disk at the stop"
            );
            assert_eq!(
                !f.tree.join("old/versions").exists(),
                moved_first,
                "{point}"
            );

            let env = MoveEnv::new(&never_in_use);
            let r = resume(&path, &env).unwrap();
            assert_eq!(r.state, JournalState::Committed, "{point}");
            assert_eq!(
                outcomes(&r),
                ["moved", "moved", "moved", "conflict"],
                "{point}"
            );
            assert_eq!(fs::read(f.tree.join("new/taken")).unwrap(), b"already here");
            let back = revert(&path).unwrap();
            assert_eq!(back.state, JournalState::Reverted, "{point}");
            assert_eq!(digest_tree(&f.tree), before, "{point}: tree after revert");
        }
    }

    /// A stop between marking an item `moving` and recording its rename: resume does not
    /// guess, the item stays where it is and is reported for review.
    #[test]
    fn move_journal_resume_reports_interrupted_item_for_review() {
        let f = fixture();
        let items = plan(&f);
        let before = digest_tree(&f.tree);
        let crash = env_with(&never_in_use, Some("after_mark_moving"));
        apply(&f.journals, "test", &items, &aliases(), NOW, &crash).unwrap_err();
        let path = journal_path(&f.journals, STAMP);
        assert_eq!(statuses(&path).1[0], ItemStatus::Moving);
        let env = MoveEnv::new(&never_in_use);
        let r = resume(&path, &env).unwrap();
        assert_eq!(outcomes(&r), ["needs_review", "moved", "moved", "conflict"]);
        assert!(f.tree.join("old/versions/a.bin").exists());
        assert!(!f.tree.join("new/versions").exists());
        // A later resume leaves the review item alone too.
        let again = resume(&path, &env).unwrap();
        assert_eq!(outcomes(&again)[0], "needs_review");
        assert!(f.tree.join("old/versions/a.bin").exists());
        revert(&path).unwrap();
        assert_eq!(digest_tree(&f.tree), before);
    }

    /// A rename that happened without its record (the source gone, the destination there):
    /// resume reports it for review instead of guessing.
    #[test]
    fn move_journal_resume_planned_item_already_at_destination_needs_review() {
        let f = fixture();
        let items = plan(&f);
        let crash = env_with(&never_in_use, Some("after_journal"));
        apply(&f.journals, "test", &items, &aliases(), NOW, &crash).unwrap_err();
        fs::rename(f.tree.join("old/path"), f.tree.join("new/path")).unwrap();
        let path = journal_path(&f.journals, STAMP);
        let r = resume(&path, &MoveEnv::new(&never_in_use)).unwrap();
        assert_eq!(outcomes(&r), ["moved", "needs_review", "moved", "conflict"]);
    }

    #[test]
    fn move_journal_symlinked_source_and_destination_folder_are_refused() {
        let f = fixture();
        let (o, n) = (f.tree.join("old"), f.tree.join("new"));
        fs::create_dir(f.tree.join("real")).unwrap();
        symlink(f.tree.join("real"), o.join("link")).unwrap();
        file(&o.join("plain"), b"x", 0o644);
        symlink(f.tree.join("real"), n.join("linked-dir")).unwrap();
        let items = vec![
            PlannedMove {
                from: o.join("link"),
                to: n.join("link"),
            },
            PlannedMove {
                from: o.join("plain"),
                to: n.join("linked-dir/plain"),
            },
        ];
        let before = digest_tree(&f.tree);
        let r = apply(
            &f.journals,
            "test",
            &items,
            &aliases(),
            NOW,
            &MoveEnv::new(&never_in_use),
        )
        .unwrap();
        assert_eq!(
            outcomes(&r),
            [
                "refused:source_is_link",
                "refused:destination_folder_is_link"
            ]
        );
        assert!(!r.changed);
        assert_eq!(digest_tree(&f.tree), before);
    }

    /// Cross-device is simulated through the injectable device check (the test machine has
    /// no guaranteed second file system).
    #[test]
    fn move_journal_cross_device_is_refused_never_copied() {
        let f = fixture();
        file(&f.tree.join("old/a"), b"a", 0o644);
        let items = vec![PlannedMove {
            from: f.tree.join("old/a"),
            to: f.tree.join("new/a"),
        }];
        let before = digest_tree(&f.tree);
        let other = |_: &Path, _: &Path| Some(false);
        let unknown = |_: &Path, _: &Path| None;
        let mut env = MoveEnv::new(&never_in_use);
        env.same_device = &other;
        let r = apply(&f.journals, "test", &items, &aliases(), NOW, &env).unwrap();
        assert_eq!(outcomes(&r), ["refused:cross_device"]);
        env.same_device = &unknown;
        let r = apply(
            &f.journals,
            "test",
            &items,
            &aliases(),
            "2026-11-14T12:00:01Z",
            &env,
        )
        .unwrap();
        assert_eq!(outcomes(&r), ["refused:device_unknown"]);
        assert_eq!(digest_tree(&f.tree), before);
        // The real check says the temp tree is one device.
        assert_eq!(
            same_device(&f.tree.join("old/a"), &f.tree.join("new")),
            Some(true)
        );
    }

    #[test]
    fn move_journal_case_only_clash_is_a_conflict() {
        let f = fixture();
        file(&f.tree.join("old/Data"), b"src", 0o644);
        file(&f.tree.join("new/data"), b"dst", 0o644);
        let items = vec![PlannedMove {
            from: f.tree.join("old/Data"),
            to: f.tree.join("new/Data"),
        }];
        let before = digest_tree(&f.tree);
        let r = apply(
            &f.journals,
            "test",
            &items,
            &aliases(),
            NOW,
            &MoveEnv::new(&never_in_use),
        )
        .unwrap();
        assert_eq!(outcomes(&r), ["conflict"]);
        assert_eq!(digest_tree(&f.tree), before);
    }

    #[test]
    fn move_journal_destination_check_edges() {
        let f = fixture();
        let n = f.tree.join("new");
        file(&n.join("Taken"), b"t", 0o644);
        // Exactly the existing name, and a case-only variant of it: not free.
        assert!(destination_free(&n.join("Taken")).is_err());
        assert!(destination_free(&n.join("taken")).is_err());
        // One letter more: free.
        assert!(destination_free(&n.join("Takenx")).is_ok());
        // The exclusive rename itself refuses an existing name, folders included.
        file(&f.tree.join("old/f"), b"f", 0o644);
        let e = rename_exclusive(&f.tree.join("old/f"), &n.join("Taken")).unwrap_err();
        assert_eq!(e.0, ItemStatus::Conflict);
        fs::create_dir(f.tree.join("old/d")).unwrap();
        fs::create_dir(n.join("d")).unwrap();
        assert!(rename_exclusive(&f.tree.join("old/d"), &n.join("d")).is_err());
        assert_eq!(fs::read(n.join("Taken")).unwrap(), b"t");
        assert!(f.tree.join("old/f").exists() && f.tree.join("old/d").exists());
    }

    #[test]
    fn move_journal_in_use_yes_and_unknown_are_skipped() {
        let f = fixture();
        let o = f.tree.join("old");
        for n in ["open", "unsure", "free"] {
            file(&o.join(n), n.as_bytes(), 0o644);
        }
        let items: Vec<_> = ["open", "unsure", "free"]
            .iter()
            .map(|n| PlannedMove {
                from: o.join(n),
                to: f.tree.join("new").join(n),
            })
            .collect();
        let check = |p: &Path| match p.file_name().and_then(|n| n.to_str()) {
            Some("open") => InUse::Yes("pid_42_open".into()),
            Some("unsure") => InUse::Unknown,
            _ => InUse::No,
        };
        let r = apply(
            &f.journals,
            "test",
            &items,
            &aliases(),
            NOW,
            &MoveEnv::new(&check),
        )
        .unwrap();
        assert_eq!(
            outcomes(&r),
            ["skipped:pid_42_open", "skipped:in_use_unknown", "moved"]
        );
        assert!(o.join("open").exists() && o.join("unsure").exists());
    }

    #[test]
    fn move_journal_revert_with_original_recreated_is_conflict_and_deletes_nothing() {
        let f = fixture();
        let items = plan(&f);
        let path = journal_path(&f.journals, STAMP);
        apply(
            &f.journals,
            "test",
            &items,
            &aliases(),
            NOW,
            &MoveEnv::new(&never_in_use),
        )
        .unwrap();
        // Something re-created the original `path` (a file) and `versions` (a folder).
        file(&f.tree.join("old/path"), b"recreated", 0o644);
        fs::create_dir(f.tree.join("old/versions")).unwrap();
        let r = revert(&path).unwrap();
        assert_eq!(r.state, JournalState::RevertPartial);
        assert_eq!(
            outcomes(&r),
            ["conflict", "conflict", "reverted", "conflict"]
        );
        assert_eq!(fs::read(f.tree.join("old/path")).unwrap(), b"recreated");
        assert_eq!(fs::read(f.tree.join("new/path")).unwrap(), b"#!/bin/sh\n");
        assert_eq!(
            fs::read(f.tree.join("new/versions/a.bin")).unwrap(),
            b"alpha"
        );
        assert_eq!(
            fs::read_dir(f.tree.join("old/versions")).unwrap().count(),
            0
        );
        // The conflicting items stay `moved` in the journal, so a later revert can retry.
        let j = read_journal(&path).unwrap();
        assert_eq!(j.items[0].status, ItemStatus::Moved);
    }

    #[test]
    fn move_journal_rejects_relative_paths_and_bad_names() {
        let f = fixture();
        let items = vec![PlannedMove {
            from: PathBuf::from("rel"),
            to: f.tree.join("new/rel"),
        }];
        let env = MoveEnv::new(&never_in_use);
        assert!(apply(&f.journals, "t", &items, &aliases(), NOW, &env).is_err());
        assert!(apply(&f.journals, "t", &[], &aliases(), "now", &env).is_err());
        assert!(fs::read_dir(&f.journals).unwrap().next().is_none());
        assert!(read_journal(&f.journals.join("other.json")).is_err());
    }
}
