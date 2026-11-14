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

/// The `export NAME=` lines of one `lane-env` run from `cwd`, after checking it succeeded.
fn exports_from(cwd: &Path, lane: &str, base: &Path) -> Vec<String> {
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .current_dir(cwd)
        .args(["lane-env", "--lane", lane, "--base"])
        .arg(base)
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(0), "{lane}: {out:?}");
    String::from_utf8_lossy(&out.stdout)
        .lines()
        .filter(|l| l.starts_with("export "))
        .map(str::to_string)
        .collect()
}

fn target_line(exports: &[String]) -> String {
    let lines: Vec<_> = exports
        .iter()
        .filter(|l| l.starts_with("export CARGO_TARGET_DIR="))
        .collect();
    assert_eq!(lines.len(), 1, "{exports:?}");
    lines[0].clone()
}

/// #2930 R-1: two linked worktrees of one lane get the same build folder, so dependencies
/// are built once per lane; another lane gets its own.
#[test]
fn two_worktrees_of_one_lane_share_the_target_dir() {
    use common::git::git;
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let repo = root.join("repo");
    let base = root.join("base");
    fs::create_dir(&repo).unwrap();
    fs::create_dir(&base).unwrap();
    git(&repo, &["init", "-q", "-b", "main"]);
    fs::write(repo.join("f"), "x\n").unwrap();
    git(&repo, &["add", "f"]);
    git(
        &repo,
        &[
            "-c",
            "user.name=F",
            "-c",
            "user.email=f@example.invalid",
            "commit",
            "-q",
            "-m",
            "base",
        ],
    );
    let (one, two) = (root.join("wt-one"), root.join("wt-two"));
    for (path, branch) in [(&one, "one"), (&two, "two")] {
        git(
            &repo,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                branch,
                path.to_str().unwrap(),
            ],
        );
    }

    let a = target_line(&exports_from(&one, "agent-7", &base));
    let b = target_line(&exports_from(&two, "agent-7", &base));
    assert_eq!(a, b, "one lane, two worktrees");
    assert!(
        a.contains(&base.join("moai-rc-target-agent-7").display().to_string()),
        "{a}"
    );
    let other = target_line(&exports_from(&two, "agent-8", &base));
    assert_ne!(a, other, "another lane gets its own folder");
}

/// #2930 R-1: every lane build runs with incremental compilation off (exactly one line).
#[test]
fn lane_env_exports_cargo_incremental_off() {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let exports = exports_from(&base, "agent-7", &base);
    let count = exports
        .iter()
        .filter(|l| *l == "export CARGO_INCREMENTAL=0")
        .count();
    assert_eq!(count, 1, "{exports:?}");
}

/// The lane-name limit at its edge: 64 characters pass, 65 are refused with stdout empty.
#[test]
fn lane_name_length_edge() {
    let dir = tempfile::TempDir::new().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let longest = "x".repeat(64);
    let exports = exports_from(&base, &longest, &base);
    assert!(target_line(&exports).ends_with(&format!("moai-rc-target-{longest}'")));

    let too_long = "x".repeat(65);
    let out = lane_env(&["--lane", &too_long, "--base", base.to_str().unwrap()]);
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8_lossy(&out.stderr).contains("invalid lane name"));
}
