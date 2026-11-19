//! SPEC-20261119-project-flow-closing (card C-59, criteria 4 and 5): the PROJECT route reports
//! an ignored profile as a "Track it / Keep it local" selection and ends with a closing commit
//! selection. Two tests read the workflow text; two run the git commands that text names in
//! a temporary repository (a temporary folder, never the real HOME, with the fixture git
//! guards of `tests/common/git.rs`), so the text only says what git really does.

mod common;
use common::git::{command, git, git_ok, git_raw};
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const PROFILE: &str = ".moai/project.yaml";

fn skill_text(rel: &str) -> String {
    std::fs::read_to_string(format!("{}/skills/moai/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// The workflow bullet that starts with `- Profile tracking` (three spaces of indent), up to the
/// next bullet or the next unindented line.
fn profile_tracking_step() -> String {
    let text = skill_text("workflows/project.md");
    let mut out = String::new();
    let mut inside = false;
    for line in text.lines() {
        if line.starts_with("   - Profile tracking") {
            inside = true;
        } else if inside && (line.starts_with("   - ") || !line.starts_with(' ')) {
            break;
        }
        if inside {
            out.push_str(line);
            out.push('\n');
        }
    }
    assert!(
        !out.is_empty(),
        "project.md has no `- Profile tracking` step"
    );
    out
}

/// The numbered step `8. **Offer the closing commit` up to the next `## ` heading.
fn closing_commit_step() -> String {
    let text = skill_text("workflows/project.md");
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

fn position(text: &str, needle: &str) -> usize {
    text.find(needle)
        .unwrap_or_else(|| panic!("{needle:?} is missing from: {text}"))
}

/// Exit code and stderr of one fixture git call (`git_raw` only says success or not, and the
/// workflow text depends on 0 against 1 against an error).
fn code(dir: &Path, args: &[&str]) -> (i32, String, String) {
    let out = command()
        .args(args)
        .current_dir(dir)
        .stdin(std::process::Stdio::null())
        .output()
        .expect("spawn git");
    (
        out.status.code().expect("git exited by signal"),
        String::from_utf8_lossy(&out.stdout).to_string(),
        String::from_utf8_lossy(&out.stderr).to_string(),
    )
}

fn repo_with_profile() -> TempDir {
    let tmp = TempDir::new().unwrap();
    git_ok(tmp.path(), &["init", "-q"]);
    fs::create_dir_all(tmp.path().join(".moai")).unwrap();
    fs::write(tmp.path().join(PROFILE), "schema_version: 2\n").unwrap();
    tmp
}

fn commit_all(dir: &Path) {
    git_ok(dir, &["add", "-A"]);
    git_ok(
        dir,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "commit",
            "-q",
            "-m",
            "fixture",
        ],
    );
}

/// AC-1 (card criterion 4): the step asks git whether the profile is tracked, then whether it is
/// ignored, names the rule, offers Track it (recommended) before Keep it local, adds one
/// reversible line, asks again, and says what a git error, a cancelled question and a rule
/// that still ignores the profile do.
#[test]
fn profile_tracking_step_states_the_checks_the_selection_and_the_undo() {
    let step = profile_tracking_step();
    for text in [
        "`git ls-files --error-unmatch",
        "`git check-ignore -q",
        "`git check-ignore -v",
        "Exit 0 means tracked and nothing more is asked",
        "Track it (recommended)",
        "Keep it local",
        "`!<profile path>`",
        "root `.gitignore`",
        "after the rule",
        "removes exactly that line",
        "a fresh clone will not have the profile",
        "cancelled, empty or unanswered",
        "neither 0 nor 1",
    ] {
        assert!(step.contains(text), "{text:?} is missing from: {step}");
    }
    assert!(
        position(&step, "`git ls-files --error-unmatch") < position(&step, "`git check-ignore -q")
            && position(&step, "`git check-ignore -q") < position(&step, "`git check-ignore -v")
            && position(&step, "Track it (recommended)") < position(&step, "Keep it local"),
        "order of the checks and the options: {step}"
    );
    // The global excludes file is the user's own setting (`guarded`): the step never edits it.
    assert!(
        step.contains("never edit the global excludes file"),
        "global excludes are left alone: {step}"
    );
}

/// AC-3 (card criterion 5): the last step lists the changed paths under the apply that named
/// them, offers the commit (recommended) before leaving them, stages only the listed paths of
/// this route, never pushes, and says what a secret name, an unanswered question and a failing
/// `git add` or `git commit` do.
#[test]
fn closing_commit_step_lists_by_apply_stages_only_those_paths_and_never_pushes() {
    let step = closing_commit_step();
    for text in [
        "`git status --porcelain=v1 -z`",
        "`go-settings-apply`",
        "`go-state-apply`",
        "`project-apply`",
        "the `generation` apply",
        "`claude-settings-apply`",
        "`worktree-move-apply`",
        "not from this route",
        "Commit these changes (recommended)",
        "Leave them uncommitted",
        "`git add -- <path>`",
        "never `-A`, `.`",
        "does not push",
        "`git reset --soft HEAD~1`",
        "`.env*`",
        "`*.key`",
        "`*credentials*`",
        "cancelled, empty or unanswered",
        "name every path already staged",
    ] {
        assert!(step.contains(text), "{text:?} is missing from: {step}");
    }
    assert!(
        position(&step, "Commit these changes (recommended)")
            < position(&step, "Leave them uncommitted"),
        "the commit is the recommended first option: {step}"
    );
    assert!(
        !step.contains("`git push"),
        "the step never tells the session to push: {step}"
    );
    // It runs after the offer of the code map (step 7) and before the headings that follow.
    let text = skill_text("workflows/project.md");
    assert!(
        position(&text, "7. **Offer the code map.**")
            < position(&text, "8. **Offer the closing commit"),
        "step 8 follows step 7"
    );
}

/// AC-2 (REQ-1, REQ-3, REQ-10): the exit codes and outputs the text relies on, in real git.
#[test]
fn the_git_commands_the_text_names_behave_as_the_text_says() {
    // Untracked and not ignored: ls-files says no match (1), check-ignore -q says not ignored
    // (1, nothing on stderr).
    let fx = repo_with_profile();
    let dir = fx.path();
    let (c, _, err) = code(dir, &["ls-files", "--error-unmatch", "--", PROFILE]);
    assert_eq!(c, 1, "untracked: ls-files --error-unmatch exits 1: {err}");
    let (c, _, err) = code(dir, &["check-ignore", "-q", "--", PROFILE]);
    assert_eq!((c, err.as_str()), (1, ""), "untracked, not ignored");

    // Tracked: ls-files exits 0.
    commit_all(dir);
    let (c, _, err) = code(dir, &["ls-files", "--error-unmatch", "--", PROFILE]);
    assert_eq!(c, 0, "tracked: ls-files --error-unmatch exits 0: {err}");

    // A user rule `*.yaml` ignores the untracked profile; `-v` names the rule and the file.
    let fx = repo_with_profile();
    let dir = fx.path();
    fs::write(dir.join(".gitignore"), "*.yaml\n").unwrap();
    let (c, _, _) = code(dir, &["check-ignore", "-q", "--", PROFILE]);
    assert_eq!(c, 0, "ignored by `*.yaml`");
    let (c, out, _) = code(dir, &["check-ignore", "-v", "--", PROFILE]);
    assert_eq!(c, 0);
    assert_eq!(
        out.trim_end(),
        format!(".gitignore:1:*.yaml\t{PROFILE}"),
        "-v names the file, the line, the rule and the path"
    );

    // `/.moai/*` ignores the contents but not the folder, so one `!<profile path>` line after
    // the rule re-includes the profile, and git answers "not ignored" again.
    let fx = repo_with_profile();
    let dir = fx.path();
    fs::write(dir.join(".gitignore"), "/.moai/*\n").unwrap();
    assert_eq!(code(dir, &["check-ignore", "-q", "--", PROFILE]).0, 0);
    fs::write(dir.join(".gitignore"), format!("/.moai/*\n!{PROFILE}\n")).unwrap();
    let (c, _, err) = code(dir, &["check-ignore", "-q", "--", PROFILE]);
    assert_eq!((c, err.as_str()), (1, ""), "the `!` line re-includes it");

    // A rule on the whole folder cannot be undone by a `!` line: git still ignores the
    // profile, which is why the text asks again and removes exactly the line it added.
    let fx = repo_with_profile();
    let dir = fx.path();
    fs::write(dir.join(".gitignore"), format!("/.moai/\n!{PROFILE}\n")).unwrap();
    assert_eq!(
        code(dir, &["check-ignore", "-q", "--", PROFILE]).0,
        0,
        "a folder rule still ignores the profile after the `!` line"
    );
    fs::write(dir.join(".gitignore"), "/.moai/\n").unwrap();
    assert_eq!(code(dir, &["check-ignore", "-q", "--", PROFILE]).0, 0);

    // An error is neither 0 nor 1: outside a repository both commands exit 128 with a message.
    let outside = TempDir::new().unwrap();
    let (c, _, err) = code(outside.path(), &["check-ignore", "-q", "--", PROFILE]);
    assert!(c != 0 && c != 1, "check-ignore outside a repo exits {c}");
    assert!(!err.is_empty(), "and says why");
    let (c, _, err) = code(
        outside.path(),
        &["ls-files", "--error-unmatch", "--", PROFILE],
    );
    assert!(c != 0 && c != 1, "ls-files outside a repo exits {c}");
    assert!(!err.is_empty(), "and says why");
}

/// AC-4 (REQ-5, REQ-7): `git status --porcelain=v1 -z` lists a deleted, a new and a user's own
/// modified path; `git add -- <path>` for exactly the two paths of the route stages the
/// deletion and the new file and leaves the user's modification unstaged.
#[test]
fn staging_listed_paths_stages_deletions_and_leaves_other_changes_alone() {
    let fx = TempDir::new().unwrap();
    let dir = fx.path();
    git_ok(dir, &["init", "-q"]);
    fs::create_dir_all(dir.join(".moai")).unwrap();
    fs::create_dir_all(dir.join("keep")).unwrap();
    fs::write(dir.join(".moai/old.json"), "{}\n").unwrap();
    fs::write(dir.join("keep/user.txt"), "mine\n").unwrap();
    commit_all(dir);

    // What a route leaves behind: a moved-away file, a new profile, and the user's own edit.
    fs::remove_file(dir.join(".moai/old.json")).unwrap();
    fs::write(dir.join(PROFILE), "schema_version: 2\n").unwrap();
    fs::write(dir.join("keep/user.txt"), "mine, edited\n").unwrap();

    let status = git_raw(dir, &["status", "--porcelain=v1", "-z"]);
    assert!(status.success, "{}", status.stderr);
    let mut entries: Vec<&str> = status
        .stdout
        .split('\0')
        .filter(|e| !e.is_empty())
        .collect();
    entries.sort_unstable();
    assert_eq!(
        entries,
        [
            " D .moai/old.json",
            " M keep/user.txt",
            "?? .moai/project.yaml"
        ],
        "the -z list the text reads"
    );

    // Stage exactly the two paths a route step named; never `-A` or `.`.
    for path in [".moai/old.json", PROFILE] {
        git_ok(dir, &["add", "--", path]);
    }
    let staged = git(dir, &["diff", "--cached", "--name-status"]);
    let mut lines: Vec<&str> = staged.lines().collect();
    lines.sort_unstable();
    assert_eq!(lines, ["A\t.moai/project.yaml", "D\t.moai/old.json"]);
    let after = git(dir, &["status", "--porcelain=v1"]);
    assert!(
        after.lines().any(|l| l == " M keep/user.txt"),
        "the user's own edit stays unstaged:\n{after}"
    );
}
