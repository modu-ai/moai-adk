//! SPEC-20261126-audit-reads-by-object-id: the audit contract and the lane rules say to read another
//! head from the auditor's own worktree by object id, and no audit text under `skills/`, `agents/`
//! or `docs/` holds a `cd` or a `git -C` that the baseline below does not cover.
//!
//! The texts are read from the checked-in tree and no process is started. The detector matches
//! text only: it does not judge which worktree a command names.

use std::collections::BTreeMap;
use std::fs;
use std::ops::Range;
use std::path::Path;

/// The ban, in the words both rule lines carry.
const CLAUSE: &str = "never `cd` or `git -C` into a worktree that is not your own";

const CONTRACT: &str = "skills/moai/references/audit.md";
const LANE_RULES: &str = "docs/lane-rules.md";

// Paths of texts that exist only inside these tests, spelled in two pieces so that the skill
// reference scan of the tree (`tests/skill_references_scan.rs`) does not read them as references
// to files of the checked-in package.
const MADE_UP: &str = concat!("skills", "/x/zz-audit.md");
const NOT_AN_AUDIT_TEXT: &str = concat!("skills", "/audit/zz.md");

/// The rule files that are audit texts whatever their name.
const NAMED: [&str; 5] = [
    LANE_RULES,
    "docs/leader-rules.md",
    "skills/moai/references/flows/next-flow.md",
    "skills/moai/references/flows/team-messages.md",
    "skills/moai/references/internal/check-rules.md",
];

/// One allowed text: the file, a substring that holds a hit and how often the file has it.
type Entry = (&'static str, &'static str, usize);

// @cc:careful the detector reads text, so a new line that names the auditor's own worktree fails
// too: such a line comes with its row here. Rows of one file must not overlap, and none quotes a
// step number.
const BASELINE: &[Entry] = &[
    // The audit agent works in the audit worktree it was given.
    ("agents/audit.md", "`git -C <worktree>`", 1),
    ("agents/audit.md", "`git -C <worktree> rev-parse HEAD`", 2),
    ("agents/audit.md", "`git -C <worktree> status`", 2),
    (
        "docs/design-review/appendix/operations/audit_done.md",
        "`git -C <자리> checkout -- <경로>`",
        1,
    ),
    (
        LANE_RULES,
        "Never move into or out of a worktree with a shell `cd`, nor with a `cd <worktree> &&` command prefix.",
        1,
    ),
    (
        LANE_RULES,
        "A worktree first entered by shell `cd` is finished as it is",
        1,
    ),
    (LANE_RULES, CLAUSE, 1),
    (CONTRACT, CLAUSE, 1),
];

/// One `cd` or `git -C` of a text: its line (from 1) and its bytes in the whole text.
struct Hit {
    line: usize,
    command: &'static str,
    at: Range<usize>,
}

fn joins_words(byte: u8) -> bool {
    byte.is_ascii_alphanumeric() || byte == b'_' || byte == b'-'
}

/// Where `word` stands in `line` with no ASCII letter, digit, `_` or `-` on either side.
fn words<'a>(line: &'a str, word: &'a str) -> impl Iterator<Item = Range<usize>> + 'a {
    let bytes = line.as_bytes();
    line.match_indices(word)
        .map(|(at, found)| at..at + found.len())
        .filter(move |at| {
            let before = at.start.checked_sub(1).map(|i| bytes[i]);
            let after = bytes.get(at.end).copied();
            !before.is_some_and(joins_words) && !after.is_some_and(joins_words)
        })
}

/// The end of the exact token `-C` that follows a `git` with nothing but options between. Blanks
/// and backticks part the tokens; other punctuation around a token is dropped, a `-` never.
fn option_c_end(after_git: &str) -> Option<usize> {
    let parts = |c: char| c.is_whitespace() || c == '`';
    let punctuation = |c: char| c.is_ascii_punctuation() && c != '-';
    let mut end = 0;
    let mut tokens = after_git.split_inclusive(parts).filter_map(|piece| {
        let start = end;
        end += piece.len();
        let kept = piece.trim_end_matches(parts).trim_end_matches(punctuation);
        let token = kept.trim_start_matches(punctuation);
        (!token.is_empty()).then(|| (token, start + kept.len()))
    });
    while let Some((token, end)) = tokens.next() {
        match token {
            "-C" => return Some(end),
            // `-c` takes the next token as its value.
            "-c" => {
                tokens.next();
            }
            _ if token.starts_with('-') => {}
            _ => return None,
        }
    }
    None
}

/// Every `cd` and `git -C` of a text. A command parted over two lines is not seen.
fn hits(text: &str) -> Vec<Hit> {
    let mut out = Vec::new();
    let mut start = 0;
    for (index, line) in text.split_inclusive('\n').enumerate() {
        let mut found = |command, at: Range<usize>| {
            out.push(Hit {
                line: index + 1,
                command,
                at: start + at.start..start + at.end,
            });
        };
        for cd in words(line, "cd") {
            found("cd", cd);
        }
        for git in words(line, "git") {
            if let Some(end) = option_c_end(&line[git.end..]) {
                found("git -C", git.start..git.end + end);
            }
        }
        start += line.len();
    }
    out
}

/// What the texts (path to content) and a baseline disagree on: an entry that does not occur its
/// recorded number of times or holds no hit, and every hit that lies inside no entry's occurrence.
fn judge(texts: &BTreeMap<String, String>, baseline: &[Entry]) -> Vec<String> {
    let mut problems = Vec::new();
    for &(file, allowed, recorded) in baseline {
        let found = texts
            .get(file)
            .map_or(0, |text| text.matches(allowed).count());
        if found != recorded {
            problems.push(format!(
                "{file}: baseline entry {allowed:?} occurs {found} times, recorded {recorded}"
            ));
        }
        if hits(allowed).is_empty() {
            problems.push(format!(
                "{file}: baseline entry {allowed:?} holds no `cd` or `git -C`"
            ));
        }
    }
    for (file, text) in texts {
        let covered: Vec<Range<usize>> = baseline
            .iter()
            .filter(|entry| entry.0 == file.as_str())
            .flat_map(|entry| text.match_indices(entry.1))
            .map(|(at, found)| at..at + found.len())
            .collect();
        for hit in hits(text) {
            let inside = |c: &Range<usize>| c.start <= hit.at.start && hit.at.end <= c.end;
            if !covered.iter().any(inside) {
                problems.push(format!(
                    "{file}:{}: `{}` is covered by no baseline entry",
                    hit.line, hit.command
                ));
            }
        }
    }
    problems
}

/// Whether a file under the three folders is an audit text: its name says so or the list names it.
fn is_audit_text(path: &str) -> bool {
    let name = path.rsplit('/').next().unwrap_or(path);
    name.to_ascii_lowercase().contains("audit") || NAMED.contains(&path)
}

/// Adds the audit texts below `folder`, whose repository path is `path`. A link is not followed.
fn walk(folder: &Path, path: &str, out: &mut BTreeMap<String, String>) {
    for entry in fs::read_dir(folder).unwrap_or_else(|e| panic!("{path}: {e}")) {
        let entry = entry.unwrap_or_else(|e| panic!("{path}: {e}"));
        let path = format!("{path}/{}", entry.file_name().to_string_lossy());
        let kind = entry.file_type().unwrap_or_else(|e| panic!("{path}: {e}"));
        if kind.is_dir() {
            walk(&entry.path(), &path, out);
        } else if kind.is_file() && is_audit_text(&path) {
            let bytes = fs::read(entry.path()).unwrap_or_else(|e| panic!("{path}: {e}"));
            out.insert(path, String::from_utf8_lossy(&bytes).into_owned());
        }
    }
}

/// The checked-in audit texts. The files that must be among them keep an empty walk from passing.
fn audit_texts() -> BTreeMap<String, String> {
    let mut out = BTreeMap::new();
    for folder in ["skills", "agents", "docs"] {
        walk(
            &Path::new(env!("CARGO_MANIFEST_DIR")).join(folder),
            folder,
            &mut out,
        );
    }
    let contracts = [
        "agents/audit.md",
        CONTRACT,
        "skills/moai/workflows/audit.md",
    ];
    for path in NAMED.into_iter().chain(contracts) {
        assert!(out.contains_key(path), "the walk did not select {path}");
    }
    out
}

/// SPEC-20261126-audit-reads-by-object-id AC-1 (REQ-1, REQ-2)
#[test]
fn ac1_the_contract_and_the_lane_rules_say_to_read_another_head_by_object_id() {
    let wanted = [
        "own worktree",
        "object id",
        "`git show <sha>:<path>`",
        "`git diff <base> <head>`",
        "`git merge-tree`",
        CLAUSE,
    ];
    let texts = audit_texts();
    let lacking: Vec<&str> = [CONTRACT, LANE_RULES]
        .into_iter()
        .filter(|file| {
            let holding = texts[*file]
                .lines()
                .filter(|line| wanted.iter().all(|part| line.contains(part)));
            holding.count() != 1
        })
        .collect();
    assert!(
        lacking.is_empty(),
        "not exactly one line holds {wanted:?} in {lacking:?}"
    );
}

/// SPEC-20261126-audit-reads-by-object-id AC-2 (REQ-3, REQ-4)
#[test]
fn ac2_the_checked_in_audit_texts_hold_only_the_baseline_cd_and_git_c() {
    let problems = judge(&audit_texts(), BASELINE);
    assert!(
        problems.is_empty(),
        "an audit reads another head by object id; remove the command or change the baseline:\n{}",
        problems.join("\n")
    );
}

/// SPEC-20261126-audit-reads-by-object-id AC-3 (REQ-3)
#[test]
fn ac3_a_new_cd_or_git_c_instruction_is_named_with_its_file_and_line() {
    let made_up = MADE_UP;
    let mut texts = BTreeMap::from([
        (CONTRACT.to_string(), audit_texts()[CONTRACT].clone()),
        (made_up.to_string(), "# Made up\n".to_string()),
    ]);
    let before = judge(&texts, BASELINE);
    let planted = [
        ("cd <path> && ls", Some("cd")),
        ("run `git -C <path> status`", Some("git -C")),
        ("git -c a=b -C <path> status", Some("git -C")),
        ("(git --no-pager -C <path> log)", Some("git -C")),
        ("`AUDIT-CLAIM`, `abcd` and `git diff`", None),
        ("a bare `-C`, `git diff -C` and cd-rom", None),
    ];
    let mut expected = Vec::new();
    for (file, text) in &mut texts {
        assert!(text.ends_with('\n'), "{file} ends with a newline");
        for (line, command) in planted {
            text.push_str(line);
            text.push('\n');
            if let Some(command) = command {
                let at = text.lines().count();
                expected.push(format!(
                    "{file}:{at}: `{command}` is covered by no baseline entry"
                ));
            }
        }
    }
    let named: Vec<String> = judge(&texts, BASELINE)
        .into_iter()
        .filter(|problem| !before.contains(problem))
        .collect();
    assert_eq!(named, expected);

    // Coverage is per hit: a second command on an allowed line is named.
    let allowed = "`git -C <worktree> status`";
    let line = format!("Run {allowed} and then cd there.\n");
    assert_eq!(
        judge(
            &BTreeMap::from([(made_up.to_string(), line)]),
            &[(made_up, allowed, 1)]
        ),
        [format!("{made_up}:1: `cd` is covered by no baseline entry")]
    );
}

/// SPEC-20261126-audit-reads-by-object-id AC-4 (REQ-4)
#[test]
fn ac4_a_baseline_count_that_differs_names_the_entry() {
    let file = MADE_UP;
    let allowed = "`git -C <worktree> status`";
    let judged = |text: String, entry: &'static str| {
        judge(
            &BTreeMap::from([(file.to_string(), text)]),
            &[(file, entry, 1)],
        )
    };
    assert_eq!(
        judged(format!("Run {allowed}.\n"), allowed),
        Vec::<String>::new()
    );
    let differs = |found: usize| {
        [format!(
            "{file}: baseline entry {allowed:?} occurs {found} times, recorded 1"
        )]
    };
    assert_eq!(
        judged(format!("{allowed}\n{allowed}\n"), allowed),
        differs(2)
    );
    assert_eq!(judged("Nothing.\n".to_string(), allowed), differs(0));
    assert_eq!(judge(&BTreeMap::new(), &[(file, allowed, 1)]), differs(0));
    assert_eq!(
        judged("abcd\n".to_string(), "abcd"),
        [format!(
            "{file}: baseline entry \"abcd\" holds no `cd` or `git -C`"
        )]
    );
}

/// SPEC-20261126-audit-reads-by-object-id AC-5 (REQ-5)
#[test]
fn ac5_audit_texts_are_chosen_by_file_name_or_by_the_named_list() {
    for path in [
        MADE_UP,
        "docs/x/zz-AUDIT-notes.md",
        "agents/zz_audit_helper.toml",
    ] {
        assert!(is_audit_text(path), "{path} is chosen by its name");
    }
    for path in NAMED {
        assert!(is_audit_text(path), "{path} is chosen by the list");
    }
    for path in [
        "docs/x/zz-rules.md",
        NOT_AN_AUDIT_TEXT,
        "docs/x/lane-rules.md",
    ] {
        assert!(!is_audit_text(path), "{path} is not chosen");
    }
    let texts = audit_texts();
    assert!(!texts.contains_key("skills/moai/SKILL.md"));
    assert!(texts.keys().all(|path| is_audit_text(path)));
}
