//! #2668 S1: the status line keeps the usage windows it is given, and `usage-limit` judges
//! them — through the real binary, with HOME pointing at a temporary folder. The real HOME
//! is never read or written.

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

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

fn now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

fn statusline(home: &Path, cwd: &Path, stdin: &str) {
    let mut child = Command::new(EXE)
        .arg("statusline")
        .current_dir(cwd)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
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

fn usage(home: &Path) -> Value {
    let out = Command::new(EXE)
        .args(["usage-limit", home.to_str().unwrap()])
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
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
    let cwd = TempDir::new().unwrap();
    assert_eq!(usage(home.path())["state"], "unknown");

    // No rate_limits (API key, before the first response): nothing recorded.
    statusline(
        home.path(),
        cwd.path(),
        r#"{"model":{"display_name":"Opus"}}"#,
    );
    assert_eq!(usage(home.path())["state"], "unknown");

    let reset = now() + 3600;
    for (used, want) in [(89, "ok"), (90, "near"), (100, "reached")] {
        statusline(
            home.path(),
            cwd.path(),
            &format!(
                r#"{{"model":{{"display_name":"Opus"}},"rate_limits":{{"five_hour":{{"used_percentage":{used},"resets_at":{reset}}}}}}}"#
            ),
        );
        let v = usage(home.path());
        assert_eq!(v["state"], want, "{used}: {v}");
        if want != "ok" {
            assert_eq!(v["window"], "five_hour");
            assert!(v["renew_until"].as_str().unwrap() > v["resets_at"].as_str().unwrap());
        }
    }
    // Only the record under the temporary HOME was written.
    let state = home.path().join(".moai-cli/state");
    let names: Vec<String> = std::fs::read_dir(&state)
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names, vec!["usage-limits.json"]);
}
