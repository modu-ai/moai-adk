//! SPEC-20261120-codex-worktree-base-note: when the pick is m-flow and `codex` is one of the hosts,
//! the PROJECT workflow says where Codex worktrees start and writes no Codex key.
//!
//! Codex has no setting for the start point of its worktrees (config keys, profile keys,
//! environment variables and flags were all read in its documents and source); its CLI
//! `--worktree` starts from the current HEAD, which is what m-flow needs.

use std::path::Path;

fn project_workflow() -> String {
    std::fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/project.md"),
    )
    .expect("skills/moai/workflows/project.md is readable")
}

/// REQ-1 and REQ-2: three short phrases say what Codex does, and no line that names Codex also
/// names a `baseRef` key.
#[test]
fn the_project_workflow_says_where_codex_worktrees_start() {
    let text = project_workflow();
    for phrase in [
        "Codex has no setting for the start point of its worktrees",
        "`--worktree` starts from the current HEAD",
        "pick the integration branch below the composer",
    ] {
        assert!(
            text.contains(phrase),
            "project.md must say {phrase:?} when the pick is m-flow and codex is a host"
        );
    }
    for line in text.lines() {
        let lower = line.to_ascii_lowercase();
        if lower.contains("codex") && line.contains("baseRef") {
            // The Claude sentence names both hosts in one bullet; only a line that gives Codex a
            // key of its own is wrong.
            assert!(
                !lower.contains("codex's baseref")
                    && !lower.contains("codex baseref")
                    && !lower.contains("codex-config-plan"),
                "project.md must not give Codex a worktree start key: {line}"
            );
        }
    }
}
