//! #2902 O3a: `lane-duty-line` prints the one line a leader types into a lane. The wording is
//! pinned here as a golden; the line is typed into another terminal, so a `--first` task with a
//! control or line-breaking character, or over 300 characters, is refused. Reads nothing and
//! writes nothing; the binary runs with an empty temporary HOME.
mod common;

use serde_json::Value;
use std::process::{Command, Output};

const BIN: &str = crate::common::MOAI_BIN;

/// The user's fixed wording (#2902 plan 6874299470); canonical source is the O1 #2900 document.
const BODY: &str = "Issue/PR 을 배차받거나 self-pick 해서 열린 Issue·PR 이 모두 0 이 될 때까지 무중단 운영. PR 을 연 뒤 감사·CI 를 기다리지 말고 바로 다음 카드(Issue 또는 감사자 없는 PR)를 맡는다. 열린 PR 상한 8. 매 카드 끝에 LANE-REPORT 를 남기고 곧바로 다음 self-pick.";

fn run(args: &[&str]) -> Output {
    let home = tempfile::tempdir().unwrap();
    Command::new(BIN)
        .arg("lane-duty-line")
        .args(args)
        .current_dir(home.path())
        .env("HOME", home.path())
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "envelope on stdout; stderr {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn line(args: &[&str]) -> String {
    let out = run(args);
    let v = envelope(&out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "lane.duty_line");
    v["result"]["line"].as_str().unwrap().to_string()
}

/// A refused call: exit 1, `invalid_input`, and no line.
fn refused(args: &[&str]) -> String {
    let out = run(args);
    let v = envelope(&out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], false, "{v}");
    assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    assert!(v.get("result").is_none(), "{v}");
    v["error"]["message"].as_str().unwrap().to_string()
}

#[test]
fn the_line_is_the_fixed_wording_for_each_host() {
    let slash = format!("/moai --mode-agent 레인 3 — {BODY}");
    assert_eq!(line(&["--lane", "3"]), slash, "claude is the default host");
    assert_eq!(line(&["--lane", "3", "--host", "claude"]), slash);
    assert_eq!(line(&["--lane", "3", "--host", "omp"]), slash);
    assert_eq!(
        line(&["--lane", "3", "--host", "codex"]),
        format!("$moai --mode-agent 레인 3 — {BODY}")
    );
}

#[test]
fn first_task_is_appended_after_the_body() {
    assert_eq!(
        line(&["--lane", "agent-1", "--first", "#2902 O3a 구현"]),
        format!("/moai --mode-agent 레인 agent-1 — {BODY} 먼저: #2902 O3a 구현")
    );
    assert_eq!(
        line(&["--lane", "3", "--host", "codex", "--first", "PR #2824 감사"]),
        format!("$moai --mode-agent 레인 3 — {BODY} 먼저: PR #2824 감사")
    );
}

#[test]
fn an_unknown_host_is_a_usage_error() {
    let out = run(&["--lane", "3", "--host", "cursor"]);
    assert_eq!(out.status.code(), Some(2));
    assert_eq!(envelope(&out)["error"]["kind"], "usage");
}

/// The name rule is `lane_registry::valid_name`: 64 characters pass, 65 are refused.
#[test]
fn the_lane_name_is_bounded_and_file_safe() {
    let sixty_four = "a".repeat(64);
    assert!(line(&["--lane", &sixty_four]).contains(&format!("레인 {sixty_four} — ")));
    let sixty_five = "a".repeat(65);
    let message = refused(&["--lane", &sixty_five]);
    assert!(!message.contains(&sixty_five), "{message}");
    for bad in ["", "-x", "a b", "a/b", "레인", "a\nb"] {
        refused(&[&format!("--lane={bad}")]);
    }
}

/// Every character that could move the cursor, end the line or reorder the text in the
/// receiving terminal is refused, and the refusal does not echo the task.
#[test]
fn first_with_a_control_character_is_refused() {
    for bad in [
        "a\tb",
        "a\nb",
        "a\rb",
        "a\u{1b}[2Jb",
        "a\u{1}b",
        "a\u{7f}b",
        "a\u{85}b",
        "a\u{9b}b",
        "a\u{202e}b",
        "a\u{2066}b",
        "a\u{200f}b",
        "a\u{feff}b",
    ] {
        let message = refused(&["--lane", "3", "--first", bad]);
        assert!(!message.contains(bad), "{message}");
    }
}

/// #2902 leader decision 6874308433: the Unicode line and paragraph separators are refused on
/// the same path; their printable neighbour U+2027 passes.
#[test]
fn first_with_a_line_or_paragraph_separator_is_refused() {
    assert!(line(&["--lane", "3", "--first", "a\u{2027}b"]).ends_with("먼저: a\u{2027}b"));
    for bad in ["a\u{2028}b", "a\u{2029}b", "\u{2028}", "\u{2029}"] {
        let message = refused(&["--lane", "3", "--first", bad]);
        assert!(message.contains("--first"), "{message}");
    }
}

/// 300 characters pass (counted as characters: 300 Hangul syllables are 900 bytes), 301 are
/// refused, and an empty task is refused.
#[test]
fn first_is_at_most_three_hundred_characters() {
    for ch in ["a", "가"] {
        let three_hundred = ch.repeat(300);
        assert!(line(&["--lane", "3", "--first", &three_hundred])
            .ends_with(&format!("먼저: {three_hundred}")));
        refused(&["--lane", "3", "--first", &ch.repeat(301)]);
    }
    refused(&["--lane", "3", "--first", ""]);
}
