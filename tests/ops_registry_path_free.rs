//! The generated operation registry names no file path (SPEC-20261122-ops-registry-path-free).
//!
//! `build.rs` writes `ops_registry.rs` into the build script's output folder. A build folder
//! that another worktree filled — shared, or warmed by an APFS clone — must not make this
//! worktree compile the other worktree's operation files, so the registry carries names and
//! help lines only, and `src/ops/mod.rs` declares the modules next to itself.

use std::{fs, path::Path, process::Command};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const REGISTRY: &str = include_str!(concat!(env!("OUT_DIR"), "/ops_registry.rs"));

/// The 1-based numbers and texts of the registry lines that hold a file path: a `#[path`
/// attribute, `src/ops/`, or the package folder `root`.
fn path_lines(text: &str, root: &str) -> Vec<(usize, String)> {
    text.lines()
        .enumerate()
        .filter(|(_, l)| {
            l.contains("#[path") || l.contains("src/ops/") || (!root.is_empty() && l.contains(root))
        })
        .map(|(i, l)| (i + 1, l.to_string()))
        .collect()
}

/// The operation file words under `src/ops/`, in file-name order.
fn op_words() -> Vec<String> {
    let mut words: Vec<String> = fs::read_dir(Path::new(ROOT).join("src/ops"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .filter_map(|n| n.strip_suffix(".rs").map(str::to_string))
        .filter(|w| w != "mod")
        .collect();
    words.sort();
    words
}

/// The words of the `with_ops!` list: an entry line starts with the file word, then the
/// variant name, then the help lines in brackets.
fn listed_words(text: &str) -> Vec<String> {
    text.lines()
        .map(str::trim)
        .filter(|l| l.ends_with(']'))
        .filter_map(|l| {
            let mut parts = l.split_whitespace();
            let word = parts.next()?;
            let variant = parts.next()?;
            let is_word = word
                .bytes()
                .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'_');
            let is_variant = variant.starts_with(|c: char| c.is_ascii_uppercase())
                && variant.bytes().all(|b| b.is_ascii_alphanumeric());
            (is_word && is_variant).then(|| word.to_string())
        })
        .collect()
}

/// The first `///` line right above `pub struct Args` in an operation file, if any.
fn first_help_line(text: &str) -> Option<String> {
    let lines: Vec<&str> = text.lines().collect();
    let at = lines
        .iter()
        .position(|l| l.trim_start().starts_with("pub struct Args"))?;
    let mut doc = Vec::new();
    for line in lines[..at].iter().rev() {
        let t = line.trim_start();
        if let Some(rest) = t.strip_prefix("///") {
            doc.push(rest.strip_prefix(' ').unwrap_or(rest).to_string());
        } else if t.starts_with("#[") {
            continue;
        } else {
            break;
        }
    }
    doc.reverse();
    doc.into_iter().next().filter(|l| !l.trim().is_empty())
}

/// AC-1: the registry is a `with_ops!` list of every operation file, in file-name order, and
/// names no path.
#[test]
fn the_registry_is_a_list_of_names_with_no_path() {
    assert!(
        REGISTRY.contains("macro_rules! with_ops"),
        "no with_ops! list in the registry:\n{REGISTRY}"
    );
    let words = op_words();
    assert!(!words.is_empty());
    assert_eq!(listed_words(REGISTRY), words);
    let found = path_lines(REGISTRY, ROOT);
    assert!(
        found.is_empty(),
        "registry lines that hold a path: {found:?}"
    );
}

/// AC-3: the path finder names each line that holds a path, and nothing in the real registry.
#[test]
fn a_path_in_the_registry_is_named() {
    let samples = [
        format!("#[path = \"{ROOT}/src/ops/alpha.rs\"]"),
        "#[path = \"/elsewhere/alpha.rs\"]".to_string(),
        "// see src/ops/alpha.rs".to_string(),
        format!("const HOME: &str = \"{ROOT}\";"),
    ];
    for sample in &samples {
        let text = format!("macro_rules! with_ops {{\n{sample}\n}}\n");
        let found = path_lines(&text, ROOT);
        assert_eq!(found, vec![(2, sample.clone())], "{sample}");
    }
    assert_eq!(path_lines("alpha Alpha [\"Alpha help\"]\n", ROOT), vec![]);
    assert_eq!(path_lines(REGISTRY, ROOT), vec![]);
}

/// AC-2: every operation answers `--help` under its kebab-case name, and the first help line
/// is the first `///` line above its `pub struct Args`.
#[test]
fn every_operation_keeps_its_name_and_help_line() {
    let exe = env!("CARGO_BIN_EXE_moai");
    for word in op_words() {
        let name = word.replace('_', "-");
        let out = Command::new(exe)
            .args([name.as_str(), "--help"])
            .env("NO_COLOR", "1")
            .output()
            .unwrap();
        let stdout = String::from_utf8_lossy(&out.stdout);
        assert!(
            out.status.success(),
            "{name} --help: {:?} {stdout} {}",
            out.status.code(),
            String::from_utf8_lossy(&out.stderr)
        );
        let text = fs::read_to_string(Path::new(ROOT).join(format!("src/ops/{word}.rs"))).unwrap();
        if let Some(line) = first_help_line(&text) {
            let first = stdout.lines().next().unwrap_or("");
            assert!(
                first.starts_with(line.trim_end()),
                "{name}: help starts {first:?}, the file says {line:?}"
            );
        }
    }
}
