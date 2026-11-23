//! `moai bundle` refusal of a tip that is not a commit (SPEC-20261124-bundle-tip-doc-refusal): the
//! refusal of a missing `origin/<name>` also names the local `refs/heads/<name>`. Every command
//! runs against a temporary HOME and a temporary repository; the real home state and the real
//! repository are not read or written.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Stdio;

/// A head argument that parses (`C-<n>=<40 hex>`); the tip is refused before any head is read.
const HEAD_ARG: &str = "C-1=0123456789abcdef0123456789abcdef01234567";

struct Scene {
    home: TempHome,
    _root: tempfile::TempDir,
    project: String,
    work: PathBuf,
}

impl Scene {
    /// A temporary HOME, a project whose profile names the m-flow integration branch `NeXT`, and a
    /// repository with one commit on `trunk` and no `NeXT` and no `origin`.
    fn new() -> Scene {
        let home = TempHome::new();
        let project = home.project("demo");
        let root = tempfile::Builder::new()
            .prefix("moai-test-tip-refusal-")
            .tempdir()
            .unwrap();
        let work = root.path().join("work");
        fs::create_dir_all(&work).unwrap();
        let s = Scene {
            home,
            _root: root,
            project: project.to_str().unwrap().to_string(),
            work,
        };
        s.git(&["init", "-q", "-b", "trunk"]);
        fs::write(s.work.join("README.md"), "tip\n").unwrap();
        s.git(&["add", "-A"]);
        s.git(&["commit", "-q", "-m", "chore: the tip"]);
        let profile = serde_json::json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
                "source": "user", "scope": "project"}],
            "work_flow": {"flow": "m-flow", "integration_branch": "NeXT"},
        });
        let path = moai_cli::state_root::path(Path::new(&s.project), "project.yaml");
        fs::write(path, profile.to_string()).unwrap();
        s
    }

    fn git(&self, args: &[&str]) -> String {
        let mut all = vec![
            "-c",
            "user.name=Test",
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "commit.gpgsign=false",
        ];
        all.extend_from_slice(args);
        let out = self
            .home
            .command_for("git", &all)
            .current_dir(&self.work)
            .stdin(Stdio::null())
            .output()
            .expect("run git");
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// `bundle check` with the tip the scene wants (`None` leaves `--tip` out); the answer and the
    /// exit code.
    fn check(&self, tip: Option<&str>) -> (Value, i32) {
        let mut args = vec![
            "bundle",
            "check",
            HEAD_ARG,
            "--project",
            &self.project,
            "--dir",
            self.work.to_str().unwrap(),
        ];
        if let Some(t) = tip {
            args.extend_from_slice(&["--tip", t]);
        }
        let out = self
            .home
            .command(&args)
            .stdin(Stdio::null())
            .output()
            .expect("run moai");
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        let v: Value = serde_json::from_str(&line).unwrap_or_else(|e| {
            panic!(
                "bundle check: {e}: {}",
                String::from_utf8_lossy(&out.stderr)
            )
        });
        (v, out.status.code().unwrap_or(-1))
    }
}

fn message(v: &Value) -> String {
    v["error"]["message"].as_str().unwrap_or("").to_string()
}

/// AC-2: with neither `NeXT` nor `origin/NeXT` the default tip is refused and both names are given.
#[test]
fn a_missing_default_tip_refusal_names_the_local_and_the_origin_ref() {
    let s = Scene::new();
    let (v, code) = s.check(None);
    assert_eq!(code, 1, "{v}");
    let text = message(&v);
    assert!(text.contains("origin/NeXT"), "names the origin ref: {text}");
    assert!(
        text.contains("refs/heads/NeXT"),
        "names the local ref: {text}"
    );
}

/// AC-3: a tip that is not of the form `origin/<name>` is refused by its own value alone.
#[test]
fn a_tip_that_is_no_origin_ref_is_refused_by_its_own_value() {
    let s = Scene::new();
    let (v, code) = s.check(Some("no-such-branch"));
    assert_eq!(code, 1, "{v}");
    let text = message(&v);
    assert!(text.contains("no-such-branch"), "names the value: {text}");
    assert!(!text.contains("refs/heads/"), "no second name: {text}");
}

/// AC-4: when the local branch is a commit, a missing `origin/NeXT` given by `--tip` is refused
/// by that value alone.
#[test]
fn an_origin_tip_is_refused_by_its_own_value_when_the_local_branch_exists() {
    let s = Scene::new();
    s.git(&["branch", "NeXT", "trunk"]);
    let (v, code) = s.check(Some("origin/NeXT"));
    assert_eq!(code, 1, "{v}");
    let text = message(&v);
    assert!(text.contains("origin/NeXT"), "names the value: {text}");
    assert!(!text.contains("refs/heads/"), "no second name: {text}");
}
