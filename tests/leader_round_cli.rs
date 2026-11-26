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
    fixture_in("")
}

/// [`fixture`] inside the folder `sub` of the temporary folder (`""` for the folder itself).
fn fixture_in(sub: &str) -> Fixture {
    let tmp = TempDir::new().unwrap();
    let mut root = tmp.path().canonicalize().unwrap();
    if !sub.is_empty() {
        root.push(sub);
        fs::create_dir_all(&root).unwrap();
    }
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

/// The action and reason of the round's only lane.
fn act(r: &Value) -> (Option<&str>, Option<&str>) {
    let a = &r["lanes"][0];
    (a["action"].as_str(), a["reason"].as_str())
}

/// #2948 audit I1: without the session's folder the real reset check answers `unknown`; the
/// finished lane is left, never cleared.
#[test]
fn a_reset_check_without_the_session_folder_leaves_the_lane() {
    let f = fixture();
    let mut obs = observations(&f);
    obs["lanes"][0]["reset"]
        .as_object_mut()
        .unwrap()
        .remove("session_cwd");
    let r = round(&f, &obs, None);
    assert_eq!(act(&r), (Some("leave"), Some("reset_unknown")), "{r}");
    assert!(
        texts(&r["lanes"][0]["findings"])
            .iter()
            .any(|u| u.starts_with("no session_cwd observation")),
        "{r}"
    );
    assert_eq!(r["lanes"][0]["send"], json!([]));
}

/// #2948 audit I1: a project the reset check cannot read (no git repository) is an error of
/// the check; the finished lane is left, never cleared.
#[test]
fn a_reset_check_that_cannot_run_leaves_the_lane() {
    let f = fixture();
    let plain = f.root.join("plain");
    fs::create_dir_all(&plain).unwrap();
    let mut args = inputs(&f, &observations(&f), None);
    args[1] = plain.to_string_lossy().into_owned();
    let r = run_round(&f, &args);
    assert_eq!(act(&r), (Some("leave"), Some("reset_unknown")), "{r}");
    assert_eq!(
        texts(&r["lanes"][0]["findings"]),
        vec!["git worktree list unavailable for the project".to_string()]
    );
    assert_eq!(r["lanes"][0]["send"], json!([]));
}

/// #2948 audit N3: the release names a project path with a space as one shell word.
#[test]
fn the_release_quotes_the_project_path() {
    let f = fixture_in("with space");
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
    assert_eq!(
        act(&r),
        (Some("clear_redispatch"), Some("card_boundary")),
        "{r}"
    );
    assert!(f.primary.to_str().unwrap().contains("with space/primary"));
    assert_eq!(
        r["lanes"][0]["release"],
        json!(format!(
            "moai session-register --release --by leader --name {LANE} --instance {instance} '{}'",
            f.primary.display()
        ))
    );
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

/// #3005 P9: a lane that printed the `LANE-IDLE` line (`lane-next` answered `idle_limit`) is
/// left, never sent the restart line, and the round names it as an idle lane: a card for the
/// leader to hand out. Without the line the same lane gets the restart line.
#[test]
fn an_idle_lane_is_left_for_a_dispatch_and_named_in_the_round() {
    let f = fixture();
    let idle = "LANE-IDLE lane=agent-29 checks=6 last=2026-11-14T09:19:00Z restart=enter MODE-AGENT again as lane agent-29";
    let mut obs = observations(&f);
    obs["lanes"][0]["terminal"]["tail"] =
        json!(format!("⏺ no task, nothing to wait for\n⏺ {idle}"));
    let r = round(&f, &obs, None);
    assert_eq!(act(&r), (Some("leave"), Some("lane_idle")), "{r}");
    assert_eq!(r["lanes"][0]["send"], json!([]), "{r}");
    assert_eq!(r["idle_lanes"], json!([LANE]), "{r}");

    obs["lanes"][0]["terminal"]["tail"] = json!("⏺ no task, nothing to wait for");
    let r = round(&f, &obs, None);
    assert_eq!(act(&r), (Some("restart"), Some("continue")), "{r}");
    assert_eq!(r["idle_lanes"], json!([]), "{r}");
}

// SPEC-20261123-leader-registration-lease, AC-5 to AC-9: the `registration` part of the round.

/// The epoch of `NOW` (2026-11-14T09:21:00Z), so registry entries can name lease ends around
/// it without re-deriving the calendar in the test.
const NOW_EPOCH: i64 = 1_794_648_060;
/// A valid instance id for a fixture entry (32 lowercase hex digits).
const LEADER_INSTANCE: &str = "0123456789abcdef0123456789abcdef";
/// The host session id of the coordination window in the fixtures.
const LEADER_SESSION: &str = "aaaaaaaa-0000-4000-8000-00000000000a";
/// Another host session id.
const OTHER_SESSION: &str = "bbbbbbbb-0000-4000-8000-00000000000b";

/// The clone's session registry folder, found through the resolver like the code under test.
fn sessions_dir(f: &Fixture) -> PathBuf {
    moai_cli::clone_root::path(&f.primary.join(".git")).join("sessions")
}

/// Writes the registry entry `leader` (mode leader) with `lease_until` and, when given, the
/// host `session`.
fn leader_entry(f: &Fixture, lease_until: i64, session: Option<&str>) {
    let dir = sessions_dir(f);
    fs::create_dir_all(&dir).unwrap();
    let mut entry = json!({
        "schema_version": 1,
        "name": "leader",
        "instance": LEADER_INSTANCE,
        "lease_until": lease_until,
        "mode": "leader",
    });
    if let Some(s) = session {
        entry["session"] = json!(s);
    }
    fs::write(dir.join("leader.json"), entry.to_string()).unwrap();
}

/// `leader-round` with `extra` arguments, run as the host session `session`.
fn round_as(f: &Fixture, session: Option<&str>, extra: &[&str]) -> Value {
    let mut args = inputs(f, &observations(f), None);
    args.extend(extra.iter().map(|s| s.to_string()));
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", &f.home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .args(&args)
        .envs(session.map(|s| ("CLAUDE_CODE_SESSION_ID", s)));
    let out = cmd.output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{e}: {}{}",
            String::from_utf8_lossy(&out.stdout),
            String::from_utf8_lossy(&out.stderr)
        )
    });
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

/// AC-5: a leader lease that ended 38 hours earlier is `expired`, named with the elapsed
/// seconds and its session (the design's 3.1 finding: nothing showed it).
#[test]
fn registration_expired_lease_names_the_elapsed_seconds() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH - 38 * 3600, Some(LEADER_SESSION));
    let r = round(&f, &observations(&f), None);
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "expired", "{r}");
    let leaders = reg["leaders"].as_array().unwrap();
    assert_eq!(leaders.len(), 1, "{r}");
    assert_eq!(leaders[0]["name"], "leader", "{r}");
    assert_eq!(leaders[0]["lease_until"], "2026-11-12T19:21:00Z", "{r}");
    assert_eq!(leaders[0]["expired_for_secs"], 38 * 3600, "{r}");
    assert_eq!(leaders[0]["session"], LEADER_SESSION, "{r}");
    assert!(leaders[0].get("remaining_secs").is_none(), "{r}");
}

/// AC-5: a live lease with a session id is `ok` and names the seconds left, not elapsed ones.
#[test]
fn registration_live_lease_with_a_session_is_ok() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH + 3600, Some(LEADER_SESSION));
    let r = round(&f, &observations(&f), None);
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "ok", "{r}");
    assert_eq!(reg["leaders"][0]["remaining_secs"], 3600, "{r}");
    assert!(reg["leaders"][0].get("expired_for_secs").is_none(), "{r}");
}

/// AC-5: a live lease whose entry has no session id is `no_session`; an ended lease wins over it.
#[test]
fn registration_a_live_entry_without_a_session_is_no_session() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH + 3600, None);
    let r = round(&f, &observations(&f), None);
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "no_session", "{r}");
    assert!(reg["leaders"][0].get("session").is_none(), "{r}");
    leader_entry(&f, NOW_EPOCH - 60, None);
    let r = round(&f, &observations(&f), None);
    assert_eq!(r["round"]["registration"]["status"], "expired", "{r}");
}

/// AC-5, AC-9: no registry folder at all is `absent` — and the part is there without a GitHub
/// snapshot while the other shared parts stay `not_run`.
#[test]
fn registration_without_a_registry_is_absent_and_not_not_run() {
    let f = fixture();
    let r = round(&f, &observations(&f), None);
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "absent", "{r}");
    assert_eq!(reg["leaders"], json!([]), "{r}");
    for key in ["close_candidates", "gate"] {
        assert_eq!(r["round"][key]["status"], "not_run", "{key}: {r}");
    }
}

/// AC-5: an entry of another mode is not a leader entry.
#[test]
fn registration_lists_only_leader_mode_entries() {
    let f = fixture();
    let dir = sessions_dir(&f);
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("agent-3.json"),
        format!(
            "{{\"schema_version\":1,\"name\":\"agent-3\",\"instance\":\"{LEADER_INSTANCE}\",\
             \"lease_until\":{},\"mode\":\"agent\"}}",
            NOW_EPOCH - 100
        ),
    )
    .unwrap();
    let r = round(&f, &observations(&f), None);
    assert_eq!(r["round"]["registration"]["status"], "absent", "{r}");
}

/// AC-5 (REQ-8): two leader entries whose leases both run are never `ok`; the status is
/// `duplicate`, both are listed, and it outranks an ended lease.
#[test]
fn registration_two_live_leaders_are_a_duplicate_not_ok() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH + 3600, Some(LEADER_SESSION));
    fs::write(
        sessions_dir(&f).join("boss.json"),
        json!({
            "schema_version": 1, "name": "boss", "instance": "fedcba9876543210fedcba9876543210",
            "lease_until": NOW_EPOCH + 7200, "mode": "leader", "session": OTHER_SESSION,
        })
        .to_string(),
    )
    .unwrap();
    let r = round(&f, &observations(&f), None);
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "duplicate", "{r}");
    assert_eq!(reg["leaders"].as_array().unwrap().len(), 2, "{r}");

    fs::write(
        sessions_dir(&f).join("old.json"),
        json!({
            "schema_version": 1, "name": "old", "instance": "00112233445566778899aabbccddeeff",
            "lease_until": NOW_EPOCH - 60, "mode": "leader",
        })
        .to_string(),
    )
    .unwrap();
    let r = round(&f, &observations(&f), None);
    assert_eq!(r["round"]["registration"]["status"], "duplicate", "{r}");

    // One live leader beside an ended one is not a duplicate.
    fs::remove_file(sessions_dir(&f).join("boss.json")).unwrap();
    let r = round(&f, &observations(&f), None);
    assert_eq!(r["round"]["registration"]["status"], "expired", "{r}");
}

/// AC-9: the note names the three words a leader needs to renew at once.
#[test]
fn registration_note_names_the_renewal_words() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH - 38 * 3600, Some(LEADER_SESSION));
    let r = round(&f, &observations(&f), None);
    let note = r["round"]["registration"]["note"].as_str().unwrap();
    for word in [
        "session-register",
        "--instance",
        "leader-round --renew-instance",
    ] {
        assert!(note.contains(word), "{word}: {note}");
    }
}

/// AC-6: an entry that cannot be read makes the part `error` with that message and leaves
/// every other shared part as it was.
#[test]
fn registration_an_unreadable_entry_makes_the_part_error() {
    let f = fixture();
    let dir = sessions_dir(&f);
    fs::create_dir_all(&dir).unwrap();
    fs::write(dir.join("leader.json"), "not json").unwrap();
    let r = round(&f, &observations(&f), None);
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "error", "{r}");
    let message = reg["message"].as_str().unwrap();
    assert!(message.contains("leader.json"), "{message}");
    for key in [
        "close_candidates",
        "handoff_candidates",
        "gate",
        "end_condition",
    ] {
        assert_eq!(r["round"][key]["status"], "not_run", "{key}: {r}");
    }
}

/// AC-8: a round without `--renew-instance` writes nothing: the whole git dir is byte for byte
/// as it was, and the answer says `writes` false.
#[test]
fn registration_read_writes_nothing() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH - 38 * 3600, Some(LEADER_SESSION));
    let before = tree_snapshot(&f.primary.join(".git"));
    let r = round(&f, &observations(&f), None);
    assert_eq!(r["writes"], false, "{r}");
    let after = tree_snapshot(&f.primary.join(".git"));
    assert_eq!(before, after);
}

/// AC-7: `--renew-instance` turns the 38 hour old entry into a live one in the same answer and
/// keeps its session id.
#[test]
fn registration_renew_instance_renews_an_ended_lease() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH - 38 * 3600, Some(LEADER_SESSION));
    let r = round_as(
        &f,
        Some(LEADER_SESSION),
        &["--renew-instance", LEADER_INSTANCE],
    );
    assert_eq!(r["writes"], true, "{r}");
    let reg = &r["round"]["registration"];
    assert_eq!(reg["renewal"]["status"], "renewed", "{r}");
    assert_eq!(reg["status"], "ok", "{r}");
    assert_eq!(reg["leaders"][0]["session"], LEADER_SESSION, "{r}");
    assert_eq!(reg["leaders"][0]["remaining_secs"], 14_400, "{r}");
    let leader_file = moai_cli::lane_registry::entry_file(&sessions_dir(&f), "leader")
        .ok()
        .flatten()
        .unwrap_or_else(|| sessions_dir(&f).join("leader.json"));
    let stored: Value = serde_json::from_str(&fs::read_to_string(leader_file).unwrap()).unwrap();
    assert_eq!(stored["instance"], LEADER_INSTANCE, "{stored}");
    assert_eq!(stored["mode"], "leader", "{stored}");
    assert_eq!(stored["session"], LEADER_SESSION, "{stored}");
}

/// AC-7: a wrong instance, and a renewal from another session, are refused: the round answers
/// the rest as without the flag, says `writes` false and leaves the entry as it was.
#[test]
fn registration_renew_instance_refusals_leave_the_entry_alone() {
    let f = fixture();
    leader_entry(&f, NOW_EPOCH - 38 * 3600, Some(LEADER_SESSION));
    let path = sessions_dir(&f).join("leader.json");
    let before = fs::read(&path).unwrap();
    for (session, instance) in [
        (LEADER_SESSION, "ffffffffffffffffffffffffffffffff"),
        (OTHER_SESSION, LEADER_INSTANCE),
    ] {
        let r = round_as(&f, Some(session), &["--renew-instance", instance]);
        let reg = &r["round"]["registration"];
        assert_eq!(reg["renewal"]["status"], "refused", "{r}");
        assert!(reg["renewal"]["message"].as_str().is_some(), "{r}");
        assert_eq!(reg["status"], "expired", "{r}");
        assert_eq!(r["writes"], false, "{r}");
        assert_eq!(fs::read(&path).unwrap(), before);
    }
}

/// AC-6, AC-7 (the dangerous path of the write): a leader entry that is a link and a registry
/// folder that is a link are errors, never an absent leader; a renewal through either is
/// refused and writes nothing, neither into the link nor into its target.
#[cfg(unix)]
#[test]
fn registration_refuses_a_linked_entry_and_a_linked_folder() {
    let f = fixture();
    let real = f.tmp.path().join("real-leader.json");
    fs::write(
        &real,
        json!({
            "schema_version": 1, "name": "leader", "instance": LEADER_INSTANCE,
            "lease_until": NOW_EPOCH - 38 * 3600, "mode": "leader", "session": LEADER_SESSION,
        })
        .to_string(),
    )
    .unwrap();
    let dir = sessions_dir(&f);
    fs::create_dir_all(&dir).unwrap();
    std::os::unix::fs::symlink(&real, dir.join("leader.json")).unwrap();
    let before = fs::read(&real).unwrap();
    let r = round_as(
        &f,
        Some(LEADER_SESSION),
        &["--renew-instance", LEADER_INSTANCE],
    );
    let reg = &r["round"]["registration"];
    assert_eq!(reg["status"], "error", "{r}");
    assert_eq!(reg["renewal"]["status"], "refused", "{r}");
    assert_eq!(r["writes"], false, "{r}");
    assert_eq!(fs::read(&real).unwrap(), before);
    assert!(fs::symlink_metadata(dir.join("leader.json"))
        .unwrap()
        .file_type()
        .is_symlink());

    fs::remove_file(dir.join("leader.json")).unwrap();
    fs::remove_dir(&dir).unwrap();
    let elsewhere = f.tmp.path().join("elsewhere");
    fs::create_dir_all(&elsewhere).unwrap();
    fs::copy(&real, elsewhere.join("leader.json")).unwrap();
    std::os::unix::fs::symlink(&elsewhere, &dir).unwrap();
    let r = round_as(
        &f,
        Some(LEADER_SESSION),
        &["--renew-instance", LEADER_INSTANCE],
    );
    assert_eq!(r["round"]["registration"]["status"], "error", "{r}");
    assert_eq!(
        r["round"]["registration"]["renewal"]["status"], "refused",
        "{r}"
    );
    assert_eq!(r["writes"], false, "{r}");
    assert_eq!(fs::read(elsewhere.join("leader.json")).unwrap(), before);
}

/// Every file below `root` as `(path, bytes)` pairs, sorted.
fn tree_snapshot(root: &Path) -> Vec<(String, Vec<u8>)> {
    fn walk(out: &mut Vec<(String, Vec<u8>)>, dir: &Path, prefix: &str) {
        let mut names: Vec<_> = fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
            .collect();
        names.sort();
        for name in names {
            let path = dir.join(&name);
            let rel = if prefix.is_empty() {
                name.clone()
            } else {
                format!("{prefix}/{name}")
            };
            if path.is_dir() {
                walk(out, &path, &rel);
            } else {
                out.push((rel, fs::read(&path).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(&mut out, root, "");
    out
}
