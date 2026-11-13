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
    let script = Path::new(MANIFEST).join("tools/check_ac_commands.py");
    let mut cmd = Command::new("python3");
    cmd.arg("-I").arg(&script);
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
