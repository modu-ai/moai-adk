//! SPEC-20261117-edit-diagnostics slice S1: `moai hook-dispatch --host claude --event
//! edit-diagnostics` runs the checker the user named in `[hooks.edit_diagnostics]` after a file
//! edit and answers with `hookSpecificOutput.additionalContext`; `hooks-plan` generates its hook
//! line only while the setting is enabled. Every case runs the real binary with a temporary HOME
//! and a temporary git project; the checker is a test script.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};
use std::time::{Duration, Instant};

/// The fake checker: `MODE` picks what it does; it touches `$MARK` when it runs; the edited
/// file is its last argument.
const CHECKER: &str = r#"#!/bin/sh
for f; do file="$f"; done
[ -n "$MARK" ] && touch "$MARK"
case "$MODE" in
  clean) exit 0 ;;
  diag)
    i=1
    while [ "$i" -le "${N:-3}" ]; do
      printf '%s:%s:7: error[E%s]: problem number %s\n' "$file" "$i" "$i" "$i"
      i=$((i + 1))
    done
    exit 1 ;;
  absolute) printf '%s/%s:2:1: warning: absolute path\n/elsewhere/x.rs:9:9: error: outside\n' "$PWD" "$file"; exit 1 ;;
  noform) echo "something went wrong"; exit 3 ;;
  sleep) sleep 30 & echo $! > "$PIDF"; sleep 30 ;;
  huge) head -c 2097152 /dev/zero | tr '\0' 'x'; exit 1 ;;
  nonutf8) printf '%s:1:1: error: bad \377\376 bytes\n' "$file"; exit 1 ;;
  secret) printf '%s:1:1: error: leaked token=abc123 here\n%s:2:1: error: plain\n' "$file" "$file"; exit 1 ;;
  slow) sleep 2; printf '%s:1:1: error: slow\n' "$file"; exit 1 ;;
  argv) for a; do printf 'ARG:%s\n' "$a"; done > "$ARGF"; exit 0 ;;
  control) printf '%s:1:1: error: bell\007 x \033[31mred\033[0m\r end\n' "$file"; exit 1 ;;
  linger) sleep 120 & echo $! > "$PIDF"; printf '%s:1:1: error: done\n' "$file"; exit 1 ;;
esac
exit 0
"#;

struct Fx {
    _dir: tempfile::TempDir,
    root: PathBuf,
    proj: PathBuf,
    home: PathBuf,
    checker: PathBuf,
}

fn fx() -> Fx {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    let (proj, home) = (root.join("proj"), root.join("home"));
    fs::create_dir_all(proj.join("src")).unwrap();
    fs::create_dir_all(proj.join(".moai-cli")).unwrap();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    common::git::git(&proj, &["init", "-q", "-b", "main"]);
    fs::write(proj.join("src/a.rs"), "fn main() {}\n").unwrap();
    let checker = root.join("checker.sh");
    common::write_executable(&checker, CHECKER);
    Fx {
        _dir: dir,
        root,
        proj,
        home,
        checker,
    }
}

impl Fx {
    fn config(&self, text: &str) {
        fs::write(self.home.join(".moai-cli/hook-config.toml"), text).unwrap();
    }

    /// An enabled setting with `extra` lines and a `rs` checker.
    fn enable(&self, extra: &str) {
        self.config(&format!(
            "[hooks.edit_diagnostics]\nenabled = true\n{extra}\ncheckers = {{ rs = [{:?}] }}\n",
            self.checker.to_str().unwrap()
        ));
    }

    fn command(&self, args: &[&str], env: &[(&str, &str)]) -> Command {
        let mut c = Command::new(common::MOAI_BIN);
        c.clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("MARK", self.root.join("ran"))
            .current_dir(&self.proj)
            .args(args);
        for (k, v) in env {
            c.env(k, v);
        }
        c
    }

    fn dispatch_as(&self, host: &str, stdin: &[u8], env: &[(&str, &str)]) -> Output {
        let mut child = self
            .command(
                &[
                    "hook-dispatch",
                    "--host",
                    host,
                    "--event",
                    "edit-diagnostics",
                ],
                env,
            )
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        use std::io::Write;
        child.stdin.take().unwrap().write_all(stdin).unwrap();
        child.wait_with_output().unwrap()
    }

    /// The answer to one edit of `file`: (exit, stdout).
    fn edit(&self, tool: &str, file: &str, env: &[(&str, &str)]) -> (i32, String) {
        let out = self.dispatch_as("claude", &input(tool, file, &self.proj), env);
        (
            out.status.code().unwrap_or(-1),
            String::from_utf8_lossy(&out.stdout).trim().to_string(),
        )
    }

    fn ran(&self) -> bool {
        let mark = self.root.join("ran");
        let ran = mark.exists();
        let _ = fs::remove_file(mark);
        ran
    }

    fn file(&self) -> String {
        self.proj.join("src/a.rs").display().to_string()
    }

    /// The hook entries `hooks-plan --host <host>` lists for edit-diagnostics.
    fn planned(&self, host: &str) -> Vec<Value> {
        let out = self
            .command(
                &["hooks-plan", self.proj.to_str().unwrap(), "--host", host],
                &[],
            )
            .output()
            .unwrap();
        assert!(out.status.success(), "{out:?}");
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        v["result"]["hooks"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|h| {
                h["command"]
                    .as_str()
                    .unwrap_or("")
                    .contains("edit-diagnostics")
            })
            .cloned()
            .collect()
    }

    fn engine_log(&self) -> String {
        fs::read_to_string(self.proj.join(".moai-cli/receipts/hook-engine.jsonl"))
            .unwrap_or_default()
    }
}

fn input(tool: &str, file: &str, cwd: &Path) -> Vec<u8> {
    json!({"session_id": "s", "hook_event_name": "PostToolUse", "tool_name": tool,
           "tool_use_id": "t1", "tool_input": {"file_path": file, "content": "x"},
           "tool_response": {"filePath": file}, "cwd": cwd})
    .to_string()
    .into_bytes()
}

fn context(stdout: &str) -> String {
    let v: Value = serde_json::from_str(stdout).unwrap_or_else(|e| panic!("{e}: {stdout}"));
    assert_eq!(
        v["hookSpecificOutput"]["hookEventName"], "PostToolUse",
        "{v}"
    );
    v["hookSpecificOutput"]["additionalContext"]
        .as_str()
        .unwrap()
        .to_string()
}

#[test]
fn off_without_the_setting_no_hook_line_and_no_run() {
    let f = fx();
    for config in [
        None,
        Some("[hooks.edit_diagnostics]\nenabled = false\ncheckers = { rs = [\"x\"] }\n"),
        Some("[hooks.outside_project_writes]\nenabled = true\n"),
    ] {
        if let Some(c) = config {
            f.config(c);
        }
        assert_eq!(f.planned("claude"), Vec::<Value>::new(), "{config:?}");
        assert_eq!(
            f.edit("Write", &f.file(), &[("MODE", "diag")]),
            (0, String::new())
        );
        assert!(!f.ran(), "{config:?}: the checker ran while off");
    }
}

#[test]
fn an_enabled_setting_generates_one_hook_line_for_claude_code_only() {
    let f = fx();
    f.enable("timeout_secs = 7");
    assert_eq!(
        f.planned("claude"),
        vec![
            json!({"event": "PostToolUse", "matcher": "Write|Edit|MultiEdit",
                    "command": "moai hook-dispatch --host claude --event edit-diagnostics",
                    "timeout": 12})
        ]
    );
    assert_eq!(f.planned("codex"), Vec::<Value>::new());
    // Codex and OMP do not run the duty in S1: an engine error, nothing checked.
    for host in ["codex", "omp"] {
        let out = f.dispatch_as(
            host,
            &input("Write", &f.file(), &f.proj),
            &[("MODE", "diag")],
        );
        assert_eq!(out.status.code(), Some(1), "{host}");
        assert!(out.stdout.is_empty());
    }
    assert!(!f.ran());
}

#[test]
fn a_malformed_setting_is_off_and_logged_and_its_bounds_hold() {
    let f = fx();
    for bad in [
        "mode = \"block\"",
        "timeout_secs = 0",
        "timeout_secs = 31",
        "max_lines = 0",
        "max_lines = 201",
        "max_chars = 199",
        "max_chars = 9001",
        "colour = true",
    ] {
        f.enable(bad);
        assert_eq!(f.planned("claude"), Vec::<Value>::new(), "{bad}");
        let before = f.engine_log().matches("invalid_edit_diagnostics").count();
        assert_eq!(
            f.edit("Write", &f.file(), &[("MODE", "diag")]),
            (0, String::new())
        );
        assert!(!f.ran(), "{bad}");
        assert_eq!(
            f.engine_log().matches("invalid_edit_diagnostics").count(),
            before + 1,
            "{bad}: one config_error line"
        );
    }
    for bad_checkers in [
        "checkers = { rs = \"x\" }",
        "checkers = { RS = [\"x\"] }",
        "checkers = { rs = [] }",
        "checkers = { rs = [\"\"] }",
    ] {
        f.config(&format!(
            "[hooks.edit_diagnostics]\nenabled = true\n{bad_checkers}\n"
        ));
        assert_eq!(f.planned("claude"), Vec::<Value>::new(), "{bad_checkers}");
    }
    f.config("[hooks.edit_diagnostics]\nenabled = \"yes\"\n");
    assert_eq!(f.edit("Write", &f.file(), &[]), (0, String::new()));
    assert!(f.engine_log().contains("invalid_edit_diagnostics"));
    // The edges of every bound are accepted.
    for good in [
        "timeout_secs = 1",
        "timeout_secs = 30",
        "max_lines = 1",
        "max_lines = 200",
        "max_chars = 200",
        "max_chars = 9000",
        "mode = \"context\"",
    ] {
        f.enable(good);
        assert_eq!(f.planned("claude").len(), 1, "{good}");
    }
    // An unusable file turns everything off with its own reason.
    fs::remove_file(f.home.join(".moai-cli/hook-config.toml")).unwrap();
    std::os::unix::fs::symlink(
        f.root.join("checker.sh"),
        f.home.join(".moai-cli/hook-config.toml"),
    )
    .unwrap();
    assert_eq!(
        f.edit("Write", &f.file(), &[("MODE", "diag")]),
        (0, String::new())
    );
    assert!(f.engine_log().contains("not_regular_file"));
    assert!(!f.ran());
}

#[test]
fn diagnostics_come_back_as_context_within_the_caps() {
    let f = fx();
    f.enable("");
    for tool in ["Write", "Edit", "MultiEdit"] {
        let (code, out) = f.edit(tool, &f.file(), &[("MODE", "diag")]);
        assert_eq!(code, 0);
        assert_eq!(
            context(&out),
            "MoAI edit-diagnostics (checker.sh, src/a.rs): 3 diagnostic line(s), exit 1\n\
             src/a.rs 1:7 error problem number 1\n\
             src/a.rs 2:7 error problem number 2\n\
             src/a.rs 3:7 error problem number 3",
            "{tool}"
        );
    }
    // A clean run says nothing; a Bash call or an unmapped extension runs nothing.
    assert_eq!(
        f.edit("Write", &f.file(), &[("MODE", "clean")]),
        (0, String::new())
    );
    assert!(f.ran());
    fs::write(f.proj.join("src/b.py"), "x\n").unwrap();
    let py = f.proj.join("src/b.py").display().to_string();
    assert_eq!(
        f.edit("Write", &py, &[("MODE", "diag")]),
        (0, String::new())
    );
    assert_eq!(
        f.edit("Bash", &f.file(), &[("MODE", "diag")]),
        (0, String::new())
    );
    assert!(!f.ran());
    // max_lines at its edge: N lines shown, N + 1 says one is not shown.
    f.enable("max_lines = 2");
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "diag"), ("N", "2")]);
    assert!(!context(&out).contains("not shown"), "{out}");
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "diag"), ("N", "3")]);
    let text = context(&out);
    assert!(
        text.contains("3 diagnostic line(s), exit 1; 1 not shown"),
        "{text}"
    );
    assert_eq!(text.lines().count(), 3);
    // max_chars bounds the whole text, summary included.
    f.enable("max_chars = 200");
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "diag"), ("N", "50")]);
    let text = context(&out);
    assert!(text.chars().count() <= 200, "{}", text.chars().count());
    assert!(text.contains("not shown"), "{text}");
    // Paths: project-relative; an absolute path outside the project is not shown.
    f.enable("");
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "absolute")]);
    let text = context(&out);
    assert!(
        text.contains("\nsrc/a.rs 2:1 warning absolute path"),
        "{text}"
    );
    assert!(
        text.contains("<outside the project> 9:9 error outside"),
        "{text}"
    );
    assert!(!text.contains(f.root.to_str().unwrap()), "{text}");
    // Output in no known form: the summary says so.
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "noform")]);
    assert_eq!(
        context(&out),
        "MoAI edit-diagnostics (checker.sh, src/a.rs): 0 diagnostic line(s), exit 3; no line in the path:line:col form"
    );
}

#[test]
fn files_outside_the_project_or_behind_a_link_are_not_checked() {
    let f = fx();
    f.enable("");
    let outside = f.root.join("other.rs");
    fs::write(&outside, "x\n").unwrap();
    std::os::unix::fs::symlink(&outside, f.proj.join("src/link.rs")).unwrap();
    fs::create_dir(f.root.join("real")).unwrap();
    fs::write(f.root.join("real/c.rs"), "x\n").unwrap();
    std::os::unix::fs::symlink(f.root.join("real"), f.proj.join("linked")).unwrap();
    let p = |rel: &str| f.proj.join(rel).display().to_string();
    for file in [
        outside.display().to_string(),
        p("src/link.rs"),
        p("linked/c.rs"),
        p("src/missing.rs"),
        p("src/../src/a.rs"),
        "src/a.rs".to_string(),
        p("src"),
    ] {
        assert_eq!(
            f.edit("Write", &file, &[("MODE", "diag")]),
            (0, String::new()),
            "{file}"
        );
        assert!(!f.ran(), "{file}: the checker ran");
    }
    // The same file through its real path is checked.
    assert_ne!(f.edit("Write", &f.file(), &[("MODE", "diag")]).1, "");
    assert!(f.ran());
}

#[test]
fn failures_are_one_line_and_never_block() {
    let f = fx();
    // A missing checker names itself (its file name only).
    f.config(
        "[hooks.edit_diagnostics]\nenabled = true\ncheckers = { rs = [\"/nonexistent/dir/lint-x\"] }\n",
    );
    let (code, out) = f.edit("Write", &f.file(), &[]);
    assert_eq!(code, 0);
    assert_eq!(
        context(&out),
        "MoAI edit-diagnostics (lint-x, src/a.rs): the checker `lint-x` was not found; this edit was not checked"
    );
    // A checker past its time is stopped with its whole process group.
    f.enable("timeout_secs = 1");
    let pidf = f.root.join("child.pid");
    let start = Instant::now();
    let (code, out) = f.edit(
        "Write",
        &f.file(),
        &[("MODE", "sleep"), ("PIDF", pidf.to_str().unwrap())],
    );
    assert_eq!(code, 0);
    assert!(
        start.elapsed() < Duration::from_secs(5),
        "{:?}",
        start.elapsed()
    );
    assert!(context(&out)
        .ends_with("stopped after 1 s with its whole process group; nothing is reported"));
    let pid: i32 = fs::read_to_string(&pidf).unwrap().trim().parse().unwrap();
    let gone = (0..200).any(|_| {
        // SAFETY: kill(2) with signal 0 only asks whether the test's own child still exists.
        let alive = unsafe { libc::kill(pid, 0) } == 0;
        if alive {
            std::thread::sleep(Duration::from_millis(10));
        }
        !alive
    });
    assert!(gone, "the checker's background child {pid} survived");
    // Huge output is read up to the bound and drained; bytes that are not UTF-8 are replaced.
    f.enable("");
    let (code, out) = f.edit("Write", &f.file(), &[("MODE", "huge")]);
    assert_eq!(code, 0);
    assert!(
        context(&out).contains("; output cut at 64 KiB per stream"),
        "{out}"
    );
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "nonutf8")]);
    let text = context(&out);
    assert!(
        text.contains("; bytes that are not UTF-8 were replaced"),
        "{text}"
    );
    assert!(
        text.contains("src/a.rs 1:1 error bad \u{fffd}\u{fffd} bytes"),
        "{text}"
    );
}

#[test]
fn credential_shaped_diagnostics_are_masked() {
    let f = fx();
    f.enable("");
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "secret")]);
    let text = context(&out);
    assert!(!text.contains("abc123"), "{text}");
    assert!(
        text.contains("src/a.rs 1:1 error [masked: credential-shaped text]"),
        "{text}"
    );
    assert!(text.contains("src/a.rs 2:1 error plain"), "{text}");
}

#[test]
fn one_check_runs_per_project_at_a_time() {
    let f = fx();
    f.enable("");
    let mut first = f
        .command(
            &[
                "hook-dispatch",
                "--host",
                "claude",
                "--event",
                "edit-diagnostics",
            ],
            &[("MODE", "slow")],
        )
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    {
        use std::io::Write;
        first
            .stdin
            .take()
            .unwrap()
            .write_all(&input("Write", &f.file(), &f.proj))
            .unwrap();
    }
    let mark = f.root.join("ran");
    let start = Instant::now();
    while !mark.exists() {
        assert!(
            start.elapsed() < Duration::from_secs(20),
            "the first check never started"
        );
        std::thread::sleep(Duration::from_millis(10));
    }
    let (code, out) = f.edit(
        "Edit",
        &f.file(),
        &[("MODE", "diag"), ("MARK", "/dev/null")],
    );
    assert_eq!(code, 0);
    assert_eq!(
        context(&out),
        "MoAI edit-diagnostics (checker.sh, src/a.rs): another check is running in this project; this edit was not checked"
    );
    let done = first.wait_with_output().unwrap();
    assert!(context(String::from_utf8_lossy(&done.stdout).trim()).contains("error slow"));
}

/// S1 audit B1 (CWE-88): a file whose name starts with `-` reaches the checker as a path
/// (`./<rel>`), never as an option; the edge is the first character of the relative path.
#[test]
fn a_file_named_like_an_option_reaches_the_checker_as_a_path() {
    let f = fx();
    f.enable("");
    fs::create_dir_all(f.proj.join("sub")).unwrap();
    let argf = f.root.join("args");
    for rel in ["-x.rs", "--config=x.rs", "x-.rs", "sub/-x.rs", "src/a.rs"] {
        fs::write(f.proj.join(rel), "x\n").unwrap();
        let file = f.proj.join(rel).display().to_string();
        let (code, out) = f.edit(
            "Write",
            &file,
            &[("MODE", "argv"), ("ARGF", argf.to_str().unwrap())],
        );
        assert_eq!((code, out.as_str()), (0, ""), "{rel}");
        let args: Vec<String> = fs::read_to_string(&argf)
            .unwrap()
            .lines()
            .map(|l| l.trim_start_matches("ARG:").to_string())
            .collect();
        assert_eq!(args, vec![format!("./{rel}")], "{rel}");
        assert!(args.iter().all(|a| !a.starts_with('-')), "{rel}: {args:?}");
    }
}

/// S1 audit N1: control characters (other than tab) and ANSI escape sequences in the checker's
/// output never reach the context.
#[test]
fn control_characters_and_escapes_are_stripped() {
    let f = fx();
    f.enable("");
    let (_, out) = f.edit("Write", &f.file(), &[("MODE", "control")]);
    let text = context(&out);
    assert!(
        text.contains("\nsrc/a.rs 1:1 error bell x red end"),
        "{text:?}"
    );
    assert!(
        !text.chars().any(|c| c.is_control() && c != '\n'),
        "{text:?}"
    );
}

/// S1 audit N2: a checker that exits normally but leaves a background child behind gets its
/// whole process group stopped too, at once (the child keeps its stdout open).
#[test]
fn a_checker_that_leaves_a_background_child_is_cleaned_up() {
    let f = fx();
    f.enable("");
    let pidf = f.root.join("linger.pid");
    let start = Instant::now();
    let (code, out) = f.edit(
        "Write",
        &f.file(),
        &[("MODE", "linger"), ("PIDF", pidf.to_str().unwrap())],
    );
    assert_eq!(code, 0);
    assert!(context(&out).contains("src/a.rs 1:1 error done"), "{out}");
    assert!(
        start.elapsed() < Duration::from_secs(5),
        "{:?}",
        start.elapsed()
    );
    let pid: i32 = fs::read_to_string(&pidf).unwrap().trim().parse().unwrap();
    let gone = (0..100).any(|_| {
        // SAFETY: kill(2) with signal 0 only asks whether the test's own grandchild exists.
        let alive = unsafe { libc::kill(pid, 0) } == 0;
        if alive {
            std::thread::sleep(Duration::from_millis(10));
        }
        !alive
    });
    if !gone {
        // SAFETY: the test's own grandchild; never leave it running.
        unsafe { libc::kill(pid, libc::SIGKILL) };
    }
    assert!(gone, "the background child {pid} was left running");
}

/// S1 audit N3: the one-check-per-project lock lives in the user's own HOME state folder
/// (`edit-diagnostics/`, mode 0700), not in the shared temporary folder; a link in its place is
/// not followed (the check still runs, without a lock).
#[test]
fn the_lock_lives_in_the_private_home_state_folder() {
    use std::os::unix::fs::PermissionsExt;
    let f = fx();
    f.enable("");
    assert_eq!(
        f.edit("Write", &f.file(), &[("MODE", "clean")]),
        (0, String::new())
    );
    let dir = f.home.join(".moai-cli/edit-diagnostics");
    let mode = fs::symlink_metadata(&dir).unwrap().permissions().mode() & 0o777;
    assert_eq!(mode, 0o700);
    let locks: Vec<_> = fs::read_dir(&dir)
        .unwrap()
        .map(|e| e.unwrap().file_name())
        .collect();
    assert_eq!(locks.len(), 1, "{locks:?}");
    assert!(locks[0].to_string_lossy().ends_with(".lock"));
    // A link in the folder's place is not followed.
    fs::remove_dir_all(&dir).unwrap();
    let elsewhere = f.root.join("elsewhere");
    fs::create_dir(&elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &dir).unwrap();
    let (code, out) = f.edit("Write", &f.file(), &[("MODE", "diag")]);
    assert_eq!(code, 0);
    assert!(context(&out).contains("3 diagnostic line(s)"), "{out}");
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
}

/// SPEC-20261117-edit-diagnostics-followups AC-2 (#3179 audit): every run of one project reuses
/// its one lock file (N runs, M files: still one), and a second project adds exactly one; the
/// files are empty and stay after the runs (D1: removing them would race).
#[test]
fn one_lock_file_per_project_is_reused() {
    let f = fx();
    f.enable("");
    let dir = f.home.join(".moai-cli/edit-diagnostics");
    let locks = || -> Vec<(String, u64)> {
        let mut v: Vec<_> = fs::read_dir(&dir)
            .unwrap()
            .map(|e| {
                let e = e.unwrap();
                let name = e.file_name().to_string_lossy().into_owned();
                (name, e.metadata().unwrap().len())
            })
            .collect();
        v.sort();
        v
    };
    fs::write(f.proj.join("src/b.rs"), "fn b() {}\n").unwrap();
    for file in ["src/a.rs", "src/b.rs", "src/a.rs"] {
        let path = f.proj.join(file).display().to_string();
        assert_eq!(f.edit("Write", &path, &[("MODE", "clean")]).0, 0);
    }
    let one = locks();
    assert_eq!(one.len(), 1, "{one:?}");
    assert_eq!(one[0].1, 0, "the lock file stays empty");
    // A second project in the same HOME: one more file, the first unchanged.
    let other = f.root.join("other");
    fs::create_dir_all(other.join("src")).unwrap();
    common::git::git(&other, &["init", "-q", "-b", "main"]);
    let file = other.join("src/c.rs");
    fs::write(&file, "fn c() {}\n").unwrap();
    let file = file.display().to_string();
    for _ in 0..2 {
        let mut child = f
            .command(
                &[
                    "hook-dispatch",
                    "--host",
                    "claude",
                    "--event",
                    "edit-diagnostics",
                ],
                &[("MODE", "clean")],
            )
            .current_dir(&other)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .unwrap();
        {
            use std::io::Write;
            let stdin = input("Write", &file, &other);
            child.stdin.take().unwrap().write_all(&stdin).unwrap();
        }
        assert!(child.wait_with_output().unwrap().status.success());
    }
    let two = locks();
    assert_eq!(two.len(), 2, "{two:?}");
    assert!(two.contains(&one[0]), "{two:?}");
}

/// SPEC-20261117-edit-diagnostics-followups AC-3 (#3179 audit, earlier N5): a malformed
/// `[hooks.edit_diagnostics]` turns off only that feature. With the same config file the write
/// guard still refuses a `.git/` write (a default rule) and a write outside the project (the
/// opt-in rule beside it), and a write inside the project still passes.
#[test]
fn a_malformed_table_leaves_the_other_guards_on() {
    let f = fx();
    f.config(&format!(
        "[hooks.edit_diagnostics]\nenabled = true\nmode = \"block\"\ncheckers = {{ rs = [{:?}] }}\n\n\
         [hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\"]\n",
        f.checker.to_str().unwrap()
    ));
    assert_eq!(
        f.edit("Write", &f.file(), &[("MODE", "diag")]),
        (0, String::new())
    );
    assert!(!f.ran(), "the malformed feature is off");
    let preflight = |path: &Path| -> String {
        let mut child = f
            .command(
                &[
                    "hook-dispatch",
                    "--host",
                    "claude",
                    "--event",
                    "tool-preflight",
                ],
                &[],
            )
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .unwrap();
        {
            use std::io::Write;
            let stdin = json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
                               "tool_input": {"file_path": path, "content": "x\n"},
                               "cwd": f.proj});
            child
                .stdin
                .take()
                .unwrap()
                .write_all(stdin.to_string().as_bytes())
                .unwrap();
        }
        let out = child.wait_with_output().unwrap();
        assert_eq!(out.status.code(), Some(0), "{out:?}");
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    };
    let denied = |stdout: &str| stdout.contains("\"permissionDecision\":\"deny\"");
    let git = preflight(&f.proj.join(".git/config"));
    assert!(denied(&git), "{git}");
    let outside = preflight(&f.root.join("outside/x.txt"));
    assert!(
        denied(&outside) && outside.contains("outside the project"),
        "{outside}"
    );
    assert_eq!(preflight(&f.proj.join("src/a.rs")), "");
}

/// SPEC REQ-11 / AC-6 budget: with the setting off the duty returns within 50 ms (median of 21
/// runs, process start included). Machine-dependent, so run on demand with `--ignored`.
#[test]
#[ignore]
fn budget_the_off_path_is_fast() {
    let f = fx();
    let mut times: Vec<Duration> = (0..21)
        .map(|_| {
            let start = Instant::now();
            assert_eq!(
                f.edit("Write", &f.file(), &[("MODE", "diag")]),
                (0, String::new())
            );
            start.elapsed()
        })
        .collect();
    times.sort();
    let median = times[times.len() / 2];
    eprintln!("budget: off-path median {median:?}");
    assert!(median < Duration::from_millis(50), "{median:?}");
}
