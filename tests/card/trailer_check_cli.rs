//! SPEC-20261117-session-trailer (SDD-6, user decision U-B 2026-11-17): `moai trailers` prints
//! the two identity trailers for a commit message (`Moai-Spec:` and, when the host gives this
//! session's id, `Moai-Session:`), and `moai trailer-check <range>` reports the commits of a
//! range whose trailers are missing or malformed. Temporary repositories and HOMEs only.

use crate::common;
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
    // The card number sits in the BODY: the body arm of the scan is pinned here, the subject
    // arm in `card_numbers_in_a_subject_fail_after_a_recorded_start` (the message-file mode
    // is a separate function with its own test).
    commit(dir, "wording\n\nwording after C-150");
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

/// The `card_number` findings of one `trailer-check` answer.
fn card_findings(v: &Value) -> Vec<&Value> {
    v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == json!("card_number"))
        .collect()
}

/// The card shapes the `card_number` findings of one answer name, in order (the detail reads
/// `... names card <shape> on line <n>; ...`).
fn card_shapes(v: &Value) -> Vec<String> {
    card_findings(v)
        .iter()
        .map(|f| {
            let detail = f["detail"].as_str().unwrap();
            let after = detail.split("names card ").nth(1).unwrap();
            after.split(' ').next().unwrap().to_string()
        })
        .collect()
}

/// The SUBJECT arm of the commit-log scan (the body arm and the message-file mode have their own
/// tests): a card number in a subject fails on line 1 wherever it stands (after a space, right
/// after an opening bracket, at the very start of the line, right after a colon), in both letter
/// cases and after a closed pair of quotes, only after the recorded start, and the commit the
/// start names is exempt.
#[test]
fn card_numbers_in_a_subject_fail_after_a_recorded_start() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let dir = tmp.path();
    common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(dir, "fix: wording after C-150");
    commit(dir, "fix: lower case c-7 counts too");
    commit(dir, "fix: \"closed\" quotes before it leave C-33 bare");
    commit(dir, "fix: the usual tail shape (C-61)");
    commit(dir, "C-62 leads this subject");
    commit(dir, "docs:C-63 follows the colon directly");
    commit(dir, "a clean follow-up");
    let range = format!("{base}..HEAD");

    // Without the recorded start the rule is off.
    let (code, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert_eq!(code, Some(0), "{v}");
    assert!(card_findings(&v).is_empty(), "{v}");

    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &range, "--card-numbers-since", &base],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    let all = ["C-150", "c-7", "C-33", "C-61", "C-62", "C-63"];
    assert_eq!(card_shapes(&v), all, "{v}");
    for f in card_findings(&v) {
        assert_eq!(f["fails"], json!(true), "{v}");
        assert!(f["detail"].as_str().unwrap().contains("on line 1"), "{v}");
    }

    // The commit the start names (the oldest one) is exempt: its shape is not reported, the
    // others still are, and the exit code stays 1.
    let named = common::git::git(dir, &["rev-parse", "HEAD~6"]);
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &range, "--card-numbers-since", &named],
        &[],
    );
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(card_shapes(&v), all[1..], "{v}");
}

/// The shapes that are not card numbers pass in a SUBJECT too: a SPEC id, a version, a
/// hexadecimal SHA, the `AC-<n>` and `SPEC-C-<n>` neighbours, a trailing letter, a word or hex
/// run that holds a `c-<digits>` candidate between letters, and a card number inside quoted
/// ledger output. The SPEC id, version, SHA and `W2-7` lines hold no `C-<digits>` candidate at
/// all, so they never reach the boundary code; the lines after them do.
#[test]
fn allowed_shapes_in_a_subject_pass() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let dir = tmp.path();
    common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    for subject in [
        "docs: reword per SPEC-20261120-commit-message-card-numbers",
        "build: v4.2647.9 from 3c3446a3dbd7d918a7d6465c0765ae66b2dafe3231b9be11cd3181737b6658ac",
        "docs: covers AC-150 and W2-7",
        "docs: the neighbour SPEC-C-9 and the shape C-150b are no card numbers",
        "docs: the word abc-12 and the hex run deadc-1beef carry a c- candidate inside",
        "fix: the take refused with \"reserved for card C-20\", recorded verbatim",
    ] {
        commit(dir, subject);
    }
    let range = format!("{base}..HEAD");
    let (code, v) = moai(
        dir,
        home.path(),
        &["trailer-check", &range, "--card-numbers-since", &base],
        &[],
    );
    assert_eq!(code, Some(0), "{v}");
    assert!(card_findings(&v).is_empty(), "{v}");
}

/// The host's attribution line, and a line a launcher puts into the trailer paragraph.
const ATTRIBUTION: &str = "Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>";
const FOREIGN: &str = "\u{1f5ff} MoAI";
const KEYED_SESSION: &str = "abc12345:123456789012";

/// The `unread_trailer_line` findings of one `trailer-check` answer.
fn unread_findings(v: &Value) -> Vec<&Value> {
    v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == json!("unread_trailer_line"))
        .collect()
}

/// The 1-based line of `needle` in `message`, counted as a person reads the message (the subject
/// is line 1).
fn line_of(message: &str, needle: &str) -> usize {
    message.lines().position(|l| l == needle).unwrap() + 1
}

/// SPEC-20261126-check-affected-trailers-flaky REQ-4, AC-3: a line that begins with a `Moai-*` key
/// in the last paragraph is unread when a foreign line sits in that paragraph (git then reads no
/// trailer from it), and the range fails with a finding that names the key and the line.
#[test]
fn an_attribution_line_in_the_last_paragraph_is_an_unread_trailer_line() {
    let spec = format!("Moai-Spec: {SPEC}");
    let session = format!("Moai-Session: {KEYED_SESSION}");
    let verified = "Moai-Verified: tests";
    // (case, message, the lines expected to be reported unread)
    let cases: Vec<(&str, String, Vec<(&str, String)>)> = vec![
        (
            "a foreign line between the Moai lines",
            format!("fix: demo\n\n{spec}\n{session}\n{FOREIGN}\n{verified}\n{ATTRIBUTION}"),
            vec![
                ("Moai-Spec", spec.clone()),
                ("Moai-Session", session.clone()),
                ("Moai-Verified", verified.to_string()),
            ],
        ),
        (
            "the foreign line as the last line of the block",
            format!("fix: demo\n\n{spec}\n{session}\n{verified}\n{ATTRIBUTION}\n{FOREIGN}"),
            vec![
                ("Moai-Spec", spec.clone()),
                ("Moai-Session", session.clone()),
                ("Moai-Verified", verified.to_string()),
            ],
        ),
        (
            // Moai-Spec stands in a paragraph before the last one (a split block, named by its own
            // finding); only the session line is in the last paragraph and unread.
            "only the session line is in the unread last paragraph",
            format!("fix: demo\n\n{spec}\n\n{session}\n{FOREIGN}"),
            vec![("Moai-Session", session.clone())],
        ),
    ];
    for (case, message, expected) in cases {
        let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        let dir = tmp.path();
        // `main`: a missing SPEC or session never fails there, so only the new finding can.
        common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
        commit(dir, "base");
        let base = common::git::git(dir, &["rev-parse", "HEAD"]);
        commit(dir, &message);
        // The premise: git itself reads no session value from this message.
        let read = common::git::git(
            dir,
            &[
                "log",
                "-1",
                "--format=%(trailers:key=Moai-Session,valueonly)",
            ],
        );
        assert!(read.is_empty(), "{case}: git reads {read:?}");
        let (code, v) = moai(
            dir,
            home.path(),
            &["trailer-check", &format!("{base}..HEAD")],
            &[],
        );
        let found = unread_findings(&v);
        assert_eq!(found.len(), expected.len(), "{case}: {v}");
        for (key, line) in &expected {
            let hit = found
                .iter()
                .find(|f| f["detail"].as_str().unwrap().contains(key))
                .unwrap_or_else(|| panic!("{case}: no finding names {key}: {v}"));
            assert_eq!(hit["fails"], json!(true), "{case}: {v}");
            let detail = hit["detail"].as_str().unwrap();
            let n = line_of(&message, line);
            assert!(
                detail.contains(line.as_str()) || detail.contains(&format!("line {n}")),
                "{case}: the finding names neither the line text nor `line {n}`: {detail}"
            );
        }
        assert_eq!(code, Some(1), "{case}: {v}");
        assert!(v["result"]["failing"].as_u64().unwrap() >= 1, "{case}: {v}");
        assert_eq!(v["result"]["ok"], json!(false), "{case}: {v}");
    }
}

/// A commit of `file` (not `f.txt`, so two branches do not conflict) with this message.
fn commit_file(dir: &Path, file: &str, message: &str) {
    std::fs::write(dir.join(file), message).unwrap();
    common::git::git_ok(dir, &["add", file]);
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

/// SPEC-20261126-check-affected-trailers-flaky REQ-4, AC-3 (the miss direction): a clean block,
/// a block git accepts with a `Signed-off-by` line in it, and a merge commit whose message ends in
/// a `# Conflicts:` comment paragraph after the trailers raise no `unread_trailer_line`.
#[test]
fn a_clean_block_and_a_signed_off_by_block_raise_no_unread_finding() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let dir = tmp.path();
    let block = format!("Moai-Spec: {SPEC}\nMoai-Session: {KEYED_SESSION}\nMoai-Verified: tests");
    common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    // A commit on a side branch, so the merge below has two parents and nothing to conflict on.
    common::git::git_ok(dir, &["switch", "-q", "-c", "side"]);
    commit_file(
        dir,
        "side.txt",
        &format!("side: clean\n\n{block}\n{ATTRIBUTION}"),
    );
    common::git::git_ok(dir, &["switch", "-q", "main"]);
    commit_file(
        dir,
        "clean.txt",
        &format!("fix: clean\n\n{block}\n{ATTRIBUTION}"),
    );
    commit_file(
        dir,
        "signed.txt",
        &format!("fix: signed\n\n{block}\nSigned-off-by: A <a@example.com>\n{ATTRIBUTION}"),
    );
    let merge = format!("merge: side\n\n{block}\n\n# Conflicts:\n#\tf.txt");
    common::git::git_ok(
        dir,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "merge",
            "-q",
            "--no-ff",
            "--cleanup=verbatim",
            "-m",
            &merge,
            "side",
        ],
    );
    // The premise: the merge message really ends in the comment paragraph.
    let body = common::git::git(dir, &["log", "-1", "--format=%B"]);
    assert!(body.contains("# Conflicts:"), "{body}");
    let range = format!("{base}..HEAD");
    let (code, v) = moai(dir, home.path(), &["trailer-check", &range], &[]);
    assert!(unread_findings(&v).is_empty(), "{v}");
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["failing"], json!(0), "{v}");
    // Every commit the range reads is judged (the merge is skipped).
    assert_eq!(v["result"]["commits"], json!(3), "{v}");
}
