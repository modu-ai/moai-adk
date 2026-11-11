//! Code-reference producer (C6c #2442): file → file `imports` and `calls` records.
//!
//! Design basis, read 2026-11-11T01:48:29Z (sources, checksums and the adopt/reject table
//! are in Issue #2442 comment 6784640144):
//! - Rust Reference, Use declarations: `use a::b::{c, d as e, f::*}`, `self` in a group.
//! - Rust Reference, Paths: a path starts with `crate`, `self`, `super` or an identifier;
//!   expression paths may carry turbofish generic arguments.
//! - Rust Reference, Modules: a body-less `mod x;` loads `x.rs` or `x/mod.rs` (both is an
//!   error); ancestor module path components are directories; `#[path]` overrides it.
//! - Rust Reference, Comments and Tokens: comments (doc comments included) and string,
//!   raw string and character literals carry no code reference.
//! - Cargo Book, Cargo Targets: the library is `src/lib.rs`, named after the package with
//!   `-` replaced by `_`; `src/main.rs` and `src/bin/` use the library by that name.
//! - Core Brain detailed design (local copy, whole-file sha256 `49e445f4…`): L3008 a
//!   syntactic call is a candidate until the callee binding is checked; L4296 unsupported
//!   languages are reported as excluded from coverage.
//!
//! Two stages keep a full rebuild and an incremental build identical (INCREMENTAL
//! EQUIVALENCE): `scan` is a pure function of one file's bytes and its result is stored
//! per origin as a `code_scan` node; `resolve` depends on the whole file set (which
//! module file exists) and runs on every build over the stored scans, so adding or
//! removing a target file changes the edges without re-reading unchanged sources.
//!
//! What is recorded: `use` paths as `imports`, and multi-segment expression paths that
//! are called (`a::b::f(…)`, `a::b::m!(…)`) as `calls`, each resolved through the `mod`
//! tree to the file that defines the deepest module the path names. Paths used only as
//! types or constants outside `use` are not recorded; macro definitions are not expanded.

use crate::ontology::{self, Extractor, OntologyEdge};
use serde_json::{json, Value};
use std::collections::{BTreeMap, BTreeSet, VecDeque};

/// Extractor identity of the code-reference edges.
pub const EXTRACTOR_NAME: &str = "code-refs";
/// Bump on any change to what `scan` or `resolve` records.
pub const EXTRACTOR_VERSION: &str = "5";
/// Node kind holding one Rust file's stored scan.
pub(crate) const SCAN_NODE_KIND: &str = "code_scan";
/// Node kind holding one Cargo manifest's library crate name.
pub(crate) const CRATE_NODE_KIND: &str = "crate_manifest";
/// `uncertainty` of a `calls` edge: the module file is resolved, the callee is not bound.
pub const CALL_UNCERTAINTY: &str = "callee-binding-not-checked";
/// `uncertainty` of an edge whose library crate name was inferred, not read.
pub const CRATE_NAME_UNCERTAINTY: &str = "crate-name-not-read";
/// `uncertainty` of a `calls` edge whose library crate name was inferred: both apply.
pub const INFERRED_CALL_UNCERTAINTY: &str = "crate-name-not-read; callee-binding-not-checked";
/// Deepest `use` group nesting followed; deeper groups are skipped, so one malformed file
/// cannot exhaust the index build's stack.
pub const MAX_USE_DEPTH: usize = 32;

const EXTERNAL_ROOTS: [&str; 5] = ["std", "core", "alloc", "proc_macro", "test"];

/// One `mod` item: `scope` is the chain of enclosing inline modules in the same file.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ModDecl {
    pub scope: Vec<String>,
    pub name: String,
    pub line: u32,
    pub inline: bool,
    pub path_attr: bool,
}

/// One unresolved reference. `alias` is the name a `use` leaf binds (None for globs and
/// `as _`); always None for calls.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Ref {
    pub kind: String,
    pub scope: Vec<String>,
    pub path: Vec<String>,
    pub alias: Option<String>,
    pub line: u32,
}

/// Everything `resolve` needs from one file.
#[derive(Debug, Default, Clone, PartialEq, Eq)]
pub struct Scan {
    pub mods: Vec<ModDecl>,
    pub refs: Vec<Ref>,
    /// Module-level item declarations; associated and local items are not resolved.
    pub definitions: Vec<Definition>,
    /// One explicit coverage gap per excluded documentation block.
    pub excluded_doc_blocks: Vec<DocExclusion>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DocExclusion {
    pub line: u32,
    pub reason: String,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Definition {
    pub scope: Vec<String>,
    pub name: String,
    pub line: u32,
}

#[derive(Debug, Clone, PartialEq)]
enum T {
    Ident(String),
    Doc(String, bool, bool),
    Lit,
    ColCol,
    P(char),
}

struct Tok {
    t: T,
    line: u32,
}

fn ident_start(c: char) -> bool {
    c.is_alphabetic() || c == '_'
}

fn ident_char(c: char) -> bool {
    c.is_alphanumeric() || c == '_'
}

/// Skips a quoted string body starting after the opening `"`; returns the index after
/// the closing quote.
fn skip_str(c: &[char], mut j: usize, line: &mut u32) -> usize {
    while j < c.len() {
        match c[j] {
            '\\' => {
                if c.get(j + 1) == Some(&'\n') {
                    *line += 1;
                }
                j += 2;
            }
            '"' => return j + 1,
            ch => {
                if ch == '\n' {
                    *line += 1;
                }
                j += 1;
            }
        }
    }
    c.len()
}

/// Raw string at `p` (pointing at `#…#"` or `"`): returns the index after it, or None
/// when `p` does not start a raw string.
fn skip_raw(c: &[char], p: usize, line: &mut u32) -> Option<usize> {
    let mut j = p;
    while c.get(j) == Some(&'#') {
        j += 1;
    }
    if c.get(j) != Some(&'"') {
        return None;
    }
    let hashes = j - p;
    j += 1;
    while j < c.len() {
        if c[j] == '"' && (1..=hashes).all(|k| c.get(j + k) == Some(&'#')) {
            return Some(j + 1 + hashes);
        }
        if c[j] == '\n' {
            *line += 1;
        }
        j += 1;
    }
    Some(c.len())
}

/// Character literal or lifetime at `i` (pointing at `'`): returns the index after it.
fn skip_quote(c: &[char], i: usize) -> usize {
    if c.get(i + 1) == Some(&'\\') {
        let mut j = i + 3;
        while j < c.len() && c[j] != '\'' && c[j] != '\n' {
            j += 1;
        }
        return (j + 1).min(c.len());
    }
    if c.get(i + 2) == Some(&'\'') {
        return i + 3;
    }
    let mut j = i + 1;
    while j < c.len() && ident_char(c[j]) {
        j += 1;
    }
    j.max(i + 1)
}

fn lex(src: &str) -> Vec<Tok> {
    let c: Vec<char> = src.chars().collect();
    let n = c.len();
    let (mut i, mut line, mut out) = (0usize, 1u32, Vec::new());
    while i < n {
        let ch = c[i];
        if ch == '\n' {
            line += 1;
            i += 1;
        } else if ch.is_whitespace() {
            i += 1;
        } else if ch == '/' && c.get(i + 1) == Some(&'/') {
            let start = i;
            while i < n && c[i] != '\n' {
                i += 1;
            }
            if c.get(start + 2) == Some(&'!')
                || (c.get(start + 2) == Some(&'/') && c.get(start + 3) != Some(&'/'))
            {
                out.push(Tok {
                    t: T::Doc(
                        c[start + 3..i].iter().collect(),
                        false,
                        c[..start]
                            .iter()
                            .rev()
                            .take_while(|ch| **ch != '\n')
                            .all(|ch| ch.is_whitespace()),
                    ),
                    line,
                });
            }
        } else if ch == '/' && c.get(i + 1) == Some(&'*') {
            let start = i;
            let first_line = line;
            let mut depth = 0usize;
            while i < n {
                if c[i] == '/' && c.get(i + 1) == Some(&'*') {
                    depth += 1;
                    i += 2;
                } else if c[i] == '*' && c.get(i + 1) == Some(&'/') {
                    depth -= 1;
                    i += 2;
                    if depth == 0 {
                        break;
                    }
                } else {
                    if c[i] == '\n' {
                        line += 1;
                    }
                    i += 1;
                }
            }
            if (c.get(start + 2) == Some(&'!')
                || (c.get(start + 2) == Some(&'*') && !matches!(c.get(start + 3), Some('*' | '/'))))
                && i >= start + 5
                && c.get(i - 2..i) == Some(&['*', '/'])
            {
                out.push(Tok {
                    t: T::Doc(c[start + 3..i - 2].iter().collect(), true, false),
                    line: first_line,
                });
            }
        } else if ch == '"' {
            let at = line;
            i = skip_str(&c, i + 1, &mut line);
            out.push(Tok {
                t: T::Lit,
                line: at,
            });
        } else if ch == '\'' {
            i = skip_quote(&c, i);
        } else if ident_start(ch) {
            let start = i;
            while i < n && ident_char(c[i]) {
                i += 1;
            }
            let word: String = c[start..i].iter().collect();
            let at = line;
            let next = c.get(i).copied();
            match (word.as_str(), next) {
                ("b" | "c", Some('"')) => {
                    i = skip_str(&c, i + 1, &mut line);
                    out.push(Tok {
                        t: T::Lit,
                        line: at,
                    });
                    continue;
                }
                ("b", Some('\'')) => {
                    i = skip_quote(&c, i);
                    out.push(Tok {
                        t: T::Lit,
                        line: at,
                    });
                    continue;
                }
                ("r" | "br" | "cr", Some('"' | '#')) => {
                    if let Some(end) = skip_raw(&c, i, &mut line) {
                        i = end;
                        out.push(Tok {
                            t: T::Lit,
                            line: at,
                        });
                        continue;
                    }
                    if word == "r" && c.get(i + 1).is_some_and(|&x| ident_start(x)) {
                        let s = i + 1;
                        i = s;
                        while i < n && ident_char(c[i]) {
                            i += 1;
                        }
                        let raw: String = c[s..i].iter().collect();
                        out.push(Tok {
                            t: T::Ident(raw),
                            line: at,
                        });
                        continue;
                    }
                }
                _ => {}
            }
            out.push(Tok {
                t: T::Ident(word),
                line: at,
            });
        } else if ch.is_ascii_digit() {
            while i < n && ident_char(c[i]) {
                i += 1;
            }
        } else if ch == ':' && c.get(i + 1) == Some(&':') {
            out.push(Tok { t: T::ColCol, line });
            i += 2;
        } else {
            out.push(Tok { t: T::P(ch), line });
            i += 1;
        }
    }
    out
}

fn at(toks: &[Tok], i: usize) -> Option<&T> {
    toks.get(i).map(|t| &t.t)
}

fn ident_at(toks: &[Tok], i: usize) -> Option<&str> {
    match at(toks, i) {
        Some(T::Ident(w)) => Some(w),
        _ => None,
    }
}

/// Whether a `#[path = …]` attribute precedes the `mod` keyword at `i`.
fn has_path_attr(toks: &[Tok], i: usize) -> bool {
    let mut k = i as isize - 1;
    while k >= 0 {
        match &toks[k as usize].t {
            T::Doc(..) => k -= 1,
            T::Ident(w) if w == "pub" => k -= 1,
            T::P(close @ (')' | ']')) => {
                let open = if *close == ')' { '(' } else { '[' };
                let mut depth = 0i32;
                let mut m = k;
                while m >= 0 {
                    match &toks[m as usize].t {
                        T::P(x) if x == close => depth += 1,
                        T::P(x) if *x == open => {
                            depth -= 1;
                            if depth == 0 {
                                break;
                            }
                        }
                        _ => {}
                    }
                    m -= 1;
                }
                if m < 0 {
                    return false;
                }
                if *close == ']' {
                    if m == 0 || at(toks, m as usize - 1) != Some(&T::P('#')) {
                        return false;
                    }
                    if ident_at(toks, m as usize + 1) == Some("path") {
                        return true;
                    }
                    k = m - 2;
                } else {
                    k = m - 1;
                }
            }
            _ => return false,
        }
    }
    false
}

/// Index after the generic argument list opening at `j` (`<`).
fn skip_angle(toks: &[Tok], mut j: usize) -> usize {
    let mut depth = 0i32;
    while j < toks.len() {
        match &toks[j].t {
            T::P('<') => depth += 1,
            T::P('>') if !(j > 0 && toks[j - 1].t == T::P('-')) => {
                depth -= 1;
                if depth == 0 {
                    return j + 1;
                }
            }
            T::P(';' | '{' | '}') => return j,
            _ => {}
        }
        j += 1;
    }
    j
}

type Leaf = (Vec<String>, Option<String>, u32);

/// Index after the brace group opening at `j`, without recursion; stops at `;`.
fn skip_group(toks: &[Tok], mut j: usize) -> usize {
    let mut depth = 0usize;
    while j < toks.len() {
        match toks[j].t {
            T::P('{') => depth += 1,
            T::P('}') => {
                depth -= 1;
                if depth == 0 {
                    return j + 1;
                }
            }
            T::P(';') => return j,
            _ => {}
        }
        j += 1;
    }
    j
}

fn use_tree(
    toks: &[Tok],
    mut j: usize,
    mut path: Vec<String>,
    leaves: &mut Vec<Leaf>,
    depth: usize,
) -> usize {
    if at(toks, j) == Some(&T::ColCol) {
        j += 1;
    }
    loop {
        match at(toks, j) {
            Some(T::Ident(w)) => {
                let w = w.clone();
                let line = toks[j].line;
                path.push(w.clone());
                j += 1;
                if at(toks, j) == Some(&T::ColCol) {
                    j += 1;
                    continue;
                }
                let mut alias = Some(w.clone());
                if ident_at(toks, j) == Some("as") {
                    if let Some(b) = ident_at(toks, j + 1) {
                        alias = (b != "_").then(|| b.to_string());
                        j += 2;
                    }
                }
                if w == "self" {
                    path.pop();
                    if alias.as_deref() == Some("self") {
                        alias = path.last().cloned();
                    }
                }
                leaves.push((path, alias, line));
                return j;
            }
            Some(T::P('*')) => {
                leaves.push((path, None, toks[j].line));
                return j + 1;
            }
            Some(T::P('{')) => {
                if depth >= MAX_USE_DEPTH {
                    return skip_group(toks, j);
                }
                j += 1;
                loop {
                    match at(toks, j) {
                        Some(T::P('}')) => return j + 1,
                        Some(T::P(',')) => j += 1,
                        None => return j,
                        _ => {
                            let before = j;
                            j = use_tree(toks, j, path.clone(), leaves, depth + 1);
                            if j == before {
                                j += 1;
                            }
                        }
                    }
                }
            }
            _ => return j,
        }
    }
}

/// Macro arguments and attribute payloads are token trees, not declarations. Mask
/// them for declaration/doc collection while preserving the existing lexical call
/// scan inside macro arguments. Unbalanced trees remain opaque through EOF.
fn opaque_declarations(toks: &[Tok]) -> Vec<bool> {
    let mut opaque = vec![false; toks.len()];
    let mut i = 0;
    while i < toks.len() {
        let opening = if at(toks, i) == Some(&T::P('#')) {
            let j = i + 1 + usize::from(at(toks, i + 1) == Some(&T::P('!')));
            (at(toks, j) == Some(&T::P('['))).then_some(j)
        } else if ident_at(toks, i).is_some() && at(toks, i + 1) == Some(&T::P('!')) {
            let j = i + 2 + usize::from(ident_at(toks, i) == Some("macro_rules"));
            matches!(at(toks, j), Some(T::P('(' | '[' | '{'))).then_some(j)
        } else {
            None
        };
        let Some(opening) = opening else {
            i += 1;
            continue;
        };
        let mut stack = Vec::new();
        let mut end = toks.len();
        for (j, tok) in toks.iter().enumerate().skip(opening) {
            match tok.t {
                T::P('(') => stack.push(')'),
                T::P('[') => stack.push(']'),
                T::P('{') => stack.push('}'),
                T::P(close @ (')' | ']' | '}')) => {
                    if stack.pop() != Some(close) {
                        break;
                    }
                    if stack.is_empty() {
                        end = j + 1;
                        break;
                    }
                }
                _ => {}
            }
        }
        opaque[i..end].fill(true);
        i = end;
    }
    opaque
}

/// Scans one Rust source text. Pure: the same text always yields the same scan.
pub fn scan(src: &str) -> Scan {
    let toks = lex(src);
    let opaque = opaque_declarations(&toks);
    let mut s = Scan::default();
    let mut scope: Vec<String> = Vec::new();
    let mut open: Vec<usize> = Vec::new();
    let mut depth = 0usize;
    let mut i = 0usize;
    while i < toks.len() {
        match &toks[i].t {
            T::Doc(text, block, pure_line) if !opaque[i] => {
                let mut lines = Vec::new();
                let mut end = i + 1;
                if *block {
                    for (offset, text) in text.lines().enumerate() {
                        lines.push((text, toks[i].line + offset as u32));
                    }
                } else {
                    lines.push((text.as_str(), toks[i].line));
                    while *pure_line && end < toks.len() && !opaque[end] {
                        let T::Doc(next, false, true) = &toks[end].t else {
                            break;
                        };
                        if toks[end].line != toks[end - 1].line + 1 {
                            break;
                        }
                        lines.push((next.as_str(), toks[end].line));
                        end += 1;
                    }
                }
                collect_doc_block(&lines, *block, &scope, &mut s);
                i = end;
            }
            T::P('{') => {
                depth += 1;
                i += 1;
            }
            T::P('}') => {
                depth = depth.saturating_sub(1);
                if open.last() == Some(&depth) {
                    open.pop();
                    scope.pop();
                }
                i += 1;
            }
            T::Ident(w) if w == "mod" && !opaque[i] => {
                let (Some(name), Some(T::P(end @ (';' | '{')))) =
                    (ident_at(&toks, i + 1), at(&toks, i + 2))
                else {
                    i += 1;
                    continue;
                };
                let inline = *end == '{';
                s.mods.push(ModDecl {
                    scope: scope.clone(),
                    name: name.to_string(),
                    line: toks[i].line,
                    inline,
                    path_attr: has_path_attr(&toks, i),
                });
                if inline {
                    scope.push(name.to_string());
                    open.push(depth);
                    depth += 1;
                }
                i += 3;
            }
            T::Ident(w) if w == "use" && !opaque[i] => {
                let mut leaves = Vec::new();
                let mut j = use_tree(&toks, i + 1, Vec::new(), &mut leaves, 0);
                while j < toks.len() && toks[j].t != T::P(';') {
                    j += 1;
                }
                for (path, alias, line) in leaves {
                    if !path.is_empty() {
                        s.refs.push(Ref {
                            kind: "imports".into(),
                            scope: scope.clone(),
                            path,
                            alias,
                            line,
                        });
                    }
                }
                i = j + 1;
            }
            T::Ident(w)
                if !opaque[i]
                    && depth == open.len()
                    && matches!(
                        w.as_str(),
                        "fn" | "struct" | "enum" | "trait" | "type" | "static" | "union"
                    ) =>
            {
                if let Some(name) = ident_at(&toks, i + 1) {
                    s.definitions.push(Definition {
                        scope: scope.clone(),
                        name: name.into(),
                        line: toks[i].line,
                    });
                }
                i += 1;
            }
            T::Ident(_) => {
                if i > 0 && matches!(toks[i - 1].t, T::ColCol | T::P('.')) {
                    i += 1;
                    continue;
                }
                let line = toks[i].line;
                let mut path = Vec::new();
                let mut j = i;
                while let Some(w) = ident_at(&toks, j) {
                    path.push(w.to_string());
                    j += 1;
                    if at(&toks, j) != Some(&T::ColCol) {
                        break;
                    }
                    if at(&toks, j + 1) == Some(&T::P('<')) {
                        j = skip_angle(&toks, j + 1);
                        if at(&toks, j) == Some(&T::ColCol) && ident_at(&toks, j + 1).is_some() {
                            j += 1;
                            continue;
                        }
                        break;
                    }
                    if ident_at(&toks, j + 1).is_none() {
                        break;
                    }
                    j += 1;
                }
                let called = match at(&toks, j) {
                    Some(T::P('(')) => true,
                    Some(T::P('!')) => matches!(at(&toks, j + 1), Some(T::P('(' | '[' | '{'))),
                    _ => false,
                };
                if path.len() >= 2 && called {
                    s.refs.push(Ref {
                        kind: "calls".into(),
                        scope: scope.clone(),
                        path,
                        alias: None,
                        line,
                    });
                }
                i = j.max(i + 1);
            }
            _ => i += 1,
        }
    }
    s
}

fn spaces(text: &str) -> usize {
    text.bytes().take_while(|c| *c == b' ').count()
}

/// Strip only comment decoration, keeping the indentation that Markdown sees.
fn doc_text(text: &str, block: bool) -> &str {
    let text = text.strip_suffix('\r').unwrap_or(text);
    let text = text.strip_prefix(' ').unwrap_or(text);
    let indent = spaces(text);
    if block && indent <= 3 {
        let rest = &text[indent..];
        if let Some(rest) = rest.strip_prefix('*') {
            if rest.is_empty() || rest.starts_with(' ') {
                return rest.strip_prefix(' ').unwrap_or(rest);
            }
        }
    }
    text
}

/// This is a block-level exclusion predicate, not a Markdown container parser.
fn unsupported_doc_reason(text: &str) -> Option<&'static str> {
    if text.contains('\t') {
        return Some("tab");
    }
    let indent = spaces(text);
    if indent >= 4 {
        return Some("indent");
    }
    let body = &text[indent..];
    if body.starts_with('>') {
        return Some("blockquote");
    }
    let digits = body.bytes().take_while(u8::is_ascii_digit).count();
    let marker_width = if matches!(body.as_bytes().first(), Some(b'-' | b'+' | b'*')) {
        1
    } else if (1..=9).contains(&digits) && matches!(body.as_bytes().get(digits), Some(b'.' | b')'))
    {
        digits + 1
    } else {
        0
    };
    if marker_width > 0
        && (body.len() == marker_width || body.as_bytes()[marker_width].is_ascii_whitespace())
    {
        return Some("list");
    }
    if body.starts_with("```") || body.starts_with("~~~") {
        return Some("fence");
    }
    // Conservatively include tag-like starts, comments, declarations and processing
    // instructions. Autolinks such as <https://example.test> are not tag starts.
    if let Some(html) = body.strip_prefix('<') {
        if html.starts_with(['!', '?']) {
            return Some("html");
        }
        let tag = html.strip_prefix('/').unwrap_or(html);
        if tag.as_bytes().first().is_some_and(u8::is_ascii_alphabetic) {
            let width = tag
                .bytes()
                .take_while(|c| c.is_ascii_alphanumeric() || *c == b'-')
                .count();
            if tag.len() == width || matches!(tag.as_bytes()[width], b' ' | b'/' | b'>') {
                return Some("html");
            }
        }
    }
    None
}

/// Inspect the entire physical doc block before emitting any candidate. Ordinary
/// attributes, source/blank lines and adjacent block comments delimit new blocks.
fn collect_doc_block(lines: &[(&str, u32)], block: bool, scope: &[String], scan: &mut Scan) {
    let reasons: BTreeSet<&str> = lines
        .iter()
        .enumerate()
        .filter_map(|(offset, (text, _))| {
            unsupported_doc_reason(doc_text(text, block && offset > 0))
        })
        .collect();
    if !reasons.is_empty() {
        scan.excluded_doc_blocks.push(DocExclusion {
            line: lines.first().unwrap().1,
            reason: reasons.into_iter().collect::<Vec<_>>().join(","),
        });
        return;
    }
    for (offset, (text, line)) in lines.iter().enumerate() {
        for path in plain_doc_paths(doc_text(text, block && offset > 0)) {
            scan.refs.push(Ref {
                kind: "documents".into(),
                scope: scope.to_vec(),
                path,
                alias: None,
                line: *line,
            });
        }
    }
}

/// Qualified backtick paths in an already accepted plain paragraph block only.
fn plain_doc_paths(text: &str) -> Vec<Vec<String>> {
    let mut paths = Vec::new();
    let mut rest = text;
    while let Some(start) = rest.find('`') {
        rest = &rest[start..];
        let count = rest.bytes().take_while(|c| *c == b'`').count();
        let delimiter = "`".repeat(count);
        rest = &rest[count..];
        let Some(end) = rest.find(&delimiter) else {
            break;
        };
        let item = rest[..end]
            .trim()
            .strip_suffix("()")
            .unwrap_or(rest[..end].trim());
        let segments: Vec<&str> = item.split("::").collect();
        if segments.len() >= 2
            && segments.iter().all(|s| {
                let mut chars = s.chars();
                chars.next().is_some_and(ident_start) && chars.all(ident_char)
            })
        {
            paths.push(segments.into_iter().map(str::to_string).collect());
        }
        rest = &rest[end + count..];
    }
    paths
}

/// Canonical text of a scan, stored as a node property. Identifiers never hold a tab,
/// a newline or `::`, so the fields need no escaping.
pub fn encode(s: &Scan) -> String {
    let mut out = String::new();
    for m in &s.mods {
        out.push_str(&format!(
            "m\t{}\t{}\t{}\t{}\t{}\n",
            m.scope.join("::"),
            m.name,
            m.line,
            u8::from(m.inline),
            u8::from(m.path_attr)
        ));
    }
    for r in &s.refs {
        out.push_str(&format!(
            "r\t{}\t{}\t{}\t{}\t{}\n",
            r.kind,
            r.scope.join("::"),
            r.path.join("::"),
            r.alias.as_deref().unwrap_or(""),
            r.line
        ));
    }
    for d in &s.definitions {
        out.push_str(&format!(
            "d\t{}\t{}\t{}\n",
            d.scope.join("::"),
            d.name,
            d.line
        ));
    }
    for excluded in &s.excluded_doc_blocks {
        out.push_str(&format!("x\t{}\t{}\n", excluded.line, excluded.reason));
    }
    out
}

fn split_path(s: &str) -> Vec<String> {
    if s.is_empty() {
        Vec::new()
    } else {
        s.split("::").map(str::to_string).collect()
    }
}

/// Inverse of `encode`; None when the stored text is not a scan.
pub fn decode(text: &str) -> Option<Scan> {
    let mut s = Scan::default();
    for line in text.lines() {
        let f: Vec<&str> = line.split('\t').collect();
        match f.as_slice() {
            ["m", scope, name, l, inline, attr] => s.mods.push(ModDecl {
                scope: split_path(scope),
                name: name.to_string(),
                line: l.parse().ok()?,
                inline: *inline == "1",
                path_attr: *attr == "1",
            }),
            ["r", kind, scope, path, alias, l]
                if matches!(*kind, "imports" | "calls" | "documents") =>
            {
                if *kind == "documents" && (path.is_empty() || split_path(path).len() < 2) {
                    return None;
                }
                s.refs.push(Ref {
                    kind: kind.to_string(),
                    scope: split_path(scope),
                    path: split_path(path),
                    alias: (!alias.is_empty()).then(|| alias.to_string()),
                    line: l.parse().ok()?,
                })
            }
            ["d", scope, name, l] => s.definitions.push(Definition {
                scope: split_path(scope),
                name: name.to_string(),
                line: l.parse().ok()?,
            }),
            ["x", l, reason]
                if !reason.is_empty()
                    && reason.split(',').all(|r| {
                        matches!(
                            r,
                            "tab" | "indent" | "blockquote" | "list" | "fence" | "html"
                        )
                    }) =>
            {
                s.excluded_doc_blocks.push(DocExclusion {
                    line: l.parse().ok()?,
                    reason: reason.to_string(),
                })
            }
            _ => return None,
        }
    }
    Some(s)
}

/// The library crate name a Cargo manifest declares: `[lib] name`, else the package
/// name with `-` replaced by `_`. None for a manifest without a package.
pub fn lib_name(manifest: &str) -> Option<String> {
    let doc: toml_edit::DocumentMut = manifest.parse().ok()?;
    if let Some(name) = doc
        .get("lib")
        .and_then(|l| l.get("name"))
        .and_then(|n| n.as_str())
    {
        return Some(name.to_string());
    }
    doc.get("package")
        .and_then(|p| p.get("name"))
        .and_then(|n| n.as_str())
        .map(|n| n.replace('-', "_"))
}

/// One resolved file → file reference, before stamping.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Resolved {
    pub kind: String,
    pub source: String,
    pub target: String,
    pub evidence: String,
    pub inferred: bool,
}

fn parent(path: &str) -> &str {
    path.rsplit_once('/').map_or("", |(d, _)| d)
}

fn join(dir: &str, rest: &str) -> String {
    if dir.is_empty() {
        rest.to_string()
    } else {
        format!("{dir}/{rest}")
    }
}

/// Crate roots of a package directory, by Cargo's target layout.
fn is_root(pkg: &str, path: &str) -> bool {
    let rel = if pkg.is_empty() {
        path
    } else {
        match path.strip_prefix(&format!("{pkg}/")) {
            Some(r) => r,
            None => return false,
        }
    };
    if matches!(rel, "src/lib.rs" | "src/main.rs" | "build.rs") {
        return true;
    }
    for dir in ["src/bin/", "tests/", "examples/", "benches/"] {
        if let Some(rest) = rel.strip_prefix(dir) {
            let parts: Vec<&str> = rest.split('/').collect();
            return match parts.as_slice() {
                [f] => f.ends_with(".rs"),
                [_, "main.rs"] => true,
                _ => false,
            };
        }
    }
    false
}

/// A module: its crate root file and its path inside that crate.
type Module = (String, Vec<String>);

struct Tree {
    /// Module → the file that holds its items.
    file_of: BTreeMap<Module, String>,
    /// File → the modules it defines as a file (not inline), one per crate reaching it.
    contexts: BTreeMap<String, Vec<Module>>,
    misses: Vec<Value>,
}

fn build_tree(
    scans: &BTreeMap<String, (String, Scan)>,
    paths: &BTreeSet<String>,
    pkgs: &[String],
) -> Tree {
    let mut t = Tree {
        file_of: BTreeMap::new(),
        contexts: BTreeMap::new(),
        misses: Vec::new(),
    };
    let mut queue: VecDeque<(String, String, Vec<String>, String)> = VecDeque::new();
    for pkg in pkgs {
        for path in scans.keys().filter(|p| is_root(pkg, p)) {
            queue.push_back((
                path.clone(),
                path.clone(),
                Vec::new(),
                parent(path).to_string(),
            ));
        }
    }
    let mut seen: BTreeSet<(String, String)> = BTreeSet::new();
    while let Some((root, file, modpath, dir)) = queue.pop_front() {
        if !seen.insert((root.clone(), file.clone())) {
            continue;
        }
        t.file_of
            .insert((root.clone(), modpath.clone()), file.clone());
        t.contexts
            .entry(file.clone())
            .or_default()
            .push((root.clone(), modpath.clone()));
        let Some((_, scan)) = scans.get(&file) else {
            continue;
        };
        for m in &scan.mods {
            let mut mp = modpath.clone();
            mp.extend(m.scope.iter().cloned());
            mp.push(m.name.clone());
            if m.inline {
                t.file_of.insert((root.clone(), mp), file.clone());
                continue;
            }
            if m.path_attr {
                t.misses.push(json!({"kind": "module_path_attribute", "path": file, "module": m.name, "line": m.line, "reason": "#[path] module files are not followed"}));
                continue;
            }
            let mut base = dir.clone();
            for s in &m.scope {
                base = join(&base, s);
            }
            base = join(&base, &m.name);
            let flat = format!("{base}.rs");
            let nested = format!("{base}/mod.rs");
            let (child, child_dir) = match (paths.contains(&flat), paths.contains(&nested)) {
                (true, false) => (flat, base.clone()),
                (false, true) => (nested, base.clone()),
                (both, _) => {
                    let (kind, reason) = if both {
                        ("module_file_ambiguous", "both x.rs and x/mod.rs exist")
                    } else {
                        (
                            "module_file_missing",
                            "neither x.rs nor x/mod.rs is in the file set",
                        )
                    };
                    t.misses.push(json!({"kind": kind, "path": file, "module": m.name, "line": m.line, "reason": reason}));
                    continue;
                }
            };
            queue.push_back((root.clone(), child, mp, child_dir));
        }
    }
    for path in scans.keys() {
        if !t.contexts.contains_key(path) {
            t.misses.push(json!({"kind": "code_reference_unreached", "path": path, "reason": "not reached from a crate root through mod declarations"}));
        }
    }
    t
}

struct Ctx<'a> {
    tree: &'a Tree,
    scan: &'a Scan,
    root: &'a str,
    lib_root: Option<&'a str>,
    lib_name: Option<&'a str>,
}

/// Where a path lands: the deepest module reached, the segments past it (an item name
/// first) and whether every segment was a module.
struct Landing {
    module: Module,
    whole: bool,
    rest: Vec<String>,
    inferred: bool,
}

impl Ctx<'_> {
    fn exists(&self, root: &str, m: &[String]) -> bool {
        self.tree
            .file_of
            .contains_key(&(root.to_string(), m.to_vec()))
    }

    fn walk(&self, root: &str, mut m: Vec<String>, rest: &[String], inferred: bool) -> Landing {
        let mut used = 0;
        for s in rest {
            let mut next = m.clone();
            next.push(s.clone());
            if !self.exists(root, &next) {
                break;
            }
            m = next;
            used += 1;
        }
        Landing {
            module: (root.to_string(), m),
            whole: used == rest.len(),
            rest: rest[used..].to_vec(),
            inferred,
        }
    }

    fn land(
        &self,
        cur: &[String],
        scope: &[String],
        segs: &[String],
        depth: u8,
    ) -> Option<Landing> {
        let first = segs.first()?;
        match first.as_str() {
            "crate" => Some(self.walk(self.root, Vec::new(), &segs[1..], false)),
            "self" => Some(self.walk(self.root, cur.to_vec(), &segs[1..], false)),
            "super" => {
                let mut m = cur.to_vec();
                let mut k = 0;
                while segs.get(k).map(String::as_str) == Some("super") {
                    m.pop()?;
                    k += 1;
                }
                Some(self.walk(self.root, m, &segs[k..], false))
            }
            "Self" => None,
            x if EXTERNAL_ROOTS.contains(&x) => None,
            x => {
                let mut child = cur.to_vec();
                child.push(x.to_string());
                if self.exists(self.root, &child) {
                    return Some(self.walk(self.root, cur.to_vec(), segs, false));
                }
                if depth < 4 {
                    if let Some(alias) = self.alias(x, scope) {
                        let mut alias_cur = cur[..cur.len() - scope.len()].to_vec();
                        alias_cur.extend(alias.scope.iter().cloned());
                        let l = self.land(&alias_cur, &alias.scope, &alias.path, depth + 1)?;
                        if l.whole {
                            let (root, m) = l.module;
                            let mut next = self.walk(&root, m, &segs[1..], l.inferred);
                            next.inferred |= l.inferred;
                            return Some(next);
                        }
                        return Some(Landing { whole: false, ..l });
                    }
                }
                let lib = self.lib_root.filter(|r| *r != self.root)?;
                if Some(x) == self.lib_name {
                    return Some(self.walk(lib, Vec::new(), &segs[1..], false));
                }
                let second = segs.get(1)?;
                if self.lib_name.is_none() && self.exists(lib, std::slice::from_ref(second)) {
                    return Some(self.walk(lib, Vec::new(), &segs[1..], true));
                }
                None
            }
        }
    }

    /// The `use` leaf binding `name` in the innermost enclosing scope of `scope`.
    fn alias(&self, name: &str, scope: &[String]) -> Option<&Ref> {
        self.scan
            .refs
            .iter()
            .filter(|r| r.kind == "imports" && r.alias.as_deref() == Some(name))
            .filter(|r| scope.starts_with(&r.scope))
            .max_by_key(|r| r.scope.len())
    }
}

/// Follows one `use` re-export: when a path stops at an item name in module M and the
/// file holding M binds that name with a `use` in M's own scope (`pub use b::item;`),
/// the reference lands where that `use` points (Rust Reference, Use declarations:
/// re-exporting). None when there is nothing to follow.
fn reexport(
    tree: &Tree,
    scans: &BTreeMap<String, (String, Scan)>,
    ctx: &Ctx<'_>,
    l: &Landing,
) -> Option<Landing> {
    let name = l.rest.first()?;
    let file = tree.file_of.get(&l.module)?;
    let (_, scan) = scans.get(file)?;
    let (root, mp) = &l.module;
    let (_, file_mod) = tree
        .contexts
        .get(file)?
        .iter()
        .filter(|(r, m)| r == root && mp.starts_with(m))
        .max_by_key(|(_, m)| m.len())?;
    let inline = mp[file_mod.len()..].to_vec();
    let fctx = Ctx {
        tree,
        scan,
        root,
        lib_root: ctx.lib_root,
        lib_name: ctx.lib_name,
    };
    let alias = fctx.alias(name, &inline).filter(|a| a.scope == inline)?;
    let next = fctx.land(mp, &inline, &alias.path, 0)?;
    let inferred = l.inferred || next.inferred;
    if !next.whole {
        return Some(Landing { inferred, ..next });
    }
    let (r, m) = next.module;
    Some(fctx.walk(&r, m, &l.rest[1..], inferred))
}

/// Resolves every stored scan against the file set. `scans` maps a Rust file path to
/// its content digest and scan; `manifests` maps a Cargo manifest directory to its
/// library crate name. Returns the references (sorted, deduplicated by evidence) and
/// the coverage misses of module resolution.
pub fn resolve(
    scans: &BTreeMap<String, (String, Scan)>,
    paths: &BTreeSet<String>,
    manifests: &BTreeMap<String, Option<String>>,
) -> (Vec<Resolved>, Vec<Value>) {
    let pkgs: Vec<String> = if manifests.is_empty() {
        vec![String::new()]
    } else {
        manifests.keys().cloned().collect()
    };
    let tree = build_tree(scans, paths, &pkgs);
    let mut out: BTreeSet<(String, String, String, String, bool)> = BTreeSet::new();
    for (file, (sha, scan)) in scans {
        for (root, modpath) in tree.contexts.get(file).into_iter().flatten() {
            let pkg = pkgs
                .iter()
                .filter(|p| is_root(p, root))
                .max_by_key(|p| p.len())
                .cloned()
                .unwrap_or_default();
            let lib_root = join(&pkg, "src/lib.rs");
            let ctx = Ctx {
                tree: &tree,
                scan,
                root,
                lib_root: scans.contains_key(&lib_root).then_some(lib_root.as_str()),
                lib_name: manifests.get(&pkg).and_then(|n| n.as_deref()),
            };
            for r in scan.refs.iter().filter(|r| r.kind != "documents") {
                let mut cur = modpath.clone();
                cur.extend(r.scope.iter().cloned());
                let Some(mut l) = ctx.land(&cur, &r.scope, &r.path, 0) else {
                    continue;
                };
                for _ in 0..4 {
                    match reexport(&tree, scans, &ctx, &l) {
                        Some(next) => l = next,
                        None => break,
                    }
                }
                let Some(target) = tree.file_of.get(&l.module) else {
                    continue;
                };
                if target == file {
                    continue;
                }
                out.insert((
                    r.kind.clone(),
                    file.clone(),
                    target.clone(),
                    format!("file:{file}#L{}@{sha}", r.line),
                    l.inferred,
                ));
            }
        }
    }
    let mut refs: Vec<Resolved> = out
        .into_iter()
        .map(|(kind, source, target, evidence, inferred)| Resolved {
            kind,
            source,
            target,
            evidence,
            inferred,
        })
        .collect();
    let (mut docs, mut misses) = resolve_docs(scans, manifests, &pkgs, &tree);
    refs.append(&mut docs);
    for (file, (_, scan)) in scans {
        for excluded in &scan.excluded_doc_blocks {
            misses.push(json!({"kind":"doc_reference_unsupported", "path":file, "line":excluded.line, "count":1, "reason":format!("unsupported doc block: {}", excluded.reason)}));
        }
    }
    misses.extend(tree.misses);
    (refs, misses)
}

/// Doc paths must identify exactly one module-level definition. Root-relative prose
/// is accepted as an inferred candidate; aliases, re-exports, associated items and
/// compiler namespace/visibility rules are not claimed to be resolved.
fn resolve_docs(
    scans: &BTreeMap<String, (String, Scan)>,
    manifests: &BTreeMap<String, Option<String>>,
    pkgs: &[String],
    tree: &Tree,
) -> (Vec<Resolved>, Vec<Value>) {
    let mut refs = Vec::new();
    let mut misses = Vec::new();
    for (file, (sha, scan)) in scans {
        for r in scan.refs.iter().filter(|r| r.kind == "documents") {
            let mut candidates = BTreeSet::new();
            for (root, modpath) in tree.contexts.get(file).into_iter().flatten() {
                let pkg = pkgs
                    .iter()
                    .filter(|p| is_root(p, root))
                    .max_by_key(|p| p.len())
                    .cloned()
                    .unwrap_or_default();
                let lib = join(&pkg, "src/lib.rs");
                let ctx = Ctx {
                    tree,
                    scan,
                    root,
                    lib_root: scans.contains_key(&lib).then_some(lib.as_str()),
                    lib_name: manifests.get(&pkg).and_then(|n| n.as_deref()),
                };
                let mut cur = modpath.clone();
                cur.extend(r.scope.iter().cloned());
                let first = &r.path[0];
                let landing = if matches!(first.as_str(), "crate" | "self" | "super") {
                    ctx.land(&cur, &r.scope, &r.path, 0)
                } else if ctx.exists(root, &[cur.clone(), vec![first.clone()]].concat()) {
                    Some(ctx.walk(root, cur, &r.path, false))
                } else if scan.refs.iter().any(|a| {
                    a.kind == "imports"
                        && a.alias.as_ref() == Some(first)
                        && r.scope.starts_with(&a.scope)
                }) || scan
                    .definitions
                    .iter()
                    .any(|d| d.name == *first && r.scope.starts_with(&d.scope))
                {
                    None
                } else if ctx.exists(root, std::slice::from_ref(first)) {
                    Some(ctx.walk(root, Vec::new(), &r.path, true))
                } else if ctx.lib_name == Some(first.as_str()) {
                    ctx.lib_root
                        .map(|lib| ctx.walk(lib, Vec::new(), &r.path[1..], false))
                } else {
                    None
                };
                let Some(l) = landing else {
                    continue;
                };
                if l.rest.len() != 1 {
                    continue;
                }
                let Some(target) = tree.file_of.get(&l.module) else {
                    continue;
                };
                let Some((target_sha, target_scan)) = scans.get(target) else {
                    continue;
                };
                let Some((_, file_scope)) = tree
                    .contexts
                    .get(target)
                    .into_iter()
                    .flatten()
                    .filter(|(rt, mp)| rt == &l.module.0 && l.module.1.starts_with(mp))
                    .max_by_key(|(_, mp)| mp.len())
                else {
                    continue;
                };
                let scope = &l.module.1[file_scope.len()..];
                for (ordinal, d) in target_scan
                    .definitions
                    .iter()
                    .enumerate()
                    .filter(|(_, d)| d.scope == scope && d.name == l.rest[0])
                {
                    candidates.insert((
                        target.clone(),
                        format!("file:{target}#L{}@{target_sha}", d.line),
                        ordinal,
                    ));
                }
            }
            if candidates.len() != 1 {
                misses.push(json!({"kind": "doc_reference_missing", "path": file, "line": r.line, "reference": r.path.join("::"), "reason": if candidates.is_empty() { "unresolved" } else { "ambiguous" }}));
                continue;
            }
            let (target, evidence, _) = candidates.into_iter().next().unwrap();
            if target != *file {
                for evidence in [format!("file:{file}#L{}@{sha}", r.line), evidence] {
                    refs.push(Resolved {
                        kind: "documents".into(),
                        source: file.clone(),
                        target: target.clone(),
                        evidence,
                        inferred: true,
                    });
                }
            }
        }
    }
    (refs, misses)
}

/// Stamps resolved references as C4 edges, one per (kind, source, target) with every
/// evidence line. An edge whose crate name was inferred anywhere is `inferred`.
pub fn edges(refs: &[Resolved], snapshot_id: &str, observed_at: &str) -> Vec<OntologyEdge> {
    let mut grouped: BTreeMap<(String, String, String), (BTreeSet<String>, bool)> = BTreeMap::new();
    for r in refs {
        let e = grouped
            .entry((r.kind.clone(), r.source.clone(), r.target.clone()))
            .or_default();
        e.0.insert(r.evidence.clone());
        e.1 |= r.inferred;
    }
    let extractor = Extractor {
        name: EXTRACTOR_NAME.into(),
        version: EXTRACTOR_VERSION.into(),
    };
    grouped
        .into_iter()
        .map(|((kind, source, target), (ev, inferred))| {
            let evidence_refs: Vec<String> = ev.into_iter().collect();
            let uncertainty = if kind == "documents" {
                Some("doc-lexical-reference; rustdoc-binding-not-checked".to_string())
            } else if inferred && kind == "calls" {
                Some(INFERRED_CALL_UNCERTAINTY.to_string())
            } else if inferred {
                Some(CRATE_NAME_UNCERTAINTY.to_string())
            } else if kind == "calls" {
                Some(CALL_UNCERTAINTY.to_string())
            } else {
                None
            };
            OntologyEdge {
                edge_id: ontology::edge_id(
                    &kind,
                    &format!("file:{source}"),
                    &format!("file:{target}"),
                    &evidence_refs,
                ),
                basis: if inferred {
                    "inferred"
                } else {
                    "static_resolved"
                }
                .into(),
                kind,
                source_id: format!("file:{source}"),
                target_id: format!("file:{target}"),
                verification: "unreviewed".into(),
                evidence_refs,
                extractor: extractor.clone(),
                snapshot_id: snapshot_id.to_string(),
                observed_at: observed_at.to_string(),
                valid_from: snapshot_id.to_string(),
                valid_to: None,
                uncertainty,
            }
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn paths_of(s: &Scan) -> Vec<(String, String)> {
        s.refs
            .iter()
            .map(|r| (r.kind.clone(), r.path.join("::")))
            .collect()
    }

    #[test]
    fn comments_strings_and_chars_yield_no_reference() {
        let src = r####"
//! crate::a::doc_line()
/// crate::b::outer_doc()
/* crate::c::block() /* nested crate::d::x() */ crate::e::y() */
fn f() {
    let s = "crate::f::in_string()";
    let r = r#"crate::g::raw() "quoted" "#;
    let b = b"crate::h::bytes()";
    let c = '"'; let q = '\''; let l: &'static str = "x";
    crate::real::call();
}
"####;
        assert_eq!(
            paths_of(&scan(src)),
            [("calls".to_string(), "crate::real::call".to_string())]
        );
    }

    #[test]
    fn use_trees_expand_groups_self_aliases_and_globs() {
        let s =
            scan("use crate::a::{self, b as c, d::{e, f::*}, g as _};\npub(crate) use ::x::y;\n");
        let got: Vec<(String, Option<String>)> = s
            .refs
            .iter()
            .map(|r| (r.path.join("::"), r.alias.clone()))
            .collect();
        assert_eq!(
            got,
            [
                ("crate::a".to_string(), Some("a".to_string())),
                ("crate::a::b".to_string(), Some("c".to_string())),
                ("crate::a::d::e".to_string(), Some("e".to_string())),
                ("crate::a::d::f".to_string(), None),
                ("crate::a::g".to_string(), None),
                ("x::y".to_string(), Some("y".to_string())),
            ]
        );
        assert!(s.refs.iter().all(|r| r.kind == "imports"));
    }

    #[test]
    fn calls_need_two_segments_and_a_call_or_macro() {
        let s = scan(
            "fn f() { a::b(); one(); x.y::z(); T::<u8>::new(); m::n!(1); p::q != r; Vec::<u8>::with_capacity(1); c::K; }",
        );
        assert_eq!(
            paths_of(&s),
            [
                ("calls".to_string(), "a::b".to_string()),
                ("calls".to_string(), "T::new".to_string()),
                ("calls".to_string(), "m::n".to_string()),
                ("calls".to_string(), "Vec::with_capacity".to_string()),
            ]
        );
    }

    #[test]
    fn inline_modules_set_the_scope_and_path_attributes_are_seen() {
        let s = scan("mod a;\n#[cfg(test)]\nmod tests {\n use super::*;\n mod deep { fn g() { x::y(); } }\n}\n#[path = \"o.rs\"]\npub mod other;\nfn h() { z::w(); }\n");
        let mods: Vec<(String, String, bool, bool)> = s
            .mods
            .iter()
            .map(|m| (m.scope.join("::"), m.name.clone(), m.inline, m.path_attr))
            .collect();
        assert_eq!(
            mods,
            [
                (String::new(), "a".into(), false, false),
                (String::new(), "tests".into(), true, false),
                ("tests".into(), "deep".into(), true, false),
                (String::new(), "other".into(), false, true),
            ]
        );
        let scopes: Vec<String> = s.refs.iter().map(|r| r.scope.join("::")).collect();
        assert_eq!(scopes, ["tests", "tests::deep", ""]);
    }

    #[test]
    fn encode_round_trips() {
        let s = scan("mod a { use crate::b::{c as d, e}; }\nfn f() { crate::g::h(); }\n");
        assert_eq!(decode(&encode(&s)), Some(s));
        assert_eq!(decode("x\tbad\n"), None);
    }

    fn files(list: &[(&str, &str)]) -> (BTreeMap<String, (String, Scan)>, BTreeSet<String>) {
        let scans = list
            .iter()
            .map(|(p, src)| (p.to_string(), ("sha256:00".to_string(), scan(src))))
            .collect();
        let paths = list.iter().map(|(p, _)| p.to_string()).collect();
        (scans, paths)
    }

    fn triples(refs: &[Resolved]) -> Vec<String> {
        let mut v: Vec<String> = refs
            .iter()
            .map(|r| {
                format!(
                    "{} {} {}{}",
                    r.source,
                    r.kind,
                    r.target,
                    if r.inferred { " ?" } else { "" }
                )
            })
            .collect();
        v.sort();
        v.dedup();
        v
    }

    #[test]
    fn resolution_follows_the_mod_tree_aliases_and_the_lib_crate() {
        let (scans, paths) = files(&[
            ("src/lib.rs", "pub mod a; pub mod forge; pub use a::item; pub fn digest() {}"),
            ("src/a.rs", "use crate::forge::outbox::{parse, op as o};\nfn f() { crate::digest(); super::forge::outbox::x(); o(); }\n#[cfg(test)] mod tests { use super::*; }"),
            ("src/forge/mod.rs", "pub mod outbox;"),
            ("src/forge/outbox.rs", "pub fn parse() {}"),
            ("src/main.rs", "use moai_cli::{a, forge};\nfn main() { moai_cli::forge::outbox::parse(); a::item(); std::fs::read(\"x\"); }"),
        ]);
        let mut manifests = BTreeMap::new();
        manifests.insert(String::new(), Some("moai_cli".to_string()));
        let (refs, misses) = resolve(&scans, &paths, &manifests);
        assert!(misses.is_empty(), "{misses:?}");
        assert_eq!(
            triples(&refs),
            [
                "src/a.rs calls src/forge/outbox.rs",
                "src/a.rs calls src/lib.rs",
                "src/a.rs imports src/forge/outbox.rs",
                "src/lib.rs imports src/a.rs",
                "src/main.rs calls src/a.rs",
                "src/main.rs calls src/forge/outbox.rs",
                "src/main.rs imports src/a.rs",
                "src/main.rs imports src/forge/mod.rs",
            ]
        );
        // Without the manifest the library name is inferred and says so.
        let (refs, _) = resolve(&scans, &paths, &BTreeMap::new());
        assert!(triples(&refs).contains(&"src/main.rs calls src/forge/outbox.rs ?".to_string()));
    }

    #[test]
    fn re_exports_are_followed_to_the_defining_file() {
        let (scans, paths) = files(&[
            ("src/lib.rs", "pub mod baseline; pub mod audit; pub mod inner { pub use crate::baseline::Tree; }\npub use baseline::tree_digest;\npub fn digest() {}"),
            ("src/baseline.rs", "pub fn tree_digest() {} pub struct Tree;"),
            ("src/audit.rs", "use crate::Tree as _;\nfn f() { crate::tree_digest(); crate::digest(); crate::inner::Tree::new(); }"),
        ]);
        let (refs, _) = resolve(&scans, &paths, &BTreeMap::new());
        let got: Vec<String> = triples(&refs)
            .into_iter()
            .filter(|t| t.starts_with("src/audit.rs"))
            .collect();
        // `crate::Tree` names nothing the crate root binds, so it stays on lib.rs.
        assert_eq!(
            got,
            [
                "src/audit.rs calls src/baseline.rs",
                "src/audit.rs calls src/lib.rs",
                "src/audit.rs imports src/lib.rs",
            ]
        );
        let lines: Vec<&str> = refs
            .iter()
            .filter(|r| r.source == "src/audit.rs" && r.target == "src/baseline.rs")
            .map(|r| r.evidence.as_str())
            .collect();
        assert_eq!(lines, ["file:src/audit.rs#L2@sha256:00"]);
    }

    #[test]
    fn unresolvable_modules_are_coverage_not_silence() {
        let (scans, paths) = files(&[
            (
                "src/lib.rs",
                "mod gone;\nmod both;\n#[path = \"x.rs\"] mod moved;",
            ),
            ("src/both.rs", ""),
            ("src/both/mod.rs", ""),
            ("tools/loose.rs", "fn f() { crate::a::b(); }"),
        ]);
        let (_, misses) = resolve(&scans, &paths, &BTreeMap::new());
        let kinds: Vec<&str> = misses.iter().map(|m| m["kind"].as_str().unwrap()).collect();
        assert_eq!(
            kinds,
            [
                "module_file_missing",
                "module_file_ambiguous",
                "module_path_attribute",
                "code_reference_unreached",
                "code_reference_unreached",
                "code_reference_unreached",
            ]
        );
    }

    #[test]
    fn a_target_file_appearing_changes_resolution_without_rescanning_the_source() {
        let src = "mod b; fn f() { crate::b::g(); }";
        let (scans, paths) = files(&[("src/lib.rs", src)]);
        let (refs, misses) = resolve(&scans, &paths, &BTreeMap::new());
        assert!(refs.is_empty());
        assert_eq!(misses[0]["kind"], "module_file_missing");
        let (scans, paths) = files(&[("src/lib.rs", src), ("src/b.rs", "pub fn g() {}")]);
        let (refs, _) = resolve(&scans, &paths, &BTreeMap::new());
        assert_eq!(triples(&refs), ["src/lib.rs calls src/b.rs"]);
    }

    #[test]
    fn lib_name_reads_lib_then_package() {
        assert_eq!(
            lib_name("[package]\nname = \"moai-cli\"\n").as_deref(),
            Some("moai_cli")
        );
        assert_eq!(
            lib_name("[package]\nname = \"a-b\"\n[lib]\nname = \"core_x\"\n").as_deref(),
            Some("core_x")
        );
        assert_eq!(lib_name("[workspace]\nmembers = []\n"), None);
    }

    #[test]
    fn stamped_edges_are_schema_valid_with_honest_basis() {
        let refs = vec![
            Resolved {
                kind: "calls".into(),
                source: "src/a.rs".into(),
                target: "src/b.rs".into(),
                evidence: "file:src/a.rs#L1@sha256:00".into(),
                inferred: false,
            },
            Resolved {
                kind: "imports".into(),
                source: "src/main.rs".into(),
                target: "src/b.rs".into(),
                evidence: "file:src/main.rs#L2@sha256:01".into(),
                inferred: true,
            },
            Resolved {
                kind: "calls".into(),
                source: "src/main.rs".into(),
                target: "src/b.rs".into(),
                evidence: "file:src/main.rs#L3@sha256:01".into(),
                inferred: true,
            },
        ];
        let e = edges(&refs, "abc", "2026-11-11T00:00:00Z");
        for edge in &e {
            ontology::validate_edge(edge).unwrap();
        }
        assert_eq!(
            (e[0].basis.as_str(), e[0].uncertainty.as_deref()),
            ("static_resolved", Some(CALL_UNCERTAINTY))
        );
        // Edges are ordered by (kind, source, target): calls a→b, calls main→b, imports main→b.
        assert_eq!(
            (e[1].basis.as_str(), e[1].uncertainty.as_deref()),
            ("inferred", Some(INFERRED_CALL_UNCERTAINTY))
        );
        assert_eq!(
            (e[2].basis.as_str(), e[2].uncertainty.as_deref()),
            ("inferred", Some(CRATE_NAME_UNCERTAINTY))
        );
    }

    #[test]
    fn use_group_nesting_is_bounded_at_max_use_depth() {
        // MAX_USE_DEPTH = 32: a leaf inside 32 nested groups is kept, inside 33 skipped.
        let nested = |n: usize| {
            format!(
                "use {}leaf{};\nfn f() {{ crate::after::call(); }}\n",
                "a::{".repeat(n),
                "}".repeat(n)
            )
        };
        let kept = scan(&nested(32));
        assert!(kept
            .refs
            .iter()
            .any(|r| r.path.last().map(String::as_str) == Some("leaf")));
        let cut = scan(&nested(33));
        assert!(!cut
            .refs
            .iter()
            .any(|r| r.path.last().map(String::as_str) == Some("leaf")));
        // A pathological depth neither overflows the stack nor hides the code after it.
        let deep = scan(&nested(100_000));
        assert_eq!(
            deep.refs.last().map(|r| r.path.join("::")),
            Some("crate::after::call".to_string())
        );
    }
}
