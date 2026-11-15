//! Project-unit `moai` skill copy (Issue #2242, EP-I / P04 HOLD).
//!
//! 04-B `update.rs` copies the `moai` skill package into the HOME skill
//! folders only; project-level copies (`.claude/skills/moai`,
//! `.agents/skills/moai` inside a project) are the P04 HOLD. This module
//! designs and exercises plan/apply/revert for those project copies with two
//! hard limits:
//!
//! - REQ-I5-01: every operation carries `opt_in` and refuses before touching
//!   anything unless it is `true`. The default is off: nothing is written
//!   without the explicit flag, and a mutation that ignores the flag fails
//!   the tests (#2226).
//! - The P04 HOLD is NOT released here: the project root and home are
//!   caller-supplied parameters, so tests run in temporary directories and
//!   real HOME paths are never defaults. Applying to the developer's actual
//!   project/home stays blocked until the HOLD is released.
//!
//! Reading order (AC-2): when both the home copy and the project copy
//! exist, the project copy is what a session in that project reads — Claude
//! documents project skills taking precedence over personal ones and Codex
//! walks from the working directory up toward `$HOME/.agents/skills`, so
//! the nearest copy wins. Both copies stay independent per host (independent
//! copy per host; symlinks refused — design textonly §4 L166). The plan reports
//! the duplication instead of guessing.
//!
//! Rollback (AC-1 revert): apply stores the before bytes of every
//! replaced folder under `<project>/.moai-cli/project-skill-rollback/
//! <request-id>/before/<host>/` plus a receipt JSON; revert restores from
//! that record (a created folder is removed, a replaced folder is restored
//! byte for byte).

use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

pub const PROJECT_SKILL_SCHEMA_VERSION: u32 = 1;
/// The reading-order note carried on duplicated pairs (AC-2).
pub const READING_ORDER_NOTE: &str =
    "project copy takes precedence over the home copy for sessions in this project (host-documented discovery order)";

fn opt_in_error() -> String {
    "explicit opt-in required; nothing written".into()
}

fn host_folder(host: &str) -> Result<&'static str> {
    match host {
        "claude" => Ok(".claude/skills"),
        "codex" => Ok(".agents/skills"),
        other => Err(format!("unknown host {other}; supported: claude, codex")),
    }
}

/// Plan/apply/revert request (#2242). The project root and home are
/// caller-supplied so tests run in temporary directories; nothing defaults
/// to the real HOME.
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ProjectSkillRequest {
    pub protocol_version: u32,
    /// REQ-I5-01: must be true or every operation refuses before any write.
    pub opt_in: bool,
    /// Caller-supplied project root (tests pass a temporary directory).
    pub project_root: PathBuf,
    /// Caller-supplied home root; omitted means the home state is not scanned.
    pub home: Option<PathBuf>,
    /// claude | codex — which host folders to plan or copy.
    pub hosts: Vec<String>,
    /// Staged `moai` skill package source directory (caller-owned).
    pub source_dir: PathBuf,
}

/// One planned copy with the reading-order context (AC-2).
#[derive(Clone, Debug, Serialize)]
pub struct ProjectSkillEntry {
    pub host: String,
    /// Project skill folder, e.g. `.claude/skills/moai`.
    pub path: String,
    /// create | unchanged | replace | blocked
    pub action: String,
    /// Why the copy is `blocked` (decision row 22, #2286); omitted otherwise.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub blocked_reason: Option<String>,
    /// Staged package digest.
    pub source_sha256: String,
    /// Current project copy digest (None when absent).
    pub current_sha256: Option<String>,
    /// Home copy digest (None when absent or home not scanned).
    pub home_sha256: Option<String>,
    /// True when both the home and project copies exist.
    pub duplicated_with_home: bool,
    pub reading_order_note: String,
}

#[derive(Clone, Debug, Serialize)]
pub struct ProjectSkillPlan {
    pub protocol_version: u32,
    pub opt_in: bool,
    pub package_digest: String,
    pub entries: Vec<ProjectSkillEntry>,
    pub warnings: Vec<String>,
    /// Binds the project, the package and every entry; `project-skill-apply --plan-digest`
    /// takes it (#2936 R9 S5c).
    pub plan_digest: String,
}

fn project_skill_dir(root: &Path, host: &str) -> Result<PathBuf> {
    Ok(root.join(host_folder(host)?).join(crate::update::SKILL_DIR))
}

/// The first folder on `parts` below `root` that is a symbolic link, is not a folder, or
/// cannot be read (#2936 R9 S4): a copy written, or removed, through it would land outside the
/// project. Stops quietly at the first folder that does not exist yet — the copy makes it.
fn unsafe_parent(root: &Path, parts: &[&str]) -> Option<String> {
    let mut path = root.to_path_buf();
    let mut shown = String::new();
    for part in parts {
        path.push(part);
        if !shown.is_empty() {
            shown.push('/');
        }
        shown.push_str(part);
        match fs::symlink_metadata(&path) {
            Ok(m) if m.file_type().is_symlink() => {
                return Some(format!("{shown} is a symbolic link"));
            }
            Ok(m) if !m.is_dir() => return Some(format!("{shown} is not a folder")),
            Ok(_) => {}
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => return None,
            Err(_) => return Some(format!("{shown} cannot be read")),
        }
    }
    None
}

/// Why a copy for `host` may not be written or removed under `root`: a link, a file or an
/// unreadable folder among the host's skill folders or the rollback folders.
fn linked_parent(root: &Path, host: &str) -> Result<Option<String>> {
    let host_parts: Vec<&str> = host_folder(host)?.split('/').collect();
    let rollback_parts = [crate::state_root::dir(root), "project-skill-rollback"];
    Ok(unsafe_parent(root, &host_parts).or_else(|| unsafe_parent(root, &rollback_parts)))
}

fn validate(req: &ProjectSkillRequest) -> Result<()> {
    if req.protocol_version != PROJECT_SKILL_SCHEMA_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    if req.hosts.is_empty() || req.hosts.len() > 2 {
        return Err("hosts must list one or two entries of claude, codex".into());
    }
    if !req.hosts.iter().all(|h| host_folder(h).is_ok()) {
        return Err("unknown host in hosts; supported: claude, codex".into());
    }
    if req.source_dir.as_os_str().is_empty() {
        return Err("source directory missing".into());
    }
    Ok(())
}

/// Digest of an installed `moai` skill directory (None when absent). The
/// staged-package digest rules apply, so both sides of the comparison use
/// the same algorithm.
fn installed_digest(dir: &Path) -> Result<Option<String>> {
    if !dir.is_dir() {
        return Ok(None);
    }
    let staged = crate::update::stage_package(dir)?;
    Ok(Some(staged.digest))
}

/// `project-skill-plan` (#2242 AC-1): read-only judgment per requested host —
/// the project copy digest, the home copy digest (reading-order report only),
/// and the action (create | unchanged | replace) against the staged package.
/// Writes nothing.
pub fn project_skill_plan(req: &ProjectSkillRequest) -> Result<ProjectSkillPlan> {
    validate(req)?;
    let staged = crate::update::stage_package(&req.source_dir)?;
    let mut entries = Vec::new();
    let mut warnings = Vec::new();
    for host in &req.hosts {
        let folder = host_folder(host)?;
        let project_dir = project_skill_dir(&req.project_root, host)?;
        // #2936 R9 S4: a linked (or non-folder) skill or rollback folder is blocked, never
        // written through; nothing below it is read either.
        if let Some(why) = linked_parent(&req.project_root, host)? {
            entries.push(ProjectSkillEntry {
                host: host.clone(),
                path: format!("{folder}/{}", crate::update::SKILL_DIR),
                action: "blocked".into(),
                blocked_reason: Some(format!(
                    "{why}; a copy made through it would land outside the project, so nothing \
                     is written"
                )),
                source_sha256: staged.digest.clone(),
                current_sha256: None,
                home_sha256: None,
                duplicated_with_home: false,
                reading_order_note: String::new(),
            });
            continue;
        }
        // Decision row 22 (#2286): an existing folder is ours only when one of our own
        // rollback receipts recorded its current digest; anything else — including a
        // folder the digest cannot read — is blocked, never replaced.
        let exists = fs::symlink_metadata(&project_dir).is_ok();
        let project_digest = if exists {
            installed_digest(&project_dir).ok().flatten()
        } else {
            None
        };
        let (action, blocked_reason) = match project_digest.as_deref() {
            None if !exists => ("create", None),
            Some(d)
                if own_copy_digests(&req.project_root, host)
                    .iter()
                    .any(|o| o == d) =>
            {
                (
                    if d == staged.digest {
                        "unchanged"
                    } else {
                        "replace"
                    },
                    None,
                )
            }
            _ => (
                "blocked",
                Some(format!(
                    "existing {folder}/{} is not a copy this tool made (no own rollback \
                     receipt records its digest); {}",
                    crate::update::SKILL_DIR,
                    crate::update::GO_ERA_GUIDANCE
                )),
            ),
        };
        let mut home_sha256 = None;
        let mut duplicated = false;
        let mut note = String::new();
        if let Some(home) = &req.home {
            // Report only: a home folder the digest cannot read is reported absent.
            let home_digest = installed_digest(&project_skill_dir(home, host)?)
                .ok()
                .flatten();
            duplicated = home_digest.is_some() && project_digest.is_some();
            if duplicated {
                note = READING_ORDER_NOTE.into();
            }
            home_sha256 = home_digest;
        }
        entries.push(ProjectSkillEntry {
            host: host.clone(),
            path: format!("{folder}/{}", crate::update::SKILL_DIR),
            action: action.into(),
            blocked_reason,
            source_sha256: staged.digest.clone(),
            current_sha256: project_digest,
            home_sha256,
            duplicated_with_home: duplicated,
            reading_order_note: note,
        });
        if duplicated {
            warnings.push(format!(
                "both the home and project copies exist for {host}; the project copy takes precedence for sessions in this project"
            ));
        }
    }
    // The digest binds the project (its canonical path), the package and every entry, so a plan
    // of one project or package cannot be applied to another (#2936 R9 S5c).
    let root = req
        .project_root
        .canonicalize()
        .map_err(|_| "project directory unavailable".to_string())?;
    let plan_digest = plan_digest_of(&root, &staged.digest, &entries, &warnings)?;
    Ok(ProjectSkillPlan {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        opt_in: req.opt_in,
        package_digest: staged.digest,
        entries,
        warnings,
        plan_digest,
    })
}

/// What a plan digest binds: the canonical project path, the package digest, every entry in
/// full and the warnings.
fn plan_digest_of(
    root: &Path,
    package_digest: &str,
    entries: &[ProjectSkillEntry],
    warnings: &[String],
) -> Result<String> {
    let bound = serde_json::to_vec(&json!({
        "project": root.to_str().ok_or("non UTF-8 project path")?,
        "package_digest": package_digest,
        "entries": entries,
        "warnings": warnings,
    }))
    .map_err(|_| "encode failed".to_string())?;
    Ok(crate::digest(&bound))
}

/// [`project_skill_apply`] of a previewed plan (#2936 R9 S5c): a permission reference is
/// required and `plan_digest` must equal a fresh plan's, both checked before anything is
/// written. The digest binds the project and the package, so a plan of another project or of
/// a package changed since the preview is refused.
pub fn project_skill_apply_confirmed(
    req: &ProjectSkillRequest,
    plan_digest: &str,
    permission: &str,
) -> Result<ProjectSkillApplyReceipt> {
    confirmed_apply(req, plan_digest, permission, &|| {}, &|_| {})
}

/// [`project_skill_apply_confirmed`] with two seams for the tests: `between_plans` runs after
/// the digest check and before the apply plans again; `before_copy` runs before each host's copy.
fn confirmed_apply(
    req: &ProjectSkillRequest,
    plan_digest: &str,
    permission: &str,
    between_plans: &dyn Fn(),
    before_copy: &dyn Fn(&str),
) -> Result<ProjectSkillApplyReceipt> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    if !req.opt_in {
        return Err(opt_in_error());
    }
    let confirmed = project_skill_plan(req)?;
    if confirmed.plan_digest != plan_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    // The apply plans again; the plan it confirmed goes down with it, so a package or a project
    // that changed between the two plans is refused before anything is written (#3002 audit B1).
    between_plans();
    apply_core(req, Some(&confirmed), before_copy)
}

/// Copies the staged skill tree into `dest` with the same refusals the
/// digest applies (symlink/special/size), through a temp directory and an
/// atomic rename: the destination is replaced whole or not at all. The
/// destination must not exist when called (the caller decides create vs
/// replace). The copied tree must hash to `expected` (the digest the plan
/// bound) before it is renamed into place: a source that changed after the
/// preview is a stale plan, and nothing reaches `dest` (#2936 R9 S5c2).
/// Returns the number of files copied.
fn copy_tree_into_vacant(src: &Path, dest: &Path, expected: &str) -> Result<usize> {
    let parent = dest.parent().ok_or("destination parent unavailable")?;
    fs::create_dir_all(parent).map_err(|_| "destination directory unavailable".to_string())?;
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let temp = parent.join(format!(
        ".moai-skillcopy-{}-{nanos}.tmp",
        std::process::id()
    ));
    fs::create_dir(&temp).map_err(|_| "temporary directory refused".to_string())?;
    let mut count = 0usize;
    let copy_result = copy_tree(src, &temp, 0, &mut count);
    if let Err(e) = copy_result {
        let _ = fs::remove_dir_all(&temp);
        return Err(e);
    }
    if !matches!(crate::package_digest(&temp), Ok(d) if d == expected) {
        let _ = fs::remove_dir_all(&temp);
        return Err("stale or altered plan; re-preview required".into());
    }
    match fs::rename(&temp, dest) {
        Ok(()) => Ok(count),
        Err(_) => {
            let _ = fs::remove_dir_all(&temp);
            Err("conflict in skill copy: existing_skill".into())
        }
    }
}

fn copy_tree(src: &Path, dest: &Path, depth: usize, count: &mut usize) -> Result<()> {
    if depth > 16 {
        return Err("skill tree nesting limit exceeded".into());
    }
    fs::create_dir_all(dest).map_err(|_| "destination directory unavailable".to_string())?;
    for entry in fs::read_dir(src).map_err(|_| "source directory unavailable".to_string())? {
        let entry = entry.map_err(|_| "source directory unavailable".to_string())?;
        let meta = fs::symlink_metadata(entry.path())
            .map_err(|_| "source metadata unavailable".to_string())?;
        if meta.file_type().is_symlink() {
            return Err("package symlink refused".into());
        }
        if meta.is_dir() {
            copy_tree(
                &entry.path(),
                &dest.join(entry.file_name()),
                depth + 1,
                count,
            )?;
        } else if meta.is_file() {
            if meta.len() > 262_144 {
                return Err("package file exceeds the size limit".into());
            }
            fs::copy(entry.path(), dest.join(entry.file_name()))
                .map_err(|_| "destination write failed".to_string())?;
            *count += 1;
        } else {
            return Err("package special file refused".into());
        }
    }
    Ok(())
}

/// One applied copy with what revert needs (AC-1 revert): the action
/// taken, the before digest, and the backup directory holding the pre-apply
/// bytes for `replace` (a `create` is reverted by removing the directory).
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AppliedCopy {
    pub host: String,
    pub path: String,
    /// create | replace | unchanged
    pub action: String,
    pub before_sha256: Option<String>,
    pub after_sha256: Option<String>,
    /// Rollback-relative backup dir for `replace` (None otherwise).
    pub backup_dir: Option<String>,
    pub files_copied: usize,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ProjectSkillApplyReceipt {
    pub schema_version: u32,
    pub request_id: String,
    pub package_digest: String,
    pub applied: Vec<AppliedCopy>,
    /// complete | partial — partial names what landed before the failure.
    pub status: String,
}

/// Revert request (#2242 AC-1 revert): names the persisted rollback
/// receipt to undo.
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ProjectSkillRevertRequest {
    pub protocol_version: u32,
    pub project_root: PathBuf,
    /// The request_id from the apply receipt.
    pub request_id: String,
}

fn rollback_root(project_root: &Path) -> PathBuf {
    crate::state_root::path(project_root, "project-skill-rollback")
}

/// Every `after_sha256` our own rollback receipts recorded for `host` — the digests of
/// copies this tool made. Unreadable or foreign receipts grant no ownership.
fn own_copy_digests(project_root: &Path, host: &str) -> Vec<String> {
    let Ok(dirs) = fs::read_dir(rollback_root(project_root)) else {
        return Vec::new();
    };
    dirs.filter_map(|e| e.ok())
        .filter_map(|e| fs::read(e.path().join("receipt.json")).ok())
        .filter_map(|b| serde_json::from_slice::<ProjectSkillApplyReceipt>(&b).ok())
        .flat_map(|r| r.applied)
        .filter(|a| a.host == host)
        .filter_map(|a| a.after_sha256)
        .collect()
}

/// `project-skill-apply` (#2242, REQ-I5-01 gate): re-plans, refuses without
/// `opt_in` before any write, refuses when the plan changed since preview,
/// then copies each requested host folder (backup → remove → rename for
/// `replace`) and persists the rollback receipt under
/// `<project>/.moai-cli/project-skill-rollback/<request_id>/receipt.json`.
pub fn project_skill_apply(req: &ProjectSkillRequest) -> Result<ProjectSkillApplyReceipt> {
    apply_core(req, None, &|_| {})
}

/// The apply. `confirmed` is the plan a caller already checked its digest against: this
/// function's own plan must be the same plan, and the package the copies are hashed against is
/// that plan's. `before_copy` runs before each host's copy, with the host's name, once the
/// package is staged (a seam for the tests: a package that changes at that point).
fn apply_core(
    req: &ProjectSkillRequest,
    confirmed: Option<&ProjectSkillPlan>,
    before_copy: &dyn Fn(&str),
) -> Result<ProjectSkillApplyReceipt> {
    if !req.opt_in {
        // REQ-I5-01: the default is off — refuse before any write.
        return Err(opt_in_error());
    }
    validate(req)?;
    let plan = project_skill_plan(req)?;
    if let Some(c) = confirmed {
        if plan.plan_digest != c.plan_digest || plan.package_digest != c.package_digest {
            return Err("stale or altered plan; re-preview required".into());
        }
    }
    // Decision row 22: a plan with any blocked host is refused whole, before any write
    // (the rollback directory included).
    if let Some(blocked) = plan.entries.iter().find(|e| e.action == "blocked") {
        return Err(format!(
            "project skill copy blocked for {}: {}; nothing written",
            blocked.host,
            blocked.blocked_reason.as_deref().unwrap_or("blocked")
        ));
    }
    let root = req
        .project_root
        .canonicalize()
        .map_err(|_| "project directory unavailable".to_string())?;
    let staged = crate::update::stage_package(&req.source_dir)?;
    // The plan above read the package once and this staging reads it again; a package that
    // changed between the two is a stale plan, refused before the rollback folder exists.
    if staged.digest != plan.package_digest {
        return Err("stale or altered plan; re-preview required".into());
    }
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let request_id = format!("project-skill-{}-{nanos}", &staged.digest[..12]);
    let rollback_dir = rollback_root(&root).join(&request_id);
    let rollback_made = missing_dirs(&root, &rollback_dir);
    fs::create_dir_all(&rollback_dir).map_err(|_| "rollback directory unavailable".to_string())?;

    let mut applied: Vec<AppliedCopy> = Vec::new();
    let step_failed =
        |applied: &[AppliedCopy], rollback_dir: &Path, host: &str, e: &str| -> String {
            // REQ-01B-06 style: persist what landed (status partial) so revert
            // can undo it; if even the receipt cannot be written, say so.
            let receipt = ProjectSkillApplyReceipt {
                schema_version: PROJECT_SKILL_SCHEMA_VERSION,
                request_id: String::new(),
                package_digest: String::new(),
                applied: applied.to_vec(),
                status: "partial".into(),
            };
            let receipt = serde_json::to_vec_pretty(&receipt)
                .map(|mut b| {
                    b.push(b'\n');
                    b
                })
                .unwrap_or_default();
            let receipt_path = rollback_dir.join("receipt.json");
            let note = match fs::write(&receipt_path, &receipt) {
                Ok(()) => format!("rollback receipt {}", rollback_dir.display()),
                Err(se) => format!("receipt not recorded: {se}"),
            };
            format!(
                "partial apply: applied [{}]; failed at {host}: {e}; {note}",
                applied
                    .iter()
                    .map(|a| a.host.as_str())
                    .collect::<Vec<_>>()
                    .join(", ")
            )
        };

    for entry in &plan.entries {
        let folder = host_folder(&entry.host)?;
        let dest = project_skill_dir(&root, &entry.host)?;
        match entry.action.as_str() {
            "unchanged" => {
                applied.push(AppliedCopy {
                    host: entry.host.clone(),
                    path: format!("{folder}/{}", crate::update::SKILL_DIR),
                    action: "unchanged".into(),
                    before_sha256: entry.current_sha256.clone(),
                    after_sha256: Some(staged.digest.clone()),
                    backup_dir: None,
                    files_copied: 0,
                });
            }
            "create" => {
                before_copy(&entry.host);
                let host_made = dest
                    .parent()
                    .map(|p| missing_dirs(&root, p))
                    .unwrap_or_default();
                match copy_tree_into_vacant(&req.source_dir, &dest, &plan.package_digest) {
                    Ok(count) => {
                        applied.push(AppliedCopy {
                            host: entry.host.clone(),
                            path: format!("{folder}/{}", crate::update::SKILL_DIR),
                            action: "create".into(),
                            before_sha256: None,
                            after_sha256: Some(staged.digest.clone()),
                            backup_dir: None,
                            files_copied: count,
                        });
                    }
                    Err(e) => {
                        if is_stale(&e) && nothing_landed(&applied) {
                            take_back(&rollback_dir, &rollback_made, &host_made);
                            return Err(e);
                        }
                        applied.push(AppliedCopy {
                            host: entry.host.clone(),
                            path: format!("{folder}/{}", crate::update::SKILL_DIR),
                            action: "create".into(),
                            before_sha256: None,
                            after_sha256: None,
                            backup_dir: None,
                            files_copied: 0,
                        });
                        return Err(step_failed(&applied, &rollback_dir, &entry.host, &e));
                    }
                }
            }
            "replace" => {
                // replace: keep the before bytes for revert, move the old tree
                // aside atomically, then rename the staged tree in.
                let backup_rel = format!("before/{folder}");
                let backup_abs = rollback_dir.join(&backup_rel);
                fs::create_dir_all(&backup_abs)
                    .map_err(|_| "rollback backup directory unavailable".to_string())?;
                let mut backup_count = 0usize;
                let backup_result = copy_tree(&dest, &backup_abs, 0, &mut backup_count);
                if let Err(e) = backup_result {
                    return Err(step_failed(
                        &applied,
                        &rollback_dir,
                        &entry.host,
                        &format!("rollback backup failed: {e}"),
                    ));
                }
                before_copy(&entry.host);
                let replaced_dir = rollback_dir.join(format!("replaced/{folder}"));
                fs::create_dir_all(replaced_dir.parent().expect("replaced parent"))
                    .map_err(|_| "rollback replaced directory unavailable".to_string())?;
                fs::rename(&dest, &replaced_dir)
                    .map_err(|_| "existing skill directory move failed".to_string())?;
                match copy_tree_into_vacant(&req.source_dir, &dest, &plan.package_digest) {
                    Ok(count) => {
                        applied.push(AppliedCopy {
                            host: entry.host.clone(),
                            path: format!("{folder}/{}", crate::update::SKILL_DIR),
                            action: "replace".into(),
                            before_sha256: entry.current_sha256.clone(),
                            after_sha256: Some(staged.digest.clone()),
                            backup_dir: Some(backup_rel.clone()),
                            files_copied: count,
                        });
                    }
                    Err(e) => {
                        // Roll the move back: the old tree returns to its place.
                        let _ = fs::remove_dir_all(&dest);
                        let _ = fs::rename(&replaced_dir, &dest);
                        if is_stale(&e) && nothing_landed(&applied) {
                            take_back(&rollback_dir, &rollback_made, &[]);
                            return Err(e);
                        }
                        return Err(step_failed(&applied, &rollback_dir, &entry.host, &e));
                    }
                }
            }
            other => {
                return Err(step_failed(
                    &applied,
                    &rollback_dir,
                    &entry.host,
                    &format!("unknown plan action {other}"),
                ));
            }
        }
    }

    let receipt = ProjectSkillApplyReceipt {
        schema_version: PROJECT_SKILL_SCHEMA_VERSION,
        request_id: request_id.clone(),
        package_digest: staged.digest,
        applied: applied.clone(),
        status: "complete".to_string(),
    };
    let receipt_path = rollback_dir.join("receipt.json");
    let bytes = serde_json::to_vec_pretty(&receipt).map_err(|_| "encode failed".to_string())?;
    fs::write(&receipt_path, bytes).map_err(|_| "receipt write failed".to_string())?;
    Ok(receipt)
}

fn is_stale(error: &str) -> bool {
    error.starts_with("stale or altered plan")
}

/// Nothing of this apply is in the project yet: no earlier host was created or replaced.
fn nothing_landed(applied: &[AppliedCopy]) -> bool {
    applied.iter().all(|a| a.action == "unchanged")
}

/// The folders on the way to `dir` below `root` that do not exist yet, deepest first: what a
/// `create_dir_all` of `dir` would make.
fn missing_dirs(root: &Path, dir: &Path) -> Vec<PathBuf> {
    let mut made = Vec::new();
    let mut at = dir.to_path_buf();
    while at != root && at.starts_with(root) && fs::symlink_metadata(&at).is_err() {
        made.push(at.clone());
        if !at.pop() {
            break;
        }
    }
    made
}

/// Takes back what an apply made when it stops before any copy landed (a stale plan found
/// while copying): its rollback folder and the empty folders it made on the way. A folder that
/// holds anything else stays (`remove_dir` refuses it).
fn take_back(rollback_dir: &Path, rollback_made: &[PathBuf], host_made: &[PathBuf]) {
    let _ = fs::remove_dir_all(rollback_dir);
    for dir in rollback_made.iter().chain(host_made) {
        let _ = fs::remove_dir(dir);
    }
}

/// `project-skill-revert` (#2242 AC-1 revert): restores what one apply
/// changed — a created folder is removed, a replaced folder is restored
/// byte for byte from the rollback backup. Reads only the persisted
/// receipt; executes only local file moves.
pub fn project_skill_revert(req: &ProjectSkillRevertRequest) -> Result<Value> {
    if req.protocol_version != PROJECT_SKILL_SCHEMA_VERSION {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let project_root = req.project_root.clone();
    let request_id = req.request_id.clone();
    if request_id.is_empty()
        || request_id.contains("..")
        || request_id.contains('/')
        || request_id.contains('\\')
    {
        return Err("invalid request id; refused".into());
    }
    let root = project_root
        .canonicalize()
        .map_err(|_| "project directory unavailable".to_string())?;
    // #2936 R9 S4: a rollback or skill folder that became a link, or a file, would send the
    // removal and the moves below outside the project; refuse before anything is read or moved.
    if let Some(why) = unsafe_parent(
        &root,
        &[crate::state_root::dir(&root), "project-skill-rollback"],
    ) {
        return Err(format!("{why}; nothing reverted"));
    }
    let rollback_dir = rollback_root(&root).join(request_id.as_str());
    let receipt_path = rollback_dir.join("receipt.json");
    let bytes = fs::read(&receipt_path)
        .map_err(|_| "receipt not found; inspect the rollback directory".to_string())?;
    let receipt: ProjectSkillApplyReceipt = serde_json::from_slice(&bytes)
        .map_err(|_| "receipt unreadable; inspect the rollback directory".to_string())?;
    for a in &receipt.applied {
        if let Some(why) = linked_parent(&root, &a.host)? {
            return Err(format!("{why}; nothing reverted"));
        }
    }
    // #2392 D3-A: a created folder is removed only while it is still exactly what the apply
    // wrote. A copy that changed since — a file the user added or edited, or a folder the
    // digest cannot read — is never removed, and the whole revert is refused before anything
    // moves, so no host is left half reverted.
    let mut changed = Vec::new();
    for a in receipt.applied.iter().filter(|a| a.action == "create") {
        let dest = project_skill_dir(&root, &a.host)?;
        if dest.is_dir() && installed_digest(&dest).ok().flatten() != a.after_sha256 {
            changed.push(a.host.clone());
        }
    }
    if !changed.is_empty() {
        return Err(format!(
            "the project copy for {} changed after apply; nothing reverted — move your own \
             files out of that skill folder, then revert again",
            changed.join(", ")
        ));
    }
    let mut reverted = Vec::new();
    for a in receipt.applied.iter().rev() {
        let dest = project_skill_dir(&root, &a.host)?;
        match a.action.as_str() {
            "create" => {
                if dest.is_dir() {
                    fs::remove_dir_all(&dest).map_err(|_| "revert remove failed".to_string())?;
                    reverted.push(format!("{} (created, removed)", a.host));
                }
            }
            "replace" => {
                let backup_abs = rollback_dir.join(
                    a.backup_dir
                        .as_deref()
                        .ok_or("receipt backup missing; inspect the rollback directory")?,
                );
                if !backup_abs.is_dir() {
                    return Err("receipt backup missing; inspect the rollback directory".into());
                }
                let trash = rollback_dir.join(format!("post-revert/{}", a.host));
                fs::create_dir_all(&trash).map_err(|_| "revert staging unavailable".to_string())?;
                // Move the applied tree aside, then move the backup tree back.
                fs::rename(&dest, trash.join("applied"))
                    .map_err(|_| "revert move failed".to_string())?;
                fs::rename(&backup_abs, &dest).map_err(|_| "revert restore failed".to_string())?;
                reverted.push(format!("{} (replaced, restored)", a.host));
            }
            "unchanged" => {}
            other => return Err(format!("unknown action {other} in the receipt")),
        }
    }
    Ok(json!({
        "state": "reverted",
        "request_id": request_id.clone(),
        "reverted": reverted,
        "network": "none",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn request(
        project: &Path,
        home: Option<&Path>,
        source: &Path,
        opt_in: bool,
    ) -> ProjectSkillRequest {
        ProjectSkillRequest {
            protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
            opt_in,
            project_root: project.to_path_buf(),
            home: home.map(|h| h.to_path_buf()),
            hosts: vec!["claude".into()],
            source_dir: source.to_path_buf(),
        }
    }

    fn write_source(dir: &Path) {
        fs::create_dir_all(dir.join("references")).unwrap();
        fs::write(
            dir.join("SKILL.md"),
            b"---\nname: moai\ndescription: fixture\n---\n",
        )
        .unwrap();
        fs::write(dir.join("references/a.md"), b"a").unwrap();
    }

    /// #2998 audit follow-up 2: the copy checks the bytes it wrote against the digest the plan
    /// bound. A source that changes after the preview is never renamed into place, and no
    /// temporary folder is left behind.
    #[test]
    fn a_source_that_changed_after_its_digest_is_not_copied() {
        let t = tempfile::tempdir().unwrap();
        let (src, dest) = (
            t.path().join("src"),
            t.path().join("out/.agents/skills/moai"),
        );
        write_source(&src);
        let digest = crate::package_digest(&src).unwrap();
        // The same bytes copy.
        assert_eq!(copy_tree_into_vacant(&src, &dest, &digest).unwrap(), 2);
        fs::remove_dir_all(&dest).unwrap();
        // Another byte after the digest was taken does not.
        fs::write(src.join("references/a.md"), b"changed").unwrap();
        let err = copy_tree_into_vacant(&src, &dest, &digest).unwrap_err();
        assert!(err.contains("stale or altered plan"), "{err}");
        assert!(!dest.exists());
        let left: Vec<_> = fs::read_dir(dest.parent().unwrap()).unwrap().collect();
        assert!(left.is_empty(), "a temporary folder stayed: {left:?}");
    }

    fn no_hook(_host: &str) {}

    /// #3002 audit B1: the confirmed apply plans twice (the digest check, then the apply's own
    /// plan). The digest it confirmed must reach the second plan, so a package that changed
    /// between the two is refused before anything is written.
    #[test]
    fn a_package_that_changed_between_the_two_plans_is_not_applied() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), None, source.path(), true);
        let confirmed = project_skill_plan(&req).unwrap();
        fs::write(source.path().join("references/a.md"), b"another byte").unwrap();
        let err = apply_core(&req, Some(&confirmed), &no_hook).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::StalePlan,
            "{err}"
        );
        assert!(
            fs::read_dir(project.path()).unwrap().next().is_none(),
            "something was written"
        );
        // The old call, which confirms nothing, plans for itself and applies.
        assert!(apply_core(&req, None, &no_hook).is_ok());
    }

    /// A project folder that appeared after the preview changes the plan; the confirmed apply
    /// calls that a stale plan (not a blocked copy) and leaves the folder as it is.
    #[test]
    fn a_project_folder_that_appeared_after_the_preview_is_a_stale_plan() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), None, source.path(), true);
        let confirmed = project_skill_plan(&req).unwrap();
        let mine = project.path().join(".claude/skills/moai");
        fs::create_dir_all(&mine).unwrap();
        fs::write(mine.join("SKILL.md"), b"mine").unwrap();
        let err = apply_core(&req, Some(&confirmed), &no_hook).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::StalePlan,
            "{err}"
        );
        assert_eq!(fs::read(mine.join("SKILL.md")).unwrap(), b"mine");
        assert!(!project.path().join(".moai-cli").exists());
    }

    /// A package that changes after the apply staged it is refused when the copy is hashed. With
    /// nothing landed yet this is a plain stale plan, and the folders this apply made (the
    /// rollback folder and the host's skill folder) are gone again.
    #[test]
    fn a_copy_that_fails_its_hash_with_nothing_landed_leaves_nothing_behind() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), None, source.path(), true);
        let altered =
            |_host: &str| fs::write(source.path().join("references/a.md"), b"late").unwrap();
        let err = apply_core(&req, None, &altered).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::StalePlan,
            "{err}"
        );
        assert!(
            fs::read_dir(project.path()).unwrap().next().is_none(),
            "left behind: {:?}",
            fs::read_dir(project.path()).unwrap().collect::<Vec<_>>()
        );
    }

    /// The same, for a host folder that already holds this tool's copy: the old tree comes back
    /// byte for byte and the rollback folder is removed.
    #[test]
    fn a_failed_replace_with_nothing_landed_restores_the_folder_and_leaves_no_rollback() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), None, source.path(), true);
        apply_core(&req, None, &no_hook).unwrap();
        fs::write(source.path().join("references/a.md"), b"v2").unwrap();
        let mine = project.path().join(".claude/skills/moai");
        let before = fs::read(mine.join("references/a.md")).unwrap();
        let rollbacks = |p: &Path| {
            fs::read_dir(p.join(".moai-cli/project-skill-rollback"))
                .unwrap()
                .count()
        };
        assert_eq!(rollbacks(project.path()), 1);
        let altered =
            |_host: &str| fs::write(source.path().join("references/a.md"), b"v3").unwrap();
        let err = apply_core(&req, None, &altered).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::StalePlan,
            "{err}"
        );
        assert_eq!(fs::read(mine.join("references/a.md")).unwrap(), before);
        assert_eq!(
            rollbacks(project.path()),
            1,
            "the failed apply left a rollback folder"
        );
    }

    /// The confirmed call hands its plan to the apply: a package that changes between the digest
    /// check and the apply's own plan is refused, and nothing is written.
    #[test]
    fn the_confirmed_call_refuses_a_package_changed_between_its_two_plans() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), None, source.path(), true);
        let digest = project_skill_plan(&req).unwrap().plan_digest;
        let change = || fs::write(source.path().join("references/a.md"), b"between").unwrap();
        let err = confirmed_apply(&req, &digest, "test:ref", &change, &no_hook).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::StalePlan,
            "{err}"
        );
        assert!(
            fs::read_dir(project.path()).unwrap().next().is_none(),
            "something was written"
        );
        // Nothing changed in between: the same call applies.
        let digest = project_skill_plan(&req).unwrap().plan_digest;
        assert!(confirmed_apply(&req, &digest, "test:ref", &|| {}, &no_hook).is_ok());
    }

    /// A stale plan found while copying the second host, after the first host landed, is still a
    /// partial apply with a receipt: what landed stays reported and revertable.
    #[test]
    fn a_stale_copy_after_an_earlier_host_landed_is_a_partial_apply() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let mut req = request(project.path(), None, source.path(), true);
        req.hosts = vec!["claude".into(), "codex".into()];
        let late = |host: &str| {
            if host == "codex" {
                fs::write(source.path().join("references/a.md"), b"late").unwrap();
            }
        };
        let err = apply_core(&req, None, &late).unwrap_err();
        assert_eq!(
            crate::contract::classify(&err),
            crate::contract::Kind::Partial,
            "{err}"
        );
        assert!(project
            .path()
            .join(".claude/skills/moai/SKILL.md")
            .is_file());
        assert!(!project.path().join(".agents/skills/moai").exists());
        let receipts: Vec<_> =
            fs::read_dir(project.path().join(".moai-cli/project-skill-rollback"))
                .unwrap()
                .collect();
        assert_eq!(receipts.len(), 1, "{receipts:?}");
    }

    /// Taking back what a refused apply made never removes a folder that was there before it.
    #[test]
    fn a_refused_copy_keeps_the_empty_folders_the_project_already_had() {
        let project = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), None, source.path(), true);
        fs::create_dir_all(project.path().join(".claude/skills")).unwrap();
        fs::create_dir_all(project.path().join(".moai-cli")).unwrap();
        let altered =
            |_host: &str| fs::write(source.path().join("references/a.md"), b"late").unwrap();
        assert!(apply_core(&req, None, &altered).is_err());
        assert!(project.path().join(".claude/skills").is_dir());
        assert!(project.path().join(".moai-cli").is_dir());
        assert!(!project.path().join(".claude/skills/moai").exists());
        assert!(!project
            .path()
            .join(".moai-cli/project-skill-rollback")
            .exists());
    }

    /// #3002 audit B3 (mutants M13, M14, M27): the plan digest binds the project, the package,
    /// every entry in full and the warnings.
    #[test]
    fn the_plan_digest_binds_project_package_entries_and_warnings() {
        let entry = |action: &str| ProjectSkillEntry {
            host: "claude".into(),
            path: ".claude/skills/moai".into(),
            action: action.into(),
            blocked_reason: None,
            source_sha256: "sha256:p".into(),
            current_sha256: None,
            home_sha256: None,
            duplicated_with_home: false,
            reading_order_note: String::new(),
        };
        let digest =
            |project: &str, package: &str, entries: &[ProjectSkillEntry], warnings: &[String]| {
                plan_digest_of(Path::new(project), package, entries, warnings).unwrap()
            };
        let base = digest("/p", "sha256:p", &[entry("create")], &[]);
        assert_eq!(base, digest("/p", "sha256:p", &[entry("create")], &[]));
        assert_ne!(
            base,
            digest("/q", "sha256:p", &[entry("create")], &[]),
            "project"
        );
        assert_ne!(
            base,
            digest("/p", "sha256:q", &[entry("create")], &[]),
            "package"
        );
        assert_ne!(
            base,
            digest("/p", "sha256:p", &[entry("replace")], &[]),
            "entry action"
        );
        let mut other = entry("create");
        other.current_sha256 = Some("sha256:c".into());
        assert_ne!(
            base,
            digest("/p", "sha256:p", &[other], &[]),
            "entry digest"
        );
        assert_ne!(
            base,
            digest("/p", "sha256:p", &[entry("create")], &["w".into()]),
            "warnings"
        );
    }

    #[test]
    fn opt_in_off_writes_nothing_and_apply_revert_round_trip() {
        let project = tempfile::TempDir::new().unwrap();
        let home = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let req = request(project.path(), Some(home.path()), source.path(), false);

        // Plan is read-only even with opt-in off.
        let plan = project_skill_plan(&req).unwrap();
        assert_eq!(plan.entries.len(), 1);
        assert_eq!(plan.entries[0].action, "create");

        // REQ-I5-01: apply without opt-in refuses before any write.
        let err = project_skill_apply(&req).unwrap_err();
        assert_eq!(err, opt_in_error());
        assert!(
            !project.path().join(".claude/skills/moai").exists(),
            "nothing may be written without opt-in"
        );

        // The reverted-and-opted-in flow creates the copy.
        let mut opted = req.clone();
        opted.opt_in = true;
        let plan = project_skill_plan(&opted).unwrap();
        assert_eq!(plan.entries[0].action, "create");
        let receipt = project_skill_apply(&opted).unwrap();
        assert_eq!(receipt.status, "complete");
        assert!(project.path().join(".claude/skills/moai/SKILL.md").exists());

        // Replace with changed source, then revert restores the first copy.
        let source2 = tempfile::TempDir::new().unwrap();
        write_source(source2.path());
        fs::write(
            source2.path().join("SKILL.md"),
            b"---\nname: moai\ndescription: v2\n---\n",
        )
        .unwrap();
        let req2 = request(project.path(), Some(home.path()), source2.path(), true);
        let plan2 = project_skill_plan(&req2).unwrap();
        assert_eq!(plan2.entries[0].action, "replace");
        let applied2 = project_skill_apply(&req2).unwrap();
        assert_eq!(applied2.status, "complete");
        let replaced =
            fs::read_to_string(project.path().join(".claude/skills/moai/SKILL.md")).unwrap();
        assert!(replaced.contains("v2"));

        let revert_req = ProjectSkillRevertRequest {
            protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
            project_root: project.path().to_path_buf(),
            request_id: applied2.request_id.clone(),
        };
        let reverted = project_skill_revert(&revert_req).unwrap();
        assert_eq!(reverted["state"], json!("reverted"));
        let restored =
            fs::read_to_string(project.path().join(".claude/skills/moai/SKILL.md")).unwrap();
        assert!(restored.contains("fixture"));
    }

    /// A Go-era `moai` skill in the HOME is only reported (the home is not written by
    /// this module); the project copy is created beside it and the home tree stays
    /// byte-identical. A Go-era copy inside the project is covered by
    /// tests/skill_name_collision.rs (decision row 22).
    #[test]
    fn project_copy_leaves_home_go_skills_untouched() {
        let project = tempfile::TempDir::new().unwrap();
        let home = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let mut go_trees = Vec::new();
        for host in [".claude/skills", ".agents/skills"] {
            let go = home.path().join(host).join("moai");
            fs::create_dir_all(&go).unwrap();
            fs::write(go.join("SKILL.md"), b"personal Go skill\n").unwrap();
            let before = crate::tree_digest(&go, &[]).unwrap();
            go_trees.push((go, before));
        }
        let mut req = request(project.path(), Some(home.path()), source.path(), true);
        req.hosts = vec!["claude".into(), "codex".into()];
        let plan = project_skill_plan(&req).unwrap();
        assert!(plan.entries.iter().all(|e| e.path.ends_with("/moai")));
        assert!(plan.entries.iter().all(|e| e.action == "create"));
        assert!(plan.entries.iter().all(|e| !e.duplicated_with_home));
        assert_eq!(project_skill_apply(&req).unwrap().status, "complete");
        for host in [".claude/skills", ".agents/skills"] {
            assert!(project.path().join(host).join("moai/SKILL.md").is_file());
        }
        for (go, before) in go_trees {
            assert_eq!(crate::tree_digest(&go, &[]).unwrap(), before);
        }
    }

    #[test]
    fn unknown_host_and_missing_home_are_typed_errors() {
        let root = tempfile::TempDir::new().unwrap();
        let source = tempfile::TempDir::new().unwrap();
        write_source(source.path());
        let mut req = request(root.path(), None, source.path(), true);
        req.hosts = vec!["zcode".into()];
        assert!(project_skill_plan(&req).is_err());
        req.hosts = vec!["claude".into()];
        // apply without home: home reporting is skipped, apply still works.
        assert!(project_skill_apply(&req).is_ok());
    }
}
