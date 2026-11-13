//! #2384 AC-2 (user decision 2026-11-10, #2389 comment 6774899518 row 37): with no forge
//! connected, a ledger operation keeps the shared forge error classification and never
//! writes a local TODO or any other file. `gh` is a fake executable (or absent from
//! `PATH`); no network is used.
#![cfg(unix)]
mod common;
use common::CleanEnv;

use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

struct Fixture {
    _dir: tempfile::TempDir,
    project: PathBuf,
    home: PathBuf,
    fake_gh: PathBuf,
    calls: PathBuf,
    /// The only `PATH` entry of every run: holds `git` and nothing else, so "no gh" does not
    /// depend on what the host has installed (#2415 audit B1: Ubuntu runners ship `gh`).
    bin: PathBuf,
}

/// The absolute path of `name` on the test process's own `PATH`.
fn which(name: &str) -> PathBuf {
    std::env::var_os("PATH")
        .and_then(|paths| {
            std::env::split_paths(&paths)
                .map(|d| d.join(name))
                .find(|p| p.is_file())
        })
        .unwrap_or_else(|| panic!("{name} not found on PATH"))
}

fn fixture() -> Fixture {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let project = base.join("project");
    let home = base.join("home");
    fs::create_dir_all(&project).unwrap();
    fs::create_dir_all(&home).unwrap();
    common::git::git(&project, &["init", "-q", "-b", "main"]);
    let calls = base.join("gh-calls");
    let fake_gh = base.join("fake-gh");
    // Answers like a gh that is installed but not signed in.
    common::write_executable(
        &fake_gh,
        &format!(
            "#!/bin/sh\necho call >> '{}'\necho 'To get started with GitHub CLI, please run:  gh auth login' >&2\nexit 4\n",
            calls.display()
        ),
    );
    let bin = base.join("bin");
    fs::create_dir_all(&bin).unwrap();
    std::os::unix::fs::symlink(which("git"), bin.join("git")).unwrap();
    Fixture {
        _dir: dir,
        project,
        home,
        fake_gh,
        calls,
        bin,
    }
}

/// Every file under `dir` except the git directory: path and length.
fn files(dir: &Path) -> Vec<(String, u64)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(p) = stack.pop() {
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            if p.file_name().is_some_and(|n| n == ".git") {
                continue;
            }
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else {
            out.push((p.to_string_lossy().into_owned(), m.len()));
        }
    }
    out.sort();
    out
}

fn forge_read(fx: &Fixture, request: &str, gh: Option<&Path>) -> (i32, Value) {
    let mut cmd = Command::new(BIN);
    cmd.arg("forge-read")
        .arg(&fx.project)
        .current_dir(&fx.project)
        .clean_env()
        .env("PATH", &fx.bin)
        .env("HOME", &fx.home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    if let Some(gh) = gh {
        cmd.env("MOAI_FORGE_GH_BIN", gh);
    }
    let mut child = cmd.spawn().unwrap();
    common::feed_stdin(&mut child, request.as_bytes());
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    (out.status.code().unwrap(), v)
}

#[test]
fn forge_read_without_a_forge_creates_nothing_and_answers_transport() {
    let fx = fixture();
    let before = (files(&fx.project), files(&fx.home));
    let repo = r#"{"schema_version":1,"owner":"acme","repo":"widgets"}"#;

    // gh installed but not signed in.
    let (code, v) = forge_read(&fx, repo, Some(&fx.fake_gh));
    assert_eq!(code, 1, "{v}");
    assert_eq!(v["ok"], false);
    assert_eq!(v["error"]["kind"], "transport", "{v}");
    assert_eq!(v["error"]["retryable"], true, "{v}");
    assert_eq!(fs::read_to_string(&fx.calls).unwrap().lines().count(), 1);

    // gh absent from PATH: the fixture's only PATH folder holds git and no gh.
    let names: Vec<String> = fs::read_dir(&fx.bin)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names, ["git"]);
    let (code, v) = forge_read(&fx, repo, None);
    assert_eq!(code, 1, "{v}");
    assert_eq!(v["error"]["kind"], "transport", "{v}");
    assert_eq!(v["error"]["retryable"], true, "{v}");

    // No repository named.
    let (code, v) = forge_read(&fx, r#"{"schema_version":1}"#, Some(&fx.fake_gh));
    assert_eq!(code, 1, "{v}");
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    assert_eq!(v["error"]["retryable"], false, "{v}");
    // The request was refused before any gh call.
    assert_eq!(fs::read_to_string(&fx.calls).unwrap().lines().count(), 1);

    // No local TODO, poll state or any other file was written.
    assert_eq!((files(&fx.project), files(&fx.home)), before);
}
