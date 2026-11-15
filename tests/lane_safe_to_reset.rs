//! Issue #2669: `lane-safe-to-reset` answers `safe` only when every check ran clean, names
//! each finding (uncommitted, unpushed, stashed work, background jobs, a claim without a
//! resume note) as `not_safe`, and says `unknown` with a reason for a check that could not run.
//! Every repository here is a temporary fixture; the binary runs with a cleared environment
//! and an empty temporary HOME.

mod common;
use common::CleanEnv;

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
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.clean_env()
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
    // The folder name match is a whole token: `agent-90` is not in `agent-9-audit-1`, and
    // `gent-9` sits inside it but not on a token boundary.
    for lane in ["agent-90", "gent-9"] {
        let other = run(&f, lane, Some(json!({"schema_version": 1})));
        assert!(
            result(&other)["worktrees"].as_array().unwrap().is_empty(),
            "{lane}: {other}"
        );
    }
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

/// Audit FAIL 6850218012 on PR #2689: a worktree named for the lane without its dash
/// (`…-agent9` for lane `agent-9`), outside every claim and not the session folder, still
/// belongs to the lane; its left work makes the lane not safe.
#[test]
fn a_dashless_lane_name_in_a_folder_or_branch_still_maps() {
    let f = fixture();
    let old = f.primary.join("wt").join("1234-old-agent9");
    g(
        &f.primary,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/1234-old-agent9",
            old.to_str().unwrap(),
            "main",
        ],
    );
    g(&old, &["push", "-q", "origin", "issue/1234-old-agent9"]);
    let v = run(&f, "agent-9", Some(clean_obs(&f)));
    let r = result(&v);
    let mapped = r["worktrees"]
        .as_array()
        .unwrap()
        .iter()
        .find(|w| w["path"].as_str().unwrap().ends_with("1234-old-agent9"))
        .unwrap_or_else(|| panic!("not mapped: {r}"));
    assert_eq!(mapped["matched_by"], json!(["folder name", "branch name"]));
    assert_eq!(r["verdict"], "safe", "{r}");
    fs::write(old.join("pending.txt"), "x\n").unwrap();
    let r = run(&f, "agent-9", Some(clean_obs(&f)));
    assert_eq!(result(&r)["verdict"], "not_safe", "{r}");
    assert_eq!(
        texts(result(&r), "findings"),
        vec![format!("{}: 1 uncommitted change(s)", old.display())]
    );
}

/// #2669 comment 6851881955 (leader's first use): a lane whose session folder is the primary
/// checkout does not own that checkout's changes — they are listed as `unattributed` and do not
/// make the lane not safe; the same changes in a lane's own worktree still do.
#[test]
fn the_primary_checkout_as_session_folder_is_not_the_lanes_work() {
    let f = fixture();
    fs::write(f.primary.join("user-notes.txt"), "x\n").unwrap();
    let mut obs = clean_obs(&f);
    obs["session_cwd"]["path"] = json!(f.primary);
    let v = run(&f, "agent-9", Some(obs));
    let r = result(&v);
    assert_eq!(r["verdict"], "safe", "{r}");
    let unattributed = texts(r, "unattributed");
    assert_eq!(unattributed.len(), 1, "{r}");
    assert!(unattributed[0].ends_with("uncommitted change(s)"), "{r}");
    let primary = r["worktrees"]
        .as_array()
        .unwrap()
        .iter()
        .find(|w| w["matched_by"] == json!(["session_cwd"]))
        .unwrap_or_else(|| panic!("{r}"));
    assert_eq!(primary["attributed"], false);
    // When the primary checkout is also on a branch the lane claims, it is the lane's work.
    g(&f.primary, &["switch", "-q", "-c", "issue/9-in-primary"]);
    let mut obs = clean_obs(&f);
    obs["session_cwd"]["path"] = json!(f.primary);
    obs["claims"]["items"]
        .as_array_mut()
        .unwrap()
        .push(json!({"issue": 9, "claim_comment": 90, "resume_note": 91}));
    let r = run(&f, "agent-9", Some(obs));
    assert_eq!(result(&r)["verdict"], "not_safe", "{r}");
    assert!(texts(result(&r), "unattributed").is_empty(), "{r}");
    g(&f.primary, &["switch", "-q", "main"]);
    // The lane's own worktree is still judged.
    fs::write(f.wt7.join("draft.txt"), "x\n").unwrap();
    let mut obs = clean_obs(&f);
    obs["session_cwd"]["path"] = json!(f.primary);
    let r = run(&f, "agent-9", Some(obs));
    assert_eq!(result(&r)["verdict"], "not_safe", "{r}");
}

/// Codex review on PR #2715: a bare repository has no primary checkout, so its first linked
/// worktree, tied to the lane only by the session folder, is still the lane's work.
#[test]
fn a_bare_repositorys_first_worktree_is_not_a_primary_checkout() {
    let f = fixture();
    let root = f.tmp.path().canonicalize().unwrap();
    git_ok(&root, &["clone", "-q", "--bare", "remote.git", "bare.git"]);
    let bare = root.join("bare.git");
    let wt = root.join("detached-wt");
    g(
        &bare,
        &[
            "worktree",
            "add",
            "-q",
            "--detach",
            wt.to_str().unwrap(),
            "main",
        ],
    );
    fs::write(wt.join("pending.txt"), "x\n").unwrap();
    let home = f.tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let obs = f.tmp.path().join("bare-obs.json");
    fs::write(
        &obs,
        json!({"schema_version": 1, "session_cwd": {"source": "fixture", "path": wt}}).to_string(),
    )
    .unwrap();
    let out = Command::new(crate::common::MOAI_BIN)
        .clean_env()
        .env("PATH", std::env::var_os("PATH").unwrap())
        .env("HOME", &home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .args([
            "lane-safe-to-reset",
            bare.to_str().unwrap(),
            "--lane",
            "agent-9",
            "--now",
            NOW,
        ])
        .arg("--observations")
        .arg(&obs)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    let r = result(&v);
    assert_eq!(r["verdict"], "not_safe", "{r}");
    assert!(texts(r, "unattributed").is_empty(), "{r}");
    assert!(
        texts(r, "findings")[0].ends_with("1 uncommitted change(s)"),
        "{r}"
    );
}

/// #2811 AC-3: a lane with one worktree in the old place (inside the primary checkout) and one
/// in the new place `<HOME>/.moai/worktrees/<project>-<id>` (outside it) gets both mapped,
/// left work in the new-place one is a finding, and a session folder inside the new-place
/// worktree belongs to that worktree, not to the primary checkout.
#[test]
fn old_and_new_place_worktrees_of_a_lane_are_both_checked() {
    let f = fixture();
    let home = f.tmp.path().canonicalize().unwrap().join("home");
    let new_wt = home.join(".moai/worktrees/primary-8-next");
    g(
        &f.primary,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/8-next",
            new_wt.to_str().unwrap(),
            "main",
        ],
    );
    g(&new_wt, &["push", "-q", "origin", "issue/8-next"]);
    fs::create_dir_all(new_wt.join("sub")).unwrap();
    let mut obs = clean_obs(&f);
    obs["claims"]["items"] = json!([
        {"issue": 7, "claim_comment": 70, "resume_note": 71},
        {"issue": 8, "claim_comment": 80, "resume_note": 81}
    ]);
    obs["session_cwd"]["path"] = json!(new_wt.join("sub"));
    let r = run(&f, "agent-9", Some(obs.clone()));
    let res = result(&r);
    assert_eq!(res["verdict"], "safe", "{res}");
    let by_path = |p: &Path| {
        res["worktrees"]
            .as_array()
            .unwrap()
            .iter()
            .find(|w| w["path"] == p.to_str().unwrap())
            .unwrap_or_else(|| panic!("no {}: {res}", p.display()))
            .clone()
    };
    assert_eq!(by_path(&f.wt7)["matched_by"], json!(["claim #7"]));
    assert_eq!(
        by_path(&new_wt)["matched_by"],
        json!(["claim #8", "session_cwd"])
    );
    // Uncommitted work in the new place is the lane's finding.
    fs::write(new_wt.join("draft.txt"), "x\n").unwrap();
    let r = run(&f, "agent-9", Some(obs));
    assert_eq!(
        texts(result(&r), "findings"),
        vec![format!("{}: 1 uncommitted change(s)", new_wt.display())]
    );
}

// --- #3015 C1: the tool reads the lane's claims itself; the caller gives only the screen ----

const CLAIM_70: &str = "moai-claim: session=agent-9 epoch=1 lease_until=2026-11-15T08:00:00Z";
const NOTE_71: &str = "moai-resume: lane=agent-9 issue=7 claim=6000070 lease_until=2026-11-15T08:00:00Z branch=issue/7-card head=0123456789abcdef0123456789abcdef01234567 uncommitted=0\nnext: push";

fn gh_comment(id: u64, login: &str, body: &str) -> Value {
    json!({"id": id, "user": {"login": login}, "body": body})
}

/// What a stub `gh` answers: the search result, Issue 7's comments (`None` answers 404) and the
/// repository's open pull requests (`None` answers an empty list, JSON `null` answers 404).
#[derive(Default)]
struct Forge {
    search: Value,
    comments_7: Option<Value>,
    pulls: Option<Value>,
}

fn forge_with(comments: Vec<Value>) -> Forge {
    Forge {
        search: json!({"total_count": 1, "incomplete_results": false, "items": [{"number": 7}]}),
        comments_7: Some(json!(comments)),
        ..Default::default()
    }
}

/// Runs the binary with flags only (no observations file) over a stub `gh` that logs every
/// request. Returns the answer and the logged request lines.
fn run_flags(
    f: &Fixture,
    lane: &str,
    flags: &[&str],
    forge: Option<&Forge>,
) -> (Value, Vec<String>) {
    let home = f.tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let gh = f.tmp.path().join("gh-stub");
    let log = f.tmp.path().join("gh.log");
    let _ = fs::remove_file(&log);
    common::write_executable(
        &gh,
        "#!/bin/sh\necho \"$@\" >> \"$GH_LOG\"\nfor a; do url=$a; done\ncase \"$url\" in\n\
         */search/issues*) body=$SEARCH_JSON ;;\n\
         */issues/7/comments*) [ -n \"$COMMENTS_7\" ] && body=$COMMENTS_7 || { printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{\"message\":\"Not Found\"}'; exit 1; } ;;\n\
         */pulls*) if [ \"$PULLS_JSON\" = null ]; then printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{\"message\":\"Not Found\"}'; exit 1; fi; body=${PULLS_JSON:-[]} ;;\n\
         *) printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{\"message\":\"Not Found\"}'; exit 1 ;;\nesac\n\
         printf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$body\"\n",
    );
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("MOAI_FORGE_GH_BIN", &gh)
        .env("GH_LOG", &log);
    if let Some(forge) = forge {
        cmd.env("SEARCH_JSON", forge.search.to_string());
        if let Some(c) = &forge.comments_7 {
            cmd.env("COMMENTS_7", c.to_string());
        }
        if let Some(p) = &forge.pulls {
            cmd.env("PULLS_JSON", p.to_string());
        }
    }
    cmd.args([
        "lane-safe-to-reset",
        f.primary.to_str().unwrap(),
        "--lane",
        lane,
        "--now",
        NOW,
    ])
    .args(flags);
    let out = cmd.output().unwrap();
    let v = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stdout)));
    let requests = fs::read_to_string(&log)
        .unwrap_or_default()
        .lines()
        .map(str::to_string)
        .collect();
    (v, requests)
}

fn clean_flags(f: &Fixture) -> Vec<String> {
    vec![
        "--repo".into(),
        "acme/widgets".into(),
        "--screen-jobs".into(),
        "0".into(),
        "--session-cwd".into(),
        f.wt7.to_str().unwrap().into(),
    ]
}

fn as_strs(v: &[String]) -> Vec<&str> {
    v.iter().map(String::as_str).collect()
}

#[cfg(unix)]
#[test]
fn without_an_observations_file_a_lane_with_a_resume_note_and_a_quiet_screen_is_safe() {
    let f = fixture();
    let forge = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    let flags = clean_flags(&f);
    let (r, requests) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    let res = result(&r);
    assert_eq!(res["verdict"], "safe", "{res}");
    assert_eq!(
        res["claims"]["items"],
        json!([{"issue": 7, "claim_comment": 6000070, "resume_note": 6000071}]),
        "{res}"
    );
    assert_eq!(res["claims"]["source"], "github:search+issue-comments");
    assert_eq!(res["screen"]["source"], "flag");
    assert_eq!(res["session_cwd"]["source"], "flag");
    // Reads only: a search, the claimed thread's comments and the open pull requests (for the
    // lane's state), every one a GET.
    assert_eq!(requests.len(), 3, "{requests:?}");
    assert!(
        requests.iter().all(|l| l.starts_with("api -X GET ")),
        "{requests:?}"
    );
}

// --- #3015 C3b: the lane's state is told from its claims, pull requests and work, not from
// a quiet screen ----

fn lane_state(r: &Value) -> &Value {
    &result(r)["lane_state"]
}

fn open_pr_7() -> Value {
    json!([{"number": 41, "head": {"ref": "issue/7-card"}, "body": "audit-tier: 1"}])
}

#[cfg(unix)]
#[test]
fn a_claim_with_an_open_pull_request_is_waiting_and_one_without_is_resumable() {
    let f = fixture();
    let mut forge = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    let flags = clean_flags(&f);
    // No pull request, a resume note: stopped, and it can be taken up again.
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "resumable", "{r}");
    assert_eq!(
        lane_state(&r)["evidence"]["open_claims"][0]["open_pr"],
        Value::Null
    );
    // The claim's own pull request is open: the lane waits for its check or audit.
    forge.pulls = Some(open_pr_7());
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "waiting", "{r}");
    assert_eq!(
        lane_state(&r)["evidence"]["open_claims"][0]["open_pr"],
        41,
        "{r}"
    );
    // The same pull request found by the claim line in its body, not by its branch.
    forge.pulls = Some(json!([{
        "number": 42, "head": {"ref": "somebody-elses-name"},
        "body": "audit-tier: 1\n\nMoai-Claim: 7#6000070"
    }]));
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "waiting", "{r}");
    assert_eq!(
        lane_state(&r)["evidence"]["open_claims"][0]["open_pr"],
        42,
        "{r}"
    );
    // A pull request of another card does not count.
    forge.pulls = Some(json!([{"number": 43, "head": {"ref": "issue/8-other"}, "body": ""}]));
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "resumable", "{r}");
}

#[cfg(unix)]
#[test]
fn a_claim_with_neither_a_note_nor_a_pull_request_is_stopped_and_a_job_is_working() {
    let f = fixture();
    let no_note = forge_with(vec![gh_comment(6000070, "GoosLab", CLAIM_70)]);
    let flags = clean_flags(&f);
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&no_note));
    assert_eq!(lane_state(&r)["state"], "stopped", "{r}");
    // A background job on the screen is a running writer, whatever else is on record.
    let with_note = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    let mut busy = clean_flags(&f);
    busy[3] = "2".into();
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&busy), Some(&with_note));
    assert_eq!(lane_state(&r)["state"], "working", "{r}");
}

#[cfg(unix)]
#[test]
fn a_lane_with_no_claim_and_nothing_left_is_idle_and_left_work_makes_it_stopped() {
    let f = fixture();
    let mut forge = forge_with(vec![]);
    forge.search = json!({"total_count": 0, "incomplete_results": false, "items": []});
    let flags = clean_flags(&f);
    let (r, requests) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "idle", "{r}");
    // No claim, no pull-request read.
    assert!(
        requests.iter().all(|l| !l.contains("/pulls")),
        "{requests:?}"
    );
    // Work left in the lane's worktree with no claim to say where it belongs.
    fs::write(f.wt7.join("draft.txt"), "x\n").unwrap();
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "stopped", "{r}");
}

#[cfg(unix)]
#[test]
fn a_worktree_that_cannot_be_checked_is_named_in_the_unknown_state() {
    let f = fixture();
    let forge = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    // The lane's worktree folder is gone but git still lists it: its work cannot be counted.
    fs::remove_dir_all(&f.wt7).unwrap();
    let flags: Vec<String> = vec![
        "--repo".into(),
        "acme/widgets".into(),
        "--screen-jobs".into(),
        "0".into(),
    ];
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
    let reasons = texts(lane_state(&r), "reasons");
    assert!(
        reasons
            .iter()
            .any(|t| t.contains("could not be checked for unsaved work")),
        "{reasons:?}"
    );
}

#[cfg(unix)]
#[test]
fn claims_given_in_a_file_read_no_pull_requests_and_the_state_says_so() {
    let f = fixture();
    let forge = forge_with(vec![]);
    let file = f.tmp.path().join("claims.json");
    fs::write(
        &file,
        json!({
            "schema_version": 1,
            "claims": {"source": "fixture", "items": [
                {"issue": 7, "claim_comment": 6000070, "resume_note": 6000071}
            ]}
        })
        .to_string(),
    )
    .unwrap();
    let flags: Vec<String> = vec![
        "--observations".into(),
        file.to_str().unwrap().into(),
        "--repo".into(),
        "acme/widgets".into(),
        "--screen-jobs".into(),
        "0".into(),
        "--session-cwd".into(),
        f.wt7.to_str().unwrap().into(),
    ];
    let (r, requests) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert!(requests.is_empty(), "no request at all: {requests:?}");
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
    let reasons = texts(lane_state(&r), "reasons");
    assert!(
        reasons
            .iter()
            .any(|t| t.contains("open pull requests were not read")),
        "{reasons:?}"
    );
}

#[cfg(unix)]
#[test]
fn a_full_page_of_pull_requests_may_hide_the_claims_own() {
    let f = fixture();
    let mut forge = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    let flags = clean_flags(&f);
    let others = |n: usize| -> Vec<Value> {
        (0..n)
            .map(|i| json!({"number": 100 + i, "head": {"ref": format!("issue/9{i:02}-x")}, "body": ""}))
            .collect()
    };
    // 99 pull requests of other cards: the read is whole, the claim has none.
    forge.pulls = Some(json!(others(99)));
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "resumable", "{r}");
    // 100 (a full page): a pull request of the claim may be on the next page.
    forge.pulls = Some(json!(others(100)));
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
    let reasons = texts(lane_state(&r), "reasons");
    assert!(
        reasons.iter().any(|t| t.contains("read only in part")),
        "{reasons:?}"
    );
    // The claim's own pull request is on the full page: nothing is hidden for it.
    let mut page = others(99);
    page.push(json!({"number": 41, "head": {"ref": "issue/7-card"}, "body": ""}));
    forge.pulls = Some(json!(page));
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "waiting", "{r}");
}

#[cfg(unix)]
#[test]
fn a_read_that_failed_makes_the_state_unknown_and_never_a_guess() {
    let f = fixture();
    let mut forge = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    // The pull requests cannot be read: neither "waiting" nor "resumable" can be told.
    forge.pulls = Some(Value::Null);
    let flags = clean_flags(&f);
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
    let reasons = texts(lane_state(&r), "reasons");
    assert!(
        reasons.iter().any(|t| t.contains("open pull requests")),
        "{reasons:?}"
    );
    // No screen observation: a running writer cannot be ruled out.
    forge.pulls = None;
    let no_screen: Vec<String> = vec![
        "--repo".into(),
        "acme/widgets".into(),
        "--session-cwd".into(),
        f.wt7.to_str().unwrap().into(),
    ];
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&no_screen), Some(&forge));
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
}

#[cfg(unix)]
#[test]
fn a_claim_without_a_resume_note_or_a_running_job_makes_the_lane_not_safe() {
    let f = fixture();
    let no_note = forge_with(vec![gh_comment(6000070, "GoosLab", CLAIM_70)]);
    let flags = clean_flags(&f);
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&no_note));
    assert_eq!(result(&r)["verdict"], "not_safe", "{r}");
    assert_eq!(
        texts(result(&r), "findings"),
        vec!["open claim #7 (comment 6000070) has no resume note"]
    );
    // A job on the screen.
    let with_note = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "GoosLab", NOTE_71),
    ]);
    let mut busy = clean_flags(&f);
    busy[3] = "2".into();
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&busy), Some(&with_note));
    assert_eq!(
        texts(result(&r), "findings"),
        vec!["2 background job(s) still running on the lane's screen"],
        "{r}"
    );
    // A resume note by someone else is no resume note.
    let forged = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000071, "Stranger", NOTE_71),
    ]);
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&forged));
    assert_eq!(result(&r)["verdict"], "not_safe", "{r}");
    // A released claim is no open claim: the lane has nothing left to note.
    let released = forge_with(vec![
        gh_comment(6000070, "GoosLab", CLAIM_70),
        gh_comment(6000072, "GoosLab", "moai-claim-release: comment=6000070"),
    ]);
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&released));
    assert_eq!(result(&r)["claims"]["items"], json!([]), "{r}");
}

#[cfg(unix)]
#[test]
fn claims_that_cannot_be_read_leave_the_lane_unknown_with_the_reason() {
    let f = fixture();
    let flags = clean_flags(&f);
    // The comments of the found thread cannot be read.
    let broken = Forge {
        search: json!({"total_count": 1, "incomplete_results": false, "items": [{"number": 7}]}),
        comments_7: None,
        ..Default::default()
    };
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&broken));
    assert_eq!(result(&r)["verdict"], "unknown", "{r}");
    let unknowns = texts(result(&r), "unknowns");
    assert_eq!(unknowns.len(), 1, "{unknowns:?}");
    assert!(
        unknowns[0].starts_with("claims could not be read from GitHub: the comments of #7"),
        "{unknowns:?}"
    );
    // The project's origin is a local path, not a GitHub repository, and no --repo is given.
    let (r, requests) = run_flags(
        &f,
        "agent-9",
        &[
            "--screen-jobs",
            "0",
            "--session-cwd",
            f.wt7.to_str().unwrap(),
        ],
        None,
    );
    assert_eq!(result(&r)["verdict"], "unknown", "{r}");
    assert!(
        texts(result(&r), "unknowns")[0].contains("not a GitHub repository"),
        "{r}"
    );
    assert!(
        requests.is_empty(),
        "no request without a repository: {requests:?}"
    );
    // A --repo that is not OWNER/REPO is refused before any request.
    let (r, requests) = run_flags(
        &f,
        "agent-9",
        &[
            "--repo",
            "acme/widgets/extra",
            "--screen-jobs",
            "0",
            "--session-cwd",
            f.wt7.to_str().unwrap(),
        ],
        None,
    );
    assert_eq!(result(&r)["verdict"], "unknown", "{r}");
    assert!(
        texts(result(&r), "unknowns")[0].contains("invalid --repo `acme/widgets/extra`"),
        "{r}"
    );
    assert!(requests.is_empty(), "{requests:?}");
    // A search that shows fewer threads than it found is reported, never read as complete.
    let cut = Forge {
        search: json!({"total_count": 31, "incomplete_results": false, "items": [{"number": 7}]}),
        comments_7: Some(json!([
            gh_comment(6000070, "GoosLab", CLAIM_70),
            gh_comment(6000071, "GoosLab", NOTE_71)
        ])),
        ..Default::default()
    };
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&cut));
    assert_eq!(result(&r)["verdict"], "unknown", "{r}");
    assert!(
        texts(result(&r), "unknowns")[0].starts_with("claims read only in part:"),
        "{r}"
    );
    // The state says the same: it never calls the lane resumable (the listed claim has a note)
    // when a claim it could not see may be open (#3051 audit B1).
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
    assert!(
        texts(lane_state(&r), "reasons")
            .iter()
            .any(|t| t.contains("open claims were read only in part")),
        "{r}"
    );
    // The search found 31 threads and showed none: no claim is listed, yet the lane is not idle.
    let hidden = Forge {
        search: json!({"total_count": 31, "incomplete_results": false, "items": []}),
        ..Default::default()
    };
    let (r, _) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&hidden));
    assert_eq!(lane_state(&r)["state"], "unknown", "{r}");
}

#[cfg(unix)]
#[test]
fn a_claim_the_search_has_not_indexed_yet_is_found_through_the_lanes_local_branch() {
    // Issue 7's branch (`issue/7-card`) is in a worktree named for the lane after the
    // fixture renames nothing: the folder `wt/7-card` does not carry the lane, so use a
    // second worktree whose folder does.
    let f = fixture();
    let lane_wt = f.primary.join("wt").join("agent-9-card");
    g(
        &f.primary,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/7-lane",
            lane_wt.to_str().unwrap(),
            "main",
        ],
    );
    g(&lane_wt, &["push", "-q", "origin", "issue/7-lane"]);
    let lagging = Forge {
        search: json!({"total_count": 0, "incomplete_results": false, "items": []}),
        comments_7: Some(json!([
            gh_comment(6000070, "GoosLab", CLAIM_70),
            gh_comment(6000071, "GoosLab", NOTE_71)
        ])),
        ..Default::default()
    };
    let flags = clean_flags(&f);
    let (r, requests) = run_flags(&f, "agent-9", &as_strs(&flags), Some(&lagging));
    assert_eq!(
        result(&r)["claims"]["items"],
        json!([{"issue": 7, "claim_comment": 6000070, "resume_note": 6000071}]),
        "{r}"
    );
    assert_eq!(requests.len(), 3, "{requests:?}");
}

#[test]
fn a_part_given_twice_or_in_the_wrong_shape_is_refused_with_the_schema_and_an_example() {
    let f = fixture();
    // The screen in the file and as a flag.
    let both = run_flags_with_file(
        &f,
        json!({
            "schema_version": 1,
            "screen": {"source": "fixture", "observed_at": "2026-11-13T08:55:00Z", "background_jobs": 0}
        }),
        &["--screen-jobs", "0"],
    );
    assert_eq!(both["ok"], false, "{both}");
    assert!(
        both["error"]["message"]
            .as_str()
            .unwrap()
            .contains("given twice"),
        "{both}"
    );
    // The session folder in the file and as a flag.
    let cwd = f.wt7.to_str().unwrap();
    let both = run_flags_with_file(
        &f,
        json!({"schema_version": 1, "session_cwd": {"source": "fixture", "path": cwd}}),
        &["--session-cwd", cwd],
    );
    assert_eq!(both["ok"], false, "{both}");
    assert!(
        both["error"]["message"]
            .as_str()
            .unwrap()
            .contains("session_cwd is given twice"),
        "{both}"
    );
    // A list where a number belongs: the message carries the field types and a full example.
    let wrong = run_flags_with_file(
        &f,
        json!({
            "schema_version": 1,
            "screen": {"source": "fixture", "observed_at": "2026-11-13T08:55:00Z", "background_jobs": []}
        }),
        &[],
    );
    assert_eq!(wrong["ok"], false, "{wrong}");
    assert_eq!(wrong["error"]["kind"], "invalid_input", "{wrong}");
    let msg = wrong["error"]["message"].as_str().unwrap();
    assert!(msg.contains("expected u64"), "{msg}");
    assert!(msg.contains("screen.background_jobs: a number"), "{msg}");
    assert!(msg.contains("\"background_jobs\":0"), "{msg}");
}

/// The binary with an observations file and extra flags, no `gh` reachable.
fn run_flags_with_file(f: &Fixture, obs: Value, flags: &[&str]) -> Value {
    let home = f.tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    let out = Command::new(crate::common::MOAI_BIN)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", &home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .args([
            "lane-safe-to-reset",
            f.primary.to_str().unwrap(),
            "--lane",
            "agent-9",
            "--now",
            NOW,
            "--observations",
        ])
        .arg(observations(f, obs))
        .args(flags)
        .output()
        .unwrap();
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|e| panic!("{e}: {}", String::from_utf8_lossy(&out.stdout)))
}

#[test]
fn help_shows_the_observations_shape_and_an_example() {
    let out = Command::new(crate::common::MOAI_BIN)
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .args(["lane-safe-to-reset", "--help"])
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout).into_owned();
    assert!(text.contains("--screen-jobs"), "{text}");
    assert!(text.contains("--session-cwd"), "{text}");
    assert!(text.contains("\"background_jobs\":0"), "{text}");
    assert!(text.contains("screen.background_jobs: a number"), "{text}");
}
