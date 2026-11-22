//! The leader round arms the loop command only in a session confirmed as the leader
//! (SPEC-20261121-loop-leader-session-only). Skill text only: the lane's loop line
//! stays as it is.

const MODE_LEADER: &str = include_str!("../skills/moai/workflows/mode-leader.md");
const ORCA: &str = include_str!("../skills/moai/references/adapters/orca.md");

fn mode_channel_step() -> &'static str {
    MODE_LEADER
        .lines()
        .find(|line| line.starts_with("1. **mode-channel**"))
        .expect("mode-leader.md has the mode-channel step")
}

#[test]
fn the_leader_workflow_arms_the_loop_only_in_a_confirmed_leader_session() {
    let step = mode_channel_step();
    // REQ-1: the loop is armed only after the session confirmed it is the leader.
    assert!(
        step.contains("Arm `/loop` only in a session that has confirmed it is the leader"),
        "mode-channel does not limit the loop to a confirmed leader session"
    );
    // REQ-2: what confirms it.
    assert!(
        step.contains("entered with `--mode-leader`"),
        "mode-channel does not define a confirmed leader session"
    );
    // REQ-3: an unconfirmed session does not arm it.
    assert!(
        step.contains("does not arm the leader round's `/loop`"),
        "mode-channel does not refuse the loop to an unconfirmed session"
    );
    // The leader still runs under the loop: the limit does not remove it.
    assert!(
        step.contains("`/loop 5m /moai --mode-leader`"),
        "the leader round must still name its loop line"
    );
}

#[test]
fn the_orca_adapter_says_the_same_for_the_leader_row() {
    let row = ORCA
        .lines()
        .find(|line| line.starts_with("| MODE-LEADER |"))
        .expect("orca.md has the MODE-LEADER row");
    assert!(
        row.contains("armed only in a confirmed leader session"),
        "the MODE-LEADER row does not say the loop is armed only in a confirmed leader session"
    );
    assert!(
        row.contains("Claude Code `/loop 5m`"),
        "the MODE-LEADER row must still name the timer"
    );
}

#[test]
fn the_lane_loop_line_is_left_as_it_is() {
    // REQ-5: the lane workflow keeps its wake rules; this change does not touch them.
    let lane = include_str!("../skills/moai/workflows/mode-agent.md");
    assert!(
        lane.contains("`/loop 5m /moai --mode-agent`"),
        "the lane workflow no longer names its fixed loop line"
    );
}
