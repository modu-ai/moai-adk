//! Slice 2 of SPEC-20261120-shipped-workflows-commands: `moai check-ready` in a temporary
//! project under a temporary HOME. The command reads the project's declarations only — the
//! SPEC documents and the profile's `test_paths` — never a git range (that is `moai check
//! affected`'s input), and writes nothing: every case compares the whole tree before and
//! after.

mod common;

use common::cli::tree_state;
use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};

/// `(exit code, stderr)` of `moai check-ready --project P`.
fn run(home: &TempHome, project: &Path) -> (i32, Value, String) {
    let p = project.to_str().unwrap();
    let out = home
        .command(&["check-ready", "--project", p])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (
        out.status.code().unwrap_or(-1),
        v,
        String::from_utf8_lossy(&out.stderr).into_owned(),
    )
}

/// A project with this generation's `.moai/` marker and no profile: the folder `check-ready`
/// runs in.
fn bare(home: &TempHome) -> PathBuf {
    let p = home.path().join("work").join("bare");
    fs::create_dir_all(p.join(".moai")).unwrap();
    fs::write(
        p.join(".moai/moai.json"),
        moai_cli::state_root::marker_bytes("test", &[]),
    )
    .unwrap();
    p
}

/// The full profile of `project` with `test_paths` declared (hand-written after
/// `TempHome::project` stubbed the file once — the stub is bound to its path first, the
/// rewrite goes last). The profile needs its required fields: without them `read_profile`
/// refuses it before the declaration is even seen.
fn declare_test_paths(project: &Path, paths: &[&str]) {
    let paths: Vec<String> = paths.iter().map(|s| s.to_string()).collect();
    let yaml = format!(
        "{{\"schema_version\":2,\"project_id\":\"check-ready\",\"purpose\":\"Check the \
         ready checks\",\"tools\":[\"Rust\"],\"development_methods\":[\"TDD\"],\
         \"hosts\":[\"claude\"],\"test_paths\":{}}}\n",
        serde_json::to_string(&paths).unwrap()
    );
    fs::write(
        project
            .join(moai_cli::PROJECT_STATE_DIR)
            .join("project.yaml"),
        yaml,
    )
    .unwrap();
}

/// A valid draft SPEC in the project, through the `spec new` command itself.
fn new_spec(home: &TempHome, project: &Path, id: &str) -> (i32, String) {
    let p = project.to_str().unwrap();
    let out = home
        .command(&[
            "spec",
            "new",
            "login-fix",
            "--title",
            "Fix login",
            "--date",
            "2026-11-20",
            "--project",
            p,
        ])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let ok = v["ok"].as_bool().unwrap();
    let made = v["result"]["id"].as_str().unwrap_or("?");
    assert_eq!(made, id, "spec new made {made:?}, not {id:?}");
    (
        out.status.code().unwrap_or(-1),
        format!("ok={ok} id={made}"),
    )
}

/// Replaces the REQ line of the SPEC file with a sentence without `shall`.
fn break_spec(project: &Path, id: &str) {
    let file = project.join(".moai/specs").join(id).join("spec.md");
    let text = fs::read_to_string(&file).unwrap();
    let broken = text
        .lines()
        .map(|l| {
            if l.starts_with("- REQ-1:") {
                "- REQ-1: The command works.".to_string()
            } else {
                l.to_string()
            }
        })
        .collect::<Vec<_>>()
        .join("\n");
    fs::write(file, broken).unwrap();
}

#[test]
fn valid_declarations_pass_and_write_nothing() {
    let home = TempHome::new();
    let p = home.project("check-ready");
    fs::create_dir_all(p.join("tests")).unwrap();
    declare_test_paths(&p, &["tests"]);
    let (code, why) = new_spec(&home, &p, "SPEC-20261120-login-fix");
    assert_eq!(code, 0, "{why}");
    let before = tree_state(home.path());
    let (code, v, err) = run(&home, &p);
    assert_eq!(code, 0, "{err}");
    let r = &v["result"];
    assert_eq!(v["operation"], "check-ready");
    assert_eq!(r["action"], "check-ready");
    assert_eq!(r["result"], "PASS");
    assert_eq!(r["profile"], "present");
    assert_eq!(r["declared_paths"], 1);
    assert_eq!(r["missing_paths"], 0);
    assert_eq!(r["specs"], "present");
    assert_eq!(r["spec_folders"], 1);
    assert_eq!(r["nothing_declared"], false);
    // The command writes nothing on the tree it checked.
    let after = tree_state(home.path());
    assert_eq!(before, after, "check-ready must not write the project tree");
}

#[test]
fn missing_declared_path_fails_naming_it() {
    let home = TempHome::new();
    let p = home.project("check-ready");
    fs::create_dir_all(p.join("tests")).unwrap();
    declare_test_paths(&p, &["tests", "spec"]);
    let (code, v, err) = run(&home, &p);
    assert_eq!(code, 1, "{v}");
    assert!(err.contains("check-ready refused"), "{err}");
    assert!(
        err.contains("missing: spec"),
        "the missing path is named: {err}"
    );
    assert!(err.contains("1 problem"), "the one missing path: {err}");
    assert!(
        !err.contains("tests"),
        "the existing path is not a problem: {err}"
    );
}

#[test]
fn broken_spec_fails() {
    let home = TempHome::new();
    let p = home.project("check-ready");
    let (code, why) = new_spec(&home, &p, "SPEC-20261120-login-fix");
    assert_eq!(code, 0, "{why}");
    break_spec(&p, "SPEC-20261120-login-fix");
    let (code, _v, err) = run(&home, &p);
    assert_eq!(code, 1, "{err}");
    assert!(err.contains("check-ready refused"), "{err}");
    assert!(err.contains("spec.md"), "the SPEC file is named: {err}");
}

#[test]
fn nothing_declared_passes_with_exit_zero() {
    let home = TempHome::new();
    let p = bare(&home);
    let before = tree_state(home.path());
    let (code, v, err) = run(&home, &p);
    assert_eq!(code, 0, "{err}");
    let r = &v["result"];
    assert_eq!(r["result"], "PASS");
    assert_eq!(r["nothing_declared"], true);
    assert_eq!(r["profile"], "absent");
    assert_eq!(r["specs"], "absent");
    assert_eq!(r["spec_folders"], 0);
    assert_eq!(tree_state(home.path()), before, "nothing is written");
}

#[test]
fn declared_paths_pass_when_spec_folder_is_absent() {
    let home = TempHome::new();
    let p = home.project("check-ready");
    fs::create_dir_all(p.join("tests")).unwrap();
    declare_test_paths(&p, &["tests"]);
    let (code, v, err) = run(&home, &p);
    assert_eq!(code, 0, "{err}");
    let r = &v["result"];
    assert_eq!(r["result"], "PASS");
    assert_eq!(r["specs"], "absent");
    assert_eq!(r["declared_paths"], 1);
    assert_eq!(r["missing_paths"], 0);
}

#[test]
fn invalid_profile_is_a_problem() {
    let home = TempHome::new();
    let p = home.project("check-ready");
    // A declared path that walks out of the project is refused by the profile itself, so
    // `check-ready` reads it as an invalid profile, not as a missing path.
    fs::write(
        p.join(moai_cli::PROJECT_STATE_DIR).join("project.yaml"),
        "{\"schema_version\":2,\"project_id\":\"check-ready\",\"test_paths\":[\"../out\"]}\n",
    )
    .unwrap();
    let (code, _v, err) = run(&home, &p);
    assert_eq!(code, 1, "{err}");
    assert!(err.contains("invalid profile"), "{err}");
}
