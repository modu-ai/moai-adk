//! #2383 AC-2 / AC-3: with no `moai` on PATH, the Claude Code status line renders the same
//! bytes. A decoy `moai` that records every call shows that nothing looks the name up: the
//! outputs with and without it are identical and its log stays empty.
//!
//! #3004 P4-2: the generated hooks now call `moai`, so they run with the Rust `moai` on PATH and
//! never block a harmless tool call; a wrong (Go-like) `moai` first on PATH makes each hook fail
//! as a non-blocking error (exit 1), which is why `hooks-apply` refuses to write them there.
//! Temporary HOME and project only.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

/// This repository's `statusLine.command` (`.claude/settings.json`, leader switch #2286
/// 6817594847): the Rust build by absolute path, no PATH lookup.
const STATUS_LINE_COMMAND: &str = "\"$HOME/.moai-cli/path/moai-cli\" statusline";

struct Fixture {
    _dir: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
    /// Holds only `moai-cli`.
    bin: PathBuf,
    /// Holds the Rust build as `moai` (what the generated hooks call).
    rust: PathBuf,
    /// Holds a `moai` that appends its argv to `decoy.log` and exits 1.
    decoy: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let dir = tempfile::TempDir::new().unwrap();
        let root = fs::canonicalize(dir.path()).unwrap();
        let home = root.join("home");
        let project = root.join("project");
        let bin = root.join("bin");
        let decoy = root.join("decoy");
        for d in [&home, &project, &bin, &decoy] {
            fs::create_dir_all(d).unwrap();
        }
        let cli = Path::new(common::MOAI_CLI_BIN);
        std::os::unix::fs::symlink(cli, bin.join("moai-cli")).unwrap();
        let rust = root.join("rust");
        fs::create_dir_all(&rust).unwrap();
        std::os::unix::fs::symlink(cli, rust.join("moai")).unwrap();
        fs::create_dir_all(home.join(".moai-cli/path")).unwrap();
        std::os::unix::fs::symlink(cli, home.join(".moai-cli/path/moai-cli")).unwrap();
        common::write_executable(
            &decoy.join("moai"),
            &format!(
                "#!/bin/sh\necho \"moai $*\" >> '{}'\nexit 1\n",
                root.join("decoy.log").display()
            ),
        );
        common::git::git_ok(&project, &["init", "-q", "-b", "main"]);
        fs::write(project.join("a.txt"), "a\n").unwrap();
        common::git::git_ok(&project, &["add", "a.txt"]);
        common::git::git_ok(
            &project,
            &[
                "-c",
                "user.name=t",
                "-c",
                "user.email=t@example.invalid",
                "commit",
                "-q",
                "-m",
                "init",
            ],
        );
        Fixture {
            _dir: dir,
            home,
            project,
            bin,
            rust,
            decoy,
        }
    }

    fn decoy_log(&self) -> String {
        fs::read_to_string(self.decoy.parent().unwrap().join("decoy.log")).unwrap_or_default()
    }

    /// PATH without any `moai`, or with the decoy first.
    fn path(&self, with_decoy: bool) -> String {
        let base = format!("{}:/usr/bin:/bin", self.bin.display());
        if with_decoy {
            format!("{}:{base}", self.decoy.display())
        } else {
            base
        }
    }

    /// PATH with the Rust `moai` (what a hook needs), or with the decoy `moai` before it.
    fn hook_path(&self, with_decoy: bool) -> String {
        let base = format!("{}:{}", self.rust.display(), self.path(false));
        if with_decoy {
            format!("{}:{base}", self.decoy.display())
        } else {
            base
        }
    }

    /// Runs `command` through `sh -c` the way a host runs a command hook or status line:
    /// cwd = the project, the given PATH, a clean environment. `(exit code, stdout)`.
    fn sh(&self, command: &str, path: &str, stdin: &[u8]) -> (i32, String) {
        let mut child = Command::new("/bin/sh")
            .args(["-c", command])
            .current_dir(&self.project)
            .clean_env()
            .env("PATH", path)
            .env("HOME", &self.home)
            .env("MOAI_HOME", self.home.join(".moai"))
            .env("NO_COLOR", "1")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, stdin);
        let out = child.wait_with_output().unwrap();
        (
            out.status.code().unwrap_or(-1),
            String::from_utf8_lossy(&out.stdout).into_owned(),
        )
    }
}

/// Neither way of saying no in a Claude Code command hook: exit 2, or a JSON answer that
/// denies (PreToolUse) or blocks (PostToolUse / Stop).
fn blocks(code: i32, stdout: &str) -> bool {
    if code == 2 {
        return true;
    }
    let Ok(v) = serde_json::from_str::<Value>(stdout.trim()) else {
        return false;
    };
    v["decision"] == "block" || v["hookSpecificOutput"]["permissionDecision"] == "deny"
}

#[test]
fn the_status_line_renders_the_same_bytes_without_moai_on_path() {
    let f = Fixture::new();
    let stdin = json!({
        "model": {"display_name": "Opus"},
        "workspace": {"current_dir": f.project, "project_dir": f.project},
    })
    .to_string();
    // A large render budget keeps the git segment independent of machine load; the command
    // in front of it is the settings value byte for byte.
    let command = format!("{STATUS_LINE_COMMAND} --render-budget-ms 3600000");
    let without = f.sh(&command, &f.path(false), stdin.as_bytes());
    let with_decoy = f.sh(&command, &f.path(true), stdin.as_bytes());
    assert_eq!(without.0, 0);
    assert!(without.1.contains("Opus"), "rendered: {:?}", without.1);
    assert!(
        without.1.contains("main"),
        "git branch shown: {:?}",
        without.1
    );
    assert_eq!(without, with_decoy);
    assert_eq!(f.decoy_log(), "", "nothing ran `moai`");
}

/// The stdin Claude Code gives each event, for a harmless tool call.
fn event_input(event: &str, tool: &str, project: &Path) -> Value {
    let mut v = json!({"session_id": "s", "hook_event_name": event, "cwd": project,
                       "transcript_path": project.join("t.jsonl")});
    match (event, tool) {
        ("SessionStart", _) => v["source"] = json!("startup"),
        ("Stop", _) => v["stop_hook_active"] = json!(false),
        (_, "Bash") => {
            v["tool_name"] = json!("Bash");
            v["tool_input"] = json!({"command": "ls"});
        }
        (_, _) => {
            v["tool_name"] = json!(tool);
            v["tool_input"] = json!({"file_path": project.join("b.txt"), "content": "b\n"});
        }
    }
    if event == "PostToolUse" {
        v["tool_response"] = json!({"success": true});
    }
    v
}

#[test]
fn generated_hooks_run_and_never_block_a_harmless_call_without_moai_on_path() {
    let f = Fixture::new();
    let out = Command::new(common::MOAI_CLI_BIN)
        .args(["hooks-plan", "--host", "claude"])
        .arg(&f.project)
        .clean_env()
        .env("HOME", &f.home)
        .env("PATH", f.hook_path(false))
        .output()
        .unwrap();
    let plan: Value = serde_json::from_slice(&out.stdout).unwrap();
    let hooks = plan["result"]["hooks"].as_array().expect("hooks").clone();
    // #2936 R9 S1: SessionStart, two PreToolUse groups (claim guard, preflight), PostToolUse;
    // plus the security-scan PostToolUse group (SPEC-20261120-security-scan-wire).
    assert_eq!(hooks.len(), 5, "{plan}");
    let mut preflight = None;
    for h in &hooks {
        let command = h["command"].as_str().unwrap();
        assert!(command.starts_with("moai "), "{command}");
        let event = h["event"].as_str().unwrap();
        let tools: Vec<&str> = match h["matcher"].as_str() {
            Some(m) => m.split('|').collect(),
            None => vec!["Bash"],
        };
        for tool in tools {
            let input = event_input(event, tool, &f.project).to_string();
            let without = f.sh(command, &f.hook_path(false), input.as_bytes());
            // Exit 0, not 127: the hook's binary ran (a missing command would not block
            // either, and would prove nothing).
            assert_eq!(without.0, 0, "{event} {tool} `{command}`: {:?}", without.1);
            assert!(
                !blocks(without.0, &without.1),
                "{event} {tool}: {:?}",
                without.1
            );
            // A Go-like `moai` first on PATH receives the hook, exits 1 and so switches the
            // guard off without blocking the call: the case the apply gate exists for.
            let with_decoy = f.sh(command, &f.hook_path(true), input.as_bytes());
            assert_eq!(
                with_decoy.0, 1,
                "{event} {tool} `{command}`: {with_decoy:?}"
            );
            assert!(!blocks(with_decoy.0, &with_decoy.1), "{event} {tool}");
        }
        if command.contains("--event tool-preflight") {
            preflight = Some(command.to_string());
        }
    }
    assert!(
        f.decoy_log().contains("moai claim-verify . --hook"),
        "the decoy received the guard: {:?}",
        f.decoy_log()
    );
    // The guards really ran: the same preflight command still denies a dangerous call.
    let mut danger = event_input("PreToolUse", "Bash", &f.project);
    danger["tool_input"]["command"] = json!("rm -rf /");
    let r = f.sh(
        &preflight.expect("a preflight hook"),
        &f.hook_path(false),
        danger.to_string().as_bytes(),
    );
    assert!(blocks(r.0, &r.1), "rm -rf / allowed: {r:?}");
}
