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
//! | a `.moai.migrating-<UTC>` beside it | `interrupted` (a stopped migration; recovery first) |
//!
//! Go markers: `.moai/manifest.json` (the Go deploy manifest), `.moai/config/config.yaml`,
//! and `.moai/config/sections/system.yaml` with a `moai:` block holding `version:`.
//! Rust marker: `.moai/moai-cli.json` = `{"schema_version": 1, "generation": "moai-cli", …}`,
//! written last by a migration (decision D1). Every path is read with `symlink_metadata`;
//! a link anywhere on the way is not a marker. Writes nothing.

use serde_json::{json, Value};
use std::{fs, path::Path};

/// The Go and Rust folder name once the state has moved (#2725).
pub const MOAI_DIR: &str = ".moai";
/// The Rust generation marker inside `.moai/` (decision D1).
pub const RUST_MARKER: &str = "moai-cli.json";
/// The marker schema this build writes and understands.
pub const MARKER_SCHEMA: u64 = 1;
/// The name prefix a migration gives the Go folder while it moves it (design §3).
pub const MIGRATING_PREFIX: &str = ".moai.migrating-";
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
    /// `.moai.migrating-*` entries beside `.moai`, sorted.
    pub interrupted: Vec<String>,
    /// Entries directly inside `.moai/` (0 unless it is a real folder).
    pub entries: u64,
    /// The earlier Rust project root (`.moai-cli/project.yaml`) is present.
    pub legacy_root: bool,
    pub reason: String,
}

impl Classified {
    pub fn to_json(&self) -> Value {
        json!({
            "generation": self.generation,
            "moai_dir": self.moai_dir,
            "go_markers": self.go_markers,
            "rust_marker": self.rust_marker,
            "interrupted": self.interrupted,
            "entries": self.entries,
            "legacy_root": self.legacy_root.then_some(crate::PROJECT_STATE_DIR),
            "reason": self.reason,
        })
    }
}

/// A regular file at `root/parts…`, reached without following any link.
fn regular_file(root: &Path, parts: &[&str]) -> Option<fs::Metadata> {
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

fn read_limited(root: &Path, parts: &[&str], limit: u64) -> Option<Vec<u8>> {
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

/// `valid`, `absent`, `newer` or `unreadable` for `.moai/moai-cli.json`.
fn rust_marker(moai: &Path) -> &'static str {
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
        Some(v) if v > MARKER_SCHEMA => "newer",
        _ => "unreadable",
    }
}

fn migrating_entries(root: &Path) -> Vec<String> {
    let mut out: Vec<String> = fs::read_dir(root)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .filter_map(|e| e.file_name().into_string().ok())
                .filter(|n| n.starts_with(MIGRATING_PREFIX))
                .collect()
        })
        .unwrap_or_default();
    out.sort();
    out
}

/// Classify `root/.moai` (design 6854765655 §2). Reads only.
pub fn classify(root: &Path) -> Classified {
    let moai = root.join(MOAI_DIR);
    let moai_dir = match fs::symlink_metadata(&moai) {
        Err(_) => "absent",
        Ok(m) if m.file_type().is_symlink() => "link",
        Ok(m) if m.is_dir() => "folder",
        Ok(_) => "other",
    };
    let legacy_root = regular_file(root, &[crate::PROJECT_STATE_DIR, "project.yaml"]).is_some();
    let interrupted = migrating_entries(root);
    let (go, marker, entries) = if moai_dir == "folder" {
        let entries = fs::read_dir(&moai).map(|rd| rd.count() as u64).unwrap_or(0);
        (go_markers(&moai), rust_marker(&moai), entries)
    } else {
        (Vec::new(), "absent", 0)
    };
    let (generation, reason) = if !interrupted.is_empty() {
        (
            "interrupted",
            format!(
                "a stopped migration left {}; run the recovery preview before anything else",
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
        go_markers: go,
        rust_marker: marker,
        interrupted,
        entries,
        legacy_root,
        reason,
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
            fs::create_dir(r.join(".moai.migrating-20261113T150000Z")).unwrap();
            write(r, ".moai/moai-cli.json", MARKER_V1);
        });
        assert_eq!(interrupted.generation, "interrupted");
        assert_eq!(
            interrupted.interrupted,
            vec![".moai.migrating-20261113T150000Z".to_string()]
        );
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
