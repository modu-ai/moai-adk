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

// ---- #2936 S-CYCLE: the settings that hold what they would take with them ----

/// An unrecorded `.claude/settings.local.json` that only names the Go folder.
const LOCAL_HOLDS_MOAI: &str = r#"{"statusLine":{"type":"command","command":"bash \"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\""}}"#;

fn cycle(style: Option<&str>, local: Option<&str>) -> (GoFx, Vec<String>) {
    let mut fx = GoFx::new();
    let referenced = fx.real_run_cycle_shape(style, local);
    fx.first_launch();
    (fx, referenced)
}

fn moved_files(plan: &Value) -> u64 {
    plan["kinds"]
        .as_object()
        .unwrap()
        .values()
        .map(|v| v.as_u64().unwrap())
        .sum()
}

/// The paths whose `freed_by` names the settings file, sorted.
fn freed(plan: &Value) -> Vec<String> {
    let mut out: Vec<String> = plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i["freed_by"] == json!(".claude/settings.json"))
        .map(|i| i["path"].as_str().unwrap().to_string())
        .collect();
    out.sort();
    out
}

/// AC 5: the settings the Go deployer wrote run 23 scripts and name the style, and they move
/// themselves, so everything they held moves too: 29 moved files become 54, and the preview
/// names the 24 paths the settings were the only holder of.
#[test]
fn the_cycle_shape_frees_the_hooks_the_settings_named() {
    let (fx, referenced) = cycle(Some("MoAI-Easy"), None);
    let plan = fx.plan();
    assert_eq!(action(&plan, ".moai"), "move", "{plan}");
    assert_eq!(action(&plan, ".claude/settings.json"), "move", "{plan}");
    assert_eq!(plan["layout"], json!("folder_moves"), "{plan}");
    assert_eq!(
        plan["kinds"],
        json!({"hook": 46, "loop": 1, "output_style": 3, "settings": 1, "workflow": 3}),
        "{plan}"
    );
    assert_eq!(moved_files(&plan), 54, "{plan}");
    let mut want: Vec<String> = referenced
        .iter()
        .map(|n| format!(".claude/hooks/moai/{n}"))
        .collect();
    want.push(".claude/output-styles/moai/moai-easy.md".into());
    want.sort();
    assert_eq!(want.len(), 24);
    assert_eq!(freed(&plan), want, "{plan}");
    // No other item carries the key.
    let carrying = plan["items"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|i| i.get("freed_by").is_some())
        .count();
    assert_eq!(carrying, 24, "{plan}");
    let empty: Vec<&str> = plan["after_move"]["empty_folders"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f.as_str().unwrap())
        .collect();
    assert!(empty.contains(&".claude/hooks/moai"), "{empty:?}");
    let sm = &plan["settings_move"];
    assert_eq!(sm["path"], json!(".claude/settings.json"), "{sm}");
    assert_eq!(sm["edited"], json!(false), "{sm}");
    assert_eq!(sm["layout"], json!("folder_moves"), "{sm}");
    assert_eq!(sm["folder_freed"], json!(true), "{sm}");
    assert_eq!(sm["still_holding"], json!([]), "{sm}");
    assert_eq!(sm["freed"].as_array().unwrap().len(), 24, "{sm}");
}

/// The same shape when `settings.local.json` also names `.moai/`: the folder stays, the
/// proven settings still move with everything they held, and the preview says what holds the
/// folder.
#[test]
fn a_local_settings_file_that_names_the_folder_keeps_it_while_the_settings_move() {
    let (fx, _) = cycle(Some("MoAI-Easy"), Some(LOCAL_HOLDS_MOAI));
    let plan = fx.plan();
    assert_eq!(action(&plan, ".moai"), "leave", "{plan}");
    assert_eq!(action(&plan, ".claude/settings.json"), "move", "{plan}");
    assert_eq!(plan["layout"], json!("folder_stays"), "{plan}");
    assert_eq!(moved_files(&plan), 54, "{plan}");
    assert_eq!(freed(&plan).len(), 24, "{plan}");
    let sm = &plan["settings_move"];
    assert_eq!(sm["layout"], json!("folder_stays"), "{sm}");
    assert_eq!(sm["folder_freed"], json!(false), "{sm}");
    assert_eq!(
        sm["still_holding"],
        json!([".claude/settings.local.json"]),
        "{sm}"
    );
}

/// AC 6: a file the local settings name stays, whatever the shared settings do.
#[test]
fn files_the_local_settings_names_stay() {
    let local = json!({
        "hooks": {"Stop": [{"hooks": [{"type": "command",
            "command": "bash .claude/hooks/moai/handle-30.sh"}]}]},
        "outputStyle": "MoAI-Learn",
    })
    .to_string();
    let (fx, _) = cycle(Some("MoAI-Easy"), Some(&local));
    let plan = fx.plan();
    assert_eq!(action(&plan, ".moai"), "move", "{plan}");
    assert_eq!(action(&plan, ".claude/settings.json"), "move", "{plan}");
    for rel in [
        ".claude/hooks/moai/handle-30.sh",
        ".claude/output-styles/moai/moai-learn.md",
    ] {
        assert_eq!(action(&plan, rel), "leave", "{rel}: {plan}");
        let it = go_project_real_run::item(&plan, rel);
        assert!(
            it["reason"]
                .as_str()
                .unwrap()
                .starts_with("referenced by .claude/settings.local.json"),
            "{rel}: {plan}"
        );
        assert!(it.get("freed_by").is_none(), "{rel}: {plan}");
    }
    // 54 minus the script and the style the local settings keep.
    assert_eq!(moved_files(&plan), 52, "{plan}");
    assert_eq!(freed(&plan).len(), 24, "{plan}");
}

fn settings_json(fx: &GoFx) -> Value {
    serde_json::from_slice(&std::fs::read(fx.root.join(".claude/settings.json")).unwrap()).unwrap()
}

/// AC 12 and the B1 pin (audit #2584): the loss the old rule feared (permissions, hooks, the
/// status line) is filled in by the very update that moves the settings. After the move the
/// install phase writes MoAI's status line, the hooks step offers `hooks-apply --host claude`
/// (there are no Go hook entries left to replace), and the merge step offers the deny rules
/// back. Each step is previewed again before its command runs: the settings steps bind the
/// bytes the step before wrote.
#[test]
fn after_the_move_the_install_phase_offers_the_hooks() {
    let (fx, _) = cycle(Some("MoAI-Easy"), None);
    // 1st preview: the install phase waits for the move.
    let update = fx.update_plan();
    let hooks = install_step(&update, "hooks").unwrap();
    assert_eq!(hooks["state"], json!("blocked"), "{hooks}");
    assert_eq!(hooks["reason"], json!("re-run after the backup phase"));
    let plan = fx.plan();
    let applied = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert!(!fx.root.join(".claude/settings.json").exists());

    // 2nd preview: no settings file, so the settings step writes MoAI's status line; the hooks
    // step waits for it (both rewrite the same file).
    let update = fx.update_plan();
    let settings = install_step(&update, "claude-settings").unwrap();
    assert_eq!(settings["state"], json!("ready"), "{settings}");
    assert_eq!(
        install_step(&update, "hooks").unwrap()["reason"],
        json!("re-run after the claude-settings apply"),
        "{update}"
    );
    let out = fx.apply_step(settings);
    assert_eq!(out["ok"], json!(true), "{out}");
    assert_eq!(
        settings_json(&fx)["statusLine"],
        json!({"type": "command", "command": "moai statusline"})
    );

    // 3rd preview: nothing of the Go generation is left in the file, and the hooks step still
    // offers the MoAI hook set, because the earlier settings (with their hooks) were moved.
    let update = fx.update_plan();
    let hooks = install_step(&update, "hooks").unwrap();
    assert_eq!(hooks["state"], json!("ready"), "{hooks}");
    let argv: Vec<&str> = hooks["apply"]["argv"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap())
        .collect();
    assert_eq!(&argv[..2], ["moai", "hooks-apply"], "{argv:?}");
    assert!(
        argv.windows(2).any(|w| w == ["--host", "claude"]),
        "{argv:?}"
    );
    let out = fx.apply_step(hooks);
    assert_eq!(out["ok"], json!(true), "{out}");
    let written = settings_json(&fx);
    assert!(written["hooks"].is_object(), "{written}");

    // 4th preview: the deny rules come back on the user's go-ahead.
    let update = fx.update_plan();
    let hooks = install_step(&update, "hooks").unwrap();
    assert_eq!(hooks["state"], json!("unchanged"), "{hooks}");
    assert_eq!(hooks["apply"], Value::Null, "{hooks}");
    let merge = go_project_real_run::step_in(&update, 2, "claude-settings-merge");
    assert_eq!(merge["state"], json!("needs_decision"), "{merge}");
    let out = fx.apply_step(merge);
    assert_eq!(out["ok"], json!(true), "{out}");
    let written = settings_json(&fx);
    assert_eq!(
        written["permissions"]["deny"],
        json!(["Read(./.env)", "Read(./secrets/**)"]),
        "{written}"
    );
    assert!(written["hooks"].is_object(), "{written}");
    assert_eq!(
        written["statusLine"],
        json!({"type": "command", "command": "moai statusline"}),
        "{written}"
    );

    // Settled: the hooks step has nothing left to offer.
    let update = fx.update_plan();
    let hooks = install_step(&update, "hooks").unwrap();
    assert_eq!(hooks["state"], json!("unchanged"), "{hooks}");
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
    // S-CYCLE (OQ2 H1): the Go settings moved into the backup with their hook entries, so the
    // step has no Go hook to replace (`go-hooks-apply` is not offered) but offers the MoAI hook
    // set through the ordinary `hooks-apply --host claude`; the Go scripts themselves still
    // have no counterpart.
    assert_eq!(hooks["state"], json!("ready"), "{hooks}");
    assert_eq!(hooks["apply"]["argv"][1], json!("hooks-apply"), "{hooks}");
    assert_eq!(hooks["apply"]["argv"][4], json!("claude"), "{hooks}");
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
