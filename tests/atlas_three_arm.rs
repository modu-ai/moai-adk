//! C10b executes all 42 scored tasks through three cumulative paths.
use serde_json::Value;
use std::{fs, path::Path, process::Command};

#[test]
fn frozen_corpus_three_arm_results_are_scored_and_each_stage_runs() {
    let out = tempfile::TempDir::new().unwrap();
    let run = Command::new("python3")
        .arg("-I")
        // -I excludes the script directory; runpy adds only our checked-in tools.
        .arg("-c")
        .arg("import runpy,sys; sys.path.insert(0,sys.argv.pop(1)); runpy.run_module('atlas_three_arm',run_name='__main__')")
        .arg(Path::new(env!("CARGO_MANIFEST_DIR")).join("tools"))
        .arg("--binary")
        .arg(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("--output-dir")
        .arg(out.path())
        .output()
        .unwrap();
    assert!(
        run.status.success(),
        "{}{}",
        String::from_utf8_lossy(&run.stdout),
        String::from_utf8_lossy(&run.stderr)
    );
    let summary: Value =
        serde_json::from_slice(&fs::read(out.path().join("summary.json")).unwrap()).unwrap();
    let arms = summary["arms"].as_array().unwrap();
    assert_eq!(summary["selection_policy"], "literal-evidence-v1");
    assert_eq!(arms.len(), 3);
    for (i, name) in ["rg", "+Map", "+Graph"].iter().enumerate() {
        let arm = &arms[i];
        assert_eq!(arm["arm"], *name);
        assert_eq!(arm["tasks"], 42);
        assert_eq!(arm["passed"], true);
        assert_eq!(arm["operation_calls"]["search"], 42);
        if i >= 1 {
            assert_eq!(arm["operation_calls"]["atlas-map"], 41);
        }
        if i == 2 {
            assert_eq!(arm["operation_calls"]["atlas-graph"], 41);
        }
        assert_eq!(arm["tool_calls"], 42 + 41 * i);
        assert!(arm["response_bytes"].as_u64().unwrap() > 0);
        assert_eq!(arm["snapshot_stale_ratio"], 0.0);
    }
    let contexts: Value =
        serde_json::from_slice(&fs::read(out.path().join("raw-contexts.json")).unwrap()).unwrap();
    assert_eq!(contexts["generation_id"], summary["generation_id"]);
    let graph: Value =
        serde_json::from_slice(&fs::read(out.path().join("Graph.json")).unwrap()).unwrap();
    let mut raw_files = 0;
    let mut raw_edges = 0;
    let mut map_edges = 0;
    let mut code_edges = 0;
    let mut expanded_tasks = 0;
    for task in graph["report"]["tasks"].as_array().unwrap() {
        let raw = &task["raw_context"]["atlas-graph"];
        if raw.is_null() {
            assert!(task["hits"].as_array().unwrap().is_empty());
            continue;
        }
        let doc = &contexts["documents"][raw["document"].as_str().unwrap()];
        let map_ref = task["raw_context"]["atlas-map"]["document"]
            .as_str()
            .unwrap();
        map_edges += contexts["documents"][map_ref]["edges"]
            .as_array()
            .unwrap()
            .len();
        let nodes = doc["nodes"].as_array().unwrap();
        let edges = doc["edges"].as_array().unwrap();
        let file_count = nodes
            .iter()
            .filter(|n| n["id"].as_str().unwrap().starts_with("file:"))
            .count();
        assert_eq!(file_count, raw["files"].as_array().unwrap().len());
        assert_eq!(nodes.len(), raw["distance"].as_object().unwrap().len());
        raw_files += file_count;
        raw_edges += edges.len();
        code_edges += edges
            .iter()
            .filter(|e| matches!(e["kind"].as_str(), Some("imports" | "calls")))
            .count();
        if file_count > task["hits"].as_array().unwrap().len() {
            expanded_tasks += 1;
        }
        assert_eq!(task["selection"]["policy"], "literal-evidence-v1");
        for answer in task["hits"].as_array().unwrap() {
            assert!(task["selection"]["support"]
                .get(answer.as_str().unwrap())
                .is_some());
        }
    }
    assert_eq!(
        raw_files,
        arms[2]["raw_file_citations"].as_u64().unwrap() as usize
    );
    // Raw graph context is retained even when its unrelated files are not answers.
    assert!(raw_files > arms[2]["citations"].as_u64().unwrap() as usize);
    assert_eq!(
        raw_edges + map_edges,
        arms[2]["view_edges_returned"].as_u64().unwrap() as usize
    );
    assert!(raw_edges > 0 && code_edges > 0 && expanded_tasks > 0);
}

#[test]
fn graph_evidence_changes_selection_not_only_invocation_counts() {
    let run = Command::new("python3")
        .arg("-I")
        .arg(Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/atlas_three_arm_selection_test.py"))
        .output()
        .unwrap();
    assert!(
        run.status.success(),
        "{}{}",
        String::from_utf8_lossy(&run.stdout),
        String::from_utf8_lossy(&run.stderr)
    );
}
