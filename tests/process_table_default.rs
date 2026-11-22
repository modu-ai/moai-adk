//! SPEC-20261122-test-process-table-default: a child command a test builds with `clean_env` reads
//! no machine-wide process table by default (the debug-build switch of
//! SPEC-20261121-cheap-process-scan); a test that judges whether a program uses a folder opts
//! back into the real table.
mod common;
use common::CleanEnv;
use std::process::Command;

/// What the child sees in the switch variable, `unset` when it has none.
fn seen(cmd: &mut Command) -> String {
    let out = cmd.output().unwrap();
    assert!(out.status.success(), "{out:?}");
    String::from_utf8(out.stdout).unwrap()
}

fn child() -> Command {
    let mut cmd = Command::new("sh");
    cmd.args(["-c", "printf %s \"${MOAI_CLI_TEST_PROCESS_TABLE-unset}\""]);
    cmd
}

/// AC-1: a `clean_env` child sees the switch set to `empty`.
#[test]
fn a_clean_env_child_skips_the_process_table() {
    assert_eq!(seen(child().clean_env()), "empty");
}

/// The slow test files whose own command helper sets the switch (named, no directory walk).
const NAMED: [&str; 10] = [
    "tests/home_install_apply_cli.rs",
    "tests/home_retire_cli.rs",
    "tests/home_state_apply.rs",
    "tests/home_state_plan.rs",
    "tests/home_install_go_retire.rs",
    "tests/home_install_plan_cli.rs",
    "tests/update_home_layer_cli.rs",
    "tests/hygiene_apply_cli.rs",
    "tests/hygiene_cli.rs",
    "tests/optimize_plan_cli.rs",
];

/// Ways a named file starts a program that holds a folder or a file for an in-use test.
const HOLDERS: [&str; 5] = [
    "Command::new(\"sleep\")",
    "Command::new(\"/bin/sleep\")",
    "Command::new(\"/bin/sh\")",
    "Command::new(\"cat\")",
    "Holder(",
];

/// AC-3: every named file sets the switch in its command helper, and every named file that
/// starts a holder program opts back into the real table.
#[test]
fn the_named_files_choose_their_table() {
    let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR"));
    for rel in NAMED {
        let text = std::fs::read_to_string(root.join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"));
        let helper = text.contains("common::process_table()") || text.contains(".clean_env()");
        assert!(helper, "{rel} sets no process table switch");
        if HOLDERS.iter().any(|h| text.contains(h)) {
            assert!(
                text.contains("use_real_process_table()"),
                "{rel} starts a holder program but never reads the real process table"
            );
        }
    }
}

/// AC-1: after `use_real_process_table` the test's children carry a value the product does not
/// read as the switch, so they run the real scan; another test keeps the default.
#[test]
fn a_test_that_judges_in_use_opts_back_in() {
    common::use_real_process_table();
    assert_eq!(seen(child().clean_env()), "real");
    assert_eq!(common::process_table(), common::REAL_PROCESS_TABLE);
    let other = std::thread::spawn(|| seen(child().clean_env()))
        .join()
        .unwrap();
    assert_eq!(
        other, "empty",
        "the choice stays with the test that made it"
    );
}
