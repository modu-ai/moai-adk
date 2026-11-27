//! `check fast` (SPEC-20261125-check-fast-and-fix): the light checks of a range and the ledger tests,
//! run first and all of them, so the failures that a hand-kept table causes are named in seconds and
//! in one answer instead of at the end of the long test run. The plan, the end-line check and the
//! ledger list are pure; [`run`] is the one place that starts programs.

use crate::check_affected::{
    judged_test_paths, read_facts, scan_violations, step, table_violations_judging, Facts, Runner,
    Stack, Step, TestFile,
};
use std::ffi::{OsStr, OsString};
use std::path::{Path, PathBuf};
use std::process::Command;

type Result<T> = std::result::Result<T, String>;

/// One name of the ledger list.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Entry {
    /// Every test of a top-level test target (or of a module of an area crate).
    Target(&'static str),
    /// Only the named tests of a test target.
    Tests(&'static str, &'static [&'static str]),
    /// A library unit test, by its path inside the library.
    Lib(&'static str),
}

/// The ledger tests: the ones that fail when a hand-kept table was not updated. They run first in
/// `check fast` and stay in the wide set of `check affected`.
pub const LEDGER: &[Entry] = &[
    Entry::Target("repo_scans_table"),
    Entry::Target("state_folder_literal_scan"),
    Entry::Target("source_manifest_facts"),
    Entry::Target("rule_history_guard"),
    Entry::Target("test_area_layout"),
    Entry::Target("git_env_scrub"),
    Entry::Tests(
        "ac_command_lint",
        &[
            "git_fixture_calls_live_in_the_shared_helper",
            "recorded_ac_commands_name_existing_targets_and_select_tests",
        ],
    ),
    Entry::Lib("check_affected::tests::the_repository_tests_pass_the_table_step_as_added_files"),
    Entry::Lib("check_affected::tests::the_repository_tests_pass_the_scan_step_as_added_files"),
];

/// The ledger list read against the test files of a checkout.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Resolved {
    /// The cargo test targets to build: a top-level target, or the crate of an area module.
    pub targets: Vec<String>,
    /// The nextest filterset terms, one per resolved entry.
    pub terms: Vec<String>,
    /// The library unit tests, by path; the library is built when there are any.
    pub lib: Vec<String>,
    /// How many names of the test directory resolved.
    pub files: usize,
    /// The names that resolve to no test target, area module or test function.
    pub missing: Vec<String>,
}

fn push_unique(list: &mut Vec<String>, item: &str) {
    if !list.iter().any(|x| x == item) {
        list.push(item.to_string());
    }
}

/// Whether `text` declares `fn <name>` as a whole identifier.
fn declares_fn(text: &str, name: &str) -> bool {
    let needle = format!("fn {name}");
    let is_ident = |c: char| c.is_alphanumeric() || c == '_';
    text.match_indices(&needle).any(|(at, _)| {
        let after = text[at + needle.len()..].chars().next();
        let before = text[..at].chars().next_back();
        after.is_some_and(|c| c == '(' || c == '<' || c.is_whitespace())
            && !before.is_some_and(is_ident)
    })
}

/// Resolve the ledger list through `tests` (the test files of `tests/`, area modules included): a
/// name that is a module of an area crate runs through that crate, narrowed to the module; a
/// top-level target is the binary itself. A name that is in no file, and a listed test function a
/// file no longer declares, are `missing`.
pub fn resolve(tests: &[TestFile]) -> Resolved {
    let mut out = Resolved::default();
    for entry in LEDGER {
        match *entry {
            Entry::Lib(path) => {
                out.lib.push(path.to_string());
                out.terms.push(format!("(kind(lib) & test(={path}))"));
            }
            Entry::Target(name) => match tests.iter().find(|t| t.name == name) {
                None => out.missing.push(name.to_string()),
                Some(file) => {
                    out.files += 1;
                    match file.area.as_deref() {
                        Some(krate) => {
                            push_unique(&mut out.targets, krate);
                            out.terms
                                .push(format!("(binary(={krate}) & test(/^{name}::/))"));
                        }
                        None => {
                            push_unique(&mut out.targets, name);
                            out.terms.push(format!("binary(={name})"));
                        }
                    }
                }
            },
            Entry::Tests(name, fns) => match tests.iter().find(|t| t.name == name) {
                None => out.missing.push(name.to_string()),
                Some(file) => {
                    out.files += 1;
                    let (krate, prefix) = match file.area.as_deref() {
                        Some(krate) => (krate, format!("{name}::")),
                        None => (name, String::new()),
                    };
                    let mut present = Vec::new();
                    for f in fns {
                        if declares_fn(&file.text, f) {
                            present.push(format!("test(={prefix}{f})"));
                        } else {
                            out.missing.push(format!("{name}::{f}"));
                        }
                    }
                    if !present.is_empty() {
                        push_unique(&mut out.targets, krate);
                        out.terms
                            .push(format!("(binary(={krate}) & ({}))", present.join(" | ")));
                    }
                }
            },
        }
    }
    out
}

/// Whether the library test `path` (`<module>::tests::<name>`) is declared in the sources of
/// `root`: the module's file, its `mod.rs` or its `tests.rs`.
fn lib_test_exists(root: &Path, path: &str) -> bool {
    let module = path.split("::").next().unwrap_or(path);
    let name = path.rsplit("::").next().unwrap_or(path);
    [
        format!("src/{module}.rs"),
        format!("src/{module}/mod.rs"),
        format!("src/{module}/tests.rs"),
    ]
    .iter()
    .any(|rel| std::fs::read_to_string(root.join(rel)).is_ok_and(|text| declares_fn(&text, name)))
}

/// The library tests of the ledger whose function cannot be found in the sources of `root`.
pub fn lib_missing(root: &Path) -> Vec<String> {
    LEDGER
        .iter()
        .filter_map(|e| match *e {
            Entry::Lib(path) => Some(path),
            _ => None,
        })
        .filter(|path| !lib_test_exists(root, path))
        .map(String::from)
        .collect()
}

/// An in-process step: `lines` are its verdict, decided when the plan is made.
fn report_step(label: &str, what: &str, lines: Vec<String>) -> Step {
    let mut s = step(label, "(in-process)", &[label, what], false);
    s.report = Some(lines);
    s
}

/// The ledger step: one heavy test run of every ledger name, or a failing in-process report that
/// names what is missing; `None` when the checkout has no ledger test file at all (REQ-6, REQ-15).
/// With nextest the run is `--no-fail-fast` and narrowed by one filterset; `cargo test` has no
/// per-binary filter, so it runs the whole target of each name and leaves the library tests to
/// the nextest run.
pub fn ledger_step(facts: &Facts, runner: Runner) -> Option<Step> {
    let r = resolve(&facts.tests);
    if r.files == 0 {
        return None;
    }
    let mut missing = r.missing.clone();
    missing.extend(lib_missing(&facts.root));
    if !missing.is_empty() {
        let lines = missing
            .iter()
            .map(|m| {
                format!(
                    "{m}: no test target, area module or test function of this name exists; restore it or fix the ledger list in src/check_fast.rs"
                )
            })
            .collect();
        return Some(report_step("ledger", "missing names", lines));
    }
    let mut args: Vec<String> = match runner {
        Runner::Nextest => ["nextest", "run", "--locked", "--offline", "--no-fail-fast"]
            .map(String::from)
            .to_vec(),
        Runner::CargoTest => ["test", "--locked", "--offline"].map(String::from).to_vec(),
    };
    if runner == Runner::Nextest && !r.lib.is_empty() {
        args.push("--lib".to_string());
    }
    for t in &r.targets {
        args.push("--test".to_string());
        args.push(t.clone());
    }
    if runner == Runner::Nextest {
        args.push("-E".to_string());
        args.push(r.terms.join(" | "));
    }
    Some(Step {
        label: "ledger".to_string(),
        program: "cargo".to_string(),
        args,
        heavy: true,
        env: Vec::new(),
        report: None,
    })
}

/// The changed files whose last byte is not a newline, as report lines (REQ-3, REQ-4). A file is
/// checked when it is a regular file that is not a link, is not empty, holds no NUL byte (a binary
/// file) and does not lie under `templates/bundle-packs/`. A file that cannot be read is a line of
/// its own that names it; a path that no longer exists is skipped.
pub fn endline_violations(root: &Path, present: &[String]) -> Vec<String> {
    let mut out = Vec::new();
    for p in present {
        if p.starts_with("templates/bundle-packs/") {
            continue;
        }
        let full = root.join(p);
        let meta = match std::fs::symlink_metadata(&full) {
            Ok(m) => m,
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => continue,
            Err(e) => {
                out.push(unreadable_line(p, &e));
                continue;
            }
        };
        if !meta.is_file() {
            continue;
        }
        match std::fs::read(&full) {
            Ok(bytes) => {
                if !bytes.is_empty() && !bytes.contains(&0) && bytes.last() != Some(&b'\n') {
                    out.push(format!("{p}: no newline at end of file; add one"));
                }
            }
            Err(e) => out.push(unreadable_line(p, &e)),
        }
    }
    out
}

fn unreadable_line(path: &str, error: &std::io::Error) -> String {
    format!("{path}: cannot be read ({error}); the end-line check could not run on it")
}

fn endline_step(facts: &Facts) -> Step {
    report_step(
        "endline",
        &format!("{} changed file(s)", facts.present.len()),
        endline_violations(&facts.root, &facts.present),
    )
}

/// The end-line step of a range, only when it has a violation to name: a clean range plans no
/// such step, so the summary words of every earlier run stay as they were.
pub fn endline_step_if_any(facts: &Facts) -> Option<Step> {
    let s = endline_step(facts);
    s.report
        .as_ref()
        .is_some_and(|l| !l.is_empty())
        .then_some(s)
}

/// Whether the scan registry's text is a list of scans whose checks can be read.
fn registry_readable(text: &str) -> bool {
    let Ok(value) = serde_json::from_str::<serde_json::Value>(text) else {
        return false;
    };
    value
        .get("scans")
        .and_then(|s| s.as_array())
        .is_some_and(|scans| {
            scans
                .iter()
                .all(|s| crate::repo_scans::checks_of(s).is_ok())
        })
}

/// The cheap step `test_tables`: the top-level test files the range adds, checked against the area
/// table, the state-folder allow list and the scan registry with the same judges `check affected`
/// uses, and every changed test file (C-431, [`judged_test_paths`]) against the allow list and the
/// scan registry. `None` when the range adds no top-level file and no changed test file breaks a
/// table, or the repository has no table.
fn table_step(facts: &Facts) -> Option<Step> {
    let added = facts
        .added
        .iter()
        .filter(|p| {
            p.strip_prefix("tests/")
                .and_then(|r| r.strip_suffix(".rs"))
                .is_some_and(|s| !s.contains('/'))
        })
        .count();
    let judged = judged_test_paths(&facts.present, &facts.moved);
    if (added == 0 && judged.is_empty())
        || (facts.areas_text.is_none()
            && facts.allow_text.is_none()
            && facts.scan_list.is_none()
            && facts.unreadable.is_empty())
    {
        return None;
    }
    let mut lines = table_violations_judging(
        &facts.added,
        &judged,
        &facts.tests,
        facts.areas_text.as_deref(),
        facts.allow_text.as_deref(),
        &facts.moved,
    );
    if let Some(list) = facts.scan_list.as_deref() {
        if !registry_readable(list) {
            lines.push(format!(
                "{}: the table exists but cannot be read (it is not JSON with a `scans` list whose entries each have readable checks); fix or restore it, it is not treated as absent",
                crate::gate::SCAN_REGISTRY
            ));
        }
    }
    lines.extend(scan_violations(
        &judged,
        &facts.tests,
        facts.scan_list.as_deref(),
        facts.tracked.as_deref(),
    ));
    for (table, error) in &facts.unreadable {
        lines.push(format!(
            "{table}: the table exists but cannot be read ({error}); fix or restore it, it is not treated as absent"
        ));
    }
    // As in `check affected`: only a miss plans the step for a range that adds no top-level file.
    if added == 0 && lines.is_empty() {
        return None;
    }
    let what = if added > 0 {
        format!("{added} added test file(s)")
    } else {
        format!("{} changed test file(s)", judged.len())
    };
    Some(report_step("test_tables", &what, lines))
}

/// The steps of `check fast`, in the order they run: the cheap ones, then the ledger. A repository
/// without Rust code gets the two checks that need no toolchain.
pub fn plan(facts: &Facts, runner: Runner) -> Vec<Step> {
    let range = format!("{}...{}", facts.base, facts.head);
    let rust = facts.stack == Stack::Rust;
    let mut out = vec![step(
        "diff_check",
        "git",
        &["diff", "--check", &range],
        false,
    )];
    if rust {
        out.push(step(
            "fmt",
            "cargo",
            &["fmt", "--all", "--", "--check"],
            false,
        ));
        let ops: Vec<&String> = facts
            .present
            .iter()
            .filter(|p| p.starts_with("src/ops/") && p.ends_with(".rs"))
            .collect();
        if !ops.is_empty() {
            let mut s = step(
                "ops_rustfmt",
                "rustfmt",
                &["--edition", "2021", "--check"],
                false,
            );
            s.args.extend(ops.into_iter().cloned());
            out.push(s);
        }
    }
    out.extend(table_step(facts));
    out.push(endline_step(facts));
    if rust {
        out.extend(ledger_step(facts, runner));
    }
    out
}

/// The words of a nextest line that mean a test did not pass on its last attempt.
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

/// `<binary> <test>` of a nextest result line: `<WORD> [<time>] (<i>/<n>) <binary> <test>`, the
/// progress counter being optional.
fn nextest_name(line: &str) -> Option<String> {
    let after = &line[line.find(']')? + 1..];
    let mut rest = after.trim_start();
    if rest.starts_with('(') {
        rest = rest[rest.find(')')? + 1..].trim_start();
    }
    let (binary, name) = rest.split_once(' ')?;
    let (binary, name) = (binary.trim(), name.trim());
    (!binary.is_empty() && !name.is_empty()).then(|| format!("{binary} {name}"))
}

/// The failing test names a step log holds, once each in log order: `<binary> <test>` from a
/// nextest line (which stderr carries into the log), `<test>` from a libtest `... FAILED` line. A
/// test that failed and then passed on a retry (`FLAKY`) is not listed.
pub fn failed_tests(log: &str) -> Vec<String> {
    let mut failed: Vec<String> = Vec::new();
    let mut flaky: Vec<String> = Vec::new();
    for line in log.lines() {
        let line = line.trim_start();
        if let Some(name) = line
            .strip_prefix("test ")
            .and_then(|r| r.strip_suffix(" ... FAILED"))
        {
            push_unique(&mut failed, name.trim());
            continue;
        }
        let mut words = line.split(' ');
        let first = words.next().unwrap_or("");
        let word = if first == "TRY" {
            words.nth(1).unwrap_or("")
        } else {
            first
        };
        if let Some(name) = nextest_name(line) {
            if FAILING.contains(&word) {
                push_unique(&mut failed, &name);
            } else if word == "FLAKY" {
                push_unique(&mut flaky, &name);
            }
        }
    }
    failed.retain(|n| !flaky.contains(n));
    failed
}

/// What `run` needs besides the range.
#[derive(Clone, Debug)]
pub struct Options {
    pub slot_dir: Option<PathBuf>,
    pub max: u32,
    pub min_free_gb: u32,
    /// Where each step's log goes (default: `<root>/target/check-fast-<head12>`).
    pub out: Option<PathBuf>,
    pub integration_ref: Option<String>,
    /// The cargo program to run (default: `cargo` from the search path).
    pub cargo: Option<PathBuf>,
}

impl Default for Options {
    fn default() -> Self {
        Options {
            slot_dir: None,
            max: 4,
            min_free_gb: 50,
            out: None,
            integration_ref: None,
            cargo: None,
        }
    }
}

fn short(sha: &str) -> &str {
    sha.get(..12).unwrap_or(sha)
}

/// One git call in `dir`, with the inherited git environment scrubbed.
fn git(dir: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let o = cmd
        .args(args)
        .current_dir(dir)
        .output()
        .map_err(|e| format!("cannot run git: {e}"))?;
    if !o.status.success() {
        return Err(format!(
            "git {} failed: {}",
            args.first().copied().unwrap_or(""),
            String::from_utf8_lossy(&o.stderr).trim()
        ));
    }
    Ok(String::from_utf8_lossy(&o.stdout).into_owned())
}

/// The folder cargo builds into: the caller's `CARGO_TARGET_DIR` (a relative one counts from the
/// repository), else `<root>/target`. The step logs go below the same folder.
fn target_dir(root: &Path) -> PathBuf {
    match std::env::var_os("CARGO_TARGET_DIR") {
        Some(t) if !t.is_empty() => root.join(t),
        _ => root.join("target"),
    }
}

/// The exit status of a finished process: its code, or 128 + the signal.
fn exit_of(status: std::process::ExitStatus) -> i32 {
    if let Some(c) = status.code() {
        return c;
    }
    #[cfg(unix)]
    {
        use std::os::unix::process::ExitStatusExt;
        if let Some(sig) = status.signal() {
            return 128 + sig;
        }
    }
    1
}

/// A check build is unnumbered: neither build-number variable reaches a step.
fn unnumbered(cmd: &mut Command) {
    for var in crate::check_affected::BUILD_NUMBER_VARS {
        cmd.env_remove(var);
    }
}

/// `cargo nextest --version` in `dir` through `cargo`: nextest when it exits 0, else `cargo test`.
fn runner_with(cargo: &OsStr, dir: &Path) -> Runner {
    let mut cmd = Command::new(cargo);
    unnumbered(&mut cmd);
    let ok = cmd
        .args(["nextest", "--version"])
        .current_dir(dir)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .status()
        .is_ok_and(|s| s.success());
    if ok {
        Runner::Nextest
    } else {
        Runner::CargoTest
    }
}

/// Run one step's program in `dir` with `extra` added to its environment; its output (stderr
/// included) goes to `log`. A program that cannot start is 127.
fn run_step(dir: &Path, program: &OsStr, s: &Step, log: &Path, extra: &[(&str, OsString)]) -> i32 {
    let Ok(file) = std::fs::File::create(log) else {
        return 127;
    };
    let Ok(err) = file.try_clone() else {
        return 127;
    };
    let mut cmd = Command::new(program);
    if s.program == "git" {
        crate::identity::scrub_git_env(&mut cmd);
    }
    cmd.args(&s.args)
        .current_dir(dir)
        .envs(s.env.iter().map(|(k, v)| (k, v)))
        .envs(extra.iter().map(|(k, v)| (k, v)))
        .stdin(std::process::Stdio::null())
        .stdout(file)
        .stderr(err);
    unnumbered(&mut cmd);
    match cmd.status() {
        Ok(status) => exit_of(status),
        Err(_) => 127,
    }
}

/// An in-process step: its verdict lines go to the log and any line makes the exit code 1.
fn run_report(lines: &[String], log: &Path) -> i32 {
    let text = if lines.is_empty() {
        "no violation\n".to_string()
    } else {
        format!("{}\n", lines.join("\n"))
    };
    let _ = std::fs::write(log, text);
    i32::from(!lines.is_empty())
}

/// Takes the one heavy slot of a run, with its slot count; the waiting lines and any refusal go to
/// `<out>/heavy-slot.log` because stdout is the JSON answer. `Err` is the slot's exit code: 125
/// refused, 75 machine too busy, 2 no usable slot folder or `--max 0`.
fn hold_slot(
    opts: &Options,
    out: &Path,
) -> std::result::Result<(crate::heavy_slot::Held, usize), i32> {
    use std::io::Write;
    let mut log: Box<dyn Write> = match std::fs::File::create(out.join("heavy-slot.log")) {
        Ok(f) => Box::new(f),
        Err(_) => Box::new(std::io::sink()),
    };
    let folder = match opts
        .slot_dir
        .clone()
        .map(Ok)
        .unwrap_or_else(crate::heavy_slot::default_dir)
    {
        Ok(d) => d,
        Err(e) => {
            let _ = writeln!(log, "heavy-slot: {e}");
            return Err(i32::from(crate::heavy_slot::USAGE));
        }
    };
    let (max, limits, notes) = match crate::heavy_slot::resolve(
        &folder,
        Some(opts.max as usize),
        Some(u64::from(opts.min_free_gb)),
        None,
        None,
    ) {
        Ok(r) => r,
        Err(e) => {
            let _ = writeln!(
                log,
                "heavy-slot: refused {e}; nothing was written and no cargo step ran"
            );
            return Err(i32::from(crate::heavy_slot::REFUSED));
        }
    };
    for n in notes {
        let _ = writeln!(log, "{n}");
    }
    if max < 1 {
        let _ = writeln!(log, "heavy-slot: --max 0 leaves no slot; no cargo step ran");
        return Err(i32::from(crate::heavy_slot::USAGE));
    }
    let name = "check-fast";
    let command: Vec<String> = std::env::args_os()
        .map(|a| a.to_string_lossy().into_owned())
        .collect();
    let held = crate::heavy_slot::hold(&folder, max, name, &command, limits, &mut log)
        .map_err(i32::from)?;
    let _ = writeln!(log, "{}", held.taken_line(name));
    Ok((held, max))
}

/// `check fast <base>..<head>` (REQ-1, REQ-2): refuse unless the checkout is the range's head and
/// no tracked file is modified, then run every step of [`plan`] and never stop at a failing one.
/// The heavy ledger step runs under one heavy slot, taken just before it; a slot that is not
/// obtained gives its exit code to that step and starts nothing. The answer lists every step and,
/// for each failed one, its exit code, its log and the failing test names.
pub fn run(dir: &Path, range: &str, opts: &Options) -> Result<serde_json::Value> {
    let facts = read_facts(dir, range)?;
    let root = facts.root.as_path();
    let at = git(root, &["rev-parse", "HEAD"])?.trim().to_string();
    if at != facts.head {
        return Err(format!(
            "check fast refused: the checkout is at {} but the range ends at {}; check out the head first",
            short(&at),
            short(&facts.head)
        ));
    }
    let dirty = git(root, &["status", "--porcelain", "--untracked-files=no"])?;
    if !dirty.trim().is_empty() {
        return Err(
            "check fast refused: a tracked file is modified; commit or restore it first"
                .to_string(),
        );
    }
    let out = opts.out.clone().unwrap_or_else(|| {
        let target = std::fs::canonicalize(root).unwrap_or_else(|_| root.to_path_buf());
        target_dir(&target).join(format!("check-fast-{}", short(&facts.head)))
    });
    std::fs::create_dir_all(&out)
        .map_err(|e| format!("check fast refused: cannot create {}: {e}", out.display()))?;
    let cargo: OsString = opts
        .cargo
        .as_ref()
        .map_or_else(|| OsString::from("cargo"), |p| p.as_os_str().to_os_string());
    // Planned with nextest until the slot is held: the runner is asked only then, so a refused
    // slot starts no cargo test at all. The labels do not depend on the runner.
    let mut results: Vec<(String, i32)> = Vec::new();
    let mut records = Vec::new();
    let mut failures = Vec::new();
    let mut held: Option<(crate::heavy_slot::Held, usize)> = None;
    let mut refused: Option<i32> = None;
    let mut tried = false;
    for (i, mut s) in plan(&facts, Runner::Nextest).into_iter().enumerate() {
        if s.heavy && !tried {
            tried = true;
            match hold_slot(opts, &out) {
                Ok(h) => {
                    let runner = runner_with(&cargo, root);
                    if let Some(again) = ledger_step(&facts, runner) {
                        s = again;
                    }
                    held = Some(h);
                }
                Err(code) => refused = Some(code),
            }
        }
        let s = &s;
        let log = out.join(format!(
            "{:02}-{}.log",
            i + 1,
            s.label.replace([':', '/'], "_")
        ));
        let program = if s.program == "cargo" {
            cargo.clone()
        } else {
            OsString::from(&s.program)
        };
        let started = std::time::Instant::now();
        let code = match (&s.report, s.heavy, &held, refused) {
            (Some(lines), _, _, _) => run_report(lines, &log),
            (None, true, Some((h, slots)), _) => {
                let mut extra = vec![(crate::heavy_slot::SLOT_FD_ENV, h.env_value())];
                if std::env::var_os("CARGO_BUILD_JOBS").is_none() {
                    let cpus = std::thread::available_parallelism().map_or(1, |n| n.get());
                    let jobs = crate::check_affected::default_jobs(cpus, *slots);
                    extra.push(("CARGO_BUILD_JOBS", jobs.to_string().into()));
                }
                run_step(root, &program, s, &log, &extra)
            }
            (None, true, None, Some(code)) => {
                let _ = std::fs::write(
                    &log,
                    format!("not run: the heavy slot was not obtained (exit {code}); see heavy-slot.log\n"),
                );
                code
            }
            _ => run_step(root, &program, s, &log, &[]),
        };
        let elapsed_ms = u64::try_from(started.elapsed().as_millis()).unwrap_or(u64::MAX);
        let mut command = vec![s.program.clone()];
        command.extend(s.args.iter().cloned());
        results.push((s.label.clone(), code));
        records.push(serde_json::json!({
            "label": s.label,
            "exit": code,
            "log": log.to_string_lossy(),
            "elapsed_ms": elapsed_ms,
            "command": command,
        }));
        if code != 0 {
            let text = std::fs::read_to_string(&log).unwrap_or_default();
            let mut failure = serde_json::json!({
                "label": s.label,
                "exit": code,
                "log": log.to_string_lossy(),
                "tests": failed_tests(&text),
            });
            // The verdict lines of an in-process step are short and are the reason it failed.
            if let Some(lines) = &s.report {
                failure["lines"] = serde_json::json!(lines);
            }
            failures.push(failure);
        }
    }
    drop(held);
    let pass = failures.is_empty();
    let mut words = vec![
        "check-fast".to_string(),
        format!("base={}", short(&facts.base)),
        format!("head={}", short(&facts.head)),
    ];
    for (label, code) in &results {
        words.push(format!(
            "{}={code}",
            label.split(':').next().unwrap_or(label)
        ));
    }
    words.push(format!("result={}", if pass { "PASS" } else { "FAIL" }));
    Ok(serde_json::json!({
        "action": "fast",
        "result": if pass { "PASS" } else { "FAIL" },
        "summary": words.join(" "),
        "steps": records,
        "failures": failures,
        "out": out.to_string_lossy(),
        "writes": true,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::check_affected::{choose, host_os, read_facts, Pick, Stack};

    const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");

    /// The facts of this checkout with an empty range: its real test files, scan list and tables.
    fn live() -> Facts {
        read_facts(Path::new(MANIFEST), "HEAD..HEAD").expect("read the facts of this checkout")
    }

    fn file(name: &str, area: Option<&str>, text: &str) -> TestFile {
        TestFile {
            name: name.to_string(),
            text: text.to_string(),
            area: area.map(String::from),
            ..Default::default()
        }
    }

    /// The test files of a checkout that holds every top-level ledger name.
    fn complete() -> Vec<TestFile> {
        let mut out = Vec::new();
        for e in LEDGER {
            match e {
                Entry::Target(n) => out.push(file(n, None, "")),
                Entry::Tests(n, fns) => {
                    let text: String = fns.iter().map(|f| format!("fn {f}() {{}}\n")).collect();
                    out.push(file(n, None, &text));
                }
                Entry::Lib(_) => {}
            }
        }
        out
    }

    fn facts_of(root: &Path, tests: Vec<TestFile>) -> Facts {
        Facts {
            root: root.to_path_buf(),
            base: "a".repeat(40),
            head: "b".repeat(40),
            changed: Vec::new(),
            present: Vec::new(),
            tests,
            scan_list: None,
            registry: None,
            added: Vec::new(),
            areas_text: None,
            allow_text: None,
            unreadable: Vec::new(),
            tracked: None,
            moved: Vec::new(),
            stack: Stack::Rust,
        }
    }

    /// C-431 (a): `check fast` names, before the ledger, what an existing test file that gains a
    /// state folder name breaks, and a module added alone to a moved area, as `check affected` does
    /// (the shapes of C-607, C-460 and C-608 on 2026-11-27).
    #[test]
    fn the_fast_table_step_reads_a_changed_file_and_a_module_added_alone() {
        let dir = tempfile::tempdir().unwrap();
        let old = crate::check_affected::earlier_state_name();
        let new = crate::check_affected::new_state_name();
        let mut modified = facts_of(
            dir.path(),
            vec![file("zz_old", None, &format!("let p = \"{old}/x\";"))],
        );
        modified.present = vec!["tests/zz_old.rs".to_string()];
        modified.allow_text = Some("# none\n".to_string());
        let lines = table_step(&modified)
            .expect("a changed test file that breaks a table plans a step")
            .report
            .expect("an in-process report");
        assert_eq!(lines.len(), 1, "{lines:?}");
        assert!(lines[0].starts_with("tests/zz_old.rs"), "{lines:?}");
        let mut nested = facts_of(
            dir.path(),
            vec![file(
                "zz_new",
                Some("lane"),
                &format!("let p = \"{new}/x\";"),
            )],
        );
        nested.present = vec!["tests/lane/zz_new.rs".to_string()];
        nested.added = nested.present.clone();
        nested.moved = vec!["lane".to_string()];
        nested.allow_text = Some("# none\n".to_string());
        let lines = table_step(&nested)
            .expect("a module added alone plans the step")
            .report
            .expect("an in-process report");
        assert_eq!(lines.len(), 1, "{lines:?}");
        assert!(lines[0].starts_with("tests/lane/zz_new.rs"), "{lines:?}");
        // A changed file that breaks nothing plans no step.
        let mut clean = facts_of(dir.path(), vec![file("zz_clean", None, "fn t() {}")]);
        clean.present = vec!["tests/zz_clean.rs".to_string()];
        clean.allow_text = Some("# none\n".to_string());
        assert!(table_step(&clean).is_none());
    }

    /// AC-6: every name of the list is a real target, area module or library test here.
    #[test]
    fn the_ledger_names_resolve_in_this_repository() {
        let facts = live();
        let r = resolve(&facts.tests);
        assert!(r.missing.is_empty(), "missing: {:?}", r.missing);
        let named = LEDGER
            .iter()
            .filter(|e| !matches!(e, Entry::Lib(_)))
            .count();
        assert_eq!(r.files, named, "every test-directory name resolves once");
        let lib = lib_missing(Path::new(MANIFEST));
        assert!(lib.is_empty(), "library tests not found: {lib:?}");
        let libs = LEDGER.iter().filter(|e| matches!(e, Entry::Lib(_))).count();
        assert_eq!(r.lib.len(), libs);
    }

    /// AC-7: the ledger step of the plan holds every ledger test whatever the pick: the plan of a
    /// wide change (the dependency file gives the full set) and the plan of a narrow one carry the
    /// same ledger step, so the ledger runs in the wide check unchanged.
    #[test]
    fn the_plan_of_a_wide_change_holds_every_ledger_target() {
        let facts = live();
        let wide = vec!["Cargo.toml".to_string()];
        let chosen = choose(
            Pick::Card,
            &wide,
            &facts.tests,
            facts.scan_list.as_deref(),
            facts.areas_text.as_deref(),
            host_os(),
        );
        assert_eq!(
            chosen.pick,
            Pick::Full,
            "the dependency file is a wide path"
        );
        let ledger = |changed: &[&str]| {
            let mut f = live();
            f.changed = changed.iter().map(|s| s.to_string()).collect();
            plan(&f, Runner::Nextest)
                .into_iter()
                .find(|s| s.label == "ledger")
                .expect("the plan has a ledger step")
        };
        let full = ledger(&["Cargo.toml"]);
        let narrow = ledger(&["docs/lane-rules.md"]);
        assert_eq!(full.args, narrow.args, "one ledger step for every pick");
        let lacking: Vec<&str> = LEDGER
            .iter()
            .filter_map(|e| match e {
                Entry::Target(n) | Entry::Tests(n, _) => Some(*n),
                Entry::Lib(_) => None,
            })
            .filter(|n| !full.args.iter().any(|a| a == n))
            .collect();
        assert!(
            lacking.is_empty(),
            "the ledger step lacks the ledger targets {lacking:?}"
        );
        assert!(
            full.args.iter().any(|a| a == "--lib"),
            "the library ledger tests run too"
        );
    }

    /// REQ-14, REQ-15: a name that resolves to nothing is missing; a limited entry whose function
    /// is gone is missing too, and one that is present resolves.
    #[test]
    fn a_name_that_resolves_to_nothing_is_missing() {
        let mut tests = complete();
        tests.retain(|t| t.name != "git_env_scrub");
        for t in tests.iter_mut().filter(|t| t.name == "ac_command_lint") {
            t.text = "fn git_fixture_calls_live_in_the_shared_helper() {}\n".to_string();
        }
        let r = resolve(&tests);
        assert_eq!(
            r.missing,
            [
                "git_env_scrub".to_string(),
                "ac_command_lint::recorded_ac_commands_name_existing_targets_and_select_tests"
                    .to_string()
            ]
        );
        assert!(!r.targets.iter().any(|t| t == "git_env_scrub"));
        let whole = resolve(&complete());
        assert!(whole.missing.is_empty(), "{:?}", whole.missing);
    }

    /// REQ-14: a name that is a module of an area crate runs through that crate, narrowed to the
    /// module; a top-level target is the binary itself; a library test is `kind(lib)`.
    #[test]
    fn a_resolved_name_runs_through_its_target() {
        let mut tests = complete();
        for t in tests.iter_mut().filter(|t| t.name == "repo_scans_table") {
            t.area = Some("spec".to_string());
        }
        let r = resolve(&tests);
        assert!(r.targets.iter().any(|t| t == "spec"));
        assert!(!r.targets.iter().any(|t| t == "repo_scans_table"));
        assert!(r
            .terms
            .contains(&"(binary(=spec) & test(/^repo_scans_table::/))".to_string()));
        assert!(r.terms.contains(&"binary(=git_env_scrub)".to_string()));
        assert!(r.terms.contains(
            &"(binary(=ac_command_lint) & (test(=git_fixture_calls_live_in_the_shared_helper) | test(=recorded_ac_commands_name_existing_targets_and_select_tests)))"
                .to_string()
        ));
        assert!(r.terms.contains(
            &"(kind(lib) & test(=check_affected::tests::the_repository_tests_pass_the_table_step_as_added_files))"
                .to_string()
        ));
    }

    /// REQ-6: a checkout with no ledger test file gets no ledger step; with one, the step is a
    /// heavy nextest run (or one `cargo test` of whole targets) and never starts clippy.
    #[test]
    fn the_ledger_step_is_none_without_a_ledger_file_and_one_run_with() {
        let dir = tempfile::tempdir().unwrap();
        let none = facts_of(dir.path(), vec![file("unrelated", None, "")]);
        assert_eq!(ledger_step(&none, Runner::Nextest), None);
        let mut ours = facts_of(dir.path(), complete());
        // The library tests live in this repository's sources, so give the checkout the same ones.
        std::fs::create_dir_all(dir.path().join("src")).unwrap();
        let lib: String = LEDGER
            .iter()
            .filter_map(|e| match e {
                Entry::Lib(p) => p.rsplit("::").next(),
                _ => None,
            })
            .map(|f| format!("fn {f}() {{}}\n"))
            .collect();
        std::fs::write(dir.path().join("src/check_affected.rs"), lib).unwrap();
        ours.stack = Stack::Rust;
        let s = ledger_step(&ours, Runner::Nextest).expect("a ledger step");
        assert_eq!(s.label, "ledger");
        assert!(s.heavy && s.report.is_none());
        assert_eq!(s.program, "cargo");
        assert_eq!(
            s.args[..5],
            ["nextest", "run", "--locked", "--offline", "--no-fail-fast"]
        );
        assert!(s.args.iter().any(|a| a == "--lib"));
        assert!(s.args.iter().any(|a| a == "-E"));
        let plain = ledger_step(&ours, Runner::CargoTest).expect("a ledger step");
        assert_eq!(plain.args[..3], ["test", "--locked", "--offline"]);
        assert!(plain.args.iter().any(|a| a == "git_env_scrub"));
        assert!(!plain.args.iter().any(|a| a == "-E"));
        // A name that is missing makes the step an in-process failure that names it.
        let mut broken = ours;
        broken.tests.retain(|t| t.name != "rule_history_guard");
        let s = ledger_step(&broken, Runner::Nextest).expect("a ledger step");
        let lines = s.report.expect("an in-process report");
        assert!(lines.iter().any(|l| l.contains("rule_history_guard")));
        assert!(!s.heavy);
    }

    /// The plan puts the ledger last and keeps the cheap steps in front of it.
    #[test]
    fn the_plan_runs_the_cheap_steps_before_the_ledger() {
        let dir = tempfile::tempdir().unwrap();
        std::fs::create_dir_all(dir.path().join("src")).unwrap();
        let mut facts = facts_of(dir.path(), complete());
        facts.present = vec!["src/ops/a.rs".to_string()];
        let lib: String = LEDGER
            .iter()
            .filter_map(|e| match e {
                Entry::Lib(p) => p.rsplit("::").next(),
                _ => None,
            })
            .map(|f| format!("fn {f}() {{}}\n"))
            .collect();
        std::fs::write(dir.path().join("src/check_affected.rs"), lib).unwrap();
        let labels: Vec<String> = plan(&facts, Runner::Nextest)
            .into_iter()
            .map(|s| s.label)
            .collect();
        assert_eq!(
            labels,
            ["diff_check", "fmt", "ops_rustfmt", "endline", "ledger"]
        );
        facts.present.clear();
        let labels: Vec<String> = plan(&facts, Runner::Nextest)
            .into_iter()
            .map(|s| s.label)
            .collect();
        assert_eq!(labels, ["diff_check", "fmt", "endline", "ledger"]);
    }

    /// The failing names come from the failure lines only, once each, with or without the progress
    /// counter and the retry prefix of nextest, and from the libtest line.
    #[test]
    fn the_failing_test_names_come_from_the_failure_lines() {
        let log = "\
        PASS [   0.010s] ( 1/9) moai-cli::a t_ok\n\
        FAIL [   0.010s] ( 2/9) moai-cli::a t_b\n\
        FAIL [   0.1s] repo_scans_table some_test\n\
  TRY 1 FAIL [   0.010s] ( 3/9) moai-cli::b t_c\n\
test d::e ... FAILED\n\
test d::f ... ok\n\
        FAIL [   0.010s] ( 2/9) moai-cli::a t_b\n";
        assert_eq!(
            failed_tests(log),
            [
                "moai-cli::a t_b",
                "repo_scans_table some_test",
                "moai-cli::b t_c",
                "d::e"
            ]
        );
        assert!(failed_tests("PASS [ 0.1s] a b\nno failure here\n").is_empty());
    }
}
