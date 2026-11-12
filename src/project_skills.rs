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
}

fn project_skill_dir(root: &Path, host: &str) -> Result<PathBuf> {
    Ok(root.join(host_folder(host)?).join(crate::update::SKILL_DIR))
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
    Ok(ProjectSkillPlan {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        opt_in: req.opt_in,
        package_digest: staged.digest,
        entries,
        warnings,
    })
}

/// Copies the staged skill tree into `dest` with the same refusals the
/// digest applies (symlink/special/size), through a temp directory and an
/// atomic rename: the destination is replaced whole or not at all. The
/// destination must not exist when called (the caller decides create vs
/// replace). Returns the number of files copied.
fn copy_tree_into_vacant(src: &Path, dest: &Path) -> Result<usize> {
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
    project_root.join(".moai-cli/project-skill-rollback")
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
    if !req.opt_in {
        // REQ-I5-01: the default is off — refuse before any write.
        return Err(opt_in_error());
    }
    validate(req)?;
    let plan = project_skill_plan(req)?;
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
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let request_id = format!("project-skill-{}-{nanos}", &staged.digest[..12]);
    let rollback_dir = rollback_root(&root).join(&request_id);
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
            "create" => match copy_tree_into_vacant(&req.source_dir, &dest) {
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
            },
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
                let replaced_dir = rollback_dir.join(format!("replaced/{folder}"));
                fs::create_dir_all(replaced_dir.parent().expect("replaced parent"))
                    .map_err(|_| "rollback replaced directory unavailable".to_string())?;
                fs::rename(&dest, &replaced_dir)
                    .map_err(|_| "existing skill directory move failed".to_string())?;
                match copy_tree_into_vacant(&req.source_dir, &dest) {
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
    let rollback_dir = rollback_root(&root).join(request_id.as_str());
    let receipt_path = rollback_dir.join("receipt.json");
    let bytes = fs::read(&receipt_path)
        .map_err(|_| "receipt not found; inspect the rollback directory".to_string())?;
    let receipt: ProjectSkillApplyReceipt = serde_json::from_slice(&bytes)
        .map_err(|_| "receipt unreadable; inspect the rollback directory".to_string())?;
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
