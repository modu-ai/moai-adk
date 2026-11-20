//! SPEC-20261119-project-worktree-move (card C-59, criterion 3): the UPDATE workflow reports a
//! git worktree registered under the Go `.moai/` folder with the worktree-move choice. The
//! tests read the workflow text and the sources that print the reason, the plan fields and the
//! apply arguments, so the text only quotes what the code really says. No host, no HOME, no
//! git and no command runs here.

use std::fs;

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The section "Earlier-generation state" of `update.md` (the move of the earlier generation's
/// state) up to the next `## ` heading. The route that moves it owns the worktree move too.
fn step_2() -> String {
    let text = read("skills/moai/workflows/update.md");
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("## Earlier-generation state") {
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
    assert!(
        !out.is_empty(),
        "update.md has no section Earlier-generation state"
    );
    out
}

/// Where `needle` first stands in `text`; a missing sentence fails with its own words.
fn at(text: &str, needle: &str) -> usize {
    text.find(needle)
        .unwrap_or_else(|| panic!("the earlier-generation section of update.md lacks: {needle}"))
}

#[test]
fn go_state_worktree_reason_is_reported_with_the_worktree_move_choice() {
    let step = step_2();
    for phrase in [
        "git worktree(s) are registered under .moai/",
        "`folder_stays`",
        "git would lose those worktrees if the folder moved",
        "Report that reason to the user in the plan's own words",
        "name the worktree move as the way to let the folder move",
        "`worktree-move-plan`",
    ] {
        assert!(
            step.contains(phrase),
            "the earlier-generation section of update.md lacks: {phrase}"
        );
    }
}

#[test]
fn worktree_move_is_previewed_then_asked_once_then_applied_only_after_the_pick() {
    let step = step_2();
    let preview = at(
        &step,
        "Run the read-only `moai worktree-move-plan <project>`",
    );
    let ask = at(&step, "ask one selection after the preview");
    let apply = at(&step, "Run `moai worktree-move-apply <project>");
    assert!(
        preview < ask && ask < apply,
        "the order must be preview, one ask, apply"
    );
    let first = at(&step, "\"Move the worktrees (recommended)\"");
    let second = at(&step, "\"Leave them where they are\"");
    assert!(first < second, "the move comes first in the selection");
    for phrase in [
        "show its `items`",
        "its `totals`",
        // Tied to the worktree-move-apply line: `go-state-apply` says the same words nearby.
        "as the preview only after the user picked the move",
        "Without a pick nothing moves",
        "without `--force`",
        "a changed project refuses the digest",
        // The hold check: no observation, no offer, and no invented empty list.
        "`hold_check_not_run`",
        "never as an empty list you did not observe",
        "report the worktree move as unavailable with the reason `hold_check_not_run` and ask nothing about it",
        // The folder moves only on a new preview that reads it as `move`.
        "run `moai go-state-plan <project>` again",
    ] {
        assert!(step.contains(phrase), "the earlier-generation section of update.md lacks: {phrase}");
    }
}

#[test]
fn quoted_names_are_the_ones_the_code_prints() {
    let go_state = read("src/go_state.rs");
    for phrase in [
        "git worktree(s) are registered under .moai/",
        "worktree-move-plan",
    ] {
        assert!(go_state.contains(phrase), "src/go_state.rs lacks: {phrase}");
    }
    assert!(
        go_state.contains("folder_stays"),
        "src/go_state.rs prints no `folder_stays` layout"
    );
    let plan = read("src/worktree_move.rs");
    for name in [
        "hold_check_not_run",
        "\"totals\"",
        "\"plan_digest\"",
        "\"items\"",
        "\"reasons\"",
        "\"destination\"",
        "\"root\"",
    ] {
        assert!(plan.contains(name), "src/worktree_move.rs lacks: {name}");
    }
    // The move is `git worktree move`, never forced.
    assert!(
        !plan.contains("\"--force\""),
        "src/worktree_move.rs passes --force to git"
    );
    let apply = read("src/ops/worktree_move_apply.rs");
    for field in [
        "plan_digest: String",
        "permission_reference: String",
        "root:",
        "observations:",
    ] {
        assert!(
            apply.contains(field),
            "src/ops/worktree_move_apply.rs lacks the argument `{field}`"
        );
    }
    let step = step_2();
    for argument in [
        "--plan-digest",
        "--permission-reference",
        "--root",
        "--observations",
    ] {
        assert!(
            step.contains(argument),
            "the earlier-generation section of update.md lacks {argument}"
        );
    }
}
