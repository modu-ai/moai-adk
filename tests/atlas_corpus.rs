//! Fixture-format tests for the C10a eval corpus and the three suite
//! definitions (Issue #2346). REQ-C10-01: every task records its ground truth
//! and its source. AC-3: FRESHNESS defines concrete expected observations;
//! these format checks do not execute C6 or C10b freshness behavior.
//!
//! These tests run against the committed fixtures only — no network, no `gh`.

use serde::{de::IgnoredAny, Deserialize};
use serde_json::Value;
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    path::PathBuf,
    sync::LazyLock,
};

fn fixture_bytes(name: &str) -> Vec<u8> {
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/atlas")
        .join(name);
    fs::read(&path).unwrap_or_else(|e| panic!("fixture {name} unreadable: {e}"))
}

fn fixture(name: &str) -> Value {
    serde_json::from_slice(&fixture_bytes(name))
        .unwrap_or_else(|e| panic!("fixture {name} is not valid JSON: {e}"))
}

#[derive(Deserialize)]
struct FrozenSource {
    schema_version: u64,
    // Only membership is needed here. Python verifies the content digests.
    files: BTreeMap<String, IgnoredAny>,
}

fn frozen_source() -> &'static FrozenSource {
    static SOURCE: LazyLock<FrozenSource> = LazyLock::new(|| {
        let source: FrozenSource =
            serde_json::from_slice(&fixture_bytes("source_snapshot.json")).unwrap();
        assert_eq!(source.schema_version, 1);
        source
    });
    &SOURCE
}

fn validate_tasks(tasks: &[Value], files: &BTreeMap<String, IgnoredAny>) -> Result<(), String> {
    if tasks.len() < 30 {
        return Err(format!(
            "the corpus needs at least 30 tasks, got {}",
            tasks.len()
        ));
    }
    let mut ids = BTreeSet::new();
    for task in tasks {
        let id = task["id"].as_str().ok_or("task id is required")?;
        if id.is_empty() || !ids.insert(id) {
            return Err("task ids must be nonempty and unique".into());
        }
        if task["source"].as_str().is_none_or(|s| s.trim().is_empty()) {
            return Err(format!("{id}: source citation is required"));
        }
        match task["kind"].as_str() {
            Some("freshness") => {
                if task["suite_ref"].as_str().is_none_or(|s| s.is_empty()) {
                    return Err(format!("{id}: freshness suite reference is required"));
                }
            }
            Some("retrieval" | "graph") => {
                let truth = task["ground_truth"]["files"]
                    .as_array()
                    .ok_or("ground truth files")?;
                if truth.is_empty() || task["pattern"].as_str().is_none_or(|s| s.is_empty()) {
                    return Err(format!("{id}: ground truth and pattern are required"));
                }
                for file in truth {
                    let path = file.as_str().ok_or("ground truth file path")?;
                    if !files.contains_key(path) {
                        return Err(format!(
                            "{id}: ground truth file {path} is outside frozen source"
                        ));
                    }
                }
            }
            _ => return Err(format!("{id}: unsupported task kind")),
        }
    }
    Ok(())
}

#[test]
fn corpus_has_at_least_thirty_tasks_with_ground_truth_and_source() {
    let corpus = fixture("corpus.json");
    assert_eq!(corpus["schema_version"], 2);
    let tasks = corpus["tasks"].as_array().expect("tasks array");
    validate_tasks(tasks, &frozen_source().files).unwrap();
}

#[test]
fn the_minimum_corpus_boundary_accepts_thirty_and_rejects_twenty_nine() {
    let corpus = fixture("corpus.json");
    let tasks = corpus["tasks"].as_array().unwrap();
    let files = &frozen_source().files;
    assert!(validate_tasks(&tasks[..29], files).is_err());
    assert_eq!(validate_tasks(&tasks[..30], files), Ok(()));
}

#[test]
fn ground_truth_membership_uses_frozen_source_not_live_paths() {
    let mut corpus = fixture("corpus.json");
    let tasks = corpus["tasks"].as_array_mut().unwrap();
    let mut files = frozen_source().files.clone();
    files.insert("src/__c10_frozen_only.rs".into(), IgnoredAny);
    tasks[0]["ground_truth"]["files"] = serde_json::json!(["src/__c10_frozen_only.rs"]);
    assert_eq!(validate_tasks(tasks, &files), Ok(()));
    // This file exists in the checkout, but is not part of the frozen src input.
    tasks[0]["ground_truth"]["files"] = serde_json::json!(["tests/atlas_corpus.rs"]);
    assert!(validate_tasks(tasks, &files).is_err());
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
