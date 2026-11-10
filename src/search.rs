//! Non-vector-first search (rc.2 C7 #2183): the retrieval pipeline of the Core Brain design
//! (`page_136c8203…` L3716–3724), without the typed-edge expansion step, which moved to #2408.
//!
//! 1. validate the project (repository top) and read the index freshness (`index_query`);
//! 2. resolve exact paths before anything else (REQ-C7-01);
//! 3. search literal strings line by line over the published generation's files, or over the
//!    tracked files when no generation is published (lower coverage, design L3740);
//! 4. typed-edge expansion is reported `unavailable` (no producer yet, design L4300);
//! 5. exact results rank ahead of lexical ones; names and paths are never inferred;
//! 6. results are deduplicated by (path, span);
//! 7. two caps cut the list — files and spans per file — and every cut is an `omitted` entry;
//! 8. before returning, each cited span is re-read and compared with the digest taken while
//!    collecting; a changed span is marked `stale`;
//! 9. the answer carries coverage, freshness and the omissions.
//!
//! Read-only: no file is written, no network is used.
use serde::Deserialize;
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::{
    collections::BTreeSet,
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

/// Step 8: re-read every cited span; `fresh` when its bytes still hash to the collected
/// digest, `stale` otherwise (changed or gone).
pub fn recheck(root: &Path, hits: &[Hit]) -> Vec<&'static str> {
    hits.iter()
        .map(|h| {
            let Ok(bytes) = fs::read(root.join(&h.path)) else {
                return "stale";
            };
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
    let (hits, mut omitted) = collect(&root, &files, q);
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
            "graph": "unavailable: no typed-edge producer yet (#2408)",
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
        let mut listed = vec!["ok.txt".to_string()];
        #[cfg(unix)]
        {
            std::os::unix::fs::symlink(
                outside.path().join("secret.txt"),
                dir.path().join("link.txt"),
            )
            .unwrap();
            listed.push("link.txt".into());
        }
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
        if cfg!(unix) {
            assert!(
                reasons.contains(&("link.txt".into(), "exact path is not a regular file".into()))
            );
        }
        assert!(!serde_json::to_string(&omitted)
            .unwrap()
            .contains(&sha256_hex(b"s1\ns2\n")));
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
