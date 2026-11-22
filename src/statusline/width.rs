//! Display width, the width budget of a line, the form picker and the spacing after an icon
//! (SPEC-20261122-statusline-flow REQ-9 and REQ-10).
//!
//! Widths are counted the way the `unicode-width` crate of `Cargo.toml` counts them: an emoji
//! written with a variation selector counts 2 columns, which is never less than what a terminal
//! draws, so a line is never wider than counted.

use unicode_width::{UnicodeWidthChar, UnicodeWidthStr};

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

/// An icon that is narrow by default (its first character counts 1 column) and is written with
/// the variation selector U+FE0F: a terminal may draw it in one column or in two, so two spaces
/// follow it and the text never touches it. Every other icon is drawn in two columns and one
/// space follows it.
pub fn icon_gap(icon: &str) -> &'static str {
    let narrow_first = icon
        .chars()
        .next()
        .and_then(UnicodeWidthChar::width)
        .is_some_and(|w| w == 1);
    if narrow_first && icon.contains('\u{FE0F}') {
        "  "
    } else {
        " "
    }
}

/// `<icon><gap><text>`: the icon with the spacing its width class asks for.
pub fn icon_text(icon: &str, text: &str) -> String {
    format!("{icon}{}{text}", icon_gap(icon))
}

/// Every icon of the three line layouts with the gap pinned for it (the table of D-6). The code
/// builds its text from [`icon_text`], and a test compares this table to the crate's width.
pub const ICONS: [(&str, &str); 17] = [
    ("🧠", " "),
    ("🤔", " "),
    ("📡", " "),
    ("🌳", " "),
    ("🧭", " "),
    ("🔗", " "),
    ("🎯", " "),
    ("👀", " "),
    ("⛔", " "),
    ("✅", " "),
    ("🤖", " "),
    ("🗿", " "),
    ("⏸️", "  "),
    ("🗂️", "  "),
    ("🅱️", "  "),
    ("⬇️", "  "),
    ("🔀", " "),
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

    #[test]
    fn the_icon_table_matches_the_width_the_crate_counts() {
        for (icon, gap) in ICONS {
            let first = icon.chars().next().unwrap();
            let narrow_first = UnicodeWidthChar::width(first) == Some(1);
            let selector = icon.contains('\u{FE0F}');
            assert_eq!(
                gap == "  ",
                narrow_first && selector,
                "{icon}: the gap of the table"
            );
            assert_eq!(icon_gap(icon), gap, "{icon}");
            // The whole sequence never counts less than two columns.
            assert!(
                display_width(icon) >= 2,
                "{icon} counts {}",
                display_width(icon)
            );
        }
    }

    #[test]
    fn two_spaces_follow_the_narrow_icons_and_one_the_wide_ones() {
        assert_eq!(icon_text("⏸️", "H7"), "⏸️  H7");
        assert_eq!(icon_text("🗂️", "D43"), "🗂️  D43");
        assert_eq!(icon_text("🎯", "c12 W"), "🎯 c12 W");
        assert_eq!(icon_text("👀", "S1"), "👀 S1");
        assert_eq!(icon_text("⛔", "B3"), "⛔ B3");
        assert_eq!(icon_text("✅", "19"), "✅ 19");
    }

    #[test]
    fn widths_are_counted_by_the_crate() {
        assert_eq!(display_width("🎯 c12 W | ⛔ B3"), 16);
        assert_eq!(
            display_width("🎯 c12 W | 👀 S11 A2 L0 | ⛔ B3 ⏸️  H7 | 🗂️  D43 R150 W14 E22 ✅ 19"),
            67
        );
        assert_eq!(display_width(""), 0);
    }
}
