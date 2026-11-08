//! Trusted runner for VERIFY (Issue #2003, 05-C).
//!
//! Runs one verification command as an argv array (never through a shell), with a timeout,
//! and records what happened: start and end, exit status, timeout, the sha256 of stdout and
//! stderr (the output itself is not kept), and a failure class. Nothing is written to disk.

use crate::digest;
use serde::{Deserialize, Serialize};
use std::io::Read;
use std::path::Path;
use std::process::{Command, Stdio};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

pub type Result<T> = std::result::Result<T, String>;

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "kebab-case")]
pub enum FailureClass {
    /// The check could not test anything: build failure, missing command, timeout, network.
    EnvironmentFailure,
    /// An assertion in the tests failed.
    TestFailure,
    /// Failed without a recognised sign either way; never counted as RED.
    Unclassified,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CheckResult {
    Passed,
    Failed,
    NotRun,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Toolchain {
    pub rustc: Option<String>,
    pub cargo: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CheckRun {
    /// Index into the approved snapshot's `verification_commands`.
    pub command_ref: usize,
    pub command: String,
    pub argv: Vec<String>,
    pub result: CheckResult,
    pub exit_code: Option<i32>,
    pub timed_out: bool,
    /// No cancellation path exists yet; recorded so a later one has a place.
    pub cancelled: bool,
    pub spawn_error: Option<String>,
    pub started_unix_ms: Option<u64>,
    pub finished_unix_ms: Option<u64>,
    pub duration_ms: Option<u64>,
    pub stdout_sha256: Option<String>,
    pub stderr_sha256: Option<String>,
    pub failure_class: Option<FailureClass>,
    /// Arguments after `--`: the test selection the command asked for.
    pub test_filters: Vec<String>,
    /// Sums of the `test result:` lines, when the output has any.
    pub tests_passed: Option<u64>,
    pub tests_failed: Option<u64>,
}

/// Splits a snapshot command into argv. No shell: a leading `VAR=value`, an option before the
/// executable, and an executable with a path separator are refused (#2003 leader decision I1).
pub fn parse_argv(command: &str) -> Result<Vec<String>> {
    let argv: Vec<String> = command.split_ascii_whitespace().map(String::from).collect();
    let Some(exe) = argv.first() else {
        return Err("command not runnable: empty".into());
    };
    if exe.contains('=') {
        return Err("command not runnable: environment assignment before the executable".into());
    }
    if exe.starts_with('-') {
        return Err("command not runnable: option before the executable".into());
    }
    if exe.contains('/') || exe.contains('\\') {
        return Err("command not runnable: executable given as a path".into());
    }
    Ok(argv)
}

/// Words that mark a failure of the environment rather than of the code under test.
const ENVIRONMENT_SIGNS: [&str; 8] = [
    "could not compile",
    "error[e",
    "failed to download",
    "failed to get `",
    "could not resolve host",
    "spurious network error",
    "failed to load manifest",
    "no such command",
];
/// Words that mark an assertion failure in the tests.
const TEST_SIGNS: [&str; 2] = ["test result: failed", "panicked at"];

/// Failure class of a finished check; `None` when it passed. Environment signs win over test
/// signs, so a run that could not build is never counted as RED. Text-based (libtest JSON is
/// unstable), so a heuristic.
pub fn classify(
    spawn_error: Option<&str>,
    timed_out: bool,
    success: bool,
    output: &str,
) -> Option<FailureClass> {
    if spawn_error.is_some() || timed_out {
        return Some(FailureClass::EnvironmentFailure);
    }
    if success {
        return None;
    }
    let lower = output.to_ascii_lowercase();
    Some(if ENVIRONMENT_SIGNS.iter().any(|s| lower.contains(s)) {
        FailureClass::EnvironmentFailure
    } else if TEST_SIGNS.iter().any(|s| lower.contains(s)) {
        FailureClass::TestFailure
    } else {
        FailureClass::Unclassified
    })
}

/// Sums `test result: … N passed; M failed` lines.
fn test_counts(output: &str) -> (Option<u64>, Option<u64>) {
    let mut found = false;
    let (mut passed, mut failed) = (0, 0);
    for line in output
        .lines()
        .filter(|l| l.trim_start().starts_with("test result:"))
    {
        found = true;
        let words: Vec<&str> = line
            .split([' ', ';', '.'])
            .filter(|w| !w.is_empty())
            .collect();
        for pair in words.windows(2) {
            match (pair[0].parse::<u64>(), pair[1]) {
                (Ok(n), "passed") => passed += n,
                (Ok(n), "failed") => failed += n,
                _ => {}
            }
        }
    }
    if found {
        (Some(passed), Some(failed))
    } else {
        (None, None)
    }
}

fn now_ms() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_or(0, |d| d.as_millis() as u64)
}

/// A declared check that this request did not run.
pub fn not_run(command_ref: usize, command: &str) -> CheckRun {
    CheckRun {
        command_ref,
        command: command.into(),
        argv: parse_argv(command).unwrap_or_default(),
        result: CheckResult::NotRun,
        exit_code: None,
        timed_out: false,
        cancelled: false,
        spawn_error: None,
        started_unix_ms: None,
        finished_unix_ms: None,
        duration_ms: None,
        stdout_sha256: None,
        stderr_sha256: None,
        failure_class: None,
        test_filters: Vec::new(),
        tests_passed: None,
        tests_failed: None,
    }
}

// @MX:WARN: [AUTO] spawns a child process and two reader threads; the child is killed on timeout
// @MX:REASON: pipes must be drained while waiting, or a chatty child blocks and looks like a hang
/// Runs one command in `dir` with the inherited environment and returns its record. A command
/// that does not parse is recorded as an environment failure without spawning anything.
pub fn run_check(dir: &Path, command_ref: usize, command: &str, timeout: Duration) -> CheckRun {
    let mut rec = not_run(command_ref, command);
    rec.result = CheckResult::Failed;
    let Ok(argv) = parse_argv(command) else {
        rec.spawn_error = Some("not-runnable".into());
        rec.failure_class = Some(FailureClass::EnvironmentFailure);
        return rec;
    };
    rec.test_filters = argv
        .iter()
        .position(|a| a == "--")
        .map(|i| argv[i + 1..].to_vec())
        .unwrap_or_default();
    rec.argv = argv.clone();
    let started = Instant::now();
    rec.started_unix_ms = Some(now_ms());
    let spawned = Command::new(&argv[0])
        .args(&argv[1..])
        .current_dir(dir)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn();
    let mut child = match spawned {
        Ok(c) => c,
        Err(e) => {
            rec.spawn_error = Some(
                match e.kind() {
                    std::io::ErrorKind::NotFound => "not-found",
                    std::io::ErrorKind::PermissionDenied => "permission-denied",
                    _ => "spawn-failed",
                }
                .into(),
            );
            rec.finished_unix_ms = Some(now_ms());
            rec.duration_ms = Some(started.elapsed().as_millis() as u64);
            rec.failure_class = Some(FailureClass::EnvironmentFailure);
            return rec;
        }
    };
    let drain = |pipe: Option<Box<dyn Read + Send>>| {
        std::thread::spawn(move || {
            let mut buf = Vec::new();
            if let Some(mut p) = pipe {
                let _ = p.read_to_end(&mut buf);
            }
            buf
        })
    };
    let out = drain(
        child
            .stdout
            .take()
            .map(|p| Box::new(p) as Box<dyn Read + Send>),
    );
    let err = drain(
        child
            .stderr
            .take()
            .map(|p| Box::new(p) as Box<dyn Read + Send>),
    );
    let status = loop {
        match child.try_wait() {
            Ok(Some(s)) => break Some(s),
            Ok(None) if started.elapsed() >= timeout => {
                let _ = child.kill();
                let _ = child.wait();
                rec.timed_out = true;
                break None;
            }
            Ok(None) => std::thread::sleep(Duration::from_millis(10)),
            Err(_) => break None,
        }
    };
    let stdout = out.join().unwrap_or_default();
    let stderr = err.join().unwrap_or_default();
    rec.finished_unix_ms = Some(now_ms());
    rec.duration_ms = Some(started.elapsed().as_millis() as u64);
    rec.exit_code = status.and_then(|s| s.code());
    rec.stdout_sha256 = Some(digest(&stdout));
    rec.stderr_sha256 = Some(digest(&stderr));
    let text = format!(
        "{}\n{}",
        String::from_utf8_lossy(&stdout),
        String::from_utf8_lossy(&stderr)
    );
    (rec.tests_passed, rec.tests_failed) = test_counts(&text);
    let success = status.is_some_and(|s| s.success());
    rec.failure_class = classify(None, rec.timed_out, success, &text);
    rec.result = if success && !rec.timed_out {
        CheckResult::Passed
    } else {
        CheckResult::Failed
    };
    rec
}

/// `rustc -V` and `cargo -V` as seen from `dir` (a project toolchain file applies there).
pub fn toolchain(dir: &Path) -> Toolchain {
    let first_line = |program: &str| {
        Command::new(program)
            .arg("-V")
            .current_dir(dir)
            .stdin(Stdio::null())
            .output()
            .ok()
            .filter(|o| o.status.success())
            .and_then(|o| String::from_utf8(o.stdout).ok())
            .and_then(|s| s.lines().next().map(|l| l.trim().to_string()))
    };
    Toolchain {
        rustc: first_line("rustc"),
        cargo: first_line("cargo"),
    }
}
