//! #2831 AC-1–3: `lane-next --snapshot` decides from fixed JSON (pull requests, comments,
//! the tip's CI, Issue threads) — each of the seven cases, the order when several hold, the
//! #2821 head-moved re-audit and the #2827 already-answered decision. No network.

mod common;
use serde_json::{json, Value};
use std::process::Command;
use tempfile::TempDir;

const EXE: &str = crate::common::MOAI_BIN;
const LANE: &str = "agent-10";

fn h(c: char) -> String {
    c.to_string().repeat(40)
}

fn next(snap: &Value) -> Value {
    let d = TempDir::new().unwrap();
    let file = d.path().join("snap.json");
    std::fs::write(&file, snap.to_string()).unwrap();
    let out = Command::new(EXE)
        .args(["lane-next", "--lane", LANE, "--repo", "o/r", "--snapshot"])
        .arg(&file)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    assert_eq!(v["operation"], json!("lane.next"));
    assert_eq!(v["result"]["writes"], json!(false));
    v["result"].clone()
}

fn kind(r: &Value) -> Value {
    r["next"]["kind"].clone()
}

fn pr(number: u64, head: &str, author: Option<&str>, comments: Value) -> Value {
    json!({"number": number, "head": head, "author_lane": author, "comments": comments})
}

fn claim(id: u64, lane: &str) -> Value {
    json!({"id": id, "body": format!("moai-claim: session={lane} epoch=1 lease_until=2026-11-14T13:00:00Z\n\nnote")})
}

fn audit_claim(id: u64, lane: &str, head: &str) -> Value {
    json!({"id": id, "body": format!("moai-audit-claim: lane={lane} head={head}")})
}

fn verdict(id: u64, head: &str, v: &str) -> Value {
    json!({"id": id, "body": format!("AUDIT #1 (x) head {head} base {} — verdict: {v}\n\nbody", h('0'))})
}

/// Every case of the list, one snapshot each, in order.
#[test]
fn each_case_is_found() {
    // 1. P0: a red tip.
    let r = next(
        &json!({"tip": {"state": "red", "tip": h('t'), "decided_by": h('t'), "detail": "ci / linux-x86_64 failure run 9"}}),
    );
    assert_eq!(kind(&r), json!("tip_red"), "{r}");
    assert_eq!(r["next"]["priority"], json!("P0"));
    assert_eq!(
        r["next"]["evidence"]["detail"],
        json!("ci / linux-x86_64 failure run 9")
    );
    // A green or unknown tip is no task.
    assert_eq!(
        next(&json!({"tip": {"state": "green"}}))["next"],
        json!(null)
    );

    // 2. My pull request: an audit FAIL on this head, a conflict, a failed CI.
    let mine = |extra: Value| {
        let mut p = pr(10, &h('a'), Some(LANE), json!([]));
        p.as_object_mut()
            .unwrap()
            .extend(extra.as_object().unwrap().clone());
        json!({"prs": [p]})
    };
    let r = next(&mine(json!({"comments": [verdict(5, &h('a'), "FAIL")]})));
    assert_eq!(kind(&r), json!("repair_audit_fail"), "{r}");
    assert_eq!(r["next"]["evidence"]["verdict_comment"], json!(5));
    // A FAIL on an earlier head is already answered by the push.
    let r = next(&mine(json!({"comments": [verdict(5, &h('b'), "FAIL")]})));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        kind(&next(&mine(json!({"conflicting": true})))),
        json!("resolve_conflict")
    );
    let r = next(&mine(json!({"ci": "red", "ci_detail": "run 7"})));
    assert_eq!(kind(&r), json!("repair_ci"));
    assert_eq!(next(&mine(json!({"ci": "pending"})))["next"], json!(null));

    // 3. An audit I reserved and have not judged.
    let r = next(
        &json!({"prs": [pr(11, &h('c'), Some("agent-28"), json!([audit_claim(1, LANE, &h('c'))]))]}),
    );
    assert_eq!(kind(&r), json!("audit"), "{r}");
    assert_eq!(r["next"]["target"]["head"], json!(h('c')));
    // Reserved for an earlier head, still not judged: audit the head now, reserving it first.
    let r = next(
        &json!({"prs": [pr(11, &h('d'), Some("agent-28"), json!([audit_claim(1, LANE, &h('c'))]))]}),
    );
    assert_eq!(kind(&r), json!("audit_new_head"), "{r}");
    // A re-audit asked for after my verdict, on the same head.
    let r = next(&json!({"prs": [pr(11, &h('c'), Some("agent-28"), json!([
        audit_claim(1, LANE, &h('c')), verdict(2, &h('c'), "FAIL"),
        {"id": 3, "body": "재감사 요청(agent-10 께) — 수리 설명"}
    ]))]}));
    assert_eq!(kind(&r), json!("reaudit_requested"), "{r}");
    assert_eq!(r["next"]["evidence"]["request_comment"], json!(3));
    // A PASS on the head now: nothing.
    let r = next(&json!({"prs": [pr(11, &h('c'), Some("agent-28"), json!([
        audit_claim(1, LANE, &h('c')), verdict(2, &h('c'), "PASS")
    ]))]}));
    assert_eq!(r["next"], json!(null), "{r}");

    // 4. Another lane's pull request with no auditor (drafts and my own are skipped).
    let r = next(&json!({"prs": [
        {"number": 12, "head": h('e'), "author_lane": "agent-29", "draft": true, "comments": []},
        pr(13, &h('f'), Some(LANE), json!([])),
        pr(14, &h('g'), Some("agent-29"), json!([])),
    ]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(14));
    // Someone else reserved it: not mine to take.
    let r = next(
        &json!({"prs": [pr(14, &h('g'), Some("agent-29"), json!([audit_claim(1, "agent-27", &h('g'))]))]}),
    );
    assert_eq!(r["next"], json!(null), "{r}");

    // 6. My claimed card with no pull request of mine for it.
    let r = next(
        &json!({"threads": [{"number": 2831, "title": "lane-next", "comments": [claim(100, LANE)]}]}),
    );
    assert_eq!(kind(&r), json!("continue_card"), "{r}");
    assert_eq!(r["next"]["evidence"]["claim_comment"], json!(100));
    // A released claim is not mine any more; an open PR of mine for the card means it waits.
    let r = next(
        &json!({"threads": [{"number": 2831, "comments": [claim(100, LANE),
        {"id": 101, "body": "CLOSE-TABLE\n\nmoai-claim-release: comment=100 by=agent-10"}]}]}),
    );
    assert_eq!(r["next"], json!(null), "{r}");
    let r = next(&json!({
        "threads": [{"number": 2831, "comments": [claim(100, LANE)]}],
        "prs": [{"number": 20, "head": h('a'), "author_lane": LANE, "claim_issue": 2831, "comments": [audit_claim(1, "agent-27", &h('a'))]}]
    }));
    assert_eq!(r["next"], json!(null), "{r}");

    // A pushed branch the open-PR limit held back: open it as a draft now (leader decision
    // #1975 6870150891) — a task, not a wait.
    let r = next(
        &json!({"threads": [{"number": 2831, "comments": [claim(100, LANE),
        {"id": 102, "body": "pr-wait: 2026-11-14T01:58:00Z\n\nbranch pushed"}]}]}),
    );
    assert_eq!(kind(&r), json!("open_draft_pr"), "{r}");
    assert_eq!(r["next"]["evidence"]["pr_wait_comment"], json!(102));
    assert!(r["next"]["action"].as_str().unwrap().contains("--draft"));
    assert_eq!(r["waiting"], json!([]));
    // Once the draft is open, the card waits on its audit: nothing to do.
    let r = next(&json!({
        "threads": [{"number": 2831, "comments": [claim(100, LANE),
            {"id": 102, "body": "pr-wait: 2026-11-14T01:58:00Z"}]}],
        "prs": [{"number": 2856, "head": h('a'), "draft": true, "author_lane": LANE, "claim_issue": 2831,
            "comments": [audit_claim(1, "agent-26", &h('a'))]}]
    }));
    assert_eq!(r["next"], json!(null), "{r}");

    // 7. An open Issue nobody claims; `hold` and claimed ones are skipped.
    let r = next(&json!({"open_issues": [
        {"number": 1, "title": "held", "labels": ["HOLD"], "comments": []},
        {"number": 2, "title": "taken", "comments": [claim(5, "agent-9")]},
        {"number": 3, "title": "free", "labels": ["P3"], "comments": [{"id": 6, "body": "context"}]},
    ]}));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(3));
}

/// AC-2 (the #2821 case): my verdict was on an earlier head and the pull request moved on
/// (a repair or an update-branch) — a delta re-audit, naming the head now in full.
#[test]
fn a_moved_head_after_my_verdict_is_a_reaudit_with_the_head_now() {
    let r = next(&json!({"prs": [pr(2821, &h('2'), Some("agent-26"), json!([
        audit_claim(1, LANE, &h('1')), verdict(2, &h('1'), "PASS"),
        {"id": 3, "body": "update-branch"}
    ]))]}));
    assert_eq!(kind(&r), json!("reaudit_head_moved"), "{r}");
    assert_eq!(r["next"]["target"]["head"], json!(h('2')));
    assert_eq!(r["next"]["evidence"]["verdict_head"], json!(h('1')));
    assert!(r["next"]["action"].as_str().unwrap().contains(&h('2')));
}

/// AC-3 (the #2827 case) and the restart case (#2831 6869108233): a decision request the
/// leader already answered — even while the lane was stopped — is "resume", not "wait"; one
/// still unanswered is listed under `waiting` and holds its card.
#[test]
fn an_answered_decision_is_resumed_and_an_open_one_waits() {
    let blocked = json!({"id": 50, "body": "LANE-BLOCKED kind=decision owner=leader #2827 B1 — (가)/(나)/(다), safe default (나)"});
    let decision = json!({"id": 51, "body": "LEADER-DECISION #2827 수리 방향 — **(나)** chosen\n\n- Decided: ..."});
    let r = next(
        &json!({"threads": [{"number": 2812, "comments": [claim(40, LANE), blocked.clone(), decision]}]}),
    );
    assert_eq!(kind(&r), json!("resume_decision"), "{r}");
    assert_eq!(r["next"]["evidence"]["decision_comment"], json!(51));
    assert_eq!(r["waiting"], json!([]));

    // Unanswered: no task from that card; it is listed as waiting.
    let r = next(&json!({"threads": [{"number": 2812, "comments": [claim(40, LANE), blocked]}]}));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["waiting"][0]["number"], json!(2812));
    assert_eq!(r["waiting"][0]["blocked_comment"], json!(50));

    // In a thread I do not claim, only a line naming me counts.
    let other = json!({"id": 60, "body": "LANE-BLOCKED kind=decision owner=leader #9 — x"});
    let named =
        json!({"id": 61, "body": "LANE-BLOCKED agent-10 kind=decision owner=leader #9 — y"});
    let d = json!({"id": 62, "body": "LEADER-DECISION #9 — (가)"});
    let r = next(&json!({"threads": [{"number": 9, "comments": [other, d.clone()]}]}));
    assert_eq!(r["next"], json!(null), "{r}");
    let r = next(&json!({"threads": [{"number": 9, "comments": [named, d]}]}));
    assert_eq!(kind(&r), json!("resume_decision"), "{r}");
}

/// When several hold at once the list order decides: removing the winner each time reveals
/// the next one, down to nothing.
#[test]
fn the_list_order_decides_when_several_hold() {
    let mut snap = json!({
        "tip": {"state": "red", "tip": h('t')},
        "prs": [
            pr(30, &h('a'), Some(LANE), json!([verdict(1, &h('a'), "FAIL")])),
            pr(31, &h('c'), Some("agent-28"), json!([audit_claim(2, LANE, &h('b')), verdict(3, &h('b'), "FAIL")])),
            pr(32, &h('d'), Some("agent-29"), json!([])),
        ],
        "threads": [
            {"number": 40, "comments": [claim(4, LANE),
                {"id": 5, "body": "LANE-BLOCKED kind=decision owner=leader #40 — x"},
                {"id": 6, "body": "LEADER-DECISION #40 — (가)"}]},
            {"number": 41, "comments": [claim(7, LANE)]},
        ],
        "open_issues": [{"number": 50, "title": "free", "comments": []}],
    });
    let order = [
        "tip_red",
        "repair_audit_fail",
        "reaudit_head_moved",
        "reserve_audit",
        "resume_decision",
        "continue_card",
        "pick_issue",
    ];
    for (i, want) in order.iter().enumerate() {
        let r = next(&snap);
        assert_eq!(kind(&r), json!(want), "step {i}: {r}");
        match *want {
            "tip_red" => snap["tip"]["state"] = json!("green"),
            "repair_audit_fail" => snap["prs"][0]["comments"] = json!([]),
            "reaudit_head_moved" => snap["prs"][1]["comments"]
                .as_array_mut()
                .unwrap()
                .push(verdict(8, &h('c'), "PASS")),
            "reserve_audit" => {
                snap["prs"][2]["comments"] = json!([audit_claim(9, "agent-27", &h('d'))])
            }
            "resume_decision" => snap["threads"][0]["comments"]
                .as_array_mut()
                .unwrap()
                .push(json!({"id": 10, "body": "moai-claim-release: comment=4 by=agent-10"})),
            "continue_card" => snap["threads"][1]["comments"]
                .as_array_mut()
                .unwrap()
                .push(json!({"id": 11, "body": "moai-claim-release: comment=7 by=agent-10"})),
            _ => snap["open_issues"] = json!([]),
        }
    }
    assert_eq!(next(&snap)["next"], json!(null));
}

/// #2873 N2: with a tip-repair pull request open, a red tip is its author's: every other lane
/// is told it is open (`waiting`) and goes on down the list — here to audit it.
#[test]
fn an_open_tip_repair_takes_p0_off_the_other_lanes() {
    let red = json!({"state": "red", "tip": h('t')});
    let repair = |author: &str, comments: Value| json!({"number": 40, "head": h('e'), "author_lane": author, "tip_repair": true, "comments": comments});
    let r = next(&json!({"tip": red, "prs": [repair("agent-28", json!([]))]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(40));
    assert_eq!(r["waiting"][0]["reason"], json!("tip_repair_open"));
    assert_eq!(r["waiting"][0]["number"], json!(40));
    // Its author repairs it as its own pull request.
    let r = next(&json!({"tip": red, "prs": [repair(LANE, json!([verdict(1, &h('e'), "FAIL")]))]}));
    assert_eq!(kind(&r), json!("repair_audit_fail"), "{r}");
    // A pull request without the line does not take P0 off anyone.
    let r = next(&json!({"tip": red, "prs": [pr(41, &h('s'), Some("agent-28"), json!([]))]}));
    assert_eq!(kind(&r), json!("tip_red"), "{r}");
    assert_eq!(r["waiting"], json!([]));
}

/// #2873 N8: a card whose pull request (its body naming this claim) has merged is not a draft
/// to open: `card_pr_merged`. A merge before the `pr-wait` line, or one for another claim,
/// leaves the draft task.
#[test]
fn a_merged_card_is_not_offered_as_a_draft() {
    let card = |merged: Value, wait: bool| {
        let mut comments = vec![claim(100, LANE)];
        if wait {
            comments.push(json!({"id": 102, "body": "pr-wait: 2026-11-14T01:58:00Z", "at": "2026-11-14T01:58:00Z"}));
        }
        json!({"threads": [{"number": 2831, "comments": comments, "merged": merged}]})
    };
    let after = json!([{"number": 2856, "claim": 100, "merged_at": "2026-11-14T03:00:00Z"}]);
    let r = next(&card(after.clone(), true));
    assert_eq!(kind(&r), json!("card_pr_merged"), "{r}");
    assert_eq!(r["next"]["evidence"]["merged_pr"], json!(2856));
    assert!(r["next"]["action"]
        .as_str()
        .unwrap()
        .contains("moai-claim-release: comment=100 by=agent-10"));
    assert_eq!(kind(&next(&card(after, false))), json!("card_pr_merged"));
    // Merged before the branch started waiting: that branch is new work.
    let before = json!([{"number": 2800, "claim": 100, "merged_at": "2026-11-14T01:00:00Z"}]);
    assert_eq!(kind(&next(&card(before, true))), json!("open_draft_pr"));
    // Another claim's merge says nothing about this one.
    let other = json!([{"number": 2801, "claim": 99, "merged_at": "2026-11-14T03:00:00Z"}]);
    assert_eq!(
        kind(&next(&card(other.clone(), true))),
        json!("open_draft_pr")
    );
    assert_eq!(kind(&next(&card(other, false))), json!("continue_card"));
    // No time on the merge: taken as later, so no second pull request is suggested.
    let untimed = json!([{"number": 2856, "claim": 100}]);
    assert_eq!(kind(&next(&card(untimed, true))), json!("card_pr_merged"));
}

/// #2873 N11: a `DECISION-APPLIED <decision id>` line after the decision ends `resume_decision`;
/// a line naming another decision does not.
#[test]
fn an_applied_decision_is_not_resumed_again() {
    let thread = |extra: Value| {
        json!({"threads": [{"number": 2812, "title": "t", "comments": [claim(40, LANE),
            {"id": 50, "body": "LANE-BLOCKED kind=decision owner=leader #2812 — (가)/(나)"},
            {"id": 51, "body": "LEADER-DECISION #2812 — (나)"}, extra]}]})
    };
    let r = next(&thread(
        json!({"id": 52, "body": "DECISION-APPLIED 51 — 6871318339 반영"}),
    ));
    assert_eq!(kind(&r), json!("continue_card"), "{r}");
    assert_eq!(r["waiting"], json!([]));
    let r = next(&thread(json!({"id": 52, "body": "DECISION-APPLIED 50"})));
    assert_eq!(kind(&r), json!("resume_decision"), "{r}");
    let r = next(&thread(json!({"id": 52, "body": "DECISION-APPLIED"})));
    assert_eq!(kind(&r), json!("resume_decision"), "{r}");
}
