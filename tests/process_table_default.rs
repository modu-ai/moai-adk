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
    cmd.args(["-c", "printf %s \"${MOAI_TEST_PROCESS_TABLE-unset}\""]);
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
const HOLDERS: [&str; 6] = [
    "Command::new(\"sleep\")",
    "Command::new(\"/bin/sleep\")",
    "Command::new(\"/bin/sh\")",
    "Command::new(\"cat\")",
    "Command::new(\"python3\")",
    "Holder(",
];

/// The top-level functions of a file as `(name, text)`, from the `fn` line to the closing `}`
/// line at the start of a line.
fn functions(text: &str) -> Vec<(String, String)> {
    let mut out = Vec::new();
    let mut open: Option<(String, String)> = None;
    for line in text.lines() {
        if let Some((_, body)) = open.as_mut() {
            body.push_str(line);
            body.push('\n');
            if line == "}" {
                out.extend(open.take());
            }
            continue;
        }
        let head = line.strip_prefix("pub ").unwrap_or(line);
        if let Some(rest) = head.strip_prefix("fn ") {
            let name = rest.split(['(', '<']).next().unwrap_or("").to_string();
            let item = (name, format!("{line}\n"));
            if line.trim_end().ends_with('}') {
                out.push(item);
            } else {
                open = Some(item);
            }
        }
    }
    out
}

/// The functions of `text` that start a holder program and neither read the real table nor
/// say why they need none (`process table not read:`).
fn unopted_holders(text: &str) -> Vec<String> {
    functions(text)
        .into_iter()
        .filter(|(_, body)| {
            HOLDERS.iter().any(|h| body.contains(h))
                && !body.contains("use_real_process_table()")
                && !body.contains("process table not read:")
        })
        .map(|(name, _)| name)
        .collect()
}

/// AC-5: the scan judges each function: a holder needs the opt-in or a stated reason in the
/// same function, and a type named `Holder` outside a function is no holder.
#[test]
fn holder_scan_reads_each_function() {
    let text = "struct Holder(u32);\n\n\
        fn opted() {\n    common::use_real_process_table();\n    Command::new(\"sleep\");\n}\n\n\
        fn reasoned() {\n    // process table not read: a lock decides\n    Command::new(\"python3\");\n}\n\n\
        fn opt_in_only() {\n    common::use_real_process_table();\n}\n\n\
        fn bare() {\n    Command::new(\"python3\");\n}\n";
    assert_eq!(unopted_holders(text), vec!["bare".to_string()]);
}

/// AC-3, AC-5: every named file sets the switch in its command helper, and every function of it
/// that starts a holder program opts back into the real table or says why it needs none.
#[test]
fn the_named_files_choose_their_table() {
    let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR"));
    for rel in NAMED {
        let text = std::fs::read_to_string(root.join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"));
        let helper = text.contains("common::process_table()") || text.contains(".clean_env()");
        assert!(helper, "{rel} sets no process table switch");
        let bare = unopted_holders(&text);
        assert!(
            bare.is_empty(),
            "{rel}: these functions start a holder program but read no real process table: {bare:?}"
        );
    }
}

/// AC-6: stopping a group-started shell holder leaves no process of its group behind.
#[cfg(unix)]
#[test]
fn a_group_started_shell_holder_leaves_no_child() {
    use std::time::{Duration, Instant};
    let mut shell = common::spawn_group(
        Command::new("/bin/sh")
            .args(["-c", "/bin/sleep 37; :", "sh"])
            .stdin(std::process::Stdio::null()),
    );
    std::thread::sleep(Duration::from_millis(300));
    let group = shell.id().to_string();
    common::kill_group(&mut shell);
    let start = Instant::now();
    let mut left = String::new();
    while start.elapsed() < Duration::from_secs(3) {
        let out = Command::new("pgrep").args(["-g", &group]).output().unwrap();
        left = String::from_utf8_lossy(&out.stdout).trim().to_string();
        if left.is_empty() {
            break;
        }
        std::thread::sleep(Duration::from_millis(100));
    }
    assert!(left.is_empty(), "group {group} still holds {left}");
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
