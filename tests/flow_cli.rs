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
/// `pulls/9/commits` must win over `pulls/9`.
const STUB_SCRIPT: &str = r##"#!/bin/sh
case "$*" in
  *pulls/9/commits*) printf '%s' "$FLOW_COMMITS";;
  *pulls/9*) printf '%s' "$FLOW_PR";;
  *pulls?state=open*) printf '%s' "$FLOW_OPEN_PRS";;
  *compare/*) printf '%s' "$FLOW_COMPARE";;
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
  *pulls/9/commits*) printf '%s' "$FLOW_COMMITS";;
  *pulls/9*) if [ "$n" -ge 2 ]; then printf '%s' "$FLOW_PR2"; else printf '%s' "$FLOW_PR"; fi;;
  *pulls?state=open*) printf '%s' "$FLOW_OPEN_PRS";;
  *compare/*) printf '%s' "$FLOW_COMPARE";;
  *commits/*/status*) printf '%s' "$FLOW_STATUS";;
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
        ("FLOW_STATUS", http200(&status_body("success"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert!(reasons_of(&v).iter().any(|r| r.contains("pr_body_claim")));
}

/// AC-3: CI failing on the head blocks.
#[test]
fn gate_verdict_failing_ci_blocks() {
    let tmp = TempDir::new().unwrap();
    let gh = install_stub(tmp.path(), STUB_SCRIPT);
    let envs = [
        ("FLOW_PR", http200(&pr_body(true))),
        ("FLOW_COMMITS", http200(&commit_list(true))),
        ("FLOW_TIP", http200(&tip_body())),
        ("FLOW_COMPARE", http200(&compare_body(3, 0, &[]))),
        ("FLOW_STATUS", http200(&status_body("failure"))),
    ];
    let (v, _) = run_gate(&gh, &envs, "pr-5a4b3c2d");
    assert_eq!(v["verdict"], json!("blocked"), "{}", v);
    assert!(reasons_of(&v).iter().any(|r| r.contains("ci_status")));
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
