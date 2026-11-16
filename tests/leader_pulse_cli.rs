//! #3015 (C5, decision 2=B): `leader-pulse` checks one round instruction and prints the
//! `orca terminal send` command line and the interval. It runs nothing: the test puts a
//! `orca` script that records any call first on PATH and proves it was never started.
mod common;

use common::CleanEnv;
use serde_json::Value;
use std::fs;
use std::path::Path;
use std::process::Command;

fn run(args: &[&str], dir: &Path) -> (i32, Value, String) {
    let bin = dir.join("bin");
    fs::create_dir_all(&bin).unwrap();
    let orca = bin.join("orca");
    fs::write(
        &orca,
        format!(
            "#!/bin/sh\necho \"$@\" >> '{}'\n",
            dir.join("orca-called").display()
        ),
    )
    .unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&orca, fs::Permissions::from_mode(0o755)).unwrap();
    }
    let out = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", dir)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.display()))
        .current_dir(dir)
        .args(args)
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    let v: Value = serde_json::from_str(&text)
        .unwrap_or_else(|_| panic!("not JSON: {text} {}", String::from_utf8_lossy(&out.stderr)));
    (out.status.code().unwrap_or(-1), v, text)
}

#[test]
fn it_prints_the_command_and_never_runs_orca() {
    let dir = tempfile::tempdir().unwrap();
    let (code, v, _) = run(
        &[
            "leader-pulse",
            "--terminal",
            "term_abc",
            "--text",
            "순환 계속",
            "--rounds",
            "2",
        ],
        dir.path(),
    );
    assert_eq!((code, &v["ok"]), (0, &Value::Bool(true)), "{v}");
    let r = &v["result"];
    assert_eq!(r["interval_secs"], 300, "the default interval: {r}");
    assert_eq!(r["rounds"], 2);
    assert_eq!(
        r["command_line"],
        "orca terminal send --terminal term_abc --text '순환 계속' --enter --wait-submit 20 --json"
    );
    assert_eq!(r["writes"], false);
    assert_eq!(r["rules"]["resend_when_turn_not_started"], false);
    assert!(
        !dir.path().join("orca-called").exists(),
        "leader-pulse must not start orca"
    );
}

#[test]
fn bad_inputs_are_refused_and_start_nothing() {
    let dir = tempfile::tempdir().unwrap();
    for args in [
        vec!["leader-pulse", "--terminal", "-x", "--text", "go"],
        vec!["leader-pulse", "--terminal", "a b", "--text", "go"],
        vec!["leader-pulse", "--terminal", "t", "--text", "a\nb"],
        vec!["leader-pulse", "--terminal", "t", "--text", "--enter"],
        vec![
            "leader-pulse",
            "--terminal",
            "t",
            "--text",
            "go",
            "--interval-secs",
            "59",
        ],
        vec![
            "leader-pulse",
            "--terminal",
            "t",
            "--text",
            "go",
            "--interval-secs",
            "3601",
        ],
        vec![
            "leader-pulse",
            "--terminal",
            "t",
            "--text",
            "go",
            "--rounds",
            "0",
        ],
    ] {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", dir.path())
            .current_dir(dir.path())
            .args(&args)
            .output()
            .unwrap();
        assert!(!out.status.success(), "{args:?} must be refused");
    }
    assert!(!dir.path().join("orca-called").exists());
}
