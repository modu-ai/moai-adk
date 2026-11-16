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

#[test]
fn nothing_of_the_go_generation_reads_unchanged_without_a_command() {
    let fx = GoFx::new();
    fx.with_profile();
    let plan = fx.update_plan();
    let step = hooks(&plan);
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
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
    Value::Object(moai_cli::hooks::hooks_object("claude").unwrap())
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
    assert_eq!(settings_now(fx)["hooks"], generated(), "the first apply");
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
        after["hooks"],
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
    assert_eq!(after["hooks"], generated(), "{after}");
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
