//! SPEC-20261123-heavy-slot-leftover (AC-1 to AC-4): `heavy-slot-acquire` gives every command it
//! starts a `CARGO_BUILD_JOBS` cap, `heavy-slot-status` prints the cap and counts the waiting
//! files of dead callers without removing any, and a caller asking for a slot removes exactly
//! those files. Every case uses its own temporary slot folder and a temporary HOME; the real
//! machine's slots are never touched.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::fs;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const EXE: &str = crate::common::MOAI_BIN;
/// Roomy machine numbers (debug-only seam), so the real disk and memory never decide.
const ROOMY: &str = "free_gb=1000,swap_pct=0";

fn cpus() -> usize {
    std::thread::available_parallelism().map_or(1, |n| n.get())
}

/// The cap rule written out again, so the test does not repeat the product's own function.
fn expected_jobs(slots: usize) -> usize {
    (cpus() / slots).max(2)
}

/// `heavy-slot-acquire --max <max> --dir <dir> -- <cmd…>` with no `CARGO_BUILD_JOBS` and no
/// inherited slot in its environment.
fn acquire(dir: &Path, home: &Path, max: &str, cmd: &[&str]) -> Command {
    let mut c = Command::new(EXE);
    c.env("HOME", home)
        .env("MOAI_TEST_MACHINE", ROOMY)
        .env_remove("MOAI_HEAVY_SLOT_FD")
        .env_remove("CARGO_BUILD_JOBS")
        .args(["heavy-slot-acquire", "--max", max, "--name", "t", "--dir"])
        .arg(dir)
        .arg("--")
        .args(cmd);
    c
}

fn ran_ok(out: &Output) {
    assert!(out.status.success(), "{out:?}");
}

fn status(dir: &Path) -> Value {
    let out = Command::new(EXE)
        .args(["heavy-slot-status", "--dir"])
        .arg(dir)
        .output()
        .unwrap();
    ran_ok(&out);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["operation"], "heavy_slot.status", "{v}");
    v["result"].clone()
}

/// A pid no process has: above any `pid_max`, and checked with `kill(pid, 0)`. A reaped child's
/// pid could be reused by an unrelated process before the case runs.
fn dead_pid() -> u32 {
    let pid = i32::MAX as u32 - 7;
    // SAFETY: signal 0 only asks whether the pid exists.
    let r = unsafe { libc::kill(pid as libc::pid_t, 0) };
    assert_eq!(r, -1, "fixture: pid {pid} must not exist");
    assert_eq!(
        std::io::Error::last_os_error().raw_os_error(),
        Some(libc::ESRCH)
    );
    pid
}

fn read_trim(p: &Path) -> String {
    fs::read_to_string(p).unwrap_or_default().trim().to_string()
}

#[test]
fn acquire_caps_cargo_build_jobs_and_a_given_value_wins() {
    let d = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slots = d.path().join("slots");

    // No value from the caller: the cap of the rule, for `--max 2`.
    let seen = d.path().join("seen");
    let script = format!("echo \"$CARGO_BUILD_JOBS\" > '{}'", seen.display());
    ran_ok(
        &acquire(&slots, home.path(), "2", &["sh", "-c", &script])
            .output()
            .unwrap(),
    );
    assert_eq!(read_trim(&seen), expected_jobs(2).to_string());

    // An empty value is no value: cargo cannot read it, so the cap replaces it.
    let empty = d.path().join("empty");
    let script = format!("echo \"$CARGO_BUILD_JOBS\" > '{}'", empty.display());
    ran_ok(
        &acquire(&slots, home.path(), "2", &["sh", "-c", &script])
            .env("CARGO_BUILD_JOBS", "")
            .output()
            .unwrap(),
    );
    assert_eq!(read_trim(&empty), expected_jobs(2).to_string());

    // A value from the caller passes through unchanged.
    let given = d.path().join("given");
    let script = format!("echo \"$CARGO_BUILD_JOBS\" > '{}'", given.display());
    ran_ok(
        &acquire(&slots, home.path(), "2", &["sh", "-c", &script])
            .env("CARGO_BUILD_JOBS", "7")
            .output()
            .unwrap(),
    );
    assert_eq!(read_trim(&given), "7");

    // A nested caller that reuses the outer slot keeps the outer value, whatever `--max` it names.
    let inner = d.path().join("inner");
    let nested = format!(
        "'{exe}' heavy-slot-acquire --max 4 --name inner --dir '{dir}' -- sh -c 'echo \"$CARGO_BUILD_JOBS\" > \"{out}\"'",
        exe = EXE,
        dir = slots.display(),
        out = inner.display()
    );
    ran_ok(
        &acquire(&slots, home.path(), "1", &["sh", "-c", &nested])
            .output()
            .unwrap(),
    );
    assert_eq!(read_trim(&inner), expected_jobs(1).to_string());
}

#[test]
fn status_prints_the_jobs_cap() {
    let d = tempfile::tempdir().unwrap();

    // No limits file: the default slot count, 3.
    let s = status(d.path());
    assert_eq!(
        s["build_jobs_cap"],
        serde_json::json!({"slots": 3, "cpus": cpus(), "jobs": expected_jobs(3)}),
        "{s}"
    );

    // A limits file gives its `max`.
    fs::write(d.path().join("limits.json"), "{\"max\": 4}").unwrap();
    let s = status(d.path());
    assert_eq!(
        s["build_jobs_cap"],
        serde_json::json!({"slots": 4, "cpus": cpus(), "jobs": expected_jobs(4)}),
        "{s}"
    );

    // A limits file that cannot be used gives no cap.
    fs::write(d.path().join("limits.json"), "{ not json").unwrap();
    let s = status(d.path());
    assert_eq!(s["build_jobs_cap"], Value::Null, "{s}");
}

/// The waiting folder of one case: what each name is, and what must happen to it.
struct Waiting {
    dir: PathBuf,
    /// Names the next caller removes: a dead pid, nobody holding the file.
    stale: Vec<String>,
    /// Names that stay.
    kept: Vec<String>,
    /// Locks held by this test until it is dropped.
    _locked: Vec<fs::File>,
}

fn waiting_folder(slots: &Path) -> Waiting {
    let dir = slots.join("waiting");
    fs::create_dir_all(&dir).unwrap();
    let dead = dead_pid();
    let me = std::process::id();
    let mut stale = Vec::new();
    let mut kept = Vec::new();
    let mut locked = Vec::new();

    // Dead pid, no process holds it: with and without a `--max` at the end.
    for name in [format!("{:020}-{dead}-2", 1), format!("{:020}-{dead}", 2)] {
        fs::write(dir.join(&name), "").unwrap();
        stale.push(name);
    }
    // Held by a process (this test), though the name's pid is dead: a live waiter stays.
    let held = format!("{:020}-{dead}-1", 3);
    let f = fs::File::create(dir.join(&held)).unwrap();
    f.lock().unwrap();
    locked.push(f);
    kept.push(held);
    // A live pid and nobody holding the file: left alone.
    let live = format!("{:020}-{me}-2", 4);
    fs::write(dir.join(&live), "").unwrap();
    kept.push(live);
    // A symlink and a folder with the name of a dead waiter: not plain files.
    let link_target = slots.join("link-target");
    fs::write(&link_target, "keep me").unwrap();
    let link = format!("{:020}-{dead}-2", 5);
    symlink(&link_target, dir.join(&link)).unwrap();
    kept.push(link);
    let folder = format!("{:020}-{dead}-3", 6);
    fs::create_dir(dir.join(&folder)).unwrap();
    kept.push(folder);
    // Names that are not `<digits>-<digits>[-<digits>]`.
    for name in ["notes.txt", "12-abc-2", "1-2-3-4", "-7-2"] {
        fs::write(dir.join(name), "").unwrap();
        kept.push(name.to_string());
    }
    // A dead waiter's name that is a second hard link of a file elsewhere: one write or removal
    // would reach the other name, so it is not a plain file of its own and stays.
    let hard = format!("{:020}-{dead}-2", 7);
    fs::hard_link(&link_target, dir.join(&hard)).unwrap();
    kept.push(hard);
    // A dead waiter's file nobody can open (mode 000): it cannot be judged, so it stays. A
    // superuser opens it anyway, so the case is left out there.
    // SAFETY: geteuid has no failure mode.
    if unsafe { libc::geteuid() } != 0 {
        use std::os::unix::fs::PermissionsExt;
        let closed = format!("{:020}-{dead}-2", 8);
        fs::write(dir.join(&closed), "").unwrap();
        fs::set_permissions(dir.join(&closed), fs::Permissions::from_mode(0o000)).unwrap();
        kept.push(closed);
    }
    stale.sort();
    kept.sort();
    Waiting {
        dir,
        stale,
        kept,
        _locked: locked,
    }
}

fn names(dir: &Path) -> Vec<String> {
    let mut v: Vec<String> = fs::read_dir(dir)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    v.sort();
    v
}

#[test]
fn a_caller_removes_only_dead_unlocked_waiting_files() {
    let d = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slots = d.path().join("slots");
    let w = waiting_folder(&slots);
    let mut all = [w.stale.clone(), w.kept.clone()].concat();
    all.sort();
    assert_eq!(names(&w.dir), all, "fixture");

    ran_ok(
        &acquire(&slots, home.path(), "3", &["true"])
            .output()
            .unwrap(),
    );

    assert_eq!(names(&w.dir), w.kept, "only the stale names are gone");
    assert_eq!(
        fs::read_to_string(slots.join("link-target")).unwrap(),
        "keep me",
        "the symlink's target is untouched"
    );
}

#[test]
fn status_counts_stale_waiting_files_and_removes_none() {
    let d = tempfile::tempdir().unwrap();
    let slots = d.path().join("slots");
    let w = waiting_folder(&slots);
    let before = names(&w.dir);

    let s = status(&slots);
    assert_eq!(s["stale_waiting"], w.stale.len(), "{s}");
    assert_eq!(s["writes"], false, "{s}");
    assert_eq!(names(&w.dir), before, "status removes nothing");

    // No waiting folder at all counts zero.
    let none = tempfile::tempdir().unwrap();
    assert_eq!(status(none.path())["stale_waiting"], 0);
}

/// B1 of the audit: `<slot folder>/waiting` that is a symlink to another folder. The cleanup
/// listed and removed the dead-named files of that other folder. Now the waiting path must be a
/// real folder of its own: a link there is refused before anything runs, nothing outside the
/// slot folder is created, listed or removed, and `status` counts zero.
#[test]
fn a_symlinked_waiting_folder_is_refused_and_nothing_outside_is_touched() {
    let d = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    let slots = d.path().join("slots");
    fs::create_dir_all(&slots).unwrap();
    let outside = d.path().join("outside");
    fs::create_dir_all(&outside).unwrap();
    let dead = dead_pid();
    // Outside files named like a dead waiter's and one more that must not be created.
    let victim = format!("{:020}-{dead}-2", 1);
    fs::write(outside.join(&victim), "victim").unwrap();
    fs::write(outside.join("other"), "other").unwrap();
    symlink(&outside, slots.join("waiting")).unwrap();
    let before = names(&outside);

    // `status` reads: the link is not a waiting folder, so it counts nothing and removes nothing.
    let st = status(&slots);
    assert_eq!(st["stale_waiting"], 0, "{st}");
    assert_eq!(names(&outside), before, "status touched the outside folder");

    // `acquire` refuses (125), COMMAND does not run, and the outside folder is as it was.
    let ran = d.path().join("ran");
    let script = format!("echo ran > '{}'", ran.display());
    let out = acquire(&slots, home.path(), "3", &["sh", "-c", &script])
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(125), "{out:?}");
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(
        err.contains("heavy-slot: refused") && err.contains("waiting"),
        "{err}"
    );
    assert!(!ran.exists(), "COMMAND must not run");
    assert_eq!(
        names(&outside),
        before,
        "acquire touched the outside folder"
    );
    assert_eq!(read_trim(&outside.join(&victim)), "victim");
    assert_eq!(read_trim(&outside.join("other")), "other");
    assert!(
        fs::symlink_metadata(slots.join("waiting"))
            .unwrap()
            .file_type()
            .is_symlink(),
        "the link itself is left for its owner"
    );
}

/// A `waiting` path that is a plain file, or a link that points nowhere, is refused the same
/// way and creates nothing behind it.
#[test]
fn a_waiting_path_that_is_a_file_or_a_dangling_link_is_refused() {
    let d = tempfile::tempdir().unwrap();
    let home = tempfile::tempdir().unwrap();
    for case in ["file", "dangling"] {
        let slots = d.path().join(case);
        fs::create_dir_all(&slots).unwrap();
        let gone = d.path().join(format!("{case}-gone"));
        if case == "file" {
            fs::write(slots.join("waiting"), "x").unwrap();
        } else {
            symlink(&gone, slots.join("waiting")).unwrap();
        }
        let out = acquire(&slots, home.path(), "3", &["true"])
            .output()
            .unwrap();
        assert_eq!(out.status.code(), Some(125), "{case}: {out:?}");
        assert!(!gone.exists(), "{case}: the link target was created");
    }
}
