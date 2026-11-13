//! Issue #2666 F1: `flake-classify` reads two `gh run view --log` texts (cargo-nextest output)
//! and sorts every test that failed in the first: passed later (flaky), failed again (still
//! failing), or absent. Per job; the panic location and the line after it travel along.

use serde_json::{json, Value};
use std::fs;
use std::process::Command;
use tempfile::TempDir;

const MAC: &str = "ci / macos-arm64";
const LINUX: &str = "ci / linux-x86_64";

fn line(job: &str, text: &str) -> String {
    format!("{job}\tUNKNOWN STEP\t2026-11-13T07:34:25.2985410Z {text}\n")
}

fn status(job: &str, s: &str, n: u32, test: &str) -> String {
    line(
        job,
        &format!("        {s} [   0.018s] ( {n:>3}/1604) {test}"),
    )
}

fn classify(first: &str, later: &str, job: Option<&str>) -> Value {
    let dir = TempDir::new().unwrap();
    let (a, b) = (dir.path().join("a.log"), dir.path().join("b.log"));
    fs::write(&a, first).unwrap();
    fs::write(&b, later).unwrap();
    let mut args = vec![
        "flake-classify".to_string(),
        "--failed".into(),
        a.to_str().unwrap().into(),
        "--later".into(),
        b.to_str().unwrap().into(),
    ];
    if let Some(j) = job {
        args.extend(["--job".into(), j.into()]);
    }
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(&args)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn tests(rows: &Value) -> Vec<String> {
    rows.as_array()
        .unwrap()
        .iter()
        .map(|r| r["test"].as_str().unwrap().to_string())
        .collect()
}

const FLAKY: &str = "moai-cli::apply_receipt apply_recover_restores_only_unchanged_targets";
const REAL: &str = "moai-cli::agents_home_cli revert_undoes_only_unchanged_files";
const GONE: &str = "moai-cli::archive_move_cli a_move_reads_back";
const UNIT: &str = "moai-cli statusline::tests::piped_stdout_keeps_color";

const H1: &str = "1111111111111111111111111111111111111111";
const H2: &str = "2222222222222222222222222222222222222222";
const BASE: &str = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";

/// The checkout of a push run: `git log -1 --format=%H` and the SHA it prints.
fn checkout(job: &str, sha: &str) -> String {
    line(job, "[command]/usr/bin/git log -1 --format=%H") + &line(job, sha)
}

/// The checkout of a pull-request run: a merge commit naming the PR head.
fn merge_checkout(job: &str, head: &str) -> String {
    line(
        job,
        &format!("HEAD is now at 9d60f5acd1 Merge {head} into {BASE}"),
    )
}

fn first() -> String {
    [
        checkout(MAC, H1),
        checkout(LINUX, H1),
        // The same function panicked on another job first: its message stays with that job.
        line(LINUX, "    thread 'apply_recover_restores_only_unchanged_targets' (7) panicked at tests/apply_receipt.rs:99:1:"),
        line(LINUX, "    linux-only message"),
        status(MAC, "PASS", 1, "moai-cli::ok always_passes"),
        status(MAC, "FAIL", 2, FLAKY),
        line(MAC, "    thread 'apply_recover_restores_only_unchanged_targets' (21936) panicked at tests/apply_receipt.rs:26:10:"),
        line(MAC, "    called `Result::unwrap()` on an `Err` value: Os { code: 2, kind: NotFound }"),
        status(MAC, "FAIL", 3, REAL),
        status(MAC, "SIGSEGV", 4, UNIT),
        status(MAC, "FAIL", 5, GONE),
        status(LINUX, "FAIL", 6, FLAKY),
        // nextest repeats the failures in its summary.
        status(MAC, "FAIL", 2, FLAKY),
    ]
    .concat()
}

fn later() -> String {
    [
        checkout(MAC, H1),
        checkout(LINUX, H1),
        status(MAC, "PASS", 2, FLAKY),
        status(MAC, "FAIL", 3, REAL),
        status(MAC, "FAIL", 4, UNIT),
        status(MAC, "PASS", 4, UNIT), // a retry that passed
        status(LINUX, "FAIL", 6, FLAKY),
    ]
    .concat()
}

#[test]
fn failed_tests_are_sorted_against_the_later_run() {
    let c = classify(&first(), &later(), None);
    // Rows come sorted by (job, test id).
    assert_eq!(tests(&c["flaky"]), vec![UNIT, FLAKY]);
    assert_eq!(c["flaky"][1]["job"], json!(MAC));
    assert_eq!(
        c["flaky"][0]["message"],
        Value::Null,
        "no panic line for it"
    );
    assert_eq!(
        c["flaky"][1]["message"],
        json!("panicked at tests/apply_receipt.rs:26:10 called `Result::unwrap()` on an `Err` value: Os { code: 2, kind: NotFound }")
    );
    // The same test on another job failed again there: per job, not flaky.
    assert_eq!(tests(&c["still_failing"]), vec![FLAKY, REAL]);
    assert_eq!(c["still_failing"][0]["job"], json!(LINUX));
    assert_eq!(tests(&c["not_in_later"]), vec![GONE]);
    // A test that passed in the first log is never reported.
    assert!(!c.to_string().contains("always_passes"));
}

#[test]
fn a_job_filter_keeps_one_job() {
    let c = classify(&first(), &later(), Some(LINUX));
    assert_eq!(tests(&c["flaky"]), Vec::<String>::new());
    assert_eq!(tests(&c["still_failing"]), vec![FLAKY]);
    let c = classify(&first(), &later(), Some("ci / macos"));
    for k in [
        "flaky",
        "head_mismatch",
        "head_unverified",
        "still_failing",
        "not_in_later",
    ] {
        assert_eq!(c[k], json!([]), "exact match only: {k}");
    }
    assert_eq!(c["first_heads"], json!({}));
}

#[test]
fn lines_that_are_not_status_lines_are_ignored() {
    let noise = [
        line(MAC, "    Starting 1604 tests across 140 binaries"),
        line(MAC, "PASS is a word in a sentence here"),
        line(MAC, "        FAIL [   0.1s] ( 1/2)"),
        "no tabs at all FAIL [ 0.1s] ( 1/2) moai-cli::x z\n".to_string(),
        "        FAIL [   0.1s] ( 1/2) moai-cli::x y\n".to_string(),
    ]
    .concat();
    let c = classify(&noise, "", None);
    // A tab-less line is all text: only one that starts with the status counts, under the
    // empty job name.
    assert_eq!(tests(&c["not_in_later"]), vec!["moai-cli::x y"]);
    assert_eq!(c["not_in_later"][0]["job"], json!(""));
}

#[test]
fn a_missing_or_oversized_log_is_refused() {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args([
            "flake-classify",
            "--failed",
            "/nonexistent/a.log",
            "--later",
            "/nonexistent/b.log",
        ])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(false));
}

#[test]
fn a_panic_line_takes_its_next_line_only_from_the_same_job() {
    let log = [
        status(MAC, "FAIL", 1, FLAKY),
        line(MAC, "thread 'apply_recover_restores_only_unchanged_targets' (1) panicked at tests/a.rs:1:2:"),
        line(LINUX, "a line of another job in between"),
        line(MAC, "the mac line that came later"),
    ]
    .concat();
    let c = classify(&log, "", None);
    assert_eq!(
        c["not_in_later"][0]["message"],
        json!("panicked at tests/a.rs:1:2")
    );
}

#[test]
fn a_tabbed_line_without_a_timestamp_keeps_its_first_word() {
    let log = format!("{MAC}\tstep\tFAIL [   0.1s] ( 1/2) moai-cli::x y\n");
    let c = classify(&log, "", None);
    assert_eq!(tests(&c["not_in_later"]), vec!["moai-cli::x y"]);
}

/// One failure first and a pass later, with the given checkout text in each log.
fn pair(first_head: &str, later_head: &str) -> Value {
    let first = first_head.to_string() + &status(MAC, "FAIL", 1, FLAKY);
    let later = later_head.to_string() + &status(MAC, "PASS", 1, FLAKY);
    classify(&first, &later, None)
}

#[test]
fn a_pass_on_another_head_is_not_flaky() {
    // Audit 6858491683 B1: runs 43446000944 and 43452036573 of one PR, the test fixed between.
    let c = pair(&merge_checkout(MAC, H1), &merge_checkout(MAC, H2));
    assert_eq!(tests(&c["flaky"]), Vec::<String>::new(), "{c}");
    assert_eq!(tests(&c["head_mismatch"]), vec![FLAKY]);
    assert_eq!(c["first_heads"], json!({MAC: H1}));
    assert_eq!(c["later_heads"], json!({MAC: H2}));

    // Same PR head on two merge commits (the base moved): flaky.
    let c = pair(&merge_checkout(MAC, H1), &merge_checkout(MAC, H1));
    assert_eq!(tests(&c["flaky"]), vec![FLAKY], "{c}");

    // A push run and a PR run of the same head.
    let c = pair(&checkout(MAC, H1), &merge_checkout(MAC, H1));
    assert_eq!(tests(&c["flaky"]), vec![FLAKY], "{c}");

    // A merge line names the PR head even after a checkout SHA (the merge commit) was printed.
    let c = pair(
        &(checkout(MAC, H2) + &merge_checkout(MAC, H1)),
        &checkout(MAC, H1),
    );
    assert_eq!(tests(&c["flaky"]), vec![FLAKY], "{c}");
}

#[test]
fn an_unread_head_leaves_the_pass_unverified() {
    let c = pair("", &checkout(MAC, H1));
    assert_eq!(tests(&c["head_unverified"]), vec![FLAKY], "{c}");
    let c = pair(&checkout(MAC, H1), "");
    assert_eq!(tests(&c["head_unverified"]), vec![FLAKY], "{c}");
    // The SHA must follow its command on the same job and be 40 hex.
    let c = pair(
        &(line(MAC, "[command]/usr/bin/git log -1 --format=%H") + &line(LINUX, H1)),
        &checkout(MAC, H1),
    );
    assert_eq!(tests(&c["head_unverified"]), vec![FLAKY], "{c}");
    let c = pair(
        &(line(MAC, "[command]/usr/bin/git log -1 --format=%H") + &line(MAC, "1111111")),
        &checkout(MAC, H1),
    );
    assert_eq!(tests(&c["head_unverified"]), vec![FLAKY], "{c}");
    // A failure that stays a failure needs no head.
    let first = status(MAC, "FAIL", 1, REAL);
    let c = classify(&first, &first, None);
    assert_eq!(tests(&c["still_failing"]), vec![REAL]);
}

#[test]
fn every_failing_status_counts_and_status_lines_are_counted() {
    let mut first = checkout(MAC, H1);
    let mut later = checkout(MAC, H1);
    for (i, s) in ["TIMEOUT", "SIGABRT", "SIGKILL"].iter().enumerate() {
        let test = format!("moai-cli::x t{i}");
        first += &status(MAC, s, 1, &test);
        later += &status(MAC, "PASS", 1, &test);
    }
    let c = classify(&first, &later, None);
    assert_eq!(
        tests(&c["flaky"]),
        vec!["moai-cli::x t0", "moai-cli::x t1", "moai-cli::x t2"]
    );
    assert_eq!(c["status_lines"], json!({"first": 3, "later": 3}));
    let c = classify("not nextest output\n", "", None);
    assert_eq!(c["status_lines"], json!({"first": 0, "later": 0}));
}
