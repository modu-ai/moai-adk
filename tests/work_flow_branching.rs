//! The lane workflow branches on the profile's work flow
//! (SPEC-20261121-work-flow-branching): an m-flow lane reads no pull request and no
//! `origin/<base>` instruction, the pull request sentences sit in a bounded block, and the
//! rules both flows share stand outside both blocks.

const LANE: &str = include_str!("../skills/moai/workflows/mode-agent.md");
const INTERVIEW: &str = include_str!("../skills/moai/references/internal/interview.md");
const SYNC: &str = include_str!("../skills/moai/workflows/sync.md");

/// The words that open and close an m-flow block.
const M_BEGIN: &str = "M-Flow only (`m-flow`):";
const M_END: &str = "End of M-Flow.";
/// The words that open and close a pull request block; the label names the flows it holds.
const PR_BEGIN: &str = "PR flow only, for `github-flow and the other pull request flows`:";
const PR_END: &str = "End of PR flow.";
/// The five branching steps.
const STEPS: [&str; 5] = ["state", "claim", "pr-gate", "merge-train", "hand-off"];
/// What an m-flow block must not tell a lane to do.
const FORBIDDEN_IN_M_FLOW: [&str; 4] = ["gh pr", "card pr", "origin/<", "--force"];
/// What only a pull request block may say.
const PR_ONLY: [&str; 3] = ["origin/<base", "gh pr ", "card pr "];

/// A numbered step is one physical line, as the other step tests read it.
fn step(key: &str) -> &'static str {
    let mark = format!(". **{key}**");
    LANE.lines()
        .find(|l| l.contains(&mark) && l.starts_with(|c: char| c.is_ascii_digit()))
        .unwrap_or_else(|| panic!("mode-agent.md has no step `{key}`"))
}

/// Every block of one kind in a line, from its opening words to its closing words.
fn blocks(line: &'static str, begin: &str, end: &str) -> Vec<&'static str> {
    let mut found = Vec::new();
    let mut rest = line;
    while let Some(at) = rest.find(begin) {
        let after = &rest[at + begin.len()..];
        let stop = after
            .find(end)
            .unwrap_or_else(|| panic!("a block opened by `{begin}` is not closed by `{end}`"));
        assert!(
            !after[..stop].contains(begin),
            "a block opened by `{begin}` is opened again before it is closed"
        );
        found.push(&after[..stop]);
        rest = &after[stop + end.len()..];
    }
    found
}

/// The line with every block of both kinds taken out: what both flows read.
fn shared(line: &'static str) -> String {
    let mut out = String::new();
    let mut rest = line;
    loop {
        let m = rest.find(M_BEGIN);
        let p = rest.find(PR_BEGIN);
        let (at, begin, end) = match (m, p) {
            (None, None) => break,
            (Some(a), Some(b)) if a < b => (a, M_BEGIN, M_END),
            (Some(a), None) => (a, M_BEGIN, M_END),
            (_, Some(b)) => (b, PR_BEGIN, PR_END),
        };
        out.push_str(&rest[..at]);
        let after = &rest[at + begin.len()..];
        let stop = after
            .find(end)
            .unwrap_or_else(|| panic!("a block opened by `{begin}` is not closed by `{end}`"));
        rest = &after[stop + end.len()..];
    }
    out.push_str(rest);
    out
}

fn m_flow_text(key: &str) -> String {
    blocks(step(key), M_BEGIN, M_END).join(" ")
}

#[test]
fn the_flow_blocks_open_and_close_in_pairs_and_do_not_nest() {
    assert_eq!(
        LANE.matches(M_BEGIN).count(),
        LANE.matches(M_END).count(),
        "an m-flow block is opened and not closed, or the reverse"
    );
    assert_eq!(
        LANE.matches(PR_BEGIN).count(),
        LANE.matches(PR_END).count(),
        "a pull request block is opened and not closed, or the reverse"
    );
    for line in LANE.lines() {
        // `blocks` panics when a block is not closed or opens again before it is closed.
        let m = blocks(line, M_BEGIN, M_END);
        let p = blocks(line, PR_BEGIN, PR_END);
        for block in m.iter().chain(p.iter()) {
            assert!(
                !block.contains(PR_BEGIN) && !block.contains(M_BEGIN),
                "a block holds the opening of another block: {block}"
            );
        }
    }
    for key in STEPS {
        let line = step(key);
        assert!(
            !blocks(line, M_BEGIN, M_END).is_empty() && !blocks(line, PR_BEGIN, PR_END).is_empty(),
            "step `{key}` does not have a block for each flow"
        );
    }
}

#[test]
fn the_state_step_reads_the_flow_and_gives_m_flow_a_local_base() {
    let outside = shared(step("state"));
    assert!(
        outside.contains("`moai work-flow-plan --project .`") && outside.contains("`profile.flow`"),
        "the state step does not read the flow with work-flow-plan outside the blocks"
    );
    assert!(
        outside.contains("is not a flow"),
        "the state step does not say that an empty or failed read is not a flow"
    );
    let text = m_flow_text("state");
    assert!(
        text.contains("integration branch") && text.contains("`HEAD`"),
        "the m-flow state block does not compare HEAD with the integration branch tip: {text}"
    );
}

#[test]
fn the_claim_step_starts_from_the_local_tip_and_names_no_card_number() {
    let text = m_flow_text("claim");
    assert!(
        text.contains("integration branch") && text.contains("worktree"),
        "the m-flow claim block does not start the worktree from the integration branch: {text}"
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
fn the_m_flow_blocks_carry_no_pull_request_or_origin_instruction() {
    for key in STEPS {
        let text = m_flow_text(key);
        for word in FORBIDDEN_IN_M_FLOW {
            assert!(
                !text.contains(word),
                "the m-flow block of `{key}` contains `{word}`: {text}"
            );
        }
    }
    let hand_off = m_flow_text("hand-off");
    assert!(
        hand_off.contains("never force-push"),
        "the m-flow hand-off block does not forbid a force push"
    );
    assert!(
        hand_off.contains("card done"),
        "the m-flow hand-off block does not close the card on the local landing evidence"
    );
    assert!(
        hand_off.contains("--integration-ref <the local integration branch>"),
        "the m-flow hand-off block does not clean up against the local integration branch"
    );
    assert!(
        m_flow_text("pr-gate").contains("no pull request"),
        "the m-flow pr-gate block does not say that no pull request is opened"
    );
    // The other lines that name a pull request carry their m-flow reading in a block too.
    for anchor in [
        "Pull requests, CI, the audit and the merge gate stay on GitHub",
        "**keep-going**",
        "**self-pick**",
    ] {
        let line = LANE
            .lines()
            .find(|l| l.contains(anchor))
            .unwrap_or_else(|| panic!("mode-agent.md has no line with `{anchor}`"));
        let text = blocks(line, M_BEGIN, M_END).join(" ");
        assert!(
            !text.is_empty(),
            "the line with `{anchor}` has no m-flow block"
        );
        for word in FORBIDDEN_IN_M_FLOW {
            assert!(
                !text.contains(word),
                "the m-flow block of the line with `{anchor}` contains `{word}`: {text}"
            );
        }
    }
}

#[test]
fn the_pull_request_sentences_stay_inside_their_blocks() {
    for key in STEPS {
        let outside = shared(step(key));
        for word in PR_ONLY {
            assert!(
                !outside.contains(word),
                "step `{key}`: `{word}` stands outside the pull request block: {outside}"
            );
        }
        assert_eq!(
            step(key)
                .matches("github-flow and the other pull request flows")
                .count(),
            blocks(step(key), PR_BEGIN, PR_END).len(),
            "step `{key}` names the label of the pull request flows outside a block"
        );
    }
}

#[test]
fn the_rules_both_flows_share_stand_outside_both_blocks() {
    // The sentences the audit named, by step: each is read by an m-flow lane and by a lane
    // of a pull request flow, so none of them may sit inside a block.
    let common: [(&str, &[&str]); 5] = [
        (
            "state",
            &[
                "Summarise the current state read-only from `moai card next`",
                "Report anything unread as unread",
                "Read the standing goal with `moai goal-show <project>`",
                "A lane only reads the goal",
            ],
        ),
        (
            "claim",
            &[
                "A dispatch is a hint, never a grant",
                "then take the card in it",
                "Take in the folder you will work in and never move it afterwards",
                "On `conflict`",
                "Keep the lease: run `card renew <card> --generation <g>`",
                "Progress never waits on the leader's reply",
                "A busy ledger (`io`, `retryable`)",
            ],
        ),
        (
            "pr-gate",
            &[
                "Before every push, also run the checks that caught pull requests only in CI",
                "go on with the round and take the next card (`self-pick` step)",
            ],
        ),
        ("merge-train", &[]),
        (
            "hand-off",
            &[
                "Before `card review`, mark each criterion of the card",
                "an audit at audit-tier 0 or 1 is started as `self-pick` (2) says",
                "so renew its lease while it is in `review`",
                "On a FAIL: `card reopen <card> --generation <g>`",
                "The leader only confirms a lane's closing",
                "A `skip` or a `trim` with `ignored_kept` paths is reported on the card, never forced",
                "Empty the card's build folder (`internal/build-folder`)",
            ],
        ),
    ];
    for (key, sentences) in common {
        let outside = shared(step(key));
        for sentence in sentences {
            assert!(
                outside.contains(sentence),
                "step `{key}`: the shared rule `{sentence}` is not outside both blocks"
            );
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
        blocks(intro, PR_BEGIN, PR_END)
            .iter()
            .any(|b| b.contains("Pull requests, CI, the audit and the merge gate stay on GitHub")),
        "the introduction does not hold the GitHub sentence in a pull request block"
    );
    assert!(
        blocks(intro, M_BEGIN, M_END)
            .iter()
            .any(|b| b.contains("a formal release")),
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
    let text = m_flow_text("merge-train");
    assert!(
        text.contains("`sync` workflow") && text.contains("step 5"),
        "the m-flow merge-train block does not point to step 5 of the sync workflow: {text}"
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

#[test]
fn the_sync_workflow_names_the_branch_like_the_lane_workflow() {
    let step5 = SYNC
        .lines()
        .find(|l| l.starts_with("5. **Pull request.**"))
        .expect("sync.md has step 5");
    assert!(
        !step5.contains("its own card branch"),
        "sync.md step 5 still calls the branch a card branch"
    );
    assert!(
        step5.contains("`spec/<slug>-<yyyymmdd>`"),
        "sync.md step 5 does not name the branch spec/<slug>-<yyyymmdd>"
    );
}
