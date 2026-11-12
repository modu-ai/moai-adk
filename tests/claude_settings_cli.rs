//! #2239 (rc.2 I2): `claude-settings-plan` / `claude-settings-apply` expose the existing
//! three-way settings merge (`src/claude.rs`) on the CLI. Every case runs in a temporary
//! project with a temporary HOME; no real user settings are read or written.
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::Path,
    process::{Command, Stdio},
};
use tempfile::TempDir;

/// Runs one operation with `home` as HOME and CLAUDE_CONFIG_DIR unset; returns the
/// exit code and the envelope.
fn run(args: &[&str], home: &Path, stdin: &Value) -> (Option<i32>, Value) {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .env("HOME", home)
        .env_remove("CLAUDE_CONFIG_DIR")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, stdin.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (out.status.code(), v)
}

fn plan(project: &Path, home: &Path, desired: &Value) -> Value {
    let (code, v) = run(
        &["claude-settings-plan", project.to_str().unwrap()],
        home,
        desired,
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["operation"], json!("claude.settings_plan"));
    v["result"].clone()
}

fn apply(project: &Path, home: &Path, desired: &Value, digest: &str) -> (Option<i32>, Value) {
    run(
        &[
            "claude-settings-apply",
            project.to_str().unwrap(),
            "--plan-digest",
            digest,
            "--permission-reference",
            "test-2239",
        ],
        home,
        desired,
    )
}

fn desired() -> Value {
    json!({"outputStyle": "MoAI-Easy",
           "statusLine": {"type": "command", "command": "python3 \"${CLAUDE_PROJECT_DIR}/.moai-cli/statusline.py\""}})
}

/// AC-1: plan writes nothing, apply creates the file and the receipt while keeping a
/// user key, and a re-plan of the same request is a no-op.
#[test]
fn plan_apply_then_replan_is_a_no_op() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    fs::create_dir_all(p.join(".claude")).unwrap();
    fs::write(
        p.join(".claude/settings.json"),
        "{\n  \"model\": \"user-choice\"\n}\n",
    )
    .unwrap();

    let first = plan(p, home.path(), &desired());
    assert_eq!(first["action"], json!("update"));
    assert_eq!(first["writes"], json!(false));
    assert!(
        !p.join(".moai-cli/manifests/claude-settings.json").exists(),
        "plan writes nothing"
    );

    let (code, v) = apply(
        p,
        home.path(),
        &desired(),
        first["plan_digest"].as_str().unwrap(),
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["state"], json!("applied"));
    let settings: Value =
        serde_json::from_slice(&fs::read(p.join(".claude/settings.json")).unwrap()).unwrap();
    assert_eq!(settings["model"], json!("user-choice"), "user key kept");
    assert_eq!(settings["outputStyle"], json!("MoAI-Easy"));
    let receipt: Value = serde_json::from_slice(
        &fs::read(p.join(".moai-cli/manifests/claude-settings.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["owned"].as_object().unwrap().len(), 2, "{receipt}");

    let again = plan(p, home.path(), &desired());
    assert_eq!(again["action"], json!("unchanged"));
    assert_eq!(again["planned_sha256"], Value::Null);
}

/// #2244 x #2238: a caller may set `permissions` (narrowing values only) through the CLI,
/// alongside the key the caller may never set, `hooks`.
#[test]
fn permissions_are_settable_through_the_cli_but_widening_modes_are_refused() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let narrowing = json!({"permissions": {"deny": ["Bash(rm:*)"], "defaultMode": "plan"}});
    let result = plan(p, home.path(), &narrowing);
    assert_eq!(result["action"], json!("create"), "{result}");
    let (code, v) = run(
        &["claude-settings-plan", p.to_str().unwrap()],
        home.path(),
        &json!({"permissions": {"defaultMode": "bypassPermissions"}}),
    );
    assert_eq!(code, Some(1), "{v}");
    assert!(v["error"]["message"]
        .as_str()
        .unwrap()
        .contains("widening modes are refused"));
    assert!(!p.join(".claude").exists(), "nothing written");
}

/// AC-1 / REQ-I2-01: keys outside the managed settings are refused and named,
/// and nothing is written; a stale digest is refused too.
#[test]
fn unmanaged_keys_and_stale_digests_are_refused() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let mut bad = desired();
    bad["env"] = json!({"MOAI_TEST_UNMANAGED": "fixture"});
    bad["hooks"] = json!({});
    let (code, v) = run(
        &["claude-settings-plan", p.to_str().unwrap()],
        home.path(),
        &bad,
    );
    assert_eq!(code, Some(1), "{v}");
    let msg = v["error"]["message"].as_str().unwrap();
    assert!(
        msg.contains("env, hooks") && msg.contains("allowed: outputStyle, statusLine"),
        "{msg}"
    );
    let (code, _) = apply(p, home.path(), &bad, &"0".repeat(64));
    assert_eq!(code, Some(1));
    assert!(!p.join(".claude").exists(), "nothing written");

    let (code, v) = apply(p, home.path(), &desired(), &"0".repeat(64));
    assert_eq!(code, Some(1));
    assert!(
        v["error"]["message"].as_str().unwrap().contains("stale"),
        "{v}"
    );
    assert!(!p.join(".claude/settings.json").exists());
}

/// Home writes are blocked: a project that is HOME, or whose `.claude` is
/// CLAUDE_CONFIG_DIR, is refused at plan time, so no digest for it exists.
#[test]
fn home_and_claude_config_dir_are_refused() {
    let home = TempDir::new().unwrap();
    let (code, v) = run(
        &["claude-settings-plan", home.path().to_str().unwrap()],
        home.path(),
        &desired(),
    );
    assert_eq!(code, Some(1), "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("home writes are blocked"),
        "{v}"
    );

    let project = TempDir::new().unwrap();
    fs::create_dir_all(project.path().join(".claude")).unwrap();
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(["claude-settings-plan", project.path().to_str().unwrap()])
        .env("HOME", home.path())
        .env("CLAUDE_CONFIG_DIR", project.path().join(".claude"))
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .and_then(|mut c| {
            // The plan is refused (exit 1) and may exit before reading stdin.
            common::feed_stdin(&mut c, desired().to_string().as_bytes());
            c.wait_with_output()
        })
        .unwrap();
    assert_eq!(out.status.code(), Some(1));
    assert!(!project.path().join(".claude/settings.json").exists());
}

/// Runs git in `dir` with the caller's git environment removed and a fixed identity.
fn git(dir: &Path, args: &[&str]) {
    let out = Command::new("git")
        .arg("-C")
        .arg(dir)
        .args(args)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_INDEX_FILE")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .output()
        .unwrap();
    assert!(out.status.success(), "git {args:?}: {out:?}");
}

/// #2581: the plan reports whether git ignores `.claude/settings.json`, and so whether a
/// session started in a linked worktree of the repository reads it. Four states: outside a
/// repository (`unknown`), ignored (`not_reached`), not ignored, and a tracked file that
/// matches an ignore pattern — git does not treat a tracked file as ignored. The facts never
/// change the plan itself: the action and digest are the same in every state.
#[test]
fn plan_reports_whether_worktree_sessions_read_the_settings_file() {
    let home = TempDir::new().unwrap();
    let project = TempDir::new().unwrap();
    let p = project.path();
    let facts = |v: &Value| {
        (
            v["settings_ignored"].as_str().unwrap().to_string(),
            v["worktree_sessions"].as_str().unwrap().to_string(),
        )
    };
    let outside = plan(p, home.path(), &desired());
    assert_eq!(
        facts(&outside),
        ("unknown".into(), "unknown".into()),
        "{outside}"
    );

    git(p, &["init", "-q"]);
    let open = plan(p, home.path(), &desired());
    assert_eq!(
        facts(&open),
        ("not_ignored".into(), "reached_when_committed".into())
    );

    fs::write(p.join(".gitignore"), ".claude/\n").unwrap();
    let ignored = plan(p, home.path(), &desired());
    assert_eq!(facts(&ignored), ("ignored".into(), "not_reached".into()));
    assert_eq!(ignored["action"], outside["action"]);
    assert_eq!(ignored["plan_digest"], outside["plan_digest"]);

    // A tracked settings file that an ignore pattern matches is checked out into every
    // worktree; git reports it as not ignored.
    fs::create_dir_all(p.join(".claude")).unwrap();
    fs::write(p.join(".claude/settings.json"), "{}\n").unwrap();
    git(p, &["add", "-f", ".claude/settings.json"]);
    let tracked = plan(p, home.path(), &desired());
    assert_eq!(
        facts(&tracked),
        ("not_ignored".into(), "reached_when_committed".into())
    );

    // The caller's GIT_DIR does not redirect the check to another repository.
    let other = TempDir::new().unwrap();
    git(other.path(), &["init", "-q"]);
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(["claude-settings-plan", p.to_str().unwrap()])
        .env("HOME", home.path())
        .env_remove("CLAUDE_CONFIG_DIR")
        .env("GIT_DIR", other.path().join(".git"))
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, desired().to_string().as_bytes());
    let v: Value = serde_json::from_slice(&child.wait_with_output().unwrap().stdout).unwrap();
    assert_eq!(v["result"]["settings_ignored"], json!("not_ignored"), "{v}");
}
