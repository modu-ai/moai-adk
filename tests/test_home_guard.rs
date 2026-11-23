//! SPEC-20261123-test-home-isolation: `tools/test_home_guard.sh` is the nextest wrapper that runs
//! every test process under an empty temporary HOME holding decoy files, and fails the test that
//! reads or changes that HOME. The wrapper is run directly here (its exit code and lines are its
//! contract) and through a scratch crate with the repository's `.config/nextest.toml`.
//! The wrapper needs `sh`, so the file is for unix hosts.
#![cfg(unix)]

use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};
use tempfile::TempDir;

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const SCRIPT: &str = "tools/test_home_guard.sh";
const CONFIG: &str = ".config/nextest.toml";

/// The decoy a test would read when it opens the user's own state.
/// The decoys are found by file name, so no state folder name is spelled in this file.
const DECOY_READ: &str = r#"find "$HOME" -type f -name user-assets.json -exec cat {} + > /dev/null"#;

fn script() -> PathBuf {
    Path::new(ROOT).join(SCRIPT)
}

/// A folder the wrapper may use as `TMPDIR`, and the runs of the wrapper under it.
struct Guard {
    tmp: TempDir,
}

impl Guard {
    fn new() -> Self {
        Guard {
            tmp: TempDir::new().unwrap(),
        }
    }

    /// Runs the wrapper with `sh -c <probe>` as the "test binary".
    fn run(&self, mode: Option<&str>, log: Option<&Path>, probe: &str) -> Output {
        let mut cmd = Command::new("sh");
        cmd.arg(script()).args(["sh", "-c", probe]);
        cmd.env("TMPDIR", self.tmp.path());
        cmd.env_remove("MOAI_HOME_GUARD");
        cmd.env_remove("MOAI_HOME_GUARD_LOG");
        if let Some(mode) = mode {
            cmd.env("MOAI_HOME_GUARD", mode);
        }
        if let Some(log) = log {
            cmd.env("MOAI_HOME_GUARD_LOG", log);
        }
        cmd.output().unwrap()
    }

    /// The names directly inside the folder, from `ls -A`.
    fn entries(&self) -> Vec<String> {
        let out = Command::new("ls")
            .arg("-A")
            .arg(self.tmp.path())
            .output()
            .unwrap();
        let mut names: Vec<String> = stdout(&out).lines().map(str::to_string).collect();
        names.sort();
        names
    }
}

fn stderr(out: &Output) -> String {
    String::from_utf8_lossy(&out.stderr).into_owned()
}

fn stdout(out: &Output) -> String {
    String::from_utf8_lossy(&out.stdout).into_owned()
}

#[test]
fn enforce_fails_a_test_that_reads_a_decoy() {
    let g = Guard::new();
    let out = g.run(Some("enforce"), None, DECOY_READ);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn enforce_fails_a_test_that_writes_into_its_home() {
    let g = Guard::new();
    let out = g.run(Some("enforce"), None, r#"echo x > "$HOME/stray.txt""#);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn enforce_fails_a_test_that_removes_a_decoy() {
    let g = Guard::new();
    let out = g.run(
        Some("enforce"),
        None,
        r#"find "$HOME" -type f -name settings.json -exec rm {} +"#,
    );
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn a_test_that_touches_nothing_passes_under_a_guarded_home() {
    let g = Guard::new();
    let out = g.run(Some("enforce"), None, r#"printf '%s\n' "$HOME""#);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(!stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
    let home = stdout(&out).trim().to_string();
    let tmp = g.tmp.path().to_string_lossy().into_owned();
    assert!(
        home.starts_with(&format!("{}/moai-test-home.", tmp.trim_end_matches('/'))),
        "HOME is {home}, TMPDIR is {tmp}"
    );
    assert_eq!(g.entries(), Vec::<String>::new(), "the wrapper left a folder");
}

#[test]
fn the_userprofile_follows_the_home() {
    let g = Guard::new();
    let out = g.run(
        Some("enforce"),
        None,
        r#"[ "$HOME" = "$USERPROFILE" ] && echo same"#,
    );
    assert_eq!(stdout(&out).trim(), "same", "{}", stderr(&out));
}

#[test]
fn a_failing_test_keeps_its_own_exit_code_and_the_line() {
    let g = Guard::new();
    let probe = format!("{DECOY_READ}; exit 7");
    let out = g.run(Some("enforce"), None, &probe);
    assert_eq!(out.status.code(), Some(7), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn the_default_mode_observes_and_never_fails_the_test() {
    let g = Guard::new();
    let out = g.run(None, None, DECOY_READ);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn the_log_file_receives_the_line_and_a_bad_log_path_changes_nothing() {
    let g = Guard::new();
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let out = g.run(Some("observe"), Some(&log), DECOY_READ);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let text = fs::read_to_string(&log).unwrap();
    assert!(text.contains("HOME-GUARD:"), "{text}");
    let bad = logs.path().join("no-such-folder").join("guard.log");
    let out = g.run(Some("enforce"), Some(&bad), DECOY_READ);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
}

#[test]
fn an_unwritable_tmpdir_stops_the_wrapper_with_70_before_the_test_runs() {
    let g = Guard::new();
    let marker = g.tmp.path().join("ran");
    let missing = g.tmp.path().join("no-such-folder");
    let probe = format!("echo ran > '{}'", marker.display());
    let out = Command::new("sh")
        .arg(script())
        .args(["sh", "-c", &probe])
        .env("TMPDIR", &missing)
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(70), "{}", stderr(&out));
    assert!(!marker.exists(), "the test ran although the wrapper stopped");
}

#[test]
fn the_wrapper_removes_only_the_folder_it_made() {
    let g = Guard::new();
    fs::create_dir(g.tmp.path().join("keep-me")).unwrap();
    fs::write(g.tmp.path().join("keep-me/file"), "x").unwrap();
    fs::create_dir(g.tmp.path().join("moai-test-home-lookalike")).unwrap();
    let out = g.run(Some("enforce"), None, "true");
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert_eq!(
        g.entries(),
        vec!["keep-me".to_string(), "moai-test-home-lookalike".to_string()]
    );
}

#[test]
fn a_tmpdir_that_is_a_link_loses_nothing_but_the_wrapper_folder() {
    let g = Guard::new();
    let real = g.tmp.path().join("real");
    fs::create_dir(&real).unwrap();
    fs::write(real.join("precious"), "x").unwrap();
    let link = g.tmp.path().join("link");
    std::os::unix::fs::symlink(&real, &link).unwrap();
    let out = Command::new("sh")
        .arg(script())
        .args(["sh", "-c", "true"])
        .env("TMPDIR", &link)
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(real.join("precious").exists());
    assert!(real.exists() && link.exists());
}

#[test]
fn a_cargo_child_of_a_test_still_finds_its_toolchain() {
    let g = Guard::new();
    let out = g.run(Some("enforce"), None, "cargo --version");
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(stdout(&out).starts_with("cargo "), "{}", stdout(&out));
}

/// The test runs under nextest, so the repository's configuration has given it a guarded HOME.
/// Under plain `cargo test` nothing wraps the process and the check does not apply.
#[test]
fn this_process_runs_under_a_guarded_home_when_nextest_runs_it() {
    if std::env::var_os("NEXTEST").is_none() {
        return;
    }
    let home = std::env::var("HOME").unwrap_or_default();
    assert!(
        home.contains("/moai-test-home."),
        "nextest ran this test but HOME is {home:?}, so the wrapper did not wrap it"
    );
}

// ---- the repository configuration, through a scratch crate ------------------------------------

const PLANTS: &str = r#"
fn home() -> std::path::PathBuf {
    std::path::PathBuf::from(std::env::var_os("HOME").unwrap())
}
fn read_by_name(name: &str) {
    let out = std::process::Command::new("find")
        .arg(home())
        .args(["-type", "f", "-name", name, "-exec", "cat", "{}", "+"])
        .output()
        .unwrap();
    assert!(out.status.success());
}
#[test]
fn plant_reads_decoy() {
    read_by_name("user-assets.json");
}
#[test]
fn plant_writes_home() {
    std::fs::write(home().join("stray.txt"), "x").unwrap();
}
#[test]
fn plant_reads_then_fails() {
    read_by_name("language.yaml");
    panic!("plant-own-failure");
}
#[test]
fn plant_idle() {}
#[test]
fn plant_home_is_guarded() {
    let h = home();
    assert!(h.to_string_lossy().contains("/moai-test-home."), "{h:?}");
}
"#;

/// The repository configuration without its per-test overrides: their filters name test binaries
/// of this repository, which a scratch crate does not have, and nextest refuses such a filter.
/// The wrapper rule, the feature switch and the version pin stay as they are.
fn config_without_overrides(config: &str) -> String {
    let mut out = String::new();
    let mut skipping = false;
    for line in config.lines() {
        if line.trim() == "[[profile.default.overrides]]" {
            skipping = true;
            continue;
        }
        if skipping {
            if line.trim_start().starts_with("test-group") {
                skipping = false;
            }
            continue;
        }
        out.push_str(line);
        out.push('\n');
    }
    out
}

enum Wrapper {
    Repository,
    Missing,
    ExitsWith(i32),
}

fn scratch_crate(wrapper: Wrapper) -> TempDir {
    let dir = TempDir::new().unwrap();
    let p = dir.path();
    fs::create_dir_all(p.join("src")).unwrap();
    fs::create_dir_all(p.join(".config")).unwrap();
    fs::create_dir_all(p.join("tools")).unwrap();
    fs::write(
        p.join("Cargo.toml"),
        "[package]\nname = \"guard_probe\"\nversion = \"0.0.0\"\nedition = \"2021\"\n\n[workspace]\n",
    )
    .unwrap();
    fs::write(p.join("src/lib.rs"), PLANTS).unwrap();
    fs::write(p.join(CONFIG), config_without_overrides(&config_text())).unwrap();
    match wrapper {
        Wrapper::Repository => {
            fs::copy(script(), p.join(SCRIPT)).unwrap();
        }
        Wrapper::Missing => {}
        Wrapper::ExitsWith(code) => {
            use std::os::unix::fs::PermissionsExt;
            fs::write(p.join(SCRIPT), format!("#!/bin/sh\nexit {code}\n")).unwrap();
            fs::set_permissions(p.join(SCRIPT), fs::Permissions::from_mode(0o755)).unwrap();
        }
    }
    dir
}

/// `(exit code, stdout and stderr)` of a nextest run of the scratch crate.
fn nextest(dir: &Path, mode: Option<&str>) -> (Option<i32>, String) {
    let mut cmd = Command::new("cargo");
    cmd.args([
        "nextest",
        "run",
        "--offline",
        "--no-fail-fast",
        "--color",
        "never",
        "--success-output",
        "immediate",
        "--failure-output",
        "immediate",
    ])
    .current_dir(dir);
    for (name, _) in std::env::vars_os() {
        let n = name.to_string_lossy().into_owned();
        let own = n.starts_with("NEXTEST") || n.starts_with("MOAI_HOME_GUARD");
        let cargo = n.starts_with("CARGO_") && n != "CARGO_HOME";
        if own || cargo {
            cmd.env_remove(&name);
        }
    }
    cmd.env("CARGO_TARGET_DIR", dir.join("target"));
    cmd.env("CARGO_INCREMENTAL", "0");
    if let Some(mode) = mode {
        cmd.env("MOAI_HOME_GUARD", mode);
    }
    let out = cmd.output().expect("cargo nextest must be installed for this test");
    let text = format!("{}\n{}", stdout(&out), stderr(&out));
    (out.status.code(), text)
}

/// The first status word nextest printed for the test `name`.
fn status(text: &str, name: &str) -> Option<String> {
    text.lines().find_map(|line| {
        let t = line.trim();
        let word = t.split_whitespace().next()?;
        let known = ["PASS", "FAIL", "SIGABRT", "SIGSEGV", "TIMEOUT", "ERROR", "EXEC-FAIL"];
        (known.contains(&word) && t.ends_with(&format!(" {name}"))).then(|| word.to_string())
    })
}

fn require_nextest() {
    let ok = Command::new("cargo")
        .args(["nextest", "--version"])
        .output()
        .map(|o| o.status.success())
        .unwrap_or(false);
    assert!(ok, "cargo-nextest is not installed, so the wrapper cannot be exercised here");
}

#[test]
fn nextest_applies_the_repository_configuration_to_a_scratch_crate() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Repository);
    let (code, text) = nextest(dir.path(), Some("enforce"));
    let st = |n: &str| status(&text, n);
    assert_eq!(st("plant_reads_decoy").as_deref(), Some("FAIL"), "{text}");
    assert_eq!(st("plant_writes_home").as_deref(), Some("FAIL"), "{text}");
    assert_eq!(st("plant_reads_then_fails").as_deref(), Some("FAIL"), "{text}");
    assert!(text.contains("plant-own-failure"), "the test's own failure is lost: {text}");
    assert_eq!(st("plant_idle").as_deref(), Some("PASS"), "{text}");
    assert_eq!(st("plant_home_is_guarded").as_deref(), Some("PASS"), "{text}");
    assert!(text.contains("HOME-GUARD:"), "{text}");
    assert_eq!(code, Some(100), "{text}");
}

#[test]
fn the_default_mode_of_the_repository_configuration_does_not_fail_a_reading_test() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Repository);
    let (_, text) = nextest(dir.path(), None);
    assert_eq!(status(&text, "plant_reads_decoy").as_deref(), Some("PASS"), "{text}");
    assert!(text.contains("HOME-GUARD:"), "{text}");
}

#[test]
fn a_missing_wrapper_does_not_let_a_test_pass() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Missing);
    let (code, text) = nextest(dir.path(), Some("enforce"));
    let st = status(&text, "plant_idle");
    assert!(st.is_some(), "the test was never attempted: {text}");
    assert_ne!(st.as_deref(), Some("PASS"), "{text}");
    assert_ne!(code, Some(0), "{text}");
}

#[test]
fn a_wrapper_that_exits_non_zero_does_not_let_a_test_pass() {
    require_nextest();
    let dir = scratch_crate(Wrapper::ExitsWith(3));
    let (code, text) = nextest(dir.path(), Some("enforce"));
    let st = status(&text, "plant_idle");
    assert!(st.is_some(), "the test was never attempted: {text}");
    assert_ne!(st.as_deref(), Some("PASS"), "{text}");
    assert_ne!(code, Some(0), "{text}");
}

// ---- the configuration file and its allowlist ---------------------------------------------------

/// The `# allow: <predicate> | reason: <text>` lines of the configuration. Each must have a
/// reason, appear in the filter of the wrapper rule, and name a test binary or a test function
/// that exists under `root`. Returns the number of entries.
fn check_allowlist(config: &str, root: &Path) -> Result<usize, String> {
    let filter_line = config
        .lines()
        .find(|l| l.contains("# home-guard-filter"))
        .unwrap_or("");
    let mut entries = 0;
    for line in config.lines() {
        let Some(rest) = line.trim_start().strip_prefix("# allow:") else {
            continue;
        };
        entries += 1;
        let (predicate, reason) = rest
            .split_once("| reason:")
            .ok_or_else(|| format!("allowlist entry without a reason: {line}"))?;
        let predicate = predicate.trim();
        if reason.trim().is_empty() {
            return Err(format!("allowlist entry with an empty reason: {line}"));
        }
        if !filter_line.contains(predicate) {
            return Err(format!("entry is not in the wrapper filter: {predicate}"));
        }
        if let Some(name) = predicate
            .strip_prefix("binary(")
            .and_then(|s| s.strip_suffix(')'))
        {
            let file = root.join("tests").join(format!("{name}.rs"));
            let dir = root.join("tests").join(name).join("main.rs");
            if !file.exists() && !dir.exists() {
                return Err(format!("entry names no test binary: {predicate}"));
            }
        } else if let Some(name) = predicate
            .strip_prefix("test(=")
            .and_then(|s| s.strip_suffix(')'))
        {
            if !function_exists(root, name) {
                return Err(format!("entry names no test function: {predicate}"));
            }
        } else {
            return Err(format!("entry is neither binary(..) nor test(=..): {predicate}"));
        }
    }
    Ok(entries)
}

/// Whether the test function `module::path::name` of the library exists: the first source file
/// the module path leads to (`src/a/b.rs`, `src/a/b/mod.rs`, then the same for `a`) holds
/// `fn name(`. One path is read, no folder is walked.
fn function_exists(root: &Path, path: &str) -> bool {
    let parts: Vec<&str> = path.split("::").collect();
    let Some((name, modules)) = parts.split_last() else {
        return false;
    };
    for depth in (1..=modules.len()).rev() {
        let base = modules[..depth].join("/");
        for rel in [format!("src/{base}.rs"), format!("src/{base}/mod.rs")] {
            if let Ok(text) = fs::read_to_string(root.join(&rel)) {
                return text.contains(&format!("fn {name}("));
            }
        }
    }
    false
}

fn config_text() -> String {
    fs::read_to_string(Path::new(ROOT).join(CONFIG)).unwrap()
}

#[test]
fn the_repository_allowlist_is_well_formed() {
    check_allowlist(&config_text(), Path::new(ROOT)).unwrap();
}

#[test]
fn an_allowlist_entry_without_a_reason_is_refused() {
    let cfg = "# allow: binary(test_home_guard)\nfilter = 'all() - binary(test_home_guard)' # home-guard-filter\n";
    let err = check_allowlist(cfg, Path::new(ROOT)).unwrap_err();
    assert!(err.contains("without a reason"), "{err}");
    let cfg = "# allow: binary(test_home_guard) | reason:   \nfilter = 'all() - binary(test_home_guard)' # home-guard-filter\n";
    let err = check_allowlist(cfg, Path::new(ROOT)).unwrap_err();
    assert!(err.contains("empty reason"), "{err}");
}

#[test]
fn an_allowlist_entry_that_names_nothing_is_refused() {
    let cfg = "# allow: binary(no_such_binary) | reason: a reason\nfilter = 'all() - binary(no_such_binary)' # home-guard-filter\n";
    let err = check_allowlist(cfg, Path::new(ROOT)).unwrap_err();
    assert!(err.contains("names no test binary"), "{err}");
    let cfg = "# allow: test(=no_such_function_anywhere) | reason: a reason\nfilter = 'all() - test(=no_such_function_anywhere)' # home-guard-filter\n";
    let err = check_allowlist(cfg, Path::new(ROOT)).unwrap_err();
    assert!(err.contains("names no test function"), "{err}");
}

#[test]
fn an_allowlist_entry_outside_the_wrapper_filter_is_refused() {
    let cfg = "# allow: binary(test_home_guard) | reason: a reason\nfilter = 'all()' # home-guard-filter\n";
    let err = check_allowlist(cfg, Path::new(ROOT)).unwrap_err();
    assert!(err.contains("not in the wrapper filter"), "{err}");
}

#[test]
fn the_configuration_enables_the_wrapper_feature_and_pins_a_nextest_version() {
    let cfg = config_text();
    assert!(
        cfg.contains(r#"experimental = ["wrapper-scripts"]"#),
        "the wrapper-scripts feature is not enabled"
    );
    assert!(cfg.contains("nextest-version"), "no nextest-version is named");
    assert!(cfg.contains("run-wrapper = 'home-guard'"), "no rule runs the wrapper");
    assert!(cfg.contains("cfg(unix)"), "the rule is not limited to unix hosts");
    assert!(cfg.contains("tools/test_home_guard.sh"), "the wrapper script is not named");
}
