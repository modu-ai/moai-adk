//! The process scans of the HOME commands run `lsof` cheaply and can be injected
//! (SPEC-20261121-cheap-process-scan), through the binary in a temporary HOME only.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{fs, path::Path};

const NOW: &str = "2026-11-21T03:00:00Z";

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

/// A `lsof` that logs its arguments, one line per call, and prints one unrelated path.
fn stub_lsof(dir: &Path, log: &Path) {
    fs::create_dir_all(dir).unwrap();
    let script = dir.join("lsof");
    fs::write(
        &script,
        format!(
            "#!/bin/sh\necho \"$*\" >> '{}'\necho 'n/nowhere/at/all'\n",
            log.display()
        ),
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

/// Linux reads `/proc` and never calls `lsof`.
fn reads_proc() -> bool {
    Path::new("/proc/self/fd").exists()
}

/// REQ-1, REQ-2: the scan leaves out name lookups and blocking kernel calls.
#[test]
fn the_process_scans_run_lsof_without_name_lookups_or_blocking_calls() {
    if reads_proc() {
        return;
    }
    let home = go_home();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(&stub, &log);
    plan(&home, &stub, &[]);
    let calls = fs::read_to_string(&log).unwrap_or_default();
    let calls: Vec<&str> = calls.lines().collect();
    assert!(!calls.is_empty(), "{calls:?}");
    for call in &calls {
        let words: Vec<&str> = call.split_whitespace().collect();
        for flag in ["-w", "-n", "-P", "-b"] {
            assert!(
                words.contains(&flag),
                "{flag} is missing from `lsof {call}`"
            );
        }
    }
}

/// REQ-4: an injected empty table is read without running `lsof` and blocks nothing.
#[test]
fn the_injected_empty_table_skips_lsof_and_blocks_nothing() {
    if reads_proc() || !cfg!(debug_assertions) {
        return;
    }
    let home = go_home();
    let stub = home.path().join("stub");
    let log = home.path().join("lsof.log");
    stub_lsof(&stub, &log);
    let p = plan(&home, &stub, &[("MOAI_CLI_TEST_PROCESS_TABLE", "empty")]);
    assert!(!log.exists(), "lsof ran: {:?}", fs::read_to_string(&log));
    let blocks = p["blocks"].as_array().cloned().unwrap_or_default();
    assert!(
        !blocks
            .iter()
            .any(|b| b == "in_use" || b == "file_open_in_use"),
        "{p}"
    );
    assert!(
        !blocks
            .iter()
            .any(|b| b.as_str().is_some_and(|s| s.ends_with("_unavailable"))),
        "{p}"
    );
}
