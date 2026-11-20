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
const PENDING: [&str; 17] = [
    "install/README.md",
    "install/install.ps1",
    "install/install.sh",
    "install/moai-switch.sh",
    "skills/moai/SKILL.md",
    "skills/moai/references/adapters/orca.md",
    "skills/moai/references/internal/interview.md",
    "skills/moai/references/internal/questions.md",
    "skills/moai/references/internal/self-check.md",
    "skills/moai/references/internal/versioning.md",
    "skills/moai/references/rule-history.md",
    "skills/moai/references/source-manifest.json",
    "skills/moai/workflows/project.md",
    "skills/moai/workflows/run.md",
    "skills/moai/workflows/update.md",
    "templates/github/ko/ISSUE_TEMPLATE/moai-investigation.yml",
    "templates/github/manifest.json",
];

/// The files this change cleaned; each has a section in the rationale document.
const CLEANED: [&str; 14] = [
    "agents/audit.md",
    "skills/moai/references/adapters/aside.md",
    "skills/moai/references/internal/build-folder.md",
    "skills/moai/references/internal/commit-trailers.md",
    "skills/moai/references/internal/git-safety.md",
    "skills/moai/references/internal/review.md",
    "skills/moai/workflows/audit.md",
    "skills/moai/workflows/mode-agent.md",
    "skills/moai/workflows/mode-leader.md",
    "skills/moai/workflows/optimize.md",
    "skills/moai/workflows/read.md",
    "skills/moai/workflows/setup.md",
    "skills/moai/workflows/sync.md",
    "skills/moai/workflows/verify.md",
];

/// The rationale document, relative to the checkout.
const RATIONALE: &str = "docs/skill-text-rationale.md";

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

/// Every kind of citation on one line: `card` (C-12), `number` (#2345), `decision` (a user or
/// leader decision followed by a date or an id) and `tool` (a development script name).
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
            if (1..=4).contains(&digits) && !next.is_ascii_digit() {
                out.push("card");
            }
        }
        if c == '#' && !is_word(before) && before != '&' {
            let digits = line[at + 1..]
                .chars()
                .take_while(char::is_ascii_digit)
                .count();
            let next = line[at + 1 + digits..].chars().next().unwrap_or(' ');
            if (3..=5).contains(&digits) && !next.is_ascii_digit() {
                out.push("number");
            }
        }
    }
    // ASCII lower case keeps every byte position, so an index into `low` is one into `line`.
    let low = line.to_ascii_lowercase();
    for word in ["issue", "issues", "pr", "prs"] {
        for (at, _) in low.match_indices(word) {
            let word_before = low[..at].chars().next_back().is_some_and(is_word);
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
                if (3..=5).contains(&digits) && !next.is_ascii_digit() {
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
    for (at, _) in low.match_indices("decision ") {
        let mut it = low[at + 9..].chars();
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
        // Found by the audit: forms without a hash, with a capital, with a possessive, a four
        // digit card, and a character whose lower case is longer than itself.
        ("Issue 2619 says so", "number"),
        ("PR 3234 was merged", "number"),
        ("see Issue#2619", "number"),
        ("Decision D7 of the review", "decision"),
        ("the user's decision 2026-11-17", "decision"),
        ("card C-1234 is done", "card"),
        ("\u{130} the user decision 2026-11-16", "decision"),
    ] {
        assert!(
            kinds(line).contains(&kind),
            "`{line}` should carry a {kind}"
        );
    }
    for line in [
        "A line starting with `LANE-BLOCKED kind` is a machine marker; `LANE-READY-INTEGRATION` too.",
        "ABC-12 is not a card and neither is the CI-3 job",
        "heading &#8212; and the colour #123456 and anchor #12",
        "the user decision is recorded with its answer",
        "ask the user for a decision before the pick",
        "a card id looks like `C-<number>` in the text",
        "issue 12 of the series, and PR 12 of the other",
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
    assert!(
        bad.is_empty(),
        "shipped files with citations:\n{}",
        bad.join("\n")
    );
}

/// AC-1: the pending list is sorted, names shipped files, and only files that still carry
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
        assert_eq!(
            citations(text),
            Vec::new(),
            "`{path}` still carries citations"
        );
    }
}

/// REQ-6: the two lane workflows state the local finish line once, and the verify workflow names
/// the release pull request's CI; the rule carries no citation label.
#[test]
fn finish_line_rule_is_stated_once_in_each_lane_workflow() {
    let files = shipped();
    let text = |path: &str| {
        files
            .iter()
            .find(|(p, _)| p == path)
            .map(|(_, t)| t.as_str())
            .unwrap_or_else(|| panic!("`{path}` is not a shipped file"))
    };
    for path in [
        "skills/moai/workflows/mode-agent.md",
        "skills/moai/workflows/mode-leader.md",
    ] {
        let t = text(path);
        assert_eq!(t.matches("Local finish line:").count(), 1, "{path}");
        for phrase in [
            "makes the merge of its head onto the current tip of the LOCAL integration branch",
            "check-affected command of the tool built in that worktree",
            "`MERGE-READY <card> merge=<40-hex> base=<40-hex> checks=<one line>`",
            "under a lock (`~/.moai/locks/<project>-<branch>.lock`)",
            "`git merge --ff-only`",
            "`git -C <primary> merge --ff-only refs/moai/lane-merge`",
            "where the host isolates a lane session to its worktree",
            "`MERGED-LOCAL`",
            "from the verified tip of the local integration branch, not from the remote one",
            "how far the branch is behind (↓) its upstream",
            "nobody opens a pull request per card, per bundle or per integration round",
            "`release/v{head}.{yearweek}.{build}`",
            "only for a real deployment",
            "a broken tip with a revert commit",
            "is closed with `card done` on its local landing evidence, not on a pull request",
        ] {
            assert!(t.contains(phrase), "{path} must say `{phrase}`");
        }
    }
    for phrase in [
        "`MERGED-LOCAL <card> <new tip> head=<head>`",
        "at audit tier 2 the card in `review` with its pushed head",
        "until `card done` accepts a card with no linked merged pull request",
    ] {
        assert!(
            text("skills/moai/workflows/mode-agent.md").contains(phrase),
            "mode-agent.md hand-off must say `{phrase}`"
        );
    }
    assert!(
        text("skills/moai/workflows/verify.md")
            .contains("the release pull request from the local integration branch to the remote one runs the full CI on Linux, macOS and Windows"),
        "verify.md names the release pull request's CI"
    );
}
