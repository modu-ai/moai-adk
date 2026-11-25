//! SPEC-20261125-trailer-paragraph-guard: git reads a message's trailers from its last paragraph
//! only, so `Moai-*` lines split from `Co-Authored-By` by a blank line are never read. The
//! commit-msg hook's mode (`trailer-check --message-file`) refuses such a message and names the
//! fix; a range reports the same shape as a finding that never fails. Temporary folders only.

use crate::common;
use common::CleanEnv;

use serde_json::Value;
use std::path::{Path, PathBuf};
use std::process::Command;

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

const SPLIT: &str = "fix(x): subject\n\nBody sentence.\n\nMoai-Spec: SPEC-20261125-trailer-paragraph-guard\nMoai-Session: 0123456789ab\nMoai-Verified: tests\n\nCo-Authored-By: A <a@example.invalid>\n";

/// The same message with the blank line before `Co-Authored-By` removed: one last paragraph.
fn joined() -> String {
    SPLIT.replace("tests\n\nCo-", "tests\nCo-")
}

fn moai(dir: &Path, args: &[&str]) -> (Option<i32>, Value) {
    let mut cmd = Command::new(common::MOAI_BIN);
    cmd.args(args)
        .clean_env()
        .env("HOME", dir)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .current_dir(dir);
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

/// Writes `message` to a file in `dir` and judges it the way the hook does.
fn judge_message(dir: &Path, message: &str) -> (Option<i32>, Value) {
    let file = dir.join("MSG");
    std::fs::write(&file, message).unwrap();
    moai(dir, &["trailer-check", "--message-file", file.to_str().unwrap()])
}

fn kinds(v: &Value) -> Vec<String> {
    v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["kind"].as_str().unwrap().to_string())
        .collect()
}

#[test]
fn ac1_a_split_block_is_refused_with_the_key_the_line_and_the_fix() {
    let tmp = tempfile::tempdir().unwrap();
    let (code, v) = judge_message(tmp.path(), SPLIT);
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(v["result"]["ok"], false, "{v}");
    assert_eq!(kinds(&v), ["split_trailer_block"; 3], "{v}");
    let first = v["result"]["findings"][0]["detail"].as_str().unwrap();
    assert!(first.contains("Moai-Spec on line 5"), "{first}");
    assert!(first.contains("remove the blank line"), "{first}");
    let last = v["result"]["findings"][2]["detail"].as_str().unwrap();
    assert!(last.contains("Moai-Verified on line 7"), "{last}");
}

#[test]
fn ac2_the_installed_hook_refuses_the_split_message_and_passes_the_joined_one() {
    let tmp = tempfile::tempdir().unwrap();
    let hook = root().join("install/commit-msg");
    // The hook finds `moai` on PATH: the folder of the built binary holds it.
    let bin_dir = Path::new(common::MOAI_BIN).parent().unwrap();
    let path = format!("{}:/usr/bin:/bin", bin_dir.display());
    let run = |message: &str| {
        let file = tmp.path().join("MSG");
        std::fs::write(&file, message).unwrap();
        Command::new("/bin/sh")
            .arg(&hook)
            .arg(&file)
            .clean_env()
            .env("HOME", tmp.path())
            .env("PATH", &path)
            .env("LC_ALL", "C")
            .current_dir(tmp.path())
            .output()
            .unwrap()
    };
    let refused = run(SPLIT);
    assert_eq!(refused.status.code(), Some(1), "{refused:?}");
    assert!(
        String::from_utf8_lossy(&refused.stdout).contains("split_trailer_block"),
        "{refused:?}"
    );
    let passed = run(&joined());
    assert_eq!(passed.status.code(), Some(0), "{passed:?}");
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

#[test]
fn ac3_a_range_names_the_split_block_and_it_adds_no_failure() {
    let tmp = tempfile::tempdir().unwrap();
    let dir = tmp.path();
    common::git::git_ok(dir, &["init", "-q", "-b", "main"]);
    commit(dir, "base");
    let base = common::git::git(dir, &["rev-parse", "HEAD"]);
    commit(dir, SPLIT);
    let range = format!("{base}..HEAD");
    // Off a `spec/` branch the unread SPEC does not fail; the named cause does not either.
    let (code, v) = moai(dir, &["trailer-check", &range, "--branch", "main"]);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["failing"], 0, "{v}");
    let split: Vec<&Value> = v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["kind"] == "split_trailer_block")
        .collect();
    assert_eq!(split.len(), 3, "{v}");
    assert!(split.iter().all(|f| f["fails"] == Value::Bool(false)), "{v}");
    // On a `spec/` branch the unread `Moai-Spec` fails exactly as it did before this check.
    let (code, v) = moai(dir, &["trailer-check", &range, "--branch", "spec/x"]);
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(v["result"]["failing"], 1, "{v}");
}

#[test]
fn ac4_whole_blocks_and_quoted_keys_pass() {
    let tmp = tempfile::tempdir().unwrap();
    let alone = "s\n\nbody\n\nMoai-Spec: SPEC-20261125-x\nMoai-Verified: none\n";
    let quoted = "s\n\nThe line Moai-Spec: SPEC-20261125-x is read from the last paragraph only.\n\nMoai-Spec: SPEC-20261125-x\nMoai-Session: 0123456789ab\n";
    let mixed = "s\n\nMoai-Spec: SPEC-20261125-x\nThis line explains the one above.\n\nCo-Authored-By: A <a@example.invalid>\n";
    for (name, message) in [
        ("joined", joined()),
        ("alone", alone.into()),
        ("quoted", quoted.into()),
        ("mixed", mixed.into()),
    ] {
        let (code, v) = judge_message(tmp.path(), &message);
        assert_eq!(code, Some(0), "{name}: {v}");
        assert!(kinds(&v).is_empty(), "{name}: {v}");
    }
}

#[test]
fn ac6_the_reference_says_one_last_paragraph_and_shows_it() {
    let doc =
        std::fs::read_to_string(root().join("skills/moai/references/internal/commit-trailers.md"))
            .unwrap();
    assert!(
        doc.contains("one last paragraph with no blank line between them"),
        "the sentence is missing"
    );
    assert!(
        doc.contains("Moai-Verified: tests\nCo-Authored-By:"),
        "the example with no blank line is missing"
    );
}
