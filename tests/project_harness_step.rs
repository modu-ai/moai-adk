//! SPEC-20261120-project-harness-research: step 6 of the project workflow proposes a project
//! harness (agents and skills) that is studied from the detected project and researched on the
//! web, with no fixed count, a source per proposal, a several-of pick and a no-web fallback.
//! Reads the shipped text; writes nothing.

use std::path::Path;

fn read(rel: &str) -> String {
    std::fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)).unwrap()
}

/// Collapse every run of whitespace to one space, so a phrase may wrap across lines.
fn flat(text: &str) -> String {
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

/// Step 6 of the Procedure: from its number to the line of step 7.
fn step_six(text: &str) -> String {
    let start = text
        .find("\n6. **")
        .expect("step 6 of the Procedure is missing");
    let rest = &text[start + 1..];
    let end = rest.find("\n7. **").expect("step 7 is missing");
    flat(&rest[..end])
}

#[test]
fn harness_step_studies_the_project_and_searches_for_sources_with_a_read_time() {
    let step = step_six(&read("skills/moai/workflows/project.md"));
    for phrase in [
        "study the detected project",
        "languages, frameworks, test and build tools, CI and folders",
        "current practice and well-regarded sources",
        "URL and read time",
    ] {
        assert!(step.contains(phrase), "step 6 must say {phrase:?}:\n{step}");
    }
}

#[test]
fn harness_step_has_no_fixed_count_and_each_proposal_names_its_reason_and_sources() {
    let step = step_six(&read("skills/moai/workflows/project.md"));
    assert!(
        !step.contains("at most three"),
        "the cap of three project agents is gone:\n{step}"
    );
    for phrase in [
        "no fixed count",
        "its purpose",
        "its scope",
        "the reason it is proposed",
        "its sources",
        "up to sixteen",
    ] {
        assert!(step.contains(phrase), "step 6 must say {phrase:?}:\n{step}");
    }
}

#[test]
fn harness_step_offers_a_pick_creates_nothing_unpicked_and_stays_portable() {
    let step = step_six(&read("skills/moai/workflows/project.md"));
    for phrase in [
        "several-of selection",
        "recommended",
        "create nothing the user did not pick",
        "no internal tool of this project, no secret, no absolute home path",
    ] {
        assert!(step.contains(phrase), "step 6 must say {phrase:?}:\n{step}");
    }
}

#[test]
fn harness_step_proposes_from_the_detected_facts_when_web_search_is_unavailable() {
    let step = step_six(&read("skills/moai/workflows/project.md"));
    for phrase in [
        "When web search is unavailable",
        "from the detected facts alone",
        "no source was read",
    ] {
        assert!(step.contains(phrase), "step 6 must say {phrase:?}:\n{step}");
    }
}

#[test]
fn harness_step_keeps_the_generation_plan_contract() {
    let step = step_six(&read("skills/moai/workflows/project.md"));
    for phrase in [
        "`native_roles`",
        "`project_skills`",
        "`prj-<name>`",
        "`generation-apply`",
    ] {
        assert!(
            step.contains(phrase),
            "step 6 must keep {phrase:?}:\n{step}"
        );
    }
}
