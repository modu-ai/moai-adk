//! The marker helper for fake executables (SPEC-20261123-ready-mutant-line, REQ-8 and REQ-9):
//! the blind case of a `touch` fake under a PATH without `touch`, and the helper that does not
//! depend on PATH. Every run uses temporary folders; no real HOME is read.
#![cfg(unix)]

mod common;

use common::fake_exe::{marker_script, write_marking_executable};
use common::CleanEnv;
use std::path::Path;
use std::process::{Command, Stdio};

/// Runs `exe` with exactly `path_env` as its PATH under `.clean_env()`: of the rest of the
/// environment only `DEVELOPER_DIR` (when that folder exists) and the process-table switch stay.
fn run_with_path(exe: &Path, path_env: &Path) -> std::process::Output {
    Command::new(exe)
        .clean_env()
        .env("PATH", path_env)
        .stdin(Stdio::null())
        .output()
        .expect("start the fake")
}

/// The blind case: the fake ran (the shell could not find `touch`, exit 127), and its marker is
/// absent all the same. An assertion "the marker is absent" cannot tell this from a fake that
/// never ran.
#[test]
fn a_touch_fake_under_a_path_without_touch_ran_and_left_no_marker() {
    let dir = tempfile::tempdir().unwrap();
    let empty = dir.path().join("empty-bin");
    std::fs::create_dir(&empty).unwrap();
    let marker = dir.path().join("ran");
    let fake = dir.path().join("fake-touch");
    common::write_executable(&fake, &format!("#!/bin/sh\ntouch '{}'\n", marker.display()));
    let out = run_with_path(&fake, &empty);
    assert_eq!(out.status.code(), Some(127), "the shell found no `touch`");
    assert!(!marker.exists(), "the blind case: it ran, and no marker");
}

/// The helper's fake leaves its marker under the same PATH.
#[test]
fn the_helpers_fake_leaves_its_marker_when_path_holds_nothing() {
    let dir = tempfile::tempdir().unwrap();
    let empty = dir.path().join("empty-bin");
    std::fs::create_dir(&empty).unwrap();
    let marker = dir.path().join("ran");
    let fake = dir.path().join("fake-marking");
    write_marking_executable(&fake, &marker, "");
    let out = run_with_path(&fake, &empty);
    assert_eq!(out.status.code(), Some(0), "{out:?}");
    assert!(marker.exists(), "the helper's fake ran and left its marker");
}

/// The script text names no external command for the marker, and a name with a space or a quote
/// still lands in one file.
#[test]
fn the_marker_path_is_quoted_and_the_rest_of_the_script_runs() {
    let dir = tempfile::tempdir().unwrap();
    let empty = dir.path().join("empty-bin");
    std::fs::create_dir(&empty).unwrap();
    let marker = dir.path().join("it's a marker");
    let fake = dir.path().join("fake-rest");
    write_marking_executable(&fake, &marker, "printf '%s' after\n");
    let out = run_with_path(&fake, &empty);
    assert_eq!(out.status.code(), Some(0), "{out:?}");
    assert_eq!(String::from_utf8_lossy(&out.stdout), "after");
    assert!(marker.exists(), "{}", marker.display());
    let text = marker_script(&marker, "");
    assert!(!text.contains("touch"), "{text}");
}
