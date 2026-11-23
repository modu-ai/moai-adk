//! The `guard-mutant:` lines of a READY note (SPEC-20261123-ready-mutant-line): one line per new
//! guard test that says which code a mutant removed or flipped and that the test failed on it, or
//! `none` with a reason when the change adds no guard test. `bundle check` reads them and lists
//! warnings; a warning never changes `ok`, `findings` or the exit status (observe first).

use serde_json::{json, Value};

/// The start of a mutant line.
pub const PREFIX: &str = "guard-mutant:";

/// One warning: its stable code and its detail.
pub type Warning = (&'static str, Value);

/// The warnings for one head. `body` is the whole text of the newest READY note that names the
/// head (its first line is the marker line and is never read as a mutant line); `None` means no
/// READY note names the head.
pub fn warnings(body: Option<&str>) -> Vec<Warning> {
    let Some(body) = body else {
        return vec![("mutant_line_missing", json!("no READY note names the head"))];
    };
    let mut out: Vec<Warning> = Vec::new();
    let mut seen = false;
    // The first line is the marker line; line numbers count it as line 1.
    for (i, raw) in body.lines().enumerate().skip(1) {
        let line = raw.trim_start();
        let line = line.strip_prefix("- ").unwrap_or(line);
        let Some(rest) = line.strip_prefix(PREFIX) else {
            continue;
        };
        seen = true;
        let number = i + 1;
        match read(rest) {
            Some(Read::Failed) | Some(Read::None) => {}
            Some(Read::Survived(test)) => {
                out.push(("mutant_survived", json!({"test": test, "line": number})))
            }
            None => out.push(("mutant_line_unreadable", json!({"line": number}))),
        }
    }
    if !seen {
        out.push((
            "mutant_line_missing",
            json!("the note body holds no `guard-mutant:` line"),
        ));
    }
    out
}

enum Read {
    /// A mutant the test failed on.
    Failed,
    /// `none | <reason>`: the change adds no guard test.
    None,
    /// A mutant the test passed on (`exit 0`), with the test's name.
    Survived(String),
}

/// What follows `guard-mutant:` on a line: `<test> | <code removed or flipped> | exit <n>` or
/// `none | <reason>`. `None` when it fits neither form.
fn read(rest: &str) -> Option<Read> {
    let fields: Vec<&str> = rest.split('|').map(str::trim).collect();
    if fields.iter().any(|f| f.is_empty()) {
        return None;
    }
    match fields.as_slice() {
        ["none", _reason] => Some(Read::None),
        [test, _removed, exit] => {
            let code = exit.strip_prefix("exit ")?;
            if code.is_empty() || !code.bytes().all(|b| b.is_ascii_digit()) {
                return None;
            }
            match code.parse::<u32>().ok()? {
                0 => Some(Read::Survived((*test).to_string())),
                _ => Some(Read::Failed),
            }
        }
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const FIRST: &str = "LANE-READY-INTEGRATION C-1 head=0000000000000000000000000000000000000000";

    fn note(rest: &str) -> String {
        format!("{FIRST}\n{rest}")
    }

    fn codes(body: &str) -> Vec<&'static str> {
        warnings(Some(body)).iter().map(|w| w.0).collect()
    }

    #[test]
    fn a_failing_mutant_line_and_a_none_line_give_no_warning() {
        let one = note("guard-mutant: cap_is_parsed | removed the cap parse | exit 101");
        assert_eq!(codes(&one), Vec::<&str>::new());
        let none = note("guard-mutant: none | the change adds no guard test");
        assert_eq!(codes(&none), Vec::<&str>::new());
        let both = note(
            "base: abc\nguard-mutant: a | removed x | exit 1\nguard-mutant: b | flipped y | exit 101",
        );
        assert_eq!(codes(&both), Vec::<&str>::new());
    }

    #[test]
    fn a_bullet_and_leading_spaces_are_allowed() {
        let n = note("  - guard-mutant: a | removed x | exit 1");
        assert_eq!(codes(&n), Vec::<&str>::new());
    }

    #[test]
    fn a_note_without_a_line_is_missing_and_a_note_no_one_wrote_is_missing_too() {
        assert_eq!(
            codes(&note("base: abc\nnothing else")),
            ["mutant_line_missing"]
        );
        assert_eq!(codes(FIRST), ["mutant_line_missing"]);
        let none_named = warnings(None);
        assert_eq!(none_named.len(), 1);
        assert_eq!(none_named[0].0, "mutant_line_missing");
        assert!(none_named[0].1.to_string().contains("no READY note names"));
    }

    #[test]
    fn the_marker_line_itself_is_never_read_as_a_mutant_line() {
        let first = format!("{FIRST} guard-mutant: a | removed x | exit 1");
        assert_eq!(codes(&first), ["mutant_line_missing"]);
        let only_first = "guard-mutant: a | removed x | exit 1";
        assert_eq!(codes(only_first), ["mutant_line_missing"]);
    }

    #[test]
    fn a_line_that_fits_neither_form_is_unreadable_and_names_its_number() {
        for bad in [
            "guard-mutant: a | removed x",
            "guard-mutant: a | removed x | exit",
            "guard-mutant: a | removed x | exit one",
            "guard-mutant: a | removed x | exit 1 extra",
            "guard-mutant: a | | exit 1",
            "guard-mutant:  | removed x | exit 1",
            "guard-mutant: none |",
            "guard-mutant: none",
            "guard-mutant: a | removed x | exit 1 | more",
            "guard-mutant: a | removed x | exit 99999999999",
            "guard-mutant:",
        ] {
            let w = warnings(Some(&note(&format!("base: x\n{bad}"))));
            assert_eq!(w.len(), 1, "{bad}: {w:?}");
            assert_eq!(w[0].0, "mutant_line_unreadable", "{bad}");
            // The marker line is line 1, `base: x` is line 2, the bad line is line 3.
            assert_eq!(w[0].1, json!({"line": 3}), "{bad}");
        }
    }

    #[test]
    fn a_mutant_that_survived_names_its_test() {
        let w = warnings(Some(&note(
            "guard-mutant: cap_is_parsed | removed the cap parse | exit 0",
        )));
        assert_eq!(w.len(), 1, "{w:?}");
        assert_eq!(w[0].0, "mutant_survived");
        assert_eq!(w[0].1, json!({"test": "cap_is_parsed", "line": 2}));
    }

    #[test]
    fn every_bad_line_is_listed_and_one_good_line_does_not_hide_a_bad_one() {
        let n = note(
            "guard-mutant: a | removed x | exit 1\nguard-mutant: b | flipped y | exit 0\nguard-mutant: oops",
        );
        assert_eq!(codes(&n), ["mutant_survived", "mutant_line_unreadable"]);
    }

    #[test]
    fn a_line_must_start_with_the_prefix() {
        // Mentioning the word in a sentence is not a mutant line.
        let n = note("see the guard-mutant: line above\nnot a guard-mutant: a | b | exit 1");
        assert_eq!(codes(&n), ["mutant_line_missing"]);
    }
}
