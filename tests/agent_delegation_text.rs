//! C-493 slice 1 (user decision 2026-11-25): a lane hands a card's plan, implementation and
//! documents to the base agents in turn. These pins keep the delegation sentences of the RUN,
//! PLAN and SYNC workflows and the implementation agent's isolation line from drifting back to
//! "may hand" wording or to a second worktree.

use std::{fs, path::Path};

fn repo_text(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap()
}

fn lacking<'a>(text: &str, phrases: &[&'a str]) -> Vec<&'a str> {
    phrases
        .iter()
        .copied()
        .filter(|p| !text.contains(p))
        .collect()
}

/// The implementation agent sets no isolation: it starts in the worktree its parent works in
/// (the card's worktree) instead of building a second one from nothing.
#[test]
fn impl_agent_uses_the_parents_worktree() {
    let text = repo_text("agents/impl.md");
    let front = text
        .strip_prefix("---\n")
        .and_then(|rest| rest.split_once("\n---\n"))
        .map(|(front, _)| front.to_string())
        .expect("an agent file starts with front matter");
    assert!(front.lines().any(|l| l == "isolation: none"));
    assert!(!front.lines().any(|l| l == "isolation: worktree"));
    let missing = lacking(
        &text,
        &[
            "You work in the worktree your parent works in",
            "never make or enter another worktree",
            "end with `RESULT impl blocked` when it does not",
        ],
    );
    assert!(missing.is_empty(), "agents/impl.md lacks {missing:?}");
}

/// RUN says in the imperative that every planned change goes to `moai-impl`, one at a time, in
/// this worktree, and why no second worktree is made.
#[test]
fn run_delegates_every_change_to_one_impl_at_a_time() {
    let text = repo_text("skills/moai/workflows/run.md");
    let missing = lacking(
        &text,
        &[
            "RUN hands every planned change to `moai-impl` with the host's agent tool; do not write the change yourself when the agent is installed.",
            "Never run two `moai-impl` at once; only `moai-research` runs several at once.",
            "the installed agent file sets `isolation: none`",
            "`isolation: worktree` would make a second worktree from the remote default branch and build it from nothing",
            "wait for that notice before you read its result or start the next step",
        ],
    );
    assert!(missing.is_empty(), "workflows/run.md lacks {missing:?}");
    assert!(
        !text.contains("the installed agent file sets `isolation: worktree`"),
        "workflows/run.md still says the impl agent sets isolation: worktree"
    );
}

/// PLAN fans research out first (several at once), then one `moai-plan`, before `moai-impl`.
#[test]
fn plan_delegates_research_then_one_plan() {
    let text = repo_text("skills/moai/workflows/plan.md");
    let missing = lacking(
        &text,
        &[
            "PLAN hands the draft to one `moai-plan` with the host's agent tool; do not draft it yourself when the agent is installed.",
            "several at once in one message and one question each",
            "wait for its `RESULT` line before RUN starts",
            "`moai-impl` starts after it and never beside it",
        ],
    );
    assert!(missing.is_empty(), "workflows/plan.md lacks {missing:?}");
}

/// SYNC hands the document writing to one `moai-docs`, after `moai-impl` has returned.
#[test]
fn sync_delegates_the_documents_after_impl() {
    let text = repo_text("skills/moai/workflows/sync.md");
    let missing = lacking(
        &text,
        &[
            "to one `moai-docs` with the host's agent tool",
            "do not write them yourself when the agent is installed",
            "It starts after `moai-impl` has returned",
            "never beside `moai-impl`",
        ],
    );
    assert!(missing.is_empty(), "workflows/sync.md lacks {missing:?}");
}
