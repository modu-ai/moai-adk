//! SPEC-20261124-home-writer-links: the forge outbox, the apply receipt store and `layout-apply`
//! refuse a symbolic link between the home state root and the folder they write, with the same
//! sentence as the claim journal (`refuse_linked_home_path`). Every fixture lives in a temporary
//! folder with a temporary HOME; the environment HOME is only ever set on a child process (the
//! test re-runs itself there), so no test changes this process's environment and none reads the
//! real HOME.
#![cfg(unix)]
mod common;

use moai_cli::apply::{recover, save_before_bytes, Receipt};
use moai_cli::forge::outbox::{self, OutboxEntry};
use moai_cli::layout;
use std::fs;
use std::os::unix::fs::symlink;
use std::path::{Path, PathBuf};
use std::process::Command;

const ROOT_REFUSED: &str = "home state root is a symlink; refused";
const FOLDER_REFUSED: &str = "home state folder is a symlink; refused";
const BEHIND_BYTES: &[u8] = b"bytes behind the link";
const SOURCE_BYTES: &[u8] = b"project side bytes";

/// Set in the child process; each names a folder of the fixture the parent built.
const CHILD_PROJECT: &str = "MOAI_WRITER_TEST_PROJECT";
const CHILD_BEHIND: &str = "MOAI_WRITER_TEST_BEHIND";
const CHILD_STATE: &str = "MOAI_WRITER_TEST_STATE";
const CHILD_HOME: &str = "MOAI_WRITER_TEST_HOME";

/// Every path below `dir`, relative and sorted, with the bytes of each file; a link is listed
/// and never followed.
fn snapshot(dir: &Path) -> Vec<(String, Option<Vec<u8>>)> {
    let mut out = Vec::new();
    let mut stack = vec![dir.to_path_buf()];
    while let Some(d) = stack.pop() {
        for e in fs::read_dir(&d).unwrap().flatten() {
            let p = e.path();
            let rel = p.strip_prefix(dir).unwrap().display().to_string();
            let kind = e.file_type().unwrap();
            if kind.is_dir() {
                out.push((rel, None));
                stack.push(p);
            } else if kind.is_file() {
                out.push((rel, Some(fs::read(&p).unwrap())));
            } else {
                out.push((rel, None));
            }
        }
    }
    out.sort();
    out
}

/// A project with a profile (`widgets`) in a git clone and one receipt file on the project
/// side, a HOME whose home state root is a real folder, and a folder `behind` with one file
/// that a link can point at, in one temporary folder.
struct Tree {
    _dir: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
    state: PathBuf,
    behind: PathBuf,
}

fn tree() -> Tree {
    let dir = tempfile::tempdir().unwrap();
    let base = fs::canonicalize(dir.path()).unwrap();
    let project = base.join("project");
    fs::create_dir_all(project.join(moai_cli::PROJECT_STATE_DIR)).unwrap();
    common::git::git(&project, &["init", "-q", "-b", "main"]);
    fs::write(
        project
            .join(moai_cli::PROJECT_STATE_DIR)
            .join("project.yaml"),
        r#"{"schema_version": 1, "project_id": "widgets", "purpose": "writer fixture", "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"]}"#,
    )
    .unwrap();
    let source = moai_cli::state_root::path(&project, "receipts");
    fs::create_dir_all(&source).unwrap();
    fs::write(source.join("old.json"), SOURCE_BYTES).unwrap();
    let home = base.join("home");
    fs::create_dir(&home).unwrap();
    let state = moai_cli::home_root::path(&home);
    fs::create_dir_all(&state).unwrap();
    let behind = base.join("behind");
    fs::create_dir(&behind).unwrap();
    fs::write(behind.join("keep.txt"), BEHIND_BYTES).unwrap();
    Tree {
        _dir: dir,
        home,
        project,
        state,
        behind,
    }
}

/// Where the link sits between the home state root and the written folder.
#[derive(Clone, Copy)]
enum At {
    /// `projects/` is a link.
    Projects,
    /// `projects/widgets/` is a link.
    Project,
    /// the written store folder itself is a link.
    Store,
    /// `projects/widgets/forge/` is a link, one folder above the outbox (`forge/outbox`).
    Forge,
    /// the home state root itself is a link.
    Root,
}

impl At {
    /// The refusal sentence of a link at this place.
    fn sentence(self) -> &'static str {
        match self {
            At::Root => ROOT_REFUSED,
            _ => FOLDER_REFUSED,
        }
    }
}

impl Tree {
    /// The home state root must still be the folder the fixture built, whatever it now holds.
    fn state_is_stable(&self) {
        assert_eq!(
            moai_cli::home_root::path(&self.home),
            self.state,
            "the fixture changed which folder is the home state root"
        );
    }

    /// The pointer a verified layout apply writes, at `dir`.
    fn pointer(&self, dir: &Path) {
        fs::create_dir_all(dir).unwrap();
        fs::write(dir.join("layout.json"), b"{}").unwrap();
    }

    /// The pointer switches `store` to the home side, behind a link at `at`.
    fn switched_behind_link(&self, at: At, store: &str) {
        match at {
            At::Projects => {
                self.pointer(&self.behind.join("widgets"));
                symlink(&self.behind, self.state.join("projects")).unwrap();
            }
            At::Project => {
                self.pointer(&self.behind);
                fs::create_dir_all(self.state.join("projects")).unwrap();
                symlink(&self.behind, self.state.join("projects/widgets")).unwrap();
            }
            At::Store => {
                let project = self.state.join("projects/widgets");
                self.pointer(&project);
                let link = project.join(store);
                fs::create_dir_all(link.parent().unwrap()).unwrap();
                symlink(&self.behind, link).unwrap();
            }
            At::Forge => unreachable!("only layout-apply meets a linked forge folder"),
            At::Root => {
                self.pointer(&self.behind.join("projects/widgets"));
                fs::create_dir_all(self.behind.join("projects/widgets").join(store)).unwrap();
                self.link_root();
                return;
            }
        }
        self.state_is_stable();
    }

    /// The home state root becomes a link to `behind`. A link under the fixture's own root name
    /// would lose that root's marker and the root would resolve to the earlier name, so the
    /// fixture's empty root is removed and the earlier name is the link; the root must resolve
    /// to that link.
    fn link_root(&self) {
        fs::remove_dir(&self.state).unwrap();
        let link = self.home.join(moai_cli::HOME_STATE_DIR);
        symlink(&self.behind, &link).unwrap();
        assert_eq!(
            moai_cli::home_root::path(&self.home),
            link,
            "the home state root does not resolve to the link"
        );
    }

    /// The pointer switches `store` to the home side on a tree with no link.
    fn switched_plain(&self, store: &str) {
        let project = self.state.join("projects/widgets");
        self.pointer(&project);
        fs::create_dir_all(project.join(store)).unwrap();
        self.state_is_stable();
    }

    /// A link at `at` and no pointer anywhere: what `layout-apply` meets before it switches.
    fn unswitched_behind_link(&self, at: At) {
        match at {
            At::Projects => symlink(&self.behind, self.state.join("projects")).unwrap(),
            At::Project => {
                fs::create_dir_all(self.state.join("projects")).unwrap();
                symlink(&self.behind, self.state.join("projects/widgets")).unwrap();
            }
            At::Forge => {
                fs::create_dir_all(self.state.join("projects/widgets")).unwrap();
                symlink(&self.behind, self.state.join("projects/widgets/forge")).unwrap();
            }
            At::Store | At::Root => unreachable!("layout-apply is not run with this link"),
        }
        self.state_is_stable();
    }
}

/// What a child process sees of the fixture the parent built.
struct Seen {
    home: PathBuf,
    project: PathBuf,
    state: PathBuf,
    behind: PathBuf,
}

impl Seen {
    fn from_env() -> Option<Seen> {
        let var = |name: &str| std::env::var_os(name).map(PathBuf::from);
        Some(Seen {
            home: var(CHILD_HOME)?,
            project: var(CHILD_PROJECT)?,
            state: var(CHILD_STATE)?,
            behind: var(CHILD_BEHIND)?,
        })
    }
}

/// The libtest name of `test` as `--exact` wants it: the module path without the crate name, and
/// the bare function name while this file is the crate root of its own test binary.
fn libtest_name(test: &str) -> String {
    match module_path!().split_once("::") {
        Some((_, inside)) => format!("{inside}::{test}"),
        None => test.to_string(),
    }
}

/// The parent builds the fixture with `setup` and re-runs `test` in a child process whose
/// environment HOME is the fixture's HOME; the child runs `body` on what it sees. The test passes
/// only when the child ran exactly that one test and it passed.
fn with_home(test: &str, setup: impl FnOnce(&Tree), body: impl FnOnce(&Seen)) {
    if let Some(seen) = Seen::from_env() {
        body(&seen);
        return;
    }
    let t = tree();
    setup(&t);
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
        .env(CHILD_HOME, &t.home)
        .env(CHILD_PROJECT, &t.project)
        .env(CHILD_STATE, &t.state)
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
}

// ---- switched_dir_checked ----

/// AC-1 (REQ-2): no pointer anywhere answers no folder and refuses nothing.
#[test]
fn switched_dir_checked_answers_none_without_a_pointer() {
    with_home(
        "switched_dir_checked_answers_none_without_a_pointer",
        |t| {
            symlink(&t.behind, t.state.join("projects")).unwrap();
            t.state_is_stable();
        },
        |seen| {
            assert_eq!(
                layout::switched_dir_checked(&seen.project, "claims"),
                Ok(None)
            );
        },
    );
}

/// AC-1 (REQ-1), control: a plain tree with the pointer answers the home side folder.
#[test]
fn switched_dir_checked_answers_the_home_folder_on_a_plain_tree() {
    with_home(
        "switched_dir_checked_answers_the_home_folder_on_a_plain_tree",
        |t| t.switched_plain("claims"),
        |seen| {
            assert_eq!(
                layout::switched_dir_checked(&seen.project, "claims"),
                Ok(Some(seen.state.join("projects/widgets/claims")))
            );
        },
    );
}

/// AC-1 (REQ-1): `projects/` is a link with the pointer behind it.
#[test]
fn switched_dir_checked_refuses_a_linked_projects_folder() {
    with_home(
        "switched_dir_checked_refuses_a_linked_projects_folder",
        |t| t.switched_behind_link(At::Projects, "claims"),
        |seen| {
            assert_eq!(
                layout::switched_dir_checked(&seen.project, "claims"),
                Err(FOLDER_REFUSED.to_string())
            );
        },
    );
}

/// AC-1 (REQ-1): `projects/widgets/` is a link with the pointer behind it.
#[test]
fn switched_dir_checked_refuses_a_linked_project_folder() {
    with_home(
        "switched_dir_checked_refuses_a_linked_project_folder",
        |t| t.switched_behind_link(At::Project, "claims"),
        |seen| {
            assert_eq!(
                layout::switched_dir_checked(&seen.project, "claims"),
                Err(FOLDER_REFUSED.to_string())
            );
        },
    );
}

// ---- the forge outbox ----

/// The outbox write and read both refuse, and nothing changed behind the link.
fn outbox_refused(test: &str, at: At) {
    with_home(
        test,
        |t| t.switched_behind_link(at, "forge/outbox"),
        |seen| {
            let before = snapshot(&seen.behind);
            let entry = OutboxEntry::new("op-1", "marker");
            assert_eq!(
                outbox::save(&seen.project, &entry).unwrap_err().message,
                at.sentence()
            );
            assert_eq!(
                outbox::load(&seen.project, "op-1").unwrap_err().message,
                at.sentence()
            );
            assert_eq!(
                snapshot(&seen.behind),
                before,
                "something changed behind the link"
            );
        },
    );
}

/// AC-2 (REQ-3): `projects/` is a link.
#[test]
fn outbox_refuses_a_linked_projects_folder() {
    outbox_refused("outbox_refuses_a_linked_projects_folder", At::Projects);
}

/// AC-2 (REQ-3): `projects/widgets/` is a link.
#[test]
fn outbox_refuses_a_linked_project_folder() {
    outbox_refused("outbox_refuses_a_linked_project_folder", At::Project);
}

/// AC-2 (REQ-3): the outbox folder itself is a link.
#[test]
fn outbox_refuses_a_linked_outbox_folder() {
    outbox_refused("outbox_refuses_a_linked_outbox_folder", At::Store);
}

/// AC-2 (REQ-3): the home state root itself is a link (SPEC-20261124-home-writer-links).
#[test]
fn outbox_refuses_a_linked_home_root() {
    outbox_refused("outbox_refuses_a_linked_home_root", At::Root);
}

/// AC-2 (REQ-3), control: a plain tree writes and reads the home side outbox.
#[test]
fn outbox_refuses_no_plain_tree() {
    with_home(
        "outbox_refuses_no_plain_tree",
        |t| t.switched_plain("forge/outbox"),
        |seen| {
            let entry = OutboxEntry::new("op-1", "marker");
            outbox::save(&seen.project, &entry).unwrap();
            assert!(seen
                .state
                .join("projects/widgets/forge/outbox/op-1.json")
                .is_file());
            assert_eq!(outbox::load(&seen.project, "op-1").unwrap(), Some(entry));
        },
    );
}

// ---- the apply receipt store ----

/// The receipt, the before bytes and the recovery read all refuse, and nothing changed behind
/// the link.
fn receipt_refused(test: &str, at: At) {
    with_home(
        test,
        |t| t.switched_behind_link(at, "receipts"),
        |seen| {
            let before = snapshot(&seen.behind);
            let receipt = Receipt::new("req-1", "apply.test", &seen.project);
            assert_eq!(receipt.save(&seen.project), Err(at.sentence().to_string()));
            assert_eq!(
                save_before_bytes(&seen.project, "req-1", "a.txt", b"x"),
                Err(at.sentence().to_string())
            );
            assert_eq!(
                recover(&seen.project, "req-1", "test:perm").unwrap_err(),
                at.sentence()
            );
            assert_eq!(
                snapshot(&seen.behind),
                before,
                "something changed behind the link"
            );
        },
    );
}

/// AC-3 (REQ-4): `projects/` is a link.
#[test]
fn receipt_refuses_a_linked_projects_folder() {
    receipt_refused("receipt_refuses_a_linked_projects_folder", At::Projects);
}

/// AC-3 (REQ-4): `projects/widgets/` is a link.
#[test]
fn receipt_refuses_a_linked_project_folder() {
    receipt_refused("receipt_refuses_a_linked_project_folder", At::Project);
}

/// AC-3 (REQ-4): the receipt folder itself is a link.
#[test]
fn receipt_refuses_a_linked_receipts_folder() {
    receipt_refused("receipt_refuses_a_linked_receipts_folder", At::Store);
}

/// AC-3 (REQ-4): the home state root itself is a link (SPEC-20261124-home-writer-links).
#[test]
fn receipt_refuses_a_linked_home_root() {
    receipt_refused("receipt_refuses_a_linked_home_root", At::Root);
}

/// AC-3 (REQ-4), control: a plain tree writes the receipt and the before bytes on the home side.
#[test]
fn receipt_refuses_no_plain_tree() {
    with_home(
        "receipt_refuses_no_plain_tree",
        |t| t.switched_plain("receipts"),
        |seen| {
            let receipt = Receipt::new("req-1", "apply.test", &seen.project);
            receipt.save(&seen.project).unwrap();
            save_before_bytes(&seen.project, "req-1", "a.txt", b"x").unwrap();
            let store = seen.state.join("projects/widgets/receipts");
            assert!(store.join("req-1.json").is_file());
            assert_eq!(fs::read(store.join("req-1.before/a.txt")).unwrap(), b"x");
        },
    );
}

// ---- layout-apply ----

/// The clone side claim journal folder exists before a plan is measured: `layout-apply` opens
/// it under its lock, and a plan measured before the folder existed would be stale by then.
fn settle_claim_journal(project: &Path) {
    drop(moai_cli::worktree::open_locked_journal_in(project, None).unwrap());
}

/// `layout::apply_at` refuses with `expect`, and neither the folder behind the link nor the
/// project side source changed.
fn layout_refused(seen: &Seen, home_root: &Path, expect: &str) {
    settle_claim_journal(&seen.project);
    let source = moai_cli::state_root::path(&seen.project, "receipts");
    let (behind, from) = (snapshot(&seen.behind), snapshot(&source));
    let plan = layout::plan(&seen.project, home_root).unwrap();
    let err =
        layout::apply_at(&seen.project, home_root, &plan.plan_digest, "test:layout").unwrap_err();
    assert_eq!(err.message, expect);
    assert_eq!(
        snapshot(&seen.behind),
        behind,
        "something changed behind the link"
    );
    assert_eq!(snapshot(&source), from, "the project side source changed");
}

/// AC-4 (REQ-5): `projects/` is a link.
#[test]
fn layout_apply_refuses_a_linked_projects_folder() {
    with_home(
        "layout_apply_refuses_a_linked_projects_folder",
        |t| t.unswitched_behind_link(At::Projects),
        |seen| layout_refused(seen, &seen.state, FOLDER_REFUSED),
    );
}

/// AC-4 (REQ-5): `projects/widgets/` is a link.
#[test]
fn layout_apply_refuses_a_linked_project_folder() {
    with_home(
        "layout_apply_refuses_a_linked_project_folder",
        |t| t.unswitched_behind_link(At::Project),
        |seen| layout_refused(seen, &seen.state, FOLDER_REFUSED),
    );
}

/// AC-4 (REQ-5): `projects/widgets/forge/` is a link to a folder with no `outbox` in it, one
/// folder above the outbox's home folder (`forge/outbox`). Before the repair the last-name check
/// found nothing and the copy followed the link (SPEC-20261124-home-writer-links).
#[test]
fn layout_apply_refuses_a_linked_forge_folder() {
    with_home(
        "layout_apply_refuses_a_linked_forge_folder",
        |t| t.unswitched_behind_link(At::Forge),
        |seen| {
            layout_refused(seen, &seen.state, FOLDER_REFUSED);
            assert!(
                !seen.state.join("projects/widgets/layout.json").exists(),
                "the pointer was written"
            );
        },
    );
}

/// AC-4 (REQ-5): the home state root itself is a link; `behind` is what it points at.
#[test]
fn layout_apply_refuses_a_linked_home_root() {
    with_home(
        "layout_apply_refuses_a_linked_home_root",
        |t| {
            let linked = t.home.join("linked-root");
            symlink(&t.behind, &linked).unwrap();
        },
        |seen| {
            let linked = seen.home.join("linked-root");
            layout_refused(seen, &linked, ROOT_REFUSED);
        },
    );
}

/// AC-5 (REQ-5, REQ-6): a first apply, with no home state root and no `projects/` yet, copies the
/// store and writes the pointer as before.
#[test]
fn layout_apply_keeps_a_first_apply_on_a_plain_tree() {
    with_home(
        "layout_apply_keeps_a_first_apply_on_a_plain_tree",
        |_| {},
        |seen| {
            settle_claim_journal(&seen.project);
            let fresh = seen.home.join("fresh-state");
            assert!(!fresh.exists());
            let plan = layout::plan(&seen.project, &fresh).unwrap();
            layout::apply_at(&seen.project, &fresh, &plan.plan_digest, "test:layout").unwrap();
            assert!(fresh.join("projects/widgets/layout.json").is_file());
            assert_eq!(
                fs::read(fresh.join("projects/widgets/receipts/old.json")).unwrap(),
                SOURCE_BYTES
            );
            assert_eq!(snapshot(&seen.behind).len(), 1);
        },
    );
}

// ---- the project home folder helpers (SPEC revision 2) ----

const STAMP: &str = "20261124T000000Z";

/// Puts a link at `at` for a folder that sits at `projects/widgets/<leaf>`: the link replaces
/// `projects/`, `projects/widgets/`, that leaf, or the home state root itself.
fn link_for(t: &Tree, at: At, leaf: &str) {
    match at {
        At::Projects => symlink(&t.behind, t.state.join("projects")).unwrap(),
        At::Project => {
            fs::create_dir_all(t.state.join("projects")).unwrap();
            symlink(&t.behind, t.state.join("projects/widgets")).unwrap();
        }
        At::Store => {
            fs::create_dir_all(t.state.join("projects/widgets")).unwrap();
            symlink(&t.behind, t.state.join("projects/widgets").join(leaf)).unwrap();
        }
        At::Root => t.link_root(),
        At::Forge => unreachable!("only layout-apply meets a linked forge folder"),
    }
    if !matches!(at, At::Root) {
        t.state_is_stable();
    }
}

/// `write_area` refuses with the sentence of the link at `at`, and nothing changed behind it.
fn write_area_refused(at: At) {
    let t = tree();
    link_for(&t, at, "receipts");
    let before = snapshot(&t.behind);
    let ph = moai_cli::project_home::resolve(&t.home, &t.project).unwrap();
    assert_eq!(ph.write_area("receipts"), Err(at.sentence().to_string()));
    assert_eq!(
        snapshot(&t.behind),
        before,
        "something changed behind the link"
    );
}

/// AC-8 (REQ-7): `projects/` is a link.
#[test]
fn write_area_refuses_a_linked_projects_folder() {
    write_area_refused(At::Projects);
}

/// AC-8 (REQ-7): `projects/widgets/` is a link.
#[test]
fn write_area_refuses_a_linked_project_folder() {
    write_area_refused(At::Project);
}

/// AC-8 (REQ-7): the area folder itself is a link.
#[test]
fn write_area_refuses_a_linked_area_folder() {
    write_area_refused(At::Store);
}

/// AC-8 (REQ-7): the home state root itself is a link.
#[test]
fn write_area_refuses_a_linked_home_root() {
    write_area_refused(At::Root);
}

/// AC-8 (REQ-7), control: a plain tree answers the area folder and creates nothing.
#[test]
fn write_area_keeps_a_plain_tree() {
    let t = tree();
    let before = snapshot(&t.state);
    let ph = moai_cli::project_home::resolve(&t.home, &t.project).unwrap();
    for area in moai_cli::project_home::AREAS {
        assert_eq!(
            ph.write_area(area),
            Ok(Some(t.state.join("projects/widgets").join(area)))
        );
    }
    assert_eq!(snapshot(&t.state), before);
}

/// AC-8 (REQ-8): a project with no profile has a key made from its path; `write_area` answers
/// no folder even with a link in the way, and reads no link.
#[test]
fn write_area_keeps_a_path_key_without_reading_links() {
    let t = tree();
    let plain = t.project.parent().unwrap().join("no-profile");
    fs::create_dir(&plain).unwrap();
    symlink(&t.behind, t.state.join("projects")).unwrap();
    t.state_is_stable();
    let ph = moai_cli::project_home::resolve(&t.home, &plain).unwrap();
    assert_eq!(
        ph.key_source,
        moai_cli::project_home::KeySource::PathHash,
        "the fixture project has a profile"
    );
    assert_eq!(ph.write_area("receipts"), Ok(None));
}

/// `backup_area` refuses with the sentence of the link at `at`, and nothing changed behind it.
fn backup_area_refused(at: At) {
    let t = tree();
    link_for(&t, at, "backups");
    let before = snapshot(&t.behind);
    let ph = moai_cli::project_home::resolve(&t.home, &t.project).unwrap();
    assert_eq!(
        ph.backup_area(STAMP, "go-state"),
        Err(at.sentence().to_string())
    );
    assert_eq!(
        snapshot(&t.behind),
        before,
        "something changed behind the link"
    );
}

/// AC-9 (REQ-9): `projects/` is a link.
#[test]
fn backup_area_refuses_a_linked_projects_folder() {
    backup_area_refused(At::Projects);
}

/// AC-9 (REQ-9): `projects/widgets/` is a link.
#[test]
fn backup_area_refuses_a_linked_project_folder() {
    backup_area_refused(At::Project);
}

/// AC-9 (REQ-9): the `backups` folder itself is a link.
#[test]
fn backup_area_refuses_a_linked_backups_folder() {
    backup_area_refused(At::Store);
}

/// AC-9 (REQ-9): the home state root itself is a link.
#[test]
fn backup_area_refuses_a_linked_home_root() {
    backup_area_refused(At::Root);
}

/// AC-9 (REQ-9), control: a plain tree answers `backups/<stamp>/<kind>`.
#[test]
fn backup_area_keeps_a_plain_tree() {
    let t = tree();
    let ph = moai_cli::project_home::resolve(&t.home, &t.project).unwrap();
    assert_eq!(
        ph.backup_area(STAMP, "go-state"),
        Ok(t.state
            .join("projects/widgets/backups")
            .join(STAMP)
            .join("go-state"))
    );
}

const AUDIT_HEAD: &str = "0123456789abcdef0123456789abcdef01234567";

/// The audit clone plan of the project the child sees, with a repository address given so that
/// no remote is read.
fn audit_plan(seen: &Seen) -> Result<serde_json::Value, String> {
    moai_cli::audit_clone::plan(
        &seen.project,
        &moai_cli::audit_clone::Request {
            pr: 1,
            head: AUDIT_HEAD,
            head_branch: "feature",
            base_branch: "main",
            repo_url: Some("https://example.com/o/r.git"),
            tmp_root: &seen.home,
        },
    )
}

/// The plan succeeds, names no evidence folder and says why; nothing changed behind the link.
fn audit_plan_refused(test: &str, at: At) {
    with_home(
        test,
        |t| link_for(t, at, "audits"),
        |seen| {
            let before = snapshot(&seen.behind);
            let plan = audit_plan(seen).unwrap();
            assert!(plan["evidence_dir"].is_null(), "{plan}");
            assert_eq!(plan["evidence_reason"], serde_json::json!(FOLDER_REFUSED));
            assert_eq!(snapshot(&seen.behind), before);
        },
    );
}

/// AC-10 (REQ-10): `projects/` is a link.
#[test]
fn audit_plan_refuses_a_linked_projects_folder() {
    audit_plan_refused("audit_plan_refuses_a_linked_projects_folder", At::Projects);
}

/// AC-10 (REQ-10): `projects/widgets/` is a link.
#[test]
fn audit_plan_refuses_a_linked_project_folder() {
    audit_plan_refused("audit_plan_refuses_a_linked_project_folder", At::Project);
}

/// AC-10 (REQ-10): the `audits` folder itself is a link.
#[test]
fn audit_plan_refuses_a_linked_audits_folder() {
    audit_plan_refused("audit_plan_refuses_a_linked_audits_folder", At::Store);
}

/// AC-10 (REQ-10), control: a plain tree names the evidence folder and gives no reason.
#[test]
fn audit_plan_keeps_a_plain_tree() {
    with_home(
        "audit_plan_keeps_a_plain_tree",
        |_| {},
        |seen| {
            let plan = audit_plan(seen).unwrap();
            assert_eq!(
                plan["evidence_dir"],
                serde_json::json!(seen
                    .state
                    .join("projects/widgets/audits/1-01234567")
                    .to_string_lossy())
            );
            assert!(plan["evidence_reason"].is_null(), "{plan}");
        },
    );
}

/// AC-10 (REQ-10): a home state root that is itself a link still refuses the whole plan (the
/// earlier `identity` check, kept).
#[test]
fn audit_plan_still_refuses_a_linked_home_root() {
    with_home(
        "audit_plan_still_refuses_a_linked_home_root",
        |t| t.link_root(),
        |seen| {
            assert_eq!(audit_plan(seen), Err(ROOT_REFUSED.to_string()));
        },
    );
}
