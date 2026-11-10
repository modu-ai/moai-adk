//! #2348 repair (PR #2382 CI): `common::feed_stdin` tolerates a child that exits before
//! reading its stdin, and only that. Each case is deterministic: the child has already
//! exited (its read end is closed) before the write starts, so no timing is involved.
#![cfg(unix)]

mod common;

use std::io::{self, Write as _};
use std::process::{Child, ChildStdin, Command, Stdio};

/// A child that never reads its stdin, already reaped, and the write end of its stdin.
fn exited_child() -> (Child, ChildStdin) {
    let mut child = Command::new("sh")
        .args(["-c", "exit 3"])
        .stdin(Stdio::piped())
        .spawn()
        .unwrap();
    let pipe = child.stdin.take().unwrap();
    assert_eq!(child.wait().unwrap().code(), Some(3));
    (child, pipe)
}

/// The precondition the helper relies on: writing to a child that exited without reading
/// fails with exactly `BrokenPipe` (SIGPIPE is ignored in Rust programs, so it is an error,
/// not a signal). 1 MiB is larger than any pipe buffer, so the write cannot just land there.
#[test]
fn writing_to_an_exited_child_is_a_broken_pipe() {
    let (_child, mut pipe) = exited_child();
    let err = pipe.write_all(&vec![b'x'; 1 << 20]).unwrap_err();
    assert_eq!(err.kind(), io::ErrorKind::BrokenPipe, "{err}");
}

#[test]
fn a_child_that_exits_without_reading_is_tolerated() {
    let (_child, pipe) = exited_child();
    common::write_stdin(pipe, &vec![b'x'; 1 << 20]);
}

#[test]
fn feed_stdin_tolerates_a_child_that_already_exited() {
    let mut child = Command::new("sh")
        .args(["-c", "exit 3"])
        .stdin(Stdio::piped())
        .spawn()
        .unwrap();
    // Wait without closing stdin: poll until the child is gone, then feed it.
    while child.try_wait().unwrap().is_none() {
        std::thread::sleep(std::time::Duration::from_millis(5));
    }
    common::feed_stdin(&mut child, &vec![b'x'; 1 << 20]);
    assert!(child.stdin.is_none(), "feed_stdin closes the pipe");
    assert_eq!(child.wait().unwrap().code(), Some(3));
}

#[test]
#[should_panic(expected = "stdin write failed")]
fn any_other_write_error_still_fails() {
    common::stdin_write_outcome(Err(io::Error::from(io::ErrorKind::PermissionDenied)));
}

#[test]
fn a_reading_child_receives_every_byte() {
    let mut child = Command::new("wc")
        .arg("-c")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, &vec![b'x'; 300_000]);
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success());
    assert_eq!(String::from_utf8_lossy(&out.stdout).trim(), "300000");
}
