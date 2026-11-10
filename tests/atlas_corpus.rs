//! Fixture-format tests for the C10a eval corpus and the three suite
//! definitions (Issue #2346). REQ-C10-01: every task records its ground truth
//! and its source. AC-3: the FRESHNESS suite is the authority and reconciles
//! with PR #2326's own freshness fixture (five covered cases are marked, three
//! extensions are recorded as gaps).
//!
//! These tests run against the committed fixtures only — no network, no `gh`.

use serde_json::Value;
use std::{collections::BTreeSet, fs, path::PathBuf};

fn fixture(name: &str) -> Value {
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/atlas")
        .join(name);
    let bytes = fs::read(&path).unwrap_or_else(|e| panic!("fixture {name} unreadable: {e}"));
    serde_json::from_slice(&bytes).unwrap_or_else(|e| panic!("fixture {name} is not valid JSON: {e}"))
}

#[test]
fn corpus_has_at_least_thirty_tasks_with_ground_truth_and_source() {
    let corpus = fixture("corpus.json");
    assert_eq!(corpus["schema_version"], 1);
    let tasks = corpus["tasks"].as_array().expect("tasks array");
    assert!(
        tasks.len() >= 30,
        "the corpus needs at least 30 tasks, got {}",
        tasks.len()
    );
    let mut ids = BTreeSet::new();
    for task in tasks {
        let id = task["id"].as_str().expect("task id");
        assert!(ids.insert(id.to_string()), "duplicate task id {id}");
        assert!(
            task["source"].as_str().is_some_and(|s| !s.trim().is_empty()),
            "{id}: the source citation is required (REQ-C10-01)"
        );
        let kind = task["kind"].as_str().expect("task kind");
        if kind == "freshness" {
            // Freshness tasks carry a suite case reference instead of an rg pattern.
            assert!(
                task["suite_ref"].as_str().is_some_and(|s| !s.trim().is_empty()),
                "{id}: a freshness task names its FRESHNESS suite case"
            );
        } else {
            let files = task["ground_truth"]["files"]
                .as_array()
                .expect("ground truth files");
            assert!(!files.is_empty(), "{id}: ground truth files are required");
            assert!(
                task["pattern"].as_str().is_some_and(|s| !s.trim().is_empty()),
                "{id}: the rg pattern is required"
            );
            for f in files {
                let rel = f.as_str().expect("ground truth file path");
                let path = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join(rel);
                assert!(
                    path.exists(),
                    "{id}: ground truth file {rel} does not exist at this tip"
                );
            }
        }
    }
}

#[test]
fn three_suites_are_defined_and_the_freshness_suite_reconciles_with_c6() {
    let suites = fixture("suites.json");
    assert_eq!(suites["schema_version"], 1);
    let suites = suites["suites"].as_array().expect("suites array");
    let ids: Vec<&str> = suites
        .iter()
        .map(|s| s["id"].as_str().expect("suite id"))
        .collect();
    assert!(ids.contains(&"FRESHNESS"), "FRESHNESS suite required");
    assert!(ids.contains(&"GRAPH"), "GRAPH suite required");
    assert!(ids.contains(&"RETRIEVAL"), "RETRIEVAL suite required");

    let freshness = suites.iter().find(|s| s["id"] == "FRESHNESS").unwrap();
    let cases = freshness["cases"].as_array().expect("FRESHNESS cases");
    // The five cases PR #2326 already covers must be marked, so the authority
    // boundary (AC-3) stays visible.
    let covered = cases
        .iter()
        .filter(|c| c["c6_status"] == "covered by PR #2326")
        .count();
    assert!(
        covered >= 5,
        "the five C6-covered cases must be marked c6_status"
    );
    // The design's P0 list adds three cases C6 does not have (rename, two
    // worktrees, scan failure) — recorded as the recorded gaps.
    let gaps = cases
        .iter()
        .filter(|c| c["c6_status"] == "gap: not implemented in C6")
        .count();
    assert!(
        gaps >= 3,
        "rename / two-worktrees / scan-failure must be recorded"
    );

    // Every non-freshness suite case must reference a corpus task that exists.
    let corpus = fixture("corpus.json");
    let task_ids: BTreeSet<String> = corpus["tasks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| t["id"].as_str().expect("task id").to_string())
        .collect();
    for suite in suites {
        if suite["id"] == "FRESHNESS" {
            continue;
        }
        let cases = suite["cases"].as_array().expect("suite cases");
        assert!(!cases.is_empty(), "{} has no cases", suite["id"]);
        for case in cases {
            let task_ref = case["task_ref"].as_str().expect("case task_ref");
            assert!(
                task_ids.contains(task_ref),
                "{} references unknown corpus task {task_ref}",
                suite["id"]
            );
        }
    }
    // Every retrieval and graph task is referenced by its suite, so nothing in
    // the corpus is orphaned from the suites that will run it.
    let referenced: BTreeSet<String> = suites
        .iter()
        .flat_map(|s| s["cases"].as_array().cloned().unwrap_or_default())
        .filter_map(|c| c["task_ref"].as_str().map(str::to_string))
        .collect();
    for task in corpus["tasks"].as_array().unwrap() {
        let kind = task["kind"].as_str().unwrap();
        if kind != "freshness" {
            let id = task["id"].as_str().unwrap();
            assert!(referenced.contains(id), "{id} is not referenced by any suite");
        }
    }
}
