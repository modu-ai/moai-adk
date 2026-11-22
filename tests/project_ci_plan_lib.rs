//! Plan and apply of the CI files (SPEC-20261122-project-ci-design, AC-3 and AC-4): the library
//! functions behind the later `project-ci-plan` and `project-ci-apply` operations, driven in
//! temporary folders only. The binary is not run here, so no HOME is involved.
//!
//! The pick rule is the apply's digest plus a permission reference; the plan itself writes
//! nothing, an existing file without a baseline is kept as a conflict, and a file the user
//! changed since the last apply is kept too.

use moai_cli::project_ci_plan::{ci_apply, ci_plan, manifest_rel, CiPlan, CiRequest};
use serde_json::{json, Value};
use std::{collections::BTreeMap, fs, path::Path};
use tempfile::TempDir;

const CHECK: &str = ".github/workflows/moai-check.yml";
const NIGHTLY: &str = ".github/workflows/moai-nightly.yml";
const PICK: &str = "fixture-approved";

fn request() -> CiRequest {
    CiRequest {
        protocol_version: 1,
        runs_per_month: Some(40),
        language: None,
        remove: Vec::new(),
    }
}

fn request_removing(paths: &[&str]) -> CiRequest {
    CiRequest {
        remove: paths.iter().map(|p| p.to_string()).collect(),
        ..request()
    }
}

fn profile(flow: &str, systems: Value, visibility: &str, budget: Option<u32>) -> Value {
    let mut ci = json!({
        "kind": "cli",
        "systems": systems,
        "deploy": {"enabled": false},
        "visibility": visibility,
        "strict": "report"
    });
    if let Some(b) = budget {
        ci["monthly_budget_usd"] = json!(b);
    }
    if ci["systems"]
        .as_array()
        .unwrap()
        .iter()
        .any(|s| s == "macos")
    {
        ci["macos_price_ack"] = json!(true);
    }
    json!({
        "schema_version": 2, "project_id": "demo", "purpose": "Build a small tool",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "work_flow": {"flow": flow, "integration_branch": "trunk"},
        "detected": {"project_kind": "existing", "manifests": ["Cargo.toml"],
                     "languages": ["rust"], "lsp": [], "mcp_servers": [], "databases": []},
        "ci": ci,
        "checks": [
            {"id": "build", "command": "cargo build"},
            {"id": "test", "command": "cargo test"}
        ]
    })
}

fn project(profile: &Value) -> TempDir {
    let root = TempDir::new().unwrap();
    set_profile(root.path(), profile);
    root
}

fn set_profile(root: &Path, profile: &Value) {
    let path = root.join(".moai-cli/project.yaml");
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, serde_json::to_vec_pretty(profile).unwrap()).unwrap();
}

fn plan(root: &TempDir) -> CiPlan {
    ci_plan(root.path(), request()).expect("plans")
}

fn apply(root: &TempDir, p: &CiPlan) -> Result<Value, String> {
    ci_apply(root.path(), p.clone(), &p.plan_digest, PICK)
}

fn action(p: &CiPlan, path: &str) -> String {
    p.files
        .iter()
        .find(|f| f.path == path)
        .unwrap_or_else(|| {
            panic!(
                "no {path} in the plan: {:?}",
                p.files.iter().map(|f| &f.path).collect::<Vec<_>>()
            )
        })
        .action
        .clone()
}

/// Every path below `root` with its bytes.
fn tree(root: &Path) -> BTreeMap<String, Vec<u8>> {
    fn walk(dir: &Path, base: &Path, out: &mut BTreeMap<String, Vec<u8>>) {
        for e in fs::read_dir(dir).unwrap().flatten() {
            let path = e.path();
            let rel = path.strip_prefix(base).unwrap().display().to_string();
            if fs::symlink_metadata(&path).unwrap().is_dir() {
                out.insert(format!("{rel}/"), Vec::new());
                walk(&path, base, out);
            } else {
                out.insert(rel, fs::read(&path).unwrap());
            }
        }
    }
    let mut out = BTreeMap::new();
    walk(root, root, &mut out);
    out
}

fn read(root: &TempDir, rel: &str) -> String {
    fs::read_to_string(root.path().join(rel)).unwrap()
}

/// AC-4: the plan writes nothing; the files appear only after the pick, and the second plan
/// has nothing left to do.
#[test]
fn the_plan_writes_nothing_and_the_pick_writes_exactly_the_plan() {
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    let before = tree(root.path());
    let p = plan(&root);
    assert_eq!(tree(root.path()), before, "the plan wrote something");
    assert_eq!(action(&p, CHECK), "create");
    assert!(
        p.files.iter().all(|f| f.path != NIGHTLY),
        "only m-flow has a nightly run"
    );

    let done = apply(&root, &p).expect("applies");
    assert_eq!(done["written"], json!([CHECK]), "{done}");
    assert!(read(&root, CHECK).starts_with("# MoAI generated this file"));
    let manifest: Value = serde_json::from_str(&read(&root, &manifest_rel(root.path()))).unwrap();
    assert_eq!(manifest["owned_files"][0]["path"], CHECK);
    assert_eq!(manifest["owned_files"].as_array().unwrap().len(), 1);

    let again = plan(&root);
    assert_eq!(action(&again, CHECK), "unchanged");
    let after = tree(root.path());
    let second = apply(&root, &again).expect("a second apply is a no-op");
    assert_eq!(second["written"], json!([]), "{second}");
    assert_eq!(tree(root.path()), after, "the no-op apply changed the tree");
}

/// AC-4: an empty or wrong pick writes nothing, and a plan changed after it was shown is stale.
#[test]
fn a_missing_wrong_or_stale_pick_writes_nothing() {
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    let p = plan(&root);
    let before = tree(root.path());
    for (digest, permission) in [
        (p.plan_digest.as_str(), ""),
        (p.plan_digest.as_str(), "null"),
        (p.plan_digest.as_str(), "placeholder"),
        ("sha256:not-the-digest", PICK),
        ("", PICK),
    ] {
        assert!(
            ci_apply(root.path(), p.clone(), digest, permission).is_err(),
            "{digest:?} {permission:?}"
        );
    }
    // A plan whose content was altered after the digest was computed.
    let mut altered = p.clone();
    altered.files[0].rendered_sha256 = "sha256:00".into();
    assert!(ci_apply(root.path(), altered, &p.plan_digest, PICK).is_err());
    // The profile changed after the plan was shown: the plan is stale.
    set_profile(
        root.path(),
        &profile("trunk-based", json!(["linux"]), "public", None),
    );
    assert!(ci_apply(root.path(), p.clone(), &p.plan_digest, PICK).is_err());
    assert!(!root.path().join(CHECK).exists());
    let after = tree(root.path());
    assert_ne!(before, after, "the fixture profile itself changed");
    assert!(after.keys().all(|k| !k.contains("workflows")), "{after:?}");
}

/// Audit B2-N1: a file that appears after the plan was shown makes the plan stale. The apply
/// refuses before it writes anything, so no earlier file of the plan is published and no
/// receipt is left behind (a partial apply was the failure this check prevents).
#[test]
fn a_target_that_appears_after_the_plan_makes_it_stale_and_nothing_is_written() {
    let root = project(&profile("m-flow", json!(["linux"]), "public", None));
    let p = plan(&root);
    assert_eq!(action(&p, CHECK), "create");
    assert_eq!(action(&p, NIGHTLY), "create");
    // The user creates the second planned file after the preview.
    let theirs = root.path().join(NIGHTLY);
    fs::create_dir_all(theirs.parent().unwrap()).unwrap();
    fs::write(&theirs, "name: mine\n").unwrap();
    let before = tree(root.path());
    let err = apply(&root, &p).unwrap_err();
    assert!(err.contains("stale"), "{err}");
    assert_eq!(tree(root.path()), before, "a refused apply wrote something");
    assert!(!root.path().join(CHECK).exists());
    assert_eq!(read(&root, NIGHTLY), "name: mine\n");

    // The same for a planned file the user edited after an earlier apply.
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    apply(&root, &plan(&root)).unwrap();
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux", "windows"]), "public", None),
    );
    let p = plan(&root);
    assert_eq!(action(&p, CHECK), "merge");
    let edited = format!("{}# edited after the preview\n", read(&root, CHECK));
    fs::write(root.path().join(CHECK), &edited).unwrap();
    let before = tree(root.path());
    assert!(apply(&root, &p).unwrap_err().contains("stale"));
    assert_eq!(tree(root.path()), before);
}

/// Audit B2-N2: a file the plan stopped owning (the nightly run after the project left
/// m-flow) stays visible as `stale` and is never touched unless the user picks it.
#[test]
fn a_file_the_plan_no_longer_owns_stays_until_the_user_picks_its_removal() {
    let root = project(&profile("m-flow", json!(["linux"]), "public", None));
    apply(&root, &plan(&root)).unwrap();
    assert!(root.path().join(NIGHTLY).exists());
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux"]), "public", None),
    );
    let p = plan(&root);
    assert_eq!(
        action(&p, CHECK),
        "merge",
        "the trigger changed with the flow"
    );
    assert_eq!(action(&p, NIGHTLY), "stale");
    let stale = p.files.iter().find(|f| f.path == NIGHTLY).unwrap();
    assert!(
        stale.detail.as_deref().unwrap_or("").contains("pick"),
        "the preview says how to remove it: {stale:?}"
    );
    let nightly = read(&root, NIGHTLY);
    let done = apply(&root, &p).unwrap();
    assert_eq!(done["written"], json!([CHECK]), "{done}");
    assert_eq!(done["removed"], json!([]), "{done}");
    assert_eq!(
        read(&root, NIGHTLY),
        nightly,
        "an unpicked stale file was touched"
    );
    // It stays recorded, so the next plan shows it again.
    let manifest: Value = serde_json::from_str(&read(&root, &manifest_rel(root.path()))).unwrap();
    assert!(manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .any(|f| f["path"] == NIGHTLY));
    assert_eq!(action(&plan(&root), NIGHTLY), "stale");
}

/// Audit B2-N2: the user picks the removal; the file goes, the manifest forgets it, the receipt
/// holds the removed bytes and a recovery brings the file back.
#[test]
fn a_picked_removal_deletes_the_unedited_file_and_can_be_recovered() {
    let root = project(&profile("m-flow", json!(["linux"]), "public", None));
    apply(&root, &plan(&root)).unwrap();
    let nightly = read(&root, NIGHTLY);
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux"]), "public", None),
    );
    let p = ci_plan(root.path(), request_removing(&[NIGHTLY])).unwrap();
    assert_eq!(action(&p, NIGHTLY), "remove");
    assert!(root.path().join(NIGHTLY).exists(), "the plan removed it");
    let done = apply(&root, &p).unwrap();
    assert!(!root.path().join(NIGHTLY).exists(), "{done}");
    assert_eq!(done["removed"], json!([NIGHTLY]), "{done}");
    let manifest: Value = serde_json::from_str(&read(&root, &manifest_rel(root.path()))).unwrap();
    assert!(manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["path"] != NIGHTLY));
    assert!(plan(&root).files.iter().all(|f| f.path != NIGHTLY));
    // The receipt can bring it back.
    let receipt = done["request_id"].as_str().unwrap();
    let back = moai_cli::apply::recover(root.path(), receipt, PICK).unwrap();
    assert!(
        back["reverted"]
            .as_array()
            .unwrap()
            .contains(&json!(NIGHTLY)),
        "{back}"
    );
    assert_eq!(read(&root, NIGHTLY), nightly);
}

/// Audit B2-N2: only a file that still equals its recorded baseline can be picked for removal;
/// an edited one, one that is still planned, and any path the plan does not own are refused,
/// and the user's own workflow files are never reachable.
#[test]
fn a_removal_is_refused_for_an_edited_planned_or_unknown_file() {
    let root = project(&profile("m-flow", json!(["linux"]), "public", None));
    apply(&root, &plan(&root)).unwrap();
    fs::write(root.path().join(".github/workflows/ci.yml"), "name: ci\n").unwrap();
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux"]), "public", None),
    );
    for path in [
        CHECK,                           // still planned
        ".github/workflows/ci.yml",      // the user's own file
        ".github/workflows/nothing.yml", // not owned at all
        "../outside.yml",
    ] {
        assert!(
            ci_plan(root.path(), request_removing(&[path])).is_err(),
            "{path} was accepted"
        );
    }
    // An edited stale file is shown as stale and cannot be picked.
    let edited = format!("{}# mine\n", read(&root, NIGHTLY));
    fs::write(root.path().join(NIGHTLY), &edited).unwrap();
    assert_eq!(action(&plan(&root), NIGHTLY), "stale");
    assert!(ci_plan(root.path(), request_removing(&[NIGHTLY])).is_err());
    assert_eq!(read(&root, NIGHTLY), edited);
    // A file edited after the removal was previewed is not deleted by the apply.
    fs::write(
        root.path().join(NIGHTLY),
        read(&root, NIGHTLY).replace("# mine\n", ""),
    )
    .unwrap();
    let p = ci_plan(root.path(), request_removing(&[NIGHTLY])).unwrap();
    assert_eq!(action(&p, NIGHTLY), "remove");
    fs::write(root.path().join(NIGHTLY), "name: mine now\n").unwrap();
    // The re-plan inside the apply refuses a removal of a file that changed since the preview.
    assert!(apply(&root, &p).unwrap_err().contains("cannot remove"));
    assert_eq!(read(&root, NIGHTLY), "name: mine now\n");
    assert_eq!(read(&root, ".github/workflows/ci.yml"), "name: ci\n");
}

/// Audit B2-N2: a stale file the user already deleted is simply forgotten by the manifest.
#[test]
fn a_stale_file_the_user_deleted_is_forgotten() {
    let root = project(&profile("m-flow", json!(["linux"]), "public", None));
    apply(&root, &plan(&root)).unwrap();
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux"]), "public", None),
    );
    fs::remove_file(root.path().join(NIGHTLY)).unwrap();
    let p = plan(&root);
    assert!(p.files.iter().all(|f| f.path != NIGHTLY), "{:?}", p.files);
    apply(&root, &p).unwrap();
    let manifest: Value = serde_json::from_str(&read(&root, &manifest_rel(root.path()))).unwrap();
    assert!(manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["path"] != NIGHTLY));
}

/// AC-4: an existing workflow file with the planned name and no baseline is kept as a
/// conflict, the apply refuses the whole plan, and other workflow files are preserved.
#[test]
fn an_existing_file_without_a_baseline_is_kept_and_other_files_are_preserved() {
    let root = project(&profile("m-flow", json!(["linux"]), "public", None));
    let mine = root.path().join(CHECK);
    fs::create_dir_all(mine.parent().unwrap()).unwrap();
    fs::write(&mine, "name: mine\non: push\n").unwrap();
    fs::write(root.path().join(".github/workflows/ci.yml"), "name: ci\n").unwrap();
    let p = plan(&root);
    assert_eq!(action(&p, CHECK), "conflict");
    assert_eq!(action(&p, NIGHTLY), "create");
    assert!(
        p.preserved
            .contains(&".github/workflows/ci.yml".to_string()),
        "{:?}",
        p.preserved
    );
    let before = tree(root.path());
    let err = apply(&root, &p).unwrap_err();
    assert!(err.contains("conflict"), "{err}");
    assert_eq!(tree(root.path()), before, "a refused apply wrote something");
}

/// AC-4: a generated file the user edited after the apply is kept, shown as a conflict, even
/// when the profile changed; an unedited one is merged to the new rendering.
#[test]
fn an_edited_generated_file_is_kept_and_an_unedited_one_follows_the_profile() {
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    apply(&root, &plan(&root)).unwrap();
    let original = read(&root, CHECK);

    // The profile changes (Windows is named); the file is untouched: merge.
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux", "windows"]), "public", None),
    );
    let p = plan(&root);
    assert_eq!(action(&p, CHECK), "merge");
    let jobs = &p.files.iter().find(|f| f.path == CHECK).unwrap().jobs;
    assert_eq!(
        jobs,
        &vec![
            "check (ubuntu-latest)".to_string(),
            "build-windows (windows-latest)".to_string()
        ]
    );
    apply(&root, &p).unwrap();
    assert!(read(&root, CHECK).contains("build-windows"));
    assert_ne!(read(&root, CHECK), original);

    // The user edits the file; the profile changes again: conflict, nothing overwritten.
    let edited = format!("{}# my own note\n", read(&root, CHECK));
    fs::write(root.path().join(CHECK), &edited).unwrap();
    set_profile(
        root.path(),
        &profile("github-flow", json!(["linux"]), "public", None),
    );
    let p = plan(&root);
    assert_eq!(action(&p, CHECK), "conflict");
    assert!(apply(&root, &p).is_err());
    assert_eq!(read(&root, CHECK), edited);
}

/// An interrupted apply left the file but no baseline: the next plan adopts it instead of
/// reporting a conflict.
#[test]
fn a_file_equal_to_the_rendering_without_a_baseline_is_adopted() {
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    let first = plan(&root);
    apply(&root, &first).unwrap();
    let manifest = root.path().join(manifest_rel(root.path()));
    fs::remove_file(&manifest).unwrap();
    let p = plan(&root);
    assert_eq!(action(&p, CHECK), "unchanged");
    assert!(p.files[0].adopt_baseline);
    apply(&root, &p).unwrap();
    assert!(manifest.exists(), "the baseline was recorded again");
}

/// REQ-4 failure line and REQ-5: no declared checks or no saved answers plans no file and
/// says why in plain words.
#[test]
fn a_project_without_checks_or_answers_plans_no_file() {
    for key in ["checks", "ci"] {
        let mut p = profile("github-flow", json!(["linux"]), "public", None);
        p.as_object_mut().unwrap().remove(key);
        let root = project(&p);
        let planned = plan(&root);
        assert!(planned.files.is_empty(), "{key}");
        assert!(planned.estimate.is_none(), "{key}");
        let said = planned.notes.join(" ");
        assert!(said.contains("nothing was planned"), "{key}: {said}");
        assert!(
            said.contains(if key == "ci" {
                "answers"
            } else {
                "declared check"
            }),
            "{key}: {said}"
        );
        // The apply of an empty plan writes no file and no manifest.
        let before = tree(root.path());
        apply(&root, &planned).unwrap();
        assert_eq!(tree(root.path()), before, "{key}");
    }
}

/// AC-3: the estimate is in the plan: free for a public repository, in dollars with the date
/// of the price table for a private one, and an estimate above the budget is flagged with the
/// cheaper option named first.
#[test]
fn the_plan_carries_the_estimate_and_flags_a_budget_overrun() {
    let public = plan(&project(&profile(
        "m-flow",
        json!(["linux", "windows"]),
        "public",
        None,
    )));
    let e = public.estimate.as_ref().expect("an estimate");
    assert!(e.free && e.mills == 0 && !e.over_budget);

    // 40 runs: Linux 5x40x6 = 1200, Windows 3x40x10 = 1200, nightly 5x30x6 = 900: 3.300 dollars.
    let private = plan(&project(&profile(
        "m-flow",
        json!(["linux", "windows"]),
        "private",
        Some(4),
    )));
    let e = private.estimate.as_ref().unwrap();
    assert_eq!((e.mills, e.minutes), (3300, 470));
    assert_eq!(e.prices_read_on, "2026-11-22");
    assert!(!e.over_budget);
    assert!(
        private
            .notes
            .iter()
            .any(|n| n.contains("3.300") && n.contains("2026-11-22")),
        "{:?}",
        private.notes
    );

    let tight = plan(&project(&profile(
        "m-flow",
        json!(["linux", "windows"]),
        "private",
        Some(3),
    )));
    assert!(tight.estimate.as_ref().unwrap().over_budget);
    let over = tight
        .notes
        .iter()
        .find(|n| n.contains("above your budget"))
        .unwrap_or_else(|| panic!("{:?}", tight.notes));
    assert!(
        over.contains("cheaper") && over.contains("nightly") && over.contains("windows"),
        "{over}"
    );
    // The tool only compares: it names the settings path and sets nothing.
    assert!(
        tight
            .notes
            .iter()
            .any(|n| n.contains("Billing") && n.contains("Budgets")),
        "{:?}",
        tight.notes
    );
}

/// The estimate follows the stated runs per month, with the floor.
#[test]
fn runs_per_month_default_to_the_floor() {
    let root = project(&profile(
        "github-flow",
        json!(["linux"]),
        "private",
        Some(100),
    ));
    let few = ci_plan(
        root.path(),
        CiRequest {
            protocol_version: 1,
            runs_per_month: Some(3),
            language: None,
            remove: Vec::new(),
        },
    )
    .unwrap();
    let none = ci_plan(
        root.path(),
        CiRequest {
            protocol_version: 1,
            runs_per_month: None,
            language: None,
            remove: Vec::new(),
        },
    )
    .unwrap();
    // 5 minutes x 20 runs x 6 mills.
    assert_eq!(few.estimate.as_ref().unwrap().mills, 600);
    assert_eq!(none.estimate.as_ref().unwrap().mills, 600);
}

/// A request the engine does not know is refused; a missing profile is an error, not a guess.
#[test]
fn a_wrong_request_or_a_missing_profile_is_an_error() {
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    assert!(ci_plan(
        root.path(),
        CiRequest {
            protocol_version: 2,
            runs_per_month: None,
            language: None,
            remove: Vec::new(),
        }
    )
    .is_err());
    let empty = TempDir::new().unwrap();
    assert!(ci_plan(empty.path(), request()).is_err());
}

/// The comment language follows the request, and a link in the way is never followed.
#[cfg(unix)]
#[test]
fn the_comment_language_follows_the_request_and_links_are_not_followed() {
    let root = project(&profile("github-flow", json!(["linux"]), "public", None));
    let ko = ci_plan(
        root.path(),
        CiRequest {
            protocol_version: 1,
            runs_per_month: None,
            language: Some("ko".into()),
            remove: Vec::new(),
        },
    )
    .unwrap();
    apply_with(&root, &ko);
    assert!(read(&root, CHECK).contains("MoAI 가 만든 파일이에요"));

    let outside = TempDir::new().unwrap();
    let linked = project(&profile("github-flow", json!(["linux"]), "public", None));
    std::os::unix::fs::symlink(outside.path(), linked.path().join(".github")).unwrap();
    let before = tree(outside.path());
    assert!(ci_plan(linked.path(), request()).is_err() || apply(&linked, &plan(&linked)).is_err());
    assert_eq!(
        tree(outside.path()),
        before,
        "a write went through the link"
    );
}

#[cfg(unix)]
fn apply_with(root: &TempDir, p: &CiPlan) {
    ci_apply(root.path(), p.clone(), &p.plan_digest, PICK).unwrap();
}
