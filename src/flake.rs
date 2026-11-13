//! Flaky test classification (Issue #2666 F1): a test that failed in one CI run and passed in a
//! later run of the same head and job is flaky (the VERIFY workflow's rule). Input is the text
//! `gh run view <id> --log` prints — one line per log line, `<job>\t<step>\t<timestamp> <text>`
//! — holding cargo-nextest output. Nothing here reads the network or writes anything.

use serde::Serialize;
use std::collections::BTreeMap;

/// Statuses cargo-nextest prints for a test that did not pass.
const FAILING: [&str; 5] = ["FAIL", "SIGSEGV", "SIGABRT", "SIGKILL", "TIMEOUT"];

/// What one log says about one test of one job.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct Outcome {
    pub failed: bool,
    pub passed: bool,
    /// The panic location line and the line after it, when the log has them.
    pub message: Option<String>,
}

/// One classified test.
#[derive(Clone, Debug, PartialEq, Serialize)]
pub struct Row {
    pub job: String,
    pub test: String,
    pub message: Option<String>,
}

#[derive(Clone, Debug, Default, PartialEq, Serialize)]
pub struct Classification {
    /// Failed in the first log and passed in the later one (a passing retry counts).
    pub flaky: Vec<Row>,
    /// Failed in both: a real failure, not flaky.
    pub still_failing: Vec<Row>,
    /// Failed in the first log and absent from the later one: undecided.
    pub not_in_later: Vec<Row>,
}

/// `(job, text)` of one `gh run view --log` line; a line without the two tabs is all text.
fn split_line(line: &str) -> (&str, &str) {
    let mut parts = line.splitn(3, '\t');
    match (parts.next(), parts.next(), parts.next()) {
        (Some(job), Some(_step), Some(rest)) => {
            // `<timestamp> <text>`: the timestamp ends with `Z`.
            let text = match rest.split_once(' ') {
                Some((ts, text)) if ts.ends_with('Z') => text,
                _ => rest,
            };
            (job, text)
        }
        _ => ("", line),
    }
}

/// `(status, test)` of a nextest status line `STATUS [ time] ( n/N) <binary> <test>`.
fn status_line(text: &str) -> Option<(&str, String)> {
    let t = text.trim();
    let (status, rest) = t.split_once(' ')?;
    if status != "PASS" && !FAILING.contains(&status) {
        return None;
    }
    let rest = rest.trim_start().strip_prefix('[')?;
    let rest = &rest[rest.find(']')? + 1..];
    let rest = rest.trim_start().strip_prefix('(')?;
    let test = rest[rest.find(')')? + 1..].trim();
    (test.contains(' ') || test.contains("::")).then(|| (status, test.to_string()))
}

/// The test function a nextest id names: the part after the binary (`moai-cli::x name` or
/// `moai-cli mod::tests::name`).
fn function(test: &str) -> &str {
    test.split_once(' ').map_or(test, |(_, name)| name)
}

/// Every test outcome in one log, keyed by `(job, test)`.
pub fn parse(log: &str) -> BTreeMap<(String, String), Outcome> {
    let mut out: BTreeMap<(String, String), Outcome> = BTreeMap::new();
    let mut panics: Vec<(String, String, String)> = Vec::new(); // (job, function, message)
    let mut pending: Option<(String, String, String)> = None;
    for line in log.lines() {
        let (job, text) = split_line(line);
        if let Some((pj, name, loc)) = pending.take() {
            if pj == job {
                panics.push((pj, name, format!("{loc} {}", text.trim())));
                continue;
            }
            panics.push((pj, name, loc));
        }
        if let Some((status, test)) = status_line(text) {
            let o = out.entry((job.to_string(), test)).or_default();
            if status == "PASS" {
                o.passed = true;
            } else {
                o.failed = true;
            }
            continue;
        }
        let t = text.trim();
        if let Some(rest) = t.strip_prefix("thread '") {
            if let (Some(end), Some(at)) = (rest.find('\''), rest.find(" panicked at ")) {
                let name = rest[..end].to_string();
                let loc = rest[at + 1..].trim_end_matches(':').to_string();
                pending = Some((job.to_string(), name, loc));
            }
        }
    }
    if let Some(p) = pending {
        panics.push(p);
    }
    for ((job, test), o) in out.iter_mut() {
        if o.failed {
            o.message = panics
                .iter()
                .find(|(pj, name, _)| pj == job && name == function(test))
                .map(|(_, _, m)| m.clone());
        }
    }
    out
}

/// Classifies every test that failed in `first` against the later log `later`, for one job
/// when `job` is set (matched exactly), else for every job.
pub fn classify(first: &str, later: &str, job: Option<&str>) -> Classification {
    let a = parse(first);
    let b = parse(later);
    let mut c = Classification::default();
    for ((j, test), o) in a
        .iter()
        .filter(|((j, _), o)| o.failed && job.is_none_or(|w| w == j))
    {
        let row = Row {
            job: j.clone(),
            test: test.clone(),
            message: o.message.clone(),
        };
        match b.get(&(j.clone(), test.clone())) {
            // A pass in the later log (a retry that passed included) is the later success.
            Some(later) if later.passed => c.flaky.push(row),
            Some(later) if later.failed => c.still_failing.push(row),
            _ => c.not_in_later.push(row),
        }
    }
    c
}
