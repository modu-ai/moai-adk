//! The star-ask CLI (SPEC-20261122-star-and-reports): the star state is
//! checked through a stand-in gh (one read-only GET to the pinned URL, no
//! write call), the line shows once when unknown or not starred, never when
//! starred or dismissed — and a dismissed home makes no network call at all
//! (B1). `--dismiss` reports the truth of its write (B3). The line comes in
//! both languages (B4).

mod common;
use common::home::TempHome;
use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::Stdio,
};
use tempfile::TempDir;

/// The stub gh: the star check reads only one endpoint; a script that prints
/// the given HTTP status envelope answers it. The calls file records every
/// call (N1: the URL and the count are pinned).
fn install_stub_gh(dir: &Path, tag: &str, status: u16, calls: &Path) -> PathBuf {
    let path = dir.join(format!("gh-{tag}"));
    let script = format!(
        "#!/bin/sh\necho \"$@\" >> {calls}\nprintf 'HTTP/1.1 {status} OK\\n\\n{{}}\\n'\n",
        calls = calls.display(),
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

fn dismiss(home: &TempHome) -> Value {
    let mut child = home
        .command(&["star-ask", "--dismiss"])
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

/// AC-1, B1, N1: with a stand-in gh, the star line shows once when not
/// starred, never when starred or dismissed; a dismissed home makes zero
/// network calls; the URL is pinned; no write call is made.
#[test]
fn ac1_the_star_line_follows_the_state_no_write_and_no_call_when_dismissed() {
    let stubs = TempDir::new().unwrap();
    let calls = stubs.path().join("calls");

    // Not starred (404): the line shows.
    let gh_404 = install_stub_gh(stubs.path(), "404", 404, &calls);
    let home = TempHome::new();
    let v = star_ask(&home, &gh_404);
    assert_eq!(v["state"], json!("not_starred"), "{v}");
    assert!(
        v["line"].as_str().is_some_and(|line| !line.is_empty()),
        "{v}"
    );

    // N1: the call went to the pinned URL.
    let recorded = fs::read_to_string(&calls).unwrap_or_default();
    assert!(
        recorded.contains("/user/starred/modu-ai/moai-adk")
            || recorded.contains("user/starred/modu-ai/moai-adk"),
        "the URL is pinned: {recorded}"
    );
    let call_count = recorded.lines().count();
    assert_eq!(call_count, 1, "exactly one call: {recorded}");

    // Starred (204): the line never shows.
    let gh_204 = install_stub_gh(stubs.path(), "204", 204, &calls);
    let v = star_ask(&home, &gh_204);
    assert_eq!(v["state"], json!("starred"), "{v}");
    assert_eq!(v["line"], json!(null), "{v}");

    // Dismissed (B3: reports the truth), then B1: zero network calls.
    let v = dismiss(&home);
    assert_eq!(v["dismissed"], json!(true), "{v}");

    // B1: a dismissed home makes no network call at all.
    let calls_before_dismiss = fs::read_to_string(&calls).unwrap().lines().count();
    let v = star_ask(&home, &gh_404);
    assert_eq!(v["state"], json!("dismissed"), "{v}");
    assert_eq!(v["line"], json!(null), "{v}");
    let calls_after = fs::read_to_string(&calls).unwrap().lines().count();
    assert_eq!(
        calls_after, calls_before_dismiss,
        "a dismissed home makes zero network calls"
    );
}

/// B4: the line exists in both languages (the module's own table).
#[test]
fn ac2_the_line_exists_in_both_languages() {
    let en = moai_cli::star_ask::star_ask_line(moai_cli::ui_text::Lang::En);
    let ko = moai_cli::star_ask::star_ask_line(moai_cli::ui_text::Lang::Ko);
    assert!(en.contains("github.com/modu-ai/moai-adk"), "{en}");
    assert!(ko.contains("github.com/modu-ai/moai-adk"), "{ko}");
    assert_ne!(en, ko, "the two languages carry different text");
}
