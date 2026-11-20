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
        let dir = tempfile::Builder::new()
            .prefix("moai-test-home-")
            .tempdir()
            .expect("create a temporary HOME");
        // REQ-4 of SPEC-20261120-test-parallelism-cap: mark the folder for macOS Spotlight
        // so a whole test run stops being indexed. Best-effort: the HOME is correct for its
        // purpose without the marker, so a failure to write it never fails the test.
        let _ = std::fs::write(dir.path().join(".metadata_never_index"), b"");
        Self { dir }
    }

    pub fn path(&self) -> &Path {
        self.dir.path()
    }

    /// A `moai` child with a clean environment whose HOME (and USERPROFILE) is this folder.
    /// Nothing of the caller's environment is passed on.
    pub fn command(&self, args: &[&str]) -> Command {
        self.command_for(super::MOAI_BIN, args)
    }

    /// Any `program` run the way [`TempHome::command`] runs `moai`: clean environment, HOME and
    /// USERPROFILE are this folder. For a ledger test that needs another process (a helper
    /// script, a signal) and must not give it the real HOME either (Issue #3068 P2).
    pub fn command_for(&self, program: impl AsRef<std::ffi::OsStr>, args: &[&str]) -> Command {
        let mut cmd = Command::new(program);
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

    /// The Rust marker `home-install-apply` of an earlier build wrote: `.moai/moai.json`,
    /// schema 1, generation `moai`. Without a Go marker beside it the HOME is `rust`, the ready
    /// state.
    pub fn make_rust_generation(&self) {
        let moai = self.path().join(".moai");
        fs::create_dir_all(&moai).unwrap();
        fs::write(
            moai.join("moai.json"),
            r#"{"schema_version":1,"generation":"moai","created_by":"moai test","project_id":"p"}"#,
        )
        .unwrap();
    }

    /// Makes `~/.moai` carry both this generation's marker and a Go marker (the Go session
    /// registry): `state_root::classify_home` answers `mixed`.
    pub fn make_mixed_generation(&self) {
        self.make_rust_generation();
        let state = self.path().join(".moai").join("state");
        fs::create_dir_all(&state).unwrap();
        fs::write(state.join("active-sessions.json"), "[]").unwrap();
    }

    /// Makes the HOME look like a Go move stopped midway: the Go folder was renamed to
    /// `~/.moai-backup` and `~/.moai` does not exist yet. `classify_home` answers `interrupted`.
    pub fn make_interrupted_generation(&self) {
        let backup = self.path().join(".moai-backup");
        fs::create_dir_all(&backup).unwrap();
        fs::write(backup.join("manifest.json"), "{}").unwrap();
    }

    /// Makes `~/.moai` a regular file, which no writer may replace or step into:
    /// `classify_home` answers `refuse`.
    pub fn make_refused_generation(&self) {
        fs::write(
            self.path().join(".moai"),
            "a file where the state folder belongs",
        )
        .unwrap();
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
