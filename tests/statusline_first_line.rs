//! The first line of the status line (SPEC-20261123-statusline-order-bars REQ-8 to REQ-11, AC-10
//! to AC-14): `🧠 <model> | 📡 <repository>, ⬇️ <issues> 🔀 <requests> | 🅱️ <integration>(🏷️ <tip>) |
//! 🌳 <checked-out branch> | 🔗 <request>`. The integration item is the configured local
//! integration branch with the first seven characters of the commit its local ref names; the
//! checked-out item is the branch the `HEAD` file names, and a detached `HEAD` leaves it out.
//!
//! Every case is a real temporary git repository (a temporary HOME and project, the shared git
//! runner that scrubs the git environment) and the real binary, with commits made so that `HEAD`
//! is at another commit than the integration tip. The expected seven characters come from
//! `git rev-parse --short=7 refs/heads/<integration>` inside the test, never from a typed value.
#![cfg(unix)]

mod common;

use common::home::TempHome;
use std::io::Write;
use std::path::PathBuf;
use std::process::Stdio;

/// The integration branch the project profile names.
const INTEGRATION: &str = "v2026.11.0";
/// The lane branch of the fixtures.
const LANE: &str = "spec/ready-mutant-line-20261123";
/// The forge counts file the status line reads (`1` issue, `0` requests, fresh for good).
const FORGE_CACHE: &str = "state/github/counts.json";

struct Repo {
    home: TempHome,
    project: PathBuf,
}

impl Repo {
    /// A project whose profile names [`INTEGRATION`], a repository on that branch with one commit
    /// (`base`), and forge counts of one issue and no request.
    fn new() -> Repo {
        let home = TempHome::new();
        let project = home.project("demo");
        common::git::git_in(&project, &["init", "-q", "-b", INTEGRATION]);
        let repo = Repo { home, project };
        repo.commit("base");
        std::fs::write(
            moai_cli::state_root::path(&repo.project, "project.yaml"),
            format!(
                "{{\"schema_version\":2,\"project_id\":\"demo\",\"work_flow\":{{\"flow\":\"m-flow\",\"integration_branch\":\"{INTEGRATION}\"}}}}\n"
            ),
        )
        .unwrap();
        let forge_cache = moai_cli::state_root::path(&repo.project, FORGE_CACHE);
        std::fs::create_dir_all(forge_cache.parent().unwrap()).unwrap();
        std::fs::write(
            forge_cache,
            "{\"open_issues\":1,\"open_prs\":0,\"fetched_at\":4102444800}\n",
        )
        .unwrap();
        repo
    }

    fn git(&self, args: &[&str]) {
        common::git::git_in(&self.project, args);
    }

    fn git_out(&self, args: &[&str]) -> String {
        common::git::git(&self.project, args)
    }

    fn commit(&self, message: &str) {
        self.git(&["commit", "-q", "--allow-empty", "-m", message]);
    }

    /// The first seven characters of the commit the local integration ref names.
    fn tip(&self) -> String {
        self.git_out(&[
            "rev-parse",
            "--short=7",
            &format!("refs/heads/{INTEGRATION}"),
        ])
    }

    /// The first seven characters of the commit `HEAD` is at.
    fn head(&self) -> String {
        self.git_out(&["rev-parse", "--short=7", "HEAD"])
    }

    /// The lane branch from the integration branch, checked out, two commits ahead of it.
    fn lane_two_ahead(&self) {
        self.git(&["checkout", "-q", "-b", LANE, INTEGRATION]);
        self.commit("lane 1");
        self.commit("lane 2");
    }

    /// One plain render of the session in the project folder: the first line. `request` adds the
    /// pull request the host reports; `columns` is the `COLUMNS` of the render, when set.
    fn first_line(&self, request: bool, columns: Option<&str>) -> String {
        let mut stdin = serde_json::json!({
            "session_id": "s-first-line",
            "cwd": self.project,
            "model": {"display_name": "Sonnet 5.5"},
            "effort": {"level": "high"},
            "thinking": {"enabled": true},
            "workspace": {
                "current_dir": self.project,
                "project_dir": self.project,
                "repo": {"host": "github.com", "owner": "modu-ai", "name": "moai-adk"},
            },
        });
        if request {
            stdin["pr"] = serde_json::json!({"number": 1234, "review_state": "pending"});
        }
        let mut cmd = self
            .home
            .command(&["statusline", "--render-budget-ms", "3600000"]);
        cmd.current_dir(&self.project)
            .env("PATH", std::env::var_os("PATH").unwrap_or_default())
            .env("TMPDIR", self.home.path().join("tmp"))
            .env("NO_COLOR", "1")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped());
        std::fs::create_dir_all(self.home.path().join("tmp")).unwrap();
        if let Some(columns) = columns {
            cmd.env("COLUMNS", columns);
        }
        let mut child = cmd.spawn().unwrap();
        child
            .stdin
            .take()
            .unwrap()
            .write_all(stdin.to_string().as_bytes())
            .unwrap();
        let out = child.wait_with_output().unwrap();
        assert_eq!(
            out.status.code(),
            Some(0),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        String::from_utf8(out.stdout)
            .unwrap()
            .lines()
            .next()
            .unwrap_or("")
            .to_string()
    }
}

/// The first line of a lane session: the model, the repository with its forge counts, the
/// integration item with `tip`, the checked-out `branch` and the request.
fn lane_line(tip: &str, branch: &str) -> String {
    format!("🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ {INTEGRATION}(🏷️ {tip}) | 🌳 {branch} | 🔗 #1234 pending")
}

/// AC-10: on a lane branch two commits ahead of the integration branch the first line names the
/// integration branch with the tip of its ref (not the commit `HEAD` is at), then the lane
/// branch, then the request — and no `🧭`, `↑`, `↓` or `+N`.
#[test]
fn the_lane_branch_shows_the_integration_tip_and_the_checked_out_branch() {
    let repo = Repo::new();
    repo.lane_two_ahead();
    let tip = repo.tip();
    assert_ne!(tip, repo.head(), "HEAD is at another commit than the tip");
    let line = repo.first_line(true, None);
    assert_eq!(line, lane_line(&tip, LANE), "{line}");
    for mark in ["🧭", "↑", "↓", "+"] {
        assert!(!line.contains(mark), "{mark} in {line}");
    }
}

/// AC-11: a detached `HEAD` (the leader) behind the integration tip shows the integration item
/// last, no `🌳`, and no short commit of `HEAD`.
#[test]
fn a_detached_leader_shows_the_integration_item_and_no_checked_out_item() {
    let repo = Repo::new();
    let behind = repo.head();
    repo.commit("landing 1");
    repo.commit("landing 2");
    repo.git(&["checkout", "-q", "--detach", &behind]);
    let tip = repo.tip();
    assert_ne!(tip, repo.head(), "the detached HEAD is behind the tip");
    let line = repo.first_line(false, None);
    assert_eq!(
        line,
        format!(
            "🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ {INTEGRATION}(🏷️ {tip})"
        ),
        "{line}"
    );
    assert!(!line.contains('🌳'), "{line}");
    assert!(!line.contains(&repo.head()), "the commit of HEAD: {line}");
}

/// AC-12: the seven characters are the commit of the local integration ref and never the one of
/// `HEAD` — with a lane branch ahead, with a detached `HEAD` behind, with the ref only in
/// `packed-refs`, and again after the ref moves.
#[test]
fn the_tip_is_the_commit_of_the_integration_ref_and_never_the_one_of_head() {
    let repo = Repo::new();
    // A lane branch ahead of the integration branch.
    repo.lane_two_ahead();
    let tip = repo.tip();
    assert_ne!(tip, repo.head());
    let line = repo.first_line(true, None);
    assert!(line.contains(&format!("(🏷️ {tip})")), "{line}");
    assert!(!line.contains(&repo.head()), "the commit of HEAD: {line}");

    // A detached `HEAD` behind it: the integration branch moves on, HEAD stays at `base`.
    repo.git(&["checkout", "-q", "--detach", &format!("{LANE}~2")]);
    repo.git(&["update-ref", &format!("refs/heads/{INTEGRATION}"), LANE]);
    let moved = repo.tip();
    assert_ne!(moved, tip, "the integration ref moved to the lane tip");
    assert_ne!(moved, repo.head());
    let line = repo.first_line(false, None);
    assert!(line.ends_with(&format!("(🏷️ {moved})")), "{line}");
    assert!(!line.contains(&repo.head()), "the commit of HEAD: {line}");

    // The ref only in `packed-refs`: no loose ref file, the same seven characters.
    repo.git(&["pack-refs", "--all", "--prune"]);
    let loose = repo.project.join(".git/refs/heads").join(INTEGRATION);
    assert!(
        !loose.exists(),
        "the loose ref is gone: {}",
        loose.display()
    );
    let packed = std::fs::read_to_string(repo.project.join(".git/packed-refs")).unwrap();
    assert!(
        packed.contains(&format!("refs/heads/{INTEGRATION}")),
        "{packed}"
    );
    assert_eq!(repo.tip(), moved);
    let line = repo.first_line(false, None);
    assert!(line.ends_with(&format!("(🏷️ {moved})")), "{line}");

    // On a lane branch with the packed ref only.
    repo.git(&["checkout", "-q", LANE]);
    repo.git(&["reset", "-q", "--hard", &format!("{LANE}~1")]);
    assert_ne!(repo.head(), moved);
    let line = repo.first_line(true, None);
    assert_eq!(line, lane_line(&moved, LANE), "{line}");
}

/// AC-13: an integration branch the profile names, with no local ref, says `n/a` — never a guess —
/// on a lane branch and on a detached `HEAD`.
#[test]
fn an_integration_branch_without_a_local_ref_says_n_a() {
    let repo = Repo::new();
    repo.lane_two_ahead();
    repo.git(&["branch", "-q", "-m", INTEGRATION, "renamed-away"]);
    assert!(!repo
        .project
        .join(".git/refs/heads")
        .join(INTEGRATION)
        .exists());
    let line = repo.first_line(true, None);
    assert_eq!(
        line,
        format!("🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ {INTEGRATION}(🏷️ n/a) | 🌳 {LANE} | 🔗 #1234 pending"),
        "{line}"
    );
    repo.git(&["checkout", "-q", "--detach", "HEAD"]);
    let line = repo.first_line(false, None);
    assert_eq!(
        line,
        format!("🧠 Sonnet 5.5 high·t | 📡 modu-ai/moai-adk, ⬇️ 1 🔀 0 | 🅱️ {INTEGRATION}(🏷️ n/a)"),
        "{line}"
    );
}

/// AC-14: a lane that is ahead of the integration branch and behind it at once, with a staged, a
/// modified and an untracked file, shows no `🧭`, `↑`, `↓` or `+N` on the first line, and keeps
/// the repository counts and the request.
#[test]
fn the_first_line_has_no_distance_and_no_ahead_behind_or_uncommitted_mark() {
    let repo = Repo::new();
    repo.git(&["checkout", "-q", "-b", LANE, INTEGRATION]);
    repo.git(&["checkout", "-q", INTEGRATION]);
    repo.commit("integration moved");
    repo.git(&["checkout", "-q", LANE]);
    repo.commit("lane 1");
    repo.commit("lane 2");
    std::fs::write(repo.project.join("tracked.txt"), "one\n").unwrap();
    repo.git(&["add", "tracked.txt"]);
    std::fs::write(repo.project.join("tracked.txt"), "two\n").unwrap();
    std::fs::write(repo.project.join("untracked.txt"), "u\n").unwrap();
    let tip = repo.tip();
    let line = repo.first_line(true, None);
    assert_eq!(line, lane_line(&tip, LANE), "{line}");
    for mark in ["🧭", "↑", "↓", "+"] {
        assert!(!line.contains(mark), "{mark} in {line}");
    }
    assert!(line.contains("⬇️ 1 🔀 0"), "the repository counts: {line}");
    assert!(line.ends_with("🔗 #1234 pending"), "the request: {line}");
}
