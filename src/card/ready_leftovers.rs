//! SPEC-20261124-lane-leftovers: a hand-back note (`LANE-READY-INTEGRATION`, `REPAIRED`, `READY `)
//! carries one `leftovers:` line that names what the lane left behind, or `none`; `card note`
//! refuses the note without it. When it keeps the note it also reads the primary checkout for
//! files that appeared there since the card was reserved and puts them in its answer (observe:
//! the answer lists them, nothing is refused because of them).

use std::path::Path;

/// The start of the line.
pub const PREFIX: &str = "leftovers:";

/// What a readable `leftovers:` line says.
#[derive(Debug, PartialEq, Eq)]
pub enum Leftovers {
    /// `leftovers: none`.
    None,
    /// `leftovers: <path> | <path>`.
    Paths(Vec<String>),
}

/// Why the body has no readable `leftovers:` line.
#[derive(Debug, PartialEq, Eq)]
pub enum Why {
    Missing,
    Repeated,
    Empty,
    Placeholder,
    NoneWithPath,
}

/// What the read of the primary checkout found.
#[derive(Debug, PartialEq, Eq)]
pub enum Shared {
    /// The paths (relative to the primary checkout) that are new since the reservation began.
    Checked(Vec<String>),
    /// The check could not be made, and the reason.
    NotChecked(String),
}

/// The `leftovers:` line of a note `body` (its first line is the marker line and is never read).
pub fn read_line(_body: &str) -> Result<Leftovers, Why> {
    Err(Why::Missing)
}

/// The refusal of a hand-back note whose `leftovers:` line is missing or unreadable; `Ok` for a
/// note that is no hand-back or reads correctly.
pub fn check_note(_card: &str, _text: &str) -> Result<(), String> {
    Ok(())
}

/// The files of the primary checkout of `project` that are untracked or ignored, are not expected
/// state and were created or changed after `since` (seconds since the epoch).
pub fn shared_checkout_new(_project: &Path, _since: i64) -> Shared {
    Shared::Checked(Vec::new())
}

#[cfg(test)]
mod tests {
    use super::*;

    const FIRST: &str = "LANE-READY-INTEGRATION C-1 head=0000000000000000000000000000000000000000";

    fn body(rest: &str) -> String {
        format!("{FIRST}\n{rest}")
    }

    #[test]
    fn the_line_is_read_in_its_forms() {
        assert_eq!(
            read_line(&body("leftovers: none\n")),
            Ok(Leftovers::None)
        );
        assert_eq!(
            read_line(&body("- leftovers: target/x/copy | target/y\n")),
            Ok(Leftovers::Paths(vec![
                "target/x/copy".to_string(),
                "target/y".to_string()
            ]))
        );
        // A line among other lines, a path with a space, and blanks around the fields.
        assert_eq!(
            read_line(&body("base=x\nleftovers:  a b  |  c \nmore\n")),
            Ok(Leftovers::Paths(vec!["a b".to_string(), "c".to_string()]))
        );
    }

    #[test]
    fn a_line_that_is_missing_or_unreadable_gives_its_reason() {
        let why = |rest: &str| read_line(&body(rest)).unwrap_err();
        assert_eq!(why("base=x\n"), Why::Missing);
        assert_eq!(why("leftovers:\n"), Why::Empty);
        assert_eq!(why("leftovers:  | \n"), Why::Empty);
        assert_eq!(why("leftovers: <path>\n"), Why::Placeholder);
        assert_eq!(why("leftovers: a | <path>\n"), Why::Placeholder);
        assert_eq!(why("leftovers: none | build/x\n"), Why::NoneWithPath);
        assert_eq!(why("leftovers: a\nleftovers: none\n"), Why::Repeated);
        assert_eq!(read_line(FIRST).unwrap_err(), Why::Missing);
    }

    #[test]
    fn a_quoted_line_is_not_a_leftovers_line() {
        let why = |rest: &str| read_line(&body(rest)).unwrap_err();
        assert_eq!(why("```\nleftovers: none\n```\n"), Why::Missing);
        assert_eq!(why("~~~\nleftovers: none\n~~~\n"), Why::Missing);
        assert_eq!(why("    leftovers: none\n"), Why::Missing);
        assert_eq!(why("\tleftovers: none\n"), Why::Missing);
        // The marker line itself is never read, even when it holds the word.
        let first_only = format!("{FIRST} leftovers: none");
        assert_eq!(read_line(&first_only).unwrap_err(), Why::Missing);
        // A fence that closes lets the next line count again.
        assert_eq!(
            read_line(&body("```\nx\n```\nleftovers: none\n")),
            Ok(Leftovers::None)
        );
    }

    #[test]
    fn only_a_hand_back_note_is_judged() {
        assert_eq!(check_note("C-1", "LANE-BLOCKED kind=decision C-1\nbody"), Ok(()));
        assert_eq!(check_note("C-1", "plain note"), Ok(()));
        for marker in ["LANE-READY-INTEGRATION", "REPAIRED", "READY"] {
            let text = format!("{marker} C-1 head=0000000000000000000000000000000000000000\nbase=x");
            let refused = check_note("C-1", &text).unwrap_err();
            assert!(refused.contains("card note refused"), "{refused}");
            assert!(refused.contains("leftovers:"), "{refused}");
            let kept = format!("{text}\nleftovers: none");
            assert_eq!(check_note("C-1", &kept), Ok(()), "{marker}");
        }
    }

    #[test]
    fn the_refusal_names_the_reason_and_a_line_to_paste() {
        let text = format!("{FIRST}\nleftovers: <path>");
        let refused = check_note("C-1", &text).unwrap_err();
        assert!(refused.contains("placeholder"), "{refused}");
        assert!(refused.contains("leftovers: none"), "{refused}");
        assert!(refused.contains("target/"), "{refused}");
    }

    #[test]
    fn a_checkout_that_cannot_be_read_is_not_checked() {
        let gone = Path::new("/nonexistent/lane-leftovers-test-checkout");
        match shared_checkout_new(gone, 0) {
            Shared::NotChecked(why) => assert!(!why.is_empty()),
            other => panic!("expected NotChecked, got {other:?}"),
        }
    }
}
