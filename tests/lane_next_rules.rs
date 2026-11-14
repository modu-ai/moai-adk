//! #2831 AC-4 (D3): the agent workflow sends a lane to `lane-next` when a task ends, says it
//! reads GitHub afresh, lets a reversible decision go on with the safe default, and names
//! only flags the command really has.

mod common;
use std::{fs, path::Path, process::Command};

fn mode_agent() -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/mode-agent.md"),
    )
    .unwrap()
}

#[test]
fn the_agent_loop_runs_lane_next_and_goes_on_with_reversible_defaults() {
    let text = mode_agent();
    for phrase in [
        "moai lane-next --lane <lane> --repo <owner>/<repo> --wait <seconds>",
        "reads GitHub afresh each time",
        "a restart, a usage-limit stop or a compacted conversation misses nothing",
        "go on with the safe default you named",
        "only a `guarded` one waits for the answer",
        "`moai lane-next --wait <seconds>` (at most 3,600",
        "Issue #2831",
        "open it as a draft (`gh pr create --draft …`)",
        "Issue #1975 comment 6870150891",
        "`open_draft_pr`",
        "`DECISION-APPLIED <decision id>`",
        "`card_pr_merged`",
        "Issue #2873",
    ] {
        assert!(text.contains(phrase), "mode-agent.md must say `{phrase}`");
    }
}

/// Every flag the workflow names exists on the command, and the limits it states match.
#[test]
fn the_flags_the_workflow_names_are_the_commands_own() {
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["lane-next", "--help"])
        .output()
        .unwrap();
    let help = String::from_utf8_lossy(&out.stdout);
    for flag in [
        "--lane",
        "--repo",
        "--wait",
        "--tip",
        "--snapshot",
        "--shared-lanes",
    ] {
        assert!(help.contains(flag), "lane-next --help lacks {flag}: {help}");
    }
    assert!(help.contains("at most 3600"), "{help}");
    let text = mode_agent();
    assert!(text.contains("at most 3,600") && text.contains("every 432 seconds"));
    assert!(
        help.contains("1 to 50")
            && text.contains("`--shared-lanes <n>`")
            && text.contains("1 to 50")
    );
}
