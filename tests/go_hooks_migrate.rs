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

/// A PATH that holds no `moai`: the hooks the replacement installs are not checked against a
/// `moai` here (#3004 P4-2 gate tests are below), so these tests do not depend on the machine.
const NO_MOAI_PATH: &str = "/nonexistent-moai-path";

/// A path under the project's state folder as the resolver answers it: `.moai` for these
/// projects, which hold no earlier `.moai-cli` (SPEC-20261121-path-root-moai-default).
fn state(root: &Path, sub: &str) -> PathBuf {
    moai_cli::state_root::path(root, sub)
}

fn apply_no_moai(
    project: &Path,
    home: Option<&Path>,
    digest: &str,
    permission: &str,
) -> Result<Value, String> {
    go_hooks::apply_in(project, home, digest, permission, NO_MOAI_PATH)
}

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

    let out = apply_no_moai(
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
    let archive = state(&root, &format!("receipts/{id}/archive/moai"));
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
    assert!(!state(&root, "manifests/claude-settings.json").exists());
    // A receipt is used once.
    assert!(go_hooks::revert(&root, &id, PERMISSION).is_err());
}

#[test]
fn a_stale_digest_or_a_missing_permission_writes_nothing() {
    let (_d, root) = go_project();
    let settings_before = fs::read(root.join(".claude/settings.json")).unwrap();
    let plan = go_hooks::plan(&root, None).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap();
    assert!(apply_no_moai(&root, None, digest, "").is_err());
    // A script changed after the preview: the digest no longer matches.
    fs::write(
        root.join(".claude/hooks/moai/handle-stop.sh"),
        "#!/bin/sh\necho x\n",
    )
    .unwrap();
    let e = apply_no_moai(&root, None, digest, PERMISSION).unwrap_err();
    assert!(e.contains("stale"), "{e}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_before
    );
    assert!(root.join(".claude/hooks/moai/handle-stop.sh").exists());
    assert!(!state(&root, "receipts").exists());
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
    let e = apply_no_moai(
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
    let out = apply_no_moai(
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
    let receipt = state(&root, "manifests/claude-settings.json");
    fs::create_dir_all(receipt.parent().unwrap()).unwrap();
    fs::write(&receipt, "not json").unwrap();
    let settings_before = fs::read(root.join(".claude/settings.json")).unwrap();
    let plan = go_hooks::plan(&root, None).unwrap();
    let e = apply_no_moai(
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

// ---- SPEC-20261119-hooks-followups: the read-back and the put-back, with installers of our own

const CLAIM_RECORD: &str = "manifests/claude-settings.json";

/// The status of the one `go-hooks-` receipt the apply left behind.
fn the_receipt_status(root: &Path) -> String {
    let mut found = Vec::new();
    for e in fs::read_dir(state(&root, "receipts")).unwrap().flatten() {
        if e.file_name().to_string_lossy().starts_with("go-hooks-") {
            let r = read_json(&e.path().join("receipt.json"));
            found.push(r["status"].as_str().unwrap().to_string());
        }
    }
    assert_eq!(found.len(), 1, "one receipt: {found:?}");
    found.remove(0)
}

struct Before {
    settings: Vec<u8>,
    scripts: Vec<(String, Vec<u8>)>,
}

impl Before {
    fn of(root: &Path) -> Before {
        Before {
            settings: fs::read(root.join(".claude/settings.json")).unwrap(),
            scripts: scripts_of(&root.join(".claude/hooks/moai")),
        }
    }

    /// Everything a failed apply must have put back, record included (it did not exist).
    fn assert_put_back(&self, root: &Path) {
        assert_eq!(
            fs::read(root.join(".claude/settings.json")).unwrap(),
            self.settings,
            "the settings file"
        );
        assert_eq!(
            scripts_of(&root.join(".claude/hooks/moai")),
            self.scripts,
            "the scripts"
        );
        assert!(!state(&root, CLAIM_RECORD).exists(), "the ownership record");
        assert_eq!(the_receipt_status(root), "rolled_back");
    }
}

/// Runs the apply with `install` as its install step; the plan is the one of the project now.
fn apply_with(root: &Path, install: go_hooks::Installer<'_>) -> Result<Value, String> {
    let plan = go_hooks::plan(root, None).unwrap();
    go_hooks::apply_with_installer(
        root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
        NO_MOAI_PATH,
        install,
    )
}

#[test]
fn an_install_that_reports_success_without_writing_the_hooks_is_put_back() {
    // N2: the read-back is the only thing between "the install said Ok" and a project whose Go
    // guards are gone with nothing in their place.
    let (_d, root) = go_project();
    let before = Before::of(&root);
    let e = apply_with(&root, &|_, _, _| Ok(())).unwrap_err();
    assert!(e.contains("does not hold the previewed hooks"), "{e}");
    assert!(e.contains("everything was put back"), "{e}");
    before.assert_put_back(&root);
}

#[test]
fn an_installer_that_writes_something_else_is_undone_with_its_record() {
    let (_d, root) = go_project();
    let before = Before::of(&root);
    let e = apply_with(&root, &|root, _, _| {
        fs::write(root.join(".claude/settings.json"), b"{\"hooks\": {}}\n").unwrap();
        fs::create_dir_all(state(&root, "manifests")).unwrap();
        fs::write(state(&root, CLAIM_RECORD), b"{\"owned\": {}}\n").unwrap();
        Ok(())
    })
    .unwrap_err();
    assert!(e.contains("does not hold the previewed hooks"), "{e}");
    before.assert_put_back(&root);
}

#[test]
fn an_installer_that_fails_is_put_back_the_same_way() {
    let (_d, root) = go_project();
    let before = Before::of(&root);
    let e = apply_with(&root, &|_, _, _| Err("the installer broke".into())).unwrap_err();
    assert!(e.contains("the installer broke"), "{e}");
    assert!(e.contains("everything was put back"), "{e}");
    before.assert_put_back(&root);
}

#[test]
fn the_read_back_failure_is_a_conflict_not_an_internal_error() {
    // N7: the file and the preview disagree, which is a refusal on the project's state.
    let (_d, root) = go_project();
    let e = apply_with(&root, &|_, _, _| Ok(())).unwrap_err();
    assert_eq!(moai_cli::contract::classify(&e).as_str(), "conflict", "{e}");
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
    let out = apply_no_moai(
        root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
    )
    .unwrap();
    out["receipt"].as_str().unwrap().to_string()
}

fn receipt_status(root: &Path, id: &str) -> Value {
    read_json(&state(&root, &format!("receipts/{id}/receipt.json")))["status"].clone()
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
    let archived = state(&root, &format!("receipts/{id}/archive/moai/handle-stop.sh"));
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
    let path = state(&root, &format!("receipts/{id}/receipt.json"));
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
    let manifests = state(&root, "manifests");
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o555)).unwrap();
    let result = go_hooks::revert(&root, &id, PERMISSION);
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o755)).unwrap();
    assert!(result.is_err(), "{result:?}");
    assert_eq!(
        fs::read(root.join(".claude/settings.json")).unwrap(),
        settings_now
    );
    assert!(!root.join(".claude/hooks/moai").exists());
    assert!(state(&root, &format!("receipts/{id}/archive/moai/handle-stop.sh")).is_file());
    assert_eq!(receipt_status(&root, &id), json!("applied"));
    assert_eq!(
        go_hooks::revert(&root, &id, PERMISSION).unwrap()["state"],
        json!("reverted")
    );
}

/// A folder with a `moai` that prints `line` for `--version`, and the PATH that puts it first.
fn fake_moai(line: &str) -> (tempfile::TempDir, String) {
    let dir = tempfile::tempdir().unwrap();
    let bin = fs::canonicalize(dir.path()).unwrap().join("bin");
    fs::create_dir_all(&bin).unwrap();
    common::write_executable(
        &bin.join("moai"),
        &format!("#!/bin/sh\nprintf '%s\\n' '{line}'\n"),
    );
    let path = format!("{}:/usr/bin:/bin", bin.display());
    (dir, path)
}

/// Every file under `root` with its bytes, for a byte-identical tree check.
fn tree(root: &Path) -> Vec<(String, Vec<u8>)> {
    fn walk(dir: &Path, root: &Path, out: &mut Vec<(String, Vec<u8>)>) {
        for e in fs::read_dir(dir).unwrap().flatten() {
            let p = e.path();
            if p.is_dir() {
                walk(&p, root, out);
            } else {
                let rel = p.strip_prefix(root).unwrap().display().to_string();
                out.push((rel, fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(root, root, &mut out);
    out.sort();
    out
}

/// #3004 P4-2: the hooks the replacement installs call `moai`. With a Go `moai` first on PATH
/// they would reach it and the claim guard would stay off, so nothing is written: no moved
/// scripts, no settings rewrite, no receipt. A Rust `moai`, or none on PATH, applies.
#[test]
fn a_go_moai_first_on_path_refuses_the_replacement_and_writes_nothing() {
    let (_d, root) = go_project();
    let before = tree(&root);
    let plan = go_hooks::plan(&root, None).unwrap();
    let digest = plan["plan_digest"].as_str().unwrap();
    let (_g, go_path) = fake_moai("moai-adk v3.2.1");
    let e = go_hooks::apply_in(&root, None, digest, PERMISSION, &go_path).unwrap_err();
    assert!(e.contains("wrong_generation"), "{e}");
    assert!(e.contains("nothing was written"), "{e}");
    assert!(e.contains("Put the Rust moai first on PATH"), "{e}");
    assert_eq!(tree(&root), before, "the tree is byte-identical");
    assert!(!root.join(".moai-cli").exists());

    // The same confirmed plan applies once a Rust `moai` is first.
    let (_r, rust_path) = fake_moai("moai 1.2.3 (rust)");
    let out = go_hooks::apply_in(&root, None, digest, PERMISSION, &rust_path).unwrap();
    assert_eq!(out["state"], json!("applied"), "{out}");
    assert!(!root.join(".claude/hooks/moai").exists());
}

/// No `moai` on PATH is reported by the install (`hooks_active: false`), not refused, as
/// `hooks-apply` always did; a Go `moai` found inside the project is not run and not refused.
#[test]
fn no_moai_on_path_or_one_inside_the_project_still_applies() {
    let (_d, root) = go_project();
    let plan = go_hooks::plan(&root, None).unwrap();
    let out = go_hooks::apply_in(
        &root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
        NO_MOAI_PATH,
    )
    .unwrap();
    assert_eq!(out["state"], json!("applied"), "{out}");

    let (_d, root) = go_project();
    let bin = root.join("bin");
    fs::create_dir_all(&bin).unwrap();
    common::write_executable(&bin.join("moai"), "#!/bin/sh\necho moai-adk v3.2.1\n");
    let plan = go_hooks::plan(&root, None).unwrap();
    let path = format!("{}:/usr/bin:/bin", bin.display());
    let out = go_hooks::apply_in(
        &root,
        None,
        plan["plan_digest"].as_str().unwrap(),
        PERMISSION,
        &path,
    )
    .unwrap();
    assert_eq!(out["state"], json!("applied"), "{out}");
}

/// Through the binary: `go-hooks-plan` carries the gate for the PATH it runs with, and
/// `go-hooks-apply` refuses under a Go `moai` with the tree untouched.
#[test]
fn the_commands_report_and_enforce_the_gate() {
    use common::CleanEnv as _;
    let (_d, root) = go_project();
    let home = tempfile::tempdir().unwrap();
    let run = |op: &str, path: &str, extra: &[&str]| -> Value {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .arg(op)
            .args(extra)
            .arg(&root)
            .env("HOME", home.path())
            .env("PATH", path)
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap()
    };
    let (_g, go_path) = fake_moai("moai-adk v3.2.1");
    let (_r, rust_path) = fake_moai("moai 1.2.3 (rust)");
    let before = tree(&root);
    let under_go = run("go-hooks-plan", &go_path, &[]);
    let under_rust = run("go-hooks-plan", &rust_path, &[]);
    assert_eq!(
        under_go["result"]["gate"]["allowed"],
        json!(false),
        "{under_go}"
    );
    assert!(under_go["result"]["gate"]["reason"]
        .as_str()
        .unwrap()
        .contains("wrong_generation"));
    assert_eq!(
        under_rust["result"]["gate"]["allowed"],
        json!(true),
        "{under_rust}"
    );
    // The gate is a report: the digest is the same under every PATH.
    assert_eq!(
        under_go["result"]["plan_digest"],
        under_rust["result"]["plan_digest"]
    );
    assert_eq!(tree(&root), before);

    let digest = under_go["result"]["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let args = [
        "--plan-digest",
        &digest,
        "--permission-reference",
        PERMISSION,
    ];
    let refused = run("go-hooks-apply", &go_path, &args);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(
        refused.to_string().contains("nothing was written"),
        "{refused}"
    );
    assert_eq!(tree(&root), before);
    let done = run("go-hooks-apply", &rust_path, &args);
    assert_eq!(done["ok"], json!(true), "{done}");
}
