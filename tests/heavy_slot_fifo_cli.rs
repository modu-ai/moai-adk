//! SPEC-20261122-check-affected-one-slot AC-5 (REQ-4): `heavy-slot-acquire` serves waiters in
//! arrival order. A newcomer never passes a live waiter that arrived earlier, even with a free
//! slot; a dead waiter's unlocked file is not counted. Every case uses its own temporary slot
//! folder and a temporary HOME; waits are bounded by progress markers plus a hang guard.
#![cfg(unix)]
mod common;

use std::fs;
use std::os::unix::process::CommandExt;
use std::path::Path;
use std::process::{Child, Command, Output, Stdio};
use std::time::{Duration, Instant};

const EXE: &str = crate::common::MOAI_BIN;
const HANG: Duration = Duration::from_secs(120);
/// Roomy machine numbers (debug-only seam), so the real disk and memory never decide.
const ROOMY: &str = "free_gb=1000,swap_pct=0";

/// A wrapper in a process group of its own, killed with its command on drop.
struct Spawned {
    child: Option<Child>,
    group: i32,
}

impl Spawned {
    /// Waits for the wrapper (bounded by the hang guard) and returns its output.
    fn finish(mut self) -> Output {
        let mut child = self.child.take().unwrap();
        let start = Instant::now();
        while child.try_wait().unwrap().is_none() {
            if start.elapsed() > HANG {
                kill_group(self.group);
                let _ = child.wait();
                panic!("hang guard: the wrapper did not end");
            }
            std::thread::sleep(Duration::from_millis(20));
        }
        child.wait_with_output().unwrap()
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

fn kill_group(group: i32) {
    // SAFETY: kill(2) on a process group this test created; ESRCH when it is already gone.
    unsafe { libc::kill(-group, libc::SIGKILL) };
}

/// `heavy-slot-acquire --max <max> --name <name> --dir <dir> -- sh -c <script>`, its stdout
/// going to `stdout` (a file the test reads while it runs).
fn acquire(dir: &Path, home: &Path, max: &str, name: &str, script: &str, stdout: &Path) -> Spawned {
    let mut c = Command::new(EXE);
    c.env("MOAI_CLI_TEST_MACHINE", ROOMY)
        .env("HOME", home)
        .env_remove("XDG_CACHE_HOME")
        .env_remove("MOAI_HEAVY_SLOT_FD")
        .args(["heavy-slot-acquire", "--max", max, "--name", name, "--dir"])
        .arg(dir)
        .args(["--", "sh", "-c", script])
        .stdin(Stdio::null())
        .stdout(fs::File::create(stdout).unwrap())
        .stderr(Stdio::piped())
        .process_group(0);
    let child = c.spawn().unwrap();
    let group = child.id() as i32;
    Spawned {
        child: Some(child),
        group,
    }
}

/// Waits (bounded by the hang guard) until `cond` holds.
fn wait_for(what: &str, mut cond: impl FnMut() -> bool) {
    let start = Instant::now();
    while !cond() {
        assert!(start.elapsed() < HANG, "hang guard: {what}");
        std::thread::sleep(Duration::from_millis(20));
    }
}

fn text(p: &Path) -> String {
    fs::read_to_string(p).unwrap_or_default()
}

/// A waiting entry named as a waiter names its own (`<ns, 20 digits>-<pid>`), older than any
/// real one.
fn old_waiter(dir: &Path, ns: u64) -> std::path::PathBuf {
    let waiting = dir.join("waiting");
    fs::create_dir_all(&waiting).unwrap();
    waiting.join(format!("{ns:020}-{}", std::process::id()))
}

#[test]
fn a_newcomer_waits_behind_an_earlier_live_waiter() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let (ran1, ran2) = (m.path().join("ran1"), m.path().join("ran2"));
    let (out1, out2) = (m.path().join("out1"), m.path().join("out2"));

    // A dead waiter (an unlocked file) that arrived first is not counted: the free slot is
    // taken at once.
    fs::write(old_waiter(d.path(), 1), "").unwrap();
    let first = acquire(
        d.path(),
        home.path(),
        "1",
        "first",
        &format!("touch '{}'", ran1.display()),
        &out1,
    );
    let o = first.finish();
    assert!(o.status.success(), "{o:?}");
    assert!(ran1.exists());
    assert!(
        !text(&out1).contains("heavy-slot: waiting"),
        "{}",
        text(&out1)
    );

    // A live waiter (its file locked) that arrived earlier: the newcomer waits behind it
    // although the one slot is free.
    let live_path = old_waiter(d.path(), 2);
    let live = fs::File::create(&live_path).unwrap();
    live.lock().unwrap();
    let newcomer = acquire(
        d.path(),
        home.path(),
        "1",
        "newcomer",
        &format!("touch '{}'", ran2.display()),
        &out2,
    );
    wait_for("the newcomer waits or runs", || {
        ran2.exists() || text(&out2).contains("heavy-slot: waiting")
    });
    assert!(!ran2.exists(), "the newcomer passed an earlier live waiter");
    assert!(
        text(&out2).contains("heavy-slot: waiting 2 of 2 for 1 slot(s)"),
        "{}",
        text(&out2)
    );
    // Several polls later it still waits.
    std::thread::sleep(Duration::from_millis(800));
    assert!(!ran2.exists(), "the newcomer passed an earlier live waiter");

    // The earlier waiter dies (its lock goes, its file stays): the newcomer starts.
    drop(live);
    let o = newcomer.finish();
    assert!(o.status.success(), "{o:?}");
    assert!(ran2.exists());
    assert!(live_path.exists(), "a dead waiter's file is left alone");
}

#[test]
fn two_waiters_start_in_arrival_order() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let (s1, r1, order) = (
        m.path().join("s1"),
        m.path().join("r1"),
        m.path().join("order"),
    );
    let holder = acquire(
        d.path(),
        home.path(),
        "1",
        "holder",
        &format!(
            "touch '{}'; while [ ! -e '{}' ]; do sleep 0.05; done",
            s1.display(),
            r1.display()
        ),
        &m.path().join("out0"),
    );
    wait_for("the holder started", || s1.exists());
    let (out1, out2) = (m.path().join("out1"), m.path().join("out2"));
    let w1 = acquire(
        d.path(),
        home.path(),
        "1",
        "w1",
        &format!("echo 1 >> '{}'", order.display()),
        &out1,
    );
    wait_for("w1 waiting", || text(&out1).contains("heavy-slot: waiting"));
    let w2 = acquire(
        d.path(),
        home.path(),
        "1",
        "w2",
        &format!("echo 2 >> '{}'", order.display()),
        &out2,
    );
    wait_for("w2 waiting", || text(&out2).contains("heavy-slot: waiting"));
    assert!(
        text(&out2).contains("heavy-slot: waiting 2 of 2 for 1 slot(s)"),
        "{}",
        text(&out2)
    );
    fs::write(&r1, "").unwrap();
    assert!(holder.finish().status.success());
    let (a, b) = (w1.finish(), w2.finish());
    assert!(a.status.success(), "{a:?}");
    assert!(b.status.success(), "{b:?}");
    assert_eq!(text(&order), "1\n2\n", "the waiters started out of order");
}
