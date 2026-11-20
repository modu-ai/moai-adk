//! SPEC-20261119-project-preview-order (card C-59, criterion 1): the PROJECT workflow walks the
//! `moai project-update-plan` preview in the order it prints and never asks about a step that
//! reads `blocked` or `re-run after …`. The tests read the workflow text and the source that
//! prints the waiting reasons, so the text only quotes what the code really says. No host, no
//! HOME and no command runs here.

use std::fs;

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn project_md() -> String {
    read("skills/moai/workflows/project.md")
}

fn update_md() -> String {
    read("skills/moai/workflows/update.md")
}

/// The numbered step `4. **Refresh the managed section` of PROJECT up to the next numbered step.
fn step_4() -> String {
    let text = project_md();
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("4. **Refresh the managed section") {
            inside = true;
        } else if inside && line.starts_with("5. ") {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "project.md has no step 4");
    out
}

/// The numbered step `<n>. **` of UPDATE whose number is `n`, up to the next numbered step.
fn update_step(n: u32) -> String {
    let text = update_md();
    let head = format!("{n}. **");
    let next = format!("{}. **", n + 1);
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with(&head) {
            inside = true;
        } else if inside && line.starts_with(&next) {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "update.md has no step {n}");
    out
}

/// The value of `pub const <name>: &str = "<value>";` in `src/project_update.rs`.
fn constant(name: &str) -> String {
    let source = read("src/project_update.rs");
    let head = format!("pub const {name}: &str = \"");
    let start = source
        .find(&head)
        .unwrap_or_else(|| panic!("src/project_update.rs has no `{head}`"))
        + head.len();
    let end = source[start..].find('"').expect("closing quote") + start;
    source[start..end].to_string()
}

const WAITING: [&str; 5] = [
    "AFTER_BACKUP",
    "AFTER_PROFILE",
    "AFTER_SETTINGS",
    "AFTER_GO_SETTINGS",
    "AFTER_INSTALL",
];

#[test]
fn step_4_asks_about_generation_only_when_it_reads_ready() {
    let step = step_4();
    for phrase in [
        "Ask about the `generation` step only when it reads `ready` in that run.",
        "run `moai project-update-plan <project>` again and ask from that new run",
        "because an earlier yes does not cover a step the user has not seen",
        "is not asked: report it with the reason the plan prints and leave it to `/moai update`",
    ] {
        assert!(step.contains(phrase), "step 4 lacks: {phrase}");
    }
    // The text speaks of the step in which `generation` is asked, before the apply it names.
    let ready = step.find("only when it reads `ready`").unwrap();
    let apply = step.find("run its printed `apply.argv`").unwrap();
    assert!(ready < apply, "the `ready` rule must come before the apply");
    // UPDATE holds the same rule for every other step.
    let ask = update_step(2);
    for phrase in [
        "never about a step that reads `blocked` or `re-run after …` behind an earlier one",
        "so no earlier yes covers it",
    ] {
        assert!(ask.contains(phrase), "update.md step 2 lacks: {phrase}");
    }
}

#[test]
fn update_asks_in_the_printed_phase_order() {
    let ask = update_step(2);
    let order =
        "phase `backup`, then `install`, then `merge`, each phase in the order of its steps";
    assert!(ask.contains(order), "update.md step 2 lacks: {order}");
    let preview_again = update_step(4);
    assert!(
        preview_again
            .contains("every install step reads `blocked` with `re-run after the backup phase`"),
        "update.md step 4 lacks the install waiting reason"
    );
    let report = update_step(5);
    assert!(
        report.contains("A `blocked` step is not offered: report it with what unblocks it"),
        "update.md step 5 lacks the blocked rule"
    );
    // The code prints the phases in that order.
    let source = read("src/project_update.rs");
    let at = |name: &str| {
        source
            .find(&format!("phase(\"{name}\""))
            .unwrap_or_else(|| panic!("src/project_update.rs has no phase `{name}`"))
    };
    assert!(at("backup") < at("install") && at("install") < at("merge"));
}

#[test]
fn quoted_waiting_reasons_are_the_ones_the_code_prints() {
    let text = format!("{}\n{}", project_md(), update_md());
    // Every waiting reason the text quotes in backticks is one the code prints.
    let known: Vec<String> = WAITING.iter().map(|n| constant(n)).collect();
    let mut quoted = 0;
    for span in text.split('`').skip(1).step_by(2) {
        if span.starts_with("re-run after") && !span.contains('…') {
            quoted += 1;
            assert!(
                known.iter().any(|k| k == span),
                "a workflow quotes `{span}`, which src/project_update.rs does not print"
            );
        }
    }
    assert!(
        quoted >= 4,
        "the text should quote the waiting reasons it relies on"
    );
    // The pairs the text names are the pairs the code builds.
    let source = read("src/project_update.rs");
    for (step, name) in [
        ("generation", "AFTER_PROFILE"),
        ("hooks", "AFTER_SETTINGS"),
        ("go-state", "AFTER_GO_SETTINGS"),
    ] {
        let sentence = format!("`{step}` reads `{}`", constant(name));
        let in_text = text.contains(&sentence)
            || (step == "generation"
                && text.contains(&format!(
                    "`generation` reads `blocked` with `{}`",
                    constant(name)
                )));
        assert!(in_text, "the workflows do not say: {sentence}");
        assert!(
            source.contains(&format!("Step::waiting(\"{step}\", \"blocked\", {name})")),
            "the code does not hold `{step}` with {name}"
        );
    }
}
