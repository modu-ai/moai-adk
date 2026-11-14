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
    Fixture {
        _tmp: tmp,
        home,
        exe,
        project,
    }
}

fn run(f: &Fixture, cwd: &Path, args: &[&str]) -> Output {
    Command::new(&f.exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
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

/// What an older release left: the `Stop` handler names another command, in the settings file
/// and in the ownership receipt alike (MoAI wrote it, the user did not touch it).
fn make_claude_hooks_old(f: &Fixture, receipt: &str) {
    let settings_path = f.project.join(".claude/settings.json");
    let mut settings: Value = serde_json::from_slice(&fs::read(&settings_path).unwrap()).unwrap();
    settings["hooks"]["Stop"][0]["hooks"][0]["command"] = json!(OLD_COMMAND);
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
    assert!(
        text.contains("moai-cli hook-receipt --host claude"),
        "{text}"
    );
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

/// Edits the `Stop` handler's timeout in the settings file only, as a user would.
fn hand_edit_claude_hooks(f: &Fixture) {
    let settings = f.project.join(".claude/settings.json");
    let mut s: Value = serde_json::from_slice(&fs::read(&settings).unwrap()).unwrap();
    s["hooks"]["Stop"][0]["hooks"][0]["timeout"] = json!(99);
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
/// conflict; a Codex hook file changed by hand is one too. The preview names them, nothing is
/// offered, and a confirmation leaves both files as they are.
#[test]
fn hand_edited_hooks_are_conflicts_and_are_never_overwritten() {
    let f = setup();
    let receipt = install_hooks(&f, "claude");
    install_hooks(&f, "codex");
    update_confirm(&f);
    make_claude_hooks_old(&f, &receipt);
    hand_edit_claude_hooks(&f);
    let settings = f.project.join(".claude/settings.json");
    let codex = f.project.join(".codex/hooks.json");
    let mut c = fs::read(&codex).unwrap();
    c.extend_from_slice(b" ");
    fs::write(&codex, &c).unwrap();
    let (s_before, c_before) = (fs::read(&settings).unwrap(), fs::read(&codex).unwrap());

    let p = update_preview(&f);
    let claude = hook_host(&p, "claude");
    assert_eq!(claude["action"], json!("conflict"), "{claude}");
    assert_eq!(claude["conflicts"], json!(["hooks"]));
    assert_eq!(claude["plan_digest"], Value::Null);
    let codex_plan = hook_host(&p, "codex");
    assert_eq!(codex_plan["action"], json!("conflict"), "{codex_plan}");
    assert!(codex_plan["reason"]
        .as_str()
        .unwrap()
        .contains("requires preservation review"));
    assert_eq!(p["nothing_to_apply"], json!(true), "{p}");

    let digest = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&f, &f.project, &["update", "--confirm", &digest]));
    assert_eq!(r["hooks_result"]["state"], json!("unchanged"), "{r}");
    assert_eq!(fs::read(&settings).unwrap(), s_before);
    assert_eq!(fs::read(&codex).unwrap(), c_before);
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
