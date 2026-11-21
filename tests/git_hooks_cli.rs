//! The git-hook feature of SPEC-20261121-go-path-removal: recognition, the retire-install
//! lifecycle and the hook texts, driven in temporary repositories. The feature module compiles
//! into this test crate by path (the same shape the unwired quality-gate slice used), so it
//! needs no `lib.rs` or `main.rs` line yet; the update layer and the CLI verbs follow the pair
//! requests on the cards that hold those files.
//! Every `git` call goes through `common::git` (the fixture runner), and every hook script
//! runs with a temporary `PATH` and a temporary `HOME` — never the real ones.

#![cfg(unix)]

#[path = "../src/git_hooks/mod.rs"]
// The crate exercises the feature only; the update submodule's items stay unused here, which
// is dead code in this include and live code in the library.
#[allow(dead_code)]
mod git_hooks;

mod common;

use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

use moai_cli::git_hooks::MANAGED;

// The exact Go-generation texts, carried from the hooks the Go build installed (the header
// line and the command line are the recognition proof, so the fixtures carry the real ones).
const GO_PRE_COMMIT: &str = "#!/bin/sh\n\
# MoAI-ADK pre-commit hook — fast subset (gofmt + go vet) + opt-in heavy gate (moai gate)\n\
# Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit\n\
if ! MOAI_PRECOMMIT=1 moai gate; then\n    exit 1\nfi\nexit 0\n";
const GO_PRE_PUSH: &str = "#!/bin/sh\n\
# MoAI-ADK pre-push hook — runs make ci-local; logs invocation outcome\n\
if ! moai hook pre-push; then\n    exit 1\nfi\nexit 0\n";
const USER_COMMIT_MSG: &str = "#!/bin/sh\n# my own commit-msg hook, not MoAI's\necho kept\n";
const HASH_OF_GO: &str = "3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac";

/// The sha256 hex of `bytes`, for hash-file fixtures that must actually match.
fn sha_hex(bytes: &[u8]) -> String {
    use sha2::{Digest, Sha256};
    let mut h = Sha256::new();
    h.update(bytes);
    h.finalize().iter().map(|b| format!("{b:02x}")).collect()
}

fn make_repo() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().expect("temp dir");
    let root = dir.path().join("repo");
    fs::create_dir_all(&root).expect("repo folder");
    common::git::git_ok(&root, &["init"]);
    (dir, root)
}

fn hooks_dir(root: &Path) -> PathBuf {
    root.join(".git").join("hooks")
}

fn write_hook(root: &Path, name: &str, text: &str) -> PathBuf {
    let path = hooks_dir(root).join(name);
    fs::write(&path, text).expect("write hook");
    make_executable(&path);
    path
}

fn write_hash_file(root: &Path, name: &str, hash: &str) {
    fs::write(
        hooks_dir(root).join(format!(".moai-{name}.sha256")),
        format!("{hash}  {name}\n"),
    )
    .expect("write hash file");
}

fn make_executable(path: &Path) {
    fs::set_permissions(path, fs::Permissions::from_mode(0o755)).expect("chmod");
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

// --- recognition (B6: recognition needs positive proof) -----------------------

#[test]
fn recognition_a_matching_hash_file_marks_a_go_generation_hook() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hash_file(&root, "pre-commit", &sha_hex(GO_PRE_COMMIT.as_bytes()));
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(names(&plan, "go_hooks"), vec!["pre-commit".to_string()]);
    assert_eq!(plan["go_hooks"][0]["evidence"], "hash file");
}

#[test]
fn recognition_the_exact_go_header_with_the_gate_line_marks_go_generation() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(names(&plan, "go_hooks"), vec!["pre-commit".to_string()]);
    assert_eq!(plan["go_hooks"][0]["evidence"], "header");
}

#[test]
fn recognition_a_stale_hash_file_keeps_the_edited_hook() {
    let (_dir, root) = make_repo();
    // The user edited the Go hook after it was installed: the hash no longer matches.
    write_hook(
        &root,
        "pre-commit",
        &GO_PRE_COMMIT.replace("exit 0", "echo edited"),
    );
    write_hash_file(&root, "pre-commit", HASH_OF_GO);
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(
        names(&plan, "go_hooks").is_empty(),
        "an edited hook is not retired: {}",
        plan["go_hooks"]
    );
    assert_eq!(names(&plan, "kept"), vec!["pre-commit".to_string()]);
    assert!(
        plan["kept"][0]["reason"]
            .as_str()
            .is_some_and(|r| r.contains("hash")),
        "the reason names the stale hash: {}",
        plan["kept"][0]
    );
}

#[test]
fn recognition_a_product_name_mention_alone_keeps_a_user_hook() {
    let (_dir, root) = make_repo();
    // The name of the product appears in the header, but neither the exact Go header nor the
    // gate command line is there: the hook is the user's.
    write_hook(
        &root,
        "pre-commit",
        "#!/bin/sh\n# runs MoAI checks before every commit\necho moai\n",
    );
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(names(&plan, "go_hooks").is_empty());
    assert_eq!(names(&plan, "kept"), vec!["pre-commit".to_string()]);
}

#[test]
fn recognition_a_go_header_deep_in_the_body_marks_nothing() {
    let (_dir, root) = make_repo();
    // The header proof reads the first lines only: the same text deep in the body of a user
    // hook retires nothing (the mutation that checks the whole file dies here).
    let body = format!("#!/bin/sh\necho start\n{}\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\n# padding line\nexit 0\n", GO_PRE_COMMIT.lines().nth(1).unwrap_or(""));
    write_hook(&root, "pre-commit", &body);
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(names(&plan, "go_hooks").is_empty(), "{}", plan["go_hooks"]);
    assert_eq!(names(&plan, "kept"), vec!["pre-commit".to_string()]);
}

#[test]
fn recognition_an_unreadable_hook_is_kept_never_deleted() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hash_file(&root, "pre-commit", &sha_hex(GO_PRE_COMMIT.as_bytes()));
    // A user hook the process cannot read: mode 000.
    let unreadable = write_hook(&root, "commit-msg", USER_COMMIT_MSG);
    fs::set_permissions(&unreadable, fs::Permissions::from_mode(0o000)).expect("chmod 000");
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(
        names(&plan, "go_hooks"),
        vec!["pre-commit".to_string()],
        "only the readable Go hook retires"
    );
    assert!(
        !names(&plan, "install").contains(&"commit-msg".to_string()),
        "an unreadable name is not installed over: {}",
        plan["install"]
    );
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("apply");
    assert!(
        unreadable.exists(),
        "the unreadable user hook still exists after the apply"
    );
}

#[test]
fn recognition_a_symlink_to_a_folder_is_kept() {
    let (_dir, root) = make_repo();
    let target = root.join("hook-target-folder");
    fs::create_dir_all(&target).expect("folder");
    #[allow(unused_must_use)]
    {
        std::os::unix::fs::symlink(&target, hooks_dir(&root).join("pre-push")).expect("link");
    }
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(names(&plan, "go_hooks").is_empty());
    assert_eq!(names(&plan, "kept"), vec!["pre-push".to_string()]);
    assert!(
        !names(&plan, "install").contains(&"pre-push".to_string()),
        "a symlinked name is not installed over"
    );
}

#[test]
fn recognition_a_drifted_rust_hook_is_kept_and_reported() {
    let (_dir, root) = make_repo();
    let installed = write_hook(&root, "pre-commit", git_hooks::PRE_COMMIT_TEXT);
    // The user edited the installed hook after the apply.
    fs::write(&installed, "#!/bin/sh\n# my edit on top\necho mine\n").expect("edit");
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(names(&plan, "go_hooks").is_empty());
    assert_eq!(names(&plan, "kept"), vec!["pre-commit".to_string()]);
}

#[test]
fn recognition_a_current_rust_hook_is_not_retired_and_not_kept() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", git_hooks::PRE_COMMIT_TEXT);
    let plan = git_hooks::plan(&root).expect("plan");
    assert!(names(&plan, "go_hooks").is_empty());
    assert!(names(&plan, "kept").is_empty());
    assert_eq!(plan["writes"], false);
}

#[test]
fn the_managed_set_is_the_two_git_hooks_without_commit_msg() {
    // The card-number scan of a commit message belongs to the repositories that own the rule;
    // the product ships pre-commit and pre-push only.
    assert_eq!(&MANAGED[..], ["pre-commit", "pre-push"]);
    assert!(
        !Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("install/git-hooks/commit-msg")
            .exists(),
        "the product ships no commit-msg hook"
    );
}

// --- lifecycle ---------------------------------------------------------------

#[test]
fn lifecycle_retires_go_hooks_and_installs_the_rust_hooks() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hook(&root, "pre-push", GO_PRE_PUSH);
    write_hash_file(&root, "pre-commit", &sha_hex(GO_PRE_COMMIT.as_bytes()));
    write_hash_file(&root, "pre-push", &sha_hex(GO_PRE_PUSH.as_bytes()));
    let plan = git_hooks::plan(&root).expect("plan");
    assert_eq!(plan["writes"], true);
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("apply");
    assert_eq!(receipt["state"], "applied");
    assert_eq!(
        names(&receipt, "retired"),
        vec!["pre-commit".to_string(), "pre-push".to_string()]
    );
    for name in ["pre-commit", "pre-push"] {
        let installed = fs::read_to_string(hooks_dir(&root).join(name)).expect("installed");
        let shipped = match name {
            "pre-commit" => git_hooks::PRE_COMMIT_TEXT,
            _ => git_hooks::PRE_PUSH_TEXT,
        };
        assert_eq!(installed, shipped, "{name} carries the shipped bytes");
    }
    assert!(!hooks_dir(&root).join(".moai-pre-commit.sha256").exists());
    assert!(!hooks_dir(&root).join(".moai-pre-push.sha256").exists());
    let stamp = receipt["stamp"].as_str().expect("stamp").to_string();
    let folder = hooks_dir(&root).join(format!("moai-retired-{stamp}"));
    let saved = fs::read_to_string(folder.join("pre-commit")).expect("saved hook");
    assert_eq!(saved, GO_PRE_COMMIT, "the retire folder keeps the bytes");
    assert!(folder.join(".moai-pre-commit.sha256").exists());
}

#[test]
fn lifecycle_keeps_a_user_hook_byte_for_byte() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    let user_msg = write_hook(&root, "commit-msg", USER_COMMIT_MSG);
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("apply");
    // A commit message hook is not managed at all: the scan it would carry belongs to the
    // repositories that own the rule, so the feature neither sees nor installs the name.
    assert!(
        !names(&receipt, "kept").contains(&"commit-msg".to_string()),
        "an unmanaged name is not reported"
    );
    let after = fs::read_to_string(&user_msg).expect("read back");
    assert_eq!(after, USER_COMMIT_MSG, "the user hook keeps its bytes");
    assert!(
        !names(&receipt, "installed").contains(&"commit-msg".to_string()),
        "an occupied name is not installed"
    );
}

#[test]
fn lifecycle_revert_works_when_a_hook_was_kept() {
    // B2: a kept user hook must not stop the revert of the retired hooks.
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hook(&root, "commit-msg", USER_COMMIT_MSG);
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("apply");
    let stamp = receipt["stamp"].as_str().expect("stamp").to_string();
    let reverted = git_hooks::revert(&root, &stamp);
    assert!(
        reverted.is_ok(),
        "the revert works beside a kept hook: {reverted:?}"
    );
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("restored"),
        GO_PRE_COMMIT
    );
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("commit-msg")).expect("kept"),
        USER_COMMIT_MSG,
        "the kept hook stays byte for byte"
    );
}

#[test]
fn lifecycle_a_second_run_changes_nothing() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("first apply");
    let second_plan = git_hooks::plan(&root).expect("second plan");
    assert_eq!(second_plan["writes"], false, "nothing left to do");
    assert!(names(&second_plan, "go_hooks").is_empty());
    assert!(names(&second_plan, "install").is_empty());
    let again = git_hooks::apply(&root, &token, "2026-11-21T22:21:00Z");
    assert!(again.is_err(), "a second apply refuses: {again:?}");
}

#[test]
fn lifecycle_revert_restores_the_go_hooks() {
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    write_hook(&root, "pre-push", GO_PRE_PUSH);
    write_hash_file(&root, "pre-commit", &sha_hex(GO_PRE_COMMIT.as_bytes()));
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("apply");
    let stamp = receipt["stamp"].as_str().expect("stamp").to_string();
    git_hooks::revert(&root, &stamp).expect("revert");
    let pre_commit = fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("restored");
    assert_eq!(
        pre_commit, GO_PRE_COMMIT,
        "the Go hook comes back byte for byte"
    );
    let pre_push = fs::read_to_string(hooks_dir(&root).join("pre-push")).expect("restored");
    assert_eq!(pre_push, GO_PRE_PUSH);
    assert!(hooks_dir(&root).join(".moai-pre-commit.sha256").exists());
    assert!(
        !hooks_dir(&root)
            .join("moai-retired-20261121T222000Z")
            .exists(),
        "the retire folder of the apply is gone"
    );
}

#[test]
fn lifecycle_revert_keeps_a_user_edited_installed_hook() {
    // The mutation that lets the revert overwrite a user-edited hook dies here: the revert
    // succeeds around the edit, and the edit survives byte for byte, reported as kept.
    let (_dir, root) = make_repo();
    write_hook(&root, "pre-commit", GO_PRE_COMMIT);
    let plan = git_hooks::plan(&root).expect("plan");
    let token = plan["confirm_token"].as_str().expect("token").to_string();
    let receipt = git_hooks::apply(&root, &token, "2026-11-21T22:20:00Z").expect("apply");
    let stamp = receipt["stamp"].as_str().expect("stamp").to_string();
    let edited = hooks_dir(&root).join("pre-push");
    fs::write(&edited, "#!/bin/sh\n# the user's own edit\necho mine\n").expect("edit");
    let reverted = git_hooks::revert(&root, &stamp).expect("the revert works around the edit");
    assert_eq!(
        fs::read_to_string(&edited).expect("the edit survives"),
        "#!/bin/sh\n# the user's own edit\necho mine\n"
    );
    let kept: Vec<String> = reverted["kept"]
        .as_array()
        .map(|rows| {
            rows.iter()
                .map(|r| r["name"].as_str().unwrap_or_default().to_string())
                .collect()
        })
        .unwrap_or_default();
    assert_eq!(
        kept,
        vec!["pre-push".to_string()],
        "the edit is reported kept"
    );
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("restored"),
        GO_PRE_COMMIT,
        "the retired hook still comes back"
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
    cmd.args(args)
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("PATH", env_path);
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

#[test]
fn hook_text_pre_push_checks_the_pushed_range() {
    let (_dir, root) = make_repo();
    common::git::git_in(
        &root,
        &["commit", "--allow-empty", "-m", "fix: clean change"],
    );
    let clean = common::git::git(&root, &["rev-parse", "HEAD"]);
    // The card-number shape, built from pieces so the source carries no literal (the scan
    // reads the commit the fixture makes, not this file).
    let shaped = format!("fix: apply {}-{} note", "C", 123);
    common::git::git_in(&root, &["commit", "--allow-empty", "-m", &shaped]);
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    let hook = write_hook(&root, "pre-push", git_hooks::PRE_PUSH_TEXT);
    let line = format!("refs/heads/main {head} refs/heads/main {clean}\n");
    let (code, stderr, _stdout) = run_hook(
        &hook,
        &root,
        &format!(
            "{}:{}",
            PathBuf::from(common::MOAI_BIN)
                .parent()
                .expect("bin folder")
                .display(),
            std::env::var("PATH").unwrap_or_default()
        ),
        &[],
        Some(&line),
    );
    assert_eq!(code, 1, "a failing range fails the push; {stderr}");
    let (code, _, _stdout) = run_hook(
        &hook,
        &root,
        &format!(
            "{}:{}",
            PathBuf::from(common::MOAI_BIN)
                .parent()
                .expect("bin folder")
                .display(),
            std::env::var("PATH").unwrap_or_default()
        ),
        &[],
        Some(&format!(
            "refs/heads/main {clean} refs/heads/main {clean}\n"
        )),
    );
    assert_eq!(code, 0, "an empty range passes");
}

#[test]
fn hook_text_pre_push_passes_a_new_branch_without_scanning_history() {
    // A new remote branch has no pushed range: the hook passes with a note instead of
    // scanning the whole history (B5).
    let (_dir, root) = make_repo();
    common::git::git_in(
        &root,
        &[
            "commit",
            "--allow-empty",
            "-m",
            "fix: a subject in letter-number shape",
        ],
    );
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    let hook = write_hook(&root, "pre-push", git_hooks::PRE_PUSH_TEXT);
    let zeros = "0".repeat(40);
    let line = format!("refs/heads/new {head} refs/heads/new {zeros}\n");
    let (code, stderr, _stdout) = run_hook(
        &hook,
        &root,
        &format!(
            "{}:{}",
            PathBuf::from(common::MOAI_BIN)
                .parent()
                .expect("bin folder")
                .display(),
            std::env::var("PATH").unwrap_or_default()
        ),
        &[],
        Some(&line),
    );
    assert_eq!(code, 0, "a new branch passes: {stderr}");
    assert!(
        stderr.contains("no remote range"),
        "the pass-through says why: {stderr}"
    );
}

#[test]
fn hook_text_pre_push_passes_a_delete_and_an_absent_moai() {
    let (_dir, root) = make_repo();
    let hook = write_hook(&root, "pre-push", git_hooks::PRE_PUSH_TEXT);
    let zeros = "0".repeat(40);
    let some = "1".repeat(40);
    let (code, _, _stdout) = run_hook(
        &hook,
        &root,
        &format!(
            "{}:{}",
            PathBuf::from(common::MOAI_BIN)
                .parent()
                .expect("bin folder")
                .display(),
            std::env::var("PATH").unwrap_or_default()
        ),
        &[],
        Some(&format!("refs/heads/main {zeros} refs/heads/main {some}\n")),
    );
    assert_eq!(code, 0, "a delete pushes no commits");
    let (code, stderr, _stdout) = run_hook(
        &hook,
        &root,
        "/usr/bin:/bin",
        &[],
        Some(&format!("refs/heads/main {some} refs/heads/main {zeros}\n")),
    );
    assert_eq!(code, 0, "without moai the push passes");
    assert!(
        stderr.contains("moai not on PATH"),
        "the pass-through says why: {stderr}"
    );
}
