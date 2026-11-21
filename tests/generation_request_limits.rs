//! SPEC-20261120-project-harness-research: `generation-plan` takes up to sixteen project agents
//! and sixteen project skills in one request, and a refusal names the field and the limit and
//! never a value of the request.
//!
//! Every case runs the built binary in a temporary project with an empty temporary HOME; the
//! real checkout and HOME are never read.

#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const LIMIT: usize = 16;

struct Fx {
    _tmp: TempDir,
    home: std::path::PathBuf,
    root: std::path::PathBuf,
}

fn fixture() -> Fx {
    let tmp = TempDir::new().unwrap();
    let base = fs::canonicalize(tmp.path()).unwrap();
    let (home, root) = (base.join("home"), base.join("project"));
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(root.join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"]
    });
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.join(".moai-cli/project.yaml"), bytes).unwrap();
    Fx {
        _tmp: tmp,
        home,
        root,
    }
}

impl Fx {
    /// The whole envelope of `generation-plan` for `request`.
    fn plan(&self, request: &Value) -> Value {
        let mut child = Command::new(EXE)
            .args(["generation-plan", self.root.to_str().unwrap()])
            .clean_env()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("LC_ALL", "C")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        let mut pipe = child.stdin.take().unwrap();
        let _ = pipe.write_all(request.to_string().as_bytes());
        drop(pipe);
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("no envelope: {}", String::from_utf8_lossy(&out.stderr)))
    }
}

fn request(roles: Vec<Value>, skills: Vec<Value>) -> Value {
    json!({
        "protocol_version": 1, "language": "en",
        "workflow_reference": "skills/moai/workflows/", "local_file": "tracked",
        "workflow_id": "review-flow", "native_roles": roles,
        "skill_projections": [], "project_skills": skills
    })
}

fn role(id: &str) -> Value {
    json!({"role_id": id, "host": "claude", "purpose": "Run the tests of the project", "tools": ["Read", "Bash"]})
}

fn skill(id: &str) -> Value {
    json!({"id": id, "purpose": "Add a subcommand with its test"})
}

fn roles(n: usize) -> Vec<Value> {
    (0..n).map(|i| role(&format!("prj-role-{i}"))).collect()
}

fn skills(n: usize) -> Vec<Value> {
    (0..n).map(|i| skill(&format!("prj-skill-{i}"))).collect()
}

fn refusal(v: &Value) -> String {
    assert_eq!(v["ok"], json!(false), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

#[test]
fn limit_sixteen_agents_and_sixteen_skills_are_planned() {
    let fx = fixture();
    let v = fx.plan(&request(roles(LIMIT), skills(LIMIT)));
    assert_eq!(v["ok"], json!(true), "{v}");
}

#[test]
fn limit_seventeen_agents_are_refused_with_the_field_and_the_limit() {
    let fx = fixture();
    let msg = refusal(&fx.plan(&request(roles(LIMIT + 1), vec![])));
    assert!(msg.contains("native_roles"), "{msg}");
    assert!(msg.contains("16"), "{msg}");
    assert!(!msg.contains("prj-role-"), "no value of the request: {msg}");
}

#[test]
fn limit_seventeen_skills_are_refused_with_the_field_and_the_limit() {
    let fx = fixture();
    let msg = refusal(&fx.plan(&request(vec![], skills(LIMIT + 1))));
    assert!(msg.contains("project_skills"), "{msg}");
    assert!(msg.contains("16"), "{msg}");
    assert!(
        !msg.contains("prj-skill-"),
        "no value of the request: {msg}"
    );
}

#[test]
fn unsafe_a_name_without_the_prefix_is_refused_with_its_field_and_index() {
    let fx = fixture();
    let msg = refusal(&fx.plan(&request(
        vec![role("prj-fine"), role("shadow-agent")],
        vec![],
    )));
    assert!(msg.contains("native_roles[1].role_id"), "{msg}");
    assert!(
        !msg.contains("shadow-agent"),
        "no value of the request: {msg}"
    );
    let msg = refusal(&fx.plan(&request(vec![], vec![skill("prj-fine"), skill("../x")])));
    assert!(msg.contains("project_skills[1].id"), "{msg}");
    assert!(!msg.contains("../x"), "no value of the request: {msg}");
}

#[test]
fn unsafe_a_host_a_purpose_and_a_tool_name_the_field() {
    let fx = fixture();
    let mut bad = role("prj-a");
    bad["host"] = json!("vim");
    let msg = refusal(&fx.plan(&request(vec![bad], vec![])));
    assert!(msg.contains("native_roles[0].host"), "{msg}");
    let mut bad = role("prj-a");
    bad["tools"] = json!(["Read", "Bash tool"]);
    let msg = refusal(&fx.plan(&request(vec![bad], vec![])));
    assert!(msg.contains("native_roles[0].tools"), "{msg}");
    let msg = refusal(&fx.plan(&request(
        vec![],
        vec![json!({"id": "prj-a", "purpose": ""})],
    )));
    assert!(msg.contains("project_skills[0].purpose"), "{msg}");
}
