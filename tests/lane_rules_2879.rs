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
            // C-82 (user decision 2026-11-17): one build folder per worktree under
            // `~/.moai/build/`, emptied when the card is done (was: "Keep the lane's one build
            // folder (`moai lane-env`)" and the self-audit building into that lane folder).
            "Empty the card's build folder (`internal/build-folder`).",
            "Every build, a self-audit's too, goes to the worktree's build folder (`internal/build-folder`).",
        ],
    );
    assert!(
        !agent.contains("lane-env"),
        "mode-agent.md still names the per-lane folder"
    );
}

/// Audit C1 on PR #2896: the blanket "never delete" rule names the card cleanup as its
/// exception, so the two instructions no longer contradict each other. C-53 S5 repair N8: the
/// unused worktree and branch of a refused `card take` is the second exception, and the
/// `claim` step names the two commands and "that branch only".
#[test]
fn the_never_delete_rule_names_the_card_cleanup_as_its_exception() {
    let agent = workflow("mode-agent.md");
    says(
        &agent,
        "mode-agent.md",
        // C-53 S5: the cleanup is self-pick item (3) since the pre-audit item left (was (4)).
        &[
            "Never delete files, branches or worktrees — except the cleanup in `self-pick` (3): `moai worktree-gc-apply` on your own merged card's worktree, with that plan's `plan_digest` (Issue #2879), and a refused take's unused worktree and branch (`claim` step) — never touch the real HOME",
            "On `conflict` (held by another session, a scope overlap, not `ready`) remove the unused worktree and its branch (`git worktree remove <w>`, `git branch -D card/<n>-<slug>`, that branch only), then take the next candidate",
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

#[test]
fn verify_names_the_build_folder_warning() {
    let verify = workflow("verify.md");
    says(
        &verify,
        "verify.md",
        &["prints one warning line when it wraps `cargo` with any other `CARGO_TARGET_DIR`, unset included"],
    );
}
