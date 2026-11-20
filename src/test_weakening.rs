//! Weakened-test detection over a unified diff (SPEC-20261120-test-weakening-detection).
//!
//! A change that only touches tests is tier 0 by its paths, and tier 0 has no audit, so a test
//! that was deleted, skipped, thinned out, slowed down with a sleep or "fixed" by rewriting its
//! golden file passed unreviewed. [`detect`] reads the text of `git diff` (merge base to head)
//! and returns one [`Finding`] for each of those five kinds, with the file and line. It is
//! deterministic and line based: no model, no network, no repository access. What it finds
//! goes to the auditor; whether it also raises the audit tier is the gate's switch
//! (`gate::TEST_WEAKENING_ENFORCED`).

use std::collections::BTreeSet;

/// One weakened-test hit: the kind, the file, the line and a short detail.
#[derive(Clone, Debug, PartialEq, Eq, serde::Serialize)]
pub struct Finding {
    /// `test_deleted`, `skip_added`, `assertions_reduced`, `sleep_added` or `golden_only`.
    pub kind: &'static str,
    pub path: String,
    /// A line of the old file for a deletion, of the new file for an addition.
    pub line: u64,
    pub detail: String,
}

impl Finding {
    /// `kind path:line detail`, the form the gate result and the review package list.
    pub fn render(&self) -> String {
        format!("{} {}:{} {}", self.kind, self.path, self.line, self.detail)
    }
}

/// One file of the diff with its removed and added lines (`(line, text)`; text without the
/// leading `-` or `+`).
#[derive(Default)]
struct File {
    path: String,
    is_new: bool,
    is_deleted: bool,
    removed: Vec<(u64, String)>,
    added: Vec<(u64, String)>,
}

fn unquote(p: &str) -> &str {
    p.strip_prefix("a/")
        .or_else(|| p.strip_prefix("b/"))
        .unwrap_or(p)
}

/// The `-a,b +c,d` of a hunk header as `(old_start, old_len, new_start, new_len)`.
fn hunk(line: &str) -> Option<(u64, u64, u64, u64)> {
    let rest = line.strip_prefix("@@ -")?;
    let (old, rest) = rest.split_once(" +")?;
    let (new, _) = rest.split_once(" @@")?;
    let range = |s: &str| -> Option<(u64, u64)> {
        match s.split_once(',') {
            Some((a, b)) => Some((a.parse().ok()?, b.parse().ok()?)),
            None => Some((s.parse().ok()?, 1)),
        }
    };
    let (o, n) = (range(old)?, range(new)?);
    Some((o.0, o.1, n.0, n.1))
}

fn parse(diff: &str) -> Vec<File> {
    let mut files: Vec<File> = Vec::new();
    let (mut old_left, mut new_left) = (0u64, 0u64);
    let (mut old_at, mut new_at) = (0u64, 0u64);
    for line in diff.lines() {
        if old_left > 0 || new_left > 0 {
            // Inside a hunk: the line starts with ' ', '-', '+' or '\' (no newline marker).
            let Some(file) = files.last_mut() else {
                continue;
            };
            if let Some(t) = line.strip_prefix('-') {
                file.removed.push((old_at, t.to_string()));
                old_at += 1;
                old_left = old_left.saturating_sub(1);
            } else if let Some(t) = line.strip_prefix('+') {
                file.added.push((new_at, t.to_string()));
                new_at += 1;
                new_left = new_left.saturating_sub(1);
            } else if line.starts_with('\\') {
                // "\ No newline at end of file"
            } else {
                old_at += 1;
                new_at += 1;
                old_left = old_left.saturating_sub(1);
                new_left = new_left.saturating_sub(1);
            }
            continue;
        }
        if let Some(rest) = line.strip_prefix("diff --git ") {
            let path = rest
                .rsplit_once(" b/")
                .map(|(_, b)| b.to_string())
                .unwrap_or_else(|| unquote(rest).to_string());
            files.push(File {
                path,
                ..Default::default()
            });
        } else if line.starts_with("new file mode") {
            if let Some(f) = files.last_mut() {
                f.is_new = true;
            }
        } else if line.starts_with("deleted file mode") {
            if let Some(f) = files.last_mut() {
                f.is_deleted = true;
            }
        } else if let Some(p) = line.strip_prefix("--- ") {
            if let Some(f) = files.last_mut() {
                if p == "/dev/null" {
                    f.is_new = true;
                } else if f.is_deleted {
                    f.path = unquote(p).to_string();
                }
            }
        } else if let Some(p) = line.strip_prefix("+++ ") {
            if let Some(f) = files.last_mut() {
                if p == "/dev/null" {
                    f.is_deleted = true;
                } else {
                    f.path = unquote(p).to_string();
                }
            }
        } else if let Some((os, ol, ns, nl)) = hunk(line) {
            old_at = os;
            new_at = ns;
            old_left = ol;
            new_left = nl;
            // A zero-length side points at the line before the change.
            if ol == 0 {
                old_at += 1;
            }
            if nl == 0 {
                new_at += 1;
            }
        }
    }
    files
}

fn file_name(path: &str) -> &str {
    path.rsplit('/').next().unwrap_or(path)
}

/// A test file by its path: below a `tests` folder, or named `test_*.py`, `*_test.py`,
/// `*_test.rs` or `*_test.go`.
pub fn is_test_path(path: &str) -> bool {
    let name = file_name(path);
    path.split('/').rev().skip(1).any(|c| c == "tests")
        || (name.starts_with("test_") && name.ends_with(".py"))
        || name.ends_with("_test.py")
        || name.ends_with("_test.rs")
        || name.ends_with("_test.go")
}

/// A golden file by its path: a `golden` folder, `.golden`, `.snap` or `.expected`, or below
/// `tests/fixtures/`.
pub fn is_golden_path(path: &str) -> bool {
    path.split('/').rev().skip(1).any(|c| c == "golden")
        || path.ends_with(".golden")
        || path.ends_with(".snap")
        || path.ends_with(".expected")
        || path.starts_with("tests/fixtures/")
        || path.contains("/tests/fixtures/")
}

fn is_doc_path(path: &str) -> bool {
    path.starts_with("docs/") || path.ends_with(".md")
}

fn ext(path: &str) -> &str {
    file_name(path)
        .rsplit_once('.')
        .map(|(_, e)| e)
        .unwrap_or("")
}

fn is_comment(t: &str) -> bool {
    t.starts_with("//") || t.starts_with('*') || t.starts_with("/*")
}

/// The test a line starts: its function name (`?` when the name is on a later line).
fn test_marker(t: &str) -> Option<Option<String>> {
    let ident = |s: &str| -> Option<String> {
        let id: String = s
            .chars()
            .take_while(|c| c.is_alphanumeric() || *c == '_')
            .collect();
        (!id.is_empty()).then_some(id)
    };
    if t.starts_with("#[test]")
        || t.starts_with("#[tokio::test")
        || t.starts_with("#[rstest")
        || t.starts_with("#[async_std::test")
    {
        return Some(None);
    }
    for pre in ["def test_", "async def test_"] {
        if let Some(rest) = t.strip_prefix(pre) {
            return Some(ident(&format!("test_{rest}")));
        }
    }
    if let Some(rest) = t.strip_prefix("func Test") {
        return Some(ident(&format!("Test{rest}")));
    }
    None
}

/// `fn name` at the start of a trimmed line (after `pub`, `async`, `unsafe`).
fn rust_fn_name(t: &str) -> Option<String> {
    let mut rest = t;
    for pre in ["pub(crate) ", "pub ", "async ", "unsafe "] {
        while let Some(r) = rest.strip_prefix(pre) {
            rest = r;
        }
    }
    let rest = rest.strip_prefix("fn ")?;
    let id: String = rest
        .chars()
        .take_while(|c| c.is_alphanumeric() || *c == '_')
        .collect();
    (!id.is_empty()).then_some(id)
}

/// The names of the tests the lines start, in order (`None` when the name is not on a line the
/// attribute is followed by within four lines).
fn tests_in(lines: &[(u64, String)]) -> Vec<(u64, Option<String>)> {
    let mut out = Vec::new();
    for (i, (n, text)) in lines.iter().enumerate() {
        let t = text.trim();
        let Some(named) = test_marker(t) else {
            continue;
        };
        let name = named.or_else(|| {
            lines[i + 1..]
                .iter()
                .take(4)
                .take_while(|(m, _)| *m <= n + 4)
                .find_map(|(_, next)| rust_fn_name(next.trim()))
        });
        out.push((*n, name));
    }
    out
}

fn is_assertion(t: &str) -> bool {
    t.starts_with("assert!(")
        || t.starts_with("assert_eq!(")
        || t.starts_with("assert_ne!(")
        || t.starts_with("assert ")
        || t.starts_with("assert(")
        || t.contains("self.assert")
        || t.starts_with("expect(")
}

fn is_skip(path: &str, t: &str) -> bool {
    match ext(path) {
        "rs" => t.starts_with("#[ignore"),
        "py" => t.starts_with("@unittest.skip") || t.starts_with("@pytest.mark.skip"),
        "go" => !is_comment(t) && t.contains("t.Skip"),
        "js" | "jsx" | "ts" | "tsx" | "mjs" | "cjs" => !is_comment(t) && t.contains(".skip("),
        _ => false,
    }
}

fn is_sleep(t: &str) -> bool {
    !is_comment(t)
        && !t.starts_with('#')
        && (t.contains("thread::sleep") || t.contains("time.sleep("))
}

/// The weakened-test findings of a unified diff (`git diff <merge base> <head>`), sorted by
/// path, line and kind. An empty list means the diff weakens no test in the five ways.
pub fn detect(diff: &str) -> Vec<Finding> {
    let files = parse(diff);
    let mut out = Vec::new();

    // Tests that are started again somewhere in the diff were moved, not deleted.
    let mut readded: BTreeSet<String> = BTreeSet::new();
    for f in &files {
        for (_, name) in tests_in(&f.added) {
            if let Some(n) = name {
                readded.insert(n);
            }
        }
    }

    for f in &files {
        // 1. A test the diff deletes or disables by dropping its attribute.
        let started_again = tests_in(&f.added).len();
        let mut unnamed_left = started_again;
        for (line, name) in tests_in(&f.removed) {
            match name {
                Some(n) if readded.contains(&n) => {}
                Some(n) => out.push(Finding {
                    kind: "test_deleted",
                    path: f.path.clone(),
                    line,
                    detail: format!("test `{n}` is deleted"),
                }),
                None if unnamed_left > 0 => unnamed_left -= 1,
                None => out.push(Finding {
                    kind: "test_deleted",
                    path: f.path.clone(),
                    line,
                    detail: "a test attribute is removed".into(),
                }),
            }
        }

        // 2. A skip added, beyond the skips the same file loses.
        let skips: Vec<&(u64, String)> = f
            .added
            .iter()
            .filter(|(_, t)| is_skip(&f.path, t.trim()))
            .collect();
        let lost = f
            .removed
            .iter()
            .filter(|(_, t)| is_skip(&f.path, t.trim()))
            .count();
        if skips.len() > lost {
            for (line, text) in skips {
                out.push(Finding {
                    kind: "skip_added",
                    path: f.path.clone(),
                    line: *line,
                    detail: format!("`{}` skips a test", text.trim()),
                });
            }
        }

        if is_test_path(&f.path) {
            // 3. More assertions removed than added in one test file.
            let gone: Vec<&(u64, String)> = f
                .removed
                .iter()
                .filter(|(_, t)| is_assertion(t.trim()))
                .collect();
            let gained = f
                .added
                .iter()
                .filter(|(_, t)| is_assertion(t.trim()))
                .count();
            if gone.len() > gained {
                out.push(Finding {
                    kind: "assertions_reduced",
                    path: f.path.clone(),
                    line: gone[0].0,
                    detail: format!("{} assertions removed, {gained} added", gone.len()),
                });
            }
            // 4. A sleep added to a test.
            for (line, text) in f.added.iter().filter(|(_, t)| is_sleep(t.trim())) {
                out.push(Finding {
                    kind: "sleep_added",
                    path: f.path.clone(),
                    line: *line,
                    detail: format!("`{}` waits in a test", text.trim()),
                });
            }
        }
    }

    // 5. Golden files rewritten and nothing but tests, golden files and documents changed.
    let only_tests_and_goldens = !files.is_empty()
        && files
            .iter()
            .all(|f| is_test_path(&f.path) || is_golden_path(&f.path) || is_doc_path(&f.path));
    if only_tests_and_goldens {
        for f in files
            .iter()
            .filter(|f| is_golden_path(&f.path) && !f.is_new)
        {
            let line = f
                .added
                .first()
                .or(f.removed.first())
                .map(|(n, _)| *n)
                .unwrap_or(1);
            out.push(Finding {
                kind: "golden_only",
                path: f.path.clone(),
                line,
                detail: "a golden file is rewritten and no source file changes".into(),
            });
        }
    }

    out.sort_by(|a, b| (&a.path, a.line, a.kind).cmp(&(&b.path, b.line, b.kind)));
    out
}
