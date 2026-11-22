//! The CI questions of the interview and the CI section of the PROJECT workflow
//! (SPEC-20261122-project-ci-design, slice C). The skill source files are the instructions a
//! session follows, so what they promise is pinned here: the six plain questions with their
//! ids, one selection each and the recommended answer first, no technical word in a question
//! the user reads, the profile fields they fill, and the commands the section runs.

use std::path::Path;

fn skill(rel: &str) -> String {
    std::fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap()
}

const IDS: [&str; 6] = [
    "ci_stack",
    "ci_kind",
    "ci_users",
    "ci_deploy",
    "ci_money",
    "ci_strict",
];

/// The table rows of the CI section of the interview note, as their cells.
fn rows(interview: &str) -> Vec<Vec<String>> {
    let section = interview
        .split("## CI questions")
        .nth(1)
        .expect("the interview note has a CI questions section");
    let section = section.split("\n## ").next().unwrap();
    section
        .lines()
        // The ids are bold, not in backticks: the older question-bank test reads every
        // backticked row after "## Question bank" and pins exactly its own list.
        .filter(|l| l.starts_with("| **ci_"))
        .map(|l| {
            l.trim_matches('|')
                .split('|')
                .map(|c| c.trim().to_string())
                .collect()
        })
        .collect()
}

#[test]
fn the_interview_note_has_the_six_ci_rows_in_order() {
    let interview = skill("references/internal/interview.md");
    let rows = rows(&interview);
    let ids: Vec<String> = rows
        .iter()
        .map(|r| r[0].trim_matches('*').to_string())
        .collect();
    assert_eq!(ids, IDS, "{rows:?}");
    for r in &rows {
        assert_eq!(
            r.len(),
            5,
            "id, ask when, question, options, goes to: {r:?}"
        );
        assert!(
            r[3].contains("recommended"),
            "the recommended option is marked: {r:?}"
        );
    }
}

/// A question and its options are read by the user, so none of them carries a technical word
/// the user was never told: the section explains each word once in its own prose instead.
#[test]
fn no_question_or_option_carries_a_technical_word() {
    let interview = skill("references/internal/interview.md");
    for r in rows(&interview) {
        for cell in [&r[2], &r[3]] {
            for word in [
                "workflow",
                "runner",
                "matrix",
                "Dependabot",
                "commit hash",
                "YAML",
                "cron",
                "pipeline",
                "artifact",
                "concurrency",
                "permissions",
            ] {
                assert!(
                    !cell.to_lowercase().contains(&word.to_lowercase()),
                    "`{word}` in the {} row: {cell}",
                    r[0]
                );
            }
        }
    }
}

#[test]
fn the_section_names_where_each_answer_goes_and_the_rules_around_them() {
    let interview = skill("references/internal/interview.md");
    let section = interview.split("## CI questions").nth(1).unwrap();
    for need in [
        "`ci`",
        "`checks`",
        "macos_price_ack",
        "monthly_budget_usd",
        "dependabot",
        "at most three questions per call",
        "never guess",
        "in the user's language",
    ] {
        assert!(section.contains(need), "the CI section lacks {need:?}");
    }
}

#[test]
fn the_project_workflow_runs_the_ci_commands_after_the_pick() {
    let project = skill("workflows/project.md");
    let section = project
        .split("## CI files")
        .nth(1)
        .expect("the PROJECT workflow has a CI files section");
    for need in [
        "project-update-plan",
        "project-ci-plan",
        "project-ci-apply",
        "`remove`",
        "project-check",
        "never edit",
        "internal/interview",
    ] {
        assert!(
            section.contains(need),
            "the CI files section lacks {need:?}"
        );
    }
}
