//! REQ-2 of SPEC-20261120-test-parallelism-cap: the wrapper's cap — a wrapped `cargo test`
//! runs libtest with the machine's test-threads (`limits.json`'s `test_threads`, default 4)
//! through `RUST_TEST_THREADS` unless the caller set that variable itself, a wrapped
//! `cargo nextest run` gets `-j <threads>`, a non-test command is untouched, and the
//! `slot taken` line names the value it applied. The wrapped command is a stub `cargo` that
//! records its environment and arguments into `RECORD_DIR`, so the assertions read what the
//! run really saw, not what the wrapper promised.

mod common;

use common::home::TempHome;
use std::fs;
use std::path::{Path, PathBuf};

/// A stub `cargo` that records its environment and arguments, then exits 0.
fn stub_cargo(home: &TempHome) -> PathBuf {
    let stub = home.path().join("stub-cargo").join("cargo");
    fs::create_dir_all(stub.parent().unwrap()).unwrap();
    common::write_executable(
        &stub,
        "#!/bin/sh\nenv > \"$RECORD_DIR/env.txt\"\nfor a in \"$@\"; do printf '%s\\n' \"$a\" >> \"$RECORD_DIR/args.txt\"; done\nexit 0\n",
    );
    stub
}

/// The slot folder with a `limits.json` that names `test_threads` and loosens the machine
/// gates, so the acquisition runs whatever the machine is doing.
fn slots(home: &TempHome, threads: u64) -> PathBuf {
    let dir = home.path().join("slots");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("limits.json"),
        format!(
            "{{\"test_threads\": {threads}, \"min_free_gb\": 1, \"max_swap_pct\": 100, \
             \"min_mem_free_pct\": 0}}"
        ),
    )
    .unwrap();
    dir
}

/// Runs `heavy-slot-acquire --dir … -- <stub cargo …>` with `RECORD_DIR` and the extra
/// environment given; `(exit code, stdout, stderr)`. `test_args` is what follows the stub
/// cargo on the command line (`test --offline`, `nextest run`, `build`).
fn acquire(
    home: &TempHome,
    slots_dir: &Path,
    cargo: &Path,
    record: &Path,
    test_args: &[&str],
    env: &[(&str, &str)],
) -> (i32, String, String) {
    let mut argv = vec![
        "heavy-slot-acquire".to_string(),
        "--dir".to_string(),
        slots_dir.to_str().unwrap().to_string(),
        "--max".to_string(),
        "1".to_string(),
        "--name".to_string(),
        "t".to_string(),
        "--".to_string(),
        cargo.to_str().unwrap().to_string(),
    ];
    argv.extend(test_args.iter().map(|s| s.to_string()));
    let refs: Vec<&str> = argv.iter().map(String::as_str).collect();
    let mut cmd = home.command_for(common::MOAI_BIN, &refs);
    cmd.env("RECORD_DIR", record);
    for (k, v) in env {
        cmd.env(k, v);
    }
    let out = cmd.output().unwrap();
    (
        out.status.code().unwrap_or(-1),
        String::from_utf8_lossy(&out.stdout).into_owned(),
        String::from_utf8_lossy(&out.stderr).into_owned(),
    )
}

fn record_of(home: &TempHome) -> (String, String) {
    (
        fs::read_to_string(home.path().join("record/env.txt")).unwrap_or_default(),
        fs::read_to_string(home.path().join("record/args.txt")).unwrap_or_default(),
    )
}

#[test]
fn a_wrapped_cargo_test_runs_libtest_with_the_cap() {
    let home = TempHome::new();
    let cargo = stub_cargo(&home);
    let slots_dir = slots(&home, 2);
    let record = home.path().join("record");
    fs::create_dir_all(&record).unwrap();
    let (code, stdout, stderr) = acquire(
        &home,
        &slots_dir,
        &cargo,
        &record,
        &["test", "--offline"],
        &[],
    );
    assert_eq!(code, 0, "{stdout} {stderr}");
    assert!(stdout.contains("test-threads=2"), "{stdout}");
    let (env_txt, _args) = record_of(&home);
    assert!(
        env_txt.contains("RUST_TEST_THREADS=2"),
        "libtest runs with the capped threads: {env_txt}"
    );
}

#[test]
fn a_caller_set_rust_test_threads_wins() {
    let home = TempHome::new();
    let cargo = stub_cargo(&home);
    let slots_dir = slots(&home, 2);
    let record = home.path().join("record");
    fs::create_dir_all(&record).unwrap();
    let (code, stdout, _stderr) = acquire(
        &home,
        &slots_dir,
        &cargo,
        &record,
        &["test", "--offline"],
        &[("RUST_TEST_THREADS", "8")],
    );
    assert_eq!(code, 0, "{stdout}");
    // The caller's own knob wins, so the line names nothing: nothing was capped.
    assert!(!stdout.contains("test-threads"), "{stdout}");
    let (env_txt, _args) = record_of(&home);
    assert!(env_txt.contains("RUST_TEST_THREADS=8"), "{env_txt}");
    assert!(!env_txt.contains("RUST_TEST_THREADS=2"), "{env_txt}");
}

#[test]
fn a_wrapped_nextest_run_gets_the_j_flag() {
    let home = TempHome::new();
    let cargo = stub_cargo(&home);
    let slots_dir = slots(&home, 2);
    let record = home.path().join("record");
    fs::create_dir_all(&record).unwrap();
    let (code, stdout, _stderr) =
        acquire(&home, &slots_dir, &cargo, &record, &["nextest", "run"], &[]);
    assert_eq!(code, 0, "{stdout}");
    assert!(stdout.contains("test-threads=2"), "{stdout}");
    let (_env_txt, args) = record_of(&home);
    let lines: Vec<&str> = args.lines().collect();
    let j_at = lines.iter().position(|l| *l == "-j").unwrap();
    assert_eq!(lines[j_at + 1], "2", "the threads follow -j: {args}");
}

#[test]
fn a_non_test_command_is_untouched() {
    let home = TempHome::new();
    let cargo = stub_cargo(&home);
    let slots_dir = slots(&home, 2);
    let record = home.path().join("record");
    fs::create_dir_all(&record).unwrap();
    let (code, stdout, _stderr) = acquire(
        &home,
        &slots_dir,
        &cargo,
        &record,
        &["build", "--offline"],
        &[],
    );
    assert_eq!(code, 0, "{stdout}");
    assert!(!stdout.contains("test-threads"), "{stdout}");
    let (env_txt, _args) = record_of(&home);
    assert!(
        !env_txt.contains("RUST_TEST_THREADS"),
        "a build is not capped: {env_txt}"
    );
}
