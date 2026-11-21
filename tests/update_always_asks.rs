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
fn the_rule_tells_the_model_to_ask_the_user_and_to_record_the_answer() {
    let text = rule();
    for phrase in [
        "ask the user one selection through the host's question tool",
        "and record the answer for the report",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn an_undo_the_person_asked_for_does_not_wait_for_the_question() {
    let text = rule();
    for phrase in [
        "An undo the person asked for by name (`moai models-revert`, step 6)",
        "is not an apply of this preview and does not wait for this question",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn the_question_is_the_one_ask_of_step_two_and_has_a_plain_text_form_for_codex() {
    let text = rule();
    for phrase in [
        "This question is the one ask of step 2 for the `ready` steps as a whole, never a second one",
        "where the host cannot ask it as a selection (Codex in Default mode)",
        "numbered plain-text choices with the recommended one first, as `internal/questions` says",
    ] {
        assert!(text.contains(phrase), "the rule lacks {phrase:?}: {text}");
    }
}

#[test]
fn the_options_of_each_case_are_named_with_the_recommended_one_first() {
    let text = rule();
    let at = |needle: &str| {
        text.find(needle)
            .unwrap_or_else(|| panic!("the rule lacks {needle:?}: {text}"))
    };
    assert!(
        at("\"Apply what is ready (recommended)\"") < at("\"Leave everything as it is\""),
        "ready case: the recommended option must come first: {text}"
    );
    assert!(
        text.contains("with \"Nothing to do, finish (recommended)\" first"),
        "the nothing-to-apply case lacks its recommended option first: {text}"
    );
    assert!(
        at("\"Apply the rest (recommended)\" first") < at("\"Stop here\" second"),
        "blocked case: the recommended option must come first: {text}"
    );
    assert!(
        !text.contains("\"Leave everything as it is (recommended)\"")
            && !text.contains("\"Stop here (recommended)\""),
        "a stopping option must not be the recommended one: {text}"
    );
}

#[test]
fn no_line_before_the_rule_names_a_command_that_applies_or_undoes() {
    let all = lines();
    let found = rule_lines(&all);
    assert_eq!(found.len(), 1, "update.md has no single rule paragraph");
    for (i, line) in all.iter().enumerate().take(found[0].0) {
        for word in [
            "--confirm",
            "update-apply",
            "-revert",
            "--permission-reference",
        ] {
            assert!(
                !line.contains(word),
                "line {} names {word:?} before the rule at line {}: {line}",
                i + 1,
                found[0].0 + 1
            );
        }
    }
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
