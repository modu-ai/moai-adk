//! #2904 (O5, #2899 B): `lane-loop-line` prints the Claude Code wake line;
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

/// The whole Claude answer, byte for byte: running lanes read this line, so it does not move
/// when other parts of the command go.
#[test]
fn the_claude_answer_is_pinned_whole() {
    let dir = tempfile::tempdir().unwrap();
    for (duty, text) in [
        (
            None,
            "/loop 5m 레인 agent-1 — 순환 계속, self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기",
        ),
        (Some("x"), "/loop 5m 레인 agent-1 — x"),
    ] {
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.clean_env()
            .env("HOME", dir.path())
            .current_dir(dir.path())
            .args(["lane-loop-line", "--host", "claude", "--lane", "agent-1"]);
        if let Some(duty) = duty {
            cmd.args(["--duty", duty]);
        }
        let out = cmd.output().unwrap();
        assert_eq!(out.status.code(), Some(0));
        assert_eq!(
            String::from_utf8(out.stdout).unwrap(),
            format!(
                "{{\"protocol_version\":1,\"operation\":\"lane.loop_line\",\"ok\":true,\"result\":{{\"host\":\"claude\",\"lane\":\"agent-1\",\"writes\":false,\"lines\":[{{\"kind\":\"prompt\",\"text\":\"{text}\"}}],\"verified\":\"documented\",\"note\":\"/loop 5m fires only while the session is idle and expires after seven days\"}}}}\n"
            )
        );
    }
}

/// Claude Code is the only host: another `--host` word, the removed ones included, is a usage
/// error that points at `--help`, no line is printed, and that help lists `claude`
/// alone. The removed `--project` flag went with the hook file it read.
#[test]
fn a_removed_host_is_a_usage_error() {
    let dir = tempfile::tempdir().unwrap();
    for host in ["codex", "omp", "cursor"] {
        let (code, v) = run(
            &[
                "lane-loop-line",
                "--host",
                host,
                "--lane",
                "agent-1",
                "--repo",
                "o/r",
            ],
            dir.path(),
        );
        assert_eq!(code, 2, "{host}: {v}");
        assert_eq!(v["error"]["kind"], json!("usage"), "{host}: {v}");
        assert!(v.get("result").is_none(), "{host}: {v}");
        let message = v["error"]["message"].as_str().unwrap();
        assert!(message.contains("see --help"), "{host}: {message}");
    }
    let help = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", dir.path())
        .current_dir(dir.path())
        .args(["lane-loop-line", "--help"])
        .output()
        .unwrap();
    assert_eq!(help.status.code(), Some(0));
    let text = String::from_utf8_lossy(&help.stdout);
    assert!(text.contains("[possible values: claude]"), "{text}");
    let (code, v) = run(
        &[
            "lane-loop-line",
            "--host",
            "claude",
            "--lane",
            "agent-1",
            "--project",
            ".",
        ],
        dir.path(),
    );
    assert_eq!(code, 2, "{v}");
    // A repository, when given, is still checked.
    let (code, v) = run(
        &[
            "lane-loop-line",
            "--host",
            "claude",
            "--lane",
            "agent-1",
            "--repo",
            "no-slash",
        ],
        dir.path(),
    );
    assert_eq!(code, 1, "{v}");
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
