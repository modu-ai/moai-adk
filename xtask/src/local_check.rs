//! `cargo xtask local-check`: the Rust form of `tools/local_check.sh` (SPEC-20261121-xtask-local-check).
//!
//! It runs the check steps of the CI `check` job, in the same order with the same arguments, in a
//! candidate (the head merged with the latest tip of the base branch) or in place (CI), reads the
//! same environment variables, prints the same lines and exits with the same code. Until the
//! callers move to this tool the texts stay the script's own, `tools/local_check.sh` included: the
//! merge gate reads `cmd=tools/local_check.sh` in the final line (`src/gate.rs`). The script's
//! header comment describes every mode and variable; `xtask/tests/local_check_parity.rs` runs both
//! on the same repository and compares them.
//!
//! One difference by design: the `script=` field is the `git hash-object` of this file, the code
//! that ran, where the script hashes its own file.

use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::sync::atomic::{AtomicBool, Ordering};

/// What the script calls itself in every message.
const NAME: &str = "tools/local_check.sh";
const DEFAULT_BASE_BRANCH: &str = "v2026.11.0";
const REMOTE: &str = "origin";
/// The file the `script=` field hashes.
const SCRIPT_SOURCE: &str = "xtask/src/local_check.rs";

/// Set by the INT and TERM handlers: the script's `trap` ends the run after the running command.
static SIGNALLED: AtomicBool = AtomicBool::new(false);

#[cfg(unix)]
mod signals {
    use super::{Ordering, SIGNALLED};

    extern "C" {
        fn signal(signum: i32, handler: usize) -> usize;
    }

    extern "C" fn on_signal(_: i32) {
        SIGNALLED.store(true, Ordering::SeqCst);
    }

    pub fn install() {
        const SIGINT: i32 = 2;
        const SIGTERM: i32 = 15;
        // SAFETY: `signal` is given a handler that only stores to an atomic.
        unsafe {
            signal(SIGINT, on_signal as *const () as usize);
            signal(SIGTERM, on_signal as *const () as usize);
        }
    }
}

#[cfg(not(unix))]
mod signals {
    pub fn install() {}
}

fn say(line: &str) {
    println!("{line}");
}

fn warn(text: &str) {
    eprintln!("{NAME}: {text}");
}

/// `${NAME:-default}`: a variable that is unset or empty is the default.
fn env_or(name: &str, default: &str) -> String {
    match std::env::var(name) {
        Ok(v) if !v.is_empty() => v,
        _ => default.to_string(),
    }
}

fn env_set(name: &str) -> bool {
    std::env::var(name).is_ok_and(|v| !v.is_empty())
}

fn is_hex40(text: &str) -> bool {
    text.len() == 40
        && text
            .bytes()
            .all(|b| b.is_ascii_digit() || (b'a'..=b'f').contains(&b))
}

/// A branch name this script accepts: `[A-Za-z0-9._/-]`, no leading `-`, no `..`, not empty.
fn valid_branch(name: &str) -> bool {
    !name.is_empty()
        && !name.starts_with('-')
        && !name.contains("..")
        && name
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'/' | b'-'))
}

/// The exit code a shell reports for a finished command: its own, or 128 plus the signal.
fn shell_code(status: std::process::ExitStatus) -> i32 {
    if let Some(code) = status.code() {
        return code;
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

/// A command that could not start, as a shell reports it.
fn spawn_code(err: &std::io::Error) -> i32 {
    if err.kind() == std::io::ErrorKind::NotFound {
        127
    } else {
        126
    }
}

/// `git` with `args` in `cwd`, output captured, stdin closed. `None` when git cannot start.
fn git(cwd: &Path, args: &[&str]) -> Option<Output> {
    Command::new("git")
        .args(args)
        .current_dir(cwd)
        .stdin(Stdio::null())
        .output()
        .ok()
}

fn text_of(out: &Output) -> String {
    String::from_utf8_lossy(&out.stdout)
        .trim_end_matches('\n')
        .to_string()
}

/// Whether `name` is an executable file in a folder of `PATH` (`command -v`).
fn on_path(name: &str) -> bool {
    let Some(path) = std::env::var_os("PATH") else {
        return false;
    };
    std::env::split_paths(&path).any(|dir| {
        std::fs::metadata(dir.join(name)).is_ok_and(|m| m.is_file() && is_executable(&m))
    })
}

#[cfg(unix)]
fn is_executable(meta: &std::fs::Metadata) -> bool {
    use std::os::unix::fs::PermissionsExt;
    meta.permissions().mode() & 0o111 != 0
}

#[cfg(not(unix))]
fn is_executable(_: &std::fs::Metadata) -> bool {
    true
}

#[derive(Clone, Copy, PartialEq)]
enum Weight {
    Heavy,
    Python,
    Light,
}

/// A signal ended the run: the final line is still printed, as a fail.
struct Aborted;

struct Check {
    root: PathBuf,
    /// The folder the steps run in: the root, or the candidate once it is checked out.
    cwd: PathBuf,
    candidate_mode: bool,
    leg: String,
    dry_run: bool,
    keep_going: bool,
    base_branch: String,
    head: String,
    diff_range: String,
    slot_max: String,
    slot_min_free: String,
    slot_name: String,
    py_slot_max: String,
    py_slot_dir: String,
    use_slot: bool,
    dirty: bool,
    failed: bool,
    stopped: bool,
    base_sha: String,
    tree_sha: String,
    rustc_tok: String,
    tmp_parent: Option<PathBuf>,
    cand: PathBuf,
    cand_added: bool,
}

/// The script's whole run; the return value is its exit code.
pub fn run() -> i32 {
    let root = match git(Path::new("."), &["rev-parse", "--show-toplevel"]) {
        Some(out) if out.status.success() => PathBuf::from(text_of(&out)),
        _ => {
            warn("cannot find the repository root; run it inside the repository");
            return 2;
        }
    };

    let mode = env_or("LOCAL_CHECK_MODE", "candidate");
    if mode != "candidate" && mode != "in-place" {
        warn(&format!(
            "LOCAL_CHECK_MODE must be candidate or in-place, not '{mode}'"
        ));
        return 2;
    }
    let candidate_mode = mode == "candidate";

    let leg = env_or("LOCAL_CHECK_LEG", "");
    if !matches!(leg.as_str(), "" | "nextest" | "rest") {
        warn(&format!(
            "LOCAL_CHECK_LEG must be unset, 'nextest' or 'rest', not '{leg}'"
        ));
        return 2;
    }
    if !leg.is_empty() {
        if candidate_mode {
            warn("LOCAL_CHECK_LEG is a CI leg for in-place mode; candidate mode always checks the full set, unset LOCAL_CHECK_LEG");
            return 2;
        }
        if std::env::var("GITHUB_ACTIONS").as_deref() != Ok("true") {
            warn("LOCAL_CHECK_LEG is for CI (in-place under GITHUB_ACTIONS); a local run always checks the full set, unset LOCAL_CHECK_LEG");
            return 2;
        }
    }

    let base_branch = env_or("LOCAL_CHECK_BASE_BRANCH", DEFAULT_BASE_BRANCH);
    let head = match git(&root, &["rev-parse", "HEAD"]) {
        Some(out) if out.status.success() => text_of(&out),
        _ => "unknown".to_string(),
    };

    let diff_range;
    if !candidate_mode {
        let base = if env_set("LOCAL_CHECK_BASE") {
            env_or("LOCAL_CHECK_BASE", "")
        } else if git(
            &root,
            &["rev-parse", "--verify", "-q", "origin/v2026.11.0^{commit}"],
        )
        .is_some_and(|o| o.status.success())
        {
            "origin/v2026.11.0".to_string()
        } else {
            "HEAD~1".to_string()
        };
        let diff_head = env_or("LOCAL_CHECK_DIFF_HEAD", "HEAD");
        diff_range = format!("{base}...{diff_head}");
        if std::env::var("GITHUB_ACTIONS").as_deref() != Ok("true") {
            warn("LOCAL_CHECK_MODE=in-place checks only this working folder and prints the short line; it is for CI, a local run should use the default candidate mode");
        }
    } else {
        if env_set("LOCAL_CHECK_BASE") || env_set("LOCAL_CHECK_DIFF_HEAD") {
            warn(&format!(
                "candidate mode ignores LOCAL_CHECK_BASE and LOCAL_CHECK_DIFF_HEAD; it checks the head merged with the latest {REMOTE}/{base_branch}"
            ));
        }
        diff_range = "<base>...<head>".to_string();
    }

    let py_slot_dir = match env_or("MOAI_PYTHON_SLOT_DIR", "") {
        dir if !dir.is_empty() => dir,
        _ if env_set("XDG_CACHE_HOME") => {
            format!("{}/moai-cli/python-slots", env_or("XDG_CACHE_HOME", ""))
        }
        _ if env_set("HOME") => format!("{}/.cache/moai-cli/python-slots", env_or("HOME", "")),
        _ => String::new(),
    };
    let use_slot = on_path("moai")
        && Command::new("moai")
            .args(["heavy-slot-acquire", "--help"])
            .stdin(Stdio::null())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .is_ok_and(|s| s.success());

    let mut check = Check {
        cwd: root.clone(),
        root,
        candidate_mode,
        leg,
        dry_run: std::env::var("LOCAL_CHECK_DRY_RUN")
            .as_deref()
            .unwrap_or("0")
            == "1",
        keep_going: std::env::var("LOCAL_CHECK_KEEP_GOING")
            .as_deref()
            .unwrap_or("0")
            == "1",
        base_branch,
        head,
        diff_range,
        slot_max: env_or("MOAI_HEAVY_SLOT_MAX", "3"),
        slot_min_free: env_or("MOAI_HEAVY_SLOT_MIN_FREE_GB", "50"),
        slot_name: env_or("MOAI_HEAVY_SLOT_NAME", "local-check"),
        py_slot_max: env_or("MOAI_PYTHON_SLOT_MAX", "2"),
        py_slot_dir,
        use_slot,
        dirty: false,
        failed: false,
        stopped: false,
        base_sha: "none".into(),
        tree_sha: "none".into(),
        rustc_tok: "unknown".into(),
        tmp_parent: None,
        cand: PathBuf::new(),
        cand_added: false,
    };

    signals::install();
    // Any unexpected abort still ends with the line, as a fail.
    let body = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| check.body()));
    if !matches!(body, Ok(Ok(()))) {
        check.failed = true;
    }
    check.finish()
}

impl Check {
    fn tree_is_dirty(&self) -> bool {
        match git(
            &self.root,
            &["status", "--porcelain", "--untracked-files=normal"],
        ) {
            Some(out) if out.status.success() => !out.stdout.is_empty(),
            // The script reads a failing `git status` as the word `error`: not empty.
            _ => true,
        }
    }

    fn signalled() -> Result<(), Aborted> {
        if SIGNALLED.load(Ordering::SeqCst) {
            Err(Aborted)
        } else {
            Ok(())
        }
    }

    fn body(&mut self) -> Result<(), Aborted> {
        if self.tree_is_dirty() {
            self.dirty = true;
            if !self.candidate_mode {
                warn("the tree has uncommitted changes; the result will be fail (dirty=1)");
            } else {
                warn("the tree has uncommitted changes and candidate mode checks only the committed HEAD; commit or stash them first. The result is fail.");
                self.stopped = true;
            }
        }

        if self.candidate_mode && !self.stopped {
            if self.dry_run {
                let _ = self.read_rustc();
            } else if !self.candidate_setup() {
                self.failed = true;
                self.stopped = true;
            }
            Self::signalled()?;
        }

        let nextest = self.on_leg(&["nextest", "rest"]);
        if nextest {
            self.step(Weight::Light, "fetch", &["cargo", "fetch", "--locked"])?;
        }
        if self.on_leg(&["nextest"]) {
            self.step(
                Weight::Heavy,
                "test",
                &[
                    "bash",
                    "-c",
                    "RUST_MIN_STACK=8388608 cargo nextest run --offline --locked --no-fail-fast 2>&1 | tee ci-test.log; exit \"${PIPESTATUS[0]}\"",
                ],
            )?;
            self.step(
                Weight::Heavy,
                "doctest",
                &[
                    "bash",
                    "-c",
                    "RUST_MIN_STACK=8388608 cargo test --offline --locked --doc 2>&1 | tee ci-doctest.log; exit \"${PIPESTATUS[0]}\"",
                ],
            )?;
        }
        if self.on_leg(&["rest"]) {
            self.step(
                Weight::Heavy,
                "clippy",
                &[
                    "cargo",
                    "clippy",
                    "--offline",
                    "--locked",
                    "--all-targets",
                    "--",
                    "-D",
                    "warnings",
                ],
            )?;
            self.step(Weight::Light, "fmt", &["cargo", "fmt", "--check"])?;
            self.step(
                Weight::Light,
                "fmt-ops",
                &["bash", "-c", "rustfmt --check --edition 2021 src/ops/*.rs"],
            )?;
            self.step(
                Weight::Python,
                "unittest",
                &[
                    "bash",
                    "-c",
                    "python3 -m unittest discover -v -s tests -p '*_test.py' 2>&1 | tee ci-unittest.log; exit \"${PIPESTATUS[0]}\"",
                ],
            )?;
            let range = self.diff_range.clone();
            self.step(
                Weight::Light,
                "diff-check",
                &["git", "diff", "--check", &range],
            )?;
        }
        Ok(())
    }

    /// Whether the selected leg is in `legs`; the unset leg matches every list.
    fn on_leg(&self, legs: &[&str]) -> bool {
        self.leg.is_empty() || legs.contains(&self.leg.as_str())
    }

    /// The version word of `rustc --version` (`rustc 1.91.0 (...)` is `1.91.0`), or failure.
    fn read_rustc(&mut self) -> bool {
        let out = Command::new("rustc")
            .arg("--version")
            .current_dir(&self.cwd)
            .stdin(Stdio::null())
            .stderr(Stdio::null())
            .output();
        let Ok(out) = out else {
            return false;
        };
        if !out.status.success() {
            return false;
        }
        let text = text_of(&out);
        let first = text.lines().next().unwrap_or("");
        let tok = first.split_whitespace().nth(1).unwrap_or("");
        let fine = !tok.is_empty()
            && tok.len() <= 32
            && tok
                .bytes()
                .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'+' | b'-'));
        if fine {
            self.rustc_tok = tok.to_string();
        }
        fine
    }

    /// Candidate mode: fetch the base tip, merge it with the head in the object database only,
    /// check the merge out as a detached worktree and enter it. `false` when it cannot; the
    /// evidence fields then tell what is known.
    fn candidate_setup(&mut self) -> bool {
        let branch = self.base_branch.clone();
        if !valid_branch(&branch) {
            warn(&format!(
                "LOCAL_CHECK_BASE_BRANCH '{branch}' is not a branch name this script accepts ([A-Za-z0-9._/-], no leading '-', no '..')"
            ));
            return false;
        }
        if self.head == "unknown" {
            warn("cannot read HEAD");
            return false;
        }
        let refspec = format!("+refs/heads/{branch}:refs/remotes/{REMOTE}/{branch}");
        let fetched = Command::new("git")
            .args([
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
                "fetch",
                "--quiet",
                "--no-write-fetch-head",
                "--no-tags",
                REMOTE,
                &refspec,
            ])
            .env("GIT_TERMINAL_PROMPT", "0")
            .current_dir(&self.cwd)
            .stdin(Stdio::null())
            .status();
        if !fetched.is_ok_and(|s| s.success()) {
            warn(&format!(
                "fetching {REMOTE}/{branch} failed; the base would be stale, so the result is fail"
            ));
            return false;
        }
        let tip = format!("refs/remotes/{REMOTE}/{branch}^{{commit}}");
        self.base_sha = match git(&self.cwd, &["rev-parse", "--verify", "-q", &tip]) {
            Some(out) if out.status.success() => text_of(&out),
            _ => "none".into(),
        };
        if !is_hex40(&self.base_sha) {
            warn(&format!(
                "{REMOTE}/{branch} is not a commit after the fetch"
            ));
            self.base_sha = "none".into();
            return false;
        }
        let merged = Command::new("git")
            .args(["merge-tree", "--write-tree", &self.base_sha, &self.head])
            .current_dir(&self.cwd)
            .stdin(Stdio::null())
            .stderr(Stdio::inherit())
            .output();
        let (code, merge_out) = match &merged {
            Ok(out) => (shell_code(out.status), text_of(out)),
            Err(e) => (spawn_code(e), String::new()),
        };
        if code == 1 {
            warn(&format!(
                "the head conflicts with {REMOTE}/{branch} ({}); merge-tree says:",
                self.base_sha
            ));
            for line in merge_out.split('\n').skip(1) {
                eprintln!("{line}");
            }
            return false;
        } else if code != 0 {
            warn(&format!("git merge-tree failed (exit {code})"));
            return false;
        }
        self.tree_sha = merge_out.lines().next().unwrap_or("").to_string();
        if !is_hex40(&self.tree_sha) {
            warn("git merge-tree did not print a tree");
            self.tree_sha = "none".into();
            return false;
        }
        let message = format!("local-check candidate: {} on {}", self.head, self.base_sha);
        let commit = Command::new("git")
            .args([
                "-c",
                "commit.gpgSign=false",
                "commit-tree",
                &self.tree_sha,
                "-p",
                &self.base_sha,
                "-p",
                &self.head,
                "-m",
                &message,
            ])
            .env("GIT_AUTHOR_NAME", "local-check")
            .env("GIT_AUTHOR_EMAIL", "local-check@invalid")
            .env("GIT_COMMITTER_NAME", "local-check")
            .env("GIT_COMMITTER_EMAIL", "local-check@invalid")
            .current_dir(&self.cwd)
            .stdin(Stdio::null())
            .stderr(Stdio::inherit())
            .output();
        let cand_commit = match commit {
            Ok(out) if out.status.success() && is_hex40(&text_of(&out)) => text_of(&out),
            _ => {
                warn("could not create the candidate commit");
                self.tree_sha = "none".into();
                return false;
            }
        };
        // Folders a killed run left behind are only reported: another run may be using one.
        let tmpdir = env_or("TMPDIR", "/tmp");
        let mut leftovers: Vec<PathBuf> = std::fs::read_dir(&tmpdir)
            .map(|rd| {
                rd.flatten()
                    .filter(|e| {
                        e.file_name()
                            .to_string_lossy()
                            .starts_with("local-check-cand.")
                    })
                    .map(|e| e.path())
                    .collect()
            })
            .unwrap_or_default();
        leftovers.sort();
        for leftover in leftovers {
            if leftover.is_dir() {
                warn(&format!(
                    "an earlier candidate folder exists: '{}' (not deleted; remove it yourself if no run uses it)",
                    leftover.display()
                ));
            }
        }
        let template = format!("{tmpdir}/local-check-cand.XXXXXX");
        let made = Command::new("mktemp")
            .args(["-d", &template])
            .stdin(Stdio::null())
            .stderr(Stdio::inherit())
            .output();
        match made {
            Ok(out) if out.status.success() => {
                self.tmp_parent = Some(PathBuf::from(text_of(&out)));
            }
            _ => {
                self.tmp_parent = None;
                warn("could not create a temporary folder");
                self.tree_sha = "none".into();
                return false;
            }
        }
        self.cand = self.tmp_parent.clone().unwrap().join("candidate");
        let cand_arg = self.cand.to_string_lossy().into_owned();
        let added = Command::new("git")
            .args([
                "worktree",
                "add",
                "--detach",
                "--quiet",
                &cand_arg,
                &cand_commit,
            ])
            .current_dir(&self.cwd)
            .stdin(Stdio::null())
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status();
        if !added.is_ok_and(|s| s.success()) {
            warn(&format!(
                "could not check the candidate out into '{cand_arg}'"
            ));
            self.tree_sha = "none".into();
            return false;
        }
        self.cand_added = true;
        self.cwd = self.cand.clone();
        say(&format!(
            "candidate: head {} merged onto {REMOTE}/{branch} {}: tree {}, commit {cand_commit}, folder {cand_arg}",
            self.head, self.base_sha, self.tree_sha
        ));
        if !self.read_rustc() {
            warn("cannot read 'rustc --version' inside the candidate; the result is fail");
            return false;
        }
        self.diff_range = format!("{}...{}", self.base_sha, self.head);
        true
    }

    /// Runs `argv` in the folder the check is in, in a slot when it is heavy or python and a
    /// slot command exists; returns the exit code a shell would report.
    fn run_step(&self, weight: Weight, label: &str, argv: &[&str]) -> i32 {
        let direct = |argv: &[&str]| -> i32 {
            match Command::new(argv[0])
                .args(&argv[1..])
                .current_dir(&self.cwd)
                .status()
            {
                Ok(status) => shell_code(status),
                Err(e) => spawn_code(&e),
            }
        };
        let slot_args = |dir: Option<&str>, max: &str, name: &str| -> Vec<String> {
            let mut v: Vec<String> = vec!["moai".into(), "heavy-slot-acquire".into()];
            if let Some(dir) = dir {
                v.extend(["--dir".into(), dir.into()]);
            }
            v.extend([
                "--max".into(),
                max.into(),
                "--min-free-gb".into(),
                self.slot_min_free.clone(),
                "--name".into(),
                name.into(),
                "--".into(),
            ]);
            v
        };
        let with_slot = |mut prefix: Vec<String>| -> Vec<String> {
            prefix.extend(argv.iter().map(|a| a.to_string()));
            prefix
        };
        if weight == Weight::Heavy && self.use_slot {
            let full = with_slot(slot_args(None, &self.slot_max, &self.slot_name));
            let refs: Vec<&str> = full.iter().map(String::as_str).collect();
            direct(&refs)
        } else if weight == Weight::Python && self.use_slot && self.py_slot_dir.is_empty() {
            // No folder can be named: never fall back to the cargo slot folder; say so and run directly.
            warn(&format!(
                "no python slot folder (MOAI_PYTHON_SLOT_DIR, XDG_CACHE_HOME and HOME are all empty); step {label} runs without a slot"
            ));
            direct(argv)
        } else if weight == Weight::Python && self.use_slot {
            let name = format!("{}-py", self.slot_name);
            let full = with_slot(slot_args(Some(&self.py_slot_dir), &self.py_slot_max, &name));
            let refs: Vec<&str> = full.iter().map(String::as_str).collect();
            direct(&refs)
        } else {
            direct(argv)
        }
    }

    fn step(&mut self, weight: Weight, label: &str, argv: &[&str]) -> Result<(), Aborted> {
        if self.stopped {
            return Ok(());
        }
        say(&format!("step {label}: {}", argv.join(" ")));
        if self.dry_run {
            return Ok(());
        }
        let code = self.run_step(weight, label, argv);
        Self::signalled()?;
        if let Some(summary) = std::env::var_os("GITHUB_STEP_SUMMARY").filter(|v| !v.is_empty()) {
            if let Ok(mut file) = std::fs::OpenOptions::new()
                .append(true)
                .create(true)
                .open(summary)
            {
                let _ = writeln!(file, "| {label} | `{}` | {code} |", argv.join(" "));
            }
        }
        if code != 0 {
            eprintln!("step {label}: exit {code}");
            self.failed = true;
            if !self.keep_going {
                self.stopped = true;
            }
        }
        Ok(())
    }

    /// Removes the candidate with `git worktree remove` (never a recursive delete) and its
    /// temporary parent with a plain `remove_dir`. A leftover is reported and fails the run.
    fn cleanup_candidate(&mut self) {
        let Some(parent) = self.tmp_parent.clone() else {
            return;
        };
        self.cwd = self.root.clone();
        if self.cand_added {
            if !self.cand.starts_with(&parent) || self.cand == parent {
                warn(&format!(
                    "refusing to remove '{}': it is not under '{}'",
                    self.cand.display(),
                    parent.display()
                ));
                self.failed = true;
                return;
            }
            let cand_arg = self.cand.to_string_lossy().into_owned();
            let removed = Command::new("git")
                .args(["worktree", "remove", "--force", &cand_arg])
                .current_dir(&self.root)
                .stdin(Stdio::null())
                .stdout(Stdio::null())
                .stderr(Stdio::null())
                .status();
            if removed.is_ok_and(|s| s.success()) {
                self.cand_added = false;
            } else {
                warn(&format!(
                    "could not remove the candidate worktree '{cand_arg}'; remove it with 'git worktree remove --force' and then 'rmdir' its folder"
                ));
                self.failed = true;
                return;
            }
        }
        if std::fs::remove_dir(&parent).is_ok() {
            self.tmp_parent = None;
        } else {
            warn(&format!(
                "could not remove the temporary folder '{}'",
                parent.display()
            ));
            self.failed = true;
        }
    }

    /// The last line, and the exit code: 0 only for `result=pass`.
    fn finish(&mut self) -> i32 {
        self.cleanup_candidate();
        if self.tree_is_dirty() {
            self.dirty = true;
        }
        // A commit made while the checks ran means the checked tree is not the reported head:
        // fail, and keep the head the run started from.
        let now = match git(&self.root, &["rev-parse", "HEAD"]) {
            Some(out) if out.status.success() => text_of(&out),
            _ => "unknown".to_string(),
        };
        if now != self.head {
            self.failed = true;
        }
        let mut pass = !(self.failed || self.dirty || self.dry_run);
        let at = Command::new("date")
            .args(["-u", "+%Y-%m-%dT%H:%M:%SZ"])
            .output()
            .map(|o| text_of(&o))
            .unwrap_or_default();
        if !self.candidate_mode {
            let flag = if self.dirty { "dirty=1 " } else { "" };
            say(&format!(
                "local-check: head={} result={} {flag}cmd={NAME} at={at}",
                self.head,
                if pass { "pass" } else { "fail" }
            ));
        } else {
            // A pass needs a real base, a real tree and a readable toolchain; never print one without.
            if pass
                && (self.base_sha == "none"
                    || self.tree_sha == "none"
                    || self.rustc_tok == "unknown")
            {
                pass = false;
            }
            let script = self.root.join(SCRIPT_SOURCE);
            let mut script_hash = git(
                &self.root,
                &["hash-object", "--", &script.to_string_lossy()],
            )
            .filter(|o| o.status.success())
            .map(|o| text_of(&o))
            .unwrap_or_else(|| "none".to_string());
            let seed = format!(
                "{} {} {} {} {}",
                self.head,
                self.base_sha,
                self.tree_sha,
                at,
                std::process::id()
            );
            let mut run_id = self.hash_stdin(&seed);
            if !is_hex40(&script_hash) || run_id.len() != 12 {
                warn("could not compute the script or run id");
                script_hash = "0".repeat(40);
                run_id = "0".repeat(12);
                pass = false;
            }
            say(&format!(
                "local-check: head={} result={} base={} tree={} script={script_hash} rustc={} run={run_id} cmd={NAME} at={at}",
                self.head,
                if pass { "pass" } else { "fail" },
                self.base_sha,
                self.tree_sha,
                self.rustc_tok
            ));
        }
        if pass {
            0
        } else {
            1
        }
    }

    /// The first 12 hex characters of `git hash-object --stdin` of `text` (the script's run id),
    /// or nothing when git cannot say. The seed has no trailing newline, like `printf '%s'`.
    fn hash_stdin(&self, text: &str) -> String {
        let child = Command::new("git")
            .args(["hash-object", "--stdin"])
            .current_dir(&self.root)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::null())
            .spawn();
        let Ok(mut child) = child else {
            return String::new();
        };
        if let Some(mut pipe) = child.stdin.take() {
            let _ = pipe.write_all(text.as_bytes());
        }
        match child.wait_with_output() {
            Ok(out) => text_of(&out).chars().take(12).collect(),
            Err(_) => String::new(),
        }
    }
}
