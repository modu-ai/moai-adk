//! #3004 P6-1: the read-only plan for the `<state-home>/path/moai` link. Every case runs copies
//! of the built binaries inside a temporary HOME (the real HOME, PATH and shell files are never
//! read), asks the real `--version` of the target, and checks that the preview writes nothing.
//! The Go-signature and `unsupported` (Windows) cases need no real binary and live with the
//! module's unit tests (`cargo test --lib path_link`); the plan inside `moai update` is checked
//! in `tests/update_user_layer_cli.rs`.

#![cfg(unix)]

mod common;

use moai_cli::path_link::plan_for;
use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::{symlink, PermissionsExt as _},
    path::{Path, PathBuf},
    time::SystemTime,
};
use tempfile::TempDir;

const EXE: &str = common::MOAI_CLI_BIN;

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
    /// `bin/moai-cli`, the running binary.
    cli: PathBuf,
    /// `bin/moai`, the hard link the installer makes.
    moai: PathBuf,
    link_dir: PathBuf,
    link: PathBuf,
}

/// A temporary HOME with both names of this build in `~/.moai-cli/versions/a/bin/` and no `path/`.
fn layout() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap();
    let bin = home.join(".moai-cli/versions/a/bin");
    let (cli, moai) = common::install_both_names(&bin);
    let link_dir = home.join(".moai-cli/path");
    let link = link_dir.join("moai");
    Fx {
        _tmp: tmp,
        home,
        cli,
        moai,
        link_dir,
        link,
    }
}

/// A second version folder `b` holding a real `bin/moai`.
fn older_version(f: &Fx) -> PathBuf {
    let bin = f.home.join(".moai-cli/versions/b/bin");
    let (_, moai) = common::install_both_names(&bin);
    moai
}

fn state(f: &Fx) -> Value {
    let p = plan_for(&f.home, &f.cli);
    p.to_json()
}

/// Replace `bin/moai` with a separate file made by `make`.
fn replace_moai(f: &Fx, make: impl FnOnce(&Path)) {
    let _ = fs::remove_file(&f.moai);
    make(&f.moai);
}

/// Every entry under `root`: path, kind, size, modification time and link text.
fn listing(root: &Path) -> Vec<(PathBuf, char, u64, SystemTime, Option<PathBuf>)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in fs::read_dir(&dir).unwrap().flatten() {
            let p = e.path();
            let m = fs::symlink_metadata(&p).unwrap();
            let kind = if m.file_type().is_symlink() {
                'l'
            } else if m.is_dir() {
                'd'
            } else {
                'f'
            };
            out.push((
                p.clone(),
                kind,
                m.len(),
                m.modified().unwrap(),
                fs::read_link(&p).ok(),
            ));
            if kind == 'd' {
                stack.push(p);
            }
        }
    }
    out.sort();
    out
}

#[test]
fn a_rust_hard_link_and_no_path_link_is_create() {
    let f = layout();
    let before = listing(&f.home);
    let s = state(&f);
    assert_eq!(s["state"], json!("create"), "{s}");
    assert_eq!(s["link"], json!(f.link.to_string_lossy()));
    assert_eq!(s["target"], json!(f.moai.to_string_lossy()));
    assert_eq!(s["reason"], Value::Null);
    assert_eq!(s["current"], Value::Null);
    assert_eq!(
        listing(&f.home),
        before,
        "a preview writes nothing, not even path/"
    );
    assert!(!f.link_dir.exists());
}

#[test]
fn a_link_to_this_build_is_unchanged_and_one_to_an_older_build_is_repoint() {
    let f = layout();
    fs::create_dir_all(&f.link_dir).unwrap();
    symlink(&f.moai, &f.link).unwrap();
    assert_eq!(state(&f)["state"], json!("unchanged"));
    fs::remove_file(&f.link).unwrap();
    symlink("../versions/a/bin/moai", &f.link).unwrap();
    let s = state(&f);
    assert_eq!(s["state"], json!("unchanged"), "a relative link: {s}");
    assert_eq!(s["current"], json!("../versions/a/bin/moai"));

    fs::remove_file(&f.link).unwrap();
    let older = older_version(&f);
    symlink(&older, &f.link).unwrap();
    let before_repoint = listing(&f.home);
    let s = state(&f);
    assert_eq!(s["state"], json!("repoint"), "{s}");
    assert_eq!(s["current"], json!(older.to_string_lossy()));
    assert_eq!(s["target"], json!(f.moai.to_string_lossy()));
    assert_eq!(listing(&f.home), before_repoint);
}

#[test]
fn a_target_that_is_not_a_rust_executable_is_blocked_with_or_without_a_link() {
    let f = layout();
    // A `#!` shim in place of the hard link.
    replace_moai(&f, |p| {
        common::write_executable(p, "#!/bin/sh\nexec moai-cli \"$@\"\n")
    });
    let s = state(&f);
    assert_eq!(s["state"], json!("blocked"), "{s}");
    assert!(s["reason"].as_str().unwrap().contains("#!"), "{s}");
    // A program that does not say `(rust)`: `true` prints nothing or another tool's line.
    let not_rust = ["/usr/bin/true", "/bin/true"]
        .iter()
        .map(Path::new)
        .find(|p| p.is_file())
        .expect("a `true` program");
    replace_moai(&f, |p| common::copy_executable(not_rust, p));
    let s = state(&f);
    assert_eq!(s["state"], json!("blocked"), "{s}");
    assert!(
        s["reason"].as_str().unwrap().contains("not the Rust build"),
        "{s}"
    );
    // The link reaches this very target: a link is not proof of Rust.
    fs::create_dir_all(&f.link_dir).unwrap();
    symlink(&f.moai, &f.link).unwrap();
    let before = listing(&f.home);
    let s = state(&f);
    assert_eq!(s["state"], json!("blocked"), "{s}");
    assert_eq!(s["current"], json!(f.moai.to_string_lossy()));
    assert_eq!(listing(&f.home), before);
}

#[test]
fn a_missing_or_unexecutable_target_is_blocked() {
    let f = layout();
    replace_moai(&f, |_| {});
    let s = state(&f);
    assert_eq!(s["state"], json!("blocked"), "{s}");
    assert!(s["reason"].as_str().unwrap().contains("missing"), "{s}");
    replace_moai(&f, |p| common::copy_executable(Path::new(EXE), p));
    assert_eq!(state(&f)["state"], json!("create"));
    fs::set_permissions(&f.moai, fs::Permissions::from_mode(0o644)).unwrap();
    let s = state(&f);
    assert_eq!(s["state"], json!("blocked"), "{s}");
    assert!(
        s["reason"].as_str().unwrap().contains("not executable"),
        "{s}"
    );
}

#[test]
fn a_shim_file_or_an_outside_link_is_foreign_and_left_alone() {
    let f = layout();
    fs::create_dir_all(&f.link_dir).unwrap();
    // A regular file (a routing shim, say).
    common::write_executable(&f.link, "#!/bin/sh\nexec \"$HOME/elsewhere/moai\" \"$@\"\n");
    let before = listing(&f.home);
    let s = state(&f);
    assert_eq!(s["state"], json!("foreign"), "{s}");
    assert_eq!(listing(&f.home), before);
    // Links that leave `versions/`: the Go install, a system tool, a `..` escape.
    let go = f.home.join("go/bin/moai");
    fs::create_dir_all(go.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(EXE), &go);
    for text in [
        go.to_string_lossy().into_owned(),
        "/usr/bin/true".to_string(),
        "../versions/../../go/bin/moai".to_string(),
        f.cli.to_string_lossy().into_owned(), // this version, but the other name
    ] {
        fs::remove_file(&f.link).unwrap();
        symlink(&text, &f.link).unwrap();
        let s = state(&f);
        assert_eq!(s["state"], json!("foreign"), "{text}: {s}");
        assert_eq!(s["current"], json!(text));
    }
}

#[test]
fn a_binary_outside_the_version_layout_is_outside_layout() {
    let f = layout();
    let stray = f.home.join("stray/moai-cli");
    fs::create_dir_all(stray.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(EXE), &stray);
    let before = listing(&f.home);
    let s = plan_for(&f.home, &stray).to_json();
    assert_eq!(s["state"], json!("outside_layout"), "{s}");
    assert_eq!(s["link"], Value::Null);
    assert_eq!(listing(&f.home), before);
}
