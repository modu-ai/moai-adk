//! #2337: every git subprocess the product starts must go through
//! `identity::scrub_git_env` (the #2033 / #2043 / #2054 class). An inherited `GIT_DIR`,
//! `GIT_WORK_TREE` or `GIT_INDEX_FILE` (set by a host hook, a worktree, or an outer git
//! command) would otherwise point git at a different repository than the one asked for.
//!
//! Two checks:
//! - a repository-wide scan: each `Command::new("git")` in `src/` must reach
//!   `scrub_git_env(` before the command runs (`.output()`, `.status()` or `.spawn()`);
//! - a behaviour repro on the claim guard, where a contaminated `GIT_DIR` turned a refusal
//!   into a pass (fail-open).
mod common;

use serde_json::json;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

/// Every `.rs` file under `dir`, sorted.
fn rust_files(dir: &Path, out: &mut Vec<PathBuf>) {
    let mut entries: Vec<_> = fs::read_dir(dir).unwrap().flatten().collect();
    entries.sort_by_key(|e| e.path());
    for e in entries {
        let p = e.path();
        if p.is_dir() {
            rust_files(&p, out);
        } else if p.extension().is_some_and(|x| x == "rs") {
            out.push(p);
        }
    }
}

/// `path:line` of each git spawn that runs without `scrub_git_env(` in between, plus the
/// number of git spawns seen (so an empty scan cannot pass).
fn unscrubbed_git_spawns(src: &Path) -> (Vec<String>, usize) {
    let mut files = Vec::new();
    rust_files(src, &mut files);
    let mut missing = Vec::new();
    let mut seen = 0;
    for file in files {
        let text = fs::read_to_string(&file).unwrap();
        let lines: Vec<&str> = text.lines().collect();
        for (i, line) in lines.iter().enumerate() {
            let code = line.split("//").next().unwrap_or("");
            if !code.contains(r#"Command::new("git")"#) {
                continue;
            }
            seen += 1;
            // The window ends at the first line that runs the command.
            let mut scrubbed = false;
            for later in &lines[i..] {
                let later = later.split("//").next().unwrap_or("");
                if later.contains("scrub_git_env(") {
                    scrubbed = true;
                    break;
                }
                if later.contains(".output()")
                    || later.contains(".status()")
                    || later.contains(".spawn()")
                {
                    break;
                }
            }
            if !scrubbed {
                let rel = file.strip_prefix(src.parent().unwrap()).unwrap();
                missing.push(format!("{}:{}", rel.display(), i + 1));
            }
        }
    }
    (missing, seen)
}

#[test]
fn every_git_subprocess_in_src_is_scrubbed() {
    let src = Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
    let (missing, seen) = unscrubbed_git_spawns(&src);
    // At 44c0bc063 there are 12 git spawn sites; a scan that finds none checks nothing.
    assert!(
        seen >= 12,
        "only {seen} git spawn sites found; the scan is broken"
    );
    assert!(
        missing.is_empty(),
        "git subprocesses without scrub_git_env (#2337): {missing:?}"
    );
}

/// The scan itself flags an unscrubbed spawn and accepts a scrubbed one, so a pass above
/// means the source is clean, not that the matcher never fires.
#[test]
fn the_scan_flags_an_unscrubbed_spawn_and_accepts_a_scrubbed_one() {
    let tmp = tempfile::tempdir().unwrap();
    let src = tmp.path().join("src");
    fs::create_dir_all(&src).unwrap();
    fs::write(
        src.join("bad.rs"),
        "fn f() {\n    let out = Command::new(\"git\")\n        .arg(\"status\")\n        .output();\n}\n",
    )
    .unwrap();
    fs::write(
        src.join("good.rs"),
        "fn g() {\n    let mut cmd = Command::new(\"git\");\n    cmd.arg(\"status\");\n    crate::identity::scrub_git_env(&mut cmd);\n    let out = cmd.output();\n}\n",
    )
    .unwrap();
    let (missing, seen) = unscrubbed_git_spawns(&src);
    assert_eq!(seen, 2);
    assert_eq!(missing, vec!["src/bad.rs:2".to_string()]);
}

fn git(dir: &Path, args: &[&str]) {
    // Keep the env-isolation probe and identity while applying the shared fixture guards.
    let mut cmd = Command::new("git");
    cmd.args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
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
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid");
    let out = cmd.output().unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A repository on `branch` with one commit and a GitHub `origin`.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    git(&root, &["commit", "-q", "--allow-empty", "-m", "base"]);
    git(
        &root,
        &[
            "remote",
            "add",
            "origin",
            "https://github.com/acme/widgets.git",
        ],
    );
    if branch != "main" {
        git(&root, &["checkout", "-q", "-b", branch]);
    }
    (dir, root)
}

/// AC-2 repro: a host hook that inherits `GIT_DIR` pointing at another repository. The
/// lane worktree is on `issue/7-widget` and the commit carries no claim token, so the guard
/// must refuse it (exit 2). Without the scrub, git reads the other repository's branch
/// (`main`), the guard decides "not a lane worktree" and lets the commit through.
#[cfg(unix)]
#[test]
fn claim_guard_ignores_an_inherited_git_dir() {
    let (_lane_dir, lane) = repo("issue/7-widget");
    let (_other_dir, other) = repo("main");
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                       "tool_input": {"command": "git commit -m work"}, "cwd": lane})
    .to_string();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["claim-verify", ".", "--hook"])
        .current_dir(&lane)
        // The refusal comes before any GitHub read, so no `gh` stub is needed.
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_DIR", other.join(".git"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.as_bytes());
    let out = child.wait_with_output().unwrap();
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert_eq!(
        out.status.code(),
        Some(2),
        "a contaminated GIT_DIR must not open the guard; stdout: {} stderr: {stderr}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(stderr.contains("claim verify: uncertain"), "{stderr}");
}
