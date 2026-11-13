//! Non-vector-first search (rc.2 C7 #2183, C7b #2408): the retrieval pipeline of the Core
//! Brain design (`page_136c8203…` L3716–3724).
//!
//! 1. validate the project (repository top) and read the index freshness (`index_query`);
//! 2. resolve exact paths before anything else (REQ-C7-01);
//! 3. search literal strings line by line over the published generation's files, or over the
//!    tracked files when no generation is published (lower coverage, design L3740);
//!    Code Map declarations: a literal equal to one argument of a `@MOAI:`/`@MX:` tag in
//!    the generation's map view cites that tag line (grade `map`);
//! 4. typed edges of the same generation's map and graph views (C6b #2347) are expanded
//!    from resolved seeds only — exact paths, exact node ids (`req:<id>`, `issue:<n>`, or a
//!    full node id) and matched declarations — within `max_hops` and `max_fanout` per node;
//!    every edge evidence span `file:<path>#L<n>@<sha256>` becomes a `graph` citation.
//!    Evidence is emitted one span per edge per round, in traversal order, so repeated
//!    evidence for one edge does not take the first slots of later traversed edges.
//!    Without a published generation with records, or when the records do not match the
//!    generation, the step is `unavailable` and the answer falls back to the lexical result
//!    (design L4300);
//! 5. exact results rank first, then map declarations, then graph evidence, then lexical
//!    matches; names and paths are never inferred (a lexical hit never seeds the expansion);
//! 6. results are deduplicated by (path, span);
//! 7. two caps cut the list — files and spans per file — and every cut is an `omitted` entry;
//! 8. before returning, each cited span is re-read and compared with the digest taken while
//!    collecting; a changed span is marked `stale`. `map`/`graph` citations are anchored
//!    to the generation's file digest, so a file edited after the index build is `stale`;
//! 9. the answer carries coverage, freshness and the omissions.
//!
//! Read-only: no file is written, no network is used.
use serde::Deserialize;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::{BTreeSet, VecDeque},
    fs,
    path::{Component, Path},
    process::Command,
};

type Result<T> = std::result::Result<T, String>;

/// Default and largest number of files in one answer.
pub const DEFAULT_MAX_FILES: usize = 200;
pub const MAX_FILES_LIMIT: usize = 2000;
/// Default and largest number of spans kept per file.
pub const DEFAULT_MAX_SPANS_PER_FILE: usize = 20;
pub const MAX_SPANS_PER_FILE_LIMIT: usize = 1000;
/// Bounds on the query itself.
pub const MAX_TERMS: usize = 64;
pub const MAX_TERM_BYTES: usize = 512;
/// Files larger than this are not scanned (recorded as omitted).
pub const MAX_FILE_BYTES: u64 = 4 * 1024 * 1024;
/// Hops of the typed-edge expansion (step 4): design L3026 expands only the relations a task
/// needs, starting from one or two hops — one hop by default, two at most.
pub const DEFAULT_MAX_HOPS: usize = 1;
pub const MAX_HOPS_LIMIT: usize = 2;
/// Edges followed from one node per hop. The design bounds fan-out without a number
/// (L3026, L3734); 20 by default and 200 at most is the #2408 lane decision, matching the
/// per-file span default.
pub const DEFAULT_MAX_FANOUT: usize = 20;
pub const MAX_FANOUT_LIMIT: usize = 200;

#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct Query {
    #[serde(default)]
    pub paths: Vec<String>,
    #[serde(default)]
    pub literals: Vec<String>,
}

#[derive(Deserialize, Default)]
#[serde(deny_unknown_fields)]
pub struct Limits {
    pub max_files: Option<usize>,
    pub max_spans_per_file: Option<usize>,
    pub max_hops: Option<usize>,
    pub max_fanout: Option<usize>,
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
pub struct SearchRequest {
    pub protocol_version: u32,
    #[serde(default)]
    pub query: Query,
    #[serde(default)]
    pub limits: Limits,
}

/// One cited span, as collected (step 3) and before the digest recheck (step 8).
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Hit {
    pub path: String,
    pub grade: &'static str,
    pub start_line: usize,
    pub end_line: usize,
    pub span_sha256: String,
    pub file_sha256: String,
    pub reason: String,
}

fn sha256_hex(bytes: &[u8]) -> String {
    format!("{:x}", Sha256::digest(bytes))
}

/// A repository-relative path of normal components only.
fn relative_ok(path: &str) -> bool {
    !path.is_empty()
        && Path::new(path)
            .components()
            .all(|c| matches!(c, Component::Normal(_)))
}

fn term_ok(term: &str) -> bool {
    !term.is_empty() && term.len() <= MAX_TERM_BYTES && !term.contains('\n')
}

/// Tracked files at HEAD, the fallback file list when no generation is published.
fn tracked_files(root: &Path) -> Vec<String> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd
        .arg("-C")
        .arg(root)
        .args(["--no-optional-locks", "ls-files", "-z"])
        .output();
    match out {
        Ok(o) if o.status.success() => o
            .stdout
            .split(|b| *b == 0)
            .filter(|p| !p.is_empty())
            .filter_map(|p| std::str::from_utf8(p).ok().map(str::to_string))
            .collect(),
        _ => Vec::new(),
    }
}

/// The bytes of lines `start..=end` (1-based), each with its terminator if present.
fn span_bytes(text: &[u8], start: usize, end: usize) -> &[u8] {
    let mut line = 1;
    let mut from = None;
    for (i, b) in text.iter().enumerate() {
        if line == start && from.is_none() {
            from = Some(i);
        }
        if *b == b'\n' {
            if line == end {
                return &text[from.unwrap_or(i)..=i];
            }
            line += 1;
        }
    }
    match from {
        Some(f) if line == end => &text[f..],
        _ => &[],
    }
}

/// Lines (1-based) of `text` that contain `term` as a byte string.
fn matching_lines(text: &[u8], term: &[u8]) -> Vec<usize> {
    text.split(|b| *b == b'\n')
        .enumerate()
        .filter(|(_, l)| l.windows(term.len()).any(|w| w == term))
        .map(|(i, _)| i + 1)
        .collect()
}

/// Steps 2–3 and 5–6: exact paths first, then literal matches, deduplicated by (path, span).
/// Returns the hits in answer order and the omissions met while reading files.
pub fn collect(root: &Path, files: &[String], query: &Query) -> (Vec<Hit>, Vec<Value>) {
    let mut hits = Vec::new();
    let mut omitted = Vec::new();
    let mut seen: BTreeSet<(String, usize, usize)> = BTreeSet::new();
    let mut push = |hits: &mut Vec<Hit>, h: Hit| {
        if seen.insert((h.path.clone(), h.start_line, h.end_line)) {
            hits.push(h);
        }
    };
    // Step 1 (access) applies to exact paths too: only a regular file in the searched file
    // list is read; anything else — outside the list, a symlink, a directory — is an omission,
    // so its existence and digest never reach the answer (pre-audit B1).
    let listed: BTreeSet<&str> = files.iter().map(String::as_str).collect();
    for path in &query.paths {
        if !listed.contains(path.as_str()) {
            omitted
                .push(json!({"path": path, "reason": "exact path not in the searched file list"}));
            continue;
        }
        match fs::symlink_metadata(root.join(path)) {
            Ok(m) if m.is_file() => {}
            _ => {
                omitted.push(json!({"path": path, "reason": "exact path is not a regular file"}));
                continue;
            }
        }
        match fs::read(root.join(path)) {
            Ok(bytes) => {
                let lines = bytes.split(|b| *b == b'\n').count().max(1);
                let end = if bytes.ends_with(b"\n") {
                    lines - 1
                } else {
                    lines
                }
                .max(1);
                push(
                    &mut hits,
                    Hit {
                        path: path.clone(),
                        grade: "exact",
                        start_line: 1,
                        end_line: end,
                        span_sha256: sha256_hex(&bytes),
                        file_sha256: sha256_hex(&bytes),
                        reason: "exact path".into(),
                    },
                );
            }
            Err(_) => omitted.push(json!({"path": path, "reason": "exact path not found"})),
        }
    }
    if query.literals.is_empty() {
        return (hits, omitted);
    }
    let mut sorted: Vec<&String> = files.iter().collect();
    sorted.sort();
    for path in sorted {
        let full = root.join(path);
        let Ok(meta) = fs::symlink_metadata(&full) else {
            continue;
        };
        if !meta.is_file() {
            continue;
        }
        if meta.len() > MAX_FILE_BYTES {
            omitted.push(json!({"path": path, "reason": "file larger than the scan limit"}));
            continue;
        }
        let Ok(bytes) = fs::read(&full) else {
            omitted.push(json!({"path": path, "reason": "file unreadable"}));
            continue;
        };
        let file_sha = sha256_hex(&bytes);
        let mut lines: BTreeSet<usize> = BTreeSet::new();
        for term in &query.literals {
            lines.extend(matching_lines(&bytes, term.as_bytes()));
        }
        for line in lines {
            push(
                &mut hits,
                Hit {
                    path: path.clone(),
                    grade: "lexical",
                    start_line: line,
                    end_line: line,
                    span_sha256: sha256_hex(span_bytes(&bytes, line, line)),
                    file_sha256: file_sha.clone(),
                    reason: "literal match".into(),
                },
            );
        }
    }
    (hits, omitted)
}

/// Step 7: keep at most `max_files` distinct files and `max_spans` spans per file, in answer
/// order; every dropped span becomes one `omitted` entry naming the cap.
pub fn cap(hits: Vec<Hit>, max_files: usize, max_spans: usize) -> (Vec<Hit>, Vec<Value>) {
    let mut kept = Vec::new();
    let mut omitted = Vec::new();
    let mut files: Vec<String> = Vec::new();
    let mut per_file: std::collections::BTreeMap<String, usize> = Default::default();
    for h in hits {
        let known = files.contains(&h.path);
        if !known && files.len() >= max_files {
            omitted
                .push(json!({"path": h.path, "start_line": h.start_line, "reason": "max_files"}));
            continue;
        }
        let n = per_file.entry(h.path.clone()).or_insert(0);
        if *n >= max_spans {
            omitted.push(
                json!({"path": h.path, "start_line": h.start_line, "reason": "max_spans_per_file"}),
            );
            continue;
        }
        *n += 1;
        if !known {
            files.push(h.path.clone());
        }
        kept.push(h);
    }
    (kept, omitted)
}

/// `file:<path>#L<n>@<sha256>` — an edge evidence ref that cites one file line. Other
/// shapes (a commit trailer) cite no file span and yield `None`.
fn evidence_span(evidence: &str) -> Option<(&str, usize, &str)> {
    let rest = evidence.strip_prefix("file:")?;
    let (location, sha) = rest.rsplit_once('@')?;
    let (path, line) = location.rsplit_once("#L")?;
    let line: usize = line.parse().ok().filter(|n| *n > 0)?;
    Some((path, line, sha))
}

/// The file a node belongs to: `file:<p>`, `tag:<p>#L<n>`, `test:<p>::<name>`.
fn node_file(id: &str) -> Option<&str> {
    if let Some(p) = id.strip_prefix("file:") {
        return Some(p);
    }
    if let Some(rest) = id.strip_prefix("tag:") {
        return rest.rsplit_once("#L").map(|(p, _)| p);
    }
    id.strip_prefix("test:")
        .and_then(|rest| rest.split_once("::").map(|(p, _)| p))
}

/// Hops and fan-out of one expansion.
#[derive(Clone, Copy, Debug)]
pub struct Bounds {
    pub max_hops: usize,
    pub max_fanout: usize,
}

/// Reads a listed regular file once per answer and keeps its digest and bytes.
struct Files<'a> {
    root: &'a Path,
    listed: BTreeSet<&'a str>,
    cache: std::collections::BTreeMap<String, Option<(String, Vec<u8>)>>,
}

impl Files<'_> {
    fn get(&mut self, path: &str) -> Option<&(String, Vec<u8>)> {
        if !self.listed.contains(path) {
            return None;
        }
        let root = self.root;
        self.cache
            .entry(path.to_string())
            .or_insert_with(|| {
                let full = root.join(path);
                let meta = fs::symlink_metadata(&full).ok()?;
                if !meta.is_file() || meta.len() > MAX_FILE_BYTES {
                    return None;
                }
                let bytes = fs::read(&full).ok()?;
                Some((sha256_hex(&bytes), bytes))
            })
            .as_ref()
    }
}

/// A citation anchored to the generation: `file_sha256` is the digest the index recorded
/// for the file (the evidence ref's or the tag node's revision), never a fresh read. The
/// span digest is taken now only when the file still has those bytes; otherwise it stays
/// empty and step 8 reports the citation `stale`.
fn anchored(
    files: &mut Files<'_>,
    path: &str,
    line: usize,
    anchor: &str,
    grade: &'static str,
    reason: String,
) -> std::result::Result<Hit, Value> {
    // The index stamps `sha256:<hex>` (`crate::digest`); citations carry bare hex.
    let anchor = anchor.strip_prefix("sha256:").unwrap_or(anchor);
    if !files.listed.contains(path) {
        return Err(json!({"path": path, "reason": "evidence outside the searched file list"}));
    }
    let span_sha256 = match files.get(path) {
        Some((sha, bytes)) if sha == anchor => {
            let span = span_bytes(bytes, line, line);
            if span.is_empty() {
                String::new()
            } else {
                sha256_hex(span)
            }
        }
        _ => String::new(),
    };
    Ok(Hit {
        path: path.to_string(),
        grade,
        start_line: line,
        end_line: line,
        span_sha256,
        file_sha256: anchor.to_string(),
        reason,
    })
}

/// Steps 3 (Code Map declarations) and 4 (typed-edge expansion) over one generation's
/// `map` and `graph` documents (`{nodes, edges, missing}`, as `index_records` returns
/// them). Seeds are resolved only: listed exact paths, exact node ids and declarations
/// whose tag argument equals a literal. Every node expands at most `max_fanout`
/// untraversed edges per hop (sorted by kind, far endpoint, edge id); the rest and the
/// edges left beyond `max_hops` are `omitted` entries with the count dropped.
/// Returns the citations in answer order, the omissions and the expansion counts.
pub fn structural(
    root: &Path,
    files: &[String],
    query: &Query,
    map: &Value,
    graph: &Value,
    bounds: Bounds,
) -> (Vec<Hit>, Vec<Value>, Value) {
    let mut io = Files {
        root,
        listed: files.iter().map(String::as_str).collect(),
        cache: Default::default(),
    };
    let mut hits = Vec::new();
    let mut omitted = Vec::new();
    let nodes: Vec<&Value> = [map, graph]
        .iter()
        .flat_map(|d| d["nodes"].as_array().into_iter().flatten())
        .collect();
    let edges: Vec<&Value> = [map, graph]
        .iter()
        .flat_map(|d| d["edges"].as_array().into_iter().flatten())
        .collect();
    let ids: BTreeSet<&str> = nodes.iter().filter_map(|n| n["id"].as_str()).collect();

    let mut seeds: BTreeSet<String> = BTreeSet::new();
    for path in &query.paths {
        if io.listed.contains(path.as_str()) {
            seeds.extend(
                ids.iter()
                    .filter(|id| node_file(id) == Some(path.as_str()))
                    .map(|id| id.to_string()),
            );
        }
    }
    for lit in &query.literals {
        for candidate in [lit.clone(), format!("req:{lit}"), format!("issue:{lit}")] {
            if ids.contains(candidate.as_str()) {
                seeds.insert(candidate);
            }
        }
    }
    // Step 3: a declaration matches when one of its arguments equals a literal exactly.
    for node in &nodes {
        if node["kind"] != "tag" {
            continue;
        }
        let props = &node["properties"];
        let args = props["args"].as_str().unwrap_or("");
        let matched = args
            .split(|c: char| c.is_whitespace() || c == ',')
            .any(|a| !a.is_empty() && query.literals.iter().any(|l| l == a));
        let (Some(id), Some(path), Some(anchor)) = (
            node["id"].as_str(),
            props["path"].as_str(),
            node["revision"].as_str(),
        ) else {
            continue;
        };
        let Some(line) = id
            .rsplit_once("#L")
            .and_then(|(_, n)| n.parse::<usize>().ok())
        else {
            continue;
        };
        if !matched {
            continue;
        }
        seeds.insert(id.to_string());
        let reason = format!(
            "Code Map declaration {}:{} {args}",
            props["namespace"].as_str().unwrap_or(""),
            props["tag_kind"].as_str().unwrap_or("")
        );
        match anchored(&mut io, path, line, anchor, "map", reason) {
            Ok(h) => hits.push(h),
            Err(o) => omitted.push(o),
        }
    }

    // Step 4: breadth-first over the typed edges, both directions.
    let mut edge_citations = Vec::new();
    let mut incident: std::collections::BTreeMap<&str, Vec<usize>> = Default::default();
    for (i, e) in edges.iter().enumerate() {
        for end in ["source_id", "target_id"] {
            if let Some(id) = e[end].as_str() {
                incident.entry(id).or_default().push(i);
            }
        }
    }
    let far = |i: usize, from: &str| -> String {
        let e = edges[i];
        let s = e["source_id"].as_str().unwrap_or("");
        let t = e["target_id"].as_str().unwrap_or("");
        if s == from { t } else { s }.to_string()
    };
    let mut traversed: BTreeSet<usize> = BTreeSet::new();
    let mut reached: BTreeSet<String> = seeds.clone();
    let mut frontier: Vec<String> = seeds.iter().cloned().collect();
    for hop in 1..=bounds.max_hops {
        let mut next = Vec::new();
        for node in &frontier {
            let mut open: Vec<usize> = incident
                .get(node.as_str())
                .into_iter()
                .flatten()
                .copied()
                .filter(|i| !traversed.contains(i))
                .collect();
            open.sort_by_key(|i| {
                (
                    edges[*i]["kind"].as_str().unwrap_or("").to_string(),
                    far(*i, node),
                    edges[*i]["edge_id"].as_str().unwrap_or("").to_string(),
                )
            });
            open.dedup();
            if open.len() > bounds.max_fanout {
                omitted.push(json!({
                    "node": node,
                    "reason": "max_fanout",
                    "dropped_edges": open.len() - bounds.max_fanout,
                }));
                open.truncate(bounds.max_fanout);
            }
            for i in open {
                traversed.insert(i);
                let e = edges[i];
                let kind = e["kind"].as_str().unwrap_or("");
                let reason = format!(
                    "{kind}: {} -> {} (hop {hop})",
                    e["source_id"].as_str().unwrap_or(""),
                    e["target_id"].as_str().unwrap_or("")
                );
                let mut citations = VecDeque::new();
                for ev in e["evidence_refs"].as_array().into_iter().flatten() {
                    let Some((path, line, sha)) = ev.as_str().and_then(evidence_span) else {
                        continue;
                    };
                    match anchored(&mut io, path, line, sha, "graph", reason.clone()) {
                        Ok(h) => citations.push_back(h),
                        Err(o) => omitted.push(o),
                    }
                }
                edge_citations.push(citations);
                let other = far(i, node);
                if reached.insert(other.clone()) {
                    next.push(other);
                }
            }
        }
        frontier = next;
    }
    for node in &frontier {
        let left = incident
            .get(node.as_str())
            .into_iter()
            .flatten()
            .filter(|i| !traversed.contains(*i))
            .collect::<BTreeSet<_>>()
            .len();
        if left > 0 {
            omitted.push(json!({"node": node, "reason": "max_hops", "dropped_edges": left}));
        }
    }
    // Give each traversed edge its first citation before an evidence-heavy edge gets
    // another slot. Keep traversal order within each round and all public caps intact.
    loop {
        let mut emitted = false;
        for citations in &mut edge_citations {
            if let Some(hit) = citations.pop_front() {
                hits.push(hit);
                emitted = true;
            }
        }
        if !emitted {
            break;
        }
    }
    let stats = json!({
        "seeds": seeds.len(),
        "edges_traversed": traversed.len(),
        "nodes_reached": reached.len(),
    });
    (hits, omitted, stats)
}

/// Missing-coverage entries of both views, counted by kind (`tag_scan_skipped`,
/// `unreadable_file`, …), so an answer says what the graph could not see.
fn missing_counts(map: &Value, graph: &Value) -> Value {
    let mut counts: std::collections::BTreeMap<String, usize> = Default::default();
    for doc in [map, graph] {
        for m in doc["missing"].as_array().into_iter().flatten() {
            *counts
                .entry(m["kind"].as_str().unwrap_or("unknown").to_string())
                .or_insert(0) += 1;
        }
    }
    json!(counts)
}

/// Step 8: re-read every cited span; `fresh` when its bytes still hash to the collected
/// digest, `stale` otherwise (changed or gone). A `map`/`graph` citation is fresh only
/// while the whole file still has the generation's digest and the span was readable.
pub fn recheck(root: &Path, hits: &[Hit]) -> Vec<&'static str> {
    hits.iter()
        .map(|h| {
            let Ok(bytes) = fs::read(root.join(&h.path)) else {
                return "stale";
            };
            if h.grade == "map" || h.grade == "graph" {
                return if !h.span_sha256.is_empty()
                    && sha256_hex(&bytes) == h.file_sha256
                    && sha256_hex(span_bytes(&bytes, h.start_line, h.end_line)) == h.span_sha256
                {
                    "fresh"
                } else {
                    "stale"
                };
            }
            let now = if h.grade == "exact" {
                sha256_hex(&bytes)
            } else {
                sha256_hex(span_bytes(&bytes, h.start_line, h.end_line))
            };
            if now == h.span_sha256 {
                "fresh"
            } else {
                "stale"
            }
        })
        .collect()
}

pub fn search(project: &Path, req: &SearchRequest) -> Result<Value> {
    if req.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let q = &req.query;
    if q.paths.is_empty() && q.literals.is_empty() {
        return Err("search refused: give at least one path or literal".into());
    }
    if q.paths.len() + q.literals.len() > MAX_TERMS {
        return Err("search refused: too many query terms".into());
    }
    if !q.paths.iter().all(|p| relative_ok(p)) {
        return Err("search refused: a path must be relative inside the repository".into());
    }
    if !q.literals.iter().all(|t| term_ok(t)) {
        return Err(
            "search refused: a literal must be non-empty, one line, at most 512 bytes".into(),
        );
    }
    let max_files = req.limits.max_files.unwrap_or(DEFAULT_MAX_FILES);
    let max_spans = req
        .limits
        .max_spans_per_file
        .unwrap_or(DEFAULT_MAX_SPANS_PER_FILE);
    if max_files == 0 || max_files > MAX_FILES_LIMIT {
        return Err("search refused: max_files must be 1..=2000".into());
    }
    if max_spans == 0 || max_spans > MAX_SPANS_PER_FILE_LIMIT {
        return Err("search refused: max_spans_per_file must be 1..=1000".into());
    }
    let bounds = Bounds {
        max_hops: req.limits.max_hops.unwrap_or(DEFAULT_MAX_HOPS),
        max_fanout: req.limits.max_fanout.unwrap_or(DEFAULT_MAX_FANOUT),
    };
    if bounds.max_hops == 0 || bounds.max_hops > MAX_HOPS_LIMIT {
        return Err("search refused: max_hops must be 1..=2".into());
    }
    if bounds.max_fanout == 0 || bounds.max_fanout > MAX_FANOUT_LIMIT {
        return Err("search refused: max_fanout must be 1..=200".into());
    }
    let root = fs::canonicalize(project).map_err(|_| "path unavailable".to_string())?;
    // Step 1: the index freshness answer is reported as it is, never re-derived here.
    let freshness = crate::index::index_query(
        &root,
        &crate::index::IndexQueryRequest {
            protocol_version: 1,
        },
    )?;
    let (generation_id, files, file_source) = match crate::index::published_files(&root) {
        Some((id, list)) => (
            Value::String(id),
            list.into_iter().map(|f| f.path).collect::<Vec<_>>(),
            "index_generation",
        ),
        None => (Value::Null, tracked_files(&root), "tracked_files_fallback"),
    };
    let (collected, mut omitted) = collect(&root, &files, q);
    // Steps 3–4 over the same generation's views; anything else is `unavailable` and the
    // answer stays lexical (design L4300).
    let records = match generation_id.as_str() {
        None => Err("no published index generation".to_string()),
        Some(gen) => {
            let answer = crate::index::index_records(
                &root,
                &crate::index::IndexQueryRequest {
                    protocol_version: 1,
                },
            )?;
            if answer["generation_id"].as_str() != Some(gen) {
                Err("generation changed during the search".into())
            } else if answer["map"].is_object() && answer["graph"].is_object() {
                Ok((answer["map"].clone(), answer["graph"].clone()))
            } else {
                Err(answer["reason"]
                    .as_str()
                    .unwrap_or("generation has no map/graph records")
                    .to_string())
            }
        }
    };
    let (structural_hits, graph_coverage) = match &records {
        Ok((map, graph)) => {
            let (h, o, stats) = structural(&root, &files, q, map, graph, bounds);
            omitted.extend(o);
            (
                h,
                json!({
                    "state": "available",
                    "views": ["map", "graph"],
                    "max_hops": bounds.max_hops,
                    "max_fanout": bounds.max_fanout,
                    "seeds": stats["seeds"],
                    "edges_traversed": stats["edges_traversed"],
                    "nodes_reached": stats["nodes_reached"],
                    "missing": missing_counts(map, graph),
                }),
            )
        }
        Err(reason) => (
            Vec::new(),
            json!({"state": "unavailable", "reason": reason}),
        ),
    };
    // Step 5–6: exact, then map and graph evidence, then lexical; one citation per span.
    let mut seen: BTreeSet<(String, usize, usize)> = BTreeSet::new();
    let (exact, lexical): (Vec<Hit>, Vec<Hit>) =
        collected.into_iter().partition(|h| h.grade == "exact");
    let hits: Vec<Hit> = exact
        .into_iter()
        .chain(structural_hits)
        .chain(lexical)
        .filter(|h| seen.insert((h.path.clone(), h.start_line, h.end_line)))
        .collect();
    let (hits, cut) = cap(hits, max_files, max_spans);
    omitted.extend(cut);
    let states = recheck(&root, &hits);
    let results: Vec<Value> = hits
        .iter()
        .zip(states)
        .map(|(h, state)| {
            json!({
                "path": h.path,
                "grade": h.grade,
                "start_line": h.start_line,
                "end_line": h.end_line,
                "span_sha256": h.span_sha256,
                "file_sha256": h.file_sha256,
                "reason": h.reason,
                "state": state,
            })
        })
        .collect();
    Ok(json!({
        "protocol_version": 1,
        "results": results,
        "omitted": omitted,
        "coverage": {
            "file_source": file_source,
            "generation_id": generation_id,
            "files_considered": files.len(),
            "graph": graph_coverage,
        },
        "freshness": freshness["state"],
        "network": "disabled",
        "writes": false,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn hit(path: &str, line: usize) -> Hit {
        Hit {
            path: path.into(),
            grade: "lexical",
            start_line: line,
            end_line: line,
            span_sha256: String::new(),
            file_sha256: String::new(),
            reason: String::new(),
        }
    }

    #[test]
    fn max_files_keeps_exactly_n_distinct_files() {
        let hits = vec![hit("a", 1), hit("b", 1), hit("c", 1), hit("a", 2)];
        // N = 2: a and b kept (a twice), c is the N+1th file and is omitted.
        let (kept, omitted) = cap(hits.clone(), 2, 20);
        let paths: Vec<&str> = kept.iter().map(|h| h.path.as_str()).collect();
        assert_eq!(paths, ["a", "b", "a"]);
        assert_eq!(omitted.len(), 1);
        assert_eq!(omitted[0]["path"], "c");
        assert_eq!(omitted[0]["reason"], "max_files");
        // N = 3: nothing is cut.
        assert_eq!(cap(hits, 3, 20).1.len(), 0);
    }

    #[test]
    fn max_spans_per_file_keeps_exactly_n_spans() {
        let hits = vec![hit("a", 1), hit("a", 2), hit("a", 3)];
        let (kept, omitted) = cap(hits.clone(), 200, 2);
        assert_eq!(kept.len(), 2);
        assert_eq!(omitted.len(), 1);
        assert_eq!(omitted[0]["start_line"], 3);
        assert_eq!(omitted[0]["reason"], "max_spans_per_file");
        assert_eq!(cap(hits, 200, 3).1.len(), 0);
    }

    #[test]
    fn lines_and_spans_are_byte_exact() {
        let text = b"alpha\nbeta gamma\ngamma\n";
        assert_eq!(matching_lines(text, b"gamma"), [2, 3]);
        assert_eq!(matching_lines(text, b"Gamma"), Vec::<usize>::new());
        assert_eq!(span_bytes(text, 2, 2), b"beta gamma\n");
        assert_eq!(span_bytes(b"x\nlast", 2, 2), b"last");
    }

    #[test]
    fn a_span_changed_after_collection_is_stale() {
        let dir = tempfile::TempDir::new().unwrap();
        fs::write(dir.path().join("f.txt"), "keep\nneedle here\n").unwrap();
        let query = Query {
            paths: vec![],
            literals: vec!["needle".into()],
        };
        let (hits, _) = collect(dir.path(), &["f.txt".into()], &query);
        assert_eq!(hits.len(), 1);
        assert_eq!(recheck(dir.path(), &hits), ["fresh"]);
        // The file changes between collection and return.
        fs::write(dir.path().join("f.txt"), "keep\nneedle moved\n").unwrap();
        assert_eq!(recheck(dir.path(), &hits), ["stale"]);
        fs::remove_file(dir.path().join("f.txt")).unwrap();
        assert_eq!(recheck(dir.path(), &hits), ["stale"]);
    }

    #[test]
    fn exact_paths_come_before_lexical_matches() {
        let dir = tempfile::TempDir::new().unwrap();
        fs::write(dir.path().join("a.txt"), "needle\n").unwrap();
        fs::write(dir.path().join("z.txt"), "other\n").unwrap();
        let query = Query {
            paths: vec!["z.txt".into()],
            literals: vec!["needle".into()],
        };
        let (hits, _) = collect(dir.path(), &["a.txt".into(), "z.txt".into()], &query);
        // z.txt sorts after a.txt, yet the exact path is first (REQ-C7-01).
        assert_eq!(hits[0].path, "z.txt");
        assert_eq!(hits[0].grade, "exact");
        assert_eq!(hits[1].path, "a.txt");
        assert_eq!(hits[1].grade, "lexical");
    }

    /// Pre-audit B1: an exact path outside the searched file list, or one that is not a
    /// regular file (a symlink to a file outside the repository), is omitted — never read,
    /// never hashed into the answer.
    #[test]
    fn exact_paths_outside_the_list_or_not_regular_files_are_omitted() {
        let outside = tempfile::TempDir::new().unwrap();
        fs::write(outside.path().join("secret.txt"), "s1\ns2\n").unwrap();
        let dir = tempfile::TempDir::new().unwrap();
        fs::write(dir.path().join(".env"), "TOKEN=x\n").unwrap();
        fs::write(dir.path().join("ok.txt"), "fine\n").unwrap();
        #[cfg(unix)]
        std::os::unix::fs::symlink(
            outside.path().join("secret.txt"),
            dir.path().join("link.txt"),
        )
        .unwrap();
        // link.txt is listed on every platform; where it does not exist (no symlink on
        // Windows) it is omitted as "not a regular file" just the same.
        let listed = vec!["ok.txt".to_string(), "link.txt".to_string()];
        let query = Query {
            paths: vec![".env".into(), "link.txt".into(), "ok.txt".into()],
            literals: vec![],
        };
        let (hits, omitted) = collect(dir.path(), &listed, &query);
        let paths: Vec<&str> = hits.iter().map(|h| h.path.as_str()).collect();
        assert_eq!(paths, ["ok.txt"]);
        let reasons: Vec<(String, String)> = omitted
            .iter()
            .map(|o| {
                (
                    o["path"].as_str().unwrap().into(),
                    o["reason"].as_str().unwrap().into(),
                )
            })
            .collect();
        assert!(reasons.contains(&(
            ".env".into(),
            "exact path not in the searched file list".into()
        )));
        assert!(reasons.contains(&("link.txt".into(), "exact path is not a regular file".into())));
        assert!(!serde_json::to_string(&omitted)
            .unwrap()
            .contains(&sha256_hex(b"s1\ns2\n")));
    }

    /// A three-file generation: `req:R` is referenced by a.rs line 1 and verified by a test
    /// at t.rs line 2; a.rs also carries a tag at line 3 (one more hop from `req:R`).
    fn graph_fixture() -> (tempfile::TempDir, Vec<String>, Value, Value) {
        let dir = tempfile::TempDir::new().unwrap();
        let a = b"// @MOAI:REQ R\nfn a() {}\n// @MX:NOTE keep\n".to_vec();
        let t = b"#[test]\nfn req_r() {}\n".to_vec();
        fs::write(dir.path().join("a.rs"), &a).unwrap();
        fs::write(dir.path().join("t.rs"), &t).unwrap();
        let (sa, st) = (crate::digest(&a), crate::digest(&t));
        let edge = |kind: &str, s: &str, t: &str, ev: String| {
            json!({"edge_id": format!("{kind}{s}{t}"), "kind": kind, "source_id": s,
                   "target_id": t, "evidence_refs": [ev]})
        };
        let map = json!({
            "nodes": [
                {"id": "file:a.rs", "kind": "file", "revision": sa, "properties": {"path": "a.rs"}},
                {"id": "tag:a.rs#L1", "kind": "tag", "revision": sa,
                 "properties": {"path": "a.rs", "namespace": "MOAI", "tag_kind": "REQ", "args": "R"}},
                {"id": "tag:a.rs#L3", "kind": "tag", "revision": sa,
                 "properties": {"path": "a.rs", "namespace": "MX", "tag_kind": "NOTE", "args": "keep"}},
            ],
            "edges": [edge("annotated_by", "file:a.rs", "tag:a.rs#L3", format!("file:a.rs#L3@{sa}"))],
            "missing": [{"kind": "tag_scan_skipped", "path": "big.bin"}],
        });
        let graph = json!({
            "nodes": [
                {"id": "req:R", "kind": "requirement", "revision": "unresolved"},
                {"id": "test:t.rs::req_r", "kind": "test", "revision": st},
            ],
            "edges": [
                edge("references_requirement", "file:a.rs", "req:R", format!("file:a.rs#L1@{sa}")),
                edge("verifies", "test:t.rs::req_r", "req:R", format!("file:t.rs#L2@{st}")),
                edge("mentions_issue", "commit:c", "issue:7", "commit:c#trailer:Refs@d".into()),
            ],
            "missing": [],
        });
        (dir, vec!["a.rs".into(), "t.rs".into()], map, graph)
    }

    fn lits(l: &[&str]) -> Query {
        Query {
            paths: vec![],
            literals: l.iter().map(|s| s.to_string()).collect(),
        }
    }

    fn cited(hits: &[Hit]) -> Vec<(String, &'static str, usize)> {
        hits.iter()
            .map(|h| (h.path.clone(), h.grade, h.start_line))
            .collect()
    }

    #[test]
    fn graph_citations_share_slots_before_repeating_an_edge() {
        let (dir, files, map, mut graph) = graph_fixture();
        let edge = &mut graph["edges"][0];
        let sha = crate::digest(&fs::read(dir.path().join("a.rs")).unwrap());
        edge["evidence_refs"] = json!([
            format!("file:a.rs#L1@{sha}"),
            format!("file:a.rs#L2@{sha}"),
            format!("file:a.rs#L3@{sha}")
        ]);
        let query = lits(&["req:R"]);
        let bounds = Bounds {
            max_hops: 1,
            max_fanout: 2,
        };
        let first = structural(dir.path(), &files, &query, &map, &graph, bounds);
        let second = structural(dir.path(), &files, &query, &map, &graph, bounds);
        assert_eq!(first, second);
        assert_eq!(
            cited(&first.0),
            [
                ("a.rs".into(), "graph", 1),
                ("t.rs".into(), "graph", 2),
                ("a.rs".into(), "graph", 2),
                ("a.rs".into(), "graph", 3),
            ]
        );
        assert_eq!(first.2["edges_traversed"], 2);
    }

    #[test]
    fn evidence_refs_parse_only_file_line_spans() {
        assert_eq!(
            evidence_span("file:src/a.rs#L12@abc"),
            Some(("src/a.rs", 12, "abc"))
        );
        assert_eq!(evidence_span("file:a.rs#L0@abc"), None);
        assert_eq!(evidence_span("commit:c#trailer:Refs@d"), None);
        assert_eq!(evidence_span("file:a.rs@abc"), None);
        assert_eq!(node_file("tag:x/y.rs#L4"), Some("x/y.rs"));
        assert_eq!(node_file("test:t.rs::m::f"), Some("t.rs"));
        assert_eq!(node_file("req:R"), None);
    }

    /// Hop bound N/N+1 with literal values: one hop from `req:R` cites a.rs:1 and t.rs:2 and
    /// records the a.rs edge it did not follow; two hops also cite a.rs:3 and drop nothing.
    #[test]
    fn max_hops_has_literal_one_and_two_boundaries() {
        let (dir, files, map, graph) = graph_fixture();
        let bounds = |h| Bounds {
            max_hops: h,
            max_fanout: 20,
        };
        let (hits, omitted, stats) = structural(
            dir.path(),
            &files,
            &lits(&["req:R"]),
            &map,
            &graph,
            bounds(1),
        );
        assert_eq!(
            cited(&hits),
            [("a.rs".into(), "graph", 1), ("t.rs".into(), "graph", 2)]
        );
        assert_eq!(
            omitted,
            [json!({"node": "file:a.rs", "reason": "max_hops", "dropped_edges": 1})]
        );
        assert_eq!(stats["edges_traversed"], 2);
        let (hits, omitted, _) = structural(
            dir.path(),
            &files,
            &lits(&["req:R"]),
            &map,
            &graph,
            bounds(2),
        );
        assert_eq!(
            cited(&hits),
            [
                ("a.rs".into(), "graph", 1),
                ("t.rs".into(), "graph", 2),
                ("a.rs".into(), "graph", 3)
            ]
        );
        assert!(omitted.is_empty(), "{omitted:?}");
    }

    /// Fan-out bound N/N+1: `req:R` has two edges; a fan-out of 1 keeps the first by kind
    /// (`references_requirement`) and records one dropped edge; 2 keeps both.
    #[test]
    fn max_fanout_has_literal_one_and_two_boundaries() {
        let (dir, files, map, graph) = graph_fixture();
        let bounds = |f| Bounds {
            max_hops: 1,
            max_fanout: f,
        };
        let (hits, omitted, _) =
            structural(dir.path(), &files, &lits(&["R"]), &map, &graph, bounds(1));
        // "R" is also a declaration argument: the map hit comes first.
        assert_eq!(
            cited(&hits),
            [("a.rs".into(), "map", 1), ("a.rs".into(), "graph", 1)]
        );
        assert!(
            omitted.contains(&json!({"node": "req:R", "reason": "max_fanout", "dropped_edges": 1}))
        );
        let (_, omitted, _) =
            structural(dir.path(), &files, &lits(&["R"]), &map, &graph, bounds(2));
        assert!(
            !omitted.iter().any(|o| o["reason"] == "max_fanout"),
            "{omitted:?}"
        );
    }

    /// Names are never inferred: a literal that only occurs as text seeds nothing, and a
    /// declaration matches a whole argument, not a substring.
    #[test]
    fn only_resolved_seeds_expand() {
        let (dir, files, map, graph) = graph_fixture();
        let b = Bounds {
            max_hops: 2,
            max_fanout: 20,
        };
        for q in [lits(&["fn a"]), lits(&["kee"]), lits(&["req:"])] {
            let (hits, omitted, stats) = structural(dir.path(), &files, &q, &map, &graph, b);
            assert!(hits.is_empty() && omitted.is_empty(), "{:?}", q.literals);
            assert_eq!(stats["seeds"], 0);
        }
        // An exact path seeds every node of that file.
        let q = Query {
            paths: vec!["t.rs".into()],
            literals: vec![],
        };
        let (hits, _, _) = structural(dir.path(), &files, &q, &map, &graph, b);
        assert_eq!(hits[0].path, "t.rs");
        assert_eq!(hits[0].start_line, 2);
        // A path outside the searched list seeds nothing.
        let (hits, _, _) = structural(dir.path(), &["a.rs".into()], &q, &map, &graph, b);
        assert!(hits.is_empty());
    }

    /// AC-2: graph citations are anchored to the generation digest. Fresh while the file
    /// keeps those bytes; edited after the index build (before collection) or between
    /// collection and return, the citation is `stale`.
    #[test]
    fn graph_citations_are_rechecked_against_the_generation_digest() {
        let (dir, files, map, graph) = graph_fixture();
        let b = Bounds {
            max_hops: 1,
            max_fanout: 20,
        };
        let (hits, _, _) = structural(dir.path(), &files, &lits(&["req:R"]), &map, &graph, b);
        assert_eq!(recheck(dir.path(), &hits), ["fresh", "fresh"]);
        // Edited between collection and return — the cited line 2 keeps its bytes, only
        // line 1 changes, so only the whole-file anchor catches it.
        fs::write(dir.path().join("t.rs"), "#[test] \nfn req_r() {}\n").unwrap();
        assert_eq!(recheck(dir.path(), &hits), ["fresh", "stale"]);
        // Edited before collection: a.rs line 1 is unchanged but the file is not the
        // indexed one, so the span is not vouched for.
        fs::write(
            dir.path().join("a.rs"),
            "// @MOAI:REQ R\nfn a() {}\n// changed\n",
        )
        .unwrap();
        let (hits, _, _) = structural(dir.path(), &files, &lits(&["req:R"]), &map, &graph, b);
        assert_eq!(hits[0].span_sha256, "");
        assert_eq!(recheck(dir.path(), &hits), ["stale", "stale"]);
    }

    #[test]
    fn missing_coverage_is_counted_by_kind() {
        let (_dir, _files, map, graph) = graph_fixture();
        assert_eq!(missing_counts(&map, &graph), json!({"tag_scan_skipped": 1}));
    }

    /// Pre-audit B3: a file of exactly MAX_FILE_BYTES is scanned; one byte more is omitted.
    #[test]
    fn the_scan_size_limit_is_4_mib() {
        let dir = tempfile::TempDir::new().unwrap();
        let mut at = vec![b'x'; 4_194_304 - 7];
        at.extend_from_slice(b"needle\n");
        assert_eq!(at.len(), 4_194_304);
        let mut over = at.clone();
        over.push(b'\n');
        fs::write(dir.path().join("at.txt"), &at).unwrap();
        fs::write(dir.path().join("over.txt"), &over).unwrap();
        let query = Query {
            paths: vec![],
            literals: vec!["needle".into()],
        };
        let (hits, omitted) = collect(dir.path(), &["at.txt".into(), "over.txt".into()], &query);
        assert_eq!(hits.len(), 1);
        assert_eq!(hits[0].path, "at.txt");
        assert_eq!(omitted.len(), 1);
        assert_eq!(omitted[0]["path"], "over.txt");
        assert_eq!(omitted[0]["reason"], "file larger than the scan limit");
    }
}
