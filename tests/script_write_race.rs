//! #2283: why a test script must not be written by the test process itself (Linux only).
//!
//! A forked child keeps every fd of its parent until its own exec. When a parallel test thread
//! forks while this process holds a write fd on a freshly written script, the script cannot be
//! exec'd until that child execs: ETXTBSY ("executable file busy"). The first test pins that
//! mechanism deterministically; the others show `common::write_executable` and
//! `common::copy_executable` survive the same concurrent forking. macOS does not refuse such an
//! exec, so all are Linux-only.
#![cfg(target_os = "linux")]
mod common;

use std::{
    fs,
    io::{PipeReader, PipeWriter, Read, Write},
    os::{
        fd::{AsRawFd, FromRawFd, OwnedFd, RawFd},
        unix::{fs::PermissionsExt, process::CommandExt},
    },
    process::Command,
    sync::{
        atomic::{AtomicBool, Ordering},
        Arc, PoisonError, RwLock,
    },
    thread,
    time::Duration,
};

/// The forking tests hold this shared, the ETXTBSY test exclusive (#2518). A child another
/// test thread forks while the ETXTBSY test's write fd is open keeps a copy until its own
/// exec, so in rc-ci (job 131198747457) the final exec there met ETXTBSY from such a child;
/// with the lock the only child holding that fd is the test's own handshake child.
static FORKS: RwLock<()> = RwLock::new(());

/// Forks a child that stays between fork and exec for `hold`, holding inherited fds.
fn fork_and_hold(hold: Duration) {
    let mut cmd = Command::new("true");
    // SAFETY: the closure only sleeps; it touches no lock or allocator state.
    unsafe {
        cmd.pre_exec(move || {
            thread::sleep(hold);
            Ok(())
        });
    }
    let _ = cmd.status();
}

/// Forks a child that, between fork and exec, reports on `forked` that it holds the inherited
/// fds and then waits for one byte on `release` (#2518: a pipe handshake instead of the old
/// 200 ms / 800 ms timing assumption). `release_writer` is the fd number of the parent's
/// release writer: the child closes its inherited copy, so a parent that dies without
/// releasing still lets the child's read end with EOF instead of blocking forever.
fn fork_and_hold_until_released(forked: PipeWriter, release: PipeReader, release_writer: RawFd) {
    let mut cmd = Command::new("true");
    let mut forked = Some(forked);
    let mut release = Some(release);
    // SAFETY: the closure only closes an inherited fd and does one pipe write and one pipe
    // read (plain syscalls); it takes no lock and does not allocate. `release_writer` is open
    // in the child because every parent fd is inherited at fork.
    unsafe {
        cmd.pre_exec(move || {
            drop(OwnedFd::from_raw_fd(release_writer));
            if let Some(mut w) = forked.take() {
                w.write_all(&[1])?;
            }
            if let Some(mut r) = release.take() {
                let mut byte = [0u8; 1];
                let _ = r.read(&mut byte);
            }
            Ok(())
        });
    }
    let _ = cmd.status();
}

/// Writes one byte on drop, so the held child is released even when an assertion fails.
struct Release(Option<PipeWriter>);
impl Drop for Release {
    fn drop(&mut self) {
        if let Some(mut w) = self.0.take() {
            let _ = w.write_all(&[1]);
        }
    }
}

#[test]
fn inherited_writer_fd_makes_exec_fail_with_etxtbsy() {
    let _alone = FORKS.write().unwrap_or_else(PoisonError::into_inner);
    let tmp = tempfile::tempdir().unwrap();
    let path = tmp.path().join("stub");
    let mut file = fs::File::create(&path).unwrap();
    file.write_all(b"#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    let (mut forked_r, forked_w) = std::io::pipe().unwrap();
    let (release_r, release_w) = std::io::pipe().unwrap();
    let release_fd = release_w.as_raw_fd();
    let release = Release(Some(release_w));
    // Another thread forks while the write fd is open; its child holds a copy.
    let forker =
        thread::spawn(move || fork_and_hold_until_released(forked_w, release_r, release_fd));
    // Wait until the child reports it has forked (and so holds the writer fd); no clock.
    let mut byte = [0u8; 1];
    forked_r.read_exact(&mut byte).unwrap();
    drop(file);
    let err = Command::new(&path)
        .status()
        .expect_err("exec must fail while a forked child holds the writer fd");
    assert_eq!(err.kind(), std::io::ErrorKind::ExecutableFileBusy, "{err}");
    // Let the child exec (dropping its copy of the fd), then wait for it.
    drop(release);
    forker.join().unwrap();
    // Once that child has exec'd, the same script runs.
    assert!(Command::new(&path).status().unwrap().success());
}

#[test]
fn write_executable_survives_concurrent_forks() {
    exec_survives_concurrent_forks(|path| {
        common::write_executable(path, "#!/bin/sh\nexit 0\n");
    });
}

#[test]
fn copy_executable_survives_concurrent_forks() {
    let src_dir = tempfile::tempdir().unwrap();
    let src = src_dir.path().join("source");
    common::write_executable(&src, "#!/bin/sh\nexit 0\n");
    exec_survives_concurrent_forks(|path| common::copy_executable(&src, path));
}

/// Makes and execs 300 executables with `make` while four threads keep forking children that
/// sit between fork and exec; every exec must succeed.
fn exec_survives_concurrent_forks(make: impl Fn(&std::path::Path)) {
    let _shared = FORKS.read().unwrap_or_else(PoisonError::into_inner);
    let tmp = tempfile::tempdir().unwrap();
    let stop = Arc::new(AtomicBool::new(false));
    let forkers: Vec<_> = (0..4)
        .map(|_| {
            let stop = Arc::clone(&stop);
            thread::spawn(move || {
                while !stop.load(Ordering::Relaxed) {
                    fork_and_hold(Duration::from_millis(5));
                }
            })
        })
        .collect();
    let mut failures = Vec::new();
    for i in 0..300 {
        let path = tmp.path().join(format!("stub-{i}"));
        make(&path);
        match Command::new(&path).status() {
            Ok(s) if s.success() => {}
            other => failures.push(format!("stub-{i}: {other:?}")),
        }
    }
    stop.store(true, Ordering::Relaxed);
    for f in forkers {
        f.join().unwrap();
    }
    assert!(
        failures.is_empty(),
        "{} failures: {:?}",
        failures.len(),
        &failures[..failures.len().min(5)]
    );
}
