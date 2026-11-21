//! The git-hook verbs (`git-hooks-plan|apply|revert`) and the update layer of
//! SPEC-20261121-go-path-removal: temporary repositories and HOME folders only, never the real
//! ones. The verbs run the built binary; the layer runs at library level, the way
//! `tests/update_switch_layer.rs` drives the switch layer. The `moai update` JSON-part test
//! stays ignored until the wiring lines of `src/update.rs` land (pair with the card that
//! holds that file, answer to note 5820).

#![cfg(unix)]

mod common;

use moai_cli::git_hooks::update;
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use tempfile::TempDir;

const NOW: &str = "2026-11-21T20:10:00Z";
const GO_PRE_COMMIT: &str = "#!/bin/sh\n# MoAI-ADK pre-commit hook - fast subset\nexit 0\n";

/// The sha256 hex of `bytes`, for hash-file fixtures that must actually match the bytes.
fn sha_hex(bytes: &[u8]) -> String {
    use sha2::{Digest, Sha256};
    let mut h = Sha256::new();
    h.update(bytes);
    h.finalize().iter().map(|b| format!("{b:02x}")).collect()
}

fn make_repo() -> (TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().expect("temp dir");
    let root = dir.path().join("repo");
    fs::create_dir_all(&root).expect("repo folder");
    common::git::git_ok(&root, &["init"]);
    (dir, root)
}

fn hooks_dir(root: &Path) -> PathBuf {
    root.join(".git").join("hooks")
}

fn put_go_hook(root: &Path) {
    let path = hooks_dir(root).join("pre-commit");
    fs::write(&path, GO_PRE_COMMIT).expect("write hook");
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).expect("chmod");
    fs::write(
        hooks_dir(root).join(".moai-pre-commit.sha256"),
        format!("{}  pre-commit\n", sha_hex(GO_PRE_COMMIT.as_bytes())),
    )
    .expect("write hash file");
}

/// Runs one verb of the built binary in the repository; answers (exit code, envelope).
fn run_verb(root: &Path, home: &Path, args: &[&str]) -> (i32, Option<Value>) {
    let out = Command::new(common::MOAI_BIN)
        .args(args)
        .current_dir(root)
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("run verb");
    let code = out.status.code().unwrap_or(-1);
    let parsed: Option<Value> = serde_json::from_slice(&out.stdout).ok();
    (code, parsed)
}

fn verb_home() -> TempDir {
    TempDir::new().expect("verb home")
}

// --- the verbs ---------------------------------------------------------------

#[test]
fn verbs_plan_previews_and_apply_needs_the_token() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let home = verb_home();
    let (code, plan) = run_verb(
        &root,
        home.path(),
        &["git-hooks-plan", root.to_str().unwrap()],
    );
    assert_eq!(code, 0, "git-hooks-plan succeeds");
    let plan = plan.expect("envelope");
    assert_eq!(plan["operation"], "git_hooks.plan");
    assert_eq!(plan["result"]["writes"], true);
    assert_eq!(plan["result"]["go_hooks"][0]["name"], "pre-commit");
    let token = plan["result"]["confirm_token"]
        .as_str()
        .unwrap()
        .to_string();

    let (_code, wrong) = run_verb(
        &root,
        home.path(),
        &[
            "git-hooks-apply",
            root.to_str().unwrap(),
            "--confirm-git-hooks",
            "git-hooks:not-the-token",
        ],
    );
    let wrong = wrong.expect("refusal envelope");
    assert_eq!(wrong["ok"], false, "a stranger token is refused: {wrong}");

    let (code, applied) = run_verb(
        &root,
        home.path(),
        &[
            "git-hooks-apply",
            root.to_str().unwrap(),
            "--confirm-git-hooks",
            &token,
        ],
    );
    assert_eq!(code, 0, "the right token applies");
    let applied = applied.expect("applied envelope");
    assert_eq!(applied["result"]["state"], "applied");
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("installed"),
        moai_cli::git_hooks::PRE_COMMIT_TEXT,
        "the shipped text is installed"
    );
    assert!(
        hooks_dir(&root).join("pre-push").exists(),
        "the free managed name is installed"
    );
    assert!(
        !hooks_dir(&root).join("commit-msg").exists(),
        "the product ships no commit-msg hook"
    );
}

#[test]
fn verbs_revert_restores_the_retired_hooks() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let home = verb_home();
    let (_code, plan) = run_verb(
        &root,
        home.path(),
        &["git-hooks-plan", root.to_str().unwrap()],
    );
    let token = plan.expect("envelope")["result"]["confirm_token"]
        .as_str()
        .unwrap()
        .to_string();
    let (_code, applied) = run_verb(
        &root,
        home.path(),
        &[
            "git-hooks-apply",
            root.to_str().unwrap(),
            "--confirm-git-hooks",
            &token,
        ],
    );
    let stamp = applied.expect("applied")["result"]["stamp"]
        .as_str()
        .unwrap()
        .to_string();

    let (code, reverted) = run_verb(
        &root,
        home.path(),
        &["git-hooks-revert", &stamp, root.to_str().unwrap()],
    );
    assert_eq!(code, 0, "the revert succeeds");
    assert_eq!(reverted.expect("reverted")["result"]["state"], "reverted");
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("restored"),
        GO_PRE_COMMIT,
        "the Go hook comes back byte for byte"
    );
    assert!(
        hooks_dir(&root).join(".moai-pre-commit.sha256").exists(),
        "the hash file comes back"
    );
    assert!(
        !hooks_dir(&root).join("pre-push").exists(),
        "the installed hook goes away"
    );
}

// --- the update layer (library level) ----------------------------------------

fn layer_plan(root: &Path) -> update::Layer {
    update::plan(Path::new(""), root).expect("layer plan")
}

#[test]
fn layer_plan_carries_the_token_summary_and_parts() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let layer = layer_plan(&root);
    assert!(layer.writes());
    let token = layer.token().expect("token").to_string();
    assert!(
        token.starts_with("git-hooks:"),
        "the token is the layer's: {token}"
    );
    let summary = layer.summary();
    assert!(summary.contains("retire pre-commit"), "summary: {summary}");
    assert!(
        summary.contains("keep"),
        "summary names the kept: {summary}"
    );
    let parts = layer.parts();
    assert_eq!(parts["go_hooks"][0]["name"], "pre-commit");
}

#[test]
fn layer_a_repository_with_nothing_to_change_carries_no_part() {
    let (_dir, root) = make_repo();
    let layer = layer_plan(&root);
    assert!(!layer.writes());
    assert!(layer.to_json().is_none());
    assert_eq!(layer.summary(), "");
}

#[test]
fn layer_check_token_refuses_a_stranger() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let layer = layer_plan(&root);
    let answer = update::check_token(&layer, Some("git-hooks:stranger"));
    assert!(answer.is_err(), "a stranger token is refused");
    assert_eq!(update::check_token(&layer, None).ok(), Some(None));
}

#[test]
fn layer_apply_and_revert_round_trip() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let layer = layer_plan(&root);
    let token = layer.token().expect("token").to_string();
    assert!(update::check_token(&layer, Some(&token)).is_ok());
    let receipt = update::apply(&root, &token, NOW).expect("apply");
    assert_eq!(receipt["state"], "applied");
    let stamp = receipt["stamp"].as_str().unwrap().to_string();
    update::revert(&root, &stamp).expect("revert");
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("restored"),
        GO_PRE_COMMIT
    );
}

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).expect("copy target");
    for entry in fs::read_dir(from).expect("copy source") {
        let p = entry.expect("copy entry").path();
        let dest = to.join(p.file_name().expect("named"));
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).expect("copy file");
        }
    }
}

/// Stages the installed package into a temporary HOME the way `tests/update_switch_cli.rs`
/// does, so `moai update` has a version folder to plan from.
fn stage_home(home: &Path) {
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).expect("versions folder");
    common::copy_executable(Path::new(common::MOAI_BIN), &folder.join("bin/moai-cli"));
    let manifest = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&manifest.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&manifest.join("agents"), &folder.join("agents"));
}

#[test]
fn update_json_part_carries_git_hooks() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let tmp = TempDir::new().expect("home");
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).expect("home folder");
    stage_home(&home);
    let exe = home.join(".moai-cli/versions/a/bin/moai-cli");
    let out = Command::new(&exe)
        .args(["update", "--json"])
        .current_dir(&root)
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", &home)
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("run update");
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("envelope");
    let part = &envelope["result"]["git_hooks"];
    assert!(
        part.is_object(),
        "the preview carries the git_hooks part: {}",
        envelope
    );
    assert_eq!(part["go_hooks"][0]["name"], "pre-commit");
    assert!(part["confirm_token"].as_str().is_some());
}

#[test]
fn update_applies_the_hook_step_with_both_confirm_flags() {
    // The digest-confirmed run applies the token parts too: the hook step must not be dropped
    // when --confirm and --confirm-git-hooks arrive together.
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    let tmp = TempDir::new().expect("home");
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).expect("home folder");
    stage_home(&home);
    let exe = home.join(".moai-cli/versions/a/bin/moai-cli");
    let run = |args: &[&str]| -> Value {
        let out = Command::new(&exe)
            .args(args)
            .current_dir(&root)
            .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
            .env("HOME", &home)
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .output()
            .expect("run update");
        serde_json::from_slice(&out.stdout).expect("envelope")
    };
    let preview = run(&["update", "--json"]);
    let digest = preview["result"]["plan_digest"]
        .as_str()
        .expect("digest")
        .to_string();
    let token = preview["result"]["git_hooks"]["confirm_token"]
        .as_str()
        .expect("token")
        .to_string();
    let applied = run(&[
        "update",
        "--json",
        "--confirm",
        &digest,
        "--confirm-git-hooks",
        &token,
    ]);
    assert_eq!(applied["ok"], true, "the confirmed run succeeds: {applied}");
    let result = &applied["result"];
    assert_eq!(
        result["git_hooks_result"]["state"], "applied",
        "the hook step applied: {}",
        result["git_hooks_result"]
    );
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("installed"),
        moai_cli::git_hooks::PRE_COMMIT_TEXT,
        "the hook step really replaced the hook"
    );
    let stamp = result["git_hooks_result"]["stamp"]
        .as_str()
        .expect("stamp")
        .to_string();
    assert_eq!(
        result["revert"]["git_hooks"].as_str(),
        Some(format!("moai git-hooks-revert {stamp}").as_str()),
    );
}
