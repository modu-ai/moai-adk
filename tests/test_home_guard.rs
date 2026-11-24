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
const DECOY_READ: &str =
    r#"find "$HOME" -type f -name user-assets.json -exec cat {} + > /dev/null"#;

/// Reads only that decoy, the way the baselined tests do: the glob lists the HOME and opens the
/// one file, and lists no other folder of it.
const KNOWN_READ: &str = r#"cat "$HOME"/.[!.]*/user-assets.json > /dev/null"#;

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
    assert_eq!(
        g.entries(),
        Vec::<String>::new(),
        "the wrapper left a folder"
    );
}

#[test]
fn every_decoy_starts_with_an_access_time_before_2001() {
    let g = Guard::new();
    // observe mode: the find of the probe reads folders, which enforce would count as a touch.
    let out = g.run(
        Some("observe"),
        None,
        r#"test -z "$(find "$HOME" -type f -newerat 2001-01-01)" && test -n "$(find "$HOME" -type f)""#,
    );
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
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
fn the_default_mode_is_enforce() {
    let g = Guard::new();
    let out = g.run(None, None, DECOY_READ);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn observe_mode_prints_the_line_and_never_fails_the_test() {
    let g = Guard::new();
    let out = g.run(Some("observe"), None, DECOY_READ);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

// ---- the baseline: tests that read the known decoy without choosing a HOME -----------------------

/// The wrapper run on a fake test binary the way nextest starts one: the binary path whose name
/// ends in a 16-digit hash, then `--exact <test name> --nocapture`. `body` is the shell text of
/// the binary; `baseline` is the baseline file the wrapper reads.
fn run_fake_binary(g: &Guard, baseline: &str, body: &str, test: &str) -> Output {
    use std::os::unix::fs::PermissionsExt;
    let bin_dir = TempDir::new().unwrap();
    let bin = bin_dir.path().join("fake_bin-0123456789abcdef");
    fs::write(&bin, format!("#!/bin/sh\n{body}\n")).unwrap();
    fs::set_permissions(&bin, fs::Permissions::from_mode(0o755)).unwrap();
    let list = bin_dir.path().join("baseline.tsv");
    fs::write(&list, baseline).unwrap();
    Command::new("sh")
        .arg(script())
        .arg(&bin)
        .args(["--exact", test, "--nocapture"])
        .env("TMPDIR", g.tmp.path())
        .env("MOAI_HOME_GUARD_BASELINE", &list)
        .env_remove("MOAI_HOME_GUARD")
        .env_remove("MOAI_HOME_GUARD_LOG")
        .output()
        .unwrap()
}

#[test]
fn a_listed_test_that_reads_only_the_known_decoy_is_tolerated() {
    let g = Guard::new();
    let out = run_fake_binary(&g, "fake_bin\tsome::test\n", KNOWN_READ, "some::test");
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(stderr(&out).contains("tolerated"), "{}", stderr(&out));
}

#[test]
fn the_baseline_next_to_the_script_is_found_without_a_path_setting() {
    // Needs a test of the baseline in the repository: its first listed line names the fake binary.
    let text = fs::read_to_string(Path::new(ROOT).join(BASELINE)).unwrap();
    let line = text
        .lines()
        .find(|l| !l.starts_with('#') && !l.is_empty())
        .unwrap();
    let (bin, name) = line.split_once('\t').unwrap();
    let g = Guard::new();
    let bin_dir = TempDir::new().unwrap();
    let fake = bin_dir.path().join(format!("{bin}-0123456789abcdef"));
    fs::write(&fake, format!("#!/bin/sh\n{KNOWN_READ}\n")).unwrap();
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let out = Command::new("sh")
        .arg(script())
        .arg(&fake)
        .args(["--exact", name, "--nocapture"])
        .env("TMPDIR", g.tmp.path())
        .env_remove("MOAI_HOME_GUARD")
        .env_remove("MOAI_HOME_GUARD_LOG")
        .env_remove("MOAI_HOME_GUARD_BASELINE")
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(stderr(&out).contains("tolerated"), "{}", stderr(&out));
}

#[test]
fn a_listed_test_that_also_writes_is_not_tolerated() {
    let g = Guard::new();
    let body = format!("{KNOWN_READ}\necho x > \"$HOME/stray.txt\"");
    let out = run_fake_binary(&g, "fake_bin\tsome::test\n", &body, "some::test");
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    assert!(!stderr(&out).contains("tolerated"), "{}", stderr(&out));
}

#[test]
fn a_listed_test_that_reads_another_decoy_is_not_tolerated() {
    let g = Guard::new();
    let body = r#"find "$HOME" -type f -name language.yaml -exec cat {} + > /dev/null"#;
    let out = run_fake_binary(&g, "fake_bin\tsome::test\n", body, "some::test");
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
}

#[test]
fn a_test_that_is_not_listed_fails_and_so_does_one_under_another_binary() {
    let g = Guard::new();
    let out = run_fake_binary(&g, "fake_bin\tother::test\n", KNOWN_READ, "some::test");
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    let out = run_fake_binary(&g, "other_bin\tsome::test\n", KNOWN_READ, "some::test");
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
}

#[test]
fn a_tolerated_test_that_fails_keeps_its_own_exit_code() {
    let g = Guard::new();
    let body = format!("{KNOWN_READ}\nexit 9");
    let out = run_fake_binary(&g, "fake_bin\tsome::test\n", &body, "some::test");
    assert_eq!(out.status.code(), Some(9), "{}", stderr(&out));
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
    assert!(
        !marker.exists(),
        "the test ran although the wrapper stopped"
    );
}

#[test]
fn the_wrapper_removes_only_the_folder_it_made() {
    let g = Guard::new();
    fs::create_dir(g.tmp.path().join("keep-me")).unwrap();
    fs::write(g.tmp.path().join("keep-me/file"), "x").unwrap();
    fs::create_dir(g.tmp.path().join("moai-test-home-lookalike")).unwrap();
    // The folder and the reference file of another wrapper process that runs at the same time.
    fs::create_dir(g.tmp.path().join("moai-test-home.1")).unwrap();
    fs::write(g.tmp.path().join("moai-test-home.1/file"), "x").unwrap();
    fs::write(g.tmp.path().join("moai-test-home.1.ref"), "").unwrap();
    let out = g.run(Some("enforce"), None, "true");
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert_eq!(
        g.entries(),
        vec![
            "keep-me".to_string(),
            "moai-test-home-lookalike".to_string(),
            "moai-test-home.1".to_string(),
            "moai-test-home.1.ref".to_string(),
        ]
    );
    assert!(g.tmp.path().join("moai-test-home.1/file").exists());
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

/// Runs `prepare` in a shell that then `exec`s the wrapper: `exec` keeps the process id, so a name
/// built from `$$` in `prepare` is the name an earlier wrapper process with a reused id left behind.
/// The shell's stdout is the process id (`prepare` prints it) and then the wrapper's.
fn run_after_prepare(g: &Guard, prepare: &str) -> Output {
    let script_text = format!("{prepare}\nexec sh \"$WRAPPER\" sh -c true");
    Command::new("sh")
        .args(["-c", &script_text])
        .env("TMPDIR", g.tmp.path())
        .env("WRAPPER", script())
        .env_remove("MOAI_HOME_GUARD")
        .env_remove("MOAI_HOME_GUARD_LOG")
        .output()
        .unwrap()
}

#[test]
fn a_folder_left_under_the_name_of_this_process_is_not_removed() {
    let g = Guard::new();
    let out = run_after_prepare(
        &g,
        r#"mkdir "$TMPDIR/moai-test-home.$$" && echo precious > "$TMPDIR/moai-test-home.$$/keep"
printf '%s' "$$""#,
    );
    let pid = stdout(&out);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let kept = g.tmp.path().join(format!("moai-test-home.{pid}/keep"));
    assert_eq!(
        fs::read_to_string(&kept).unwrap_or_default(),
        "precious\n",
        "the wrapper removed a folder it did not make: {}",
        kept.display()
    );
}

#[test]
fn a_link_left_in_the_place_of_the_reference_file_keeps_its_target() {
    let g = Guard::new();
    let out = run_after_prepare(
        &g,
        r#"echo precious > "$TMPDIR/target"
ln -s "$TMPDIR/target" "$TMPDIR/moai-test-home.$$.ref"
printf '%s' "$$""#,
    );
    let pid = stdout(&out);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert_eq!(
        fs::read_to_string(g.tmp.path().join("target")).unwrap_or_default(),
        "precious\n",
        "the wrapper wrote through a link it did not make"
    );
    assert!(
        fs::symlink_metadata(g.tmp.path().join(format!("moai-test-home.{pid}.ref"))).is_ok(),
        "the wrapper removed a link it did not make"
    );
}

/// A probe that removes the reference file the wrapper compares times with (`mktemp` made it under
/// `$TMPDIR` as `moai-test-home-ref.XXXXXX`).
const VANISH_REF: &str = r#"rm -f "$TMPDIR"/moai-test-home-ref.*"#;

#[test]
fn a_reference_file_that_vanishes_does_not_let_a_test_pass() {
    let g = Guard::new();
    let out = g.run(Some("enforce"), None, VANISH_REF);
    assert_ne!(
        out.status.code(),
        Some(0),
        "the comparison could not run and the test passed: {}",
        stderr(&out)
    );
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
    assert_eq!(
        g.entries(),
        Vec::<String>::new(),
        "the wrapper left a folder"
    );
}

#[test]
fn a_home_that_vanishes_does_not_let_a_test_pass() {
    let g = Guard::new();
    let out = g.run(Some("enforce"), None, r#"rm -rf "$HOME""#);
    assert_ne!(
        out.status.code(),
        Some(0),
        "the comparison could not run and the test passed: {}",
        stderr(&out)
    );
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

#[test]
fn a_comparison_that_cannot_run_keeps_the_test_exit_code_and_never_fails_observe() {
    let g = Guard::new();
    let failing = format!("{VANISH_REF}; exit 7");
    let out = g.run(Some("enforce"), None, &failing);
    assert_eq!(out.status.code(), Some(7), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
    let out = g.run(Some("observe"), None, VANISH_REF);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(stderr(&out).contains("HOME-GUARD:"), "{}", stderr(&out));
}

/// Every end of the wrapper that a shell can catch leaves neither the HOME folder nor the
/// reference file in its `TMPDIR`: a normal exit, a failing test and a HUP, INT or TERM sent to
/// the wrapper while the test runs (the probe signals its parent, the wrapper's shell, which runs
/// its trap once the probe is over). A SIGKILL cannot be caught and is not covered.
#[test]
fn a_wrapper_that_ends_leaves_nothing_in_its_tmpdir() {
    let cases: [(&str, i32); 5] = [
        ("exit 0", 0),
        ("exit 7", 7),
        ("kill -TERM $PPID", 143),
        ("kill -INT $PPID", 143),
        ("kill -HUP $PPID", 143),
    ];
    for (probe, code) in cases {
        let g = Guard::new();
        let out = g.run(Some("enforce"), None, probe);
        assert_eq!(out.status.code(), Some(code), "{probe}: {}", stderr(&out));
        assert_eq!(
            g.entries(),
            Vec::<String>::new(),
            "the wrapper left a name behind after `{probe}`"
        );
    }
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
#[test]
fn plant_tmpdir_is_private() {
    let t = std::env::var("TMPDIR").unwrap_or_default();
    assert!(t.ends_with("/private-tmp"), "TMPDIR is {t:?}");
}
"#;

/// The folder inside a scratch crate that its nextest run gets as `TMPDIR`, so the names the
/// wrapper makes never reach the system temporary folder.
const PRIVATE_TMP: &str = "private-tmp";

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
    nextest_with_log(dir, mode, None)
}

/// [`nextest`] with `MOAI_HOME_GUARD_LOG` set to `log` outside the scratch crate, so the test can
/// read whether the variable reaches the wrapper through nextest.
fn nextest_with_log(dir: &Path, mode: Option<&str>, log: Option<&Path>) -> (Option<i32>, String) {
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
    // The wrapper makes its HOME folder and reference file under `TMPDIR`: a private folder of
    // the scratch crate keeps them out of the system temporary folder, and the crate's `TempDir`
    // removes it when the test ends, panic or not.
    let private_tmp = dir.join(PRIVATE_TMP);
    fs::create_dir_all(&private_tmp).unwrap();
    cmd.env("TMPDIR", &private_tmp);
    if let Some(mode) = mode {
        cmd.env("MOAI_HOME_GUARD", mode);
    }
    if let Some(log) = log {
        cmd.env("MOAI_HOME_GUARD_LOG", log);
    }
    let out = cmd
        .output()
        .expect("cargo nextest must be installed for this test");
    let text = format!("{}\n{}", stdout(&out), stderr(&out));
    let left = names_with_prefix(&dir.join(PRIVATE_TMP), "moai-test-home");
    assert!(
        left.is_empty(),
        "the wrapper left {left:?} in the private TMPDIR of the nextest run"
    );
    (out.status.code(), text)
}

/// The names directly inside `folder` that start with `prefix` (none when the folder is absent).
fn names_with_prefix(folder: &Path, prefix: &str) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(folder)
        .map(|rd| {
            rd.filter_map(|e| e.ok())
                .map(|e| e.file_name().to_string_lossy().into_owned())
                .filter(|n| n.starts_with(prefix))
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    names
}

/// The first status word nextest printed for the test `name`.
fn status(text: &str, name: &str) -> Option<String> {
    text.lines().find_map(|line| {
        let t = line.trim();
        let word = t.split_whitespace().next()?;
        let known = [
            "PASS",
            "FAIL",
            "SIGABRT",
            "SIGSEGV",
            "TIMEOUT",
            "ERROR",
            "EXEC-FAIL",
        ];
        (known.contains(&word) && t.ends_with(&format!(" {name}"))).then(|| word.to_string())
    })
}

fn require_nextest() {
    let ok = Command::new("cargo")
        .args(["nextest", "--version"])
        .output()
        .map(|o| o.status.success())
        .unwrap_or(false);
    assert!(
        ok,
        "cargo-nextest is not installed, so the wrapper cannot be exercised here"
    );
}

#[test]
fn nextest_applies_the_repository_configuration_to_a_scratch_crate() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Repository);
    let (code, text) = nextest(dir.path(), Some("enforce"));
    let st = |n: &str| status(&text, n);
    assert_eq!(st("plant_reads_decoy").as_deref(), Some("FAIL"), "{text}");
    assert_eq!(st("plant_writes_home").as_deref(), Some("FAIL"), "{text}");
    assert_eq!(
        st("plant_reads_then_fails").as_deref(),
        Some("FAIL"),
        "{text}"
    );
    assert!(
        text.contains("plant-own-failure"),
        "the test's own failure is lost: {text}"
    );
    assert_eq!(st("plant_idle").as_deref(), Some("PASS"), "{text}");
    assert_eq!(
        st("plant_home_is_guarded").as_deref(),
        Some("PASS"),
        "{text}"
    );
    assert_eq!(
        st("plant_tmpdir_is_private").as_deref(),
        Some("PASS"),
        "{text}"
    );
    assert!(text.contains("HOME-GUARD:"), "{text}");
    assert_eq!(code, Some(100), "{text}");
}

#[test]
fn the_default_mode_of_the_repository_configuration_fails_a_reading_test() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Repository);
    let (_, text) = nextest(dir.path(), None);
    assert_eq!(
        status(&text, "plant_reads_decoy").as_deref(),
        Some("FAIL"),
        "{text}"
    );
    assert!(text.contains("HOME-GUARD:"), "{text}");
}

#[test]
fn the_observe_mode_of_the_repository_configuration_does_not_fail_a_reading_test() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Repository);
    let (_, text) = nextest(dir.path(), Some("observe"));
    assert_eq!(
        status(&text, "plant_reads_decoy").as_deref(),
        Some("PASS"),
        "{text}"
    );
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
            return Err(format!(
                "entry is neither binary(..) nor test(=..): {predicate}"
            ));
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
    assert!(
        cfg.contains("nextest-version"),
        "no nextest-version is named"
    );
    assert!(
        cfg.contains("run-wrapper = 'home-guard'"),
        "no rule runs the wrapper"
    );
    assert!(
        cfg.contains("cfg(unix)"),
        "the rule is not limited to unix hosts"
    );
    assert!(
        cfg.contains("tools/test_home_guard.sh"),
        "the wrapper script is not named"
    );
}

// ---- the baseline file --------------------------------------------------------------------------

const BASELINE: &str = "tests/fixtures/test_home_guard/baseline.tsv";

/// The most lines the baseline may hold. The observation of SPEC-20261123-test-home-isolation (D-5)
/// found 389 tests, and the file only shrinks: remove lines, then lower this number to the new
/// count. A test that reads the decoy and is not listed fails instead of being added.
const BASELINE_CAP: usize = 361;

/// Every `binary<TAB>test` line of a baseline: sorted without repeats, within `cap`, and naming a
/// test that exists. The library tests use the binary name `moai_cli`; the name is a module path
/// that leads to one source file. An integration test is found by its last name segment in
/// `tests/<binary>.rs`. Returns the number of lines.
fn check_baseline(text: &str, root: &Path, cap: usize) -> Result<usize, String> {
    let lines: Vec<&str> = text
        .lines()
        .filter(|l| !l.trim().is_empty() && !l.starts_with('#'))
        .collect();
    if lines.len() > cap {
        return Err(format!(
            "the baseline has {} lines, more than its cap {cap}",
            lines.len()
        ));
    }
    for pair in lines.windows(2) {
        if pair[0] >= pair[1] {
            return Err(format!(
                "the baseline is not sorted without repeats at: {}",
                pair[1]
            ));
        }
    }
    let mut sources: std::collections::BTreeMap<String, String> = Default::default();
    for line in &lines {
        let (bin, name) = line
            .split_once('\t')
            .filter(|(b, n)| !b.is_empty() && !n.is_empty() && !n.contains('\t'))
            .ok_or_else(|| format!("malformed baseline line: {line}"))?;
        let exists = if bin == "moai_cli" {
            function_exists(root, name)
        } else {
            let short = name.rsplit("::").next().unwrap_or(name);
            let src = sources.entry(bin.to_string()).or_insert_with(|| {
                fs::read_to_string(root.join("tests").join(format!("{bin}.rs"))).unwrap_or_default()
            });
            src.contains(&format!("fn {short}("))
        };
        if !exists {
            return Err(format!("the baseline line names no existing test: {line}"));
        }
    }
    Ok(lines.len())
}

#[test]
fn the_repository_baseline_stays_within_its_cap_and_names_real_tests() {
    let text = fs::read_to_string(Path::new(ROOT).join(BASELINE)).unwrap();
    let n = check_baseline(&text, Path::new(ROOT), BASELINE_CAP).unwrap();
    assert!(
        n > 0,
        "an empty baseline means the cap and this file should be removed"
    );
}

#[test]
fn a_baseline_that_grew_is_refused() {
    let text = "test_home_guard\ta_baseline_that_grew_is_refused\n";
    let err = check_baseline(text, Path::new(ROOT), 0).unwrap_err();
    assert!(err.contains("more than its cap"), "{err}");
}

#[test]
fn an_unsorted_or_repeating_baseline_is_refused() {
    let a = "test_home_guard\tthe_default_mode_is_enforce";
    let b = "test_home_guard\ta_baseline_that_grew_is_refused";
    let err = check_baseline(&format!("{a}\n{b}\n"), Path::new(ROOT), 10).unwrap_err();
    assert!(err.contains("not sorted"), "{err}");
    let err = check_baseline(&format!("{a}\n{a}\n"), Path::new(ROOT), 10).unwrap_err();
    assert!(err.contains("not sorted"), "{err}");
}

#[test]
fn a_baseline_line_that_names_nothing_is_refused() {
    let root = Path::new(ROOT);
    for text in [
        "test_home_guard\tno_such_test_anywhere\n",
        "no_such_binary\tsome_test\n",
        "moai_cli\tno_such_module::tests::some_test\n",
    ] {
        let err = check_baseline(text, root, 10).unwrap_err();
        assert!(err.contains("names no existing test"), "{text}: {err}");
    }
    let err = check_baseline("no tab on this line\n", root, 10).unwrap_err();
    assert!(err.contains("malformed"), "{err}");
}

// ---- the detail lines (SPEC-20261124-home-guard-flake) ---------------------------------------------
//
// A `touched` verdict leaves evidence in the log file and changes nothing else: the verdict, the exit
// code and the `HOME-GUARD:` line stay what the tests above pin.

const DETAIL: &str = "HOME-GUARD-DETAIL:";

fn detail_lines(log: &Path) -> Vec<String> {
    fs::read_to_string(log)
        .unwrap_or_default()
        .lines()
        .filter(|l| l.starts_with(DETAIL))
        .map(str::to_string)
        .collect()
}

/// The value of `key=` in one detail line, up to the next space.
fn field<'a>(line: &'a str, key: &str) -> Option<&'a str> {
    let k = format!("{key}=");
    line.split_whitespace()
        .find_map(|w| w.strip_prefix(k.as_str()))
}

fn digits(s: &str) -> bool {
    !s.is_empty() && s.bytes().all(|c| c.is_ascii_digit())
}

/// Seconds with a fraction: `1795505797.048996`.
fn has_fraction(t: &str) -> bool {
    t.split_once('.')
        .is_some_and(|(a, b)| digits(a) && digits(b))
}

fn seconds(t: &str) -> f64 {
    t.parse().unwrap_or_else(|_| panic!("not a time: {t:?}"))
}

fn accessed(p: &Path) -> std::time::SystemTime {
    fs::metadata(p).unwrap().accessed().unwrap()
}

/// The wrapper run with a folder of stand-in commands in front of `PATH`.
fn run_with_fake_path(g: &Guard, fake: &Path, mode: &str, log: &Path, probe: &str) -> Output {
    let path = format!(
        "{}:{}",
        fake.display(),
        std::env::var("PATH").unwrap_or_default()
    );
    Command::new("sh")
        .arg(script())
        .args(["sh", "-c", probe])
        .env("TMPDIR", g.tmp.path())
        .env("PATH", path)
        .env("MOAI_HOME_GUARD", mode)
        .env("MOAI_HOME_GUARD_LOG", log)
        .output()
        .unwrap()
}

/// A stand-in command that always fails, the way a host without it would.
fn failing_command(dir: &Path, name: &str) {
    use std::os::unix::fs::PermissionsExt;
    fs::write(dir.join(name), "#!/bin/sh\nexit 1\n").unwrap();
    fs::set_permissions(dir.join(name), fs::Permissions::from_mode(0o755)).unwrap();
}

/// AC-1: one detail line per hit entry, with times in seconds with a fraction, and none when the
/// test touched nothing.
#[test]
fn a_touched_verdict_writes_one_detail_line_per_hit_entry() {
    let g = Guard::new();
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let out = g.run(Some("observe"), Some(&log), DECOY_READ);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let printed = stderr(&out);
    let listed = printed
        .lines()
        .find(|l| l.starts_with("HOME-GUARD: touched its HOME"))
        .and_then(|l| l.split_once(" -> "))
        .map(|(_, hits)| hits.split_whitespace().count())
        .unwrap_or_else(|| panic!("no touched line: {printed}"));
    let hits: Vec<String> = detail_lines(&log)
        .into_iter()
        .filter(|l| field(l, "kind") == Some("hit"))
        .collect();
    assert_eq!(hits.len(), listed, "{hits:#?}\n{printed}");
    for l in &hits {
        assert!(l.starts_with(&format!("{DETAIL} kind=hit ")), "{l}");
        for key in ["atime", "mtime", "ref", "end"] {
            let t = field(l, key).unwrap_or_else(|| panic!("no {key} in {l}"));
            assert!(has_fraction(t), "{key} has no fraction in {l}");
        }
        assert!(
            seconds(field(l, "ref").unwrap()) <= seconds(field(l, "end").unwrap()),
            "{l}"
        );
    }
    let read = hits
        .iter()
        .find(|l| field(l, "path") == Some(".moai/user-assets.json"))
        .unwrap_or_else(|| panic!("the decoy read is not named: {hits:#?}"));
    // The read came after the reference file was made and before the comparison.
    let (r, a, e) = (
        seconds(field(read, "ref").unwrap()),
        seconds(field(read, "atime").unwrap()),
        seconds(field(read, "end").unwrap()),
    );
    assert!(r < a && a <= e, "ref {r} < atime {a} <= end {e}");

    let quiet = logs.path().join("quiet.log");
    let out = g.run(Some("enforce"), Some(&quiet), r#"printf '%s\n' "$HOME""#);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    assert!(detail_lines(&quiet).is_empty());
}

/// AC-2: a guard folder of another test is read with `stat` only, so its access times stay put.
#[test]
fn another_guard_folder_is_named_by_its_access_time_and_never_opened() {
    let g = Guard::new();
    let peer = g.tmp.path().join("moai-test-home.PEER01");
    fs::create_dir_all(peer.join(".moai")).unwrap();
    let file = peer.join(".moai/user-assets.json");
    fs::write(&file, "").unwrap();
    let old = Command::new("touch")
        .args(["-a", "-t", "200001010000"])
        .args([&file, &peer.join(".moai"), &peer])
        .status()
        .unwrap();
    assert!(old.success());
    let before = [
        accessed(&file),
        accessed(&peer.join(".moai")),
        accessed(&peer),
    ];
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let out = g.run(Some("observe"), Some(&log), DECOY_READ);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let peers: Vec<String> = detail_lines(&log)
        .into_iter()
        .filter(|l| field(l, "kind") == Some("peer"))
        .collect();
    assert_eq!(peers.len(), 1, "{peers:#?}");
    assert_eq!(field(&peers[0], "folder"), Some("moai-test-home.PEER01"));
    let secs = before[0]
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs()
        .to_string();
    let t = field(&peers[0], "atime").unwrap();
    assert!(
        has_fraction(t) && t.starts_with(&format!("{secs}.")),
        "{} vs {secs}",
        peers[0]
    );
    let after = [
        accessed(&file),
        accessed(&peer.join(".moai")),
        accessed(&peer),
    ];
    assert_eq!(
        before, after,
        "the wrapper opened or listed the other guard folder"
    );
}

/// AC-6: nothing in a detail line says where the user's folders are or what the test was asked.
#[test]
fn the_detail_lines_hold_relative_paths_and_no_argument() {
    let g = Guard::new();
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let probe = format!("{DECOY_READ}; : ARGUMENT-MARKER-9f3");
    let out = g.run(Some("observe"), Some(&log), &probe);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let lines = detail_lines(&log);
    assert!(!lines.is_empty());
    let mut banned = vec![
        g.tmp.path().to_string_lossy().into_owned(),
        "ARGUMENT-MARKER".to_string(),
        ROOT.to_string(),
    ];
    if let Ok(h) = std::env::var("HOME") {
        if h.len() > 1 {
            banned.push(h);
        }
    }
    for l in &lines {
        for b in &banned {
            assert!(!l.contains(b.as_str()), "{b:?} is in {l}");
        }
    }
    let paths: Vec<&str> = lines
        .iter()
        .filter(|l| field(l, "kind") == Some("hit"))
        .filter_map(|l| field(l, "path"))
        .collect();
    for want in [".", ".moai", ".moai/user-assets.json"] {
        assert!(paths.contains(&want), "{want} is missing in {paths:?}");
    }
    assert!(paths.iter().all(|p| !p.starts_with('/')), "{paths:?}");
}

/// AC-7: a running `moai*` process is named by id and file name; its arguments and folder are not
/// kept; a process table that cannot be read leaves one line and the same verdict.
#[test]
fn a_moai_process_is_named_by_id_and_file_name_only() {
    let g = Guard::new();
    let bin = TempDir::new().unwrap();
    // A link, not a copy: a copied system binary loses its signature and macOS kills it at start.
    std::os::unix::fs::symlink("/bin/sleep", bin.path().join("moai-probe")).unwrap();
    let mut child = Command::new(bin.path().join("moai-probe"))
        .arg("31.4159")
        .spawn()
        .unwrap();
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let out = g.run(Some("observe"), Some(&log), DECOY_READ);
    let _ = child.kill();
    let _ = child.wait();
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let procs: Vec<String> = detail_lines(&log)
        .into_iter()
        .filter(|l| field(l, "kind") == Some("proc"))
        .collect();
    let mine = procs
        .iter()
        .find(|l| field(l, "pid") == Some(child.id().to_string().as_str()))
        .unwrap_or_else(|| panic!("pid {} is not named: {procs:#?}", child.id()));
    assert_eq!(field(mine, "name"), Some("moai-probe"), "{mine}");
    for l in &procs {
        assert!(!l.contains("31.4159"), "an argument is kept: {l}");
        assert!(
            !l.contains(bin.path().to_str().unwrap()),
            "a folder is kept: {l}"
        );
    }
}

#[test]
fn an_unreadable_process_table_leaves_one_line_and_the_same_verdict() {
    let g = Guard::new();
    let fake = TempDir::new().unwrap();
    failing_command(fake.path(), "ps");
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let out = run_with_fake_path(&g, fake.path(), "enforce", &log, DECOY_READ);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    let lines = detail_lines(&log);
    let unavailable = lines
        .iter()
        .filter(|l| l.contains("procs=unavailable"))
        .count();
    assert_eq!(unavailable, 1, "{lines:#?}");
    assert!(
        lines.iter().all(|l| field(l, "kind") != Some("proc")),
        "{lines:#?}"
    );
}

/// AC-5: without `perl` the times are whole seconds and the verdict is the same.
#[test]
fn without_perl_the_times_are_whole_seconds_and_the_verdict_is_the_same() {
    let g = Guard::new();
    let fake = TempDir::new().unwrap();
    failing_command(fake.path(), "perl");
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let out = run_with_fake_path(&g, fake.path(), "enforce", &log, DECOY_READ);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
    let hits: Vec<String> = detail_lines(&log)
        .into_iter()
        .filter(|l| field(l, "kind") == Some("hit"))
        .collect();
    assert!(!hits.is_empty());
    for l in &hits {
        for key in ["atime", "mtime", "ref", "end"] {
            let t = field(l, key).unwrap_or_else(|| panic!("no {key} in {l}"));
            assert!(digits(t), "{key} is not whole seconds in {l}");
        }
    }
}

/// AC-4: a detail line that cannot be written leaves the exit code and the printed line as they
/// were, and no shell error of its own.
#[test]
fn a_detail_that_cannot_be_written_changes_nothing() {
    let g = Guard::new();
    let logs = TempDir::new().unwrap();
    let bad = logs.path().join("no-such-folder").join("guard.log");
    let out = g.run(Some("observe"), Some(&bad), DECOY_READ);
    assert_eq!(out.status.code(), Some(0), "{}", stderr(&out));
    let printed = stderr(&out);
    assert!(
        printed.contains("HOME-GUARD: touched its HOME"),
        "{printed}"
    );
    assert!(!printed.contains("No such file"), "{printed}");
    let out = g.run(Some("enforce"), Some(&bad), DECOY_READ);
    assert_eq!(out.status.code(), Some(1), "{}", stderr(&out));
}

/// AC-8: `MOAI_HOME_GUARD_LOG` set outside the scratch crate reaches the wrapper through nextest.
#[test]
fn the_log_variable_reaches_the_wrapper_through_nextest() {
    require_nextest();
    let dir = scratch_crate(Wrapper::Repository);
    let logs = TempDir::new().unwrap();
    let log = logs.path().join("guard.log");
    let (_, text) = nextest_with_log(dir.path(), Some("observe"), Some(&log));
    assert_eq!(
        status(&text, "plant_reads_decoy").as_deref(),
        Some("PASS"),
        "{text}"
    );
    let lines = detail_lines(&log);
    assert!(
        lines.iter().any(|l| field(l, "kind") == Some("hit")
            && field(l, "path") == Some(".moai/user-assets.json")),
        "the log has no detail line: {}",
        fs::read_to_string(&log).unwrap_or_default()
    );
}
