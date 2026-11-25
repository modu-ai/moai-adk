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

/// The build and cleanup rules of the RUN workflow (it carries the retired VERIFY route,
/// SPEC-20261122-next-flow-skills-agents) live in `references/internal/build-slots.md` since
/// SPEC-20261125-run-sync-procedure-receipt.
fn build_slots() -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai/references/internal/build-slots.md"),
    )
    .unwrap()
}

#[test]
fn verify_states_the_safe_ignored_rule_and_leaves_bulk_cleanup_to_the_leader() {
    // The commands are the ones `moai worktree-gc-plan --help` and `moai worktree-gc-apply --help`
    // list: plan takes `--worktree <PATH>` (repeatable), `--max-checked` and `--cursor`; apply
    // takes `--integration-ref`, `--plan-digest` and `--permission-reference` and has no
    // `--force`. The old line `moai worktree-gc-apply --worktree <path>` is not in the text any
    // more: a lane previews its own worktree with the plan, then applies the plan's digest.
    let slots = build_slots();
    says(
        &slots,
        "build-slots.md",
        &[
            "a lane removes its own merged card worktree",
            "`--worktree <path>` (repeatable) plans only that worktree",
            "`moai worktree-gc-apply <project> --integration-ref <ref> --plan-digest <digest> --permission-reference <ref>`",
            "(no `--force`; branches are kept)",
            "byte-identical copies of the primary checkout's same path",
            "stays the leader's periodic step: `moai worktree-gc-plan --max-checked N --cursor <next_cursor>",
        ],
    );
    for (file, text) in [("build-slots.md", &slots), ("run.md", &workflow("run.md"))] {
        assert!(
            !text.contains("a worktree itself is removed only after a separate confirmation"),
            "{file} still carries the rule #2879 replaced"
        );
    }
}

// Retired with the MODE-LEADER route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_leader_runs_the_same_plan_periodically_with_a_cursor` pinned the leader's periodic disk hygiene.

#[test]
fn verify_names_the_build_folder_warning() {
    says(
        &build_slots(),
        "build-slots.md",
        // Card C-82 follow-up: only a folder shared with or inside another worktree (or one that
        // names a retired build folder in the home) is named (was: "with any other
        // `CARGO_TARGET_DIR`, unset included").
        &[
            "When the slot command wraps `cargo` with a `CARGO_TARGET_DIR` shared with another worktree, inside another worktree,",
            "it prints one warning line saying to unset the variable",
        ],
    );
}
