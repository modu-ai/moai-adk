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
