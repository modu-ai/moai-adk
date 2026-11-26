//! SPEC-20261126-scratch-cargo-wrapper: a test that starts cargo goes through
//! `tests/common/cargo.rs`, which starts it without the compiler wrapper of the outer run.
//!
//! The scan reads one line at a time and sees two shapes: a process whose program names cargo,
//! and a read of the `CARGO` variable. A shell text or a `moai` command that starts cargo is not
//! seen here.

#[path = "common/cargo.rs"]
mod scratch_cargo;

use std::ffi::OsStr;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

/// The call the scan looks for, assembled so that this file does not report itself.
const START: &str = concat!("Command::", "new(");

/// Reads of the variable that names the running cargo, assembled for the same reason.
const READS: [&str; 3] = [
    concat!("var(\"CAR", "GO\")"),
    concat!("var_os(\"CAR", "GO\")"),
    concat!("env!(\"CAR", "GO\")"),
];

/// Names Cargo sets that say nothing about the program being started.
const NOT_CARGO: [&str; 3] = ["CARGO_BIN_EXE", "CARGO_MANIFEST_DIR", "CARGO_TARGET_TMPDIR"];

/// Test files that may start cargo themselves, each with its reason. A file counts at the top of
/// `tests/` or in one area folder below it.
const EXCEPTIONS: [(&str, &str); 1] = [(
    "check_affected_touch.rs",
    "starts the toolchain's cargo and removes RUSTC_WRAPPER itself; moving it to the helper \
     waits for the work that holds the file",
)];

/// The text of a call's argument: up to the parenthesis that closes it, else the rest of the line.
fn argument(rest: &str) -> &str {
    let mut depth = 0usize;
    for (at, c) in rest.char_indices() {
        match c {
            '(' => depth += 1,
            ')' if depth == 0 => return &rest[..at],
            ')' => depth -= 1,
            _ => {}
        }
    }
    rest
}

fn starts_cargo(line: &str) -> bool {
    line.match_indices(START).any(|(at, _)| {
        let mut program = argument(&line[at + START.len()..]).to_string();
        for name in NOT_CARGO {
            program = program.replace(name, "");
        }
        program.to_ascii_lowercase().contains("cargo")
    })
}

/// The numbers of the lines that start cargo without the helper.
fn direct_starts(text: &str) -> Vec<usize> {
    text.lines()
        .enumerate()
        .filter(|(_, line)| {
            let line = line.trim_start();
            !line.starts_with("//")
                && (starts_cargo(line) || READS.iter().any(|read| line.contains(read)))
        })
        .map(|(at, _)| at + 1)
        .collect()
}

fn rust_files(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for entry in fs::read_dir(&dir).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() {
                stack.push(path);
            } else if path.extension().is_some_and(|e| e == "rs") {
                out.push(path);
            }
        }
    }
    out.sort();
    out
}

fn exempt(relative: &Path) -> bool {
    if relative == Path::new("common/cargo.rs") {
        return true;
    }
    relative.components().count() <= 2
        && relative
            .file_name()
            .is_some_and(|name| EXCEPTIONS.iter().any(|(file, _)| name == OsStr::new(file)))
}

/// `tests/<file>:<line>` for every direct start under `tests`.
fn violations(tests: &Path) -> Vec<String> {
    let mut out = Vec::new();
    for file in rust_files(tests) {
        let relative = file.strip_prefix(tests).unwrap();
        if exempt(relative) {
            continue;
        }
        for line in direct_starts(&fs::read_to_string(&file).unwrap()) {
            out.push(format!("tests/{}:{line}", relative.display()));
        }
    }
    out
}

#[test]
fn scratch_cargo_starts_go_through_the_shared_helper() {
    let found = violations(&Path::new(env!("CARGO_MANIFEST_DIR")).join("tests"));
    assert!(
        found.is_empty(),
        "test files start cargo without the shared helper, so the outer compiler wrapper reaches \
         their scratch crates; use tests/common/cargo.rs:\n{}",
        found.join("\n")
    );
}

#[test]
fn the_scan_names_the_file_and_line_of_a_direct_start() {
    let dir = tempfile::TempDir::new().unwrap();
    let by_name = format!("fn f() {{\n    {START}\"cargo\");\n}}\n");
    let by_path = format!("fn f() {{ std::process::{START}bin.join(\"cargo\")); }}\n");
    let by_call = format!("\n\nfn f() {{ {START}cargo_path()); }}\n");
    let by_variable = format!("fn f() {{ let c = std::env::{}; }}\n", READS[0]);
    let clean = format!(
        "// {START}\"cargo\") in a comment\n\
         fn f() {{ scratch_cargo::cargo_at(\"cargo\"); }}\n\
         fn g() {{ {START}env!(\"CARGO_BIN_EXE_moai\")).arg(\"cargo\"); }}\n\
         fn h() {{ {START}\"git\"); }}\n"
    );
    for (relative, text) in [
        ("a_name.rs", &by_name),
        ("b_path.rs", &by_path),
        ("area/c_call.rs", &by_call),
        ("d_variable.rs", &by_variable),
        ("e_clean.rs", &clean),
        ("common/cargo.rs", &by_name),
        ("check_affected_touch.rs", &by_path),
        ("spec/check_affected_touch.rs", &by_path),
        ("spec/deep/check_affected_touch.rs", &by_path),
        ("other/cargo.rs", &by_name),
    ] {
        let file = dir.path().join(relative);
        fs::create_dir_all(file.parent().unwrap()).unwrap();
        fs::write(file, text).unwrap();
    }
    assert_eq!(
        violations(dir.path()),
        [
            "tests/a_name.rs:2",
            "tests/area/c_call.rs:3",
            "tests/b_path.rs:1",
            "tests/d_variable.rs:1",
            "tests/other/cargo.rs:2",
            "tests/spec/deep/check_affected_touch.rs:1",
        ]
    );
}

#[test]
fn the_helper_removes_every_wrapper_variable() {
    assert_eq!(
        scratch_cargo::WRAPPER_VARS,
        [
            "RUSTC_WRAPPER",
            "CARGO_BUILD_RUSTC_WRAPPER",
            "RUSTC_WORKSPACE_WRAPPER",
            "CARGO_BUILD_RUSTC_WORKSPACE_WRAPPER",
        ]
    );
    let removed = |command: &Command| {
        scratch_cargo::WRAPPER_VARS.iter().all(|name| {
            command
                .get_envs()
                .any(|(key, value)| key == OsStr::new(name) && value.is_none())
        })
    };
    assert!(removed(&scratch_cargo::cargo_at(
        scratch_cargo::cargo_path()
    )));
    let mut set = Command::new("true");
    for name in scratch_cargo::WRAPPER_VARS {
        set.env(name, "/nonexistent/moai-no-wrapper");
    }
    assert!(!removed(&set));
    scratch_cargo::without_wrapper(&mut set);
    assert!(removed(&set));
}
