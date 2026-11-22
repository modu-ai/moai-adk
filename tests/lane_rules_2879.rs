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

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `a_lane_cleans_its_card_after_the_closing_table` pinned the lane's card cleanup after its pull request merged.

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_never_delete_rule_names_the_card_cleanup_as_its_exception` pinned the lane's never-delete exceptions, which left with the route.

/// The RUN workflow carries the retired VERIFY route (SPEC-20261122-next-flow-skills-agents).
#[test]
fn verify_states_the_safe_ignored_rule_and_leaves_bulk_cleanup_to_the_leader() {
    let run = workflow("run.md");
    says(
        &run,
        "run.md",
        &[
            "removes its own merged card worktree with `moai worktree-gc-apply --worktree <path>` (Issue #2879",
            "byte-identical copies of the primary checkout's same path",
            "stays the leader's periodic step (`worktree-gc-plan --max-checked N --cursor <next_cursor>`",
        ],
    );
    assert!(
        !run.contains("a worktree itself is removed only after a separate confirmation"),
        "run.md still carries the rule #2879 replaced"
    );
}

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_leader_runs_the_same_plan_periodically_with_a_cursor` pinned the leader's periodic disk hygiene.

#[test]
fn verify_names_the_build_folder_warning() {
    let run = workflow("run.md");
    says(
        &run,
        "run.md",
        // Card C-82 follow-up: only a folder shared with or inside another worktree is named
        // (was: "with any other `CARGO_TARGET_DIR`, unset included").
        &["prints one warning line when it wraps `cargo` with a `CARGO_TARGET_DIR` shared with another worktree or inside another worktree"],
    );
}
