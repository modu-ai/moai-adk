//! SPEC-20261121-update-installs-hooks: `moai update` offers the Claude Code and Codex hook sets
//! the same way where they are not installed yet: planned as `create` in the preview, written on
//! the one confirmation with one revert receipt, and removed again by `hooks-revert`.
//! Everything runs in a temporary HOME and a temporary git project.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test-approval";

fn copy_tree(from: &Path, to: &Path) {
    fs::create_dir_all(to).unwrap();
    for entry in fs::read_dir(from).unwrap() {
        let p = entry.unwrap().path();
        let dest = to.join(p.file_name().unwrap());
        if p.is_dir() {
            copy_tree(&p, &dest);
        } else {
            fs::copy(&p, &dest).unwrap();
        }
    }
}

struct Fixture {
    _tmp: TempDir,
    home: PathBuf,
    exe: PathBuf,
    project: PathBuf,
    path: String,
}

/// A temporary HOME with the binary, the real package and the agent sources in a version folder,
/// and a git project beside it with no hook set and no profile.
fn setup() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&repo.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&repo.join("agents"), &folder.join("agents"));
    let project = root.join("project");
    fs::create_dir_all(&project).unwrap();
    common::git::git_ok(&project, &["init", "-q"]);
    let bin = root.join("bin");
    fs::create_dir_all(&bin).unwrap();
    std::os::unix::fs::symlink(&exe, bin.join("moai")).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        exe,
        project,
        path: format!("{}:/usr/bin:/bin", bin.display()),
    }
}

fn command(f: &Fixture, args: &[&str]) -> Command {
    let mut c = Command::new(&f.exe);
    c.args(args)
        .clean_env()
        .env("PATH", &f.path)
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(&f.project);
    c
}

fn run(f: &Fixture, args: &[&str]) -> Output {
    command(f, args).stdin(Stdio::null()).output().unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("not json: {}", String::from_utf8_lossy(&out.stdout)))
}

fn result(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn call_with_stdin(f: &Fixture, args: &[&str], input: &[u8]) -> Value {
    let mut child = command(f, args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input);
    result(&child.wait_with_output().unwrap())
}

/// Writes a profile that lists `hosts`, the way PROJECT writes it.
fn profile_with_hosts(f: &Fixture, hosts: &[&str]) {
    let p = f.project.to_str().unwrap();
    let detect = result(&run(f, &["project-detect", p]));
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": hosts,
        "detected": detect["detected"],
    });
    let plan = call_with_stdin(
        f,
        &["project-plan", p],
        &serde_json::to_vec(&profile).unwrap(),
    );
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    call_with_stdin(
        f,
        &[
            "project-apply",
            p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            REF,
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
}

fn update_preview(f: &Fixture) -> Value {
    result(&run(f, &["update"]))
}

fn confirm_with(f: &Fixture, digest: &str) -> Output {
    run(f, &["update", "--confirm", digest])
}

fn update_confirm(f: &Fixture) -> Value {
    let digest = update_preview(f)["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    result(&confirm_with(f, &digest))
}

fn host_names(preview: &Value) -> Vec<String> {
    preview["hooks"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap().to_string())
        .collect()
}

fn hook_host<'a>(preview: &'a Value, host: &str) -> &'a Value {
    preview["hooks"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(host))
        .unwrap_or_else(|| panic!("no {host} in {}", preview["hooks"]))
}

fn claude_file(f: &Fixture) -> PathBuf {
    f.project.join(".claude/settings.json")
}

fn codex_file(f: &Fixture) -> PathBuf {
    f.project.join(".codex/hooks.json")
}

/// The one receipt folder the confirmed update left under the project's state folder.
fn receipt_dirs(f: &Fixture) -> Vec<PathBuf> {
    let mut out = Vec::new();
    for state in [".moai", ".moai-cli"] {
        let dir = f.project.join(state).join("receipts");
        if let Ok(entries) = fs::read_dir(&dir) {
            for e in entries.flatten() {
                let p = e.path();
                let named = p
                    .file_name()
                    .and_then(|n| n.to_str())
                    .is_some_and(|n| n.starts_with("hooks-"));
                if p.is_dir() && named {
                    out.push(p);
                }
            }
        }
    }
    out
}

const USER_HOOK: &str = r#"{"hooks":{"PreToolUse":[{"matcher":"*","hooks":[{"type":"command","command":"/usr/local/bin/other-tool"}]}]}}"#;

/// AC-1 (REQ-1, REQ-3): with no MoAI hooks and no profile, the preview offers both hosts as a first
/// install and the preview itself writes nothing.
#[test]
fn first_install_is_previewed() {
    let f = setup();
    let p = update_preview(&f);
    assert_eq!(host_names(&p), ["claude", "codex"], "{p}");
    for host in ["claude", "codex"] {
        let h = hook_host(&p, host);
        assert_eq!(h["action"], json!("create"), "{h}");
        assert!(h["plan_digest"].is_string(), "{h}");
    }
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    assert!(
        !claude_file(&f).exists(),
        "the preview wrote the Claude hooks"
    );
    assert!(
        !codex_file(&f).exists(),
        "the preview wrote the Codex hooks"
    );
}

/// AC-2 (REQ-1): the profile's hosts decide the offer; OMP is never part of it.
#[test]
fn the_profile_hosts_decide_the_offer() {
    let only_claude = setup();
    profile_with_hosts(&only_claude, &["claude"]);
    assert_eq!(host_names(&update_preview(&only_claude)), ["claude"]);

    let only_codex = setup();
    profile_with_hosts(&only_codex, &["codex"]);
    assert_eq!(host_names(&update_preview(&only_codex)), ["codex"]);

    let all = setup();
    profile_with_hosts(&all, &["claude", "codex", "omp"]);
    assert_eq!(host_names(&update_preview(&all)), ["claude", "codex"]);
}

/// AC-3 (REQ-2): a hooks key the user set is a conflict that is shown and left; the other host
/// is still installed.
#[test]
fn a_conflict_is_shown_and_left() {
    let f = setup();
    fs::create_dir_all(f.project.join(".claude")).unwrap();
    fs::write(claude_file(&f), USER_HOOK).unwrap();
    let p = update_preview(&f);
    let claude = hook_host(&p, "claude");
    assert_eq!(claude["action"], json!("conflict"), "{claude}");
    assert!(
        claude["conflicts"]
            .as_array()
            .is_some_and(|c| !c.is_empty())
            || claude["reason"].is_string(),
        "a conflict names what is in the way: {claude}"
    );
    assert_eq!(hook_host(&p, "codex")["action"], json!("create"), "{p}");
    let done = update_confirm(&f);
    assert_eq!(
        fs::read_to_string(claude_file(&f)).unwrap(),
        USER_HOOK,
        "the user's file was written: {done}"
    );
    assert!(
        codex_file(&f).exists(),
        "the Codex install was stopped: {done}"
    );
}

/// AC-4 (REQ-4, REQ-5): one confirmation writes both host files and leaves one receipt that names
/// both hosts.
#[test]
fn confirm_installs_both_with_one_receipt() {
    let f = setup();
    let done = update_confirm(&f);
    assert!(claude_file(&f).exists(), "{done}");
    assert!(codex_file(&f).exists(), "{done}");
    let hooks = &done["hooks_result"];
    assert_eq!(hooks["state"], json!("applied"), "{hooks}");
    let receipts = receipt_dirs(&f);
    assert_eq!(receipts.len(), 1, "{receipts:?} {hooks}");
    let receipt: Value =
        serde_json::from_slice(&fs::read(receipts[0].join("receipt.json")).unwrap()).unwrap();
    let hosts: Vec<&str> = receipt["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap())
        .collect();
    assert_eq!(hosts, ["claude", "codex"], "{receipt}");
    // Nothing is left to offer once both are in place.
    assert_eq!(update_preview(&f)["nothing_to_apply"], json!(true));
}

/// AC-5 (REQ-6): `hooks-revert` with the receipt removes what the first install created.
#[test]
fn revert_removes_what_the_install_created() {
    let f = setup();
    let done = update_confirm(&f);
    let receipt = done["hooks_result"]["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let p = f.project.to_str().unwrap();
    let out = result(&run(
        &f,
        &[
            "hooks-revert",
            p,
            "--receipt",
            &receipt,
            "--permission-reference",
            REF,
        ],
    ));
    assert_eq!(out["state"], json!("reverted"), "{out}");
    assert!(!claude_file(&f).exists(), "{out}");
    assert!(!codex_file(&f).exists(), "{out}");
}

/// AC-6 (REQ-7): a plan digest from before a change in the project is refused and writes nothing.
#[test]
fn a_stale_digest_is_refused() {
    let f = setup();
    let digest = update_preview(&f)["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    // After the preview, the user puts their own hooks into the Claude settings.
    fs::create_dir_all(f.project.join(".claude")).unwrap();
    fs::write(claude_file(&f), USER_HOOK).unwrap();
    let out = confirm_with(&f, &digest);
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(false), "a stale digest was accepted: {v}");
    assert_eq!(fs::read_to_string(claude_file(&f)).unwrap(), USER_HOOK);
    assert!(!codex_file(&f).exists(), "nothing is written on a refusal");
}
