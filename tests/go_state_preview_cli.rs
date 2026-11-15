//! #2936 R-fix R-A (PLAN 6880040242, R1 and R6 of the real run 6879901297): before the Go
//! files move, the preview says what this generation puts in their place (`counterparts`)
//! and what the move leaves behind (`after_move`: folders that end up empty, include or
//! ignore lines that name a moved path, the settings wiring that goes with the settings
//! file). Output styles are compared by their frontmatter name, without case (O-6).
//!
//! Every case builds a temporary project with an empty temporary HOME. A preview leaves the
//! whole tree byte for byte as it was (AC-8).

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{action, go_state_preview, item, GoFx};
use serde_json::{json, Value};

/// `worktree_include::LIMIT`: a larger include or ignore list is not read.
const LIST_LIMIT: usize = 1024 * 1024;

/// A Go project with one file of each kind the real run moved, `.claude/settings.json` not
/// yet written.
fn go_fx() -> GoFx {
    let mut fx = GoFx::new();
    for n in ["handle-a", "handle-b", "handle-c"] {
        fx.deployed(
            &format!(".claude/hooks/moai/{n}.sh"),
            &format!("#!/bin/sh\necho {n}\n"),
        );
    }
    fx.deployed(
        ".claude/output-styles/moai/moai.md",
        "---\nname: MoAI\n---\nbody\n",
    );
    fx.deployed(
        ".claude/output-styles/moai/moai-easy.md",
        "---\nname: MoAI-Easy\n---\nbody\n",
    );
    fx.deployed(".claude/workflows/w1.md", "# w1\n");
    fx.deployed(".claude/commands/moai/x.md", "# x\n");
    fx.deployed(".claude/rules/r.md", "# r\n");
    fx.deployed(".claude/loop.md", "# loop\n");
    fx
}

/// A settings file whose hooks run the named scripts, with an optional output style and
/// deny rules.
fn settings_body(scripts: &[&str], style: Option<&str>, deny: &[&str]) -> String {
    let handlers: Vec<Value> = scripts
        .iter()
        .map(|s| json!({"type": "command", "command": format!("bash .claude/hooks/moai/{s}")}))
        .collect();
    let mut v = json!({
        "hooks": {"PreToolUse": [{"matcher": "*", "hooks": handlers}]},
        "permissions": {"deny": deny},
    });
    if let Some(s) = style {
        v["outputStyle"] = json!(s);
    }
    v.to_string()
}

fn counterpart<'a>(plan: &'a Value, kind: &str) -> &'a Value {
    plan["counterparts"]
        .as_array()
        .unwrap_or_else(|| panic!("no counterparts: {plan}"))
        .iter()
        .find(|r| r["kind"] == kind)
        .unwrap_or_else(|| panic!("no counterpart row for {kind}: {}", plan["counterparts"]))
}

fn empty_folders(plan: &Value) -> Vec<String> {
    plan["after_move"]["empty_folders"]
        .as_array()
        .unwrap_or_else(|| panic!("no after_move: {plan}"))
        .iter()
        .map(|f| f.as_str().unwrap().to_string())
        .collect()
}

// ---- AC-1, AC-2: the counterpart of each moved kind ----

#[test]
fn every_moved_kind_gets_a_counterpart_row() {
    let fx = go_fx();
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], None, &[]),
    );
    let plan = fx.plan();
    assert_eq!(
        plan["kinds"],
        json!({"command": 1, "hook": 2, "loop": 1, "output_style": 2, "rule": 1, "workflow": 1}),
        "{plan}"
    );
    assert_eq!(
        item(&plan, ".claude/hooks/moai/handle-b.sh")["kind"],
        "hook"
    );
    assert_eq!(item(&plan, ".claude/loop.md")["kind"], "loop");
    for (kind, moved, text) in [
        ("hook", 2, "reinstalled by hooks"),
        ("output_style", 2, "user layer after moai update"),
        ("workflow", 1, "no counterpart \u{2014} will be gone"),
        ("command", 1, "no counterpart \u{2014} will be gone"),
        ("rule", 1, "no counterpart \u{2014} will be gone"),
        ("loop", 1, "no counterpart \u{2014} will be gone"),
    ] {
        let row = counterpart(&plan, kind);
        assert_eq!(row["moved"], json!(moved), "{kind}: {row}");
        assert_eq!(row["counterpart"], json!(text), "{kind}: {row}");
    }
    assert_eq!(
        plan["counterparts"].as_array().unwrap().len(),
        6,
        "one row per moved kind and no other: {}",
        plan["counterparts"]
    );
}

#[test]
fn an_output_style_name_resolves_only_from_home() {
    let fx = go_fx();
    fx.manifest();
    let plan = fx.plan();
    let note = counterpart(&plan, "output_style")["note"].as_str().unwrap();
    assert!(
        note.contains("resolves from HOME's output-styles only"),
        "{note}"
    );
}

/// The miss direction: a kind with nothing moving gets no row, and a plan that moves nothing
/// has no rows at all.
#[test]
fn a_kind_that_does_not_move_has_no_row() {
    let mut fx = GoFx::new();
    fx.deployed(".claude/workflows/w1.md", "# w1\n");
    fx.manifest();
    let plan = fx.plan();
    let rows = plan["counterparts"].as_array().unwrap();
    assert_eq!(rows.len(), 1, "{plan}");
    assert_eq!(rows[0]["kind"], "workflow");

    let empty = GoFx::new();
    empty.write(".claude/keep.md", "x\n");
    let plan = empty.plan();
    assert_eq!(plan["counterparts"], json!([]), "{plan}");
    assert_eq!(plan["kinds"], json!({}), "{plan}");
}

#[test]
fn the_update_preview_carries_the_counterparts_and_the_after_move() {
    let fx = go_fx();
    fx.manifest();
    fx.write(".gitignore", ".claude/loop.md\n");
    let plan = fx.plan();
    let update = fx.update_plan();
    let preview = go_state_preview(&update);
    assert_eq!(preview["counterparts"], plan["counterparts"], "{preview}");
    assert_eq!(preview["after_move"], plan["after_move"], "{preview}");
    assert_eq!(preview["kinds"], plan["kinds"], "{preview}");
    assert_eq!(
        preview["after_move"]["dangling"][0]["file"], ".gitignore",
        "{preview}"
    );
}

// ---- AC-4: only the folders that end up empty ----

#[test]
fn only_the_folders_that_end_up_empty_are_listed() {
    let fx = go_fx();
    fx.manifest();
    // handle-a.sh is named by the settings, so it stays and keeps its folders.
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], None, &[]),
    );
    let plan = fx.plan();
    assert_eq!(
        empty_folders(&plan),
        [
            ".claude/commands",
            ".claude/commands/moai",
            ".claude/output-styles",
            ".claude/output-styles/moai",
            ".claude/rules",
            ".claude/workflows",
        ],
        "{plan}"
    );
    assert_eq!(plan["after_move"]["unknown"], json!(false), "{plan}");
}

#[test]
fn when_every_script_moves_the_hook_folders_and_the_claude_folder_are_listed() {
    let fx = go_fx();
    fx.manifest();
    let plan = fx.plan();
    let folders = empty_folders(&plan);
    for f in [".claude/hooks/moai", ".claude/hooks", ".claude"] {
        assert!(folders.contains(&f.to_string()), "{f} in {folders:?}");
    }
}

/// The referenced-script exception at its edge: one script left keeps `hooks/moai`, none left
/// empties it; 22 and 23 referenced of 46 both keep it.
#[test]
fn a_referenced_script_keeps_its_folder() {
    for (referenced, hooks_empty) in [(0usize, true), (1, false), (22, false), (23, false)] {
        let mut fx = GoFx::new();
        let names: Vec<String> = (0..46).map(|i| format!("hook-{i:02}.sh")).collect();
        for n in &names {
            fx.deployed(&format!(".claude/hooks/moai/{n}"), &format!("# {n}\n"));
        }
        fx.manifest();
        let refs: Vec<&str> = names.iter().take(referenced).map(String::as_str).collect();
        fx.write(".claude/settings.json", &settings_body(&refs, None, &[]));
        let plan = fx.plan();
        assert_eq!(
            plan["kinds"]["hook"],
            json!(46 - referenced),
            "{referenced} referenced: {plan}"
        );
        assert_eq!(
            empty_folders(&plan).contains(&".claude/hooks/moai".to_string()),
            hooks_empty,
            "{referenced} referenced: {:?}",
            empty_folders(&plan)
        );
    }
}

/// The walk's limit at its edge (N passes, N+1 is unknown): the entries read here are the 3
/// of `.claude/workflows` and the 2 of `.claude`.
#[test]
fn the_empty_folder_walk_stops_at_its_limit() {
    let mut fx = GoFx::new();
    for n in 1..=3 {
        fx.deployed(&format!(".claude/workflows/w{n}.md"), "# w\n");
    }
    fx.manifest();
    fx.write(".claude/settings.json", "{}\n");
    let at = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", "5")]);
    assert_eq!(at["after_move"]["unknown"], json!(false), "N: {at}");
    assert_eq!(empty_folders(&at), [".claude/workflows"], "N: {at}");

    let over = fx.plan_env(&[("MOAI_CLI_TEST_UNRECORDED_LIMIT", "4")]);
    assert_eq!(over["after_move"]["unknown"], json!(true), "N+1: {over}");
    // What was settled before the walk stopped is still true; nothing else is listed.
    assert!(
        empty_folders(&over)
            .iter()
            .all(|f| f == ".claude/workflows"),
        "N+1: {over}"
    );
}

// ---- AC-5: include and ignore lines that name a moved path ----

#[test]
fn a_line_that_names_a_moved_path_is_dangling() {
    let fx = go_fx();
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], None, &[]),
    );
    fx.write(
        ".worktreeinclude",
        ".claude/workflows/\n.claude/hooks/moai/handle-a.sh\n",
    );
    fx.write(".gitignore", "node_modules/\n.claude/loop.md\n");
    let plan = fx.plan();
    let dangling = plan["after_move"]["dangling"].as_array().unwrap();
    assert_eq!(dangling.len(), 2, "{plan}");
    let by = |file: &str| {
        dangling
            .iter()
            .find(|d| d["file"] == file)
            .unwrap_or_else(|| panic!("no {file} row: {dangling:?}"))
    };
    let (include, ignore) = (by(".worktreeinclude"), by(".gitignore"));
    assert_eq!(include["line"], json!(1), "{include}");
    assert_eq!(include["pattern"], json!(".claude/workflows/"), "{include}");
    assert_eq!(
        include["paths"],
        json!([".claude/workflows/w1.md"]),
        "{include}"
    );
    assert_eq!(ignore["line"], json!(2), "{ignore}");
    assert_eq!(ignore["pattern"], json!(".claude/loop.md"), "{ignore}");
    assert_eq!(plan["after_move"]["unknown"], json!(false), "{plan}");
}

/// The miss direction: lines that select only what stays, or nothing, are not dangling.
#[test]
fn a_line_that_names_what_stays_is_not_dangling() {
    let fx = go_fx();
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], None, &[]),
    );
    fx.write(
        ".worktreeinclude",
        ".claude/hooks/moai/handle-a.sh\n.claude/settings.local.json\n",
    );
    fx.write(".gitignore", "node_modules/\n*.log\n");
    let plan = fx.plan();
    assert_eq!(plan["after_move"]["dangling"], json!([]), "{plan}");
}

/// A list of exactly the size limit is read; one byte more is not, and the preview says it
/// does not know.
#[test]
fn the_include_list_is_read_up_to_its_size_limit() {
    let fx = go_fx();
    fx.manifest();
    let list = |len: usize| {
        let head = ".claude/loop.md\n";
        format!("{head}#{}", "x".repeat(len - head.len() - 1))
    };
    fx.write(".worktreeinclude", &list(LIST_LIMIT));
    let at = fx.plan();
    assert_eq!(
        at["after_move"]["dangling"].as_array().unwrap().len(),
        1,
        "N: {at}"
    );
    assert_eq!(at["after_move"]["unknown"], json!(false), "N: {at}");

    fx.write(".worktreeinclude", &list(LIST_LIMIT + 1));
    let over = fx.plan();
    assert_eq!(over["after_move"]["dangling"], json!([]), "N+1: {over}");
    assert_eq!(over["after_move"]["unknown"], json!(true), "N+1: {over}");
}

#[test]
fn the_ignore_list_is_read_up_to_its_size_limit() {
    let fx = go_fx();
    fx.manifest();
    let list = |len: usize| {
        let head = ".claude/loop.md\n";
        format!("{head}#{}", "x".repeat(len - head.len() - 1))
    };
    fx.write(".gitignore", &list(LIST_LIMIT));
    let at = fx.plan();
    assert_eq!(
        at["after_move"]["dangling"][0]["file"], ".gitignore",
        "N: {at}"
    );
    fx.write(".gitignore", &list(LIST_LIMIT + 1));
    let over = fx.plan();
    assert_eq!(over["after_move"]["dangling"], json!([]), "N+1: {over}");
    assert_eq!(over["after_move"]["unknown"], json!(true), "N+1: {over}");
}

// ---- AC-6: the settings wiring that goes with a moved settings file ----

#[test]
fn a_moved_settings_file_reports_the_wiring_that_goes_with_it() {
    let mut fx = go_fx();
    fx.deployed(
        ".claude/settings.json",
        &settings_body(
            &["handle-a.sh", "handle-b.sh"],
            None,
            &["Bash(rm -rf:*)", "Read(./.env)", "Read(./secrets/**)"],
        ),
    );
    fx.manifest();
    let plan = fx.plan();
    assert_eq!(action(&plan, ".claude/settings.json"), "move", "{plan}");
    let wiring = &plan["after_move"]["wiring_lost"];
    assert_eq!(wiring["file"], json!(".claude/settings.json"), "{wiring}");
    assert_eq!(wiring["go_hook_entries"], json!(2), "{wiring}");
    assert_eq!(wiring["deny_rules"], json!(3), "{wiring}");
    assert_eq!(wiring["keys"], json!(["hooks", "permissions"]), "{wiring}");
}

/// The miss direction: settings that stay lose nothing.
#[test]
fn settings_that_stay_report_no_lost_wiring() {
    let fx = go_fx();
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        &settings_body(&["handle-a.sh"], None, &["Read(./.env)"]),
    );
    let plan = fx.plan();
    assert!(plan["after_move"].is_object(), "{plan}");
    assert_eq!(plan["after_move"]["wiring_lost"], Value::Null, "{plan}");
}

// ---- AC-7, O-6: output styles compared by name ----

fn style_fx(style: &str, files: &[(&str, &str)]) -> GoFx {
    let mut fx = GoFx::new();
    for (rel, body) in files {
        fx.deployed(rel, body);
    }
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        &settings_body(&[], Some(style), &[]),
    );
    fx
}

#[test]
fn a_style_is_kept_when_its_frontmatter_name_matches_without_case() {
    let fx = style_fx(
        "moai-easy",
        &[
            (
                ".claude/output-styles/moai/custom.md",
                "---\nname: MoAI-Easy\n---\nb\n",
            ),
            (
                ".claude/output-styles/moai/other.md",
                "---\nname: Other\n---\nb\n",
            ),
        ],
    );
    let plan = fx.plan();
    let kept = item(&plan, ".claude/output-styles/moai/custom.md");
    assert_eq!(kept["action"], "leave", "{kept}");
    assert!(
        kept["reason"].as_str().unwrap().contains("outputStyle"),
        "{kept}"
    );
    assert_eq!(action(&plan, ".claude/output-styles/moai/other.md"), "move");
}

#[test]
fn a_style_is_kept_when_its_file_name_matches_without_case() {
    let fx = style_fx(
        "MoAI-Easy",
        &[
            (
                ".claude/output-styles/moai/moai-easy.md",
                "no frontmatter\n",
            ),
            (".claude/output-styles/moai/moai.md", "no frontmatter\n"),
        ],
    );
    let plan = fx.plan();
    assert_eq!(
        action(&plan, ".claude/output-styles/moai/moai-easy.md"),
        "leave"
    );
    assert_eq!(action(&plan, ".claude/output-styles/moai/moai.md"), "move");
}

/// The exact spelling kept working before; it still does.
#[test]
fn a_style_named_exactly_is_kept() {
    let fx = style_fx(
        "moai-easy",
        &[(
            ".claude/output-styles/moai/moai-easy.md",
            "no frontmatter\n",
        )],
    );
    let plan = fx.plan();
    assert_eq!(
        action(&plan, ".claude/output-styles/moai/moai-easy.md"),
        "leave"
    );
}

// ---- AC-8: the new fields are preview-only, not part of the digest ----

/// S-DIGEST (#2936 plan 6887264857 §3): preview-only data is not bound. A dangling ignore
/// line lives in `after_move`, which the apply never acts on, so a line added between the
/// preview and the apply shows in the fresh preview but keeps the confirmed digest valid.
/// (Before S-DIGEST the whole preview was hashed and this was refused.)
#[test]
fn a_new_dangling_line_shows_in_the_preview_but_keeps_the_plan_digest() {
    let fx = go_fx();
    fx.manifest();
    let before = fx.plan();
    fx.write(".gitignore", ".claude/loop.md\n");
    let after = fx.plan();
    assert_ne!(
        before["after_move"]["dangling"], after["after_move"]["dangling"],
        "the new line shows in the preview: {after}"
    );
    assert_eq!(before["plan_digest"], after["plan_digest"], "{after}");
    let applied = fx.apply(before["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    assert_eq!(applied["result"]["state"], json!("committed"), "{applied}");
}

// ---- AC-11, O-2: an emptied skill mirror folder is warned about and kept ----

#[test]
fn an_emptied_agents_skills_folder_is_warned_about_and_kept_by_the_apply() {
    let mut fx = GoFx::new();
    fx.deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    fx.manifest();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    let plan = fx.plan();
    assert_eq!(action(&plan, ".agents/skills/moai-a"), "move", "{plan}");
    let folders = empty_folders(&plan);
    assert!(
        folders.contains(&".agents/skills".to_string()),
        "{folders:?}"
    );
    let warned = plan["warnings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| w.as_str().unwrap())
        .find(|w| w.contains(".agents/skills"))
        .unwrap_or_else(|| panic!("no warning about .agents/skills: {}", plan["warnings"]));
    assert!(warned.contains("the apply deletes no folder"), "{warned}");

    let applied = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    let left = fx.root.join(".agents/skills");
    assert!(left.is_dir(), "the apply deleted a folder");
    assert_eq!(std::fs::read_dir(&left).unwrap().count(), 0);
}

/// The miss direction: a mirror folder that keeps something is not warned about.
#[test]
fn a_skills_folder_that_keeps_something_is_not_warned_about() {
    let mut fx = GoFx::new();
    fx.deployed(".claude/skills/moai-a/SKILL.md", "# a\n");
    fx.manifest();
    fx.link(".agents/skills/moai-a", "../../.claude/skills/moai-a");
    fx.write(".agents/skills/mine/SKILL.md", "mine\n");
    let plan = fx.plan();
    assert!(!empty_folders(&plan).contains(&".agents/skills".to_string()));
    assert!(
        !plan["warnings"]
            .to_string()
            .contains("the apply deletes no folder"),
        "{}",
        plan["warnings"]
    );
}

// ---- #2936 S0 (F6, F16, F24): the preview says what the move does with the Go folder ----

/// Where the new `.moai/` keeps the dated backup; the apply fills in `<UTC>`.
const BACKUP_AREA: &str = ".moai/backups/<UTC>";

/// A Go project that only has files under `.moai/`: no `.claude/` file can move on its own, so
/// the `.moai/` item alone decides what the plan does.
fn folder_only_fx() -> GoFx {
    let mut fx = GoFx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.manifest();
    fx
}

/// The preview `project-update-plan` embeds carries the same three fields as the plan.
fn assert_embedded(fx: &GoFx, plan: &Value) {
    let update = fx.update_plan();
    let preview = go_state_preview(&update);
    for key in ["layout", "go_backup", "backup_to"] {
        assert_eq!(preview[key], plan[key], "{key}: {preview}");
    }
}

#[test]
fn layout_folder_moves_names_the_go_backup_and_the_backup_area() {
    let fx = go_fx();
    fx.manifest();
    let plan = fx.plan();
    assert_eq!(plan["layout"], json!("folder_moves"), "{plan}");
    assert_eq!(plan["go_backup"], json!(".moai-backup"), "{plan}");
    assert_eq!(plan["backup_to"], json!(BACKUP_AREA), "{plan}");
    assert_embedded(&fx, &plan);
    let update = fx.update_plan();
    assert_eq!(update["phases"][0]["steps"][0]["state"], "needs_decision");
}

#[test]
fn layout_folder_stays_when_the_settings_name_the_folder() {
    let fx = folder_only_fx();
    fx.write(
        ".claude/settings.json",
        "{\"statusLine\": \".moai/status.sh\"}\n",
    );
    let plan = fx.plan();
    assert_eq!(plan["layout"], json!("folder_stays"), "{plan}");
    assert_eq!(action(&plan, ".moai"), "leave", "{plan}");
    // Nothing moves, so there is no Go backup name and no backup area, not an empty string.
    assert_eq!(plan["go_backup"], Value::Null, "{plan}");
    assert_eq!(plan["backup_to"], Value::Null, "{plan}");
    assert_embedded(&fx, &plan);
    // F24: no step waits on a confirmation, no command is printed.
    let update = fx.update_plan();
    let step = &update["phases"][0]["steps"][0];
    assert_ne!(step["state"], "needs_decision", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
}

/// The nested-only case: the one thing in the way is a checkout inside `.moai/`.
#[test]
fn layout_folder_stays_for_a_nested_checkout_and_the_reason_says_so() {
    let fx = folder_only_fx();
    fx.write(".moai/nested/.git/HEAD", "ref: refs/heads/main\n");
    let plan = fx.plan();
    assert_eq!(plan["layout"], json!("folder_stays"), "{plan}");
    assert_eq!(plan["go_backup"], Value::Null, "{plan}");
    assert_eq!(plan["backup_to"], Value::Null, "{plan}");
    let reason = item(&plan, ".moai")["reason"].as_str().unwrap();
    assert!(reason.starts_with("1 nested checkout(s)"), "{reason}");
    assert!(
        reason.contains("the .moai/ folder stays where it is and is not moved into a backup"),
        "{reason}"
    );
    assert_embedded(&fx, &plan);
}

/// The edge of F24: a staying folder with one proven `.claude/` file to move still makes a
/// backup, and the plan names where; with no such file it names none.
#[test]
fn a_staying_folder_names_a_backup_only_when_a_file_moves() {
    let mut fx = GoFx::new();
    fx.deployed(".moai/config/sections/a.yaml", "a: 1\n");
    fx.deployed(".claude/workflows/w1.md", "# w1\n");
    fx.manifest();
    fx.write(
        ".claude/settings.json",
        "{\"statusLine\": \".moai/status.sh\"}\n",
    );
    let one = fx.plan();
    assert_eq!(one["layout"], json!("folder_stays"), "N=1: {one}");
    assert_eq!(
        action(&one, ".claude/workflows/w1.md"),
        "move",
        "N=1: {one}"
    );
    assert_eq!(one["go_backup"], Value::Null, "N=1: {one}");
    // The state folder in use is not `.moai/`: that folder is the one that stays.
    assert_eq!(
        one["backup_to"],
        json!(format!("{}/backups/<UTC>", moai_cli::PROJECT_STATE_DIR)),
        "N=1: {one}"
    );
    assert_embedded(&fx, &one);

    std::fs::remove_file(fx.root.join(".claude/workflows/w1.md")).unwrap();
    let zero = fx.plan();
    assert_eq!(zero["layout"], json!("folder_stays"), "N=0: {zero}");
    assert_eq!(zero["backup_to"], Value::Null, "N=0: {zero}");
}

#[test]
fn layout_legacy_into_new_moves_the_earlier_root_into_a_new_moai() {
    let fx = GoFx::new();
    fx.write(
        &format!("{}/specs/SPEC-1/spec.md", moai_cli::PROJECT_STATE_DIR),
        "# spec\n",
    );
    let plan = fx.plan();
    assert_eq!(plan["layout"], json!("legacy_into_new"), "{plan}");
    // No Go folder is renamed, so there is no Go backup name; the receipt area is in the new
    // `.moai/`.
    assert_eq!(plan["go_backup"], Value::Null, "{plan}");
    assert_eq!(plan["backup_to"], json!(BACKUP_AREA), "{plan}");
    assert_embedded(&fx, &plan);
}

#[test]
fn layout_nothing_has_no_go_backup_and_no_backup_area() {
    let fx = GoFx::new();
    fx.write(".claude/keep.md", "x\n");
    let plan = fx.plan();
    assert_eq!(plan["layout"], json!("nothing"), "{plan}");
    assert_eq!(plan["go_backup"], Value::Null, "{plan}");
    assert_eq!(plan["backup_to"], Value::Null, "{plan}");
    assert_embedded(&fx, &plan);
    let update = fx.update_plan();
    let step = &update["phases"][0]["steps"][0];
    assert_eq!(step["state"], "unchanged", "{step}");
    assert_eq!(step["apply"], Value::Null, "{step}");
}

/// After the apply the Go folder is gone from the project (only its backup is left), so the
/// same project reads `nothing`.
#[test]
fn after_the_apply_the_same_project_reads_nothing() {
    let fx = go_fx();
    fx.manifest();
    let plan = fx.plan();
    assert_eq!(plan["layout"], json!("folder_moves"), "{plan}");
    let applied = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(applied["ok"], json!(true), "{applied}");
    let again = fx.plan();
    assert_eq!(again["layout"], json!("nothing"), "{again}");
    assert_eq!(again["go_backup"], Value::Null, "{again}");
    assert_eq!(again["backup_to"], Value::Null, "{again}");
}
