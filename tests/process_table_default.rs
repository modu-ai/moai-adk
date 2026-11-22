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
