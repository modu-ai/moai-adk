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

/// SPEC-20261122-ac-lint-no-cargo: the lint reads test names from the sources and starts no
/// cargo, so the nextest special case that made this binary run alone (#2661) has no reason.
#[test]
fn nextest_has_no_special_case_for_this_binary() {
    let config = fs::read_to_string(Path::new(MANIFEST).join(".config/nextest.toml"))
        .expect(".config/nextest.toml must exist");
    let special = config
        .split("[[profile.default.overrides]]")
        .skip(1)
        .any(|block| {
            let block = block.split("\n[").next().unwrap_or("");
            block.lines().any(|line| {
                line.split_once('=').is_some_and(|(k, v)| {
                    k.trim() == "filter" && v.contains("binary(ac_command_lint)")
                })
            })
        });
    assert!(
        !special,
        "no override for binary(ac_command_lint) may stay in .config/nextest.toml:\n{config}"
    );
}

/// Runs the lint with a fake `cargo` first on PATH, which shadows the real one and records each
/// start in `cargo-started`. Returns the output and whether cargo was started.
#[cfg(unix)]
fn run_lint_without_cargo(args: &[&str]) -> (std::process::Output, bool) {
    use std::os::unix::fs::PermissionsExt;
    let bin = tempfile::TempDir::new().unwrap();
    let marker = bin.path().join("cargo-started");
    let fake = bin.path().join("cargo");
    fs::write(
        &fake,
        format!("#!/bin/sh\necho \"$@\" >> '{}'\nexit 1\n", marker.display()),
    )
    .unwrap();
    fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    let path = std::env::var_os("PATH").unwrap_or_default();
    let path = std::env::join_paths(
        std::iter::once(bin.path().to_path_buf()).chain(std::env::split_paths(&path)),
    )
    .unwrap();
    let out = Command::new("python3")
        .arg("-I")
        .arg(Path::new(MANIFEST).join("tools/check_ac_commands.py"))
        .args(args)
        .env("PATH", path)
        .output()
        .expect("spawn the AC command lint");
    (out, marker.exists())
}

#[cfg(unix)]
fn text_of(out: &std::process::Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

#[cfg(unix)]
#[test]
fn the_repository_lint_starts_no_cargo_and_ends_within_five_seconds() {
    let started = std::time::Instant::now();
    let (out, cargo_started) = run_lint_without_cargo(&["--tree", MANIFEST]);
    let took = started.elapsed();
    assert!(
        !cargo_started,
        "the lint must not start cargo:\n{}",
        text_of(&out)
    );
    assert!(
        out.status.success(),
        "the recorded commands must pass without cargo:\n{}",
        text_of(&out)
    );
    assert!(
        took < std::time::Duration::from_secs(5),
        "the lint took {took:?}; it must end within 5 s"
    );
}

/// Writes `files` (relative path, text) under a new temporary tree.
#[cfg(unix)]
fn tree(files: &[(&str, &str)]) -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    for (rel, text) in files {
        let path = dir.path().join(rel);
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        fs::write(path, text).unwrap();
    }
    dir
}

#[cfg(unix)]
const SHAPES: &str = r##"// a brace in a comment: {
/* a block comment with a brace } and a /* nested */ one */
const TEXT: &str = "not a module { mod fake { #[test] fn in_a_string() {} } }";
const RAW: &str = r#"#[test] fn in_a_raw_string() {}"#;
const OPEN: char = '{';
fn borrow<'a>(x: &'a str) -> &'a str { x }

#[test]
fn top_level() {}

mod inner {
    #[test]
    // a comment between the attributes
    #[ignore = "kept for the list"]
    fn nested_one() {
        let _ = "}";
    }

    mod deeper {
        #[tokio::test]
        async fn deepest() {}
    }
}

mod helpers;

#[path = "elsewhere/odd_name.rs"]
mod renamed;

#[test]
fn after_the_modules() {}
"##;

#[cfg(unix)]
#[test]
fn test_names_are_read_from_the_sources_with_their_module_paths() {
    let dir = tree(&[
        ("tests/shapes.rs", SHAPES),
        ("tests/helpers.rs", "#[test]\nfn from_a_module_file() {}\n"),
        (
            "tests/elsewhere/odd_name.rs",
            "#[test]\nfn from_a_path_attribute() {}\n",
        ),
        (
            "docs/good.md",
            "`cargo test --test shapes top_level`\n\
             `cargo test --test shapes inner::nested_one`\n\
             `cargo test --test shapes inner::deeper::deepest`\n\
             `cargo test --test shapes helpers::from_a_module_file`\n\
             `cargo test --test shapes renamed::from_a_path_attribute`\n\
             `cargo test --test shapes after_the_modules`\n",
        ),
        (
            "docs/bad.md",
            "`cargo test --test shapes in_a_string`\n\
             `cargo test --test shapes in_a_raw_string`\n\
             `cargo test --test shapes inner::after_the_modules`\n\
             `cargo test --test shapes deeper::nested_one`\n",
        ),
    ]);
    let root = dir.path().to_str().unwrap();
    let (out, cargo_started) = run_lint_without_cargo(&["--tree", root]);
    let text = text_of(&out);
    assert!(!cargo_started, "the lint must not start cargo:\n{text}");
    let violations: Vec<&str> = text
        .lines()
        .filter(|l| l.starts_with("VIOLATION:"))
        .collect();
    assert!(
        violations.iter().all(|l| l.contains("docs/bad.md")),
        "every good filter must select a read test:\n{text}"
    );
    for filter in [
        "in_a_string",
        "in_a_raw_string",
        "inner::after_the_modules",
        "deeper::nested_one",
    ] {
        assert!(
            violations
                .iter()
                .any(|l| l.contains(&format!("'{filter}'"))),
            "the filter {filter} names no test and must be a violation:\n{text}"
        );
    }
    assert_eq!(violations.len(), 4, "{text}");
    assert!(!out.status.success(), "{text}");
}

#[cfg(unix)]
#[test]
fn a_filter_on_macro_made_tests_says_the_names_cannot_be_read() {
    let dir = tree(&[
        (
            "tests/made.rs",
            "macro_rules! case {\n    ($name:ident) => {\n        #[test]\n        fn $name() {}\n    };\n}\n\
             case!(made_by_macro);\n",
        ),
        ("docs/record.md", "`cargo test --test made made_by_macro`\n"),
    ]);
    let root = dir.path().to_str().unwrap();
    let (out, cargo_started) = run_lint_without_cargo(&["--tree", root]);
    let text = text_of(&out);
    assert!(!cargo_started, "the lint must not start cargo:\n{text}");
    assert!(!out.status.success(), "{text}");
    assert!(
        text.lines().any(|l| l.starts_with("VIOLATION:")
            && l.contains("docs/record.md")
            && l.contains("macro")),
        "the violation must say the names are made by a macro:\n{text}"
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
