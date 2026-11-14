//! #2904 (O5, #2899 B): `lane-loop-line` prints each host's wake line and reads the Codex lane's
//! Stop hook from the project; `lane-resume-note` writes a card's resume note and
//! `lane-resume-check` reads it back. The binary runs in a clean environment with a temporary
//! project; nothing here reads GitHub or runs git.
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

/// AC-1: the three hosts' lines, byte for byte.
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
    let (_, codex) = run(
        &[
            "lane-loop-line",
            "--host",
            "codex",
            "--lane",
            "agent-1",
            "--repo",
            "modu-ai/moai-adk",
        ],
        dir.path(),
    );
    assert_eq!(
        codex["result"]["lines"],
        json!([
            {"kind": "shell", "text": "codex features enable goals"},
            {"kind": "prompt", "text": "/goal 저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(HOLD·needs:decision 제외) — 레인 agent-1: 매 회차 $moai --mode-agent 로 self-pick (1)부터. compact·clear 뒤 첫 행동은 내 카드 Issue 의 moai-resume 메모 읽기"},
        ])
    );
    assert_eq!(codex["result"]["verified"], json!("not_run"));
    let (_, third) = run(
        &["lane-loop-line", "--host", "omp", "--lane", "agent-1"],
        dir.path(),
    );
    assert_eq!(
        third["result"]["lines"],
        json!([{"kind": "flag", "text": "--goal \"저장소에 열린 PR이 0이고 시작할 수 있는 Issue가 0임(HOLD·needs:decision 제외)\""}])
    );
    for v in [&claude, &codex, &third] {
        assert_eq!(v["result"]["writes"], json!(false));
    }
    // A bad host or a codex line without its repository is a refusal, not a line.
    let (code, bad) = run(
        &["lane-loop-line", "--host", "cursor", "--lane", "agent-1"],
        dir.path(),
    );
    assert_ne!(code, 0, "{bad}");
    let (code, bad) = run(
        &["lane-loop-line", "--host", "codex", "--lane", "agent-1"],
        dir.path(),
    );
    assert_ne!(code, 0, "{bad}");
}

/// AC-4: no lane Stop hook in the project's `.codex/hooks.json` → a `LANE-BLOCKED … wake` line;
/// the lane's own hook → none, with trust left `unknown`.
#[test]
fn missing_hook_reports_lane_blocked_wake() {
    let dir = tempfile::tempdir().unwrap();
    let args = [
        "lane-loop-line",
        "--host",
        "codex",
        "--lane",
        "agent-1",
        "--repo",
        "o/r",
    ];
    let (_, v) = run(&args, dir.path());
    assert_eq!(v["result"]["stop_hook"]["state"], json!("missing"));
    let line = v["result"]["blocked_line"].as_str().unwrap();
    assert!(
        line.starts_with(
            "LANE-BLOCKED agent-1 wake codex Stop hook missing · kind=outside owner=user next="
        ),
        "{line}"
    );
    fs::create_dir(dir.path().join(".codex")).unwrap();
    let hooks = json!({"hooks": {"Stop": [{"hooks": [{"type": "command",
        "command": "moai-cli lane-stop-hook --lane agent-1 --repo o/r", "timeout": 120}]}]}});
    fs::write(dir.path().join(".codex/hooks.json"), hooks.to_string()).unwrap();
    let (_, v) = run(&args, dir.path());
    assert_eq!(v["result"]["stop_hook"]["state"], json!("installed"));
    assert_eq!(v["result"]["stop_hook"]["trust"], json!("unknown"));
    assert_eq!(v["result"]["blocked_line"], Value::Null);
    // Another project folder is read with --project.
    let other = tempfile::tempdir().unwrap();
    let mut with_project = args.to_vec();
    let p = dir.path().to_str().unwrap();
    with_project.extend(["--project", p]);
    let (_, v) = run(&with_project, other.path());
    assert_eq!(v["result"]["stop_hook"]["state"], json!("installed"));
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
