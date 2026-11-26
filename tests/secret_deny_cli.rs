//! `secret-deny-plan` (SPEC-20261122-secret-file-deny-rules, slice A): temporary projects and
//! HOME folders only. The verb runs the built binary; the walk bound is driven at library level.

#![cfg(unix)]

mod common;

use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::os::unix::fs::{symlink, PermissionsExt};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use tempfile::TempDir;

/// A value planted in a key file; deliberately not shaped like any credential.
const PLANTED: &str = "planted-value-not-a-secret-7f3a";

/// This repository's `.claudeignore` as read on 2026-11-26, comments and order kept.
const REPOSITORY_LIST: &str = include_str!("fixtures/secret_deny/claudeignore");

/// Its entries in file order with the rule each one converts to (the table under the SPEC's
/// Plan) and `yes` for a recommended entry: `entry<TAB>rule<TAB>yes|no`.
const EXPECTED_RULES: &str = include_str!("fixtures/secret_deny/rules.tsv");

fn expected_rules() -> Vec<(&'static str, &'static str, bool)> {
    EXPECTED_RULES
        .lines()
        .map(|line| {
            let mut cells = line.split('\t');
            let mut cell = || cells.next().expect("three cells in rules.tsv");
            (cell(), cell(), cell() == "yes")
        })
        .collect()
}

fn project() -> (TempDir, PathBuf) {
    let dir = TempDir::new().expect("temp dir");
    let root = dir.path().join("project");
    fs::create_dir_all(&root).expect("project folder");
    common::git::git_ok(&root, &["init"]);
    (dir, root)
}

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().expect("parent")).expect("parent folder");
    fs::write(path, text).expect("write fixture file");
}

/// Runs the preview of the built binary; answers (stdout, stderr, envelope).
fn preview(root: &Path) -> (String, String, Value) {
    let home = TempDir::new().expect("verb home");
    let out = Command::new(common::MOAI_BIN)
        .arg("secret-deny-plan")
        .arg(root)
        .current_dir(root)
        .env("PATH", "/usr/sbin:/usr/bin:/bin:/sbin")
        .env("HOME", home.path())
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("run secret-deny-plan");
    let stdout = String::from_utf8_lossy(&out.stdout).into_owned();
    let stderr = String::from_utf8_lossy(&out.stderr).into_owned();
    assert_eq!(
        out.status.code(),
        Some(0),
        "preview exits 0: {stdout}{stderr}"
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).expect("envelope");
    assert_eq!(envelope["operation"], "secret_deny.plan");
    assert_eq!(
        fs::read_dir(home.path()).expect("home").count(),
        0,
        "the preview writes nothing under HOME"
    );
    (stdout, stderr, envelope)
}

fn strings(value: &Value) -> Vec<String> {
    value
        .as_array()
        .expect("array")
        .iter()
        .map(|v| v.as_str().expect("string").to_string())
        .collect()
}

fn rule_names(list: &Value) -> Vec<String> {
    list.as_array()
        .expect("rule array")
        .iter()
        .map(|r| r["rule"].as_str().expect("rule").to_string())
        .collect()
}

fn find<'a>(list: &'a Value, rule: &str) -> &'a Value {
    list.as_array()
        .expect("rule array")
        .iter()
        .find(|r| r["rule"] == rule)
        .unwrap_or_else(|| panic!("no rule {rule} in {list}"))
}

/// Every path under `root` (`.git` aside) with what it holds: bytes, a link target or a folder.
fn snapshot(root: &Path) -> BTreeMap<PathBuf, Vec<u8>> {
    fn visit(dir: &Path, into: &mut BTreeMap<PathBuf, Vec<u8>>) {
        for entry in fs::read_dir(dir).expect("read dir") {
            let path = entry.expect("entry").path();
            if path.file_name().is_some_and(|n| n == ".git") {
                continue;
            }
            let meta = fs::symlink_metadata(&path).expect("metadata");
            if meta.is_dir() {
                into.insert(path.clone(), b"folder".to_vec());
                visit(&path, into);
            } else {
                into.insert(path.clone(), fs::read(&path).expect("read"));
            }
        }
    }
    let mut into = BTreeMap::new();
    visit(root, &mut into);
    into
}

#[test]
fn preview_lists_rules_and_writes_nothing() {
    let (_dir, root) = project();
    for rel in [
        ".env",
        "apps/web/.env",
        "apps/web/.env.local",
        "certs/server.pem",
        "certs/server.key",
        "keys/ci-service-account.json",
        ".npmrc",
        "README.md",
        "src/env.rs",
        "service-account.json",
    ] {
        put(&root, rel, "x\n");
    }
    put(&root, ".claude/settings.json", "{\"model\": \"opus\"}\n");
    let before = snapshot(&root);

    let (_, _, envelope) = preview(&root);
    let result = &envelope["result"];
    let keys = &result["key_rules"];
    // `.pypirc` matched nothing, so it proposes no rule.
    assert_eq!(
        rule_names(keys),
        [
            "Read(.env)",
            "Read(.env.*)",
            "Read(*.pem)",
            "Read(*.key)",
            "Read(*-service-account.json)",
            "Read(.npmrc)",
        ]
    );
    assert_eq!(
        strings(&find(keys, "Read(.env)")["matched"]),
        [".env", "apps/web/.env"]
    );
    assert_eq!(
        strings(&find(keys, "Read(.env.*)")["matched"]),
        ["apps/web/.env.local"]
    );
    assert_eq!(
        strings(&find(keys, "Read(*-service-account.json)")["matched"]),
        ["keys/ci-service-account.json"]
    );
    let settings = result["settings_file"].as_str().expect("settings file");
    assert!(
        settings.ends_with("project/.claude/settings.json"),
        "{settings}"
    );
    assert_eq!(result["claudeignore"]["state"], "absent");
    assert_eq!(result["walk"]["complete"], true);
    assert_eq!(snapshot(&root), before, "the preview changes no file");
}

#[test]
fn claudeignore_entries_are_picked_one_by_one() {
    let (_dir, root) = project();
    let table = expected_rules();
    assert_eq!(table.len(), 34, "the SPEC table holds 34 entries");
    // Forms the repository's list does not hold: a negation and a root-anchored folder.
    let list = format!("{REPOSITORY_LIST}!.env.example\n/generated/\n!certs/dev.pem\n");
    put(&root, ".claudeignore", &list);
    put(&root, ".env.production", "x\n");
    put(&root, ".env.example", "x\n");
    put(&root, "web/node_modules/pkg/index.js", "x\n");
    put(&root, "build.log", "x\n");
    put(&root, "generated/a.txt", "x\n");
    put(&root, "sub/generated/b.txt", "x\n");
    put(&root, "docs-site/public/page.html", "x\n");
    put(&root, "logs", "a file, not a folder\n");
    common::git::git_ok(&root, &["add", "-f", ".env.production", "build.log"]);

    let (_, _, envelope) = preview(&root);
    let ignore = &envelope["result"]["claudeignore"];
    assert_eq!(ignore["state"], "present");
    let entries = ignore["entries"].as_array().expect("entries");
    assert_eq!(entries.len(), table.len() + 3, "no blank or comment line");
    for (entry, expected) in entries.iter().zip(table) {
        assert_eq!(entry["entry"], expected.0);
        assert_eq!(entry["rule"], expected.1, "{}", expected.0);
        assert_eq!(entry["recommended"], expected.2, "{}", expected.0);
        let warning = entry["warning"].as_str().expect("warning");
        assert!(warning.contains("blocks editing"), "{warning}");
    }
    // Line numbers are those of the file, comments and blank lines counted.
    assert_eq!(entries[0]["line"], 5);

    let list = &ignore["entries"];
    let negation = find(list, "Read(!.env.example)");
    assert_eq!(negation["recommended"], false);
    assert_eq!(strings(&negation["matched"]), [".env.example"]);
    assert!(
        negation["warning"].as_str().unwrap().contains("carves"),
        "a negation says what it does instead: {negation}"
    );
    // A negation naming a key file is still never recommended: it would open the file.
    assert_eq!(find(list, "Read(!certs/dev.pem)")["recommended"], false);
    let anchored = find(list, "Read(/generated/**/*)");
    assert_eq!(strings(&anchored["matched"]), ["generated/a.txt"]);

    let production = find(list, "Read(.env.production)");
    assert_eq!(strings(&production["matched"]), [".env.production"]);
    assert_eq!(strings(&production["tracked"]), [".env.production"]);
    let modules = find(list, "Read(**/node_modules/**/*)");
    assert_eq!(
        strings(&modules["matched"]),
        ["web/node_modules/pkg/index.js"]
    );
    assert!(strings(&modules["tracked"]).is_empty());
    assert_eq!(
        strings(&find(list, "Read(*.log)")["tracked"]),
        ["build.log"]
    );
    assert_eq!(
        strings(&find(list, "Read(docs-site/public/**/*)")["matched"]),
        ["docs-site/public/page.html"]
    );
    // A folder entry matches no file of that name, and an entry with no file is still listed.
    assert!(strings(&find(list, "Read(**/logs/**/*)")["matched"]).is_empty());
    assert!(strings(&find(list, "Read(*.pyc)")["matched"]).is_empty());
}

#[test]
fn matched_file_is_never_opened() {
    let (_dir, root) = project();
    put(&root, ".env", &format!("TOKEN={PLANTED}\n"));
    put(&root, "locked/server.key", PLANTED);
    let locked = root.join("locked/server.key");
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o000)).expect("chmod");
    put(&root, ".claudeignore", ".env\n*.key\n");

    let (stdout, stderr, envelope) = preview(&root);
    let keys = &envelope["result"]["key_rules"];
    assert_eq!(strings(&find(keys, "Read(.env)")["matched"]), [".env"]);
    assert_eq!(
        strings(&find(keys, "Read(*.key)")["matched"]),
        ["locked/server.key"]
    );
    let all = format!("{stdout}{stderr}");
    assert!(
        !all.contains(PLANTED),
        "a matched file's content is printed"
    );
    assert!(
        !all.to_ascii_lowercase().contains("permission denied"),
        "a matched file was opened: {all}"
    );
    fs::set_permissions(&locked, fs::Permissions::from_mode(0o600)).expect("chmod back");
}

#[test]
fn a_linked_ignore_list_is_not_read() {
    let (_dir, root) = project();
    put(&root, ".env", &format!("{PLANTED}\n"));
    symlink(root.join(".env"), root.join(".claudeignore")).expect("list link");

    let (stdout, stderr, envelope) = preview(&root);
    let ignore = &envelope["result"]["claudeignore"];
    assert_eq!(ignore["state"], "not_read", "{ignore}");
    assert!(ignore["entries"].as_array().expect("entries").is_empty());
    assert!(
        !format!("{stdout}{stderr}").contains(PLANTED),
        "the link's target was read as the list"
    );
}

#[test]
fn the_preview_runs_no_program_the_repository_configures() {
    let (dir, root) = project();
    put(&root, "README.md", "x\n");
    common::git::git_ok(&root, &["add", "README.md"]);
    let marker = dir.path().join("fsmonitor-ran");
    let program = dir.path().join("fsmonitor.sh");
    fs::write(&program, format!("#!/bin/sh\n: > '{}'\n", marker.display())).expect("write program");
    fs::set_permissions(&program, fs::Permissions::from_mode(0o755)).expect("chmod");
    common::git::git_ok(
        &root,
        &["config", "core.fsmonitor", program.to_str().expect("utf-8")],
    );

    let (_, _, envelope) = preview(&root);
    assert_eq!(envelope["result"]["tracked_files"], "known");
    assert!(!marker.exists(), "the preview ran the repository's program");
}

#[test]
fn the_walk_is_bounded_and_stays_in_the_project() {
    let (dir, root) = project();
    let outside = dir.path().join("outside");
    put(&outside, "stolen.pem", "x\n");
    put(&root, ".gitignore", "ignored/\n");
    put(&root, "ignored/deploy.key", "x\n");
    put(&root, "nested/.git/HEAD", "ref: refs/heads/main\n");
    put(&root, "nested/inner.pem", "x\n");
    put(&root, "linked-checkout/.git", "gitdir: /nowhere\n");
    put(&root, "linked-checkout/other.pem", "x\n");
    put(&root, ".git/hooks/own.pem", "x\n");
    symlink(&outside, root.join("linked-folder")).expect("folder link");

    let (_, _, envelope) = preview(&root);
    let result = &envelope["result"];
    assert_eq!(
        strings(&find(&result["key_rules"], "Read(*.key)")["matched"]),
        ["ignored/deploy.key"]
    );
    assert_eq!(
        rule_names(&result["key_rules"]),
        ["Read(*.key)"],
        "nothing from .git, a nested checkout or behind a link"
    );
    assert_eq!(
        strings(&result["walk"]["skipped_nested_checkouts"]),
        ["linked-checkout", "nested"]
    );
    assert_eq!(result["walk"]["bound"], 100_000);
    assert_eq!(result["walk"]["complete"], true);

    // Past the bound the list says it is partial; the walk takes shallow entries first, so a
    // key file at the root is found however large a folder below it is.
    for n in 0..40 {
        put(&root, &format!("bulk/file-{n:02}.txt"), "x\n");
    }
    put(&root, ".env", "x\n");
    let partial = moai_cli::secret_deny::plan_with_bound(&root, 12).expect("plan");
    assert_eq!(partial["walk"]["bound"], 12);
    assert_eq!(partial["walk"]["complete"], false);
    let note = partial["walk"]["note"].as_str().expect("note");
    assert!(note.contains("partial") && note.contains("12"), "{note}");
    assert_eq!(
        strings(&find(&partial["key_rules"], "Read(.env)")["matched"]),
        [".env"]
    );
}

#[test]
fn preview_states_rule_forms_and_trust_behavior() {
    let (_dir, root) = project();
    put(&root, ".env", "x\n");
    put(&root, "a.pem", "x\n");
    put(&root, ".claudeignore", "*.log\n!keep.log\nbuild/\n");

    let (_, _, envelope) = preview(&root);
    let result = &envelope["result"];
    let mut rules = rule_names(&result["key_rules"]);
    rules.extend(rule_names(&result["claudeignore"]["entries"]));
    assert_eq!(rules.len(), 5, "{rules:?}");
    for rule in &rules {
        assert!(rule.starts_with("Read("), "{rule}");
    }
    let applies = result["applies"].as_str().expect("applies");
    for part in ["trust", "headless", "--setting-sources user", "SDK"] {
        assert!(applies.contains(part), "{part} missing: {applies}");
    }
    let no_effect = result["claudeignore"]["note"].as_str().expect("note");
    assert!(no_effect.contains("no effect"), "{no_effect}");
}
