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

/// The status line's plain text for one payload: `NO_COLOR`, the hidden clock and render-budget
/// seams of the goldens (`tests/statusline_golden.rs`), a temporary HOME, TMPDIR and project.
fn statusline_text(home: &Path, tmpdir: &Path, cwd: &Path, stdin: &str) -> String {
    let mut child = Command::new(EXE)
        .args([
            "statusline",
            "--render-budget-ms",
            "3600000",
            "--now-unix",
            "1738429000",
        ])
        .current_dir(cwd)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("TMPDIR", tmpdir)
        .env("LC_ALL", "C")
        .env("NO_COLOR", "1")
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
    String::from_utf8(out.stdout).unwrap()
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

/// SPEC-20261122-glm-profile REQ-6 / AC-5: a GLM-shaped payload (the gateway model, the declared
/// 1M window, no `rate_limits` object) shows neither the five-hour nor the seven-day window, and
/// a payload that carries both renders them. The golden strings are the whole plain output.
#[test]
fn glm_absent_windows_are_left_out_and_present_windows_render() {
    let home = TempDir::new().unwrap();
    let tmp = TempDir::new().unwrap();
    let cwd = TempDir::new().unwrap();
    let project = cwd.path().to_str().unwrap();
    let payload = |rate_limits: &str| {
        format!(
            r#"{{"model":{{"id":"glm-5.4[1m]","display_name":"glm-5.4[1m]"}},"workspace":{{"current_dir":"{project}","project_dir":"{project}"}},"context_window":{{"context_window_size":1000000,"used_percentage":12}}{rate_limits}}}"#
        )
    };

    // The second line names the binary's version, which changes with every build; the goldens
    // keep the place with `<binary-version>` as `tests/fixtures/statusline-claude` does.
    let golden_form = |text: String| {
        text.lines()
            .map(|line| match line.strip_prefix("🗿 ") {
                Some(_) => "🗿 <binary-version>".to_string(),
                None => line.to_string(),
            })
            .collect::<Vec<_>>()
            .join("\n")
    };
    let glm = golden_form(statusline_text(
        home.path(),
        tmp.path(),
        cwd.path(),
        &payload(""),
    ));
    assert_eq!(
        glm, "🧠 glm-5.4\n🗿 <binary-version>\n🔋 CW █░░░░░░░░░ 12%",
        "{glm}"
    );

    let reset_5h = 1738429000 + 3600;
    let reset_7d = 1738429000 + 3 * 86400;
    let with = golden_form(statusline_text(
        home.path(),
        tmp.path(),
        cwd.path(),
        &payload(&format!(
            r#","rate_limits":{{"five_hour":{{"used_percentage":25,"resets_at":{reset_5h}}},"seven_day":{{"used_percentage":40,"resets_at":{reset_7d}}}}}"#
        )),
    ));
    assert_eq!(
        with,
        "🧠 glm-5.4\n🗿 <binary-version>\n🔋 CW █░░░░░░░░░ 12% | 🔋 5H ██░░░░░░░░ 25% (1h 0m) | 🔋 7D ████░░░░░░ 40% (3d 0h)",
        "{with}"
    );
    // Nothing under the temporary HOME.
    assert_eq!(std::fs::read_dir(home.path()).unwrap().count(), 0);
}
