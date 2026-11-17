//! SPEC-20261117-spec-before-code (card C-86, slice 1; user decision 2026-11-17 "권고 말고 강제"):
//! on a SPEC work branch `spec/<yyyymmdd>-<slug>`, tool-preflight refuses a file write inside
//! the work tree outside `.moai/specs/` until `.moai/specs/SPEC-<yyyymmdd>-<slug>/spec.md`
//! exists and names that id. Every case runs the real binary in a temporary repository with a
//! temporary HOME (as `tests/hook_dispatch_cli.rs`); a real host session is not run here.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const BRANCH: &str = "spec/20261117-demo";
const SPEC: &str = ".moai/specs/SPEC-20261117-demo/spec.md";

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

fn preflight(host: &str, cwd: &Path, input: &Value) -> Run {
    let mut cmd = Command::new(BIN);
    cmd.args(["hook-dispatch", "--host", host, "--event", "tool-preflight"])
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

fn git(cwd: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

/// A temporary repository on `branch` with one commit and its own HOME.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", branch]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    fs::write(root.join("a.txt"), "a\n").unwrap();
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    git(&root, &["add", "a.txt", ".gitignore"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    (dir, root)
}

fn write(path: &str, cwd: &Path) -> Value {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
           "tool_input": {"file_path": path, "content": "fn main() {}\n"}, "cwd": cwd})
}

fn spec_text(id: &str) -> String {
    format!("---\nid: {id}\ntitle: \"demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-17\n---\n\n## Intent\n\nx\n")
}

fn assert_deny(r: &Run) -> String {
    assert_eq!(r.code, 0, "deny exits 0: {}", r.stderr);
    let v: Value = serde_json::from_str(r.stdout.trim()).expect("stdout is one JSON object");
    assert_eq!(v["hookSpecificOutput"]["permissionDecision"], "deny", "{v}");
    v["hookSpecificOutput"]["permissionDecisionReason"]
        .as_str()
        .unwrap()
        .to_string()
}

fn assert_allow(r: &Run) {
    assert_eq!((r.code, r.stdout.as_str()), (0, ""), "allow: {}", r.stderr);
}

/// SPEC-20261117-spec-before-code AC-1: a code write waits for the SPEC; the SPEC itself can
/// always be written, and once it names the branch's id the code write goes ahead.
#[test]
fn a_code_write_on_a_spec_branch_waits_for_the_branch_spec() {
    let (_d, root) = repo(BRANCH);
    let reason = assert_deny(&preflight("claude", &root, &write("src/main.rs", &root)));
    assert!(reason.starts_with("MoAI tool-preflight:"), "{reason}");
    assert!(
        reason.contains(".moai/specs/SPEC-<yyyymmdd>-<slug>/spec.md"),
        "{reason}"
    );
    // The reason is a fixed string: no path or branch from the input is echoed.
    assert!(
        !reason.contains("main.rs") && !reason.contains("demo"),
        "{reason}"
    );
    // C-86 audit N4: it also names the way to give the SPEC the branch's date.
    assert!(
        reason.contains("`moai spec new --date YYYY-MM-DD`"),
        "{reason}"
    );
    for path in ["src/main.rs", "README.md", "tests/a.rs", "a.txt"] {
        assert_deny(&preflight("claude", &root, &write(path, &root)));
    }
    // Writing the SPEC (and anything else under .moai/specs/) is never refused by this rule.
    assert_allow(&preflight("claude", &root, &write(SPEC, &root)));
    assert_allow(&preflight(
        "claude",
        &root,
        &write(".moai/specs/SPEC-20261117-demo/notes.md", &root),
    ));
    fs::create_dir_all(root.join(".moai/specs/SPEC-20261117-demo")).unwrap();
    fs::write(root.join(SPEC), spec_text("SPEC-20261117-demo")).unwrap();
    assert_allow(&preflight("claude", &root, &write("src/main.rs", &root)));
    let absolute = root.join("src/main.rs");
    assert_allow(&preflight(
        "claude",
        &root,
        &write(absolute.to_str().unwrap(), &root),
    ));
}

/// AC-2: a SPEC file that names another id, or a folder without spec.md, does not count.
#[test]
fn a_spec_of_another_id_or_a_bare_folder_does_not_count() {
    let (_d, root) = repo(BRANCH);
    let folder = root.join(".moai/specs/SPEC-20261117-demo");
    fs::create_dir_all(&folder).unwrap();
    assert_deny(&preflight("claude", &root, &write("src/main.rs", &root)));
    fs::write(root.join(SPEC), spec_text("SPEC-20261117-other")).unwrap();
    assert_deny(&preflight("claude", &root, &write("src/main.rs", &root)));
    // An `id:` line after the frontmatter is not the frontmatter's id.
    fs::write(
        root.join(SPEC),
        "---\ntitle: x\n---\n\nid: SPEC-20261117-demo\n",
    )
    .unwrap();
    assert_deny(&preflight("claude", &root, &write("src/main.rs", &root)));
    // A folder named spec.md is not the file.
    fs::remove_file(root.join(SPEC)).unwrap();
    fs::create_dir_all(root.join(SPEC)).unwrap();
    assert_deny(&preflight("claude", &root, &write("src/main.rs", &root)));
}

/// AC-3: a path that leaves `.moai/specs/` through `..` or a symbolic link is judged where it
/// lands.
#[test]
fn a_path_that_leaves_the_specs_folder_is_judged_where_it_lands() {
    let (_d, root) = repo(BRANCH);
    fs::create_dir_all(root.join(".moai/specs")).unwrap();
    fs::create_dir_all(root.join("src")).unwrap();
    assert_deny(&preflight(
        "claude",
        &root,
        &write(".moai/specs/../../src/main.rs", &root),
    ));
    std::os::unix::fs::symlink(root.join("src"), root.join(".moai/specs/out")).unwrap();
    assert_deny(&preflight(
        "claude",
        &root,
        &write(".moai/specs/out/main.rs", &root),
    ));
}

/// AC-4: other branches, writes outside the work tree and Bash are not this rule's.
#[test]
fn other_branches_outside_writes_and_bash_are_left_alone() {
    for branch in [
        "main",
        "issue/77-k",
        "spec/notadate-demo",
        "spec/2026111-demo",
        "spec/20261117-",
        "spec/20261117-Demo",
        "spec/20261117-demo/sub",
        "card/86-x",
    ] {
        let (_d, root) = repo(branch);
        assert_allow(&preflight("claude", &root, &write("src/main.rs", &root)));
    }
    let (_d, root) = repo(BRANCH);
    let outside = tempfile::TempDir::new().unwrap();
    let elsewhere = outside.path().join("x.rs");
    assert_allow(&preflight(
        "claude",
        &root,
        &write(elsewhere.to_str().unwrap(), &root),
    ));
    let bash = json!({"hook_event_name": "PreToolUse", "tool_name": "Bash",
                      "tool_input": {"command": "echo hi > src/main.rs"}, "cwd": root});
    assert_allow(&preflight("claude", &root, &bash));
}

/// AC-5: the rule reads a linked worktree's branch, and covers Edit, MultiEdit and Codex
/// `apply_patch` the same way.
#[test]
fn linked_worktrees_and_every_write_tool_are_covered() {
    let (_d, main) = repo("main");
    let parent = tempfile::TempDir::new().unwrap();
    let wt = fs::canonicalize(parent.path()).unwrap().join("wt");
    git(
        &main,
        &["worktree", "add", "-q", "-b", BRANCH, wt.to_str().unwrap()],
    );
    fs::create_dir_all(wt.join("home")).unwrap();
    assert_deny(&preflight("claude", &wt, &write("src/main.rs", &wt)));
    let edit = json!({"hook_event_name": "PreToolUse", "tool_name": "Edit",
                      "tool_input": {"file_path": "a.txt", "old_string": "a", "new_string": "b"},
                      "cwd": wt});
    assert_deny(&preflight("claude", &wt, &edit));
    let multi = json!({"hook_event_name": "PreToolUse", "tool_name": "MultiEdit",
                       "tool_input": {"file_path": "a.txt",
                                      "edits": [{"old_string": "a", "new_string": "b"}]},
                       "cwd": wt});
    assert_deny(&preflight("claude", &wt, &multi));
    let patch = json!({"hook_event_name": "PreToolUse", "tool_name": "apply_patch",
                       "tool_input": {"command": "*** Begin Patch\n*** Update File: a.txt\n@@\n-a\n+b\n*** End Patch"},
                       "cwd": wt});
    let r = preflight("codex", &wt, &patch);
    assert!(
        r.code == 2 || r.stdout.contains("\"deny\""),
        "codex: {} {}",
        r.code,
        r.stdout
    );
    fs::create_dir_all(wt.join(".moai/specs/SPEC-20261117-demo")).unwrap();
    fs::write(wt.join(SPEC), spec_text("SPEC-20261117-demo")).unwrap();
    assert_allow(&preflight("claude", &wt, &edit));
    assert_allow(&preflight("codex", &wt, &patch));
    // The primary checkout on `main` is untouched by its worktree's branch.
    assert_allow(&preflight("claude", &main, &write("src/main.rs", &main)));
}

/// `to` written relative to the folder `from` (both absolute real paths).
fn relative(from: &Path, to: &Path) -> PathBuf {
    let (a, b): (Vec<_>, Vec<_>) = (from.components().collect(), to.components().collect());
    let common = a.iter().zip(&b).take_while(|(x, y)| x == y).count();
    let mut out = PathBuf::new();
    for _ in common..a.len() {
        out.push("..");
    }
    for part in &b[common..] {
        out.push(part);
    }
    out
}

/// C-86 audit N1: a linked worktree whose `.git` file names its git folder with a RELATIVE
/// `gitdir:` (git `worktree.useRelativePaths`, written by hand here) is read from the work
/// tree, not from the hook's working folder: the session sits in a subfolder, so a path joined
/// against the process folder would miss the branch and let the write through.
#[test]
fn a_relative_gitdir_is_read_from_the_work_tree() {
    let (_d, main) = repo("main");
    let parent = tempfile::TempDir::new().unwrap();
    let wt = fs::canonicalize(parent.path()).unwrap().join("wt");
    git(
        &main,
        &["worktree", "add", "-q", "-b", BRANCH, wt.to_str().unwrap()],
    );
    let pointer = fs::read_to_string(wt.join(".git")).unwrap();
    let target = PathBuf::from(pointer.trim().strip_prefix("gitdir:").unwrap().trim());
    let rel = relative(&wt, &target);
    assert!(rel.is_relative(), "{}", rel.display());
    fs::write(wt.join(".git"), format!("gitdir: {}\n", rel.display())).unwrap();
    let sub = wt.join("src");
    for cwd in [&wt, &sub] {
        fs::create_dir_all(cwd.join("home")).unwrap();
        let file = wt.join("src/main.rs");
        assert_deny(&preflight(
            "claude",
            cwd,
            &write(file.to_str().unwrap(), cwd),
        ));
    }
    fs::create_dir_all(wt.join(".moai/specs/SPEC-20261117-demo")).unwrap();
    fs::write(wt.join(SPEC), spec_text("SPEC-20261117-demo")).unwrap();
    let file = wt.join("src/main.rs");
    assert_allow(&preflight(
        "claude",
        &sub,
        &write(file.to_str().unwrap(), &sub),
    ));
}

/// C-86 audit N2: the work tree is the one holding the WRITTEN file (the nearest `.git`, a
/// folder or a file, above it), not the one holding the session's folder. A worktree nested in
/// the primary checkout (`.claude/worktrees/<x>`) has its own branch, in both directions.
#[test]
fn the_written_file_decides_the_work_tree_in_both_directions() {
    // The primary checkout on a SPEC branch without its SPEC; a nested worktree on `work`.
    let (_d, primary) = repo(BRANCH);
    let nested = primary.join(".claude/worktrees/x");
    git(
        &primary,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "work",
            nested.to_str().unwrap(),
        ],
    );
    fs::create_dir_all(nested.join("home")).unwrap();
    let in_nested = nested.join("src/lib.rs").display().to_string();
    let in_primary = primary.join("src/lib.rs").display().to_string();
    // From the primary: the nested worktree's file is on `work` and goes ahead.
    assert_allow(&preflight("claude", &primary, &write(&in_nested, &primary)));
    assert_deny(&preflight(
        "claude",
        &primary,
        &write(&in_primary, &primary),
    ));
    // From the nested worktree: the primary's file is on the SPEC branch and waits.
    assert_deny(&preflight("claude", &nested, &write(&in_primary, &nested)));
    assert_allow(&preflight("claude", &nested, &write(&in_nested, &nested)));

    // The other way round: the primary on `main`, a nested worktree on the SPEC branch.
    let (_e, main) = repo("main");
    let spec_wt = main.join(".claude/worktrees/y");
    git(
        &main,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            BRANCH,
            spec_wt.to_str().unwrap(),
        ],
    );
    fs::create_dir_all(spec_wt.join("home")).unwrap();
    let in_spec_wt = spec_wt.join("src/lib.rs").display().to_string();
    let in_main = main.join("src/lib.rs").display().to_string();
    assert_deny(&preflight("claude", &main, &write(&in_spec_wt, &main)));
    assert_allow(&preflight("claude", &main, &write(&in_main, &main)));
    assert_allow(&preflight("claude", &spec_wt, &write(&in_main, &spec_wt)));
    // Its SPEC lets its code through, from either folder.
    fs::create_dir_all(spec_wt.join(".moai/specs/SPEC-20261117-demo")).unwrap();
    fs::write(spec_wt.join(SPEC), spec_text("SPEC-20261117-demo")).unwrap();
    assert_allow(&preflight("claude", &main, &write(&in_spec_wt, &main)));
}
