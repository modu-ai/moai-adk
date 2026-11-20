//! SPEC-20261120-user-asset-bundles: the Go `moai bundle add | remove | list` port. The
//! commands run in a temporary HOME; the manifest, its lock and journal keep the Go paths,
//! and the list and summary lines keep the Go output contract. The expected pack data is
//! read from the repository's `templates/bundle-packs/packs.json` — the same file the binary
//! embeds — so a catalog change that forgets the data file fails here first.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

/// The pack data as the repository holds it (the binary embeds the same file).
fn packs_data() -> Value {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join("templates/bundle-packs/packs.json");
    serde_json::from_str(&fs::read_to_string(path).unwrap()).unwrap()
}

/// A file the repo ships under templates/bundle-packs/, by its data-file path.
fn repo_pack_file(rel: &str) -> Vec<u8> {
    fs::read(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("templates/bundle-packs")
            .join(rel),
    )
    .unwrap()
}

/// Every file under one data-file path (a folder in the data), relative to it, sorted.
fn repo_pack_files(rel: &str) -> Vec<PathBuf> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        for entry in fs::read_dir(dir).unwrap() {
            let p = entry.unwrap().path();
            if p.is_dir() {
                walk(&p, out);
            } else {
                out.push(p);
            }
        }
    }
    let mut out = Vec::new();
    walk(
        &Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("templates/bundle-packs")
            .join(rel),
        &mut out,
    );
    out.sort();
    out
}

fn run(home: &TempHome, args: &[&str]) -> Output {
    home.command(args).output().unwrap()
}

/// The manifest key of one backend file, assembled so the literal never reads as a
/// `skills/…` repository reference (the file lives in a user's HOME, not in the repo).
fn backend_key() -> String {
    ["claude-skills", "moai-domain-backend", "SKILL.md"].join("/")
}

fn stdout(out: &Output) -> String {
    String::from_utf8_lossy(&out.stdout).into_owned()
}

fn stderr(out: &Output) -> String {
    String::from_utf8_lossy(&out.stderr).into_owned()
}

/// The Go list line for one pack.
fn go_list_line(name: &str, skills: usize, agents: usize) -> String {
    format!("{name:<10} {skills:>3} skills {agents:>3} agents")
}

/// The operation envelope of one run (the Rust output form; `result.text` carries the Go text).
fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("the envelope: {e}; stdout was {}", stdout(out)))
}

#[test]
fn list_prints_every_pack_in_the_go_format_in_name_order() {
    let data = packs_data();
    let mut expected = Vec::new();
    for (name, pack) in data["packs"].as_object().unwrap() {
        let mut skills = 0;
        let mut agents = 0;
        for e in pack["entries"].as_array().unwrap() {
            match e["kind"].as_str().unwrap() {
                "skill" => skills += 1,
                "agent" => agents += 1,
                other => panic!("unknown kind {other}"),
            }
        }
        expected.push(go_list_line(name, skills, agents));
    }
    expected.sort();

    let home = TempHome::new();
    let out = run(&home, &["bundle", "list"]);
    assert!(
        out.status.success(),
        "exit {:?}: {}",
        out.status.code(),
        stderr(&out)
    );
    let env = envelope(&out);
    let lines: Vec<&str> = env["result"]["lines"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l.as_str().unwrap())
        .collect();
    assert_eq!(
        lines, expected,
        "the list lines must match the Go format in name order"
    );
    assert_eq!(env["result"]["text"].as_str().unwrap(), expected.join("\n"));
}

#[test]
fn an_unknown_pack_name_refuses_like_go_and_touches_nothing() {
    let home = TempHome::new();
    let out = run(&home, &["bundle", "add", "nosuchpack"]);
    assert!(!out.status.success(), "an unknown pack must fail");
    let message = envelope(&out)["error"]["message"]
        .as_str()
        .unwrap_or("")
        .to_string();
    assert!(
        message.contains("unknown bundle \"nosuchpack\""),
        "the Go refusal names the value: {message}"
    );
    let data = packs_data();
    let mut names: Vec<&str> = data["packs"]
        .as_object()
        .unwrap()
        .keys()
        .map(String::as_str)
        .collect();
    names.sort();
    assert!(
        message.contains(&format!("valid bundles: {}", names.join(", "))),
        "the Go refusal lists the valid names: {message}"
    );
    let out = run(&home, &["bundle", "remove", "nosuchpack"]);
    assert!(
        !out.status.success(),
        "an unknown pack must fail on remove too"
    );
    assert!(
        !home.path().join(".moai/user-assets.json").exists(),
        "a refusal writes no manifest"
    );
}

#[test]
fn add_installs_the_pack_records_the_go_manifest_and_prints_the_summary() {
    let home = TempHome::new();
    let out = run(&home, &["bundle", "add", "backend"]);
    assert!(
        out.status.success(),
        "exit {:?}: {}",
        out.status.code(),
        stderr(&out)
    );
    let env = envelope(&out);

    // The summary line: the Go shape with the pack's installed count and no removals.
    let data = packs_data();
    let entries = data["packs"]["backend"]["entries"].as_array().unwrap();
    let mut expected_files = Vec::new();
    for e in entries {
        for f in repo_pack_files(e["path"].as_str().unwrap()) {
            let rel = f
                .strip_prefix(Path::new(env!("CARGO_MANIFEST_DIR")).join("templates/bundle-packs"))
                .unwrap()
                .to_string_lossy()
                .into_owned();
            expected_files.push(rel);
        }
    }
    let text = env["result"]["text"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    let first = text.lines().next().unwrap_or("").to_string();
    assert_eq!(
        first,
        format!(
            "bundle add backend: {} installed, 0 refreshed, 0 removed, 0 collision-skipped, 0 divergence-preserved",
            expected_files.len()
        ),
        "the Go summary line: {first:?}"
    );

    // Every installed file exists under the deploy root with the shipped bytes.
    for rel in &expected_files {
        let under_root = rel
            .strip_prefix(".claude/skills/")
            .unwrap_or_else(|| panic!("backend entries deploy to claude-skills: {rel}"));
        let target = home.path().join(".claude").join("skills").join(under_root);
        let want = repo_pack_file(rel);
        assert_eq!(
            fs::read(&target).unwrap_or_else(|e| panic!("{}: {e}", target.display())),
            want,
            "{} carries the shipped bytes",
            target.display()
        );
    }

    // The manifest: the Go shape, one entry per file, provenance filled.
    let manifest: Value = serde_json::from_str(
        &fs::read_to_string(home.path().join(".moai/user-assets.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(manifest["schema_version"], 1);
    assert_eq!(
        manifest["bundles"],
        Value::Array(vec![Value::from("backend")])
    );
    let files = manifest["files"].as_object().unwrap();
    assert_eq!(
        files.len(),
        expected_files.len(),
        "one manifest entry per file"
    );
    for (key, entry) in files {
        assert!(key.starts_with("claude-skills/"), "{key}");
        assert_eq!(entry["bundle"], "backend", "{key}");
        assert!(
            entry["installed_at"]
                .as_str()
                .is_some_and(|s| !s.is_empty()),
            "{key}"
        );
        assert!(
            entry["moai_version"]
                .as_str()
                .is_some_and(|s| !s.is_empty()),
            "{key}"
        );
        let sha = entry["sha256"].as_str().unwrap_or_default();
        assert_eq!(sha.len(), 64, "{key}: a sha256 column");
    }
}

#[test]
fn remove_of_a_lone_pack_deletes_its_files_and_clears_the_journal() {
    let home = TempHome::new();
    run(&home, &["bundle", "add", "backend"]);
    let out = run(&home, &["bundle", "remove", "backend"]);
    assert!(
        out.status.success(),
        "exit {:?}: {}",
        out.status.code(),
        stderr(&out)
    );
    let env = envelope(&out);
    let text = env["result"]["text"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    let first = text.lines().next().unwrap_or("").to_string();
    assert!(
        first.starts_with("bundle remove backend: 0 installed, 0 refreshed, ")
            && first.contains(" removed, 0 collision-skipped, 0 divergence-preserved"),
        "the Go remove summary line: {first:?}"
    );

    let data = packs_data();
    let entries = data["packs"]["backend"]["entries"].as_array().unwrap();
    let removed = first
        .split(" refreshed, ")
        .nth(1)
        .and_then(|s| s.split(" removed,").next())
        .and_then(|s| s.trim().parse::<usize>().ok())
        .expect("the removed count");
    let mut expected_files = 0;
    for e in entries {
        expected_files += repo_pack_files(e["path"].as_str().unwrap()).len();
    }
    assert_eq!(removed, expected_files, "every file of the pack is removed");
    for e in entries {
        assert!(
            !home
                .path()
                .join(".claude/skills")
                .join(
                    e["path"]
                        .as_str()
                        .unwrap()
                        .strip_prefix(".claude/skills/")
                        .unwrap()
                )
                .join("SKILL.md")
                .exists(),
            "{} is gone",
            e["name"].as_str().unwrap()
        );
    }
    let manifest: Value = serde_json::from_str(
        &fs::read_to_string(home.path().join(".moai/user-assets.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(
        manifest["bundles"],
        Value::Array(Vec::new()),
        "the selection is empty"
    );
    assert_eq!(
        manifest["files"].as_object().unwrap().len(),
        0,
        "no manifest entries survive"
    );
    assert!(
        !home.path().join(".moai/user-assets-journal.json").exists(),
        "a remove clears the journal"
    );
}

#[test]
fn a_journal_of_an_interrupted_run_is_reconciled_as_a_collision() {
    let home = TempHome::new();
    // A target whose bytes a user changed, plus a journal that intended our bytes: the
    // mismatched target is preserved as a collision, never overwritten.
    let target = home
        .path()
        .join(".claude/skills/moai-domain-backend/SKILL.md");
    fs::create_dir_all(target.parent().unwrap()).unwrap();
    fs::write(&target, "the user's own bytes\n").unwrap();
    let journal = format!(
        r#"{{"schema_version":1,"bundles_selection":["backend"],"entries":[{{"path":"{key}","expected_sha256":"{}","bundle":"backend","moai_version":"test","installed_at":"2026-11-20T00:00:00Z","complete":false}}]}}"#,
        "0".repeat(64),
        key = backend_key()
    );
    fs::create_dir_all(home.path().join(".moai")).unwrap();
    fs::write(home.path().join(".moai/user-assets-journal.json"), journal).unwrap();

    let out = run(&home, &["bundle", "add", "backend"]);
    assert!(
        out.status.success(),
        "exit {:?}: {}",
        out.status.code(),
        stderr(&out)
    );
    let env = envelope(&out);
    let text = env["result"]["text"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    assert!(
        text.contains(&format!(
            "collision (skipped, your file kept): {}",
            backend_key()
        )),
        "the mismatched target is a collision: {text}"
    );
    assert_eq!(
        fs::read(&target).unwrap(),
        b"the user's own bytes\n",
        "the user's bytes are never overwritten"
    );
    let manifest: Value = serde_json::from_str(
        &fs::read_to_string(home.path().join(".moai/user-assets.json")).unwrap(),
    )
    .unwrap();
    let collisions = manifest["collisions"].as_array().unwrap();
    assert!(
        collisions.iter().any(|c| c["path"] == backend_key()
            && c["first_seen_at"].as_str().is_some_and(|s| !s.is_empty())),
        "the collision is recorded with its first-seen time: {collisions:?}"
    );
    assert!(
        !home.path().join(".moai/user-assets-journal.json").exists(),
        "the run clears the reconciled journal"
    );
}
