//! `review-package` (#2614): the deterministic review package an AUDIT hands to a reviewer.
//!
//! It holds the diff from the merge base to the head, the changed files, the review
//! instructions (`REVIEW.md`, its `@` includes, and the subdirectory `REVIEW.md` files on the
//! changed paths) and the project context (`AGENTS.md`, `CLAUDE.md` at the root and above each
//! changed file). Instructions and context are read from the BASE commit only, never from the
//! change under review, following Claude Code's Code Review documentation
//! (code.claude.com/docs/en/code-review-customize): one level of `@name` includes beside the
//! root file, at most eight different files, root plus includes under 150,000 bytes combined
//! (a file that would pass the limit is skipped whole), more than 64 Markdown links skip the
//! root file, subdirectory files share their own 150,000 bytes, and directories with test,
//! fixture, vendored or generated names are skipped. Branch protection cannot be checked
//! offline, so includes always apply and the package says `protection: not_checked`
//! (design 6839486269, decision D1). Skill includes are not supported yet (D2).
//!
//! Git runs with the caller's configuration and environment removed
//! (`audit_delta::command`) and with fixed diff options, so the same base and head give the
//! same bytes on every machine and in every tool (AC-2). The repository's own diff settings are
//! overridden on the command line, attributes come from an empty tree instead of the worktree,
//! and a repository with `.git/info/attributes` is refused (audit FAIL 6839909483 B1). Paths and
//! diff text that are not UTF-8 are carried with U+FFFD replacements (deterministic, and
//! `diff_sha256` hashes that replaced text). Reads only; writes nothing.
//!
//! The package also carries a `checklist` (SPEC-20261120-review-checklist-by-diff): the items of
//! `tools/review_checklist.json`, read from the BASE commit like `REVIEW.md`, whose trigger the
//! diff meets. An item fires when a changed path is in its `trigger.paths`
//! (`repo_scans::in_scope`) and, when it has `diff_pattern`, an added line of that path contains
//! one of the strings literally. A table that is not valid is skipped whole with its reason.

use serde::Serialize;
use serde_json::Value;
use std::collections::BTreeSet;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

pub const REVIEW_FILE: &str = "REVIEW.md";
/// The checklist table inside a repository.
pub const CHECKLIST_FILE: &str = "tools/review_checklist.json";
pub const BYTE_LIMIT: usize = 150_000;
pub const INCLUDE_LIMIT: usize = 8;
pub const LINK_LIMIT: usize = 64;
const CONTEXT_FILES: [&str; 2] = ["AGENTS.md", "CLAUDE.md"];
/// Directory names whose subdirectory `REVIEW.md` is not read, matched whole and in any case
/// at any depth ([customize]).
const SKIPPED_DIRS: [&str; 31] = [
    "test",
    "tests",
    "__tests__",
    "testdata",
    "fixtures",
    "__fixtures__",
    "golden",
    "corpus",
    "snapshots",
    "__snapshots__",
    "example",
    "examples",
    "sample",
    "samples",
    "demo",
    "demos",
    "generated",
    "vendor",
    "third_party",
    "third-party",
    "thirdparty",
    "node_modules",
    "bower_components",
    "external",
    "extern",
    "deps",
    "godeps",
    "pods",
    "carthage",
    "site-packages",
    "contrib",
];

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Changed {
    pub path: String,
    pub status: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct File {
    pub path: String,
    /// `root`, `include` or `subdirectory` for review files; the file name for context.
    pub role: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub included_by: Option<String>,
    pub bytes: usize,
    pub sha256: String,
    pub text: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Skipped {
    pub path: String,
    pub reason: String,
}

/// Why a checklist item is in the package: the changed paths that met its trigger and the first
/// `diff_pattern` string found in them (`null` for an item without patterns).
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Matched {
    pub paths: Vec<String>,
    pub pattern: Option<String>,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct ChecklistItem {
    pub id: String,
    pub check: String,
    pub source: String,
    pub added: String,
    pub matched: Matched,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Package {
    pub schema_version: u32,
    pub base: String,
    pub head: String,
    pub merge_base: String,
    pub protection: String,
    pub changed: Vec<Changed>,
    pub diff: String,
    pub diff_sha256: String,
    pub review: Vec<File>,
    pub context: Vec<File>,
    pub checklist: Vec<ChecklistItem>,
    pub skipped: Vec<Skipped>,
    pub package_digest: String,
}

/// One item of the checklist table, checked for every field the judge reads.
#[derive(Debug, Clone, PartialEq, Eq)]
struct Item {
    id: String,
    scope: Vec<(String, Vec<String>)>,
    patterns: Vec<String>,
    check: String,
    source: String,
    added: String,
}

const TOP_KEYS: [&str; 3] = ["schema_version", "comment", "items"];
const ITEM_KEYS: [&str; 5] = ["id", "trigger", "check", "source", "added"];
const TRIGGER_KEYS: [&str; 2] = ["paths", "diff_pattern"];
const PATH_KEYS: [&str; 2] = ["prefix", "suffixes"];

fn unknown_key(obj: &serde_json::Map<String, Value>, allowed: &[&str]) -> Option<String> {
    obj.keys().find(|k| !allowed.contains(&k.as_str())).cloned()
}

fn text_of<'a>(v: &'a Value, what: &str) -> std::result::Result<&'a str, String> {
    v.as_str()
        .filter(|s| !s.trim().is_empty())
        .ok_or_else(|| format!("{what} is not a non-empty string"))
}

/// A list whose entries are all non-empty strings; the list itself may be empty.
fn strings_of(v: &Value, what: &str) -> std::result::Result<Vec<String>, String> {
    let list = v
        .as_array()
        .ok_or_else(|| format!("{what} is not a list"))?;
    list.iter()
        .map(|s| {
            s.as_str()
                .filter(|s| !s.is_empty())
                .map(str::to_string)
                .ok_or_else(|| format!("{what} holds an empty or non-string entry"))
        })
        .collect()
}

/// Whether `s` names an `Issue #<n>`, a `PR #<n>` or a `SPEC-<8 digits>-<slug>`.
fn names_source(s: &str) -> bool {
    let number_after = |prefix: &str| {
        s.match_indices(prefix)
            .any(|(i, _)| s[i + prefix.len()..].starts_with(|c: char| c.is_ascii_digit()))
    };
    if number_after("Issue #") || number_after("PR #") {
        return true;
    }
    s.match_indices("SPEC-").any(|(i, _)| {
        let b = s[i + 5..].as_bytes();
        b.len() > 9
            && b[..8].iter().all(u8::is_ascii_digit)
            && b[8] == b'-'
            && (b[9].is_ascii_lowercase() || b[9].is_ascii_digit())
    })
}

/// `YYYY-MM-DD` with a month of 01 to 12 and a day of 01 to 31.
fn valid_date(s: &str) -> bool {
    let b = s.as_bytes();
    s.is_ascii()
        && b.len() == 10
        && b[4] == b'-'
        && b[7] == b'-'
        && b.iter()
            .enumerate()
            .all(|(i, c)| i == 4 || i == 7 || c.is_ascii_digit())
        && matches!(s[5..7].parse::<u32>(), Ok(1..=12))
        && matches!(s[8..10].parse::<u32>(), Ok(1..=31))
}

fn parse_item(n: usize, v: &Value) -> std::result::Result<Item, String> {
    let at = |m: String| format!("item {n}: {m}");
    let obj = v.as_object().ok_or_else(|| at("is not an object".into()))?;
    if let Some(k) = unknown_key(obj, &ITEM_KEYS) {
        return Err(at(format!("unknown key `{k}`")));
    }
    if let Some(k) = ITEM_KEYS.iter().find(|k| !obj.contains_key(**k)) {
        return Err(at(format!("missing `{k}`")));
    }
    let id = text_of(&obj["id"], "`id`").map_err(&at)?;
    if !id
        .chars()
        .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '-')
    {
        return Err(at(format!(
            "`id` {id:?} is not lowercase letters, digits and hyphens"
        )));
    }
    let check = text_of(&obj["check"], "`check`").map_err(&at)?;
    let source = text_of(&obj["source"], "`source`").map_err(&at)?;
    if !names_source(source) {
        return Err(at("`source` names no Issue #<n>, PR #<n> or SPEC-<id>".into()));
    }
    let added = text_of(&obj["added"], "`added`").map_err(&at)?;
    if !valid_date(added) {
        return Err(at("`added` is not a YYYY-MM-DD date".into()));
    }
    let trigger = obj["trigger"]
        .as_object()
        .ok_or_else(|| at("`trigger` is not an object".into()))?;
    if let Some(k) = unknown_key(trigger, &TRIGGER_KEYS) {
        return Err(at(format!("unknown trigger key `{k}`")));
    }
    let paths = trigger
        .get("paths")
        .and_then(Value::as_array)
        .filter(|p| !p.is_empty())
        .ok_or_else(|| at("`trigger.paths` is not a non-empty list".into()))?;
    let mut scope = Vec::new();
    for entry in paths {
        let entry = entry
            .as_object()
            .ok_or_else(|| at("a `trigger.paths` entry is not an object".into()))?;
        if let Some(k) = unknown_key(entry, &PATH_KEYS) {
            return Err(at(format!("unknown `trigger.paths` key `{k}`")));
        }
        let prefix = entry
            .get("prefix")
            .and_then(Value::as_str)
            .ok_or_else(|| at("a `trigger.paths` entry has no string `prefix`".into()))?;
        let suffixes = match entry.get("suffixes") {
            None => Vec::new(),
            Some(v) => strings_of(v, "`suffixes`").map_err(&at)?,
        };
        scope.push((prefix.to_string(), suffixes));
    }
    let patterns = match trigger.get("diff_pattern") {
        None => Vec::new(),
        Some(v) => {
            let list = strings_of(v, "`diff_pattern`").map_err(&at)?;
            if list.is_empty() {
                return Err(at("`diff_pattern` is an empty list".into()));
            }
            list
        }
    };
    Ok(Item {
        id: id.to_string(),
        scope,
        patterns,
        check: check.to_string(),
        source: source.to_string(),
        added: added.to_string(),
    })
}

/// The items of a table's text, or the first problem found.
fn parse_checklist(text: &str) -> std::result::Result<Vec<Item>, String> {
    let doc: Value = serde_json::from_str(text).map_err(|e| format!("not JSON: {e}"))?;
    let top = doc
        .as_object()
        .ok_or("the table is not an object".to_string())?;
    if let Some(k) = unknown_key(top, &TOP_KEYS) {
        return Err(format!("unknown key `{k}`"));
    }
    if top.get("schema_version").and_then(Value::as_u64) != Some(1) {
        return Err("unsupported schema_version; supported: 1".into());
    }
    let list = top
        .get("items")
        .and_then(Value::as_array)
        .ok_or("`items` is not a list".to_string())?;
    let mut items: Vec<Item> = Vec::new();
    for (n, v) in list.iter().enumerate() {
        let item = parse_item(n, v)?;
        if items.iter().any(|i| i.id == item.id) {
            return Err(format!("item {n}: duplicate `id` {:?}", item.id));
        }
        items.push(item);
    }
    Ok(items)
}

/// The diff's per-file sections in order, when there is exactly one for each changed path.
fn diff_sections(diff: &str, expected: usize) -> Option<Vec<&str>> {
    let mut starts: Vec<usize> = Vec::new();
    let mut at = 0;
    for line in diff.split_inclusive('\n') {
        if line.starts_with("diff --git ") {
            starts.push(at);
        }
        at += line.len();
    }
    if starts.len() != expected {
        return None;
    }
    Some(
        starts
            .iter()
            .enumerate()
            .map(|(i, s)| &diff[*s..starts.get(i + 1).copied().unwrap_or(diff.len())])
            .collect(),
    )
}

/// The added lines of one file's section, without the leading `+`: only lines after the first
/// hunk header count, so the `+++` file header is never one.
fn added_lines(section: &str) -> Vec<&str> {
    let mut in_hunk = false;
    let mut out = Vec::new();
    for line in section.lines() {
        if line.starts_with("@@ ") {
            in_hunk = true;
        } else if in_hunk {
            if let Some(text) = line.strip_prefix('+') {
                out.push(text);
            }
        }
    }
    out
}

/// The items the diff triggers, in table order. `sections` is the diff split per changed path;
/// without it the items that need a `diff_pattern` are not judged.
fn judge_checklist(
    items: &[Item],
    changed: &[Changed],
    sections: Option<&[&str]>,
) -> Vec<ChecklistItem> {
    let mut out = Vec::new();
    for item in items {
        let in_scope: Vec<usize> = changed
            .iter()
            .enumerate()
            .filter(|(_, c)| crate::repo_scans::in_scope(&c.path, &item.scope))
            .map(|(i, _)| i)
            .collect();
        let mut paths = Vec::new();
        let mut pattern: Option<usize> = None;
        if item.patterns.is_empty() {
            paths = in_scope.iter().map(|&i| changed[i].path.clone()).collect();
        } else if let Some(sections) = sections {
            for i in in_scope {
                let lines = added_lines(sections[i]);
                let found = item
                    .patterns
                    .iter()
                    .position(|p| lines.iter().any(|l| l.contains(p.as_str())));
                if let Some(k) = found {
                    paths.push(changed[i].path.clone());
                    pattern = Some(pattern.map_or(k, |f| f.min(k)));
                }
            }
        }
        if paths.is_empty() {
            continue;
        }
        out.push(ChecklistItem {
            id: item.id.clone(),
            check: item.check.clone(),
            source: item.source.clone(),
            added: item.added.clone(),
            matched: Matched {
                paths,
                pattern: pattern.map(|k| item.patterns[k].clone()),
            },
        });
    }
    out
}

/// The checklist of the BASE commit judged against the diff, and the skips it leaves. No table
/// at base is an empty list without a skip, like a missing `REVIEW.md`.
fn read_checklist(
    repo: &Path,
    base: &str,
    changed: &[Changed],
    diff: &str,
) -> Result<(Vec<ChecklistItem>, Vec<Skipped>)> {
    let skip = |reason: &str| Skipped {
        path: CHECKLIST_FILE.into(),
        reason: reason.into(),
    };
    let text = match blob_text(repo, base, CHECKLIST_FILE)? {
        None => return Ok((Vec::new(), Vec::new())),
        Some(Err(reason)) => return Ok((Vec::new(), vec![skip(reason)])),
        Some(Ok(text)) => text,
    };
    if text.len() > BYTE_LIMIT {
        return Ok((Vec::new(), vec![skip("byte-limit")]));
    }
    let items = match parse_checklist(&text) {
        Ok(items) => items,
        Err(problem) => {
            return Ok((
                Vec::new(),
                vec![skip(&format!("invalid-checklist: {problem}"))],
            ))
        }
    };
    let sections = diff_sections(diff, changed.len());
    let mut skips = Vec::new();
    if sections.is_none() && items.iter().any(|i| !i.patterns.is_empty()) {
        skips.push(skip("diff-sections-mismatch"));
    }
    Ok((
        judge_checklist(&items, changed, sections.as_deref()),
        skips,
    ))
}

fn git(repo: &Path, args: &[&str]) -> Result<Vec<u8>> {
    let out = crate::audit_delta::command(repo)
        .args(args)
        .output()
        .map_err(|e| format!("review package git failed: cannot execute Git: {e}"))?;
    if !out.status.success() {
        return Err(format!(
            "review package git failed: {}",
            crate::identity::failure_summary(&out.stderr)
        ));
    }
    Ok(out.stdout)
}

fn null_device() -> &'static str {
    if cfg!(windows) {
        "NUL"
    } else {
        "/dev/null"
    }
}

/// `git --attr-source=<empty tree> -c … diff`: the configuration keys a repository could set
/// and the command line has no option for.
fn diff_prefix(attr_source: &str) -> Vec<&str> {
    vec![
        attr_source,
        "-c",
        "core.quotePath=true",
        "-c",
        "core.bigFileThreshold=512m",
        "-c",
        "diff.suppressBlankEmpty=false",
        "-c",
        "diff.noprefix=false",
        "-c",
        "diff.mnemonicPrefix=false",
        "diff",
    ]
}

/// The repository's top level. Every path in the package is relative to it, and git runs
/// there, so PROJECT given as a subfolder reads the same root `REVIEW.md` and the same diff
/// (follow-up 6840613078).
fn toplevel(repo: &Path) -> Result<std::path::PathBuf> {
    let out = git(repo, &["rev-parse", "--show-toplevel"])?;
    let top = String::from_utf8_lossy(&out).trim().to_string();
    if top.is_empty() {
        return Err("not inside a git work tree".into());
    }
    Ok(std::path::PathBuf::from(top))
}

fn commit_of(repo: &Path, rev: &str) -> Result<String> {
    if rev.is_empty() || rev.starts_with('-') {
        return Err(format!("invalid revision {rev:?}"));
    }
    let out = git(
        repo,
        &[
            "rev-parse",
            "--verify",
            "--end-of-options",
            &format!("{rev}^{{commit}}"),
        ],
    )
    .map_err(|_| format!("unknown revision {rev:?}"))?;
    Ok(String::from_utf8_lossy(&out).trim().to_string())
}

/// What a path is in `commit`: `Some(Ok(text))` for a regular UTF-8 file, `Some(Err(reason))`
/// for anything else that exists there, `None` when it is absent.
fn blob_text(
    repo: &Path,
    commit: &str,
    path: &str,
) -> Result<Option<std::result::Result<String, &'static str>>> {
    let listing = git(repo, &["ls-tree", "-z", commit, "--", path])?;
    let Some(entry) = listing.split(|b| *b == 0).find(|e| !e.is_empty()) else {
        return Ok(None);
    };
    let entry = String::from_utf8_lossy(entry);
    let (meta, name) = entry.split_once('\t').ok_or("unreadable tree entry")?;
    // `ls-tree -- <path>` also answers for `.` and other pathspecs; only the exact entry counts.
    if name != path {
        return Ok(None);
    }
    let fields: Vec<&str> = meta.split(' ').collect();
    if fields.len() != 3 || fields[1] != "blob" || !(fields[0] == "100644" || fields[0] == "100755")
    {
        return Ok(Some(Err("not-a-regular-file")));
    }
    let bytes = git(repo, &["cat-file", "blob", fields[2]])?;
    Ok(Some(String::from_utf8(bytes).map_err(|_| "not-utf8")))
}

fn file(path: &str, role: &str, included_by: Option<&str>, text: String) -> File {
    File {
        path: path.to_string(),
        role: role.to_string(),
        included_by: included_by.map(str::to_string),
        bytes: text.len(),
        sha256: crate::digest(text.as_bytes()),
        text,
    }
}

/// Markdown links counted toward [`LINK_LIMIT`]: inline links `[text](destination)` and
/// reference definitions `[label]: destination` (up to three spaces of indentation, also inside
/// block quotes and list items, destination on the same or the next line). Spaces,
/// tabs and one line break may come before a destination and a destination may be written in
/// `<…>`, as CommonMark allows (AC-4 record 6841710561). `http:`, `https:`, `mailto:` and `#`
/// destinations are not counted.
fn counted_links(text: &str) -> usize {
    let mut count = 0;
    let mut rest = text;
    while let Some(at) = rest.find("](") {
        let after = &rest[at + 2..];
        if counted_destination(destination(after)) {
            count += 1;
        }
        rest = after;
    }
    // Definitions: the destination is read from the text itself, not from the line, so one
    // that sits on the next line counts (audit FAIL 6841945971 B1).
    let mut offset = 0;
    for line in text.split_inclusive('\n') {
        let start = offset;
        offset += line.len();
        let body = strip_containers(line);
        let indent = body.len() - body.trim_start_matches(' ').len();
        let Some(label) = body[indent..].strip_prefix('[').filter(|_| indent <= 3) else {
            continue;
        };
        let Some(close) = label.find("]:") else {
            continue;
        };
        let after = start + (line.len() - label.len()) + close + 2;
        if !label[..close].contains(']') && counted_destination(destination(&text[after..])) {
            count += 1;
        }
    }
    count
}

/// A line without its block-quote markers (`>`) and list-item markers (`-`, `*`, `+`, `1.`,
/// `1)`), each with up to three spaces before it: a definition inside them still counts.
fn strip_containers(line: &str) -> &str {
    let mut rest = line;
    loop {
        let indent = rest.len() - rest.trim_start_matches(' ').len();
        if indent > 3 {
            return rest;
        }
        let body = &rest[indent..];
        if let Some(r) = body.strip_prefix('>') {
            rest = r.strip_prefix(' ').unwrap_or(r);
            continue;
        }
        let digits = body.len() - body.trim_start_matches(|c: char| c.is_ascii_digit()).len();
        let marker = if body.starts_with(['-', '*', '+']) {
            1
        } else if (1..=9).contains(&digits) && body[digits..].starts_with(['.', ')']) {
            digits + 1
        } else {
            0
        };
        match body[marker..].chars().next() {
            Some(' ' | '\t') if marker > 0 => rest = &body[marker + 1..],
            _ => return rest,
        }
    }
}

/// The destination that starts `after`: leading spaces and tabs and at most one line break
/// skipped, then a `<…>` destination unwrapped or a bare one up to `)` or whitespace.
fn destination(after: &str) -> &str {
    let mut rest = after;
    let mut line_breaks = 0;
    loop {
        if let Some(r) = rest.strip_prefix([' ', '\t', '\r']) {
            rest = r;
        } else if let Some(r) = rest.strip_prefix('\n').filter(|_| line_breaks == 0) {
            line_breaks += 1;
            rest = r;
        } else {
            break;
        }
    }
    if let Some(inner) = rest.strip_prefix('<') {
        return inner.split(['>', '\n']).next().unwrap_or("").trim();
    }
    rest.split([')', ' ', '\t', '\n']).next().unwrap_or("")
}

fn counted_destination(destination: &str) -> bool {
    let lower = destination.to_ascii_lowercase();
    !destination.is_empty()
        && !lower.starts_with("http:")
        && !lower.starts_with("https:")
        && !lower.starts_with("mailto:")
        && !destination.starts_with('#')
}

/// `@name` lines of the root file, in order, without repeats.
fn include_lines(text: &str) -> Vec<String> {
    let mut seen = BTreeSet::new();
    let mut out = Vec::new();
    for line in text.lines() {
        let line = line.strip_suffix('\r').unwrap_or(line);
        if let Some(name) = line.strip_prefix('@') {
            if !name.is_empty()
                && !name.contains(char::is_whitespace)
                && seen.insert(name.to_string())
            {
                out.push(name.to_string());
            }
        }
    }
    out
}

fn skipped_dir(dir: &str) -> bool {
    dir.split('/').any(|part| {
        let part = part.to_ascii_lowercase();
        SKIPPED_DIRS.contains(&part.as_str())
    })
}

/// Every proper ancestor directory of the changed paths, shallow first, sorted.
fn ancestor_dirs(changed: &[Changed]) -> Vec<String> {
    let mut dirs = BTreeSet::new();
    for c in changed {
        let mut path = c.path.as_str();
        while let Some((parent, _)) = path.rsplit_once('/') {
            dirs.insert(parent.to_string());
            path = parent;
        }
    }
    let mut dirs: Vec<String> = dirs.into_iter().collect();
    dirs.sort_by(|a, b| {
        a.matches('/')
            .count()
            .cmp(&b.matches('/').count())
            .then(a.cmp(b))
    });
    dirs
}

pub fn package(repo: &Path, base: &str, head: &str) -> Result<Package> {
    let top = toplevel(repo)?;
    let repo = top.as_path();
    let base = commit_of(repo, base)?;
    let head = commit_of(repo, head)?;
    let merge_base = String::from_utf8_lossy(&git(repo, &["merge-base", &base, &head])?)
        .trim()
        .to_string();
    if merge_base.is_empty() {
        return Err("base and head share no history".into());
    }
    // Attributes come from an empty tree, never from the worktree, whose `.gitattributes` may
    // be uncommitted or belong to another branch (B1). `.git/info/attributes` is read whatever
    // the attribute source, so a repository that has one is refused.
    let info =
        String::from_utf8_lossy(&git(repo, &["rev-parse", "--git-path", "info/attributes"])?)
            .trim()
            .to_string();
    let info = if Path::new(&info).is_absolute() {
        Path::new(&info).to_path_buf()
    } else {
        repo.join(&info)
    };
    if std::fs::read(&info).is_ok_and(|b| b.iter().any(|c| !c.is_ascii_whitespace())) {
        return Err(format!(
            "{} sets attributes that would change the package; remove it or build the package in another clone",
            info.display()
        ));
    }
    let empty_tree =
        String::from_utf8_lossy(&git(repo, &["hash-object", "-t", "tree", "--stdin"])?)
            .trim()
            .to_string();
    let attr_source = format!("--attr-source={empty_tree}");
    // Every diff setting a repository's own `.git/config` could change is fixed here, on the
    // command line, which wins over that file (audit FAIL 6839909483 B1).
    let fixed = [
        "--no-ext-diff",
        "--full-index",
        "--inter-hunk-context=0",
        "--indent-heuristic",
        "--no-relative",
        "--ignore-submodules=none",
        "--submodule=short",
        "--no-textconv",
        "--no-color",
        "--no-renames",
        "--src-prefix=a/",
        "--dst-prefix=b/",
        "--diff-algorithm=myers",
        "-U3",
    ];
    let order_none = format!("-O{}", null_device());
    let mut args = diff_prefix(&attr_source);
    args.push(&order_none);
    args.extend(["--name-status", "-z"]);
    args.extend(fixed);
    args.extend([merge_base.as_str(), head.as_str()]);
    let listing = git(repo, &args)?;
    let fields: Vec<String> = listing
        .split(|b| *b == 0)
        .filter(|f| !f.is_empty())
        .map(|f| String::from_utf8_lossy(f).to_string())
        .collect();
    let changed: Vec<Changed> = fields
        .chunks(2)
        .filter(|pair| pair.len() == 2)
        .map(|pair| Changed {
            status: pair[0].clone(),
            path: pair[1].clone(),
        })
        .collect();
    let mut args = diff_prefix(&attr_source);
    args.push(&order_none);
    args.extend(fixed);
    args.extend([merge_base.as_str(), head.as_str()]);
    let diff = String::from_utf8_lossy(&git(repo, &args)?).to_string();

    let mut review = Vec::new();
    let mut skipped = Vec::new();
    // Root file and its includes.
    match blob_text(repo, &base, REVIEW_FILE)? {
        None => {}
        Some(Err(reason)) => skipped.push(Skipped {
            path: REVIEW_FILE.into(),
            reason: reason.into(),
        }),
        Some(Ok(root)) => {
            let mut files = vec![file(REVIEW_FILE, "root", None, root.clone())];
            let mut used = root.len();
            // Different include files taken so far (the root file is not one of them).
            let mut includes = 0;
            if used > BYTE_LIMIT {
                skipped.push(Skipped {
                    path: REVIEW_FILE.into(),
                    reason: "byte-limit".into(),
                });
                files.clear();
            }
            let root_read = !files.is_empty();
            for name in include_lines(&root).into_iter().filter(|_| root_read) {
                let reason = if name.starts_with(".claude/skills/") {
                    Some("skill-include-not-supported")
                } else if name.contains('/') || name.contains('\\') || name == "." || name == ".." {
                    Some("include-not-beside-root")
                } else if includes >= INCLUDE_LIMIT {
                    Some("include-count")
                } else {
                    None
                };
                if let Some(reason) = reason {
                    skipped.push(Skipped {
                        path: name,
                        reason: reason.into(),
                    });
                    continue;
                }
                // A git failure on one include line skips that include, not the package (B2).
                let found = blob_text(repo, &base, &name).unwrap_or(Some(Err("unreadable")));
                match found {
                    None => skipped.push(Skipped {
                        path: name,
                        reason: "missing".into(),
                    }),
                    Some(Err(reason)) => skipped.push(Skipped {
                        path: name,
                        reason: reason.into(),
                    }),
                    Some(Ok(text)) => {
                        includes += 1;
                        if used + text.len() > BYTE_LIMIT {
                            skipped.push(Skipped {
                                path: name,
                                reason: "byte-limit".into(),
                            });
                        } else {
                            used += text.len();
                            files.push(file(&name, "include", Some(REVIEW_FILE), text));
                        }
                    }
                }
            }
            let links: usize = files.iter().map(|f| counted_links(&f.text)).sum();
            if links > LINK_LIMIT {
                skipped.push(Skipped {
                    path: REVIEW_FILE.into(),
                    reason: "too-many-links".into(),
                });
            } else {
                review.extend(files);
            }
        }
    }
    // Subdirectory files on the changed paths, sharing their own byte limit.
    let dirs = ancestor_dirs(&changed);
    let mut used = 0;
    for dir in &dirs {
        let path = format!("{dir}/{REVIEW_FILE}");
        match blob_text(repo, &base, &path)? {
            None => {}
            Some(_) if skipped_dir(dir) => skipped.push(Skipped {
                path,
                reason: "skipped-directory-name".into(),
            }),
            Some(Err(reason)) => skipped.push(Skipped {
                path,
                reason: reason.into(),
            }),
            Some(Ok(text)) => {
                if used + text.len() > BYTE_LIMIT {
                    skipped.push(Skipped {
                        path,
                        reason: "byte-limit".into(),
                    });
                } else {
                    used += text.len();
                    review.push(file(&path, "subdirectory", None, text));
                }
            }
        }
    }
    // Context: the root and every ancestor directory, each file within the byte limit.
    let mut context = Vec::new();
    for dir in std::iter::once(String::new()).chain(dirs) {
        for name in CONTEXT_FILES {
            let path = if dir.is_empty() {
                name.to_string()
            } else {
                format!("{dir}/{name}")
            };
            match blob_text(repo, &base, &path)? {
                None => {}
                Some(Err(reason)) => skipped.push(Skipped {
                    path,
                    reason: reason.into(),
                }),
                Some(Ok(text)) if text.len() > BYTE_LIMIT => skipped.push(Skipped {
                    path,
                    reason: "byte-limit".into(),
                }),
                Some(Ok(text)) => context.push(file(&path, name, None, text)),
            }
        }
    }
    let (checklist, checklist_skips) = read_checklist(repo, &base, &changed, &diff)?;
    skipped.extend(checklist_skips);
    let mut package = Package {
        schema_version: 1,
        diff_sha256: crate::digest(diff.as_bytes()),
        base,
        head,
        merge_base,
        protection: "not_checked".into(),
        changed,
        diff,
        review,
        context,
        checklist,
        skipped,
        package_digest: String::new(),
    };
    let body = serde_json::to_vec(&package).map_err(|_| "encode failed".to_string())?;
    package.package_digest = crate::digest(&body);
    Ok(package)
}

/// One finding accepted by [`check`].
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Finding {
    pub severity: String,
    pub file: String,
    pub line: u64,
    pub title: String,
    pub evidence: String,
    pub method: String,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct Counts {
    pub important: usize,
    pub nit: usize,
    pub pre_existing: usize,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct CheckReport {
    pub schema_version: u32,
    pub head: String,
    /// `FAIL` when any finding is `important` (design D3), else `PASS`.
    pub verdict: String,
    pub counts: Counts,
    pub findings: Vec<Finding>,
    /// The post body: a `Severity | File:Line | Issue` table and a machine line.
    pub body: String,
    pub digest: String,
}

/// How far (in lines) quoted evidence may sit from the cited line.
pub const EVIDENCE_WINDOW: u64 = 3;
/// The marker of the post body's machine-readable line.
pub const MACHINE_MARKER: &str = "moai-review-severity";
const SEVERITIES: [&str; 3] = ["important", "nit", "pre_existing"];

fn squash(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// Text for one table cell: one line, `|` escaped, and `&`, `<`, `>` as entities so a title or
/// path can neither split the table nor open an HTML comment that forges the machine line
/// (pre-audit 6840720403 B1, N1).
fn table_cell(text: &str) -> String {
    squash(text)
        .replace('&', "&amp;")
        .replace('<', "&lt;")
        .replace('>', "&gt;")
        .replace('|', "\\|")
}

fn text_field<'a>(v: &'a serde_json::Value, key: &str) -> Option<&'a str> {
    v.get(key)?.as_str().filter(|s| !s.trim().is_empty())
}

/// `review-check` (#2614 AC-3): every finding needs a known severity, a `file` that is a
/// regular file in HEAD, a `line` inside it, a title, evidence quoted from the code within
/// [`EVIDENCE_WINDOW`] lines of that line (whitespace-insensitive), and a verification with a
/// method and the result `confirmed`. Any malformed finding refuses the whole set, naming each
/// one by index and reason, so nothing unverified is posted. Reads git only.
pub fn check(repo: &Path, head: &str, input: &serde_json::Value) -> Result<CheckReport> {
    let top = toplevel(repo)?;
    let repo = top.as_path();
    if input.get("schema_version").and_then(|v| v.as_u64()) != Some(1) {
        return Err("unsupported schema_version; supported: 1".into());
    }
    let list = input
        .get("findings")
        .and_then(|v| v.as_array())
        .ok_or("findings must be an array")?;
    let head = commit_of(repo, head)?;
    let mut problems = Vec::new();
    let mut accepted: Vec<Finding> = Vec::new();
    for (i, f) in list.iter().enumerate() {
        let severity = f.get("severity").and_then(|v| v.as_str()).unwrap_or("");
        if !SEVERITIES.contains(&severity) {
            problems.push(format!("finding {i}: severity"));
            continue;
        }
        // The machine line's marker is reserved; a title naming it could pass for a second
        // machine line in the posted body (pre-audit 6840720403 B1).
        let Some(title) = text_field(f, "title").filter(|t| !t.contains(MACHINE_MARKER)) else {
            problems.push(format!("finding {i}: title"));
            continue;
        };
        let Some(evidence) = text_field(f, "evidence") else {
            problems.push(format!("finding {i}: evidence"));
            continue;
        };
        let verification = f.get("verification");
        let method = verification.and_then(|v| text_field(v, "method"));
        let result = verification
            .and_then(|v| v.get("result"))
            .and_then(|v| v.as_str());
        let Some(method) = method.filter(|_| result == Some("confirmed")) else {
            problems.push(format!("finding {i}: not-verified"));
            continue;
        };
        let file = f.get("file").and_then(|v| v.as_str()).unwrap_or("");
        let plain = !file.is_empty()
            && !file.starts_with('/')
            && !file.contains('\\')
            && file
                .split('/')
                .all(|p| !p.is_empty() && p != "." && p != "..");
        let text = match plain.then(|| blob_text(repo, &head, file)) {
            Some(Ok(Some(Ok(text)))) => text,
            _ => {
                problems.push(format!("finding {i}: file-not-in-head"));
                continue;
            }
        };
        let lines: Vec<&str> = text.lines().collect();
        let line = f.get("line").and_then(|v| v.as_u64()).unwrap_or(0);
        if line == 0 || line > lines.len() as u64 {
            problems.push(format!("finding {i}: line-out-of-range"));
            continue;
        }
        let from = line.saturating_sub(EVIDENCE_WINDOW).max(1) as usize - 1;
        let to = (line + EVIDENCE_WINDOW).min(lines.len() as u64) as usize;
        if !squash(&lines[from..to].join("\n")).contains(&squash(evidence)) {
            problems.push(format!("finding {i}: evidence-not-at-line"));
            continue;
        }
        let finding = Finding {
            severity: severity.into(),
            file: file.into(),
            line,
            title: title.into(),
            evidence: evidence.into(),
            method: method.into(),
        };
        if accepted
            .iter()
            .any(|a| a.file == finding.file && a.line == finding.line && a.title == finding.title)
        {
            problems.push(format!("finding {i}: duplicate"));
            continue;
        }
        accepted.push(finding);
    }
    if !problems.is_empty() {
        return Err(format!("review findings refused: {}", problems.join("; ")));
    }
    let rank = |s: &str| {
        SEVERITIES
            .iter()
            .position(|x| *x == s)
            .unwrap_or(SEVERITIES.len())
    };
    accepted.sort_by(|a, b| {
        rank(&a.severity)
            .cmp(&rank(&b.severity))
            .then(a.file.cmp(&b.file))
            .then(a.line.cmp(&b.line))
            .then(a.title.cmp(&b.title))
    });
    let count = |s: &str| accepted.iter().filter(|f| f.severity == s).count();
    let counts = Counts {
        important: count("important"),
        nit: count("nit"),
        pre_existing: count("pre_existing"),
    };
    let mut body = String::from("| Severity | File:Line | Issue |\n| --- | --- | --- |\n");
    for f in &accepted {
        let label = match f.severity.as_str() {
            "important" => "Important",
            "nit" => "Nit",
            _ => "Pre-existing",
        };
        body.push_str(&format!(
            "| {label} | {}:{} | {} |\n",
            table_cell(&f.file),
            f.line,
            table_cell(&f.title)
        ));
    }
    body.push_str(&format!(
        "\n<!-- {MACHINE_MARKER}: {{\"important\":{},\"nit\":{},\"pre_existing\":{}}} -->\n",
        counts.important, counts.nit, counts.pre_existing
    ));
    let mut report = CheckReport {
        schema_version: 1,
        head,
        verdict: if counts.important > 0 { "FAIL" } else { "PASS" }.into(),
        counts,
        findings: accepted,
        body,
        digest: String::new(),
    };
    let bytes = serde_json::to_vec(&report).map_err(|_| "encode failed".to_string())?;
    report.digest = crate::digest(&bytes);
    Ok(report)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn links_count_only_relative_targets() {
        assert_eq!(
            counted_links("[a](x.md) [b](https://e) [c](HTTP://e) [d](mailto:m) [e](#f)"),
            1
        );
        assert_eq!(counted_links("no links ] ( here"), 0);
    }

    /// AC-4 record 6841710561 (found by Codex CLI): CommonMark allows spaces, tabs and one
    /// line break before an inline link's destination, a destination in `<…>`, and reference
    /// definitions `[label]: destination`. All of them count; the exempt schemes stay exempt.
    #[test]
    fn links_with_space_angle_or_reference_destinations_count() {
        assert_eq!(counted_links("[a]( x.md)"), 1);
        assert_eq!(counted_links("[a](\tx.md)"), 1);
        assert_eq!(counted_links("[a](\n  x.md)"), 1);
        assert_eq!(counted_links("[a](<x y.md>)"), 1);
        assert_eq!(counted_links("[a]( https://e)"), 0);
        assert_eq!(counted_links("[a](< mailto:m >)"), 0);
        assert_eq!(counted_links("[a]( #f)"), 0);
        assert_eq!(
            counted_links("[ref]: x.md\n   [two]: <y.md>\n[web]: https://e\n"),
            2
        );
        // Four spaces of indentation is a code block, not a definition.
        assert_eq!(counted_links("    [ref]: x.md\n"), 0);
        // Two line breaks end the link: nothing is counted.
        assert_eq!(counted_links("[a](\n\nx.md)"), 0);
    }

    /// Audit FAIL 6841945971 B1: a definition's destination may sit on the next line, and a
    /// definition inside a block quote or a list item is a definition too.
    #[test]
    fn definitions_on_the_next_line_or_in_containers_count() {
        assert_eq!(counted_links("[ref]:\nx.md\n"), 1);
        assert_eq!(counted_links("[ref]:\n   x.md\n"), 1);
        assert_eq!(counted_links("[ref]:\n\nx.md\n"), 0);
        assert_eq!(counted_links("[ref]:\tx.md\n"), 1);
        assert_eq!(counted_links("[ref]: x.md \"t\"\n"), 1);
        assert_eq!(counted_links("[ref]: x.md\r\n"), 1);
        assert_eq!(counted_links("> [ref]: x.md\n"), 1);
        assert_eq!(counted_links(">> [ref]: x.md\n"), 1);
        assert_eq!(counted_links("- [ref]: x.md\n"), 1);
        assert_eq!(counted_links("12. [ref]: x.md\n"), 1);
        assert_eq!(counted_links("> - [ref]: x.md\n"), 1);
        assert_eq!(counted_links("> [web]: https://e\n"), 0);
    }

    #[test]
    fn include_lines_are_whole_lines_without_spaces_or_repeats() {
        assert_eq!(
            include_lines("@a.md\n @b.md\n@c d.md\n@a.md\r\n@e.md\r\ntext @f.md\n@\n"),
            ["a.md", "e.md"]
        );
    }

    #[test]
    fn a_source_names_an_issue_a_pr_or_a_spec() {
        assert!(names_source("Issue #2401"));
        assert!(names_source("see PR #7"));
        assert!(names_source("SPEC-20261120-review-checklist-by-diff"));
        assert!(names_source("SPEC-20261120-x1"));
        for no in [
            "",
            "C-12",
            "Issue #",
            "Issue # 5",
            "issue #5",
            "SPEC-2026112-x",
            "SPEC-20261120-",
            "SPEC-20261120-Abc",
        ] {
            assert!(!names_source(no), "{no:?}");
        }
    }

    #[test]
    fn a_date_is_year_month_day_in_ascii() {
        assert!(valid_date("2026-11-20"));
        assert!(valid_date("2026-01-31"));
        for no in [
            "2026-13-01",
            "2026-00-10",
            "2026-11-32",
            "2026-11-00",
            "2026-1-1",
            "2026/11/20",
            "２０２６-11-20",
            "2026-11-20 ",
            "",
        ] {
            assert!(!valid_date(no), "{no:?}");
        }
    }

    #[test]
    fn diff_sections_pair_one_to_one_with_the_changed_paths() {
        let diff = "diff --git a/a b/a\n+++ b/a\n@@ -1 +1 @@\n-x\n+y\ndiff --git a/b b/b\nbinary\n";
        let sections = diff_sections(diff, 2).unwrap();
        assert_eq!(sections.len(), 2);
        assert!(sections[0].starts_with("diff --git a/a"));
        assert!(sections[1].starts_with("diff --git a/b"));
        assert!(diff_sections(diff, 1).is_none());
        assert!(diff_sections(diff, 3).is_none());
        assert_eq!(diff_sections("", 0), Some(Vec::new()));
        // A content line that merely contains the words is not a header.
        let inner = "diff --git a/a b/a\n@@ -0,0 +1 @@\n+diff --git a/x b/x\n";
        assert_eq!(diff_sections(inner, 1).unwrap().len(), 1);
    }

    #[test]
    fn only_lines_after_a_hunk_header_that_start_with_plus_are_added() {
        let section = "diff --git a/a b/a\n--- a/a\n+++ b/a\n@@ -1,3 +1,3 @@\n ctx\n-gone\n+new\n\\ No newline at end of file\n@@ -9 +9 @@\n+++ also added\n";
        assert_eq!(added_lines(section), ["new", "++ also added"]);
        assert!(added_lines("diff --git a/a b/a\nnew file mode 100644\n").is_empty());
    }

    fn item(id: &str, prefix: &str, patterns: &[&str]) -> Item {
        Item {
            id: id.into(),
            scope: vec![(prefix.into(), Vec::new())],
            patterns: patterns.iter().map(|p| p.to_string()).collect(),
            check: "c".into(),
            source: "Issue #1".into(),
            added: "2026-11-20".into(),
        }
    }

    fn changed(paths: &[&str]) -> Vec<Changed> {
        paths
            .iter()
            .map(|p| Changed {
                path: p.to_string(),
                status: "M".into(),
            })
            .collect()
    }

    #[test]
    fn items_fire_by_path_and_by_added_pattern_in_that_path() {
        let ch = changed(&["a/x", "b/y"]);
        let diff = "diff --git a/a/x b/a/x\n@@ -1 +1 @@\n+needle\ndiff --git a/b/y b/b/y\n@@ -1 +1 @@\n+other\n";
        let sections = diff_sections(diff, 2).unwrap();
        let items = [
            item("path", "a/", &[]),
            item("hit", "", &["zzz", "needle"]),
            item("wrong-path", "b/", &["needle"]),
            item("none", "c/", &[]),
        ];
        let got = judge_checklist(&items, &ch, Some(&sections));
        let ids: Vec<&str> = got.iter().map(|g| g.id.as_str()).collect();
        assert_eq!(ids, ["path", "hit"]);
        assert_eq!(got[0].matched.paths, ["a/x"]);
        assert_eq!(got[0].matched.pattern, None);
        assert_eq!(got[1].matched.paths, ["a/x"]);
        assert_eq!(got[1].matched.pattern.as_deref(), Some("needle"));
        // Without sections only the items that need no pattern are judged.
        let got = judge_checklist(&items, &ch, None);
        let ids: Vec<&str> = got.iter().map(|g| g.id.as_str()).collect();
        assert_eq!(ids, ["path"]);
    }

    #[test]
    fn a_table_is_refused_for_its_first_problem() {
        let ok = r#"{"schema_version":1,"items":[{"id":"a","trigger":{"paths":[{"prefix":""}]},"check":"c","source":"PR #3","added":"2026-11-20"}]}"#;
        assert_eq!(parse_checklist(ok).unwrap().len(), 1);
        let bad = ok.replace("\"source\":\"PR #3\",", "");
        assert_eq!(parse_checklist(&bad).unwrap_err(), "item 0: missing `source`");
        assert!(parse_checklist("[]").is_err());
        assert!(parse_checklist(r#"{"schema_version":1}"#).is_err());
    }

    #[test]
    fn skipped_directory_names_match_whole_parts_in_any_case() {
        assert!(skipped_dir("src/Tests"));
        assert!(skipped_dir("contrib/x"));
        assert!(!skipped_dir("src/testing"));
        assert!(!skipped_dir("latest"));
    }
}
