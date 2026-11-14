//! Issue #2902 O3b: `leader-round` turns the leader's screen observations into one action per
//! lane, running the real `lane-safe-to-reset` check and reading the session registry, and
//! writes nothing. Every repository here is a temporary fixture; the binary runs with a
//! cleared environment, an empty temporary HOME and no git configuration of the machine.

mod common;
use common::CleanEnv;

use common::git::git_ok;
use serde_json::{json, Value};
use std::collections::BTreeMap;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::Command;
use tempfile::TempDir;

const NOW: &str = "2026-11-14T09:21:00Z";
const READ_AT: &str = "2026-11-14T09:20:00Z";
const LANE: &str = "agent-29";

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
    root: PathBuf,
    primary: PathBuf,
    wt: PathBuf,
    home: PathBuf,
}

/// A remote, a primary checkout on `main` and the lane's worktree on `issue/2902-card`, both
/// pushed; an empty HOME.
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
    let wt = primary.join("wt").join("2902-card");
    g(
        &primary,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/2902-card",
            wt.to_str().unwrap(),
            "main",
        ],
    );
    g(&wt, &["push", "-q", "origin", "issue/2902-card"]);
    let home = root.join("home");
    fs::create_dir_all(&home).unwrap();
    Fixture {
        tmp,
        root,
        primary,
        wt,
        home,
    }
}

/// The lane's screen after `LANE-DONE`: idle, input box empty, claim #2902 with its resume
/// note, the session in the lane's worktree.
fn observations(f: &Fixture) -> Value {
    json!({
        "schema_version": 1,
        "lanes": [{
            "lane": LANE,
            "host": "claude",
            "read_at": READ_AT,
            "source": "fixture",
            "terminal": {
                "handle": "term-29", "status": "running", "source": "screen",
                "tail": "⏺ PR #2912 merged.\n\n⏺ LANE-DONE agent-29 #2902 — PR #2912 MERGED",
                "limited": false, "truncated": false, "draft": null
            },
            "pane": "working",
            "turn": "idle",
            "input_box": "empty",
            "context_used_pct": 40,
            "wake_armed": false,
            "background_jobs": 0,
            "reset": {
                "claims": {"source": "fixture",
                    "items": [{"issue": 2902, "claim_comment": 6874287412_u64, "resume_note": 6874287500_u64}]},
                "session_cwd": {"source": "fixture", "path": f.wt}
            }
        }]
    })
}

fn moai(f: &Fixture, args: &[&str]) -> Value {
    let out = Command::new(crate::common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .args(args)
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{e}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    })
}

/// Runs `leader-round` on the primary checkout with the observations (and a snapshot).
fn round(f: &Fixture, obs: &Value, snapshot: Option<&Value>) -> Value {
    let args = inputs(f, obs, snapshot);
    run_round(f, &args)
}

fn run_round(f: &Fixture, args: &[String]) -> Value {
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let v = moai(f, &args);
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

/// Writes the input files and returns the `leader-round` arguments.
fn inputs(f: &Fixture, obs: &Value, snapshot: Option<&Value>) -> Vec<String> {
    let input = f.tmp.path().join("input");
    fs::create_dir_all(&input).unwrap();
    let obs_path = input.join("observations.json");
    fs::write(&obs_path, obs.to_string()).unwrap();
    let mut args = vec![
        "leader-round".to_string(),
        f.primary.to_string_lossy().into_owned(),
        "--observations".into(),
        obs_path.to_string_lossy().into_owned(),
        "--now".into(),
        NOW.into(),
    ];
    if let Some(s) = snapshot {
        let p = input.join("snapshot.json");
        fs::write(&p, s.to_string()).unwrap();
        args.extend([
            "--snapshot".into(),
            p.to_string_lossy().into_owned(),
            "--repo".into(),
            "o/r".into(),
        ]);
    }
    args
}

fn texts(v: &Value) -> Vec<String> {
    v.as_array()
        .unwrap()
        .iter()
        .map(|s| s.as_str().unwrap().to_string())
        .collect()
}

/// AC3: a finished card on a lane that `lane-safe-to-reset` finds safe is cleared and
/// redispatched with the duty line, and the release names the instance the lane registered.
#[test]
fn a_finished_safe_lane_is_cleared_and_redispatched() {
    let f = fixture();
    let reg = moai(
        &f,
        &[
            "session-register",
            "--name",
            LANE,
            "--now",
            READ_AT,
            f.primary.to_str().unwrap(),
        ],
    );
    let instance = reg["result"]["instance"].as_str().unwrap().to_string();
    let r = round(&f, &observations(&f), None);
    let a = &r["lanes"][0];
    assert_eq!(
        (a["action"].as_str(), a["reason"].as_str()),
        (Some("clear_redispatch"), Some("card_boundary")),
        "{r}"
    );
    let duty = moai(&f, &["lane-duty-line", "--lane", LANE]);
    assert_eq!(
        a["send"],
        json!(["/clear", duty["result"]["line"].as_str().unwrap()])
    );
    assert_eq!(
        a["release"],
        json!(format!(
            "moai session-register --release --by leader --name {LANE} --instance {instance} {}",
            f.primary.display()
        ))
    );
    assert_eq!(r["writes"], false);
    // No GitHub state: the round's shared parts are not run, and row 7 cannot see the end.
    for key in [
        "close_candidates",
        "handoff_candidates",
        "gate",
        "end_condition",
    ] {
        assert_eq!(r["round"][key]["status"], "not_run", "{key}: {r}");
    }
}

/// AC4: an unpushed commit in the lane's worktree makes the real reset check `not_safe`; the
/// lane is left with the finding and nothing is sent.
#[test]
fn an_unpushed_head_leaves_the_lane_with_the_finding() {
    let f = fixture();
    fs::write(f.wt.join("b.txt"), "b\n").unwrap();
    g(&f.wt, &["add", "b.txt"]);
    g(&f.wt, &["commit", "-q", "-m", "b"]);
    let r = round(&f, &observations(&f), None);
    let a = &r["lanes"][0];
    assert_eq!(
        (a["action"].as_str(), a["reason"].as_str()),
        (Some("leave"), Some("reset_not_safe")),
        "{r}"
    );
    let findings = texts(&a["findings"]);
    assert_eq!(findings.len(), 1, "{findings:?}");
    assert!(
        findings[0].ends_with("1 unpushed commit(s)"),
        "{findings:?}"
    );
    assert_eq!(a["send"], json!([]));
    assert_eq!(a["release"], Value::Null);
}

/// Every path under `dir` with its kind, bytes and modification time.
fn tree(dir: &Path, out: &mut BTreeMap<PathBuf, (String, Vec<u8>, Option<std::time::SystemTime>)>) {
    for e in fs::read_dir(dir).unwrap() {
        let e = e.unwrap();
        let p = e.path();
        let m = fs::symlink_metadata(&p).unwrap();
        let mtime = m.modified().ok();
        if m.is_dir() {
            out.insert(p.clone(), ("dir".into(), Vec::new(), mtime));
            tree(&p, out);
        } else if m.file_type().is_symlink() {
            let target = fs::read_link(&p).unwrap();
            out.insert(
                p,
                (
                    "link".into(),
                    target.to_string_lossy().as_bytes().to_vec(),
                    None,
                ),
            );
        } else {
            out.insert(p.clone(), ("file".into(), fs::read(&p).unwrap(), mtime));
        }
    }
}

/// AC13: a round with a snapshot writes nothing — the repositories with their `.git`, the
/// remote, HOME and the session registry are the same before and after, and no registry
/// folder is created.
#[test]
fn a_round_writes_nothing() {
    let f = fixture();
    let lease_end = "2026-11-14T09:21:00Z";
    let snapshot = json!({
        "threads": [
            {"number": 2902, "comments": [
                {"id": 1, "body": format!("moai-claim: session=agent-7 epoch=1 lease_until={lease_end}")},
                {"id": 2, "body": "CLOSE-TABLE #2902\n| AC | met |"}
            ]},
            {"number": 2903, "comments": [
                {"id": 3, "body": "moai-claim: session=agent-8 epoch=1 lease_until=2026-11-14T09:21:01Z"}
            ]}
        ],
        "cards": [{"number": 2902}, {"number": 2903}]
    });
    // The inputs are written first, so the comparison covers only the run itself.
    let args = inputs(&f, &observations(&f), Some(&snapshot));
    let mut before = BTreeMap::new();
    tree(&f.root, &mut before);
    let r = run_round(&f, &args);
    let mut after = BTreeMap::new();
    tree(&f.root, &mut after);
    assert!(before == after, "the round changed the tree");
    assert!(
        !after
            .keys()
            .any(|p| p.components().any(|c| c.as_os_str() == ".moai-cli")),
        "a .moai-cli folder exists"
    );
    assert!(before.keys().any(|p| p.ends_with(".git/HEAD")));
    // The lane is cleared (safe), but with no registered session there is nothing to release.
    let a = &r["lanes"][0];
    assert_eq!(a["action"], "clear_redispatch", "{r}");
    assert_eq!(a["release"], Value::Null);
    assert!(texts(&a["notes"])
        .iter()
        .any(|n| n.contains("nothing to release")));
    // The shared parts, read from the snapshot: the lease-end second is expired, one second
    // later still live.
    assert_eq!(
        r["round"]["close_candidates"]["items"],
        json!([{"issue": 2902, "comment": 2}])
    );
    assert_eq!(
        r["round"]["handoff_candidates"]["items"],
        json!([{"issue": 2902, "claim_comment": 1, "session": "agent-7", "epoch": 1, "lease_until": lease_end}])
    );
    assert_eq!(
        r["round"]["gate"]["command"],
        json!(format!(
            "moai gate-loop --once --fetch --repo o/r --dir {}",
            f.primary.display()
        ))
    );
    assert_eq!(r["round"]["end_condition"]["met"], false);
    assert_eq!(r["api_calls"], 0);
}
