//! SPEC-20261125-update-closing-commit: `workflows/update.md` ends its procedure with the closing
//! commit offer of the project route (step 8, after the leftover-space step 7), says what an
//! uncommitted move does, reads the three things a Go-generation update leaves, and names the commit
//! in its completion check. The test reads the workflow by path and checks the named words; it never
//! walks the repository.

use std::fs;
use std::path::Path;

/// The file's text with every run of whitespace turned into one space, so a rewrapped line does
/// not break a check.
fn update() -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/update.md");
    let text = fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

fn has(text: &str, words: &[&str]) {
    for w in words {
        assert!(text.contains(w), "update.md is missing `{w}`");
    }
}

fn lacks(text: &str, words: &[&str]) {
    for w in words {
        assert!(!text.contains(w), "update.md still has `{w}`");
    }
}

#[test]
fn closing_commit_is_the_last_step() {
    let t = update();
    let at = |needle: &str| {
        t.find(needle)
            .unwrap_or_else(|| panic!("update.md is missing `{needle}`"))
    };
    // After the re-run, after taking it back and after the leftover-space step, before the
    // exceptions and the completion check.
    assert!(at("5. **Re-run and report.**") < at("8. **Offer the closing commit.**"));
    assert!(at("6. **Taking it back.**") < at("8. **Offer the closing commit.**"));
    assert!(at("7. **Leftover space") < at("8. **Offer the closing commit.**"));
    assert!(at("8. **Offer the closing commit.**") < at("## Exceptions and Permissions"));
    has(
        &t,
        &[
            "Run it after the re-run of step 5 has been reported on and before the final report",
            "so the commit holds what the steps wrote",
            "The closing commit is reported as committed (with its id and the way back), left uncommitted by the user's pick, or not offered with its reason.",
        ],
    );
    // The completion check is the last section.
    assert!(
        at("The closing commit is reported as committed") > at("## Completion Check"),
        "the commit belongs to the completion check"
    );
}

#[test]
fn closing_commit_lists_marks_and_warns() {
    let t = update();
    has(
        &t,
        &[
            "Read `git status --porcelain=v1 -z --untracked-files=all` in the project folder",
            "List every changed path under the apply that moved or wrote it",
            "`go-settings-apply`, `go-state-apply`, `project-apply`, the `generation` apply, `claude-settings-apply`, `worktree-move-apply`, `hooks-apply`, `claude-md-apply`",
            "a path `go-state-apply` moved or deleted is listed one by one, not as a count",
            "and so is every tracked path `claude-md-apply` moved (its `tracked_moved`, which git shows as deleted)",
            "Mark every listed path that `git ls-files` shows as tracked",
            "includes changes from before this route",
            "a path no step gave goes under \"not from this route\"",
            "`.moai-backup/`, `.moai/backups/`, `.moai/receipts/`, `.moai-cli/backups/`, `.moai-cli/receipts/`, `.moai.rust-*` or a `.before/` folder goes under \"kept local\" and is never staged",
            // The warning the report asks for: an uncommitted move of a tracked file comes back.
            "A move that is not committed comes back: a tracked file that a step moved returns at the next `git checkout`, `git stash pop` or `git switch`, and in every worktree made from that branch",
            "say this before the question",
            "\"Commit these changes (recommended)\" or \"Leave them uncommitted\"",
            "Before the selection, show the proposed commit (the listed paths) as a dry-run diff",
            "an existing commit is never amended",
            "read the index with `git diff --cached --name-only -z`",
            "\"already staged by you\"",
            "stage exactly the listed paths with `git add -- <path>`, one at a time",
            "never `-A`, `.` or a path under \"not from this route\" or \"kept local\"",
            "run `git commit -- <path>` with every listed path after the `--`",
            "`git reset --soft HEAD~1` as the way back",
            "the commit does not push",
            "`.env*`, `*.key`, `*.pem`, `*credentials*` or `*secret*`",
            "do not retry in a loop",
            "A cancelled, empty or unanswered question is \"Leave them uncommitted\" and writes nothing",
        ],
    );
}

#[test]
fn m_flow_commits_on_a_branch_of_its_own() {
    let t = update();
    has(
        &t,
        &[
            "On an m-flow project (`work_flow.flow` is `m-flow` in `project.yaml`) whose current branch is its `integration_branch`, never commit on that branch",
            "first `git switch -c spec/<slug>-<yyyymmdd>`",
            "no card number",
            "which keeps the work tree as it is",
            "or leave the changes uncommitted",
        ],
    );
    // `moai spec take` refuses a work tree that changed outside the SPEC's folder: not the way here.
    lacks(&t, &["moai spec take"]);
}

#[test]
fn go_generation_leftovers_are_read_out() {
    let t = update();
    has(
        &t,
        &[
            "**Reading what a Go-generation update left.**",
            "When the plan has a `go-residue` step",
            "`go_same`",
            "`go_changed`",
            "`go_marker`",
            "`user_data`",
            "each item shows `tracked` and `resurrected`",
            "a secret-looking name is shown only as a count",
            "The `claude-md` step has the next paragraph of its own",
            "Read `leftovers` as folders and files of the earlier generation that no step uses",
            "`.moai-cli/`, `.moai.rust-*`, `.moai-backups/`, a dead `.update.lock`",
            "report an unanswered leftover as what is left, never as done",
            "when the `claude-md` step asks about `CLAUDE.md`, report that file once, with that step",
        ],
    );
    // The product no longer offers an "import only" pick: the Reading paragraph must not name one.
    let start = t
        .find("**Reading what a Go-generation update left.**")
        .expect("update.md is missing the Reading paragraph");
    let end = t[start..]
        .find("**The picks")
        .map(|i| start + i)
        .expect("update.md is missing the `**The picks` paragraph after the Reading paragraph");
    lacks(&t[start..end], &["import only"]);
}

/// The paragraph about the `claude-md` step, up to the picks paragraph.
fn claude_md_paragraph(text: &str) -> &str {
    let start = text
        .find("**The Go-generation `CLAUDE.md` (the `claude-md` step).**")
        .expect("update.md is missing the claude-md paragraph");
    let end = text[start..]
        .find("**The picks")
        .expect("the claude-md paragraph stands before the picks paragraph");
    &text[start..start + end]
}

/// SPEC-20261126-claude-md-fresh-merge AC-8: the two steps of the route name the `claude-md`
/// step, first of the merge phase, with its two choices.
#[test]
fn claude_md_step_is_asked_first_in_the_merge_phase() {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/update.md");
    let raw = fs::read_to_string(path).unwrap();
    let step = |head: &str| {
        raw.lines()
            .find(|l| l.starts_with(head))
            .unwrap_or_else(|| panic!("update.md has no step {head}"))
    };
    has(
        step("1. **Preview everything once.**"),
        &["`merge` (first the `claude-md` step"],
    );
    has(
        step("2. **Ask.**"),
        &[
            "The `claude-md` step comes first of the merge phase",
            "`replace` (recommended) or `leave`",
            "the one yes for the `ready` steps never covers it",
        ],
    );
}

/// SPEC-20261126-claude-md-fresh-merge AC-8: how the step is read, both choices, the two
/// destinations, the preview with the picks before the apply, the decline, the restart notice,
/// the undo, the writers that keep away and the backup step the step waits behind.
#[test]
fn claude_md_step_is_read_as_the_commands_print_it() {
    let t = update();
    let p = claude_md_paragraph(&t);
    has(
        p,
        &[
            "reads `needs_decision` while the root `CLAUDE.md` is a Go-generation instruction file",
            "`unchanged` when it is none, `skip` once the user declined it",
            "into a backup folder and writes a new `CLAUDE.md`",
            "nothing is deleted",
            "`sources`",
            "`warnings`",
            "`preview.blocked`",
            "ask one selection, `replace` (recommended) or `leave`",
            "For `replace`",
            "`moai claude-md-plan <project>`",
            "`same_as`",
            "`hidden_lines`",
            "`moved_paths`",
            "`apply_blocked`",
            "`shared` (the new `CLAUDE.md`, shared through git",
            "`personal` (`CLAUDE.local.md`, ignored by git)",
            "a hunk nobody picks stays only in the backup",
            "without a `personal` pick no `CLAUDE.local.md` is written",
            "settings apply at once, while running sessions and lanes read changed instructions only after a restart",
            "preview again before any other merge step",
            "For `leave`",
            "`moai project-update-decline … --step claude-md",
            "writes only the decline record",
            "`claude-md-revert`",
            "`undo_order`",
            "`agents-picks` and `agents-local-merge` read `skip`",
            "no `append-managed`",
            "the optional `go-residue` step included, offer that step's choice first",
            "because the instruction step waits for it",
        ],
    );
    let at = |needle: &str| {
        p.find(needle)
            .unwrap_or_else(|| panic!("the claude-md paragraph is missing `{needle}`"))
    };
    // The preview with the picks comes between the first preview and the apply.
    let first = at("`moai claude-md-plan <project>`");
    let picks = at("`moai claude-md-plan <project> --keep <id>=shared --keep <id>=personal`");
    let yes = at("ask for the yes");
    let apply = at("run exactly the `apply.argv` that run printed");
    assert!(first < picks && picks < yes && yes < apply, "{p}");
}

/// The picks of the older files are not asked while the Go-generation file holds them.
#[test]
fn the_picks_paragraph_leaves_out_a_go_generation_claude_md() {
    has(
        &update(),
        &["is not yet in its CLAUDE file, and the root `CLAUDE.md` is no Go-generation file, the merge phase's `agents-picks` step reads `needs_decision`"],
    );
}
