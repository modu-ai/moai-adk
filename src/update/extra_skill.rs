//! The skill folders `moai update` installs beside the moai skill
//! (SPEC-20261122-html-report-install).
//!
//! A release may carry more skill folders next to `<version folder>/skills/moai`; today that is
//! `moai-html-report`, the internal report skill. Each folder found there is one part of the user
//! layer: copied into the `skills` folder of the Claude config folder ([`crate::host_home`]:
//! `CLAUDE_CONFIG_DIR`, else `~/.claude`), planned three-way per file like the moai skill, and
//! bound into the layer's `plan_digest`. A release without the folder has no part, so its preview
//! keeps the keys it had.
//!
//! The rules are the moai skill's: a package file on disk that no receipt of this skill owns blocks
//! the part (the reason names the folder and says moai did not install it; nothing is written
//! there), a pending receipt blocks it until it is recovered, and a newer installed version blocks
//! it unless `--allow-older` was given. The part is planned and applied on its own: a blocked moai
//! skill copy does not stop it, and a blocked part does not stop the moai skill.
//!
//! Receipts live in their own namespace, `installations/<id>/hosts/<skill>/claude.json`, with
//! known-good copies under `hosts/<skill>/claude/known-good`. The receipt path, the receipt choice
//! and the known-good copies of `update.rs` name the moai folder; this module keeps copies of them
//! that take the skill name.

use super::{
    ancestor_symlinks, cross_filesystem, current_install_id, earlier_install_ids, file_result,
    file_result_with, layout_folder, load_host_receipt, now_rfc3339, owned_map, owned_vec,
    reload_state, scan_projection, sha256_hex, stage_package, validate_permission, writable,
    FilePlan, HostReceipt, HostReceiptEntry, ProjectionScan, ReceiptSource, StagedPackage,
};
use crate::apply::{decide, publish_exclusive, publish_replace, read_opt, Decision};
use crate::host_home::{Host, Place};
use serde::Serialize;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

/// The skill folders installed beside `skills/moai`, in plan order.
pub const SKILLS: [&str; 1] = ["moai-html-report"];

/// The one host these folders go to.
const HOST: &str = "claude";

/// One skill folder's plan.
#[derive(Clone, Debug, Serialize, PartialEq)]
pub struct Part {
    pub skill: String,
    pub host: String,
    pub projection_root: String,
    /// ok | blocked
    pub preflight: String,
    pub blocked_reason: Option<String>,
    pub files: Vec<FilePlan>,
    /// The `install_id` of the earlier version folder whose receipt gives the baseline.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub receipt_from: Option<String>,
    /// The warning line when `--allow-older` plans this build over a newer installed version.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub going_back: Option<String>,
    /// The staged folder's tree digest; `None` when the folder could not be staged.
    pub package_digest: Option<String>,
}

/// The skill folders part of one `moai update` preview.
pub struct Layer {
    pub parts: Vec<Part>,
    /// The digest over `parts`; the layer's combined digest binds it.
    pub plan_digest: String,
    /// The staged folder of each part (same order), `None` where staging failed.
    staged: Vec<Option<StagedPackage>>,
}

impl Layer {
    /// True when confirming would write at least one file.
    pub fn writes(&self) -> bool {
        self.parts.iter().any(|p| {
            p.preflight != "blocked"
                && p.files
                    .iter()
                    .any(|f| f.action == "create" || f.action == "update")
        })
    }

    /// The preview value, only when the release carries another skill folder.
    pub fn to_json(&self) -> Option<Value> {
        (!self.parts.is_empty())
            .then(|| json!({"parts": self.parts, "plan_digest": self.plan_digest}))
    }

    /// One line per part for the terminal question, under the moai skill's line.
    pub fn summary(&self) -> String {
        let mut out = String::new();
        for p in &self.parts {
            let mut counts: BTreeMap<&str, usize> = BTreeMap::new();
            for f in &p.files {
                *counts.entry(f.action.as_str()).or_default() += 1;
            }
            let counts: Vec<String> = counts.iter().map(|(a, n)| format!("{a} {n}")).collect();
            out.push_str(&format!(
                "  {} {}: {} — {}{}\n",
                p.host,
                p.skill,
                p.preflight,
                counts.join(", "),
                p.blocked_reason
                    .as_ref()
                    .map(|r| format!(" ({r})"))
                    .unwrap_or_default()
            ));
            if let Some(w) = &p.going_back {
                out.push_str(&format!("  warning: {} {}: {w}\n", p.host, p.skill));
            }
        }
        out
    }
}

/// Plans every skill folder found beside `package` (the moai skill folder of the release).
/// Writes nothing. A folder that is absent has no part; one that is a link or a file, or that
/// cannot be staged, is a blocked part.
pub fn plan(home: &Path, package: &Path, allow_older: bool) -> Result<Layer> {
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    let mut parts = Vec::new();
    let mut staged = Vec::new();
    let mut install_id: Option<String> = None;
    if let Some(beside) = package.parent() {
        for skill in SKILLS {
            let src = beside.join(skill);
            let stage = match fs::symlink_metadata(&src) {
                Err(_) => continue,
                Ok(m) if m.file_type().is_symlink() => Err(format!(
                    "the installed {skill} folder {} is a symlink; refused",
                    src.display()
                )),
                Ok(m) if !m.is_dir() => Err(format!(
                    "the installed {skill} path {} is not a folder; refused",
                    src.display()
                )),
                Ok(_) => stage_package(&src)
                    .map_err(|e| format!("the installed {skill} folder cannot be read: {e}")),
            };
            let id = match &install_id {
                Some(id) => id.clone(),
                None => {
                    let id = current_install_id(&canonical_home)?;
                    install_id = Some(id.clone());
                    id
                }
            };
            let (part, pkg) = plan_part(&canonical_home, &id, skill, stage, allow_older)?;
            parts.push(part);
            staged.push(pkg);
        }
    }
    let plan_digest = crate::digest(&serde_json::to_vec(&parts).map_err(|_| "encode failed")?);
    Ok(Layer {
        parts,
        plan_digest,
        staged,
    })
}

/// The folder of `skill` under the Claude config folder of `home`.
fn place(home: &Path, skill: &str) -> Result<Place> {
    crate::host_home::place(Host::Claude, home, &format!("skills/{skill}"))
}

fn plan_part(
    home: &Path,
    install_id: &str,
    skill: &str,
    stage: Result<StagedPackage>,
    allow_older: bool,
) -> Result<(Part, Option<StagedPackage>)> {
    let place = place(home, skill);
    let root = place
        .as_ref()
        .map(Place::path)
        .unwrap_or_else(|_| home.join(".claude/skills").join(skill));
    let root_text = root.to_string_lossy().into_owned();
    let staged = match stage {
        Ok(s) => s,
        Err(reason) => {
            return Ok((
                Part {
                    skill: skill.to_string(),
                    host: HOST.to_string(),
                    projection_root: root_text,
                    preflight: "blocked".into(),
                    blocked_reason: Some(reason),
                    files: Vec::new(),
                    receipt_from: None,
                    going_back: None,
                    package_digest: None,
                },
                None,
            ))
        }
    };
    let (mut blocked, scan) = match &place {
        Err(e) => (Some(e.clone()), None),
        Ok(_) if root.to_str().is_none() => (
            Some(
                "the skill folder path is not UTF-8 (check CLAUDE_CONFIG_DIR); this part is \
                 left out"
                    .to_string(),
            ),
            None,
        ),
        Ok(p) => preflight(home, p),
    };
    let mut receipt_from = None;
    let mut receipt = None;
    let mut going_back = None;
    if blocked.is_none() {
        let source = resolve_receipt(home, install_id, skill)?;
        receipt_from = source.from;
        receipt = source.receipt;
        if let Some(reason) = source.blocked {
            blocked = Some(reason);
        } else if let Some(r) = receipt.as_ref().filter(|r| r.status == "pending") {
            blocked = Some(format!(
                "the {skill} receipt {} is pending; nothing is written to {root_text} until it \
                 is recovered",
                receipt_path(home, &r.install_id, skill, false)?.display()
            ));
        } else if let Some(r) = receipt
            .as_ref()
            .filter(|r| crate::headver::is_newer(&r.package_version, crate::BUILD_VERSION))
        {
            if allow_older {
                going_back = Some(format!(
                    "going back: the installed {skill} version {} is newer than this build's \
                     {}; --allow-older plans this older build over it",
                    r.package_version,
                    crate::BUILD_VERSION
                ));
            } else {
                blocked = Some(format!(
                    "the installed {skill} version {} is newer than this build's {}; this build \
                     never updates back unless asked: preview again with `moai update \
                     --allow-older`",
                    r.package_version,
                    crate::BUILD_VERSION
                ));
            }
        }
    }
    let scan = if blocked.is_some() { None } else { scan };
    let owned = owned_map(receipt.as_ref());
    let mut files = Vec::new();
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
        files.push(FilePlan {
            host: HOST.to_string(),
            path: rel.clone(),
            action: action.to_string(),
            baseline_sha256: baseline,
            current_sha256: current,
            new_sha256: Some(new_sha),
        });
    }
    // Files on disk the folder does not carry: never deleted, reported.
    if let Some(scan) = &scan {
        for (rel, current_sha) in &scan.files {
            if staged.files.contains_key(rel) {
                continue;
            }
            files.push(FilePlan {
                host: HOST.to_string(),
                path: rel.clone(),
                action: if owned.contains_key(rel) {
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
    // The ownership rule of the moai skill: a package file present with no baseline means a
    // folder moai did not install; the part is blocked and nothing is written there.
    let unmanaged: Vec<&str> = files
        .iter()
        .filter(|f| {
            f.baseline_sha256.is_none()
                && f.current_sha256.is_some()
                && staged.files.contains_key(&f.path)
        })
        .map(|f| f.path.as_str())
        .collect();
    if !unmanaged.is_empty() {
        blocked = Some(format!(
            "the folder {root_text} holds {} file(s) that moai did not install ({}); moai \
             leaves the folder as it is — move it aside yourself to let moai install {skill} \
             there",
            unmanaged.len(),
            unmanaged.join(", ")
        ));
    }
    let part = Part {
        skill: skill.to_string(),
        host: HOST.to_string(),
        projection_root: root_text,
        preflight: if blocked.is_some() { "blocked" } else { "ok" }.into(),
        blocked_reason: blocked,
        files,
        receipt_from,
        going_back,
        package_digest: Some(staged.digest.clone()),
    };
    Ok((part, Some(staged)))
}

/// The preflight of `update::preflight` for the folder `place`: links from the host folder down,
/// a writable parent for an absent folder, a clean scan, a writable folder and one filesystem.
fn preflight(home: &Path, place: &Place) -> (Option<String>, Option<ProjectionScan>) {
    let root = place.path();
    let base = &place.base;
    let top = if base == home {
        home.to_path_buf()
    } else {
        base.parent()
            .map_or_else(|| base.clone(), Path::to_path_buf)
    };
    if let Err(e) = ancestor_symlinks(&top, &root) {
        return (Some(e), None);
    }
    let meta = match fs::symlink_metadata(&root) {
        Err(_) => {
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

/// Refuses a symlink in any existing component between the state root and `dir`.
fn refuse_links_below_state(dir: &Path, what: &str) -> Result<()> {
    let mut cursor = dir;
    let mut parts: Vec<PathBuf> = Vec::new();
    while let Some(parent) = cursor.parent() {
        parts.push(cursor.to_path_buf());
        cursor = parent;
        if cursor
            .file_name()
            .is_some_and(crate::home_root::is_state_dir_name)
        {
            break;
        }
    }
    for part in parts.iter().rev() {
        if let Ok(m) = fs::symlink_metadata(part) {
            if m.file_type().is_symlink() {
                return Err(format!("installations {what} path is a symlink; refused"));
            }
        }
    }
    Ok(())
}

/// `installations/<install_id>/hosts/<skill>/claude.json`; `create` also creates the folders
/// (apply only).
fn receipt_path(home: &Path, install_id: &str, skill: &str, create: bool) -> Result<PathBuf> {
    let dir = crate::home_root::path(home)
        .join("installations")
        .join(install_id)
        .join("hosts")
        .join(skill);
    if create {
        refuse_links_below_state(&dir, "receipt")?;
        fs::create_dir_all(&dir).map_err(|_| "host receipt directory unavailable")?;
    }
    Ok(dir.join(format!("{HOST}.json")))
}

fn save_receipt(home: &Path, install_id: &str, skill: &str, receipt: &HostReceipt) -> Result<()> {
    let path = receipt_path(home, install_id, skill, true)?;
    let bytes =
        serde_json::to_vec_pretty(receipt).map_err(|_| "receipt encode failed".to_string())?;
    let before = read_opt(&path)?;
    publish_replace(&path, before.as_deref(), &bytes)
}

/// `update::resolve_host_receipt` for the receipts of `skill`: this installation's own receipt
/// (or one it moved from), else in the version layout the receipt of another version folder that
/// finished last. A pending receipt of an earlier folder that is newer than every finished one, an
/// unreadable one, or two folders finishing in the same second block the part.
fn resolve_receipt(canonical_home: &Path, install_id: &str, skill: &str) -> Result<ReceiptSource> {
    let mut own = load_host_receipt(&receipt_path(canonical_home, install_id, skill, false)?)?;
    if own.is_none() {
        for old in crate::home_install::earlier_ids_of(canonical_home, install_id) {
            let path = receipt_path(canonical_home, &old, skill, false)?;
            if let Ok(Some(r)) = load_host_receipt(&path) {
                let newer = own.as_ref().is_none_or(|o| {
                    r.finished_at.clone().unwrap_or(r.started_at.clone())
                        > o.finished_at.clone().unwrap_or(o.started_at.clone())
                });
                if r.install_id == old && r.host == HOST && newer {
                    own = Some(r);
                }
            }
        }
    }
    let plain = |receipt| ReceiptSource {
        receipt,
        from: None,
        blocked: None,
    };
    if own.as_ref().is_some_and(|r| r.status == "pending") {
        return Ok(plain(own));
    }
    let exe = std::env::current_exe().map_err(|_| "current executable unavailable".to_string())?;
    let exe = fs::canonicalize(exe).map_err(|_| "binary path unavailable".to_string())?;
    let exe = crate::install::identity_exe(exe);
    let Some((versions, own_folder)) = layout_folder(canonical_home, &exe) else {
        return Ok(plain(own));
    };
    let exe_name = exe.file_name().ok_or("binary path unavailable")?.to_owned();
    let installations = crate::home_root::path(canonical_home).join("installations");
    let blocked = |reason: String| ReceiptSource {
        receipt: None,
        from: None,
        blocked: Some(reason),
    };
    let finished_key = |r: &HostReceipt| r.finished_at.clone().unwrap_or(r.started_at.clone());
    let mut best: Option<(String, Option<String>, HostReceipt)> =
        own.map(|r| (finished_key(&r), None, r));
    let mut newest_pending: Option<(String, String)> = None;
    let mut tie = false;
    let entries = match fs::read_dir(&versions) {
        Ok(e) => e,
        Err(_) => return Ok(plain(best.map(|(_, _, r)| r))),
    };
    let mut folders: Vec<PathBuf> = Vec::new();
    for entry in entries {
        let entry = entry.map_err(|_| "versions directory unreadable".to_string())?;
        let name = PathBuf::from(entry.file_name());
        if name == own_folder {
            continue;
        }
        match fs::symlink_metadata(entry.path()) {
            Ok(m) if m.is_dir() => folders.push(name),
            _ => continue,
        }
    }
    folders.sort();
    for (_, folder_id) in earlier_install_ids(&versions, &folders, &exe_name)? {
        if folder_id == install_id {
            continue;
        }
        let mut ids = vec![folder_id.clone()];
        ids.extend(crate::home_install::earlier_ids_of(
            canonical_home,
            &folder_id,
        ));
        let mut chosen: Option<(String, HostReceipt)> = None;
        for id in ids {
            if id == install_id {
                continue;
            }
            match fs::symlink_metadata(installations.join(&id)) {
                Ok(m) if m.is_dir() => {}
                _ => continue,
            }
            let path = receipt_path(canonical_home, &id, skill, false)?;
            let receipt = match load_host_receipt(&path) {
                Ok(Some(r)) => r,
                Ok(None) => continue,
                Err(e) => {
                    return Ok(blocked(format!(
                        "the {skill} receipt of the earlier installation {id} is unreadable \
                         ({e}); nothing was adopted"
                    )))
                }
            };
            if receipt.install_id != id || receipt.host != HOST {
                return Ok(blocked(format!(
                    "the {skill} receipt of the earlier installation {id} names another \
                     installation or host; nothing was adopted"
                )));
            }
            if receipt.status == "pending" {
                if newest_pending
                    .as_ref()
                    .is_none_or(|(k, _)| receipt.started_at > *k)
                {
                    newest_pending = Some((receipt.started_at.clone(), id));
                }
                continue;
            }
            if chosen
                .as_ref()
                .is_none_or(|(_, c)| finished_key(&receipt) > finished_key(c))
            {
                chosen = Some((id, receipt));
            }
        }
        let Some((id, receipt)) = chosen else {
            continue;
        };
        let key = finished_key(&receipt);
        match &best {
            Some((k, _, _)) if key < *k => {}
            Some((k, from, _)) if key == *k => tie |= from.is_some(),
            _ => {
                tie = false;
                best = Some((key, Some(id), receipt));
            }
        }
    }
    if let Some((started, id)) = newest_pending {
        if best.as_ref().is_none_or(|(k, _, _)| started >= *k) {
            return Ok(blocked(format!(
                "the {skill} receipt of the earlier installation {id} is pending; run \
                 recovery-plan with that version's binary first"
            )));
        }
    }
    if tie {
        return Ok(blocked(
            "two earlier installations finished in the same second; the baseline is ambiguous \
             and nothing was adopted"
                .into(),
        ));
    }
    Ok(match best {
        Some((_, from, r)) => ReceiptSource {
            receipt: Some(r),
            from,
            blocked: None,
        },
        None => plain(None),
    })
}

/// Known-good copies of every file an `update` overwrites, before the transaction starts, under
/// `installations/<id>/hosts/<skill>/claude/known-good/<hex>` (`update::known_good_copies`).
fn known_good_copies(
    home: &Path,
    install_id: &str,
    skill: &str,
    root: &Path,
    writes: &[&FilePlan],
) -> Result<()> {
    for w in writes {
        if w.action != "update" {
            continue;
        }
        let bytes = read_opt(&root.join(&w.path))?.ok_or("known-good: update target missing")?;
        let hex = format!("{:x}", Sha256::digest(&bytes));
        let dir = crate::home_root::path(home)
            .join("installations")
            .join(install_id)
            .join("hosts")
            .join(skill)
            .join(HOST)
            .join("known-good");
        refuse_links_below_state(&dir, "known-good")?;
        fs::create_dir_all(&dir).map_err(|_| "known-good directory unavailable")?;
        let path = dir.join(&hex);
        if fs::symlink_metadata(&path).is_ok() {
            let m = fs::symlink_metadata(&path).map_err(|_| "known-good copy state unavailable")?;
            if !m.is_file() {
                return Err("known-good digest path is not a regular file".into());
            }
            let existing = fs::read(&path).map_err(|_| "known-good copy read failed")?;
            if format!("{:x}", Sha256::digest(&existing)) == hex {
                continue;
            }
            publish_replace(&path, Some(&existing), &bytes)?;
        } else {
            publish_exclusive(&path, &bytes)?;
        }
        let stored = fs::read(&path).map_err(|_| "known-good copy read failed")?;
        if format!("{:x}", Sha256::digest(&stored)) != hex {
            return Err("known-good copy does not hold its digest's bytes".into());
        }
    }
    Ok(())
}

/// Applies the confirmed parts. The plan is made again first and must equal `expected`; then each
/// part walks the moai skill's transaction: known-good copies → pending receipt → each file
/// published and read back → committed receipt. A blocked part is never written.
pub fn apply(
    home: &Path,
    package: &Path,
    expected: &Layer,
    permission: &str,
    allow_older: bool,
) -> Result<Value> {
    validate_permission(permission)?;
    if expected.parts.is_empty() {
        return Ok(json!({"state": "unchanged", "parts": []}));
    }
    let fresh = plan(home, package, allow_older)?;
    if fresh.parts != expected.parts {
        return Ok(json!({
            "state": "refused",
            "reason": "the plan of the skill folders beside the moai skill changed after the confirmed preview; nothing written",
        }));
    }
    let canonical_home = fs::canonicalize(home).map_err(|_| "home directory unavailable")?;
    crate::home_root::adopt_new(&canonical_home, "update-apply")?;
    let install_id = current_install_id(&canonical_home)?;
    let mut outcomes = Vec::new();
    let (mut published, mut failed, mut conflict) = (false, false, false);
    for (part, staged) in fresh.parts.iter().zip(&fresh.staged) {
        let o = match staged.as_ref().filter(|_| part.preflight != "blocked") {
            None => outcome(
                part,
                "blocked",
                Vec::new(),
                None,
                ("no", "no files written"),
            ),
            Some(staged) => apply_part(&canonical_home, &install_id, part, staged)?,
        };
        published |= o["status"] == json!("applied");
        failed |= o["status"] == json!("failed");
        conflict |= o["status"] == json!("conflict");
        outcomes.push(o);
    }
    let state = if failed {
        "partial"
    } else if outcomes.iter().all(|o| o["status"] == json!("blocked")) {
        "blocked"
    } else if published {
        "applied"
    } else if conflict {
        "conflict"
    } else {
        "unchanged"
    };
    Ok(json!({"state": state, "parts": outcomes}))
}

fn outcome(
    part: &Part,
    status: &str,
    files: Vec<Value>,
    receipt_status: Option<&str>,
    reload: (&str, &str),
) -> Value {
    json!({
        "skill": part.skill,
        "host": part.host,
        "status": status,
        "files": files,
        "receipt_status": receipt_status,
        "reload_required": reload.0,
        "reload_basis": reload.1,
        "blocked_reason": part.blocked_reason,
    })
}

fn apply_part(home: &Path, install_id: &str, part: &Part, staged: &StagedPackage) -> Result<Value> {
    let skill = part.skill.as_str();
    let root = place(home, skill)?.path();
    let writes: Vec<&FilePlan> = part
        .files
        .iter()
        .filter(|f| f.action == "create" || f.action == "update")
        .collect();
    if writes.is_empty() {
        let conflict = part.files.iter().any(|f| f.action == "conflict");
        let files = part
            .files
            .iter()
            .map(|f| file_result(f, false, None))
            .collect();
        let status = if conflict { "conflict" } else { "unchanged" };
        return Ok(outcome(
            part,
            status,
            files,
            None,
            ("no", "no files written"),
        ));
    }
    if let Err(e) = known_good_copies(home, install_id, skill, &root, &writes) {
        let basis = format!("known-good copy failed; part not started: {e}");
        return Ok(outcome(
            part,
            "failed",
            Vec::new(),
            None,
            ("no", basis.as_str()),
        ));
    }
    receipt_path(home, install_id, skill, true)?;
    let previous = resolve_receipt(home, install_id, skill)?.receipt;
    let pre_owned = owned_map(previous.as_ref());
    let top_skills_existed = root
        .parent()
        .is_some_and(|p| fs::symlink_metadata(p).is_ok());
    let mut receipt = HostReceipt {
        schema_version: crate::install::RECEIPT_SCHEMA_VERSION,
        install_id: install_id.to_string(),
        host: HOST.to_string(),
        operation: "update.apply".into(),
        plan_digest: crate::digest(&serde_json::to_vec(part).map_err(|_| "encode failed")?),
        package_digest: staged.digest.clone(),
        package_version: crate::BUILD_VERSION.to_string(),
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
        owned: owned_vec(HOST, &pre_owned),
        backup: None,
    };
    save_receipt(home, install_id, skill, &receipt)?;
    let mut owned_now = pre_owned.clone();
    let mut count = 0usize;
    let mut failed: Option<String> = None;
    for (idx, w) in writes.iter().enumerate() {
        let target = root.join(&w.path);
        let new_bytes = staged
            .files
            .get(&w.path)
            .ok_or_else(|| "package file disappeared".to_string())?;
        let before = match read_opt(&target) {
            Ok(b) => b,
            Err(e) => {
                failed = Some(e);
                receipt.entries[idx].error = failed.clone();
                break;
            }
        };
        if before.as_deref().map(sha256_hex) != w.current_sha256 {
            failed = Some("target changed during apply; temporary file discarded".into());
            receipt.entries[idx].error = failed.clone();
            break;
        }
        let result = if w.action == "create" {
            publish_exclusive(&target, new_bytes)
        } else {
            publish_replace(&target, before.as_deref(), new_bytes)
        };
        if let Err(e) = result {
            failed = Some(e);
            receipt.entries[idx].error = failed.clone();
            break;
        }
        let written = read_opt(&target)?.as_deref().map(sha256_hex);
        if written.as_deref() != w.new_sha256.as_deref() {
            failed = Some("published bytes do not match the plan".into());
            receipt.entries[idx].error = failed.clone();
            break;
        }
        receipt.entries[idx].published = true;
        owned_now.insert(w.path.clone(), written.expect("sha of a written file"));
        count += 1;
    }
    let files: Vec<Value> = writes
        .iter()
        .zip(&receipt.entries)
        .map(|(f, e)| file_result_with(f, e))
        .collect();
    if failed.is_some() {
        // The receipt stays pending with the ownership it had before this apply.
        receipt.finished_at = Some(now_rfc3339());
        save_receipt(home, install_id, skill, &receipt)?;
        return Ok(outcome(
            part,
            "failed",
            files,
            Some("pending"),
            (
                "unknown",
                "apply failed mid-transaction; the part's receipt is pending",
            ),
        ));
    }
    for w in &writes {
        let written = read_opt(&root.join(&w.path))?;
        if written.as_deref().map(sha256_hex).as_deref() != w.new_sha256.as_deref() {
            return Err("published bytes do not match the plan".into());
        }
    }
    receipt.owned = owned_vec(HOST, &owned_now);
    receipt.status = "committed".into();
    receipt.finished_at = Some(now_rfc3339());
    save_receipt(home, install_id, skill, &receipt)?;
    let (reload, basis) = reload_state(HOST, top_skills_existed, count > 0);
    let mut all = files;
    for f in part
        .files
        .iter()
        .filter(|f| f.action != "create" && f.action != "update")
    {
        all.push(file_result(f, false, None));
    }
    Ok(outcome(
        part,
        "applied",
        all,
        Some("committed"),
        (reload, basis.as_str()),
    ))
}
