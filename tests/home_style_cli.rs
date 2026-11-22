//! #2875 (#2871 I3): the MoAI-Easy output style in the user layer, against a temporary HOME
//! only — never the real one. Covers create, idempotence, a changed source, the user's own
//! style file and setting (conflicts, bytes kept), symlinks, a stale digest, and revert. The
//! Codex block is no longer written (SPEC-20261122-codex-home-update-removed).

mod common;
use moai_cli::home_style::{apply, plan, render_style_file, revert, Plan, SOURCE, STYLE_NAME};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};

const PERM: &str = "test-permission";
const STYLE: &str = ".claude/output-styles/MoAI-Easy.md";
const SETTINGS: &str = ".claude/settings.json";

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
    let style = "---\nname: MoAI-Easy\n---\n\nmy own style\n";
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
    assert!(!other.join("MoAI-Easy.md").exists());
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
        "{\"a\": 1, \"outputStyle\": \"MoAI-Easy\", \"b\": 2}\n",
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
    for wrong in ["moai-easy", "MOAI-EASY"] {
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
    write(&home, SETTINGS, "{\"outputStyle\": \"MoAI-Easy\"}\n");
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
            "{\"outputStyle\": \"MoAI-Easy\"}\n",
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
    let records = home.join(".moai/home-style");
    let (_, id_a) = folder_install(&home, None);
    let (_, id_b) = folder_install(&home, Some(&second));
    assert!(id_a.is_some() && id_b.is_some());
    assert!(
        records.join("claude.json").is_file(),
        "the default folder keeps its record name"
    );
    assert!(records.join(second_record_name(&second)).is_file());
    let style_b = second.join("output-styles/MoAI-Easy.md");
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
    fs::create_dir_all(home.join(".moai-cli/keep")).unwrap();
    fs::write(home.join(".moai-cli/keep/a.txt"), "a\n").unwrap();
    let before = listing(&home.join(".moai-cli"));
    let second = home.join("profiles/b");
    folder_install(&home, Some(&second));
    assert!(home
        .join(".moai/home-style")
        .join(second_record_name(&second))
        .is_file());
    assert_eq!(listing(&home.join(".moai-cli")), before);
    // And with no earlier folder at all, none appears.
    let (_t2, fresh) = project_with(&[]);
    assert!(moai_cli::home_root::adopt_new(&fresh, "test").unwrap());
    folder_install(&fresh, Some(&fresh.join("p")));
    assert!(!fresh.join(".moai-cli").exists());
}

/// AC-6 (REQ-7): while the HOME's state root is not this generation's, a second folder is blocked
/// with a reason and nothing is written for it; the default folder plans as before.
#[test]
fn a_second_folder_is_blocked_while_the_state_root_is_not_this_generations() {
    let (_t, home) = home();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    let second = home.join("profiles/b");
    let h = home.to_str().unwrap();
    let pkg = package();
    let pkg = pkg.to_str().unwrap();
    let in_b = folder_run(&home, Some(&second), &["home-style-plan", h, pkg]);
    let targets = in_b["result"]["targets"].as_array().unwrap();
    for t in targets {
        assert_eq!(t["action"], "blocked", "{t}");
        let d = t["detail"].as_str().unwrap();
        assert!(d.contains("state") && d.contains("generation"), "{d}");
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
            "{\n    \"a\": 1,\n    \"n\": 12345678901234567890123,\n    \"f\": 1.50,\n    \"outputStyle\": \"MoAI-Easy\"\n}\n",
        ),
        ("{\"a\":1.0}", "{\"a\":1.0, \"outputStyle\": \"MoAI-Easy\"}"),
        ("{ }\n", "{\"outputStyle\": \"MoAI-Easy\" }\n"),
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
    let rewritten = "{\"n\":1.5,\"outputStyle\":\"MoAI-Easy\"}";
    fs::write(other.join(SETTINGS), rewritten).unwrap();
    let r = revert(&other, &id, PERM).unwrap();
    assert_eq!(r["skipped"][0]["target"], "claude:setting");
    assert_eq!(fs::read_to_string(other.join(SETTINGS)).unwrap(), rewritten);
}

/// #2875 audit B2: the style file's frontmatter is plain YAML — every value that holds `: ` is
/// quoted, so a strict parser reads `name: MoAI-Easy`.
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
    assert!(front.starts_with("name: MoAI-Easy\n"));
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
    let receipts = home.join(".moai-cli/home-style/receipts");
    fs::rename(&receipts, outside.join("receipts")).unwrap();
    std::os::unix::fs::symlink(outside.join("receipts"), &receipts).unwrap();
    let outside_receipt = outside.join("receipts").join(&id).join("receipt.json");
    let before = fs::read(&outside_receipt).unwrap();
    assert!(revert(&home, &id, PERM).unwrap_err().contains("symlink"));
    assert_eq!(fs::read(&outside_receipt).unwrap(), before);
    assert!(home.join(STYLE).exists());
}
