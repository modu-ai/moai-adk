//! The git-hook verbs (`git-hooks-plan|apply|revert`) and the update layer of
//! SPEC-20261121-go-path-removal: temporary repositories and HOME folders only, never the real
//! ones. The verbs run the built binary; the layer runs at library level, the way
//! `tests/update_switch_layer.rs` drives the switch layer; the `moai update` runs use a staged
//! temporary HOME.

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
// The real Go-generation text (header line and gate command line), as the Go build wrote it.
const GO_PRE_COMMIT: &str = "#!/bin/sh\n\
# MoAI-ADK pre-commit hook — fast subset (gofmt + go vet) + opt-in heavy gate (moai gate)\n\
# Bypass via: SKIP_MOAI_PRECOMMIT=1 git commit\n\
if ! MOAI_PRECOMMIT=1 moai gate; then\n    exit 1\nfi\nexit 0\n";

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
fn stage_home(home: &Path) -> PathBuf {
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).expect("versions folder");
    common::copy_executable(Path::new(common::MOAI_BIN), &folder.join("bin/moai-cli"));
    let manifest = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&manifest.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&manifest.join("agents"), &folder.join("agents"));
    folder.join("bin/moai-cli")
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
fn update_applies_the_hook_step_with_the_token_alone() {
    // A run with only the git-hook token (no digest) applies the hook step and nothing else.
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
    let token = preview["result"]["git_hooks"]["confirm_token"]
        .as_str()
        .expect("token")
        .to_string();
    let applied = run(&["update", "--json", "--confirm-git-hooks", &token]);
    assert_eq!(
        applied["result"]["git_hooks_result"]["state"], "applied",
        "the token-only run applied the hook step: {applied}"
    );
    assert_eq!(
        fs::read_to_string(hooks_dir(&root).join("pre-commit")).expect("installed"),
        moai_cli::git_hooks::PRE_COMMIT_TEXT,
        "the hook was really replaced"
    );
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

// --- the guard hook and the stale record (SPEC-20261126-guard-record-repoint) ---------------

const GUARD_HOOK: &str = "reference-transaction";
const CURRENT_RECORD: &str = "refs/heads/NeXT\n";
const STALE_RECORD: &str = "refs/heads/main\n";

/// The profile of the project, naming `branch` as the integration branch (or none), the way
/// `tests/integration_guard_install.rs` writes it.
fn write_guard_profile(root: &Path, branch: Option<&str>) {
    let flow = match branch {
        Some(b) => {
            format!(", \"work_flow\": {{\"flow\": \"next-flow\", \"integration_branch\": \"{b}\"}}")
        }
        None => String::new(),
    };
    let cfg = moai_cli::state_root::path(root, "project.yaml");
    fs::create_dir_all(cfg.parent().expect("state folder")).expect("state folder");
    fs::write(
        &cfg,
        format!(
            "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"]{flow}}}"
        ),
    )
    .expect("profile");
}

/// The guard record, in the common git folder of the repository at `root`.
fn guard_record(root: &Path) -> PathBuf {
    root.join(".git").join(moai_cli::card::guard::RECORD_FILE)
}

fn write_guard_record(root: &Path, text: &str) {
    let path = guard_record(root);
    fs::create_dir_all(path.parent().expect("record folder")).expect("record folder");
    fs::write(path, text).expect("record");
}

/// A staged temporary HOME and the `moai update` of its package, run in a project folder.
struct GuardUpdate {
    _tmp: TempDir,
    exe: PathBuf,
    home: PathBuf,
}

impl GuardUpdate {
    fn new() -> Self {
        let tmp = TempDir::new().expect("home");
        let home = tmp.path().join("home");
        fs::create_dir_all(&home).expect("home folder");
        let exe = stage_home(&home);
        GuardUpdate {
            _tmp: tmp,
            exe,
            home,
        }
    }

    /// `moai <args>` in `cwd`; answers the envelope.
    fn run(&self, cwd: &Path, args: &[&str]) -> Value {
        let out = Command::new(&self.exe)
            .args(args)
            .current_dir(cwd)
            .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
            .env("HOME", &self.home)
            .env("LC_ALL", "C")
            .stdin(Stdio::null())
            .output()
            .expect("run update");
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "no envelope: {} / {}",
                String::from_utf8_lossy(&out.stdout),
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn preview(&self, cwd: &Path) -> Value {
        let v = self.run(cwd, &["update", "--json"]);
        assert_eq!(v["ok"], true, "the preview succeeds: {v}");
        v
    }
}

/// The names the git-hook part of an update preview installs.
fn installs(preview: &Value) -> Vec<String> {
    preview["result"]["git_hooks"]["install"]
        .as_array()
        .map(|rows| {
            rows.iter()
                .filter_map(|r| r.as_str().or_else(|| r["name"].as_str()).map(String::from))
                .collect()
        })
        .unwrap_or_default()
}

/// The notices of an update preview that point at `moai card guard-repoint`.
fn repoint_notices(preview: &Value) -> Vec<Value> {
    preview["result"]["notices"]
        .as_array()
        .unwrap_or_else(|| panic!("the preview carries a notices list: {preview}"))
        .iter()
        .filter(|n| n.to_string().contains("guard-repoint"))
        .cloned()
        .collect()
}

fn token_of(preview: &Value) -> String {
    preview["result"]["git_hooks"]["confirm_token"]
        .as_str()
        .unwrap_or_else(|| panic!("the preview offers no git-hook token: {preview}"))
        .to_string()
}

fn assert_hook_installed(root: &Path) {
    let path = hooks_dir(root).join(GUARD_HOOK);
    assert_eq!(
        fs::read_to_string(&path).expect("the guard hook is installed"),
        moai_cli::card::guard::hook_text(),
        "the shipped bytes"
    );
    assert_eq!(
        fs::metadata(&path).expect("hook").permissions().mode() & 0o777,
        0o755,
        "the hook is executable"
    );
}

/// AC-4: a repository that holds a current record and no hook is offered the guard hook, the
/// token alone installs it, and the answer carries `guard`; the plan and the apply were given the
/// same guard answer, so the preview's token is the one the apply accepts.
#[test]
fn update_offers_the_guard_hook_with_the_token_alone() {
    let (_dir, root) = make_repo();
    write_guard_profile(&root, Some("NeXT"));
    write_guard_record(&root, CURRENT_RECORD);
    let env = GuardUpdate::new();
    let preview = env.preview(&root);
    assert!(
        preview["result"]["git_hooks"].is_object(),
        "the preview carries a git_hooks part: {preview}"
    );
    assert_eq!(installs(&preview), [GUARD_HOOK], "{preview}");
    let token = token_of(&preview);
    assert_eq!(
        token_of(&env.preview(&root)),
        token,
        "the same repository offers the same token"
    );
    let applied = env.run(&root, &["update", "--json", "--confirm-git-hooks", &token]);
    let result = &applied["result"]["git_hooks_result"];
    assert_eq!(result["state"], "applied", "{applied}");
    assert_hook_installed(&root);
    assert!(
        result["guard"].is_object(),
        "the answer carries guard: {applied}"
    );
    assert_eq!(result["guard"]["record"], "present", "{applied}");
    assert_eq!(
        fs::read_to_string(guard_record(&root)).expect("record"),
        CURRENT_RECORD,
        "the record is not rewritten"
    );
}

/// AC-4: the digest-confirmed run applies the guard hook too, with the same guard answer as the
/// preview (the second apply site of `moai update`).
#[test]
fn update_offers_the_guard_hook_with_both_confirm_flags() {
    let (_dir, root) = make_repo();
    write_guard_profile(&root, Some("NeXT"));
    write_guard_record(&root, CURRENT_RECORD);
    let env = GuardUpdate::new();
    let preview = env.preview(&root);
    let digest = preview["result"]["plan_digest"]
        .as_str()
        .expect("digest")
        .to_string();
    let token = token_of(&preview);
    let applied = env.run(
        &root,
        &[
            "update",
            "--json",
            "--confirm",
            &digest,
            "--confirm-git-hooks",
            &token,
        ],
    );
    assert_eq!(applied["ok"], true, "the confirmed run succeeds: {applied}");
    let result = &applied["result"]["git_hooks_result"];
    assert_eq!(result["state"], "applied", "{applied}");
    assert_hook_installed(&root);
    assert_eq!(result["guard"]["record"], "present", "{applied}");
}

/// AC-4: a profile that names the integration branch is enough: with no record the guard hook is
/// offered, and the apply seeds the record from the profile (`written`, said in the answer).
#[test]
fn update_offers_the_guard_hook_from_the_profile_alone_and_seeds_the_record() {
    let (_dir, root) = make_repo();
    write_guard_profile(&root, Some("NeXT"));
    let env = GuardUpdate::new();
    let preview = env.preview(&root);
    assert_eq!(installs(&preview), [GUARD_HOOK], "{preview}");
    let token = token_of(&preview);
    let applied = env.run(&root, &["update", "--json", "--confirm-git-hooks", &token]);
    let result = &applied["result"]["git_hooks_result"];
    assert_eq!(result["state"], "applied", "{applied}");
    assert_hook_installed(&root);
    assert_eq!(result["guard"]["record"], "written", "{applied}");
    assert_eq!(
        fs::read_to_string(guard_record(&root)).expect("record"),
        CURRENT_RECORD
    );
}

/// AC-4: beside a Go-generation hook that retires, a current record adds the guard hook to the
/// two other free names.
#[test]
fn update_offers_the_guard_hook_beside_a_retired_go_hook() {
    let (_dir, root) = make_repo();
    put_go_hook(&root);
    write_guard_profile(&root, Some("NeXT"));
    write_guard_record(&root, CURRENT_RECORD);
    let env = GuardUpdate::new();
    let preview = env.preview(&root);
    assert_eq!(
        installs(&preview),
        ["pre-commit", "pre-push", GUARD_HOOK],
        "{preview}"
    );
    let token = token_of(&preview);
    let applied = env.run(&root, &["update", "--json", "--confirm-git-hooks", &token]);
    assert_eq!(
        applied["result"]["git_hooks_result"]["state"], "applied",
        "{applied}"
    );
    assert_hook_installed(&root);
    assert!(hooks_dir(&root).join("pre-commit").is_file());
    assert!(hooks_dir(&root).join("pre-push").is_file());
}

/// AC-4, the direction that stays closed: a stale record offers no guard hook. With nothing else
/// to write the preview has no `git_hooks` part; beside a retiring Go hook the part offers the
/// other two names and not the guard hook, and applying it leaves the guard hook out and the
/// record as it was.
#[test]
fn update_offers_the_guard_hook_not_for_a_stale_record() {
    let (_dir, root) = make_repo();
    write_guard_profile(&root, Some("NeXT"));
    write_guard_record(&root, STALE_RECORD);
    let env = GuardUpdate::new();
    let preview = env.preview(&root);
    assert!(
        preview["result"]["git_hooks"].is_null(),
        "a stale record offers nothing to install: {preview}"
    );
    let stranger = env.run(
        &root,
        &[
            "update",
            "--json",
            "--confirm-git-hooks",
            "git-hooks:stranger",
        ],
    );
    assert_eq!(stranger["ok"], false, "no token is offered: {stranger}");
    assert!(!hooks_dir(&root).join(GUARD_HOOK).exists());
    put_go_hook(&root);
    let preview = env.preview(&root);
    assert_eq!(installs(&preview), ["pre-commit", "pre-push"], "{preview}");
    let token = token_of(&preview);
    let applied = env.run(&root, &["update", "--json", "--confirm-git-hooks", &token]);
    let result = &applied["result"]["git_hooks_result"];
    assert_eq!(result["state"], "applied", "{applied}");
    assert!(
        !hooks_dir(&root).join(GUARD_HOOK).exists(),
        "the guard hook is withheld while the record is stale"
    );
    assert_eq!(result["guard"]["record"], "present", "{applied}");
    assert_eq!(
        fs::read_to_string(guard_record(&root)).expect("record"),
        STALE_RECORD,
        "the update never rewrites the record"
    );
}

/// AC-4, the other closed direction: no record and no integration branch in the profile offers
/// no guard hook, with or without a Go hook that retires.
#[test]
fn update_offers_the_guard_hook_not_without_a_record_or_a_profile_branch() {
    let (_dir, root) = make_repo();
    write_guard_profile(&root, None);
    let env = GuardUpdate::new();
    let preview = env.preview(&root);
    assert!(
        preview["result"]["git_hooks"].is_null(),
        "nothing to install: {preview}"
    );
    put_go_hook(&root);
    let preview = env.preview(&root);
    assert_eq!(installs(&preview), ["pre-commit", "pre-push"], "{preview}");
    assert!(!guard_record(&root).exists());
}

/// AC-5: the preview of a stale primary worktree carries one read-only notice that names the
/// record, the profile branch and `moai card guard-repoint`, has no `git_hooks` part when the
/// hook is current, and keeps the notice outside `plan_digest`; an unreadable profile and a
/// linked worktree add none.
#[test]
fn update_notice_names_a_stale_record() {
    let (dir, root) = make_repo();
    write_guard_profile(&root, Some("NeXT"));
    write_guard_record(&root, CURRENT_RECORD);
    // One commit first, before any hook exists, so the linked worktree below has a HEAD.
    common::git::git_ok(
        &root,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "commit",
            "--allow-empty",
            "-q",
            "-m",
            "init",
        ],
    );
    // The hook is current: installed through the verbs, the way a person gets it.
    let verbs_home = verb_home();
    let (code, plan) = run_verb(
        &root,
        verbs_home.path(),
        &["git-hooks-plan", root.to_str().unwrap()],
    );
    assert_eq!(code, 0, "git-hooks-plan succeeds");
    let token = plan.expect("plan envelope")["result"]["confirm_token"]
        .as_str()
        .expect("token")
        .to_string();
    let (code, applied) = run_verb(
        &root,
        verbs_home.path(),
        &[
            "git-hooks-apply",
            root.to_str().unwrap(),
            "--confirm-git-hooks",
            &token,
        ],
    );
    assert_eq!(code, 0, "git-hooks-apply succeeds: {applied:?}");
    assert_hook_installed(&root);
    let env = GuardUpdate::new();
    // A current record: no notice and no git_hooks part, and the digest to compare with.
    let current = env.preview(&root);
    assert!(repoint_notices(&current).is_empty(), "{current}");
    assert!(current["result"]["git_hooks"].is_null(), "{current}");
    let current_digest = current["result"]["plan_digest"]
        .as_str()
        .expect("digest")
        .to_string();
    // The record goes stale: the profile now names another branch than the record.
    write_guard_record(&root, STALE_RECORD);
    let stale = env.preview(&root);
    let notices = repoint_notices(&stale);
    assert_eq!(notices.len(), 1, "exactly one notice: {stale}");
    let reason = notices[0]["reason"].as_str().unwrap_or_else(|| {
        panic!("the notice carries a reason text: {}", notices[0]);
    });
    for needle in ["main", "NeXT", "moai card guard-repoint"] {
        assert!(reason.contains(needle), "{needle} is not in `{reason}`");
    }
    assert!(
        stale["result"]["git_hooks"].is_null(),
        "a current hook leaves nothing for the git-hook layer: {stale}"
    );
    assert_eq!(
        stale["result"]["plan_digest"].as_str(),
        Some(current_digest.as_str()),
        "the notice is outside plan_digest"
    );
    assert_eq!(
        fs::read_to_string(guard_record(&root)).expect("record"),
        STALE_RECORD,
        "the preview writes nothing"
    );
    // An unreadable profile adds no notice.
    let cfg = moai_cli::state_root::path(&root, "project.yaml");
    fs::write(&cfg, "{ this is not a profile").expect("broken profile");
    let broken = env.preview(&root);
    assert!(repoint_notices(&broken).is_empty(), "{broken}");
    write_guard_profile(&root, Some("NeXT"));
    // A linked worktree adds no notice.
    let linked = dir.path().join("linked");
    common::git::git_ok(
        &root,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            linked.to_str().unwrap(),
        ],
    );
    write_guard_profile(&linked, Some("NeXT"));
    let from_linked = env.preview(&linked);
    assert!(repoint_notices(&from_linked).is_empty(), "{from_linked}");
}
