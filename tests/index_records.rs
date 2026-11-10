//! #2347 (rc.2 C6b) Map/Graph records in an index generation. Each fixture is a
//! temporary git repository driven through the library build and query:
//! AC-1 both views exist in one generation and pass the C4 schema; AC-2 an incremental
//! build equals a full rebuild byte for byte (and really reused stored fragments);
//! AC-3 Map and Graph come from the same generation and file set, and a view that does
//! not match its manifest makes the records query `unavailable`.
use moai_cli::index::{index_build, index_records, IndexBuildRequest, IndexQueryRequest};
use moai_cli::ontology::validate_document;
use serde_json::Value;
use std::{
    collections::BTreeSet,
    fs,
    path::{Path, PathBuf},
    process::Command,
};

fn git(dir: &Path, args: &[&str]) {
    let status = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.com")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.com")
        .env("GIT_AUTHOR_DATE", "2026-11-10T01:02:03Z")
        .env("GIT_COMMITTER_DATE", "2026-11-10T01:02:03Z")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?} failed");
}

fn write(root: &Path, rel: &str, body: &str) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

fn commit(root: &Path, message: &str) {
    git(root, &["add", "-A"]);
    git(root, &["commit", "-q", "--allow-empty", "-m", message]);
}

const LIB: &str = "// @MOAI:REQ REQ-C6B-01\npub fn pay() {}\n";
const TESTS: &str = "#[test]\nfn req_c6b_01_pays_once() {}\n";
const OTHER: &str = "// @MOAI:REQ REQ-C6B-02\n// @MX:NOTE keep\npub fn other() {}\n";

fn setup() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("p");
    fs::create_dir_all(&root).unwrap();
    git(&root, &["init", "-q", "-b", "main"]);
    write(&root, ".gitignore", ".moai-cli/index/\n");
    write(&root, "src/lib.rs", LIB);
    write(&root, "src/other.rs", OTHER);
    write(&root, "tests/pay.rs", TESTS);
    write(&root, "README.md", "plain\n");
    commit(&root, "first\n\nRefs: #2347");
    (tmp, root)
}

fn build(root: &Path) -> Value {
    index_build(
        root,
        &IndexBuildRequest {
            protocol_version: 1,
        },
    )
    .unwrap()
}

fn records(root: &Path) -> Value {
    index_records(
        root,
        &IndexQueryRequest {
            protocol_version: 1,
        },
    )
    .unwrap()
}

fn gen_dir(root: &Path, gen: &str) -> PathBuf {
    root.join(".moai-cli/index/generations").join(gen)
}

fn edges(doc: &Value) -> BTreeSet<(String, String, String)> {
    doc["edges"]
        .as_array()
        .unwrap()
        .iter()
        .map(|e| {
            (
                e["kind"].as_str().unwrap().to_string(),
                e["source_id"].as_str().unwrap().to_string(),
                e["target_id"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

fn t(kind: &str, source: &str, target: &str) -> (String, String, String) {
    (kind.into(), source.into(), target.into())
}

fn view_bytes(root: &Path, gen: &str) -> [Vec<u8>; 2] {
    let dir = gen_dir(root, gen);
    [
        fs::read(dir.join("map.jsonl")).unwrap(),
        fs::read(dir.join("graph.jsonl")).unwrap(),
    ]
}

#[test]
fn one_generation_holds_schema_valid_map_and_graph_records() {
    let (_tmp, root) = setup();
    let built = build(&root);
    let gen = built["generation_id"].as_str().unwrap();
    assert_eq!(built["records"]["extracted_files"], 5);
    assert_eq!(built["records"]["reused_files"], 0);
    let dir = gen_dir(&root, gen);
    for name in ["manifest.json", "map.jsonl", "graph.jsonl"] {
        assert!(dir.join(name).is_file(), "{name}");
    }

    let q = records(&root);
    assert_eq!(q["state"], "fresh");
    assert_eq!(q["generation_id"], gen);
    for view in ["map", "graph"] {
        validate_document(&q[view]).unwrap();
    }
    assert_eq!(
        edges(&q["map"]),
        BTreeSet::from([
            t("annotated_by", "file:src/lib.rs", "tag:src/lib.rs#L1"),
            t("annotated_by", "file:src/other.rs", "tag:src/other.rs#L1"),
            t("annotated_by", "file:src/other.rs", "tag:src/other.rs#L2"),
        ])
    );
    let head = String::from_utf8(
        Command::new("git")
            .args(["rev-parse", "HEAD"])
            .current_dir(&root)
            .output()
            .unwrap()
            .stdout,
    )
    .unwrap();
    let head = head.trim();
    assert_eq!(
        edges(&q["graph"]),
        BTreeSet::from([
            t(
                "references_requirement",
                "file:src/lib.rs",
                "req:REQ-C6B-01"
            ),
            t(
                "references_requirement",
                "file:src/other.rs",
                "req:REQ-C6B-02"
            ),
            t(
                "verifies",
                "test:tests/pay.rs::req_c6b_01_pays_once",
                "req:REQ-C6B-01"
            ),
            t("mentions_issue", &format!("commit:{head}"), "issue:2347"),
        ])
    );
    // Time is content: every edge carries the base commit time and snapshot.
    for view in ["map", "graph"] {
        for e in q[view]["edges"].as_array().unwrap() {
            assert_eq!(e["observed_at"], "2026-11-10T01:02:03Z");
            assert_eq!(e["snapshot_id"], head);
        }
    }
}

#[test]
fn incremental_build_equals_a_full_rebuild() {
    let (_tmp, root) = setup();
    build(&root);
    // Modify, add, delete the only file that references REQ-C6B-02, change an
    // untouched file's neighbour set, and move HEAD with a new commit.
    write(
        &root,
        "src/lib.rs",
        "// @MOAI:REQ REQ-C6B-01, REQ-C6B-03\npub fn pay() {}\n",
    );
    write(
        &root,
        "src/new.rs",
        "// @MX:TODO later\n#[test]\nfn req_c6b_03_new() {}\n",
    );
    fs::remove_file(root.join("src/other.rs")).unwrap();
    commit(&root, "second\n\nRefs: #2348");

    let incremental = build(&root);
    let gen = incremental["generation_id"].as_str().unwrap().to_string();
    assert_eq!(incremental["action"], "created");
    // Three unchanged entries (.gitignore, README.md, tests/pay.rs) reuse stored
    // fragments; the modified and the added file are extracted.
    assert_eq!(incremental["records"]["reused_files"], 3);
    assert_eq!(incremental["records"]["extracted_files"], 2);
    let incremental_views = view_bytes(&root, &gen);
    let incremental_manifest: Value =
        serde_json::from_slice(&fs::read(gen_dir(&root, &gen).join("manifest.json")).unwrap())
            .unwrap();

    fs::remove_dir_all(root.join(".moai-cli/index")).unwrap();
    let full = build(&root);
    assert_eq!(full["generation_id"], gen.as_str());
    assert_eq!(full["records"]["reused_files"], 0);
    assert_eq!(full["records"]["extracted_files"], 5);
    let full_views = view_bytes(&root, &gen);
    assert_eq!(incremental_views[0], full_views[0], "map.jsonl differs");
    assert_eq!(incremental_views[1], full_views[1], "graph.jsonl differs");
    let full_manifest: Value =
        serde_json::from_slice(&fs::read(gen_dir(&root, &gen).join("manifest.json")).unwrap())
            .unwrap();
    assert_eq!(incremental_manifest["records"], full_manifest["records"]);

    // The deleted file's requirement is gone from both views, the new one is present.
    let q = records(&root);
    let text = serde_json::to_string(&q["graph"]).unwrap();
    assert!(!text.contains("REQ-C6B-02"), "stale requirement kept");
    assert!(text.contains("req:REQ-C6B-03"));
    assert!(!serde_json::to_string(&q["map"])
        .unwrap()
        .contains("src/other.rs"));
}

#[test]
fn map_and_graph_share_the_generation_and_file_set() {
    let (_tmp, root) = setup();
    let built = build(&root);
    let gen = built["generation_id"].as_str().unwrap();
    let manifest: Value =
        serde_json::from_slice(&fs::read(gen_dir(&root, gen).join("manifest.json")).unwrap())
            .unwrap();
    let paths: BTreeSet<String> = manifest["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| format!("file:{}", f["path"].as_str().unwrap()))
        .collect();
    let mut headers = Vec::new();
    for name in ["map.jsonl", "graph.jsonl"] {
        let bytes = fs::read(gen_dir(&root, gen).join(name)).unwrap();
        let first: Value =
            serde_json::from_str(std::str::from_utf8(&bytes).unwrap().lines().next().unwrap())
                .unwrap();
        headers.push(first["header"].clone());
    }
    for (h, view) in headers.iter().zip(["map", "graph"]) {
        assert_eq!(h["view"], view);
        assert_eq!(h["generation_id"], gen);
        assert_eq!(h["base_commit"], manifest["base_commit"]);
        assert_eq!(h["file_set"], manifest["records"]["file_set"]);
    }
    // Every Map file node is a manifest entry and every manifest entry is on the Map.
    let q = records(&root);
    let map_files: BTreeSet<String> = q["map"]["nodes"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|n| n["kind"] == "file")
        .map(|n| n["id"].as_str().unwrap().to_string())
        .collect();
    assert_eq!(map_files, paths);
    // Every Graph file or test node names a manifest path.
    for n in q["graph"]["nodes"].as_array().unwrap() {
        if let Some(p) = n["properties"]["path"].as_str() {
            assert!(paths.contains(&format!("file:{p}")), "{p}");
        }
    }
}

#[test]
fn a_view_that_does_not_match_its_manifest_is_unavailable() {
    for (view, tamper) in [
        ("map.jsonl", "append"),
        ("graph.jsonl", "append"),
        ("graph.jsonl", "remove"),
    ] {
        let (_tmp, root) = setup();
        let built = build(&root);
        let gen = built["generation_id"].as_str().unwrap();
        let path = gen_dir(&root, gen).join(view);
        if tamper == "append" {
            let mut bytes = fs::read(&path).unwrap();
            bytes.extend_from_slice(b"{\"origin\":null,\"missing\":{\"kind\":\"x\"}}\n");
            fs::write(&path, bytes).unwrap();
        } else {
            fs::remove_file(&path).unwrap();
        }
        let q = records(&root);
        assert_eq!(q["state"], "unavailable", "{view} {tamper}");
        assert!(
            q["map"].is_null() && q["graph"].is_null(),
            "{view} {tamper}"
        );
        // A damaged view is never reused by the next build of the same snapshot.
        let err = index_build(
            &root,
            &IndexBuildRequest {
                protocol_version: 1,
            },
        )
        .unwrap_err();
        assert!(err.contains("refusing reuse"), "{err}");
    }
}

#[test]
fn a_generation_without_views_is_unavailable_to_the_records_query() {
    let (_tmp, root) = setup();
    let built = build(&root);
    let gen = built["generation_id"].as_str().unwrap();
    let manifest_path = gen_dir(&root, gen).join("manifest.json");
    let mut manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    manifest.as_object_mut().unwrap().remove("records");
    fs::write(&manifest_path, serde_json::to_vec(&manifest).unwrap()).unwrap();
    let q = records(&root);
    assert_eq!(q["state"], "unavailable");
    assert_eq!(q["reason"], "generation has no map/graph records");
}
