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
    /// False when stdout or stderr could not be read within the grace period after the command
    /// ended; their digests are null then. None when nothing was spawned.
    pub output_complete: Option<bool>,
}

/// Splits on ASCII whitespace; single and double quotes group words and are removed, the other
/// quote is literal inside them, and adjacent parts join into one word. No escapes, no
/// expansion. An unterminated quote is refused (#2392 decision 6774899518 row 18: simple
/// quotes now, a structured argv input later).
fn split_words(command: &str) -> Result<Vec<String>> {
    let mut words = Vec::new();
    let mut word = String::new();
    let mut in_word = false;
    let mut quote: Option<char> = None;
    for ch in command.chars() {
        match quote {
            Some(q) if ch == q => quote = None,
            Some(_) => word.push(ch),
            None if ch == '\'' || ch == '"' => {
                quote = Some(ch);
                in_word = true;
            }
            None if ch.is_ascii_whitespace() => {
                if in_word {
                    words.push(std::mem::take(&mut word));
                    in_word = false;
                }
            }
            None => {
                word.push(ch);
                in_word = true;
            }
        }
    }
    if quote.is_some() {
        return Err("command not runnable: unterminated quote".into());
    }
    if in_word {
        words.push(word);
    }
    Ok(words)
}

/// Splits a snapshot command into argv. No shell: a leading `VAR=value`, an option before the
/// executable, and an executable with a path separator are refused (#2003 leader decision I1),
/// judged on the executable word after quote removal.
pub fn parse_argv(command: &str) -> Result<Vec<String>> {
    let argv = split_words(command)?;
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
/// Cargo first, then Python (pytest, unittest): a test module that cannot be imported tested
/// nothing (#2392 decision 6774899518 row 20 = A; other runners stay `Unclassified`).
const ENVIRONMENT_SIGNS: [&str; 10] = [
    "could not compile",
    "error[e",
    "failed to download",
    "failed to get `",
    "could not resolve host",
    "spurious network error",
    "failed to load manifest",
    "no such command",
    "importerror",
    "modulenotfounderror",
];
/// Words that mark an assertion failure in the tests: cargo, then Python's `AssertionError`
/// (unittest, and pytest without assertion rewriting).
const TEST_SIGNS: [&str; 3] = ["test result: failed", "panicked at", "assertionerror"];

/// pytest's rewritten assertion: an `E` detail line whose text starts with `assert`.
fn pytest_assert_line(lower: &str) -> bool {
    lower.lines().any(|l| {
        l.strip_prefix('e')
            .is_some_and(|rest| rest.starts_with(' ') && rest.trim_start().starts_with("assert"))
    })
}

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
    } else if TEST_SIGNS.iter().any(|s| lower.contains(s)) || pytest_assert_line(&lower) {
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
        output_complete: None,
    }
}

/// How long the output readers may take after the command ended (or was ended).
const OUTPUT_GRACE: Duration = Duration::from_secs(2);

/// Ends every process left in the command's process group (unix: `kill -KILL -- -<pgid>`, as
/// argv). An empty group makes `kill` fail, which is fine. Without process groups, nothing.
fn end_group(pid: u32) {
    #[cfg(unix)]
    {
        let _ = Command::new("kill")
            .args(["-KILL", "--", &format!("-{pid}")])
            .stdin(Stdio::null())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status();
    }
    #[cfg(not(unix))]
    let _ = pid;
}

// @MX:WARN: [AUTO] spawns a process group and two reader threads; the group is killed when the
// command ends or times out, and the readers are waited for at most OUTPUT_GRACE
// @MX:REASON: a grandchild holding the pipes kept an unbounded join waiting past the timeout
// (agent-2 FAIL on 5413e0221)
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
    let mut cmd = Command::new(&argv[0]);
    cmd.args(&argv[1..])
        .current_dir(dir)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    // Its own process group, so everything the command starts can be ended with it.
    #[cfg(unix)]
    std::os::unix::process::CommandExt::process_group(&mut cmd, 0);
    let spawned = cmd.spawn();
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
    // Readers hand their bytes over a channel, so the wait for them can be bounded.
    let drain = |pipe: Option<Box<dyn Read + Send>>| {
        let (tx, rx) = std::sync::mpsc::channel();
        std::thread::spawn(move || {
            let mut buf = Vec::new();
            if let Some(mut p) = pipe {
                let _ = p.read_to_end(&mut buf);
            }
            let _ = tx.send(buf);
        });
        rx
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
    // Whatever the command left running (background children holding the pipes) ends with it.
    end_group(child.id());
    let stdout = out.recv_timeout(OUTPUT_GRACE).ok();
    let stderr = err.recv_timeout(OUTPUT_GRACE).ok();
    rec.output_complete = Some(stdout.is_some() && stderr.is_some());
    rec.finished_unix_ms = Some(now_ms());
    rec.duration_ms = Some(started.elapsed().as_millis() as u64);
    rec.exit_code = status.and_then(|s| s.code());
    rec.stdout_sha256 = stdout.as_deref().map(digest);
    rec.stderr_sha256 = stderr.as_deref().map(digest);
    let (stdout, stderr) = (stdout.unwrap_or_default(), stderr.unwrap_or_default());
    let text = format!(
        "{}\n{}",
        String::from_utf8_lossy(&stdout),
        String::from_utf8_lossy(&stderr)
    );
    (rec.tests_passed, rec.tests_failed) = test_counts(&text);
    let success = status.is_some_and(|s| s.success());
    rec.failure_class = classify(None, rec.timed_out, success, &text);
    // A test command whose output says it ran zero tests tested nothing: `not_run`, not a pass
    // (#2392 decision 6774899518 row 19 = A). Only cargo `test result:` lines are counted; a
    // command without them keeps its exit-code result.
    let ran_zero = rec.tests_passed == Some(0) && rec.tests_failed == Some(0);
    rec.result = if success && !rec.timed_out && ran_zero {
        CheckResult::NotRun
    } else if success && !rec.timed_out {
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
