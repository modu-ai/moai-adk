//! #2936 R-fix R-B (PLAN 6880040242, O-1): `project-update-plan` has an install step `hooks`
//! that replaces the Go-generation hook entries in `.claude/settings.json`. It prints
//! `go-hooks-apply` when there is something to replace, waits for a decision without a
//! command when the replacement is blocked, and reads `unchanged` when nothing of the Go
//! generation is left. Every case builds a temporary project with an empty temporary HOME.

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{install_step, GoFx};
use serde_json::{json, Value};
use std::fs;
use std::process::{Command, Stdio};

const PLACEHOLDER: &str = "<permission-reference>";
const INSTALL_ORDER: [&str; 5] = [
    "profile",
    "generation",
    "claude-settings",
    "hooks",
    "github",
];

const DENY: [&str; 2] = ["Bash(rm -rf:*)", "Read(./.env)"];

/// A settings file whose hooks run the named Go scripts, plus the deny rules a user keeps.
fn settings_body(scripts: &[&str], other: bool) -> String {
    let mut handlers: Vec<Value> = scripts
        .iter()
        .map(|s| json!({"type": "command", "command": format!("bash .claude/hooks/moai/{s}")}))
        .collect();
    if other {
        handlers.push(json!({"type": "command", "command": "/usr/local/bin/other-tool"}));
    }
    json!({
        "hooks": {"PreToolUse": [{"matcher": "*", "hooks": handlers}]},
        "permissions": {"deny": DENY},
    })
    .to_string()
}

fn project_with_go_hooks() -> GoFx {
    let fx = GoFx::new();
    fx.write(".claude/hooks/moai/handle-a.sh", "#!/bin/sh\necho a\n");
    fx.write(".claude/hooks/moai/handle-b.sh", "#!/bin/sh\necho b\n");
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh", "handle-b.sh"], false),
    );
    fx.with_profile();
    fx
}

fn hooks(plan: &Value) -> &Value {
    install_step(plan, "hooks").unwrap_or_else(|| panic!("no hooks step: {plan}"))
}

/// The preview a user sees once the `claude-settings` step is applied: that step and `hooks`
/// rewrite the same file, so `hooks` is planned only against the file the settings step wrote
/// (#3028 audit B1).
fn after_settings(fx: &GoFx) -> Value {
    let plan = fx.update_plan();
    let step = settings(&plan);
    if step["state"] == "ready" {
        let out = fx.apply_step(step);
        assert_eq!(out["ok"], json!(true), "{out}");
    }
    fx.update_plan()
}

#[test]
fn the_install_phase_lists_hooks_after_the_claude_settings() {
    let fx = GoFx::new();
    fx.with_profile();
    let plan = fx.update_plan();
    let names: Vec<&str> = plan["phases"][1]["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["step"].as_str().unwrap())
        .collect();
    // The host steps (#2936 R9) follow `github`; this pin is the order up to there.
    assert!(names.len() >= INSTALL_ORDER.len(), "{plan}");
    assert_eq!(names[..INSTALL_ORDER.len()], INSTALL_ORDER, "{plan}");
}

/// SPEC-20261119-hooks-followups AC-1 (N1, leader decision B): a profile that lists `claude` and
/// a project with no hooks at all (no settings file, no record, no move) keeps the step
/// `unchanged` with no command (the rule of the settings-move receipt stays), but its reason
/// names the gap and the command to run by hand, and `preview.configured` reads as `hooks-status`
/// does. Before, the step said only "no Go-generation hooks are left".
#[test]
fn a_project_with_nothing_yet_names_the_gap_and_prints_no_command() {
    let fx = GoFx::new();
    fx.with_profile();
    let plan = after_settings(&fx);
    let own = hooks_plan(&fx);
    // The settings step wrote the file first, so the plan reads `update` (it reads `create` only
    // where no file exists, a state the step sequence never leaves behind).
    assert_eq!(own["action"], "update", "{own}");
    let step = hooks(&plan);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(reason.contains("no MoAI hook set is installed"), "{step}");
    assert!(reason.contains("hooks-plan reads update"), "{step}");
    assert!(reason.contains("moai hooks-apply"), "{step}");
    assert!(reason.contains("--host claude"), "{step}");
    assert_eq!(
        step["preview"]["configured"], own["hook_status"]["configured"],
        "{step}"
    );
    // The command it names does install the set, and the step then reads current.
    let applied = fx.run_stdin(
        &[
            "hooks-apply",
            fx.p(),
            "--host",
            "claude",
            "--plan-digest",
            own["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            go_project_real_run::REF,
        ],
        b"",
    );
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert_eq!(
        normalize_commands(settings_now(&fx)["hooks"].clone()),
        generated()
    );
    let after = fx.update_plan();
    assert_eq!(hooks(&after)["state"], "unchanged");
    assert_eq!(hooks(&after)["reason"], "the MoAI hook set is current");
}

#[test]
fn go_hook_entries_make_a_ready_step_that_prints_go_hooks_apply() {
    let fx = project_with_go_hooks();
    let plan = after_settings(&fx);
    let step = hooks(&plan);
    assert_eq!(step["state"], "ready", "{step}");
    let own = fx.call(&["go-hooks-plan", fx.p()], b"");
    assert_eq!(step["plan_digest"], own["plan_digest"], "{step}");
    assert_eq!(
        step["apply"]["argv"],
        json!([
            "moai",
            "go-hooks-apply",
            fx.p(),
            "--plan-digest",
            own["plan_digest"],
            "--permission-reference",
            PLACEHOLDER
        ]),
        "{step}"
    );
    assert_eq!(step["apply"]["stdin"], Value::Null, "{step}");
    assert_eq!(
        step["preview"]["go_entries"].as_array().unwrap().len(),
        2,
        "{step}"
    );

    let out = fx.apply_step(step);
    assert_eq!(out["ok"], json!(true), "{out}");
    // O-1: the hook entries change, `permissions` is not touched.
    let after: Value =
        serde_json::from_slice(&fs::read(fx.root.join(".claude/settings.json")).unwrap()).unwrap();
    assert_eq!(after["permissions"]["deny"], json!(DENY), "{after}");
    assert!(!after.to_string().contains(".claude/hooks/moai"), "{after}");
    let again = fx.update_plan();
    assert_eq!(hooks(&again)["state"], "unchanged", "{again}");
}

#[test]
fn hooks_of_other_tools_block_the_step_and_print_no_command() {
    let fx = GoFx::new();
    fx.write(".claude/hooks/moai/handle-a.sh", "#!/bin/sh\n");
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], true),
    );
    fx.with_profile();
    let plan = fx.update_plan();
    let step = hooks(&plan);
    assert_eq!(step["state"], "needs_decision", "{step}");
    assert!(
        step["reason"]
            .as_str()
            .unwrap()
            .contains("other_hooks_present"),
        "{step}"
    );
    assert_eq!(step["apply"], Value::Null, "{step}");
}

#[test]
fn without_a_profile_the_hooks_step_waits_with_the_other_install_steps() {
    let fx = GoFx::new();
    fx.write(".claude/hooks/moai/handle-a.sh", "#!/bin/sh\n");
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], false),
    );
    let plan = fx.update_plan();
    let step = hooks(&plan);
    assert_eq!(step["state"], "needs_decision", "{step}");
    assert_eq!(step["reason"], "run /moai project", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
}

/// AC-3: the Go project from the real run. Before the move every install step waits; after
/// the move and a profile, the scripts the settings still name get their replacement.
#[test]
fn after_the_move_the_hooks_step_prints_go_hooks_apply() {
    let mut fx = GoFx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    for n in ["handle-a", "handle-b", "handle-c"] {
        fx.deployed(
            &format!(".claude/hooks/moai/{n}.sh"),
            &format!("#!/bin/sh\necho {n}\n"),
        );
    }
    fx.deployed(".claude/workflows/w1.md", "# w1\n");
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], false),
    );

    let before = fx.update_plan();
    let step = hooks(&before);
    assert_eq!(step["state"], "blocked", "{step}");
    assert_eq!(step["reason"], "re-run after the backup phase", "{step}");

    let plan = fx.plan();
    let applied = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert_eq!(applied["result"]["state"], "committed", "{applied}");
    assert!(
        !fx.root.join(".claude/hooks/moai/handle-b.sh").exists(),
        "the unreferenced script moved"
    );
    assert!(
        fx.root.join(".claude/hooks/moai/handle-a.sh").exists(),
        "the referenced script stayed"
    );

    fx.with_profile();
    let after = after_settings(&fx);
    let step = hooks(&after);
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["apply"]["argv"][1], "go-hooks-apply", "{step}");
    assert_eq!(
        step["preview"]["go_entries"].as_array().unwrap().len(),
        1,
        "{step}"
    );
    let out = fx.apply_step(step);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(hooks(&fx.update_plan())["state"], "unchanged");
}

fn settings(plan: &Value) -> &Value {
    install_step(plan, "claude-settings")
        .unwrap_or_else(|| panic!("no claude-settings step: {plan}"))
}

/// #3028 audit B1: `claude-settings` and `hooks` rewrite the same `.claude/settings.json`, and
/// the `go-hooks-apply` digest binds that file's bytes. Applying both from one preview in the
/// printed order made the second command `stale_plan` every time, so the `hooks` step waits
/// for a pending `claude-settings` apply and is planned again afterwards.
#[test]
fn hooks_wait_for_a_pending_claude_settings_apply() {
    let fx = project_with_go_hooks();
    let plan = fx.update_plan();
    let first = settings(&plan);
    assert_eq!(first["state"], "ready", "the precondition: {first}");
    let step = hooks(&plan);
    assert_eq!(step["state"], "blocked", "{step}");
    assert_eq!(
        step["reason"], "re-run after the claude-settings apply",
        "{step}"
    );
    assert_eq!(step["apply"], Value::Null, "{step}");
    assert_eq!(step["plan_digest"], Value::Null, "{step}");

    let out = fx.apply_step(first);
    assert_eq!(out["ok"], json!(true), "{out}");
    let again = fx.update_plan();
    assert_ne!(settings(&again)["state"], "ready", "{}", settings(&again));
    let step = hooks(&again);
    assert_eq!(step["state"], "ready", "{step}");
    let out = fx.apply_step(step);
    assert_eq!(
        out["ok"],
        json!(true),
        "the planned-again digest applies: {out}"
    );
    assert_eq!(hooks(&fx.update_plan())["state"], "unchanged");
}

/// The profile-error path builds its own step list; the same wait holds there.
#[test]
fn hooks_wait_for_the_settings_apply_when_the_profile_is_unreadable() {
    let fx = project_with_go_hooks();
    let stored = [".moai/project.yaml", ".moai-cli/project.yaml"]
        .into_iter()
        .find(|p| fx.root.join(p).is_file())
        .expect("the stored profile");
    fx.write(stored, "{ not a profile");
    let plan = fx.update_plan();
    let first = settings(&plan);
    assert!(first["apply"].is_object(), "the precondition: {first}");
    let step = hooks(&plan);
    assert_eq!(step["state"], "blocked", "{step}");
    assert_eq!(
        step["reason"], "re-run after the claude-settings apply",
        "{step}"
    );
    assert_eq!(step["apply"], Value::Null, "{step}");
}

// ---- C-55: the ownership record outlives the settings file it describes ----

fn generated() -> Value {
    normalize_commands(Value::Object(
        moai_cli::hooks::hooks_object("claude").unwrap(),
    ))
}

/// The exec-form entries name the absolute path of the binary that rendered them
/// (SPEC-20261122-next-flow-hooks REQ-5): an in-test render names this test's executable
/// while the file was written by the spawned `moai`, so comparisons normalize every
/// command to its last path segment.
fn normalize_commands(v: Value) -> Value {
    match v {
        Value::Object(mut m) => {
            if m.contains_key("command") && m.contains_key("args") {
                // A generated entry: the absolute path names whichever binary rendered it
                // (the spawned `moai` for files, this test executable in-process), so both
                // sides become the same constant token.
                m.insert("command".into(), json!("moai"));
            }
            Value::Object(
                m.into_iter()
                    .map(|(k, v)| (k, normalize_commands(v)))
                    .collect(),
            )
        }
        Value::Array(a) => Value::Array(a.into_iter().map(normalize_commands).collect()),
        other => other,
    }
}

/// The settings file's `hooks`, normalized the same way.
fn settings_hooks_normalized(fx: &GoFx) -> Value {
    normalize_commands(settings_now(fx)["hooks"].clone())
}

fn settings_now(fx: &GoFx) -> Value {
    serde_json::from_slice(&fs::read(fx.root.join(".claude/settings.json")).unwrap()).unwrap()
}

fn hooks_plan(fx: &GoFx) -> Value {
    fx.call(&["hooks-plan", fx.p(), "--host", "claude"], b"")
}

/// A project whose hooks step ran once (the ownership record names the MoAI hook set) and
/// whose settings file and scripts were then put back as they were before it, the way a
/// `git reset` restores tracked files and leaves the untracked record. Returns the restored
/// settings bytes.
fn reset_after_a_hooks_step(fx: &GoFx) -> Vec<u8> {
    let plan = after_settings(fx);
    let old = fs::read(fx.root.join(".claude/settings.json")).unwrap();
    let out = fx.apply_step(hooks(&plan));
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(
        settings_hooks_normalized(fx),
        generated(),
        "the first apply"
    );
    fs::write(fx.root.join(".claude/settings.json"), &old).unwrap();
    fx.write(".claude/hooks/moai/handle-a.sh", "#!/bin/sh\necho a\n");
    fx.write(".claude/hooks/moai/handle-b.sh", "#!/bin/sh\necho b\n");
    assert!(
        [".moai", ".moai-cli"].iter().any(|state| fx
            .root
            .join(state)
            .join("manifests/claude-settings.json")
            .is_file()),
        "the record stays"
    );
    old
}

/// The reported sequence: the record says the MoAI hooks are installed, the file holds the Go
/// entries again. The step must end with the previewed hook set in the file, the plan and the
/// status must agree afterwards, and the receipt's revert must put the scripts and the
/// entries back.
#[test]
fn a_stale_record_does_not_keep_the_hooks_step_from_writing_the_hook_set() {
    let fx = project_with_go_hooks();
    let old = reset_after_a_hooks_step(&fx);

    let plan = after_settings(&fx);
    let step = hooks(&plan);
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["apply"]["argv"][1], "go-hooks-apply", "{step}");
    let before_apply = fs::read(fx.root.join(".claude/settings.json")).unwrap();
    let out = fx.apply_step(step);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(out["result"]["state"], "applied", "{out}");

    let after = settings_now(&fx);
    assert_eq!(
        normalize_commands(after["hooks"].clone()),
        generated(),
        "an applied hooks step leaves the previewed MoAI hook set in the file: {after}"
    );
    assert_eq!(after["permissions"]["deny"], json!(DENY), "{after}");
    assert!(!fx.root.join(".claude/hooks/moai").exists());
    let own = hooks_plan(&fx);
    assert_eq!(own["action"], "unchanged", "{own}");
    assert_eq!(own["hook_status"]["configured"], "yes", "{own}");
    assert_eq!(hooks(&fx.update_plan())["state"], "unchanged");

    let receipt = out["result"]["receipt"].as_str().unwrap();
    let back = fx.call(
        &[
            "go-hooks-revert",
            fx.p(),
            "--receipt",
            receipt,
            "--permission-reference",
            go_project_real_run::REF,
        ],
        b"",
    );
    assert_eq!(back["state"], "reverted", "{back}");
    assert_eq!(back["scripts_restored"], json!(2), "{back}");
    assert_eq!(
        fs::read(fx.root.join(".claude/settings.json")).unwrap(),
        before_apply
    );
    assert_eq!(before_apply, old, "no settings apply ran in between");
    for (name, body) in [
        ("handle-a.sh", "#!/bin/sh\necho a\n"),
        ("handle-b.sh", "#!/bin/sh\necho b\n"),
    ] {
        assert_eq!(
            fs::read_to_string(fx.root.join(".claude/hooks/moai").join(name)).unwrap(),
            body
        );
    }
}

/// The state the reported run ended in: no Go entry, no `hooks` key, a record that names the
/// MoAI hook set. Neither plan may read `unchanged` while the status reads `drifted`; the step
/// prints `hooks-apply`, and applying it writes the set.
#[test]
fn a_file_without_the_recorded_hooks_reads_ready_and_agrees_with_the_status() {
    let fx = project_with_go_hooks();
    let plan = after_settings(&fx);
    let out = fx.apply_step(hooks(&plan));
    assert_eq!(out["ok"], json!(true), "{out}");
    let mut s = settings_now(&fx);
    s.as_object_mut().unwrap().remove("hooks");
    fx.write(
        ".claude/settings.json",
        &serde_json::to_string_pretty(&s).unwrap(),
    );

    let own = hooks_plan(&fx);
    assert_eq!(own["hook_status"]["configured"], "drifted", "{own}");
    assert_eq!(own["action"], "update", "{own}");
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["plan_digest"], own["plan_digest"], "{step}");
    assert_eq!(
        step["apply"]["argv"],
        json!([
            "moai",
            "hooks-apply",
            fx.p(),
            "--host",
            "claude",
            "--plan-digest",
            own["plan_digest"],
            "--permission-reference",
            PLACEHOLDER
        ]),
        "{step}"
    );

    let out = fx.apply_step(step);
    assert_eq!(out["ok"], json!(true), "{out}");
    let after = settings_now(&fx);
    assert_eq!(
        normalize_commands(after["hooks"].clone()),
        generated(),
        "{after}"
    );
    assert_eq!(after["permissions"]["deny"], json!(DENY), "{after}");
    let own = hooks_plan(&fx);
    assert_eq!(own["action"], "unchanged", "{own}");
    assert_eq!(own["hook_status"]["configured"], "yes", "{own}");
    assert_eq!(hooks(&fx.update_plan())["state"], "unchanged");
}

/// The other direction: a `hooks` key the file still has is not the lost-key case. A user's
/// edit of the current set is kept as before (the three-way rule `update` pins in
/// `a_hand_edit_of_a_current_hook_set_is_kept`): no command is printed and nothing is written.
#[test]
fn recorded_hooks_that_were_edited_are_kept_and_get_no_command() {
    let fx = project_with_go_hooks();
    let plan = after_settings(&fx);
    let out = fx.apply_step(hooks(&plan));
    assert_eq!(out["ok"], json!(true), "{out}");
    let mut s = settings_now(&fx);
    s["hooks"]["PostToolUse"] = json!([]);
    fx.write(
        ".claude/settings.json",
        &serde_json::to_string_pretty(&s).unwrap(),
    );
    let edited = fs::read(fx.root.join(".claude/settings.json")).unwrap();

    let own = hooks_plan(&fx);
    assert_eq!(own["hook_status"]["configured"], "drifted", "{own}");
    assert_eq!(own["action"], "unchanged", "{own}");
    assert_eq!(own["planned_sha256"], Value::Null, "{own}");
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
    let applied = fx.run_stdin(
        &[
            "hooks-apply",
            fx.p(),
            "--host",
            "claude",
            "--plan-digest",
            own["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            go_project_real_run::REF,
        ],
        b"",
    );
    assert_eq!(applied["result"]["state"], "unchanged", "{applied}");
    assert_eq!(
        fs::read(fx.root.join(".claude/settings.json")).unwrap(),
        edited
    );
}

// ---- SPEC-20261119-hooks-followups: what the step says when the file and the record disagree

/// The command `hooks` prints for the Claude host, for the plan digest of `own` (`hooks-plan`).
/// Runs `hooks-apply` as the step's own argv would: the spawned binary with its directory
/// first on PATH (the legacy gate needs the Rust `moai` first; the entries' absolute path
/// makes an in-process apply digest a different command than the spawned plan).
fn run_hooks_apply(fx: &GoFx, argv: &Value, digest: &str) -> Value {
    use common::CleanEnv as _;
    use std::io::Write as _;
    let bin_dir = std::path::Path::new(common::MOAI_BIN)
        .parent()
        .unwrap()
        .to_path_buf();
    // The argv holds the program name, the action, the project, the digest flag with the
    // PLAN's digest (the step's own digest may differ), the permission placeholder and more;
    // rebuild it with the step's digest and a real permission reference.
    let mut args: Vec<String> = Vec::new();
    let mut skip_next = false;
    for (n, word) in argv.as_array().unwrap().iter().enumerate() {
        if skip_next {
            skip_next = false;
            continue;
        }
        let w = word.as_str().unwrap_or_default();
        if n == 0 {
            continue; // the program name
        }
        if w == "--plan-digest" {
            args.push("--plan-digest".into());
            args.push(digest.to_string());
            skip_next = true;
            continue;
        }
        if word == &json!(PLACEHOLDER) {
            args.push(go_project_real_run::REF.to_string());
            continue;
        }
        args.push(w.to_string());
    }
    let arg_refs: Vec<&str> = args.iter().map(String::as_str).collect();
    let mut child = Command::new(common::MOAI_BIN)
        .args(&arg_refs)
        .clean_env()
        .env("HOME", &fx.home)
        .env("PATH", format!("{}:/usr/bin:/bin", bin_dir.display()))
        .env("LC_ALL", "C")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(b"");
    drop(pipe);
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "hooks-apply {arg_refs:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn hooks_apply_argv(fx: &GoFx, own: &Value) -> Value {
    json!([
        "moai",
        "hooks-apply",
        fx.p(),
        "--host",
        "claude",
        "--plan-digest",
        own["plan_digest"],
        "--permission-reference",
        PLACEHOLDER
    ])
}

fn write_settings(fx: &GoFx, value: &Value) {
    fx.write(
        ".claude/settings.json",
        &serde_json::to_string_pretty(value).unwrap(),
    );
}

/// Applies the hooks step once, so the record names the MoAI hook set; returns the record's path.
fn with_the_hook_set_installed(fx: &GoFx) -> std::path::PathBuf {
    let plan = after_settings(fx);
    let out = fx.apply_step(hooks(&plan));
    assert_eq!(out["ok"], json!(true), "{out}");
    record_path(fx)
}

fn record_path(fx: &GoFx) -> std::path::PathBuf {
    [".moai-cli", ".moai"]
        .iter()
        .map(|state| fx.root.join(state).join("manifests/claude-settings.json"))
        .find(|p| p.is_file())
        .expect("the ownership record")
}

/// A profile with the given hosts, written the way PROJECT writes it.
fn profile_with_hosts(fx: &GoFx, hosts: &[&str]) {
    let detect = fx.call(&["project-detect", fx.p()], b"");
    let profile = json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": hosts,
        "detected": detect["detected"],
    });
    let plan = fx.call(
        &["project-plan", fx.p()],
        &serde_json::to_vec(&profile).unwrap(),
    );
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    fx.call(
        &[
            "project-apply",
            fx.p(),
            "--plan-digest",
            &digest,
            "--permission-reference",
            go_project_real_run::REF,
        ],
        &serde_json::to_vec(&plan).unwrap(),
    );
}

/// N1: a settings file that exists but holds no `hooks` key, and no record: the step agrees with
/// `hooks-status` (`drifted`) and `hooks-plan` (`update`) in its words, and prints no command.
#[test]
fn a_file_without_hooks_and_without_a_record_says_drifted_like_the_status() {
    let fx = GoFx::new();
    fx.with_profile();
    write_settings(&fx, &json!({"permissions": {"deny": DENY}}));
    let plan = after_settings(&fx);
    let own = hooks_plan(&fx);
    assert_eq!(own["hook_status"]["configured"], "drifted", "{own}");
    assert_eq!(own["action"], "update", "{own}");
    let step = hooks(&plan);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
    assert_eq!(step["preview"]["configured"], "drifted", "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(reason.contains("hooks-plan reads update"), "{step}");
    assert!(reason.contains("moai hooks-plan"), "{step}");
    assert!(reason.contains("moai hooks-apply"), "{step}");
    assert!(!reason.contains("is current"), "{step}");
}

/// N1: the hooks plan is consulted only where the profile lists `claude`.
#[test]
fn a_profile_without_claude_leaves_the_hooks_step_unchanged_without_a_command() {
    let fx = GoFx::new();
    profile_with_hosts(&fx, &["codex"]);
    let plan = fx.update_plan();
    let step = hooks(&plan);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
}

/// N3: a user's own `hooks` key with no record is the three-way conflict of `hooks-plan`, never
/// "current"; without a move or a record the step stays `unchanged` and prints no command, and
/// says so.
#[test]
fn a_hooks_key_of_the_users_own_with_no_record_says_conflict_and_prints_no_command() {
    let fx = GoFx::new();
    fx.with_profile();
    write_settings(
        &fx,
        &json!({"hooks": {"Stop": [{"hooks": [{"type": "command", "command": "echo mine"}]}]}}),
    );
    let plan = after_settings(&fx);
    let own = hooks_plan(&fx);
    assert_eq!(own["action"], "conflict", "{own}");
    let step = hooks(&plan);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(reason.contains("hooks-plan reads conflict"), "{step}");
    assert!(!reason.contains("is current"), "{step}");
}

/// N3: an edit of the recorded set stays (state `unchanged`, no command) but the step does not
/// call it current: the status reads `drifted` and so does the step.
#[test]
fn a_hooks_key_that_is_not_the_moai_set_is_not_called_current() {
    let fx = project_with_go_hooks();
    with_the_hook_set_installed(&fx);
    let mut s = settings_now(&fx);
    s["hooks"]["PostToolUse"] = json!([]);
    write_settings(&fx, &s);
    assert_eq!(hooks_plan(&fx)["hook_status"]["configured"], "drifted");
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(!reason.contains("is current"), "{step}");
    assert!(reason.contains("not the MoAI hook set"), "{step}");
    assert_eq!(step["preview"]["configured"], "drifted", "{step}");
}

/// N3: `{}` and `null` leave no hook in effect, and nothing prints a command to change that: the
/// step waits for a decision and names the way out (the file is the user's).
#[test]
fn an_empty_or_null_hooks_key_reads_needs_decision_and_names_the_way_out() {
    for empty in [json!({}), Value::Null] {
        let fx = project_with_go_hooks();
        with_the_hook_set_installed(&fx);
        let mut s = settings_now(&fx);
        s["hooks"] = empty.clone();
        write_settings(&fx, &s);
        let update = fx.update_plan();
        let step = hooks(&update);
        assert_eq!(step["state"], "needs_decision", "{empty}: {step}");
        assert_eq!(step["apply"], Value::Null, "{empty}: {step}");
        let reason = step["reason"].as_str().unwrap();
        assert!(reason.contains("no hook is in effect"), "{empty}: {step}");
        assert!(reason.contains("remove the `hooks` key"), "{empty}: {step}");
        assert_eq!(step["preview"]["hooks_in_effect"], json!(false), "{step}");
    }
}

/// N4: the status line MoAI wrote is gone from the file while the record still owns it. The
/// plan keeps it removed (the user's edit), and the step says so instead of "current".
#[test]
fn a_status_line_removed_by_hand_is_named_not_called_current() {
    let fx = project_with_go_hooks();
    let plan = fx.update_plan();
    let first = settings(&plan);
    assert_eq!(first["state"], "ready", "the precondition: {first}");
    let out = fx.apply_step(first);
    assert_eq!(out["ok"], json!(true), "{out}");
    let mut s = settings_now(&fx);
    assert!(s["statusLine"].is_object(), "the precondition: {s}");
    // While the file still holds the status line, the step does not say it was removed.
    let settled = fx.update_plan();
    let held = settings(&settled);
    assert_eq!(held["state"], "unchanged", "{held}");
    assert!(
        !held["reason"].as_str().unwrap().contains("removed"),
        "{held}"
    );
    assert_eq!(
        held["preview"]["status_line_removed"],
        Value::Null,
        "{held}"
    );
    s.as_object_mut().unwrap().remove("statusLine");
    write_settings(&fx, &s);
    let update = fx.update_plan();
    let step = settings(&update);
    assert_eq!(step["state"], "unchanged", "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(reason.contains("status line"), "{step}");
    assert!(reason.contains("removed"), "{step}");
    assert!(reason.contains("not put back"), "{step}");
    assert_ne!(reason, "the MoAI-owned keys are current", "{step}");
    assert_eq!(
        step["preview"]["status_line_removed"],
        json!(true),
        "{step}"
    );
}

/// N5: an unreadable record blocks the step; the reason names the file and the next step.
#[test]
fn an_unreadable_record_names_its_file_and_the_next_step() {
    let fx = project_with_go_hooks();
    let record = with_the_hook_set_installed(&fx);
    fs::write(&record, "not json").unwrap();
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "blocked", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(reason.contains("manifests/claude-settings.json"), "{step}");
    assert!(reason.contains("no moai command repairs it"), "{step}");
    assert!(reason.contains("plan again"), "{step}");
}

/// The record and the file of a project that took an earlier release's hook set: the same
/// older set in both, written the way that release did (the `moai-cli` command names).
fn hold_an_older_set(fx: &GoFx) {
    let record = with_the_hook_set_installed(fx);
    let older = Value::Object(moai_cli::hooks::legacy_hooks("claude", None).unwrap());
    let mut s = settings_now(fx);
    s["hooks"] = older.clone();
    write_settings(fx, &s);
    let mut r: Value = serde_json::from_slice(&fs::read(&record).unwrap()).unwrap();
    r["owned"]["hooks"] = older;
    r["settings_sha256"] = json!(go_project_real_run::sha(
        &fs::read(fx.root.join(".claude/settings.json")).unwrap()
    ));
    fs::write(&record, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
}

/// N6: an upgrade where the file and the record hold the same older set is `ready`.
#[test]
fn an_older_hook_set_held_by_the_file_and_the_record_reads_ready() {
    let fx = project_with_go_hooks();
    hold_an_older_set(&fx);
    let own = hooks_plan(&fx);
    assert_eq!(own["action"], "update", "{own}");
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "ready", "{step}");
    assert_eq!(step["apply"]["argv"], hooks_apply_argv(&fx, &own), "{step}");
    // The writer runs as the command's own argv: the spawned binary, with its own directory
    // first on PATH so the `moai-cli` rewrite's gate sees the Rust `moai` (the exec-form
    // entries name the absolute path of the binary that renders them, so an in-process apply
    // would digest a different command than the spawned plan did).
    let argv = hooks_apply_argv(&fx, &own);
    let out = run_hooks_apply(&fx, &argv, step["plan_digest"].as_str().unwrap());
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(out["result"]["state"], "applied", "{out}");
    assert_eq!(
        normalize_commands(settings_now(&fx)["hooks"].clone()),
        generated()
    );
    assert_eq!(hooks(&fx.update_plan())["state"], "unchanged");
}

/// N6: the older set with a user's edit on top is a conflict, and nothing is printed.
#[test]
fn an_older_hook_set_with_a_users_edit_reads_conflict() {
    let fx = project_with_go_hooks();
    hold_an_older_set(&fx);
    let mut s = settings_now(&fx);
    s["hooks"]["PostToolUse"] = json!([]);
    write_settings(&fx, &s);
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "conflict", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
}

/// N6: `hooks-apply --host claude` over an existing file prints its own `hooks-revert`; the
/// printed undo order no longer says such a project has no command.
#[test]
fn the_undo_order_names_hooks_revert_for_a_hooks_apply_update() {
    let fx = project_with_go_hooks();
    let plan = fx.update_plan();
    let hooks_note = plan["undo_order"]["steps"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["step"] == "hooks")
        .unwrap()["note"]
        .as_str()
        .unwrap();
    assert!(hooks_note.contains("hooks-revert"), "{hooks_note}");
    assert!(!hooks_note.contains("has no command here"), "{hooks_note}");
    assert!(
        hooks_note.contains("a first install has nothing to revert"),
        "{hooks_note}"
    );
}

/// Card C-107 slice 2: the profile mutant the merged hooks SPEC recorded as surviving
/// (`let claude = true;` at the profile check) is now pinned. A project whose profile does not
/// list claude and whose settings carry no hooks key has its hooks step decided by the Go plan
/// alone: `unchanged` with the Go reason — not the claude-plan words the mutant's `true` would
/// print ("no MoAI hook set is installed … hooks-plan reads …").
#[test]
fn a_profile_without_claude_keeps_the_go_reason_on_the_hooks_step() {
    let fx = GoFx::new();
    profile_with_hosts(&fx, &["codex"]);
    write_settings(&fx, &json!({"model": "opus"}));
    let plan = after_settings(&fx);
    let step = hooks(&plan);
    assert_eq!(step["state"], "unchanged", "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert_eq!(reason, "no Go-generation hooks are left", "{step}");
    assert!(!reason.contains("hooks-plan"), "{step}");
}

/// Card C-107 slice 2 (audit follow-up): a symlinked ownership record is refused and the
/// refusal names the record file, like an unreadable one does — the settings-state refusals
/// ("symlink target refused") reach the step undecorated today.
#[test]
fn a_symlinked_record_refusal_names_the_file() {
    let fx = project_with_go_hooks();
    let record = with_the_hook_set_installed(&fx);
    fs::remove_file(&record).unwrap();
    std::os::unix::fs::symlink("/nonexistent/claude-settings.json", &record).unwrap();
    let update = fx.update_plan();
    let step = hooks(&update);
    assert_eq!(step["state"], "blocked", "{step}");
    let reason = step["reason"].as_str().unwrap();
    assert!(reason.contains("manifests/claude-settings.json"), "{step}");
    assert!(reason.contains("symlink"), "{step}");
}
