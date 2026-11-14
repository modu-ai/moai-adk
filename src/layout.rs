//! P09 opt-in home layout (#2175): the receipt, outbox and claim-journal stores can
//! move to `~/.moai-cli/projects/<project_id>/` — copy without deleting the originals,
//! verify, then switch the pointer (REQ-B8-01). The pointer is `layout.json` inside the
//! home project directory: while it exists, the three stores resolve to the home side
//! and the project-side directories stay as read-compatible frozen history. Removing
//! the marker reverts every reader and writer to the project side (the originals were
//! never touched), so the default remains the project side until an explicit opt-in.
//!
//! `layout-plan` measures and binds a plan (writes nothing); `layout-apply` executes it
//! (copy → verify → switch). The home tree this defines is
//! `~/.moai-cli/projects/<project_id>/{receipts/, forge/outbox/, claims/, layout.json}` —
//! the P09 decision document lives in `docs/home-layout-p09.md`.
use crate::profile::read_profile;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// The three store kinds this layout moves: (kind, project-side directory relative to
/// the project's state folder — `.moai-cli` or, once migrated, `.moai` (#2725 S2b) —, and
/// home-side directory relative to `~/.moai-cli/projects/<project_id>/`). The claim
/// journal's project-side root is the git common dir (`<common>/.moai-cli/claims`),
/// resolved per clone, so its middle entry is not used.
pub const KINDS: &[(&str, &str, &str)] = &[
    ("receipts", "receipts", "receipts"),
    ("outbox", "forge/outbox", "forge/outbox"),
    ("claims", "claims", "claims"),
];

/// The default home root: `$HOME/.moai-cli` (the card's fixed location). `None` without
/// a HOME — callers then stay on the project side.
pub fn default_home_root() -> Option<PathBuf> {
    std::env::var_os("HOME")
        .map(PathBuf::from)
        .filter(|h| !h.as_os_str().is_empty())
        .map(|h| crate::home_root::path(&h))
}

/// The active home project directory — `Some` only when `layout.json` (the pointer a
/// verified apply wrote) exists there. No profile, no HOME root, or no marker means the
/// project side stays authoritative (AC-2: the default paths are unchanged).
pub fn active_home(project: &Path, home_root: &Path) -> Option<PathBuf> {
    let (profile, _) = read_profile(project).ok()?;
    let dir = home_root.join("projects").join(&profile.project_id);
    dir.join("layout.json").is_file().then_some(dir)
}

/// The `$HOME/.moai-cli` wrapper store modules use: the directory a store kind resolves
/// to once the pointer is switched, `None` keeps the project side authoritative.
pub fn switched_dir(project: &Path, subdir: &str) -> Option<PathBuf> {
    let home_root = default_home_root()?;
    active_home(project, &home_root).map(|dir| dir.join(subdir))
}

/// One store entry of a layout plan: what would move, how big it is, and where it goes.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct LayoutEntry {
    pub kind: String,
    /// Project-side source directory (empty when the store has no files yet).
    pub source: String,
    /// Home-side target directory (under `~/.moai-cli/projects/<project_id>/`).
    pub target: String,
    pub files: u64,
    pub bytes: u64,
    /// `moai-tree-v1` digest of the source tree (empty when absent) — the apply result
    /// must reproduce it on the home side before the pointer switches.
    pub source_digest: String,
}

/// A bound layout plan. `plan_digest` covers the sorted entries, so any change to the
/// source tree between plan and apply invalidates the confirmed digest (stale guard).
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
pub struct LayoutPlan {
    pub schema_version: u32,
    pub project_id: String,
    pub home_root: String,
    pub entries: Vec<LayoutEntry>,
    pub plan_digest: String,
}

/// The claim journal's project-side root lives on the git common dir, not the project
/// root. Resolves per clone; absent for a non-repository or common-dir-less input.
fn claims_legacy_dir(project: &Path) -> Result<Option<PathBuf>> {
    let id = crate::identity::identity(project)?;
    let Some(common) = id["clone_anchor"].as_str() else {
        return Ok(None);
    };
    source_dir(crate::clone_root::path(Path::new(common)).join("claims"))
}

fn source_dir(dir: PathBuf) -> Result<Option<PathBuf>> {
    match std::fs::metadata(&dir) {
        Ok(meta) if meta.is_dir() => Ok(Some(dir)),
        Ok(_) => Err("source is not a directory".into()),
        Err(error) if error.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("source metadata unavailable".into()),
    }
}

fn legacy_dir(project: &Path, kind: &str) -> Result<Option<PathBuf>> {
    let Some(rel) = KINDS
        .iter()
        .find(|(k, _, _)| *k == kind)
        .map(|(_, legacy, _)| *legacy)
    else {
        return Ok(None);
    };
    if kind == "claims" {
        return claims_legacy_dir(project);
    }
    source_dir(crate::state_root::path(project, rel))
}

fn home_dir(home_root: &Path, project_id: &str, kind: &str) -> PathBuf {
    let sub = KINDS
        .iter()
        .find(|(k, _, _)| *k == kind)
        .map(|(_, _, sub)| *sub)
        .unwrap_or(kind);
    home_root.join("projects").join(project_id).join(sub)
}

fn measure_fb(root: &Path, excludes: &[&str]) -> (u64, u64, u64) {
    let (mut files, mut bytes, mut errors) = (0u64, 0u64, 0u64);
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        let entries = match std::fs::read_dir(&dir) {
            Ok(entries) => entries,
            Err(_) => {
                errors += 1;
                continue;
            }
        };
        for entry in entries {
            let entry = match entry {
                Ok(entry) => entry,
                Err(_) => {
                    errors += 1;
                    continue;
                }
            };
            let p = entry.path();
            let rel = p
                .strip_prefix(root)
                .map(|r| r.to_string_lossy())
                .unwrap_or_default();
            if excluded(rel.as_ref(), excludes) {
                continue;
            }
            if p.is_dir() {
                stack.push(p);
            } else {
                match entry.metadata() {
                    Ok(meta) => {
                        files += 1;
                        bytes += meta.len();
                    }
                    Err(_) => errors += 1,
                }
            }
        }
    }
    (files, bytes, errors)
}

/// The journal's `.lock` is a serialization handle, not store data: it moves neither in
/// the plan measure nor the digest nor the copy, or the apply's own `Journal::open`
/// would make every plan stale.
fn excluded(rel: &str, excludes: &[&str]) -> bool {
    excludes.contains(&rel)
}

fn kind_excludes(kind: &str) -> &'static [&'static str] {
    if kind == "claims" {
        &[".lock"]
    } else {
        &[]
    }
}

fn entry_for(
    project: &Path,
    home_root: &Path,
    kind: &str,
    project_id: &str,
) -> Result<LayoutEntry> {
    let target = home_dir(home_root, project_id, kind);
    let excludes = kind_excludes(kind);
    Ok(match legacy_dir(project, kind)? {
        Some(source) => {
            let (files, bytes, errors) = measure_fb(&source, excludes);
            if errors != 0 {
                return Err(format!("{errors} entry inspection errors"));
            }
            let source_digest = crate::tree_digest(&source, excludes)
                .map_err(|_| "source digest unavailable".to_string())?;
            LayoutEntry {
                kind: kind.to_string(),
                source: source.to_string_lossy().into_owned(),
                target: target.to_string_lossy().into_owned(),
                files,
                bytes,
                source_digest,
            }
        }
        None => LayoutEntry {
            kind: kind.to_string(),
            source: String::new(),
            target: target.to_string_lossy().into_owned(),
            files: 0,
            bytes: 0,
            source_digest: String::new(),
        },
    })
}

/// Measures the three stores and binds the plan. Writes nothing.
pub fn plan(project: &Path, home_root: &Path) -> Result<LayoutPlan> {
    let project_id = identity_of(project)?;
    let mut entries = Vec::with_capacity(KINDS.len());
    let mut errors = Vec::new();
    for (kind, _, _) in KINDS {
        match entry_for(project, home_root, kind, &project_id) {
            Ok(entry) => entries.push(entry),
            Err(error) => errors.push(format!("{kind}: {error}")),
        }
    }
    if !errors.is_empty() {
        return Err(format!(
            "layout source unreadable: {}; nothing applied",
            errors.join("; ")
        ));
    }
    entries.sort_by(|a, b| a.kind.cmp(&b.kind));
    let canonical = serde_json::to_vec(&entries).map_err(|_| "encode failed".to_string())?;
    Ok(LayoutPlan {
        schema_version: SCHEMA_VERSION,
        project_id,
        home_root: home_root.to_string_lossy().into_owned(),
        plan_digest: crate::digest(&canonical),
        entries,
    })
}

fn identity_of(project: &Path) -> Result<String> {
    let (profile, _) = read_profile(project)?;
    Ok(profile.project_id)
}

fn check_destinations(project: &Path, home_root: &Path) -> Result<()> {
    let project_id = identity_of(project)?;
    let home_project = home_root.join("projects").join(&project_id);
    if home_project.join("layout.json").is_file() {
        return Err(
            "conflict in home layout: already switched; remove the home layout.json marker to revert first"
                .into(),
        );
    }
    for (kind, _, subdir) in KINDS {
        match std::fs::symlink_metadata(home_project.join(subdir)) {
            Ok(_) => {
                return Err(format!(
                    "conflict in home layout: home target for {kind} already exists; refusing overwrite; remove or merge it first; nothing applied"
                ));
            }
            Err(error) if error.kind() == std::io::ErrorKind::NotFound => {}
            Err(_) => return Err("home layout target unavailable".into()),
        }
    }
    Ok(())
}

/// Copies one source tree to its target without touching the source, then verifies the
/// copy reproduces the source digest (REQ-B8-01: copy + verify). The entry's excludes (the
/// journal `.lock`) are skipped in both the walk and the verify, matching the plan.
fn copy_verified(entry: &LayoutEntry) -> Result<()> {
    let source = Path::new(&entry.source);
    let target = Path::new(&entry.target);
    let excludes = kind_excludes(&entry.kind);
    if let Some(parent) = target.parent() {
        std::fs::create_dir_all(parent)
            .map_err(|_| "home layout target unavailable".to_string())?;
    }
    // Reserve this target exclusively: a concurrent creator never becomes our copy.
    std::fs::create_dir(target).map_err(|_| "home layout target unavailable".to_string())?;
    let mut stack = vec![source.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in std::fs::read_dir(&dir).map_err(|_| "layout source unreadable".to_string())? {
            let from = e
                .map_err(|_| "layout source unreadable".to_string())?
                .path();
            let rel = from
                .strip_prefix(source)
                .map_err(|_| "layout walk failed".to_string())?;
            if excluded(rel.to_string_lossy().as_ref(), excludes) {
                continue;
            }
            let to = target.join(rel);
            if from.is_dir() {
                stack.push(from);
                std::fs::create_dir_all(&to)
                    .map_err(|_| "home layout target unavailable".to_string())?;
            } else {
                std::fs::copy(&from, &to).map_err(|_| "home layout copy failed".to_string())?;
            }
        }
    }
    let copied = crate::tree_digest(target, excludes)
        .map_err(|_| "home layout verify failed".to_string())?;
    if copied != entry.source_digest {
        // The verified copy failed: remove the partial copy, keep the source intact.
        // Safe since repair (a): the target never pre-existed, so everything under it
        // was created by this apply.
        let _ = std::fs::remove_dir_all(target);
        return Err("home layout verify failed; partial copy removed".into());
    }
    Ok(())
}

/// Executes a confirmed plan: validate the permission reference, refuse when the home
/// targets already exist, hold the claim journal's advisory lock across the
/// plan→copy→switch window, then copy every entry, verify the copy reproduces the
/// source digest, and switch the pointer (write `layout.json`). The originals are never
/// deleted (REQ-B8-01). A plan whose digest no longer matches the tree is refused
/// (stale); a second apply while the pointer is already switched is a conflict.
///
/// Repairs from the 2026-11-09 audit (issue #2175 FAIL):
/// - direction (a): a home target that already exists refuses the apply before any
///   write. The home side may hold data written while the pointer was switched during
///   an earlier apply — copying over it, or deleting it on a failed verify, destroyed
///   the only copy. Nothing is overwritten or removed; the caller merges or removes the
///   home side first, and a revert→re-apply sequence therefore cannot lose data.
/// - the claim journal lock: a claim-acquire inside the copy window used to land on the
///   project side and be missed by the switch; under the lock it sees busy (REQ-06A-05,
///   never waits) and retries after the switch, landing on the home side.
pub fn apply_at(
    project: &Path,
    home_root: &Path,
    confirmed: &str,
    permission: &str,
) -> std::result::Result<Value, crate::contract::OpError> {
    // Caller authorization remains host-owned; this reference is not an authorization verifier.
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    check_destinations(project, home_root)?;
    // The journal lock spans plan→copy→switch so the window closes for claimants.
    let journal_lock = crate::worktree::open_locked_journal(project)?;
    Ok(apply_locked(project, home_root, confirmed, &journal_lock)?)
}

fn apply_locked(
    project: &Path,
    home_root: &Path,
    confirmed: &str,
    _journal_lock: &crate::claim::Journal,
) -> Result<Value> {
    // The fast preflight is not authoritative across lock acquisition.
    check_destinations(project, home_root)?;
    let fresh = plan(project, home_root)?;
    if fresh.plan_digest != confirmed {
        return Err("stale or altered plan; re-preview required".into());
    }
    let mut copied = Vec::new();
    for entry in &fresh.entries {
        if entry.files == 0 {
            continue;
        }
        copy_verified(entry)?;
        copied.push(json!({"kind": entry.kind, "files": entry.files, "bytes": entry.bytes}));
    }
    let marker = home_root
        .join("projects")
        .join(&fresh.project_id)
        .join("layout.json");
    if let Some(parent) = marker.parent() {
        std::fs::create_dir_all(parent)
            .map_err(|_| "home layout target unavailable".to_string())?;
    }
    let marker_value = json!({
        "schema_version": SCHEMA_VERSION,
        "project_id": fresh.project_id,
        "plan_digest": fresh.plan_digest,
        "entries": fresh.entries.iter().map(|e| json!({"kind": e.kind, "target": e.target})).collect::<Vec<_>>(),
    });
    let bytes =
        serde_json::to_vec_pretty(&marker_value).map_err(|_| "encode failed".to_string())?;
    std::fs::write(&marker, bytes).map_err(|_| "home layout marker write failed".to_string())?;
    Ok(json!({
        "switched": true,
        "copied": copied,
        "marker": marker.to_string_lossy(),
        "project_id": fresh.project_id,
        "project_side_kept": true,
    }))
}

/// The per-store active directory, used by the three stores' readers/writers once the
/// pointer is switched (`None` = project side stays authoritative).
pub fn store_dir(project: &Path, home_root: &Path, kind: &str) -> Option<PathBuf> {
    let project_id = identity_of(project).ok()?;
    active_home(project, home_root).map(|_| home_dir(home_root, &project_id, kind))
}

/// CLI wrappers over the `$HOME/.moai-cli` default root.
pub fn plan_cli(project: &Path) -> Result<Value> {
    let home_root = default_home_root().ok_or("home root unavailable; HOME is not set")?;
    let plan = plan(project, &home_root)?;
    serde_json::to_value(&plan).map_err(|_| "encode failed".to_string())
}

pub fn apply_cli(
    project: &Path,
    confirmed: &str,
    permission: &str,
) -> std::result::Result<Value, crate::contract::OpError> {
    let home_root = default_home_root().ok_or("home root unavailable; HOME is not set")?;
    apply_at(project, &home_root, confirmed, permission)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    #[test]
    fn plan_refuses_without_a_profile() {
        let tmp = tempfile::tempdir().unwrap();
        let err = plan(&tmp.path().join("p"), tmp.path()).unwrap_err();
        assert!(
            err.contains("profile") || err.contains("unavailable"),
            "{err}"
        );
    }

    /// A minimal profile-bearing project (no git needed: the target guard fires before
    /// the journal lock is taken).
    fn fixture_project(root: &Path) -> PathBuf {
        let project = root.join("project");
        fs::create_dir_all(project.join(crate::PROJECT_STATE_DIR)).unwrap();
        fs::write(
            project.join(crate::project_state!("project.yaml")),
            r#"{"schema_version": 1, "project_id": "widgets", "purpose": "layout fixture", "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]}"#,
        )
        .unwrap();
        project
    }

    /// Audit repair (a): an existing home target refuses the apply before any write —
    /// the home side may hold data written during a previously switched window.
    #[test]
    fn apply_refuses_when_a_home_target_already_exists() {
        let tmp = tempfile::tempdir().unwrap();
        let project = fixture_project(tmp.path());
        let home_root = tmp.path().join("home");
        let existing = home_dir(&home_root, "widgets", "receipts");
        fs::create_dir_all(&existing).unwrap();
        fs::write(
            existing.join("req-newer.json"),
            b"{\"written while switched\":true}",
        )
        .unwrap();
        let confirmed = plan(&project, &home_root).unwrap().plan_digest;
        let err = apply_at(&project, &home_root, &confirmed, "test:layout").unwrap_err();
        assert!(
            err.message
                .contains("home target for receipts already exists"),
            "{err}"
        );
        assert_eq!(
            fs::read(existing.join("req-newer.json")).unwrap(),
            b"{\"written while switched\":true}".to_vec(),
            "the existing home data is untouched"
        );
        assert!(
            !home_root.join("projects/widgets/layout.json").exists(),
            "the pointer is never written by a refused apply"
        );
    }

    #[test]
    fn apply_refuses_a_placeholder_permission_reference() {
        let tmp = tempfile::tempdir().unwrap();
        let project = fixture_project(tmp.path());
        let home_root = tmp.path().join("home");
        let confirmed = plan(&project, &home_root).unwrap().plan_digest;
        for permission in ["", "null", "placeholder"] {
            let err = apply_at(&project, &home_root, &confirmed, permission).unwrap_err();
            assert_eq!(
                err.message, "permission reference required",
                "{permission:?}"
            );
        }
        assert!(
            !home_root.join("projects").exists(),
            "a refused apply writes nothing"
        );
    }

    #[test]
    fn migration_rechecks_destinations_inside_the_locked_section() {
        for switched in [false, true] {
            let tmp = tempfile::tempdir().unwrap();
            let project = fixture_project(tmp.path());
            let home_root = tmp.path().join("home");
            let confirmed = plan(&project, &home_root).unwrap().plan_digest;
            check_destinations(&project, &home_root).unwrap();
            let journal = crate::claim::Journal::open(&tmp.path().join("journal")).unwrap();
            assert!(journal.try_lock().unwrap());
            let target = home_dir(&home_root, "widgets", "receipts");
            fs::create_dir_all(&target).unwrap();
            fs::write(target.join("newer.json"), b"newer home data").unwrap();
            let marker = home_root.join("projects/widgets/layout.json");
            if switched {
                fs::write(&marker, b"committed pointer").unwrap();
            }
            let error = apply_locked(&project, &home_root, &confirmed, &journal).unwrap_err();
            assert!(
                error.contains(if switched {
                    "already switched"
                } else {
                    "already exists"
                }),
                "{error}"
            );
            assert_eq!(
                fs::read(target.join("newer.json")).unwrap(),
                b"newer home data"
            );
            if switched {
                assert_eq!(fs::read(&marker).unwrap(), b"committed pointer");
            } else {
                assert!(!marker.exists());
            }
        }
    }

    #[test]
    fn copying_refuses_a_target_created_after_the_preflight() {
        let tmp = tempfile::tempdir().unwrap();
        let project = fixture_project(tmp.path());
        let source = project.join(crate::project_state!("receipts"));
        fs::create_dir_all(&source).unwrap();
        fs::write(source.join("old.json"), b"old project data").unwrap();
        let home_root = tmp.path().join("home");
        let entry = entry_for(&project, &home_root, "receipts", "widgets").unwrap();
        check_destinations(&project, &home_root).unwrap();
        let target = Path::new(&entry.target);
        fs::create_dir_all(target).unwrap();
        fs::write(target.join("newer.json"), b"only home copy").unwrap();
        assert!(copy_verified(&entry).is_err());
        assert_eq!(
            fs::read(target.join("newer.json")).unwrap(),
            b"only home copy"
        );
        assert!(!target.join("old.json").exists());
        assert_eq!(
            fs::read(source.join("old.json")).unwrap(),
            b"old project data"
        );
        assert!(!home_root.join("projects/widgets/layout.json").exists());
    }
}
