//! Issue #2669: `lane-safe-to-reset` answers `safe` only when every check ran clean, names
//! each finding (uncommitted, unpushed, stashed work, background jobs, a claim without a
//! resume note) as `not_safe`, and says `unknown` with a reason for a check that could not run.
//! Every repository here is a temporary fixture; the binary runs with a cleared environment
//! and an empty temporary HOME.

mod common;

use common::git::git_ok;
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use tempfile::TempDir;

const NOW: &str = "2026-11-13T09:00:00Z";

fn g(dir: &Path, args: &[&str]) {
    let mut full = vec![
        "-c",
        "user.name=Fixture",
        "-c",
        "user.email=fixture@example.invalid",
        "-c",
        "commit.gpgsign=false",
    ];
    full.extend_from_slice(args);
    git_ok(dir, &full);
}

struct Fixture {
    tmp: TempDir,
    primary: PathBuf,
    wt7: PathBuf,
}

/// A remote, a primary checkout on `main` (pushed), and a worktree on `issue/7-card`
/// (pushed), plus a detached audit worktree named for lane `agent-9`.
fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = tmp.path().canonicalize().unwrap();
    let remote = root.join("remote.git");
    let primary = root.join("primary");
    git_ok(&root, &["init", "-q", "--bare", "-b", "main", "remote.git"]);
    git_ok(&root, &["init", "-q", "-b", "main", "primary"]);
    fs::write(primary.join("a.txt"), "a\n").unwrap();
    g(&primary, &["add", "a.txt"]);
    g(&primary, &["commit", "-q", "-m", "a"]);
    g(
        &primary,
        &["remote", "add", "origin", remote.to_str().unwrap()],
    );
    g(&primary, &["push", "-q", "origin", "main"]);
    let wt7 = primary.join("wt").join("7-card");
    g(
        &primary,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/7-card",
            wt7.to_str().unwrap(),
            "main",
        ],
    );
    g(&wt7, &["push", "-q", "origin", "issue/7-card"]);
    let audit = primary.join("wt").join("agent-9-audit-1");
    g(
        &primary,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            audit.to_str().unwrap(),
            "main",
        ],
    );
    Fixture { tmp, primary, wt7 }
}

fn observations(f: &Fixture, value: Value) -> PathBuf {
    let path = f.tmp.path().join("obs.json");
    fs::write(&path, value.to_string()).unwrap();
    path
}

/// Clean observations for lane `agent-9`, which holds Issue 7 and works in `wt7`.
fn clean_obs(f: &Fixture) -> Value {
    json!({
        "schema_version": 1,
        "claims": {"source": "fixture", "items": [{"issue": 7, "claim_comment": 70, "resume_note": 71}]},
        "session_cwd": {"source": "fixture", "path": f.wt7},
        "screen": {"source": "fixture", "observed_at": "2026-11-13T08:55:00Z", "background_jobs": 0}
    })
}

fn run(f: &Fixture, lane: &str, obs: Option<Value>) -> Value {
    let home = f.tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.env_clear()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", &home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .args([
            "lane-safe-to-reset",
            f.primary.to_str().unwrap(),
            "--lane",
            lane,
            "--now",
            NOW,
        ]);
    if let Some(o) = obs {
        cmd.arg("--observations").arg(observations(f, o));
    }
    let out = cmd.output().unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stdout)))
}

fn result(v: &Value) -> &Value {
    assert_eq!(v["ok"], true, "{v}");
    &v["result"]
}

fn texts(v: &Value, key: &str) -> Vec<String> {
    v[key]
        .as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

#[test]
fn a_clean_lane_is_safe_and_its_worktrees_are_mapped_with_reasons() {
    let f = fixture();
    let v = run(&f, "agent-9", Some(clean_obs(&f)));
    let r = result(&v);
    assert_eq!(r["verdict"], "safe", "{r}");
    assert!(
        texts(r, "findings").is_empty() && texts(r, "unknowns").is_empty(),
        "{r}"
    );
    let mapped: Vec<(String, Value)> = r["worktrees"]
        .as_array()
        .unwrap()
        .iter()
        .map(|w| {
            let p = w["path"].as_str().unwrap();
            (
                Path::new(p)
                    .file_name()
                    .unwrap()
                    .to_string_lossy()
                    .into_owned(),
                w["matched_by"].clone(),
            )
        })
        .collect();
    // The primary checkout contains the session folder but is not the session's worktree.
    assert_eq!(
        mapped,
        vec![
            ("7-card".to_string(), json!(["claim #7", "session_cwd"])),
            ("agent-9-audit-1".to_string(), json!(["folder name"])),
        ]
    );
    // `agent-90` is another lane: the folder name match is a whole token.
    let other = run(&f, "agent-90", Some(json!({"schema_version": 1})));
    assert!(result(&other)["worktrees"].as_array().unwrap().is_empty());
}

#[test]
fn each_kind_of_left_work_makes_the_lane_not_safe() {
    let f = fixture();
    // Uncommitted.
    fs::write(f.wt7.join("draft.txt"), "x\n").unwrap();
    let r = run(&f, "agent-9", Some(clean_obs(&f)));
    assert_eq!(result(&r)["verdict"], "not_safe");
    assert!(
        texts(result(&r), "findings")[0].ends_with("1 uncommitted change(s)"),
        "{r}"
    );
    // Stashed on the lane's branch (the stash empties the tree again).
    g(&f.wt7, &["stash", "push", "-q", "-u", "-m", "wip"]);
    let r = run(&f, "agent-9", Some(clean_obs(&f)));
    assert_eq!(texts(result(&r), "findings").len(), 1, "{r}");
    assert!(
        texts(result(&r), "findings")[0].ends_with("1 stash entry(s)"),
        "{r}"
    );
    g(&f.wt7, &["stash", "drop", "-q"]);
    // Unpushed: a commit on no remote-tracking ref.
    fs::write(f.wt7.join("b.txt"), "b\n").unwrap();
    g(&f.wt7, &["add", "b.txt"]);
    g(&f.wt7, &["commit", "-q", "-m", "b"]);
    let r = run(&f, "agent-9", Some(clean_obs(&f)));
    assert_eq!(
        texts(result(&r), "findings"),
        vec![format!("{}: 1 unpushed commit(s)", f.wt7.display())]
    );
    g(&f.wt7, &["push", "-q", "origin", "issue/7-card"]);
    assert_eq!(
        result(&run(&f, "agent-9", Some(clean_obs(&f))))["verdict"],
        "safe"
    );
    // A background job on the screen.
    let mut obs = clean_obs(&f);
    obs["screen"]["background_jobs"] = json!(1);
    let r = run(&f, "agent-9", Some(obs));
    assert_eq!(
        texts(result(&r), "findings"),
        vec!["1 background job(s) still running on the lane's screen"]
    );
    // An open claim without a resume note.
    let mut obs = clean_obs(&f);
    obs["claims"]["items"][0]["resume_note"] = Value::Null;
    let r = run(&f, "agent-9", Some(obs));
    assert_eq!(
        texts(result(&r), "findings"),
        vec!["open claim #7 (comment 70) has no resume note"]
    );
}

#[test]
fn a_check_that_cannot_run_makes_the_lane_unknown_with_its_reason() {
    let f = fixture();
    // No observations at all: claims, session folder and screen are unchecked.
    let r = run(&f, "agent-9", None);
    assert_eq!(result(&r)["verdict"], "unknown");
    assert_eq!(texts(result(&r), "unknowns").len(), 3, "{r}");
    // Screen age: 600 s is accepted, 601 s is not (NOW = 09:00:00).
    let mut obs = clean_obs(&f);
    obs["screen"]["observed_at"] = json!("2026-11-13T08:50:00Z");
    assert_eq!(
        result(&run(&f, "agent-9", Some(obs.clone())))["verdict"],
        "safe"
    );
    obs["screen"]["observed_at"] = json!("2026-11-13T08:49:59Z");
    let r = run(&f, "agent-9", Some(obs.clone()));
    assert_eq!(result(&r)["verdict"], "unknown");
    assert_eq!(
        texts(result(&r), "unknowns"),
        vec!["screen observation is 601 s old; the limit is 600 s"]
    );
    // An observation from the future is not trusted either.
    obs["screen"]["observed_at"] = json!("2026-11-13T09:00:01Z");
    assert_eq!(result(&run(&f, "agent-9", Some(obs)))["verdict"], "unknown");
    // A claim with no local worktree, and a session folder outside the repository.
    let mut obs = clean_obs(&f);
    obs["claims"]["items"]
        .as_array_mut()
        .unwrap()
        .push(json!({"issue": 8, "claim_comment": 80, "resume_note": 81}));
    obs["session_cwd"]["path"] = json!(f.tmp.path().join("home"));
    let r = run(&f, "agent-9", Some(obs));
    let unknowns = texts(result(&r), "unknowns");
    assert_eq!(result(&r)["verdict"], "unknown", "{r}");
    assert!(
        unknowns.iter().any(|u| u.contains("is in no worktree")),
        "{unknowns:?}"
    );
    assert!(
        unknowns
            .iter()
            .any(|u| u.contains("open claim #8 has no local")),
        "{unknowns:?}"
    );
    // A finding outranks an unknown: only the session folder is observed (claims and screen
    // stay unknown), and that worktree holds an uncommitted file.
    fs::write(f.wt7.join("draft.txt"), "x\n").unwrap();
    let only_cwd =
        json!({"schema_version": 1, "session_cwd": {"source": "fixture", "path": f.wt7}});
    let r = run(&f, "agent-9", Some(only_cwd));
    assert_eq!(result(&r)["verdict"], "not_safe", "{r}");
    assert_eq!(texts(result(&r), "unknowns").len(), 2, "{r}");
}

#[test]
fn bad_input_is_refused_as_invalid_input() {
    let f = fixture();
    for (lane, obs) in [
        ("agent 9", json!({"schema_version": 1})),
        ("agent-9", json!({"schema_version": 2})),
        ("agent-9", json!({"schema_version": 1, "extra": true})),
    ] {
        let v = run(&f, lane, Some(obs));
        assert_eq!(v["ok"], false, "{v}");
        assert_eq!(v["error"]["kind"], "invalid_input", "{v}");
    }
}
