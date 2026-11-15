//! #3015 C2: `lane-next` offers two lanes that ask back to back two different targets, by
//! keeping the offer of each registered lane for a short lease in the clone's session registry
//! folder. Synthetic repositories in a temporary folder with a temporary HOME; the snapshots
//! are fixed JSON and no GitHub request is made.
#![cfg(unix)]
mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;

const BIN: &str = crate::common::MOAI_BIN;

fn h(c: char) -> String {
    c.to_string().repeat(40)
}

struct Fixture {
    _dir: tempfile::TempDir,
    repo: PathBuf,
    home: PathBuf,
}

impl Fixture {
    fn new() -> Self {
        let dir = tempfile::tempdir().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let repo = base.join("repo");
        let home = base.join("home");
        fs::create_dir_all(&repo).unwrap();
        fs::create_dir_all(&home).unwrap();
        common::git::git(&repo, &["init", "-q", "-b", "main"]);
        Fixture {
            _dir: dir,
            repo,
            home,
        }
    }

    fn proposals(&self) -> PathBuf {
        self.repo.join(".git/.moai-cli/sessions/proposals")
    }

    fn register(&self, name: &str) {
        let out = Command::new(BIN)
            .arg("session-register")
            .arg(&self.repo)
            .args(["--name", name])
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["ok"], json!(true), "{v}");
    }

    /// `lane-next --snapshot` for `lane`; `dir` is passed as `--dir` when given.
    fn next(&self, lane: &str, snap: &Value, dir: Option<&Path>) -> Value {
        let file = self.home.join("snap.json");
        fs::write(&file, snap.to_string()).unwrap();
        let mut cmd = Command::new(BIN);
        cmd.args(["lane-next", "--lane", lane, "--repo", "o/r", "--snapshot"])
            .arg(&file)
            .current_dir(&self.home)
            .env("HOME", &self.home)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null");
        if let Some(d) = dir {
            cmd.arg("--dir").arg(d);
        }
        let out = cmd.output().unwrap();
        let v: Value = serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(
            v["result"]["writes"],
            json!(false),
            "GitHub is never written: {v}"
        );
        v["result"].clone()
    }
}

/// Two other lanes' pull requests with no auditor.
fn two_prs() -> Value {
    json!({"prs": [
        {"number": 40, "head": h('a'), "author_lane": "agent-28", "comments": []},
        {"number": 41, "head": h('b'), "author_lane": "agent-28", "comments": []},
    ]})
}

fn number(r: &Value) -> u64 {
    r["next"]["target"]["number"].as_u64().unwrap()
}

#[test]
fn two_sessions_calling_back_to_back_get_different_targets() {
    let fx = Fixture::new();
    fx.register("agent-1");
    fx.register("agent-2");
    fx.register("agent-3");
    let a = fx.next("agent-1", &two_prs(), Some(&fx.repo));
    let b = fx.next("agent-2", &two_prs(), Some(&fx.repo));
    assert_eq!(number(&a), 40, "{a}");
    assert_eq!(number(&b), 41, "{b}");
    assert_eq!(a["proposal"]["status"], "recorded", "{a}");
    assert_eq!(
        b["pr_skipped"][0]["reason"], "proposed_to_other_lane",
        "{b}"
    );
    assert_eq!(b["pr_skipped"][0]["by"], "agent-1", "{b}");
    // Both targets are taken: a third lane is offered none, and says why.
    let c = fx.next("agent-3", &two_prs(), Some(&fx.repo));
    assert!(c["next"].is_null(), "{c}");
    assert_eq!(c["proposal"]["status"], "none", "{c}");
    // The lane asking again gets its own target again with its first lease.
    let a2 = fx.next("agent-1", &two_prs(), Some(&fx.repo));
    assert_eq!(number(&a2), 40, "{a2}");
    assert_eq!(a2["proposal"]["status"], "kept", "{a2}");
    assert_eq!(a2["proposal"]["lease_until"], a["proposal"]["lease_until"]);
}

#[test]
fn a_snapshot_without_dir_keeps_nothing() {
    let fx = Fixture::new();
    fx.register("agent-1");
    let a = fx.next("agent-1", &two_prs(), None);
    assert!(a.get("proposal").is_none(), "{a}");
    assert!(!fx.proposals().exists(), "no offer was written");
    // The registry stays as it was: only the registration file and its lock.
    let b = fx.next("agent-2", &two_prs(), None);
    assert_eq!(number(&a), number(&b));
}

#[test]
fn a_lane_that_never_registered_blocks_nobody() {
    let fx = Fixture::new();
    fx.register("agent-2");
    let a = fx.next("agent-1", &two_prs(), Some(&fx.repo));
    assert_eq!(a["proposal"]["status"], "caller_not_registered", "{a}");
    let b = fx.next("agent-2", &two_prs(), Some(&fx.repo));
    assert_eq!(number(&b), 40, "{b}");
}

#[test]
fn an_unreadable_proposal_leaves_the_old_decision_and_says_so() {
    let fx = Fixture::new();
    fx.register("agent-1");
    fs::create_dir_all(fx.proposals()).unwrap();
    fs::write(fx.proposals().join("agent-2.json"), "{broken").unwrap();
    let a = fx.next("agent-1", &two_prs(), Some(&fx.repo));
    assert_eq!(number(&a), 40, "{a}");
    assert_eq!(a["proposal"]["status"], "unavailable", "{a}");
    assert!(
        !a["proposal"].to_string().contains("broken"),
        "no file text in the reason: {a}"
    );
}

#[test]
fn a_folder_without_a_registry_says_unavailable_and_still_decides() {
    let fx = Fixture::new();
    // A git clone where nobody registered: no registry folder yet.
    let a = fx.next("agent-1", &two_prs(), Some(&fx.repo));
    assert_eq!(number(&a), 40, "{a}");
    assert_eq!(a["proposal"]["status"], "unavailable", "{a}");
    assert!(
        !fx.proposals().exists(),
        "the folder is not created by lane-next"
    );
    // A folder that is no git clone at all.
    let plain = fx.home.join("plain");
    fs::create_dir_all(&plain).unwrap();
    let b = fx.next("agent-1", &two_prs(), Some(&plain));
    assert_eq!(number(&b), 40, "{b}");
    assert_eq!(b["proposal"]["status"], "unavailable", "{b}");
}
