//! SPEC-20261124-launcher-retire-role-flags D-4: the launcher no longer numbers sessions. A
//! session is named by `--name` only, and a launch without `--name` takes no name and no lock.
//! A named launch holds `<sessions dir>/<name>.live` for as long as its host runs (the OS drops
//! the lock on any exit; SPEC-20261120-launcher-liveness REQ-4). A launch refused before the
//! host starts (bad user settings, no token) starts no host and takes no lock, and a live path
//! that is a link or not a regular file is refused. Real launches run a stub `claude`
//! executable from a temporary PATH in a temporary git project; no real host starts.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::os::unix::fs::PermissionsExt;
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
    time::{Duration, Instant},
};

struct Fixture {
    project: tempfile::TempDir,
    bin: tempfile::TempDir,
    home: tempfile::TempDir,
}

fn git(dir: &Path, args: &[&str]) {
    let ok = common::git::command()
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(args)
        .current_dir(dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .unwrap()
        .success();
    assert!(ok, "git {args:?}");
}

/// A host stub that leaves a mark and exits at once.
const STUB: &str = "#!/bin/sh
printf started > \"$MOAI_TEST_HOST_MARKER\"
exit 0
";

fn fixture() -> Fixture {
    let project = tempfile::tempdir().unwrap();
    git(project.path(), &["init", "-q"]);
    // An initialized project: with settings present no launch runs the first-launch
    // initialization, so these tests measure the locks only (tests/launcher_auto_init.rs).
    let settings = moai_cli::state_root::path(project.path(), "project.yaml");
    fs::create_dir_all(settings.parent().unwrap()).unwrap();
    fs::write(settings, "{}\n").unwrap();
    let bin = tempfile::tempdir().unwrap();
    let path = bin.path().join("claude");
    fs::write(&path, STUB).unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    Fixture {
        project,
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

fn command(f: &Fixture, args: &[&str]) -> Command {
    let mut c = Command::new(crate::common::MOAI_BIN);
    c.args(args)
        .current_dir(f.project.path())
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", f.home.path().join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL");
    c
}

fn run(f: &Fixture, args: &[&str]) -> Output {
    command(f, args).output().unwrap()
}

/// The session name the launcher reported, after asserting it succeeded.
fn name(out: &Output) -> String {
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["name"].as_str().unwrap_or("").to_string()
}

fn sessions(f: &Fixture) -> PathBuf {
    moai_cli::clone_root::path(&f.project.path().join(".git")).join("sessions")
}

fn live_path(f: &Fixture, name: &str) -> PathBuf {
    sessions(f).join(format!("{name}.live"))
}

/// The `.live` files in the sessions folder, sorted; empty when there is no folder.
fn lives(f: &Fixture) -> Vec<String> {
    let mut names: Vec<String> = fs::read_dir(sessions(f))
        .map(|d| {
            d.map(|e| e.unwrap().file_name().to_string_lossy().to_string())
                .filter(|n| n.ends_with(".live"))
                .collect()
        })
        .unwrap_or_default();
    names.sort();
    names
}

fn host_started(f: &Fixture) -> bool {
    f.home.path().join("host-started").exists()
}

/// A refused launch: exit 1, nothing on stdout, a message on stderr.
fn refused(out: &Output) -> String {
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert_eq!(out.status.code(), Some(1), "{stderr}");
    assert!(
        out.stdout.is_empty(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(!stderr.is_empty());
    stderr
}

/// A named launch takes its name, holds its `<name>.live` lock while the host runs (the lock
/// file stays behind, unlocked), and a launch without `--name` takes no name and no lock.
#[test]
fn a_named_launch_leaves_its_live_file_and_a_plain_launch_none() {
    let f = fixture();
    assert_eq!(name(&run(&f, &["cc", "--name", "x"])), "x");
    assert_eq!(name(&run(&f, &["glm", "--name", "y"])), "y");
    assert_eq!(name(&run(&f, &["glm"])), "");
    assert_eq!(name(&run(&f, &["cc"])), "");
    assert_eq!(lives(&f), vec!["x.live", "y.live"]);
    assert!(host_started(&f));
}

/// A `--dry-run` takes no lock and creates nothing, whatever the name.
#[test]
fn a_dry_run_takes_no_lock() {
    let f = fixture();
    assert_eq!(name(&run(&f, &["cc", "--name", "x", "--dry-run"])), "x");
    assert_eq!(lives(&f), Vec::<String>::new());
    assert!(!sessions(&f).exists());
    assert!(!host_started(&f));
}

/// The decided argv of a real launch with a name: pinned byte for byte, against the stub host.
#[test]
fn a_named_launch_runs_the_host_with_the_decided_argv() {
    let f = fixture();
    let out = run(&f, &["cc", "--name", "x"]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let argv: Vec<&str> = v["argv"]
        .as_array()
        .unwrap_or_else(|| panic!("no argv in {v}"))
        .iter()
        .map(|a| a.as_str().unwrap())
        .collect();
    assert_eq!(
        argv,
        [
            "claude",
            "--name",
            "x",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--permission-mode",
            "auto"
        ]
    );
}

/// A launch that is refused before the host starts takes no lock and runs no host: `glm` with
/// no token, and a model with a space in it.
#[test]
fn a_launch_refused_before_the_host_starts_takes_no_lock() {
    let f = fixture();
    let out = command(&f, &["glm", "--name", "x"])
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1));
    assert_ne!(
        run(&f, &["cc", "--name", "x", "--model", "two words"])
            .status
            .code(),
        Some(0)
    );
    assert_eq!(lives(&f), Vec::<String>::new());
    assert!(!host_started(&f));
}

/// A `.live` path that is a symlink is refused and nothing is written or followed.
#[test]
fn a_live_path_that_is_a_symlink_is_refused_and_nothing_is_written() {
    let f = fixture();
    let real = f.project.path().join("elsewhere");
    fs::write(&real, "3\n").unwrap();
    fs::create_dir_all(sessions(&f)).unwrap();
    std::os::unix::fs::symlink(&real, live_path(&f, "x")).unwrap();
    let err = refused(&run(&f, &["cc", "--name", "x"]));
    assert!(err.contains("symlink"), "{err}");
    assert_eq!(fs::read_to_string(&real).unwrap(), "3\n");
    assert!(!host_started(&f));
    // A dangling link is a link too.
    let g = fixture();
    fs::create_dir_all(sessions(&g)).unwrap();
    std::os::unix::fs::symlink(g.project.path().join("missing"), live_path(&g, "x")).unwrap();
    refused(&run(&g, &["cc", "--name", "x"]));
}

/// A `.live` path that is a directory or a FIFO is refused, and a FIFO does not block the launch.
#[test]
fn a_live_path_that_is_not_a_regular_file_is_refused_without_blocking() {
    let f = fixture();
    fs::create_dir_all(live_path(&f, "x")).unwrap();
    let err = refused(&run(&f, &["cc", "--name", "x"]));
    assert!(err.contains("regular file"), "{err}");
    let g = fixture();
    fs::create_dir_all(sessions(&g)).unwrap();
    assert!(Command::new("mkfifo")
        .arg(live_path(&g, "x"))
        .status()
        .unwrap()
        .success());
    let mut child = command(&g, &["cc", "--name", "x"])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let start = Instant::now();
    while child.try_wait().unwrap().is_none() {
        // #2511: a hang guard, so the bound is generous.
        if start.elapsed() > Duration::from_secs(120) {
            let _ = child.kill();
            let _ = child.wait();
            panic!("launcher blocked on a FIFO");
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    refused(&child.wait_with_output().unwrap());
    assert!(!host_started(&g));
}

/// A name that is already running is refused at once, with no wait and no second host.
#[test]
fn a_name_that_is_already_running_is_refused_at_once() {
    let f = fixture();
    fs::create_dir_all(sessions(&f)).unwrap();
    let held = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(live_path(&f, "x"))
        .unwrap();
    held.lock().unwrap();
    let err = refused(&run(&f, &["cc", "--name", "x"]));
    assert!(err.contains("already running"), "{err}");
    assert!(!host_started(&f));
    drop(held);
    assert_eq!(name(&run(&f, &["cc", "--name", "x"])), "x");
}

fn launcher_config(f: &Fixture) -> PathBuf {
    moai_cli::home_root::path(f.home.path()).join("config/launcher.yaml")
}

/// A launch refused for its user settings starts no host, prints nothing on stdout, shows no
/// secret and takes no lock.
fn assert_config_refusal_before_side_effects(f: &Fixture) {
    for host in ["cc", "glm"] {
        for extra in [&[][..], &["--permissions"][..], &["--name", "x"][..]] {
            let mut args = vec![host];
            args.extend_from_slice(extra);
            let out = run(f, &args);
            assert_eq!(out.status.code(), Some(1), "{args:?}");
            assert!(out.stdout.is_empty(), "{args:?}");
            assert!(!String::from_utf8_lossy(&out.stderr).contains("SECRET"));
            assert!(!host_started(f), "{args:?}");
            assert_eq!(lives(f), Vec::<String>::new(), "{args:?}");
        }
    }
}

#[test]
fn invalid_launcher_config_never_executes_host_or_takes_a_lock() {
    for bad in [
        r#"{"launcher":{"role_access":"SECRET-invalid-access"}}"#,
        r#"{"launcher":{"role_access":null}}"#,
        r#"{"launcher":{"role_access":false}}"#,
        r#"{"launcher":{"role_access":42}}"#,
        r#"{"launcher":{"role_access":[]}}"#,
        r#"{"launcher":{"role_access":{}}}"#,
        r#"{"launcher":null}"#,
        r#"{"launcher":[]}"#,
        r#"{"launcher":"SECRET-invalid-structure"}"#,
        "null",
        "[]",
        "42",
        "{ SECRET-malformed",
    ] {
        let f = fixture();
        fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
        fs::write(launcher_config(&f), bad).unwrap();
        assert_config_refusal_before_side_effects(&f);
    }
}

#[test]
fn nonregular_launcher_config_never_executes_host_or_takes_a_lock() {
    let f = fixture();
    fs::create_dir_all(launcher_config(&f)).unwrap();
    assert_config_refusal_before_side_effects(&f);
}

#[test]
fn unreadable_launcher_config_never_executes_host_or_takes_a_lock() {
    let f = fixture();
    fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
    fs::write(
        launcher_config(&f),
        r#"{"launcher":{"role_access":"prompt"}}"#,
    )
    .unwrap();
    fs::set_permissions(launcher_config(&f), fs::Permissions::from_mode(0o000)).unwrap();
    // Elevated CI users can read mode-000 files; exercise refusal only when unreadable.
    let unreadable = fs::File::open(launcher_config(&f)).is_err();
    if unreadable {
        assert_config_refusal_before_side_effects(&f);
    }
    fs::set_permissions(launcher_config(&f), fs::Permissions::from_mode(0o600)).unwrap();
}

#[test]
fn symlink_launcher_config_is_refused_without_following_it() {
    let f = fixture();
    fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
    let target = f.home.path().join("synthetic-config");
    let contents = r#"{"launcher":{"role_access":"full"}}"#;
    fs::write(&target, contents).unwrap();
    std::os::unix::fs::symlink(&target, launcher_config(&f)).unwrap();
    assert_config_refusal_before_side_effects(&f);
    assert_eq!(fs::read_to_string(target).unwrap(), contents);
}

/// Valid settings start the stub host on both hosts.
#[test]
fn valid_launcher_settings_execute_the_stub_hosts() {
    for setting in ["full", "prompt"] {
        let f = fixture();
        fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
        fs::write(
            launcher_config(&f),
            format!(r#"{{"launcher":{{"role_access":"{setting}"}}}}"#),
        )
        .unwrap();
        for host in ["cc", "glm"] {
            assert_eq!(name(&run(&f, &[host, "--name", "x"])), "x");
            assert_eq!(
                fs::read(f.home.path().join("host-started")).unwrap(),
                b"started"
            );
        }
    }
}
