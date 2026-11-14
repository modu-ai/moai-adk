//! The product CLI helpers that group test files share (#2834 R1): run one `moai` command with a
//! JSON value on stdin, unwrap a success envelope, create a fixture project profile, and write or
//! snapshot a fixture tree. Moved from `tests/product_cli.rs` unchanged.

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// Runs `moai ARGS` with `input` (JSON) on stdin and returns its output.
#[allow(dead_code)] // every test crate includes `common`; not every one runs the product CLI
pub fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(super::MOAI_BIN)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    // The child may validate its arguments and exit before reading stdin; on a fast
    // runner that surfaces as EPIPE here. The tests assert the child's exit status and
    // output, never that the stdin write completed, so only a broken pipe is tolerated.
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
    {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    child.wait_with_output().unwrap()
}

/// Result of an internal operation: asserts the success envelope (01-C #1990) and unwraps it.
#[allow(dead_code)] // as above
pub fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// v2 profile input (REQ-03A-01/08): `project-plan` accepts schema_version 2 only;
/// v1 files stay readable through `read_profile` with defaulted new fields.
#[allow(dead_code)] // as above
pub fn profile(pin: Option<String>) -> Value {
    let mut v = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
    if let Some(pin) = pin {
        v["package_pin"] = json!(pin);
    }
    v
}

/// Creates the project profile `input` in `root` through `project-plan` and `project-apply`.
#[allow(dead_code)] // as above
pub fn create(root: &TempDir, input: &Value) -> Value {
    let p = root.path().to_str().unwrap();
    // An earlier-root project: its state stays in `.moai-cli/` (#2696 C2b gives a project
    // with no state `.moai/`; that layout is pinned in tests/state_root_cli.rs).
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let preview = cli(&["project-plan", p], input);
    assert!(preview.status.success(), "{:?}", preview.stderr);
    let plan: Value = ok_result(&preview);
    let result = cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    assert!(result.status.success(), "{:?}", result.stderr);
    plan
}

/// Writes `bytes` to `root/rel`, creating its parent folders.
#[allow(dead_code)] // as above
pub fn write_file(root: &std::path::Path, rel: &str, bytes: &[u8]) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}
/// Every path and byte under `dir`, `.git` included, without following symlinks.
#[allow(dead_code)] // as above
pub fn tree_state(dir: &std::path::Path) -> Vec<(String, Vec<u8>)> {
    fn walk(root: &std::path::Path, dir: &std::path::Path, out: &mut Vec<(String, Vec<u8>)>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.file_type().is_symlink() {
                let t = fs::read_link(&p).unwrap();
                out.push((rel, t.to_string_lossy().as_bytes().to_vec()));
            } else if m.is_dir() {
                out.push((rel, Vec::new()));
                walk(root, &p, out);
            } else {
                out.push((rel, fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, dir, &mut out);
    out.sort();
    out
}
