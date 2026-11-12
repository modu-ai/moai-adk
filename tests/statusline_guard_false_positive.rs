//! #2503 AC-2: a statusline golden re-check was refused once because a case name holds the
//! words of a guarded command (`b01-git-clean`; PR #2482, report 6810975788). The refusing
//! guard was not named in that record. These cases pin that the repository's own
//! tool-preflight guards (`hook_guards`, #2391) read commands, not words: the golden command
//! is neither denied nor asked about, while a real forced `git clean` still is.

use moai_cli::hook_guards::{bash_confirm, bash_guard};

const GOLDEN: &str =
    "python3 -I tools/statusline_golden.py --bin \"$HOME/.moai-cli/path/moai-cli\" \
                      --argv statusline --check";

fn assert_allowed(command: &str) {
    assert_eq!(bash_guard(command), None, "denied: {command}");
    assert_eq!(bash_confirm(command), None, "asked: {command}");
}

#[test]
fn a_golden_check_naming_the_git_clean_case_is_allowed() {
    assert_allowed(&format!("{GOLDEN} --case b01-git-clean"));
    assert_allowed(&format!("{GOLDEN} --case b01-git-clean --variant ansi256"));
    assert_allowed(&format!(
        "{GOLDEN} --case 'b01-git-clean' --case b02-git-dirty-ahead"
    ));
    // The whole set, with no case name on the line at all.
    assert_allowed(GOLDEN);
}

#[test]
fn a_golden_check_with_a_command_substitution_is_allowed() {
    assert_allowed(&format!(
        "B=$(git rev-parse --show-toplevel) && cd \"$B\" && {GOLDEN} --case b01-git-clean"
    ));
}

#[test]
fn a_forced_git_clean_is_still_asked_about() {
    // Control: the same words as a real command still reach the confirmation class.
    assert!(bash_confirm("git clean -fd").is_some());
    assert!(bash_confirm(&format!("{GOLDEN} --case b01-git-clean && git clean -fd")).is_some());
}
