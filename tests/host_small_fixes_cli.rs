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
    let (code, v) = check(tmp.path(), home.path(), &base);
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["ok"], json!(true), "{v}");
    let mut got = kinds(&v);
    got.sort();
    assert_eq!(
        got,
        vec![
            ("malformed_cloud_session".to_string(), false),
            ("malformed_cloud_session".to_string(), false),
            ("missing_session".to_string(), false),
            ("missing_session".to_string(), false),
            // Off a `spec/` branch a missing SPEC is named and does not fail either.
            ("missing_spec".to_string(), false),
            ("missing_spec".to_string(), false),
        ],
        "{v}"
    );
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
