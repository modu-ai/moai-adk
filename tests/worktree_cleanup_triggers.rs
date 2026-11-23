//! SPEC-20261123-worktree-slots AC-6: what starts a worktree cleanup.
//! - a deleted branch makes the `reference-transaction` hook append one line naming the ref to a
//!   queue file and remove nothing;
//! - a landing cleans the landed card's worktrees through the existing plan and apply, on the
//!   plan's own digest, and the queue is read at the same time (the landing half is in
//!   `tests/card_land_cli.rs`, where the landing fixture lives).
//!
//! Every test runs on temporary repositories with a hook installed from the module's text.
mod common;

use common::git::git;
use moai_cli::card::guard;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Output;

/// The queue the hook appends to, in the common git folder.
const QUEUE: &str = "moai/worktree-cleanup-queue";

struct Repo {
    _dir: tempfile::TempDir,
    root: PathBuf,
}

fn commit(dir: &Path, message: &str) {
    git(
        dir,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "--allow-empty",
            "-m",
            message,
        ],
    );
}

/// A repository with the branch `integ` recorded as the protected one and the hook installed.
fn repo() -> Repo {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap().join("repo");
    fs::create_dir_all(&root).unwrap();
    git(&root, &["init", "-q", "-b", "integ"]);
    fs::write(root.join("a.txt"), "a\n").unwrap();
    git(&root, &["add", "a.txt"]);
    commit(&root, "one");
    // Work happens on a side branch: a switch back onto the protected one needs the guard program.
    git(&root, &["switch", "-q", "-c", "side"]);
    let hooks = root.join(".git/hooks");
    fs::create_dir_all(&hooks).unwrap();
    common::write_executable(&hooks.join("reference-transaction"), guard::hook_text());
    fs::create_dir_all(root.join(".git/moai")).unwrap();
    fs::write(
        root.join(".git/moai/integration-branch"),
        "refs/heads/integ\n",
    )
    .unwrap();
    Repo { _dir: dir, root }
}

impl Repo {
    fn queue(&self) -> Option<String> {
        fs::read_to_string(self.root.join(".git").join(QUEUE)).ok()
    }

    /// git with the hook on, no `moai` anywhere on the path: the queue part must not need it.
    fn try_git(&self, args: &[&str]) -> Output {
        common::git::command()
            .args(args)
            .current_dir(&self.root)
            .env("PATH", "/usr/bin:/bin")
            .env("HOME", &self.root)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap()
    }

    fn worktrees(&self) -> String {
        git(&self.root, &["worktree", "list", "--porcelain"])
    }
}

/// AC-6: a deleted branch adds one queue line and removes nothing: the linked worktree, its
/// build folder and the other branches stay as they were.
#[test]
fn a_deleted_branch_adds_one_queue_line_and_removes_nothing() {
    let r = repo();
    git(&r.root, &["branch", "card/x"]);
    git(&r.root, &["branch", "card/y"]);
    let wt = r.root.parent().unwrap().join("wt");
    git(
        &r.root,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            wt.to_str().unwrap(),
            "integ",
        ],
    );
    fs::create_dir_all(wt.join("target/debug")).unwrap();
    fs::write(wt.join("target/debug/lib.rlib"), "x").unwrap();
    let before = r.worktrees();
    assert_eq!(r.queue(), None, "nothing is queued before a delete");

    let out = r.try_git(&["branch", "-D", "card/x"]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(r.queue().as_deref(), Some("refs/heads/card/x\n"));
    let out = r.try_git(&["branch", "-D", "card/y"]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(
        r.queue().as_deref(),
        Some("refs/heads/card/x\nrefs/heads/card/y\n"),
        "one line per deleted branch"
    );

    assert_eq!(
        r.worktrees(),
        before,
        "a delete removed or moved a worktree"
    );
    assert!(wt.join("target/debug/lib.rlib").is_file());
    assert!(wt.join("a.txt").is_file());
}

/// A transaction that deletes nothing, or that does not finish, queues nothing: only the
/// committed state of a delete does. A symbolic write and a delete of a tag add no line either.
#[test]
fn only_a_committed_branch_delete_queues_a_line() {
    let r = repo();
    git(&r.root, &["branch", "card/x"]);
    git(&r.root, &["tag", "t1"]);
    // Create, move, switch, tag delete.
    for args in [
        &["branch", "card/new"][..],
        &["switch", "-q", "card/new"],
        &["switch", "-q", "side"],
        &["tag", "-d", "t1"],
    ] {
        let out = r.try_git(args);
        assert!(
            out.status.success(),
            "{args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
    }
    commit(&r.root, "two");
    assert_eq!(r.queue(), None, "an ordinary change queued a line");

    // A delete that is prepared and then aborted never reaches the committed state.
    let aborted = common::git::command()
        .args(["update-ref", "--stdin"])
        .current_dir(&r.root)
        .env("PATH", "/usr/bin:/bin")
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::null())
        .stderr(std::process::Stdio::null())
        .spawn()
        .and_then(|mut child| {
            use std::io::Write;
            child
                .stdin
                .take()
                .unwrap()
                .write_all(b"start\ndelete refs/heads/card/x\nprepare\nabort\n")?;
            child.wait()
        })
        .unwrap();
    assert!(aborted.success());
    assert_eq!(r.queue(), None, "an aborted delete queued a line");
    assert_eq!(
        git(&r.root, &["rev-parse", "--verify", "refs/heads/card/x"]).len(),
        40
    );
}

/// The queue never weakens the guard: deleting the integration branch is still refused, the
/// branch stays and no line is queued for it.
#[test]
fn deleting_the_integration_branch_is_still_refused_and_queues_nothing() {
    let r = repo();
    let out = r.try_git(&["branch", "-D", "integ"]);
    assert!(!out.status.success(), "the guard let the delete pass");
    assert_eq!(
        git(&r.root, &["rev-parse", "--verify", "refs/heads/integ"]).len(),
        40
    );
    assert_eq!(r.queue(), None);
}

/// A cleanup that fails after it took the queue puts every line back and leaves no taken file:
/// the integration ref it is given does not exist, so the plan refuses once the card's worktree
/// is known. This is the entry the verdict step calls, run with `Trigger::Verdict`.
#[test]
fn a_cleanup_that_fails_after_taking_the_queue_puts_every_line_back() {
    use common::home::TempHome;
    use moai_cli::worktree_cleanup::{self, Request, Trigger};
    let home = TempHome::new();
    let project = fs::canonicalize(home.project("demo")).unwrap();
    moai_cli::ledger::register(home.path(), "demo").unwrap();
    git(&project, &["init", "-q", "-b", "integ"]);
    fs::write(project.join("a.txt"), "a\n").unwrap();
    git(&project, &["add", "a.txt"]);
    commit(&project, "one");
    let wt = fs::canonicalize(home.path()).unwrap().join("wt");
    git(
        &project,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            wt.to_str().unwrap(),
            "integ",
        ],
    );
    let p = project.to_str().unwrap();
    let run = |args: &[&str]| {
        let out = home
            .command(args)
            .env(
                "CLAUDE_CODE_SESSION_ID",
                "aaaaaaaa-0000-4000-8000-00000000000a",
            )
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stdout)
        );
        serde_json::from_slice::<serde_json::Value>(&out.stdout).unwrap()
    };
    let added = run(&[
        "card",
        "add",
        "--title",
        "t",
        "--scope",
        "src/a.rs",
        "--project",
        p,
    ]);
    let name = added["result"]["name"].as_str().unwrap().to_string();
    run(&["card", "ready", &name, "--project", p]);
    run(&[
        "card",
        "take",
        &name,
        "--worktree",
        wt.to_str().unwrap(),
        "--project",
        p,
    ]);

    let queue = project.join(".git").join(QUEUE);
    fs::create_dir_all(queue.parent().unwrap()).unwrap();
    fs::write(&queue, "refs/heads/card/gone\nrefs/heads/card/other\n").unwrap();
    let got = worktree_cleanup::run(
        &project,
        home.path(),
        &Request {
            card_id: 1,
            integration_ref: "refs/heads/nope",
            trigger: Trigger::Verdict,
        },
    );
    assert_eq!(got["trigger"], serde_json::json!("verdict"), "{got}");
    assert_eq!(got["applied"], serde_json::json!(false), "{got}");
    assert!(
        got["error"].as_str().unwrap().contains("integration ref"),
        "{got}"
    );
    assert_eq!(
        fs::read_to_string(&queue).unwrap(),
        "refs/heads/card/gone\nrefs/heads/card/other\n",
        "the queue lost a line"
    );
    let leftovers: Vec<String> = fs::read_dir(queue.parent().unwrap())
        .unwrap()
        .flatten()
        .map(|e| e.file_name().to_string_lossy().to_string())
        .filter(|n| n.contains("draining"))
        .collect();
    assert!(leftovers.is_empty(), "a taken file was left: {leftovers:?}");
    assert!(wt.is_dir(), "a failed cleanup removed a worktree");
}
