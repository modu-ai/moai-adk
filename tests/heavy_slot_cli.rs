//! #2677 (AC-1, AC-2): `heavy-slot-acquire` / `heavy-slot-status` — a machine-wide limit on
//! concurrent heavy runs, the product port of `tools/heavy_check_slot.py` (#2491). Every case
//! uses its own temporary slot folder (`--dir`); nothing under the real HOME cache is touched.
//! Waits are bounded by progress markers plus a generous hang guard, never a bare budget.
#![cfg(unix)]

use serde_json::Value;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Child, Command, Output, Stdio},
    time::{Duration, Instant},
};

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const HANG: Duration = Duration::from_secs(120);

fn acquire(dir: &Path, max: &str, name: &str, cmd: &[&str]) -> Command {
    let mut c = Command::new(EXE);
    c.args(["heavy-slot-acquire", "--max", max, "--name", name, "--dir"])
        .arg(dir)
        .arg("--")
        .args(cmd);
    c
}

fn status(dir: &Path) -> Value {
    let out = Command::new(EXE)
        .args(["heavy-slot-status", "--dir"])
        .arg(dir)
        .output()
        .unwrap();
    assert!(out.status.success(), "{out:?}");
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["operation"], "heavy_slot.status", "{v}");
    v["result"].clone()
}

/// Waits (bounded by the hang guard) until `cond` holds.
fn wait_for(what: &str, mut cond: impl FnMut() -> bool) {
    let start = Instant::now();
    while !cond() {
        assert!(start.elapsed() < HANG, "hang guard: {what}");
        std::thread::sleep(Duration::from_millis(20));
    }
}

/// A shell command that touches `started`, then waits until `release` exists.
fn hold(started: &Path, release: &Path) -> String {
    format!(
        "touch '{}'; while [ ! -e '{}' ]; do sleep 0.05; done",
        started.display(),
        release.display()
    )
}

fn finish(child: Child) -> Output {
    child.wait_with_output().unwrap()
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// The command's own exit code comes back: 0, 3, and 128 + the signal for a signal death.
#[test]
fn the_commands_exit_code_comes_back() {
    let d = tempfile::tempdir().unwrap();
    for (cmd, code) in [("exit 0", 0), ("exit 3", 3), ("kill -TERM $$", 128 + 15)] {
        let out = acquire(d.path(), "2", "t", &["sh", "-c", cmd])
            .output()
            .unwrap();
        assert_eq!(out.status.code(), Some(code), "{cmd}: {}", text(&out));
    }
    assert_eq!(status(d.path())["held"], serde_json::json!([]));
}

/// AC-2: a failing command is reported with the load averages and the free disk space of the
/// working folder; a passing one is not.
#[test]
fn a_failure_reports_load_and_disk() {
    let d = tempfile::tempdir().unwrap();
    let out = acquire(d.path(), "1", "lane-x", &["sh", "-c", "exit 1"])
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1));
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(err.contains("lane-x exited 1"), "{err}");
    assert!(err.contains("load "), "{err}");
    assert!(err.contains("disk free "), "{err}");
    let ok = acquire(d.path(), "1", "lane-x", &["true"])
        .output()
        .unwrap();
    assert!(!String::from_utf8_lossy(&ok.stderr).contains("disk free"));
}

/// AC-1: with one slot the second caller waits — printing its place with load and disk — and
/// runs only after the first ends; `status` names the holder meanwhile.
#[test]
fn with_one_slot_the_second_caller_waits_its_turn() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let (s1, r1, s2) = (
        m.path().join("s1"),
        m.path().join("r1"),
        m.path().join("s2"),
    );
    let first = acquire(d.path(), "1", "first", &["sh", "-c", &hold(&s1, &r1)])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    wait_for("first started", || s1.exists());
    let held = status(d.path());
    assert_eq!(held["held"][0]["name"], "first", "{held}");
    assert_eq!(held["held"][0]["slot"], 0);

    let second = acquire(
        d.path(),
        "1",
        "second",
        &["sh", "-c", &format!("touch '{}'", s2.display())],
    )
    .stdout(Stdio::piped())
    .stderr(Stdio::piped())
    .spawn()
    .unwrap();
    // The second registers as a waiter; it never starts while the first holds the slot.
    wait_for("second waiting", || {
        fs::read_dir(d.path().join("waiting"))
            .map(|r| r.count() == 1)
            .unwrap_or(false)
    });
    assert!(!s2.exists(), "second ran while the slot was held");
    fs::write(&r1, "").unwrap();
    let a = finish(first);
    let b = finish(second);
    assert!(a.status.success(), "{}", text(&a));
    assert!(b.status.success(), "{}", text(&b));
    assert!(s2.exists());
    let waited = text(&b);
    assert!(
        waited.contains("heavy-slot: waiting 1 of 1 for 1 slot(s)"),
        "{waited}"
    );
    assert!(
        waited.contains("load ") && waited.contains("disk free "),
        "{waited}"
    );
    assert_eq!(status(d.path())["held"], serde_json::json!([]));
}

/// The lock lives on the file handed to the command: killing the wrapper with SIGKILL keeps
/// the slot held while the command runs, and the slot is free once the command ends.
#[test]
fn the_slot_stays_held_by_the_command_after_the_wrapper_is_killed() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let (s1, r1) = (m.path().join("s1"), m.path().join("r1"));
    let mut wrapper = acquire(d.path(), "1", "w", &["sh", "-c", &hold(&s1, &r1)])
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .unwrap();
    wait_for("command started", || s1.exists());
    wrapper.kill().unwrap();
    wrapper.wait().unwrap();
    assert_eq!(status(d.path())["held"].as_array().unwrap().len(), 1);
    fs::write(&r1, "").unwrap();
    wait_for("slot released", || {
        status(d.path())["held"].as_array().unwrap().is_empty()
    });
}

/// SIGTERM to the wrapper reaches the command, and the wrapper returns 128 + 15.
#[test]
fn a_signal_to_the_wrapper_reaches_the_command() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let (s1, got) = (m.path().join("s1"), m.path().join("got"));
    let script = format!(
        "trap 'touch \"{}\"; exit 143' TERM; touch '{}'; while :; do sleep 0.05; done",
        got.display(),
        s1.display()
    );
    let wrapper = acquire(d.path(), "1", "sig", &["sh", "-c", &script])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    wait_for("command started", || s1.exists());
    let pid = wrapper.id().to_string();
    assert!(Command::new("kill")
        .args(["-TERM", &pid])
        .status()
        .unwrap()
        .success());
    let out = finish(wrapper);
    assert!(
        got.exists(),
        "the command never saw SIGTERM: {}",
        text(&out)
    );
    assert_eq!(out.status.code(), Some(143), "{}", text(&out));
}

/// A slot path that is not a plain file of its own (here a symlink) is refused with 125
/// before anything is written; the command does not run and the link target is untouched.
#[test]
fn a_symlinked_slot_is_refused_before_anything_runs() {
    let d = tempfile::tempdir().unwrap();
    let outside = tempfile::tempdir().unwrap();
    let target = outside.path().join("victim");
    fs::write(&target, "keep\n").unwrap();
    std::os::unix::fs::symlink(&target, d.path().join("slot-0")).unwrap();
    let ran = outside.path().join("ran");
    let out = acquire(
        d.path(),
        "1",
        "x",
        &["sh", "-c", &format!("touch '{}'", ran.display())],
    )
    .output()
    .unwrap();
    assert_eq!(out.status.code(), Some(125), "{}", text(&out));
    assert!(text(&out).contains("not a plain file"), "{}", text(&out));
    assert!(!ran.exists());
    assert_eq!(fs::read_to_string(&target).unwrap(), "keep\n");
    let st = status(d.path());
    assert_eq!(st["refused"], serde_json::json!(["slot-0"]), "{st}");
}

/// Bad arguments: `--max 0` or no command is a usage error (2); nothing is created.
#[test]
fn bad_arguments_are_refused() {
    let d = tempfile::tempdir().unwrap();
    let slots = d.path().join("slots");
    let out = acquire(&slots, "0", "x", &["true"]).output().unwrap();
    assert_eq!(out.status.code(), Some(2), "{}", text(&out));
    let out = Command::new(EXE)
        .args(["heavy-slot-acquire", "--max", "1", "--name", "x", "--dir"])
        .arg(&slots)
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(2), "{}", text(&out));
    assert!(!slots.exists());
}

fn python_tool() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("tools/heavy_check_slot.py")
}

/// The Rust command and the repository's Python tool share one folder and one lock: with one
/// slot, whichever holds it makes the other wait (each direction).
#[test]
fn the_python_tool_and_the_command_take_turns_in_one_folder() {
    for python_first in [true, false] {
        let d = tempfile::tempdir().unwrap();
        let m = tempfile::tempdir().unwrap();
        let (s1, r1, s2) = (
            m.path().join("s1"),
            m.path().join("r1"),
            m.path().join("s2"),
        );
        let py = |cmd: &str| {
            let mut c = Command::new("python3");
            c.arg("-I")
                .arg(python_tool())
                .args(["acquire", "--max", "1", "--name", "py", "--dir"])
                .arg(d.path())
                .args(["--", "sh", "-c", cmd]);
            c
        };
        let hold_cmd = hold(&s1, &r1);
        let then_cmd = format!("touch '{}'", s2.display());
        let (mut a, mut b) = if python_first {
            (
                py(&hold_cmd),
                acquire(d.path(), "1", "rs", &["sh", "-c", &then_cmd]),
            )
        } else {
            (
                acquire(d.path(), "1", "rs", &["sh", "-c", &hold_cmd]),
                py(&then_cmd),
            )
        };
        let first = a
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        wait_for("holder started", || s1.exists());
        let second = b
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        wait_for("second waiting", || {
            fs::read_dir(d.path().join("waiting"))
                .map(|r| r.count() == 1)
                .unwrap_or(false)
        });
        assert!(
            !s2.exists(),
            "python_first={python_first}: both ran at once"
        );
        fs::write(&r1, "").unwrap();
        let (x, y) = (finish(first), finish(second));
        assert!(x.status.success(), "{}", text(&x));
        assert!(y.status.success(), "{}", text(&y));
        assert!(s2.exists());
    }
}
