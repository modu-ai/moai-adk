//! #2350 / #2337: the statusline's git read is scrubbed like every product git call. An
//! inherited `GIT_DIR` / `GIT_WORK_TREE` / `GIT_INDEX_FILE` pointing at another repository
//! (set by a host hook or an outer git command) must not change the rendered branch or
//! counts: the render describes the repository of the session directory only.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use std::{
    fs,
    io::Write as _,
    path::Path,
    process::{Command, Stdio},
};

fn git(dir: &Path, args: &[&str]) {
    let status = common::git::command()
        .args(args)
        .current_dir(dir)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_INDEX_FILE")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?}");
}

/// One plain render with `stdin` naming `project`; `budget_ms` is the hidden render-budget
/// seam (a large value keeps this independent of machine load).
fn render(home: &Path, project: &Path, extra: &[(&str, &Path)], budget_ms: u64) -> String {
    let stdin = format!(
        r#"{{"model":{{"display_name":"Opus"}},"workspace":{{"current_dir":"{0}","project_dir":"{0}"}}}}"#,
        project.display()
    );
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.args(["statusline", "--render-budget-ms", &budget_ms.to_string()])
        .current_dir(project)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("HOME", home)
        .env("MOAI_HOME", home.join(".moai"))
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped());
    for (k, v) in extra {
        cmd.env(k, v);
    }
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success());
    String::from_utf8(out.stdout).unwrap()
}

#[test]
fn an_inherited_git_dir_cannot_change_the_rendered_branch_or_counts() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(home.join(".moai")).unwrap();

    // The session's repository: branch `main`, clean.
    let project = root.join("project");
    fs::create_dir(&project).unwrap();
    git(&project, &["init", "-q", "-b", "main"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git(&project, &["add", "a.txt"]);
    git(&project, &["commit", "-q", "-m", "a"]);

    // Another repository: branch `intruder`, one modified and two untracked files.
    let other = root.join("other");
    fs::create_dir(&other).unwrap();
    git(&other, &["init", "-q", "-b", "intruder"]);
    fs::write(other.join("b.txt"), "b\n").unwrap();
    git(&other, &["add", "b.txt"]);
    git(&other, &["commit", "-q", "-m", "b"]);
    fs::write(other.join("b.txt"), "changed\n").unwrap();
    fs::write(other.join("u1.txt"), "u\n").unwrap();
    fs::write(other.join("u2.txt"), "u\n").unwrap();

    let clean = render(&home, &project, &[], WITHIN);
    assert!(clean.contains("🅱️ main"), "{clean}");
    assert!(clean.contains("💾 +0 M0 ?0"), "{clean}");
    // #2756: the repository line is the second of the three lines, not the last.
    let repository_line = clean.lines().find(|l| l.starts_with("🅱️")).unwrap();
    assert_eq!(repository_line, "🅱️ main | 💾 +0 M0 ?0", "{clean}");
    assert!(!clean.contains("📁"), "{clean}");

    let git_dir = other.join(".git");
    let index = git_dir.join("index");
    let contaminated = render(
        &home,
        &project,
        &[
            ("GIT_DIR", git_dir.as_path()),
            ("GIT_WORK_TREE", other.as_path()),
            ("GIT_INDEX_FILE", index.as_path()),
        ],
        WITHIN,
    );
    assert_eq!(contaminated, clean);
    assert!(!contaminated.contains("intruder"), "{contaminated}");
}

/// A budget no git call can miss under any load.
const WITHIN: u64 = 3_600_000;

/// The over-budget path, deterministically: with an already expired budget no git is
/// started, so the repository row is omitted entirely. Without the seam (budget ignored)
/// git would run and this fails.
#[test]
fn an_expired_render_budget_drops_the_git_segment() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    fs::create_dir_all(home.join(".moai")).unwrap();
    let project = root.join("project");
    fs::create_dir(&project).unwrap();
    git(&project, &["init", "-q", "-b", "main"]);

    let within = render(&home, &project, &[], WITHIN);
    // #2756: the repository line sits between the environment and usage lines.
    assert_eq!(
        within.lines().nth(1),
        Some("🅱️ main | 💾 +0 M0 ?0"),
        "{within}"
    );
    for _ in 0..5 {
        let over = render(&home, &project, &[], 0);
        assert!(
            !over.contains("🅱️") && !over.contains("💾") && !over.contains("📁"),
            "{over}"
        );
        let within_without_repo: Vec<_> = within.lines().filter(|l| !l.starts_with("🅱️")).collect();
        let over_lines: Vec<_> = over.lines().collect();
        assert_eq!(over_lines, within_without_repo, "{over}");
    }
}
