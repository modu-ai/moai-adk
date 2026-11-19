//! Card C-118 (SPEC-20261119-generated-files-no-lock): a generated shared file
//! (`skills/moai/references/source-manifest.json`) listed in two cards' write scopes no longer
//! locks the second card out; every other overlap is refused as before, and the refusal names
//! its own pair.
//!
//! Every command runs against a temporary HOME and a temporary project; the real `~/.moai` is
//! not read or written. The one repository file read is `.gitattributes` (AC-2).

mod common;

use common::home::TempHome;
use moai_cli::ledger;
use serde_json::{json, Value};
use std::process::Output;

const SESSION_VAR: &str = "CLAUDE_CODE_SESSION_ID";
const S_A: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
const S_B: &str = "bbbbbbbb-0000-4000-8000-00000000000b";
const S_C: &str = "cccccccc-0000-4000-8000-00000000000c";

const MANIFEST: &str = "skills/moai/references/source-manifest.json";
const PLAN_MD: &str = "skills/moai/workflows/plan.md";
const RUN_MD: &str = "skills/moai/workflows/run.md";

fn as_session(home: &TempHome, session: &str, args: &[&str]) -> Output {
    home.command(args)
        .env(SESSION_VAR, session)
        .output()
        .expect("run moai")
}

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON ({e}): {}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn result(out: &Output) -> Value {
    assert_eq!(
        out.status.code(),
        Some(0),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn refusal(out: &Output) -> (String, String) {
    assert_eq!(
        out.status.code(),
        Some(1),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v = envelope(out);
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap().to_string(),
        v["error"]["message"].as_str().unwrap().to_string(),
    )
}

fn registered() -> (TempHome, String) {
    let home = TempHome::new();
    let project = home.project("demo");
    ledger::register(home.path(), "demo").unwrap();
    (home, project.to_str().unwrap().to_string())
}

/// Adds a card with the given scope entries and makes it ready; returns its name.
fn ready_card(home: &TempHome, project: &str, scope: &[&str]) -> String {
    let mut args = vec!["card", "add", "--title", "t"];
    for s in scope {
        args.extend_from_slice(&["--scope", s]);
    }
    args.extend_from_slice(&["--project", project]);
    let added = result(&as_session(home, S_C, &args));
    let name = added["name"].as_str().unwrap().to_string();
    result(&as_session(
        home,
        S_C,
        &["card", "ready", &name, "--project", project],
    ));
    name
}

fn take(home: &TempHome, project: &str, card: &str, session: &str) -> Output {
    as_session(home, session, &["card", "take", card, "--project", project])
}

#[test]
fn two_cards_may_both_list_the_generated_manifest() {
    let (home, project) = registered();
    let a = ready_card(&home, &project, &[PLAN_MD, MANIFEST]);
    let b = ready_card(&home, &project, &[RUN_MD, MANIFEST]);
    result(&take(&home, &project, &a, S_A));
    let taken = result(&take(&home, &project, &b, S_B));
    // The ledger keeps a scope sorted, so the manifest (`references/`) comes before `workflows/`.
    assert_eq!(
        taken["reservation"]["write_scope"],
        json!([MANIFEST, RUN_MD]),
        "{taken}"
    );
    // The manifest alone is no reason either.
    let c = ready_card(&home, &project, &[MANIFEST]);
    result(&take(&home, &project, &c, S_C));
}

#[test]
fn a_real_overlap_is_still_refused_and_names_its_own_pair() {
    let (home, project) = registered();
    let a = ready_card(&home, &project, &[PLAN_MD, MANIFEST]);
    let same_file = ready_card(&home, &project, &[PLAN_MD, MANIFEST]);
    let broad = ready_card(&home, &project, &["skills/moai/**"]);
    result(&take(&home, &project, &a, S_A));

    // The same skill file: refused, and the refusal is about that file, not the manifest.
    let (kind, msg) = refusal(&take(&home, &project, &same_file, S_B));
    assert_eq!(kind, "conflict", "{msg}");
    assert!(msg.contains(PLAN_MD), "{msg}");
    assert!(!msg.contains("source-manifest"), "{msg}");

    // A folder pattern that reaches plan.md: refused for plan.md; the manifest pair never carries it.
    let (kind, msg) = refusal(&take(&home, &project, &broad, S_B));
    assert_eq!(kind, "conflict", "{msg}");
    assert!(msg.contains(PLAN_MD), "{msg}");
    assert!(!msg.contains("source-manifest"), "{msg}");
}

#[test]
fn the_ready_listing_agrees_with_take() {
    let (home, project) = registered();
    let a = ready_card(&home, &project, &[PLAN_MD, MANIFEST]);
    let b = ready_card(&home, &project, &[RUN_MD, MANIFEST]);
    result(&take(&home, &project, &a, S_A));
    // `next` takes the project folder as its last word.
    let listing = result(&as_session(&home, S_B, &["card", "next", &project]));
    assert_eq!(listing["next"], json!(b), "{listing}");
    // The held card is listed as `held`; nothing is skipped for the shared manifest.
    assert_eq!(skipped(&listing), vec![(a.clone(), "held".to_string())]);

    // A card whose scope really overlaps is still skipped, with the overlap as the reason.
    let c = ready_card(&home, &project, &[PLAN_MD]);
    result(&take(&home, &project, &b, S_B));
    let listing = result(&as_session(&home, S_C, &["card", "next", &project]));
    assert_eq!(listing["next"], Value::Null, "{listing}");
    let got = skipped(&listing);
    assert!(
        got.contains(&(c.clone(), "scope_overlap".to_string())),
        "{got:?}"
    );
    assert!(
        !got.iter().any(|(n, r)| n == &b && r == "scope_overlap"),
        "{got:?}"
    );
}

/// `(name, reason)` of every card the `next` listing skipped.
fn skipped(listing: &Value) -> Vec<(String, String)> {
    listing["skipped"]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| {
            (
                s["name"].as_str().unwrap().to_string(),
                s["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn card_check_still_requires_the_path_in_scope() {
    let (home, project) = registered();
    let a = ready_card(&home, &project, &[PLAN_MD, MANIFEST]);
    let d = ready_card(&home, &project, &[RUN_MD]);
    let ta = result(&take(&home, &project, &a, S_A));
    let td = result(&take(&home, &project, &d, S_B));
    let gen = |t: &Value| t["reservation"]["generation"].as_i64().unwrap().to_string();

    let ok = result(&as_session(
        &home,
        S_A,
        &[
            "card",
            "check",
            &a,
            "--generation",
            &gen(&ta),
            "--path",
            MANIFEST,
            "--project",
            &project,
        ],
    ));
    assert_eq!(ok["allowed"], json!(true), "{ok}");

    let (kind, msg) = refusal(&as_session(
        &home,
        S_B,
        &[
            "card",
            "check",
            &d,
            "--generation",
            &gen(&td),
            "--path",
            MANIFEST,
            "--project",
            &project,
        ],
    ));
    assert_eq!(kind, "conflict", "{msg}");
    assert!(msg.contains("not under what"), "{msg}");
}

#[test]
fn the_list_equals_the_gitattributes_merge_lines() {
    let text = std::fs::read_to_string(
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(".gitattributes"),
    )
    .expect(".gitattributes");
    let mut attributed: Vec<&str> = text
        .lines()
        .map(str::trim)
        .filter(|l| !l.starts_with('#') && l.split_whitespace().any(|w| w == "merge=moai-manifest"))
        .filter_map(|l| l.split_whitespace().next())
        .collect();
    attributed.sort_unstable();
    let mut listed: Vec<&str> = moai_cli::scope::GENERATED_SHARED_PATHS.to_vec();
    listed.sort_unstable();
    assert_eq!(listed, attributed);
    assert!(listed.contains(&MANIFEST));
}
