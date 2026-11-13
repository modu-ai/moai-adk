//! #2407: the C6 incremental index judged by the FRESHNESS cases of
//! `tests/fixtures/atlas/suites.json` (the fixture is the case authority, #2346 AC-3).
//! Each case gets its own temporary repository: setup → `index-build` → mutation →
//! `index-query`, and the query is compared with the case's `expected`. A second test
//! perturbs each runnable case's expectation and requires the judge to reject it, so a
//! judge that accepts anything cannot pass.
mod common;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
    sync::atomic::{AtomicUsize, Ordering},
};

const SUITES: &str = "tests/fixtures/atlas/suites.json";

fn cases() -> Vec<Value> {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(SUITES);
    let suites: Value = serde_json::from_slice(&fs::read(path).unwrap()).unwrap();
    let suite = suites["suites"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["id"] == "FRESHNESS")
        .expect("FRESHNESS suite")
        .clone();
    suite["cases"].as_array().unwrap().clone()
}

fn with_env(mut command: Command, dir: &Path) -> Command {
    command
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com");
    command
}

fn git(dir: &Path, args: &[&str]) {
    let mut command = with_env(Command::new("git"), dir);
    command.args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"]);
    let status = command.args(args).status().unwrap();
    assert!(status.success(), "git {args:?} failed in {}", dir.display());
}

fn cli_output(dir: &Path, op: &str) -> Output {
    let mut child = with_env(Command::new(crate::common::MOAI_BIN), dir)
        .args([op, "."])
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(br#"{"protocol_version": 1}"#)
        .unwrap();
    child.wait_with_output().unwrap()
}

fn cli(dir: &Path, op: &str) -> Value {
    let out = cli_output(dir, op);
    assert!(
        out.status.success(),
        "{op} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice::<Value>(&out.stdout).unwrap()["result"].clone()
}

fn pointer_gen(dir: &Path) -> Option<String> {
    let bytes = fs::read(dir.join(".moai-cli/index/current.json")).ok()?;
    let v: Value = serde_json::from_slice(&bytes).ok()?;
    v["generation_id"].as_str().map(str::to_string)
}

struct Repo {
    _tmp: tempfile::TempDir,
    main: PathBuf,
    worktrees: Vec<(String, PathBuf)>,
}

impl Repo {
    fn dir(&self, worktree: Option<&str>) -> &Path {
        match worktree {
            None => &self.main,
            Some(name) => {
                &self
                    .worktrees
                    .iter()
                    .find(|(n, _)| n == name)
                    .unwrap_or_else(|| panic!("unknown worktree {name}"))
                    .1
            }
        }
    }
}

fn setup(case: &Value) -> Repo {
    let tmp = tempfile::tempdir().unwrap();
    let main = tmp.path().join("project");
    fs::create_dir_all(&main).unwrap();
    git(&main, &["init", "-q", "--initial-branch=main"]);
    for (path, contents) in case["setup"]["tracked_files"].as_object().unwrap() {
        fs::create_dir_all(main.join(path).parent().unwrap()).unwrap();
        fs::write(main.join(path), contents.as_str().unwrap()).unwrap();
    }
    git(&main, &["add", "."]);
    git(&main, &["commit", "-q", "-m", "init"]);
    let mut worktrees = Vec::new();
    for name in case["setup"]["worktrees"].as_array().into_iter().flatten() {
        let name = name.as_str().unwrap().to_string();
        let path = tmp.path().join(&name);
        git(
            &main,
            &["worktree", "add", "-q", "-b", &name, path.to_str().unwrap()],
        );
        worktrees.push((name, path));
    }
    Repo {
        _tmp: tmp,
        main,
        worktrees,
    }
}

fn mutate(repo: &Repo, m: &Value) {
    let dir = repo.dir(m["worktree"].as_str());
    let path = m["path"].as_str().unwrap_or_default();
    let contents = m["contents"].as_str().unwrap_or_default();
    match m["operation"].as_str().unwrap() {
        "add_tracked" => {
            fs::write(dir.join(path), contents).unwrap();
            git(dir, &["add", path]);
        }
        "add_untracked" | "modify" | "modify_worktree" => {
            fs::write(dir.join(path), contents).unwrap();
        }
        "delete_without_git_rm" => fs::remove_file(dir.join(path)).unwrap(),
        "git_mv" => git(
            dir,
            &["mv", m["from"].as_str().unwrap(), m["to"].as_str().unwrap()],
        ),
        "checkout_new_commit" => {
            assert_eq!(m["same_files"], json!(true), "only same_files is defined");
            git(
                dir,
                &["checkout", "-q", "-b", m["branch"].as_str().unwrap()],
            );
            git(dir, &["commit", "-q", "--allow-empty", "-m", "other"]);
        }
        other => panic!("unknown mutation operation {other}"),
    }
}

/// Compares one directory's query with an expectation (`state`, `changes`,
/// `generation_unchanged`).
fn compare(label: &str, dir: &Path, before: &Option<String>, want: &Value) -> Result<(), String> {
    compare_at(label, dir, before, want, false)
}

fn compare_at(
    label: &str,
    dir: &Path,
    before: &Option<String>,
    want: &Value,
    other_worktree: bool,
) -> Result<(), String> {
    validate_expected_at(want, other_worktree)?;
    let got = cli(dir, "index-query");
    if got["state"] != want["state"] {
        return Err(format!(
            "{label}: state {} != {}",
            got["state"], want["state"]
        ));
    }
    for (key, value) in want["changes"].as_object().into_iter().flatten() {
        let seen = &got["changes"][key];
        if seen != value {
            return Err(format!("{label}: changes.{key} {seen} != {value}"));
        }
    }
    if let Some(expected) = want.get("generation_unchanged") {
        let actual = json!(pointer_gen(dir) == *before);
        if actual != *expected {
            return Err(format!(
                "{label}: generation_unchanged {actual} != {expected}"
            ));
        }
    }
    Ok(())
}

/// Every accepted key must have a real observation in compare() or judge().
fn validate_expected(want: &Value) -> Result<(), String> {
    validate_expected_at(want, false)
}

fn validate_expected_at(want: &Value, other_worktree: bool) -> Result<(), String> {
    for key in want.as_object().ok_or("expected must be an object")?.keys() {
        let supported = matches!(key.as_str(), "state" | "changes" | "generation_unchanged")
            || (!other_worktree && matches!(key.as_str(), "other_worktree" | "cache_scope"))
            || (other_worktree && key == "worktree");
        if !supported {
            return Err(format!("unconsumed expected key: {key}"));
        }
    }
    if let Some(changes) = want.get("changes") {
        for key in changes
            .as_object()
            .ok_or("changes must be an object")?
            .keys()
        {
            if !matches!(
                key.as_str(),
                "added" | "deleted" | "modified" | "new_files" | "branch_switched"
            ) {
                return Err(format!("unconsumed changes key: {key}"));
            }
        }
    }
    if let Some(other) = want.get("other_worktree") {
        validate_expected_at(other, true)?;
    }
    if want.get("cache_scope").is_some_and(|v| v != "per_worktree") {
        return Err("unsupported cache_scope".into());
    }
    Ok(())
}

/// Runs one case end to end; `Err` names the first difference.
fn judge(test: &str, case: &Value) -> Result<Option<String>, String> {
    if case["mutation"]["operation"] == "fail_scan" {
        return match f08_prepare(test, case) {
            Ok(failure) => {
                let observed = failure.observe_failure();
                compare_failure_expected(&observed, &case["expected"])?;
                failure.blocked.restore();
                eprintln!("F-08: native fail_scan and all expected fields checked; gap: the original fixture has no initial code-reference edge; consumer enforcement is not measured");
                Ok(None)
            }
            Err(reason) => Ok(Some(reason)),
        };
    }
    let id = case["case_id"].as_str().unwrap();
    let repo = setup(case);
    let mut before: Vec<(Option<String>, Option<String>)> = Vec::new();
    let mut dirs: Vec<Option<String>> = vec![None];
    dirs.extend(repo.worktrees.iter().map(|(n, _)| Some(n.clone())));
    if case["setup"]["build_index"] == json!(true) {
        for d in &dirs {
            let dir = repo.dir(d.as_deref());
            cli(dir, "index-build");
            before.push((d.clone(), pointer_gen(dir)));
        }
    }
    let gen_of = |name: Option<&str>| {
        before
            .iter()
            .find(|(d, _)| d.as_deref() == name)
            .and_then(|(_, g)| g.clone())
    };
    let m = &case["mutation"];
    mutate(&repo, m);
    let target = m["worktree"].as_str();
    compare(id, repo.dir(target), &gen_of(target), &case["expected"])?;
    if let Some(other) = case["expected"].get("other_worktree") {
        let name = other["worktree"].as_str().unwrap();
        compare_at(
            &format!("{id} ({name})"),
            repo.dir(Some(name)),
            &gen_of(Some(name)),
            other,
            true,
        )?;
    }
    if case["expected"]["cache_scope"] == json!("per_worktree") {
        for (name, path) in &repo.worktrees {
            if !path.join(".moai-cli/index/current.json").is_file() {
                return Err(format!(
                    "{id}: worktree {name} has no index pointer of its own"
                ));
            }
        }
    }
    Ok(None)
}

#[test]
fn freshness_cases_from_the_fixture() {
    let all = cases();
    assert_eq!(all.len(), 8, "the FRESHNESS suite has eight cases");
    let mut not_run = Vec::new();
    for case in &all {
        let id = case["case_id"].as_str().unwrap();
        match judge("freshness_cases_from_the_fixture", case) {
            Ok(None) => eprintln!("passed {id}"),
            Ok(Some(reason)) => {
                eprintln!("not_run {id}: {reason}");
                not_run.push(id.to_string());
            }
            Err(e) => panic!("{e}"),
        }
    }
    assert!(not_run.iter().all(|id| id == "F-08"));
    eprintln!(
        "FRESHNESS: {} passed, {} not_run",
        all.len() - not_run.len(),
        not_run.len()
    );
}

/// #2477 scope 2: the fixture's `c6_status` text states what this file runs. The judge
/// reads no `c6_status`, so only the wording is checked here.
#[test]
fn c6_status_states_that_every_case_is_covered() {
    for case in cases() {
        let id = case["case_id"].as_str().unwrap();
        let status = case["c6_status"].as_str().unwrap();
        assert!(status.starts_with("covered by "), "{id}: {status}");
    }
}

/// #2477 scope 1: a fault that cannot be set up is an error where F-08 must run (Linux,
/// macOS) and a stated not_run elsewhere.
#[test]
fn f08_not_run_is_refused_where_f08_must_run() {
    let err = f08_not_run_allowed(true, "synthetic reason").unwrap_err();
    assert!(err.contains("synthetic reason"), "{err}");
    assert_eq!(f08_not_run_allowed(false, "synthetic reason"), Ok(()));
    assert_eq!(
        F08_REQUIRED,
        cfg!(any(target_os = "linux", target_os = "macos"))
    );
}

/// Every F-08 setup goes through [`f08_prepare`], and [`F08_SITES`] counts them, so the
/// summary line cannot miss a site: the direct calls plus the two `judge` callers, each of
/// which reaches F-08 once through the fixture's single `fail_scan` case.
#[test]
fn f08_sites_match_the_calls_in_this_file() {
    // Whitespace removed, so a call rustfmt splits over lines still counts.
    let source: String = include_str!("freshness_fixture.rs")
        .split_whitespace()
        .collect();
    let gated = concat!("f08_prepare", "(\"");
    let judged = concat!("judge", "(\"");
    let raw = concat!("prepare_failed_scan", "(");
    assert_eq!(
        source.matches(gated).count() + source.matches(judged).count(),
        F08_SITES
    );
    let fail_scan = cases()
        .iter()
        .filter(|c| c["mutation"]["operation"] == "fail_scan")
        .count();
    assert_eq!(fail_scan, 1);
    // The definition and the one call inside `f08_prepare`.
    assert_eq!(source.matches(raw).count(), 2);
}

/// Mutation per case: a wrong expected state must make the judge fail.
#[test]
fn the_judge_rejects_a_wrong_expectation_in_every_runnable_case() {
    for case in cases() {
        let id = case["case_id"].as_str().unwrap().to_string();
        let mut wrong = case.clone();
        let state = wrong["expected"]["state"].as_str().unwrap();
        let flipped = if state == "fresh" { "stale" } else { "fresh" };
        wrong["expected"]["state"] = json!(flipped);
        match judge(
            "the_judge_rejects_a_wrong_expectation_in_every_runnable_case",
            &wrong,
        ) {
            Err(_) => (),
            Ok(Some(reason)) => eprintln!("not_run {id} wrong-expectation check: {reason}"),
            Ok(None) => panic!("{id}: the judge accepted state {flipped}"),
        }
    }
}

#[test]
fn freshness_judge_rejects_unconsumed_expected_keys() {
    let repo = setup(&cases()[0]);
    cli(&repo.main, "index-build");
    let before = pointer_gen(&repo.main);
    for key in ["edge_state", "file_exists", "deleted_files", "typo"] {
        let mut want = json!({"state": "fresh"});
        want[key] = json!(null);
        assert!(
            compare("unconsumed", &repo.main, &before, &want).is_err(),
            "the judge silently accepted {key}"
        );
    }
    assert!(validate_expected(&json!({"state":"fresh", "changes":{"typo":null}})).is_err());
    assert!(validate_expected(&json!({"state":"fresh", "other_worktree":{"typo":null}})).is_err());
    assert!(compare(
        "root metadata",
        &repo.main,
        &before,
        &json!({"state":"fresh", "worktree":"ignored"})
    )
    .is_err());
    assert!(compare(
        "false generation",
        &repo.main,
        &before,
        &json!({"state":"fresh", "generation_unchanged":false})
    )
    .is_err());
    assert!(validate_expected(&json!({"state":"fresh", "other_worktree":{"worktree":"other", "cache_scope":"per_worktree"}})).is_err());
}

fn f08() -> Value {
    cases()
        .into_iter()
        .find(|c| c["case_id"] == "F-08")
        .unwrap()
}

/// chmod is a fixture fault, never a product-only fault-injection switch.
struct BlockedRead {
    path: PathBuf,
    original: fs::Permissions,
}

impl BlockedRead {
    #[cfg(unix)]
    fn new(path: PathBuf) -> Result<Self, String> {
        use std::os::unix::fs::PermissionsExt;
        let original = fs::metadata(&path).unwrap().permissions();
        let mut denied = original.clone();
        denied.set_mode(0o000);
        fs::set_permissions(&path, denied).unwrap();
        let guard = Self { path, original };
        match fs::read(&guard.path) {
            Err(e) if e.kind() == std::io::ErrorKind::PermissionDenied => Ok(guard),
            Ok(_) => Err("mode 000 remains readable (privileged user or filesystem)".into()),
            Err(e) => panic!("unexpected fault setup error: {e}"),
        }
    }

    #[cfg(not(unix))]
    fn new(_path: PathBuf) -> Result<Self, String> {
        Err("Unix permission-based scan failure is unsupported on this platform".into())
    }

    fn restore(&self) {
        fs::set_permissions(&self.path, self.original.clone()).unwrap();
    }
}

impl Drop for BlockedRead {
    fn drop(&mut self) {
        // Explicit restoration is asserted on the success path; also restore on unwind.
        if let Err(e) = fs::set_permissions(&self.path, self.original.clone()) {
            eprintln!("permission restoration failed: {e}");
        }
    }
}

struct FailedScan {
    // Restore before Repo drops its temporary directory, including during unwind.
    blocked: BlockedRead,
    repo: Repo,
    pointer: Vec<u8>,
    generation: String,
    records: Value,
}

fn records(dir: &Path) -> Value {
    moai_cli::index::index_records(
        dir,
        &moai_cli::index::IndexQueryRequest {
            protocol_version: 1,
        },
    )
    .unwrap()
}

/// Platforms whose test runners can raise the mode-000 read fault, so F-08 must run there
/// (#2477). Windows has no mode-based read fault and keeps a stated not_run.
const F08_REQUIRED: bool = cfg!(any(target_os = "linux", target_os = "macos"));
/// The tests that set up the F-08 fault: one [`f08_prepare`] call each.
const F08_SITES: usize = 6;
static F08_RUN: AtomicUsize = AtomicUsize::new(0);
static F08_NOT_RUN: AtomicUsize = AtomicUsize::new(0);
static F08_DONE: AtomicUsize = AtomicUsize::new(0);

/// Writes to the process stderr directly. libtest captures `eprintln!` of a passing test, so
/// only a direct write reaches the rc-ci log as evidence that F-08 ran.
fn evidence(line: &str) {
    let _ = writeln!(std::io::stderr(), "{line}");
}

/// The gate for an F-08 fault that could not be set up: an error where F-08 must run, a
/// stated not_run elsewhere.
fn f08_not_run_allowed(required: bool, reason: &str) -> Result<(), String> {
    if required {
        Err(format!(
            "F-08 must run on this platform but the read fault could not be set up: {reason} (run the tests as an unprivileged user)"
        ))
    } else {
        Ok(())
    }
}

/// Sets up F-08 for `test` and records the outcome in the evidence lines. `Err` is an allowed
/// not_run (with its reason); a not_run where F-08 must run fails the test. The test that
/// completes the last of [`F08_SITES`] prints the summary line.
fn f08_prepare(test: &str, case: &Value) -> Result<FailedScan, String> {
    let outcome = prepare_failed_scan(case);
    match &outcome {
        Ok(_) => {
            F08_RUN.fetch_add(1, Ordering::SeqCst);
            evidence(&format!("F-08 run: {test}"));
        }
        Err(reason) => {
            if let Err(e) = f08_not_run_allowed(F08_REQUIRED, reason) {
                evidence(&format!("F-08 FAILED: {test}: {e}"));
                panic!("{e}");
            }
            F08_NOT_RUN.fetch_add(1, Ordering::SeqCst);
            evidence(&format!("F-08 not_run: {test}: {reason}"));
        }
    }
    if F08_DONE.fetch_add(1, Ordering::SeqCst) + 1 == F08_SITES {
        evidence(&format!(
            "F-08 summary: run={} not_run={} of {F08_SITES} on {} (F-08 required: {F08_REQUIRED})",
            F08_RUN.load(Ordering::SeqCst),
            F08_NOT_RUN.load(Ordering::SeqCst),
            std::env::consts::OS,
        ));
    }
    outcome
}

fn prepare_failed_scan(case: &Value) -> Result<FailedScan, String> {
    assert_eq!(case["mutation"]["operation"], "fail_scan");
    assert_eq!(case["mutation"]["file_removed"], false);
    let repo = setup(case);
    cli(&repo.main, "index-build");
    let pointer = fs::read(repo.main.join(".moai-cli/index/current.json")).unwrap();
    let generation = pointer_gen(&repo.main).unwrap();
    let records = records(&repo.main);
    assert_eq!(records["state"], "fresh");
    let path = repo.main.join(case["mutation"]["path"].as_str().unwrap());
    let blocked = BlockedRead::new(path)?;
    Ok(FailedScan {
        blocked,
        repo,
        pointer,
        generation,
        records,
    })
}

impl FailedScan {
    fn observe_failure(&self) -> Value {
        let dir = &self.repo.main;
        let out = cli_output(dir, "index-build");
        assert!(
            !out.status.success(),
            "the source-read failure must refuse the build"
        );
        let diagnostic = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        assert!(diagnostic.contains("read failed"), "{diagnostic}");
        let got = cli(dir, "index-query");
        let pointer = fs::read(dir.join(".moai-cli/index/current.json")).unwrap();
        assert_eq!(pointer_gen(dir).as_deref(), Some(self.generation.as_str()));
        assert_eq!(got["generation_id"], self.generation);
        assert_eq!(
            pointer, self.pointer,
            "a failed build must not publish a generation"
        );
        assert!(
            got["changes"].is_null(),
            "an unreadable scan must not invent an empty changes list"
        );

        // This observation is git's tracked-deletion list, not fabricated query.changes.
        let mut command = with_env(Command::new("git"), dir);
        let deleted = command
            .args([
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
                "diff",
                "--name-only",
                "--diff-filter=D",
                "-z",
                "HEAD",
                "--",
            ])
            .output()
            .unwrap();
        assert!(deleted.status.success());
        let paths: Vec<&str> = std::str::from_utf8(&deleted.stdout)
            .unwrap()
            .split('\0')
            .filter(|p| !p.is_empty())
            .collect();

        let retained = records(dir);
        assert_eq!(retained["map"], self.records["map"]);
        assert_eq!(retained["graph"], self.records["graph"]);
        // #2407 leader decision 6793757063: availability of THIS failed scan, not
        // deletion of retained historical edges. This is test evidence, not a new
        // product field or proof that every consumer rejects partial records.
        let unavailable = !out.status.success()
            && diagnostic.contains("read failed")
            && got["state"] == "partial"
            && pointer == self.pointer;
        json!({
            "state": got["state"],
            "file_exists": self.blocked.path.is_file(),
            "deleted_files": paths,
            "generation_unchanged": pointer == self.pointer,
            "edge_state": if unavailable { "unavailable" } else { "available" },
        })
    }
}

#[test]
fn f08_failed_scan_preserves_file_and_published_generation() {
    let case = f08();
    // An allowed not_run (Windows) is recorded by `f08_prepare`.
    let Ok(failure) = f08_prepare(
        "f08_failed_scan_preserves_file_and_published_generation",
        &case,
    ) else {
        return;
    };
    compare_failure_expected(&failure.observe_failure(), &case["expected"]).unwrap();
    failure.blocked.restore();
}

#[test]
fn f08_recovery_clears_partial_after_read_access_returns() {
    let case = f08();
    // An allowed not_run (Windows) is recorded by `f08_prepare`.
    let Ok(failure) = f08_prepare(
        "f08_recovery_clears_partial_after_read_access_returns",
        &case,
    ) else {
        return;
    };
    compare_failure_expected(&failure.observe_failure(), &case["expected"]).unwrap();
    failure.blocked.restore();
    assert_eq!(cli(&failure.repo.main, "index-build")["status"], "complete");
    assert_eq!(cli(&failure.repo.main, "index-query")["state"], "fresh");
    assert!(!failure
        .repo
        .main
        .join(".moai-cli/index/last-build.json")
        .exists());
}

#[test]
fn failed_scan_retains_historical_edges_with_partial_freshness() {
    // Separate synthetic setup: the authoritative F-08 fixture and expectations stay intact.
    let mut case = f08();
    case["setup"]["tracked_files"] = json!({
        "src/lib.rs": "pub mod a; pub mod b;\n",
        "src/a.rs": "use crate::b;\n",
        "src/b.rs": "pub fn callee() {}\n"
    });
    case["mutation"]["path"] = json!("src/a.rs");
    // An allowed not_run (Windows) is recorded by `f08_prepare`.
    let Ok(failure) = f08_prepare(
        "failed_scan_retains_historical_edges_with_partial_freshness",
        &case,
    ) else {
        return;
    };
    assert!(
        failure.records["graph"]["edges"]
            .as_array()
            .unwrap()
            .iter()
            .any(|e| e["kind"] == "imports"
                && e["source_id"] == "file:src/a.rs"
                && e["target_id"] == "file:src/b.rs"
                && e["evidence_refs"].as_array().is_some_and(|v| !v.is_empty())),
        "the control must contain a real evidence-backed edge before the failure"
    );
    compare_failure_expected(&failure.observe_failure(), &case["expected"]).unwrap();
    failure.blocked.restore();
}

/// All F-08 fields come from independent observations; unknown or missing keys fail.
fn compare_failure_expected(observed: &Value, want: &Value) -> Result<(), String> {
    let expected = want.as_object().ok_or("F-08 expected must be an object")?;
    for (key, value) in expected {
        let actual = observed
            .get(key)
            .ok_or_else(|| format!("unconsumed F-08 expected key: {key}"))?;
        if actual != value {
            return Err(format!("F-08: {key} {actual} != {value}"));
        }
    }
    for key in observed.as_object().unwrap().keys() {
        if !expected.contains_key(key) {
            return Err(format!("F-08 missing expected key: {key}"));
        }
    }
    Ok(())
}

#[test]
fn f08_judge_rejects_every_wrong_expected_field() {
    let case = f08();
    // An allowed not_run (Windows) is recorded by `f08_prepare`.
    let Ok(failure) = f08_prepare("f08_judge_rejects_every_wrong_expected_field", &case) else {
        return;
    };
    let observed = failure.observe_failure();
    compare_failure_expected(&observed, &case["expected"]).unwrap();
    for (key, wrong) in [
        ("state", json!("fresh")),
        ("file_exists", json!(false)),
        ("deleted_files", json!(["a.rs"])),
        ("edge_state", json!("available")),
        ("generation_unchanged", json!(false)),
        ("typo", Value::Null),
    ] {
        let mut expected = case["expected"].clone();
        expected[key] = wrong;
        assert!(
            compare_failure_expected(&observed, &expected).is_err(),
            "accepted wrong {key}"
        );
    }
    let mut missing = case["expected"].clone();
    missing.as_object_mut().unwrap().remove("edge_state");
    assert!(compare_failure_expected(&observed, &missing).is_err());
    failure.blocked.restore();
}
