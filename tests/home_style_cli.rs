//! #2875 (#2871 I3): the MoAI-Style output style in the user layer, against a temporary HOME
//! only — never the real one. Covers create, idempotence, a changed source, the user's own
//! style file and setting (conflicts, bytes kept), symlinks, a stale digest, and revert. The
//! Codex block is no longer written (SPEC-20261122-codex-home-update-removed).
//!
//! SPEC-20261127-output-style-moai-style: the style was named MoAI-Easy before; the tests named
//! `ac1_` to `ac8_` at the end cover the move of an installed old file and old value (plan,
//! apply, revert, receipts of the earlier format, links). The old name is written out below only
//! as that move's source (`OLD_*`).

mod common;
use moai_cli::home_style::{apply, plan, render_style_file, revert, Plan, SOURCE, STYLE_NAME};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const STYLE: &str = ".claude/output-styles/MoAI-Style.md";
const SETTINGS: &str = ".claude/settings.json";
/// The earlier style file and `outputStyle` value: what `moai update` moves away from.
const OLD_STYLE: &str = ".claude/output-styles/MoAI-Easy.md";
const OLD_NAME: &str = "MoAI-Easy";

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
fn a_fresh_home_gets_the_style_file_and_the_default_once() {
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "create");
    assert_eq!(action(&p, "setting"), "set");
    let kinds: Vec<&str> = p.targets.iter().map(|t| t.kind.as_str()).collect();
    assert_eq!(kinds, ["style_file", "setting"], "no Codex block target");
    let hosts: Vec<&str> = p.not_supported.iter().map(|u| u.host.as_str()).collect();
    assert_eq!(hosts, ["kiro", "cursor"]);
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    assert_eq!(
        fs::read_to_string(home.join(STYLE)).unwrap(),
        render_style_file(&src())
    );
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));
    let style = fs::read_to_string(home.join(STYLE)).unwrap();
    assert!(style.ends_with(&src()));
    assert!(!home.join(".codex").exists(), "no Codex file is made");
    let again = plan(&home, &package()).unwrap();
    for kind in ["style_file", "setting"] {
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
    let style = "---\nname: MoAI-Style\n---\n\nmy own style\n";
    let set = "{\"outputStyle\": \"Explanatory\"}\n";
    write(&home, STYLE, style);
    write(&home, SETTINGS, set);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "conflict");
    // The default is never pointed at a style file MoAI does not own.
    assert_eq!(action(&p, "setting"), "skipped");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), style);
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), set);
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
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "blocked");
    assert_eq!(action(&p, "setting"), "skipped");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    // Nothing went through the linked folder.
    assert!(!other.join("MoAI-Style.md").exists());
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
    write(&home, SETTINGS, "{\"a\": 1}\n");
    let p = plan(&home, &package()).unwrap();
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let id = r["receipt"].as_str().unwrap().to_string();
    // Claude Code rewrites settings.json on its own: the key still goes, the new key stays.
    write(
        &home,
        SETTINGS,
        "{\"a\": 1, \"outputStyle\": \"MoAI-Style\", \"b\": 2}\n",
    );
    let r = revert(&home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted");
    assert!(!home.join(STYLE).exists());
    assert_eq!(settings(&home), json!({"a": 1, "b": 2}));
    assert!(revert(&home, &id, PERM)
        .unwrap_err()
        .contains("already reverted"));
    // After the revert nothing is owned: a plan creates again.
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "create");
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
    assert_eq!(r["reverted"], json!([]));
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), style);
    assert_eq!(settings(&home), json!({"outputStyle": "Learning"}));
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
        "MoAI 🗿",
        "four steps",
        "evidence",
        "46 `─`",
        "▓▓▓▓▓▓░░░░",
        "the user's language",
    ] {
        assert!(s.contains(item), "{item}");
    }
}

/// SPEC-20261122-output-style-install-checks AC-1 (REQ-1): `outputStyle` is case-sensitive, so a
/// value that differs only in case is a conflict whose detail says so; the exact name is fine.
#[test]
fn the_setting_plan_says_the_style_name_is_case_sensitive() {
    for wrong in ["moai-style", "MOAI-STYLE"] {
        let (_t, home) = home();
        write(
            &home,
            SETTINGS,
            &format!("{{\"outputStyle\": \"{wrong}\"}}\n"),
        );
        let p = plan(&home, &package()).unwrap();
        assert_eq!(action(&p, "setting"), "conflict", "{wrong}");
        let d = detail(&p, "setting");
        assert!(
            d.contains("case-sensitive") && d.contains("default style"),
            "{d}"
        );
    }
    let (_t, home) = home();
    write(&home, SETTINGS, "{\"outputStyle\": \"MoAI-Style\"}\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "unchanged");
    // A different name is still only the user's choice, with no case note.
    write(&home, SETTINGS, "{\"outputStyle\": \"Explanatory\"}\n");
    let d = detail(&plan(&home, &package()).unwrap(), "setting");
    assert!(
        d.contains("Explanatory") && !d.contains("case-sensitive"),
        "{d}"
    );
}

/// A temporary project folder with the named settings files.
fn project_with(files: &[(&str, &str)]) -> (tempfile::TempDir, PathBuf) {
    let (t, path) = home();
    for (rel, bytes) in files {
        write(&path, rel, bytes);
    }
    (t, path)
}

fn reasons(notices: &[Value]) -> Vec<String> {
    notices
        .iter()
        .map(|n| n["reason"].as_str().unwrap().to_string())
        .collect()
}

/// AC-2 (REQ-2): a project value other than the style hides it in that project; the notice names
/// the file and the value, and neither file is written.
#[test]
fn a_project_setting_that_hides_the_style_gives_a_notice_and_writes_nothing() {
    let local = "{\"outputStyle\": \"Learning\"}\n";
    let shared = "{\"outputStyle\": \"Explanatory\", \"x\": 1}\n";
    let (_t, project) = project_with(&[
        (".claude/settings.local.json", local),
        (".claude/settings.json", shared),
    ]);
    let (_h, home) = home();
    let p = plan(&home, &package()).unwrap();
    let notices = moai_cli::home_style::notices(&p, &project);
    let r = reasons(&notices);
    assert!(
        r.iter().any(|s| s.contains("settings.local.json")
            && s.contains("Learning")
            && s.contains("this project uses")),
        "{r:?}"
    );
    assert!(
        r.iter()
            .any(|s| s.contains("settings.json") && s.contains("Explanatory")),
        "{r:?}"
    );
    assert!(notices
        .iter()
        .all(|n| n["part"] == "style" && n["host"] == "claude"));
    assert_eq!(
        fs::read_to_string(project.join(".claude/settings.local.json")).unwrap(),
        local
    );
    assert_eq!(
        fs::read_to_string(project.join(".claude/settings.json")).unwrap(),
        shared
    );
    // The same name, or no `outputStyle`, hides nothing.
    let (_t2, quiet) = project_with(&[
        (
            ".claude/settings.local.json",
            "{\"outputStyle\": \"MoAI-Style\"}\n",
        ),
        (".claude/settings.json", "{\"a\": 1}\n"),
    ]);
    assert!(!reasons(&moai_cli::home_style::notices(&p, &quiet))
        .iter()
        .any(|s| s.contains("this project uses")));
}

/// AC-2 (REQ-3): a file that cannot be read as a JSON object, or has no string value, gives no
/// notice and does not stop the rest.
#[test]
fn an_unreadable_project_setting_gives_no_notice() {
    let (_h, home) = home();
    let p = plan(&home, &package()).unwrap();
    for bytes in ["{", "[1]", "{\"outputStyle\": 3}\n", ""] {
        let (_t, project) = project_with(&[(".claude/settings.local.json", bytes)]);
        let n = moai_cli::home_style::notices(&p, &project);
        assert!(
            !reasons(&n).iter().any(|s| s.contains("this project uses")),
            "{bytes:?}"
        );
    }
    // A folder in the file's place and a missing project folder are the same quiet path.
    let (_t, project) = project_with(&[]);
    fs::create_dir_all(project.join(".claude/settings.json")).unwrap();
    assert!(!reasons(&moai_cli::home_style::notices(&p, &project))
        .iter()
        .any(|s| s.contains("this project uses")));
    let gone = project.join("missing");
    let _ = moai_cli::home_style::notices(&p, &gone);
}

/// AC-3 (REQ-4): an update that creates or replaces the style file points to `/status`; one that
/// changes nothing does not.
#[test]
fn the_update_points_to_status_after_a_style_write() {
    let (_h, home) = home();
    let (_t, project) = project_with(&[]);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "create");
    assert!(reasons(&moai_cli::home_style::notices(&p, &project))
        .iter()
        .any(|s| s.contains("/status")));
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let again = plan(&home, &package()).unwrap();
    assert_eq!(action(&again, "style_file"), "unchanged");
    assert!(!reasons(&moai_cli::home_style::notices(&again, &project))
        .iter()
        .any(|s| s.contains("/status")));
    // A style file the update takes over (`merge`) is a write too and gets the pointer.
    let (_h2, taken) = project_with(&[]);
    write(&taken, STYLE, &render_style_file(&src()));
    let p = plan(&taken, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "merge");
    let notices = moai_cli::home_style::notices(&p, &project);
    assert!(reasons(&notices).iter().any(|s| s.contains("/status")));
    // Every notice names the style part and the host, the shape the update's notice rows read.
    assert!(
        notices
            .iter()
            .all(|n| n["part"] == "style" && n["host"] == "claude"),
        "{notices:?}"
    );
}

/// AC-4 (REQ-5), D-5: a receipt written before it named its folder reverts in the default
/// folder only; in another folder's environment it changes nothing.
#[test]
fn a_receipt_without_a_folder_reverts_only_in_the_default_folder() {
    let (_t, home) = home();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    let second = home.join("profiles/b");
    let (_, id) = folder_install(&home, None);
    let id = id.unwrap();
    let receipt = moai_cli::home_root::path(&home)
        .join("home-style/receipts")
        .join(&id)
        .join("receipt.json");
    let mut r: Value = serde_json::from_slice(&fs::read(&receipt).unwrap()).unwrap();
    assert_eq!(
        r["config_folder"],
        json!(home.join(".claude").to_str().unwrap())
    );
    r.as_object_mut().unwrap().remove("config_folder");
    fs::write(&receipt, serde_json::to_vec_pretty(&r).unwrap()).unwrap();
    let style = fs::read(home.join(STYLE)).unwrap();
    let h = home.to_str().unwrap();
    let args = [
        "home-style-revert",
        h,
        "--receipt",
        id.as_str(),
        "--permission-reference",
        PERM,
    ];
    let (ok, v) = folder_run_any(&home, Some(&second), &args);
    assert!(!ok, "{v}");
    assert!(v.to_string().contains("names no config folder"), "{v}");
    assert_eq!(fs::read(home.join(STYLE)).unwrap(), style);
    let v = folder_run(&home, None, &args);
    assert_eq!(v["result"]["state"], json!("reverted"), "{v}");
    assert!(!home.join(STYLE).exists());
}

/// `home-style-*` through the binary for one Claude config folder of a temporary HOME: `folder`
/// is `CLAUDE_CONFIG_DIR` (the HOME's own default folder when `None`).
fn folder_run(home: &Path, folder: Option<&Path>, args: &[&str]) -> Value {
    let mut c = std::process::Command::new(crate::common::MOAI_BIN);
    c.args(args)
        .env("HOME", home)
        .env("MOAI_ACCOUNT_HOME", home)
        .env_remove("CLAUDE_CONFIG_DIR");
    if let Some(f) = folder {
        c.env("CLAUDE_CONFIG_DIR", f);
    }
    let out = c.output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    serde_json::from_slice(&out.stdout).unwrap()
}

/// Plans and applies in one folder; returns the plan result and the receipt id (when it wrote).
fn folder_install(home: &Path, folder: Option<&Path>) -> (Value, Option<String>) {
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    let p = folder_run(home, folder, &["home-style-plan", h, pkg]);
    let digest = p["result"]["plan_digest"].as_str().unwrap().to_string();
    let a = folder_run(
        home,
        folder,
        &[
            "home-style-apply",
            h,
            pkg,
            "--plan-digest",
            &digest,
            "--permission-reference",
            PERM,
        ],
    );
    let id = a["result"]["receipt"].as_str().map(str::to_string);
    (p["result"].clone(), id)
}

fn folder_action(plan: &Value, kind: &str) -> String {
    plan["targets"]
        .as_array()
        .unwrap()
        .iter()
        .find(|t| t["kind"] == kind)
        .map(|t| t["action"].as_str().unwrap().to_string())
        .unwrap()
}

/// The record name of a folder that is not the HOME's default: the first 12 hex digits of the
/// sha256 of the folder path (the SPEC's D-1).
fn second_record_name(folder: &Path) -> String {
    let d = moai_cli::digest(folder.to_str().unwrap().as_bytes());
    format!("claude-{}.json", &d["sha256:".len()..][..12])
}

/// AC-4 (REQ-5): each config folder has its own record; one folder's edit conflicts only there,
/// and reverting one folder leaves the other's file and record alone.
#[test]
fn two_config_folders_keep_two_records() {
    let (_t, home) = home();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    let second = home.join("profiles/b");
    let records = moai_cli::home_root::path(&home).join("home-style");
    let (_, id_a) = folder_install(&home, None);
    let (_, id_b) = folder_install(&home, Some(&second));
    assert!(id_a.is_some() && id_b.is_some());
    assert!(
        records.join("claude.json").is_file(),
        "the default folder keeps its record name"
    );
    assert!(records.join(second_record_name(&second)).is_file());
    let style_b = second.join("output-styles/MoAI-Style.md");
    assert_eq!(
        fs::read(&style_b).unwrap(),
        fs::read(home.join(STYLE)).unwrap()
    );
    // An edit in folder B is a conflict in B only.
    fs::write(&style_b, "my own text\n").unwrap();
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    let in_b = folder_run(&home, Some(&second), &["home-style-plan", h, pkg]);
    assert_eq!(folder_action(&in_b["result"], "style_file"), "conflict");
    let in_a = folder_run(&home, None, &["home-style-plan", h, pkg]);
    assert_eq!(folder_action(&in_a["result"], "style_file"), "unchanged");
    // Reverting folder A leaves folder B's file and record as they are.
    fs::write(&style_b, fs::read(home.join(STYLE)).unwrap()).unwrap();
    let record_b = fs::read(records.join(second_record_name(&second))).unwrap();
    folder_run(
        &home,
        None,
        &[
            "home-style-revert",
            h,
            "--receipt",
            id_a.as_deref().unwrap(),
            "--permission-reference",
            PERM,
        ],
    );
    assert!(!home.join(STYLE).exists());
    assert!(style_b.is_file());
    assert_eq!(
        fs::read(records.join(second_record_name(&second))).unwrap(),
        record_b
    );
}

/// One `moai` run in a config folder like [`folder_run`], without asserting success: the exit
/// status and the envelope on stdout.
fn folder_run_any(home: &Path, folder: Option<&Path>, args: &[&str]) -> (bool, Value) {
    let mut c = std::process::Command::new(crate::common::MOAI_BIN);
    c.args(args)
        .env("HOME", home)
        .env("MOAI_ACCOUNT_HOME", home)
        .env_remove("CLAUDE_CONFIG_DIR");
    if let Some(f) = folder {
        c.env("CLAUDE_CONFIG_DIR", f);
    }
    let out = c.output().unwrap();
    let v = serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    (out.status.success(), v)
}

/// AC-4 (REQ-5), D-5: a receipt belongs to the config folder it was applied in. Reverting
/// folder B's receipt from folder A's environment (the variable unset) changes nothing in either
/// folder, leaves the receipt usable, and names the folder it belongs to; run in folder B it then
/// reverts B and leaves A alone.
#[test]
fn a_receipt_reverts_only_in_the_config_folder_it_was_applied_in() {
    let (_t, home) = home();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    let second = home.join("profiles/b");
    let records = moai_cli::home_root::path(&home).join("home-style");
    let (_, id_a) = folder_install(&home, None);
    let (_, id_b) = folder_install(&home, Some(&second));
    let id_b = id_b.unwrap();
    assert!(id_a.is_some());
    let style_a = home.join(STYLE);
    let style_b = second.join("output-styles/MoAI-Style.md");
    let snapshot = |p: &Path| fs::read(p).ok();
    let before = (
        snapshot(&style_a),
        snapshot(&home.join(SETTINGS)),
        snapshot(&records.join("claude.json")),
        snapshot(&style_b),
        snapshot(&second.join("settings.json")),
        snapshot(&records.join(second_record_name(&second))),
    );
    let h = home.to_str().unwrap();
    let revert_b = [
        "home-style-revert",
        h,
        "--receipt",
        id_b.as_str(),
        "--permission-reference",
        PERM,
    ];
    let (ok, v) = folder_run_any(&home, None, &revert_b);
    // The one intended answer is a refusal (failed exit, `ok: false`), never a run
    // that reports some other state.
    assert!(!ok, "{v}");
    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(v.get("result").is_none(), "{v}");
    let text = v.to_string();
    assert!(
        text.contains(second.to_str().unwrap()),
        "the refusal names the receipt's folder: {v}"
    );
    let after = (
        snapshot(&style_a),
        snapshot(&home.join(SETTINGS)),
        snapshot(&records.join("claude.json")),
        snapshot(&style_b),
        snapshot(&second.join("settings.json")),
        snapshot(&records.join(second_record_name(&second))),
    );
    assert_eq!(after, before, "nothing changed in either folder");
    // In its own folder the same receipt still reverts B, and A stays.
    let v = folder_run(&home, Some(&second), &revert_b);
    assert_eq!(v["result"]["state"], "reverted", "{v}");
    assert!(!style_b.exists());
    assert_eq!(snapshot(&style_a), before.0);
    assert_eq!(snapshot(&records.join("claude.json")), before.2);
}

/// The names and sha256 of every file below `root`, sorted; empty when `root` is absent.
fn listing(root: &Path) -> Vec<(String, String)> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        let Ok(rd) = fs::read_dir(&dir) else { continue };
        for e in rd {
            let p = e.unwrap().path();
            if p.is_dir() {
                stack.push(p);
            } else {
                out.push((
                    p.strip_prefix(root).unwrap().display().to_string(),
                    moai_cli::digest(&fs::read(&p).unwrap()),
                ));
            }
        }
    }
    out.sort();
    out
}

/// AC-5 (REQ-6): a second folder's new record goes below this generation's state root; the
/// earlier state folder of the HOME is the same before and after.
#[test]
fn installing_into_a_second_config_folder_creates_nothing_under_the_earlier_state_folder() {
    let (_t, home) = home();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    // The earlier generation's HOME state folder, by its named constant.
    let earlier = home.join(moai_cli::HOME_STATE_DIR);
    fs::create_dir_all(earlier.join("keep")).unwrap();
    fs::write(earlier.join("keep/a.txt"), "a\n").unwrap();
    let before = listing(&earlier);
    let second = home.join("profiles/b");
    folder_install(&home, Some(&second));
    assert!(moai_cli::home_root::path(&home)
        .join("home-style")
        .join(second_record_name(&second))
        .is_file());
    assert_eq!(listing(&earlier), before);
    // And with no earlier folder at all, none appears.
    let (_t2, fresh) = project_with(&[]);
    assert!(moai_cli::home_root::adopt_new(&fresh, "test").unwrap());
    folder_install(&fresh, Some(&fresh.join("p")));
    assert!(!fresh.join(moai_cli::HOME_STATE_DIR).exists());
}

/// AC-6 (REQ-7): while the HOME's state root is not this generation's, a second folder is blocked
/// with a reason and nothing is written for it; the default folder plans as before.
#[test]
fn a_second_folder_is_blocked_while_the_state_root_is_not_this_generations() {
    let (_t, home) = home();
    // A HOME on the earlier generation's state folder (by its named constant).
    fs::create_dir_all(home.join(moai_cli::HOME_STATE_DIR)).unwrap();
    let second = home.join("profiles/b");
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    let in_b = folder_run(&home, Some(&second), &["home-style-plan", h, pkg]);
    let targets = in_b["result"]["targets"].as_array().unwrap();
    for t in targets {
        assert_eq!(t["action"], "blocked", "{t}");
        let d = t["detail"].as_str().unwrap();
        assert!(d.contains("env -u CLAUDE_CONFIG_DIR moai update"), "{d}");
    }
    let digest = in_b["result"]["plan_digest"].as_str().unwrap().to_string();
    let applied = folder_run(
        &home,
        Some(&second),
        &[
            "home-style-apply",
            h,
            pkg,
            "--plan-digest",
            &digest,
            "--permission-reference",
            PERM,
        ],
    );
    assert_eq!(applied["result"]["state"], "unchanged");
    assert!(
        !second.exists(),
        "no file or folder is made for the blocked folder"
    );
    let in_a = folder_run(&home, None, &["home-style-plan", h, pkg]);
    assert_eq!(folder_action(&in_a["result"], "style_file"), "create");
    assert_eq!(folder_action(&in_a["result"], "setting"), "set");
}

/// SPEC-20261127-update-linked-profile-display AC-3 (REQ-5, REQ-6): a second folder blocked on a
/// HOME of the earlier generation says which folder it is and both ways out in plain words, and a
/// block for another reason keeps its own reason and next step.
#[cfg(unix)]
#[test]
fn a_blocked_second_folder_names_the_way_out_without_internal_words() {
    let (_t, home) = home();
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    fs::create_dir_all(home.join(moai_cli::HOME_STATE_DIR)).unwrap();
    let second = home.join("profiles/b");
    let planned = folder_run(&home, Some(&second), &["home-style-plan", h, pkg]);
    let targets = planned["result"]["targets"].as_array().unwrap();
    assert!(!targets.is_empty(), "{planned}");
    let default_styles = home.join(".claude/output-styles");
    for t in targets {
        assert_eq!(t["action"], "blocked", "{t}");
        let d = t["detail"].as_str().unwrap();
        for inside in ["state root", "generation", "moved state folder"] {
            assert!(!d.contains(inside), "{inside}: {d}");
        }
        assert!(
            d.contains(second.to_str().unwrap()),
            "names the folder: {d}"
        );
        assert!(d.contains("next:"), "{d}");
        assert!(
            d.contains("env -u CLAUDE_CONFIG_DIR moai update"),
            "the way out without the variable: {d}"
        );
        assert!(
            d.contains(&format!(
                "link {} to {}",
                second.join("output-styles").display(),
                default_styles.display()
            )),
            "the way out by a link: {d}"
        );
    }

    // A link that the rules refuse keeps its own reason and next step.
    let (_t, home) = fresh();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    let h = home.to_str().unwrap();
    let profile = home.join("profiles/q");
    fs::create_dir_all(&profile).unwrap();
    std::os::unix::fs::symlink(
        home.join("outside/no-such-folder"),
        profile.join("output-styles"),
    )
    .unwrap();
    let planned = folder_run(&home, Some(&profile), &["home-style-plan", h, pkg]);
    let t = planned["result"]["targets"]
        .as_array()
        .unwrap()
        .iter()
        .find(|t| t["kind"] == "style_file")
        .unwrap();
    assert_eq!(t["action"], "blocked", "{planned}");
    let d = t["detail"].as_str().unwrap();
    assert!(d.contains("next:"), "{d}");
    assert!(
        !d.contains("env -u CLAUDE_CONFIG_DIR"),
        "a refused link is not a second folder: {d}"
    );
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

/// #2875 audit B1: the key goes in as text — every other byte (big integers, a trailing zero,
/// four-space indentation, order) stays — and a revert of an untouched file gives the before
/// bytes back byte for byte.
#[test]
fn settings_keep_every_byte_through_apply_and_revert() {
    for (before, after) in [
        (
            "{\n    \"a\": 1,\n    \"n\": 12345678901234567890123,\n    \"f\": 1.50\n}\n",
            "{\n    \"a\": 1,\n    \"n\": 12345678901234567890123,\n    \"f\": 1.50,\n    \"outputStyle\": \"MoAI-Style\"\n}\n",
        ),
        ("{\"a\":1.0}", "{\"a\":1.0, \"outputStyle\": \"MoAI-Style\"}"),
        ("{ }\n", "{\"outputStyle\": \"MoAI-Style\" }\n"),
    ] {
        let (_t, home) = home();
        write(&home, SETTINGS, before);
        let p = plan(&home, &package()).unwrap();
        assert_eq!(action(&p, "setting"), "set", "{before:?}");
        let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), after);
        let id = r["receipt"].as_str().unwrap().to_string();
        revert(&home, &id, PERM).unwrap();
        assert_eq!(
            fs::read(home.join(SETTINGS)).unwrap(),
            before.as_bytes(),
            "{before:?}"
        );
    }
}

/// #2875 audit B1: a file changed after the apply loses only the inserted text; when that text
/// is gone (the file was rewritten another way), the revert skips it and writes nothing.
#[test]
fn a_changed_settings_file_loses_only_the_inserted_text_or_is_skipped() {
    let before = "{\n    \"n\": 1.50\n}\n";
    let (_t2, other) = home();
    let (_t, home) = home();
    write(&home, SETTINGS, before);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let changed = fs::read_to_string(home.join(SETTINGS))
        .unwrap()
        .replace("{\n", "{\n    \"mine\": 10.0,\n");
    fs::write(home.join(SETTINGS), &changed).unwrap();
    revert(&home, &id, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\n    \"mine\": 10.0,\n    \"n\": 1.50\n}\n"
    );
    // Rewritten another way: the exact inserted text is not there, so nothing changes.
    write(&other, SETTINGS, before);
    let p = plan(&other, &package()).unwrap();
    let id = apply(&other, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let rewritten = "{\"n\":1.5,\"outputStyle\":\"MoAI-Style\"}";
    fs::write(other.join(SETTINGS), rewritten).unwrap();
    let r = revert(&other, &id, PERM).unwrap();
    assert_eq!(r["skipped"][0]["target"], "claude:setting");
    assert_eq!(fs::read_to_string(other.join(SETTINGS)).unwrap(), rewritten);
}

/// #2875 audit B2: the style file's frontmatter is plain YAML — every value that holds `: ` is
/// quoted, so a strict parser reads `name: MoAI-Style`.
#[test]
fn the_style_frontmatter_is_valid_yaml() {
    let out = render_style_file(&src());
    let front = out
        .strip_prefix("---\n")
        .and_then(|r| r.split_once("\n---\n"))
        .unwrap()
        .0;
    let mut keys = Vec::new();
    for line in front.lines() {
        let (key, value) = line.split_once(": ").unwrap();
        keys.push(key);
        if value.len() >= 2 && value.starts_with('"') && value.ends_with('"') {
            assert!(!value[1..value.len() - 1].contains('"'), "{line}");
        } else {
            assert!(!value.contains(": ") && !value.contains(" #"), "{line}");
        }
    }
    assert_eq!(keys, ["name", "description", "keep-coding-instructions"]);
    assert!(front.starts_with("name: MoAI-Style\n"));
}

/// #2875 audit B3: a `receipts` folder swapped for a link is never followed by a revert.
#[cfg(unix)]
#[test]
fn revert_refuses_a_linked_receipts_folder() {
    let (_o, outside) = home();
    let (_t, home) = home();
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let receipts = moai_cli::home_root::path(&home).join("home-style/receipts");
    fs::rename(&receipts, outside.join("receipts")).unwrap();
    std::os::unix::fs::symlink(outside.join("receipts"), &receipts).unwrap();
    let outside_receipt = outside.join("receipts").join(&id).join("receipt.json");
    let before = fs::read(&outside_receipt).unwrap();
    assert!(revert(&home, &id, PERM).unwrap_err().contains("symlink"));
    assert_eq!(fs::read(&outside_receipt).unwrap(), before);
    assert!(home.join(STYLE).exists());
}

// SPEC-20261120-reply-language-rule revision 2: besides `outputStyle`, the settings target plans
// the Claude Code `language` key (the saved conversation language's lower-case English name), in
// one file replacement with one receipt entry.

/// A new temporary HOME for the tests below, which shadow `home` with the path they get.
fn fresh() -> (tempfile::TempDir, PathBuf) {
    home()
}

/// The HOME's state folder and its saved conversation language, as the installer leaves them.
fn save_language(home: &Path, code: &str) {
    moai_cli::home_root::adopt_new(home, "test").unwrap();
    let r = moai_cli::language::set_home(home, code).unwrap();
    assert_eq!(r["written"], json!(true), "{r}");
}

/// A saved conversation language MoAI has no name for (`fr`): `language-set` refuses such a code,
/// so the record is written the way the installer's file holds it.
fn save_unnamed_language(home: &Path, code: &str) {
    moai_cli::home_root::adopt_new(home, "test").unwrap();
    let file = moai_cli::home_root::path(home)
        .join(moai_cli::language::HOME_FILE[0])
        .join(moai_cli::language::HOME_FILE[1]);
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(
        file,
        format!("language:\n  conversation_language: {code}\n"),
    )
    .unwrap();
}

/// The `language` value MoAI plans for a saved code: the lower-case English name of its table.
fn language_value(code: &str) -> Value {
    json!(moai_cli::language::english_name(code)
        .unwrap()
        .to_lowercase())
}

/// The keys the settings target of a plan writes, in order.
fn key_list(p: &Plan) -> Vec<(String, Value)> {
    p.targets
        .iter()
        .find(|t| t.kind == "setting")
        .unwrap()
        .keys
        .iter()
        .map(|k| (k.key.clone(), k.value.clone()))
        .collect()
}

fn pair(key: &str, value: Value) -> (String, Value) {
    (key.to_string(), value)
}

fn receipt_json(home: &Path, id: &str) -> Value {
    let path = moai_cli::home_root::path(home)
        .join("home-style/receipts")
        .join(id)
        .join("receipt.json");
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

/// AC-1 (REQ-1, REQ-4): a saved Korean conversation language plans `language = korean` beside
/// `outputStyle`, for no settings file, a file without the key, and a file where only the language
/// is missing (the style already ours, the user's own style, or a style file MoAI does not own).
#[test]
fn language_key_plan_sets_the_saved_conversation_language() {
    let name = language_value("ko");
    assert_eq!(name, json!("korean"));
    let (_t, home) = fresh();
    save_language(&home, "ko");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "set");
    assert_eq!(
        key_list(&p),
        [
            pair("outputStyle", json!(STYLE_NAME)),
            pair("language", name.clone())
        ]
    );
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    assert_eq!(
        settings(&home),
        json!({"outputStyle": STYLE_NAME, "language": "korean"})
    );
    let again = plan(&home, &package()).unwrap();
    assert_eq!(action(&again, "setting"), "unchanged");
    assert!(key_list(&again).is_empty());
    // A language MoAI wrote itself is not reported as "the user's value" on every later plan.
    let own_note = again
        .targets
        .iter()
        .find(|t| t.kind == "setting")
        .and_then(|t| t.detail.clone());
    assert_eq!(own_note, None, "{own_note:?}");

    // A settings file without the key keeps its own keys and order.
    let (_t, home) = fresh();
    save_language(&home, "ko");
    write(&home, SETTINGS, "{\"a\": 1}\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(
        key_list(&p),
        [
            pair("outputStyle", json!(STYLE_NAME)),
            pair("language", name.clone())
        ]
    );
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let s = settings(&home);
    let order: Vec<&String> = s.as_object().unwrap().keys().collect();
    assert_eq!(order, ["a", "outputStyle", "language"]);
    assert_eq!(s["language"], "korean");

    // Only the language is missing: the style is already the default.
    let (_t, home) = fresh();
    save_language(&home, "ko");
    write(&home, SETTINGS, "{\"outputStyle\": \"MoAI-Style\"}\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "set");
    assert_eq!(key_list(&p), [pair("language", name.clone())]);
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(
        settings(&home),
        json!({"outputStyle": "MoAI-Style", "language": "korean"})
    );

    // The user's own output style is kept and still says so; the language is filled.
    let (_t, home) = fresh();
    save_language(&home, "ko");
    let set = "{\"outputStyle\": \"Explanatory\"}\n";
    write(&home, SETTINGS, set);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "set");
    assert_eq!(key_list(&p), [pair("language", name.clone())]);
    assert!(detail(&p, "setting").contains("Explanatory"));
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(
        settings(&home),
        json!({"outputStyle": "Explanatory", "language": "korean"})
    );

    // A style file MoAI does not own keeps `outputStyle` out; the language is still filled.
    let (_t, home) = fresh();
    save_language(&home, "ko");
    let style = "---\nname: MoAI-Style\n---\n\nmy own style\n";
    write(&home, STYLE, style);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "conflict");
    assert_eq!(action(&p, "setting"), "set");
    assert_eq!(key_list(&p), [pair("language", name)]);
    assert!(detail(&p, "setting").contains("the style file is conflict"));
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), style);
    assert_eq!(settings(&home), json!({"language": "korean"}));

    // A settings file that is not an object stops both keys.
    let (_t, home) = fresh();
    save_language(&home, "ko");
    write(&home, SETTINGS, "[1, 2]\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "conflict");
    assert!(key_list(&p).is_empty());
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), "[1, 2]\n");
}

/// AC-2 (REQ-2): a `language` key that is already there is kept whatever it holds (another name,
/// the empty string, a number, `null`, the same name) and the plan's detail says so, except for
/// the name MoAI would write itself, which stays without a note.
#[test]
fn language_key_user_value_is_kept_and_said_so() {
    for (text, shown) in [
        ("\"japanese\"", "japanese"),
        ("\"\"", "\"\""),
        ("5", "5"),
        ("null", "null"),
        ("\"korean\"", "korean"),
    ] {
        let (_t, home) = fresh();
        save_language(&home, "ko");
        let before = format!("{{\"language\": {text}}}\n");
        write(&home, SETTINGS, &before);
        let p = plan(&home, &package()).unwrap();
        assert_eq!(action(&p, "setting"), "set", "{text}");
        assert_eq!(
            key_list(&p),
            [pair("outputStyle", json!(STYLE_NAME))],
            "{text}"
        );
        let d = detail(&p, "setting");
        if text == "\"korean\"" {
            // The value MoAI would write itself: kept as it is, and nothing said about it.
            assert!(!d.contains("kept"), "{d}");
        } else {
            assert!(
                d.contains("language") && d.contains(shown) && d.contains("kept"),
                "{d}"
            );
        }
        apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        // Every byte of the user's key stays; only `outputStyle` is added after it.
        assert_eq!(
            fs::read_to_string(home.join(SETTINGS)).unwrap(),
            format!("{{\"language\": {text}, \"outputStyle\": \"MoAI-Style\"}}\n"),
            "{text}"
        );
    }
    // Style already ours and the language the user's: nothing to write at all.
    let (_t, home) = fresh();
    save_language(&home, "ko");
    let both = "{\"outputStyle\": \"MoAI-Style\", \"language\": \"japanese\"}\n";
    write(&home, SETTINGS, both);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "unchanged");
    assert!(detail(&p, "setting").contains("japanese"));
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), both);
}

/// AC-3 (REQ-3): no saved conversation language, or a saved one MoAI has no name for (`fr`), plans
/// no `language` key and names the reason.
#[test]
fn language_key_unknown_language_names_the_reason() {
    let (_t, home) = fresh();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(key_list(&p), [pair("outputStyle", json!(STYLE_NAME))]);
    let d = detail(&p, "setting");
    assert!(
        d.contains("language") && d.contains("no conversation language saved"),
        "{d}"
    );
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));

    let (_t, home) = fresh();
    save_unnamed_language(&home, "fr");
    assert_eq!(
        moai_cli::language::home_conversation(&home).as_deref(),
        Some("fr")
    );
    assert!(moai_cli::language::english_name("fr").is_none());
    let p = plan(&home, &package()).unwrap();
    assert_eq!(key_list(&p), [pair("outputStyle", json!(STYLE_NAME))]);
    let d = detail(&p, "setting");
    assert!(
        d.contains("language") && d.contains("`fr`") && d.contains("not set"),
        "{d}"
    );
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(settings(&home), json!({"outputStyle": STYLE_NAME}));

    // Nothing else to write and no language to name: the target stays unchanged, with the reason.
    let (_t, home) = fresh();
    save_unnamed_language(&home, "fr");
    write(&home, SETTINGS, "{\"outputStyle\": \"MoAI-Style\"}\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "unchanged");
    assert!(key_list(&p).is_empty());
    assert!(detail(&p, "setting").contains("`fr`"));
}

/// The names of the reverted and skipped targets of a revert answer, for one target.
fn skipped_of(rev: &Value, target: &str) -> Vec<Value> {
    rev["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|s| s["target"] == target)
        .cloned()
        .collect()
}

fn reverted(rev: &Value, target: &str) -> bool {
    rev["reverted"]
        .as_array()
        .unwrap()
        .iter()
        .any(|w| w == target)
}

/// AC-4 (REQ-4, REQ-5, REQ-6, REQ-11): one apply writes both keys in one file replacement with one
/// receipt entry; a revert removes exactly the keys that still hold what was written, keeps the
/// user's other keys, leaves a key changed afterwards and says so, and a receipt without a key
/// list still reverts through its `inserted` text.
#[test]
fn language_key_one_apply_one_receipt_revert_removes_only_those() {
    let before = "{\n  \"theme\": \"dark\",\n  \"n\": 1.50\n}\n";
    let after = "{\n  \"theme\": \"dark\",\n  \"n\": 1.50,\n  \"outputStyle\": \"MoAI-Style\",\n  \"language\": \"korean\"\n}\n";
    // A HOME with `before` applied: its receipt id and the apply's answer.
    let applied = |code: Option<&str>| {
        let (t, home) = fresh();
        if let Some(code) = code {
            save_language(&home, code);
        }
        write(&home, SETTINGS, before);
        let p = plan(&home, &package()).unwrap();
        let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
        let id = r["receipt"].as_str().unwrap().to_string();
        (t, home, id, r)
    };

    // One write, one receipt entry: the entry's after hash is the final file with both keys.
    let (_t, home, id, r) = applied(Some("ko"));
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), after);
    let written = r["written"].as_array().unwrap();
    assert_eq!(written.iter().filter(|w| *w == "claude:setting").count(), 1);
    let receipt = receipt_json(&home, &id);
    let entries: Vec<&Value> = receipt["entries"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["kind"] == "setting")
        .collect();
    assert_eq!(entries.len(), 1, "{receipt}");
    let names: Vec<&str> = entries[0]["inserted_keys"]
        .as_array()
        .unwrap()
        .iter()
        .map(|k| k["key"].as_str().unwrap())
        .collect();
    assert_eq!(names, ["outputStyle", "language"]);
    assert_eq!(
        entries[0]["after_sha256"],
        json!(moai_cli::digest(after.as_bytes()))
    );

    // A file nobody touched gets its before bytes back byte for byte.
    let rev = revert(&home, &id, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), before);
    assert!(reverted(&rev, "claude:setting"));
    assert!(skipped_of(&rev, "claude:setting").is_empty(), "{rev}");

    // Another key added since: both keys still go, the new key stays, nothing is reported.
    let (_t, home, id, _) = applied(Some("ko"));
    fs::write(
        home.join(SETTINGS),
        after.replace("{\n", "{\n  \"b\": 2,\n"),
    )
    .unwrap();
    let rev = revert(&home, &id, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\n  \"b\": 2,\n  \"theme\": \"dark\",\n  \"n\": 1.50\n}\n"
    );
    assert!(skipped_of(&rev, "claude:setting").is_empty(), "{rev}");

    // The language changed since: it stays and is reported by name; `outputStyle` still goes.
    let (_t, home, id, _) = applied(Some("ko"));
    fs::write(
        home.join(SETTINGS),
        after.replace("\"korean\"", "\"english\""),
    )
    .unwrap();
    let rev = revert(&home, &id, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\n  \"theme\": \"dark\",\n  \"n\": 1.50,\n  \"language\": \"english\"\n}\n"
    );
    assert!(reverted(&rev, "claude:setting"));
    let skipped = skipped_of(&rev, "claude:setting");
    assert_eq!(skipped.len(), 1, "{rev}");
    assert_eq!(skipped[0]["reason"], "user_changed_since_apply");
    assert_eq!(skipped[0]["keys"], json!(["language"]));

    // The output style changed since: it stays, the language goes.
    let (_t, home, id, _) = applied(Some("ko"));
    fs::write(
        home.join(SETTINGS),
        after.replace("\"MoAI-Style\"", "\"Learning\""),
    )
    .unwrap();
    let rev = revert(&home, &id, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\n  \"theme\": \"dark\",\n  \"n\": 1.50,\n  \"outputStyle\": \"Learning\"\n}\n"
    );
    let skipped = skipped_of(&rev, "claude:setting");
    assert_eq!(skipped[0]["keys"], json!(["outputStyle"]));

    // Both changed since: nothing is written and the target is skipped as before.
    let (_t, home, id, _) = applied(Some("ko"));
    let both = after
        .replace("\"MoAI-Style\"", "\"Learning\"")
        .replace("\"korean\"", "\"english\"");
    fs::write(home.join(SETTINGS), &both).unwrap();
    let rev = revert(&home, &id, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(SETTINGS)).unwrap(), both);
    assert!(!reverted(&rev, "claude:setting"));
    let skipped = skipped_of(&rev, "claude:setting");
    assert_eq!(skipped.len(), 1, "{rev}");
    assert_eq!(skipped[0]["reason"], "user_changed_since_apply");

    // A receipt written before the key list (`outputStyle` alone, only `inserted`) still reverts.
    let (_t, home, id, _) = applied(None);
    let path = moai_cli::home_root::path(&home)
        .join("home-style/receipts")
        .join(&id)
        .join("receipt.json");
    let mut old = receipt_json(&home, &id);
    for e in old["entries"].as_array_mut().unwrap() {
        if e["kind"] == "setting" {
            assert!(e["inserted"].is_string(), "{e}");
            e.as_object_mut().unwrap().remove("inserted_keys");
        }
    }
    fs::write(&path, serde_json::to_vec_pretty(&old).unwrap()).unwrap();
    let now = fs::read_to_string(home.join(SETTINGS)).unwrap();
    fs::write(home.join(SETTINGS), now.replace("{\n", "{\n  \"b\": 2,\n")).unwrap();
    let rev = revert(&home, &id, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\n  \"b\": 2,\n  \"theme\": \"dark\",\n  \"n\": 1.50\n}\n"
    );
    assert!(reverted(&rev, "claude:setting"));
}

/// AC-5 (REQ-7): `CLAUDE_CONFIG_DIR` names the folder whose `settings.json` is filled, through the
/// binary; the temporary HOME's default Claude folder is not made or touched, a key the user holds
/// in that folder is kept, and the revert runs in that folder too.
#[test]
fn language_key_config_dir_profile_gets_its_own_settings() {
    let (_t, home) = fresh();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    save_language(&home, "ko");
    let second = home.join("profiles/b");
    let (planned, id) = folder_install(&home, Some(&second));
    let setting_of = |plan: &Value| -> Value {
        plan["targets"]
            .as_array()
            .unwrap()
            .iter()
            .find(|t| t["kind"] == "setting")
            .unwrap()
            .clone()
    };
    let setting = setting_of(&planned);
    assert_eq!(
        setting["path"],
        json!(second.join("settings.json").to_str().unwrap())
    );
    assert_eq!(
        setting["keys"],
        json!([
            {"key": "outputStyle", "value": STYLE_NAME},
            {"key": "language", "value": "korean"}
        ])
    );
    let written: Value =
        serde_json::from_slice(&fs::read(second.join("settings.json")).unwrap()).unwrap();
    assert_eq!(
        written,
        json!({"outputStyle": STYLE_NAME, "language": "korean"})
    );
    assert!(
        !home.join(".claude").exists(),
        "the default folder is not touched"
    );

    // A key the user holds in another profile's folder is kept.
    let third = home.join("profiles/c");
    fs::create_dir_all(&third).unwrap();
    fs::write(
        third.join("settings.json"),
        "{\"language\": \"japanese\"}\n",
    )
    .unwrap();
    let (planned, _) = folder_install(&home, Some(&third));
    assert_eq!(
        setting_of(&planned)["keys"],
        json!([{"key": "outputStyle", "value": STYLE_NAME}])
    );
    assert_eq!(
        fs::read_to_string(third.join("settings.json")).unwrap(),
        "{\"language\": \"japanese\", \"outputStyle\": \"MoAI-Style\"}\n"
    );
    assert!(!home.join(".claude").exists());

    // The revert runs where the apply ran and removes the file it created.
    let id = id.unwrap();
    let h = home.to_str().unwrap();
    let v = folder_run(
        &home,
        Some(&second),
        &[
            "home-style-revert",
            h,
            "--receipt",
            &id,
            "--permission-reference",
            PERM,
        ],
    );
    assert_eq!(v["result"]["state"], "reverted", "{v}");
    assert!(!second.join("settings.json").exists());
    assert!(!home.join(".claude").exists());
}

// SPEC-20261127-output-style-moai-style: the style is `MoAI-Style`; `moai update` moves an
// installed `MoAI-Easy.md` and an `outputStyle` of `MoAI-Easy` through the receipt.

/// The render of the text MoAI shipped under the earlier name (`tests/fixtures/home-style/render-8.md`,
/// the text before this SPEC): a file with these bytes is MoAI's own old style file.
fn old_render() -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/home-style/render-8.md"),
    )
    .unwrap()
}

fn record_path(home: &Path) -> PathBuf {
    moai_cli::home_root::path(home).join("home-style/claude.json")
}

/// The owned-text record as the installer writes it.
fn record_naming(text: &str) -> String {
    format!(
        "{{\n  \"schema_version\": 1,\n  \"owned_sha256\": \"{}\"\n}}\n",
        moai_cli::digest(text.as_bytes())
    )
}

fn kinds(p: &Plan) -> Vec<String> {
    p.targets.iter().map(|t| t.kind.clone()).collect()
}

/// A HOME that an earlier `moai update` installed: the old style file (its text named by the
/// record) and the given `settings.json`.
fn home_with_old(settings_text: &str) -> (tempfile::TempDir, PathBuf) {
    let (t, home) = home();
    write(&home, OLD_STYLE, &old_render());
    write_record(&home, &old_render());
    write(&home, SETTINGS, settings_text);
    (t, home)
}

fn write_record(home: &Path, text: &str) {
    let path = record_path(home);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, record_naming(text)).unwrap();
}

/// The settings text an earlier install left, with a value and spacing the move must not touch.
const OLD_SETTINGS: &str =
    "{\n  \"a\": 1,\n  \"outputStyle\": \"MoAI-Easy\",\n  \"b\": [1,  2]\n}\n";

/// AC-1 (REQ-1): a first install names the style file and the value `MoAI-Style`, writes no old
/// file and plans no old-file target.
#[test]
fn ac1_a_first_install_writes_the_new_names_and_no_old_file() {
    let (_t, home) = home();
    assert_eq!(STYLE_NAME, "MoAI-Style");
    assert_ne!(STYLE_NAME, OLD_NAME);
    let p = plan(&home, &package()).unwrap();
    let file = p.targets.iter().find(|t| t.kind == "style_file").unwrap();
    assert!(
        file.path.ends_with("output-styles/MoAI-Style.md"),
        "{}",
        file.path
    );
    assert_eq!(file.action, "create");
    assert_eq!(key_list(&p), [pair("outputStyle", json!("MoAI-Style"))]);
    assert_eq!(kinds(&p), ["style_file", "setting"]);
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    let text = fs::read_to_string(home.join(STYLE)).unwrap();
    assert!(text.starts_with("---\nname: MoAI-Style\n"), "{text}");
    assert_eq!(settings(&home), json!({"outputStyle": "MoAI-Style"}));
    assert!(!home.join(OLD_STYLE).exists(), "no old file is made");
}

/// AC-2 (REQ-2, REQ-3, REQ-4, REQ-10, REQ-14): an old file with MoAI's bytes is retired and the
/// exact old value is replaced as text; the apply writes the new file, then the setting, then
/// removes the old file, keeps the old bytes in the receipt and points the record at the new file.
#[test]
fn ac2_an_old_file_and_value_are_retired_and_replaced_as_text() {
    let (_t, home) = home_with_old(OLD_SETTINGS);
    let old = old_render();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(kinds(&p), ["style_file", "setting", "old_style_file"]);
    assert_eq!(action(&p, "style_file"), "create");
    assert_eq!(action(&p, "setting"), "replace");
    assert_eq!(action(&p, "old_style_file"), "retire");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "applied");
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        OLD_SETTINGS.replace("\"MoAI-Easy\"", "\"MoAI-Style\""),
        "one value text changes and every other byte stays"
    );
    assert!(!home.join(OLD_STYLE).exists());
    assert_eq!(
        fs::read_to_string(home.join(STYLE)).unwrap(),
        render_style_file(&src())
    );
    let id = r["receipt"].as_str().unwrap().to_string();
    let receipt = receipt_json(&home, &id);
    let entries = receipt["entries"].as_array().unwrap();
    let order: Vec<&str> = entries
        .iter()
        .map(|e| e["kind"].as_str().unwrap())
        .collect();
    assert_eq!(order, ["style_file", "setting", "old_style_file"]);
    let old_entry = &entries[2];
    assert_eq!(old_entry["action"], "retire");
    assert_eq!(
        old_entry["before_sha256"],
        json!(moai_cli::digest(old.as_bytes()))
    );
    let kept = moai_cli::home_root::path(&home)
        .join("home-style/receipts")
        .join(&id)
        .join("claude-old_style_file.before");
    assert_eq!(fs::read_to_string(kept).unwrap(), old);
    assert_eq!(
        fs::read_to_string(record_path(&home)).unwrap(),
        record_naming(&render_style_file(&src())),
        "the record names the new file"
    );
    // A HOME without an old file has no old-file target.
    let (_t2, clean) = fresh();
    assert!(!kinds(&plan(&clean, &package()).unwrap()).contains(&"old_style_file".to_string()));
}

/// AC-3 (REQ-11, REQ-12): the revert gives back the old file, the old value text, the settings
/// bytes and the record; a settings file edited elsewhere gets only the value text back; a new
/// file changed since is skipped by name.
#[test]
fn ac3_the_revert_gives_back_the_old_file_the_value_and_the_record() {
    let old = old_render();
    let (_t, home) = home_with_old(OLD_SETTINGS);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let r = revert(&home, &id, PERM).unwrap();
    assert_eq!(r["state"], "reverted");
    for target in [
        "claude:old_style_file",
        "claude:setting",
        "claude:style_file",
    ] {
        assert!(reverted(&r, target), "{target}: {r}");
    }
    assert_eq!(fs::read_to_string(home.join(OLD_STYLE)).unwrap(), old);
    assert!(!home.join(STYLE).exists());
    assert_eq!(
        fs::read(home.join(SETTINGS)).unwrap(),
        OLD_SETTINGS.as_bytes()
    );
    assert_eq!(
        fs::read_to_string(record_path(&home)).unwrap(),
        record_naming(&old),
        "the record names the old file again"
    );

    // The user edited the settings elsewhere after the apply: only the value text goes back.
    let (_t2, home) = home_with_old(OLD_SETTINGS);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let edited = fs::read_to_string(home.join(SETTINGS))
        .unwrap()
        .replace("\"b\": [1,  2]", "\"b\": [1,  2],\n  \"c\": 3");
    fs::write(home.join(SETTINGS), &edited).unwrap();
    revert(&home, &id, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        OLD_SETTINGS.replace("\"b\": [1,  2]", "\"b\": [1,  2],\n  \"c\": 3")
    );

    // The new style file changed since the apply: that target is skipped and named.
    let (_t3, home) = home_with_old(OLD_SETTINGS);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    let mine = fs::read_to_string(home.join(STYLE)).unwrap() + "mine\n";
    fs::write(home.join(STYLE), &mine).unwrap();
    let r = revert(&home, &id, PERM).unwrap();
    let skipped = skipped_of(&r, "claude:style_file");
    assert_eq!(skipped.len(), 1, "{r}");
    assert_eq!(skipped[0]["reason"], "user_changed_since_apply");
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), mine);
    assert_eq!(fs::read_to_string(home.join(OLD_STYLE)).unwrap(), old);
    assert_eq!(
        fs::read(home.join(SETTINGS)).unwrap(),
        OLD_SETTINGS.as_bytes()
    );
}

/// REQ-11, found by a guard mutant: the revert never writes over a file that is now at the
/// old path. The user's file stays as it is and the target is skipped as
/// `user_changed_since_apply`; the other targets still revert.
#[test]
fn revert_never_overwrites_a_file_put_at_the_old_style_path_since_the_apply() {
    let (_t, home) = home_with_old(OLD_SETTINGS);
    let p = plan(&home, &package()).unwrap();
    let id = apply(&home, &package(), &p.plan_digest, PERM).unwrap()["receipt"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(!home.join(OLD_STYLE).exists());
    fs::write(home.join(OLD_STYLE), "my own style\n").unwrap();
    let r = revert(&home, &id, PERM).unwrap();
    let skipped = skipped_of(&r, "claude:old_style_file");
    assert_eq!(skipped.len(), 1, "{r}");
    assert_eq!(skipped[0]["reason"], "user_changed_since_apply");
    assert_eq!(
        fs::read_to_string(home.join(OLD_STYLE)).unwrap(),
        "my own style\n"
    );
    assert!(reverted(&r, "claude:style_file"), "{r}");
}

/// AC-4 (REQ-5, REQ-6, REQ-7): an old file the user edited is a `conflict`; it and `outputStyle`
/// stay byte for byte, the new file is still written, and a notice names the old file and says
/// how to choose the style.
#[test]
fn ac4_an_old_file_the_user_edited_stays_with_its_value_and_a_notice_says_so() {
    let (_t, home) = home_with_old("{\"outputStyle\": \"MoAI-Easy\"}\n");
    let edited = format!("{}my own line\n", old_render());
    fs::write(home.join(OLD_STYLE), &edited).unwrap();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "old_style_file"), "conflict");
    assert_eq!(action(&p, "style_file"), "create");
    assert_ne!(action(&p, "setting"), "replace");
    let (_pt, project) = project_with(&[]);
    let said = reasons(&moai_cli::home_style::notices(&p, &project));
    assert!(
        said.iter()
            .any(|r| r.contains("MoAI-Easy.md") && r.contains("MoAI-Style")),
        "{said:?}"
    );
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(fs::read_to_string(home.join(OLD_STYLE)).unwrap(), edited);
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\"outputStyle\": \"MoAI-Easy\"}\n"
    );
    assert_eq!(
        fs::read_to_string(home.join(STYLE)).unwrap(),
        render_style_file(&src())
    );
}

/// AC-5 (REQ-4): with no old file the value `MoAI-Easy` is replaced, an absent value is added as
/// before, and any other value (a near miss in case, the user's own style) is a `conflict` left
/// as written; an own old file does not make another value replaceable.
#[test]
fn ac5_the_value_is_replaced_only_when_it_is_exactly_the_old_name() {
    let (_t, home) = home();
    write(&home, SETTINGS, "{\"outputStyle\": \"MoAI-Easy\"}\n");
    let p = plan(&home, &package()).unwrap();
    assert_eq!(kinds(&p), ["style_file", "setting"]);
    assert_eq!(action(&p, "setting"), "replace");
    apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\"outputStyle\": \"MoAI-Style\"}\n"
    );
    let (_t, home) = fresh();
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "setting"), "set");
    for other in ["moai-easy", "MOAI-EASY", "Explanatory"] {
        let text = format!("{{\"outputStyle\": \"{other}\"}}\n");
        // With no old file, and with MoAI's own old file next to the value.
        for old_file in [false, true] {
            let (_t, home) = fresh();
            write(&home, SETTINGS, &text);
            if old_file {
                write(&home, OLD_STYLE, &old_render());
                write_record(&home, &old_render());
            }
            let p = plan(&home, &package()).unwrap();
            assert_eq!(action(&p, "setting"), "conflict", "{other} {old_file}");
            apply(&home, &package(), &p.plan_digest, PERM).unwrap();
            assert_eq!(
                fs::read_to_string(home.join(SETTINGS)).unwrap(),
                text,
                "{other} {old_file}"
            );
        }
    }
}

/// AC-6 (REQ-13): a receipt written before this SPEC (entries without a path, an `outputStyle`
/// text recorded in `inserted`) reverts with the old path and the old value; the new file name
/// is never written or removed for it.
#[test]
fn ac6_an_old_format_receipt_reverts_with_the_old_path_and_the_old_value() {
    let (_t, home) = home();
    let old = old_render();
    write(&home, OLD_STYLE, &old);
    write(
        &home,
        SETTINGS,
        "{\"a\": 1, \"outputStyle\": \"MoAI-Easy\", \"b\": 2}\n",
    );
    let sha = |text: &str| moai_cli::digest(text.as_bytes());
    let dir = moai_cli::home_root::path(&home).join("home-style/receipts/style-1-0");
    fs::create_dir_all(&dir).unwrap();
    let receipt = json!({
        "schema_version": 1,
        "id": "style-1-0",
        "plan_digest": "sha256:fixture",
        "permission_reference": PERM,
        "status": "committed",
        "entries": [
            {
                "host": "claude", "kind": "style_file", "action": "create",
                "before_sha256": null, "after_sha256": sha(&old), "baseline_before": null
            },
            {
                "host": "claude", "kind": "setting", "action": "set",
                "before_sha256": sha("{\"a\": 1}\n"),
                "after_sha256": sha("{\"a\": 1, \"outputStyle\": \"MoAI-Easy\"}\n"),
                "baseline_before": null,
                "inserted": ", \"outputStyle\": \"MoAI-Easy\""
            }
        ]
    });
    fs::write(
        dir.join("receipt.json"),
        serde_json::to_vec_pretty(&receipt).unwrap(),
    )
    .unwrap();
    let r = revert(&home, "style-1-0", PERM).unwrap();
    assert_eq!(r["state"], "reverted", "{r}");
    assert!(reverted(&r, "claude:style_file"), "{r}");
    assert!(reverted(&r, "claude:setting"), "{r}");
    assert!(!home.join(OLD_STYLE).exists(), "the old file is removed");
    assert!(!home.join(STYLE).exists(), "the new name is never written");
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\"a\": 1, \"b\": 2}\n"
    );
}

/// AC-7 (REQ-8): a `MoAI-Style.md` that is not MoAI's is a `conflict`; the old file and the
/// value stay byte for byte and nothing is written.
#[test]
fn ac7_a_style_file_of_the_user_named_like_the_new_one_leaves_the_old_file_and_value() {
    let (_t, home) = home_with_old("{\"outputStyle\": \"MoAI-Easy\"}\n");
    let mine = "---\nname: MoAI-Style\n---\n\nmy own style\n";
    write(&home, STYLE, mine);
    let p = plan(&home, &package()).unwrap();
    assert_eq!(action(&p, "style_file"), "conflict");
    assert_eq!(action(&p, "old_style_file"), "skipped");
    assert_ne!(action(&p, "setting"), "replace");
    let r = apply(&home, &package(), &p.plan_digest, PERM).unwrap();
    assert_eq!(r["state"], "unchanged");
    assert_eq!(fs::read_to_string(home.join(STYLE)).unwrap(), mine);
    assert_eq!(
        fs::read_to_string(home.join(OLD_STYLE)).unwrap(),
        old_render()
    );
    assert_eq!(
        fs::read_to_string(home.join(SETTINGS)).unwrap(),
        "{\"outputStyle\": \"MoAI-Easy\"}\n"
    );
}

/// AC-8 (REQ-9): in a `CLAUDE_CONFIG_DIR` profile whose `output-styles` and `settings.json` are
/// links into the default folder the old file is moved by the same link rules; a link the rules
/// refuse (outside the default folder, pointing at nothing) blocks the old-file target with the
/// style file's reason and nothing is moved.
#[cfg(unix)]
#[test]
fn ac8_the_old_file_follows_the_link_rules_of_the_style_file() {
    use std::os::unix::fs::symlink;
    let is_link = |p: &Path| fs::symlink_metadata(p).unwrap().file_type().is_symlink();
    let h_of = |home: &Path| home.to_str().unwrap().to_string();
    let pkg = package();
    let pkg = pkg.to_str().unwrap().to_string();

    // Links into the default folder are followed for the old file too.
    let (_t, home) = home();
    assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
    let profile = home.join("profiles/p");
    write(&home, OLD_STYLE, &old_render());
    write(&home, SETTINGS, "{\"outputStyle\": \"MoAI-Easy\"}\n");
    fs::create_dir_all(&profile).unwrap();
    symlink(
        home.join(".claude/output-styles"),
        profile.join("output-styles"),
    )
    .unwrap();
    symlink(
        home.join(".claude/settings.json"),
        profile.join("settings.json"),
    )
    .unwrap();
    let (planned, id) = folder_install(&home, Some(&profile));
    assert_eq!(
        folder_action(&planned, "old_style_file"),
        "retire",
        "{planned}"
    );
    assert_eq!(folder_action(&planned, "setting"), "replace", "{planned}");
    assert!(id.is_some());
    assert!(!home.join(OLD_STYLE).exists());
    assert!(home.join(STYLE).is_file());
    assert_eq!(settings(&home), json!({"outputStyle": "MoAI-Style"}));
    assert!(is_link(&profile.join("output-styles")) && is_link(&profile.join("settings.json")));

    // A link the rules refuse blocks both file targets with the same reason; nothing moves.
    for pointing in ["styles-elsewhere", "no-such-folder"] {
        let (_t, home) = fresh();
        assert!(moai_cli::home_root::adopt_new(&home, "test").unwrap());
        let profile = home.join("profiles/q");
        let target = home.join("outside").join(pointing);
        if pointing == "styles-elsewhere" {
            write_at_path(&target.join("MoAI-Easy.md"), &old_render());
        }
        fs::create_dir_all(&profile).unwrap();
        symlink(&target, profile.join("output-styles")).unwrap();
        write_at_path(
            &profile.join("settings.json"),
            "{\"outputStyle\": \"MoAI-Easy\"}\n",
        );
        let h = h_of(&home);
        let planned = folder_run(&home, Some(&profile), &["home-style-plan", &h, &pkg]);
        let targets = planned["result"]["targets"].as_array().unwrap();
        let of = |kind: &str| targets.iter().find(|t| t["kind"] == kind).unwrap();
        assert_eq!(
            of("style_file")["action"],
            "blocked",
            "{pointing}: {planned}"
        );
        assert_eq!(
            of("old_style_file")["action"],
            "blocked",
            "{pointing}: {planned}"
        );
        assert_eq!(
            of("old_style_file")["detail"],
            of("style_file")["detail"],
            "{pointing}"
        );
        assert_ne!(of("setting")["action"], "replace", "{pointing}");
        let digest = planned["result"]["plan_digest"]
            .as_str()
            .unwrap()
            .to_string();
        let applied = folder_run(
            &home,
            Some(&profile),
            &[
                "home-style-apply",
                &h,
                &pkg,
                "--plan-digest",
                &digest,
                "--permission-reference",
                PERM,
            ],
        );
        assert_eq!(applied["result"]["state"], "unchanged", "{pointing}");
        assert_eq!(
            fs::read_to_string(profile.join("settings.json")).unwrap(),
            "{\"outputStyle\": \"MoAI-Easy\"}\n"
        );
        if pointing == "styles-elsewhere" {
            assert_eq!(
                fs::read_to_string(target.join("MoAI-Easy.md")).unwrap(),
                old_render()
            );
            assert!(!target.join("MoAI-Style.md").exists());
        }
    }
}

fn write_at_path(path: &Path, text: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, text).unwrap();
}
