//! Issue #3096 S7b (leader decision 6905681774): the launcher hands an effort the operator asked
//! for to Codex as `-c model_reasoning_effort="<value>"`, passes nothing when none was asked
//! (the user's own Codex settings decide, the launcher picks no default), and says so in a note
//! when the word is not one Codex takes. Every case uses `--dry-run` in a temporary git project;
//! no host is started and no Codex setting file is read.
mod common;
use serde_json::Value;
use std::process::Command;

const NOT_PASSED: &str = "--effort was not passed to codex";

/// Runs the launcher with `--dry-run` in a fresh git project and a home that does not exist.
fn dry(args: &[&str]) -> (i32, Value) {
    let project = tempfile::tempdir().unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap();
    assert!(init.success());
    let out = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .arg("--dry-run")
        .current_dir(project.path())
        .env(
            "HOME",
            std::env::temp_dir().join("moai-cli-test-no-such-home"),
        )
        .output()
        .unwrap();
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let v = serde_json::from_str(&stdout)
        .unwrap_or_else(|e| panic!("not json ({e}): {stdout} {:?}", out.stderr));
    (out.status.code().unwrap_or(-1), v)
}

fn argv(v: &Value) -> Vec<String> {
    serde_json::from_value(v["argv"].clone()).unwrap()
}

fn notes(v: &Value) -> String {
    v["notes"].to_string()
}

fn effort_arg(e: &str) -> String {
    format!("model_reasoning_effort=\"{e}\"")
}

#[test]
fn every_effort_codex_takes_goes_out_as_a_config_override() {
    for e in ["low", "medium", "high", "xhigh", "max"] {
        let (code, v) = dry(&["codex", "--effort", e]);
        assert_eq!(code, 0, "{e}");
        assert_eq!(argv(&v), ["codex", "-c", &effort_arg(e)], "{e}");
        assert!(!notes(&v).contains(NOT_PASSED), "{e}: {}", notes(&v));
    }
}

#[test]
fn without_a_request_nothing_about_effort_is_passed_and_no_note_is_added() {
    let (code, v) = dry(&["codex"]);
    assert_eq!(code, 0);
    assert_eq!(argv(&v), ["codex"]);
    assert!(!notes(&v).contains("effort"), "{}", notes(&v));
    // A model alone must not bring an effort with it either: the launcher sets no default.
    let (_, v) = dry(&["codex", "--model", "gpt-6.1-terra"]);
    assert_eq!(argv(&v), ["codex", "-m", "gpt-6.1-terra"]);
}

#[test]
fn a_word_codex_does_not_take_is_dropped_with_a_note() {
    for e in ["ultracode", "ultra", "turbo", "high\"; touch x; \""] {
        let (code, v) = dry(&["codex", "--effort", e]);
        assert_eq!(code, 0, "{e}");
        let a = argv(&v);
        assert!(
            !a.iter()
                .any(|x| x == "-c" || x.contains("model_reasoning_effort")),
            "{e}: {a:?}"
        );
        assert!(notes(&v).contains(NOT_PASSED), "{e}: {}", notes(&v));
    }
}

#[test]
fn the_model_comes_first_and_the_prompt_stays_last() {
    let (_, v) = dry(&[
        "codex",
        "-a",
        "--name",
        "agent-9",
        "--model",
        "gpt-6.1-terra",
        "--effort",
        "xhigh",
    ]);
    assert_eq!(
        argv(&v),
        [
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "-m",
            "gpt-6.1-terra",
            "-c",
            &effort_arg("xhigh"),
            "MoAI session name: agent-9"
        ]
    );
}

#[test]
fn claude_hosts_keep_passing_effort_as_a_flag() {
    let (_, v) = dry(&["cc", "--effort", "high"]);
    assert_eq!(argv(&v), ["claude", "--model", "opus", "--effort", "high"]);
    assert!(!argv(&v).iter().any(|x| x == "-c"));
}
