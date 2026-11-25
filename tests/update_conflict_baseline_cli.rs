//! SPEC-20261125-update-conflict-baseline: `moai update` no longer leaves a skill file that is
//! already the package's (or an earlier generation's) as a false conflict, and a user's own edit
//! stays a conflict.
//!
//! Every case runs in a temporary HOME with copies of this build's binary; the real HOME, the
//! real Claude folder and both real state folders are never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

#[path = "update_adopt_cli/trees.rs"]
mod trees;
use trees::{copy_tree, names_in, tree};

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
    thread::sleep,
    time::Duration,
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// A file every package carries, edited per generation below.
const FILE: &str = "workflows/work.md";
/// The HOME state folder of the earlier name and the one of the new name, from the library, so
/// this file spells neither.
const EARLIER: &str = moai_cli::HOME_STATE_DIR;
const NEWER: &str = moai_cli::home_root::MOAI_DIR;

fn package() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
}

fn agent_sources() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("agents")
}

struct Home {
    _tmp: TempDir,
    home: PathBuf,
}

fn home() -> Home {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(moai_cli::home_root::path(&home)).unwrap();
    Home { _tmp: tmp, home }
}

/// A version folder `<state folder>/versions/<name>` with a copy of this build's binary, the
/// package and the agent sources; `edit` changes the package copy. Returns the binary.
fn install(home: &Path, name: &str, edit: impl FnOnce(&Path)) -> PathBuf {
    let folder = moai_cli::home_root::path(home).join("versions").join(name);
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    copy_tree(&package(), &folder.join("skills/moai"));
    copy_tree(&agent_sources(), &folder.join("agents"));
    edit(&folder.join("skills/moai"));
    exe
}

/// The package of a generation whose `FILE` ends with `tail`.
fn with_tail(tail: &'static str) -> impl FnOnce(&Path) {
    move |pkg| {
        let path = pkg.join(FILE);
        let mut text = fs::read_to_string(&path).unwrap();
        text.push_str(tail);
        fs::write(path, text).unwrap();
    }
}

fn run(exe: &Path, home: &Path, args: &[&str]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn result(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn skill_dir(home: &Path) -> PathBuf {
    home.join(".claude/skills/moai")
}

fn actions(preview: &Value) -> BTreeMap<String, String> {
    preview["plan"]["hosts"][0]["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn digest_of(preview: &Value) -> String {
    preview["plan_digest"].as_str().unwrap().to_string()
}

/// `update`, then `update --confirm` with its digest.
fn apply(exe: &Path, home: &Path) -> Value {
    let p = result(&run(exe, home, &["update"]));
    result(&run(exe, home, &["update", "--confirm", &digest_of(&p)]))
}

/// The owned digest of `FILE` in every host receipt under the state folder, newest last.
fn owned_digests(home: &Path) -> Vec<String> {
    let dir = moai_cli::home_root::path(home).join("installations");
    let mut receipts: Vec<Value> = Vec::new();
    for id in names_in(&dir) {
        if let Ok(bytes) = fs::read(dir.join(id).join("hosts/moai/claude.json")) {
            receipts.push(serde_json::from_slice(&bytes).unwrap());
        }
    }
    receipts.sort_by_key(|r| r["finished_at"].as_str().unwrap_or("").to_string());
    receipts
        .iter()
        .filter_map(|r| {
            r["owned"]
                .as_array()?
                .iter()
                .find(|f| f["path"] == json!(FILE))
                .and_then(|f| f["sha256"].as_str().map(str::to_string))
        })
        .collect()
}

fn sha_of(path: &Path) -> String {
    format!("sha256:{:x}", Sha256::digest(fs::read(path).unwrap()))
}

/// An older generation applied, the file then made current by other means (the baseline of the
/// receipt stays at the older bytes), and one more generation that changes the file again.
#[test]
fn ac1_a_package_equal_file_moves_its_baseline_and_the_next_package_is_an_update() {
    let h = home();
    let old = install(&h.home, "old", with_tail("\nold generation\n"));
    apply(&old, &h.home);
    let old_sha = owned_digests(&h.home)
        .pop()
        .expect("old receipt owns the file");
    // The file is now the current package's, written by another route; the receipt still owns
    // the old bytes.
    fs::copy(package().join(FILE), skill_dir(&h.home).join(FILE)).unwrap();
    // Receipts finish in whole seconds: the next one must finish in a later second.
    sleep(Duration::from_millis(1100));
    let current = install(&h.home, "a", |_| {});
    let skills_before = tree(&skill_dir(&h.home));

    let applied = apply(&current, &h.home);

    assert_ne!(applied["mode"], json!("refused"), "{applied}");
    assert_eq!(
        tree(&skill_dir(&h.home)),
        skills_before,
        "no skill byte was rewritten"
    );
    let owned = owned_digests(&h.home);
    assert_ne!(
        owned.last().unwrap(),
        &old_sha,
        "the baseline moved: {owned:?}"
    );
    assert_eq!(
        owned.last().unwrap(),
        &sha_of(&package().join(FILE)),
        "the receipt owns the package's bytes"
    );

    // A later generation changes the file again: that is an update, not a conflict.
    sleep(Duration::from_millis(1100));
    let later = install(&h.home, "b", with_tail("\nlater generation\n"));
    let preview = result(&run(&later, &h.home, &["update"]));
    assert_eq!(actions(&preview)[FILE], "update", "{preview}");
}

/// The disk holds the file of an earlier generation (not the baseline, not the package): it is
/// our own copy, so it is an update.
#[test]
fn ac2_an_earlier_generation_copy_of_an_owned_file_is_an_update_not_a_conflict() {
    let h = home();
    let old = install(&h.home, "old", with_tail("\nold generation\n"));
    apply(&old, &h.home);
    let mid = install(&h.home, "mid", with_tail("\nmiddle generation\n"));
    fs::copy(
        mid.parent()
            .unwrap()
            .parent()
            .unwrap()
            .join("skills/moai")
            .join(FILE),
        skill_dir(&h.home).join(FILE),
    )
    .unwrap();
    sleep(Duration::from_millis(1100));
    let current = install(&h.home, "a", |_| {});

    let preview = result(&run(&current, &h.home, &["update"]));

    assert_eq!(actions(&preview)[FILE], "update", "{preview}");
    result(&run(
        &current,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert_eq!(
        sha_of(&skill_dir(&h.home).join(FILE)),
        sha_of(&package().join(FILE)),
        "the file is now the package's"
    );
}

/// The user's own edit matches no generation: it stays a conflict and its bytes are kept.
#[test]
fn ac3_a_users_own_edit_stays_a_conflict_and_is_kept() {
    let h = home();
    let old = install(&h.home, "old", with_tail("\nold generation\n"));
    apply(&old, &h.home);
    let edited = skill_dir(&h.home).join(FILE);
    fs::write(&edited, "the user's own words\n").unwrap();
    sleep(Duration::from_millis(1100));
    let current = install(&h.home, "a", |_| {});

    let preview = result(&run(&current, &h.home, &["update"]));

    assert_eq!(actions(&preview)[FILE], "conflict", "{preview}");
    result(&run(
        &current,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert_eq!(
        fs::read_to_string(&edited).unwrap(),
        "the user's own words\n",
        "the user's bytes are kept"
    );
}

// ---- the human screens ----------------------------------------------------------------------

use moai_cli::ui::Painter;
use moai_cli::ui_text::Lang;
use moai_cli::update_tux::{screen, Asked, View};

fn fixture(name: &str) -> Value {
    let p = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join(format!("tests/fixtures/update_tux/{name}.json"));
    serde_json::from_str(&fs::read_to_string(p).unwrap()).unwrap()
}

fn view(lang: Lang) -> View {
    View {
        painter: Painter {
            color: false,
            unicode: true,
            lang,
        },
        width: 120,
        home: "/home/u".into(),
        project: None,
    }
}

fn conflict_file(path: &str) -> Value {
    json!({
        "path": path, "action": "conflict", "before_sha256": null,
        "after_sha256": null, "published": false, "error": null
    })
}

/// Files of an applied host that stayed in conflict are named, with their number, although other
/// files of the host were written.
#[test]
fn ac4_an_applied_host_names_the_files_that_stayed_in_conflict() {
    let mut v = fixture("applied");
    assert_eq!(v["result"]["hosts"][0]["status"], json!("applied"));
    let files = v["result"]["hosts"][0]["files"].as_array_mut().unwrap();
    files.push(conflict_file("workflows/work.md"));
    files.push(conflict_file("references/flows/next-flow.md"));

    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&v, Asked::Yes, &view(lang));
        assert!(out.contains("workflows/work.md"), "{lang:?}: {out}");
        assert!(
            out.contains("references/flows/next-flow.md"),
            "{lang:?}: {out}"
        );
        let word = if lang == Lang::Ko {
            "충돌"
        } else {
            "conflict"
        };
        assert!(out.contains(word), "{lang:?}: {out}");
    }

    let clean = screen(&fixture("applied"), Asked::Yes, &view(Lang::En));
    assert!(!clean.contains("workflows/work.md"), "{clean}");
    assert!(!clean.contains("conflict"), "no conflict, no line: {clean}");
}

/// A linked routing file reads "linked", and the flow rule's detail (the earlier build's file
/// left in place) is a line of the preview.
#[test]
fn ac5_a_linked_routing_file_reads_linked_and_the_flow_detail_is_shown() {
    let mut v = fixture("preview-changes");
    let note = "an earlier build wrote /home/u/.claude/rules/moai/flow.md; it is left in place for you to delete";
    v["routing"]["hosts"] = json!([
        {
            "host": "claude", "config_dir": "/home/u/.claude",
            "file": "/home/u/.claude/rules/moai.md", "preflight": "ok", "blocked_reason": null,
            "section": {"path": "/home/u/.claude/rules/moai.md", "action": "unchanged", "detail": null},
            "flow": {"path": "/home/u/.claude/rules/moai-flow.md", "action": "unchanged", "detail": note}
        },
        {
            "host": "claude", "config_dir": "/home/u/.orca",
            "file": "/home/u/.orca/rules/moai.md", "preflight": "ok", "blocked_reason": null,
            "section": {"path": "/home/u/.orca/rules/moai.md", "action": "linked", "detail": "reached through a link"},
            "flow": null
        }
    ]);

    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&v, Asked::NotAsked, &view(lang));
        assert!(
            out.contains("left in place for you to delete"),
            "{lang:?}: {out}"
        );
        assert!(out.contains("moai/flow.md"), "{lang:?}: {out}");
    }
    let out = screen(&v, Asked::NotAsked, &view(Lang::En));
    assert!(out.contains("moai.md linked"), "{out}");
    assert!(!out.contains("moai.md skipped"), "{out}");
}

/// The Applied screen shows the next command of a switch that was not confirmed, and the revert
/// lines name the real home and say what the permission reference is.
#[test]
fn ac6_the_applied_screen_shows_the_switch_command_and_a_readable_revert_line() {
    let mut v = fixture("applied");
    v["switch"] = json!({"apply_flag": "--confirm-switch switch:0123456789abcdef"});
    assert!(v["revert"]["agents"].as_str().unwrap().contains("<HOME>"));

    let out = screen(&v, Asked::Yes, &view(Lang::En));
    assert!(
        out.contains("moai update --confirm-switch switch:0123456789abcdef"),
        "{out}"
    );
    assert!(
        out.contains("moai agents-home-revert /home/u --receipt"),
        "{out}"
    );
    assert!(!out.contains("<HOME>"), "{out}");
    assert!(out.contains("who allowed the revert"), "{out}");

    let ko = screen(&v, Asked::Yes, &view(Lang::Ko));
    assert!(ko.contains("허락한 근거"), "{ko}");
    assert!(!ko.contains("<HOME>"), "{ko}");

    // A switch that was confirmed in this run shows no next command.
    v["switch_result"] =
        json!({"state": "applied", "saved": "/home/u/saved", "revert": "moai switch-revert"});
    let done = screen(&v, Asked::Yes, &view(Lang::En));
    assert!(!done.contains("moai update --confirm-switch"), "{done}");
}

// ---- revision 2: replace the conflicts after a backup, remove the earlier flow rule ------------

/// `run` with extra environment variables.
fn run_env(exe: &Path, home: &Path, args: &[&str], env: &[(&str, &str)]) -> Output {
    Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .envs(env.iter().copied())
        .current_dir(home)
        .stdin(Stdio::null())
        .output()
        .unwrap()
}

/// An old generation applied, the user's own words in `FILE`, then the current package: the
/// home, the binary that made the state and the preview of it.
fn user_edited_state() -> (Home, PathBuf, Value) {
    user_edited_state_in(home())
}

/// `user_edited_state` in a HOME the caller made.
fn user_edited_state_in(h: Home) -> (Home, PathBuf, Value) {
    let old = install(&h.home, "old", with_tail("\nold generation\n"));
    apply(&old, &h.home);
    fs::write(skill_dir(&h.home).join(FILE), "the user's own words\n").unwrap();
    sleep(Duration::from_millis(1100));
    let current = install(&h.home, "a", |_| {});
    let preview = result(&run(&current, &h.home, &["update"]));
    assert_eq!(actions(&preview)[FILE], "conflict", "{preview}");
    (h, current, preview)
}

fn backup_dirs(home: &Path, suffix: &str) -> Vec<PathBuf> {
    let root = moai_cli::home_root::path(home).join("backups");
    names_in(&root)
        .into_iter()
        .filter(|n| n.ends_with(suffix))
        .map(|n| root.join(n))
        .collect()
}

/// A conflicting host offers the replacement with the files, the backup folder and the flag; the
/// human preview shows the command; no conflict, no offer and no line.
#[test]
fn ac8_a_conflict_offers_the_replacement_and_the_screen_shows_its_command() {
    let (_h, _exe, preview) = user_edited_state();
    let offer = &preview["plan"]["hosts"][0]["conflict_replace"];
    let files: Vec<&str> = offer["files"]
        .as_array()
        .unwrap_or_else(|| panic!("no offer: {preview}"))
        .iter()
        .map(|f| f["path"].as_str().unwrap())
        .collect();
    assert_eq!(files, vec![FILE], "{offer}");
    assert!(
        offer["backup_root"].as_str().unwrap().contains("backups"),
        "{offer}"
    );
    let flag = offer["apply_flag"].as_str().unwrap();
    assert!(
        flag.starts_with("--confirm-conflict-replace conflict-replace:claude:"),
        "{flag}"
    );
    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&preview, Asked::NotAsked, &view(lang));
        assert!(
            out.contains(&format!(
                "moai update --confirm {} {flag}",
                digest_of(&preview)
            )),
            "{lang:?}: {out}"
        );
    }

    let clean = fixture("preview-changes");
    assert!(
        clean["plan"]["hosts"][0]["conflict_replace"].is_null(),
        "the fixture has no conflict"
    );
    let out = screen(&clean, Asked::NotAsked, &view(Lang::En));
    assert!(!out.contains("--confirm-conflict-replace"), "{out}");
}

/// Without the flag a confirmed update writes nothing to the conflict; with it the original is
/// copied first and the package bytes replace it; a failed copy replaces nothing.
#[test]
fn ac9_the_flag_backs_the_conflict_up_and_replaces_it_and_nothing_without_it() {
    // A confirmed run applies the rest of the plan too, so each case starts from its own state.
    let state = || {
        let (h, exe, preview) = user_edited_state();
        let digest = digest_of(&preview);
        let token = preview["plan"]["hosts"][0]["conflict_replace"]["confirm_token"]
            .as_str()
            .unwrap_or_else(|| panic!("no offer: {preview}"))
            .to_string();
        let target = skill_dir(&h.home).join(FILE);
        (h, exe, digest, token, target)
    };
    let words = "the user's own words\n";

    // No flag: the user's bytes stay, no backup folder is made.
    let (h, exe, digest, _, target) = state();
    result(&run(&exe, &h.home, &["update", "--confirm", &digest]));
    assert_eq!(fs::read_to_string(&target).unwrap(), words);
    assert!(backup_dirs(&h.home, "-claude-conflict").is_empty());

    // A failed copy: nothing is replaced and no backup folder is left half made.
    let (h, exe, digest, token, target) = state();
    let args = [
        "update",
        "--confirm",
        digest.as_str(),
        "--confirm-conflict-replace",
        token.as_str(),
    ];
    run_env(
        &exe,
        &h.home,
        &args,
        &[("MOAI_TEST_FAULT", "conflict_backup_copy")],
    );
    assert_eq!(
        fs::read_to_string(&target).unwrap(),
        words,
        "a failed backup replaces nothing"
    );
    assert!(backup_dirs(&h.home, "-claude-conflict").is_empty());

    // With the flag: the copy first, then the package's bytes, and the receipt owns them.
    let (h, exe, digest, token, target) = state();
    let args = [
        "update",
        "--confirm",
        digest.as_str(),
        "--confirm-conflict-replace",
        token.as_str(),
    ];
    let applied = result(&run(&exe, &h.home, &args));
    assert_ne!(applied["mode"], json!("refused"), "{applied}");
    assert_eq!(sha_of(&target), sha_of(&package().join(FILE)));
    let dirs = backup_dirs(&h.home, "-claude-conflict");
    assert_eq!(dirs.len(), 1, "{dirs:?}");
    let copy = fs::read_to_string(dirs[0].join("files").join(FILE)).unwrap();
    assert_eq!(copy, "the user's own words\n");
    assert_eq!(
        owned_digests(&h.home).last().unwrap(),
        &sha_of(&package().join(FILE))
    );
}

fn write_flow_receipt(home: &Path, bytes: &[u8]) {
    let dir = moai_cli::home_root::path(home).join("home-routing/receipts/routing-1");
    fs::create_dir_all(&dir).unwrap();
    let after = format!("sha256:{:x}", Sha256::digest(bytes));
    let receipt = json!({
        "schema_version": 1, "id": "routing-1", "plan_digest": "sha256:0",
        "permission_reference": "test", "status": "committed",
        "entries": [{
            "host": "claude", "kind": "flow",
            "config_dir": home.join(".claude").to_string_lossy(),
            "action": "create", "before_sha256": null, "after_sha256": after,
            "baseline_before": null, "created_dirs": []
        }]
    });
    fs::write(
        dir.join("receipt.json"),
        serde_json::to_vec(&receipt).unwrap(),
    )
    .unwrap();
}

fn write_old_flow_rule(home: &Path, bytes: &[u8]) -> PathBuf {
    let file = home.join(".claude/rules/moai/flow.md");
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    fs::write(&file, bytes).unwrap();
    file
}

/// The earlier flow rule file is removed after a backup only when its digest is one an earlier
/// routing receipt recorded; any other bytes stay and are only named.
#[test]
fn ac10_the_earlier_flow_rule_is_removed_after_a_backup_only_when_it_is_ours() {
    // Ours: the bytes an earlier receipt wrote.
    let h = home();
    let exe = install(&h.home, "a", |_| {});
    let file = write_old_flow_rule(&h.home, b"flow rule of an earlier build\n");
    write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
    let preview = result(&run(&exe, &h.home, &["update"]));
    let plan = &preview["routing"]["hosts"][0]["legacy_flow"];
    assert_eq!(plan["action"], json!("remove"), "{preview}");
    result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert!(!file.exists(), "the file is removed");
    let dirs = backup_dirs(&h.home, "-claude-flow-rule");
    assert_eq!(dirs.len(), 1, "{dirs:?}");
    assert_eq!(
        fs::read_to_string(dirs[0].join("rules/moai/flow.md")).unwrap(),
        "flow rule of an earlier build\n"
    );
    let applied = fixture("applied");
    let mut shown = applied.clone();
    shown["routing_result"] = json!({
        "state": "applied",
        "legacy_flow_removed": [{"path": file.to_string_lossy(), "backup": dirs[0].to_string_lossy()}]
    });
    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&shown, Asked::Yes, &view(lang));
        assert!(out.contains("-claude-flow-rule"), "{lang:?}: {out}");
    }
    assert!(!screen(&applied, Asked::Yes, &view(Lang::En)).contains("-claude-flow-rule"));

    // A failed backup: the file stays and no backup folder is left.
    let h = home();
    let exe = install(&h.home, "a", |_| {});
    let file = write_old_flow_rule(&h.home, b"flow rule of an earlier build\n");
    write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
    let preview = result(&run(&exe, &h.home, &["update"]));
    run_env(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
        &[("MOAI_TEST_FAULT", "flow_rule_backup")],
    );
    assert_eq!(
        fs::read_to_string(&file).unwrap(),
        "flow rule of an earlier build\n",
        "a failed backup removes nothing"
    );
    assert!(backup_dirs(&h.home, "-claude-flow-rule").is_empty());

    // Not ours: the user's own words in the same place stay, and the preview only names them.
    let h = home();
    let exe = install(&h.home, "a", |_| {});
    let file = write_old_flow_rule(&h.home, b"the user's own rule\n");
    write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
    let preview = result(&run(&exe, &h.home, &["update"]));
    assert!(
        preview["routing"]["hosts"][0]["legacy_flow"].is_null(),
        "{preview}"
    );
    result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert_eq!(fs::read_to_string(&file).unwrap(), "the user's own rule\n");
    assert!(backup_dirs(&h.home, "-claude-flow-rule").is_empty());
}

// ---- SPEC-20261125-update-screen-facts: the update screen and its model step tell the facts ----

fn applied_with(parts: Value) -> Value {
    let mut v = json!({
        "mode": "applied",
        "applied": true,
        "result": {"status": "applied", "hosts": [{
            "host": "claude", "status": "applied",
            "files": [{"path": "SKILL.md", "action": "unchanged", "published": false}]
        }]},
        "agents_result": {"state": "unchanged"},
        "routing_result": {"state": "unchanged"},
        "style_result": {"state": "unchanged"},
        "bin_names_result": {"state": "unchanged"},
        "hooks_result": {"state": "unchanged"}
    });
    for (k, p) in parts.as_object().unwrap() {
        v[k] = p.clone();
    }
    v
}

/// A run that only adopted files or moved receipt baselines changed nothing a person can see, so
/// it does not open the model questions on its own; a written file or a written part does.
#[test]
fn screen_facts_ac1_only_a_visible_change_opens_the_model_step() {
    use moai_cli::update::{model_step_opens, visible_change};
    assert!(!visible_change(&applied_with(json!({}))), "adopt only");
    let one_file = json!({"result": {"status": "applied", "hosts": [{
        "host": "claude", "status": "applied",
        "files": [{"path": "SKILL.md", "action": "update", "published": true}]
    }]}});
    assert!(visible_change(&applied_with(one_file)), "a skill file");
    for (key, state) in [
        ("agents_result", json!({"state": "applied"})),
        (
            "agents_result",
            json!({"state": "partial", "written": ["a"]}),
        ),
        ("routing_result", json!({"state": "applied"})),
        ("style_result", json!({"state": "applied"})),
        ("bin_names_result", json!({"state": "applied"})),
        ("hooks_result", json!({"state": "partial"})),
        ("extra_skills_result", json!({"state": "applied"})),
        ("keys_result", json!({"state": "applied"})),
    ] {
        let v = applied_with(json!({ key: state.clone() }));
        assert!(visible_change(&v), "{key} {state}");
    }
    let nothing = applied_with(json!({"agents_result": {"state": "partial", "written": []}}));
    assert!(!visible_change(&nothing), "a partial that wrote nothing");

    // scripted, terminal, wrote, models saved
    assert!(
        !model_step_opens(false, true, false, true),
        "saved, nothing seen"
    );
    assert!(model_step_opens(false, true, false, false), "nothing saved");
    assert!(
        model_step_opens(false, true, true, true),
        "a visible change"
    );
    assert!(!model_step_opens(true, true, true, false), "a scripted run");
}

/// Every end of the model step has one line on both screens; a report that asked nothing has none.
#[test]
fn screen_facts_ac2_the_model_step_result_has_one_line() {
    let path = moai_cli::update_models::settings_path(Path::new("/home/u"));
    let tail = path.file_name().unwrap().to_string_lossy().into_owned();
    let cases = [
        (
            json!({"state": "applied"}),
            "Model settings saved:",
            "모델 설정을 저장했어요",
        ),
        (
            json!({"state": "unchanged"}),
            "Model settings left as they were",
            "모델 설정은 그대로예요",
        ),
        (
            json!({"state": "nothing_picked"}),
            "Model settings left as they were",
            "모델 설정은 그대로예요",
        ),
        (
            json!({"state": "declined"}),
            "Model settings not saved",
            "모델 설정을 저장하지 않았어요",
        ),
        (
            json!({"state": "cancelled"}),
            "Model questions cancelled",
            "모델 질문을 취소했어요",
        ),
        (
            json!({"state": "error", "message": "the file broke"}),
            "Model settings problem: the file broke",
            "모델 설정 문제: the file broke",
        ),
    ];
    for (models, en, ko) in &cases {
        let mut v = fixture("applied");
        v["models"] = models.clone();
        let out = screen(&v, Asked::Yes, &view(Lang::En));
        assert!(out.contains(en), "{models}: {out}");
        assert_eq!(out.matches("Model ").count(), 1, "one line: {out}");
        let out = screen(&v, Asked::Yes, &view(Lang::Ko));
        assert!(out.contains(ko), "{models}: {out}");
    }
    let mut saved = fixture("applied");
    saved["models"] = json!({"state": "applied"});
    for lang in [Lang::En, Lang::Ko] {
        let out = screen(&saved, Asked::Yes, &view(lang));
        assert!(
            out.contains(&tail),
            "the saved line names the settings file: {out}"
        );
    }
    let mut report = fixture("applied");
    report["models"] = json!({"state": "ok", "values": []});
    let out = screen(&report, Asked::Yes, &view(Lang::En));
    assert!(
        !out.contains("Model settings") && !out.contains("Model questions"),
        "{out}"
    );
    // The nothing-to-apply screen of a person's terminal shows it as well.
    let mut latest = fixture("preview-latest");
    latest["models"] = json!({"state": "declined"});
    let out = screen(&latest, Asked::NotAsked, &view(Lang::En));
    assert!(out.contains("Model settings not saved"), "{out}");
}

/// The Checking row and the Applied row of a host that adopted files say how many; a host that
/// adopted none has neither the words nor the key.
#[test]
fn screen_facts_ac3_the_adopted_files_are_counted() {
    let h = home();
    let exe = install(&h.home, "a", |_| {});
    copy_tree(&package(), &skill_dir(&h.home));
    let preview = result(&run(&exe, &h.home, &["update"]));
    let n = preview["plan"]["hosts"][0]["adopt"]["files"]
        .as_array()
        .unwrap_or_else(|| panic!("no adopt list: {preview}"))
        .len();
    assert!(n > 0);
    let out = screen(&preview, Asked::NotAsked, &view(Lang::En));
    assert!(out.contains(&format!("{n} to adopt")), "{out}");
    let out = screen(&preview, Asked::NotAsked, &view(Lang::Ko));
    assert!(
        out.contains("기록 예정") && out.contains(&format!("{n}개")),
        "{out}"
    );

    let applied = result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert_eq!(
        applied["result"]["hosts"][0]["adopted"],
        json!(n),
        "{applied}"
    );
    let out = screen(&applied, Asked::Yes, &view(Lang::En));
    assert!(out.contains(&format!("{n} adopted")), "{out}");
    let out = screen(&applied, Asked::Yes, &view(Lang::Ko));
    assert!(
        out.contains("기록함") && out.contains(&format!("{n}개")),
        "{out}"
    );

    let plain = fixture("applied");
    assert!(plain["result"]["hosts"][0]["adopted"].is_null());
    assert!(!screen(&plain, Asked::Yes, &view(Lang::En)).contains("adopted"));
    let (_h2, _exe2, preview2) = user_edited_state();
    assert!(
        preview2["plan"]["hosts"][0]["adopt"].is_null(),
        "no adopt here: {preview2}"
    );
    assert!(!screen(&preview2, Asked::NotAsked, &view(Lang::En)).contains("to adopt"));
}

/// A plan with nothing to write never says everything is up to date above an open conflict, a
/// skipped line or a blocked host; it says nothing is left to apply, that the rest is up to date
/// and which rows are still open. With no open row the screen keeps its single line.
#[test]
fn screen_facts_ac4_the_headline_follows_the_rows() {
    let conflicts = || {
        let mut v = fixture("preview-latest");
        let files = v["plan"]["hosts"][0]["files"].as_array_mut().unwrap();
        files[0]["action"] = json!("conflict");
        files[1]["action"] = json!("conflict");
        v
    };
    let skipped = || {
        let mut v = fixture("preview-latest");
        v["routing"]["hosts"][0]["section"]["action"] = json!("held");
        v
    };
    for (name, v, open) in [
        ("conflict", conflicts(), "skills claude"),
        ("skipped", skipped(), "routing"),
        ("blocked", fixture("preview-blocked"), "skills claude"),
    ] {
        let out = screen(&v, Asked::NotAsked, &view(Lang::En));
        assert!(!out.contains("Everything is up to date"), "{name}: {out}");
        assert!(
            !out.contains("clear the blocked part first"),
            "{name}: {out}"
        );
        assert!(out.contains("Nothing left to apply"), "{name}: {out}");
        assert!(out.contains("the rest is up to date"), "{name}: {out}");
        let line = out
            .lines()
            .find(|l| l.contains("Still open:"))
            .unwrap_or_else(|| panic!("{name}: {out}"));
        assert!(line.contains(open), "{name}: {line}");
        let ko = screen(&v, Asked::NotAsked, &view(Lang::Ko));
        assert!(
            !ko.contains("모두 최신이에요") && ko.contains("아직 열린 줄"),
            "{name}: {ko}"
        );
    }
    let out = screen(&fixture("preview-latest"), Asked::NotAsked, &view(Lang::En));
    assert!(
        out.contains("Everything is up to date") && !out.contains("Still open"),
        "{out}"
    );
}

// ---- revision 3: backups go under the new state folder; the emptied rules folder is removed ----

/// A HOME whose state folder resolves to the earlier one (it has that folder and no marker), the
/// way a real HOME that also holds another tool's folder of the new name does. A plain `home()`
/// resolves to the new one, which hid the defect.
fn earlier_home() -> Home {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(home.join(EARLIER)).unwrap();
    assert_eq!(moai_cli::home_root::path(&home), home.join(EARLIER));
    Home { _tmp: tmp, home }
}

/// One backup folder with this suffix stands under the new state folder's backups, nothing was made under the
/// earlier state folder, and the HOME's state folder did not move.
fn only_new_root(h: &Home, suffix: &str) {
    let made: Vec<String> = names_in(&h.home.join(NEWER).join("backups"))
        .into_iter()
        .filter(|n| n.ends_with(suffix))
        .collect();
    assert_eq!(
        made.len(),
        1,
        "one {suffix} backup under the new backups folder: {made:?}"
    );
    assert!(
        !h.home.join(EARLIER).join("backups").exists(),
        "nothing under the earlier state folder"
    );
    assert_eq!(
        moai_cli::home_root::path(&h.home),
        h.home.join(EARLIER),
        "the state folder did not move"
    );
}

#[test]
fn ac11_backups_go_under_the_moai_backups_folder_whichever_state_folder_the_home_uses() {
    let want = |h: &Home| json!(h.home.join(NEWER).join("backups").to_string_lossy());
    // A conflict replacement, on a plain earlier HOME and on one whose new-name folder holds another
    // tool's files, which stay as they are.
    for foreign in [false, true] {
        let h = earlier_home();
        let theirs = h.home.join(NEWER).join("other-tool/data.txt");
        if foreign {
            fs::create_dir_all(theirs.parent().unwrap()).unwrap();
            fs::write(&theirs, b"theirs\n").unwrap();
        }
        let (h, exe, preview) = user_edited_state_in(h);
        let offer = &preview["plan"]["hosts"][0]["conflict_replace"];
        assert_eq!(offer["backup_root"], want(&h), "{preview}");
        let token = offer["confirm_token"].as_str().unwrap().to_string();
        result(&run(
            &exe,
            &h.home,
            &[
                "update",
                "--confirm",
                &digest_of(&preview),
                "--confirm-conflict-replace",
                &token,
            ],
        ));
        only_new_root(&h, "-claude-conflict");
        if foreign {
            assert_eq!(fs::read(&theirs).unwrap(), b"theirs\n");
        }
    }

    // An adoption.
    let h = earlier_home();
    let exe = install(&h.home, "a", |_| {});
    copy_tree(&package(), &skill_dir(&h.home));
    let preview = result(&run(&exe, &h.home, &["update"]));
    assert_eq!(
        preview["plan"]["hosts"][0]["adopt"]["backup_root"],
        want(&h),
        "{preview}"
    );
    result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    only_new_root(&h, "-claude-adopt");

    // The removal of the earlier flow rule.
    let h = earlier_home();
    let exe = install(&h.home, "a", |_| {});
    write_old_flow_rule(&h.home, b"flow rule of an earlier build\n");
    write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
    let preview = result(&run(&exe, &h.home, &["update"]));
    assert_eq!(
        preview["routing"]["hosts"][0]["legacy_flow"]["backup_root"],
        want(&h),
        "{preview}"
    );
    result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    only_new_root(&h, "-claude-flow-rule");
}

/// An earlier HOME whose new-name folder cannot take a backup: a link leading out, or a folder that holds
/// a Go marker.
fn blocked_home(kind: &str) -> Home {
    let h = earlier_home();
    match kind {
        "link" => {
            let elsewhere = h.home.parent().unwrap().join("elsewhere");
            fs::create_dir_all(&elsewhere).unwrap();
            std::os::unix::fs::symlink(&elsewhere, h.home.join(NEWER)).unwrap();
        }
        _ => {
            fs::create_dir_all(h.home.join(NEWER)).unwrap();
            fs::write(h.home.join(NEWER).join("credentials.yaml"), b"x: 1\n").unwrap();
        }
    }
    h
}

fn has_step(list: &Value, step: &str) -> bool {
    list.as_array()
        .is_some_and(|l| l.iter().any(|b| b["step"] == json!(step)))
}

fn nothing_made(h: &Home) {
    assert!(!h.home.join(EARLIER).join("backups").exists());
    let outside = h.home.parent().unwrap().join("elsewhere");
    assert!(names_in(&outside).is_empty(), "no write through the link");
}

#[test]
fn ac12_a_root_that_cannot_take_a_backup_leaves_the_steps_that_need_one_undone() {
    for kind in ["link", "go"] {
        // A conflict stays as it is: no offer, a named reason, bytes kept.
        let (h, exe, preview) = user_edited_state_in(blocked_home(kind));
        let host = &preview["plan"]["hosts"][0];
        assert!(host["conflict_replace"].is_null(), "{kind}: {preview}");
        assert!(
            has_step(&host["backup_blocked"], "conflict-replace"),
            "{kind}: {preview}"
        );
        for (lang, line) in [
            (Lang::En, "Backup folder unusable"),
            (Lang::Ko, "백업 폴더를 쓸 수 없어요"),
        ] {
            let out = screen(&preview, Asked::NotAsked, &view(lang));
            assert!(out.contains(line), "{kind} {lang:?}: {out}");
        }
        result(&run(
            &exe,
            &h.home,
            &["update", "--confirm", &digest_of(&preview)],
        ));
        assert_eq!(
            fs::read_to_string(skill_dir(&h.home).join(FILE)).unwrap(),
            "the user's own words\n",
            "{kind}: the conflict file is kept"
        );
        nothing_made(&h);

        // An adoption is not made.
        let h = blocked_home(kind);
        let exe = install(&h.home, "a", |_| {});
        copy_tree(&package(), &skill_dir(&h.home));
        let before = tree(&skill_dir(&h.home));
        let preview = result(&run(&exe, &h.home, &["update"]));
        let host = &preview["plan"]["hosts"][0];
        assert!(host["adopt"].is_null(), "{kind}: {preview}");
        assert!(
            has_step(&host["backup_blocked"], "adopt"),
            "{kind}: {preview}"
        );
        let _ = run(
            &exe,
            &h.home,
            &["update", "--confirm", &digest_of(&preview)],
        );
        assert_eq!(
            tree(&skill_dir(&h.home)),
            before,
            "{kind}: no skill byte changed"
        );
        nothing_made(&h);

        // The earlier flow rule stays.
        let h = blocked_home(kind);
        let exe = install(&h.home, "a", |_| {});
        let file = write_old_flow_rule(&h.home, b"flow rule of an earlier build\n");
        write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
        let preview = result(&run(&exe, &h.home, &["update"]));
        let routing = &preview["routing"]["hosts"][0];
        assert!(routing["legacy_flow"].is_null(), "{kind}: {preview}");
        assert!(
            has_step(&routing["backup_blocked"], "flow-rule-removal"),
            "{kind}: {preview}"
        );
        let _ = run(
            &exe,
            &h.home,
            &["update", "--confirm", &digest_of(&preview)],
        );
        assert_eq!(
            fs::read_to_string(&file).unwrap(),
            "flow rule of an earlier build\n",
            "{kind}: the flow rule file stays"
        );
        nothing_made(&h);
    }
}

/// Every file under `root` whose text holds `from` gets `to` instead; returns how many changed.
fn rewrite_in_tree(root: &Path, from: &str, to: &str) -> usize {
    let mut n = 0;
    for name in names_in(root) {
        let p = root.join(&name);
        if p.is_dir() {
            n += rewrite_in_tree(&p, from, to);
        } else if let Ok(text) = fs::read_to_string(&p) {
            if text.contains(from) {
                fs::write(&p, text.replace(from, to)).unwrap();
                n += 1;
            }
        }
    }
    n
}

/// `run` with a JSON document on the standard input.
fn run_stdin(exe: &Path, home: &Path, args: &[&str], stdin: &Value) -> Output {
    use std::io::Write as _;
    let mut child = Command::new(exe)
        .args(args)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", home)
        .env("LC_ALL", "C")
        .current_dir(home)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut pipe = child.stdin.take().unwrap();
    let _ = pipe.write_all(stdin.to_string().as_bytes());
    drop(pipe);
    child.wait_with_output().unwrap()
}

/// A HOME whose skill folder was backed up and replaced by a confirmed update, with the backup
/// folder then moved to `<HOME>/<place>` and the receipt pointed at it, the way an earlier update
/// or a stray folder would leave it. Returns the HOME, the skill folder, its original files, the
/// folder that now holds the backup and the backup folder's name.
fn replaced_with_backup_moved_to(
    place: &str,
) -> (
    Home,
    PathBuf,
    std::collections::BTreeMap<String, Vec<u8>>,
    PathBuf,
    String,
) {
    let h = earlier_home();
    let stage = h.home.parent().unwrap().join("stage");
    fs::create_dir_all(stage.join("references")).unwrap();
    fs::write(
        stage.join("SKILL.md"),
        "---\nname: moai\ndescription: fixture package\n---\n# MoAI\nrust\n",
    )
    .unwrap();
    fs::write(stage.join("references/routes.json"), "{\"routes\":1}\n").unwrap();
    let skill = skill_dir(&h.home);
    fs::create_dir_all(skill.join("modules")).unwrap();
    fs::write(
        skill.join("SKILL.md"),
        "---\nname: moai\n---\n# Go-era MoAI\n",
    )
    .unwrap();
    fs::write(skill.join("modules/run.md"), "go run module\n").unwrap();
    let original = tree(&skill);
    let exe = Path::new(EXE);
    let (home_s, stage_s) = (h.home.to_str().unwrap(), stage.to_str().unwrap());

    let plan = result(&run(exe, &h.home, &["update-plan", home_s, stage_s]));
    let row = plan["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["host"] == json!("claude"))
        .expect("host row")
        .clone();
    let token = row["backup_then_replace"]["confirm_token"]
        .as_str()
        .unwrap_or_else(|| panic!("no replace offer: {plan}"))
        .to_string();
    let digest = digest_of(&plan);
    let applied = envelope(&run_stdin(
        exe,
        &h.home,
        &[
            "update-apply",
            home_s,
            stage_s,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
            "--confirm-backup-replace",
            &token,
        ],
        &plan,
    ));
    assert_eq!(applied["ok"], json!(true), "{applied}");

    // The backup stands under the new backups folder; move it to `place` and point the receipt there.
    let (new_root, old_root) = (h.home.join(NEWER).join("backups"), h.home.join(place));
    let made = names_in(&new_root);
    assert_eq!(made.len(), 1, "{made:?}");
    fs::create_dir_all(&old_root).unwrap();
    fs::rename(new_root.join(&made[0]), old_root.join(&made[0])).unwrap();
    fs::remove_dir(&new_root).unwrap();
    let moved = rewrite_in_tree(
        &h.home.join(EARLIER).join("installations"),
        &new_root.to_string_lossy(),
        &old_root.to_string_lossy(),
    );
    assert!(moved > 0, "the receipt names the backup folder");
    (h, skill, original, old_root, made[0].clone())
}

/// A backup an earlier update left under the earlier state folder is still read by the recovery,
/// and what the recovery only reads of it is the same afterwards.
#[test]
fn ac13_recovery_reads_an_earlier_backup_under_the_earlier_state_folder() {
    let (h, skill, original, old_root, name) =
        replaced_with_backup_moved_to(&format!("{EARLIER}/backups"));
    let new_root = h.home.join(NEWER).join("backups");
    let exe = Path::new(EXE);
    let home_s = h.home.to_str().unwrap();
    let before = tree(&old_root);

    let rp = result(&run(exe, &h.home, &["recovery-plan", home_s]));
    let rb = rp["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["host"] == json!("claude"))
        .expect("host row")["restore_backup"]
        .clone();
    assert_eq!(rb["action"], json!("restore_backup"), "{rp}");
    let r = result(&run_stdin(
        exe,
        &h.home,
        &[
            "recovery-apply",
            home_s,
            "--plan-digest",
            rp["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &rp,
    ));
    assert_eq!(r["status"], json!("success"), "{r}");
    assert_eq!(tree(&skill), original, "the original folder is back");
    // The recovery moves the original tree out of its backup by design; what it only reads (the
    // manifest, the file list) is as it was, the folder is where it was, and nothing was made
    // under the new state folder.
    let after = tree(&old_root);
    assert!(old_root.join(&name).is_dir(), "the backup folder stays");
    let manifest = format!("{name}/MANIFEST.json");
    assert_eq!(
        after.get(&manifest),
        before.get(&manifest),
        "manifest bytes"
    );
    for (path, bytes) in &before {
        if !path.contains("/moai/") {
            assert_eq!(after.get(path), Some(bytes), "{path} is as it was");
        }
    }
    assert!(
        !new_root.exists(),
        "nothing was made under the new state folder"
    );
}

/// The case the retirement of the earlier folder makes real: the HOME's state has since moved to the
/// new folder (a valid marker, the receipts there), while the backup an earlier update made still
/// stands under the earlier folder. The recovery reads it from there.
#[test]
fn ac13_recovery_reads_the_earlier_backup_once_the_state_is_the_new_folder() {
    let (h, skill, original, old_root, name) =
        replaced_with_backup_moved_to(&format!("{EARLIER}/backups"));
    let (earlier, newer) = (h.home.join(EARLIER), h.home.join(NEWER));
    fs::create_dir_all(&newer).unwrap();
    fs::rename(earlier.join("installations"), newer.join("installations")).unwrap();
    fs::write(
        newer.join(moai_cli::state_root::RUST_MARKER),
        moai_cli::state_root::marker_bytes("test", &[]),
    )
    .unwrap();
    assert_eq!(moai_cli::home_root::path(&h.home), newer, "the state moved");
    assert!(old_root.join(&name).is_dir());
    let exe = Path::new(EXE);
    let home_s = h.home.to_str().unwrap();
    let rp = result(&run(exe, &h.home, &["recovery-plan", home_s]));
    let rb = rp["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["host"] == json!("claude"))
        .expect("host row")["restore_backup"]
        .clone();
    assert_eq!(rb["action"], json!("restore_backup"), "{rp}");
    let r = result(&run_stdin(
        exe,
        &h.home,
        &[
            "recovery-apply",
            home_s,
            "--plan-digest",
            rp["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &rp,
    ));
    assert_eq!(r["status"], json!("success"), "{r}");
    assert_eq!(tree(&skill), original, "the original folder is back");
    assert!(old_root.join(&name).is_dir(), "the backup folder stays");
}

/// The other side of AC-13: a receipt that names a backup folder outside the known backup folders
/// is not read, so the recovery restores nothing from it and leaves it as it is.
#[test]
fn ac13_recovery_does_not_read_a_backup_outside_the_known_folders() {
    let (h, skill, _original, elsewhere, name) = replaced_with_backup_moved_to("stray-backups");
    let exe = Path::new(EXE);
    let home_s = h.home.to_str().unwrap();
    let replaced = tree(&skill);
    let before = tree(&elsewhere);
    let rp = result(&run(exe, &h.home, &["recovery-plan", home_s]));
    let rb = rp["hosts"]
        .as_array()
        .unwrap()
        .iter()
        .find(|x| x["host"] == json!("claude"))
        .expect("host row")["restore_backup"]
        .clone();
    assert_ne!(rb["action"], json!("restore_backup"), "{rp}");
    assert_eq!(tree(&skill), replaced, "the replaced package is untouched");
    assert_eq!(tree(&elsewhere), before, "the stray backup is untouched");
    assert!(elsewhere.join(&name).is_dir());
}

#[test]
fn ac14_the_emptied_rules_folder_is_removed_and_a_folder_with_other_files_stays() {
    let folder_of = |h: &Home| h.home.join(".claude/rules/moai");
    // Emptied: the folder is gone, its parent stays, the screen names it.
    let h = home();
    let exe = install(&h.home, "a", |_| {});
    let file = write_old_flow_rule(&h.home, b"flow rule of an earlier build\n");
    write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
    let preview = result(&run(&exe, &h.home, &["update"]));
    let applied = result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert!(!file.exists());
    assert!(!folder_of(&h).exists(), "the emptied folder is gone");
    assert!(h.home.join(".claude/rules").is_dir(), "its parent stays");
    let removed = &applied["routing_result"]["legacy_flow_removed"][0];
    assert_eq!(
        removed["folder_removed"],
        json!(folder_of(&h).to_string_lossy()),
        "{applied}"
    );
    let mut shown = fixture("applied");
    shown["routing_result"] = json!({
        "state": "applied",
        "legacy_flow_removed": [{
            "path": "/home/u/.claude/rules/moai/flow.md",
            "backup": "/home/u/state/backups/20261125T000000Z-claude-flow-rule",
            "folder_removed": "/home/u/.claude/rules/moai"
        }]
    });
    for (lang, words) in [(Lang::En, "empty folder"), (Lang::Ko, "빈 폴더")] {
        let out = screen(&shown, Asked::Yes, &view(lang));
        assert!(
            out.contains(words) && out.contains("rules/moai"),
            "{lang:?}: {out}"
        );
    }

    // Another file in the folder: the folder and that file stay, and no such line is shown.
    let h = home();
    let exe = install(&h.home, "a", |_| {});
    let file = write_old_flow_rule(&h.home, b"flow rule of an earlier build\n");
    let mine = folder_of(&h).join("mine.md");
    fs::write(&mine, b"my own rule\n").unwrap();
    write_flow_receipt(&h.home, b"flow rule of an earlier build\n");
    let preview = result(&run(&exe, &h.home, &["update"]));
    let applied = result(&run(
        &exe,
        &h.home,
        &["update", "--confirm", &digest_of(&preview)],
    ));
    assert!(!file.exists(), "the flow rule file itself is removed");
    assert_eq!(fs::read(&mine).unwrap(), b"my own rule\n");
    assert!(
        applied["routing_result"]["legacy_flow_removed"][0]["folder_removed"].is_null(),
        "{applied}"
    );
    let mut shown = fixture("applied");
    shown["routing_result"] = json!({
        "state": "applied",
        "legacy_flow_removed": [{
            "path": "/home/u/.claude/rules/moai/flow.md",
            "backup": "/home/u/state/backups/20261125T000000Z-claude-flow-rule"
        }]
    });
    let out = screen(&shown, Asked::Yes, &view(Lang::En));
    assert!(!out.contains("empty folder"), "{out}");
}
