//! The lane workflow branches on the profile's work flow
//! (SPEC-20261121-work-flow-branching): an m-flow lane reads no pull request and no
//! `origin/<base>` instruction, and the pull request sentences stay under their label.

const LANE: &str = include_str!("../skills/moai/workflows/mode-agent.md");
const INTERVIEW: &str = include_str!("../skills/moai/references/internal/interview.md");

/// The label every pull request sentence of a branching step stays under.
const LABEL: &str = "github-flow and the other pull request flows";
/// The words that open the m-flow clause of a branching step.
const M_FLOW: &str = "With `m-flow`";
/// The five branching steps.
const STEPS: [&str; 5] = ["state", "claim", "pr-gate", "merge-train", "hand-off"];
/// What an m-flow clause must not tell a lane to do.
const FORBIDDEN_IN_M_FLOW: [&str; 4] = ["gh pr", "card pr", "origin/<", "--force"];

/// A numbered step is one physical line, as the other step tests read it.
fn step(key: &str) -> &'static str {
    let mark = format!(". **{key}**");
    LANE.lines()
        .find(|l| l.contains(&mark) && l.starts_with(|c: char| c.is_ascii_digit()))
        .unwrap_or_else(|| panic!("mode-agent.md has no step `{key}`"))
}

/// The m-flow clause of a step: from its opening words to the label of the pull request flows.
fn m_flow_clause(key: &str) -> &'static str {
    let line = step(key);
    let start = line
        .find(M_FLOW)
        .unwrap_or_else(|| panic!("step `{key}` has no m-flow clause"));
    let end = line
        .find(LABEL)
        .unwrap_or_else(|| panic!("step `{key}` has no label for the pull request flows"));
    assert!(
        start < end,
        "step `{key}`: the m-flow clause comes before the label"
    );
    &line[start..end]
}

#[test]
fn the_state_step_reads_the_flow_and_gives_m_flow_a_local_base() {
    let state = step("state");
    assert!(
        state.contains("`moai work-flow-plan --project .`") && state.contains("`profile.flow`"),
        "the state step does not read the flow with work-flow-plan"
    );
    assert!(
        state.contains("is not a flow"),
        "the state step does not say that an empty or failed read is not a flow"
    );
    let clause = m_flow_clause("state");
    assert!(
        clause.contains("integration branch") && clause.contains("`HEAD`"),
        "the m-flow state clause does not compare HEAD with the integration branch tip: {clause}"
    );
}

#[test]
fn the_claim_step_starts_from_the_local_tip_and_names_no_card_number() {
    let clause = m_flow_clause("claim");
    assert!(
        clause.contains("integration branch") && clause.contains("worktree"),
        "the m-flow claim clause does not start the worktree from the integration branch: {clause}"
    );
    assert!(
        step("claim").contains("spec/<slug>-<yyyymmdd>"),
        "the claim step does not name the branch spec/<slug>-<yyyymmdd>"
    );
    assert!(
        !LANE.contains("card/<n>"),
        "mode-agent.md still names a branch with the card number"
    );
}

#[test]
fn the_m_flow_clauses_carry_no_pull_request_or_origin_instruction() {
    for key in STEPS {
        let clause = m_flow_clause(key);
        for word in FORBIDDEN_IN_M_FLOW {
            assert!(
                !clause.contains(word),
                "the m-flow clause of `{key}` contains `{word}`: {clause}"
            );
        }
    }
    let hand_off = m_flow_clause("hand-off");
    assert!(
        hand_off.contains("never force-push"),
        "the m-flow hand-off clause does not forbid a force push"
    );
    assert!(
        hand_off.contains("card done"),
        "the m-flow hand-off clause does not close the card on the local landing evidence"
    );
    assert!(
        m_flow_clause("pr-gate").contains("no pull request"),
        "the m-flow pr-gate clause does not say that no pull request is opened"
    );
}

#[test]
fn the_pull_request_sentences_stay_under_their_label() {
    for key in STEPS {
        let line = step(key);
        let label_at = line
            .find(LABEL)
            .unwrap_or_else(|| panic!("step `{key}` has no label for the pull request flows"));
        assert_eq!(
            line.matches(LABEL).count(),
            1,
            "step `{key}` names the label more than once"
        );
        for word in ["origin/<base", "gh pr create", "card pr"] {
            for (at, _) in line.match_indices(word) {
                assert!(
                    at > label_at,
                    "step `{key}`: `{word}` sits before the label of the pull request flows"
                );
            }
        }
    }
}

#[test]
fn the_pull_request_only_at_a_formal_release_is_said_in_both_places() {
    let intro = LANE
        .lines()
        .find(|l| l.contains("Pull requests, CI, the audit and the merge gate stay on GitHub"))
        .expect("mode-agent.md keeps the sentence that pull requests stay on GitHub");
    assert!(
        intro.contains("a formal release"),
        "the introduction line does not say that m-flow opens a pull request only for a formal release"
    );
    let entry = INTERVIEW
        .lines()
        .find(|l| l.starts_with("- **m-flow**"))
        .expect("interview.md has the m-flow entry");
    assert!(
        entry.contains("a formal release"),
        "the m-flow entry does not say that a pull request is opened only for a formal release"
    );
    assert!(
        !entry.contains("one pull request"),
        "the m-flow entry still says one pull request per SPEC"
    );
}

#[test]
fn the_m_flow_merge_text_is_pointed_to_and_not_copied() {
    let clause = m_flow_clause("merge-train");
    assert!(
        clause.contains("`sync` workflow") && clause.contains("step 5"),
        "the m-flow merge-train clause does not point to step 5 of the sync workflow: {clause}"
    );
    for copied in [
        "refs/moai/lane-merge",
        "moai-op pr_create",
        "pr-body-render",
    ] {
        assert!(
            !LANE.contains(copied),
            "mode-agent.md copies the canonical M-Flow text (`{copied}`)"
        );
    }
}
