//! SPEC-20261125-check-affected-exit-code: `check affected` exits non-zero unless the run ends in
//! `result=PASS` (AC-1), stops after a failed step and lists the steps it did not start in
//! `skipped` unless `--keep-going` is given (AC-2), and takes the fork point of a base that is not
//! an ancestor of the head as its base (AC-3). Temporary git repositories, a temporary HOME,
//! temporary slot folders and fake `cargo` and `rustfmt` scripts only: the real tools, the real
//! HOME and the machine's heavy slots are never used.
#![cfg(unix)]
mod common;

use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;

/// Roomy machine numbers (debug-only seam), so the test machine's real disk and memory never
/// decide the outcome.
const ROOMY: &str = "free_gb=1000,swap_pct=0";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

fn init(p: &Path) {
    git_ok(p, &["init", "-q", "-b", "main"]);
    git_ok(p, &["config", "user.email", "t@example.invalid"]);
    git_ok(p, &["config", "user.name", "t"]);
}

/// Commits every change under `p` and returns the new commit.
fn commit_all(p: &Path, message: &str) -> String {
    git_ok(p, &["add", "-A"]);
    git_ok(p, &["commit", "-q", "-m", message]);
    git(p, &["rev-parse", "HEAD"]).trim().to_string()
}

/// The files of a small Rust project: changing `src/update_models.rs` selects the target
/// `update_command_cli`, so the run plans `diff_check`, `fmt`, `clippy` and `test`.
fn rust_files(p: &Path, text: &str) {
    put(p, "README.md", "x\n");
    put(p, "src/update_models.rs", text);
    put(p, "tests/update_command_cli.rs", "fn main() {}\n");
}

/// A repository, the base the run is given, the head the checkout is at and, when the base is no
/// ancestor of the head, the fork point `git merge-base` names.
struct Repo {
    dir: tempfile::TempDir,
    base: String,
    head: String,
    fork: Option<String>,
}

impl Repo {
    /// The base is the commit the head grew from.
    fn rust() -> Repo {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        init(p);
        rust_files(p, "// a\n");
        let base = commit_all(p, "base");
        put(p, "src/update_models.rs", "// b\n");
        let head = commit_all(p, "change");
        Repo {
            dir,
            base,
            head,
            fork: None,
        }
    }

    /// A repository without Rust code whose profile declares `checks` (id, command), or has no
    /// profile at all.
    fn web(checks: Option<&[(&str, &str)]>) -> Repo {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        init(p);
        put(p, "package.json", "{}\n");
        put(p, "src/index.ts", "export const a = 1;\n");
        if let Some(checks) = checks {
            let rows: Vec<Value> = checks
                .iter()
                .map(|(id, command)| serde_json::json!({"id": id, "command": command}))
                .collect();
            let profile = serde_json::json!({
                "schema_version": 2, "project_id": "demo", "purpose": "A small web app",
                "tools": ["TypeScript"], "development_methods": ["TDD"], "hosts": ["claude"],
                "checks": rows,
            });
            put(p, ".moai/project.yaml", &profile.to_string());
        }
        let base = commit_all(p, "base");
        put(p, "src/index.ts", "export const a = 2;\n");
        let head = commit_all(p, "change");
        Repo {
            dir,
            base,
            head,
            fork: None,
        }
    }

    /// The integration branch moved on after the lane branch was cut: the base the run is given
    /// is the tip of `main`, the head is the lane's, and the fork point is the commit both grew
    /// from.
    fn ahead() -> Repo {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        init(p);
        rust_files(p, "// a\n");
        let fork = commit_all(p, "fork point");
        git_ok(p, &["checkout", "-q", "-b", "lane"]);
        put(p, "src/update_models.rs", "// lane\n");
        let head = commit_all(p, "lane change");
        git_ok(p, &["checkout", "-q", "main"]);
        put(p, "README.md", "integration moved on\n");
        let base = commit_all(p, "integration change");
        git_ok(p, &["checkout", "-q", "lane"]);
        assert_eq!(git(p, &["rev-parse", "HEAD"]).trim(), head);
        Repo {
            dir,
            base,
            head,
            fork: Some(fork),
        }
    }

    /// The head's history shares no commit with the base's.
    fn unrelated() -> Repo {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        init(p);
        rust_files(p, "// a\n");
        let base = commit_all(p, "one history");
        git_ok(p, &["checkout", "-q", "--orphan", "other"]);
        put(p, "src/update_models.rs", "// other\n");
        let head = commit_all(p, "another history");
        // `git merge-base` exits 1 and prints nothing when there is no common commit.
        let found = common::git::git_raw(p, &["merge-base", &base, &head]);
        assert!(
            !found.success && found.stdout.trim().is_empty(),
            "the fixture must have no merge base: {}",
            found.stdout
        );
        Repo {
            dir,
            base,
            head,
            fork: None,
        }
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

/// Fake tools. A fake `cargo` appends `cargo <args>` to `calls.log`, exits 101 for `cargo
/// nextest …` (as cargo does for a subcommand that is not installed, so the tests are one `cargo
/// test` call) and exits 3 when its first argument is the word in `FAKE_FAIL`. The file is
/// written with the shell's own redirection, never a command from PATH.
struct Tools {
    dir: tempfile::TempDir,
}

impl Tools {
    fn new() -> Tools {
        let dir = tempfile::tempdir().unwrap();
        common::write_executable(
            &dir.path().join("cargo"),
            "#!/bin/sh\necho \"cargo $*\" >> \"$FAKE_CALLS\"\nif [ \"$1\" = nextest ]; then exit 101; fi\nif [ -n \"$FAKE_FAIL\" ] && [ \"$1\" = \"$FAKE_FAIL\" ]; then exit 3; fi\nexit 0\n",
        );
        common::write_executable(
            &dir.path().join("rustfmt"),
            "#!/bin/sh\necho \"rustfmt $*\" >> \"$FAKE_CALLS\"\nexit 0\n",
        );
        Tools { dir }
    }

    fn calls(&self) -> PathBuf {
        self.dir.path().join("calls.log")
    }

    fn lines(&self) -> Vec<String> {
        fs::read_to_string(self.calls())
            .unwrap_or_default()
            .lines()
            .map(String::from)
            .collect()
    }

    /// The calls that start the tests: one `cargo test …` or one `cargo nextest run …`. The probe
    /// `cargo nextest --version` is no test step.
    fn test_calls(&self) -> Vec<String> {
        self.lines()
            .into_iter()
            .filter(|l| l.starts_with("cargo test ") || l.starts_with("cargo nextest run"))
            .collect()
    }

    /// PATH with the fakes first and only `git` and the system's shell tools after them.
    fn path(&self) -> String {
        let git_dir = Command::new("sh")
            .args(["-c", "dirname \"$(command -v git)\""])
            .output()
            .unwrap();
        format!(
            "{}:{}:/usr/bin:/bin",
            self.dir.path().display(),
            String::from_utf8_lossy(&git_dir.stdout).trim()
        )
    }
}

/// A slot folder whose limits file allows `max` heavy runs; 0 makes every hold refuse (125).
fn slot_dir_with(max: u32) -> tempfile::TempDir {
    let d = tempfile::tempdir().unwrap();
    fs::write(
        d.path().join("limits.json"),
        format!(r#"{{"max":{max},"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}}"#),
    )
    .unwrap();
    d
}

/// What one `moai check affected` call did: its exit code and the envelope it printed.
struct Outcome {
    code: Option<i32>,
    stdout: String,
    stderr: String,
    envelope: Value,
}

impl Outcome {
    /// The envelope's `result`: the answer of the command, whatever the exit code.
    fn answer(&self) -> &Value {
        if self.envelope.is_null() || self.envelope.get("result").is_none() {
            panic!(
                "no answer on stdout (exit {:?}): {} / {}",
                self.code, self.stdout, self.stderr
            );
        }
        &self.envelope["result"]
    }

    /// The summary line of a run.
    fn line(&self) -> &str {
        self.answer()["summary"]
            .as_str()
            .unwrap_or_else(|| panic!("no summary in {}", self.answer()))
    }

    /// The exit code, which must be a real one (a signal gives none).
    fn exit(&self) -> i32 {
        self.code
            .unwrap_or_else(|| panic!("the process was killed: {}", self.stderr))
    }

    /// The labels of the steps that ran, in order.
    fn labels(&self) -> Vec<String> {
        strings(&self.answer()["steps"])
    }

    /// The labels `skipped` lists; an absent, null or empty field is none.
    fn skipped(&self) -> Vec<String> {
        strings(&self.answer()["skipped"])
    }
}

/// The strings of a JSON array; a step object counts by its `label`. Anything else is none.
fn strings(v: &Value) -> Vec<String> {
    v.as_array()
        .map(|a| {
            a.iter()
                .filter_map(|x| match x {
                    Value::String(s) => Some(s.clone()),
                    Value::Object(_) => x["label"].as_str().map(String::from),
                    _ => None,
                })
                .collect()
        })
        .unwrap_or_default()
}

/// Everything a run needs besides the repository: the fakes, a slot folder and a temporary HOME.
struct Run {
    tools: Tools,
    slots: tempfile::TempDir,
    home: tempfile::TempDir,
    /// The first argument of the fake `cargo` that makes it exit 3 (empty: none).
    fail: &'static str,
}

impl Run {
    fn new() -> Run {
        Run::with(4, "")
    }

    /// A run whose fake `cargo` fails for `step` (for example `clippy`).
    fn failing(step: &'static str) -> Run {
        Run::with(4, step)
    }

    fn with(slots: u32, fail: &'static str) -> Run {
        Run {
            tools: Tools::new(),
            slots: slot_dir_with(slots),
            home: tempfile::tempdir().unwrap(),
            fail,
        }
    }

    /// `moai check affected <range> --dir … --slot-dir …` followed by `extra`.
    fn exec_range(&self, repo: &Repo, range: &str, extra: &[&str]) -> Outcome {
        let out = Command::new(BIN)
            .args(["check", "affected", range, "--dir"])
            .arg(repo.root())
            .arg("--slot-dir")
            .arg(self.slots.path())
            .args(extra)
            .env("HOME", self.home.path())
            .env_remove("MOAI_HOME")
            .env_remove("XDG_CACHE_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("MOAI_HEAVY_SLOT_FD")
            .env_remove("CARGO_BUILD_JOBS")
            .env_remove("RUST_TEST_THREADS")
            .env_remove("NEXTEST_TEST_THREADS")
            .env("MOAI_TEST_MACHINE", ROOMY)
            .env("PATH", self.tools.path())
            .env("FAKE_CALLS", self.tools.calls())
            .env("FAKE_FAIL", self.fail)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let stdout = String::from_utf8_lossy(&out.stdout).to_string();
        let envelope = stdout
            .find('{')
            .and_then(|at| serde_json::from_str(&stdout[at..]).ok())
            .unwrap_or(Value::Null);
        Outcome {
            code: out.status.code(),
            stdout,
            stderr: String::from_utf8_lossy(&out.stderr).to_string(),
            envelope,
        }
    }

    fn exec(&self, repo: &Repo, extra: &[&str]) -> Outcome {
        self.exec_range(repo, &repo.range(), extra)
    }
}

/// The keys of a summary line, in order.
fn keys(line: &str) -> Vec<&str> {
    line.split(' ')
        .skip(1)
        .filter_map(|w| w.split_once('=').map(|(k, _)| k))
        .collect()
}

/// The value of the summary word `key=`.
fn word<'a>(line: &'a str, key: &str) -> Option<&'a str> {
    let prefix = format!("{key}=");
    line.split(' ')
        .find_map(|w| w.strip_prefix(prefix.as_str()))
}

fn short(sha: &str) -> &str {
    &sha[..12]
}

// ---------------------------------------------------------------------------------------------
// AC-1 (REQ-1, REQ-2): the exit code follows the result word.

/// A run whose step fails prints the answer with `result` FAIL and exits non-zero.
#[test]
fn ac1_a_failed_step_exits_non_zero_with_result_fail() {
    let repo = Repo::rust();
    let run = Run::failing("clippy");
    let o = run.exec(&repo, &[]);
    assert_eq!(o.answer()["result"], "FAIL", "{}", o.stdout);
    assert!(o.line().ends_with(" result=FAIL"), "{}", o.line());
    assert_ne!(
        o.exit(),
        0,
        "a FAIL run exited 0; stdout: {} stderr: {}",
        o.stdout,
        o.stderr
    );
}

/// The same for a declared check of a repository without Rust code, with `--keep-going` not
/// given: the failed check still gives FAIL and a non-zero exit.
#[test]
fn ac1_a_failed_declared_check_exits_non_zero_with_result_fail() {
    let repo = Repo::web(Some(&[("lint", "exit 3")]));
    let o = Run::new().exec(&repo, &[]);
    assert_eq!(o.answer()["result"], "FAIL", "{}", o.stdout);
    assert!(o.line().contains(" check=3 "), "{}", o.line());
    assert_ne!(o.exit(), 0, "a FAIL run exited 0: {}", o.stdout);
}

/// A repository without Rust code that declares no check (no profile, or a profile without
/// `checks`) verified nothing: `result` is `not_run` and the exit is non-zero.
#[test]
fn ac1_nothing_declared_exits_non_zero_with_result_not_run() {
    let no_checks = serde_json::json!({
        "schema_version": 2, "project_id": "demo", "purpose": "A small web app",
        "tools": ["TypeScript"], "development_methods": ["TDD"], "hosts": ["claude"],
    })
    .to_string();
    let cases: [(&str, Repo); 2] = [
        ("no profile", Repo::web(None)),
        ("no checks", {
            let repo = Repo::web(None);
            let p = repo.root();
            put(p, ".moai/project.yaml", &no_checks);
            let head = commit_all(p, "profile without checks");
            Repo { head, ..repo }
        }),
    ];
    for (what, repo) in cases {
        let o = Run::new().exec(&repo, &[]);
        assert_eq!(o.answer()["result"], "not_run", "{what}: {}", o.stdout);
        assert!(o.line().ends_with(" result=not_run"), "{what}");
        assert_ne!(o.exit(), 0, "{what}: a not_run run exited 0: {}", o.stdout);
    }
}

/// A passing run exits 0, for a Rust repository and for one with a declared check.
#[test]
fn ac1_a_passing_run_exits_zero() {
    let rust = Repo::rust();
    let o = Run::new().exec(&rust, &[]);
    assert_eq!(o.answer()["result"], "PASS", "{}", o.stdout);
    assert_eq!(o.exit(), 0, "{} / {}", o.stdout, o.stderr);

    let web = Repo::web(Some(&[("lint", "echo lint")]));
    let o = Run::new().exec(&web, &[]);
    assert_eq!(o.answer()["result"], "PASS", "{}", o.stdout);
    assert_eq!(o.exit(), 0, "{} / {}", o.stdout, o.stderr);
}

/// `--list` runs nothing and has no result: it exits 0, also where a run would be `not_run`.
#[test]
fn ac1_list_exits_zero() {
    for (what, repo) in [
        ("rust", Repo::rust()),
        ("nothing declared", Repo::web(None)),
    ] {
        let o = Run::new().exec(&repo, &["--list"]);
        assert_eq!(o.exit(), 0, "{what}: {} / {}", o.stdout, o.stderr);
        assert!(o.answer()["steps"].is_array(), "{what}: {}", o.stdout);
        assert!(o.answer().get("result").is_none(), "{what}: {}", o.stdout);
    }
}

// ---------------------------------------------------------------------------------------------
// AC-2 (REQ-3, REQ-4, REQ-7): a failed step stops the run unless `--keep-going` is given.

/// A failing clippy step starts no test step: the fake `cargo` never sees `test`, the answer lists
/// `test` in `skipped`, names `clippy` in `stopped_at`, and the summary has no word for the step
/// that did not run and ends in the stop.
#[test]
fn ac2_a_failed_clippy_step_skips_the_test_step() {
    let repo = Repo::rust();
    let run = Run::failing("clippy");
    let o = run.exec(&repo, &[]);
    let line = o.line();
    assert_eq!(o.answer()["result"], "FAIL", "{line}");
    assert!(
        run.tools.test_calls().is_empty(),
        "the test step ran after a failed clippy: {:?}",
        run.tools.lines()
    );
    assert!(
        run.tools
            .lines()
            .iter()
            .any(|l| l.starts_with("cargo clippy ")),
        "clippy never ran: {:?}",
        run.tools.lines()
    );
    assert_eq!(o.skipped(), ["test"], "{}", o.answer());
    assert_eq!(o.answer()["stopped_at"], "clippy", "{}", o.answer());
    assert_eq!(
        o.labels(),
        ["diff_check", "fmt", "clippy"],
        "{}",
        o.answer()
    );
    assert!(line.ends_with(" stopped_at=clippy result=FAIL"), "{line}");
    assert_eq!(
        keys(line),
        [
            "base",
            "head",
            "targets",
            "scans",
            "test-threads",
            "diff_check",
            "fmt",
            "clippy",
            "stopped_at",
            "result"
        ],
        "{line}"
    );
}

/// `--keep-going` runs every planned step after the failure: the test step runs, nothing is
/// skipped, there is no `stopped_at`, and the result is still FAIL.
#[test]
fn ac2_keep_going_runs_the_test_step_after_a_failed_clippy_step() {
    let repo = Repo::rust();
    let run = Run::failing("clippy");
    let o = run.exec(&repo, &["--keep-going"]);
    let line = o.line();
    assert_eq!(
        run.tools.test_calls().len(),
        1,
        "--keep-going must run the test step once: {:?}",
        run.tools.lines()
    );
    assert_eq!(o.answer()["result"], "FAIL", "{line}");
    assert!(
        o.answer()["stopped_at"].is_null(),
        "no stop with --keep-going: {}",
        o.answer()
    );
    assert!(o.skipped().is_empty(), "{}", o.answer());
    assert_eq!(
        o.labels(),
        ["diff_check", "fmt", "clippy", "test"],
        "{}",
        o.answer()
    );
    assert!(!line.contains("stopped_at"), "{line}");
    assert!(line.ends_with(" clippy=3 test=0 result=FAIL"), "{line}");
}

/// The direction the stop lets through: a run in which no step fails runs every step, skips none
/// and has no `stopped_at`.
#[test]
fn ac2_a_run_without_a_failure_skips_nothing() {
    let repo = Repo::rust();
    let run = Run::new();
    let o = run.exec(&repo, &[]);
    assert_eq!(o.answer()["result"], "PASS", "{}", o.stdout);
    assert_eq!(run.tools.test_calls().len(), 1, "{:?}", run.tools.lines());
    assert!(o.skipped().is_empty(), "{}", o.answer());
    assert!(o.answer()["stopped_at"].is_null(), "{}", o.answer());
    assert!(!o.line().contains("stopped_at"), "{}", o.line());
}

/// REQ-7: a heavy slot that is not obtained still gives its exit code to every heavy step, and no
/// cargo step but the cheap `fmt` starts. The stop of REQ-3 does not turn the second heavy step
/// into a skipped one.
#[test]
fn ac2_a_refused_slot_still_gives_its_code_to_every_heavy_step() {
    let repo = Repo::rust();
    let run = Run::with(0, "");
    let o = run.exec(&repo, &[]);
    let line = o.line();
    assert!(line.ends_with(" clippy=125 test=125 result=FAIL"), "{line}");
    assert!(o.skipped().is_empty(), "{}", o.answer());
    let exits: Vec<(String, i64)> = o.answer()["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| {
            (
                s["label"].as_str().unwrap().to_string(),
                s["exit"].as_i64().unwrap(),
            )
        })
        .collect();
    for label in ["clippy", "test"] {
        assert!(
            exits.contains(&(label.to_string(), 125)),
            "{label} must carry the slot's code: {exits:?}"
        );
    }
    let cargo: Vec<String> = run
        .tools
        .lines()
        .into_iter()
        .filter(|l| l.starts_with("cargo "))
        .collect();
    assert!(
        !cargo.is_empty() && cargo.iter().all(|l| l.starts_with("cargo fmt ")),
        "a heavy cargo step started: {cargo:?}"
    );
}

// ---------------------------------------------------------------------------------------------
// AC-3 (REQ-5, REQ-6, REQ-8): the base is a commit the head grew from.

/// A base that is ahead of the lane (not an ancestor of the head): the summary names the merge
/// base as `base=` and the given base as `base_from=` before the result, the answer carries
/// `base_given`, and `card note` accepts the receipt. Both the bare run and `--card` do it.
#[test]
fn ac3_a_base_ahead_of_the_lane_is_replaced_by_the_fork_point() {
    let repo = Repo::ahead();
    let fork = repo.fork.clone().unwrap();
    assert_eq!(
        git(repo.root(), &["merge-base", &repo.base, &repo.head]).trim(),
        fork,
        "the fixture's fork point"
    );
    for extra in [&["--card"][..], &[][..]] {
        let o = Run::new().exec(&repo, extra);
        let line = o.line();
        assert_eq!(o.answer()["result"], "PASS", "{extra:?}: {line}");
        assert_eq!(word(line, "base"), Some(short(&fork)), "{extra:?}: {line}");
        let from = format!(" base_from={}", short(&repo.base));
        let at = line
            .find(&from)
            .unwrap_or_else(|| panic!("{extra:?}: no{from} in {line}"));
        assert!(
            at < line.rfind(" result=").unwrap(),
            "{extra:?}: base_from after the result: {line}"
        );
        assert!(line.ends_with(" result=PASS"), "{extra:?}: {line}");
        assert_eq!(
            word(line, "base_from"),
            Some(short(&repo.base)),
            "{extra:?}: {line}"
        );
        assert_eq!(
            o.answer()["base_given"],
            repo.base.as_str(),
            "{extra:?}: {}",
            o.answer()
        );
        // The receipt on disk is the line, and `card note` keeps a note of the head on it.
        let out = Path::new(o.answer()["out"].as_str().unwrap());
        assert_eq!(
            fs::read_to_string(out.join("summary.txt")).unwrap(),
            format!("{line}\n"),
            "{extra:?}"
        );
        let note = format!("LANE-READY-INTEGRATION C-1 head={}", repo.head);
        let accepted = moai_cli::card::ready_receipt::check_note(repo.root(), "C-1", &note);
        assert!(accepted.is_ok(), "{extra:?}: {accepted:?} for {line}");
        // The base the receipt names is a commit the head grew from.
        let base = word(line, "base").unwrap();
        git_ok(
            repo.root(),
            &["merge-base", "--is-ancestor", base, &repo.head],
        );
    }
}

/// A base that is an ancestor of the head: the line is written exactly as before, with no
/// `base_from`, and the answer has no `base_given`. The bare line keeps its words and their order.
#[test]
fn ac3_an_ancestor_base_writes_the_line_as_before() {
    let repo = Repo::rust();
    for extra in [&[][..], &["--card"][..]] {
        let o = Run::new().exec(&repo, extra);
        let line = o.line();
        assert_eq!(o.answer()["result"], "PASS", "{extra:?}: {line}");
        assert_eq!(word(line, "base"), Some(short(&repo.base)), "{line}");
        assert!(!line.contains("base_from"), "{extra:?}: {line}");
        assert!(
            o.answer().get("base_given").is_none_or(Value::is_null),
            "{extra:?}: {}",
            o.answer()
        );
    }
    let o = Run::new().exec(&repo, &[]);
    assert_eq!(
        keys(o.line()),
        [
            "base",
            "head",
            "targets",
            "scans",
            "test-threads",
            "diff_check",
            "fmt",
            "clippy",
            "test",
            "result"
        ],
        "{}",
        o.line()
    );
}

/// Unrelated histories have no merge base: the run is refused as before (a non-zero exit and an
/// error that names the missing merge base) and writes no `base_from`.
#[test]
fn ac3_unrelated_histories_are_refused_as_before() {
    let repo = Repo::unrelated();
    let o = Run::new().exec(&repo, &[]);
    assert_ne!(o.exit(), 0, "{} / {}", o.stdout, o.stderr);
    assert!(o.stdout.contains("no merge base"), "{}", o.stdout);
    assert!(!o.stdout.contains("base_from"), "{}", o.stdout);
    assert!(!o.stdout.contains("base_given"), "{}", o.stdout);
}
