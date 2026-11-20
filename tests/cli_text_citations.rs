//! SPEC-20261120-cli-text-citations: the text the commands print carries no card number, no Issue
//! or pull-request number, no citation of a user or leader decision and no name of a moai-adk
//! development script. The scan reads the `src/` files of this checkout, leaves out comments and the
//! inline test module, and looks at what is left (strings included); nothing is run and no HOME is
//! touched.
//!
//! `PENDING` lists the source files that still carry such text. It can only shrink: a file that
//! was cleaned must leave the list in the same change, so the scan keeps the rest clean.

use std::{
    fs,
    path::{Path, PathBuf},
};

/// Source files that still carry citations, sorted. Remove an entry when its file is cleaned.
const PENDING: [&str; 15] = [
    "src/bundle.rs",
    "src/claim_timeout.rs",
    "src/claude.rs",
    "src/code_refs.rs",
    "src/flake_record.rs",
    "src/gate.rs",
    "src/gate_merge.rs",
    "src/go_state.rs",
    "src/home_layout.rs",
    "src/hooks.rs",
    "src/hosts/codex.rs",
    "src/lane_env.rs",
    "src/optimize.rs",
    "src/project_update.rs",
    "src/session.rs",
];

/// The files this change cleaned; each has a section in the rationale document.
const CLEANED: [&str; 16] = [
    "src/claim.rs",
    "src/claim_verify.rs",
    "src/gate_schedule.rs",
    "src/handoff.rs",
    "src/heavy_slot.rs",
    "src/help.rs",
    "src/home_state.rs",
    "src/launcher_init.rs",
    "src/ledger/open.rs",
    "src/project_hosts.rs",
    "src/recovery.rs",
    "src/shim_check.rs",
    "src/worktree.rs",
    "src/worktree_gc.rs",
    "src/worktree_include.rs",
    "src/worktree_report.rs",
];

/// The rationale document, relative to the checkout.
const RATIONALE: &str = "docs/cli-text-rationale.md";

/// Names of moai-adk's own development scripts and repository, which a user does not have.
const DEV_WORDS: [&str; 9] = [
    "moai-adk repository",
    "local_check.sh",
    "check-guard",
    "M-Flow",
    "tools/self_check.sh",
    "tools/source_manifest.py",
    "tools/build_folder.py",
    "tools/smoke/",
    "tools/ci_scope.py",
];

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

fn is_word(c: char) -> bool {
    c.is_alphanumeric() || c == '_'
}

/// The text of `line` from byte `at`, trimmed at the start.
fn after(line: &str, at: usize) -> &str {
    line[at..].trim_start()
}

/// The source of a file without comments and without the inline test module: comments become
/// blanks (line breaks stay, so line numbers hold), strings and the rest stay as they are.
fn code_of(text: &str) -> String {
    let end = text.find("\n#[cfg(test)]\n").map_or(text.len(), |i| i + 1);
    let chars: Vec<char> = text[..end].chars().collect();
    let mut out = String::with_capacity(end);
    let mut i = 0;
    while i < chars.len() {
        let c = chars[i];
        let next = chars.get(i + 1).copied().unwrap_or(' ');
        if c == '/' && next == '/' {
            while i < chars.len() && chars[i] != '\n' {
                i += 1;
            }
        } else if c == '/' && next == '*' {
            let mut depth = 0;
            while i < chars.len() {
                if chars[i] == '/' && chars.get(i + 1) == Some(&'*') {
                    depth += 1;
                    i += 2;
                } else if chars[i] == '*' && chars.get(i + 1) == Some(&'/') {
                    depth -= 1;
                    i += 2;
                    if depth == 0 {
                        break;
                    }
                } else {
                    if chars[i] == '\n' {
                        out.push('\n');
                    }
                    i += 1;
                }
            }
        } else if c == '"' {
            out.push(c);
            i += 1;
            while i < chars.len() && chars[i] != '"' {
                if chars[i] == '\\' && i + 1 < chars.len() {
                    out.push(chars[i]);
                    i += 1;
                }
                out.push(chars[i]);
                i += 1;
            }
            if i < chars.len() {
                out.push('"');
                i += 1;
            }
        } else if c == 'r' && (next == '"' || next == '#') && !out.ends_with(is_word) {
            // A raw string: `r"…"` or `r#"…"#`.
            let mut j = i + 1;
            let mut hashes = 0;
            while chars.get(j) == Some(&'#') {
                hashes += 1;
                j += 1;
            }
            if chars.get(j) == Some(&'"') {
                let close: Vec<char> = std::iter::once('"')
                    .chain(std::iter::repeat_n('#', hashes))
                    .collect();
                out.extend(&chars[i..=j]);
                i = j + 1;
                while i < chars.len() && !chars[i..].starts_with(&close) {
                    out.push(chars[i]);
                    i += 1;
                }
                let stop = (i + close.len()).min(chars.len());
                out.extend(&chars[i..stop]);
                i = stop;
            } else {
                out.push(c);
                i += 1;
            }
        } else if c == '\'' {
            // A character literal (`'x'`, `'\n'`, `'"'`) is copied whole; a lifetime is not.
            let literal = if next == '\\' { 4 } else { 3 };
            let whole: String = chars[i..(i + literal).min(chars.len())].iter().collect();
            if next != '\'' && whole.ends_with('\'') && whole.chars().count() >= 3 {
                out.push_str(&whole);
                i += whole.chars().count();
            } else {
                out.push(c);
                i += 1;
            }
        } else {
            out.push(c);
            i += 1;
        }
    }
    out
}

/// Every kind of citation on one line: `card` (C-82; the sample ids C-1 to C-9 are not), `number`
/// (#2345, Issue 2345, PR 2345), `decision` (a user or leader decision followed by a date or an
/// id) and `tool` (a development script name).
fn kinds(line: &str) -> Vec<&'static str> {
    let mut out = Vec::new();
    let chars: Vec<(usize, char)> = line.char_indices().collect();
    for (k, &(at, c)) in chars.iter().enumerate() {
        let before = if k == 0 { ' ' } else { chars[k - 1].1 };
        if c == 'C' && line[at..].starts_with("C-") && !is_word(before) && before != '-' {
            let digits = line[at + 2..]
                .chars()
                .take_while(char::is_ascii_digit)
                .count();
            let next = line[at + 2 + digits..].chars().next().unwrap_or(' ');
            if (2..=4).contains(&digits) && !next.is_ascii_digit() {
                out.push("card");
            }
        }
        if c == '#' && !is_word(before) && before != '&' {
            let digits = line[at + 1..]
                .chars()
                .take_while(char::is_ascii_digit)
                .count();
            let next = line[at + 1 + digits..].chars().next().unwrap_or(' ');
            // `#8839EF` is a colour: a number is not followed by a letter or a digit.
            if (3..=5).contains(&digits) && !is_word(next) {
                out.push("number");
            }
        }
    }
    // ASCII lower case keeps every byte position, so an index into `low` is one into `line`.
    let low = line.to_ascii_lowercase();
    for word in ["issue", "issues", "pr", "prs"] {
        for (at, _) in low.match_indices(word) {
            // `--issue 2904` is a sample option, not a citation: a `-` before the word excuses it.
            let word_before = low[..at]
                .chars()
                .next_back()
                .is_some_and(|c| is_word(c) || c == '-');
            let rest = &low[at + word.len()..];
            // `Issue #2619` is the `#` rule's; `Issue 2619` and `Issue#2619` are this one's.
            let digits_at = match rest.chars().next() {
                Some('#') => Some(1),
                Some(' ') => {
                    let trimmed = rest.trim_start();
                    (!trimmed.starts_with('#')).then_some(rest.len() - trimmed.len())
                }
                _ => None,
            };
            if let (false, Some(skip)) = (word_before, digits_at) {
                let digits = rest[skip..]
                    .chars()
                    .take_while(char::is_ascii_digit)
                    .count();
                let next = rest[skip + digits..].chars().next().unwrap_or(' ');
                if (3..=5).contains(&digits) && !is_word(next) {
                    out.push("number");
                }
            }
        }
    }
    for phrase in [
        "user decision",
        "leader decision",
        "user selection",
        "user's decision",
        "leader's decision",
    ] {
        for (at, _) in low.match_indices(phrase) {
            let rest = after(line, at + phrase.len());
            let rest = rest.strip_prefix(',').map_or(rest, str::trim_start);
            let token: String = rest.chars().take_while(|c| is_word(*c)).collect();
            let first = token.chars().next().unwrap_or(' ');
            if first.is_ascii_digit() || (first.is_ascii_uppercase() && token.len() <= 3) {
                out.push("decision");
            }
        }
    }
    for phrase in ["사용자 결정", "리더 결정"] {
        if line.contains(phrase) {
            out.push("decision");
        }
    }
    for (at, _) in low.match_indices("decision ") {
        let token: String = low[at + 9..].chars().take_while(|c| is_word(*c)).collect();
        let mut it = token.chars();
        let letter_digit =
            matches!(it.next(), Some('d' | 'u')) && it.next().is_some_and(|c| c.is_ascii_digit());
        if letter_digit || (token.len() >= 7 && token.chars().all(|c| c.is_ascii_digit())) {
            out.push("decision");
        }
    }
    for word in DEV_WORDS {
        if line.contains(word) {
            out.push("tool");
        }
    }
    out
}

/// Every citation of a source text as `(line number, kind)`, comments and the test module left out.
fn citations(text: &str) -> Vec<(usize, &'static str)> {
    code_of(text)
        .lines()
        .enumerate()
        .flat_map(|(i, l)| kinds(l).into_iter().map(move |k| (i + 1, k)))
        .collect()
}

/// Every source file of `src/` as `(path relative to the checkout with `/`, text)`, sorted.
fn sources() -> Vec<(String, String)> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                walk(&path, out);
            } else if path.extension().is_some_and(|e| e == "rs") {
                out.push(path);
            }
        }
    }
    let mut files = Vec::new();
    walk(&root().join("src"), &mut files);
    let mut out: Vec<(String, String)> = files
        .iter()
        .filter_map(|p| {
            let text = fs::read_to_string(p).ok()?;
            let rel = p
                .strip_prefix(root())
                .ok()?
                .components()
                .map(|c| c.as_os_str().to_string_lossy().into_owned())
                .collect::<Vec<_>>()
                .join("/");
            Some((rel, text))
        })
        .collect();
    out.sort();
    out
}

/// AC-1: the scan sees each kind of citation and leaves comments, the test module, colours and the
/// sample ids alone.
#[test]
fn scan_flags_each_kind_and_leaves_comments_alone() {
    for (line, kind) in [
        ("\"merge train (card C-82)\"", "card"),
        ("\"project check (#2496)\"", "number"),
        ("\"moved (Issue #2633)\"", "number"),
        ("\"see Issue 2619\"", "number"),
        ("\"PR 3234 merged\"", "number"),
        ("\"(user decision 2026-11-10)\"", "decision"),
        ("\"user decision A 2026-11-15\"", "decision"),
        ("\"leader decision 6903832699 says\"", "decision"),
        ("\"(decision D2)\"", "decision"),
        ("\"decision 6886824285 item 4\"", "decision"),
        ("\"run tools/local_check.sh\"", "tool"),
        ("\"the M-Flow bundle\"", "tool"),
    ] {
        assert!(
            citations(line).iter().any(|(_, k)| *k == kind),
            "`{line}` should carry a {kind}"
        );
    }
    for text in [
        "let a = 1; // (#2561) a trailing comment\n",
        "\"lane-resume-note --lane agent-1 --issue 2904 --claim 6874514212\"\n",
        "// (#2561) a comment line\nlet b = 2;\n",
        "/* (#2561 and C-82) a block comment */ let c = 3;\n",
        "/// (user decision 2026-11-10) a doc comment\nfn f() {}\n",
        "let color = \"#8839EF\"; let other = \"#209FB5\";\n",
        "\"card note C-1 --kind report\" \"card ready C-9\"\n",
        "\"ABC-12 and CI-3\" \"a leader decision is recorded with its answer\"\n",
        "\"resume #{number} with the leader decision in comment {}\"\n",
        "fn code() {}\n#[cfg(test)]\nmod tests { fn t() { let _ = \"#3333 C-82\"; } }\n",
    ] {
        assert_eq!(citations(text), Vec::new(), "`{text}` should be clean");
    }
    // A `//` inside a string is no comment, and a raw string and a character literal are read whole.
    for text in [
        "let url = \"http://x (#2561)\";\n",
        "let raw = r#\"has \"quotes\" and (#2561)\"#;\n",
        "let q = '\"'; let s = \"(#2561)\";\n",
    ] {
        assert!(
            citations(text).iter().any(|(_, k)| *k == "number"),
            "`{text}` should carry a number"
        );
    }
}

/// AC-1: no source file outside the pending list carries a citation.
#[test]
fn scan_source_files_outside_the_pending_list_are_clean() {
    let mut bad = Vec::new();
    for (path, text) in sources() {
        if PENDING.contains(&path.as_str()) {
            continue;
        }
        let found = citations(&text);
        if !found.is_empty() {
            let first: Vec<String> = found
                .iter()
                .take(4)
                .map(|(l, k)| format!("line {l} {k}"))
                .collect();
            bad.push(format!("{path}: {} ({})", found.len(), first.join(", ")));
        }
    }
    assert!(
        bad.is_empty(),
        "source files with citations:\n{}",
        bad.join("\n")
    );
}

/// AC-1: the pending list is sorted, names source files, and only files that still carry
/// citations (a cleaned file has to leave it).
#[test]
fn scan_pending_list_only_shrinks() {
    let mut sorted = PENDING.to_vec();
    sorted.sort_unstable();
    sorted.dedup();
    assert_eq!(
        sorted,
        PENDING.to_vec(),
        "PENDING is sorted and has no duplicates"
    );
    let files = sources();
    for path in PENDING {
        let text = files
            .iter()
            .find(|(p, _)| p == path)
            .unwrap_or_else(|| panic!("PENDING names `{path}`, which is not a source file"));
        assert!(
            !citations(&text.1).is_empty(),
            "`{path}` carries no citation any more: remove it from PENDING"
        );
    }
}

/// AC-2: the rationale document has one section per cleaned file, titled with its path.
#[test]
fn rationale_has_a_section_per_cleaned_file() {
    let text = fs::read_to_string(root().join(RATIONALE))
        .unwrap_or_else(|e| panic!("{RATIONALE} is missing: {e}"));
    for path in CLEANED {
        let heading = format!("## `{path}`");
        assert!(
            text.lines().any(|l| l.trim_end() == heading),
            "{RATIONALE} has no section `{heading}`"
        );
    }
}

/// AC-2: a cleaned file carries no citation and is not on the pending list.
#[test]
fn rationale_cleaned_files_carry_no_citation() {
    let files = sources();
    for path in CLEANED {
        assert!(!PENDING.contains(&path), "`{path}` is cleaned, not pending");
        let (_, text) = files
            .iter()
            .find(|(p, _)| p == path)
            .unwrap_or_else(|| panic!("`{path}` is not a source file"));
        assert_eq!(
            citations(text),
            Vec::new(),
            "`{path}` still carries citations"
        );
    }
}
