//! `ac-commands-check`: the recorded-AC-command lint (#2352), ported from
//! `tools/check_ac_commands.py` (tools slice 2 — same scan, same judgement, same wording).
//!
//! A recorded `cargo test --test <name> [filters…]` line must name an existing Rust test
//! target and select at least one test — cargo exits 0 with "running 0 tests" for a filter
//! that matches nothing, which has silently passed as verification more than once (survey
//! 2.11). Reads only: the one cargo listing per test target is a per-invocation cache,
//! never written.

use std::collections::{BTreeMap, BTreeSet};
use std::path::{Path, PathBuf};
use std::process::Command;

/// Scan recorded cargo-test commands, or judge one file
///
/// Flags every recorded `cargo test --test <name> [filters…]` command that names a target
/// with no `tests/<name>.rs`, a target whose listing has no tests, or filters that select
/// none. `--tree` scans `docs/**/*.md`, `.moai-cli/reports/*.md` and the issue-body records
/// numbered 2153 or later (frozen records stay out — a historical record is not edited to
/// match a renamed test); `--file` judges the one file against the caller's working folder.
/// The result carries the tool's summary line and one violation line per finding, in the
/// wording the Python tool used. A clean tree succeeds; any violation fails.
#[derive(clap::Args)]
#[command(group(clap::ArgGroup::new("source").required(true).args(["tree", "file"])))]
pub struct Args {
    /// The repository tree to scan.
    #[arg(long)]
    tree: Option<PathBuf>,
    /// One recorded file to judge (its targets are judged against the caller's cwd).
    #[arg(long)]
    file: Option<PathBuf>,
}

pub const OPERATION: &str = "ac_commands.check";

pub fn run(args: Args) -> Result<serde_json::Value, String> {
    let (root, files) = match (args.tree, args.file) {
        (Some(tree), None) => (tree.clone(), collect_tree_files(&tree)),
        (None, Some(file)) => (
            std::env::current_dir().unwrap_or_else(|_| PathBuf::from(".")),
            vec![file],
        ),
        // clap's exclusive required group answers the usage case with its own exit.
        _ => unreachable!("clap requires exactly one of --tree and --file"),
    };
    let mut cache: BTreeMap<String, BTreeSet<String>> = BTreeMap::new();
    let mut violations = Vec::new();
    for path in &files {
        match std::fs::read_to_string(path) {
            Ok(text) => violations.extend(violations_in(path, &text, &root, &mut cache)),
            Err(e) => violations.push(format!("{}: unreadable ({e})", path.display())),
        }
    }
    let summary = format!(
        "checked {} recorded files, {} --test targets listed",
        files.len(),
        cache.len()
    );
    if violations.is_empty() {
        return Ok(serde_json::json!({
            "summary": summary,
            "checked": files.len(),
            "targets_listed": cache.len(),
        }));
    }
    // The failure path of the envelope ops (the lane-handoff precedent): the lines the
    // caller reads go to stderr in the tool's order — summary first, one line per
    // violation — and the exit is 1 through `Err`.
    eprintln!("{summary}");
    for v in &violations {
        eprintln!("VIOLATION: {v}");
    }
    Err(format!("{} violation(s)", violations.len()))
}

/// The tree's recorded files, in the Python tool's order: `docs/**/*.md` sorted, then
/// `.moai-cli/reports/*.md` sorted, then the issue-body records numbered 2153 or later
/// sorted — the groups are each sorted, never mixed (a violation line names its file in
/// the order the tool judged them). The frozen RC1 mirrors (< 2153) stay out.
fn collect_tree_files(root: &Path) -> Vec<PathBuf> {
    let mut docs = Vec::new();
    let docs_root = root.join("docs");
    let mut stack = vec![docs_root.clone()];
    while let Some(dir) = stack.pop() {
        let Ok(entries) = std::fs::read_dir(&dir) else {
            continue;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                stack.push(path);
            } else if path.extension().is_some_and(|e| e == "md") {
                docs.push(path);
            }
        }
    }
    docs.sort();
    let reports = root.join(moai_cli::fixed_state!("reports"));
    let mut listed = Vec::new();
    if let Ok(entries) = std::fs::read_dir(&reports) {
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_file() && path.extension().is_some_and(|e| e == "md") {
                listed.push(path);
            }
        }
    }
    listed.sort();
    let mut bodies = Vec::new();
    if let Ok(entries) = std::fs::read_dir(reports.join("issue-bodies")) {
        for entry in entries.flatten() {
            let path = entry.path();
            if !path.is_file() || path.extension().is_some_and(|e| e != "md") {
                continue;
            }
            let Some(stem) = path.file_stem().and_then(|s| s.to_str()) else {
                continue;
            };
            // `(\d+)-` at the start of the file name; a record without the number is out.
            let Some(dash) = stem.find('-') else {
                continue;
            };
            if let Ok(n) = stem[..dash].parse::<u64>() {
                if n >= 2153 {
                    bodies.push(path);
                }
            }
        }
    }
    bodies.sort();
    docs.extend(listed);
    docs.extend(bodies);
    docs
}

/// Python's `str.splitlines()` — the boundaries Python breaks lines on are wider than
/// Rust's `\n`: `\v`, `\f`, the file-separator group, NEL and the Unicode separators
/// count too (the audit's c14 — a `\x0b` inside a recorded line is a line break to the
/// tool, so a span never crosses it).
fn py_lines(text: &str) -> Vec<&str> {
    let mut out = Vec::new();
    let mut start = 0usize;
    let mut idx = 0usize;
    let chars: Vec<(usize, char)> = text.char_indices().collect();
    while idx < chars.len() {
        let (pos, ch) = chars[idx];
        let mut skip = 1usize;
        let boundary = match ch {
            '\n' => true,
            '\r' => {
                if chars.get(idx + 1).is_some_and(|(_, n)| *n == '\n') {
                    skip = 2;
                }
                true
            }
            '\u{0B}' | '\u{0C}' | '\u{1C}' | '\u{1D}' | '\u{1E}' | '\u{85}' | '\u{2028}'
            | '\u{2029}' => true,
            _ => false,
        };
        if boundary {
            out.push(&text[start..pos]);
            start = pos + ch.len_utf8() + if skip == 2 { 1 } else { 0 };
        }
        idx += skip;
    }
    out.push(&text[start..]);
    out
}

/// Candidate command strings: fenced-block lines and backtick code spans, as the tool had
/// them (table rows that only mention commands are handled by this: only the quoted span is
/// tokenized, trailing prose never becomes a filter).
fn candidates(text: &str) -> Vec<&str> {
    let mut out = Vec::new();
    let mut in_fence = false;
    for line in py_lines(text) {
        if line.trim_start().starts_with("```") {
            in_fence = !in_fence;
            continue;
        }
        if in_fence {
            out.push(line);
            continue;
        }
        // `` `([^`\n]+)` `` per line, scanning the way the regex does: from each
        // backtick, a non-empty run up to the next backtick is a span and the scan moves
        // past the closer; an empty pair — or a backtick with no closer — moves the scan
        // one character, so the second backtick of a doubled pair opens the span inside
        // (``b`` yields `b`), and an unclosed one is no span at all (the audit's B3).
        let mut scan = 0usize;
        while let Some(open_rel) = line[scan..].find('`') {
            let open = scan + open_rel;
            let Some(close_rel) = line[open + 1..].find('`') else {
                break;
            };
            let close = open + 1 + close_rel;
            if close > open + 1 {
                out.push(&line[open + 1..close]);
                scan = close + 1;
            } else {
                scan = open + 1;
            }
        }
    }
    out
}

/// Shell-like tokenization good enough for a recorded cargo command: repeated `K=V ` env
/// prefixes are stripped, then one pair of wrapping backticks.
fn tokens_of(candidate: &str) -> Vec<String> {
    let mut line = candidate.trim();
    loop {
        let stripped = strip_env_prefix(line);
        if stripped == line {
            break;
        }
        line = stripped;
    }
    if line.contains('`') {
        line = line.trim_matches('`');
    }
    line.split_whitespace().map(str::to_string).collect()
}

/// The regex `^(?:[A-Z_][A-Z0-9_]*=\S+\s+)+` one repetition at a time.
fn strip_env_prefix(line: &str) -> &str {
    let mut rest = line;
    loop {
        let Some((head, tail)) = rest.split_once(char::is_whitespace) else {
            let tail = "";
            let _ = tail;
            return rest;
        };
        if !is_env_word(head) {
            return rest;
        }
        let next = tail.trim_start();
        if next.is_empty() {
            return next;
        }
        rest = next;
    }
}

/// `[A-Z_][A-Z0-9_]*=\S+`: uppercase-and-underscore key with a value that holds no space.
fn is_env_word(word: &str) -> bool {
    let Some((key, value)) = word.split_once('=') else {
        return false;
    };
    let mut chars = key.chars();
    let first = match chars.next() {
        Some(c) if c == '_' || c.is_ascii_uppercase() => c,
        _ => return false,
    };
    let _ = first;
    if !chars.all(|c| c == '_' || c.is_ascii_uppercase() || c.is_ascii_digit()) {
        return false;
    }
    !value.is_empty() && !value.chars().any(char::is_whitespace)
}

/// Positional filter words after `--test <name>`, up to the next flag token.
fn filters_after_target(tokens: &[String]) -> Vec<String> {
    let Some(idx) = tokens.iter().position(|t| t == "--test") else {
        return Vec::new();
    };
    let mut out = Vec::new();
    for token in &tokens[idx + 2..] {
        if token.starts_with('-') {
            break;
        }
        out.push(token.clone());
    }
    out
}

/// `cargo\s+test\b` anywhere in the candidate — the tool's regex searches the whole
/// string, so an env-prefixed command (`CARGO_HOME=… cargo test …`) matches too; the
/// `\b` means a `testing` word after `test` does not.
fn contains_cargo_test(candidate: &str) -> bool {
    let mut rest = candidate;
    while let Some(i) = rest.find("cargo") {
        let after = &rest[i + "cargo".len()..];
        let trimmed = after.trim_start();
        if trimmed.len() < after.len()
            && trimmed.starts_with("test")
            && !trimmed["test".len()..]
                .chars()
                .next()
                .is_some_and(|c| c.is_ascii_alphanumeric() || c == '_')
        {
            return true;
        }
        rest = after;
    }
    false
}

/// The first `--test <name>` the tool's regex finds — the flag needs `--test\s+`, so
/// `--test-threads` is never a `--test` and never names a target (the audit's B3), and
/// when the word after it is not a target name at all (`./x`) the regex keeps scanning
/// and the next `--test` names the target (the audit's e1).
fn test_target_name(tokens: &[String]) -> Option<String> {
    let joined = tokens.join(" ");
    let mut scan = 0usize;
    while scan <= joined.len() {
        let rel = joined[scan..].find("--test")?;
        let idx = scan + rel;
        let after = &joined[idx + "--test".len()..];
        if !after.starts_with(char::is_whitespace) {
            scan = idx + "--test".len();
            continue;
        }
        let rest = after.trim_start();
        let end = rest
            .find(|c: char| !(c.is_ascii_alphanumeric() || c == '_' || c == '-'))
            .unwrap_or(rest.len());
        let name = &rest[..end];
        if name.is_empty() {
            scan = idx + "--test".len();
            continue;
        }
        return Some(name.to_string());
    }
    None
}

/// cargo's own filter rule: a test is selected when ANY positional filter is a substring of
/// the full test path; no filters means "select everything".
fn selects_any(filters: &[String], tests: &BTreeSet<String>) -> bool {
    if filters.is_empty() {
        return true;
    }
    filters
        .iter()
        .any(|f| tests.iter().any(|t| t.contains(f.as_str())))
}

/// The tool clips the recorded command at 170 characters after trimming it.
fn clip(candidate: &str) -> String {
    let trimmed = candidate.trim();
    trimmed.chars().take(170).collect()
}

fn violations_in(
    path: &Path,
    text: &str,
    root: &Path,
    cache: &mut BTreeMap<String, BTreeSet<String>>,
) -> Vec<String> {
    let html;
    let text = if path.extension().is_some_and(|e| e == "html") {
        html = strip_tags(text);
        html.as_str()
    } else {
        text
    };
    let mut found = Vec::new();
    for candidate in candidates(text) {
        if !contains_cargo_test(candidate) || !candidate.contains("--test") {
            continue;
        }
        let tokens = tokens_of(candidate);
        let Some(name) = test_target_name(&tokens) else {
            continue;
        };
        if !root.join("tests").join(format!("{name}.rs")).is_file() {
            found.push(format!(
                "{}: target tests/{name}.rs does not exist :: {}",
                path.display(),
                clip(candidate)
            ));
            continue;
        }
        let filters: Vec<String> = filters_after_target(&tokens)
            .into_iter()
            .filter(|f| !f.starts_with('<'))
            .collect();
        let tests = list_tests(root, &name, cache);
        if tests.is_empty() {
            found.push(format!(
                "{}: target {name} lists no tests :: {}",
                path.display(),
                clip(candidate)
            ));
            continue;
        }
        if !selects_any(&filters, &tests) {
            found.push(format!(
                "{}: filters [{}] select 0 tests in {name} ({} available) :: {}",
                path.display(),
                filters
                    .iter()
                    .map(|f| format!("'{f}'"))
                    .collect::<Vec<_>>()
                    .join(", "),
                tests.len(),
                clip(candidate)
            ));
        }
    }
    found
}

/// One listing per target for the whole invocation — `cargo test --test <name> -- --list`,
/// the `path: test` lines becoming the set of full test paths.
fn list_tests(
    root: &Path,
    name: &str,
    cache: &mut BTreeMap<String, BTreeSet<String>>,
) -> BTreeSet<String> {
    if let Some(tests) = cache.get(name) {
        return tests.clone();
    }
    let output = Command::new("cargo")
        .args([
            "test",
            "--offline",
            "--locked",
            "--test",
            name,
            "--",
            "--list",
        ])
        .current_dir(root)
        .output();
    let mut tests = BTreeSet::new();
    if let Ok(output) = output {
        for line in String::from_utf8_lossy(&output.stdout).lines() {
            let line = line.trim();
            if line.contains(": test") {
                if let Some((path, _)) = line.split_once(':') {
                    tests.insert(path.trim().to_string());
                }
            }
        }
    }
    cache.insert(name.to_string(), tests.clone());
    tests
}

/// `re.sub(r"<[^>]+>", " ", text)`: each run from `<` to the next `>` becomes one space.
fn strip_tags(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    let mut rest = text;
    while let Some(open) = rest.find('<') {
        let Some(close_rel) = rest[open..].find('>') else {
            break;
        };
        out.push_str(&rest[..open]);
        out.push(' ');
        rest = &rest[open + close_rel + 1..];
    }
    out.push_str(rest);
    out
}
