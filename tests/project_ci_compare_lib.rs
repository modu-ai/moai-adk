//! AC-2 of SPEC-20261122-project-ci-design (REQ-4, REQ-5): the commands in the planned
//! workflow's run steps, the profile's declared commands and the commands the project's local
//! check runs are one list in one order. The test reads all three and compares them; a
//! workflow that drifted from the profile is a finding.
//!
//! The workflow is published through the binary with a temporary HOME of its own: an apply
//! keeps a receipt in the HOME folder, so a library call here would reach the real one. The
//! reading and comparing of the lists is the library's.

mod common;
use common::home::TempHome;
use moai_cli::project_check::{compare, declared, run, workflow_run_steps, Options};
use serde_json::{json, Value};
use std::{fs, io::Write, path::Path, process::Stdio};
use tempfile::TempDir;

const CHECK: &str = ".github/workflows/moai-check.yml";

fn project(flow: &str, systems: Value, commands: &[(&str, &str)]) -> TempDir {
    let root = TempDir::new().unwrap();
    let checks: Vec<Value> = commands
        .iter()
        .map(|(id, command)| json!({"id": id, "command": command}))
        .collect();
    let mut ci = json!({
        "kind": "cli", "systems": systems, "deploy": {"enabled": false},
        "visibility": "public", "strict": "report"
    });
    if ci["systems"]
        .as_array()
        .unwrap()
        .iter()
        .any(|s| s == "macos")
    {
        ci["macos_price_ack"] = json!(true);
    }
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "work_flow": {"flow": flow, "integration_branch": "trunk"},
        "detected": {"project_kind": "existing", "manifests": ["Cargo.toml"],
                     "languages": ["rust"], "lsp": [], "mcp_servers": [], "databases": []},
        "ci": ci, "checks": checks
    });
    let path = root
        .path()
        .join(moai_cli::state_root::rel(root.path(), "project.yaml"));
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, serde_json::to_vec_pretty(&profile).unwrap()).unwrap();
    root
}

/// The result of one `moai` operation run in `home` with `stdin` as its input.
fn answer(home: &TempHome, args: &[&str], stdin: &Value) -> Value {
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// Plans and applies the CI files of `root` and returns the check workflow's text.
fn publish(root: &Path) -> String {
    let home = TempHome::new();
    let project = root.to_str().unwrap();
    let plan = answer(
        &home,
        &["project-ci-plan", project],
        &json!({"protocol_version": 1}),
    );
    answer(
        &home,
        &[
            "project-ci-apply",
            project,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    fs::read_to_string(root.join(CHECK)).unwrap()
}

const COMMANDS: [(&str, &str); 4] = [
    ("build", "echo build >> ran.txt"),
    ("test", "echo test >> ran.txt"),
    ("lint", "echo lint >> ran.txt"),
    ("format", "echo format >> ran.txt"),
];

/// The three lists are equal for every flow the project can choose and with extra systems.
#[test]
fn the_workflow_the_profile_and_the_local_check_are_one_list_in_one_order() {
    let want: Vec<String> = COMMANDS.iter().map(|(_, c)| c.to_string()).collect();
    for flow in [
        "next-flow",
        "m-flow",
        "github-flow",
        "trunk-based",
        "git-flow",
    ] {
        for systems in [json!(["linux"]), json!(["linux", "macos", "windows"])] {
            let root = project(flow, systems.clone(), &COMMANDS);
            let text = publish(root.path());

            let in_workflow = workflow_run_steps(&text);
            let in_profile: Vec<String> = declared(root.path())
                .unwrap()
                .into_iter()
                .map(|c| c.command)
                .collect();
            let report = run(root.path(), &Options { timeout_secs: 20 }).unwrap();
            let in_local: Vec<String> = report["steps"]
                .as_array()
                .unwrap()
                .iter()
                .map(|s| s["command"].as_str().unwrap().to_string())
                .collect();
            let ran = fs::read_to_string(root.path().join("ran.txt")).unwrap();

            assert_eq!(in_profile, want, "{flow} {systems}");
            assert_eq!(in_workflow, want, "{flow} {systems}\n{text}");
            assert_eq!(in_local, want, "{flow} {systems}");
            assert_eq!(ran, "build\ntest\nlint\nformat\n", "{flow} {systems}");
            assert_eq!(
                compare(&declared(root.path()).unwrap(), &text),
                Vec::<String>::new(),
                "{flow} {systems}"
            );
        }
    }
}

/// A workflow that no longer matches the profile is a finding: a command declared since, a
/// command removed since and a change of order are each named.
#[test]
fn a_workflow_that_drifted_from_the_profile_is_a_finding() {
    let root = project("github-flow", json!(["linux"]), &COMMANDS);
    let text = publish(root.path());
    let drifted = [
        (
            "removed from the profile",
            vec![
                ("build", "echo build >> ran.txt"),
                ("test", "echo test >> ran.txt"),
            ],
        ),
        (
            "added to the profile",
            vec![
                ("build", "echo build >> ran.txt"),
                ("test", "echo test >> ran.txt"),
                ("lint", "echo lint >> ran.txt"),
                ("format", "echo format >> ran.txt"),
                ("docs", "echo docs >> ran.txt"),
            ],
        ),
        (
            "reordered in the profile",
            vec![
                ("test", "echo test >> ran.txt"),
                ("build", "echo build >> ran.txt"),
                ("lint", "echo lint >> ran.txt"),
                ("format", "echo format >> ran.txt"),
            ],
        ),
    ];
    for (why, list) in drifted {
        let other = project("github-flow", json!(["linux"]), &list);
        let findings = compare(&declared(other.path()).unwrap(), &text);
        assert!(!findings.is_empty(), "{why}: no finding");
    }
}
