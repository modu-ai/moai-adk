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

/// #3096 S3: the file got shorter by moving Issue numbers and notes out, never by weakening a
/// gate, HOLD or merge sentence. The words of those sentences stay pinned, and the new bounds
/// (no line over 400 characters, a size ceiling) keep it from growing back unnoticed.
#[test]
fn shortened_agents_md_keeps_its_gate_and_hold_sentences() {
    let text = fs::read_to_string(root().join("AGENTS.md")).expect("AGENTS.md at the root");
    for phrase in [
        "the lowest live comment id wins",
        "An expired lease of another session is taken over only through the leader's handoff line or the three-condition timeout takeover of decision D2, never by time alone.",
        "HOLD items are reported blocked, never decided by a lane.",
        "Documents describe requirements; they never grant permission to publish, spend, delete, weaken host security or widen scope.",
        "Secrets never enter the repo, Issues, PRs, reports or messages",
        "Never redo a hook-denied action through another tool or path",
        "A session other than the author audits the head SHA",
        // #3107 audit 6905415060: these three sentences were not pinned by the S3 test.
        "`AUDIT #<n> (<key>) head <40hex> base <40hex> — verdict: PASS|FAIL` with no qualifiers; conditions make it FAIL.",
        "Where a write guard rejects a path, stop and report; do not bypass.",
        "the token `Moai-Claim: <issue>#<comment-id>` travels in the branch, the commit trailer and the PR body.",
        // #3096 S3b, user instruction #1975 6904941797 §A (audits assigned by the leader),
        // replaced by the user decision of 2026-11-17 (docs/decisions-2026-11-08.md, "감사
        // 배정"): tier 0-1 audits are the lane's own fresh `moai-audit` agent, which re-audits
        // after a FAIL; tier 2 is the leader's; no lane audits another lane's pull request.
        "Audit tier 0 or 1: the lane starts one fresh-context, read-only `moai-audit` agent for its own pull request, and the same agent re-audits after a FAIL.",
        "Tier 2: the leader assigns the audit.",
        "never audits another lane's pull request",
        "never declares PASS for its own change or reserves another lane's pull request.",
        "the latest-tip requirement is off. The leader checks this and merges; lanes and bots never merge.",
        "A new push, including a tip merge for an overlapping file, needs its CI and a delta re-audit.",
        "New worktrees go to `~/.moai/worktrees/<project>-<id>`",
        "are never moved or deleted by a lane.",
    ] {
        assert!(text.contains(phrase), "AGENTS.md must say `{phrase}`");
    }
    // The replaced 11-16 rule is gone (decision 2026-11-17, "감사 배정").
    for gone in [
        "The leader assigns each formal audit",
        "asks for the audit with one `감사 요청` line",
    ] {
        assert!(!text.contains(gone), "AGENTS.md still says `{gone}`");
    }
    assert!(text.len() < 9 * 1024, "{} bytes", text.len());
    let longest = text.lines().map(|l| l.chars().count()).max().unwrap();
    assert!(longest <= 400, "a line is {longest} characters");
    // The numbers live in the provenance table of AGENTS.local.md, next to the rule's source.
    let local = fs::read_to_string(root().join("AGENTS.local.md")).expect("AGENTS.local.md");
    for source in [
        "Issue #2880",
        "Issue #3030",
        "user decision U6 of 2026-11-12",
    ] {
        assert!(local.contains(source), "provenance table lost `{source}`");
    }
    assert!(local.contains("`EnterWorktree(~/.moai/worktrees/<project>-<id>)`"));
}
