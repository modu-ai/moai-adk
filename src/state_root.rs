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
//! Rust marker: `.moai/moai.json` = `{"schema_version": 1, "generation": "moai", …}`, written
//! last by a migration (decision D1), or the earlier `.moai/moai-cli.json` with `"generation":
//! "moai-cli"` (the rename of #2871, leader decision 6886824285 item 2: readers accept both
//! names, writers write only the new one). Each name pairs with its own generation value; a
//! mismatched pair is not a marker. `moai.json` alone decides when it exists. Every path is read with
//! `symlink_metadata`; a link anywhere on the way is not a marker. Writes nothing.
//!
//! A Go backup is the Go folder after `go-state-apply` renamed it (#2725 rev2 D2, #2696):
//! exactly `.moai-backup`, or `.moai-backup-<UTC>` when that name was taken. The Go tool's own
//! `.moai-backups/` (plural) is not one.

use serde_json::{json, Value};
use std::{fs, path::Path};

/// The Go and Rust folder name once the state has moved (#2725).
pub const MOAI_DIR: &str = ".moai";
/// The Rust generation marker inside `.moai/` that every writer writes (decision D1), with
/// `"generation": "moai"` ([`NEW_GENERATION`]). #2871 marker rename (leader decision 6886824285
/// item 2): the writers flipped to this name after the readers learned both; an existing
/// [`LEGACY_MARKER`] is read, never rewritten, renamed or deleted. The one exception (leader
/// audit 6900128841 B1): `home-install-apply` and `home-install-revert` of an install an
/// earlier build began write, find and move the marker under the name that goes with the
/// `generation` in its saved bytes, so a resumed install ends with `moai-cli.json` and its
/// revert takes that file into the journal folder.
pub const RUST_MARKER: &str = NEW_MARKER;
/// The marker name of this generation, paired with [`NEW_GENERATION`].
pub const NEW_MARKER: &str = "moai.json";
/// The earlier marker name, paired with [`LEGACY_GENERATION`]. Read by every reader; written
/// or moved only by `home-install-apply`/`home-install-revert` for an install an earlier build
/// began (see [`RUST_MARKER`]).
pub const LEGACY_MARKER: &str = "moai-cli.json";
/// `"generation"` inside [`LEGACY_MARKER`].
pub const LEGACY_GENERATION: &str = "moai-cli";
/// `"generation"` inside [`NEW_MARKER`].
pub const NEW_GENERATION: &str = "moai";

/// Either marker file name, ASCII case-insensitive (APFS and NTFS fold the case, so `MOAI.JSON`
/// is the same file there).
pub fn is_marker_name(name: &str) -> bool {
    name.eq_ignore_ascii_case(LEGACY_MARKER) || name.eq_ignore_ascii_case(NEW_MARKER)
}
/// The marker schema this build writes and understands.
pub const MARKER_SCHEMA: u64 = 1;
/// #2871 I2 R1: a HOME marker with this schema makes the HOME this generation's first
/// (`rust-first`, user decision D-M1 6876633358), Go markers or not. Only
/// [`classify_home`] reads it; projects and older builds refuse it as `newer`. Only
/// `home-install-apply` writes it ([`home_marker_bytes`]), as the last step of the move.
pub const HOME_MARKER_SCHEMA: u64 = 2;
/// The earlier generation's project root folder, for the one place that states it as a leftover
/// and for the workflow reference that still names it (#2936 L2b).
pub const EARLIER_PROJECT_DIR: &str = crate::PROJECT_STATE_DIR;
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
    /// The marker file that decided `rust_marker` (`moai.json` or `moai-cli.json`); `None` when
    /// there is no marker.
    pub marker_file: Option<&'static str>,
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
        // Likewise `marker_file` appears only when a marker exists.
        if let Some(file) = self.marker_file {
            v["marker_file"] = json!(file);
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

/// What one marker file says (`.moai/moai-cli.json` or `.moai/moai.json`): `valid`, `absent`, `newer` or `unreadable`; `schema-2` too when
/// `accept` is [`HOME_MARKER_SCHEMA`] (HOME only). Old builds read schema 2 as `newer` and
/// refuse it. `schema` is the number found for `valid`, `schema-2` and `newer`.
struct OneMarker {
    status: &'static str,
    schema: Option<u64>,
}

/// Read `moai/<name>` and demand `generation` as its generation: `moai-cli.json` carries
/// `"moai-cli"`, `moai.json` carries `"moai"` and no other pairing is a marker. A file that is
/// not a regular file (a link, a folder, a FIFO) is `unreadable` and never opened. `Absent`
/// follows the old rule for the old name (any failed `lstat`) and only a missing name for the
/// new one, so a name that exists but cannot be examined still decides.
fn read_one_marker(moai: &Path, name: &str, generation: &str, accept: u64) -> OneMarker {
    let unreadable = OneMarker {
        status: "unreadable",
        schema: None,
    };
    match fs::symlink_metadata(moai.join(name)) {
        Err(e) if name == LEGACY_MARKER || e.kind() == std::io::ErrorKind::NotFound => {
            return OneMarker {
                status: "absent",
                schema: None,
            }
        }
        Err(_) => return unreadable,
        Ok(m) if !m.is_file() => return unreadable,
        Ok(_) => {}
    }
    let Some(bytes) = read_limited(moai, &[name], MARKER_LIMIT) else {
        return unreadable;
    };
    let Ok(Value::Object(map)) = serde_json::from_slice::<Value>(&bytes) else {
        return unreadable;
    };
    if map.get("generation").and_then(Value::as_str) != Some(generation) {
        return unreadable;
    }
    match map.get("schema_version").and_then(Value::as_u64) {
        Some(MARKER_SCHEMA) => OneMarker {
            status: "valid",
            schema: Some(MARKER_SCHEMA),
        },
        Some(HOME_MARKER_SCHEMA) if accept >= HOME_MARKER_SCHEMA => OneMarker {
            status: "schema-2",
            schema: Some(HOME_MARKER_SCHEMA),
        },
        Some(v) if v > MARKER_SCHEMA => OneMarker {
            status: "newer",
            schema: Some(v),
        },
        _ => unreadable,
    }
}

/// The decided marker of a `.moai/`: its status string, the file that decided it, and, when the
/// two names disagree, why (#2871 marker rename, leader decision 6886824285 item 2).
struct Marker {
    status: &'static str,
    file: Option<&'static str>,
    conflict: Option<String>,
}

/// Read both marker names; reads only, nothing is renamed, rewritten or deleted.
///
/// - `moai.json` present (any `lstat` answer but "not found"): it alone decides. The old file is
///   never a fallback for it.
/// - `moai.json` absent: `moai-cli.json` decides, exactly as before the rename.
/// - Both present and the new one readable: the new one decides unless the old one is `newer`
///   or carries a higher schema than the new one (`markers disagree`, refused). An old
///   schema 1 beside a new schema 2 is the expected pair of a HOME install, and an unreadable
///   old file beside a good new one is ignored.
fn rust_marker(moai: &Path, accept: u64) -> Marker {
    let new = read_one_marker(moai, NEW_MARKER, NEW_GENERATION, accept);
    let old = read_one_marker(moai, LEGACY_MARKER, LEGACY_GENERATION, accept);
    if new.status == "absent" {
        return Marker {
            status: old.status,
            file: (old.status != "absent").then_some(LEGACY_MARKER),
            conflict: None,
        };
    }
    let decided = |conflict: Option<String>| Marker {
        status: if conflict.is_some() {
            "newer"
        } else {
            new.status
        },
        file: Some(NEW_MARKER),
        conflict,
    };
    if !matches!(new.status, "valid" | "schema-2") {
        return decided(None);
    }
    let new_schema = new.schema.unwrap_or(MARKER_SCHEMA);
    let disagrees = old.status == "newer"
        || (matches!(old.status, "valid" | "schema-2") && old.schema.unwrap_or(0) > new_schema);
    if disagrees {
        // Both files carry a schema here (`newer`, `valid` and `schema-2` all do).
        let old_schema = old.schema.unwrap_or(0);
        return decided(Some(format!(
            "markers disagree: {LEGACY_MARKER} is schema {old_schema} and {NEW_MARKER} is schema \
             {new_schema}; this build leaves both alone"
        )));
    }
    decided(None)
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
    let (go, found, entries) = if moai_dir == "folder" {
        let entries = fs::read_dir(&moai).map(|rd| rd.count() as u64).unwrap_or(0);
        (go_markers(&moai), rust_marker(&moai, accept), entries)
    } else {
        (
            Vec::new(),
            Marker {
                status: "absent",
                file: None,
                conflict: None,
            },
            0,
        )
    };
    let marker = found.status;
    // The file that decided, for the reasons below.
    let decided_by = found.file.unwrap_or(NEW_MARKER);
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
                    found.conflict.clone().unwrap_or_else(|| {
                        format!(
                            "{MOAI_DIR}/{decided_by} is from a newer moai-cli; this build leaves it alone"
                        )
                    }),
                ),
                ("unreadable", _) => (
                    "refuse",
                    format!("{MOAI_DIR}/{decided_by} is not a marker this build can read"),
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
        marker_file: found.file,
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
    map.insert("generation".into(), json!(NEW_GENERATION));
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

/// The schema-2 marker `home-install-apply` writes last (#2871 U3c, user decision D-M1): the
/// two fields [`classify_home`] reads, who wrote it and the stamp of the install it closes.
pub fn home_marker_bytes(created_by: &str, stamp: &str) -> Vec<u8> {
    let mut map = serde_json::Map::new();
    map.insert("schema_version".into(), json!(HOME_MARKER_SCHEMA));
    map.insert("generation".into(), json!(NEW_GENERATION));
    map.insert(
        "created_by".into(),
        json!(format!("moai-cli {} {created_by}", crate::BUILD_VERSION)),
    );
    map.insert("install_stamp".into(), json!(stamp));
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

    /// #2871 U3c: the marker `home-install-apply` writes is the schema-2 marker the HOME verdict
    /// reads, and it carries the install stamp; a project root refuses it as `newer`.
    #[test]
    fn the_home_marker_bytes_make_a_home_rust_first() {
        let bytes = home_marker_bytes("home-install-apply", "20261114T120000Z");
        let text = String::from_utf8(bytes.clone()).unwrap();
        let v: Value = serde_json::from_str(&text).unwrap();
        assert_eq!(v["schema_version"], json!(HOME_MARKER_SCHEMA));
        assert_eq!(v["generation"], json!(NEW_GENERATION));
        assert_eq!(v["install_stamp"], json!("20261114T120000Z"));
        assert!(v["created_by"]
            .as_str()
            .unwrap()
            .ends_with("home-install-apply"));
        assert!(text.ends_with('\n'));
        let c = home_verdict(|h| {
            fs::create_dir_all(h.join(".moai")).unwrap();
            fs::write(h.join(".moai").join(RUST_MARKER), &bytes).unwrap();
        });
        assert_eq!((c.generation, c.rust_marker), ("rust-first", "schema-2"));
        let project = verdict(|r| {
            fs::create_dir_all(r.join(".moai")).unwrap();
            fs::write(r.join(".moai").join(RUST_MARKER), &bytes).unwrap();
        });
        assert_eq!(project.rust_marker, "newer");
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

    // ---- #2871 marker rename, PR-A: read both names, write unchanged ----

    /// Both file names with the generation each one is paired with.
    const NAMES: [(&str, &str); 2] = [("moai-cli.json", "moai-cli"), ("moai.json", "moai")];

    fn body(generation: &str, schema: u64) -> String {
        format!(r#"{{"schema_version":{schema},"generation":"{generation}"}}"#)
    }

    fn put(root: &Path, name: &str, text: &str) {
        write(root, &format!(".moai/{name}"), text);
    }

    fn padded(text: &str, len: usize) -> String {
        let mut s = text.to_string();
        while s.len() < len {
            s.push(' ');
        }
        s
    }

    /// Project and HOME verdicts of the same setup: `(generation, rust_marker, marker_file)`.
    fn both(setup: impl Fn(&Path)) -> [(&'static str, &'static str, Option<&'static str>); 2] {
        let p = verdict(&setup);
        let h = home_verdict(&setup);
        [
            (p.generation, p.rust_marker, p.marker_file),
            (h.generation, h.rust_marker, h.marker_file),
        ]
    }

    #[test]
    fn the_marker_names_and_generations_are_pinned() {
        assert_eq!(RUST_MARKER, "moai.json", "writers write the new name");
        assert_eq!((NEW_MARKER, NEW_GENERATION), ("moai.json", "moai"));
        assert_eq!(
            (LEGACY_MARKER, LEGACY_GENERATION),
            ("moai-cli.json", "moai-cli")
        );
        for name in ["moai.json", "moai-cli.json", "MOAI.JSON", "Moai-Cli.Json"] {
            assert!(is_marker_name(name), "{name}");
        }
        for name in [
            "moai.json.bak",
            "moai",
            "moai-json",
            "xmoai.json",
            "",
            "moai.json ",
        ] {
            assert!(!is_marker_name(name), "{name:?}");
        }
        // The bytes a writer produces carry the new pair, and read back as a valid marker.
        let v: Value = serde_json::from_slice(&marker_bytes("t", &[])).unwrap();
        assert_eq!(v["generation"], json!("moai"));
        assert_eq!(v["schema_version"], json!(MARKER_SCHEMA));
        let c = verdict(|r| {
            put(
                r,
                RUST_MARKER,
                &String::from_utf8(marker_bytes("t", &[])).unwrap(),
            )
        });
        assert_eq!((c.generation, c.marker_file), ("rust", Some("moai.json")));
    }

    /// A new project gets `.moai/moai.json` with `"moai"` and no `moai-cli.json`; an old-name
    /// marker beside it (not a new project: `.moai` exists) is read, and left byte for byte.
    #[test]
    fn adopt_new_writes_only_the_new_name() {
        let tmp = tempfile::tempdir().unwrap();
        assert!(adopt_new(tmp.path(), "test").unwrap());
        let moai = tmp.path().join(".moai");
        assert!(moai.join("moai.json").is_file());
        assert!(fs::symlink_metadata(moai.join("moai-cli.json")).is_err());
        let v: Value = serde_json::from_slice(&fs::read(moai.join("moai.json")).unwrap()).unwrap();
        assert_eq!(v["generation"], json!("moai"));
        let c = classify(tmp.path());
        assert_eq!((c.generation, c.marker_file), ("rust", Some("moai.json")));
        // A project that already has an old-name marker is not new: nothing is written, the
        // old file stays as it was and keeps deciding.
        let old = tempfile::tempdir().unwrap();
        put(old.path(), "moai-cli.json", MARKER_V1);
        let before = snapshot(old.path());
        assert!(!adopt_new(old.path(), "test").unwrap());
        assert_eq!(
            snapshot(old.path()),
            before,
            "an old-name marker is left untouched"
        );
        assert_eq!(dir(old.path()), MOAI_DIR);
        assert_eq!(classify(old.path()).marker_file, Some("moai-cli.json"));
    }

    /// Each name alone: valid for a project, schema 2 for a HOME, newer and unreadable refused,
    /// and the reason names the file that decided.
    #[test]
    fn either_marker_name_decides_alone() {
        for (name, generation) in NAMES {
            let v1 = body(generation, 1);
            let v2 = body(generation, 2);
            let got = both(|r| put(r, name, &v1));
            assert_eq!(got[0], ("rust", "valid", Some(name)), "{name}");
            assert_eq!(got[1], ("rust", "valid", Some(name)), "{name}");
            // Schema 2: a HOME is rust-first, a project refuses it as newer.
            let got = both(|r| put(r, name, &v2));
            assert_eq!(got[0], ("refuse", "newer", Some(name)), "{name}");
            assert_eq!(got[1], ("rust-first", "schema-2", Some(name)), "{name}");
            // Schema 3 is newer for both; schema 0 and a wrong shape are unreadable.
            let got = both(|r| put(r, name, &body(generation, 3)));
            assert_eq!(got[0], ("refuse", "newer", Some(name)), "{name}");
            assert_eq!(got[1], ("refuse", "newer", Some(name)), "{name}");
            assert!(verdict(|r| put(r, name, &body(generation, 3)))
                .reason
                .contains(name));
            for bad in [
                body(generation, 0),
                format!(r#"{{"schema_version":"1","generation":"{generation}"}}"#),
                format!(r#"{{"generation":"{generation}"}}"#),
                r#"{"schema_version":1}"#.to_string(),
                "[1]".to_string(),
                "not json".to_string(),
            ] {
                let got = both(|r| put(r, name, &bad));
                assert_eq!(got[0], ("refuse", "unreadable", Some(name)), "{name} {bad}");
                assert_eq!(got[1], ("refuse", "unreadable", Some(name)), "{name} {bad}");
                let c = verdict(|r| put(r, name, &bad));
                assert!(c.reason.contains(name), "{name}: {}", c.reason);
            }
        }
        // `moai.json` beside Go markers is `mixed`, as the old name is.
        let c = verdict(|r| {
            put(r, "moai.json", &body("moai", 1));
            write(r, ".moai/manifest.json", "{}");
        });
        assert_eq!((c.generation, c.marker_file), ("mixed", Some("moai.json")));
        // `dir` follows the verdict: the new name alone is a `rust` project.
        let tmp = tempfile::tempdir().unwrap();
        put(tmp.path(), "moai.json", &body("moai", 1));
        assert_eq!(dir(tmp.path()), MOAI_DIR);
        assert!(!is_new_project(tmp.path()));
    }

    /// The name and the generation travel together: either mismatch is not a marker.
    #[test]
    fn a_name_with_the_other_generation_is_refused() {
        for (name, other) in [("moai-cli.json", "moai"), ("moai.json", "moai-cli")] {
            for schema in [1, 2] {
                let got = both(|r| put(r, name, &body(other, schema)));
                assert_eq!(
                    got[0],
                    ("refuse", "unreadable", Some(name)),
                    "{name} {schema}"
                );
                assert_eq!(
                    got[1],
                    ("refuse", "unreadable", Some(name)),
                    "{name} {schema}"
                );
            }
        }
        // A good pair beside a crossed one: the crossed `moai.json` decides alone (refuse), and
        // a crossed old file beside a good new one is ignored.
        let got = both(|r| {
            put(r, "moai-cli.json", &body("moai-cli", 1));
            put(r, "moai.json", &body("moai-cli", 1));
        });
        assert_eq!(got[0], ("refuse", "unreadable", Some("moai.json")));
        let got = both(|r| {
            put(r, "moai-cli.json", &body("moai", 1));
            put(r, "moai.json", &body("moai", 1));
        });
        assert_eq!(got[0], ("rust", "valid", Some("moai.json")));
    }

    /// Classifying never writes, renames or deletes: one tree snapshot before and after, for
    /// each name alone and for both, in a project and in a HOME.
    #[test]
    fn classify_changes_nothing_for_either_name() {
        let setups: [Vec<(&str, String)>; 4] = [
            vec![("moai-cli.json", body("moai-cli", 1))],
            vec![("moai.json", body("moai", 1))],
            vec![
                ("moai-cli.json", body("moai-cli", 1)),
                ("moai.json", body("moai", 2)),
            ],
            vec![
                ("moai-cli.json", body("moai-cli", 3)),
                ("moai.json", body("moai", 1)),
            ],
        ];
        for files in &setups {
            let tmp = tempfile::tempdir().unwrap();
            let r = tmp.path();
            for (name, text) in files {
                put(r, name, text);
            }
            write(r, ACTIVE, "[]");
            let before = snapshot(r);
            let _ = classify(r);
            let _ = classify_home(r);
            let _ = classify_with(r, crate::home_state::home_go_markers);
            let _ = dir(r);
            let _ = is_new_project(r);
            assert_eq!(snapshot(r), before, "{files:?}");
        }
    }

    #[test]
    fn the_marker_size_limit_is_64_kib_at_both_names() {
        for (name, generation) in NAMES {
            // A project reads schema 1, a HOME schema 2: 65536 bytes pass, 65537 do not.
            let p1 = body(generation, 1);
            let p2 = body(generation, 2);
            let at = verdict(|r| put(r, name, &padded(&p1, 65536)));
            assert_eq!((at.generation, at.rust_marker), ("rust", "valid"), "{name}");
            let over = verdict(|r| put(r, name, &padded(&p1, 65537)));
            assert_eq!(
                (over.generation, over.rust_marker),
                ("refuse", "unreadable"),
                "{name}"
            );
            let at = home_verdict(|h| put(h, name, &padded(&p2, 65536)));
            assert_eq!(
                (at.generation, at.rust_marker),
                ("rust-first", "schema-2"),
                "{name}"
            );
            let over = home_verdict(|h| put(h, name, &padded(&p2, 65537)));
            assert_eq!(
                (over.generation, over.rust_marker),
                ("refuse", "unreadable"),
                "{name}"
            );
        }
    }

    /// A link, a folder or a FIFO in the place of either marker is `unreadable`; none is
    /// followed or opened (opening a FIFO would block), and nothing is written.
    #[cfg(unix)]
    #[test]
    fn a_link_a_folder_or_a_fifo_is_no_marker_at_either_name() {
        use std::os::unix::fs::symlink;
        for (name, generation) in NAMES {
            for home in [false, true] {
                let schema = if home { 2 } else { 1 };
                let classify_it = |r: &Path| {
                    if home {
                        classify_home(r)
                    } else {
                        classify(r)
                    }
                };
                // A link to a good marker.
                let tmp = tempfile::tempdir().unwrap();
                let r = tmp.path();
                write(r, "outside/m.json", &body(generation, schema));
                fs::create_dir(r.join(".moai")).unwrap();
                symlink(r.join("outside/m.json"), r.join(".moai").join(name)).unwrap();
                let before = snapshot(r);
                let c = classify_it(r);
                assert_eq!(
                    (c.generation, c.rust_marker, c.marker_file),
                    ("refuse", "unreadable", Some(name)),
                    "link {name} home={home}"
                );
                assert_eq!(snapshot(r), before, "link {name}");
                // A dangling link too.
                let tmp = tempfile::tempdir().unwrap();
                let r = tmp.path();
                fs::create_dir(r.join(".moai")).unwrap();
                symlink(r.join("nowhere"), r.join(".moai").join(name)).unwrap();
                let c = classify_it(r);
                assert_eq!(
                    (c.generation, c.rust_marker),
                    ("refuse", "unreadable"),
                    "dangling {name} home={home}"
                );
                // A folder.
                let tmp = tempfile::tempdir().unwrap();
                let r = tmp.path();
                fs::create_dir_all(r.join(".moai").join(name)).unwrap();
                let c = classify_it(r);
                assert_eq!(
                    (c.generation, c.rust_marker, c.marker_file),
                    ("refuse", "unreadable", Some(name)),
                    "folder {name} home={home}"
                );
                // A FIFO: `lstat` says it is no regular file, so it is never opened.
                let tmp = tempfile::tempdir().unwrap();
                let r = tmp.path();
                fs::create_dir(r.join(".moai")).unwrap();
                let status = std::process::Command::new("mkfifo")
                    .arg(r.join(".moai").join(name))
                    .status()
                    .unwrap();
                assert!(status.success());
                let before = snapshot(r);
                let c = classify_it(r);
                assert_eq!(
                    (c.generation, c.rust_marker, c.marker_file),
                    ("refuse", "unreadable", Some(name)),
                    "fifo {name} home={home}"
                );
                assert_eq!(snapshot(r), before, "fifo {name}");
            }
        }
    }

    /// Both files present: `moai.json` alone decides when it exists; the old file is no
    /// fallback for it, and is ignored unless it is newer or higher than the new one.
    #[test]
    fn with_both_marker_files_the_new_one_decides() {
        let old = |schema| body("moai-cli", schema);
        let new = |schema| body("moai", schema);
        // Equal and readable: new wins, and the verdict says so.
        let got = both(|r| {
            put(r, "moai-cli.json", &old(1));
            put(r, "moai.json", &new(1));
        });
        assert_eq!(got[0], ("rust", "valid", Some("moai.json")));
        assert_eq!(got[1], ("rust", "valid", Some("moai.json")));
        // An old v1 beside a new v2 is the expected pair of a HOME install, not a conflict.
        let got = both(|r| {
            put(r, "moai-cli.json", &old(1));
            put(r, "moai.json", &new(2));
        });
        assert_eq!(got[1], ("rust-first", "schema-2", Some("moai.json")));
        // For a project the new v2 is newer than this build.
        assert_eq!(got[0], ("refuse", "newer", Some("moai.json")));
        // An unreadable old file beside a good new one is ignored.
        for bad in [
            "not json",
            "[1]",
            r#"{"schema_version":1,"generation":"moai"}"#,
        ] {
            let got = both(|r| {
                put(r, "moai-cli.json", bad);
                put(r, "moai.json", &new(1));
            });
            assert_eq!(got[0], ("rust", "valid", Some("moai.json")), "{bad}");
            assert_eq!(got[1], ("rust", "valid", Some("moai.json")), "{bad}");
        }
        // The old file as a folder or a link is ignored too.
        let c = verdict(|r| {
            put(r, "moai.json", &new(1));
            fs::create_dir_all(r.join(".moai/moai-cli.json")).unwrap();
        });
        assert_eq!((c.generation, c.marker_file), ("rust", Some("moai.json")));
        // A new file that is unreadable or newer refuses on its own state; the good old file
        // is not a fallback.
        for (text, marker) in [
            ("not json", "unreadable"),
            (
                r#"{"schema_version":1,"generation":"moai-cli"}"#,
                "unreadable",
            ),
            (r#"{"schema_version":3,"generation":"moai"}"#, "newer"),
        ] {
            let got = both(|r| {
                put(r, "moai-cli.json", &old(1));
                put(r, "moai.json", text);
            });
            assert_eq!(got[0], ("refuse", marker, Some("moai.json")), "{text}");
            assert_eq!(got[1], ("refuse", marker, Some("moai.json")), "{text}");
            let c = verdict(|r| {
                put(r, "moai-cli.json", &old(1));
                put(r, "moai.json", text);
            });
            assert!(c.reason.contains("moai.json"), "{}", c.reason);
        }
        // A `moai.json` that is a link or a folder is "present": the old file does not decide.
        #[cfg(unix)]
        {
            let c = verdict(|r| {
                put(r, "moai-cli.json", &old(1));
                write(r, "outside/m.json", &new(1));
                std::os::unix::fs::symlink(r.join("outside/m.json"), r.join(".moai/moai.json"))
                    .unwrap();
            });
            assert_eq!((c.generation, c.marker_file), ("refuse", Some("moai.json")));
        }
        let c = verdict(|r| {
            put(r, "moai-cli.json", &old(1));
            fs::create_dir_all(r.join(".moai/moai.json")).unwrap();
        });
        assert_eq!((c.generation, c.rust_marker), ("refuse", "unreadable"));
        // `moai.json` absent: the old name decides exactly as before.
        let c = verdict(|r| put(r, "moai-cli.json", &old(1)));
        assert_eq!(
            (c.generation, c.marker_file),
            ("rust", Some("moai-cli.json"))
        );
    }

    /// Two markers that disagree refuse and say so: the old file newer than this build, or with
    /// a higher schema than the new one. Nothing is touched.
    #[test]
    fn markers_that_disagree_are_refused() {
        // Old v2 beside new v1 (a HOME reads both): refuse.
        let tmp = tempfile::tempdir().unwrap();
        let h = tmp.path();
        put(h, "moai-cli.json", &body("moai-cli", 2));
        put(h, "moai.json", &body("moai", 1));
        let before = snapshot(h);
        let c = classify_home(h);
        assert_eq!((c.generation, c.rust_marker), ("refuse", "newer"));
        assert!(c.reason.contains("markers disagree"), "{}", c.reason);
        assert!(c.reason.contains("moai-cli.json") && c.reason.contains("moai.json"));
        assert_eq!(c.marker_file, Some("moai.json"));
        assert_eq!(snapshot(h), before, "nothing written");
        // Old schema 3 (newer than this build) beside a good new file: refuse, home and project.
        for schema in [3, 7] {
            let got = both(|r| {
                put(r, "moai-cli.json", &body("moai-cli", schema));
                put(r, "moai.json", &body("moai", 1));
            });
            assert_eq!(got[0].0, "refuse", "{schema}");
            assert_eq!(got[1].0, "refuse", "{schema}");
        }
        // A project reads an old v2 as newer, so beside a new v1 it disagrees too.
        let c = verdict(|r| {
            put(r, "moai-cli.json", &body("moai-cli", 2));
            put(r, "moai.json", &body("moai", 1));
        });
        assert_eq!(c.generation, "refuse");
        assert!(c.reason.contains("markers disagree"), "{}", c.reason);
        // Equal schemas and a lower old schema are not a disagreement.
        let c = home_verdict(|h| {
            put(h, "moai-cli.json", &body("moai-cli", 2));
            put(h, "moai.json", &body("moai", 2));
        });
        assert_eq!(c.generation, "rust-first");
    }

    /// `marker_file` is in the JSON only when a marker exists; every other key is as before.
    #[test]
    fn the_json_names_the_deciding_file_only_when_there_is_one() {
        let none = verdict(|_| {}).to_json();
        assert!(none.get("marker_file").is_none(), "{none}");
        let unknown = verdict(|r| write(r, ".moai/notes.txt", "x")).to_json();
        assert!(unknown.get("marker_file").is_none(), "{unknown}");
        for (name, generation) in NAMES {
            let j = verdict(|r| put(r, name, &body(generation, 1))).to_json();
            assert_eq!(j["marker_file"], json!(name), "{j}");
            assert_eq!(j["rust_marker"], json!("valid"), "{j}");
            assert_eq!(j["generation"], json!("rust"), "{j}");
        }
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
