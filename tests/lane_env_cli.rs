//! #2841: `moai lane-env --lane NAME [--base DIR]` prints export lines a shell evaluates in one
//! line; a refusal prints nothing on stdout, so `eval "$(…)"` never applies half an environment.
//! Card C-82 follow-up (user 2026-11-17 08:22Z, "빌드 슬롯도 이번 버전에 맞게 개선"):
//! `CARGO_TARGET_DIR` is the worktree's own build folder `<HOME>/.moai/build/<name>`, named and
//! recorded as `tools/build_folder.py env` does. Every case uses a temporary HOME.
#![cfg(unix)]
mod common;

use common::git::git;
use common::CleanEnv;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

/// A repository `<root>/repo` with one commit, a temporary HOME `<root>/the home` and a base
/// folder for the cargo homes; every path is a real path.
struct Fx {
    _dir: tempfile::TempDir,
    root: PathBuf,
    repo: PathBuf,
    home: PathBuf,
    base: PathBuf,
}

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (repo, home) = (root.join("repo"), root.join("the home"));
    // A base with a space and a quote: the shell must still get one word per value.
    let base = root.join("it's a base");
    for d in [&repo, &home, &base] {
        fs::create_dir(d).unwrap();
    }
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
    Fx {
        _dir: dir,
        root,
        repo,
        home,
        base,
    }
}

impl Fx {
    /// A linked worktree at `path` on a new branch.
    fn worktree(&self, path: &Path) -> PathBuf {
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        static N: std::sync::atomic::AtomicUsize = std::sync::atomic::AtomicUsize::new(0);
        let branch = format!("b{}", N.fetch_add(1, std::sync::atomic::Ordering::Relaxed));
        git(
            &self.repo,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                &branch,
                path.to_str().unwrap(),
            ],
        );
        path.to_path_buf()
    }

    fn folder(&self, name: &str) -> PathBuf {
        self.home.join(".moai").join("build").join(name)
    }
}

/// `moai lane-env` run from `cwd` with only PATH, HOME and git's config pinned.
fn lane_env(cwd: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .current_dir(cwd)
        .arg("lane-env")
        .args(args)
        .output()
        .unwrap()
}

/// What a POSIX shell sees after the documented one line `e=$(moai lane-env …) && eval "$e"`
/// (a bare `eval "$(…)"` returns 0 on an empty string, POSIX `eval`, so it cannot stop on a
/// refusal).
fn evaluated(cwd: &Path, home: &Path, base: &Path, lane: &str) -> (i32, String) {
    let out = Command::new("sh")
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("MOAI", common::MOAI_BIN)
        .env("BASE", base)
        .env("LANE", lane)
        .current_dir(cwd)
        .args([
            "-c",
            r#"e=$("$MOAI" lane-env --lane "$LANE" --base "$BASE") && eval "$e" && printf '%s\n%s\n%s\n' "$CARGO_TARGET_DIR" "${CARGO_HOME-unset}" "$CARGO_INCREMENTAL""#,
        ])
        .output()
        .unwrap();
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

fn record(folder: &Path) -> Value {
    serde_json::from_slice(&fs::read(folder.join(".moai-build.json")).unwrap()).unwrap()
}

fn mode(path: &Path) -> u32 {
    use std::os::unix::fs::PermissionsExt;
    fs::metadata(path).unwrap().permissions().mode() & 0o777
}

#[test]
fn eval_sets_the_worktree_folder_and_the_lane_cargo_home() {
    let f = fx();
    let wt = f.worktree(&f.root.join("wt-one"));
    let target = f.folder("repo-wt-one");

    let (code, out) = evaluated(&wt, &f.home, &f.base, "agent-7");
    assert_eq!(code, 0, "{out}");
    assert_eq!(out, format!("{}\nunset\n0\n", target.display()));
    // The folder, its parents (0700) and the owner record of C-82 are there.
    assert_eq!(mode(&f.home.join(".moai")), 0o700);
    assert_eq!(mode(&f.home.join(".moai/build")), 0o700);
    assert_eq!(mode(&target), 0o700);
    assert_eq!(
        record(&target),
        serde_json::json!({
            "schema_version": 1,
            "git_common_dir": f.repo.join(".git").display().to_string(),
            "worktree": wt.display().to_string(),
        })
    );

    fs::create_dir_all(f.base.join("moai-rc-product-cargo/registry/src")).unwrap();
    let (_, out) = evaluated(&wt, &f.home, &f.base, "agent-7");
    let shared = f.base.join("moai-rc-product-cargo");
    assert_eq!(
        out,
        format!("{}\n{}\n0\n", target.display(), shared.display())
    );

    fs::create_dir_all(f.base.join("moai-rc-cargo-agent-7/registry/src")).unwrap();
    let (_, out) = evaluated(&wt, &f.home, &f.base, "agent-7");
    let own = f.base.join("moai-rc-cargo-agent-7");
    assert_eq!(out, format!("{}\n{}\n0\n", target.display(), own.display()));
    // The former per-lane folder is not made.
    assert!(!f.base.join("moai-rc-target-agent-7").exists());
}

#[test]
fn refusals_print_nothing_on_stdout() {
    let f = fx();
    let base_s = f.base.to_str().unwrap();
    let elsewhere = f.root.join("elsewhere");
    fs::create_dir(&elsewhere).unwrap();
    let other_home = f.root.join("linked home");
    fs::create_dir(&other_home).unwrap();
    std::os::unix::fs::symlink(&elsewhere, other_home.join(".moai")).unwrap();
    let file_home = f.root.join("file home");
    fs::create_dir_all(file_home.join(".moai/build")).unwrap();
    fs::write(file_home.join(".moai/build/repo"), "x").unwrap();
    // A folder whose owner record names another worktree.
    let taken_home = f.root.join("taken home");
    let taken = taken_home.join(".moai/build/repo");
    fs::create_dir_all(&taken).unwrap();
    fs::write(
        taken.join(".moai-build.json"),
        r#"{"schema_version": 1, "git_common_dir": "/x/.git", "worktree": "/x"}"#,
    )
    .unwrap();
    let not_a_repo = f.root.join("not a repo");
    fs::create_dir(&not_a_repo).unwrap();

    for (cwd, home, args, says) in [
        (
            &f.repo,
            &other_home,
            vec!["--lane", "agent-7", "--base", base_s],
            "is a link",
        ),
        (
            &f.repo,
            &file_home,
            vec!["--lane", "agent-7", "--base", base_s],
            "is not a folder",
        ),
        (
            &f.repo,
            &taken_home,
            vec!["--lane", "agent-7", "--base", base_s],
            "belongs to the worktree /x",
        ),
        (
            &not_a_repo,
            &f.home,
            vec!["--lane", "agent-7", "--base", base_s],
            "git rev-parse",
        ),
        (
            &f.repo,
            &f.home,
            vec!["--lane", "../x", "--base", base_s],
            "invalid lane name",
        ),
        (
            &f.repo,
            &f.home,
            vec!["--lane", "a b", "--base", base_s],
            "invalid lane name",
        ),
        (
            &f.repo,
            &f.home,
            vec!["--lane", "agent-8", "--base", "tmp"],
            "not an absolute path",
        ),
    ] {
        let out = lane_env(cwd, home, &args);
        assert_eq!(out.status.code(), Some(1), "{args:?}: {out:?}");
        assert!(out.stdout.is_empty(), "{args:?}: stdout must stay empty");
        let err = String::from_utf8_lossy(&out.stderr);
        assert!(err.contains(says), "{args:?}: {err}");
    }
    // Nothing was written through the link, and the other worktree's record stays.
    // (Following the linked `.moai` would have made `build` in its target.)
    assert!(!elsewhere.join("build").exists());
    assert_eq!(record(&taken)["worktree"], "/x");
    // A relative HOME is refused too.
    let out = lane_env(
        &f.repo,
        Path::new("rel"),
        &["--lane", "agent-7", "--base", base_s],
    );
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8_lossy(&out.stderr).contains("HOME must be an absolute folder"));
    // The documented line stops at the failed substitution's `&&`: nothing is exported.
    let (code, out) = evaluated(&f.repo, &other_home, &f.base, "agent-7");
    assert_ne!(code, 0);
    assert_eq!(out, "");
}

/// C-82 slot audit N-a: USERPROFILE stands in for an unset HOME on Windows only. On this
/// (Unix) host, with HOME unset and only USERPROFILE set, `lane-env` refuses with stdout empty
/// and makes nothing under USERPROFILE: `resolve` passes `cfg!(windows)`, not `true`.
#[test]
fn userprofile_alone_is_not_a_home_off_windows() {
    let f = fx();
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("USERPROFILE", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .current_dir(&f.repo)
        .args(["lane-env", "--lane", "agent-7", "--base"])
        .arg(&f.base)
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1), "{out:?}");
    assert!(out.stdout.is_empty());
    assert!(
        String::from_utf8_lossy(&out.stderr).contains("HOME must be an absolute folder"),
        "{out:?}"
    );
    assert!(
        !f.home.join(".moai").exists(),
        "nothing made under USERPROFILE"
    );
}

#[test]
fn a_missing_lane_is_a_usage_error() {
    let f = fx();
    let out = lane_env(&f.repo, &f.home, &[]);
    assert_eq!(out.status.code(), Some(2));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8_lossy(&out.stderr).contains("--lane"));
}

/// The `export NAME=` lines of one `lane-env` run from `cwd`, after checking it succeeded.
fn exports_from(cwd: &Path, home: &Path, lane: &str, base: &Path) -> Vec<String> {
    let out = lane_env(
        cwd,
        home,
        &["--lane", lane, "--base", base.to_str().unwrap()],
    );
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

/// C-82: one build folder per worktree, never shared between worktrees; the lane does not
/// change it (was #2930 R-1: one folder for every worktree of a lane).
#[test]
fn each_worktree_gets_its_own_folder() {
    let f = fx();
    let one = f.worktree(&f.root.join("wt-one"));
    let two = f.worktree(&f.root.join("wt-two"));
    let a = target_line(&exports_from(&one, &f.home, "agent-7", &f.base));
    let b = target_line(&exports_from(&two, &f.home, "agent-7", &f.base));
    assert_ne!(a, b, "two worktrees of one lane");
    assert!(
        a.contains(&f.folder("repo-wt-one").display().to_string()),
        "{a}"
    );
    let other_lane = target_line(&exports_from(&one, &f.home, "agent-8", &f.base));
    assert_eq!(a, other_lane, "the lane does not name the folder");
    // From a subfolder of the primary checkout: the project name alone.
    fs::create_dir(f.repo.join("sub")).unwrap();
    let primary = target_line(&exports_from(
        &f.repo.join("sub"),
        &f.home,
        "agent-7",
        &f.base,
    ));
    assert!(
        primary.contains(&format!("{}'", f.folder("repo").display())),
        "{primary}"
    );
}

/// #2930 R-1: every lane build runs with incremental compilation off (exactly one line).
#[test]
fn lane_env_exports_cargo_incremental_off() {
    let f = fx();
    let exports = exports_from(&f.repo, &f.home, "agent-7", &f.base);
    let count = exports
        .iter()
        .filter(|l| *l == "export CARGO_INCREMENTAL=0")
        .count();
    assert_eq!(count, 1, "{exports:?}");
}

/// The lane-name limit at its edge: 64 characters pass, 65 are refused with stdout empty.
#[test]
fn lane_name_length_edge() {
    let f = fx();
    let longest = "x".repeat(64);
    let exports = exports_from(&f.repo, &f.home, &longest, &f.base);
    assert!(target_line(&exports).ends_with("/.moai/build/repo'"));

    let too_long = "x".repeat(65);
    let out = lane_env(
        &f.repo,
        &f.home,
        &["--lane", &too_long, "--base", f.base.to_str().unwrap()],
    );
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8_lossy(&out.stderr).contains("invalid lane name"));
}

/// `python3 -I tools/build_folder.py env` from `cwd`: (exit code, evaluated CARGO_TARGET_DIR).
fn python_env(cwd: &Path, home: &Path) -> (i32, String) {
    let tool = Path::new(env!("CARGO_MANIFEST_DIR")).join("tools/build_folder.py");
    let out = Command::new("sh")
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("TOOL", tool)
        .current_dir(cwd)
        .args([
            "-c",
            r#"e=$(python3 -I "$TOOL" env) && eval "$e" && printf '%s\n' "$CARGO_TARGET_DIR""#,
        ])
        .output()
        .unwrap();
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// One source of truth: for the same repository, worktree and HOME, `moai lane-env` and
/// `tools/build_folder.py env` name the same folder and write the same owner record, and each
/// accepts the record the other wrote.
#[test]
fn the_folder_and_record_match_tools_build_folder() {
    let f = fx();
    let cases: Vec<(PathBuf, &str)> = vec![
        (f.repo.clone(), "repo"),
        (f.worktree(&f.root.join("wt-one")), "repo-wt-one"),
        (f.worktree(&f.root.join("repo-x")), "repo-x"),
        (f.worktree(&f.root.join("other/repo")), "repo"),
        (f.worktree(&f.root.join("a b+c")), "repo-a_b_c"),
        (f.worktree(&f.root.join("한글")), "repo-__"),
        (
            f.worktree(&f.repo.join(".claude/worktrees/agent-1")),
            "repo-agent-1",
        ),
    ];
    for (i, (wt, name)) in cases.iter().enumerate() {
        let (py_home, rs_home) = (f.root.join(format!("py{i}")), f.root.join(format!("rs{i}")));
        fs::create_dir(&py_home).unwrap();
        fs::create_dir(&rs_home).unwrap();
        let (pc, py) = python_env(wt, &py_home);
        let (rc, rs) = evaluated(wt, &rs_home, &f.base, "agent-7");
        assert_eq!((pc, rc), (0, 0), "{name}: {py} / {rs}");
        let py_dir = py.trim_end().to_string();
        let rs_dir = rs.lines().next().unwrap().to_string();
        assert_eq!(
            py_dir,
            py_home.join(".moai/build").join(name).display().to_string()
        );
        assert_eq!(
            rs_dir,
            rs_home.join(".moai/build").join(name).display().to_string()
        );
        assert_eq!(
            record(Path::new(&py_dir)),
            record(Path::new(&rs_dir)),
            "{name}"
        );
        // Each accepts the other's record: Python after Rust, Rust after Python.
        assert_eq!(
            python_env(wt, &rs_home),
            (0, format!("{rs_dir}\n")),
            "{name}"
        );
        assert_eq!(evaluated(wt, &py_home, &f.base, "agent-7").0, 0, "{name}");
    }
}
