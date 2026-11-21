//! One process snapshot serves both use checks, tests inject it from a table file and a stuck
//! `lsof` is stopped (SPEC-20261121-process-scan-snapshot), through the binary in a temporary
//! HOME only.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    time::{Duration, Instant},
};

const NOW: &str = "2026-11-21T05:00:00Z";

/// A temporary HOME whose `.moai` looks like the Go generation's.
fn go_home() -> TempHome {
    let home = TempHome::new();
    for (rel, body) in [
        (
            ".moai/user-assets.json",
            r#"{"schema_version": 1, "files": {}}"#,
        ),
        (".moai/state/active-sessions.json", "[]"),
        (".moai/credentials.yaml", "token: PLACEHOLDER\n"),
        (".moai/config/participation.yaml", "x: 1\n"),
    ] {
        let p = home.path().join(rel);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, body).unwrap();
    }
    home
}

/// A `lsof` script that logs its arguments, one line per call, then runs `then`.
fn stub_lsof(dir: &Path, log: &Path, then: &str) {
    fs::create_dir_all(dir).unwrap();
    let script = dir.join("lsof");
    fs::write(
        &script,
        format!("#!/bin/sh\necho \"$*\" >> '{}'\n{then}\n", log.display()),
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&script, fs::Permissions::from_mode(0o755)).unwrap();
    }
}

fn plan(home: &TempHome, stub_dir: &Path, env: &[(&str, &str)]) -> Value {
    let path = format!("{}:/usr/bin:/bin", stub_dir.display());
    let mut c = home.command(&[
        "home-state-plan",
        "--home",
        home.path().to_str().unwrap(),
        "--now",
        NOW,
    ]);
    c.env("PATH", path);
    for (k, v) in env {
        c.env(k, v);
    }
    let out = c.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn blocks(p: &Value) -> Vec<String> {
    p["blocks"]
        .as_array()
        .cloned()
        .unwrap_or_default()
        .iter()
        .filter_map(|b| b.as_str().map(str::to_string))
        .collect()
}

/// Linux reads `/proc` and never calls `lsof`.
fn reads_proc() -> bool {
    Path::new("/proc/self/fd").exists()
}

/// REQ-1: both checks of one HOME plan are served by a single `lsof` call.
#[test]
fn one_lsof_call_serves_both_checks() {
    if reads_proc() {
        return;
    }
    let home = go_home();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(
        &stub,
        &log,
        "printf 'fcwd\\nn/nowhere\\nf3\\nn/nowhere/file\\n'",
    );
    plan(&home, &stub, &[]);
    let calls = fs::read_to_string(&log).unwrap_or_default();
    let calls: Vec<&str> = calls.lines().collect();
    assert_eq!(calls.len(), 1, "{calls:?}");
    let words: Vec<&str> = calls[0].split_whitespace().collect();
    for flag in ["-w", "-n", "-P", "-b", "-F", "fn"] {
        assert!(
            words.contains(&flag),
            "{flag} is missing from `lsof {}`",
            calls[0]
        );
    }
}

/// REQ-5: a table file drives the decisions and `lsof` is never run.
#[test]
fn an_injected_table_file_drives_the_decisions_without_lsof() {
    if reads_proc() || !cfg!(debug_assertions) {
        return;
    }
    let home = go_home();
    let moai = fs::canonicalize(home.path().join(".moai")).unwrap();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(&stub, &log, "echo 'n/nowhere'");
    let table = home.path().join("table.txt");

    // A working folder inside: named by the working-folder check (and, as lsof lists a working
    // folder among open files, by the open-file check too).
    fs::write(&table, format!("cwd {}\n", moai.join("config").display())).unwrap();
    let b = blocks(&plan(
        &home,
        &stub,
        &[("MOAI_CLI_TEST_PROCESS_TABLE", table.to_str().unwrap())],
    ));
    assert!(b.contains(&"in_use".to_string()), "{b:?}");

    // Only an open file inside.
    fs::write(
        &table,
        format!("open {}\n", moai.join("credentials.yaml").display()),
    )
    .unwrap();
    let b = blocks(&plan(
        &home,
        &stub,
        &[("MOAI_CLI_TEST_PROCESS_TABLE", table.to_str().unwrap())],
    ));
    assert_eq!(b, ["file_open_in_use"], "{b:?}");

    // A sibling whose name only starts the same, and paths elsewhere: nothing is blocked.
    let sibling = format!("{}-other/x", moai.display());
    fs::write(
        &table,
        format!("cwd {sibling}\nopen {sibling}\nopen /elsewhere/file\n"),
    )
    .unwrap();
    let b = blocks(&plan(
        &home,
        &stub,
        &[("MOAI_CLI_TEST_PROCESS_TABLE", table.to_str().unwrap())],
    ));
    assert!(b.is_empty(), "{b:?}");

    assert!(!log.exists(), "lsof ran: {:?}", fs::read_to_string(&log));
}

/// REQ-6: an `lsof` that never answers is stopped and the check says it cannot tell.
#[test]
fn a_stuck_lsof_is_stopped_and_the_check_reports_unavailable() {
    if reads_proc() || !cfg!(debug_assertions) {
        return;
    }
    let home = go_home();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(&stub, &log, "exec sleep 20");
    let start = Instant::now();
    let p = plan(&home, &stub, &[("MOAI_CLI_TEST_LSOF_TIMEOUT_MS", "500")]);
    let took = start.elapsed();
    assert!(took < Duration::from_secs(12), "took {took:?}");
    let b = blocks(&p);
    assert!(b.iter().any(|x| x == "process_check_unavailable"), "{b:?}");
}
