//! SPEC-20261120-security-scan-notify: the scanner of the `security-scan` duty. It reads the
//! text a write tool just wrote (never the file's content on disk) line by line and reports
//! which of six classes a line holds; the engine turns that into a notification, a
//! `would_notify` line or nothing.
//!
//! The shapes are hand-written byte matchers that follow the Go guardian's regexes
//! (`internal/hook/security/patterns.go`, read 2026-11-20) with three deliberate differences,
//! listed in the SPEC: one line at a time, a look-ahead of [`LOOK_AHEAD`] bytes beyond the fixed
//! lead-in of a shape, and no match for an `.innerHTML ==` comparison. Because every shape sees
//! at most [`LOOK_AHEAD`] bytes past its lead-in, the time is proportional to the text length.
//! A shape never decides a refusal: a miss costs one advisory, a false hit costs one sentence.

use serde_json::Value;
use std::path::{Component, Path};

/// How far a shape looks past the end of its fixed lead-in (SPEC REQ-4).
pub(super) const LOOK_AHEAD: usize = 512;
/// Text longer than this is not scanned (SPEC REQ-5).
pub(super) const MAX_TEXT_BYTES: usize = 1024 * 1024;
/// The most findings one notification lists, and its byte bound (SPEC REQ-7).
const LISTED_MAX: usize = 5;
const NOTE_MAX_BYTES: usize = 800;
/// Room kept for the "and N more" line while findings are still to be listed.
const OVERFLOW_RESERVE: usize = 40;
/// The opening words of every notification.
const HEADER: &str = "MoAI security-scan (advisory, nothing was blocked): the text you just wrote holds patterns worth a look. Check them before you go on.";

/// The six classes of the table, in table order.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub(super) enum Class {
    UnsafeDeserialization,
    DomInjectionXss,
    HardcodedSecret,
    CodeInjectionEval,
    SqlInjection,
    CommandInjection,
}

pub(super) const CLASSES: [Class; 6] = [
    Class::UnsafeDeserialization,
    Class::DomInjectionXss,
    Class::HardcodedSecret,
    Class::CodeInjectionEval,
    Class::SqlInjection,
    Class::CommandInjection,
];

impl Class {
    pub(super) fn name(self) -> &'static str {
        match self {
            Class::UnsafeDeserialization => "unsafe-deserialization",
            Class::DomInjectionXss => "dom-injection-xss",
            Class::HardcodedSecret => "hardcoded-secret",
            Class::CodeInjectionEval => "code-injection-eval",
            Class::SqlInjection => "sql-injection",
            Class::CommandInjection => "command-injection",
        }
    }

    pub(super) fn severity(self) -> &'static str {
        match self {
            Class::HardcodedSecret => "critical",
            _ => "high",
        }
    }

    fn found_in(self, line: &Line) -> bool {
        match self {
            Class::UnsafeDeserialization => unsafe_deserialization(line),
            Class::DomInjectionXss => dom_injection(line),
            Class::HardcodedSecret => hardcoded_secret(line),
            Class::CodeInjectionEval => code_injection(line),
            Class::SqlInjection => sql_injection(line),
            Class::CommandInjection => command_injection(line),
        }
    }
}

/// One class found on one line (1-based) of a written text.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(super) struct Finding {
    pub class: Class,
    pub line: usize,
}

/// A line and its ASCII-lowercased copy (same byte positions).
struct Line<'a> {
    text: &'a str,
    lower: String,
}

/// Every class found in `text`, by line and then in table order, each class at most once per
/// line.
pub(super) fn scan(text: &str) -> Vec<Finding> {
    let mut found = Vec::new();
    for (index, raw) in text.split('\n').enumerate() {
        let line = Line {
            text: raw.strip_suffix('\r').unwrap_or(raw),
            lower: raw.to_ascii_lowercase(),
        };
        for class in CLASSES {
            if class.found_in(&line) {
                found.push(Finding {
                    class,
                    line: index + 1,
                });
            }
        }
    }
    found
}

fn is_word(b: u8) -> bool {
    b.is_ascii_alphanumeric() || b == b'_'
}

fn is_blank(b: u8) -> bool {
    b == b' ' || b == b'\t'
}

fn is_quote(b: u8) -> bool {
    matches!(b, b'"' | b'\'' | b'`')
}

/// The at most [`LOOK_AHEAD`] bytes of `bytes` from `from` on: everything a shape may look at
/// beyond its lead-in.
fn window(bytes: &[u8], from: usize) -> &[u8] {
    let rest = bytes.get(from..).unwrap_or(&[]);
    &rest[..rest.len().min(LOOK_AHEAD)]
}

/// Index of the first byte at or after `from` that is not a blank (the length when none is).
fn skip_blanks(w: &[u8], from: usize) -> usize {
    let from = from.min(w.len());
    from + w[from..].iter().take_while(|b| is_blank(**b)).count()
}

/// Whether `w`, after optional blanks, continues with `lit`.
fn blanks_then(w: &[u8], lit: &[u8]) -> bool {
    w[skip_blanks(w, 0)..].starts_with(lit)
}

/// `(start, end)` of every occurrence of `anchor` in `hay`.
fn anchors<'a>(hay: &'a str, anchor: &'a str) -> impl Iterator<Item = (usize, usize)> + 'a {
    hay.match_indices(anchor)
        .map(move |(at, _)| (at, at + anchor.len()))
}

/// A word boundary before the byte at `at` (ASCII letters, digits and `_` are word bytes).
fn bounded_before(bytes: &[u8], at: usize) -> bool {
    at == 0 || !is_word(bytes[at - 1])
}

/// `name`, optional blanks, `(`.
fn call_of(text: &str, name: &str) -> bool {
    anchors(text, name).any(|(_, end)| blanks_then(window(text.as_bytes(), end), b"("))
}

/// `new`, blanks, `word`, optional blanks, `(`.
fn new_of(text: &str, word: &str) -> bool {
    anchors(text, "new").any(|(_, end)| {
        let w = window(text.as_bytes(), end);
        let at = skip_blanks(w, 0);
        at > 0 && w[at..].starts_with(word.as_bytes()) && blanks_then(&w[at + word.len()..], b"(")
    })
}

fn unsafe_deserialization(line: &Line) -> bool {
    let text = line.text;
    if ["yaml.load", "torch.load", "Marshal.load"]
        .iter()
        .any(|name| call_of(text, name))
    {
        return true;
    }
    // `pickle.load` or `pickle.loads`.
    if anchors(text, "pickle.load").any(|(_, end)| {
        let w = window(text.as_bytes(), end);
        blanks_then(w.strip_prefix(b"s").unwrap_or(w), b"(")
    }) {
        return true;
    }
    if new_of(text, "ObjectInputStream") {
        return true;
    }
    anchors(text, "unserialize").any(|(at, end)| {
        bounded_before(text.as_bytes(), at) && blanks_then(window(text.as_bytes(), end), b"(")
    })
}

fn dom_injection(line: &Line) -> bool {
    let text = line.text;
    if text.contains("dangerouslySetInnerHTML") || call_of(text, "document.write") {
        return true;
    }
    // `v-html`, optional blanks, `=`.
    if anchors(text, "v-html").any(|(_, end)| blanks_then(window(text.as_bytes(), end), b"=")) {
        return true;
    }
    // `.innerHTML`, optional blanks, a single `=` (not `==` or `===`).
    anchors(text, ".innerHTML").any(|(_, end)| {
        let w = window(text.as_bytes(), end);
        let at = skip_blanks(w, 0);
        w.get(at) == Some(&b'=') && w.get(at + 1) != Some(&b'=')
    })
}

/// The window after a name (`password`, `api_key`): optional blanks, `=` or `:=` or `<-`,
/// optional blanks, a quote (`"` or `'`), at least one byte that is no quote and a quote.
fn assigns_quoted_value(w: &[u8]) -> bool {
    let rest = &w[skip_blanks(w, 0)..];
    let op = if rest.starts_with(b":=") || rest.starts_with(b"<-") {
        2
    } else if rest.starts_with(b"=") {
        1
    } else {
        return false;
    };
    let rest = &rest[op..];
    let rest = &rest[skip_blanks(rest, 0)..];
    if !matches!(rest.first(), Some(b'"' | b'\'')) {
        return false;
    }
    matches!(
        rest[1..].iter().position(|b| matches!(b, b'"' | b'\'')),
        Some(n) if n >= 1
    )
}

fn hardcoded_secret(line: &Line) -> bool {
    let lower = &line.lower;
    let bytes = lower.as_bytes();
    if anchors(lower, "password").any(|(_, end)| assigns_quoted_value(window(bytes, end))) {
        return true;
    }
    // `api`, an optional `_` or `-`, `key`: the name ends after `key`.
    anchors(lower, "api").any(|(_, end)| {
        let key = end + usize::from(matches!(bytes.get(end), Some(b'_' | b'-')));
        bytes
            .get(key..)
            .is_some_and(|rest| rest.starts_with(b"key"))
            && assigns_quoted_value(window(bytes, key + "key".len()))
    })
}

fn code_injection(line: &Line) -> bool {
    let text = line.text;
    anchors(text, "eval").any(|(at, end)| {
        bounded_before(text.as_bytes(), at) && blanks_then(window(text.as_bytes(), end), b"(")
    }) || new_of(text, "Function")
}

fn sql_injection(line: &Line) -> bool {
    let bytes = line.lower.as_bytes();
    // A quote, a statement keyword, the rest of the string without a quote, the closing quote,
    // optional blanks and `+`: `"SELECT … " + x`.
    for (at, b) in bytes.iter().enumerate() {
        if !is_quote(*b) {
            continue;
        }
        let w = window(bytes, at + 1);
        let start = skip_blanks(w, 0);
        for keyword in ["select", "insert", "update", "delete", "drop"] {
            if !w[start..].starts_with(keyword.as_bytes()) {
                continue;
            }
            let end = start + keyword.len();
            if w.get(end).is_some_and(|b| is_word(*b)) {
                continue;
            }
            // The tail is looked at within the look-ahead bound beyond the keyword.
            let tail = window(bytes, at + 1 + end);
            if let Some(close) = tail.iter().position(|b| is_quote(*b)) {
                if blanks_then(&tail[close + 1..], b"+") {
                    return true;
                }
            }
        }
    }
    // `execute` or `query`, optional blanks, `(`, then a `+` before the first `)`.
    ["execute", "query"].iter().any(|name| {
        anchors(&line.lower, name).any(|(_, end)| {
            let w = window(bytes, end);
            let at = skip_blanks(w, 0);
            if w.get(at) != Some(&b'(') {
                return false;
            }
            let after = window(bytes, end + at + 1);
            matches!(
                after.iter().position(|b| *b == b')'),
                Some(close) if after[..close].contains(&b'+')
            )
        })
    })
}

fn command_injection(line: &Line) -> bool {
    let text = line.text;
    let bytes = text.as_bytes();
    // `shell`, optional blanks, `=`, optional blanks, `True`.
    let shell = anchors(text, "shell").any(|(_, end)| {
        let w = window(bytes, end);
        let at = skip_blanks(w, 0);
        if w.get(at) != Some(&b'=') {
            return false;
        }
        let rest = &w[at + 1..];
        rest[skip_blanks(rest, 0)..].starts_with(b"True")
    });
    if shell || call_of(text, "os.system") || call_of(text, "child_process.exec") {
        return true;
    }
    // A quote, `sh` or `bash`, blanks and `-c` that ends a word: `"sh -c …"`.
    bytes.iter().enumerate().any(|(at, b)| {
        if !is_quote(*b) {
            return false;
        }
        let w = window(bytes, at + 1);
        ["sh", "bash"].iter().any(|shell| {
            let Some(rest) = w.strip_prefix(shell.as_bytes()) else {
                return false;
            };
            let blanks = skip_blanks(rest, 0);
            blanks > 0
                && rest[blanks..].starts_with(b"-c")
                && !rest.get(blanks + 2).is_some_and(|b| is_word(*b))
        })
    })
}

/// The path of a written file as the notification names it: relative to the project, no `./`,
/// or `None` when the file lies outside the project (a `..`, another root). An absolute path
/// that does not start with `cwd` is resolved through the file system once (a link such as
/// `/var` for `/private/var`) and judged again; the file's content is never read.
fn project_relative(raw: &str, cwd: &Path) -> Option<String> {
    let path = Path::new(raw);
    let resolved;
    let rel = if path.is_absolute() {
        match path.strip_prefix(cwd) {
            Ok(rel) => rel,
            Err(_) => {
                resolved = std::fs::canonicalize(path).ok()?;
                resolved.strip_prefix(cwd).ok()?
            }
        }
    } else {
        path
    };
    let mut parts: Vec<&str> = Vec::new();
    for component in rel.components() {
        match component {
            Component::Normal(part) => parts.push(part.to_str()?),
            Component::CurDir => {}
            _ => return None,
        }
    }
    if parts.is_empty() {
        return None;
    }
    let joined = parts.join("/");
    // A name the model reads must not be able to forge a line: control characters and absurd
    // lengths are replaced by a fixed word.
    if joined.len() > 256 || joined.chars().any(char::is_control) {
        return Some("(unnamed)".to_string());
    }
    Some(joined)
}

/// The text written to one file of a call.
struct Written {
    path: String,
    text: String,
    lines: usize,
}

impl Written {
    fn new(path: &str) -> Written {
        Written {
            path: path.trim().to_string(),
            text: String::new(),
            lines: 0,
        }
    }

    fn push_line(&mut self, line: &str) {
        if self.lines > 0 {
            self.text.push('\n');
        }
        self.text.push_str(line);
        self.lines += 1;
    }
}

/// The files a write call writes text to, each with the text it wrote: the Claude Code tools
/// `Write`, `Edit`, `MultiEdit` and `NotebookEdit` (the written pieces joined by a newline), and
/// Codex `apply_patch` (the added lines of each file of the patch; context and removed lines are
/// not written text).
fn written_texts(tool: &str, tool_input: &Value) -> Vec<Written> {
    let mut out = Vec::new();
    match tool {
        "Write" | "Edit" | "MultiEdit" | "NotebookEdit" => {
            let path = tool_input["file_path"]
                .as_str()
                .or_else(|| tool_input["notebook_path"].as_str());
            let mut pieces: Vec<&str> = Vec::new();
            for key in ["content", "new_string", "new_source"] {
                pieces.extend(tool_input[key].as_str());
            }
            for edit in tool_input["edits"].as_array().into_iter().flatten() {
                pieces.extend(edit["new_string"].as_str());
            }
            if let (Some(path), false) = (path, pieces.is_empty()) {
                let mut written = Written::new(path);
                for piece in pieces {
                    written.push_line(piece);
                }
                out.push(written);
            }
        }
        "apply_patch" => {
            let mut current: Option<usize> = None;
            for line in tool_input["command"].as_str().unwrap_or("").lines() {
                let named = ["*** Add File: ", "*** Update File: ", "*** Move to: "]
                    .iter()
                    .find_map(|head| line.strip_prefix(head));
                if let Some(path) = named {
                    out.push(Written::new(path));
                    current = Some(out.len() - 1);
                } else if line.starts_with("*** Delete File: ") {
                    current = None;
                } else if let (Some(added), Some(at)) = (line.strip_prefix('+'), current) {
                    out[at].push_line(added);
                }
            }
            out.retain(|written| written.lines > 0);
        }
        _ => {}
    }
    out
}

/// What one call's written text holds: `(project-relative path, finding)` for every file that
/// lies inside the project, holds no NUL character and is at most [`MAX_TEXT_BYTES`] long.
pub(super) fn scan_call(tool: &str, tool_input: &Value, cwd: &Path) -> Vec<(String, Finding)> {
    let mut all = Vec::new();
    for written in written_texts(tool, tool_input) {
        if written.text.len() > MAX_TEXT_BYTES || written.text.contains('\0') {
            continue;
        }
        let Some(path) = project_relative(&written.path, cwd) else {
            continue;
        };
        for finding in scan(&written.text) {
            all.push((path.clone(), finding));
        }
    }
    all
}

/// The notification: fixed words, at most [`LISTED_MAX`] findings within [`NOTE_MAX_BYTES`], the
/// rest counted. It never repeats the matched text.
pub(super) fn notification(found: &[(String, Finding)]) -> String {
    let mut text = HEADER.to_string();
    let mut listed = 0;
    for (index, (path, finding)) in found.iter().enumerate().take(LISTED_MAX) {
        let entry = format!(
            "\n- {} ({}) {}:{}",
            finding.class.name(),
            finding.class.severity(),
            path,
            finding.line
        );
        let reserve = if index + 1 < found.len() {
            OVERFLOW_RESERVE
        } else {
            0
        };
        if text.len() + entry.len() + reserve > NOTE_MAX_BYTES {
            break;
        }
        text.push_str(&entry);
        listed += 1;
    }
    if listed < found.len() {
        text.push_str(&format!(
            "\n- and {} more findings not listed.",
            found.len() - listed
        ));
    }
    text
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    fn classes(text: &str) -> Vec<(Class, usize)> {
        scan(text).iter().map(|f| (f.class, f.line)).collect()
    }

    /// The table of the SPEC: each class at its edge, one byte past its edge and in prose.
    #[test]
    fn each_class_is_found_at_its_edge_and_not_beyond_it() {
        use Class::*;
        let name = format!("{}{}", "pass", "word");
        let cases: Vec<(String, Option<Class>)> = vec![
            ("pickle.loads(b)".into(), Some(UnsafeDeserialization)),
            ("pickle.loadx(b)".into(), None),
            ("a.innerHTML=b".into(), Some(DomInjectionXss)),
            ("a.innerHTML==b".into(), None),
            (format!("{name}='x'"), Some(HardcodedSecret)),
            (format!("{name}=''"), None),
            ("x.eval(y)".into(), Some(CodeInjectionEval)),
            ("my_eval(y)".into(), None),
            ("'select a' +b".into(), Some(SqlInjection)),
            ("'selection' +b".into(), None),
            ("os.system(c)".into(), Some(CommandInjection)),
            ("os.systems(c)".into(), None),
        ];
        for (text, want) in cases {
            let got: Vec<Class> = scan(&text).iter().map(|f| f.class).collect();
            assert_eq!(got, want.into_iter().collect::<Vec<_>>(), "{text:?}");
        }
    }

    /// The look-ahead is the 512 bytes after the lead-in: the blanks before `(` count.
    #[test]
    fn the_look_ahead_ends_512_bytes_after_the_lead_in() {
        let at = |blanks: usize| format!("eval{}(x)", " ".repeat(blanks));
        assert_eq!(classes(&at(LOOK_AHEAD - 1)).len(), 1);
        assert!(classes(&at(LOOK_AHEAD)).is_empty());
    }

    #[test]
    fn a_class_counts_once_per_line_and_lines_count_from_one() {
        use Class::*;
        assert_eq!(
            classes("ok\r\neval(a); eval(b)\neval(os.system(c))"),
            vec![
                (CodeInjectionEval, 2),
                (CodeInjectionEval, 3),
                (CommandInjection, 3)
            ]
        );
    }

    #[test]
    fn paths_are_project_relative_or_none_and_never_forge_a_line() {
        let cwd = Path::new("/p/proj");
        for (raw, want) in [
            ("/p/proj/src/a.rs", Some("src/a.rs")),
            ("./src/a.rs", Some("src/a.rs")),
            ("src/a.rs", Some("src/a.rs")),
            ("../a.rs", None),
            ("src/../a.rs", None),
            ("/p/other/a.rs", None),
            ("/p/proj", None),
            ("", None),
            ("src/a\n- fake", Some("(unnamed)")),
        ] {
            assert_eq!(project_relative(raw, cwd).as_deref(), want, "{raw:?}");
        }
        let long = format!("src/{}", "d".repeat(260));
        assert_eq!(project_relative(&long, cwd).as_deref(), Some("(unnamed)"));
    }

    #[test]
    fn a_patch_scans_added_lines_of_each_file_and_skips_deleted_files() {
        let patch = "*** Begin Patch\n*** Add File: a.py\n+x\n+eval(1)\n\
                     *** Delete File: b.py\n+eval(2)\n*** Update File: c.py\n-eval(3)\n\
                     eval(4)\n+ok\n*** Move to: d.py\n+os.system(1)\n*** End Patch";
        let found = scan_call("apply_patch", &json!({ "command": patch }), Path::new("/p"));
        let got: Vec<(String, Class, usize)> = found
            .into_iter()
            .map(|(path, f)| (path, f.class, f.line))
            .collect();
        assert_eq!(
            got,
            vec![
                ("a.py".to_string(), Class::CodeInjectionEval, 2),
                ("d.py".to_string(), Class::CommandInjection, 1),
            ]
        );
    }

    #[test]
    fn the_notification_holds_five_findings_and_800_bytes_and_counts_the_rest() {
        let many = |n: usize, path: &str| -> Vec<(String, Finding)> {
            (1..=n)
                .map(|line| {
                    (
                        path.to_string(),
                        Finding {
                            class: Class::HardcodedSecret,
                            line,
                        },
                    )
                })
                .collect()
        };
        let five = notification(&many(5, "a.py"));
        assert_eq!(five.lines().filter(|l| l.starts_with("- ")).count(), 5);
        assert!(!five.contains("more findings"), "{five}");
        let seven = notification(&many(7, "a.py"));
        assert!(
            seven.contains("- and 2 more findings not listed."),
            "{seven}"
        );
        let long = notification(&many(7, &"d".repeat(250)));
        assert!(long.len() <= NOTE_MAX_BYTES, "{} bytes", long.len());
        assert!(long.starts_with(HEADER));
    }
}
