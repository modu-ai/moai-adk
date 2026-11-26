//! What earlier generations leave behind that no step of the project update removes, and what the
//! user can do about each (SPEC-20261127-update-leftovers-action, plan side).
//!
//! #2936 L2b (completion condition 3 of the Issue): every leftover is a fact first (path, one
//! line, and that the update does not remove it). A fact needs its proof:
//!
//! - the earlier state folder: a real folder while the state already lives in the new state
//!   folder (state_root `rust`). Before that it is the live state folder (or moves with the
//!   go-state step) and is no leftover.
//! - `CLAUDE.md`: a regular file with Go proof (the Go manifest records it, or it holds a Go
//!   curator block) and no managed section of this generation, while the Go state folder is not
//!   there for the go-state step to move it with. It stays a fact; another step owns it.
//! - a Go update lock (in the state folder or in the Go backup folder): the lock file whose
//!   recorded process is not running (`kill -0`; a process that cannot be judged, or a lock that
//!   cannot be read, is not stated).
//!
//! For a project whose state generation is exactly `rust` each of five kinds also carries what
//! the user can do about it: an `action` (the printed command of the archive apply, one plan
//! digest for every item) or a `keep_only` reason, and a `keep` command that records "leave it
//! here" (`declined`). The five kinds are the earlier state folder, a `<state>.rust-<UTC>/`
//! folder an earlier revert set aside, the Go `.moai-backups/` folder, and the Go update lock in
//! the state folder and in the Go backup folder. The plan never moves, writes or deletes anything
//! and never reads a file's contents except a lock's bytes and the go-state receipts it must
//! classify; a link is never followed.

use super::declined::{self, Record};
use super::guide::OpenLeftover;
use crate::{go_residue, go_state, home_state, state_root};
use serde_json::{json, Value};
use std::{fs, path::Path};

/// What every leftover says it is not (#2936 L2b).
const LEFTOVER_NOTE: &str = "not removed by update";
/// A lock file is a few lines of JSON; this only refuses absurd inputs.
pub const LOCK_LIMIT: u64 = 4096;
/// How many entry names a folder's fact lists.
const LEFTOVER_NAMES: usize = 8;
/// How many entries the nested-checkout walk looks at before it gives up (unknown keeps the
/// folder where it is).
const WALK_LIMIT: u64 = 20_000;
/// The subcommand an `action` prints.
pub const APPLY_OPERATION: &str = "project-update-leftovers-apply";
/// The Go tool's backups folder, which is not the Go backup folder (`state_root::GO_BACKUP`).
pub const GO_BACKUPS: &str = ".moai-backups";
/// The file name of a Go update lock.
const LOCK_FILE: &str = ".update.lock";
/// The root instruction files a go-state apply may have copied into a backup.
const ROOT_FILES: [&str; 3] = ["CLAUDE.md", "AGENTS.md", "AGENTS.local.md"];
/// How the lock fact names the lock's place.
const STATE_LOCK_PLACE: &str = "the Go update lock";
const BACKUP_LOCK_PLACE: &str = "the Go update lock in the Go backup folder";

/// The five kinds of leftover that can be archived or kept.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    /// The earlier state folder.
    Earlier,
    /// A `<state>.rust-<name>/` folder.
    Aside,
    /// The Go `.moai-backups/` folder.
    Backups,
    /// The Go update lock in the state folder.
    StateLock,
    /// The Go update lock in the Go backup folder.
    BackupLock,
}

impl Kind {
    pub fn is_folder(self) -> bool {
        matches!(self, Kind::Earlier | Kind::Aside | Kind::Backups)
    }
}

/// The name a revert sets this generation's state folder aside under, up to the UTC stamp.
pub fn aside_prefix() -> String {
    format!("{}.rust-", state_root::MOAI_DIR)
}

fn state_lock() -> String {
    format!("{}/{LOCK_FILE}", state_root::MOAI_DIR)
}

fn backup_lock() -> String {
    format!("{}/{LOCK_FILE}", state_root::GO_BACKUP)
}

/// The path without the slash a folder's printed path ends in.
pub fn rel_of(path: &str) -> &str {
    path.strip_suffix('/').unwrap_or(path)
}

/// The kind of a printed leftover path, `None` for anything that is not one of the five (so
/// `CLAUDE.md`, a path outside the project and an empty string are none). A folder path ends in
/// `/`, a lock path does not; a set-aside folder is any `<state>.rust-<name>/` here, and whether
/// its name carries a UTC stamp is a reason to keep it, not a reason to hide it.
pub fn known_path(path: &str) -> Option<Kind> {
    if path == format!("{}/", state_root::EARLIER_PROJECT_DIR) {
        return Some(Kind::Earlier);
    }
    if path == format!("{GO_BACKUPS}/") {
        return Some(Kind::Backups);
    }
    if path == state_lock() {
        return Some(Kind::StateLock);
    }
    if path == backup_lock() {
        return Some(Kind::BackupLock);
    }
    path.strip_suffix('/')?
        .strip_prefix(aside_prefix().as_str())
        .filter(|name| !name.is_empty() && !name.contains('/'))
        .map(|_| Kind::Aside)
}

/// One leftover of the five kinds, as the plan reads it now.
#[derive(Clone, Debug)]
pub struct Item {
    /// As printed: a folder ends in `/`.
    pub path: String,
    pub kind: Kind,
    pub fact: String,
    /// The source sha256 now (`listing_digest` of a folder, the bytes' sha256 of a lock); `None`
    /// when it cannot be read.
    pub sha256: Option<String>,
    /// `Ok` when the archive apply may move it, else the reason it can only be kept.
    pub archive: Result<(), String>,
}

#[cfg(unix)]
fn name_bytes(name: &std::ffi::OsStr) -> Vec<u8> {
    use std::os::unix::ffi::OsStrExt;
    name.as_bytes().to_vec()
}

#[cfg(not(unix))]
fn name_bytes(name: &std::ffi::OsStr) -> Vec<u8> {
    name.to_string_lossy().into_owned().into_bytes()
}

/// The source sha256 of a folder (REQ-11): its sorted top-level entry names, kinds and file
/// sizes, nothing below the first level and no file's contents, so a changed name or size reads
/// as a new source and the same size with other bytes does not. A directory's own size is not
/// part of it (it moves on some systems without any change). The folder itself must be a real
/// folder; a link is refused, never followed.
pub fn listing_digest(dir: &Path) -> Result<String, String> {
    let own = fs::symlink_metadata(dir).map_err(|_| "the folder cannot be read".to_string())?;
    if !own.is_dir() {
        return Err("it is not a real folder".into());
    }
    let mut rows: Vec<(Vec<u8>, &'static str, u64)> = Vec::new();
    let entries = fs::read_dir(dir).map_err(|_| "the folder cannot be listed".to_string())?;
    for entry in entries {
        let entry = entry.map_err(|_| "the folder cannot be listed".to_string())?;
        let meta = entry
            .metadata()
            .map_err(|_| "an entry of the folder cannot be read".to_string())?;
        let (kind, size) = if meta.file_type().is_symlink() {
            ("link", 0)
        } else if meta.is_dir() {
            ("dir", 0)
        } else if meta.is_file() {
            ("file", meta.len())
        } else {
            ("other", 0)
        };
        rows.push((name_bytes(&entry.file_name()), kind, size));
    }
    rows.sort();
    let mut bytes = Vec::new();
    for (name, kind, size) in rows {
        bytes.extend_from_slice(&(name.len() as u64).to_be_bytes());
        bytes.extend_from_slice(&name);
        bytes.push(0);
        bytes.extend_from_slice(kind.as_bytes());
        bytes.push(0);
        bytes.extend_from_slice(&size.to_be_bytes());
    }
    Ok(crate::digest(&bytes))
}

/// The one plan digest of a plan's leftovers: the path and source sha256 of every item the archive
/// apply may move, in the order the plan lists them. It depends on nothing else (not on the user's
/// kept records and not on whether another step is waiting), so the apply can rebuild it from disk.
pub fn plan_digest(items: &[Item]) -> String {
    let mut text = String::from("leftovers\n");
    for item in items.iter().filter(|i| i.archive.is_ok()) {
        text.push_str(&format!(
            "{}\0{}\n",
            item.path,
            item.sha256.as_deref().unwrap_or_default()
        ));
    }
    crate::digest(text.as_bytes())
}

/// What a go-state receipt's state word says about the folder that holds it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ReceiptState {
    Pending,
    Committed,
    Reverted,
    /// A state word this version does not know.
    Unknown,
}

/// The state word of a go-state receipt, as the go-state writer spells it.
pub fn receipt_state(word: &str) -> ReceiptState {
    match word {
        "pending" => ReceiptState::Pending,
        "committed" => ReceiptState::Committed,
        "reverted" => ReceiptState::Reverted,
        _ => ReceiptState::Unknown,
    }
}

/// Whether a go-state receipt in `state` lets its folder be archived (REQ-4, REQ-5). A pending
/// receipt is an apply that did not finish and an unknown state is not judged, so neither moves.
/// A receipt in a folder a revert set aside (`set_aside`) must be `reverted`: a `committed` one
/// there means a revert stopped. In the earlier state folder a `committed` receipt is a finished
/// move and is fine. `stamp` names the receipt's folder in the reason.
pub fn receipt_verdict(state: ReceiptState, set_aside: bool, stamp: &str) -> Result<(), String> {
    let name = crate::card::input::echo(stamp);
    match state {
        ReceiptState::Pending => Err(format!(
            "a go-state receipt inside ({name}) is still pending, so an apply did not finish"
        )),
        ReceiptState::Unknown => Err(format!(
            "a go-state receipt inside ({name}) is in a state this version does not know"
        )),
        ReceiptState::Committed if set_aside => Err(format!(
            "a go-state receipt inside ({name}) is committed, so a revert stopped before it finished"
        )),
        _ => Ok(()),
    }
}

/// Judges every go-state receipt below `<folder>/backups/<name>/`. Only that place is read, each
/// through the go-state reader's own bounds (no link, a regular file, at most the receipt limit)
/// and parsed as the go-state receipt it is; anything else is unreadable and keeps the folder.
fn receipts(root: &Path, folder: &str, set_aside: bool) -> Result<(), String> {
    let backups = format!("{folder}/backups");
    match fs::symlink_metadata(root.join(&backups)) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(()),
        Err(_) => {
            return Err("its backups entry cannot be read, so its receipts are not judged".into())
        }
        Ok(m) if m.is_dir() => {}
        Ok(_) => {
            return Err(
                "its backups entry is not a plain folder, so its receipts are not judged".into(),
            )
        }
    }
    let mut names = Vec::new();
    let listing = fs::read_dir(root.join(&backups))
        .map_err(|_| "its backups folder cannot be listed, so its receipts are not judged")?;
    for entry in listing {
        let name = entry
            .ok()
            .and_then(|e| e.file_name().into_string().ok())
            .ok_or("a name in its backups folder cannot be read, so its receipts are not judged")?;
        names.push(name);
    }
    names.sort();
    for name in names {
        let place = format!("{backups}/{name}");
        // Only a real folder can hold a receipt; a file or a link there holds none we read.
        if !fs::symlink_metadata(root.join(&place)).is_ok_and(|m| m.is_dir()) {
            continue;
        }
        let file = format!("{place}/{}", state_root::GO_STATE_RECEIPT);
        match fs::symlink_metadata(root.join(&file)) {
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => continue,
            _ => {}
        }
        let unreadable = || {
            format!(
                "a go-state receipt inside ({}) cannot be read as a receipt (a link, over the \
                 size limit or not a receipt)",
                crate::card::input::echo(&name)
            )
        };
        let text = go_state::read_bounded_text(root, &file, state_root::GO_STATE_RECEIPT_MAX)
            .ok()
            .flatten()
            .ok_or_else(unreadable)?;
        let receipt: go_state::Receipt = serde_json::from_str(&text).map_err(|_| unreadable())?;
        receipt_verdict(receipt_state(&receipt.state), set_aside, &name)?;
    }
    Ok(())
}

/// REQ-4: no merge step still reads a backed-up copy of a root instruction file or of the shared
/// settings from the earlier state folder. A reader that cannot tell keeps the folder.
fn merge_source_guard(root: &Path) -> Result<(), String> {
    let inside = format!("{}/", state_root::EARLIER_PROJECT_DIR);
    let unsure = || {
        "it cannot be told whether a merge step still reads a backed-up copy from it".to_string()
    };
    for name in ROOT_FILES {
        match go_state::backed_up_root_file(root, name) {
            Ok(Some(b)) if b.rel.starts_with(&inside) => {
                return Err(format!(
                    "a merge step still reads the backed-up copy of {name} from it"
                ))
            }
            Ok(_) => {}
            Err(_) => return Err(unsure()),
        }
    }
    match go_state::backed_up_settings(root) {
        Ok(Some(b)) if b.rel.starts_with(&inside) => {
            Err("a merge step still reads the backed-up copy of the settings from it".into())
        }
        Ok(_) => Ok(()),
        Err(_) => Err(unsure()),
    }
}

/// Whether `<root>/<rel>` is on the same filesystem as the backup area, so one rename moves it.
/// The area may not exist yet: its nearest existing ancestor stands for it.
#[cfg(unix)]
fn same_filesystem(root: &Path, rel: &str) -> bool {
    use std::os::unix::fs::MetadataExt;
    let Ok(item) = fs::symlink_metadata(root.join(rel)) else {
        return false;
    };
    let mut area = root.join(state_root::MOAI_DIR).join("backups");
    loop {
        if let Ok(meta) = fs::symlink_metadata(&area) {
            return meta.dev() == item.dev();
        }
        if !area.pop() || !area.starts_with(root) {
            return false;
        }
    }
}

#[cfg(not(unix))]
fn same_filesystem(_root: &Path, _rel: &str) -> bool {
    true
}

/// A real folder at `rel`, not a link.
fn real_dir(root: &Path, rel: &str) -> bool {
    fs::symlink_metadata(root.join(rel)).is_ok_and(|m| m.is_dir())
}

/// REQ-3 to REQ-6: why a folder cannot be archived, `Ok` when it can.
fn judge_folder(root: &Path, kind: Kind, rel: &str) -> Result<(), String> {
    if kind == Kind::Aside
        && !rel
            .strip_prefix(aside_prefix().as_str())
            .is_some_and(state_root::utc_stamp)
    {
        return Err(format!(
            "its name has no UTC stamp after {}, so it is not a folder an earlier revert set aside",
            aside_prefix()
        ));
    }
    if !go_residue::contained(root, rel) {
        return Err("its path is not a plain relative path or passes through a link".into());
    }
    match go_state::nested_checkouts_under(root, rel, WALK_LIMIT) {
        Some(0) => {}
        Some(_) => {
            return Err(
                "it holds a nested checkout (a folder with its own .git), which a move would \
                 take with it"
                    .into(),
            )
        }
        None => return Err("it holds too many entries to rule out a nested checkout".into()),
    }
    if !same_filesystem(root, rel) {
        return Err(
            "it is on another filesystem than the backup area, so one rename cannot move it".into(),
        );
    }
    match kind {
        Kind::Earlier => {
            receipts(root, rel, false)?;
            merge_source_guard(root)
        }
        Kind::Aside => receipts(root, rel, true),
        _ => Ok(()),
    }
}

/// The part of a folder's fact that says what it holds: a few entry names, never read.
fn held_phrase(dir: &Path) -> String {
    let mut names: Vec<String> = fs::read_dir(dir)
        .map(|rd| {
            rd.flatten()
                .map(|e| e.file_name().to_string_lossy().into_owned())
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    match names.len() {
        0 => "is an empty folder".to_string(),
        n => {
            // File names are the project's, not ours (AUDIT #3117 E4): a long one is cut and a
            // control or direction character shows as `?`.
            let shown = names
                .iter()
                .take(LEFTOVER_NAMES)
                .map(|n| crate::card::input::echo(n))
                .collect::<Vec<_>>()
                .join(", ");
            let more = n.saturating_sub(LEFTOVER_NAMES);
            if more > 0 {
                format!("still holds {n} entries ({shown}, and {more} more)")
            } else {
                format!("still holds {n} entries ({shown})")
            }
        }
    }
}

fn folder_fact(root: &Path, kind: Kind, rel: &str) -> String {
    let held = held_phrase(&root.join(rel));
    match kind {
        Kind::Earlier => format!(
            "the earlier generation's state folder {held}; this generation's state is in .moai/"
        ),
        Kind::Aside => format!(
            "a copy of this generation's state folder that an earlier revert set aside {held}"
        ),
        _ => format!("the Go tool's backups folder {held}"),
    }
}

fn folder_item(root: &Path, kind: Kind, name: &str) -> Item {
    let sha256 = listing_digest(&root.join(name)).ok();
    let archive = if sha256.is_some() {
        judge_folder(root, kind, name)
    } else {
        Err("its top-level listing cannot be read".into())
    };
    Item {
        path: format!("{name}/"),
        kind,
        fact: folder_fact(root, kind, name),
        sha256,
        archive,
    }
}

/// The `<state>.rust-<name>` real folders directly under the project, sorted.
fn asides(root: &Path) -> Vec<String> {
    let prefix = aside_prefix();
    let mut names: Vec<String> = fs::read_dir(root)
        .map(|rd| {
            rd.flatten()
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| n.len() > prefix.len() && n.starts_with(&prefix))
                .collect()
        })
        .unwrap_or_default();
    names.retain(|n| real_dir(root, n));
    names.sort();
    names
}

/// The fact and the source sha256 of a Go update lock at `rel` whose recorded process is not
/// running; `None` for no lock, a link, a lock that cannot be read as the Go payload
/// `{pid, started_at, hostname}`, or a process that is alive or cannot be judged. The bytes are
/// read once; the hostname and every other key stay unprinted.
pub fn dead_update_lock(root: &Path, rel: &str, place: &str) -> Option<(String, String)> {
    let path = root.join(rel);
    let m = fs::symlink_metadata(&path).ok()?;
    if !m.is_file() || m.len() > LOCK_LIMIT || crate::apply::target(root, rel).is_err() {
        return None;
    }
    let bytes = fs::read(&path).ok()?;
    if bytes.len() as u64 > LOCK_LIMIT {
        return None;
    }
    let payload: Value = serde_json::from_slice(&bytes).ok()?;
    let pid = payload["pid"]
        .as_i64()
        .filter(|p| *p > 0 && *p <= i64::from(i32::MAX))?;
    if home_state::pid_alive(pid)? {
        return None;
    }
    // The time is shown only when it looks like one: the file is not trusted text.
    let since = payload["started_at"]
        .as_str()
        .filter(|s| {
            s.len() <= 32
                && s.chars()
                    .all(|c| c.is_ascii_digit() || matches!(c, '-' | ':' | 'T' | 'Z' | '.' | '+'))
        })
        .map(|s| format!(" (started {s})"))
        .unwrap_or_default();
    Some((
        format!("{place} names process {pid}{since}, which is not running"),
        crate::digest(&bytes),
    ))
}

/// REQ-7: a lock is archivable as a regular file whose process is not running (what
/// [`dead_update_lock`] proved), reached without a link and on the backup area's filesystem.
fn judge_lock(root: &Path, rel: &str) -> Result<(), String> {
    if !go_residue::contained(root, rel) {
        return Err("its path passes through a link".into());
    }
    if !same_filesystem(root, rel) {
        return Err(
            "it is on another filesystem than the backup area, so it cannot be moved by one link"
                .into(),
        );
    }
    Ok(())
}

fn lock_item(root: &Path, kind: Kind, rel: &str, place: &str) -> Option<Item> {
    let (fact, sha) = dead_update_lock(root, rel, place)?;
    Some(Item {
        path: rel.to_string(),
        kind,
        fact,
        sha256: Some(sha),
        archive: judge_lock(root, rel),
    })
}

/// The five kinds as the project has them now: folders first (the earlier state folder, the
/// set-aside folders, the Go backups folder), then the locks. Only a project whose state
/// generation is exactly `rust` has any; every other generation gets none.
pub fn survey(root: &Path, generation: &str) -> Vec<Item> {
    if generation != "rust" {
        return Vec::new();
    }
    let mut out = Vec::new();
    let earlier = state_root::EARLIER_PROJECT_DIR;
    if real_dir(root, earlier) {
        out.push(folder_item(root, Kind::Earlier, earlier));
    }
    for name in asides(root) {
        out.push(folder_item(root, Kind::Aside, &name));
    }
    if real_dir(root, GO_BACKUPS) {
        out.push(folder_item(root, Kind::Backups, GO_BACKUPS));
    }
    out.extend(lock_item(
        root,
        Kind::StateLock,
        &state_lock(),
        STATE_LOCK_PLACE,
    ));
    out.extend(lock_item(
        root,
        Kind::BackupLock,
        &backup_lock(),
        BACKUP_LOCK_PLACE,
    ));
    out
}

/// The source sha256 `path` has now, `None` when it is not listed (or cannot be read) now. Read
/// only; this is what a kept record is compared against.
pub fn source_sha256(root: &Path, path: &str) -> Option<String> {
    let generation = state_root::classify(root).generation;
    survey(root, generation)
        .into_iter()
        .find(|i| i.path == path)
        .and_then(|i| i.sha256)
}

fn plain_entry(path: &str, fact: String) -> Value {
    json!({"path": path, "fact": fact, "note": LEFTOVER_NOTE})
}

/// The entry of one of the five kinds: the fact, what the user can do and whether they already
/// said to keep it for its current source.
fn entry(item: &Item, shown: &str, digest: &str, others_pending: bool, record: &Record) -> Value {
    let key = format!("{}{}", declined::LEFTOVER_PREFIX, item.path);
    let kept = item
        .sha256
        .as_deref()
        .is_some_and(|sha| record.get(&key).map(String::as_str) == Some(sha));
    let mut v = plain_entry(&item.path, item.fact.clone());
    v["kept"] = json!(kept);
    if let Some(sha) = &item.sha256 {
        v["keep"] = json!(declined::decline_argv(shown, &key, sha));
        if kept {
            v["ask_again"] = json!(declined::ask_again_argv(shown, &key));
        }
    }
    match &item.archive {
        Err(reason) => v["keep_only"] = json!(reason),
        // Another backup step's move would make this plan digest stale: the command waits.
        Ok(()) if others_pending => {
            v["action_waits"] = json!(
                "the command to move it is printed after the earlier backup steps have run and \
                 the plan is made again"
            );
        }
        Ok(()) => {
            let mut words = super::argv(APPLY_OPERATION, shown, digest);
            words.extend(["--item".to_string(), item.path.clone()]);
            v["action"] = json!(words);
        }
    }
    v
}

/// The `"leftovers"` array of the plan, in the order the facts have always had (the earlier state
/// folder, the folders after it, `CLAUDE.md`, the locks).
///
/// - `generation`: the state generation word of the project.
/// - `others_pending`: another backup-phase step still has a command to run (REQ-2): every
///   leftover then prints its `keep` and no `action`.
/// - `record`: the user's recorded answers (`declined::read`).
pub fn view(root: &Path, generation: &str, others_pending: bool, record: &Record) -> Vec<Value> {
    let shown = root.to_str().unwrap_or_default();
    let items = survey(root, generation);
    let digest = plan_digest(&items);
    let mut out: Vec<Value> = items
        .iter()
        .filter(|i| i.kind.is_folder())
        .map(|i| entry(i, shown, &digest, others_pending, record))
        .collect();
    if generation != "go" {
        if let Some(proof) = go_state::go_claude_md_proof(root) {
            let how = if proof == "manifest" {
                "the Go manifest records it"
            } else {
                "it holds a Go curator block"
            };
            out.push(plain_entry(
                "CLAUDE.md",
                format!("a Go-generation instruction file ({how}) is still in the project root"),
            ));
        }
    }
    if generation == "rust" {
        out.extend(
            items
                .iter()
                .filter(|i| !i.kind.is_folder())
                .map(|i| entry(i, shown, &digest, others_pending, record)),
        );
    } else if let Some((fact, _)) = dead_update_lock(root, &state_lock(), STATE_LOCK_PLACE) {
        // The state folder has not moved yet: the go-state step carries the lock along.
        out.push(plain_entry(&state_lock(), fact));
    }
    out
}

/// The leftovers of the five kinds that still wait for an answer: neither archived nor kept for
/// the source they have now (REQ-8). `CLAUDE.md` is not one of them. The flag says whether an
/// archive is offered for the item (a printed command or one that waits), as opposed to a keep
/// only.
pub fn open(view: &[Value]) -> Vec<OpenLeftover<'_>> {
    view.iter()
        .filter(|l| l["kept"] == json!(false))
        .filter_map(|l| {
            Some(OpenLeftover {
                path: l["path"].as_str()?,
                archive: l.get("keep_only").is_none(),
            })
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    const STAMP: &str = "20261101T000000Z";

    fn project() -> tempfile::TempDir {
        tempfile::tempdir().unwrap()
    }

    fn put(root: &Path, rel: &str, body: &str) {
        let path = root.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, body).unwrap();
    }

    /// A go-state receipt the reader accepts, in `<folder>/backups/<stamp>/`.
    fn receipt(root: &Path, folder: &str, stamp: &str, state: &str) {
        let body = json!({
            "schema_version": 1, "project": root.to_string_lossy(), "backup": stamp,
            "plan_digest": "sha256:x", "permission_reference": "r", "state": state,
            "moves": [],
        });
        put(
            root,
            &format!("{folder}/backups/{stamp}/{}", state_root::GO_STATE_RECEIPT),
            &body.to_string(),
        );
    }

    fn only(items: &[Item], path: &str) -> Item {
        items.iter().find(|i| i.path == path).cloned().unwrap()
    }

    /// The listing digest covers sorted names, kinds and file sizes of the first level only: it is
    /// stable, and a new or renamed entry or another file size changes it, but other bytes of the
    /// same size and anything below the first level do not.
    #[test]
    fn the_listing_digest_is_stable_and_follows_names_kinds_and_sizes() {
        let tmp = project();
        let dir = tmp.path().join("d");
        fs::create_dir(&dir).unwrap();
        fs::write(dir.join("a.txt"), "one").unwrap();
        let first = listing_digest(&dir).unwrap();
        assert_eq!(listing_digest(&dir).unwrap(), first);
        assert!(first.starts_with("sha256:") && first.len() == "sha256:".len() + 64);

        fs::write(dir.join("a.txt"), "ONE").unwrap();
        assert_eq!(
            listing_digest(&dir).unwrap(),
            first,
            "same size, other bytes"
        );
        fs::create_dir(dir.join("sub")).unwrap();
        let with_sub = listing_digest(&dir).unwrap();
        assert_ne!(with_sub, first, "a new top-level folder");
        fs::write(dir.join("sub/deep.txt"), "x").unwrap();
        assert_eq!(
            listing_digest(&dir).unwrap(),
            with_sub,
            "below the first level"
        );

        fs::write(dir.join("a.txt"), "three").unwrap();
        let sized = listing_digest(&dir).unwrap();
        assert_ne!(sized, with_sub, "another size");
        fs::rename(dir.join("a.txt"), dir.join("b.txt")).unwrap();
        assert_ne!(listing_digest(&dir).unwrap(), sized, "another name");
        // The same name as a folder reads as another kind.
        fs::remove_file(dir.join("b.txt")).unwrap();
        let without = listing_digest(&dir).unwrap();
        fs::create_dir(dir.join("b.txt")).unwrap();
        assert_ne!(listing_digest(&dir).unwrap(), without);
    }

    #[cfg(unix)]
    #[test]
    fn the_listing_digest_never_follows_a_link_and_refuses_a_link_as_the_folder() {
        let tmp = project();
        let real = tmp.path().join("real");
        fs::create_dir(&real).unwrap();
        fs::write(real.join("a"), "x").unwrap();
        let link = tmp.path().join("link");
        std::os::unix::fs::symlink("real", &link).unwrap();
        assert!(listing_digest(&link).is_err());
        assert!(listing_digest(&tmp.path().join("missing")).is_err());
        // A link inside is an entry of kind `link`, whatever it points at.
        let before = listing_digest(&real).unwrap();
        std::os::unix::fs::symlink("a", real.join("l")).unwrap();
        assert_ne!(listing_digest(&real).unwrap(), before);
    }

    /// Only the five kinds are known paths: a set-aside folder is any `<state>.rust-<name>/`, and
    /// nothing else (a stamped Go backup folder's lock included) is.
    #[test]
    fn only_the_five_kinds_are_known_paths() {
        let earlier = format!("{}/", state_root::EARLIER_PROJECT_DIR);
        let aside = format!("{}{STAMP}/", aside_prefix());
        let notes = format!("{}notes/", aside_prefix());
        for (path, kind) in [
            (earlier.as_str(), Kind::Earlier),
            (aside.as_str(), Kind::Aside),
            (notes.as_str(), Kind::Aside),
            (".moai-backups/", Kind::Backups),
            (".moai/.update.lock", Kind::StateLock),
            (".moai-backup/.update.lock", Kind::BackupLock),
        ] {
            assert_eq!(known_path(path), Some(kind), "{path}");
        }
        let bare_earlier = state_root::EARLIER_PROJECT_DIR;
        let bare_aside = aside.trim_end_matches('/').to_string();
        let empty_aside = aside_prefix() + "/";
        let nested_aside = format!("{}a/b/", aside_prefix());
        let stamped_lock = format!(".moai-backup-{STAMP}/.update.lock");
        for path in [
            "",
            "/",
            "CLAUDE.md",
            "src/",
            bare_earlier,
            bare_aside.as_str(),
            empty_aside.as_str(),
            nested_aside.as_str(),
            ".moai-backups",
            ".moai-backup/",
            ".moai/.update.lock/",
            stamped_lock.as_str(),
            "../.moai-backups/",
        ] {
            assert_eq!(known_path(path), None, "{path}");
        }
        assert!(Kind::Earlier.is_folder() && !Kind::BackupLock.is_folder());
    }

    #[test]
    fn a_receipt_state_word_is_classified_and_judged() {
        assert_eq!(receipt_state("pending"), ReceiptState::Pending);
        assert_eq!(receipt_state("committed"), ReceiptState::Committed);
        assert_eq!(receipt_state("reverted"), ReceiptState::Reverted);
        for other in ["", "Committed", "done", "reverted "] {
            assert_eq!(receipt_state(other), ReceiptState::Unknown, "{other:?}");
        }
        // A reverted receipt never holds a folder; an unfinished or unknown one always does.
        for set_aside in [false, true] {
            assert!(receipt_verdict(ReceiptState::Reverted, set_aside, "s").is_ok());
            let pending = receipt_verdict(ReceiptState::Pending, set_aside, "s").unwrap_err();
            assert!(pending.contains("pending"), "{pending}");
            let unknown = receipt_verdict(ReceiptState::Unknown, set_aside, "s").unwrap_err();
            assert!(unknown.contains("does not know"), "{unknown}");
        }
        // A committed receipt is a finished move in the earlier state folder and a stopped revert
        // in a set-aside folder.
        assert!(receipt_verdict(ReceiptState::Committed, false, "s").is_ok());
        let stopped = receipt_verdict(ReceiptState::Committed, true, "s").unwrap_err();
        assert!(stopped.contains("committed"), "{stopped}");
    }

    /// REQ-3 to REQ-6 on a real folder tree: what can be archived and what only kept.
    #[test]
    fn folders_are_archivable_or_keep_only_by_the_rules_of_their_kind() {
        let tmp = project();
        let root = tmp.path();
        let earlier = format!("{}/", state_root::EARLIER_PROJECT_DIR);
        receipt(
            root,
            state_root::EARLIER_PROJECT_DIR,
            "20261102T000000Z",
            "committed",
        );
        let good = format!("{}{STAMP}", aside_prefix());
        receipt(root, &good, "20261103T000000Z", "reverted");
        let stopped = format!("{}20261105T000000Z", aside_prefix());
        receipt(root, &stopped, "20261103T000000Z", "committed");
        let notes = format!("{}notes", aside_prefix());
        put(root, &format!("{notes}/todo.txt"), "mine");
        put(root, ".moai-backups/sub/.git/HEAD", "ref");
        let items = survey(root, "rust");

        assert!(only(&items, &earlier).archive.is_ok());
        assert!(only(&items, &format!("{good}/")).archive.is_ok());
        let reason = only(&items, &format!("{stopped}/")).archive.unwrap_err();
        assert!(reason.contains("committed"), "{reason}");
        let reason = only(&items, &format!("{notes}/")).archive.unwrap_err();
        assert!(reason.contains("stamp"), "{reason}");
        let reason = only(&items, ".moai-backups/").archive.unwrap_err();
        assert!(reason.contains("checkout"), "{reason}");
        for item in &items {
            assert!(item.sha256.is_some(), "{}", item.path);
        }
        // Every path of the five is listed once, folders before locks.
        assert_eq!(items.len(), 5);

        // A pending receipt, an unreadable one and one in an unknown state hold the earlier folder.
        for (body, keyword) in [("pending", "pending"), ("weird", "does not know")] {
            receipt(
                root,
                state_root::EARLIER_PROJECT_DIR,
                "20261102T000000Z",
                body,
            );
            let reason = only(&survey(root, "rust"), &earlier).archive.unwrap_err();
            assert!(reason.contains(keyword), "{body}: {reason}");
        }
        put(
            root,
            &format!(
                "{}/backups/20261102T000000Z/{}",
                state_root::EARLIER_PROJECT_DIR,
                state_root::GO_STATE_RECEIPT
            ),
            "not json",
        );
        let reason = only(&survey(root, "rust"), &earlier).archive.unwrap_err();
        assert!(reason.contains("cannot be read"), "{reason}");
    }

    /// Only a project whose state generation is exactly `rust` has any of the five.
    #[test]
    fn no_other_generation_lists_the_five_kinds() {
        let tmp = project();
        put(tmp.path(), ".moai-backups/a", "x");
        for generation in ["go", "mixed", "none", "moai-cli"] {
            assert!(survey(tmp.path(), generation).is_empty(), "{generation}");
        }
        assert_eq!(survey(tmp.path(), "rust").len(), 1);
    }

    /// The view carries an `action` for an item that can move, a `keep_only` reason for one that
    /// cannot, a `keep` for both and neither `action` nor `keep_only` while another step waits;
    /// a record of the current source turns `kept` on and an old one does not.
    #[test]
    fn the_view_prints_an_action_or_a_reason_and_always_a_keep() {
        let tmp = project();
        let root = tmp.path();
        put(root, ".moai-backups/a.txt", "x");
        put(root, &format!("{}notes/t", aside_prefix()), "x");
        let shown = root.to_str().unwrap();
        let idle = view(root, "rust", false, &Record::new());
        let key = |p: &str| format!("{}{p}", declined::LEFTOVER_PREFIX);
        assert_eq!(idle.len(), 2);
        let backups = idle.iter().find(|l| l["path"] == ".moai-backups/").unwrap();
        assert_eq!(backups["kept"], json!(false));
        assert_eq!(backups["action"][1], json!(APPLY_OPERATION));
        assert_eq!(backups["action"][2], json!(shown));
        assert_eq!(backups["action"][7], json!("--item"));
        assert_eq!(backups["action"][8], json!(".moai-backups/"));
        assert!(backups.get("keep_only").is_none() && backups.get("ask_again").is_none());
        assert_eq!(backups["keep"][4], json!(key(".moai-backups/")));
        let notes_path = format!("{}notes/", aside_prefix());
        let notes = idle
            .iter()
            .find(|l| l["path"] == notes_path.as_str())
            .unwrap();
        assert!(notes.get("action").is_none());
        assert!(notes["keep_only"].as_str().unwrap().contains("stamp"));
        assert!(notes["keep"].is_array());
        // The action's plan digest is the one `plan_digest` computes over the archivable items.
        let digest = plan_digest(&survey(root, "rust"));
        assert_eq!(backups["action"][4], json!(digest));

        // While another step waits: the keep stays, no action, no reason.
        let waiting = view(root, "rust", true, &Record::new());
        let backups = waiting
            .iter()
            .find(|l| l["path"] == ".moai-backups/")
            .unwrap();
        assert!(backups.get("action").is_none() && backups.get("keep_only").is_none());
        assert!(backups["action_waits"].is_string() && backups["keep"].is_array());

        // A record of the current source keeps the item; a record of another source does not.
        let sha = source_sha256_of(root, ".moai-backups/");
        let kept = view(
            root,
            "rust",
            false,
            &Record::from([(key(".moai-backups/"), sha)]),
        );
        let backups = kept.iter().find(|l| l["path"] == ".moai-backups/").unwrap();
        assert_eq!(backups["kept"], json!(true));
        assert!(backups["ask_again"].is_array() && backups["action"].is_array());
        let stale = view(
            root,
            "rust",
            false,
            &Record::from([(key(".moai-backups/"), "sha256:old".to_string())]),
        );
        let backups = stale
            .iter()
            .find(|l| l["path"] == ".moai-backups/")
            .unwrap();
        assert_eq!(backups["kept"], json!(false));
        // One open item is a question; a kept one and a fact without `kept` are not.
        assert_eq!(open(&idle).len(), 2);
        let notes_open = open(&idle)
            .into_iter()
            .find(|o| o.path == notes_path)
            .unwrap();
        assert!(!notes_open.archive);
        assert_eq!(open(&kept).len(), 1);
    }

    fn source_sha256_of(root: &Path, path: &str) -> String {
        survey(root, "rust")
            .into_iter()
            .find(|i| i.path == path)
            .and_then(|i| i.sha256)
            .unwrap()
    }

    #[cfg(unix)]
    #[test]
    fn a_dead_lock_is_listed_with_its_bytes_sha_and_a_live_one_is_not() {
        let tmp = project();
        let root = tmp.path();
        let mut child = std::process::Command::new("true").spawn().unwrap();
        let dead = child.id();
        child.wait().unwrap();
        let body =
            json!({"pid": dead, "started_at": "2026-11-01T00:00:00Z", "hostname": "secret-host"})
                .to_string();
        put(root, ".moai-backup/.update.lock", &body);
        let items = survey(root, "rust");
        let lock = only(&items, ".moai-backup/.update.lock");
        assert_eq!(lock.kind, Kind::BackupLock);
        assert_eq!(
            lock.sha256.as_deref(),
            Some(crate::digest(body.as_bytes()).as_str())
        );
        assert!(lock.archive.is_ok());
        assert!(lock.fact.contains("Go backup folder") && lock.fact.contains("not running"));
        assert!(!lock.fact.contains("secret-host"), "{}", lock.fact);
        let alive = json!({"pid": std::process::id()}).to_string();
        put(root, ".moai-backup/.update.lock", &alive);
        assert!(survey(root, "rust").is_empty());
        put(root, ".moai-backup/.update.lock", &"x".repeat(5000));
        assert!(survey(root, "rust").is_empty());
    }
}
