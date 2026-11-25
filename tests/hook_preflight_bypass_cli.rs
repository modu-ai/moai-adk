//! SPEC-20261125-tool-preflight-bypass-guards: in `bypassPermissions` mode the `tool-preflight`
//! hook judges a Bash command with the bash-boundary handler and turns the force-push
//! confirmation into a deny; every other mode answers as before. Every case runs the real binary
//! (`hook-dispatch --host claude --event tool-preflight`) against a temporary world: a primary
//! checkout on the integration branch, the session's own linked worktree, a plain folder, all
//! under a temporary HOME. The world lives under this crate's `target/` folder because the
//! system temporary folder is itself a boundary root of the judgment. No judged command is ever
//! executed: the hook only reads the text.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::Instant;

/// A folder that exists on no machine and lies outside every boundary root.
const OUT: &str = "/opt/moai-boundary-outside";
const BYPASS: &str = "bypassPermissions";
/// The modes that are not bypass, as the hook input names them, and an input without one.
const OTHER_MODES: [&str; 5] = ["default", "plan", "acceptEdits", "auto", "dontAsk"];

/// Commands that touch the integration branch `{I}`: a deny in bypass mode.
const REF_DENY: [&str; 5] = [
    "git branch -f {I} HEAD",
    "git branch -M {I}",
    "git branch -D {I}",
    "git update-ref refs/heads/{I} HEAD",
    "git push origin work:{I}",
];
/// The same moves on a branch that is not protected: allowed.
const REF_ALLOW: [&str; 3] = [
    "git branch -f lane/other HEAD",
    "git update-ref refs/heads/lane/other HEAD",
    "git push origin work:lane/other",
];
/// Recursive deletes outside the session checkout and the temporary folders: a deny in bypass
/// mode. `{HOME}` is the fixture HOME, a folder of the world and not a temporary root.
const RM_DENY: [&str; 3] = [
    "rm -rf {OUT}",
    "rm -r /usr/local/lib/moai-fixture",
    "rm -r {HOME}/data",
];
/// Deletes that stay as they are: inside the worktree by a relative path, and inside a
/// temporary folder (the flow rule allows both).
const RM_ALLOW: [&str; 3] = [
    "rm -r build",
    "rm -rf ./build/x",
    "rm -rf /tmp/moai-fixture-x",
];
/// The force pushes the confirmation class asks about (never a protected branch).
const FORCE: [&str; 4] = [
    "git push origin +feature",
    "git push --force origin feature",
    "git push -f origin feature",
    "git push --force-with-lease origin feature",
];

struct Run {
    code: i32,
    stdout: String,
    stderr: String,
}

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

fn under_a_temporary_root(path: &Path) -> bool {
    ["/tmp", "/private/tmp", "/var/tmp", "/private/var/tmp"]
        .iter()
        .any(|root| path.starts_with(root))
        || path.starts_with("/var/folders")
        || path.starts_with("/private/var/folders")
}

/// The folder the worlds are built in: this crate's `target/` folder, or the folder the caller
/// names when the checkout itself lies under a temporary root.
fn fixture_base() -> PathBuf {
    let manifest = fs::canonicalize(env!("CARGO_MANIFEST_DIR")).unwrap();
    let base = manifest.join("target").join("hook-preflight-bypass-tests");
    if !under_a_temporary_root(&base) {
        return base;
    }
    let named = std::env::var_os("MOAI_BOUNDARY_TEST_WORLD")
        .map(PathBuf::from)
        .expect("this checkout lies under a temporary root; set MOAI_BOUNDARY_TEST_WORLD");
    assert!(
        !under_a_temporary_root(&named),
        "{named:?} is a temporary root"
    );
    named
}

struct World {
    _tmp: tempfile::TempDir,
    home: PathBuf,
    own: PathBuf,
    plain: PathBuf,
    /// The integration branch the project profile names.
    branch: String,
}

impl World {
    /// A primary checkout on `branch`, named by the profile, and a linked worktree `lane/own`.
    fn new(branch: &str) -> World {
        let base = fixture_base();
        fs::create_dir_all(&base).unwrap();
        let tmp = tempfile::Builder::new()
            .prefix("hpb-")
            .tempdir_in(&base)
            .unwrap();
        let root = fs::canonicalize(tmp.path()).unwrap();
        let primary = root.join("primary");
        let home = root.join("home");
        let plain = root.join("plain");
        for dir in [&primary, &home, &plain] {
            fs::create_dir_all(dir).unwrap();
        }
        // A checkout with no `.moai/project.yaml`: the boundary judgment would deny here, so only
        // the project check keeps it from running.
        git(&plain, &["init", "-q", "-b", branch]);
        git(&primary, &["init", "-q", "-b", branch]);
        fs::create_dir_all(moai_cli::state_root::path(&primary, "")).unwrap();
        fs::write(
            moai_cli::state_root::path(&primary, "project.yaml"),
            format!(
                "{{\"schema_version\":2,\"project_id\":\"demo\",\"work_flow\":{{\"flow\":\"m-flow\",\"integration_branch\":\"{branch}\"}}}}\n"
            ),
        )
        .unwrap();
        fs::create_dir_all(primary.join("build")).unwrap();
        fs::write(primary.join("build/x"), "x\n").unwrap();
        git(&primary, &["add", "-A"]);
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
            own,
            plain,
            branch: branch.to_owned(),
        }
    }

    fn fill(&self, text: &str) -> String {
        text.replace("{I}", &self.branch)
            .replace("{OUT}", OUT)
            .replace("{HOME}", &self.home.display().to_string())
    }

    /// One `tool-preflight` call from `cwd`; `mode` is the `permission_mode` of the input, or
    /// absent for `None`.
    fn call(&self, cwd: &Path, command: &str, mode: Option<Value>) -> Run {
        let mut input = json!({
            "session_id": "fixture",
            "cwd": cwd,
            "hook_event_name": "PreToolUse",
            "tool_name": "Bash",
            "tool_input": { "command": self.fill(command) },
        });
        if let Some(mode) = mode {
            input["permission_mode"] = mode;
        }
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
        let mut child = cmd.spawn().unwrap();
        common::feed_stdin(&mut child, input.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        Run {
            code: out.status.code().unwrap(),
            stdout: String::from_utf8_lossy(&out.stdout).into_owned(),
            stderr: String::from_utf8_lossy(&out.stderr).into_owned(),
        }
    }

    fn bypass(&self, command: &str) -> Answer {
        answer(&self.call(&self.own, command, Some(json!(BYPASS))), command)
    }
}

/// What the hook said: nothing (the call goes ahead), an `ask`, or a `deny` with its reason.
#[derive(Debug, PartialEq, Eq)]
enum Answer {
    Allow,
    Ask(String),
    Deny(String),
}

fn answer(run: &Run, what: &str) -> Answer {
    assert_eq!(run.code, 0, "exit code for {what:?}: {}", run.stderr);
    if run.stdout.trim().is_empty() {
        return Answer::Allow;
    }
    let value: Value = serde_json::from_str(run.stdout.trim())
        .unwrap_or_else(|e| panic!("stdout for {what:?} is not JSON ({e}): {}", run.stdout));
    let inner = &value["hookSpecificOutput"];
    let reason = inner["permissionDecisionReason"]
        .as_str()
        .unwrap_or_default()
        .to_owned();
    match inner["permissionDecision"].as_str() {
        Some("ask") => Answer::Ask(reason),
        Some("deny") => Answer::Deny(reason),
        other => panic!("decision for {what:?}: {other:?}"),
    }
}

fn deny_with(a: Answer, prefix: &str, what: &str) {
    match a {
        Answer::Deny(reason) => assert!(
            reason.starts_with(prefix),
            "reason for {what:?} is {reason:?}, want prefix {prefix:?}"
        ),
        other => panic!("{what:?}: want a deny with {prefix:?}, got {other:?}"),
    }
}

// ------------------------------------------------------------------ AC-1

#[test]
fn ac1_a_move_of_the_integration_branch_is_denied_in_bypass_mode() {
    // The branch name is read from the profile: `NeXT` and a second name.
    for branch in ["NeXT", "release-line"] {
        let world = World::new(branch);
        for command in REF_DENY {
            deny_with(
                world.bypass(command),
                "moai-boundary/protected-ref:",
                command,
            );
        }
        for command in REF_ALLOW {
            assert_eq!(world.bypass(command), Answer::Allow, "{branch}: {command}");
        }
    }
}

// ------------------------------------------------------------------ AC-2

#[test]
fn ac2_a_delete_outside_the_checkout_is_denied_in_bypass_mode() {
    let world = World::new("NeXT");
    for command in RM_DENY {
        deny_with(
            world.bypass(command),
            "moai-boundary/delete-outside:",
            command,
        );
    }
    for command in RM_ALLOW {
        assert_eq!(world.bypass(command), Answer::Allow, "{command}");
    }
}

// ------------------------------------------------------------------ AC-3

#[test]
fn ac3_a_force_push_is_denied_in_bypass_mode_and_still_asked_otherwise() {
    let world = World::new("NeXT");
    for command in FORCE {
        let reason = match world.bypass(command) {
            Answer::Deny(reason) => reason,
            other => panic!("{command:?}: want a deny, got {other:?}"),
        };
        assert!(
            reason.starts_with("MoAI tool-preflight: a force push"),
            "reason for {command:?}: {reason}"
        );
        assert!(
            !reason.contains("feature"),
            "the reason echoes the command: {reason}"
        );
        match answer(
            &world.call(&world.own, command, Some(json!("default"))),
            command,
        ) {
            Answer::Ask(_) => {}
            other => panic!("{command:?} in default mode: want an ask, got {other:?}"),
        }
    }
}

// ------------------------------------------------------------------ AC-4

#[test]
fn ac4_every_other_mode_answers_as_before() {
    let world = World::new("NeXT");
    let mut modes: Vec<Option<Value>> = OTHER_MODES.iter().map(|m| Some(json!(m))).collect();
    modes.push(None);
    modes.push(Some(json!(5)));
    modes.push(Some(Value::Null));
    for mode in modes {
        for command in REF_DENY.iter().chain(RM_DENY.iter()) {
            let run = world.call(&world.own, command, mode.clone());
            assert_eq!(
                answer(&run, command),
                Answer::Allow,
                "mode {mode:?}: {command}"
            );
        }
        for command in FORCE {
            let run = world.call(&world.own, command, mode.clone());
            assert!(
                matches!(answer(&run, command), Answer::Ask(_)),
                "mode {mode:?}: {command}"
            );
        }
    }
}

// ------------------------------------------------------------------ AC-5

#[test]
fn ac5_outside_a_moai_project_nothing_new_is_judged() {
    let world = World::new("NeXT");
    for command in REF_DENY.iter().chain(RM_DENY.iter()) {
        let run = world.call(&world.plain, command, Some(json!(BYPASS)));
        assert_eq!(
            answer(&run, command),
            Answer::Allow,
            "plain folder: {command}"
        );
    }
}

// ------------------------------------------------------------------ AC-8

#[test]
fn ac8_the_latency_of_an_allowed_and_a_denied_call_is_reported() {
    let world = World::new("NeXT");
    let measure = |command: &str| {
        let mut ms: Vec<f64> = (0..15)
            .map(|_| {
                let start = Instant::now();
                let run = world.call(&world.own, command, Some(json!(BYPASS)));
                let took = start.elapsed().as_secs_f64() * 1000.0;
                assert_eq!(run.code, 0, "{}", run.stderr);
                took
            })
            .collect();
        ms.sort_by(|a, b| a.partial_cmp(b).unwrap());
        (ms[ms.len() / 2], ms[(ms.len() * 9) / 10])
    };
    let (allow50, allow90) = measure("git status");
    let (deny50, deny90) = measure("git branch -f NeXT HEAD");
    println!(
        "hook-dispatch tool-preflight, bypassPermissions, profile {}: allowed p50 {allow50:.1} ms p90 {allow90:.1} ms; denied p50 {deny50:.1} ms p90 {deny90:.1} ms",
        if cfg!(debug_assertions) { "debug" } else { "release" }
    );
}
