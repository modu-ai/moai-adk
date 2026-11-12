//! #2561: Go-generation project state — the read-only preview (`go-state-plan`).
//!
//! The Go tool records what it deployed into a project in `.moai/manifest.json`
//! (`internal/manifest/types.go`: provenance `template_managed` means "deployed from a
//! template and not modified by the user"; hashes are `sha256:<hex>`). That file is read
//! as untrusted input. A path is offered for moving only when every check holds:
//! a plain relative key (no root, no `..`, no backslash), under `.moai/` or `.claude/`,
//! no path component is a symlink, the path is a regular file, the manifest records it as
//! `template_managed`, its bytes still hash to the recorded `deployed_hash`, and
//! neither `.claude/settings.json` nor `.claude/settings.local.json` references it, and it
//! is not one of those two live settings files. Anything else stays, with its reason;
//! files the manifest does not record are never offered and are counted per folder.
//! Without a usable manifest nothing moves. Which of these rules apply to `.moai/` as a
//! whole is a pending decision (#2561 comment 6827973256); [`policy`] is the one place
//! that choice lives. The preview writes nothing; the apply and revert below it write
//! only under the project's `.moai-cli/backups/` and move the planned files.

use serde::{Deserialize, Serialize};
// Stage 2 (`go-state-apply`, `go-state-revert-plan`, `go-state-revert`) is below `plan`.
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, HashMap},
    fs,
    path::{Component, Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const SCHEMA_VERSION: u32 = 1;
const MANIFEST: &str = ".moai/manifest.json";
const SETTINGS: &str = ".claude/settings.json";
const SETTINGS_LOCAL: &str = ".claude/settings.local.json";
/// The Go manifest of this checkout is about 180 KiB; this only refuses absurd inputs.
const MANIFEST_LIMIT: u64 = 16 * 1024 * 1024;
const FILE_LIMIT: u64 = 64 * 1024 * 1024;
const ROOTS: [&str; 2] = [".moai", ".claude"];
/// Decision 1 (#2561 comment 6831345102): `.moai/` moves as a whole folder.
const MOAI_DIR: &str = ".moai";
/// AC-5: what is restored from the moved folder, and where (Rust reads SPECs there).
const SPECS_FROM: &str = ".moai/specs";
const SPECS_TO: &str = ".moai-cli/specs";

#[derive(Deserialize)]
struct Manifest {
    version: String,
    files: HashMap<String, Entry>,
}

#[derive(Deserialize)]
struct Entry {
    provenance: String,
    #[serde(default)]
    deployed_hash: String,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct Item {
    pub path: String,
    /// `move` or `leave`.
    pub action: String,
    pub reason: String,
}

#[derive(Clone, Debug, Serialize)]
pub struct GoStatePlan {
    pub schema_version: u32,
    pub project: String,
    pub manifest_version: Option<String>,
    pub items: Vec<Item>,
    /// Regular files under `.moai/` and `.claude/` the manifest does not record, counted per
    /// second-level folder (`.moai/specs`, `.claude/worktrees`, …). They are never moved.
    pub unrecorded: BTreeMap<String, u64>,
    /// AC-5: files copied back from the moved `.moai/specs` into `.moai-cli/specs`
    /// (`copy`), or not (`conflict`, `conflict_link`, `excluded_secret`, `excluded_unclassifiable`,
    /// `leave_link`). Empty unless the folder moves.
    pub restores: Vec<Restore>,
    pub warnings: Vec<String>,
    pub plan_digest: String,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq)]
pub struct Restore {
    pub from: String,
    pub to: String,
    pub sha256: String,
    pub action: String,
}

/// The restore candidates under `.moai/specs`: regular files are copied unless the
/// target exists (never overwritten) or the content is or may be a secret (the `migrate`
/// filter, fail-closed); links are never followed.
fn plan_restores(project: &Path) -> Vec<Restore> {
    let mut out = Vec::new();
    let mut stack = vec![project.join(SPECS_FROM)];
    while let Some(dir) = stack.pop() {
        let Ok(entries) = fs::read_dir(&dir) else {
            continue;
        };
        for e in entries.flatten() {
            let p = e.path();
            let Ok(m) = fs::symlink_metadata(&p) else {
                continue;
            };
            if m.is_dir() {
                stack.push(p);
                continue;
            }
            let Some(rel) = p
                .strip_prefix(project.join(SPECS_FROM))
                .ok()
                .and_then(|r| r.to_str())
                .map(|r| r.replace('\\', "/"))
            else {
                continue;
            };
            let from = format!("{SPECS_FROM}/{rel}");
            let to = format!("{SPECS_TO}/{rel}");
            let (sha, action) = if !m.is_file() {
                (String::new(), "leave_link")
            } else {
                match fs::read(&p) {
                    Ok(bytes) if (bytes.len() as u64) <= FILE_LIMIT => {
                        let sha = format!("sha256:{:x}", Sha256::digest(&bytes));
                        let action = match crate::migrate::sensitive_content(&bytes, true) {
                            Ok(true) => "excluded_secret",
                            Err(_) => "excluded_unclassifiable",
                            Ok(false) if linked_component(project, &to) => "conflict_link",
                            Ok(false) if fs::symlink_metadata(project.join(&to)).is_ok() => {
                                "conflict"
                            }
                            Ok(false) => "copy",
                        };
                        (sha, action)
                    }
                    _ => (String::new(), "excluded_unclassifiable"),
                }
            };
            out.push(Restore {
                from,
                to,
                sha256: sha,
                action: action.into(),
            });
        }
    }
    out.sort_by(|a, b| a.from.cmp(&b.from));
    out
}

fn sha256_file(path: &Path) -> Result<String> {
    let meta = fs::symlink_metadata(path).map_err(|_| "file metadata unavailable")?;
    if !meta.is_file() || meta.len() > FILE_LIMIT {
        return Err("not a bounded regular file".into());
    }
    let bytes = fs::read(path).map_err(|_| "file read failed")?;
    Ok(format!("sha256:{:x}", Sha256::digest(bytes)))
}

/// A plain relative key: normal components only, no backslash, at least one component.
fn plain_relative(key: &str) -> bool {
    !key.is_empty()
        && !key.contains('\\')
        && Path::new(key)
            .components()
            .all(|c| matches!(c, Component::Normal(_)))
}

/// Whether any component from the project root down to `rel` is a symlink.
fn linked_component(project: &Path, rel: &str) -> bool {
    let mut cur = project.to_path_buf();
    for part in Path::new(rel).components() {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => return true,
            Ok(_) => {}
            Err(_) => return false,
        }
    }
    false
}

/// The decision for one recorded path. The single place the move rule lives.
fn policy(
    project: &Path,
    key: &str,
    entry: &Entry,
    settings: &[(&'static str, String)],
    folder_moves: bool,
) -> (&'static str, &'static str) {
    if !plain_relative(key) {
        return ("leave", "not a plain relative path");
    }
    let top = Path::new(key)
        .components()
        .next()
        .and_then(|c| c.as_os_str().to_str())
        .unwrap_or("");
    if !ROOTS.contains(&top) {
        return ("leave", "outside .moai/ and .claude/");
    }
    // Decision 1: everything under `.moai/` follows the folder (one rename), so no file
    // there moves on its own; when the folder stays, all of it stays.
    if top == MOAI_DIR {
        return if folder_moves {
            ("with_folder", "moves with .moai/ as a whole")
        } else {
            ("leave", "the .moai/ folder stays (see the .moai item)")
        };
    }
    if key == MANIFEST {
        return ("leave", "the Go manifest itself");
    }
    // AUDIT #2584 B1: the settings files are the live host settings (permissions.deny,
    // hooks, statusLine) even when the Go manifest proves them unchanged. Moving them is a
    // policy question (decision 1 of #2561 comment 6827973256), never a default.
    if key == SETTINGS || key == SETTINGS_LOCAL {
        return (
            "leave",
            "live host settings; moving them drops permissions, hooks and the status line",
        );
    }
    if linked_component(project, key) {
        return ("leave", "a path component is a symlink");
    }
    let path = project.join(key);
    match fs::symlink_metadata(&path) {
        Err(_) => return ("leave", "missing"),
        Ok(m) if !m.is_file() => return ("leave", "not a regular file"),
        Ok(_) => {}
    }
    if entry.provenance != "template_managed" {
        return ("leave", "not template_managed");
    }
    match sha256_file(&path) {
        Ok(h) if !entry.deployed_hash.is_empty() && h == entry.deployed_hash => {}
        Ok(_) => return ("leave", "changed since the Go deploy"),
        Err(_) => return ("leave", "not a bounded regular file"),
    }
    for (file, text) in settings {
        if referenced(text, key) {
            return (
                "leave",
                if *file == SETTINGS_LOCAL {
                    "referenced by .claude/settings.local.json"
                } else {
                    "referenced by .claude/settings.json"
                },
            );
        }
    }
    ("move", "proven Go deploy, unchanged")
}

/// `.claude/settings.json` names the file by path (hooks, statusline) or, for an output
/// style, by the style name in `outputStyle`.
fn referenced(settings: &str, key: &str) -> bool {
    if settings.is_empty() {
        return false;
    }
    if settings.contains(key) {
        return true;
    }
    if let Some(name) = key
        .strip_prefix(".claude/output-styles/")
        .and_then(|r| Path::new(r).file_stem())
        .and_then(|s| s.to_str())
    {
        let style = serde_json::from_str::<serde_json::Value>(settings)
            .ok()
            .and_then(|v| v["outputStyle"].as_str().map(str::to_string));
        return style.as_deref() == Some(name);
    }
    false
}

fn read_bounded_text(project: &Path, rel: &str, limit: u64) -> Result<Option<String>> {
    if linked_component(project, rel) {
        return Err(format!("{rel}: a path component is a symlink; refused"));
    }
    let path = project.join(rel);
    match fs::symlink_metadata(&path) {
        Err(_) => Ok(None),
        Ok(m) if !m.is_file() || m.len() > limit => {
            Err(format!("{rel} is not a bounded regular file"))
        }
        Ok(_) => fs::read_to_string(&path)
            .map(Some)
            .map_err(|_| format!("{rel} read failed")),
    }
}

fn count_unrecorded(project: &Path, recorded: &HashMap<String, Entry>) -> BTreeMap<String, u64> {
    let mut counts = BTreeMap::new();
    for root in ROOTS {
        let base = project.join(root);
        let Ok(m) = fs::symlink_metadata(&base) else {
            continue;
        };
        if !m.is_dir() {
            continue;
        }
        let mut stack: Vec<PathBuf> = vec![base];
        while let Some(dir) = stack.pop() {
            let Ok(entries) = fs::read_dir(&dir) else {
                continue;
            };
            for e in entries.flatten() {
                let p = e.path();
                let Ok(m) = fs::symlink_metadata(&p) else {
                    continue;
                };
                if m.is_dir() {
                    stack.push(p);
                    continue;
                }
                if !m.is_file() {
                    continue;
                }
                let Ok(rel) = p.strip_prefix(project) else {
                    continue;
                };
                let rel = rel.to_string_lossy().replace('\\', "/");
                if recorded.contains_key(&rel) || rel == MANIFEST {
                    continue;
                }
                let folder: String = rel.splitn(3, '/').take(2).collect::<Vec<_>>().join("/");
                *counts.entry(folder).or_insert(0) += 1;
            }
        }
    }
    counts
}

/// The read-only preview over one project folder.
pub fn plan(project: &Path) -> Result<GoStatePlan> {
    let project = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    // A linked `.moai` (or any linked component above the manifest) is refused by the
    // manifest read below, with or without a manifest file.
    let mut warnings = vec![
        "Running Claude Code sessions in this project may read .claude/hooks, rules and output \
         styles; apply only while none is running (this preview cannot tell)."
            .to_string(),
    ];
    let manifest = match read_bounded_text(&project, MANIFEST, MANIFEST_LIMIT)? {
        None => None,
        Some(text) => serde_json::from_str::<Manifest>(&text).ok(),
    };
    let mut settings = Vec::new();
    for file in [SETTINGS, SETTINGS_LOCAL] {
        let text = read_bounded_text(&project, file, MANIFEST_LIMIT)?.unwrap_or_default();
        settings.push((file, text));
    }
    let mut items = Vec::new();
    let (version, recorded) = match manifest {
        Some(m) => (Some(m.version), m.files),
        None => {
            warnings.push(format!(
                "no usable Go manifest at {MANIFEST}: nothing is proven tool-made, so nothing moves"
            ));
            (None, HashMap::new())
        }
    };
    // The `.moai/` folder item. Only with a usable manifest (nothing is proven otherwise),
    // only a real folder (a linked `.moai` was refused by the manifest read above), and
    // never while either settings file names a path under `.moai/` (decision 2: what the
    // settings reference stays, fail-closed; cleanup is #2581).
    let folder_is_dir = fs::symlink_metadata(project.join(MOAI_DIR))
        .map(|m| m.is_dir())
        .unwrap_or(false);
    let referenced_by_settings = settings.iter().any(|(_, text)| text.contains(".moai/"));
    let folder_moves = version.is_some() && folder_is_dir && !referenced_by_settings;
    if version.is_some() && folder_is_dir {
        items.push(Item {
            path: MOAI_DIR.into(),
            action: if folder_moves { "move" } else { "leave" }.into(),
            reason: if folder_moves {
                "the Go .moai/ folder as a whole (decision 1, #2561 comment 6831345102)"
            } else {
                "settings reference .moai/; the folder stays until they are cleaned (#2581)"
            }
            .into(),
        });
    }
    if folder_moves {
        warnings.push(
            "Hooks that write under .moai/ recreate it while sessions run; a revert refuses \
             while .moai/ exists again."
                .to_string(),
        );
    }
    let mut keys: Vec<&String> = recorded.keys().collect();
    keys.sort();
    for key in keys {
        let (action, reason) = policy(&project, key, &recorded[key], &settings, folder_moves);
        items.push(Item {
            path: key.clone(),
            action: action.into(),
            reason: reason.into(),
        });
    }
    let unrecorded = if version.is_some() {
        count_unrecorded(&project, &recorded)
    } else {
        BTreeMap::new()
    };
    let restores = if folder_moves {
        plan_restores(&project)
    } else {
        Vec::new()
    };
    let mut plan = GoStatePlan {
        schema_version: SCHEMA_VERSION,
        project: project
            .to_str()
            .ok_or("non UTF-8 project path")?
            .to_string(),
        manifest_version: version,
        items,
        unrecorded,
        restores,
        warnings,
        plan_digest: String::new(),
    };
    let bytes = serde_json::to_vec(&plan).map_err(|_| "encode failed")?;
    plan.plan_digest = format!("sha256:{:x}", Sha256::digest(bytes));
    Ok(plan)
}

// ---- stage 2: apply and revert ----
//
// The apply moves each planned file, one by one, into
// `<project>/.moai-cli/backups/<stamp>/go-state/<same relative path>`; the receipt beside
// it (`go-state-receipt.json`) is written `pending` before the first move and records
// every move as it happens, so a revert also works after a failure midway. A move is a
// hard link to the backup path followed by removing the source: the link fails when the
// backup path exists (no overwrite) or lies on another filesystem (no copy fallback).
// Emptied folders stay, so a revert brings the tree back exactly. The policy (which files
// move) is the preview's; only files proven tool-made move until decision 1 of
// #2561 comment 6827973256 says otherwise.

const BACKUPS: &str = ".moai-cli/backups";
const RECEIPT: &str = "go-state-receipt.json";

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Moved {
    pub path: String,
    /// A file's sha256, or for the folder its `baseline::tree_digest` taken right after
    /// the move (empty until then).
    pub sha256: String,
    pub moved: bool,
    /// `folder` for the `.moai/` rename; absent (a file) in stage-2 receipts.
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub kind: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub schema_version: u32,
    pub project: String,
    pub backup: String,
    pub plan_digest: String,
    /// The caller's `--permission-reference`, recorded as given (not an authorization).
    pub permission_reference: String,
    /// `pending` → `committed` → `reverted`.
    pub state: String,
    pub moves: Vec<Moved>,
    /// AC-5: the copies written under `.moai-cli/specs`, each with its sha256.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub restored: Vec<Restored>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct Restored {
    pub path: String,
    pub sha256: String,
}

/// A backup name is the apply's UTC stamp, `YYYYMMDDTHHMMSSZ`, and nothing else.
fn valid_stamp(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 16
        && b[8] == b'T'
        && b[15] == b'Z'
        && b[..8].iter().chain(&b[9..15]).all(u8::is_ascii_digit)
}

fn stamp_now() -> String {
    crate::baseline::utc_now().replace(['-', ':'], "")
}

/// Every existing component of `rel` under `project` must be a real folder (not a link).
fn refuse_linked(project: &Path, rel: &str) -> Result<()> {
    if linked_component(project, rel) {
        return Err(format!("{rel}: a path component is a symlink; refused"));
    }
    Ok(())
}

fn save_receipt(project: &Path, receipt: &Receipt, before: Option<&[u8]>) -> Result<Vec<u8>> {
    let rel = format!("{BACKUPS}/{}/{RECEIPT}", receipt.backup);
    refuse_linked(project, &rel)?;
    let bytes = serde_json::to_vec_pretty(receipt).map_err(|_| "encode failed")?;
    let path = project.join(&rel);
    match before {
        None => crate::apply::publish_exclusive(&path, &bytes)?,
        Some(b) => crate::apply::publish_replace(&path, Some(b), &bytes)?,
    }
    Ok(bytes)
}

/// Create the folders of `rel` (relative to `project`) one by one, refusing a link.
fn make_dirs(project: &Path, rel: &str) -> Result<()> {
    let mut cur = project.to_path_buf();
    for part in Path::new(rel).components() {
        cur.push(part);
        match fs::symlink_metadata(&cur) {
            Ok(m) if m.file_type().is_symlink() => {
                return Err(format!("{rel}: a path component is a symlink; refused"))
            }
            Ok(m) if m.is_dir() => {}
            Ok(_) => return Err(format!("{rel}: a path component is not a folder")),
            Err(_) => fs::create_dir(&cur).map_err(|_| format!("{rel}: folder create failed"))?,
        }
    }
    Ok(())
}

/// Move one file without overwriting: hard link, then remove the source.
fn move_exclusive(from: &Path, to: &Path) -> Result<()> {
    fs::hard_link(from, to).map_err(|e| {
        if e.kind() == std::io::ErrorKind::AlreadyExists {
            "move refused: the destination already exists".to_string()
        } else if e.raw_os_error() == Some(18) {
            // EXDEV on Linux and macOS.
            "cross-filesystem move refused; the project and its backup must share one filesystem"
                .to_string()
        } else {
            "move failed (hard link)".to_string()
        }
    })?;
    fs::remove_file(from).map_err(|_| "move failed: source not removed (both copies kept)".into())
}

/// Move one folder with a single rename: the source must be a real folder (not a link),
/// the destination must not exist, and another filesystem is refused (no copy fallback).
fn move_folder(from: &Path, to: &Path) -> Result<()> {
    match fs::symlink_metadata(from) {
        Ok(m) if m.is_dir() => {}
        Ok(_) => return Err("folder move refused: the source is not a real folder".into()),
        Err(_) => return Err("folder move refused: the source is missing".into()),
    }
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

/// Apply the preview whose digest the caller confirmed. Nothing is written unless a fresh
/// preview has exactly that digest.
pub fn apply_with_stamp(
    project: &Path,
    plan_digest: &str,
    permission_reference: &str,
    stamp: &str,
) -> Result<Receipt> {
    let fresh = plan(project)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or unconfirmed plan: the digest does not match a fresh go-state-plan; nothing written".into());
    }
    if permission_reference.trim().is_empty() {
        return Err("a permission reference is required".into());
    }
    if !valid_stamp(stamp) {
        return Err("invalid backup stamp".into());
    }
    let project = PathBuf::from(&fresh.project);
    let planned: Vec<&Item> = fresh.items.iter().filter(|i| i.action == "move").collect();
    let mut receipt = Receipt {
        schema_version: SCHEMA_VERSION,
        project: fresh.project.clone(),
        backup: stamp.to_string(),
        plan_digest: fresh.plan_digest.clone(),
        permission_reference: permission_reference.to_string(),
        state: "unchanged".into(),
        moves: Vec::new(),
        restored: Vec::new(),
    };
    if planned.is_empty() {
        return Ok(receipt);
    }
    if let Some(c) = fresh
        .restores
        .iter()
        .find(|r| r.action.starts_with("conflict"))
    {
        return Err(if c.action == "conflict_link" {
            format!(
                "{}: a path component is a symlink; refused; nothing written",
                c.to
            )
        } else {
            format!(
                "restore conflict: {} already exists and is never overwritten; nothing written",
                c.to
            )
        });
    }
    let backup_rel = format!("{BACKUPS}/{stamp}");
    // `make_dirs` refuses a linked `.moai-cli` or `.moai-cli/backups`; the stamp folder
    // itself is created exclusively, so a link already at that name fails the create.
    make_dirs(&project, BACKUPS)?;
    fs::create_dir(project.join(&backup_rel)).map_err(|e| {
        if e.kind() == std::io::ErrorKind::AlreadyExists {
            "exclusive create refused: backup already exists".to_string()
        } else {
            "backup folder create failed".to_string()
        }
    })?;
    // Files first, the folder last: one rename that either happened or did not.
    for item in planned.iter().filter(|i| i.path != MOAI_DIR) {
        let h = sha256_file(&project.join(&item.path))?;
        receipt.moves.push(Moved {
            path: item.path.clone(),
            sha256: h,
            moved: false,
            kind: String::new(),
        });
    }
    if planned.iter().any(|i| i.path == MOAI_DIR) {
        receipt.moves.push(Moved {
            path: MOAI_DIR.into(),
            sha256: String::new(),
            moved: false,
            kind: "folder".into(),
        });
    }
    receipt.state = "pending".into();
    let mut saved = save_receipt(&project, &receipt, None)?;
    for i in 0..receipt.moves.len() {
        let rel = receipt.moves[i].path.clone();
        if receipt.moves[i].kind == "folder" {
            let dest_rel = format!("{backup_rel}/go-state/{MOAI_DIR}");
            make_dirs(&project, &format!("{backup_rel}/go-state"))?;
            move_folder(&project.join(MOAI_DIR), &project.join(&dest_rel))?;
            receipt.moves[i].sha256 = crate::baseline::tree_digest(&project.join(&dest_rel), &[])?;
            receipt.moves[i].moved = true;
            saved = save_receipt(&project, &receipt, Some(&saved))?;
            // AC-5: copy from the backup (the original stays there), never overwriting.
            for r in fresh.restores.iter().filter(|r| r.action == "copy") {
                let src_rel = format!("{backup_rel}/go-state/{}", r.from);
                refuse_linked(&project, &src_rel)?;
                let bytes = fs::read(project.join(&src_rel)).map_err(|_| "restore read failed")?;
                if format!("sha256:{:x}", Sha256::digest(&bytes)) != r.sha256 {
                    return Err(format!(
                        "{} changed after the preview; restore stopped",
                        r.from
                    ));
                }
                let parent = Path::new(&r.to)
                    .parent()
                    .and_then(|p| p.to_str())
                    .ok_or("restore path")?
                    .to_string();
                make_dirs(&project, &parent)?;
                crate::apply::publish_exclusive(&project.join(&r.to), &bytes)?;
                receipt.restored.push(Restored {
                    path: r.to.clone(),
                    sha256: r.sha256.clone(),
                });
                saved = save_receipt(&project, &receipt, Some(&saved))?;
            }
            continue;
        }
        refuse_linked(&project, &rel)?;
        let dest_rel = format!("{backup_rel}/go-state/{rel}");
        let parent = Path::new(&dest_rel)
            .parent()
            .and_then(|p| p.to_str())
            .ok_or("backup path")?
            .to_string();
        make_dirs(&project, &parent)?;
        move_exclusive(&project.join(&rel), &project.join(&dest_rel))?;
        receipt.moves[i].moved = true;
        saved = save_receipt(&project, &receipt, Some(&saved))?;
    }
    receipt.state = "committed".into();
    save_receipt(&project, &receipt, Some(&saved))?;
    Ok(receipt)
}

pub fn apply(project: &Path, plan_digest: &str, permission_reference: &str) -> Result<Receipt> {
    apply_with_stamp(project, plan_digest, permission_reference, &stamp_now())
}

#[derive(Clone, Debug, Serialize)]
pub struct RevertItem {
    pub path: String,
    /// `restore`, `conflict` (something is at the original path again), `missing` (the
    /// backup copy is gone or changed) or `skip` (never moved).
    pub action: String,
}

#[derive(Clone, Debug, Serialize)]
pub struct RevertPlan {
    pub schema_version: u32,
    pub project: String,
    pub backup: String,
    pub receipt_state: String,
    pub items: Vec<RevertItem>,
    pub plan_digest: String,
}

fn read_receipt(project: &Path, backup: &str) -> Result<(Receipt, Vec<u8>)> {
    if !valid_stamp(backup) {
        return Err("invalid backup name: expected the apply's UTC stamp YYYYMMDDTHHMMSSZ".into());
    }
    let rel = format!("{BACKUPS}/{backup}/{RECEIPT}");
    let text = read_bounded_text(project, &rel, MANIFEST_LIMIT)?
        .ok_or_else(|| format!("{rel}: no go-state receipt"))?;
    let receipt: Receipt =
        serde_json::from_str(&text).map_err(|_| format!("{rel}: unreadable receipt"))?;
    if receipt.backup != backup || receipt.project != project.to_string_lossy() {
        return Err("receipt scope mismatch".into());
    }
    if receipt.moves.iter().any(|m| !plain_relative(&m.path)) {
        return Err("receipt holds a path that is not plain relative; refused".into());
    }
    // The receipt is not its own authority: it may only name what an apply can move.
    if receipt.moves.iter().any(|m| {
        let top = Path::new(&m.path)
            .components()
            .next()
            .and_then(|c| c.as_os_str().to_str())
            .unwrap_or("");
        !ROOTS.contains(&top) || m.path == MANIFEST
    }) {
        return Err("receipt holds a path outside .moai/ and .claude/; refused".into());
    }
    // A folder entry is exactly `.moai`; a file entry is never the folder itself.
    if receipt.moves.iter().any(|m| match m.kind.as_str() {
        "folder" => m.path != MOAI_DIR,
        "" => m.path == MOAI_DIR,
        _ => true,
    }) {
        return Err("receipt holds an entry of the wrong kind; refused".into());
    }
    if receipt
        .restored
        .iter()
        .any(|r| !plain_relative(&r.path) || !r.path.starts_with(&format!("{SPECS_TO}/")))
    {
        return Err("receipt holds a restored path outside .moai-cli/specs/; refused".into());
    }
    Ok((receipt, text.into_bytes()))
}

/// Read-only: what a revert of one apply would do.
pub fn revert_plan(project: &Path, backup: &str) -> Result<RevertPlan> {
    let project = fs::canonicalize(project).map_err(|_| "project folder unavailable")?;
    let (receipt, _) = read_receipt(&project, backup)?;
    let mut items = Vec::new();
    for r in &receipt.restored {
        let action = if linked_component(&project, &r.path) {
            "keep_changed"
        } else {
            match sha256_file(&project.join(&r.path)) {
                Ok(h) if h == r.sha256 => "remove_restored",
                Ok(_) => "keep_changed",
                Err(_) => "skip",
            }
        };
        items.push(RevertItem {
            path: r.path.clone(),
            action: action.into(),
        });
    }
    for m in &receipt.moves {
        if m.kind == "folder" {
            let dest_rel = format!("{BACKUPS}/{backup}/go-state/{MOAI_DIR}");
            let source_present = fs::symlink_metadata(project.join(MOAI_DIR)).is_ok();
            let backup_dir = !linked_component(&project, &dest_rel)
                && fs::symlink_metadata(project.join(&dest_rel))
                    .map(|md| md.is_dir())
                    .unwrap_or(false);
            let action = if source_present {
                // Never merged: `.moai/` came back (a live hook) or was never moved.
                if m.moved {
                    "conflict"
                } else {
                    "skip"
                }
            } else if !backup_dir {
                "missing"
            } else if m.sha256.is_empty() {
                // Moved but the receipt write did not follow: the disk says where it is.
                "restore"
            } else {
                match crate::baseline::tree_digest(&project.join(&dest_rel), &[]) {
                    Ok(d) if d == m.sha256 => "restore",
                    _ => "missing",
                }
            };
            items.push(RevertItem {
                path: m.path.clone(),
                action: action.into(),
            });
            continue;
        }
        let action = if !m.moved {
            "skip"
        } else if fs::symlink_metadata(project.join(&m.path)).is_ok()
            || linked_component(&project, &m.path)
        {
            "conflict"
        } else {
            let dest_rel = format!("{BACKUPS}/{backup}/go-state/{}", m.path);
            if linked_component(&project, &dest_rel) {
                "missing"
            } else {
                match sha256_file(&project.join(&dest_rel)) {
                    Ok(h) if h == m.sha256 => "restore",
                    _ => "missing",
                }
            }
        };
        items.push(RevertItem {
            path: m.path.clone(),
            action: action.into(),
        });
    }
    let mut plan = RevertPlan {
        schema_version: SCHEMA_VERSION,
        project: receipt.project.clone(),
        backup: backup.to_string(),
        receipt_state: receipt.state.clone(),
        items,
        plan_digest: String::new(),
    };
    let bytes = serde_json::to_vec(&plan).map_err(|_| "encode failed")?;
    plan.plan_digest = format!("sha256:{:x}", Sha256::digest(bytes));
    Ok(plan)
}

/// Move every backed-up file back, all or nothing up front: any `conflict` or `missing`
/// item refuses the whole revert before the first move. Never overwrites.
pub fn revert(
    project: &Path,
    backup: &str,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<RevertPlan> {
    let fresh = revert_plan(project, backup)?;
    if fresh.plan_digest != plan_digest {
        return Err("stale or unconfirmed revert plan: the digest does not match a fresh go-state-revert-plan; nothing written".into());
    }
    if permission_reference.trim().is_empty() {
        return Err("a permission reference is required".into());
    }
    if fresh.receipt_state == "reverted" {
        return Err("this apply was already reverted".into());
    }
    if let Some(bad) = fresh
        .items
        .iter()
        .find(|i| i.action == "conflict" || i.action == "missing")
    {
        return Err(format!(
            "revert refused: {} is {}; nothing written",
            bad.path, bad.action
        ));
    }
    let project = PathBuf::from(&fresh.project);
    let (mut receipt, mut saved) = read_receipt(&project, backup)?;
    // Unchanged restored copies go; a changed one stays and is reported (`keep_changed`).
    for item in fresh.items.iter().filter(|i| i.action == "remove_restored") {
        refuse_linked(&project, &item.path)?;
        fs::remove_file(project.join(&item.path)).map_err(|_| "restored copy not removed")?;
        receipt.restored.retain(|r| r.path != item.path);
        saved = save_receipt(&project, &receipt, Some(&saved))?;
    }
    for item in fresh.items.iter().filter(|i| i.action == "restore") {
        let parent = Path::new(&item.path)
            .parent()
            .and_then(|p| p.to_str())
            .unwrap_or("")
            .to_string();
        if !parent.is_empty() {
            make_dirs(&project, &parent)?;
        }
        let src = format!("{BACKUPS}/{backup}/go-state/{}", item.path);
        if item.path == MOAI_DIR {
            move_folder(&project.join(src), &project.join(MOAI_DIR))?;
        } else {
            move_exclusive(&project.join(src), &project.join(&item.path))?;
        }
        if let Some(m) = receipt.moves.iter_mut().find(|m| m.path == item.path) {
            m.moved = false;
        }
        saved = save_receipt(&project, &receipt, Some(&saved))?;
    }
    receipt.state = "reverted".into();
    save_receipt(&project, &receipt, Some(&saved))?;
    Ok(fresh)
}
