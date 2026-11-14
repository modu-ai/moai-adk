//! #2877 (I5): replacing a project's Go-generation hooks — plan, apply, revert. Temporary
//! projects only; no real HOME and no host session.

#![cfg(unix)]

mod common;

use moai_cli::go_hooks;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const PERMISSION: &str = "test-permission-2877";

fn go_handler(script: &str) -> Value {
    json!({
        "type": "command",
        "command": "bash",
        "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; exit 0",
                 format!("${{CLAUDE_PROJECT_DIR}}/.claude/hooks/moai/{script}")],
        "timeout": 10
    })
}

/// A Go-generation project: only Go hook entries, a Go `permissions` key, the script folder.
fn go_project() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let settings = json!({
        "permissions": {"allow": ["Bash(ls:*)"]},
        "statusLine": {"type": "command", "command": "echo go"},
        "hooks": {
            "SessionStart": [{"matcher": "startup", "hooks": [go_handler("handle-session-start.sh")]}],
            "PreToolUse": [{"matcher": "Write|Edit|Bash", "hooks": [go_handler("handle-pre-tool.sh")]}],
            "Stop": [{"hooks": [go_handler("handle-stop.sh")]}]
        }
    });
    fs::create_dir_all(root.join(".claude/hooks/moai")).unwrap();
    fs::write(
        root.join(".claude/settings.json"),
        serde_json::to_vec_pretty(&settings).unwrap(),
    )
    .unwrap();
    for s in [
        "handle-session-start.sh",
        "handle-pre-tool.sh",
        "handle-stop.sh",
        "handle-notification.sh",
    ] {
        fs::write(
            root.join(".claude/hooks/moai").join(s),
            format!("#!/bin/sh\n# {s}\nexit 0\n"),
        )
        .unwrap();
    }
    (dir, root)
}

fn read_json(path: &Path) -> Value {
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

fn scripts_of(dir: &Path) -> Vec<(String, Vec<u8>)> {
    let mut out: Vec<_> = fs::read_dir(dir)
        .unwrap()
        .flatten()
        .map(|e| {
            (
                e.file_name().to_string_lossy().into_owned(),
                fs::read(e.path()).unwrap(),
            )
        })
        .collect();
    out.sort();
    out
}

#[test]
fn apply_replaces_the_go_hooks_and_revert_puts_every_byte_back() {
    let (_d, root) = go_project();
    let settings_path = root.join(".claude/settings.json");
    let settings_before = fs::read(&settings_path).unwrap();
    let scripts_before = scripts_of(&root.join(".claude/hooks/moai"));

    let plan = go_hooks::plan(&root, None).unwrap();
    assert_eq!(plan["action"], json!("replace"), "{plan}");
    assert_eq!(plan["blocks"], json!([]));
    assert_eq!(
        fs::read(&settings_path).unwrap(),
        settings_before,
        "plan writes nothing"
    );

    let out = go_hooks::apply(
        &root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    assert_eq!(out["state"], json!("applied"), "{out}");
    assert_eq!(out["removed_entries"], json!(3));
    assert_eq!(out["scripts_archived"], json!(4));
    let id = out["receipt"].as_str().unwrap().to_string();

    // The Go entries are gone, the MoAI set is in, other keys are as they were.
    let after = read_json(&settings_path);
    let hooks_text = after["hooks"].to_string();
    assert!(!hooks_text.contains(".claude/hooks/moai/"), "{hooks_text}");
    assert_eq!(
        after["hooks"],
        Value::Object(moai_cli::hooks::hooks_object("claude").unwrap())
    );
    let before: Value = serde_json::from_slice(&settings_before).unwrap();
    assert_eq!(after["permissions"], before["permissions"]);
    assert_eq!(after["statusLine"], before["statusLine"]);
    // The scripts moved into the receipt folder, byte for byte; nothing was deleted.
    assert!(!root.join(".claude/hooks/moai").exists());
    let archive = root.join(format!(".moai-cli/receipts/{id}/archive/moai"));
    assert_eq!(scripts_of(&archive), scripts_before);
    // `hooks-apply`'s ownership receipt now owns the hooks, so `update` sees them installed.
    assert!(moai_cli::hooks::installed(&root, "claude").unwrap());
    assert_eq!(
        moai_cli::hooks::plan(&root, "claude", None).unwrap()["action"],
        json!("unchanged")
    );
    // A second plan has nothing left to replace.
    assert_eq!(
        go_hooks::plan(&root, None).unwrap()["action"],
        json!("none")
    );

    let back = go_hooks::revert(&root, &id, PERMISSION).unwrap();
    assert_eq!(back["state"], json!("reverted"), "{back}");
    assert_eq!(fs::read(&settings_path).unwrap(), settings_before);
    assert_eq!(scripts_of(&root.join(".claude/hooks/moai")), scripts_before);
    assert!(!root
        .join(".moai-cli/manifests/claude-settings.json")
        .exists());
    // A receipt is used once.
    assert!(go_hooks::revert(&root, &id, PERMISSION).is_err());
}

#[test]
fn a_stale_digest_or_a_missing_permission_writes_nothing() {
    let (_d, root) = go_project();
    let settings_before = fs::read(root.join(".claude/settings.json")).unwrap();
    let plan = go_hooks::plan(&root, None).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap();
    assert!(go_hooks::apply(&root, None, digest, "").is_err());
    // A script changed after the preview: the digest no longer matches.
    fs::write(
        root.join(".claude/hooks/moai/handle-stop.sh"),
        "#!/bin/sh\necho x\n",
    )
    .unwrap();
    let e = go_hooks::apply(&root, None, digest, PERMISSION).unwrap_err();
    assert!(e.contains("stale"), "{e}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_before
    );
    assert!(root.join(".claude/hooks/moai/handle-stop.sh").exists());
    assert!(!root.join(".moai-cli/receipts").exists());
}

#[test]
fn hooks_of_another_tool_block_the_replacement() {
    let (_d, root) = go_project();
    let path = root.join(".claude/settings.json");
    let mut s = read_json(&path);
    s["hooks"]["Notification"] =
        json!([{"hooks": [{"type": "command", "command": "orca notify"}]}]);
    fs::write(&path, serde_json::to_vec_pretty(&s).unwrap()).unwrap();
    let plan = go_hooks::plan(&root, None).unwrap();
    assert_eq!(plan["action"], json!("blocked"), "{plan}");
    assert_eq!(plan["blocks"], json!(["other_hooks_present"]));
    let e = go_hooks::apply(
        &root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap_err();
    assert!(e.contains("blocked"), "{e}");
    assert!(root.join(".claude/hooks/moai").exists());
}

#[test]
fn revert_refuses_once_the_settings_changed_since_the_apply() {
    let (_d, root) = go_project();
    let plan = go_hooks::plan(&root, None).unwrap();
    let out = go_hooks::apply(
        &root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    let id = out["receipt"].as_str().unwrap();
    let path = root.join(".claude/settings.json");
    let mut s = read_json(&path);
    s["model"] = json!("user-changed");
    fs::write(&path, serde_json::to_vec_pretty(&s).unwrap()).unwrap();
    let edited = fs::read(&path).unwrap();
    let e = go_hooks::revert(&root, id, PERMISSION).unwrap_err();
    assert!(e.contains("changed since the apply"), "{e}");
    assert_eq!(fs::read(&path).unwrap(), edited);
    assert!(!root.join(".claude/hooks/moai").exists());
    assert!(go_hooks::revert(&root, "../escape", PERMISSION).is_err());
}

#[test]
fn a_failed_install_puts_everything_back() {
    // An ownership receipt nobody can read makes `hooks-plan` fail after the Go entries were
    // taken out: the settings, the receipt and the scripts must all be as before.
    let (_d, root) = go_project();
    let receipt = root.join(".moai-cli/manifests/claude-settings.json");
    fs::create_dir_all(receipt.parent().unwrap()).unwrap();
    fs::write(&receipt, "not json").unwrap();
    let settings_before = fs::read(root.join(".claude/settings.json")).unwrap();
    let plan = go_hooks::plan(&root, None).unwrap();
    let e = go_hooks::apply(
        &root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap_err();
    assert!(e.contains("put back"), "{e}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_before
    );
    assert_eq!(fs::read(&receipt).unwrap(), b"not json");
    assert!(root.join(".claude/hooks/moai/handle-stop.sh").exists());
}

#[test]
fn the_commands_are_wired() {
    let (_d, root) = go_project();
    let home = tempfile::tempdir().unwrap();
    let out = Command::new(common::MOAI_BIN)
        .arg("go-hooks-plan")
        .arg(&root)
        .env("HOME", home.path())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["result"]["action"], json!("replace"), "{v}");
    assert_eq!(v["operation"], json!("go_hooks.plan"));
}

fn applied(root: &Path) -> String {
    let plan = go_hooks::plan(root, None).unwrap();
    let out = go_hooks::apply(
        root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    out["receipt"].as_str().unwrap().to_string()
}

fn receipt_status(root: &Path, id: &str) -> Value {
    read_json(&root.join(format!(".moai-cli/receipts/{id}/receipt.json")))["status"].clone()
}

/// #2893 audit B1: a revert never writes through a link — neither a linked `.claude/hooks`
/// nor a linked `.claude` — and leaves the receipt usable once the folder is real again.
#[test]
fn revert_refuses_a_linked_hooks_or_claude_folder() {
    let (_d, root) = go_project();
    let id = applied(&root);
    let outside = tempfile::tempdir().unwrap();
    fs::remove_dir(root.join(".claude/hooks")).unwrap();
    std::os::unix::fs::symlink(outside.path(), root.join(".claude/hooks")).unwrap();
    let settings_now = fs::read(root.join(".claude/settings.json")).unwrap();
    let e = go_hooks::revert(&root, &id, PERMISSION).unwrap_err();
    assert!(e.contains(".claude/hooks is a link"), "{e}");
    assert_eq!(
        fs::read_dir(outside.path()).unwrap().count(),
        0,
        "nothing moved outside"
    );
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_now
    );
    assert_eq!(receipt_status(&root, &id), json!("applied"));
    // A real folder again: the same receipt reverts.
    fs::remove_file(root.join(".claude/hooks")).unwrap();
    assert_eq!(
        go_hooks::revert(&root, &id, PERMISSION).unwrap()["state"],
        json!("reverted")
    );
    assert!(root.join(".claude/hooks/moai/handle-stop.sh").is_file());

    // `.claude` itself as a link.
    let (_d2, root) = go_project();
    let id = applied(&root);
    let away = root.join("claude-away");
    fs::rename(root.join(".claude"), &away).unwrap();
    std::os::unix::fs::symlink(&away, root.join(".claude")).unwrap();
    let e = go_hooks::revert(&root, &id, PERMISSION).unwrap_err();
    assert!(e.contains(".claude is a link"), "{e}");
    assert!(!away.join("hooks/moai").exists());
    assert_eq!(receipt_status(&root, &id), json!("applied"));
}

/// #2893 audit note 2: an archived script changed while archived is not put back.
#[test]
fn revert_refuses_an_archived_script_changed_since_the_apply() {
    let (_d, root) = go_project();
    let id = applied(&root);
    let archived = root.join(format!(
        ".moai-cli/receipts/{id}/archive/moai/handle-stop.sh"
    ));
    fs::write(&archived, "#!/bin/sh\necho changed\n").unwrap();
    let settings_now = fs::read(root.join(".claude/settings.json")).unwrap();
    let e = go_hooks::revert(&root, &id, PERMISSION).unwrap_err();
    assert!(e.contains("handle-stop.sh changed since the apply"), "{e}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_now
    );
    assert!(!root.join(".claude/hooks/moai").exists());
}

/// #2893 audit note 3: an interrupted apply's receipt names the manual recovery.
#[test]
fn a_pending_receipt_names_the_manual_recovery() {
    let (_d, root) = go_project();
    let id = applied(&root);
    let path = root.join(format!(".moai-cli/receipts/{id}/receipt.json"));
    let mut r = read_json(&path);
    r["status"] = json!("pending");
    fs::write(&path, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    let e = go_hooks::revert(&root, &id, PERMISSION).unwrap_err();
    assert!(e.contains("interrupted") && e.contains("by hand"), "{e}");
}

/// A revert that fails after the scripts moved back puts them into the archive again and leaves
/// the settings as the apply wrote them, so the receipt still matches and can be used later.
#[test]
fn a_revert_that_fails_midway_moves_the_scripts_back_into_the_archive() {
    use std::os::unix::fs::PermissionsExt;
    let (_d, root) = go_project();
    let id = applied(&root);
    let settings_now = fs::read(root.join(".claude/settings.json")).unwrap();
    let manifests = root.join(".moai-cli/manifests");
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o555)).unwrap();
    let result = go_hooks::revert(&root, &id, PERMISSION);
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o755)).unwrap();
    assert!(result.is_err(), "{result:?}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_now
    );
    assert!(!root.join(".claude/hooks/moai").exists());
    assert!(root
        .join(format!(
            ".moai-cli/receipts/{id}/archive/moai/handle-stop.sh"
        ))
        .is_file());
    assert_eq!(receipt_status(&root, &id), json!("applied"));
    assert_eq!(
        go_hooks::revert(&root, &id, PERMISSION).unwrap()["state"],
        json!("reverted")
    );
}
