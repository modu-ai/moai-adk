//! SPEC-20261120-plan-output-write-set: the output of the moai-plan agent names the write set,
//! the shared one-line spots, the tests that read the changed text and the current holders; the
//! check-completeness reference asks a port for a verdict parity; the chain table carries the
//! flaky-test and the repair-after-FAIL paths. Reads the shipped sources; writes nothing.

use std::path::Path;

fn read(rel: &str) -> String {
    std::fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap()
}

/// The text of one `## <name>` section of a markdown file, up to the next `## `.
fn section<'a>(text: &'a str, name: &str) -> &'a str {
    let head = format!("\n## {name}\n");
    let start = text
        .find(&head)
        .unwrap_or_else(|| panic!("no section `{name}`"))
        + head.len();
    let rest = &text[start..];
    &rest[..rest.find("\n## ").unwrap_or(rest.len())]
}

/// Collapse every run of whitespace to one space, so a phrase may wrap across lines.
fn flat(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

#[test]
fn output_lists_the_write_set_shared_spots_reading_tests_and_holders_in_order() {
    let plan = read("agents/plan.md");
    let output = section(&plan, "Output");
    let mut at = 0;
    for name in [
        "`Steps`",
        "`Write set`",
        "`Shared spots`",
        "`Tests that read changed text`",
        "`Current holders`",
        "`Files written`",
    ] {
        let found = output[at..]
            .find(name)
            .unwrap_or_else(|| panic!("Output lacks {name} after offset {at}:\n{output}"));
        at += found + name.len();
    }
    let out = flat(output);
    for phrase in [
        "every file the plan changes, by exact path",
        "a module list, a help table, an operation map, a golden file or a manifest",
        "found by searching the tests for the path",
    ] {
        assert!(out.contains(phrase), "Output must say {phrase:?}:\n{out}");
    }
}

#[test]
fn holders_are_read_from_the_card_ledger_and_say_no_ledger_without_one() {
    let plan = read("agents/plan.md");
    let verification = flat(section(&plan, "Verification"));
    for phrase in [
        "card ledger",
        "current holder of each file",
        "`moai card list`",
        "`moai card show`",
        "`free`",
        "`no ledger`",
    ] {
        assert!(
            verification.contains(phrase),
            "Verification must say {phrase:?}:\n{verification}"
        );
    }
}

#[test]
fn parity_a_port_keeps_the_old_tools_verdicts_over_the_same_real_inputs() {
    let text = read("skills/moai/references/internal/check-completeness.md");
    let seventh = flat(
        text.lines()
            .find(|l| l.starts_with("7. "))
            .expect("an item 7"),
    );
    for phrase in [
        "A port or rewrite keeps the old tool's verdicts",
        "the same real inputs",
        "every input whose verdict differs",
        "unless the SPEC names it as an intended change",
    ] {
        assert!(
            seventh.contains(phrase),
            "item 7 must say {phrase:?}:\n{seventh}"
        );
    }
}

#[test]
fn chains_carry_the_flaky_test_and_repair_paths_and_the_first_auditor_rule() {
    let text = read("skills/moai/references/chains/index.md");
    let rows: Vec<&str> = text
        .lines()
        .filter(|l| l.starts_with("| ") && !l.starts_with("| Request") && !l.starts_with("| ---"))
        .collect();
    assert_eq!(rows.len(), 10, "{rows:#?}");
    let flaky = rows
        .iter()
        .find(|r| r.contains("flaky"))
        .expect("a flaky-test row");
    assert!(flaky.contains("VERIFY"), "{flaky}");
    let repair = rows
        .iter()
        .find(|r| r.contains("FAIL"))
        .expect("a repair-after-FAIL row");
    for word in ["RUN", "VERIFY", "SYNC", "AUDIT", "moai-impl", "moai-audit"] {
        assert!(repair.contains(word), "{repair}");
    }
    let flat_text = flat(&text);
    assert!(
        flat_text.contains(
            "The re-audit after a FAIL goes to the first auditor, who judges the repair again and every changed criterion."
        ),
        "the first-auditor rule is missing"
    );
}
