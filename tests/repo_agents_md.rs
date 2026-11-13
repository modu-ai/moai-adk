//! This repository's own `AGENTS.md` is tracked (user decision 2026-11-13, Issue #2667).
//! Claude Code and Codex read `AGENTS.md` only inside the session's own working tree; a lane
//! worktree of this repository therefore gets the project rules only when git carries the
//! file (measured in a temporary repository, Issue comment 6848071468).
mod common;

use std::{fs, path::Path};

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

/// #2667 AC-1: no ignore rule hides `AGENTS.md` again — a re-added `/AGENTS.md` line in
/// `.gitignore` would bring back worktrees without the project rules.
#[test]
fn agents_md_is_not_git_ignored() {
    let out = common::git::git_raw(
        root(),
        &["check-ignore", "--no-index", "-v", "--", "AGENTS.md"],
    );
    assert!(
        !out.success && out.stdout.is_empty(),
        "AGENTS.md is git-ignored again: {}{}",
        out.stdout,
        out.stderr
    );
    assert!(out.stderr.is_empty(), "check-ignore failed: {}", out.stderr);
}

/// #2667: the tracked file keeps the shape `agents-apply` maintains — exactly one managed
/// section, within Codex's 32 KiB instruction budget, and no machine-specific absolute path.
#[test]
fn tracked_agents_md_keeps_one_managed_section_and_no_machine_paths() {
    let text = fs::read_to_string(root().join("AGENTS.md")).expect("AGENTS.md at the root");
    assert_eq!(
        text.matches("<!-- moai:managed:begin id=moai-workflow -->")
            .count(),
        1
    );
    assert_eq!(
        text.matches("<!-- moai:managed:end id=moai-workflow -->")
            .count(),
        1
    );
    assert!(text.len() < 32 * 1024, "{} bytes", text.len());
    for machine in ["/Users/", "/home/", "C:\\Users\\"] {
        assert!(!text.contains(machine), "AGENTS.md names {machine}");
    }
}
