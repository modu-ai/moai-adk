//! C-33 S3b (SPEC-20261119-update-switch-layer, AC-6): `moai update` previews the name-switch
//! script's replacement (`switch`), applies it only with its own `--confirm-switch <token>`, and
//! puts the saved copy back with `--revert-switch <stamp>`. Temporary HOME folders only; the real
//! HOME and the real `~/.moai-cli/path/moai` are never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::PermissionsExt,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// The old script (S2a's pinned copy) and the new one the binary carries.
const V1: &[u8] = include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1");
const V2: &[u8] = include_bytes!("../install/moai-switch.sh");

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

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli`, the real package beside it and
/// the old name-switch script as `~/.moai-cli/path/moai`.
fn setup() -> (TempDir, PathBuf, PathBuf) {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let folder = home.join(".moai-cli/versions/a");
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    let manifest = Path::new(env!("CARGO_MANIFEST_DIR"));
    copy_tree(&manifest.join("skills/moai"), &folder.join("skills/moai"));
    copy_tree(&manifest.join("agents"), &folder.join("agents"));
    let moai = home.join(".moai-cli/path/moai");
    fs::create_dir_all(moai.parent().unwrap()).unwrap();
    fs::write(&moai, V1).unwrap();
    fs::set_permissions(&moai, fs::Permissions::from_mode(0o755)).unwrap();
    (tmp, home, exe)
}

fn moai_file(home: &Path) -> PathBuf {
    home.join(".moai-cli/path/moai")
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "not JSON ({e}): {} {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    assert!(
        !out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["error"]["message"].as_str().unwrap().to_string()
}

fn preview(exe: &Path, home: &Path) -> Value {
    result(&run(exe, home, &["update"]))
}

/// The names in `~/.moai-cli/path/`, sorted.
fn names(home: &Path) -> Vec<String> {
    let mut v: Vec<String> = fs::read_dir(home.join(".moai-cli/path"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    v.sort();
    v
}

#[test]
fn the_preview_offers_the_switch_and_writes_nothing() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let s = &p["switch"];
    assert_eq!(s["action"], json!("replace"), "{p}");
    assert_eq!(s["writes"], json!(false));
    let token = s["confirm_token"].as_str().unwrap();
    assert!(token.starts_with("switch:"), "{token}");
    assert_eq!(s["apply_flag"], json!(format!("--confirm-switch {token}")));
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
    assert_eq!(names(&home), vec!["moai"]);
    // The layer's own digest and the other parts are untouched by the new part.
    assert!(p["plan_digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn confirm_switch_alone_replaces_only_the_script_and_keeps_the_old_one() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    assert_eq!(r["switch_result"]["state"], json!("applied"), "{r}");
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V2);
    let saved = home.join(format!(".moai-cli/path/moai.before-{stamp}"));
    assert_eq!(fs::read(&saved).unwrap(), V1);
    assert_eq!(
        r["revert"]["switch"],
        json!(format!("moai-cli update --revert-switch {stamp}")),
        "{r}"
    );
    // The skills layer was not confirmed, so it was not written.
    assert!(!home.join(".claude/skills/moai/SKILL.md").exists());
    // Nothing is offered for the new script.
    assert!(preview(&exe, &home).get("switch").is_none());
}

#[test]
fn confirm_and_confirm_switch_together_apply_the_layer_then_the_script() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(
        &exe,
        &home,
        &["update", "--confirm", &top, "--confirm-switch", &t],
    ));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["switch_result"]["state"], json!("applied"), "{r}");
    assert!(home.join(".claude/skills/moai/SKILL.md").is_file());
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V2);
    // The undo line is reported and comes first in the revert order.
    let stamp = r["switch_result"]["stamp"].as_str().unwrap();
    assert_eq!(
        r["revert"]["switch"],
        json!(format!("moai-cli update --revert-switch {stamp}")),
        "{r}"
    );
    assert_eq!(r["revert_order"][0], json!("switch"), "{r}");
}

#[test]
fn the_layer_digest_alone_never_confirms_the_switch() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm", &top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert!(r.get("switch_result").is_none(), "{r}");
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
    assert_eq!(names(&home), vec!["moai"]);
}

#[test]
fn a_token_that_is_not_the_previews_is_refused_before_any_write() {
    let (_t, home, exe) = setup();
    for bad in ["switch:0000000000000000", "nonsense", ""] {
        let msg = refused(&run(&exe, &home, &["update", "--confirm-switch", bad]));
        assert!(!msg.is_empty());
        assert_eq!(fs::read(moai_file(&home)).unwrap(), V1, "{bad}");
        assert_eq!(names(&home), vec!["moai"], "{bad}");
        assert!(!home.join(".claude").exists(), "{bad}");
    }
}

#[test]
fn revert_switch_puts_the_saved_copy_back_without_the_package() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();

    // A binary with no package beside it still reverts: only the saved copy and `moai` are read.
    let bare = home.join(".moai-cli/versions/b/bin/moai-cli");
    fs::create_dir_all(bare.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(EXE), &bare);
    let out = result(&run(&bare, &home, &["update", "--revert-switch", &stamp]));
    assert_eq!(out["mode"], json!("revert_switch"), "{out}");
    assert_eq!(out["switch_result"]["state"], json!("reverted"), "{out}");
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);

    // A second revert finds `moai` is no longer the new script and changes nothing.
    let msg = refused(&run(&bare, &home, &["update", "--revert-switch", &stamp]));
    assert!(!msg.is_empty());
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
}

#[test]
fn revert_switch_is_refused_beside_any_confirm_flag_and_for_a_bad_stamp() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();
    for extra in [
        vec!["--confirm", "sha256:x"],
        vec!["--confirm-cleanup", "x"],
        vec!["--confirm-home", "x"],
        vec!["--confirm-switch", "x"],
    ] {
        let mut args = vec!["update", "--revert-switch", stamp.as_str()];
        args.extend(extra.iter());
        let msg = refused(&run(&exe, &home, &args));
        assert!(msg.contains("cannot be combined"), "{msg}");
        assert_eq!(fs::read(moai_file(&home)).unwrap(), V2, "{args:?}");
    }
    for bad in ["../x", "2026-11-19", "20261119T074722Z/../x", ""] {
        let msg = refused(&run(&exe, &home, &["update", "--revert-switch", bad]));
        assert!(!msg.is_empty(), "{bad}");
        assert_eq!(fs::read(moai_file(&home)).unwrap(), V2, "{bad}");
    }
}

/// Audit B1 (d99930df2a): on a terminal `--revert-switch` drew the preview's screen, which called a
/// reverted result "did not finish" and "up to date". The result the CLI prints is drawn as one
/// line that says the old script is back, in both languages and without colour or symbols.
#[test]
fn the_revert_switch_result_is_drawn_as_a_success_line() {
    use moai_cli::ui::Painter;
    use moai_cli::ui_text::Lang;
    use moai_cli::update_tux::{screen, Asked, View};

    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();
    let out = result(&run(&exe, &home, &["update", "--revert-switch", &stamp]));
    assert_eq!(out["mode"], json!("revert_switch"), "{out}");

    for (lang, ok_word) in [(Lang::En, "put back"), (Lang::Ko, "되돌렸어요")] {
        let view = View {
            painter: Painter {
                color: false,
                unicode: false,
                lang,
            },
            width: 80,
            home: "/home/u".into(),
            project: None,
        };
        let text = screen(&out, Asked::NotAsked, &view);
        assert!(text.contains(ok_word), "{text}");
        assert_eq!(text.lines().count(), 1, "one line: {text}");
        for wrong in ["did not finish", "끝나지 않았어요", "up to date", "Next"] {
            assert!(!text.contains(wrong), "{wrong}: {text}");
        }
    }
}
