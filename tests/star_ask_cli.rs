//! The star-ask CLI (SPEC-20261122-star-and-reports): the star state is
//! checked through a stand-in gh (one read-only GET), the line shows once
//! when unknown or not starred, never when starred or dismissed, no write
//! call exists on the check path, and the dismissal marker survives.

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    io::Write as _,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

/// The stub gh: the star check reads only one endpoint; a script that prints
/// the given HTTP status envelope answers it. The writes file records any
/// call with a method other than GET (the no-write proof).
fn install_stub_gh(dir: &Path, tag: &str, status: u16, writes: &Path) -> PathBuf {
    let path = dir.join(format!("gh-{tag}"));
    let script = format!(
        "#!/bin/sh\nif echo \"$@\" | grep -qv GET; then echo \"$@\" >> {writes}; fi\nprintf 'HTTP/1.1 {status} OK\\n\\n{{}}\\n'\n",
        writes = writes.display(),
    );
    fs::write(&path, script).unwrap();
    make_executable(&path);
    path
}

fn make_executable(path: &Path) {
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt as _;
        let mut perms = fs::metadata(path).unwrap().permissions();
        perms.set_mode(0o755);
        fs::set_permissions(path, perms).unwrap();
    }
}

fn star_ask(home: &TempHome, gh: &Path) -> Value {
    let mut child = home
        .command(&["star-ask"])
        .env("MOAI_FORGE_GH_BIN", gh)
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child.stdin.take();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn dismiss(home: &TempHome) {
    let mut child = home
        .command(&["star-ask", "--dismiss"])
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let _ = child.stdin.take();
    let out = child.wait_with_output().unwrap();
    assert!(out.status.success());
}

/// AC-1: with a stand-in gh, the star line shows once when not starred,
/// never when starred or dismissed, and no write call is made.
#[test]
fn ac1_the_star_line_follows_the_state_and_no_write_call_is_made() {
    let stubs = TempDir::new().unwrap();
    let writes = stubs.path().join("writes");

    // Not starred (404): the line shows.
    let gh_404 = install_stub_gh(stubs.path(), "404", 404, &writes);
    let home = TempHome::new();
    let v = star_ask(&home, &gh_404);
    assert_eq!(v["state"], json!("not_starred"), "{v}");
    assert!(
        v["line"].as_str().is_some_and(|line| !line.is_empty()),
        "{v}"
    );

    // Starred (204): the line never shows.
    let gh_204 = install_stub_gh(stubs.path(), "204", 204, &writes);
    let v = star_ask(&home, &gh_204);
    assert_eq!(v["state"], json!("starred"), "{v}");
    assert_eq!(v["line"], json!(null), "{v}");

    // Dismissed: the line never shows regardless of the state.
    dismiss(&home);
    let v = star_ask(&home, &gh_404);
    assert_eq!(v["state"], json!("not_starred"), "{v}");
    assert_eq!(v["line"], json!(null), "{v}");

    // No write call was made: the writes file is empty.
    assert_eq!(
        fs::read_to_string(&writes).unwrap_or_default(),
        "",
        "only GET calls reached the stub"
    );
}

/// AC-1 continued: gh absent or failing answers unknown and the line still
/// shows once (the check's failure answer, not an error).
#[test]
fn ac1b_a_failing_gh_answers_unknown_and_the_line_still_shows() {
    let stubs = TempDir::new().unwrap();
    let writes = stubs.path().join("writes");
    let gh_fail = stubs.path().join("gh-fail");
    fs::write(&gh_fail, "#!/bin/sh\nprintf 'this is not HTTP\\n'\n").unwrap();
    make_executable(&gh_fail);
    let home = TempHome::new();
    let v = star_ask(&home, &gh_fail);
    assert_eq!(v["state"], json!("unknown"), "{v}");
    assert!(
        v["line"].as_str().is_some_and(|line| !line.is_empty()),
        "{v}"
    );
}
