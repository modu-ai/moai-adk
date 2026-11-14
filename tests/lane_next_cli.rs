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
    next_with(snap, &[])
}

fn next_with(snap: &Value, args: &[&str]) -> Value {
    let d = TempDir::new().unwrap();
    let file = d.path().join("snap.json");
    std::fs::write(&file, snap.to_string()).unwrap();
    let out = Command::new(EXE)
        .args(["lane-next", "--lane", LANE, "--repo", "o/r", "--snapshot"])
        .arg(&file)
        .args(args)
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

/// #2664 E1: a pull request from outside the project (author association not OWNER, MEMBER or
/// COLLABORATOR) or one that does not target the integration branch is the leader's to report
/// to the user: no lane reserves or runs its audit; it is listed in `waiting` as `external_pr`.
/// A pull request whose association and base were not read stays a lane pull request.
#[test]
fn an_external_pull_request_is_not_audited_by_a_lane() {
    let ext = |n: u64, assoc: &str, base: &str, comments: Value| {
        json!({"number": n, "head": h('x'), "author_lane": "agent-29", "author_association": assoc,
            "base": base, "comments": comments})
    };
    // An outside contributor's pull request into the integration branch.
    for assoc in [
        "CONTRIBUTOR",
        "FIRST_TIME_CONTRIBUTOR",
        "FIRST_TIMER",
        "NONE",
    ] {
        let r = next(&json!({"tip_branch": "v2026.11.0",
            "prs": [ext(50, assoc, "v2026.11.0", json!([]))]}));
        assert_eq!(r["next"], json!(null), "{assoc}: {r}");
        assert_eq!(
            r["waiting"][0]["reason"],
            json!("external_pr"),
            "{assoc}: {r}"
        );
        assert_eq!(r["waiting"][0]["number"], json!(50));
    }
    // A member's pull request into another branch than the integration branch.
    let r = next(&json!({"tip_branch": "v2026.11.0",
        "prs": [ext(51, "MEMBER", "main", json!([]))]}));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["waiting"][0]["reason"], json!("external_pr"), "{r}");
    // Even with this lane's reservation on it, the audit is not run by the lane.
    let r = next(&json!({"tip_branch": "v2026.11.0",
        "prs": [ext(52, "NONE", "v2026.11.0", json!([audit_claim(1, LANE, &h('x'))]))]}));
    assert_eq!(r["next"], json!(null), "{r}");
    // Members, owners and collaborators into the integration branch stay lane work, and so does
    // a pull request whose association, base or tip branch was not read.
    for assoc in ["OWNER", "MEMBER", "COLLABORATOR"] {
        let r = next(&json!({"tip_branch": "v2026.11.0",
            "prs": [ext(53, assoc, "v2026.11.0", json!([]))]}));
        assert_eq!(kind(&r), json!("reserve_audit"), "{assoc}: {r}");
        assert_eq!(r["waiting"], json!([]), "{assoc}: {r}");
    }
    let r = next(&json!({"prs": [ext(54, "", "", json!([]))]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    let r = next(&json!({"tip_branch": "v2026.11.0", "prs": [ext(58, "MEMBER", "", json!([]))]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "no base read: {r}");
    let r = next(&json!({"prs": [ext(55, "MEMBER", "main", json!([]))]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "no tip branch read: {r}");
    // An external pull request never hides the next lane task behind it.
    let r = next(&json!({"tip_branch": "v2026.11.0", "prs": [
        ext(56, "NONE", "v2026.11.0", json!([])),
        pr(57, &h('y'), Some("agent-29"), json!([])),
    ]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(57));
    assert_eq!(r["waiting"][0]["number"], json!(56));
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

/// #2880 W3 (AC-2): SYNC before new work — a merged card of mine with no close table is a P1
/// task ahead of an audit I reserved; once `CLOSE-TABLE` is posted after the claim it is no
/// task, and another open pull request of mine for the card keeps it in progress.
#[test]
fn a_merged_card_without_its_close_table_comes_before_audits() {
    let snap = |extra: Vec<Value>, prs: Value| {
        let mut comments = vec![claim(100, LANE)];
        comments.extend(extra);
        json!({"threads": [{"number": 2831, "comments": comments,
            "merged": [{"number": 2856, "claim": 100, "merged_at": "2026-11-14T03:00:00Z"}]}],
            "prs": prs})
    };
    let audit = json!([pr(
        11,
        &h('c'),
        Some("agent-28"),
        json!([audit_claim(1, LANE, &h('c'))])
    )]);
    let r = next(&snap(vec![], audit.clone()));
    assert_eq!(kind(&r), json!("card_pr_merged"), "{r}");
    assert_eq!(r["next"]["priority"], json!("P1"));
    assert!(r["next"]["action"]
        .as_str()
        .unwrap()
        .contains("CLOSE-TABLE #2831"));
    // Close table posted: the audit is next, and the card is no task.
    let table = json!({"id": 101, "body": "CLOSE-TABLE #2831\n\n| AC | ... |"});
    let r = next(&snap(vec![table.clone()], audit));
    assert_eq!(kind(&r), json!("audit"), "{r}");
    let r = next(&snap(vec![table], json!([])));
    assert_eq!(r["next"], json!(null), "{r}");
    // A table line that is not the first line does not count.
    let quoted = json!({"id": 102, "body": "note\nCLOSE-TABLE #2831"});
    assert_eq!(
        kind(&next(&snap(vec![quoted], json!([])))),
        json!("card_pr_merged")
    );
    // Another open pull request of mine for the card: still in progress, not closed yet.
    let open = json!([{"number": 2870, "head": h('a'), "author_lane": LANE, "claim_issue": 2831,
        "comments": [audit_claim(1, "agent-26", &h('a'))]}]);
    assert_eq!(next(&snap(vec![], open))["next"], json!(null));
    // #2891 audit B1 (Ma): an earlier stage's close table, posted before this claim (#2825: S6a
    // CLOSE-TABLE then the S6b claim), does not close the new card.
    let staged = json!({"threads": [{"number": 2831, "comments": [
            {"id": 90, "body": "CLOSE-TABLE #2831 (S6a)\n\n| AC | ... |"}, claim(100, LANE)],
        "merged": [{"number": 2856, "claim": 100, "merged_at": "2026-11-14T03:00:00Z"}]}]});
    let r = next(&staged);
    assert_eq!(kind(&r), json!("card_pr_merged"), "{r}");
    assert_eq!(r["next"]["priority"], json!("P1"));
    // #2891 audit B1 (Mf): another lane's open pull request on the same (split) Issue does not
    // hold my card's SYNC (#2802, #2825).
    let theirs = json!([{"number": 2871, "head": h('b'), "author_lane": "agent-29", "claim_issue": 2831,
        "comments": [audit_claim(1, "agent-26", &h('b'))]}]);
    let r = next(&snap(vec![], theirs));
    assert_eq!(kind(&r), json!("card_pr_merged"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2831));
}

/// #2901 AC1 (the #2345 case): a lane whose only live claim is on a card held for a decision
/// (`needs:decision`) is not told to continue it — it gets the next free card, and the held
/// card is named under `skipped`. Without the label the same claim is the lane's next step.
#[test]
fn a_lane_whose_only_claim_is_held_gets_the_next_free_card() {
    let free = json!([{"number": 2400, "title": "free", "labels": ["P2"], "comments": []}]);
    let r = next(&json!({
        "threads": [{"number": 2345, "title": "held", "labels": ["P1", "needs:decision"],
            "comments": [claim(100, LANE)]}],
        "open_issues": free.clone(),
    }));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2400));
    assert_eq!(
        r["skipped"],
        json!([{"issue": 2345, "reason": "held", "labels": ["needs:decision"]}])
    );
    let r = next(&json!({
        "threads": [{"number": 2345, "title": "not held", "labels": ["P1"],
            "comments": [claim(100, LANE)]}],
        "open_issues": free,
    }));
    assert_eq!(kind(&r), json!("continue_card"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2345));
    assert_eq!(r["skipped"], json!([]));
}

/// #2901 AC2: every reason a card is passed over is named in `skipped` — on hold, left out
/// by the goal, an epic (title or open sub-issues), blocked by an open Issue (body line or
/// GitHub), claimed by another lane — and the first card none of them stops is picked.
#[test]
fn each_skip_reason_is_named() {
    let r = next(&json!({
        "keep_set": [{"number": 2, "reason": "운영자 보유"}],
        "open_issues": [
            {"number": 1, "title": "held", "labels": ["status:hold"], "comments": []},
            {"number": 2, "title": "kept", "comments": []},
            {"number": 3, "title": "[rc.2][EP-X] epic", "comments": []},
            {"number": 4, "title": "parent", "open_children": 2, "comments": []},
            {"number": 5, "title": "waits", "body": "Blocked-By: #1", "comments": []},
            {"number": 6, "title": "waits on GitHub", "blocked_by": [1], "comments": []},
            {"number": 7, "title": "taken", "comments": [claim(9, "agent-9")]},
            {"number": 8, "title": "free", "labels": ["threshold"], "body": "Blocked-By: #999", "comments": []},
        ]
    }));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(8));
    assert_eq!(
        r["skipped"],
        json!([
            {"issue": 1, "reason": "held", "labels": ["status:hold"]},
            {"issue": 2, "reason": "keep_set", "keep_reason": "운영자 보유"},
            {"issue": 3, "reason": "epic", "open_children": 0},
            {"issue": 4, "reason": "epic", "open_children": 2},
            {"issue": 5, "reason": "blocked_by", "blocked_by": [1]},
            {"issue": 6, "reason": "blocked_by", "blocked_by": [1]},
            {"issue": 7, "reason": "claimed", "claim_comment": 9},
        ])
    );
    // An older snapshot without the open Issue list cannot tell the end condition.
    assert_eq!(
        r["end_condition"],
        json!({"open_prs": null, "startable": null, "met": null})
    );
}

/// PR #2916 review: when the open Issue list was read only in part, a `Blocked-By:` reference
/// past the last number read may still be open — the card is skipped as `blocked_by_unknown`,
/// not picked; with the whole list the same reference is closed and the card is picked.
#[test]
fn a_blocker_past_a_partial_list_is_not_read_as_closed() {
    let waits = json!({"number": 1, "title": "waits", "body": "Blocked-By: #150", "comments": []});
    let free = json!({"number": 2, "title": "free", "comments": []});
    let partial = next(&json!({
        "cards": [waits.clone(), free.clone()],
        "cards_truncated": true,
        "open_issues": [waits.clone(), free.clone()],
    }));
    assert_eq!(kind(&partial), json!("pick_issue"), "{partial}");
    assert_eq!(partial["next"]["target"]["number"], json!(2));
    assert_eq!(
        partial["skipped"],
        json!([{"issue": 1, "reason": "blocked_by_unknown", "blocked_by": [150]}])
    );
    let whole = next(&json!({
        "cards": [waits.clone(), free.clone()],
        "open_issues": [waits, free],
    }));
    assert_eq!(kind(&whole), json!("pick_issue"), "{whole}");
    assert_eq!(whole["next"]["target"]["number"], json!(1));
}

/// #2901 AC5: `end_condition` counts the open pull requests and the startable open Issues of
/// the list (`cards`); met only when both are zero, null while the list was read in part.
#[test]
fn the_end_condition_counts_open_prs_and_startable_cards() {
    let held = json!({"number": 1, "title": "held", "labels": ["HOLD"]});
    let free = json!({"number": 2, "title": "free"});
    let cases = [
        (
            json!({"cards": [held.clone()]}),
            json!({"open_prs": 0, "startable": 0, "met": true}),
        ),
        (
            json!({"cards": [held.clone(), free]}),
            json!({"open_prs": 0, "startable": 1, "met": false}),
        ),
        (
            json!({"cards": [held.clone()], "prs": [pr(9, &h('a'), Some("agent-28"), json!([audit_claim(1, "agent-27", &h('a'))]))]}),
            json!({"open_prs": 1, "startable": 0, "met": false}),
        ),
        (
            json!({"cards": [held], "cards_truncated": true}),
            json!({"open_prs": 0, "startable": null, "met": null}),
        ),
    ];
    for (snap, want) in cases {
        let r = next(&snap);
        assert_eq!(r["end_condition"], want, "{snap}");
    }
}

/// #2933 P-3: with no pull request work, a startable Issue nobody claims is the result even
/// when the five oldest startable ones are claimed by other lanes.
#[test]
fn startable_issue_is_picked_when_no_pr_work() {
    let issues: Vec<Value> = (1..=6u64)
        .map(|n| {
            let comments = if n <= 5 {
                json!([claim(n * 10, &format!("agent-{}", 20 + n))])
            } else {
                json!([])
            };
            json!({"number": n, "title": format!("issue {n}"), "comments": comments})
        })
        .collect();
    let r = next(&json!({"open_issues": issues}));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(6));
    let skipped = r["skipped"].as_array().unwrap();
    assert_eq!(skipped.len(), 5, "{r}");
    assert!(
        skipped.iter().all(|s| s["reason"] == json!("claimed")),
        "{r}"
    );
}

/// #2933 S1: with nothing to do, the result says why — each pull request passed over with its
/// reason (`pr_skipped`), each Issue skipped, and with `--explain` one line for each and a last
/// line for the end condition. The snapshot's `now` is echoed.
#[test]
fn explain_says_why_nothing_is_next() {
    let snap = json!({
        "now": "2026-11-14T12:00:00Z",
        "prs": [
            pr(5, &h('a'), Some("agent-3"), json!([audit_claim(9, "agent-3", &h('a'))])),
            {"number": 6, "head": h('b'), "author_lane": "agent-4", "draft": true, "comments": []},
        ],
        "cards": [{"number": 7, "title": "held", "labels": ["HOLD"]}],
    });
    let r = next_with(&snap, &["--explain"]);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["pr_skipped"],
        json!([
            {"pr": 5, "reason": "audit_reserved", "by": "agent-3", "comment": 9},
            {"pr": 6, "reason": "draft"},
        ]),
        "{r}"
    );
    assert_eq!(r["now"], json!("2026-11-14T12:00:00Z"));
    assert_eq!(
        r["explain"],
        json!([
            "PR #5 passed over: audit_reserved by agent-3 (comment 9)",
            "PR #6 passed over: draft",
            "Issue #7 skipped: held",
            "end_condition: open_prs=2 startable=0 met=false",
        ]),
        "{r}"
    );
}

/// #2933 E: a lane with two live claims on open cards takes no new Issue — `waiting` names
/// the cap; one live claim does not stop it, and a card closed (`CLOSE-TABLE`) or merged
/// does not count.
#[test]
fn two_live_cards_stop_pick_issue() {
    let free = json!([{"number": 2600, "title": "free", "comments": []}]);
    // Each claimed card has the lane's own open pull request, so step 6 passes over it.
    let snap = |cards: &[(u64, Value)]| {
        let threads: Vec<Value> = cards
            .iter()
            .map(|(n, extra)| {
                let mut t =
                    json!({"number": n, "title": "card", "comments": [claim(n * 10, LANE)]});
                t.as_object_mut()
                    .unwrap()
                    .extend(extra.as_object().unwrap().clone());
                t
            })
            .collect();
        let prs: Vec<Value> = cards
            .iter()
            .map(|(n, _)| {
                json!({"number": n + 100, "head": h('a'), "author_lane": LANE,
                    "claim_issue": n, "comments": []})
            })
            .collect();
        json!({"threads": threads, "prs": prs, "open_issues": free.clone()})
    };
    let cap = |cards: Vec<u64>| json!([{"reason": "card_cap", "cards": cards, "limit": 2}]);
    // N = 1: the free Issue is offered.
    let r = next(&snap(&[(2501, json!({}))]));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2600));
    assert_eq!(r["waiting"], json!([]));
    // N = 2: no new Issue; the cap waits.
    let r = next(&snap(&[(2501, json!({})), (2502, json!({}))]));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["waiting"], cap(vec![2501, 2502]), "{r}");
    // A card with its close table posted after the claim does not count.
    let closed = json!({"comments": [claim(25020, LANE),
        {"id": 25021, "body": "CLOSE-TABLE #2502\n\n| AC | ... |"}]});
    let r = next(&snap(&[(2501, json!({})), (2502, closed)]));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["waiting"], json!([]));
    // A card whose pull request for this claim merged does not count.
    let merged =
        json!({"merged": [{"number": 90, "claim": 25020, "merged_at": "2026-11-14T03:00:00Z"}]});
    let r = next(&snap(&[(2501, json!({})), (2502, merged)]));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["waiting"], json!([]));
    // Without its own pull requests the lane continues its oldest card (step 6 comes first);
    // with two such cards the cap is named as well, with one it is not.
    let bare = |ns: &[u64]| {
        let threads: Vec<Value> = ns
            .iter()
            .map(|n| json!({"number": n, "title": "card", "comments": [claim(n * 10, LANE)]}))
            .collect();
        json!({"threads": threads, "open_issues": free.clone()})
    };
    let r = next(&bare(&[2501]));
    assert_eq!(kind(&r), json!("continue_card"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2501));
    assert_eq!(r["waiting"], json!([]));
    let r = next(&bare(&[2501, 2502]));
    assert_eq!(kind(&r), json!("continue_card"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2501));
    assert_eq!(r["waiting"], cap(vec![2501, 2502]), "{r}");
}

/// #2933 S1: the lane's own pull request with nothing to fix and its own draft are named too,
/// and naming them changes nothing about the task chosen.
#[test]
fn each_passed_over_pr_names_its_reason() {
    let snap = json!({"prs": [
        pr(3, &h('c'), Some(LANE), json!([])),
        {"number": 4, "head": h('d'), "author_lane": LANE, "draft": true, "comments": []},
        pr(8, &h('e'), Some("agent-29"), json!([])),
    ]});
    let r = next(&snap);
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(8));
    assert_eq!(
        r["pr_skipped"],
        json!([{"pr": 3, "reason": "own_no_blocker"}, {"pr": 4, "reason": "draft"}]),
        "{r}"
    );
}

/// #2933 S1: without `--explain` the result has no `explain` key and is otherwise the same
/// bytes; a snapshot without `now` gives no `now` (the clock fills it only when reading GitHub).
#[test]
fn explain_is_only_added_when_asked() {
    let snap = json!({"prs": [pr(5, &h('a'), Some("agent-3"), json!([audit_claim(9, "agent-3", &h('a'))]))]});
    let plain = next(&snap);
    assert!(plain.get("explain").is_none(), "{plain}");
    assert!(plain.get("now").is_none(), "{plain}");
    let mut explained = next_with(&snap, &["--explain"]);
    assert!(explained["explain"].is_array(), "{explained}");
    explained.as_object_mut().unwrap().remove("explain");
    assert_eq!(explained.to_string(), plain.to_string());
}
