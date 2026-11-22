//! #2831 AC-4 (D3): the agent workflow sends a lane to `lane-next` when a task ends, says it
//! reads GitHub afresh, lets a reversible decision go on with the safe default, and names
//! only flags the command really has.

use std::{fs, path::Path};

fn work() -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/work.md"))
        .unwrap()
}

/// Re-pointed to WORK with the retired MODE-AGENT route (SPEC-20261122-next-flow-skills-agents
/// REQ-3): the next card comes from the ledger, read afresh each time; the name stays.
#[test]
fn the_agent_loop_runs_lane_next_and_goes_on_with_reversible_defaults() {
    let text = work();
    // Card C-53 S5: the next task comes from the ledger (`moai card next`); `lane-next` reads
    // Issue claims, so the loop no longer names it. WORK opens no pull request, so the pull
    // request, draft, decision-applied and tip-repair phrases left with MODE-AGENT.
    for phrase in [
        "Read the state afresh each time (`moai card next`) — never what you remember",
        "a restart, a usage-limit stop or a compacted conversation misses nothing",
    ] {
        assert!(text.contains(phrase), "work.md must say `{phrase}`");
    }
    assert!(!text.contains("lane-next"), "work.md names `lane-next`");
}

/// #2880 W4 (AC-3): the project rules say the lane's PLAN comment is part of the design, next
/// to the Issue body.
#[test]
fn agents_md_names_the_plan_comment_as_part_of_the_design() {
    let agents =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("AGENTS.md")).unwrap();
    // The Issue number (#2880) moved to the provenance table in AGENTS.local.md (#3096 S3); the
    // rule's words stay pinned.
    assert!(
        agents.contains("Each Issue body plus the lane's `PLAN #<issue>` comment is the design;")
    );
    assert!(!agents.contains("Each Issue body is the design;"));
    let local =
        fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("AGENTS.local.md")).unwrap();
    assert!(local.contains(
        "| Authority 2: the lane's `PLAN #<issue>` comment is part of the design | Issue #2880 |"
    ));
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_flags_the_workflow_names_are_the_commands_own` pinned the `lane-next` and `moai wait` flags MODE-AGENT showed, and WORK names neither.
