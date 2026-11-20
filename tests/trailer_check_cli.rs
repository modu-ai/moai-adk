//! SPEC-20261117-session-trailer (SDD-6, user decision U-B 2026-11-17): `moai trailers` prints
//! the two identity trailers for a commit message (`Moai-Spec:` and, when the host gives this
//! session's id, `Moai-Session:`), and `moai trailer-check <range>` reports the commits of a
//! range whose trailers are missing or malformed. Temporary repositories and HOMEs only.

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::path::{Path, PathBuf};
use std::process::Command;

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

const SPEC: &str = "SPEC-20261117-session-trailer";
const SESSION: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f";

fn moai(dir: &Path, home: &Path, args: &[&str], env: &[(&str, &str)]) -> (Option<i32>, Value) {
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.args(args)
        .clean_env()
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .current_dir(dir);
    for (k, v) in env {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code(), v)
}

fn commit(dir: &Path, message: &str) {
    std::fs::write(dir.join("f.txt"), message).unwrap();
    common::git::git_ok(dir, &["add", "f.txt"]);
    common::git::git_ok(
        dir,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "commit",
            "-q",
            "-m",
            message,
        ],
    );
}

/// A repository with one commit at `dir`, for the worktree key.
fn repo(dir: &Path) {
    common::git::git_ok(dir, &["init", "-q", "-b", "spec/20261117-session-trailer"]);
    commit(dir, "base");
}

/// The `Moai-Session` value `moai trailers` prints in `dir` for this session id.
fn session_value(dir: &Path, home: &Path) -> String {
    let (code, v) = moai(
        dir,
        home,
        &["trailers", "--spec", SPEC],
        &[("CLAUDE_CODE_SESSION_ID", SESSION)],
    );
    assert_eq!(code, Some(0), "{v}");
    v["result"]["session"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// Re-audit B3 (leader decision): a subagent inherits its parent's session id, so the value
/// adds a key of the worktree: two worktrees of one repository with the same session id get
/// different values, and one worktree gets the same value every time.
#[test]
fn the_session_value_names_the_worktree() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let main = tmp.path().join("main");
    std::fs::create_dir_all(&main).unwrap();
    repo(&main);
    let other = tmp.path().join("other");
    common::git::git_ok(
        &main,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "spec/20261117-other",
            other.to_str().unwrap(),
        ],
    );
    let a = session_value(&main, home.path());
    let b = session_value(&other, home.path());
    assert!(a.starts_with(&format!("{SESSION}:")), "{a}");
    assert!(b.starts_with(&format!("{SESSION}:")), "{b}");
    assert_eq!(a.len(), SESSION.len() + 1 + 12, "{a}");
    assert_ne!(a, b, "two worktrees, one session id");
    assert_eq!(
        a,
        session_value(&main, home.path()),
        "one worktree, one value"
    );
    // A subfolder of a worktree is the same worktree.
    std::fs::create_dir_all(main.join("sub")).unwrap();
    assert_eq!(a, session_value(&main.join("sub"), home.path()));
}

#[test]
fn trailers_name_the_spec_and_this_sessions_id() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    repo(tmp.path());
    let (code, v) = moai(
        tmp.path(),
        home.path(),
        &["trailers", "--spec", SPEC],
        &[("CLAUDE_CODE_SESSION_ID", SESSION)],
    );
    assert_eq!(code, Some(0), "{v}");
    let r = &v["result"];
    let value = r["session"].as_str().unwrap().to_string();
    assert_eq!(
        r["text"],
        json!(format!("Moai-Spec: {SPEC}\nMoai-Session: {value}\n"))
    );
    assert_eq!(r["session_source"], json!("env:CLAUDE_CODE_SESSION_ID"));
    // Outside a git worktree there is no worktree key: the SPEC line only, with the reason.
    let bare = tempfile::tempdir().unwrap();
    let (code, v) = moai(
        bare.path(),
        home.path(),
        &["trailers", "--spec", SPEC],
        &[("CLAUDE_CODE_SESSION_ID", SESSION)],
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["text"], json!(format!("Moai-Spec: {SPEC}\n")));
    assert_eq!(v["result"]["session_missing"], json!(true));
    // No id from the host: only the SPEC line, and the reason is said.
    let (code, v) = moai(tmp.path(), home.path(), &["trailers", "--spec", SPEC], &[]);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["text"], json!(format!("Moai-Spec: {SPEC}\n")));
    assert_eq!(v["result"]["session_missing"], json!(true));
    assert!(v["result"]["reason"]
        .as_str()
        .unwrap()
        .contains("CLAUDE_CODE_SESSION_ID"));
    // A malformed host value is not written as a trailer.
    let (_, v) = moai(
        tmp.path(),
        home.path(),
        &["trailers", "--spec", SPEC],
        &[("CLAUDE_CODE_SESSION_ID", "agent-7 x")],
    );
    assert_eq!(v["result"]["session_missing"], json!(true), "{v}");
    // Only a new-form SPEC id is written; a card number never is.
    for bad in [
        "C-87",
        "SPEC-AUTH-001",
        "SPEC-20261317-x",
        "SPEC-20261117-Bad",
    ] {
        let (code, v) = moai(tmp.path(), home.path(), &["trailers", "--spec", bad], &[]);
        assert_eq!(code, Some(1), "{bad}: {v}");
    }
}

#[test]
fn trailer_check_reports_missing_and_malformed_trailers() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let dir = tmp.path();
    common::git::git_ok(dir, &["init", "-q", "-b", "spec/20261117-session-trailer"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(
        dir,
        &format!("good\n\nMoai-Spec: {SPEC}\nMoai-Session: {SESSION}:0123456789ab"),
    );
    let range = format!("{base}..HEAD");
    let (code, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["commits"], json!(1));

    commit(dir, "no trailers at all");
    commit(dir, &format!("spec only\n\nMoai-Spec: {SPEC}"));
    commit(dir, "card number\n\nMoai-Spec: C-87\nMoai-Session: agent-7");
    // Audit B1: a failing check exits 1 and still prints its result.
    let (code, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(
        v["ok"],
        json!(true),
        "the envelope still carries the result: {v}"
    );
    let r = &v["result"];
    assert_eq!(r["ok"], json!(false), "{v}");
    assert_eq!(r["commits"], json!(4));
    let kinds: Vec<&str> = r["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["kind"].as_str().unwrap())
        .collect();
    assert_eq!(
        kinds,
        [
            "missing_spec",
            "missing_session",
            "missing_session",
            "malformed_spec",
            "malformed_session"
        ],
        "{v}"
    );
    // A missing session is reported but does not fail the check; a missing SPEC fails it
    // only on a `spec/` branch.
    assert_eq!(r["failing"], json!(3), "{v}");
    common::git::git_ok(dir, &["switch", "-q", "-c", "main"]);
    let (code, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(v["result"]["failing"], json!(2), "{v}");

    // Re-audit N4: a detached checkout (CI) names the branch it judges; the name decides
    // whether a missing Moai-Spec fails, not the checkout's own HEAD.
    common::git::git_ok(dir, &["switch", "-q", "--detach"]);
    let (_, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert_eq!(v["result"]["branch"], json!("HEAD"), "{v}");
    assert_eq!(v["result"]["failing"], json!(2), "{v}");
    let (code, v) = moai(
        dir,
        home.path(),
        &[
            "trailer-check",
            &range,
            "--branch",
            "spec/20261117-session-trailer",
        ],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(
        v["result"]["branch"],
        json!("spec/20261117-session-trailer")
    );
    assert_eq!(v["result"]["failing"], json!(3), "{v}");

    // Re-audit B3 transition: a bare id from before the worktree key is reported, not failed;
    // a value whose key is not 12 hex digits is malformed.
    let next = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(
        dir,
        &format!("legacy\n\nMoai-Spec: {SPEC}\nMoai-Session: {SESSION}"),
    );
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &format!("{next}..HEAD")],
        &[],
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(
        v["result"]["findings"][0]["kind"],
        json!("legacy_session"),
        "{v}"
    );
    let next = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(
        dir,
        &format!("short key\n\nMoai-Spec: {SPEC}\nMoai-Session: {SESSION}:0123"),
    );
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &format!("{next}..HEAD")],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(
        v["result"]["findings"][0]["kind"],
        json!("malformed_session"),
        "{v}"
    );

    // A range that reads as an option is refused before git runs.
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", "--", "--output=x"],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    assert!(!dir.join("x").exists());
}

/// Verify wave: card numbers in a commit's subject or body fail after a recorded start, and
/// only then. The branch is `main` so missing-trailer findings never fail and the exit code
/// says only what the card-number rule decided.
#[test]
fn card_numbers_fail_after_a_recorded_start() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let dir = tmp.path();
    common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(dir, "wording after C-150");
    commit(dir, "a clean follow-up");
    let range = format!("{base}..HEAD");

    // Without the recorded start the rule is off and the same history passes.
    let (code, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["card_numbers_from"], json!(null), "{v}");

    // With the base named as the start, the commit after it fails on its shape.
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &range, "--card-numbers-since", &base],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(v["result"]["ok"], json!(false), "{v}");
    let findings: Vec<&Value> = v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == json!("card_number"))
        .collect();
    assert_eq!(findings.len(), 1, "{v}");
    assert_eq!(findings[0]["fails"], json!(true), "{v}");
    assert!(
        findings[0]["detail"].as_str().unwrap().contains("C-150"),
        "{v}"
    );

    // The commit the start names is itself exempt: everything after it is clean.
    let named = common::git::git(dir, &["rev-parse", "HEAD~"]);
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &range, "--card-numbers-since", &named],
        &[],
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
    assert_eq!(v["result"]["card_numbers_from"], json!(1), "{v}");
}

/// The shapes that are not card numbers pass: a SPEC id, a version, a hexadecimal SHA, the
/// `AC-<n>` neighbour, and a card number inside quoted ledger output.
#[test]
fn spec_ids_versions_shas_and_quoted_output_pass() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let dir = tmp.path();
    common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(
        dir,
        "docs: reword per SPEC-20261120-commit-message-card-numbers\n\n\
         built v4.2647.9 from 3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac\n\
         covers AC-150 and W2-7\n\
         the take refused with \"reserved for card C-20\", recorded verbatim",
    );
    let range = format!("{base}..HEAD");
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &range, "--card-numbers-since", &base],
        &[],
    );
    assert_eq!(code, Some(0), "{v}");
    let cards: Vec<&Value> = v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == json!("card_number"))
        .collect();
    assert!(cards.is_empty(), "{v}");
}

/// The same scan judges one draft message file, always on: the commit-msg hook's mode.
#[test]
fn message_file_judges_a_draft() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let message = tmp.path().join("COMMIT_EDITMSG");
    std::fs::write(&message, "fix: wording after C-9\n").unwrap();
    let (code, v) = moai(
        tmp.path(),
        home.path(),
        &["trailer-check", "--message-file", message.to_str().unwrap()],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(v["result"]["ok"], json!(false), "{v}");
    std::fs::write(
        &message,
        "docs: reword per SPEC-20261120-commit-message-card-numbers\n",
    )
    .unwrap();
    let (code, v) = moai(
        tmp.path(),
        home.path(),
        &["trailer-check", "--message-file", message.to_str().unwrap()],
        &[],
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
}

/// The provided commit-msg hook calls the same check on the file git hands it.
#[test]
fn the_commit_msg_hook_calls_the_message_file_check() {
    let hook = std::fs::read_to_string(root().join("install/commit-msg")).expect("hook file");
    assert!(hook.contains("trailer-check"), "{hook}");
    assert!(hook.contains("--message-file"), "{hook}");
    assert!(hook.contains("\"$1\""), "{hook}");
}
