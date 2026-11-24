//! SPEC-20261122-check-affected-one-slot AC-5 (REQ-4): `heavy-slot-acquire` serves waiters in
//! arrival order. A newcomer never passes a live waiter that arrived earlier and can use the same
//! slot, even with a free slot; a dead waiter's unlocked file is not counted; with mixed `--max`
//! an earlier waiter keeps a newcomer off only the slots it can use itself. REQ-5: a nested
//! caller reuses a slot only through the descriptor its holder handed down. Every case uses its
//! own temporary slot folder and a temporary HOME; waits are bounded by progress markers plus a
//! hang guard.
#![cfg(unix)]
mod common;

use std::fs;
use std::os::unix::io::{AsRawFd, RawFd};
use std::os::unix::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Output, Stdio};
use std::time::{Duration, Instant};

const EXE: &str = crate::common::MOAI_BIN;
const HANG: Duration = Duration::from_secs(120);
/// How long a caller that should start at once may take before the case fails (it takes one
/// look of 200 ms at most when the rule holds).
const AT_ONCE: Duration = Duration::from_secs(20);
/// Roomy machine numbers (debug-only seam), so the real disk and memory never decide.
const ROOMY: &str = "free_gb=1000,swap_pct=0";
/// The descriptor a nested caller is handed.
const CHILD_FD: RawFd = 7;

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

    /// Stops (SIGSTOP) or resumes (SIGCONT) the wrapper itself.
    fn signal(&self, sig: libc::c_int) {
        // SAFETY: kill(2) on the wrapper this test started.
        assert_eq!(unsafe { libc::kill(self.group, sig) }, 0);
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

fn spawn(mut c: Command, stdout: &Path) -> Spawned {
    c.env("MOAI_TEST_MACHINE", ROOMY)
        .env_remove("XDG_CACHE_HOME")
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

/// `heavy-slot-acquire --max <max> --name <name> --dir <dir> -- sh -c <script>`, its stdout
/// going to `stdout` (a file the test reads while it runs).
fn acquire(dir: &Path, home: &Path, max: &str, name: &str, script: &str, stdout: &Path) -> Spawned {
    let mut c = Command::new(EXE);
    c.env("HOME", home)
        .env_remove("MOAI_HEAVY_SLOT_FD")
        .args(["heavy-slot-acquire", "--max", max, "--name", name, "--dir"])
        .arg(dir)
        .args(["--", "sh", "-c", script]);
    spawn(c, stdout)
}

/// A nested `heavy-slot-acquire --max 1` that finds `MOAI_HEAVY_SLOT_FD=<value>`. With `pass`,
/// that descriptor of this test is handed to the child alone as descriptor 7 (it shares this
/// test's open file description); without it, descriptor 7 is closed in the child.
fn nested(
    dir: &Path,
    home: &Path,
    value: &str,
    pass: Option<RawFd>,
    script: &str,
    stdout: &Path,
) -> Spawned {
    let mut c = Command::new(EXE);
    c.env("HOME", home)
        .env("MOAI_HEAVY_SLOT_FD", value)
        .args([
            "heavy-slot-acquire",
            "--max",
            "1",
            "--name",
            "nested",
            "--dir",
        ])
        .arg(dir)
        .args(["--", "sh", "-c", script]);
    // SAFETY: dup2, fcntl and close are async-signal-safe; nothing else runs before exec.
    unsafe {
        c.pre_exec(move || {
            match pass {
                Some(fd) if fd == CHILD_FD => {
                    if libc::fcntl(fd, libc::F_SETFD, 0) < 0 {
                        return Err(std::io::Error::last_os_error());
                    }
                }
                Some(fd) => {
                    if libc::dup2(fd, CHILD_FD) < 0 {
                        return Err(std::io::Error::last_os_error());
                    }
                }
                None => {
                    libc::close(CHILD_FD);
                }
            }
            Ok(())
        });
    }
    spawn(c, stdout)
}

/// Waits (bounded by the hang guard) until `cond` holds.
fn wait_for(what: &str, cond: impl FnMut() -> bool) {
    wait_within(HANG, &format!("hang guard: {what}"), cond);
}

/// Waits until `cond` holds; fails with `why` after `limit`.
fn wait_within(limit: Duration, why: &str, mut cond: impl FnMut() -> bool) {
    let start = Instant::now();
    while !cond() {
        assert!(start.elapsed() < limit, "{why}");
        std::thread::sleep(Duration::from_millis(20));
    }
}

fn text(p: &Path) -> String {
    fs::read_to_string(p).unwrap_or_default()
}

fn stderr(o: &Output) -> String {
    String::from_utf8_lossy(&o.stderr).into_owned()
}

/// A waiting entry named as a waiter of an older build or the Python twin names its own
/// (`<ns, 20 digits>-<pid>`, no max), older than any real one.
fn old_waiter(dir: &Path, ns: u64) -> PathBuf {
    let waiting = dir.join("waiting");
    fs::create_dir_all(&waiting).unwrap();
    waiting.join(format!("{ns:020}-{}", std::process::id()))
}

/// A live waiting entry that names its `--max` (`<ns>-<pid>-<max>`), its lock held by this test
/// until the returned file is dropped.
fn live_waiter(dir: &Path, ns: u64, max: usize) -> fs::File {
    let path = old_waiter(dir, ns).with_file_name(format!("{ns:020}-{}-{max}", std::process::id()));
    let f = fs::File::create(path).unwrap();
    f.lock().unwrap();
    f
}

/// Slot `i` of `dir` held by this test (a process apart from every wrapper).
fn hold_slot(dir: &Path, i: usize) -> fs::File {
    let f = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(dir.join(format!("slot-{i}")))
        .unwrap();
    f.lock().unwrap();
    f
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

/// Two waiters start in the order they arrived. The first is stopped (SIGSTOP: alive, its
/// waiting file still locked) when the slot frees, so only the order rule keeps the second, which
/// keeps looking, from taking it; once the first goes on, it starts first.
#[test]
fn two_waiters_start_in_arrival_order() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let order = m.path().join("order");
    let slot0 = hold_slot(d.path(), 0);
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
    w1.signal(libc::SIGSTOP);
    drop(slot0);
    // Five looks of w2 at a free slot it must leave to the stopped, earlier w1.
    std::thread::sleep(Duration::from_millis(1000));
    assert_eq!(text(&order), "", "w2 passed the earlier waiter w1");
    w1.signal(libc::SIGCONT);
    let (a, b) = (w1.finish(), w2.finish());
    assert!(a.status.success(), "{a:?}");
    assert!(b.status.success(), "{b:?}");
    assert_eq!(text(&order), "1\n2\n", "the waiters started out of order");
}

/// Mixed `--max` (the standing setup: callers with 1 to 5 slots share one folder). Slot 0 is
/// held by a third process. A narrow waiter (`--max 1`) that arrived first cannot use slot 1, so
/// a wide newcomer (`--max 2`) takes slot 1 at once, without a waiting line. When slot 0 frees,
/// the narrow waiter takes it before a narrow one that arrived later (the first is stopped then,
/// so only the order rule holds the later one back).
#[test]
fn a_wider_newcomer_takes_the_slot_a_narrower_earlier_waiter_cannot_use() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let order = m.path().join("order");
    let slot0 = hold_slot(d.path(), 0);
    let (out_n1, out_w, out_n2) = (m.path().join("n1"), m.path().join("w"), m.path().join("n2"));
    let n1 = acquire(
        d.path(),
        home.path(),
        "1",
        "narrow1",
        &format!("echo n1 >> '{}'", order.display()),
        &out_n1,
    );
    wait_for("narrow1 waiting", || {
        text(&out_n1).contains("heavy-slot: waiting")
    });
    let (ws, wr) = (m.path().join("ws"), m.path().join("wr"));
    let wide = acquire(
        d.path(),
        home.path(),
        "2",
        "wide",
        &format!(
            "touch '{}'; while [ ! -e '{}' ]; do sleep 0.05; done",
            ws.display(),
            wr.display()
        ),
        &out_w,
    );
    wait_within(
        AT_ONCE,
        "slot 1 stayed idle: the wide caller waited behind a narrow waiter that cannot use it",
        || ws.exists(),
    );
    assert!(
        !text(&out_w).contains("heavy-slot: waiting"),
        "the wide caller waited: {}",
        text(&out_w)
    );
    assert_eq!(text(&order), "", "narrow1 ran while slot 0 was held");

    let n2 = acquire(
        d.path(),
        home.path(),
        "1",
        "narrow2",
        &format!("echo n2 >> '{}'", order.display()),
        &out_n2,
    );
    wait_for("narrow2 waiting", || {
        text(&out_n2).contains("heavy-slot: waiting 2 of 2 for 1 slot(s)")
    });
    n1.signal(libc::SIGSTOP);
    drop(slot0);
    std::thread::sleep(Duration::from_millis(1000));
    assert_eq!(text(&order), "", "narrow2 passed the earlier narrow1");
    n1.signal(libc::SIGCONT);
    let (a, b) = (n1.finish(), n2.finish());
    assert!(a.status.success(), "{a:?}");
    assert!(b.status.success(), "{b:?}");
    assert_eq!(
        text(&order),
        "n1\nn2\n",
        "the narrow waiters started out of order"
    );
    assert!(
        stderr(&a).contains("slot 0 taken by narrow1"),
        "{}",
        stderr(&a)
    );
    fs::write(&wr, "").unwrap();
    let w = wide.finish();
    assert!(w.status.success(), "{w:?}");
    assert!(
        stderr(&w).contains("slot 1 taken by wide"),
        "{}",
        stderr(&w)
    );
}

/// The waiting file names its waiter's `--max`, and an earlier live waiter counts only for the
/// slots it can use: one with `--max 1` leaves the free slot 1 to a `--max 2` newcomer, one
/// with `--max 2` keeps it.
#[test]
fn an_earlier_waiter_counts_only_for_the_slots_it_can_use() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let _slot0 = hold_slot(d.path(), 0);
    let (ran_a, out_a) = (m.path().join("ran_a"), m.path().join("out_a"));
    let narrow = live_waiter(d.path(), 1, 1);
    let a = acquire(
        d.path(),
        home.path(),
        "2",
        "a",
        &format!("touch '{}'", ran_a.display()),
        &out_a,
    );
    wait_within(
        AT_ONCE,
        "slot 1 stayed idle behind a waiter that can use slot 0 only",
        || ran_a.exists(),
    );
    let o = a.finish();
    assert!(o.status.success(), "{o:?}");
    assert!(stderr(&o).contains("slot 1 taken by a"), "{}", stderr(&o));
    assert!(
        !text(&out_a).contains("heavy-slot: waiting"),
        "{}",
        text(&out_a)
    );
    drop(narrow);

    let (ran_b, out_b) = (m.path().join("ran_b"), m.path().join("out_b"));
    let wide = live_waiter(d.path(), 2, 2);
    let b = acquire(
        d.path(),
        home.path(),
        "2",
        "b",
        &format!("touch '{}'", ran_b.display()),
        &out_b,
    );
    wait_for("b waits or runs", || {
        ran_b.exists() || text(&out_b).contains("heavy-slot: waiting")
    });
    std::thread::sleep(Duration::from_millis(600));
    assert!(
        !ran_b.exists(),
        "b passed an earlier waiter that can use slot 1"
    );
    drop(wide);
    let o = b.finish();
    assert!(o.status.success(), "{o:?}");
    assert!(stderr(&o).contains("slot 1 taken by b"), "{}", stderr(&o));
}

/// REQ-5, the reuse direction: the descriptor of the process that holds slot 0, handed down,
/// is reused at once.
#[test]
fn the_holders_own_descriptor_is_reused() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slot0 = hold_slot(d.path(), 0);
    let (ran, out) = (m.path().join("ran"), m.path().join("out"));
    let value = format!("{CHILD_FD}:{}", d.path().join("slot-0").display());
    let child = nested(
        d.path(),
        home.path(),
        &value,
        Some(slot0.as_raw_fd()),
        &format!("touch '{}'", ran.display()),
        &out,
    );
    wait_within(AT_ONCE, "the inherited slot was not reused", || {
        ran.exists()
    });
    let o = child.finish();
    assert!(o.status.success(), "{o:?}");
    assert!(
        stderr(&o).contains("heavy-slot: nested reuses slot 0 it inherited"),
        "{}",
        stderr(&o)
    );
    assert!(
        !text(&out).contains("heavy-slot: waiting"),
        "{}",
        text(&out)
    );
}

/// REQ-5, the refuse direction: a nested caller whose variable does not lead to the holder's own
/// descriptor queues for slot 0 while this test holds it, then takes it the ordinary way.
fn queues_until_slot0_frees(slot0: fs::File, child: Spawned, out: &Path, ran: &Path) {
    wait_for("the nested caller waits or runs", || {
        ran.exists() || text(out).contains("heavy-slot: waiting")
    });
    assert!(
        !ran.exists(),
        "it ran while another process held slot 0: {}",
        text(out)
    );
    // Let go of the lock itself (the description may be shared with the child).
    slot0.unlock().unwrap();
    drop(slot0);
    let o = child.finish();
    assert!(o.status.success(), "{o:?}");
    assert!(ran.exists());
    let e = stderr(&o);
    assert!(e.contains("slot 0 taken by nested"), "{e}");
    assert!(!e.contains("reuses"), "{e}");
}

#[test]
fn a_descriptor_on_another_file_is_not_reused() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slot0 = hold_slot(d.path(), 0);
    let other = fs::File::create(d.path().join("other")).unwrap();
    let (ran, out) = (m.path().join("ran"), m.path().join("out"));
    let value = format!("{CHILD_FD}:{}", d.path().join("slot-0").display());
    let script = format!("touch '{}'", ran.display());
    let child = nested(
        d.path(),
        home.path(),
        &value,
        Some(other.as_raw_fd()),
        &script,
        &out,
    );
    queues_until_slot0_frees(slot0, child, &out, &ran);
}

#[test]
fn a_closed_descriptor_is_not_reused() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slot0 = hold_slot(d.path(), 0);
    let (ran, out) = (m.path().join("ran"), m.path().join("out"));
    let value = format!("{CHILD_FD}:{}", d.path().join("slot-0").display());
    let script = format!("touch '{}'", ran.display());
    let child = nested(d.path(), home.path(), &value, None, &script, &out);
    queues_until_slot0_frees(slot0, child, &out, &ran);
}

/// A held file that is no `slot-<digits>` (its descriptor handed down, so every other step would
/// pass) is not a slot.
#[test]
fn a_held_file_that_is_no_slot_is_not_reused() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slot0 = hold_slot(d.path(), 0);
    let held = fs::File::create(d.path().join("slot-x0")).unwrap();
    held.lock().unwrap();
    let (ran, out) = (m.path().join("ran"), m.path().join("out"));
    let value = format!("{CHILD_FD}:{}", d.path().join("slot-x0").display());
    let script = format!("touch '{}'", ran.display());
    let child = nested(
        d.path(),
        home.path(),
        &value,
        Some(held.as_raw_fd()),
        &script,
        &out,
    );
    queues_until_slot0_frees(slot0, child, &out, &ran);
}

/// The variable names a slot of another folder: not this folder's slot, even with the holder's
/// own descriptor of this folder's slot 0 handed down.
#[test]
fn a_slot_named_in_another_folder_is_not_reused() {
    let d = tempfile::tempdir().unwrap();
    let e = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slot0 = hold_slot(d.path(), 0);
    fs::write(e.path().join("slot-0"), "").unwrap();
    let (ran, out) = (m.path().join("ran"), m.path().join("out"));
    let value = format!("{CHILD_FD}:{}", e.path().join("slot-0").display());
    let script = format!("touch '{}'", ran.display());
    let child = nested(
        d.path(),
        home.path(),
        &value,
        Some(slot0.as_raw_fd()),
        &script,
        &out,
    );
    queues_until_slot0_frees(slot0, child, &out, &ran);
}

/// A caller that opens slot 0 itself and names that descriptor, while another process holds the
/// lock: same file, slot held, but not the holder's open file description — it queues.
#[test]
fn a_slot_file_opened_anew_is_not_reused() {
    let d = tempfile::tempdir().unwrap();
    let m = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slot0 = hold_slot(d.path(), 0);
    let (ran, out) = (m.path().join("ran"), m.path().join("out"));
    let slot_path = d.path().join("slot-0");
    let mut c = Command::new("sh");
    c.env("HOME", home.path())
        .env(
            "MOAI_HEAVY_SLOT_FD",
            format!("{CHILD_FD}:{}", slot_path.display()),
        )
        .args([
            "-c",
            "exec 7<>\"$1\"; exec \"$2\" heavy-slot-acquire --max 1 --name nested --dir \"$3\" -- sh -c \"$4\"",
            "sh",
        ])
        .arg(&slot_path)
        .arg(EXE)
        .arg(d.path())
        .arg(format!("touch '{}'", ran.display()));
    let child = spawn(c, &out);
    queues_until_slot0_frees(slot0, child, &out, &ran);
}
