//! #2283: why a test script must not be written by the test process itself (Linux only).
//!
//! A forked child keeps every fd of its parent until its own exec. When a parallel test thread
//! forks while this process holds a write fd on a freshly written script, the script cannot be
//! exec'd until that child execs: ETXTBSY ("executable file busy"). The first test pins that
//! mechanism deterministically; the second shows `common::write_executable` survives the same
//! concurrent forking. macOS does not refuse such an exec, so both are Linux-only.
#![cfg(target_os = "linux")]
mod common;

use std::{
    fs,
    io::Write,
    os::unix::{fs::PermissionsExt, process::CommandExt},
    process::Command,
    sync::{
        atomic::{AtomicBool, Ordering},
        Arc,
    },
    thread,
    time::Duration,
};

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

#[test]
fn inherited_writer_fd_makes_exec_fail_with_etxtbsy() {
    let tmp = tempfile::tempdir().unwrap();
    let path = tmp.path().join("stub");
    let mut file = fs::File::create(&path).unwrap();
    file.write_all(b"#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    // Another thread forks while the write fd is open; its child holds a copy.
    let forker = thread::spawn(|| fork_and_hold(Duration::from_millis(800)));
    thread::sleep(Duration::from_millis(200));
    drop(file);
    let err = Command::new(&path)
        .status()
        .expect_err("exec must fail while a forked child holds the writer fd");
    assert_eq!(err.kind(), std::io::ErrorKind::ExecutableFileBusy, "{err}");
    forker.join().unwrap();
    // Once that child has exec'd, the same script runs.
    assert!(Command::new(&path).status().unwrap().success());
}

#[test]
fn write_executable_survives_concurrent_forks() {
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
        common::write_executable(&path, "#!/bin/sh\nexit 0\n");
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
