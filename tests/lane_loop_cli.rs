//! `lane-loop-line` prints the Claude lane's wake line and refuses any other host;
//! `lane-resume-note` writes a card's resume note and `lane-resume-check` reads it back. The
//! binary runs in a clean environment with a temporary project; nothing here reads GitHub or
//! runs git.
mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::Command;

const HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

fn run(args: &[&str], cwd: &Path) -> (i32, Value) {
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", cwd)
        .current_dir(cwd)
        .args(args)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("not JSON: {}", String::from_utf8_lossy(&out.stdout)));
    (out.status.code().unwrap_or(-1), v)
}

/// AUDIT #2917 I1: a saved body of exactly 1 MiB is read; one byte more is refused whole, so a
/// second note past the cap is never missed by a partial read.
#[test]
fn a_body_over_one_mib_is_refused_not_cut() {
    let dir = tempfile::tempdir().unwrap();
    let note = format!(
        "moai-resume: lane=a issue=1 claim=1234567 lease_until=2026-11-14T14:00:00Z branch=b head={HEAD} uncommitted=0\nnext: x\n"
    );
    let limit = moai_cli::lane_loop::READ_LIMIT as usize;
    let file = dir.path().join("c.md");
    let mut at = note.clone();
    at.push_str(&" ".repeat(limit - note.len()));
    fs::write(&file, &at).unwrap();
    let args = ["lane-resume-check", "--file", file.to_str().unwrap()];
    let (code, v) = run(&args, dir.path());
    assert_eq!((code, &v["result"]["ok"]), (0, &json!(true)), "{v}");
    fs::write(&file, format!("{at} ")).unwrap();
    let (code, v) = run(&args, dir.path());
    assert_ne!(code, 0, "one byte over: {v}");
    // More, with a second note after the cap: refused, not read as one note.
    let mut over = at.clone();
    over.push('\n');
    over.push_str(&note);
    fs::write(&file, &over).unwrap();
    let (code, v) = run(&args, dir.path());
    assert_ne!(code, 0, "{v}");
    assert!(v.to_string().contains("over 1048576 bytes"), "{v}");
}

/// AC-1: the Claude line, byte for byte.
#[test]
fn host_lines_match_goldens() {
    let dir = tempfile::tempdir().unwrap();
    let (code, claude) = run(
        &["lane-loop-line", "--host", "claude", "--lane", "agent-1"],
        dir.path(),
    );
    assert_eq!(code, 0, "{claude}");
    assert_eq!(claude["operation"], json!("lane.loop_line"));
    assert_eq!(
        claude["result"]["lines"],
        json!([{"kind": "prompt", "text": "/loop 5m /moai --mode-agent 레인 agent-1 — 순환 계속, self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기"}])
    );
    assert_eq!(claude["result"]["writes"], json!(false));
    // A bad host is a refusal, not a line.
    let (code, bad) = run(
        &["lane-loop-line", "--host", "cursor", "--lane", "agent-1"],
        dir.path(),
    );
    assert_ne!(code, 0, "{bad}");
}

/// SPEC lane-goal-leader-codex-removed AC-1: the WHOLE answer of `lane-loop-line --host claude`
/// is pinned with literal strings (not the source constants), for the default duty and for a
/// custom `--duty`, so removing the Codex and OMP parts cannot change what a running Claude lane
/// reads: the envelope and the result with their keys in order, every value, the one prompt line
/// and the note.
#[test]
fn the_claude_answer_is_pinned_whole() {
    let dir = tempfile::tempdir().unwrap();
    let note = "/loop 5m fires only while the session is idle and expires after seven days";
    let cases = [
        (
            vec!["lane-loop-line", "--host", "claude", "--lane", "agent-1"],
            "agent-1",
            "/loop 5m /moai --mode-agent 레인 agent-1 — 순환 계속, self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기",
        ),
        (
            vec![
                "lane-loop-line",
                "--host",
                "claude",
                "--lane",
                "agent-2",
                "--duty",
                "맡은 카드 마무리, 그 뒤 self-pick",
            ],
            "agent-2",
            "/loop 5m /moai --mode-agent 레인 agent-2 — 맡은 카드 마무리, 그 뒤 self-pick",
        ),
    ];
    for (args, lane, prompt) in cases {
        let (code, v) = run(&args, dir.path());
        assert_eq!(code, 0, "{v}");
        let envelope: Vec<&str> = v.as_object().unwrap().keys().map(String::as_str).collect();
        assert_eq!(
            envelope,
            ["protocol_version", "operation", "ok", "result"],
            "{v}"
        );
        assert_eq!(v["protocol_version"], json!(1), "{v}");
        assert_eq!(v["operation"], json!("lane.loop_line"), "{v}");
        assert_eq!(v["ok"], json!(true), "{v}");
        let result = v["result"].as_object().unwrap();
        let keys: Vec<&str> = result.keys().map(String::as_str).collect();
        assert_eq!(
            keys,
            ["host", "lane", "writes", "lines", "verified", "note"],
            "{v}"
        );
        assert_eq!(result["host"], json!("claude"), "{v}");
        assert_eq!(result["lane"], json!(lane), "{v}");
        assert_eq!(result["writes"], json!(false), "{v}");
        assert_eq!(
            result["lines"],
            json!([{"kind": "prompt", "text": prompt}]),
            "{v}"
        );
        assert_eq!(result["verified"], json!("documented"), "{v}");
        assert_eq!(result["note"], json!(note), "{v}");
        // The whole result at once: no key may be added or dropped.
        assert_eq!(
            v["result"],
            json!({
                "host": "claude",
                "lane": lane,
                "writes": false,
                "lines": [{"kind": "prompt", "text": prompt}],
                "verified": "documented",
                "note": note,
            }),
            "{v}"
        );
    }
}

/// AC-2: the removed hosts are refused with one line that names claude, and `--host claude`
/// still prints its line. The line is the error message of the envelope and the one line on
/// standard error, and nothing else is printed on the error path.
#[test]
fn a_retired_host_is_refused_with_one_line() {
    let dir = tempfile::tempdir().unwrap();
    for host in ["codex", "omp"] {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", dir.path())
            .current_dir(dir.path())
            .args(["lane-loop-line", "--host", host, "--lane", "agent-1"])
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout);
        let v: Value = serde_json::from_str(&text).unwrap_or_else(|_| panic!("not JSON: {text}"));
        assert_ne!(out.status.code(), Some(0), "{host}: {v}");
        assert_eq!(v["ok"], json!(false), "{host}: {v}");
        assert!(v.get("result").is_none(), "{host}: {v}");
        let message = v["error"]["message"].as_str().unwrap();
        assert_eq!(message.lines().count(), 1, "{host}: {message:?}");
        assert!(message.contains("claude"), "{host}: {message}");
        assert!(message.contains(host), "{host}: {message}");
        // One envelope line on standard output, and the same message as the one line on
        // standard error: nothing else is printed.
        assert_eq!(text.trim_end().lines().count(), 1, "{host}: {text}");
        let stderr = String::from_utf8_lossy(&out.stderr);
        assert_eq!(stderr.trim_end(), message, "{host}: {stderr:?}");
    }
    let (code, ok) = run(
        &["lane-loop-line", "--host", "claude", "--lane", "agent-1"],
        dir.path(),
    );
    assert_eq!(code, 0, "{ok}");
    assert_eq!(ok["result"]["host"], json!("claude"), "{ok}");
}

/// AC-3: a note written by `lane-resume-note` is read back by `lane-resume-check`; a note with
/// uncommitted work is refused.
#[test]
fn resume_note_render_then_check() {
    let dir = tempfile::tempdir().unwrap();
    let (code, v) = run(
        &[
            "lane-resume-note",
            "--lane",
            "agent-1",
            "--issue",
            "2904",
            "--claim",
            "6874514212",
            "--lease-until",
            "2026-11-14T14:00:00Z",
            "--branch",
            "issue/2904-x",
            "--head",
            HEAD,
            "--uncommitted",
            "0",
            "--next",
            "run the tests",
            "--next",
            "open the PR",
            "--wait",
            "PR #2910 merge",
            "--clears",
            "leader",
        ],
        dir.path(),
    );
    assert_eq!(code, 0, "{v}");
    let text = v["result"]["text"].as_str().unwrap();
    assert_eq!(
        text,
        format!(
            "moai-resume: lane=agent-1 issue=2904 claim=6874514212 lease_until=2026-11-14T14:00:00Z branch=issue/2904-x head={HEAD} uncommitted=0\nnext: run the tests\nnext: open the PR\nwait: PR #2910 merge clears=leader\n"
        )
    );
    let file = dir.path().join("comment.md");
    fs::write(&file, format!("상태: 대기\n\n{text}\n🗿 MoAI\n")).unwrap();
    let (code, v) = run(
        &["lane-resume-check", "--file", file.to_str().unwrap()],
        dir.path(),
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["ok"], json!(true));
    assert_eq!(v["result"]["note"]["head"], json!(HEAD));
    assert_eq!(
        v["result"]["note"]["next"],
        json!(["run the tests", "open the PR"])
    );
    assert_eq!(
        v["result"]["note"]["wait"],
        json!({"condition": "PR #2910 merge", "clears": "leader"})
    );
    // Uncommitted work is refused when writing and when reading.
    let (code, _) = run(
        &[
            "lane-resume-note",
            "--lane",
            "agent-1",
            "--issue",
            "2904",
            "--claim",
            "6874514212",
            "--lease-until",
            "2026-11-14T14:00:00Z",
            "--branch",
            "b",
            "--head",
            HEAD,
            "--uncommitted",
            "2",
            "--next",
            "x",
        ],
        dir.path(),
    );
    assert_ne!(code, 0);
    fs::write(&file, text.replacen("uncommitted=0", "uncommitted=2", 1)).unwrap();
    let (_, v) = run(
        &["lane-resume-check", "--file", file.to_str().unwrap()],
        dir.path(),
    );
    assert_eq!(v["result"]["ok"], json!(false), "{v}");
    assert!(v["result"]["problems"][0]
        .as_str()
        .unwrap()
        .contains("uncommitted"));
}
