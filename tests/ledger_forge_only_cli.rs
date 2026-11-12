//! #2384 (user decision 2026-11-10): the work ledger is forge Issues only. The local TODO
//! promotion commands are removed, so they are usage errors and absent from the help.

use std::process::Command;

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

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
        assert_eq!(stdout, "", "{name}");
        assert!(
            stderr.contains("unrecognized subcommand"),
            "{name}: {stderr}"
        );
    }
}

#[test]
fn help_lists_no_todo_promotion_command() {
    let (code, stdout, stderr) = run(&["--help"]);
    assert_eq!(code, 0, "{stderr}");
    // The help is the real command list (a known forge command is in it).
    assert!(stdout.contains("forge-read"), "{stdout}");
    assert!(!stdout.contains("todo-promote"), "{stdout}");
}
