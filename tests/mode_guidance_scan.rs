//! SPEC-20261120-mode-guidance-rules: the lane and leader workflows carry the rules the
//! day's supervision showed were missing. The test reads the two workflow files by path and
//! checks the named words; it never walks the repository.

use std::fs;
use std::path::Path;

/// The file's text with every run of whitespace turned into one space, so a rewrapped line does
/// not break a check.
fn flat(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(rel);
    let text = fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

fn agent() -> String {
    flat("skills/moai/workflows/mode-agent.md")
}

fn leader() -> String {
    flat("skills/moai/workflows/mode-leader.md")
}

fn has(text: &str, what: &str, words: &[&str]) {
    for w in words {
        assert!(text.contains(w), "{what} is missing `{w}`");
    }
}

fn lacks(text: &str, what: &str, words: &[&str]) {
    for w in words {
        assert!(!text.contains(w), "{what} still has `{w}`");
    }
}

#[test]
fn agent_runs_long_commands_in_the_background() {
    let t = agent();
    has(
        &t,
        "mode-agent.md",
        &[
            "## Long commands",
            "longer than two minutes",
            "host background task",
            "timeout of at most 120 seconds",
        ],
    );
    // REQ-2: the old way to wait is gone.
    lacks(
        &t,
        "mode-agent.md",
        &[
            "run `moai wait` in the foreground",
            "in the foreground and do not end the turn",
        ],
    );
    has(
        &t,
        "mode-agent.md",
        &["run `moai wait` as a background task"],
    );
}

#[test]
fn waiting_is_switching_and_the_leader_answers_it() {
    let a = agent();
    has(
        &a,
        "mode-agent.md",
        &[
            "## Waiting is switching",
            "Never wait.",
            "(1) leave a `RESUME` note on the card",
            "and stop your own background jobs and watches",
            "`moai lane-safe-to-reset` answers `not_safe` while one runs",
            "(2) give the reservation back with `card release`",
            "A card in `review` is never named in a `SWITCH` line: `card release` refuses it",
            "a new session cannot take it until its lease ends",
            "A wait that one of your own jobs carries is not switched away",
            "after arming the wake of the `wake` step",
            "or (iii) the turn ends on the `SWITCH` line of `Waiting is switching`",
            "A turn that ends on a `SWITCH` line arms this wake first",
            "no card of yours is in `review`",
            "(3) print `SWITCH <card> → <next>` as the last line of the screen",
            "(4) end the turn on that line and wait for the leader to send `/clear` (Claude Code) or `/new` (Codex CLI)",
            "The `SWITCH` line is the only line a turn may end on while a card waits",
            "If no `/clear` or `/new` has arrived by your next wake, go on with `self-pick`",
            "follow `Waiting is switching`",
            "the leader dispatches the earlier card again",
            "reads its `RESUME` note and takes the card again",
            "at most three cards at once and renews the lease of each",
        ],
    );
    let l = leader();
    has(
        &l,
        "mode-leader.md",
        &[
            "posted `SWITCH <card> → <next>` gets `/clear` (Claude Code) or `/new` (Codex CLI)",
            "`moai lane-safe-to-reset` answers `safe` and the card the line names is released",
            "A lane that holds a card in `review` gets no `/clear` or `/new`",
            "dispatch the earlier card again",
            "count the `SWITCH` lines in the round summary",
        ],
    );
}

#[test]
fn compact_request_is_posted_by_the_lane_and_answered_by_the_leader() {
    let a = agent();
    has(
        &a,
        "mode-agent.md",
        &[
            "`LANE-COMPACT <lane> card=<C-n> resume=<note ref>`",
            "`ACK <lane> resume <ref> <UTC>`",
        ],
    );
    let l = leader();
    has(
        &l,
        "mode-leader.md",
        &[
            "`LANE-COMPACT`",
            "`/compact`",
            "`moai lane-safe-to-reset`",
            "`ACK … resume`",
        ],
    );
}

#[test]
fn no_hand_written_times_in_notes() {
    let t = agent();
    has(
        &t,
        "mode-agent.md",
        &[
            "Never type a time into a card note, a report line or a resume note",
            "paste the output of `date -u`",
            "leave the time out",
        ],
    );
}

#[test]
fn supervision_tools_are_named_in_the_leader_workflow() {
    let t = leader();
    has(
        &t,
        "mode-leader.md",
        &[
            "`moai lane-screen --scan`",
            "`moai card timeline <card>`",
            "follow MODE-AGENT `Long commands`",
        ],
    );
}

#[test]
fn session_reset_follows_every_card_and_compact_follows_a_full_context() {
    let a = agent();
    has(
        &a,
        "mode-agent.md",
        &[
            "## Session reset",
            "One context does not carry many cards",
            "post `LANE-DONE`, then wait for the leader's `/clear`",
            "If no `/clear` has arrived by your next wake",
            "Claude Code: `/clear` starts a new conversation with empty context",
            "`/compact [instructions]` summarizes the conversation so far",
            "Codex CLI: `/new` starts a new chat in the same session",
            "(`/clear` also starts one and clears the terminal view first)",
            "`/compact` summarizes the visible chat to free tokens",
            "At 80% or more of the context figure the host shows, post `LANE-COMPACT <lane>",
            "After a `/clear` the session id is new",
            "after writing it, follow `Session reset` and go on in the same turn",
        ],
    );
    // The old rule sent a lane to its next pick right after LANE-DONE, in the same context.
    lacks(
        &a,
        "mode-agent.md",
        &["after writing it, go to the next pick in the same turn"],
    );
    // The old way to spend a wait took another job at once, which no `/clear` could precede.
    lacks(
        &a,
        "mode-agent.md",
        &[
            "take other work in the same turn",
            "take the next job in the same turn",
            "Then start the next round without waiting for the merge",
            "never wait and never end the turn",
            "keep renewing the lease of a card in `review` while this session lives",
            "stays held and is named with `review`",
        ],
    );
    let l = leader();
    has(
        &l,
        "mode-leader.md",
        &[
            "gets `/clear` (Claude Code) or `/new` (Codex CLI) once `moai lane-safe-to-reset` answers `safe`, and its next dispatch as the first message after the reset",
            "Claude Code and Codex CLI both take `/clear` or `/new` and `/compact`",
        ],
    );
}

#[test]
fn the_waiting_steps_come_in_order_and_the_safety_words_stay() {
    let a = agent();
    // (1) note, (2) release, (3) SWITCH line, (4) end the turn on it: a swapped order is a bug.
    let at = |needle: &str| {
        a.find(needle)
            .unwrap_or_else(|| panic!("mode-agent.md is missing `{needle}`"))
    };
    assert!(at("(1) leave a `RESUME` note") < at("(2) give the reservation back"));
    assert!(at("(2) give the reservation back") < at("(3) print `SWITCH"));
    assert!(at("(3) print `SWITCH") < at("(4) end the turn on that line"));
    has(
        &a,
        "mode-agent.md",
        &[
            "nothing uncommitted or unpushed",
            "once the resume note is current",
            "`moai check affected`, an audit agent, `moai wait`",
            "go on with other work of the card in the same turn, or follow `Waiting is switching` when it has none",
            "once the card has nothing else to do, follow `Waiting is switching`",
        ],
    );
    let l = leader();
    has(
        &l,
        "mode-leader.md",
        &[
            "count it as working only after you have read its `ACK … resume` line",
            "its context is at 80% or more at a card boundary",
            "then give the lane `/clear` (Claude Code) or `/new` (Codex CLI) and its next task",
        ],
    );
    lacks(
        &l,
        "mode-leader.md",
        &[
            "its context is at 85% or more",
            "or let the lane continue its own pick order (MODE-AGENT workflow, `self-pick` step)",
            "takes its own next card (MODE-AGENT `keep-going`)",
        ],
    );
}

#[test]
fn repair_evidence_is_asked_for_in_the_report_step() {
    let t = agent();
    has(
        &t,
        "mode-agent.md",
        &[
            "A repair READY line",
            "the reproduction before and after the fix",
            "under `Not done`",
            "the handoff note you wrote",
        ],
    );
}
