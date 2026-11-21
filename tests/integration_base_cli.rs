//! `moai integration-base` (SPEC-20261121-integration-base-sha, slice S1): the exact commit of the
//! project's local integration branch, resolved from `refs/heads/<branch>` only, and a check of
//! whether a worktree holds it.
//!
//! Everything runs in a temporary HOME, a temporary repository and a local bare repository as the
//! remote: never the network, the real HOME or a real project.

mod common;

use common::home::TempHome;
use serde_json::{json, Value};
use std::fs;
use std::path::PathBuf;
use std::process::Output;

fn envelope(out: &Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|_| {
        panic!(
            "{} / {}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

fn ok(out: &Output) -> Value {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(0), "{v}");
    assert_eq!(v["operation"], json!("integration.base"), "{v}");
    v["result"].clone()
}

/// `(kind, message)` of a refusal.
fn refused(out: &Output) -> (String, String) {
    let v = envelope(out);
    assert_eq!(out.status.code(), Some(1), "{v}");
    assert_eq!(v["ok"], json!(false), "{v}");
    (
        v["error"]["kind"].as_str().unwrap_or_default().to_string(),
        v["error"]["message"]
            .as_str()
            .unwrap_or_default()
            .to_string(),
    )
}

struct World {
    home: TempHome,
    project: PathBuf,
}

impl World {
    /// A HOME and a repository `demo` on `main` with one commit.
    fn new() -> Self {
        let home = TempHome::new();
        let project = home.project("demo");
        let w = World { home, project };
        w.git(&["init", "-q", "-b", "main"]);
        fs::write(w.project.join("a.txt"), "a\n").unwrap();
        w.git(&["add", "-A"]);
        w.commit("init");
        w
    }

    fn p(&self) -> &str {
        self.project.to_str().unwrap()
    }

    fn git(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    fn commit(&self, message: &str) {
        commit_in(&self.project, message);
    }

    /// A committed profile with `flow` and, when given, the integration branch.
    fn profile(&self, flow: &str, branch: Option<&str>) {
        let work_flow = match branch {
            Some(b) => format!("{{\"flow\":\"{flow}\",\"integration_branch\":\"{b}\"}}"),
            None => format!("{{\"flow\":\"{flow}\"}}"),
        };
        let body = format!(
            "{{\"schema_version\":2,\"project_id\":\"demo\",\"purpose\":\"Fixture\",\
             \"tools\":[\"Rust\"],\"development_methods\":[\"TDD\"],\"hosts\":[\"claude\"],\
             \"decisions\":[{{\"id\":\"d1\",\"topic\":\"tools\",\"value\":\"Rust\",\
             \"source\":\"user\",\"scope\":\"project\"}}],\"work_flow\":{work_flow}}}"
        );
        fs::write(self.project.join(".moai-cli").join("project.yaml"), body).unwrap();
        self.git(&["add", "-A"]);
        self.commit("profile");
    }

    fn run(&self, extra: &[&str]) -> Output {
        let mut args = vec!["integration-base", self.p()];
        args.extend_from_slice(extra);
        self.home.command(&args).output().expect("run moai")
    }

    /// A bare remote `origin` that has moved on without this repository: the local `main` has
    /// one commit the remote lacks and `origin/main` one the local branch lacks. Returns
    /// `(local, remote)`, both from `git rev-parse`.
    fn diverge(&self) -> (String, String) {
        let bare = self.home.path().join("origin.git");
        let bare_s = bare.to_str().unwrap();
        common::git::git_ok(
            self.home.path(),
            &["init", "-q", "--bare", "-b", "main", bare_s],
        );
        self.git(&["remote", "add", "origin", bare_s]);
        self.git(&["push", "-q", "origin", "main"]);
        fs::write(self.project.join("local.txt"), "l\n").unwrap();
        self.git(&["add", "-A"]);
        self.commit("local only");
        let local = self.git(&["rev-parse", "HEAD"]);
        let other = self.home.path().join("other");
        common::git::git_ok(
            self.home.path(),
            &["clone", "-q", bare_s, other.to_str().unwrap()],
        );
        fs::write(other.join("remote.txt"), "r\n").unwrap();
        common::git::git_ok(&other, &["add", "-A"]);
        commit_in(&other, "remote only");
        common::git::git_ok(&other, &["push", "-q", "origin", "main"]);
        self.git(&["fetch", "-q", "origin"]);
        let remote = self.git(&["rev-parse", "origin/main"]);
        assert_ne!(local, remote);
        (local, remote)
    }
}

fn commit_in(dir: &std::path::Path, message: &str) {
    common::git::git_ok(
        dir,
        &[
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
            "-c",
            "commit.gpgsign=false",
            "commit",
            "-q",
            "-m",
            message,
        ],
    );
}

#[test]
fn a_diverged_local_branch_resolves_to_the_local_commit() {
    let w = World::new();
    let (_, remote) = w.diverge();
    // The profile commit moves the local branch, so the expected commit is read after it.
    w.profile("m-flow", Some("main"));
    let local = w.git(&["rev-parse", "refs/heads/main"]);
    let r = ok(&w.run(&[]));
    assert_eq!(r["policy"], json!("m-flow"), "{r}");
    assert_eq!(r["branch"], json!("main"), "{r}");
    assert_eq!(r["source"], json!("local"), "{r}");
    assert_eq!(r["sha"], json!(local), "{r}");
    assert_ne!(r["sha"], json!(remote), "{r}");
    assert_eq!(r["writes"], json!(false), "{r}");
}

#[test]
fn a_missing_local_branch_stops_with_the_next_step_and_never_reads_origin() {
    let w = World::new();
    w.diverge();
    // The integration branch exists on the remote only: `origin/m-flow/integration`.
    w.git(&["branch", "m-flow/integration"]);
    w.git(&["push", "-q", "origin", "m-flow/integration"]);
    w.git(&["branch", "-q", "-D", "m-flow/integration"]);
    let remote_only = w.git(&["rev-parse", "origin/m-flow/integration"]);
    w.profile("m-flow", Some("m-flow/integration"));
    let out = w.run(&[]);
    let (kind, message) = refused(&out);
    assert_eq!(kind, "not_found", "{message}");
    assert!(message.contains("m-flow/integration"), "{message}");
    assert!(message.contains("git branch"), "the next step: {message}");
    assert!(
        message.contains("origin"),
        "it says origin is not used: {message}"
    );
    let all = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(
        !all.contains(&remote_only),
        "no remote commit id is printed: {all}"
    );
}

#[test]
fn a_tag_or_remote_ref_with_the_branch_name_never_stands_in_for_the_branch() {
    let w = World::new();
    // A tag called `rel` on one commit, the branch `rel` on another.
    w.git(&["tag", "rel"]);
    fs::write(w.project.join("b.txt"), "b\n").unwrap();
    w.git(&["add", "-A"]);
    w.commit("ahead");
    w.git(&["branch", "rel"]);
    let branch_sha = w.git(&["rev-parse", "refs/heads/rel"]);
    let tag_sha = w.git(&["rev-parse", "refs/tags/rel"]);
    assert_ne!(branch_sha, tag_sha);
    w.profile("m-flow", Some("rel"));
    let r = ok(&w.run(&[]));
    assert_eq!(r["sha"], json!(branch_sha), "{r}");

    // With only the tag left, the branch is missing: the tag is not a fallback.
    w.git(&["branch", "-q", "-D", "rel"]);
    let (kind, message) = refused(&w.run(&[]));
    assert_eq!(kind, "not_found", "{message}");
}

#[test]
fn another_flow_no_branch_and_no_profile_have_no_policy() {
    let w = World::new();
    // No profile beyond the stub the project folder starts with.
    let r = ok(&w.run(&[]));
    assert_eq!(r["policy"], json!("none"), "{r}");
    assert_eq!(r["sha"], Value::Null, "{r}");
    assert_eq!(r["reason"], json!("profile_unreadable"), "{r}");

    w.profile("github-flow", Some("main"));
    let r = ok(&w.run(&[]));
    assert_eq!(r["policy"], json!("none"), "{r}");
    assert_eq!(r["sha"], Value::Null, "{r}");
    assert_eq!(r["reason"], json!("flow_not_m_flow"), "{r}");

    w.profile("m-flow", None);
    let r = ok(&w.run(&[]));
    assert_eq!(r["policy"], json!("none"), "{r}");
    assert_eq!(r["reason"], json!("no_integration_branch"), "{r}");
}

#[test]
fn the_command_writes_nothing() {
    let w = World::new();
    w.diverge();
    w.profile("m-flow", Some("main"));
    let before = w.home.snapshot();
    ok(&w.run(&[]));
    assert_eq!(
        w.home.snapshot(),
        before,
        "no file, folder, size or time changed"
    );
    // A refusal writes nothing either.
    w.profile("m-flow", Some("not-there"));
    let before = w.home.snapshot();
    refused(&w.run(&[]));
    assert_eq!(w.home.snapshot(), before);
}

#[test]
fn a_worktree_from_the_local_commit_holds_it_and_one_from_origin_does_not() {
    let w = World::new();
    let (_, remote) = w.diverge();
    w.profile("m-flow", Some("main"));
    let local = w.git(&["rev-parse", "refs/heads/main"]);

    // A lane worktree started from the local commit, with that commit recorded for its branch.
    let good = w.home.path().join("wt-local");
    w.git(&[
        "worktree",
        "add",
        "-q",
        good.to_str().unwrap(),
        "-b",
        "lane-a",
        &local,
    ]);
    w.git(&["config", "branch.lane-a.moai-base-sha", &local]);
    let r = ok(&w.run(&["--worktree", good.to_str().unwrap()]));
    let wt = &r["worktree"];
    assert_eq!(wt["head"], json!(local), "{r}");
    assert_eq!(wt["contains_base"], json!(true), "{r}");
    assert_eq!(wt["recorded_sha"], json!(local), "{r}");
    assert_eq!(wt["matches_recorded"], json!(true), "{r}");

    // A lane worktree started from the diverged `origin/main`: it lacks the local commit.
    let bad = w.home.path().join("wt-origin");
    w.git(&[
        "worktree",
        "add",
        "-q",
        bad.to_str().unwrap(),
        "-b",
        "lane-b",
        "origin/main",
    ]);
    let r = ok(&w.run(&["--worktree", bad.to_str().unwrap()]));
    let wt = &r["worktree"];
    assert_eq!(wt["head"], json!(remote), "{r}");
    assert_eq!(wt["contains_base"], json!(false), "{r}");
    assert_eq!(wt["recorded_sha"], Value::Null, "{r}");
    assert_eq!(wt["matches_recorded"], json!(false), "{r}");

    // A recorded commit that is not the local integration commit does not match.
    w.git(&["config", "branch.lane-b.moai-base-sha", &remote]);
    let r = ok(&w.run(&["--worktree", bad.to_str().unwrap()]));
    assert_eq!(r["worktree"]["recorded_sha"], json!(remote), "{r}");
    assert_eq!(r["worktree"]["matches_recorded"], json!(false), "{r}");
}

#[test]
fn a_worktree_check_without_a_policy_or_for_a_bad_path_says_so() {
    let w = World::new();
    // No m-flow policy: there is no commit to compare with, and the worktree part says why.
    let r = ok(&w.run(&["--worktree", w.p()]));
    assert_eq!(r["policy"], json!("none"), "{r}");
    assert_eq!(r["worktree"], Value::Null, "{r}");

    w.profile("m-flow", Some("main"));
    let missing = w.home.path().join("not-a-worktree");
    let (kind, message) = refused(&w.run(&["--worktree", missing.to_str().unwrap()]));
    assert_eq!(kind, "invalid_input", "{message}");
    assert!(message.contains("integration-base refused"), "{message}");
}
