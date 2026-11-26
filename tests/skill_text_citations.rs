//! SPEC-20261126-shipped-text-citations: a file that ships to users carries no card number, no
//! Issue or pull-request number, no citation of a user or leader decision and no name of a
//! moai-adk development script. The scan reads `skills/`, `agents/`, `templates/` and `install/`
//! of this checkout; nothing is run and no HOME is touched.
//!
//! What it does not flag, on purpose: the machine markers of the work flow (`LANE-BLOCKED` and
//! the like, which the tool reads), SPEC ids, CSS colour codes, one card-id format example, and
//! the ported pack content under `templates/bundle-packs/`.

use std::{
    fs,
    path::{Path, PathBuf},
};

/// The shipped folders, relative to the checkout.
const SHIPPED: [&str; 4] = ["skills", "agents", "templates", "install"];

/// Shipped folders the scan does not read: ported pack content with its own vocabulary (colour
/// codes, upstream issue numbers, taxonomy ids), none of which cites this project's development.
const EXEMPT_PREFIXES: [&str; 1] = ["templates/bundle-packs/"];

/// Lines that show a format and cite nothing: `(file, text the line starts with)`.
const FORMAT_LINES: [(&str, &str); 1] = [("skills/moai/references/audit.md", "AUDIT C-12 (")];

/// Shipped files that still carry citations, sorted. It can only shrink: a file that was cleaned
/// leaves the list in the same change.
const PENDING: [&str; 1] = ["skills/moai/references/source-manifest.json"];

/// Names of moai-adk's own development scripts and repository, which a user does not have.
const DEV_WORDS: [&str; 8] = [
    "moai-adk repository",
    "local_check.sh",
    "check-guard",
    "tools/self_check.sh",
    "tools/source_manifest.py",
    "tools/build_folder.py",
    "tools/smoke/",
    "tools/ci_scope.py",
];

/// Words that make a citation of the person or window that decided.
const DECISIONS: [&str; 7] = [
    "user decision",
    "leader decision",
    "user selection",
    "user's decision",
    "leader's decision",
    "사용자 결정",
    "리더 결정",
];

fn is_word(c: char) -> bool {
    c.is_alphanumeric() || c == '_'
}

/// Whether `text` starts with a number: ASCII digits whose count is in `lengths`, not running
/// into an ASCII letter (`#87867F` is a colour code, while a Korean particle may follow).
fn digits_in(text: &str, lengths: std::ops::RangeInclusive<usize>) -> bool {
    let n = text.chars().take_while(char::is_ascii_digit).count();
    lengths.contains(&n)
        && !text[n..]
            .chars()
            .next()
            .is_some_and(|c| c.is_ascii_alphanumeric())
}

/// Whether the `#` at byte `at` starts the value of a CSS colour declaration.
fn is_css_colour(line: &str, at: usize) -> bool {
    let before = line[..at].trim_end();
    ["color:", "background:"]
        .iter()
        .any(|p| before.ends_with(p))
}

/// Every kind of citation on one line: `card` (C-12), `number` (#2345), `decision` (a user or
/// leader decision with its date or id) and `tool` (a development script name).
fn kinds(line: &str) -> Vec<&'static str> {
    let mut out = Vec::new();
    let mut before = ' ';
    for (at, c) in line.char_indices() {
        let rest = &line[at..];
        if c == 'C'
            && rest.starts_with("C-")
            && !is_word(before)
            && before != '-'
            && digits_in(&rest[2..], 1..=4)
        {
            out.push("card");
        }
        if c == '#'
            && !is_word(before)
            && before != '&'
            && digits_in(&rest[1..], 3..=5)
            && !is_css_colour(line, at)
        {
            out.push("number");
        }
        before = c;
    }
    // ASCII lower case keeps every byte position, so an index into `low` is one into `line`.
    let low = line.to_ascii_lowercase();
    // `#2619` is the rule above; `Issue 2619`, `PR 3234` and `Issue#2619` are these.
    for word in ["issue", "issues", "pr", "prs"] {
        for (at, _) in low.match_indices(word) {
            if low[..at].chars().next_back().is_some_and(is_word) {
                continue;
            }
            let rest = &low[at + word.len()..];
            let number = match rest.strip_prefix('#') {
                Some(after_hash) => Some(after_hash),
                None => rest
                    .starts_with(' ')
                    .then(|| rest.trim_start())
                    .filter(|r| !r.starts_with('#')),
            };
            if number.is_some_and(|n| digits_in(n, 3..=5)) {
                out.push("number");
            }
        }
    }
    for phrase in DECISIONS {
        for (at, _) in low.match_indices(phrase) {
            let mut end = at + phrase.len();
            // The plural cites like the singular: `user decisions (<date>)`.
            let after = &low[end..];
            if after.starts_with('s') && !after[1..].chars().next().is_some_and(is_word) {
                end += 1;
            }
            // `(user decision)` names who decided and nothing a reader can follow.
            if low[..at].ends_with('(') && low[end..].starts_with(')') {
                out.push("decision");
            }
            let rest = line[end..].trim_start();
            let rest = rest.strip_prefix(',').map_or(rest, str::trim_start);
            let rest = rest.strip_prefix("of ").unwrap_or(rest);
            // What is cited may stand in parentheses: `user decision (<date or id>)`.
            let rest = rest.strip_prefix('(').unwrap_or(rest);
            let mut it = rest.chars();
            let (first, second) = (it.next().unwrap_or(' '), it.next().unwrap_or(' '));
            if first.is_ascii_digit()
                || first == '#'
                || (first.is_ascii_uppercase() && (second.is_ascii_digit() || second == '-'))
            {
                out.push("decision");
            }
        }
    }
    for (at, found) in low.match_indices("decision ") {
        let mut it = low[at + found.len()..].chars();
        if matches!(it.next(), Some('d' | 'u')) && it.next().is_some_and(|c| c.is_ascii_digit()) {
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

/// The cited lines of one shipped file as `(line number, kinds)`, its format lines left out.
fn citations(path: &str, text: &str) -> Vec<(usize, Vec<&'static str>)> {
    text.lines()
        .enumerate()
        .filter(|(_, line)| {
            !FORMAT_LINES
                .iter()
                .any(|(file, start)| *file == path && line.trim_start().starts_with(start))
        })
        .map(|(i, line)| (i + 1, kinds(line)))
        .filter(|(_, found)| !found.is_empty())
        .collect()
}

/// Every shipped text file as `(path relative to the checkout with `/`, text)`, sorted. A link is
/// never followed.
fn shipped() -> Vec<(String, String)> {
    fn walk(dir: &Path, rel: &str, out: &mut Vec<(String, PathBuf)>) {
        let entries = fs::read_dir(dir).unwrap_or_else(|e| panic!("{}: {e}", dir.display()));
        for entry in entries {
            let entry = entry.unwrap();
            let rel = format!("{rel}/{}", entry.file_name().to_string_lossy());
            let kind = entry.file_type().unwrap();
            if kind.is_dir() {
                walk(&entry.path(), &rel, out);
            } else if kind.is_file() {
                out.push((rel, entry.path()));
            }
        }
    }
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut files = Vec::new();
    for top in SHIPPED {
        walk(&root.join(top), top, &mut files);
    }
    files.sort();
    files
        .into_iter()
        .map(|(rel, path)| {
            let bytes = fs::read(&path).unwrap_or_else(|e| panic!("{rel}: {e}"));
            (rel, String::from_utf8_lossy(&bytes).into_owned())
        })
        .collect()
}

fn exempt(path: &str) -> bool {
    EXEMPT_PREFIXES.iter().any(|p| path.starts_with(p))
}

/// The text of the shipped file `path`.
fn text_of<'a>(files: &'a [(String, String)], path: &str) -> &'a str {
    files
        .iter()
        .find(|(p, _)| p == path)
        .map(|(_, t)| t.as_str())
        .unwrap_or_else(|| panic!("`{path}` is not a shipped file"))
}

/// AC-1: the scan sees each kind of citation and leaves markers, formats and colours alone.
#[test]
fn scan_flags_each_kind_and_leaves_markers_alone() {
    let mut problems = Vec::new();
    for (line, kind) in [
        ("Merge train (card C-51 and more)", "card"),
        ("card C-1234 is done", "card"),
        ("(see Issue #2619)", "number"),
        ("PR #3234 merged", "number"),
        ("Issue 2619 says so", "number"),
        ("PR 3234 was merged", "number"),
        ("see Issue#2619", "number"),
        ("# rule (#2935 L1)", "number"),
        ("the color of #2434; see there", "number"),
        ("(user decision 2026-11-16)", "decision"),
        ("(user decision of 2026-11-13 on it)", "decision"),
        ("integration gate of user decision U6", "decision"),
        ("MoAI uses it that way (user decision).", "decision"),
        ("a leader decision 2026-11-17 says", "decision"),
        ("Decision D7 of the review", "decision"),
        ("the user's decision 2026-11-17", "decision"),
        ("\u{130} the user decision 2026-11-16", "decision"),
        ("사용자 결정 2026-11-17", "decision"),
        ("리더 결정 D7 에 따라", "decision"),
        // A phrase followed by what it cites in parentheses, singular and plural.
        (
            "User decision (2026-11-22): the artifact version uses",
            "decision",
        ),
        (
            "User decisions (2026-11-22): html files use Pretendard",
            "decision",
        ),
        ("the leader decision (U-B 2026-11-17) says", "decision"),
        ("leader decisions 2026-11-17 say so", "decision"),
        ("both user's decisions (#2634) hold", "decision"),
        ("MoAI does it that way (user decisions).", "decision"),
        ("사용자 결정 (2026-11-17) 에 따라", "decision"),
        ("Issue #2434의 결정", "number"),
        ("in the moai-adk repository run it", "tool"),
        ("run `local_check.sh` first", "tool"),
        ("the check-guard job", "tool"),
    ] {
        if !kinds(line).contains(&kind) {
            problems.push(format!("`{line}` should carry a {kind}"));
        }
    }
    for line in [
        "`LANE-BLOCKED kind=decision` and `LANE-READY-INTEGRATION` are markers the tool reads",
        "ABC-12 is not a card and neither is the CI-3 job",
        "a card id looks like `C-<n>` in the text",
        "heading &#8212; and the colour #123456 and anchor #12",
        "        color: #555;",
        "background: #0000;",
        "| `--g500` | `#87867F` | Secondary text |",
        "--ivory: #17150F;",
        "    placeholder: '근거가 기록되고 사용자 결정이 났다'",
        "the user decision is recorded with its answer",
        "  needs:decision  a user decision is pending; work waits for it",
        "ask the user for a decision before the pick",
        "user decisions are recorded with their answers",
        "setting it is the user's decision (`guarded`); nothing does it",
        "a user decision (see the table) opens the work",
        "issue 12 of the series, and PR 12 of the other",
        "# (SPEC-20261117-installer-tty)",
    ] {
        if !kinds(line).is_empty() {
            problems.push(format!("`{line}` should be clean: {:?}", kinds(line)));
        }
    }
    assert!(problems.is_empty(), "{}", problems.join("\n"));
}

/// AC-2: no shipped file outside the pending list carries a citation.
#[test]
fn scan_shipped_files_outside_the_pending_list_are_clean() {
    let files = shipped();
    for expected in [
        "agents/audit.md",
        "install/install.sh",
        "skills/moai/references/source-manifest.json",
        "templates/github/manifest.json",
    ] {
        assert!(
            files.iter().any(|(p, _)| p == expected),
            "the walk does not list {expected}"
        );
    }
    let mut bad = Vec::new();
    for (path, text) in files {
        if exempt(&path) || PENDING.contains(&path.as_str()) {
            continue;
        }
        let found = citations(&path, &text);
        if !found.is_empty() {
            let lines: Vec<String> = found.iter().map(|(l, k)| format!("{l}:{}", k[0])).collect();
            bad.push(format!("{path}: {}", lines.join(" ")));
        }
    }
    assert!(
        bad.is_empty(),
        "shipped files with citations:\n{}",
        bad.join("\n")
    );
}

/// AC-3: every exception is named and still needed: the one exempt folder, each format line and
/// each pending file.
#[test]
fn scan_exceptions_stay_named_and_needed() {
    assert_eq!(EXEMPT_PREFIXES, ["templates/bundle-packs/"]);
    let files = shipped();
    for (path, start) in FORMAT_LINES {
        let lines: Vec<&str> = text_of(&files, path)
            .lines()
            .filter(|l| l.trim_start().starts_with(start))
            .collect();
        assert_eq!(lines.len(), 1, "{path}: one line starts with `{start}`");
        assert!(
            !kinds(lines[0]).is_empty(),
            "{path}: the format line needs no exception any more"
        );
    }
    for path in PENDING {
        assert!(!exempt(path), "{path} is pending and exempt at once");
        assert!(
            !citations(path, text_of(&files, path)).is_empty(),
            "`{path}` carries no citation any more: remove it from PENDING"
        );
    }
}

/// AC-4: the installer document says how the MoAI skill and agents are offered: the question
/// both scripts ask and, in its options table, the option that answers it.
#[test]
fn installer_document_names_the_skill_question_and_the_yes_option() {
    const QUESTION: &str = "Install them now? [y/N]";
    let files = shipped();
    for script in ["install/install.sh", "install/install.ps1"] {
        let text = text_of(&files, script);
        assert!(text.contains(QUESTION), "{script} asks `{QUESTION}`");
        assert!(text.contains("--yes"), "{script} takes --yes");
    }
    let doc = text_of(&files, "install/README.md");
    let mut missing = Vec::new();
    if !doc.contains(QUESTION) {
        missing.push(format!("the question `{QUESTION}`"));
    }
    if !doc.lines().any(|l| l.starts_with("| `--yes` (`-yes`")) {
        missing.push("an options table row for `--yes` (`-yes`)".to_string());
    }
    assert!(
        missing.is_empty(),
        "install/README.md lacks {}",
        missing.join(" and ")
    );
}
