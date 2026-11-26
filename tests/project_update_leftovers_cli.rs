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
//! No test here spells the state folder's name: the scan of `state_folder_literal_scan` counts
//! every spelling, so the names come from the `state_root` constants.

#![cfg(unix)]

mod common;
#[path = "common/go_project_real_run.rs"]
mod go_project_real_run;

use go_project_real_run::{step_in, GoFx, REF};
use moai_cli::project_update::{declined::OPERATION_WORD, PERMISSION_PLACEHOLDER};
use moai_cli::state_root::{self, EARLIER_PROJECT_DIR, GO_BACKUP, GO_STATE_RECEIPT, MOAI_DIR};
use serde_json::{json, Value};
use std::fs;

const APPLY_WORD: &str = "project-update-leftovers-apply";
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
