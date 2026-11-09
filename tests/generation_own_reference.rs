//! generation-plan targeting its own workflow location (Issue #2274). New-file discipline:
//! existing fixtures and goldens stay untouched.
//!
//! An ordinary project has no `skills/moai/workflows/`. Generation writes
//! `.moai-cli/workflows/<id>/` itself, so naming that as the `workflow_reference` must work
//! from an empty project (REQ-1), while a direct `agents-plan` with a reference that does not
//! exist is still refused (REQ-2).

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

/// An empty user project and a temporary HOME; nothing reaches the real home.
struct Fixture {
    _tmp: TempDir,
    project: std::path::PathBuf,
    home: std::path::PathBuf,
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let project = tmp.path().join("project");
    let home = tmp.path().join("home");
    fs::create_dir_all(&project).unwrap();
    fs::create_dir_all(&home).unwrap();
    Fixture {
        _tmp: tmp,
        project,
        home,
    }
}

fn cli(f: &Fixture, args: &[&str], input: &Value) -> (bool, Value) {
    let mut child = Command::new(EXE)
        .args(args)
        .env("HOME", &f.home)
        .env_remove("CODEX_HOME")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(input.to_string().as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out.status.success(), v)
}

fn p(f: &Fixture) -> &str {
    f.project.to_str().unwrap()
}

/// `<op>-plan` then `<op>-apply` with the plan's digest; returns the apply result.
fn plan_apply(f: &Fixture, op: &str, input: &Value) -> Value {
    let (ok, plan) = cli(f, &[&format!("{op}-plan"), p(f)], input);
    assert!(ok, "{op}-plan: {plan}");
    let plan = plan["result"].clone();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let (ok, applied) = cli(
        f,
        &[
            &format!("{op}-apply"),
            p(f),
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    assert!(ok, "{op}-apply: {applied}");
    applied["result"].clone()
}

fn profile() -> Value {
    json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"})
}

fn generation(reference: &str) -> Value {
    json!({
        "protocol_version": 1,
        "language": "en",
        "workflow_reference": reference,
        "local_file": "none",
        "workflow_id": "review-flow",
        "native_roles": [],
        "skill_projections": []
    })
}

fn agents(reference: &str) -> Value {
    json!({"protocol_version":1,"language":"en","workflow_reference":reference,"local_file":"none"})
}

fn agents_md(project: &Path) -> String {
    fs::read_to_string(project.join("AGENTS.md")).unwrap()
}

/// AC-1: empty project → project → generation → agents, every step exit 0, and AGENTS.md
/// points at the generated `.moai-cli/workflows/review-flow/`.
#[test]
fn an_empty_project_generates_into_its_own_workflow_location() {
    let f = fixture();
    assert!(!f.project.join("skills").exists());
    plan_apply(&f, "project", &profile());
    let reference = ".moai-cli/workflows/review-flow/";
    assert!(!f.project.join(reference).exists(), "not there before");
    plan_apply(&f, "generation", &generation(reference));
    assert!(f.project.join(reference).join("SKILL.md").is_file());
    assert!(
        agents_md(&f.project).contains(reference),
        "{}",
        agents_md(&f.project)
    );
    // agents-plan on its own now finds the generated directory.
    plan_apply(&f, "agents", &agents(reference));
    assert!(agents_md(&f.project).contains(reference));
    assert!(
        fs::read_dir(&f.home).unwrap().next().is_none(),
        "HOME untouched"
    );
}

/// The parent location `.moai-cli/workflows/` is generation's output too, with or without
/// the trailing slash.
#[test]
fn the_workflows_parent_is_also_generations_own_location() {
    for reference in [".moai-cli/workflows/", ".moai-cli/workflows"] {
        let f = fixture();
        plan_apply(&f, "project", &profile());
        plan_apply(&f, "generation", &generation(reference));
        assert!(
            agents_md(&f.project).contains(".moai-cli/workflows/"),
            "{reference}"
        );
    }
}

/// AC-2 (REQ-2): a direct agents-plan, and a generation-plan naming any other missing
/// directory, are still refused before anything is written.
#[test]
fn a_missing_reference_is_still_refused_outside_generations_own_location() {
    let f = fixture();
    plan_apply(&f, "project", &profile());
    for reference in [
        ".moai-cli/workflows/review-flow/",
        ".moai-cli/workflows/",
        "skills/moai/workflows/",
    ] {
        let (ok, v) = cli(&f, &["agents-plan", p(&f)], &agents(reference));
        assert!(!ok, "{reference}: {v}");
        assert_eq!(v["error"]["kind"], json!("not_found"), "{reference}: {v}");
        assert!(
            v["error"]["message"]
                .as_str()
                .unwrap()
                .contains("workflow reference directory missing"),
            "{v}"
        );
    }
    // A generation reference that is not its own output (another id, another tree) keeps the check.
    for reference in [".moai-cli/workflows/other-flow/", "docs/workflows/"] {
        let (ok, v) = cli(&f, &["generation-plan", p(&f)], &generation(reference));
        assert!(!ok, "{reference}: {v}");
        assert_eq!(v["error"]["kind"], json!("not_found"), "{reference}: {v}");
    }
    assert!(!f.project.join("AGENTS.md").exists());
    assert!(!f.project.join(".moai-cli/workflows").exists());
}

/// REQ-3: with generation's own location (the exception applies), the managed-marker
/// refusal still holds; a reference escaping the project is refused as before.
#[test]
fn the_exception_keeps_the_other_refusals() {
    let f = fixture();
    // A profile carrying marker syntax, written directly as the 03-C fixtures do.
    let mut marked = profile();
    marked["purpose"] = json!("Build <!-- moai:managed:end --> an editor");
    let state = f.project.join(".moai-cli");
    fs::create_dir_all(&state).unwrap();
    fs::write(
        state.join("project.yaml"),
        serde_json::to_vec_pretty(&marked).unwrap(),
    )
    .unwrap();
    let (ok, v) = cli(
        &f,
        &["generation-plan", p(&f)],
        &generation(".moai-cli/workflows/review-flow/"),
    );
    assert!(!ok, "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("managed-marker"),
        "{v}"
    );
    let (ok, v) = cli(&f, &["generation-plan", p(&f)], &generation("../outside/"));
    assert!(!ok, "{v}");
    assert!(!f.project.join("AGENTS.md").exists());
}
