//! SPEC-20261122-check-affected-one-slot (AC-1 to AC-4, AC-6): `check affected` holds one heavy
//! slot for all of its cargo steps, runs the tests in one call, gives a refused slot's code to
//! every cargo step and sets a default `CARGO_BUILD_JOBS`. Temporary git repositories, a
//! temporary HOME, temporary slot folders and fake `cargo`, `python3` and `rustfmt` scripts
//! only: the real tools and the machine's heavy slots are never used.
#![cfg(unix)]
mod common;

use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::os::unix::process::CommandExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

const BIN: &str = crate::common::MOAI_BIN;

/// Roomy machine numbers (debug-only seam), so the test machine's real disk and memory never
/// decide the outcome.
const ROOMY: &str = "free_gb=1000,swap_pct=0";

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository with one commit on top of the base that changes `changed`.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
    head: String,
}

impl Fixture {
    fn new(changed: &[(&str, &str)]) -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "src/update_models.rs", "// a\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/bundle_ops_test.py", "#\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        for (rel, text) in changed {
            put(p, rel, text);
        }
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "change"]);
        let head = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        Fixture { dir, base, head }
    }

    /// The Rust target `update_command_cli` is selected, no Python module.
    fn rust_target() -> Fixture {
        Fixture::new(&[("src/update_models.rs", "// b\n")])
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head)
    }
}

/// Fake tools. Each call appends one line `<tool> <args> | FD=<MOAI_HEAVY_SLOT_FD>
/// JOBS=<CARGO_BUILD_JOBS>` to `calls.log`. A fake `cargo` made `without_nextest` exits 101 for
/// `cargo nextest …`, as cargo does for a subcommand that is not installed. With `FAKE_MOAI` set,
/// each call also writes `heavy-slot-status --dir $FAKE_SLOTS` as it is during that call to
/// `calls.log.<tool>-<first arg>.status`.
struct Tools {
    dir: tempfile::TempDir,
}

impl Tools {
    fn new(with_nextest: bool) -> Tools {
        let dir = tempfile::tempdir().unwrap();
        for n in ["cargo", "python3", "rustfmt"] {
            let nextest = if n == "cargo" && !with_nextest {
                "if [ \"$1\" = nextest ]; then exit 101; fi\n"
            } else {
                ""
            };
            let script = format!(
                "#!/bin/sh\necho \"{n} $* | FD=$MOAI_HEAVY_SLOT_FD JOBS=$CARGO_BUILD_JOBS\" >> \"$FAKE_CALLS\"\nif [ -n \"$FAKE_MOAI\" ]; then \"$FAKE_MOAI\" heavy-slot-status --dir \"$FAKE_SLOTS\" > \"$FAKE_CALLS.{n}-$1.status\"; fi\n{nextest}exit 0\n"
            );
            let path = dir.path().join(n);
            fs::write(&path, script).unwrap();
            use std::os::unix::fs::PermissionsExt;
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
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

    /// The calls that start with `prefix`.
    fn starting(&self, prefix: &str) -> Vec<String> {
        self.lines()
            .into_iter()
            .filter(|l| l.starts_with(prefix))
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

/// Everything a run needs besides the fixture: the fakes, a slot folder, the log folder and a
/// temporary HOME.
struct Run {
    tools: Tools,
    slots: tempfile::TempDir,
    out: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Run {
    fn new(with_nextest: bool, max: u32) -> Run {
        Run {
            tools: Tools::new(with_nextest),
            slots: slot_dir_with(max),
            out: tempfile::tempdir().unwrap(),
            home: tempfile::tempdir().unwrap(),
        }
    }

    /// `moai check affected <range> --dir … --slot-dir … --out …` with the fakes on PATH.
    fn command(&self, fx: &Fixture, extra: &[&str]) -> Command {
        let mut cmd = Command::new(BIN);
        cmd.args(["check", "affected", &fx.range(), "--dir"])
            .arg(fx.root())
            .arg("--slot-dir")
            .arg(self.slots.path())
            .arg("--out")
            .arg(self.out.path())
            .args(extra)
            .env("HOME", self.home.path())
            .env_remove("MOAI_HOME")
            .env_remove("XDG_CACHE_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("MOAI_HEAVY_SLOT_FD")
            .env("MOAI_CLI_TEST_MACHINE", ROOMY)
            .env("PATH", self.tools.path())
            .env("FAKE_CALLS", self.tools.calls())
            .stdin(Stdio::null());
        cmd
    }

    fn slot_log(&self) -> String {
        fs::read_to_string(self.out.path().join("heavy-slot.log")).unwrap_or_default()
    }
}

/// The `result` of a successful answer on stdout (anything before the JSON is skipped).
fn answer(stdout: &[u8], stderr: &[u8]) -> Value {
    let text = String::from_utf8_lossy(stdout);
    let at = text.find('{').unwrap_or(0);
    let v: Value = serde_json::from_str(&text[at..])
        .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(stderr)));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn run(mut cmd: Command) -> Value {
    let out = cmd.output().unwrap();
    answer(&out.stdout, &out.stderr)
}

/// The keys of a summary line, in order.
fn keys(line: &str) -> Vec<&str> {
    line.split(' ')
        .skip(1)
        .filter_map(|w| w.split_once('=').map(|(k, _)| k))
        .collect()
}

/// The `FD=` and `JOBS=` values of a call line.
fn seen(line: &str, key: &str) -> String {
    let (_, env) = line.rsplit_once(" | ").unwrap();
    env.split(' ')
        .find_map(|w| w.strip_prefix(&format!("{key}=")))
        .unwrap_or("")
        .to_string()
}

/// AC-1 (REQ-1, REQ-3): one run takes one slot, every heavy cargo call runs under that slot,
/// and the tests are one `cargo nextest run` call with no `cargo test`.
#[test]
fn one_run_takes_one_slot_and_one_nextest_call() {
    let fx = Fixture::rust_target();
    let r = Run::new(true, 4);
    let v = run(r.command(&fx, &[]));
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");
    assert_eq!(
        keys(line),
        [
            "base",
            "head",
            "targets",
            "scans",
            "diff_check",
            "fmt",
            "clippy",
            "test",
            "result"
        ],
        "{line}"
    );
    // One slot taken, by this run, once.
    let log = r.slot_log();
    assert_eq!(
        log.matches("heavy-slot: slot ").count(),
        1,
        "one slot record: {log}"
    );
    assert!(log.contains("slot 0 taken by check-affected"), "{log}");
    // One test call, by nextest, with the library, the `moai-cli` binary (not every binary's
    // harness: no `--bins`, D-5) and the selected target.
    let nextest = r.tools.starting("cargo nextest run ");
    assert_eq!(nextest.len(), 1, "{:?}", r.tools.lines());
    assert!(
        nextest[0].starts_with(
            "cargo nextest run --locked --offline --no-fail-fast --lib --bin moai-cli --test update_command_cli |"
        ),
        "{nextest:?}"
    );
    assert!(!nextest[0].contains("--bins"), "{nextest:?}");
    assert!(
        r.tools.starting("cargo test").is_empty(),
        "{:?}",
        r.tools.lines()
    );
    // Clippy and the tests ran under the same inherited slot file.
    let clippy = r.tools.starting("cargo clippy ");
    assert_eq!(clippy.len(), 1, "{:?}", r.tools.lines());
    let fd = seen(&clippy[0], "FD");
    assert!(fd.ends_with("/slot-0") && fd.contains(':'), "{fd}");
    assert_eq!(seen(&nextest[0], "FD"), fd);
    // The cheap `fmt` step runs before the slot is taken.
    assert_eq!(seen(&r.tools.starting("cargo fmt ")[0], "FD"), "");
    assert_eq!(v["runner"], "nextest", "{v}");
    let test = v["steps"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["label"] == "test")
        .unwrap();
    assert_eq!(test["command"][1], "nextest", "{test}");
    assert_eq!(test["exit"], 0, "{test}");
}

/// AC-2 (REQ-3): when `cargo nextest --version` fails, the tests are one `cargo test` call.
#[test]
fn without_nextest_one_cargo_test_call() {
    let fx = Fixture::rust_target();
    let r = Run::new(false, 4);
    let v = run(r.command(&fx, &[]));
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");
    let tests = r.tools.starting("cargo test ");
    assert_eq!(tests.len(), 1, "{:?}", r.tools.lines());
    assert!(
        tests[0].starts_with(
            "cargo test --locked --offline --no-fail-fast --lib --bin moai-cli --test update_command_cli |"
        ),
        "{tests:?}"
    );
    assert!(!tests[0].contains("--bins"), "{tests:?}");
    assert!(
        r.tools.starting("cargo nextest run").is_empty(),
        "{:?}",
        r.tools.lines()
    );
    assert_eq!(v["runner"], "cargo-test", "{v}");
}

/// AC-3 (REQ-2): a refused slot (an unusable limits file, 125) and a machine too short of disk
/// (75) give their code to every cargo step — `clippy` and `test` — and no cargo starts but the
/// cheap `fmt`.
#[test]
fn a_refused_slot_marks_every_cargo_step_and_runs_none() {
    // A Rust target selected: clippy and test.
    let fx = Fixture::rust_target();
    let r = Run::new(true, 0);
    let v = run(r.command(&fx, &[]));
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=125 test=125 result=FAIL"), "{line}");
    assert!(r.slot_log().contains("refused"), "{}", r.slot_log());
    // Only a Python module changed: no target is selected, the two cargo steps still get the
    // code, and there is no `build_cli` or `python` step.
    let only_python = Fixture::new(&[("tests/bundle_ops_test.py", "# b\n")]);
    let r2 = Run::new(true, 0);
    let v = run(r2.command(&only_python, &[]));
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=125 test=125 result=FAIL"), "{line}");
    assert!(r2.tools.starting("python3").is_empty());
    // A machine below the free-disk floor (the caller's 50 GB is stricter than the file's 0).
    let r3 = Run::new(true, 4);
    let mut busy = r3.command(&fx, &[]);
    busy.env("MOAI_CLI_TEST_MACHINE", "free_gb=1,swap_pct=0");
    let v = run(busy);
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=75 test=75 result=FAIL"), "{line}");
    for t in [&r.tools, &r2.tools, &r3.tools] {
        let cargo = t.starting("cargo ");
        assert!(
            cargo.iter().all(|l| l.starts_with("cargo fmt ")),
            "a cargo step ran: {cargo:?}"
        );
        assert!(!cargo.is_empty(), "the cheap fmt step ran");
    }
}

/// AC-4 (REQ-6): without a caller value every heavy cargo step gets `CARGO_BUILD_JOBS` = the
/// larger of 2 and the logical CPUs divided by the effective slot count; a caller's value passes
/// through unchanged.
#[test]
fn cargo_build_jobs_defaults_from_the_slot_count_and_a_caller_value_wins() {
    let fx = Fixture::rust_target();
    let cpus = std::thread::available_parallelism().unwrap().get();
    // The limits file allows 4; `--max 2` is stricter and wins, so the effective count is 2.
    for (flags, slots) in [(&[][..], 4usize), (&["--max", "2"][..], 2)] {
        let r = Run::new(true, 4);
        let mut cmd = r.command(&fx, flags);
        cmd.env_remove("CARGO_BUILD_JOBS");
        let v = run(cmd);
        assert!(
            v["summary"].as_str().unwrap().ends_with(" result=PASS"),
            "{v}"
        );
        let want = (cpus / slots).max(2).to_string();
        let heavy: Vec<String> = r
            .tools
            .lines()
            .into_iter()
            .filter(|l| l.starts_with("cargo clippy ") || l.starts_with("cargo nextest run "))
            .collect();
        assert_eq!(heavy.len(), 2, "{heavy:?}");
        for l in &heavy {
            assert_eq!(seen(l, "JOBS"), want, "{slots} slots, {cpus} cpus: {l}");
        }
    }
    let r = Run::new(true, 4);
    let mut cmd = r.command(&fx, &[]);
    cmd.env("CARGO_BUILD_JOBS", "7");
    run(cmd);
    let heavy = r.tools.starting("cargo clippy ");
    assert_eq!(seen(&heavy[0], "JOBS"), "7", "{heavy:?}");
    assert_eq!(
        seen(&r.tools.starting("cargo nextest run ")[0], "JOBS"),
        "7"
    );
}

/// AC-6 (REQ-5): `check affected` run inside a one-slot `heavy-slot-acquire` on the same folder
/// reuses the slot it inherited instead of waiting for itself. A hang guard ends the case.
#[test]
fn a_check_inside_a_one_slot_acquire_does_not_wait_on_itself() {
    let fx = Fixture::rust_target();
    let r = Run::new(true, 1);
    let inner = r.command(&fx, &[]);
    let mut outer = Command::new(BIN);
    outer
        .args([
            "heavy-slot-acquire",
            "--max",
            "1",
            "--name",
            "outer",
            "--dir",
        ])
        .arg(r.slots.path())
        .arg("--")
        .arg(inner.get_program())
        .args(inner.get_args());
    for (k, v) in inner.get_envs() {
        match v {
            Some(v) => outer.env(k, v),
            None => outer.env_remove(k),
        };
    }
    let mut child = outer
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .process_group(0)
        .spawn()
        .unwrap();
    let group = child.id() as i32;
    let start = Instant::now();
    loop {
        if child.try_wait().unwrap().is_some() {
            break;
        }
        if start.elapsed() > Duration::from_secs(60) {
            // SAFETY: kill(2) on the process group this test created.
            unsafe { libc::kill(-group, libc::SIGKILL) };
            let _ = child.wait();
            panic!(
                "hang guard: the check waited on its own slot; slot log: {}",
                r.slot_log()
            );
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success(), "{out:?}");
    let v = answer(&out.stdout, &out.stderr);
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");
    let log = r.slot_log();
    assert!(log.contains("reuses slot 0 it inherited"), "{log}");
    assert_eq!(r.tools.starting("cargo nextest run ").len(), 1);
}

/// The `held` list of the slot status a fake call wrote (`<tool>-<first arg>`).
fn held_during(r: &Run, call: &str) -> Vec<Value> {
    let path = format!("{}.{call}.status", r.tools.calls().display());
    let text = fs::read_to_string(&path).unwrap_or_else(|e| panic!("{path}: {e}"));
    let v: Value = serde_json::from_str(&text).unwrap_or_else(|_| panic!("{path}: {text}"));
    let status = if v.get("result").is_some() {
        &v["result"]
    } else {
        &v
    };
    status["held"]
        .as_array()
        .unwrap_or_else(|| panic!("{path}: {text}"))
        .clone()
}

/// D-2 and the slot record: while a cargo step runs, the one slot is held and its record names
/// this run and its own command line. A changed Python module adds no step: nothing calls
/// `python3` (SPEC-20261123-rust-scans-before-ready REQ-3).
#[test]
fn the_slot_record_names_the_run_and_a_changed_python_module_adds_no_step() {
    let fx = Fixture::new(&[
        ("src/update_models.rs", "// b\n"),
        ("tests/bundle_ops_test.py", "# b\n"),
    ]);
    let r = Run::new(true, 4);
    let mut cmd = r.command(&fx, &[]);
    cmd.env("FAKE_MOAI", BIN).env("FAKE_SLOTS", r.slots.path());
    let v = run(cmd);
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");
    assert!(!line.contains("python"), "{line}");
    let held = held_during(&r, "cargo-clippy");
    assert_eq!(held.len(), 1, "{held:?}");
    assert_eq!(held[0]["name"], "check-affected", "{held:?}");
    let command: Vec<&str> = held[0]["command"]
        .as_array()
        .unwrap_or_else(|| panic!("{held:?}"))
        .iter()
        .filter_map(Value::as_str)
        .collect();
    let range = fx.range();
    for word in ["check", "affected", range.as_str()] {
        assert!(command.contains(&word), "{word} in {command:?}");
    }
    assert_eq!(
        held_during(&r, "cargo-nextest").len(),
        1,
        "the test call holds it"
    );
    assert!(r.tools.starting("python3").is_empty());
}

/// A slot count of 0 and a slot folder that cannot be named (no HOME, no XDG_CACHE_HOME, no
/// `--slot-dir`) are code 2 for every cargo step, and no cargo starts but the cheap `fmt`.
#[test]
fn no_slot_count_or_no_slot_folder_is_code_2_for_every_cargo_step() {
    let fx = Fixture::rust_target();
    let r = Run::new(true, 4);
    let v = run(r.command(&fx, &["--max", "0"]));
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=2 test=2 result=FAIL"), "{line}");
    assert!(
        r.slot_log().contains("--max 0 leaves no slot"),
        "{}",
        r.slot_log()
    );

    let r2 = Run::new(true, 4);
    let mut cmd = Command::new(BIN);
    cmd.args(["check", "affected", &fx.range(), "--dir"])
        .arg(fx.root())
        .arg("--out")
        .arg(r2.out.path())
        .env_remove("HOME")
        .env_remove("MOAI_HOME")
        .env_remove("XDG_CACHE_HOME")
        .env_remove("CARGO_TARGET_DIR")
        .env_remove("MOAI_HEAVY_SLOT_FD")
        .env("MOAI_CLI_TEST_MACHINE", ROOMY)
        .env("PATH", r2.tools.path())
        .env("FAKE_CALLS", r2.tools.calls())
        .stdin(Stdio::null());
    let v = run(cmd);
    let line = v["summary"].as_str().unwrap();
    assert!(line.ends_with(" clippy=2 test=2 result=FAIL"), "{line}");
    assert!(
        r2.slot_log().contains("HOME is not set"),
        "{}",
        r2.slot_log()
    );
    for t in [&r.tools, &r2.tools] {
        let cargo = t.starting("cargo ");
        assert!(
            !cargo.is_empty() && cargo.iter().all(|l| l.starts_with("cargo fmt ")),
            "{cargo:?}"
        );
    }
}
