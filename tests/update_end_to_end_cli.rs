//! #2936 L2b (completion conditions 1 to 3 of comment 6904409512): `/moai update` on a
//! Go-generation project reaches the end. Each test starts from one shape of the reproduction
//! (comment 6904353302) and runs the loop the UPDATE flow runs: read `project-update-plan`,
//! execute the first step that prints a command exactly as printed (argv and stdin, with a
//! permission reference), read the plan again, and stop at `all_unchanged = true` or when no step
//! prints a command.
//!
//! - G: the Go manifest does not record `.claude/settings.json`.
//! - H: it does, and a Rust settings receipt owns a key.
//! - E: half moved: a go-state receipt committed, the Go rule files deleted from the work tree
//!   (not committed to git), an earlier `.moai-cli/` and a `.moai/.update.lock` of a dead pid.
//!
//! Every case uses a temporary project, an empty temporary HOME and one fake Rust launcher at
//! `HOME/.moai-cli/path/moai-cli`; nothing outside the temporary folders is read or written.

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{step_in, GoFx, Node, REF};
use moai_cli::status_line::RUST_HELP_LINE;
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet};
use std::fs;

const SETTINGS: &str = ".claude/settings.json";
const LOCAL: &str = ".claude/settings.local.json";
const LAUNCHER: &str = ".moai-cli/path/moai-cli";
const LOCAL_BODY: &str = "{\"env\":{\"LOCAL\":\"1\"}}\n";
/// More steps than any shape needs; a loop that reaches it has a cycle.
const MAX_STEPS: usize = 14;

/// A launcher that answers `statusline --help` the way the Rust build does.
fn rust_launcher() -> String {
    format!(
        "#!/bin/sh\nif [ \"$1\" = statusline ] && [ \"$2\" = --help ]; then\n  echo '{RUST_HELP_LINE}'\n  echo 'Usage: moai statusline'\nfi\n"
    )
}

/// The Go tool's settings: three Go-made keys, Go hook entries and keys that are not its business.
fn go_settings() -> Value {
    let hook = |n: u32| json!({"type": "command", "command": format!("bash .claude/hooks/moai/handle-{n:02}.sh")});
    json!({
        "permissions": {"deny": ["Read(./.env)"], "allow": ["Bash(ls:*)"]},
        "hooks": {"PreToolUse": [{"matcher": "*", "hooks": [hook(0), hook(1)]}]},
        "statusLine": {"type": "command", "command": "$CLAUDE_PROJECT_DIR/.moai/status_line.sh"},
        "model": "opus",
        "subagentStatusLine": {"type": "command",
                               "command": "\"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\"",
                               "refreshInterval": 60},
        "plansDirectory": ".moai/plans",
        "env": {"FOO": "1"},
    })
}

/// A Go-generation project without its settings file in the manifest yet: the status script, the
/// plans, two hook scripts, a rule file, a root `CLAUDE.md`, the user's own sources and local
/// settings, and the fake launcher.
fn base_fx() -> GoFx {
    let mut fx = GoFx::new();
    fx.deployed(".moai/status_line.sh", "#!/bin/sh\necho status\n");
    for n in 0..2 {
        fx.deployed(
            &format!(".claude/hooks/moai/handle-{n:02}.sh"),
            &format!("#!/bin/sh\necho {n}\n"),
        );
    }
    fx.deployed(".claude/rules/moai/r1.md", "rule one\n");
    fx.deployed("CLAUDE.md", "# Go CLAUDE.md\n");
    fx.write(".moai/plans/a.md", "# a\n");
    fx.write(".moai/plans/sub/b.md", "# b\n");
    fx.write("src/main.rs", "fn main() {}\n");
    fx.write(LOCAL, LOCAL_BODY);
    let path = fx.home.join(LAUNCHER);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    common::write_executable(&path, &rust_launcher());
    fx
}

/// Shape G: the Go manifest does not record the settings file.
fn shape_g() -> GoFx {
    let fx = base_fx();
    fx.manifest();
    fx.write(SETTINGS, &go_settings().to_string());
    fx.with_profile();
    fx
}

/// Shape H: the manifest records the settings file and a Rust settings receipt owns a key.
fn shape_h() -> GoFx {
    let mut fx = base_fx();
    fx.deployed(SETTINGS, &go_settings().to_string());
    fx.manifest();
    fx.write(
        ".moai-cli/manifests/claude-settings.json",
        &json!({"schema_version": 1, "owned": {"permissions": go_settings()["permissions"]}})
            .to_string(),
    );
    fx.with_profile();
    fx
}

fn git(fx: &GoFx, args: &[&str]) {
    let mut all = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    all.extend_from_slice(args);
    common::git::git(&fx.root, &all);
}

/// The pid of a process that has finished: no live process has it now.
fn dead_pid() -> u32 {
    let mut child = std::process::Command::new("true").spawn().unwrap();
    let pid = child.id();
    child.wait().unwrap();
    pid
}

/// Shape E: an earlier go-state apply (committed receipt) moved the Go rule files out of the
/// work tree while the Go `.moai/` stayed (the settings name it); git has the deletions
/// uncommitted. The earlier `.moai-cli/` holds the project file and an empty manifests folder,
/// and the Go update lock names a process that is gone.
fn shape_e() -> (GoFx, u32) {
    let fx = base_fx();
    fx.manifest();
    fx.write(SETTINGS, &go_settings().to_string());
    git(&fx, &["init", "-q", "-b", "main"]);
    git(&fx, &["add", "."]);
    git(&fx, &["commit", "-q", "-m", "init"]);
    let plan = fx.plan();
    let out = fx.apply(plan["plan_digest"].as_str().unwrap());
    assert_eq!(out["ok"], json!(true), "{out}");
    assert!(
        !fx.root.join(".claude/rules/moai/r1.md").exists(),
        "the earlier apply moved the rule file: {out}"
    );
    assert!(
        fx.root.join(".moai/manifest.json").is_file(),
        "the Go folder stayed: {out}"
    );
    fx.with_profile();
    fs::create_dir_all(fx.root.join(".moai-cli/manifests")).unwrap();
    let pid = dead_pid();
    fx.write(
        ".moai/.update.lock",
        &json!({"pid": pid, "started_at": "2026-11-01T00:00:00Z", "hostname": "old-host"})
            .to_string(),
    );
    (fx, pid)
}

/// What the loop saw.
struct Walk {
    /// The step each round executed, in order.
    order: Vec<String>,
    /// What each executed step printed, in the same order.
    applied: Vec<(String, Value)>,
    /// The plan that ended the loop.
    last: Value,
}

fn steps_of(plan: &Value) -> Vec<&Value> {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().unwrap())
        .collect()
}

/// The steps that are neither `unchanged` nor `skip`: `(step, state)`.
fn open_steps(plan: &Value) -> Vec<(String, String)> {
    steps_of(plan)
        .into_iter()
        .filter(|s| !matches!(s["state"].as_str().unwrap(), "unchanged" | "skip"))
        .map(|s| {
            (
                s["step"].as_str().unwrap().to_string(),
                s["state"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

/// The loop of the UPDATE flow. Every command is the printed one, and must succeed.
fn walk(fx: &GoFx) -> Walk {
    let mut order = Vec::new();
    let mut applied = Vec::new();
    for _ in 0..MAX_STEPS {
        let plan = fx.update_plan();
        if plan["all_unchanged"] == json!(true) {
            return Walk {
                order,
                applied,
                last: plan,
            };
        }
        let Some(step) = steps_of(&plan)
            .into_iter()
            .find(|s| s["apply"].is_object())
            .cloned()
        else {
            return Walk {
                order,
                applied,
                last: plan,
            };
        };
        let name = step["step"].as_str().unwrap().to_string();
        let out = fx.apply_step(&step);
        assert_eq!(out["ok"], json!(true), "{name}: {out}");
        applied.push((name.clone(), out["result"].clone()));
        order.push(name);
    }
    panic!("the loop did not end in {MAX_STEPS} steps: {order:?}");
}

/// Every file body in the project and HOME, git's own folder left out.
fn bodies(fx: &GoFx) -> BTreeSet<Vec<u8>> {
    fx.tree()
        .into_iter()
        .filter(|(p, _)| !p.components().any(|c| c.as_os_str() == ".git"))
        .filter_map(|(_, n)| match n {
            Node::File(b) => Some(b),
            _ => None,
        })
        .collect()
}

/// The project files (git's folder and the fixture's launcher left out) with their bytes.
fn project_files(fx: &GoFx) -> Vec<(String, Vec<u8>)> {
    fx.tree()
        .into_iter()
        .filter(|(p, _)| p.starts_with(&fx.root))
        .filter(|(p, _)| !p.components().any(|c| c.as_os_str() == ".git"))
        .filter_map(|(p, n)| match n {
            Node::File(b) => Some((p.strip_prefix(&fx.root).unwrap().display().to_string(), b)),
            _ => None,
        })
        .collect()
}

/// After the walk: nothing exists outside the two temporary folders, and every file the project
/// started with is still somewhere, byte for byte (moved into a backup or kept by a receipt);
/// no step deleted a user's file. Returns the paths whose bytes are not kept elsewhere.
fn assert_contained_and_kept(fx: &GoFx, start: &[(String, Vec<u8>)]) {
    let mut siblings: Vec<String> = fs::read_dir(fx.root.parent().unwrap())
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    siblings.sort();
    assert_eq!(siblings, ["home", "project"]);
    let all = bodies(fx);
    let lost: Vec<&str> = start
        .iter()
        .filter(|(_, b)| !all.contains(b))
        .map(|(p, _)| p.as_str())
        .collect();
    assert!(lost.is_empty(), "files whose bytes are gone: {lost:?}");
}

/// The `path`s of a plan's leftovers.
fn leftover_paths(plan: &Value) -> Vec<&str> {
    plan["leftovers"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l["path"].as_str().unwrap())
        .collect()
}

fn leftover<'a>(plan: &'a Value, path: &str) -> &'a Value {
    plan["leftovers"]
        .as_array()
        .unwrap()
        .iter()
        .find(|l| l["path"] == path)
        .unwrap_or_else(|| panic!("no leftover {path}: {plan}"))
}

/// The common end of G, H and E: the loop ran the steps in `expected` order and stopped at
/// `all_unchanged = true`, the Rust marker is in place, nothing was lost, and a second loop has
/// nothing to do.
fn assert_walks_to_the_end(fx: &GoFx, expected: &[&str]) -> Walk {
    let start = project_files(fx);
    let first = fx.update_plan();
    assert_eq!(first["all_unchanged"], json!(false), "{first}");
    let w = walk(fx);
    assert_eq!(w.order, expected, "{:?}", open_steps(&w.last));
    assert_eq!(w.last["all_unchanged"], json!(true), "{}", w.last);
    assert!(open_steps(&w.last).is_empty(), "{}", w.last);
    // The Rust marker is the last thing a migration writes: this generation owns `.moai/` now.
    let state = &w.last["state_root"];
    assert_eq!(state["generation"], json!("rust"), "{state}");
    assert_eq!(state["rust_marker"], json!("valid"), "{state}");
    assert!(fx.root.join(".moai/moai.json").is_file());
    assert_contained_and_kept(fx, &start);
    // Done is done: a second loop executes nothing and still reads `all_unchanged`.
    let again = walk(fx);
    assert!(again.order.is_empty(), "{:?}", again.order);
    assert_eq!(again.last["all_unchanged"], json!(true), "{}", again.last);
    w
}

/// The order every shape runs: the settings clean-up, the move of the Go folder, then the
/// install steps that have something to write (the MoAI files, the hook set, the host files).
const ORDER: [&str; 6] = [
    "go-settings",
    "go-state",
    "generation",
    "hooks",
    "codex-hooks",
    "codex-config",
];

#[test]
fn shape_g_reaches_all_unchanged() {
    let fx = shape_g();
    let before: Value = serde_json::from_slice(&fs::read(fx.root.join(SETTINGS)).unwrap()).unwrap();
    let first = fx.update_plan();
    // The reproduction's loop: the Go keys hold the folder, and the first command cleans them.
    assert_eq!(
        step_in(&first, 0, "go-settings")["state"],
        "ready",
        "{first}"
    );
    assert_eq!(
        step_in(&first, 0, "go-state")["state"],
        "blocked",
        "{first}"
    );
    let w = assert_walks_to_the_end(&fx, &ORDER);
    // The status line the clean-up wrote is the Rust launcher's; the Go keys are gone.
    let after: Value = serde_json::from_slice(&fs::read(fx.root.join(SETTINGS)).unwrap()).unwrap();
    for key in ["plansDirectory", "subagentStatusLine"] {
        assert!(
            after.get(key).is_none() && before.get(key).is_some(),
            "{key}"
        );
    }
    assert_eq!(after["model"], json!("opus"));
    assert_eq!(after["permissions"], before["permissions"]);
    assert!(
        !after["statusLine"].to_string().contains(".moai/"),
        "{after}"
    );
    // What the Go tool left behind is told as facts and still exists: the earlier state folder
    // (empty here) and, in the backup, the Go folder.
    assert_eq!(leftover_paths(&w.last), [".moai-cli/"], "{}", w.last);
    let fact = leftover(&w.last, ".moai-cli/")["fact"].as_str().unwrap();
    assert!(fact.contains("empty folder"), "{fact}");
    assert!(fx.root.join(".moai-cli").is_dir());
    assert!(fx.root.join(".moai-backup/manifest.json").is_file());
    assert!(
        fx.root.join(".moai/plans/a.md").is_file(),
        "the plans were moved, not deleted"
    );
}

#[test]
fn shape_h_reaches_all_unchanged() {
    let fx = shape_h();
    let first = fx.update_plan();
    assert_eq!(
        step_in(&first, 0, "go-settings")["state"],
        "ready",
        "{first}"
    );
    let w = assert_walks_to_the_end(&fx, &ORDER);
    // The Rust settings receipt still owns `permissions`, and the file keeps the user's rules.
    let after: Value = serde_json::from_slice(&fs::read(fx.root.join(SETTINGS)).unwrap()).unwrap();
    assert_eq!(
        after["permissions"],
        go_settings()["permissions"],
        "{after}"
    );
    assert_eq!(leftover_paths(&w.last), [".moai-cli/"], "{}", w.last);
}

#[test]
fn shape_e_reaches_all_unchanged() {
    let (fx, pid) = shape_e();
    let first = fx.update_plan();
    // Half moved: the rule file is gone from the work tree, the earlier state folder is there
    // and the dead lock is stated as a fact, not removed.
    assert!(!fx.root.join(".claude/rules/moai/r1.md").exists());
    assert_eq!(
        step_in(&first, 0, "go-settings")["state"],
        "ready",
        "{first}"
    );
    let state = step_in(&first, 0, "go-state");
    assert_eq!(state["state"], "blocked", "{state}");
    assert!(
        state["reason"]
            .as_str()
            .unwrap()
            .contains("held by .claude/settings.json"),
        "{state}"
    );
    assert_eq!(leftover_paths(&first), [".moai/.update.lock"], "{first}");
    let fact = leftover(&first, ".moai/.update.lock")["fact"]
        .as_str()
        .unwrap();
    assert!(
        fact.contains(&format!("process {pid}")) && fact.contains("not running"),
        "{fact}"
    );
    assert_eq!(
        leftover(&first, ".moai/.update.lock")["note"],
        json!("not removed by update")
    );
    assert!(
        fx.root.join(".moai/.update.lock").is_file(),
        "the preview removed nothing"
    );

    let w = assert_walks_to_the_end(&fx, &ORDER);
    // The lock went into the Go backup with the folder; nothing deleted it. The earlier receipt
    // folder of the first half move is still in the earlier state folder.
    assert!(fx.root.join(".moai-backup/.update.lock").is_file());
    assert_eq!(leftover_paths(&w.last), [".moai-cli/"], "{}", w.last);
    let fact = leftover(&w.last, ".moai-cli/")["fact"].as_str().unwrap();
    assert!(fact.contains("still holds 1 entries (backups)"), "{fact}");
}

// ---- the local settings file ----

/// The local file that holds Go references (D1 condition 3: it is never written).
const LOCAL_WITH_GO_REFS: &str = "{\"statusLine\":{\"type\":\"command\",\"command\":\"$CLAUDE_PROJECT_DIR/.moai/status_line.sh\"},\"plansDirectory\":\".moai/plans\"}\n";

/// Go references in `settings.local.json` hold the Go folder: the update never edits that file, so
/// the loop ends with one blocked step whose reason names the file and the keys, says the update
/// does not edit it and gives the command that follows the user's edit. After the edit the loop
/// reaches the end.
#[test]
fn go_references_in_the_local_settings_are_named_and_left_to_the_user() {
    let fx = shape_g();
    fx.write(LOCAL, LOCAL_WITH_GO_REFS);
    let w = walk(&fx);
    // The settings clean-up ran first. The other moves (rule files and the like) do not depend on
    // the folder and went through; the folder itself stays while the local file names it.
    assert_eq!(w.order, ORDER, "{:?}", open_steps(&w.last));
    assert_eq!(w.last["all_unchanged"], json!(false), "{}", w.last);
    let state = step_in(&w.last, 0, "go-state");
    assert_eq!(state["state"], "blocked", "{state}");
    assert_eq!(state["apply"], Value::Null, "{state}");
    let why = state["reason"].as_str().unwrap();
    assert!(
        why.contains("held by .claude/settings.local.json (plansDirectory, statusLine)")
            && why.contains(
                "this update does not edit .claude/settings.local.json; edit or remove these \
                 keys yourself (plansDirectory, statusLine), then run `moai project-update-plan` \
                 again"
            ),
        "{why}"
    );
    // Only the local file holds Go keys, so the sentence about keys "this update does not
    // change" (for the project settings) is not there.
    assert!(
        !why.contains("this update does not change those keys"),
        "{why}"
    );
    // The values are never printed.
    assert!(
        !w.last.to_string().contains("CLAUDE_PROJECT_DIR"),
        "{}",
        w.last
    );
    // The file is byte for byte the user's.
    assert_eq!(
        fs::read_to_string(fx.root.join(LOCAL)).unwrap(),
        LOCAL_WITH_GO_REFS
    );
    // That blocked step is the only one left, and the Go folder is where it was.
    assert_eq!(
        open_steps(&w.last),
        [("go-state".to_string(), "blocked".to_string())],
        "{}",
        w.last
    );
    assert!(fx.root.join(".moai/manifest.json").is_file());

    // The user edits the file as the reason says; the same loop now reaches the end.
    fx.write(LOCAL, LOCAL_BODY);
    let rest = walk(&fx);
    assert_eq!(rest.last["all_unchanged"], json!(true), "{}", rest.last);
    // The install steps ran while the folder was held, with the workflow reference in the earlier
    // state folder; once that state moved, the generation step follows it to `.moai/` and
    // rewrites the managed section instead of waiting for a folder that is gone.
    assert_eq!(
        rest.order,
        ["go-state", "generation"],
        "{:?}",
        open_steps(&rest.last)
    );
    let agents = fs::read_to_string(fx.root.join("AGENTS.md")).unwrap();
    assert!(
        agents.contains(".moai/workflows/project/") && !agents.contains(".moai-cli/workflows"),
        "{agents}"
    );
    assert!(fx.root.join(".moai/workflows/project").is_dir());
    assert_eq!(rest.last["state_root"]["rust_marker"], json!("valid"));
}

// ---- undo (AUDIT #3117 E1) ----

/// Every entry of the project, git's folder left out: a file's bytes, a link's target, a folder.
fn project_nodes(fx: &GoFx) -> BTreeMap<String, Node> {
    fx.tree()
        .into_iter()
        .filter(|(p, _)| p.starts_with(&fx.root))
        .filter(|(p, _)| !p.components().any(|c| c.as_os_str() == ".git"))
        .map(|(p, n)| (p.strip_prefix(&fx.root).unwrap().display().to_string(), n))
        .collect()
}

/// What the chain adds and an undo does not take away: the receipts, the set-aside folder of the
/// Rust state, and the files of the three host steps that have no revert command.
fn may_stay_behind(rel: &str) -> bool {
    rel == ".codex"
        || rel.starts_with(".codex/")
        || rel == ".omp"
        || rel.starts_with(".omp/")
        || rel.starts_with(".moai-cli/receipts")
        || rel.starts_with(".moai.rust-")
        || matches!(
            rel,
            ".moai-cli/manifests/codex-config.json"
                | ".moai-cli/manifests/codex-hooks.json"
                | ".moai-cli/manifests/omp-hooks.json"
        )
}

fn applied<'a>(w: &'a Walk, step: &str) -> &'a Value {
    &w.applied
        .iter()
        .find(|(name, _)| name == step)
        .unwrap_or_else(|| panic!("{step} was not applied: {:?}", w.order))
        .1
}

/// The preview prints how to take the chain back, last applied first; running exactly that, with
/// the ids the apply results printed, leaves every byte the project started with in place. Only
/// what the steps added (receipts, the Rust state's set-aside folder, the host files that have no
/// revert command) is left. Out of that order the go-state revert is refused.
#[test]
fn the_printed_undo_order_takes_the_chain_back_to_the_start_bytes() {
    let fx = shape_g();
    // The user's own instruction file: the move keeps it in the backup, the generation step writes
    // a managed one in its place, and the go-state revert must find that one still there or gone.
    fx.write("AGENTS.md", "# AGENTS\n\nmine\n");
    let start = project_nodes(&fx);
    let w = walk(&fx);
    assert_eq!(w.order, ORDER, "{:?}", open_steps(&w.last));
    let steps = w.last["undo_order"]["steps"].as_array().unwrap();
    let printed: Vec<&str> = steps.iter().map(|s| s["step"].as_str().unwrap()).collect();
    let mut reverse = ORDER.to_vec();
    reverse.reverse();
    assert_eq!(printed, reverse, "{}", w.last["undo_order"]);
    let commands = |step: &str| -> Vec<String> {
        steps.iter().find(|s| s["step"] == step).unwrap()["commands"]
            .as_array()
            .unwrap()
            .iter()
            .map(|c| c.as_str().unwrap().to_string())
            .collect()
    };
    for step in ["codex-config", "codex-hooks"] {
        let note = steps.iter().find(|s| s["step"] == step).unwrap()["note"].as_str();
        assert!(
            note.is_some_and(|n| n.contains("no revert command")),
            "{step}"
        );
        assert_eq!(commands(step), Vec::<String>::new(), "{step}");
    }
    // What stays is said, not left to be read as "everything is back".
    let stays = w.last["undo_order"]["stays"].as_str().unwrap();
    assert!(
        stays.contains("receipts") && stays.contains(".moai.rust-"),
        "{stays}"
    );
    let id = |step: &str, key: &str| applied(&w, step)[key].as_str().unwrap().to_string();
    let fills = [
        ("<receipt of the hooks step>", id("hooks", "receipt")),
        (
            "<agents_receipt of the generation step>",
            id("generation", "agents_receipt"),
        ),
        (
            "<receipt of the generation step>",
            id("generation", "receipt"),
        ),
        ("<backup of the go-state step>", id("go-state", "backup")),
        (
            "<receipt of the go-settings step>",
            id("go-settings", "receipt"),
        ),
        ("<reference>", REF.to_string()),
        ("<PROJECT>", fx.p().to_string()),
    ];
    // A printed command, run as printed: its placeholders filled (the digest from the plan that
    // came just before it), no placeholder left, the first word (`moai`) dropped.
    let digest = std::cell::RefCell::new(String::new());
    let run_printed = |cmd: &str| -> Value {
        let mut line = cmd.to_string();
        for (placeholder, value) in &fills {
            line = line.replace(placeholder, value);
        }
        line = line.replace("<digest of that plan>", &digest.borrow());
        assert!(!line.contains('<'), "a placeholder is left: {line}");
        let args: Vec<&str> = line.split_whitespace().skip(1).collect();
        let out = fx.run(&args);
        if let Some(d) = out["result"]["plan_digest"].as_str() {
            *digest.borrow_mut() = d.to_string();
        }
        out
    };

    // Out of order the go-state revert is refused and writes nothing.
    let before = project_nodes(&fx);
    let early = commands("go-state");
    assert_eq!(early.len(), 2, "{early:?}");
    assert_eq!(run_printed(&early[0])["ok"], json!(true));
    let refused = run_printed(&early[1]);
    assert_eq!(refused["ok"], json!(false), "{refused}");
    assert!(
        refused["error"]["message"]
            .as_str()
            .unwrap_or_default()
            .contains("AGENTS.md is conflict"),
        "{refused}"
    );
    assert_eq!(
        project_nodes(&fx),
        before,
        "a refused revert writes nothing"
    );

    // The printed order: every printed command of every step, one after the other.
    let mut ran = 0;
    for step in steps {
        for cmd in step["commands"].as_array().unwrap() {
            let out = run_printed(cmd.as_str().unwrap());
            assert_eq!(out["ok"], json!(true), "{}: {cmd}: {out}", step["step"]);
            ran += 1;
        }
    }
    assert_eq!(ran, 6, "hooks 1, generation 2, go-state 2, go-settings 1");

    let end = project_nodes(&fx);
    for (path, node) in &start {
        assert_eq!(end.get(path), Some(node), "{path} is not as it started");
    }
    let extra: Vec<&String> = end
        .keys()
        .filter(|p| !start.contains_key(*p) && !may_stay_behind(p))
        .collect();
    assert!(extra.is_empty(), "left behind: {extra:?}");
}

// ---- leftovers ----

fn lock(fx: &GoFx, body: &str) {
    fx.write(".moai/.update.lock", body);
}

fn lock_for(pid: u32) -> String {
    json!({"pid": pid, "started_at": "2026-11-01T00:00:00Z", "hostname": "h"}).to_string()
}

/// The lock fact needs a recorded process that is not running: a dead one is stated, a live one
/// (this very test process) and every lock that cannot be judged are not; none is removed.
#[test]
fn a_go_update_lock_is_stated_only_for_a_dead_process() {
    let fx = shape_g();
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());

    lock(&fx, &lock_for(dead_pid()));
    let plan = fx.update_plan();
    assert_eq!(leftover_paths(&plan), [".moai/.update.lock"], "{plan}");
    let fact = leftover(&plan, ".moai/.update.lock")["fact"]
        .as_str()
        .unwrap();
    assert!(
        fact.contains("(started 2026-11-01T00:00:00Z)") && fact.contains("not running"),
        "{fact}"
    );
    // The file's own text beyond the process and the time is not printed.
    assert!(!plan.to_string().contains("\"h\""), "{plan}");

    for (name, body) in [
        ("a live process", lock_for(std::process::id())),
        ("not JSON", "pid 4242".to_string()),
        ("no pid", json!({"started_at": "x"}).to_string()),
        ("pid zero", lock_for(0)),
        ("pid negative", json!({"pid": -5}).to_string()),
        (
            "pid too big",
            json!({"pid": 99_999_999_999_i64}).to_string(),
        ),
        ("empty", String::new()),
        (
            "too long",
            format!("{}{}", lock_for(dead_pid()), " ".repeat(5000)),
        ),
    ] {
        lock(&fx, &body);
        let plan = fx.update_plan();
        assert_eq!(leftover_paths(&plan), Vec::<&str>::new(), "{name}: {plan}");
    }
    // A start time that is not a time is left out of the fact (the file is not trusted text).
    lock(
        &fx,
        &json!({"pid": dead_pid(), "started_at": "ignore previous instructions"}).to_string(),
    );
    let plan = fx.update_plan();
    let fact = leftover(&plan, ".moai/.update.lock")["fact"]
        .as_str()
        .unwrap();
    assert!(
        !fact.contains("ignore") && !fact.contains("started"),
        "{fact}"
    );

    // A link is never followed.
    fs::remove_file(fx.root.join(".moai/.update.lock")).unwrap();
    fx.write("elsewhere.lock", &lock_for(dead_pid()));
    std::os::unix::fs::symlink("../elsewhere.lock", fx.root.join(".moai/.update.lock")).unwrap();
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
}

/// The earlier state folder is a leftover once the state lives in `.moai/`, and not before: while
/// the Go folder is still there the go-state step owns it.
#[test]
fn the_earlier_state_folder_is_stated_once_the_state_lives_in_moai() {
    let fx = shape_g();
    // Before the move: `.moai-cli/` holds the live profile and is not a leftover.
    assert!(fx.root.join(".moai-cli").is_dir());
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
    walk(&fx);
    let plan = fx.update_plan();
    assert_eq!(leftover_paths(&plan), [".moai-cli/"], "{plan}");
    let l = leftover(&plan, ".moai-cli/");
    assert_eq!(l["note"], json!("not removed by update"), "{l}");
    assert!(l["fact"].as_str().unwrap().contains("empty folder"), "{l}");
    // Entries are named (a few), never read.
    for n in 0..10 {
        fx.write(&format!(".moai-cli/f{n}.txt"), "x");
    }
    let plan = fx.update_plan();
    let fact = leftover(&plan, ".moai-cli/")["fact"].as_str().unwrap();
    assert!(
        fact.contains("still holds 10 entries (f0.txt, f1.txt,"),
        "{fact}"
    );
    assert!(fact.contains("and 2 more)"), "{fact}");
    assert!(
        fx.root.join(".moai-cli/f9.txt").is_file(),
        "nothing was removed"
    );
    // AUDIT #3117 E4: a file name is the project's text. A long one is cut, and a control character
    // (an escape sequence) shows as `?`; the file itself is not touched.
    // Sorts before the f*.txt names, so it is among the eight shown.
    let long = "0".repeat(200);
    fx.write(&format!(".moai-cli/{long}"), "x");
    fx.write(".moai-cli/\u{1b}[31mred", "x");
    let plan = fx.update_plan();
    let fact = leftover(&plan, ".moai-cli/")["fact"].as_str().unwrap();
    assert!(!fact.contains('\u{1b}'), "{fact:?}");
    assert!(fact.contains("?[31mred"), "{fact:?}");
    assert!(!fact.contains(&"0".repeat(60)), "{fact:?}");
    assert!(fact.contains('…'), "{fact:?}");
    assert!(fx.root.join(format!(".moai-cli/{long}")).is_file());
    fs::remove_file(fx.root.join(format!(".moai-cli/{long}"))).unwrap();
    fs::remove_file(fx.root.join(".moai-cli/\u{1b}[31mred")).unwrap();
    // Gone, or a link: nothing to state.
    for n in 0..10 {
        fs::remove_file(fx.root.join(format!(".moai-cli/f{n}.txt"))).unwrap();
    }
    fs::remove_dir(fx.root.join(".moai-cli")).unwrap();
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
    fx.write("elsewhere/x", "x");
    std::os::unix::fs::symlink("elsewhere", fx.root.join(".moai-cli")).unwrap();
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
}

const CURATOR_BLOCK: &str =
    "# Notes\n<!-- moai:learned-start -->\nlearned\n<!-- moai:learned-end -->\n";

/// A root `CLAUDE.md` is a leftover when it has Go proof and the Go folder is not there for the
/// go-state step to move it with.
#[test]
fn a_go_claude_md_is_stated_only_with_go_proof_and_without_the_go_folder() {
    // A Go curator block, no `.moai/` at all.
    let fx = GoFx::new();
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
    fx.write("CLAUDE.md", CURATOR_BLOCK);
    let plan = fx.update_plan();
    assert_eq!(leftover_paths(&plan), ["CLAUDE.md"], "{plan}");
    let l = leftover(&plan, "CLAUDE.md");
    assert!(
        l["fact"].as_str().unwrap().contains("Go curator block"),
        "{l}"
    );
    assert_eq!(l["note"], json!("not removed by update"), "{l}");
    assert_eq!(
        fs::read_to_string(fx.root.join("CLAUDE.md")).unwrap(),
        CURATOR_BLOCK
    );

    // No proof: the user's own file.
    fx.write("CLAUDE.md", "# My project\n");
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
    // This generation's managed section wins over a Go block.
    fx.write(
        "CLAUDE.md",
        &format!("<!-- moai:managed:begin id=x -->\n{CURATOR_BLOCK}"),
    );
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());
    // A link is never followed.
    fs::remove_file(fx.root.join("CLAUDE.md")).unwrap();
    fx.write("other.md", CURATOR_BLOCK);
    std::os::unix::fs::symlink("other.md", fx.root.join("CLAUDE.md")).unwrap();
    assert_eq!(leftover_paths(&fx.update_plan()), Vec::<&str>::new());

    // While the Go folder is there the go-state step owns the file, whatever the proof.
    let mut go = GoFx::new();
    go.deployed("CLAUDE.md", "# Go CLAUDE.md\n");
    go.manifest();
    go.write("CLAUDE.md", CURATOR_BLOCK);
    assert_eq!(leftover_paths(&go.update_plan()), Vec::<&str>::new());

    // The Go manifest records the file next to a Rust marker (state_root `mixed`): proof by
    // manifest, and the folder is not moved by the go-state step.
    let mut mixed = GoFx::new();
    mixed.deployed("CLAUDE.md", "# Go CLAUDE.md\n");
    mixed.manifest();
    mixed.write(
        ".moai/moai.json",
        &json!({"schema_version": 1, "generation": "moai"}).to_string(),
    );
    let plan = mixed.update_plan();
    assert_eq!(plan["state_root"]["generation"], json!("mixed"), "{plan}");
    let l = leftover(&plan, "CLAUDE.md");
    assert!(
        l["fact"]
            .as_str()
            .unwrap()
            .contains("the Go manifest records it"),
        "{l}"
    );
}
