//! SPEC-20261120-question-bank-generic: the interview question bank is generic. It has no
//! decision-model row, asks the writing language as three rows, offers the versioning schemes
//! in the decided order after the work flow, and the three internal reference files name no
//! development tool of this repository, no card number and no Issue number. The tests read the
//! shipped text only; no host, no HOME and no command runs here.

use std::fs;

const BANK_FILES: [&str; 3] = [
    "skills/moai/references/internal/interview.md",
    "skills/moai/references/internal/questions.md",
    "skills/moai/references/internal/versioning.md",
];

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn interview() -> String {
    read("skills/moai/references/internal/interview.md")
}

/// The question bank's table rows: the cells of every line that starts with "| `".
fn rows() -> Vec<Vec<String>> {
    let text = interview();
    let bank = text
        .split("## Question bank")
        .nth(1)
        .expect("interview.md has a question bank");
    bank.lines()
        .filter(|l| l.starts_with("| `"))
        .map(|l| {
            l.trim_matches('|')
                .split(" | ")
                .map(|c| c.trim().to_string())
                .collect()
        })
        .collect()
}

fn ids() -> Vec<String> {
    rows()
        .iter()
        .map(|r| r[0].trim_matches('`').to_string())
        .collect()
}

fn index(id: &str) -> usize {
    ids()
        .iter()
        .position(|i| i == id)
        .unwrap_or_else(|| panic!("the bank has no `{id}` row"))
}

#[test]
fn the_bank_has_no_decision_model() {
    assert!(
        !ids().iter().any(|i| i == "decision_model"),
        "the bank still has a decision_model row"
    );
    let text = interview();
    for word in [
        "decision_model",
        "Jev",
        "OpenAI",
        "key question",
        "decision model",
    ] {
        assert!(!text.contains(word), "interview.md still says {word:?}");
    }
}

#[test]
fn the_bank_asks_three_language_rows_and_fixes_the_rest_to_english() {
    for id in [
        "language_conversation",
        "language_commits",
        "language_documents",
    ] {
        let row = &rows()[index(id)];
        assert_eq!(row.len(), 5, "five cells: {row:?}");
        assert!(
            row[4].contains("profile `language`"),
            "`{id}` must fill the profile language: {row:?}"
        );
    }
    let text = interview();
    let sentence = "Every other language item is English and is not asked";
    assert!(text.contains(sentence), "interview.md lacks: {sentence}");
    // The old per-item rows are not rows of the bank.
    for old in [
        "agent_messages",
        "skill_instructions",
        "code_comments",
        "instruction_files",
        "issue_pr_records",
    ] {
        assert!(
            !ids().iter().any(|i| i == old),
            "`{old}` must not be a bank row"
        );
    }
}

#[test]
fn versioning_options_follow_the_decided_order_and_come_after_the_work_flow() {
    let row = &rows()[index("versioning")];
    let options = &row[3];
    let keep = options.find("keep the existing one").expect("keep option");
    let semver = options
        .find("SemVer (recommended)")
        .expect("SemVer recommended");
    assert!(keep < semver, "a detected scheme comes first: {options}");
    let after_semver = &options[semver..];
    let headver = after_semver.find("HeadVer").expect("HeadVer option");
    let calver = after_semver.find("CalVer").expect("CalVer option");
    assert!(headver < calver, "HeadVer before CalVer: {options}");
    let text = interview();
    let sentence = "when the work flow is m-flow, HeadVer is labelled \"suggested with m-flow\" and SemVer stays first and recommended";
    assert!(text.contains(sentence), "interview.md lacks: {sentence}");
    for later in ["versioning", "versioning_format", "versioning_head"] {
        assert!(
            index("work_flow") < index(later),
            "the `work_flow` row must come before `{later}`"
        );
    }
}

/// Shapes the shipped reference files must not carry (the repository's own tools, card and
/// Issue numbers, dated user-decision citations, decision providers, Kiro and Cursor).
fn forbidden_in(text: &str) -> Vec<String> {
    let mut hits = Vec::new();
    for word in [
        "tools/local_check.sh",
        "check-guard",
        "SKILL.md case",
        "SKILL.md sorts",
        "Kiro",
        "Cursor",
        "Jev",
        "OpenAI",
        "decision_model",
        "user decision 20",
        "decision 20",
        "Issue #",
    ] {
        if text.contains(word) {
            hits.push(word.to_string());
        }
    }
    let bytes = text.as_bytes();
    for (i, _) in text.match_indices("C-") {
        let digits = bytes[i + 2..]
            .iter()
            .take_while(|b| b.is_ascii_digit())
            .count();
        let before_ok = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
        if digits > 0 && before_ok {
            hits.push(text[i..i + 2 + digits].to_string());
        }
    }
    for (i, _) in text.match_indices('#') {
        let digits = bytes[i + 1..]
            .iter()
            .take_while(|b| b.is_ascii_digit())
            .count();
        if digits >= 3 {
            hits.push(text[i..i + 1 + digits].to_string());
        }
    }
    hits
}

#[test]
fn the_bank_files_name_no_card_issue_or_repository_tool() {
    for rel in BANK_FILES {
        let hits = forbidden_in(&read(rel));
        assert!(hits.is_empty(), "{rel} still names {hits:?}");
    }
    // The autonomy row says what is asked first in its own words.
    let row = &rows()[index("autonomy")];
    assert!(
        !row[2].contains("SKILL.md") && !row[3].contains("SKILL.md"),
        "the autonomy row points at another file's case number: {row:?}"
    );
    assert!(
        row[3].contains("cannot be undone"),
        "the autonomy row must say what is asked first: {row:?}"
    );
}
