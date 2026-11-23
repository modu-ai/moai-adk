//! SPEC-20261123-home-state-retire: archive the earlier HOME state folder whole. The earlier
//! folder is renamed, by one rename on one file system, to `~/.moai/archive/moai-cli-<UTC>/`.
//! Nothing inside the earlier folder is copied, deleted or opened: the preview reads its names,
//! kinds and modes by `lstat` alone, so a vault inside keeps its owner-only modes and its bytes,
//! and the revert is the same rename backwards. No command other than this one reads the
//! earlier folder (user decision 6504).
//!
//! The apply writes two files of its own into `~/.moai/archive/` before the rename, in this
//! order: the manifest (`state-retire-manifest-<UTC>.json`), then the move journal
//! ([`crate::move_journal`], which publishes it and moves). A crash between the two leaves a
//! manifest and no journal; the preview and the apply of the same clock take that manifest as
//! their own (it names this kind, HOME and stamp) and replace it by one rename, so the rerun
//! finishes. Another file at that name keeps the clock `stamp_taken`.
//!
//! The step stops while a program uses a path under the folder, while the running binary lies
//! inside it, and where the folder, `~/.moai`, its `archive` folder or the archive name is not
//! what the move needs. Shell files are never read or written: the preview only says which
//! `PATH` entries of the calling process name the earlier folder.

use crate::home_state::{stamp, InUse, CANDIDATE_LIMIT};
use crate::move_journal::{self, MoveEnv, PlannedMove};
use crate::state_root::MOAI_DIR;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::ffi::OsString;
use std::fs;
use std::path::{Path, PathBuf};

/// The folder under `~/.moai` that holds the archived earlier folders.
pub const ARCHIVE_DIR: &str = "archive";
/// The journal `kind` of this step.
pub const KIND: &str = "state-retire";
const MANIFEST_PREFIX: &str = "state-retire-manifest-";
const ARCHIVE_PREFIX: &str = "moai-cli-";
/// Test seam of the apply: stop at a step like a crash would (debug builds only): `after_manifest`
/// (this module's, between the manifest and the journal) or a point of [`crate::move_journal`].
pub const FAULT_VAR: &str = "MOAI_CLI_TEST_STATE_RETIRE_FAULT";

/// What the calling process knows that the library does not read itself.
pub struct Facts {
    /// The running binary, when the system says where it is.
    pub exe: Option<PathBuf>,
    /// The `PATH` variable.
    pub path: Option<OsString>,
}

impl Facts {
    pub fn current() -> Facts {
        Facts {
            exe: std::env::current_exe().ok(),
            path: std::env::var_os("PATH"),
        }
    }
}

/// Which digest an apply is confirmed by.
#[derive(Clone, Copy, PartialEq, Eq)]
pub enum Confirm {
    /// `plan_digest`: bound to the clock of the preview (the archive name).
    Plan,
    /// `content_digest`: the same preview without the clock.
    Content,
}

impl Confirm {
    fn key(self) -> &'static str {
        match self {
            Confirm::Plan => "plan_digest",
            Confirm::Content => "content_digest",
        }
    }
}

/// The earlier state folder. This module is the one place that names it on purpose.
fn earlier(home: &Path) -> PathBuf {
    home.join(crate::HOME_STATE_DIR)
}

fn canon(p: &Path) -> PathBuf {
    fs::canonicalize(p).unwrap_or_else(|_| p.to_path_buf())
}

fn kind_of(p: &Path) -> &'static str {
    crate::home_retire::kind_of(p)
}

#[cfg(unix)]
fn mode_text(m: &fs::Metadata) -> String {
    use std::os::unix::fs::PermissionsExt;
    format!("{:o}", m.permissions().mode() & 0o7777)
}

#[cfg(not(unix))]
fn mode_text(_: &fs::Metadata) -> String {
    String::new()
}

/// The names of the archive of `stamp`: the root, the archived folder, the journal, the manifest.
struct Names {
    root: PathBuf,
    folder: PathBuf,
    journal: PathBuf,
    manifest: PathBuf,
}

fn names(home: &Path, stamp: &str) -> Names {
    let root = home.join(MOAI_DIR).join(ARCHIVE_DIR);
    Names {
        folder: root.join(format!("{ARCHIVE_PREFIX}{stamp}")),
        journal: move_journal::journal_path(&root, stamp),
        manifest: root.join(format!("{MANIFEST_PREFIX}{stamp}.json")),
        root,
    }
}

/// The most a manifest is read for: 500 listed entries stay far below it.
const MANIFEST_READ_MAX: u64 = 1 << 20;

/// True when the file at the manifest name is the one this step wrote for this HOME and clock:
/// a regular file that parses and names this kind, HOME and stamp. A crash between the manifest
/// and the journal leaves exactly this file, so a rerun of the same clock takes it as its own
/// instead of calling the name taken. Anything else at that name is not ours.
fn own_manifest(n: &Names, home: &Path, stamp: &str) -> bool {
    let Ok(m) = fs::symlink_metadata(&n.manifest) else {
        return false;
    };
    if !m.is_file() || m.len() > MANIFEST_READ_MAX {
        return false;
    }
    let Some(v) = fs::read(&n.manifest)
        .ok()
        .and_then(|b| serde_json::from_slice::<Value>(&b).ok())
    else {
        return false;
    };
    v["kind"] == KIND
        && v["stamp"].as_str() == Some(stamp)
        && v["home"].as_str() == Some(home.display().to_string().as_str())
}

/// The nearest folder at or above `p` that exists (the file system the archive will be on).
fn existing_ancestor(p: &Path) -> Option<&Path> {
    p.ancestors().find(|a| fs::symlink_metadata(a).is_ok())
}

/// The `PATH` entries that name the earlier folder, as written.
fn path_entries(path: Option<&OsString>, old: &Path) -> Vec<String> {
    let Some(path) = path else {
        return Vec::new();
    };
    let canon_old = canon(old);
    std::env::split_paths(path)
        .filter(|e| e.starts_with(old) || e.starts_with(&canon_old))
        .map(|e| e.display().to_string())
        .collect()
}

const GUIDANCE: &str = "No shell file is read or edited. The launcher and the PATH links inside the \
earlier folder move with it. Run `moai update` first (it installs under the new state folder), \
put the folder it names on PATH, and remove any PATH line that still names the earlier folder yourself.";

/// The preview for the HOME at `home`; `now` (`YYYY-MM-DDTHH:MM:SSZ`) names the archive.
/// Reads names, kinds and modes by `lstat`; opens no file and follows no link. Writes nothing.
pub fn plan(home: &Path, now: &str, facts: &Facts) -> Result<Value, String> {
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let stamp = stamp(now)?;
    let n = names(home, &stamp);
    let old = earlier(home);
    let moai = home.join(MOAI_DIR);

    let mut blocks: Vec<String> = Vec::new();
    let mut entries: Vec<Value> = Vec::new();
    let mut total = 0usize;
    let present = match fs::symlink_metadata(&old) {
        Err(_) => false,
        Ok(m) if m.file_type().is_symlink() => {
            blocks.push("earlier_folder_is_a_link".into());
            true
        }
        Ok(m) if !m.is_dir() => {
            blocks.push("earlier_folder_not_a_folder".into());
            true
        }
        Ok(_) => true,
    };
    if present && blocks.is_empty() {
        match fs::symlink_metadata(&moai) {
            Ok(m) if !m.is_dir() => blocks.push("moai_folder_not_a_folder".into()),
            Ok(_) => match fs::symlink_metadata(&n.root) {
                Ok(m) if !m.is_dir() => blocks.push("archive_not_a_folder".into()),
                Ok(_) => {
                    let taken = [&n.folder, &n.journal]
                        .iter()
                        .any(|p| fs::symlink_metadata(p).is_ok());
                    let manifest_taken = fs::symlink_metadata(&n.manifest).is_ok()
                        && !own_manifest(&n, home, &stamp);
                    if taken || manifest_taken {
                        blocks.push("stamp_taken".into());
                    }
                }
                Err(_) => {}
            },
            Err(_) => {}
        }
        match fs::read_dir(&old) {
            Err(_) => blocks.push("earlier_folder_unreadable".into()),
            Ok(rd) => {
                let mut unreadable = 0usize;
                let mut found: Vec<(String, PathBuf)> = Vec::new();
                for e in rd.flatten() {
                    match e.file_name().into_string() {
                        Ok(name) => found.push((name, e.path())),
                        Err(_) => unreadable += 1,
                    }
                }
                found.sort();
                total = found.len() + unreadable;
                if unreadable > 0 {
                    blocks.push(format!("unreadable_names:{unreadable}"));
                }
                if total > CANDIDATE_LIMIT {
                    blocks.push("not_fully_listed".into());
                }
                for (name, p) in found.into_iter().take(CANDIDATE_LIMIT) {
                    let mode = fs::symlink_metadata(&p).ok().map(|m| mode_text(&m));
                    entries.push(json!({"name": name, "kind": kind_of(&p), "mode": mode}));
                }
            }
        }
        if blocks.is_empty() {
            let parent = existing_ancestor(&n.root).unwrap_or(&moai);
            match move_journal::same_device(&old, parent) {
                Some(true) => {}
                Some(false) => blocks.push("different_file_system".into()),
                None => blocks.push("device_unknown".into()),
            }
        }
        if blocks.is_empty() {
            let c = canon(&old);
            if facts
                .exe
                .as_deref()
                .is_some_and(|e| e.starts_with(&old) || canon(e).starts_with(&c))
            {
                blocks.push("running_binary_inside".into());
            }
        }
        if blocks.is_empty() {
            match InUse::read().uses(&old, &canon(&old)) {
                Ok(None) => {}
                Ok(Some(why)) | Err(why) => blocks.push(format!("in_use:{why}")),
            }
        }
    }

    let action = if !present {
        "none"
    } else if !blocks.is_empty() {
        "blocked"
    } else {
        "retire"
    };
    let destination = (action == "retire").then(|| n.folder.display().to_string());
    let path_entries = path_entries(facts.path.as_ref(), &old);
    let mut warnings: Vec<String> = Vec::new();
    if action == "retire" && !fs::symlink_metadata(moai.join("versions")).is_ok_and(|m| m.is_dir())
    {
        warnings.push(
            "the new state folder holds no version folder yet: run `moai update` before the \
             earlier folder is archived"
                .into(),
        );
    }
    let shown: Vec<Value> = entries.clone();
    let bound = json!({
        "mode": "state-retire", "home": home.display().to_string(), "action": action,
        "blocks": blocks, "destination": destination, "entries": shown,
    });
    let plan_digest = crate::digest(serde_json::to_string(&bound).unwrap_or_default().as_bytes());
    // The confirmation without the clock: the archive name is out, so a preview and an apply at
    // different times can still agree.
    let stable = json!({
        "mode": "state-retire", "home": home.display().to_string(), "action": action,
        "blocks": blocks, "entries": entries,
    });
    let content_digest = crate::digest(
        serde_json::to_string(&stable)
            .unwrap_or_default()
            .as_bytes(),
    );
    Ok(json!({
        "schema_version": 1,
        "mode": "state-retire",
        "home": home.display().to_string(),
        "action": action,
        "source": present.then(|| old.display().to_string()),
        "destination": destination,
        "blocks": blocks,
        "entries": entries,
        "total": total,
        "path_entries": path_entries,
        "path_guidance": GUIDANCE,
        "warnings": warnings,
        "plan_digest": plan_digest,
        "content_digest": content_digest,
        "apply": "home-archive-apply with the same --now and this plan_digest (or this content_digest) and a permission reference; on a real HOME it runs only after the user confirms",
        "writes": false,
    }))
}

fn make_dir(p: &Path) -> Result<(), String> {
    let mut b = fs::DirBuilder::new();
    #[cfg(unix)]
    std::os::unix::fs::DirBuilderExt::mode(&mut b, 0o700);
    b.create(p).map_err(|e| format!("{}: {e}", p.display()))
}

/// Write `bytes` to `path` and flush them: `exclusive` refuses a name that exists; otherwise a
/// leftover file of an earlier crash (the temporary name of the manifest) is overwritten.
fn write_file(path: &Path, bytes: &[u8], exclusive: bool) -> Result<(), String> {
    use std::io::Write;
    let mut opts = fs::OpenOptions::new();
    opts.write(true);
    if exclusive {
        opts.create_new(true);
    } else {
        opts.create(true).truncate(true);
    }
    let mut f = opts
        .open(path)
        .map_err(|e| format!("{}: {e}", path.display()))?;
    f.write_all(bytes)
        .and_then(|()| f.sync_all())
        .map_err(|e| format!("{}: {e}", path.display()))
}

/// Archive the earlier folder of an identical fresh preview whose action is `retire`. A journal
/// of the same clock that an earlier run left (a crash) is finished instead, when its manifest
/// carries the same digest; a manifest without a journal (a crash between the two) is replaced
/// by the fresh one and the move goes on.
pub fn apply(
    home: &Path,
    now: &str,
    confirm: Confirm,
    expected: &str,
    permission_reference: &str,
    facts: &Facts,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    let stamp = stamp(now)?;
    let n = names(home, &stamp);
    let old = earlier(home);

    let resuming = fs::symlink_metadata(&n.manifest).is_ok_and(|m| m.is_file())
        && fs::symlink_metadata(&n.journal).is_ok_and(|m| m.is_file());
    let fault = test_fault();
    let procs = InUse::read();
    let check = |p: &Path| match procs.uses(p, &canon(p)) {
        Ok(None) => move_journal::InUse::No,
        Ok(Some(why)) => move_journal::InUse::Yes(why.to_string()),
        Err(_) => move_journal::InUse::Unknown,
    };
    let env = MoveEnv {
        in_use: &check,
        same_device: &move_journal::same_device,
        fault: fault.as_deref(),
    };

    let report = if resuming {
        let manifest: Value = fs::read(&n.manifest)
            .ok()
            .and_then(|b| serde_json::from_slice(&b).ok())
            .ok_or_else(|| format!("{}: not a readable manifest", n.manifest.display()))?;
        if manifest["kind"] != KIND
            || manifest["home"].as_str() != Some(home.display().to_string().as_str())
            || manifest[confirm.key()].as_str() != Some(expected)
        {
            return Err(
                "a journal of this clock belongs to another plan; nothing was written".into(),
            );
        }
        move_journal::resume(&n.journal, &env)
            .map_err(|e| format!("{e}; home-archive-revert puts back what moved"))?
    } else {
        let fresh = plan(home, now, facts)?;
        if fresh[confirm.key()].as_str() != Some(expected) {
            return Err("stale or altered plan: the recomputed digest differs; re-preview required, nothing was written".into());
        }
        if fresh["action"] != "retire" {
            return Err(format!(
                "the preview's action is `{}`, not `retire`; nothing was written",
                fresh["action"].as_str().unwrap_or("?")
            ));
        }
        let moai = home.join(MOAI_DIR);
        for dir in [&moai, &n.root] {
            match fs::symlink_metadata(dir) {
                Ok(m) if m.is_dir() => {}
                Ok(_) => {
                    return Err(format!(
                        "{} is not a folder; nothing was written",
                        dir.display()
                    ))
                }
                Err(_) => make_dir(dir).map_err(|e| format!("{e}; nothing was written"))?,
            }
        }
        let manifest = json!({
            "schema_version": 1,
            "kind": KIND,
            "stamp": stamp,
            "home": home.display().to_string(),
            "created_at": now,
            "plan_digest": fresh["plan_digest"],
            "content_digest": fresh["content_digest"],
            "permission_reference": permission_reference,
            "journal": n.journal.display().to_string(),
            "from": old.display().to_string(),
            "to": n.folder.display().to_string(),
            "entries": fresh["entries"],
        });
        let bytes = format!("{manifest:#}\n");
        if own_manifest(&n, home, &stamp) {
            // A crash left this step's own manifest without a journal: replace it by one
            // rename, so the name never holds a half-written file.
            let tmp = n.manifest.with_extension("json.tmp");
            write_file(&tmp, bytes.as_bytes(), false)?;
            fs::rename(&tmp, &n.manifest).map_err(|e| format!("{}: {e}", n.manifest.display()))?;
        } else {
            write_file(&n.manifest, bytes.as_bytes(), true)?;
        }
        if fault.as_deref() == Some("after_manifest") {
            return Err("injected crash at after_manifest".into());
        }
        let items = [PlannedMove {
            from: old.clone(),
            to: n.folder.clone(),
        }];
        move_journal::apply(&n.root, KIND, &items, &BTreeMap::new(), now, &env).map_err(|e| {
            format!(
                "{e}; the journal in {} lets home-archive-revert put back what moved",
                n.root.display()
            )
        })?
    };
    let moved = report.items.iter().filter(|i| i.outcome == "moved").count();
    Ok(json!({
        "schema_version": 1,
        "state": if moved == report.items.len() { "applied" } else { "partial" },
        "stamp": stamp,
        "home": home.display().to_string(),
        "archive": n.folder.display().to_string(),
        "manifest": n.manifest.display().to_string(),
        "journal": report.journal,
        "journal_state": report.state,
        "moved": moved,
        "not_moved": report.items.len() - moved,
        "items": report.items.iter().map(|i| json!({"outcome": i.outcome})).collect::<Vec<_>>(),
        "revert": format!("home-archive-revert --home {} --stamp {stamp} --permission-reference <ref>", home.display()),
        "path_guidance": GUIDANCE,
    }))
}

#[cfg(debug_assertions)]
fn test_fault() -> Option<String> {
    std::env::var(FAULT_VAR).ok().filter(|v| !v.is_empty())
}

#[cfg(not(debug_assertions))]
fn test_fault() -> Option<String> {
    None
}

/// The newest journal of this step in `root`, or the one of `stamp`.
fn find_journal(root: &Path, stamp: Option<&str>) -> Result<PathBuf, String> {
    if let Some(s) = stamp {
        if !crate::state_root::utc_stamp(s) {
            return Err(format!("stamp `{s}` is not YYYYMMDDTHHMMSSZ"));
        }
        return Ok(move_journal::journal_path(root, s));
    }
    let mut found: Vec<String> = Vec::new();
    for e in fs::read_dir(root)
        .map_err(|e| format!("{}: {e}", root.display()))?
        .flatten()
    {
        let name = e.file_name().to_string_lossy().into_owned();
        let Some(s) = name
            .strip_prefix(move_journal::PREFIX)
            .and_then(|n| n.strip_suffix(".json"))
        else {
            continue;
        };
        if crate::state_root::utc_stamp(s) {
            found.push(s.to_string());
        }
    }
    found.sort();
    found
        .iter()
        .rev()
        .map(|s| move_journal::journal_path(root, s))
        .find(|p| move_journal::read_journal(p).is_ok_and(|j| j.kind == KIND))
        .ok_or_else(|| {
            format!(
                "no archive journal in {}; nothing was moved",
                root.display()
            )
        })
}

/// Put the archived folder back: the journal of `stamp` (the newest when `None`) is replayed in
/// reverse, the entry only while it is still the one that was moved and its old place is free.
/// Deletes nothing.
pub fn revert(
    home: &Path,
    stamp: Option<&str>,
    permission_reference: &str,
) -> Result<Value, String> {
    if !crate::contract::permission_reference_ok(permission_reference) {
        return Err("permission reference required".into());
    }
    if !home.is_absolute() {
        return Err("HOME is not an absolute path".into());
    }
    let root = home.join(MOAI_DIR).join(ARCHIVE_DIR);
    if !fs::symlink_metadata(&root).is_ok_and(|m| m.is_dir()) {
        return Err(format!(
            "{} is not a folder; nothing was moved",
            root.display()
        ));
    }
    if fs::symlink_metadata(home.join(MOAI_DIR)).is_ok_and(|m| !m.is_dir()) {
        return Err("the new state folder is a link or not a folder; nothing was moved".into());
    }
    let path = find_journal(&root, stamp)?;
    let journal = move_journal::read_journal(&path)?;
    if journal.kind != KIND {
        return Err(format!(
            "{} is not an archive journal; nothing was moved",
            path.display()
        ));
    }
    let report = move_journal::revert(&path)?;
    Ok(json!({
        "schema_version": 1,
        "state": report.state,
        "home": home.display().to_string(),
        "journal": report.journal,
        "changed": report.changed,
        "items": report.items.iter().map(|i| json!({"outcome": i.outcome})).collect::<Vec<_>>(),
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn path_entries_keep_only_the_entries_under_the_earlier_folder() {
        let home = tempfile::tempdir().unwrap();
        let old = home.path().join(crate::HOME_STATE_DIR);
        let path = std::env::join_paths([
            old.join("path"),
            PathBuf::from("/usr/bin"),
            home.path().join(".moai-cli-other"),
        ])
        .unwrap();
        assert_eq!(
            path_entries(Some(&path), &old),
            [old.join("path").display().to_string()]
        );
        assert!(path_entries(None, &old).is_empty());
    }

    #[test]
    fn the_names_of_one_stamp_sit_in_the_archive_folder() {
        let n = names(Path::new("/h"), "20261123T120000Z");
        assert_eq!(
            n.folder,
            Path::new("/h/.moai/archive/moai-cli-20261123T120000Z")
        );
        assert_eq!(
            n.journal,
            Path::new("/h/.moai/archive/move-journal-20261123T120000Z.json")
        );
    }
}
