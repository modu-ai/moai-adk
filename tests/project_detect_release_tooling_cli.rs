//! `project-detect` names the release tooling a project already uses and the versioning scheme
//! that tooling assumes (a release drafter reads SemVer bumps from labels, semantic-release from
//! commit types). The `release_tooling` fact lists each tool with the file that shows it, says
//! which scheme they assume, and names a `conflict` line when the versions the project already
//! shows follow another scheme, so the versioning question can warn before a pick breaks the
//! tool. Through the binary in temporary folders only; nothing is written.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{fs, process::Stdio};
use tempfile::TempDir;

/// `project-detect` of `root` through the binary with a temporary HOME of its own (a clean
/// environment: nothing of this machine's HOME is read).
fn detect(root: &TempDir) -> Value {
    let home = TempHome::new();
    let out = home
        .command(&["project-detect", root.path().to_str().unwrap()])
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn write(root: &TempDir, rel: &str, text: &str) {
    let path = root.path().join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

const DRAFTER: &str =
    "name-template: 'v$RESOLVED_VERSION'\nversion-resolver:\n  minor:\n    labels: [feature]\n";

#[test]
fn release_tooling_none_is_an_empty_list_without_a_conflict() {
    let root = TempDir::new().unwrap();
    write(
        &root,
        "package.json",
        "{\"name\":\"a\",\"version\":\"4.2647.3\"}\n",
    );
    let fact = &detect(&root)["release_tooling"];
    assert_eq!(fact["tools"], json!([]), "{fact}");
    assert_eq!(fact["assumes"], Value::Null, "{fact}");
    assert_eq!(fact["conflict"], Value::Null, "{fact}");
}

#[test]
fn release_tooling_names_each_tool_by_its_file_and_the_scheme_it_assumes() {
    let root = TempDir::new().unwrap();
    write(&root, ".github/release-drafter.yml", DRAFTER);
    write(&root, ".releaserc.json", "{\"branches\":[\"main\"]}\n");
    write(&root, "release-please-config.json", "{}\n");
    write(&root, ".changeset/config.json", "{}\n");
    let fact = &detect(&root)["release_tooling"];
    assert_eq!(
        fact["tools"],
        json!([
            {"tool": "changesets", "file": ".changeset/config.json", "assumes": "semver"},
            {"tool": "release-drafter", "file": ".github/release-drafter.yml", "assumes": "semver"},
            {"tool": "release-please", "file": "release-please-config.json", "assumes": "semver"},
            {"tool": "semantic-release", "file": ".releaserc.json", "assumes": "semver"},
        ]),
        "{fact}"
    );
    assert_eq!(fact["assumes"], json!("semver"), "{fact}");
}

#[test]
fn release_tooling_conflicts_with_a_version_that_follows_another_scheme() {
    let root = TempDir::new().unwrap();
    write(&root, ".github/release-drafter.yml", DRAFTER);
    write(
        &root,
        "package.json",
        "{\"name\":\"a\",\"version\":\"4.2647.3\"}\n",
    );
    let result = detect(&root);
    let versioning = result["versioning"]["scheme"].as_str().unwrap().to_string();
    assert_ne!(
        versioning, "semver",
        "the fixture must show another scheme: {result}"
    );
    let line = result["release_tooling"]["conflict"]
        .as_str()
        .unwrap_or_else(|| panic!("a conflict line is expected: {}", result["release_tooling"]));
    assert!(line.contains("release-drafter"), "{line}");
    assert!(line.contains("semver"), "{line}");
    assert!(line.contains(&versioning), "{line}");
}

#[test]
fn release_tooling_has_no_conflict_when_the_versions_already_follow_its_scheme() {
    let root = TempDir::new().unwrap();
    write(&root, ".github/release-drafter.yml", DRAFTER);
    write(
        &root,
        "package.json",
        "{\"name\":\"a\",\"version\":\"1.4.2\"}\n",
    );
    let result = detect(&root);
    assert_eq!(result["versioning"]["scheme"], json!("semver"), "{result}");
    assert_eq!(
        result["release_tooling"]["conflict"],
        Value::Null,
        "{result}"
    );
}

#[test]
fn release_tooling_without_a_version_to_compare_names_no_conflict() {
    let root = TempDir::new().unwrap();
    write(&root, ".github/release-drafter.yaml", DRAFTER);
    let result = detect(&root);
    assert_eq!(
        result["release_tooling"]["tools"][0]["file"],
        json!(".github/release-drafter.yaml")
    );
    assert_eq!(
        result["release_tooling"]["conflict"],
        Value::Null,
        "{result}"
    );
}

#[test]
fn release_tooling_never_follows_a_link_to_a_config() {
    let root = TempDir::new().unwrap();
    write(&root, "elsewhere/drafter.yml", DRAFTER);
    fs::create_dir_all(root.path().join(".github")).unwrap();
    #[cfg(unix)]
    std::os::unix::fs::symlink(
        root.path().join("elsewhere/drafter.yml"),
        root.path().join(".github/release-drafter.yml"),
    )
    .unwrap();
    let fact = &detect(&root)["release_tooling"];
    assert_eq!(fact["tools"], json!([]), "{fact}");
}
