//! SPEC-20261124-leader-withdraw-rule: the leader rules tell the leader to keep a card that carries
//! a HOLD or an unmet prerequisite out of ready through `moai card withdraw`. The test reads one
//! file; no host, no HOME and no command runs here.

use std::fs;

const LEADER_RULES: &str = concat!(env!("CARGO_MANIFEST_DIR"), "/docs/leader-rules.md");

/// REQ-1, REQ-2: item 6 of "What a round does" names the command and the condition.
#[test]
fn hold_cards_leave_ready_through_card_withdraw() {
    let text = fs::read_to_string(LEADER_RULES).unwrap();
    let item = text
        .lines()
        .find(|l| l.starts_with("6. Keep HOLD cards out of ready"))
        .unwrap_or_else(|| panic!("no item starts with `6. Keep HOLD cards out of ready`"));
    for phrase in ["`moai card withdraw`", "HOLD or an unmet prerequisite"] {
        assert!(item.contains(phrase), "the item lacks {phrase}: {item}");
    }
}
