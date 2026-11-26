//! The CI step of the PROJECT workflow (SPEC-20261122-project-ci-design, REQ-1, REQ-2, AC-1,
//! D-7 and D-8). The skill source files are the instructions a session follows, so what they
//! promise is pinned here: the step stands outside the interview and its question bank, comes
//! after the interview and the applied project files, is offered with one selection, asks the
//! six plain rows in at most two calls, saves the answers through a plan the user approves, and
//! writes files only after a pick of their own.

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

const HEADING: &str = "## CI step";

fn project() -> String {
    skill("workflows/project.md")
}

/// The CI step section of the PROJECT workflow, up to the next `## ` heading.
fn section() -> String {
    let text = project();
    assert_eq!(
        text.lines().filter(|l| l.starts_with(HEADING)).count(),
        1,
        "the PROJECT workflow has exactly one CI step section"
    );
    let after = &text[text.find(&format!("\n{HEADING}")).unwrap() + 1..];
    let end = after[HEADING.len()..]
        .find("\n## ")
        .map_or(after.len(), |at| at + HEADING.len());
    after[..end].to_string()
}

/// The question rows of the CI step, as their cells.
fn rows() -> Vec<Vec<String>> {
    section()
        .lines()
        .filter(|l| l.starts_with("| **ci_"))
        .map(|l| {
            l.trim_matches('|')
                .split('|')
                .map(|c| c.trim().to_string())
                .collect()
        })
        .collect()
}

fn position(text: &str, needle: &str) -> usize {
    text.find(needle)
        .unwrap_or_else(|| panic!("{needle:?} is missing from: {text}"))
}

fn holds(text: &str, needles: &[&str]) {
    for need in needles {
        assert!(text.contains(need), "the CI step lacks {need:?}: {text}");
    }
}

/// REQ-1, D-8: the step comes after the interview step and after the project files are applied,
/// both by where it stands and by what it says.
#[test]
fn the_ci_step_comes_after_the_interview_and_the_applied_project_files() {
    let text = project();
    let at = position(&text, &format!("\n{HEADING}"));
    for step in [
        "\n3. **Interview.**",
        "\n4. **Plan and apply.**",
        "\n5. **Refresh the managed section.**",
    ] {
        assert!(
            position(&text, step) < at,
            "the CI step stands before {step:?}"
        );
    }
    holds(
        &section(),
        &[
            "not part of the interview",
            "after that interview is finished and the project files are applied",
            "`project-apply` of step 4",
        ],
    );
    let procedure = &text[position(&text, "\n## Procedure")..position(&text, "\n## Resume")];
    assert!(
        procedure.contains("run the CI step"),
        "no step of the procedure sends the session to the CI step"
    );
}

/// REQ-1, D-8: one selection offers the step, only while the profile holds no CI answers, and
/// "not now" asks nothing, records nothing and is offered again.
#[test]
fn the_offer_is_one_selection_and_not_now_asks_and_records_nothing() {
    let s = section();
    holds(
        &s,
        &[
            "Only while the profile holds no CI answers",
            "one selection of its own",
            "\"set up automatic checks now\" (recommended)",
            "\"not now\"",
            "ask nothing more",
            "record nothing",
            "offered again on the next run",
            "A cancelled, empty or unanswered offer is \"not now\"",
        ],
    );
    assert!(
        position(&s, "\"set up automatic checks now\" (recommended)") < position(&s, "\"not now\""),
        "the recommended answer is not first"
    );
}

/// REQ-1, REQ-2: the six rows in order, each with its recommended answer and its destination.
#[test]
fn the_ci_step_has_the_six_rows_in_order_with_their_destinations() {
    let rows = rows();
    let ids: Vec<String> = rows
        .iter()
        .map(|r| r[0].trim_matches('*').to_string())
        .collect();
    assert_eq!(ids, IDS, "{rows:?}");
    let destinations: [&[&str]; 6] = [
        &["profile `checks`"],
        &["`ci.kind`"],
        &["`ci.systems`"],
        &["`ci.deploy`"],
        &["`ci.visibility`", "`ci.monthly_budget_usd`"],
        &["`ci.strict`"],
    ];
    for (r, needles) in rows.iter().zip(destinations) {
        assert_eq!(
            r.len(),
            5,
            "id, ask when, question, options, goes to: {r:?}"
        );
        assert!(
            r[3].contains("recommended"),
            "the recommended option is marked: {r:?}"
        );
        holds(&r[4], needles);
    }
}

/// A question and its options are read by the user, so none of them carries a technical word
/// the user was never told: the step explains each word once in its own prose instead.
#[test]
fn no_question_or_option_carries_a_technical_word() {
    let rows = rows();
    assert_eq!(rows.len(), IDS.len(), "no rows to read: {rows:?}");
    for r in rows {
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

/// REQ-1, REQ-2: the detected facts are confirmed first, the question calls are capped, and an
/// unanswered row is held, never guessed.
#[test]
fn the_detected_facts_come_first_and_the_question_calls_are_capped() {
    let s = section();
    holds(
        &s,
        &[
            "package managers",
            "each with its source",
            "a detected fact is confirmed, never asked as an open question",
            "the user can correct it",
            "at most six questions",
            "at most three questions per call",
            "at most two calls",
            "never guess",
            "HOLD",
            "ask it again on the next run",
            "in the user's language",
        ],
    );
    assert!(
        position(&s, "each with its source") < position(&s, "| **ci_stack**"),
        "the detected facts are confirmed before the first row"
    );
}

/// D-8: the answers are saved through one more plan and apply the user approves, and the file
/// pick comes after that, in a call of its own.
#[test]
fn the_answers_are_saved_by_an_approved_plan_before_the_file_pick() {
    let s = section();
    holds(
        &s,
        &[
            "`ci`",
            "`checks`",
            "one more `moai project-plan <project>`",
            "`moai project-apply <project>` only after the user approves",
            "the file pick comes in its own call",
            "A cancelled, empty or unanswered pick writes nothing",
        ],
    );
    let order = [
        "\"set up automatic checks now\" (recommended)",
        "| **ci_strict**",
        "one more `moai project-plan <project>`",
        "`moai project-update-plan <project>`",
        "the file pick comes in its own call",
    ];
    for pair in order.windows(2) {
        assert!(
            position(&s, pair[0]) < position(&s, pair[1]),
            "{:?} comes before {:?}",
            pair[0],
            pair[1]
        );
    }
}

/// The rules around the files: conflict and stale files, the removal pick, the deploy file, the
/// settings shown as text and the local check.
#[test]
fn the_ci_step_runs_the_ci_commands_after_the_pick() {
    holds(
        &section(),
        &[
            "project-update-plan",
            "project-ci-plan",
            "project-ci-apply",
            "explain every job in plain words",
            "the cheaper option first",
            "`conflict`",
            "`stale`",
            "`remove`",
            "the deploy file exists only after the yes",
            "holds no deploy command",
            "as text steps",
            "never edit a repository setting",
            "`moai project-check <project>`",
            "Never edit a generated workflow file by hand",
            "macos_price_ack",
            "ci.dependabot",
        ],
    );
}

/// D-7: the interview's question bank holds no CI section and no CI row.
#[test]
fn the_interview_question_bank_holds_no_ci_section_and_no_ci_row() {
    let bank = skill("references/internal/interview.md");
    for (at, line) in bank.lines().enumerate() {
        assert!(
            !(line.starts_with('#') && line.contains("CI ")),
            "interview.md:{} has a CI section: {line}",
            at + 1
        );
        assert!(
            !line.contains("ci_"),
            "interview.md:{} names a CI row: {line}",
            at + 1
        );
    }
}
