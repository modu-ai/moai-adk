//! #2871 U2b: `moai update` previews the Go generation's `~/.moai` items (the home layer) beside
//! the user layer, retires them only with its own `--confirm-home <token>`, and prints the
//! revert. Temporary HOME folders only; the real HOME is never read or written. Secrets are
//! synthetic names with placeholder content.

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

/// A temporary HOME with `~/.moai-cli/versions/a/bin/moai-cli` and the real package beside it.
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
    (tmp, home, exe)
}

fn write(root: &Path, rel: &str, body: &str) {
    let p = root.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, body).unwrap();
}

/// A Go-generation `~/.moai`: tool assets, a secret name, user data and a stale registry.
fn go_items(home: &Path) {
    write(
        home,
        ".moai/user-assets.json",
        r#"{"schema_version": 1, "files": {}}"#,
    );
    write(home, ".moai/state/active-sessions.json", "[]");
    write(home, ".moai/credentials.yaml", "token: PLACEHOLDER\n");
    write(home, ".moai/cache/a.bin", "x");
    write(home, ".moai/run/x.pid", "1\n");
    write(home, ".moai/specs/SPEC-1/spec.md", "# user spec\n");
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        // `lsof` lives in /usr/sbin: with it out of reach the process table cannot be read and
        // the home part rightly moves nothing (see the unreadable-table test below).
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn result(out: &Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    assert!(!out.status.success());
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["error"]["message"].as_str().unwrap().to_string()
}

/// Every entry under HOME except the version folder: relative path, kind and bytes.
fn state(home: &Path) -> Vec<(String, String, Vec<u8>)> {
    fn walk(dir: &Path, home: &Path, skip: &Path, out: &mut Vec<(String, String, Vec<u8>)>) {
        let mut names: Vec<_> = fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().path())
            .collect();
        names.sort();
        for p in names {
            if p.starts_with(skip) {
                continue;
            }
            let rel = p.strip_prefix(home).unwrap().display().to_string();
            if p.is_dir() {
                out.push((rel, "dir".into(), Vec::new()));
                walk(&p, home, skip, out);
            } else {
                out.push((rel, "file".into(), fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(home, home, &home.join(".moai-cli/versions"), &mut out);
    out
}

fn preview(exe: &Path, home: &Path) -> Value {
    result(&run(exe, home, &["update"]))
}

fn token(p: &Value) -> String {
    p["home"]["confirm_token"].as_str().unwrap().to_string()
}

#[test]
fn update_home_preview_lists_the_go_items_and_writes_nothing() {
    let (_t, home, exe) = setup();
    let plain = preview(&exe, &home);
    assert!(
        plain.get("home").is_none(),
        "no Go items, no home part: {plain}"
    );
    go_items(&home);
    let before = state(&home);
    let p = preview(&exe, &home);
    assert_eq!(state(&home), before, "a preview writes nothing");
    assert_eq!(p["home"]["action"], json!("retire"), "{p}");
    assert_eq!(p["home"]["generation"], json!("go"), "{p}");
    assert_eq!(p["home"]["writes"], json!(false));
    let t = token(&p);
    assert!(t.starts_with("home:"), "{t}");
    assert_eq!(
        p["home"]["apply_flag"],
        json!(format!("--confirm-home {t}"))
    );
    // `cache`, `run`, `state`, `credentials.yaml`, `user-assets.json` retire; `specs` stays.
    assert_eq!(p["home"]["counts"]["retire"], json!(5), "{p}");
    let specs = p["home"]["entries"]
        .as_array()
        .unwrap()
        .iter()
        .find(|e| e["name"] == "specs")
        .unwrap();
    assert_eq!(specs["action"], json!("keep"), "{specs}");
    // The home part is not in the one digest `--confirm` takes.
    assert_eq!(
        p["plan_digest"], plain["plan_digest"],
        "home is in the digest"
    );
}

#[test]
fn update_confirm_alone_moves_no_go_item() {
    let (_t, home, exe) = setup();
    go_items(&home);
    let p = preview(&exe, &home);
    let top = p["plan_digest"].as_str().unwrap();
    let r = result(&run(&exe, &home, &["update", "--confirm", top]));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert!(r.get("home_result").is_none(), "{r}");
    assert!(r["revert"].get("home").is_none(), "{r}");
    assert!(home.join(".moai/cache/a.bin").is_file());
    assert!(!home.join(".moai-go-retired").exists());
}

#[test]
fn update_confirm_home_retires_the_go_items_and_the_printed_revert_undoes_it() {
    let (_t, home, exe) = setup();
    go_items(&home);
    let before = state(&home);
    let p = preview(&exe, &home);
    let t = token(&p);
    let r = result(&run(&exe, &home, &["update", "--confirm-home", &t]));
    assert_eq!(r["home_result"]["state"], json!("applied"), "{r}");
    assert_eq!(r["home_result"]["moved"], json!(5), "{r}");
    // The layer stays a preview: only the home part was confirmed.
    assert_eq!(r["applied"], json!(false), "{r}");
    assert!(!home.join(".moai/cache").exists());
    assert!(!home.join(".moai/credentials.yaml").exists());
    assert!(
        home.join(".moai/specs/SPEC-1/spec.md").is_file(),
        "user data stays"
    );
    // The next preview has nothing left to retire.
    let again = preview(&exe, &home);
    assert!(
        again["home"].is_null() || again["home"]["action"] == json!("none"),
        "{again}"
    );
    // The printed revert command puts everything back.
    let revert = r["revert"]["home"].as_str().unwrap();
    assert!(revert.contains("home-retire-revert"), "{revert}");
    let stamp = revert
        .split_whitespace()
        .skip_while(|w| *w != "--stamp")
        .nth(1)
        .unwrap();
    let out = run(
        &exe,
        &home,
        &[
            "home-retire-revert",
            "--stamp",
            stamp,
            "--permission-reference",
            "user approved in the fixture",
        ],
    );
    let v = result(&out);
    assert_eq!(v["state"], json!("reverted"), "{v}");
    let after: Vec<_> = state(&home)
        .into_iter()
        .filter(|(n, ..)| !n.starts_with(".moai-go-retired"))
        .collect();
    assert_eq!(after, before);
}

#[test]
fn update_confirm_and_confirm_home_together_apply_the_layer_then_the_home_part() {
    let (_t, home, exe) = setup();
    go_items(&home);
    let p = preview(&exe, &home);
    let top = p["plan_digest"].as_str().unwrap().to_string();
    let t = token(&p);
    let r = result(&run(
        &exe,
        &home,
        &["update", "--confirm", &top, "--confirm-home", &t],
    ));
    assert_eq!(r["applied"], json!(true), "{r}");
    assert_eq!(r["home_result"]["state"], json!("applied"), "{r}");
    assert!(home.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(!home.join(".moai/cache").exists());
}

#[test]
fn update_home_refuses_a_wrong_stale_or_doubled_token_before_any_write() {
    let (_t, home, exe) = setup();
    // No Go items: a token has nothing to match.
    let e = refused(&run(
        &exe,
        &home,
        &["update", "--confirm-home", "home:0000000000000000"],
    ));
    assert!(e.contains("home token"), "{e}");
    go_items(&home);
    let p = preview(&exe, &home);
    let t = token(&p);
    let before = state(&home);
    let e = refused(&run(
        &exe,
        &home,
        &["update", "--confirm-home", "home:0000000000000000"],
    ));
    assert!(e.contains("home token"), "{e}");
    // A cleanup token and a home token together: both touch `~/.moai`, so neither runs.
    let cleanup = p["cleanup"]["confirm_token"]
        .as_str()
        .expect("the `cache` folder is a cleanup candidate")
        .to_string();
    let e = refused(&run(
        &exe,
        &home,
        &[
            "update",
            "--confirm-home",
            &t,
            "--confirm-cleanup",
            &cleanup,
        ],
    ));
    assert!(e.contains("home") && e.contains("cleanup"), "{e}");
    assert_eq!(state(&home), before, "a refused run wrote something");
    // A top-level entry appeared after the preview: the token is stale.
    write(&home, ".moai/appeared-later.bin", "z");
    let before = state(&home);
    let e = refused(&run(&exe, &home, &["update", "--confirm-home", &t]));
    assert!(e.contains("home token"), "{e}");
    assert_eq!(state(&home), before);
    assert!(!home.join(".moai-go-retired").exists());
}

/// A machine whose process table cannot be read moves nothing: the part is shown with every
/// movable item "in use" and no token, and a made-up token is refused.
#[test]
fn update_home_with_an_unreadable_process_table_offers_nothing() {
    let (_t, home, exe) = setup();
    go_items(&home);
    let before = state(&home);
    let out = Command::new(&exe)
        .arg("update")
        .clean_env()
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", &home)
        .env("LC_ALL", "C")
        .env("MOAI_CLI_TEST_PROCESS_CHECK", "unavailable")
        .current_dir(&home)
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let p = result(&out);
    assert_eq!(p["home"]["action"], json!("none"), "{p}");
    assert!(p["home"]["confirm_token"].is_null(), "{p}");
    assert_eq!(p["home"]["counts"]["retire"], json!(0), "{p}");
    assert_eq!(p["home"]["counts"]["skip-in-use"], json!(5), "{p}");
    assert_eq!(state(&home), before);
}

#[test]
fn update_home_shows_a_blocked_part_without_a_token() {
    let (_t, home, exe) = setup();
    go_items(&home);
    let me = std::process::id();
    write(
        &home,
        ".moai/state/active-sessions.json",
        &format!(r#"[{{"pid": {me}}}]"#),
    );
    let p = preview(&exe, &home);
    assert_eq!(p["home"]["action"], json!("blocked"), "{p}");
    assert!(p["home"]["confirm_token"].is_null(), "{p}");
    assert!(p["home"]["apply_flag"].is_null(), "{p}");
    assert!(
        p["home"]["blocks"].to_string().contains("go_session_live"),
        "{p}"
    );
    let e = refused(&run(
        &exe,
        &home,
        &["update", "--confirm-home", "home:0000000000000000"],
    ));
    assert!(e.contains("home token"), "{e}");
}
