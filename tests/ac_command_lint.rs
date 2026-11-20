//! Hygiene lints for the test suite and recorded verification commands (#2352).
//!
//! - git fixture calls must go through `tests/common/git.rs`: a test file that
//!   spawns git outside that helper (and without the gc/maintenance guards in its
//!   wrapper) misses the guards that keep fixture repositories stable
//!   (#2082/#2092), and the scan points new tests at the shared helper.
//! - recorded `cargo test --test X …` verification commands must name an existing
//!   target and select at least one test — cargo exits 0 with "running 0 tests"
//!   for a filter that matches nothing, which has silently passed as verification
//!   more than once (survey 2.11, s01-scope-05).

mod common;

use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");

fn tests_dir() -> PathBuf {
    Path::new(MANIFEST).join("tests")
}

fn rs_sources(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(dir) = stack.pop() {
        let entries = match fs::read_dir(&dir) {
            Ok(entries) => entries,
            Err(_) => continue,
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                stack.push(path);
            } else if path.extension().is_some_and(|e| e == "rs") {
                out.push(path);
            }
        }
    }
    out.sort();
    out
}

/// Files that carry their own git spawns must guard them: a wrapper that spawns git
/// without the fixture guards is exactly what this card routes into
/// `tests/common/git.rs`. This file is exempt — it quotes the pattern it scans for.
fn files_with_unguarded_git(dir: &Path) -> Vec<String> {
    let mut violations = Vec::new();
    for path in rs_sources(dir) {
        let rel = path
            .strip_prefix(dir)
            .unwrap()
            .to_string_lossy()
            .to_string();
        if rel.starts_with("common/") || rel == "ac_command_lint.rs" {
            continue;
        }
        let text =
            fs::read_to_string(&path).unwrap_or_else(|e| panic!("read {}: {e}", path.display()));
        let has_spawn = text.contains("Command::new(\"git\")");
        let guarded = text.contains("gc.auto=0") && text.contains("maintenance.auto=false");
        if has_spawn && !guarded {
            violations.push(format!("tests/{rel}"));
        }
    }
    violations
}

#[test]
fn git_fixture_calls_live_in_the_shared_helper() {
    let violations = files_with_unguarded_git(&tests_dir());
    assert!(
        violations.is_empty(),
        "test files spawning git without the gc/maintenance fixture guards — route \
         them through tests/common/git.rs or add the guards:\n{}",
        violations.join("\n")
    );
}

/// Behavioural check: the guards must reach git itself, not just appear in the
/// helper's text (its doc comment names them, so a text match alone survives a
/// mutation of the actual argument values).
#[test]
fn git_helper_applies_the_guards_to_every_call() {
    let dir = tempfile::TempDir::new().unwrap();
    common::git::git_ok(dir.path(), &["init", "-q"]);
    assert_eq!(
        common::git::git(dir.path(), &["config", "--get", "gc.auto"]),
        "0",
        "git() must run with gc.auto=0"
    );
    assert_eq!(
        common::git::git(dir.path(), &["config", "--get", "maintenance.auto"]),
        "false",
        "git() must run with maintenance.auto=false"
    );
    let raw = common::git::git_raw(dir.path(), &["config", "--get", "gc.auto"]);
    assert_eq!(raw.stdout.trim(), "0", "git_raw() must run with gc.auto=0");
    // git_pinned only asserts success; a value pattern makes `git config --get`
    // exit 1 unless the shared factory supplied exactly the expected guard.
    common::git::git_pinned(dir.path(), &["config", "--get", "gc.auto", "^0$"], None);
    common::git::git_pinned(
        dir.path(),
        &["config", "--get", "maintenance.auto", "^false$"],
        None,
    );
    let global = common::git::git_raw(dir.path(), &["config", "--global", "--list"]);
    assert!(
        global.stdout.trim().is_empty(),
        "GIT_CONFIG_GLOBAL must point at an empty config: {}",
        global.stdout
    );
}

/// Runs the recorded-AC-command lint with the given arguments and returns its exit code.
fn run_lint(args: &[&str]) -> std::process::Output {
    // The port runs as the `moai ac-commands-check` command (tools slice 2, the
    // `pr-self-audit-check` raw-text precedent): the tests drive the same command the users
    // and the lanes would, never the deleted Python tool.
    let mut cmd = Command::new(common::MOAI_CLI_BIN);
    cmd.arg("ac-commands-check");
    for arg in args {
        cmd.arg(arg);
    }
    cmd.output().expect("spawn the AC command lint")
}

/// #2661: the lint calls `cargo test --test <name> -- --list`, and that cargo call re-copies
/// target/debug/moai-cli even when nothing is rebuilt. Under cargo-nextest every test of this
/// binary must run alone, or a test spawning the binary in that moment fails with NotFound.
#[test]
fn nextest_runs_this_binary_alone() {
    let config = fs::read_to_string(Path::new(MANIFEST).join(".config/nextest.toml"))
        .expect(".config/nextest.toml must exist");
    let alone = config
        .split("[[profile.default.overrides]]")
        .skip(1)
        .any(|block| {
            let block = block.split("\n[").next().unwrap_or("");
            let value = |key: &str| {
                block.lines().find_map(|line| {
                    let (k, v) = line.split_once('=')?;
                    (k.trim() == key).then(|| v.trim().trim_matches(['\'', '"']).to_string())
                })
            };
            value("filter").as_deref() == Some("binary(ac_command_lint)")
                && value("threads-required").as_deref() == Some("num-test-threads")
        });
    assert!(
        alone,
        "an override with filter = 'binary(ac_command_lint)' and \
         threads-required = 'num-test-threads' must stay in .config/nextest.toml:\n{config}"
    );
}

#[test]
fn recorded_ac_commands_name_existing_targets_and_select_tests() {
    let out = run_lint(&["--tree", MANIFEST]);
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        out.status.success(),
        "recorded `cargo test --test …` commands must name existing targets and \
         select at least one test:\n{stdout}{stderr}"
    );
}

#[test]
fn lint_flags_a_zero_selection_command() {
    let dir = tempfile::TempDir::new().unwrap();
    let report = dir.path().join("zero-select-report.md");
    fs::write(
        &report,
        "verification: `cargo test --offline --locked --test product_cli \
         definitely_not_a_selected_test -- --nocapture`\n",
    )
    .unwrap();
    let file = report.to_str().unwrap();
    let out = run_lint(&["--file", file]);
    let combined = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(
        !out.status.success(),
        "a `--test` command whose filters select nothing must be flagged:\n{combined}"
    );
    assert!(
        combined.contains(file),
        "violation must name the file: {combined}"
    );
}

// ---- the `moai ac-commands` contract (tools slice 2) ----
//
// The stub `cargo` answers every `cargo test … -- --list` from one env block and logs each
// call, so the cache, the wording classes and the scan-set edges are asserted from observed
// calls, not from time. The stubs are shell scripts, so these tests are Unix-only: a
// Windows stub needs its own flavour (batch) and stays open — see the port card's plan.

const STUB_CARGO: &str = "#!/bin/sh
printf '%s\\n' \"$*\" >> \"$CARGO_CALL_LOG\"
case \"$*\" in
  *--list*)
    for a in \"$@\"; do
      case \"$a\" in
        fixture_target) printf '%b\\n' \"$CARGO_LIST_CONTENT\" ;;
      esac
    done
    ;;
esac
";

struct AcFixture {
    dir: tempfile::TempDir,
    path_env: String,
}

/// A tree with the scan roots and the stub cargo on PATH; `target_file` also lays down
/// `tests/<name>.rs` for the two targets the stub's listing knows (`fixture_target`) or
/// answers with nothing about (`empty_target`).
fn ac_fixture(target_file: bool) -> AcFixture {
    let dir = tempfile::TempDir::new().unwrap();
    let bin = dir.path().join("stubbin");
    fs::create_dir_all(&bin).unwrap();
    fs::write(bin.join("cargo"), STUB_CARGO).unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(bin.join("cargo"), fs::Permissions::from_mode(0o755)).unwrap();
    }
    fs::create_dir_all(dir.path().join("docs")).unwrap();
    fs::create_dir_all(dir.path().join(".moai-cli/reports/issue-bodies")).unwrap();
    if target_file {
        fs::create_dir_all(dir.path().join("tests")).unwrap();
        fs::write(dir.path().join("tests/fixture_target.rs"), "fn main() {}\n").unwrap();
        fs::write(dir.path().join("tests/empty_target.rs"), "fn main() {}\n").unwrap();
    }
    let path_env = format!(
        "{}:{}",
        bin.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    AcFixture { dir, path_env }
}

fn record(dir: &tempfile::TempDir, rel: &str, body: &str) -> String {
    let path = dir.path().join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, body).unwrap();
    path.to_string_lossy().to_string()
}

fn lint_with_stub(f: &AcFixture, args: &[&str]) -> std::process::Output {
    let mut cmd = Command::new(common::MOAI_CLI_BIN);
    cmd.arg("ac-commands-check")
        .args(args)
        .current_dir(f.dir.path())
        .env("PATH", &f.path_env)
        .env("CARGO_CALL_LOG", f.dir.path().join("cargo-calls.log"))
        .env(
            "CARGO_LIST_CONTENT",
            "fixture_target::run_a: test\nfixture_target::run_b: test\n",
        );
    cmd.output()
        .expect("spawn the AC command lint with the stub cargo")
}

#[test]
#[cfg(unix)]
fn an_issue_body_2152_stays_unjudged_and_2153_is_judged() {
    let f = ac_fixture(true);
    let early = record(
        &f.dir,
        ".moai-cli/reports/issue-bodies/2152-early.md",
        "verified by `cargo test --offline --locked --test fixture_target no_such_filter`\n",
    );
    let _ = early;
    let late = record(
        &f.dir,
        ".moai-cli/reports/issue-bodies/2153-new.md",
        "verified by `cargo test --offline --locked --test fixture_target no_such_filter`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "the 2153 record's zero-select command must be flagged:\n{stderr}"
    );
    assert!(
        stderr.contains(&late),
        "the violation names the 2153 file:\n{stderr}"
    );
    assert!(
        !stderr.contains("2152-early"),
        "the frozen 2152 record stays unjudged:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn an_env_prefixed_command_is_judged_after_the_prefix() {
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "run: `CARGO_HOME=/tmp/x MOAI_B=1 cargo test --offline --locked --test fixture_target no_such_filter`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "the env-prefixed command must still be judged:\n{stderr}"
    );
    assert!(
        stderr
            .contains("filters ['no_such_filter'] select 0 tests in fixture_target (1 available)"),
        "the prefix never becomes a filter or a target:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn a_placeholder_filter_is_skipped() {
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "here: `cargo test --offline --locked --test fixture_target <이름>`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    assert!(
        out.status.success(),
        "a <placeholder> filter is a template, not a violation:\n{}",
        String::from_utf8_lossy(&out.stdout)
    );
}

#[test]
#[cfg(unix)]
fn a_missing_target_is_flagged_with_its_wording() {
    let f = ac_fixture(false);
    record(
        &f.dir,
        "docs/a.md",
        "check: `cargo test --offline --locked --test ghost_target one_filter`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "a command naming a target with no tests/<name>.rs must be flagged:\n{stderr}"
    );
    assert!(
        stderr.contains("target tests/ghost_target.rs does not exist"),
        "the wording class of a missing target:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn a_target_that_lists_no_tests_is_flagged() {
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "check: `cargo test --offline --locked --test empty_target one_filter`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "a target whose listing has no tests must be flagged:\n{stderr}"
    );
    assert!(
        stderr.contains("target empty_target lists no tests"),
        "the wording class of an empty listing:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn one_listing_serves_every_recorded_file_of_the_same_target() {
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "`cargo test --offline --locked --test fixture_target missing_filter`\n",
    );
    record(
        &f.dir,
        ".moai-cli/reports/r.md",
        "`cargo test --offline --locked --test fixture_target missing_filter`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    assert!(
        !out.status.success(),
        "both zero-select commands must be flagged:\n{stdout}"
    );
    let log = fs::read_to_string(f.dir.path().join("cargo-calls.log")).unwrap_or_default();
    assert_eq!(
        log.lines().filter(|l| l.contains("--list")).count(),
        1,
        "one listing per target per invocation:\n{log}"
    );
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        stderr.contains("checked 2 recorded files, 1 --test targets listed"),
        "the summary line keeps its shape on the failure path:\n{stderr}"
    );
    assert_eq!(
        stderr.matches("VIOLATION: ").count(),
        2,
        "both files are judged from the one listing:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn file_mode_judges_the_one_file_against_the_callers_cwd() {
    let f = ac_fixture(true);
    let file = record(
        &f.dir,
        "notes/a.md",
        "`cargo test --offline --locked --test fixture_target missing_filter`\n",
    );
    let out = lint_with_stub(&f, &["--file", &file]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "--file judges the one file:\n{stderr}"
    );
    assert!(
        stderr.contains(&file),
        "the violation names the given file:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn a_doubled_backtick_pair_opens_the_span_inside() {
    // The Python regex scans one character at a time: in ``cmd`` the second backtick
    // opens the span whose closer is the third — the audit found the Rust port missing
    // these (17 findings against 1 on the rc1-issue-design record).
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "x ``cargo test --offline --locked --test fixture_target no_such_filter`` y\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "the span inside a doubled pair is a span:\n{stderr}"
    );
    assert!(
        stderr
            .contains("filters ['no_such_filter'] select 0 tests in fixture_target (1 available)"),
        "the doubled pair's span was judged:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn an_unclosed_backtick_is_not_a_span() {
    // `` `([^`\n]+)` `` needs a closer: an unclosed backtick never becomes a span, so its
    // line is not judged at all.
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "x `cargo test --offline --locked --test ghost_target one\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        out.status.success(),
        "an unclosed backtick is not a span — no violation:\n{stderr}"
    );
    assert!(
        !stderr.contains("does not exist"),
        "the unclosed line never reached the judge:\n{stderr}"
    );
}

#[test]
#[cfg(unix)]
fn a_double_dash_thread_flag_is_not_a_target() {
    // `--test\s+` requires the whitespace: `--test-threads` is not a `--test` — the
    // audit found the Rust port naming `-threads` as the target here.
    let f = ac_fixture(true);
    record(
        &f.dir,
        "docs/a.md",
        "run: `cargo test --offline --locked --test-threads 4 --test fixture_target one`\n",
    );
    let out = lint_with_stub(&f, &["--tree", f.dir.path().to_str().unwrap()]);
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert!(
        !out.status.success(),
        "the real `--test fixture_target` is judged and `one` selects none:\n{stderr}"
    );
    assert!(
        stderr.contains("filters ['one'] select 0 tests in fixture_target (1 available)"),
        "the target is fixture_target, judged by the tool's filter rule:\n{stderr}"
    );
    assert!(
        !stderr.contains("does not exist"),
        "-threads was never named as a target:\n{stderr}"
    );
}
