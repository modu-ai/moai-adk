//! This repository's own `CLAUDE.md` is tracked. It holds what `AGENTS.md` (tracked since the
//! user decision 2026-11-13, Issue #2667) and the shared rules of `AGENTS.local.md` held; both
//! originals are removed (user decision 2026-11-22, `SPEC-20261122-claude-md-layout` Plan 5,
//! AC-6). Claude Code loads `CLAUDE.md` from the session's own working tree, so a lane worktree
//! gets the project rules only when git carries the file; `CLAUDE.local.md` is personal and
//! stays ignored.
mod common;

use std::{fs, path::Path};

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

fn claude_md() -> String {
    fs::read_to_string(root().join("CLAUDE.md")).expect("CLAUDE.md at the root")
}

/// The first line of the repository-specific part, the former `AGENTS.local.md`.
const LOCAL_HEADING: &str = "## moai-adk local development rules";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-workflow -->";
const END: &str = "<!-- moai:managed:end id=moai-workflow -->";

/// The part of `CLAUDE.md` that came from `AGENTS.md`: the authored text before
/// [`LOCAL_HEADING`] and the managed section from its begin marker to the end of the file.
fn generic_part(text: &str) -> String {
    let (authored, rest) = text
        .split_once(&format!("\n{LOCAL_HEADING}\n"))
        .unwrap_or_else(|| panic!("CLAUDE.md has no `{LOCAL_HEADING}` heading"));
    let managed = &rest[rest.find(BEGIN).expect("the managed section")..];
    format!("{authored}\n{managed}")
}

/// #2667 AC-1, AC-6: no ignore rule hides `CLAUDE.md` — a re-added `/CLAUDE.md` line in
/// `.gitignore` would bring back worktrees without the project rules.
#[test]
fn claude_md_is_not_git_ignored() {
    let out = common::git::git_raw(
        root(),
        &["check-ignore", "--no-index", "-v", "--", "CLAUDE.md"],
    );
    assert!(
        !out.success && out.stdout.is_empty(),
        "CLAUDE.md is git-ignored: {}{}",
        out.stdout,
        out.stderr
    );
    assert!(out.stderr.is_empty(), "check-ignore failed: {}", out.stderr);
}

/// AC-6: the personal file stays ignored, and the two originals are no longer tracked.
#[test]
fn claude_local_md_is_ignored_and_the_agents_files_are_gone() {
    let out = common::git::git_raw(
        root(),
        &["check-ignore", "--no-index", "-v", "--", "CLAUDE.local.md"],
    );
    assert!(
        out.success && out.stdout.contains("CLAUDE.local.md"),
        "CLAUDE.local.md is not git-ignored: {}{}",
        out.stdout,
        out.stderr
    );
    // An index entry is looked up by name (`:<path>`), so the test reads only the named paths;
    // `Cargo.toml` shows the lookup itself works.
    let in_index =
        |path: &str| common::git::git_raw(root(), &["cat-file", "-e", &format!(":{path}")]).success;
    assert!(
        in_index("Cargo.toml"),
        "the index lookup does not work here"
    );
    for gone in ["AGENTS.md", "AGENTS.local.md"] {
        assert!(!in_index(gone), "{gone} is still tracked");
    }
}

/// #2667: the tracked file keeps the shape `agents-apply` maintains — exactly one managed
/// section, at the end, within the file's 44 KiB budget, and no machine-specific absolute path.
#[test]
fn tracked_claude_md_keeps_one_managed_section_and_no_machine_paths() {
    let text = claude_md();
    assert_eq!(text.matches(BEGIN).count(), 1);
    assert_eq!(text.matches(END).count(), 1);
    assert!(
        text.trim_end().ends_with(END),
        "the managed section is not the last part"
    );
    assert!(text.len() < 44 * 1024, "{} bytes", text.len());
    for machine in ["/Users/", "/home/", "C:\\Users\\"] {
        assert!(!text.contains(machine), "CLAUDE.md names {machine}");
    }
}

/// #3096 S3: the file got shorter by moving Issue numbers and notes out, never by weakening a
/// gate, HOLD or merge sentence. The words of those sentences stay pinned, and the bounds of
/// the part that came from `AGENTS.md` (no line over 400 characters, a size ceiling) keep it
/// from growing back unnoticed.
#[test]
fn shortened_generic_part_keeps_its_gate_and_hold_sentences() {
    let full = claude_md();
    let text = generic_part(&full);
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
        // Leader decision 2026-11-17 ~07:1xZ (docs/decisions-2026-11-08.md, "FAIL 판정과 tier 0
        // 병합"), as in the mode-leader merge-train: the PASS is per tier, a FAIL stops every tier.
        "Tier 1 or 2 also needs a non-author PASS that cites that run; tier 0 needs `ci` alone.",
        "A posted FAIL stops the merge at every tier.",
        "New worktrees go to `~/.moai/worktrees/<project>-<id>`",
        "are never moved or deleted by a lane.",
    ] {
        assert!(text.contains(phrase), "CLAUDE.md must say `{phrase}`");
    }
    // The replaced 11-16 rule is gone (decision 2026-11-17, "감사 배정").
    for gone in [
        "The leader assigns each formal audit",
        "asks for the audit with one `감사 요청` line",
        // The tier-blind merge rule replaced by the leader decision of 2026-11-17 ~07:1xZ.
        "a successful `ci` run and a non-author PASS that cites that run",
    ] {
        assert!(!text.contains(gone), "CLAUDE.md still says `{gone}`");
    }
    assert!(text.len() < 9 * 1024, "{} bytes", text.len());
    let longest = text.lines().map(|l| l.chars().count()).max().unwrap();
    assert!(longest <= 400, "a line is {longest} characters");
    // The numbers live in the provenance table of the repository part, next to the rule's source.
    for source in [
        "Issue #2880",
        "Issue #3030",
        "user decision U6 of 2026-11-12",
    ] {
        assert!(full.contains(source), "provenance table lost `{source}`");
        assert!(!text.contains(source), "the generic part names `{source}`");
    }
    assert!(full.contains("`EnterWorktree(~/.moai/worktrees/<project>-<id>)`"));
}
