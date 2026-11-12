//! Host projection UPDATE (Issue #1999, 04-B).
//!
//! `update-plan` inventories every host projection (Claude `~/.claude/skills/moai/`,
//! Codex `~/.agents/skills/moai/` — the design-verified paths; project-level projections
//! are a P04 HOLD and are not included), compares each file three-way (baseline = the
//! ownership manifest of the latest host receipt, current = bytes on disk, new = the
//! staged package) and writes nothing. `update-apply` recomputes the plan and refuses
//! any other digest, then walks the host transaction (REQ-04B-04): pending receipt
//! (fsynced publish) → per-file unique temp + fsync + rename through the 01-B engine
//! (`apply::publish_replace` / `apply::publish_exclusive`, each re-verifying the target
//! and fsyncing the parent directory) → full re-verification → committed receipt. A
//! failed host keeps its `pending` receipt with the PRE-apply ownership manifest
//! (recovery is 04-C) and never touches a host that already committed. A crashed
//! apply therefore reads back as per-file `conflict` where the stage differs and
//! `unchanged` where the bytes already match — never as a silent overwrite.
//!
//! Host receipts live beside the 04-A install receipt:
//! `~/.moai-cli/installations/<install_id>/hosts/moai/<host>.json`. The `install_id` is the
//! 04-A derivation, obtained through the public `install::plan_receipt` so this module
//! does not modify 04-A. Host config keys (Claude `settings.json`, Codex `config.toml`)
//! are out of scope (P04 HOLD — the REQs name projections only). Projections are
//! independent copies; a symlinked projection is `blocked`, never followed (REQ-04B-08),
//! and user files are never deleted or overwritten (REQ-04B-09).

use crate::apply::{decide, publish_exclusive, publish_replace, read_opt, relative_ok, Decision};
use crate::contract::PROTOCOL_VERSION;
use crate::install::{plan_receipt, InstallerOwner};
use serde::{Deserialize, Serialize};
use serde_json::json;
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const UPDATE_SCHEMA_VERSION: u32 = 1;
/// Host projection roots under the home, in fixed plan order. 02-B/02-C expose explicit
/// target paths rather than constants, so the design-verified values stand in here
/// (Issue impl plan step 1).
pub(crate) const HOSTS: [&str; 2] = ["claude", "codex"];
/// Same bounded-file rule as every other state file.
pub(crate) const FILE_LIMIT: u64 = 262_144;
const WALK_DEPTH: usize = 32;
const WALK_ENTRIES: usize = 8192;
const WALK_FILES: usize = 4096;
const WALK_BYTES: usize = 16 * 1024 * 1024;

fn skills_dir(host: &str) -> &'static str {
    match host {
        "claude" => ".claude/skills",
        "codex" => ".agents/skills",
        other => unreachable!("unknown host {other}"),
    }
}

/// The installed skill's projection directory name (#2286): `moai`, the skill
/// package name; the executable, product and state root stay `moai-cli`. The name
/// is shared with the earlier-generation (Go-era) MoAI `moai` skill, so the two no
/// longer coexist by name: a pre-existing `skills/moai` whose files the install
/// receipt does not own is blocked with guidance and never replaced automatically
/// (decision row 22). The host receipt namespace `hosts/<SKILL_DIR>/` follows this
/// name, so receipts written for the earlier `skills/moai-cli` root never supply a
/// baseline for this one.
pub(crate) const SKILL_DIR: &str = "moai";

/// Guidance appended wherever a `moai` skill folder this tool did not write is found
/// (decision row 22, #2286): block and guide, never replace automatically. The
/// backup-then-replace path is a separate, user-confirmed procedure (#2483): the plan
/// previews it and `update-apply --confirm-backup-replace <token>` runs it.
pub(crate) const GO_ERA_GUIDANCE: &str = "this can be an earlier-generation (Go-era) \
     MoAI `moai` skill at that path; moai-cli will not replace it — move it aside \
     yourself, or wait for a confirmed backup-then-replace procedure";

pub(crate) fn projection_root(home: &Path, host: &str) -> PathBuf {
    home.join(skills_dir(host)).join(SKILL_DIR)
}

/// UTC RFC3339 second precision; local copy of install.rs's private formatter.
pub(crate) fn now_rfc3339() -> String {
    let secs = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs())
        .unwrap_or(0);
    let days = (secs / 86400) as i64;
    let rem = secs % 86400;
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z.rem_euclid(146_097);
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = if m <= 2 { y + 1 } else { y };
    format!(
        "{y:04}-{m:02}-{d:02}T{:02}:{:02}:{:02}Z",
        rem / 3600,
        (rem % 3600) / 60,
        rem % 60
    )
}

pub(crate) fn sha256_hex(bytes: &[u8]) -> String {
    format!("sha256:{:x}", Sha256::digest(bytes))
}

// --- staged package -----------------------------------------------------------------

/// The staged new package: `moai-package-tree-v1` digest plus every file's bytes.
pub struct StagedPackage {
    pub digest: String,
    pub files: BTreeMap<String, Vec<u8>>,
}

/// Stages the package directory. `package_digest` is the single source of the digest
/// (and of the SKILL.md/symlink/size rules); the walk below mirrors its limits to also
/// carry the bytes.
pub fn stage_package(src: &Path) -> Result<StagedPackage> {
    let digest = crate::package_digest(src)?;
    let mut files = BTreeMap::new();
    let mut entries = 0usize;
    let mut total = 0usize;
    walk_files(src, src, &mut files, &mut entries, &mut total, 0)?;
    Ok(StagedPackage { digest, files })
}

fn walk_files(
    root: &Path,
    dir: &Path,
    files: &mut BTreeMap<String, Vec<u8>>,
    entries: &mut usize,
    total: &mut usize,
    depth: usize,
) -> Result<()> {
    if depth > WALK_DEPTH {
        return Err("package nesting limit exceeded".into());
    }
    for entry in fs::read_dir(dir).map_err(|_| "package directory unavailable")? {
        *entries += 1;
        if *entries > WALK_ENTRIES {
            return Err("package entry limit exceeded".into());
        }
        let path = entry.map_err(|_| "package entry unavailable")?.path();
        let meta = fs::symlink_metadata(&path).map_err(|_| "package metadata unavailable")?;
        if meta.file_type().is_symlink() {
            return Err("package symlink refused".into());
        }
        if meta.is_dir() {
            walk_files(root, &path, files, entries, total, depth + 1)?;
        } else if meta.is_file() {
            if meta.len() > FILE_LIMIT {
                return Err("package file exceeds the size limit".into());
            }
            let rel = path
                .strip_prefix(root)
                .map_err(|_| "package path invalid")?
                .to_str()
                .ok_or("non UTF-8 package path")?
                .replace('\\', "/");
            if !relative_ok(&rel) {
                return Err("nonportable package path".into());
            }
            let bytes = fs::read(&path).map_err(|_| "package read failed")?;
            *total += bytes.len();
            if *total > WALK_BYTES {
                return Err("package too large".into());
            }
            files.insert(rel, bytes);
        } else {
            return Err("package special file refused".into());
        }
        if files.len() > WALK_FILES {
            return Err("package file limit exceeded".into());
        }
    }
    Ok(())
}

// --- plan types ---------------------------------------------------------------------

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct FilePlan {
    pub host: String,
    pub path: String,
    /// create | update | unchanged | conflict | preserved | remove_blocked
    pub action: String,
    pub baseline_sha256: Option<String>,
    pub current_sha256: Option<String>,
    pub new_sha256: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct HostPlan {
    pub host: String,
    pub projection_root: String,
    /// ok | blocked (REQ-04B-02: one blocked host never stops the others)
    pub preflight: String,
    pub blocked_reason: Option<String>,
    pub files: Vec<FilePlan>,
    /// #2483: offered only for a host blocked by an unmanaged existing package (a
    /// clean scan, nothing owned). Left out otherwise, so other plans keep their bytes.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub backup_then_replace: Option<BackupThenReplace>,
}

/// One file of the folder a backup-then-replace moves (#2483).
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct BackupFile {
    pub path: String,
    pub sha256: String,
    /// Permission bits (`mode & 0o7777`); 0 where the platform has none.
    pub mode: u32,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReplaceWith {
    pub package_digest: String,
    pub file_count: usize,
}

/// The backup-then-replace preview of one host (#2483, decision row 22). The original
/// folder is moved whole into `<backup_root>/<apply UTC>-<host>/moai` (a same-filesystem
/// rename: no byte is rewritten, nothing is deleted) beside a `MANIFEST.json` of
/// `source_files`, and the package is created in its place. It runs only when
/// `update-apply` carries `confirm_token`, which is bound to `source_tree_sha256`.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct BackupThenReplace {
    pub source_root: String,
    pub source_files: Vec<BackupFile>,
    pub source_file_count: usize,
    pub source_tree_sha256: String,
    pub backup_root: String,
    pub backup_layout: String,
    pub replace_with: ReplaceWith,
    pub confirm_token: String,
    pub apply_flag: String,
}

/// The backup a backup-then-replace receipt points at (#2483).
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ReplaceBackup {
    pub dir: String,
    pub manifest_sha256: String,
    pub source_tree_sha256: String,
    pub file_count: usize,
}

const BACKUP_MANIFEST: &str = "MANIFEST.json";
/// The moved original folder inside a backup directory.
pub(crate) const BACKUP_TREE: &str = "moai";

fn file_mode(meta: &fs::Metadata) -> u32 {
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        meta.permissions().mode() & 0o7777
    }
    #[cfg(not(unix))]
    {
        let _ = meta;
        0
    }
}

/// Every regular file under `root` with its sha256 and mode, sorted by path. Refuses
/// symlinks and special files; bounded like the projection scan.
pub(crate) fn source_files(root: &Path) -> Result<Vec<BackupFile>> {
    fn walk(
        root: &Path,
        dir: &Path,
        out: &mut Vec<BackupFile>,
        total: &mut usize,
        depth: usize,
    ) -> Result<()> {
        if depth > WALK_DEPTH {
            return Err("backup source nesting limit exceeded".into());
        }
        for entry in fs::read_dir(dir).map_err(|_| "backup source directory unavailable")? {
            let path = entry.map_err(|_| "backup source entry unavailable")?.path();
            let meta =
                fs::symlink_metadata(&path).map_err(|_| "backup source metadata unavailable")?;
            if meta.file_type().is_symlink() {
                return Err("backup source contains a symlink; refused".into());
            }
            if meta.is_dir() {
                walk(root, &path, out, total, depth + 1)?;
            } else if meta.is_file() {
                if meta.len() > FILE_LIMIT {
                    return Err("backup source file exceeds the size limit".into());
                }
                let rel = path
                    .strip_prefix(root)
                    .map_err(|_| "backup source path invalid")?
                    .to_str()
                    .ok_or("non UTF-8 backup source path")?
                    .replace('\\', "/");
                let bytes = fs::read(&path).map_err(|_| "backup source read failed")?;
                *total += bytes.len();
                if *total > WALK_BYTES || out.len() >= WALK_FILES {
                    return Err("backup source too large".into());
                }
                out.push(BackupFile {
                    path: rel,
                    sha256: sha256_hex(&bytes),
                    mode: file_mode(&meta),
                });
            } else {
                return Err("backup source special file refused".into());
            }
        }
        Ok(())
    }
    let meta = fs::symlink_metadata(root).map_err(|_| "backup source unavailable")?;
    if meta.file_type().is_symlink() || !meta.is_dir() {
        return Err("backup source is not a plain directory".into());
    }
    let mut out = Vec::new();
    walk(root, root, &mut out, &mut 0, 0)?;
    out.sort_by(|a, b| a.path.cmp(&b.path));
    Ok(out)
}

/// One digest over the sorted (path, sha256, mode) list.
pub(crate) fn source_tree_sha256(files: &[BackupFile]) -> String {
    let mut text = String::new();
    for f in files {
        text.push_str(&format!("{}\0{}\0{:o}\n", f.path, f.sha256, f.mode));
    }
    sha256_hex(text.as_bytes())
}

/// Test seam (#2483 AC-4): `MOAI_CLI_TEST_FAULT=<point>` fails the backup-then-replace
/// at that point. Read only in debug builds; a release build never consults it.
pub(crate) fn test_fault(point: &str) -> Result<()> {
    #[cfg(debug_assertions)]
    if std::env::var("MOAI_CLI_TEST_FAULT").as_deref() == Ok(point) {
        return Err(format!("injected test fault at {point}"));
    }
    let _ = point;
    Ok(())
}

/// Test seam (#2483 AC-5): `MOAI_CLI_TEST_FAULT=corrupt_after_move` appends one byte to
/// the first moved file, standing in for a writer that changes the folder between the
/// move and the manifest check. Debug builds only.
fn test_corrupt_moved(moved: &Path, files: &[BackupFile]) {
    #[cfg(debug_assertions)]
    if std::env::var("MOAI_CLI_TEST_FAULT").as_deref() == Ok("corrupt_after_move") {
        if let Some(first) = files.first() {
            if let Ok(mut f) = fs::OpenOptions::new()
                .append(true)
                .open(moved.join(&first.path))
            {
                use std::io::Write as _;
                let _ = f.write_all(b"x");
            }
        }
    }
    let _ = (moved, files);
}

/// Refuses a symlink in any existing component between `~/.moai-cli` and `dir`.
pub(crate) fn refuse_state_symlinks(dir: &Path, what: &str) -> Result<()> {
    let mut cursor = dir;
    loop {
        if let Ok(m) = fs::symlink_metadata(cursor) {
            if m.file_type().is_symlink() {
                return Err(format!("{what} path is a symlink; refused"));
            }
        }
        if cursor.file_name().map(|n| n == crate::HOME_STATE_DIR) == Some(true) {
            return Ok(());
        }
        cursor = match cursor.parent() {
            Some(p) => p,
            None => return Ok(()),
        };
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct ProtocolRange {
    pub min: u32,
    pub max: u32,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct FileHash {
    pub host: String,
    pub path: String,
    pub sha256: String,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct UpdatePlan {
    pub schema_version: u32,
    pub protocol_version: u32,
    pub protocol_range: ProtocolRange,
    /// Canonical home; binds the plan so a replay under another HOME is refused.
    pub home: String,
    pub scope: Vec<String>,
    pub source_version: Option<String>,
    pub target_version: String,
    pub package_digest: String,
    pub hosts: Vec<HostPlan>,
    pub expected_file_hashes: Vec<FileHash>,
    pub rollback_manifest: Vec<FileHash>,
    pub plan_digest: String,
}

// --- host receipt -------------------------------------------------------------------

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct HostReceiptEntry {
    pub path: String,
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: Option<String>,
    pub published: bool,
    pub error: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct HostReceipt {
    pub schema_version: u32,
    pub install_id: String,
    pub host: String,
    pub operation: String,
    pub plan_digest: String,
    pub package_digest: String,
    pub package_version: String,
    /// The previous receipt's `package_version` — the version this apply replaces
    /// (REQ-04C-06 recovery input, #2089). Absent on a first apply. `default` keeps
    /// pre-#2089 receipts readable.
    #[serde(default)]
    pub version_before: Option<String>,
    /// pending | committed (REQ-04B-04/05: a failed apply stays pending) | recovered
    /// (04-C #2000: a pending host whose changes were all reverted; read like committed)
    pub status: String,
    pub started_at: String,
    pub finished_at: Option<String>,
    /// This apply's file actions (before/after per file).
    pub entries: Vec<HostReceiptEntry>,
    /// Ownership manifest: every file this host owns and its sha at this receipt's
    /// commit point. The plan's baseline comes from here (REQ-04B-01). A pending
    /// receipt keeps the PRE-apply manifest, so a crashed apply surfaces as per-file
    /// conflict instead of a silent overwrite (04-C recovers).
    pub owned: Vec<FileHash>,
    /// #2268 additive (REQ-3): whether the projection root already existed before this
    /// apply. On a first apply (`version_before` none) into a pre-existing root, the
    /// created entries landed inside a previously unmanaged folder — recovery lists
    /// them. serde-default keeps pre-#2268 receipts readable.
    #[serde(default)]
    pub top_skills_existed_before: bool,
    /// #2483: set by a backup-then-replace apply; recovery restores from it.
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub backup: Option<ReplaceBackup>,
}

/// Receipt path; `create` also creates the directory chain (apply only — the plan must
/// not write anything, so it passes `false`).
/// The skill namespace keeps pre-rename receipts from owning the new projection.
pub(crate) fn host_receipt_path(
    home: &Path,
    install_id: &str,
    host: &str,
    create: bool,
) -> Result<PathBuf> {
    let dir = home
        .join(crate::HOME_STATE_DIR)
        .join("installations")
        .join(install_id)
        .join("hosts")
        .join(SKILL_DIR);
    if create {
        // Refuse a symlink in any component below ~/.moai-cli before creating.
        let mut cursor = dir.as_path();
        let mut parts: Vec<PathBuf> = Vec::new();
        while let Some(parent) = cursor.parent() {
            parts.push(cursor.to_path_buf());
            cursor = parent;
            if cursor.file_name().map(|n| n == crate::HOME_STATE_DIR) == Some(true) {
                break;
            }
        }
        for part in parts.iter().rev() {
            if let Ok(m) = fs::symlink_metadata(part) {
                if m.file_type().is_symlink() {
                    return Err("installations receipt path is a symlink; refused".into());
                }
            }
        }
        fs::create_dir_all(&dir).map_err(|_| "host receipt directory unavailable")?;
    }
    Ok(dir.join(format!("{host}.json")))
}

pub(crate) fn load_host_receipt(path: &Path) -> Result<Option<HostReceipt>> {
    match fs::symlink_metadata(path) {
        Ok(m) if m.file_type().is_symlink() => Err("host receipt is a symlink; refused".into()),
        Ok(_) => {
            let meta = fs::metadata(path).map_err(|_| "host receipt metadata unavailable")?;
            if !meta.is_file() || meta.len() > FILE_LIMIT {
                return Err("host receipt is not a bounded regular file".into());
            }
            let bytes = fs::read(path).map_err(|_| "host receipt read failed")?;
            let receipt: HostReceipt = serde_json::from_slice(&bytes)
                .map_err(|_| "invalid host receipt; no values were logged".to_string())?;
            Ok(Some(receipt))
        }
        Err(_) => Ok(None),
    }
}

/// Writes the receipt through the 01-B publish engine (temp + fsync + rename + parent
/// directory fsync), REQ-04B-04(a).
pub(crate) fn save_host_receipt(
    home: &Path,
    install_id: &str,
    receipt: &HostReceipt,
) -> Result<()> {
    let path = host_receipt_path(home, install_id, &receipt.host, true)?;
    let bytes =
        serde_json::to_vec_pretty(receipt).map_err(|_| "receipt encode failed".to_string())?;
    let before = read_opt(&path)?;
    publish_replace(&path, before.as_deref(), &bytes)
}

/// REQ-04C-03 (#2089): preserve every file this host will overwrite as a known-good
/// copy BEFORE the apply transaction starts (the pending receipt included) — a copy
/// failure starts nothing for the host. Copies live at
/// `~/.moai-cli/installations/<id>/hosts/moai/<host>/known-good/<hex>`, where `<hex>` is the
/// bare-hex sha256 of the exact before bytes — a filename-valid form like `install_id`;
/// the receipt's `before_sha256` carries the same digest with its `sha256:` prefix, so
/// recovery derives the path by stripping that prefix. `create` actions have no before
/// bytes and get no copy. Exclusive create: an existing regular file under the same
/// digest name is reused only after verifying it holds that digest's bytes — a
/// mismatched copy (a partial write from a failed save) is replaced with the true
/// bytes, and every fresh copy is read back and verified. No retention or GC: copies
/// accumulate for the life of the installation.
// @MX:DEBT: known-good copies accumulate with no retention or GC
// @MX:CEILING: RC1 fixture scale — a handful of hosts and files per installation
// @MX:UPGRADE: the real installer ships (P12 release matrix) — add retention/GC then
fn known_good_copies(
    home: &Path,
    install_id: &str,
    host: &str,
    root: &Path,
    writes: &[&FilePlan],
) -> Result<()> {
    for w in writes {
        if w.action != "update" {
            continue;
        }
        let target = root.join(&w.path);
        let bytes = read_opt(&target)?.ok_or("known-good: update target missing")?;
        let hex = format!("{:x}", Sha256::digest(&bytes));
        let dir = home
            .join(crate::HOME_STATE_DIR)
            .join("installations")
            .join(install_id)
            .join("hosts")
            .join(SKILL_DIR)
            .join(host)
            .join("known-good");
        // Refuse a symlink in any component below ~/.moai-cli before creating (same
        // hygiene as the receipt path).
        let mut cursor = dir.as_path();
        let mut parts: Vec<PathBuf> = Vec::new();
        while let Some(parent) = cursor.parent() {
            parts.push(cursor.to_path_buf());
            cursor = parent;
            if cursor.file_name().map(|n| n == crate::HOME_STATE_DIR) == Some(true) {
                break;
            }
        }
        for part in parts.iter().rev() {
            if let Ok(m) = fs::symlink_metadata(part) {
                if m.file_type().is_symlink() {
                    return Err("installations known-good path is a symlink; refused".into());
                }
            }
        }
        fs::create_dir_all(&dir).map_err(|_| "known-good directory unavailable")?;
        let path = dir.join(&hex);
        if fs::symlink_metadata(&path).is_ok() {
            // A file under this digest name already exists: reuse only a regular
            // file that actually holds its digest's bytes — a partial write from a
            // failed earlier save must not survive as a "known-good" copy. A
            // directory or symlink planted at the path is a refusal, not a reuse;
            // a mismatched regular file is replaced with the true bytes.
            let m = fs::symlink_metadata(&path).map_err(|_| "known-good copy state unavailable")?;
            if !m.is_file() {
                return Err("known-good digest path is not a regular file".into());
            }
            let existing = fs::read(&path).map_err(|_| "known-good copy read failed")?;
            if format!("{:x}", Sha256::digest(&existing)) == hex {
                continue;
            }
            // 01-B temp+rename (REQ-01B-03): stage() re-verifies the on-disk bytes
            // right before the rename, and the rename swaps the inode, so a hard
            // link sharing the corrupt copy keeps its own bytes — an in-place
            // truncate would corrupt every link (#2103).
            publish_replace(&path, Some(&existing), &bytes)?;
        } else {
            // Exclusive create through the same engine (REQ-01B-04): the temp file
            // is hard-linked into place and the publish fails explicitly if the
            // path appeared meanwhile.
            publish_exclusive(&path, &bytes)?;
        }
        // Read-back verification: what is on disk must be the digest the name
        // promises (save-time digest re-verification).
        let stored = fs::read(&path).map_err(|_| "known-good copy read failed")?;
        if format!("{:x}", Sha256::digest(&stored)) != hex {
            return Err("known-good copy does not hold its digest's bytes".into());
        }
    }
    Ok(())
}

fn owned_map(receipt: Option<&HostReceipt>) -> BTreeMap<String, String> {
    let mut map = BTreeMap::new();
    if let Some(r) = receipt {
        for f in &r.owned {
            map.insert(f.path.clone(), f.sha256.clone());
        }
    }
    map
}

fn owned_vec(host: &str, map: &BTreeMap<String, String>) -> Vec<FileHash> {
    map.iter()
        .map(|(p, s)| FileHash {
            host: host.to_string(),
            path: p.clone(),
            sha256: s.clone(),
        })
        .collect()
}

// --- inventory / preflight ----------------------------------------------------------

struct ProjectionScan {
    files: BTreeMap<String, String>, // rel -> sha256
}

/// Reads the projection tree (shas only). Only called on an existing root: any
/// structural hazard (symlink, special file, oversized/unreadable file) becomes an
/// `Err` that the caller turns into the host's `blocked` reason (REQ-04B-02).
fn scan_projection(root: &Path) -> std::result::Result<ProjectionScan, String> {
    let meta =
        fs::symlink_metadata(root).map_err(|_| "projection directory unavailable".to_string())?;
    if meta.file_type().is_symlink() {
        return Err("projection path is a symlink; refused".into());
    }
    if !meta.is_dir() {
        return Err("projection path is not a directory".into());
    }
    let mut files = BTreeMap::new();
    let mut entries = 0usize;
    let mut total = 0usize;
    scan_walk(root, root, &mut files, &mut entries, &mut total, 0)?;
    Ok(ProjectionScan { files })
}

fn scan_walk(
    root: &Path,
    dir: &Path,
    files: &mut BTreeMap<String, String>,
    entries: &mut usize,
    total: &mut usize,
    depth: usize,
) -> std::result::Result<(), String> {
    if depth > WALK_DEPTH {
        return Err("projection nesting limit exceeded".into());
    }
    for entry in fs::read_dir(dir).map_err(|_| "projection directory unavailable")? {
        *entries += 1;
        if *entries > WALK_ENTRIES {
            return Err("projection entry limit exceeded".into());
        }
        let path = entry.map_err(|_| "projection entry unavailable")?.path();
        let meta = fs::symlink_metadata(&path).map_err(|_| "projection metadata unavailable")?;
        if meta.file_type().is_symlink() {
            return Err("projection contains a symlink; refused".into());
        }
        if meta.is_dir() {
            scan_walk(root, &path, files, entries, total, depth + 1)?;
        } else if meta.is_file() {
            if meta.len() > FILE_LIMIT {
                return Err("projection file exceeds the size limit".into());
            }
            let rel = path
                .strip_prefix(root)
                .map_err(|_| "projection path invalid")?
                .to_str()
                .ok_or("non UTF-8 projection path")?
                .replace('\\', "/");
            let bytes = fs::read(&path).map_err(|_| "projection read failed")?;
            *total += bytes.len();
            if *total > WALK_BYTES {
                return Err("projection too large".into());
            }
            files.insert(rel, sha256_hex(&bytes));
        } else {
            return Err("projection special file refused".into());
        }
        if files.len() > WALK_FILES {
            return Err("projection file limit exceeded".into());
        }
    }
    Ok(())
}

/// Refuses a symlink in any existing component between the home and the projection root.
fn ancestor_symlinks(home: &Path, root: &Path) -> Result<()> {
    let mut cursor = root.to_path_buf();
    while cursor.starts_with(home) && cursor != home {
        match fs::symlink_metadata(&cursor) {
            Ok(m) if m.file_type().is_symlink() => {
                return Err("projection path is a symlink; refused".into())
            }
            Ok(_) => (),
            Err(_) => (), // missing components are created at apply time
        }
        cursor = cursor
            .parent()
            .ok_or("projection path invalid")?
            .to_path_buf();
    }
    Ok(())
}

/// Cross-filesystem projection (unix device ids): receipt and files would live on
/// different filesystems, weakening the crash ordering. Non-unix platforms skip the
/// check and make no claim.
#[cfg(unix)]
fn cross_filesystem(home: &Path, root: &Path) -> Result<bool> {
    use std::os::unix::fs::MetadataExt;
    let state = home.join(crate::HOME_STATE_DIR);
    match (fs::metadata(&state), fs::metadata(root)) {
        (Ok(a), Ok(b)) => Ok(a.dev() != b.dev()),
        _ => Ok(false),
    }
}

#[cfg(not(unix))]
fn cross_filesystem(_home: &Path, _root: &Path) -> Result<bool> {
    Ok(false)
}

fn writable(path: &Path) -> bool {
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::metadata(path)
            .map(|m| m.permissions().mode() & 0o200 != 0)
            .unwrap_or(false)
    }
    #[cfg(not(unix))]
    {
        fs::metadata(path).is_ok()
    }
}

/// Runs the host preflight (REQ-04B-02) and, when safe, the projection scan in one
/// pass. Returns the `blocked` reason (if any) and the scan (an absent projection is
/// an empty scan, not a block). Every hazard becomes a reason; nothing escapes as a
/// global error.
fn preflight(home: &Path, host: &str) -> (Option<String>, Option<ProjectionScan>) {
    let root = projection_root(home, host);
    if let Err(e) = ancestor_symlinks(home, &root) {
        return (Some(e), None);
    }
    let meta = match fs::symlink_metadata(&root) {
        Err(_) => {
            // Absent projection: apply may create it, so the nearest existing ancestor
            // must be writable.
            let mut cursor = match root.parent() {
                Some(p) => p.to_path_buf(),
                None => return (Some("projection path invalid".into()), None),
            };
            loop {
                if fs::symlink_metadata(&cursor).is_ok() {
                    if !writable(&cursor) {
                        return (
                            Some("skills parent directory is not writable; refused".into()),
                            None,
                        );
                    }
                    break;
                }
                match cursor.parent() {
                    Some(p) => cursor = p.to_path_buf(),
                    None => break,
                }
            }
            return (
                None,
                Some(ProjectionScan {
                    files: BTreeMap::new(),
                }),
            );
        }
        Ok(m) => m,
    };
    if meta.file_type().is_symlink() {
        return (Some("projection path is a symlink; refused".into()), None);
    }
    let scan = match scan_projection(&root) {
        Ok(s) => s,
        Err(e) => return (Some(e), None),
    };
    if !writable(&root) {
        return (
            Some("projection directory is not writable; refused".into()),
            None,
        );
    }
    if cross_filesystem(home, &root).unwrap_or(false) {
        return (
            Some("cross-filesystem projection; receipt and files must share one filesystem".into()),
            None,
        );
    }
    (None, Some(scan))
}

// --- plan ---------------------------------------------------------------------------

/// Computes the current installation id through 04-A's public API (no 04-A changes).
pub(crate) fn current_install_id(home: &Path) -> Result<String> {
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable".to_string())?;
    Ok(plan_receipt(&exe, home, InstallerOwner::Unknown)?.install_id)
}

/// Read-only three-way plan over every host (REQ-04B-01/02/03). Writes nothing.
pub fn plan_update(home: &Path, staged: &StagedPackage) -> Result<UpdatePlan> {
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let install_id = current_install_id(&canonical_home)?;
    let mut hosts = Vec::new();
    let mut expected = Vec::new();
    let mut rollback = Vec::new();
    let mut latest_version: Option<String> = None;

    for host in HOSTS {
        let root = projection_root(&canonical_home, host);
        let root_text = root
            .to_str()
            .ok_or("non UTF-8 projection path")?
            .to_string();
        let (blocked, scan) = preflight(&canonical_home, host);
        let receipt = if blocked.is_some() {
            None
        } else {
            let path = host_receipt_path(&canonical_home, &install_id, host, false)?;
            load_host_receipt(&path)?
        };
        // REQ-04C-04 (04-C #2000): a pending host takes no new apply until recovered; the
        // other hosts plan as usual.
        let (blocked, scan) = match &receipt {
            Some(r) if r.status == "pending" => (
                Some("host receipt is pending; run recovery-plan first".to_string()),
                None,
            ),
            _ => (blocked, scan),
        };
        if receipt.is_some() && latest_version.is_none() {
            latest_version = receipt.as_ref().map(|r| r.package_version.clone());
        }
        let owned = owned_map(receipt.as_ref());
        let mut files = Vec::new();

        // Package files vs disk, three-way per file (REQ-04B-01).
        for (rel, new_bytes) in &staged.files {
            let new_sha = sha256_hex(new_bytes);
            let baseline = owned.get(rel).cloned();
            let current = scan.as_ref().and_then(|s| s.files.get(rel).cloned());
            let action = match decide(
                baseline.as_deref(),
                current.as_deref(),
                Some(new_sha.as_str()),
            ) {
                Decision::KeepCurrent => {
                    if current.as_deref() == Some(new_sha.as_str()) {
                        "unchanged"
                    } else {
                        // desired == baseline: the template did not change; keep the
                        // user's edit (AC-04B-03).
                        "preserved"
                    }
                }
                Decision::Apply => {
                    if current.is_none() {
                        "create"
                    } else {
                        "update"
                    }
                }
                Decision::Conflict => "conflict",
            };
            if action == "create" || action == "update" {
                expected.push(FileHash {
                    host: (*host).to_string(),
                    path: rel.clone(),
                    sha256: new_sha.clone(),
                });
                if let Some(before) = &current {
                    rollback.push(FileHash {
                        host: (*host).to_string(),
                        path: rel.clone(),
                        sha256: before.clone(),
                    });
                }
            }
            files.push(FilePlan {
                host: (*host).to_string(),
                path: rel.clone(),
                action: action.to_string(),
                baseline_sha256: baseline,
                current_sha256: current,
                new_sha256: Some(new_sha),
            });
        }

        // Disk files the package does not carry (REQ-04B-09): never deleted, reported.
        if let Some(scan) = &scan {
            for (rel, current_sha) in &scan.files {
                if staged.files.contains_key(rel) {
                    continue;
                }
                let owned_here = owned.contains_key(rel);
                files.push(FilePlan {
                    host: (*host).to_string(),
                    path: rel.clone(),
                    action: if owned_here {
                        // We applied it before and the new package dropped it: removal
                        // would be needed, but RC1 never deletes (04-C decides).
                        "remove_blocked"
                    } else {
                        "preserved"
                    }
                    .to_string(),
                    baseline_sha256: owned.get(rel).cloned(),
                    current_sha256: Some(current_sha.clone()),
                    new_sha256: None,
                });
            }
        }

        files.sort_by(|a, b| a.path.cmp(&b.path));

        // REQ-1 (#2268, fail-closed): pre-existing content the install receipt does not
        // own means the projection root holds a foreign or previous-generation package
        // (the agent-9 repro: an old-generation SKILL.md with 12 pending RC1 creates
        // around it). The host is blocked — apply writes nothing for it. A managed
        // root (every staged file owned) keeps the normal three-way path (AC-2).
        // REQ-2: the reason names the condition ('unmanaged existing package') and the
        // safe alternatives — never an automatic overwrite.
        let unmanaged: Vec<&str> = files
            .iter()
            .filter(|f| {
                f.baseline_sha256.is_none()
                    && f.current_sha256.is_some()
                    && staged.files.contains_key(&f.path)
            })
            .map(|f| f.path.as_str())
            .collect();
        // #2483: a host blocked ONLY by the unmanaged package (clean scan, no other
        // reason, nothing owned) gets the backup-then-replace preview.
        let backup_then_replace = if !unmanaged.is_empty() && blocked.is_none() && owned.is_empty()
        {
            source_files(&root).ok().map(|source| {
                let tree = source_tree_sha256(&source);
                let hex = tree.strip_prefix("sha256:").unwrap_or(&tree);
                let token = format!("backup-replace:{host}:{}", &hex[..16]);
                BackupThenReplace {
                    source_root: root_text.clone(),
                    source_file_count: source.len(),
                    source_files: source,
                    source_tree_sha256: tree.clone(),
                    backup_root: canonical_home
                        .join(crate::HOME_STATE_DIR)
                        .join("backups")
                        .to_string_lossy()
                        .into_owned(),
                    backup_layout: format!(
                        "<apply UTC>-{host}/{BACKUP_TREE} (the folder, moved as is) and \
                         <apply UTC>-{host}/{BACKUP_MANIFEST}"
                    ),
                    replace_with: ReplaceWith {
                        package_digest: staged.digest.clone(),
                        file_count: staged.files.len(),
                    },
                    apply_flag: format!("--confirm-backup-replace {token}"),
                    confirm_token: token,
                }
            })
        } else {
            None
        };
        let (preflight, blocked) = if !unmanaged.is_empty() {
            (
                "blocked".to_string(),
                Some(format!(
                    "unmanaged existing package: {} file(s) present in the projection root \
                     that install receipt {install_id} does not own ({}); \
                     safe alternatives: install to a separate location, or a \
                     user-confirmed backup-then-replace — never automatic; \
                     per-file actions below are informational only; {GO_ERA_GUIDANCE}",
                    unmanaged.len(),
                    unmanaged.join(", "),
                )),
            )
        } else {
            let reason = blocked.clone();
            (
                if reason.is_some() { "blocked" } else { "ok" }.into(),
                reason,
            )
        };
        hosts.push(HostPlan {
            host: (*host).to_string(),
            projection_root: root_text,
            preflight,
            blocked_reason: match (&backup_then_replace, blocked) {
                (Some(b), Some(reason)) => Some(format!(
                    "{reason}; this host offers a backup-then-replace preview \
                     (backup_then_replace): confirm with {}",
                    b.apply_flag
                )),
                (_, reason) => reason,
            },
            files,
            backup_then_replace,
        });
    }

    let plan = UpdatePlan {
        schema_version: UPDATE_SCHEMA_VERSION,
        protocol_version: PROTOCOL_VERSION,
        protocol_range: ProtocolRange {
            min: PROTOCOL_VERSION,
            max: PROTOCOL_VERSION,
        },
        home: canonical_home
            .to_str()
            .ok_or("non UTF-8 home path")?
            .to_string(),
        scope: HOSTS.iter().map(|h| (*h).to_string()).collect(),
        source_version: latest_version,
        target_version: env!("CARGO_PKG_VERSION").to_string(),
        package_digest: staged.digest.clone(),
        hosts,
        expected_file_hashes: expected,
        rollback_manifest: rollback,
        plan_digest: String::new(),
    };
    let mut core = plan.clone();
    core.plan_digest = String::new();
    let mut plan = plan;
    plan.plan_digest = crate::digest(&serde_json::to_vec(&core).map_err(|_| "encode failed")?);
    Ok(plan)
}

// --- apply --------------------------------------------------------------------------

fn validate_permission(permission: &str) -> Result<()> {
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    Ok(())
}

/// One host's apply outcome.
struct HostOutcome {
    host: String,
    status: String, // applied | unchanged | conflict | blocked | failed
    files: Vec<serde_json::Value>,
    receipt_status: Option<String>,
    reload_required: String, // no | yes | unknown
    reload_basis: String,
    blocked_reason: Option<String>,
    /// #2483: the backup directory of a backup-then-replace.
    backup_dir: Option<String>,
}

/// Applies the plan host by host (REQ-04B-03..06). Refuses anything but a freshly
/// recomputed identical digest before touching a single file. `confirm_backup_replace`
/// carries the tokens of the hosts whose backup-then-replace the user confirmed
/// (#2483); a token the plan did not offer is refused before any write.
pub fn apply_update(
    home: &Path,
    staged: &StagedPackage,
    plan: &UpdatePlan,
    plan_digest: &str,
    permission_reference: &str,
    confirm_backup_replace: &[String],
) -> Result<serde_json::Value> {
    validate_permission(permission_reference)?;
    let fresh = plan_update(home, staged)?;
    // The whole plan BODY must equal a fresh recompute (a tampered body with a stolen
    // digest field is still refused), and the caller's digest must bind both.
    if plan.schema_version != UPDATE_SCHEMA_VERSION
        || plan.protocol_version != PROTOCOL_VERSION
        || *plan != fresh
        || fresh.plan_digest != plan_digest
        || plan.plan_digest != plan_digest
    {
        return Err("stale or altered plan; re-preview required".into());
    }
    for token in confirm_backup_replace {
        let offered = fresh.hosts.iter().any(|h| {
            h.backup_then_replace
                .as_ref()
                .is_some_and(|b| &b.confirm_token == token)
        });
        if !offered {
            return Err(
                "confirmation token does not match any backup-then-replace in the plan; nothing written"
                    .into(),
            );
        }
    }
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let install_id = current_install_id(&canonical_home)?;

    let mut outcomes: Vec<HostOutcome> = Vec::new();
    let mut published_anywhere = false;
    let mut any_failed = false;
    let mut any_conflict_only = false;

    for host_plan in &fresh.hosts {
        let host = host_plan.host.as_str();
        if host_plan.preflight == "blocked" {
            if let Some(btr) = host_plan
                .backup_then_replace
                .as_ref()
                .filter(|b| confirm_backup_replace.iter().any(|t| t == &b.confirm_token))
            {
                // #2483: the user confirmed this host's backup-then-replace.
                let o = backup_then_replace(
                    &canonical_home,
                    &install_id,
                    host,
                    staged,
                    &fresh,
                    plan_digest,
                    btr,
                )?;
                any_failed |= o.status == "failed";
                published_anywhere |= o.status == "replaced";
                outcomes.push(o);
                continue;
            }
            // REQ-04B-02: a blocked host is never written; the others go on.
            outcomes.push(HostOutcome {
                host: host.to_string(),
                status: "blocked".into(),
                files: Vec::new(),
                receipt_status: None,
                reload_required: "no".into(),
                reload_basis: "no files written".into(),
                blocked_reason: host_plan.blocked_reason.clone(),
                backup_dir: None,
            });
            continue;
        }

        let root = projection_root(&canonical_home, host);
        let writes: Vec<&FilePlan> = host_plan
            .files
            .iter()
            .filter(|f| f.action == "create" || f.action == "update")
            .collect();
        let conflict_only = host_plan.files.iter().any(|f| f.action == "conflict");

        if writes.is_empty() {
            // Nothing to write: report, never touch (REQ-04B-07 rerun; conflicts keep
            // the user bytes, REQ-04B-09).
            let status = if conflict_only {
                "conflict"
            } else {
                "unchanged"
            };
            any_conflict_only |= conflict_only;
            let files = host_plan
                .files
                .iter()
                .map(|f| file_result(f, false, None))
                .collect();
            outcomes.push(HostOutcome {
                host: host.to_string(),
                status: status.into(),
                files,
                receipt_status: None,
                reload_required: "no".into(),
                reload_basis: "no files written".into(),
                blocked_reason: None,
                backup_dir: None,
            });
            continue;
        }

        // REQ-04C-03 (#2089): known-good copies of every overwritten file, BEFORE the
        // transaction starts; a copy failure starts nothing for this host.
        if let Err(e) = known_good_copies(&canonical_home, &install_id, host, &root, &writes) {
            any_failed = true;
            outcomes.push(HostOutcome {
                host: host.to_string(),
                status: "failed".into(),
                files: Vec::new(),
                receipt_status: None,
                reload_required: "no".into(),
                reload_basis: format!("known-good copy failed; host not started: {e}"),
                blocked_reason: None,
                backup_dir: None,
            });
            continue;
        }

        // Transaction start: pending receipt BEFORE the first file write (REQ-04B-04a),
        // carrying the pre-apply ownership manifest.
        let receipt_path = host_receipt_path(&canonical_home, &install_id, host, true)?;
        let previous = load_host_receipt(&receipt_path)?;
        let pre_owned = owned_map(previous.as_ref());
        let top_skills_existed = fs::symlink_metadata(home.join(skills_dir(host))).is_ok();
        let mut receipt = HostReceipt {
            schema_version: crate::install::RECEIPT_SCHEMA_VERSION,
            install_id: install_id.clone(),
            host: host.to_string(),
            operation: "update.apply".into(),
            plan_digest: plan_digest.to_string(),
            package_digest: staged.digest.clone(),
            package_version: fresh.target_version.clone(),
            version_before: previous.as_ref().map(|p| p.package_version.clone()),
            top_skills_existed_before: top_skills_existed,
            status: "pending".into(),
            started_at: now_rfc3339(),
            finished_at: None,
            entries: writes
                .iter()
                .map(|f| HostReceiptEntry {
                    path: f.path.clone(),
                    action: f.action.clone(),
                    before_sha256: f.current_sha256.clone(),
                    after_sha256: f.new_sha256.clone(),
                    published: false,
                    error: None,
                })
                .collect(),
            owned: owned_vec(host, &pre_owned),
            backup: None,
        };
        save_host_receipt(&canonical_home, &install_id, &receipt)?;

        // (b) per file: 01-B engine (unique temp + fsync + rename / exclusive create),
        // each re-verifying the target right before publish and fsyncing its directory.
        let mut owned_now = pre_owned.clone();
        let mut host_published = 0usize;
        let mut failed: Option<String> = None;
        for (idx, w) in writes.iter().enumerate() {
            let target = root.join(&w.path);
            let new_bytes = staged
                .files
                .get(&w.path)
                .ok_or_else(|| "package file disappeared".to_string())?;
            // Re-verify the current bytes still match the plan (REQ-04B-03); an
            // unreadable target is THIS host's failure, not a global error.
            let before = match read_opt(&target) {
                Ok(b) => b,
                Err(e) => {
                    failed = Some(e);
                    receipt.entries[idx].error = failed.clone();
                    break;
                }
            };
            let before_sha = before.as_deref().map(sha256_hex);
            if before_sha != w.current_sha256 {
                failed = Some("target changed during apply; temporary file discarded".into());
                receipt.entries[idx].error = failed.clone();
                break;
            }
            let result = if w.action == "create" {
                publish_exclusive(&target, new_bytes)
            } else {
                publish_replace(&target, before.as_deref(), new_bytes)
            };
            match result {
                Ok(()) => {
                    let written = read_opt(&target)?;
                    let written_sha = written.as_deref().map(sha256_hex);
                    if written_sha.as_deref() != w.new_sha256.as_deref() {
                        failed = Some("published bytes do not match the plan".into());
                        receipt.entries[idx].error = failed.clone();
                        break;
                    }
                    receipt.entries[idx].published = true;
                    owned_now.insert(w.path.clone(), written_sha.expect("sha of a written file"));
                    host_published += 1;
                }
                Err(e) => {
                    failed = Some(e);
                    receipt.entries[idx].error = failed.clone();
                    break;
                }
            }
        }
        published_anywhere |= host_published > 0;

        if let Some(_err) = failed {
            // REQ-04B-05: keep the receipt pending WITH THE PRE-APPLY OWNERSHIP
            // manifest, report failed, never roll back another host (04-C recovers).
            any_failed = true;
            receipt.finished_at = Some(now_rfc3339());
            save_host_receipt(&canonical_home, &install_id, &receipt)?;
            let files = writes
                .iter()
                .zip(&receipt.entries)
                .map(|(f, e)| file_result_with(f, e))
                .collect();
            outcomes.push(HostOutcome {
                host: host.to_string(),
                status: "failed".into(),
                files,
                receipt_status: Some("pending".into()),
                reload_required: "unknown".into(),
                reload_basis: "apply failed mid-transaction; host state is pending (04-C recovery)"
                    .into(),
                blocked_reason: None,
                backup_dir: None,
            });
            continue;
        }

        // (c) every touched directory was fsynced by publish_replace/publish_exclusive;
        // (d) full re-verification, then committed (ownership manifest moves forward).
        for w in &writes {
            let target = root.join(&w.path);
            let written = read_opt(&target)?;
            if written.as_deref().map(sha256_hex).as_deref() != w.new_sha256.as_deref() {
                return Err("published bytes do not match the plan".into());
            }
        }
        receipt.owned = owned_vec(host, &owned_now);
        receipt.status = "committed".into();
        receipt.finished_at = Some(now_rfc3339());
        save_host_receipt(&canonical_home, &install_id, &receipt)?;

        let (reload, basis) = reload_state(host, top_skills_existed, host_published > 0);
        let mut all: Vec<serde_json::Value> = writes
            .iter()
            .zip(&receipt.entries)
            .map(|(f, e)| file_result_with(f, e))
            .collect();
        for f in host_plan
            .files
            .iter()
            .filter(|f| f.action != "create" && f.action != "update")
        {
            all.push(file_result(f, false, None));
        }
        outcomes.push(HostOutcome {
            host: host.to_string(),
            status: "applied".into(),
            files: all,
            receipt_status: Some("committed".into()),
            reload_required: reload.into(),
            reload_basis: basis,
            blocked_reason: None,
            backup_dir: None,
        });
    }

    let overall = if any_failed {
        "partial"
    } else if !outcomes.is_empty() && outcomes.iter().all(|o| o.status == "blocked") {
        "blocked"
    } else if published_anywhere {
        "applied"
    } else if any_conflict_only {
        "conflict"
    } else {
        "unchanged"
    };
    Ok(json!({
        "status": overall,
        "plan_digest": plan_digest,
        "package_digest": staged.digest,
        "package_version": fresh.target_version,
        "hosts": outcomes.iter().map(host_outcome_json).collect::<Vec<_>>(),
        "host_loaded": false,
        "execution_verified": "not_run",
        "reload_note": "host loading is not measured here (08-C owns the live check)",
        "writes": false,
    }))
}

/// UTC `YYYYMMDDTHHMMSSZ` for backup directory names.
fn compact_stamp() -> String {
    now_rfc3339().replace(['-', ':'], "")
}

/// #2483: one confirmed backup-then-replace. Order: fresh scan equal to the preview →
/// exclusive backup directory → `MANIFEST.json` → pending receipt (carrying `backup`) →
/// move the original folder into the backup (same-filesystem rename) → verify the moved
/// tree against the manifest → create the package files → verify → committed. Before the
/// move nothing in the projection changed; after it the original sits intact in the
/// backup — so at every point the original or the backup is whole, and a failure after
/// the pending receipt reads as `pending` in `recovery-plan`.
fn backup_then_replace(
    home: &Path,
    install_id: &str,
    host: &str,
    staged: &StagedPackage,
    plan: &UpdatePlan,
    plan_digest: &str,
    btr: &BackupThenReplace,
) -> Result<HostOutcome> {
    let mut receipt: Option<HostReceipt> = None;
    let mut backup_dir: Option<PathBuf> = None;
    let run = run_backup_then_replace(
        home,
        install_id,
        host,
        staged,
        plan,
        plan_digest,
        btr,
        &mut receipt,
        &mut backup_dir,
    );
    let dir_text = backup_dir
        .as_ref()
        .map(|d| d.to_string_lossy().into_owned());
    let created: Vec<serde_json::Value> = receipt
        .as_ref()
        .map(|r| {
            r.entries
                .iter()
                .map(|e| {
                    json!({"path": e.path, "action": e.action, "before_sha256": e.before_sha256,
                       "after_sha256": if e.published { e.after_sha256.clone() } else { None },
                       "published": e.published, "error": e.error})
                })
                .collect()
        })
        .unwrap_or_default();
    match run {
        Ok(()) => {
            let (reload, basis) = reload_state(host, true, true);
            Ok(HostOutcome {
                host: host.to_string(),
                status: "replaced".into(),
                files: created,
                receipt_status: Some("committed".into()),
                reload_required: reload.into(),
                reload_basis: basis,
                blocked_reason: None,
                backup_dir: dir_text,
            })
        }
        Err(e) => match receipt {
            None => Ok(HostOutcome {
                host: host.to_string(),
                status: "failed".into(),
                files: Vec::new(),
                receipt_status: None,
                reload_required: "no".into(),
                reload_basis: format!(
                    "backup-then-replace stopped before the original folder moved: {e}"
                ),
                blocked_reason: None,
                backup_dir: dir_text,
            }),
            Some(mut r) => {
                r.finished_at = Some(now_rfc3339());
                save_host_receipt(home, install_id, &r)?;
                Ok(HostOutcome {
                    host: host.to_string(),
                    status: "failed".into(),
                    files: created,
                    receipt_status: Some("pending".into()),
                    reload_required: "unknown".into(),
                    reload_basis: format!(
                        "backup-then-replace failed after the pending receipt: {e}; run recovery-plan"
                    ),
                    blocked_reason: None,
                    backup_dir: dir_text,
                })
            }
        },
    }
}

#[allow(clippy::too_many_arguments)]
fn run_backup_then_replace(
    home: &Path,
    install_id: &str,
    host: &str,
    staged: &StagedPackage,
    plan: &UpdatePlan,
    plan_digest: &str,
    btr: &BackupThenReplace,
    receipt_out: &mut Option<HostReceipt>,
    backup_out: &mut Option<PathBuf>,
) -> Result<()> {
    let root = projection_root(home, host);
    if source_files(&root)? != btr.source_files {
        return Err("projection changed since the plan; nothing moved".into());
    }
    let backups = home.join(crate::HOME_STATE_DIR).join("backups");
    refuse_state_symlinks(&backups, "backup")?;
    fs::create_dir_all(&backups).map_err(|_| "backup directory unavailable")?;
    refuse_state_symlinks(&backups, "backup")?;
    let dir = backups.join(format!("{}-{host}", compact_stamp()));
    fs::create_dir(&dir).map_err(|_| "backup directory already exists or is unavailable")?;
    *backup_out = Some(dir.clone());

    let manifest = json!({
        "schema_version": 1,
        "host": host,
        "source_root": btr.source_root,
        "source_tree_sha256": btr.source_tree_sha256,
        "plan_digest": plan_digest,
        "created_at": now_rfc3339(),
        "files": btr.source_files,
    });
    let manifest_bytes =
        serde_json::to_vec_pretty(&manifest).map_err(|_| "manifest encode failed".to_string())?;
    publish_exclusive(&dir.join(BACKUP_MANIFEST), &manifest_bytes)?;

    let previous = load_host_receipt(&host_receipt_path(home, install_id, host, false)?)?;
    let mut receipt = HostReceipt {
        schema_version: crate::install::RECEIPT_SCHEMA_VERSION,
        install_id: install_id.to_string(),
        host: host.to_string(),
        operation: "update.backup_then_replace".into(),
        plan_digest: plan_digest.to_string(),
        package_digest: staged.digest.clone(),
        package_version: plan.target_version.clone(),
        version_before: previous.as_ref().map(|p| p.package_version.clone()),
        top_skills_existed_before: true,
        status: "pending".into(),
        started_at: now_rfc3339(),
        finished_at: None,
        entries: staged
            .files
            .iter()
            .map(|(rel, bytes)| HostReceiptEntry {
                path: rel.clone(),
                action: "create".into(),
                before_sha256: None,
                after_sha256: Some(sha256_hex(bytes)),
                published: false,
                error: None,
            })
            .collect(),
        owned: Vec::new(),
        backup: Some(ReplaceBackup {
            dir: dir.to_string_lossy().into_owned(),
            manifest_sha256: sha256_hex(&manifest_bytes),
            source_tree_sha256: btr.source_tree_sha256.clone(),
            file_count: btr.source_files.len(),
        }),
    };
    save_host_receipt(home, install_id, &receipt)?;
    *receipt_out = Some(receipt.clone());
    test_fault("after_backup")?;

    let moved = dir.join(BACKUP_TREE);
    fs::rename(&root, &moved)
        .map_err(|_| "moving the original folder into the backup failed; nothing moved")?;
    test_corrupt_moved(&moved, &btr.source_files);
    if source_files(&moved).ok().as_ref() != Some(&btr.source_files) {
        let _ = fs::rename(&moved, &root);
        return Err("the moved folder does not match the manifest; moved back".into());
    }

    let mut owned = BTreeMap::new();
    for (idx, (rel, bytes)) in staged.files.iter().enumerate() {
        let target = root.join(rel);
        if let Err(e) = publish_exclusive(&target, bytes) {
            receipt.entries[idx].error = Some(e.clone());
            *receipt_out = Some(receipt.clone());
            return Err(e);
        }
        let written = read_opt(&target)?.map(|b| sha256_hex(&b));
        if written != receipt.entries[idx].after_sha256 {
            *receipt_out = Some(receipt.clone());
            return Err("published bytes do not match the package".into());
        }
        receipt.entries[idx].published = true;
        owned.insert(rel.clone(), written.expect("sha of a written file"));
        *receipt_out = Some(receipt.clone());
        if idx == 0 {
            test_fault("mid_replace")?;
        }
    }
    receipt.owned = owned_vec(host, &owned);
    receipt.status = "committed".into();
    receipt.finished_at = Some(now_rfc3339());
    save_host_receipt(home, install_id, &receipt)?;
    *receipt_out = Some(receipt);
    Ok(())
}

/// REQ-04B-06: reload state and its documented basis per host.
fn reload_state(
    host: &str,
    top_skills_existed_before: bool,
    wrote: bool,
) -> (&'static str, String) {
    match host {
        "claude" => {
            if wrote && !top_skills_existed_before {
                (
                    "yes",
                    "the top-level skills directory was created; running sessions need /reload-skills (code.claude.com/docs/en/skills, 2026-11-08)".into(),
                )
            } else if wrote {
                (
                    "no",
                    "Claude reflects skill file changes live (code.claude.com/docs/en/skills, 2026-11-08)".into(),
                )
            } else {
                ("no", "no files written".into())
            }
        }
        "codex" => {
            if wrote {
                (
                    "unknown",
                    "Codex detects skill changes automatically; restart if they are not shown (learn.chatgpt.com/docs/build-skills, 2026-11-08)".into(),
                )
            } else {
                ("no", "no files written".into())
            }
        }
        other => unreachable!("unknown host {other}"),
    }
}

fn file_result(f: &FilePlan, published: bool, error: Option<String>) -> serde_json::Value {
    json!({
        "path": f.path,
        "action": f.action,
        "before_sha256": f.current_sha256,
        "after_sha256": if published { f.new_sha256.clone() } else { None },
        "published": published,
        "error": error,
    })
}

fn file_result_with(f: &FilePlan, e: &HostReceiptEntry) -> serde_json::Value {
    json!({
        "path": f.path,
        "action": f.action,
        "before_sha256": e.before_sha256,
        "after_sha256": if e.published { e.after_sha256.clone() } else { None },
        "published": e.published,
        "error": e.error.clone(),
    })
}

fn host_outcome_json(o: &HostOutcome) -> serde_json::Value {
    let mut v = host_outcome_fields(o);
    if let Some(dir) = &o.backup_dir {
        v["backup_dir"] = json!(dir);
    }
    v
}

fn host_outcome_fields(o: &HostOutcome) -> serde_json::Value {
    json!({
        "host": o.host,
        "status": o.status,
        "files": o.files,
        "receipt_status": o.receipt_status,
        "reload_required": o.reload_required,
        "reload_basis": o.reload_basis,
        "blocked_reason": o.blocked_reason,
        "host_loaded": false,
        "execution_verified": "not_run",
    })
}
