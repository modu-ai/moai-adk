//! #2404: `scope-check` drops an overlapping `pr_ready` claim from the conflict candidates only
//! when the forge shows every PR it opened as closed (merged or not); a lookup that cannot be
//! settled keeps the conflict and says why. Literal fixtures: open / merged / closed-unmerged /
//! lookup failure / non-numeric response id / unconfirmed second PR, plus the no-call case.
//! Each test uses a temporary repository and HOME and a stub `gh` (MOAI_FORGE_GH_BIN) that
//! answers per PR number and logs every call; no network, no real `gh`.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const KEY_A: &str = "aaaaaaaaaaaaaaaa";
const KEY_B: &str = "bbbbbbbbbbbbbbbb";

/// PR 11 open, 12 merged, 13 closed without merge, 14 a server error, 15 no `state` field.
const STUB: &str = r#"#!/bin/sh
echo "$*" >> "$(dirname "$0")/calls.log"
case "$*" in
  *pulls/11*) printf 'HTTP/2.0 200 OK\r\n\r\n{"number":11,"state":"open","merged_at":null}';;
  *pulls/12*) printf 'HTTP/2.0 200 OK\r\n\r\n{"number":12,"state":"closed","merged_at":"2026-11-10T16:47:23Z"}';;
  *pulls/13*) printf 'HTTP/2.0 200 OK\r\n\r\n{"number":13,"state":"closed","merged_at":null}';;
  *pulls/14*) printf 'HTTP/2.0 500 Internal Server Error\r\n\r\n{}';;
  *pulls/15*) printf 'HTTP/2.0 200 OK\r\n\r\n{"number":15}';;
  *) printf 'HTTP/2.0 404 Not Found\r\n\r\n{}';;
esac
"#;

fn repo() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap().join("primary");
    fs::create_dir_all(root.join("home")).unwrap();
    fs::write(root.join("README.md"), "x\n").unwrap();
    common::git::git_ok(&root, &["init", "-q", "-b", "main"]);
    common::git::git_ok(&root, &["add", "README.md"]);
    common::git::git_ok(
        &root,
        &[
            "-c",
            "user.email=t@example.invalid",
            "-c",
            "user.name=t",
            "commit",
            "-q",
            "-m",
            "init",
        ],
    );
    common::write_executable(&root.join("gh-stub"), STUB);
    (tmp, root)
}

fn pr_create(response_id: &str, state: &str) -> Value {
    json!({"kind": "pr_create", "operation_id": format!("pr_create-101-{KEY_B}-{response_id}-1"),
           "state": state, "response_id": response_id, "epoch": 1})
}

fn record(key: &str, issue: i64, state: &str, effects: Vec<Value>) -> Value {
    json!({
        "schema_version": 1,
        "task_ref": {"provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
                     "repository_id": 101, "issue_id": issue * 10, "issue_number": issue},
        "task_key": key,
        "logical_key": format!("L{issue}"),
        "owner": {"kind": "session", "id": "agent-x"},
        "epoch": 1,
        "base_revision": "a".repeat(40),
        "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": "a".repeat(40)},
        "write_scope": ["src/lib.rs"],
        "spec_ref": {"id": "S", "revision": "r", "digest": null},
        "comment_id": issue * 100,
        "claim_token": format!("Moai-Claim: {issue}#{}", issue * 100),
        "lease_until": "2099-01-01T00:00:00Z",
        "state": state,
        "effects": effects,
        "last_observed_at": "2026-11-10T12:00:00Z"
    })
}

fn seed(root: &Path, other: Value) {
    let dir = root.join(".git/.moai-cli/claims");
    fs::create_dir_all(&dir).unwrap();
    let lines: String = [record(KEY_A, 1, "running", vec![]), other]
        .iter()
        .map(|r| format!("{}\n", json!({"event": "claimed", "record": r})))
        .collect();
    fs::write(dir.join("journal.log"), lines).unwrap();
}

/// Runs `scope-check --task KEY_A`; returns the result and the stub's call count.
fn scope_check(root: &Path) -> (Value, usize) {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(["scope-check"])
        .arg(root)
        .args(["--task", KEY_A])
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("MOAI_FORGE_GH_BIN", root.join("gh-stub"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let env: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(env["ok"], true, "{env}");
    let calls = fs::read_to_string(root.join("calls.log"))
        .map(|s| s.lines().count())
        .unwrap_or(0);
    (env["result"].clone(), calls)
}

fn decide(other: Value) -> (Value, usize) {
    let (_t, root) = repo();
    seed(&root, other);
    scope_check(&root)
}

#[test]
fn an_open_pr_keeps_the_conflict() {
    let (r, calls) = decide(record(
        KEY_B,
        2,
        "pr_ready",
        vec![pr_create("11", "confirmed")],
    ));
    assert_eq!(r["decision"], "blocked");
    assert_eq!(r["conflicts"][0]["pr_status"]["pr_state"], "open");
    assert_eq!(r["skipped_ended"], json!([]));
    assert_eq!(calls, 1);
}

#[test]
fn a_merged_pr_is_skipped() {
    let (r, calls) = decide(record(
        KEY_B,
        2,
        "pr_ready",
        vec![pr_create("12", "confirmed")],
    ));
    assert_eq!(r["decision"], "clear");
    assert_eq!(r["conflicts"], json!([]));
    let s = &r["skipped_ended"][0];
    assert_eq!(s["other_task"], KEY_B);
    assert_eq!(s["pr_status"]["prs"][0]["number"], 12);
    assert_eq!(
        s["pr_status"]["prs"][0]["merged_at"],
        "2026-11-10T16:47:23Z"
    );
    assert_eq!(r["record_update"]["applied"], false, "nothing written");
    assert_eq!(calls, 1);
}

#[test]
fn a_pr_closed_without_merge_is_skipped() {
    let (r, _) = decide(record(
        KEY_B,
        2,
        "pr_ready",
        vec![pr_create("13", "confirmed")],
    ));
    assert_eq!(r["decision"], "clear");
    assert_eq!(
        r["skipped_ended"][0]["pr_status"]["prs"][0]["merged_at"],
        Value::Null
    );
}

#[test]
fn an_unsettled_lookup_keeps_the_conflict_as_uncertain() {
    for (label, effects) in [
        ("server error", vec![pr_create("14", "confirmed")]),
        ("no state field", vec![pr_create("15", "confirmed")]),
        ("not found", vec![pr_create("99", "confirmed")]),
        (
            "non-numeric response id",
            vec![pr_create("abc", "confirmed")],
        ),
        ("empty response id", vec![pr_create("", "confirmed")]),
        // `+12` parses as 12 (merged) — only the digits-only rule keeps it unsettled.
        ("signed response id", vec![pr_create("+12", "confirmed")]),
        ("no pr_create at all", vec![]),
        // #2329: the first PR merged, a second one is still only intended.
        (
            "second PR intended",
            vec![pr_create("12", "confirmed"), pr_create("13", "intended")],
        ),
        (
            "second PR uncertain",
            vec![pr_create("12", "confirmed"), pr_create("13", "uncertain")],
        ),
    ] {
        let (r, _) = decide(record(KEY_B, 2, "pr_ready", effects));
        assert_eq!(r["decision"], "blocked", "{label}");
        assert_eq!(
            r["conflicts"][0]["pr_status"]["pr_state"], "uncertain",
            "{label}"
        );
        assert_eq!(r["skipped_ended"], json!([]), "{label}");
    }
}

#[test]
fn every_pr_of_the_claim_must_be_closed() {
    let (r, calls) = decide(record(
        KEY_B,
        2,
        "pr_ready",
        vec![pr_create("12", "confirmed"), pr_create("11", "confirmed")],
    ));
    assert_eq!(r["decision"], "blocked", "one of two PRs still open");
    assert_eq!(calls, 2);
    let (r, _) = decide(record(
        KEY_B,
        2,
        "pr_ready",
        vec![pr_create("12", "confirmed"), pr_create("13", "confirmed")],
    ));
    assert_eq!(r["decision"], "clear", "both closed");
}

#[test]
fn only_overlapping_pr_ready_claims_are_looked_up() {
    // A running claim with a merged PR is not handed off: no lookup, the conflict stays.
    let (r, calls) = decide(record(
        KEY_B,
        2,
        "running",
        vec![pr_create("12", "confirmed")],
    ));
    assert_eq!(r["decision"], "blocked");
    assert_eq!(calls, 0);
    assert!(r["conflicts"][0].get("pr_status").is_none());
    // A pr_ready claim that does not overlap: no lookup.
    let mut far = record(KEY_B, 2, "pr_ready", vec![pr_create("12", "confirmed")]);
    far["write_scope"] = json!(["docs/elsewhere.md"]);
    let (r, calls) = decide(far);
    assert_eq!(r["decision"], "clear");
    assert_eq!(calls, 0);
}

/// Audit B1 (6774380195): a GitHub Enterprise claim must not be looked up on github.com, where
/// the same owner/repo/number could be an unrelated closed PR.
#[test]
fn a_github_claim_on_another_host_is_uncertain_without_a_lookup() {
    let mut other = record(KEY_B, 2, "pr_ready", vec![pr_create("12", "confirmed")]);
    other["task_ref"]["host"] = json!("ghe.example.com");
    let (r, calls) = decide(other);
    assert_eq!(r["decision"], "blocked");
    assert_eq!(r["conflicts"][0]["pr_status"]["pr_state"], "uncertain");
    assert_eq!(r["skipped_ended"], json!([]));
    assert_eq!(calls, 0);
}

/// With no overlapping handed-off claim the forge is not touched at all, not even to build the
/// transport: a broken `MOAI_FORGE_GH_BIN` cannot fail the check.
#[test]
fn no_handed_off_overlap_needs_no_forge_setup() {
    let (_t, root) = repo();
    seed(&root, record(KEY_B, 2, "running", vec![]));
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(["scope-check"])
        .arg(&root)
        .args(["--task", KEY_A])
        .current_dir(&root)
        .env("HOME", root.join("home"))
        .env("MOAI_FORGE_GH_BIN", "relative/not-a-file")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap();
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["ok"], true, "{env}");
    assert_eq!(env["result"]["decision"], "blocked");
}

#[test]
fn a_non_github_claim_is_uncertain_without_a_lookup() {
    let mut other = record(KEY_B, 2, "pr_ready", vec![pr_create("12", "confirmed")]);
    other["task_ref"]["provider"] = json!("gitlab");
    other["task_ref"]["host"] = json!("gitlab.com");
    let (r, calls) = decide(other);
    assert_eq!(r["decision"], "blocked");
    assert_eq!(r["conflicts"][0]["pr_status"]["pr_state"], "uncertain");
    assert_eq!(calls, 0);
}
