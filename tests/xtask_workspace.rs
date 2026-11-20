//! SPEC-20261121-xtask-local-check REQ-1 to REQ-3: the repository's own tools live in the
//! workspace member `xtask`, run as `cargo xtask <tool>`, and never reach a shipped build or
//! the shipped skills. `cargo metadata` answers what cargo itself builds.

use serde_json::Value;
use std::{fs, path::Path, process::Command};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

fn metadata() -> Value {
    let cargo = std::env::var("CARGO").unwrap_or_else(|_| "cargo".to_string());
    let out = Command::new(cargo)
        .args(["metadata", "--no-deps", "--offline", "--format-version", "1"])
        .current_dir(ROOT)
        .output()
        .expect("cargo metadata starts");
    assert!(
        out.status.success(),
        "cargo metadata failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice(&out.stdout).expect("cargo metadata prints JSON")
}

fn package<'a>(meta: &'a Value, name: &str) -> &'a Value {
    meta["packages"]
        .as_array()
        .unwrap()
        .iter()
        .find(|p| p["name"] == name)
        .unwrap_or_else(|| panic!("no package {name} in the workspace"))
}

/// REQ-1: the member exists without dependencies and the alias runs it.
#[test]
fn the_workspace_has_the_xtask_member_and_the_alias() {
    let meta = metadata();
    let members: Vec<&str> = meta["workspace_members"]
        .as_array()
        .unwrap()
        .iter()
        .map(|m| m.as_str().unwrap())
        .collect();
    assert_eq!(members.len(), 2, "{members:?}");
    assert!(
        members.iter().any(|m| m.contains("xtask")),
        "xtask is not a workspace member: {members:?}"
    );
    let xtask = package(&meta, "xtask");
    assert_eq!(
        xtask["dependencies"].as_array().unwrap().len(),
        0,
        "xtask has dependencies: {}",
        xtask["dependencies"]
    );
    assert!(
        xtask["targets"]
            .as_array()
            .unwrap()
            .iter()
            .any(|t| t["name"] == "xtask" && t["kind"][0] == "bin"),
        "xtask has no binary target: {}",
        xtask["targets"]
    );

    let config = fs::read_to_string(Path::new(ROOT).join(".cargo/config.toml")).unwrap();
    let doc: toml_edit::DocumentMut = config.parse().unwrap();
    assert_eq!(
        doc["alias"]["xtask"].as_str(),
        Some("run --quiet --package xtask --"),
        "{config}"
    );
}

/// REQ-2: the root package alone is built by default, and `xtask` is neither published nor
/// marked for distribution, and `moai-cli` does not depend on it.
#[test]
fn the_shipped_build_does_not_contain_xtask() {
    let meta = metadata();
    let defaults = meta["workspace_default_members"].as_array().unwrap();
    assert_eq!(defaults.len(), 1, "{defaults:?}");
    assert!(defaults[0].as_str().unwrap().contains("moai-cli"));

    let xtask = package(&meta, "xtask");
    assert_eq!(xtask["publish"], serde_json::json!([]), "xtask is published");
    assert_eq!(
        xtask["metadata"]["dist"]["dist"],
        serde_json::json!(false),
        "xtask is marked for distribution: {}",
        xtask["metadata"]
    );

    let moai = package(&meta, "moai-cli");
    assert!(
        moai["dependencies"]
            .as_array()
            .unwrap()
            .iter()
            .all(|d| d["name"] != "xtask"),
        "moai-cli depends on xtask"
    );
    let bins: Vec<&str> = moai["targets"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|t| t["kind"][0] == "bin")
        .map(|t| t["name"].as_str().unwrap())
        .collect();
    assert_eq!(bins, ["moai-cli", "moai"], "the shipped binaries changed");
}

fn files_under(dir: &Path, out: &mut Vec<std::path::PathBuf>) {
    let Ok(entries) = fs::read_dir(dir) else {
        return;
    };
    for entry in entries.flatten() {
        let path = entry.path();
        let kind = entry.file_type().unwrap();
        if kind.is_dir() {
            files_under(&path, out);
        } else if kind.is_file() {
            out.push(path);
        }
    }
}

/// REQ-3: xtask is maintainer tooling; the skills and agents a user installs never name it.
#[test]
fn no_shipped_skill_or_agent_names_xtask() {
    let mut files = Vec::new();
    for dir in ["skills", "agents"] {
        files_under(&Path::new(ROOT).join(dir), &mut files);
    }
    assert!(files.len() > 10, "found only {} shipped files", files.len());
    let naming: Vec<_> = files
        .iter()
        .filter(|f| {
            fs::read(f).is_ok_and(|bytes| {
                String::from_utf8_lossy(&bytes)
                    .to_ascii_lowercase()
                    .contains("xtask")
            })
        })
        .collect();
    assert!(naming.is_empty(), "shipped files name xtask: {naming:?}");
}
