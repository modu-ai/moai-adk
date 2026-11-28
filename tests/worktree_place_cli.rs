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

/// Points git at an empty global config for this whole test binary, so no test here reads the
/// runner's own git settings (its global ignore list could hide the folder under test).
fn isolate_git_config() {
    static ONCE: std::sync::Once = std::sync::Once::new();
    ONCE.call_once(|| {
        std::env::set_var("GIT_CONFIG_GLOBAL", "/dev/null");
        std::env::set_var("GIT_CONFIG_SYSTEM", "/dev/null");
    });
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
    isolate_git_config();
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
    isolate_git_config();
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
    isolate_git_config();
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
    isolate_git_config();
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

/// Scene 12 (REQ-10, audit B1): a call made from a subfolder checks and writes at the root of
/// the repository, so a second call from the same subfolder adds nothing.
#[test]
fn worktree_place_status_a_call_from_a_subfolder_works_at_the_root() {
    isolate_git_config();
    let t = tempfile::tempdir().unwrap();
    let p = t.path();
    repo(p);
    std::fs::create_dir_all(p.join("sub/deeper")).unwrap();
    make_place_folder(p);
    home_workdir::ensure_place_ignored(&p.join("sub/deeper")).unwrap();
    home_workdir::ensure_place_ignored(&p.join("sub")).unwrap();
    assert_eq!(exclude_lines(&p.join(".git")), 1);
    assert_eq!(git(p, &["status", "--porcelain"]), "");
}

/// Scene 13 (REQ-10, audit N1): an exclude file whose last line has no newline gets one
/// before the new line, so the earlier line is not joined to it.
#[test]
fn worktree_place_status_a_last_line_without_a_newline_is_kept_whole() {
    isolate_git_config();
    let t = tempfile::tempdir().unwrap();
    let p = t.path();
    repo(p);
    std::fs::create_dir_all(p.join(".git/info")).unwrap();
    std::fs::write(p.join(".git/info/exclude"), "# mine").unwrap();
    home_workdir::ensure_place_ignored(p).unwrap();
    assert_eq!(
        std::fs::read_to_string(p.join(".git/info/exclude")).unwrap(),
        "# mine\n/.claude/worktrees/\n"
    );
}

// ---- REQ-13: a place that runs through a link is refused ----

/// Where a link test points: a folder outside the project that must stay empty.
#[cfg(unix)]
fn link_world() -> (tempfile::TempDir, std::path::PathBuf, std::path::PathBuf) {
    let t = tempfile::tempdir().unwrap();
    let project = t.path().join("project");
    std::fs::create_dir_all(&project).unwrap();
    let outside = t.path().join("outside");
    std::fs::create_dir_all(&outside).unwrap();
    (t, project, outside)
}

/// Scene 8 (REQ-13): `.claude` that is a link is refused with the linked path named, and the
/// folder it points at gets nothing.
#[cfg(unix)]
#[test]
fn worktree_place_links_a_linked_claude_folder_is_refused() {
    let (_t, project, outside) = link_world();
    std::os::unix::fs::symlink(&outside, project.join(".claude")).unwrap();
    let folder = project.join(".claude/worktrees/project-1-k");
    let err = home_workdir::refuse_linked_place(&project, &folder).unwrap_err();
    assert!(
        err.contains(&project.join(".claude").display().to_string()),
        "{err}"
    );
    assert_eq!(std::fs::read_dir(&outside).unwrap().count(), 0);
}

/// Scene 9 (REQ-13): `.claude/worktrees` that is a link is refused with that path named.
#[cfg(unix)]
#[test]
fn worktree_place_links_a_linked_worktrees_folder_is_refused() {
    let (_t, project, outside) = link_world();
    std::fs::create_dir_all(project.join(".claude")).unwrap();
    std::os::unix::fs::symlink(&outside, project.join(".claude/worktrees")).unwrap();
    let folder = project.join(".claude/worktrees/project-1-k");
    let err = home_workdir::refuse_linked_place(&project, &folder).unwrap_err();
    assert!(
        err.contains(&project.join(".claude/worktrees").display().to_string()),
        "{err}"
    );
    assert_eq!(std::fs::read_dir(&outside).unwrap().count(), 0);
}

/// Scene 10 (REQ-13): a worktree folder that is itself a link, even a dangling one, is refused.
#[cfg(unix)]
#[test]
fn worktree_place_links_a_linked_worktree_folder_is_refused() {
    let (_t, project, outside) = link_world();
    std::fs::create_dir_all(project.join(".claude/worktrees")).unwrap();
    let folder = project.join(".claude/worktrees/project-1-k");
    std::os::unix::fs::symlink(outside.join("missing"), &folder).unwrap();
    let err = home_workdir::refuse_linked_place(&project, &folder).unwrap_err();
    assert!(err.contains(&folder.display().to_string()), "{err}");
}

/// Audit B2 (REQ-13): the place is read ignoring case, so a path spelled `.CLAUDE/worktrees`
/// still has its owner, and a link in a middle part of the path (here a group folder inside the
/// place) is refused, whatever the spelling of the parts before it.
#[cfg(unix)]
#[test]
fn worktree_place_links_every_part_below_the_owner_is_checked() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("p");
    std::fs::create_dir_all(project.join(".claude/worktrees")).unwrap();
    let elsewhere = tmp.path().join("elsewhere");
    std::fs::create_dir_all(&elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, project.join(".claude/worktrees/group")).unwrap();
    let folder = project.join(".claude/worktrees/group/w");
    assert!(home_workdir::refuse_linked_place(&project, &folder)
        .unwrap_err()
        .contains("group"));
    // The owner is found whatever the case of the place parts.
    let upper = project.join(".CLAUDE/Worktrees/w");
    assert_eq!(
        home_workdir::place_owner(&project, &upper),
        Some(project.clone())
    );
    // A folder below the project but outside the place has no owner.
    assert_eq!(
        home_workdir::place_owner(&project, &project.join("claude/worktrees/w")),
        None
    );
}

/// Audit B3 (REQ-13): the verdict refuses a `.` or `..` part anywhere in the path, plainly, and
/// leaves a path outside the place and a plain place path alone.
#[test]
fn worktree_place_verdict_refuses_dot_parts_and_passes_plain_paths() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("p");
    std::fs::create_dir_all(project.join(".claude/worktrees")).unwrap();
    let refused = |t: &Path| home_workdir::place_verdict(&project, t).unwrap_err().reason;
    assert_eq!(
        refused(&project.join(".claude/zz/../worktrees/w")),
        "place_path_refused"
    );
    assert_eq!(
        refused(&project.join(".claude/../.claude/worktrees/w")),
        "place_path_refused"
    );
    assert_eq!(refused(Path::new("a/./b")), "place_path_refused");
    assert_eq!(refused(Path::new("./w")), "place_path_refused");
    assert_eq!(refused(Path::new("../w")), "place_path_refused");
    // Plain paths: in the place, below a missing place, outside the project.
    assert!(home_workdir::place_verdict(&project, &project.join(".claude/worktrees/w")).is_ok());
    assert!(home_workdir::place_verdict(&project, &tmp.path().join("elsewhere/w")).is_ok());
    assert!(home_workdir::place_verdict(&project, Path::new("relative-name")).is_ok());
}

/// Audit B3 (REQ-13): a link anywhere between the owner and the folder is named, and the verdict
/// of a path that is clean says nothing.
#[cfg(unix)]
#[test]
fn worktree_place_verdict_names_a_link_in_any_part() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("p");
    std::fs::create_dir_all(project.join(".claude/worktrees")).unwrap();
    let elsewhere = tmp.path().join("elsewhere");
    std::fs::create_dir_all(&elsewhere).unwrap();
    std::os::unix::fs::symlink(&elsewhere, project.join(".claude/worktrees/group")).unwrap();
    let r = home_workdir::place_verdict(&project, &project.join(".claude/worktrees/group/w"))
        .unwrap_err();
    assert_eq!(r.reason, "place_is_a_link");
    assert!(r.detail.contains("group"), "{r:?}");
    // The same path through a linked alias of the project is found as well.
    let alias = tmp.path().join("alias");
    std::os::unix::fs::symlink(&project, &alias).unwrap();
    let r = home_workdir::place_verdict(&project, &alias.join(".claude/worktrees/group/w"))
        .unwrap_err();
    assert_eq!(r.reason, "place_is_a_link");
}

/// Audit B3 (REQ-13): a path that is in the place by name but whose real path is not under the
/// real place is refused as well: here a changed case where no `.claude` exists yet, so the
/// spelling is a folder of its own.
#[test]
fn worktree_place_verdict_refuses_a_path_that_resolves_outside_the_place() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("p");
    std::fs::create_dir_all(&project).unwrap();
    let r =
        home_workdir::place_verdict(&project, &project.join(".CLAUDE/worktrees/w")).unwrap_err();
    assert_eq!(r.reason, "place_path_refused");
    assert!(r.detail.contains("resolves outside"), "{r:?}");
    // The plain spelling passes, with or without the place made yet.
    assert!(home_workdir::place_verdict(&project, &project.join(".claude/worktrees/w")).is_ok());
}

/// Scene 11 (REQ-13): real folders, or none yet, are accepted.
#[test]
fn worktree_place_links_real_folders_and_missing_ones_are_accepted() {
    let t = tempfile::tempdir().unwrap();
    let project = t.path().join("project");
    std::fs::create_dir_all(&project).unwrap();
    let folder = project.join(".claude/worktrees/project-1-k");
    home_workdir::refuse_linked_place(&project, &folder).unwrap();
    std::fs::create_dir_all(project.join(".claude/worktrees")).unwrap();
    home_workdir::refuse_linked_place(&project, &folder).unwrap();
}

/// Scene 9 (REQ-6, AC-4): the tree digest that RUN records and AUDIT compares leaves the
/// project place out, so a worktree made inside the project does not change it, and the
/// project's other `.claude` files still count.
#[test]
fn worktree_place_scans_the_tree_digest_skips_the_place() {
    let project = tempfile::tempdir().unwrap();
    let root = project.path();
    std::fs::write(root.join("a.txt"), "a\n").unwrap();
    let digest = || moai_cli::tree_digest(root, &moai_cli::run::TREE_EXCLUDES).unwrap();
    let before = digest();
    let inside = root.join(".claude/worktrees/p-1-k/src");
    std::fs::create_dir_all(&inside).unwrap();
    std::fs::write(inside.join("b.rs"), "fn b() {}\n").unwrap();
    assert_eq!(
        digest(),
        before,
        "a file under the place changed the digest"
    );
    std::fs::write(root.join(".claude/settings.json"), "{}\n").unwrap();
    assert_ne!(
        digest(),
        before,
        "the project's own .claude file must still count"
    );
}

/// Scene 10 (REQ-6, AC-4): the affected-test selection reads changed paths only, and a path
/// under the place selects nothing (a pin: it holds by construction today).
#[test]
fn worktree_place_scans_the_selection_ignores_the_place() {
    let changed = vec![
        ".claude/worktrees/p-1-k/src/ops/worktree.rs".to_string(),
        ".claude/worktrees/p-1-k/tests/worktree_cli.rs".to_string(),
    ];
    let sel = moai_cli::check_affected::select(&changed, &[], None, "macos");
    assert!(sel.rust.is_empty(), "{sel:?}");
    assert_eq!(sel.scans, moai_cli::check_affected::Scans::None);
    assert_eq!(moai_cli::check_affected::area(&changed[0]), None);
}

/// AC-12 (REQ-15, REQ-16, leader note 6026): the read-only `.worktreeinclude` check names a line
/// that selects `CLAUDE.local.md` — written out, through a pattern, nested, and kept when a later
/// negation only un-selects another file — and names nothing for a list of build files or a
/// list whose later line un-selects it. The check writes nothing.
#[test]
fn worktree_place_include_names_a_line_that_copies_claude_local_md() {
    let named = |list: &str| -> Vec<String> {
        let tmp = tempfile::tempdir().unwrap();
        std::fs::write(tmp.path().join(".worktreeinclude"), list).unwrap();
        let w = moai_cli::worktree_include::warnings(tmp.path());
        assert_eq!(
            std::fs::read_to_string(tmp.path().join(".worktreeinclude")).unwrap(),
            list,
            "the check wrote the file"
        );
        w.into_iter()
            .filter(|l| l.contains("CLAUDE.local.md"))
            .collect()
    };
    for (list, line) in [
        ("CLAUDE.local.md\n", "line 1"),
        ("build/gen.rs\n*.local.md\n", "line 2"),
        ("**/CLAUDE.local.md\n", "line 1"),
        ("*.md\n!README.md\n", "line 1"),
    ] {
        let w = named(list);
        assert_eq!(w.len(), 1, "{list:?}: {w:?}");
        assert!(w[0].contains(line), "{list:?}: {w:?}");
    }
    assert!(named("build/gen.rs\nconfig/local.toml\n").is_empty());
    assert!(named("*.md\n!CLAUDE.local.md\n").is_empty());
}
/// Audit B1 (REQ-13): the owner of a place path is found by what the ancestors resolve to. A
/// path spelled through a linked alias of the project finds the project, a linked `.claude` is
/// not followed to reach it, and a path outside the place has no owner.
#[cfg(unix)]
#[test]
fn worktree_place_links_the_owner_is_found_by_what_the_path_resolves_to() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("p");
    std::fs::create_dir_all(&project).unwrap();
    let alias = tmp.path().join("alias");
    std::os::unix::fs::symlink(&project, &alias).unwrap();
    let via = alias.join(".claude/worktrees/w");
    assert_eq!(
        home_workdir::place_owner(&project, &via),
        Some(alias.clone())
    );
    assert_eq!(
        home_workdir::place_owner(&project, &project.join(".claude/worktrees/w")),
        Some(project.clone())
    );
    // Outside the place: no owner, even though the parent is the project.
    assert_eq!(
        home_workdir::place_owner(&project, &project.join("w")),
        None
    );
    assert_eq!(
        home_workdir::place_owner(&project, &tmp.path().join("other/.claude/worktrees/w")),
        None
    );
}

// ---- SPEC-20261122-worktree-place revisions 6 and 7: the open outcomes ----
//
// The tests below run the built binary on temporary repositories with a temporary HOME and a
// cleared environment. A path outside the place (REQ-2) and a linked place on the slot and the
// move (REQ-23) are observed, not refused: the command goes on as before, answers
// `would_refuse`, and the call that creates or moves records one receipt line.

use common::CleanEnv;
use serde_json::{json, Value};
use std::path::PathBuf;

/// The receipt log of a place the command would refuse.
const PLACE_LOG: &str = "worktree-place.jsonl";

/// A primary checkout `proj` with a HOME beside it, under one canonical temporary folder. The
/// repository ignores the state folders and `target/`, never `.claude/`.
struct World {
    _dir: tempfile::TempDir,
    base: PathBuf,
    repo: PathBuf,
    home: PathBuf,
}

fn world() -> World {
    let dir = tempfile::tempdir().unwrap();
    let base = std::fs::canonicalize(dir.path()).unwrap();
    let primary = base.join("proj");
    let home = base.join("home");
    std::fs::create_dir_all(&primary).unwrap();
    std::fs::create_dir_all(&home).unwrap();
    repo(&primary);
    std::fs::write(primary.join(".gitignore"), ".moai/\n.moai-cli/\ntarget/\n").unwrap();
    git_ok(&primary, &["add", ".gitignore"]);
    git_ok(&primary, &["commit", "-q", "-m", "ignore"]);
    World {
        _dir: dir,
        base,
        repo: primary,
        home,
    }
}

fn text(path: &Path) -> String {
    path.to_str().unwrap().to_string()
}

impl World {
    fn place(&self) -> PathBuf {
        self.repo.join(".claude/worktrees")
    }

    /// The binary in `cwd`, with a cleared environment whose HOME is this world's.
    fn command(&self, cwd: &Path, args: &[&str]) -> std::process::Command {
        let mut c = std::process::Command::new(common::MOAI_BIN);
        c.args(args)
            .current_dir(cwd)
            .clean_env()
            .env("HOME", &self.home)
            .env("USERPROFILE", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null");
        c
    }

    /// The answer envelope of one run in `cwd`.
    fn run_in(&self, cwd: &Path, args: &[&str]) -> Value {
        envelope(&self.command(cwd, args).output().unwrap(), args)
    }

    fn add_worktree(&self, path: &Path, branch: &str) {
        git_ok(
            &self.repo,
            &["worktree", "add", "-q", "-b", branch, &text(path)],
        );
    }

    /// A project state folder, so that a receipt log of this project has a place to go.
    fn state_folder(&self) {
        let folder = moai_cli::state_root::dir(&self.repo);
        std::fs::create_dir_all(self.repo.join(folder)).unwrap();
    }

    /// Every line of every `receipts/worktree-place.jsonl` under this world (the project's state
    /// folder or the temporary HOME). Links are not followed.
    fn place_log(&self) -> Vec<Value> {
        let mut found = Vec::new();
        let mut stack = vec![self.base.clone()];
        while let Some(dir) = stack.pop() {
            for entry in std::fs::read_dir(&dir).unwrap() {
                let entry = entry.unwrap();
                let kind = entry.file_type().unwrap();
                let path = entry.path();
                if kind.is_dir() {
                    stack.push(path);
                } else if kind.is_file()
                    && entry.file_name() == PLACE_LOG
                    && dir.file_name().is_some_and(|n| n == "receipts")
                {
                    for line in std::fs::read_to_string(&path).unwrap().lines() {
                        found.push(serde_json::from_str(line).unwrap());
                    }
                }
            }
        }
        found
    }
}

fn envelope(out: &std::process::Output, args: &[&str]) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{args:?}: stdout is not JSON: {e}\n{}\n{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// The one receipt line a creating or moving call records, with its kind and code.
fn assert_one_place_line(w: &World, code: &str) {
    let lines = w.place_log();
    assert_eq!(lines.len(), 1, "one line in {PLACE_LOG}: {lines:?}");
    assert_eq!(lines[0]["kind"], json!("would_refuse"), "{lines:?}");
    assert_eq!(lines[0]["code"], json!(code), "{lines:?}");
}

/// The claim record and comment files `worktree-plan` and `worktree-apply` read.
fn claim_args(w: &World) -> Vec<String> {
    const COMMENT: i64 = 6_721_450_649;
    let head = git(&w.repo, &["rev-parse", "HEAD"]);
    let record = w.base.join("record.json");
    let comment = w.base.join("comment.json");
    let body = json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
            "repository_id": 101, "issue_id": 7001, "issue_number": 7
        },
        "task_key": "aaaaaaaaaaaaaaaa",
        "logical_key": "06-C",
        "owner": {"kind": "session", "id": "agent-1-0dd806"},
        "epoch": 1,
        "base_revision": head,
        "worktree": {"common_dir": "/x/.git", "git_dir": "/x/.git", "path": "/x", "head": head},
        "write_scope": ["src/worktree.rs"],
        "spec_ref": {"id": "SPEC-RC1-06-C", "revision": "proposed", "digest": "sha256:abc"},
        "comment_id": COMMENT,
        "claim_token": format!("Moai-Claim: 7#{COMMENT}"),
        "lease_until": "2026-11-08T15:25:00Z",
        "state": "running",
        "effects": [],
        "last_observed_at": "2026-11-08T12:00:00Z"
    });
    std::fs::write(&record, body.to_string()).unwrap();
    std::fs::write(
        &comment,
        json!({
            "id": COMMENT,
            "body": "moai-claim: session=agent-1-0dd806 epoch=1 lease_until=2026-11-08T15:25:00Z"
        })
        .to_string(),
    )
    .unwrap();
    vec![
        "--record".into(),
        text(&record),
        "--claim-comment".into(),
        text(&comment),
    ]
}

/// `worktree-plan` (no digest) or `worktree-apply` (with the plan's digest) on the primary.
fn worktree_call(w: &World, path: Option<&str>, digest: Option<&str>) -> Value {
    let claim = claim_args(w);
    let project = text(&w.repo);
    let command = if digest.is_some() {
        "worktree-apply"
    } else {
        "worktree-plan"
    };
    let mut args = vec![command, project.as_str(), "--branch-prefix", "lane"];
    if let Some(p) = path {
        args.extend(["--path", p]);
    }
    if let Some(d) = digest {
        // The host's free space must not decide these tests.
        args.extend(["--plan-digest", d, "--disk-floor-bytes", "0"]);
    }
    args.extend(claim.iter().map(String::as_str));
    w.run_in(&w.repo, &args)
}

fn planned(w: &World, path: Option<&str>) -> Value {
    let env = worktree_call(w, path, None);
    assert_eq!(env["ok"], json!(true), "{path:?}: {env}");
    env["result"].clone()
}

fn applied(w: &World, path: Option<&str>, plan: &Value) -> Value {
    let env = worktree_call(w, path, Some(plan["plan_digest"].as_str().unwrap()));
    assert_eq!(env["ok"], json!(true), "{path:?}: {env}");
    env["result"].clone()
}

/// The paths outside the place that stay accepted and are reported: an outside folder, a deeper
/// path inside the place and the place of a linked worktree itself.
fn outside_cases() -> Vec<(&'static str, World, PathBuf)> {
    ["outside folder", "deeper path", "linked worktree's place"]
        .into_iter()
        .map(|case| {
            let w = world();
            w.state_folder();
            let path = match case {
                "outside folder" => w.base.join("outside-wt"),
                "deeper path" => w.place().join("group/deep"),
                _ => {
                    let linked = w.base.join("linked");
                    w.add_worktree(&linked, "side");
                    linked.join(".claude/worktrees/own")
                }
            };
            (case, w, path)
        })
        .collect()
}

fn assert_outside_report(w: &World, case: &str, result: &Value) {
    let report = &result["would_refuse"];
    assert_eq!(
        report["code"],
        json!("path_outside_place"),
        "{case}: no `would_refuse` for a path outside the place: {result}"
    );
    assert!(
        report["why"]
            .as_str()
            .is_some_and(|why| why.contains(&text(&w.place()))),
        "{case}: `why` names the primary place {}: {result}",
        w.place().display()
    );
}

/// AC-1 (REQ-2, observed): a path directly inside the place has no report; any other explicit
/// `--path` still plans and answers `would_refuse` naming the primary place. A plan writes
/// nothing. Red today: the plan has no `would_refuse`.
#[test]
fn worktree_place_outside_plan_reports_the_path_and_writes_nothing() {
    let w = world();
    w.state_folder();
    for direct in [
        ".claude/worktrees/direct".to_string(),
        text(&w.place().join("direct")),
    ] {
        let plan = planned(&w, Some(&direct));
        assert!(plan["would_refuse"].is_null(), "{direct}: {plan}");
    }
    assert!(w.place_log().is_empty());
    // Every case first shows that its plan answers and writes nothing; the reports follow.
    let mut reports = Vec::new();
    for (case, w, path) in outside_cases() {
        let list = git(&w.repo, &["worktree", "list", "--porcelain"]);
        let plan = planned(&w, Some(&text(&path)));
        assert_eq!(plan["writes"], json!(false), "{case}: {plan}");
        assert!(!path.exists(), "{case}: the plan made the folder");
        assert_eq!(
            git(&w.repo, &["worktree", "list", "--porcelain"]),
            list,
            "{case}"
        );
        assert_eq!(
            git(&w.repo, &["for-each-ref", "refs/heads/lane"]),
            "",
            "{case}"
        );
        assert!(w.place_log().is_empty(), "{case}: a plan records no line");
        reports.push((case, w, plan));
    }
    for (case, w, plan) in &reports {
        assert_outside_report(w, case, plan);
    }
}

/// AC-1 (REQ-2, observed): the apply creates the worktree outside the place as before, answers
/// the plan's `would_refuse` and records one line; a path directly inside the place records
/// none. Red today: no `would_refuse` and no line.
#[test]
fn worktree_place_outside_apply_creates_reports_and_records_one_line() {
    let w = world();
    w.state_folder();
    let direct = text(&w.place().join("direct"));
    let plan = planned(&w, Some(&direct));
    let made = applied(&w, Some(&direct), &plan);
    assert_eq!(made["decision"], json!("created"), "{made}");
    assert!(made["would_refuse"].is_null(), "{made}");
    assert!(w.place_log().is_empty());
    // Every case first shows that it is created as before; the reports follow.
    let mut reports = Vec::new();
    for (case, w, path) in outside_cases() {
        let path_text = text(&path);
        let plan = planned(&w, Some(&path_text));
        let made = applied(&w, Some(&path_text), &plan);
        assert_eq!(made["decision"], json!("created"), "{case}: {made}");
        assert!(path.join("README.md").is_file(), "{case}: {made}");
        reports.push((case, w, plan, made));
    }
    for (case, w, plan, made) in &reports {
        assert_outside_report(w, case, made);
        assert_eq!(made["would_refuse"], plan["would_refuse"], "{case}");
        assert_one_place_line(w, "path_outside_place");
    }
}

/// AC-3 (REQ-5): the plan names no separate build folder. Red today: it names a build folder
/// under HOME.
#[test]
fn worktree_place_build_folder_the_plan_names_none() {
    let w = world();
    let plan = planned(&w, None);
    assert_eq!(plan["build_folder"], Value::Null, "{plan}");
}

/// AC-3 (REQ-5): an apply creates the worktree and nothing under HOME, so builds use the
/// default folder inside the worktree. Red today: it creates a build folder under HOME.
#[test]
fn worktree_place_build_folder_the_apply_creates_nothing_under_home() {
    let w = world();
    let plan = planned(&w, None);
    let made = applied(&w, None, &plan);
    assert_eq!(made["decision"], json!("created"), "{made}");
    let mut under_home = Vec::new();
    let mut stack = vec![w.home.clone()];
    while let Some(dir) = stack.pop() {
        for entry in std::fs::read_dir(&dir).unwrap() {
            let entry = entry.unwrap();
            if entry.file_type().unwrap().is_dir() {
                stack.push(entry.path());
            }
            under_home.push(entry.path());
        }
    }
    assert!(
        under_home.is_empty(),
        "the apply wrote under HOME: {under_home:?}: {made}"
    );
    assert_eq!(made["build_folder"], Value::Null, "{made}");
    assert!(made.get("build_folder_error").is_none(), "{made}");
}

/// AC-4 (REQ-6): the tag walk of a project reads nothing under `.claude/worktrees/`; names that
/// only look like it are still read, and a worktree scanned as the root reads its own files.
/// Red today: the walk of the project lists the files under the place.
#[test]
fn worktree_place_scans_the_tag_walk_skips_the_place() {
    let t = tempfile::tempdir().unwrap();
    let root = t.path();
    let kept = [
        ".claude/hooks/h.rs",
        ".claude/worktrees-old/b.rs",
        "a.rs",
        "docs/worktrees/c.rs",
    ];
    let skipped = [".claude/worktrees/e.rs", ".claude/worktrees/p-1/src/d.rs"];
    for rel in kept.iter().chain(&skipped) {
        let path = root.join(rel);
        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        std::fs::write(path, "x\n").unwrap();
    }
    let scanned = |dir: &Path| -> Vec<String> {
        let report = moai_cli::atlas_tag::scan_path(dir).unwrap();
        let mut files: Vec<String> = report["files"]
            .as_array()
            .unwrap()
            .iter()
            .map(|f| {
                Path::new(f["path"].as_str().unwrap())
                    .strip_prefix(dir)
                    .unwrap()
                    .components()
                    .map(|c| c.as_os_str().to_string_lossy().into_owned())
                    .collect::<Vec<_>>()
                    .join("/")
            })
            .collect();
        files.sort();
        files
    };
    assert_eq!(scanned(&root.join(".claude/worktrees/p-1")), ["src/d.rs"]);
    assert_eq!(scanned(root), kept, "the walk read files under the place");
}

/// AC-8 (REQ-10, D-15): slots made by `worktree-slot ensure` leave the main tree's status as it
/// was in a repository that does not ignore the folder, with one exclude line. Red today: the
/// status shows `?? .claude/`.
#[test]
fn worktree_place_status_the_slot_leaves_the_main_tree_clean() {
    let w = world();
    assert_eq!(git(&w.repo, &["status", "--porcelain"]), "");
    for slot in ["lane-01", "lane-02"] {
        let env = slot_ensure(&w, slot);
        assert_eq!(env["result"]["added"], json!(true), "{env}");
    }
    assert_eq!(git(&w.repo, &["status", "--porcelain"]), "");
    assert_eq!(exclude_lines(&w.repo.join(".git")), 1);
}

/// `worktree-slot ensure <slot>` at HEAD; the call must succeed.
fn slot_ensure(w: &World, slot: &str) -> Value {
    let head = git(&w.repo, &["rev-parse", "HEAD"]);
    let project = text(&w.repo);
    let env = w.run_in(
        &w.repo,
        &[
            "worktree-slot",
            "ensure",
            slot,
            "--at",
            &head,
            "--project",
            &project,
            // The host's free disk share must not decide these tests.
            "--min-free-percent",
            "0",
        ],
    );
    assert_eq!(env["ok"], json!(true), "{env}");
    env
}

/// Makes `<repo>/<part>` a link to an empty folder outside the repository and returns that
/// folder.
#[cfg(unix)]
fn link_the_place(w: &World, part: &str) -> PathBuf {
    let outside = w.base.join("outside");
    std::fs::create_dir_all(&outside).unwrap();
    let link = w.repo.join(part);
    std::fs::create_dir_all(link.parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&outside, link).unwrap();
    outside
}

#[cfg(unix)]
fn assert_link_report(w: &World, part: &str, result: &Value) {
    let report = &result["would_refuse"];
    assert_eq!(
        report["code"],
        json!("place_is_a_link"),
        "no `would_refuse` for a linked {part}: {result}"
    );
    assert!(
        report["why"]
            .as_str()
            .is_some_and(|why| why.contains(&text(&w.repo.join(part)))),
        "`why` names the linked path: {result}"
    );
}

/// An unlinked slot first (no report, no line), then a slot through the linked `part`: added as
/// before, reported and recorded once.
#[cfg(unix)]
fn observe_slot_over_a_linked_place(part: &str) {
    let plain = world();
    plain.state_folder();
    let env = slot_ensure(&plain, "lane-01");
    assert!(env["result"]["would_refuse"].is_null(), "{env}");
    assert!(plain.place_log().is_empty());

    let w = world();
    w.state_folder();
    link_the_place(&w, part);
    let env = slot_ensure(&w, "lane-01");
    assert_eq!(env["result"]["added"], json!(true), "{env}");
    assert!(w.place().join("lane-01/README.md").is_file(), "{env}");
    assert_link_report(&w, part, &env["result"]);
    assert_one_place_line(&w, "place_is_a_link");
}

/// AC-15 (REQ-23): the slot over a linked `.claude`. Red today: no `would_refuse`, no line.
#[cfg(unix)]
#[test]
fn worktree_place_observe_the_slot_reports_a_linked_claude_folder() {
    observe_slot_over_a_linked_place(".claude");
}

/// AC-15 (REQ-23): the slot over a linked `.claude/worktrees`. Red today: the same.
#[cfg(unix)]
#[test]
fn worktree_place_observe_the_slot_reports_a_linked_worktrees_folder() {
    observe_slot_over_a_linked_place(".claude/worktrees");
}

// ---- the move (AC-14: REQ-20 to REQ-22, D-11) ----

/// A worktree on a new branch whose HEAD a remote-tracking ref holds, so nothing but its place
/// decides whether it may move.
#[cfg(unix)]
fn pushed_worktree(w: &World, path: &Path, branch: &str) {
    w.add_worktree(path, branch);
    git_ok(
        path,
        &[
            "update-ref",
            &format!("refs/remotes/origin/{branch}"),
            "HEAD",
        ],
    );
}

/// A world with one movable worktree, `clean`, in the earlier project state folder.
#[cfg(unix)]
fn move_world() -> World {
    let w = world();
    pushed_worktree(&w, &w.repo.join(".moai-cli/worktrees/clean"), "clean");
    w
}

/// The move preview (no digest) or the move (with the preview's digest) of the checkout
/// `from`, with an observation that nothing is held.
#[cfg(unix)]
fn move_call(w: &World, from: &Path, root: Option<&Path>, digest: Option<&str>) -> Value {
    let observations = w.base.join("observations.json");
    std::fs::write(
        &observations,
        json!({"schema_version": 1, "holds": {"source": "fixture", "branches": []}}).to_string(),
    )
    .unwrap();
    let command = if digest.is_some() {
        "worktree-move-apply"
    } else {
        "worktree-move-plan"
    };
    let (project, observations) = (text(from), text(&observations));
    let root = root.map(text);
    let mut args = vec![
        command,
        project.as_str(),
        "--observations",
        observations.as_str(),
    ];
    if let Some(r) = &root {
        args.extend(["--root", r.as_str()]);
    }
    if let Some(d) = digest {
        args.extend([
            "--plan-digest",
            d,
            "--permission-reference",
            "user approved the moves in the fixture",
        ]);
    }
    w.run_in(from, &args)
}

#[cfg(unix)]
fn move_result(w: &World, from: &Path, root: Option<&Path>, digest: Option<&str>) -> Value {
    let env = move_call(w, from, root, digest);
    assert_eq!(env["ok"], json!(true), "from {}: {env}", from.display());
    env["result"].clone()
}

#[cfg(unix)]
fn move_item<'a>(result: &'a Value, path: &Path) -> &'a Value {
    let path = std::fs::canonicalize(path).unwrap();
    result["items"]
        .as_array()
        .unwrap()
        .iter()
        .find(|i| i["path"] == json!(text(&path)))
        .unwrap_or_else(|| panic!("no item for {} in {result}", path.display()))
}

/// AC-14 (REQ-20, REQ-22): without `--root` the root is the primary checkout's
/// `.claude/worktrees`, also from a linked worktree, and that one folder is accepted when it is
/// named; every other root inside a worktree stays refused, and a call without a readable HOME
/// and without `--root` stays refused. Red today: the root is the home worktree folder.
#[cfg(unix)]
#[test]
fn worktree_place_move_default_root_is_the_primary_place() {
    let w = move_world();
    let linked = w.base.join("linked");
    pushed_worktree(&w, &linked, "side");
    // Unchanged behaviour: no readable HOME means no default root.
    for from in [&w.repo, &linked] {
        for home in [None, Some("relative/home")] {
            let project = text(from);
            let args = ["worktree-move-plan", project.as_str()];
            let mut c = w.command(from, &args);
            c.env_remove("USERPROFILE");
            match home {
                None => c.env_remove("HOME"),
                Some(h) => c.env("HOME", h),
            };
            let env = envelope(&c.output().unwrap(), &args);
            assert_eq!(env["ok"], json!(false), "{home:?}: {env}");
            assert!(env.to_string().contains("--root"), "{home:?}: {env}");
        }
    }
    // Unchanged behaviour: any other root inside a worktree is refused.
    for root in [
        w.repo.join("inside"),
        w.repo.join(".moai-cli/worktrees/clean/deeper"),
        w.place().join("sub"),
        linked.join(".claude/worktrees"),
    ] {
        let env = move_call(&w, &w.repo, Some(&root), None);
        assert_eq!(env["ok"], json!(false), "{}: {env}", root.display());
        assert!(env.to_string().contains("inside the worktree"), "{env}");
    }
    let place = text(&w.place());
    for from in [&w.repo, &linked] {
        let plan = move_result(&w, from, None, None);
        assert_eq!(
            plan["root"],
            json!(place),
            "the default root, from {}: {plan}",
            from.display()
        );
        assert_eq!(plan["project_root"], json!(place), "{plan}");
        let clean = move_item(&plan, &w.repo.join(".moai-cli/worktrees/clean"));
        assert_eq!(
            clean["destination"],
            json!(text(&w.place().join("proj-clean"))),
            "{plan}"
        );
    }
    assert!(
        !w.repo.join(".claude").exists(),
        "a preview creates nothing"
    );
    let named = move_result(&w, &w.repo, Some(&w.place()), None);
    assert_eq!(named["root"], json!(place), "{named}");
}

/// AC-14 (REQ-21): without `--root` a worktree in the place, at any depth, and one in each old
/// home place is kept, and only the worktree outside them moves. Red today: the worktree in the
/// place is planned for a move to the home folder.
#[cfg(unix)]
#[test]
fn worktree_place_move_keeps_the_place_and_the_old_places_without_a_root() {
    let w = move_world();
    let here = w.place().join("here");
    let kept = [
        (here.clone(), "here", Some("claude")),
        (w.place().join("group/deep"), "deep", None),
        (
            w.home.join(".moai/worktrees/proj-home"),
            "at-home",
            Some("home"),
        ),
        (
            w.home.join(".moai-cli/worktrees/proj-0123456789ab/earlier"),
            "earlier",
            Some("old_root"),
        ),
        (
            w.home.join(".moai/worktrees/projects/proj/feat"),
            "feat",
            Some("go"),
        ),
    ];
    for (path, branch, _) in &kept {
        pushed_worktree(&w, path, branch);
    }
    for from in [&w.repo, &here] {
        let plan = move_result(&w, from, None, None);
        for (path, branch, location) in &kept {
            let item = move_item(&plan, path);
            assert_eq!(item["action"], json!("keep"), "{branch}: {item}");
            if let Some(location) = location {
                assert_eq!(item["location"], json!(location), "{branch}: {item}");
            }
        }
        let clean = move_item(&plan, &w.repo.join(".moai-cli/worktrees/clean"));
        assert_eq!(clean["action"], json!("move"), "{plan}");
        assert_eq!(
            plan["totals"],
            json!({"move": 1, "keep": kept.len(), "blocked": 1}),
            "{plan}"
        );
    }
}

/// AC-14 (REQ-10, REQ-20): a move into the place writes the exclude line, so the main tree's
/// status stays clean. Red today: the destination is under the home folder.
#[cfg(unix)]
#[test]
fn worktree_place_move_apply_writes_the_exclude_line() {
    let w = move_world();
    assert_eq!(git(&w.repo, &["status", "--porcelain"]), "");
    let source = w.repo.join(".moai-cli/worktrees/clean");
    let destination = w.place().join("proj-clean");
    let plan = move_result(&w, &w.repo, None, None);
    assert_eq!(
        move_item(&plan, &source)["destination"],
        json!(text(&destination)),
        "{plan}"
    );
    let digest = plan["plan_digest"].as_str().unwrap();
    let done = move_result(&w, &w.repo, None, Some(digest));
    assert_eq!(done["state"], json!("applied"), "{done}");
    assert!(destination.join("README.md").is_file(), "{done}");
    assert_eq!(git(&w.repo, &["status", "--porcelain"]), "");
    assert_eq!(exclude_lines(&w.repo.join(".git")), 1);
}

/// AC-14: the help of the preview and of the apply names the new default. Red today: the preview
/// names the home worktree folder and neither names `.claude/worktrees`.
#[test]
fn worktree_place_move_help_names_the_new_default() {
    let w = world();
    for command in ["worktree-move-plan", "worktree-move-apply"] {
        let out = w.command(&w.repo, &[command, "--help"]).output().unwrap();
        assert!(out.status.success(), "{command}: {out:?}");
        let help = String::from_utf8(out.stdout)
            .unwrap()
            .split_whitespace()
            .collect::<Vec<_>>()
            .join(" ");
        assert!(
            !help.contains("default `<HOME>/.moai/worktrees`"),
            "{command} --help still says \"default `<HOME>/.moai/worktrees`\": {help}"
        );
        assert!(
            help.contains(".claude/worktrees"),
            "{command} --help does not name .claude/worktrees: {help}"
        );
    }
}

/// An unlinked move first (no report, no line), then a move whose place runs through the linked
/// `part`, without `--root` or with the place named: it goes on, the preview and the move
/// report it, and only the move records a line.
#[cfg(unix)]
fn observe_move_over_a_linked_place(part: &str, name_the_place: bool) {
    let plain = move_world();
    plain.state_folder();
    let other = plain.base.join("other-root");
    let plan = move_result(&plain, &plain.repo, Some(&other), None);
    assert!(plan["would_refuse"].is_null(), "{plan}");
    let digest = plan["plan_digest"].as_str().unwrap();
    let done = move_result(&plain, &plain.repo, Some(&other), Some(digest));
    assert!(done["would_refuse"].is_null(), "{done}");
    assert!(plain.place_log().is_empty());

    let w = move_world();
    w.state_folder();
    link_the_place(&w, part);
    let place = w.place();
    let root = name_the_place.then_some(place.as_path());
    let source = w.repo.join(".moai-cli/worktrees/clean");
    let plan = move_result(&w, &w.repo, root, None);
    assert!(source.join("README.md").is_file(), "{plan}");
    assert!(w.place_log().is_empty(), "a preview records no line");
    let digest = plan["plan_digest"].as_str().unwrap();
    let done = move_result(&w, &w.repo, root, Some(digest));
    let moved_behind_the_link = || {
        assert_eq!(done["state"], json!("applied"), "{done}");
        assert!(place.join("proj-clean/README.md").is_file(), "{done}");
        assert!(!source.exists(), "{done}");
    };
    // With the place named the move already goes on today, so that is shown first; without a
    // root it goes there only once the place is the default, so the reports come first.
    if name_the_place {
        moved_behind_the_link();
    }
    assert_link_report(&w, part, &plan);
    assert_link_report(&w, part, &done);
    assert_one_place_line(&w, "place_is_a_link");
    moved_behind_the_link();
}

/// AC-15 (REQ-23): the default root over a linked `.claude`. Red today: the default root is the
/// home folder and nothing is reported.
#[cfg(unix)]
#[test]
fn worktree_place_observe_the_default_move_reports_a_linked_claude_folder() {
    observe_move_over_a_linked_place(".claude", false);
}

/// AC-15 (REQ-23): the default root over a linked `.claude/worktrees`.
#[cfg(unix)]
#[test]
fn worktree_place_observe_the_default_move_reports_a_linked_worktrees_folder() {
    observe_move_over_a_linked_place(".claude/worktrees", false);
}

/// AC-15 (REQ-23): the place named as `--root` over a linked `.claude`. Red today: the move
/// goes behind the link and reports nothing.
#[cfg(unix)]
#[test]
fn worktree_place_observe_a_named_move_reports_a_linked_claude_folder() {
    observe_move_over_a_linked_place(".claude", true);
}

/// AC-15 (REQ-23): the same over a linked `.claude/worktrees`.
#[cfg(unix)]
#[test]
fn worktree_place_observe_a_named_move_reports_a_linked_worktrees_folder() {
    observe_move_over_a_linked_place(".claude/worktrees", true);
}

// ---- AC-7 (REQ-9, D-14): no text orders or defaults to the home folder for new worktrees ----

/// A checked-in text with comment markers, string continuations and line breaks folded away, so
/// a phrase is found however its lines are wrapped.
fn folded(path: &Path) -> String {
    let raw = std::fs::read_to_string(path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
    raw.lines()
        .map(|line| {
            let line = line.trim_start();
            let line = line
                .strip_prefix("///")
                .or_else(|| line.strip_prefix("//!"))
                .unwrap_or(line);
            line.trim_end_matches('\\')
        })
        .flat_map(str::split_whitespace)
        .collect::<Vec<_>>()
        .join(" ")
}

/// Fails with every phrase of `phrases` the checked-in file `rel` still holds. A text that only
/// names an old place to recognise it matches none of them.
fn assert_phrases_gone(rel: &str, phrases: &[&str]) {
    let found = folded(&Path::new(env!("CARGO_MANIFEST_DIR")).join(rel));
    let still: Vec<&&str> = phrases.iter().filter(|p| found.contains(**p)).collect();
    assert!(
        still.is_empty(),
        "{rel} still orders or defaults to the home worktree folder: {still:?}"
    );
}

#[test]
fn worktree_place_texts_help_table() {
    assert_phrases_gone(
        "src/help.rs",
        &[
            "default `<HOME>/.moai/worktrees/<project>-<issue>-<logical_key>`",
            "기본 `<HOME>/.moai/worktrees`",
        ],
    );
}

#[test]
fn worktree_place_texts_move_plan_arguments() {
    assert_phrases_gone(
        "src/ops/worktree_move_plan.rs",
        &["default `<HOME>/.moai/worktrees`"],
    );
}

#[test]
fn worktree_place_texts_move_module() {
    assert_phrases_gone(
        "src/worktree_move.rs",
        &[
            "The root defaults to `<HOME>/.moai/worktrees`",
            "The common root when the caller names none: `<HOME>/.moai/worktrees`",
        ],
    );
}

#[test]
fn worktree_place_texts_audit_probe_advice() {
    assert_phrases_gone(
        "src/audit_clone.rs",
        &["with <path> under `~/.moai/worktrees/<project>/<name>/`"],
    );
}

#[test]
fn worktree_place_texts_product_path_document() {
    assert_phrases_gone(
        "docs/rust-product-path.md",
        &["The default path is `<HOME>/.moai/worktrees/<project>-<issue>-<logical_key>`"],
    );
}

#[test]
fn worktree_place_texts_move_plan_appendix() {
    assert_phrases_gone(
        "docs/design-review/appendix/operations/worktree_move_plan.md",
        &[
            "`--root` 가 없으면 `<HOME>/.moai/worktrees` 다",
            "`home` — 이미 루트 안",
        ],
    );
}

#[test]
fn worktree_place_texts_source_manifest() {
    assert_phrases_gone(
        "skills/moai/references/source-manifest.json",
        &["new worktrees under `~/.moai/worktrees/<project>-<id>`"],
    );
}

/// Every file under `skills/` and `agents/`: none orders a new worktree into the home folder.
#[test]
fn worktree_place_texts_every_skill_and_agent_file() {
    let orders = [
        "new worktrees under `~/.moai/worktrees",
        "new worktrees under `<HOME>/.moai/worktrees",
        "under `~/.moai/worktrees/<project>",
        "default `<HOME>/.moai/worktrees",
    ];
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let mut stack = vec![root.join("skills"), root.join("agents")];
    let (mut read, mut still) = (0, Vec::new());
    while let Some(dir) = stack.pop() {
        for entry in std::fs::read_dir(&dir).unwrap() {
            let entry = entry.unwrap();
            if entry.file_type().unwrap().is_dir() {
                stack.push(entry.path());
            } else if entry.file_type().unwrap().is_file() {
                let bytes = std::fs::read(entry.path()).unwrap();
                let body = String::from_utf8_lossy(&bytes);
                read += 1;
                for order in orders.iter().filter(|o| body.contains(**o)) {
                    still.push(format!("{}: {order:?}", entry.path().display()));
                }
            }
        }
    }
    assert!(read > 0, "no file was read");
    assert!(
        still.is_empty(),
        "texts that order the home worktree folder for new worktrees: {still:#?}"
    );
}

// ---- AC-5 (REQ-7, REQ-19): the worktree base key of the project settings ----
//
// Pins, green today: the key is written only when the request asks for it, which is the request
// made after the user picks the lane setup.

/// One `claude-settings-*` run on the repository with `request` on stdin; the answer's result.
fn settings_call(w: &World, verb: &str, extra: &[&str], request: &Value) -> Value {
    let project = text(&w.repo);
    let mut args = vec![verb, project.as_str()];
    args.extend_from_slice(extra);
    let mut child = w
        .command(&w.repo, &args)
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, request.to_string().as_bytes());
    let env = envelope(&child.wait_with_output().unwrap(), &args);
    assert_eq!(env["ok"], json!(true), "{env}");
    env["result"].clone()
}

/// Plans (which writes nothing) and applies `request` on a project without a settings file, and
/// returns the settings file the apply wrote.
fn plan_and_apply_settings(request: &Value) -> Value {
    let w = world();
    let file = w.repo.join(".claude/settings.json");
    let plan = settings_call(&w, "claude-settings-plan", &[], request);
    assert!(!file.exists(), "the plan wrote the settings file");
    let digest = plan["plan_digest"].as_str().unwrap();
    let extra = [
        "--plan-digest",
        digest,
        "--permission-reference",
        "test-worktree-place",
    ];
    settings_call(&w, "claude-settings-apply", &extra, request);
    serde_json::from_slice(&std::fs::read(&file).unwrap()).unwrap()
}

/// Pin, AC-5 (REQ-19): a request without `worktree` leaves the settings without that key.
#[test]
fn worktree_place_settings_base_a_request_without_it_writes_no_worktree_key() {
    let request = json!({"outputStyle": "MoAI-Easy"});
    let written = plan_and_apply_settings(&request);
    assert!(written.get("worktree").is_none(), "{written}");
    assert_eq!(written, request);
}

/// Pin, AC-5 (REQ-7): the request for the lane setup writes `worktree.baseRef` as `"head"` and
/// nothing else.
#[test]
fn worktree_place_settings_base_the_asked_request_writes_exactly_head() {
    let request = json!({"worktree": {"baseRef": "head"}});
    assert_eq!(plan_and_apply_settings(&request), request);
}
