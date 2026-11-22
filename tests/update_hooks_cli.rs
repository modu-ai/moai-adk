//! #2853: the hook layer of `moai update`. The MoAI hook set already installed in the project the
//! command runs from is previewed with the rest of the user layer, applied on the same
//! confirmation with a revert receipt, and left alone (as a conflict) where the user changed it.
//! Everything runs in a temporary HOME and a temporary project (leader decision D3).

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
    /// The PATH every command runs with: a folder with the Rust build as `moai` (the generated
    /// hooks call it, and the rewrite of an older set needs it, #3004 P4-2), then the system.
    path: String,
}

/// A temporary HOME with the binary, the real package and the agent sources in a version folder,
/// and a git project beside it.
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

fn run(f: &Fixture, cwd: &Path, args: &[&str]) -> Output {
    Command::new(&f.exe)
        .args(args)
        .clean_env()
        .env("PATH", &f.path)
        .env("HOME", &f.home)
        .env("LC_ALL", "C")
        .current_dir(cwd)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn update_preview(f: &Fixture) -> Value {
    result(&run(f, &f.project, &["update"]))
}

fn update_confirm(f: &Fixture) -> Value {
    let p = update_preview(f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    result(&run(f, &f.project, &["update", "--confirm", &digest]))
}

/// Installs the hook set for `host` with `hooks-plan` / `hooks-apply`, as a lane does today.
/// Returns the ownership receipt's path relative to the project.
fn install_hooks(f: &Fixture, host: &str) -> String {
    let p = f.project.to_str().unwrap();
    let plan = result(&run(f, &f.project, &["hooks-plan", p, "--host", host]));
    let digest = plan["plan_digest"].as_str().unwrap();
    let applied = result(&run(
        f,
        &f.project,
        &[
            "hooks-apply",
            p,
            "--host",
            host,
            "--plan-digest",
            digest,
            "--permission-reference",
            "test-approval",
        ],
    ));
    applied["receipt_path"].as_str().unwrap().to_string()
}

fn hook_host<'a>(preview: &'a Value, host: &str) -> &'a Value {
    preview["hooks"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|h| h["host"] == json!(host))
        .unwrap_or_else(|| panic!("no {host} in {}", preview["hooks"]))
}

const OLD_COMMAND: &str = "moai-cli hook-receipt --host claude --from-an-older-release";

/// What an older release left: the `PostToolUse` handler names another command, in the settings
/// file and in the ownership receipt alike (MoAI wrote it, the user did not touch it).
fn make_claude_hooks_old(f: &Fixture, receipt: &str) {
    let settings_path = f.project.join(".claude/settings.json");
    let mut settings: Value = serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
    settings["hooks"]["PostToolUse"][0]["hooks"][0]["command"] = json!(OLD_COMMAND);
    let mut bytes = serde_json::to_vec_pretty(&settings).unwrap();
    bytes.push(b'\n');
    fs::write(&settings_path, &bytes).unwrap();
    let receipt_path = f.project.join(receipt);
    let mut r: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    r["owned"]["hooks"] = settings["hooks"].clone();
    r["settings_sha256"] = json!(moai_cli::digest(&bytes));
    let mut rb = serde_json::to_vec_pretty(&r).unwrap();
    rb.push(b'\n');
    fs::write(&receipt_path, rb).unwrap();
}

const RECEIPT_STOP: &str = r#"[{"hooks":[{"type":"command","command":"moai-cli hook-receipt --host HOST","timeout":10}]}]"#;

/// What a release before #2936 R9 S1 installed for Claude Code: the four-event set, with
/// `SessionStart` and `Stop` as plain receipts. Written to the settings file and the ownership
/// receipt alike, so the user is taken not to have touched it.
fn make_claude_hooks_four_event(f: &Fixture, receipt: &str) {
    let settings_path = f.project.join(".claude/settings.json");
    let mut settings: Value = serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
    let group: Value = serde_json::from_str(&RECEIPT_STOP.replace("HOST", "claude")).unwrap();
    settings["hooks"]["SessionStart"] = group.clone();
    settings["hooks"]["Stop"] = group;
    let mut bytes = serde_json::to_vec_pretty(&settings).unwrap();
    bytes.push(b'\n');
    fs::write(&settings_path, &bytes).unwrap();
    let receipt_path = f.project.join(receipt);
    let mut r: Value = serde_json::from_slice(&fs::read(&receipt_path).unwrap()).unwrap();
    r["owned"]["hooks"] = settings["hooks"].clone();
    r["settings_sha256"] = json!(moai_cli::digest(&bytes));
    let mut rb = serde_json::to_vec_pretty(&r).unwrap();
    rb.push(b'\n');
    fs::write(&receipt_path, rb).unwrap();
}

/// #2936 R9 S1 AC-3: a project that holds the earlier four-event set sees it as an update, the
/// confirmation drops `Stop` and moves `SessionStart` to the session-check command, and the next
/// preview has nothing to apply.
#[test]
fn an_installed_four_event_set_is_replanned() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    update_confirm(&f); // the rest of the user layer is current from here on
    make_claude_hooks_four_event(&f, &receipt);

    let p = update_preview(&f);
    assert_eq!(hook_host(&p, "claude")["action"], json!("update"), "{p}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("applied"), "{r}");

    let claude: Value =
        serde_json::from_slice(&fs::read(f.project.join(".claude/settings.json")).unwrap())
            .unwrap();
    assert!(claude["hooks"].get("Stop").is_none(), "{claude}");
    assert_eq!(
        claude["hooks"]["SessionStart"][0]["hooks"][0]["command"],
        json!("moai hook-dispatch --host claude --event session-check")
    );
    let again = update_preview(&f);
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    assert_eq!(hook_host(&again, "claude")["action"], json!("unchanged"));
}

/// #2936 R9 S1 AC-4 (miss direction): an earlier four-event set that the user edited is not
/// rewritten. Claude Code stops at a `hooks` conflict; the file is byte-identical after the
/// confirmation, so no `Stop` entry is deleted.
#[test]
fn an_edited_four_event_set_is_kept() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    update_confirm(&f);
    make_claude_hooks_four_event(&f, &receipt);
    let settings = f.project.join(".claude/settings.json");
    let mut s: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    s["hooks"]["Stop"][0]["hooks"][0]["timeout"] = json!(99);
    fs::write(&settings, serde_json::to_vec_pretty(&s).unwrap()).unwrap();
    let s_before = fs::read(&settings).unwrap();

    let p = update_preview(&f);
    let claude = hook_host(&p, "claude");
    assert_eq!(claude["action"], json!("conflict"), "{claude}");
    assert_eq!(claude["conflicts"], json!(["hooks"]));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(fs::read(&settings).unwrap(), s_before);
}

/// AC-1: an old hook entry shows as an update in the preview, the confirmation applies it with a
/// receipt, the layer then has nothing to apply, and `hooks-revert` brings both files back.
#[test]
fn an_old_hook_set_is_previewed_applied_and_reverted() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    update_confirm(&f); // the rest of the user layer is current from here on
    make_claude_hooks_old(&f, &receipt);
    let settings = f.project.join(".claude/settings.json");
    let old_settings = fs::read(&settings).unwrap();
    let old_receipt = fs::read(f.project.join(&receipt)).unwrap();

    let p = update_preview(&f);
    assert_eq!(p["nothing_to_apply"], json!(false), "{p}");
    assert_eq!(
        p["hooks"]["project"],
        json!(f.project.display().to_string())
    );
    let claude = hook_host(&p, "claude");
    assert_eq!(claude["action"], json!("update"), "{claude}");
    assert!(claude["plan_digest"].is_string());
    assert_eq!(
        fs::read(&settings).unwrap(),
        old_settings,
        "a preview writes nothing"
    );

    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["hooks_result"]["state"], json!("applied"), "{r}");
    let text = fs::read_to_string(&settings).unwrap();
    assert!(!text.contains(OLD_COMMAND), "{text}");
    assert!(text.contains("moai hook-receipt --host claude"), "{text}");
    assert!(!text.contains("moai-cli"), "{text}");
    let id = r["hooks_result"]["receipt"].as_str().unwrap().to_string();
    assert!(r["revert"]["hooks"].as_str().unwrap().contains(&id), "{r}");

    let again = update_preview(&f);
    assert_eq!(again["nothing_to_apply"], json!(true), "{again}");
    assert_eq!(hook_host(&again, "claude")["action"], json!("unchanged"));

    let p_text = f.project.to_str().unwrap();
    let reverted = result(&run(
        &f,
        &f.project,
        &[
            "hooks-revert",
            p_text,
            "--receipt",
            &id,
            "--permission-reference",
            "test-approval",
        ],
    ));
    assert_eq!(reverted["reverted"], json!(["claude"]), "{reverted}");
    assert_eq!(fs::read(&settings).unwrap(), old_settings);
    assert_eq!(fs::read(f.project.join(&receipt)).unwrap(), old_receipt);
    // The restored set is MoAI's again, so the next preview offers the same update.
    assert_eq!(
        hook_host(&update_preview(&f), "claude")["action"],
        json!("update")
    );
    // A receipt reverts once.
    let twice = run(
        &f,
        &f.project,
        &[
            "hooks-revert",
            p_text,
            "--receipt",
            &id,
            "--permission-reference",
            "test-approval",
        ],
    );
    assert!(!twice.status.success());
}

/// `hooks-revert` skips a host whose file changed after the apply and writes nothing to it.
#[test]
fn a_revert_skips_a_host_changed_after_the_apply() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    update_confirm(&f);
    make_claude_hooks_old(&f, &receipt);
    let r = update_confirm(&f);
    let id = r["hooks_result"]["receipt"].as_str().unwrap().to_string();
    let settings = f.project.join(".claude/settings.json");
    let written = fs::read(&settings).unwrap();
    let mut edited = written.clone();
    edited.extend_from_slice(b"\n");
    fs::write(&settings, &edited).unwrap();
    let revert = || {
        result(&run(
            &f,
            &f.project,
            &[
                "hooks-revert",
                f.project.to_str().unwrap(),
                "--receipt",
                &id,
                "--permission-reference",
                "test-approval",
            ],
        ))
    };
    let reverted = revert();
    assert_eq!(reverted["state"], json!("unchanged"), "{reverted}");
    assert_eq!(reverted["reverted"], json!([]), "{reverted}");
    assert_eq!(
        reverted["skipped"][0]["reason"],
        json!("user_changed_since_apply")
    );
    assert_eq!(fs::read(&settings).unwrap(), edited);
    // Nothing was restored, so the receipt still works once the user puts the file back.
    fs::write(&settings, &written).unwrap();
    let again = revert();
    assert_eq!(again["state"], json!("reverted"), "{again}");
    assert_eq!(again["reverted"], json!(["claude"]), "{again}");
}

/// The receipt id is the only thing that keeps `hooks-revert` inside the receipt folder: an id
/// with a path step, without the `hooks-` prefix, or longer than 64 characters is refused before
/// any path is built; a 64-character id of the right shape is looked up (and not found).
#[test]
fn a_receipt_id_of_another_shape_is_refused() {
    let f = setup();
    let p = f.project.to_str().unwrap();
    let revert = |id: &str| {
        run(
            &f,
            &f.project,
            &[
                "hooks-revert",
                p,
                "--receipt",
                id,
                "--permission-reference",
                "test-approval",
            ],
        )
    };
    let len65 = format!("hooks-{}", "a".repeat(59));
    let len64 = format!("hooks-{}", "a".repeat(58));
    assert_eq!((len64.len(), len65.len()), (64, 65));
    for id in [
        "hooks-../../x",
        "../hooks-1",
        "hooks-1/../../x",
        "agents-1",
        "hooks",
        len65.as_str(),
    ] {
        let out = revert(id);
        assert!(!out.status.success(), "{id}");
        let text = String::from_utf8_lossy(&out.stdout);
        assert!(text.contains("unknown hook receipt id"), "{id}: {text}");
    }
    let out = revert(&len64);
    assert!(!out.status.success());
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("hook receipt not found"), "{text}");
}

/// Edits the `PostToolUse` handler's timeout in the settings file only, as a user would.
fn hand_edit_claude_hooks(f: &Fixture) {
    let settings = f.project.join(".claude/settings.json");
    let mut s: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    s["hooks"]["PostToolUse"][0]["hooks"][0]["timeout"] = json!(99);
    fs::write(&settings, serde_json::to_vec_pretty(&s).unwrap()).unwrap();
}

/// A user's edit of a current hook set is kept: MoAI has nothing new to write there, so the
/// host stays `unchanged` and the edit survives a confirmation (the three-way rule of
/// `claude-settings-*`).
#[test]
fn a_hand_edit_of_a_current_hook_set_is_kept() {
    let f = setup();
    install_hooks(&f, "claude");
    update_confirm(&f);
    hand_edit_claude_hooks(&f);
    let settings = f.project.join(".claude/settings.json");
    let edited = fs::read(&settings).unwrap();
    let p = update_preview(&f);
    assert_eq!(hook_host(&p, "claude")["action"], json!("unchanged"), "{p}");
    assert_eq!(p["nothing_to_apply"], json!(true));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(fs::read(&settings).unwrap(), edited);
}

/// AC-2: when MoAI has a newer hook set but the user edited the installed one, the entry is a
/// conflict. The preview names it, nothing is offered, and a confirmation leaves the file as it is.
#[test]
fn hand_edited_hooks_are_conflicts_and_are_never_overwritten() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    update_confirm(&f);
    make_claude_hooks_old(&f, &receipt);
    hand_edit_claude_hooks(&f);
    let settings = f.project.join(".claude/settings.json");
    let s_before = fs::read(&settings).unwrap();

    let p = update_preview(&f);
    let claude = hook_host(&p, "claude");
    assert_eq!(claude["action"], json!("conflict"), "{claude}");
    assert_eq!(claude["conflicts"], json!(["hooks"]));
    assert_eq!(claude["plan_digest"], Value::Null);
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");

    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("unchanged"), "{r}");
    assert_eq!(fs::read(&settings).unwrap(), s_before);
}

/// AC-3: a current hook set adds nothing to apply; a project without MoAI hooks, and a folder
/// outside any project, give an empty layer with its reason, and nothing to apply either.
#[test]
fn a_current_or_absent_hook_set_leaves_nothing_to_apply() {
    let f = setup();
    install_hooks(&f, "claude");
    update_confirm(&f);
    let p = update_preview(&f);
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");
    assert_eq!(hook_host(&p, "claude")["action"], json!("unchanged"));

    let bare = setup();
    update_confirm(&bare);
    let p = update_preview(&bare);
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");
    assert_eq!(p["hooks"]["hosts"], json!([]));
    let reasons: Vec<&str> = p["hooks"]["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["reason"].as_str().unwrap())
        .collect();
    assert_eq!(reasons, ["no MoAI hook set is installed in this project"]);

    let outside = result(&run(&bare, &bare.home, &["update"]));
    assert_eq!(outside["hooks"]["project"], Value::Null, "{outside}");
    assert_eq!(outside["nothing_to_apply"], json!(true));
}

/// A Codex or OMP hook file an earlier release wrote, with its ownership receipt, is no part of
/// the hook layer any more: the preview names no such host and the confirmation leaves both
/// files as they were.
#[test]
fn update_leaves_the_hook_files_of_a_removed_host_alone() {
    let f = setup();
    install_hooks(&f, "claude");
    let files = [
        (".codex/hooks.json", "{\"hooks\": {}}\n"),
        (".moai-cli/manifests/codex-hooks.json", "{}\n"),
        (".omp/extensions/moai-hooks.ts", "// earlier module\n"),
        (".moai-cli/manifests/omp-hooks.json", "{}\n"),
    ];
    for (rel, text) in files {
        let path = f.project.join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    let preview = update_preview(&f);
    let hosts: Vec<&str> = preview["hooks"]["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .map(|h| h["host"].as_str().unwrap())
        .collect();
    assert_eq!(hosts, ["claude"], "{preview}");
    update_confirm(&f);
    for (rel, text) in files {
        assert_eq!(
            fs::read_to_string(f.project.join(rel)).unwrap(),
            text,
            "{rel}"
        );
    }
}

/// The confirmation binds the hook layer: a hook file changed after the preview refuses the
/// whole update before any write.
#[test]
fn a_hook_file_changed_after_the_preview_refuses_the_confirmation() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    update_confirm(&f);
    make_claude_hooks_old(&f, &receipt);
    let p = update_preview(&f);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let settings = f.project.join(".claude/settings.json");
    let mut s = fs::read(&settings).unwrap();
    s.extend_from_slice(b"\n");
    fs::write(&settings, &s).unwrap();
    let out = run(&f, &f.project, &["update", "--confirm", &digest]);
    assert!(!out.status.success());
    assert!(String::from_utf8_lossy(&out.stdout).contains("stale or altered plan"));
    assert_eq!(fs::read(&settings).unwrap(), s);
}

/// Walks `dir`, returning every folder and every file below it with its permission bits.
fn modes_below(dir: &Path, out: &mut Vec<(PathBuf, bool, u32)>) {
    use std::os::unix::fs::PermissionsExt;
    for e in fs::read_dir(dir).unwrap() {
        let e = e.unwrap();
        let meta = fs::symlink_metadata(e.path()).unwrap();
        out.push((e.path(), meta.is_dir(), meta.permissions().mode() & 0o777));
        if meta.is_dir() {
            modes_below(&e.path(), out);
        }
    }
}

/// Audit FAIL 6882751326 B1 (#2985): the hook receipt keeps the bytes of every file it replaces.
/// A settings file can hold a key the user protected with 0600, so the receipt folders are
/// owner-only and the kept copies are 0600 whatever the umask, and git is told to skip them.
#[test]
fn the_hook_receipt_keeps_a_protected_settings_file_privately_under_any_umask() {
    use std::os::unix::fs::PermissionsExt;
    for umask in ["000", "022", "077"] {
        let f = setup();
        let receipt = install_hooks(&f, "claude");
        update_confirm(&f); // the rest of the user layer is current from here on
        let settings_path = f.project.join(".claude/settings.json");
        let mut settings: Value =
            serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
        settings["env"] = json!({"TOKEN_FOR_TEST": "TESTVALUE-not-secret"});
        fs::write(
            &settings_path,
            serde_json::to_vec_pretty(&settings).unwrap(),
        )
        .unwrap();
        fs::set_permissions(&settings_path, fs::Permissions::from_mode(0o600)).unwrap();
        make_claude_hooks_four_event(&f, &receipt);

        let digest = update_preview(&f)["plan_digest"]
            .as_str()
            .unwrap()
            .to_string();
        let out = Command::new("/bin/sh")
            .args(["-c", &format!("umask {umask}; exec \"$0\" \"$@\"")])
            .arg(&f.exe)
            .args(["update", "--confirm", &digest])
            .clean_env()
            .env("PATH", &f.path)
            .env("HOME", &f.home)
            .env("LC_ALL", "C")
            .current_dir(&f.project)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let r = result(&out);
        assert_eq!(
            r["hooks_result"]["state"],
            json!("applied"),
            "umask {umask}: {r}"
        );

        let receipts = f.project.join(".moai-cli/receipts");
        let mut seen = Vec::new();
        modes_below(&receipts, &mut seen);
        let own = fs::symlink_metadata(&receipts)
            .unwrap()
            .permissions()
            .mode()
            & 0o777;
        assert_eq!(own & 0o077, 0, "umask {umask}: receipts folder is {own:o}");
        let mut kept = 0;
        for (path, is_dir, mode) in &seen {
            if path.file_name().is_some_and(|n| n == ".gitignore") {
                continue;
            }
            assert_eq!(
                mode & 0o077,
                0,
                "umask {umask}: {} is {mode:o}",
                path.display()
            );
            if !is_dir && path.to_string_lossy().contains("/before/") {
                let text = fs::read_to_string(path).unwrap();
                kept += usize::from(text.contains("TESTVALUE-not-secret"));
            }
        }
        assert_eq!(
            kept, 1,
            "umask {umask}: the kept settings copy must exist: {seen:?}"
        );
        assert_eq!(
            fs::read_to_string(receipts.join(".gitignore")).unwrap(),
            "*\n",
            "umask {umask}"
        );
    }
}
