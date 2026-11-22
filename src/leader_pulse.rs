//! `moai leader-pulse` (Issue #3015 C5, plan comment 6905006699, leader decision 2=B in #1975):
//! a leader that has no timer of its own types one round instruction into its own terminal every
//! N seconds. This module only builds what the skill needs for that: the checked instruction,
//! the interval and the argument list of the one `orca terminal send` call. It never runs
//! `orca`, never sleeps and never reads a terminal; the loop and the call belonged to the skill's
//! leader workflow, which left the package (the card loop is `skills/moai/workflows/work.md`).
//!
//! Rules the output carries (decision 3=A): a pulse whose turn did not start is reported, not
//! sent again; the command never creates a terminal or a session.
//!
//! Measured on a live Codex leader: not yet (AC-5 needs a signed-in Codex session), so the
//! receipt of `orca terminal send --json` is not interpreted here.

use serde_json::{json, Value};

/// Seconds between two pulses when `--interval-secs` is not given.
pub const DEFAULT_INTERVAL_SECS: u64 = 300;
/// Shortest and longest interval accepted, in seconds.
pub const MIN_INTERVAL_SECS: u64 = 60;
pub const MAX_INTERVAL_SECS: u64 = 3600;
/// Longest instruction, in characters.
pub const MAX_TEXT_CHARS: usize = 1000;
/// Longest terminal handle, in characters.
pub const MAX_TERMINAL_CHARS: usize = 200;
/// Longest run, in rounds.
pub const MAX_ROUNDS: u32 = 10_000;
/// How long `orca terminal send` watches for the prompt to start a turn before it answers.
pub const WAIT_SUBMIT_SECS: u64 = 20;

type Result<T> = std::result::Result<T, String>;

/// A runtime-issued handle: letters, digits and `. _ : -`, not starting with `-`.
pub fn check_terminal(handle: &str) -> Result<()> {
    let chars = handle.chars().count();
    if chars == 0 || chars > MAX_TERMINAL_CHARS {
        return Err(format!(
            "the terminal handle must be 1 to {MAX_TERMINAL_CHARS} characters, not {chars}"
        ));
    }
    if handle.starts_with('-') {
        return Err("the terminal handle must not start with `-`".into());
    }
    if !handle
        .chars()
        .all(|c| c.is_ascii_alphanumeric() || matches!(c, '.' | '_' | ':' | '-'))
    {
        return Err("the terminal handle may hold only letters, digits and `. _ : -`".into());
    }
    Ok(())
}

/// One line of text: not empty, at most [`MAX_TEXT_CHARS`], no control character (a newline or
/// an escape would act as a key), not starting with `-` (it could read as an option).
pub fn check_text(text: &str) -> Result<()> {
    if text.trim().is_empty() {
        return Err("the instruction text is empty".into());
    }
    let chars = text.chars().count();
    if chars > MAX_TEXT_CHARS {
        return Err(format!(
            "the instruction text is {chars} characters; the limit is {MAX_TEXT_CHARS}"
        ));
    }
    if text.chars().any(char::is_control) {
        return Err("the instruction text must be one line without control characters".into());
    }
    if text.starts_with('-') {
        return Err(
            "the instruction text must not start with `-` (it could read as an option)".into(),
        );
    }
    Ok(())
}

/// The arguments after `orca` for one pulse. Every value is its own argument.
pub fn send_args(terminal: &str, text: &str) -> Vec<String> {
    [
        "terminal",
        "send",
        "--terminal",
        terminal,
        "--text",
        text,
        "--enter",
        "--wait-submit",
        &WAIT_SUBMIT_SECS.to_string(),
        "--json",
    ]
    .iter()
    .map(|s| s.to_string())
    .collect()
}

/// A word safe to paste into a POSIX shell: plain words stay as they are, anything else goes in
/// single quotes with each `'` written `'\''`.
pub fn shell_word(word: &str) -> String {
    let plain = !word.is_empty()
        && word
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || matches!(c, '-' | '_' | '.' | ':' | '/' | '='));
    if plain {
        word.to_string()
    } else {
        format!("'{}'", word.replace('\'', "'\\''"))
    }
}

/// The checked pulse: the text, the interval, the rounds and the one `orca` command.
pub fn build(terminal: &str, text: &str, interval_secs: u64, rounds: Option<u32>) -> Result<Value> {
    check_terminal(terminal)?;
    check_text(text)?;
    if !(MIN_INTERVAL_SECS..=MAX_INTERVAL_SECS).contains(&interval_secs) {
        return Err(format!(
            "the interval must be {MIN_INTERVAL_SECS} to {MAX_INTERVAL_SECS} seconds, not {interval_secs}"
        ));
    }
    if let Some(n) = rounds {
        if !(1..=MAX_ROUNDS).contains(&n) {
            return Err(format!(
                "rounds must be 1 to {MAX_ROUNDS} (leave it out to run until stopped), not {n}"
            ));
        }
    }
    let args = send_args(terminal, text);
    let command_line = std::iter::once("orca".to_string())
        .chain(args.iter().map(|a| shell_word(a)))
        .collect::<Vec<_>>()
        .join(" ");
    Ok(json!({
        "schema_version": 1,
        "terminal": terminal,
        "text": text,
        "interval_secs": interval_secs,
        "rounds": rounds,
        "program": "orca",
        "args": args,
        "command_line": command_line,
        "rules": {
            "create_terminal": false,
            "resend_when_turn_not_started": false,
            "on_turn_not_started": "report",
            "between_pulses": "wait interval_secs, then run the command once",
        },
        "writes": false,
        "verified": "not_run",
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_interval_has_hard_edges() {
        for (secs, ok) in [
            (59, false),
            (60, true),
            (3600, true),
            (3601, false),
            (0, false),
        ] {
            assert_eq!(build("t1", "go", secs, None).is_ok(), ok, "{secs} seconds");
        }
        assert_eq!(DEFAULT_INTERVAL_SECS, 300);
        assert!((MIN_INTERVAL_SECS..=MAX_INTERVAL_SECS).contains(&DEFAULT_INTERVAL_SECS));
    }

    #[test]
    fn rounds_have_hard_edges() {
        for (n, ok) in [
            (0, false),
            (1, true),
            (MAX_ROUNDS, true),
            (MAX_ROUNDS + 1, false),
        ] {
            assert_eq!(build("t1", "go", 60, Some(n)).is_ok(), ok, "{n} rounds");
        }
        assert!(build("t1", "go", 60, None).is_ok());
    }

    #[test]
    fn a_terminal_handle_is_one_plain_word() {
        let longest = "a".repeat(MAX_TERMINAL_CHARS);
        let too_long = "a".repeat(MAX_TERMINAL_CHARS + 1);
        for good in ["t", "term_1", "a.b:c-d", longest.as_str()] {
            assert!(check_terminal(good).is_ok(), "{good:?}");
        }
        for bad in [
            "",
            "-x",
            "--help",
            "a b",
            "a;b",
            "a\nb",
            "a/b",
            "한글",
            too_long.as_str(),
        ] {
            assert!(check_terminal(bad).is_err(), "{bad:?}");
        }
    }

    #[test]
    fn the_text_is_one_line_and_cannot_be_an_option() {
        let longest = "x".repeat(MAX_TEXT_CHARS);
        let too_long = "x".repeat(MAX_TEXT_CHARS + 1);
        for good in ["go", "순환 계속", "a - b", longest.as_str()] {
            assert!(check_text(good).is_ok(), "{good:?}");
        }
        for bad in [
            "",
            "   ",
            "-go",
            "--help",
            "a\nb",
            "a\rb",
            "a\u{1b}[31mb",
            "a\tb",
            "a\u{0}b",
            too_long.as_str(),
        ] {
            assert!(check_text(bad).is_err(), "{bad:?}");
        }
    }

    #[test]
    fn the_arguments_are_separate_and_send_one_prompt_with_enter() {
        let args = send_args("term_abc", "순환 계속 — self-pick (1)부터");
        assert_eq!(
            args,
            [
                "terminal",
                "send",
                "--terminal",
                "term_abc",
                "--text",
                "순환 계속 — self-pick (1)부터",
                "--enter",
                "--wait-submit",
                "20",
                "--json"
            ]
        );
        // `create` and `--interrupt` are never part of a pulse.
        assert!(!args.iter().any(|a| a == "create" || a == "--interrupt"));
    }

    #[test]
    fn a_shell_word_is_quoted_unless_plain() {
        assert_eq!(shell_word("term_1"), "term_1");
        assert_eq!(shell_word("--enter"), "--enter");
        assert_eq!(shell_word("a b"), "'a b'");
        assert_eq!(shell_word("it's"), "'it'\\''s'");
        assert_eq!(shell_word("$(x)"), "'$(x)'");
        // A single word with a variable must not reach the shell unquoted.
        assert_eq!(shell_word("$HOME"), "'$HOME'");
        assert_eq!(shell_word("~"), "'~'");
        assert_eq!(shell_word("a*"), "'a*'");
        assert_eq!(shell_word(""), "''");
    }

    #[test]
    fn the_output_names_the_command_and_the_rules() {
        let v = build("term_abc", "go on", 120, Some(2)).unwrap();
        assert_eq!(v["interval_secs"], 120);
        assert_eq!(v["rounds"], 2);
        assert_eq!(v["program"], "orca");
        assert_eq!(
            v["command_line"],
            "orca terminal send --terminal term_abc --text 'go on' --enter --wait-submit 20 --json"
        );
        assert_eq!(v["rules"]["create_terminal"], false);
        assert_eq!(v["rules"]["resend_when_turn_not_started"], false);
        assert_eq!(v["rules"]["on_turn_not_started"], "report");
        assert_eq!(v["writes"], false);
        assert_eq!(build("t", "go", 60, None).unwrap()["rounds"], Value::Null);
    }
}
