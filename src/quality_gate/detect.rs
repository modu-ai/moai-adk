//! Language detection: the root marker plus a bounded walk below it (`detectToolchains`).

use super::table::toolchains;
use std::collections::HashSet;
use std::path::{Path, PathBuf};

/// How many levels below the project root the marker walk looks.
pub const MARKER_SCAN_DEPTH: usize = 4;

/// Dependency, build and cache folders no scan descends into (`sourceScanSkipDirs`).
pub const SKIP_DIRS: &[&str] = &[
    ".git",
    ".hg",
    ".svn",
    "node_modules",
    "vendor",
    ".venv",
    "venv",
    "site-packages",
    "__pycache__",
    ".tox",
    ".nox",
    ".mypy_cache",
    ".ruff_cache",
    ".pytest_cache",
    "dist",
    "build",
    "target",
    ".next",
    ".output",
];

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Detected {
    /// The language's position in the toolchain table.
    pub table_index: usize,
    /// The directory whose marker matched; the language's steps run here.
    pub root: PathBuf,
}

fn marker_matches(dir: &Path, marker: &str) -> bool {
    if let Some(suffix) = marker.strip_prefix('*') {
        // `*.csproj`: any file in `dir` whose name ends with the suffix.
        return std::fs::read_dir(dir)
            .map(|rd| {
                rd.flatten()
                    .any(|e| e.file_name().to_string_lossy().ends_with(suffix))
            })
            .unwrap_or(false);
    }
    dir.join(marker)
        .metadata()
        .map(|m| !m.is_dir())
        .unwrap_or(false)
}

fn match_at(dir: &Path, table: &[super::table::Toolchain], skip: &HashSet<usize>) -> Vec<usize> {
    table
        .iter()
        .enumerate()
        .filter(|(i, t)| !skip.contains(i) && t.markers.iter().any(|m| marker_matches(dir, m)))
        .map(|(i, _)| i)
        .collect()
}

/// Every language toolchain the project holds, in run order: the root match first (the first
/// table entry that matches), then the nested modules in the Go build's `WalkDir` order — depth
/// first, names sorted within a folder — one entry per language, never descending below a
/// folder whose marker matched and never deeper than [`MARKER_SCAN_DEPTH`].
pub fn detect(project_dir: &Path) -> Vec<Detected> {
    let table = toolchains();
    let mut found = Vec::new();
    let mut seen: HashSet<usize> = HashSet::new();
    if let Some(&i) = match_at(project_dir, &table, &seen).first() {
        found.push(Detected {
            table_index: i,
            root: project_dir.to_path_buf(),
        });
        seen.insert(i);
    }
    walk(project_dir, 1, &table, &mut seen, &mut found);
    found
}

fn walk(
    dir: &Path,
    depth: usize,
    table: &[super::table::Toolchain],
    seen: &mut HashSet<usize>,
    found: &mut Vec<Detected>,
) {
    if depth > MARKER_SCAN_DEPTH {
        return;
    }
    let mut subs: Vec<PathBuf> = match std::fs::read_dir(dir) {
        Ok(rd) => rd
            .flatten()
            .filter(|e| e.file_type().map(|t| t.is_dir()).unwrap_or(false))
            .filter(|e| !SKIP_DIRS.contains(&e.file_name().to_string_lossy().as_ref()))
            .map(|e| e.path())
            .collect(),
        Err(_) => return,
    };
    subs.sort();
    for sub in subs {
        let hits = match_at(&sub, table, seen);
        if hits.is_empty() {
            walk(&sub, depth + 1, table, seen, found);
            continue;
        }
        // What lives below a module root belongs to that module.
        for i in hits {
            seen.insert(i);
            found.push(Detected {
                table_index: i,
                root: sub.clone(),
            });
        }
    }
}
