//! SPEC-20261125-check-affected-rerun: a repeated `moai check affected` run on a head whose earlier
//! run failed re-runs only the tests that failed, the test step retries a failing test and lists
//! one that passed on a retry as FLAKY, and a ledger in the project state folder counts the FLAKY
//! names (`moai check flaky`).
//!
//! The earlier run is read from the check folder it left (`summary.txt` and the test step log);
//! nothing else is trusted. Anything that cannot be read, or that names no failed test, gives the
//! full set and says why (`rerun=full reason=<why>`): the full set is the safe default.

use crate::check_affected::{Facts, Pick, Selection, Step};
use std::collections::BTreeSet;
use std::path::{Path, PathBuf};
use std::process::Command;

/// How many times nextest re-runs a failing test (REQ-6).
pub const RETRIES: u32 = 2;

/// The ledger of FLAKY names, one JSON object per line, in the project state folder (REQ-9).
const LEDGER_FILE: &str = "check-flaky.jsonl";

/// The ledger of the project at `root`, in the folder that holds its state now.
pub fn ledger_path(root: &Path) -> PathBuf {
    crate::state_root::path(root, LEDGER_FILE)
}

/// A name seen this many times is marked `fix` by `moai check flaky` (REQ-10).
const FIX_AT: usize = 2;

/// One test of a nextest log: the binary id (`moai-cli::product_cli`) and the test name.
#[derive(Clone, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct TestName {
    pub binary: String,
    pub name: String,
}

impl TestName {
    /// `binary name`, the form the log, `flaky.txt` and the ledger use.
    pub fn text(&self) -> String {
        format!("{} {}", self.binary, self.name)
    }
}

/// The first words of a nextest line that mean a test did not pass on its last attempt.
const FAILING: [&str; 12] = [
    "FAIL",
    "SIGABRT",
    "SIGSEGV",
    "SIGBUS",
    "SIGILL",
    "SIGFPE",
    "SIGKILL",
    "SIGTERM",
    "SIGTRAP",
    "TIMEOUT",
    "LEAK-FAIL",
    "EXECFAIL",
];

/// The test a nextest result line names: `<WORD> [..] (i/n) <binary id> <test name>`.
fn named(line: &str) -> Option<TestName> {
    let after = &line[line.find(']')? + 1..];
    let rest = after[after.find(')')? + 1..].trim();
    let (binary, name) = rest.split_once(' ')?;
    let (binary, name) = (binary.trim(), name.trim());
    (!binary.is_empty() && !name.is_empty()).then(|| TestName {
        binary: binary.to_string(),
        name: name.to_string(),
    })
}

/// The tests a step log says failed and the tests it says passed on a retry (`FLAKY`). With
/// retries on, nextest prints every attempt as `TRY n <STATUS>` and a test that fails all of them
/// only ever as `TRY n FAIL`; such a line counts as a failure, and a name that later passed on a
/// retry is dropped from the failures below (real output measured in `target/rerun/danger`).
pub fn read_log(text: &str) -> (BTreeSet<TestName>, BTreeSet<TestName>) {
    let (mut failed, mut flaky) = (BTreeSet::new(), BTreeSet::new());
    for line in text.lines() {
        let line = line.trim_start();
        let mut words = line.split(' ');
        let first = words.next().unwrap_or("");
        let word = if first == "TRY" {
            words.nth(1).unwrap_or("")
        } else {
            first
        };
        if let Some(t) = named(line) {
            if FAILING.contains(&word) {
                failed.insert(t);
            } else if word == "FLAKY" {
                flaky.insert(t);
            }
        }
    }
    // A name that failed and then passed on a retry is listed once, as FLAKY.
    failed.retain(|t| !flaky.contains(t));
    (failed, flaky)
}

/// The words of an earlier run's one-line summary that the plan needs.
struct Earlier {
    folder: PathBuf,
    base: String,
    head: String,
    selection: String,
    test: Option<i32>,
    failed: bool,
}

fn read_earlier(folder: &Path) -> Option<Earlier> {
    let text = std::fs::read_to_string(folder.join("summary.txt")).ok()?;
    let line = text.trim_end();
    if line.contains('\n') {
        return None;
    }
    let words: Vec<&str> = line.split(' ').collect();
    if words.first() != Some(&"check-affected") {
        return None;
    }
    let get = |key: &str| {
        words
            .iter()
            .find_map(|w| w.strip_prefix(key))
            .map(str::to_string)
    };
    Some(Earlier {
        folder: folder.to_path_buf(),
        base: get("base=")?,
        head: get("head=")?,
        selection: get("selection=").unwrap_or_else(|| "full".to_string()),
        test: get("test=").and_then(|v| v.parse().ok()),
        failed: words.last() == Some(&"result=FAIL"),
    })
}

/// The newest finished run among `out` and the `check-affected-*` folders beside it.
fn find_earlier(out: &Path) -> Option<Earlier> {
    let mut folders = vec![out.to_path_buf()];
    if let Some(Ok(entries)) = out.parent().map(std::fs::read_dir) {
        for e in entries.flatten() {
            let name = e.file_name().to_string_lossy().into_owned();
            if name.starts_with("check-affected-") && !name.contains(".run-") && e.path() != out {
                folders.push(e.path());
            }
        }
    }
    folders
        .into_iter()
        .filter_map(|f| {
            let at = std::fs::metadata(f.join("summary.txt"))
                .ok()?
                .modified()
                .ok()?;
            Some((at, f))
        })
        .max_by_key(|(at, _)| *at)
        .and_then(|(_, f)| read_earlier(&f))
}

/// The test step log of a check folder: the highest-numbered `NN-test.log`.
fn test_log(folder: &Path) -> Option<PathBuf> {
    let mut logs: Vec<PathBuf> = std::fs::read_dir(folder)
        .ok()?
        .flatten()
        .map(|e| e.path())
        .filter(|p| p.to_string_lossy().ends_with("-test.log"))
        .collect();
    logs.sort();
    logs.pop()
}

/// What a run does about an earlier run of the same project (REQ-1 to REQ-5, REQ-11).
#[derive(Clone, Debug)]
pub enum Mode {
    /// No earlier failed run to use, or `--full`: today's selection, with the retries.
    Plain,
    /// An earlier run failed but cannot be used: the full set, and the summary says why.
    Full(String),
    /// The same head failed before: only the failed tests run (`of` is the earlier folder).
    Same { of: PathBuf, failed: Vec<TestName> },
    /// The head moved since the earlier failed run: the targets of the diff between the two heads
    /// and the earlier failures run.
    Moved {
        of: PathBuf,
        failed: Vec<TestName>,
        selection: Selection,
    },
}

impl Mode {
    /// The folder a narrowed run continues, if it narrows.
    pub fn continues(&self) -> Option<&Path> {
        match self {
            Mode::Same { of, .. } | Mode::Moved { of, .. } => Some(of),
            _ => None,
        }
    }

    /// The summary words this mode adds (REQ-4, REQ-5, REQ-12).
    pub fn words(&self, kept: Option<&Path>) -> Vec<String> {
        match self {
            Mode::Full(why) => vec!["rerun=full".to_string(), format!("reason={why}")],
            Mode::Same { of, .. } | Mode::Moved { of, .. } => {
                let shown = kept.unwrap_or(of);
                vec![format!("rerun_of={}", shown.display())]
            }
            Mode::Plain => Vec::new(),
        }
    }
}

fn plain_name(s: &str) -> bool {
    !s.is_empty()
        && s.chars()
            .all(|c| c.is_ascii_alphanumeric() || c == '_' || c == ':')
}

fn plain_binary(s: &str) -> bool {
    !s.is_empty()
        && s.chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '_' | ':' | '-' | '/'))
}

fn git(dir: &Path, args: &[&str]) -> Option<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd.args(args).current_dir(dir).output().ok()?;
    out.status
        .success()
        .then(|| String::from_utf8_lossy(&out.stdout).trim().to_string())
}

fn spaced(why: &str) -> String {
    why.replace(' ', "_")
}

/// Decide what this run does about the earlier run. `pick` is the selection the run uses, `full`
/// is `--full` or a wide check (REQ-11: no earlier run is read), `out` is the check folder of this
/// run.
pub fn plan(dir: &Path, facts: &Facts, pick: Pick, full: bool, out: &Path) -> Mode {
    if full {
        return Mode::Plain;
    }
    let Some(earlier) = find_earlier(out) else {
        return Mode::Plain;
    };
    if !earlier.failed {
        return Mode::Plain;
    }
    let short = |sha: &str| sha.chars().take(12).collect::<String>();
    if earlier.base != short(&facts.base) {
        return Mode::Full("base_changed".to_string());
    }
    if earlier.selection != pick.word() {
        return Mode::Full("selection_changed".to_string());
    }
    if earlier.test.is_none() {
        return Mode::Full("stopped_before_test".to_string());
    }
    let Some(log) = test_log(&earlier.folder) else {
        return Mode::Full("log_missing".to_string());
    };
    let Ok(text) = std::fs::read_to_string(&log) else {
        return Mode::Full("log_unreadable".to_string());
    };
    let (failed, _) = read_log(&text);
    if failed.is_empty() {
        return Mode::Full("no_failed_test".to_string());
    }
    let readable = failed.iter().all(|t| {
        plain_name(&t.name)
            && plain_binary(&t.binary)
            && (t.binary == "moai-cli" || t.binary.starts_with("moai-cli::"))
    });
    if !readable {
        return Mode::Full("unreadable_name".to_string());
    }
    let failed: Vec<TestName> = failed.into_iter().collect();
    if earlier.head == short(&facts.head) {
        return Mode::Same {
            of: earlier.folder,
            failed,
        };
    }
    let Some(from) = git(
        dir,
        &[
            "rev-parse",
            "--verify",
            "--quiet",
            &format!("{}^{{commit}}", earlier.head),
        ],
    ) else {
        return Mode::Full("earlier_head_unknown".to_string());
    };
    let mut ancestry = Command::new("git");
    crate::identity::scrub_git_env(&mut ancestry);
    let ancestor = ancestry
        .args(["merge-base", "--is-ancestor", &from, &facts.head])
        .current_dir(dir)
        .status()
        .is_ok_and(|s| s.success());
    if !ancestor {
        return Mode::Full("head_not_descendant".to_string());
    }
    let Ok(diff) = crate::check_affected::read_facts(dir, &format!("{from}..{}", facts.head))
    else {
        return Mode::Full("diff_unreadable".to_string());
    };
    if let Some(class) = diff
        .changed
        .iter()
        .find_map(|p| crate::check_affected::wide_class(p))
    {
        return Mode::Full(format!("wide_{}", spaced(class)));
    }
    let chosen = crate::check_affected::choose(
        pick,
        &diff.changed,
        &diff.tests,
        diff.scan_list.as_deref(),
        diff.areas_text.as_deref(),
        crate::check_affected::host_os(),
    );
    if chosen.pick != pick {
        return Mode::Full("wide_path".to_string());
    }
    Mode::Moved {
        of: earlier.folder,
        failed,
        selection: chosen.selection,
    }
}

/// `--flag value` pairs removed from `args`; the values, in order.
fn remove_pairs(args: &mut Vec<String>, flag: &str) -> Vec<String> {
    let mut values = Vec::new();
    let mut i = 0;
    while i < args.len() {
        if args[i] == flag && i + 1 < args.len() {
            values.push(args.remove(i + 1));
            args.remove(i);
        } else {
            i += 1;
        }
    }
    values
}

fn failed_filter(failed: &[TestName]) -> String {
    failed
        .iter()
        .map(|t| format!("(binary_id(={}) & test(={}))", t.binary, t.name))
        .collect::<Vec<_>>()
        .join(" | ")
}

/// Give the nextest `test` step the retries (REQ-6) and, when the mode narrows, the filter that
/// runs only the failed tests (and, for a moved head, the targets of the diff). Any other step,
/// and a `cargo test` step, stays as it is.
pub fn tune(step: &mut Step, mode: &Mode) {
    let nextest =
        step.program == "cargo" && step.args.first().map(String::as_str) == Some("nextest");
    if step.label != "test" || !nextest || step.args.len() < 2 {
        return;
    }
    step.args.splice(
        2..2,
        [
            "--retries".to_string(),
            RETRIES.to_string(),
            "--flaky-result".to_string(),
            "pass".to_string(),
        ],
    );
    let (failed, keep) = match mode {
        Mode::Same { failed, .. } => (failed, false),
        Mode::Moved { failed, .. } => (failed, true),
        _ => return,
    };
    let mut args = std::mem::take(&mut step.args);
    let old_filter = remove_pairs(&mut args, "-E").pop();
    let mut targets: Vec<String> = Vec::new();
    if keep {
        let mut i = 0;
        while i + 1 < args.len() {
            if args[i] == "--test" {
                targets.push(args[i + 1].clone());
            }
            i += 1;
        }
    } else {
        remove_pairs(&mut args, "--test");
    }
    let present = targets.clone();
    for t in failed {
        if let Some(target) = t.binary.strip_prefix("moai-cli::") {
            if !target.starts_with("bin/") && !targets.iter().any(|x| x == target) {
                targets.push(target.to_string());
            }
        }
    }
    for t in targets.iter().filter(|t| !present.contains(t)) {
        args.push("--test".to_string());
        args.push(t.clone());
    }
    let wanted = failed_filter(failed);
    let filter = if keep {
        let base = old_filter.unwrap_or_else(|| {
            let mut parts = vec!["kind(lib)".to_string(), "kind(bin)".to_string()];
            parts.extend(present.iter().map(|t| format!("binary_id(=moai-cli::{t})")));
            parts.join(" | ")
        });
        format!("({base}) | {wanted}")
    } else {
        wanted
    };
    args.push("-E".to_string());
    args.push(filter);
    step.args = args;
}

/// Move the folder a narrowed run continues to `<folder>.run-<n>` (the first free `n`) so the run
/// can write a new one at the same place; the answer names where it went (REQ-12). A folder
/// elsewhere (the run of another head) stays where it is.
pub fn keep_earlier(mode: &Mode, out: &Path) -> Option<PathBuf> {
    let of = mode.continues()?;
    if of != out {
        return Some(of.to_path_buf());
    }
    let name = out.file_name()?.to_string_lossy().into_owned();
    (1..)
        .map(|n| out.with_file_name(format!("{name}.run-{n}")))
        .find(|p| !p.exists())
        .and_then(|to| std::fs::rename(out, &to).ok().map(|()| to))
}

/// The summary line with `extra` words put before its last word, `result=<verdict>`.
pub fn with_words(line: &str, extra: &[String]) -> String {
    if extra.is_empty() {
        return line.to_string();
    }
    match line.rfind(" result=") {
        Some(at) => format!("{} {}{}", &line[..at], extra.join(" "), &line[at..]),
        None => line.to_string(),
    }
}

/// The FLAKY names of the test step log in `out` (empty when the run has no such log).
pub fn flaky_of(out: &Path) -> Vec<TestName> {
    let Some(log) = test_log(out) else {
        return Vec::new();
    };
    let text = std::fs::read_to_string(log).unwrap_or_default();
    read_log(&text).1.into_iter().collect()
}

/// Write `flaky.txt` in `out` and append each name to the ledger of the project at `root`
/// (REQ-7, REQ-9). A failure is returned as text for the caller to put on stderr; it never changes
/// the result of the run.
pub fn record(root: &Path, out: &Path, head: &str, names: &[TestName]) -> Result<(), String> {
    if names.is_empty() {
        return Ok(());
    }
    let list: Vec<String> = names.iter().map(TestName::text).collect();
    let mut problems = Vec::new();
    if let Err(e) = std::fs::write(out.join("flaky.txt"), format!("{}\n", list.join("\n"))) {
        problems.push(format!("cannot write flaky.txt: {e}"));
    }
    let ledger = ledger_path(root);
    let append = || -> std::io::Result<()> {
        use std::io::Write;
        if let Some(parent) = ledger.parent() {
            std::fs::create_dir_all(parent)?;
        }
        let mut file = std::fs::OpenOptions::new()
            .create(true)
            .append(true)
            .open(&ledger)?;
        let at = crate::audit::utc_now();
        let mut text = String::new();
        for name in &list {
            let row = serde_json::json!({
                "name": name,
                "time": at,
                "head": head.chars().take(12).collect::<String>(),
                "folder": out.to_string_lossy(),
            });
            text.push_str(&format!("{row}\n"));
        }
        file.write_all(text.as_bytes())
    };
    if let Err(e) = append() {
        problems.push(format!("cannot append to {}: {e}", ledger.display()));
    }
    if problems.is_empty() {
        Ok(())
    } else {
        Err(problems.join("; "))
    }
}

/// `moai check flaky`: each ledger name with its count, a name seen twice marked `fix` (REQ-10).
pub fn report(dir: &Path) -> Result<serde_json::Value, String> {
    let root = git(dir, &["rev-parse", "--show-toplevel"])
        .map(PathBuf::from)
        .unwrap_or_else(|| dir.to_path_buf());
    let ledger = ledger_path(&root);
    let mut counts: std::collections::BTreeMap<String, usize> = Default::default();
    match std::fs::read_to_string(&ledger) {
        Ok(text) => {
            for (n, line) in text
                .lines()
                .enumerate()
                .filter(|(_, l)| !l.trim().is_empty())
            {
                let name = serde_json::from_str::<serde_json::Value>(line)
                    .ok()
                    .and_then(|v| v["name"].as_str().map(str::to_string))
                    .ok_or_else(|| {
                        format!(
                            "check flaky refused: line {} of {} is not a ledger line",
                            n + 1,
                            ledger.display()
                        )
                    })?;
                *counts.entry(name).or_default() += 1;
            }
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(e) => {
            return Err(format!(
                "check flaky refused: cannot read {}: {e}",
                ledger.display()
            ))
        }
    }
    let mut rows: Vec<(String, usize)> = counts.into_iter().collect();
    rows.sort_by(|a, b| b.1.cmp(&a.1).then_with(|| a.0.cmp(&b.0)));
    let fix = rows.iter().filter(|(_, c)| *c >= FIX_AT).count();
    let flaky: Vec<serde_json::Value> = rows
        .into_iter()
        .map(|(name, count)| {
            let status = if count >= FIX_AT { "fix" } else { "seen" };
            serde_json::json!({"name": name, "count": count, "status": status})
        })
        .collect();
    Ok(serde_json::json!({
        "action": "flaky",
        "ledger": ledger.to_string_lossy(),
        "flaky": flaky,
        "fix": fix,
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    const LOG: &str = "\
        PASS [   0.010s] ( 1/9) moai-cli::a t_a
        TRY 1 FAIL [   0.010s] ( 2/9) moai-cli::a t_b
        FAIL [   0.010s] ( 2/9) moai-cli::a t_b
       FLAKY 2/3 [   0.010s] ( 3/9) moai-cli::b t_c
        TRY 1 FAIL [   0.010s] ( 3/9) moai-cli::b t_c
        SIGABRT [   0.010s] ( 4/9) moai-cli t_lib::deep
        FAIL [   0.010s] ( 2/9) moai-cli::a t_b
";

    #[test]
    fn a_log_gives_the_failed_and_the_flaky_names() {
        let (failed, flaky) = read_log(LOG);
        let text = |s: &BTreeSet<TestName>| s.iter().map(TestName::text).collect::<Vec<_>>();
        assert_eq!(text(&failed), ["moai-cli t_lib::deep", "moai-cli::a t_b"]);
        assert_eq!(text(&flaky), ["moai-cli::b t_c"]);
    }

    /// The shape real nextest 0.9.152 printed with `--retries 2` (measured in the dangerous-path
    /// run): a test that fails every attempt only ever appears as `TRY n FAIL`.
    const REAL: &str = "\
  TRY 1 FAIL [   0.010s] (───) dz::fails always_fails
  TRY 1 FAIL [   0.010s] (───) dz::flaky flaky_once
  TRY 2 FAIL [   0.009s] (───) dz::fails always_fails
  TRY 2 PASS [   0.009s] (1/2) dz::flaky flaky_once
  TRY 3 FAIL [   0.026s] (2/2) dz::fails always_fails
     Summary [   0.048s] 2 tests run: 1 passed (1 flaky), 1 failed, 0 skipped
   FLAKY 2/3 [   0.009s] (1/2) dz::flaky flaky_once
  TRY 3 FAIL [   0.026s] (2/2) dz::fails always_fails
";

    #[test]
    fn a_test_that_fails_every_attempt_is_read_from_its_try_lines() {
        let (failed, flaky) = read_log(REAL);
        let text = |s: &BTreeSet<TestName>| s.iter().map(TestName::text).collect::<Vec<_>>();
        assert_eq!(text(&failed), ["dz::fails always_fails"]);
        assert_eq!(text(&flaky), ["dz::flaky flaky_once"]);
    }

    #[test]
    fn words_go_before_the_result_word() {
        let line = "check-affected base=a head=b test=0 result=PASS";
        assert_eq!(
            with_words(line, &["flaky=1".to_string()]),
            "check-affected base=a head=b test=0 flaky=1 result=PASS"
        );
        assert_eq!(with_words(line, &[]), line);
    }

    fn test_step(args: &[&str]) -> Step {
        Step {
            label: "test".to_string(),
            program: "cargo".to_string(),
            args: args.iter().map(|s| s.to_string()).collect(),
            heavy: true,
            env: Vec::new(),
            report: None,
        }
    }

    #[test]
    fn a_same_head_rerun_keeps_only_the_failed_tests() {
        let mode = Mode::Same {
            of: PathBuf::from("x"),
            failed: vec![TestName {
                binary: "moai-cli::a".to_string(),
                name: "t_b".to_string(),
            }],
        };
        let mut step = test_step(&[
            "nextest",
            "run",
            "--locked",
            "--no-fail-fast",
            "--lib",
            "--test",
            "z",
            "-E",
            "kind(lib)",
        ]);
        tune(&mut step, &mode);
        let line = step.args.join(" ");
        assert!(line.contains("--retries 2 --flaky-result pass"), "{line}");
        assert!(line.contains("--test a"), "{line}");
        assert!(
            !line.contains("--test z") && !line.contains("kind(lib)"),
            "{line}"
        );
        assert!(
            line.ends_with("-E (binary_id(=moai-cli::a) & test(=t_b))"),
            "{line}"
        );
    }

    #[test]
    fn a_plain_run_only_gets_the_retries_and_cargo_test_nothing() {
        let mut step = test_step(&["nextest", "run", "--locked"]);
        tune(&mut step, &Mode::Plain);
        assert_eq!(
            step.args.join(" "),
            "nextest run --retries 2 --flaky-result pass --locked"
        );
        let mut other = test_step(&["test", "--locked"]);
        tune(&mut other, &Mode::Plain);
        assert_eq!(other.args.join(" "), "test --locked");
    }
}
