//! The read-only residue survey (SPEC-20261121-go-residue-detect): the files the Go
//! generation left behind, proven against the first manifest it can still find. The survey
//! reads only — the plan writes nothing, so the tree and every digest it reports are the
//! same before and after. Secret-shaped names are counted, never opened; links are never
//! followed; the walk stops at fixed depth and file-count limits.
//!
//! Classes (REQ-3): `go_same` is byte-identical to a `deployed_hash` of the found manifest,
//! `go_changed` is named by the manifest with different bytes, `go_marker` is one of the Go
//! markers back after a committed move, `user_data` is surveyed but named by no manifest,
//! and `leftover_folder` is an earlier-generation folder kept whole. The receipt scan marks
//! `resurrected` for files a committed go-state receipt says moved and that are back.

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    path::Path,
};

use crate::home_state::is_secret;
use crate::worktree_include::secret_name;

/// The walk stops after this many files.
const WALK_LIMIT: usize = 4000;
/// The walk descends at most this many levels under the state folder.
const WALK_DEPTH: usize = 8;
/// Receipt stamps read for `resurrected` and the receipt-named backup manifest.
const RECEIPT_SCAN: usize = 64;
/// Go markers the state-root classifier reads; a copy back on disk is `go_marker`.
const MARKERS: [&str; 3] = [
    "manifest.json",
    "config/config.yaml",
    "config/sections/system.yaml",
];
/// Folders the survey never enters or reports: this generation's state and the user-data
/// folders go-state restores (their files are this generation's or the user's own).
const SKIPPED: [&str; 15] = [
    "moai.json",
    "project.yaml",
    "manifests",
    "receipts",
    "index",
    "backups",
    "specs",
    "plans",
    "reports",
    "docs",
    "memory",
    "decisions",
    "research",
    "todo",
    "release-notes",
];

/// One surveyed file or folder, with the action the next slice would take.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Item {
    /// Slash-separated path relative to the project.
    pub path: String,
    /// `go_same`, `go_changed`, `go_marker`, `user_data` or `leftover_folder`.
    pub class: &'static str,
    /// `sha256:<hex>` of the bytes on disk; never read for a secret-shaped name.
    pub sha256: Option<String>,
    pub tracked: bool,
    /// A committed go-state receipt says this moved and it is back.
    pub resurrected: bool,
    /// `archive` (the next slice moves it to a backup area) or `keep` (user data).
    pub suggested_action: &'static str,
}

/// One survey: the items, the manifest that proved them, the secret-shaped name count.
#[derive(Debug, Default, PartialEq, Eq)]
pub struct Survey {
    pub items: Vec<Item>,
    /// The manifest that proved the classified files, relative to the project; `None` when
    /// no manifest proves anything and nothing is a candidate.
    pub proof: Option<String>,
    /// Secret-shaped names among the candidates when a proof exists; `0` without one.
    pub secret_names: usize,
}

/// Reads the manifest named by the chain and returns its path (relative to the project)
/// with its `files` map. The chain: `.moai/manifest.json`, a `manifest.json` under any
/// `.moai-backup*` folder, a `manifest.json` under `.moai/backups/go-markers-*`, then the
/// backup manifest the newest committed go-state receipt names.
fn find_manifest(project: &Path) -> Option<(String, BTreeMap<String, String>)> {
    let live = ".moai/manifest.json";
    if let Some(map) = read_manifest(project, live) {
        return Some((live.into(), map));
    }
    if let Ok(entries) = fs::read_dir(project) {
        let mut folders: Vec<String> = entries
            .filter_map(|e| e.ok())
            .filter_map(|e| e.file_name().into_string().ok())
            .filter(|n| n.starts_with(".moai-backup"))
            .collect();
        folders.sort();
        for folder in folders {
            let rel = format!("{folder}/manifest.json");
            if let Some(map) = read_manifest(project, &rel) {
                return Some((rel, map));
            }
        }
    }
    if let Ok(entries) = fs::read_dir(project.join(".moai/backups")) {
        let mut stamps: Vec<String> = entries
            .filter_map(|e| e.ok())
            .filter_map(|e| e.file_name().into_string().ok())
            .filter(|n| n.starts_with("go-markers-"))
            .collect();
        stamps.sort();
        if let Some(newest) = stamps.pop() {
            let rel = format!(".moai/backups/{newest}/manifest.json");
            if let Some(map) = read_manifest(project, &rel) {
                return Some((rel, map));
            }
        }
    }
    for stamp in receipt_stamps(project) {
        if let Ok(receipt) =
            fs::read_to_string(project.join(format!(".moai/backups/{stamp}/go-state-receipt.json")))
        {
            if let Ok(r) = serde_json::from_str::<Value>(&receipt) {
                if let Some(backup) = r.get("go_backup").and_then(|b| b.as_str()) {
                    let rel = format!("{backup}/manifest.json");
                    if let Some(map) = read_manifest(project, &rel) {
                        return Some((rel, map));
                    }
                }
            }
        }
    }
    None
}

/// Reads one manifest and flattens it to `path -> deployed_hash`; a manifest that is not a
/// real file or does not parse is skipped (the chain moves on).
fn read_manifest(project: &Path, rel: &str) -> Option<BTreeMap<String, String>> {
    let meta = fs::symlink_metadata(project.join(rel)).ok()?;
    if !meta.is_file() {
        return None;
    }
    let text = fs::read_to_string(project.join(rel)).ok()?;
    let v: Value = serde_json::from_str(&text).ok()?;
    let files = v.get("files")?.as_object()?;
    Some(
        files
            .iter()
            .filter_map(|(path, e)| {
                let hash = e.get("deployed_hash")?.as_str()?.to_string();
                Some((path.clone(), hash))
            })
            .collect(),
    )
}

/// Receipt stamp folders under the state backup area, newest first.
fn receipt_stamps(project: &Path) -> Vec<String> {
    let mut stamps: Vec<String> = match fs::read_dir(project.join(".moai/backups")) {
        Ok(entries) => entries
            .filter_map(|e| e.ok())
            .filter_map(|e| e.file_name().into_string().ok())
            .filter(|n| n.len() == 17 && n.starts_with('2'))
            .collect(),
        Err(_) => Vec::new(),
    };
    stamps.sort();
    stamps.reverse();
    stamps.truncate(RECEIPT_SCAN);
    stamps
}

/// The move records of every readable receipt: path -> it moved.
fn moved_by_receipts(project: &Path) -> BTreeSet<String> {
    let mut moved = BTreeSet::new();
    for stamp in receipt_stamps(project) {
        let Ok(text) = fs::read_to_string(
            project.join(format!(".moai/backups/{stamp}/go-state-receipt.json")),
        ) else {
            continue;
        };
        let Ok(r) = serde_json::from_str::<Value>(&text) else {
            continue;
        };
        if let Some(moves) = r.get("moves").and_then(|m| m.as_array()) {
            for m in moves {
                if m.get("moved").and_then(|b| b.as_bool()).unwrap_or(false) {
                    if let Some(p) = m.get("path").and_then(|p| p.as_str()) {
                        moved.insert(p.to_string());
                    }
                }
            }
        }
    }
    moved
}

/// `sha256:<hex>` of the file's bytes, never called for a secret-shaped name.
fn file_hash(path: &Path) -> Option<String> {
    let bytes = fs::read(path).ok()?;
    Some(format!("sha256:{:x}", Sha256::digest(&bytes)))
}

/// Whether git tracks the path in this project (one listing, never per file).
fn tracked_paths(project: &Path) -> BTreeSet<String> {
    let mut cmd = std::process::Command::new("git");
    cmd.arg("-C").arg(project).args(["ls-files", "-z"]);
    // Every git subprocess in src runs with the scrubbed environment (#2598).
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd.output();
    match out {
        Ok(o) if o.status.success() => String::from_utf8_lossy(&o.stdout)
            .split('\0')
            .filter(|p| !p.is_empty())
            .map(str::to_string)
            .collect(),
        _ => BTreeSet::new(),
    }
}

/// The surveyed files under `.moai/`, the walk's own limits and skip list applied; links are
/// listed by name and never followed, secret-shaped names are counted and never opened.
fn walk(project: &Path) -> (Vec<String>, usize) {
    let mut files = Vec::new();
    let mut secrets = 0usize;
    let mut stack = vec![(String::from(".moai"), 0usize)];
    while let Some((rel, depth)) = stack.pop() {
        if files.len() >= WALK_LIMIT {
            break;
        }
        let full = project.join(&rel);
        let Ok(entries) = fs::read_dir(&full) else {
            continue;
        };
        for e in entries.filter_map(|e| e.ok()) {
            if files.len() >= WALK_LIMIT {
                break;
            }
            let Ok(name) = e.file_name().into_string() else {
                continue;
            };
            let child = format!("{rel}/{name}");
            let Ok(meta) = fs::symlink_metadata(e.path()) else {
                continue;
            };
            if meta.file_type().is_symlink() {
                // A link is never followed; its name is surveyed as a plain name.
                files.push(child);
                continue;
            }
            if meta.is_dir() {
                if depth < WALK_DEPTH && !SKIPPED.contains(&name.as_str()) {
                    stack.push((child, depth + 1));
                }
                continue;
            }
            if is_secret(&name) || secret_name(&name) {
                secrets += 1;
                continue;
            }
            files.push(child);
        }
    }
    files.sort();
    (files, secrets)
}

/// One survey of the project. With no proving manifest nothing is a candidate: the answer
/// is empty with no proof and zero secret names (the plan still succeeds).
pub fn survey(project: &Path) -> Survey {
    let Some((proof, named)) = find_manifest(project) else {
        return Survey::default();
    };
    let (walked, secrets) = walk(project);
    let tracked = tracked_paths(project);
    let moved = moved_by_receipts(project);
    let mut items: Vec<Item> = Vec::new();

    // Manifest-named files: the proof decides go_same against go_changed; a named path
    // outside the state folder is surveyed too (never through a link, never a secret).
    for (rel, deployed) in &named {
        if is_secret(rel) || secret_name(rel) {
            continue;
        }
        let full = project.join(rel);
        let Ok(meta) = fs::symlink_metadata(&full) else {
            continue;
        };
        if !meta.is_file() {
            continue;
        }
        let actual = file_hash(&full);
        let class = if actual.as_deref() == Some(deployed.as_str()) {
            "go_same"
        } else {
            "go_changed"
        };
        items.push(Item {
            path: rel.clone(),
            class,
            sha256: actual,
            tracked: tracked.contains(rel),
            resurrected: moved.contains(rel),
            suggested_action: "archive",
        });
    }
    // Walked files the manifest does not name: a Go marker back on disk, else the user's
    // own data. This generation's own state files are never surveyed, and a link is named
    // but never read (its digest stays empty so no outside content can leak into the item).
    for rel in &walked {
        if named.contains_key(rel) {
            continue;
        }
        let name = rel.rsplit('/').next().unwrap_or(rel);
        if SKIPPED.contains(&name) {
            continue;
        }
        let marker = rel
            .strip_prefix(".moai/")
            .is_some_and(|rest| MARKERS.contains(&rest));
        let is_link =
            fs::symlink_metadata(project.join(rel)).is_ok_and(|m| m.file_type().is_symlink());
        items.push(Item {
            path: rel.clone(),
            class: if marker { "go_marker" } else { "user_data" },
            sha256: if marker || is_link {
                None
            } else {
                file_hash(&project.join(rel))
            },
            tracked: tracked.contains(rel) && !marker,
            resurrected: marker || moved.contains(rel),
            suggested_action: if marker { "archive" } else { "keep" },
        });
    }
    // Earlier-generation folders kept whole. The earlier root's name is spelled only in
    // lib.rs; every other file takes the constant (#2725 S1).
    for folder in [crate::PROJECT_STATE_DIR, ".moai-backups"] {
        let rel = folder.to_string();
        if fs::symlink_metadata(project.join(folder)).is_ok_and(|m| m.is_dir()) {
            items.push(Item {
                path: rel,
                class: "leftover_folder",
                sha256: None,
                tracked: false,
                resurrected: false,
                suggested_action: "archive",
            });
        }
    }
    items.sort_by(|a, b| a.path.cmp(&b.path));
    Survey {
        items,
        proof: Some(proof),
        secret_names: secrets,
    }
}

/// The step's `preview` value: the items with every field the SPEC requires.
pub fn preview(survey: &Survey) -> Value {
    json!({
        "items": survey.items.iter().map(|i| json!({
            "path": i.path,
            "class": i.class,
            "sha256": i.sha256,
            "tracked": i.tracked,
            "resurrected": i.resurrected,
            "suggested_action": i.suggested_action,
        })).collect::<Vec<_>>(),
        "proof": survey.proof.clone().unwrap_or_else(|| "none".into()),
        "secret_names": survey.secret_names,
    })
}
