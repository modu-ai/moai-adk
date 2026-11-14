//! #2936 R9 S5d: `omp-agents-plan` and `omp-agents-apply` — the project `.omp/agents` files
//! through the binary: a read-only preview, an apply that needs its digest and a permission
//! reference, a re-run that has nothing to do, the user's own agent left alone, a linked
//! `.omp` refused. Temporary folders only; the real HOME is never read or written.

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{fs, path::Path, process::Command};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const PERMISSION: &str = "fixture-approved";

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for e in fs::read_dir(from).unwrap() {
        let e = e.unwrap();
        let target = to.join(e.file_name());
        if e.file_type().unwrap().is_dir() {
            copy_tree(&e.path(), &target);
        } else {
            fs::copy(e.path(), target).unwrap();
        }
    }
}

struct Fx {
    _tmp: TempDir,
    root: std::path::PathBuf,
    project: String,
    package: String,
    agents: String,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    let (project, package, agents) = (
        root.join("project"),
        root.join("package"),
        root.join("agents"),
    );
    fs::create_dir_all(&project).unwrap();
    copy_tree(&repo.join("skills/moai"), &package);
    copy_tree(&repo.join("agents"), &agents);
    let s = |p: &Path| p.to_str().unwrap().to_string();
    Fx {
        project: s(&project),
        package: s(&package),
        agents: s(&agents),
        root,
        _tmp: tmp,
    }
}

impl Fx {
    fn run(&self, args: &[&str]) -> Value {
        let out = Command::new(EXE)
            .args(args)
            .clean_env()
            .env("HOME", self.root.join("home"))
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "{args:?}: {}{}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn plan(&self) -> Value {
        let v = self.run(&[
            "omp-agents-plan",
            &self.project,
            &self.package,
            &self.agents,
        ]);
        assert_eq!(v["ok"], json!(true), "{v}");
        v["result"].clone()
    }

    fn apply(&self, digest: &str, permission: &str) -> Value {
        self.run(&[
            "omp-agents-apply",
            &self.project,
            &self.package,
            &self.agents,
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ])
    }

    fn agent_files(&self) -> Vec<String> {
        let dir = Path::new(&self.project).join(".omp/agents");
        let mut names: Vec<String> = fs::read_dir(dir)
            .map(|d| {
                d.flatten()
                    .map(|e| e.file_name().to_string_lossy().into_owned())
                    .collect()
            })
            .unwrap_or_default();
        names.sort();
        names
    }
}

fn digest(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
}

#[test]
fn the_preview_creates_every_agent_and_writes_nothing() {
    let f = fx();
    let plan = f.plan();
    let files = plan["files"].as_array().unwrap();
    assert!(files.len() >= 5, "{plan}");
    for file in files {
        assert_eq!(file["action"], "create", "{file}");
    }
    assert_eq!(plan["preflight"], "ok", "{plan}");
    assert!(
        plan["shadows_user_agents"]
            .as_str()
            .unwrap()
            .contains("~/.omp/agent/agents"),
        "{plan}"
    );
    assert!(
        !Path::new(&f.project).join(".omp").exists(),
        "the preview created .omp"
    );
}

#[test]
fn the_apply_writes_the_planned_files_and_a_second_plan_is_unchanged() {
    let f = fx();
    let plan = f.plan();
    let out = f.apply(&digest(&plan), PERMISSION);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(out["result"]["state"], "applied", "{out}");
    let names = f.agent_files();
    assert!(names.len() >= 5, "{names:?}");
    assert!(names.iter().all(|n| n.starts_with("moai-")), "{names:?}");
    let again = f.plan();
    for file in again["files"].as_array().unwrap() {
        assert_eq!(file["action"], "unchanged", "{file}");
    }
    let second = f.apply(&digest(&again), PERMISSION);
    assert_eq!(second["result"]["state"], "unchanged", "{second}");
}

#[test]
fn a_wrong_digest_or_an_empty_permission_writes_nothing() {
    let f = fx();
    let plan = f.plan();
    let stale = f.apply(&format!("sha256:{}", "0".repeat(64)), PERMISSION);
    assert_eq!(stale["ok"], json!(false), "{stale}");
    let no_permission = f.apply(&digest(&plan), "");
    assert_eq!(no_permission["ok"], json!(false), "{no_permission}");
    assert!(
        !Path::new(&f.project).join(".omp").exists(),
        "a refused apply created .omp"
    );
}

#[test]
fn the_users_own_agent_is_never_read_or_changed() {
    let f = fx();
    let own = Path::new(&f.project).join(".omp/agents/reviewer.md");
    fs::create_dir_all(own.parent().unwrap()).unwrap();
    fs::write(&own, "my own reviewer\n").unwrap();
    let plan = f.plan();
    assert!(
        !plan.to_string().contains("reviewer.md"),
        "the plan names the user's file: {plan}"
    );
    assert_eq!(f.apply(&digest(&plan), PERMISSION)["ok"], json!(true));
    assert_eq!(fs::read_to_string(&own).unwrap(), "my own reviewer\n");
}

#[cfg(unix)]
#[test]
fn a_linked_omp_folder_is_blocked_and_nothing_lands_beyond_it() {
    let f = fx();
    let outside = f.root.join("elsewhere");
    fs::create_dir_all(&outside).unwrap();
    std::os::unix::fs::symlink(&outside, Path::new(&f.project).join(".omp")).unwrap();
    let plan = f.plan();
    assert_eq!(plan["preflight"], "blocked", "{plan}");
    let out = f.apply(&digest(&plan), PERMISSION);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(
        fs::read_dir(&outside).unwrap().count(),
        0,
        "the apply wrote through the link"
    );
}
