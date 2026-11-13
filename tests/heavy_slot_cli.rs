//! #2677 (AC-1, AC-2): `heavy-slot-acquire` / `heavy-slot-status` — a machine-wide limit on
//! concurrent heavy runs, the product port of `tools/heavy_check_slot.py` (#2491). Every case
//! uses its own temporary slot folder (`--dir`); nothing under the real HOME cache is touched.
//! Waits are bounded by progress markers plus a generous hang guard, never a bare budget.
#![cfg(unix)]

use serde_json::Value;
use std::{
    fs,
    io::Read,
    os::unix::process::CommandExt,
    path::{Path, PathBuf},
    process::{Child, Command, Output, Stdio},
    time::{Duration, Instant},
};

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");
const HANG: Duration = Duration::from_secs(120);

/// A spawned wrapper in a process group of its own. Whatever happens to the test — a failed
/// assertion, a hang — the whole group (the wrapper and the command it started) is killed on
/// drop, so no holder loop outlives the test (an earlier run left orphans looping for hours).
struct Spawned {
    child: Option<Child>,
    group: i32,
}

fn spawn(mut c: Command) -> Spawned {
    c.process_group(0);
    let child = c.spawn().unwrap();
    let group = child.id() as i32;
    Spawned {
        child: Some(child),
        group,
    }
}

fn kill_group(group: i32) {
    // SAFETY: kill(2) on a process group this test created; ESRCH when it is already gone.
    unsafe { libc::kill(-group, libc::SIGKILL) };
}

impl Spawned {
    fn id(&self) -> u32 {
        self.child.as_ref().unwrap().id()
    }

    /// Waits for the wrapper (bounded by the hang guard), kills what is left of its group, and
    /// returns its output.
    fn finish(mut self) -> Output {
        let mut child = self.child.take().unwrap();
        let start = Instant::now();
        let status = loop {
            if let Some(s) = child.try_wait().unwrap() {
                break s;
            }
            if start.elapsed() > HANG {
                kill_group(self.group);
                let _ = child.wait();
                panic!("hang guard: the wrapper did not end");
            }
            std::thread::sleep(Duration::from_millis(20));
        };
        kill_group(self.group);
        let (mut stdout, mut stderr) = (Vec::new(), Vec::new());
        if let Some(mut o) = child.stdout.take() {
            o.read_to_end(&mut stdout).unwrap();
        }
        if let Some(mut e) = child.stderr.take() {
            e.read_to_end(&mut stderr).unwrap();
        }
        Output {
            status,
            stdout,
            stderr,
        }
    }

    fn kill_wrapper_only(&mut self) {
        let c = self.child.as_mut().unwrap();
        c.kill().unwrap();
        c.wait().unwrap();
    }
}

impl Drop for Spawned {
    fn drop(&mut self) {
        kill_group(self.group);
        if let Some(mut c) = self.child.take() {
            let _ = c.wait();
        }
    }
}

/// Runs `c` to the end in its own process group, bounded by the hang guard.
fn run(mut c: Command) -> Output {
    c.stdout(Stdio::piped()).stderr(Stdio::piped());
    spawn(c).finish()
}

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
        let out = run(acquire(d.path(), "2", "t", &["sh", "-c", cmd]));
        assert_eq!(out.status.code(), Some(code), "{cmd}: {}", text(&out));
    }
    assert_eq!(status(d.path())["held"], serde_json::json!([]));
}

/// AC-2: a failing command is reported with the load averages and the free disk space of the
/// working folder; a passing one is not.
#[test]
fn a_failure_reports_load_and_disk() {
    let d = tempfile::tempdir().unwrap();
    let out = run(acquire(d.path(), "1", "lane-x", &["sh", "-c", "exit 1"]));
    assert_eq!(out.status.code(), Some(1));
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(err.contains("lane-x exited 1"), "{err}");
    assert!(err.contains("load "), "{err}");
    assert!(err.contains("disk free "), "{err}");
    let ok = run(acquire(d.path(), "1", "lane-x", &["true"]));
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
    let mut c = acquire(d.path(), "1", "first", &["sh", "-c", &hold(&s1, &r1)]);
    c.stdout(Stdio::piped()).stderr(Stdio::piped());
    let first = spawn(c);
    wait_for("first started", || s1.exists());
    let held = status(d.path());
    assert_eq!(held["held"][0]["name"], "first", "{held}");
    assert_eq!(held["held"][0]["slot"], 0);

    let mut c = acquire(
        d.path(),
        "1",
        "second",
        &["sh", "-c", &format!("touch '{}'", s2.display())],
    );
    c.stdout(Stdio::piped()).stderr(Stdio::piped());
    let second = spawn(c);
    // The second registers as a waiter; it never starts while the first holds the slot.
    wait_for("second waiting", || {
        fs::read_dir(d.path().join("waiting"))
            .map(|r| r.count() == 1)
            .unwrap_or(false)
    });
    assert!(!s2.exists(), "second ran while the slot was held");
    fs::write(&r1, "").unwrap();
    let a = first.finish();
    let b = second.finish();
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
    let mut c = acquire(d.path(), "1", "w", &["sh", "-c", &hold(&s1, &r1)]);
    c.stdout(Stdio::null()).stderr(Stdio::null());
    let mut wrapper = spawn(c);
    wait_for("command started", || s1.exists());
    wrapper.kill_wrapper_only();
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
    let mut c = acquire(d.path(), "1", "sig", &["sh", "-c", &script]);
    c.stdout(Stdio::piped()).stderr(Stdio::piped());
    let wrapper = spawn(c);
    wait_for("command started", || s1.exists());
    // To the wrapper's pid only, never the group: the command must get it by forwarding.
    let pid = wrapper.id().to_string();
    assert!(Command::new("kill")
        .args(["-TERM", &pid])
        .status()
        .unwrap()
        .success());
    let out = wrapper.finish();
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
    let out = run(acquire(
        d.path(),
        "1",
        "x",
        &["sh", "-c", &format!("touch '{}'", ran.display())],
    ));
    assert_eq!(out.status.code(), Some(125), "{}", text(&out));
    assert!(text(&out).contains("not a plain file"), "{}", text(&out));
    assert!(!ran.exists());
    assert_eq!(fs::read_to_string(&target).unwrap(), "keep\n");
    let st = status(d.path());
    assert_eq!(st["refused"], serde_json::json!(["slot-0"]), "{st}");
}

/// A slot file that is a hard link to a file elsewhere is refused the same way (mutation M5
/// survived without this case): the other file keeps its bytes and the command does not run.
#[test]
fn a_hard_linked_slot_is_refused_before_anything_runs() {
    let d = tempfile::tempdir().unwrap();
    let target = d.path().join("elsewhere");
    fs::write(&target, "keep\n").unwrap();
    fs::hard_link(&target, d.path().join("slot-0")).unwrap();
    let ran = d.path().join("ran");
    let out = run(acquire(
        d.path(),
        "1",
        "x",
        &["sh", "-c", &format!("touch '{}'", ran.display())],
    ));
    assert_eq!(out.status.code(), Some(125), "{}", text(&out));
    assert!(text(&out).contains("not a plain file"), "{}", text(&out));
    assert!(!ran.exists());
    assert_eq!(fs::read_to_string(&target).unwrap(), "keep\n");
}

/// Bad arguments: `--max 0` or no command is a usage error (2); nothing is created.
#[test]
fn bad_arguments_are_refused() {
    let d = tempfile::tempdir().unwrap();
    let slots = d.path().join("slots");
    let out = run(acquire(&slots, "0", "x", &["true"]));
    assert_eq!(out.status.code(), Some(2), "{}", text(&out));
    let mut c = Command::new(EXE);
    c.args(["heavy-slot-acquire", "--max", "1", "--name", "x", "--dir"])
        .arg(&slots);
    let out = run(c);
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
        a.stdout(Stdio::piped()).stderr(Stdio::piped());
        let first = spawn(a);
        wait_for("holder started", || s1.exists());
        b.stdout(Stdio::piped()).stderr(Stdio::piped());
        let second = spawn(b);
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
        let (x, y) = (first.finish(), second.finish());
        assert!(x.status.success(), "{}", text(&x));
        assert!(y.status.success(), "{}", text(&y));
        assert!(s2.exists());
    }
}
