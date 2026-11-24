//! SPEC-20261121-go-path-removal revision 6 (REQ-13 to REQ-18): `moai update` previews the end
//! of the name-switch routing script (`switch`), replaces the script by a link to the Rust
//! launcher only with its own `--confirm-switch <token>`, and puts the kept copy back with
//! `--revert-switch <stamp>`. Temporary HOME folders only; the real HOME and the real
//! `~/.moai-cli/path/moai` are never read or written. The v1 and v2 scripts are read from the
//! preserved copies of the v2 SPEC at build time.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    os::unix::fs::{symlink, PermissionsExt},
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// The two scripts the step knows: the preserved copies of the v2 SPEC.
const V1: &[u8] = include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1");
const V2: &[u8] = include_bytes!("../.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v2");
/// The relative link the confirmed run puts in place of the script.
const LINK: &str = "moai-cli";

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

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli` (this build), the real package
/// beside it, a launcher link of the kind an install lays beside the script, into the version
/// folder, and `script` as `~/.moai-cli/path/moai`.
fn setup_with(script: &[u8]) -> (TempDir, PathBuf, PathBuf) {
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
    fs::write(&moai, script).unwrap();
    fs::set_permissions(&moai, fs::Permissions::from_mode(0o755)).unwrap();
    symlink(
        "../versions/a/bin/moai-cli",
        moai_file(&home).with_file_name(LINK),
    )
    .unwrap();
    (tmp, home, exe)
}

fn setup() -> (TempDir, PathBuf, PathBuf) {
    setup_with(V1)
}

fn moai_file(home: &Path) -> PathBuf {
    home.join(".moai-cli/path/moai")
}

fn launcher_file(home: &Path) -> PathBuf {
    moai_file(home).with_file_name(LINK)
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

/// Whether `path/moai` is the relative link the confirmed run makes.
fn is_the_link(home: &Path) -> bool {
    fs::read_link(moai_file(home)).is_ok_and(|t| t == Path::new(LINK))
}

#[test]
fn the_preview_offers_the_link_and_writes_nothing() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let s = &p["switch"];
    assert_eq!(s["action"], json!("link"), "{p}");
    assert_eq!(s["script"], json!("v1"), "{p}");
    assert_eq!(s["link_target"], json!(LINK), "{p}");
    assert_eq!(
        s["launcher"],
        json!(launcher_file(&home).display().to_string()),
        "{p}"
    );
    assert_eq!(s["writes"], json!(false));
    let token = s["confirm_token"].as_str().unwrap();
    assert!(token.starts_with("switch:"), "{token}");
    assert_eq!(s["apply_flag"], json!(format!("--confirm-switch {token}")));
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
    assert_eq!(names(&home), vec!["moai", LINK]);
    // The layer's own digest and the other parts are untouched by the new part.
    assert!(p["plan_digest"].as_str().unwrap().starts_with("sha256:"));
}

#[test]
fn v2_gets_the_link_too() {
    let (_t, home, exe) = setup_with(V2);
    let p = preview(&exe, &home);
    assert_eq!(p["switch"]["action"], json!("link"), "{p}");
    assert_eq!(p["switch"]["script"], json!("v2"), "{p}");
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    assert_eq!(r["switch_result"]["state"], json!("applied"), "{r}");
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();
    assert!(is_the_link(&home));
    assert_eq!(
        fs::read(moai_file(&home).with_file_name(format!("moai.before-{stamp}"))).unwrap(),
        V2
    );
}

#[test]
fn confirm_switch_alone_makes_the_link_and_keeps_the_old_script() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    assert_eq!(r["switch_result"]["state"], json!("applied"), "{r}");
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();
    assert!(is_the_link(&home), "moai is a relative link to moai-cli");
    let saved = home.join(format!(".moai-cli/path/moai.before-{stamp}"));
    assert_eq!(
        fs::read(&saved).unwrap(),
        V1,
        "the old script is kept byte for byte"
    );
    assert_eq!(
        fs::metadata(&saved).unwrap().permissions().mode() & 0o777,
        0o755
    );
    assert_eq!(
        r["revert"]["switch"],
        json!(format!("moai update --revert-switch {stamp}")),
        "{r}"
    );
    // Through the link the Rust build answers: the build's own version line.
    let out = Command::new(moai_file(&home))
        .arg("--version")
        .clean_env()
        .env("HOME", &home)
        .env("PATH", "/usr/bin:/bin")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(
        String::from_utf8_lossy(&out.stdout).lines().next(),
        Some(format!("moai {} (rust)", moai_cli::BUILD_VERSION).as_str())
    );
    // The skills layer was not confirmed, so it was not written.
    assert!(!home.join(".claude/skills/moai/SKILL.md").exists());
    // Nothing is offered for the link.
    assert!(preview(&exe, &home).get("switch").is_none());
}

#[test]
fn confirm_and_confirm_switch_together_apply_the_layer_then_the_link() {
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
    assert!(is_the_link(&home));
    // The undo line is reported and comes first in the revert order.
    let stamp = r["switch_result"]["stamp"].as_str().unwrap();
    assert_eq!(
        r["revert"]["switch"],
        json!(format!("moai update --revert-switch {stamp}")),
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
    assert_eq!(names(&home), vec!["moai", LINK]);
}

#[test]
fn a_token_that_is_not_the_previews_is_refused_before_any_write() {
    let (_t, home, exe) = setup();
    for bad in ["switch:0000000000000000", "nonsense", ""] {
        let msg = refused(&run(&exe, &home, &["update", "--confirm-switch", bad]));
        assert!(!msg.is_empty());
        assert_eq!(fs::read(moai_file(&home)).unwrap(), V1, "{bad}");
        assert_eq!(names(&home), vec!["moai", LINK], "{bad}");
        assert!(!home.join(".claude").exists(), "{bad}");
    }
}

/// Without a launcher shown to be the Rust build the step waits: the preview says so and gives
/// no token, and a token that is passed writes nothing.
#[test]
fn without_a_rust_launcher_the_step_is_held_and_nothing_is_written() {
    let (_t, home, exe) = setup();
    fs::remove_file(launcher_file(&home)).unwrap();
    let p = preview(&exe, &home);
    let s = &p["switch"];
    assert_eq!(s["action"], json!("held"), "{p}");
    assert!(
        s["confirm_token"].is_null() && s["apply_flag"].is_null(),
        "{p}"
    );
    assert!(s["reason"].as_str().unwrap().contains(LINK), "{p}");
    let msg = refused(&run(
        &exe,
        &home,
        &["update", "--confirm-switch", "switch:0000000000000000"],
    ));
    assert!(msg.contains("preview again"), "{msg}");
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
    assert_eq!(names(&home), vec!["moai"]);

    // A launcher that answers with a Go line holds the step the same way.
    fs::write(
        launcher_file(&home),
        "#!/bin/sh\necho 'moai-adk v3.2.0-rc.23'\n",
    )
    .unwrap();
    fs::set_permissions(launcher_file(&home), fs::Permissions::from_mode(0o755)).unwrap();
    let p = preview(&exe, &home);
    assert_eq!(p["switch"]["action"], json!("held"), "{p}");
    assert!(p["switch"]["confirm_token"].is_null(), "{p}");
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
}

/// The step that replaced v1 by v2 is gone: a file that is neither known script is left alone
/// and named, even beside a Rust launcher.
#[test]
fn an_unknown_script_is_left_alone_and_named() {
    let (_t, home, exe) = setup_with(b"#!/bin/sh\necho mine\n");
    let p = preview(&exe, &home);
    assert_eq!(p["switch"]["action"], json!("left"), "{p}");
    assert!(p["switch"]["confirm_token"].is_null(), "{p}");
    assert!(
        p["switch"]["reason"].as_str().unwrap().contains("not"),
        "{p}"
    );
    assert!(refused(&run(
        &exe,
        &home,
        &["update", "--confirm-switch", "switch:0000000000000000"]
    ))
    .contains("preview again"));
    assert_eq!(
        fs::read(moai_file(&home)).unwrap(),
        b"#!/bin/sh\necho mine\n"
    );
}

#[test]
fn revert_switch_puts_the_kept_copy_back_without_the_package() {
    let (_t, home, exe) = setup();
    let p = preview(&exe, &home);
    let t = p["switch"]["confirm_token"].as_str().unwrap().to_string();
    let r = result(&run(&exe, &home, &["update", "--confirm-switch", &t]));
    let stamp = r["switch_result"]["stamp"].as_str().unwrap().to_string();

    // A binary with no package beside it still reverts: only the kept copy and `moai` are read.
    let bare = home.join(".moai-cli/versions/b/bin/moai-cli");
    fs::create_dir_all(bare.parent().unwrap()).unwrap();
    common::copy_executable(Path::new(EXE), &bare);
    let out = result(&run(&bare, &home, &["update", "--revert-switch", &stamp]));
    assert_eq!(out["mode"], json!("revert_switch"), "{out}");
    assert_eq!(out["switch_result"]["state"], json!("reverted"), "{out}");
    assert_eq!(out["switch_result"]["script"], json!("v1"), "{out}");
    assert_eq!(fs::read(moai_file(&home)).unwrap(), V1);
    assert!(!is_the_link(&home));

    // A second revert finds no kept copy and changes nothing.
    let msg = refused(&run(&bare, &home, &["update", "--revert-switch", &stamp]));
    assert!(msg.contains("no saved copy"), "{msg}");
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
        assert!(is_the_link(&home), "{args:?}");
    }
    for bad in ["../x", "2026-11-24", "20261124T074722Z/../x", ""] {
        let msg = refused(&run(&exe, &home, &["update", "--revert-switch", bad]));
        assert!(!msg.is_empty(), "{bad}");
        assert!(is_the_link(&home), "{bad}");
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
        for wrong in [
            "did not finish",
            "끝나지 않았어요",
            "up to date",
            "Next",
            "v1",
            "v2",
        ] {
            assert!(!text.contains(wrong), "{wrong}: {text}");
        }
    }
}
