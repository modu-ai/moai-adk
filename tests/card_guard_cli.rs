//! The guard of the integration branch (SPEC-20261121-ref-guard). Every test runs a real git in a
//! temporary repository whose integration branch is `Integ` (mixed case on purpose), with the
//! hook installed from the text the guard module provides and the `moai` of this build found the
//! way a client with no moai on its path finds it: in `$HOME/.moai/bin`. Nothing here touches the
//! real repository, its hooks or its ledger.
//!
//! The repository: `a` is the first commit and `Integ` points at it; `lane` points at `b`, a child
//! of `a`; `c` is another child of `a`, `d` a child of `b`, and `x` a commit with no history in
//! common with them. The primary worktree is detached so `Integ` can be deleted, except in the
//! variant that has the branch checked out in a linked worktree.

#![cfg(unix)]

mod common;

use common::home::TempHome;
use moai_cli::card::guard;
use serde_json::json;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const INTEG: &str = "Integ";
const ZERO: &str = "0000000000000000000000000000000000000000";

fn git_executable() -> PathBuf {
    let path = std::env::var_os("PATH").expect("PATH");
    std::env::split_paths(&path)
        .map(|dir| dir.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

struct Repo {
    home: TempHome,
    dir: tempfile::TempDir,
    repo: PathBuf,
    lane_wt: PathBuf,
    /// A directory on the git runner's path that holds a stand-in `moai`, when a test wants one.
    standin: Option<PathBuf>,
    /// Extra environment for every git call.
    env: Vec<(String, String)>,
    a: String,
    b: String,
    c: String,
    d: String,
    x: String,
}

impl Repo {
    fn new(checked_out: bool) -> Self {
        let home = TempHome::new();
        let dir = tempfile::Builder::new()
            .prefix("moai-test-guard-")
            .tempdir()
            .unwrap();
        let repo = dir.path().join("r");
        fs::create_dir_all(&repo).unwrap();
        let mut w = Repo {
            lane_wt: dir.path().join("lane-wt"),
            home,
            dir,
            repo: repo.clone(),
            standin: None,
            env: Vec::new(),
            a: String::new(),
            b: String::new(),
            c: String::new(),
            d: String::new(),
            x: String::new(),
        };
        w.git(&repo, &["init", "-q", "-b", "main"]);
        fs::write(repo.join("a.txt"), "a\n").unwrap();
        w.git(&repo, &["add", "a.txt"]);
        w.git(&repo, &["commit", "-q", "-m", "init"]);
        w.a = w.git(&repo, &["rev-parse", "HEAD"]);
        let a = w.a.clone();
        w.b = w.commit_tree(&a, &[&a], "lane head");
        w.c = w.commit_tree(&a, &[&a], "another child");
        let b = w.b.clone();
        w.d = w.commit_tree(&a, &[&b], "after the lane head");
        w.x = w.commit_tree(&a, &[], "no history in common");
        w.git(&repo, &["branch", INTEG, &a]);
        w.git(&repo, &["branch", "lane", &b]);
        if checked_out {
            let wi = w.dir.path().join("integ-wt");
            w.git(
                &repo,
                &["worktree", "add", "-q", wi.to_str().unwrap(), INTEG],
            );
        } else {
            w.git(&repo, &["checkout", "-q", "--detach"]);
        }
        let lane_wt = w.lane_wt.clone();
        w.git(
            &repo,
            &["worktree", "add", "-q", lane_wt.to_str().unwrap(), "lane"],
        );
        w
    }

    fn common(&self) -> PathBuf {
        self.repo.join(".git")
    }

    fn integ_wt(&self) -> PathBuf {
        self.dir.path().join("integ-wt")
    }

    /// A commit with the tree of `tree_of` and the given parents; no ref moves.
    fn commit_tree(&self, tree_of: &str, parents: &[&str], message: &str) -> String {
        let tree = self.git(&self.repo, &["rev-parse", &format!("{tree_of}^{{tree}}")]);
        let mut args = vec!["commit-tree", tree.as_str()];
        for parent in parents {
            args.push("-p");
            args.push(parent);
        }
        args.push("-m");
        args.push(message);
        self.git(&self.repo, &args)
    }

    fn path_var(&self) -> String {
        let mut dirs: Vec<PathBuf> = Vec::new();
        if let Some(standin) = &self.standin {
            dirs.push(standin.clone());
        }
        dirs.push(git_executable().parent().unwrap().to_path_buf());
        dirs.push(PathBuf::from("/usr/bin"));
        dirs.push(PathBuf::from("/bin"));
        std::env::join_paths(dirs)
            .unwrap()
            .to_string_lossy()
            .into_owned()
    }

    fn command(&self, dir: &Path, args: &[&str]) -> Command {
        let mut cmd = Command::new(git_executable());
        cmd.arg("-C")
            .arg(dir)
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
                "-c",
                "gc.auto=0",
                "-c",
                "maintenance.auto=false",
            ])
            .args(args)
            .env_clear()
            .env("HOME", self.home.path())
            .env("PATH", self.path_var())
            .env("LC_ALL", "C")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null");
        for (key, value) in &self.env {
            cmd.env(key, value);
        }
        cmd
    }

    /// A git call that must succeed; returns its trimmed stdout.
    fn git(&self, dir: &Path, args: &[&str]) -> String {
        let out = self
            .command(dir, args)
            .stdin(Stdio::null())
            .output()
            .unwrap();
        assert!(
            out.status.success(),
            "git {args:?}: {}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8_lossy(&out.stdout).trim().to_string()
    }

    /// A git call that may fail.
    fn try_git(&self, dir: &Path, args: &[&str]) -> Output {
        self.command(dir, args)
            .stdin(Stdio::null())
            .output()
            .unwrap()
    }

    /// A git call that reads `input` on its standard input.
    fn try_git_stdin(&self, dir: &Path, args: &[&str], input: &str) -> Output {
        let mut child = self
            .command(dir, args)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::write_stdin(child.stdin.take().unwrap(), input.as_bytes());
        child.wait_with_output().unwrap()
    }

    /// The commit a reference names now, or `None`.
    fn rev(&self, reference: &str) -> Option<String> {
        let out = self.try_git(&self.repo, &["rev-parse", "-q", "--verify", reference]);
        out.status
            .success()
            .then(|| String::from_utf8_lossy(&out.stdout).trim().to_string())
    }

    fn integ(&self) -> String {
        self.rev("refs/heads/Integ").expect("Integ exists")
    }

    fn write_record(&self, reference: &str) {
        fs::create_dir_all(self.common().join("moai")).unwrap();
        fs::write(
            self.common().join(guard::RECORD_FILE),
            format!("{reference}\n"),
        )
        .unwrap();
    }

    fn install_hook(&self) {
        let hooks = self.common().join("hooks");
        fs::create_dir_all(&hooks).unwrap();
        common::write_executable(&hooks.join("reference-transaction"), guard::hook_text());
    }

    /// The `moai` of this build where a client with no moai on its path looks for it.
    fn link_moai(&self) {
        let bin = self.home.path().join(".moai/bin");
        fs::create_dir_all(&bin).unwrap();
        std::os::unix::fs::symlink(common::MOAI_BIN, bin.join("moai")).unwrap();
    }

    /// The record, the hook and the moai the hook finds.
    fn install(&self) {
        self.write_record("refs/heads/Integ");
        self.install_hook();
        self.link_moai();
    }

    fn write_intent(&self, reference: &str, old: &str, new: &str) {
        fs::create_dir_all(self.common().join("moai")).unwrap();
        fs::write(
            self.common().join(guard::INTENT_FILE),
            json!({"ref": reference, "old": old, "new": new, "by": "test", "at": 1}).to_string(),
        )
        .unwrap();
    }

    fn intent_exists(&self) -> bool {
        self.common().join(guard::INTENT_FILE).exists()
    }

    fn remove_intent(&self) {
        let _ = fs::remove_file(self.common().join(guard::INTENT_FILE));
    }

    /// Asserts a refusal by the guard and returns git's message.
    fn refused(&self, out: &Output) -> String {
        let message = String::from_utf8_lossy(&out.stderr).into_owned();
        assert!(!out.status.success(), "the call was not refused: {message}");
        assert!(
            message.contains("moai ref guard"),
            "the refusal is not the guard's: {message}"
        );
        message
    }

    /// The names of every branch reference, sorted.
    fn branches(&self) -> Vec<String> {
        let mut names: Vec<String> = self
            .git(
                &self.repo,
                &["for-each-ref", "--format=%(refname)", "refs/heads"],
            )
            .lines()
            .map(str::to_string)
            .collect();
        names.sort();
        names
    }
}

fn tree_clean(w: &Repo) {
    assert_eq!(w.integ(), w.a, "the integration branch moved");
}

// ---- AC-1: every way of moving the integration branch is refused ----------------------------

/// AC-1: with the hook installed, the ways of moving, deleting or renaming the integration
/// branch that need no moai leave it where it was, and an environment variable typed by hand is
/// no pass.
#[test]
fn ac1_every_way_of_moving_the_integration_branch_is_refused() {
    let mut w = Repo::new(false);
    w.install();
    let lane = w.lane_wt.clone();
    let repo = w.repo.clone();
    let (a, b) = (w.a.clone(), w.b.clone());
    let attempts: Vec<(&str, &Path, Vec<&str>)> = vec![
        (
            "update-ref",
            &repo,
            vec!["update-ref", "refs/heads/Integ", &b, &a],
        ),
        (
            "update-ref from a lane worktree",
            &lane,
            vec!["update-ref", "refs/heads/Integ", &b],
        ),
        ("branch -f", &lane, vec!["branch", "-f", INTEG, "lane"]),
        ("branch -D", &lane, vec!["branch", "-D", INTEG]),
        ("branch -m", &lane, vec!["branch", "-m", INTEG, "Gone"]),
        (
            "push .",
            &lane,
            vec!["push", "-q", ".", "HEAD:refs/heads/Integ"],
        ),
        (
            "fetch .",
            &lane,
            vec!["fetch", "-q", ".", "lane:refs/heads/Integ"],
        ),
    ];
    for (label, dir, args) in attempts {
        let out = w.try_git(dir, &args);
        let message = w.refused(&out);
        assert!(message.contains("moai card land"), "{label}: {message}");
        tree_clean(&w);
        assert!(
            w.rev("refs/heads/Gone").is_none(),
            "{label}: a branch was made"
        );
    }
    let out = w.try_git_stdin(
        &lane,
        &["update-ref", "--stdin"],
        &format!("update refs/heads/Integ {b}\n"),
    );
    w.refused(&out);
    tree_clean(&w);
    // An environment variable typed by hand is no pass.
    w.env.push(("MOAI_LANDER".to_string(), "1".to_string()));
    let out = w.try_git(&lane, &["update-ref", "refs/heads/Integ", &b, &a]);
    w.refused(&out);
    tree_clean(&w);

    // With the branch checked out in a linked worktree: a commit and a reset --hard on it.
    let mut v = Repo::new(true);
    v.install();
    let wi = v.integ_wt();
    let out = v.try_git(&wi, &["commit", "-q", "--allow-empty", "-m", "direct"]);
    v.refused(&out);
    tree_clean(&v);
    let out = v.try_git(&wi, &["reset", "-q", "--hard", "lane"]);
    v.refused(&out);
    tree_clean(&v);
    v.env.clear();
}

// ---- AC-2: the intent record --------------------------------------------------------------

/// AC-2: the move the intent record names passes and removes the record; every other record
/// refuses; a finished move cannot be replayed; no record allows a deletion; a creation needs a
/// record that names zero as the old value.
#[test]
fn ac2_the_move_the_intent_record_names_passes_and_only_that_one() {
    let w = Repo::new(false);
    w.install();
    let (a, b, c, d, x) = (
        w.a.clone(),
        w.b.clone(),
        w.c.clone(),
        w.d.clone(),
        w.x.clone(),
    );
    let repo = w.repo.clone();

    // The named move passes and consumes the record.
    w.write_intent("refs/heads/Integ", &a, &b);
    w.git(&repo, &["update-ref", "refs/heads/Integ", &b, &a]);
    assert_eq!(w.integ(), b);
    assert!(!w.intent_exists(), "the record of a finished move stays");

    // A replay of the finished move: the record is back, the branch is not where it names.
    w.write_intent("refs/heads/Integ", &a, &b);
    w.refused(&w.try_git(&repo, &["update-ref", "refs/heads/Integ", &a, &b]));
    assert_eq!(w.integ(), b);

    // Every other record refuses: another new value, another old value, another ref, a new value
    // that is no descendant, and a deletion.
    let cases: Vec<(&str, String, String, String, Vec<String>)> = vec![
        (
            "another new value",
            "refs/heads/Integ".into(),
            b.clone(),
            d.clone(),
            vec![
                "update-ref".into(),
                "refs/heads/Integ".into(),
                c.clone(),
                b.clone(),
            ],
        ),
        (
            "another old value",
            "refs/heads/Integ".into(),
            a.clone(),
            d.clone(),
            vec![
                "update-ref".into(),
                "refs/heads/Integ".into(),
                d.clone(),
                b.clone(),
            ],
        ),
        (
            "another ref",
            "refs/heads/lane".into(),
            b.clone(),
            d.clone(),
            vec![
                "update-ref".into(),
                "refs/heads/Integ".into(),
                d.clone(),
                b.clone(),
            ],
        ),
        (
            "no descendant",
            "refs/heads/Integ".into(),
            b.clone(),
            x.clone(),
            vec![
                "update-ref".into(),
                "refs/heads/Integ".into(),
                x.clone(),
                b.clone(),
            ],
        ),
        (
            "a deletion",
            "refs/heads/Integ".into(),
            b.clone(),
            ZERO.into(),
            vec![
                "update-ref".into(),
                "-d".into(),
                "refs/heads/Integ".into(),
                b.clone(),
            ],
        ),
    ];
    for (label, reference, old, new, args) in cases {
        w.write_intent(&reference, &old, &new);
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        let out = w.try_git(&repo, &args);
        let message = w.refused(&out);
        assert!(message.contains("intent"), "{label}: {message}");
        assert_eq!(w.integ(), b, "{label}: the branch moved");
    }

    // The named move passes again after all of that.
    w.write_intent("refs/heads/Integ", &b, &d);
    w.git(&repo, &["update-ref", "refs/heads/Integ", &d, &b]);
    assert_eq!(w.integ(), d);
    assert!(!w.intent_exists());

    // A creation: refused without a record, passes with one that names zero as the old value.
    w.write_record("refs/heads/Fresh");
    w.remove_intent();
    w.refused(&w.try_git(&repo, &["update-ref", "refs/heads/Fresh", &a]));
    assert!(w.rev("refs/heads/Fresh").is_none());
    w.write_intent("refs/heads/Fresh", ZERO, &a);
    w.git(&repo, &["update-ref", "refs/heads/Fresh", &a]);
    assert_eq!(w.rev("refs/heads/Fresh"), Some(a));
}

// ---- AC-3: packing and gc are not deletions ------------------------------------------------

/// AC-3: `pack-refs` and `gc` pass in the loose and in the packed state and from a linked
/// worktree, the branch keeps its value, and a delete that names no old value refuses in the
/// loose-only and in the packed-only state.
#[test]
fn ac3_packing_and_gc_pass_and_a_real_delete_does_not() {
    let w = Repo::new(false);
    w.install();
    let repo = w.repo.clone();
    let lane = w.lane_wt.clone();
    w.git(&repo, &["pack-refs", "--all"]);
    tree_clean(&w);
    w.git(&repo, &["gc", "-q"]);
    tree_clean(&w);
    w.git(&repo, &["pack-refs", "--all"]);
    w.git(&repo, &["pack-refs", "--all", "--no-prune"]);
    w.git(&lane, &["pack-refs", "--all"]);
    tree_clean(&w);
    // The packed-only state: deletes refuse.
    assert!(!w.common().join("refs/heads/Integ").exists());
    w.refused(&w.try_git(&repo, &["branch", "-D", INTEG]));
    tree_clean(&w);
    w.refused(&w.try_git(&repo, &["update-ref", "-d", "refs/heads/Integ"]));
    tree_clean(&w);

    // The loose-only state, in a repository of its own.
    let v = Repo::new(false);
    v.install();
    assert!(v.common().join("refs/heads/Integ").exists());
    v.refused(&v.try_git(&v.repo, &["branch", "-D", INTEG]));
    tree_clean(&v);
    v.refused(&v.try_git(&v.repo, &["update-ref", "-d", "refs/heads/Integ"]));
    tree_clean(&v);
}

// ---- AC-4: case variants -------------------------------------------------------------------

/// AC-4: creating a branch whose name differs only in case, one whose first path part differs
/// only in case, and moving or deleting such a branch that already exists, refuse in the packed
/// state with the recorded spelling and the landing command in the message; another name passes.
#[test]
fn ac4_case_variants_of_the_integration_branch_are_refused() {
    let w = Repo::new(false);
    w.install();
    let repo = w.repo.clone();
    let b = w.b.clone();
    w.git(&repo, &["pack-refs", "--all"]);
    let before = w.branches();
    for args in [
        vec!["branch", "integ"],
        vec!["branch", "INTEG"],
        vec!["branch", "integ/x"],
        vec!["update-ref", "refs/heads/integ", &b],
    ] {
        let out = w.try_git(&repo, &args);
        let message = w.refused(&out);
        assert!(
            message.contains("`Integ`") && message.contains("moai card land"),
            "{args:?}: {message}"
        );
        assert_eq!(w.branches(), before, "{args:?}: a branch appeared");
    }
    w.git(&repo, &["branch", "other"]);

    // A variant that already exists (made before the hook): moving and deleting it refuse.
    let v = Repo::new(false);
    v.git(&v.repo, &["pack-refs", "--all"]);
    v.git(&v.repo, &["update-ref", "refs/heads/INTEG", &v.b]);
    v.git(&v.repo, &["pack-refs", "--all"]);
    v.install();
    let existing = v.branches();
    v.refused(&v.try_git(&v.repo, &["update-ref", "refs/heads/INTEG", &v.c]));
    v.refused(&v.try_git(&v.repo, &["branch", "-D", "INTEG"]));
    assert_eq!(v.branches(), existing);
    assert_eq!(v.rev("refs/heads/INTEG"), Some(v.b.clone()));
}

// ---- AC-5: no alias --------------------------------------------------------------------------

/// AC-5: a symbolic ref to the integration branch or a variant refuses, the integration branch
/// made symbolic refuses, and a worktree that checks the branch out keeps working.
#[test]
fn ac5_an_alias_of_the_integration_branch_is_refused() {
    let w = Repo::new(false);
    w.install();
    let repo = w.repo.clone();
    for args in [
        vec!["symbolic-ref", "refs/heads/alias", "refs/heads/Integ"],
        vec!["symbolic-ref", "refs/heads/integ", "refs/heads/Integ"],
        vec!["symbolic-ref", "refs/heads/Integ", "refs/heads/lane"],
    ] {
        let out = w.try_git(&repo, &args);
        assert!(!out.status.success(), "{args:?} passed");
        tree_clean(&w);
    }
    assert!(w.rev("refs/heads/alias").is_none());
    let out = w.try_git(
        &repo,
        &["symbolic-ref", "refs/heads/alias", "refs/heads/Integ"],
    );
    w.refused(&out);
    // A worktree's HEAD is no ref under refs/heads: checking the branch out keeps working.
    w.git(&repo, &["checkout", "-q", INTEG]);
    w.git(&repo, &["checkout", "-q", "--detach"]);
    tree_clean(&w);
}

// ---- AC-6: the shell filter and the missing program ------------------------------------------

/// A stand-in `moai` that records its arguments and fails at the state in `STANDIN_FAIL_AT`.
const STANDIN: &str = "#!/bin/sh\necho \"$@\" >> \"$STANDIN_LOG\"\n[ \"$3\" = \"$STANDIN_FAIL_AT\" ] && exit 1\nexit 0\n";

fn calls(log: &Path) -> Vec<String> {
    fs::read_to_string(log)
        .unwrap_or_default()
        .lines()
        .map(str::to_string)
        .collect()
}

/// AC-6: the hook starts the guard only for a transaction that touches the recorded branch, a
/// variant or a symbolic value, and only at the prepared and committed states; with no moai it
/// refuses those and nothing else; with no record nothing starts.
#[test]
fn ac6_the_hook_starts_the_guard_only_where_it_is_needed() {
    let mut w = Repo::new(false);
    w.install();
    let log = w.dir.path().join("standin.log");
    let standin_dir = w.dir.path().join("standin");
    fs::create_dir_all(&standin_dir).unwrap();
    common::write_executable(&standin_dir.join("moai"), STANDIN);
    w.standin = Some(standin_dir);
    w.env
        .push(("STANDIN_LOG".into(), log.to_str().unwrap().into()));
    let repo = w.repo.clone();
    let lane = w.lane_wt.clone();
    let (a, b) = (w.a.clone(), w.b.clone());

    // An unrelated branch: no program starts at any state.
    w.git(&lane, &["commit", "-q", "--allow-empty", "-m", "lane work"]);
    assert!(calls(&log).is_empty(), "{:?}", calls(&log));

    // A move of the recorded branch: the guard at prepared and at committed, nothing else.
    w.git(&repo, &["update-ref", "refs/heads/Integ", &b, &a]);
    assert_eq!(
        calls(&log),
        vec!["card guard prepared", "card guard committed"]
    );

    // A variant, a nested variant and a symbolic value, with the guard failing at prepared: one
    // call each, and no call at the aborted state git runs afterwards.
    w.git(&repo, &["pack-refs", "--all"]);
    w.env.push(("STANDIN_FAIL_AT".into(), "prepared".into()));
    for args in [
        vec!["branch", "integ"],
        vec!["branch", "INTEG/x"],
        vec!["symbolic-ref", "refs/heads/alias", "refs/heads/Integ"],
    ] {
        fs::write(&log, "").unwrap();
        let out = w.try_git(&repo, &args);
        assert!(!out.status.success(), "{args:?} passed");
        assert_eq!(calls(&log), vec!["card guard prepared"], "{args:?}");
    }

    // No moai anywhere: a transaction on the recorded branch refuses and names the cause, an
    // unrelated commit passes.
    w.standin = None;
    w.env.retain(|(k, _)| k != "STANDIN_FAIL_AT");
    fs::remove_dir_all(w.home.path().join(".moai")).ok();
    let out = w.try_git(&repo, &["update-ref", "refs/heads/Integ", &a]);
    assert!(!out.status.success());
    let message = String::from_utf8_lossy(&out.stderr).into_owned();
    assert!(
        message.contains("moai") && message.contains("not found"),
        "{message}"
    );
    w.git(
        &lane,
        &["commit", "-q", "--allow-empty", "-m", "more lane work"],
    );

    // No record: nothing starts and the transaction passes.
    let standin_dir = w.dir.path().join("standin");
    w.standin = Some(standin_dir);
    fs::remove_file(w.common().join(guard::RECORD_FILE)).unwrap();
    fs::write(&log, "").unwrap();
    w.git(&repo, &["update-ref", "refs/heads/Integ", &a]);
    assert!(calls(&log).is_empty(), "{:?}", calls(&log));
}

// ---- AC-9: the one reader of the name --------------------------------------------------------

fn profile_for(branch: &str) -> String {
    format!(
        "{{\"schema_version\": 1, \"project_id\": \"demo\", \"purpose\": \"fixture\", \"tools\": [\"rust\"], \"development_methods\": [\"sdd\"], \"hosts\": [\"claude\"], \"work_flow\": {{\"flow\": \"m-flow\", \"integration_branch\": \"{branch}\"}}}}"
    )
}

fn write_profile(project: &Path, branch: &str) {
    let cfg = project.join(".moai-cli/project.yaml");
    fs::create_dir_all(cfg.parent().unwrap()).unwrap();
    fs::write(cfg, profile_for(branch)).unwrap();
}

/// AC-9: a repository with no record takes the name from the profile once and writes the record,
/// a profile that later disagrees is reported and the record wins, and no record with no profile
/// value leaves nothing protected.
#[test]
fn ac9_the_record_is_the_one_value_and_the_profile_only_seeds_it() {
    let w = Repo::new(false);
    // No record and no profile value: nothing to protect, nothing written.
    assert_eq!(guard::protected(&w.repo, true), Ok(None));
    assert!(!w.common().join(guard::RECORD_FILE).exists());
    // A profile value seeds the record, once.
    write_profile(&w.repo, "Integ");
    let seeded = guard::protected(&w.repo, true).unwrap().unwrap();
    assert_eq!(seeded.full_ref, "refs/heads/Integ");
    assert!(seeded.written);
    assert_eq!(seeded.profile_disagrees, None);
    assert_eq!(
        fs::read_to_string(w.common().join(guard::RECORD_FILE)).unwrap(),
        "refs/heads/Integ\n"
    );
    // A profile that later disagrees is reported and the record wins; nothing is rewritten.
    write_profile(&w.repo, "Other");
    let again = guard::protected(&w.repo, true).unwrap().unwrap();
    assert_eq!(again.full_ref, "refs/heads/Integ");
    assert!(!again.written);
    assert_eq!(again.profile_disagrees.as_deref(), Some("Other"));
    // Reading without leave to write seeds nothing.
    let v = Repo::new(false);
    write_profile(&v.repo, "Integ");
    assert_eq!(guard::protected(&v.repo, false), Ok(None));
    assert!(!v.common().join(guard::RECORD_FILE).exists());
}

// ---- AC-10: the hook text and the aborted transaction ------------------------------------------

/// AC-10: the hook text is a shell script that holds the guard's command and no branch name and
/// reads the record, and a transaction the guard refuses as a whole leaves the intent record where
/// it was.
#[test]
fn ac10_the_hook_text_has_no_branch_name_and_an_aborted_transaction_keeps_the_record() {
    let text = guard::hook_text();
    assert!(text.starts_with("#!/bin/sh\n"), "{text}");
    assert!(text.contains("card guard"), "{text}");
    assert!(text.contains("moai/integration-branch"), "{text}");
    for name in ["integ", "main", "master", "next"] {
        assert!(
            !text.to_lowercase().contains(&format!("refs/heads/{name}")),
            "the hook text spells the branch {name}"
        );
    }
    let w = Repo::new(false);
    w.install();
    let (a, b) = (w.a.clone(), w.b.clone());
    w.write_intent("refs/heads/Integ", &a, &b);
    // One transaction: the allowed move of the integration branch and a variant's creation.
    let out = w.try_git_stdin(
        &w.repo,
        &["update-ref", "--stdin"],
        &format!("start\nupdate refs/heads/Integ {b} {a}\ncreate refs/heads/integ {b}\nprepare\ncommit\n"),
    );
    w.refused(&out);
    tree_clean(&w);
    assert!(
        w.intent_exists(),
        "an aborted transaction consumed the record"
    );
}

// ---- AC-12: the document and the help --------------------------------------------------------

/// AC-12: the card operations document names the `card guard` action, its states, the intent
/// record and the guard's refusals, and the help says that git runs it.
#[test]
fn ac12_the_document_and_the_help_name_the_guard() {
    let doc = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("docs/design-review/appendix/operations/card.md"),
    )
    .unwrap();
    for needle in [
        "card guard",
        "prepared",
        "committed",
        "moai/land-intent",
        "moai/integration-branch",
    ] {
        assert!(doc.contains(needle), "the card document lacks {needle}");
    }
    let home = TempHome::new();
    let out = home
        .command(&["card", "guard", "--help"])
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let help = String::from_utf8_lossy(&out.stdout).into_owned();
    assert!(help.contains("git runs"), "{help}");
}

// ---- AC-13: a worktree whose profile names another branch --------------------------------------

/// AC-13: in a linked worktree whose checked-out profile names another integration branch the
/// moves of the recorded branch still refuse, and the refusal reports that the profile disagrees.
#[test]
fn ac13_a_worktree_profile_that_names_another_branch_does_not_disarm_the_guard() {
    let w = Repo::new(false);
    w.install();
    write_profile(&w.lane_wt, "Old");
    let (a, b) = (w.a.clone(), w.b.clone());
    let out = w.try_git(&w.lane_wt, &["update-ref", "refs/heads/Integ", &b, &a]);
    let message = w.refused(&out);
    assert!(
        message.contains("profile") && message.contains("Old"),
        "{message}"
    );
    tree_clean(&w);
    let out = w.try_git(&w.lane_wt, &["branch", "-f", INTEG, "lane"]);
    w.refused(&out);
    tree_clean(&w);
}
