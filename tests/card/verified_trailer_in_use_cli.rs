//! The `Moai-Verified` trailer in use (SPEC-20261120-verified-trailer-in-use): the guidance files
//! that tell an implementer to write the trailer name it and its five values in one pinned
//! sentence each. The sentence is built from the trailer module's own constants, so a kind added
//! to or dropped from the module fails the pin until the guidance says the same.
//! Nothing here runs a command or touches the real HOME.

use moai_cli::trailers::{VERIFIED_KEY, VERIFIED_KINDS};
use std::fs;
use std::path::Path;

/// The five kinds as the guidance writes them: the four kinds of check in a list, then the
/// kind that says no check ran.
fn kinds_in_words() -> String {
    let (none, checks) = VERIFIED_KINDS.split_last().expect("five kinds");
    let checks: Vec<String> = checks.iter().map(|k| format!("`{k}`")).collect();
    format!(
        "only the kinds of check you ran ({}), and `{none}` when no check ran.",
        checks.join(", ")
    )
}

fn read(path: &str) -> String {
    let full = Path::new(env!("CARGO_MANIFEST_DIR")).join(path);
    fs::read_to_string(&full).unwrap_or_else(|e| panic!("read {}: {e}", full.display()))
}

/// AC-2 (REQ-3): the implementation agent's steps tell it to end a commit message with the
/// trailers `moai trailers` prints and to name only the kinds of check it ran.
#[test]
fn the_implementation_agent_names_the_trailer_and_its_five_values() {
    let sentence = format!(
        "When the task says to commit, end each commit message with the trailers that \
         `moai trailers --spec <SPEC id> --verified <kinds>` prints, and name in `{VERIFIED_KEY}` {}",
        kinds_in_words()
    );
    let text = read("agents/impl.md");
    assert!(
        text.lines().any(|line| line.contains(&sentence)),
        "agents/impl.md has no line holding the pinned sentence:\n{sentence}"
    );
}
