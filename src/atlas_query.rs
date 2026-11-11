//! Code Map / Code Graph read ops (rc.2 C9 #2185).
//!
//! Design basis (Core Brain detailed design, local text copy
//! `references/chatgpt-space/page_136c8203bf2881919eb1f543f72f87ea.md`, whole-file sha256
//! `49e445f4…47fa`): Code Map answers "what is where" (L2859-2869); the graph walk explores
//! typed relations of the same snapshot's Map and Graph with a bounded hop count and shows
//! coverage when the graph is absent (L4300).
//!
//! Both ops read the records of the published index generation through
//! `index::index_records` (C6b #2347), which already verifies that the two views belong to
//! the same generation and file set. When no records can be read the answer is
//! `unavailable` (REQ-C9-01); `stale` and `partial` pass through with their records.
//!
//! What the records hold today: file, tag, requirement, test, commit and issue nodes with
//! `annotated_by`, `references_requirement`, `verifies` and `mentions_issue` edges, and
//! file → file `imports` and `calls` edges of Rust files (C6c #2442, `src/code_refs.rs`).
//! Qualified doc-comment references add `documents` edges with mention and definition
//! evidence (C6d #2450); these are inferred lexical references, not executable calls.
//! There are no symbol or module records, so symbol questions are reported `unavailable`
//! in `coverage`, never answered by inference.
//!
//! Read-only: no file is written, no network is used. Answer-only data (hop distances,
//! cuts) lives in sibling fields so `map` and `graph` stay C4 documents.
use crate::index::{index_records, IndexQueryRequest};
use crate::ontology::validate_document;
use serde::Deserialize;
use serde_json::{json, Map, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    path::{Component, Path},
};

type Result<T> = std::result::Result<T, String>;

/// Default and largest number of files in one map answer.
pub const DEFAULT_MAX_FILES: usize = 200;
pub const MAX_FILES_LIMIT: usize = 2000;
/// Default and largest hop count of one graph walk.
pub const DEFAULT_MAX_HOPS: usize = 1;
pub const MAX_HOPS_LIMIT: usize = 3;
/// Default and largest number of edges in one graph answer.
pub const DEFAULT_MAX_EDGES: usize = 500;
pub const MAX_EDGES_LIMIT: usize = 5000;
/// Largest number of query terms (paths + prefixes, or start ids).
pub const MAX_TERMS: usize = 64;
/// Largest byte length of one query term.
pub const MAX_TERM_BYTES: usize = 512;

/// Coverage note for questions the records cannot answer.
pub const SYMBOLS_UNAVAILABLE: &str =
    "unavailable: the index records hold no symbol or module records; imports and calls are file to file (Rust only)";

#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct MapQuery {
    /// Exact repository-relative file paths.
    #[serde(default)]
    pub paths: Vec<String>,
    /// Directory prefixes: `src` selects every file under `src/`.
    #[serde(default)]
    pub prefixes: Vec<String>,
}

#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct MapLimits {
    pub max_files: Option<usize>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct MapRequest {
    pub protocol_version: u32,
    #[serde(default)]
    pub query: MapQuery,
    #[serde(default)]
    pub limits: MapLimits,
}

#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct GraphQuery {
    /// Node ids the walk starts from (`file:<path>`, `req:<id>`, `test:<path>::<name>`, …).
    #[serde(default)]
    pub start: Vec<String>,
}

#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct GraphLimits {
    pub max_hops: Option<usize>,
    pub max_edges: Option<usize>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct GraphRequest {
    pub protocol_version: u32,
    #[serde(default)]
    pub query: GraphQuery,
    #[serde(default)]
    pub limits: GraphLimits,
}

/// A repository-relative path of normal components only.
fn relative_ok(path: &str) -> bool {
    !path.is_empty()
        && path.len() <= MAX_TERM_BYTES
        && Path::new(path)
            .components()
            .all(|c| matches!(c, Component::Normal(_)))
}

fn limit(value: Option<usize>, default: usize, max: usize, name: &str) -> Result<usize> {
    let v = value.unwrap_or(default);
    if v == 0 || v > max {
        return Err(format!("refused: {name} must be 1..={max}"));
    }
    Ok(v)
}

/// Reads the published generation's views. `Err(answer)` is the `unavailable` answer to
/// return as it is; `Ok` carries the freshness answer and both C4 documents.
fn records(project: &Path) -> Result<std::result::Result<(Value, Value, Value), Value>> {
    let answer = index_records(
        project,
        &IndexQueryRequest {
            protocol_version: 1,
        },
    )?;
    let (map, graph) = (answer["map"].clone(), answer["graph"].clone());
    if map.is_null() || graph.is_null() {
        let reason = answer["reason"]
            .as_str()
            .unwrap_or("no published index generation")
            .to_string();
        return Ok(Err(json!({
            "protocol_version": 1,
            "state": "unavailable",
            "reason": reason,
            "generation_id": answer["generation_id"],
            "coverage": {"symbols": SYMBOLS_UNAVAILABLE},
            "network": "disabled",
            "writes": false,
        })));
    }
    Ok(Ok((answer, map, graph)))
}

fn head(answer: &Value) -> Map<String, Value> {
    let mut out = Map::new();
    out.insert("protocol_version".into(), json!(1));
    out.insert("state".into(), answer["state"].clone());
    out.insert("reason".into(), answer["reason"].clone());
    out.insert("generation_id".into(), answer["generation_id"].clone());
    out.insert("network".into(), json!("disabled"));
    out.insert("writes".into(), json!(false));
    out
}

fn array<'a>(doc: &'a Value, key: &str) -> impl Iterator<Item = &'a Value> {
    doc[key].as_array().into_iter().flatten()
}

fn path_of(record: &Value) -> Option<&str> {
    record["properties"]["path"]
        .as_str()
        .or_else(|| record["path"].as_str())
}

fn checked(doc: Value) -> Result<Value> {
    validate_document(&doc).map_err(|e| format!("answer is not a C4 document: {e}"))?;
    Ok(doc)
}

/// Selects the Map records of the requested files: their file nodes, their tag nodes, the
/// `annotated_by` edges between them and their missing-coverage entries. An empty query
/// selects every file. Files are taken in path order; the ones past `max_files` are
/// `omitted`, and a requested path that is not in the generation is `omitted` too.
pub fn select_map(map: &Value, query: &MapQuery, max_files: usize) -> (Value, Vec<Value>, usize) {
    let all: BTreeSet<&str> = array(map, "nodes")
        .filter(|n| n["kind"] == "file")
        .filter_map(path_of)
        .collect();
    let dirs: Vec<String> = query
        .prefixes
        .iter()
        .map(|p| format!("{}/", p.trim_end_matches('/')))
        .collect();
    let every = query.paths.is_empty() && query.prefixes.is_empty();
    let mut omitted = Vec::new();
    for p in &query.paths {
        if !all.contains(p.as_str()) {
            omitted.push(json!({"path": p, "reason": "path not in the generation"}));
        }
    }
    let mut chosen: BTreeSet<&str> = BTreeSet::new();
    for path in &all {
        let wanted = every
            || query.paths.iter().any(|p| p == path)
            || dirs.iter().any(|d| path.starts_with(d.as_str()));
        if !wanted {
            continue;
        }
        if chosen.len() >= max_files {
            omitted.push(json!({"path": path, "reason": "max_files"}));
            continue;
        }
        chosen.insert(path);
    }
    let in_chosen = |r: &Value| path_of(r).is_some_and(|p| chosen.contains(p));
    let nodes: Vec<Value> = array(map, "nodes")
        .filter(|n| in_chosen(n))
        .cloned()
        .collect();
    let ids: BTreeSet<&str> = nodes.iter().filter_map(|n| n["id"].as_str()).collect();
    let edges: Vec<Value> = array(map, "edges")
        .filter(|e| e["source_id"].as_str().is_some_and(|s| ids.contains(s)))
        .cloned()
        .collect();
    let missing: Vec<Value> = array(map, "missing")
        .filter(|m| in_chosen(m))
        .cloned()
        .collect();
    let total = all.len();
    (
        json!({"nodes": nodes, "edges": edges, "missing": missing}),
        omitted,
        total,
    )
}

/// The result of one bounded walk over the union of the Map and Graph views.
pub struct Traversal {
    /// C4 document of the reached nodes and taken edges.
    pub document: Value,
    /// Hop distance of every reached node id (start nodes at 0).
    pub distance: BTreeMap<String, usize>,
    /// Unknown starts, edges cut by `max_edges`, endpoints without a node record.
    pub omitted: Vec<Value>,
}

/// Walks typed edges in both directions from `start`, at most `max_hops` hops and
/// `max_edges` edges, in a deterministic order (frontier and incident edges sorted by id).
/// Nodes are the union of both views keyed by id; an edge endpoint with no node record is
/// listed in `omitted`, never synthesized. Library entry point for search (#2408).
pub fn traverse(
    map: &Value,
    graph: &Value,
    start: &[String],
    max_hops: usize,
    max_edges: usize,
) -> Traversal {
    let mut nodes: BTreeMap<String, Value> = BTreeMap::new();
    let mut omitted = Vec::new();
    for n in array(map, "nodes").chain(array(graph, "nodes")) {
        let Some(id) = n["id"].as_str() else { continue };
        match nodes.get(id) {
            None => {
                nodes.insert(id.to_string(), n.clone());
            }
            Some(seen) if seen != n => {
                omitted.push(json!({"id": id, "reason": "node record differs between views"}))
            }
            Some(_) => {}
        }
    }
    let mut edges: BTreeMap<String, Value> = BTreeMap::new();
    for e in array(map, "edges").chain(array(graph, "edges")) {
        if let Some(id) = e["edge_id"].as_str() {
            edges.entry(id.to_string()).or_insert_with(|| e.clone());
        }
    }
    let mut incident: BTreeMap<&str, BTreeSet<&str>> = BTreeMap::new();
    for (id, e) in &edges {
        for end in ["source_id", "target_id"] {
            if let Some(n) = e[end].as_str() {
                incident.entry(n).or_default().insert(id.as_str());
            }
        }
    }
    let mut distance: BTreeMap<String, usize> = BTreeMap::new();
    let mut frontier: BTreeSet<String> = BTreeSet::new();
    for s in start {
        if nodes.contains_key(s) || incident.contains_key(s.as_str()) {
            distance.insert(s.clone(), 0);
            frontier.insert(s.clone());
        } else {
            omitted.push(json!({"id": s, "reason": "start node not in the generation"}));
        }
    }
    let mut taken: BTreeSet<&str> = BTreeSet::new();
    let mut cut: BTreeSet<&str> = BTreeSet::new();
    for hop in 0..max_hops {
        let mut next: BTreeSet<String> = BTreeSet::new();
        for node in &frontier {
            for &edge_id in incident.get(node.as_str()).into_iter().flatten() {
                if taken.contains(edge_id) || cut.contains(edge_id) {
                    continue;
                }
                if taken.len() >= max_edges {
                    cut.insert(edge_id);
                    continue;
                }
                taken.insert(edge_id);
                let e = &edges[edge_id];
                for end in ["source_id", "target_id"] {
                    let Some(other) = e[end].as_str() else {
                        continue;
                    };
                    if !distance.contains_key(other) {
                        distance.insert(other.to_string(), hop + 1);
                        next.insert(other.to_string());
                    }
                }
            }
        }
        frontier = next;
        if frontier.is_empty() {
            break;
        }
    }
    for edge_id in &cut {
        omitted.push(json!({"edge_id": edge_id, "reason": "max_edges"}));
    }
    let mut out_nodes = Vec::new();
    for id in distance.keys() {
        match nodes.get(id) {
            Some(n) => out_nodes.push(n.clone()),
            None => omitted.push(json!({"id": id, "reason": "edge endpoint has no node record"})),
        }
    }
    let reached_paths: BTreeSet<&str> = out_nodes
        .iter()
        .filter(|n| n["kind"] == "file")
        .filter_map(path_of)
        .collect();
    let missing: Vec<Value> = array(map, "missing")
        .chain(array(graph, "missing"))
        .filter(|m| path_of(m).is_some_and(|p| reached_paths.contains(p)))
        .cloned()
        .collect();
    let out_edges: Vec<Value> = taken.iter().map(|id| edges[*id].clone()).collect();
    Traversal {
        document: json!({"nodes": out_nodes, "edges": out_edges, "missing": missing}),
        distance,
        omitted,
    }
}

/// `atlas.map`: the Code Map records of the requested files (REQ-C9-01: read-only,
/// `unavailable` without an index).
pub fn atlas_map(project: &Path, req: &MapRequest) -> Result<Value> {
    if req.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let q = &req.query;
    if q.paths.len() + q.prefixes.len() > MAX_TERMS {
        return Err("refused: too many query terms".into());
    }
    if !q.paths.iter().chain(&q.prefixes).all(|p| relative_ok(p)) {
        return Err(
            "refused: a path or prefix must be relative inside the repository, at most 512 bytes"
                .into(),
        );
    }
    let max_files = limit(
        req.limits.max_files,
        DEFAULT_MAX_FILES,
        MAX_FILES_LIMIT,
        "max_files",
    )?;
    let (answer, map, _) = match records(project)? {
        Ok(v) => v,
        Err(unavailable) => return Ok(unavailable),
    };
    let (doc, omitted, total) = select_map(&map, q, max_files);
    let returned = array(&doc, "nodes").filter(|n| n["kind"] == "file").count();
    let mut out = head(&answer);
    out.insert("map".into(), checked(doc)?);
    out.insert("omitted".into(), json!(omitted));
    out.insert(
        "coverage".into(),
        json!({
            "view": "map",
            "files_in_generation": total,
            "files_returned": returned,
            "symbols": SYMBOLS_UNAVAILABLE,
        }),
    );
    Ok(Value::Object(out))
}

/// `atlas.graph`: a bounded walk over the same generation's Map and Graph views
/// (REQ-C9-01: read-only, `unavailable` without an index).
pub fn atlas_graph(project: &Path, req: &GraphRequest) -> Result<Value> {
    if req.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let start = &req.query.start;
    if start.is_empty() {
        return Err("refused: give at least one start node id".into());
    }
    if start.len() > MAX_TERMS {
        return Err("refused: too many query terms".into());
    }
    if !start
        .iter()
        .all(|s| !s.is_empty() && s.len() <= MAX_TERM_BYTES && !s.contains('\n'))
    {
        return Err("refused: a start id must be non-empty, one line, at most 512 bytes".into());
    }
    let max_hops = limit(
        req.limits.max_hops,
        DEFAULT_MAX_HOPS,
        MAX_HOPS_LIMIT,
        "max_hops",
    )?;
    let max_edges = limit(
        req.limits.max_edges,
        DEFAULT_MAX_EDGES,
        MAX_EDGES_LIMIT,
        "max_edges",
    )?;
    let (answer, map, graph) = match records(project)? {
        Ok(v) => v,
        Err(unavailable) => return Ok(unavailable),
    };
    let t = traverse(&map, &graph, start, max_hops, max_edges);
    let kinds: BTreeSet<&str> = array(&map, "edges")
        .chain(array(&graph, "edges"))
        .filter_map(|e| e["kind"].as_str())
        .collect();
    let mut out = head(&answer);
    out.insert("graph".into(), checked(t.document)?);
    out.insert("distance".into(), json!(t.distance));
    out.insert("omitted".into(), json!(t.omitted));
    out.insert(
        "coverage".into(),
        json!({
            "views": ["map", "graph"],
            "max_hops": max_hops,
            "edge_kinds_in_generation": kinds,
            "symbols": SYMBOLS_UNAVAILABLE,
        }),
    );
    Ok(Value::Object(out))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn node(id: &str, kind: &str, path: Option<&str>) -> Value {
        let mut n = json!({"id": id, "kind": kind, "revision": "sha256:x"});
        if let Some(p) = path {
            n["properties"] = json!({"path": p});
        }
        n
    }

    fn edge(id: &str, kind: &str, s: &str, t: &str) -> Value {
        json!({"edge_id": id, "kind": kind, "source_id": s, "target_id": t})
    }

    fn chain() -> (Value, Value) {
        // file:a -annotated_by-> tag:a#L1 ; file:a -references_requirement-> req:R
        // test:t -verifies-> req:R ; file:b -references_requirement-> req:Q
        let map = json!({
            "nodes": [node("file:a", "file", Some("a")), node("tag:a#L1", "tag", Some("a")),
                      node("file:b", "file", Some("b")), node("file:c/d", "file", Some("c/d")),
                      node("file:cd", "file", Some("cd"))],
            "edges": [edge("e1", "annotated_by", "file:a", "tag:a#L1")],
            "missing": [{"kind": "tag_scan_skipped", "path": "b", "reason": "x"}],
        });
        let graph = json!({
            "nodes": [node("file:a", "file", Some("a")), node("req:R", "requirement", None),
                      node("test:t", "test", Some("t"))],
            "edges": [edge("e2", "references_requirement", "file:a", "req:R"),
                      edge("e3", "verifies", "test:t", "req:R"),
                      edge("e4", "references_requirement", "file:b", "req:Q")],
            "missing": [],
        });
        (map, graph)
    }

    fn ids(doc: &Value) -> Vec<&str> {
        array(doc, "nodes")
            .filter_map(|n| n["id"].as_str())
            .collect()
    }

    #[test]
    fn hops_bound_the_walk_and_distances_are_recorded() {
        let (m, g) = chain();
        let one = traverse(&m, &g, &["file:a".into()], 1, 100);
        assert_eq!(ids(&one.document), ["file:a", "req:R", "tag:a#L1"]);
        let two = traverse(&m, &g, &["file:a".into()], 2, 100);
        assert_eq!(
            ids(&two.document),
            ["file:a", "req:R", "tag:a#L1", "test:t"]
        );
        assert_eq!(two.distance["test:t"], 2);
        assert_eq!(two.distance["file:a"], 0);
        assert!(two.omitted.is_empty());
    }

    #[test]
    fn unknown_starts_cut_edges_and_dangling_endpoints_are_omitted() {
        let (m, g) = chain();
        let t = traverse(&m, &g, &["file:zz".into(), "file:a".into()], 1, 1);
        let reasons: Vec<&str> = t
            .omitted
            .iter()
            .filter_map(|o| o["reason"].as_str())
            .collect();
        assert!(reasons.contains(&"start node not in the generation"));
        assert!(reasons.contains(&"max_edges"));
        assert_eq!(array(&t.document, "edges").count(), 1);
        // req:Q is an endpoint with no node record: listed, never synthesized.
        let d = traverse(&m, &g, &["file:b".into()], 1, 100);
        assert_eq!(ids(&d.document), ["file:b"]);
        assert!(d
            .omitted
            .contains(&json!({"id": "req:Q", "reason": "edge endpoint has no node record"})));
        assert_eq!(d.document["missing"][0]["path"], "b");
    }

    #[test]
    fn map_selection_by_path_prefix_and_cap() {
        let (m, _) = chain();
        let q = MapQuery {
            paths: vec!["a".into(), "nope".into()],
            prefixes: vec!["c/".into()],
        };
        let (doc, omitted, total) = select_map(&m, &q, 10);
        assert_eq!(total, 4);
        assert_eq!(ids(&doc), ["file:a", "tag:a#L1", "file:c/d"]);
        assert_eq!(array(&doc, "edges").count(), 1);
        assert_eq!(
            omitted,
            [json!({"path": "nope", "reason": "path not in the generation"})]
        );
        // A prefix is a directory: "c" does not select a sibling file "cd".
        let (doc, _, _) = select_map(
            &m,
            &MapQuery {
                paths: vec![],
                prefixes: vec!["c".into()],
            },
            10,
        );
        assert_eq!(ids(&doc), ["file:c/d"]);
        let (doc, omitted, _) = select_map(&m, &MapQuery::default(), 2);
        assert_eq!(ids(&doc), ["file:a", "tag:a#L1", "file:b"]);
        assert_eq!(
            omitted,
            [
                json!({"path": "c/d", "reason": "max_files"}),
                json!({"path": "cd", "reason": "max_files"})
            ]
        );
        assert_eq!(doc["missing"][0]["path"], "b");
    }

    #[test]
    fn limits_have_literal_n_and_n_plus_one_boundaries() {
        assert_eq!(limit(Some(2000), 200, MAX_FILES_LIMIT, "f"), Ok(2000));
        assert!(limit(Some(2001), 200, MAX_FILES_LIMIT, "f").is_err());
        assert_eq!(limit(Some(3), 1, MAX_HOPS_LIMIT, "h"), Ok(3));
        assert!(limit(Some(4), 1, MAX_HOPS_LIMIT, "h").is_err());
        assert_eq!(limit(Some(5000), 500, MAX_EDGES_LIMIT, "e"), Ok(5000));
        assert!(limit(Some(5001), 500, MAX_EDGES_LIMIT, "e").is_err());
        assert!(limit(Some(0), 1, 3, "h").is_err());
        assert!(relative_ok(&"a".repeat(512)));
        assert!(!relative_ok(&"a".repeat(513)));
        assert!(!relative_ok("../x"));
        assert!(!relative_ok("/x"));
    }
}
