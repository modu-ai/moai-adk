//! The archive apply and its revert for the leftovers of an earlier generation
//! (SPEC-20261127-update-leftovers-action, Plan steps 5 and 6).
//!
//! The apply moves each chosen leftover (a folder by one rename, a lock file by one hard link and
//! one unlink) to `<state>/backups/<UTC>/leftovers/<same path>` and deletes nothing; the revert
//! puts the items back. The two keep the discipline of the go-residue apply:
//!
//! - nothing is written unless a fresh survey has exactly the plan digest the caller confirmed
//!   and every named item is one the fresh survey lists as archivable, reached without a link;
//! - the pending receipt is saved under the one size limit its reader uses and read back before
//!   the first move, and it is marked `committed` only after every move reads back;
//! - the receipt names each item by path, kind, entry count, size and top-level listing digest
//!   and nothing else: no file's contents, no child name, no lock payload;
//! - the revert trusts the receipt for nothing: every item path passes the path guard first
//!   (one that fails makes the whole revert move nothing), and each item goes back only when its
//!   original place is empty and the archived digest is the receipt's, judged by the disk and not
//!   by the receipt's `moved` flags.
//!
//! A secret-shaped name inside a moved folder is counted and never opened; the move is by name.

use super::leftovers::{self, rel_of, Item, Kind, LOCK_LIMIT};
use crate::go_residue::contained;
use crate::go_state::{
    has_secret_component, make_dirs, make_private_dirs, move_exclusive, parent_of,
    read_bounded_text,
};
use crate::state_root::{self, utc_stamp, GO_STATE_RECEIPT_MAX, MOAI_DIR};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{fs, path::Path};

type Result<T> = std::result::Result<T, String>;

/// The subcommand the printed undo row names.
pub const REVERT_OPERATION: &str = "project-update-leftovers-revert";
/// The receipt's file name under `<state>/backups/<UTC>/`. It is not `go-state-receipt.json`, so
/// the state classifier and the go-state revert never read it.
const RECEIPT_FILE: &str = "leftovers-receipt.json";
/// The folder under the stamp folder that holds the moved items.
const AREA: &str = "leftovers";
/// One size limit for encoding, replacing and reading the receipt.
const RECEIPT_MAX: u64 = GO_STATE_RECEIPT_MAX;
/// A receipt names paths the user may want private.
const PRIVATE_FILE: u32 = 0o600;
/// How many following seconds a stamp may be moved on when its folder is taken already.
const STAMP_TRIES: i64 = 120;
/// The `MOAI_TEST_FAULT` point that stands in for a backup area on another filesystem.
const FAULT_CROSS_DEVICE: &str = "leftovers_cross_device";

/// The refusal for a digest that is not the fresh plan's (the phrase makes
/// `contract::classify` say `stale_plan`).
const STALE: &str =
    "stale or altered plan: what project-update-leftovers-apply would move no longer matches \
     the preview that was confirmed (a leftover changed, appeared or went, or the digest is from \
     an older version); nothing was written; run `moai project-update-plan <project>` again and \
     confirm again with the new plan_digest of the leftovers";

/// One item of the receipt.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
struct Entry {
    /// As the plan prints it: a folder ends in `/`.
    path: String,
    /// `folder` or `file`.
    kind: String,
    /// A folder's top-level entry count (0 for a file).
    entries: u64,
    /// A file's size, or the sum of a folder's top-level file sizes.
    size: u64,
    /// The item's source sha256 of the plan.
    listing_digest: String,
    moved: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
struct Receipt {
    schema_version: u32,
    project: String,
    backup: String,
    plan_digest: String,
    /// The caller's `--permission-reference`, recorded as given (not an authorization).
    permission_reference: String,
    /// `pending` -> `committed` -> `reverted`.
    state: String,
    items: Vec<Entry>,
}

fn receipt_rel(stamp: &str) -> String {
    format!("{MOAI_DIR}/backups/{stamp}/{RECEIPT_FILE}")
}

fn area_rel(stamp: &str) -> String {
    format!("{MOAI_DIR}/backups/{stamp}/{AREA}")
}

/// The receipt's bytes; a receipt over `limit` is never written, because its reader would refuse
/// to read it back.
fn encode(receipt: &Receipt, limit: u64) -> Result<Vec<u8>> {
    let bytes = serde_json::to_vec_pretty(receipt).map_err(|_| "encode failed".to_string())?;
    if bytes.len() as u64 > limit {
        return Err(format!("the receipt is larger than {limit} bytes"));
    }
    Ok(bytes)
}

/// The first stamp, from now on, whose backup folder does not exist yet: two applies in one
/// second keep two receipts.
fn free_stamp(root: &Path) -> Result<String> {
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);
    for i in 0..STAMP_TRIES {
        let stamp = crate::baseline::utc_from_epoch(now + i).replace(['-', ':'], "");
        if fs::symlink_metadata(root.join(format!("{MOAI_DIR}/backups/{stamp}"))).is_err() {
            return Ok(stamp);
        }
    }
    Err(
        "project-update-leftovers-apply refused: no free backup stamp in the next two minutes; \
         nothing was written"
            .into(),
    )
}

/// Move one folder with a single rename: the source must be a real folder (not a link), the
/// destination must not exist, and another filesystem is refused (no copy fallback).
fn rename_folder(from: &Path, to: &Path) -> Result<()> {
    match fs::symlink_metadata(from) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => return Err("folder move refused: the source is not a real folder".into()),
        Err(_) => return Err("folder move refused: the source is missing".into()),
    }
    // A rename onto an empty folder would succeed, so the destination is looked at first.
    if fs::symlink_metadata(to).is_ok() {
        return Err("folder move refused: the destination already exists".into());
    }
    fs::rename(from, to).map_err(|e| {
        if e.raw_os_error() == Some(18) {
            "cross-filesystem move refused; the project and its backup must share one filesystem"
                .to_string()
        } else {
            "folder move failed (rename)".to_string()
        }
    })
}

/// The sha256 of a lock file's bytes: a regular file of at most the lock limit, never a link.
fn file_digest(path: &Path) -> Option<String> {
    let meta = fs::symlink_metadata(path).ok()?;
    if !meta.is_file() || meta.len() > LOCK_LIMIT {
        return None;
    }
    let bytes = fs::read(path).ok()?;
    (bytes.len() as u64 <= LOCK_LIMIT).then(|| crate::digest(&bytes))
}

/// The source sha256 `path` has now: the top-level listing digest of a folder, the bytes' digest
/// of a lock. Neither opens a name inside a folder.
fn digest_of(path: &Path, kind: Kind) -> Option<String> {
    if kind.is_folder() {
        leftovers::listing_digest(path).ok()
    } else {
        file_digest(path)
    }
}

/// What one more look at an item says right before it is written to or moved: the folder's
/// listing is still the plan's, the lock's bytes are still the plan's and its process is still
/// not running (the survey judged that once; a process may have started since).
fn still_as_planned(root: &Path, item: &Item) -> Result<()> {
    let rel = rel_of(&item.path);
    if !contained(root, rel) {
        return Err(format!(
            "{} is absolute, holds a `..` component or passes through a link",
            item.path
        ));
    }
    let planned = item.sha256.as_deref().unwrap_or_default();
    if item.kind.is_folder() {
        let now = leftovers::listing_digest(&root.join(rel)).ok();
        if now.as_deref() != Some(planned) {
            return Err(format!("{} changed after the preview", item.path));
        }
        return Ok(());
    }
    match leftovers::dead_update_lock(root, rel, "the lock") {
        Some((_, sha)) if sha == planned => Ok(()),
        _ => Err(format!(
            "{} is no longer a lock of a process that is not running, or its bytes changed after \
             the preview",
            item.path
        )),
    }
}

/// What one pass over a folder's top level counts: the entries, the sum of the file sizes and the
/// secret-shaped file names. Only names and `lstat` data are read; nothing is opened.
struct Counted {
    entries: u64,
    size: u64,
    secret_shaped: u64,
}

fn count_item(root: &Path, item: &Item) -> Counted {
    let path = root.join(rel_of(&item.path));
    if !item.kind.is_folder() {
        let size = fs::symlink_metadata(&path).map(|m| m.len()).unwrap_or(0);
        return Counted {
            entries: 0,
            size,
            secret_shaped: 0,
        };
    }
    let mut out = Counted {
        entries: 0,
        size: 0,
        secret_shaped: 0,
    };
    let Ok(listing) = fs::read_dir(&path) else {
        return out;
    };
    for entry in listing.flatten() {
        out.entries += 1;
        let Ok(meta) = entry.metadata() else {
            continue;
        };
        if meta.is_file() {
            out.size += meta.len();
        }
        if !meta.is_dir() && has_secret_component(&entry.file_name().to_string_lossy()) {
            out.secret_shaped += 1;
        }
    }
    out
}

/// How many files git tracks under `rels` (the index is read; no file is). Zero when the project
/// is not a git repository or git cannot answer.
fn tracked_count(root: &Path, rels: &[&str]) -> u64 {
    let mut cmd = std::process::Command::new("git");
    cmd.args(["--no-optional-locks", "--literal-pathspecs", "-C"])
        .arg(root)
        .args(["ls-files", "-z", "--"])
        .args(rels);
    crate::identity::scrub_git_env(&mut cmd);
    cmd.stdin(std::process::Stdio::null())
        .stderr(std::process::Stdio::null());
    match cmd.output() {
        Ok(out) if out.status.success() => out
            .stdout
            .split(|b| *b == 0)
            .filter(|p| !p.is_empty())
            .count() as u64,
        _ => 0,
    }
}

/// Every check that needs no write, in the order that costs least: the path guard of each item
/// and of its place in the archive, one more look at each item (the lock's process included),
/// then the filesystem of the backup area. The fault point `leftovers_cross_device` stands in
/// for a backup area on another filesystem; the real device compare is part of the fresh
/// survey's verdict that chose these items.
fn preflight(root: &Path, stamp: &str, chosen: &[Item]) -> Result<()> {
    let area = area_rel(stamp);
    for item in chosen {
        still_as_planned(root, item)?;
        if !contained(root, &format!("{area}/{}", rel_of(&item.path))) {
            return Err(format!(
                "the backup place of {} passes through a link",
                item.path
            ));
        }
    }
    crate::update::test_fault(FAULT_CROSS_DEVICE).map_err(|_| {
        format!(
            "simulated cross-device check (test fault {FAULT_CROSS_DEVICE}): the backup area is \
             taken to be on another filesystem than the items"
        )
    })
}

fn refused(what: &str) -> String {
    format!("project-update-leftovers-apply refused: {what}; nothing was written")
}

/// Archive the chosen leftovers of the preview whose digest the caller confirmed. Nothing is
/// written unless a fresh survey has exactly that digest and every item is archivable.
pub fn apply(
    project: &Path,
    plan_digest_given: &str,
    permission_reference: &str,
    picks: &[String],
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("a permission reference is required".into());
    }
    if picks.is_empty() {
        return Err(refused("name at least one `--item <path>`"));
    }
    for (at, pick) in picks.iter().enumerate() {
        let shown = crate::card::input::echo(pick);
        if leftovers::known_path(pick).is_none() {
            return Err(refused(&format!(
                "{shown} is not one of the leftover paths a plan prints (it may be absolute, hold \
                 `..` or name another place)"
            )));
        }
        if picks[..at].contains(pick) {
            return Err(refused(&format!("{shown} is named twice")));
        }
    }
    let root = fs::canonicalize(project)
        .map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    for pick in picks {
        if !contained(&root, rel_of(pick)) {
            return Err(refused(&format!(
                "{} passes through a link",
                crate::card::input::echo(pick)
            )));
        }
    }
    let generation = state_root::classify(&root).generation;
    let items = leftovers::survey(&root, generation);
    if leftovers::plan_digest(&items) != plan_digest_given {
        return Err(STALE.into());
    }
    for pick in picks {
        match items.iter().find(|i| &i.path == pick) {
            None => {
                return Err(refused(&format!(
                    "{} is not listed by the plan now",
                    crate::card::input::echo(pick)
                )))
            }
            Some(i) if i.archive.is_err() || i.sha256.is_none() => {
                return Err(refused(&format!(
                    "{} cannot be archived: {}",
                    i.path,
                    i.archive.clone().err().unwrap_or_default()
                )))
            }
            Some(_) => {}
        }
    }
    // The survey's order, so that the receipt and the result do not depend on the argv order.
    let chosen: Vec<Item> = items
        .into_iter()
        .filter(|i| picks.contains(&i.path))
        .collect();
    let stamp = free_stamp(&root)?;
    apply_items(
        &root,
        &stamp,
        &chosen,
        plan_digest_given,
        permission_reference,
    )
}

/// The apply of an already chosen list under `stamp`: the checks that need no write, the pending
/// receipt saved and read back, one move per item (each looked at once more right before), the
/// read-back of every moved item, and the receipt `committed`.
fn apply_items(
    root: &Path,
    stamp: &str,
    chosen: &[Item],
    digest: &str,
    permission: &str,
) -> Result<Value> {
    preflight(root, stamp, chosen).map_err(|e| refused(&e))?;
    let counted: Vec<Counted> = chosen.iter().map(|i| count_item(root, i)).collect();
    let rels: Vec<&str> = chosen.iter().map(|i| rel_of(&i.path)).collect();
    let tracked_moved = tracked_count(root, &rels);
    let secret_shaped: u64 = counted.iter().map(|c| c.secret_shaped).sum();

    let area = area_rel(stamp);
    let mut receipt = Receipt {
        schema_version: 1,
        project: root.to_string_lossy().into_owned(),
        backup: stamp.to_string(),
        plan_digest: digest.to_string(),
        permission_reference: permission.to_string(),
        state: "pending".into(),
        items: chosen
            .iter()
            .zip(&counted)
            .map(|(i, c)| Entry {
                path: i.path.clone(),
                kind: if i.kind.is_folder() { "folder" } else { "file" }.into(),
                entries: c.entries,
                size: c.size,
                listing_digest: i.sha256.clone().unwrap_or_default(),
                moved: false,
            })
            .collect(),
    };
    let pending = encode(&receipt, RECEIPT_MAX).map_err(|e| refused(&e))?;
    let receipt_path = root.join(receipt_rel(stamp));
    make_private_dirs(root, &format!("{MOAI_DIR}/backups/{stamp}")).map_err(|e| {
        refused(&format!(
            "the backup area is not usable ({e}); nothing was moved"
        ))
    })?;
    crate::apply::publish_exclusive_mode(&receipt_path, &pending, Some(PRIVATE_FILE))
        .map_err(|e| refused(&format!("the pending receipt could not be saved ({e})")))?;
    if fs::read(&receipt_path).ok().as_deref() != Some(pending.as_slice()) {
        return Err(refused("the pending receipt does not read back"));
    }

    let revert_hint = format!(
        "run `moai {REVERT_OPERATION} <project> --receipt {stamp} --permission-reference <ref>` to \
         put the moved items back"
    );
    for (moved, item) in chosen.iter().enumerate() {
        let to_rel = format!("{area}/{}", rel_of(&item.path));
        let step = (|| -> Result<()> {
            still_as_planned(root, item)?;
            make_private_dirs(root, &parent_of(&to_rel)?)?;
            let (from, to) = (root.join(rel_of(&item.path)), root.join(&to_rel));
            if item.kind.is_folder() {
                rename_folder(&from, &to)
            } else {
                move_exclusive(&from, &to)
            }
        })();
        let Err(e) = step else { continue };
        // Record what did move, so the revert and a reader see it; best effort.
        for entry in receipt.items.iter_mut().take(moved) {
            entry.moved = true;
        }
        if moved > 0 {
            if let Ok(after) = encode(&receipt, RECEIPT_MAX) {
                let _ = crate::apply::publish_replace_limited(
                    &receipt_path,
                    Some(&pending),
                    &after,
                    RECEIPT_MAX,
                );
            }
            return Err(format!(
                "partial apply: {moved} of {} items moved; stopped at {}: {e}; the receipt stays \
                 pending; {revert_hint}",
                chosen.len(),
                item.path
            ));
        }
        return Err(format!(
            "project-update-leftovers-apply refused: {}: {e}; no item moved; the pending receipt \
             stays; {revert_hint}",
            item.path
        ));
    }
    // Read every moved item back before the receipt says `committed`.
    for item in chosen {
        let archived = root.join(format!("{area}/{}", rel_of(&item.path)));
        let back = digest_of(&archived, item.kind);
        let gone = fs::symlink_metadata(root.join(rel_of(&item.path))).is_err();
        if back != item.sha256 || !gone {
            return Err(format!(
                "partial apply: {} items moved but {} does not read back from the backup; the \
                 receipt stays pending; {revert_hint}",
                chosen.len(),
                item.path
            ));
        }
    }
    receipt.state = "committed".into();
    for entry in &mut receipt.items {
        entry.moved = true;
    }
    let committed = encode(&receipt, RECEIPT_MAX)?;
    crate::apply::publish_replace_limited(&receipt_path, Some(&pending), &committed, RECEIPT_MAX)
        .map_err(|e| {
        format!(
            "partial apply: {} items moved but the receipt could not be marked committed \
                 ({e}); {revert_hint}",
            chosen.len()
        )
    })?;
    Ok(json!({
        "state": "committed",
        "receipt": stamp,
        "backup": area,
        "plan_digest": digest,
        "items": chosen.iter().map(|i| json!({"path": i.path, "outcome": "moved"})).collect::<Vec<_>>(),
        "tracked_moved": tracked_moved,
        "warning": (tracked_moved > 0).then(|| format!(
            "{tracked_moved} of the moved files are tracked by git; a checkout brings them back \
             unless the deletions are committed"
        )),
        "secret_shaped": secret_shaped,
        "revert": {"argv": [
            "moai", REVERT_OPERATION, root.to_string_lossy(), "--receipt", stamp,
            "--permission-reference", super::PERMISSION_PLACEHOLDER,
        ]},
    }))
}

/// Whether one receipt item passes the path guard: one of the five leftover paths, of the kind
/// the receipt says, reached (here and in the archive) without a link, with a sha256 digest.
fn entry_guard_ok(root: &Path, area: &str, entry: &Entry) -> bool {
    let Some(kind) = leftovers::known_path(&entry.path) else {
        return false;
    };
    let kind_word = if kind.is_folder() { "folder" } else { "file" };
    let hex_ok = entry
        .listing_digest
        .strip_prefix("sha256:")
        .is_some_and(|h| h.len() == 64 && h.bytes().all(|b| b.is_ascii_hexdigit()));
    let rel = rel_of(&entry.path);
    entry.kind == kind_word
        && hex_ok
        && contained(root, rel)
        && contained(root, &format!("{area}/{rel}"))
}

/// Put the items of one apply back. The receipt is untrusted: every item path passes the guard
/// first (when one fails nothing moves and the whole revert is a `conflict`), and each item goes
/// back only when its original place is empty and the archived digest is the receipt's; anything
/// else is left as it is and reported.
pub fn revert(project: &Path, stamp: &str, permission_reference: &str) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("a permission reference is required".into());
    }
    if !utc_stamp(stamp) {
        return Err("invalid receipt name: expected the apply's UTC stamp YYYYMMDDTHHMMSSZ".into());
    }
    let root = fs::canonicalize(project)
        .map_err(|_| crate::contract::PROJECT_PATH_UNAVAILABLE.to_string())?;
    let rel = receipt_rel(stamp);
    let text = read_bounded_text(&root, &rel, RECEIPT_MAX)?
        .ok_or_else(|| format!("{rel}: no leftovers receipt"))?;
    let mut receipt: Receipt =
        serde_json::from_str(&text).map_err(|_| format!("{rel}: unreadable receipt"))?;
    if receipt.backup != stamp || receipt.project != root.to_string_lossy() {
        return Err("receipt scope mismatch; refused".into());
    }
    if !matches!(receipt.state.as_str(), "pending" | "committed" | "reverted") {
        return Err("receipt state is not one an apply writes; refused".into());
    }
    if receipt.state == "reverted" {
        return Ok(json!({"state": "reverted", "receipt": stamp, "items": [],
                         "note": "already reverted"}));
    }
    let area = area_rel(stamp);
    let mut seen: Vec<&str> = Vec::new();
    let mut guard: Vec<bool> = Vec::new();
    for entry in &receipt.items {
        let fresh = !seen.contains(&entry.path.as_str());
        seen.push(&entry.path);
        guard.push(fresh && entry_guard_ok(&root, &area, entry));
    }
    if guard.iter().any(|ok| !ok) {
        let items: Vec<Value> = receipt
            .items
            .iter()
            .zip(&guard)
            .map(|(e, ok)| {
                let reason = if *ok {
                    "not touched: another item of the receipt fails the path guard"
                } else {
                    "its path is not one the apply writes, is named twice, or passes through a link"
                };
                json!({"path": e.path, "outcome": "conflict", "reason": reason})
            })
            .collect();
        return Ok(
            json!({"state": "conflict", "receipt": stamp, "items": items,
                         "note": "nothing was moved"}),
        );
    }

    let mut outcomes: Vec<Value> = Vec::new();
    let mut clean = true;
    let mut restored = 0usize;
    for entry in &receipt.items {
        let Some(kind) = leftovers::known_path(&entry.path) else {
            continue;
        };
        let item_rel = rel_of(&entry.path);
        let original = root.join(item_rel);
        let backup = root.join(format!("{area}/{item_rel}"));
        let original_there = fs::symlink_metadata(&original).is_ok();
        let backup_there = fs::symlink_metadata(&backup).is_ok();
        let digest_is =
            |path: &Path| digest_of(path, kind).as_deref() == Some(entry.listing_digest.as_str());
        let (outcome, detail) = if !backup_there {
            // Never moved (the original is still there as the receipt recorded it) or lost
            // (neither is there). Another thing at the original place is a conflict.
            if !original_there {
                (
                    "missing",
                    "neither the archived copy nor the original place has it",
                )
            } else if digest_is(&original) {
                (
                    "restored",
                    "it is at its original place as the receipt recorded it",
                )
            } else {
                (
                    "conflict",
                    "the original place holds something other than the receipt records",
                )
            }
        } else if original_there {
            ("conflict", "the original place is taken")
        } else if !digest_is(&backup) {
            ("conflict", "the archived copy differs from the receipt")
        } else {
            let put_back = (|| -> Result<()> {
                make_dirs(&root, &parent_of(item_rel)?)?;
                if !digest_is(&backup) {
                    return Err(format!(
                        "{} changed in the archive; move stopped",
                        entry.path
                    ));
                }
                if kind.is_folder() {
                    rename_folder(&backup, &original)
                } else {
                    move_exclusive(&backup, &original)
                }
            })();
            if let Err(e) = put_back {
                return Err(format!(
                    "partial apply: the revert restored {restored} items and stopped at {}: {e}; \
                     run the same revert again to continue",
                    entry.path
                ));
            }
            restored += 1;
            ("restored", "put back from the archive")
        };
        clean &= outcome == "restored";
        outcomes.push(json!({"path": entry.path, "outcome": outcome, "detail": detail}));
    }
    if clean {
        let before = fs::read(root.join(&rel)).map_err(|_| "receipt read failed".to_string())?;
        receipt.state = "reverted".into();
        for entry in &mut receipt.items {
            entry.moved = false;
        }
        let after = encode(&receipt, RECEIPT_MAX)?;
        crate::apply::publish_replace_limited(&root.join(&rel), Some(&before), &after, RECEIPT_MAX)
            .map_err(|e| {
                format!("partial apply: the items are back but the receipt could not be marked reverted ({e})")
            })?;
    }
    Ok(json!({
        "state": if clean { "reverted" } else { "conflict" },
        "receipt": stamp,
        "items": outcomes,
    }))
}

/// Whether a leftovers receipt that is not `reverted` exists under the state folder's backups:
/// the printed undo order then begins with the leftovers revert (REQ-22). A receipt that cannot
/// be read counts as not reverted; nothing is written.
pub fn open_receipt_exists(root: &Path) -> bool {
    let backups = root.join(MOAI_DIR).join("backups");
    if !fs::symlink_metadata(&backups).is_ok_and(|m| m.is_dir()) {
        return false;
    }
    let Ok(listing) = fs::read_dir(&backups) else {
        return false;
    };
    for entry in listing.flatten() {
        let Some(name) = entry.file_name().to_str().map(str::to_string) else {
            continue;
        };
        if !utc_stamp(&name) {
            continue;
        }
        match read_bounded_text(root, &receipt_rel(&name), RECEIPT_MAX) {
            Ok(None) => {}
            Ok(Some(text)) => {
                let state = serde_json::from_str::<Value>(&text)
                    .ok()
                    .and_then(|v| v["state"].as_str().map(str::to_string));
                if state.as_deref() != Some("reverted") {
                    return true;
                }
            }
            Err(_) => return true,
        }
    }
    false
}

#[cfg(test)]
mod tests {
    use super::*;

    const STAMP: &str = "20261101T000000Z";

    fn project() -> (tempfile::TempDir, std::path::PathBuf) {
        let tmp = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        (tmp, root)
    }

    /// The pid of a process that has finished: no live process has it now.
    fn dead_pid() -> u32 {
        let mut child = std::process::Command::new("true").spawn().unwrap();
        let pid = child.id();
        child.wait().unwrap();
        pid
    }

    fn lock_item(root: &Path, pid: u32) -> Item {
        let rel = format!("{}/.update.lock", state_root::GO_BACKUP);
        let body =
            json!({"pid": pid, "started_at": "2026-11-01T00:00:00Z", "hostname": "h"}).to_string();
        fs::create_dir_all(root.join(&rel).parent().unwrap()).unwrap();
        fs::write(root.join(&rel), &body).unwrap();
        Item {
            path: rel,
            kind: Kind::BackupLock,
            fact: String::new(),
            sha256: Some(crate::digest(body.as_bytes())),
            archive: Ok(()),
        }
    }

    /// REQ-7, REQ-14 (AC-6): a lock whose bytes did not change since the plan but whose process
    /// is running when the apply starts is refused before any write; the same lock with a dead
    /// process moves. This is the apply's own check, apart from the survey that chose the item.
    #[test]
    fn a_lock_whose_process_is_alive_when_the_apply_starts_is_refused() {
        let (_tmp, root) = project();
        let live = lock_item(&root, std::process::id());
        let err = apply_items(&root, STAMP, std::slice::from_ref(&live), "d", "ref").unwrap_err();
        assert!(err.contains("not running"), "{err}");
        assert!(root.join(&live.path).is_file(), "the lock stays");
        assert!(
            fs::symlink_metadata(root.join(MOAI_DIR)).is_err(),
            "nothing was written: {err}"
        );

        let dead = lock_item(&root, dead_pid());
        let out = apply_items(&root, STAMP, std::slice::from_ref(&dead), "d", "ref").unwrap();
        assert_eq!(out["state"], json!("committed"), "{out}");
        assert!(fs::symlink_metadata(root.join(&dead.path)).is_err());
        let archived = root.join(format!("{}/{}", area_rel(STAMP), dead.path));
        assert!(archived.is_file());
    }

    /// The receipt keeps one size limit: a receipt of exactly the limit is encoded, one byte
    /// more is refused (N and N+1).
    #[test]
    fn the_receipt_is_encoded_up_to_its_limit_and_refused_above_it() {
        let receipt = Receipt {
            schema_version: 1,
            project: "p".into(),
            backup: STAMP.into(),
            plan_digest: "d".into(),
            permission_reference: "r".into(),
            state: "pending".into(),
            items: vec![Entry {
                path: ".moai-backups/".into(),
                kind: "folder".into(),
                entries: 1,
                size: 2,
                listing_digest: "sha256:x".into(),
                moved: false,
            }],
        };
        let n = serde_json::to_vec_pretty(&receipt).unwrap().len() as u64;
        assert_eq!(encode(&receipt, n).unwrap().len() as u64, n);
        let err = encode(&receipt, n - 1).unwrap_err();
        assert!(err.contains("larger than"), "{err}");
        // The limit the reader uses is the one the writer keeps.
        assert_eq!(RECEIPT_MAX, GO_STATE_RECEIPT_MAX);
    }

    /// REQ-12: a folder moves by one rename and never onto a place that is taken, not even an
    /// empty folder (which a bare rename would replace); a file or a missing source is refused.
    #[test]
    fn rename_folder_refuses_an_occupied_target() {
        let (_tmp, root) = project();
        let from = root.join("from");
        fs::create_dir_all(&from).unwrap();
        fs::write(from.join("a.txt"), "a").unwrap();
        let empty = root.join("empty");
        fs::create_dir_all(&empty).unwrap();
        let full = root.join("full");
        fs::create_dir_all(&full).unwrap();
        fs::write(full.join("b.txt"), "b").unwrap();
        for target in [&empty, &full] {
            let err = rename_folder(&from, target).unwrap_err();
            assert!(err.contains("already exists"), "{err}");
            assert_eq!(fs::read(from.join("a.txt")).unwrap(), b"a");
        }
        assert!(rename_folder(&root.join("missing"), &root.join("x")).is_err());
        assert!(rename_folder(&from.join("a.txt"), &root.join("x")).is_err());
        rename_folder(&from, &root.join("moved")).unwrap();
        assert_eq!(fs::read(root.join("moved/a.txt")).unwrap(), b"a");
        assert!(fs::symlink_metadata(&from).is_err());
    }
}
