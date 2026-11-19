//! SPEC-20261120-skill-text-citations: the files that ship to users carry no card number, no
//! Issue or pull-request number, no citation of a user or leader decision and no name of a moai-adk
//! development script. The scan reads the shipped files of this checkout (`skills/`, `agents/`,
//! `templates/`, `install/`); nothing is run and no HOME is touched.
//!
//! `PENDING` lists the shipped files that still carry such text. It can only shrink: a file that
//! was cleaned must leave the list in the same change, so the scan keeps the rest clean.

use std::{
    fs,
    path::{Path, PathBuf},
};

/// The shipped folders, relative to the checkout.
const SHIPPED: [&str; 4] = ["skills", "agents", "templates", "install"];

/// Shipped files that still carry citations, sorted. Remove an entry when its file is cleaned.
const PENDING: [&str; 18] = [
    "agents/audit.md",
    "install/README.md",
    "install/install.ps1",
    "install/install.sh",
    "install/moai-switch.sh",
    "skills/moai/references/internal/interview.md",
    "skills/moai/references/internal/questions.md",
    "skills/moai/references/internal/versioning.md",
    "skills/moai/references/rule-history.md",
    "skills/moai/references/source-manifest.json",
    "skills/moai/workflows/audit.md",
    "skills/moai/workflows/mode-agent.md",
    "skills/moai/workflows/mode-leader.md",
    "skills/moai/workflows/project.md",
    "skills/moai/workflows/update.md",
    "skills/moai/workflows/verify.md",
    "templates/github/ko/ISSUE_TEMPLATE/moai-investigation.yml",
    "templates/github/manifest.json",
];

/// The files this change cleaned; each has a section in the rationale document.
const CLEANED: [&str; 13] = [
    "skills/moai/SKILL.md",
    "skills/moai/references/adapters/aside.md",
    "skills/moai/references/adapters/orca.md",
    "skills/moai/references/internal/build-folder.md",
    "skills/moai/references/internal/commit-trailers.md",
    "skills/moai/references/internal/git-safety.md",
    "skills/moai/references/internal/review.md",
    "skills/moai/references/internal/self-check.md",
    "skills/moai/workflows/optimize.md",
    "skills/moai/workflows/read.md",
    "skills/moai/workflows/run.md",
    "skills/moai/workflows/setup.md",
    "skills/moai/workflows/sync.md",
];

/// The rationale document, relative to the checkout.
const RATIONALE: &str = "docs/skill-text-rationale.md";

/// Names of moai-adk's own development scripts and repository, which a user does not have.
const DEV_WORDS: [&str; 9] = [
    "moai-adk repository",
    "local_check",
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

/// Every kind of citation on one line: `card` (C-12), `number` (#2345), `decision` (a user or
/// leader decision followed by a date or an id) and `tool` (a development script name).
fn kinds(line: &str) -> Vec<&'static str> {
    let mut out = Vec::new();
    let chars: Vec<(usize, char)> = line.char_indices().collect();
    for (k, &(at, c)) in chars.iter().enumerate() {
        let before = if k == 0 { ' ' } else { chars[k - 1].1 };
        if c == 'C' && line[at..].starts_with("C-") && !is_word(before) && before != '-' {
            let digits = line[at + 2..].chars().take_while(char::is_ascii_digit).count();
            let next = line[at + 2 + digits..].chars().next().unwrap_or(' ');
            if (1..=3).contains(&digits) && !next.is_ascii_digit() {
                out.push("card");
            }
        }
        if c == '#' && !is_word(before) && before != '&' {
            let digits = line[at + 1..].chars().take_while(char::is_ascii_digit).count();
            let next = line[at + 1 + digits..].chars().next().unwrap_or(' ');
            if (3..=5).contains(&digits) && !next.is_ascii_digit() {
                out.push("number");
            }
        }
    }
    let low = line.to_lowercase();
    for phrase in ["user decision", "leader decision", "user selection"] {
        for (at, _) in low.match_indices(phrase) {
            let rest = after(line, at + phrase.len());
            let rest = rest.strip_prefix(',').map_or(rest, str::trim_start);
            let mut it = rest.chars();
            let (first, second) = (it.next().unwrap_or(' '), it.next().unwrap_or(' '));
            if first.is_ascii_digit()
                || (first.is_ascii_uppercase() && (second.is_ascii_digit() || second == '-'))
            {
                out.push("decision");
            }
        }
    }
    for phrase in ["사용자 결정", "리더 결정"] {
        if line.contains(phrase) {
            out.push("decision");
        }
    }
    for (at, _) in line.match_indices("decision ") {
        let mut it = line[at + 9..].chars();
        if matches!(it.next(), Some('D' | 'U')) && it.next().is_some_and(|c| c.is_ascii_digit()) {
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

/// Every citation of a text as `(line number, kind)`.
fn citations(text: &str) -> Vec<(usize, &'static str)> {
    text.lines()
        .enumerate()
        .flat_map(|(i, l)| kinds(l).into_iter().map(move |k| (i + 1, k)))
        .collect()
}

/// Every shipped text file as `(path relative to the checkout with `/`, text)`, sorted.
fn shipped() -> Vec<(String, String)> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                walk(&path, out);
            } else {
                out.push(path);
            }
        }
    }
    let mut files = Vec::new();
    for top in SHIPPED {
        walk(&root().join(top), &mut files);
    }
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

/// AC-1: the scan sees each kind of citation and leaves the machine markers and plain words alone.
#[test]
fn scan_flags_each_kind_and_leaves_markers_alone() {
    for (line, kind) in [
        ("Merge train (card C-51 and more)", "card"),
        ("(see Issue #2619)", "number"),
        ("PR #3234 merged", "number"),
        ("(user decision 2026-11-16)", "decision"),
        ("integration gate of user decision U6", "decision"),
        ("a leader decision 2026-11-17 says", "decision"),
        ("decision D7 of the review", "decision"),
        ("사용자 결정 2026-11-17", "decision"),
        ("in the moai-adk repository run it", "tool"),
        ("run `tools/self_check.sh` first", "tool"),
        ("the M-Flow bundle", "tool"),
    ] {
        assert!(kinds(line).contains(&kind), "`{line}` should carry a {kind}");
    }
    for line in [
        "A line starting with `LANE-BLOCKED kind` is a machine marker; `LANE-READY-INTEGRATION` too.",
        "ABC-12 is not a card and neither is the CI-3 job",
        "heading &#8212; and the colour #123456 and anchor #12",
        "the user decision is recorded with its answer",
        "ask the user for a decision before the pick",
        "a card id looks like `C-<number>` in the text",
    ] {
        assert!(kinds(line).is_empty(), "`{line}` should be clean: {:?}", kinds(line));
    }
}

/// AC-1: no shipped file outside the pending list carries a citation.
#[test]
fn scan_shipped_files_outside_the_pending_list_are_clean() {
    let mut bad = Vec::new();
    for (path, text) in shipped() {
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
    assert!(bad.is_empty(), "shipped files with citations:\n{}", bad.join("\n"));
}

/// AC-1: the pending list is sorted, names shipped files, and only files that still carry
/// citations (a cleaned file has to leave it).
#[test]
fn scan_pending_list_only_shrinks() {
    let mut sorted = PENDING.to_vec();
    sorted.sort_unstable();
    sorted.dedup();
    assert_eq!(sorted, PENDING.to_vec(), "PENDING is sorted and has no duplicates");
    let files = shipped();
    for path in PENDING {
        let text = files
            .iter()
            .find(|(p, _)| p == path)
            .unwrap_or_else(|| panic!("PENDING names `{path}`, which is not a shipped file"));
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
    let files = shipped();
    for path in CLEANED {
        assert!(!PENDING.contains(&path), "`{path}` is cleaned, not pending");
        let (_, text) = files
            .iter()
            .find(|(p, _)| p == path)
            .unwrap_or_else(|| panic!("`{path}` is not a shipped file"));
        assert_eq!(citations(text), Vec::new(), "`{path}` still carries citations");
    }
}
