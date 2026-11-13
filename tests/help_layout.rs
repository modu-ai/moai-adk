//! #2802 S1 (design #2799 D2 sections 9 and 11): the people-facing help, pinned as goldens, and
//! the machine envelope left untouched by it.
//!
//! The goldens are the help a person sees when stdout is not a terminal (no logo): the root help
//! and `help --all`, in the default width, `NO_COLOR=1 TERM=dumb`, and `COLUMNS=60`. Rewrite
//! them with `MOAI_HELP_GOLDEN_WRITE=1 cargo test --test help_layout` and review the diff.
mod common;

use common::CleanEnv;
use std::path::{Path, PathBuf};
use std::process::Command;

fn golden_dir() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("tests/fixtures/help")
}

fn run(bin: &str, args: &[&str], env: &[(&str, &str)]) -> std::process::Output {
    let home = tempfile::TempDir::new().unwrap();
    let mut cmd = Command::new(bin);
    cmd.args(args)
        .clean_env()
        .env("HOME", home.path())
        .env("PATH", "/usr/bin:/bin");
    for (k, v) in env {
        cmd.env(k, v);
    }
    cmd.output().unwrap()
}

/// (golden name, arguments, environment).
type Case = (
    &'static str,
    &'static [&'static str],
    &'static [(&'static str, &'static str)],
);

const CASES: [Case; 6] = [
    ("root", &["--help"], &[]),
    (
        "root-dumb",
        &["--help"],
        &[("NO_COLOR", "1"), ("TERM", "dumb")],
    ),
    ("root-60", &["--help"], &[("COLUMNS", "60")]),
    ("all", &["help", "--all"], &[]),
    (
        "all-dumb",
        &["help", "--all"],
        &[("NO_COLOR", "1"), ("TERM", "dumb")],
    ),
    ("all-60", &["help", "--all"], &[("COLUMNS", "60")]),
];

#[test]
fn help_matches_its_goldens() {
    let write = std::env::var_os("MOAI_HELP_GOLDEN_WRITE").is_some();
    let mut failed = Vec::new();
    for (name, args, env) in CASES {
        let out = run(common::MOAI_BIN, args, env);
        assert!(out.status.success(), "{name}: {out:?}");
        let text = String::from_utf8(out.stdout).unwrap();
        let path = golden_dir().join(format!("{name}.txt"));
        if write {
            std::fs::create_dir_all(golden_dir()).unwrap();
            std::fs::write(&path, &text).unwrap();
            continue;
        }
        let want = std::fs::read_to_string(&path)
            .unwrap_or_else(|_| panic!("missing golden {}", path.display()));
        if text != want {
            failed.push(name);
        }
    }
    assert!(
        failed.is_empty(),
        "help differs from its golden: {failed:?}"
    );
}

/// D2 section 9: the root help stays near 40 lines and inside 60 columns at `COLUMNS=60`, under
/// both names.
#[test]
fn the_root_help_is_short_and_fits_sixty_columns() {
    for bin in [common::MOAI_BIN, common::MOAI_CLI_BIN] {
        let wide = String::from_utf8(run(bin, &["--help"], &[]).stdout).unwrap();
        assert!(
            wide.lines().count() <= 45,
            "{bin}: {} lines",
            wide.lines().count()
        );
        let narrow = String::from_utf8(run(bin, &["--help"], &[("COLUMNS", "60")]).stdout).unwrap();
        for line in narrow.lines() {
            assert!(line.chars().count() <= 60, "{bin}: {line:?}");
        }
    }
}

/// `Usage:` names the binary that ran (D2 section 9); `--version` stays the one signature line
/// (D3 decision 2).
#[test]
fn usage_names_the_typed_binary_and_the_version_line_is_unchanged() {
    for (bin, name) in [
        (common::MOAI_BIN, "moai"),
        (common::MOAI_CLI_BIN, "moai-cli"),
    ] {
        let help = String::from_utf8(run(bin, &["--help"], &[]).stdout).unwrap();
        assert!(
            help.contains(&format!("Usage: {name} <command> [options]")),
            "{help}"
        );
        assert!(
            help.contains(&format!("Run '{name} help --all' for every command.")),
            "{help}"
        );
        let sub = String::from_utf8(run(bin, &["identity", "--help"], &[]).stdout).unwrap();
        assert!(sub.contains(&format!("Usage: {name} identity")), "{sub}");
        let version = String::from_utf8(run(bin, &["--version"], &[]).stdout).unwrap();
        assert!(version.starts_with("moai-cli ") && version.trim_end().ends_with(" (rust)"));
    }
}

/// The help layer never reaches stdout's envelope (D2 section 1): a usage error prints the same
/// envelope bytes under both names, while only stderr names the binary.
#[test]
fn a_usage_error_envelope_is_byte_identical_under_both_names() {
    for args in [&["identity"][..], &["go-state-plan", "--nope", "."][..]] {
        let a = run(common::MOAI_BIN, args, &[]);
        let b = run(common::MOAI_CLI_BIN, args, &[]);
        assert_eq!(a.status.code(), Some(2), "{args:?}");
        assert_eq!(a.stdout, b.stdout, "{args:?}");
        let envelope: serde_json::Value = serde_json::from_slice(&a.stdout).unwrap();
        assert_eq!(envelope["ok"], serde_json::json!(false), "{args:?}");
        assert_eq!(
            envelope["error"]["kind"],
            serde_json::json!("usage"),
            "{args:?}"
        );
    }
}

/// Every command appears in `help --all`, and the root help lists only the root ones.
#[test]
fn help_all_lists_every_command_and_the_root_a_subset() {
    let all = String::from_utf8(run(common::MOAI_BIN, &["help", "--all"], &[]).stdout).unwrap();
    let root = String::from_utf8(run(common::MOAI_BIN, &["--help"], &[]).stdout).unwrap();
    // A row is `  <word>` then spaces and its about, or `  <word>` alone when it has none.
    let listed = |text: &str, word: &str| {
        text.lines().any(|l| {
            l.strip_prefix("  ")
                .and_then(|r| r.strip_prefix(word))
                .is_some_and(|r| r.is_empty() || r.starts_with(' '))
        })
    };
    for g in moai_cli::help::GROUPS {
        assert!(all.contains(&format!("\n{}\n", g.heading)), "{}", g.heading);
        for (word, on_root) in g.commands {
            assert!(listed(&all, word), "help --all lacks {word}");
            if !["cc", "codex", "glm"].contains(word) {
                assert_eq!(listed(&root, word), *on_root, "root help and {word}");
            }
        }
    }
    assert!(root.contains("  cc, codex, glm "), "{root}");
}
