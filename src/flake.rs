//! Flaky test classification (Issue #2666 F1): a test that failed in one CI run and passed in a
//! later run of the same head and job is flaky (the VERIFY workflow's rule). Input is the text
//! `gh run view <id> --log` prints — one line per log line, `<job>\t<step>\t<timestamp> <text>`
//! — holding cargo-nextest output. Nothing here reads the network or writes anything.
//!
//! The head of each job is read from the log too ([`heads`]): a pull-request run checks out a
//! merge commit whose checkout line names the PR head (`HEAD is now at <x> Merge <head> into
//! <base>`); otherwise the SHA `git log -1 --format=%H` prints after the checkout. A pass on
//! another head is not flakiness — the code may have been fixed — so a failed test that passed
//! later is `flaky` only when both heads of its job are read and equal (audit 6858491683 B1).

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
    /// Failed in the first log and passed in the later one, on the same head (a passing retry
    /// counts).
    pub flaky: Vec<Row>,
    /// Failed first, passed later, but the job's two heads differ: a fix, not flakiness.
    pub head_mismatch: Vec<Row>,
    /// Failed first, passed later, but a head could not be read from one of the logs.
    pub head_unverified: Vec<Row>,
    /// Failed in both: a real failure, not flaky.
    pub still_failing: Vec<Row>,
    /// Failed in the first log and absent from the later one: undecided.
    pub not_in_later: Vec<Row>,
    /// The head each job ran on, per log.
    pub first_heads: BTreeMap<String, String>,
    pub later_heads: BTreeMap<String, String>,
    /// Status lines read per log: zero means the log was not cargo-nextest output.
    pub status_lines: BTreeMap<String, usize>,
}

fn is_sha(s: &str) -> bool {
    s.len() == 40 && s.bytes().all(|b| b.is_ascii_hexdigit())
}

/// The head each job ran on (see the module doc); a job without a readable head is absent.
pub fn heads(log: &str) -> BTreeMap<String, String> {
    let mut merge_heads: BTreeMap<String, String> = BTreeMap::new();
    let mut checked_out: BTreeMap<String, String> = BTreeMap::new();
    let mut waiting: Option<String> = None; // job whose `git log -1 --format=%H` output is next
    for line in log.lines() {
        let (job, text) = split_line(line);
        let t = text.trim();
        if let Some(w) = waiting.take() {
            if w == job && is_sha(t) {
                checked_out.entry(w).or_insert_with(|| t.to_string());
                continue;
            }
        }
        if t.starts_with("[command]") && t.ends_with(" log -1 --format=%H") {
            waiting = Some(job.to_string());
        } else if let Some(rest) = t.strip_prefix("HEAD is now at ") {
            let words: Vec<&str> = rest.split_whitespace().collect();
            if let [_, "Merge", head, "into", base, ..] = words[..] {
                if is_sha(head) && is_sha(base) {
                    merge_heads
                        .entry(job.to_string())
                        .or_insert_with(|| head.to_string());
                }
            }
        }
    }
    checked_out.extend(merge_heads);
    checked_out
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

/// How many cargo-nextest status lines one log holds.
fn status_lines(log: &str) -> usize {
    log.lines()
        .filter(|l| status_line(split_line(l).1).is_some())
        .count()
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
    let keep = |m: BTreeMap<String, String>| -> BTreeMap<String, String> {
        m.into_iter()
            .filter(|(j, _)| job.is_none_or(|w| w == j))
            .collect()
    };
    let mut c = Classification {
        first_heads: keep(heads(first)),
        later_heads: keep(heads(later)),
        status_lines: [
            ("first".to_string(), status_lines(first)),
            ("later".to_string(), status_lines(later)),
        ]
        .into(),
        ..Classification::default()
    };
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
            Some(later) if later.passed => match (c.first_heads.get(j), c.later_heads.get(j)) {
                (Some(h1), Some(h2)) if h1 == h2 => c.flaky.push(row),
                (Some(_), Some(_)) => c.head_mismatch.push(row),
                _ => c.head_unverified.push(row),
            },
            Some(later) if later.failed => c.still_failing.push(row),
            _ => c.not_in_later.push(row),
        }
    }
    c
}
