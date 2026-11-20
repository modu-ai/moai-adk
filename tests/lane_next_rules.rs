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
    // Card C-53 S5: the next task comes from the ledger (`moai card next`) and a wait is
    // `moai wait`; `lane-next` reads Issue claims, so MODE-AGENT no longer names it, nor its
    // `open_draft_pr` and `card_pr_merged` tasks (was: `moai lane-next --lane … --wait`, "reads
    // GitHub afresh each time", its 3,600-second cap, Issue #2831, `open_draft_pr`,
    // `card_pr_merged`, Issue #2873).
    for phrase in [
        "Read the state afresh each time (`moai card next`, your pull requests) — never what you remember",
        "a restart, a usage-limit stop or a compacted conversation misses nothing",
        "go on with the safe default you named",
        "only a `guarded` one waits for the answer",
        "or a background `moai wait`",
        "open it as a draft (`gh pr create --draft …`)",
        "the merge gate marks a draft ready once it has its PASS",
        "`DECISION-APPLIED <decision id>`",
        "A card of yours whose pull request has merged: `card pr`, `card done` and the cleanup",
        "While a pull request with a `moai-tip-repair:` line is open, a red tip is that pull request's author's to repair",
    ] {
        assert!(text.contains(phrase), "mode-agent.md must say `{phrase}`");
    }
    assert!(
        !text.contains("lane-next"),
        "mode-agent.md still names `lane-next`"
    );
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
        "--explain",
    ] {
        assert!(help.contains(flag), "lane-next --help lacks {flag}: {help}");
    }
    assert!(help.contains("at most 3600"), "{help}");
    assert!(help.contains("1 to 50"), "{help}");
    // C-53 S5: MODE-AGENT waits with `moai wait` (was: `lane-next --wait`, its 3,600-second
    // cap, its 432-second read interval and `--shared-lanes`); every flag the lane's watch shows
    // is the command's own.
    let out = Command::new(crate::common::MOAI_BIN)
        .args(["wait", "--help"])
        .output()
        .unwrap();
    assert!(out.status.success());
    let wait = String::from_utf8_lossy(&out.stdout);
    let text = mode_agent();
    for (shown, flag) in [
        ("`moai wait --repo <owner/repo> <kind>`", "--repo"),
        ("`ci --head <h>`", "--head"),
        (
            "`comment --issue <pr> --after <id> --prefix AUDIT`",
            "--prefix",
        ),
        ("`pr --pr <pr> --head <h>`", "--pr"),
    ] {
        assert!(text.contains(shown), "mode-agent.md must show {shown}");
        assert!(wait.contains(flag), "wait --help lacks {flag}: {wait}");
    }
    for flag in ["--issue", "--after"] {
        assert!(wait.contains(flag), "wait --help lacks {flag}: {wait}");
    }
}
