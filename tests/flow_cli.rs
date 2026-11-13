//! Flow control fixtures (A13 #2279). AC-2: the WIP cap refuses at the boundary and
//! allows below it. AC-3: the integration gate verdict checks tip containment,
//! changed-since-audit, claim lines (merge-exempt), the PR body claim, CI status and
//! head stability — each check observed failing at least once. All GitHub reads are
//! replayed through a stub `gh` binary that answers from environment-set canned HTTP
//! envelopes (read-only; no status posting, merging or ruleset change is exercised).

mod common;

use serde_json::{json, Value};
use std::{
    path::{Path, PathBuf},
    process::{Command, Stdio},
};
use tempfile::TempDir;

const EXE: &str = env!("CARGO_BIN_EXE_moai-cli");

/// The HTTP envelope `parse_gh_output` expects: a status line, an empty header block,
/// then the JSON body. Status 200 keeps the forge client happy on every read.
fn http200(body: &str) -> String {
    format!("HTTP/1.1 200 OK\n\n{body}")
}

/// The stub `gh` every fixture uses. It dispatches on the request URL (the full
/// `gh api -X GET -i … <url>` argument list arrives as `$*`) and prints the
/// environment-supplied envelope for the matched endpoint. Pattern order matters:
/// `pulls/9/commits` must win over `pulls/9`, and the tip's check runs (#2665) over the
/// head's; without `FLOW_TIP_CHECK_RUNS` the tip shares the head's runs.
const STUB_SCRIPT: &str = r##"#!/bin/sh
case "$*" in
  *commits/tip-*/check-runs*) printf '%s' "${FLOW_TIP_CHECK_RUNS:-$FLOW_CHECK_RUNS}";;
  *pulls/9/commits*) printf '%s' "$FLOW_COMMITS";;
  *pulls/9*) printf '%s' "$FLOW_PR";;
  *pulls?state=open*) printf '%s' "$FLOW_OPEN_PRS";;
  *compare/*) printf '%s' "$FLOW_COMPARE";;
  *commits/*/check-runs*) printf '%s' "$FLOW_CHECK_RUNS";;
  *commits/*/status*) printf '%s' "$FLOW_STATUS";;
  *) printf '%s' "$FLOW_TIP";;
esac
"##;

/// Same endpoints, but the PR is re-read with a moving head: the first `pulls/9`
/// read answers FLOW_PR, every later one answers FLOW_PR2 (counted via a file whose
/// absolute path the fixture supplies).
const HEAD_MOVE_STUB: &str = r##"#!/bin/sh
n=$(cat "$FLOW_COUNT_FILE" 2>/dev/null || echo 0)
n=$((n + 1))
printf '%s' "$n" > "$FLOW_COUNT_FILE"
case "$*" in
  *commits/tip-*/check-runs*) printf '%s' "${FLOW_TIP_CHECK_RUNS:-$FLOW_CHECK_RUNS}";;
  *pulls/9/commits*) printf '%s' "$FLOW_COMMITS";;
  *pulls/9*) if [ "$n" -ge 2 ]; then printf '%s' "$FLOW_PR2"; else printf '%s' "$FLOW_PR"; fi;;
  *pulls?state=open*) printf '%s' "$FLOW_OPEN_PRS";;
  *compare/*) printf '%s' "$FLOW_COMPARE";;
  *commits/*/check-runs*) printf '%s' "$FLOW_CHECK_RUNS";;
  *commits/*/status*) printf '%s' "$FLOW_STATUS";;
  *) printf '%s' "$FLOW_TIP";;
esac
"##;

/// The required check appears only on the last page, so truncating a successful
/// ten-page response or accepting an eleven-page response changes the verdict.
const PAGINATED_CHECKS_STUB: &str = r##"#!/bin/sh
case "$*" in
  *commits/tip-*/check-runs*) printf 'HTTP/1.1 200 OK\n\n%s' "$FLOW_LAST_CHECKS";;
  *pulls/9/commits*) printf '%s' "$FLOW_COMMITS";;
  *pulls/9*) printf '%s' "$FLOW_PR";;
  *compare/*) printf '%s' "$FLOW_COMPARE";;
  *commits/*/check-runs*)
    n=$(cat "$FLOW_PAGE_COUNT" 2>/dev/null || echo 0)
    n=$((n + 1))
    printf '%s' "$n" > "$FLOW_PAGE_COUNT"
    printf 'HTTP/1.1 200 OK\n'
    if [ "$n" -lt "$FLOW_CHECK_PAGES" ]; then
      printf 'link: <https://api.github.com/repos/acme/widgets/commits/pr-5a4b3c2d/check-runs?per_page=100&page=%s>; rel="next"\n' "$((n + 1))"
      printf '\n%s' '{"check_runs":[]}'
    else
      printf '\n%s' "$FLOW_LAST_CHECKS"
    fi;;
  *) printf '%s' "$FLOW_TIP";;
esac
"##;

fn install_stub(dir: &Path, script: &str) -> PathBuf {
    let path = dir.join("gh-stub");
    common::write_executable(&path, script);
    path
}

fn run_flow(gh: &Path, envs: &[(&str, String)], args: &[&str]) -> (Value, std::process::Output) {
    let home = std::env::temp_dir().join("moai-flow-cli-home");
    let mut cmd = Command::new(EXE);
    cmd.args(args).env("HOME", &home);
    for (k, v) in envs {
        cmd.env(k, v);
    }
    cmd.env("MOAI_FORGE_GH_BIN", gh);
    let out = cmd
        .stdin(Stdio::null())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .output()
        .unwrap();
    // The CLI wraps every op in the protocol envelope; tests assert on the payload.
    let v: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    (v["result"].clone(), out)
}

/// The audited PR: head `pr-5a4b3c2d`, lane agent-6, body carries the claim line.
fn pr_body(claim: bool) -> String {
    let marker = if claim {
        "\nMoai-Claim: 2200#12345"
    } else {
        ""
    };
    json!({
        "number": 9,
        "head": {"ref": "issue/2200-x", "sha": "pr-5a4b3c2d"},
        "user": {"login": "agent-6"},
        "body": format!("Refs #2200{marker}"),
    })
    .to_string()
}

/// Two lane-authored commits (parents=1) plus one merge commit (parents=2, claim
/// omitted — merge commits are records and exempt, decision 06:02). With `claim`,
/// both lane commits carry the claim line; without it, both are missing theirs.
fn commit_list(claim: bool) -> String {
    let marker = if claim {
        "\n\nMoai-Claim: 2200#12345"
    } else {
        ""
    };
    json!([
        {"sha": "aaa1", "parents": [{"sha": "root"}], "commit": {"message": format!("feat(x): one{marker}")}},
        {"sha": "bbb2", "parents": [{"sha": "aaa1"}], "commit": {"message": format!("fix(x): two{marker}")}},
        {"sha": "mrg3", "parents": [{"sha": "aaa1"}, {"sha": "v2026.11.0"}], "commit": {"message": "Merge branch 'issue/2200-x'"}},
    ])
    .to_string()
}

fn tip_body() -> String {
    json!({"sha": "tip-1a2b3c4d"}).to_string()
}

fn compare_body(ahead: u64, behind: u64, files: &[&str]) -> String {
    json!({
        "status": if behind == 0 && ahead > 0 { "ahead" } else if ahead == 0 && behind == 0 { "identical" } else { "diverged" },
        "ahead_by": ahead,
        "behind_by": behind,
        "files": files.iter().map(|f| json!({"filename": f})).collect::<Vec<_>>(),
    })
    .to_string()
}

fn status_body(state: &str) -> String {
    json!({"state": state}).to_string()
}

/// The `commits/<sha>/check-runs` payload: an aggregate list the CI judgment reads
/// after the #2335 fix (a combined-status `state` is no longer the source).
fn check_runs_body(runs: &[Value]) -> String {
    json!({"total_count": runs.len(), "check_runs": runs}).to_string()
}

fn check_run(name: &str, status: &str, conclusion: &str) -> Value {
    json!({
        "name": name,
        "status": status,
        "conclusion": if conclusion.is_empty() { Value::Null } else { json!(conclusion) },
    })
}

/// rc-ci succeeded; an unrelated deployment check failed and the combined commit
/// status is poisoned to `failure` (the measured shape of PR #2333, 23:52Z).
fn poisoned_combined_ok_checks() -> [(String, String); 2] {
    [
        (
            "FLOW_CHECK_RUNS".into(),
            http200(&check_runs_body(&[
                check_run("rc-ci", "completed", "success"),
                check_run("vercel", "completed", "failure"),
            ])),
        ),
        ("FLOW_STATUS".into(), http200(&status_body("failure"))),
    ]
}

const OWNER_ARGS: [&str; 4] = ["--owner", "acme", "--repo", "widgets"];

fn gate_args(audit_head: &str) -> Vec<String> {
    let mut v: Vec<String> = ["flow-gate-verdict"]
        .iter()
        .chain(OWNER_ARGS.iter())
        .map(|s| s.to_string())
        .collect();
    v.push("--pr".into());
    v.push("9".into());
    v.push("--audit-head".into());
    v.push(audit_head.to_string());
    v
}

fn run_gate(gh: &Path, envs: &[(&str, String)], audit_head: &str) -> (Value, std::process::Output) {
    let owned: Vec<String> = gate_args(audit_head);
    let args: Vec<&str> = owned.iter().map(|s| s.as_str()).collect();
    run_flow(gh, envs, &args)
}

fn reasons_of(v: &Value) -> Vec<String> {
    v["reasons"]
        .as_array()
        .map(|a| {
            a.iter()
                .map(|r| r.as_str().unwrap_or_default().to_string())
                .collect()
        })
        .unwrap_or_default()
}

/// AC-2: the WIP cap at its boundary — lane at cap (1/1) refuses, one below (0/1)
/// allows, a second lane PR over the cap refuses, and the total cap refuses even
/// when the lane itself owns nothing (8/8 with a different author).
#[test]
fn wip_cap_refuses_at_boundary_and_allows_below() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let args = [
        "flow-wip-check",
        OWNER_ARGS[0],
        OWNER_ARGS[1],
        OWNER_ARGS[2],
        OWNER_ARGS[3],
        "--lane",
        "agent-6",
        "--lane-cap",
        "1",
        "--total-cap",
        "8",
    ];

    // At the lane cap: 1/1 → refuse_assignment with both counts reported.
    let open = json!([
        {"number": 1, "head": {"ref": "issue/1-a"}, "user": {"login": "agent-6"}, "draft": false}
    ]);
    let envs = [("FLOW_OPEN_PRS", http200(&open.to_string()))];
    let (v, out) = run_flow(&gh, &envs, &args);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["schema_version"], json!(1));
    assert_eq!(v["lane_open"], json!(1));
    assert_eq!(v["total_open"], json!(1));
    assert_eq!(v["lane_cap"], json!(1));
    assert_eq!(v["total_cap"], json!(8));
    assert_eq!(v["exceeded"], json!(true));
    assert_eq!(v["verdict"], json!("refuse_assignment"));
    assert_eq!(v["writes"], json!(false));

    // Below the cap: 0/1 → allow.
    let envs = [("FLOW_OPEN_PRS", http200("[]"))];
    let (v, out) = run_flow(&gh, &envs, &args);
    assert!(out.status.success());
    assert_eq!(v["lane_open"], json!(0));
    assert_eq!(v["verdict"], json!("allow"));

    // Over the lane cap: 2/1 → refuse.
    let open = json!([
        {"number": 1, "head": {"ref": "issue/1-a"}, "user": {"login": "agent-6"}, "draft": false},
        {"number": 2, "head": {"ref": "issue/2-b"}, "user": {"login": "agent-6"}, "draft": false}
    ]);
    let envs = [("FLOW_OPEN_PRS", http200(&open.to_string()))];
    let (v, _) = run_flow(&gh, &envs, &args);
    assert_eq!(v["verdict"], json!("refuse_assignment"));

    // Total cap reached by other lanes: 8/8 with the lane itself empty → refuse.
    let others: Vec<Value> = (0..8)
        .map(|i| {
            json!({
                "number": 100 + i,
                "head": {"ref": format!("issue/{}", 100 + i)},
                "user": {"login": "agent-other"},
                "draft": false
            })
        })
        .collect();
    let envs = [("FLOW_OPEN_PRS", http200(&json!(others).to_string()))];
    let (v, _) = run_flow(&gh, &envs, &args);
    assert_eq!(v["lane_open"], json!(0));
    assert_eq!(v["total_open"], json!(8));
    assert_eq!(v["verdict"], json!("refuse_assignment"));
}

/// AC-3, ok path: tip contained, head at the audited head, claims on every lane
/// commit, the merge commit exempt, body claim present, CI success, head stable.
#[test]
fn gate_verdict_ok_with_merge_commit_exempt() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, out) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["verdict"], json!("ok"), "{}", v);
    assert_eq!(v["head"], json!("pr-5a4b3c2d"));
    assert_eq!(v["tip"], json!("tip-1a2b3c4d"));
    assert_eq!(v["tip_ci"]["state"], json!("green"), "{v}");
    assert_eq!(v["writes"], json!(false));
    assert!(v["checks"]
        .as_array()
        .unwrap()
        .iter()
        .all(|c| c["pass"] == json!(true)));
    assert!(reasons_of(&v).is_empty());
}

/// AC-3: behind the tip (ahead_by 0, behind_by 1) — tip containment fails.
#[test]
fn gate_verdict_behind_tip_reports_behind() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(0, 1, &["src/x.rs"]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert!(reasons_of(&v).iter().any(|r| r.contains("tip_contained")));
}

/// AC-3: the head moved past the audited head and files differ —
/// changed_since_audit fails (the update-branch need is visible, not posted).
#[test]
fn gate_verdict_changed_since_audit_reports_files() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        // The same compare envelope serves both reads (tip…head, audit…head);
        // the audit comparison is what this fixture exercises.
        (
            "FLOW_COMPARE",
            http200(&compare_body(1, 0, &["src/flow.rs"])),
        ),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "audit-9f8e7d6c");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert_eq!(v["changed_since_audit"], json!(true));
    assert_eq!(v["changed_files"], json!(["src/flow.rs"]));
    assert!(reasons_of(&v)
        .iter()
        .any(|r| r.contains("changed_since_audit")));
}

/// AC-3: a lane-authored commit without its claim line blocks, naming the sha;
/// the merge commit without one does not (06:02 exemption).
#[test]
fn gate_verdict_claim_less_lane_commit_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(false))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    let rs = reasons_of(&v);
    assert!(rs.iter().any(|r| r.contains("commit_claim_lines")), "{}", v);
    assert!(
        rs.iter().any(|r| r.contains("aaa1") && r.contains("bbb2")),
        "{}",
        v
    );
    // The merge commit is not named as missing its claim.
    assert!(!rs.iter().any(|r| r.contains("mrg3")));
}

/// AC-3: a PR body without the claim line blocks.
#[test]
fn gate_verdict_missing_body_claim_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(false))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert!(reasons_of(&v).iter().any(|r| r.contains("pr_body_claim")));
}

/// AC-3: CI failing on the head blocks. The combined status here is `success` while
/// the required `rc-ci` check run concluded `failure` — the judgment must read the
/// check run, so the combined state cannot mask it (the inverse of #2333's poisoning).
#[test]
fn gate_verdict_failing_ci_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[
                check_run("rc-ci", "completed", "failure"),
                check_run("vercel", "completed", "success"),
            ])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert!(reasons_of(&v).iter().any(|r| r.contains("ci_status")));
}

/// #2335 RED-1: an unrelated failure poisons the combined commit status (Vercel), but
/// the required `rc-ci` check run succeeded — the verdict must be `ok`. The judgment
/// reads the check run, never the combined state.
#[test]
fn gate_verdict_ignores_poisoned_combined_status() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let mut envs: Vec<(String, String)> = vec![
        ("FLOW_PR".into(), http200(&pr_body(true))),
        ("FLOW_COMMITS".into(), http200(&commit_list(true))),
        ("FLOW_TIP".into(), http200(&tip_body())),
        ("FLOW_COMPARE".into(), http200(&compare_body(3, 0, &[]))),
    ];
    envs.extend(poisoned_combined_ok_checks());
    let refs: Vec<(&str, String)> = envs.iter().map(|(k, v)| (k.as_str(), v.clone())).collect();
    let (v, _) = run_gate(&gh, &refs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("ok"), "{}", v);
    let ci = v["checks"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["check"] == json!("ci_status"))
        .unwrap()
        .clone();
    assert_eq!(ci["pass"], json!(true));
    assert!(
        // #2563: the default required name is the renamed `ci` check.
        ci["detail"]
            .as_str()
            .unwrap()
            .contains("required check runs: ci: success"),
        "{}",
        ci["detail"]
    );
}

/// #2335 RED-2: the branch diverged from the tip (ahead_by 3, behind_by 1) — the tip
/// is NOT contained even though ahead_by > 0; the behind count must fail the check.
#[test]
fn gate_verdict_diverged_from_tip_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 1, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    let rs = reasons_of(&v);
    assert!(rs.iter().any(|r| r.contains("tip_contained")), "{}", v);
    assert!(
        rs.iter().any(|r| r.contains("behind_by 1")),
        "the reason must name the behind count: {:?}",
        rs
    );
}

/// #2335: a required check run that has not concluded yet (in_progress) blocks — the
/// gate cannot pass on CI that has not finished.
#[test]
fn gate_verdict_pending_ci_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run("rc-ci", "in_progress", "")])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    let rs = reasons_of(&v);
    assert!(rs.iter().any(|r| r.contains("ci_status")), "{}", v);
    assert!(rs.iter().any(|r| r.contains("pending")), "{:?}", rs);
}

/// #2335: a required check run that does not exist on the head blocks by name — an
/// unrelated check (`vercel`) present on the same list is ignored.
#[test]
fn gate_verdict_missing_required_check_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "vercel",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    let rs = reasons_of(&v);
    assert!(
        rs.iter().any(|r| r.contains("ci check run not found")),
        "{:?}",
        rs
    );
}

/// #2335: the required-check list is configurable — two required names pass only when
/// both check runs succeeded, and one failure blocks by name.
#[test]
fn gate_verdict_required_checks_list_is_configurable() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let gate = |runs: &[Value]| {
        let envs = [
            ("FLOW_PR", http200(&pr_body(true))),
            ("FLOW_COMMITS", http200(&commit_list(true))),
            ("FLOW_TIP", http200(&tip_body())),
            ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
            ("FLOW_CHECK_RUNS", http200(&check_runs_body(runs))),
            ("FLOW_STATUS", http200(&status_body("success"))),
        ];
        let mut owned = gate_args("pr-5a4b3c2d");
        owned.push("--required-checks".into());
        owned.push("rc-ci,docs-check".into());
        let args: Vec<&str> = owned.iter().map(|s| s.as_str()).collect();
        run_flow(&gh, &envs, &args)
    };
    // Both required checks succeeded → ok.
    let (v, out) = gate(&[
        check_run("rc-ci", "completed", "success"),
        check_run("docs-check", "completed", "success"),
        check_run("vercel", "completed", "failure"),
    ]);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["verdict"], json!("ok"), "{}", v);
    // One required check failed → blocked, the unrelated one is not named.
    let (v, _) = gate(&[
        check_run("rc-ci", "completed", "success"),
        check_run("docs-check", "completed", "failure"),
        check_run("vercel", "completed", "failure"),
    ]);
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    let rs = reasons_of(&v);
    assert!(rs.iter().any(|r| r.contains("docs-check")), "{:?}", rs);
    assert!(!rs.iter().any(|r| r.contains("vercel")));
}

/// AC-3: the head moved between the two PR reads — head_recheck fails (the audit
/// and the verdict no longer describe the same commit).
#[test]
fn gate_verdict_head_moved_between_reads_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), HEAD_MOVE_STUB);
    let count = tmp.path().join("flow-count");
    let moved = json!({
        "number": 9,
        "head": {"ref": "issue/2200-x", "sha": "pr-pushed-9e8d7c6b"},
        "user": {"login": "agent-6"},
        "body": "Refs #2200\nMoai-Claim: 2200#12345",
    });
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_PR2", http200(&moved.to_string())),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
        ("FLOW_COUNT_FILE", count.to_str().unwrap().to_string()),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert_eq!(v["head"], json!("pr-5a4b3c2d"));
    assert!(
        reasons_of(&v).iter().any(|r| r.contains("head_recheck")),
        "{}",
        v
    );
}

/// Audit F1: an empty policy must block regardless of the returned CI result.
/// Every other check passes, so no unrelated failure can mask a missing CI guard.
fn assert_empty_required_checks_block(required: &str) {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    for conclusion in ["failure", "success"] {
        let envs = [
            ("FLOW_PR", http200(&pr_body(true))),
            ("FLOW_COMMITS", http200(&commit_list(true))),
            ("FLOW_TIP", http200(&tip_body())),
            ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
            (
                "FLOW_CHECK_RUNS",
                http200(&check_runs_body(&[check_run(
                    "rc-ci",
                    "completed",
                    conclusion,
                )])),
            ),
            ("FLOW_STATUS", http200(&status_body(conclusion))),
        ];
        let mut owned = gate_args("pr-5a4b3c2d");
        owned.extend(["--required-checks".to_string(), required.to_string()]);
        let args: Vec<&str> = owned.iter().map(String::as_str).collect();
        let (v, out) = run_flow(&gh, &envs, &args);
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        assert_eq!(
            v["verdict"],
            json!("blocked"),
            "required={required:?}, CI={conclusion}: {v}"
        );
        let checks = v["checks"].as_array().unwrap();
        let ci = checks
            .iter()
            .find(|c| c["check"] == json!("ci_status"))
            .unwrap();
        assert_eq!(ci["pass"], json!(false));
        assert_eq!(
            ci["detail"],
            json!("required check runs: no required checks configured")
        );
        assert!(checks
            .iter()
            .filter(|c| c["check"] != json!("ci_status"))
            .all(|c| c["pass"] == json!(true)));
        assert_eq!(
            reasons_of(&v),
            vec!["ci_status: required check runs: no required checks configured"]
        );
        assert_eq!(v["writes"], json!(false));
    }
}

#[test]
fn gate_verdict_empty_required_checks_blocks() {
    assert_empty_required_checks_block("");
}

#[test]
fn gate_verdict_comma_only_required_checks_blocks() {
    assert_empty_required_checks_block(",");
}

#[test]
fn gate_verdict_whitespace_required_checks_blocks() {
    assert_empty_required_checks_block(" , ");
}

/// Every matching run must succeed, regardless of response ordering.
#[test]
fn gate_verdict_duplicate_required_names_all_must_succeed() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    for (first_status, first_conclusion, second_status, second_conclusion, verdict) in [
        ("completed", "failure", "completed", "success", "blocked"),
        ("completed", "success", "completed", "failure", "blocked"),
        ("in_progress", "", "completed", "success", "blocked"),
        ("completed", "success", "in_progress", "", "blocked"),
        ("completed", "success", "completed", "success", "ok"),
    ] {
        let envs = [
            ("FLOW_PR", http200(&pr_body(true))),
            ("FLOW_COMMITS", http200(&commit_list(true))),
            ("FLOW_TIP", http200(&tip_body())),
            ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
            (
                "FLOW_CHECK_RUNS",
                http200(&check_runs_body(&[
                    check_run("rc-ci", first_status, first_conclusion),
                    check_run("rc-ci", second_status, second_conclusion),
                ])),
            ),
            ("FLOW_STATUS", http200(&status_body("success"))),
        ];
        let (v, out) = run_gate(&gh, &envs, "pr-5a4b3c2d");
        assert!(out.status.success());
        assert_eq!(v["verdict"], json!(verdict), "{v}");
        let ci = v["checks"]
            .as_array()
            .unwrap()
            .iter()
            .find(|c| c["check"] == json!("ci_status"))
            .unwrap();
        assert_eq!(ci["pass"], json!(verdict == "ok"));
        assert_eq!(v["writes"], json!(false));
    }
}

fn run_paginated_check_runs(pages: usize) -> (Value, std::process::Output, usize) {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), PAGINATED_CHECKS_STUB);
    let count = tmp.path().join("page-count");
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        ("FLOW_PAGE_COUNT", count.to_str().unwrap().to_string()),
        ("FLOW_CHECK_PAGES", pages.to_string()),
        (
            "FLOW_LAST_CHECKS",
            check_runs_body(&[check_run("rc-ci", "completed", "success")]),
        ),
    ];
    let (v, out) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    let requested = std::fs::read_to_string(count).unwrap().parse().unwrap();
    (v, out, requested)
}

#[test]
fn gate_verdict_accepts_ten_check_run_pages() {
    let (v, out, requested) = run_paginated_check_runs(10);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["verdict"], json!("ok"), "{v}");
    assert_eq!(requested, 10);
    assert_eq!(v["writes"], json!(false));
}

#[test]
fn gate_verdict_blocks_eleven_check_run_pages() {
    let (v, out, requested) = run_paginated_check_runs(11);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["verdict"], json!("blocked"), "{v}");
    assert_eq!(requested, 10, "never request a page beyond the cap");
    let ci = v["checks"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["check"] == json!("ci_status"))
        .unwrap();
    assert_eq!(ci["pass"], json!(false));
    assert_eq!(
        ci["detail"],
        json!("required check runs: check runs unavailable")
    );
    assert_eq!(v["writes"], json!(false));
}

/// #2563: the aggregate CI check is renamed `rc-ci` → `ci`. Through the transition the
/// two spellings are one required check: a run under either name is found, and a
/// failed or pending run under either name still blocks (the union of both names keeps
/// the every-run-succeeds rule). A matrix job name (`ci / linux-x86_64`) is not the
/// aggregate and never satisfies it. Both `--required-checks` spellings behave the same.
#[test]
fn gate_verdict_reads_both_ci_check_names_without_weakening() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let gate = |required: Option<&str>, runs: &[Value]| {
        let envs = [
            ("FLOW_PR", http200(&pr_body(true))),
            ("FLOW_COMMITS", http200(&commit_list(true))),
            ("FLOW_TIP", http200(&tip_body())),
            ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
            ("FLOW_CHECK_RUNS", http200(&check_runs_body(runs))),
            ("FLOW_STATUS", http200(&status_body("success"))),
        ];
        let mut owned = gate_args("pr-5a4b3c2d");
        if let Some(r) = required {
            owned.push("--required-checks".into());
            owned.push(r.into());
        }
        let args: Vec<&str> = owned.iter().map(|s| s.as_str()).collect();
        run_flow(&gh, &envs, &args).0
    };
    let ok = |name: &str| check_run(name, "completed", "success");
    let failed = |name: &str| check_run(name, "completed", "failure");
    let pending = |name: &str| check_run(name, "in_progress", "");
    for required in [None, Some("rc-ci"), Some("ci")] {
        for (runs, verdict) in [
            (vec![ok("ci")], "ok"),
            (vec![ok("rc-ci")], "ok"),
            (vec![ok("rc-ci"), ok("ci")], "ok"),
            (vec![failed("rc-ci"), ok("ci")], "blocked"),
            (vec![ok("rc-ci"), failed("ci")], "blocked"),
            (vec![ok("rc-ci"), pending("ci")], "blocked"),
            (vec![ok("ci / linux-x86_64")], "blocked"),
            (
                vec![ok("rc-ci / linux-x86_64"), ok("cii"), ok("CI")],
                "blocked",
            ),
        ] {
            let v = gate(required, &runs);
            assert_eq!(v["verdict"], json!(verdict), "{required:?} {runs:?} → {v}");
        }
    }
}

// ---- #2665: the tip's own CI (tip watch) ----

fn tip_run(conclusion: &str, run: u64) -> Value {
    let status = if conclusion.is_empty() {
        "in_progress"
    } else {
        "completed"
    };
    let mut r = check_run("rc-ci", status, conclusion);
    r["details_url"] = json!(format!(
        "https://github.com/acme/widgets/actions/runs/{run}/job/{}",
        run + 1
    ));
    r["output"] = json!({"title": format!("rc-ci {conclusion}")});
    r
}

/// A green PR against a tip whose own CI is `tip_runs`; `repair` adds the tip-repair line.
fn gate_with_tip(tip: Option<String>, repair: bool) -> Value {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let mut pr: Value = serde_json::from_str(&pr_body(true)).unwrap();
    if repair {
        pr["body"] = json!(format!(
            "{}\nmoai-tip-repair: tip-1a2b3c4d run 777",
            pr["body"].as_str().unwrap()
        ));
    }
    let mut envs = vec![
        ("FLOW_PR", http200(&pr.to_string())),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    if let Some(t) = tip {
        envs.push(("FLOW_TIP_CHECK_RUNS", t));
    }
    let (v, out) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    v
}

/// AC-1/AC-2/AC-4: a red tip stops every merge but a tip repair, and names the run.
#[test]
fn gate_verdict_red_tip_blocks_and_names_the_run() {
    let v = gate_with_tip(
        Some(http200(&check_runs_body(&[tip_run("failure", 777)]))),
        false,
    );
    assert_eq!(v["verdict"], json!("blocked"), "{v}");
    assert_eq!(v["tip_ci"]["state"], json!("red"));
    let failed = &v["tip_ci"]["failed"][0];
    assert_eq!(failed["check"], json!("rc-ci"));
    assert_eq!(failed["run_id"], json!(777));
    assert_eq!(failed["title"], json!("rc-ci failure"));
    let reasons = reasons_of(&v);
    assert!(
        reasons
            .iter()
            .any(|r| r.starts_with("tip_ci:") && r.contains("777")),
        "{reasons:?}"
    );
    // Only the tip_ci check fails: the PR itself is fine.
    assert_eq!(reasons.len(), 1, "{reasons:?}");
}

/// AC-1: a PR marked as the tip repair passes a red tip.
#[test]
fn gate_verdict_red_tip_passes_a_tip_repair() {
    let v = gate_with_tip(
        Some(http200(&check_runs_body(&[tip_run("failure", 777)]))),
        true,
    );
    assert_eq!(v["verdict"], json!("ok"), "{v}");
    assert_eq!(v["tip_ci"]["state"], json!("red"));
    assert_eq!(v["tip_ci"]["repair"], json!(true));
}

/// AC-3: a failure re-run green on the same SHA is flaky, not a stop.
#[test]
fn gate_verdict_flaky_tip_does_not_block() {
    let v = gate_with_tip(
        Some(http200(&check_runs_body(&[
            tip_run("failure", 777),
            tip_run("success", 778),
        ]))),
        false,
    );
    assert_eq!(v["verdict"], json!("ok"), "{v}");
    assert_eq!(v["tip_ci"]["state"], json!("flaky"));
    // Named by the required check (the default `ci` covers both spellings, #2563).
    assert_eq!(v["tip_ci"]["flaky"], json!(["ci"]));
}

/// A tip CI still running (or not started) is not a stop; a re-run in progress after a
/// failure is not red either.
#[test]
fn gate_verdict_pending_tip_does_not_block() {
    for runs in [
        vec![tip_run("", 777)],
        vec![],
        vec![tip_run("failure", 777), tip_run("", 778)],
    ] {
        let v = gate_with_tip(Some(http200(&check_runs_body(&runs))), false);
        assert_eq!(v["verdict"], json!("ok"), "{v}");
        assert_eq!(v["tip_ci"]["state"], json!("pending"), "{v}");
    }
}

/// An unreadable tip CI is never a pass: a server error (retried, then a transport
/// failure) and a status the client hands back as is (404) — not an empty, pending list.
#[test]
fn gate_verdict_unreadable_tip_ci_blocks() {
    for response in [
        "HTTP/1.1 500 Internal Server Error\n\n{}",
        "HTTP/1.1 404 Not Found\n\n{\"message\":\"Not Found\"}",
    ] {
        let v = gate_with_tip(Some(response.into()), false);
        assert_eq!(v["verdict"], json!("blocked"), "{response}: {v}");
        assert_eq!(v["tip_ci"]["state"], json!("unknown"), "{response}: {v}");
    }
}

/// The tip is read with `filter=all`, so an earlier failed attempt stays visible.
#[test]
fn gate_verdict_reads_all_tip_attempts() {
    let tmp = TempDir::new().unwrap();
    let log = tmp.path().join("calls");
    let script = format!(
        "#!/bin/sh\necho \"$*\" >> '{}'\n{}",
        log.display(),
        STUB_SCRIPT.trim_start_matches("#!/bin/sh\n")
    );
    let gh = install_stub(tmp.path(), &script);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("ok"), "{v}");
    let calls = std::fs::read_to_string(log).unwrap();
    assert!(
        calls.contains("commits/tip-1a2b3c4d/check-runs?filter=all&per_page=100"),
        "{calls}"
    );
}

/// #2683 audit B1: with two required checks, one red and one still running, red wins —
/// the run that failed is not hidden behind the one still going.
#[test]
fn gate_verdict_red_beats_pending_across_required_checks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let mut lint = tip_run("", 900);
    lint["name"] = json!("lint");
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[
                check_run("rc-ci", "completed", "success"),
                check_run("lint", "completed", "success"),
            ])),
        ),
        (
            "FLOW_TIP_CHECK_RUNS",
            http200(&check_runs_body(&[tip_run("failure", 777), lint])),
        ),
    ];
    let mut args = gate_args("pr-5a4b3c2d");
    args.push("--required-checks".into());
    args.push("ci,lint".into());
    let args: Vec<&str> = args.iter().map(String::as_str).collect();
    let (v, out) = run_flow(&gh, &envs, &args);
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["tip_ci"]["state"], json!("red"), "{v}");
    assert_eq!(v["verdict"], json!("blocked"), "{v}");
}

/// #2683 audit B2: the repair marker counts only at the start of a line; a sentence that
/// mentions it (as a PR body explaining the rule does) is not a repair.
#[test]
fn gate_verdict_repair_marker_only_at_line_start() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let mut pr: Value = serde_json::from_str(&pr_body(true)).unwrap();
    pr["body"] = json!(format!(
        "{}\n본문에 `moai-tip-repair:` 줄이 있는 PR만 통과한다.\n  - see moai-tip-repair: below",
        pr["body"].as_str().unwrap()
    ));
    let envs = [
        ("FLOW_PR", http200(&pr.to_string())),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        (
            "FLOW_CHECK_RUNS",
            http200(&check_runs_body(&[check_run(
                "rc-ci",
                "completed",
                "success",
            )])),
        ),
        (
            "FLOW_TIP_CHECK_RUNS",
            http200(&check_runs_body(&[tip_run("failure", 777)])),
        ),
    ];
    let (v, out) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    assert_eq!(v["tip_ci"]["repair"], json!(false), "{v}");
    assert_eq!(v["verdict"], json!("blocked"), "{v}");
}

/// #2683 audit N1: a tip run cancelled (for example by a concurrency group) with no
/// success is red: the tip has no passing CI, so only its repair merges.
#[test]
fn gate_verdict_cancelled_tip_is_red() {
    let v = gate_with_tip(
        Some(http200(&check_runs_body(&[tip_run("cancelled", 777)]))),
        false,
    );
    assert_eq!(v["tip_ci"]["state"], json!("red"), "{v}");
    assert_eq!(v["verdict"], json!("blocked"), "{v}");
}
