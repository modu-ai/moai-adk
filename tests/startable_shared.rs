//! #2901 AC6: `lane-next` and `goal-block` judge "startable" with one function
//! (`moai_cli::startable::judge`), so the same open Issues give the same verdict in both —
//! the same Issues passed over for the same reasons, the same count left, the same end
//! condition. No network: `lane-next --snapshot` and `goal-block` on stdin.

mod common;
use serde_json::{json, Value};
use std::io::Write as _;
use std::process::{Command, Stdio};
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

/// One open Issue of each kind, as the `lane-next` snapshot (`cards`) and `goal-block`
/// (`issues`) both carry it.
fn issues() -> Vec<Value> {
    vec![
        json!({"number": 10, "title": "on hold", "labels": ["HOLD"]}),
        json!({"number": 11, "title": "[rc.2][EP-A] epic"}),
        json!({"number": 12, "title": "waits", "body": "Blocked-By: #13"}),
        json!({"number": 13, "title": "free", "labels": ["P2"]}),
        json!({"number": 14, "title": "parent", "open_children": 1}),
        json!({"number": 15, "title": "waits on GitHub", "blocked_by": [13]}),
        json!({"number": 16, "title": "kept"}),
        json!({"number": 17, "title": "free too", "labels": ["threshold"]}),
        json!({"number": 18, "title": "closed blocker", "body": "Blocked-By: #999"}),
        json!({"number": 19, "title": "decision", "labels": ["needs:decision"]}),
    ]
}

fn keep_set() -> Value {
    json!([{"number": 16, "reason": "운영자 보유"}])
}

fn lane_next(snap: &Value) -> Value {
    let d = TempDir::new().unwrap();
    let file = d.path().join("snap.json");
    std::fs::write(&file, snap.to_string()).unwrap();
    let out = Command::new(EXE)
        .args([
            "lane-next",
            "--lane",
            "agent-10",
            "--repo",
            "o/r",
            "--snapshot",
        ])
        .arg(&file)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn goal_block(fixture: &Value) -> Value {
    let mut child = Command::new(EXE)
        .arg("goal-block")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(fixture.to_string().as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn reasons(entries: &Value) -> Vec<(u64, String)> {
    entries
        .as_array()
        .unwrap_or_else(|| panic!("not a list: {entries}"))
        .iter()
        .map(|e| {
            (
                e["issue"].as_u64().unwrap(),
                e["reason"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn lane_next_and_goal_block_agree_on_startable() {
    let want = vec![
        (10, "held".to_string()),
        (11, "epic".into()),
        (12, "blocked_by".into()),
        (14, "epic".into()),
        (15, "blocked_by".into()),
        (16, "keep_set".into()),
        (19, "held".into()),
    ];
    for prs in [0usize, 1] {
        let snap = json!({
            "cards": issues(),
            "keep_set": keep_set(),
            "prs": (0..prs).map(|_| json!({"number": 900, "head": "a".repeat(40), "author_lane": "agent-28",
                "comments": [{"id": 1, "body": format!("moai-audit-claim: lane=agent-27 head={}", "a".repeat(40))}]}))
                .collect::<Vec<_>>(),
        });
        let mut gb_issues = issues();
        for i in &mut gb_issues {
            i["state"] = json!("open");
        }
        let fixture = json!({
            "milestone": "rc.2",
            "issues": gb_issues,
            "prs": (0..prs).map(|_| json!({"number": 900, "state": "open"})).collect::<Vec<_>>(),
            "keep_set": keep_set(),
        });
        let ln = lane_next(&snap);
        let gb = goal_block(&fixture);
        assert_eq!(reasons(&ln["skipped"]), want, "lane-next {ln}");
        assert_eq!(reasons(&gb["held"]), want, "goal-block {gb}");
        // The entries themselves are the same bytes: one judgment, one entry shape.
        assert_eq!(ln["skipped"], gb["held"]);
        assert_eq!(gb["startable"], json!([13, 17, 18]));
        assert_eq!(ln["end_condition"]["startable"], json!(3));
        assert_eq!(ln["end_condition"], gb["end_condition"], "{prs} open PR(s)");
        assert_eq!(gb["end_condition"]["met"], json!(false));
    }
    // Only held cards and no open pull request: both say the loop is over.
    let only_held = [issues()[0].clone(), issues()[9].clone()];
    let ln = lane_next(&json!({"cards": only_held}));
    let gb = goal_block(
        &json!({"milestone": "rc.2", "issues": only_held.iter().map(|i| {
        let mut i = i.clone();
        i["state"] = json!("open");
        i
    }).collect::<Vec<_>>()}),
    );
    let met = json!({"open_prs": 0, "startable": 0, "met": true});
    assert_eq!(ln["end_condition"], met, "{ln}");
    assert_eq!(gb["end_condition"], met, "{gb}");
    assert_eq!(gb["agent"]["loop_line"], Value::Null, "{gb}");
}
