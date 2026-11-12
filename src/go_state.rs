//! #2561: Go-generation project state — the read-only preview (`go-state-plan`).
//!
//! The Go tool records what it deployed into a project in `.moai/manifest.json`
//! (`internal/manifest/types.go`: provenance `template_managed` means "deployed from a
//! template and not modified by the user"; hashes are `sha256:<hex>`). That file is read
//! as untrusted input. A path is offered for moving only when every check holds:
//! a plain relative key (no root, no `..`, no backslash), under `.moai/` or `.claude/`,
//! no path component is a symlink, the path is a regular file, the manifest records it as
//! `template_managed`, its bytes still hash to the recorded `deployed_hash`, and
//! `.claude/settings.json` does not reference it. Anything else stays, with its reason;
//! files the manifest does not record are never offered and are counted per folder.
//! Without a usable manifest nothing moves. Which of these rules apply to `.moai/` as a
//! whole is a pending decision (#2561 comment 6827973256); [`policy`] is the one place
//! that choice lives. This module writes nothing.

use serde::{Deserialize, Serialize};
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
/// The Go manifest of this checkout is about 180 KiB; this only refuses absurd inputs.
const MANIFEST_LIMIT: u64 = 16 * 1024 * 1024;
const FILE_LIMIT: u64 = 64 * 1024 * 1024;
const ROOTS: [&str; 2] = [".moai", ".claude"];

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
    pub warnings: Vec<String>,
    pub plan_digest: String,
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
    settings: &str,
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
    if key == MANIFEST {
        return ("leave", "the Go manifest itself");
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
    if referenced(settings, key) {
        return ("leave", "referenced by .claude/settings.json");
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
    let moai = project.join(".moai");
    if fs::symlink_metadata(&moai).is_ok_and(|m| m.file_type().is_symlink()) {
        return Err(".moai is a symlink; refused".into());
    }
    let mut warnings = vec![
        "Running Claude Code sessions in this project may read .claude/hooks, rules and output \
         styles; apply only while none is running (this preview cannot tell)."
            .to_string(),
    ];
    let manifest = match read_bounded_text(&project, MANIFEST, MANIFEST_LIMIT)? {
        None => None,
        Some(text) => serde_json::from_str::<Manifest>(&text).ok(),
    };
    let settings = read_bounded_text(&project, SETTINGS, MANIFEST_LIMIT)?.unwrap_or_default();
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
    let mut keys: Vec<&String> = recorded.keys().collect();
    keys.sort();
    for key in keys {
        let (action, reason) = policy(&project, key, &recorded[key], &settings);
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
    let mut plan = GoStatePlan {
        schema_version: SCHEMA_VERSION,
        project: project
            .to_str()
            .ok_or("non UTF-8 project path")?
            .to_string(),
        manifest_version: version,
        items,
        unrecorded,
        warnings,
        plan_digest: String::new(),
    };
    let bytes = serde_json::to_vec(&plan).map_err(|_| "encode failed")?;
    plan.plan_digest = format!("sha256:{:x}", Sha256::digest(bytes));
    Ok(plan)
}
