//! #2401 (leader decision 6864527006): lane rules from the observations of 2026-11-13, pinned
//! in the shipped workflows. The verdict example in AUDIT is fed to the parser that reads card
//! verdicts, so the documented shape and the code cannot drift apart.
use moai_cli::gate::parse_verdict;
use std::{fs, path::Path};

fn workflow(name: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/workflows")
            .join(name),
    )
    .unwrap()
}

fn says(text: &str, file: &str, phrases: &[&str]) {
    for p in phrases {
        assert!(text.contains(p), "{file} must say `{p}`");
    }
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `mode_agent_gates_the_opening_exempts_tip_repairs_and_verifies_before_a_merge` pinned the pull-request opening gate, and WORK opens no pull request.

#[test]
fn run_puts_a_new_test_in_a_file_of_its_own() {
    says(
        &workflow("run.md"),
        "run.md",
        &[
            "Add a test to a file of its own under `tests/`",
            "`tests/product_cli.rs`",
        ],
    );
}

/// The fenced example after AUDIT's **Verdict note.** step, without its list indent.
fn verdict_example(audit: &str) -> String {
    let step = audit
        .find("**Verdict note.**")
        .expect("AUDIT has a Verdict note step");
    let open = step + audit[step..].find("```text\n").expect("an example") + "```text\n".len();
    let close = open + audit[open..].find("```").expect("a closed example");
    audit[open..close]
        .lines()
        .map(str::trim)
        .filter(|l| !l.is_empty())
        .collect::<Vec<_>>()
        .join("\n")
}

/// SPEC-20261122-next-flow-skills-agents (slice B audit N1): the verdict is a card note, not a
/// pull-request comment; its first line keeps the shape the card-verdict reader parses.
#[test]
fn the_documented_verdict_example_passes_the_card_verdict_parser() {
    let audit = workflow("audit.md");
    says(
        &audit,
        "audit.md",
        &[
            "`AUDIT <card> (<one-line scope>) head <40-hex> base <40-hex> — verdict: PASS|FAIL`",
            "lowercase hex",
            "nothing after the verdict word",
            "The next line names the auditing session and its reservation note",
        ],
    );
    let example = verdict_example(&audit);
    let mut lines = example.lines();
    let first = lines.next().unwrap();
    assert!(first.starts_with("AUDIT C-"), "{first}");
    assert_eq!(
        parse_verdict(first),
        Some((
            "997af563452bfb36f2ce722a66e9068dae191568".to_string(),
            "9c8a8c51c9c059be389deeaebb19a3e45709b273".to_string(),
            "PASS".to_string()
        )),
        "{first}"
    );
    let second = lines.next().unwrap_or_default();
    assert!(second.starts_with("auditor: "), "{example}");
}

/// The rules the AUDIT text states — lowercase hex, nothing after the verdict word, the em dash —
/// are the parser's: each broken form is refused.
#[test]
fn each_rule_the_text_states_is_one_the_parser_enforces() {
    let example = verdict_example(&workflow("audit.md"));
    let first = example.lines().next().unwrap();
    let upper = first.replace("997af563", "997AF563");
    assert_eq!(parse_verdict(&upper), None, "uppercase hex");
    assert_eq!(
        parse_verdict(&format!("{first} (CI 43493998238)")),
        None,
        "text after the verdict word"
    );
    assert_eq!(
        parse_verdict(&first.replace(" — ", " - ")),
        None,
        "a hyphen instead of the em dash"
    );
}
