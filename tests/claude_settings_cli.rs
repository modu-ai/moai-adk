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

    common::git::git_ok(p, &["init", "-q"]);
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
    common::git::git_ok(p, &["add", "-f", ".claude/settings.json"]);
    let tracked = plan(p, home.path(), &desired());
    assert_eq!(
        facts(&tracked),
        ("not_ignored".into(), "reached_when_committed".into())
    );

    // The caller's GIT_DIR does not redirect the check to another repository.
    let other = TempDir::new().unwrap();
    common::git::git_ok(other.path(), &["init", "-q"]);
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

/// Runs `claude-settings-plan` / `-apply` with extra flags (e.g. `--scope local`).
fn run_with(
    op: &str,
    project: &Path,
    flags: &[&str],
    home: &Path,
    stdin: &Value,
) -> (Option<i32>, Value) {
    let mut args = vec![op, project.to_str().unwrap()];
    args.extend_from_slice(flags);
    run(&args, home, stdin)
}

fn sl() -> Value {
    json!({"statusLine": {"type": "command", "command": "moai-cli statusline"}})
}

/// #2581: `--scope local` plans and applies `.claude/settings.local.json` — the file Claude
/// Code also reads for worktree sessions — with its own receipt. Other keys in the local file
/// (`env`, user permissions) are kept, its 0600 mode stays, `settings.json` is not touched,
/// and a re-plan is a no-op. The default scope is unchanged (`project`).
#[test]
fn local_scope_writes_the_local_file_with_its_own_receipt() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    fs::create_dir_all(p.join(".claude")).unwrap();
    let local = p.join(".claude/settings.local.json");
    fs::write(&local, "{\n  \"env\": {\"X\": \"1\"}\n}\n").unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&local, fs::Permissions::from_mode(0o600)).unwrap();
    }

    let (code, v) = run_with(
        "claude-settings-plan",
        p,
        &["--scope", "local"],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(0), "{v}");
    let plan = &v["result"];
    assert_eq!(plan["scope"], json!("local"));
    assert_eq!(plan["settings_path"], json!(".claude/settings.local.json"));
    assert_eq!(plan["action"], json!("update"));
    assert_eq!(
        plan["worktree_sessions"],
        json!("reached_from_main_checkout")
    );
    let digest = plan["plan_digest"].as_str().unwrap().to_string();

    let (code, v) = run_with(
        "claude-settings-apply",
        p,
        &[
            "--scope",
            "local",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-2581",
        ],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(
        v["result"]["files_changed"],
        json!([".claude/settings.local.json"])
    );
    let written: Value = serde_json::from_slice(&fs::read(&local).unwrap()).unwrap();
    assert_eq!(
        written["env"],
        json!({"X": "1"}),
        "other local keys are kept"
    );
    assert_eq!(written["statusLine"], sl()["statusLine"]);
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        assert_eq!(
            fs::metadata(&local).unwrap().permissions().mode() & 0o777,
            0o600
        );
    }
    assert!(
        !p.join(".claude/settings.json").exists(),
        "the project file is untouched"
    );
    assert!(p
        .join(".moai-cli/manifests/claude-settings-local.json")
        .is_file());
    assert!(!p.join(".moai-cli/manifests/claude-settings.json").exists());

    let (_, again) = run_with(
        "claude-settings-plan",
        p,
        &["--scope", "local"],
        home.path(),
        &sl(),
    );
    assert_eq!(again["result"]["action"], json!("unchanged"), "{again}");
    let default = plan_of(p, home.path());
    assert_eq!(default["scope"], json!("project"));
    assert_eq!(default["settings_path"], json!(".claude/settings.json"));
}

fn plan_of(p: &Path, home: &Path) -> Value {
    plan(p, home, &sl())
}

/// #2581: a local settings file the apply creates is owner-only (0600), like Claude Code's own.
#[cfg(unix)]
#[test]
fn a_created_local_file_is_owner_only() {
    use std::os::unix::fs::PermissionsExt;
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let (_, v) = run_with(
        "claude-settings-plan",
        p,
        &["--scope", "local"],
        home.path(),
        &sl(),
    );
    assert_eq!(v["result"]["action"], json!("create"), "{v}");
    let digest = v["result"]["plan_digest"].as_str().unwrap().to_string();
    let (code, v) = run_with(
        "claude-settings-apply",
        p,
        &[
            "--scope",
            "local",
            "--plan-digest",
            &digest,
            "--permission-reference",
            "test-2581",
        ],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(0), "{v}");
    let mode = fs::metadata(p.join(".claude/settings.local.json"))
        .unwrap()
        .permissions()
        .mode();
    assert_eq!(mode & 0o777, 0o600);
}

/// #2581 boundary: with no file and no receipt on either side, the project and local plans
/// still differ in digest, so a digest planned for one scope never applies the other.
#[test]
fn the_plan_digest_binds_the_scope() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let project_digest = plan_of(p, home.path())["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    let (_, v) = run_with(
        "claude-settings-plan",
        p,
        &["--scope", "local"],
        home.path(),
        &sl(),
    );
    let local_digest = v["result"]["plan_digest"].as_str().unwrap().to_string();
    assert_ne!(project_digest, local_digest);
    let (code, v) = run_with(
        "claude-settings-apply",
        p,
        &[
            "--scope",
            "local",
            "--plan-digest",
            &project_digest,
            "--permission-reference",
            "test-2581",
        ],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(1), "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("stale or altered plan"),
        "{v}"
    );
    assert!(!p.join(".claude").exists(), "nothing written");
    let (code, _) = run_with(
        "claude-settings-plan",
        p,
        &["--scope", "home"],
        home.path(),
        &sl(),
    );
    assert_ne!(code, Some(0), "an unknown scope is refused");
}

/// #2581: in a linked worktree the local scope is refused and names the main checkout —
/// Claude Code reads settings.local.json there — and nothing is written in the worktree.
#[test]
fn local_scope_in_a_linked_worktree_names_the_main_checkout() {
    let (home, main) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let m = main.path();
    common::git::git_ok(m, &["init", "-q"]);
    fs::write(m.join("README.md"), "x\n").unwrap();
    common::git::git_ok(m, &["add", "README.md"]);
    common::git::git_ok(
        m,
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
    let wt = m.join("wt");
    common::git::git_ok(m, &["worktree", "add", "-q", wt.to_str().unwrap()]);
    let (code, v) = run_with(
        "claude-settings-plan",
        &wt,
        &["--scope", "local"],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(1), "{v}");
    let msg = v["error"]["message"].as_str().unwrap();
    assert!(
        msg.contains("local scope refused in a linked worktree"),
        "{msg}"
    );
    let canon = fs::canonicalize(m).unwrap();
    assert!(msg.contains(canon.to_str().unwrap()), "{msg}");
    assert!(!wt.join(".claude").exists());
    // The project scope in the same worktree still plans as before.
    let (code, v) = run_with("claude-settings-plan", &wt, &[], home.path(), &sl());
    assert_eq!(code, Some(0), "{v}");
}

/// #2383 range 2 (leader decision 6846938102): a `.claude/settings.json` with a user-set
/// `statusLine` MoAI never owned.
fn user_status_line(p: &Path) -> Vec<u8> {
    fs::create_dir_all(p.join(".claude")).unwrap();
    let bytes = b"{\n  \"model\": \"user-choice\",\n  \"statusLine\": {\n    \"type\": \"command\",\n    \"command\": \"orca statusline\"\n  }\n}\n".to_vec();
    fs::write(p.join(".claude/settings.json"), &bytes).unwrap();
    bytes
}

fn backups(p: &Path) -> Vec<std::path::PathBuf> {
    match fs::read_dir(p.join(".moai-cli/settings-backups")) {
        Err(_) => vec![],
        Ok(rd) => {
            let mut v: Vec<_> = rd.map(|e| e.unwrap().path()).collect();
            v.sort();
            v
        }
    }
}

/// #2383: a conflict on a key the caller may set previews a confirmed replacement — the key,
/// its current and new value, the backup folder and a `confirm_token` — and still writes
/// nothing; the plain apply still refuses.
#[test]
fn a_conflicting_key_previews_a_confirmed_replacement_and_writes_nothing() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let before = user_status_line(p);
    let r = plan(p, home.path(), &sl());
    assert_eq!(r["action"], json!("conflict"), "{r}");
    assert_eq!(r["conflicts"], json!(["statusLine"]));
    let c = &r["confirm_replace"];
    assert_eq!(c["keys"], json!(["statusLine"]), "{r}");
    assert_eq!(
        c["current"]["statusLine"],
        json!({"type": "command", "command": "orca statusline"})
    );
    assert_eq!(c["replace_with"]["statusLine"], sl()["statusLine"]);
    assert_eq!(c["backup_root"], json!(".moai-cli/settings-backups"));
    let token = c["confirm_token"].as_str().unwrap();
    assert!(token.starts_with("settings-replace:project:"), "{token}");
    assert_eq!(
        token.len(),
        "settings-replace:project:".len() + 16,
        "{token}"
    );
    assert_eq!(r["writes"], json!(false));

    let (code, v) = apply(p, home.path(), &sl(), r["plan_digest"].as_str().unwrap());
    assert_eq!(code, Some(1), "{v}");
    assert!(v["error"]["message"].as_str().unwrap().contains("conflict"));
    assert_eq!(fs::read(p.join(".claude/settings.json")).unwrap(), before);
    assert!(backups(p).is_empty());
    assert!(!p.join(".moai-cli/manifests/claude-settings.json").exists());
}

fn apply_replace(p: &Path, home: &Path, digest: &str, token: &str) -> (Option<i32>, Value) {
    run_with(
        "claude-settings-apply",
        p,
        &[
            "--plan-digest",
            digest,
            "--permission-reference",
            "test-2383",
            "--confirm-replace",
            token,
        ],
        home,
        &sl(),
    )
}

/// #2383: with the token, the original file is kept byte for byte in a new backup folder,
/// only the conflicting key is replaced (other user keys stay), MoAI owns it from now on,
/// the answer names the backup and the command that puts it back, and a re-plan is a no-op.
/// Putting the backup back restores the user's file and the plan offers the replacement again.
#[test]
fn the_confirmed_replacement_keeps_the_original_then_replaces_only_that_key() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let before = user_status_line(p);
    let r = plan(p, home.path(), &sl());
    let token = r["confirm_replace"]["confirm_token"].as_str().unwrap();
    let (code, v) = apply_replace(p, home.path(), r["plan_digest"].as_str().unwrap(), token);
    assert_eq!(code, Some(0), "{v}");
    let out = &v["result"];
    assert_eq!(out["state"], json!("replaced"), "{out}");
    assert_eq!(out["replaced_keys"], json!(["statusLine"]));

    let dirs = backups(p);
    assert_eq!(dirs.len(), 1, "{dirs:?}");
    let name = dirs[0].file_name().unwrap().to_str().unwrap().to_string();
    assert!(name.ends_with("-project"), "{name}");
    assert_eq!(fs::read(dirs[0].join("settings.json")).unwrap(), before);
    let rel = format!(".moai-cli/settings-backups/{name}");
    assert_eq!(out["backup_dir"], json!(rel));
    let revert = out["revert"].as_str().unwrap();
    assert!(
        revert.contains(&format!("{rel}/settings.json"))
            && revert.contains(".claude/settings.json"),
        "{revert}"
    );

    let settings: Value =
        serde_json::from_slice(&fs::read(p.join(".claude/settings.json")).unwrap()).unwrap();
    assert_eq!(
        settings["model"],
        json!("user-choice"),
        "other user key kept"
    );
    assert_eq!(settings["statusLine"], sl()["statusLine"]);
    let receipt: Value = serde_json::from_slice(
        &fs::read(p.join(".moai-cli/manifests/claude-settings.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(
        receipt["owned"]["statusLine"],
        sl()["statusLine"],
        "{receipt}"
    );
    assert_eq!(plan(p, home.path(), &sl())["action"], json!("unchanged"));

    // The revert the answer names: copy the backup back. The user's value then counts as a
    // change made after MoAI's (the three-way rule, user drift): the same request leaves it as
    // it is and writes nothing.
    fs::copy(
        dirs[0].join("settings.json"),
        p.join(".claude/settings.json"),
    )
    .unwrap();
    assert_eq!(fs::read(p.join(".claude/settings.json")).unwrap(), before);
    let again = plan(p, home.path(), &sl());
    assert_eq!(again["action"], json!("unchanged"), "{again}");
    assert!(again.get("confirm_replace").is_none(), "{again}");
    let (code, v) = apply(
        p,
        home.path(),
        &sl(),
        again["plan_digest"].as_str().unwrap(),
    );
    assert_eq!(code, Some(0), "{v}");
    assert_eq!(v["result"]["state"], json!("unchanged"));
    assert_eq!(fs::read(p.join(".claude/settings.json")).unwrap(), before);
}

/// #2383: a request with a conflicting key and a new key replaces the one and adds the other,
/// exactly as a plan without the conflict would; the file and the receipt agree on both.
#[test]
fn a_confirmed_replacement_also_applies_the_requests_other_keys() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    user_status_line(p);
    let both = desired();
    let r = plan(p, home.path(), &both);
    assert_eq!(r["confirm_replace"]["keys"], json!(["statusLine"]), "{r}");
    let token = r["confirm_replace"]["confirm_token"].as_str().unwrap();
    let (code, v) = run_with(
        "claude-settings-apply",
        p,
        &[
            "--plan-digest",
            r["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test-2383",
            "--confirm-replace",
            token,
        ],
        home.path(),
        &both,
    );
    assert_eq!(code, Some(0), "{v}");
    let settings: Value =
        serde_json::from_slice(&fs::read(p.join(".claude/settings.json")).unwrap()).unwrap();
    assert_eq!(settings["outputStyle"], both["outputStyle"], "{settings}");
    assert_eq!(settings["statusLine"], both["statusLine"]);
    assert_eq!(settings["model"], json!("user-choice"));
    assert_eq!(
        r["confirm_replace"]["planned_sha256"],
        v["result"]["settings_sha256_after"]
    );
    let receipt: Value = serde_json::from_slice(
        &fs::read(p.join(".moai-cli/manifests/claude-settings.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(receipt["owned"]["outputStyle"], both["outputStyle"]);
    assert_eq!(receipt["owned"]["statusLine"], both["statusLine"]);
    assert_eq!(plan(p, home.path(), &both)["action"], json!("unchanged"));
}

/// #2383: a token the plan did not offer, or one from before the file changed, refuses the
/// whole apply before any write — no backup folder, the file and receipt as they were.
#[test]
fn a_wrong_or_stale_token_refuses_before_any_write() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    let before = user_status_line(p);
    let r = plan(p, home.path(), &sl());
    let digest = r["plan_digest"].as_str().unwrap();
    let token = r["confirm_replace"]["confirm_token"].as_str().unwrap();

    for bad in ["settings-replace:project:0000000000000000", "x", ""] {
        let (code, v) = apply_replace(p, home.path(), digest, bad);
        assert_eq!(code, Some(1), "{bad}: {v}");
        assert!(
            v["error"]["message"].as_str().unwrap().contains("confirm"),
            "{bad}: {v}"
        );
    }
    assert_eq!(fs::read(p.join(".claude/settings.json")).unwrap(), before);
    assert!(backups(p).is_empty());

    // The user edits the key after the preview: the old digest and token no longer apply.
    let edited = String::from_utf8(before.clone())
        .unwrap()
        .replace("orca statusline", "orca statusline --wide");
    fs::write(p.join(".claude/settings.json"), &edited).unwrap();
    let (code, v) = apply_replace(p, home.path(), digest, token);
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(
        fs::read_to_string(p.join(".claude/settings.json")).unwrap(),
        edited
    );
    assert!(backups(p).is_empty());
    assert!(!p.join(".moai-cli/manifests/claude-settings.json").exists());

    // A fresh preview of the edited file with the old confirmation: the user confirmed
    // replacing a different current value, so the old token is refused even with the new
    // digest, and the new preview offers a new token.
    let fresh_plan = plan(p, home.path(), &sl());
    let new_token = fresh_plan["confirm_replace"]["confirm_token"]
        .as_str()
        .unwrap();
    assert_ne!(new_token, token);
    let (code, v) = apply_replace(
        p,
        home.path(),
        fresh_plan["plan_digest"].as_str().unwrap(),
        token,
    );
    assert_eq!(code, Some(1), "{v}");
    assert_eq!(
        fs::read_to_string(p.join(".claude/settings.json")).unwrap(),
        edited
    );
    assert!(backups(p).is_empty());

    // A token on a plan without any conflict is refused too.
    let fresh = TempDir::new().unwrap();
    let f = plan(fresh.path(), home.path(), &sl());
    assert_eq!(f["action"], json!("create"));
    assert!(f.get("confirm_replace").is_none(), "{f}");
    let (code, v) = apply_replace(
        fresh.path(),
        home.path(),
        f["plan_digest"].as_str().unwrap(),
        token,
    );
    assert_eq!(code, Some(1), "{v}");
    assert!(!fresh.path().join(".claude").exists());
}

/// #2383: the local scope replaces in `.claude/settings.local.json`, keeps that file in its
/// own backup folder, and its token names the scope.
#[test]
fn the_local_scope_replaces_its_own_file() {
    let (home, project) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let p = project.path();
    fs::create_dir_all(p.join(".claude")).unwrap();
    let before =
        b"{\n  \"statusLine\": {\n    \"type\": \"command\",\n    \"command\": \"mine\"\n  }\n}\n";
    fs::write(p.join(".claude/settings.local.json"), before).unwrap();
    let (code, r) = run_with(
        "claude-settings-plan",
        p,
        &["--scope", "local"],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(0), "{r}");
    let r = &r["result"];
    let token = r["confirm_replace"]["confirm_token"].as_str().unwrap();
    assert!(token.starts_with("settings-replace:local:"), "{token}");
    let (code, v) = run_with(
        "claude-settings-apply",
        p,
        &[
            "--plan-digest",
            r["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test-2383",
            "--scope",
            "local",
            "--confirm-replace",
            token,
        ],
        home.path(),
        &sl(),
    );
    assert_eq!(code, Some(0), "{v}");
    let dirs = backups(p);
    assert_eq!(dirs.len(), 1);
    assert!(dirs[0].to_str().unwrap().ends_with("-local"));
    assert_eq!(
        fs::read(dirs[0].join("settings.local.json")).unwrap(),
        before
    );
    assert!(!p.join(".claude/settings.json").exists());
    let local: Value =
        serde_json::from_slice(&fs::read(p.join(".claude/settings.local.json")).unwrap()).unwrap();
    assert_eq!(local["statusLine"], sl()["statusLine"]);
}
