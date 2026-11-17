//! #3042: the agent workflow says what happens to a lease that ended. Card C-53 S5 moved the
//! round to the card ledger: `card renew` needs a live lease (`moai card renew --help`), so an
//! ended lease goes to the leader; the `claim-acquire` behaviour stays pinned in
//! `tests/claim_cli.rs`.

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
    // C-53 S5: a card lane renews with `card renew`, which needs a live lease; an ended lease
    // is the leader's to settle (was: `moai claim-acquire` renewing this session's own expired
    // Issue claim record with a next-epoch claim line, and decision D2's timeout takeover).
    for phrase in [
        "Keep the lease: run `card renew <card> --generation <g>` before the lease ends",
        "`renew` needs a lease that has not ended, so an ended lease is the leader's to settle",
        "`LANE-BLOCKED … kind=claim owner=leader`",
    ] {
        assert!(
            step.contains(phrase),
            "{phrase:?} missing from the claim step"
        );
    }
    for gone in ["claim-acquire", "claim-timeout-check"] {
        assert!(!step.contains(gone), "the claim step still names `{gone}`");
    }
}
