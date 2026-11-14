//! #2903 (#2899 O4): `close-check --snapshot` answers whether a lane may close its card — the
//! close table must name the Issue, every criterion row must be met, and no open pull request
//! may still claim it. No network.

mod common;
use serde_json::{json, Value};
use std::process::Command;
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;

const TABLE: &str = "CLOSE-TABLE #7 — agent-7, PR #9 병합\n\n| 기준 | 증거 | 판정 |\n| --- | --- | --- |\n| AC-1 | test | 충족 |\n| AC-2 | none | 해당 없음 |\n\n🗿 MoAI";

fn check(snap: &Value) -> (bool, Value) {
    let d = TempDir::new().unwrap();
    let file = d.path().join("snap.json");
    std::fs::write(&file, snap.to_string()).unwrap();
    let out = Command::new(EXE)
        .args(["close-check", "7", "--repo", "o/r", "--snapshot"])
        .arg(&file)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["operation"], json!("close.check"), "{v}");
    (v["ok"] == json!(true), v)
}

fn verdict(snap: &Value) -> Value {
    let (ok, v) = check(snap);
    assert!(ok, "{v}");
    assert_eq!(v["result"]["writes"], json!(false));
    v["result"].clone()
}

fn snap(state: &str, comments: Value, prs: Value) -> Value {
    json!({"number": 7, "state": state, "comments": comments, "prs": prs})
}

fn c(id: u64, body: &str) -> Value {
    json!({"id": id, "body": body})
}

#[test]
fn a_met_table_with_no_open_pull_request_may_close() {
    let r = verdict(&snap(
        "OPEN",
        json!([c(1, "moai-claim: session=agent-7 epoch=1"), c(2, TABLE)]),
        json!([{"number": 9, "state": "MERGED", "claim_issue": 7}]),
    ));
    assert_eq!(r["verdict"], json!("close"), "{r}");
    assert_eq!(r["close_table_comment"], json!(2));
    assert_eq!(r["rows"], json!(2));
    assert_eq!(r["merged_prs"], json!([9]));
    assert!(
        r["next"].as_str().unwrap().contains("gh issue close 7"),
        "{r}"
    );
}

#[test]
fn each_reason_keeps_the_issue_open() {
    // An unmet row, kept as written.
    let unmet = TABLE.replace("| AC-2 | none | 해당 없음 |", "| AC-2 | later | 후속 #12 |");
    let r = verdict(&snap("OPEN", json!([c(2, &unmet)]), json!([])));
    assert_eq!(r["verdict"], json!("keep_open"), "{r}");
    assert_eq!(r["unmet"], json!(["| AC-2 | later | 후속 #12 |"]));
    // An open pull request that claims this Issue; one that only mentions it does not count.
    let r = verdict(&snap(
        "OPEN",
        json!([c(2, TABLE)]),
        json!([{"number": 11, "state": "OPEN", "claim_issue": 7}, {"number": 12, "state": "OPEN", "claim_issue": 8}]),
    ));
    assert_eq!(r["verdict"], json!("keep_open"), "{r}");
    assert_eq!(r["open_prs"], json!([11]));
    let r = verdict(&snap(
        "OPEN",
        json!([c(2, TABLE)]),
        json!([{"number": 12, "state": "OPEN", "claim_issue": 8}, {"number": 13, "state": "OPEN"}]),
    ));
    assert_eq!(r["verdict"], json!("close"), "{r}");
    // A table without criterion rows.
    let r = verdict(&snap(
        "OPEN",
        json!([c(2, "CLOSE-TABLE #7\n\nall done")]),
        json!([]),
    ));
    assert_eq!(r["verdict"], json!("keep_open"), "{r}");
    assert_eq!(r["rows"], json!(0));
}

#[test]
fn the_newest_table_of_this_issue_decides() {
    let unmet = TABLE.replace("| AC-1 | test | 충족 |", "| AC-1 | test | 미충족 |");
    // Older unmet table, newer met one: close.
    let r = verdict(&snap("OPEN", json!([c(2, &unmet), c(3, TABLE)]), json!([])));
    assert_eq!(
        (r["verdict"].clone(), r["close_table_comment"].clone()),
        (json!("close"), json!(3))
    );
    // Newer unmet table wins.
    let r = verdict(&snap("OPEN", json!([c(3, TABLE), c(4, &unmet)]), json!([])));
    assert_eq!(r["verdict"], json!("keep_open"), "{r}");
    // A table of another Issue, a quoted table, a table line not first: none of them count.
    for body in [
        TABLE.replace("#7 —", "#70 —"),
        format!("> {TABLE}"),
        format!("note\n{TABLE}"),
    ] {
        let r = verdict(&snap("OPEN", json!([c(5, &body)]), json!([])));
        assert_eq!(r["verdict"], json!("no_table"), "{body}\n{r}");
    }
}

#[test]
fn closed_unread_and_mismatched_snapshots() {
    let r = verdict(&snap("CLOSED", json!([c(2, TABLE)]), json!([])));
    assert_eq!(r["verdict"], json!("already_closed"));
    let mut s = snap("OPEN", json!([c(2, "hello")]), json!([]));
    s["partial"] = json!(["comments: read the newest 100 of 140"]);
    let r = verdict(&s);
    assert_eq!(r["verdict"], json!("uncertain"), "{r}");
    let mut s = snap("OPEN", json!([]), json!([]));
    s["number"] = json!(8);
    let (ok, v) = check(&s);
    assert!(!ok, "{v}");
}
