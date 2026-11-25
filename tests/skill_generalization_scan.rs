//! Generalization scan of the shipped skill texts that no other change rewrites: the adapter
//! note, the git-safety, commit-trailers and loop references, the output style and the ten
//! language texts. A user project reads them, so they carry no retired host, no Issue, pull
//! request or card number, no internal tool path and no empty cell in a defaults table; the
//! manifest's origin reasons carry no retired host either.
//!
//! The texts are read from the checked-in tree; no process is started.

use std::fs;
use std::path::{Path, PathBuf};

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

/// The files this scan covers, relative to `skills/moai/references`.
fn covered() -> Vec<String> {
    let mut files: Vec<String> = [
        "adapters/orca.md",
        "internal/git-safety.md",
        "internal/commit-trailers.md",
        "internal/loop.md",
        "home/output-style.md",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect();
    let mut languages: Vec<String> = fs::read_dir(package().join("references/languages"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter(|n| n.ends_with(".md"))
        .map(|n| format!("languages/{n}"))
        .collect();
    languages.sort();
    assert_eq!(languages.len(), 10, "the language texts: {languages:?}");
    files.extend(languages);
    files
}

fn read(rel: &str) -> String {
    fs::read_to_string(package().join("references").join(rel))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The name of the state folder an earlier generation used. It is built at run time so that this
/// file does not itself spell a state folder name.
fn old_state_folder() -> String {
    format!(".{}-cli", "moai")
}

/// The words of a text, split at everything that is not a letter or a digit.
fn words(text: &str) -> Vec<&str> {
    text.split(|c: char| !c.is_ascii_alphanumeric())
        .filter(|w| !w.is_empty())
        .collect()
}

#[test]
fn no_retired_host_or_internal_tool_path_in_the_covered_texts() {
    for rel in covered() {
        let text = read(&rel);
        for banned in ["Kiro", "Cursor", "Codex", "OMP"] {
            assert!(
                !words(&text).contains(&banned),
                "{rel} names the retired host {banned}"
            );
        }
        for banned in [
            "Issue #",
            "tools/",
            old_state_folder().as_str(),
            "Go generation",
        ] {
            assert!(!text.contains(banned), "{rel} contains {banned:?}");
        }
    }
}

#[test]
fn no_issue_pull_request_or_card_number_in_the_covered_texts() {
    for rel in covered() {
        let text = read(&rel);
        let bytes = text.as_bytes();
        for i in 0..bytes.len().saturating_sub(4) {
            if bytes[i] == b'#' && bytes[i + 1..i + 5].iter().all(u8::is_ascii_digit) {
                panic!("{rel} names an Issue or pull request number at byte {i}");
            }
        }
        for i in 0..bytes.len().saturating_sub(2) {
            let starts_word = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
            if starts_word
                && bytes[i] == b'C'
                && bytes[i + 1] == b'-'
                && bytes[i + 2].is_ascii_digit()
            {
                panic!("{rel} names a card number at byte {i}");
            }
        }
        assert!(!text.contains("SPEC-2026"), "{rel} names a SPEC id");
    }
}

#[test]
fn no_table_of_the_covered_texts_has_an_empty_cell() {
    for rel in covered() {
        let text = read(&rel);
        for (n, line) in text.lines().enumerate() {
            let row = line.trim();
            if !row.starts_with('|') || !row.ends_with('|') {
                continue;
            }
            let inner = &row[1..row.len() - 1];
            let cells: Vec<&str> = inner.split('|').map(str::trim).collect();
            let separator = cells
                .iter()
                .all(|c| !c.is_empty() && c.chars().all(|ch| ch == '-' || ch == ':'));
            if separator {
                continue;
            }
            assert!(
                cells.iter().all(|c| !c.is_empty()),
                "{rel}:{}: a table row has an empty cell (write `n/a: <reason>`): {line}",
                n + 1
            );
        }
    }
}

#[test]
fn the_manifest_origin_reasons_name_no_retired_host_or_route() {
    let manifest: serde_json::Value = serde_json::from_str(
        &fs::read_to_string(package().join("references/source-manifest.json")).unwrap(),
    )
    .unwrap();
    let files = manifest["files"].as_object().expect("`files` is an object");
    for (path, entry) in files {
        let reason = entry["origin"]["reason"].as_str().unwrap_or_default();
        for banned in ["Kiro", "Cursor", "Codex", "OMP"] {
            assert!(
                !words(reason).contains(&banned),
                "{path}: the reason names the retired host {banned}"
            );
        }
        for banned in [
            old_state_folder().as_str(),
            "--mode-agent",
            "--mode-leader",
            "OPTIMIZE",
            "tools/",
        ] {
            assert!(
                !reason.contains(banned),
                "{path}: the reason contains {banned:?}"
            );
        }
    }
}
