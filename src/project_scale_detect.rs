//! The scale reading of `project-detect` (SPEC-20261122-project-scale, REQ-1 to REQ-4): a small
//! project, a large single tree or a monorepo.
//!
//! The kind is a pure function of numbers the folder shows: the tracked-file count from the
//! entry count in the header of the git index (no program is started), the workspace markers of
//! the root, and the folders below the root that hold a manifest. It is a fact beside
//! `detected` and is never recorded inside it; the user's pick (later slices) overrides it.
//! The two limits are moai's own defaults, because the official documentation and the blog post
//! give no number (decision D-1 of the SPEC).
//!
//! The index of a linked worktree is read from the admin folder its `.git` pointer names, only
//! when that folder names the project's `.git` back; a `.git` that is a link is never followed.
//! When the index cannot be read, the files of a bounded walk are counted and `lower_bound`
//! says the count may be short.

use crate::project_detect::{
    is_file_in, nested_manifests, read_in, MAX_DEPTH, MAX_ENTRIES, SKIP_DIRS,
};
use serde::Serialize;
use std::{
    collections::BTreeSet,
    fs,
    io::Read,
    path::{Path, PathBuf},
};

/// Tracked files from which a tree without workspace markers reads as a large single tree.
pub const LARGE_TREE_FILES: u64 = 5_000;
/// Manifest folders below the root from which a tree reads as a monorepo.
pub const MONOREPO_MANIFEST_FOLDERS: usize = 3;
/// Packages listed at most.
const MAX_PACKAGES: usize = 64;
/// Largest `.git` pointer file read.
const POINTER_LIMIT: u64 = 4096;

/// Root files whose presence alone marks a workspace.
const MARKER_FILES: &[&str] = &[
    "go.work",
    "lerna.json",
    "nx.json",
    "pnpm-workspace.yaml",
    "rush.json",
    "turbo.json",
];

/// What the folder shows about its size and layout.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct ScaleFact {
    /// `small`, `large_tree` or `monorepo`.
    pub kind: String,
    pub tracked_files: u64,
    /// True when the count is the files of a bounded walk, not the index entries.
    pub lower_bound: bool,
    /// Folders below the root that hold a manifest.
    pub manifest_folders: usize,
    /// Workspace markers found at the root, sorted.
    pub markers: Vec<String>,
    /// The manifest folders, sorted, at most 64.
    pub packages: Vec<String>,
}

/// The kind from the three numbers: a marker or [`MONOREPO_MANIFEST_FOLDERS`] folders is a
/// monorepo, otherwise [`LARGE_TREE_FILES`] tracked files is a large tree.
pub fn classify(tracked_files: u64, manifest_folders: usize, markers: usize) -> &'static str {
    if markers > 0 || manifest_folders >= MONOREPO_MANIFEST_FOLDERS {
        "monorepo"
    } else if tracked_files >= LARGE_TREE_FILES {
        "large_tree"
    } else {
        "small"
    }
}

/// The entry count in the first twelve bytes of a git index: `DIRC`, version 2 to 4, count, all
/// big-endian. `None` for anything else.
pub fn index_entry_count(header: &[u8]) -> Option<u64> {
    if header.len() < 12 || &header[..4] != b"DIRC" {
        return None;
    }
    let version = u32::from_be_bytes(header[4..8].try_into().ok()?);
    if !(2..=4).contains(&version) {
        return None;
    }
    Some(u32::from_be_bytes(header[8..12].try_into().ok()?) as u64)
}

/// The workspace markers of the root: marker files, a `workspaces` key of the root
/// `package.json` and a `[workspace]` table of the root `Cargo.toml`.
fn markers(root: &Path) -> Vec<String> {
    let mut found: BTreeSet<String> = MARKER_FILES
        .iter()
        .filter(|name| is_file_in(root, &[**name]))
        .map(|name| name.to_string())
        .collect();
    if read_in(root, &["package.json"])
        .and_then(|text| serde_json::from_str::<serde_json::Value>(&text).ok())
        .is_some_and(|json| json.get("workspaces").is_some())
    {
        found.insert("package.json".into());
    }
    if read_in(root, &["Cargo.toml"])
        .and_then(|text| text.parse::<toml_edit::DocumentMut>().ok())
        .is_some_and(|doc| doc.get("workspace").is_some_and(|item| item.is_table()))
    {
        found.insert("Cargo.toml".into());
    }
    found.into_iter().collect()
}

/// The git index file of the project, or `None` when it cannot be trusted: no `.git`, a `.git`
/// that is a link, a pointer that is not the pointer of a linked worktree of this folder.
fn index_path(root: &Path) -> Option<PathBuf> {
    let dot_git = root.join(".git");
    let meta = fs::symlink_metadata(&dot_git).ok()?;
    let admin = if meta.is_dir() {
        dot_git.clone()
    } else if meta.is_file() && meta.len() <= POINTER_LIMIT {
        let text = fs::read_to_string(&dot_git).ok()?;
        let target = text.trim().strip_prefix("gitdir:")?.trim();
        let target = Path::new(target);
        let admin = if target.is_absolute() { target.to_path_buf() } else { root.join(target) };
        if !fs::symlink_metadata(&admin).ok()?.is_dir() {
            return None;
        }
        // The admin folder must name this project's `.git` back.
        let back = fs::read_to_string(admin.join("gitdir")).ok()?;
        let same = |a: &Path, b: &Path| match (fs::canonicalize(a), fs::canonicalize(b)) {
            (Ok(a), Ok(b)) => a == b,
            _ => a == b,
        };
        if !same(Path::new(back.trim()), &dot_git) {
            return None;
        }
        admin
    } else {
        return None;
    };
    let index = admin.join("index");
    fs::symlink_metadata(&index).ok()?.is_file().then_some(index)
}

fn tracked_from_index(root: &Path) -> Option<u64> {
    let mut header = [0u8; 12];
    fs::File::open(index_path(root)?).ok()?.read_exact(&mut header).ok()?;
    index_entry_count(&header)
}

/// The regular files of a bounded walk: [`MAX_DEPTH`] folders deep, [`MAX_ENTRIES`] entries,
/// the folders every project skips left out, links never followed.
fn walked_files(dir: &Path, depth: usize, seen: &mut usize) -> u64 {
    let Ok(entries) = fs::read_dir(dir) else {
        return 0;
    };
    let mut files = 0;
    for entry in entries.flatten() {
        *seen += 1;
        if *seen > MAX_ENTRIES {
            break;
        }
        let Ok(kind) = entry.file_type() else {
            continue;
        };
        let name = entry.file_name();
        // The repository's own entry is not a project file, whether folder or pointer file.
        if name == ".git" {
            continue;
        }
        if kind.is_dir() {
            if depth < MAX_DEPTH && !SKIP_DIRS.contains(&name.to_string_lossy().as_ref()) {
                files += walked_files(&entry.path(), depth + 1, seen);
            }
        } else if kind.is_file() {
            files += 1;
        }
    }
    files
}

/// The scale of the project folder. Reads files only.
pub fn read(root: &Path) -> ScaleFact {
    let (tracked_files, lower_bound) = match tracked_from_index(root) {
        Some(count) => (count, false),
        None => (walked_files(root, 0, &mut 0), true),
    };
    let packages: BTreeSet<String> = nested_manifests(root)
        .iter()
        .filter_map(|path| path.rsplit_once('/').map(|(folder, _)| folder.to_string()))
        .collect();
    let markers = markers(root);
    let kind = classify(tracked_files, packages.len(), markers.len()).to_string();
    ScaleFact {
        kind,
        tracked_files,
        lower_bound,
        manifest_folders: packages.len(),
        markers,
        packages: packages.into_iter().take(MAX_PACKAGES).collect(),
    }
}
