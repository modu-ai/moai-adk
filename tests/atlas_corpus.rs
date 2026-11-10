//! Fixture-format tests for the C10a eval corpus and the three suite
//! definitions (Issue #2346). REQ-C10-01: every task records its ground truth
//! and its source. AC-3: FRESHNESS defines concrete expected observations;
//! these format checks do not execute C6 or C10b freshness behavior.
//!
//! These tests run against the committed fixtures only — no network, no `gh`.

use serde_json::Value;
use std::{collections::BTreeSet, fs, path::PathBuf};

fn fixture(name: &str) -> Value {
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/atlas")
        .join(name);
    let bytes = fs::read(&path).unwrap_or_else(|e| panic!("fixture {name} unreadable: {e}"));
    serde_json::from_slice(&bytes)
        .unwrap_or_else(|e| panic!("fixture {name} is not valid JSON: {e}"))
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
            task["source"]
                .as_str()
                .is_some_and(|s| !s.trim().is_empty()),
            "{id}: the source citation is required (REQ-C10-01)"
        );
        let kind = task["kind"].as_str().expect("task kind");
        if kind == "freshness" {
            // Freshness tasks carry a suite case reference instead of an rg pattern.
            assert!(
                task["suite_ref"]
                    .as_str()
                    .is_some_and(|s| !s.trim().is_empty()),
                "{id}: a freshness task names its FRESHNESS suite case"
            );
        } else {
            let files = task["ground_truth"]["files"]
                .as_array()
                .expect("ground truth files");
            assert!(!files.is_empty(), "{id}: ground truth files are required");
            assert!(
                task["pattern"]
                    .as_str()
                    .is_some_and(|s| !s.trim().is_empty()),
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
fn three_suites_resolve_tasks_and_concrete_freshness_outcomes() {
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

    // Every suite case must resolve to its corpus task, including freshness.
    let corpus = fixture("corpus.json");
    let task_ids: BTreeSet<String> = corpus["tasks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| t["id"].as_str().expect("task id").to_string())
        .collect();
    for suite in suites {
        let cases = suite["cases"].as_array().expect("suite cases");
        assert!(!cases.is_empty(), "{} has no cases", suite["id"]);
        for case in cases {
            let task_ref = case["task_ref"].as_str().expect("case task_ref");
            assert!(
                task_ids.contains(task_ref),
                "{} references unknown corpus task {task_ref}",
                suite["id"]
            );
            if suite["id"] == "FRESHNESS" {
                let case_id = case["case_id"].as_str().expect("freshness case id");
                case["setup"].as_object().expect("concrete freshness setup");
                case["mutation"]
                    .as_object()
                    .expect("concrete freshness mutation");
                let expected = case["expected"]
                    .as_object()
                    .expect("freshness expected outcomes");
                let state = expected["state"]
                    .as_str()
                    .expect("expected freshness state");
                assert!(
                    ["fresh", "stale", "partial", "unavailable"].contains(&state),
                    "{case_id}: unsupported freshness state"
                );
                assert_eq!(
                    expected["generation_unchanged"], true,
                    "{case_id}: read-only query"
                );
                if state == "stale" {
                    let changes = expected["changes"].as_object().expect("observable changes");
                    assert!(
                        changes.values().any(|v| v.as_bool() == Some(true)
                            || v.as_array().is_some_and(|paths| !paths.is_empty())),
                        "{case_id}: stale needs an observable changed path or commit"
                    );
                } else if state == "partial" {
                    assert_eq!(expected["file_exists"], true);
                    assert_eq!(expected["edge_state"], "unavailable");
                    assert_eq!(expected["deleted_files"], serde_json::json!([]));
                }
                if case["mutation"]["operation"] == "modify_worktree" {
                    assert_eq!(expected["other_worktree"]["state"], "fresh");
                    assert_eq!(expected["other_worktree"]["generation_unchanged"], true);
                    assert_ne!(
                        case["mutation"]["worktree"],
                        expected["other_worktree"]["worktree"]
                    );
                }
                let task = corpus["tasks"]
                    .as_array()
                    .unwrap()
                    .iter()
                    .find(|task| task["id"] == task_ref)
                    .unwrap();
                assert_eq!(task["kind"], "freshness");
                assert_eq!(task["suite_ref"], case_id);
                assert_eq!(task["ground_truth"]["expected_state"], expected["state"]);
            }
        }
    }
    // No corpus task may be orphaned from the suite that will run it.
    let referenced: BTreeSet<String> = suites
        .iter()
        .flat_map(|s| s["cases"].as_array().cloned().unwrap_or_default())
        .filter_map(|c| c["task_ref"].as_str().map(str::to_string))
        .collect();
    for task in corpus["tasks"].as_array().unwrap() {
        let id = task["id"].as_str().unwrap();
        assert!(
            referenced.contains(id),
            "{id} is not referenced by any suite"
        );
    }
}
