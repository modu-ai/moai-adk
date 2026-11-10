//! `decision-append` (rc.2 #2403 item 4): adds rows to the end of a decisions log's single
//! table, only when the log in the worktree and index is byte-identical to HEAD and HEAD is
//! the base the caller names. Existing rows are never re-rendered, edited or removed; the
//! default is a preview that writes nothing.
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

const LOG: &str = "docs/decisions-test.md";
const HEADER: &str = "# 사용자 결정 기록 (test)\n\n안내 문단.\n\n| 시각 | 주제 | 결정 | 적용 Issue | 비고 |\n| --- | --- | --- | --- | --- |\n";
// The second row carries a bare `|` inside a code span, as the real log does (rows 76, 80):
// existing rows must be kept byte for byte and never split on it.
const ROWS: &str = "| 10:00Z | a | `cc|codex -l|-a` | #1 | — |\n| 11:00Z | b | 결정 b | #2 | — |\n";
const TAIL: &str = "\n부분 결정 분리: 뒤 문단.\n";

fn git(dir: &Path, args: &[&str]) -> String {
    common::git::git(dir, args)
}

fn repo_with(log: &str) -> (TempDir, String) {
    let tmp = TempDir::new().unwrap();
    let dir = tmp.path();
    git(dir, &["init", "-q", "-b", "main"]);
    fs::create_dir_all(dir.join("docs")).unwrap();
    fs::write(dir.join(LOG), log).unwrap();
    git(dir, &["add", "-A"]);
    git(
        dir,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            "init",
        ],
    );
    let head = git(dir, &["rev-parse", "HEAD"]).trim().to_string();
    (tmp, head)
}

fn standard_log() -> String {
    format!("{HEADER}{ROWS}{TAIL}")
}

fn run(dir: &Path, extra: &[&str], request: &Value) -> (i32, Value) {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("decision-append")
        .arg(dir)
        .args(["--log", LOG])
        .args(extra)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_INDEX_FILE")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let bytes = serde_json::to_vec(request).unwrap();
    let mut stdin = child.stdin.take().unwrap();
    common::stdin_write_outcome(stdin.write_all(&bytes));
    drop(stdin);
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("stdout not JSON: {}", String::from_utf8_lossy(&out.stdout)));
    (out.status.code().unwrap(), v)
}

fn request(base: &str, rows: Value) -> Value {
    json!({"protocol_version": 1, "base": base, "rows": rows})
}

fn one_row() -> Value {
    json!([["12:00Z", "c", "결정 c", "#3", "—"]])
}

fn apply_flags() -> [&'static str; 3] {
    ["--apply", "--permission-reference", "test-ref"]
}

fn error_message(v: &Value) -> String {
    v["error"]["message"].as_str().unwrap_or("").to_string()
}

#[test]
fn preview_writes_nothing_and_reports_the_insert() {
    let (tmp, head) = repo_with(&standard_log());
    let before = fs::read(tmp.path().join(LOG)).unwrap();
    let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
    assert_eq!(code, 0, "{v}");
    let r = &v["result"];
    assert_eq!(r["mode"], json!("preview"));
    assert_eq!(r["rows_added"], json!(1));
    assert_eq!(r["lines_removed"], json!(0));
    assert_eq!(r["remote_checked"], json!(false));
    assert_eq!(r["writes"], json!([]));
    assert_eq!(fs::read(tmp.path().join(LOG)).unwrap(), before);
    assert_eq!(git(tmp.path(), &["status", "--porcelain"]), "");
}

#[test]
fn apply_adds_rows_after_the_last_row_and_keeps_every_existing_byte() {
    let (tmp, head) = repo_with(&standard_log());
    let rows = json!([
        ["12:00Z", "c", "결정 c", "#3", "—"],
        ["12:30Z", "d", "결정 d", "#4", "—"]
    ]);
    let (code, v) = run(tmp.path(), &apply_flags(), &request(&head, rows));
    assert_eq!(code, 0, "{v}");
    let after = fs::read_to_string(tmp.path().join(LOG)).unwrap();
    let want = format!(
        "{HEADER}{ROWS}| 12:00Z | c | 결정 c | #3 | — |\n| 12:30Z | d | 결정 d | #4 | — |\n{TAIL}"
    );
    assert_eq!(after, want);
    assert_eq!(v["result"]["mode"], json!("applied"));
    assert_eq!(v["result"]["rows_added"], json!(2));
    assert_eq!(v["result"]["writes"], json!([LOG]));
    // The only change is two added lines; nothing is removed.
    let stat = git(tmp.path(), &["diff", "--numstat", "--", LOG]);
    assert_eq!(stat.trim(), format!("2\t0\t{LOG}"));
    // No temporary file is left next to the log.
    let names: Vec<String> = fs::read_dir(tmp.path().join("docs"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names, vec!["decisions-test.md".to_string()]);
}

#[test]
fn a_second_apply_without_a_commit_is_blocked() {
    let (tmp, head) = repo_with(&standard_log());
    assert_eq!(
        run(tmp.path(), &apply_flags(), &request(&head, one_row())).0,
        0
    );
    let after_first = fs::read(tmp.path().join(LOG)).unwrap();
    let (code, v) = run(tmp.path(), &apply_flags(), &request(&head, one_row()));
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], json!("blocked"), "{v}");
    assert!(error_message(&v).contains("differs from HEAD"), "{v}");
    assert_eq!(fs::read(tmp.path().join(LOG)).unwrap(), after_first);
}

#[test]
fn a_base_other_than_head_is_blocked() {
    let (tmp, _head) = repo_with(&standard_log());
    let (code, v) = run(tmp.path(), &[], &request(&"a".repeat(40), one_row()));
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], json!("blocked"), "{v}");
    assert!(
        error_message(&v).contains("HEAD is not the given base"),
        "{v}"
    );
}

#[test]
fn a_locally_edited_log_is_blocked_even_when_the_base_matches() {
    // A stale primary checkout whose file was edited by hand is exactly the 21-line loss case.
    let (tmp, head) = repo_with(&standard_log());
    fs::write(tmp.path().join(LOG), format!("{HEADER}{ROWS}")).unwrap();
    let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
    assert_ne!(code, 0);
    assert_eq!(v["error"]["kind"], json!("blocked"), "{v}");
    assert!(error_message(&v).contains("differs from HEAD"), "{v}");
}

#[test]
fn a_staged_change_to_the_log_is_blocked() {
    let (tmp, head) = repo_with(&standard_log());
    fs::write(tmp.path().join(LOG), format!("{HEADER}{TAIL}")).unwrap();
    git(tmp.path(), &["add", LOG]);
    // Restore the worktree copy so only the index differs from HEAD.
    fs::write(tmp.path().join(LOG), standard_log()).unwrap();
    let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
    assert_ne!(code, 0);
    assert!(error_message(&v).contains("differs from HEAD"), "{v}");
}

#[test]
fn apply_requires_a_permission_reference() {
    let (tmp, head) = repo_with(&standard_log());
    let (code, v) = run(tmp.path(), &["--apply"], &request(&head, one_row()));
    assert_ne!(code, 0);
    assert!(
        error_message(&v).contains("permission reference required"),
        "{v}"
    );
    assert_eq!(
        fs::read_to_string(tmp.path().join(LOG)).unwrap(),
        standard_log()
    );
}

#[test]
fn permission_reference_placeholders_and_the_128_char_edge() {
    // Refused values: the three placeholders, blank, 129 characters, a control character.
    // Each leaves the log byte for byte as it was.
    let long_ok = "r".repeat(128);
    let long_bad = "r".repeat(129);
    for bad in [
        "null",
        "placeholder",
        "   ",
        long_bad.as_str(),
        "ref\u{7}bell",
    ] {
        let (tmp, head) = repo_with(&standard_log());
        let (code, v) = run(
            tmp.path(),
            &["--apply", "--permission-reference", bad],
            &request(&head, one_row()),
        );
        assert_ne!(code, 0, "{bad:?} -> {v}");
        assert!(
            error_message(&v).contains("permission reference required"),
            "{bad:?} -> {v}"
        );
        assert_eq!(
            fs::read_to_string(tmp.path().join(LOG)).unwrap(),
            standard_log(),
            "{bad:?}"
        );
    }
    // 128 characters is the largest accepted reference.
    let (tmp, head) = repo_with(&standard_log());
    let (code, v) = run(
        tmp.path(),
        &["--apply", "--permission-reference", &long_ok],
        &request(&head, one_row()),
    );
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["mode"], json!("applied"));
}

#[test]
fn rows_go_after_the_table_not_after_a_later_pipe_line() {
    // A paragraph after the table holds a line that starts with `|`. The table ends at the
    // first line that is not a row, so the insert point stays after the table's last row.
    let stray_tail = "\n문단.\n| stray |\n";
    let log = format!("{HEADER}{ROWS}{stray_tail}");
    let (tmp, head) = repo_with(&log);
    let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
    assert_eq!(code, 0, "{v}");
    // HEADER is 6 lines and ROWS 2, so the table's last row is line 8.
    assert_eq!(v["result"]["insert_after_line"], json!(8), "{v}");
    let (code, v) = run(tmp.path(), &apply_flags(), &request(&head, one_row()));
    assert_eq!(code, 0, "{v}");
    assert_eq!(
        fs::read_to_string(tmp.path().join(LOG)).unwrap(),
        format!("{HEADER}{ROWS}| 12:00Z | c | 결정 c | #3 | — |\n{stray_tail}")
    );
}

#[test]
fn zero_rows_are_refused_and_one_row_is_accepted() {
    let (tmp, head) = repo_with(&standard_log());
    let (code, v) = run(tmp.path(), &[], &request(&head, json!([])));
    assert_ne!(code, 0);
    assert!(error_message(&v).contains("at least one row"), "{v}");
    assert_eq!(run(tmp.path(), &[], &request(&head, one_row())).0, 0);
}

#[test]
fn a_row_with_the_wrong_cell_count_is_refused() {
    let (tmp, head) = repo_with(&standard_log());
    // Header has 5 columns: 4 and 6 are refused, 5 passes.
    for (cells, ok) in [
        (json!([["a", "b", "c", "d"]]), false),
        (json!([["a", "b", "c", "d", "e", "f"]]), false),
        (json!([["a", "b", "c", "d", "e"]]), true),
    ] {
        let (code, v) = run(tmp.path(), &[], &request(&head, cells.clone()));
        assert_eq!(code == 0, ok, "{cells} -> {v}");
        if !ok {
            assert!(error_message(&v).contains("cells"), "{v}");
        }
    }
}

#[test]
fn a_bare_pipe_or_a_line_break_in_a_new_cell_is_refused() {
    let (tmp, head) = repo_with(&standard_log());
    for bad in ["x | y", "`a|b`", "line\nbreak", "cr\rhere"] {
        let rows = json!([["12:00Z", "c", bad, "#3", "—"]]);
        let (code, v) = run(tmp.path(), &[], &request(&head, rows));
        assert_ne!(code, 0, "{bad:?} -> {v}");
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{v}");
    }
    // An escaped pipe is allowed, as row 104 of the real log writes it.
    let rows = json!([["12:00Z", "c", "`leader\\|agent`", "#3", "—"]]);
    assert_eq!(run(tmp.path(), &[], &request(&head, rows)).0, 0);
}

#[test]
fn a_repeated_time_key_is_a_warning_not_a_refusal() {
    let (tmp, head) = repo_with(&standard_log());
    let rows = json!([["11:00Z", "again", "결정", "#9", "—"]]);
    let (code, v) = run(tmp.path(), &[], &request(&head, rows));
    assert_eq!(code, 0, "{v}");
    let warnings = v["result"]["warnings"].as_array().unwrap();
    assert_eq!(warnings.len(), 1, "{v}");
    assert!(warnings[0].as_str().unwrap().contains("11:00Z"), "{v}");
    let rows = json!([["12:00Z", "new", "결정", "#9", "—"]]);
    assert_eq!(
        run(tmp.path(), &[], &request(&head, rows)).1["result"]["warnings"],
        json!([])
    );
}

#[test]
fn a_log_without_exactly_one_table_is_refused() {
    for log in [
        "# 제목\n\n표가 없는 문서.\n".to_string(),
        format!(
            "{HEADER}{ROWS}\n{}",
            &HEADER[HEADER.find("| 시각").unwrap()..]
        ),
    ] {
        let (tmp, head) = repo_with(&log);
        let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
        assert_ne!(code, 0, "{log}");
        assert!(error_message(&v).contains("exactly one table"), "{v}");
    }
}

#[test]
fn a_last_row_without_a_line_terminator_is_refused() {
    let (tmp, head) = repo_with(&format!("{HEADER}{}", ROWS.trim_end_matches('\n')));
    let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
    assert_ne!(code, 0);
    assert!(error_message(&v).contains("line terminator"), "{v}");
}

#[test]
fn log_paths_outside_the_repository_or_untracked_are_refused() {
    let (tmp, head) = repo_with(&standard_log());
    fs::write(tmp.path().join("docs/untracked.md"), standard_log()).unwrap();
    // Staged but never committed: in the index, absent from HEAD.
    fs::write(tmp.path().join("docs/staged.md"), standard_log()).unwrap();
    git(tmp.path(), &["add", "docs/staged.md"]);
    // `std::os::unix` exists only on Unix; the symlink case is Unix-only so that the
    // Windows `--all-targets` build still compiles this test.
    #[cfg(unix)]
    std::os::unix::fs::symlink(LOG, tmp.path().join("docs/link.md")).unwrap();
    let symlink_case: &[(&str, &str)] = if cfg!(unix) {
        &[("docs/link.md", "symlink")]
    } else {
        &[]
    };
    let cases = [
        ("../outside.md", "relative path inside the repository"),
        ("/etc/hosts", "relative path inside the repository"),
        ("docs/untracked.md", "not tracked"),
        ("docs/staged.md", "not tracked"),
        ("docs/missing.md", "not tracked"),
    ];
    for &(path, needle) in cases.iter().chain(symlink_case) {
        let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .arg("decision-append")
            .arg(tmp.path())
            .args(["--log", path])
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        let mut stdin = child.stdin.take().unwrap();
        common::stdin_write_outcome(
            stdin.write_all(&serde_json::to_vec(&request(&head, one_row())).unwrap()),
        );
        drop(stdin);
        let out = child.wait_with_output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_ne!(out.status.code(), Some(0), "{path}");
        assert!(error_message(&v).contains(needle), "{path}: {v}");
    }
}

#[test]
fn the_real_decisions_log_previews_cleanly() {
    // The repository's own log (bare `|` inside code spans in some rows) must parse as one
    // table and preview an append without touching it.
    let real = Path::new(env!("CARGO_MANIFEST_DIR")).join("docs/decisions-2026-11-08.md");
    let text = fs::read_to_string(&real).unwrap();
    let (tmp, head) = repo_with(&text);
    let (code, v) = run(tmp.path(), &[], &request(&head, one_row()));
    assert_eq!(code, 0, "{v}");
    assert_eq!(v["result"]["lines_removed"], json!(0));
    assert_eq!(fs::read_to_string(tmp.path().join(LOG)).unwrap(), text);
}
