//! The terminal speaks English whatever the language settings say (Issue #2825 R1, user
//! decision 6878265192: the CLI's human output is English only; the user's language belongs to
//! the LLM session).
//!
//! The S4 summary, the S5 progress lines, the `key-remove` confirm screen and the S3 error frame
//! are compared with the English goldens under `tests/fixtures/ui/en/`, with `MOAI_LANG=ko`, a
//! Korean locale and a profile `user_chat: ko` as well: every run gives the English bytes, on a
//! terminal and on a pipe, and the stdout envelope is the same bytes. Every run uses a temporary
//! HOME (under `/tmp`, short enough that the 60-column title never cuts it, so the golden's
//! `<home>` stands for the whole path on macOS and Linux alike). Moved here from the Python
//! module `ui_korean_tty_test.py` (SPEC-20261120-python-binary-tests-to-rust).

#![cfg(unix)]

mod common;
use common::pty::{env_with, piped_with, stderr_on_terminal, Env, Seen, Terminal};
use common::screens::strip_colors;
use std::{
    fs,
    os::unix::fs::{symlink, PermissionsExt as _},
    path::{Path, PathBuf},
};
use tempfile::TempDir;

const BIN: &str = common::MOAI_CLI_BIN;
const PACKAGE: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/skills/moai");
const GOLDEN: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/tests/fixtures/ui");
const KOREAN: [(&str, &str); 1] = [("MOAI_LANG", "ko")];
const KOREAN_LOCALES: [(&str, &str); 2] = [("LANG", "ko_KR.UTF-8"), ("LC_ALL", "ko_KR.UTF-8")];

fn golden(name: &str) -> String {
    fs::read_to_string(Path::new(GOLDEN).join("en").join(format!("{name}.golden"))).unwrap()
}

struct Fx {
    tmp: TempDir,
    home: PathBuf,
    work: PathBuf,
}

fn fx() -> Fx {
    let tmp = TempDir::new_in("/tmp").unwrap();
    let home = tmp.path().join("home");
    let work = tmp.path().join("work");
    fs::create_dir_all(&home).unwrap();
    fs::create_dir_all(&work).unwrap();
    Fx { tmp, home, work }
}

fn env_for(home: &Path, extra: &[(&str, &str)]) -> Env {
    env_with(
        &[
            ("PATH", "/usr/bin:/bin"),
            ("HOME", home.to_str().unwrap()),
            ("LANG", "en_US.UTF-8"),
            ("TERM", "xterm-256color"),
            ("COLUMNS", "60"),
            ("NO_COLOR", "1"),
        ],
        extra,
    )
}

/// `text` with each temporary path as `<home>` and every digest as `sha256:<digest>`.
fn normal(text: &str, paths: &[&Path]) -> String {
    let mut names: Vec<String> = Vec::new();
    for p in paths {
        names.push(fs::canonicalize(p).unwrap().display().to_string());
        names.push(p.display().to_string());
    }
    names.sort_by_key(|n| std::cmp::Reverse(n.len()));
    names.dedup();
    let mut text = text.to_string();
    for n in &names {
        text = text.replace(n.as_str(), "<home>");
    }
    let mut out = String::new();
    let mut rest = text.as_str();
    while let Some(at) = rest.find("sha256:") {
        let (head, tail) = rest.split_at(at + "sha256:".len());
        out.push_str(head);
        let hex = tail
            .bytes()
            .take_while(|b| b.is_ascii_hexdigit() && !b.is_ascii_uppercase());
        if hex.count() >= 64 {
            out.push_str("<digest>");
            rest = &tail[64..];
        } else {
            rest = tail;
        }
    }
    out.push_str(rest);
    out
}

fn hangul(text: &str) -> bool {
    text.chars().any(|c| ('가'..='힣').contains(&c))
}

fn json(text: &str) -> serde_json::Value {
    serde_json::from_str(text).unwrap_or_else(|e| panic!("not JSON ({e}): {text:?}"))
}

/// `home-routing-plan` with stderr on a terminal when `tty`; stdout is a pipe.
fn capture_summary(home: &Path, cwd: &Path, tty: bool, extra: &[(&str, &str)]) -> Seen {
    let args = ["home-routing-plan", home.to_str().unwrap(), PACKAGE];
    let env = env_for(home, extra);
    if tty {
        stderr_on_terminal(BIN, cwd, &env, &args)
    } else {
        piped_with(BIN, cwd, &env, &args, None)
    }
}

/// `(stdout, terminal text)` of `key-remove` answered `n` on a terminal, after a `key-set`.
fn capture_key_refusal(home: &Path, work: &Path, extra: &[(&str, &str)]) -> (String, String) {
    let mut all = vec![("LANG", "C"), ("TERM", "xterm")];
    all.extend_from_slice(extra);
    let env = env_for(home, &all);
    let set = piped_with(
        BIN,
        work,
        &env,
        &["key-set", "glm", "--shared"],
        Some("not-a-real-key\n"),
    );
    assert_eq!(set.code, Some(0), "{}", set.stderr);
    let mut term = Terminal::new(BIN)
        .args(["key-remove", "glm", "--shared"])
        .cwd(work)
        .stdout_not_a_terminal()
        .answer(b"n\n");
    for (k, v) in &env {
        term = term.env(k, v);
    }
    let s = term.run();
    assert_eq!(s.code, Some(0), "{}", s.text);
    (s.stdout, strip_colors(&s.text))
}

/// The progress lines `index-build` writes on a terminal.
fn capture_index_stages(root: &Path, home: &Path, extra: &[(&str, &str)]) -> String {
    let mut all = vec![("LANG", "C"), ("TERM", "xterm")];
    all.extend_from_slice(extra);
    let mut term = Terminal::new(BIN)
        .args(["index-build", root.to_str().unwrap()])
        .cwd(root)
        .stdout_not_a_terminal()
        .stdin_text("{\"protocol_version\":1}");
    for (k, v) in env_for(home, &all) {
        term = term.env(k, v);
    }
    let s = term.run();
    let text = strip_colors(&s.text);
    assert_eq!(s.code, Some(0), "{text}");
    text.lines()
        .filter(|l| l.starts_with("->"))
        .map(|l| format!("{l}\n"))
        .collect()
}

fn git_project(base: &Path, name: &str) -> PathBuf {
    let root = base.join(name);
    fs::create_dir_all(&root).unwrap();
    fs::write(root.join("a.md"), "# a\n").unwrap();
    common::git::git_in(&root, &["init", "-q"]);
    common::git::git_in(&root, &["add", "a.md"]);
    common::git::git_in(&root, &["commit", "-qm", "a"]);
    root
}

/// The summary's stderr with colors and temporary paths normalized, after checking it exited 0.
fn summary(f: &Fx, extra: &[(&str, &str)]) -> (String, String) {
    let s = capture_summary(&f.home, &f.home, true, extra);
    assert_eq!(s.code, Some(0), "{}", s.stdout);
    (s.stdout, normal(&strip_colors(&s.stderr), &[&f.home]))
}

// ---- the summary and the confirm and progress screens ----

#[test]
fn the_english_summary_keeps_its_bytes() {
    let f = fx();
    assert_eq!(summary(&f, &[]).1, golden("home-routing-plan"));
}

#[test]
fn moai_lang_ko_keeps_the_english_summary_and_envelope() {
    let f = fx();
    let (en_out, _) = summary(&f, &[]);
    let (ko_out, err) = summary(&f, &KOREAN);
    assert_eq!(err, golden("home-routing-plan"));
    assert_eq!(
        ko_out, en_out,
        "the stdout envelope does not follow the language"
    );
}

#[test]
fn a_korean_locale_keeps_the_english_summary() {
    let f = fx();
    for extra in [
        [("LANG", "ko_KR.UTF-8")],
        [("LC_ALL", "ko_KR.UTF-8")],
        [("LC_MESSAGES", "ko_KR.UTF-8")],
    ] {
        assert_eq!(
            summary(&f, &extra).1,
            golden("home-routing-plan"),
            "{extra:?}"
        );
    }
}

#[test]
fn a_korean_project_profile_keeps_the_english_summary() {
    let f = fx();
    let project = f.tmp.path().join("proj");
    fs::create_dir_all(&project).unwrap();
    // The first `moai cc` in a folder prepares it (#2935 L3: `moai init` is gone); a stand-in
    // `claude` that exits at once takes the session's place.
    let fake = f.tmp.path().join("fakebin");
    fs::create_dir_all(&fake).unwrap();
    fs::write(fake.join("claude"), "#!/bin/sh\nexit 0\n").unwrap();
    fs::set_permissions(fake.join("claude"), fs::Permissions::from_mode(0o755)).unwrap();
    let path = format!("{}:/usr/bin:/bin", fake.display());
    let cc = piped_with(
        BIN,
        &project,
        &env_for(&f.home, &[("PATH", path.as_str())]),
        &["cc"],
        Some(""),
    );
    assert_eq!(cc.code, Some(0), "{}", cc.stderr);
    // The profile lives under a hidden state folder; walk instead of globbing.
    let mut files = Vec::new();
    let mut todo = vec![project.clone()];
    while let Some(dir) = todo.pop() {
        for entry in fs::read_dir(&dir).unwrap() {
            let p = entry.unwrap().path();
            if p.is_dir() {
                todo.push(p);
            } else if p.file_name().is_some_and(|n| n == "project.yaml") {
                files.push(p);
            }
        }
    }
    assert_eq!(files.len(), 1, "{files:?}");
    let mut profile = json(&fs::read_to_string(&files[0]).unwrap());
    profile["language"] = serde_json::json!({ "user_chat": "ko" });
    fs::write(&files[0], serde_json::to_string(&profile).unwrap()).unwrap();
    let s = capture_summary(&f.home, &project, true, &[]);
    assert_eq!(s.code, Some(0), "{}", s.stdout);
    assert_eq!(
        normal(&strip_colors(&s.stderr), &[&f.home]),
        golden("home-routing-plan")
    );
}

#[test]
fn a_pipe_stays_english_whatever_the_language_says() {
    let f = fx();
    let (tty_out, _) = summary(&f, &[]);
    for extra in [KOREAN.to_vec(), KOREAN_LOCALES.to_vec()] {
        let s = capture_summary(&f.home, &f.home, false, &extra);
        assert_eq!(s.code, Some(0), "{}", s.stdout);
        assert_eq!(s.stderr, "", "{extra:?}: nothing human on a pipe");
        assert_eq!(s.stdout, tty_out, "{extra:?}: the same envelope bytes");
    }
}

#[test]
fn plain_output_under_a_korean_setting_is_english() {
    let f = fx();
    let (_, err) = summary(&f, &[("MOAI_OUTPUT", "plain"), ("MOAI_LANG", "ko")]);
    assert!(!err.contains('\u{1b}'), "{err:?}");
    assert!(err.contains("ACTION"), "{err}");
    assert!(err.contains("-> next:"), "{err}");
    assert!(!hangul(&err), "{err}");
}

#[test]
fn the_key_remove_screen_is_english_under_a_korean_setting() {
    let f = fx();
    let want = golden("key-remove-refused");
    let (out, term) = capture_key_refusal(&f.home, &f.work, &[]);
    assert_eq!(normal(&term, &[&f.home]), want);
    let (out_ko, term_ko) = capture_key_refusal(&f.home, &f.work, &KOREAN);
    assert_eq!(normal(&term_ko, &[&f.home]), want);
    assert_eq!(out_ko, out, "the refusal answers with the same envelope");
    // The key is still there (both answers were no): a pipe never asks, in any language, and
    // answers with the same preview envelope.
    let piped = piped_with(
        BIN,
        &f.work,
        &env_for(&f.home, &[("LANG", "C"), ("MOAI_LANG", "ko")]),
        &["key-remove", "glm", "--shared"],
        None,
    );
    assert_eq!(piped.code, Some(0), "{}", piped.stderr);
    assert_eq!(json(&piped.stdout)["result"]["exists"], true);
    assert!(!piped.stderr.contains("[y/N]"), "{}", piped.stderr);
    assert!(!piped.stderr.contains("지울까요"), "{}", piped.stderr);
    assert_eq!(piped.stdout, out);
    // `MOAI_OUTPUT=quiet` still asks (without it the answer is no), in the same language.
    let (_, quiet) = capture_key_refusal(
        &f.home,
        &f.work,
        &[("MOAI_OUTPUT", "quiet"), ("MOAI_LANG", "ko")],
    );
    assert_eq!(normal(&quiet, &[&f.home]), want);
}

#[test]
fn the_index_stages_are_english_under_a_korean_setting() {
    let f = fx();
    let want = golden("index-build-stages");
    let a = git_project(f.tmp.path(), "a");
    assert_eq!(capture_index_stages(&a, &f.home, &[]), want);
    let b = git_project(f.tmp.path(), "b");
    assert_eq!(capture_index_stages(&b, &f.home, &KOREAN), want);
}

// ---- the S3 error frame ----

// English labels, clap's first line and the English next step under every language setting
// (R1, user decision 6878265192).
const ERROR_CASES: [(&str, &[&str]); 4] = [
    ("usage-missing", &["project-plan"]),
    ("usage-unknown", &["project-plan", "--nope", "{proj}"]),
    (
        "usage-invalid",
        &[
            "lane-next",
            "--lane",
            "a",
            "--repo",
            "o/r",
            "--wait",
            "bogus",
        ],
    ),
    (
        "operation-permission",
        &[
            "hooks-apply",
            "--host",
            "claude",
            "{proj}",
            "--plan-digest",
            "x",
            "--permission-reference",
            "",
        ],
    ),
];

/// `ERROR_CASES[name]` with `NO_COLOR=1` and `TERM=xterm-256color`; stderr on a terminal when
/// `tty`.
fn capture_error(name: &str, home: &Path, tty: bool, exe: &str, extra: &[(&str, &str)]) -> Seen {
    let proj = home.join("p");
    fs::create_dir_all(&proj).unwrap();
    let proj = proj.to_str().unwrap().to_string();
    let cases = ERROR_CASES.iter().find(|(n, _)| *n == name).unwrap().1;
    let args: Vec<String> = cases.iter().map(|a| a.replace("{proj}", &proj)).collect();
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let mut all = vec![("NO_COLOR", "1")];
    all.extend_from_slice(extra);
    let env = env_with(
        &[
            ("PATH", "/usr/bin:/bin"),
            ("HOME", home.to_str().unwrap()),
            ("LANG", "en_US.UTF-8"),
            ("TERM", "xterm-256color"),
        ],
        &all,
    );
    if tty {
        stderr_on_terminal(exe, home, &env, &args)
    } else {
        piped_with(exe, home, &env, &args, None)
    }
}

fn frame(name: &str, home: &Path, extra: &[(&str, &str)]) -> Seen {
    let s = capture_error(name, home, true, BIN, extra);
    Seen {
        stderr: normal(&s.stderr, &[home]),
        ..s
    }
}

fn error_home() -> (TempDir, PathBuf) {
    let tmp = TempDir::new_in("/tmp").unwrap();
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    (tmp, home)
}

#[test]
fn english_frames_keep_their_bytes() {
    let (_tmp, home) = error_home();
    for (name, _) in ERROR_CASES {
        let s = frame(name, &home, &[]);
        assert_eq!(s.stderr, golden(&format!("error-{name}")), "{name}");
    }
}

#[test]
fn a_korean_setting_keeps_the_english_frames_and_envelope() {
    let (_tmp, home) = error_home();
    for extra in [KOREAN.to_vec(), vec![("LANG", "ko_KR.UTF-8")]] {
        for (name, _) in ERROR_CASES {
            let en = frame(name, &home, &[]);
            let ko = frame(name, &home, &extra);
            assert_eq!(
                ko.stderr,
                golden(&format!("error-{name}")),
                "{name} {extra:?}"
            );
            assert_eq!(
                (ko.code, &ko.stdout),
                (en.code, &en.stdout),
                "{name}: same exit, envelope"
            );
        }
    }
}

#[test]
fn a_pipe_keeps_the_english_bytes_in_any_language() {
    let (_tmp, home) = error_home();
    for (name, _) in ERROR_CASES {
        let plain = capture_error(name, &home, false, BIN, &[]);
        for extra in [KOREAN.to_vec(), KOREAN_LOCALES.to_vec()] {
            assert_eq!(
                capture_error(name, &home, false, BIN, &extra),
                plain,
                "{name} {extra:?}"
            );
        }
    }
}

// The next step names the program `moai` whatever name ran (#3004 P1), so an executable name with
// an OSC title and BEL never reaches the frame: no control byte, and the step reads as it does for
// any name.
#[test]
fn a_control_character_in_the_program_name_is_shown_as_text() {
    let (tmp, home) = error_home();
    let evil_dir = tmp.path().join("evil");
    fs::create_dir_all(&evil_dir).unwrap();
    let evil = evil_dir.join("mo\u{1b}]0;T\u{7}ai");
    symlink(BIN, &evil).unwrap();
    let s = capture_error(
        "usage-missing",
        &home,
        true,
        evil.to_str().unwrap(),
        &KOREAN,
    );
    assert!(!s.stderr.contains("\u{1b}]"), "{:?}", s.stderr);
    assert!(!s.stderr.contains('\u{7}'), "{:?}", s.stderr);
    assert!(!s.stderr.contains("\\x1b"), "{:?}", s.stderr);
    assert!(
        s.stderr
            .contains("run `moai project-plan --help` to see its arguments"),
        "{:?}",
        s.stderr
    );
}
