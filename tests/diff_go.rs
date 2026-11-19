//! SPEC-20261119-diff-go (card C-104): differential tests against the Go build `moai.old`, the
//! oracle for behaviour the Rust build keeps. Each case runs both binaries on the same input in
//! its own temporary HOME and folder, never the real ones and never the network. The Go build
//! is looked up in `MOAI_GO_BIN`, else `$HOME/go/bin/moai.old` (a path read, nothing written);
//! when it is absent the tests print why and pass without comparing (CI has no Go build).
//! What differs on purpose is listed in `tests/fixtures/diff_go/approved-differences.json`.
//! This is not a port of Go behaviour: only a behaviour the Rust build chose to keep is
//! compared, and a difference the Rust build chose is written down there with its reason.

#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::Value;
use std::fs;
use std::io::Write;
use std::path::PathBuf;
use std::process::{Command, Stdio};
use tempfile::TempDir;

const RUST: &str = crate::common::MOAI_BIN;
const APPROVED: &str = include_str!("fixtures/diff_go/approved-differences.json");

/// The Go build for these two lookup values, or the reason there is none: `MOAI_GO_BIN` first,
/// else `go/bin/moai.old` below `home`.
fn locate(var: Option<PathBuf>, home: Option<PathBuf>) -> Result<PathBuf, String> {
    let path = var
        .or_else(|| home.map(|h| h.join("go/bin/moai.old")))
        .ok_or("no MOAI_GO_BIN and no HOME to look in")?;
    if path.is_file() {
        Ok(path)
    } else {
        Err(format!(
            "the Go build {} is absent (set MOAI_GO_BIN)",
            path.display()
        ))
    }
}

/// The Go build, or `None` with the reason printed.
fn go_binary() -> Option<PathBuf> {
    match locate(
        std::env::var_os("MOAI_GO_BIN").map(PathBuf::from),
        std::env::var_os("HOME").map(PathBuf::from),
    ) {
        Ok(path) => Some(path),
        Err(reason) => {
            eprintln!("diff_go: skipped, {reason}");
            None
        }
    }
}

struct Run {
    code: Option<i32>,
    stdout: String,
    stderr: String,
}

/// Runs `exe args` with `stdin` in a fresh temporary HOME and folder.
fn run(exe: &std::path::Path, args: &[&str], stdin: &str) -> Run {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let home = root.join("home");
    let work = root.join("work");
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&work).unwrap();
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_CEILING_DIRECTORIES", &root)
        .env("LC_ALL", "C")
        .current_dir(&work)
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
    Run {
        code: out.status.code(),
        stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
        stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
    }
}

/// The text without terminal colour codes.
fn plain(text: &str) -> String {
    let mut out = String::new();
    let mut chars = text.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '\u{1b}' && chars.peek() == Some(&'[') {
            for n in chars.by_ref() {
                if n.is_ascii_alphabetic() {
                    break;
                }
            }
        } else {
            out.push(c);
        }
    }
    out
}

fn approved(case: &str) -> bool {
    let list: Value = serde_json::from_str(APPROVED).unwrap();
    list["differences"]
        .as_array()
        .unwrap()
        .iter()
        .any(|d| d["case"] == case && d["reason"].as_str().is_some_and(|r| !r.is_empty()))
}

/// Which of `names` the status line shows.
fn shown(r: &Run, names: &[&str]) -> Vec<String> {
    let text = plain(&r.stdout);
    names
        .iter()
        .filter(|n| text.contains(**n))
        .map(|n| n.to_string())
        .collect()
}

/// The status line reads the model name from the payload the same way in both builds: a later
/// duplicate key wins, a string is a name, key case is ignored, and a payload that is not valid
/// or has a wrong type shows no model (measured against moai.old v3.2.1-rc.6, the rules of
/// `src/statusline/godecode.rs`).
#[test]
fn the_status_line_reads_the_model_of_a_payload_like_the_go_build() {
    let Some(go) = go_binary() else { return };
    let rust = PathBuf::from(RUST);
    // (case, payload, names the payload could show, names it must show)
    let cases: Vec<(&str, &str, Vec<&str>, Vec<&str>)> = vec![
        (
            "display_name_of_an_object",
            r#"{"model":{"id":"x","display_name":"Alpha Name"}}"#,
            vec!["Alpha Name"],
            vec!["Alpha Name"],
        ),
        (
            "a_later_duplicate_key_wins",
            r#"{"model":{"display_name":"First Name"},"model":{"display_name":"Second Name"}}"#,
            vec!["First Name", "Second Name"],
            vec!["Second Name"],
        ),
        (
            "a_string_is_the_name",
            r#"{"model":"string-model-name"}"#,
            vec!["string-model-name"],
            vec!["string-model-name"],
        ),
        (
            "key_case_is_ignored",
            r#"{"Model":{"display_name":"Folded Name"}}"#,
            vec!["Folded Name"],
            vec!["Folded Name"],
        ),
        (
            "a_wrong_type_shows_no_model",
            r#"{"model":5,"session_id":"Never Shown"}"#,
            vec!["Never Shown"],
            vec![],
        ),
        (
            "invalid_json_shows_no_model",
            r#"{"model":"#,
            vec!["Never Shown"],
            vec![],
        ),
        (
            "an_empty_object_shows_no_model",
            "{}",
            vec!["Never Shown"],
            vec![],
        ),
    ];
    for (case, payload, names, expected) in cases {
        let g = run(&go, &["statusline"], payload);
        let r = run(&rust, &["statusline"], payload);
        let (gs, rs) = (shown(&g, &names), shown(&r, &names));
        assert_eq!(
            gs, expected,
            "{case}: the Go build shows {gs:?} (oracle sanity)"
        );
        assert_eq!(
            rs,
            gs,
            "{case}: Rust {rs:?} against Go {gs:?}\nrust: {}",
            plain(&r.stdout)
        );
    }
}

/// The layout of the status line is the Rust build's own: the two texts differ, and that is on
/// the approved list with its reason.
#[test]
fn the_layout_of_the_status_line_differs_on_purpose() {
    let Some(go) = go_binary() else { return };
    assert!(
        approved("layout_of_the_status_line"),
        "not on the approved list"
    );
    let payload = r#"{"model":{"display_name":"Layout Probe"}}"#;
    let g = run(&go, &["statusline"], payload);
    let r = run(&PathBuf::from(RUST), &["statusline"], payload);
    assert_ne!(
        plain(&g.stdout),
        plain(&r.stdout),
        "the texts are equal now: remove the entry from the approved list"
    );
}

/// A Go-generation hook call: Go answers an allow decision, the Rust build ignores it and prints
/// nothing on stdout. Both exit 0 and neither blocks.
#[test]
fn a_go_hook_call_is_ignored_by_the_rust_build_on_purpose() {
    let Some(go) = go_binary() else { return };
    assert!(approved("hook_call_pre_tool"), "not on the approved list");
    let payload = r#"{"session_id":"s1","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls"}}"#;
    let g = run(&go, &["hook", "pre-tool"], payload);
    let r = run(&PathBuf::from(RUST), &["hook", "pre-tool"], payload);
    assert_eq!(g.code, Some(0), "the Go build exits 0: {}", g.stderr);
    assert_eq!(r.code, Some(0), "the Rust build exits 0: {}", r.stderr);
    let answer: Value = serde_json::from_str(g.stdout.trim()).expect("the Go build answers JSON");
    assert_eq!(
        answer["hookSpecificOutput"]["permissionDecision"], "allow",
        "{}",
        g.stdout
    );
    assert_eq!(
        r.stdout.trim(),
        "",
        "the Rust build prints nothing on stdout"
    );
    assert!(
        r.stderr.contains("ignored a Go-generation hook call"),
        "{}",
        r.stderr
    );
}

/// Without the Go build the tests print the reason and pass: the lookup answers the reason for
/// a path that is no file, for no value at all, and finds a file given by the variable or by
/// the home folder.
#[test]
fn an_absent_go_build_skips_with_a_reason() {
    let tmp = TempDir::new().unwrap();
    let home = tmp.path().join("home");
    let missing = tmp.path().join("no-such-binary");
    let reason = locate(Some(missing.clone()), Some(home.clone())).unwrap_err();
    assert!(
        reason.contains("absent") && reason.contains("MOAI_GO_BIN"),
        "{reason}"
    );
    assert!(locate(None, Some(home.clone()))
        .unwrap_err()
        .contains("absent"));
    assert!(locate(None, None).unwrap_err().contains("no MOAI_GO_BIN"));
    let bin = home.join("go/bin/moai.old");
    fs::create_dir_all(bin.parent().unwrap()).unwrap();
    fs::write(&bin, "").unwrap();
    assert_eq!(locate(None, Some(home.clone())).unwrap(), bin);
    let given = tmp.path().join("given");
    fs::write(&given, "").unwrap();
    assert_eq!(locate(Some(given.clone()), Some(home)).unwrap(), given);
}
