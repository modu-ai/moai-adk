//! #2841: `moai lane-env --lane NAME [--base DIR]` prints export lines a shell evaluates in one
//! line; a refusal prints nothing on stdout, so `eval "$(…)"` never applies half an environment.
//! A lane builds in its worktree's own `target/`: `lane-env` unsets `CARGO_TARGET_DIR`, makes
//! nothing under HOME and writes no owner record (SPEC-20261123-lane-env-own-target). Every case
//! uses a temporary HOME.
#![cfg(unix)]
mod common;

use common::git::git;
use common::CleanEnv;
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

/// The value a shell already holds in `CARGO_TARGET_DIR` before it evaluates the lines: the
/// retired build folder the former `lane-env` handed out.
const OLD_TARGET: &str = "/old/home/.moai/build/repo-wt-one";

/// What a POSIX shell sees after the documented one line `e=$(moai lane-env …) && eval "$e"`
/// (a bare `eval "$(…)"` returns 0 on an empty string, POSIX `eval`, so it cannot stop on a
/// refusal), started with [`OLD_TARGET`] in `CARGO_TARGET_DIR`.
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
        .env("CARGO_TARGET_DIR", OLD_TARGET)
        .current_dir(cwd)
        .args([
            "-c",
            r#"e=$("$MOAI" lane-env --lane "$LANE" --base "$BASE") && eval "$e" && printf '%s\n%s\n%s\n' "${CARGO_TARGET_DIR-unset}" "${CARGO_HOME-unset}" "$CARGO_INCREMENTAL""#,
        ])
        .output()
        .unwrap();
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// Every path the tree under `dir` holds, `dir` itself left out (nothing is followed).
fn names_under(dir: &Path) -> Vec<String> {
    let mut found = Vec::new();
    let mut todo = vec![dir.to_path_buf()];
    while let Some(d) = todo.pop() {
        let Ok(rd) = fs::read_dir(&d) else { continue };
        for e in rd.flatten() {
            found.push(e.path().display().to_string());
            if e.file_type().map(|t| t.is_dir()).unwrap_or(false) {
                todo.push(e.path());
            }
        }
    }
    found.sort();
    found
}

/// The lines of one `lane-env` run from `cwd`, after checking it succeeded.
fn lines_from(cwd: &Path, home: &Path, lane: &str, base: &Path) -> Vec<String> {
    let out = lane_env(
        cwd,
        home,
        &["--lane", lane, "--base", base.to_str().unwrap()],
    );
    assert_eq!(out.status.code(), Some(0), "{lane}: {out:?}");
    String::from_utf8_lossy(&out.stdout)
        .lines()
        .map(str::to_string)
        .collect()
}

/// AC-1: the shell started with the retired folder in `CARGO_TARGET_DIR`; after the eval the
/// variable is gone, so cargo builds in the worktree's own `target/`.
#[test]
fn eval_unsets_cargo_target_dir_and_keeps_the_cargo_home() {
    let f = fx();
    let wt = f.worktree(&f.root.join("wt-one"));

    let (code, out) = evaluated(&wt, &f.home, &f.base, "agent-7");
    assert_eq!(code, 0, "{out}");
    assert_eq!(out, "unset\nunset\n0\n");
    let lines = lines_from(&wt, &f.home, "agent-7", &f.base);
    assert!(
        lines.iter().any(|l| l == "unset CARGO_TARGET_DIR"),
        "{lines:?}"
    );
    assert!(
        !lines.iter().any(|l| l.contains("export CARGO_TARGET_DIR")),
        "{lines:?}"
    );

    fs::create_dir_all(f.base.join("moai-rc-product-cargo/registry/src")).unwrap();
    let (_, out) = evaluated(&wt, &f.home, &f.base, "agent-7");
    let shared = f.base.join("moai-rc-product-cargo");
    assert_eq!(out, format!("unset\n{}\n0\n", shared.display()));

    fs::create_dir_all(f.base.join("moai-rc-cargo-agent-7/registry/src")).unwrap();
    let (_, out) = evaluated(&wt, &f.home, &f.base, "agent-7");
    let own = f.base.join("moai-rc-cargo-agent-7");
    assert_eq!(out, format!("unset\n{}\n0\n", own.display()));
    // The former per-lane folder is not made.
    assert!(!f.base.join("moai-rc-target-agent-7").exists());
}

/// AC-2: `lane-env` makes nothing under HOME and needs none: the folder, its parents and the
/// owner record the former one wrote are not there, with a HOME, without one and with a
/// relative one.
#[test]
fn lane_env_makes_nothing_under_home() {
    let f = fx();
    let wt = f.worktree(&f.root.join("wt-one"));
    fs::create_dir(f.repo.join("sub")).unwrap();
    let base = f.base.to_str().unwrap();
    let args = ["--lane", "agent-7", "--base", base];

    let mut shapes = Vec::new();
    for cwd in [&wt, &f.repo, &f.repo.join("sub")] {
        let out = lane_env(cwd, &f.home, &args);
        assert_eq!(out.status.code(), Some(0), "{cwd:?}: {out:?}");
        shapes.push(String::from_utf8_lossy(&out.stdout).into_owned());
    }
    assert_eq!(
        names_under(&f.home),
        Vec::<String>::new(),
        "HOME stays empty"
    );

    // No HOME at all, and a relative one: the same lines, nothing made under the checkout.
    let before = names_under(&f.repo);
    for home in [None, Some(Path::new("rel"))] {
        let mut c = Command::new(common::MOAI_BIN);
        c.clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .current_dir(&f.repo)
            .arg("lane-env")
            .args(args);
        if let Some(h) = home {
            c.env("HOME", h);
        }
        let out = c.output().unwrap();
        assert_eq!(out.status.code(), Some(0), "{home:?}: {out:?}");
        assert_eq!(String::from_utf8_lossy(&out.stdout), shapes[1], "{home:?}");
    }
    assert_eq!(
        names_under(&f.repo),
        before,
        "nothing made under the checkout"
    );
}

#[test]
fn refusals_print_nothing_on_stdout() {
    let f = fx();
    let base_s = f.base.to_str().unwrap();
    let not_a_repo = f.root.join("not a repo");
    fs::create_dir(&not_a_repo).unwrap();

    for (cwd, args, says) in [
        (
            &not_a_repo,
            vec!["--lane", "agent-7", "--base", base_s],
            "git rev-parse",
        ),
        (
            &f.repo,
            vec!["--lane", "../x", "--base", base_s],
            "invalid lane name",
        ),
        (
            &f.repo,
            vec!["--lane", "a b", "--base", base_s],
            "invalid lane name",
        ),
        (
            &f.repo,
            vec!["--lane", "agent-8", "--base", "tmp"],
            "not an absolute path",
        ),
    ] {
        let out = lane_env(cwd, &f.home, &args);
        assert_eq!(out.status.code(), Some(1), "{args:?}: {out:?}");
        assert!(out.stdout.is_empty(), "{args:?}: stdout must stay empty");
        let err = String::from_utf8_lossy(&out.stderr).into_owned();
        assert!(err.contains(says), "{args:?}: {err}");
    }
    assert_eq!(names_under(&f.home), Vec::<String>::new());
    // The documented line stops at the failed substitution's `&&`: nothing is exported.
    let (code, out) = evaluated(&not_a_repo, &f.home, &f.base, "agent-7");
    assert_ne!(code, 0);
    assert_eq!(out, "");
}

#[test]
fn a_missing_lane_is_a_usage_error() {
    let f = fx();
    let out = lane_env(&f.repo, &f.home, &[]);
    assert_eq!(out.status.code(), Some(2));
    assert!(out.stdout.is_empty());
    assert!(String::from_utf8_lossy(&out.stderr).contains("--lane"));
}

/// The comment line that says where cargo builds: the worktree's own `target/`, whichever lane
/// asks and from whichever folder of the worktree.
#[test]
fn the_comment_names_the_worktrees_own_target() {
    let f = fx();
    let one = f.worktree(&f.root.join("wt-one"));
    let two = f.worktree(&f.root.join("wt-two"));
    let says = |lines: &[String]| {
        let hits: Vec<_> = lines
            .iter()
            .filter(|l| l.starts_with("# CARGO_TARGET_DIR"))
            .cloned()
            .collect();
        assert_eq!(hits.len(), 1, "{lines:?}");
        hits[0].clone()
    };
    let a = says(&lines_from(&one, &f.home, "agent-7", &f.base));
    let b = says(&lines_from(&two, &f.home, "agent-7", &f.base));
    assert!(a.contains(&one.join("target").display().to_string()), "{a}");
    assert!(b.contains(&two.join("target").display().to_string()), "{b}");
    assert_eq!(
        a,
        says(&lines_from(&one, &f.home, "agent-8", &f.base)),
        "the lane does not name the folder"
    );
    fs::create_dir(one.join("sub")).unwrap();
    assert_eq!(
        a,
        says(&lines_from(&one.join("sub"), &f.home, "agent-7", &f.base))
    );
}

/// #2930 R-1: every lane build runs with incremental compilation off (exactly one line).
#[test]
fn lane_env_exports_cargo_incremental_off() {
    let f = fx();
    let lines = lines_from(&f.repo, &f.home, "agent-7", &f.base);
    let count = lines
        .iter()
        .filter(|l| *l == "export CARGO_INCREMENTAL=0")
        .count();
    assert_eq!(count, 1, "{lines:?}");
}

/// The lane-name limit at its edge: 64 characters pass, 65 are refused with stdout empty.
#[test]
fn lane_name_length_edge() {
    let f = fx();
    let longest = "x".repeat(64);
    let lines = lines_from(&f.repo, &f.home, &longest, &f.base);
    assert!(
        lines.iter().any(|l| l == "unset CARGO_TARGET_DIR"),
        "{lines:?}"
    );

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
