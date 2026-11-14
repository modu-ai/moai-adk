//! #2725 S2a: which generation owns a project's `.moai/` — read only.
//!
//! The project state moves from `.moai-cli/` to `.moai/` (#2725, leader decisions 6854854683),
//! the folder name the Go generation already uses. Nothing may overwrite or move a `.moai/`
//! before this classification says whose it is (design 6854765655 §2):
//!
//! | `.moai/` | verdict |
//! | --- | --- |
//! | absent | `none` |
//! | a link or another non-folder | `refuse` (never followed) |
//! | valid Rust marker only | `rust` |
//! | Go markers only | `go` |
//! | both | `mixed` (nothing moves; the user decides) |
//! | a folder with neither, empty | `empty` |
//! | a folder with neither, not empty | `unknown` (nothing moves; the user decides) |
//! | unreadable or newer Rust marker | `refuse` |
//! | not `rust`, with a Go backup beside it and `.moai` absent or holding a go-state receipt | `interrupted` (a stopped migration; recovery first) |
//! | HOME only ([`classify_home`]): schema-2 Rust marker, Go markers or not | `rust-first` (#2871 I2, D-M1) |
//!
//! Go markers: `.moai/manifest.json` (the Go deploy manifest), `.moai/config/config.yaml`,
//! and `.moai/config/sections/system.yaml` with a `moai:` block holding `version:`.
//! Rust marker: `.moai/moai-cli.json` = `{"schema_version": 1, "generation": "moai-cli", …}`,
//! written last by a migration (decision D1). Every path is read with `symlink_metadata`;
//! a link anywhere on the way is not a marker. Writes nothing.
//!
//! A Go backup is the Go folder after `go-state-apply` renamed it (#2725 rev2 D2, #2696):
//! exactly `.moai-backup`, or `.moai-backup-<UTC>` when that name was taken. The Go tool's own
//! `.moai-backups/` (plural) is not one.

use serde_json::{json, Value};
use std::{fs, path::Path};

/// The Go and Rust folder name once the state has moved (#2725).
pub const MOAI_DIR: &str = ".moai";
/// The Rust generation marker inside `.moai/` (decision D1).
pub const RUST_MARKER: &str = "moai-cli.json";
/// The marker schema this build writes and understands.
pub const MARKER_SCHEMA: u64 = 1;
/// #2871 I2 R1: a HOME marker with this schema makes the HOME this generation's first
/// (`rust-first`, user decision D-M1 6876633358), Go markers or not. Only
/// [`classify_home`] reads it; projects and older builds refuse it as `newer`. Nothing writes
/// it yet.
pub const HOME_MARKER_SCHEMA: u64 = 2;
/// The name `go-state-apply` gives the Go folder (#2725 rev2 D2); never overwritten.
pub const GO_BACKUP: &str = ".moai-backup";
/// The receipt a `go-state-apply` that moved the Go folder keeps under `.moai/backups/<UTC>/`.
pub const GO_STATE_RECEIPT: &str = "go-state-receipt.json";
/// Markers are small JSON objects; anything larger is not one.
const MARKER_LIMIT: u64 = 64 * 1024;
/// `system.yaml` is a few KiB; a larger file is not read for the version key.
const SYSTEM_YAML_LIMIT: u64 = 1024 * 1024;

/// What a classification found. `generation` is one of the eight verdicts above.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Classified {
    pub generation: &'static str,
    /// `absent`, `folder`, `link` or `other` — what `.moai` itself is.
    pub moai_dir: &'static str,
    /// Go markers found, as paths relative to the project.
    pub go_markers: Vec<String>,
    /// `valid`, `absent`, `newer` or `unreadable`.
    pub rust_marker: &'static str,
    /// The Go backups beside `.moai` that make the verdict `interrupted`, sorted (empty
    /// for every other verdict).
    pub interrupted: Vec<String>,
    /// Entries directly inside `.moai/` (0 unless it is a real folder).
    pub entries: u64,
    /// The earlier Rust project root (`.moai-cli/project.yaml`) is present.
    pub legacy_root: bool,
    pub reason: String,
    /// `rust-first` only (#2871 I2 R1): the Go markers this generation now outranks; their items
    /// retire in a later step, nothing moves here.
    pub go_retire_pending: Vec<String>,
}

impl Classified {
    pub fn to_json(&self) -> Value {
        let mut v = json!({
            "generation": self.generation,
            "moai_dir": self.moai_dir,
            "go_markers": self.go_markers,
            "rust_marker": self.rust_marker,
            "interrupted": self.interrupted,
            "entries": self.entries,
            "legacy_root": self.legacy_root.then_some(crate::PROJECT_STATE_DIR),
            "reason": self.reason,
        });
        // Only a `rust-first` verdict carries the key, so every other verdict's JSON is unchanged.
        if self.generation == "rust-first" {
            v["go_retire_pending"] = json!(self.go_retire_pending);
        }
        v
    }
}

/// A regular file at `root/parts…`, reached without following any link.
pub(crate) fn regular_file(root: &Path, parts: &[&str]) -> Option<fs::Metadata> {
    let mut path = root.to_path_buf();
    for (i, part) in parts.iter().enumerate() {
        path.push(part);
        // `symlink_metadata` describes a link as a link, so neither `is_file` nor `is_dir`
        // below accepts one: a link at any step ends the walk (mutant K4 showed a separate
        // `is_symlink` check changed nothing).
        let meta = fs::symlink_metadata(&path).ok()?;
        let last = i + 1 == parts.len();
        if last {
            return meta.is_file().then_some(meta);
        }
        if !meta.is_dir() {
            return None;
        }
    }
    None
}

pub(crate) fn read_limited(root: &Path, parts: &[&str], limit: u64) -> Option<Vec<u8>> {
    let meta = regular_file(root, parts)?;
    if meta.len() > limit {
        return None;
    }
    let mut path = root.to_path_buf();
    path.extend(parts);
    fs::read(path).ok()
}

/// `system.yaml` names the Go version under a top-level `moai:` block.
fn system_yaml_has_version(text: &str) -> bool {
    let mut in_moai = false;
    for line in text.lines() {
        if line.trim().is_empty() || line.trim_start().starts_with('#') {
            continue;
        }
        if !line.starts_with(' ') && !line.starts_with('\t') {
            in_moai = line.trim_end() == "moai:";
            continue;
        }
        if in_moai && line.trim_start().starts_with("version:") {
            return true;
        }
    }
    false
}

fn go_markers(moai: &Path) -> Vec<String> {
    let mut found = Vec::new();
    if regular_file(moai, &["manifest.json"]).is_some() {
        found.push(format!("{MOAI_DIR}/manifest.json"));
    }
    if regular_file(moai, &["config", "config.yaml"]).is_some() {
        found.push(format!("{MOAI_DIR}/config/config.yaml"));
    }
    let system = ["config", "sections", "system.yaml"];
    if read_limited(moai, &system, SYSTEM_YAML_LIMIT)
        .and_then(|b| String::from_utf8(b).ok())
        .is_some_and(|t| system_yaml_has_version(&t))
    {
        found.push(format!("{MOAI_DIR}/config/sections/system.yaml"));
    }
    found
}

/// `valid`, `absent`, `newer` or `unreadable` for `.moai/moai-cli.json`; `schema-2` too when
/// `accept` is [`HOME_MARKER_SCHEMA`] (HOME only). Old builds read schema 2 as `newer` and
/// refuse it.
fn rust_marker(moai: &Path, accept: u64) -> &'static str {
    let path = moai.join(RUST_MARKER);
    match fs::symlink_metadata(&path) {
        Err(_) => return "absent",
        Ok(m) if !m.is_file() => return "unreadable",
        Ok(_) => {}
    }
    let Some(bytes) = read_limited(moai, &[RUST_MARKER], MARKER_LIMIT) else {
        return "unreadable";
    };
    let Ok(Value::Object(map)) = serde_json::from_slice::<Value>(&bytes) else {
        return "unreadable";
    };
    if map.get("generation").and_then(Value::as_str) != Some("moai-cli") {
        return "unreadable";
    }
    match map.get("schema_version").and_then(Value::as_u64) {
        Some(MARKER_SCHEMA) => "valid",
        Some(HOME_MARKER_SCHEMA) if accept >= HOME_MARKER_SCHEMA => "schema-2",
        Some(v) if v > MARKER_SCHEMA => "newer",
        _ => "unreadable",
    }
}

/// A UTC stamp `YYYYMMDDTHHMMSSZ`, the only suffix a backup or apply name carries.
pub fn utc_stamp(s: &str) -> bool {
    let b = s.as_bytes();
    b.len() == 16
        && b[8] == b'T'
        && b[15] == b'Z'
        && b[..8].iter().chain(&b[9..15]).all(u8::is_ascii_digit)
}

/// `.moai-backup` or `.moai-backup-<UTC>`, and nothing else (not the Go `.moai-backups`).
pub fn go_backup_name(name: &str) -> bool {
    name == GO_BACKUP
        || name
            .strip_prefix(GO_BACKUP)
            .and_then(|r| r.strip_prefix('-'))
            .is_some_and(utc_stamp)
}

/// The Go backups directly under `root`, whatever each entry is, sorted.
fn go_backups(root: &Path) -> Vec<String> {
    let mut out: Vec<String> = fs::read_dir(root)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| go_backup_name(n))
                .collect()
        })
        .unwrap_or_default();
    out.sort();
    out
}

/// `.moai/backups/<UTC>/go-state-receipt.json` exists as a regular file: this `.moai/` is the
/// one a go-state apply made (or a revert has not finished leaving).
fn holds_go_state_receipt(moai: &Path) -> bool {
    fs::read_dir(moai.join("backups"))
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| utc_stamp(n))
                .any(|n| regular_file(moai, &["backups", &n, GO_STATE_RECEIPT]).is_some())
        })
        .unwrap_or(false)
}

/// Classify `root/.moai` (design 6854765655 §2). Reads only.
pub fn classify(root: &Path) -> Classified {
    classify_with(root, go_markers)
}

/// The same verdict table over `root/.moai` with another set of Go markers: the HOME state
/// folder `~/.moai` has different Go files than a project's (#2723 design 6860356009 §2,
/// #2725 rev2 — one table and one set of rules for both). `go` returns the markers found,
/// as paths relative to `root`.
pub fn classify_with(root: &Path, go: fn(&Path) -> Vec<String>) -> Classified {
    classify_inner(root, go, MARKER_SCHEMA)
}

/// The HOME verdict (#2871 I2 R1): the table above with the HOME's Go markers, plus one row —
/// a schema-2 marker is `rust-first`, Go markers or not; the Go markers found are reported as
/// `go_retire_pending` and nothing moves. A HOME without a schema-2 marker gets exactly the
/// verdict of [`classify_with`] (`mixed` stays for schema 1). The way to the layout root
/// ([`crate::home_layout::RUST_HOME_LAYOUT`]) must be real folders, or the verdict is `refuse`.
pub fn classify_home(home: &Path) -> Classified {
    classify_inner(home, crate::home_state::home_go_markers, HOME_MARKER_SCHEMA)
}

fn classify_inner(root: &Path, go: fn(&Path) -> Vec<String>, accept: u64) -> Classified {
    let go_markers = go;
    let moai = root.join(MOAI_DIR);
    let moai_dir = match fs::symlink_metadata(&moai) {
        Err(_) => "absent",
        Ok(m) if m.file_type().is_symlink() => "link",
        Ok(m) if m.is_dir() => "folder",
        Ok(_) => "other",
    };
    let legacy_root = regular_file(root, &[crate::PROJECT_STATE_DIR, "project.yaml"]).is_some();
    let (go, marker, entries) = if moai_dir == "folder" {
        let entries = fs::read_dir(&moai).map(|rd| rd.count() as u64).unwrap_or(0);
        (go_markers(&moai), rust_marker(&moai, accept), entries)
    } else {
        (Vec::new(), "absent", 0)
    };
    // #2696: a go-state apply renames the Go folder to a backup, makes a new `.moai/` holding
    // its receipt and writes the marker last; a revert removes the marker first. A backup
    // beside a `.moai` that is missing, or that is the apply's own folder without a valid
    // marker, is a migration stopped midway in either direction. A backup beside the Go
    // folder or beside any other `.moai/` is not (a finished revert, or the user's folder).
    let backups = go_backups(root);
    // #2696 C4a: the earlier-root move has no Go backup, so its own `.moai/` holding a receipt
    // without a valid marker is a stopped migration by itself.
    let stopped = marker != "valid"
        && marker != "schema-2"
        && ((!backups.is_empty() && moai_dir == "absent")
            || (moai_dir == "folder" && holds_go_state_receipt(&moai)));
    let interrupted = if stopped {
        if backups.is_empty() {
            vec![MOAI_DIR.to_string()]
        } else {
            backups
        }
    } else {
        Vec::new()
    };
    let (generation, reason) = if stopped {
        (
            "interrupted",
            format!(
                "a go-state migration stopped midway ({} and a .moai that is not this \
                 generation's); run go-state-revert-plan before anything else",
                interrupted.join(", ")
            ),
        )
    } else {
        match moai_dir {
            "absent" => ("none", "no .moai/ here".to_string()),
            "link" => ("refuse", ".moai is a link; it is never followed".to_string()),
            "other" => ("refuse", ".moai is neither a folder nor a link".to_string()),
            _ => match (marker, go.is_empty()) {
                ("newer", _) => (
                    "refuse",
                    format!("{MOAI_DIR}/{RUST_MARKER} is from a newer moai-cli; this build leaves it alone"),
                ),
                ("unreadable", _) => (
                    "refuse",
                    format!("{MOAI_DIR}/{RUST_MARKER} is not a marker this build can read"),
                ),
                ("schema-2", _) => match crate::home_layout::check_root_in(
                    crate::home_layout::RUST_HOME_LAYOUT,
                    root,
                ) {
                    Ok(()) => (
                        "rust-first",
                        "this generation's HOME; Go items stay until they retire".to_string(),
                    ),
                    Err(e) => ("refuse", e),
                },
                ("valid", true) => ("rust", "this generation's state".to_string()),
                ("valid", false) => (
                    "mixed",
                    format!(
                        "both the Rust marker and Go markers ({}) are present; nothing moves until the user decides",
                        go.join(", ")
                    ),
                ),
                (_, false) => (
                    "go",
                    format!("Go-generation state ({}); offer the migration preview", go.join(", ")),
                ),
                (_, true) if entries == 0 => ("empty", "an empty .moai/ folder".to_string()),
                (_, true) => (
                    "unknown",
                    format!(
                        ".moai/ holds {entries} entries but no generation marker; nothing moves until the user decides"
                    ),
                ),
            },
        }
    };
    Classified {
        generation,
        moai_dir,
        go_retire_pending: if generation == "rust-first" {
            go.clone()
        } else {
            Vec::new()
        },
        go_markers: go,
        rust_marker: marker,
        interrupted,
        entries,
        legacy_root,
        reason,
    }
}

/// The project-relative folder that holds `root`'s state now (#2725 S2b): `.moai` once
/// [`classify`] says `rust`, otherwise the earlier `.moai-cli` (decision D7: both are read,
/// one writes). A project gets `.moai/` only through a migration that writes the marker last,
/// so every project without a valid marker keeps exactly the paths it had before.
pub fn dir(root: &Path) -> &'static str {
    if classify(root).generation == "rust" {
        MOAI_DIR
    } else {
        crate::PROJECT_STATE_DIR
    }
}

/// #2696 C2b: a project with no MoAI state at all — verdict `none` (no `.moai` of any kind,
/// no stopped migration) and no earlier `.moai-cli` — starts on this generation's folder.
pub fn is_new_project(root: &Path) -> bool {
    fs::symlink_metadata(root.join(crate::PROJECT_STATE_DIR)).is_err()
        && classify(root).generation == "none"
}

/// The folder a writer that creates state uses: `.moai` for a new project (which
/// [`adopt_new`] then marks), otherwise [`dir`]. A preview and its apply both ask this, so the
/// path the user approved is the path written.
pub fn create_dir(root: &Path) -> &'static str {
    if is_new_project(root) {
        MOAI_DIR
    } else {
        dir(root)
    }
}

/// The marker this generation writes last (decision D1), with `extra` fields beside the two
/// that [`classify`] reads.
pub fn marker_bytes(created_by: &str, extra: &[(&str, &str)]) -> Vec<u8> {
    let mut map = serde_json::Map::new();
    map.insert("schema_version".into(), json!(MARKER_SCHEMA));
    map.insert("generation".into(), json!("moai-cli"));
    map.insert(
        "created_by".into(),
        json!(format!("moai-cli {} {created_by}", crate::BUILD_VERSION)),
    );
    for (k, v) in extra {
        map.insert((*k).into(), json!(v));
    }
    let mut bytes = serde_json::to_vec_pretty(&Value::Object(map)).expect("marker encodes");
    bytes.push(b'\n');
    bytes
}

/// Make a new project's `.moai/` and its marker (#2696 C2b), so every later read and write
/// resolves to `.moai`. Does nothing for any other project and says whether it wrote. The
/// folder is created exclusively and the marker published without overwriting.
pub fn adopt_new(root: &Path, created_by: &str) -> Result<bool, String> {
    if !is_new_project(root) {
        return Ok(false);
    }
    fs::create_dir(root.join(MOAI_DIR)).map_err(|_| format!("{MOAI_DIR}: folder create failed"))?;
    crate::apply::publish_exclusive(
        &root.join(MOAI_DIR).join(RUST_MARKER),
        &marker_bytes(created_by, &[]),
    )?;
    Ok(true)
}

/// `root/<dir(root)>/<sub>`; an empty `sub` is the folder itself.
pub fn path(root: &Path, sub: &str) -> std::path::PathBuf {
    let base = root.join(dir(root));
    if sub.is_empty() {
        base
    } else {
        base.join(sub)
    }
}

/// `<dir(root)>/<sub>`, the project-relative spelling receipts and outputs record.
pub fn rel(root: &Path, sub: &str) -> String {
    if sub.is_empty() {
        dir(root).to_string()
    } else {
        format!("{}/{sub}", dir(root))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn write(root: &Path, rel: &str, body: &str) {
        let p = root.join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }

    const MARKER_V1: &str = r#"{"schema_version":1,"generation":"moai-cli","created_by":"moai-cli test","project_id":"p"}"#;
    const SYSTEM_YAML: &str = "# System\nmoai:\n  # MoAI-ADK version\n  version: \"2.4.0\"\n";

    fn verdict(setup: impl FnOnce(&Path)) -> Classified {
        let tmp = tempfile::tempdir().unwrap();
        setup(tmp.path());
        classify(tmp.path())
    }

    #[test]
    fn every_row_of_the_design_table() {
        assert_eq!(verdict(|_| {}).generation, "none");
        assert_eq!(
            verdict(|r| fs::create_dir(r.join(".moai")).unwrap()).generation,
            "empty"
        );
        assert_eq!(
            verdict(|r| write(r, ".moai/notes.txt", "mine")).generation,
            "unknown"
        );
        assert_eq!(
            verdict(|r| write(r, ".moai/manifest.json", "{}")).generation,
            "go"
        );
        assert_eq!(
            verdict(|r| write(r, ".moai/config/config.yaml", "x: 1\n")).generation,
            "go"
        );
        assert_eq!(
            verdict(|r| write(r, ".moai/config/sections/system.yaml", SYSTEM_YAML)).generation,
            "go"
        );
        assert_eq!(
            verdict(|r| write(r, ".moai/moai-cli.json", MARKER_V1)).generation,
            "rust"
        );
        let mixed = verdict(|r| {
            write(r, ".moai/moai-cli.json", MARKER_V1);
            write(r, ".moai/manifest.json", "{}");
        });
        assert_eq!(mixed.generation, "mixed");
        assert_eq!(mixed.go_markers, vec![".moai/manifest.json".to_string()]);
        assert_eq!(
            verdict(|r| write(r, ".moai", "a file")).generation,
            "refuse"
        );
        let interrupted = verdict(|r| {
            write(r, ".moai-backup/manifest.json", "{}");
            write(
                r,
                ".moai/backups/20261113T150000Z/go-state-receipt.json",
                "{}",
            );
        });
        assert_eq!(interrupted.generation, "interrupted");
        assert_eq!(interrupted.interrupted, vec![".moai-backup".to_string()]);
    }

    /// #2696: which Go backups beside which `.moai` mean a stopped migration.
    #[test]
    fn a_go_backup_means_interrupted_only_beside_a_missing_or_unfinished_moai() {
        const RECEIPT: &str = ".moai/backups/20261113T150000Z/go-state-receipt.json";
        let backup = |r: &Path| write(r, ".moai-backup/manifest.json", "{}");
        // Renamed, nothing made yet; and made, marker not yet written (or already removed).
        assert_eq!(verdict(backup).generation, "interrupted");
        let c = verdict(|r| {
            backup(r);
            write(r, RECEIPT, "{}");
        });
        assert_eq!(c.generation, "interrupted");
        // The stamped name counts too; both are listed.
        let c = verdict(|r| {
            backup(r);
            write(r, ".moai-backup-20261113T150000Z/manifest.json", "{}");
        });
        assert_eq!(
            c.interrupted,
            vec![
                ".moai-backup".to_string(),
                ".moai-backup-20261113T150000Z".to_string()
            ]
        );
        // Finished: the marker is valid.
        let c = verdict(|r| {
            backup(r);
            write(r, RECEIPT, "{}");
            write(r, ".moai/moai-cli.json", MARKER_V1);
        });
        assert_eq!((c.generation, c.interrupted.len()), ("rust", 0));
        // Beside the Go folder (a finished revert) or a folder without the receipt.
        let c = verdict(|r| {
            backup(r);
            write(r, ".moai/manifest.json", "{}");
        });
        assert_eq!((c.generation, c.interrupted.len()), ("go", 0));
        let c = verdict(|r| {
            backup(r);
            write(r, ".moai/notes.txt", "mine");
        });
        assert_eq!(c.generation, "unknown");
        // Not backups: the Go tool's plural folder, other suffixes, a receipt not under a stamp.
        for other in [
            ".moai-backups",
            ".moai-backup-old",
            ".moai-backup-20261113T150000",
            ".moai-backup20261113T150000Z",
        ] {
            let c = verdict(|r| write(r, &format!("{other}/x"), "x"));
            assert_eq!(c.generation, "none", "{other}");
        }
        let c = verdict(|r| {
            backup(r);
            write(r, ".moai/backups/latest/go-state-receipt.json", "{}");
        });
        assert_eq!(c.generation, "unknown");
        // #2696 C4a: the earlier-root move has no Go backup; its own `.moai/` holding a receipt
        // and no valid marker is interrupted by itself, and names `.moai`.
        let c = verdict(|r| write(r, RECEIPT, "{}"));
        assert_eq!(
            (c.generation, c.interrupted.clone()),
            ("interrupted", vec![".moai".to_string()])
        );
        let c = verdict(|r| {
            write(r, RECEIPT, "{}");
            write(r, ".moai/moai-cli.json", MARKER_V1);
        });
        assert_eq!(c.generation, "rust");
    }

    #[test]
    fn utc_stamps_are_exactly_sixteen_characters() {
        assert!(utc_stamp("20261113T150000Z"));
        for bad in [
            "",
            "20261113T150000",
            "20261113T150000ZZ",
            "20261113X150000Z",
            "2026111aT150000Z",
            "20261113T15000aZ",
            "20261113T150000z",
        ] {
            assert!(!utc_stamp(bad), "{bad}");
        }
    }

    #[test]
    fn the_rust_marker_must_be_exactly_this_schema() {
        let marker = |body: &str| verdict(|r| write(r, ".moai/moai-cli.json", body));
        let c = marker(r#"{"schema_version":2,"generation":"moai-cli"}"#);
        assert_eq!((c.generation, c.rust_marker), ("refuse", "newer"));
        for bad in [
            r#"{"schema_version":0,"generation":"moai-cli"}"#,
            r#"{"schema_version":1,"generation":"moai"}"#,
            r#"{"schema_version":1}"#,
            r#"{"schema_version":"1","generation":"moai-cli"}"#,
            "[1]",
            "not json",
        ] {
            let c = marker(bad);
            assert_eq!(
                (c.generation, c.rust_marker),
                ("refuse", "unreadable"),
                "{bad}"
            );
        }
        // A folder named like the marker is not one.
        let c = verdict(|r| fs::create_dir_all(r.join(".moai/moai-cli.json")).unwrap());
        assert_eq!((c.generation, c.rust_marker), ("refuse", "unreadable"));
    }

    #[test]
    fn the_marker_size_limit_is_64_kib() {
        // 65536 bytes: a valid marker padded with spaces still counts; 65537 does not.
        let padded = |len: usize| {
            let mut s = MARKER_V1.to_string();
            while s.len() < len {
                s.push(' ');
            }
            s
        };
        let at = verdict(|r| write(r, ".moai/moai-cli.json", &padded(65536)));
        assert_eq!(at.rust_marker, "valid");
        let over = verdict(|r| write(r, ".moai/moai-cli.json", &padded(65537)));
        assert_eq!(
            (over.generation, over.rust_marker),
            ("refuse", "unreadable")
        );
    }

    #[test]
    fn system_yaml_counts_only_with_a_version_under_moai() {
        for (text, want) in [
            (SYSTEM_YAML, true),
            ("moai:\n  version: 1\n", true),
            ("moai:\n  update_check_frequency: daily\n", false),
            ("other:\n  version: 1\nmoai:\n  name: x\n", false),
            ("version: 1\n", false),
            ("moai:\nnext:\n  version: 1\n", false),
        ] {
            assert_eq!(system_yaml_has_version(text), want, "{text:?}");
        }
        let c = verdict(|r| write(r, ".moai/config/sections/system.yaml", "moai:\n  name: x\n"));
        assert_eq!(c.generation, "unknown");
    }

    #[cfg(unix)]
    #[test]
    fn links_are_never_markers() {
        use std::os::unix::fs::symlink;
        // `.moai` itself a link to a real Rust folder: refused, not followed.
        let c = verdict(|r| {
            write(r, "elsewhere/moai-cli.json", MARKER_V1);
            symlink(r.join("elsewhere"), r.join(".moai")).unwrap();
        });
        assert_eq!((c.generation, c.moai_dir), ("refuse", "link"));
        // A linked marker file or a linked `config` folder is not a marker.
        let c = verdict(|r| {
            write(r, "outside/m.json", MARKER_V1);
            fs::create_dir(r.join(".moai")).unwrap();
            symlink(r.join("outside/m.json"), r.join(".moai/moai-cli.json")).unwrap();
        });
        assert_eq!((c.generation, c.rust_marker), ("refuse", "unreadable"));
        let c = verdict(|r| {
            write(r, "outside/config.yaml", "x: 1\n");
            fs::create_dir(r.join(".moai")).unwrap();
            symlink(r.join("outside"), r.join(".moai/config")).unwrap();
        });
        assert_eq!(c.generation, "unknown");
        assert!(c.go_markers.is_empty());
    }

    const MARKER_V2: &str = r#"{"schema_version":2,"generation":"moai-cli"}"#;
    const ACTIVE: &str = ".moai/state/active-sessions.json";

    fn home_verdict(setup: impl FnOnce(&Path)) -> Classified {
        let tmp = tempfile::tempdir().unwrap();
        setup(tmp.path());
        classify_home(tmp.path())
    }

    /// Every name and its bytes under `root`, links described not followed (a no-write check).
    fn snapshot(root: &Path) -> Vec<(String, Vec<u8>)> {
        let mut out = Vec::new();
        let mut stack = vec![root.to_path_buf()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                let rel = p.strip_prefix(root).unwrap().display().to_string();
                if m.is_dir() {
                    stack.push(p);
                    out.push((rel, b"dir".to_vec()));
                } else if m.is_file() {
                    out.push((rel, fs::read(&p).unwrap_or_default()));
                } else {
                    out.push((rel, b"other".to_vec()));
                }
            }
        }
        out.sort();
        out
    }

    /// #2871 I2 R1 (D-M1 6876633358): a schema-2 marker on a HOME is this generation first;
    /// the Go markers beside it are reported as retiring, nothing moves.
    #[test]
    fn a_schema_2_home_marker_is_rust_first() {
        let c = home_verdict(|h| {
            write(h, ".moai/moai-cli.json", MARKER_V2);
            write(h, ACTIVE, "[]");
        });
        assert_eq!((c.generation, c.rust_marker), ("rust-first", "schema-2"));
        assert_eq!(c.go_retire_pending, vec![ACTIVE.to_string()]);
        assert_eq!(c.to_json()["go_retire_pending"], json!([ACTIVE]));
        // Rust-first without Go markers too.
        let c = home_verdict(|h| write(h, ".moai/moai-cli.json", MARKER_V2));
        assert_eq!(c.generation, "rust-first");
        assert!(c.go_retire_pending.is_empty());
        // Every HOME without a schema-2 marker keeps its verdict, and its JSON keys.
        for (setup, want) in [
            (Box::new(|_: &Path| {}) as Box<dyn Fn(&Path)>, "none"),
            (
                Box::new(|h: &Path| fs::create_dir(h.join(".moai")).unwrap()),
                "empty",
            ),
            (Box::new(|h: &Path| write(h, ACTIVE, "[]")), "go"),
            (
                Box::new(|h: &Path| write(h, ".moai/moai-cli.json", MARKER_V1)),
                "rust",
            ),
            (
                Box::new(|h: &Path| {
                    write(h, ".moai/moai-cli.json", MARKER_V1);
                    write(h, ACTIVE, "[]");
                }),
                "mixed",
            ),
        ] {
            let c = home_verdict(|h| setup(h));
            assert_eq!(c.generation, want);
            assert!(c.to_json().get("go_retire_pending").is_none(), "{want}");
            let tmp = tempfile::tempdir().unwrap();
            setup(tmp.path());
            assert_eq!(
                classify_home(tmp.path()),
                classify_with(tmp.path(), crate::home_state::home_go_markers),
                "{want}"
            );
        }
        // Projects never read schema 2: an older-style refusal, as before.
        let c = verdict(|r| {
            write(r, ".moai/moai-cli.json", MARKER_V2);
            write(r, ".moai/manifest.json", "{}");
        });
        assert_eq!((c.generation, c.rust_marker), ("refuse", "newer"));
    }

    #[test]
    fn schema_2_home_marker_boundaries() {
        let refused = |c: &Classified, marker: &str| {
            assert_eq!((c.generation, c.rust_marker), ("refuse", marker));
        };
        // Schema 3 is newer than this build.
        refused(
            &home_verdict(|h| {
                write(
                    h,
                    ".moai/moai-cli.json",
                    r#"{"schema_version":3,"generation":"moai-cli"}"#,
                )
            }),
            "newer",
        );
        // 65536 bytes still reads; 65537 does not.
        let padded = |len: usize| {
            let mut s = MARKER_V2.to_string();
            while s.len() < len {
                s.push(' ');
            }
            s
        };
        let at = home_verdict(|h| write(h, ".moai/moai-cli.json", &padded(65536)));
        assert_eq!(at.generation, "rust-first");
        let tmp = tempfile::tempdir().unwrap();
        let h = tmp.path();
        write(h, ".moai/moai-cli.json", &padded(65537));
        let before = snapshot(h);
        refused(&classify_home(h), "unreadable");
        assert_eq!(snapshot(h), before, "nothing written");
        // A folder in the marker's place.
        refused(
            &home_verdict(|h| fs::create_dir_all(h.join(".moai/moai-cli.json")).unwrap()),
            "unreadable",
        );
        #[cfg(unix)]
        {
            use std::os::unix::fs::symlink;
            let tmp = tempfile::tempdir().unwrap();
            let h = tmp.path();
            write(h, "outside/m.json", MARKER_V2);
            fs::create_dir(h.join(".moai")).unwrap();
            symlink(h.join("outside/m.json"), h.join(".moai/moai-cli.json")).unwrap();
            let before = snapshot(h);
            refused(&classify_home(h), "unreadable");
            assert_eq!(snapshot(h), before, "nothing written");
            // A FIFO is never opened (opening it would block).
            let tmp = tempfile::tempdir().unwrap();
            let h = tmp.path();
            fs::create_dir(h.join(".moai")).unwrap();
            let status = std::process::Command::new("mkfifo")
                .arg(h.join(".moai/moai-cli.json"))
                .status()
                .unwrap();
            assert!(status.success());
            let before = snapshot(h);
            refused(&classify_home(h), "unreadable");
            assert_eq!(snapshot(h), before, "nothing written");
            // A linked `.moai` holding a schema-2 marker is not followed.
            let c = home_verdict(|h| {
                write(h, "elsewhere/moai-cli.json", MARKER_V2);
                symlink(h.join("elsewhere"), h.join(".moai")).unwrap();
            });
            assert_eq!((c.generation, c.moai_dir), ("refuse", "link"));
        }
    }

    /// The secret Go marker is seen by its metadata only: a 0o000 `credentials.yaml` beside a
    /// schema-2 marker is listed, never opened.
    #[cfg(unix)]
    #[test]
    fn home_classification_never_opens_secret_files() {
        use std::os::unix::fs::PermissionsExt;
        let tmp = tempfile::tempdir().unwrap();
        let h = tmp.path();
        write(h, ".moai/moai-cli.json", MARKER_V2);
        write(h, ".moai/credentials.yaml", "token: x\n");
        let secret = h.join(".moai/credentials.yaml");
        fs::set_permissions(&secret, fs::Permissions::from_mode(0o000)).unwrap();
        let c = classify_home(h);
        fs::set_permissions(&secret, fs::Permissions::from_mode(0o600)).unwrap();
        assert_eq!(c.generation, "rust-first");
        assert_eq!(
            c.go_retire_pending,
            vec![".moai/credentials.yaml".to_string()]
        );
    }

    #[test]
    fn the_legacy_root_is_reported_beside_the_verdict() {
        let c = verdict(|r| {
            write(
                r,
                &format!("{}/project.yaml", crate::PROJECT_STATE_DIR),
                "{}",
            )
        });
        assert_eq!(c.generation, "none");
        assert!(c.legacy_root);
        assert_eq!(c.to_json()["legacy_root"], json!(crate::PROJECT_STATE_DIR));
        assert_eq!(verdict(|_| {}).to_json()["legacy_root"], Value::Null);
    }
}
