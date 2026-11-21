//! SPEC-20261121-update-always-asks: the UPDATE workflow asks one selection right after every
//! preview and applies nothing before the answer. The tests read the workflow as text; no
//! command, host or HOME runs here.

use std::fs;

const START: &str = "**Always ask after the preview.**";

fn lines() -> Vec<String> {
    fs::read_to_string(format!(
        "{}/skills/moai/workflows/update.md",
        env!("CARGO_MANIFEST_DIR")
    ))
    .expect("skills/moai/workflows/update.md")
    .lines()
    .map(str::to_string)
    .collect()
}

/// The index and text of the lines that start the rule paragraph.
fn rule_lines(all: &[String]) -> Vec<(usize, &String)> {
    all.iter()
        .enumerate()
        .filter(|(_, l)| l.starts_with(START))
        .collect()
}

/// The one rule paragraph, or a failure that says what is wrong.
fn rule() -> String {
    let all = lines();
    let found = rule_lines(&all);
    assert_eq!(
        found.len(),
        1,
        "update.md must hold exactly one paragraph starting with {START}, found {}",
        found.len()
    );
    found[0].1.clone()
}

#[test]
fn the_rule_asks_one_selection_through_the_host_after_the_preview_in_both_hosts() {
    let text = rule();
    for phrase in [
        "Right after the preview is shown",
        "`/moai update`",
        "`$moai update`",
        "alike",
        "one selection through the host's question tool",
        "the recommended option first",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn the_rule_covers_a_preview_with_nothing_to_apply_and_one_with_a_blocked_step() {
    let text = rule();
    for phrase in [
        "`nothing to apply — no change` and still ask",
        "`blocked: <its reason>`",
        "ask whether to apply the rest",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn no_apply_runs_in_a_turn_where_the_question_has_no_answer() {
    let text = rule();
    for phrase in [
        "Never run an applying command",
        "`moai update --confirm`",
        "`--confirm-cleanup`",
        "`--confirm-home`",
        "the apply of any step",
        "in a turn where this question has no answer",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn an_unanswered_question_is_open_and_an_earlier_continue_does_not_answer() {
    let text = rule();
    for phrase in [
        "a cancelled, empty or unanswered question is reported as open, never as a yes",
        "an earlier \"continue\" does not answer a new preview",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn order_the_rule_stands_before_the_first_line_that_names_confirm() {
    let all = lines();
    let found = rule_lines(&all);
    assert_eq!(found.len(), 1, "update.md has no single rule paragraph");
    let at = found[0].0;
    let first_confirm = all
        .iter()
        .position(|l| l.contains("--confirm"))
        .expect("update.md names --confirm somewhere");
    assert!(
        at <= first_confirm,
        "line {} names --confirm before the rule at line {}",
        first_confirm + 1,
        at + 1
    );
}

#[test]
fn order_the_steps_that_apply_come_after_the_rule() {
    let all = lines();
    let found = rule_lines(&all);
    assert_eq!(found.len(), 1, "update.md has no single rule paragraph");
    let at = found[0].0;
    for marker in [
        "1. **Preview everything once.**",
        "2. **Ask.**",
        "4. **After the backup phase",
    ] {
        let step = all
            .iter()
            .position(|l| l.starts_with(marker))
            .unwrap_or_else(|| panic!("update.md has no step starting {marker:?}"));
        assert!(
            at < step,
            "the step {marker:?} (line {}) comes before the rule (line {})",
            step + 1,
            at + 1
        );
    }
}
