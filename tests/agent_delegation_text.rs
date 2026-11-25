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

/// The chain table's lane row names the three base agents a lane hands a card to, in order, and
/// the rule under it says the lane audits and lands nothing; AUDIT says the implementation agent
/// works in the lane's worktree.
#[test]
fn chains_and_audit_name_the_lane_delegation() {
    let chains = repo_text("skills/moai/references/chains/index.md");
    let row = chains
        .lines()
        .find(|l| l.starts_with("| Work through the cards |"))
        .expect("the chain table has the WORK row");
    assert!(
        row.contains("| WORK | moai-plan, moai-impl, moai-docs |"),
        "the WORK row: {row}"
    );
    let missing = lacking(
        &chains,
        &[
            "hands its plan, implementation and documents to `moai-plan`, `moai-impl` and `moai-docs` in turn",
            "It audits and lands nothing.",
        ],
    );
    assert!(missing.is_empty(), "chains/index.md lacks {missing:?}");
    let audit = repo_text("skills/moai/workflows/audit.md");
    assert!(
        audit.contains("The implementation agent works in the lane's worktree (RUN workflow)."),
        "workflows/audit.md Isolation line"
    );
    assert!(!audit.contains("The implementation agent keeps its worktree"));
}

/// The lane route (C-493 criteria 6 to 9, 11): its steps are one action each and run take,
/// worktree, plan, impl, docs, submit, message, wait; the plan, impl and docs steps hand the work
/// to the base agents; a lane audits and lands nothing; the route says the first check, the
/// branch name, the status line before a wait and the `## Base agents` rules.
#[test]
fn work_route_delegates_plan_impl_and_docs_in_turn() {
    let text = repo_text("skills/moai/workflows/work.md");
    let procedure = text
        .split("## Procedure\n")
        .nth(1)
        .and_then(|rest| rest.split("\n## ").next())
        .expect("work.md has a Procedure section");
    let keys: Vec<&str> = procedure
        .lines()
        .filter_map(|l| l.split_once(". **"))
        .filter(|(n, _)| n.trim().parse::<usize>().is_ok())
        .filter_map(|(_, rest)| rest.split("**").next())
        .collect();
    assert_eq!(
        keys,
        ["take", "worktree", "plan", "impl", "docs", "submit", "message", "wait"]
    );
    let step = |key: &str| {
        procedure
            .lines()
            .find(|l| l.contains(&format!(". **{key}** — ")))
            .unwrap_or_else(|| panic!("no step {key}"))
    };
    for (key, phrase) in [
        ("take", "First run `pwd` and confirm that the folder and the `project_id` of the `moai card next` answer are the project this session was started for"),
        ("take", "Name the branch `spec/<slug>-<yyyymmdd>`"),
        ("take", "`SPEC-20261125-card-no-file-locks` gives `spec/card-no-file-locks-20261125`, and `spec/20261125-card-no-file-locks` is wrong"),
        ("plan", "hands the draft to `moai-plan` after `moai-research` has settled each open fact"),
        ("impl", "hands each planned change to `moai-impl`, one at a time"),
        ("docs", "hands the SPEC Result, the document mismatches and the `@cc:` comments to `moai-docs` after `moai-impl` has returned"),
        ("submit", "Run `moai card submit <card> --generation <g>`"),
        ("submit", "`leftovers:` line and one `audit-tier: 0|1|2 <reason>` line"),
        ("message", "`LANE-DONE <card> <state> head=<10 hex> (<name>)`"),
    ] {
        assert!(step(key).contains(phrase), "step {key} lacks {phrase:?}");
    }
    // A lane audits and lands nothing: no audit or land step, and no `card land` command.
    assert!(!procedure.contains(". **audit** — ") && !procedure.contains(". **land** — "));
    assert!(!text.contains("moai card land"));
    let missing = lacking(
        &text,
        &[
            "Whenever you start a background run or begin to wait (for an agent, a check or the leader's message), first write one line: what is running, what you wait for and what comes next.",
            "## When a step stops",
            "the leader sends `REPAIR <card> fail head=<head>` after a first FAIL",
            "A lane hands each card's plan, implementation and documents to the base agents in turn with the host's agent tool; do not do them yourself when the agents are installed.",
            "then `moai-impl` for each planned change one at a time",
            "They work in this worktree (`moai-impl` sets `isolation: none`)",
        ],
    );
    assert!(missing.is_empty(), "workflows/work.md lacks {missing:?}");
    assert_eq!(text.matches("## Base agents\n").count(), 1);
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
