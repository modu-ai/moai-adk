//! #2495: the installed binary is identified by its real file, so calling it through a link
//! (the decided layout `~/.moai-cli/path/moai-cli`, decisions L165) works. Only temporary
//! homes and fixture links are used; the real HOME and the installed binary are untouched.
#![cfg(unix)]

use moai_cli::install::{plan_receipt, record_receipt, InstallerOwner};
use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn staged_package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .env_clear()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .expect("spawn moai-cli")
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).expect("envelope JSON on stdout");
    assert_eq!(
        v["ok"],
        json!(true),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    v["result"].clone()
}

/// `current_exe` keeps the link on macOS and resolves it on Linux (std `_NSGetExecutablePath`
/// vs `/proc/self/exe`), so a CLI started through a link only knows it on macOS.
fn link_is_visible_to_the_cli() -> bool {
    cfg!(target_vendor = "apple")
}

// AC-1 (library): a link and its target plan the same receipt and install id.
#[test]
fn a_link_plans_the_same_receipt_as_its_target() {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let link = tmp.path().join("link-moai-cli");
    symlink(EXE, &link).unwrap();
    let chain = tmp.path().join("chain-moai-cli");
    symlink(&link, &chain).unwrap();

    let real = plan_receipt(Path::new(EXE), &home, InstallerOwner::Unknown).unwrap();
    for path in [&link, &chain] {
        let via = plan_receipt(path, &home, InstallerOwner::Unknown)
            .unwrap_or_else(|e| panic!("{}: {e}", path.display()));
        assert_eq!(via.install_id, real.install_id);
        assert_eq!(via.receipt.binary_path, real.receipt.binary_path);
        assert_eq!(via.receipt.binary_sha256, real.receipt.binary_sha256);
    }
}

// AC-2: a link to something that is not a regular file, or a dangling link, is refused.
#[test]
fn a_link_to_a_directory_or_nothing_is_refused() {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    fs::create_dir_all(tmp.path().join("dir")).unwrap();
    fs::create_dir_all(&home).unwrap();
    let to_dir = tmp.path().join("to-dir");
    symlink(tmp.path().join("dir"), &to_dir).unwrap();
    let dangling = tmp.path().join("dangling");
    symlink(tmp.path().join("missing"), &dangling).unwrap();

    let err = plan_receipt(&to_dir, &home, InstallerOwner::Unknown).unwrap_err();
    assert_eq!(err, "binary path is not a regular file");
    let err = plan_receipt(&dangling, &home, InstallerOwner::Unknown).unwrap_err();
    assert_eq!(err, "binary path unavailable");
}

// AC-1 (CLI): update-plan through a link succeeds with the same plan as the real path.
#[test]
fn update_plan_through_a_link_matches_the_real_path() {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(&home).unwrap();
    let link = home.join("bin-link");
    symlink(EXE, &link).unwrap();
    let stage = staged_package();
    let args = [
        "update-plan",
        home.to_str().unwrap(),
        stage.to_str().unwrap(),
    ];

    let real = run(Path::new(EXE), &home, &args);
    let via = run(&link, &home, &args);
    assert!(
        real.status.success(),
        "{}",
        String::from_utf8_lossy(&real.stdout)
    );
    assert!(
        via.status.success(),
        "{}",
        String::from_utf8_lossy(&via.stdout)
    );
    assert_eq!(result(&via), result(&real));
    // Planning wrote nothing into the home: only the fixture link is there.
    let names: Vec<_> = fs::read_dir(&home)
        .unwrap()
        .map(|e| e.unwrap().file_name())
        .collect();
    assert_eq!(names, ["bin-link"]);
}

// AC-3: install-inspect identifies itself the same way through a link, and says so.
#[test]
fn install_inspect_through_a_link_is_still_self() {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(&home).unwrap();
    let plan = plan_receipt(Path::new(EXE), &home, InstallerOwner::Unknown).unwrap();
    record_receipt(&home, &plan).unwrap();
    let link = home.join("bin-link");
    symlink(EXE, &link).unwrap();

    let real = result(&run(Path::new(EXE), &home, &["install-inspect", "--json"]));
    let via = result(&run(&link, &home, &["install-inspect", "--json"]));
    assert_eq!(real["self"]["state"], json!("self"));
    assert_eq!(via["self"]["state"], real["self"]["state"]);
    assert_eq!(via["self"]["current_exe"], real["self"]["current_exe"]);
    assert_eq!(via["self"]["binary_sha256"], real["self"]["binary_sha256"]);
    assert_eq!(real["self"]["invoked_via_link"], json!(false));
    assert_eq!(
        via["self"]["invoked_via_link"],
        json!(link_is_visible_to_the_cli())
    );
    if link_is_visible_to_the_cli() {
        assert_eq!(via["self"]["invoked_exe"], json!("~/bin-link"));
    } else {
        assert_eq!(via["self"]["invoked_exe"], real["self"]["current_exe"]);
    }
}
