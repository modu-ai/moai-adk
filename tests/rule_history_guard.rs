//! SPEC-20261119-rule-prune-leader (card C-103, last slice): `references/rule-history.md` is a
//! record the other rule files point to, so three things in it must not drift quietly:
//! every row of a mechanical inventory names the test files that read its line (or says
//! `text only`) and its section heading names the tip it was written from; every test file the
//! record names exists; and no entry of a history section is still word for word in the live rule
//! file it was moved from (a moved rule that came back would be counted twice). A sentence that
//! was merged into another file's copy may stay there: only its source file is checked.
//!
//! The row text of a mechanical inventory is not compared with the live files: other cards edit
//! the workflows all the time, and a text match would block their merges (SPEC decision D-3).
//! The tests read files of this checkout only; nothing is run.

use std::{fs, path::PathBuf};

const HISTORY_FILE: &str = "skills/moai/references/rule-history.md";
const LIVE_FILES: [&str; 4] = [
    "skills/moai/workflows/mode-leader.md",
    "skills/moai/workflows/mode-agent.md",
    "skills/moai/workflows/audit.md",
    "skills/moai/references/internal/core-rules.md",
];

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The `## ` sections of a document: the heading text and the lines below it.
fn sections(text: &str) -> Vec<(String, Vec<String>)> {
    let mut out: Vec<(String, Vec<String>)> = Vec::new();
    for line in text.lines() {
        if let Some(heading) = line.strip_prefix("## ") {
            out.push((heading.trim().to_string(), Vec::new()));
        } else if let Some((_, body)) = out.last_mut() {
            body.push(line.to_string());
        }
    }
    out
}

/// `name.rs` or `name.py`, with at most one folder in front (`common/git.rs`).
fn is_test_file_name(name: &str) -> bool {
    let name = match name.split_once('/') {
        Some((dir, file))
            if !dir.is_empty() && dir.chars().all(|c| c.is_ascii_lowercase() || c == '_') =>
        {
            file
        }
        Some(_) => return false,
        None => name,
    };
    let Some((stem, ext)) = name.rsplit_once('.') else {
        return false;
    };
    !stem.is_empty()
        && stem.chars().all(|c| c.is_ascii_alphanumeric() || c == '_')
        && (ext == "rs" || ext == "py")
}

/// A heading of a mechanical inventory ends with ` @ <short sha of the tip it was written from>`.
fn heading_has_a_tip(heading: &str) -> bool {
    heading.rsplit_once(" @ ").is_some_and(|(_, sha)| {
        (7..=40).contains(&sha.len())
            && sha
                .chars()
                .all(|c| c.is_ascii_hexdigit() && !c.is_ascii_uppercase())
    })
}

/// The enforcer field of an inventory row: the text after the last ` · `.
fn row_field(row: &str) -> Option<&str> {
    row.rsplit_once(" · ").map(|(_, field)| field.trim())
}

/// The names of a `pins: a.rs, b.rs (+3 more)` field; `None` when it is not one.
fn pin_names(field: &str) -> Option<Vec<&str>> {
    let names = field.strip_prefix("pins: ")?;
    let names = match names.rsplit_once(" (+") {
        Some((head, tail))
            if tail
                .strip_suffix(" more)")
                .is_some_and(|n| !n.is_empty() && n.chars().all(|c| c.is_ascii_digit())) =>
        {
            head
        }
        _ => names,
    };
    Some(names.split(", ").collect())
}

fn inventory_problems(text: &str) -> Vec<String> {
    let mut problems = Vec::new();
    for (heading, body) in sections(text) {
        if !heading.starts_with("Mechanical inventory — ") {
            continue;
        }
        if !heading_has_a_tip(&heading) {
            problems.push(format!(
                "heading `{heading}` does not end with ` @ <short sha>`"
            ));
        }
        for row in body.iter().filter(|l| l.starts_with("- ")) {
            let ok = match row_field(row) {
                Some("text only") => true,
                Some(field) => pin_names(field)
                    .is_some_and(|names| names.iter().all(|name| is_test_file_name(name))),
                None => false,
            };
            if !ok {
                let head: String = row.chars().take(70).collect();
                problems.push(format!(
                    "row `{head}…` does not end with `· pins: <test files>` or `· text only`"
                ));
            }
        }
    }
    problems
}

/// Test file names the record writes as an enforcer: the `pins:` fields of inventory rows and
/// the backticked `*.rs` / `*.py` names of the live table (a name with a `/` is a path).
fn named_files(text: &str) -> Vec<String> {
    let mut names = Vec::new();
    for (heading, body) in sections(text) {
        if heading.starts_with("Mechanical inventory — ") {
            for row in body.iter().filter(|l| l.starts_with("- ")) {
                if let Some(pins) = row_field(row).and_then(pin_names) {
                    names.extend(pins.into_iter().map(str::to_string));
                }
            }
        } else if heading.starts_with("Live inventory — ") {
            for row in body.iter().filter(|l| l.starts_with('|')) {
                for (i, piece) in row.split('`').enumerate() {
                    let name = piece.trim();
                    let last = name.rsplit('/').next().unwrap_or(name);
                    if i % 2 == 1 && is_test_file_name(last) {
                        names.push(name.to_string());
                    }
                }
            }
        }
    }
    names
}

fn missing_files(text: &str, exists: impl Fn(&str) -> bool) -> Vec<String> {
    let mut missing: Vec<String> = named_files(text)
        .into_iter()
        .filter(|n| !exists(n))
        .collect();
    missing.sort();
    missing.dedup();
    missing
}

fn squash(s: &str) -> String {
    s.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// The moved text of every history entry: `N. text`, `N (tail). text` or `- text`, without a
/// trailing `*(note)*` and without a ` · reason: …` part.
fn history_entries(text: &str) -> Vec<(String, String)> {
    let mut out = Vec::new();
    for (heading, body) in sections(text) {
        if !heading.starts_with("History — moved from ") {
            continue;
        }
        for line in body {
            let line = line.trim();
            let entry = if let Some(rest) = line.strip_prefix("- ") {
                rest
            } else if let Some((num, rest)) = line.split_once(". ") {
                let num_ok = num.chars().next().is_some_and(|c| c.is_ascii_digit())
                    && num
                        .chars()
                        .all(|c| c.is_ascii_digit() || " ()tail".contains(c));
                if !num_ok {
                    continue;
                }
                rest
            } else {
                continue;
            };
            let entry = entry.split(" · reason:").next().unwrap_or(entry);
            let entry = match entry.rfind(" *(") {
                Some(at) if entry.ends_with(")*") => &entry[..at],
                _ => entry,
            };
            out.push((heading.clone(), squash(entry)));
        }
    }
    out
}

/// The live rule file a history heading says its entries were moved from.
fn source_file(heading: &str) -> Option<&'static str> {
    let from = heading.strip_prefix("History — moved from ")?;
    LIVE_FILES.into_iter().find(|f| f.ends_with(from))
}

fn live_entries<'a>(entries: &'a [(String, String)], live: &str) -> Vec<&'a str> {
    let live = squash(live);
    entries
        .iter()
        .filter(|(_, text)| text.len() >= 20 && live.contains(text.as_str()))
        .map(|(_, text)| text.as_str())
        .collect()
}

#[test]
fn every_inventory_row_names_its_enforcers_or_text_only() {
    let problems = inventory_problems(&read(HISTORY_FILE));
    assert!(
        problems.is_empty(),
        "rule-history.md inventories: {problems:#?}"
    );
}

#[test]
fn every_test_file_named_in_the_history_exists() {
    let text = read(HISTORY_FILE);
    assert!(
        !named_files(&text).is_empty(),
        "the record names no test file at all"
    );
    // a pin names a file under `tests/`; the live table may also name a path from the root
    let missing = missing_files(&text, |name| {
        root().join("tests").join(name).exists() || root().join(name).exists()
    });
    assert!(
        missing.is_empty(),
        "rule-history.md names files that do not exist: {missing:?}"
    );
}

#[test]
fn no_history_entry_is_still_live() {
    let text = read(HISTORY_FILE);
    let entries = history_entries(&text);
    assert!(
        entries.len() >= 12,
        "the history sections hold {} entries",
        entries.len()
    );
    for rel in LIVE_FILES {
        let mine: Vec<(String, String)> = entries
            .iter()
            .filter(|(heading, _)| source_file(heading) == Some(rel))
            .cloned()
            .collect();
        // a file nothing was moved from has no entry to compare
        let back = live_entries(&mine, &read(rel));
        assert!(
            back.is_empty(),
            "{rel} holds history entries word for word: {back:#?}"
        );
    }
    for (heading, _) in &entries {
        assert!(
            source_file(heading).is_some(),
            "history heading `{heading}` names no live rule file"
        );
    }
}

#[test]
fn the_guard_flags_a_row_without_enforcers_and_a_heading_without_a_tip() {
    let text = "## Mechanical inventory — workflows/x.md\n\n\
                - L1 #1 — A rule · pins: a_test.rs, b_test.py\n\
                - L2 #1 — Another rule · text only\n\
                - L3 #1 — A third rule without a field\n\
                - L4 #1 — A fourth · pins: not a file name\n\
                - L5 #1 — A fifth · pins: common/git.rs, a_test.rs (+12 more)\n\
                - L6 #1 — A sixth · pins: a_test.rs (+more)\n";
    let problems = inventory_problems(text);
    assert_eq!(problems.len(), 4, "{problems:#?}");
    assert!(problems[0].contains("heading"), "{problems:#?}");
    let fixed = text.replace("x.md", "x.md @ 190de2d8d7");
    assert_eq!(inventory_problems(&fixed).len(), 3);
    assert_eq!(
        pin_names("pins: a_test.rs, common/git.rs (+3 more)"),
        Some(vec!["a_test.rs", "common/git.rs"])
    );
}

#[test]
fn the_guard_flags_a_named_file_that_does_not_exist() {
    let text = "## Live inventory — core-rules.md (1 rules)\n\n| # | R | S | Enforcer |\n| --- | --- | --- | --- |\n\
                | 3 | r | #1 | `here_test.rs`; `src/gone.rs` |\n\n\
                ## Mechanical inventory — workflows/x.md @ 190de2d8d7\n\n\
                - L1 #1 — A rule · pins: here_test.rs, gone_test.py\n";
    assert_eq!(named_files(text).len(), 4);
    let missing = missing_files(text, |name| name == "here_test.rs");
    assert_eq!(missing, ["gone_test.py", "src/gone.rs"]);
}

#[test]
fn the_guard_flags_a_history_entry_that_is_still_live() {
    let text = "## History — moved from core-rules.md\n\n\
                1. Never predict durations in plans or reports; give order and priority.\n\
                7 (tail). Code comments follow the project's language setting. *(a note)*\n\n\
                ## History — moved from workflows/x.md\n\n\
                - A moved sentence that is long enough to count. · reason: one source\n";
    let entries = history_entries(text);
    assert_eq!(entries.len(), 3, "{entries:#?}");
    assert_eq!(
        entries[1].1,
        "Code comments follow the project's language setting."
    );
    assert_eq!(
        entries[2].1,
        "A moved sentence that is long enough to count."
    );
    let live = "A rule.\nNever predict durations in plans or reports;\ngive order and priority.\n";
    assert_eq!(
        live_entries(&entries, live),
        ["Never predict durations in plans or reports; give order and priority."]
    );
}
