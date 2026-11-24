//! SPEC-20261120-recurring-problems-source, as `SPEC-20261125-completion-flow` changed it:
//! `CLAUDE.md` tells every session that a problem that happens a second time adds one line to
//! `docs/later.md` that names the change at the source, and `docs/claude-md-rule-provenance.md`
//! records where the rule came from (moved there by `SPEC-20261123-claude-md-rewrite`). The tests
//! read files only; no host, no HOME and no command runs here.

use std::fs;

fn read(path: &str) -> String {
    fs::read_to_string(format!("{}/{path}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{path}: {e}"))
}

fn text() -> String {
    read("CLAUDE.md")
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
    assert!(!out.is_empty(), "no section {heading}");
    out
}

/// The rule's bullet in `## Standing principles`: from its first words to the end of that line.
fn rule() -> String {
    let principles = section(&text(), "Standing principles");
    principles
        .lines()
        .find(|l| l.starts_with("- Recurring problems change the source"))
        .unwrap_or_else(|| {
            panic!("Standing principles has no `Recurring problems change the source` item")
        })
        .to_string()
}

/// A row of the provenance table, found by its first cell.
fn provenance_row_for(first_cell: &str) -> String {
    let table = section(
        &read("docs/claude-md-rule-provenance.md"),
        "Rule provenance",
    );
    table
        .lines()
        .find(|l| l.contains(first_cell))
        .unwrap_or_else(|| panic!("Rule provenance has no row for `{first_cell}`"))
        .to_string()
}

/// The provenance row for the rule.
fn provenance_row() -> String {
    provenance_row_for("Standing principles: recurring problems change the source")
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
fn the_principles_section_states_the_rule() {
    let rule = rule();
    for phrase in [
        "when the same class of problem happens a second time",
        "a rule broken twice",
        "one CI or integration failure cause twice",
        "a lane block with one cause twice",
        "a manual leader step repeated",
        "adds one line to `docs/later.md` that names the cause and the change at the source",
        "(check, test, command, guard or generated guidance)",
        "the leader turns such lines into cards when work is free",
        "never relaxes a gate, a HOLD or a user rule",
    ] {
        assert!(rule.contains(phrase), "the rule lacks: {phrase}");
    }
}

/// The RSI bullet in `## Standing principles` (user instruction of 2026-11-24T22:19:11Z).
#[test]
fn the_principles_section_names_recursive_self_improvement() {
    let principles = section(&text(), "Standing principles");
    let rsi = principles
        .lines()
        .find(|l| l.starts_with("- Recursive self-improvement (RSI)"))
        .unwrap_or_else(|| panic!("Standing principles has no RSI item"))
        .to_string();
    for phrase in [
        "`CLAUDE.local.md`",
        "`skills/**`, `agents/**`",
        "At every round and when every card ends",
        "the same class a second time is fixed in that round",
        "the card that already owns the result (a new card only when none does)",
        "measure before and after with a `moai` command",
        "revert or fix again",
        "Removing comes before adding",
        "a gate, a HOLD or a user rule is never weakened",
        "team rules to this file through a card",
        "product defaults to the templates through a SPEC",
    ] {
        assert!(rsi.contains(phrase), "the RSI item lacks: {phrase}");
    }
    let hits = card_numbers(&rsi);
    assert!(
        hits.is_empty(),
        "the RSI item names a card number: {hits:?}"
    );
    // It is the one place that says what the rule asks; the evolution section only points to it.
    let evolves = section(&text(), "How this file evolves");
    assert!(
        evolves.contains("RSI rule of Standing principles"),
        "{evolves}"
    );
    // The rule is recorded with its source, outside the table of the earlier rule.
    let row = provenance_row_for("Standing principles: Recursive self-improvement (RSI)");
    assert!(row.contains("2026-11-24T22:19:11Z"), "{row}");
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

/// The Korean wording item of `## Language`.
fn response_language() -> String {
    section(&text(), "Language")
        .lines()
        .find(|l| l.starts_with("- Korean text uses the words developers actually say"))
        .unwrap_or_else(|| panic!("Language has no Korean wording item"))
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
    // The sentence itself, not only its words: `uses` must not turn into `avoids`, and the
    // list opens with the plain word 테스트, which `플레이키 테스트` alone would also contain.
    assert!(
        item.contains("Korean text uses the words developers actually say (테스트, 브랜치, 커밋,"),
        "the item lacks the opening of the sentence: {item}"
    );
    assert!(
        item.contains("never literal coinages such as 시험, 가지, 변이"),
        "the item names no literal coinage: {item}"
    );
    let hits = card_numbers(&item);
    assert!(hits.is_empty(), "the item names a card number: {hits:?}");
    let row = provenance_row_for("Language: Korean text uses developer terms");
    assert!(row.contains("user instruction 2026-11-20"), "{row}");
    assert_eq!(row.matches(" | ").count(), 1, "{row}");
    let hits = card_numbers(&row);
    assert!(hits.is_empty(), "the row names a card number: {hits:?}");
}

/// Sentences that an earlier trim for the size budget removed by mistake; they are rules, not
/// evidence, so they stay, now in the file that holds their topic.
#[test]
fn the_rule_sentences_kept_while_making_room_are_still_there() {
    for (path, phrase) in [
        (
            "docs/lane-rules.md",
            "An audit clone builds in its own `target/` too.",
        ),
        (
            ".moai/specs/SPEC-20261123-claude-md-rewrite/claude-local-draft.md",
            "Filling the shared folder again from crates.io needs the network and is the leader's act.",
        ),
        (
            "CLAUDE.md",
            "text typed into a session's chat input (including what the leader sends with `orca terminal send`)",
        ),
    ] {
        assert!(read(path).contains(phrase), "{path} lost: {phrase}");
    }
}

#[test]
fn the_file_stays_inside_its_budget() {
    let lines = text().lines().count();
    assert!(lines <= 200, "CLAUDE.md has {lines} lines");
}
