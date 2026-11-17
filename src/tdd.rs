//! The TDD phase record and the GREEN test lock (card C-86 slice 3,
//! SPEC-20261117-green-test-lock; user decision 2026-11-17 "권고 말고 강제", research proposal B).
//!
//! - [`TestPaths`]: one table of test paths per language the project detector knows
//!   ([`TEST_PATTERNS`]); a project uses the rows of its own languages and `common`, plus the
//!   paths it declares (profile `test_paths`). `run`'s oracle check and the tool-preflight lock
//!   read the same table.
//! - `moai tdd red|green|reopen|done|status`: the record `<git-dir>/moai/tdd.json` and the log
//!   `<git-dir>/moai/tdd-log.jsonl` live in the worktree's own git folder, so they are never
//!   committed and a linked worktree has its own. `red` needs a head commit with the trailer
//!   `Moai-Phase: red` that changes a test path, and checks that fail with assertion failures
//!   only ([`crate::run::red_confirmed`]); `green` reruns the same checks and needs them all to
//!   pass; `reopen --reason` ends the lock and logs why (leader decision D-6: no per-file
//!   unlock); `done` ends it. Records are written to a temporary file and renamed; a record, log
//!   or folder that is a symbolic link is refused.
//! - [`lock_refusal`]: the tool-preflight rule A6, read from the file system only (no program):
//!   while a record is GREEN or REFACTOR, a write to a path the RED commit changed, or a new file
//!   on a test path (leader decision D-7), is refused. A missing or unreadable record locks
//!   nothing.

use crate::run::{red_confirmed, ResolvedMethod};
use crate::verify::{self, CheckResult, CheckRun};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

type Result<T> = std::result::Result<T, String>;

/// One test-path rule, matched against a repository-relative path split on `/`.
pub enum Pattern {
    /// A folder named exactly this, anywhere in the path.
    Dir(&'static str),
    /// A folder whose name ends with this (`Tests`, `App.Tests`).
    DirEndsWith(&'static str),
    /// A file name that ends with this and is longer than it.
    Suffix(&'static str),
    /// A file name that holds this after its first character and before its last.
    Infix(&'static str),
    /// A file name with this start and this end, longer than both together.
    PrefixExt(&'static str, &'static str),
    /// A file named exactly this.
    Name(&'static str),
}

use Pattern::{Dir, DirEndsWith, Infix, Name, PrefixExt, Suffix};

const JS: &[Pattern] = &[
    Infix(".test."),
    Infix(".spec."),
    Dir("__tests__"),
    Dir("test"),
    Dir("tests"),
];
const C: &[Pattern] = &[Dir("test"), Dir("tests")];
const JVM: &[Pattern] = &[Dir("test")];
const DOTNET: &[Pattern] = &[DirEndsWith("Tests")];

/// Test paths per detected language (`crate::project_detect`), and the paths every language
/// shares. The language guides under `skills/moai/references/languages/` name the same
/// conventions.
pub const TEST_PATTERNS: &[(&str, &[Pattern])] = &[
    ("rust", &[Dir("tests"), Dir("benches")]),
    (
        "python",
        &[
            Dir("tests"),
            PrefixExt("test_", ".py"),
            Suffix("_test.py"),
            Name("conftest.py"),
        ],
    ),
    ("javascript", JS),
    ("typescript", JS),
    ("go", &[Suffix("_test.go"), Dir("testdata")]),
    ("java", JVM),
    ("kotlin", JVM),
    (
        "ruby",
        &[
            Dir("spec"),
            Dir("test"),
            Suffix("_spec.rb"),
            Suffix("_test.rb"),
        ],
    ),
    ("php", &[Dir("tests"), Suffix("Test.php")]),
    ("c", C),
    ("cpp", C),
    ("csharp", DOTNET),
    ("swift", DOTNET),
    ("common", &[Dir("fixtures"), Dir("golden")]),
];

/// The most declared test paths a profile holds, and the longest one.
pub const MAX_DECLARED: usize = 32;
const MAX_DECLARED_CHARS: usize = 200;

/// The parts of a repository-relative path: `\` read as `/`, a leading `./` and empty or `.`
/// parts dropped.
fn parts(path: &str) -> Vec<String> {
    path.replace('\\', "/")
        .split('/')
        .filter(|p| !p.is_empty() && *p != ".")
        .map(str::to_string)
        .collect()
}

fn matches(pattern: &Pattern, dirs: &[String], name: &str) -> bool {
    match pattern {
        Dir(d) => dirs.iter().any(|x| x == d),
        DirEndsWith(s) => dirs.iter().any(|x| x.ends_with(s)),
        Suffix(s) => name.len() > s.len() && name.ends_with(s),
        Infix(i) => name
            .find(i)
            .is_some_and(|at| at > 0 && at + i.len() < name.len()),
        PrefixExt(p, e) => {
            name.len() > p.len() + e.len() && name.starts_with(p) && name.ends_with(e)
        }
        Name(n) => name == *n,
    }
}

/// Whether `path` (relative to the repository root) is a test path of a project whose
/// languages are unknown: it matches a pattern of any row of [`TEST_PATTERNS`], or it is or
/// lies under one of the `declared` paths. A project whose languages are known uses
/// [`TestPaths::for_project`] instead.
pub fn is_test_path(path: &str, declared: &[String]) -> bool {
    test_path_in(path, None, declared)
}

/// The rows of `languages` and `common` (every row when `languages` is `None`), then the
/// declared paths.
fn test_path_in(path: &str, languages: Option<&[String]>, declared: &[String]) -> bool {
    let all = parts(path);
    let Some((name, dirs)) = all.split_last() else {
        return false;
    };
    let table = TEST_PATTERNS
        .iter()
        .filter(|(lang, _)| {
            *lang == "common" || languages.is_none_or(|l| l.iter().any(|x| x == lang))
        })
        .flat_map(|(_, patterns)| patterns.iter())
        .any(|p| matches(p, dirs, name));
    table
        || declared.iter().any(|d| {
            let d = parts(d);
            !d.is_empty() && all.len() >= d.len() && all[..d.len()] == d[..]
        })
}

/// The test paths of one project (green-test-lock audit N1): the [`TEST_PATTERNS`] rows of its
/// languages, the `common` row and its declared `test_paths`. The languages are the profile's
/// `detected.languages`; without them, the languages of the manifests at the project root
/// ([`crate::project_detect::manifest_languages`]); when neither names one, every row applies.
/// Files only, no program.
pub struct TestPaths {
    languages: Option<Vec<String>>,
    declared: Vec<String>,
}

impl TestPaths {
    pub fn for_project(root: &Path) -> Self {
        let profile = crate::read_profile(root).ok().map(|(p, _)| p);
        let declared = profile
            .as_ref()
            .and_then(|p| p.test_paths.clone())
            .unwrap_or_default();
        let languages = profile
            .and_then(|p| p.detected)
            .map(|d| d.languages)
            .filter(|l| !l.is_empty())
            .or_else(|| {
                let l = crate::project_detect::manifest_languages(root);
                (!l.is_empty()).then_some(l)
            });
        TestPaths {
            languages,
            declared,
        }
    }

    /// The languages whose rows apply, or `None` when every row does.
    pub fn languages(&self) -> Option<&[String]> {
        self.languages.as_deref()
    }

    pub fn matches(&self, path: &str) -> bool {
        test_path_in(path, self.languages.as_deref(), &self.declared)
    }
}

/// Declared test paths: at most [`MAX_DECLARED`], each relative, without `..`, a control
/// character or more than 200 characters.
pub fn validate_declared(paths: &[String]) -> Result<()> {
    let ok = paths.len() <= MAX_DECLARED
        && paths.iter().all(|p| {
            let p = p.trim();
            !p.is_empty()
                && p.chars().count() <= MAX_DECLARED_CHARS
                && !p.starts_with('/')
                && !p.starts_with('\\')
                && !p.contains(':')
                && !p.chars().any(char::is_control)
                && !parts(p).iter().any(|x| x == "..")
        });
    if ok {
        Ok(())
    } else {
        Err(
            "invalid test_paths: at most 32 relative paths without `..`; no values were logged"
                .into(),
        )
    }
}

/// The record file and the log, in `<git-dir>/moai/`.
pub const RECORD: &str = "tdd.json";
pub const LOG: &str = "tdd-log.jsonl";
const FOLDER: &str = "moai";
const MAX_RECORD_BYTES: u64 = 1024 * 1024;
const MAX_REASON_CHARS: usize = 500;
const MAX_CHECKS: usize = 8;

/// The phase record. `phase` is `green` until the RED checks pass, then `refactor`.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Record {
    pub schema_version: u32,
    pub spec: String,
    pub phase: String,
    pub red_commit: String,
    /// Test paths the RED commit changed, relative to the work tree.
    pub locked: Vec<String>,
    /// The commands whose assertion failures made the RED.
    pub checks: Vec<String>,
    pub at_unix: u64,
}

/// The nearest folder at or above `dir` that holds `.git` (a folder or a file), resolved.
pub fn work_tree_top(dir: &Path) -> Option<PathBuf> {
    let absolute = if dir.is_absolute() {
        dir.to_path_buf()
    } else {
        std::env::current_dir().ok()?.join(dir)
    };
    let mut d = crate::hook_config::canonical(&absolute);
    loop {
        if fs::symlink_metadata(d.join(".git")).is_ok() {
            return Some(d);
        }
        if !d.pop() {
            return None;
        }
    }
}

/// The work tree's own git folder: `.git` itself, or the folder a `.git` file names after
/// `gitdir:` (a linked worktree). Read from the file system only.
pub fn git_dir(top: &Path) -> Option<PathBuf> {
    let git = top.join(".git");
    if git.is_dir() {
        return Some(git);
    }
    let pointer = fs::read_to_string(&git).ok()?;
    let target = PathBuf::from(pointer.trim().strip_prefix("gitdir:")?.trim());
    Some(if target.is_absolute() {
        target
    } else {
        top.join(target)
    })
}

fn is_link(path: &Path) -> bool {
    fs::symlink_metadata(path).is_ok_and(|m| m.file_type().is_symlink())
}

fn valid_record(r: &Record) -> bool {
    r.schema_version == 1
        && matches!(r.phase.as_str(), "green" | "refactor")
        && r.red_commit.len() == 40
        && r.red_commit.bytes().all(|b| b.is_ascii_hexdigit())
}

/// The record of `git_dir`, or `None` when there is none, it cannot be read, it is a link, its
/// folder is a link, or it is not a valid record.
pub fn read_record(git_dir: &Path) -> Option<Record> {
    let folder = git_dir.join(FOLDER);
    let file = folder.join(RECORD);
    if is_link(&folder) || is_link(&file) {
        return None;
    }
    let meta = fs::metadata(&file).ok()?;
    if !meta.is_file() || meta.len() > MAX_RECORD_BYTES {
        return None;
    }
    let record: Record = serde_json::from_slice(&fs::read(&file).ok()?).ok()?;
    valid_record(&record).then_some(record)
}

const LOCKED: &str = "MoAI tool-preflight: the TDD cycle of this worktree is in GREEN or REFACTOR; the tests of its RED commit and new test files are locked. End the cycle first: `moai tdd done` when it is finished, or `moai tdd reopen --reason <why>` to change a test.";

/// A6: why a write of `paths` (relative to `cwd`, or absolute) is refused while this
/// worktree's record is GREEN or REFACTOR, or `None`. Reads files only.
pub fn lock_refusal(paths: &[String], cwd: &Path) -> Option<&'static str> {
    // #3177 audit N1: each path is judged by the record of the work tree that holds it (the
    // nearest `.git` above it, as A5), not by the session folder's work tree; its test paths
    // are that work tree's ([`TestPaths::for_project`]).
    let refused = paths.iter().any(|p| {
        let path = Path::new(p);
        let absolute = if path.is_absolute() {
            path.to_path_buf()
        } else {
            cwd.join(path)
        };
        let resolved = crate::hook_config::canonical(&absolute);
        let Some(top) = work_tree_top(&resolved) else {
            return false;
        };
        let Some(record) = git_dir(&top).and_then(|g| read_record(&g)) else {
            return false;
        };
        let Ok(rel) = resolved.strip_prefix(&top) else {
            return false;
        };
        let rel = rel.to_string_lossy().replace('\\', "/");
        record.locked.contains(&rel)
            || (TestPaths::for_project(&top).matches(&rel)
                && fs::symlink_metadata(&resolved).is_err())
    });
    refused.then_some(LOCKED)
}

// ---- the `moai tdd` command ----

fn now() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_or(0, |d| d.as_secs())
}

fn git(top: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .args(args)
        .current_dir(top)
        .output()
        .map_err(|_| "git could not be started".to_string())?;
    if !out.status.success() {
        return Err(format!("git {} failed", args.first().unwrap_or(&"")));
    }
    Ok(String::from_utf8_lossy(&out.stdout).into_owned())
}

/// The work tree, its git folder and the SPEC id of its branch `spec/<yyyymmdd>-<slug>`.
fn locate(project: &Path) -> Result<(PathBuf, PathBuf, String)> {
    let top = work_tree_top(project).ok_or("not inside a git work tree")?;
    let gd = git_dir(&top).ok_or("the work tree's git folder cannot be read")?;
    let head = fs::read_to_string(gd.join("HEAD")).map_err(|_| "HEAD cannot be read")?;
    let branch = head
        .trim()
        .strip_prefix("ref: refs/heads/")
        .ok_or("HEAD is detached; a TDD cycle runs on a SPEC work branch")?;
    let spec = spec_of_branch(branch)
        .ok_or("the branch is not a SPEC work branch spec/<yyyymmdd>-<slug>")?;
    Ok((top, gd, spec))
}

/// `spec/<yyyymmdd>-<slug>` → `SPEC-<yyyymmdd>-<slug>`.
fn spec_of_branch(branch: &str) -> Option<String> {
    let (date, slug) = branch.strip_prefix("spec/")?.split_once('-')?;
    let ok = date.len() == 8
        && date.bytes().all(|b| b.is_ascii_digit())
        && crate::spec_doc::valid_slug(slug);
    ok.then(|| format!("SPEC-{date}-{slug}"))
}

/// Whether the last paragraph of a commit message holds the line `Moai-Phase: red`.
fn has_red_trailer(message: &str) -> bool {
    message
        .trim_end()
        .rsplit("\n\n")
        .next()
        .is_some_and(|block| block.lines().any(|l| l.trim_end() == "Moai-Phase: red"))
}

fn folder(gd: &Path) -> Result<PathBuf> {
    let f = gd.join(FOLDER);
    if is_link(&f) {
        return Err("the record folder is a symbolic link; refused".into());
    }
    if !f.exists() {
        fs::create_dir(&f).map_err(|_| "the record folder cannot be created")?;
    }
    if !f.is_dir() {
        return Err("the record folder is not a folder; refused".into());
    }
    Ok(f)
}

fn write_record(gd: &Path, record: &Record) -> Result<()> {
    let f = folder(gd)?;
    let target = f.join(RECORD);
    if is_link(&target) {
        return Err("the record is a symbolic link; refused".into());
    }
    let tmp = f.join(format!("{RECORD}.tmp-{}", std::process::id()));
    let _ = fs::remove_file(&tmp);
    let mut file = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(&tmp)
        .map_err(|_| "the record cannot be written")?;
    let bytes = serde_json::to_vec_pretty(record).expect("typed record");
    file.write_all(&bytes)
        .and_then(|_| file.sync_all())
        .map_err(|_| "the record cannot be written")?;
    fs::rename(&tmp, &target).map_err(|_| "the record cannot be written".into())
}

fn remove_record(gd: &Path) -> Result<()> {
    let f = folder(gd)?;
    let target = f.join(RECORD);
    if is_link(&target) {
        return Err("the record is a symbolic link; refused".into());
    }
    fs::remove_file(&target).map_err(|_| "the record cannot be removed".into())
}

fn append_log(gd: &Path, entry: &Value) -> Result<()> {
    let f = folder(gd)?;
    let log = f.join(LOG);
    if is_link(&log) {
        return Err("the log is a symbolic link; refused".into());
    }
    let mut file = fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(&log)
        .map_err(|_| "the log cannot be written")?;
    let mut line = entry.to_string();
    line.push('\n');
    file.write_all(line.as_bytes())
        .map_err(|_| "the log cannot be written".into())
}

fn tdd_method() -> ResolvedMethod {
    ResolvedMethod {
        change_strategy: "tdd".into(),
        domain_modeling: "not-required".into(),
        legacy_mode: None,
    }
}

fn run_checks(top: &Path, checks: &[String], timeout: Duration) -> Vec<CheckRun> {
    checks
        .iter()
        .enumerate()
        .map(|(i, c)| verify::run_check(top, i, c, timeout))
        .collect()
}

fn check_summary(runs: &[CheckRun]) -> Value {
    json!(runs
        .iter()
        .map(|r| json!({
            "command": r.command, "result": r.result, "exit_code": r.exit_code,
            "failure_class": r.failure_class,
        }))
        .collect::<Vec<_>>())
}

/// `moai tdd red`: records the head commit as the RED of this worktree's cycle.
pub fn red(project: &Path, checks: &[String], timeout: Duration) -> Result<Value> {
    if checks.is_empty() || checks.len() > MAX_CHECKS {
        return Err(
            "tdd red refused: name 1 to 8 checks with --check; nothing was recorded".into(),
        );
    }
    let (top, gd, spec) = locate(project).map_err(|e| format!("tdd red refused: {e}"))?;
    if read_record(&gd).is_some() {
        return Err("tdd red refused: a cycle is already open; end it with `moai tdd done` or `moai tdd reopen --reason <why>`".into());
    }
    let head = git(&top, &["rev-parse", "HEAD"]).map_err(|e| format!("tdd red refused: {e}"))?;
    let head = head.trim().to_string();
    let message = git(&top, &["log", "-1", "--format=%B", "HEAD"])
        .map_err(|e| format!("tdd red refused: {e}"))?;
    if !has_red_trailer(&message) {
        return Err("tdd red refused: the head commit has no `Moai-Phase: red` trailer; nothing was recorded".into());
    }
    let changed = git(
        &top,
        &[
            "diff-tree",
            "--no-commit-id",
            "--name-only",
            "-r",
            "--root",
            "HEAD",
        ],
    )
    .map_err(|e| format!("tdd red refused: {e}"))?;
    let tests = TestPaths::for_project(&top);
    let locked: Vec<String> = changed
        .lines()
        .filter(|p| tests.matches(p))
        .map(str::to_string)
        .collect();
    if locked.is_empty() {
        return Err(
            "tdd red refused: the head commit changes no test path; nothing was recorded".into(),
        );
    }
    let runs = run_checks(&top, checks, timeout);
    if red_confirmed(&tdd_method(), Some("red"), &runs) != Some(true) {
        return Err(format!(
            "tdd red refused: the checks did not fail with assertion failures only ({}); nothing \
             was recorded",
            check_summary(&runs)
        ));
    }
    let record = Record {
        schema_version: 1,
        spec: spec.clone(),
        phase: "green".into(),
        red_commit: head.clone(),
        locked: locked.clone(),
        checks: checks.to_vec(),
        at_unix: now(),
    };
    write_record(&gd, &record).map_err(|e| format!("tdd red refused: {e}"))?;
    append_log(
        &gd,
        &json!({"at_unix": record.at_unix, "event": "red", "spec": spec, "commit": head, "locked": locked}),
    )?;
    Ok(
        json!({"action": "red", "phase": "green", "record": record, "checks": check_summary(&runs), "writes": true}),
    )
}

/// `moai tdd green`: moves the record to REFACTOR when every RED check passes at the head.
pub fn green(project: &Path, timeout: Duration) -> Result<Value> {
    let (top, gd, _) = locate(project).map_err(|e| format!("tdd green refused: {e}"))?;
    let mut record = read_record(&gd)
        .ok_or("tdd green refused: no open TDD cycle (run `moai tdd red` first)")?;
    if record.phase != "green" {
        return Err("tdd green refused: the cycle is already in REFACTOR".into());
    }
    let runs = run_checks(&top, &record.checks, timeout);
    if !runs.iter().all(|r| r.result == CheckResult::Passed) {
        return Err(format!(
            "tdd green refused: not every RED check passes ({}); the cycle stays in GREEN",
            check_summary(&runs)
        ));
    }
    let head = git(&top, &["rev-parse", "HEAD"]).map_err(|e| format!("tdd green refused: {e}"))?;
    record.phase = "refactor".into();
    record.at_unix = now();
    write_record(&gd, &record).map_err(|e| format!("tdd green refused: {e}"))?;
    append_log(
        &gd,
        &json!({"at_unix": record.at_unix, "event": "green", "spec": record.spec, "commit": head.trim()}),
    )?;
    Ok(
        json!({"action": "green", "phase": "refactor", "record": record, "checks": check_summary(&runs), "writes": true}),
    )
}

/// `moai tdd reopen --reason`: ends the lock and logs why (leader decision D-6).
pub fn reopen(project: &Path, reason: &str) -> Result<Value> {
    let reason = reason.trim();
    if reason.is_empty()
        || reason.chars().count() > MAX_REASON_CHARS
        || reason.chars().any(char::is_control)
    {
        return Err("tdd reopen refused: --reason is one line of 1 to 500 characters".into());
    }
    let (_, gd, _) = locate(project).map_err(|e| format!("tdd reopen refused: {e}"))?;
    let record = read_record(&gd).ok_or("tdd reopen refused: no open TDD cycle")?;
    remove_record(&gd).map_err(|e| format!("tdd reopen refused: {e}"))?;
    append_log(
        &gd,
        &json!({"at_unix": now(), "event": "reopen", "spec": record.spec, "commit": record.red_commit, "reason": reason}),
    )?;
    Ok(
        json!({"action": "reopen", "phase": null, "spec": record.spec, "reason": reason, "writes": true}),
    )
}

/// `moai tdd done`: ends the cycle; the log stays.
pub fn done(project: &Path) -> Result<Value> {
    let (_, gd, _) = locate(project).map_err(|e| format!("tdd done refused: {e}"))?;
    let record = read_record(&gd).ok_or("tdd done refused: no open TDD cycle")?;
    remove_record(&gd).map_err(|e| format!("tdd done refused: {e}"))?;
    append_log(
        &gd,
        &json!({"at_unix": now(), "event": "done", "spec": record.spec, "commit": record.red_commit}),
    )?;
    Ok(json!({"action": "done", "phase": null, "spec": record.spec, "writes": true}))
}

/// `moai tdd status`: the record (or `null`) and the log, read only.
pub fn status(project: &Path) -> Result<Value> {
    let top = work_tree_top(project).ok_or("not inside a git work tree")?;
    let gd = git_dir(&top).ok_or("the work tree's git folder cannot be read")?;
    let log_path = gd.join(FOLDER).join(LOG);
    let log: Vec<Value> = if is_link(&log_path) {
        Vec::new()
    } else {
        fs::read_to_string(&log_path)
            .unwrap_or_default()
            .lines()
            .filter_map(|l| serde_json::from_str(l).ok())
            .collect()
    };
    let record = read_record(&gd);
    // Audit N2: a log whose last event opened or advanced a cycle while no record can be read
    // means the record was removed or damaged outside `moai tdd`. Reported only; nothing locks.
    let last = log
        .last()
        .and_then(|e| e["event"].as_str())
        .map(str::to_string);
    let open_in_log = matches!(last.as_deref(), Some("red" | "green"));
    let problem = (open_in_log && record.is_none()).then(|| {
        format!(
            "the log's last event is `{}` but the record is missing or unreadable; nothing is \
             locked. Start again with `moai tdd red` on a new RED commit",
            last.as_deref().unwrap_or_default()
        )
    });
    Ok(json!({
        "action": "status",
        "phase": record.as_ref().map(|r| r.phase.clone()),
        "record": record,
        "log": log,
        "problem": problem,
        "writes": false,
    }))
}
