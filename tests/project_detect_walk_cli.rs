//! `project-detect` finds sub-package manifests in a large monorepo and says when its list is
//! partial (SPEC-20261121-detect-walk-breadth), through the binary in temporary folders only.

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

const PARTIAL: &str = "manifest list is partial";

fn run(args: &[&str]) -> Output {
    // A temporary HOME, so the run never reads the real one.
    let home = TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    match pipe.write_all(b"") {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(pipe);
    child.wait_with_output().unwrap()
}

fn detect(root: &TempDir) -> Value {
    let out = run(&["project-detect", root.path().to_str().unwrap()]);
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn write(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn manifests(result: &Value) -> Vec<&str> {
    result["detected"]["manifests"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| m.as_str().unwrap())
        .collect()
}

fn partial_lines(result: &Value) -> usize {
    result["next_actions"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|l| l.as_str().unwrap().contains(PARTIAL))
        .count()
}

/// More entries than the walk budget (5000) sit in one folder three levels down that sorts
/// before `apps/`.
fn a_project_with_a_large_early_folder() -> TempDir {
    let root = TempDir::new().unwrap();
    let big = root.path().join("aaa/b/c");
    fs::create_dir_all(&big).unwrap();
    for i in 0..6000 {
        fs::write(big.join(format!("f{i:05}.txt")), "x").unwrap();
    }
    write(root.path(), "apps/x/package.json", "{}");
    write(root.path(), "apps/id/go.mod", "module x\n");
    root
}

/// SPEC-20261121-detect-walk-breadth REQ-1, REQ-3: the walk reads shallow folders first, so a
/// large folder that sorts early does not use up the budget before `apps/` is read; the list
/// is reported as partial because the budget did run out.
#[test]
fn a_large_early_folder_does_not_hide_a_later_packages_manifest() {
    let root = a_project_with_a_large_early_folder();
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/id/go.mod", "apps/x/package.json"],
        "{result}"
    );
    assert_eq!(
        result["manifest_walk"]["complete"],
        json!(false),
        "{result}"
    );
    assert_eq!(result["manifest_walk"]["over_cap"], json!(0), "{result}");
}

/// REQ-2: tool, cache and backup folders (names starting with a dot) are not walked for
/// manifests; a root manifest is read whatever the folder names are.
#[test]
fn hidden_folders_are_not_walked_for_manifests() {
    let root = TempDir::new().unwrap();
    write(root.path(), "package.json", "{}");
    write(root.path(), ".cache/pkg/package.json", "{}");
    write(root.path(), ".vercel/x/package.json", "{}");
    write(root.path(), ".moai-backup/y/go.mod", "module y\n");
    write(root.path(), "apps/x/package.json", "{}");
    let result = detect(&root);
    assert_eq!(
        manifests(&result),
        ["apps/x/package.json", "package.json"],
        "{result}"
    );
    assert_eq!(result["manifest_walk"]["complete"], json!(true), "{result}");
}

/// REQ-3: a small project reports a complete walk and nothing over the cap; 70 manifests are
/// 64 listed and 6 over the cap.
#[test]
fn the_manifest_walk_fact_reports_completeness_and_the_cap() {
    let small = TempDir::new().unwrap();
    write(small.path(), "package.json", "{}");
    write(small.path(), "apps/a/package.json", "{}");
    let result = detect(&small);
    assert_eq!(
        result["manifest_walk"],
        json!({"complete": true, "over_cap": 0}),
        "{result}"
    );

    let many = TempDir::new().unwrap();
    for i in 0..70 {
        write(many.path(), &format!("pkg{i:02}/package.json"), "{}");
    }
    let result = detect(&many);
    assert_eq!(manifests(&result).len(), 64, "{result}");
    assert_eq!(
        result["manifest_walk"],
        json!({"complete": true, "over_cap": 6}),
        "{result}"
    );
}

/// REQ-4: one `next_actions` line says the list is partial when the walk stopped or the cap cut
/// the list, and none when the list is complete.
#[test]
fn a_partial_list_adds_one_next_action_and_a_complete_one_adds_none() {
    let stopped = a_project_with_a_large_early_folder();
    assert_eq!(partial_lines(&detect(&stopped)), 1);

    let many = TempDir::new().unwrap();
    for i in 0..65 {
        write(many.path(), &format!("pkg{i:02}/package.json"), "{}");
    }
    assert_eq!(partial_lines(&detect(&many)), 1);

    let small = TempDir::new().unwrap();
    write(small.path(), "apps/a/package.json", "{}");
    assert_eq!(partial_lines(&detect(&small)), 0);
}
