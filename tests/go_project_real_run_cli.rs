//! #2936 R-fix R-E (PLAN 6880040242, real run 6879901297): the whole update of a Go-generation
//! project in the shape the real run found, through the real CLI. The project has 46 hook
//! scripts of which the settings file runs 23, three output styles, three workflows and a
//! `loop.md`, all recorded as deployed, and the placeholder profile a first `moai cc` launch
//! writes. The preview names what moves, what is left behind and what is put in its place;
//! the apply moves exactly that; the next preview shows the hooks step with its command.
//!
//! Every case builds a temporary project with an empty temporary HOME and a stub `claude`; no
//! real host starts and nothing outside the temporary folders is read or written. A preview
//! leaves the whole tree byte for byte as it was.

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{action, go_state_preview, install_step, GoFx};
use serde_json::{json, Value};

fn moving(plan: &Value, kind: &str) -> usize {
    plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["kind"] == kind && i["action"] == "move")
        .count()
}

fn real_run(style: Option<&str>) -> (GoFx, Vec<String>) {
    let mut fx = GoFx::new();
    let referenced = fx.real_run_shape(style);
    fx.first_launch();
    (fx, referenced)
}

#[test]
fn the_preview_of_the_real_run_shape_counts_what_moves_and_what_stays() {
    let (fx, referenced) = real_run(Some("MoAI-Easy"));
    let plan = fx.plan();
    // 23 of the 46 scripts are run by the settings and stay; the other 23 move.
    assert_eq!(moving(&plan, "hook"), 23, "{plan}");
    for n in &referenced {
        assert_eq!(
            action(&plan, &format!(".claude/hooks/moai/{n}")),
            "leave",
            "{n}: {plan}"
        );
    }
    // The settings name `MoAI-Easy`, so `moai-easy.md` stays (O-6) and the other two move.
    assert_eq!(moving(&plan, "output_style"), 2, "{plan}");
    assert_eq!(
        action(&plan, ".claude/output-styles/moai/moai-easy.md"),
        "leave",
        "{plan}"
    );
    assert_eq!(moving(&plan, "workflow"), 3, "{plan}");
    assert_eq!(moving(&plan, "loop"), 1, "{plan}");
    assert_eq!(
        plan["kinds"],
        json!({"hook": 23, "loop": 1, "output_style": 2, "workflow": 3}),
        "{plan}"
    );
    // The settings file itself never moves.
    assert!(
        !plan["items"]
            .as_array()
            .unwrap()
            .iter()
            .any(|i| i["path"] == ".claude/settings.json" && i["action"] == "move"),
        "{plan}"
    );
    // The workflows folder ends up empty; the hooks and styles folders keep files.
    let empty: Vec<&str> = plan["after_move"]["empty_folders"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f.as_str().unwrap())
        .collect();
    assert!(empty.contains(&".claude/workflows"), "{empty:?}");
    assert!(!empty.contains(&".claude/hooks/moai"), "{empty:?}");
}

#[test]
fn without_an_output_style_the_third_style_moves_too() {
    let (fx, _) = real_run(None);
    let plan = fx.plan();
    // 23 hooks + 3 styles + 3 workflows + loop.md.
    assert_eq!(moving(&plan, "hook"), 23, "{plan}");
    assert_eq!(moving(&plan, "output_style"), 3, "{plan}");
    let total: u64 = plan["kinds"]
        .as_object()
        .unwrap()
        .values()
        .map(|v| v.as_u64().unwrap())
        .sum();
    assert_eq!(total, 30, "{plan}");
}

#[test]
fn the_update_preview_matches_and_leaves_the_tree_alone() {
    let (fx, _) = real_run(Some("MoAI-Easy"));
    let plan = fx.plan();
    let update = fx.update_plan();
    let preview = go_state_preview(&update);
    assert_eq!(preview["kinds"], plan["kinds"], "{preview}");
    assert_eq!(preview["counterparts"], plan["counterparts"], "{preview}");
    assert_eq!(preview["after_move"], plan["after_move"], "{preview}");
    // The hooks step is in the install phase; until the backup phase is applied it waits.
    let hooks = install_step(&update, "hooks").unwrap_or_else(|| panic!("no hooks step: {update}"));
    assert_eq!(hooks["state"], json!("blocked"), "{hooks}");
    assert_eq!(
        hooks["reason"],
        json!("re-run after the backup phase"),
        "{hooks}"
    );
}

#[test]
fn the_apply_moves_exactly_what_the_preview_named() {
    let (fx, referenced) = real_run(Some("MoAI-Easy"));
    let plan = fx.plan();
    let applied = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    let root = &fx.root;
    // The referenced scripts and the named style stay where they were.
    for n in &referenced {
        assert!(
            root.join(format!(".claude/hooks/moai/{n}")).is_file(),
            "{n}"
        );
    }
    assert!(root
        .join(".claude/output-styles/moai/moai-easy.md")
        .is_file());
    assert!(root.join(".claude/settings.json").is_file());
    // Everything else is gone from its place and sits in a backup.
    assert!(!root.join(".claude/hooks/moai/handle-45.sh").exists());
    assert!(!root.join(".claude/output-styles/moai/moai.md").exists());
    assert!(!root.join(".claude/workflows/plan.md").exists());
    assert!(!root.join(".claude/loop.md").exists());
    // The 29 moved files sit in the backup, under their old paths.
    let backup = applied["result"]["backup"].as_str().unwrap();
    let areas: Vec<_> = [".moai", ".moai-cli"]
        .iter()
        .map(|d| root.join(d).join("backups").join(backup))
        .filter(|a| a.is_dir())
        .collect();
    assert_eq!(areas.len(), 1, "{areas:?}");
    let moved = areas[0].join("go-state/.claude");
    let count = |dir: &str| std::fs::read_dir(moved.join(dir)).unwrap().count();
    assert_eq!(count("hooks/moai"), 23);
    assert_eq!(count("output-styles/moai"), 2);
    assert_eq!(count("workflows"), 3);
    assert!(moved.join("loop.md").is_file());
    // The next preview shows the hooks step with its command (what the real run lacked). The
    // settings step rewrites the same file, so `hooks` waits for it (#3028 audit B1): apply the
    // settings step when it has a command, then preview again.
    let mut update = fx.update_plan();
    let waiting = install_step(&update, "hooks").map(|h| h["state"].clone());
    if let Some(settings) =
        install_step(&update, "claude-settings").filter(|s| s["apply"] != Value::Null)
    {
        assert_eq!(waiting, Some(json!("blocked")), "{update}");
        let out = fx.apply_step(settings);
        assert_eq!(out["ok"], json!(true), "{out}");
        update = fx.update_plan();
    }
    let hooks = install_step(&update, "hooks").unwrap_or_else(|| panic!("no hooks step: {update}"));
    assert_eq!(
        hooks["apply"]["argv"][1],
        json!("go-hooks-apply"),
        "{hooks}"
    );
}

/// The `counterpart` text of the `hook` row of a preview.
fn hook_counterpart(preview: &Value) -> &str {
    preview["counterparts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["kind"] == "hook")
        .unwrap_or_else(|| panic!("no hook row: {}", preview["counterparts"]))["counterpart"]
        .as_str()
        .unwrap()
}

/// The `hooks` step after the backup phase is applied; the settings step rewrites the same
/// file, so it is applied first when it has a command (#3028 audit B1).
fn hooks_step_after(fx: &GoFx, plan: &Value) -> Value {
    let applied = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    let mut update = fx.update_plan();
    if let Some(settings) =
        install_step(&update, "claude-settings").filter(|s| s["apply"] != Value::Null)
    {
        let out = fx.apply_step(settings);
        assert_eq!(out["ok"], json!(true), "{out}");
        update = fx.update_plan();
    }
    install_step(&update, "hooks")
        .unwrap_or_else(|| panic!("no hooks step: {update}"))
        .clone()
}

/// #2936 F-A: the `hook` counterpart in the preview agrees with what the `hooks` step does
/// once the move is applied. Every hook script moves together with the settings file that
/// carries the deny rules: the step finds no Go hook left, so the preview must not say the
/// hooks are "reinstalled by hooks".
#[test]
fn the_hook_counterpart_agrees_with_the_hooks_step_when_everything_moves() {
    let mut fx = GoFx::new();
    for n in 0..3 {
        fx.deployed(
            &format!(".claude/hooks/moai/handle-{n}.sh"),
            &format!("#!/bin/sh\necho {n}\n"),
        );
    }
    fx.deployed(
        ".claude/settings.json",
        "{\"permissions\":{\"deny\":[\"Read(./.env)\"]}}",
    );
    fx.manifest();
    fx.first_launch();
    let plan = fx.plan();
    assert_eq!(moving(&plan, "hook"), 3, "{plan}");
    assert_eq!(
        hook_counterpart(&plan),
        "no counterpart \u{2014} will be gone",
        "{plan}"
    );
    let hooks = hooks_step_after(&fx, &plan);
    assert_eq!(hooks["state"], json!("unchanged"), "{hooks}");
    assert_eq!(hooks["apply"], Value::Null, "{hooks}");
}

/// The kept direction: scripts move, the settings stay with their Go entries, and the step
/// prints the replacement command, as the preview said.
#[test]
fn the_hook_counterpart_agrees_with_the_hooks_step_when_entries_stay() {
    let (fx, _) = real_run(Some("MoAI-Easy"));
    let plan = fx.plan();
    assert_eq!(hook_counterpart(&plan), "reinstalled by hooks", "{plan}");
    let hooks = hooks_step_after(&fx, &plan);
    assert_eq!(hooks["state"], json!("ready"), "{hooks}");
    assert_eq!(
        hooks["apply"]["argv"][1],
        json!("go-hooks-apply"),
        "{hooks}"
    );
}
