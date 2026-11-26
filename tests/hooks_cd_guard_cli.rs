//! SPEC-20261126-shell-cd-guard: the `tool-preflight` hook refuses a `cd`, `pushd` or `popd`
//! that would move the folder of a session working in a linked worktree of a MoAI project, and
//! leaves every command that does not move it alone. Every case runs the real binary
//! (`hook-dispatch --host claude --event tool-preflight`) against a temporary world made with a
//! real `git worktree add`, under a temporary HOME. No judged command is ever executed: the hook
//! only reads the text.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

const MODES: [&str; 2] = ["default", "bypassPermissions"];
const PREFIX: &str = "moai-boundary/session-cd: ";

/// Commands that move the folder of the session shell.
const MOVES: [&str; 16] = [
    "(ls) && cd src",
    "time cd src",
    "command -- cd src",
    "cd src",
    "cd",
    "ls && cd src",
    "ls; cd src",
    "false || cd src",
    "ls\ncd src",
    "{ cd src; }",
    "if true; then cd src; fi",
    "X=1 cd src",
    "builtin cd src",
    "eval 'cd src'",
    "pushd src",
    "popd",
];

/// Commands that hold the letters or run a `cd` that leaves the session folder where it is.
const STAYS: [&str; 19] = [
    // A `)` that the lexer takes for the end of a substitution or an arithmetic expansion
    // leaves the parentheses of the line unbalanced: such a line is not split.
    "v=$(cat <<'EOF'\n1) a\ncd src\nEOF\n)",
    "v=$(\n# 1) a\ncd src\n)",
    "( echo $((2*(1+1))); cd src )",
    "f () { cd src; }",
    "echo \"cd src\"",
    "grep 'cd ' a.txt",
    "printf '%s' 'a && cd src'",
    "ls # cd src",
    "cat <<'EOF'\ncd src\nEOF",
    "ls cd",
    "(cd src && ls)",
    "( cd src && ls ) && pwd",
    "v=$(cd src && pwd)",
    "echo `cd src; pwd`",
    "bash -c 'cd src && ls'",
    "sh -c \"cd src\"",
    "command -v cd",
    "/usr/bin/cd src",
    "git -C /tmp status",
];

fn git(dir: &Path, args: &[&str]) {
    let out = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args([
            "-c",
            "user.name=Fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .expect("spawn git");
    assert!(
        out.status.success(),
        "git {args:?} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
}

struct World {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    primary: PathBuf,
    own: PathBuf,
}

impl World {
    /// A primary checkout and a linked worktree `lane/own`. With `profile` the primary holds a
    /// MoAI project profile; `tracked` commits it, so the worktree holds it too.
    fn new(profile: bool, tracked: bool) -> World {
        let tmp = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let primary = root.join("primary");
        let home = root.join("home");
        for dir in [&primary, &home] {
            fs::create_dir_all(dir).unwrap();
        }
        git(&primary, &["init", "-q", "-b", "NeXT"]);
        fs::create_dir_all(primary.join("src")).unwrap();
        fs::write(primary.join("src/x"), "x\n").unwrap();
        if profile {
            fs::create_dir_all(moai_cli::state_root::path(&primary, "")).unwrap();
            fs::write(
                moai_cli::state_root::path(&primary, "project.yaml"),
                "{\"schema_version\":2,\"project_id\":\"demo\"}\n",
            )
            .unwrap();
        }
        if tracked {
            git(&primary, &["add", "-A"]);
        } else {
            git(&primary, &["add", "src"]);
        }
        git(&primary, &["commit", "-q", "-m", "fixture"]);
        let own = root.join("wt-own");
        git(
            &primary,
            &[
                "worktree",
                "add",
                "-q",
                "-b",
                "lane/own",
                own.to_str().unwrap(),
            ],
        );
        World {
            _tmp: tmp,
            home,
            primary,
            own,
        }
    }

    /// One `tool-preflight` call from `cwd` in `mode`; `project_dir` sets `CLAUDE_PROJECT_DIR`.
    /// The deny reason, or `None` when the hook says nothing.
    fn call(
        &self,
        cwd: &Path,
        command: &str,
        mode: &str,
        project_dir: Option<&Path>,
    ) -> Option<String> {
        let input = json!({
            "session_id": "fixture",
            "cwd": cwd,
            "hook_event_name": "PreToolUse",
            "tool_name": "Bash",
            "tool_input": { "command": command },
            "permission_mode": mode,
        });
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args([
            "hook-dispatch",
            "--host",
            "claude",
            "--event",
            "tool-preflight",
        ])
        .current_dir(cwd)
        .env("HOME", &self.home)
        .env_remove("CLAUDE_PROJECT_DIR")
        .env_remove("MOAI_FORGE_GH_BIN")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
        if let Some(dir) = project_dir {
            cmd.env("CLAUDE_PROJECT_DIR", dir);
        }
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, input.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        assert_eq!(
            out.status.code(),
            Some(0),
            "{command:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        let text = String::from_utf8_lossy(&out.stdout);
        if text.trim().is_empty() {
            return None;
        }
        let value: Value = serde_json::from_str(text.trim())
            .unwrap_or_else(|e| panic!("stdout for {command:?} is not JSON ({e}): {text}"));
        let inner = &value["hookSpecificOutput"];
        assert_eq!(
            inner["permissionDecision"],
            json!("deny"),
            "{command:?}: {value}"
        );
        Some(
            inner["permissionDecisionReason"]
                .as_str()
                .unwrap()
                .to_owned(),
        )
    }
}

/// AC-1 (REQ-1): every way of moving the session folder is denied in a worktree session, with one
/// line that names the other ways; also from a subfolder, from a session the host started in the
/// primary checkout, and where the profile is not tracked into the worktree.
#[test]
fn a_worktree_session_cd_is_denied() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in MOVES {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            assert!(reason.starts_with(PREFIX), "{command:?}: {reason}");
            assert_eq!(reason.lines().count(), 1, "one line: {reason}");
            for way in [
                "absolute path",
                "--manifest-path",
                "plain single",
                "object id",
            ] {
                assert!(reason.contains(way), "{way} not named: {reason}");
            }
        }
    }
    let from_sub = w.call(&w.own.join("src"), "cd ..", "default", None);
    assert!(from_sub.is_some_and(|r| r.starts_with(PREFIX)), "subfolder");
    let started_in_primary = w.call(&w.own, "cd src", "bypassPermissions", Some(&w.primary));
    assert!(
        started_in_primary.is_some_and(|r| r.starts_with(PREFIX)),
        "a session started in the primary checkout"
    );

    let untracked = World::new(true, false);
    assert!(
        !moai_cli::state_root::path(&untracked.own, "project.yaml").exists(),
        "the profile is not in this worktree"
    );
    let reason = untracked.call(&untracked.own, "cd src", "default", None);
    assert!(
        reason.is_some_and(|r| r.starts_with(PREFIX)),
        "an untracked profile"
    );
}

/// SPEC-less card C-583: the refusal never recommends a form the lane rules forbid. A subshell
/// `( cd <path> && … )`, `git -C`, a bare `cd <path>`, `pushd` and `popd` must not come back into
/// the text, in either permission mode and for every way of moving the session folder.
#[test]
fn the_session_cd_refusal_recommends_no_cd_form() {
    const FORBIDDEN: [&str; 7] = ["( cd", "(cd", "cd <", "cd ..", "git -c", "pushd", "popd"];
    let w = World::new(true, true);
    for mode in MODES {
        for command in MOVES {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            let text = reason
                .strip_prefix(PREFIX)
                .unwrap_or_else(|| panic!("prefix: {reason}"))
                .to_ascii_lowercase();
            for form in FORBIDDEN {
                assert!(
                    !text.contains(form),
                    "{command:?}: the refusal recommends {form:?}: {reason}"
                );
            }
        }
    }
}

/// AC-2 (REQ-2, REQ-3): letters in a string, a comment or a here-document, a subshell, a
/// substitution, another shell's string and a program called `cd` are not denied.
#[test]
fn a_cd_that_does_not_move_the_session_is_allowed() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in STAYS {
            let got = w.call(&w.own, command, mode, None);
            assert!(
                !got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
                "denied in {mode}: {command:?}: {got:?}"
            );
        }
    }
}

/// SPEC-20261126-shell-cd-guard revision 2, REQ-2 and REQ-3, AC-6: these inputs leave the
/// session folder where it is, yet the rule denies them because the shell reader does not follow
/// these forms; the test pins the behaviour until the reader does.
#[test]
fn a_known_false_denial_stays_denied() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in [
            "v=$(\n# 1) cd src (\n)",
            "v=$(\n# 1) a; cd src; echo (\n)",
            "v=$(cat <<'EOF'\n1) a\ncd src\n(\nEOF\n)",
            "cat <<EOF$X\nEOF\ncd src\nEOF$X",
        ] {
            let reason = w
                .call(&w.own, command, mode, None)
                .unwrap_or_else(|| panic!("allowed in {mode}: {command:?}"));
            assert!(reason.starts_with(PREFIX), "{command:?}: {reason}");
        }
    }
}

/// The dangerous path of the rule (REQ-1, REQ-4): a session folder reached through a link is
/// judged by the checkout it leads to, and judging a compound command runs nothing of it.
#[test]
fn a_linked_session_folder_is_judged_by_where_it_leads() {
    let w = World::new(true, true);
    let into_own = w.primary.parent().unwrap().join("link-into-own");
    std::os::unix::fs::symlink(w.own.join("src"), &into_own).unwrap();
    let into_primary = w.own.join("link-into-primary");
    std::os::unix::fs::symlink(w.primary.join("src"), &into_primary).unwrap();
    for mode in MODES {
        let got = w.call(&into_own, "ls && cd .. && rm -f src/x", mode, None);
        assert!(
            got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
            "a link into the worktree, {mode}: {got:?}"
        );
        let got = w.call(&into_primary, "cd .. && ls", mode, None);
        assert!(
            !got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
            "a link into the primary checkout, {mode}: {got:?}"
        );
    }
    for checkout in [&w.own, &w.primary] {
        assert!(checkout.join("src/x").is_file(), "a judged command ran");
    }
}

/// AC-3 (REQ-4): the primary checkout session, a linked worktree of a folder that is no MoAI
/// project and a command the lexer cannot split keep their `cd`.
#[test]
fn a_session_outside_a_worktree_keeps_cd() {
    let w = World::new(true, true);
    for mode in MODES {
        for command in ["cd src", "ls && cd src", "pushd src"] {
            let got = w.call(&w.primary, command, mode, None);
            assert_eq!(got, None, "primary session, {mode}: {command:?}");
        }
        let got = w.call(&w.own, "case a in a) cd src;; esac", mode, None);
        assert!(
            !got.as_deref().is_some_and(|r| r.starts_with(PREFIX)),
            "unsplittable, {mode}: {got:?}"
        );
    }
    let plain = World::new(false, false);
    for mode in MODES {
        let got = plain.call(&plain.own, "cd src", mode, None);
        assert_eq!(got, None, "no MoAI project, {mode}");
    }
}
