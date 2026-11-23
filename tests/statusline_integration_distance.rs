//! The lane branch is still measured against the integration branch, whatever upstream the
//! checked-out branch has configured itself (SPEC-20261123-statusline-integration-distance), but the
//! numbers are not on the screen since SPEC-20261123-statusline-order-bars REQ-11: the first line
//! holds the integration item with its tip and the checked-out branch, and no `🧭`, `↑`, `↓` or
//! `+N`. The cases run the real binary against a real repository in a temporary folder with the
//! real `git`.
#![cfg(unix)]

mod common;
use common::CleanEnv;

use std::{
    fs,
    io::Write as _,
    path::PathBuf,
    process::{Command, Stdio},
};

/// A budget no git call can miss under any load.
const WITHIN: u64 = 3_600_000;
/// The integration branch of the fixture.
const INTEGRATION: &str = "next";

struct Fixture {
    _guard: tempfile::TempDir,
    home: PathBuf,
    project: PathBuf,
}

impl Fixture {
    /// A repository on `next` with one commit, the project profile naming `next` as the
    /// integration branch. No remote, no cache folder: every render asks git.
    fn new() -> Fixture {
        let guard = tempfile::tempdir().unwrap();
        let root = fs::canonicalize(guard.path()).unwrap();
        let home = root.join("home");
        let project = root.join("project");
        fs::create_dir_all(home.join(moai_cli::home_root::MOAI_DIR)).unwrap();
        fs::create_dir_all(&project).unwrap();
        common::git::git_ok(&project, &["init", "-q", "-b", INTEGRATION]);
        let f = Fixture {
            _guard: guard,
            home,
            project,
        };
        f.commit("base");
        let state = moai_cli::state_root::path(&f.project, "");
        fs::create_dir_all(&state).unwrap();
        fs::write(
            state.join("project.yaml"),
            format!(
                "{{\"schema_version\":2,\"project_id\":\"demo\",\"work_flow\":{{\"flow\":\"m-flow\",\"integration_branch\":\"{INTEGRATION}\"}}}}\n"
            ),
        )
        .unwrap();
        f
    }

    fn git(&self, args: &[&str]) {
        common::git::git_in(&self.project, args);
    }

    /// One empty commit on the checked-out branch.
    fn commit(&self, message: &str) {
        self.git(&["commit", "-q", "--allow-empty", "-m", message]);
    }

    /// `branch` started at the current commit of `from` and checked out.
    fn branch_from(&self, branch: &str, from: &str) {
        self.git(&["checkout", "-q", "-b", branch, from]);
    }

    /// The branch `name` gets its own upstream configuration, the way a pushed lane branch has
    /// it: the remote `origin` and the branch of the same name on it. `refs/heads/<name>` is
    /// what `git push -u origin <name>` writes.
    fn own_upstream(&self, name: &str) {
        self.git(&["config", &format!("branch.{name}.remote"), "origin"]);
        self.git(&[
            "config",
            &format!("branch.{name}.merge"),
            &format!("refs/heads/{name}"),
        ]);
    }

    /// The lines of one render of the session in the project folder.
    fn render(&self) -> String {
        let stdin = serde_json::json!({
            "model": {"display_name": "Opus"},
            "workspace": {"current_dir": self.project, "project_dir": self.project},
        });
        let mut cmd = Command::new(common::MOAI_BIN);
        cmd.args(["statusline", "--render-budget-ms", &WITHIN.to_string()])
            .current_dir(&self.project)
            .clean_env()
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("HOME", &self.home)
            .env("MOAI_HOME", self.home.join(moai_cli::home_root::MOAI_DIR))
            .env("NO_COLOR", "1")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped());
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stdin.to_string().as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        assert!(out.status.success());
        String::from_utf8(out.stdout).unwrap()
    }
}

/// The first line of one render.
fn first_line(out: &str) -> &str {
    out.lines().next().unwrap_or("")
}

/// What the first line of the session is: the integration branch with the first seven characters
/// of its local ref (`git rev-parse --short=7`, or `n/a` when the ref is gone), then the
/// checked-out branch — and no `🧭` item, no `↑` or `↓` and no `+N`, however far the branch is
/// from the integration branch (SPEC-20261123-statusline-order-bars REQ-9 and REQ-11).
fn assert_first_line(f: &Fixture, out: &str, tip: Option<&str>, checked_out: &str) {
    let tip = match tip {
        Some(branch) => common::git::git(
            &f.project,
            &["rev-parse", "--short=7", &format!("refs/heads/{branch}")],
        ),
        None => "n/a".to_string(),
    };
    assert_eq!(
        first_line(out),
        format!("🧠 Opus | 🅱️ {INTEGRATION}(🏷️ {tip}) | 🌳 {checked_out}"),
        "{out}"
    );
    for mark in ["🧭", "↑", "↓", "+"] {
        assert!(
            !first_line(out).contains(mark),
            "{mark} on the first line: {out}"
        );
    }
}

/// Two lane commits, one commit that only the integration branch has: ahead 2, behind 1.
fn lane_two_ahead_one_behind(f: &Fixture, lane: &str) {
    // The lane starts at the current `next`, then `next` moves on by one commit.
    f.branch_from(lane, INTEGRATION);
    f.git(&["checkout", "-q", INTEGRATION]);
    f.commit("integration moved");
    f.git(&["checkout", "-q", lane]);
    f.commit("lane 1");
    f.commit("lane 2");
}

/// The card's case: a lane branch whose own upstream is set (`origin`, the branch of its own
/// name) is still measured against the integration branch, which the code keeps doing; the line
/// shows no distance (SPEC-20261123-statusline-order-bars REQ-11) and the integration tip is the
/// commit of `next`, not the one `HEAD` is at.
#[test]
fn a_branch_with_its_own_upstream_shows_the_integration_tip_and_no_distance() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    f.own_upstream("lane");
    let out = f.render();
    assert_first_line(&f, &out, Some(INTEGRATION), "lane");
}

/// The same lane without any own upstream shows the same line.
#[test]
fn a_branch_without_an_upstream_shows_the_same_line() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    let out = f.render();
    assert_first_line(&f, &out, Some(INTEGRATION), "lane");
}

/// An own upstream that names a different branch of another remote changes nothing either.
#[test]
fn an_upstream_that_names_another_branch_does_not_change_the_line() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    f.git(&["config", "branch.lane.remote", "origin"]);
    f.git(&["config", "branch.lane.merge", "refs/heads/somewhere-else"]);
    let out = f.render();
    assert_first_line(&f, &out, Some(INTEGRATION), "lane");
}

/// A branch whose own upstream is the integration branch itself on a remote shows the same line.
#[test]
fn an_upstream_that_names_the_integration_branch_does_not_change_the_line() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    f.git(&["config", "branch.lane.remote", "origin"]);
    f.git(&[
        "config",
        "branch.lane.merge",
        &format!("refs/heads/{INTEGRATION}"),
    ]);
    let out = f.render();
    assert_first_line(&f, &out, Some(INTEGRATION), "lane");
}

/// A branch level with the integration branch has no zero shown either: no `↑0 ↓0`.
#[test]
fn a_branch_level_with_the_integration_branch_shows_no_zero() {
    let f = Fixture::new();
    f.branch_from("lane", INTEGRATION);
    f.own_upstream("lane");
    let out = f.render();
    assert_first_line(&f, &out, Some(INTEGRATION), "lane");
}

/// The integration branch itself, checked out: both items name it, with the commit `HEAD` is at,
/// and no mark shows although its upstream configuration exists.
#[test]
fn the_integration_branch_itself_shows_both_items() {
    let f = Fixture::new();
    f.own_upstream(INTEGRATION);
    let out = f.render();
    assert_first_line(&f, &out, Some(INTEGRATION), INTEGRATION);
}

/// When the integration branch is not a local branch, nothing is measured and the commit part says
/// `n/a`: no guessed commit.
#[test]
fn an_integration_branch_that_is_missing_locally_says_n_a() {
    let f = Fixture::new();
    f.branch_from("lane", INTEGRATION);
    f.commit("lane 1");
    f.git(&["branch", "-q", "-m", INTEGRATION, "renamed"]);
    f.own_upstream("lane");
    let out = f.render();
    assert_first_line(&f, &out, None, "lane");
}
