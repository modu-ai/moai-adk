//! Map and Graph records of an index generation (C6b #2347).
//!
//! Design basis (Core Brain detailed design; local text copy
//! `references/chatgpt-space/page_136c8203bf2881919eb1f543f72f87ea.md`, 4,450 lines,
//! retrieved 2026-11-08T23:19:26Z; body sha256 `747ac160…57cb` as its header line
//! states and as re-measured by the #2347 lane — `49e445f4…47fa` is the whole file
//! including that header line; the live Space page was not re-read): one generation
//! directory holds `manifest.json`, `map.jsonl`,
//! `graph.jsonl` (L3816-3819); Map and Graph are two views built under the same
//! repository identity and snapshot and never silently show different file sets or
//! commits (L2936); a full rebuild and an incremental build of the same input give the
//! same semantic result, failure coverage included (INCREMENTAL EQUIVALENCE,
//! L4057-4061). `lexical/` (L3820) is out of scope here.
//!
//! The two views:
//! - Map (`map.jsonl`): one `file:<path>` node per manifest entry, one `tag:<path>#L<n>`
//!   node per C2 tag (`@MOAI:` / `@MX:`), and a `file annotated_by tag` edge with the
//!   line evidence. Files the tag scanner cannot read are listed as `missing`.
//! - Graph (`graph.jsonl`): the C5 explicit links (`references_requirement`,
//!   `verifies`, `mentions_issue`) and their `missing` coverage (`src/spec_links.rs`),
//!   plus the C6c file → file code references (`imports`, `calls`) of Rust files
//!   (`src/code_refs.rs`, #2442). Each Rust file stores its scan as a per-origin
//!   `code_scan` node; the edges depend on the whole file set, so they are resolved
//!   again on every build from the stored scans and never reused as stored lines.
//!   C6d #2450 also stores qualified doc-comment candidates and module-level item
//!   declarations in those scans. `documents` edges require a unique declaration
//!   and carry both mention and definition evidence; ambiguity is missing coverage.
//!
//! Every record line carries `origin`: the manifest path whose bytes produced it, or
//! `null` for records derived from the whole snapshot (requirement nodes, commit
//! history, triangle gaps). An incremental build reuses the per-origin records of files
//! whose digest is unchanged and recomputes every global record; the result is
//! byte-identical to a full rebuild because both are canonical (sorted lines) and edge
//! stamps (`snapshot_id`, `observed_at`) are re-applied, never carried over.
//!
//! Time is content here: `observed_at` is the commit time of the manifest's base
//! commit, so the same input always yields the same records.

use crate::{atlas_tag, code_refs, ontology, spec_links};
use ontology::{OntologyEdge, OntologyNode};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use spec_links::Builder;
use std::collections::BTreeMap;

type Result<T> = std::result::Result<T, String>;

/// Bump on any change to the record shape or to what a view contains.
pub const RECORDS_VERSION: &str = "1";
/// The record, link-extractor and tag-parser versions that produced a view. Stored
/// fragments are reused only under the same string: reused edges are restamped with
/// the current extractor identity, so output of an older extractor must never be
/// relabelled as current.
pub fn versions() -> String {
    format!(
        "records={RECORDS_VERSION};{}={};tag-parser={};{}={}",
        spec_links::EXTRACTOR_NAME,
        spec_links::EXTRACTOR_VERSION,
        atlas_tag::PARSER_VERSION,
        code_refs::EXTRACTOR_NAME,
        code_refs::EXTRACTOR_VERSION
    )
}

/// A view larger than the index state read limit is refused, not truncated.
pub(crate) fn within_limit(len: u64, limit: u64) -> Result<()> {
    if len > limit {
        return Err(format!(
            "index records byte limit exceeded ({limit}); refused, not truncated"
        ));
    }
    Ok(())
}

/// Extractor identity of the map view's tag edges.
pub const MAP_EXTRACTOR: &str = "index-map";

/// One manifest entry's records in both views, before stamping.
#[derive(Default)]
pub(crate) struct Fragment {
    pub(crate) map: Builder,
    pub(crate) graph: Builder,
}

/// What the build learned about one manifest entry's bytes.
pub(crate) enum Content<'a> {
    /// A regular file whose bytes were read and matched the manifest digest.
    Bytes(&'a [u8]),
    /// A regular file not read because no view needs its bytes (over the tag limit
    /// and not a code file).
    Oversize,
    /// A symlink or gitlink entry.
    NotRegular,
}

/// Whether the build must read this regular file's bytes for either view.
pub(crate) fn needs_bytes(path: &str, bytes: u64) -> bool {
    bytes <= atlas_tag::FILE_LIMIT || spec_links::is_code_path(path)
}

fn representable(path: &str) -> bool {
    !path.contains('#') && !path.contains('@')
}

/// Extracts one manifest entry's records from its verified content.
pub(crate) fn extract_file(path: &str, sha256: &str, content: Content<'_>) -> Fragment {
    let mut f = Fragment::default();
    let file_id = format!("file:{path}");
    f.map.node(&file_id, "file", sha256, &[("path", path)]);
    match content {
        Content::NotRegular => {
            f.map.miss(
                "tag_scan_skipped",
                json!({"path": path, "reason": "not a regular file"}),
            );
            if spec_links::is_code_path(path) && representable(path) {
                f.graph.miss(
                    "unreadable_file",
                    json!({"path": path, "reason": "not a regular file"}),
                );
            }
        }
        Content::Oversize => {
            f.map.miss(
                "tag_scan_skipped",
                json!({"path": path, "reason": "tag file exceeds the size limit"}),
            );
        }
        Content::Bytes(bytes) => {
            map_tags(&mut f.map, path, sha256, &file_id, bytes);
            if spec_links::is_code_path(path) {
                spec_links::file_links(&mut f.graph, path, bytes.to_vec());
            }
            code_scan(&mut f.graph, path, sha256, bytes);
        }
    }
    f
}

/// The per-file half of the code-reference producer: a Rust file's scan, a Cargo
/// manifest's library name, or the coverage note for a code file no scanner reads.
fn code_scan(b: &mut Builder, path: &str, sha256: &str, bytes: &[u8]) {
    let name = path.rsplit('/').next().unwrap_or(path);
    if name == "Cargo.toml" {
        let lib = std::str::from_utf8(bytes)
            .ok()
            .and_then(code_refs::lib_name);
        let lib = lib.as_deref().unwrap_or("");
        b.node(
            &format!("code-crate:{path}"),
            code_refs::CRATE_NODE_KIND,
            sha256,
            &[("path", path), ("lib_name", lib)],
        );
        return;
    }
    if !spec_links::is_code_path(path) || !representable(path) {
        return;
    }
    if !path.ends_with(".rs") {
        b.miss(
            "code_reference_unsupported",
            json!({"path": path, "reason": "no code-reference scanner for this language"}),
        );
        return;
    }
    let Ok(text) = std::str::from_utf8(bytes) else {
        b.miss(
            "code_reference_skipped",
            json!({"path": path, "reason": "not UTF-8"}),
        );
        return;
    };
    let scan = code_refs::encode(&code_refs::scan(text));
    b.node(
        &format!("code-scan:{path}"),
        code_refs::SCAN_NODE_KIND,
        sha256,
        &[("path", path), ("scan", &scan)],
    );
}

/// The global half: resolves every stored scan against the file set and returns the
/// stamped edges keyed by source path, plus the module-resolution coverage misses.
fn code_edges(
    graph: &Builder,
    paths: &std::collections::BTreeSet<String>,
    snapshot_id: &str,
    observed_at: &str,
) -> (Vec<OntologyEdge>, Vec<Value>) {
    let mut scans = BTreeMap::new();
    let mut manifests = BTreeMap::new();
    let mut misses = Vec::new();
    for node in graph.nodes.values() {
        let Some(path) = node.properties.get("path") else {
            continue;
        };
        if node.kind == code_refs::SCAN_NODE_KIND {
            let stored = node
                .properties
                .get("scan")
                .and_then(|t| code_refs::decode(t));
            match stored {
                Some(scan) => {
                    scans.insert(path.clone(), (node.revision.clone(), scan));
                }
                None => misses.push(json!({"kind": "code_reference_skipped", "path": path, "reason": "stored scan unreadable"})),
            }
        } else if node.kind == code_refs::CRATE_NODE_KIND {
            let dir = path.rsplit_once('/').map_or("", |(d, _)| d).to_string();
            let lib = node
                .properties
                .get("lib_name")
                .filter(|n| !n.is_empty())
                .cloned();
            manifests.insert(dir, lib);
        }
    }
    let (refs, mut tree_misses) = code_refs::resolve(&scans, paths, &manifests);
    misses.append(&mut tree_misses);
    (code_refs::edges(&refs, snapshot_id, observed_at), misses)
}

fn map_tags(b: &mut Builder, path: &str, sha256: &str, file_id: &str, bytes: &[u8]) {
    if bytes.len() as u64 > atlas_tag::FILE_LIMIT {
        b.miss(
            "tag_scan_skipped",
            json!({"path": path, "reason": "tag file exceeds the size limit"}),
        );
        return;
    }
    let Ok(text) = std::str::from_utf8(bytes) else {
        b.miss(
            "tag_scan_skipped",
            json!({"path": path, "reason": "not UTF-8"}),
        );
        return;
    };
    let report = atlas_tag::scan_text(text);
    for inv in report["invalid"].as_array().into_iter().flatten() {
        b.miss(
            "invalid_marker",
            json!({"path": path, "line": inv["line"], "reason": inv["reason"]}),
        );
    }
    let tags = report["nodes"].as_array().cloned().unwrap_or_default();
    if !tags.is_empty() && !representable(path) {
        b.miss(
            "path_unrepresentable",
            json!({"path": path, "reason": "# or @ cannot appear in an evidence file-uri"}),
        );
        return;
    }
    for tag in tags {
        let line = tag["line"].as_u64().unwrap_or(0);
        let tag_id = format!("tag:{path}#L{line}");
        let mut props: Vec<(&str, String)> = vec![
            ("path", path.to_string()),
            ("namespace", tag["namespace"].as_str().unwrap_or("").into()),
            ("tag_kind", tag["kind"].as_str().unwrap_or("").into()),
            ("parser_version", atlas_tag::PARSER_VERSION.to_string()),
        ];
        for key in ["sub", "args"] {
            if let Some(v) = tag[key].as_str() {
                props.push((key, v.to_string()));
            }
        }
        let props: Vec<(&str, &str)> = props.iter().map(|(k, v)| (*k, v.as_str())).collect();
        b.node(&tag_id, "tag", sha256, &props);
        b.edge(
            "annotated_by",
            file_id,
            &tag_id,
            format!("file:{path}#L{line}@{sha256}"),
        );
    }
}

fn is_global_node(id: &str) -> bool {
    id.starts_with("req:")
}

/// Rebuilds one entry's fragment from the record lines a previous generation stored
/// under that origin. Requirement nodes are global lines, so they are re-derived from
/// the edges exactly as `spec_links::file_links` creates them.
pub(crate) fn fragment_from_lines(map: &[Value], graph: &[Value]) -> Result<Fragment> {
    let mut f = Fragment::default();
    for (lines, b) in [(map, &mut f.map), (graph, &mut f.graph)] {
        for line in lines {
            if let Some(node) = line.get("node") {
                let n: OntologyNode = serde_json::from_value(node.clone())
                    .map_err(|_| "stored record node unreadable".to_string())?;
                b.nodes.insert(n.id.clone(), n);
            } else if let Some(edge) = line.get("edge") {
                let e: OntologyEdge = serde_json::from_value(edge.clone())
                    .map_err(|_| "stored record edge unreadable".to_string())?;
                // Code references depend on the whole file set: re-resolved, never reused.
                if e.extractor.name == code_refs::EXTRACTOR_NAME {
                    continue;
                }
                if let Some(id) = e.target_id.strip_prefix("req:") {
                    b.requirement(id);
                }
                for r in e.evidence_refs {
                    b.edge(&e.kind, &e.source_id, &e.target_id, r);
                }
            } else if let Some(m) = line.get("missing") {
                b.missing.push(m.clone());
            } else {
                return Err("stored record line has no record".into());
            }
        }
    }
    Ok(f)
}

/// The serialized views of one generation.
pub(crate) struct Views {
    pub(crate) map: Vec<u8>,
    pub(crate) graph: Vec<u8>,
    pub(crate) map_records: usize,
    pub(crate) graph_records: usize,
}

/// Digest of the manifest's file table — the "same file set" key both headers carry.
pub(crate) fn file_set_digest(files: &[crate::index::IndexedFile]) -> String {
    let mut h = Sha256::new();
    h.update(b"moai-index-file-set-v1\0");
    for f in files {
        for part in [f.path.as_bytes(), f.sha256.as_bytes()] {
            h.update((part.len() as u64).to_be_bytes());
            h.update(part);
        }
        h.update(f.bytes.to_be_bytes());
    }
    format!("sha256:{:x}", h.finalize())
}

pub(crate) fn header(view: &str, generation_id: &str, base_commit: &str, file_set: &str) -> Value {
    json!({"header": {
        "records_version": versions(),
        "view": view,
        "generation_id": generation_id,
        "base_commit": base_commit,
        "file_set": file_set,
    }})
}

/// Stamps and serializes both views. `fragments` is keyed by manifest path; `history`
/// holds the commit → issue records of the base commit. Both documents pass
/// `ontology::validate_document` before any byte is returned.
pub(crate) fn assemble(
    fragments: BTreeMap<String, Fragment>,
    history: Builder,
    head: (&str, &str, &str),
    observed_at: &str,
) -> Result<Views> {
    let (generation_id, base_commit, file_set) = head;
    let paths: std::collections::BTreeSet<String> = fragments.keys().cloned().collect();
    let mut origin_of: [BTreeMap<String, String>; 2] = Default::default();
    let mut edge_origin: [BTreeMap<(String, String, String), String>; 2] = Default::default();
    let mut miss_origin: [Vec<(String, Value)>; 2] = Default::default();
    let mut all = [Builder::default(), history];
    for (path, frag) in fragments {
        for (i, part) in [frag.map, frag.graph].into_iter().enumerate() {
            for (id, node) in part.nodes {
                if !is_global_node(&id) {
                    origin_of[i].insert(id.clone(), path.clone());
                }
                all[i].nodes.entry(id).or_insert(node);
            }
            for (key, ev) in part.edges {
                edge_origin[i].insert(key.clone(), path.clone());
                all[i].edges.entry(key).or_default().extend(ev);
            }
            for m in part.missing {
                miss_origin[i].push((path.clone(), m.clone()));
                all[i].missing.push(m);
            }
        }
    }
    let [map_b, mut graph_b] = all;
    let (code, code_misses) = code_edges(&graph_b, &paths, base_commit, observed_at);
    graph_b.missing.extend(code_misses);
    let mut docs = [
        spec_links::finish_with(
            map_b,
            MAP_EXTRACTOR,
            RECORDS_VERSION,
            base_commit,
            observed_at,
        )?,
        spec_links::finish(graph_b, base_commit, observed_at)?,
    ];
    for e in code {
        edge_origin[1].insert(
            (e.kind.clone(), e.source_id.clone(), e.target_id.clone()),
            e.source_id.trim_start_matches("file:").to_string(),
        );
        let v = serde_json::to_value(&e).map_err(|_| "index records encode failed".to_string())?;
        if let Some(list) = docs[1]["edges"].as_array_mut() {
            list.push(v);
        }
    }
    ontology::validate_document(&docs[1])
        .map_err(|e| format!("code-reference records failed the C4 schema: {e}"))?;
    let mut out: Vec<(Vec<u8>, usize)> = Vec::new();
    for (i, (doc, view)) in docs.iter().zip(["map", "graph"]).enumerate() {
        let mut lines: Vec<String> = Vec::new();
        for n in doc["nodes"].as_array().into_iter().flatten() {
            let origin = n["id"].as_str().and_then(|id| origin_of[i].get(id));
            lines.push(line(origin, "node", n)?);
        }
        for e in doc["edges"].as_array().into_iter().flatten() {
            let key = (
                e["kind"].as_str().unwrap_or("").to_string(),
                e["source_id"].as_str().unwrap_or("").to_string(),
                e["target_id"].as_str().unwrap_or("").to_string(),
            );
            lines.push(line(edge_origin[i].get(&key), "edge", e)?);
        }
        // Per-file misses keep their origin; the rest (history, triangle gaps) are
        // global. Matching consumes one stored miss per emitted one.
        let mut pending = std::mem::take(&mut miss_origin[i]);
        for m in doc["missing"].as_array().into_iter().flatten() {
            let origin = pending
                .iter()
                .position(|(_, v)| v == m)
                .map(|at| pending.remove(at).0);
            lines.push(line(origin.as_ref(), "missing", m)?);
        }
        lines.sort();
        let count = lines.len();
        let mut bytes = serde_json::to_vec(&header(view, generation_id, base_commit, file_set))
            .map_err(|_| "index records encode failed".to_string())?;
        bytes.push(b'\n');
        for l in lines {
            bytes.extend_from_slice(l.as_bytes());
            bytes.push(b'\n');
        }
        out.push((bytes, count));
    }
    let (graph, graph_records) = out.pop().ok_or("index records encode failed")?;
    let (map, map_records) = out.pop().ok_or("index records encode failed")?;
    Ok(Views {
        map,
        graph,
        map_records,
        graph_records,
    })
}

fn line(origin: Option<&String>, key: &str, record: &Value) -> Result<String> {
    let mut v = serde_json::Map::new();
    v.insert("origin".into(), origin.map_or(Value::Null, |o| json!(o)));
    v.insert(key.into(), record.clone());
    serde_json::to_string(&Value::Object(v)).map_err(|_| "index records encode failed".into())
}

/// Parses a stored view: the header plus record lines, each validated as JSON.
pub(crate) fn parse_view(bytes: &[u8]) -> Result<(Value, Vec<Value>)> {
    let text = std::str::from_utf8(bytes).map_err(|_| "index records unreadable".to_string())?;
    let mut lines = text.lines();
    let header: Value = lines
        .next()
        .and_then(|l| serde_json::from_str(l).ok())
        .ok_or("index records header unreadable")?;
    let header = header
        .get("header")
        .cloned()
        .ok_or("index records header unreadable")?;
    let records = lines
        .map(|l| serde_json::from_str::<Value>(l).map_err(|_| "index records unreadable".into()))
        .collect::<Result<Vec<_>>>()?;
    Ok((header, records))
}

/// Groups stored record lines by origin (`None` for global lines).
pub(crate) fn by_origin(records: Vec<Value>) -> BTreeMap<Option<String>, Vec<Value>> {
    let mut out: BTreeMap<Option<String>, Vec<Value>> = BTreeMap::new();
    for r in records {
        let origin = r["origin"].as_str().map(str::to_string);
        out.entry(origin).or_default().push(r);
    }
    out
}

/// Re-assembles stored lines into `{nodes, edges, missing}` and checks the C4 schema.
pub(crate) fn document(records: &[Value]) -> Result<Value> {
    let pick =
        |key: &str| -> Vec<Value> { records.iter().filter_map(|r| r.get(key).cloned()).collect() };
    let doc = json!({"nodes": pick("node"), "edges": pick("edge"), "missing": pick("missing")});
    ontology::validate_document(&doc).map_err(|e| format!("index records invalid: {e}"))?;
    Ok(doc)
}

#[cfg(test)]
mod tests {
    use super::*;

    const AT: &str = "2026-11-10T00:00:00Z";

    fn views(files: &[(&str, &[u8])]) -> Views {
        let mut fragments = BTreeMap::new();
        for (path, bytes) in files {
            let sha = crate::digest(bytes);
            fragments.insert(
                path.to_string(),
                extract_file(path, &sha, Content::Bytes(bytes)),
            );
        }
        assemble(fragments, Builder::default(), ("gen-x", "base", "fs"), AT).unwrap()
    }

    #[test]
    fn map_view_carries_files_tags_and_schema_valid_edges() {
        let v = views(&[
            ("a.rs", b"// @MOAI:REQ REQ-C6B-01\nfn x() {}\n"),
            ("b.md", b"plain\n"),
        ]);
        let (header, records) = parse_view(&v.map).unwrap();
        assert_eq!(header["view"], "map");
        let doc = document(&records).unwrap();
        let mut ids: Vec<&str> = doc["nodes"]
            .as_array()
            .unwrap()
            .iter()
            .map(|n| n["id"].as_str().unwrap())
            .collect();
        ids.sort();
        assert_eq!(ids, ["file:a.rs", "file:b.md", "tag:a.rs#L1"]);
        let edge = &doc["edges"][0];
        assert_eq!(edge["kind"], "annotated_by");
        assert_eq!(edge["extractor"]["name"], MAP_EXTRACTOR);
        assert_eq!(edge["observed_at"], AT);
        // The graph view links the same tag to its requirement.
        let (_, graph) = parse_view(&v.graph).unwrap();
        let g = document(&graph).unwrap();
        assert_eq!(g["edges"][0]["kind"], "references_requirement");
    }

    fn map_misses(f: &Fragment) -> Vec<String> {
        f.map
            .missing
            .iter()
            .map(|m| format!("{}:{}", m["kind"], m["reason"]))
            .collect()
    }

    /// The tag scan limit, literal N and N+1: 262,144 bytes are scanned (the tag on
    /// line 1 is kept), 262,145 are reported as skipped with no tag node.
    #[test]
    fn tag_scan_limit_has_literal_n_and_n_plus_one_boundaries() {
        for (len, scanned) in [(262_144usize, true), (262_145, false)] {
            let mut bytes = b"<!-- @MX:NOTE edge -->\n".to_vec();
            bytes.resize(len, b'x');
            let f = extract_file("doc.md", "sha256:x", Content::Bytes(&bytes));
            assert_eq!(f.map.nodes.contains_key("tag:doc.md#L1"), scanned, "{len}");
            let skipped = vec![r#""tag_scan_skipped":"tag file exceeds the size limit""#];
            if scanned {
                assert!(map_misses(&f).is_empty(), "{len}");
            } else {
                assert_eq!(map_misses(&f), skipped, "{len}");
            }
        }
        // The build reads a non-code file only up to the same limit; code files are
        // always read for the link view.
        assert!(needs_bytes("doc.md", 262_144));
        assert!(!needs_bytes("doc.md", 262_145));
        assert!(needs_bytes("big.rs", 262_145));
    }

    /// #2442: a Rust file is always read (code files need bytes at any size), so the
    /// `Oversize` branch never skips a code-reference scan.
    #[test]
    fn code_files_are_read_at_any_size() {
        assert!(needs_bytes("src/huge.rs", u64::MAX));
        assert!(!needs_bytes("docs/huge.md", u64::MAX));
    }

    #[test]
    fn unread_entries_are_listed_as_missing_coverage() {
        let over = extract_file("big.md", "sha256:x", Content::Oversize);
        assert_eq!(
            map_misses(&over),
            [r#""tag_scan_skipped":"tag file exceeds the size limit""#]
        );
        let link = extract_file("ln.rs", "symlink-sha256:x", Content::NotRegular);
        assert_eq!(
            map_misses(&link),
            [r#""tag_scan_skipped":"not a regular file""#]
        );
        assert_eq!(link.graph.missing[0]["kind"], "unreadable_file");
        for f in [&over, &link] {
            assert_eq!(f.map.nodes.len(), 1, "only the file node");
        }
    }

    /// The records byte cap, literal N and N+1 (the index state read limit).
    #[test]
    fn records_byte_cap_has_literal_n_and_n_plus_one_boundaries() {
        assert!(within_limit(33_554_432, 33_554_432).is_ok());
        assert!(within_limit(33_554_433, 33_554_432).is_err());
    }

    #[test]
    fn stored_fragments_rebuild_identical_bytes() {
        let files: &[(&str, &[u8])] = &[
            (
                "a.rs",
                b"// @MOAI:REQ REQ-C6B-01\n#[test]\nfn req_c6b_01_x() {}\n",
            ),
            ("b.rs", b"// @MOAI:REQ REQ-C6B-01, bad\n"),
            ("c.bin", &[0xff, 0xfe]),
        ];
        let full = views(files);
        let (_, map) = parse_view(&full.map).unwrap();
        let (_, graph) = parse_view(&full.graph).unwrap();
        let mut map = by_origin(map);
        let mut graph = by_origin(graph);
        let mut fragments = BTreeMap::new();
        for (path, _) in files {
            let key = Some(path.to_string());
            let m = map.remove(&key).unwrap_or_default();
            let g = graph.remove(&key).unwrap_or_default();
            fragments.insert(path.to_string(), fragment_from_lines(&m, &g).unwrap());
        }
        let again = assemble(fragments, Builder::default(), ("gen-x", "base", "fs"), AT).unwrap();
        assert_eq!(full.map, again.map);
        assert_eq!(full.graph, again.graph);
    }
}
