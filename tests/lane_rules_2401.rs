//! #2401 (leader decision 6864527006): lane rules from the observations of 2026-11-13, pinned
//! in the shipped workflows. The verdict example in AUDIT is fed to the leader gate's own
//! parsers, so the documented shape and the gate cannot drift apart.
use moai_cli::gate::{auditor, parse_verdict};
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

#[test]
fn mode_agent_gates_the_opening_exempts_tip_repairs_and_verifies_before_a_merge() {
    let agent = workflow("mode-agent.md");
    says(
        &agent,
        "mode-agent.md",
        &[
            // P-C: the count and the opening are one command.
            "Gate the opening on that count in the same command",
            "-lt <N> ] && gh pr create",
            // P-B: a tip repair is outside the open-PR limit.
            "A pull request whose body carries a `moai-tip-repair:` line is outside this count",
            "never convert it to a draft",
            // P-E: a tip merge is a commit.
            "A tip merge is a commit too: run `moai claim-verify` before `git merge`",
        ],
    );
}

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

/// The fenced example after AUDIT's **Verdict comment.** step, without its list indent.
fn verdict_example(audit: &str) -> String {
    let step = audit
        .find("**Verdict comment.**")
        .expect("AUDIT has a Verdict comment step");
    let open = step + audit[step..].find("```text\n").expect("an example") + "```text\n".len();
    let close = open + audit[open..].find("```").expect("a closed example");
    audit[open..close]
        .lines()
        .map(str::trim)
        .filter(|l| !l.is_empty())
        .collect::<Vec<_>>()
        .join("\n")
}

#[test]
fn the_documented_verdict_example_passes_the_leader_gate_parsers() {
    let audit = workflow("audit.md");
    says(
        &audit,
        "audit.md",
        &[
            "`AUDIT #<pr> (<one-line scope>) head <40-hex> base <40-hex> — verdict: PASS|FAIL`",
            "`^AUDIT #[0-9]+ .* head [0-9a-f]{40} base [0-9a-f]{40} — verdict: (PASS|FAIL)$`",
            "lowercase hex",
            "nothing after the verdict word",
            "The next line names the auditing lane and its reservation comment",
        ],
    );
    let example = verdict_example(&audit);
    let first = example.lines().next().unwrap();
    assert_eq!(
        parse_verdict(first),
        Some((
            "997af563452bfb36f2ce722a66e9068dae191568".to_string(),
            "9c8a8c51c9c059be389deeaebb19a3e45709b273".to_string(),
            "PASS".to_string()
        )),
        "{first}"
    );
    assert_eq!(auditor(&example).as_deref(), Some("agent-27"), "{example}");
}

/// The rules the AUDIT text states — lowercase hex, nothing after the verdict word, the lane on
/// the line after the first — are the gate's: each broken form is refused.
#[test]
fn each_rule_the_text_states_is_one_the_gate_enforces() {
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
    assert_eq!(
        auditor(&format!("{first}\n감사: 레인 이름 없음")),
        None,
        "a verdict with no lane name after its first line has no auditor"
    );
    let other_first = example.replacen("감사: agent-27", "감사: agent-9 대신 agent-27", 1);
    assert_eq!(
        auditor(&other_first).as_deref(),
        Some("agent-9"),
        "the first lane named wins"
    );
}
