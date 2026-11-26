//! SPEC-20261125-check-affected-rerun (AC-1 to AC-7): a repeated `check affected` run on a head
//! whose earlier run failed re-runs only the tests that failed, a head that moved runs the diff
//! targets and the earlier failures, the test step retries and lists a test that passes on retry
//! as FLAKY, a ledger counts the FLAKY names, `--full` keeps the full set, and the combined
//! `summary.txt` stays one receipt line. Temporary git repositories, a temporary HOME and a fake
//! `cargo` that prints nextest-shaped lines only: the real tools and the machine's heavy slots
//! are never used.
#![cfg(unix)]
mod common;

use common::git::{git, git_ok};
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const ROOMY: &str = "free_gb=1000,swap_pct=0";

/// The fake tests, in the order the fake prints them; each belongs to the binary `moai-cli::<target>`.
const TESTS: [(&str, &str); 4] = [
    ("update_command_cli", "t_a"),
    ("update_command_cli", "t_b"),
    ("update_command_cli", "t_c"),
    ("bundle_run_cli", "t_d"),
];

/// A fake `cargo` that answers `nextest --version`, prints a nextest-shaped line per test of
/// `FAKE_TESTS` that the `-E` filter keeps, and records every call. A name in `FAKE_FAIL` fails on
/// every attempt, a name in `FAKE_FLAKY` fails unless `--retries` is at least 1 and then prints
/// `FLAKY`. It exits 100 when a test failed, as nextest does.
const FAKE_CARGO: &str = r#"#!/bin/sh
echo "cargo $*" >> "$FAKE_CALLS"
if [ "$1" != nextest ]; then
  if [ -n "$FAKE_STEP_FAIL" ] && [ "$1" = "$FAKE_STEP_FAIL" ]; then exit 1; fi
  exit 0
fi
if [ "$2" = "--version" ]; then exit 0; fi
filter=""; retries=0
while [ $# -gt 0 ]; do
  case "$1" in
    -E) filter="$2"; shift;;
    --retries) retries="$2"; shift;;
  esac
  shift
done
code=0; n=0
for pair in $FAKE_TESTS; do
  bin="${pair%%:*}"; t="${pair##*:}"
  if [ -n "$filter" ]; then
    case "$filter" in
      *"test(=$t)"*) ;;
      *"binary_id(=moai-cli::$bin)"*)
        # the binary is selected as a whole unless the filter names tests of it
        case "$filter" in *"binary_id(=moai-cli::$bin) & test("*) continue;; esac;;
      *) continue;;
    esac
  fi
  n=$((n+1))
  case " $FAKE_FAIL " in
    *" $t "*) echo "        FAIL [   0.010s] ($n/9) moai-cli::$bin $t"; code=100; continue;;
  esac
  case " $FAKE_FLAKY " in
    *" $t "*)
      if [ "$retries" -ge 1 ]; then
        echo "        FLAKY 2/3 [   0.010s] ($n/9) moai-cli::$bin $t"
      else
        echo "        FAIL [   0.010s] ($n/9) moai-cli::$bin $t"; code=100
      fi
      continue;;
  esac
  echo "        PASS [   0.010s] ($n/9) moai-cli::$bin $t"
done
exit $code
"#;

fn put(root: &Path, rel: &str, text: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}

/// A repository with a base commit, a first change (module `update_models`, test
/// `update_command_cli`) and, on demand, further commits.
struct Fixture {
    dir: tempfile::TempDir,
    base: String,
}

impl Fixture {
    fn new() -> Fixture {
        let dir = tempfile::tempdir().unwrap();
        let p = dir.path();
        git_ok(p, &["init", "-q", "-b", "main"]);
        git_ok(p, &["config", "user.email", "t@example.invalid"]);
        git_ok(p, &["config", "user.name", "t"]);
        put(p, "README.md", "x\n");
        put(p, "Cargo.toml", "[package]\nname = \"x\"\n");
        put(p, "src/update_models.rs", "// a\n");
        put(p, "src/bundle_pack.rs", "// a\n");
        put(p, "tests/update_command_cli.rs", "fn main() {}\n");
        put(p, "tests/bundle_run_cli.rs", "fn main() {}\n");
        git_ok(p, &["add", "-A"]);
        git_ok(p, &["commit", "-q", "-m", "base"]);
        let base = git(p, &["rev-parse", "HEAD"]).trim().to_string();
        let fx = Fixture { dir, base };
        fx.commit(&[("src/update_models.rs", "// b\n")]);
        fx
    }

    fn root(&self) -> &Path {
        self.dir.path()
    }

    fn commit(&self, files: &[(&str, &str)]) -> String {
        for (rel, text) in files {
            put(self.root(), rel, text);
        }
        git_ok(self.root(), &["add", "-A"]);
        git_ok(self.root(), &["commit", "-q", "-m", "change"]);
        self.head()
    }

    fn head(&self) -> String {
        git(self.root(), &["rev-parse", "HEAD"]).trim().to_string()
    }

    fn range(&self) -> String {
        format!("{}..{}", self.base, self.head())
    }
}

/// The fake tool folder, the slot folder, the parent of the log folders and a temporary HOME.
struct Run {
    tools: tempfile::TempDir,
    slots: tempfile::TempDir,
    outs: tempfile::TempDir,
    home: tempfile::TempDir,
}

impl Run {
    fn new() -> Run {
        let tools = tempfile::tempdir().unwrap();
        for n in ["cargo", "rustfmt"] {
            let body = if n == "cargo" {
                FAKE_CARGO.to_string()
            } else {
                "#!/bin/sh\nexit 0\n".to_string()
            };
            let path = tools.path().join(n);
            fs::write(&path, body).unwrap();
            fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
        }
        let slots = tempfile::tempdir().unwrap();
        fs::write(
            slots.path().join("limits.json"),
            r#"{"max":4,"min_free_gb":0,"max_swap_pct":100,"min_mem_free_pct":0}"#,
        )
        .unwrap();
        Run {
            tools,
            slots,
            outs: tempfile::tempdir().unwrap(),
            home: tempfile::tempdir().unwrap(),
        }
    }

    fn calls(&self) -> PathBuf {
        self.tools.path().join("calls.log")
    }

    /// The `cargo nextest run` calls so far, oldest first.
    fn nextest_runs(&self) -> Vec<String> {
        fs::read_to_string(self.calls())
            .unwrap_or_default()
            .lines()
            .filter(|l| l.starts_with("cargo nextest run"))
            .map(String::from)
            .collect()
    }

    fn out(&self) -> PathBuf {
        self.outs.path().join("check")
    }

    /// `moai check affected <range> --dir … --out …` with the fakes on PATH. `fail`, `flaky`
    /// and `step_fail` steer the fake cargo for this one call.
    fn go(&self, fx: &Fixture, extra: &[&str], fail: &str, flaky: &str, step_fail: &str) -> Value {
        let tests: Vec<String> = TESTS.iter().map(|(b, t)| format!("{b}:{t}")).collect();
        let git_dir = Command::new("sh")
            .args(["-c", "dirname \"$(command -v git)\""])
            .output()
            .unwrap();
        let path = format!(
            "{}:{}:/usr/bin:/bin",
            self.tools.path().display(),
            String::from_utf8_lossy(&git_dir.stdout).trim()
        );
        let out = Command::new(BIN)
            .args(["check", "affected", &fx.range(), "--dir"])
            .arg(fx.root())
            .arg("--slot-dir")
            .arg(self.slots.path())
            .arg("--out")
            .arg(self.out())
            .args(extra)
            .env("HOME", self.home.path())
            .env_remove("MOAI_HOME")
            .env_remove("XDG_CACHE_HOME")
            .env_remove("CARGO_TARGET_DIR")
            .env_remove("MOAI_HEAVY_SLOT_FD")
            .env_remove("RUST_TEST_THREADS")
            .env_remove("NEXTEST_TEST_THREADS")
            .env("MOAI_TEST_MACHINE", ROOMY)
            .env("PATH", path)
            .env("FAKE_CALLS", self.calls())
            .env("FAKE_TESTS", tests.join(" "))
            .env("FAKE_FAIL", fail)
            .env("FAKE_FLAKY", flaky)
            .env("FAKE_STEP_FAIL", step_fail)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        let text = String::from_utf8_lossy(&out.stdout);
        let at = text.find('{').unwrap_or(0);
        let v: Value = serde_json::from_str(&text[at..])
            .unwrap_or_else(|_| panic!("{text} / {}", String::from_utf8_lossy(&out.stderr)));
        assert_eq!(v["ok"], true, "{v}");
        v["result"].clone()
    }

    fn flaky_cmd(&self, fx: &Fixture) -> std::process::Output {
        Command::new(BIN)
            .args(["check", "flaky", "--dir"])
            .arg(fx.root())
            .env("HOME", self.home.path())
            .env_remove("MOAI_HOME")
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }
}

fn summary(v: &Value) -> String {
    v["summary"].as_str().unwrap().to_string()
}

/// AC-1 (REQ-1, REQ-2, REQ-12): a second run on a head whose first run failed one test runs only
/// that test, says which folder it continues, and passes when the test passes.
#[test]
fn ac1_a_second_run_runs_only_the_failed_test_and_combines_the_result() {
    let fx = Fixture::new();
    let r = Run::new();
    let first = r.go(&fx, &[], "t_b", "", "");
    assert_eq!(first["result"], "FAIL", "{first}");
    assert!(
        !summary(&first).contains("rerun_of="),
        "{}",
        summary(&first)
    );
    let second = r.go(&fx, &[], "", "", "");
    let line = summary(&second);
    assert!(line.contains(" rerun_of="), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
    let runs = r.nextest_runs();
    assert_eq!(runs.len(), 2, "{runs:?}");
    let last = &runs[1];
    assert!(last.contains("test(=t_b)"), "{last}");
    assert!(
        !last.contains("test(=t_a)") && !last.contains("test(=t_c)"),
        "{last}"
    );
    // The folder the second run continues is kept and named.
    let earlier = line
        .split(' ')
        .find_map(|w| w.strip_prefix("rerun_of="))
        .unwrap();
    assert!(
        Path::new(earlier).join("summary.txt").is_file(),
        "{earlier}"
    );
    // The one-line receipt (REQ-12).
    let text = fs::read_to_string(r.out().join("summary.txt")).unwrap();
    assert_eq!(text.trim_end(), line, "{text}");
}

/// AC-1 (REQ-2): while the test still fails the combined result stays FAIL and the log names it.
#[test]
fn ac1_a_test_that_still_fails_keeps_the_result_fail() {
    let fx = Fixture::new();
    let r = Run::new();
    r.go(&fx, &[], "t_b", "", "");
    let second = r.go(&fx, &[], "t_b", "", "");
    let line = summary(&second);
    assert!(line.contains(" rerun_of="), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    let log: String = fs::read_dir(r.out())
        .unwrap()
        .map(|e| e.unwrap().path())
        .filter(|p| p.to_string_lossy().ends_with("-test.log"))
        .map(|p| fs::read_to_string(p).unwrap())
        .collect();
    assert!(log.contains("FAIL") && log.contains(" t_b"), "{log}");
}

/// AC-2 (REQ-3): a head that moved runs the targets of the diff between the two heads and the
/// earlier failure, not the targets the earlier run already passed.
#[test]
fn ac2_a_moved_head_runs_the_diff_targets_and_the_earlier_failures() {
    let fx = Fixture::new();
    let r = Run::new();
    let first = r.go(&fx, &[], "t_b", "", "");
    assert_eq!(first["result"], "FAIL", "{first}");
    fx.commit(&[("src/bundle_pack.rs", "// b\n")]);
    let second = r.go(&fx, &[], "", "", "");
    let line = summary(&second);
    assert!(line.contains(" rerun_of="), "{line}");
    assert!(!line.contains("rerun=full"), "{line}");
    let last = r.nextest_runs().pop().unwrap();
    assert!(last.contains("--test bundle_run_cli"), "{last}");
    assert!(last.contains("test(=t_b)"), "{last}");
    assert!(!last.contains("test(=t_a)"), "{last}");
}

/// AC-2 (REQ-4): a diff that touches a wide path gives the full set and says why.
#[test]
fn ac2_a_wide_path_in_the_moved_diff_gives_the_full_set() {
    let fx = Fixture::new();
    let r = Run::new();
    r.go(&fx, &[], "t_b", "", "");
    fx.commit(&[("Cargo.toml", "[package]\nname = \"y\"\n")]);
    let second = r.go(&fx, &[], "", "", "");
    let line = summary(&second);
    assert!(line.contains(" rerun=full reason="), "{line}");
    let last = r.nextest_runs().pop().unwrap();
    assert!(!last.contains(" -E "), "{last}");
}

/// AC-3 (REQ-6, REQ-7): the test step retries; a test that passes on retry is FLAKY in the
/// summary line, the step log and `flaky.txt`, and the run passes.
#[test]
fn ac3_a_test_that_passes_on_retry_is_listed_as_flaky() {
    let fx = Fixture::new();
    let r = Run::new();
    let v = r.go(&fx, &[], "", "t_c", "");
    let line = summary(&v);
    assert!(line.contains(" flaky=1"), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
    let call = r.nextest_runs().pop().unwrap();
    assert!(call.contains("--retries 2"), "{call}");
    assert!(call.contains("--flaky-result pass"), "{call}");
    let names = fs::read_to_string(r.out().join("flaky.txt")).unwrap();
    assert!(names.contains("t_c"), "{names}");
    let log: String = fs::read_dir(r.out())
        .unwrap()
        .map(|e| e.unwrap().path())
        .filter(|p| p.to_string_lossy().ends_with("-test.log"))
        .map(|p| fs::read_to_string(p).unwrap())
        .collect();
    assert!(log.contains("FLAKY"), "{log}");
}

/// AC-3 (REQ-8): a test that fails every attempt fails the step and lists no FLAKY.
#[test]
fn ac3_a_test_that_fails_every_attempt_fails_the_step() {
    let fx = Fixture::new();
    let r = Run::new();
    let v = r.go(&fx, &[], "t_b", "", "");
    let line = summary(&v);
    assert!(line.contains(" test=100 "), "{line}");
    assert!(line.ends_with(" result=FAIL"), "{line}");
    assert!(!line.contains(" flaky="), "{line}");
}

/// AC-3 (REQ-6): a clean run keeps the summary keys it has today (no `flaky=`, no `rerun_of=`).
#[test]
fn ac3_a_clean_run_keeps_the_summary_keys() {
    let fx = Fixture::new();
    let r = Run::new();
    let v = r.go(&fx, &[], "", "", "");
    let line = summary(&v);
    assert!(
        !line.contains("flaky=") && !line.contains("rerun"),
        "{line}"
    );
    assert!(line.ends_with(" clippy=0 test=0 result=PASS"), "{line}");
}

/// AC-4 (REQ-5): a test log that names no failed test, a missing test log and an earlier run that
/// stopped before the test step give the full set and say why.
#[test]
fn ac4_an_unusable_earlier_run_gives_the_full_set() {
    for case in ["no_failure_line", "log_missing", "stopped_early"] {
        let fx = Fixture::new();
        let r = Run::new();
        match case {
            "stopped_early" => {
                let first = r.go(&fx, &[], "", "", "fmt");
                assert_eq!(first["result"], "FAIL", "{first}");
            }
            _ => {
                r.go(&fx, &[], "t_b", "", "");
                let logs: Vec<PathBuf> = fs::read_dir(r.out())
                    .unwrap()
                    .map(|e| e.unwrap().path())
                    .filter(|p| p.to_string_lossy().ends_with("-test.log"))
                    .collect();
                assert_eq!(logs.len(), 1, "{case}");
                if case == "log_missing" {
                    fs::remove_file(&logs[0]).unwrap();
                } else {
                    fs::write(&logs[0], "        PASS [   0.010s] (1/1) moai-cli::x t\n").unwrap();
                }
            }
        }
        let second = r.go(&fx, &[], "", "", "");
        let line = summary(&second);
        assert!(line.contains(" rerun=full reason="), "{case}: {line}");
        let last = r.nextest_runs().pop().unwrap();
        assert!(!last.contains(" -E "), "{case}: {last}");
        assert!(!line.contains(" rerun_of="), "{case}: {line}");
    }
}

/// AC-5 (REQ-9, REQ-10): each FLAKY name is appended to the ledger, `moai check flaky` counts it
/// and marks a name seen twice `fix`, and an unwritable ledger does not change the run result.
#[test]
fn ac5_the_flaky_ledger_counts_names_and_marks_a_repeat() {
    let fx = Fixture::new();
    let r = Run::new();
    r.go(&fx, &[], "", "t_c", "");
    let once = r.flaky_cmd(&fx);
    let text = String::from_utf8_lossy(&once.stdout).to_string();
    assert!(once.status.success(), "{text}");
    let v: Value = serde_json::from_str(&text[text.find('{').unwrap()..]).unwrap();
    let row = &v["result"]["flaky"][0];
    assert_eq!(row["name"], "moai-cli::update_command_cli t_c", "{v}");
    assert_eq!(row["count"], 1, "{v}");
    assert_ne!(row["status"], "fix", "{v}");
    r.go(&fx, &[], "", "t_c", "");
    let twice = r.flaky_cmd(&fx);
    let text = String::from_utf8_lossy(&twice.stdout).to_string();
    let v: Value = serde_json::from_str(&text[text.find('{').unwrap()..]).unwrap();
    assert_eq!(v["result"]["flaky"][0]["count"], 2, "{v}");
    assert_eq!(v["result"]["flaky"][0]["status"], "fix", "{v}");
}

#[test]
fn ac5_an_unwritable_ledger_does_not_change_the_result() {
    let fx = Fixture::new();
    let r = Run::new();
    // A folder where the ledger file belongs (the `local` folder of the project's state folder):
    // the append fails.
    fs::create_dir_all(
        fx.root()
            .join(moai_cli::state_root::MOAI_DIR)
            .join(moai_cli::state_root::LOCAL_DIR)
            .join("check-flaky.jsonl"),
    )
    .unwrap();
    let v = r.go(&fx, &[], "", "t_c", "");
    let line = summary(&v);
    assert!(line.contains(" flaky=1"), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
}

/// AC-6 (REQ-11): `--full` after a failed run reads no earlier run and runs the full set.
#[test]
fn ac6_full_runs_the_full_set_and_still_records_flaky() {
    let fx = Fixture::new();
    let r = Run::new();
    r.go(&fx, &[], "t_b", "", "");
    let v = r.go(&fx, &["--full"], "", "t_c", "");
    let line = summary(&v);
    assert!(!line.contains(" rerun_of="), "{line}");
    assert!(line.contains(" flaky=1"), "{line}");
    let last = r.nextest_runs().pop().unwrap();
    assert!(!last.contains(" -E "), "{last}");
    assert!(last.contains("--retries 2"), "{last}");
}

/// AC-7 (REQ-12): the combined `summary.txt` is the one receipt line with the keys the receipt
/// reader needs (`base=`, `head=`, `result=`) and names the folder it continues.
#[test]
fn ac7_the_combined_summary_is_one_receipt_line() {
    let fx = Fixture::new();
    let r = Run::new();
    r.go(&fx, &[], "t_b", "", "");
    let v = r.go(&fx, &[], "", "", "");
    let text = fs::read_to_string(r.out().join("summary.txt")).unwrap();
    assert_eq!(text.trim_end().lines().count(), 1, "{text}");
    let line = text.trim_end();
    assert!(line.starts_with("check-affected base="), "{line}");
    assert!(
        line.contains(&format!(" head={} ", &fx.head()[..12])),
        "{line}"
    );
    assert!(line.contains(" rerun_of="), "{line}");
    assert!(line.ends_with(" result=PASS"), "{line}");
    assert_eq!(summary(&v), line);
}
