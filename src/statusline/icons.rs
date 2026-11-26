//! The card-flow icon set, written once (SPEC-20261124-card-flow-icons REQ-1).
//!
//! One card walks `📋 → 🔨 → 👀 → 🚀 → ✅`; `⛔ ⏸️ 🙋` stand beside that road. The status line
//! (the flow items and the icon of the card in hand), the progress board of the output style and
//! the lane reports all name these icons, and every writer and every test reads them from here:
//! a changed icon is a change of this file, the pinned render of the style text and the goldens.
//! `🔴` and `❌` are no card states; the board keeps them beside the set ([`EXTRA`]).

/// One icon of the set: its key, the icon, the words of the flow sentence and the hint on its
/// line of the sample board.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Icon {
    pub key: &'static str,
    pub icon: &'static str,
    /// What the icon says in the flow sentence and the beside sentence.
    pub meaning: &'static str,
    /// What follows `<step> ·` on the sample board line of the icon.
    pub hint: &'static str,
}

/// A card added and waiting its turn.
pub const TODO: &str = "📋";
/// A lane works on the card.
pub const WORK: &str = "🔨";
/// The head was handed back and an auditor checks it.
pub const CHECK: &str = "👀";
/// The check passed: the card waits to land or to be closed.
pub const READY: &str = "🚀";
/// Landed or closed.
pub const DONE: &str = "✅";
/// The lane asked for a decision.
pub const BLOCKED: &str = "⛔";
/// Withdrawn or folded: no one owes it anything. Written with the variation selector U+FE0F (the
/// emoji presentation, two cells wide), counted by `width::display_width` like `🗂️` was.
pub const PAUSED: &str = "⏸️";
/// A user's pick is needed.
pub const QUESTION: &str = "🙋";
/// A risk (board only).
pub const RISK: &str = "🔴";
/// Failed or cancelled (board only).
pub const FAILED: &str = "❌";

/// The road of one card, in order.
pub const CARD_FLOW: [Icon; 5] = [
    Icon {
        key: "todo",
        icon: TODO,
        meaning: "to do, waiting its turn",
        hint: "to do, waiting its turn",
    },
    Icon {
        key: "work",
        icon: WORK,
        meaning: "work in progress",
        hint: "what is happening now",
    },
    Icon {
        key: "check",
        icon: CHECK,
        meaning: "in check (audit)",
        hint: "in check (audit)",
    },
    Icon {
        key: "ready",
        icon: READY,
        meaning: "check passed, waiting to land",
        hint: "check passed, waiting to land",
    },
    Icon {
        key: "done",
        icon: DONE,
        meaning: "done (landed or closed)",
        hint: "done: landed or closed",
    },
];

/// What stands beside the road: the order of the status line is the order here.
pub const BESIDE: [Icon; 3] = [
    Icon {
        key: "blocked",
        icon: BLOCKED,
        meaning: "blocked",
        hint: "what it is blocked on",
    },
    Icon {
        key: "paused",
        icon: PAUSED,
        meaning: "paused (withdrawn or folded)",
        hint: "paused, and why",
    },
    Icon {
        key: "question",
        icon: QUESTION,
        meaning: "needs the user's pick",
        hint: "the pick it needs from you",
    },
];

/// The two icons of the board that are no card state.
pub const EXTRA: [Icon; 2] = [
    Icon {
        key: "risk",
        icon: RISK,
        meaning: "risk",
        hint: "the risk and what it threatens",
    },
    Icon {
        key: "failed",
        icon: FAILED,
        meaning: "failed or cancelled",
        hint: "failed or cancelled, and why",
    },
];

/// The icons of the earlier card flow that left with this set: the old flow
/// (`🗂️ 🟡 📨 🛬 🚧`) and the audit and the repair icons that have a twin in the road now
/// (`🔍 🔁`). No shipped text holds one (SPEC-20261124-card-flow-icons REQ-10).
pub const RETIRED_FLOW: [&str; 7] = ["🗂️", "🟡", "📨", "🛬", "🚧", "🔍", "🔁"];

/// Every icon that left the board: [`RETIRED_FLOW`] and the earlier board's `⬜ 🟢 🔵`.
pub const RETIRED: [&str; 10] = ["🗂️", "🟡", "📨", "🛬", "🚧", "🔍", "🔁", "⬜", "🟢", "🔵"];

/// The lines of the sample board in the order the board keeps them: what needs the user, what is
/// finished, the rest from the nearest to done, then what failed.
pub const BOARD_ORDER: [&str; 10] = [
    QUESTION, RISK, DONE, READY, CHECK, WORK, BLOCKED, PAUSED, TODO, FAILED,
];

/// The icons that carry a count on the status line: the road and what stands beside it.
pub fn status_line_icons() -> impl Iterator<Item = &'static str> {
    CARD_FLOW.iter().chain(BESIDE.iter()).map(|i| i.icon)
}

fn find(icon: &str) -> Option<&'static Icon> {
    CARD_FLOW
        .iter()
        .chain(BESIDE.iter())
        .chain(EXTRA.iter())
        .find(|i| i.icon == icon)
}

/// The flow as one sentence: `📋 to do, waiting its turn → 🔨 … → ✅ done (landed or closed)`.
pub fn flow_sentence() -> String {
    CARD_FLOW
        .iter()
        .map(|i| format!("{} {}", i.icon, i.meaning))
        .collect::<Vec<_>>()
        .join(" → ")
}

/// What stands beside the flow as one sentence: `Beside the flow: ⛔ blocked · … · ❌ failed or
/// cancelled.`
pub fn beside_sentence() -> String {
    let items = BESIDE
        .iter()
        .chain(EXTRA.iter())
        .map(|i| format!("{} {}", i.icon, i.meaning))
        .collect::<Vec<_>>()
        .join(" · ");
    format!("Beside the flow: {items}.")
}

/// The sample line of the board for one icon: `[✅] <step> · <done: landed or closed>`.
pub fn board_line(icon: &str) -> Option<String> {
    find(icon).map(|i| format!("[{}] <step> · <{}>", i.icon, i.hint))
}

/// The sample board, one line per icon in [`BOARD_ORDER`].
pub fn board_lines() -> Vec<String> {
    BOARD_ORDER.iter().filter_map(|i| board_line(i)).collect()
}

/// The icon of one key of the road or the beside list, for a test that names a state.
pub fn by_key(key: &str) -> Option<&'static str> {
    CARD_FLOW
        .iter()
        .chain(BESIDE.iter())
        .chain(EXTRA.iter())
        .find(|i| i.key == key)
        .map(|i| i.icon)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::statusline::width;
    use std::collections::HashSet;

    /// AC-1: eight different icons with their keys and meanings, none retired, every one on the
    /// width list.
    #[test]
    fn the_table_holds_eight_different_icons_none_retired_all_on_the_width_list() {
        let icons: Vec<&str> = status_line_icons().collect();
        assert_eq!(icons.len(), 8);
        let all: HashSet<&str> = CARD_FLOW
            .iter()
            .chain(BESIDE.iter())
            .chain(EXTRA.iter())
            .map(|i| i.icon)
            .collect();
        assert_eq!(all.len(), 10, "two entries are the same icon");
        let keys: HashSet<&str> = CARD_FLOW
            .iter()
            .chain(BESIDE.iter())
            .chain(EXTRA.iter())
            .map(|i| i.key)
            .collect();
        assert_eq!(keys.len(), 10, "two entries have the same key");
        for icon in &all {
            assert!(!RETIRED.contains(icon), "{icon} is retired");
        }
        for icon in &icons {
            assert!(
                width::ICONS.contains(icon),
                "{icon} is not on the width list"
            );
        }
        for entry in CARD_FLOW.iter().chain(BESIDE.iter()).chain(EXTRA.iter()) {
            assert!(
                !entry.meaning.is_empty() && !entry.hint.is_empty(),
                "{}",
                entry.key
            );
        }
        assert_eq!(by_key("paused"), Some(PAUSED));
        assert_eq!(by_key("nothing"), None);
    }

    /// The board lists every icon of the table once, in the order the board keeps.
    #[test]
    fn the_board_order_names_every_icon_once() {
        let lines = board_lines();
        assert_eq!(lines.len(), 10);
        let set: HashSet<&str> = BOARD_ORDER.iter().copied().collect();
        assert_eq!(set.len(), 10);
        assert_eq!(lines[0], "[🙋] <step> · <the pick it needs from you>");
        assert_eq!(lines[2], "[✅] <step> · <done: landed or closed>");
    }

    /// The two sentences the style texts spell.
    #[test]
    fn the_sentences_follow_the_table() {
        assert_eq!(
            flow_sentence(),
            "📋 to do, waiting its turn → 🔨 work in progress → 👀 in check (audit) → 🚀 check \
             passed, waiting to land → ✅ done (landed or closed)"
        );
        assert_eq!(
            beside_sentence(),
            "Beside the flow: ⛔ blocked · ⏸️ paused (withdrawn or folded) · 🙋 needs the \
             user's pick · 🔴 risk · ❌ failed or cancelled."
        );
    }
}
