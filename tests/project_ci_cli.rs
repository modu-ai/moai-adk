//! The two operations of the CI files (SPEC-20261122-project-ci-design, REQ-6 to REQ-8):
//! `project-ci-plan` previews, `project-ci-apply` publishes exactly the plan the user picked.
//! The library is tested in `project_ci_plan_lib.rs`; here the binary is driven in temporary
//! folders with a temporary HOME of its own, so no test reads or writes the real one.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Output, Stdio},
};
use tempfile::TempDir;

const CHECK: &str = ".github/workflows/moai-check.yml";
const NIGHTLY: &str = ".github/workflows/moai-nightly.yml";

fn run(args: &[&str], stdin: &Value) -> Output {
    run_with_env(args, stdin, &[])
}

/// [`run`] with `env` added to the child's clean environment.
fn run_with_env(args: &[&str], stdin: &Value, env: &[(&str, &str)]) -> Output {
    let home = TempHome::new();
    let mut child = home
        .command(args)
        .envs(env.iter().copied())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(stdin.to_string().as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    v
}

fn profile(flow: &str) -> Value {
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "work_flow": {"flow": flow, "integration_branch": "trunk"},
        "detected": {"project_kind": "existing", "manifests": ["Cargo.toml"],
                     "languages": ["rust"], "lsp": [], "mcp_servers": [], "databases": []},
        "ci": {"kind": "cli", "systems": ["linux"], "deploy": {"enabled": false},
               "visibility": "public", "strict": "report"},
        "checks": [
            {"id": "build", "command": "cargo build"},
            {"id": "test", "command": "cargo test"}
        ]
    })
}

fn project(flow: &str) -> TempDir {
    let root = TempDir::new().unwrap();
    set_profile(root.path(), flow);
    root
}

fn set_profile(root: &Path, flow: &str) {
    save(root, &profile(flow));
}

fn save(root: &Path, profile: &Value) {
    let path = root.join(moai_cli::state_root::rel(root, "project.yaml"));
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, serde_json::to_vec_pretty(profile).unwrap()).unwrap();
}

/// The files a fresh project with this profile would get, as (path, digest of the text).
fn planned_text(profile: &Value) -> Vec<(String, String)> {
    let root = TempDir::new().unwrap();
    save(root.path(), profile);
    plan(&root, json!({"protocol_version": 1}))["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["rendered_sha256"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn plan(root: &TempDir, request: Value) -> Value {
    ok(&run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &request,
    ))
}

fn apply(root: &TempDir, plan: &Value, digest: &str, permission: &str) -> Output {
    run(
        &[
            "project-ci-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ],
        plan,
    )
}

fn digest(plan: &Value) -> String {
    plan["plan_digest"].as_str().unwrap().to_string()
}

fn actions(plan: &Value) -> Vec<(String, String)> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// REQ-6: the plan prints the files, the estimate and the digest, and writes nothing.
#[test]
fn the_plan_previews_and_writes_nothing() {
    let root = project("m-flow");
    let before = fs::read_dir(root.path()).unwrap().count();
    let out = run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &json!({"protocol_version": 1}),
    );
    let v = envelope(&out);
    assert_eq!(v["operation"], "project.ci_plan", "{v}");
    let p = ok(&out);
    assert_eq!(
        actions(&p),
        vec![
            (CHECK.to_string(), "create".to_string()),
            (NIGHTLY.to_string(), "create".to_string())
        ],
        "{p}"
    );
    assert!(p["estimate"]["free"].as_bool().unwrap(), "{p}");
    assert!(digest(&p).starts_with("sha256:"), "{p}");
    assert!(!root.path().join(".github").exists());
    assert_eq!(fs::read_dir(root.path()).unwrap().count(), before);
}

/// REQ-7: the pick is the digest plus a permission reference; the files exist only after it,
/// and a second apply changes nothing.
#[test]
fn the_pick_writes_exactly_the_plan_and_a_second_apply_changes_nothing() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    let out = apply(&root, &p, &digest(&p), "fixture-approved");
    let v = envelope(&out);
    assert_eq!(v["operation"], "project.ci_apply", "{v}");
    let done = ok(&out);
    assert_eq!(done["written"], json!([CHECK, NIGHTLY]), "{done}");
    assert!(root.path().join(CHECK).is_file());
    assert!(fs::read_to_string(root.path().join(CHECK))
        .unwrap()
        .contains("persist-credentials: false"));

    let again = plan(&root, json!({"protocol_version": 1}));
    assert_eq!(
        actions(&again),
        vec![
            (CHECK.to_string(), "unchanged".to_string()),
            (NIGHTLY.to_string(), "unchanged".to_string())
        ]
    );
    let second = ok(&apply(&root, &again, &digest(&again), "fixture-approved"));
    assert_eq!(second["written"], json!([]), "{second}");
    assert!(second["request_id"].is_null(), "{second}");
}

/// REQ-7: no permission reference, a wrong digest, a plan changed by hand and a plan the
/// profile has since outgrown each write nothing.
#[test]
fn a_missing_wrong_or_stale_pick_writes_nothing() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    for (d, permission) in [
        (digest(&p), ""),
        (digest(&p), "null"),
        ("sha256:not-the-digest".to_string(), "fixture-approved"),
    ] {
        refused(&apply(&root, &p, &d, permission));
    }
    let mut altered = p.clone();
    altered["files"][0]["rendered_sha256"] = json!("sha256:00");
    refused(&apply(&root, &altered, &digest(&p), "fixture-approved"));
    set_profile(root.path(), "github-flow");
    refused(&apply(&root, &p, &digest(&p), "fixture-approved"));
    assert!(!root.path().join(".github").exists());
}

/// A request that is not the typed shape is refused by the plan with the operation named.
#[test]
fn a_wrong_request_is_refused_with_its_operation_named() {
    let root = project("m-flow");
    for request in [json!({}), json!({"protocol_version": 2}), json!([])] {
        let v = refused(&run(
            &["project-ci-plan", root.path().to_str().unwrap()],
            &request,
        ));
        assert_eq!(v["operation"], "project.ci_plan", "{v}");
    }
}

/// B2-N2 through the binary: the removal of a file the plan stopped owning is a pick in the
/// request, previewed as `remove`, applied with the digest.
#[test]
fn a_file_the_plan_stopped_owning_is_removed_only_when_picked() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    ok(&apply(&root, &p, &digest(&p), "fixture-approved"));
    set_profile(root.path(), "github-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    assert!(actions(&p).contains(&(NIGHTLY.to_string(), "stale".to_string())));
    let picked = plan(&root, json!({"protocol_version": 1, "remove": [NIGHTLY]}));
    assert!(actions(&picked).contains(&(NIGHTLY.to_string(), "remove".to_string())));
    assert!(root.path().join(NIGHTLY).exists(), "the plan removed it");
    let done = ok(&apply(&root, &picked, &digest(&picked), "fixture-approved"));
    assert_eq!(done["removed"], json!([NIGHTLY]), "{done}");
    assert!(!root.path().join(NIGHTLY).exists());
    // A pick for a path the plan does not own is refused.
    refused(&run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &json!({"protocol_version": 1, "remove": [".github/workflows/ci.yml"]}),
    ));
}

/// AC-1 (REQ-3), the answers that reach a workflow file: for each of the four profiles a
/// change of the confirmed checks, of the operating systems or of the deploy answer changes the
/// planned text.
// @cc:later AC-1 asks the same of the kind, money and strictness answers; the generator writes
// none of them into a file yet, so they are not covered here.
#[test]
fn a_change_of_checks_systems_or_deploy_changes_the_planned_text_for_four_profiles() {
    for kind in ["web", "api", "cli", "library"] {
        let mut base = profile("m-flow");
        base["ci"]["kind"] = json!(kind);
        let before = planned_text(&base);
        assert_eq!(
            before,
            planned_text(&base),
            "{kind}: the plan is not stable"
        );
        for (answer, pointer, value) in [
            ("checks", "/checks/1/command", json!("cargo test --release")),
            ("systems", "/ci/systems", json!(["linux", "windows"])),
            (
                "deploy",
                "/ci/deploy",
                json!({"enabled": true, "target": "production"}),
            ),
        ] {
            let mut changed = base.clone();
            *changed.pointer_mut(pointer).unwrap() = value;
            assert_ne!(planned_text(&changed), before, "{kind}: {answer}");
        }
    }
}

/// D-5: the tool applies no repository setting. With the answers whose settings the preview
/// shows as text (blocking checks, a deploy place), planning and applying run neither `gh` nor
/// `curl`: both are replaced on PATH by scripts that record each call.
#[cfg(unix)]
#[test]
fn the_settings_are_shown_as_text_and_no_command_is_run_to_apply_them() {
    use std::os::unix::fs::PermissionsExt;
    let bin = TempDir::new().unwrap();
    let calls = bin.path().join("calls.log");
    for name in ["gh", "curl"] {
        let path = bin.path().join(name);
        fs::write(
            &path,
            format!("#!/bin/sh\necho \"{name} $*\" >> \"${{0%/*}}/calls.log\"\n"),
        )
        .unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        // The recorder records: without this the missing log below would prove nothing.
        assert!(std::process::Command::new(&path)
            .status()
            .unwrap()
            .success());
    }
    assert_eq!(fs::read_to_string(&calls).unwrap(), "gh \ncurl \n");
    fs::remove_file(&calls).unwrap();

    let search = format!("{}:/usr/bin:/bin", bin.path().display());
    let env = [("PATH", search.as_str())];
    let mut answers = profile("github-flow");
    answers["ci"]["strict"] = json!("block");
    answers["ci"]["deploy"] = json!({"enabled": true, "target": "production"});
    let root = TempDir::new().unwrap();
    save(root.path(), &answers);
    let project = root.path().to_str().unwrap();
    let p = ok(&run_with_env(
        &["project-ci-plan", project],
        &json!({"protocol_version": 1}),
        &env,
    ));
    let notes = p["notes"].to_string();
    assert!(notes.contains("Settings, Branches"), "{notes}");
    assert!(notes.contains("Required reviewers"), "{notes}");
    ok(&run_with_env(
        &[
            "project-ci-apply",
            project,
            "--plan-digest",
            &digest(&p),
            "--permission-reference",
            "fixture-approved",
        ],
        &p,
        &env,
    ));
    assert!(root.path().join(CHECK).is_file());
    assert!(
        !calls.exists(),
        "a command was run: {}",
        fs::read_to_string(&calls).unwrap_or_default()
    );
}

/// A folder that stands where a file picked for removal was is refused, by the apply of the
/// earlier preview and by a new plan; the folder, what it holds and the other file stay.
#[test]
fn a_folder_at_a_picked_removal_path_is_refused_and_kept() {
    let root = project("m-flow");
    let p = plan(&root, json!({"protocol_version": 1}));
    ok(&apply(&root, &p, &digest(&p), "fixture-approved"));
    set_profile(root.path(), "github-flow");
    let pick = json!({"protocol_version": 1, "remove": [NIGHTLY]});
    let picked = plan(&root, pick.clone());
    assert!(actions(&picked).contains(&(NIGHTLY.to_string(), "remove".to_string())));
    let check = fs::read(root.path().join(CHECK)).unwrap();

    let at = root.path().join(NIGHTLY);
    fs::remove_file(&at).unwrap();
    fs::create_dir(&at).unwrap();
    fs::write(at.join("mine.txt"), "mine").unwrap();
    refused(&apply(&root, &picked, &digest(&picked), "fixture-approved"));
    refused(&run(
        &["project-ci-plan", root.path().to_str().unwrap()],
        &pick,
    ));
    assert_eq!(fs::read_to_string(at.join("mine.txt")).unwrap(), "mine");
    assert_eq!(fs::read(root.path().join(CHECK)).unwrap(), check);
}
