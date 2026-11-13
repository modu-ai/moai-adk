//! Flow control fixtures (A13 #2279). AC-2: the WIP cap refuses at the boundary and
//! allows below it. GitHub reads are replayed through a stub `gh` binary that answers from
//! environment-set canned HTTP envelopes (read-only). The integration gate verdict fixtures
//! went with `flow-gate-verdict` (#2674 L3); `gate-check` covers the merge gate, G8 included
//! (`tests/gate_check_cli.rs`, `tests/gate_collect.rs`).

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

/// The stub `gh`: the WIP check reads only the open pull request list, so every request
/// answers the environment-supplied envelope `FLOW_OPEN_PRS`.
const STUB_SCRIPT: &str = r##"#!/bin/sh
printf '%s' "$FLOW_OPEN_PRS"
"##;

fn install_stub(dir: &Path, script: &str) -> PathBuf {
    let path = dir.join("gh-stub");
    common::write_executable(&path, script);
    path
}

fn run_flow(gh: &Path, envs: &[(&str, String)], args: &[&str]) -> (Value, std::process::Output) {
    // A fresh HOME per call: parallel test processes never share one folder (#2666).
    let home_dir = tempfile::TempDir::new().unwrap();
    let home = home_dir.path();
    let mut cmd = Command::new(EXE);
    cmd.args(args).env("HOME", home);
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

const OWNER_ARGS: [&str; 4] = ["--owner", "acme", "--repo", "widgets"];

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
