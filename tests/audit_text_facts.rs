//! SPEC-20261125-audit-text-facts: the two audit texts (`skills/moai/workflows/audit.md` and
//! `skills/moai/references/audit.md`) read as one ordered procedure, state the tier lines, the four
//! FAIL kinds and the verdict line once, and say only what the commands they name do.
//!
//! The texts are read from the checked-in tree; the only process started is the binary's own
//! `--help`, with a temporary HOME.

mod common;

use std::path::Path;
use std::process::Command;
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

fn read(rel: &str) -> String {
    std::fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn workflow() -> String {
    read("workflows/audit.md")
}

fn contract() -> String {
    read("references/audit.md")
}

/// The lines of the `## <heading>` section, up to the next `## ` heading.
fn section<'a>(text: &'a str, heading: &str) -> Vec<&'a str> {
    let mut out = Vec::new();
    let mut inside = false;
    for line in text.lines() {
        if let Some(h) = line.strip_prefix("## ") {
            inside = h.trim() == heading;
            continue;
        }
        if inside {
            out.push(line);
        }
    }
    out
}

/// A line that starts a numbered list item (`1. `, `10. `) at the margin.
fn is_numbered(line: &str) -> bool {
    let digits = line.chars().take_while(char::is_ascii_digit).count();
    digits > 0 && line[digits..].starts_with(". ")
}

/// AC-1 (REQ-1): one numbered list in work order, and no second procedure list.
#[test]
fn ac1_the_procedure_is_one_numbered_list_in_work_order() {
    let text = workflow();
    assert_eq!(
        text.lines().filter(|l| *l == "## Procedure").count(),
        1,
        "one Procedure section"
    );
    assert!(
        !text.lines().any(|l| l.starts_with("## Review steps")),
        "the Review steps section is folded into the Procedure"
    );
    let procedure = section(&text, "Procedure");
    let items: Vec<&str> = procedure
        .iter()
        .copied()
        .filter(|l| is_numbered(l))
        .collect();
    let in_file = text.lines().filter(|l| is_numbered(l)).count();
    assert_eq!(
        items.len(),
        in_file,
        "every numbered list item of the file stands in the Procedure"
    );
    let keys = [
        "AUDIT-CLAIM",
        "ready note",
        "Audit tier",
        "moai review-package",
        "moai review-check",
        "moai card verdict",
        "moai worktree-gc-plan",
    ];
    let mut last = None;
    for key in keys {
        let at = items
            .iter()
            .position(|i| i.contains(key))
            .unwrap_or_else(|| panic!("no Procedure item names {key:?}"));
        assert!(
            last.is_none_or(|l| at > l),
            "{key:?} comes at item {at}, not after item {last:?}"
        );
        last = Some(at);
    }
    let first = items[0];
    assert!(
        first.contains("AUDIT-CLAIM"),
        "the list starts with the reservation"
    );
}

/// AC-2 (REQ-2): the tier lines, the FAIL four-kinds sentence and the verdict-line form stand in
/// the contract once, and the workflow points to their sections.
#[test]
fn ac2_tier_lines_fail_kinds_and_verdict_form_have_one_home() {
    let wf = workflow();
    let ct = contract();
    for start in [
        "- docs-only (tier 0)",
        "- test-only (tier 0",
        "- code (tier 1)",
        "- high-risk (tier 2)",
    ] {
        let in_contract = ct.lines().filter(|l| l.starts_with(start)).count();
        let in_workflow = wf.lines().filter(|l| l.starts_with(start)).count();
        assert_eq!((in_contract, in_workflow), (1, 0), "{start}");
    }
    assert_eq!(
        ct.matches("A verdict is FAIL for four things only").count(),
        1
    );
    for phrase in [
        "four things only",
        "behaviour that differs from the card's acceptance criteria",
        "a weakened gate",
    ] {
        assert!(!wf.contains(phrase), "workflow restates {phrase:?}");
    }
    assert_eq!(
        ct.matches(
            "`AUDIT <card> (<one-line scope>) head <40-hex> base <40-hex> — verdict: PASS|FAIL`"
        )
        .count(),
        1,
        "the verdict line form"
    );
    assert!(
        !wf.contains("AUDIT <card> (<one-line scope>)"),
        "workflow restates the form"
    );
    assert!(
        !wf.contains("AUDIT C-12 "),
        "workflow carries the example line"
    );
    for pointer in [
        "section \"Audit tiers\"",
        "section \"Who records the verdict\"",
        "section \"What blocks\"",
    ] {
        assert!(wf.contains(pointer), "workflow lacks the pointer {pointer}");
    }
}

/// The `moai <command>` names the two texts put in inline code, one or two words each.
fn named_commands(text: &str) -> Vec<String> {
    // A lowercase word with dashes: a command name, not an argument such as `PROJECT` or `--head`.
    fn is_word(w: &str) -> bool {
        w.chars().next().is_some_and(|c| c.is_ascii_lowercase())
            && w.chars().all(|c| c.is_ascii_lowercase() || c == '-')
    }
    // The command groups that have subcommands: only these take a second word.
    const GROUPS: [&str; 4] = ["card", "bundle", "check", "spec"];
    let mut out = Vec::new();
    for part in text.split("`moai ").skip(1) {
        let mut tokens = part.split_whitespace();
        let Some(first) = tokens.next() else { continue };
        let closed = first.ends_with('`');
        let first = first.trim_end_matches('`');
        if !is_word(first) {
            continue;
        }
        let mut name = first.to_string();
        if !closed && GROUPS.contains(&first) {
            if let Some(second) = tokens.next() {
                let second = second.trim_end_matches('`');
                if is_word(second) {
                    name = format!("{first} {second}");
                }
            }
        }
        out.push(name);
    }
    out.sort();
    out.dedup();
    out
}

/// AC-3 (REQ-3, REQ-4): the false sentences are gone, the `no_pass` refusal is stated, and every
/// command either text names answers `--help`.
#[test]
fn ac3_the_texts_state_what_the_commands_do() {
    let wf = workflow();
    let ct = contract();
    for text in [&wf, &ct] {
        for banned in [
            "not implemented",
            "reads no verdict",
            "storing verdicts",
            "the lander reads the change and lands it",
            "audit-reduce",
        ] {
            assert!(!text.contains(banned), "a text still says {banned:?}");
        }
    }
    assert!(ct.contains("`no_pass`") && ct.contains("in every mode, tier 0 included"));
    assert!(ct.contains("`--receipt-step NAME=EXIT`"));
    assert!(wf.contains("refuses a head that has no independent PASS"));

    let home = TempDir::new().unwrap();
    let mut names = named_commands(&wf);
    names.extend(named_commands(&ct));
    names.sort();
    names.dedup();
    assert!(
        names.len() >= 8,
        "the scan found too few commands: {names:?}"
    );
    let mut failing = Vec::new();
    for name in &names {
        let args: Vec<&str> = name.split(' ').collect();
        let out = Command::new(EXE)
            .args(&args)
            .arg("--help")
            .env("HOME", home.path())
            .output()
            .unwrap();
        if out.status.code() != Some(0) || out.stdout.is_empty() {
            failing.push(name.clone());
        }
    }
    assert!(
        failing.is_empty(),
        "commands the texts name that have no --help: {failing:?}"
    );
}

/// AC-4 (REQ-5): the base rule for a lane head and for a merge head.
#[test]
fn ac4_the_contract_states_the_base_of_a_verdict() {
    let ct = contract();
    assert!(
        ct.contains("the base the lane's ready note names when the head is the lane's own head")
    );
    assert!(ct.contains("the tip the head will land on"));
    assert!(ct.contains("accepts only a base that is an ancestor of the head"));
}

/// The floor-list paragraph of the contract: it mirrors the prefixes the tier code reads, so it
/// names repository tools by path (SPEC D-1: the floor list stays as it is).
fn without_floor_list(text: &str) -> String {
    text.lines()
        .filter(|l| !l.starts_with("The tier a ready note states never goes below the floor"))
        .collect::<Vec<_>>()
        .join("\n")
}

/// AC-5 (REQ-6): no Issue number, no internal tool path, no retired host.
#[test]
fn ac5_the_texts_carry_no_issue_number_internal_tool_or_retired_host() {
    for (name, text) in [
        ("workflows/audit.md", workflow()),
        ("references/audit.md", without_floor_list(&contract())),
    ] {
        for banned in ["Issue #", "tools/", "Kiro", "Cursor", "Codex", "OMP"] {
            assert!(!text.contains(banned), "{name} contains {banned:?}");
        }
        let bytes = text.as_bytes();
        let hash_digits = (0..bytes.len().saturating_sub(4))
            .any(|i| bytes[i] == b'#' && bytes[i + 1..i + 5].iter().all(u8::is_ascii_digit));
        assert!(!hash_digits, "{name} names an issue or pull request number");
    }
}

/// AC-6 (REQ-7, REQ-8): the release pull request part keeps its four items, and the worktree
/// removal is the leader's plan and apply with their real options.
#[test]
fn ac6_the_release_part_and_the_worktree_removal_are_kept_short_and_true() {
    let wf = workflow();
    let release = section(&wf, "Release pull request").join("\n");
    for needle in [
        "moai pr-self-audit-check --reaudit --file BODY.md",
        "ci: gate_unavailable(<cause> since <since>)",
        "its run id, head and conclusion",
        "author_association",
    ] {
        assert!(
            release.contains(needle),
            "the release part lacks {needle:?}"
        );
    }
    assert!(wf.contains("moai worktree-gc-plan PROJECT --integration-ref <branch>"));
    assert!(wf.contains("`--plan-digest`"));
    for text in [&wf, &contract()] {
        assert!(
            !text.contains("worktree-gc-apply --worktree"),
            "an option the command lacks"
        );
    }
}

/// The rules REQ-9 keeps; each phrase has to stay in the contract.
const KEPT_RULES: [&str; 6] = [
    "behaviour that differs from the card's acceptance criteria",
    "data loss",
    "a security problem",
    "a weakened gate",
    "The author of a change never audits it",
    "The auditing session records the verdict itself",
];

/// The kept rules that the text lacks.
fn rule_problems(contract: &str) -> Vec<&'static str> {
    KEPT_RULES
        .iter()
        .copied()
        .filter(|p| !contract.contains(p))
        .collect()
}

/// AC-7 (REQ-9): the FAIL four kinds, the author rule and the verdict-recording rule stay, and a
/// text without one of them is found (the guard mutant of this SPEC).
#[test]
fn ac7_the_kept_gate_rules_stay_and_a_missing_one_is_found() {
    let ct = contract();
    assert_eq!(rule_problems(&ct), Vec::<&str>::new());
    for phrase in KEPT_RULES {
        let mutant = ct.replacen(phrase, "", 1);
        assert_eq!(
            rule_problems(&mutant),
            vec![phrase],
            "dropping {phrase:?} passes"
        );
    }
}
