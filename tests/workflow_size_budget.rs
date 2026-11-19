//! SPEC-20261119-flaky-policy (card C-99, Verify D): a check that a re-run on the same SHA turned
//! green is never simply passed. The merge train in `mode-leader.md` lets the merge through and
//! records the test, both run ids and the SHA on one flaky card. Pinned here: what the rule says,
//! that the "not a stop" sentence points to it, and that it stays inside its byte budget.
//!
//! The tests read the workflow file of this checkout; nothing is run and no HOME is touched.

use std::{fs, path::Path};

const FILE: &str = "mode-leader.md";
/// Where the rule starts and the words it ends with; the tests cut the rule out between them.
const RULE_START: &str = "Flaky rule (card C-99): ";
const RULE_END: &str = "touches that test.";
/// SPEC REQ-3: the rule is at most this many bytes.
const RULE_BUDGET: usize = 700;
/// The bound `profile::read` puts on every package file.
const PACKAGE_FILE_BOUND: usize = 65536;

fn workflow() -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(FILE),
    )
    .unwrap()
}

/// The rule's text, from its start words to its end words.
fn rule(text: &str) -> &str {
    let start = text
        .find(RULE_START)
        .unwrap_or_else(|| panic!("{FILE} has no `{RULE_START}…` rule"));
    let rest = &text[start..];
    let end = rest
        .find(RULE_END)
        .unwrap_or_else(|| panic!("the flaky rule does not end with `{RULE_END}`"));
    &rest[..end + RULE_END.len()]
}

/// The sentence of the merge-train step that calls the re-run result "not a stop".
fn not_a_stop_sentence(text: &str) -> &str {
    let at = text
        .find("turned green is flaky, not a stop")
        .expect("the merge-train step still calls a same-SHA re-run pass flaky, not a stop");
    let from = text[..at].rfind("(Issues #2665").unwrap_or(at);
    let to = text[at..].find(')').map_or(text.len(), |i| at + i + 1);
    &text[from..to]
}

#[test]
fn flaky_rule_names_the_card_and_what_it_records() {
    let text = workflow();
    let rule = rule(&text);
    for needle in [
        // The card: one per test, found by its title, a note on the open one.
        "`flaky: <test>`",
        "moai card add --kind task --title",
        "one card per test",
        "--kind report",
        // What it keeps.
        "the test",
        "the failed run id",
        "the passing run id",
        "the SHA",
        // The merge is not stopped, and the missing-card path (REQ-4).
        "does not stop the merge",
        "merge comment",
        "the card is missing",
        // A re-run's pass is no proof for a change that touches the test.
        "never cited as the verification of a change that",
    ] {
        assert!(
            rule.contains(needle),
            "the flaky rule lacks `{needle}`: {rule}"
        );
    }
}

#[test]
fn the_not_a_stop_sentence_points_to_the_flaky_rule() {
    let text = workflow();
    let sentence = not_a_stop_sentence(&text);
    assert!(
        sentence.contains("flaky rule"),
        "the `not a stop` sentence does not point to the flaky rule: {sentence}"
    );
}

#[test]
fn flaky_rule_stays_within_its_byte_budget() {
    let text = workflow();
    let rule = rule(&text);
    assert!(
        rule.len() <= RULE_BUDGET,
        "the flaky rule is {} bytes, over the {RULE_BUDGET}-byte budget",
        rule.len()
    );
    // N and N+1 on the bound itself: the file is under it, and the bound is not the budget.
    assert!(
        text.len() <= PACKAGE_FILE_BOUND,
        "{FILE} is {} bytes, over the {PACKAGE_FILE_BOUND}-byte package file bound",
        text.len()
    );
    const { assert!(RULE_BUDGET < PACKAGE_FILE_BOUND) };
}

#[test]
fn rule_cutting_reads_the_text_between_its_markers() {
    let sample = format!("before {RULE_START}one {RULE_END} after {RULE_END}");
    assert_eq!(rule(&sample), format!("{RULE_START}one {RULE_END}"));
    let sentence = "x (Issues #2665, #2777; a failure that a re-run on the same SHA turned green is flaky, not a stop) y";
    assert_eq!(
        not_a_stop_sentence(sentence),
        "(Issues #2665, #2777; a failure that a re-run on the same SHA turned green is flaky, not a stop)"
    );
}
