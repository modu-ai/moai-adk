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
