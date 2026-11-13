//! #2699: `cargo test` points the macOS git shim at the Command Line Tools, so a git call made
//! under a fresh temporary HOME no longer starts `xcodebuild -license check` each time
//! (measured in the Issue: tests/worktree_cli.rs 27 → 0 xcodebuild processes, 7.1 s → 3.3 s).
use std::{fs, path::Path};

const CLT: &str = "/Library/Developer/CommandLineTools";

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
