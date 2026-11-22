//! SPEC-20261122-check-light-stop: the repository's build config turns incremental builds off
//! (REQ-7), the self check's default base is the local integration branch (REQ-8) and the
//! operation page of `check` describes the stop before the heavy steps (REQ-10). Everything
//! runs in temporary folders or reads files of this repository; the real HOME is never used.
mod common;
use common::git::{git_ok, git_raw};
use common::CleanEnv;
use std::{fs, path::Path, path::PathBuf, process::Command};

fn repo_path(rel: &str) -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)
}

fn repo_text(rel: &str) -> String {
    fs::read_to_string(repo_path(rel)).unwrap()
}

/// REQ-7: `.cargo/config.toml` says `[build] incremental = false` (a caller's `CARGO_INCREMENTAL`
/// still decides for its run, which cargo's own rule gives).
#[test]
fn the_cargo_config_turns_incremental_off() {
    let doc: toml_edit::DocumentMut = repo_text(".cargo/config.toml")
        .parse()
        .expect("the cargo config is TOML");
    let incremental = doc
        .get("build")
        .and_then(|b| b.get("incremental"))
        .and_then(|v| v.as_bool());
    assert_eq!(incremental, Some(false), "{doc}");
    // The two settings that were there before stay.
    let env = doc.get("env").expect("the [env] table stays");
    assert!(env.get("RUST_MIN_STACK").is_some());
    assert!(env.get("DEVELOPER_DIR").is_some());
}

/// A PATH with the real `git` and the system tools only.
fn path_with_git() -> String {
    let dir = Command::new("sh")
        .args(["-c", "dirname \"$(command -v git)\""])
        .output()
        .unwrap();
    format!(
        "{}:/usr/bin:/bin",
        String::from_utf8_lossy(&dir.stdout).trim()
    )
}

/// REQ-8: in a repository whose only integration branch is the local `v2026.11.0` and that has
/// no `origin`, the self check without `SELF_CHECK_BASE` runs its diff check against that
/// branch: a trailing space committed on top of it fails the check (exit 2 of `git diff
/// --check`), where an unresolved base would make it `not_run` (REQ-9).
#[test]
fn the_self_check_base_is_the_local_integration_branch() {
    let dir = tempfile::tempdir().unwrap();
    let p = dir.path();
    git_ok(p, &["init", "-q", "-b", "v2026.11.0"]);
    git_ok(p, &["config", "user.email", "t@example.invalid"]);
    git_ok(p, &["config", "user.name", "t"]);
    fs::create_dir_all(p.join("tools")).unwrap();
    fs::copy(
        repo_path("tools/self_check.sh"),
        p.join("tools/self_check.sh"),
    )
    .unwrap();
    git_ok(p, &["add", "-A"]);
    git_ok(p, &["commit", "-q", "-m", "base"]);
    // The change sits on a branch of its own, so the integration branch is the merge base.
    git_ok(p, &["checkout", "-q", "-b", "feature"]);
    fs::write(p.join("x.txt"), "a b \n").unwrap();
    git_ok(p, &["add", "-A"]);
    git_ok(p, &["commit", "-q", "-m", "change"]);
    assert!(
        !git_raw(p, &["remote"]).stdout.contains("origin"),
        "the fixture has no origin"
    );

    let out = Command::new("bash")
        .args(["tools/self_check.sh", "--pre-ready"])
        .current_dir(p)
        .clean_env()
        .env("PATH", path_with_git())
        .env("HOME", p)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env_remove("SELF_CHECK_BASE")
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).to_string();
    let line = text
        .lines()
        .find(|l| l.starts_with("self-check: diff-check "))
        .unwrap_or_else(|| panic!("no diff-check line: {text}"));
    assert!(line.starts_with("self-check: diff-check exit="), "{line}");
    assert!(!line.contains("not_run"), "{line}");
    assert!(!line.contains("exit=0"), "{line}");
    assert!(
        !text.contains("origin/v2026.11.0"),
        "the default no longer names the remote branch: {text}"
    );
}

/// REQ-10: the page names the stop, its summary word and the two kinds of step.
#[test]
fn the_operation_page_describes_the_stop() {
    let page = repo_text("docs/design-review/appendix/operations/check.md");
    for need in [
        "stopped_at",
        "가벼운 단계",
        "무거운 단계",
        "SPEC-20261122-check-light-stop",
    ] {
        assert!(page.contains(need), "the page lacks {need:?}");
    }
}
