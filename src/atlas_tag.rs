//! `@MOAI:` / `@MX:` dual-read tag scanner (C2 #2178).
//!
//! Grammar v1 (unifies the two historical spellings found in the design corpus —
//! top §9 L1184 `@MOAI:REQ SPEC-X/REQ-007` (space form) and Core Brain
//! L3472-3473 `@MOAI:SPEC: SPEC-PAYMENT-001` (colon form); the conflict and the
//! "define schema + parser version before adding sub-lines" note are recorded in
//! review-group2 §태그 문법):
//!
//! ```text
//! line    := prefix? "@MOAI:" payload | prefix? "@MX:" payload
//! payload := KIND ( ":" SEGMENT )*           // SEGMENT = SUB or args part
//! KIND    := [A-Z][A-Z0-9_]*                 // uppercase, at least one char
//! ```
//!
//! A payload without `:` carries args only (`@MOAI:REQ SPEC-X/REQ-007`); each
//! additional `:` segment is a SUB when it is a bare uppercase run (`SPEC:REQ`),
//! otherwise the final segment is the free-text ARGS (`SPEC: args with spaces`).
//! The `@MX:` namespace parses with the same rules and normalizes into the same
//! node shape — mixed-namespace files are read in full and duplicates are
//! diagnosed, never rewritten (REQ-C2-01/02: this scanner is read-only; the
//! migration writer belongs to a later card).
//!
//! Parser version: every node and report carries `parser_version = 1`; new
//! sub-lines MUST bump it (Core Brain L3475).

use serde_json::{json, Value};
use std::{
    collections::BTreeMap,
    fs,
    path::{Path, PathBuf},
};

type Result<T> = std::result::Result<T, String>;

pub const PARSER_VERSION: u32 = 1;
const FILE_LIMIT: u64 = 262_144;
const WALK_DEPTH: usize = 16;
const WALK_FILES: usize = 512;

/// One normalized tag node (both namespaces land here).
#[derive(Clone, Debug)]
pub struct TagNode {
    pub namespace: &'static str,
    pub kind: String,
    pub sub: Option<String>,
    pub args: Option<String>,
    pub line: usize,
    pub raw: String,
}

/// A line that carries a marker but does not parse under grammar v1.
#[derive(Clone, Debug)]
pub struct InvalidTag {
    pub namespace: &'static str,
    pub line: usize,
    pub raw: String,
    pub reason: String,
}

struct Parsed {
    namespace: &'static str,
    kind: String,
    sub: Option<String>,
    args: Option<String>,
}

fn uppercase_run(s: &str) -> Option<&str> {
    let end = s
        .char_indices()
        .find(|(_, c)| !c.is_ascii_uppercase() && !c.is_ascii_digit() && *c != '_')
        .map(|(i, _)| i)
        .unwrap_or(s.len());
    if end > 0 {
        Some(&s[..end])
    } else {
        None
    }
}

/// Parses one marker payload under grammar v1. The payload is everything after
/// the `@MOAI:` / `@MX:` marker on the line.
///
/// - one segment (`KIND args` or bare `KIND`): the args keep everything after the
///   uppercase kind run, separated by whitespace — junk glued to the kind is invalid.
/// - two segments (`KIND:X`): `X` is a SUB only when it is a bare uppercase run
///   (`@MOAI:SPEC:REQ` → sub REQ, no args); otherwise it is free-text args
///   (`@MX:SPEC: SPEC-PAYMENT-001`).
/// - three or more segments (`KIND:SUB:args…`): the second segment must be a bare
///   uppercase run (the SUB) and the rest — colons preserved — are the args.
fn parse_payload(payload: &str) -> Result<(String, Option<String>, Option<String>)> {
    let segments: Vec<&str> = payload.split(':').collect();
    if segments.len() == 1 {
        // `KIND args` (space-separated) or a bare KIND. The raw remainder must
        // start with whitespace; trimming before the check would accept junk.
        let kind = uppercase_run(segments[0]).unwrap_or("").to_string();
        if kind.is_empty() {
            let shown = segments[0];
            return Err(format!(
                "kind must be an uppercase run ([A-Z][A-Z0-9_]*), found {shown:?}"
            ));
        }
        let raw_rest = segments[0][kind.len()..].to_string();
        let rest = raw_rest.trim().to_string();
        if rest.is_empty() {
            return Ok((kind, None, None));
        }
        if !raw_rest.starts_with(' ') && !raw_rest.starts_with('\t') {
            return Err(format!(
                "junk {rest:?} directly after kind {kind:?}; separate args with a space"
            ));
        }
        return Ok((kind, None, Some(rest.to_string())));
    }
    let kind = uppercase_run(segments[0]).unwrap_or("").to_string();
    if kind.is_empty() || kind.len() != segments[0].len() {
        return Err(format!(
            "kind must be an uppercase run ([A-Z][A-Z0-9_]*), found {:?}",
            segments[0]
        ));
    }
    if segments.len() == 2 {
        let seg = segments[1].trim();
        if seg.is_empty() {
            return Err(format!("empty segment after kind {kind:?}"));
        }
        if let Some(sub) = uppercase_run(seg) {
            if sub.len() == seg.len() {
                return Ok((kind, Some(sub.to_string()), None));
            }
        }
        return Ok((kind, None, Some(seg.to_string())));
    }
    let sub_raw = segments[1];
    let sub = uppercase_run(sub_raw).unwrap_or("").to_string();
    if sub.is_empty() || sub.len() != sub_raw.len() {
        return Err(format!(
            "invalid sub {sub_raw:?} (must be a bare uppercase run)"
        ));
    }
    let args = segments[2..].join(":");
    let args = args.trim();
    if args.is_empty() {
        Ok((kind, Some(sub.to_string()), None))
    } else {
        Ok((kind, Some(sub.to_string()), Some(args.to_string())))
    }
}

fn parse_line(line: &str) -> Option<Result<Parsed>> {
    let marker = if let Some(pos) = line.find("@MOAI:") {
        let payload = line[pos + "@MOAI:".len()..].to_string();
        ("MOAI", payload)
    } else if let Some(pos) = line.find("@MX:") {
        let payload = line[pos + "@MX:".len()..].to_string();
        ("MX", payload)
    } else {
        return None;
    };
    Some(parse_payload(&marker.1).map(|(kind, sub, args)| Parsed {
        namespace: marker.0,
        kind,
        sub,
        args,
    }))
}

fn scan_file(path: &Path) -> Result<Value> {
    let bytes = fs::read(path).map_err(|_| "tag file read failed".to_string())?;
    if bytes.len() as u64 > FILE_LIMIT {
        return Err("tag file exceeds the size limit".into());
    }
    let text = String::from_utf8(bytes).map_err(|_| "tag file is not UTF-8".to_string())?;
    let mut nodes = Vec::new();
    let mut invalid = Vec::new();
    for (idx, line) in text.lines().enumerate() {
        if let Some(parsed) = parse_line(line) {
            match parsed {
                Ok(p) => nodes.push((idx + 1, line.to_string(), p)),
                Err(reason) => invalid.push((idx + 1, line.to_string(), reason)),
            }
        }
    }
    Ok(json!({
        "nodes": nodes.into_iter().map(|(line, raw, p)| json!({
            "namespace": p.namespace,
            "kind": p.kind,
            "sub": p.sub,
            "args": p.args,
            "line": line,
            "raw": raw,
            "parser_version": PARSER_VERSION,
        })).collect::<Vec<_>>(),
        "invalid": invalid.into_iter().map(|(line, raw, reason)| json!({
            "line": line, "raw": raw, "reason": reason,
        })).collect::<Vec<_>>(),
    }))
}

fn walk_files(root: &Path, out: &mut Vec<PathBuf>, depth: usize) -> Result<()> {
    if depth > WALK_DEPTH {
        return Err("tag scan nesting limit exceeded".into());
    }
    for entry in fs::read_dir(root).map_err(|_| "tag scan directory unavailable".to_string())? {
        let path = entry
            .map_err(|_| "tag scan entry unavailable".to_string())?
            .path();
        let meta =
            fs::symlink_metadata(&path).map_err(|_| "tag scan metadata unavailable".to_string())?;
        if meta.file_type().is_symlink() {
            continue; // links are skipped, never followed (read-only scan)
        }
        if meta.is_dir() {
            walk_files(&path, out, depth + 1)?;
        } else if meta.is_file() && meta.len() <= FILE_LIMIT {
            out.push(path);
        }
    }
    if out.len() > WALK_FILES {
        return Err("tag scan file limit exceeded".into());
    }
    Ok(())
}

/// Read-only dual-read scan (REQ-C2-01/02): reads both `@MOAI:` and `@MX:`
/// markers, normalizes them into one node shape, diagnoses duplicates (same
/// namespace+kind+sub seen more than once), and never writes. `path` may be a
/// file or a directory.
pub fn scan_path(path: &Path) -> Result<Value> {
    let meta = fs::symlink_metadata(path).map_err(|_| "tag scan path unavailable".to_string())?;
    let mut reports = Vec::new();
    let paths: Vec<PathBuf> = if meta.is_dir() {
        let mut found = Vec::new();
        walk_files(path, &mut found, 0)?;
        found.sort();
        found
    } else if meta.is_file() {
        vec![path.to_path_buf()]
    } else {
        return Err("tag scan path is not a file or directory".into());
    };
    let mut total_nodes = 0usize;
    let mut total_invalid = 0usize;
    let mut duplicates: Vec<Value> = Vec::new();
    for file in &paths {
        let report = scan_file(file)?;
        let nodes = report["nodes"].as_array().cloned().unwrap_or_default();
        let invalid = report["invalid"].as_array().cloned().unwrap_or_default();
        total_nodes += nodes.len();
        total_invalid += invalid.len();

        // Duplicate diagnosis is per file (REQ-C2-01): the same kind+sub declared
        // in BOTH namespaces inside one file is the migration duplicate. The same
        // tag repeated across separate files is normal practice, not a duplicate.
        let mut per_file: BTreeMap<String, Vec<Value>> = BTreeMap::new();
        for node in &nodes {
            let key = format!(
                "{}|{}",
                node["kind"].as_str().unwrap_or(""),
                node["sub"].as_str().unwrap_or(""),
            );
            per_file.entry(key).or_default().push(node.clone());
        }
        let file_dups: Vec<Value> = per_file
            .iter()
            .filter(|(_, group)| {
                let namespaces: Vec<&str> = group
                    .iter()
                    .map(|n| n["namespace"].as_str().unwrap_or(""))
                    .collect();
                namespaces.contains(&"MOAI") && namespaces.contains(&"MX")
            })
            .map(|(key, group)| {
                json!({
                    "file": file.to_string_lossy(),
                    "key": key,
                    "namespaces": group.iter().map(|n| n["namespace"].clone()).collect::<Vec<_>>(),
                    "lines": group.iter().map(|n| n["line"].clone()).collect::<Vec<_>>(),
                })
            })
            .collect();
        if !file_dups.is_empty() {
            duplicates.extend(file_dups);
        }
        reports.push(json!({
            "path": file.to_string_lossy(),
            "nodes": nodes,
            "invalid": invalid,
        }));
    }
    Ok(json!({
        "parser_version": PARSER_VERSION,
        "files": reports,
        "file_count": reports.len(),
        "node_count": total_nodes,
        "invalid_count": total_invalid,
        "duplicates": duplicates,
        "duplicate_diagnosis_only": true,
        "modified": 0,
        "writes": false,
    }))
}
