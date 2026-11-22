//! The scale reading of `project-detect` (SPEC-20261122-project-scale, REQ-1 to REQ-4, AC-2):
//! small project, large single tree or monorepo, from the entry count in the header of the git
//! index and the workspace markers and manifest folders of the tree. Driven in temporary
//! folders only; no program is started and no HOME is involved.
//!
//! The kind is a fact beside `detected` and is never stored inside it.

use moai_cli::project_detect::detect_cli;
use moai_cli::project_scale_detect::{
    classify, index_entry_count, read, LARGE_TREE_FILES, MONOREPO_MANIFEST_FOLDERS,
};
use std::{fs, path::Path};
use tempfile::TempDir;

fn mk(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// The first twelve bytes of a git index: the `DIRC` signature, a version and the entry count,
/// both big-endian.
fn index_bytes(version: u32, entries: u32) -> Vec<u8> {
    let mut bytes = b"DIRC".to_vec();
    bytes.extend_from_slice(&version.to_be_bytes());
    bytes.extend_from_slice(&entries.to_be_bytes());
    bytes
}

fn with_index(root: &Path, entries: u32) {
    fs::create_dir_all(root.join(".git")).unwrap();
    fs::write(root.join(".git/index"), index_bytes(2, entries)).unwrap();
}

#[test]
fn the_thresholds_and_markers_decide_the_kind() {
    assert_eq!(LARGE_TREE_FILES, 5_000);
    assert_eq!(MONOREPO_MANIFEST_FOLDERS, 3);
    // tracked files: 4,999 is small and 5,000 is a large tree, with no marker and no package.
    assert_eq!(classify(0, 0, 0), "small");
    assert_eq!(classify(4_999, 0, 0), "small");
    assert_eq!(classify(5_000, 0, 0), "large_tree");
    assert_eq!(classify(25_494, 2, 0), "large_tree");
    // manifest folders: two stay below the line, three are a monorepo, whatever the file count.
    assert_eq!(classify(100, 2, 0), "small");
    assert_eq!(classify(100, 3, 0), "monorepo");
    assert_eq!(classify(50_000, 3, 0), "monorepo");
    // one marker is enough, and wins over a large file count.
    assert_eq!(classify(100, 0, 1), "monorepo");
    assert_eq!(classify(50_000, 0, 1), "monorepo");

    // the same lines on real folders: each marker alone.
    for (rel, text) in [
        ("pnpm-workspace.yaml", "packages:\n  - 'apps/*'\n"),
        ("lerna.json", "{}"),
        ("nx.json", "{}"),
        ("turbo.json", "{}"),
        ("rush.json", "{}"),
        ("go.work", "go 1.22\n"),
        ("package.json", r#"{"name":"r","workspaces":["apps/*"]}"#),
        ("Cargo.toml", "[workspace]\nmembers = [\"crates/*\"]\n"),
    ] {
        let dir = TempDir::new().unwrap();
        mk(dir.path(), rel, text);
        let fact = read(dir.path());
        assert_eq!(fact.kind, "monorepo", "marker {rel}");
        assert_eq!(fact.markers, [rel], "marker {rel}");
    }

    // a root package.json without `workspaces` and a Cargo.toml without `[workspace]` are not markers.
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "package.json", r#"{"name":"r","dependencies":{}}"#);
    mk(dir.path(), "Cargo.toml", "[package]\nname = \"r\"\n");
    let fact = read(dir.path());
    assert_eq!(fact.kind, "small");
    assert!(fact.markers.is_empty());

    // manifest folders on real folders: two stay small, three are a monorepo and are listed.
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "packages/a/package.json", "{}");
    mk(dir.path(), "packages/b/package.json", "{}");
    assert_eq!(read(dir.path()).kind, "small");
    mk(dir.path(), "packages/c/Cargo.toml", "[package]\nname = \"c\"\n");
    let fact = read(dir.path());
    assert_eq!(fact.kind, "monorepo");
    assert_eq!(fact.manifest_folders, 3);
    assert_eq!(fact.packages, ["packages/a", "packages/b", "packages/c"]);

    // two manifests in one folder are one folder.
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "packages/a/package.json", "{}");
    mk(dir.path(), "packages/a/tsconfig.json", "{}");
    mk(dir.path(), "packages/b/package.json", "{}");
    assert_eq!(read(dir.path()).manifest_folders, 2);

    // folders that are never part of the project do not count.
    let dir = TempDir::new().unwrap();
    for name in ["x", "y", "z"] {
        mk(dir.path(), &format!("node_modules/{name}/package.json"), "{}");
    }
    assert_eq!(read(dir.path()).kind, "small");
}

#[test]
fn the_tracked_file_count_comes_from_the_index_header() {
    assert_eq!(index_entry_count(&index_bytes(2, 4_999)), Some(4_999));
    assert_eq!(index_entry_count(&index_bytes(3, 5_000)), Some(5_000));
    assert_eq!(index_entry_count(&index_bytes(4, u32::MAX)), Some(u32::MAX as u64));
    // anything that is not a header of a known version gives no count.
    assert_eq!(index_entry_count(b""), None);
    assert_eq!(index_entry_count(&index_bytes(2, 1)[..11]), None);
    let mut bad = index_bytes(2, 9);
    bad[0] = b'X';
    assert_eq!(index_entry_count(&bad), None);
    assert_eq!(index_entry_count(&index_bytes(1, 9)), None);
    assert_eq!(index_entry_count(&index_bytes(5, 9)), None);

    for (entries, kind) in [(4_999u32, "small"), (5_000, "large_tree")] {
        let dir = TempDir::new().unwrap();
        mk(dir.path(), "src/main.rs", "fn main() {}\n");
        with_index(dir.path(), entries);
        let fact = read(dir.path());
        assert_eq!(fact.tracked_files, entries as u64);
        assert!(!fact.lower_bound);
        assert_eq!(fact.kind, kind, "{entries} files");
    }
}

#[test]
fn a_missing_or_unreadable_index_counts_the_walk_and_says_it_is_a_lower_bound() {
    // no `.git` at all
    let dir = TempDir::new().unwrap();
    for i in 0..3 {
        mk(dir.path(), &format!("src/f{i}.txt"), "x");
    }
    let fact = read(dir.path());
    assert_eq!(fact.tracked_files, 3);
    assert!(fact.lower_bound);
    assert_eq!(fact.kind, "small");

    // a `.git` folder whose index is not a header
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "a.txt", "x");
    mk(dir.path(), ".git/index", "not an index");
    let fact = read(dir.path());
    assert!(fact.lower_bound);
    assert_eq!(fact.tracked_files, 1, "the .git folder itself is not counted");

    // a `.git` folder with no index yet
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "a.txt", "x");
    fs::create_dir_all(dir.path().join(".git")).unwrap();
    assert!(read(dir.path()).lower_bound);
}

#[test]
fn a_linked_worktree_reads_its_own_index_and_a_wrong_pointer_is_not_followed() {
    let main = TempDir::new().unwrap();
    let admin = main.path().join(".git/worktrees/w");
    let wt = TempDir::new().unwrap();
    fs::create_dir_all(&admin).unwrap();
    fs::write(admin.join("index"), index_bytes(2, 7_000)).unwrap();
    fs::write(admin.join("gitdir"), format!("{}\n", wt.path().join(".git").display())).unwrap();
    fs::write(wt.path().join(".git"), format!("gitdir: {}\n", admin.display())).unwrap();
    mk(wt.path(), "a.txt", "x");
    let fact = read(wt.path());
    assert_eq!(fact.tracked_files, 7_000);
    assert!(!fact.lower_bound);
    assert_eq!(fact.kind, "large_tree");

    // the folder the pointer names does not name this folder back: not followed, so a lower bound.
    let other = TempDir::new().unwrap();
    fs::write(admin.join("gitdir"), format!("{}\n", other.path().join(".git").display())).unwrap();
    let fact = read(wt.path());
    assert!(fact.lower_bound);
    assert_eq!(fact.kind, "small");
}

#[cfg(unix)]
#[test]
fn a_git_entry_that_is_a_link_is_not_followed() {
    let real = TempDir::new().unwrap();
    with_index(real.path(), 9_000);
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "a.txt", "x");
    std::os::unix::fs::symlink(real.path().join(".git"), dir.path().join(".git")).unwrap();
    let fact = read(dir.path());
    assert!(fact.lower_bound);
    assert_eq!(fact.kind, "small");
}

#[test]
fn the_scale_is_a_fact_beside_detected_and_never_inside_it() {
    let dir = TempDir::new().unwrap();
    mk(dir.path(), "pnpm-workspace.yaml", "packages:\n  - 'apps/*'\n");
    mk(dir.path(), "apps/web/package.json", "{}");
    with_index(dir.path(), 6_000);
    let before = fs::read_to_string(dir.path().join("pnpm-workspace.yaml")).unwrap();
    let out = detect_cli(dir.path()).unwrap();
    assert_eq!(out["scale"]["kind"], "monorepo");
    assert_eq!(out["scale"]["tracked_files"], 6_000);
    assert_eq!(out["scale"]["lower_bound"], false);
    assert_eq!(out["scale"]["markers"][0], "pnpm-workspace.yaml");
    assert_eq!(out["scale"]["packages"][0], "apps/web");
    // the profile value `detected` is unchanged: no scale key inside it.
    assert!(out["detected"].get("scale").is_none());
    assert_eq!(out["profile_written"], false);
    // detection writes nothing.
    assert_eq!(fs::read_to_string(dir.path().join("pnpm-workspace.yaml")).unwrap(), before);
    assert!(!dir.path().join(".moai-cli").exists() && !dir.path().join(".moai").exists());
}
