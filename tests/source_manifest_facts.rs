//! The facts of this repository's `skills/moai/references/source-manifest.json` (Issue #2837):
//! every file under `skills/moai/` has one entry with its sha256, the
//! entries are in sorted key order with no key twice, the file is the text
//! `json.dumps(indent=2, ensure_ascii=False)` plus a newline gives, and `.gitattributes` names
//! the merge driver for it.
//!
//! The tool's own behaviour (check/write on temporary repositories, the merge driver) stays in
//! the Python module until the tool itself is a `moai` command; this file reads only the checked-in
//! tree and runs no tool, no shell and no git.
//!
//! Moved here from `source_manifest_test.py` (SPEC-20261123-rust-scans-before-ready).

use serde::de::{Deserializer, MapAccess, Visitor};
use serde::Deserialize;
use serde_json::{Map, Value};
use sha2::{Digest, Sha256};
use std::{fmt, fs, path::Path};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const PACKAGE: &str = "skills/moai";
const MANIFEST: &str = "references/source-manifest.json";
const MARKERS: [&str; 4] = ["<<<<<<< ", "||||||| ", "=======", ">>>>>>> "];

fn manifest_path() -> std::path::PathBuf {
    Path::new(ROOT).join(PACKAGE).join(MANIFEST)
}

fn manifest_text() -> String {
    fs::read_to_string(manifest_path()).expect("the skill package manifest is readable")
}

/// A JSON object read as its pairs in file order, so a key twice is seen instead of "last wins".
struct Pairs(Vec<(String, Value)>);

impl<'de> Deserialize<'de> for Pairs {
    fn deserialize<D: Deserializer<'de>>(d: D) -> Result<Self, D::Error> {
        struct V;
        impl<'de> Visitor<'de> for V {
            type Value = Pairs;
            fn expecting(&self, f: &mut fmt::Formatter) -> fmt::Result {
                f.write_str("a JSON object")
            }
            fn visit_map<A: MapAccess<'de>>(self, mut map: A) -> Result<Pairs, A::Error> {
                let mut out = Vec::new();
                while let Some(pair) = map.next_entry::<String, Value>()? {
                    out.push(pair);
                }
                Ok(Pairs(out))
            }
        }
        d.deserialize_map(V)
    }
}

#[derive(Deserialize)]
struct Doc {
    files: Pairs,
}

fn keys_of_files(text: &str) -> Vec<String> {
    let doc: Doc = serde_json::from_str(text).expect("the manifest is JSON with a `files` object");
    doc.files.0.into_iter().map(|(k, _)| k).collect()
}

fn sha256_hex(bytes: &[u8]) -> String {
    Sha256::digest(bytes)
        .iter()
        .map(|b| format!("{b:02x}"))
        .collect()
}

/// Every file under the package except the manifest, as `/`-separated paths, sorted. A link or a
/// file that is not a regular file is a failure (a FIFO would block a read).
fn package_files() -> Vec<String> {
    fn walk(dir: &Path, base: &Path, out: &mut Vec<String>) {
        let mut entries: Vec<_> = fs::read_dir(dir).unwrap().flatten().collect();
        entries.sort_by_key(|e| e.file_name());
        for entry in entries {
            let path = entry.path();
            let rel = path
                .strip_prefix(base)
                .unwrap()
                .to_string_lossy()
                .replace('\\', "/");
            let kind = fs::symlink_metadata(&path).unwrap().file_type();
            assert!(!kind.is_symlink(), "{rel} is a link");
            if kind.is_dir() {
                walk(&path, base, out);
            } else {
                assert!(kind.is_file(), "{rel} is not a regular file");
                if rel != MANIFEST {
                    out.push(rel);
                }
            }
        }
    }
    let package = Path::new(ROOT).join(PACKAGE);
    let mut out = Vec::new();
    walk(&package, &package, &mut out);
    out.sort();
    out
}

/// The folders from the repository root down to the manifest are real folders and the manifest
/// a regular file, never a link that would send a write outside the repository (#2849).
#[test]
fn the_manifest_path_has_no_link() {
    let mut here = Path::new(ROOT).to_path_buf();
    let parts: Vec<&str> = PACKAGE.split('/').chain(MANIFEST.split('/')).collect();
    for (i, part) in parts.iter().enumerate() {
        here.push(part);
        let kind = fs::symlink_metadata(&here)
            .unwrap_or_else(|e| panic!("{} is missing: {e}", here.display()))
            .file_type();
        assert!(!kind.is_symlink(), "{} is a link", here.display());
        if i + 1 == parts.len() {
            assert!(kind.is_file(), "{} is not a regular file", here.display());
        } else {
            assert!(kind.is_dir(), "{} is not a folder", here.display());
        }
    }
}

#[test]
fn the_manifest_and_the_package_hold_no_conflict_marker() {
    let marker = |line: &str| {
        MARKERS.iter().any(|m| {
            line.starts_with(m) && (*m != "=======" || line.trim_end_matches(['\r', '\n']) == *m)
        })
    };
    assert!(
        !manifest_text().lines().any(marker),
        "the manifest holds a conflict marker"
    );
    // In a skill file only the opening and closing markers count: a bare `=======` line is also
    // a Markdown heading underline.
    let package = Path::new(ROOT).join(PACKAGE);
    let unresolved: Vec<String> = package_files()
        .into_iter()
        .filter(|rel| {
            let text = String::from_utf8_lossy(&fs::read(package.join(rel)).unwrap()).into_owned();
            text.lines()
                .any(|l| l.starts_with("<<<<<<< ") || l.starts_with(">>>>>>> "))
        })
        .collect();
    assert!(
        unresolved.is_empty(),
        "conflict markers remain in: {unresolved:?}"
    );
}

#[test]
fn a_duplicate_key_is_never_read_as_last_wins() {
    // The detector itself: a text merge of two orders can leave one key twice.
    let keys = keys_of_files(r#"{"files": {"a": 1, "a": 2}}"#);
    assert_eq!(keys, ["a", "a"]);
    // This repository's manifest has each key once.
    let keys = keys_of_files(&manifest_text());
    let mut seen = std::collections::BTreeSet::new();
    for key in &keys {
        assert!(
            seen.insert(key),
            "the manifest has the key `{key}` twice; keep one entry"
        );
    }
}

#[test]
fn files_are_in_sorted_key_order() {
    // One order for every tree, whatever the history, so two branches that hold the same entry
    // hold it on the same line and a text merge never keeps an entry twice.
    let keys = keys_of_files(&manifest_text());
    let mut sorted = keys.clone();
    sorted.sort();
    assert_eq!(
        keys, sorted,
        "`files` of the manifest is not in sorted key order"
    );
}

#[test]
fn every_package_file_has_one_entry_and_every_entry_a_file() {
    let present = package_files();
    let keys = keys_of_files(&manifest_text());
    let missing: Vec<&String> = present.iter().filter(|r| !keys.contains(r)).collect();
    assert!(
        missing.is_empty(),
        "no entry (write its origin first): {missing:?}"
    );
    let dropped: Vec<&String> = keys.iter().filter(|k| !present.contains(k)).collect();
    assert!(
        dropped.is_empty(),
        "entries of files that are gone: {dropped:?}"
    );
}

/// The manifest text the tool writes: every sha256 recomputed from the file on disk, `files` in
/// sorted key order, every other key in place, two-space indent, non-ASCII kept, one newline.
fn rendered() -> String {
    let text = manifest_text();
    let data: Value = serde_json::from_str(&text).expect("the manifest is JSON");
    let package = Path::new(ROOT).join(PACKAGE);
    let mut out = Map::new();
    for (key, value) in data.as_object().expect("the manifest is an object") {
        if key != "files" {
            out.insert(key.clone(), value.clone());
            continue;
        }
        let mut files: Vec<(&String, &Value)> = value
            .as_object()
            .expect("`files` is an object")
            .iter()
            .collect();
        files.sort_by(|a, b| a.0.cmp(b.0));
        let mut sorted = Map::new();
        for (rel, entry) in files {
            let mut entry = entry.as_object().expect("an entry is an object").clone();
            let sha =
                sha256_hex(&fs::read(package.join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}")));
            entry.insert("sha256".into(), Value::String(sha));
            sorted.insert(rel.clone(), Value::Object(entry));
        }
        out.insert(key.clone(), Value::Object(sorted));
    }
    serde_json::to_string_pretty(&Value::Object(out)).unwrap() + "\n"
}

#[test]
fn this_repository_manifest_is_what_the_tool_writes() {
    let text = manifest_text();
    let want = rendered();
    if text != want {
        let line = text
            .lines()
            .zip(want.lines())
            .position(|(a, b)| a != b)
            .map_or_else(|| "its length".to_string(), |i| format!("line {}", i + 1));
        panic!("the manifest differs from what `write` produces, first at {line}: recompute the sha256 values");
    }
}

/// What `.gitattributes` says about the `merge` attribute of one path, as git reads it.
#[derive(Debug, PartialEq)]
enum Merge {
    Unspecified,
    Set,
    Unset,
    Driver(String),
}

/// `*` stays inside one folder name, `**` crosses folders, `**/` also matches no folder, and `?`
/// is one character that is not `/`.
fn glob_match(pattern: &[u8], text: &[u8]) -> bool {
    match pattern {
        [] => text.is_empty(),
        [b'*', b'*', b'/', rest @ ..] => {
            glob_match(rest, text)
                || (0..text.len()).any(|i| text[i] == b'/' && glob_match(rest, &text[i + 1..]))
        }
        [b'*', b'*', rest @ ..] => (0..=text.len()).any(|i| glob_match(rest, &text[i..])),
        [b'*', rest @ ..] => {
            let reach = text.iter().position(|c| *c == b'/').unwrap_or(text.len());
            (0..=reach).any(|i| glob_match(rest, &text[i..]))
        }
        [b'?', rest @ ..] => {
            matches!(text.first(), Some(c) if *c != b'/') && glob_match(rest, &text[1..])
        }
        [c, rest @ ..] => text.first() == Some(c) && glob_match(rest, &text[1..]),
    }
}

/// Does a `.gitattributes` pattern (kept at the repository root) name this file? A pattern with
/// no slash matches the file name in any folder; one with a slash is read from the root.
fn pattern_names(pattern: &str, path: &str) -> bool {
    if pattern.ends_with('/') {
        return false;
    }
    let anchored = pattern.trim_start_matches('/');
    if anchored.contains('/') {
        glob_match(anchored.as_bytes(), path.as_bytes())
    } else {
        let name = path.rsplit('/').next().unwrap_or(path);
        glob_match(anchored.as_bytes(), name.as_bytes())
    }
}

/// The last line that names the path and speaks of `merge` decides, as in git: `merge` sets it,
/// `-merge` unsets it, `!merge` leaves it unspecified and `merge=<driver>` names the driver. A
/// `[attr]` macro that mentions merge is refused, because what it sets cannot be told here.
fn merge_attribute(attributes: &str, path: &str) -> Result<Merge, String> {
    let mut state = Merge::Unspecified;
    for line in attributes.lines().map(str::trim) {
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        if line.starts_with("[attr]") {
            if line.contains("merge") {
                return Err(format!("a macro line speaks of merge: {line}"));
            }
            continue;
        }
        let mut words = line.split_whitespace();
        let Some(pattern) = words.next() else {
            continue;
        };
        if !pattern_names(pattern, path) {
            continue;
        }
        for word in words {
            if word == "merge" {
                state = Merge::Set;
            } else if word == "-merge" {
                state = Merge::Unset;
            } else if word == "!merge" {
                state = Merge::Unspecified;
            } else if let Some(driver) = word.strip_prefix("merge=") {
                state = Merge::Driver(driver.to_string());
            }
        }
    }
    Ok(state)
}

#[test]
fn the_attribute_reader_reads_the_last_line_that_names_the_path() {
    let path = format!("{PACKAGE}/{MANIFEST}");
    let driver = format!("{path} merge=moai-manifest\n");
    let read = |text: &str| merge_attribute(text, &path).expect("the lines are readable");
    assert_eq!(read(&driver), Merge::Driver("moai-manifest".into()));
    // Lines that do not name the file, or come before the driver line, change nothing.
    for other in [
        "skills/moai/other/*.json -merge\n",
        "*.md merge=union\n",
        "# skills/moai/references/*.json -merge\n",
        "templates/** merge=union\n",
        "skills/moai/references/ merge=union\n",
    ] {
        assert_eq!(
            read(&format!("{driver}{other}")),
            Merge::Driver("moai-manifest".into()),
            "{other}"
        );
    }
    assert_eq!(
        read(&format!("*.json -merge\n{driver}")),
        Merge::Driver("moai-manifest".into())
    );
    // A later line that names the file, by a glob, by the file name alone or by `**`, wins.
    for (later, want) in [
        ("skills/moai/references/*.json -merge", Merge::Unset),
        ("skills/**/*.json !merge", Merge::Unspecified),
        ("**/references/*.json merge", Merge::Set),
        (
            "source-manifest.json merge=ours",
            Merge::Driver("ours".into()),
        ),
        (
            "/skills/moai/references/source-manifest.json merge=union",
            Merge::Driver("union".into()),
        ),
        (
            "skills/moai/references/source-manifest.json -merge",
            Merge::Unset,
        ),
    ] {
        assert_eq!(read(&format!("{driver}{later}\n")), want, "{later}");
    }
    assert!(merge_attribute("[attr]sync merge=union\n", &path).is_err());
}

#[test]
fn this_repository_names_the_driver_for_the_manifest() {
    let path = format!("{PACKAGE}/{MANIFEST}");
    let attributes = fs::read_to_string(Path::new(ROOT).join(".gitattributes")).unwrap();
    assert_eq!(
        merge_attribute(&attributes, &path),
        Ok(Merge::Driver("moai-manifest".into())),
        ".gitattributes does not leave the manifest driver on {path}"
    );
    // A `.gitattributes` in a folder on the way down would speak after the root one.
    let mut folder = Path::new(ROOT).to_path_buf();
    for part in path.split('/').take(path.split('/').count() - 1) {
        folder.push(part);
        assert!(
            fs::symlink_metadata(folder.join(".gitattributes")).is_err(),
            "{} holds its own .gitattributes",
            folder.display()
        );
    }
}
