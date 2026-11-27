//! The `update-plan` help tells a reader what the command is next to `update` and
//! `project-update-plan`, and the command keeps its name and its arguments (the name is on hold,
//! so only the words around it change). The test runs the real binary with a temporary HOME.
#![cfg(unix)]
mod common;

use common::home::TempHome;

/// The help of `moai <args>` with every run of white space folded to one space.
fn help(home: &TempHome, args: &[&str]) -> String {
    let out = home.command(args).output().expect("run moai");
    assert!(out.status.success(), "{args:?} failed: {out:?}");
    String::from_utf8_lossy(&out.stdout)
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ")
}

/// The first line says what `update-plan` previews in terms of the command a person runs, and the
/// body says which command to run instead and how it differs from `project-update-plan`.
#[test]
fn the_update_plan_help_says_what_to_run_instead() {
    let home = TempHome::new();
    let text = help(&home, &["update-plan", "--help"]);
    for sentence in [
        "Preview what `update` writes to host copies",
        "the low-level step under `update`",
        "to update MoAI run `update`",
        "to preview a project's own files run `project-update-plan`",
    ] {
        assert!(
            text.contains(sentence),
            "`update-plan --help` lacks {sentence:?}: {text}"
        );
    }
}

/// The name and the two arguments are what they were: the help still shows the usage line and the
/// hold note, and the command list still carries the name.
#[test]
fn the_update_plan_name_and_arguments_are_unchanged() {
    let home = TempHome::new();
    let text = help(&home, &["update-plan", "--help"]);
    assert!(
        text.contains("update-plan <HOME> <STAGE>"),
        "the usage line changed: {text}"
    );
    assert!(
        text.contains("the command name is P06 HOLD"),
        "the hold note went missing: {text}"
    );
    let all = help(&home, &["help", "--all"]);
    assert!(
        all.contains("update-plan Preview what `update` writes to host copies"),
        "the command list lacks the new line: {all}"
    );
}
