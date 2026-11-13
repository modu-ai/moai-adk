//! #2392 F6: native `scope-check` inspects the write_scope on disk and blocks any link it
//! reaches. Only disposable fixture paths are used; nothing outside the temp dir is touched.
#![cfg(unix)]
mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{fs, os::unix::fs::symlink, path::Path, process::Command};

const MINE: &str = "aaaaaaaaaaaaaaaa";
const OTHER: &str = "bbbbbbbbbbbbbbbb";

fn record(root: &Path, head: &str, key: &str, scope: &[&str]) -> Value {
    json!({
        "schema_version": 1,
        "task_ref": {"provider":"github","host":"github.com","owner":"fixture","repo":"fixture","repository_id":101,"issue_id":23920,"issue_number":2392},
        "task_key": key, "logical_key": key, "owner": {"kind":"session","id":"fixture"},
        "epoch": 1, "base_revision": head,
        "worktree": {"common_dir":root.join(".git"),"git_dir":root.join(".git"),"path":root,"head":head},
        "write_scope": scope, "spec_ref": {"id":"S","revision":"r","digest":null},
        "comment_id": 1, "claim_token": "Moai-Claim: 2392#1", "lease_until": "2099-01-01T00:00:00Z",
        "state": "running", "effects": [], "last_observed_at": "2026-11-11T21:00:00Z"
    })
}

struct Fixture {
    _tmp: tempfile::TempDir,
    root: std::path::PathBuf,
    outside: std::path::PathBuf,
    head: String,
}

/// A committed synthetic project with `safe/` and a sibling `outside/sentinel`.
fn fixture() -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let parent = fs::canonicalize(tmp.path()).unwrap();
    let root = parent.join("project");
    let outside = parent.join("outside");
    fs::create_dir_all(root.join("safe")).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    fs::create_dir_all(&outside).unwrap();
    fs::write(outside.join("sentinel"), b"ORIGINAL").unwrap();
    fs::write(root.join("README.md"), "fixture\n").unwrap();
    common::git::git_ok(&root, &["init", "-q", "-b", "main"]);
    common::git::git_ok(&root, &["add", "README.md"]);
    common::git::git_ok(
        &root,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "commit",
            "-q",
            "-m",
            "fixture",
        ],
    );
    let head = common::git::git(&root, &["rev-parse", "HEAD"]);
    Fixture {
        _tmp: tmp,
        root,
        outside,
        head,
    }
}

fn journal(f: &Fixture) -> std::path::PathBuf {
    f.root.join(".git/.moai-cli/claims/journal.log")
}

fn seed(f: &Fixture, records: &[Value]) {
    fs::create_dir_all(journal(f).parent().unwrap()).unwrap();
    let lines: String = records
        .iter()
        .map(|r| format!("{}\n", json!({"event": "claimed", "record": r})))
        .collect();
    fs::write(journal(f), lines).unwrap();
}

fn scope_check(f: &Fixture, extra: &[&str]) -> Value {
    let output = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("scope-check")
        .arg(&f.root)
        .args(["--task", MINE])
        .args(extra)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", f.root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    let envelope: Value = serde_json::from_slice(&output.stdout).unwrap();
    assert!(
        output.status.success() && envelope["ok"] == true,
        "{envelope}: {}",
        String::from_utf8_lossy(&output.stderr)
    );
    envelope["result"].clone()
}

/// The decision and `reason:path` of every finding for `scope`, alone in the journal.
fn check(f: &Fixture, scope: &[&str]) -> (String, Vec<String>) {
    seed(f, &[record(&f.root, &f.head, MINE, scope)]);
    let r = scope_check(f, &[]);
    assert_eq!(r["physical_check"]["root"], json!(f.root), "{r}");
    let findings = r["physical_check"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|x| {
            format!(
                "{}:{}",
                x["reason"].as_str().unwrap(),
                x["path"].as_str().unwrap()
            )
        })
        .collect();
    (r["decision"].as_str().unwrap().to_string(), findings)
}

#[test]
fn links_reached_by_any_scope_form_block() {
    let f = fixture();
    symlink(&f.outside, f.root.join("safe/link")).unwrap();
    let blocked = |scope: &[&str], path: &str| {
        let (decision, findings) = check(&f, scope);
        assert_eq!(decision, "blocked", "{scope:?}");
        assert_eq!(findings, [format!("symlink:{path}")], "{scope:?}");
    };
    blocked(&["safe/**"], "safe/link");
    blocked(&["safe"], "safe/link");
    blocked(&["safe/link"], "safe/link");
    blocked(&["safe/link/sentinel"], "safe/link");
    blocked(&["safe/*"], "safe/link");
    blocked(&["safe/l*k"], "safe/link");
    // Names match case-insensitively, like `overlap`: a variant spelling still reaches the link.
    blocked(&["SAFE/**"], "safe/link");
    blocked(&["docs/a.md", "safe/**"], "safe/link");
    // An entry with no path part covers the whole repository (Codex review 4545024323).
    blocked(&["."], "safe/link");
    blocked(&["./"], "safe/link");
    assert_eq!(fs::read(f.outside.join("sentinel")).unwrap(), b"ORIGINAL");
}

#[test]
fn internal_dangling_and_nested_links_block_without_traversal() {
    let f = fixture();
    fs::create_dir_all(f.root.join("safe/a/b")).unwrap();
    fs::create_dir_all(f.root.join("other")).unwrap();
    symlink(f.root.join("other"), f.root.join("safe/a/b/inner")).unwrap();
    let (decision, findings) = check(&f, &["safe/**"]);
    assert_eq!(
        (decision.as_str(), findings),
        ("blocked", vec!["symlink:safe/a/b/inner".to_string()])
    );
    fs::remove_file(f.root.join("safe/a/b/inner")).unwrap();
    symlink(f.root.join("missing"), f.root.join("safe/a/dead")).unwrap();
    let (decision, findings) = check(&f, &["safe/a"]);
    assert_eq!(
        (decision.as_str(), findings),
        ("blocked", vec!["symlink:safe/a/dead".to_string()])
    );
    // A link to itself (a cycle) is observed, never followed.
    fs::remove_file(f.root.join("safe/a/dead")).unwrap();
    symlink(f.root.join("safe/a/loop"), f.root.join("safe/a/loop")).unwrap();
    let (decision, findings) = check(&f, &["safe/**"]);
    assert_eq!(
        (decision.as_str(), findings),
        ("blocked", vec!["symlink:safe/a/loop".to_string()])
    );
}

#[test]
fn absolute_and_parent_entries_block() {
    let f = fixture();
    symlink(&f.outside, f.root.join("safe/link")).unwrap();
    let (decision, findings) = check(&f, &["safe/link/../x"]);
    assert_eq!(decision, "blocked");
    assert_eq!(findings, ["parent_component:safe/link/../x"]);
    let (decision, findings) = check(&f, &["../outside/**"]);
    assert_eq!(
        (decision.as_str(), findings),
        (
            "blocked",
            vec!["parent_component:../outside/**".to_string()]
        )
    );
    let abs = format!("{}/**", f.outside.display());
    let (decision, findings) = check(&f, &[abs.as_str()]);
    assert_eq!(
        (decision.as_str(), findings),
        ("blocked", vec![format!("absolute:{abs}")])
    );
}

#[test]
fn a_whole_repository_entry_without_links_stays_clear() {
    // Negative control for the root walk: `.`-like entries block only when a link exists.
    let f = fixture();
    fs::create_dir_all(f.root.join("safe/sub")).unwrap();
    fs::write(f.root.join("safe/sub/a.rs"), b"").unwrap();
    for scope in [&["."][..], &["./"], &["./."], &[" "]] {
        let (decision, findings) = check(&f, scope);
        assert_eq!(decision, "clear", "{scope:?}: {findings:?}");
        assert!(findings.is_empty(), "{scope:?}");
    }
}

#[test]
fn plain_scopes_and_links_outside_the_scope_stay_clear() {
    let f = fixture();
    fs::create_dir_all(f.root.join("safe/sub")).unwrap();
    fs::write(f.root.join("safe/sub/a.rs"), b"").unwrap();
    fs::create_dir_all(f.root.join("elsewhere")).unwrap();
    symlink(&f.outside, f.root.join("elsewhere/link")).unwrap();
    symlink(&f.outside, f.root.join("safe/link.md")).unwrap();
    for scope in [
        &["safe/sub/**"][..],
        &["safe/sub"],
        &["safe/*.rs"],
        &["safe/sub/*.rs"],
        &["safe/new/file.rs"],
        &["README.md", "safe/sub/a.rs"],
    ] {
        let (decision, findings) = check(&f, scope);
        assert_eq!(decision, "clear", "{scope:?}: {findings:?}");
        assert!(findings.is_empty(), "{scope:?}");
    }
    let r = scope_check(&f, &[]);
    assert_eq!(r["record_update_request"], Value::Null);
}

#[test]
fn integration_order_does_not_lift_a_link_block_and_the_journal_records_it() {
    let f = fixture();
    symlink(&f.outside, f.root.join("safe/link")).unwrap();
    seed(
        &f,
        &[
            record(&f.root, &f.head, MINE, &["safe/**"]),
            record(&f.root, &f.head, OTHER, &["safe/**"]),
        ],
    );
    let r = scope_check(&f, &["--integration-order", OTHER]);
    assert_eq!(r["decision"], "blocked", "{r}");
    assert_eq!(r["conflicts"][0]["allowed_by_integration_order"], true);
    assert_eq!(r["record_update_request"]["state"], "blocked");
    assert_eq!(r["record_update_request"]["reason"], "scope_symlink");
    let text = fs::read_to_string(journal(&f)).unwrap();
    let last: Value = serde_json::from_str(text.lines().last().unwrap()).unwrap();
    assert_eq!(last["record"]["task_key"], MINE);
    assert_eq!(last["record"]["state"], "blocked");
    let effect = &last["record"]["effects"][0];
    assert_eq!(effect["kind"], "scope_symlink");
    assert_eq!(effect["findings"][0]["path"], "safe/link");
    // The same outcome again adds no event.
    let events = text.lines().count();
    scope_check(&f, &["--integration-order", OTHER]);
    assert_eq!(
        fs::read_to_string(journal(&f)).unwrap().lines().count(),
        events
    );
}
