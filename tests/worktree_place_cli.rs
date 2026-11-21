//! The worktree-place keystone (SPEC-20261122-worktree-place, slice: the path keystone and
//! the classifier): a worktree moai creates lands inside the project's `.claude/worktrees/`,
//! a worktree that Claude Code or Orca made there is accepted with any folder name, and the
//! old places keep their names in the classifier. Nothing here touches the real HOME.

use moai_cli::home_workdir;
use std::path::Path;

/// Scene 1 (REQ-1, REQ-3): the create path builds the project place for a moai-made
/// worktree, with the kept `<project>-<id>` name and no HOME argument.
#[test]
fn worktree_place_creation_lands_in_the_project() {
    let dir = home_workdir::worktree_dir(Path::new("/src/repo"), "1-k").unwrap();
    assert_eq!(
        dir,
        Path::new("/src/repo").join(".claude/worktrees/repo-1-k")
    );
    let dir = home_workdir::worktree_dir(Path::new("/src/My Repo (copy)"), "12-a b").unwrap();
    assert_eq!(
        dir,
        Path::new("/src/My Repo (copy)").join(".claude/worktrees/My-Repo-copy-12-a-b")
    );
}

/// Scene 2 (REQ-2, REQ-4): a path that Claude Code or Orca made under the project place is
/// accepted with whatever folder name it has, and a path outside is `other` to the classifier.
#[test]
fn worktree_place_host_made_is_accepted_and_outside_is_other() {
    let h = Path::new("/home/u");
    let p = Path::new("/src/repo");
    // Any folder name under the project place is a worktree of the project place.
    assert_eq!(
        home_workdir::location_of(Path::new("/src/repo/.claude/worktrees/whatever"), h, p),
        "claude"
    );
    assert_eq!(
        home_workdir::location_of(
            Path::new("/src/repo/.claude/worktrees/worktree-topic"),
            h,
            p
        ),
        "claude"
    );
    // Outside the project place stays outside.
    assert_eq!(
        home_workdir::location_of(Path::new("/elsewhere/wt"), h, p),
        "other"
    );
    assert_eq!(
        home_workdir::location_of(Path::new("/src/repo/.claude/worktrees/topic/sub"), h, p),
        "other"
    );
}

/// Scene 3 (REQ-7, REQ-8): the classifier keeps recognising the old places — the home
/// folder, the Go generation folder and the earlier project state folder.
#[test]
fn worktree_place_old_places_keep_their_names() {
    let h = Path::new("/home/u");
    let p = Path::new("/src/repo");
    let loc = |s: &str| home_workdir::location_of(Path::new(s), h, p);
    assert_eq!(loc("/home/u/.moai/worktrees/repo-1-k"), "home");
    assert_eq!(loc("/home/u/.moai/worktrees/projects/repo/feat"), "go");
    assert_eq!(loc("/src/repo/.moai-cli/worktrees/2811-x"), "project");
}

// ---- REQ-10, REQ-11: a creation keeps the main tree's status clean ----

mod common;
use common::git::{git, git_ok};

/// A repository with one commit; the repository's own `core.excludesFile` points nowhere, so
/// the runner's global ignore list cannot hide the folder this test is about.
fn repo(dir: &Path) {
    git_ok(dir, &["init", "-q", "-b", "main"]);
    git_ok(dir, &["config", "core.excludesFile", "/dev/null"]);
    git_ok(dir, &["config", "user.email", "t@example.invalid"]);
    git_ok(dir, &["config", "user.name", "t"]);
    std::fs::write(dir.join("README.md"), "x\n").unwrap();
    git_ok(dir, &["add", "README.md"]);
    git_ok(dir, &["commit", "-q", "-m", "init"]);
}

fn make_place_folder(dir: &Path) {
    std::fs::create_dir_all(dir.join(".claude/worktrees/x")).unwrap();
    std::fs::write(dir.join(".claude/worktrees/x/f.txt"), "f\n").unwrap();
}

fn exclude_lines(common_dir: &Path) -> usize {
    std::fs::read_to_string(common_dir.join("info/exclude"))
        .unwrap_or_default()
        .lines()
        .filter(|l| *l == "/.claude/worktrees/")
        .count()
}

/// Scene 4 (REQ-10): a repository that does not ignore the folder gets the one exclude line,
/// its status is clean afterwards, and a second call adds nothing.
#[test]
fn worktree_place_status_an_unignored_folder_gets_one_exclude_line() {
    let t = tempfile::tempdir().unwrap();
    let p = t.path();
    repo(p);
    make_place_folder(p);
    assert_eq!(git(p, &["status", "--porcelain"]), "?? .claude/");
    home_workdir::ensure_place_ignored(p).unwrap();
    assert_eq!(git(p, &["status", "--porcelain"]), "");
    home_workdir::ensure_place_ignored(p).unwrap();
    assert_eq!(exclude_lines(&p.join(".git")), 1);
}

/// Scene 5 (REQ-10): a folder that git already ignores through a tracked ignore file leaves
/// the exclude file exactly as it was.
#[test]
fn worktree_place_status_an_already_ignored_folder_writes_nothing() {
    let t = tempfile::tempdir().unwrap();
    let p = t.path();
    repo(p);
    std::fs::write(p.join(".gitignore"), ".claude/worktrees/\n").unwrap();
    git_ok(p, &["add", ".gitignore"]);
    git_ok(p, &["commit", "-q", "-m", "ignore"]);
    let before = std::fs::read(p.join(".git/info/exclude")).ok();
    home_workdir::ensure_place_ignored(p).unwrap();
    assert_eq!(std::fs::read(p.join(".git/info/exclude")).ok(), before);
    assert_eq!(exclude_lines(&p.join(".git")), 0);
}

/// Scene 6 (REQ-10): called from a linked worktree, the line goes to the shared common
/// folder, so the primary tree is clean too.
#[test]
fn worktree_place_status_a_linked_worktree_writes_the_common_folder() {
    let t = tempfile::tempdir().unwrap();
    let p = t.path().join("primary");
    std::fs::create_dir_all(&p).unwrap();
    repo(&p);
    let linked = t.path().join("linked");
    git_ok(
        &p,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "side",
            linked.to_str().unwrap(),
        ],
    );
    make_place_folder(&p);
    home_workdir::ensure_place_ignored(&linked).unwrap();
    assert_eq!(exclude_lines(&p.join(".git")), 1);
    assert!(
        linked.join(".git").is_file(),
        "the linked worktree keeps a .git file"
    );
    assert_eq!(git(&p, &["status", "--porcelain"]), "");
}

/// Scene 7 (REQ-11): an exclude file that is a link is not followed or written; the call
/// returns the error and the file the link points at keeps its bytes.
#[cfg(unix)]
#[test]
fn worktree_place_status_an_exclude_file_that_is_a_link_is_refused() {
    let t = tempfile::tempdir().unwrap();
    let p = t.path().join("repo");
    std::fs::create_dir_all(&p).unwrap();
    repo(&p);
    let outside = t.path().join("outside.txt");
    std::fs::write(&outside, "keep\n").unwrap();
    let exclude = p.join(".git/info/exclude");
    let _ = std::fs::remove_file(&exclude);
    std::fs::create_dir_all(p.join(".git/info")).unwrap();
    std::os::unix::fs::symlink(&outside, &exclude).unwrap();
    let err = home_workdir::ensure_place_ignored(&p).unwrap_err();
    assert!(err.contains("exclude"), "{err}");
    assert_eq!(std::fs::read_to_string(&outside).unwrap(), "keep\n");
}
