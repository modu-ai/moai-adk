//! Explicit SPEC link extraction — REQ ↔ code ↔ test (C5 #2181, REQ-C5-01).
//!
//! Reads three explicit sources from a git work tree and emits C4 ontology records
//! (`src/ontology.rs`). Every edge has `basis: "explicit"` and
//! `verification: "unreviewed"`; nothing is inferred (REQ-C5-01):
//!
//! | source | grammar | edge |
//! | --- | --- | --- |
//! | tag in a tracked code file | C2 grammar v1, kind `REQ` in either namespace (`@MOAI:REQ REQ-C5-01`, `@MX:REQ …`); args split on whitespace and commas | `file:<path>` → `references_requirement` → `req:<id>` |
//! | test declaration outside comments/literals | a Rust fn whose complete attribute block carries `#[test]` or `#[tokio::test]`, or a Python `def test_…`; the leaf name carries `req_<segs>_<digits>` | `test:<path>::<module-or-class-scope>::<fn>` → `verifies` → `req:<id>` (top-level names have no scope prefix) |
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

#[derive(Clone, Copy)]
struct CodeToken<'a> {
    text: &'a str,
    line: usize,
    column: usize,
}

/// A borrowed token stream; literals and comments never become declarations.
struct CodeTokens<'a> {
    text: &'a str,
    rust: bool,
    pos: usize,
    line: usize,
    column: usize,
}

impl<'a> CodeTokens<'a> {
    fn new(text: &'a str, rust: bool) -> Self {
        Self {
            text,
            rust,
            pos: 0,
            line: 1,
            column: 0,
        }
    }

    fn advance(&mut self, end: usize) {
        for b in self.text.as_bytes()[self.pos..end].iter().copied() {
            match b {
                b'\n' => {
                    self.line += 1;
                    self.column = 0;
                }
                b'\t' => self.column += 8 - self.column % 8,
                _ => self.column += 1,
            }
        }
        self.pos = end;
    }

    fn quoted_end(&self, start: usize, width: usize) -> usize {
        let b = self.text.as_bytes();
        let mut end = start + width;
        while end < b.len() {
            if b[end] == b'\\' {
                end = (end + 2).min(b.len());
            } else if b[end..].starts_with(&b[start..start + width]) {
                return end + width;
            } else {
                end += 1;
            }
        }
        b.len()
    }
}

impl<'a> Iterator for CodeTokens<'a> {
    type Item = CodeToken<'a>;

    fn next(&mut self) -> Option<Self::Item> {
        let b = self.text.as_bytes();
        while self.pos < b.len() {
            let start = self.pos;
            if b[start].is_ascii_whitespace() {
                self.advance(start + 1);
                continue;
            }
            if (!self.rust && b[start] == b'#') || (self.rust && b[start..].starts_with(b"//")) {
                let end = b[start..]
                    .iter()
                    .position(|v| *v == b'\n')
                    .map_or(b.len(), |n| start + n);
                self.advance(end);
                continue;
            }
            if self.rust && b[start..].starts_with(b"/*") {
                let mut end = start + 2;
                let mut depth = 1;
                while end < b.len() && depth > 0 {
                    if b[end..].starts_with(b"/*") {
                        depth += 1;
                        end += 2;
                    } else if b[end..].starts_with(b"*/") {
                        depth -= 1;
                        end += 2;
                    } else {
                        end += 1;
                    }
                }
                self.advance(end);
                continue;
            }
            // Byte/C prefixes do not turn raw contents into escaped ordinary strings.
            let literal_start = if self.rust
                && matches!(b[start], b'b' | b'c')
                && matches!(b.get(start + 1), Some(b'r' | b'"'))
            {
                start + 1
            } else {
                start
            };
            if self.rust && b[literal_start] == b'r' {
                let mut quote = literal_start + 1;
                while b.get(quote) == Some(&b'#') {
                    quote += 1;
                }
                if b.get(quote) == Some(&b'"') {
                    let hashes = quote - literal_start - 1;
                    let mut end = quote + 1;
                    while end < b.len() {
                        if b[end] == b'"'
                            && b.get(end + 1..end + 1 + hashes)
                                .is_some_and(|s| s.iter().all(|v| *v == b'#'))
                        {
                            end += 1 + hashes;
                            break;
                        }
                        end += 1;
                    }
                    self.advance(end);
                    continue;
                }
            }
            if b[literal_start] == b'"' || (!self.rust && b[start] == b'\'') {
                let width = if !self.rust && b[start..].starts_with(&[b[start]; 3]) {
                    3
                } else {
                    1
                };
                self.advance(self.quoted_end(literal_start, width));
                continue;
            }
            if self.rust && b[start] == b'\'' {
                let mut end = start + 1;
                if b.get(end) == Some(&b'\\') {
                    end += 2;
                    while end < b.len() && b[end] != b'\'' && b[end] != b'\n' {
                        end += 1;
                    }
                } else if let Some(c) = self.text[end..].chars().next() {
                    end += c.len_utf8();
                }
                if b.get(end) == Some(&b'\'') {
                    self.advance(end + 1);
                    continue;
                }
            }
            let name_start = if self.rust
                && b[start..].starts_with(b"r#")
                && self.text[start + 2..]
                    .chars()
                    .next()
                    .is_some_and(|c| c.is_alphabetic() || c == '_')
            {
                start + 2
            } else {
                start
            };
            let first = self.text[name_start..].chars().next()?;
            let end = if first.is_alphanumeric() || first == '_' {
                self.text[name_start..]
                    .char_indices()
                    .find(|(_, c)| !c.is_alphanumeric() && *c != '_')
                    .map_or(b.len(), |(n, _)| name_start + n)
            } else {
                start + first.len_utf8()
            };
            let token = CodeToken {
                text: &self.text[name_start..end],
                line: self.line,
                column: self.column,
            };
            self.advance(end);
            return Some(token);
        }
        None
    }
}

fn scoped_name(scopes: &[(usize, &str)], name: &str) -> String {
    let mut out =
        String::with_capacity(name.len() + scopes.iter().map(|(_, s)| s.len() + 2).sum::<usize>());
    for (_, scope) in scopes {
        out.push_str(scope);
        out.push_str("::");
    }
    out.push_str(name);
    out
}

fn skip_token_group<'a>(tokens: &mut impl Iterator<Item = CodeToken<'a>>) {
    let mut depth = 1usize;
    for token in tokens.by_ref() {
        match token.text {
            "(" | "[" | "{" => depth += 1,
            ")" | "]" | "}" => depth -= 1,
            _ => {}
        }
        if depth == 0 {
            break;
        }
    }
}

fn rust_tests(text: &str) -> Vec<(usize, String)> {
    let mut tokens = CodeTokens::new(text, true).peekable();
    let mut scopes = Vec::new();
    let mut depth = 0usize;
    let mut module = None;
    let mut is_test = false;
    let mut out = Vec::new();
    while let Some(token) = tokens.next() {
        match token.text {
            "#" if tokens.peek().is_some_and(|t| t.text == "[") => {
                tokens.next();
                let mut balance = 1;
                let mut prefix = [None; 6];
                for (n, t) in tokens.by_ref().enumerate() {
                    if n < prefix.len() {
                        prefix[n] = Some(t.text);
                    }
                    match t.text {
                        "[" => balance += 1,
                        "]" => balance -= 1,
                        _ => {}
                    }
                    if balance == 0 {
                        break;
                    }
                }
                is_test |= prefix[..2] == [Some("test"), Some("]")]
                    || (prefix[..4] == [Some("tokio"), Some(":"), Some(":"), Some("test")]
                        && matches!(prefix[4], Some("(" | "]")));
            }
            "macro_rules" if tokens.peek().is_some_and(|t| t.text == "!") => {
                tokens.next();
                tokens.next();
                if tokens
                    .next()
                    .is_some_and(|t| matches!(t.text, "(" | "[" | "{"))
                {
                    skip_token_group(&mut tokens);
                }
                is_test = false;
            }
            "!" if tokens
                .peek()
                .is_some_and(|t| matches!(t.text, "(" | "[" | "{")) =>
            {
                tokens.next();
                skip_token_group(&mut tokens);
                is_test = false;
            }
            "mod" => {
                is_test = false;
                module = tokens.next().map(|t| t.text);
            }
            "fn" => {
                if let Some(name) = tokens.next() {
                    if is_test {
                        out.push((token.line, scoped_name(&scopes, name.text)));
                    }
                }
                is_test = false;
            }
            "{" => {
                depth += 1;
                if let Some(name) = module.take() {
                    scopes.push((depth, name));
                }
                is_test = false;
            }
            "}" => {
                depth = depth.saturating_sub(1);
                while scopes.last().is_some_and(|(d, _)| *d > depth) {
                    scopes.pop();
                }
                is_test = false;
            }
            ";" | "struct" | "enum" | "impl" | "trait" | "const" | "static" | "use" | "type" => {
                module = None;
                is_test = false;
            }
            _ => {}
        }
    }
    out
}

fn python_tests(text: &str) -> Vec<(usize, String)> {
    let mut tokens = CodeTokens::new(text, false);
    let mut scopes = Vec::new();
    let mut last_line = 0;
    let mut brackets = 0usize;
    let mut async_line = None;
    let mut async_indent = 0;
    let mut out = Vec::new();
    while let Some(token) = tokens.next() {
        let first = token.line != last_line && brackets == 0;
        if first {
            while scopes
                .last()
                .is_some_and(|(indent, _)| *indent >= token.column)
            {
                scopes.pop();
            }
            async_line = None;
        }
        last_line = token.line;
        if first && token.text == "async" {
            async_line = Some(token.line);
            async_indent = token.column;
        } else if (first || async_line == Some(token.line)) && matches!(token.text, "class" | "def")
        {
            if let Some(name) = tokens.next() {
                if token.text == "def" && name.text.starts_with("test") {
                    out.push((token.line, scoped_name(&scopes, name.text)));
                }
                let indent = if async_line == Some(token.line) {
                    async_indent
                } else {
                    token.column
                };
                scopes.push((indent, name.text));
            }
            async_line = None;
        }
        match token.text {
            "(" | "[" | "{" => brackets += 1,
            ")" | "]" | "}" => brackets = brackets.saturating_sub(1),
            _ => {}
        }
    }
    out
}

/// Explicit declarations only; names include their enclosing module/class scope.
fn test_functions(path: &str, text: &str) -> Vec<(usize, String)> {
    if path.ends_with(".py") {
        python_tests(text)
    } else if path.ends_with(".rs") {
        rust_tests(text)
    } else {
        Vec::new()
    }
}

fn git(root: &Path, args: &[&str]) -> Result<String> {
    let mut cmd = Command::new("git");
    cmd.arg("-C").arg(root).args(args);
    crate::identity::scrub_git_env(&mut cmd);
    let out = cmd.output().map_err(|_| "git unavailable".to_string())?;
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
pub(crate) struct Builder {
    pub(crate) nodes: BTreeMap<String, OntologyNode>,
    pub(crate) edges: BTreeMap<(String, String, String), BTreeSet<String>>,
    pub(crate) missing: Vec<Value>,
}

impl Builder {
    pub(crate) fn node(&mut self, id: &str, kind: &str, revision: &str, props: &[(&str, &str)]) {
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

    pub(crate) fn requirement(&mut self, id: &str) {
        self.node(
            &format!("req:{id}"),
            "requirement",
            "unresolved",
            &[("revision_source", "not read: SPEC documents are outside C5")],
        );
    }

    pub(crate) fn edge(&mut self, kind: &str, source: &str, target: &str, evidence: String) {
        self.edges
            .entry((kind.to_string(), source.to_string(), target.to_string()))
            .or_default()
            .insert(evidence);
    }

    pub(crate) fn miss(&mut self, kind: &str, detail: Value) {
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
        if !is_code_path(path) || unrepresentable(&mut b, path) {
            continue;
        }
        let full = root.join(path);
        if !fs::symlink_metadata(&full).is_ok_and(|m| m.is_file()) {
            b.miss(
                "unreadable_file",
                json!({"path": path, "reason": "not a regular file"}),
            );
            continue;
        }
        let Ok(bytes) = fs::read(&full) else {
            b.miss(
                "unreadable_file",
                json!({"path": path, "reason": "read failed"}),
            );
            continue;
        };
        file_links(&mut b, path, bytes);
    }
    commit_links(&mut b, root, None)?;
    finish(b, snapshot_id, observed_at)
}

/// Whether `extract` reads this tracked path (code extensions only).
pub(crate) fn is_code_path(path: &str) -> bool {
    CODE_EXTENSIONS.contains(&path.rsplit('.').next().unwrap_or(""))
}

/// Records a path that cannot appear in an evidence file-uri; true when skipped.
fn unrepresentable(b: &mut Builder, path: &str) -> bool {
    if path.contains('#') || path.contains('@') {
        b.miss(
            "path_unrepresentable",
            json!({"path": path, "reason": "# or @ cannot appear in an evidence file-uri"}),
        );
        return true;
    }
    false
}

/// The per-file part of `extract`: REQ tags and REQ-named test declarations of one
/// code file, from bytes the caller already read (and, in the C6b index build,
/// verified against the manifest digest). Callers filter with `is_code_path`.
pub(crate) fn file_links(b: &mut Builder, path: &str, bytes: Vec<u8>) {
    if unrepresentable(b, path) {
        return;
    }
    let bytes_len = bytes.len() as u64;
    let file_digest = digest(&bytes);
    let Ok(text) = String::from_utf8(bytes) else {
        b.miss(
            "unreadable_file",
            json!({"path": path, "reason": "not UTF-8"}),
        );
        return;
    };
    let evidence = |line: usize| format!("file:{path}#L{line}@{file_digest}");
    let file_id = format!("file:{path}");

    let tags = if bytes_len > atlas_tag::FILE_LIMIT {
        Err("tag file exceeds the size limit")
    } else {
        Ok(atlas_tag::scan_text(&text))
    };
    match tags {
        Err(e) => b.miss("unreadable_file", json!({"path": path, "reason": e})),
        Ok(file) => {
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
        let leaf = name.rsplit("::").next().unwrap_or(&name);
        let ids = req_ids_in_test_name(leaf);
        let test_id = format!("test:{path}::{name}");
        for id in ids {
            b.node(
                &test_id,
                "test",
                &file_digest,
                &[("path", path), ("name", leaf)],
            );
            b.requirement(&id);
            b.edge("verifies", &test_id, &format!("req:{id}"), evidence(line));
        }
    }
}

/// The history part of `extract`: commit → issue mentions reachable from `rev`
/// (HEAD when `None`).
pub(crate) fn commit_links(b: &mut Builder, root: &Path, rev: Option<&str>) -> Result<()> {
    // Read raw bodies: a footer or a separate coauthor paragraph must not hide
    // explicit protocol lines. %x1e separates commits; %x1f separates SHA and body.
    let mut args = vec!["log", "--format=%H%x1f%B%x1e"];
    args.extend(rev);
    let log = git(root, &args)?;
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
    Ok(())
}

/// The global part of `extract`: triangle gaps, record serialization and the C4
/// schema check over everything the builder collected.
pub(crate) fn finish(b: Builder, snapshot_id: &str, observed_at: &str) -> Result<Value> {
    finish_with(
        b,
        EXTRACTOR_NAME,
        EXTRACTOR_VERSION,
        snapshot_id,
        observed_at,
    )
}

/// `finish` under another extractor identity (the C6b map view records tags, not links).
pub(crate) fn finish_with(
    mut b: Builder,
    extractor_name: &str,
    extractor_version: &str,
    snapshot_id: &str,
    observed_at: &str,
) -> Result<Value> {
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
        name: extractor_name.into(),
        version: extractor_version.into(),
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
        "extractor": {"name": extractor_name, "version": extractor_version},
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
