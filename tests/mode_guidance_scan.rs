//! SPEC-20261120-mode-guidance-rules: the lane and leader workflows carry the six rules the
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
fn waiting_lanes_take_the_next_card_and_the_leader_dispatches_them() {
    let a = agent();
    has(
        &a,
        "mode-agent.md",
        &[
            "`LANE-WAIT <lane> on=check|audit|merge card=<C-n>`",
            "second worktree",
            "at most two cards at once",
        ],
    );
    let l = leader();
    has(
        &l,
        "mode-leader.md",
        &[
            "`LANE-WAIT on=check|audit|merge`",
            "give it the next ready card",
            "two cards with a worktree",
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
