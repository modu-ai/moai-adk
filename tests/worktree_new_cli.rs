//! SPEC-20261124-lane-worktree-command: `moai worktree-new <name> --branch <branch>` makes a
//! lane's worktree at `<primary checkout>/.claude/worktrees/<name>` on a new branch cut from the
//! integration branch, reads `git worktree list` again and prints the absolute path. A name
//! that is not one plain folder name, a taken folder or branch, an unknown base, a link in the
//! place and a read-back that does not match are refused or reported; nothing is ever deleted.
//! Every command runs on a temporary repository in a clean environment.
#![cfg(unix)]

mod common;

use common::git::{git, git_ok, git_raw};
use common::home::TempHome;
use serde_json::Value;
use std::fs;
use std::path::{Path, PathBuf};

const BRANCH: &str = "spec/x-20261124";
/// A profile whose integration branch is `main`.
const PROFILE: &str = r#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"], "work_flow": {"flow": "next-flow", "integration_branch": "main"}}"#;
/// The same project with no work_flow: no integration branch to cut from.
const PLAIN: &str = r#"{"schema_version": 1, "project_id": "demo", "purpose": "fixture", "tools": ["rust"], "development_methods": ["sdd"], "hosts": ["claude"]}"#;

struct Repo {
    dir: tempfile::TempDir,
    home: TempHome,
    root: PathBuf,
    /// The tip of `main`, the integration branch.
    tip: String,
}

/// What one `worktree-new` run answered: the exit code, the JSON on stdout when there is any and
/// everything it said.
struct Out {
    code: i32,
    json: Value,
    said: String,
}

fn commit(root: &Path, message: &str) {
    git_ok(root, &["add", "-A"]);
    git_ok(
        root,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            message,
        ],
    );
}

/// A repository on `main` with the profile `profile`, then checked out on `side` one commit
/// further, so the checked-out HEAD is not the integration tip.
fn repo(profile: &str) -> Repo {
    let dir = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(dir.path()).unwrap().join("repo");
    let state = root.join(moai_cli::PROJECT_STATE_DIR);
    fs::create_dir_all(&state).unwrap();
    git_ok(&root, &["init", "-q", "-b", "main"]);
    fs::write(state.join("project.yaml"), profile).unwrap();
    fs::write(root.join(".gitignore"), "/target/\n/.claude/\n").unwrap();
    fs::write(root.join("one.txt"), "one\n").unwrap();
    commit(&root, "one");
    let tip = git(&root, &["rev-parse", "main"]);
    git_ok(&root, &["checkout", "-q", "-b", "side"]);
    fs::write(root.join("two.txt"), "two\n").unwrap();
    commit(&root, "two");
    Repo {
        dir,
        home: TempHome::new(),
        root,
        tip,
    }
}

impl Repo {
    fn place(&self) -> PathBuf {
        self.root.join(".claude/worktrees")
    }

    fn run_in(&self, cwd: &Path, args: &[&str]) -> Out {
        let out = self
            .home
            .command(&["worktree-new"])
            .args(args)
            .current_dir(cwd)
            .output()
            .expect("run moai");
        self.out_of(out)
    }

    fn run(&self, args: &[&str]) -> Out {
        self.run_in(&self.root, args)
    }

    fn out_of(&self, out: std::process::Output) -> Out {
        let stdout = String::from_utf8_lossy(&out.stdout).to_string();
        let stderr = String::from_utf8_lossy(&out.stderr).to_string();
        Out {
            code: out.status.code().unwrap_or(-1),
            json: serde_json::from_str(&stdout).unwrap_or(Value::Null),
            said: format!("{stdout}{stderr}"),
        }
    }

    /// The branches, the registered worktrees and the folders of the place: what a refused run
    /// must leave as it found them.
    fn state(&self) -> (String, String, Vec<String>) {
        let branches = git(&self.root, &["branch", "--format=%(refname:short)"]);
        let listed = git(&self.root, &["worktree", "list", "--porcelain"]);
        let mut folders: Vec<String> = fs::read_dir(self.place())
            .map(|d| {
                d.map(|e| e.unwrap().file_name().to_string_lossy().to_string())
                    .collect()
            })
            .unwrap_or_default();
        folders.sort();
        (branches, listed, folders)
    }

    /// A first run that must succeed, so a refusal test cannot pass because the command is
    /// missing.
    fn control(&self) -> Out {
        let out = self.run(&["lane-ok", "--branch", "spec/ok-20261124"]);
        assert_eq!(out.code, 0, "the control run failed: {}", out.said);
        out
    }
}

fn path_of(out: &Out) -> String {
    out.json["result"]["path"]
        .as_str()
        .unwrap_or_else(|| panic!("the answer has no path: {}", out.said))
        .to_string()
}

/// REQ-1, REQ-2: the worktree is made in the primary's place on a new branch at the integration
/// tip (not at the HEAD that is checked out), the absolute path is printed and git lists it.
#[test]
fn ac1_the_worktree_is_made_on_a_new_branch_at_the_integration_tip() {
    let r = repo(PROFILE);
    assert_ne!(git(&r.root, &["rev-parse", "HEAD"]), r.tip);
    let out = r.run(&["lane-a", "--branch", BRANCH]);
    assert_eq!(out.code, 0, "{}", out.said);
    let want = r.place().join("lane-a");
    let path = path_of(&out);
    assert_eq!(path, want.to_str().unwrap());
    assert!(Path::new(&path).is_absolute() && want.is_dir());
    assert_eq!(out.json["result"]["branch"], BRANCH);
    assert_eq!(out.json["result"]["base"], r.tip.as_str());
    assert_eq!(git(&want, &["rev-parse", "HEAD"]), r.tip);
    assert_eq!(
        git(&want, &["rev-parse", "--abbrev-ref", "HEAD"]),
        BRANCH,
        "the worktree is on the new branch"
    );
    let listed = git(&r.root, &["worktree", "list", "--porcelain"]);
    assert!(listed.contains(&format!("worktree {path}")), "{listed}");
    assert!(
        listed.contains(&format!("branch refs/heads/{BRANCH}")),
        "{listed}"
    );
    // The base can be named: a commit other than the integration tip.
    let head = git(&r.root, &["rev-parse", "HEAD"]);
    let second = r.run(&["lane-b", "--branch", "spec/y-20261124", "--base", "side"]);
    assert_eq!(second.code, 0, "{}", second.said);
    assert_eq!(second.json["result"]["base"], head.as_str());
    assert_eq!(git(&r.place().join("lane-b"), &["rev-parse", "HEAD"]), head);
}

/// REQ-3: a name that is not one plain folder name is refused and nothing is made.
#[test]
fn ac2_a_name_that_is_not_one_plain_folder_name_is_refused() {
    let r = repo(PROFILE);
    r.control();
    let before = r.state();
    let long = "a".repeat(65);
    let names = [
        "a/b",
        "../x",
        "..",
        ".",
        ".hidden",
        "a..b",
        "A",
        "-a",
        "a b",
        "a_b",
        "",
        "/tmp/abs",
        "lane-a/",
        long.as_str(),
    ];
    for name in names {
        let out = r.run(&["--branch", BRANCH, "--", name]);
        assert_ne!(out.code, 0, "the name {name:?} was accepted: {}", out.said);
        assert!(
            !out.said.contains("\"path\""),
            "the name {name:?} printed a path: {}",
            out.said
        );
        assert_eq!(
            r.state(),
            before,
            "the name {name:?} changed the repository"
        );
    }
    // A plain name of the largest length is the control of the length rule.
    let longest = "a".repeat(64);
    let ok = r.run(&[longest.as_str(), "--branch", BRANCH]);
    assert_eq!(ok.code, 0, "{}", ok.said);
}

/// REQ-4: a taken folder or branch, an unknown base, a place that is a link and no way to know a
/// base are refused and change nothing.
#[test]
fn ac3_a_taken_name_an_unknown_base_and_a_linked_place_are_refused() {
    let r = repo(PROFILE);
    r.control();
    // The folder is taken, by anything.
    fs::create_dir_all(r.place().join("lane-taken")).unwrap();
    let before = r.state();
    let out = r.run(&["lane-taken", "--branch", BRANCH]);
    assert_ne!(out.code, 0, "{}", out.said);
    assert!(out.said.contains("lane-taken"), "{}", out.said);
    assert_eq!(
        out.json["error"]["kind"], "invalid_input",
        "a refusal is not an internal error: {}",
        out.said
    );
    // The branch is taken.
    git_ok(&r.root, &["branch", "spec/taken-20261124", "main"]);
    let before_branch = r.state();
    let out = r.run(&["lane-b", "--branch", "spec/taken-20261124"]);
    assert_ne!(out.code, 0, "{}", out.said);
    assert!(out.said.contains("spec/taken-20261124"), "{}", out.said);
    assert_eq!(r.state(), before_branch);
    assert_ne!(before, before_branch, "the control of the fixture");
    // The base names no commit.
    let out = r.run(&["lane-c", "--branch", BRANCH, "--base", "no-such-ref"]);
    assert_ne!(out.code, 0, "{}", out.said);
    assert!(out.said.contains("no-such-ref"), "{}", out.said);
    assert_eq!(r.state(), before_branch);
    // The branch name is not a branch name, or is spelled as a full reference.
    for bad in ["bad..name", "refs/heads/spec/full-20261124"] {
        let out = r.run(&["lane-d", "--branch", bad]);
        assert_ne!(out.code, 0, "{bad}: {}", out.said);
        assert_eq!(r.state(), before_branch, "{bad}");
    }
    // No integration branch and no --base: nothing to cut from; with --base it works.
    let plain = repo(PLAIN);
    let out = plain.run(&["lane-a", "--branch", BRANCH]);
    assert_ne!(out.code, 0, "{}", out.said);
    assert!(out.said.contains("integration"), "{}", out.said);
    assert!(!plain.place().join("lane-a").exists());
    let ok = plain.run(&["lane-a", "--branch", BRANCH, "--base", "main"]);
    assert_eq!(ok.code, 0, "{}", ok.said);
    // A link at the place: refused, and nothing appears behind it.
    let linked = repo(PROFILE);
    let elsewhere = linked.dir.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::create_dir_all(linked.root.join(".claude")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, linked.place()).unwrap();
    let out = linked.run(&["lane-a", "--branch", BRANCH]);
    assert_ne!(out.code, 0, "{}", out.said);
    assert!(out.said.contains("link"), "{}", out.said);
    assert_eq!(fs::read_dir(&elsewhere).unwrap().count(), 0);
    assert!(!git(&linked.root, &["branch", "--list", BRANCH]).contains(BRANCH));
}

/// REQ-5: run from inside a linked worktree, the new worktree is still in the primary's place.
#[test]
fn ac4_run_from_a_linked_worktree_the_new_one_is_in_the_primary_place() {
    let r = repo(PROFILE);
    let first = r.run(&["lane-a", "--branch", BRANCH]);
    assert_eq!(first.code, 0, "{}", first.said);
    let inside = r.place().join("lane-a");
    let out = r.run_in(&inside, &["lane-c", "--branch", "spec/z-20261124"]);
    assert_eq!(out.code, 0, "{}", out.said);
    let want = r.place().join("lane-c");
    assert_eq!(path_of(&out), want.to_str().unwrap());
    assert!(want.is_dir());
    assert!(
        !inside.join(".claude").exists(),
        "a folder was made below the worktree the command ran in"
    );
    let listed = git(&r.root, &["worktree", "list", "--porcelain"]);
    assert!(
        listed.contains(&format!("worktree {}", want.display())),
        "{listed}"
    );
}

/// The first `git` on the real PATH.
fn real_git() -> PathBuf {
    std::env::var_os("PATH")
        .into_iter()
        .flat_map(|p| std::env::split_paths(&p).collect::<Vec<_>>())
        .map(|d| d.join("git"))
        .find(|p| p.is_file())
        .expect("a git on PATH")
}

/// REQ-2, REQ-8: a list that does not hold the new worktree on the new branch is a failure that
/// names the path; the folder stays where it is and nothing is deleted. The throwaway `git`
/// shim answers the list that follows the add with another branch.
#[test]
fn ac5_a_read_back_that_does_not_match_fails_and_names_the_path() {
    let r = repo(PROFILE);
    let shim = r.dir.path().join("shim");
    fs::create_dir_all(&shim).unwrap();
    let mark = shim.join("added");
    common::write_executable(
        &shim.join("git"),
        &format!(
            "#!/bin/sh\ncase \"$*\" in\n  *\"worktree add\"*) \"{real}\" \"$@\"; rc=$?; : > \"{mark}\"; exit $rc ;;\n  *\"worktree list\"*) if [ -e \"{mark}\" ]; then \"{real}\" \"$@\" | sed 's#^branch refs/heads/.*#branch refs/heads/elsewhere#'; exit 0; fi ;;\nesac\nexec \"{real}\" \"$@\"\n",
            real = real_git().display(),
            mark = mark.display()
        ),
    );
    let out = r.out_of(
        r.home
            .command(&["worktree-new"])
            .args(["lane-a", "--branch", BRANCH])
            .env("PATH", format!("{}:/usr/bin:/bin", shim.display()))
            .current_dir(&r.root)
            .output()
            .expect("run moai"),
    );
    let want = r.place().join("lane-a");
    assert_ne!(
        out.code, 0,
        "a mismatching read-back was accepted: {}",
        out.said
    );
    assert!(out.said.contains(want.to_str().unwrap()), "{}", out.said);
    assert!(
        !out.said.contains("\"path\""),
        "a failed read-back printed a path: {}",
        out.said
    );
    assert!(want.is_dir(), "the folder was removed");
    let listed = git(&r.root, &["worktree", "list", "--porcelain"]);
    assert!(
        listed.contains(&format!("worktree {}", want.display())),
        "{listed}"
    );
    assert!(git_raw(&r.root, &["rev-parse", "--verify", BRANCH]).success);
}

/// REQ-6: the help names the options the command has and none that forces, deletes, prunes or
/// moves.
#[test]
fn ac6_the_help_has_no_option_that_forces_deletes_prunes_or_moves() {
    let home = TempHome::new();
    let out = home
        .command(&["worktree-new", "--help"])
        .output()
        .expect("run moai");
    assert!(out.status.success(), "`worktree-new --help` failed");
    let help = String::from_utf8_lossy(&out.stdout).to_string();
    for want in ["--branch", "--base", "--project"] {
        assert!(help.contains(want), "{want} is missing: {help}");
    }
    for never in ["--force", "--remove", "--delete", "--prune", "--move"] {
        assert!(!help.contains(never), "{never} is in the help: {help}");
    }
}

/// REQ-7: work.md names the command where it makes a card's worktree, in the numbered step and
/// in the paragraph that gives the order.
#[test]
fn ac7_work_md_names_the_command_in_both_worktree_places() {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/work.md"),
    )
    .unwrap();
    let step = text
        .lines()
        .find(|l| l.starts_with("2. **worktree**"))
        .expect("the worktree step of work.md");
    assert!(step.contains("moai worktree-new"), "{step}");
    let paragraph = text
        .lines()
        .find(|l| l.starts_with("Make a card's worktree"))
        .expect("the paragraph that makes a card's worktree");
    assert!(paragraph.contains("moai worktree-new"), "{paragraph}");
}
