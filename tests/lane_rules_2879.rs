//! #2879 (disk hygiene, D1·D4·AC-3): the shipped workflows require a lane to remove its own
//! merged card worktree with `worktree-gc` after the closing table, to audit in one reused
//! detached worktree, and the leader to run the same plan periodically with a cursor.
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
fn a_lane_cleans_its_card_after_the_closing_table() {
    let agent = workflow("mode-agent.md");
    says(
        &agent,
        "mode-agent.md",
        &[
            "clean up that card (Issue #2879, required)",
            "`moai worktree-gc-plan --integration-ref origin/<base> --worktree <the card's worktree> .`",
            "then `moai worktree-gc-apply` with the same arguments, its `plan_digest`",
            "never forced",
            "Keep the lane's one build folder (`moai lane-env`)",
            "never in a new worktree per audit (Issue #2879)",
        ],
    );
}

#[test]
fn verify_states_the_safe_ignored_rule_and_leaves_bulk_cleanup_to_the_leader() {
    let verify = workflow("verify.md");
    says(
        &verify,
        "verify.md",
        &[
            "removes its own merged card worktree with `moai worktree-gc-apply --worktree <path>` (Issue #2879",
            "byte-identical copies of the primary checkout's same path",
            "stays the leader's periodic step (`worktree-gc-plan --max-checked N --cursor <next_cursor>`",
        ],
    );
    assert!(
        !verify.contains("a worktree itself is removed only after a separate confirmation"),
        "verify.md still carries the rule #2879 replaced"
    );
}

#[test]
fn the_leader_runs_the_same_plan_periodically_with_a_cursor() {
    let leader = workflow("mode-leader.md");
    says(
        &leader,
        "mode-leader.md",
        &[
            "Disk hygiene is part of supervision (Issue #2879)",
            "`moai worktree-gc-plan --integration-ref origin/<base> --max-checked <N> --cursor <the last plan's next_cursor> .`",
            "while free space is under the heavy-slot limit",
            "Record the merged-worktree count and free space before and after",
        ],
    );
}
