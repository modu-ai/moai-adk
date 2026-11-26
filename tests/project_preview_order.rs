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

/// The update workflow: it owns the backup and merge steps (the project workflow offers the install
/// steps only and names `/moai update` for the rest), so their waiting reasons are quoted there.
fn update_md() -> String {
    read("skills/moai/workflows/update.md")
}

/// The numbered step `5. **Refresh the managed section` up to the next numbered step.
fn step_5() -> String {
    let text = project_md();
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("5. **Refresh the managed section") {
            inside = true;
        } else if inside && line.starts_with("6. ") {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "project.md has no step 5");
    out
}

/// The `## Completion Check` section up to the next `## ` heading.
fn completion_check() -> String {
    let text = project_md();
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("## Completion Check") {
            inside = true;
            continue;
        } else if inside && line.starts_with("## ") {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(!out.is_empty(), "project.md has no Completion Check");
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

const WAITING: [&str; 6] = [
    "AFTER_BACKUP",
    "AFTER_PROFILE",
    "AFTER_SETTINGS",
    "AFTER_GO_SETTINGS",
    "AFTER_INSTALL",
    // The reports step waits for its own apply; `update.md` quotes it.
    "AFTER_REPORTS",
];

#[test]
fn step_5_asks_about_generation_only_when_it_reads_ready() {
    let step = step_5();
    for phrase in [
        "Ask about the `generation` step only when it reads `ready` in that run.",
        "Ask nothing about a step that reads `blocked` or `re-run after …`",
        "apply the earlier step its reason names, run `moai project-update-plan <project>` again and ask from that new run",
        "because an earlier yes does not cover a step the user has not seen",
        "an earlier step the user declined or left out stays unapplied, and the step behind it is reported with the reason the plan prints, not asked",
    ] {
        assert!(step.contains(phrase), "step 5 lacks: {phrase}");
    }
    // The text speaks of the step in which `generation` is asked, before the apply it names.
    let ready = step.find("only when it reads `ready`").unwrap();
    let apply = step.find("run its printed `apply.argv`").unwrap();
    assert!(ready < apply, "the `ready` rule must come before the apply");
}

#[test]
fn completion_check_offers_steps_in_the_printed_phase_order() {
    let check = completion_check();
    // The sentences that hold for the install steps stay in the project workflow's check.
    for phrase in [
        "a waiting step is reported with that reason and asked about only after the earlier step is applied and the plan is run again",
        "as UPDATE step 5 says",
    ] {
        assert!(check.contains(phrase), "the Completion Check lacks: {phrase}");
    }
    // The offer of the project workflow is the install steps only (REQ-28): the backup and merge
    // steps are named to `/moai update`, in one sentence that names all three.
    let order =
        "phase `backup`, then `install`, then `merge`, each phase in the order of its steps";
    assert!(
        !check.contains(order),
        "the Completion Check still offers the backup and merge phases: {order}"
    );
    assert!(
        check
            .split(". ")
            .any(|s| s.contains("`/moai update`") && s.contains("backup") && s.contains("merge")),
        "the Completion Check must name `/moai update` for the backup and merge steps"
    );
    // The phase order and the merge steps' waiting reasons are the update workflow's text now.
    let update = update_md();
    for phrase in [
        order,
        "the merge steps read `re-run after the backup phase` while the backup phase still has a move to make and `re-run after the install phase` behind the install phase",
    ] {
        assert!(update.contains(phrase), "update.md lacks: {phrase}");
    }
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
    // The reasons are quoted by the two workflows together: the project workflow for the install
    // steps, the update workflow for the backup and merge steps it owns.
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
        // A step of the backup phase (`go-state`) waits behind an earlier step in the update
        // workflow's text only: the project workflow no longer names it.
        let home = if step == "go-state" {
            update_md()
        } else {
            text.clone()
        };
        let in_text = home.contains(&sentence)
            || ((step == "generation" || step == "hooks")
                && home.contains(&format!(
                    "`{step}` reads `blocked` with `{}`",
                    constant(name)
                )));
        assert!(in_text, "the workflows do not say: {sentence}");
        assert!(
            source.contains(&format!("Step::waiting(\"{step}\", \"blocked\", {name})")),
            "the code does not hold `{step}` with {name}"
        );
    }
}
