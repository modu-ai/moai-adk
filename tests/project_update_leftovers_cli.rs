//! SPEC-20261127-update-leftovers-action, plan side: `project-update-plan` gives each leftover of
//! an earlier generation an archive action or a reason it can only be kept, and a recorded keep
//! counts. Every case runs the built binary in a temporary project with an empty temporary HOME
//! (the shared `GoFx` fixture); the real checkout and HOME are never read.
//!
//! The JSON these tests fix (the GREEN slice follows it; no wording beyond one keyword per
//! reason is pinned). A leftover entry keeps `path`, `fact` and `note`. For the five kinds (the
//! earlier state folder, a `<state>.rust-<name>/` set-aside folder, the Go backups folder, the
//! Go update lock at the state folder and at the Go backup folder) it gains:
//!
//! - `kept`: always present, a bool, true while `manifests/update-declined.json` holds the
//!   entry's current source sha256 under the key `leftover:<path>` (`path` as printed). A kept
//!   entry still prints its `action` or `keep_only`.
//! - `keep`: always present, the argv of `project-update-decline`, exactly the form
//!   `declined::decline_argv` prints: `moai project-update-decline <project> --step
//!   leftover:<path> --source-sha256 <sha256:hex> --permission-reference <placeholder>`.
//! - `ask_again`: present only while `kept`, the argv `declined::ask_again_argv` prints for that
//!   step.
//! - `action`: the argv `moai project-update-leftovers-apply <project> --plan-digest <sha256:hex>
//!   --permission-reference <placeholder> --item <path>`, one plan digest for all items of a plan.
//!   Absent while the item cannot be archived and while another backup-phase step still has a
//!   command.
//! - `keep_only`: a non-empty reason string, present when the item cannot be archived (at most
//!   one keyword per case is pinned below). A folder path ends in `/`, a lock path does not.
//!
//! A `CLAUDE.md` leftover carries none of these keys and never holds `all_unchanged` false.
//! `finish.open` of an update with no question lists an open leftover as `{"leftover": <path>,
//! "say": <sentence>}`. The apply result prints the receipt's UTC name as `receipt`, and the
//! printed `undo_order` begins with a row `leftovers` (command `project-update-leftovers-revert`
//! with the placeholder `<receipt of the leftovers step>`) only while an archive receipt that is
//! not reverted exists (the end-to-end walk of shape E reads that row).
//!
//! The apply and the revert (the words below are what the GREEN slice follows; nothing else of a
//! message is pinned):
//!
//! - `project-update-leftovers-apply <project> --plan-digest D --permission-reference R --item P
//!   [--item P...]`: the command is built from the printed `action` (one digest, the placeholder
//!   filled), one `--item` more per extra path. Success: `result.state` `committed`;
//!   `result.receipt` the UTC name `YYYYMMDDTHHMMSSZ`; `result.items` one `{"path": P,
//!   "outcome": "moved"}` per chosen path; `result.tracked_moved` the number of moved files git
//!   tracks (0 when none); `result.warning` null when it is 0, else one sentence with that number,
//!   the word `checkout` and the word `committed`; `result.secret_shaped` the number of
//!   secret-shaped top-level file names inside the moved folders (counted, never printed).
//! - A refusal is an error envelope (`ok` false) that writes nothing. A plan digest that is not
//!   the fresh plan's is kind `stale_plan` (the phrase `stale or altered plan` of the go-residue
//!   refusal); the other refusals (a rejected permission reference, an unknown item, an absolute
//!   path, a `..`, a link, a live lock process, a receipt that cannot be saved) pin only that the
//!   command ran and refused. The simulated cross-device check is the fault point
//!   `leftovers_cross_device` of `MOAI_TEST_FAULT` (the `update::test_fault` seam); its refusal
//!   says `simulated`; the apply stops before the backup area and the pending receipt exist.
//! - The archive: `backups/<UTC>/leftovers/<path without the trailing slash>` under this
//!   generation's state folder; the receipt: `backups/<UTC>/leftovers-receipt.json` there, with
//!   `state` (`pending`, `committed`,
//!   `reverted`) and `items`, one object per item with `path` (as printed), `kind` (`folder` or
//!   `file`), `entries` (a folder's top-level entry count), `size`, `listing_digest` (the item's
//!   source sha256 of the plan) and `moved`. It holds no file's contents and no child name.
//! - `project-update-leftovers-revert <project> --receipt <UTC> --permission-reference R`: the
//!   flag is given in addition to the SPEC's minimum because `go-residue-revert` requires it and
//!   the printed undo row names `<reference>`. Success: `result.state` `reverted` (every item
//!   back) or `conflict`; `result.receipt`; `result.items` one `{"path", "outcome"}` per receipt
//!   item with outcome `restored`, `conflict` or `missing`. An item whose path fails the path
//!   guard makes the whole revert move nothing: `ok` true, `state` `conflict`, that item
//!   `conflict`.
//! - The undo row: step `leftovers`, one command that names `project-update-leftovers-revert`,
//!   `<PROJECT>`, `<receipt of the leftovers step>` and `<reference>` (the placeholders the other
//!   rows use), so that no other placeholder is left once those three are filled.
//!
//! No test here spells the state folder's name: the scan of `state_folder_literal_scan` counts
//! every spelling, so the names come from the `state_root` constants.

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{sha, step_in, GoFx, REF};
use moai_cli::project_update::leftovers::{listing_digest, rel_of};
use moai_cli::project_update::{declined::OPERATION_WORD, PERMISSION_PLACEHOLDER};
use moai_cli::state_root::{self, EARLIER_PROJECT_DIR, GO_BACKUP, GO_STATE_RECEIPT, MOAI_DIR};
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::os::unix::fs::PermissionsExt as _;
use std::path::{Path, PathBuf};

const APPLY_WORD: &str = "project-update-leftovers-apply";
const REVERT_WORD: &str = "project-update-leftovers-revert";
/// The fault point of the simulated cross-device check (`MOAI_TEST_FAULT`).
const CROSS_DEVICE: &str = "leftovers_cross_device";
/// The receipt's file name under `backups/<UTC>/` of the state folder, and the folder that holds
/// the items.
const RECEIPT_FILE: &str = "leftovers-receipt.json";
const ARCHIVE: &str = "leftovers";
/// A stem no other text of a result, a receipt or a plan fact contains.
const SECRET_STEM: &str = "zq41";
/// The set-aside folder's stamp, and the stamps of the receipts inside the folders.
const ASIDE_STAMP: &str = "20261101T000000Z";
const EARLIER_RECEIPT_STAMP: &str = "20261102T000000Z";
const ASIDE_RECEIPT_STAMP: &str = "20261103T000000Z";
/// The Go backups folder (not the Go backup folder `GO_BACKUP`).
const BACKUPS: &str = ".moai-backups/";
const CURATOR_BLOCK: &str =
    "# Notes\n<!-- moai:learned-start -->\nlearned\n<!-- moai:learned-end -->\n";

fn earlier() -> String {
    format!("{EARLIER_PROJECT_DIR}/")
}

fn aside() -> String {
    format!("{MOAI_DIR}.rust-{ASIDE_STAMP}/")
}

fn state_lock() -> String {
    format!("{MOAI_DIR}/.update.lock")
}

fn backup_lock() -> String {
    format!("{GO_BACKUP}/.update.lock")
}

/// The five kinds, in the order the SPEC lists them.
fn five() -> Vec<String> {
    vec![
        earlier(),
        aside(),
        BACKUPS.to_string(),
        state_lock(),
        backup_lock(),
    ]
}

/// The pid of a process that has finished: no live process has it now.
fn dead_pid() -> u32 {
    let mut child = std::process::Command::new("true").spawn().unwrap();
    let pid = child.id();
    child.wait().unwrap();
    pid
}

fn lock_body() -> String {
    json!({"pid": dead_pid(), "started_at": "2026-11-01T00:00:00Z", "hostname": "h"}).to_string()
}

/// A go-state receipt a reader accepts: every required field, the canonical project path, and the
/// set-aside fields when the receipt sits in a `<state>.rust-<UTC>/` folder.
fn receipt(fx: &GoFx, stamp: &str, state: &str, aside_folder: Option<&str>) -> String {
    let mut r = json!({
        "schema_version": 1, "project": fx.p(), "backup": stamp,
        "plan_digest": "sha256:fixture", "permission_reference": "r",
        "state": state, "moves": [],
    });
    if let Some(folder) = aside_folder {
        r["go_backup"] = json!(GO_BACKUP);
        r["rust_aside"] = json!(folder);
    }
    r.to_string()
}

/// A receipt in the earlier state folder.
fn earlier_receipt(fx: &GoFx, state: &str) {
    fx.write(
        &format!("{EARLIER_PROJECT_DIR}/backups/{EARLIER_RECEIPT_STAMP}/{GO_STATE_RECEIPT}"),
        &receipt(fx, EARLIER_RECEIPT_STAMP, state, None),
    );
}

/// A receipt in the set-aside folder `folder` (its name without the trailing slash).
fn aside_receipt(fx: &GoFx, folder: &str, state: &str) {
    fx.write(
        &format!("{folder}/backups/{ASIDE_RECEIPT_STAMP}/{GO_STATE_RECEIPT}"),
        &receipt(fx, ASIDE_RECEIPT_STAMP, state, Some(folder)),
    );
}

fn steps_of(plan: &Value) -> Vec<&Value> {
    plan["phases"]
        .as_array()
        .unwrap()
        .iter()
        .flat_map(|p| p["steps"].as_array().unwrap())
        .collect()
}

/// The steps that still want something (neither `unchanged`, `skip` nor `optional`).
fn open_steps(plan: &Value) -> Vec<String> {
    steps_of(plan)
        .into_iter()
        .filter(|s| {
            !matches!(
                s["state"].as_str().unwrap(),
                "unchanged" | "skip" | "optional"
            )
        })
        .map(|s| s["step"].as_str().unwrap().to_string())
        .collect()
}

/// How many steps of the backup phase print a command.
fn backup_commands(plan: &Value) -> usize {
    plan["phases"][0]["steps"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|s| s["apply"].is_object())
        .count()
}

/// A project whose update preview reads `all_unchanged`: a profile, then every printed `ready`
/// command applied; the state lives in the new folder and no leftover exists.
fn settled() -> GoFx {
    let fx = GoFx::new();
    fx.with_profile();
    for step in steps_of(&fx.update_plan())
        .into_iter()
        .filter(|s| s["state"] == "ready")
    {
        let out = fx.apply_step(step);
        assert_eq!(out["ok"], json!(true), "{step}: {out}");
    }
    let plan = fx.update_plan();
    assert_eq!(plan["state_root"]["generation"], json!("rust"), "{plan}");
    assert_eq!(plan["all_unchanged"], json!(true), "{plan}");
    assert_eq!(plan["leftovers"], json!([]), "{plan}");
    fx
}

/// The fixture is sound when no step wants anything: a problem here is the fixture's, not the
/// behaviour under test.
fn assert_steps_settled(plan: &Value) {
    assert_eq!(open_steps(plan), Vec::<String>::new(), "{plan}");
    assert_eq!(backup_commands(plan), 0, "{plan}");
}

fn leftover_paths(plan: &Value) -> Vec<String> {
    let mut paths: Vec<String> = plan["leftovers"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l["path"].as_str().unwrap().to_string())
        .collect();
    paths.sort();
    paths
}

/// The entry for `path`; a missing one is an assertion failure, not a panic in a lookup.
fn entry<'a>(plan: &'a Value, path: &str) -> &'a Value {
    let found = plan["leftovers"]
        .as_array()
        .unwrap()
        .iter()
        .find(|l| l["path"] == path);
    assert!(found.is_some(), "no leftover {path}: {}", plan["leftovers"]);
    found.unwrap()
}

/// An argv value as words; `None` when it is not an array of strings.
fn words(v: &Value) -> Option<Vec<&str>> {
    v.as_array()?.iter().map(Value::as_str).collect()
}

/// The word after `flag`.
fn after<'a>(argv: &[&'a str], flag: &str) -> Option<&'a str> {
    let at = argv.iter().position(|w| *w == flag)?;
    argv.get(at + 1).copied()
}

fn is_sha256(s: &str) -> bool {
    s.strip_prefix("sha256:")
        .is_some_and(|h| h.len() == 64 && h.bytes().all(|b| b.is_ascii_hexdigit()))
}

/// The entry's `keep` argv, checked against the contract; returns its source sha256.
fn keep_sha(fx: &GoFx, l: &Value) -> String {
    let path = l["path"].as_str().unwrap();
    let argv = words(&l["keep"]);
    assert!(argv.is_some(), "{path} prints no keep argv: {l}");
    let argv = argv.unwrap();
    assert_eq!(argv[..3], ["moai", OPERATION_WORD, fx.p()], "{l}");
    assert_eq!(
        after(&argv, "--step").map(String::from),
        Some(format!("leftover:{path}")),
        "{l}"
    );
    let sha = after(&argv, "--source-sha256").unwrap_or_default();
    assert!(is_sha256(sha), "{path}: {l}");
    assert_eq!(
        after(&argv, "--permission-reference"),
        Some(PERMISSION_PLACEHOLDER),
        "{l}"
    );
    sha.to_string()
}

/// The entry's `action` argv, checked against the contract; returns its plan digest.
fn action_digest(fx: &GoFx, l: &Value) -> String {
    let path = l["path"].as_str().unwrap();
    let argv = words(&l["action"]);
    assert!(argv.is_some(), "{path} prints no action: {l}");
    let argv = argv.unwrap();
    assert_eq!(argv[..3], ["moai", APPLY_WORD, fx.p()], "{l}");
    assert_eq!(after(&argv, "--item"), Some(path), "{l}");
    assert_eq!(
        after(&argv, "--permission-reference"),
        Some(PERMISSION_PLACEHOLDER),
        "{l}"
    );
    let digest = after(&argv, "--plan-digest").unwrap_or_default();
    assert!(is_sha256(digest), "{path}: {l}");
    digest.to_string()
}

/// `argv` with the value after `flag` replaced.
fn with(argv: &[&str], flag: &str, value: &str) -> Vec<String> {
    let at = argv.iter().position(|w| *w == flag).expect("the flag");
    let mut out: Vec<String> = argv.iter().map(|w| w.to_string()).collect();
    out[at + 1] = value.to_string();
    out
}

/// Runs a printed command exactly as printed, the placeholder filled with a reference; returns
/// the whole envelope.
fn run_printed<S: AsRef<str>>(fx: &GoFx, argv: &[S]) -> Value {
    assert_eq!(argv[0].as_ref(), "moai");
    let args: Vec<&str> = argv[1..]
        .iter()
        .map(|w| match w.as_ref() {
            PERMISSION_PLACEHOLDER => REF,
            other => other,
        })
        .collect();
    fx.run(&args)
}

/// The `keep` command of `path`'s entry, run as printed; it must succeed.
fn keep(fx: &GoFx, path: &str) {
    let plan = fx.update_plan();
    let l = entry(&plan, path);
    keep_sha(fx, l);
    let out = run_printed(fx, &words(&l["keep"]).unwrap());
    assert_eq!(out["ok"], json!(true), "keep {path}: {out}");
}

/// The five kinds, each in the shape the SPEC's AC-1 names: the earlier folder with a committed
/// receipt, a set-aside folder with a reverted one, the Go backups folder, and a dead lock at
/// each of the two lock paths.
fn add_the_five(fx: &GoFx) {
    earlier_receipt(fx, "committed");
    let folder = aside();
    aside_receipt(fx, folder.trim_end_matches('/'), "reverted");
    fx.write(&format!("{BACKUPS}note.txt"), "an earlier backup\n");
    fx.write(&state_lock(), &lock_body());
    fx.write(&backup_lock(), &lock_body());
}

/// AC-1 (REQ-1, REQ-2, REQ-8): five leftovers, each with an `action` and a `keep`; while an
/// earlier backup step still has a command every leftover prints its `keep` and no `action`;
/// a `CLAUDE.md` leftover is none of the five kinds.
#[test]
fn leftover_plan_gives_each_item_an_action_or_a_keep_only_reason() {
    let fx = settled();
    add_the_five(&fx);
    let idle = fx.update_plan();
    assert_steps_settled(&idle);
    // An earlier backup step with a command: an untracked earlier report in a repository.
    common::git::git_ok(&fx.root, &["init", "-q"]);
    fx.write(
        &format!("{MOAI_DIR}/reports/earlier.md"),
        "an earlier report",
    );
    let waiting = fx.update_plan();
    assert_eq!(
        step_in(&waiting, 0, "reports")["state"],
        "ready",
        "{waiting}"
    );
    let moved = fx.apply_step(step_in(&waiting, 0, "reports"));
    assert_eq!(moved["ok"], json!(true), "{moved}");
    let after_report = fx.update_plan();
    assert_steps_settled(&after_report);
    // A Go curator block in a project root with no managed section.
    let claude = GoFx::new();
    claude.write("CLAUDE.md", CURATOR_BLOCK);
    let claude_plan = claude.update_plan();

    let mut expected = five();
    expected.sort();
    for plan in [&idle, &after_report] {
        assert_eq!(leftover_paths(plan), expected, "{}", plan["leftovers"]);
        let mut digests = Vec::new();
        for path in &expected {
            let l = entry(plan, path);
            assert!(l["fact"].as_str().is_some_and(|s| !s.is_empty()), "{l}");
            assert!(l["note"].as_str().is_some_and(|s| !s.is_empty()), "{l}");
            keep_sha(&fx, l);
            assert_eq!(l["kept"], json!(false), "{l}");
            assert!(l.get("keep_only").is_none(), "{l}");
            digests.push(action_digest(&fx, l));
        }
        digests.dedup();
        assert_eq!(
            digests.len(),
            1,
            "one plan digest for all items: {digests:?}"
        );
        assert_eq!(plan["all_unchanged"], json!(false), "{plan}");
    }
    // While the reports step has a command, no action is printed (that move would make the plan
    // digest of the leftovers stale); the keep stays.
    assert_eq!(
        leftover_paths(&waiting),
        expected,
        "{}",
        waiting["leftovers"]
    );
    for path in &expected {
        let l = entry(&waiting, path);
        keep_sha(&fx, l);
        assert!(l.get("action").is_none(), "{path}: {l}");
    }
    // `CLAUDE.md` is stated as a fact only.
    let l = entry(&claude_plan, "CLAUDE.md");
    for key in ["action", "keep", "keep_only", "kept", "ask_again"] {
        assert!(l.get(key).is_none(), "{key}: {l}");
    }
}

/// AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): an earlier folder with a pending receipt, a set-aside folder
/// with a committed receipt, a folder with a nested checkout and a set-aside folder whose name has
/// no stamp are each listed with `keep_only` and their reason, and no `action`.
#[test]
fn leftover_folder_with_an_unfinished_receipt_is_keep_only() {
    let fx = settled();
    earlier_receipt(&fx, "pending");
    let committed = aside();
    aside_receipt(&fx, committed.trim_end_matches('/'), "committed");
    fx.write(&format!("{BACKUPS}sub/.git/HEAD"), "ref: refs/heads/main\n");
    let notes = format!("{MOAI_DIR}.rust-notes/");
    fx.write(&format!("{notes}todo.txt"), "mine\n");
    let plan = fx.update_plan();
    assert_steps_settled(&plan);

    let cases = [
        (earlier(), "pending"),
        (committed, "committed"),
        (BACKUPS.to_string(), "checkout"),
        (notes, "stamp"),
    ];
    for (path, keyword) in &cases {
        let l = entry(&plan, path);
        let reason = l["keep_only"].as_str().unwrap_or_default();
        assert!(
            reason.to_lowercase().contains(keyword),
            "{path}: want a reason with `{keyword}`: {l}"
        );
        assert!(l.get("action").is_none(), "{path}: {l}");
        keep_sha(&fx, l);
        assert_eq!(l["kept"], json!(false), "{l}");
    }
}

/// AC-7 (REQ-8, REQ-9, REQ-10, REQ-11): an open leftover keeps `all_unchanged` false; recording
/// its keep makes it `kept` and `all_unchanged` true; a keep with a wrong sha256 or for a path
/// that is none of the five kinds is refused with nothing written; a changed top-level name or
/// size asks again while a same-size change of a file's bytes does not; `--ask-again` clears the
/// record.
#[test]
fn leftover_keep_is_recorded_counted_and_asked_again_when_the_folder_changes() {
    let fx = settled();
    let path = earlier();
    fx.write(&format!("{EARLIER_PROJECT_DIR}/a.txt"), "one");
    let open = fx.update_plan();
    assert_steps_settled(&open);
    let l = entry(&open, &path);
    assert_eq!(open["all_unchanged"], json!(false), "{open}");
    assert_eq!(l["kept"], json!(false), "{l}");
    let first = keep_sha(&fx, l);
    let printed = words(&l["keep"]).unwrap();

    // Refused, with the whole tree as it was: a sha256 that is not the source's now, and a path
    // that is none of the five kinds (with the source's real sha256).
    let before = fx.tree();
    let wrong = with(
        &printed,
        "--source-sha256",
        &format!("sha256:{}", "0".repeat(64)),
    );
    let out = run_printed(&fx, &wrong);
    assert_eq!(out["ok"], json!(false), "{out}");
    for other in ["leftover:src/", "leftover:CLAUDE.md", "leftover:"] {
        let out = run_printed(&fx, &with(&printed, "--step", other));
        assert_eq!(out["ok"], json!(false), "{other}: {out}");
    }
    assert_eq!(fx.tree(), before, "a refused keep writes nothing");

    // The right keep is recorded under `leftover:<path>` and counts.
    let out = run_printed(&fx, &printed);
    assert_eq!(out["ok"], json!(true), "{out}");
    let record: Value = serde_json::from_slice(
        &fs::read(state_root::path(&fx.root, "manifests/update-declined.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(
        record["declined"][format!("leftover:{path}")]["source_sha256"],
        json!(first),
        "{record}"
    );
    let done = fx.update_plan();
    let l = entry(&done, &path);
    assert_eq!(l["kept"], json!(true), "{l}");
    assert!(words(&l["ask_again"]).is_some(), "{l}");
    assert_eq!(done["all_unchanged"], json!(true), "{done}");

    // A new top-level name asks again, with another source sha256.
    fx.write(&format!("{EARLIER_PROJECT_DIR}/b.txt"), "two");
    let named = fx.update_plan();
    let l = entry(&named, &path);
    assert_eq!(l["kept"], json!(false), "{l}");
    assert_eq!(named["all_unchanged"], json!(false), "{named}");
    let second = keep_sha(&fx, l);
    assert_ne!(second, first);
    keep(&fx, &path);
    assert_eq!(entry(&fx.update_plan(), &path)["kept"], json!(true));

    // A new size of a top-level file asks again.
    fx.write(&format!("{EARLIER_PROJECT_DIR}/a.txt"), "three");
    let sized = fx.update_plan();
    let l = entry(&sized, &path);
    assert_eq!(l["kept"], json!(false), "{l}");
    assert_ne!(keep_sha(&fx, l), second);
    keep(&fx, &path);
    assert_eq!(entry(&fx.update_plan(), &path)["kept"], json!(true));

    // The sha256 covers names, kinds and sizes only: the same size with other bytes stays kept.
    fx.write(&format!("{EARLIER_PROJECT_DIR}/a.txt"), "THREE");
    let same = fx.update_plan();
    assert_eq!(entry(&same, &path)["kept"], json!(true), "{same}");
    assert_eq!(same["all_unchanged"], json!(true), "{same}");

    // `--ask-again` clears the record and the item is open again.
    let l = entry(&same, &path);
    let ask_again = words(&l["ask_again"]).unwrap();
    assert!(ask_again.contains(&"--ask-again"), "{l}");
    let out = run_printed(&fx, &ask_again);
    assert_eq!(out["ok"], json!(true), "{out}");
    let record: Value = serde_json::from_slice(
        &fs::read(state_root::path(&fx.root, "manifests/update-declined.json")).unwrap(),
    )
    .unwrap();
    assert!(
        record["declined"].get(format!("leftover:{path}")).is_none(),
        "{record}"
    );
    let again = fx.update_plan();
    assert_eq!(entry(&again, &path)["kept"], json!(false), "{again}");
    assert_eq!(again["all_unchanged"], json!(false), "{again}");
}

/// AC-17 (REQ-24): when no step asks, the finish section lists an open leftover in `finish.open`
/// and does not say that nothing is left; once the leftover is kept the finish says it again and
/// lists no leftover.
#[test]
fn leftover_open_item_is_listed_by_the_guide() {
    let fx = settled();
    let path = earlier();
    fx.write(&format!("{EARLIER_PROJECT_DIR}/a.txt"), "one");
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    assert!(plan["ask"].is_null(), "no step asks: {plan}");
    let finish = &plan["finish"];
    assert!(finish.is_object(), "{plan}");
    let listed: Vec<&Value> = finish["open"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|e| e["leftover"] == path.as_str())
        .collect();
    assert_eq!(listed.len(), 1, "{finish}");
    assert!(
        listed[0]["say"].as_str().is_some_and(|s| !s.is_empty()),
        "{finish}"
    );
    assert!(
        !finish["say"]
            .as_str()
            .unwrap_or_default()
            .to_lowercase()
            .contains("nothing is left"),
        "an open leftover is not settled: {finish}"
    );

    keep(&fx, &path);
    let done = fx.update_plan();
    let finish = &done["finish"];
    assert!(
        finish["say"]
            .as_str()
            .unwrap_or_default()
            .to_lowercase()
            .contains("nothing"),
        "{finish}"
    );
    assert!(
        finish["open"]
            .as_array()
            .unwrap()
            .iter()
            .all(|e| e.get("leftover").is_none()),
        "{finish}"
    );
}

// ---- the archive apply and its revert (Plan steps 5 and 6) ----

/// One tree entry as a snapshot sees it: its kind, the sha256 of a file's bytes or the target of a
/// link, and its permission bits. Nothing is followed.
#[derive(Debug, PartialEq, Eq, Clone)]
struct Seen {
    kind: &'static str,
    content: String,
    mode: u32,
}

fn seen(path: &Path) -> Seen {
    let meta = fs::symlink_metadata(path).unwrap();
    let mode = meta.permissions().mode() & 0o7777;
    let (kind, content) = if meta.file_type().is_symlink() {
        ("link", fs::read_link(path).unwrap().display().to_string())
    } else if meta.is_dir() {
        ("dir", String::new())
    } else {
        // A file the test itself cannot read (mode 000) is recorded as such.
        let bytes = fs::read(path).map(|b| sha(&b));
        ("file", bytes.unwrap_or_else(|_| "unreadable".to_string()))
    };
    Seen {
        kind,
        content,
        mode,
    }
}

/// Every entry at or below `top` (`top` itself under the empty key), keyed by its path below
/// `top`; nothing when `top` is not there.
fn snapshot_of(top: &Path) -> BTreeMap<String, Seen> {
    let mut out = BTreeMap::new();
    if fs::symlink_metadata(top).is_err() {
        return out;
    }
    let mut stack = vec![top.to_path_buf()];
    while let Some(p) = stack.pop() {
        let entry = seen(&p);
        if entry.kind == "dir" {
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        }
        out.insert(p.strip_prefix(top).unwrap().display().to_string(), entry);
    }
    out
}

/// The whole project and the whole temporary HOME: paths, kinds, sha256 and modes.
fn snapshot(fx: &GoFx) -> BTreeMap<String, Seen> {
    let mut all = BTreeMap::new();
    for (tag, top) in [("project", &fx.root), ("home", &fx.home)] {
        for (rel, entry) in snapshot_of(top) {
            all.insert(format!("{tag}/{rel}"), entry);
        }
    }
    all
}

fn assert_same(before: &BTreeMap<String, Seen>, after: &BTreeMap<String, Seen>, what: &str) {
    let mut changed: Vec<&String> = before
        .keys()
        .chain(after.keys())
        .filter(|k| before.get(*k) != after.get(*k))
        .collect();
    changed.sort();
    changed.dedup();
    assert!(
        changed.is_empty(),
        "{what}: the tree changed at {:?}",
        &changed[..changed.len().min(10)]
    );
}

fn assert_unchanged(fx: &GoFx, before: &BTreeMap<String, Seen>, what: &str) {
    assert_same(before, &snapshot(fx), what);
}

/// A refusal of a command that exists: not the answer for an unknown command.
fn assert_refused(out: &Value, what: &str) {
    assert_eq!(out["ok"], json!(false), "{what}: {out}");
    let message = out["error"]["message"].as_str().unwrap_or_default();
    assert!(
        !message.contains("unknown command"),
        "{what}: the command is not there yet: {out}"
    );
}

/// The printed apply command for `paths` (one digest for all of them): the `action` of the first
/// path, then one more `--item` for each other path.
fn apply_argv(fx: &GoFx, plan: &Value, paths: &[String]) -> Vec<String> {
    let first_digest = action_digest(fx, entry(plan, &paths[0]));
    for path in &paths[1..] {
        assert_eq!(action_digest(fx, entry(plan, path)), first_digest, "{path}");
    }
    let mut argv: Vec<String> = words(&entry(plan, &paths[0])["action"])
        .unwrap()
        .iter()
        .map(|w| w.to_string())
        .collect();
    for path in &paths[1..] {
        argv.extend(["--item".to_string(), path.clone()]);
    }
    argv
}

/// A printed command without its first word (`moai`), the permission placeholder filled.
fn filled(argv: &[String]) -> Vec<&str> {
    argv.iter()
        .skip(1)
        .map(|w| {
            if w == PERMISSION_PLACEHOLDER {
                REF
            } else {
                w.as_str()
            }
        })
        .collect()
}

/// The outcome word of `path` in a result's `items` (`absent` when it is not listed).
fn outcome(result: &Value, path: &str) -> String {
    result["items"]
        .as_array()
        .into_iter()
        .flatten()
        .find(|i| i["path"] == path)
        .and_then(|i| i["outcome"].as_str())
        .unwrap_or("absent")
        .to_string()
}

/// The receipt name an apply result prints; it is the UTC stamp the archive folder carries.
fn receipt_of(out: &Value) -> String {
    let stamp = out["result"]["receipt"]
        .as_str()
        .unwrap_or_default()
        .to_string();
    assert!(
        state_root::utc_stamp(&stamp),
        "the result names its receipt by a UTC stamp: {out}"
    );
    stamp
}

/// Where the apply puts `item` (a printed path): `backups/<UTC>/leftovers/<same path>` under the
/// state folder.
fn archived(fx: &GoFx, stamp: &str, item: &str) -> PathBuf {
    fx.root.join(format!(
        "{MOAI_DIR}/backups/{stamp}/{ARCHIVE}/{}",
        rel_of(item)
    ))
}

fn receipt_path(fx: &GoFx, stamp: &str) -> PathBuf {
    fx.root
        .join(format!("{MOAI_DIR}/backups/{stamp}/{RECEIPT_FILE}"))
}

fn revert_args<'a>(fx: &'a GoFx, stamp: &'a str) -> Vec<&'a str> {
    vec![
        REVERT_WORD,
        fx.p(),
        "--receipt",
        stamp,
        "--permission-reference",
        REF,
    ]
}

/// The step names of the printed undo order, first row first.
fn undo_rows(plan: &Value) -> Vec<&str> {
    plan["undo_order"]["steps"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s["step"].as_str().unwrap())
        .collect()
}

fn sha_of(path: &Path) -> String {
    sha(&fs::read(path).unwrap())
}

/// AC-3 (REQ-12, REQ-15, REQ-17, REQ-19): one apply moves the five kinds into the archive, byte
/// for byte, and writes a receipt that holds no file's contents and no child name; a second plan
/// shows them gone; the revert puts every byte, mode and listing digest back.
#[test]
fn leftover_apply_moves_and_revert_puts_back_every_byte() {
    let fx = settled();
    add_the_five(&fx);
    let paths = five();
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let argv = apply_argv(&fx, &plan, &paths);
    // What each item is before: its subtree and its source sha256 (the plan's listing digest).
    let places: Vec<BTreeMap<String, Seen>> = paths
        .iter()
        .map(|p| snapshot_of(&fx.root.join(rel_of(p))))
        .collect();
    let digests: Vec<String> = paths
        .iter()
        .map(|p| keep_sha(&fx, entry(&plan, p)))
        .collect();
    let counts: Vec<usize> = paths
        .iter()
        .map(|p| {
            fs::read_dir(fx.root.join(rel_of(p)))
                .map(|d| d.count())
                .unwrap_or(0)
        })
        .collect();
    for (p, d) in paths.iter().zip(&digests).filter(|(p, _)| p.ends_with('/')) {
        assert_eq!(&listing_digest(&fx.root.join(rel_of(p))).unwrap(), d, "{p}");
    }

    let out = run_printed(&fx, &argv);
    assert_eq!(out["ok"], json!(true), "{out}");
    let r = &out["result"];
    assert_eq!(r["state"], json!("committed"), "{r}");
    let stamp = receipt_of(&out);
    for p in &paths {
        assert_eq!(outcome(r, p), "moved", "{p}: {r}");
        assert!(
            fs::symlink_metadata(fx.root.join(rel_of(p))).is_err(),
            "{p} is still at its place"
        );
    }
    for (p, before) in paths.iter().zip(&places) {
        assert_eq!(
            &snapshot_of(&archived(&fx, &stamp, p)),
            before,
            "{p} differs in the archive"
        );
    }

    // The receipt names each item by path, kind, count, size and listing digest only.
    let text = fs::read_to_string(receipt_path(&fx, &stamp)).unwrap();
    let receipt: Value = serde_json::from_str(&text).unwrap();
    assert_eq!(receipt["state"], json!("committed"), "{text}");
    let items = receipt["items"].as_array().cloned().unwrap_or_default();
    assert_eq!(items.len(), paths.len(), "{text}");
    for ((p, d), count) in paths.iter().zip(&digests).zip(&counts) {
        let found = items.iter().find(|i| i["path"] == p.as_str());
        assert!(found.is_some(), "{p} is not in the receipt: {text}");
        let item = found.unwrap();
        assert_eq!(item["listing_digest"], json!(d), "{item}");
        let kind = if p.ends_with('/') { "folder" } else { "file" };
        assert_eq!(item["kind"], json!(kind), "{item}");
        assert!(item["size"].is_u64(), "{item}");
        if p.ends_with('/') {
            assert_eq!(item["entries"], json!(count), "{item}");
        }
    }
    for private in [
        "note.txt",
        "an earlier backup",
        EARLIER_RECEIPT_STAMP,
        ASIDE_RECEIPT_STAMP,
        "started_at",
        "hostname",
    ] {
        assert!(
            !text.contains(private),
            "the receipt holds {private}: {text}"
        );
    }

    // A second plan shows them gone, and nothing is left to ask.
    let after = fx.update_plan();
    assert_eq!(
        leftover_paths(&after),
        Vec::<String>::new(),
        "{}",
        after["leftovers"]
    );
    assert_eq!(after["all_unchanged"], json!(true), "{after}");

    // The revert restores every byte, mode and listing digest.
    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("reverted"), "{back}");
    for p in &paths {
        assert_eq!(outcome(&back["result"], p), "restored", "{p}: {back}");
    }
    for ((p, before), d) in paths.iter().zip(&places).zip(&digests) {
        assert_eq!(
            &snapshot_of(&fx.root.join(rel_of(p))),
            before,
            "{p} is not as it was"
        );
        if p.ends_with('/') {
            assert_eq!(&listing_digest(&fx.root.join(rel_of(p))).unwrap(), d, "{p}");
        }
    }
    let receipt: Value =
        serde_json::from_slice(&fs::read(receipt_path(&fx, &stamp)).unwrap()).unwrap();
    assert_eq!(receipt["state"], json!("reverted"), "{receipt}");
    // A reverted apply is found again by the next plan, with the same source sha256 each.
    let again = fx.update_plan();
    let mut expected = five();
    expected.sort();
    assert_eq!(leftover_paths(&again), expected, "{}", again["leftovers"]);
    for (p, d) in paths.iter().zip(&digests) {
        assert_eq!(&keep_sha(&fx, entry(&again, p)), d, "{p}");
    }
}

/// AC-4 (REQ-13): a made-up digest, a malformed permission reference, an unknown item (alone and
/// beside valid ones) and a digest from before a change each refuse, and the whole tree (paths,
/// kinds, sha256 and modes, project and HOME) is the same before and after.
#[test]
fn leftover_apply_refuses_stale_digest_bad_reference_and_unknown_item() {
    let fx = settled();
    add_the_five(&fx);
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let argv = apply_argv(&fx, &plan, &five());
    let printed: Vec<&str> = argv.iter().map(String::as_str).collect();
    let before = snapshot(&fx);

    // A made-up digest.
    let made_up = run_printed(
        &fx,
        &with(
            &printed,
            "--plan-digest",
            &format!("sha256:{}", "0".repeat(64)),
        ),
    );
    assert_eq!(made_up["ok"], json!(false), "{made_up}");
    assert_eq!(made_up["error"]["kind"], json!("stale_plan"), "{made_up}");
    assert_unchanged(&fx, &before, "a made-up digest");

    // A permission reference the contract refuses.
    for bad in ["placeholder", "null", " ", "ref\nline"] {
        let words: Vec<&str> = printed
            .iter()
            .map(|w| {
                if *w == PERMISSION_PLACEHOLDER {
                    bad
                } else {
                    *w
                }
            })
            .collect();
        let refused = fx.run(&words[1..]);
        assert_refused(&refused, bad);
        assert_unchanged(&fx, &before, bad);
    }

    // An item the plan does not list, alone and beside the valid ones.
    let missing = format!("{MOAI_DIR}.rust-20991231T000000Z/");
    let at = printed.iter().position(|w| *w == "--item").unwrap();
    for item in ["src/", "CLAUDE.md", missing.as_str()] {
        let mut alone: Vec<&str> = printed[..at].to_vec();
        alone.extend(["--item", item]);
        assert_refused(&run_printed(&fx, &alone), item);
        assert_unchanged(&fx, &before, item);
        let mut beside = printed.clone();
        beside.extend(["--item", item]);
        assert_refused(&run_printed(&fx, &beside), item);
        assert_unchanged(&fx, &before, item);
    }

    // The folder changed after the plan was read: its digest is not the fresh plan's.
    fx.write(&format!("{BACKUPS}late.txt"), "added after the plan");
    let before = snapshot(&fx);
    let stale = run_printed(&fx, &printed);
    assert_eq!(stale["ok"], json!(false), "{stale}");
    assert_eq!(stale["error"]["kind"], json!("stale_plan"), "{stale}");
    assert_unchanged(&fx, &before, "a digest from before a change");
}

/// AC-5 (REQ-13, REQ-20): a link item, an absolute path and a `..` are never moved, and a receipt
/// forged to name a place outside the project (through `..`, as an absolute path or through a
/// link) makes the revert report that item `conflict` before it moves anything; the outside files
/// keep their sha256.
#[test]
fn leftover_apply_never_moves_a_link_or_an_escaping_path() {
    let fx = settled();
    earlier_receipt(&fx, "committed");
    let base = fx.root.parent().unwrap().to_path_buf();
    let outside_dir = base.join("outside_dir");
    fs::create_dir_all(&outside_dir).unwrap();
    fs::write(outside_dir.join("y.txt"), "outside y\n").unwrap();
    fs::write(outside_dir.join("lock.json"), lock_body()).unwrap();
    // A linked Go backups folder and a linked lock: neither is a real item, so neither is listed.
    fx.link(rel_of(BACKUPS), "../outside_dir");
    fx.link(&state_lock(), "../../outside_dir/lock.json");
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    assert_eq!(leftover_paths(&plan), [earlier()], "{}", plan["leftovers"]);

    let argv = apply_argv(&fx, &plan, &[earlier()]);
    let printed: Vec<&str> = argv.iter().map(String::as_str).collect();
    let at = printed.iter().position(|w| *w == "--item").unwrap();
    let outside_before = snapshot_of(&outside_dir);
    let outside_sha = sha_of(&outside_dir.join("y.txt"));
    let before = snapshot(&fx);
    let absolute_outside = format!("{}/", outside_dir.display());
    let absolute_inside = format!("{}/{}", fx.p(), earlier());
    let through_the_link = format!("{BACKUPS}y.txt");
    let linked_lock = state_lock();
    let dotdot_inside = format!("{}/../outside_dir/", earlier());
    let hostile: [&str; 8] = [
        BACKUPS,
        rel_of(BACKUPS),
        &through_the_link,
        &linked_lock,
        &absolute_outside,
        &absolute_inside,
        "../outside_dir/",
        &dotdot_inside,
    ];
    for item in hostile {
        let mut alone: Vec<&str> = printed[..at].to_vec();
        alone.extend(["--item", item]);
        assert_refused(&run_printed(&fx, &alone), item);
        let mut beside = printed.clone();
        beside.extend(["--item", item]);
        assert_refused(&run_printed(&fx, &beside), item);
        assert_unchanged(&fx, &before, item);
        assert_eq!(snapshot_of(&outside_dir), outside_before, "{item}");
        assert_eq!(sha_of(&outside_dir.join("y.txt")), outside_sha, "{item}");
    }

    // The real item applies; a receipt forged to name a place outside the project is then refused
    // by the revert, which moves nothing. A file a trusting revert would take from the archive is
    // planted for each forged path.
    let done = run_printed(&fx, &argv);
    assert_eq!(done["ok"], json!(true), "{done}");
    let stamp = receipt_of(&done);
    let receipt_file = receipt_path(&fx, &stamp);
    let original = fs::read_to_string(&receipt_file).unwrap();
    let genuine: Value = serde_json::from_str(&original).unwrap();
    let body = "forged\n";
    let area = fx
        .root
        .join(format!("{MOAI_DIR}/backups/{stamp}/{ARCHIVE}"))
        .display()
        .to_string();
    let forged_paths = [
        "../outside_dir/new.txt".to_string(),
        format!("{}/abs.txt", outside_dir.display()),
        format!("{BACKUPS}new.txt"),
    ];
    for forged in &forged_paths {
        let planted = PathBuf::from(format!("{area}/{forged}"));
        fs::create_dir_all(planted.parent().unwrap()).unwrap();
        fs::write(&planted, body).unwrap();
        let mut item = genuine["items"][0].clone();
        item["path"] = json!(forged);
        item["kind"] = json!("file");
        item["entries"] = json!(0);
        item["size"] = json!(body.len());
        item["listing_digest"] = json!(sha(body.as_bytes()));
        if item.get("moved").is_some() {
            item["moved"] = json!(true);
        }
        let mut receipt = genuine.clone();
        receipt["items"].as_array_mut().unwrap().push(item);
        fs::write(&receipt_file, receipt.to_string()).unwrap();

        let before = snapshot(&fx);
        let outside_before = snapshot_of(&outside_dir);
        let back = fx.run(&revert_args(&fx, &stamp));
        assert_eq!(back["ok"], json!(true), "{forged}: {back}");
        assert_eq!(
            back["result"]["state"],
            json!("conflict"),
            "{forged}: {back}"
        );
        assert_eq!(
            outcome(&back["result"], forged),
            "conflict",
            "{forged}: {back}"
        );
        assert_unchanged(&fx, &before, forged);
        assert_eq!(snapshot_of(&outside_dir), outside_before, "{forged}");
        assert_eq!(sha_of(&outside_dir.join("y.txt")), outside_sha, "{forged}");
    }

    // The genuine receipt reverts: the refusals above were the forged items' doing.
    fs::write(&receipt_file, &original).unwrap();
    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("reverted"), "{back}");
    assert!(fx.root.join(earlier()).is_dir());
}

/// AC-6 (REQ-7, REQ-14): a lock whose process came alive after the plan, a lock whose bytes
/// changed after the plan (its process still dead) and a live lock named beside a valid item are
/// not moved; the unchanged dead lock moves, byte for byte.
#[test]
fn leftover_lock_is_archived_only_while_its_process_is_dead() {
    let fx = settled();
    let lock = state_lock();
    let dead = lock_body();
    fx.write(&lock, &dead);
    // A second item keeps a printed command to copy while the lock is not listed.
    fx.write(&format!("{BACKUPS}note.txt"), "an earlier backup\n");
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let for_lock = apply_argv(&fx, &plan, &[lock.clone()]);

    // The process came alive after the plan.
    let live = json!({"pid": std::process::id(), "started_at": "2026-11-01T00:00:00Z",
                      "hostname": "h"})
    .to_string();
    fx.write(&lock, &live);
    let before = snapshot(&fx);
    assert_refused(&run_printed(&fx, &for_lock), "a live process");
    assert_unchanged(&fx, &before, "a live process");
    // A fresh plan does not list a live lock; naming it beside a valid item refuses the whole.
    let fresh = fx.update_plan();
    assert!(
        !leftover_paths(&fresh).contains(&lock),
        "{}",
        fresh["leftovers"]
    );
    let mut beside = apply_argv(&fx, &fresh, &[BACKUPS.to_string()]);
    beside.extend(["--item".to_string(), lock.clone()]);
    assert_refused(
        &run_printed(&fx, &beside),
        "a live lock beside a valid item",
    );
    assert_unchanged(&fx, &before, "a live lock beside a valid item");

    // The bytes changed after the plan, the process still dead.
    fx.write(&lock, &dead);
    let plan = fx.update_plan();
    let stale = apply_argv(&fx, &plan, &[lock.clone()]);
    let other = json!({"pid": dead_pid(), "started_at": "2026-11-02T00:00:00Z", "hostname": "h"})
        .to_string();
    fx.write(&lock, &other);
    let before = snapshot(&fx);
    let out = run_printed(&fx, &stale);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert_eq!(out["error"]["kind"], json!("stale_plan"), "{out}");
    assert_unchanged(&fx, &before, "changed bytes");

    // Control: the lock as it is now, a fresh plan, moves it.
    let plan = fx.update_plan();
    let done = run_printed(&fx, &apply_argv(&fx, &plan, &[lock.clone()]));
    assert_eq!(done["ok"], json!(true), "{done}");
    let stamp = receipt_of(&done);
    assert_eq!(outcome(&done["result"], &lock), "moved", "{done}");
    assert!(fs::symlink_metadata(fx.root.join(&lock)).is_err());
    assert_eq!(
        fs::read_to_string(archived(&fx, &stamp, &lock)).unwrap(),
        other
    );
}

/// AC-8 (REQ-14): with the simulated cross-device fault the apply refuses before the backup area
/// or the pending receipt exists and moves nothing, and says the check was simulated; without the
/// fault the same command moves both items.
#[test]
fn leftover_apply_stops_before_any_move_on_another_filesystem() {
    let fx = settled();
    fx.write(&format!("{BACKUPS}note.txt"), "an earlier backup\n");
    fx.write(&state_lock(), &lock_body());
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let argv = apply_argv(&fx, &plan, &[BACKUPS.to_string(), state_lock()]);
    let args = filled(&argv);
    let before = snapshot(&fx);

    let out = fx.run_env(&args, &[("MOAI_TEST_FAULT", CROSS_DEVICE)]);
    assert_eq!(out["ok"], json!(false), "{out}");
    assert!(
        out.to_string().to_lowercase().contains("simulated"),
        "the refusal says the check was simulated: {out}"
    );
    assert_unchanged(&fx, &before, "the simulated cross-device check");
    assert!(
        snapshot(&fx).keys().all(|k| !k.ends_with(RECEIPT_FILE)),
        "a pending receipt was written"
    );

    let done = fx.run(&args);
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(outcome(&done["result"], BACKUPS), "moved", "{done}");
    assert_eq!(outcome(&done["result"], &state_lock()), "moved", "{done}");
}

/// AC-9 (REQ-21): an apply that moves no tracked file says so without a warning; one that moves
/// two tracked files names the count and says a checkout brings them back unless the deletions
/// are committed.
#[test]
fn leftover_apply_says_how_many_tracked_files_moved() {
    let fx = settled();
    common::git::git_ok(&fx.root, &["init", "-q"]);
    let (a, b, c) = (
        format!("{BACKUPS}a.txt"),
        format!("{BACKUPS}b.txt"),
        format!("{BACKUPS}c.txt"),
    );
    fx.write(&a, "tracked one");
    fx.write(&b, "tracked two");
    fx.write(&c, "untracked");
    fx.write(&state_lock(), &lock_body());
    common::git::git_ok(&fx.root, &["add", "-f", "--", a.as_str(), b.as_str()]);

    // The lock is not tracked: nothing to warn about.
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let done = run_printed(&fx, &apply_argv(&fx, &plan, &[state_lock()]));
    assert_eq!(done["ok"], json!(true), "{done}");
    assert_eq!(done["result"]["tracked_moved"], json!(0), "{done}");
    assert!(done["result"]["warning"].is_null(), "{done}");

    // Two of the folder's three files are tracked.
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let done = run_printed(&fx, &apply_argv(&fx, &plan, &[BACKUPS.to_string()]));
    assert_eq!(done["ok"], json!(true), "{done}");
    let r = &done["result"];
    assert_eq!(r["tracked_moved"], json!(2), "{r}");
    let warning = r["warning"].as_str().unwrap_or_default();
    assert!(
        warning
            .split(|c: char| !c.is_ascii_digit())
            .any(|t| t == "2"),
        "the warning names the count: {r}"
    );
    assert!(
        warning.contains("checkout") && warning.contains("committed"),
        "the warning says a checkout brings them back unless the deletions are committed: {r}"
    );
}

/// AC-10 (REQ-18): secret-shaped files with mode 000 inside an archived folder move and revert
/// without being opened, are counted, and their names are in neither the result, the receipt nor
/// the revert's result.
#[test]
fn leftover_apply_never_opens_a_secret_shaped_name() {
    let fx = settled();
    fx.write(&format!("{BACKUPS}note.txt"), "an earlier backup\n");
    let secrets: Vec<String> = ["a", "b"]
        .iter()
        .map(|n| format!("{BACKUPS}{SECRET_STEM}-{n}.pem"))
        .collect();
    for s in &secrets {
        fx.write(s, "key material");
    }
    // The plan harness reads every file, so it runs before the files become unreadable.
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let argv = apply_argv(&fx, &plan, &[BACKUPS.to_string()]);
    for s in &secrets {
        fs::set_permissions(fx.root.join(s), fs::Permissions::from_mode(0o000)).unwrap();
    }

    let out = run_printed(&fx, &argv);
    assert_eq!(out["ok"], json!(true), "{out}");
    let stamp = receipt_of(&out);
    assert_eq!(out["result"]["secret_shaped"], json!(2), "{out}");
    let text = fs::read_to_string(receipt_path(&fx, &stamp)).unwrap();
    for shown in [out.to_string(), text] {
        assert!(
            !shown.contains(SECRET_STEM),
            "a secret-shaped name is printed: {shown}"
        );
    }
    let archive = archived(&fx, &stamp, BACKUPS);
    for s in &secrets {
        let name = Path::new(s).file_name().unwrap();
        let mode = fs::symlink_metadata(archive.join(name)).map(|m| m.permissions().mode());
        assert_eq!(
            mode.map(|m| m & 0o777).ok(),
            Some(0),
            "{s} was moved by its name"
        );
    }

    let back = fx.run(&revert_args(&fx, &stamp));
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("reverted"), "{back}");
    assert!(!back.to_string().contains(SECRET_STEM), "{back}");
    for s in &secrets {
        // Back at their places, still unreadable; the mode goes back to a plain one for tidiness
        // (a file of mode 000 is removed with its folder all the same).
        let place = fx.root.join(s);
        assert!(fs::symlink_metadata(&place).is_ok(), "{s} is back");
        fs::set_permissions(&place, fs::Permissions::from_mode(0o600)).unwrap();
    }
}

/// AC-11 (REQ-22): the printed undo order has no `leftovers` row before an apply, starts with it
/// after one (the row runs as printed) and drops it again once the receipt is reverted.
#[test]
fn leftover_undo_order_names_the_revert_only_after_an_apply() {
    let fx = settled();
    fx.write(&format!("{BACKUPS}note.txt"), "an earlier backup\n");
    let waiting = fx.update_plan();
    assert!(
        !undo_rows(&waiting).contains(&"leftovers"),
        "no apply yet, no leftovers row: {}",
        waiting["undo_order"]
    );
    let done = run_printed(&fx, &apply_argv(&fx, &waiting, &[BACKUPS.to_string()]));
    assert_eq!(done["ok"], json!(true), "{done}");
    let stamp = receipt_of(&done);

    let after = fx.update_plan();
    assert_eq!(
        undo_rows(&after).first(),
        Some(&"leftovers"),
        "{}",
        after["undo_order"]
    );
    let commands = after["undo_order"]["steps"][0]["commands"]
        .as_array()
        .unwrap();
    assert_eq!(commands.len(), 1, "{commands:?}");
    let printed = commands[0].as_str().unwrap();
    assert!(
        printed.contains(REVERT_WORD) && printed.contains("<receipt of the leftovers step>"),
        "{printed}"
    );
    let line = printed
        .replace("<receipt of the leftovers step>", &stamp)
        .replace("<reference>", REF)
        .replace("<PROJECT>", fx.p());
    assert!(!line.contains('<'), "a placeholder is left: {line}");
    let args: Vec<&str> = line.split_whitespace().skip(1).collect();
    let back = fx.run(&args);
    assert_eq!(back["ok"], json!(true), "{back}");
    assert_eq!(back["result"]["state"], json!("reverted"), "{back}");

    let reverted = fx.update_plan();
    assert!(
        !undo_rows(&reverted).contains(&"leftovers"),
        "a reverted receipt starts no order: {}",
        reverted["undo_order"]
    );
}

/// AC-18 (REQ-15, REQ-16): with the backup area unwritable the apply moves no item and reports a
/// failure about the receipt or the backup area; the tree is the same before and after; with the
/// area writable again the same command applies.
#[test]
fn leftover_apply_moves_nothing_when_the_receipt_cannot_be_saved() {
    let fx = settled();
    add_the_five(&fx);
    let plan = fx.update_plan();
    assert_steps_settled(&plan);
    let argv = apply_argv(&fx, &plan, &five());
    let backups = fx.root.join(format!("{MOAI_DIR}/backups"));
    fs::create_dir_all(&backups).unwrap();
    let writable = fs::metadata(&backups).unwrap().permissions();
    fs::set_permissions(&backups, fs::Permissions::from_mode(0o555)).unwrap();
    let before = snapshot(&fx);
    let refused = run_printed(&fx, &argv);
    let after = snapshot(&fx);
    // Restore the mode before any assertion, so the temporary folder can clean up.
    fs::set_permissions(&backups, writable).unwrap();
    assert_refused(&refused, "an unwritable backup area");
    let message = refused["error"]["message"]
        .as_str()
        .unwrap_or_default()
        .to_lowercase();
    assert!(
        message.contains("receipt") || message.contains("backup"),
        "the failure names the receipt or the backup area: {refused}"
    );
    assert_same(&before, &after, "an unwritable backup area");

    let done = run_printed(&fx, &argv);
    assert_eq!(done["ok"], json!(true), "{done}");
}
