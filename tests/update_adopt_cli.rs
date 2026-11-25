//! SPEC-20261123-update-adopt-unowned: `moai update` adopts a skill folder whose files the
//! install receipt does not own but whose bytes equal the package or an earlier generation, and
//! the version folder layout is recognised under either state folder name.
//!
//! Every case runs in a temporary HOME with a copy of this build's binary; the real HOME, the
//! real `~/.claude` and both real state folders are never read or written.

#![cfg(unix)]

mod common;
use common::CleanEnv;

// Folder helpers live in a module of their own: they walk only temporary HOMEs and the named
// package folders, and the test file itself then reads no tree.
#[path = "update_adopt_cli/trees.rs"]
mod trees;
use trees::{copy_tree, names_in, tree};

use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
    process::{Command, Output, Stdio},
};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
/// The two names a HOME state folder can have: the earlier one and the new one.
const EARLIER: &str = ".moai-cli";
const NEWER: &str = ".moai";

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

/// A temporary HOME whose state folder is the earlier one (kept by a HOME that has it and no
/// marker), with the skill folder of the host `claude` left absent.
fn home() -> Home {
    let tmp = TempDir::new().unwrap();
    let home = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(home.join(EARLIER)).unwrap();
    Home { _tmp: tmp, home }
}

/// A version folder `<home>/<state>/versions/<name>` holding a copy of this build's binary, the
/// package and the agent sources; `edit` may change the package copy. Returns the binary.
fn install(home: &Path, state: &str, name: &str, edit: impl FnOnce(&Path)) -> PathBuf {
    let folder = home.join(state).join("versions").join(name);
    fs::create_dir_all(folder.join("bin")).unwrap();
    let exe = folder.join("bin/moai-cli");
    common::copy_executable(Path::new(EXE), &exe);
    copy_tree(&package(), &folder.join("skills/moai"));
    copy_tree(&agent_sources(), &folder.join("agents"));
    edit(&folder.join("skills/moai"));
    exe
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

fn host(preview: &Value) -> &Value {
    &preview["plan"]["hosts"][0]
}

fn actions(host: &Value) -> BTreeMap<String, String> {
    host["files"]
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

/// The backup folders: always under the HOME's `.moai`, whichever folder holds the state
/// (SPEC-20261125-update-conflict-baseline revision 3, REQ-17).
fn backups(home: &Path) -> Vec<String> {
    names_in(&home.join(".moai/backups"))
}

fn host_receipts(home: &Path) -> Vec<Value> {
    let dir = home.join(EARLIER).join("installations");
    let mut out = Vec::new();
    for id in names_in(&dir) {
        if let Ok(bytes) = fs::read(dir.join(id).join("hosts/moai/claude.json")) {
            out.push(serde_json::from_slice(&bytes).unwrap());
        }
    }
    out
}

// ---- AC-1: a byte copy of the package is adopted in the preview, nothing is written --------

#[test]
fn a_byte_copy_of_the_package_is_adopted_in_the_preview_and_nothing_is_written() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    copy_tree(&package(), &skill_dir(&h.home));
    let before = tree(&h.home);

    let p = result(&run(&exe, &h.home, &["update"]));

    assert_eq!(tree(&h.home), before, "a preview writes nothing");
    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("ok"), "{p}");
    let adopt = &claude["adopt"];
    let files = adopt["files"].as_array().expect("an adopt list");
    assert_eq!(files.len(), tree(&package()).len(), "{adopt}");
    assert!(
        files.iter().all(|f| f["matched"] == json!("package")),
        "{adopt}"
    );
    assert!(
        actions(claude).values().all(|a| a == "unchanged"),
        "{claude}"
    );
    assert_eq!(
        p["nothing_to_apply"],
        json!(false),
        "an adoption is a write: {p}"
    );
}

// ---- AC-2: an earlier generation matches; a file of the user blocks everything ---------------

fn old_skill_md(pkg: &Path) {
    let p = pkg.join("SKILL.md");
    let mut text = fs::read_to_string(&p).unwrap();
    text.push_str("\nan earlier generation line\n");
    fs::write(p, text).unwrap();
}

#[test]
fn a_copy_of_an_earlier_version_folder_is_adopted_with_the_generation_it_matched() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    install(&h.home, EARLIER, "old", old_skill_md);
    copy_tree(
        &h.home.join(EARLIER).join("versions/old/skills/moai"),
        &skill_dir(&h.home),
    );

    let p = result(&run(&exe, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("ok"), "{p}");
    let files = claude["adopt"]["files"].as_array().expect("an adopt list");
    let skill = files
        .iter()
        .find(|f| f["path"] == json!("SKILL.md"))
        .expect("SKILL.md is adopted");
    assert!(
        skill["matched"].as_str().unwrap().starts_with("generation"),
        "{skill}"
    );
    assert!(
        skill["matched"].as_str().unwrap().contains("old"),
        "{skill}"
    );
    let acts = actions(claude);
    assert_eq!(acts["SKILL.md"], "update", "{claude}");
    assert!(
        acts.iter()
            .filter(|(path, _)| path.as_str() != "SKILL.md")
            .all(|(_, a)| a == "unchanged"),
        "{acts:?}"
    );
}

/// A skill folder that is a copy of the earlier folder `old`'s package, with `old` beside `exe`.
fn old_copy_in(h: &Home) {
    install(&h.home, EARLIER, "old", old_skill_md);
    copy_tree(
        &h.home.join(EARLIER).join("versions/old/skills/moai"),
        &skill_dir(&h.home),
    );
}

#[test]
fn a_build_outside_the_layout_lends_nothing_and_stays_blocked() {
    let h = home();
    old_copy_in(&h);
    let dir = h.home.join("elsewhere/bin");
    fs::create_dir_all(&dir).unwrap();
    let outside = dir.join("moai-cli");
    common::copy_executable(Path::new(EXE), &outside);
    copy_tree(&package(), &dir.join("../skills/moai"));
    copy_tree(&agent_sources(), &dir.join("../agents"));

    let p = result(&run(&outside, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("blocked"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
}

#[test]
fn a_linked_version_folder_lends_nothing_and_stays_blocked() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    old_copy_in(&h);
    let versions = h.home.join(EARLIER).join("versions");
    let moved = h.home.join("moved-old");
    fs::rename(versions.join("old"), &moved).unwrap();
    std::os::unix::fs::symlink(&moved, versions.join("old")).unwrap();

    let p = result(&run(&exe, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("blocked"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
}

#[test]
fn one_file_of_the_user_keeps_the_host_blocked_with_nothing_adopted() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    copy_tree(&package(), &skill_dir(&h.home));
    fs::write(skill_dir(&h.home).join("SKILL.md"), "my own skill text\n").unwrap();
    let before = tree(&h.home);

    let p = result(&run(&exe, &h.home, &["update"]));

    assert_eq!(tree(&h.home), before);
    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("blocked"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
    let reason = claude["blocked_reason"].as_str().unwrap();
    assert!(reason.contains("unmanaged existing package"), "{reason}");
    assert!(reason.contains("SKILL.md"), "{reason}");
}

/// A sibling version folder that has the same path with other bytes lends nothing for a file the
/// user edited: only equal bytes are ours (the sha256 of the file picks the generation).
#[test]
fn a_user_edit_of_a_path_an_earlier_version_folder_has_is_not_adopted() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    install(&h.home, EARLIER, "old", old_skill_md);
    copy_tree(&package(), &skill_dir(&h.home));
    let edited = skill_dir(&h.home).join("SKILL.md");
    fs::write(&edited, "my own skill text, not any generation\n").unwrap();
    let before = tree(&h.home);

    let p = result(&run(&exe, &h.home, &["update"]));

    assert_eq!(tree(&h.home), before, "a preview writes nothing");
    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("blocked"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    run(&exe, &h.home, &["update", "--confirm", &digest]);
    assert_eq!(
        fs::read(&edited).unwrap(),
        b"my own skill text, not any generation\n",
        "the user's bytes stay"
    );
}

/// With two earlier folders that disagree on a file, the file is adopted from the folder whose
/// bytes it equals, and the answer names that folder.
#[test]
fn a_file_equal_to_the_second_of_two_earlier_generations_names_that_generation() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    install(&h.home, EARLIER, "old1", |pkg| {
        fs::write(pkg.join("SKILL.md"), "first earlier generation\n").unwrap();
    });
    install(&h.home, EARLIER, "old2", |pkg| {
        fs::write(pkg.join("SKILL.md"), "second earlier generation\n").unwrap();
    });
    copy_tree(&package(), &skill_dir(&h.home));
    fs::write(
        skill_dir(&h.home).join("SKILL.md"),
        "second earlier generation\n",
    )
    .unwrap();

    let p = result(&run(&exe, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("ok"), "{p}");
    let skill = claude["adopt"]["files"]
        .as_array()
        .expect("an adopt list")
        .iter()
        .find(|f| f["path"] == json!("SKILL.md"))
        .expect("SKILL.md is adopted");
    assert!(
        skill["matched"].as_str().unwrap().contains("old2"),
        "{skill}"
    );
}

/// A link inside an earlier version folder lends nothing either: `skills` linked to a folder
/// elsewhere is not followed, so the bytes found there never make a user's file ours.
#[test]
fn a_linked_skills_folder_inside_an_earlier_version_folder_lends_nothing() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    old_copy_in(&h);
    let old = h.home.join(EARLIER).join("versions/old");
    let moved = h.home.join("moved-skills");
    fs::rename(old.join("skills"), &moved).unwrap();
    std::os::unix::fs::symlink(&moved, old.join("skills")).unwrap();
    let user_bytes = fs::read(skill_dir(&h.home).join("SKILL.md")).unwrap();
    let before = tree(&h.home);

    let p = result(&run(&exe, &h.home, &["update"]));

    assert_eq!(tree(&h.home), before, "a preview writes nothing");
    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("blocked"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    run(&exe, &h.home, &["update", "--confirm", &digest]);
    assert_eq!(
        fs::read(skill_dir(&h.home).join("SKILL.md")).unwrap(),
        user_bytes,
        "the user's bytes stay"
    );
}

/// The package folder `skills/moai` of an earlier version folder linked elsewhere lends nothing.
#[test]
fn a_linked_package_folder_inside_an_earlier_version_folder_lends_nothing() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    old_copy_in(&h);
    let skills = h.home.join(EARLIER).join("versions/old/skills");
    let moved = h.home.join("moved-moai");
    fs::rename(skills.join("moai"), &moved).unwrap();
    std::os::unix::fs::symlink(&moved, skills.join("moai")).unwrap();

    let p = result(&run(&exe, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("blocked"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
}

// ---- AC-3: the confirmed run backs up, adopts and applies -------------------------------------

#[test]
fn the_confirmed_run_backs_up_then_owns_the_adopted_files() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    install(&h.home, EARLIER, "old", old_skill_md);
    copy_tree(
        &h.home.join(EARLIER).join("versions/old/skills/moai"),
        &skill_dir(&h.home),
    );
    let adopted_before = tree(&skill_dir(&h.home));
    let p = result(&run(&exe, &h.home, &["update"]));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    assert!(backups(&h.home).is_empty());

    let r = result(&run(&exe, &h.home, &["update", "--confirm", &digest]));

    assert_eq!(r["mode"], json!("applied"), "{r}");
    let names = backups(&h.home);
    assert_eq!(names.len(), 1, "{names:?}");
    assert!(names[0].ends_with("-claude-adopt"), "{names:?}");
    let dir = h.home.join(".moai/backups").join(&names[0]);
    let manifest: Value =
        serde_json::from_slice(&fs::read(dir.join("MANIFEST.json")).unwrap()).unwrap();
    assert_eq!(
        manifest["files"].as_array().unwrap().len(),
        adopted_before.len(),
        "{manifest}"
    );
    assert_eq!(
        tree(&dir.join("files")),
        adopted_before,
        "the backup holds every adopted file as it was"
    );
    assert_eq!(
        tree(&skill_dir(&h.home)),
        tree(&package()),
        "the skill folder then equals the package"
    );
    let receipts = host_receipts(&h.home);
    assert_eq!(receipts.len(), 1, "{receipts:?}");
    let receipt = &receipts[0];
    assert_eq!(receipt["status"], json!("committed"), "{receipt}");
    assert!(receipt["backup"].is_null(), "{receipt}");
    assert_eq!(
        receipt["owned"].as_array().unwrap().len(),
        tree(&package()).len(),
        "{receipt}"
    );
}

#[test]
fn a_file_changed_after_the_preview_is_refused_and_nothing_is_written() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    copy_tree(&package(), &skill_dir(&h.home));
    let p = result(&run(&exe, &h.home, &["update"]));
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    fs::write(
        skill_dir(&h.home).join("SKILL.md"),
        "changed after the preview\n",
    )
    .unwrap();
    let before = tree(&h.home);

    let out = run(&exe, &h.home, &["update", "--confirm", &digest]);
    let v = envelope(&out);

    assert_eq!(v["ok"], json!(false), "{v}");
    assert!(
        v["error"]["message"]
            .as_str()
            .unwrap()
            .contains("stale or altered plan"),
        "{v}"
    );
    assert_eq!(tree(&h.home), before, "nothing was written");
}

// ---- AC-4: the next update does not block ---------------------------------------------------

#[test]
fn the_next_update_lists_no_adoption_and_does_not_block() {
    let h = home();
    let exe = install(&h.home, EARLIER, "a", |_| {});
    copy_tree(&package(), &skill_dir(&h.home));
    let digest = result(&run(&exe, &h.home, &["update"]))["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    result(&run(&exe, &h.home, &["update", "--confirm", &digest]));

    let p = result(&run(&exe, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("ok"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
    assert!(
        actions(claude).values().all(|a| a == "unchanged"),
        "{claude}"
    );
}

// ---- AC-5: the version folder layout under either state folder name -------------------------

#[test]
fn a_binary_in_the_moai_versions_folder_is_inside_the_layout() {
    let h = home();
    let exe = install(&h.home, NEWER, "a", |_| {});

    let p = result(&run(
        &exe,
        &h.home,
        &["bin-names-plan", "--home", h.home.to_str().unwrap()],
    ));

    assert_eq!(p["state"], json!("create"), "{p}");
}

#[test]
fn a_binary_elsewhere_is_outside_the_layout_and_the_reason_names_both_folders() {
    let h = home();
    let dir = h.home.join("elsewhere/bin");
    fs::create_dir_all(&dir).unwrap();
    let exe = dir.join("moai-cli");
    common::copy_executable(Path::new(EXE), &exe);

    let p = result(&run(
        &exe,
        &h.home,
        &["bin-names-plan", "--home", h.home.to_str().unwrap()],
    ));

    assert_eq!(p["state"], json!("outside_layout"), "{p}");
    let reason = p["reason"].as_str().unwrap();
    assert!(reason.contains(&format!("{NEWER}/versions")), "{reason}");
    assert!(reason.contains(&format!("{EARLIER}/versions")), "{reason}");
}

#[test]
fn the_receipt_of_a_sibling_folder_under_moai_versions_is_the_baseline() {
    let h = home();
    let old = install(&h.home, NEWER, "old", old_skill_md);
    let digest = result(&run(&old, &h.home, &["update"]))["plan_digest"]
        .as_str()
        .unwrap()
        .to_string();
    result(&run(&old, &h.home, &["update", "--confirm", &digest]));
    let exe = install(&h.home, NEWER, "a", |_| {});

    let p = result(&run(&exe, &h.home, &["update"]));

    let claude = host(&p);
    assert_eq!(claude["preflight"], json!("ok"), "{p}");
    assert!(claude.get("adopt").is_none(), "{claude}");
    assert!(claude["receipt_from"].is_string(), "{claude}");
    assert_eq!(actions(claude)["SKILL.md"], "update", "{claude}");
}
