//! The `ci` step of `project-update-plan` (SPEC-20261122-project-ci-design REQ-6, REQ-7): once a
//! project has saved its CI answers, the install phase of the update preview shows the CI
//! workflow files next to the `github` step, with the printed command that applies them; a
//! project without CI answers shows no such step. Every case runs the built binary in a
//! temporary project with a temporary HOME; the plan leaves the project tree as it was.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{fs, process::Output};
use tempfile::TempDir;

const CHECK: &str = ".github/workflows/moai-check.yml";
const PLACEHOLDER: &str = "<permission-reference>";

fn profile(detected: Value, with_ci: bool) -> Value {
    let mut p = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "work_flow": {"flow": "github-flow", "integration_branch": "trunk"},
        "detected": detected
    });
    if with_ci {
        p["ci"] = json!({"kind": "cli", "systems": ["linux"], "deploy": {"enabled": false},
                         "visibility": "public", "strict": "report"});
        p["checks"] = json!([
            {"id": "build", "command": "cargo build"},
            {"id": "test", "command": "cargo test"}
        ]);
    }
    p
}

/// A project whose profile is written the way PROJECT writes it: the detected facts, then
/// `project-plan` and `project-apply` (a hand-written old-layout profile would leave the backup
/// phase pending and every install step waiting).
fn project(with_ci: bool) -> TempDir {
    let root = TempDir::new().unwrap();
    fs::write(
        root.path().join("Cargo.toml"),
        "[package]\nname = \"demo\"\nversion = \"0.1.0\"\n",
    )
    .unwrap();
    save(&root, &profile(detected(&root), with_ci));
    root
}

fn detected(root: &TempDir) -> Value {
    let p = root.path().to_str().unwrap();
    envelope(&call(&["project-detect", p], b""))["detected"].clone()
}

/// Saves `profile` through `project-plan` and `project-apply`.
fn save(root: &TempDir, profile: &Value) {
    let p = root.path().to_str().unwrap();
    let body = serde_json::to_vec(profile).unwrap();
    let plan = envelope(&call(&["project-plan", p], &body));
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let applied = call(
        &[
            "project-apply",
            p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
    assert!(
        applied.status.success(),
        "{}",
        String::from_utf8_lossy(&applied.stdout)
    );
}

/// Whether the list at `pointer` of the update preview names the `ci` step.
fn names_ci(plan: &Value, pointer: &str) -> bool {
    plan.pointer(pointer)
        .and_then(Value::as_array)
        .is_some_and(|steps| steps.iter().any(|s| s["step"] == "ci"))
}

fn envelope(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "not JSON: {e}\n{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn call(args: &[&str], stdin: &[u8]) -> Output {
    use std::io::Write;
    let mut child = TempHome::new()
        .command(args)
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(stdin);
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn plan(root: &TempDir) -> Value {
    let out = call(&["project-update-plan", root.path().to_str().unwrap()], b"");
    envelope(&out)
}

/// Every path below `dir` with its bytes.
fn tree(dir: &std::path::Path) -> std::collections::BTreeMap<String, Vec<u8>> {
    fn walk(
        d: &std::path::Path,
        base: &std::path::Path,
        out: &mut std::collections::BTreeMap<String, Vec<u8>>,
    ) {
        for e in fs::read_dir(d).unwrap().flatten() {
            let path = e.path();
            let rel = path.strip_prefix(base).unwrap().display().to_string();
            if fs::symlink_metadata(&path).unwrap().is_dir() {
                walk(&path, base, out);
            } else {
                out.insert(rel, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = std::collections::BTreeMap::new();
    walk(dir, dir, &mut out);
    out
}

fn install_steps(plan: &Value) -> Vec<Value> {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .find(|p| p["phase"] == "install")
        .unwrap()["steps"]
        .as_array()
        .unwrap()
        .clone()
}

fn ci(plan: &Value) -> Option<Value> {
    install_steps(plan).into_iter().find(|s| s["step"] == "ci")
}

/// A project that never saved CI answers has no `ci` step, so the install phase reads as before.
#[test]
fn a_project_without_ci_answers_shows_no_ci_step() {
    let root = project(false);
    let p = plan(&root);
    assert!(ci(&p).is_none(), "{p}");
}

/// With CI answers the step sits right after `github`, is `ready` with the workflow file to
/// create, prints the apply command of `project-ci-apply` with the plan's digest and the
/// permission placeholder, and the plan writes nothing.
#[test]
fn a_project_with_ci_answers_shows_the_ci_step_after_github() {
    let root = project(true);
    let before = tree(root.path());
    let p = plan(&root);
    assert!(
        !root.path().join(".github").exists(),
        "the plan wrote a workflow"
    );
    assert_eq!(tree(root.path()), before, "the plan wrote something");
    let names: Vec<String> = install_steps(&p)
        .iter()
        .map(|s| s["step"].as_str().unwrap().to_string())
        .collect();
    let at = names.iter().position(|n| n == "github").unwrap();
    assert_eq!(names[at + 1], "ci", "{names:?}");

    let step = ci(&p).unwrap();
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["preview"]["files"][0]["path"], CHECK, "{step}");
    assert_eq!(step["preview"]["files"][0]["action"], "create", "{step}");
    assert!(
        step["preview"]["estimate"]["free"].as_bool().unwrap(),
        "{step}"
    );
    let argv: Vec<&str> = step["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect();
    assert_eq!(argv[0], "moai");
    assert_eq!(argv[1], "project-ci-apply");
    assert_eq!(argv[3], "--plan-digest");
    assert_eq!(Some(argv[4]), step["plan_digest"].as_str());
    assert_eq!(
        argv.iter().filter(|w| w.contains('<')).collect::<Vec<_>>(),
        vec![&PLACEHOLDER]
    );
}

/// The printed command applies: the file appears, and the next preview says `unchanged`.
#[test]
fn the_printed_command_applies_and_the_next_preview_is_unchanged() {
    let root = project(true);
    let p = plan(&root);
    let step = ci(&p).unwrap();
    let argv: Vec<String> = step["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap().replace(PLACEHOLDER, "fixture-approved"))
        .collect();
    let args: Vec<&str> = argv[1..].iter().map(String::as_str).collect();
    let out = call(&args, &serde_json::to_vec(&step["apply"]["stdin"]).unwrap());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    assert!(root.path().join(CHECK).is_file());
    let again = plan(&root);
    assert_eq!(ci(&again).unwrap()["state"], "unchanged", "{again}");
    assert!(ci(&again).unwrap()["apply"].is_null());
}

/// A workflow file the project already has by that name, with no recorded baseline, is a
/// conflict the preview names; nothing is overwritten.
#[test]
fn a_workflow_file_of_the_user_is_a_conflict_and_is_kept() {
    let root = project(true);
    let mine = root.path().join(CHECK);
    fs::create_dir_all(mine.parent().unwrap()).unwrap();
    fs::write(&mine, "name: mine\n").unwrap();
    let p = plan(&root);
    let step = ci(&p).unwrap();
    assert_eq!(step["state"], "conflict", "{step}");
    assert!(step["apply"].is_null(), "{step}");
    assert_eq!(fs::read_to_string(&mine).unwrap(), "name: mine\n");
}

/// A ready step reads its own sentence, is one of the changes the one question applies and is
/// not named there as guarded.
#[test]
fn the_ready_ci_step_says_what_it_does_and_the_question_runs_it() {
    let root = project(true);
    let p = plan(&root);
    let step = ci(&p).unwrap();
    assert_eq!(step["next"], "apply", "{step}");
    assert_eq!(
        step["say"], "Write or refresh the files that run your project's checks automatically.",
        "{step}"
    );
    let asked = p["ask"]["steps"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["step"] == "ci")
        .unwrap_or_else(|| panic!("the question does not list the step: {p}"));
    assert_eq!(asked["guarded"], false, "{asked}");
    assert!(
        p["chain"]["apply_order"]
            .as_array()
            .unwrap()
            .contains(&json!("ci")),
        "{p}"
    );
}

/// AC-5 (REQ-9): with a generated file the user edited and a project that changed since, the
/// update plans again from the saved answers, asks nothing about the step and keeps the file as
/// a conflict.
#[test]
fn an_edited_file_of_a_changed_project_stays_a_conflict_and_nothing_is_asked() {
    let root = project(true);
    let step = ci(&plan(&root)).unwrap();
    let argv: Vec<String> = step["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap().replace(PLACEHOLDER, "fixture-approved"))
        .collect();
    let args: Vec<&str> = argv[1..].iter().map(String::as_str).collect();
    envelope(&call(
        &args,
        &serde_json::to_vec(&step["apply"]["stdin"]).unwrap(),
    ));
    let file = root.path().join(CHECK);
    let edited = format!("{}# mine\n", fs::read_to_string(&file).unwrap());
    fs::write(&file, &edited).unwrap();
    // The project changed since the apply: one more declared check.
    let mut changed = profile(detected(&root), true);
    changed["checks"]
        .as_array_mut()
        .unwrap()
        .push(json!({"id": "lint", "command": "cargo clippy"}));
    save(&root, &changed);
    let before = tree(root.path());

    let p = plan(&root);
    let step = ci(&p).unwrap();
    assert_eq!(step["state"], "conflict", "{step}");
    assert_eq!(step["preview"]["files"][0]["path"], CHECK, "{step}");
    assert_eq!(step["preview"]["files"][0]["action"], "conflict", "{step}");
    assert_eq!(step["next"], "report", "{step}");
    assert!(step["apply"].is_null(), "{step}");
    assert!(step["choices"].is_null(), "{step}");
    for list in ["/ask/steps", "/ask/follows", "/ask/picks"] {
        assert!(!names_ci(&p, list), "{list} asks about the step: {p}");
    }
    assert_eq!(tree(root.path()), before, "the plan wrote something");
    assert_eq!(fs::read_to_string(&file).unwrap(), edited);
}
