//! File apply engine (Issue #1989, 01-B): the three-way decision, atomic per-file
//! publish and per-file receipts with recovery. This module is the shared apply engine
//! of 03-A/03-C (generation) and 04-B/04-C (UPDATE and partial-failure recovery).
//!
//! A multi-file apply is NOT a transaction: each file is published atomically
//! (unique temp in the target directory → write → sync_all → rename, or an exclusive
//! create via hard link → directory sync_all), the receipt records exactly what this
//! request published, and `apply-recover` reverts exactly those files — only while
//! their bytes still match the receipt's `after` digest.

use crate::identity;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs::{self, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// HOLD-B1 (user-confirmed 2026-11-08, ADK_1989_CONFIRMED_20261108_01, P09 pending):
/// receipts live under the project
/// state dir and stay out of git through the agents-managed `.gitignore` block. The
/// constant is the swap point for the later home-side layout. Worktree deletion loses
/// the receipt with the tree — a documented known gap.
// @MX:DEBT [P09 pending] Project-local receipt directory, HOLD-B1 A user-confirmed 2026-11-08 (ADK_1989_CONFIRMED_20261108_01).
// @MX:CEILING Worktree deletion loses the receipt with the tree; the recovery gap is documented in docs/rust-product-path.md.
// @MX:UPGRADE Move to the home-side runs layout when P09 resolves — this constant is the single swap point.
pub const RECEIPTS_DIR: &str = crate::project_state!("receipts");
const RECEIPT_SCHEMA_VERSION: u32 = 1;
/// Same bounded-file rule as every other state file.
const RECEIPT_LIMIT: u64 = 262_144;

#[derive(Debug, PartialEq, Eq)]
pub enum Decision {
    KeepCurrent,
    Apply,
    Conflict,
}

/// Three-way decision on digests. Missing values are distinct from present values and
/// drift away from the baseline is never overwritten (REQ-01B-01).
pub fn decide(baseline: Option<&str>, current: Option<&str>, desired: Option<&str>) -> Decision {
    if current == desired || desired == baseline {
        Decision::KeepCurrent
    } else if current == baseline {
        Decision::Apply
    } else {
        Decision::Conflict
    }
}

/// Shared repo-relative path rule of the managed files (moved from agents.rs).
pub fn relative_ok(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 256
        && !s.starts_with('/')
        && !s.contains('\\')
        && !s.contains(':')
        && !s.chars().any(char::is_control)
        && !s.split('/').any(|seg| seg == "..")
}

pub fn root_of(root: &Path) -> Result<PathBuf> {
    let root = root
        .canonicalize()
        .map_err(|_| "project directory unavailable")?;
    if !root.is_dir() {
        return Err("project must be a directory".into());
    }
    Ok(root)
}

/// Resolves a repo-relative target and refuses symlinks on the target or any ancestor
/// below the project root (moved from agents.rs, messages unchanged).
pub fn target(root: &Path, rel: &str) -> Result<PathBuf> {
    let path = root.join(rel);
    let mut cursor = path.clone();
    loop {
        match fs::symlink_metadata(&cursor) {
            Ok(meta) if meta.file_type().is_symlink() => {
                return Err("symlink target refused".into())
            }
            Ok(_) => (),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
            Err(_) => return Err("target metadata unavailable".into()),
        }
        match cursor.parent() {
            Some(parent) if parent.starts_with(root) && parent != root => cursor = parent.into(),
            _ => break,
        }
    }
    Ok(path)
}

pub fn read_opt(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::metadata(path) {
        Ok(m) if !m.is_file() || m.len() > RECEIPT_LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
    }
}

fn cross_device(e: &std::io::Error) -> bool {
    // EXDEV (Linux/macOS errno 18) / ERROR_NOT_SAME_DEVICE (Windows 17).
    matches!(e.raw_os_error(), Some(18) | Some(17))
}

fn sync_parent(path: &Path) {
    if let Some(parent) = path.parent() {
        if let Ok(dir) = fs::File::open(parent) {
            let _ = dir.sync_all();
        }
    }
}

/// Stages `after` into a unique temporary file in the target's directory (create_new,
/// write, sync_all), re-verifying the target still matches `before` before publishing.
/// The temporary file carries the target's permissions on unix and is removed on every
/// failure path (REQ-01B-03).
fn stage(path: &Path, before: Option<&[u8]>, after: &[u8]) -> Result<PathBuf> {
    let parent = path.parent().ok_or("target parent unavailable")?;
    fs::create_dir_all(parent).map_err(|_| "target directory unavailable")?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let temp = parent.join(format!(".moai-apply-{}-{nanos}.tmp", std::process::id()));
    let mut file = OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&temp)
        .map_err(|_| "temporary file refused")?;
    #[cfg(unix)]
    if let Ok(meta) = fs::metadata(path) {
        let _ = fs::set_permissions(&temp, meta.permissions());
    }
    let written = file
        .write_all(after)
        .and_then(|_| file.sync_all())
        .map_err(|_| "temporary write failed".to_string());
    // The re-check may itself fail (for example a permission change mid-apply); the
    // temporary file is removed on that path too, never left behind.
    let verified = read_opt(path);
    if verified.is_err() || written.is_err() || verified.unwrap().as_deref() != before {
        let _ = fs::remove_file(&temp);
        return Err(match written {
            Err(message) => message,
            _ => "target changed during apply; temporary file discarded".into(),
        });
    }
    Ok(temp)
}

/// Atomic replace (REQ-01B-03): stage → rename → directory sync. A rename across
/// filesystems is refused, never downgraded to copy+delete (REQ-01B-05).
pub fn publish_replace(path: &Path, before: Option<&[u8]>, after: &[u8]) -> Result<()> {
    let temp = stage(path, before, after)?;
    match fs::rename(&temp, path) {
        Ok(()) => {}
        Err(e) => {
            let _ = fs::remove_file(&temp);
            return Err(if cross_device(&e) {
                "cross-filesystem move refused; temporary and target must share one filesystem"
                    .into()
            } else {
                "atomic replace failed".into()
            });
        }
    }
    sync_parent(path);
    Ok(())
}

/// Exclusive create (REQ-01B-04, HOLD-B2): stage → publish via hard link, which fails
/// when the target appeared meanwhile (`exclusive create refused`). Chosen over
/// renameat2 (OS branching, libc dependency) and over a direct `create_new` write
/// (a crash could leave a partial target behind). A filesystem without hard links
/// fails the publish explicitly; no fallback, no support claim.
// @MX:NOTE [HOLD-B2 A] Hard-link publish verified on macOS/APFS fixtures only; a
// no-hard-link filesystem fails explicitly and no wider OS/FS claim is made.
pub fn publish_exclusive(path: &Path, after: &[u8]) -> Result<()> {
    let temp = stage(path, None, after)?;
    let published = fs::hard_link(&temp, path)
        .map_err(|_| "exclusive create refused")
        .map(|_| ());
    let _ = fs::remove_file(&temp);
    published?;
    sync_parent(path);
    Ok(())
}

/// One file's apply outcome (REQ-01B-06). `decision` carries the three-way action
/// (`create|append|merge|unchanged|conflict`) or `stale_target` when the apply refused
/// the file (REQ-01B-02).
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ReceiptEntry {
    pub path: String,
    pub decision: String,
    pub before_sha256: Option<String>,
    pub after_sha256: Option<String>,
    pub published: bool,
    pub error: Option<String>,
}

/// Per-request receipt (shared interface of 03-C/04-B/04-C). The JSON schema carries
/// digests only; the reverted-to bytes are kept beside the receipt under
/// `<request-id>.before/` so recovery can restore actual content (REQ-01B-08).
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub schema_version: u32,
    pub request_id: String,
    pub operation: String,
    pub clone_id: Option<String>,
    pub workspace_id: Option<String>,
    pub entries: Vec<ReceiptEntry>,
    /// complete | partial | refused
    pub status: String,
}

impl Receipt {
    pub fn new(request_id: &str, operation: &str, root: &Path) -> Self {
        let (clone_id, workspace_id) = identity_fields(root);
        Receipt {
            schema_version: RECEIPT_SCHEMA_VERSION,
            request_id: request_id.into(),
            operation: operation.into(),
            clone_id,
            workspace_id,
            entries: Vec::new(),
            status: "partial".into(),
        }
    }

    pub fn entry(
        &mut self,
        path: &str,
        decision: &str,
        before: Option<&str>,
        after: Option<&str>,
        published: bool,
        error: Option<String>,
    ) {
        self.entries.push(ReceiptEntry {
            path: path.into(),
            decision: decision.into(),
            before_sha256: before.map(str::to_string),
            after_sha256: after.map(str::to_string),
            published,
            error,
        });
    }

    /// Writes the receipt through the same publish procedure as a target file
    /// (REQ-01B-06). Callers save after every published file, so a crash leaves the
    /// newest state on disk; the window between a publish and its receipt save is the
    /// known recovery gap documented in docs/rust-product-path.md.
    pub fn save(&self, root: &Path) -> Result<()> {
        let path = receipt_path(root, &self.request_id)?;
        let bytes = serde_json::to_vec_pretty(self).map_err(|_| "receipt encode failed")?;
        let before = read_opt(&path)?;
        publish_replace(&path, before.as_deref(), &bytes)
    }
}

/// Receipt file name for `--receipt`; only the request ids this engine creates pass.
fn receipt_id_ok(s: &str) -> bool {
    !s.is_empty()
        && s.len() <= 128
        && s.bytes()
            .all(|b| b.is_ascii_alphanumeric() || b == b'-' || b == b'_')
}

/// The receipt store directory, honouring the P09 opt-in home layout (#2175): once
/// `layout.json` is switched in the home project directory, receipts resolve there;
/// otherwise the project-side store stays authoritative (read-compatible fallback).
fn receipts_dir(root: &Path) -> Result<PathBuf> {
    Ok(crate::layout::switched_dir(root, "receipts").unwrap_or_else(|| root.join(RECEIPTS_DIR)))
}

pub fn receipt_path(root: &Path, request_id: &str) -> Result<PathBuf> {
    let dir = receipts_dir(root)?;
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.file_type().is_symlink() => {
            return Err("receipt directory is a symlink; refused".into())
        }
        Ok(m) if !m.is_dir() => return Err("receipt path is not a directory".into()),
        Ok(_) => (),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => (),
        Err(_) => return Err("receipt directory unavailable".into()),
    }
    Ok(dir.join(format!("{request_id}.json")))
}

/// New unique request id: operation, a plan-digest head (which request it belongs to)
/// and the nanosecond clock (which run it is). File-name-safe characters only.
pub fn new_request_id(operation: &str, plan_digest: &str) -> String {
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let hex = plan_digest.strip_prefix("sha256:").unwrap_or(plan_digest);
    let head: String = hex.chars().take(12).collect();
    format!("{}-{}-{}", operation.replace('.', "-"), head, nanos)
}

/// Best-effort clone/workspace identity of the project root (01-A). Identity is
/// recorded for traceability; a resolver failure never fails the apply.
fn identity_fields(root: &Path) -> (Option<String>, Option<String>) {
    match identity::identity(root) {
        Ok(v) if v["identity_status"] == "ok" => (
            v["clone_id"].as_str().map(str::to_string),
            v["workspace_id"].as_str().map(str::to_string),
        ),
        _ => (None, None),
    }
}

/// Keeps the before bytes of a published file so recovery can restore content
/// (REQ-01B-08). Stored beside the receipt under `<request-id>.before/`, same
/// directory policy and publish procedure as the receipt itself.
pub fn save_before_bytes(root: &Path, request_id: &str, rel: &str, bytes: &[u8]) -> Result<()> {
    let dir = receipts_dir(root)?.join(format!("{request_id}.before"));
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.file_type().is_symlink() => {
            return Err("receipt backup directory is a symlink; refused".into())
        }
        Ok(_) => (),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            fs::create_dir_all(&dir).map_err(|_| "receipt backup directory unavailable")?
        }
        Err(_) => return Err("receipt backup directory unavailable".into()),
    }
    let path = dir.join(rel);
    let before = read_opt(&path)?;
    publish_replace(&path, before.as_deref(), bytes)
}

fn before_bytes_path(root: &Path, request_id: &str, rel: &str) -> Result<PathBuf> {
    let dir = receipts_dir(root)?.join(format!("{request_id}.before"));
    match fs::symlink_metadata(&dir) {
        Ok(m) if m.file_type().is_symlink() => {
            return Err("receipt backup directory is a symlink; refused".into())
        }
        Ok(_) => (),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return Err("receipt backup directory unavailable".into())
        }
        Err(_) => return Err("receipt backup directory unavailable".into()),
    }
    Ok(dir.join(rel))
}

/// REQ-01B-09: no receipt or recovery under Go-generation state; the receipt store
/// itself is also out of bounds for recovery (a receipt must not be "recovered").
fn guard_recoverable_path(rel: &str) -> Result<()> {
    if !relative_ok(rel) {
        return Err("receipt path is not a bounded relative path; refused".into());
    }
    let blocked = [
        (".moai", "go generation state path"),
        (RECEIPTS_DIR, "receipt store path"),
    ];
    for (prefix, why) in blocked {
        if rel == prefix || rel.starts_with(&format!("{prefix}/")) {
            return Err(format!("receipt names a {why}; refused"));
        }
    }
    Ok(())
}

/// REQ-01B-08: reverts exactly the files one apply request published — and only while
/// the file's current bytes still match the receipt's `after` digest. A file changed
/// meanwhile is skipped as `user_changed_since_apply`, never overwritten. A created
/// file is removed again; a replaced file is restored from the stored before bytes
/// through the same publish procedure. This is a mutation: it requires the same
/// permission reference as an apply (REQ-01C-03).
pub fn recover(root: &Path, receipt_id: &str, permission: &str) -> Result<Value> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    if !receipt_id_ok(receipt_id) {
        return Err("invalid receipt id; refused".into());
    }
    let root = root_of(root)?;
    let path = receipt_path(&root, receipt_id)?;
    let bytes = read_opt(&path)?.ok_or("receipt not found")?;
    let receipt: Receipt = serde_json::from_slice(&bytes)
        .map_err(|_| "receipt unreadable; inspect before recovering".to_string())?;
    // Every path is checked before anything changes, so a refused receipt changes nothing.
    let published: Vec<(&ReceiptEntry, PathBuf)> = receipt
        .entries
        .iter()
        .filter(|e| e.published)
        .map(|e| {
            guard_recoverable_path(&e.path)?;
            Ok((e, target(&root, &e.path)?))
        })
        .collect::<Result<_>>()?;
    let mut reverted: Vec<String> = Vec::new();
    let mut skipped = Vec::new();
    for (i, (entry, target)) in published.iter().enumerate() {
        let step = || -> Result<bool> {
            let current = read_opt(target)?;
            let current_digest = current.as_deref().map(crate::digest);
            if current_digest.as_deref() != entry.after_sha256.as_deref() {
                return Ok(false);
            }
            match &entry.before_sha256 {
                None => {
                    // The request created this file; reverting means removing it.
                    fs::remove_file(target).map_err(|_| "target remove failed")?;
                }
                Some(before_digest) => {
                    let bytes = read_opt(&before_bytes_path(&root, receipt_id, &entry.path)?)?
                        .ok_or("receipt backup not found; inspect before recovering")?;
                    if crate::digest(&bytes) != *before_digest {
                        return Err(
                            "receipt backup digest mismatch; inspect before recovering".into()
                        );
                    }
                    // The stage-time re-check compares against the file's current bytes, so
                    // a change between the digest check and the rename still refuses.
                    publish_replace(target, current.as_deref(), &bytes)?;
                }
            }
            Ok(true)
        };
        match step() {
            Ok(true) => reverted.push(entry.path.clone()),
            Ok(false) => {
                skipped.push(json!({"path": entry.path, "reason": "user_changed_since_apply"}))
            }
            // Nothing changed yet: the plain refusal is the whole truth.
            Err(e) if reverted.is_empty() => return Err(e),
            // Part of the request is already reverted (Issue #2072): say exactly what.
            Err(e) => {
                let skipped_list: Vec<String> = skipped
                    .iter()
                    .map(|s| {
                        format!(
                            "{}: {}",
                            s["path"].as_str().unwrap_or(""),
                            s["reason"].as_str().unwrap_or("")
                        )
                    })
                    .collect();
                let rest: Vec<&str> = published[i + 1..]
                    .iter()
                    .map(|(e, _)| e.path.as_str())
                    .collect();
                return Err(format!(
                    "partial apply: recover {receipt_id} reverted [{}]; failed at {}: {e}; skipped [{}]; not attempted [{}]",
                    reverted.join(", "),
                    entry.path,
                    skipped_list.join(", "),
                    rest.join(", "),
                ));
            }
        }
    }
    Ok(json!({
        "state": if reverted.is_empty() { "nothing_to_revert" } else { "reverted" },
        "receipt": receipt_id,
        "operation": receipt.operation,
        "reverted": reverted,
        "skipped": skipped,
        "host_verified": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn three_way_decisions() {
        // The eight rows of the decision table (REQ-01B-01).
        assert_eq!(decide(None, None, Some("new")), Decision::Apply); // create
        assert_eq!(
            decide(Some("old"), Some("old"), Some("new")),
            Decision::Apply
        ); // clean update
        assert_eq!(decide(None, Some("x"), Some("x")), Decision::KeepCurrent);
        assert_eq!(
            decide(Some("old"), Some("user"), Some("old")),
            Decision::KeepCurrent
        ); // user edit is final
        assert_eq!(
            decide(Some("old"), Some("new"), Some("new")),
            Decision::KeepCurrent
        ); // already applied
        assert_eq!(decide(None, Some("user"), Some("new")), Decision::Conflict);
        assert_eq!(
            decide(Some("old"), Some("user"), Some("new")),
            Decision::Conflict
        ); // user edit + template change
        assert_eq!(decide(Some("old"), None, Some("new")), Decision::Conflict);
    }

    #[test]
    fn relative_paths_are_bounded() {
        assert!(relative_ok("skills/moai/workflows/"));
        assert!(relative_ok(crate::project_state!(
            "manifests/generation.json"
        )));
        assert!(!relative_ok("/abs"));
        assert!(!relative_ok("../up"));
        assert!(!relative_ok("a/../b"));
        assert!(!relative_ok("C:\\x"));
        assert!(!relative_ok(""));
    }

    #[test]
    fn request_ids_are_file_name_safe() {
        let id = new_request_id("agents.apply", "sha256:0123456789abcdef");
        assert!(id.starts_with("agents-apply-0123456789ab-"));
        assert!(receipt_id_ok(&id));
        assert!(!receipt_id_ok("../escape"));
        assert!(!receipt_id_ok(""));
    }

    #[test]
    fn recover_guard_blocks_go_and_receipt_paths() {
        assert!(guard_recoverable_path("AGENTS.md").is_ok());
        assert!(guard_recoverable_path(crate::project_state!("manifests/generation.json")).is_ok());
        assert!(guard_recoverable_path(".moai/state/x.json").is_err());
        assert!(guard_recoverable_path(".moai").is_err());
        assert!(guard_recoverable_path(crate::project_state!("receipts/x.json")).is_err());
        assert!(guard_recoverable_path("/abs").is_err());
        assert!(guard_recoverable_path("a/../b").is_err());
    }
}
