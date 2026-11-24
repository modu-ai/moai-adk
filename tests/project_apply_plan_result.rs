//! SPEC-20261124-moai-project-interview (REQ-11, REQ-12): `moai project-apply` reads the `result`
//! object of `moai project-plan` exactly as printed. A plan of a project whose profile names an
//! integration branch carries `work_flow_remote` (the remote check), and the apply used to refuse
//! that key as `invalid typed JSON input`. Temporary HOME, temporary repository and bare remote;
//! never the network or the real HOME.

mod common;

use common::MOAI_BIN;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

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

/// A project repository on `main` with a bare remote `origin` that already holds `main`, and a
/// HOME for the run.
struct World {
    _dir: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let dir = tempfile::tempdir().unwrap();
        let root = dir.path().to_path_buf();
        let home = root.join("home");
        let project = root.join("demo");
        let remote = root.join("remote.git");
        fs::create_dir_all(&home).unwrap();
        let init = common::git::command()
            .arg("init")
            .arg("-q")
            .arg("-b")
            .arg("main")
            .arg(&project)
            .output()
            .expect("git init");
        assert!(init.status.success(), "git init");
        let bare = common::git::command()
            .arg("init")
            .arg("-q")
            .arg("--bare")
            .arg("-b")
            .arg("main")
            .arg(&remote)
            .output()
            .expect("git init --bare");
        assert!(bare.status.success(), "git init --bare");
        git(
            &project,
            &["remote", "add", "origin", remote.to_str().unwrap()],
        );
        fs::write(project.join("seed.txt"), "base\n").unwrap();
        git(&project, &["add", "--", "seed.txt"]);
        git(
            &project,
            &[
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "commit.gpgsign=false",
                "commit",
                "-q",
                "-m",
                "base",
            ],
        );
        git(&project, &["push", "-q", "origin", "main"]);
        World {
            _dir: dir,
            home,
            project,
        }
    }

    fn run(&self, args: &[&str], stdin: &[u8]) -> Output {
        let mut child = Command::new(MOAI_BIN)
            .args(args)
            .env("HOME", &self.home)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        child.stdin.take().unwrap().write_all(stdin).unwrap();
        child.wait_with_output().unwrap()
    }

    /// The `project-plan` answer of a profile that names `main` as its integration branch.
    fn plan(&self) -> Value {
        let profile = json!({
            "schema_version": 2, "project_id": "demo", "purpose": "Fixture",
            "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
            "decisions": [{"id": "d1", "topic": "tools", "value": "Rust",
                "source": "user", "scope": "project"}],
            "work_flow": {"flow": "github-flow", "integration_branch": "main"},
        });
        let out = self.run(
            &["project-plan", self.project.to_str().unwrap()],
            profile.to_string().as_bytes(),
        );
        assert!(
            out.status.success(),
            "plan failed: {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
        envelope["result"].clone()
    }

    fn apply(&self, plan: &Value) -> Output {
        self.run(
            &[
                "project-apply",
                self.project.to_str().unwrap(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                "fixture-approved",
            ],
            plan.to_string().as_bytes(),
        )
    }
}

fn said(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// AC-8 (REQ-11): the plan of a project with a remote carries `work_flow_remote`, and the apply
/// takes that `result` exactly as printed and writes the profile.
#[test]
fn apply_takes_the_plan_result_that_carries_work_flow_remote() {
    let world = World::new();
    let plan = world.plan();
    assert!(
        plan.get("work_flow_remote").is_some(),
        "the fixture must produce a plan with `work_flow_remote`: {plan}"
    );
    let out = world.apply(&plan);
    assert!(
        out.status.success(),
        "apply refused the plan result as printed: {}",
        said(&out)
    );
    let profile = fs::read_to_string(moai_cli::state_root::path(&world.project, "project.yaml"))
        .expect("the apply must write the profile");
    assert!(profile.contains("Fixture"), "{profile}");
}

/// AC-8 (REQ-12): any other unknown key is still refused as invalid typed JSON input, and
/// nothing is written.
#[test]
fn apply_still_refuses_another_unknown_key() {
    let world = World::new();
    let mut plan = world.plan();
    // Take the one tolerated key out first, so the refusal can only come from the other key.
    plan.as_object_mut().unwrap().remove("work_flow_remote");
    plan["not_a_plan_key"] = json!(true);
    let out = world.apply(&plan);
    assert!(
        !out.status.success(),
        "an unknown key must be refused: {}",
        said(&out)
    );
    assert!(
        said(&out).contains("invalid typed JSON input"),
        "{}",
        said(&out)
    );
    assert!(
        !moai_cli::state_root::path(&world.project, "project.yaml").exists(),
        "a refused apply must write nothing"
    );
}

/// AC-8 (REQ-11): a plan without `work_flow_remote` still applies, so the fix does not make the
/// key mandatory.
#[test]
fn apply_takes_a_plan_result_without_work_flow_remote() {
    let world = World::new();
    let mut plan = world.plan();
    plan.as_object_mut().unwrap().remove("work_flow_remote");
    let out = world.apply(&plan);
    assert!(out.status.success(), "{}", said(&out));
}
