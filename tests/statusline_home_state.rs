//! Issue #2455 (user decision 2026-11-11): the statusline reads its HOME state files
//! only from `$HOME/.moai-cli/state/`, never from the Go folder `$HOME/.moai/`, and
//! `MOAI_HOME` (the Go tool's variable) no longer redirects it. `MOAI_LAST_MODEL` is an
//! environment value, not a file read, and keeps its precedence over the model file.
//!
//! Every case runs the real binary with a cleared environment and a temporary HOME; the
//! working directory is a separate temporary folder outside that HOME. Paths are spelled
//! literally on purpose: a test that went through the product constant would follow a
//! regression instead of catching it.

use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

struct Fixture {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    work: PathBuf,
}

fn fixture() -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let home = tmp.path().join("home");
    let work = tmp.path().join("work");
    std::fs::create_dir_all(&home).unwrap();
    std::fs::create_dir_all(&work).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        work,
    }
}

fn write(base: &Path, rel: &str, bytes: &[u8]) {
    let path = base.join(rel);
    std::fs::create_dir_all(path.parent().unwrap()).unwrap();
    std::fs::write(path, bytes).unwrap();
}

/// One plain render with stdin `{}`: no model in the input, so the model segment can only
/// come from `MOAI_LAST_MODEL` or the HOME model file.
fn render(f: &Fixture, extra: &[(&str, &str)]) -> String {
    render_in(f, &f.work, extra)
}

fn render_in(f: &Fixture, cwd: &Path, extra: &[(&str, &str)]) -> String {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.args(["statusline", "--render-budget-ms", "3600000"])
        .current_dir(cwd)
        .env_clear()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", &f.home)
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    for (k, v) in extra {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    child.stdin.take().unwrap().write_all(b"{}").unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "statusline exit {:?}: {}",
        out.status.code(),
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap()
}

const GO_MODEL: &[u8] = b"go-marker-model\n";
const GO_TASK: &[u8] =
    br#"{"active_task": {"command": "gomarker", "spec_id": "SPEC-GO-1", "stage": "red"}}"#;
const RC_MODEL: &[u8] = b"rc-marker-model\n";
const RC_TASK: &[u8] =
    br#"{"active_task": {"command": "rcmarker", "spec_id": "SPEC-RC-1", "stage": "green"}}"#;

fn write_go_state(f: &Fixture) {
    write(&f.home, ".moai/state/last-model.txt", GO_MODEL);
    write(&f.home, ".moai/state/last-session-state.json", GO_TASK);
}

fn write_rc_state(f: &Fixture) {
    write(&f.home, ".moai-cli/state/last-model.txt", RC_MODEL);
    write(&f.home, ".moai-cli/state/last-session-state.json", RC_TASK);
}

fn assert_no_go_markers(out: &str) {
    assert!(
        !out.contains("go-marker-model"),
        "Go model file was read: {out}"
    );
    assert!(
        !out.contains("gomarker"),
        "Go session state was read: {out}"
    );
}

/// A sorted listing of every entry below `dir` with its bytes, to prove nothing changed.
fn snapshot(dir: &Path) -> Vec<(PathBuf, Vec<u8>)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in std::fs::read_dir(&d).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p.clone());
                out.push((p, Vec::new()));
            } else {
                out.push((p.clone(), std::fs::read(&p).unwrap()));
            }
        }
    }
    out.sort();
    out
}

/// AC-1: only the Go files exist — neither marker appears, with or without `MOAI_HOME`
/// pointing at the Go folder, and the Go folder is left byte for byte as it was.
#[test]
fn go_home_state_files_are_never_read() {
    let f = fixture();
    write_go_state(&f);
    let before = snapshot(&f.home);

    let out = render(&f, &[]);
    assert_no_go_markers(&out);
    assert!(!out.contains("🤖"), "{out}");
    assert!(!out.contains("📋"), "{out}");

    let go_home = f.home.join(".moai");
    let out = render(&f, &[("MOAI_HOME", go_home.to_str().unwrap())]);
    assert_no_go_markers(&out);
    assert!(!out.contains("📋"), "{out}");

    assert_eq!(snapshot(&f.home), before, "HOME changed during render");
}

/// AC-2: the same files under `.moai-cli/state/` are read. With both locations filled the
/// RC files win and the Go markers stay out. The session task (📋) is never shown since the new
/// three-line layout (#2376, user decision 3), even when its state file is present.
#[test]
fn moai_cli_home_state_files_are_read() {
    let f = fixture();
    write_rc_state(&f);
    let out = render(&f, &[]);
    assert!(out.contains("🤖 rc-marker-model"), "{out}");
    assert!(!out.contains("📋") && !out.contains("rcmarker"), "{out}");

    write_go_state(&f);
    let go_home = f.home.join(".moai");
    for extra in [vec![], vec![("MOAI_HOME", go_home.to_str().unwrap())]] {
        let out = render(&f, &extra);
        assert!(out.contains("🤖 rc-marker-model"), "{out}");
        assert!(!out.contains("📋") && !out.contains("rcmarker"), "{out}");
        assert_no_go_markers(&out);
    }
}

/// `MOAI_LAST_MODEL` is kept and still wins over the model file.
#[test]
fn moai_last_model_still_overrides_the_model_file() {
    let f = fixture();
    write_rc_state(&f);
    let out = render(&f, &[("MOAI_LAST_MODEL", "env-marker-model")]);
    assert!(out.contains("🤖 env-marker-model"), "{out}");
    assert!(!out.contains("rc-marker-model"), "{out}");
}

/// AC-2 boundaries for the model file: absent, 0 bytes, whitespace only. Each renders
/// (exit 0) without a model segment; one visible character is enough to show it.
#[test]
fn model_file_boundaries() {
    let f = fixture();
    let out = render(&f, &[]);
    assert!(!out.contains("🤖"), "absent: {out}");

    for (case, bytes) in [("empty", &b""[..]), ("whitespace", &b" \n\t\n"[..])] {
        write(&f.home, ".moai-cli/state/last-model.txt", bytes);
        let out = render(&f, &[]);
        assert!(!out.contains("🤖"), "{case}: {out}");
    }

    write(&f.home, ".moai-cli/state/last-model.txt", b"m");
    let out = render(&f, &[]);
    assert!(out.contains("🤖 m"), "one character: {out}");
}

/// AC-2 boundaries for the session state: absent, 0 bytes, broken JSON, no `active_task`,
/// an empty command, a minimal command. Each renders (exit 0), and none shows a task segment —
/// the new layout never shows 📋 (#2376, user decision 3).
#[test]
fn session_state_boundaries() {
    let f = fixture();
    let out = render(&f, &[]);
    assert!(!out.contains("📋"), "absent: {out}");

    for (case, bytes) in [
        ("empty", &b""[..]),
        ("broken", &br#"{"active_task": {"command": "rcmarker""#[..]),
        ("no active_task", &br#"{"current_branch": "main"}"#[..]),
        ("empty command", &br#"{"active_task": {"command": ""}}"#[..]),
    ] {
        write(&f.home, ".moai-cli/state/last-session-state.json", bytes);
        let out = render(&f, &[]);
        assert!(!out.contains("📋"), "{case}: {out}");
    }

    write(
        &f.home,
        ".moai-cli/state/last-session-state.json",
        br#"{"active_task": {"command": "r"}}"#,
    );
    let out = render(&f, &[]);
    assert!(
        !out.trim().is_empty(),
        "minimal command still renders: {out}"
    );
    assert!(!out.contains("📋"), "minimal command: {out}");
}

/// AC-5 (leader scope 6792437161): with the cwd deep below HOME and only
/// `HOME/.moai/config/sections/system.yaml` present, its template version never reaches
/// the output; at HOME itself neither. A project folder directly below HOME is still read
/// (#2389 project fallback).
#[test]
fn home_dot_moai_is_not_project_configuration() {
    let f = fixture();
    write(
        &f.home,
        ".moai/config/sections/system.yaml",
        b"moai:\n  template_version: v9.9.9\n",
    );
    let deep = f.home.join("a/b/c");
    std::fs::create_dir_all(&deep).unwrap();
    for cwd in [&deep, &f.home] {
        let out = render_in(&f, cwd, &[]);
        assert!(!out.contains("9.9.9"), "{cwd:?}: {out}");
    }

    let proj = f.home.join("proj");
    write(
        &proj,
        ".moai/config/sections/system.yaml",
        b"moai:\n  template_version: v1.2.3\n",
    );
    let sub = proj.join("sub");
    std::fs::create_dir_all(&sub).unwrap();
    for cwd in [&proj, &sub] {
        let out = render_in(&f, cwd, &[]);
        assert!(out.contains("1.2.3"), "{cwd:?}: {out}");
        assert!(!out.contains("9.9.9"), "{cwd:?}: {out}");
    }
}
