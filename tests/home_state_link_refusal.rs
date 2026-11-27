//! SPEC-20261124-home-state-link-refusal: a symbolic link between the home state root and the
//! folder that is written is refused, for the claim journal root (`journal_root_in`) and for the
//! worktree move root (`worktree-move-plan` and `-apply`). Every fixture lives in a temporary
//! folder with a temporary HOME; the environment HOME of a test is only ever set on a child
//! process, so no test changes this process's environment and none reads the real HOME.
#![cfg(unix)]
mod common;

use moai_cli::layout::refuse_linked_home_path;
use serde_json::Value;
use std::fs;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::Command;

const ROOT_REFUSED: &str = "home state root is a symlink; refused";
const FOLDER_REFUSED: &str = "home state folder is a symlink; refused";

/// Set in the child process of the environment HOME test; names the project the child works on.
const CHILD_PROJECT: &str = "MOAI_LINK_TEST_PROJECT";
/// Set in the child process; names the folder behind the link.
const CHILD_BEHIND: &str = "MOAI_LINK_TEST_BEHIND";

/// Every path below `dir`, relative and sorted; a link is listed and never followed.
fn listing(dir: &Path) -> Vec<String> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            out.push(p.strip_prefix(dir).unwrap().display().to_string());
            if e.file_type().unwrap().is_dir() {
                stack.push(p);
            }
        }
    }
    out.sort();
    out
}

// ---- refuse_linked_home_path ----

/// A temporary home state root with a folder `projects/widgets/claims` below it, and a folder
/// `behind` beside it that a link can point at.
struct Plain {
    _dir: tempfile::TempDir,
    base: PathBuf,
    root: PathBuf,
    behind: PathBuf,
}

fn plain() -> Plain {
    let dir = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let root = base.join("state");
    fs::create_dir_all(root.join("projects/widgets/claims")).unwrap();
    let behind = base.join("behind");
    fs::create_dir(&behind).unwrap();
    Plain {
        _dir: dir,
        base,
        root,
        behind,
    }
}

/// AC-1 (REQ-1, REQ-3): each level of the path, the root first and the written folder last, is
/// refused when it is a link, and the sentence names the root or a folder below it. The middle
/// levels are the ones the earlier checks (the root, the last name) never looked at.
#[test]
fn refuse_linked_home_path_names_the_linked_component() {
    let written = |p: &Plain| p.root.join("projects/widgets/claims");
    // The root itself.
    let p = plain();
    let link = p.base.join("state-link");
    symlink(&p.root, &link).unwrap();
    assert_eq!(
        refuse_linked_home_path(&link, &link.join("projects/widgets/claims")),
        Err(ROOT_REFUSED.to_string())
    );
    // Each folder below the root.
    for below in ["projects", "projects/widgets", "projects/widgets/claims"] {
        let p = plain();
        let at = p.root.join(below);
        fs::remove_dir_all(&at).unwrap();
        fs::create_dir_all(at.parent().unwrap()).unwrap();
        symlink(&p.behind, &at).unwrap();
        assert_eq!(
            refuse_linked_home_path(&p.root, &written(&p)),
            Err(FOLDER_REFUSED.to_string()),
            "a link at {below}"
        );
    }
}

/// AC-1 (REQ-1), control: the same tree with no link passes, and so does a path whose last
/// folders do not exist yet (a journal folder that is created on first use).
#[test]
fn refuse_linked_home_path_accepts_a_plain_tree() {
    let p = plain();
    assert_eq!(
        refuse_linked_home_path(&p.root, &p.root.join("projects/widgets/claims")),
        Ok(())
    );
    assert_eq!(
        refuse_linked_home_path(&p.root, &p.root.join("projects/widgets/not-yet/claims")),
        Ok(())
    );
    assert_eq!(refuse_linked_home_path(&p.root, &p.root), Ok(()));
}

/// AC-2 (REQ-2): a path outside the home state root is accepted whatever links its own parents
/// are; a trailing slash or a `.` inside a path below the root does not change the answer.
#[test]
fn refuse_linked_home_path_ignores_a_path_outside_the_root() {
    let p = plain();
    let real = p.base.join("real");
    fs::create_dir_all(real.join("x/y")).unwrap();
    let linked = p.base.join("linked");
    symlink(&real, &linked).unwrap();
    assert_eq!(
        refuse_linked_home_path(&p.root, &linked.join("x/y")),
        Ok(())
    );
    assert_eq!(
        refuse_linked_home_path(&p.root, Path::new("/nowhere/at/all")),
        Ok(())
    );
    // Below the root, with a link in the middle: the plain, the slash and the dot spellings agree.
    let at = p.root.join("projects");
    fs::remove_dir_all(&at).unwrap();
    symlink(&p.behind, &at).unwrap();
    let refused = Err(FOLDER_REFUSED.to_string());
    assert_eq!(
        refuse_linked_home_path(&p.root, &p.root.join("projects/widgets")),
        refused
    );
    assert_eq!(
        refuse_linked_home_path(
            &p.root,
            &PathBuf::from(format!("{}/projects/widgets/", p.root.display()))
        ),
        refused
    );
    assert_eq!(
        refuse_linked_home_path(
            &p.root,
            &PathBuf::from(format!("{}/./projects/./widgets", p.root.display()))
        ),
        refused
    );
}

// ---- journal_root_in ----

/// A project with a profile (`widgets`) in a git clone, and a HOME whose home state root is a
/// real folder, in one temporary folder. `behind` is a folder beside them for a link to point at.
struct Journal {
    _dir: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
    state: PathBuf,
    behind: PathBuf,
}

fn journal_tree() -> Journal {
    let dir = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let project = base.join("project");
    fs::create_dir_all(project.join(moai_cli::PROJECT_STATE_DIR)).unwrap();
    common::git::git(&project, &["init", "-q", "-b", "main"]);
    fs::write(
        project
            .join(moai_cli::PROJECT_STATE_DIR)
            .join("project.yaml"),
        r#"{"schema_version": 1, "project_id": "widgets", "purpose": "link fixture", "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]}"#,
    )
    .unwrap();
    let home = base.join("home");
    fs::create_dir(&home).unwrap();
    let state = moai_cli::home_root::path(&home);
    fs::create_dir_all(&state).unwrap();
    let behind = base.join("behind");
    fs::create_dir(&behind).unwrap();
    Journal {
        _dir: dir,
        home,
        project,
        state,
        behind,
    }
}

impl Journal {
    /// The pointer a verified layout apply writes, at `dir`.
    fn pointer(&self, dir: &Path) {
        fs::create_dir_all(dir).unwrap();
        fs::write(dir.join("layout.json"), b"{}").unwrap();
    }

    /// The home state root must still be the folder the fixture built, whatever it now holds.
    fn state_is_stable(&self) {
        assert_eq!(
            moai_cli::home_root::path(&self.home),
            self.state,
            "the fixture changed which folder is the home state root"
        );
    }

    /// `projects/` is a link to `behind`, and the pointer is behind it.
    fn linked_projects(&self) {
        self.pointer(&self.behind.join("widgets"));
        symlink(&self.behind, self.state.join("projects")).unwrap();
        self.state_is_stable();
    }

    /// `projects/widgets` is a link to `behind`, and the pointer is behind it.
    fn linked_project_folder(&self) {
        self.pointer(&self.behind);
        fs::create_dir_all(self.state.join("projects")).unwrap();
        symlink(&self.behind, self.state.join("projects/widgets")).unwrap();
        self.state_is_stable();
    }

    /// `projects/widgets` is real with the pointer, and `claims` is a link to `behind`.
    fn linked_claims(&self) {
        self.pointer(&self.state.join("projects/widgets"));
        symlink(&self.behind, self.state.join("projects/widgets/claims")).unwrap();
        self.state_is_stable();
    }

    /// The journal root refuses with the folder sentence, the write entry that opens the journal
    /// refuses too, and nothing appeared behind the link.
    fn refused(&self) {
        let before = listing(&self.behind);
        assert_eq!(
            moai_cli::claim::journal_root_in(&self.project, Some(&self.home)),
            Err(FOLDER_REFUSED.to_string())
        );
        let opened = moai_cli::worktree::open_locked_journal_in(&self.project, Some(&self.home));
        assert_eq!(opened.unwrap_err().message, FOLDER_REFUSED);
        assert_eq!(
            listing(&self.behind),
            before,
            "something appeared behind the link"
        );
    }
}

/// AC-3 (REQ-4): `projects/` is a link, the case the earlier check missed (probe S3).
#[test]
fn journal_root_in_refuses_a_linked_projects_folder() {
    let t = journal_tree();
    t.linked_projects();
    t.refused();
}

/// AC-3 (REQ-4): `projects/<id>/` is a link.
#[test]
fn journal_root_in_refuses_a_linked_project_folder() {
    let t = journal_tree();
    t.linked_project_folder();
    t.refused();
}

/// AC-3 (REQ-4): the journal folder `claims` itself is a link.
#[test]
fn journal_root_in_refuses_a_linked_claims_folder() {
    let t = journal_tree();
    t.linked_claims();
    t.refused();
}

/// AC-5 (REQ-5): with a link on `projects/` and no pointer anywhere, the journal stays on the
/// clone side and nothing under the home state is written, so there is nothing to refuse.
#[test]
fn journal_root_in_keeps_the_clone_side_answer_without_a_pointer() {
    let t = journal_tree();
    symlink(&t.behind, t.state.join("projects")).unwrap();
    t.state_is_stable();
    let clone_side = moai_cli::claim::journal_root_in(&t.project, None).unwrap();
    assert!(
        !clone_side.starts_with(&t.state),
        "{}",
        clone_side.display()
    );
    assert_eq!(
        moai_cli::claim::journal_root_in(&t.project, Some(&t.home)),
        Ok(clone_side)
    );
}

/// AC-5 (REQ-5), control: a plain tree with the pointer answers the home side journal folder.
#[test]
fn journal_root_in_keeps_the_home_side_answer_on_a_plain_tree() {
    let t = journal_tree();
    t.pointer(&t.state.join("projects/widgets"));
    t.state_is_stable();
    assert_eq!(
        moai_cli::claim::journal_root_in(&t.project, Some(&t.home)),
        Ok(t.state.join("projects/widgets/claims"))
    );
}

/// The libtest name of `test` as `--exact` wants it: the module path without the crate name, and
/// the bare function name while this file is the crate root of its own test binary.
fn libtest_name(test: &str) -> String {
    match module_path!().split_once("::") {
        Some((_, inside)) => format!("{inside}::{test}"),
        None => test.to_string(),
    }
}

/// AC-4 (REQ-4): the environment HOME names a home state root whose `projects/` is a link with
/// the pointer behind it. The adapter `journal_root` and the write entry `open_locked_journal`
/// refuse, and nothing appears behind the link. The parent only builds the fixture; the child
/// process carries the HOME, so this process's environment is never changed.
#[test]
fn journal_root_refuses_a_linked_home_folder_in_a_child() {
    let test = "journal_root_refuses_a_linked_home_folder_in_a_child";
    let (Some(project), Some(behind)) = (
        std::env::var_os(CHILD_PROJECT).map(PathBuf::from),
        std::env::var_os(CHILD_BEHIND).map(PathBuf::from),
    ) else {
        let t = journal_tree();
        t.linked_projects();
        let out = Command::new(std::env::current_exe().unwrap())
            .args([
                "--exact",
                &libtest_name(test),
                "--nocapture",
                "--test-threads=1",
            ])
            .env("HOME", &t.home)
            .env_remove("XDG_CONFIG_HOME")
            .env_remove("GIT_CONFIG_GLOBAL")
            .env(CHILD_PROJECT, &t.project)
            .env(CHILD_BEHIND, &t.behind)
            .output()
            .unwrap();
        let text = format!(
            "{}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        );
        assert!(out.status.success(), "child failed:\n{text}");
        assert!(
            text.contains("1 passed; 0 failed"),
            "child did not run exactly {test}:\n{text}"
        );
        return;
    };
    let before = listing(&behind);
    assert_eq!(
        moai_cli::claim::journal_root(&project),
        Err(FOLDER_REFUSED.to_string())
    );
    let opened = moai_cli::worktree::open_locked_journal(&project);
    assert_eq!(opened.unwrap_err().message, FOLDER_REFUSED);
    assert_eq!(
        listing(&behind),
        before,
        "something appeared behind the link"
    );
}

// ---- worktree-move-plan and worktree-move-apply ----

/// A clone with one commit on `main` and a clean linked worktree, a HOME beside it, and a folder
/// `behind` for a link to point at, in one temporary folder.
struct Move {
    _dir: tempfile::TempDir,
    base: PathBuf,
    repo: PathBuf,
    home: PathBuf,
    behind: PathBuf,
    worktree: PathBuf,
}

fn move_tree() -> Move {
    let dir = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let repo = base.join("proj");
    let home = base.join("home");
    let behind = base.join("behind");
    for d in [&repo, &home, &behind] {
        fs::create_dir_all(d).unwrap();
    }
    let git = |dir: &Path, args: &[&str]| {
        let mut all = vec![
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ];
        all.extend_from_slice(args);
        common::git::git(dir, &all)
    };
    git(&repo, &["init", "-q", "-b", "main"]);
    fs::write(repo.join("a.txt"), "a\n").unwrap();
    git(&repo, &["add", "a.txt"]);
    git(&repo, &["commit", "-q", "-m", "init"]);
    git(&repo, &["update-ref", "refs/remotes/origin/main", "HEAD"]);
    let worktree = base.join("clean-worktree");
    git(
        &repo,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1-clean",
            worktree.to_str().unwrap(),
        ],
    );
    git(
        &worktree,
        &["update-ref", "refs/remotes/origin/issue/1-clean", "HEAD"],
    );
    Move {
        _dir: dir,
        base,
        repo,
        home,
        behind,
        worktree,
    }
}

impl Move {
    /// The HOME state folder of the worktree work folders, through its resolver.
    fn state(&self) -> PathBuf {
        moai_cli::home_workdir::root(&self.home)
    }

    /// The default move root, below [`Move::state`].
    fn worktrees(&self) -> PathBuf {
        moai_cli::home_workdir::worktrees_root(&self.home)
    }

    /// Runs the CLI in the clone with this HOME and returns the envelope on stdout.
    fn run(&self, args: &[&str]) -> Value {
        let out = Command::new(common::MOAI_BIN)
            .args(args)
            .current_dir(&self.repo)
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap();
        serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
            panic!(
                "envelope on stdout; stderr {}",
                String::from_utf8_lossy(&out.stderr)
            )
        })
    }

    fn plan(&self, extra: &[&str]) -> Value {
        let repo = self.repo.to_str().unwrap();
        let mut args = vec!["worktree-move-plan", repo];
        args.extend_from_slice(extra);
        self.run(&args)
    }

    /// The plan is refused with `sentence`, and nothing appeared behind the link.
    fn plan_refused(&self, extra: &[&str], sentence: &str) {
        let before = listing(&self.behind);
        let v = self.plan(extra);
        assert_eq!(v["ok"], Value::Bool(false), "{v}");
        assert!(v.to_string().contains(sentence), "{sentence}: {v}");
        assert_eq!(
            listing(&self.behind),
            before,
            "something appeared behind the link"
        );
    }

    fn obs(&self) -> PathBuf {
        let p = self.base.join("obs.json");
        fs::write(
            &p,
            br#"{"schema_version": 1, "holds": {"source": "fixture", "branches": []}}"#,
        )
        .unwrap();
        p
    }
}

/// AC-6 (REQ-6): the HOME state folder is itself a link; the explicit `--root` is below it.
#[test]
fn worktree_move_plan_refuses_a_linked_state_folder() {
    let m = move_tree();
    symlink(&m.behind, m.state()).unwrap();
    let root = m.worktrees();
    m.plan_refused(&["--root", root.to_str().unwrap()], ROOT_REFUSED);
}

/// AC-6 (REQ-6): the `worktrees` folder of the HOME state folder, given as the explicit `--root`,
/// is a link.
#[test]
fn worktree_move_plan_refuses_a_linked_worktrees_folder() {
    let m = move_tree();
    fs::create_dir(m.state()).unwrap();
    symlink(&m.behind, m.worktrees()).unwrap();
    let root = m.worktrees();
    m.plan_refused(&["--root", root.to_str().unwrap()], FOLDER_REFUSED);
}

/// AC-6 (REQ-6): an explicit `--root` below the HOME state folder whose parent folder is a link.
#[test]
fn worktree_move_plan_refuses_a_linked_folder_below_an_explicit_root() {
    let m = move_tree();
    fs::create_dir_all(m.worktrees()).unwrap();
    symlink(&m.behind, m.worktrees().join("sub")).unwrap();
    let root = m.worktrees().join("sub/x");
    m.plan_refused(&["--root", root.to_str().unwrap()], FOLDER_REFUSED);
}

/// AC-7 (REQ-7): an explicit `--root` outside the HOME state folder whose parent is a link (the
/// shape of a temporary folder on macOS) still plans.
#[test]
fn worktree_move_plan_accepts_a_linked_root_outside_the_home_state() {
    let m = move_tree();
    let linked = m.base.join("linked");
    symlink(&m.behind, &linked).unwrap();
    let root = linked.join("common-root");
    let v = m.plan(&["--root", root.to_str().unwrap()]);
    assert_eq!(v["ok"], Value::Bool(true), "{v}");
    assert!(v["result"]["items"].is_array(), "{v}");
}

/// AC-8 (REQ-8): `worktree-move-apply` recomputes the plan, so it is refused before the digest
/// is compared, no worktree moves, and nothing appears behind the link.
#[test]
fn worktree_move_apply_refuses_a_linked_root_and_moves_nothing() {
    let m = move_tree();
    fs::create_dir(m.state()).unwrap();
    symlink(&m.behind, m.worktrees()).unwrap();
    let before = listing(&m.behind);
    let obs = m.obs();
    let root = m.worktrees();
    let v = m.run(&[
        "worktree-move-apply",
        m.repo.to_str().unwrap(),
        "--root",
        root.to_str().unwrap(),
        "--observations",
        obs.to_str().unwrap(),
        "--plan-digest",
        "sha256:0000",
        "--permission-reference",
        "approved",
    ]);
    assert_eq!(v["ok"], Value::Bool(false), "{v}");
    let text = v.to_string();
    assert!(text.contains(FOLDER_REFUSED), "{v}");
    assert!(
        !text.contains("stale or altered plan"),
        "the link was not seen first: {v}"
    );
    assert!(m.worktree.join("a.txt").exists(), "the worktree moved");
    assert_eq!(
        listing(&m.behind),
        before,
        "something appeared behind the link"
    );
}
