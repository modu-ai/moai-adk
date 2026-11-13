//! #2791 AC-1: `bin-names-plan` / `bin-names-apply` make `bin/moai` a hard link to
//! `bin/moai-cli` in the version folder the binary runs from. Every case runs copies of the
//! built binaries inside a temporary HOME; the real HOME, PATH and shell files are never read.

#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::os::unix::fs::MetadataExt as _;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_CLI_BIN;
const MOAI: &str = common::MOAI_BIN;
const NOW: &str = "2026-11-13T21:00:00Z";

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`; returns (guard, home, exe).
fn layout() -> (TempDir, PathBuf, PathBuf) {
    let d = TempDir::new().unwrap();
    let home = fs::canonicalize(d.path()).unwrap();
    let bin = home.join(".moai-cli/versions/a/bin");
    fs::create_dir_all(&bin).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    (d, home, exe)
}

fn call(exe: &Path, home: &Path, args: &[&str]) -> Value {
    let out = Command::new(exe)
        .args(args)
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn plan(exe: &Path, home: &Path) -> Value {
    let v = call(
        exe,
        home,
        &["bin-names-plan", "--home", home.to_str().unwrap()],
    );
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn apply(exe: &Path, home: &Path, digest: &str, perm: &str) -> Value {
    call(
        exe,
        home,
        &[
            "bin-names-apply",
            "--home",
            home.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            perm,
            "--now",
            NOW,
        ],
    )
}

fn names(dir: &Path) -> Vec<(String, Vec<u8>)> {
    let mut v: Vec<_> = fs::read_dir(dir)
        .unwrap()
        .flatten()
        .map(|e| {
            let p = e.path();
            let bytes = fs::read(&p).unwrap_or_default();
            (e.file_name().to_string_lossy().into_owned(), bytes)
        })
        .collect();
    v.sort();
    v
}

fn same_file(a: &Path, b: &Path) -> bool {
    let (m1, m2) = (fs::metadata(a).unwrap(), fs::metadata(b).unwrap());
    (m1.dev(), m1.ino()) == (m2.dev(), m2.ino())
}

fn version_line(exe: &Path) -> String {
    let out = Command::new(exe).arg("--version").output().unwrap();
    String::from_utf8_lossy(&out.stdout)
        .lines()
        .next()
        .unwrap_or("")
        .to_string()
}

#[test]
fn create_links_the_alias_and_a_rerun_is_unchanged() {
    let (_d, home, exe) = layout();
    let bin = exe.parent().unwrap().to_path_buf();
    let alias = bin.join("moai");
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("create"), "{p}");
    let d = p["plan_digest"].as_str().unwrap().to_string();

    // Refusals write nothing.
    let before = names(&bin);
    for (digest, perm, why) in [
        (d.as_str(), "", "permission reference"),
        ("sha256:00", "approved", "stale or altered plan"),
    ] {
        let a = apply(&exe, &home, digest, perm);
        assert_eq!(a["ok"], json!(false), "{why}: {a}");
        assert!(a.to_string().contains(why), "{why}: {a}");
        assert_eq!(names(&bin), before, "{why}");
    }

    let a = apply(&exe, &home, &d, "approved");
    assert_eq!(a["ok"], json!(true), "{a}");
    assert_eq!(a["result"]["state"], json!("applied"), "{a}");
    assert!(same_file(&exe, &alias), "one file, two names");
    assert_eq!(version_line(&alias), version_line(&exe));
    assert!(version_line(&alias).ends_with(" (rust)"));

    // Planned from the `moai` name the same folder is found, and nothing is left to do.
    let again = plan(&alias, &home);
    assert_eq!(again["state"], json!("unchanged"), "{again}");
    let a = apply(
        &alias,
        &home,
        again["plan_digest"].as_str().unwrap(),
        "approved",
    );
    assert_eq!(a["result"]["written"], json!([]), "{a}");
}

#[test]
fn a_separate_copy_of_this_build_is_moved_aside_never_deleted() {
    let (_d, home, exe) = layout();
    let bin = exe.parent().unwrap().to_path_buf();
    let alias = bin.join("moai");
    common::copy_executable(Path::new(MOAI), &alias);
    let copy_bytes = fs::read(&alias).unwrap();
    assert_ne!(
        copy_bytes,
        fs::read(&exe).unwrap(),
        "two binaries of one build"
    );
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("replace_copy"), "{p}");

    // The aside name is taken: refused before anything moves.
    let aside = bin.join("moai.replaced-20261113T210000Z");
    fs::write(&aside, "taken").unwrap();
    let before = names(&bin);
    let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["ok"], json!(false), "{a}");
    assert!(a.to_string().contains("already exists"), "{a}");
    assert_eq!(names(&bin), before);
    fs::remove_file(&aside).unwrap();

    let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["ok"], json!(true), "{a}");
    assert_eq!(fs::read(&aside).unwrap(), copy_bytes, "the copy is kept");
    assert!(same_file(&exe, &alias));
}

#[test]
fn another_program_a_link_or_a_folder_is_left_alone() {
    let (_d, home, exe) = layout();
    let bin = exe.parent().unwrap().to_path_buf();
    let alias = bin.join("moai");
    // The Go `moai` under this name: reported, never replaced.
    common::write_executable(&alias, "#!/bin/sh\necho 'moai-adk 3.2.0'\n");
    let p = plan(&exe, &home);
    assert_eq!(p["state"], json!("blocked"), "{p}");
    assert!(
        p["reason"].as_str().unwrap().contains("moai-adk 3.2.0"),
        "{p}"
    );
    let before = names(&bin);
    let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
    assert_eq!(a["result"]["written"], json!([]), "{a}");
    assert_eq!(names(&bin), before);

    fs::remove_file(&alias).unwrap();
    std::os::unix::fs::symlink(&exe, &alias).unwrap();
    assert_eq!(plan(&exe, &home)["state"], json!("blocked"));
    fs::remove_file(&alias).unwrap();
    fs::create_dir(&alias).unwrap();
    assert_eq!(plan(&exe, &home)["state"], json!("blocked"));
}

#[test]
fn outside_the_version_folder_nothing_is_linked() {
    let d = TempDir::new().unwrap();
    let root = fs::canonicalize(d.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    // `bin/` but not under `versions/<v>/`, and a `versions/<v>/` folder not named `bin`.
    for rel in ["elsewhere/bin", "home/.moai-cli/versions/a/other"] {
        let dir = root.join(rel);
        fs::create_dir_all(&dir).unwrap();
        let exe = dir.join("moai-cli");
        common::copy_executable(Path::new(EXE), &exe);
        let p = plan(&exe, &home);
        assert_eq!(p["state"], json!("outside_layout"), "{rel}: {p}");
        let a = apply(&exe, &home, p["plan_digest"].as_str().unwrap(), "approved");
        assert_eq!(a["result"]["written"], json!([]), "{rel}: {a}");
        assert!(!dir.join("moai").exists(), "{rel}");
    }
}
