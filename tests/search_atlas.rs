//! #2183 AC-1: the C10a corpus (#2346) measured through `moai-cli search`.
//!
//! The frozen source snapshot is written into a temporary git project, indexed, and each
//! scored task (retrieval T-*, graph G-*) is searched with its `pattern` as the only literal —
//! no exact ids or paths, so this measures the lexical step alone. The file set of every task
//! must equal the committed rg-only baseline (no regression), nothing may be cut by a cap,
//! and the hit report must pass `tools/atlas_baseline.py --judge` against the committed
//! thresholds.
mod common;

use serde_json::{json, Value};
use std::{
    collections::BTreeSet,
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");

fn fixture(name: &str) -> Value {
    let path = Path::new(MANIFEST).join("tests/fixtures/atlas").join(name);
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

fn git(dir: &Path, args: &[&str]) {
    let ok = common::git::command()
        .args(["-c", "user.name=t", "-c", "user.email=t@example.invalid"])
        .args(args)
        .current_dir(dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .status()
        .unwrap()
        .success();
    assert!(ok, "git {args:?}");
}

fn run(dir: &Path, op: &str, input: &Value) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg(op)
        .arg(dir)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut stdin = child.stdin.take().unwrap();
    match stdin.write_all(input.to_string().as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(stdin);
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{op}: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v.get("result").cloned().unwrap_or(v)
}

#[test]
fn c10a_corpus_through_search_matches_the_rg_baseline_and_passes_the_judge() {
    let snapshot = fixture("source_snapshot.json");
    let corpus = fixture("corpus.json");
    let baseline = fixture("baseline.json");

    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    git(root, &["init", "-q"]);
    for (path, entry) in snapshot["files"].as_object().unwrap() {
        let target = root.join(path);
        fs::create_dir_all(target.parent().unwrap()).unwrap();
        fs::write(&target, entry["content"].as_str().unwrap()).unwrap();
    }
    git(root, &["add", "-A"]);
    git(root, &["commit", "-q", "-m", "frozen snapshot"]);
    run(root, "index-build", &json!({"protocol_version": 1}));

    let expected: std::collections::BTreeMap<String, BTreeSet<String>> = baseline["tasks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| {
            let hits = t["hits"]
                .as_array()
                .unwrap()
                .iter()
                .map(|h| h.as_str().unwrap().to_string())
                .collect();
            (t["id"].as_str().unwrap().to_string(), hits)
        })
        .collect();

    let mut report_tasks = Vec::new();
    for task in corpus["tasks"].as_array().unwrap() {
        if task["kind"] == "freshness" {
            continue;
        }
        let id = task["id"].as_str().unwrap();
        let pattern = task["pattern"].as_str().unwrap();
        let r = run(
            root,
            "search",
            &json!({
                "protocol_version": 1,
                "query": {"literals": [pattern]},
                "limits": {"max_files": 2000, "max_spans_per_file": 1000},
            }),
        );
        assert_eq!(r["coverage"]["file_source"], "index_generation", "{id}");
        assert_eq!(
            r["omitted"],
            json!([]),
            "{id}: a cap or a read cut the answer"
        );
        let mut hits: Vec<String> = Vec::new();
        for h in r["results"].as_array().unwrap() {
            assert_eq!(h["state"], "fresh", "{id}");
            let p = h["path"].as_str().unwrap().to_string();
            if !hits.contains(&p) {
                hits.push(p);
            }
        }
        let got: BTreeSet<String> = hits.iter().cloned().collect();
        assert_eq!(
            Some(&got),
            expected.get(id),
            "{id}: search file set differs from the rg-only baseline"
        );
        report_tasks.push(json!({"id": id, "hits": hits}));
    }
    assert_eq!(
        report_tasks.len(),
        expected.len(),
        "every scored task searched"
    );

    let report = json!({
        "schema_version": baseline["schema_version"],
        "corpus_sha256": baseline["corpus_sha256"],
        "source_snapshot_sha256": baseline["source_snapshot_sha256"],
        "tasks": report_tasks,
    });
    let report_path = root.join("search-report.json");
    fs::write(&report_path, report.to_string()).unwrap();
    let judged = Command::new("python3")
        .arg("-I")
        .arg(Path::new(MANIFEST).join("tools/atlas_baseline.py"))
        .arg("--judge")
        .arg(&report_path)
        .output()
        .expect("spawn the atlas judge");
    assert!(
        judged.status.success(),
        "judge refused the search report:\n{}{}",
        String::from_utf8_lossy(&judged.stdout),
        String::from_utf8_lossy(&judged.stderr)
    );
}
