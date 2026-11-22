//! #2668 S1: the status line keeps the usage windows it is given, and `usage-limit` judges
//! them — through the real binary, with HOME and TMPDIR pointing at temporary folders. The real
//! HOME and the real temporary folder are never read or written. The record lives in the status
//! line's cache folder under TMPDIR (SPEC-20261122-statusline-limits REQ-7), never under HOME.

mod common;
use common::CleanEnv;
use serde_json::Value;
use std::{
    io::Write,
    path::Path,
    process::{Command, Stdio},
    time::{SystemTime, UNIX_EPOCH},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

fn statusline(home: &Path, tmpdir: &Path, cwd: &Path, stdin: &str) {
    let mut child = Command::new(EXE)
        .arg("statusline")
        .current_dir(cwd)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("TMPDIR", tmpdir)
        .env("LC_ALL", "C")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
}

fn usage(home: &Path, tmpdir: &Path) -> Value {
    let out = Command::new(EXE)
        .args(["usage-limit", home.to_str().unwrap()])
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("TMPDIR", tmpdir)
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(v["operation"], "usage.limit", "{v}");
    v["result"].clone()
}

#[test]
fn statusline_render_feeds_usage_limit_in_a_temporary_home() {
    let home = TempDir::new().unwrap();
    let tmp = TempDir::new().unwrap();
    let cwd = TempDir::new().unwrap();
    assert_eq!(usage(home.path(), tmp.path())["state"], "unknown");

    // No rate_limits (API key, before the first response): nothing recorded.
    statusline(
        home.path(),
        tmp.path(),
        cwd.path(),
        r#"{"model":{"display_name":"Opus"}}"#,
    );
    assert_eq!(usage(home.path(), tmp.path())["state"], "unknown");

    let reset = now() + 3600;
    for (used, want) in [(89, "ok"), (90, "near"), (100, "reached")] {
        statusline(
            home.path(),
            tmp.path(),
            cwd.path(),
            &format!(
                r#"{{"model":{{"display_name":"Opus"}},"rate_limits":{{"five_hour":{{"used_percentage":{used},"resets_at":{reset}}}}}}}"#
            ),
        );
        let v = usage(home.path(), tmp.path());
        assert_eq!(v["state"], want, "{used}: {v}");
        if want != "ok" {
            assert_eq!(v["window"], "five_hour");
            assert!(v["renew_until"].as_str().unwrap() > v["resets_at"].as_str().unwrap());
        }
    }
    // Nothing under the temporary HOME; the record the reader names is the one the product
    // function names, below the temporary TMPDIR.
    assert_eq!(std::fs::read_dir(home.path()).unwrap().count(), 0);
    let record = moai_cli::usage_limit::record_path(home.path(), tmp.path().to_str()).unwrap();
    assert!(record.is_file(), "{record:?}");
    assert!(record.starts_with(tmp.path()), "{record:?}");
    assert_eq!(
        usage(home.path(), tmp.path())["record"],
        record.to_str().unwrap()
    );
}
