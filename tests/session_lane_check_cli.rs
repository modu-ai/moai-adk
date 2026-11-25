//! SPEC-20261125-session-names-project-prefix (REQ-7, REQ-8, REQ-12): `moai session-register
//! --lane-check <name>` says whether a session name is a lane of this clone's registry before the
//! leader sends it a message, and `--names` prints the names of this project. Both read only. Every
//! case runs in a temporary git project with a temporary HOME; nothing of this machine is read.
#![cfg(unix)]
mod common;

use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

struct Fixture {
    _root: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn new(folder: &str) -> Self {
        let root = tempfile::tempdir().unwrap();
        let base = fs::canonicalize(root.path()).unwrap();
        let repo = base.join(folder);
        let home = base.join("home");
        fs::create_dir_all(&repo).unwrap();
        fs::create_dir_all(&home).unwrap();
        assert!(common::git::command()
            .args(["init", "-q"])
            .current_dir(&repo)
            .status()
            .unwrap()
            .success());
        Fixture {
            _root: root,
            repo,
            home,
        }
    }

    fn sessions(&self) -> PathBuf {
        moai_cli::session::agent_seq_dir(&self.repo).unwrap()
    }

    fn run(&self, args: &[&str]) -> (Output, Value) {
        let out = Command::new(common::MOAI_BIN)
            .arg("session-register")
            .args(args)
            .current_dir(&self.repo)
            .env_clear()
            .env("HOME", &self.home)
            .env("PATH", "/usr/bin:/bin")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .output()
            .unwrap();
        let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
        (out, v)
    }

    /// The answer's `result` of a successful command.
    fn answer(&self, args: &[&str]) -> Value {
        let (out, v) = self.run(args);
        assert!(out.status.success(), "{args:?}: {}", text(&out));
        assert_eq!(v["ok"], true, "{args:?}: {v}");
        v["result"].clone()
    }
}

fn text(out: &Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// Every file of `dir` with its bytes, to compare a folder before and after a read.
fn snapshot(dir: &Path) -> BTreeMap<String, Vec<u8>> {
    let mut seen = BTreeMap::new();
    if let Ok(entries) = fs::read_dir(dir) {
        for e in entries.flatten() {
            seen.insert(
                e.file_name().to_string_lossy().into_owned(),
                fs::read(e.path()).unwrap_or_default(),
            );
        }
    }
    seen
}

/// AC-3 (REQ-8): a name whose live lock is held is a lane, and it is not any more once the lock
/// is let go.
#[test]
fn ac3_a_held_live_lock_is_a_lane() {
    let f = Fixture::new("demo");
    let held = moai_cli::session::lock_named_live(&f.sessions(), "demo-agent-ben").unwrap();
    let said = f.answer(&["--lane-check", "demo-agent-ben"]);
    assert_eq!(said["status"], "lane", "{said}");
    assert_eq!(said["basis"], "live_lock", "{said}");
    assert_eq!(said["name"], "demo-agent-ben", "{said}");
    drop(held);
    let said = f.answer(&["--lane-check", "demo-agent-ben"]);
    assert_eq!(said["status"], "not_lane", "{said}");
    assert!(
        said["reason"].as_str().is_some_and(|r| !r.is_empty()),
        "{said}"
    );
}

/// AC-3 (REQ-8): a registry entry whose lease runs is a lane; one whose lease is over is not.
#[test]
fn ac3_a_running_lease_is_a_lane_and_an_ended_one_is_not() {
    let f = Fixture::new("demo");
    let registered = f.answer(&[
        "--name",
        "demo-agent-cara",
        "--mode",
        "agent",
        "--lease-secs",
        "3600",
        f.repo.to_str().unwrap(),
    ]);
    assert!(registered["instance"].is_string(), "{registered}");
    let said = f.answer(&["--lane-check", "demo-agent-cara"]);
    assert_eq!(said["status"], "lane", "{said}");
    assert_eq!(said["basis"], "registry_lease", "{said}");
    let said = f.answer(&[
        "--lane-check",
        "demo-agent-cara",
        "--now",
        "2099-01-01T00:00:00Z",
    ]);
    assert_eq!(said["status"], "not_lane", "{said}");
}

/// AC-3 (REQ-7, REQ-8): a name with neither, a name the registry refuses, an unreadable entry and
/// a symlinked sessions folder all answer `not_lane` with the reason, and never `lane`.
#[test]
fn ac3_every_other_name_is_not_a_lane() {
    let f = Fixture::new("demo");
    let said = f.answer(&["--lane-check", "demo-agent-dana"]);
    assert_eq!(said["status"], "not_lane", "{said}");
    let said = f.answer(&["--lane-check", "bad name"]);
    assert_eq!(said["status"], "not_lane", "{said}");
    // An old shape of a name is judged by the same rule, not by its spelling: no lock, no lane.
    let said = f.answer(&["--lane-check", "agent-ben"]);
    assert_eq!(said["status"], "not_lane", "{said}");
    // An entry that cannot be read is no proof of a lane.
    fs::create_dir_all(f.sessions()).unwrap();
    fs::write(f.sessions().join("demo-agent-eli.json"), "{not json").unwrap();
    let said = f.answer(&["--lane-check", "demo-agent-eli"]);
    assert_eq!(said["status"], "not_lane", "{said}");
    assert!(
        said["reason"].as_str().is_some_and(|r| !r.is_empty()),
        "{said}"
    );
    // A sessions folder that is a link never answers `lane`, even with a lock held inside it.
    // The entry is a good one with a running lease: only the link makes the answer `not_lane`.
    let g = Fixture::new("linked");
    g.answer(&[
        "--name",
        "linked-agent-ben",
        "--mode",
        "agent",
        g.repo.to_str().unwrap(),
    ]);
    assert_eq!(
        g.answer(&["--lane-check", "linked-agent-ben"])["status"],
        "lane"
    );
    let target = g._root.path().join("elsewhere");
    let sessions = g.sessions();
    fs::rename(&sessions, &target).unwrap();
    std::os::unix::fs::symlink(&target, &sessions).unwrap();
    let said = g.answer(&["--lane-check", "linked-agent-ben"]);
    assert_eq!(said["status"], "not_lane", "{said}");
    assert!(
        said["reason"]
            .as_str()
            .is_some_and(|r| r.contains("symlink")),
        "{said}"
    );
}

/// AC-3 (REQ-8): the check writes nothing: it makes no sessions folder and changes no file.
#[test]
fn ac3_the_check_writes_nothing() {
    let f = Fixture::new("demo");
    assert!(!f.sessions().exists());
    f.answer(&["--lane-check", "demo-agent-ben"]);
    f.answer(&["--names"]);
    assert!(!f.sessions().exists(), "the check made the sessions folder");
    let _held = moai_cli::session::lock_named_live(&f.sessions(), "demo-agent-ben").unwrap();
    fs::write(f.sessions().join("demo-agent-cara.json"), "{}").unwrap();
    let before = snapshot(&f.sessions());
    f.answer(&["--lane-check", "demo-agent-ben"]);
    f.answer(&["--lane-check", "demo-agent-cara"]);
    f.answer(&["--names"]);
    assert_eq!(before, snapshot(&f.sessions()));
}

/// AC-3 (REQ-12): `--names` prints the project name the launcher uses, the leader session name and
/// the agent name prefix of this project; outside a git repository it is an error that says so.
#[test]
fn ac3_names_prints_the_names_of_this_project() {
    let f = Fixture::new("some folder");
    let said = f.answer(&["--names"]);
    assert_eq!(said["project"], "some-folder", "{said}");
    assert_eq!(said["leader"], "some-folder-leader", "{said}");
    assert_eq!(said["agent_prefix"], "some-folder-agent-", "{said}");
    let state = f.repo.join(moai_cli::PROJECT_STATE_DIR);
    fs::create_dir_all(&state).unwrap();
    fs::write(state.join("project.yaml"), "{\"project_id\": \"demo\"}\n").unwrap();
    let said = f.answer(&["--names"]);
    assert_eq!(said["project"], "demo", "{said}");
    assert_eq!(said["leader"], "demo-leader", "{said}");
    assert_eq!(said["agent_prefix"], "demo-agent-", "{said}");
    let plain = tempfile::tempdir().unwrap();
    let out = Command::new(common::MOAI_BIN)
        .args(["session-register", "--names"])
        .current_dir(plain.path())
        .env_clear()
        .env("HOME", &f.home)
        .env("PATH", "/usr/bin:/bin")
        .env("GIT_CEILING_DIRECTORIES", plain.path().parent().unwrap())
        .output()
        .unwrap();
    assert!(!out.status.success(), "{}", text(&out));
    assert!(text(&out).contains("git repository"), "{}", text(&out));
}
