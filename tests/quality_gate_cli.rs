//! SPEC-20261120-rust-quality-gate: the command's own behaviour — the pre-commit marker, the
//! enabled switch, the summary, the run lock. Names start with the SPEC's verify words
//! (`summary_`, `lock_`) so the filters pick them.

#[path = "../src/quality_gate/mod.rs"]
#[allow(dead_code)]
mod quality_gate;

use quality_gate::lock;
use quality_gate::{run, Invocation};
use std::ffi::OsString;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use tempfile::TempDir;

fn bin_dir(tools: &[(&str, &str)]) -> TempDir {
    let d = TempDir::new().unwrap();
    for (name, body) in tools {
        let p = d.path().join(name);
        fs::write(&p, format!("#!/bin/sh\n{body}\n")).unwrap();
        fs::set_permissions(&p, fs::Permissions::from_mode(0o755)).unwrap();
    }
    d
}

fn path_of(bin: &TempDir) -> OsString {
    std::env::join_paths([bin.path(), Path::new("/usr/bin"), Path::new("/bin")]).unwrap()
}

fn go_project() -> TempDir {
    let p = TempDir::new().unwrap();
    fs::write(p.path().join("go.mod"), "module x\n").unwrap();
    fs::write(p.path().join("main.go"), "package x\n").unwrap();
    p
}

fn config(project: &Path, text: &str) {
    let d = project.join(".moai/config/sections");
    fs::create_dir_all(&d).unwrap();
    fs::write(d.join("gate.yaml"), text).unwrap();
}

fn gate(project: &Path, bin: &TempDir, marker: bool) -> (i32, String) {
    let inv = Invocation {
        project_dir: project.to_path_buf(),
        precommit_marker: marker,
        path: path_of(bin),
    };
    let mut err = Vec::new();
    let code = run(&inv, &mut err);
    (code, String::from_utf8_lossy(&err).into_owned())
}

const SKIP_NOTICE: &str = "[moai gate] pre-commit context: heavy gate not opted in — skipping the project-wide steps (opt in via gate.pre_commit.enabled: true in .moai/config/sections/gate.yaml)\n";

/// AC-1 (REQ-1): with the marker and the default config no step runs and the notice is exact;
/// opting in runs the steps again.
#[test]
fn the_precommit_marker_skips_unless_opted_in() {
    let p = go_project();
    let bin = bin_dir(&[("go", "echo ran >&2; exit 3")]);
    let (code, err) = gate(p.path(), &bin, true);
    assert_eq!(code, 0, "{err}");
    assert_eq!(err, SKIP_NOTICE);
    config(p.path(), "gate:\n  pre_commit:\n    enabled: true\n");
    let (code, err) = gate(p.path(), &bin, true);
    assert_eq!(
        code, 1,
        "opted in: the steps run and the failing one fails: {err}"
    );
    // Without the marker the gate runs whatever the opt-in says.
    let q = go_project();
    let (code, err) = gate(q.path(), &bin, false);
    assert_eq!(code, 1, "{err}");
}

/// AC-1 (REQ-10): `enabled: false` prints nothing and exits 0 even for a project whose step
/// would fail.
#[test]
fn a_disabled_gate_passes_silently() {
    let p = go_project();
    config(p.path(), "gate:\n  enabled: false\n");
    let bin = bin_dir(&[("go", "exit 3")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0);
    assert_eq!(err, "", "no output at all");
}

/// AC-5 (REQ-8): the passing run prints the summary: header with the configured count, one row
/// per step in run order, the executed rows with duration, start time and command.
#[test]
fn summary_lists_every_configured_step_in_order() {
    let p = go_project();
    let bin = bin_dir(&[("go", "exit 0")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0, "{err}");
    let at = err
        .find("quality gate steps (4 configured):")
        .unwrap_or_else(|| panic!("{err}"));
    let rows: Vec<&str> = err[at..].lines().skip(1).take(4).collect();
    assert!(rows[0].starts_with("  - go vet: executed in "), "{rows:?}");
    assert!(rows[0].ends_with(" — go vet ./..."), "{rows:?}");
    assert!(
        rows[0].contains(" at 20"),
        "start time is an RFC 3339 stamp: {rows:?}"
    );
    assert_eq!(rows[1], "  - typecheck: skipped — no default for this language; set gate.typecheck.command to enable one");
    assert_eq!(
        rows[2],
        "  - golangci-lint: skipped — the optional tool \"golangci-lint\" was not found on PATH"
    );
    assert!(
        rows[3].starts_with("  - go test: executed in ") && rows[3].ends_with(" — go test ./..."),
        "{rows:?}"
    );
}

/// AC-5 (REQ-8): a failed step prints the failure block, then the summary where the steps after
/// it read `not reached`.
#[test]
fn summary_marks_the_steps_after_a_failure_as_not_reached() {
    let p = go_project();
    let bin = bin_dir(&[("go", "echo vet-broke >&2; exit 3")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 1, "{err}");
    let failed = err
        .find("quality gate failed: go vet\n\nvet-broke\n\nquality gate steps (4 configured):\n")
        .unwrap_or_else(|| panic!("{err}"));
    let tail = &err[failed..];
    assert!(tail.contains("\n  - typecheck: not reached\n"), "{tail}");
    assert!(
        tail.contains("\n  - golangci-lint: not reached\n"),
        "{tail}"
    );
    assert!(tail.contains("\n  - go test: not reached"), "{tail}");
}

/// AC-5 (REQ-8): the axes this build does not run yet are named, not left out (D-2).
#[test]
fn summary_names_the_axes_not_ported_yet() {
    let p = go_project();
    let bin = bin_dir(&[("go", "exit 0")]);
    let (_, err) = gate(p.path(), &bin, false);
    for axis in ["graph-freshness", "ast-grep"] {
        assert!(
            err.contains(&format!("  - {axis}: not run — not ported yet")),
            "{axis}: {err}"
        );
    }
    // A node project also names the node-only decisions (test and lint selection, typecheck).
    let n = TempDir::new().unwrap();
    fs::write(n.path().join("package.json"), "{}").unwrap();
    let (_, err) = gate(
        n.path(),
        &bin_dir(&[("npm", "exit 0"), ("npx", "exit 0")]),
        false,
    );
    assert!(err.contains("not run — not ported yet"), "{err}");
    assert!(err.contains("npm"), "the node steps are named: {err}");
}

/// AC-5 (REQ-8): a project with no recognised language prints nothing and passes.
#[test]
fn summary_is_absent_when_no_language_is_detected() {
    let p = TempDir::new().unwrap();
    let (code, err) = gate(p.path(), &bin_dir(&[]), false);
    assert_eq!(code, 0);
    assert!(!err.contains("quality gate steps"), "{err}");
}

/// AC-5 (REQ-9): the lock file is created under `.moai/state/` and is free again after a run.
#[test]
fn lock_is_taken_and_released_by_a_run() {
    let p = go_project();
    let bin = bin_dir(&[("go", "exit 0")]);
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0, "{err}");
    let lock_path = p.path().join(".moai/state/gate-run.lock");
    assert!(lock_path.exists());
    assert!(lock::acquire(p.path()).is_ok(), "released after the run");
}

/// AC-5 (REQ-9): a held lock makes the run wait its budget, say so, then run unserialized; the
/// lock never decides the exit code.
#[test]
fn lock_held_by_another_run_waits_then_runs_unserialized() {
    let p = go_project();
    config(p.path(), "gate:\n  timeouts:\n    lock_wait: 1\n");
    let held = lock::acquire(p.path()).expect("first acquire");
    // `acquire` records this process as the owner; a second acquire in the same process must
    // be refused as held (flock is per open file description).
    assert!(matches!(
        lock::acquire(p.path()),
        Err(lock::LockError::Held)
    ));
    let bin = bin_dir(&[("go", "exit 0")]);
    let started = std::time::Instant::now();
    let (code, err) = gate(p.path(), &bin, false);
    assert_eq!(code, 0, "{err}");
    assert!(started.elapsed().as_millis() >= 900, "waited the budget");
    let pid = std::process::id();
    assert!(
        err.contains(&format!(
            "gate-run lock: held by pid {pid} — waiting (budget 1s)"
        )),
        "{err}"
    );
    assert!(
        err.contains("gate-run lock: wait budget 1s expired — running unserialized"),
        "{err}"
    );
    drop(held);
}

/// AC-5 (REQ-9): the lock file's owner record is the Go build's shape, so a Go run and a Rust
/// run on one tree can read each other's holder.
#[test]
fn lock_owner_record_is_json_with_pid_and_created_at() {
    let p = go_project();
    let held = lock::acquire(p.path()).unwrap();
    let raw = fs::read_to_string(p.path().join(".moai/state/gate-run.lock")).unwrap();
    assert!(
        raw.contains(&format!("\"pid\": {}", std::process::id())),
        "{raw}"
    );
    assert!(raw.contains("\"created_at\": \"20"), "{raw}");
    drop(held);
    let _ = PathBuf::new();
}
