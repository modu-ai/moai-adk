//! SPEC-20261121-report-rules-in-core-rules: the completion-report rule and the report-number
//! rule are rules 23 and 24 of `references/internal/core-rules.md`, the skill entry keeps one line
//! that says so, and the decision order, the HOLD categories and the decision record stay in the
//! skill entry. The test reads the named files by path; it never walks the repository.

use std::{fs, path::Path};

fn skill(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("skills/moai")
        .join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

/// The completion-report rule, word for word as the skill entry carried it.
const COMPLETION: &str = "Report completion only after verifying the actual changes and confirming an independent AUDIT result. A command acknowledgement or the existence of a file is not success. A local report file is saved under `~/.moai/reports/<project>/` in the user-level MoAI home, never under the repository.";

/// The report-number rule, word for word as the skill entry carried it.
const NUMBERS: &str = "Take every number and state in a report from a command run for that report, and name the command next to it. An empty output is not zero: say whether the command prints nothing for \"none\" or also on a failure, and count a paged list only after reading every page. After a write, read the written thing back and report what the read shows, not what was sent. Write a time only from a clock read in the same step (`date -u`).";

#[test]
fn the_core_rules_carry_the_moved_report_rules() {
    let rules = skill("references/internal/core-rules.md");
    assert!(
        rules.contains(&format!("\n23. {COMPLETION}\n")),
        "rule 23 is the completion-report rule"
    );
    assert!(
        rules.contains(&format!("\n24. {NUMBERS}\n")),
        "rule 24 is the report-number rule"
    );
    // A section of its own after rule 22, so the numbers keep their order.
    let at = |needle: &str| {
        rules
            .find(needle)
            .unwrap_or_else(|| panic!("core-rules.md lacks {needle:?}"))
    };
    assert!(at("\n22. ") < at("\n## Reports\n"));
    assert!(at("\n## Reports\n") < at("\n23. "));
}

#[test]
fn the_skill_entry_points_to_the_core_rules_and_keeps_the_decision_order() {
    let entry = skill("SKILL.md");
    // The two moved rules are gone from the entry (their first sentences are the markers).
    for moved in [
        "Report completion only after verifying the actual changes",
        "Take every number and state in a report from a command run",
        "Write a time only from a clock read in the same step",
    ] {
        assert!(!entry.contains(moved), "SKILL.md still carries {moved:?}");
    }
    // One line says where they went, by the route table's entry and not by a package path (the
    // skill entry names only the route table: `skill_package_frontmatter_is_host_neutral`).
    let pointer = entry
        .lines()
        .find(|l| {
            l.contains("rules 23 and 24 of the core rules (the route table's `core_rules` entry)")
        })
        .expect("SKILL.md points to rules 23 and 24 of the core rules");
    assert!(
        !pointer.contains("references/"),
        "the pointer names no package path: {pointer}"
    );
    assert!(
        pointer.len() < 260,
        "the pointer is one short line: {pointer}"
    );
    // The decision order, the HOLD categories and the record fields stay where they are defined.
    for kept in [
        "the five HOLD categories",
        "Independent audit and the CI gate are never delegated.",
        "Every decision record names what was decided, its options, the default, whether it is reversible",
        "a decision without a record does not count.",
    ] {
        assert!(entry.contains(kept), "SKILL.md lost {kept:?}");
    }
}
