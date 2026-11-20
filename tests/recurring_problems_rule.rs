//! SPEC-20261120-recurring-problems-source: `AGENTS.local.md` tells the leader that a problem that
//! happens twice is fixed by a card that changes moai-adk itself, and records where the rule came
//! from. The tests read the file only; no host, no HOME and no command runs here.

use std::fs;

fn text() -> String {
    fs::read_to_string(format!("{}/AGENTS.local.md", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("AGENTS.local.md: {e}"))
}

/// The lines of the `## <heading>` section, up to the next `## ` heading.
fn section(text: &str, heading: &str) -> String {
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line == format!("## {heading}") {
            inside = true;
            continue;
        } else if inside && line.starts_with("## ") {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "AGENTS.local.md has no section {heading}");
    out
}

/// The rule's bullet in `## Leader supervision`: from its first words to the end of that line.
fn rule() -> String {
    let supervision = section(&text(), "Leader supervision");
    supervision
        .lines()
        .find(|l| l.starts_with("- Recurring problems change the source"))
        .unwrap_or_else(|| {
            panic!("Leader supervision has no `Recurring problems change the source` item")
        })
        .to_string()
}

/// The provenance row for the rule.
fn provenance_row() -> String {
    let table = section(&text(), "Rule provenance");
    table
        .lines()
        .find(|l| l.contains("Leader supervision: recurring problems change the source"))
        .unwrap_or_else(|| panic!("Rule provenance has no row for the recurring-problems rule"))
        .to_string()
}

/// Card numbers (`C-` and digits, not inside a longer word) in `text`.
fn card_numbers(text: &str) -> Vec<String> {
    let bytes = text.as_bytes();
    let mut hits = Vec::new();
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
    hits
}

#[test]
fn the_supervision_section_states_the_rule() {
    let rule = rule();
    for phrase in [
        "(recursive self-improvement)",
        "when the same class of problem happens twice",
        "a rule broken twice",
        "one CI or integration failure cause twice",
        "a lane block with one cause twice",
        "a manual leader step repeated",
        "the leader opens a card that round which changes moai-adk itself",
        "check, test, command, guard or generated guidance",
        "so the class cannot recur",
        "adds the cause to Rule provenance",
        "from ledger notes (LANE-BLOCKED kinds, audit FAIL classes, CI failure causes)",
        "follows the normal SPEC, audit and integration path",
        "never relaxes a gate, a HOLD or a user rule",
    ] {
        assert!(rule.contains(phrase), "the rule lacks: {phrase}");
    }
}

#[test]
fn the_provenance_table_names_the_instruction_and_its_causes() {
    let row = provenance_row();
    for phrase in [
        "user instruction 2026-11-20",
        "session-bottleneck-review-2026-11-20.md",
        "narrow test selection",
        "file reservations",
        "card-number commit messages",
    ] {
        assert!(row.contains(phrase), "the provenance row lacks: {phrase}");
    }
    // The row sits inside the table, so it has the table's two cells.
    assert_eq!(row.matches(" | ").count(), 1, "{row}");
    for (what, text) in [("the rule", rule()), ("the provenance row", row)] {
        let hits = card_numbers(&text);
        assert!(hits.is_empty(), "{what} names a card number: {hits:?}");
    }
}

/// The Response language item of `## Machine notes`.
fn response_language() -> String {
    section(&text(), "Machine notes")
        .lines()
        .find(|l| l.starts_with("- Response language:"))
        .unwrap_or_else(|| panic!("Machine notes has no Response language item"))
        .to_string()
}

#[test]
fn the_korean_wording_rule_and_its_provenance_row_are_there() {
    let item = response_language();
    for term in [
        "테스트",
        "브랜치",
        "커밋",
        "푸시",
        "머지",
        "리뷰",
        "워크트리",
        "뮤턴트",
        "픽스처",
        "골든 파일",
        "플레이키 테스트",
    ] {
        assert!(item.contains(term), "the item lacks the term {term}");
    }
    assert!(
        item.contains("never literal coinages such as 시험, 가지, 변이"),
        "the item names no literal coinage: {item}"
    );
    let table = section(&text(), "Rule provenance");
    let row = table
        .lines()
        .find(|l| l.contains("Machine notes: Korean text uses developer terms"))
        .unwrap_or_else(|| panic!("Rule provenance has no row for the Korean wording rule"));
    assert!(row.contains("user instruction 2026-11-20"), "{row}");
    assert_eq!(row.matches(" | ").count(), 1, "{row}");
}

#[test]
fn the_file_stays_inside_its_budget() {
    let len = text().len();
    assert!(len < 32 * 1024, "AGENTS.local.md is {len} bytes");
}
