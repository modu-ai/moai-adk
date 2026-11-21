//! The git-hook feature of SPEC-20261121-go-path-removal: recognition, the retire-install
//! lifecycle and the hook texts, driven in temporary repositories. The feature module compiles
//! into this test crate by path (the same shape the unwired quality-gate slice used), so it
//! needs no `lib.rs` or `main.rs` line yet; the update layer and the CLI verbs follow the pair
//! requests on the cards that hold those files.
//! Every `git` call goes through `common::git` (the fixture runner), and every hook script
//! runs with a temporary `PATH` — never the real HOME.

#![cfg(unix)]

#[path = "../src/git_hooks/mod.rs"]
mod git_hooks;

mod common;

use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

// A representative Go-generation pre-commit: the real header lines plus a stub body — the
// feature recognises hooks by their marks, not by their length.
const GO_PRE_COMMIT: &str = "#!/bin/sh\n\
# MoAI-ADK pre-commit hook - fast subset (gofmt + go vet) + opt-in heavy gate (moai gate)\n\
# Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit\n\
exit 0\n";
const GO_PRE_PUSH: &str = "#!/bin/sh\n\
# MoAI-ADK pre-push hook - commit message convention via moai hook pre-push\n\
exit 0\n";
const RUST_COMMIT_MSG: &str = "#!/bin/sh\n\
# MoAI commit-msg hook (Rust): refuse card numbers in the commit message.\n\
exit 0\n";
const USER_COMMIT_MSG: &str = "#!/bin/sh\n# my own commit-msg hook, not MoAI's\necho kept\n";
const HASH_LINE: &str =
    "3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac  pre-commit\n";

fn make_repo() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().expect("temp dir");
    let root = dir.path().join("repo");
    fs::create_dir_all(&root).expect("repo folder");
    common::git::git_ok(&root, &["init"]);
    common::git::git_in(
        &root,
        &["commit", "--allow-empty", "-m", "root: first commit"],
    );
    (dir, root)
}

fn hooks_dir(root: &Path) -> PathBuf {
    root.join(".git").join("hooks")
}

fn write_hook(root: &Path, name: &str, text: &str) {
    let path = hooks_dir(root).join(name);
    fs::write(&path, text).expect("write hook");
    make_executable(&path);
}

fn write_hash_file(root: &Path, name: &str) {
    fs::write(
        hooks_dir(root).join(format!(".moai-{name}.sha256")),
        HASH_LINE,
    )
    .expect("write hash file");
}

fn make_executable(path: &Path) {
    use std::os::unix::fs::PermissionsExt;
    fs::set_permissions(path, fs::Permissions::from_mode(0o755)).expect("chmod");
}

fn bin_dir() -> PathBuf {
    PathBuf::from(common::MOAI_BIN)
        .parent()
        .expect("binary folder")
        .to_path_buf()
}

fn path_with_moai() -> String {
    format!(
        "{}:{}",
        bin_dir().display(),
        std::env::var("PATH").unwrap_or_default()
    )
}

fn path_without_moai() -> String {
    "/usr/bin:/bin".to_string()
}

fn names(value: &serde_json::Value, key: &str) -> Vec<String> {
    value[key]
        .as_array()
        .map(|rows| {
            rows.iter()
                .map(|r| r["name"].as_str().unwrap_or_default().to_string())
                .collect()
        })
        .unwrap_or_default()
}

// --- recognition -------------------------------------------------------------

#[test]
fn recognition_hash_file_marks_a_go_generation_hook() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", "#!/bin/sh\necho plain\n");
    write_hash_file(&root, "pre-commit");
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(names(&plan, "go_hooks"), vec!["pre-commit".to_string()]);
    assert_eq!(
        plan["go_hooks"][0]["evidence"], "hash file",
        "the hash file is the evidence"
    );
}

#[test]
fn recognition_header_marks_a_go_generation_hook_without_a_hash_file() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(names(&plan, "go_hooks"), vec!["pre-commit".to_string()]);
    assert_eq!(plan["go_hooks"][0]["evidence"], "header");
}

#[test]
fn recognition_a_rust_marked_hook_is_current_and_not_retired() {
    let (_dir, root) = make_repo();
    write_hook(&root, "commit-msg", RUST_COMMIT_MSG);
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(
        names(&plan, "go_hooks").is_empty(),
        "a (Rust) hook is not Go generation: {}",
        plan["go_hooks"]
    );
    assert!(
        !names(&plan, "kept").contains(&"commit-msg".to_string()),
        "a MoAI hook is not a user hook"
    );
}

#[test]
fn recognition_a_hook_without_moai_marks_is_kept() {
    let (_dir, root) = make_repo();
    write_hook(&root, "commit-msg", USER_COMMIT_MSG);
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(names(&plan, "kept"), vec!["commit-msg".to_string()]);
    assert!(names(&plan, "go_hooks").is_empty());
    assert!(
        !names(&plan, "install").contains(&"commit-msg".to_string()),
        "an occupied name is not installed: {}",
        plan["install"]
    );
}

// --- lifecycle ---------------------------------------------------------------

#[test]
fn lifecycle_retires_go_hooks_and_installs_the_rust_hooks() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hook(&root, "pre-push", GO_PRE_PUSH);
    write_hash_file(&root, "pre-commit");
    write_hash_file(&root, "pre-push");
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(plan["writes"], true);
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T19:40:00Z").expect("apply");
    assert_eq!(receipt["state"], "applied");
    assert_eq!(
        names(&receipt, "retired"),
        vec!["pre-commit".to_string(), "pre-push".to_string()]
    );
    for name in ["pre-commit", "pre-push", "commit-msg"] {
        let installed = fs::read_to_string(hooks_dir(&root).join(name)).expect("installed");
        let shipped = match name {
            "pre-commit" => git_hooks::PRE_COMMIT_TEXT,
            "pre-push" => git_hooks::PRE_PUSH_TEXT,
            _ => git_hooks::COMMIT_MSG_TEXT,
        };
        assert_eq!(installed, shipped, "{name} carries the shipped bytes");
    }
    assert!(
        !hooks_dir(&root).join(".moai-pre-commit.sha256").exists(),
        "the hash file moved away"
    );
    assert!(
        !hooks_dir(&root).join(".moai-pre-push.sha256").exists(),
        "the hash file moved away"
    );
    let stamp = receipt["stamp"].as_str().expect("stamp").to_string();
    let folder = hooks_dir(&root).join(format!("moai-retired-{stamp}"));
    let saved = fs::read_to_string(folder.join("pre-commit")).expect("saved hook");
    assert_eq!(saved, GO_PRE_COMMIT, "the retire folder keeps the bytes");
    assert!(
        folder.join(".moai-pre-commit.sha256").exists(),
        "the hash file is kept for the revert"
    );
}

#[test]
fn lifecycle_keeps_a_user_hook_byte_for_byte() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hook(&root, "commit-msg", USER_COMMIT_MSG);
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T19:40:00Z").expect("apply");
    assert_eq!(
        names(&receipt, "kept"),
        vec!["commit-msg".to_string()],
        "the user hook is listed as kept"
    );
    let after = fs::read_to_string(hooks_dir(&root).join("commit-msg")).expect("read back");
    assert_eq!(after, USER_COMMIT_MSG, "the user hook keeps its bytes");
    assert!(
        !names(&receipt, "installed").contains(&"commit-msg".to_string()),
        "an occupied name is not installed"
    );
}

#[test]
fn lifecycle_a_second_run_changes_nothing() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    git_hooks::apply(&root, &token, "2026-11-21T19:40:00Z").expect("first apply");
    let second_plan = git_hooks::plan(&root).expect("second plan");
    assert_eq!(second_plan["writes"], false, "nothing left to do");
    assert!(names(&second_plan, "go_hooks").is_empty());
    assert!(names(&second_plan, "install").is_empty());
    let again = git_hooks::apply(&root, &token, "2026-11-21T19:41:00Z");
    assert!(again.is_err(), "a second apply refuses: {again:?}");
}

#[test]
fn lifecycle_revert_restores_the_go_hooks() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hook(&root, "pre-push", GO_PRE_PUSH);
    write_hash_file(&root, "pre-commit");
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T19:40:00Z").expect("apply");
    let stamp = receipt["stamp"].as_str().expect("stamp").to_string();
    git_hooks::revert(&root, &stamp).expect("revert");
    let pre_commit = fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("restored");
    assert_eq!(
        pre_commit, GO_PRE_COMMIT,
        "the Go hook comes back byte for byte"
    );
    let pre_push = fs::read_to_string(hooks_dir(&root).join("pre-push")).expect("restored");
    assert_eq!(pre_push, GO_PRE_PUSH);
    assert!(
        hooks_dir(&root).join(".moai-pre-commit.sha256").exists(),
        "the hash file comes back"
    );
    assert!(
        !hooks_dir(&root).join("commit-msg").exists(),
        "the installed Rust hook is removed on the revert"
    );
    let leftover: Vec<String> = fs::read_dir(hooks_dir(&root))
        .expect("hooks folder")
        .filter_map(|e| e.ok())
        .map(|e| e.file_name().to_string_lossy().to_string())
        .collect();
    assert!(
        leftover.iter().all(|n| !n.starts_with("moai-retired-")),
        "the retire folder is gone: {leftover:?}"
    );
}

// --- hook texts --------------------------------------------------------------

fn run_hook(
    hook_path: &Path,
    root: &Path,
    env_path: &str,
    args: &[&str],
    stdin: Option<&str>,
) -> (i32, String, String) {
    let mut cmd = Command::new(hook_path);
    cmd.args(args).current_dir(root).env("PATH", env_path);
    match stdin {
        Some(text) => {
            cmd.stdin(Stdio::piped())
                .stdout(Stdio::piped())
                .stderr(Stdio::piped());
            let mut child = cmd.spawn().expect("spawn hook");
            use std::io::Write;
            child
                .stdin
                .as_mut()
                .expect("stdin")
                .write_all(text.as_bytes())
                .expect("write stdin");
            let out = child.wait_with_output().expect("wait");
            (
                out.status.code().unwrap_or(-1),
                String::from_utf8_lossy(&out.stderr).to_string(),
                String::from_utf8_lossy(&out.stdout).to_string(),
            )
        }
        None => {
            cmd.stdin(Stdio::null());
            let out = cmd.output().expect("run hook");
            (
                out.status.code().unwrap_or(-1),
                String::from_utf8_lossy(&out.stderr).to_string(),
                String::from_utf8_lossy(&out.stdout).to_string(),
            )
        }
    }
}

fn write_message(root: &Path, subject: &str) -> PathBuf {
    let path = root.join("message.txt");
    fs::write(&path, format!("{subject}\n\nbody line\n")).expect("message file");
    path
}

#[test]
fn hook_text_commit_msg_passes_clean_trailerless_and_merge_messages() {
    let (_dir, root) = make_repo();
    let hook = hooks_dir(&root).join("commit-msg");
    fs::write(&hook, git_hooks::COMMIT_MSG_TEXT).expect("write hook");
    make_executable(&hook);
    for subject in [
        "docs(spec): the inventory refresh",
        "a message without trailers",
        "merge(m-flow): SPEC-20261121-x head 743924d81b",
    ] {
        let (code, stderr, _stdout) = run_hook(
            &hook,
            &root,
            &path_with_moai(),
            &[&write_message(&root, subject).to_string_lossy()],
            None,
        );
        assert_eq!(
            code, 0,
            "clean subject {subject:?} passes; stderr: {stderr}"
        );
    }
}

#[test]
fn hook_text_commit_msg_refuses_a_card_number() {
    let (_dir, root) = make_repo();
    let hook = hooks_dir(&root).join("commit-msg");
    fs::write(&hook, git_hooks::COMMIT_MSG_TEXT).expect("write hook");
    make_executable(&hook);
    let message = write_message(&root, "fix(statusline): apply C-203 feedback");
    let (code, stderr, _stdout) = run_hook(
        &hook,
        &root,
        &path_with_moai(),
        &[&message.to_string_lossy()],
        None,
    );
    assert_eq!(code, 1, "a card number fails the hook; stderr: {stderr}");
}

#[test]
fn hook_text_pre_push_checks_the_pushed_range() {
    let (_dir, root) = make_repo();
    common::git::git_in(
        &root,
        &["commit", "--allow-empty", "-m", "fix: clean change"],
    );
    let clean = common::git::git(&root, &["rev-parse", "HEAD"]);
    common::git::git_in(
        &root,
        &["commit", "--allow-empty", "-m", "fix: apply C-160 note"],
    );
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    let hook = hooks_dir(&root).join("pre-push");
    fs::write(&hook, git_hooks::PRE_PUSH_TEXT).expect("write hook");
    make_executable(&hook);
    let line = format!("refs/heads/main {head} refs/heads/main {clean}\n");
    let (code, stderr, _stdout) = run_hook(&hook, &root, &path_with_moai(), &[], Some(&line));
    assert_eq!(
        code, 1,
        "a card number in the range fails the push; {stderr}"
    );
    // The same range without the offending commit passes.
    let (code, _, _stdout) = run_hook(
        &hook,
        &root,
        &path_with_moai(),
        &[],
        Some(&format!(
            "refs/heads/main {clean} refs/heads/main {clean}\n"
        )),
    );
    assert_eq!(code, 0, "an empty range passes");
}

#[test]
fn hook_text_pre_push_scans_a_new_branch_from_its_root() {
    let (_dir, root) = make_repo();
    let base = common::git::git(&root, &["rev-parse", "HEAD"]);
    common::git::git_in(
        &root,
        &["commit", "--allow-empty", "-m", "fix: apply C-313 note"],
    );
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    let hook = hooks_dir(&root).join("pre-push");
    fs::write(&hook, git_hooks::PRE_PUSH_TEXT).expect("write hook");
    make_executable(&hook);
    let zeros = "0".repeat(40);
    let line = format!("refs/heads/new {head} refs/heads/new {zeros}\n");
    let (code, stderr, _stdout) = run_hook(&hook, &root, &path_with_moai(), &[], Some(&line));
    assert_eq!(code, 1, "a new branch is scanned from its root; {stderr}");
    let _ = base;
}

#[test]
fn hook_text_pre_push_passes_a_delete_and_an_absent_moai() {
    let (_dir, root) = make_repo();
    common::git::git_in(
        &root,
        &["commit", "--allow-empty", "-m", "fix: apply C-999 note"],
    );
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    let hook = hooks_dir(&root).join("pre-push");
    fs::write(&hook, git_hooks::PRE_PUSH_TEXT).expect("write hook");
    make_executable(&hook);
    let zeros = "0".repeat(40);
    let (code, _, _stdout) = run_hook(
        &hook,
        &root,
        &path_with_moai(),
        &[],
        Some(&format!("refs/heads/main {zeros} refs/heads/main {head}\n")),
    );
    assert_eq!(code, 0, "a delete pushes no commits");
    let (code, stderr, _stdout) = run_hook(
        &hook,
        &root,
        &path_without_moai(),
        &[],
        Some(&format!("refs/heads/main {head} refs/heads/main {zeros}\n")),
    );
    assert_eq!(code, 0, "without moai the push passes");
    assert!(
        stderr.contains("moai not on PATH"),
        "the pass-through says why: {stderr}"
    );
}
