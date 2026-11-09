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
//! the P09 결정안 lives in `docs/home-layout-p09.md`.
use crate::profile::read_profile;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

pub const SCHEMA_VERSION: u32 = 1;
/// The three store kinds this layout moves: (kind, project-side directory relative to
/// the project root, home-side directory relative to
/// `~/.moai-cli/projects/<project_id>/`). The claim journal's project-side root is the
/// git common dir (`<common>/.moai-cli/claims`), resolved per clone — the table here
/// carries the project-root-relative spelling used when the common dir equals the
/// project's own `.git` (the primary checkout).
pub const KINDS: &[(&str, &str, &str)] = &[
    ("receipts", ".moai-cli/receipts", "receipts"),
    ("outbox", ".moai-cli/forge/outbox", "forge/outbox"),
    ("claims", ".moai-cli/claims", "claims"),
];

/// The default home root: `$HOME/.moai-cli` (the card's fixed location). `None` without
/// a HOME — callers then stay on the project side.
pub fn default_home_root() -> Option<PathBuf> {
    std::env::var_os("HOME")
        .map(PathBuf::from)
        .filter(|h| !h.as_os_str().is_empty())
        .map(|h| h.join(crate::HOME_STATE_DIR))
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
fn claims_legacy_dir(project: &Path) -> Option<PathBuf> {
    let id = crate::identity::identity(project).ok()?;
    let common = id["clone_anchor"].as_str()?;
    let dir = Path::new(common).join(".moai-cli").join("claims");
    dir.is_dir().then_some(dir)
}

fn legacy_dir(project: &Path, kind: &str) -> Option<PathBuf> {
    let rel = KINDS
        .iter()
        .find(|(k, _, _)| *k == kind)
        .map(|(_, legacy, _)| *legacy)?;
    if kind == "claims" {
        return claims_legacy_dir(project);
    }
    let dir = project.join(rel);
    dir.is_dir().then_some(dir)
}

fn home_dir(home_root: &Path, project_id: &str, kind: &str) -> PathBuf {
    let sub = KINDS
        .iter()
        .find(|(k, _, _)| *k == kind)
        .map(|(_, _, sub)| *sub)
        .unwrap_or(kind);
    home_root.join("projects").join(project_id).join(sub)
}

fn measure_fb(root: &Path) -> (u64, u64) {
    let (mut files, mut bytes) = (0u64, 0u64);
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        let Ok(entries) = std::fs::read_dir(&dir) else {
            continue;
        };
        for e in entries.flatten() {
            let p = e.path();
            if p.is_dir() {
                stack.push(p);
            } else if let Ok(m) = e.metadata() {
                files += 1;
                bytes += m.len();
            }
        }
    }
    (files, bytes)
}

fn entry_for(project: &Path, home_root: &Path, kind: &str, project_id: &str) -> LayoutEntry {
    let target = home_dir(home_root, project_id, kind);
    match legacy_dir(project, kind) {
        Some(source) => {
            let (files, bytes) = measure_fb(&source);
            LayoutEntry {
                kind: kind.to_string(),
                source: source.to_string_lossy().into_owned(),
                target: target.to_string_lossy().into_owned(),
                files,
                bytes,
                source_digest: crate::tree_digest(&source, &[]).unwrap_or_default(),
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
    }
}

/// Measures the three stores and binds the plan. Writes nothing.
pub fn plan(project: &Path, home_root: &Path) -> Result<LayoutPlan> {
    let project_id = identity_of(project)?;
    let mut entries: Vec<LayoutEntry> = KINDS
        .iter()
        .map(|(k, _, _)| entry_for(project, home_root, k, &project_id))
        .collect();
    entries.sort_by(|a, b| a.kind.cmp(&b.kind));
    let canonical = serde_json::to_vec(&entries).map_err(|_| "encode failed".to_string())?;
    Ok(LayoutPlan {
        schema_version: SCHEMA_VERSION,
        project_id: project_id.clone(),
        home_root: home_root.to_string_lossy().into_owned(),
        plan_digest: crate::digest(&canonical),
        entries,
    })
}

fn identity_of(project: &Path) -> Result<String> {
    let (profile, _) = read_profile(project)?;
    Ok(profile.project_id)
}

/// Copies one source tree to its target without touching the source, then verifies the
/// copy reproduces the source digest (REQ-B8-01: 복사·검증).
fn copy_verified(entry: &LayoutEntry) -> Result<()> {
    let source = Path::new(&entry.source);
    let target = Path::new(&entry.target);
    std::fs::create_dir_all(target).map_err(|_| "home layout target unavailable".to_string())?;
    let mut stack = vec![source.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in std::fs::read_dir(&dir).map_err(|_| "layout source unreadable".to_string())? {
            let from = e
                .map_err(|_| "layout source unreadable".to_string())?
                .path();
            let rel = from
                .strip_prefix(source)
                .map_err(|_| "layout walk failed".to_string())?;
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
    let copied =
        crate::tree_digest(target, &[]).map_err(|_| "home layout verify failed".to_string())?;
    if copied != entry.source_digest {
        // The verified copy failed: remove the partial copy, keep the source intact.
        let _ = std::fs::remove_dir_all(target);
        return Err("home layout verify failed; partial copy removed".into());
    }
    Ok(())
}

/// Executes a confirmed plan: copy every entry, verify the copy reproduces the source
/// digest, then switch the pointer (write `layout.json`). The originals are never
/// deleted (REQ-B8-01). A plan whose digest no longer matches the tree is refused
/// (stale); a second apply while the pointer is already switched is a conflict.
pub fn apply_at(project: &Path, home_root: &Path, confirmed: &str) -> Result<Value> {
    if active_home(project, home_root).is_some() {
        return Err(
            "conflict in home layout: already switched; remove the home layout.json marker to revert first"
                .into(),
        );
    }
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

pub fn apply_cli(project: &Path, confirmed: &str) -> Result<Value> {
    let home_root = default_home_root().ok_or("home root unavailable; HOME is not set")?;
    apply_at(project, &home_root, confirmed)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn plan_refuses_without_a_profile() {
        let tmp = tempfile::tempdir().unwrap();
        let err = plan(&tmp.path().join("p"), tmp.path()).unwrap_err();
        assert!(
            err.contains("profile") || err.contains("unavailable"),
            "{err}"
        );
    }
}
