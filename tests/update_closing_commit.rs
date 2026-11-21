//! SPEC-20261121-update-closing-commit: `workflows/update.md` ends its procedure with the closing
//! commit offer of the project route, says what an uncommitted move does, reads the three things a
//! Go-generation update leaves, and names the commit in its completion check. The test reads the
//! workflow by path and checks the named words; it never walks the repository.

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
    // After the re-run and after taking it back, before the exceptions and the completion check.
    assert!(at("5. **Re-run and report.**") < at("7. **Offer the closing commit.**"));
    assert!(at("6. **Taking it back.**") < at("7. **Offer the closing commit.**"));
    assert!(at("7. **Offer the closing commit.**") < at("## Exceptions and Permissions"));
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
            "`go-settings-apply`, `go-state-apply`, `project-apply`, the `generation` apply, `claude-settings-apply`, `worktree-move-apply`, `hooks-apply`",
            "a path `go-state-apply` moved or deleted is listed one by one, not as a count",
            "Mark every listed path that `git ls-files` shows as tracked",
            "includes changes from before this route",
            "a path no step gave goes under \"not from this route\"",
            "`.moai-backup/`, `.moai/backups/`, `.moai/receipts/`, `.moai-cli/backups/`, `.moai-cli/receipts/`, `.moai.rust-*` or a `.before/` folder goes under \"kept local\" and is never staged",
            // The warning the report asks for: an uncommitted move of a tracked file comes back.
            "A move that is not committed comes back: a tracked file that a step moved returns at the next `git checkout`, `git stash pop` or `git switch`, and in every worktree made from that branch",
            "say this before the question",
            "\"Commit these changes (recommended)\" or \"Leave them uncommitted\"",
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
            "When the plan has a `claude-md` step",
            "replace it with the managed `CLAUDE.md` (recommended), import only, or leave it",
            "the Go text goes to the backup",
            "Read `leftovers` as folders and files of the earlier generation that no step uses",
            "`.moai-cli/`, `.moai.rust-*`, `.moai-backups/`, a dead `.update.lock`",
            "report an unanswered leftover as what is left, never as done",
        ],
    );
}
