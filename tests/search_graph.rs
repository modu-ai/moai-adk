//! #2408 AC-3: frozen GRAPH cases judged through native search, not atlas-graph.
mod common;

use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    path::Path,
    process::{Command, Stdio},
};

fn fixture(name: &str) -> Value {
    serde_json::from_slice(
        &fs::read(
            Path::new(env!("CARGO_MANIFEST_DIR"))
                .join("tests/fixtures/atlas")
                .join(name),
        )
        .unwrap(),
    )
    .unwrap()
}

fn git(root: &Path, args: &[&str]) {
    let out = common::git::command()
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(args)
        .current_dir(root)
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
}

fn run(root: &Path, op: &str, input: &Value) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg(op)
        .arg(root)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{op}: {}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["ok"], true, "{envelope}");
    envelope["result"].clone()
}

fn frozen() -> (tempfile::TempDir, Value) {
    let snapshot = fixture("source_snapshot.json");
    let dir = tempfile::tempdir().unwrap();
    git(dir.path(), &["init", "-q"]);
    for (path, entry) in snapshot["files"].as_object().unwrap() {
        let dest = dir.path().join(path);
        fs::create_dir_all(dest.parent().unwrap()).unwrap();
        fs::write(dest, entry["content"].as_str().unwrap()).unwrap();
    }
    git(dir.path(), &["add", "."]);
    git(dir.path(), &["commit", "-q", "-m", "frozen snapshot"]);
    let built = run(dir.path(), "index-build", &json!({"protocol_version":1}));
    assert_eq!(built["status"], "complete");
    (dir, snapshot)
}

fn task(id: &str) -> Value {
    fixture("corpus.json")["tasks"]
        .as_array()
        .unwrap()
        .iter()
        .find(|t| t["id"] == id)
        .unwrap()
        .clone()
}

fn kinds(id: &str) -> &'static [&'static str] {
    match id {
        "G-002" => &["documents"],
        "G-003" | "G-007" => &["imports"],
        "G-009" => &["calls", "imports"],
        _ => &["calls"],
    }
}

fn search(root: &Path, source: &str, expanded: bool) -> Value {
    let mut input = json!({"protocol_version":1, "query":{"paths":[source]}});
    if expanded {
        input["limits"] = json!({"max_fanout":200, "max_spans_per_file":1000});
    }
    let r = run(root, "search", &input);
    assert_eq!(r["freshness"], "fresh");
    assert_eq!(r["coverage"]["graph"]["state"], "available");
    assert_eq!(r["coverage"]["graph"]["max_hops"], 1);
    assert_eq!(
        r["coverage"]["graph"]["max_fanout"],
        if expanded { 200 } else { 20 }
    );
    let mut counts = BTreeMap::<&str, usize>::new();
    for hit in r["results"].as_array().unwrap() {
        assert_eq!(hit["state"], "fresh", "{hit}");
        *counts.entry(hit["path"].as_str().unwrap()).or_default() += 1;
    }
    assert!(counts
        .values()
        .all(|n| *n <= if expanded { 1000 } else { 20 }));
    assert_eq!(r["results"][0]["grade"], "exact");
    assert_eq!(r["writes"], false);
    r
}

/// A direct edge's own evidence is required: lexical hits and wider walks cannot pass.
fn found(r: &Value, id: &str, source: &str, target: &str) -> bool {
    r["results"].as_array().unwrap().iter().any(|h| {
        h["grade"] == "graph"
            && h["state"] == "fresh"
            && kinds(id).iter().any(|kind| {
                h["reason"] == format!("{kind}: file:{source} -> file:{target} (hop 1)")
            })
    })
}

fn assert_case(id: &str) {
    let (dir, snapshot) = frozen();
    let t = task(id);
    let relation = t["ground_truth"]["relation"].as_array().unwrap();
    let (source, target) = (relation[0].as_str().unwrap(), relation[2].as_str().unwrap());
    assert!(snapshot["files"].get(target).is_some());
    for expanded in [false, true] {
        let r = search(dir.path(), source, expanded);
        assert!(
            found(&r, id, source, target),
            "{id}: no direct graph citation (expanded={expanded}); omitted={}",
            r["omitted"]
        );
    }
}

macro_rules! graph_case {
    ($name:ident, $id:literal) => {
        #[test]
        fn $name() {
            assert_case($id);
        }
    };
}
graph_case!(g001_search_direct_relation, "G-001");
graph_case!(g002_search_direct_relation, "G-002");
graph_case!(g003_search_direct_relation, "G-003");
graph_case!(g004_search_direct_relation, "G-004");
graph_case!(g005_search_direct_relation, "G-005");
graph_case!(g007_search_direct_relation, "G-007");
graph_case!(g008_search_direct_relation, "G-008");
graph_case!(g009_search_direct_relation, "G-009");

fn assert_matrix(expanded: bool) {
    assert_eq!(moai_cli::search::DEFAULT_MAX_SPANS_PER_FILE, 20);
    assert_eq!(moai_cli::search::DEFAULT_MAX_FANOUT, 20);
    assert_eq!(moai_cli::search::DEFAULT_MAX_HOPS, 1);
    let (dir, snapshot) = frozen();
    let suites = fixture("suites.json");
    let suite = suites["suites"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["id"] == "GRAPH")
        .unwrap();
    assert_eq!(suite["cases"].as_array().unwrap().len(), 9);
    {
        let mut hits = Vec::new();
        let mut gaps = Vec::new();
        for case in suite["cases"].as_array().unwrap() {
            let id = case["task_ref"].as_str().unwrap();
            let t = task(id);
            let rel = t["ground_truth"]["relation"].as_array().unwrap();
            let (source, target) = (rel[0].as_str().unwrap(), rel[2].as_str().unwrap());
            let r = search(dir.path(), source, expanded);
            if found(&r, id, source, target) {
                eprintln!("search GRAPH found {id} expanded={expanded}");
                hits.push(id.to_string());
            } else {
                let reason = if snapshot["files"].get(target).is_none() {
                    "target is not a frozen file; concept identity/alias HOLD".to_string()
                } else {
                    format!("direct relation citation absent; omitted={}", r["omitted"])
                };
                eprintln!("search GRAPH gap {id} expanded={expanded}: {reason}");
                gaps.push(id.to_string());
            }
        }
        eprintln!(
            "search GRAPH expanded={expanded}: {} found; gaps={gaps:?}",
            hits.len()
        );
        assert_eq!(gaps, ["G-006"], "only the approved concept gap may remain");
        assert_eq!(hits.len(), 8);
    }
}

#[test]
fn default_graph_matrix_has_eight_direct_hits_and_only_the_decided_gap() {
    assert_matrix(false);
}

#[test]
fn expanded_graph_matrix_has_eight_direct_hits_and_only_the_decided_gap() {
    assert_matrix(true);
}
