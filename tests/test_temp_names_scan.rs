//! SPEC-20261124-test-temp-names: no test under `tests/` names a scratch folder under the system
//! temp folder from the process id.
//!
//! Such a name is predictable. A test that creates it and does not remove it leaves a folder, and
//! a later test process with the same pid (nextest reuses pids over a long run) meets the leftover
//! (`remote origin already exists`, card C-417). A `tempfile::TempDir` is unique and removes itself.
//!
//! The scan is a text match on code: comments and the inside of string literals are blanked, the
//! rest is cut into statements at `;`, `{` and `}`, and a statement that holds both the temp folder
//! call and the process id call is reported with the line of the pid call. The pid used as data (a
//! key name, a lock owner, a file name inside a `TempDir`) shares no statement with the temp folder
//! call and is not reported. A pid name hidden behind a helper function, a temp folder bound to a
//! variable in an earlier statement, and a bare `id()` after `use std::process::id` are not seen
//! (SPEC-20261124-pid-scan-followup, Non-goals). Every file under `tests/` is read, this one too.

use std::fs;
use std::path::{Path, PathBuf};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");

const TEMP_CALL: &str = "temp_dir()";
const PID_CALL: &str = "process::id()";

/// The code view of Rust source: comments and the inside of string and character literals become
/// spaces, newlines stay, so a position keeps its line.
fn code_view(src: &str) -> String {
    let c: Vec<char> = src.chars().collect();
    let mut out: Vec<char> = Vec::with_capacity(c.len());
    let blank = |ch: char| if ch == '\n' { '\n' } else { ' ' };
    let mut i = 0;
    while i < c.len() {
        let ch = c[i];
        let next = c.get(i + 1).copied().unwrap_or('\0');
        if ch == '/' && next == '/' {
            while i < c.len() && c[i] != '\n' {
                out.push(' ');
                i += 1;
            }
        } else if ch == '/' && next == '*' {
            let mut depth = 0;
            while i < c.len() {
                if c[i] == '/' && c.get(i + 1) == Some(&'*') {
                    depth += 1;
                    out.extend([' ', ' ']);
                    i += 2;
                } else if c[i] == '*' && c.get(i + 1) == Some(&'/') {
                    depth -= 1;
                    out.extend([' ', ' ']);
                    i += 2;
                    if depth == 0 {
                        break;
                    }
                } else {
                    out.push(blank(c[i]));
                    i += 1;
                }
            }
        } else if let Some((open_len, hashes)) = raw_literal_open(&c, i) {
            // The prefix (`r`, `br`, `cr`), the hashes and the opening quote stay as they are.
            out.extend(&c[i..i + open_len]);
            i += open_len;
            while i < c.len() {
                if c[i] == '"' && (0..hashes).all(|k| c.get(i + 1 + k) == Some(&'#')) {
                    out.extend(&c[i..i + 1 + hashes]);
                    i += 1 + hashes;
                    break;
                }
                out.push(blank(c[i]));
                i += 1;
            }
        } else if ch == '"' {
            out.push('"');
            i += 1;
            while i < c.len() {
                if c[i] == '\\' {
                    out.push(' ');
                    out.push(blank(c.get(i + 1).copied().unwrap_or(' ')));
                    i += 2;
                } else if c[i] == '"' {
                    out.push('"');
                    i += 1;
                    break;
                } else {
                    out.push(blank(c[i]));
                    i += 1;
                }
            }
        } else if ch == '\'' {
            // A character literal (`'x'`, `'\n'`, `'"'`); anything else is a lifetime.
            let len = if next == '\\' {
                // `'\n'`, `'\''`, `'\u{1F}'`: the closing quote is the first one after the escape.
                c.get(i + 3..)
                    .and_then(|rest| rest.iter().position(|&x| x == '\''))
                    .map(|p| p + 4)
            } else if c.get(i + 2) == Some(&'\'') {
                Some(3)
            } else {
                None
            };
            match len {
                Some(n) => {
                    out.extend(std::iter::repeat_n(' ', n));
                    i += n;
                }
                None => {
                    out.push(ch);
                    i += 1;
                }
            }
        } else {
            out.push(ch);
            i += 1;
        }
    }
    out.into_iter().collect()
}

/// The number of `#` of a raw string whose `r` is at `i` (`r"`, `r#"`), or None.
fn raw_open(c: &[char], i: usize) -> Option<usize> {
    let mut k = i + 1;
    while c.get(k) == Some(&'#') {
        k += 1;
    }
    (c.get(k) == Some(&'"')).then_some(k - i - 1)
}

/// The length of the opening of a raw literal that starts at `i` (`r"`, `br#"`, `cr"`) and its
/// number of `#`, or None. The prefix must not be the end of a longer name. A raw byte string or a
/// raw C string is raw like `r"…"`: a backslash in it is not an escape.
fn raw_literal_open(c: &[char], i: usize) -> Option<(usize, usize)> {
    if i > 0 && (c[i - 1].is_alphanumeric() || c[i - 1] == '_') {
        return None;
    }
    let r_at = match c[i] {
        'r' => i,
        'b' | 'c' if c.get(i + 1) == Some(&'r') => i + 1,
        _ => return None,
    };
    let hashes = raw_open(c, r_at)?;
    Some((r_at - i + 1 + hashes + 1, hashes))
}

/// The 1-based lines of the pid calls that share a statement with the temp folder call.
fn pid_names_under_temp(src: &str) -> Vec<usize> {
    let view = code_view(src);
    let mut lines = Vec::new();
    let mut start = 0;
    let mut stmt_line = 1;
    let mut line = 1;
    let bytes: Vec<(usize, char)> = view.char_indices().collect();
    let flush = |from: usize, to: usize, first_line: usize, lines: &mut Vec<usize>| {
        let text = &view[from..to];
        if text.contains(TEMP_CALL) {
            let mut at = 0;
            while let Some(p) = text[at..].find(PID_CALL) {
                let abs = at + p;
                lines.push(first_line + text[..abs].matches('\n').count());
                at = abs + PID_CALL.len();
            }
        }
    };
    for (idx, ch) in bytes {
        if ch == ';' || ch == '{' || ch == '}' {
            flush(start, idx, stmt_line, &mut lines);
            start = idx + 1;
            stmt_line = line;
        }
        if ch == '\n' {
            line += 1;
        }
    }
    flush(start, view.len(), stmt_line, &mut lines);
    lines
}

fn rust_files(dir: &Path, out: &mut Vec<PathBuf>) {
    let mut entries: Vec<_> = fs::read_dir(dir)
        .unwrap_or_else(|e| panic!("{}: {e}", dir.display()))
        .map(|e| e.unwrap())
        .collect();
    entries.sort_by_key(|e| e.file_name());
    for e in entries {
        let kind = e.file_type().unwrap();
        if kind.is_symlink() {
            continue;
        }
        let path = e.path();
        if kind.is_dir() {
            rust_files(&path, out);
        } else if path.extension().is_some_and(|x| x == "rs") {
            out.push(path);
        }
    }
}

/// REQ-2, AC-2: a pid name joined onto the temp folder is reported, on one line or over several.
#[test]
fn temp_names_scan_flags_a_pid_name_under_the_temp_folder() {
    let one = "fn f() {\n    let p = std::env::temp_dir().join(format!(\"x-{}\", std::process::id()));\n}\n";
    assert_eq!(pid_names_under_temp(one), vec![2]);

    let many = "fn f() {\n    let n = 1;\n    let root =\n        std::env::temp_dir().join(format!(\"c-{}-{}\", std::process::id(), n));\n    g();\n}\n";
    assert_eq!(pid_names_under_temp(many), vec![4]);

    let bare = "fn f() { let base = temp_dir().join(format!(\"a-{}\", process::id())); }";
    assert_eq!(pid_names_under_temp(bare), vec![1]);

    let two = "let a = temp_dir().join(format!(\"{}\", std::process::id()));\nlet b = temp_dir().join(format!(\"{}\", std::process::id()));\n";
    assert_eq!(pid_names_under_temp(two), vec![1, 2]);
}

/// REQ-3, AC-3: a pid used as data, a comment and a string literal are not reported.
#[test]
fn temp_names_scan_ignores_a_pid_that_is_data_and_text_that_is_not_code() {
    let key = "let key = format!(\"moai-test-key-{}\", std::process::id());\n";
    assert!(pid_names_under_temp(key).is_empty());

    let owner = "let me = std::process::id();\nlet dir = std::env::temp_dir();\n";
    assert!(pid_names_under_temp(owner).is_empty(), "two statements");

    let block =
        "fn f() { let a = std::env::temp_dir(); }\nfn g() { let me = std::process::id(); }\n";
    assert!(pid_names_under_temp(block).is_empty(), "two blocks");

    let comment = "// let p = temp_dir().join(format!(\"{}\", std::process::id()));\n/* temp_dir() std::process::id() */\nlet x = 1;\n";
    assert!(pid_names_under_temp(comment).is_empty(), "comments");

    let text =
        "let s = \"temp_dir() std::process::id()\";\nlet r = r#\"temp_dir() process::id()\"#;\n";
    assert!(pid_names_under_temp(text).is_empty(), "string literals");

    // A brace inside a string does not cut a statement; a quote inside a character literal does
    // not open a string.
    let braces = "let p = temp_dir().join(format!(\"a-{}-{}\", 1, 2));\nlet q = '\"';\nlet me = process::id();\n";
    assert!(
        pid_names_under_temp(braces).is_empty(),
        "format braces and a quote char"
    );
    let cut = "let p = temp_dir().join(format!(\"a-{}\", process::id()));\n";
    assert_eq!(
        pid_names_under_temp(cut),
        vec![1],
        "format braces keep one statement"
    );
}

/// REQ-1, AC-1 (SPEC-20261124-pid-scan-followup): `br"…"` and `cr"…"` are raw literals, so a
/// backslash before the closing quote does not swallow the code after them, and text inside them
/// is not scanned.
#[test]
fn temp_names_scan_reads_raw_byte_strings_as_raw() {
    let byte = r##"let a = br"C:\";
let p = temp_dir().join(format!("{}", process::id()));
"##;
    assert_eq!(pid_names_under_temp(byte), vec![2], "br with a backslash");

    let hashed = r##"let a = br#"a"b"#;
let p = temp_dir().join(format!("{}", process::id()));
"##;
    assert_eq!(pid_names_under_temp(hashed), vec![2], "br with hashes");

    let cstr = r##"let a = cr"x\";
let p = temp_dir().join(format!("{}", process::id()));
"##;
    assert_eq!(pid_names_under_temp(cstr), vec![2], "cr with a backslash");

    let inside = r##"let a = br"temp_dir() process::id()";
let b = br#"temp_dir() "x" process::id()"#;
let c = cr"temp_dir() process::id()";
"##;
    assert!(
        pid_names_under_temp(inside).is_empty(),
        "text inside raw literals"
    );
}

/// The number of files read under `<root>/tests` and the `file:line` of every pid name found.
fn tree_violations(root: &Path) -> (usize, Vec<String>) {
    let mut files = Vec::new();
    rust_files(&root.join("tests"), &mut files);
    let count = files.len();
    let mut found = Vec::new();
    for file in files {
        let rel = file
            .strip_prefix(root)
            .unwrap()
            .to_string_lossy()
            .replace('\\', "/");
        let text = String::from_utf8_lossy(&fs::read(&file).unwrap()).into_owned();
        for line in pid_names_under_temp(&text) {
            found.push(format!("{rel}:{line}"));
        }
    }
    (count, found)
}

/// REQ-2, AC-4: the checked-in `tests/` folder holds no pid name under the temp folder.
#[test]
fn temp_names_tree_has_no_pid_names_under_the_temp_folder() {
    let (count, found) = tree_violations(Path::new(ROOT));
    assert!(count > 100, "the walk found {count} files");
    assert!(
        found.is_empty(),
        "a scratch folder named from the process id (use tempfile::TempDir): {found:?}"
    );
}

/// REQ-2, AC-2 (SPEC-20261124-pid-scan-followup): the walk reads every file, one named like the
/// scan file included, so a violation written into the scan file itself is reported.
#[test]
fn temp_names_tree_walk_reads_a_file_named_like_the_scan_file() {
    let dir = tempfile::tempdir().unwrap();
    let tests = dir.path().join("tests");
    fs::create_dir_all(&tests).unwrap();
    let bad = "fn f() {\n    let p = std::env::temp_dir().join(format!(\"x-{}\", std::process::id()));\n}\n";
    fs::write(tests.join("test_temp_names_scan.rs"), bad).unwrap();
    fs::write(tests.join("clean.rs"), "fn g() {}\n").unwrap();
    let (count, found) = tree_violations(dir.path());
    assert_eq!(count, 2);
    assert_eq!(found, vec!["tests/test_temp_names_scan.rs:2".to_string()]);
}
