//! The two names an independent check is recorded under (SPEC-20261127-review-term-rename).
//!
//! The act was called audit and the product now calls it review. A record written before the
//! rename keeps its earlier spelling and is never rewritten, so a reader that looks for the
//! spelling of a record reads both: the verdict word that starts a note or comment, the tier
//! marker of a pull request body, and the markers of a reservation and its withdrawals. This is
//! the one place the names are listed; a reader asks it instead of spelling a name itself, so a
//! name added or dropped later changes here only. A line that starts with neither name is no
//! record, and a reader refuses it closed exactly as it did before the rename.

/// The word a verdict note or comment starts with, the earlier name first.
pub const VERDICT_WORDS: [&str; 2] = ["AUDIT", "REVIEW"];

/// The marker of the tier a pull request body declares (`<marker>: <tier> <reason>`).
pub const TIER_MARKERS: [&str; 2] = ["audit-tier", "review-tier"];

/// The marker of a reservation comment (`<marker>: lane=<lane> head=<40hex>`) and of its
/// withdrawal (`<marker> 철회: <lane> (<claim id>)`).
pub const CLAIM_MARKERS: [&str; 2] = ["moai-audit-claim", "moai-review-claim"];

/// The marker of a whole-lane withdrawal (`<marker>: lane=<lane>`).
pub const WITHDRAW_MARKERS: [&str; 2] = ["moai-audit-withdraw", "moai-review-withdraw"];

/// The other marker of a whole-lane withdrawal (`<marker>: lane=<lane>`).
pub const CLAIM_WITHDRAW_MARKERS: [&str; 2] =
    ["moai-audit-claim-withdraw", "moai-review-claim-withdraw"];

/// What follows `<name><tail>` when `line` starts with one of `names` and then `tail`.
pub fn strip_any<'a>(line: &'a str, names: &[&str], tail: &str) -> Option<&'a str> {
    names
        .iter()
        .find_map(|name| line.strip_prefix(name)?.strip_prefix(tail))
}

/// Whether `line` starts with one of `names` and then `tail`.
pub fn starts_with_any(line: &str, names: &[&str], tail: &str) -> bool {
    strip_any(line, names, tail).is_some()
}

/// Whether `line` starts with a verdict word and then `tail` (for a pull request comment the
/// tail is ` #`, for a card note a space).
pub fn starts_verdict(line: &str, tail: &str) -> bool {
    starts_with_any(line, &VERDICT_WORDS, tail)
}

/// Whether `text` starts with a verdict word at all, however the line goes on: a comment that
/// does but is no verdict line is a near miss the gate names instead of passing over.
pub fn starts_with_verdict_word(text: &str) -> bool {
    VERDICT_WORDS.iter().any(|word| text.starts_with(word))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_name_is_read_only_whole_and_only_at_the_start() {
        for word in VERDICT_WORDS {
            assert!(starts_verdict(&format!("{word} #9 (x)"), " #"));
            assert!(!starts_verdict(&format!(" {word} #9"), " #"));
            assert!(!starts_verdict(&format!("{word}X #9"), " #"));
            assert!(!starts_verdict(&format!("{word}#9"), " #"));
        }
        assert!(!starts_verdict("INSPECT #9", " #"));
        assert_eq!(
            strip_any("review-tier: 2 why", &TIER_MARKERS, ": "),
            Some("2 why")
        );
        assert_eq!(strip_any("audit-tier:0", &TIER_MARKERS, ": "), None);
    }
}
