//! #2503 AC-2: a statusline golden re-check was refused once because a case name holds the
//! words of a guarded command (`b01-git-clean`; PR #2482, report 6810975788). The refusing
//! guard was not named in that record. These cases pin that the repository's own
//! tool-preflight guards (`hook_guards`, #2391) read commands, not words: the golden command
//! is neither denied nor asked about, while a real forced `git clean` still is. The golden
//! harness is the Rust test `tests/statusline_golden.rs` (SPEC-20261125-rust-only-finish): it
//! is re-recorded by name, and `MOAI_GOLDEN_CASES` / `MOAI_GOLDEN_VARIANTS` pick cases.

use moai_cli::hook_guards::{bash_confirm, bash_guard};

const GOLDEN: &str = "cargo test --test statusline_golden -- --ignored record_statusline_goldens";

fn assert_allowed(command: &str) {
    assert_eq!(bash_guard(command), None, "denied: {command}");
    assert_eq!(bash_confirm(command), None, "asked: {command}");
}

#[test]
fn a_golden_record_naming_the_git_clean_case_is_allowed() {
    assert_allowed(&format!("MOAI_GOLDEN_CASES=b01-git-clean {GOLDEN}"));
    assert_allowed(&format!(
        "MOAI_GOLDEN_CASES=b01-git-clean MOAI_GOLDEN_VARIANTS=ansi256 {GOLDEN}"
    ));
    assert_allowed(&format!(
        "MOAI_GOLDEN_CASES='b01-git-clean,b02-git-dirty-ahead' {GOLDEN}"
    ));
    // The whole set, with no case name on the line at all.
    assert_allowed(GOLDEN);
}

#[test]
fn a_golden_record_with_a_command_substitution_is_allowed() {
    assert_allowed(&format!(
        "B=$(git rev-parse --show-toplevel) && cd \"$B\" && MOAI_GOLDEN_CASES=b01-git-clean {GOLDEN}"
    ));
}

#[test]
fn a_forced_git_clean_is_still_asked_about() {
    // Control: the same words as a real command still reach the confirmation class.
    assert!(bash_confirm("git clean -fd").is_some());
    assert!(bash_confirm(&format!(
        "MOAI_GOLDEN_CASES=b01-git-clean {GOLDEN} && git clean -fd"
    ))
    .is_some());
}
