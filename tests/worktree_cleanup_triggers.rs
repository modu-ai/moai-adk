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

// ---- SPEC-20261123-landing-cleanup-real-clock: the landing cleanup on the real clock ---------
//
// None of these tests sets `NOW_SEAM`: every worktree is seconds old, so a cleanup that waited for
// the recent-change window or the empty-worktree guard would keep all of them.

mod real_clock {
    use super::{commit, git, Path, PathBuf, QUEUE};
    use crate::common::home::TempHome;
    use moai_cli::worktree_cleanup::{self, Request, Trigger};
    use serde_json::Value;
    use std::fs;

    const S1: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
    const S2: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

    /// The plan reads the claims of the ledger under the process's own `HOME`, as the `moai`
    /// program does under the `HOME` it is started with. These tests run in this process, so each
    /// one points `HOME` at its own temporary home and holds this lock while it does.
    static HOME_LOCK: std::sync::Mutex<()> = std::sync::Mutex::new(());

    struct Env {
        home: TempHome,
        project: PathBuf,
        n: std::cell::Cell<u32>,
        _lock: std::sync::MutexGuard<'static, ()>,
    }

    impl Env {
        fn new() -> Env {
            let lock = HOME_LOCK.lock().unwrap_or_else(|e| e.into_inner());
            let home = TempHome::new();
            std::env::set_var("HOME", home.path());
            let project = fs::canonicalize(home.project("demo")).unwrap();
            moai_cli::ledger::register(home.path(), "demo").unwrap();
            git(&project, &["init", "-q", "-b", "integ"]);
            fs::write(project.join("a.txt"), "a\n").unwrap();
            fs::write(project.join(".gitignore"), "target\n").unwrap();
            git(&project, &["add", "a.txt", ".gitignore"]);
            commit(&project, "one");
            Env {
                home,
                project,
                n: std::cell::Cell::new(0),
                _lock: lock,
            }
        }

        fn moai(&self, session: &str, args: &[&str]) -> Value {
            let out = self
                .home
                .command(args)
                .env("CLAUDE_CODE_SESSION_ID", session)
                .output()
                .unwrap();
            assert!(
                out.status.success(),
                "{args:?}: {}",
                String::from_utf8_lossy(&out.stdout)
            );
            serde_json::from_slice::<Value>(&out.stdout).unwrap()
        }

        /// A detached linked worktree of `integ`, outside the project.
        fn worktree(&self, name: &str) -> PathBuf {
            let at = fs::canonicalize(self.home.path()).unwrap().join(name);
            self.worktree_at(&at)
        }

        fn worktree_at(&self, at: &Path) -> PathBuf {
            fs::create_dir_all(at.parent().unwrap()).unwrap();
            git(
                &self.project,
                &[
                    "worktree",
                    "add",
                    "-q",
                    "--detach",
                    at.to_str().unwrap(),
                    "integ",
                ],
            );
            fs::canonicalize(at).unwrap()
        }

        /// A card with its own scope, taken by `session` in `wt`; returns its name.
        fn card(&self, session: &str, wt: &Path, branch: Option<&str>) -> String {
            self.n.set(self.n.get() + 1);
            let scope = format!("apps/{}/", self.n.get());
            let p = self.project.to_str().unwrap();
            let added = self.moai(
                session,
                &[
                    "card",
                    "add",
                    "--title",
                    "t",
                    "--scope",
                    &scope,
                    "--project",
                    p,
                ],
            );
            let name = added["result"]["name"].as_str().unwrap().to_string();
            self.moai(session, &["card", "ready", &name, "--project", p]);
            let mut take = vec!["card", "take", &name, "--worktree", wt.to_str().unwrap()];
            if let Some(b) = branch {
                take.extend(["--branch", b]);
            }
            take.extend(["--project", p]);
            self.moai(session, &take);
            name
        }

        /// The card's reservation ended, as a landing's close ends it.
        fn end(&self, session: &str, name: &str) {
            let p = self.project.to_str().unwrap();
            self.moai(
                session,
                &["card", "release", name, "--generation", "1", "--project", p],
            );
        }

        /// A card that ended in `wt`: the card a landing would have closed.
        fn landed_card(&self, wt: &Path, branch: Option<&str>) -> String {
            let name = self.card(S1, wt, branch);
            self.end(S1, &name);
            name
        }

        fn clean(&self, card: &str) -> Value {
            worktree_cleanup::run(
                &self.project,
                self.home.path(),
                &Request {
                    card_id: card.trim_start_matches("C-").parse().unwrap(),
                    integration_ref: "integ",
                    trigger: Trigger::Landing,
                },
            )
        }

        fn registered(&self, wt: &Path) -> bool {
            git(&self.project, &["worktree", "list", "--porcelain"])
                .contains(wt.file_name().unwrap().to_str().unwrap())
        }

        fn queue(&self) -> PathBuf {
            let q = self.project.join(".git").join(QUEUE);
            fs::create_dir_all(q.parent().unwrap()).unwrap();
            q
        }
    }

    /// The control: a clean, merged worktree of the landed card goes on the real clock, with or
    /// without build output.
    #[test]
    fn a_landing_removes_the_cards_clean_merged_worktree_on_the_real_clock() {
        let e = Env::new();
        let plain = e.worktree("wt-plain");
        let built = e.worktree("wt-built");
        fs::create_dir_all(built.join("target/debug")).unwrap();
        fs::write(built.join("target/debug/.cargo-lock"), "").unwrap();
        fs::write(built.join("target/debug/lib.rlib"), "x").unwrap();
        for wt in [&plain, &built] {
            let card = e.landed_card(wt, None);
            let got = e.clean(&card);
            assert_eq!(got["error"], Value::Null, "{got}");
            assert_eq!(got["worktrees"][0]["action"], "remove", "{got}");
            assert_eq!(got["worktrees"][0]["status"], "removed", "{got}");
            assert!(!wt.exists(), "{got}");
            assert!(!e.registered(wt));
        }
    }

    /// The same real-clock run keeps what the safe criteria keep, and nothing outside a
    /// worktree changes: a dirty tree, an unmerged commit, an untracked file, a linked `target`
    /// that points outside, another card's claim and a `.claude/worktrees` path.
    #[test]
    fn a_landing_on_the_real_clock_keeps_every_worktree_the_safe_criteria_keep() {
        let e = Env::new();
        let outside = fs::canonicalize(e.home.path()).unwrap().join("outside");
        fs::create_dir_all(&outside).unwrap();
        fs::write(outside.join("keep.txt"), "do not touch\n").unwrap();
        let bystander = fs::canonicalize(e.home.path())
            .unwrap()
            .join("bystander.txt");
        fs::write(&bystander, "mine\n").unwrap();

        type Spoil = Box<dyn Fn(&Env, &Path)>;
        let cases: Vec<(&str, &str, Spoil)> = vec![
            (
                "wt-tracked",
                "tracked_changes",
                Box::new(|_, wt| fs::write(wt.join("a.txt"), "changed\n").unwrap()),
            ),
            (
                "wt-unmerged",
                "not_merged",
                Box::new(|_, wt| {
                    fs::write(wt.join("b.txt"), "b\n").unwrap();
                    git(wt, &["add", "b.txt"]);
                    commit(wt, "not in the tip");
                }),
            ),
            (
                "wt-untracked",
                "untracked_files",
                Box::new(|_, wt| fs::write(wt.join("note.txt"), "n\n").unwrap()),
            ),
            (
                "wt-linked-target",
                "ignored_files_outside_build_output",
                Box::new({
                    let outside = outside.clone();
                    move |_, wt| {
                        #[cfg(unix)]
                        std::os::unix::fs::symlink(&outside, wt.join("target")).unwrap();
                    }
                }),
            ),
            (
                "wt-claimed",
                "live_claim",
                Box::new(|e, wt| {
                    e.card(S2, wt, None);
                }),
            ),
        ];
        for (name, reason, spoil) in cases {
            let wt = e.worktree(name);
            let card = e.landed_card(&wt, None);
            spoil(&e, &wt);
            let got = e.clean(&card);
            assert_eq!(got["worktrees"][0]["action"], "skip", "{name}: {got}");
            assert!(
                got["worktrees"][0]["reasons"].to_string().contains(reason),
                "{name}: {got}"
            );
            assert_eq!(got["applied"], false, "{name}: {got}");
            assert!(wt.is_dir() && e.registered(&wt), "{name} was removed");
        }
        // The preserved place.
        let under = e.project.join(".claude/worktrees/lane-x");
        let wt = e.worktree_at(&under);
        let card = e.landed_card(&wt, None);
        let got = e.clean(&card);
        assert!(
            got["worktrees"][0]["reasons"]
                .to_string()
                .contains("preserved_location"),
            "{got}"
        );
        assert!(wt.is_dir() && e.registered(&wt));
        assert_eq!(
            fs::read_to_string(outside.join("keep.txt")).unwrap(),
            "do not touch\n"
        );
        assert_eq!(fs::read_to_string(&bystander).unwrap(), "mine\n");
    }

    /// `card land --home DIR` reads the ledger of DIR, and the cleanup after it must too: another
    /// card's live claim is read from the same home, not from the process's own `HOME`. With
    /// `HOME` pointing at an empty home, the claim of the second live card still keeps the
    /// worktree the landed card shared with it; the same landing in a worktree nobody else holds
    /// is still removed (the control, so the keep is not a blanket keep).
    #[test]
    fn a_landing_reads_the_other_cards_claim_from_the_given_home_not_from_the_process_home() {
        let e = Env::new();
        let shared = e.worktree("wt-shared");
        let landed = e.landed_card(&shared, Some("card/shared"));
        e.card(S2, &shared, None);
        let alone = e.worktree("wt-alone");
        let landed_alone = e.landed_card(&alone, None);

        // The process's own HOME is another, empty home; the ledger is the one `e.home` holds.
        let other = TempHome::new();
        std::env::set_var("HOME", other.path());

        let got = e.clean(&landed);
        assert_eq!(got["error"], Value::Null, "{got}");
        assert_eq!(got["worktrees"][0]["action"], "skip", "{got}");
        assert!(
            got["worktrees"][0]["reasons"]
                .to_string()
                .contains("live_claim"),
            "{got}"
        );
        assert_eq!(got["applied"], false, "{got}");
        assert!(
            shared.is_dir() && e.registered(&shared),
            "the shared worktree was removed: {got}"
        );

        let got = e.clean(&landed_alone);
        assert_eq!(got["worktrees"][0]["status"], "removed", "{got}");
        assert!(!alone.exists(), "{got}");
    }

    /// A queue line whose worktree was skipped for a temporary reason stays in the queue; one
    /// whose worktree went, or is kept for a lasting reason, does not.
    #[test]
    fn a_queue_line_skipped_for_a_temporary_reason_stays_in_the_queue() {
        let e = Env::new();
        let own = e.worktree("wt-own");
        let landing = e.landed_card(&own, Some("card/own"));

        let claimed = e.worktree("wt-claimed");
        e.landed_card(&claimed, Some("card/claimed"));
        e.card(S2, &claimed, None);

        let dirty = e.worktree("wt-dirty");
        e.landed_card(&dirty, Some("card/dirty"));
        fs::write(dirty.join("a.txt"), "changed\n").unwrap();

        let fresh = e.worktree("wt-fresh");
        fs::create_dir_all(fresh.join("target/debug")).unwrap();
        fs::write(fresh.join("target/debug/.cargo-lock"), "").unwrap();
        fs::write(fresh.join("target/debug/lib.rlib"), "x").unwrap();
        e.landed_card(&fresh, Some("card/fresh"));

        let queue = e.queue();
        fs::write(
            &queue,
            "refs/heads/card/own\nrefs/heads/card/claimed\nrefs/heads/card/dirty\nrefs/heads/card/fresh\n",
        )
        .unwrap();
        let got = e.clean(&landing);
        assert_eq!(got["error"], Value::Null, "{got}");
        assert!(!own.exists(), "{got}");
        assert_eq!(
            fs::read_to_string(&queue).unwrap(),
            "refs/heads/card/claimed\nrefs/heads/card/fresh\n",
            "{got}"
        );
        assert!(dirty.is_dir() && claimed.is_dir() && fresh.is_dir());
        let left: Vec<String> = fs::read_dir(queue.parent().unwrap())
            .unwrap()
            .flatten()
            .map(|d| d.file_name().to_string_lossy().to_string())
            .filter(|n| n.contains("draining"))
            .collect();
        assert!(left.is_empty(), "{left:?}");
    }

    /// A taken file left by a run that died is read by the next run and removed after it.
    #[test]
    fn a_leftover_draining_file_is_read_by_the_next_run() {
        let e = Env::new();
        let own = e.worktree("wt-own");
        let landing = e.landed_card(&own, Some("card/own"));
        let old = e.worktree("wt-old");
        e.landed_card(&old, Some("card/old"));
        // Kept for a lasting reason, so its line is used up and no queue is left.
        fs::write(old.join("a.txt"), "changed\n").unwrap();
        let queue = e.queue();
        let leftover = queue.with_file_name("worktree-cleanup-queue.draining-424242");
        fs::write(&leftover, "refs/heads/card/old\n").unwrap();
        let got = e.clean(&landing);
        assert_eq!(got["error"], Value::Null, "{got}");
        assert_eq!(
            got["queue"]["gone"],
            serde_json::json!(["refs/heads/card/old"]),
            "{got}"
        );
        assert!(!leftover.exists(), "the leftover file was not used up");
        assert!(!queue.exists());
        assert!(!own.exists());
        assert!(old.is_dir(), "a dirty worktree was removed");
    }
}
