//! rc.2 A3 (#2162) `leader-tick` over recorded exchanges (`RecordedTransport`, injected
//! clock, no network), in the fixture style of `tests/claim_cli.rs`. The state file is real
//! (a temporary project), so "report only what changed" is tested across ticks.
use moai_cli::forge::policy::Clock;
mod common;

use moai_cli::forge::transport::{Exchange, RecordedTransport};
use moai_cli::leader::{tick_cli, TickRequest};
use serde_json::{json, Value};
use std::{cell::Cell, fs, path::Path};

const API: &str = "https://api.github.com/repos/acme/widgets";
const LEASE: &str = "2030-01-01T00:00:00Z";

struct FakeClock(Cell<i64>);
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.0.get()
    }
    fn sleep(&self, secs: u64) {
        self.0.set(self.0.get() + secs as i64);
    }
}

fn get(url: &str, body: Value) -> Exchange {
    common::exchange("GET", url, common::json_response(200, body))
}

/// One open PR #5 (`Refs #7`) at `head`, claimed by `author`, with the given verdict.
fn one_pr(head: &str, author: &str, verdict: &str, ci: &str) -> Vec<Exchange> {
    vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([{"number": 5, "head": {"sha": head}, "body": "x\n\nRefs #7\n"}]),
        ),
        get(
            &format!("{API}/pulls/5"),
            json!({"mergeable_state": "clean"}),
        ),
        get(
            &format!("{API}/commits/{head}/status"),
            json!({"statuses": [{"context": "moai/audit", "state": "success"}]}),
        ),
        get(
            &format!("{API}/commits/{head}/check-runs"),
            json!({"check_runs": [{"name": "rc-ci", "status": "completed", "conclusion": ci}]}),
        ),
        get(
            &format!("{API}/issues/7/comments?per_page=100"),
            json!([
                {"id": 100, "body": format!("moai-claim: session={author} epoch=1 lease_until={LEASE}")},
                {"id": 101, "body": format!("AUDIT #7 (A1) head {head} base b — verdict: {verdict}")},
            ]),
        ),
    ]
}

fn request(lanes: &[&str]) -> TickRequest {
    TickRequest {
        schema_version: 1,
        owner: "acme".into(),
        repo: "widgets".into(),
        base_branch: "v2026.11.0".into(),
        lanes: lanes.iter().map(|l| l.to_string()).collect(),
    }
}

fn tick(project: &Path, lanes: &[&str], exchanges: Vec<Exchange>) -> Value {
    let t = RecordedTransport::new(exchanges);
    let clock = FakeClock(Cell::new(1_794_000_000));
    let out = tick_cli(project, request(lanes), &t, &clock).unwrap();
    assert_eq!(t.remaining(), 0, "every recorded exchange was used");
    assert!(
        t.sent.borrow().iter().all(|r| r.method == "GET"),
        "GET only"
    );
    out
}

/// AC-1: the first tick reports the PR; an identical second tick reports nothing and keeps
/// the state byte-identical; a new head is reported again, for that PR only.
#[test]
fn unchanged_prs_are_not_reported_and_changes_are() {
    let tmp = tempfile::tempdir().unwrap();
    let p = tmp.path();
    let first = tick(p, &[], one_pr("h1", "agent-1", "PASS", "success"));
    assert_eq!(first["changes"].as_array().unwrap().len(), 1, "{first}");
    assert_eq!(first["changes"][0]["now"]["author"], json!("agent-1"));
    assert_eq!(first["changes"][0]["now"]["verdict"], json!("PASS"));
    // #2678: the count is said in words, and "no change" is said outright.
    assert_eq!(
        first["summary"],
        json!("1 open pull request(s) read; 1 changed since the previous tick")
    );
    let state = p.join(".moai-cli/forge/leader-tick.json");
    let saved = fs::read(&state).unwrap();

    let second = tick(p, &[], one_pr("h1", "agent-1", "PASS", "success"));
    assert_eq!(second["changes"], json!([]), "{second}");
    assert_eq!(second["hints"], json!([]));
    assert_eq!(
        second["summary"],
        json!("1 open pull request(s) read; no change since the previous tick")
    );
    assert_eq!(fs::read(&state).unwrap(), saved, "state is byte-identical");

    let third = tick(p, &[], one_pr("h2", "agent-1", "none", "success"));
    let changes = third["changes"].as_array().unwrap();
    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0]["pr"], json!(5));
    assert_eq!(changes[0]["before"]["head"], json!("h1"));
    assert_eq!(changes[0]["now"]["head"], json!("h2"));
}

/// AC-2: a FAIL verdict on the head drafts a hint for the first free lane that is not the
/// author; with only the author free there is no hint target.
#[test]
fn repair_hint_excludes_the_author_lane() {
    let tmp = tempfile::tempdir().unwrap();
    let out = tick(
        tmp.path(),
        &["agent-1", "agent-3"],
        one_pr("h1", "agent-1", "FAIL", "success"),
    );
    let hints = out["hints"].as_array().unwrap();
    assert_eq!(hints.len(), 1, "{out}");
    assert_eq!(hints[0]["hint_to"], json!("agent-3"));
    assert_eq!(hints[0]["author"], json!("agent-1"));
    assert!(hints[0]["comment_body"]
        .as_str()
        .unwrap()
        .starts_with("assignee-hint: agent-3"));

    let tmp = tempfile::tempdir().unwrap();
    let out = tick(
        tmp.path(),
        &["agent-1"],
        one_pr("h1", "agent-1", "none", "failure"),
    );
    assert_eq!(out["hints"][0]["hint_to"], Value::Null, "{out}");
}

/// AC-3: the interval is 300 by default and the profile's `leader.pr_check_interval_secs`
/// when set; the tick reports which.
#[test]
fn interval_comes_from_the_profile_or_the_default() {
    let tmp = tempfile::tempdir().unwrap();
    let out = tick(tmp.path(), &[], one_pr("h1", "agent-1", "PASS", "success"));
    assert_eq!(out["interval_secs"], json!(300));
    assert_eq!(out["interval_source"], json!("default"));

    let tmp = tempfile::tempdir().unwrap();
    fs::create_dir_all(tmp.path().join(".moai-cli")).unwrap();
    fs::write(
        tmp.path().join(".moai-cli/project.yaml"),
        json!({"schema_version": 2, "project_id": "demo", "purpose": "p", "tools": ["Rust"],
               "development_methods": ["TDD"], "hosts": ["claude"],
               "leader": {"pr_check_interval_secs": 60}})
        .to_string(),
    )
    .unwrap();
    let out = tick(tmp.path(), &[], one_pr("h1", "agent-1", "PASS", "success"));
    assert_eq!(out["interval_secs"], json!(60));
    assert_eq!(out["interval_source"], json!("profile"));
}

/// #2563: the CI check renamed `rc-ci` → `ci`. The tick reads either spelling and the
/// newest run (largest id) decides, whatever order the API lists them in; a matrix job
/// name never counts.
#[test]
fn ci_reads_both_names_and_the_newest_run_decides() {
    let with_runs = |runs: Value| {
        let mut ex = one_pr("h1", "agent-1", "PASS", "success");
        ex[3] = get(
            &format!("{API}/commits/h1/check-runs"),
            json!({ "check_runs": runs }),
        );
        ex
    };
    let run = |id: i64, name: &str, c: &str| json!({"id": id, "name": name, "status": "completed", "conclusion": c});
    for (runs, ci) in [
        (json!([run(1, "ci", "success")]), "success"),
        (json!([run(1, "rc-ci", "failure")]), "failure"),
        (
            json!([run(2, "ci", "failure"), run(1, "rc-ci", "success")]),
            "failure",
        ),
        (
            json!([run(1, "rc-ci", "success"), run(2, "ci", "failure")]),
            "failure",
        ),
        (
            json!([run(1, "ci", "failure"), run(2, "rc-ci", "success")]),
            "success",
        ),
        (json!([run(1, "ci / linux-x86_64", "success")]), "none"),
    ] {
        let tmp = tempfile::tempdir().unwrap();
        let out = tick(tmp.path(), &[], with_runs(runs.clone()));
        assert_eq!(out["changes"][0]["now"]["ci"], json!(ci), "{runs} → {out}");
    }
}
