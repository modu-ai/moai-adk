//! The OMP removal, slice one (SPEC-20261120-drop-omp): `moai omp` answers with one line and
//! a non-zero exit, and no host list of the launcher offers OMP any more. Synthetic git
//! repositories in a temporary folder with a temporary HOME; the real `~/.moai` is not read
//! or written.
#![cfg(unix)]
mod common;

use common::git::git;
use serde_json::Value;
use std::fs;
use std::path::PathBuf;
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;

struct Fixture {
    _dir: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let dir = tempfile::tempdir().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let repo = base.join("repo");
        let home = base.join("home");
        fs::create_dir_all(&repo).unwrap();
        fs::create_dir_all(&home).unwrap();
        git(&repo, &["init", "-q", "-b", "main"]);
        Fixture {
            _dir: dir,
            repo,
            home,
        }
    }

    fn moai(&self, args: &[&str]) -> std::process::Output {
        Command::new(BIN)
            .args(args)
            .current_dir(&self.repo)
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap()
    }
}

/// REQ-1: `moai omp`, with or without role and name flags, prints one line that says the
/// support was removed and exits non-zero, and composes no session.
#[test]
fn the_omp_launcher_answers_one_line_and_fails() {
    let f = Fixture::new();
    for args in [
        vec!["omp"],
        vec!["omp", "-l"],
        vec!["omp", "-a"],
        vec!["omp", "-a", "--name", "lane-x"],
        vec!["omp", "--json"],
    ] {
        let out = f.moai(&args);
        assert_ne!(out.status.code(), Some(0), "{args:?} exits non-zero");
        let stdout = String::from_utf8_lossy(&out.stdout);
        let stderr = String::from_utf8_lossy(&out.stderr);
        let said = format!("{stdout}{stderr}");
        let lines = said
            .lines()
            .filter(|l| !l.trim().is_empty())
            .filter(|l| !l.starts_with('{'))
            .count();
        assert!(
            said.to_lowercase().contains("removed"),
            "{args:?}: the notice says the support was removed: {said}"
        );
        assert!(
            said.to_lowercase().contains("omp"),
            "{args:?}: the notice names omp: {said}"
        );
        assert!(lines <= 1, "{args:?}: one line, not more: {said}");
    }
}

/// REQ-2: the launcher's host list offers no OMP: the dry-run of the other hosts still works
/// and nowhere in `moai --help` or the session composition does OMP appear as a host.
#[test]
fn no_host_list_offers_omp_any_more() {
    let f = Fixture::new();
    let help = String::from_utf8_lossy(&f.moai(&["--help"]).stdout).into_owned();
    assert!(
        !help.contains("omp -l") && !help.contains("omp -a"),
        "the root help offers no omp launcher"
    );
    for line in help.lines() {
        assert!(
            !line.trim().starts_with("omp") || !line.contains("-l|-a"),
            "no omp row: {line}"
        );
    }
    // A claude dry run still names its own host only.
    let out = f.moai(&["cc", "--dry-run", "--json"]);
    assert_eq!(out.status.code(), Some(0));
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["host"], "claude");
}
