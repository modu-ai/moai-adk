//! SPEC-20261122-host-small-fixes: three small fixes from the Claude Code docs. Slice 1 pins the
//! cloud session trailer in `moai trailer-check` (REQ-3, REQ-4, REQ-5, REQ-8) and the README's
//! section on how MoAI ADK relates to Anthropic and Claude Code (REQ-6, REQ-7). Temporary
//! repositories and HOMEs only; the READMEs are read from the tree.

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::path::{Path, PathBuf};
use std::process::Command;

const SPEC: &str = "SPEC-20261122-host-small-fixes";
const SESSION: &str = "428e289a-1c67-492e-a23b-a5aca15b4d1f:0123456789ab";
/// The shape a cloud session gives (cloud-environments.md): the claude.ai session link.
const LINK: &str = "https://claude.ai/code/session_011CUmZ7rWq2pXk9Lb3vT4dE";

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn moai(dir: &Path, home: &Path, args: &[&str]) -> (Option<i32>, Value) {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .clean_env()
        .env("HOME", home)
        .env("PATH", "/usr/bin:/bin")
        .env("LC_ALL", "C")
        .current_dir(dir)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.code(), v)
}

fn commit(dir: &Path, message: &str) {
    std::fs::write(dir.join("f.txt"), message).unwrap();
    common::git::git_ok(dir, &["add", "f.txt"]);
    common::git::git_ok(
        dir,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "commit",
            "-q",
            "-m",
            message,
        ],
    );
}

/// A repository on `branch` with one base commit; returns the base commit.
fn repo(dir: &Path, branch: &str) -> String {
    common::git::git_ok(dir, &["init", "-q", "-b", branch]);
    commit(dir, "base");
    common::git::git(dir, &["rev-parse", "HEAD"])
        .trim()
        .to_string()
}

/// Runs `trailer-check <base>..HEAD` with the card-number rule on from `base`.
fn check(dir: &Path, home: &Path, base: &str) -> (Option<i32>, Value) {
    moai(
        dir,
        home,
        &[
            "trailer-check",
            &format!("{base}..HEAD"),
            "--card-numbers-since",
            base,
        ],
    )
}

fn kinds(v: &Value) -> Vec<(String, bool)> {
    v["result"]["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["kind"].as_str().unwrap().to_string(),
                f["fails"].as_bool().unwrap(),
            )
        })
        .collect()
}

/// AC-2 (REQ-3, REQ-8): a commit with all of moai's trailers and a `Claude-Session` line passes
/// with no finding at all; the card-number rule finds nothing in the link.
#[test]
fn cloud_trailer_beside_moai_trailers_adds_no_finding() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let base = repo(tmp.path(), "spec/host-small-fixes-20261122");
    commit(
        tmp.path(),
        &format!(
            "feat: x\n\nMoai-Spec: {SPEC}\nMoai-Session: {SESSION}\nMoai-Verified: tests\nClaude-Session: {LINK}"
        ),
    );
    let (code, v) = check(tmp.path(), home.path(), &base);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
    assert_eq!(kinds(&v), Vec::<(String, bool)>::new(), "{v}");
}

/// AC-3 (REQ-4): no `Moai-Session` and one well-formed link: `cloud_session` names the link and
/// does not fail; `missing_session` is not reported.
#[test]
fn cloud_trailer_without_moai_session_names_the_cloud_session() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let base = repo(tmp.path(), "spec/host-small-fixes-20261122");
    commit(
        tmp.path(),
        &format!("feat: x\n\nMoai-Spec: {SPEC}\nMoai-Verified: none\nClaude-Session: {LINK}"),
    );
    let (code, v) = check(tmp.path(), home.path(), &base);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(kinds(&v), vec![("cloud_session".to_string(), false)], "{v}");
    let detail = v["result"]["findings"][0]["detail"].as_str().unwrap();
    assert!(detail.contains(LINK), "{detail}");
}

/// AC-4 (REQ-5): another shape and a doubled link each give `malformed_cloud_session` beside
/// `missing_session`, neither failing; on a branch that is not `spec/` the result stays ok.
#[test]
fn cloud_trailer_of_another_shape_or_doubled_is_named_without_failing() {
    let (tmp, home) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    let base = repo(tmp.path(), "main");
    commit(
        tmp.path(),
        "feat: odd\n\nMoai-Verified: none\nClaude-Session: https://example.invalid/session_1",
    );
    commit(
        tmp.path(),
        &format!(
            "feat: twice\n\nMoai-Verified: none\nClaude-Session: {LINK}\nClaude-Session: {LINK}"
        ),
    );
    // The link's start with no id, and an id holding a character other than a letter or digit.
    for odd in [
        "https://claude.ai/code/session_",
        "https://claude.ai/code/session_011CU-mZ7",
    ] {
        commit(
            tmp.path(),
            &format!("feat: {odd}\n\nMoai-Verified: none\nClaude-Session: {odd}"),
        );
    }
    let (code, v) = check(tmp.path(), home.path(), &base);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
    let mut got = kinds(&v);
    got.sort();
    let mut want = Vec::new();
    // Off a `spec/` branch a missing SPEC is named and does not fail either.
    for kind in ["malformed_cloud_session", "missing_session", "missing_spec"] {
        want.extend(std::iter::repeat_n((kind.to_string(), false), 4));
    }
    assert_eq!(got, want, "{v}");
}

/// The five statements of the README section (AC-5), each found by words every spelling of the
/// statement must hold, per README.
fn statements(korean: bool) -> [(&'static str, &'static [&'static str]); 5] {
    if korean {
        [
            ("runs Claude Code", &["Claude Code를 실행"]),
            (
                "not built, endorsed or partnered",
                &["Anthropic", "만들", "보증", "제휴"],
            ),
            (
                "no Claude Code or Anthropic name or logo",
                &["이름", "로고"],
            ),
            (
                "does not modify the executable",
                &["실행 파일", "바꾸지 않"],
            ),
            ("no credentials", &["자격 증명", "세션 토큰"]),
        ]
    } else {
        [
            ("runs Claude Code", &["runs Claude Code"]),
            (
                "not built, endorsed or partnered",
                &["Anthropic", "built", "endorsed", "partnered"],
            ),
            (
                "no Claude Code or Anthropic name or logo",
                &["name", "logo"],
            ),
            (
                "does not modify the executable",
                &["does not modify", "executable"],
            ),
            ("no credentials", &["credentials", "session tokens"]),
        ]
    }
}

/// The section under `heading` up to the next `## ` heading.
fn section<'a>(text: &'a str, heading: &str) -> Option<&'a str> {
    let start = text.find(&format!("\n{heading}\n"))? + 1;
    let body = &text[start + heading.len()..];
    Some(match body.find("\n## ") {
        Some(end) => &body[..end],
        None => body,
    })
}

/// AC-5 (REQ-6, REQ-7): each README holds one section on how MoAI ADK relates to Anthropic and
/// Claude Code, with its five statements; a missing one is named with its file.
#[test]
fn readme_states_how_moai_relates_to_anthropic() {
    for (file, heading, korean) in [
        (
            "README.md",
            "## Relation to Anthropic and Claude Code",
            false,
        ),
        ("README.ko.md", "## Anthropic·Claude Code와의 관계", true),
    ] {
        let text = std::fs::read_to_string(root().join(file)).unwrap();
        assert_eq!(
            text.matches(&format!("\n{heading}\n")).count(),
            1,
            "{file} must hold exactly one section {heading:?}"
        );
        let body = section(&text, heading).unwrap();
        for (name, words) in statements(korean) {
            for w in words {
                assert!(
                    body.contains(w),
                    "{file}: the statement {name:?} lacks {w:?} in {heading:?}"
                );
            }
        }
    }
}

/// The install command the official page gives for native Windows (setup.md); the refusal names it.
const INSTALL: &str = "irm https://claude.ai/install.ps1 | iex";

fn files(list: &[&str]) -> Vec<PathBuf> {
    list.iter().map(PathBuf::from).collect()
}

/// AC-1 (REQ-1, REQ-2): on Windows a name that ends in `.cmd` or `.bat` is refused whatever its
/// case or folder, even when a native `claude.exe` is also on the list, and the refusal names the
/// install command. The function takes the target, so this runs on every platform.
#[test]
fn windows_a_cmd_or_bat_name_is_refused_with_the_install_command() {
    let exe = files(&[r"C:\Users\u\.local\bin\claude.exe"]);
    for name in [
        "claude.cmd",
        "claude.BAT",
        r"C:\Users\u\AppData\Roaming\npm\claude.cmd",
        "/c/npm/claude.bat",
    ] {
        for list in [Vec::new(), exe.clone()] {
            let got = moai_cli::claude::choose_program(name, true, &list);
            assert!(
                got.is_err(),
                "{name} with {list:?} must be refused: {got:?}"
            );
            let msg = got.unwrap_err();
            assert!(msg.contains(INSTALL), "{name}: {msg}");
            assert!(
                msg.contains(name),
                "{name}: the refusal names the file: {msg}"
            );
        }
    }
}

/// AC-1 (REQ-1): a name that ends in `.exe` is started as it is, whatever else is on the list.
#[test]
fn windows_a_claude_exe_is_started_as_it_is() {
    let shims = files(&[r"C:\npm\claude.cmd", r"C:\npm\claude.bat"]);
    for name in [
        "claude.exe",
        r"C:\Users\u\.local\bin\claude.exe",
        "claude.EXE",
    ] {
        let got = moai_cli::claude::choose_program(name, true, &shims);
        assert_eq!(got, Ok(PathBuf::from(name)), "{name}");
    }
}

/// AC-1 (REQ-1): for the bare name `claude` the native executable on the list is chosen, never a
/// `.cmd` or `.bat` file, in whatever order the list holds them.
#[test]
fn windows_a_bare_name_picks_the_exe_and_never_the_shim() {
    let exe = r"C:\Users\u\.local\bin\claude.exe";
    for list in [
        files(&[r"C:\npm\claude.cmd", exe]),
        files(&[exe, r"C:\npm\claude.cmd"]),
        files(&[r"C:\npm\claude.bat", r"C:\npm\claude.cmd", exe]),
    ] {
        let got = moai_cli::claude::choose_program("claude", true, &list);
        assert_eq!(got, Ok(PathBuf::from(exe)), "{list:?}");
    }
}

/// AC-1 (REQ-2): when the only Claude Code a bare name finds is a `.cmd` or `.bat` file, moai
/// refuses to start it and names the file and the native install command.
#[test]
fn windows_a_bare_name_with_only_a_shim_is_refused() {
    for list in [
        files(&[r"C:\npm\claude.cmd"]),
        files(&[r"C:\npm\claude.bat"]),
        files(&[r"C:\a\claude.bat", r"C:\b\claude.cmd"]),
    ] {
        let got = moai_cli::claude::choose_program("claude", true, &list);
        assert!(got.is_err(), "{list:?} must be refused: {got:?}");
        let msg = got.unwrap_err();
        assert!(msg.contains(INSTALL), "{list:?}: {msg}");
        let first = list[0].to_string_lossy().into_owned();
        assert!(msg.contains(&first), "{list:?}: names the file: {msg}");
    }
}

/// AC-1: a bare name that resolves to no file is left to the start itself, whose error names it.
#[test]
fn windows_a_bare_name_with_no_file_is_left_to_the_start() {
    let got = moai_cli::claude::choose_program("claude", true, &[]);
    assert_eq!(got, Ok(PathBuf::from("claude")));
}

/// AC-1 and the card's second criterion: every other target keeps every name as it was given,
/// a `.cmd` name and the shims on the list included.
#[test]
fn windows_other_targets_keep_every_name_as_given() {
    let list = files(&["/x/claude.cmd", "/y/claude.bat", "/z/claude.exe"]);
    for name in [
        "claude",
        "claude.cmd",
        "claude.exe",
        "/usr/local/bin/claude",
    ] {
        let got = moai_cli::claude::choose_program(name, false, &list);
        assert_eq!(got, Ok(PathBuf::from(name)), "{name}");
    }
}

/// The list the function reads: for a bare name, each search folder in order, and in each folder
/// `.exe`, then `.cmd`, then `.bat`; a name with a folder part is read beside itself. Only files
/// that exist, and a longer name is not mistaken for the name.
#[test]
fn windows_candidates_follow_the_search_folders_and_the_endings() {
    let (a, b) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
    for (dir, name) in [
        (&a, "claude.cmd"),
        (&b, "claude.exe"),
        (&b, "claude.cmd"),
        (&b, "claude.bat"),
        (&b, "claudex.exe"),
    ] {
        std::fs::write(dir.path().join(name), b"").unwrap();
    }
    std::fs::create_dir(a.path().join("claude.exe")).unwrap();
    let dirs = vec![a.path().to_path_buf(), b.path().to_path_buf()];
    let got = moai_cli::claude::candidates_in(&dirs, "claude");
    assert_eq!(
        got,
        vec![
            a.path().join("claude.cmd"),
            b.path().join("claude.exe"),
            b.path().join("claude.cmd"),
            b.path().join("claude.bat"),
        ]
    );
    let beside = b.path().join("claude");
    let got = moai_cli::claude::candidates_in(&[], beside.to_str().unwrap());
    assert_eq!(
        got,
        vec![
            b.path().join("claude.exe"),
            b.path().join("claude.cmd"),
            b.path().join("claude.bat"),
        ]
    );
}

/// The Windows-only start, pinned under `cfg(windows)`: a named shim is refused by the real
/// start function too. Not run on this platform; see the card's READY.
#[cfg(windows)]
#[test]
fn windows_the_start_function_refuses_a_named_shim() {
    let got = moai_cli::claude::start_program("claude.cmd");
    assert!(got.is_err(), "{got:?}");
    assert!(got.unwrap_err().contains(INSTALL));
}

/// Off Windows the start function changes nothing (the card's second criterion).
#[cfg(not(windows))]
#[test]
fn windows_the_start_function_changes_nothing_off_windows() {
    for name in [
        "claude",
        "claude.cmd",
        "claude.exe",
        "/usr/local/bin/claude",
    ] {
        let got = moai_cli::claude::start_program(name);
        assert_eq!(got, Ok(PathBuf::from(name)), "{name}");
    }
}
