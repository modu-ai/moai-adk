//! #2903 (#2899 O4) stage 2: the agent workflow has the lane close its own card — close table,
//! `moai close-check`, then close on `close` or keep open with follow-up cards on `keep_open` —
//! and the leader workflow only confirms it. The flags the workflow names are the command's own.

mod common;
use std::process::Command;

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_agent_workflow_closes_its_own_card_through_close_check` pinned the lane's hand-off step, and `moai card land` now closes the card.

/// The command, its verdict words and the flag the workflow names exist.
#[test]
fn the_workflow_names_only_what_close_check_has() {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["close-check", "--help"])
        .output()
        .unwrap();
    assert!(out.status.success());
    let help = String::from_utf8_lossy(&out.stdout);
    assert!(help.contains("--repo"), "{help}");
    for verdict in ["close", "keep_open", "no_table", "uncertain"] {
        assert!(help.contains(verdict), "{verdict} not in help: {help}");
    }
    for heading in ["상태", "판정", "결과"] {
        assert!(
            moai_cli::close_check::VERDICT_HEADINGS.contains(&heading),
            "{heading}"
        );
    }
}
