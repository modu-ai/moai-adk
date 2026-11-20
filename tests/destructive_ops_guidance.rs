//! SPEC-20261120-destructive-ops-checklists: the shipped guidance for code that removes or moves
//! data, for units and limits, and for external-tool versions is in the references, and the
//! audit and implementation agents read it. The tests read the shipped files as they are.

use std::fs;
use std::path::PathBuf;

fn read(rel: &str) -> String {
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()))
}

/// The phrases of `text` that are absent from it.
fn missing<'a>(text: &str, phrases: &[&'a str]) -> Vec<&'a str> {
    phrases
        .iter()
        .copied()
        .filter(|p| !text.contains(p))
        .collect()
}

const DESTRUCTIVE_OPS: &str = "skills/moai/references/internal/destructive-ops.md";
const BOUNDARY_CHECKS: &str = "skills/moai/references/internal/boundary-checks.md";

const DESTRUCTIVE_PHRASES: [&str; 11] = [
    "Every component of the path",
    "not only the last name",
    "Machine output in its NUL form",
    "exactly one record matches",
    "Nothing becomes unreachable",
    "Save the way back first",
    "when the save fails, remove nothing",
    "Leave what was just made",
    "Another owner's contents stay",
    "Fail closed",
    "Break each guard on purpose",
];

const BOUNDARY_PHRASES: [&str; 8] = [
    "Units and limits",
    "counts the same unit on both sides",
    "the same on the side that writes and the side that reads",
    "N, N+1 and empty on both sides",
    "External tools are a boundary too",
    "oldest version on each supported platform",
    "checks support at run time",
    "drives the fallback branch",
];

#[test]
fn destructive_ops_names_every_check() {
    let text = read(DESTRUCTIVE_OPS);
    assert!(
        text.lines()
            .nth(2)
            .is_some_and(|l| l.starts_with("When to read:")),
        "the third line says when to read it:\n{text}"
    );
    assert_eq!(missing(&text, &DESTRUCTIVE_PHRASES), Vec::<&str>::new());
    assert!(
        text.contains("`.git`") && text.contains("nested repository"),
        "the nested-repository rule names `.git`"
    );
}

#[test]
fn boundary_checks_state_units_limits_and_tool_versions() {
    let text = read(BOUNDARY_CHECKS);
    assert_eq!(missing(&text, &BOUNDARY_PHRASES), Vec::<&str>::new());
    // The two new items follow the earlier seven, which keep their numbers.
    for n in 1..=9 {
        assert!(
            text.lines().any(|l| l.starts_with(&format!("{n}. **"))),
            "item {n} is numbered in place:\n{text}"
        );
    }
}

#[test]
fn both_agents_read_destructive_ops() {
    for agent in ["agents/audit.md", "agents/impl.md"] {
        let text = read(agent);
        let reads = text
            .lines()
            .find(|l| l.starts_with("reads:"))
            .unwrap_or_else(|| panic!("{agent} has no reads: line"));
        assert!(
            reads
                .trim_start_matches("reads:")
                .split(',')
                .any(|r| r.trim() == "internal/destructive-ops"),
            "{agent}: {reads}"
        );
    }
}

/// The checks above compare against a phrase list: dropping any phrase from the text must show
/// up as that phrase missing, so a deleted sentence cannot pass unnoticed.
#[test]
fn the_guidance_test_fails_when_an_item_is_missing() {
    let ops = read(DESTRUCTIVE_OPS);
    for p in DESTRUCTIVE_PHRASES {
        assert_eq!(
            missing(&ops.replacen(p, "", 1), &DESTRUCTIVE_PHRASES),
            vec![p],
            "removing {p:?} must be reported"
        );
    }
    let boundary = read(BOUNDARY_CHECKS);
    for p in BOUNDARY_PHRASES {
        assert_eq!(
            missing(&boundary.replacen(p, "", 1), &BOUNDARY_PHRASES),
            vec![p],
            "removing {p:?} must be reported"
        );
    }
}
