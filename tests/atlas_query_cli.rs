//! #2185 (rc.2 C9) `moai-cli atlas-map` / `atlas-graph` on temporary git projects.
//! AC-1 fixture queries over a built generation (both answers are C4 documents);
//! AC-2 the unavailable path (no generation, a tampered view); stale passes through
//! with its records; no op writes a byte; refusals; and the GRAPH suite (#2346) is
//! answered honestly — a relation the records cannot hold is never reported as found
//! while coverage says the code-reference extractor is unavailable.
mod common;

use moai_cli::ontology::validate_document;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeMap,
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};

const MANIFEST: &str = env!("CARGO_MANIFEST_DIR");
const LIB: &str = "// @MOAI:REQ REQ-C9-01\npub fn pay() {}\n";
const TESTS: &str = "#[test]\nfn req_c9_01_pays_once() {}\n";
const OTHER: &str = "// @MX:NOTE keep\npub fn other() {}\n";

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

fn write(root: &Path, rel: &str, body: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

fn project() -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    git(root, &["init", "-q"]);
    write(root, ".gitignore", ".moai-cli/index/\n");
    write(root, "src/lib.rs", LIB);
    write(root, "src/other.rs", OTHER);
    write(root, "tests/pay.rs", TESTS);
    git(root, &["add", "-A"]);
    git(root, &["commit", "-q", "-m", "base"]);
    dir
}

fn run(dir: &Path, op: &str, input: &Value) -> (i32, Value) {
    let mut child = Command::new(crate::common::MOAI_BIN)
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
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    let v = v.get("result").cloned().unwrap_or(v);
    (out.status.code().unwrap_or(-1), v)
}

fn ok(dir: &Path, op: &str, input: &Value) -> Value {
    let (code, v) = run(dir, op, input);
    assert_eq!(code, 0, "{op}: {v}");
    v
}

fn build(dir: &Path) {
    ok(dir, "index-build", &json!({"protocol_version": 1}));
}

/// Digest of every file under `root` (index directory included), keyed by path.
fn tree(root: &Path) -> BTreeMap<String, String> {
    let mut out = BTreeMap::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for entry in fs::read_dir(&dir).unwrap() {
            let entry = entry.unwrap();
            let path = entry.path();
            let rel = path
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .to_string();
            let meta = fs::symlink_metadata(&path).unwrap();
            if meta.is_dir() {
                out.insert(rel, "dir".into());
                stack.push(path);
            } else {
                // An unreadable file (the partial-build fixture) is compared by size.
                let digest = match fs::read(&path) {
                    Ok(bytes) => format!("{:x}", Sha256::digest(bytes)),
                    Err(_) => format!("unreadable:{}", meta.len()),
                };
                out.insert(rel, digest);
            }
        }
    }
    out
}

/// Runs an op and asserts the project tree is byte-identical afterwards.
fn read_only(dir: &Path, op: &str, input: &Value) -> (i32, Value) {
    let before = tree(dir);
    let answer = run(dir, op, input);
    assert_eq!(before, tree(dir), "{op} wrote into the project");
    answer
}

fn ids(doc: &Value) -> Vec<String> {
    let mut v: Vec<String> = doc["nodes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|n| n["id"].as_str().unwrap().to_string())
        .collect();
    v.sort();
    v
}

#[test]
fn ac1_map_query_returns_the_files_tags_and_edges_as_a_c4_document() {
    let p = project();
    build(p.path());
    let (code, r) = read_only(
        p.path(),
        "atlas-map",
        &json!({"protocol_version": 1, "query": {"paths": ["src/lib.rs"], "prefixes": ["tests"]}}),
    );
    assert_eq!(code, 0, "{r}");
    assert_eq!(r["state"], "fresh");
    assert!(r["generation_id"].as_str().unwrap().starts_with("gen-"));
    validate_document(&r["map"]).unwrap();
    assert_eq!(
        ids(&r["map"]),
        ["file:src/lib.rs", "file:tests/pay.rs", "tag:src/lib.rs#L1"]
    );
    let edge = &r["map"]["edges"][0];
    assert_eq!(edge["kind"], "annotated_by");
    assert_eq!(edge["source_id"], "file:src/lib.rs");
    assert_eq!(edge["target_id"], "tag:src/lib.rs#L1");
    assert_eq!(r["omitted"], json!([]));
    assert_eq!(r["coverage"]["files_in_generation"], 4);
    assert_eq!(r["coverage"]["files_returned"], 2);
    assert!(r["coverage"]["symbols"]
        .as_str()
        .unwrap()
        .starts_with("unavailable"));
    assert_eq!(r["writes"], false);
    assert_eq!(r["network"], "disabled");
}

#[test]
fn ac1_graph_walk_joins_map_and_graph_views_within_the_hop_bound() {
    let p = project();
    build(p.path());
    let start = json!(["req:REQ-C9-01"]);
    let (code, one) = read_only(
        p.path(),
        "atlas-graph",
        &json!({"protocol_version": 1, "query": {"start": start}, "limits": {"max_hops": 1}}),
    );
    assert_eq!(code, 0, "{one}");
    validate_document(&one["graph"]).unwrap();
    assert_eq!(
        ids(&one["graph"]),
        [
            "file:src/lib.rs",
            "req:REQ-C9-01",
            "test:tests/pay.rs::req_c9_01_pays_once"
        ]
    );
    let (_, two) = read_only(
        p.path(),
        "atlas-graph",
        &json!({"protocol_version": 1, "query": {"start": start}, "limits": {"max_hops": 2}}),
    );
    validate_document(&two["graph"]).unwrap();
    // The second hop crosses into the map view: the file's tag.
    assert_eq!(two["distance"]["tag:src/lib.rs#L1"], 2);
    assert_eq!(two["distance"]["req:REQ-C9-01"], 0);
    let kinds: Vec<&str> = two["graph"]["edges"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| e["kind"].as_str().unwrap())
        .collect();
    for k in ["annotated_by", "references_requirement", "verifies"] {
        assert!(kinds.contains(&k), "{k} in {kinds:?}");
    }
    // An unknown start is omitted, not an error.
    let (_, unknown) = read_only(
        p.path(),
        "atlas-graph",
        &json!({"protocol_version": 1, "query": {"start": ["file:nope.rs"]}}),
    );
    assert_eq!(unknown["graph"]["nodes"], json!([]));
    assert_eq!(
        unknown["omitted"],
        json!([{"id": "file:nope.rs", "reason": "start node not in the generation"}])
    );
}

#[test]
fn ac2_without_a_generation_both_ops_answer_unavailable() {
    let p = project();
    for (op, input) in [
        ("atlas-map", json!({"protocol_version": 1})),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": ["file:src/lib.rs"]}}),
        ),
    ] {
        let (code, r) = read_only(p.path(), op, &input);
        assert_eq!(code, 0, "{op}: {r}");
        assert_eq!(r["state"], "unavailable", "{op}");
        assert!(r["reason"].as_str().is_some(), "{op}");
        assert!(r.get("map").is_none() && r.get("graph").is_none(), "{op}");
        assert_eq!(r["writes"], false);
    }
}

#[test]
fn ac2_a_tampered_view_makes_both_ops_unavailable() {
    let p = project();
    build(p.path());
    let gens = p.path().join(".moai-cli/index/generations");
    let gen = fs::read_dir(&gens).unwrap().next().unwrap().unwrap().path();
    let mut bytes = fs::read(gen.join("graph.jsonl")).unwrap();
    bytes.extend_from_slice(b"{\"origin\":null,\"missing\":{\"kind\":\"forged\"}}\n");
    fs::write(gen.join("graph.jsonl"), bytes).unwrap();
    for (op, input) in [
        ("atlas-map", json!({"protocol_version": 1})),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": ["file:src/lib.rs"]}}),
        ),
    ] {
        let (code, r) = read_only(p.path(), op, &input);
        assert_eq!(code, 0, "{op}: {r}");
        assert_eq!(r["state"], "unavailable", "{op}: {r}");
    }
}

#[test]
fn a_stale_index_passes_through_with_its_records() {
    let p = project();
    build(p.path());
    write(p.path(), "src/lib.rs", "pub fn changed() {}\n");
    let (code, r) = read_only(
        p.path(),
        "atlas-map",
        &json!({"protocol_version": 1, "query": {"paths": ["src/lib.rs"]}}),
    );
    assert_eq!(code, 0, "{r}");
    assert_eq!(r["state"], "stale");
    // The records are the published generation's, not the working tree's.
    assert!(ids(&r["map"]).contains(&"tag:src/lib.rs#L1".to_string()));
}

#[cfg(unix)]
fn unreadable(path: &Path, yes: bool) {
    use std::os::unix::fs::PermissionsExt;
    let mut perm = fs::metadata(path).unwrap().permissions();
    perm.set_mode(if yes { 0o000 } else { 0o644 });
    fs::set_permissions(path, perm).unwrap();
}

/// A failed build after a good one leaves the pointer: `partial` passes through with the
/// previous generation's records. A failed first build has no pointer: `unavailable`.
#[cfg(unix)]
#[test]
fn partial_passes_through_with_a_pointer_and_is_unavailable_without_one() {
    let p = project();
    build(p.path());
    let victim = p.path().join("src/other.rs");
    unreadable(&victim, true);
    let (code, _) = run(p.path(), "index-build", &json!({"protocol_version": 1}));
    assert_ne!(code, 0, "the failing build must not succeed");
    let (code, r) = read_only(
        p.path(),
        "atlas-map",
        &json!({"protocol_version": 1, "query": {"paths": ["src/lib.rs"]}}),
    );
    assert_eq!(code, 0, "{r}");
    assert_eq!(r["state"], "partial");
    assert!(ids(&r["map"]).contains(&"tag:src/lib.rs#L1".to_string()));
    let (_, g) = read_only(
        p.path(),
        "atlas-graph",
        &json!({"protocol_version": 1, "query": {"start": ["req:REQ-C9-01"]}}),
    );
    assert_eq!(g["state"], "partial");
    assert!(ids(&g["graph"]).contains(&"file:src/lib.rs".to_string()));
    unreadable(&victim, false);

    let q = project();
    let victim = q.path().join("src/other.rs");
    unreadable(&victim, true);
    let (code, _) = run(q.path(), "index-build", &json!({"protocol_version": 1}));
    assert_ne!(code, 0);
    let (code, r) = read_only(q.path(), "atlas-map", &json!({"protocol_version": 1}));
    assert_eq!(code, 0, "{r}");
    assert_eq!(r["state"], "unavailable");
    assert!(r.get("map").is_none());
    unreadable(&victim, false);
}

#[test]
fn refusals_exit_nonzero_and_limits_have_literal_boundaries() {
    let p = project();
    build(p.path());
    let refused = [
        ("atlas-map", json!({"protocol_version": 2})),
        (
            "atlas-map",
            json!({"protocol_version": 1, "query": {"paths": ["../x"]}}),
        ),
        (
            "atlas-map",
            json!({"protocol_version": 1, "limits": {"max_files": 2001}}),
        ),
        (
            "atlas-map",
            json!({"protocol_version": 1, "limits": {"max_files": 0}}),
        ),
        ("atlas-map", json!({"protocol_version": 1, "bogus": 1})),
        ("atlas-graph", json!({"protocol_version": 1})),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": ["req:REQ-C9-01"]}, "limits": {"max_hops": 4}}),
        ),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": ["req:REQ-C9-01"]}, "limits": {"max_edges": 5001}}),
        ),
    ];
    for (op, input) in refused {
        let (code, r) = read_only(p.path(), op, &input);
        assert_ne!(code, 0, "{op} {input}: {r}");
    }
    let accepted = [
        (
            "atlas-map",
            json!({"protocol_version": 1, "limits": {"max_files": 2000}}),
        ),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": ["req:REQ-C9-01"]}, "limits": {"max_hops": 3, "max_edges": 5000}}),
        ),
    ];
    for (op, input) in accepted {
        let (code, r) = read_only(p.path(), op, &input);
        assert_eq!(code, 0, "{op} {input}: {r}");
    }
    // MAX_TERMS (64) and the 512-byte start id, literal N and N+1.
    let terms = |n: usize| -> Vec<String> { (0..n).map(|i| format!("p{i}.rs")).collect() };
    let starts = |n: usize| -> Vec<String> { (0..n).map(|i| format!("file:p{i}.rs")).collect() };
    let id_of = |len: usize| format!("file:{}", "x".repeat(len - 5));
    assert_eq!(id_of(512).len(), 512);
    let boundaries = [
        (
            "atlas-map",
            json!({"protocol_version": 1, "query": {"paths": terms(64)}}),
            true,
        ),
        (
            "atlas-map",
            json!({"protocol_version": 1, "query": {"paths": terms(65)}}),
            false,
        ),
        (
            "atlas-map",
            json!({"protocol_version": 1, "query": {"paths": terms(32), "prefixes": terms(33)}}),
            false,
        ),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": starts(64)}}),
            true,
        ),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": starts(65)}}),
            false,
        ),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": [id_of(512)]}}),
            true,
        ),
        (
            "atlas-graph",
            json!({"protocol_version": 1, "query": {"start": [id_of(513)]}}),
            false,
        ),
    ];
    for (op, input, accepted) in boundaries {
        let (code, r) = read_only(p.path(), op, &input);
        assert_eq!(code == 0, accepted, "{op}: {r}");
    }
    // max_files 1 keeps one file and names the rest.
    let (_, r) = read_only(
        p.path(),
        "atlas-map",
        &json!({"protocol_version": 1, "limits": {"max_files": 1}}),
    );
    assert_eq!(r["coverage"]["files_returned"], 1);
    assert_eq!(r["omitted"].as_array().unwrap().len(), 3);
}

fn fixture(name: &str) -> Value {
    let path = Path::new(MANIFEST).join("tests/fixtures/atlas").join(name);
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

/// The frozen GRAPH relations require direct evidence, not incidental multi-hop reach.
/// #2450 decision 6793749743 permits typed targets and `documents` edges without changing
/// the corpus, truth or baseline. Non-file concept targets remain unresolved under H1;
/// only G-006 may remain a gap. The seven original code relations retain exact regressions.
#[test]
fn graph_suite_relations_are_found_or_their_gap_is_stated() {
    let (dir, snapshot) = snapshot_project();
    let root = dir.path();
    let corpus = fixture("corpus.json");
    let suites = fixture("suites.json");
    let tasks: BTreeMap<&str, &Value> = corpus["tasks"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| (t["id"].as_str().unwrap(), t))
        .collect();
    let graph = suites["suites"]
        .as_array()
        .unwrap()
        .iter()
        .find(|s| s["id"] == "GRAPH")
        .unwrap();
    let mut found = Vec::new();
    let mut gaps: BTreeMap<String, String> = BTreeMap::new();
    for case in graph["cases"].as_array().unwrap() {
        let id = case["task_ref"].as_str().unwrap();
        let rel = tasks[id]["ground_truth"]["relation"].as_array().unwrap();
        let (source, target) = (rel[0].as_str().unwrap(), rel[2].as_str().unwrap());
        let from = format!("file:{source}");
        let to = snapshot["files"]
            .get(target)
            .map(|_| format!("file:{target}"));
        let r = ok(
            root,
            "atlas-graph",
            &json!({"protocol_version": 1, "query": {"start": [from]}, "limits": {"max_hops": 3, "max_edges": 5000}}),
        );
        assert_eq!(r["state"], "fresh", "{id}");
        validate_document(&r["graph"]).unwrap();
        // The start file is in the generation: a miss below is a missing relation,
        // not an unknown start.
        assert_eq!(
            r["distance"][&from], 0,
            "{id}: {from} not in the generation"
        );
        let direct: Vec<String> = to
            .as_ref()
            .map(|_| direct_edges(root, source, target))
            .unwrap_or_default()
            .into_iter()
            .filter(|e| {
                e.starts_with(&format!("imports {from} "))
                    || e.starts_with(&format!("calls {from} "))
                    || e.starts_with(&format!("documents {from} "))
            })
            .collect();
        if !direct.is_empty() {
            found.push(id.to_string());
            continue;
        }
        let reason = if snapshot["files"].get(target).is_none() {
            format!("target {target:?} is not a file of the snapshot")
        } else {
            let wider = match to.as_ref().and_then(|to| r["distance"].get(to)) {
                Some(d) => format!("; reached only through other files at hop {d}"),
                None => String::new(),
            };
            format!("no direct imports/calls/documents edge from {source} to {target}{wider}")
        };
        gaps.insert(id.to_string(), reason);
    }
    for (id, reason) in &gaps {
        assert!(!reason.trim().is_empty(), "{id}: a gap needs its reason");
        eprintln!("GRAPH gap {id}: {reason}");
    }
    eprintln!(
        "GRAPH suite through atlas-graph: {}/{} found by a direct edge, {} gaps stated",
        found.len(),
        found.len() + gaps.len(),
        gaps.len()
    );
    assert_eq!(found.len() + gaps.len(), 9, "every GRAPH case queried");
    assert_eq!(
        gaps.keys().map(String::as_str).collect::<Vec<_>>(),
        ["G-006"],
        "only the concept identity/alias HOLD may remain (decision 6793749743)"
    );
    assert_eq!(found.len(), 8);
}

/// Builds the frozen GRAPH snapshot (#2346) as a committed repository with an index.
fn snapshot_project() -> (tempfile::TempDir, Value) {
    let snapshot = fixture("source_snapshot.json");
    let dir = tempfile::TempDir::new().unwrap();
    let root = dir.path();
    git(root, &["init", "-q"]);
    write(root, ".gitignore", ".moai-cli/index/\n");
    for (path, entry) in snapshot["files"].as_object().unwrap() {
        write(root, path, entry["content"].as_str().unwrap());
    }
    git(root, &["add", "-A"]);
    git(root, &["commit", "-q", "-m", "frozen snapshot"]);
    build(root);
    (dir, snapshot)
}

/// The direct code-reference edges (#2442) between the two files of each GRAPH case:
/// `kind source target L<lines>` for every edge whose endpoints are exactly the pair,
/// read from a one-hop walk so a dense graph cannot reach the target by accident.
fn direct_edges(root: &Path, from: &str, to: &str) -> Vec<String> {
    let r = ok(
        root,
        "atlas-graph",
        &json!({"protocol_version": 1, "query": {"start": [format!("file:{from}")]}, "limits": {"max_hops": 1, "max_edges": 5000}}),
    );
    let mut out = Vec::new();
    for e in r["graph"]["edges"].as_array().unwrap() {
        let (s, t) = (
            e["source_id"].as_str().unwrap(),
            e["target_id"].as_str().unwrap(),
        );
        let pair = [format!("file:{from}"), format!("file:{to}")];
        if !(pair.contains(&s.to_string()) && pair.contains(&t.to_string())) {
            continue;
        }
        let lines: Vec<String> = e["evidence_refs"]
            .as_array()
            .unwrap()
            .iter()
            .map(|r| {
                let r = r.as_str().unwrap();
                r[r.find('#').unwrap() + 1..r.find('@').unwrap()].to_string()
            })
            .collect();
        out.push(format!(
            "{} {} {} {} {}",
            e["kind"].as_str().unwrap(),
            s,
            t,
            e["basis"].as_str().unwrap(),
            lines.join(",")
        ));
    }
    out.sort();
    out
}

/// Per-case evidence for the GRAPH suite (#2442 AC-3): the direct edge, its direction,
/// basis and evidence lines, measured on the frozen snapshot. The suite walk is three hops
/// in both directions, so reaching a target there does not prove a direct reference.
/// G-008/G-009 retain their inferred library crate names. G-002 is a documents edge
/// with both the mention and definition evidence; it is never a call. G-006 retains H1.
#[test]
fn graph_cases_have_direct_code_reference_edges() {
    let (dir, _) = snapshot_project();
    let corpus = fixture("corpus.json");
    let want: BTreeMap<&str, Vec<&str>> = BTreeMap::from([
        (
            "G-001",
            vec!["calls file:src/archive.rs file:src/profile.rs static_resolved L136"],
        ),
        (
            "G-002",
            vec!["documents file:src/pr_link.rs file:src/forge/outbox.rs inferred L150,L11"],
        ),
        (
            "G-003",
            vec!["imports file:src/handoff.rs file:src/forge/outbox.rs static_resolved L28"],
        ),
        (
            "G-004",
            vec!["calls file:src/decision.rs file:src/lib.rs static_resolved L190"],
        ),
        (
            "G-005",
            vec!["calls file:src/audit.rs file:src/baseline.rs static_resolved L257"],
        ),
        ("G-006", vec![]),
        (
            "G-007",
            vec!["imports file:src/flow.rs file:src/forge/transport.rs static_resolved L8"],
        ),
        (
            "G-008",
            vec!["calls file:src/main.rs file:src/atlas_tag.rs inferred L1426"],
        ),
        (
            "G-009",
            vec![
                "calls file:src/main.rs file:src/contract.rs inferred L1688,L1714,L1720",
                "imports file:src/main.rs file:src/contract.rs inferred L3",
            ],
        ),
    ]);
    let mut seen = 0;
    for t in corpus["tasks"].as_array().unwrap() {
        if t["kind"] != "graph" {
            continue;
        }
        let id = t["id"].as_str().unwrap();
        let rel = t["ground_truth"]["relation"].as_array().unwrap();
        let got = direct_edges(
            dir.path(),
            rel[0].as_str().unwrap(),
            rel[2].as_str().unwrap(),
        );
        assert_eq!(got, want[id], "{id}");
        seen += 1;
    }
    assert_eq!(seen, 9);
}

#[test]
fn narrow_frozen_doc_exclusion_counts_are_visible_without_excluding_g002() {
    let (dir, snapshot) = snapshot_project();
    let got = moai_cli::index::index_records(
        dir.path(),
        &moai_cli::index::IndexQueryRequest {
            protocol_version: 1,
        },
    )
    .unwrap();
    assert_eq!(got["state"], "fresh");
    let excluded: Vec<&Value> = got["graph"]["missing"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|m| m["kind"] == "doc_reference_unsupported")
        .collect();
    let count: u64 = excluded.iter().map(|m| m["count"].as_u64().unwrap()).sum();
    assert!(
        count > 0,
        "unsupported blocks must remain visible as coverage gaps"
    );
    assert!(excluded
        .iter()
        .all(|m| m["count"] == 1 && m["reason"].as_str().is_some_and(|s| !s.is_empty())));
    let pr_link = snapshot["files"]["src/pr_link.rs"]["content"]
        .as_str()
        .unwrap();
    assert!(
        moai_cli::code_refs::scan(pr_link)
            .excluded_doc_blocks
            .is_empty(),
        "G-002 must remain in a supported plain block"
    );
    assert!(excluded.iter().all(|m| m["path"] != "src/pr_link.rs"));
    eprintln!(
        "NARROW_FROZEN_EXCLUDED_BLOCKS: {count}; diagnostic entries: {}",
        serde_json::to_string(&excluded).unwrap()
    );
}
