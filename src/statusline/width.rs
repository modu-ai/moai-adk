//! Display width, the width budget of a line, the form picker and the spacing after an icon
//! (SPEC-20261122-statusline-flow REQ-9 and REQ-10).
//!
//! Widths are counted the way the `unicode-width` crate of `Cargo.toml` counts them: an emoji
//! written with a variation selector counts 2 columns, which is never less than what a terminal
//! draws, so a line is never wider than counted.

use unicode_width::UnicodeWidthStr;

/// The margin the host keeps on a line (notifications and the cursor column): the budget of a
/// line is `COLUMNS` minus this minus the configured padding.
pub const LINE_MARGIN: usize = 2;

/// The display width of `text`.
pub fn display_width(text: &str) -> usize {
    UnicodeWidthStr::width(text)
}

/// The width a line may take: `COLUMNS` minus [`LINE_MARGIN`] minus `padding`, never below zero.
/// `None` when `COLUMNS` is missing, empty, zero or not a number: the widest form is written then.
pub fn budget(columns: Option<&str>, padding: usize) -> Option<usize> {
    let columns: usize = columns?.trim().parse().ok().filter(|c| *c > 0)?;
    Some(columns.saturating_sub(LINE_MARGIN.saturating_add(padding)))
}

/// The index of the form to write. `widths` lists the forms from the widest to the narrowest:
/// the first one that fits the `budget` wins, the last one is written when none fits, and the
/// first one when there is no budget. `widths` is never empty.
pub fn pick(widths: &[usize], budget: Option<usize>) -> usize {
    let Some(budget) = budget else {
        return 0;
    };
    widths
        .iter()
        .position(|w| *w <= budget)
        .unwrap_or(widths.len().saturating_sub(1))
}

/// `<icon> <text>`: exactly one space after every icon, also after one that is narrow by default
/// and written with the variation selector U+FE0F (`🗂️`, `🅱️`, `⬇️`), which used to take two
/// (SPEC-20261122-statusline-limits-left REQ-5). The width of the result is counted from the
/// text itself ([`display_width`]), so the form picker measures what is written.
pub fn icon_text(icon: &str, text: &str) -> String {
    format!("{icon} {text}")
}

/// Every icon the three line layouts can print. The code builds its text from [`icon_text`], and
/// the tests scan every rendered form for the one space after each of them.
pub const ICONS: [&str; 22] = [
    // Line 1: who and where.
    "🧠", "🤔", "📡", "⬇️", "🔀", "🌳", "🅱️", "🧭", "🔗",
    // The flow line: the one icon set of the progress board and the lane reports.
    "🎯", "🗂️", "🟡", "📨", "🛬", "✅", "🚧", "🙋",
    // Line 3: the limits and the versions.
    "🔋", "🪫", "💰", "🤖", "🗿",
];

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_budget_subtracts_the_margin_and_the_padding() {
        assert_eq!(budget(Some("120"), 0), Some(118));
        assert_eq!(budget(Some("80"), 0), Some(78));
        assert_eq!(budget(Some("80"), 3), Some(75));
        assert_eq!(budget(Some(" 60 "), 0), Some(58));
        // Smaller than the margin: zero, never an underflow.
        assert_eq!(budget(Some("2"), 0), Some(0));
        assert_eq!(budget(Some("1"), 5), Some(0));
        // A padding as large as a number can be never overflows: the budget is zero, so the
        // narrowest form is written (audit N1 of the flow line step).
        assert_eq!(budget(Some("80"), usize::MAX), Some(0));
        assert_eq!(budget(Some("80"), usize::MAX - 1), Some(0));
        assert_eq!(budget(Some("80"), usize::MAX - LINE_MARGIN), Some(0));
    }

    #[test]
    fn a_missing_empty_zero_or_non_numeric_columns_gives_no_budget() {
        for columns in [
            None,
            Some(""),
            Some("0"),
            Some("abc"),
            Some("-5"),
            Some("8 0"),
        ] {
            assert_eq!(budget(columns, 0), None, "{columns:?}");
        }
    }

    #[test]
    fn the_picker_changes_exactly_at_each_boundary() {
        let widths = [67, 38, 31, 16];
        assert_eq!(
            pick(&widths, Some(67)),
            0,
            "equal to the width keeps the form"
        );
        assert_eq!(pick(&widths, Some(66)), 1, "one less drops to the next");
        assert_eq!(pick(&widths, Some(38)), 1);
        assert_eq!(pick(&widths, Some(37)), 2);
        assert_eq!(pick(&widths, Some(31)), 2);
        assert_eq!(pick(&widths, Some(30)), 3);
        assert_eq!(pick(&widths, Some(16)), 3);
        assert_eq!(
            pick(&widths, Some(15)),
            3,
            "the narrowest is written when none fits"
        );
        assert_eq!(pick(&widths, Some(0)), 3);
        assert_eq!(pick(&widths, None), 0, "no budget writes the widest");
        assert_eq!(pick(&[5], Some(1)), 0, "one form is always the form");
    }

    /// Every icon of the table counts two columns as a whole, with or without a variation
    /// selector, so the width of a form is the same whatever icon it holds; and no icon is
    /// listed twice.
    #[test]
    fn every_icon_of_the_table_counts_two_columns() {
        for icon in ICONS {
            assert_eq!(display_width(icon), 2, "{icon}");
        }
        let mut unique = ICONS.to_vec();
        unique.sort_unstable();
        unique.dedup();
        assert_eq!(unique.len(), ICONS.len(), "an icon is listed twice");
    }

    /// SPEC-20261123-statusline-order-bars REQ-9: the tag icon of the integration item is in the
    /// table the one-space scans walk, and the distance icon, which no layout prints any more,
    /// is not (a table entry nothing prints would be a scan that can never see its icon).
    #[test]
    fn the_table_holds_the_tag_icon_and_not_the_distance_icon() {
        assert!(ICONS.contains(&"\u{1F3F7}\u{FE0F}"));
        assert_eq!(
            icon_text("\u{1F3F7}\u{FE0F}", "b3ba2b3"),
            "\u{1F3F7}\u{FE0F} b3ba2b3"
        );
        assert!(!ICONS.contains(&"\u{1F9ED}"));
    }

    /// SPEC-20261122-statusline-limits-left REQ-5: exactly one space follows every icon, the
    /// ones written with a variation selector (`🗂️`, `🅱️`, `⬇️`) too.
    #[test]
    fn one_space_follows_every_icon() {
        for icon in ICONS {
            assert_eq!(icon_text(icon, "x"), format!("{icon} x"), "{icon}");
        }
        assert_eq!(icon_text("🗂️", "150"), "🗂️ 150");
        assert_eq!(icon_text("🅱️", "main"), "🅱️ main");
        assert_eq!(icon_text("⬇️", "7"), "⬇️ 7");
        assert_eq!(icon_text("✅", "19"), "✅ 19");
    }

    #[test]
    fn widths_are_counted_by_the_crate() {
        assert_eq!(display_width("🚧 3 🙋 7"), 9);
        assert_eq!(display_width("🗂️ 150 → 🟡 14"), 14);
        assert_eq!(display_width("🅱️ main"), 7);
        assert_eq!(display_width(""), 0);
    }
}
