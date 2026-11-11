//! UPDATE partial-failure recovery (Issue #2000, 04-C).
//!
//! 04-B leaves a host whose apply failed with a `pending` receipt that lists this request's
//! files (`before_sha256`/`after_sha256`) and keeps the pre-apply ownership manifest.
//! `recovery-plan` reads every host receipt and classifies each file of a pending host by
//! its CURRENT bytes (never by the receipt's `published` flag: a crash between a rename and
//! the receipt save leaves `published: false` on a file that already carries `after`):
//!
//! - `matches_after`: this request's change, revertible — a created file (`before` none) is
//!   removed; a replaced file is restored from its known-good copy
//!   `~/.moai-cli/installations/<id>/hosts/moai/<host>/known-good/<sha256 hex>` (REQ-04C-03),
//!   and without that copy the file is `conflict` (`no_known_good`) instead;
//! - `matches_before`: never applied, nothing to do;
//! - `conflict`: anything else (a user edit, a symlink) — preserved;
//! - `missing`: the file is gone — preserved, reported.
//!
//! The plan writes nothing. `recovery-apply` recomputes it, refuses another digest, and for
//! each revertible file re-reads the bytes right before the remove/replace (REQ-04C-02).
//! A committed or recovered host is never written. A host with nothing left to resolve is
//! marked `recovered`; one with conflicts stays `pending`. A failure part way stops there,
//! records what was reverted beside the receipt and never retries (REQ-04C-07).
//!
//! Leader RC1 decisions (2026-11-08, recorded on #2000): recovery is triggered only by these
//! two explicit commands (P06 = A); the known-good location is the REQ-04C-03 path (P09).
//! 04-B stores the known-good copies and the receipt's `version_before` (#2089/#2091); a
//! receipt written before that, or a first apply, has neither, so a replaced file is then
//! `conflict(no_known_good)` and `version_before` is null with its reason.
// @MX:DEBT [P09 retention] Known-good copies have no retention period and no GC in RC1.
// UPGRADE when P09 retention is decided: prune copies no committed/pending receipt names.

use crate::apply::{publish_replace, read_opt, relative_ok};
use crate::contract::PROTOCOL_VERSION;
use crate::update::{
    current_install_id, host_receipt_path, load_host_receipt, now_rfc3339, projection_root,
    save_host_receipt, sha256_hex, HostReceipt, FILE_LIMIT, HOSTS,
};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

const RECOVERY_SCHEMA_VERSION: u32 = 1;

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct RecoveryFile {
    pub path: String,
    /// matches_before | matches_after | conflict | missing
    pub class: String,
    /// For `conflict`: `no_known_good` | `user_changed` | `symlink` | `not_a_regular_file`.
    pub reason: Option<String>,
    /// remove | restore | none
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: Option<String>,
    pub current_sha256: Option<String>,
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct RecoveryHost {
    pub host: String,
    /// none (no receipt) | pending | committed | recovered | blocked
    pub receipt_status: String,
    pub blocked_reason: Option<String>,
    pub package_version: Option<String>,
    /// The receipt's `version_before` (#2091): the version the failed apply replaced.
    pub version_before: Option<String>,
    pub files: Vec<RecoveryFile>,
}

/// Why `version_before` is null, or null when it is known (REQ-04C-06).
fn version_before_basis(h: &RecoveryHost) -> Value {
    match (&h.version_before, h.receipt_status.as_str()) {
        (Some(_), _) => Value::Null,
        (None, "none") => json!("no host receipt"),
        (None, "blocked") => json!("host receipt unreadable or refused"),
        (None, _) => json!(
            "not recorded: a first apply on this host, or a receipt written before version_before existed (#2091)"
        ),
    }
}

#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(deny_unknown_fields)]
pub struct RecoveryPlan {
    pub schema_version: u32,
    pub protocol_version: u32,
    pub home: String,
    pub install_id: String,
    pub hosts: Vec<RecoveryHost>,
    pub plan_digest: String,
}

/// `hosts/moai/<host>/known-good/<hex>` beside the host receipt.
fn known_good_path(home: &Path, install_id: &str, host: &str, sha: &str) -> Result<PathBuf> {
    let receipt = host_receipt_path(home, install_id, host, false)?;
    let hex = sha.strip_prefix("sha256:").unwrap_or(sha);
    if hex.len() != 64 || !hex.bytes().all(|b| b.is_ascii_hexdigit()) {
        return Err("invalid host receipt digest".into());
    }
    let dir = receipt
        .parent()
        .ok_or("host receipt directory unavailable")?;
    Ok(dir.join(host).join("known-good").join(hex))
}

/// The known-good bytes for `sha`, only if they are a bounded regular file (no symlink in
/// the store) whose digest is exactly `sha`.
fn known_good(home: &Path, install_id: &str, host: &str, sha: &str) -> Result<Option<Vec<u8>>> {
    let path = known_good_path(home, install_id, host, sha)?;
    let mut cursor = path.as_path();
    let stop = home.join(crate::HOME_STATE_DIR);
    while cursor.starts_with(&stop) && cursor != stop {
        if let Ok(m) = fs::symlink_metadata(cursor) {
            if m.file_type().is_symlink() {
                return Ok(None);
            }
        }
        cursor = match cursor.parent() {
            Some(p) => p,
            None => break,
        };
    }
    match fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() && m.len() <= FILE_LIMIT => {
            let bytes = fs::read(&path).map_err(|_| "known-good read failed")?;
            Ok((sha256_hex(&bytes) == sha).then_some(bytes))
        }
        _ => Ok(None),
    }
}

/// True when any component from the projection root down to `target` is a symlink.
fn symlink_on_path(root: &Path, target: &Path) -> bool {
    let mut cursor = target;
    loop {
        if let Ok(m) = fs::symlink_metadata(cursor) {
            if m.file_type().is_symlink() {
                return true;
            }
        }
        if cursor == root {
            return false;
        }
        cursor = match cursor.parent() {
            Some(p) if p.starts_with(root) => p,
            _ => return false,
        };
    }
}

fn classify_file(
    home: &Path,
    install_id: &str,
    host: &str,
    root: &Path,
    entry: &crate::update::HostReceiptEntry,
) -> Result<RecoveryFile> {
    let mut file = RecoveryFile {
        path: entry.path.clone(),
        class: "conflict".into(),
        reason: None,
        action: "none".into(),
        before_sha256: entry.before_sha256.clone(),
        after_sha256: entry.after_sha256.clone(),
        current_sha256: None,
    };
    if !relative_ok(&entry.path) {
        return Err("host receipt path is not a bounded relative path; refused".into());
    }
    let target = root.join(&entry.path);
    if symlink_on_path(root, &target) {
        file.reason = Some("symlink".into());
        return Ok(file);
    }
    let current = match read_opt(&target) {
        Ok(c) => c,
        Err(_) => {
            file.reason = Some("not_a_regular_file".into());
            return Ok(file);
        }
    };
    file.current_sha256 = current.as_deref().map(sha256_hex);
    let current_sha = file.current_sha256.as_deref();
    if current_sha.is_some() && current_sha == entry.after_sha256.as_deref() {
        match entry.before_sha256.as_deref() {
            None => {
                file.class = "matches_after".into();
                file.action = "remove".into();
            }
            Some(before) => {
                if known_good(home, install_id, host, before)?.is_some() {
                    file.class = "matches_after".into();
                    file.action = "restore".into();
                } else {
                    file.reason = Some("no_known_good".into());
                }
            }
        }
    } else if current_sha == entry.before_sha256.as_deref() {
        file.class = "matches_before".into();
    } else if current_sha.is_none() {
        file.class = "missing".into();
    } else {
        file.reason = Some("user_changed".into());
    }
    Ok(file)
}

/// REQ-04C-01: read-only classification of every host. Writes nothing.
pub fn recovery_plan(home: &Path) -> Result<RecoveryPlan> {
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let install_id = current_install_id(&canonical_home)?;
    let mut hosts = Vec::new();
    for host in HOSTS {
        let path = host_receipt_path(&canonical_home, &install_id, host, false)?;
        let mut h = RecoveryHost {
            host: host.to_string(),
            receipt_status: "none".into(),
            blocked_reason: None,
            package_version: None,
            version_before: None,
            files: Vec::new(),
        };
        let receipt = match load_host_receipt(&path) {
            Ok(r) => r,
            Err(e) => {
                // A damaged or substituted receipt: nothing about this host can be trusted.
                h.receipt_status = "blocked".into();
                h.blocked_reason = Some(e);
                hosts.push(h);
                continue;
            }
        };
        let Some(receipt) = receipt else {
            hosts.push(h);
            continue;
        };
        h.package_version = Some(receipt.package_version.clone());
        h.version_before = receipt.version_before.clone();
        h.receipt_status = receipt.status.clone();
        // REQ-3 (#2268): a FIRST apply (`version_before` none) into a projection root
        // that already existed (`top_skills_existed_before`) landed created files
        // inside a previously unmanaged folder — list them so the operator sees what
        // an apply added there. Fresh installs into an empty folder (the normal path)
        // never trigger this.
        if receipt.version_before.is_none() && receipt.top_skills_existed_before {
            for entry in &receipt.entries {
                if entry.before_sha256.is_none() && entry.published {
                    h.files.push(crate::recovery::RecoveryFile {
                        path: entry.path.clone(),
                        class: "added_into_pre_existing_root".into(),
                        reason: Some(
                            "first apply landed this file in a folder that already existed \
                             (unmanaged root at apply time, #2268)"
                                .into(),
                        ),
                        action: "none".into(),
                        before_sha256: None,
                        after_sha256: entry.after_sha256.clone(),
                        current_sha256: None,
                    });
                }
            }
        }
        if receipt.status == "pending" {
            let root = projection_root(&canonical_home, host);
            if fs::symlink_metadata(&root).is_ok_and(|m| m.file_type().is_symlink()) {
                h.receipt_status = "blocked".into();
                h.blocked_reason = Some("projection path is a symlink; refused".into());
            } else {
                for entry in &receipt.entries {
                    h.files.push(classify_file(
                        &canonical_home,
                        &install_id,
                        host,
                        &root,
                        entry,
                    )?);
                }
            }
        }
        hosts.push(h);
    }
    let mut plan = RecoveryPlan {
        schema_version: RECOVERY_SCHEMA_VERSION,
        protocol_version: PROTOCOL_VERSION,
        home: canonical_home
            .to_str()
            .ok_or("non UTF-8 home path")?
            .to_string(),
        install_id,
        hosts,
        plan_digest: String::new(),
    };
    plan.plan_digest = crate::digest(&serde_json::to_vec(&plan).map_err(|_| "encode failed")?);
    Ok(plan)
}

/// What one recovery did on one host, saved beside its receipt (REQ-04C-07).
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RecoveryRecord {
    pub schema_version: u32,
    pub host: String,
    pub plan_digest: String,
    pub started_at: String,
    pub finished_at: String,
    /// recovered | pending (conflicts remain) | failed (stopped part way; no retry)
    pub status: String,
    pub reverted: Vec<String>,
    pub conflicts: Vec<Value>,
    pub failed: Option<Value>,
    pub not_attempted: Vec<String>,
}

fn save_record(home: &Path, install_id: &str, record: &RecoveryRecord) -> Result<()> {
    let receipt = host_receipt_path(home, install_id, &record.host, true)?;
    let path = receipt.with_file_name(format!("{}.recovery.json", record.host));
    if fs::symlink_metadata(&path).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err("recovery record is a symlink; refused".into());
    }
    let bytes = serde_json::to_vec_pretty(record).map_err(|_| "encode failed".to_string())?;
    let before = read_opt(&path)?;
    publish_replace(&path, before.as_deref(), &bytes)
}

fn validate_permission(permission: &str) -> Result<()> {
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    Ok(())
}

fn conflict_json(f: &RecoveryFile) -> Value {
    json!({"path": f.path, "class": f.class, "reason": f.reason,
           "current_sha256": f.current_sha256})
}

/// REQ-04C-02..07: reverts only `matches_after` files of pending hosts.
pub fn recovery_apply(
    home: &Path,
    plan: &RecoveryPlan,
    plan_digest: &str,
    permission_reference: &str,
) -> Result<Value> {
    validate_permission(permission_reference)?;
    let fresh = recovery_plan(home)?;
    if *plan != fresh || fresh.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let home = PathBuf::from(&fresh.home);
    let install_id = fresh.install_id.as_str();
    let mut reports = Vec::new();
    let mut reverted_anywhere: Vec<String> = Vec::new();
    let mut conflicts_remain = false;

    for h in &fresh.hosts {
        let applied: Vec<&str> = h
            .files
            .iter()
            .filter(|f| f.class != "matches_before")
            .map(|f| f.path.as_str())
            .collect();
        if h.receipt_status != "pending" {
            // Committed, recovered, absent or blocked: never written (REQ-04C-05).
            conflicts_remain |= h.receipt_status == "blocked";
            reports.push(json!({"host": h.host, "receipt_status": h.receipt_status,
                "blocked_reason": h.blocked_reason, "version_before": h.version_before,
                "version_before_basis": version_before_basis(h),
                "version_after": h.package_version, "applied": [], "failed": [],
                "recovered": [], "conflicts": [], "written": false}));
            continue;
        }
        let root = projection_root(&home, &h.host);
        let started = now_rfc3339();
        let conflicts: Vec<Value> = h
            .files
            .iter()
            .filter(|f| f.class == "conflict" || f.class == "missing")
            .map(conflict_json)
            .collect();
        let todo: Vec<&crate::recovery::RecoveryFile> =
            h.files.iter().filter(|f| f.action != "none").collect();
        let mut reverted: Vec<String> = Vec::new();
        for (i, f) in todo.iter().enumerate() {
            let target = root.join(&f.path);
            let step = || -> Result<()> {
                // Re-verify right before the write: still exactly this request's bytes.
                if symlink_on_path(&root, &target) {
                    return Err("target became a symlink; refused".into());
                }
                let current = read_opt(&target)?;
                if current.as_deref().map(sha256_hex).as_deref() != f.after_sha256.as_deref() {
                    return Err("target changed since recovery-plan".into());
                }
                match (f.action.as_str(), f.before_sha256.as_deref()) {
                    ("remove", None) => {
                        fs::remove_file(&target).map_err(|_| "target remove failed".to_string())
                    }
                    ("restore", Some(before)) => {
                        let bytes = known_good(&home, install_id, &h.host, before)?
                            .ok_or("known-good copy changed since recovery-plan")?;
                        publish_replace(&target, current.as_deref(), &bytes)
                    }
                    _ => Err("invalid recovery action".into()),
                }
            };
            if let Err(e) = step() {
                let rest: Vec<String> = todo[i + 1..].iter().map(|f| f.path.clone()).collect();
                let record = RecoveryRecord {
                    schema_version: RECOVERY_SCHEMA_VERSION,
                    host: h.host.clone(),
                    plan_digest: plan_digest.to_string(),
                    started_at: started.clone(),
                    finished_at: now_rfc3339(),
                    status: "failed".into(),
                    reverted: reverted.clone(),
                    conflicts: conflicts.clone(),
                    failed: Some(json!({"path": f.path, "reason": e})),
                    not_attempted: rest.clone(),
                };
                save_record(&home, install_id, &record)?;
                reverted_anywhere.extend(reverted.iter().cloned());
                if reverted_anywhere.is_empty() {
                    return Err(format!(
                        "conflict in recovery: {} {} could not be reverted ({e}); nothing reverted; the receipt stays pending, no retry",
                        h.host, f.path
                    ));
                }
                return Err(format!(
                    "partial apply: recovery {} reverted [{}]; failed at {}: {e}; conflicts [{}]; not attempted [{}]",
                    h.host,
                    reverted_anywhere.join(", "),
                    f.path,
                    conflicts
                        .iter()
                        .filter_map(|c| c["path"].as_str())
                        .collect::<Vec<_>>()
                        .join(", "),
                    rest.join(", "),
                ));
            }
            reverted.push(f.path.clone());
        }
        let status = if conflicts.is_empty() {
            "recovered"
        } else {
            "pending"
        };
        conflicts_remain |= !conflicts.is_empty();
        if !reverted.is_empty() || status == "recovered" {
            let path = host_receipt_path(&home, install_id, &h.host, false)?;
            let mut receipt: HostReceipt =
                load_host_receipt(&path)?.ok_or("host receipt disappeared")?;
            if receipt.status != "pending" {
                return Err("host receipt changed since recovery-plan".into());
            }
            receipt.status = status.into();
            receipt.finished_at = Some(now_rfc3339());
            save_host_receipt(&home, install_id, &receipt)?;
        }
        save_record(
            &home,
            install_id,
            &RecoveryRecord {
                schema_version: RECOVERY_SCHEMA_VERSION,
                host: h.host.clone(),
                plan_digest: plan_digest.to_string(),
                started_at: started,
                finished_at: now_rfc3339(),
                status: status.into(),
                reverted: reverted.clone(),
                conflicts: conflicts.clone(),
                failed: None,
                not_attempted: Vec::new(),
            },
        )?;
        reverted_anywhere.extend(reverted.iter().cloned());
        reports.push(json!({"host": h.host, "receipt_status": status,
            "blocked_reason": null,
            "version_before": h.version_before,
            "version_before_basis": version_before_basis(h),
            "version_after": h.package_version, "applied": applied, "failed": [],
            "recovered": reverted, "conflicts": conflicts, "written": true}));
    }
    Ok(json!({
        "status": if conflicts_remain { "partial" } else { "success" },
        "plan_digest": plan_digest,
        "hosts": reports,
        "retry": "none: a failure is reported, never retried",
        "host_loaded": false,
    }))
}
