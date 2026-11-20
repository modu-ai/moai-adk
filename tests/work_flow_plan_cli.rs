//! SPEC-20261120-mflow-integration-branch-choice: `moai work-flow-plan` — the first-setup
//! question evidence. It prints the branch candidates (a pull-request candidate ref is
//! never canonical), one recommended branch with its reason (the recorded profile value on
//! a re-run, else the remote default, else the current branch), the recorded work-flow
//! state, and the safe-choice wording when the recorded target is missing or occupied by
//! another worktree. It creates nothing: every run leaves the branch list unchanged.
//! Temporary HOME, temporary repository, bare remote; never the network or the real HOME.
#![cfg(unix)]

mod common;

use common::MOAI_BIN;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicUsize, Ordering};

static NEXT: AtomicUsize = AtomicUsize::new(0);

fn git(dir: &Path, args: &[&str]) {
    let ok = common::git::command()
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .expect("git");
    assert!(
        ok.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&ok.stderr)
    );
}

fn commit(dir: &Path, message: &str) {
    let ok = common::git::command()
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "commit",
            "-q",
            "-m",
            message,
        ])
        .output()
        .unwrap();
    assert!(ok.status.success());
}

struct World {
    home: PathBuf,
    project: PathBuf,
}

impl World {
    /// A project repository on `main` with a bare `origin` and a HOME for the run.
    fn new() -> Self {
        let n = NEXT.fetch_add(1, Ordering::SeqCst);
        let root = std::env::temp_dir().join(format!("wfplan-{}-{n}", std::process::id()));
        fs::create_dir_all(&root).unwrap();
        let home = root.join("home");
        let project = root.join("demo");
        let remote = root.join("remote.git");
        fs::create_dir_all(&home).unwrap();
        let w = World {
            home,
            project: project.clone(),
        };
        let ok = common::git::command()
            .arg("init")
            .arg("-q")
            .arg("-b")
            .arg("main")
            .arg(&project)
            .output()
            .expect("git init");
        assert!(ok.status.success(), "git init");
        let ok = common::git::command()
            .arg("init")
            .arg("-q")
            .arg("--bare")
            .arg("-b")
            .arg("main")
            .arg(&remote)
            .output()
            .expect("git init --bare");
        assert!(ok.status.success(), "git init --bare");
        git(
            &project,
            &["remote", "add", "origin", remote.to_str().unwrap()],
        );
        fs::write(project.join("seed.txt"), "base\n").unwrap();
        git(&project, &["add", "-A"]);
        commit(&project, "base");
        git(&project, &["push", "-q", "origin", "main"]);
        w
    }

    /// `moai work-flow-plan --project <project>`; the answer's `result`.
    fn plan(&self) -> Value {
        let mut cmd = Command::new(MOAI_BIN);
        cmd.arg("work-flow-plan")
            .arg("--project")
            .arg(self.project.to_str().unwrap())
            .env("HOME", &self.home)
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        let out = cmd.output().expect("run moai");
        let line = String::from_utf8_lossy(&out.stdout)
            .lines()
            .next()
            .unwrap_or("")
            .to_string();
        let v: Value = serde_json::from_str(&line).unwrap_or_else(|e| {
            panic!(
                "work-flow-plan: {e}: stdout {} stderr {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        });
        assert_eq!(v["ok"], true, "work-flow-plan: {v}");
        v["result"].clone()
    }

    fn write_profile(&self, work_flow: Option<Value>) {
        let mut profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
                "source": "user", "scope": "project"}],
        });
        if let Some(wf) = work_flow {
            profile["work_flow"] = wf;
        }
        let path = self.project.join(".moai-cli").join("project.yaml");
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, profile.to_string()).unwrap();
    }

    /// `origin/HEAD` points at `origin/main` (the local symbolic ref, no network).
    fn set_origin_head(&self) {
        git(
            &self.project,
            &[
                "symbolic-ref",
                "refs/remotes/origin/HEAD",
                "refs/remotes/origin/main",
            ],
        );
    }

    fn branch(&self, name: &str) {
        git(&self.project, &["branch", name]);
    }

    fn branches(&self) -> Vec<String> {
        let out = common::git::command()
            .arg("-C")
            .arg(&self.project)
            .args(["for-each-ref", "refs/heads/", "--format=%(refname:short)"])
            .output()
            .expect("git for-each-ref");
        String::from_utf8_lossy(&out.stdout)
            .lines()
            .map(str::to_string)
            .collect()
    }
}

/// AC-1 (first run): with no recorded value and no `origin/HEAD`, the recommendation is
/// the current branch with the current-branch reason, and the local branches are canonical.
#[test]
fn the_first_run_recommends_the_current_branch() {
    let w = World::new();
    let v = w.plan();
    assert_eq!(v["recommendation"]["branch"], "main");
    let reason = v["recommendation"]["reason"].as_str().unwrap();
    assert!(reason.contains("current branch"), "reason: {reason}");
    let cands = v["candidates"].as_array().unwrap();
    let main = cands
        .iter()
        .find(|c| c["name"] == "main")
        .expect("main is a candidate");
    assert_eq!(main["canonical"], true);
}

/// AC-1 (discovery rule): with `origin/HEAD` set and nothing recorded, the recommendation
/// is the remote's default branch with that reason.
#[test]
fn the_remote_default_branch_is_recommended_before_the_current_one() {
    let w = World::new();
    w.set_origin_head();
    let v = w.plan();
    assert_eq!(v["recommendation"]["branch"], "main");
    let reason = v["recommendation"]["reason"].as_str().unwrap();
    assert!(reason.contains("default branch"), "reason: {reason}");
}

/// AC-1 (re-run): a recorded `work_flow.integration_branch` wins the recommendation with
/// the recorded-answer reason, so the second run proposes the saved choice without asking.
#[test]
fn a_rerun_recommends_the_recorded_branch() {
    let w = World::new();
    w.branch("v2026.11.0");
    w.write_profile(Some(
        json!({"flow": "m-flow", "integration_branch": "v2026.11.0"}),
    ));
    let v = w.plan();
    assert_eq!(v["recommendation"]["branch"], "v2026.11.0");
    let reason = v["recommendation"]["reason"].as_str().unwrap();
    assert!(reason.contains("recorded"), "reason: {reason}");
}

/// AC-3: a pull-request candidate ref prints with `canonical: false` and a candidate
/// label, and is never recommended as the integration target.
#[test]
fn a_pr_candidate_ref_is_never_canonical() {
    let w = World::new();
    git(
        &w.project,
        &["update-ref", "refs/remotes/origin/pr/42", "HEAD"],
    );
    let v = w.plan();
    let pr = v["candidates"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["name"] == "origin/pr/42")
        .expect("the pr ref is a candidate");
    assert_eq!(pr["canonical"], false, "pr ref: {pr}");
    let label = pr["label"].as_str().unwrap_or("");
    assert!(label.contains("candidate"), "label: {label}");
    assert_ne!(v["recommendation"]["branch"], "origin/pr/42");
}

/// AC-2: a recorded target the repository does not have reports `missing-target` with the
/// safe choices, and the run creates no branch.
#[test]
fn a_missing_recorded_target_is_reported() {
    let w = World::new();
    w.write_profile(Some(
        json!({"flow": "m-flow", "integration_branch": "ghost"}),
    ));
    let before = w.branches();
    let v = w.plan();
    assert_eq!(v["state"]["kind"], "missing-target");
    let msg = v["state"]["message"].as_str().unwrap();
    assert!(msg.contains("existing branch"), "message: {msg}");
    assert!(msg.contains("creates nothing"), "message: {msg}");
    assert_eq!(w.branches(), before);
}

/// AC-2: a worktree that already has the recorded branch checked out reports
/// `occupied-worktree` naming that worktree, and the run creates no branch.
#[test]
fn an_occupied_worktree_is_reported() {
    let w = World::new();
    w.branch("v2026.11.0");
    let other = w.project.parent().unwrap().join("demo-occupied");
    git(
        &w.project,
        &["worktree", "add", other.to_str().unwrap(), "v2026.11.0"],
    );
    w.write_profile(Some(
        json!({"flow": "m-flow", "integration_branch": "v2026.11.0"}),
    ));
    let before = w.branches();
    let v = w.plan();
    assert_eq!(v["state"]["kind"], "occupied-worktree");
    let wt = v["state"]["worktree"].as_str().unwrap();
    assert!(wt.ends_with("demo-occupied"), "worktree: {wt}");
    let msg = v["state"]["message"].as_str().unwrap();
    assert!(msg.contains("creates nothing"), "message: {msg}");
    assert_eq!(w.branches(), before);
}
