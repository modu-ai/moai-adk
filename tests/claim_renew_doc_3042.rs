//! #3042: the agent workflow says what `claim-acquire` does with an expired claim, in the words
//! the command's own note uses (the behaviour is pinned in `tests/claim_cli.rs`).

use std::{fs, path::Path};

#[test]
fn the_agent_workflow_says_what_the_command_does_with_an_expired_claim() {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/mode-agent.md"),
    )
    .unwrap();
    let step = text
        .lines()
        .find(|l| l.starts_with("5. **claim**"))
        .expect("claim step");
    // The same words the command's note uses (pinned in tests/claim_cli.rs).
    for phrase in [
        "`moai claim-acquire` does the same for your own expired record in the local journal (Issue #3042)",
        "the same session name gets `claimed` with the next epoch and a new claim line",
        "another session's expired record still answers `expired`",
        "`moai claim-timeout-check`",
        "decision D2",
    ] {
        assert!(step.contains(phrase), "{phrase:?} missing from the claim step");
    }
}
