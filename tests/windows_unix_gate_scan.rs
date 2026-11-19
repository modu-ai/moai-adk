//! SPEC-20261119-windows-unix-gate (card C-110): a file under `tests/` that names
//! `std::os::unix` outside an item that `cfg` removes on Windows breaks the push run's
//! `windows-build` job (`cargo build --all-targets` on Windows). A pull request runs Linux and
//! macOS only, so no PR check saw `tests/hooks_cli.rs:1391` (`std::os::unix::fs::symlink`
//! without a gate) before it reached the tip. This scan reads the same fact on Linux and macOS:
//! every `os::unix` mention in `tests/**/*.rs` must sit in an item, a block, a statement or a
//! whole file that a `cfg` attribute turns off for `target_os = "windows"`.
//!
//! The reading is lexical, not a compiler: comments, strings and character literals are blanked,
//! then the code is walked once. A `#[cfg(...)]` attribute whose expression is false on Windows
//! gates the next brace block or the next statement that ends in `;`; `#![cfg(...)]` gates the
//! file. A gate on something with neither a block nor a `;` (a struct field, a match arm) lasts
//! until the next block or statement: that is a missed catch, never a false alarm.

use std::fs;
use std::path::{Path, PathBuf};

/// Blanks comments, string literals (also raw ones) and character literals and keeps every
/// newline and every byte offset, so line numbers and brace matching survive.
fn strip(src: &str) -> String {
    let b = src.as_bytes();
    let n = b.len();
    let mut out = b.to_vec();
    let blank = |out: &mut Vec<u8>, from: usize, to: usize| {
        for byte in &mut out[from..to.min(n)] {
            if *byte != b'\n' {
                *byte = b' ';
            }
        }
    };
    let ident = |c: u8| c.is_ascii_alphanumeric() || c == b'_';
    let mut i = 0;
    while i < n {
        if b[i..].starts_with(b"//") {
            let j = b[i..].iter().position(|&c| c == b'\n').map_or(n, |p| i + p);
            blank(&mut out, i, j);
            i = j;
        } else if b[i..].starts_with(b"/*") {
            let (mut depth, mut j) = (1, i + 2);
            while j < n && depth > 0 {
                if b[j..].starts_with(b"/*") {
                    depth += 1;
                    j += 2;
                } else if b[j..].starts_with(b"*/") {
                    depth -= 1;
                    j += 2;
                } else {
                    j += 1;
                }
            }
            blank(&mut out, i, j);
            i = j;
        } else if b[i] == b'r'
            && (i == 0 || !ident(b[i - 1]) || (b[i - 1] == b'b' && (i < 2 || !ident(b[i - 2]))))
            && raw_string_open(&b[i..]).is_some()
        {
            let (hashes, open) = raw_string_open(&b[i..]).unwrap();
            let mut close = vec![b'"'];
            close.extend(std::iter::repeat_n(b'#', hashes));
            let from = i + open;
            let j = (from..n)
                .find(|&k| b[k..].starts_with(&close))
                .map_or(n, |k| k + close.len());
            blank(&mut out, i, j);
            i = j;
        } else if b[i] == b'"' {
            let mut j = i + 1;
            while j < n && b[j] != b'"' {
                j += if b[j] == b'\\' { 2 } else { 1 };
            }
            blank(&mut out, i, j + 1);
            i = j + 1;
        } else if b[i] == b'\'' {
            // A character literal, or a lifetime (`'a`), which has no closing quote.
            let close = if b.get(i + 1) == Some(&b'\\') {
                (i + 2..n.min(i + 12)).find(|&k| b[k] == b'\'')
            } else {
                let len = src[i + 1..].chars().next().map_or(0, char::len_utf8);
                (len > 0 && b.get(i + 1 + len) == Some(&b'\'')).then_some(i + 1 + len)
            };
            match close {
                Some(k) => {
                    blank(&mut out, i, k + 1);
                    i = k + 1;
                }
                None => i += 1,
            }
        } else {
            i += 1;
        }
    }
    String::from_utf8(out).expect("blanking keeps UTF-8")
}

/// `r"`, `r#"`, `r##"` …: the number of hashes and the length of the opening.
fn raw_string_open(b: &[u8]) -> Option<(usize, usize)> {
    let hashes = b[1..].iter().take_while(|&&c| c == b'#').count();
    (b.get(1 + hashes) == Some(&b'"')).then_some((hashes, 2 + hashes))
}

/// Value of a `cfg` expression when the target is Windows: `Some(true)` or `Some(false)` for
/// what this scan can decide (`unix`, `windows`, `target_os`, `target_family`), `None` for a name
/// it cannot (`test`, `debug_assertions`, a feature, `target_vendor`). `None` never gates: the
/// item stays in the build as far as the scan knows, and `not(...)` of it stays `None` too.
fn cfg_on_windows(expr: &str) -> Option<bool> {
    let expr = expr.trim();
    for op in ["all", "any", "not"] {
        if let Some(inner) = expr
            .strip_prefix(op)
            .map(str::trim_start)
            .and_then(|r| r.strip_prefix('('))
            .and_then(|r| r.strip_suffix(')'))
        {
            let values: Vec<Option<bool>> = split_top_level(inner)
                .iter()
                .map(|p| cfg_on_windows(p))
                .collect();
            return match op {
                // One false part settles `all`; one true part settles `any`; otherwise a
                // part the scan cannot decide leaves the whole unknown.
                "all" if values.contains(&Some(false)) => Some(false),
                "any" if values.contains(&Some(true)) => Some(true),
                "all" | "any" if values.contains(&None) => None,
                "all" => Some(true),
                "any" => Some(false),
                _ => values.first().copied().flatten().map(|v| !v),
            };
        }
    }
    match expr {
        "unix" => Some(false),
        "windows" => Some(true),
        _ => match expr.split_once('=') {
            Some((key, value)) => {
                let value = value.trim().trim_matches('"');
                match key.trim() {
                    "target_family" | "target_os" => Some(value == "windows"),
                    _ => None,
                }
            }
            None => None,
        },
    }
}

/// Whether the text after `os::` names the `unix` module: `unix::fs::…`, or `unix` as one of
/// the names of a brace group (`{raw, unix}`, `{self, unix::fs::PermissionsExt}`).
fn names_unix(after_os: &str) -> bool {
    let word_is_unix = |s: &str| {
        s.trim_start()
            .strip_prefix("unix")
            .is_some_and(|tail| !tail.starts_with(|c: char| c.is_ascii_alphanumeric() || c == '_'))
    };
    let rest = after_os.trim_start();
    let Some(group) = rest.strip_prefix('{') else {
        return word_is_unix(rest);
    };
    let (mut depth, mut end) = (1usize, group.len());
    for (k, ch) in group.char_indices() {
        match ch {
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    end = k;
                    break;
                }
            }
            _ => {}
        }
    }
    split_top_level(&group[..end])
        .iter()
        .any(|part| word_is_unix(part))
}

fn split_top_level(s: &str) -> Vec<String> {
    let (mut parts, mut depth, mut cur) = (Vec::new(), 0i32, String::new());
    for ch in s.chars() {
        // Parentheses for `cfg(all(a, b))`, braces for `use a::{b, c}`.
        match ch {
            '(' | '{' => depth += 1,
            ')' | '}' => depth -= 1,
            _ => {}
        }
        if ch == ',' && depth == 0 {
            parts.push(std::mem::take(&mut cur));
        } else {
            cur.push(ch);
        }
    }
    if !cur.trim().is_empty() {
        parts.push(cur);
    }
    parts
}

/// What one file says: whether `#![cfg(...)]` removes the whole file on Windows, and the
/// 1-based lines that name `os::unix` outside anything that `cfg` removes.
#[derive(Debug, PartialEq)]
struct Reading {
    file_gated: bool,
    lines: Vec<usize>,
}

fn read_file(src: &str) -> Reading {
    let code = strip(src);
    let b = code.as_bytes();
    let n = b.len();
    let ident = |c: u8| c.is_ascii_alphanumeric() || c == b'_';
    let mut lines = Vec::new();
    let mut stack: Vec<bool> = Vec::new();
    let (mut pending, mut file_gated, mut paren) = (false, false, 0usize);
    let mut i = 0;
    while i < n {
        if b[i..].starts_with(b"#[") || b[i..].starts_with(b"#![") {
            let inner = b[i..].starts_with(b"#![");
            let open = i + if inner { 3 } else { 2 };
            let (mut depth, mut j) = (1, open);
            while j < n && depth > 0 {
                match b[j] {
                    b'[' => depth += 1,
                    b']' => depth -= 1,
                    _ => {}
                }
                j += 1;
            }
            // The attribute is read from the source, not from the blanked copy: the string in
            // `target_os = "windows"` is the value. Offsets agree and both ends are ASCII.
            let body = src[open..j.saturating_sub(1).max(open)].trim();
            if let Some(expr) = body
                .strip_prefix("cfg")
                .map(str::trim_start)
                .and_then(|r| r.strip_prefix('('))
                .and_then(|r| r.strip_suffix(')'))
            {
                if cfg_on_windows(expr) == Some(false) {
                    if inner {
                        file_gated = true;
                    } else {
                        pending = true;
                    }
                }
            }
            i = j;
            continue;
        }
        match b[i] {
            b'{' => {
                let outer = stack.last().copied().unwrap_or(false);
                stack.push(pending || outer);
                pending = false;
            }
            b'}' => {
                stack.pop();
            }
            b'(' | b'[' => paren += 1,
            b')' | b']' => paren = paren.saturating_sub(1),
            b';' if paren == 0 => pending = false,
            b'o' if b[i..].starts_with(b"os::") && (i == 0 || !ident(b[i - 1])) => {
                let gated = file_gated || pending || stack.last().copied().unwrap_or(false);
                if names_unix(&code[i + 4..]) && !gated {
                    lines.push(code[..i].matches('\n').count() + 1);
                }
            }
            _ => {}
        }
        i += 1;
    }
    Reading { file_gated, lines }
}

/// Every `.rs` file under `tests/`, sorted, as (path relative to the crate root with `/`, text).
fn test_files(root: &Path) -> Vec<(String, String)> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        let mut entries: Vec<PathBuf> = fs::read_dir(dir)
            .unwrap_or_else(|e| panic!("read {}: {e}", dir.display()))
            .map(|e| e.unwrap().path())
            .collect();
        entries.sort();
        for p in entries {
            if p.is_dir() {
                walk(&p, out);
            } else if p.extension().is_some_and(|e| e == "rs") {
                out.push(p);
            }
        }
    }
    let mut paths = Vec::new();
    walk(&root.join("tests"), &mut paths);
    paths
        .into_iter()
        .map(|p| {
            let rel = p
                .strip_prefix(root)
                .unwrap()
                .to_string_lossy()
                .replace('\\', "/");
            let text = fs::read_to_string(&p).unwrap_or_else(|e| panic!("read {rel}: {e}"));
            (rel, text)
        })
        .collect()
}

/// The ungated mentions of the given files as `path:line` strings. A helper under
/// `tests/common/` (not `mod.rs`) is read through its includers: its mentions are fine when it
/// is pulled in by at least one file and every such file is itself removed on Windows.
fn ungated(files: &[(String, String)]) -> Vec<String> {
    let readings: Vec<Reading> = files.iter().map(|(_, t)| read_file(t)).collect();
    let mut found = Vec::new();
    for (k, (path, _)) in files.iter().enumerate() {
        if readings[k].lines.is_empty() {
            continue;
        }
        let helper = path.contains("/common/") && !path.ends_with("/mod.rs");
        if helper {
            let name = path.rsplit('/').next().unwrap();
            let needle = format!("common/{name}\"");
            let includers: Vec<&Reading> = files
                .iter()
                .enumerate()
                .filter(|(j, (_, text))| *j != k && text.contains(&needle))
                .map(|(j, _)| &readings[j])
                .collect();
            if !includers.is_empty() && includers.iter().all(|r| r.file_gated) {
                continue;
            }
        }
        for line in &readings[k].lines {
            found.push(format!("{path}:{line}"));
        }
    }
    found
}

/// The scan must see more than this many `.rs` files under `tests/` (the tree holds a few
/// hundred), so a moved folder or a failed walk is a failure and not an empty, passing list.
const MIN_FILES: usize = 100;

fn saw_enough_files(count: usize) -> bool {
    count > MIN_FILES
}

#[test]
fn no_file_under_tests_names_std_os_unix_outside_cfg_unix() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let files = test_files(root);
    assert!(
        saw_enough_files(files.len()),
        "the scan saw {} files under tests/ (it needs more than {MIN_FILES}); the folder moved?",
        files.len()
    );
    let found = ungated(&files);
    assert!(
        found.is_empty(),
        "these lines name `os::unix` where Windows still compiles them; the push run's \
         windows-build job fails on them. Put `#[cfg(unix)]` on the item, block or statement \
         (or `#![cfg(unix)]` on the file), or write a Windows branch:\n{}",
        found.join("\n")
    );
}

fn lines(src: &str) -> Vec<usize> {
    read_file(src).lines
}

#[test]
fn an_ungated_mention_is_found_with_its_line() {
    assert_eq!(
        lines("fn t() {\n    std::os::unix::fs::symlink(a, b).unwrap();\n}\n"),
        [2]
    );
    // The brace form of a use names `unix` too.
    assert_eq!(
        lines("use std::os::{unix::fs::PermissionsExt, raw};\n"),
        [1]
    );
    // Another name that merely starts with `unix` is not it.
    assert!(lines("use std::os::unixish::x;\nfn f() { os::unix_like(); }\n").is_empty());
}

#[test]
fn a_cfg_that_turns_the_item_off_on_windows_gates_it() {
    for gate in [
        "#[cfg(unix)]",
        "#[cfg(not(windows))]",
        "#[cfg(target_os = \"macos\")]",
        "#[cfg(target_os = \"linux\")]",
        "#[cfg(target_family = \"unix\")]",
        "#[cfg(all(unix, not(target_os = \"macos\")))]",
        "#[cfg(all(unix, test))]",
        "#[cfg( unix )]",
    ] {
        let src = format!("{gate}\nfn t() {{\n    std::os::unix::fs::symlink(a, b);\n}}\n");
        assert!(lines(&src).is_empty(), "{gate}");
    }
}

#[test]
fn the_file_count_floor_has_a_boundary() {
    assert!(!saw_enough_files(0));
    assert!(!saw_enough_files(MIN_FILES));
    assert!(saw_enough_files(MIN_FILES + 1));
}

#[test]
fn a_name_the_scan_cannot_decide_never_gates() {
    // `None` stays `None` through `not`, and one false part still settles `all`.
    for gate in [
        "#[cfg(not(test))]",
        "#[cfg(not(feature = \"x\"))]",
        "#[cfg(target_vendor = \"apple\")]",
        "#[cfg(any(feature = \"x\", unix))]",
        "#[cfg(not(debug_assertions))]",
        "#[cfg(all(feature = \"x\", not(test)))]",
    ] {
        let src = format!("{gate}\nfn t() {{\n    std::os::unix::fs::symlink(a, b);\n}}\n");
        assert_eq!(lines(&src), [3], "{gate}");
    }
    let src =
        "#[cfg(all(unix, feature = \"x\"))]\nfn t() {\n    std::os::unix::fs::symlink(a, b);\n}\n";
    assert!(lines(src).is_empty());
    let src = "#[cfg(not(any(windows, feature = \"x\")))]\nfn t() {\n    std::os::unix::fs::symlink(a, b);\n}\n";
    assert!(lines(src).is_empty());
}

#[test]
fn the_names_of_a_brace_group_are_read_one_by_one() {
    for (src, expected) in [
        ("use std::os::{raw, unix};\n", vec![1]),
        ("use std::os::{self, unix::fs::PermissionsExt};\n", vec![1]),
        // The line is the one that holds `os::`, even when `unix` sits on a later line.
        ("use std::os::{ raw ,\n    unix::fs };\n", vec![1]),
        // A `unix` inside a nested group belongs to `other::`, not to `os::`.
        ("use std::os::{raw, unixish, other::{a, unix}};\n", vec![]),
        ("use std::os::{raw, fd};\n", vec![]),
        ("use a::os::{b, unix_x};\nuse std::os::unix_like;\n", vec![]),
    ] {
        assert_eq!(lines(src), expected, "{src}");
    }
}

#[test]
fn a_cfg_that_keeps_the_item_on_windows_does_not_gate_it() {
    for gate in [
        "#[cfg(any(unix, windows))]",
        "#[cfg(not(unix))]",
        "#[cfg(windows)]",
        "#[cfg(test)]",
        "#[cfg(target_os = \"windows\")]",
        "#[cfg(target_family = \"windows\")]",
        "#[cfg(any(unix, test))]",
        "#[test]",
    ] {
        let src = format!("{gate}\nfn t() {{\n    std::os::unix::fs::symlink(a, b);\n}}\n");
        assert_eq!(lines(&src), [3], "{gate}");
    }
}

#[test]
fn a_gate_covers_exactly_the_item_it_stands_on() {
    // A gated block inside a function: the line after the block is not covered.
    let src = "fn t() {\n    #[cfg(unix)]\n    {\n        std::os::unix::fs::symlink(a, b);\n    }\n    std::os::unix::fs::symlink(c, d);\n}\n";
    assert_eq!(lines(src), [6]);
    // A gated statement ends at its `;`.
    let src = "#[cfg(unix)]\nuse std::os::unix::fs::symlink;\nfn t() {\n    std::os::unix::fs::symlink(a, b);\n}\n";
    assert_eq!(lines(src), [4]);
    // A block nested inside a gated item is gated too; the same shape without the gate is not.
    let src = "#[cfg(unix)]\nfn t() {\n    if x {\n        std::os::unix::fs::symlink(a, b);\n    }\n}\nfn u() {\n    if x {\n        std::os::unix::fs::symlink(a, b);\n    }\n}\n";
    assert_eq!(lines(src), [9]);
    // A `;` inside brackets does not end the wait for the item's block.
    let src = "#[cfg(unix)]\nfn t(a: [u8; 3]) -> [u8; 2] {\n    std::os::unix::fs::symlink(a, b);\n    [0; 2]\n}\n";
    assert!(lines(src).is_empty());
    // A use inside a gated module is covered; the next module is not.
    let src = "#[cfg(unix)]\nmod a {\n    use std::os::unix::fs::symlink;\n}\nmod b {\n    use std::os::unix::fs::symlink;\n}\n";
    assert_eq!(lines(src), [6]);
    // Other attributes between the gate and the item do not lose it.
    let src =
        "#[cfg(unix)]\n#[test]\n#[ignore]\nfn t() {\n    std::os::unix::fs::symlink(a, b);\n}\n";
    assert!(lines(src).is_empty());
}

#[test]
fn an_inner_cfg_gates_the_whole_file() {
    let src = "//! doc\n\n#![cfg(unix)]\n\nfn t() {\n    std::os::unix::fs::symlink(a, b);\n}\n";
    let reading = read_file(src);
    assert!(reading.file_gated && reading.lines.is_empty());
    let reading = read_file("#![cfg(any(unix, windows))]\nuse std::os::unix::fs::symlink;\n");
    assert!(!reading.file_gated);
    assert_eq!(reading.lines, [2]);
}

#[test]
fn comments_strings_and_lifetimes_do_not_confuse_the_reading() {
    let src = concat!(
        "// std::os::unix::fs::symlink\n",
        "/* std::os::unix /* nested */ still a comment std::os::unix */\n",
        "const A: &str = \"std::os::unix::fs::symlink\";\n",
        "const B: &str = r#\"std::os::unix \"quoted\" \"#;\n",
        "const C: &[u8] = b\"std::os::unix\";\n",
        "fn f<'a>(x: &'a str, c: char) -> bool {\n",
        "    let q = '\"';\n",
        "    let e = '\\'';\n",
        "    std::os::unix::fs::symlink(x, x).is_ok() && c == q\n",
        "}\n",
    );
    assert_eq!(lines(src), [9]);
}

fn file(path: &str, text: &str) -> (String, String) {
    (path.to_string(), text.to_string())
}

#[test]
fn a_helper_under_common_is_read_through_its_includers() {
    let helper = file(
        "tests/common/helper.rs",
        "pub fn link() {\n    std::os::unix::fs::symlink(a, b);\n}\n",
    );
    let gated = file(
        "tests/a_cli.rs",
        "#![cfg(unix)]\n#[path = \"common/helper.rs\"]\nmod helper;\n",
    );
    let open = file(
        "tests/b_cli.rs",
        "#[path = \"common/helper.rs\"]\nmod helper;\n",
    );
    // Nobody includes it: the helper is judged alone.
    assert_eq!(
        ungated(std::slice::from_ref(&helper)),
        ["tests/common/helper.rs:2"]
    );
    // Every includer is removed on Windows: fine.
    assert!(ungated(&[helper.clone(), gated.clone()]).is_empty());
    // One includer keeps Windows in: the helper's line is reported.
    assert_eq!(
        ungated(&[helper.clone(), gated, open]),
        ["tests/common/helper.rs:2"]
    );
    // `mod.rs` is never read through includers: every test file includes it.
    let module = file(
        "tests/common/mod.rs",
        "pub fn link() {\n    std::os::unix::fs::symlink(a, b);\n}\n",
    );
    let gated = file("tests/a_cli.rs", "#![cfg(unix)]\nmod common;\n");
    assert_eq!(ungated(&[module, gated]), ["tests/common/mod.rs:2"]);
}
