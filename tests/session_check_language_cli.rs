//! The session-check conversation line reads the HOME record only (the language card's delta
//! audit, B6): the value `language-set --home` (the installer) saved decides, a project
//! profile's earlier `language.user_chat` stays recorded data and never decides it, and with
//! nothing in HOME no language is named. The former pins lived in tests/hook_dispatch_cli.rs
//! (that file is another card's write scope); this file is the B6 regression set. Every case
//! runs the real binary in a temporary git repository with a temporary HOME.
#![cfg(unix)]

mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

const BIN: &str = crate::common::MOAI_BIN;

fn git(cwd: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

fn repo() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    git(&root, &["init", "-q"]);
    git(&root, &["config", "user.name", "fixture"]);
    git(&root, &["config", "user.email", "fixture@example.invalid"]);
    git(&root, &["config", "commit.gpgsign", "false"]);
    fs::write(root.join("a.txt"), "a\n").unwrap();
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    git(&root, &["add", "a.txt", ".gitignore"]);
    git(&root, &["commit", "-q", "-m", "init"]);
    (dir, root)
}

fn run(args: &[&str], cwd: &Path, stdin: &[u8]) -> (i32, String) {
    let mut child = Command::new(Path::new(BIN))
        .args(args)
        .current_dir(cwd)
        .env("HOME", cwd.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    use std::io::Write as _;
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(stdin);
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        ),
    )
}

fn session_context(root: &Path, source: &str) -> String {
    let input = json!({"hook_event_name": "SessionStart", "session_id": "s", "model": "m",
                       "source": source});
    let mut child = Command::new(Path::new(BIN))
        .args([
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "session-check",
        ])
        .current_dir(root)
        .env("HOME", root.join("home"))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    use std::io::Write as _;
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(input.to_string().as_bytes());
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .or_else(|| v["context"].as_str())
        .unwrap()
        .to_string()
}

fn profile(root: &Path, language: Value) {
    let mut p = json!({"schema_version": 2, "project_id": "demo", "purpose": "x",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]});
    p["language"] = language;
    fs::create_dir_all(root.join(".moai-cli")).unwrap();
    fs::write(
        root.join(".moai-cli/project.yaml"),
        serde_json::to_vec_pretty(&p).unwrap(),
    )
    .unwrap();
}

fn home_sentence(name: &str, code: &str) -> String {
    format!(
        "Conversation language: {name} (`{code}`), from the language recorded in HOME: talk \
         with the user in it and relay subagent reports to the user in it."
    )
}

/// B6 regression: the HOME record decides for every source; a profile's earlier `user_chat`
/// never outranks it (the old order let a later `language-set --home` miss existing sessions),
/// and with nothing in HOME nothing is named even when a profile holds a value.
#[test]
fn session_check_names_the_home_recorded_language_only() {
    let (_d, root) = repo();
    let sources = ["startup", "resume", "compact"];
    for source in sources {
        let ctx = session_context(&root, source);
        assert!(!ctx.contains("Conversation language"), "{source}: {ctx}");
    }
    // The installer's record, written the way the install does it.
    fs::create_dir_all(root.join("home/.moai-cli")).unwrap();
    let (code, out) = run(&["language-set", "--home", "ja"], &root, b"");
    assert_eq!(code, 0, "{out}");
    let home_line = home_sentence("Japanese", "ja");
    for source in sources {
        assert!(
            session_context(&root, source).contains(&home_line),
            "{source}"
        );
    }
    // A profile's stored user_chat stays data: the HOME record still decides.
    profile(&root, json!({"user_chat": "ko", "code_comments": "en"}));
    for source in sources {
        let ctx = session_context(&root, source);
        assert!(ctx.contains(&home_line), "{source}: {ctx}");
        assert!(!ctx.contains("Korean"), "{source}: {ctx}");
    }
    // A profile with a code outside the list, or one that does not read, changes nothing.
    profile(&root, json!({"user_chat": "fr"}));
    assert!(session_context(&root, "startup").contains(&home_line));
    fs::write(root.join(".moai-cli/project.yaml"), "{not yaml: [").unwrap();
    assert!(session_context(&root, "startup").contains(&home_line));
    // Nothing in HOME: no language is named even with a profile that holds one.
    fs::remove_dir_all(root.join("home")).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    profile(&root, json!({"user_chat": "ko"}));
    let ctx = session_context(&root, "compact");
    assert!(!ctx.contains("Conversation language"), "{ctx}");
    assert!(!ctx.contains("Korean"), "{ctx}");
}

/// The sentence names HOME as its only source: the profile's `language.user_chat` wording is
/// gone with the old order.
#[test]
fn session_check_sentence_names_home_as_its_source() {
    let (_d, root) = repo();
    fs::create_dir_all(root.join("home/.moai-cli")).unwrap();
    let (code, out) = run(&["language-set", "--home", "ko"], &root, b"");
    assert_eq!(code, 0, "{out}");
    let ctx = session_context(&root, "startup");
    assert!(ctx.contains("from the language recorded in HOME"), "{ctx}");
    assert!(!ctx.contains("the project's `language.user_chat`"), "{ctx}");
}
