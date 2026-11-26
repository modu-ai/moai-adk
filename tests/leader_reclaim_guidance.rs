//! SPEC-20261124-leader-withdraw-rule: the leader rules tell the leader to keep a card that carries
//! a HOLD or an unmet prerequisite out of ready through `moai card withdraw`. The test reads one
//! file; no host, no HOME and no command runs here. SPEC-20261125-land-sync-merge: the same file
//! also pins that the leader rules name `moai card land-sync <branch>` and `work_flow.sync_sources`
//! and carry no card number.

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

/// SPEC-20261125-land-sync-merge AC-8: the leader rules name the command and the profile key and
/// carry no card number (a card number is the letter C, a hyphen and digits that do not continue
/// a longer word, so `SPEC-2026...` is not one).
#[test]
fn leader_rules_name_land_sync_and_its_key_without_a_card_number() {
    let text = fs::read_to_string(LEADER_RULES).unwrap();
    let text = text.split_whitespace().collect::<Vec<_>>().join(" ");
    for s in ["`moai card land-sync <branch>", "`work_flow.sync_sources`"] {
        assert!(text.contains(s), "docs/leader-rules.md lacks {s}");
    }
    if let Some(found) = card_number(&text) {
        panic!("docs/leader-rules.md names a card number: {found}");
    }
}

/// The first card number in `text`: `C-` and a digit, where the `C` does not end a longer word.
fn card_number(text: &str) -> Option<String> {
    let bytes = text.as_bytes();
    bytes.windows(3).enumerate().find_map(|(i, w)| {
        let after_word = i > 0 && bytes[i - 1].is_ascii_alphanumeric();
        (w[0] == b'C' && w[1] == b'-' && w[2].is_ascii_digit() && !after_word)
            .then(|| String::from_utf8_lossy(&bytes[i..(i + 8).min(bytes.len())]).into_owned())
    })
}

/// The card-number scan refuses a card number and lets the lookalikes through.
#[test]
fn card_number_scan_refuses_a_card_number_and_passes_a_spec_id() {
    assert_eq!(card_number("see (C-523) here").as_deref(), Some("C-523) h"));
    assert_eq!(card_number("C-7").as_deref(), Some("C-7"));
    assert_eq!(card_number("SPEC-20261125-land-sync-merge"), None);
    assert_eq!(card_number("a card `C-n` or C-"), None);
}
