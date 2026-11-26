//! The star-ask CLI (SPEC-20261122-star-and-reports): the star state is
//! checked through a stand-in gh (one read-only GET to the pinned URL, no
//! write call), the line shows once when unknown or not starred, never when
//! starred or dismissed — and a dismissed home makes no network call at all.
//! `--dismiss` answers one shape on every outcome: an object with `written`
//! and `reason`, the retired folder and a refused write named in the reason.
//! The line comes from the shared UI text table in the stream's language.

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
/// call (the URL and the count are pinned).
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

/// AC-1: with a stand-in gh, the star line shows once when not starred,
/// never when starred or dismissed; a dismissed home makes zero network
/// calls; the URL is pinned; no write call is made. The line is the UI text
/// table's row for the stream's language (a pipe is English).
#[test]
fn ac1_the_star_line_follows_the_state_no_write_and_no_call_when_dismissed() {
    let stubs = TempDir::new().unwrap();
    let calls = stubs.path().join("calls");

    // Not starred (404): the line shows, and it is the table's English row.
    let gh_404 = install_stub_gh(stubs.path(), "404", 404, &calls);
    let home = TempHome::new();
    let v = star_ask(&home, &gh_404);
    assert_eq!(v["state"], json!("not_starred"), "{v}");
    let en = moai_cli::ui_text::text(
        moai_cli::ui_text::Msg::StarAskLine,
        moai_cli::ui_text::Lang::En,
    );
    assert_eq!(v["line"], json!(en), "{v}");

    // The call went to the pinned URL.
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

    // Dismissed (the truth of the write, one shape), then zero network calls.
    let v = dismiss(&home);
    assert_eq!(v["dismissed"]["written"], json!(true), "{v}");
    assert_eq!(v["dismissed"]["reason"], json!(null), "{v}");

    // A dismissed home makes no network call at all.
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

/// The line lives in the shared UI text table in both languages, and the
/// module function serves the table's rows (a mutant with its own constant
/// cannot survive).
#[test]
fn ac2_the_line_comes_from_the_ui_text_table_in_both_languages() {
    let en = moai_cli::ui_text::text(
        moai_cli::ui_text::Msg::StarAskLine,
        moai_cli::ui_text::Lang::En,
    );
    let ko = moai_cli::ui_text::text(
        moai_cli::ui_text::Msg::StarAskLine,
        moai_cli::ui_text::Lang::Ko,
    );
    assert_eq!(
        moai_cli::star_ask::star_ask_line(moai_cli::ui_text::Lang::En),
        en
    );
    assert_eq!(
        moai_cli::star_ask::star_ask_line(moai_cli::ui_text::Lang::Ko),
        ko
    );
    assert!(en.contains("github.com/modu-ai/moai-adk"), "{en}");
    assert!(ko.contains("github.com/modu-ai/moai-adk"), "{ko}");
    assert_ne!(en, ko, "the two languages carry different text");
}

/// A home whose resolver answers the retired folder (an earlier-generation
/// home folder exists): `--dismiss` writes nothing, exits 0, keeps the one
/// shape and names the retired folder in the reason.
#[test]
fn dismiss_on_a_retired_folder_home_names_the_reason_and_writes_nothing() {
    let home = TempHome::new();
    fs::create_dir_all(home.path().join(moai_cli::HOME_STATE_DIR)).unwrap();
    let v = dismiss(&home);
    assert_eq!(v["dismissed"]["written"], json!(false), "{v}");
    let reason = v["dismissed"]["reason"].as_str().unwrap_or_default();
    assert!(
        reason.contains("retired"),
        "the reason names the retired folder: {v}"
    );
    assert!(
        !home
            .path()
            .join(moai_cli::HOME_STATE_DIR)
            .join("star-ask-dismissed")
            .exists(),
        "nothing new is created under the retired folder"
    );
}

/// A read-only state folder: the write is refused, `--dismiss` exits 0,
/// keeps the one shape and carries the refusal's reason.
#[test]
fn dismiss_on_a_read_only_home_reports_the_refused_write() {
    let home = TempHome::new();
    let v = dismiss(&home);
    assert_eq!(v["dismissed"]["written"], json!(true), "{v}");
    let state = home.path().join(moai_cli::home_root::MOAI_DIR);
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt as _;
        fs::remove_file(state.join("star-ask-dismissed")).unwrap();
        let mut perms = fs::metadata(&state).unwrap().permissions();
        perms.set_mode(0o555);
        fs::set_permissions(&state, perms).unwrap();
    }
    let v = dismiss(&home);
    assert_eq!(v["dismissed"]["written"], json!(false), "{v}");
    let reason = v["dismissed"]["reason"].as_str().unwrap_or_default();
    assert!(
        reason.contains("cannot be written"),
        "the reason carries the refusal: {v}"
    );
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt as _;
        let mut perms = fs::metadata(&state).unwrap().permissions();
        perms.set_mode(0o755);
        fs::set_permissions(&state, perms).unwrap();
    }
}
