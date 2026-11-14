//! #2875 (#2871 I3): the MoAI-Easy output style in the user layer, against a temporary HOME
//! only — never the real one. Covers create, idempotence, a changed source, the user's own
//! style file, setting and hand-written block (conflicts, bytes kept), symlinks, a stale digest,
//! and revert.

mod common;
use moai_cli::home_style::{
    apply, plan, render_block, render_style_file, revert, Plan, SOURCE, STYLE_NAME,
};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const STYLE: &str = ".claude/output-styles/MoAI-Easy.md";
const SETTINGS: &str = ".claude/settings.json";
const AGENTS: &str = ".codex/AGENTS.md";

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let path = fs::canonicalize(dir.path()).unwrap();
    (dir, path)
}

fn action(p: &Plan, kind: &str) -> String {
    p.targets
        .iter()
        .find(|t| t.kind == kind)
        .map(|t| t.action.clone())
        .unwrap()
}

fn detail(p: &Plan, kind: &str) -> String {
    p.targets
        .iter()
        .find(|t| t.kind == kind)
        .and_then(|t| t.detail.clone())
        .unwrap_or_default()
}

fn src() -> String {
    fs::read_to_string(package().join(SOURCE)).unwrap()
}

fn settings(home: &Path) -> Value {
    serde_json::from_slice(&fs::read(home.join(SETTINGS)).unwrap()).unwrap()
}

fn write(home: &Path, rel: &str, bytes: &str) {
    let p = home.join(rel);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(p, bytes).unwrap();
}

/// A temporary package holding only what a plan reads: `SKILL.md` and the style source.
fn temp_package(source: &str) -> tempfile::TempDir {
    let pkg = tempfile::tempdir().unwrap();
    fs::write(pkg.path().join("SKILL.md"), "package\n").unwrap();
    let p = pkg.path().join(SOURCE);
    fs::create_dir_all(p.parent().unwrap()).unwrap();
    fs::write(&p, source).unwrap();
    pkg
}

#[test]
fn a_fresh_home_gets_the_style_file_the_default_and_the_codex_block_once() {
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "create");
    assert_eq!(action(&p, "setting"), "set");
    assert_eq!(action(&p, "block"), "create");
    let hosts: Vec<&str> = p.not_supported.iter().map(|u| u.host.as_str()).collect();
    assert_eq!(hosts, ["omp", "kiro", "cursor"]);
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    assert_eq!(
        fs::read_to_string(home.join(STYLE)).unwrap(),
        render_style_file(&src())
    );
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));
    assert_eq!(
        fs::read_to_string(home.join(AGENTS)).unwrap(),
        render_block(&src())
    );
    // One source for both hosts: the Codex block carries the same text as the style file body.
    let style = fs::read_to_string(home.join(STYLE)).unwrap();
    assert!(style.ends_with(&src()));
    assert!(fs::read_to_string(home.join(AGENTS))
        .unwrap()
        .contains(&src()));
    let again = plan(&home, &package()).unwrap();
    for kind in ["style_file", "setting", "block"] {
        assert_eq!(action(&again, kind), "unchanged", "{kind}");
    }
    let r = apply(&home, &package(), &again.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
}

#[test]
fn the_other_settings_keys_keep_their_values_and_order() {
    let (_t, home) = home();
    write(
        &home,
        SETTINGS,
        "{\"z\": 1, \"hooks\": {\"Stop\": []}, \"a\": \"keep\"}\n",
    );
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "set");
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let s = settings(&home);
    let keys: Vec<&String> = s.as_object().unwrap().keys().collect();
    assert_eq!(keys, ["z", "hooks", "a", "outputStyle"]);
    assert_eq!(s["hooks"], json!({"Stop": []}));
    assert_eq!(s["a"], "keep");
}

#[test]
fn the_users_own_choices_are_conflicts_and_keep_every_byte() {
    let (_t, home) = home();
    let style = "---\nname: MoAI-Easy\n---\n\nmy own style\n";
    let set = "{\"outputStyle\": \"Explanatory\"}\n";
    let agents = "notes\n<!-- moai:output-style moai-easy BEGIN -->\nmine\n<!-- moai:output-style moai-easy END -->\n";
    write(&home, STYLE, style);
    write(&home, SETTINGS, set);
    write(&home, AGENTS, agents);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "conflict");
    // The default is never pointed at a style file MoAI does not own.
    assert_eq!(action(&p, "setting"), "skipped");
    assert_eq!(action(&p, "block"), "conflict");
    assert!(detail(&p, "block").contains("hand-written"));
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), style);
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), set);
    assert_eq!(fs::read_to_string(home.join(AGENTS)).unwrap(), agents);
    // With the style file gone, a different outputStyle is still the user's choice.
    fs::remove_file(home.join(STYLE)).unwrap();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "create");
    assert_eq!(action(&p, "setting"), "conflict");
    assert!(detail(&p, "setting").contains("Explanatory"));
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), set);
}

#[test]
fn a_settings_file_that_is_not_an_object_is_a_conflict() {
    let (_t, home) = home();
    write(&home, SETTINGS, "[1, 2]\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "conflict");
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), "[1, 2]\n");
}

#[test]
fn a_changed_source_updates_only_what_moai_wrote() {
    let (_t, home) = home();
    let one = temp_package("## Style\n\nOne.\n");
    let p = plan(&home, one.path()).unwrap();
    apply(&home, one.path(), &p.plan_digest, PERM).unwrap();
    let two = temp_package("## Style\n\nTwo.\n");
    let p = plan(&home, two.path()).unwrap();
    assert_eq!(action(&p, "style_file"), "merge");
    assert_eq!(action(&p, "setting"), "unchanged");
    assert_eq!(action(&p, "block"), "merge");
    let r = apply(&home, two.path(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    assert!(fs::read_to_string(home.join(STYLE))
        .unwrap()
        .ends_with("Two.\n"));
    // An edit after that apply makes the next source change a conflict, never an overwrite.
    let edited = fs::read_to_string(home.join(STYLE)).unwrap() + "my line\n";
    fs::write(home.join(STYLE), &edited).unwrap();
    let three = temp_package("## Style\n\nThree.\n");
    let p = plan(&home, three.path()).unwrap();
    assert_eq!(action(&p, "style_file"), "conflict");
    apply(&home, three.path(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), edited);
}

#[cfg(unix)]
#[test]
fn a_symlinked_folder_or_file_blocks_its_target() {
    let (_o, other) = home();
    let (_t, home) = home();
    fs::create_dir_all(home.join(".claude")).unwrap();
    std::os::unix::fs::symlink(&other, home.join(".claude/output-styles")).unwrap();
    fs::create_dir_all(home.join(".codex")).unwrap();
    fs::write(other.join("agents.md"), "elsewhere\n").unwrap();
    std::os::unix::fs::symlink(other.join("agents.md"), home.join(AGENTS)).unwrap();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "blocked");
    assert_eq!(action(&p, "setting"), "skipped");
    assert_eq!(action(&p, "block"), "blocked");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    // Nothing went through the linked folder.
    assert!(!other.join("MoAI-Easy.md").exists());
    assert_eq!(
        fs::read_to_string(other.join("agents.md")).unwrap(),
        "elsewhere\n"
    );
}

#[test]
fn a_stale_digest_a_changed_file_or_a_missing_permission_is_refused() {
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    assert!(apply(&home, &package(), "0", PERM)
        .unwrap_err()
        .contains("stale"));
    assert!(apply(&home, &package(), &p.plan_digest, "")
        .unwrap_err()
        .contains("permission"));
    write(&home, SETTINGS, "{}\n");
    assert!(apply(&home, &package(), &p.plan_digest, PERM)
        .unwrap_err()
        .contains("stale"));
    assert!(!home.join(STYLE).exists());
}

#[test]
fn revert_removes_what_it_created_and_restores_what_it_changed() {
    let (_t, home) = home();
    let user_agents = "Codex notes\n";
    write(&home, AGENTS, user_agents);
    write(&home, SETTINGS, "{\"a\": 1}\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "block"), "append");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    // Claude Code rewrites settings.json on its own: the key still goes, the new key stays.
    write(
        &home,
        SETTINGS,
        "{\"a\": 1, \"outputStyle\": \"MoAI-Easy\", \"b\": 2}\n",
    );
    let r = revert(&home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted");
    assert!(!home.join(STYLE).exists());
    assert_eq!(settings(&home), json!({"a": 1, "b": 2}));
    assert_eq!(fs::read_to_string(home.join(AGENTS)).unwrap(), user_agents);
    assert!(revert(&home, &id, PERM)
        .unwrap_err()
        .contains("already reverted"));
    // After the revert nothing is owned: a plan creates again.
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "create");
    assert_eq!(action(&p, "block"), "append");
}

#[test]
fn revert_skips_whatever_the_user_changed_since() {
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    let style = fs::read_to_string(home.join(STYLE)).unwrap() + "mine\n";
    fs::write(home.join(STYLE), &style).unwrap();
    write(&home, SETTINGS, "{\"outputStyle\": \"Learning\"}\n");
    let r = revert(&home, &id, PERM).unwrap();
    let skipped: Vec<&str> = r["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["target"].as_str().unwrap())
        .collect();
    assert_eq!(skipped, ["claude:setting", "claude:style_file"]);
    assert_eq!(r["reverted"], json!(["codex:block"]));
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), style);
    assert_eq!(settings(&home), json!({"outputStyle": "Learning"}));
    assert!(!home.join(AGENTS).exists());
}

#[test]
fn a_bad_style_source_is_refused() {
    let (_t, home) = home();
    for bad in [
        "no newline",
        "---\nname: x\n---\nbody\n",
        "<!-- moai:managed:begin id=x -->\n",
    ] {
        let pkg = temp_package(bad);
        assert!(plan(&home, pkg.path()).is_err(), "{bad:?}");
    }
}

#[test]
fn the_shipped_source_carries_the_style_items() {
    let s = src();
    for item in [
        "MoAI-Easy",
        "four steps",
        "evidence",
        "46 `─`",
        "▓▓▓▓▓▓░░░░",
        "user's conversation language",
    ] {
        assert!(s.contains(item), "{item}");
    }
}

/// The three operations, through the binary, against a temporary HOME passed as an argument.
#[test]
fn the_operations_run_through_the_binary() {
    let (_t, home) = home();
    let bin = crate::common::MOAI_BIN;
    let run = |args: &[&str]| {
        let out = std::process::Command::new(bin)
            .args(args)
            .env("HOME", &home)
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        serde_json::from_slice::<Value>(&out.stdout).unwrap()
    };
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    let p = run(&["home-style-plan", h, pkg]);
    assert_eq!(p["operation"], "home.style_plan");
    let digest = p["result"]["plan_digest"].as_str().unwrap().to_string();
    let a = run(&[
        "home-style-apply",
        h,
        pkg,
        "--plan-digest",
        &digest,
        "--permission-reference",
        PERM,
    ]);
    assert_eq!(a["operation"], "home.style_apply");
    let id = a["result"]["receipt"].as_str().unwrap().to_string();
    let r = run(&[
        "home-style-revert",
        h,
        "--receipt",
        &id,
        "--permission-reference",
        PERM,
    ]);
    assert_eq!(r["result"]["state"], "reverted");
    assert!(!home.join(STYLE).exists() && !home.join(SETTINGS).exists());
}
