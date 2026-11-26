//! The generic local check of a project (SPEC-20261122-project-ci-design REQ-5): it runs the
//! commands the project profile declares, in the declared order, in the project folder, so the
//! local run and the generated workflow are one list. This repository's own `check affected`
//! and `tools/local_check.sh` are not this: they belong to this repository only.
//!
//! Every step runs even when an earlier one failed; a command that hangs, or whose child
//! still holds its output, is stopped at the limit; a project with no declared commands gets a
//! plain explanation, never a green result.
//! The steps are the commands a person confirmed in the profile, run through `sh -c` with
//! git's repository variables taken out of the environment, so a check that runs `git` sees the
//! project's own repository.

use crate::project_ci::CheckEntry;
use serde_json::{json, Value};
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// The default stop for one command, in seconds.
pub const DEFAULT_TIMEOUT_SECS: u64 = 1800;
/// The last bytes kept of each of a step's two output streams: the answer holds the standard
/// output's last bytes and then the standard error's.
const OUTPUT_TAIL: usize = 4096;
/// How long the output of a stopped step may take to close before the run stops waiting for
/// it. Only a child that left the step's process group can hold it that long.
#[cfg(unix)]
const DRAIN_GRACE: std::time::Duration = std::time::Duration::from_secs(2);

pub struct Options {
    /// A step whose command still runs, or whose output is still held open, after this many
    /// seconds is stopped and counted as a failure.
    pub timeout_secs: u64,
}

impl Default for Options {
    fn default() -> Self {
        Options {
            timeout_secs: DEFAULT_TIMEOUT_SECS,
        }
    }
}

/// The profile's declared checks, in the declared order. A project without any gets a plain
/// explanation.
pub fn declared(root: &Path) -> Result<Vec<CheckEntry>> {
    let (profile, _) = crate::profile::read_profile(root)?;
    match profile.checks {
        Some(checks) if !checks.is_empty() => Ok(checks),
        _ => Err(
            "no declared checks in the project profile, so there is nothing to run; confirm the detected check commands first"
                .into(),
        ),
    }
}

/// Runs the declared commands and answers with one entry per step and one summary line.
pub fn run(root: &Path, options: &Options) -> Result<Value> {
    let checks = declared(root)?;
    let mut steps = Vec::new();
    let mut words = Vec::new();
    let mut failed = false;
    for c in &checks {
        let outcome = run_one(root, &c.command, options.timeout_secs);
        if outcome.exit != Some(0) {
            failed = true;
        }
        words.push(format!(
            "{}={}",
            c.id,
            outcome
                .exit
                .map_or("timeout".to_string(), |code| code.to_string())
        ));
        steps.push(json!({
            "id": c.id,
            "command": c.command,
            "exit": outcome.exit,
            "timed_out": outcome.timed_out,
            "seconds": outcome.seconds,
            "output_tail": outcome.tail,
        }));
    }
    let result = if failed { "FAIL" } else { "PASS" };
    Ok(json!({
        "steps": steps,
        "result": result,
        "summary": format!("project-check {} result={result}", words.join(" ")),
    }))
}

struct Outcome {
    /// `None` when the step timed out or was killed by a signal.
    exit: Option<i32>,
    timed_out: bool,
    seconds: u64,
    tail: String,
}

#[cfg(unix)]
fn run_one(root: &Path, command: &str, timeout_secs: u64) -> Outcome {
    use std::os::unix::process::CommandExt;
    use std::process::{Command, Stdio};
    use std::sync::mpsc;
    use std::time::{Duration, Instant};

    let started = Instant::now();
    let failed = |why: String| Outcome {
        exit: Some(127),
        timed_out: false,
        seconds: started.elapsed().as_secs(),
        tail: why,
    };
    let mut cmd = Command::new("sh");
    cmd.arg("-c")
        .arg(command)
        .current_dir(root)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    crate::identity::scrub_git_env(&mut cmd);
    // The step runs in its own process group, so a stop reaches what the shell started.
    cmd.process_group(0);
    let mut child = match cmd.spawn() {
        Ok(c) => c,
        Err(_) => return failed("the command could not be started".into()),
    };
    // Each stream is read on its own thread, which says when the stream has closed. The
    // threads are never joined: a pipe can stay open longer than the step is allowed to run.
    let (closed, closing) = mpsc::channel();
    let streams = [
        child.stdout.take().map(|out| reader(out, closed.clone())),
        child.stderr.take().map(|err| reader(err, closed.clone())),
    ];
    let mut open = streams.iter().flatten().count();
    drop(closed);
    let limit = Duration::from_secs(timeout_secs);
    let tick = Duration::from_millis(20);
    let mut timed_out = false;
    let status = loop {
        // @cc:careful the shell is not reaped while its output is open: until it is, its pid
        // names its process group and no other process can be given that number, so the group
        // stop below cannot reach a stranger.
        if open == 0 {
            match child.try_wait() {
                Ok(Some(s)) => break Some(s),
                Ok(None) => {}
                Err(_) => break None,
            }
        }
        if started.elapsed() >= limit {
            timed_out = true;
            break None;
        }
        if open == 0 {
            std::thread::sleep(tick);
        } else if closing.recv_timeout(tick).is_ok() {
            open -= 1;
        }
    };
    if status.is_none() {
        if let Ok(pid) = i32::try_from(child.id()) {
            // SAFETY: kill(2) on the process group this function's spawn created.
            unsafe { libc::kill(-pid, libc::SIGKILL) };
        }
        let _ = child.kill();
        let _ = child.wait();
        let stopped = Instant::now();
        while open > 0
            && closing
                .recv_timeout(DRAIN_GRACE.saturating_sub(stopped.elapsed()))
                .is_ok()
        {
            open -= 1;
        }
    }
    let mut bytes = Vec::new();
    for kept in streams.iter().flatten() {
        bytes.extend_from_slice(
            &kept
                .lock()
                .unwrap_or_else(std::sync::PoisonError::into_inner),
        );
    }
    let text = String::from_utf8_lossy(&bytes).into_owned();
    Outcome {
        exit: status.and_then(|s| s.code()),
        timed_out,
        seconds: started.elapsed().as_secs(),
        tail: text,
    }
}

#[cfg(not(unix))]
fn run_one(_root: &Path, _command: &str, _timeout_secs: u64) -> Outcome {
    Outcome {
        exit: Some(127),
        timed_out: false,
        seconds: 0,
        tail: "the local check needs a Unix shell".into(),
    }
}

/// Reads `pipe` on a thread of its own, keeping its last [`OUTPUT_TAIL`] bytes in the
/// returned buffer as they arrive, and sends on `closed` once the pipe has ended.
#[cfg(unix)]
fn reader(
    mut pipe: impl std::io::Read + Send + 'static,
    closed: std::sync::mpsc::Sender<()>,
) -> std::sync::Arc<std::sync::Mutex<Vec<u8>>> {
    let kept = std::sync::Arc::new(std::sync::Mutex::new(Vec::new()));
    let into = kept.clone();
    std::thread::spawn(move || {
        let mut buf = [0u8; 4096];
        loop {
            match pipe.read(&mut buf) {
                Ok(0) | Err(_) => break,
                Ok(n) => {
                    let mut kept = into
                        .lock()
                        .unwrap_or_else(std::sync::PoisonError::into_inner);
                    kept.extend_from_slice(&buf[..n]);
                    if kept.len() > OUTPUT_TAIL {
                        let cut = kept.len() - OUTPUT_TAIL;
                        kept.drain(..cut);
                    }
                }
            }
        }
        // The run may have stopped waiting; then nobody receives.
        let _ = closed.send(());
    });
    kept
}

/// The commands of the `check` job's run steps, in order. Another job's steps are not part of
/// the list: `check` is the one job every generated workflow has, and it runs the declared
/// commands in the declared order.
pub fn workflow_run_steps(text: &str) -> Vec<String> {
    let indent = |l: &str| l.len() - l.trim_start().len();
    let mut out = Vec::new();
    let mut in_jobs = false;
    let mut in_check = false;
    let mut block: Option<usize> = None;
    for line in text.lines() {
        if line == "jobs:" {
            in_jobs = true;
            continue;
        }
        if !in_jobs {
            continue;
        }
        if !line.trim().is_empty() && indent(line) == 2 && line.trim_end().ends_with(':') {
            in_check = line.trim() == "check:";
            block = None;
            continue;
        }
        if !in_check {
            continue;
        }
        if let Some(at) = block {
            if line.trim().is_empty() {
                continue;
            }
            if indent(line) > at {
                out.push(line.trim().to_string());
                continue;
            }
            block = None;
        }
        if line.trim_end().ends_with("run: |") {
            block = Some(indent(line));
        }
    }
    out
}

/// The findings of a workflow against the declared commands (REQ-4): a declared command the
/// workflow lacks, a run step that is not declared and a different order are each named. No
/// finding means the two lists are equal.
pub fn compare(declared: &[CheckEntry], workflow: &str) -> Vec<String> {
    let steps = workflow_run_steps(workflow);
    let mut findings = Vec::new();
    for c in declared {
        if !steps.contains(&c.command) {
            findings.push(format!(
                "the declared command `{}` is missing from the workflow",
                c.command
            ));
        }
    }
    for s in &steps {
        if !declared.iter().any(|c| c.command == *s) {
            findings.push(format!(
                "the workflow runs `{s}`, which is not declared in the profile"
            ));
        }
    }
    if findings.is_empty() {
        let want: Vec<&str> = declared.iter().map(|c| c.command.as_str()).collect();
        let have: Vec<&str> = steps.iter().map(String::as_str).collect();
        if want != have {
            findings.push(
                "the workflow runs the declared commands in a different order than the profile declares them"
                    .into(),
            );
        }
    }
    findings
}
