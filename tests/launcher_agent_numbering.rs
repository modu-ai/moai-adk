//! #2359 (user decision 2026-11-10, comment 6760976883; resolves #2361): `-a` without `--name`
//! is named `agent-N`. N is counted per project (one clone, all its worktrees), shared across
//! `cc`, `glm` and `codex`, only increases and is never handed out twice, also under concurrent
//! launches. `--dry-run`, `--name` and `-l` take no number, and a store that cannot be read or
//! written is an error, never a silent `agent-1`. Real launches run stub `claude` / `codex`
//! executables from a temporary PATH in a temporary git project; no real host starts.
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

fn fixture() -> Fixture {
    let project = tempfile::tempdir().unwrap();
    git(project.path(), &["init", "-q"]);
    let bin = tempfile::tempdir().unwrap();
    for host in ["claude", "codex"] {
        let path = bin.path().join(host);
        fs::write(
            &path,
            "#!/bin/sh\nprintf started > \"$MOAI_TEST_HOST_MARKER\"\nexit 0\n",
        )
        .unwrap();
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    Fixture {
        project,
        bin,
        home: tempfile::tempdir().unwrap(),
    }
}

fn command(f: &Fixture, dir: &Path, args: &[&str]) -> Command {
    let mut c = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    c.args(args)
        .current_dir(dir)
        .env("PATH", format!("{}:/usr/bin:/bin", f.bin.path().display()))
        .env("HOME", f.home.path())
        .env("USERPROFILE", f.home.path())
        .env("MOAI_TEST_HOST_MARKER", f.home.path().join("host-started"))
        .env("ANTHROPIC_AUTH_TOKEN", "moai-test-dummy-not-real")
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR);
    c
}

fn run_in(f: &Fixture, dir: &Path, args: &[&str]) -> Output {
    command(f, dir, args).output().unwrap()
}
fn run(f: &Fixture, args: &[&str]) -> Output {
    run_in(f, f.project.path(), args)
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

fn store(f: &Fixture) -> PathBuf {
    f.project.path().join(".git/.moai-cli/sessions/agent-seq")
}
fn stored(f: &Fixture) -> Option<String> {
    fs::read_to_string(store(f)).ok()
}
fn seed(f: &Fixture, text: &str) {
    fs::create_dir_all(store(f).parent().unwrap()).unwrap();
    fs::write(store(f), text).unwrap();
}

/// A refused launch: exit 1, nothing on stdout, a message that names the way out.
fn refused(out: &Output) -> String {
    let stderr = String::from_utf8_lossy(&out.stderr).to_string();
    assert_eq!(out.status.code(), Some(1), "{stderr}");
    assert!(
        out.stdout.is_empty(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(stderr.contains("pass --name"), "{stderr}");
    assert!(!stderr.contains("agent-1"), "no silent fallback: {stderr}");
    stderr
}

#[test]
fn numbers_count_up_across_hosts_and_live_in_the_clone_common_dir() {
    let f = fixture();
    let names: Vec<String> = [
        &["cc", "-a"][..],
        &["glm", "-a"][..],
        &["codex", "-a"][..],
        &["-a", "codex"][..],
        &["cc", "-a"][..],
    ]
    .iter()
    .map(|args| name(&run(&f, args)))
    .collect();
    assert_eq!(
        names,
        ["agent-1", "agent-2", "agent-3", "agent-4", "agent-5"]
    );
    assert_eq!(stored(&f).as_deref(), Some("5\n"));
    // The counter sits under `.git`, so git never lists it.
    let status = common::git::command()
        .args(["status", "--porcelain", "--ignored"])
        .current_dir(f.project.path())
        .output()
        .unwrap();
    assert!(
        status.stdout.is_empty(),
        "{}",
        String::from_utf8_lossy(&status.stdout)
    );
    // codex carries the name in its first prompt.
    let v: Value = serde_json::from_slice(&run(&f, &["codex", "-a"]).stdout).unwrap();
    assert_eq!(
        v["argv"].as_array().unwrap().last().unwrap(),
        "MoAI session name: agent-6"
    );
}

#[test]
fn dry_run_previews_and_takes_no_number() {
    let f = fixture();
    for _ in 0..3 {
        assert_eq!(name(&run(&f, &["cc", "-a", "--dry-run"])), "agent-1");
    }
    assert_eq!(stored(&f), None, "a dry run creates no store");
    assert_eq!(name(&run(&f, &["glm", "-a"])), "agent-1");
    let out = run(&f, &["codex", "-a", "--dry-run"]);
    assert_eq!(name(&out), "agent-2");
    assert_eq!(stored(&f).as_deref(), Some("1\n"));
}

#[test]
fn name_leader_and_roleless_take_no_number() {
    let f = fixture();
    assert_eq!(name(&run(&f, &["cc", "-a", "--name", "x"])), "x");
    assert_eq!(name(&run(&f, &["codex", "-l"])), "leader");
    assert_eq!(name(&run(&f, &["-l"])), "leader");
    assert_eq!(name(&run(&f, &["glm"])), "");
    assert_eq!(stored(&f), None);
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
}

#[test]
fn a_launch_refused_before_the_host_starts_takes_no_number() {
    let f = fixture();
    // glm without any token is blocked; an invalid model is refused.
    let out = command(&f, f.project.path(), &["glm", "-a"])
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .output()
        .unwrap();
    assert_eq!(out.status.code(), Some(1));
    assert_ne!(
        run(&f, &["cc", "-a", "--model", "two words"]).status.code(),
        Some(0)
    );
    assert_eq!(stored(&f), None);
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
}

#[test]
fn concurrent_launches_never_share_a_number() {
    const N: usize = 8;
    let f = fixture();
    // Each launch holds the lock 150 ms between read and write, so without the lock every
    // launch would read the same last number.
    let children: Vec<_> = (0..N)
        .map(|i| {
            let host = ["cc", "glm", "codex"][i % 3];
            command(&f, f.project.path(), &[host, "-a"])
                .env(moai_cli::session::AGENT_SEQ_TEST_HOLD_VAR, "150")
                .stdout(Stdio::piped())
                .stderr(Stdio::piped())
                .spawn()
                .unwrap()
        })
        .collect();
    let mut names: Vec<String> = children
        .into_iter()
        .map(|c| name(&c.wait_with_output().unwrap()))
        .collect();
    names.sort_by_key(|n| n.trim_start_matches("agent-").parse::<u64>().unwrap());
    let want: Vec<String> = (1..=N).map(|n| format!("agent-{n}")).collect();
    assert_eq!(names, want);
    assert_eq!(stored(&f), Some(format!("{N}\n")));
}

#[test]
fn the_counter_continues_from_the_stored_number_n_to_n_plus_1() {
    for (seed_text, want, after) in [
        ("41\n", "agent-42", "42\n"),
        ("41", "agent-42", "42\n"),
        ("0\n", "agent-1", "1\n"),
        (
            "18446744073709551614\n",
            "agent-18446744073709551615",
            "18446744073709551615\n",
        ),
    ] {
        let f = fixture();
        seed(&f, seed_text);
        assert_eq!(
            name(&run(&f, &["cc", "-a", "--dry-run"])),
            want,
            "{seed_text:?}"
        );
        assert_eq!(name(&run(&f, &["cc", "-a"])), want, "{seed_text:?}");
        assert_eq!(stored(&f).as_deref(), Some(after), "{seed_text:?}");
    }
}

#[test]
fn a_store_that_is_not_one_number_is_an_error_and_stays_as_it_was() {
    for bad in [
        "",
        "\n",
        "4x\n",
        " 4\n",
        "4 \n",
        "-1\n",
        "4\n\n",
        "+4\n",
        "18446744073709551616\n",
    ] {
        let f = fixture();
        seed(&f, bad);
        for args in [&["cc", "-a"][..], &["codex", "-a", "--dry-run"][..]] {
            let err = refused(&run(&f, args));
            assert!(err.contains("agent-seq"), "{bad:?}: {err}");
        }
        assert_eq!(stored(&f).as_deref(), Some(bad));
    }
    // The last number: nothing comes after it. 21 bytes, the largest store that is read.
    let f = fixture();
    seed(&f, "18446744073709551615\n");
    let err = refused(&run(&f, &["cc", "-a"]));
    assert!(err.contains("exhausted"), "{err}");
    assert_eq!(stored(&f).as_deref(), Some("18446744073709551615\n"));
}

#[test]
fn the_store_size_limit_is_21_bytes() {
    // 21 bytes (20 digits + newline) is read, leading zeros and all.
    let f = fixture();
    seed(&f, "00000000000000000041\n");
    assert_eq!("00000000000000000041\n".len(), 21);
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-42");
    assert_eq!(stored(&f).as_deref(), Some("42\n"));
    // 22 bytes is refused before it is read, although it would parse as 41.
    let f = fixture();
    seed(&f, "000000000000000000041\n");
    assert_eq!("000000000000000000041\n".len(), 22);
    for args in [&["cc", "-a"][..], &["cc", "-a", "--dry-run"][..]] {
        refused(&run(&f, args));
    }
    assert_eq!(stored(&f).as_deref(), Some("000000000000000000041\n"));
}

#[test]
fn a_store_path_that_is_not_a_regular_file_is_refused_without_blocking() {
    // A directory.
    let f = fixture();
    fs::create_dir_all(store(&f)).unwrap();
    refused(&run(&f, &["cc", "-a"]));
    // A FIFO: a plain read would block forever.
    let f = fixture();
    fs::create_dir_all(store(&f).parent().unwrap()).unwrap();
    assert!(Command::new("mkfifo")
        .arg(store(&f))
        .status()
        .unwrap()
        .success());
    let mut child = command(&f, f.project.path(), &["cc", "-a"])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let start = Instant::now();
    while child.try_wait().unwrap().is_none() {
        if start.elapsed() > Duration::from_secs(20) {
            let _ = child.kill();
            panic!("launcher blocked on a FIFO store");
        }
        std::thread::sleep(Duration::from_millis(50));
    }
    refused(&child.wait_with_output().unwrap());
    // A symlink, even to a good file.
    let f = fixture();
    let real = f.project.path().join("elsewhere");
    fs::write(&real, "3\n").unwrap();
    fs::create_dir_all(store(&f).parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&real, store(&f)).unwrap();
    let err = refused(&run(&f, &["cc", "-a"]));
    assert!(err.contains("symlink"), "{err}");
    assert_eq!(fs::read_to_string(&real).unwrap(), "3\n");
}

#[test]
fn outside_a_git_project_numbering_is_an_error_but_a_given_name_works() {
    let f = fixture();
    let plain = tempfile::tempdir().unwrap();
    for args in [&["cc", "-a"][..], &["codex", "-a", "--dry-run"][..]] {
        let err = refused(&run_in(&f, plain.path(), args));
        assert!(err.contains("git repository"), "{err}");
    }
    assert_eq!(
        name(&run_in(&f, plain.path(), &["cc", "-a", "--name", "n1"])),
        "n1"
    );
    assert_eq!(name(&run_in(&f, plain.path(), &["cc", "-l"])), "leader");
}

#[test]
fn worktrees_of_one_clone_share_the_counter() {
    let f = fixture();
    git(
        f.project.path(),
        &["commit", "-q", "--allow-empty", "-m", "base"],
    );
    let wt = f.project.path().join("wt");
    git(
        f.project.path(),
        &["worktree", "add", "-q", "-b", "lane", "wt"],
    );
    assert_eq!(name(&run(&f, &["cc", "-a"])), "agent-1");
    assert_eq!(name(&run_in(&f, &wt, &["codex", "-a"])), "agent-2");
    assert_eq!(name(&run(&f, &["glm", "-a"])), "agent-3");
    assert_eq!(stored(&f).as_deref(), Some("3\n"));
}

fn launcher_config(f: &Fixture) -> PathBuf {
    f.home.path().join(".moai-cli/config/launcher.yaml")
}

fn assert_config_refusal_before_side_effects(f: &Fixture) {
    for host in ["cc", "glm", "codex"] {
        for role in [&["-l"][..], &["-a"][..], &[][..]] {
            for cli in [false, true] {
                let mut args = vec![host];
                args.extend_from_slice(role);
                if cli {
                    args.push("--permissions");
                }
                let out = run(f, &args);
                assert_eq!(out.status.code(), Some(1), "{args:?}");
                assert!(out.stdout.is_empty(), "{args:?}");
                assert!(!String::from_utf8_lossy(&out.stderr).contains("SECRET"));
                assert!(!f.home.path().join("host-started").exists(), "{args:?}");
                assert_eq!(stored(f).as_deref(), Some("00041\n"), "{args:?}");
            }
        }
    }
}

#[test]
fn invalid_launcher_config_never_executes_host_or_allocates_number() {
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
        seed(&f, "00041\n");
        fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
        fs::write(launcher_config(&f), bad).unwrap();
        assert_config_refusal_before_side_effects(&f);
    }
}

#[test]
fn nonregular_launcher_config_never_executes_host_or_allocates_number() {
    let f = fixture();
    seed(&f, "00041\n");
    fs::create_dir_all(launcher_config(&f)).unwrap();
    assert_config_refusal_before_side_effects(&f);
}

#[test]
fn unreadable_launcher_config_never_executes_host_or_allocates_number() {
    let f = fixture();
    seed(&f, "00041\n");
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
fn valid_launcher_settings_execute_fake_hosts_and_allocate_normally() {
    for setting in ["full", "prompt"] {
        let f = fixture();
        seed(&f, "00041\n");
        fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
        fs::write(
            launcher_config(&f),
            format!(r#"{{"launcher":{{"role_access":"{setting}"}}}}"#),
        )
        .unwrap();
        for (host, number) in [("cc", 42), ("glm", 43), ("codex", 44)] {
            assert_eq!(name(&run(&f, &[host, "-a"])), format!("agent-{number}"));
            assert_eq!(stored(&f), Some(format!("{number}\n")));
            assert_eq!(
                fs::read(f.home.path().join("host-started")).unwrap(),
                b"started"
            );
        }
    }
}

#[test]
fn symlink_launcher_config_is_refused_without_following_it() {
    let f = fixture();
    seed(&f, "00041\n");
    fs::create_dir_all(launcher_config(&f).parent().unwrap()).unwrap();
    let target = f.home.path().join("synthetic-config");
    let contents = r#"{"launcher":{"role_access":"full"}}"#;
    fs::write(&target, contents).unwrap();
    std::os::unix::fs::symlink(&target, launcher_config(&f)).unwrap();
    assert_config_refusal_before_side_effects(&f);
    assert_eq!(fs::read_to_string(target).unwrap(), contents);
}

/// #2361 AC-5 as amended by the 06:45Z user decision (`-l`, `-a` = full-access mode): a
/// launch with `--name X` or `-l` takes no number and starts the host with exactly this argv,
/// pinned byte for byte per host. Real launches against the stub hosts, not `--dry-run`.
#[test]
fn name_and_leader_launch_the_decided_full_access_argv_and_take_no_number() {
    let f = fixture();
    let cases: [(&[&str], &[&str]); 6] = [
        (
            &["cc", "-l"],
            &[
                "claude",
                "--name",
                "leader",
                "--model",
                "opus",
                "--effort",
                "medium",
                "--dangerously-skip-permissions",
            ],
        ),
        (
            &["glm", "-l"],
            &[
                "claude",
                "--name",
                "leader",
                "--model",
                "opus",
                "--effort",
                "medium",
                "--dangerously-skip-permissions",
            ],
        ),
        (
            &["codex", "-l"],
            &[
                "codex",
                "-s",
                "danger-full-access",
                "-a",
                "never",
                "MoAI session name: leader",
            ],
        ),
        (
            &["cc", "-a", "--name", "x"],
            &[
                "claude",
                "--name",
                "x",
                "--model",
                "opus",
                "--effort",
                "medium",
                "--dangerously-skip-permissions",
            ],
        ),
        (
            &["glm", "-a", "--name", "x"],
            &[
                "claude",
                "--name",
                "x",
                "--model",
                "opus",
                "--effort",
                "medium",
                "--dangerously-skip-permissions",
            ],
        ),
        (
            &["codex", "-a", "--name", "x"],
            &[
                "codex",
                "-s",
                "danger-full-access",
                "-a",
                "never",
                "MoAI session name: x",
            ],
        ),
    ];
    for (args, want) in cases {
        let out = run(&f, args);
        assert!(
            out.status.success(),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        let argv: Vec<&str> = v["argv"]
            .as_array()
            .unwrap_or_else(|| panic!("{args:?}: no argv in {v}"))
            .iter()
            .map(|a| a.as_str().unwrap())
            .collect();
        assert_eq!(argv, want, "{args:?}");
    }
    assert_eq!(stored(&f), None, "--name and -l take no number");
}
