//! #2871 follow-up S1: `moai setting list|get` through the real binary. Every case runs with a
//! temporary HOME; the real home and its settings are never read. The command only reads, so each
//! case also compares the whole HOME tree before and after.
#![cfg(unix)]

mod common;
use common::CleanEnv;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;
const FILE: &str = ".moai-cli/config/launcher.yaml";

fn home(file: Option<&str>) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    if let Some(text) = file {
        let path = home.join(FILE);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    (dir, home)
}

fn tree(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(dir: &Path, root: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        for entry in fs::read_dir(dir).unwrap() {
            let path = entry.unwrap().path();
            let name = path.strip_prefix(root).unwrap().display().to_string();
            if path.is_dir() {
                out.insert(format!("{name}/"), vec![]);
                walk(&path, root, out);
            } else {
                out.insert(name, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn run(home: &Path, args: &[&str]) -> (Value, String) {
    let before = tree(home);
    let output = Command::new(BIN)
        .clean_env()
        .args(args)
        .current_dir(home)
        .env("HOME", home)
        .env("GIT_CEILING_DIRECTORIES", home)
        .output()
        .unwrap();
    assert_eq!(tree(home), before, "{args:?} changed the HOME tree");
    let stdout = String::from_utf8(output.stdout).unwrap();
    let stderr = String::from_utf8(output.stderr).unwrap();
    let v: Value = serde_json::from_str(stdout.trim()).unwrap_or(Value::Null);
    (v, format!("{stdout}{stderr}"))
}

fn entry<'a>(v: &'a Value, key: &str) -> &'a Value {
    v["result"]["settings"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["key"] == key)
        .unwrap()
}

#[test]
fn list_without_a_file_shows_the_defaults_and_writes_nothing() {
    let (_d, home) = home(None);
    let (v, _) = run(&home, &["setting", "list"]);
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "setting.read");
    assert_eq!(v["result"]["file"]["state"], "absent");
    assert_eq!(v["result"]["writes"], false);
    assert_eq!(
        entry(&v, "install.hosts")["value"],
        json!(["claude", "codex"])
    );
    assert_eq!(entry(&v, "install.hosts")["source"], "default");
    assert_eq!(entry(&v, "launcher.role_access")["value"], "full");
}

#[test]
fn list_and_get_show_the_file_value_with_its_source() {
    let (_d, home) = home(Some(
        r#"{"install":{"hosts":["omp"]},"launcher":{"role_access":"prompt"}}"#,
    ));
    let (v, _) = run(&home, &["setting", "list"]);
    assert_eq!(entry(&v, "install.hosts")["value"], json!(["omp"]));
    assert_eq!(entry(&v, "install.hosts")["source"], "file");
    assert_eq!(v["result"]["file"]["state"], "present");
    let (g, _) = run(&home, &["setting", "get", "launcher.role_access"]);
    assert_eq!(g["result"]["setting"]["value"], "prompt");
    assert_eq!(g["result"]["setting"]["valid"], true);
}

#[test]
fn an_unknown_key_or_a_wrong_form_is_refused() {
    let (_d, home) = home(Some(r#"{"launcher":{"token":"abc-not-shown"}}"#));
    for args in [
        &["setting", "get", "launcher.token"][..],
        &["setting", "get"][..],
        &["setting", "list", "install.hosts"][..],
        &["setting", "set", "install.hosts"][..],
    ] {
        let (v, text) = run(&home, args);
        assert_ne!(v["ok"], true, "{args:?}: {text}");
        assert!(!text.contains("abc-not-shown"), "{args:?}: {text}");
    }
}

#[test]
fn an_invalid_value_is_flagged_and_never_quoted() {
    let secret = "not-a-host-9f3a-xyz";
    let (_d, home) = home(Some(&format!(r#"{{"install":{{"hosts":["{secret}"]}}}}"#)));
    let (v, text) = run(&home, &["setting", "list"]);
    assert_eq!(entry(&v, "install.hosts")["valid"], false);
    assert!(!text.contains(secret), "{text}");
}

#[test]
fn a_broken_file_is_an_error_not_a_default() {
    let (_d, home) = home(Some("{not json"));
    let (v, text) = run(&home, &["setting", "list"]);
    assert_ne!(v["ok"], true, "{text}");
    assert!(!text.contains("{not json"), "{text}");
}

#[test]
fn the_command_is_listed_in_the_full_help() {
    let (_d, home) = home(None);
    let (_, text) = run(&home, &["help", "--all"]);
    assert!(text.contains("setting"), "{text}");
}
