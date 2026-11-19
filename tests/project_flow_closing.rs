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
        "at the end of the root `.gitignore`",
        "after the rule",
        "outside any block MoAI manages",
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
        "`git status --porcelain=v1 -z --untracked-files=all`",
        "one entry",
        "`go-settings-apply`",
        "`go-state-apply`",
        "`project-apply`",
        "the `generation` apply",
        "`claude-settings-apply`",
        "`worktree-move-apply`",
        "not from this route",
        "kept local",
        "`.moai-backup/`",
        "`.moai/backups/`",
        "`.moai/receipts/`",
        "`.before/`",
        "`*.pem`",
        "`*secret*`",
        "anywhere in it",
        "project's language setting",
        "after the offer of the Completion Check",
        "fails (not a repository, or git cannot answer)",
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

/// The paragraph that starts with `Closing-commit snapshots` (under `## Procedure`, before
/// step 1): where the status lists are read and how each path is given to a step.
fn snapshot_paragraph() -> String {
    let text = skill_text("workflows/project.md");
    let start = position(&text, "Closing-commit snapshots (card C-59)");
    let tail = &text[start..];
    let end = tail.find("\n\n").unwrap_or(tail.len());
    assert!(
        position(&text, "## Procedure") < start && start < position(&text, "1. **Detect.**"),
        "the snapshot paragraph sits between `## Procedure` and step 1"
    );
    tail[..end].to_string()
}

/// REQ-13 (audit B2): the apply results name no paths (`project-apply` names none), so the
/// text gives each path to the step whose snapshot first showed it.
#[test]
fn procedure_snapshots_the_status_around_every_apply() {
    let para = snapshot_paragraph();
    for text in [
        "`git status --porcelain=v1 -z --untracked-files=all`",
        "before the first apply",
        "After each apply",
        "the \"Track it\" edit",
        "new or whose status changed",
        "name none",
        "already changed before the route began",
        "never given to an apply",
    ] {
        assert!(para.contains(text), "{text:?} is missing from: {para}");
    }
}

/// The name rule the text states for step 8: a secret-looking name anywhere in the path.
fn looks_secret(path: &str) -> bool {
    path.split('/').any(|part| {
        part.starts_with(".env")
            || part.ends_with(".key")
            || part.ends_with(".pem")
            || part.contains("credentials")
            || part.contains("secret")
    })
}

/// The backup and receipt paths the text keeps local.
fn kept_local(path: &str) -> bool {
    [".moai-backup/", ".moai/backups/", ".moai/receipts/"]
        .iter()
        .any(|p| path.starts_with(p))
        || path.split('/').any(|part| part == ".before")
}

/// The sorted `-z` entries (`XY path`) of one status read.
fn status_entries(dir: &Path, untracked_all: bool) -> Vec<String> {
    let mut args = vec!["status", "--porcelain=v1", "-z"];
    if untracked_all {
        args.push("--untracked-files=all");
    }
    let out = git_raw(dir, &args);
    assert!(out.success, "{}", out.stderr);
    let mut entries: Vec<String> = out
        .stdout
        .split('\0')
        .filter(|e| !e.is_empty())
        .map(str::to_string)
        .collect();
    entries.sort_unstable();
    entries
}

fn fresh_repo_with_one_commit() -> TempDir {
    let tmp = TempDir::new().unwrap();
    git_ok(tmp.path(), &["init", "-q"]);
    fs::write(tmp.path().join("README.md"), "project\n").unwrap();
    commit_all(tmp.path());
    tmp
}

fn write_file(dir: &Path, rel: &str, body: &str) {
    let path = dir.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

/// REQ-5, REQ-8, REQ-14 (audit B1): on a fresh project git shows an untracked folder as one
/// entry, so a name rule cannot see a `credentials` file inside it; with
/// `--untracked-files=all` every file is listed, the name rules apply to each, and staging only
/// the paths that pass leaves the secret, the backup and the receipt out.
#[test]
fn untracked_folders_collapse_unless_every_file_is_asked_for() {
    let fx = fresh_repo_with_one_commit();
    let dir = fx.path();
    write_file(dir, ".moai/project.yaml", "schema_version: 2\n");
    write_file(dir, ".moai/receipts/r.json", "{}\n");
    write_file(dir, ".moai-backup/config/credentials.yaml", "token: x\n");
    write_file(dir, ".moai-backup/config/a.yaml", "a: 1\n");

    assert_eq!(
        status_entries(dir, false),
        ["?? .moai-backup/", "?? .moai/"],
        "without the flag the folders are one entry each"
    );
    let all = status_entries(dir, true);
    assert_eq!(
        all,
        [
            "?? .moai-backup/config/a.yaml",
            "?? .moai-backup/config/credentials.yaml",
            "?? .moai/project.yaml",
            "?? .moai/receipts/r.json"
        ],
        "with the flag every file is listed"
    );

    let paths: Vec<&str> = all.iter().map(|e| &e[3..]).collect();
    let secret: Vec<&&str> = paths.iter().filter(|p| looks_secret(p)).collect();
    assert_eq!(secret, [&".moai-backup/config/credentials.yaml"]);
    let stageable: Vec<&&str> = paths
        .iter()
        .filter(|p| !looks_secret(p) && !kept_local(p))
        .collect();
    assert_eq!(stageable, [&".moai/project.yaml"]);

    for path in stageable {
        git_ok(dir, &["add", "--", path]);
    }
    assert_eq!(
        git(dir, &["diff", "--cached", "--name-only"]).trim_end(),
        ".moai/project.yaml",
        "only the profile is staged"
    );
}

/// REQ-13 (audit B2, N2): the paths each step wrote are the entries that are new in the status
/// list read after it, so a result that names no path still gives its files to the right step,
/// the user's earlier changes stay out, and the "Track it" edit of `.gitignore` is a step too.
#[test]
fn snapshot_differences_give_each_new_path_to_the_step_that_wrote_it() {
    let fx = fresh_repo_with_one_commit();
    let dir = fx.path();
    // The user's own changes before the route began.
    write_file(dir, "notes.txt", "mine\n");
    write_file(dir, "README.md", "project, edited\n");
    let s0 = status_entries(dir, true);

    let new_since = |before: &[String], after: &[String]| -> Vec<String> {
        after
            .iter()
            .filter(|e| !before.contains(e))
            .cloned()
            .collect()
    };

    // `project-apply` writes the profile and its sidecars and names none of them.
    write_file(dir, ".moai/project.yaml", "schema_version: 2\n");
    write_file(dir, ".moai/manifests/profile-baseline.json", "{}\n");
    let s1 = status_entries(dir, true);
    assert_eq!(
        new_since(&s0, &s1),
        [
            "?? .moai/manifests/profile-baseline.json",
            "?? .moai/project.yaml"
        ]
    );

    // The `generation` apply writes AGENTS.md; the user's `notes.txt` is not given to it.
    write_file(dir, "AGENTS.md", "managed\n");
    let s2 = status_entries(dir, true);
    assert_eq!(new_since(&s1, &s2), ["?? AGENTS.md"]);

    // The "Track it" edit appends one line to the root `.gitignore`.
    write_file(dir, ".gitignore", "*.yaml\n!.moai/project.yaml\n");
    let s3 = status_entries(dir, true);
    assert_eq!(new_since(&s2, &s3), ["?? .gitignore"]);

    // What was already changed before the route is in the first list, never given to a step.
    assert!(
        s0.iter().any(|e| e == " M README.md") && s0.iter().any(|e| e == "?? notes.txt"),
        "{s0:?}"
    );
    for (name, list) in [("s1", &s1), ("s2", &s2), ("s3", &s3)] {
        assert!(
            s0.iter().all(|e| list.contains(e)),
            "{name} still holds the earlier changes"
        );
    }
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
