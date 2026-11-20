//! SPEC-20261120-closing-commit-branch: under the m-flow work flow, the closing commit of the
//! PROJECT route is never made on the integration branch. Two tests read the workflow text
//! (step 8 of `skills/moai/workflows/project.md`); one runs the git commands that text names in
//! a temporary repository (a temporary folder, never the real HOME, with the fixture git guards
//! of `tests/common/git.rs`), so the text only says what git really does.

mod common;
use common::git::{command, git, git_ok};
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const INTEGRATION: &str = "main";
const NEW_BRANCH: &str = "spec/project-setup-20261120";

fn closing_commit_step() -> String {
    let text = std::fs::read_to_string(format!(
        "{}/skills/moai/workflows/project.md",
        env!("CARGO_MANIFEST_DIR")
    ))
    .expect("project.md");
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("8. **Offer the closing commit") {
            inside = true;
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
        "project.md has no `8. **Offer the closing commit` step"
    );
    out
}

fn assert_has(step: &str, phrases: &[&str]) {
    for p in phrases {
        assert!(step.contains(p), "{p:?} is missing from: {step}");
    }
}

/// AC-1: the step says when the rule applies (m-flow, the current branch is the profile's
/// integration branch), what it offers instead, that nothing is committed on that branch, what
/// stays as it was in every other case, and that an unanswered question commits nothing.
#[test]
fn closing_commit_on_the_integration_branch_offers_a_new_branch() {
    let step = closing_commit_step();
    assert_has(
        &step,
        &[
            "`work_flow.flow` is `m-flow`",
            "`git branch --show-current`",
            "`work_flow.integration_branch`",
            "never commit on that branch",
            "Commit on a new branch (recommended)",
            "Leave them uncommitted",
            "Any other flow, any other branch and a profile without an integration branch keep the selection below",
            "cancelled, empty or unanswered",
        ],
    );
    let new_branch = step.find("Commit on a new branch (recommended)").unwrap();
    let leave = step.find("or \"Leave them uncommitted\"").unwrap();
    assert!(
        new_branch < leave,
        "the new branch is the recommended first option: {step}"
    );
}

/// AC-2: the new branch is named by the SPEC branch rule without a card number, is made from the
/// current commit with `git switch -c`, and a failure shows git's message and commits nothing.
#[test]
fn the_new_branch_is_named_by_the_spec_rule_and_failure_commits_nothing() {
    let step = closing_commit_step();
    assert_has(
        &step,
        &[
            "`spec/<slug>-<yyyymmdd>`",
            "no card number",
            "`git switch -c <name>`",
            "from the current commit",
            "when `git switch -c` fails, show git's message and commit nothing",
        ],
    );
}

fn commit_env(dir: &Path, args: &[&str]) -> (i32, String) {
    let mut full = vec![
        "-c",
        "user.name=fixture",
        "-c",
        "user.email=fixture@example.invalid",
    ];
    full.extend_from_slice(args);
    let out = command()
        .args(&full)
        .current_dir(dir)
        .stdin(std::process::Stdio::null())
        .output()
        .expect("spawn git");
    (
        out.status.code().expect("git exited by signal"),
        String::from_utf8_lossy(&out.stderr).to_string(),
    )
}

fn repo_on_integration_branch() -> TempDir {
    let tmp = TempDir::new().unwrap();
    git_ok(tmp.path(), &["init", "-q", "-b", INTEGRATION]);
    fs::write(tmp.path().join("README.md"), "first\n").unwrap();
    git_ok(tmp.path(), &["add", "--", "README.md"]);
    let (code, err) = commit_env(tmp.path(), &["commit", "-q", "-m", "first"]);
    assert_eq!(code, 0, "{err}");
    tmp
}

/// AC-3: the commands the step names, run on the integration branch, leave that branch where it
/// was and put the commit on the new branch only; when the branch name is taken, `git switch -c`
/// fails, the working tree and the integration branch stay as they were and nothing is committed.
#[test]
fn the_commands_leave_the_integration_branch_untouched() {
    let tmp = repo_on_integration_branch();
    let dir = tmp.path();
    let before = git(dir, &["rev-parse", INTEGRATION]);
    assert_eq!(
        git(dir, &["branch", "--show-current"]).trim(),
        INTEGRATION,
        "the fixture starts on the integration branch"
    );
    fs::create_dir_all(dir.join(".moai")).unwrap();
    fs::write(dir.join(".moai/project.yaml"), "schema_version: 2\n").unwrap();

    git_ok(dir, &["switch", "-c", NEW_BRANCH]);
    git_ok(dir, &["add", "--", ".moai/project.yaml"]);
    let (code, err) = commit_env(
        dir,
        &[
            "commit",
            "-q",
            "-m",
            "project setup",
            "--",
            ".moai/project.yaml",
        ],
    );
    assert_eq!(code, 0, "{err}");

    assert_eq!(
        git(dir, &["rev-parse", INTEGRATION]),
        before,
        "the integration branch did not move"
    );
    assert_eq!(git(dir, &["branch", "--show-current"]).trim(), NEW_BRANCH);
    assert_eq!(
        git(
            dir,
            &[
                "rev-list",
                "--count",
                &format!("{INTEGRATION}..{NEW_BRANCH}")
            ]
        )
        .trim(),
        "1",
        "the new branch holds exactly the one commit"
    );

    // A second run on the integration branch with the name already taken: the switch fails.
    git_ok(dir, &["switch", "-q", INTEGRATION]);
    fs::write(dir.join("notes.txt"), "second\n").unwrap();
    let out = command()
        .args(["switch", "-c", NEW_BRANCH])
        .current_dir(dir)
        .stdin(std::process::Stdio::null())
        .output()
        .expect("spawn git");
    assert!(!out.status.success(), "the taken name must fail");
    assert!(
        !String::from_utf8_lossy(&out.stderr).trim().is_empty(),
        "git says why, and the step shows that message"
    );
    assert_eq!(git(dir, &["branch", "--show-current"]).trim(), INTEGRATION);
    assert_eq!(git(dir, &["rev-parse", INTEGRATION]), before);
    assert!(
        dir.join("notes.txt").exists(),
        "the file stays in the working tree"
    );
}
