//! SPEC-20261125-skill-entry-order (AC-1 to AC-5): the skill entry gives one numbered start
//! order, the decision classification stands once in `internal/decide`, the core rules carry
//! their own text, and the entry files carry no Issue word, `semi` name, card number or Go
//! PATH sentence. Reads the package files and the SPEC only: no process is started.
use std::{fs, path::Path};

fn package(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn spec() -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join(".moai/specs/SPEC-20261125-skill-entry-order/spec.md"),
    )
    .unwrap()
}

/// True when `text` holds `word` as a whole word (letters and digits on both sides break it).
fn has_word(text: &str, word: &str) -> bool {
    let bytes = text.as_bytes();
    text.match_indices(word).any(|(at, _)| {
        let before = at == 0 || !bytes[at - 1].is_ascii_alphanumeric();
        let end = at + word.len();
        let after = end >= bytes.len() || !bytes[end].is_ascii_alphanumeric();
        before && after
    })
}

/// True when `text` holds a card number such as `C-520`.
fn has_card_number(text: &str) -> bool {
    text.match_indices("C-").any(|(at, _)| {
        let before = at == 0 || !text.as_bytes()[at - 1].is_ascii_alphanumeric();
        let digit = text.as_bytes().get(at + 2).is_some_and(u8::is_ascii_digit);
        before && digit
    })
}

/// AC-1 (REQ-1): the entry has a numbered start order that names the route table, the chains
/// table, the core rules and the chosen route text, in that order, each with its path.
#[test]
fn ac1_the_start_order_is_numbered_and_names_four_stops_in_order() {
    let skill = package("SKILL.md");
    let at = skill
        .find("Start order")
        .expect("SKILL.md has no `Start order` heading or line");
    let numbered: Vec<&str> = skill[at..]
        .lines()
        .skip(1)
        .take_while(|l| l.trim().is_empty() || l.starts_with(|c: char| c.is_ascii_digit()))
        .filter(|l| !l.trim().is_empty())
        .collect();
    assert_eq!(numbered.len(), 4, "{numbered:?}");
    for (n, (line, path)) in numbered
        .iter()
        .zip([
            "references/routes.json",
            "references/chains/index.md",
            "references/internal/core-rules.md",
            "workflows/",
        ])
        .enumerate()
    {
        assert!(
            line.starts_with(&format!("{}. ", n + 1)),
            "step {} is not numbered: {line}",
            n + 1
        );
        assert!(
            line.contains(path),
            "step {} must name {path}: {line}",
            n + 1
        );
    }
}

/// AC-2 (REQ-2, REQ-3): `decide.md` holds the four cases as one table with the five HOLD
/// categories, the `guarded` settings and the release-tag rule once, and `SKILL.md` restates none.
#[test]
fn ac2_the_decision_table_stands_once_in_decide_md() {
    let decide = package("references/internal/decide.md");
    for n in 1..=4 {
        let row = format!("| {n} |");
        assert_eq!(
            decide.matches(&row).count(),
            1,
            "decide.md must hold row {n} once"
        );
    }
    for phrase in [
        "money",
        "legal",
        "external_contract",
        "secret",
        "real_data",
        "guarded",
        "pushing a tag or publishing a release",
        "`full` included",
        "no deadline",
    ] {
        assert!(decide.contains(phrase), "decide.md lost `{phrase}`");
    }
    let skill = package("SKILL.md");
    for restated in [
        "external_contract",
        "real_data",
        "`guarded`",
        "case (2)",
        "case (4)",
        "no deadline",
    ] {
        assert!(
            !skill.contains(restated),
            "SKILL.md restates the classification: `{restated}`"
        );
    }
    assert!(
        skill.contains("`internal/decide`"),
        "SKILL.md must point to `internal/decide`"
    );
}

/// AC-3 (REQ-4): every row of the sentence table pairs an old sentence with a new place, and a
/// new place names a file of the package.
#[test]
fn ac3_the_sentence_table_pairs_every_old_sentence_with_a_new_place() {
    let text = spec();
    let at = text
        .find("## Sentence table")
        .expect("the SPEC has a sentence table");
    let rows: Vec<Vec<&str>> = text[at..]
        .lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| #") && !l.starts_with("| ---"))
        .map(|l| l.trim_matches('|').split('|').map(str::trim).collect())
        .collect();
    assert!(rows.len() >= 14, "{} rows", rows.len());
    for row in &rows {
        assert_eq!(row.len(), 3, "{row:?}");
        assert!(!row[1].is_empty() && !row[2].is_empty(), "{row:?}");
        assert!(
            ["decide.md", "SKILL.md", "internal/decide"]
                .iter()
                .any(|file| row[2].contains(file)),
            "the new place of {} names no file: {}",
            row[0],
            row[2]
        );
    }
}

/// AC-4 (REQ-5): each core rule has a text of its own and none points to the history file or a
/// history entry.
#[test]
fn ac4_every_core_rule_has_its_own_text() {
    let rules = package("references/internal/core-rules.md");
    let mut seen = 0;
    for line in rules.lines() {
        let Some((n, rest)) = line.split_once(". ") else {
            continue;
        };
        if n.parse::<usize>().is_err() {
            continue;
        }
        seen += 1;
        assert!(rest.len() >= 40, "rule {n} is only a stub: {rest}");
        assert!(
            !rest.contains("history"),
            "rule {n} points to history: {rest}"
        );
    }
    assert!(seen >= 25, "{seen} rules");
    assert!(
        !rules.contains("rule-history"),
        "core-rules.md names the history file"
    );
}

/// AC-5 (REQ-6, REQ-7): no Issue word, `semi` name, card number or Go PATH sentence in the
/// entry files, and every pointer to the classification names `internal/decide`.
#[test]
fn ac5_the_entry_files_carry_no_old_words_and_point_to_decide() {
    for file in [
        "SKILL.md",
        "references/internal/decide.md",
        "references/internal/questions.md",
    ] {
        let text = package(file);
        for word in ["Issue", "Issues", "semi"] {
            assert!(!has_word(&text, word), "{file} holds the word `{word}`");
        }
        assert!(!has_card_number(&text), "{file} holds a card number");
        assert!(
            !text.contains("earlier Go"),
            "{file} holds the Go PATH sentence"
        );
    }
    for file in [
        "references/internal/versioning.md",
        "references/internal/interview.md",
        "references/internal/questions.md",
    ] {
        let text = package(file);
        for old in [
            "`SKILL.md` case (2)",
            "SKILL.md calls case (2)",
            "SKILL.md sorts",
        ] {
            assert!(
                !text.contains(old),
                "{file} still points to SKILL.md for the classification: `{old}`"
            );
        }
    }
}
