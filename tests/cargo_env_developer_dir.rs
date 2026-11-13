//! #2699: `cargo test` points the macOS git shim at the Command Line Tools, so a git call made
//! under a fresh temporary HOME no longer starts `xcodebuild -license check` each time
//! (measured in the Issue: tests/worktree_cli.rs 27 → 0 xcodebuild processes, 7.1 s → 3.3 s).
mod common;
use common::CleanEnv;
use std::{fs, path::Path, process::Command};

const CLT: &str = "/Library/Developer/CommandLineTools";

/// AUDIT #2704 B1: a child started with a cleared environment still gets the developer folder,
/// so moai-cli's git calls in `env_clear` tests skip the license check too.
#[test]
fn a_cleared_child_environment_keeps_the_developer_dir() {
    let out = Command::new("/usr/bin/env")
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .output()
        .expect("/usr/bin/env");
    let child = String::from_utf8_lossy(&out.stdout);
    let lines: Vec<&str> = child.lines().collect();
    assert!(lines.contains(&"PATH=/usr/bin:/bin"), "{child}");
    match std::env::var("DEVELOPER_DIR") {
        Ok(dir) if Path::new(&dir).is_dir() => {
            assert!(
                lines.contains(&format!("DEVELOPER_DIR={dir}").as_str()),
                "{child}"
            );
            assert_eq!(lines.len(), 2, "nothing else survives the clear: {child}");
        }
        _ => assert!(!child.contains("DEVELOPER_DIR="), "{child}"),
    }
}

/// AUDIT #2704 B1: no test clears a child's environment directly; `.clean_env()` is the one
/// place that does and passes the developer folder on.
#[test]
fn tests_clear_child_environments_only_through_clean_env() {
    let needle = format!(".env_{}()", "clear");
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join("tests");
    let mut found = Vec::new();
    for entry in fs::read_dir(&dir).unwrap() {
        let path = entry.unwrap().path();
        if path.extension().is_some_and(|e| e == "rs")
            && fs::read_to_string(&path).unwrap().contains(&needle)
        {
            found.push(path.file_name().unwrap().to_string_lossy().into_owned());
        }
    }
    let helper = fs::read_to_string(dir.join("common/mod.rs")).unwrap();
    assert_eq!(
        helper.matches(&needle).count(),
        1,
        "tests/common/mod.rs clears once"
    );
    assert!(
        found.is_empty(),
        "use .clean_env() from tests/common instead: {found:?}"
    );
}

/// The repository's cargo config sets the variable for every test process, without `force`.
#[test]
fn cargo_config_points_the_git_shim_at_the_command_line_tools() {
    let text = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(".cargo/config.toml"))
        .expect(".cargo/config.toml");
    let env_table = text.split("[env]").nth(1).expect("an [env] table");
    let line = env_table
        .lines()
        .find(|l| l.trim_start().starts_with("DEVELOPER_DIR"))
        .expect("DEVELOPER_DIR in [env]");
    assert_eq!(line.trim(), format!("DEVELOPER_DIR = \"{CLT}\""));
}

/// On macOS the value reaches the test process itself (cargo `[env]`), unless the caller set
/// its own DEVELOPER_DIR before running cargo.
#[cfg(target_os = "macos")]
#[test]
fn the_test_process_carries_the_developer_dir() {
    let got = std::env::var("DEVELOPER_DIR").expect("DEVELOPER_DIR reaches the test process");
    assert!(!got.is_empty());
    if got == CLT {
        assert!(
            Path::new(CLT).join("usr/bin/git").exists(),
            "{CLT} has no usr/bin/git: set DEVELOPER_DIR to an installed developer folder"
        );
    }
}
