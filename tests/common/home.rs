//! A temporary HOME for tests that touch HOME-keyed state (Issue #3068 P1, requirement 9): a
//! fresh folder, a child process whose environment holds nothing of the real one, and a tree
//! snapshot to prove what a command did or did not create. Tests of the ledger use only this;
//! none reads or writes the real `~/.moai`, `~/.claude` or `~/.codex`.

use super::CleanEnv;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use tempfile::TempDir;

/// A temporary HOME. The folder is removed when the value is dropped.
#[allow(dead_code)] // every test crate includes `common`; not every one needs a HOME
pub struct TempHome {
    dir: TempDir,
}

/// One entry of a [`TempHome::snapshot`]: path relative to the HOME, kind, size, modification
/// time (nanoseconds since the epoch; for a folder this changes whenever an entry is added or
/// removed) and the bytes of a regular file.
#[allow(dead_code)]
#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub struct Entry {
    pub path: String,
    pub kind: &'static str,
    pub len: u64,
    pub mtime_ns: u128,
    pub bytes: Vec<u8>,
}

#[allow(dead_code)]
impl TempHome {
    /// An empty HOME.
    pub fn new() -> Self {
        Self {
            dir: tempfile::Builder::new()
                .prefix("moai-test-home-")
                .tempdir()
                .expect("create a temporary HOME"),
        }
    }

    pub fn path(&self) -> &Path {
        self.dir.path()
    }

    /// A `moai` child with a clean environment whose HOME (and USERPROFILE) is this folder.
    /// Nothing of the caller's environment is passed on.
    pub fn command(&self, args: &[&str]) -> Command {
        let mut cmd = Command::new(super::MOAI_BIN);
        cmd.args(args)
            .clean_env()
            .env("HOME", self.path())
            .env("USERPROFILE", self.path());
        cmd
    }

    /// A project folder under this HOME's own scratch area with the profile a project has:
    /// `<project>/.moai-cli/project.yaml` holding `project_id`. Returns the project folder,
    /// which is not inside `~/.moai`.
    pub fn project(&self, project_id: &str) -> PathBuf {
        let project = self.path().join("work").join(project_id);
        let state = project.join(moai_cli::PROJECT_STATE_DIR);
        fs::create_dir_all(&state).unwrap();
        fs::write(
            state.join("project.yaml"),
            format!("{{\"schema_version\":2,\"project_id\":\"{project_id}\"}}\n"),
        )
        .unwrap();
        project
    }

    /// Makes `~/.moai` look like the Go generation's (a Go session registry), the HOME the
    /// ledger must not write into yet.
    pub fn make_go_generation(&self) {
        let state = self.path().join(".moai").join("state");
        fs::create_dir_all(&state).unwrap();
        fs::write(state.join("active-sessions.json"), "[]").unwrap();
    }

    /// Every entry under the HOME (symlinks not followed), sorted. Two snapshots are equal only
    /// when no entry was added, removed or changed, folder modification times included.
    pub fn snapshot(&self) -> Vec<Entry> {
        fn walk(root: &Path, dir: &Path, out: &mut Vec<Entry>) {
            for e in fs::read_dir(dir).unwrap() {
                let p = e.unwrap().path();
                let m = fs::symlink_metadata(&p).unwrap();
                let mtime_ns = m
                    .modified()
                    .ok()
                    .and_then(|t| t.duration_since(std::time::UNIX_EPOCH).ok())
                    .map_or(0, |d| d.as_nanos());
                let kind = if m.file_type().is_symlink() {
                    "link"
                } else if m.is_dir() {
                    "dir"
                } else {
                    "file"
                };
                let bytes = if kind == "file" {
                    fs::read(&p).unwrap()
                } else {
                    Vec::new()
                };
                out.push(Entry {
                    path: p.strip_prefix(root).unwrap().to_string_lossy().into_owned(),
                    kind,
                    len: m.len(),
                    mtime_ns,
                    bytes,
                });
                if kind == "dir" {
                    walk(root, &p, out);
                }
            }
        }
        let mut out = Vec::new();
        walk(self.path(), self.path(), &mut out);
        out.sort();
        out
    }

    /// The file names (not paths) directly inside `dir`, sorted.
    pub fn names_in(dir: &Path) -> Vec<String> {
        let mut names: Vec<String> = fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
            .collect();
        names.sort();
        names
    }
}

impl Default for TempHome {
    fn default() -> Self {
        Self::new()
    }
}
