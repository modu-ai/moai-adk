//! SPEC-20261121-generic-flow-questions: the project questions about the git flow and about
//! autonomy use plain words and recommend GitHub flow. The tests read the question bank and the
//! PROJECT workflow as text; no command, host or HOME runs here.

use std::fs;

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn bank() -> String {
    read("skills/moai/references/internal/interview.md")
}

/// The cells of the table row `| `id` | …`: id, when, question, options, goes to.
fn row(id: &str) -> Vec<String> {
    bank()
        .lines()
        .find(|l| l.starts_with(&format!("| `{id}` |")))
        .unwrap_or_else(|| panic!("no `{id}` row"))
        .trim_matches('|')
        .split(" | ")
        .map(|c| c.trim().to_string())
        .collect()
}

/// The section `## Git work flow …` up to the end of the file or the next `## ` heading.
fn flow_section() -> String {
    let text = bank();
    let start = text.find("## Git work flow").expect("flow section");
    let rest = &text[start + 3..];
    let end = rest.find("\n## ").map(|at| at + 3).unwrap_or(rest.len());
    text[start..start + end].to_string()
}

#[test]
fn work_flow_row_lists_github_flow_first_and_recommended() {
    let options = &row("work_flow")[3];
    assert!(
        options.starts_with("GitHub flow (recommended)"),
        "the work_flow options must start with `GitHub flow (recommended)`: {options}"
    );
}

#[test]
fn work_flow_row_describes_m_flow_in_plain_words() {
    let options = &row("work_flow")[3];
    assert!(options.contains("m-flow"), "{options}");
    assert!(
        options.contains("when several AI agents work at the same time"),
        "the m-flow option must say when it fits, in plain words: {options}"
    );
    for other in ["trunk-based", "git-flow"] {
        assert!(
            options.contains(other),
            "the options lost {other}: {options}"
        );
    }
}

#[test]
fn work_flow_row_uses_no_maintainer_words() {
    let options = row("work_flow")[3].to_lowercase();
    for word in ["spec", "worktree", "audit", "leader", "card"] {
        assert!(
            !options.contains(word),
            "the work_flow options show the user the word `{word}`: {options}"
        );
    }
}

#[test]
fn flow_section_names_github_flow_the_recommended_default() {
    let section = flow_section();
    assert!(
        section.contains("GitHub flow is the recommended default"),
        "the flow section must name GitHub flow the recommended default"
    );
    assert!(
        !section.contains("m-flow is the recommended default"),
        "the flow section still calls m-flow the recommended default"
    );
}

#[test]
fn flow_section_tells_the_model_to_describe_m_flow_in_plain_words_first() {
    let section = flow_section();
    for phrase in [
        "describe m-flow to the user in plain words",
        "until the user has picked it",
    ] {
        assert!(
            section.contains(phrase),
            "the flow section lacks {phrase:?}"
        );
    }
}

#[test]
fn autonomy_row_says_what_is_asked_first_in_plain_words() {
    let cells = row("autonomy");
    assert!(
        cells[3].contains("hard to undo, or touches money, secrets or deletion"),
        "the autonomy options must name what is asked first in plain words: {}",
        cells[3]
    );
}

#[test]
fn autonomy_row_shows_the_user_no_internal_terms() {
    let cells = row("autonomy");
    for (name, cell) in [("question", &cells[2]), ("options", &cells[3])] {
        for term in ["case (2)", "SKILL.md"] {
            assert!(
                !cell.contains(term),
                "the autonomy {name} shows the user `{term}`: {cell}"
            );
        }
    }
}

#[test]
fn autonomy_sorting_rule_for_the_model_stands_outside_the_table() {
    let outside: String = bank()
        .lines()
        .filter(|l| !l.starts_with('|'))
        .collect::<Vec<_>>()
        .join("\n");
    assert!(
        outside.contains("Sort the `autonomy` answer the way SKILL.md sorts an open question"),
        "the model's sorting rule must be a sentence outside the table"
    );
}

#[test]
fn project_workflow_asks_the_work_flow_row_with_github_flow_first() {
    let text = read("skills/moai/workflows/project.md");
    let line = text
        .lines()
        .find(|l| l.contains("the `work_flow` row"))
        .expect("project.md has no line asking the work_flow row");
    assert!(
        line.contains("GitHub flow first and recommended"),
        "project.md must ask GitHub flow first and recommended: {line}"
    );
    assert!(
        !line.contains("m-flow first and recommended"),
        "project.md still asks m-flow first and recommended: {line}"
    );
    assert!(
        line.contains(
            "m-flow (described to the user as \"when several AI agents work at the same time\")"
        ),
        "project.md must tell the model to describe m-flow in plain words: {line}"
    );
}
