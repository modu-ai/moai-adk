//! SPEC-20261117-green-test-lock (card C-86 slice 3): the test-path table, `moai tdd
//! red|green|reopen|done|status` and the tool-preflight rule A6 that locks the RED tests during
//! GREEN and REFACTOR. Every case runs the real binary in a temporary repository with a
//! temporary HOME; the checks are small `sh` scripts inside that repository.
#![cfg(unix)]
mod common;

use moai_cli::run::{oracle_changes, oracle_changes_in};
use moai_cli::tdd::{is_test_path, validate_declared, TestPaths, TEST_PATTERNS};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const BRANCH: &str = "spec/20261117-demo";
const SPEC: &str = ".moai/specs/SPEC-20261117-demo/spec.md";
const SPEC_TEXT: &str = "---\nid: SPEC-20261117-demo\ntitle: \"demo\"\nstatus: draft\nrevision: 1\ncreated: 2026-11-17\n---\n\n## Intent\n\nx\n";
/// Fails with a cargo-style assertion summary until `src/lib.txt` says `fixed`.
const CHECK: &str = "sh check.sh";
const CHECK_SH: &str =
    "if grep -q fixed src/lib.txt; then echo 'test result: ok. 1 passed; 0 failed'; \
                        else echo 'test result: FAILED. 0 passed; 1 failed'; exit 101; fi\n";

// ---- the table ----

/// AC-6 (REQ-11): every language the project detector can report has a row.
#[test]
fn every_detected_language_has_test_patterns() {
    let source = include_str!("../src/project_detect.rs");
    let mut langs = std::collections::BTreeSet::new();
    for table in ["const MANIFESTS", "const EXTENSIONS"] {
        let start = source.find(table).unwrap();
        let end = start + source[start..].find("];").unwrap();
        for line in source[start..end].lines() {
            let quoted: Vec<&str> = line.split('"').skip(1).step_by(2).collect();
            if quoted.len() == 2 {
                langs.insert(quoted[1].to_string());
            }
        }
    }
    assert!(langs.contains("rust") && langs.len() >= 13, "{langs:?}");
    for lang in &langs {
        assert!(
            TEST_PATTERNS.iter().any(|(l, _)| l == lang),
            "no test patterns for {lang}"
        );
    }
}

/// AC-6 (REQ-11), N and N+1: each pattern against the closest path it must not take.
#[test]
fn test_paths_match_their_convention_and_not_its_neighbours() {
    let none: [String; 0] = [];
    for yes in [
        "tests/a.rs",
        "crates/x/tests/a.rs",
        "benches/b.rs",
        "test_login.py",
        "pkg/login_test.py",
        "tests/conftest.py",
        "conftest.py",
        "src/login.test.ts",
        "src/login.spec.js",
        "src/__tests__/a.ts",
        "pkg/server_test.go",
        "pkg/testdata/in.json",
        "src/test/java/AppTest.java",
        "spec/login_spec.rb",
        "lib/login_test.rb",
        "src/LoginTest.php",
        "App.Tests/Login.cs",
        "Tests/AppTests/LoginTests.swift",
        "crates/report/golden/summary.txt",
        "a/fixtures/b.json",
        "./tests/a.rs",
        "tests\\a.rs",
    ] {
        assert!(is_test_path(yes, &none), "{yes}");
    }
    for no in [
        "src/lib.rs",
        "testing/a.rs",
        "tests",
        "src/tests.rs",
        "test.py",
        "test_.py",
        "contest.py",
        "_test.py",
        "src/login.ts",
        "src/.test.ts",
        "src/a.test",
        "server_test.go.bak",
        "_test.go",
        "src/Test.php",
        "src/latest/a.rs",
        "specs/a.md",
        ".moai/specs/SPEC-1/spec.md",
        "",
    ] {
        assert!(!is_test_path(no, &none), "{no}");
    }
    let declared = vec!["qa/".to_string(), "e2e/smoke.sh".to_string()];
    assert!(is_test_path("qa/login.sh", &declared));
    assert!(is_test_path("e2e/smoke.sh", &declared));
    assert!(!is_test_path("qa2/login.sh", &declared));
    assert!(!is_test_path("e2e/smoke.sh.bak", &declared));
}

/// Audit N1: a project uses the rows of its own languages and `common`. This repository's
/// profile detects Python and Rust, so its SPEC templates under `templates/spec/` and Ruby or
/// JavaScript conventions are not test paths here; `tests/`, `*_test.py` and golden files are.
#[test]
fn this_repository_uses_only_its_own_languages() {
    // A temporary project holding this repository's profile, byte for byte.
    let dir = tempfile::TempDir::new().unwrap();
    fs::create_dir_all(dir.path().join(".moai-cli")).unwrap();
    fs::write(
        dir.path().join(".moai-cli/project.yaml"),
        include_str!("../.moai-cli/project.yaml"),
    )
    .unwrap();
    let here = TestPaths::for_project(dir.path());
    assert_eq!(
        here.languages(),
        Some(&["python".to_string(), "rust".to_string()][..])
    );
    for no in [
        "templates/spec/full.md",
        "templates/spec/minimal.md",
        "spec/login_spec.rb",
        "src/login.test.ts",
        "test/a.c",
        "pkg/server_test.go",
    ] {
        assert!(!here.matches(no), "{no}");
        let unknown_project = is_test_path(no, &[]);
        assert!(unknown_project, "{no} is a test path of an unknown project");
    }
    for yes in [
        "tests/a.rs",
        "benches/b.rs",
        "tools/check_test.py",
        "crates/report/golden/summary.txt",
    ] {
        assert!(here.matches(yes), "{yes}");
    }
}

/// Audit N1 without a profile: the root manifests name the languages (N: one manifest; N+1: a
/// second adds its rows); no manifest at all keeps every row. `run`'s oracle list narrows the
/// same way.
#[test]
fn root_manifests_narrow_the_rows_and_none_keeps_them_all() {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    let paths: Vec<String> = [
        "spec/login_spec.rb",
        "tests/a.rs",
        "tools/check_test.py",
        "a/golden/x.txt",
    ]
    .map(String::from)
    .to_vec();
    assert_eq!(TestPaths::for_project(root).languages(), None);
    assert_eq!(oracle_changes_in(root, &paths), oracle_changes(&paths));
    assert_eq!(oracle_changes(&paths).len(), 4);
    fs::write(root.join("Cargo.toml"), "[package]\n").unwrap();
    let rust = TestPaths::for_project(root);
    assert_eq!(rust.languages(), Some(&["rust".to_string()][..]));
    assert!(!rust.matches("spec/login_spec.rb") && rust.matches("tests/a.rs"));
    assert_eq!(
        oracle_changes_in(root, &paths),
        ["a/golden/x.txt", "tests/a.rs"]
    );
    fs::write(root.join("Gemfile"), "source 'x'\n").unwrap();
    let both = TestPaths::for_project(root);
    assert!(both.matches("spec/login_spec.rb") && both.matches("tests/a.rs"));
    assert!(!both.matches("tools/check_test.py"));
}

/// The declared list at its edges: 32 paths and 200 characters are taken, one more is not.
#[test]
fn declared_test_paths_are_bounded_and_relative() {
    let n = |count: usize| (0..count).map(|i| format!("qa{i}/")).collect::<Vec<_>>();
    assert!(validate_declared(&n(32)).is_ok());
    assert!(validate_declared(&n(33)).is_err());
    assert!(validate_declared(&["a".repeat(200)]).is_ok());
    assert!(validate_declared(&["a".repeat(201)]).is_err());
    for bad in ["/abs", "\\abs", "C:/x", "a/../b", "..", "", " ", "a\nb"] {
        assert!(validate_declared(&[bad.to_string()]).is_err(), "{bad:?}");
    }
}

// ---- the command and the hook ----

struct Run {
    code: i32,
    stdout: String,
}

fn git(cwd: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

fn commit(root: &Path, message: &str) {
    git(root, &["add", "-A"]);
    git(root, &["commit", "-q", "-m", message]);
}

/// A temporary repository on `branch` with a failing check, an older test and its own HOME.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q", "-b", branch]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    fs::write(root.join("check.sh"), CHECK_SH).unwrap();
    fs::create_dir_all(root.join("src")).unwrap();
    fs::write(root.join("src/lib.txt"), "broken\n").unwrap();
    fs::create_dir_all(root.join("tests")).unwrap();
    fs::write(root.join("tests/old_test.rs"), "old\n").unwrap();
    // A5 (spec-before-code) refuses code writes on a SPEC branch until its SPEC exists.
    fs::create_dir_all(root.join(SPEC).parent().unwrap()).unwrap();
    fs::write(root.join(SPEC), SPEC_TEXT).unwrap();
    commit(&root, "init");
    (dir, root)
}

/// A RED commit: a new test and the trailer.
fn red_commit(root: &Path) {
    fs::write(root.join("tests/login_test.rs"), "assert fixed\n").unwrap();
    commit(root, "test: login fails first\n\nMoai-Phase: red");
}

fn moai(root: &Path, args: &[&str]) -> Run {
    let out = Command::new(BIN)
        .args(args)
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
    }
}

fn ok(r: &Run) -> Value {
    let v: Value =
        serde_json::from_str(r.stdout.lines().next().unwrap_or("")).unwrap_or(json!(null));
    assert_eq!((r.code, &v["ok"]), (0, &json!(true)), "{}", r.stdout);
    v["result"].clone()
}

fn refused(r: &Run) -> String {
    let v: Value =
        serde_json::from_str(r.stdout.lines().next().unwrap_or("")).unwrap_or(json!(null));
    assert!(
        r.code != 0 && v["ok"] == json!(false),
        "{} {}",
        r.code,
        r.stdout
    );
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

fn tdd(root: &Path, args: &[&str]) -> Run {
    let mut all = vec!["tdd"];
    all.extend_from_slice(args);
    moai(root, &all)
}

fn write_input(path: &str, cwd: &Path) -> Value {
    json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
           "tool_input": {"file_path": path, "content": "x\n"}, "cwd": cwd})
}

fn preflight(cwd: &Path, input: &Value) -> Run {
    let mut child = Command::new(BIN)
        .args([
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ])
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    Run {
        code: out.status.code().unwrap(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
    }
}

fn denied(r: &Run) -> bool {
    r.code == 0 && r.stdout.contains("\"permissionDecision\":\"deny\"")
}

fn allowed(r: &Run) -> bool {
    r.code == 0 && r.stdout.is_empty()
}

fn record_path(root: &Path) -> PathBuf {
    root.join(".git/moai/tdd.json")
}

fn log_events(root: &Path) -> Vec<String> {
    fs::read_to_string(root.join(".git/moai/tdd-log.jsonl"))
        .unwrap_or_default()
        .lines()
        .map(|l| {
            serde_json::from_str::<Value>(l).unwrap()["event"]
                .as_str()
                .unwrap()
                .to_string()
        })
        .collect()
}

/// AC-1 (REQ-1, REQ-2): RED needs the trailer, a test path and an assertion failure; each
/// missing piece is refused and nothing is recorded.
#[test]
fn red_needs_the_trailer_a_test_path_and_an_assertion_failure() {
    let (_d, root) = repo(BRANCH);
    // No trailer.
    fs::write(root.join("tests/login_test.rs"), "assert fixed\n").unwrap();
    commit(&root, "test: login fails first");
    assert!(refused(&tdd(&root, &["red", "--check", CHECK])).contains("Moai-Phase: red"));
    // A trailer, but no test path.
    fs::write(root.join("src/other.txt"), "x\n").unwrap();
    commit(&root, "chore: other\n\nMoai-Phase: red");
    assert!(refused(&tdd(&root, &["red", "--check", CHECK])).contains("no test path"));
    // A trailer and a test path, but the check passes.
    fs::write(root.join("tests/login_test.rs"), "assert fixed 2\n").unwrap();
    commit(&root, "test: more\n\nMoai-Phase: red");
    assert!(refused(&tdd(&root, &["red", "--check", "true"])).contains("assertion failures only"));
    // A failure that is not an assertion (a build error) is not RED either.
    let build = "sh -c 'echo could not compile; exit 101'";
    assert!(refused(&tdd(&root, &["red", "--check", build])).contains("assertion failures only"));
    assert!(!record_path(&root).exists());
    assert!(log_events(&root).is_empty());
    // All three: the cycle starts in GREEN with the RED commit's test locked.
    let r = ok(&tdd(&root, &["red", "--check", CHECK]));
    assert_eq!(r["phase"], json!("green"));
    assert_eq!(r["record"]["locked"], json!(["tests/login_test.rs"]));
    assert_eq!(r["record"]["spec"], json!("SPEC-20261117-demo"));
    // A second RED while the cycle is open is refused.
    assert!(refused(&tdd(&root, &["red", "--check", CHECK])).contains("already open"));
    // Not a SPEC work branch: refused.
    let (_d2, other) = repo("main");
    red_commit(&other);
    assert!(refused(&tdd(&other, &["red", "--check", CHECK])).contains("SPEC work branch"));
}

/// AC-2 (REQ-3, REQ-4, REQ-12): in GREEN the RED tests and new test files are refused, other
/// writes (code, an existing unrelated test) pass; without a record nothing is locked.
#[test]
fn locked_tests_refuse_other_writes_pass() {
    let (_d, root) = repo(BRANCH);
    red_commit(&root);
    assert!(allowed(&preflight(
        &root,
        &write_input("tests/login_test.rs", &root)
    )));
    ok(&tdd(&root, &["red", "--check", CHECK]));
    let r = preflight(&root, &write_input("tests/login_test.rs", &root));
    assert!(denied(&r), "{}", r.stdout);
    assert!(r.stdout.contains("moai tdd reopen"), "{}", r.stdout);
    let absolute = root.join("tests/login_test.rs");
    assert!(denied(&preflight(
        &root,
        &write_input(absolute.to_str().unwrap(), &root)
    )));
    assert!(denied(&preflight(
        &root,
        &write_input("tests/new_test.rs", &root)
    )));
    assert!(denied(&preflight(
        &root,
        &write_input("src/new.test.ts", &root)
    )));
    assert!(allowed(&preflight(
        &root,
        &write_input("src/lib.txt", &root)
    )));
    assert!(allowed(&preflight(
        &root,
        &write_input("tests/old_test.rs", &root)
    )));
    // A corrupt record locks nothing, and status reads no phase.
    fs::write(record_path(&root), "{oops").unwrap();
    assert!(allowed(&preflight(
        &root,
        &write_input("tests/login_test.rs", &root)
    )));
    assert_eq!(ok(&tdd(&root, &["status"]))["phase"], json!(null));
}

/// AC-3 (REQ-5, REQ-6): GREEN waits for every RED check to pass, then REFACTOR keeps the lock.
#[test]
fn green_needs_every_red_test_passing() {
    let (_d, root) = repo(BRANCH);
    red_commit(&root);
    ok(&tdd(&root, &["red", "--check", CHECK]));
    assert!(refused(&tdd(&root, &["green"])).contains("stays in GREEN"));
    assert_eq!(ok(&tdd(&root, &["status"]))["phase"], json!("green"));
    fs::write(root.join("src/lib.txt"), "fixed\n").unwrap();
    assert_eq!(ok(&tdd(&root, &["green"]))["phase"], json!("refactor"));
    assert!(refused(&tdd(&root, &["green"])).contains("already in REFACTOR"));
    assert!(denied(&preflight(
        &root,
        &write_input("tests/login_test.rs", &root)
    )));
}

/// Audit N2: a record removed outside `moai tdd` (the log still ends in `red` or `green`) is
/// reported by `status` and locks nothing; after `done` or `reopen` there is no report.
#[test]
fn status_reports_an_open_log_without_a_record() {
    let (_d, root) = repo(BRANCH);
    red_commit(&root);
    ok(&tdd(&root, &["red", "--check", CHECK]));
    assert_eq!(ok(&tdd(&root, &["status"]))["problem"], json!(null));
    fs::remove_file(record_path(&root)).unwrap();
    let s = ok(&tdd(&root, &["status"]));
    assert_eq!(s["phase"], json!(null));
    let problem = s["problem"].as_str().unwrap_or_default();
    assert!(
        problem.contains("`red`") && problem.contains("missing or unreadable"),
        "{s}"
    );
    assert!(allowed(&preflight(
        &root,
        &write_input("tests/login_test.rs", &root)
    )));
    // An unreadable record after `green` is reported the same way.
    fs::write(root.join("src/lib.txt"), "fixed\n").unwrap();
    fs::write(root.join("tests/login_test.rs"), "assert fixed 2\n").unwrap();
    commit(&root, "test: again\n\nMoai-Phase: red");
    fs::write(root.join("src/lib.txt"), "broken\n").unwrap();
    ok(&tdd(&root, &["red", "--check", CHECK]));
    fs::write(root.join("src/lib.txt"), "fixed\n").unwrap();
    ok(&tdd(&root, &["green"]));
    fs::write(record_path(&root), "{oops").unwrap();
    let problem = ok(&tdd(&root, &["status"]))["problem"].clone();
    let names_green = problem.as_str().unwrap_or_default().contains("`green`");
    assert!(names_green, "{problem}");
    // A log that ends in `reopen` is not a problem.
    let (_d2, other) = repo(BRANCH);
    red_commit(&other);
    ok(&tdd(&other, &["red", "--check", CHECK]));
    ok(&tdd(&other, &["reopen", "--reason", "x"]));
    assert_eq!(ok(&tdd(&other, &["status"]))["problem"], json!(null));
}

/// Audit N4 (the auditor's mutant M2): a record whose phase is not `green` or `refactor`
/// (`done`, `red`, an unknown word) locks nothing and reads as no cycle.
#[test]
fn a_record_with_another_phase_locks_nothing() {
    let (_d, root) = repo(BRANCH);
    red_commit(&root);
    ok(&tdd(&root, &["red", "--check", CHECK]));
    let mut record: Value = serde_json::from_slice(&fs::read(record_path(&root)).unwrap()).unwrap();
    for (phase, locks) in [
        ("refactor", true),
        ("done", false),
        ("red", false),
        ("GREEN", false),
        ("", false),
    ] {
        record["phase"] = json!(phase);
        fs::write(record_path(&root), record.to_string()).unwrap();
        let r = preflight(&root, &write_input("tests/login_test.rs", &root));
        assert_eq!(denied(&r), locks, "{phase:?}: {}", r.stdout);
        assert_eq!(allowed(&r), !locks, "{phase:?}: {}", r.stdout);
        let expected = if locks { json!(phase) } else { json!(null) };
        assert_eq!(ok(&tdd(&root, &["status"]))["phase"], expected, "{phase:?}");
    }
}

/// AC-4 (REQ-7, REQ-8): reopen needs a reason, logs it and ends the lock; done ends the cycle
/// and keeps the log; with no cycle both are refused.
#[test]
fn reopen_and_done_keep_the_log() {
    let (_d, root) = repo(BRANCH);
    assert!(refused(&tdd(&root, &["done"])).contains("no open TDD cycle"));
    assert!(refused(&tdd(&root, &["reopen", "--reason", "x"])).contains("no open TDD cycle"));
    red_commit(&root);
    ok(&tdd(&root, &["red", "--check", CHECK]));
    for bad in ["", "   ", &"r".repeat(501)] {
        assert!(refused(&tdd(&root, &["reopen", "--reason", bad])).contains("--reason"));
    }
    ok(&tdd(&root, &["reopen", "--reason", &"r".repeat(500)]));
    let r = ok(&tdd(&root, &["status"]));
    assert_eq!(r["phase"], json!(null));
    assert_eq!(r["log"][1]["reason"], json!("r".repeat(500)));
    assert!(allowed(&preflight(
        &root,
        &write_input("tests/login_test.rs", &root)
    )));
    // A new RED, then done.
    fs::write(root.join("tests/login_test.rs"), "assert fixed again\n").unwrap();
    commit(&root, "test: again\n\nMoai-Phase: red");
    ok(&tdd(&root, &["red", "--check", CHECK]));
    ok(&tdd(&root, &["done"]));
    assert!(!record_path(&root).exists());
    assert_eq!(log_events(&root), ["red", "reopen", "red", "done"]);
}

/// Links are refused: a record folder that is a symbolic link stops `red`, and a record that is
/// a link is not read (no lock).
#[test]
fn a_linked_record_folder_or_record_is_refused() {
    let (_d, root) = repo(BRANCH);
    red_commit(&root);
    let elsewhere = tempfile::TempDir::new().unwrap();
    std::os::unix::fs::symlink(elsewhere.path(), root.join(".git/moai")).unwrap();
    assert!(refused(&tdd(&root, &["red", "--check", CHECK])).contains("symbolic link"));
    assert!(fs::read_dir(elsewhere.path()).unwrap().next().is_none());
    fs::remove_file(root.join(".git/moai")).unwrap();
    ok(&tdd(&root, &["red", "--check", CHECK]));
    let real = elsewhere.path().join("tdd.json");
    fs::rename(record_path(&root), &real).unwrap();
    std::os::unix::fs::symlink(&real, record_path(&root)).unwrap();
    assert!(allowed(&preflight(
        &root,
        &write_input("tests/login_test.rs", &root)
    )));
    assert!(refused(&tdd(&root, &["done"])).contains("no open TDD cycle"));
}

/// A5 runs before A6: with the SPEC gone, a write to a locked test gets the SPEC-first reason;
/// with it back, the lock's.
#[test]
fn spec_before_code_is_judged_before_the_test_lock() {
    let (_d, root) = repo(BRANCH);
    red_commit(&root);
    ok(&tdd(&root, &["red", "--check", CHECK]));
    fs::remove_file(root.join(SPEC)).unwrap();
    let r = preflight(&root, &write_input("tests/login_test.rs", &root));
    let spec_first = denied(&r) && r.stdout.contains("moai spec new");
    assert!(spec_first, "{}", r.stdout);
    assert!(!r.stdout.contains("moai tdd reopen"), "{}", r.stdout);
    fs::write(root.join(SPEC), SPEC_TEXT).unwrap();
    let r = preflight(&root, &write_input("tests/login_test.rs", &root));
    let locked = denied(&r) && r.stdout.contains("moai tdd reopen");
    assert!(locked, "{}", r.stdout);
}

/// The record lives in the linked worktree's own git folder: its cycle does not lock the
/// primary checkout.
#[test]
fn a_linked_worktree_keeps_its_own_cycle() {
    let (_d, main) = repo("main");
    let parent = tempfile::TempDir::new().unwrap();
    let wt = fs::canonicalize(parent.path()).unwrap().join("wt");
    git(
        &main,
        &["worktree", "add", "-q", "-b", BRANCH, wt.to_str().unwrap()],
    );
    fs::create_dir_all(wt.join("home")).unwrap();
    red_commit(&wt);
    ok(&tdd(&wt, &["red", "--check", CHECK]));
    assert!(denied(&preflight(
        &wt,
        &write_input("tests/login_test.rs", &wt)
    )));
    assert!(allowed(&preflight(
        &main,
        &write_input("tests/new_test.rs", &main)
    )));
    assert!(!main.join(".git/moai/tdd.json").exists());
    // #3177 audit N1: the written path's own work tree decides, not the session's folder. A
    // session in the primary checkout writing into the GREEN worktree is refused (a locked test
    // and a new test file); a session in the worktree writing into the primary is not.
    for target in ["tests/login_test.rs", "tests/another_test.rs"] {
        let into_wt = wt.join(target).display().to_string();
        let r = preflight(&main, &write_input(&into_wt, &main));
        assert!(denied(&r), "{target}: {}", r.stdout);
    }
    let into_main = main.join("tests/new_test.rs").display().to_string();
    assert!(allowed(&preflight(&wt, &write_input(&into_main, &wt))));
}
