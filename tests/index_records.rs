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
    git_at(dir, args, "2026-11-10T01:02:03Z");
}

fn git_at(dir: &Path, args: &[&str], date: &str) {
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
        .env("GIT_AUTHOR_DATE", date)
        .env("GIT_COMMITTER_DATE", date)
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
    // A later committer time: reused fragments must be restamped with it.
    git(&root, &["add", "-A"]);
    git_at(
        &root,
        &["commit", "-q", "-m", "second\n\nRefs: #2348"],
        "2026-11-10T05:06:07Z",
    );

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
    for view in ["map", "graph"] {
        for e in q[view]["edges"].as_array().unwrap() {
            assert_eq!(e["observed_at"], "2026-11-10T05:06:07Z", "{view}");
        }
    }
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

#[test]
fn fragments_of_other_extractor_versions_are_never_reused() {
    let (_tmp, root) = setup();
    let built = build(&root);
    let gen = built["generation_id"].as_str().unwrap();
    let manifest_path = gen_dir(&root, gen).join("manifest.json");
    let mut manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    // Simulate a generation written by an older link extractor.
    manifest["records"]["version"] = Value::from("records=1;spec-links=0;tag-parser=1");
    fs::write(&manifest_path, serde_json::to_vec(&manifest).unwrap()).unwrap();
    write(&root, "README.md", "changed\n");
    commit(&root, "touch");
    let next = build(&root);
    assert_eq!(next["records"]["reused_files"], 0);
    assert_eq!(next["records"]["extracted_files"], 5);
}

#[test]
fn unreadable_history_is_missing_coverage_not_a_failed_build() {
    let (_tmp, root) = setup();
    // A missing parent commit object makes the history walk fail while the file scan
    // (the index and HEAD) still succeeds.
    let parent = String::from_utf8(
        Command::new("git")
            .args(["rev-parse", "HEAD"])
            .current_dir(&root)
            .output()
            .unwrap()
            .stdout,
    )
    .unwrap();
    let parent = parent.trim();
    commit(&root, "second");
    fs::remove_file(
        root.join(".git/objects")
            .join(&parent[..2])
            .join(&parent[2..]),
    )
    .unwrap();
    let built = build(&root);
    assert_eq!(built["status"], "complete");
    let q = records(&root);
    let missing = q["graph"]["missing"].as_array().unwrap();
    assert!(
        missing.iter().any(|m| m["kind"] == "history_unavailable"),
        "{missing:?}"
    );
    assert!(edges(&q["graph"]).iter().all(|e| e.0 != "mentions_issue"));
}

fn issues(q: &Value) -> BTreeSet<String> {
    edges(&q["graph"])
        .into_iter()
        .filter(|e| e.0 == "mentions_issue")
        .map(|e| e.2)
        .collect()
}

fn has_miss(q: &Value, kind: &str) -> bool {
    q["graph"]["missing"]
        .as_array()
        .unwrap()
        .iter()
        .any(|m| m["kind"] == kind)
}

/// Pre-audit B1 (#2347 comment 6774913904): unshallowing a clone keeps HEAD and every
/// tracked byte but extends the reachable history, so the Graph changes. The new
/// history must give a new generation (no "refusing reuse" on the old id), the
/// incremental build must still equal a full rebuild, and the shallow boundary must be
/// visible as missing coverage while it exists.
#[test]
fn unshallowing_a_clone_builds_a_new_generation_equal_to_a_full_rebuild() {
    let (tmp, producer) = setup();
    commit(&producer, "second\n\nRefs: #42");
    let url = format!("file://{}", producer.display());
    git(
        tmp.path(),
        &["clone", "-q", "--depth", "1", &url, "shallow"],
    );
    let root = tmp.path().join("shallow");

    let shallow = build(&root);
    let q = records(&root);
    assert_eq!(issues(&q), BTreeSet::from(["issue:42".to_string()]));
    assert!(has_miss(&q, "history_shallow"));

    git(&root, &["fetch", "-q", "--unshallow", "origin"]);
    let full_history = build(&root);
    assert_ne!(full_history["generation_id"], shallow["generation_id"]);
    assert_eq!(full_history["records"]["reused_files"], 5);
    let gen = full_history["generation_id"].as_str().unwrap().to_string();
    let q = records(&root);
    assert_eq!(q["state"], "fresh");
    assert_eq!(
        issues(&q),
        BTreeSet::from(["issue:42".to_string(), "issue:2347".to_string()])
    );
    assert!(!has_miss(&q, "history_shallow"));

    let incremental = view_bytes(&root, &gen);
    fs::remove_dir_all(root.join(".moai-cli/index")).unwrap();
    let rebuilt = build(&root);
    assert_eq!(rebuilt["generation_id"], gen.as_str());
    assert_eq!(rebuilt["records"]["reused_files"], 0);
    assert_eq!(view_bytes(&root, &gen), incremental);
}

/// #2442 AC-2: code references depend on the file set, not only on the source bytes. A
/// target module file that appears or disappears while the referencing file is unchanged
/// (and therefore reused) still changes the edges, and the incremental views equal a full
/// rebuild byte for byte.
#[test]
fn code_references_follow_the_file_set_and_rebuild_identically() {
    let (_tmp, root) = setup();
    write(
        &root,
        "src/lib.rs",
        "pub mod pay;\npub fn digest() {}\nfn f() { crate::pay::once(); }\n",
    );
    write(
        &root,
        "src/user.rs",
        "use crate::pay;\nfn g() { crate::digest(); }\n",
    );
    commit(&root, "lib without the pay file");
    let first = build(&root);
    let q = records(&root);
    assert!(!edges(&q["graph"]).contains(&t("calls", "file:src/lib.rs", "file:src/pay.rs")));
    let misses = serde_json::to_string(&q["graph"]["missing"]).unwrap();
    assert!(misses.contains("module_file_missing"), "{misses}");
    assert!(
        misses.contains("code_reference_unreached"),
        "src/user.rs has no mod line"
    );
    assert_eq!(first["records"]["reused_files"], 0);

    // Only the target appears: src/lib.rs and src/user.rs bytes are unchanged.
    write(&root, "src/pay.rs", "pub fn once() {}\n");
    commit(&root, "add the pay file");
    let incremental = build(&root);
    let gen = incremental["generation_id"].as_str().unwrap().to_string();
    assert_eq!(incremental["records"]["extracted_files"], 1);
    let q = records(&root);
    let got = edges(&q["graph"]);
    assert!(got.contains(&t("calls", "file:src/lib.rs", "file:src/pay.rs")));
    for e in q["graph"]["edges"].as_array().unwrap() {
        if e["kind"] == "calls" && e["target_id"] == "file:src/pay.rs" {
            assert_eq!(e["basis"], "static_resolved");
            assert_eq!(e["uncertainty"], "callee-binding-not-checked");
            assert_eq!(e["extractor"]["name"], "code-refs");
            let ev = e["evidence_refs"][0].as_str().unwrap();
            assert!(ev.starts_with("file:src/lib.rs#L3@sha256:"), "{ev}");
        }
    }
    let incremental_views = view_bytes(&root, &gen);
    fs::remove_dir_all(root.join(".moai-cli/index")).unwrap();
    let full = build(&root);
    assert_eq!(full["generation_id"], gen.as_str());
    assert_eq!(full["records"]["reused_files"], 0);
    let full_views = view_bytes(&root, &gen);
    assert_eq!(incremental_views[1], full_views[1], "graph.jsonl differs");
    assert_eq!(incremental_views[0], full_views[0], "map.jsonl differs");

    // The target disappears again: the edge goes, the gap is stated.
    fs::remove_file(root.join("src/pay.rs")).unwrap();
    commit(&root, "drop the pay file");
    build(&root);
    let q = records(&root);
    assert!(!edges(&q["graph"]).contains(&t("calls", "file:src/lib.rs", "file:src/pay.rs")));
    assert!(serde_json::to_string(&q["graph"]["missing"])
        .unwrap()
        .contains("module_file_missing"));
}

/// #2450: changing only a documented definition re-resolves reused source scans.
/// Preserve each previous index outside the fixture instead of deleting it.
#[test]
fn doc_references_rebuild_identically_when_only_the_definition_changes() {
    let (tmp, root) = setup();
    write(&root, "src/lib.rs", "pub mod user; pub mod service;\n");
    write(&root, "src/user.rs", "//! `crate::service::parse`\n");
    write(&root, "src/service.rs", "pub fn absent() {}\n");
    commit(&root, "doc initially unresolved");
    build(&root);
    for (step, target, found, reason) in [
        (0, "pub fn parse() {}\n", true, ""),
        (1, "pub fn renamed() {}\n", false, "unresolved"),
        (
            2,
            "pub fn parse() {}\npub fn parse() {}\n",
            false,
            "ambiguous",
        ),
        (3, "pub fn parse() {}\n", true, ""),
    ] {
        write(&root, "src/service.rs", target);
        commit(&root, &format!("definition step {step}"));
        let inc = build(&root);
        assert_eq!(inc["records"]["extracted_files"], 1);
        assert!(inc["records"]["reused_files"].as_u64().unwrap() > 0);
        let gen = inc["generation_id"].as_str().unwrap();
        let q = records(&root);
        validate_document(&q["graph"]).unwrap();
        assert_eq!(
            edges(&q["graph"]).contains(&t("documents", "file:src/user.rs", "file:src/service.rs")),
            found
        );
        if !found {
            assert!(q["graph"]["missing"]
                .as_array()
                .unwrap()
                .iter()
                .any(|m| m["kind"] == "doc_reference_missing" && m["reason"] == reason));
        }
        let incremental = view_bytes(&root, gen);
        fs::rename(
            root.join(".moai-cli/index"),
            tmp.path().join(format!("saved-index-{step}")),
        )
        .unwrap();
        let full = build(&root);
        assert_eq!(full["generation_id"], gen);
        assert_eq!(full["records"]["reused_files"], 0);
        assert_eq!(view_bytes(&root, gen), incremental);
    }
}

#[test]
fn excluded_doc_blocks_survive_reuse_and_disappear_when_the_block_becomes_plain() {
    let (tmp, root) = setup();
    write(&root, "src/lib.rs", "pub mod user; pub mod service;\n");
    write(
        &root,
        "src/user.rs",
        "//! See `crate::service::parse`.\n//! > excluded\n",
    );
    write(&root, "src/service.rs", "pub fn parse() {}\n");
    commit(&root, "excluded documentation");
    build(&root);
    for (step, plain) in [(0, false), (1, true)] {
        if plain {
            write(&root, "src/user.rs", "//! See `crate::service::parse`.\n");
        } else {
            write(
                &root,
                "src/service.rs",
                "pub fn parse() {}\npub fn other() {}\n",
            );
        }
        commit(&root, &format!("doc block step {step}"));
        let inc = build(&root);
        assert!(inc["records"]["reused_files"].as_u64().unwrap() > 0);
        let gen = inc["generation_id"].as_str().unwrap();
        let q = records(&root);
        validate_document(&q["graph"]).unwrap();
        assert_eq!(
            edges(&q["graph"]).contains(&t("documents", "file:src/user.rs", "file:src/service.rs")),
            plain
        );
        let count: u64 = q["graph"]["missing"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|m| m["kind"] == "doc_reference_unsupported")
            .map(|m| m["count"].as_u64().unwrap())
            .sum();
        assert_eq!(count, u64::from(!plain));
        let incremental = view_bytes(&root, gen);
        fs::rename(
            root.join(".moai-cli/index"),
            tmp.path().join(format!("narrow-saved-{step}")),
        )
        .unwrap();
        let full = build(&root);
        assert_eq!(full["generation_id"], gen);
        assert_eq!(view_bytes(&root, gen), incremental);
    }
}
