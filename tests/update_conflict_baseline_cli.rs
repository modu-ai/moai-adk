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
use trees::{copy_tree, names_in};

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
const STATE: &str = ".moai";
/// A file every package carries, edited per generation below.
const FILE: &str = "workflows/work.md";

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
    fs::create_dir_all(home.join(STATE)).unwrap();
    Home { _tmp: tmp, home }
}

/// A version folder `<home>/.moai/versions/<name>` with a copy of this build's binary, the
/// package and the agent sources; `edit` changes the package copy. Returns the binary.
fn install(home: &Path, name: &str, edit: impl FnOnce(&Path)) -> PathBuf {
    let folder = home.join(STATE).join("versions").join(name);
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
    let dir = home.join(STATE).join("installations");
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

    let applied = apply(&current, &h.home);

    assert_ne!(applied["mode"], json!("refused"), "{applied}");
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
    assert_eq!(
        sha_of(&skill_dir(&h.home).join(FILE)),
        sha_of(&package().join(FILE)),
        "no skill byte was rewritten"
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
        json!({"state": "applied", "saved": "/home/u/.moai/saved", "revert": "moai switch-revert"});
    let done = screen(&v, Asked::Yes, &view(Lang::En));
    assert!(!done.contains("moai update --confirm-switch"), "{done}");
}
