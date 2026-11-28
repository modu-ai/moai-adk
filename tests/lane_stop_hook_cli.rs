//! SPEC-20261121-lane-goal-leader-codex-removed: the lane Stop hook was removed with its host,
//! but a hook file an earlier release wrote still runs `lane-stop-hook`. Such a run must neither
//! block nor fail: one line on standard error, nothing on standard output, exit 0, and nothing
//! written in the project or the HOME. Standard input is never read.
mod common;

use common::CleanEnv;
use std::fs;
use std::io::Read as _;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// Every path below `dir`, relative and sorted.
fn listing(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            if e.file_type().unwrap().is_dir() {
                stack.push(p.clone());
            }
            out.push(p.strip_prefix(dir).unwrap().to_path_buf());
        }
    }
    out.sort();
    out
}

/// Runs the command with standard input held open and never written, and fails when it does not
/// end by itself: a command that read standard input would hang here.
fn run_with_open_stdin(root: &Path, args: &[&str]) -> (Option<i32>, String, String) {
    let mut child = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", root.join("home"))
        .env("PATH", "/usr/bin:/bin")
        .current_dir(root.join("project"))
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let held = child.stdin.take().unwrap();
    let start = Instant::now();
    let status = loop {
        if let Some(status) = child.try_wait().unwrap() {
            break status;
        }
        if start.elapsed() > Duration::from_secs(60) {
            child.kill().unwrap();
            child.wait().unwrap();
            panic!("{args:?} did not end while standard input was open");
        }
        std::thread::sleep(Duration::from_millis(20));
    };
    drop(held);
    let (mut out, mut err) = (String::new(), String::new());
    child
        .stdout
        .take()
        .unwrap()
        .read_to_string(&mut out)
        .unwrap();
    child
        .stderr
        .take()
        .unwrap()
        .read_to_string(&mut err)
        .unwrap();
    (status.code(), out, err)
}

/// The command lines old hook files carry (with either flag set, with none, and with values the
/// old command refused) all get the same one-line answer.
#[test]
fn an_old_stop_hook_gets_one_line_and_exit_zero_and_writes_nothing() {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir(root.join("home")).unwrap();
    let project = root.join("project");
    // The state folder the old command counted into.
    fs::create_dir_all(project.join(moai_cli::state_root::dir(&project))).unwrap();
    let before = listing(&root);
    for args in [
        &["lane-stop-hook", "--lane", "agent-1", "--repo", "o/r"][..],
        &[
            "lane-stop-hook",
            "--lane",
            "agent-1",
            "--repo",
            "o/r",
            "--snapshot",
            "missing.json",
        ],
        &["lane-stop-hook", "--lane", "a b", "--repo", "no-slash"],
        &["lane-stop-hook"],
    ] {
        let (code, out, err) = run_with_open_stdin(&root, args);
        assert_eq!(code, Some(0), "{args:?}: {err}");
        assert_eq!(out, "", "{args:?}");
        assert_eq!(
            err,
            "moai lane-stop-hook: the lane Stop hook was removed; MoAI serves Claude Code only\n",
            "{args:?}"
        );
        assert_eq!(listing(&root), before, "{args:?} wrote something");
    }
}
