//! SPEC-20261119-flaky-policy (card C-99, Verify D): a check that a re-run on the same SHA turned
//! green is never simply passed. The flaky rule lets the merge through and records the test, both
//! run ids and the SHA on one flaky card. It moved from the merge train of the retired
//! `mode-leader.md` to `run.md` (SPEC-20261122-next-flow-skills-agents). Pinned here: what the
//! rule says and that it stays inside its byte budget.
//!
//! The tests read the workflow file of this checkout; nothing is run and no HOME is touched.

use std::{fs, path::Path};

const FILE: &str = "run.md";
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

/// SPEC-20261119-flaky-rule-ac (card C-108 criterion 3): the `card add` example of the rule carries
/// a criterion, so the flaky card it makes can be closed (`card done` refuses a card without a met
/// criterion), and the criterion says what closes the card.
#[test]
fn flaky_rule_card_add_example_carries_a_criterion() {
    let text = workflow();
    let rule = rule(&text);
    let head = "moai card add --kind task --title \"flaky: <test>\"";
    let at = rule
        .find(head)
        .unwrap_or_else(|| panic!("the flaky rule lost its `card add` example: {rule}"));
    let example = &rule[at..];
    let example = &example[..example
        .find("`)")
        .expect("the example ends with a backtick")];
    let criterion = example
        .strip_prefix(head)
        .and_then(|rest| rest.strip_prefix(" --ac \""))
        .and_then(|rest| rest.strip_suffix('"'))
        .unwrap_or_else(|| panic!("the example has no ` --ac \"…\"` criterion: {example}"));
    assert!(
        !criterion.contains('"') && criterion.len() > 10,
        "the criterion is one quoted phrase: {criterion}"
    );
    for word in ["fixed", "retired", "reason"] {
        assert!(
            criterion.contains(word),
            "the criterion does not say what closes the card (`{word}`): {criterion}"
        );
    }
}

// `the_not_a_stop_sentence_points_to_the_flaky_rule` is retired: the merge train is retired
// (SPEC-20261122-next-flow-skills-agents REQ-2).

/// The one place the budget bound is applied; the tests below drive it at N and N+1.
fn budget_ok(len: usize) -> bool {
    len <= RULE_BUDGET
}

#[test]
fn flaky_rule_stays_within_its_byte_budget() {
    let text = workflow();
    let rule = rule(&text);
    assert!(
        budget_ok(rule.len()),
        "the flaky rule is {} bytes, over the {RULE_BUDGET}-byte budget",
        rule.len()
    );
    // N and N+1 on the bound itself, through the one helper the check uses: 700 passes,
    // 701 fails (audit N10: the edge is committed, not only probed).
    assert!(budget_ok(RULE_BUDGET), "the bound rejects its own N");
    assert!(!budget_ok(RULE_BUDGET + 1), "the bound accepts N+1");
    // The file stays far under the package-file bound (its own check, not the budget helper's).
    assert!(
        text.len() <= PACKAGE_FILE_BOUND,
        "{FILE} is {} bytes, over the {PACKAGE_FILE_BOUND}-byte package file bound",
        text.len()
    );
}

// `the_flaky_rule_sits_inside_the_merge_train_step` is retired: the merge train is retired
// (SPEC-20261122-next-flow-skills-agents REQ-2).

#[test]
fn rule_cutting_reads_the_text_between_its_markers() {
    let sample = format!("before {RULE_START}one {RULE_END} after {RULE_END}");
    assert_eq!(rule(&sample), format!("{RULE_START}one {RULE_END}"));
}
