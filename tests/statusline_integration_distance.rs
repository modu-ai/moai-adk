//! The `🧭` item and the `↑↓` marks of the branch item are measured against the integration
//! branch, whatever upstream the checked-out branch has configured itself. The cases run the real
//! binary against a real repository in a temporary folder with the real `git`.
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

/// The text of the `🧭` item (from the icon to the next ` | ` or the end of its line).
fn distance_item(out: &str) -> Option<&str> {
    let line = out.lines().find(|l| l.contains("🧭"))?;
    let from = line.find("🧭")?;
    let rest = &line[from..];
    Some(rest.split(" | ").next().unwrap_or(rest))
}

/// The line that carries the branch item.
fn branch_item_line(out: &str) -> &str {
    out.lines().find(|l| l.contains("🅱️")).unwrap_or("")
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
/// name) is still measured against the integration branch: `↑2 ↓1`, not `↑0 ↓0`.
#[test]
fn a_branch_with_its_own_upstream_is_measured_against_the_integration_branch() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    f.own_upstream("lane");
    let out = f.render();
    assert_eq!(
        distance_item(&out),
        Some("🧭 next ↑2 ↓1"),
        "the distance item must count against `{INTEGRATION}`: {out}"
    );
    let line = branch_item_line(&out);
    assert!(line.contains("lane"), "{out}");
    assert!(
        line.contains("↑2"),
        "the branch item's ahead mark uses the same count: {out}"
    );
}

/// The same lane without any own upstream shows the same numbers (the case that already worked).
#[test]
fn a_branch_without_an_upstream_is_measured_against_the_integration_branch() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    let out = f.render();
    assert_eq!(distance_item(&out), Some("🧭 next ↑2 ↓1"), "{out}");
    assert!(branch_item_line(&out).contains("↑2"), "{out}");
}

/// An own upstream that names a different branch of another remote changes nothing either.
#[test]
fn an_upstream_that_names_another_branch_does_not_change_the_distance() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    f.git(&["config", "branch.lane.remote", "origin"]);
    f.git(&["config", "branch.lane.merge", "refs/heads/somewhere-else"]);
    let out = f.render();
    assert_eq!(distance_item(&out), Some("🧭 next ↑2 ↓1"), "{out}");
}

/// A branch whose own upstream is the integration branch itself on a remote stays measured
/// against the local integration branch.
#[test]
fn an_upstream_that_names_the_integration_branch_keeps_the_local_numbers() {
    let f = Fixture::new();
    lane_two_ahead_one_behind(&f, "lane");
    f.git(&["config", "branch.lane.remote", "origin"]);
    f.git(&[
        "config",
        "branch.lane.merge",
        &format!("refs/heads/{INTEGRATION}"),
    ]);
    let out = f.render();
    assert_eq!(distance_item(&out), Some("🧭 next ↑2 ↓1"), "{out}");
}

/// A branch level with the integration branch shows `↑0 ↓0` — a real zero, not a failure.
#[test]
fn a_branch_level_with_the_integration_branch_shows_a_real_zero() {
    let f = Fixture::new();
    f.branch_from("lane", INTEGRATION);
    f.own_upstream("lane");
    let out = f.render();
    assert_eq!(distance_item(&out), Some("🧭 next ↑0 ↓0"), "{out}");
}

/// The integration branch itself has no `🧭` item and shows git's own numbers: with an upstream
/// that is behind by one, `↑1` on the branch item is what it has not pushed.
#[test]
fn the_integration_branch_itself_keeps_its_own_display() {
    let f = Fixture::new();
    f.own_upstream(INTEGRATION);
    let out = f.render();
    assert!(
        distance_item(&out).is_none(),
        "no distance item on the integration branch: {out}"
    );
    assert!(branch_item_line(&out).contains(INTEGRATION), "{out}");
    // No remote exists, so git's own numbers are zeros and no mark shows.
    assert!(!branch_item_line(&out).contains('↑'), "{out}");
}

/// When the integration branch is not a local branch, nothing is measured: `↑0 ↓0`, as before.
#[test]
fn an_integration_branch_that_is_missing_locally_measures_nothing() {
    let f = Fixture::new();
    f.branch_from("lane", INTEGRATION);
    f.commit("lane 1");
    f.git(&["branch", "-q", "-m", INTEGRATION, "renamed"]);
    f.own_upstream("lane");
    let out = f.render();
    assert_eq!(distance_item(&out), Some("🧭 next ↑0 ↓0"), "{out}");
}
