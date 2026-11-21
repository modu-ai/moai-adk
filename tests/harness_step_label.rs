//! SPEC-20261121-harness-step-label: the step of `moai project-update-plan` that writes the managed
//! `AGENTS.md` section and the project files shows people the name `Project harness`
//! (`프로젝트 맞춤 하네스` in Korean), while its id `generation`, its commands and its manifest
//! stay. The tests read the source and the two workflow texts; no command, host or HOME runs.

use std::fs;

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The value of `const INSTALL_REASON: &str = "<value>";` with its line continuations joined.
fn install_reason() -> String {
    let source = read("src/project_update.rs");
    let head = "const INSTALL_REASON: &str = \"";
    let start = source
        .find(head)
        .unwrap_or_else(|| panic!("src/project_update.rs has no `{head}`"))
        + head.len();
    let end = source[start..].find('"').expect("closing quote") + start;
    let mut out = String::new();
    let mut rest = &source[start..end];
    while let Some(at) = rest.find("\\\n") {
        out.push_str(&rest[..at]);
        rest = rest[at + 2..].trim_start();
    }
    out.push_str(rest);
    out
}

#[test]
fn install_reason_names_the_step_by_its_display_name() {
    let reason = install_reason();
    assert!(
        reason.contains("project harness"),
        "the install phase reason must name `project harness`: {reason}"
    );
}

#[test]
fn install_reason_does_not_say_generation() {
    let reason = install_reason();
    assert!(
        !reason.contains("generation"),
        "the install phase reason still says `generation`: {reason}"
    );
}

#[test]
fn install_reason_keeps_the_other_steps_in_order() {
    let reason = install_reason();
    let mut at = 0;
    for step in [
        "profile",
        "project harness",
        "Claude Code settings",
        "Claude Code hooks",
        "GitHub templates",
        "the Codex and OMP hook sets",
    ] {
        let found = reason[at..]
            .find(step)
            .unwrap_or_else(|| panic!("`{step}` missing or out of order in: {reason}"));
        at += found + step.len();
    }
}

#[test]
fn identifiers_stay_as_they_were() {
    let update = read("src/project_update.rs");
    for needle in [
        "name: \"generation\"",
        "argv(\"generation-apply\"",
        "manifests/generation.json",
    ] {
        assert!(
            update.contains(needle),
            "src/project_update.rs lost `{needle}`"
        );
    }
    let help = read("src/help.rs");
    for needle in [
        "(\"generation-plan\", false)",
        "(\"generation-apply\", false)",
    ] {
        assert!(help.contains(needle), "src/help.rs lost `{needle}`");
    }
}

/// The line of `rel` that holds the display-name sentence, or none.
fn naming_line(rel: &str) -> Option<String> {
    read(rel)
        .lines()
        .find(|l| l.contains("Project harness") && l.contains("프로젝트 맞춤 하네스"))
        .map(str::to_string)
}

#[test]
fn workflow_text_names_the_step_in_english_and_korean() {
    for rel in [
        "skills/moai/workflows/update.md",
        "skills/moai/workflows/project.md",
    ] {
        assert!(
            naming_line(rel).is_some(),
            "{rel} has no line naming `Project harness` and `프로젝트 맞춤 하네스`"
        );
    }
}

#[test]
fn workflow_text_tells_to_translate_and_keeps_the_id_for_tools() {
    for rel in [
        "skills/moai/workflows/update.md",
        "skills/moai/workflows/project.md",
    ] {
        let line = naming_line(rel).unwrap_or_else(|| panic!("{rel} has no naming line"));
        assert!(
            line.contains("translate"),
            "{rel}: the naming line must tell to translate the name for another language: {line}"
        );
        assert!(
            line.contains("`generation`"),
            "{rel}: the naming line must keep the id `generation` for tools: {line}"
        );
    }
}
