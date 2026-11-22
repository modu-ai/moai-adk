//! #2904 stage 2 (O5, #2899 B, user decision D-O2): the lane Stop command. `lane-stop-hook`
//! answers a Stop with `{"decision":"block",…}` until the end condition. No host file carries the
//! command any more (the Codex project layer was removed), so every case runs the command line
//! itself, in a temporary project and HOME, with saved `lane-next` snapshots: no git, no network.
mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::io::Write as _;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

struct Fixture {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
}

fn setup() -> Fixture {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let (home, project) = (root.join("home"), root.join("project"));
    fs::create_dir_all(&home).unwrap();
    // The state folder a block count is written into (an install used to make it).
    fs::create_dir_all(project.join(".moai-cli")).unwrap();
    Fixture {
        _tmp: tmp,
        home,
        project,
    }
}

fn run_in(f: &Fixture, exe: &str, args: &[&str], stdin: &[u8]) -> Output {
    let mut child = Command::new(exe)
        .clean_env()
        .env("HOME", &f.home)
        .env("PATH", "/usr/bin:/bin")
        .current_dir(&f.project)
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child.stdin.take().unwrap().write_all(stdin).unwrap();
    child.wait_with_output().unwrap()
}

/// The command line a lane Stop hook runs, without its leading `moai`.
const COMMAND: &str = "lane-stop-hook --lane agent-1 --repo o/r";

fn fixture(name: &str) -> String {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/lane_loop")
        .join(name)
        .to_str()
        .unwrap()
        .to_string()
}

/// Runs the lane Stop command with `--snapshot` added.
fn stop(f: &Fixture, snapshot: &str, stdin: &[u8]) -> (i32, String) {
    let mut args: Vec<String> = COMMAND.split_whitespace().map(str::to_string).collect();
    args.extend(["--snapshot".to_string(), fixture(snapshot)]);
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let out = run_in(f, common::MOAI_BIN, &args, stdin);
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8(out.stdout).unwrap(),
    )
}

const STOP_INPUT: &[u8] = br#"{"session_id":"s","transcript_path":null,"cwd":".","hook_event_name":"Stop","model":"m","turn_id":"t","stop_hook_active":true}"#;

fn block_reason(out: &str) -> String {
    assert!(out.ends_with('\n') && out.lines().count() == 1, "{out:?}");
    let v: Value = serde_json::from_str(out).unwrap();
    assert_eq!(v["decision"], json!("block"), "{v}");
    assert_eq!(v.as_object().unwrap().len(), 2, "{v}");
    v["reason"].as_str().unwrap().to_string()
}

fn state_file(f: &Fixture) -> PathBuf {
    moai_cli::lane_loop::stop_state_path(&f.project, "agent-1")
}

/// AC-2a: the Stop command blocks while the lane has work or open pull requests remain, passes at
/// the end condition, and passes an unknown reading the second time.
#[test]
fn stop_command_blocks_until_end_condition() {
    let f = setup();
    // stop_hook_active is ignored (Q1 (a)).
    let (code, out) = stop(&f, "work.json", STOP_INPUT);
    assert_eq!(code, 0);
    let reason = block_reason(&out);
    assert!(reason.contains("lane-next: P2 pick_issue #7"), "{reason}");
    assert!(
        reason.contains("moai lane-next --lane agent-1 --repo o/r"),
        "{reason}"
    );
    // No text read from GitHub (the Issue title) reaches the reason.
    assert!(!reason.contains("nobody claims"), "{reason}");
    // Bad or empty stdin still gets an answer.
    for input in [&b"not json"[..], b""] {
        let (code, out) = stop(&f, "work.json", input);
        assert_eq!(code, 0);
        block_reason(&out);
    }
    let (_, out) = stop(&f, "wait_prs.json", STOP_INPUT);
    let reason = block_reason(&out);
    assert!(
        reason.contains("moai lane-next --lane agent-1 --repo o/r --wait 1800"),
        "{reason}"
    );
    // The end condition: nothing on stdout, exit 0.
    assert_eq!(stop(&f, "end.json", STOP_INPUT), (0, String::new()));
    // Unknown (a partial read, an unreadable snapshot): block once, then pass.
    for snapshot in ["partial.json", "missing.json"] {
        block_reason(&stop(&f, snapshot, STOP_INPUT).1);
        assert_eq!(stop(&f, snapshot, STOP_INPUT), (0, String::new()));
    }
    // A malformed command line is a non-blocking hook error: exit 1, stdout empty.
    let out = run_in(
        &f,
        common::MOAI_BIN,
        &["lane-stop-hook", "--lane", "a"],
        b"",
    );
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
    let out = run_in(
        &f,
        common::MOAI_BIN,
        &["lane-stop-hook", "--lane", "a b", "--repo", "o/r"],
        b"",
    );
    assert_eq!(out.status.code(), Some(1));
    assert!(out.stdout.is_empty());
}

/// The cap: 48 blocks in a row for one next task, the 49th stop passes; a new next task
/// starts the count again.
#[test]
fn stop_command_passes_after_48_blocks_in_a_row() {
    let f = setup();
    block_reason(&stop(&f, "work.json", STOP_INPUT).1);
    let path = state_file(&f);
    let mut st: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(st["blocks"], json!(1));
    st["blocks"] = json!(46);
    fs::write(&path, st.to_string()).unwrap();
    block_reason(&stop(&f, "work.json", STOP_INPUT).1); // 47th
    block_reason(&stop(&f, "work.json", STOP_INPUT).1); // 48th
    assert_eq!(stop(&f, "work.json", STOP_INPUT), (0, String::new())); // 49th
    let st: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(st["blocks"], json!(48));
    let reason = block_reason(&stop(&f, "work_other.json", STOP_INPUT).1);
    assert!(reason.contains("#8"), "{reason}");
    let st: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(st["blocks"], json!(1));
}

/// #2926 audit B1: two next tasks that alternate block 96 stops, and the 97th passes.
#[test]
fn alternating_next_tasks_pass_after_96_blocks() {
    let f = setup();
    let snapshots = ["work.json", "work_other.json"];
    for i in 0..96 {
        block_reason(&stop(&f, snapshots[i % 2], STOP_INPUT).1);
    }
    let path = state_file(&f);
    let st: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(st["since_pass"], json!(96));
    assert_eq!(stop(&f, "work.json", STOP_INPUT), (0, String::new())); // 97th
    block_reason(&stop(&f, "work_other.json", STOP_INPUT).1);
}
