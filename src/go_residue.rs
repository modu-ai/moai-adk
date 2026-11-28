//! The read-only residue survey (SPEC-20261121-go-residue-detect): the files the Go
//! generation left behind, proven against the first manifest it can still find. The survey
//! reads only — the plan writes nothing, so the tree and every digest it reports are the
//! same before and after. Secret-shaped names are counted, never opened; links are never
//! followed; the walk stops at fixed depth and file-count limits.
//!
//! Classes (REQ-3): `go_same` is a `template_managed` entry of the found manifest whose bytes
//! equal its `deployed_hash`, `go_changed` is a `template_managed` entry with different bytes or
//! a `user_modified` one, `go_marker` is one of the Go markers back after a committed move,
//! `user_data` is surveyed but named by no manifest, or named with any other provenance (the
//! manifest records a `user_created` file's own bytes, so a matching hash proves nothing), and
//! `leftover_folder` is an earlier-generation folder kept whole. The receipt scan marks
//! `resurrected` for files a committed go-state receipt says moved and that are back.
//!
//! SPEC-20261126-go-residue-in-use: a `go_same` or `go_changed` file that a project reader still
//! names (`readers`: an `@` import, a settings string, a script those name, the status line's
//! own files) is `go_in_use` with the action `keep` and a `used_by` list of reader paths and
//! kinds, never a setting value. When a reader that exists cannot be read, every `go_same` and
//! `go_changed` file is kept and the reader is listed under `unread_readers`. The survey also
//! lists, report only and with the action `keep`, the local rules and harness files of REQ-5
//! (`report_only_path`).

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    path::Path,
};

use crate::home_state::is_secret;
use crate::worktree_include::secret_name;

mod readers;

/// The walk stops after this many files.
const WALK_LIMIT: usize = 4000;
/// The report-only listing stops after this many files, and descends at most `WALK_DEPTH` levels.
const LISTED_LIMIT: usize = 1000;
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
    /// `go_same`, `go_changed`, `go_in_use`, `go_marker`, `user_data` or `leftover_folder`.
    pub class: &'static str,
    /// `sha256:<hex>` of the bytes on disk; never read for a secret-shaped name.
    pub sha256: Option<String>,
    pub tracked: bool,
    /// A committed go-state receipt says this moved and it is back.
    pub resurrected: bool,
    /// `archive` (the next slice moves it to a backup area) or `keep` (user data).
    pub suggested_action: &'static str,
    /// For `go_in_use`: each reader's project path and how it names the file (`import`,
    /// `settings`, `script` or `statusline`); empty for every other class.
    pub used_by: Vec<(String, &'static str)>,
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
    /// Manifest-named paths the path guard refused (absolute, `..`, through a link): counted,
    /// never read.
    pub refused: usize,
    /// Readers that exist and could not be read; any one keeps every `go_same` and
    /// `go_changed` file in place.
    pub unread: Vec<readers::Unread>,
    /// The report-only listing stopped at its limit.
    pub listing_cut: bool,
}

/// Files another step owns: the host settings (the go-settings and claude-settings steps) and
/// the root instruction files (the go-state and claude-md steps). They are surveyed and shown,
/// but the apply never moves them.
const OWNED_ELSEWHERE: [&str; 6] = [
    ".claude/settings.json",
    ".claude/settings.local.json",
    "CLAUDE.md",
    "CLAUDE.local.md",
    "AGENTS.md",
    "AGENTS.local.md",
];

/// The one path guard (SPEC-20261125-go-residue-apply REQ-3), used by the survey, the apply and
/// the revert before they read, write or move anything at `rel`: a plain relative path (normal
/// components only, so no absolute path and no `..`) none of whose existing components, from the
/// project root down, is a link.
pub(crate) fn contained(project: &Path, rel: &str) -> bool {
    crate::go_state::plain_relative(rel) && !crate::go_state::linked_component(project, rel)
}

/// A secret-shaped name anywhere in the path (`.moai/.env`, `.moai/config/credentials.json`),
/// judged component by component like go-state does. The walk judges bare file names, but a
/// manifest names whole paths, so the survey, the apply and the revert all use this one check.
pub(crate) fn secret_path(rel: &str) -> bool {
    crate::go_state::has_secret_component(rel)
}

/// Reads the manifest named by the chain and returns its path (relative to the project)
/// with its `files` map. The chain: `.moai/manifest.json`, a `manifest.json` under any
/// `.moai-backup*` folder, a `manifest.json` under `.moai/backups/go-markers-*`, then the
/// backup manifest the newest committed go-state receipt names.
pub(crate) fn find_manifest(project: &Path) -> Option<(String, BTreeMap<String, Recorded>)> {
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

/// What one manifest entry records: the bytes of the file and who the Go tool says wrote it.
pub(crate) struct Recorded {
    pub(crate) deployed_hash: String,
    /// The `sha256:<hex>` of the bytes the Go tool would have written for the path, when the
    /// entry holds a non-empty one.
    pub(crate) template_hash: Option<String>,
    provenance: String,
}

/// Reads one manifest and flattens it to `path -> Recorded`; a manifest that is not a
/// real file or does not parse is skipped (the chain moves on).
fn read_manifest(project: &Path, rel: &str) -> Option<BTreeMap<String, Recorded>> {
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
                let entry = Recorded {
                    deployed_hash: e.get("deployed_hash")?.as_str()?.to_string(),
                    template_hash: e
                        .get("template_hash")
                        .and_then(Value::as_str)
                        .filter(|hash| !hash.is_empty())
                        .map(str::to_string),
                    // No provenance reads as none of the template ones.
                    provenance: e
                        .get("provenance")
                        .and_then(Value::as_str)
                        .unwrap_or_default()
                        .to_string(),
                };
                Some((path.clone(), entry))
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
    // Every git subprocess in src runs with the scrubbed environment.
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
    let (walked, mut secrets) = walk(project);
    let tracked = tracked_paths(project);
    let moved = moved_by_receipts(project);
    let mut items: Vec<Item> = Vec::new();
    let mut refused = 0usize;

    // Manifest-named files: the proof decides go_same against go_changed; a named path
    // outside the state folder is surveyed too (never through a link, never a secret). A named
    // path the guard refuses is counted and never joined to the project, so an absolute path,
    // a `..` or a link cannot make the survey read a file outside it.
    for (rel, recorded) in &named {
        if !contained(project, rel) {
            refused += 1;
            continue;
        }
        if secret_path(rel) {
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
        // The Go manifest records the bytes on disk for a file it did not deploy too, so a
        // matching hash proves the Go tool wrote them only under `template_managed`. As in
        // go-state: an edited template file is asked about, anything else is the user's own.
        let same = actual.as_deref() == Some(recorded.deployed_hash.as_str());
        let class = match recorded.provenance.as_str() {
            "template_managed" if same => "go_same",
            "template_managed" | "user_modified" => "go_changed",
            _ => "user_data",
        };
        items.push(Item {
            path: rel.clone(),
            class,
            sha256: actual,
            tracked: tracked.contains(rel),
            resurrected: moved.contains(rel),
            suggested_action: if class == "user_data" || report_only_path(rel) {
                "keep"
            } else if OWNED_ELSEWHERE.contains(&rel.as_str()) {
                "owned_elsewhere"
            } else {
                "archive"
            },
            used_by: Vec::new(),
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
            used_by: Vec::new(),
        });
    }
    // Earlier-generation folders kept whole. The earlier root's name comes from the
    // resolver module's constant, so this file never spells it.
    for folder in [crate::state_root::EARLIER_PROJECT_DIR, ".moai-backups"] {
        let rel = folder.to_string();
        if fs::symlink_metadata(project.join(folder)).is_ok_and(|m| m.is_dir()) {
            items.push(Item {
                path: rel,
                class: "leftover_folder",
                sha256: None,
                tracked: false,
                resurrected: false,
                suggested_action: "archive",
                used_by: Vec::new(),
            });
        }
    }
    // REQ-5: the local rules and harness files the walk above never reaches, listed to be seen.
    let (listed, listed_secrets, listing_cut) = list_report_only(project);
    secrets += listed_secrets;
    let known: BTreeSet<String> = items.iter().map(|i| i.path.clone()).collect();
    for rel in listed.into_iter().filter(|rel| !known.contains(rel)) {
        let full = project.join(&rel);
        let is_link = fs::symlink_metadata(&full).is_ok_and(|m| m.file_type().is_symlink());
        items.push(Item {
            sha256: if is_link { None } else { file_hash(&full) },
            tracked: tracked.contains(&rel),
            resurrected: false,
            path: rel,
            class: "user_data",
            suggested_action: "keep",
            used_by: Vec::new(),
        });
    }
    // REQ-1 to REQ-3: what a reader still names is `go_in_use`; one unread reader keeps the rest.
    let targets: BTreeSet<String> = items
        .iter()
        .filter(|i| matches!(i.class, "go_same" | "go_changed"))
        .map(|i| i.path.clone())
        .collect();
    let found = readers::collect(project, &targets);
    for item in items
        .iter_mut()
        .filter(|i| matches!(i.class, "go_same" | "go_changed"))
    {
        if let Some(readers) = found.used_by.get(&item.path) {
            item.class = "go_in_use";
            item.used_by = readers.iter().cloned().collect();
        }
        if item.class == "go_in_use" || !found.unread.is_empty() {
            item.suggested_action = "keep";
        }
    }
    items.sort_by(|a, b| a.path.cmp(&b.path));
    Survey {
        items,
        proof: Some(proof),
        secret_names: secrets,
        refused,
        unread: found.unread,
        listing_cut,
    }
}

/// REQ-5: a path of the local rules (`.moai/docs/local-rules/`), the harness commands
/// (`.claude/commands/harness/`) or a harness workflow (`.claude/workflows/hns-*`). Shown with
/// the action `keep` and never part of an apply bundle.
fn report_only_path(rel: &str) -> bool {
    rel.starts_with(".moai/docs/local-rules/")
        || rel.starts_with(".claude/commands/harness/")
        || rel
            .strip_prefix(".claude/workflows/")
            .is_some_and(|rest| rest.starts_with("hns-"))
}

/// The files of the report-only places, sorted, with the count of secret-shaped names (never
/// opened) and whether the listing stopped at [`LISTED_LIMIT`]. Links are listed by name and
/// never followed, and a place whose path passes through a link is skipped. The listing starts
/// at the three places only, so it never looks under `.claude/worktrees`.
fn list_report_only(project: &Path) -> (Vec<String>, usize, bool) {
    let mut files = Vec::new();
    let mut secrets = 0usize;
    let mut cut = false;
    let mut places: Vec<String> = vec![
        ".moai/docs/local-rules".into(),
        ".claude/commands/harness".into(),
    ];
    if contained(project, ".claude/workflows") {
        if let Ok(entries) = fs::read_dir(project.join(".claude/workflows")) {
            let mut named: Vec<String> = entries
                .filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| n.starts_with("hns-"))
                .map(|n| format!(".claude/workflows/{n}"))
                .collect();
            named.sort();
            places.extend(named);
        }
    }
    let shaped = |name: &str| is_secret(name) || secret_name(name);
    'places: for place in places {
        if !contained(project, &place) {
            continue;
        }
        if shaped(place.rsplit('/').next().unwrap_or(&place)) {
            secrets += 1;
            continue;
        }
        let mut stack = vec![(place, 0usize)];
        while let Some((rel, depth)) = stack.pop() {
            let Ok(meta) = fs::symlink_metadata(project.join(&rel)) else {
                continue;
            };
            if meta.is_dir() {
                let Ok(entries) = fs::read_dir(project.join(&rel)) else {
                    continue;
                };
                let mut names: Vec<String> = entries
                    .filter_map(|e| e.ok())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .collect();
                // Popped from the end, so the sorted order is walked first to last.
                names.sort_by(|a, b| b.cmp(a));
                for name in names {
                    if shaped(&name) {
                        secrets += 1;
                    } else if depth < WALK_DEPTH {
                        stack.push((format!("{rel}/{name}"), depth + 1));
                    } else {
                        cut = true;
                    }
                }
                continue;
            }
            if files.len() >= LISTED_LIMIT {
                cut = true;
                break 'places;
            }
            files.push(rel);
        }
    }
    files.sort();
    (files, secrets, cut)
}

/// The step's `preview` value: the items with every field the SPEC requires.
pub fn preview(survey: &Survey) -> Value {
    json!({
        "items": survey.items.iter().map(|i| {
            let mut item = json!({
                "path": i.path,
                "class": i.class,
                "sha256": i.sha256,
                "tracked": i.tracked,
                "resurrected": i.resurrected,
                "suggested_action": i.suggested_action,
            });
            // A reader's path and the way it names the file, never a value it holds (REQ-1).
            if i.class == "go_in_use" {
                item["used_by"] = i.used_by.iter()
                    .map(|(path, kind)| json!({"path": path, "kind": kind}))
                    .collect();
            }
            item
        }).collect::<Vec<_>>(),
        "proof": survey.proof.clone().unwrap_or_else(|| "none".into()),
        "secret_names": survey.secret_names,
        "refused_paths": survey.refused,
        "unread_readers": survey.unread.iter()
            .map(|u| json!({"file": u.file, "reason": u.reason}))
            .collect::<Vec<_>>(),
        "listing_cut": survey.listing_cut,
    })
}

/// The readers that stand in the way, as the apply's refusal names them: the file's readers for
/// a `go_in_use` item, the unread readers for a file kept because of them.
pub(crate) fn kept_because(survey: &Survey, path: &str) -> Option<String> {
    let item = survey.items.iter().find(|i| i.path == path)?;
    if item.class == "go_in_use" {
        let readers: Vec<String> = item
            .used_by
            .iter()
            .map(|(reader, kind)| format!("{reader} ({kind})"))
            .collect();
        return Some(format!("it is still read by {}", readers.join(", ")));
    }
    if matches!(item.class, "go_same" | "go_changed") && item.suggested_action == "keep" {
        let files: Vec<&str> = survey.unread.iter().map(|u| u.file.as_str()).collect();
        let why = if files.is_empty() {
            "it is a local rule or harness file the survey only lists".to_string()
        } else {
            format!("a reader could not be read: {}", files.join(", "))
        };
        return Some(why);
    }
    None
}
