//! SPEC-20261120-spec-branch-followups (card C-133, follow-up of C-112's audit notes N8, N9, N5):
//! the remote-name boundary of `spec take`, the hint that must end in one branch form, and the
//! hook reason for a name that reads as two SPEC ids because of its repair suffix. Every case runs
//! the real binary in a temporary repository with a temporary HOME; no real session or network.
#![cfg(unix)]
mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BIN: &str = crate::common::MOAI_BIN;
const OLD: &str = "spec/20261117-login-fix";
const NEW: &str = "spec/login-fix-20261117";

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    v["result"].clone()
}

fn refused(out: &Output) -> String {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    v["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_string()
}

/// A HOME, a registered ledger and a repository `demo` on `main` with one commit.
struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        ledger::register(home.path(), "demo").unwrap();
        let w = World { home, project };
        w.git(&["init", "-q", "-b", "main"]);
        fs::write(w.project.join("a.txt"), "a\n").unwrap();
        w.git(&["add", "-A"]);
        w.commit("init");
        w
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    fn commit(&self, message: &str) {
        common::git::git_ok(
            &self.project,
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

    fn run(&self, args: &[&str]) -> Output {
        self.home.command(args).output().expect("run moai")
    }

    fn take(&self, id: &str) -> Output {
        self.run(&["spec", "take", id, "--project", self.p()])
    }

    /// A draft SPEC `SPEC-20261117-<slug>` that passes `spec check` (placeholders filled).
    fn draft_spec(&self, slug: &str) -> String {
        let r = ok(&self.run(&[
            "spec",
            "new",
            slug,
            "--title",
            "Login fix",
            "--minimal",
            "--date",
            "2026-11-17",
            "--project",
            self.p(),
        ]));
        let id = r["id"].as_str().unwrap().to_string();
        let path = self.project.join(".moai/specs").join(&id).join("spec.md");
        fs::write(&path, filled(&fs::read_to_string(&path).unwrap())).unwrap();
        id
    }

    fn branches(&self) -> String {
        self.git(&["branch", "--list", "--format=%(refname:short)"])
    }

    /// Branches, HEAD and the status: what a refusal must leave as it was.
    fn state(&self) -> (String, String, String) {
        (
            self.git(&[
                "branch",
                "--list",
                "--format=%(refname:short) %(objectname)",
            ]),
            self.git(&["rev-parse", "HEAD"]),
            self.git(&["status", "--porcelain"]),
        )
    }
}

/// A template's text with every `<…>` placeholder filled with the word `filled`.
fn filled(text: &str) -> String {
    let mut out = String::new();
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push_str("filled");
        rest = &rest[open + close + 1..];
    }
    out.push_str(rest);
    out
}

/// N8 (REQ-1, AC-1): a remote branch whose name only starts with the SPEC's older branch name (a
/// repair branch `-r2`, any longer name) is another branch: it does not make the older one exist
/// on a remote, so `spec take` goes ahead with the name-first branch. The refusal for the exact
/// name stays (`name_first_take_refuses_an_old_form_branch_that_exists_only_on_a_remote`).
#[test]
fn a_remote_branch_that_only_starts_with_the_old_name_does_not_block_take() {
    let w = World::new();
    let id = w.draft_spec("login-fix");
    for name in [
        "spec/20261117-login-fix-r2",
        "spec/20261117-login-fix-extra",
    ] {
        w.git(&["update-ref", &format!("refs/remotes/origin/{name}"), "HEAD"]);
    }
    let result = ok(&w.take(&id));
    assert_eq!(result["branch"], NEW, "{result}");
    assert!(
        w.branches().lines().any(|b| b == NEW),
        "the name-first branch is made: {}",
        w.branches()
    );
}

/// N9 (REQ-2, AC-2): the refusal for an older branch that exists only on a remote names steps
/// that leave one branch form. With the name-first branch also local, switching to the older
/// branch alone would make both forms and the next take would refuse again; the hint says to
/// merge the name-first branch into it and delete the name-first one, and following exactly that
/// makes the take go ahead. Without a local name-first branch the single `git switch` is enough.
#[test]
fn the_remote_only_hint_ends_in_one_form() {
    // Name-first branch local, older branch on a remote only.
    let w = World::new();
    let id = w.draft_spec("login-fix");
    w.git(&["branch", NEW]);
    w.git(&["update-ref", &format!("refs/remotes/origin/{OLD}"), "HEAD"]);
    let before = w.state();
    let message = refused(&w.take(&id));
    assert!(message.contains("remote"), "{message}");
    assert!(
        message.contains("merge") && message.contains("delete"),
        "the hint names the merge and the delete: {message}"
    );
    assert_eq!(w.state(), before, "a refusal changes nothing");
    // Follow the hint: switch to the older branch, merge the name-first one, delete it.
    w.git(&["switch", OLD]);
    w.git(&["merge", "--no-edit", NEW]);
    w.git(&["branch", "-d", NEW]);
    w.git(&["switch", "main"]);
    let message = w.take(&id);
    let v = envelope(&message);
    assert_eq!(
        message.status.code(),
        Some(0),
        "after the hint's steps the take goes ahead: {v}"
    );
    assert_eq!(v["result"]["branch"], OLD, "{v}");

    // No local name-first branch: the one `git switch` of the hint is enough.
    let w = World::new();
    let id = w.draft_spec("login-fix");
    w.git(&["update-ref", &format!("refs/remotes/origin/{OLD}"), "HEAD"]);
    let message = refused(&w.take(&id));
    assert!(message.contains("git switch"), "{message}");
    w.git(&["switch", OLD]);
    w.git(&["switch", "main"]);
    let result = ok(&w.take(&id));
    assert_eq!(result["branch"], OLD, "{result}");
}

/// A temporary MoAI repository (a committed `.moai/` folder) on `branch` with one commit.
fn repo(branch: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap();
    fs::create_dir_all(root.join("home")).unwrap();
    for args in [
        vec!["init", "-q", "-b", branch],
        vec!["config", "user.name", "fixture"],
        vec!["config", "user.email", "fixture@example.invalid"],
        vec!["config", "commit.gpgsign", "false"],
    ] {
        common::git::git_ok(&root, &args);
    }
    fs::create_dir_all(root.join(".moai")).unwrap();
    fs::write(root.join(".moai/.keep"), "").unwrap();
    fs::write(root.join("a.txt"), "a\n").unwrap();
    fs::write(root.join(".gitignore"), "home/\n").unwrap();
    common::git::git_ok(&root, &["add", "-A"]);
    common::git::git_ok(&root, &["commit", "-q", "-m", "init"]);
    (dir, root)
}

fn preflight(cwd: &Path, input: &Value) -> (i32, String) {
    let mut cmd = Command::new(BIN);
    cmd.args([
        "hook-dispatch",
        "--host",
        "claude",
        "--event",
        "tool-preflight",
    ])
    .current_dir(cwd)
    .env("HOME", cwd.join("home"))
    .env_remove("MOAI_FORGE_GH_BIN")
    .env("GIT_CONFIG_GLOBAL", "/dev/null")
    .env("GIT_CONFIG_SYSTEM", "/dev/null")
    .stdin(Stdio::piped())
    .stdout(Stdio::piped())
    .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    (
        out.status.code().unwrap(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
    )
}

/// N5 (REQ-3, AC-3): `spec/20261117-demo-r2` reads as `SPEC-20261117-demo-r2` and, without its
/// repair suffix, as `SPEC-20261117-demo`. With no SPEC folder of exactly one of them the code
/// write is refused, and the reason says the repair suffix is part of why the name reads two
/// ways (the older reason named only the two branch forms). The reason is a fixed string.
#[test]
fn the_hook_reason_names_the_repair_suffix() {
    let (_d, root) = repo("spec/20261117-demo-r2");
    let input = json!({"hook_event_name": "PreToolUse", "tool_name": "Write",
        "tool_input": {"file_path": "src/main.rs", "content": "fn main() {}\n"}, "cwd": root});
    let (code, stdout) = preflight(&root, &input);
    assert_eq!(code, 0);
    let v: Value = serde_json::from_str(stdout.trim()).expect("a deny is one JSON object");
    assert_eq!(v["hookSpecificOutput"]["permissionDecision"], "deny", "{v}");
    let reason = v["hookSpecificOutput"]["permissionDecisionReason"]
        .as_str()
        .unwrap();
    assert!(reason.contains("reads as two SPEC ids"), "{reason}");
    assert!(reason.contains("repair suffix"), "{reason}");
    assert!(
        !reason.contains("demo"),
        "no branch name is echoed: {reason}"
    );
}
