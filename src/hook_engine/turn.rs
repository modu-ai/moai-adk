//! SPEC-20261121-security-scan-turn: what the `security-turn` duty (Stop, Claude Code) reads. A
//! turn's changes are the added lines of the tracked changes against HEAD and the text of every
//! untracked file the project does not ignore. Each file is scanned on its own with the scanner
//! of `security-scan`; for a tracked file the whole working-tree file is scanned (so a shape that
//! needs its neighbours still sees them) and only the findings on added lines are kept.
//!
//! Git is asked twice (`git diff HEAD` and `git ls-files --others`), each call with a three
//! second limit and a bound on what is read; any failure gives no answer at all, never a partial
//! one. Nothing here writes: the project's index is not refreshed (`GIT_OPTIONAL_LOCKS=0`).

use super::scan::{self, Finding};
use std::io::Read;
use std::path::Path;
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};

/// The diff output the duty reads; a longer diff is not read at all.
const DIFF_MAX_BYTES: usize = 4 * 1024 * 1024;
/// The untracked files the duty reads.
const UNTRACKED_MAX_FILES: usize = 200;
/// How long one git call may run.
const GIT_TIMEOUT: Duration = Duration::from_secs(3);
/// Folders that hold the tool's own state, never the project's code.
const OWN_FOLDERS: [&str; 2] = [".moai-cli/", ".git/"];

/// One finding of the turn: the project-relative path, the class and line, and the trimmed text
/// of the line (used for the digest only; it is never printed or logged).
pub(super) struct TurnFinding {
    pub path: String,
    pub finding: Finding,
    pub text: String,
}

/// What a bounded git call printed: at most `cap` bytes, and whether more was cut off.
struct GitRun {
    data: Vec<u8>,
    cut: bool,
}

/// Runs `git -C cwd <args>` with the limits above. `None` when git cannot be started, runs past
/// [`GIT_TIMEOUT`], or exits with a failure while its output was complete.
fn run_git(cwd: &Path, args: &[&str], cap: usize) -> Option<GitRun> {
    let mut cmd = Command::new("git");
    crate::identity::scrub_git_env(&mut cmd);
    let mut child = cmd
        .arg("-C")
        .arg(cwd)
        .args(args)
        .env("GIT_OPTIONAL_LOCKS", "0")
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()
        .ok()?;
    let stdout = child.stdout.take()?;
    // Reading stops one byte past the cap: the child then ends on a closed pipe, and the extra
    // byte tells a cut output from one that is exactly at the cap.
    let reader = std::thread::spawn(move || {
        let mut data = Vec::new();
        let _ = stdout.take(cap as u64 + 1).read_to_end(&mut data);
        data
    });
    let started = Instant::now();
    let status = loop {
        match child.try_wait() {
            Ok(Some(status)) => break Some(status),
            Ok(None) if started.elapsed() >= GIT_TIMEOUT => {
                let _ = child.kill();
                let _ = child.wait();
                break None;
            }
            Ok(None) => std::thread::sleep(Duration::from_millis(5)),
            Err(_) => {
                let _ = child.kill();
                let _ = child.wait();
                break None;
            }
        }
    };
    let mut data = reader.join().ok()?;
    let status = status?;
    let cut = data.len() > cap;
    if cut {
        data.truncate(cap);
    }
    if !cut && !status.success() {
        return None;
    }
    Some(GitRun { data, cut })
}

/// The text of a regular file inside the project that the scanner may read: not a link, at most
/// one mebibyte, no NUL character, valid UTF-8.
fn read_text(path: &Path) -> Option<String> {
    let meta = std::fs::symlink_metadata(path).ok()?;
    if !meta.is_file() || meta.len() > scan::MAX_TEXT_BYTES as u64 {
        return None;
    }
    let mut bytes = Vec::new();
    std::fs::File::open(path)
        .ok()?
        .take(scan::MAX_TEXT_BYTES as u64 + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    if bytes.len() > scan::MAX_TEXT_BYTES || bytes.contains(&0) {
        return None;
    }
    String::from_utf8(bytes).ok()
}

/// `a[,b]` of a hunk header as (start, count), the count defaulting to 1.
fn span(word: &str) -> Option<(usize, usize)> {
    let word = word.trim_start_matches(['-', '+']);
    match word.split_once(',') {
        Some((start, count)) => Some((start.parse().ok()?, count.parse().ok()?)),
        None => Some((word.parse().ok()?, 1)),
    }
}

/// The new-side path of a `+++ ` header: `None` for a deleted file, a name git had to quote and
/// a header without the `b/` prefix.
fn new_path(rest: &str) -> Option<String> {
    let rest = rest.strip_suffix('\t').unwrap_or(rest);
    if rest == "/dev/null" || rest.starts_with('"') {
        return None;
    }
    rest.strip_prefix("b/").map(str::to_string)
}

/// The files of a zero-context diff with the line ranges (start, count) they gained, in the
/// order git printed them. Lines inside a hunk are counted off by the header's lengths, so an
/// added line that itself reads `++ b/x` is never taken for a file header.
fn added_ranges(diff: &[u8]) -> Vec<(String, Vec<(usize, usize)>)> {
    let text = String::from_utf8_lossy(diff);
    let mut files: Vec<(String, Vec<(usize, usize)>)> = Vec::new();
    let mut current: Option<usize> = None;
    let (mut old_left, mut new_left) = (0usize, 0usize);
    for line in text.split('\n') {
        if old_left > 0 || new_left > 0 {
            match line.as_bytes().first() {
                Some(b'+') => new_left = new_left.saturating_sub(1),
                Some(b'-') => old_left = old_left.saturating_sub(1),
                _ => {}
            }
            continue;
        }
        if line.starts_with("diff --git ") {
            current = None;
        } else if let Some(rest) = line.strip_prefix("+++ ") {
            current = new_path(rest).map(|path| {
                files.push((path, Vec::new()));
                files.len() - 1
            });
        } else if let Some(rest) = line.strip_prefix("@@ ") {
            let head = rest.split(" @@").next().unwrap_or("");
            let Some((old, new)) = head.split_once(' ') else {
                continue;
            };
            let (Some((_, old_count)), Some((start, count))) = (span(old), span(new)) else {
                continue;
            };
            old_left = old_count;
            new_left = count;
            if let (Some(at), true) = (current, count > 0) {
                files[at].1.push((start, count));
            }
        }
    }
    files
}

fn own_state(path: &str) -> bool {
    OWN_FOLDERS.iter().any(|folder| path.starts_with(folder))
}

/// The findings the scanner reports for `text`, as the turn's findings of the file `path`;
/// `keep` decides by line number which ones count.
fn findings_of(path: &str, text: &str, keep: impl Fn(usize) -> bool, out: &mut Vec<TurnFinding>) {
    let lines: Vec<&str> = text.split('\n').collect();
    for finding in scan::scan(text) {
        if keep(finding.line) {
            out.push(TurnFinding {
                path: path.to_string(),
                finding,
                text: lines
                    .get(finding.line - 1)
                    .map_or(String::new(), |l| l.trim().to_string()),
            });
        }
    }
}

/// What the turn changed that the scanner flags, tracked files first (git's path order) and then
/// untracked files (git's path order). `None` when the answer cannot be had: no repository, no
/// commit, a failed or slow git call. An empty list is an answer.
pub(super) fn collect(cwd: &Path) -> Option<Vec<TurnFinding>> {
    run_git(cwd, &["rev-parse", "--verify", "-q", "HEAD"], 256)?;
    let diff = run_git(
        cwd,
        &[
            "-c",
            "core.quotePath=false",
            "diff",
            "HEAD",
            "-U0",
            "--no-ext-diff",
            "--no-color",
            "--no-renames",
            "--no-textconv",
            "--relative",
        ],
        DIFF_MAX_BYTES,
    )?;
    let mut untracked = run_git(
        cwd,
        &["ls-files", "--others", "--exclude-standard", "-z"],
        DIFF_MAX_BYTES,
    )?;
    let mut out = Vec::new();
    // A diff longer than the bound is not read at all: a cut hunk is no hunk.
    if !diff.cut {
        for (path, ranges) in added_ranges(&diff.data) {
            if own_state(&path) {
                continue;
            }
            let Some(name) = scan::project_relative(&path, cwd) else {
                continue;
            };
            let Some(text) = read_text(&cwd.join(&path)) else {
                continue;
            };
            findings_of(
                &name,
                &text,
                |line| {
                    ranges
                        .iter()
                        .any(|(start, count)| line >= *start && line < start + count)
                },
                &mut out,
            );
        }
    }
    if untracked.cut {
        // The last entry may be a name cut in two.
        if let Some(at) = untracked.data.iter().rposition(|b| *b == 0) {
            untracked.data.truncate(at + 1);
        } else {
            untracked.data.clear();
        }
    }
    let names = untracked
        .data
        .split(|b| *b == 0)
        .filter(|name| !name.is_empty())
        .filter_map(|name| std::str::from_utf8(name).ok())
        .filter(|name| !own_state(name) && !name.ends_with('/'))
        .take(UNTRACKED_MAX_FILES);
    for path in names {
        let Some(name) = scan::project_relative(path, cwd) else {
            continue;
        };
        let Some(text) = read_text(&cwd.join(path)) else {
            continue;
        };
        findings_of(&name, &text, |_| true, &mut out);
    }
    Some(out)
}

/// A digest (16 hex digits) of a set of findings: path, class and the trimmed text of each line,
/// so a line that only moves does not make the set a new one.
pub(super) fn digest_of(found: &[TurnFinding]) -> String {
    let mut bytes = Vec::new();
    for item in found {
        bytes.extend_from_slice(item.path.as_bytes());
        bytes.push(0);
        bytes.extend_from_slice(item.finding.class.name().as_bytes());
        bytes.push(0);
        bytes.extend_from_slice(item.text.as_bytes());
        bytes.push(b'\n');
    }
    crate::digest(&bytes).trim_start_matches("sha256:")[..16].to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn ranges(diff: &str) -> Vec<(String, Vec<(usize, usize)>)> {
        added_ranges(diff.as_bytes())
    }

    /// An added line that reads like a file header is counted off by the hunk's length and never
    /// starts a new file; a deleted file, a pure removal and a quoted name add no range.
    #[test]
    fn a_hunk_is_counted_off_by_its_header() {
        let diff = "diff --git a/a.py b/a.py\n--- a/a.py\n+++ b/a.py\n@@ -1 +1,2 @@\n-x\n+++ b/evil.py\n+y\n\
                    diff --git a/gone.py b/gone.py\n--- a/gone.py\n+++ /dev/null\n@@ -1 +0,0 @@\n-z\n\
                    diff --git a/q.py b/q.py\n--- a/q.py\n+++ \"b/q\\n.py\"\n@@ -0,0 +1 @@\n+w\n\
                    diff --git a/n.py b/n.py\n--- a/n.py\n+++ b/n.py\t\n@@ -3,2 +5,0 @@\n-p\n-q\n";
        assert_eq!(
            ranges(diff),
            [
                ("a.py".to_string(), vec![(1, 2)]),
                ("n.py".to_string(), vec![]),
            ]
        );
    }

    /// A hunk header without a count means one line.
    #[test]
    fn a_missing_count_is_one_line() {
        let diff = "diff --git a/a.py b/a.py\n--- a/a.py\n+++ b/a.py\n@@ -7 +7 @@\n-a\n+b\n@@ -9,0 +10,3 @@\n+1\n+2\n+3\n";
        assert_eq!(ranges(diff), [("a.py".to_string(), vec![(7, 1), (10, 3)])]);
    }

    /// The digest ignores where a line stands and changes with its text, path or class.
    #[test]
    fn the_digest_follows_text_path_and_class_not_the_line_number() {
        let item = |path: &str, line: usize, text: &str| TurnFinding {
            path: path.to_string(),
            finding: Finding {
                class: scan::Class::CodeInjectionEval,
                line,
            },
            text: text.to_string(),
        };
        let base = digest_of(&[item("a.py", 2, "eval(x)")]);
        assert_eq!(base, digest_of(&[item("a.py", 9, "eval(x)")]));
        assert_ne!(base, digest_of(&[item("b.py", 2, "eval(x)")]));
        assert_ne!(base, digest_of(&[item("a.py", 2, "eval(y)")]));
        assert_eq!(base.len(), 16);
    }
}
