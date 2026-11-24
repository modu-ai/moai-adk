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
//! there), and a newer installed version blocks it unless `--allow-older` was given. The plan also
//! blocks the part when a path a write goes through has a shape the apply would fail on (a file
//! where a folder is needed, a folder where a file is, a folder moai cannot write in), so a confirm
//! does not stop halfway on it. The part is planned and applied on its own: a blocked moai skill
//! copy does not stop it, and a blocked part does not stop the moai skill.
//!
//! An apply that still fails halfway undoes its own files: the files it created are removed (only
//! while they hold the bytes it wrote) with the folders it made, the files it replaced get their
//! earlier bytes back, and the receipt is marked `recovered`. `recovery-plan` does not read these
//! receipts, so only an undo that itself fails leaves a receipt pending; a pending receipt blocks
//! the part, and its reason names the receipt, the folder and the way out by hand.
//!
//! Receipts live in their own namespace, `installations/<id>/hosts/<skill>/claude.json`, with
//! known-good copies under `hosts/<skill>/claude/known-good`. The receipt path, the receipt choice
//! and the known-good copies of `update.rs` name the moai folder; this module keeps copies of them
//! that take the skill name.

use super::{
    ancestor_symlinks, cross_filesystem, current_install_id, earlier_install_ids, file_result,
    file_result_with, folder_receipt_ids, layout_folder, load_host_receipt, now_rfc3339, owned_map,
    owned_vec, reload_state, scan_projection, sha256_hex, stage_package, test_fault,
    validate_permission, writable, FilePlan, HostReceipt, HostReceiptEntry, ProjectionScan,
    ReceiptSource, StagedPackage,
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
    /// Whether the top skills folder of each part (same order) existed when this plan was made.
    /// The confirm makes its plan before any apply step, so the receipt records the folder as it
    /// was before the moai skill's apply of the same run could create it.
    top_existed: Vec<bool>,
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
    let mut top_existed = Vec::new();
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
            top_existed.push(
                Path::new(&part.projection_root)
                    .parent()
                    .is_some_and(|p| fs::symlink_metadata(p).is_ok()),
            );
            parts.push(part);
            staged.push(pkg);
        }
    }
    let plan_digest = crate::digest(&serde_json::to_vec(&parts).map_err(|_| "encode failed")?);
    Ok(Layer {
        parts,
        plan_digest,
        staged,
        top_existed,
    })
}

/// The folder of `skill` under the Claude config folder of `home`.
fn place(home: &Path, skill: &str) -> Result<Place> {
    crate::host_home::place(Host::Claude, home, &format!("skills/{skill}"))
}

/// The path of [`place`], or where it would be under `~/.claude` when it cannot be resolved.
fn folder(home: &Path, skill: &str) -> PathBuf {
    place(home, skill)
        .as_ref()
        .map(Place::path)
        .unwrap_or_else(|_| home.join(".claude/skills").join(skill))
}

/// The blocked reason of a pending receipt. `recovery-plan` does not read the receipts of these
/// skills, so the reason gives the way out by hand.
fn pending_reason(skill: &str, receipt: &Path, of: &str, root: &Path) -> String {
    format!(
        "the {skill} receipt {}{of} is pending: an apply stopped and could not undo its own \
         files, so moai writes nothing to {} while it is pending; compare that folder with the \
         receipt's entries, move the folder aside, then remove the receipt to let moai install \
         {skill} again",
        receipt.display(),
        root.display()
    )
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
            blocked = Some(pending_reason(
                skill,
                &receipt_path(home, &r.install_id, skill, false)?,
                "",
                &root,
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
    if blocked.is_none() {
        blocked = shape_blocker(&root, skill, &files);
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
/// a writable folder above an absent one, a clean scan, a writable folder and one filesystem.
/// Every reason names the skill folder.
fn preflight(home: &Path, place: &Place) -> (Option<String>, Option<ProjectionScan>) {
    let root = place.path();
    let shown = root.display();
    let refuse = |reason: String| (Some(reason), None);
    let base = &place.base;
    let top = if base == home {
        home.to_path_buf()
    } else {
        base.parent()
            .map_or_else(|| base.clone(), Path::to_path_buf)
    };
    let meta = fs::symlink_metadata(&root);
    if meta.as_ref().is_ok_and(|m| m.file_type().is_symlink()) {
        return refuse(format!("the skill folder {shown} is a symlink; refused"));
    }
    if ancestor_symlinks(&top, &root).is_err() {
        return refuse(format!(
            "a folder above the skill folder {shown} is a symlink; refused"
        ));
    }
    let meta = match meta {
        Err(_) => {
            let mut cursor = match root.parent() {
                Some(p) => p.to_path_buf(),
                None => return refuse(format!("the skill folder path {shown} is invalid")),
            };
            loop {
                if let Ok(m) = fs::symlink_metadata(&cursor) {
                    if !m.is_dir() {
                        return refuse(format!(
                            "{} is not a folder, but the skill folder {shown} goes below it; \
                             refused",
                            cursor.display()
                        ));
                    }
                    if !writable(&cursor) {
                        return refuse(format!(
                            "the folder {} that would hold the skill folder {shown} is not \
                             writable; refused",
                            cursor.display()
                        ));
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
    if !meta.is_dir() {
        return refuse(format!(
            "the skill folder path {shown} is not a folder; refused"
        ));
    }
    let scan = match scan_projection(&root) {
        Ok(s) => s,
        Err(e) => return refuse(format!("the skill folder {shown}: {e}")),
    };
    if !writable(&root) {
        return refuse(format!("the skill folder {shown} is not writable; refused"));
    }
    if cross_filesystem(home, &root).unwrap_or(false) {
        return refuse(format!(
            "the skill folder {shown} is on another filesystem than the receipts; receipt and \
             files must share one filesystem"
        ));
    }
    (None, Some(scan))
}

/// The first path a write of this plan would fail on, as a blocked reason: below `root`, a part
/// of a written file's path that exists but is not a folder, a folder there that moai cannot
/// write in, or the file's own name taken by something that is not a regular file. The plan
/// blocks the part on it, so a confirm never stops halfway on a shape it could have seen. The
/// scan has already refused links and special files inside the folder.
fn shape_blocker(root: &Path, skill: &str, files: &[FilePlan]) -> Option<String> {
    let aside = format!("move it aside to let moai install {skill} there");
    for f in files
        .iter()
        .filter(|f| f.action == "create" || f.action == "update")
    {
        let names: Vec<&str> = f.path.split('/').collect();
        let mut cursor = root.to_path_buf();
        for (i, name) in names.iter().enumerate() {
            cursor.push(name);
            let Ok(meta) = fs::symlink_metadata(&cursor) else {
                // Absent from here down: the apply creates it.
                break;
            };
            let shown = cursor.display();
            let last = i + 1 == names.len();
            if meta.file_type().is_symlink() {
                return Some(format!(
                    "{shown} is a symlink, and the package file {} goes there; {aside}",
                    f.path
                ));
            }
            if last && !meta.is_file() {
                return Some(format!(
                    "{shown} is not a regular file, but the package puts the file {} there; \
                     {aside}",
                    f.path
                ));
            }
            if !last && !meta.is_dir() {
                return Some(format!(
                    "{shown} is not a folder, but the package file {} goes below it; {aside}",
                    f.path
                ));
            }
            if !last && !writable(&cursor) {
                return Some(format!(
                    "the folder {shown} is not writable, and the package file {} goes in it; \
                     refused",
                    f.path
                ));
            }
        }
    }
    None
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
    for (folder, folder_id) in earlier_install_ids(&versions, &folders, &exe_name)? {
        if folder_id == install_id {
            continue;
        }
        let ids = folder_receipt_ids(canonical_home, &versions, &folder, &folder_id, &exe_name);
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
            return Ok(blocked(pending_reason(
                skill,
                &receipt_path(canonical_home, &id, skill, false)?,
                &format!(" of the earlier installation {id}"),
                &folder(canonical_home, skill),
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
    let (mut published, mut failed, mut conflict, mut stuck) = (false, false, false, false);
    // The top folder as the confirmed plan saw it, before any apply step of this run.
    for ((part, staged), top_existed) in fresh
        .parts
        .iter()
        .zip(&fresh.staged)
        .zip(&expected.top_existed)
    {
        let o = match staged.as_ref().filter(|_| part.preflight != "blocked") {
            None => outcome(
                part,
                "blocked",
                Vec::new(),
                None,
                ("no", "no files written"),
            ),
            Some(staged) => apply_part(&canonical_home, &install_id, part, staged, *top_existed)?,
        };
        published |= o["status"] == json!("applied");
        failed |= o["status"] == json!("failed");
        conflict |= o["status"] == json!("conflict");
        stuck |= o["receipt_status"] == json!("pending");
        outcomes.push(o);
    }
    // A failed part whose undo put everything back wrote nothing: only a pending receipt, or
    // another part that did write, makes the layer partial.
    let state = if stuck || (failed && published) {
        "partial"
    } else if failed {
        "failed"
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

/// One part's transaction. `top_skills_existed` is the top skills folder as the confirmed plan
/// saw it, before any apply step of this run.
fn apply_part(
    home: &Path,
    install_id: &str,
    part: &Part,
    staged: &StagedPackage,
    top_skills_existed: bool,
) -> Result<Value> {
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
    // Measured before the first publish: only these folders are this apply's to remove.
    let made = missing_folders(&root, &writes);
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
    // The bytes each published `update` replaced, for the undo.
    let mut replaced: Vec<Option<Vec<u8>>> = vec![None; writes.len()];
    for (idx, w) in writes.iter().enumerate() {
        let target = root.join(&w.path);
        let Some(new_bytes) = staged.files.get(&w.path) else {
            failed = Some("package file disappeared".into());
            receipt.entries[idx].error = failed.clone();
            break;
        };
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
        // Test seam: `MOAI_TEST_FAULT=extra_skill_last_publish` fails the last publish as if
        // the OS refused it (debug builds only).
        let result = if idx + 1 == writes.len() {
            test_fault("extra_skill_last_publish")
        } else {
            Ok(())
        }
        .and_then(|()| {
            if w.action == "create" {
                publish_exclusive(&target, new_bytes)
            } else {
                publish_replace(&target, before.as_deref(), new_bytes)
            }
        });
        if let Err(e) = result {
            failed = Some(e);
            receipt.entries[idx].error = failed.clone();
            break;
        }
        // Published from here: a read-back that fails or differs leaves the file to the undo,
        // which takes back only the bytes this apply wrote.
        receipt.entries[idx].published = true;
        replaced[idx] = before;
        let written = match read_opt(&target) {
            Ok(b) => b.as_deref().map(sha256_hex),
            Err(e) => {
                failed = Some(e);
                receipt.entries[idx].error = failed.clone();
                break;
            }
        };
        if written.as_deref() != w.new_sha256.as_deref() {
            failed = Some("published bytes do not match the plan".into());
            receipt.entries[idx].error = failed.clone();
            break;
        }
        owned_now.insert(w.path.clone(), written.expect("sha of a written file"));
        count += 1;
    }
    if failed.is_none() {
        for w in &writes {
            let written = read_opt(&root.join(&w.path)).ok().flatten();
            if written.as_deref().map(sha256_hex).as_deref() != w.new_sha256.as_deref() {
                failed = Some(format!(
                    "published bytes of {} do not match the plan",
                    w.path
                ));
                break;
            }
        }
    }
    let files: Vec<Value> = writes
        .iter()
        .zip(&receipt.entries)
        .map(|(f, e)| file_result_with(f, e))
        .collect();
    if let Some(why) = failed {
        // The receipt keeps the ownership it had before this apply in either case.
        let undo = undo_own_writes(&root, &writes, &receipt.entries, &replaced, &made);
        receipt.finished_at = Some(now_rfc3339());
        let path = receipt_path(home, install_id, skill, false)?;
        let (status, reload, basis) = match undo {
            Ok(()) => {
                receipt.status = "recovered".into();
                (
                    "recovered",
                    "no",
                    format!(
                        "apply failed ({why}); the files this apply created were removed and \
                         the ones it replaced got their earlier bytes back, so {} is as it was; \
                         the receipt {} is recovered and the next preview plans the part again",
                        root.display(),
                        path.display()
                    ),
                )
            }
            Err(e) => (
                "pending",
                "unknown",
                format!(
                    "apply failed ({why}) and undoing its own files failed too ({e}); {}",
                    pending_reason(skill, &path, "", &root)
                ),
            ),
        };
        save_receipt(home, install_id, skill, &receipt)?;
        return Ok(outcome(
            part,
            "failed",
            files,
            Some(status),
            (reload, basis.as_str()),
        ));
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

/// The folders from `root` down to each write's folder that do not exist yet, deepest first:
/// the ones this apply creates, and so the only ones its undo may remove.
fn missing_folders(root: &Path, writes: &[&FilePlan]) -> Vec<PathBuf> {
    let mut made: Vec<PathBuf> = Vec::new();
    for w in writes {
        let mut dir = root.join(&w.path);
        while let Some(parent) = dir.parent() {
            if !parent.starts_with(root) || fs::symlink_metadata(parent).is_ok() {
                break;
            }
            if !made.iter().any(|m| m == parent) {
                made.push(parent.to_path_buf());
            }
            dir = parent.to_path_buf();
        }
    }
    made.sort_by_key(|p| std::cmp::Reverse(p.components().count()));
    made
}

/// Undoes this apply's published writes, last first: a file it created is removed, a file it
/// replaced gets the bytes it replaced back, each only while it still holds the bytes this apply
/// wrote; a symlink found there is left alone. Then the folders this apply made are removed while
/// they are empty. A file this apply did not write is never touched. The error names every file
/// left as it is.
fn undo_own_writes(
    root: &Path,
    writes: &[&FilePlan],
    entries: &[HostReceiptEntry],
    replaced: &[Option<Vec<u8>>],
    made: &[PathBuf],
) -> Result<()> {
    let mut left = Vec::new();
    for (idx, w) in writes.iter().enumerate().rev() {
        if !entries[idx].published {
            continue;
        }
        let target = root.join(&w.path);
        let step = || -> Result<()> {
            if fs::symlink_metadata(&target).is_ok_and(|m| m.file_type().is_symlink()) {
                return Err("a symlink is there now".into());
            }
            let current = read_opt(&target)?;
            if current.as_deref().map(sha256_hex).as_deref() != w.new_sha256.as_deref() {
                return Err("it no longer holds the bytes this apply wrote".into());
            }
            match (w.action.as_str(), &replaced[idx]) {
                ("create", _) => fs::remove_file(&target).map_err(|_| "remove failed".to_string()),
                ("update", Some(before)) => publish_replace(&target, current.as_deref(), before),
                _ => Err("the replaced bytes are unknown".into()),
            }
        };
        if let Err(e) = step() {
            left.push(format!("{} ({e})", w.path));
        }
    }
    for dir in made {
        // `remove_dir` removes only an empty folder: whatever another writer put there stays.
        let _ = fs::remove_dir(dir);
    }
    if left.is_empty() {
        Ok(())
    } else {
        Err(format!("left as it is: {}", left.join(", ")))
    }
}
