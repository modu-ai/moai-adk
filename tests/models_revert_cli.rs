//! SPEC-20261119-models-revert-command (card C-47 slice 5): the command `models-revert` puts the
//! launcher settings back from the receipt the model stage of `moai update` left. The save is
//! made through the library in a temporary HOME; the revert runs the real binary.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use moai_cli::home_root;
use moai_cli::update_models::{settings_path, stage};
use serde_json::{json, Value};
use std::fs;
use std::io::Cursor;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const REF: &str = "test-approval";
/// A text that must never appear in the command's output.
const MARKER: &str = "BEFORE-MARKER-TEXT";

struct Home {
    _tmp: TempDir,
    path: PathBuf,
}

fn home() -> Home {
    let tmp = TempDir::new().unwrap();
    let path = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(path.join(".claude")).unwrap();
    Home { _tmp: tmp, path }
}

/// Writes the settings file, saves the claude defaults through the stage and returns the
/// earlier text and the receipt id.
fn saved(h: &Home) -> (String, String) {
    let before = format!("{{ \"note\" : \"{MARKER}\" ,\"launcher\":{{}} }}\n");
    let path = settings_path(&h.path);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, &before).unwrap();
    let mut out = Vec::new();
    let result = stage(
        &h.path,
        &mut Cursor::new("1\n\n\n\n\ny\n".to_string()),
        &mut out,
    );
    assert_eq!(result["state"], json!("applied"), "{result}");
    (before, result["receipt"].as_str().unwrap().to_string())
}

fn run(args: &[&str], home: &Path) -> Output {
    Command::new(EXE)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .current_dir(home.parent().unwrap())
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn revert(h: &Home, receipt: &str, reference: &str) -> Output {
    run(
        &[
            "models-revert",
            h.path.to_str().unwrap(),
            "--receipt",
            receipt,
            "--permission-reference",
            reference,
        ],
        &h.path,
    )
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "not JSON: {} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn marker_file(h: &Home, id: &str) -> PathBuf {
    home_root::path(&h.path)
        .join("receipts")
        .join(format!("{id}.reverted.json"))
}

#[test]
fn models_revert_puts_the_earlier_file_back() {
    let h = home();
    let (before, id) = saved(&h);
    let path = settings_path(&h.path);
    assert_ne!(fs::read_to_string(&path).unwrap(), before);

    let out = revert(&h, &id, REF);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("models.revert"), "{v}");
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert_eq!(v["result"]["receipt"], json!(id));
    assert_eq!(fs::read_to_string(&path).unwrap(), before, "byte for byte");
    assert!(marker_file(&h, &id).is_file(), "the marker is written");
    let printed =
        String::from_utf8_lossy(&out.stdout).to_string() + &String::from_utf8_lossy(&out.stderr);
    assert!(!printed.contains(MARKER), "no settings text in the output");
}

#[test]
fn models_revert_refuses_what_it_cannot_prove_is_its_own() {
    let h = home();
    let (_before, id) = saved(&h);
    let path = settings_path(&h.path);
    let saved_text = fs::read_to_string(&path).unwrap();
    let cases: Vec<(String, &str, &str, &str)> = vec![
        (
            id.clone(),
            "",
            "permission reference required",
            "an empty reference",
        ),
        (
            id.clone(),
            "placeholder",
            "permission reference required",
            "a placeholder reference",
        ),
        (
            "keys-1-1".to_string(),
            REF,
            "invalid models receipt id",
            "an id of another shape",
        ),
        (
            "models-9-9".to_string(),
            REF,
            "models receipt not found",
            "an unknown receipt",
        ),
    ];
    for (given, reference, reason, why) in cases {
        let out = revert(&h, &given, reference);
        assert!(!out.status.success(), "{why}: must fail");
        let v = envelope(&out);
        assert_eq!(v["ok"], json!(false), "{why}: {v}");
        assert_eq!(v["error"]["message"], json!(reason), "{why}: {v}");
        assert_eq!(
            fs::read_to_string(&path).unwrap(),
            saved_text,
            "{why}: unchanged"
        );
    }
    assert!(!marker_file(&h, &id).exists(), "no marker after refusals");
}

#[test]
fn models_revert_skips_a_file_changed_after_the_save() {
    let h = home();
    let (_before, id) = saved(&h);
    let path = settings_path(&h.path);
    let changed = format!("{} ", fs::read_to_string(&path).unwrap());
    fs::write(&path, &changed).unwrap();

    let out = revert(&h, &id, REF);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["state"], json!("skipped"), "{v}");
    assert_eq!(v["result"]["reason"], json!("changed_since_apply"), "{v}");
    assert_eq!(fs::read_to_string(&path).unwrap(), changed, "left alone");
    assert!(!marker_file(&h, &id).exists(), "no marker after a skip");
}

#[test]
fn models_revert_is_in_the_help() {
    let h = home();
    let all = run(&["help", "--all"], &h.path);
    let text =
        String::from_utf8_lossy(&all.stdout).to_string() + &String::from_utf8_lossy(&all.stderr);
    assert!(text.contains("models-revert"), "{text}");
    let one = run(&["help", "models-revert"], &h.path);
    let text =
        String::from_utf8_lossy(&one.stdout).to_string() + &String::from_utf8_lossy(&one.stderr);
    assert!(text.contains("--receipt"), "{text}");
    assert!(text.contains("--permission-reference"), "{text}");
    // Audit N3: the per-command help prints the example row of `EXAMPLES`.
    assert!(
        text.contains("moai models-revert --receipt <models-id> --permission-reference <ref> ~"),
        "{text}"
    );
}
