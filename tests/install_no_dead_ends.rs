//! SPEC-20261121-install-no-dead-ends: the install flow leaves no dead ends. Two pins ride
//! with this slice: the update preview offers the guided backup-and-replace for an unmanaged
//! skills/moai folder (REQ-3), and an apply without its flag never touches that folder. The
//! guarded role-access pick (REQ-4) is pinned as a contract test that stays ignored until the
//! session slice lands (the reservation of `src/session.rs` sits with another card). Every
//! home is temporary; the real HOME, the installed binary and the network are never touched.

#![cfg(unix)]

mod common;
use common::{CleanEnv, MOAI_BIN};

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

fn skill_md(body: &str) -> String {
    format!("---\nname: moai\ndescription: fixture package\n---\n# MoAI\n{body}\n")
}

struct Fx {
    _tmp: TempDir,
    home: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    Fx { _tmp: tmp, home }
}

/// `~/.moai-cli/versions/<version>/bin/moai-cli` with its package beside it, and an empty
/// agent source folder `agents/` — the layout an installed version folder has, so `update`
/// sees the staged package the same way it would after a real install.
fn install_version(home: &Path, version: &str) -> PathBuf {
    let folder = home.join(".moai-cli/versions").join(version);
    let bin = folder.join("bin");
    fs::create_dir_all(&bin).unwrap();
    fs::create_dir_all(folder.join("agents")).unwrap();
    let exe = bin.join("moai-cli");
    common::copy_executable(Path::new(MOAI_BIN), &exe);
    fs::hard_link(&exe, bin.join("moai")).unwrap();
    for (rel, body) in [
        ("SKILL.md", skill_md("staged v1")),
        ("workflows/update.md", "# UPDATE\nv1\n".to_string()),
    ] {
        let p = folder.join("skills/moai").join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    exe
}

/// A Go-era `skills/moai` folder no install receipt owns: the situation the guided
/// backup-and-replace exists for.
fn plant_foreign_package(home: &Path) -> PathBuf {
    let p = home.join(".claude/skills/moai/SKILL.md");
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(&p, "---\nname: moai\n---\nGo-era fixture\n").unwrap();
    p
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap()
        .wait_with_output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope JSON on stdout: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// The plan digest exactly as the installer reads it: the `confirm_with` sentence is the one
/// documented non-interactive handle on the plan (install/install.sh parses the same field).
fn digest_of(out: &Output) -> String {
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    let prefix = "\"confirm_with\":\"moai update --confirm sha256:";
    let start = text
        .find(prefix)
        .unwrap_or_else(|| panic!("confirm_with names the digest: {text}"));
    let hex: String = text[start + prefix.len()..].chars().take(64).collect();
    assert!(
        hex.len() == 64 && hex.chars().all(|c| c.is_ascii_hexdigit()),
        "digest is 64 hex digits, got: {hex}"
    );
    format!("sha256:{hex}")
}

fn claude_host(result: &Value) -> Value {
    result["plan"]["hosts"]
        .as_array()
        .unwrap_or_else(|| panic!("plan.hosts: {result}"))
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .cloned()
        .unwrap_or_else(|| panic!("claude host in plan.hosts: {result}"))
}

/// REQ-3: the preview meets an unmanaged Go-era skills/moai folder, blocks the host for that
/// reason and offers the guided backup-and-replace with its apply flag spelled out.
#[test]
fn update_preview_offers_guided_backup_replace() {
    let f = fx();
    let exe = install_version(&f.home, "4.2647.1");
    plant_foreign_package(&f.home);

    let out = run(&exe, &f.home, &["update"]);
    let v = envelope(&out);
    assert_eq!(
        v["ok"],
        json!(true),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let result = &v["result"];
    let host = claude_host(result);
    assert_eq!(
        host["preflight"],
        json!("blocked"),
        "unmanaged package blocks: {host}"
    );
    let reason = host["blocked_reason"]
        .as_str()
        .or_else(|| host["blocked"].as_str())
        .unwrap_or_default();
    assert!(
        reason.contains("backup-then-replace"),
        "the blocked reason names the guided path: {reason}"
    );
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(
        text.contains("--confirm-backup-replace backup-replace:claude:"),
        "the preview offers the apply flag: {text}"
    );
}

/// REQ-3: an apply that confirms the plan without the backup-and-replace flag leaves the
/// blocked host's foreign folder exactly as it was (other hosts may still apply; the
/// envelope's overall flag reflects them, per-host results are the contract here).
#[test]
fn update_apply_without_backup_flag_leaves_foreign_folder_untouched() {
    let f = fx();
    let exe = install_version(&f.home, "4.2647.1");
    let foreign = plant_foreign_package(&f.home);
    let before = fs::read(&foreign).unwrap();

    let preview = run(&exe, &f.home, &["update"]);
    let digest = digest_of(&preview);

    let apply = run(&exe, &f.home, &["update", "--confirm", &digest]);
    let v = envelope(&apply);
    let hosts = v["result"]["result"]["hosts"]
        .as_array()
        .unwrap_or_else(|| {
            panic!(
                "hosts in the apply result: {}",
                String::from_utf8_lossy(&apply.stdout)
            )
        });
    let claude = hosts
        .iter()
        .find(|h| h["host"] == json!("claude"))
        .unwrap_or_else(|| panic!("claude host in the apply result: {v}"));
    assert_eq!(
        claude["status"],
        json!("blocked"),
        "the unmanaged host stays blocked without the flag: {claude}"
    );
    assert!(
        claude["files"]
            .as_array()
            .is_none_or(|files| files.is_empty()),
        "no file of the foreign folder is written: {claude}"
    );
    let after = fs::read(&foreign).unwrap();
    assert_eq!(
        before, after,
        "the foreign SKILL.md keeps its bytes without the flag"
    );
}

/// REQ-4 (contract first): a fresh HOME reports the pending guarded pick in the dry-run
/// answer; a HOME with `launcher.role_access` recorded reports no pending pick. Ignored until
/// the session slice lands: the reservation of `src/session.rs` sits with another card, and
/// until then no answer carries the field (observed RED 2026-11-21, card note).
#[test]
#[ignore = "REQ-4 contract: the guarded pick lands with the session slice (src/session.rs is reserved by another card); observed RED 2026-11-21"]
fn dry_run_reports_pending_role_access_pick() {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    let project = tmp.path().join("project");
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&project).unwrap();
    for dir in [".claude/skills/moai", ".agents/skills/moai"] {
        fs::create_dir_all(home.join(dir)).unwrap();
        fs::write(home.join(dir).join("SKILL.md"), "skill\n").unwrap();
    }

    let launch = |args: &[&str]| {
        Command::new(MOAI_BIN)
            .args(args)
            .clean_env()
            .env("HOME", &home)
            .env("USERPROFILE", &home)
            .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
            .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
            .current_dir(&project)
            .stdin(Stdio::null())
            .output()
            .unwrap()
    };

    let out = launch(&["cc", "-l", "--dry-run"]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(
        v["role_access_pick"],
        json!("pending"),
        "a fresh HOME reports the pending guarded pick: {v}"
    );

    let config = home.join(".moai-cli/config/launcher.yaml");
    fs::create_dir_all(config.parent().unwrap()).unwrap();
    fs::write(&config, "launcher:\n  role_access: prompt\n").unwrap();

    let out = launch(&["cc", "-l", "--dry-run"]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert!(
        v.get("role_access_pick").is_none(),
        "a recorded answer means no pending pick: {v}"
    );
}
