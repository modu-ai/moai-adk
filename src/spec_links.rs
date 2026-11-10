//! Explicit SPEC link extraction — REQ ↔ code ↔ test (C5 #2181, REQ-C5-01).
//!
//! Reads three explicit sources from a git work tree and emits C4 ontology records
//! (`src/ontology.rs`). Every edge has `basis: "explicit"` and
//! `verification: "unreviewed"`; nothing is inferred (REQ-C5-01):
//!
//! | source | grammar | edge |
//! | --- | --- | --- |
//! | tag in a tracked code file | C2 grammar v1, kind `REQ` in either namespace (`@MOAI:REQ REQ-C5-01`, `@MX:REQ …`); args split on whitespace and commas | `file:<path>` → `references_requirement` → `req:<id>` |
//! | test name | a Rust fn whose attribute block carries `#[test]` or `#[tokio::test]`, or a Python `def test_…`; the name carries `req_<segs>_<digits>` (`req_08c_01` → `REQ-08C-01`) | `test:<path>::<fn>` → `verifies` → `req:<id>` |
//! | commit protocol line | exact raw-body lines in each commit reachable from HEAD: `Moai-Claim: <issue>#<id>` or `Refs: #<issue>` (ASCII decimal ids) | `commit:<sha>` → `mentions_issue` → `issue:<n>` |
//!
//! The REQ ↔ code ↔ test triangle closes through the requirement node; no
//! file ↔ test edge is ever emitted (the only candidates would be static or
//! semantic matches, which REQ-C5-01 forbids). Edge kinds follow the DR-07
//! appendix: `references_requirement` is a reference, not a claim of satisfaction
//! (`implements` is a candidate relation and stays out), and `mentions_issue` is a
//! mention, never resolution evidence.
//!
//! Everything that looks like a link but is not explicit enough lands in `missing`
//! (AC-2) instead of becoming an edge. Requirement and issue nodes carry
//! `revision: "unresolved"`: this extractor does not read SPEC documents or the forge.
//!
//! Library only; no CLI operation is added (lane decision, #2181).

use crate::{atlas_tag, digest, ontology};
use ontology::{Extractor, OntologyEdge, OntologyNode};
use serde_json::{json, Value};
use std::{
    collections::{BTreeMap, BTreeSet},
    fs,
    path::Path,
    process::Command,
};

type Result<T> = std::result::Result<T, String>;

pub const EXTRACTOR_NAME: &str = "spec-links";
pub const EXTRACTOR_VERSION: &str = "1";
const CODE_EXTENSIONS: [&str; 3] = ["rs", "py", "sh"];

/// `REQ-<SEG>(-<SEG>)*-<digits>`, segments `[A-Z0-9]+`.
pub fn is_req_id(s: &str) -> bool {
    let Some(rest) = s.strip_prefix("REQ-") else {
        return false;
    };
    let segs: Vec<&str> = rest.split('-').collect();
    segs.len() >= 2
        && segs.iter().all(|g| {
            !g.is_empty()
                && g.bytes()
                    .all(|b| b.is_ascii_uppercase() || b.is_ascii_digit())
        })
        && segs
            .last()
            .is_some_and(|g| g.bytes().all(|b| b.is_ascii_digit()))
}

/// REQ ids named by a test function: every `req_<segs>_<digits>` token, where the token
/// starts the name or follows `_`, the segments are `[a-z0-9]+`, at least one segment
/// precedes the number, and the number is the first all-digit segment.
pub fn req_ids_in_test_name(name: &str) -> Vec<String> {
    let parts: Vec<&str> = name.split('_').collect();
    let mut out = Vec::new();
    for (i, p) in parts.iter().enumerate() {
        if *p != "req" {
            continue;
        }
        let mut segs = Vec::new();
        for q in &parts[i + 1..] {
            if q.is_empty()
                || !q
                    .bytes()
                    .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit())
            {
                break;
            }
            segs.push(q.to_ascii_uppercase());
            if q.bytes().all(|b| b.is_ascii_digit()) {
                break;
            }
        }
        let numbered = segs
            .last()
            .is_some_and(|g| g.bytes().all(|b| b.is_ascii_digit()));
        // `is_req_id` requires a segment before the number, so `req_2195_x` is refused.
        if numbered {
            let id = format!("REQ-{}", segs.join("-"));
            if is_req_id(&id) && !out.contains(&id) {
                out.push(id);
            }
        }
    }
    out
}

/// Rust and Python test functions in one file: (line, name).
fn test_functions(path: &str, text: &str) -> Vec<(usize, String)> {
    let lines: Vec<&str> = text.lines().collect();
    let mut out = Vec::new();
    for (i, line) in lines.iter().enumerate() {
        let t = line.trim_start();
        if path.ends_with(".py") {
            if let Some(rest) = t.strip_prefix("def ") {
                let name: String = rest
                    .chars()
                    .take_while(|c| c.is_alphanumeric() || *c == '_')
                    .collect();
                if name.starts_with("test") {
                    out.push((i + 1, name));
                }
            }
            continue;
        }
        let Some(name) = rust_fn_name(t) else {
            continue;
        };
        // The attribute block directly above: attributes, doc and line comments.
        let mut j = i;
        let mut is_test = false;
        while j > 0 {
            j -= 1;
            let a = lines[j].trim();
            if a.starts_with("#[") {
                if a == "#[test]" || a.starts_with("#[tokio::test") {
                    is_test = true;
                }
            } else if !a.starts_with("//") {
                break;
            }
        }
        if is_test {
            out.push((i + 1, name));
        }
    }
    out
}

fn rust_fn_name(t: &str) -> Option<String> {
    let mut s = t;
    for prefix in ["pub(crate) ", "pub ", "async ", "unsafe "] {
        s = s.strip_prefix(prefix).unwrap_or(s);
    }
    let rest = s.strip_prefix("fn ")?;
    let name: String = rest
        .chars()
        .take_while(|c| c.is_alphanumeric() || *c == '_')
        .collect();
    (!name.is_empty()).then_some(name)
}

fn git(root: &Path, args: &[&str]) -> Result<String> {
    let out = Command::new("git")
        .arg("-C")
        .arg(root)
        .args(args)
        .output()
        .map_err(|_| "git unavailable".to_string())?;
    if !out.status.success() {
        return Err(format!("git {} failed", args.join(" ")));
    }
    String::from_utf8(out.stdout).map_err(|_| "git output is not UTF-8".to_string())
}

/// Project protocol lines are independent of Git's final trailer-block rules.
fn issue_trailer(line: &str) -> Option<(&'static str, &str)> {
    let decimal = |value: &str| !value.is_empty() && value.bytes().all(|b| b.is_ascii_digit());
    if let Some(value) = line.strip_prefix("Moai-Claim: ") {
        let (issue, claim) = value.split_once('#')?;
        return (decimal(issue) && decimal(claim)).then_some(("Moai-Claim", issue));
    }
    let issue = line.strip_prefix("Refs: #")?;
    decimal(issue).then_some(("Refs", issue))
}

/// Collected records before serialization: edges keyed by (kind, source, target) so
/// repeated explicit references merge their evidence into one edge.
#[derive(Default)]
struct Builder {
    nodes: BTreeMap<String, OntologyNode>,
    edges: BTreeMap<(String, String, String), BTreeSet<String>>,
    missing: Vec<Value>,
}

impl Builder {
    fn node(&mut self, id: &str, kind: &str, revision: &str, props: &[(&str, &str)]) {
        self.nodes
            .entry(id.to_string())
            .or_insert_with(|| OntologyNode {
                id: id.to_string(),
                kind: kind.to_string(),
                revision: revision.to_string(),
                properties: props
                    .iter()
                    .map(|(k, v)| (k.to_string(), v.to_string()))
                    .collect(),
                tombstone: None,
            });
    }

    fn requirement(&mut self, id: &str) {
        self.node(
            &format!("req:{id}"),
            "requirement",
            "unresolved",
            &[("revision_source", "not read: SPEC documents are outside C5")],
        );
    }

    fn edge(&mut self, kind: &str, source: &str, target: &str, evidence: String) {
        self.edges
            .entry((kind.to_string(), source.to_string(), target.to_string()))
            .or_default()
            .insert(evidence);
    }

    fn miss(&mut self, kind: &str, detail: Value) {
        let mut row = json!({"kind": kind});
        if let (Some(r), Some(d)) = (row.as_object_mut(), detail.as_object()) {
            r.extend(d.clone());
        }
        self.missing.push(row);
    }
}

/// Extracts explicit REQ ↔ code ↔ test links and commit → issue mentions from the git
/// work tree at `root`. `snapshot_id` names the snapshot the edges belong to (the
/// caller passes the HEAD commit); `observed_at` is RFC3339 and supplied by the caller,
/// so the output is a pure function of the tree, the history and the two inputs.
/// The result is validated with `ontology::validate_document` before it is returned.
pub fn extract(root: &Path, snapshot_id: &str, observed_at: &str) -> Result<Value> {
    if !ontology::is_rfc3339(observed_at) {
        return Err(format!("observed_at {observed_at:?} is not RFC3339"));
    }
    if snapshot_id.trim().is_empty() {
        return Err("snapshot_id must be non-empty".into());
    }
    let mut b = Builder::default();

    let tracked = git(root, &["ls-files", "-z"])?;
    let mut files: Vec<&str> = tracked.split('\0').filter(|p| !p.is_empty()).collect();
    files.sort();
    for path in files {
        let ext = path.rsplit('.').next().unwrap_or("");
        if !CODE_EXTENSIONS.contains(&ext) {
            continue;
        }
        if path.contains('#') || path.contains('@') {
            b.miss(
                "path_unrepresentable",
                json!({"path": path, "reason": "# or @ cannot appear in an evidence file-uri"}),
            );
            continue;
        }
        let full = root.join(path);
        let Ok(bytes) = fs::read(&full) else {
            b.miss(
                "unreadable_file",
                json!({"path": path, "reason": "read failed"}),
            );
            continue;
        };
        let file_digest = digest(&bytes);
        let Ok(text) = String::from_utf8(bytes) else {
            b.miss(
                "unreadable_file",
                json!({"path": path, "reason": "not UTF-8"}),
            );
            continue;
        };
        let evidence = |line: usize| format!("file:{path}#L{line}@{file_digest}");
        let file_id = format!("file:{path}");

        match atlas_tag::scan_path(&full) {
            Err(e) => b.miss("unreadable_file", json!({"path": path, "reason": e})),
            Ok(report) => {
                let file = &report["files"][0];
                for inv in file["invalid"].as_array().into_iter().flatten() {
                    b.miss(
                        "invalid_marker",
                        json!({"path": path, "line": inv["line"], "reason": inv["reason"]}),
                    );
                }
                for tag in file["nodes"].as_array().into_iter().flatten() {
                    if tag["kind"] != "REQ" {
                        continue;
                    }
                    let line = tag["line"].as_u64().unwrap_or(0) as usize;
                    let args = tag["args"].as_str().unwrap_or("");
                    let tokens: Vec<&str> = args
                        .split(|c: char| c.is_whitespace() || c == ',')
                        .filter(|t| !t.is_empty())
                        .collect();
                    if tokens.is_empty() {
                        b.miss("req_tag_without_id", json!({"path": path, "line": line}));
                    }
                    for t in tokens {
                        if !is_req_id(t) {
                            b.miss(
                                "malformed_req_ref",
                                json!({"path": path, "line": line, "token": t}),
                            );
                            continue;
                        }
                        b.node(&file_id, "file", &file_digest, &[("path", path)]);
                        b.requirement(t);
                        b.edge(
                            "references_requirement",
                            &file_id,
                            &format!("req:{t}"),
                            evidence(line),
                        );
                    }
                }
            }
        }

        for (line, name) in test_functions(path, &text) {
            let ids = req_ids_in_test_name(&name);
            let test_id = format!("test:{path}::{name}");
            for id in ids {
                b.node(
                    &test_id,
                    "test",
                    &file_digest,
                    &[("path", path), ("name", &name)],
                );
                b.requirement(&id);
                b.edge("verifies", &test_id, &format!("req:{id}"), evidence(line));
            }
        }
    }

    // Read raw bodies: a footer or a separate coauthor paragraph must not hide
    // explicit protocol lines. %x1e separates commits; %x1f separates SHA and body.
    let log = git(root, &["log", "--format=%H%x1f%B%x1e"])?;
    for rec in log.split('\x1e') {
        let Some((sha, body)) = rec.trim_start_matches('\n').split_once('\x1f') else {
            continue;
        };
        if sha.is_empty() {
            continue;
        }
        let msg_digest = digest(body.as_bytes());
        let commit_id = format!("commit:{sha}");
        let mut trailer_issues = BTreeSet::new();
        for t in body.lines() {
            if !t.starts_with("Moai-Claim:") && !t.starts_with("Refs:") {
                continue;
            }
            let Some((key, n)) = issue_trailer(t) else {
                b.miss(
                    "malformed_issue_trailer",
                    json!({"commit": sha, "trailer": t}),
                );
                continue;
            };
            trailer_issues.insert(n.to_string());
            b.node(&commit_id, "commit", sha, &[]);
            b.node(
                &format!("issue:{n}"),
                "issue",
                "unresolved",
                &[("revision_source", "not read: the forge is outside C5")],
            );
            b.edge(
                "mentions_issue",
                &commit_id,
                &format!("issue:{n}"),
                format!("commit:{sha}#trailer:{key}@{msg_digest}"),
            );
        }
        // A colonless `Refs #N` line is reported rather than inferred as a link.
        for line in body.lines() {
            if let Some(n) = line.trim().strip_prefix("Refs #") {
                let n = n.trim();
                if !n.is_empty()
                    && n.bytes().all(|c| c.is_ascii_digit())
                    && !trailer_issues.contains(n)
                {
                    b.miss("issue_ref_not_trailer", json!({"commit": sha, "issue": n}));
                }
            }
        }
    }

    // Triangle gaps (AC-2): a requirement referenced only from code or only from tests.
    let mut by_req: BTreeMap<String, (bool, bool)> = BTreeMap::new();
    for (kind, _, target) in b.edges.keys() {
        let e = by_req.entry(target.clone()).or_default();
        match kind.as_str() {
            "references_requirement" => e.0 = true,
            "verifies" => e.1 = true,
            _ => {}
        }
    }
    for (req, (code, test)) in by_req {
        if !req.starts_with("req:") {
            continue;
        }
        if code && !test {
            b.miss("req_without_test", json!({"requirement": req}));
        }
        if test && !code {
            b.miss("req_without_code", json!({"requirement": req}));
        }
    }

    let extractor = Extractor {
        name: EXTRACTOR_NAME.into(),
        version: EXTRACTOR_VERSION.into(),
    };
    let edges: Vec<OntologyEdge> = b
        .edges
        .into_iter()
        .map(|((kind, source, target), ev)| {
            let evidence_refs: Vec<String> = ev.into_iter().collect();
            OntologyEdge {
                edge_id: ontology::edge_id(&kind, &source, &target, &evidence_refs),
                kind,
                source_id: source,
                target_id: target,
                basis: "explicit".into(),
                verification: "unreviewed".into(),
                evidence_refs,
                extractor: extractor.clone(),
                snapshot_id: snapshot_id.to_string(),
                observed_at: observed_at.to_string(),
                valid_from: snapshot_id.to_string(),
                valid_to: None,
                uncertainty: None,
            }
        })
        .collect();
    let nodes: Vec<OntologyNode> = b.nodes.into_values().collect();
    let doc = json!({
        "nodes": nodes,
        "edges": edges,
        "missing": b.missing,
        "extractor": {"name": EXTRACTOR_NAME, "version": EXTRACTOR_VERSION},
        "snapshot_id": snapshot_id,
        "basis_inferred": 0,
    });
    ontology::validate_document(&doc)
        .map_err(|e| format!("extractor output failed the C4 schema: {e}"))?;
    Ok(doc)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn req_id_grammar() {
        for ok in ["REQ-C5-01", "REQ-08C-01", "REQ-SEC-HARDEN-002"] {
            assert!(is_req_id(ok), "{ok}");
        }
        for bad in [
            "REQ-01",
            "REQ-C5",
            "REQ-c5-01",
            "REQ--01",
            "REQ-C5-01x",
            "XREQ-C5-01",
        ] {
            assert!(!is_req_id(bad), "{bad}");
        }
    }

    #[test]
    fn test_name_tokens() {
        assert_eq!(
            req_ids_in_test_name("req_c5_01_only_explicit"),
            ["REQ-C5-01"]
        );
        assert_eq!(
            req_ids_in_test_name("test_req_08c_01_and_req_sec_harden_002"),
            ["REQ-08C-01", "REQ-SEC-HARDEN-002"]
        );
        assert!(req_ids_in_test_name("required_fields_are_kept").is_empty());
        assert!(
            req_ids_in_test_name("req_2195_kept").is_empty(),
            "a number needs a segment before it"
        );
        assert!(req_ids_in_test_name("req_c5_only").is_empty(), "no number");
    }
}
