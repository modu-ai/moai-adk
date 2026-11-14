//! #2384 (user decision 2026-11-10): the work ledger is forge Issues only. The local TODO
//! promotion commands are removed, so they are usage errors and absent from the help.

mod common;
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;

fn run(args: &[&str]) -> (i32, String, String) {
    let home = tempfile::tempdir().unwrap();
    let out = Command::new(BIN)
        .args(args)
        .current_dir(home.path())
        .env("HOME", home.path())
        .output()
        .unwrap();
    (
        out.status.code().unwrap(),
        String::from_utf8_lossy(&out.stdout).into_owned(),
        String::from_utf8_lossy(&out.stderr).into_owned(),
    )
}

#[test]
fn removed_todo_promotion_commands_are_usage_errors() {
    for name in ["todo-promote-preview", "todo-promote-apply"] {
        let (code, stdout, stderr) = run(&[name, "."]);
        assert_eq!(code, 2, "{name}: {stderr}");
        // #2799 D3 decision 1 (a): the `cli` usage envelope, never the removed operation's.
        assert_eq!(
            stdout,
            "{\"protocol_version\":1,\"operation\":\"cli\",\"ok\":false,\"error\":{\"kind\":\"usage\",\"message\":\"unknown command; see --help\",\"retryable\":false}}\n",
            "{name}"
        );
        assert!(
            stderr.contains("unrecognized subcommand"),
            "{name}: {stderr}"
        );
    }
}

#[test]
fn help_lists_no_todo_promotion_command() {
    // `help --all` is the full command list (#2802 S1); the root help lists only the commands
    // a person types.
    let (code, stdout, stderr) = run(&["help", "--all"]);
    assert_eq!(code, 0, "{stderr}");
    // The help is the real command list (a known forge command is in it).
    assert!(stdout.contains("forge-read"), "{stdout}");
    assert!(!stdout.contains("todo-promote"), "{stdout}");
}
