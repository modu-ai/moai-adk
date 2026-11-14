//! #2841: `moai lane-env --lane NAME [--base DIR]` prints export lines a shell evaluates in one
//! line; a refusal prints nothing on stdout, so `eval "$(…)"` never applies half an environment.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use std::fs;
use std::path::Path;
use std::process::{Command, Output};

fn lane_env(args: &[&str]) -> Output {
    Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .arg("lane-env")
        .args(args)
        .output()
        .unwrap()
}

/// What a POSIX shell sees after the documented one line `e=$(moai lane-env …) && eval "$e"`
/// (a bare `eval "$(…)"` returns 0 on an empty string, POSIX `eval`, so it cannot stop on a
/// refusal).
fn evaluated(base: &Path, lane: &str) -> (i32, String) {
    let out = Command::new("sh")
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("MOAI", common::MOAI_BIN)
        .env("BASE", base)
        .env("LANE", lane)
        .args([
            "-c",
            r#"e=$("$MOAI" lane-env --lane "$LANE" --base "$BASE") && eval "$e" && printf '%s\n%s\n' "$CARGO_TARGET_DIR" "${CARGO_HOME-unset}""#,
        ])
        .output()
        .unwrap();
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

#[test]
fn eval_sets_the_lane_folders() {
    let dir = tempfile::TempDir::new().unwrap();
    // A base with a space and a quote: the shell must still get one word per value.
    let base = fs::canonicalize(dir.path()).unwrap().join("it's a base");
    fs::create_dir(&base).unwrap();
    let target = base.join("moai-rc-target-agent-7");

    let (code, out) = evaluated(&base, "agent-7");
    assert_eq!(code, 0, "{out}");
    assert_eq!(out, format!("{}\nunset\n", target.display()));

    fs::create_dir_all(base.join("moai-rc-product-cargo/registry/src")).unwrap();
    let (_, out) = evaluated(&base, "agent-7");
    assert_eq!(
        out,
        format!(
            "{}\n{}\n",
            target.display(),
            base.join("moai-rc-product-cargo").display()
        )
    );

    fs::create_dir_all(base.join("moai-rc-cargo-agent-7/registry/src")).unwrap();
    let (_, out) = evaluated(&base, "agent-7");
    assert_eq!(
        out,
        format!(
            "{}\n{}\n",
            target.display(),
            base.join("moai-rc-cargo-agent-7").display()
        )
    );
    // Reads only: the target folder is not created.
    assert!(!target.exists());
}

#[test]
fn refusals_print_nothing_on_stdout() {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    std::os::unix::fs::symlink(&base, base.join("moai-rc-target-agent-7")).unwrap();
    let base_s = base.to_str().unwrap();
    for (args, says) in [
        (
            vec!["--lane", "agent-7", "--base", base_s],
            "is a symbolic link",
        ),
        (
            vec!["--lane", "../x", "--base", base_s],
            "invalid lane name",
        ),
        (vec!["--lane", "a b", "--base", base_s], "invalid lane name"),
        (
            vec!["--lane", "agent-8", "--base", "tmp"],
            "not an absolute path",
        ),
    ] {
        let out = lane_env(&args);
        assert_eq!(out.status.code(), Some(1), "{args:?}");
        assert!(out.stdout.is_empty(), "{args:?}: stdout must stay empty");
        let err = String::from_utf8_lossy(&out.stderr);
        assert!(err.contains(says), "{args:?}: {err}");
    }
    // The documented line stops at the failed substitution's `&&`: nothing is exported.
    let (code, out) = evaluated(&base, "agent-7");
    assert_ne!(code, 0);
    assert_eq!(out, "");
}

#[test]
fn a_missing_lane_is_a_usage_error() {
    let out = lane_env(&[]);
    assert_eq!(out.status.code(), Some(2));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8_lossy(&out.stderr).contains("--lane"));
}
