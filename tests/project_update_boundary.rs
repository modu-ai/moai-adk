//! SPEC-20261120-project-update-boundary: `/moai project` describes the project only, and
//! `/moai update` owns the earlier-generation moves, the installs and the merges. The tests read
//! the shipped workflow text only; no host, no HOME and no command runs here.

use std::fs;

const BOUNDARY: &str = "`/moai project` only describes the project; moving the earlier generation's state aside and installing or merging MoAI's managed files belongs to `/moai update`.";
const ONE_LINE: &str =
    "When `moai project-update-plan` still shows a step that does not read `unchanged`, end with this one line: \"Run `/moai update` to finish the remaining updates.\"";

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn project() -> String {
    read("skills/moai/workflows/project.md")
}

fn update() -> String {
    read("skills/moai/workflows/update.md")
}

#[test]
fn project_holds_no_move_install_or_merge_step() {
    let project = project();
    for word in [
        "go-settings-apply",
        "go-state-plan",
        "go-state-apply",
        "go-state-revert",
        "worktree-move-plan",
        "worktree-move-apply",
        "claude-settings-plan",
        "claude-settings-apply",
        "hooks-apply",
        "agents-merge",
        "Earlier-generation state",
        "state_root.generation",
        "baseRef",
    ] {
        assert!(!project.contains(word), "project.md still holds {word:?}");
    }
    let update = update();
    for word in [
        "go-settings-apply",
        "go-state-plan",
        "go-state-apply",
        "go-state-revert",
        "worktree-move-plan",
        "worktree-move-apply",
        "claude-settings-plan",
        "worktree.baseRef",
        "folder_stays",
        "root_files",
    ] {
        assert!(update.contains(word), "update.md lacks {word:?}");
    }
    // The two Codex and m-flow sentences of the worktree base setting.
    assert!(
        update.contains("Codex: CLI --worktree starts from the current HEAD, which matches m-flow; in the Codex app pick the integration branch when creating a worktree. No Codex setting is written."),
        "update.md lacks the Codex worktree sentence"
    );
}

#[test]
fn the_boundary_sentence_is_in_both_files_and_project_ends_with_one_line() {
    let project = project();
    let update = update();
    assert!(
        project.contains(BOUNDARY),
        "project.md lacks the boundary sentence"
    );
    assert!(
        update.contains(BOUNDARY),
        "update.md lacks the boundary sentence"
    );
    let closing = project
        .find("Offer the closing commit")
        .expect("project.md keeps the closing commit step");
    let one_line = project
        .find(ONE_LINE)
        .unwrap_or_else(|| panic!("project.md lacks the one line: {ONE_LINE}"));
    assert!(
        one_line > closing,
        "the one line comes after the closing commit step"
    );
    assert_eq!(
        project.matches(ONE_LINE).count(),
        1,
        "the one line is stated once"
    );
}

#[test]
fn project_asks_little_and_offers_no_decision_model() {
    let project = project();
    for sentence in [
        "at most three questions per call and at most four calls",
        "A value that reading already knows is confirmed with the user, not asked.",
        "the conversation language, the language of commit messages and the language of documents",
    ] {
        assert!(project.contains(sentence), "project.md lacks: {sentence}");
    }
    for word in [
        "Decision model",
        "decision_model",
        "decision model",
        "at most three project agents",
        "purpose, users",
        "ask for purpose",
    ] {
        assert!(!project.contains(word), "project.md still says {word:?}");
    }
    // Step 6 keeps no count; the way proposals are made is one line.
    assert!(
        project.contains(
            "How the proposals are made is set by the harness proposal; offer none until it is."
        ),
        "project.md lacks the one-line pointer for project agents and skills"
    );
}

/// Shapes the rewritten workflow files must not carry.
fn forbidden_in(text: &str) -> Vec<String> {
    let mut hits = Vec::new();
    for word in [
        "OMP",
        "/skill:moai",
        "user decision",
        "Issue #",
        "SPEC-2026",
        "tools/local_check.sh",
        "check-guard",
        "Kiro",
        "Cursor",
    ] {
        if text.contains(word) {
            hits.push(word.to_string());
        }
    }
    let bytes = text.as_bytes();
    for (i, _) in text.match_indices("C-") {
        let digits = bytes[i + 2..]
            .iter()
            .take_while(|b| b.is_ascii_digit())
            .count();
        let before_ok = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
        if digits > 0 && before_ok {
            hits.push(text[i..i + 2 + digits].to_string());
        }
    }
    for (i, _) in text.match_indices('#') {
        let digits = bytes[i + 1..]
            .iter()
            .take_while(|b| b.is_ascii_digit())
            .count();
        if digits >= 3 {
            hits.push(text[i..i + 1 + digits].to_string());
        }
    }
    hits
}

#[test]
fn the_workflows_name_no_card_issue_or_omp() {
    for rel in [
        "skills/moai/workflows/project.md",
        "skills/moai/workflows/update.md",
    ] {
        let hits = forbidden_in(&read(rel));
        assert!(hits.is_empty(), "{rel} still names {hits:?}");
    }
}
