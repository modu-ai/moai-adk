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
    // Run in the temp folder, never in the checkout under test: a count written by a run with
    // no `--dir` would land in this repository's own state folder (#3005 P8 N6).
    let out = Command::new(EXE)
        .current_dir(d.path())
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
    // A FAIL on an earlier head is answered by the push and a re-audit request (#2933 S3);
    // the push alone leaves the request to post.
    let r = next(&mine(json!({"comments": [verdict(5, &h('b'), "FAIL")]})));
    assert_eq!(kind(&r), json!("request_reaudit"), "{r}");
    let r = next(&mine(json!({"comments": [verdict(5, &h('b'), "FAIL"),
        {"id": 6, "body": "REPAIR #10 — B1 fixed"}]})));
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
    // Another lane's live claim is not this lane's card (#2954 audit N1): N = 1 own + 1 other.
    let mut s = snap(&[(2501, json!({}))]);
    s["threads"]
        .as_array_mut()
        .unwrap()
        .push(json!({"number": 2502, "title": "card", "comments": [claim(25020, "agent-99")]}));
    let r = next(&s);
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2600));
    assert_eq!(r["waiting"], json!([]));
}

/// #2954 audit B1: a card the lane cannot work on now — on hold (`hold`, `needs:decision`)
/// or waiting for a decision (`LANE-BLOCKED kind=decision` with no leader decision yet) — is
/// not counted toward the card cap; the cap still stops at two startable cards.
#[test]
fn held_and_decision_cards_do_not_count_toward_the_cap() {
    let free = json!([{"number": 2600, "title": "free", "comments": []}]);
    let card = |n: u64, labels: &[&str], comments: Vec<Value>| {
        let mut c = vec![claim(n * 10, LANE)];
        c.extend(comments);
        json!({"number": n, "title": "card", "labels": labels, "comments": c})
    };
    let no_cap = |r: &Value| {
        assert!(
            !r["waiting"]
                .as_array()
                .unwrap()
                .iter()
                .any(|w| w["reason"] == json!("card_cap")),
            "{r}"
        )
    };
    // A held card and a card waiting on `needs:decision`: the free Issue is offered.
    let r = next(&json!({
        "threads": [card(2501, &["hold"], vec![]), card(2502, &["needs:decision"], vec![])],
        "open_issues": free.clone(),
    }));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2600));
    no_cap(&r);
    // Two cards whose decision request has no answer yet: the free Issue is offered.
    let blocked = |id: u64| json!({"id": id, "body": "LANE-BLOCKED kind=decision owner=leader #1 — (가)/(나)"});
    let r = next(&json!({
        "threads": [card(2501, &[], vec![blocked(25011)]), card(2502, &[], vec![blocked(25021)])],
        "open_issues": free.clone(),
    }));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(2600));
    no_cap(&r);
    // One held card and two live startable ones (each with the lane's open pull request): the
    // cap holds at the two live cards.
    let prs: Vec<Value> = [2502u64, 2503]
        .iter()
        .map(|n| {
            json!({"number": n + 100, "head": h('a'), "author_lane": LANE,
                "claim_issue": n, "comments": []})
        })
        .collect();
    let r = next(&json!({
        "threads": [card(2501, &["hold"], vec![]), card(2502, &[], vec![]), card(2503, &[], vec![])],
        "prs": prs,
        "open_issues": free,
    }));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["waiting"],
        json!([{"reason": "card_cap", "cards": [2502, 2503], "limit": 2}]),
        "{r}"
    );
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

/// #2933 S3 ①: my pull request's head moved past an audit FAIL on an older head and no
/// re-audit request follows that FAIL — ask for the re-audit (P1, ahead of reserving another
/// lane's audit). A request after the FAIL (`REPAIR…`, `재감사 요청`, `re-audit request`) ends it;
/// one before the FAIL does not. A FAIL on the head now is still a repair.
#[test]
fn own_pr_with_new_head_after_fail_asks_reaudit() {
    let own = |comments: Value| json!({"prs": [pr(60, &h('b'), Some(LANE), comments)]});
    let r = next(&own(json!([
        verdict(5, &h('a'), "FAIL"),
        {"id": 6, "body": "pushed the fix"}
    ])));
    assert_eq!(kind(&r), json!("request_reaudit"), "{r}");
    assert_eq!(r["next"]["priority"], json!("P1"));
    assert_eq!(r["next"]["target"]["head"], json!(h('b')));
    assert_eq!(
        r["next"]["evidence"],
        json!({"fail_comment": 5, "fail_head": h('a'), "head": h('b')})
    );
    assert_eq!(r["pr_skipped"], json!([]), "{r}");
    for body in [
        "REPAIR #60 — B1 fixed at the new head",
        "재감사 요청(agent-27 께) — 수리 설명",
        "Re-audit request: head moved",
    ] {
        let r = next(&own(
            json!([verdict(5, &h('a'), "FAIL"), {"id": 7, "body": body}]),
        ));
        assert_eq!(r["next"], json!(null), "{body}: {r}");
        assert_eq!(
            r["pr_skipped"],
            json!([{"pr": 60, "reason": "own_no_blocker"}])
        );
    }
    // A request posted before the FAIL answers an earlier round, not this one.
    let r = next(&own(json!([
        {"id": 4, "body": "REPAIR #60 — round 1"},
        verdict(5, &h('a'), "FAIL")
    ])));
    assert_eq!(kind(&r), json!("request_reaudit"), "{r}");
    // A FAIL on the head now: repair, not a request.
    let r = next(&own(json!([verdict(5, &h('b'), "FAIL")])));
    assert_eq!(kind(&r), json!("repair_audit_fail"), "{r}");
    // A PASS on an older head is the auditor's delta re-audit, and a verdict on the head now
    // after the FAIL ends it.
    assert_eq!(
        next(&own(json!([verdict(5, &h('a'), "PASS")])))["next"],
        json!(null)
    );
    assert_eq!(
        next(&own(json!([
            verdict(5, &h('a'), "FAIL"),
            verdict(8, &h('b'), "PASS")
        ])))["next"],
        json!(null)
    );
    // Ahead of reserving another lane's audit.
    let r = next(&json!({"prs": [
        pr(59, &h('e'), Some("agent-29"), json!([])),
        pr(60, &h('b'), Some(LANE), json!([verdict(5, &h('a'), "FAIL")])),
    ]}));
    assert_eq!(kind(&r), json!("request_reaudit"), "{r}");
}

/// #2933 S3 ⑥: a pull request (not a draft) with a PASS on its head now and that head's CI
/// green, still open — read the merge gate's stop reason with `moai gate-loop --once --dry-run`
/// (after reserving an unreserved audit, before a leader decision). A PASS on an older head,
/// a later FAIL on the head, a red, pending or unread CI and a draft are none. My own pull
/// request counts too: its author can act on the reason.
#[test]
fn passed_green_unmerged_pr_points_to_gate_dry_run() {
    let passed = |author: &str, vhead: char, ci: Value, extra: Vec<Value>| {
        let mut comments = vec![
            audit_claim(1, "agent-27", &h(vhead)),
            verdict(2, &h(vhead), "PASS"),
        ];
        comments.extend(extra);
        json!({"number": 70, "head": h('c'), "author_lane": author, "ci": ci, "ci_run": 123,
            "comments": comments})
    };
    let r = next(&json!({"prs": [passed("agent-29", 'c', json!("green"), vec![])]}));
    assert_eq!(kind(&r), json!("unstick_passed_pr"), "{r}");
    assert_eq!(r["next"]["priority"], json!("P1"));
    assert_eq!(
        r["next"]["evidence"],
        json!({"pass_comment": 2, "head": h('c'), "ci_run": 123})
    );
    assert!(r["next"]["action"]
        .as_str()
        .unwrap()
        .contains("moai gate-loop --once --dry-run"));
    assert_eq!(r["pr_skipped"], json!([]), "{r}");
    let none = |p: Value, why: &str| {
        let r = next(&json!({"prs": [p]}));
        assert_eq!(r["next"], json!(null), "{why}: {r}");
    };
    none(
        passed("agent-29", 'b', json!("green"), vec![]),
        "PASS on an older head",
    );
    none(passed("agent-29", 'c', json!("red"), vec![]), "red CI");
    none(
        passed("agent-29", 'c', json!("pending"), vec![]),
        "pending CI",
    );
    none(passed("agent-29", 'c', json!(null), vec![]), "CI not read");
    none(
        passed(
            "agent-29",
            'c',
            json!("green"),
            vec![verdict(3, &h('c'), "FAIL")],
        ),
        "a later FAIL on the head",
    );
    let mut draft = passed("agent-29", 'c', json!("green"), vec![]);
    draft["draft"] = json!(true);
    none(draft, "draft");
    // My own pull request with its PASS and green CI.
    let r = next(&json!({"prs": [passed(LANE, 'c', json!("green"), vec![])]}));
    assert_eq!(kind(&r), json!("unstick_passed_pr"), "{r}");
    // Order: an unreserved audit first, then this, then a leader decision.
    let mut snap = json!({
        "prs": [passed("agent-29", 'c', json!("green"), vec![]), pr(71, &h('d'), Some("agent-28"), json!([]))],
        "threads": [{"number": 40, "comments": [claim(4, LANE),
            {"id": 5, "body": "LANE-BLOCKED kind=decision owner=leader #40 — x"},
            {"id": 6, "body": "LEADER-DECISION #40 — (가)"}]}],
    });
    assert_eq!(kind(&next(&snap)), json!("reserve_audit"));
    snap["prs"][1]["comments"] = json!([audit_claim(9, "agent-27", &h('d'))]);
    assert_eq!(kind(&next(&snap)), json!("unstick_passed_pr"));
    snap["prs"][0]["ci"] = json!("pending");
    assert_eq!(kind(&next(&snap)), json!("resume_decision"));
}

/// #2973 T2 (AC4–AC6): an audit record in another form that names the head now is no reason to
/// ask for a second reservation. The PR goes to `waiting` with the record's comment id; the list
/// goes on to the next PR; a record of an older head, or one with a head too short to tell, changes
/// nothing; and nothing is written.
#[test]
fn a_foreign_audit_record_on_the_head_now_waits_instead_of_reserving() {
    let head = "a0f30f6865e273a42e39c6a2e818d772704496be";
    let record = json!({"id": 6858185395u64, "body": "## 독립 감사 델타 (리드 배정 · lane-14) head a0f30f686\n\nPASS 근거"});
    // Alone: nothing to do, and the PR is listed as waiting with the record.
    let r = next(&json!({"prs": [pr(872, head, Some("agent-29"), json!([record.clone()]))]}));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["waiting"],
        json!([{"number": 872, "reason": "foreign_audit_record",
                "comment": 6858185395u64, "head": head}]),
        "{r}"
    );
    // The list goes on: the next PR with no record is still reserved.
    let r = next(&json!({"prs": [
        pr(872, head, Some("agent-29"), json!([record.clone()])),
        pr(873, &h('b'), Some("agent-29"), json!([])),
    ]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(873));
    assert_eq!(r["waiting"][0]["number"], json!(872));
    // A record of an older head: the head now still needs a reservation.
    let r = next(&json!({"prs": [pr(
        872,
        &format!("b0{}", &head[2..]),
        Some("agent-29"),
        json!([record.clone()]),
    )]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    // A head of 6 digits is too short to tell.
    let short = json!({"id": 7, "body": "독립 감사 head a0f30f"});
    let r = next(&json!({"prs": [pr(872, head, Some("agent-29"), json!([short]))]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    // A reservation by another lane still wins (the record adds nothing): audit_reserved.
    let r = next(&json!({"prs": [pr(
        872,
        head,
        Some("agent-29"),
        json!([record, audit_claim(8, "agent-3", head)]),
    )]}));
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["pr_skipped"][0]["reason"], json!("audit_reserved"), "{r}");
}

// --- #2933 S4 ④: a second auditor for a reservation nobody judges ---------------------------

/// When the reservation of the pull requests below was posted.
const RESERVED_AT: &str = "2026-11-14T10:00:00Z";

fn at(mut c: Value, when: &str) -> Value {
    c["at"] = json!(when);
    c
}

fn reservation() -> Value {
    at(audit_claim(500, "agent-3", &h('a')), RESERVED_AT)
}

fn takeover_request(id: u64, lane: &str, comment: u64, head: &str) -> Value {
    at(
        json!({"id": id, "body": format!("moai-audit-takeover-request: comment={comment} lane={lane} head={head}")}),
        "2026-11-14T10:31:00Z",
    )
}

/// The decision for pull request 5 (author `agent-29`) with `comments`, at `now`.
fn second(now: &str, comments: Value) -> Value {
    next(&json!({"now": now, "prs": [pr(5, &h('a'), Some("agent-29"), comments)]}))
}

fn assert_reserved_not_second(r: &Value, ctx: &str) {
    assert_eq!(r["next"], json!(null), "{ctx}: {r}");
    assert_eq!(
        r["pr_skipped"],
        json!([{"pr": 5, "reason": "audit_reserved", "by": "agent-3", "comment": 500}]),
        "{ctx}: {r}"
    );
}

/// ④ boundary on the injected clock: 29:59 after the reservation is still `audit_reserved`; 30:00
/// and 30:01 ask for a second auditor, with the one line to post and the evidence. The line
/// says the leader's handoff makes it effective — a lane never takes over on its own.
#[test]
fn a_reservation_without_a_verdict_for_thirty_minutes_asks_for_a_second_auditor() {
    let r = second("2026-11-14T10:29:59Z", json!([reservation()]));
    assert_reserved_not_second(&r, "29:59");
    for (now, idle) in [
        ("2026-11-14T10:30:00Z", 1800),
        ("2026-11-14T10:30:01Z", 1801),
    ] {
        let r = second(now, json!([reservation()]));
        assert_eq!(kind(&r), json!("request_second_audit"), "{now}: {r}");
        assert_eq!(r["next"]["priority"], json!("P1"), "{r}");
        assert_eq!(
            r["next"]["target"],
            json!({"type": "pull_request", "number": 5, "head": h('a')}),
            "{r}"
        );
        let action = r["next"]["action"].as_str().unwrap();
        assert!(
            action.contains(&format!(
                "`moai-audit-takeover-request: comment=500 lane={LANE} head={}`",
                h('a')
            )),
            "{action}"
        );
        assert!(action.contains("exactly one line"), "{action}");
        assert!(action.contains("moai-claim-handoff"), "{action}");
        assert!(action.contains("by=leader"), "{action}");
        assert!(action.contains("never take over on your own"), "{action}");
        assert_eq!(
            r["next"]["evidence"],
            json!({"reservation_comment": 500, "reserved_by": "agent-3", "head": h('a'),
                   "idle_secs": idle}),
            "{r}"
        );
        // The pull request is not "passed over" when it is the task; nothing is written.
        assert_eq!(r["pr_skipped"], json!([]), "{r}");
        assert_eq!(r["waiting"], json!([]), "{r}");
    }
}

/// ④ also needs the reserving lane quiet for 30:00: a comment that names it after the
/// reservation restarts the count (29:59 after it still holds, 30:00 asks); another lane's
/// comment, or one that only looks like its name (`agent-33`), does not.
#[test]
fn a_reserving_lane_that_is_still_commenting_is_not_replaced() {
    let note = |id: u64, body: &str, when: &str| at(json!({"id": id, "body": body}), when);
    for body in [
        "agent-3: reading the diff, verdict soon",
        "audit progress\nlane=agent-3 files 3/7",
        "audit progress\nauditor: agent-3.",
    ] {
        let comments = json!([reservation(), note(510, body, "2026-11-14T10:10:00Z")]);
        let r = second("2026-11-14T10:39:59Z", comments.clone());
        assert_reserved_not_second(&r, &format!("{body:?} 29:59"));
        let r = second("2026-11-14T10:40:00Z", comments);
        assert_eq!(
            kind(&r),
            json!("request_second_audit"),
            "{body:?} 30:00: {r}"
        );
        assert_eq!(r["next"]["evidence"]["idle_secs"], json!(1800), "{r}");
    }
    // Other lanes talking on the pull request do not count as the reserving lane's activity.
    let r = second(
        "2026-11-14T10:30:00Z",
        json!([
            reservation(),
            note(
                510,
                "agent-7 asks about the schedule",
                "2026-11-14T10:20:00Z"
            ),
            note(511, "lane=agent-33 files 1/2", "2026-11-14T10:25:00Z"),
            note(512, "agent-3x is another name", "2026-11-14T10:29:00Z"),
        ]),
    );
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    // A comment of the reserving lane with no time is not proof of silence.
    let r = second(
        "2026-11-14T11:30:00Z",
        json!([
            reservation(),
            json!({"id": 510, "body": "agent-3 still reading"})
        ]),
    );
    assert_reserved_not_second(&r, "undated activity");
}

/// Decision B: the lowest request comment id wins. My request waits (`takeover_requested`), a
/// lower id of another lane makes mine lose (`takeover_lost`), a request that names another
/// reservation or head counts for nothing; neither is a task.
#[test]
fn the_lowest_takeover_request_wins_and_the_others_wait() {
    let now = "2026-11-14T10:40:00Z";
    // My request is the only one.
    let r = second(
        now,
        json!([reservation(), takeover_request(600, LANE, 500, &h('a'))]),
    );
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["waiting"],
        json!([{"number": 5, "reason": "takeover_requested", "comment": 500,
                "request_comment": 600, "by": LANE, "head": h('a')}]),
        "{r}"
    );
    assert_eq!(r["pr_skipped"], json!([]), "{r}");
    // Another lane's request has a lower id: mine (and a lane that has not asked) lose.
    for mine in [
        json!([takeover_request(700, LANE, 500, &h('a'))]),
        json!([]),
    ] {
        let mut comments = vec![
            reservation(),
            takeover_request(600, "agent-12", 500, &h('a')),
        ];
        comments.extend(mine.as_array().unwrap().clone());
        let r = second(now, json!(comments));
        assert_eq!(r["next"], json!(null), "{r}");
        assert_eq!(
            r["waiting"],
            json!([{"number": 5, "reason": "takeover_lost", "comment": 500,
                    "request_comment": 600, "by": "agent-12", "head": h('a')}]),
            "{r}"
        );
    }
    // Mine has the lower id: I am the one waiting for the leader.
    let r = second(
        now,
        json!([
            reservation(),
            takeover_request(600, LANE, 500, &h('a')),
            takeover_request(700, "agent-12", 500, &h('a')),
        ]),
    );
    assert_eq!(
        r["waiting"][0]["reason"],
        json!("takeover_requested"),
        "{r}"
    );
    // A request for another reservation or another head is not for this one: still a task.
    let r = second(
        now,
        json!([
            reservation(),
            takeover_request(600, "agent-12", 499, &h('a')),
            takeover_request(601, "agent-12", 500, &h('b')),
        ]),
    );
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
}

/// ④ is never asked when the reserving lane is the one asking, when a verdict is on the head,
/// when the head moved since the reservation (that reservation is stale), or when the clock or
/// the reservation's time is not known.
#[test]
fn no_second_auditor_for_a_verdict_a_moved_head_or_an_unknown_time() {
    let old = "2026-11-14T12:00:00Z";
    // A verdict on the head, either way, ends the question.
    for v in ["PASS", "FAIL"] {
        let r = second(old, json!([reservation(), verdict(510, &h('a'), v)]));
        assert_reserved_not_second(&r, v);
    }
    // A verdict on an older head does not.
    let r = second(old, json!([reservation(), verdict(510, &h('b'), "FAIL")]));
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    // The head moved since the reservation: stale, the existing rule applies.
    let moved = at(audit_claim(500, "agent-3", &h('b')), RESERVED_AT);
    let r = second(old, json!([moved]));
    assert_reserved_not_second(&r, "head moved");
    // Another form of audit record on the head now: someone wrote the audit.
    let record = json!({"id": 510, "body": format!("## 독립 감사 델타 head {}", &h('a')[..9])});
    let r = second(old, json!([reservation(), record]));
    assert_reserved_not_second(&r, "foreign record");
    // No clock, or a reservation without a time: unknown, never a guess.
    let r = next(&json!({"prs": [pr(5, &h('a'), Some("agent-29"), json!([reservation()]))]}));
    assert_reserved_not_second(&r, "no now");
    let r = second(old, json!([audit_claim(500, "agent-3", &h('a'))]));
    assert_reserved_not_second(&r, "no reservation time");
    // The lowest reservation on the head holds it (`audit-claim`'s rule): a later one by another
    // lane is inert and restarts nothing.
    let r = second(
        "2026-11-14T10:30:00Z",
        json!([
            reservation(),
            at(audit_claim(520, "agent-4", &h('a')), "2026-11-14T10:20:00Z")
        ]),
    );
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    assert_eq!(
        r["next"]["evidence"]["reserved_by"],
        json!("agent-3"),
        "{r}"
    );
}

/// The decision of `lane` (not the default one) on the snapshot.
fn next_for(lane: &str, snap: &Value) -> Value {
    let d = TempDir::new().unwrap();
    let file = d.path().join("snap.json");
    std::fs::write(&file, snap.to_string()).unwrap();
    let out = Command::new(EXE)
        .current_dir(d.path())
        .args(["lane-next", "--lane", lane, "--repo", "o/r", "--snapshot"])
        .arg(&file)
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// The leader's handoff of reservation 500 (`audit-claim`'s only way to move a hold).
fn handoff(id: u64, to: &str, when: &str) -> Value {
    at(
        json!({"id": id, "body": format!("moai-claim-handoff: comment=500 to={to} by=leader")}),
        when,
    )
}

/// After the leader's handoff the winner is the auditor: it gets the normal `audit` task, not a
/// wait; the lane that reserved first, and everyone else, get no audit (`audit_reserved` by the
/// new holder). A handoff to a session spelling of the lane (`agent-10-59cb6a`) is the lane's.
#[test]
fn a_handoff_makes_the_winner_the_auditor_and_no_one_else() {
    let now = "2026-11-14T10:40:00Z";
    let snap = |to: &str| {
        json!({"now": now, "prs": [pr(5, &h('a'), Some("agent-29"), json!([
            reservation(),
            takeover_request(600, LANE, 500, &h('a')),
            handoff(700, to, "2026-11-14T10:35:00Z"),
        ]))]})
    };
    for to in [LANE, "agent-10-59cb6a"] {
        let r = next_for(LANE, &snap(to));
        assert_eq!(kind(&r), json!("audit"), "{to}: {r}");
        assert_eq!(r["next"]["target"]["number"], json!(5), "{r}");
        assert_eq!(r["waiting"], json!([]), "{to}: {r}");
        // The lane that reserved first is told nothing to audit.
        let r = next_for("agent-3", &snap(to));
        assert_eq!(r["next"], json!(null), "{to}: {r}");
    }
    // Handed to another lane: that lane audits, I (a third lane) and the first holder do not;
    // the PR is `audit_reserved` by the new holder, and no takeover wait is left for the old
    // reservation.
    let to_other = snap("agent-12");
    assert_eq!(kind(&next_for("agent-12", &to_other)), json!("audit"));
    for third in [LANE, "agent-3"] {
        let r = next_for(third, &to_other);
        assert_eq!(r["next"], json!(null), "{third}: {r}");
        assert_eq!(r["waiting"], json!([]), "{third}: {r}");
        assert_eq!(
            r["pr_skipped"],
            json!([{"pr": 5, "reason": "audit_reserved", "by": "agent-12", "comment": 700}]),
            "{third}: {r}"
        );
    }
    // A handoff line that is not the leader's, or names another reservation, moves nothing.
    let not_leader = at(
        json!({"id": 700, "body": "moai-claim-handoff: comment=500 to=agent-12 by=agent-5"}),
        "2026-11-14T10:35:00Z",
    );
    let r = second(
        now,
        json!([
            reservation(),
            takeover_request(600, LANE, 500, &h('a')),
            not_leader
        ]),
    );
    assert_eq!(
        r["waiting"][0]["reason"],
        json!("takeover_requested"),
        "{r}"
    );
}

/// The 30:00 and the holder's silence run against the effective holder and restart at the
/// handoff comment: 29:59 after it is `audit_reserved`, 30:00 asks again. A request posted
/// before the handoff belongs to the old round and is not read; the holder's own comments after
/// the handoff restart the count.
#[test]
fn the_second_audit_clock_restarts_at_the_handoff() {
    let handed = handoff(700, "agent-12", "2026-11-14T10:35:00Z");
    let old_request = takeover_request(600, LANE, 500, &h('a'));
    let comments = json!([reservation(), old_request, handed]);
    let r = second("2026-11-14T11:04:59Z", comments.clone());
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["waiting"], json!([]), "{r}");
    assert_eq!(r["pr_skipped"][0]["by"], json!("agent-12"), "{r}");
    // Just before the handoff's own 30:00 it would already be past the reservation's.
    let r = second("2026-11-14T11:05:00Z", comments);
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    assert_eq!(
        r["next"]["evidence"]["reserved_by"],
        json!("agent-12"),
        "{r}"
    );
    assert_eq!(
        r["next"]["evidence"]["reservation_comment"],
        json!(500),
        "{r}"
    );
    assert_eq!(r["next"]["evidence"]["idle_secs"], json!(1800), "{r}");
    // The handed-off auditor commenting after the handoff restarts the count.
    let note = at(
        json!({"id": 710, "body": "agent-12: reading the diff"}),
        "2026-11-14T10:50:00Z",
    );
    let comments = json!([
        reservation(),
        handoff(700, "agent-12", "2026-11-14T10:35:00Z"),
        note
    ]);
    let r = second("2026-11-14T11:19:59Z", comments.clone());
    assert_eq!(r["next"], json!(null), "{r}");
    let r = second("2026-11-14T11:20:00Z", comments);
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    // A request after the handoff is for the new round: the lowest such id waits for the leader.
    let r = second(
        "2026-11-14T11:20:00Z",
        json!([
            reservation(),
            handoff(700, "agent-12", "2026-11-14T10:35:00Z"),
            at(
                json!({"id": 710, "body": "agent-12: reading the diff"}),
                "2026-11-14T10:50:00Z"
            ),
            takeover_request(720, LANE, 500, &h('a')),
        ]),
    );
    assert_eq!(
        r["waiting"][0]["reason"],
        json!("takeover_requested"),
        "{r}"
    );
    assert_eq!(r["waiting"][0]["request_comment"], json!(720), "{r}");
}

/// A comment that names the auditor by a session spelling (`<letters>-<digits>-<fragment>`,
/// the rule `audit-claim` compares lanes by) is the auditor's activity; `agent-33` and
/// `agent-3x` are other lanes.
#[test]
fn an_alias_spelling_of_the_auditor_keeps_it_active() {
    let note = |body: &str| at(json!({"id": 510, "body": body}), "2026-11-14T10:10:00Z");
    for (reserved_lane, body) in [
        ("agent-3", "session=agent-3-59cb6a reading the diff"),
        ("agent-3", "agent-3-hBp17: reading the diff"),
        ("agent-3", "audit progress\nlane=agent-3-2293 files 3/7"),
        ("agent-3-59cb6a", "agent-3: reading the diff"),
    ] {
        let comments = json!([
            at(audit_claim(500, reserved_lane, &h('a')), RESERVED_AT),
            note(body),
        ]);
        let r = second("2026-11-14T10:39:59Z", comments.clone());
        assert_eq!(
            r["next"],
            json!(null),
            "{reserved_lane} {body:?} 29:59: {r}"
        );
        assert_eq!(r["pr_skipped"][0]["reason"], json!("audit_reserved"), "{r}");
        let r = second("2026-11-14T10:40:00Z", comments);
        assert_eq!(
            kind(&r),
            json!("request_second_audit"),
            "{body:?} 30:00: {r}"
        );
    }
    // Other lanes' numbers or fragments do not count, aliased or not.
    for body in [
        "agent-33 reading",
        "session=agent-33-59cb6a",
        "agent-3x reading",
        "agent-30-2293 hi",
    ] {
        let comments = json!([reservation(), note(body)]);
        let r = second("2026-11-14T10:30:00Z", comments);
        assert_eq!(kind(&r), json!("request_second_audit"), "{body:?}: {r}");
    }
}

/// The place of ④ in the list: after my own pull request's work and after reserving a pull
/// request that has no auditor (③), before an open Issue; a PR whose request I am waiting on
/// does not hold the list back.
#[test]
fn the_second_audit_comes_after_reserving_and_before_issues() {
    let now = "2026-11-14T10:40:00Z";
    let reserved = pr(5, &h('a'), Some("agent-29"), json!([reservation()]));
    let free = pr(8, &h('c'), Some("agent-29"), json!([]));
    let issue = json!({"number": 1, "title": "free", "comments": []});
    // ③ first.
    let r = next(&json!({"now": now, "prs": [reserved.clone(), free.clone()]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(8));
    // My own pull request's FAIL first.
    let mine = pr(3, &h('d'), Some(LANE), json!([verdict(9, &h('d'), "FAIL")]));
    let r = next(&json!({"now": now, "prs": [reserved.clone(), mine]}));
    assert_eq!(kind(&r), json!("repair_audit_fail"), "{r}");
    // Before an Issue.
    let r = next(&json!({"now": now, "prs": [reserved.clone()], "open_issues": [issue.clone()]}));
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    // Waiting on my request, the list goes on: the free PR is reserved, the Issue is picked.
    let asked = pr(
        5,
        &h('a'),
        Some("agent-29"),
        json!([reservation(), takeover_request(600, LANE, 500, &h('a'))]),
    );
    let r = next(&json!({"now": now, "prs": [asked.clone(), free]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(
        r["waiting"][0]["reason"],
        json!("takeover_requested"),
        "{r}"
    );
    let r = next(&json!({"now": now, "prs": [asked], "open_issues": [issue]}));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(
        r["waiting"][0]["reason"],
        json!("takeover_requested"),
        "{r}"
    );
}

// --- #2933 S5 ⑤: repair help for a FAIL nobody repairs --------------------------------------

/// When the FAIL verdicts of the pull requests below were posted.
const FAILED_AT: &str = "2026-11-14T10:00:00Z";

/// The auditor `agent-3` reserved pull request 5 and judged it FAIL on the head `a`.
fn failed() -> Vec<Value> {
    vec![
        at(audit_claim(500, "agent-3", &h('a')), "2026-11-14T09:00:00Z"),
        at(verdict(510, &h('a'), "FAIL"), FAILED_AT),
    ]
}

fn offer(id: u64, lane: &str, number: u64, fail: u64, head: &str) -> Value {
    at(
        json!({"id": id, "body": format!("moai-repair-help-offer: pr={number} lane={lane} fail={fail} head={head}")}),
        "2026-11-14T10:31:00Z",
    )
}

fn note(id: u64, body: &str, when: &str) -> Value {
    at(json!({"id": id, "body": body}), when)
}

/// The decision for pull request 5 (author `agent-29`) with `comments`, at `now`.
fn help(now: &str, comments: Vec<Value>) -> Value {
    next(&json!({"now": now, "prs": [pr(5, &h('a'), Some("agent-29"), json!(comments))]}))
}

fn assert_failed_not_helped(r: &Value, ctx: &str) {
    assert_eq!(r["next"], json!(null), "{ctx}: {r}");
    assert_eq!(
        r["pr_skipped"],
        json!([{"pr": 5, "reason": "audit_reserved", "by": "agent-3", "comment": 500}]),
        "{ctx}: {r}"
    );
}

/// ⑤ boundary on the injected clock: 29:59 after the FAIL is still `audit_reserved`; 30:00 and
/// 30:01 offer repair help, with the one line to post. The line is only an offer: no push before
/// the author's consent or the leader's handoff.
#[test]
fn a_fail_nobody_repairs_for_thirty_minutes_gets_a_help_offer() {
    assert_failed_not_helped(&help("2026-11-14T10:29:59Z", failed()), "29:59");
    for (now, idle) in [
        ("2026-11-14T10:30:00Z", 1800),
        ("2026-11-14T10:30:01Z", 1801),
    ] {
        let r = help(now, failed());
        assert_eq!(kind(&r), json!("offer_repair_help"), "{now}: {r}");
        assert_eq!(r["next"]["priority"], json!("P1"), "{r}");
        assert_eq!(
            r["next"]["target"],
            json!({"type": "pull_request", "number": 5, "head": h('a')}),
            "{r}"
        );
        let action = r["next"]["action"].as_str().unwrap();
        assert!(
            action.contains(&format!(
                "`moai-repair-help-offer: pr=5 lane={LANE} fail=510 head={}`",
                h('a')
            )),
            "{action}"
        );
        assert!(action.contains("exactly one line"), "{action}");
        assert!(action.contains("only an offer"), "{action}");
        assert!(action.contains("never push"), "{action}");
        assert!(action.contains("agent-29"), "{action}");
        assert!(action.contains("moai-claim-handoff"), "{action}");
        assert!(action.contains("by=leader"), "{action}");
        assert_eq!(
            r["next"]["evidence"],
            json!({"fail_comment": 510, "failed_by": "agent-3", "author_lane": "agent-29",
                   "head": h('a'), "idle_secs": idle}),
            "{r}"
        );
        // The pull request is not "passed over" when it is the task; nothing is written.
        assert_eq!(r["pr_skipped"], json!([]), "{r}");
        assert_eq!(r["waiting"], json!([]), "{r}");
    }
}

/// ⑤ also needs the author lane silent on the pull request for 30:00 after the FAIL: a comment
/// that names it restarts the count (29:59 after it still holds, 30:00 offers); other lanes'
/// comments, or one that only looks like its name, do not; a comment of the author with no time
/// is not proof of silence.
#[test]
fn an_author_who_is_still_commenting_is_not_offered_help() {
    for body in [
        "agent-29: reading the findings",
        "REPAIR\nlane=agent-29 fixing 2/5",
        "by=agent-29.",
    ] {
        let mut comments = failed();
        comments.push(note(520, body, "2026-11-14T10:10:00Z"));
        assert_failed_not_helped(&help("2026-11-14T10:39:59Z", comments.clone()), body);
        let r = help("2026-11-14T10:40:00Z", comments);
        assert_eq!(kind(&r), json!("offer_repair_help"), "{body:?} 30:00: {r}");
        assert_eq!(r["next"]["evidence"]["idle_secs"], json!(1800), "{r}");
    }
    let mut others = failed();
    others.extend([
        note(
            520,
            "agent-7 asks about the schedule",
            "2026-11-14T10:20:00Z",
        ),
        note(521, "lane=agent-299 files 1/2", "2026-11-14T10:25:00Z"),
        note(522, "agent-29x is another name", "2026-11-14T10:29:00Z"),
    ]);
    let r = help("2026-11-14T10:30:00Z", others);
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
    let mut undated = failed();
    undated.push(json!({"id": 520, "body": "agent-29 is fixing it"}));
    assert_failed_not_helped(&help("2026-11-14T11:30:00Z", undated), "undated activity");
}

/// ⑤ is for a FAIL on the head now that is still the newest verdict: a verdict on an older head,
/// a newer verdict (any head), a PASS, or no verdict leaves it to the existing rules.
#[test]
fn no_help_offer_for_a_moved_head_a_newer_verdict_or_an_unknown_time() {
    let old = "2026-11-14T13:00:00Z";
    // The FAIL was on an older head: the author pushed since (the head moved).
    let moved = vec![
        at(audit_claim(500, "agent-3", &h('b')), "2026-11-14T09:00:00Z"),
        at(verdict(510, &h('b'), "FAIL"), FAILED_AT),
    ];
    let r = help(old, moved);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["pr_skipped"][0]["reason"], json!("audit_reserved"), "{r}");
    // A newer verdict on the head: the PASS ends the repair; a newer one on another head too.
    for newer in [verdict(520, &h('a'), "PASS"), verdict(520, &h('b'), "FAIL")] {
        let mut comments = failed();
        comments.push(at(newer, "2026-11-14T10:10:00Z"));
        let r = help(old, comments);
        assert_ne!(kind(&r), json!("offer_repair_help"), "{r}");
        assert_eq!(r["next"], json!(null), "{r}");
    }
    // A reservation with no verdict yet is ④'s, not ⑤'s.
    let r = help(old, vec![failed().remove(0)]);
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    // An unknown clock, or a FAIL with no time: not proven, nothing is offered.
    let r = next(&json!({"prs": [pr(5, &h('a'), Some("agent-29"), json!(failed()))]}));
    assert_failed_not_helped(&r, "no now");
    let untimed = vec![
        failed().remove(0),
        verdict(510, &h('a'), "FAIL"), // no `at`
    ];
    assert_failed_not_helped(&help(old, untimed), "no FAIL time");
    // The author lane is not known: its silence cannot be read.
    let r = next(&json!({"now": old, "prs": [pr(5, &h('a'), None, json!(failed()))]}));
    assert_failed_not_helped(&r, "no author lane");
}

/// Not for the author lane (its FAIL is its own repair), and not for the lane that posted the
/// FAIL (it keeps the re-audit).
#[test]
fn neither_the_author_nor_the_fails_own_auditor_is_offered_help() {
    let old = "2026-11-14T13:00:00Z";
    let mine = next(&json!({"now": old, "prs": [pr(5, &h('a'), Some(LANE), json!(failed()))]}));
    assert_eq!(kind(&mine), json!("repair_audit_fail"), "{mine}");
    // I reserved and judged it: the re-audit is mine, no offer.
    let comments = vec![
        at(audit_claim(500, LANE, &h('a')), "2026-11-14T09:00:00Z"),
        at(verdict(510, &h('a'), "FAIL"), FAILED_AT),
    ];
    let r = help(old, comments);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["pr_skipped"],
        json!([{"pr": 5, "reason": "audit_reserved", "by": LANE, "comment": 500}]),
        "{r}"
    );
    assert_eq!(r["waiting"], json!([]), "{r}");
}

/// One offer per FAIL and head: my offer waits (`help_offered`); a lower comment id of another
/// lane makes mine lose, and a lane that has not offered, too (`help_offer_lost`); an offer for
/// another pull request, FAIL or head, or by the author lane, counts for nothing.
#[test]
fn the_lowest_help_offer_wins_and_the_others_wait() {
    let now = "2026-11-14T10:40:00Z";
    let with = |extra: Vec<Value>| {
        let mut comments = failed();
        comments.extend(extra);
        help(now, comments)
    };
    let r = with(vec![offer(600, LANE, 5, 510, &h('a'))]);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["waiting"],
        json!([{"number": 5, "reason": "help_offered", "comment": 510,
                "offer_comment": 600, "by": LANE, "head": h('a')}]),
        "{r}"
    );
    assert_eq!(r["pr_skipped"], json!([]), "{r}");
    for mine in [vec![offer(700, LANE, 5, 510, &h('a'))], vec![]] {
        let mut extra = vec![offer(600, "agent-12", 5, 510, &h('a'))];
        extra.extend(mine);
        let r = with(extra);
        assert_eq!(r["next"], json!(null), "{r}");
        assert_eq!(
            r["waiting"],
            json!([{"number": 5, "reason": "help_offer_lost", "comment": 510,
                    "offer_comment": 600, "by": "agent-12", "head": h('a')}]),
            "{r}"
        );
    }
    // Mine has the lower id: I am the one whose offer stands.
    let r = with(vec![
        offer(600, LANE, 5, 510, &h('a')),
        offer(700, "agent-12", 5, 510, &h('a')),
    ]);
    assert_eq!(r["waiting"][0]["reason"], json!("help_offered"), "{r}");
    // Offers for another PR, FAIL or head, or one by the author lane itself (long enough ago
    // that it is not the author's recent activity), are not for this one.
    let r = with(vec![
        offer(600, "agent-12", 6, 510, &h('a')),
        offer(601, "agent-12", 5, 509, &h('a')),
        offer(602, "agent-12", 5, 510, &h('b')),
        at(
            offer(603, "agent-29", 5, 510, &h('a')),
            "2026-11-14T10:05:00Z",
        ),
    ]);
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
}

/// The place of ⑤ in the list: after my own pull request's work, after reserving a pull
/// request that has no auditor (③), after ④ and after `unstick_passed_pr`, before an open
/// Issue; a PR whose offer I wait on does not hold the list back.
#[test]
fn the_help_offer_comes_after_the_audit_steps_and_before_issues() {
    let now = "2026-11-14T10:40:00Z";
    let failed_pr = pr(5, &h('a'), Some("agent-29"), json!(failed()));
    let issue = json!({"number": 1, "title": "free", "comments": []});
    // Before an Issue.
    let r = next(&json!({"now": now, "prs": [failed_pr.clone()], "open_issues": [issue.clone()]}));
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
    // ③ first.
    let free = pr(8, &h('c'), Some("agent-29"), json!([]));
    let r = next(&json!({"now": now, "prs": [failed_pr.clone(), free]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["next"]["target"]["number"], json!(8));
    // My own pull request's FAIL first.
    let mine = pr(3, &h('d'), Some(LANE), json!([verdict(9, &h('d'), "FAIL")]));
    let r = next(&json!({"now": now, "prs": [failed_pr.clone(), mine]}));
    assert_eq!(kind(&r), json!("repair_audit_fail"), "{r}");
    // ④ before ⑤, even when the PR numbers say otherwise.
    let stale = pr(
        9,
        &h('e'),
        Some("agent-29"),
        json!([at(
            audit_claim(30, "agent-4", &h('e')),
            "2026-11-14T09:00:00Z"
        )]),
    );
    let r = next(&json!({"now": now, "prs": [failed_pr.clone(), stale]}));
    assert_eq!(kind(&r), json!("request_second_audit"), "{r}");
    // 4b `unstick_passed_pr` before ⑤.
    let passed = json!({"number": 70, "head": h('f'), "author_lane": "agent-29", "ci": "green",
        "ci_run": 123, "comments": [audit_claim(1, "agent-27", &h('f')), verdict(2, &h('f'), "PASS")]});
    let r = next(&json!({"now": now, "prs": [failed_pr, passed]}));
    assert_eq!(kind(&r), json!("unstick_passed_pr"), "{r}");
    // Waiting on my offer, the list goes on to the Issue.
    let mut comments = failed();
    comments.push(offer(600, LANE, 5, 510, &h('a')));
    let asked = pr(5, &h('a'), Some("agent-29"), json!(comments));
    let r = next(&json!({"now": now, "prs": [asked], "open_issues": [issue]}));
    assert_eq!(kind(&r), json!("pick_issue"), "{r}");
    assert_eq!(r["waiting"][0]["reason"], json!("help_offered"), "{r}");
}

/// An offer line is neither a verdict, a reservation nor an audit record in another form: on a
/// pull request with no auditor it leaves `reserve_audit` alone, even when the offering lane's
/// name says `audit`.
#[test]
fn a_help_offer_line_is_not_read_as_an_audit_record() {
    let r = next(&json!({"now": "2026-11-14T10:40:00Z", "prs": [pr(
        5,
        &h('a'),
        Some("agent-29"),
        json!([offer(600, "audit-2", 5, 510, &h('a'))]),
    )]}));
    assert_eq!(kind(&r), json!("reserve_audit"), "{r}");
    assert_eq!(r["waiting"], json!([]), "{r}");
}

/// Lane names are compared as `audit-claim` compares them (`same_lane`): the session spellings
/// `agent-29-59cb6a`, `session=agent-29-59cb6a` and `agent-29-hBp17:` are the lane `agent-29`;
/// `agent-299-…` is another lane. An author who keeps commenting under such a spelling is active.
#[test]
fn an_author_commenting_under_a_session_spelling_is_still_active() {
    for body in [
        "session=agent-29-59cb6a reading the findings",
        "agent-29-hBp17: fixing 2/5",
        "REPAIR\nlane=agent-29-2293 pushing soon",
    ] {
        let mut comments = failed();
        comments.push(note(520, body, "2026-11-14T10:10:00Z"));
        assert_failed_not_helped(&help("2026-11-14T10:39:59Z", comments.clone()), body);
        let r = help("2026-11-14T10:40:00Z", comments);
        assert_eq!(kind(&r), json!("offer_repair_help"), "{body:?} 30:00: {r}");
        assert_eq!(r["next"]["evidence"]["idle_secs"], json!(1800), "{r}");
    }
    // Another lane whose name only begins like the author's is not the author.
    let mut comments = failed();
    comments.push(note(
        520,
        "session=agent-299-59cb6a still here",
        "2026-11-14T10:20:00Z",
    ));
    let r = help("2026-11-14T10:30:00Z", comments);
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
}

/// The lane that reserved the failed audit keeps the re-audit also when the reservation spells
/// its session (`agent-10-59cb6a` is this lane, `agent-10`): no offer, the PR is passed over.
#[test]
fn the_fails_auditor_under_a_session_spelling_is_not_offered_help() {
    let comments = vec![
        at(
            audit_claim(500, "agent-10-59cb6a", &h('a')),
            "2026-11-14T09:00:00Z",
        ),
        at(verdict(510, &h('a'), "FAIL"), FAILED_AT),
    ];
    let r = help("2026-11-14T13:00:00Z", comments);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["pr_skipped"],
        json!([{"pr": 5, "reason": "audit_reserved", "by": "agent-10-59cb6a", "comment": 500}]),
        "{r}"
    );
}

/// An offer line's `lane=` is compared as a lane: `agent-10-hBp17` is this lane (its offer
/// stands: `help_offered`), `agent-12-2293` is `agent-12` (a lower id makes mine lose), and an
/// offer written under the author's own session spelling is the author's, which counts for nothing.
#[test]
fn offers_under_session_spellings_count_as_their_lane() {
    let now = "2026-11-14T10:40:00Z";
    let with = |extra: Vec<Value>| {
        let mut comments = failed();
        comments.extend(extra);
        help(now, comments)
    };
    let r = with(vec![offer(600, "agent-10-hBp17", 5, 510, &h('a'))]);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(
        r["waiting"],
        json!([{"number": 5, "reason": "help_offered", "comment": 510,
                "offer_comment": 600, "by": "agent-10-hBp17", "head": h('a')}]),
        "{r}"
    );
    let r = with(vec![
        offer(600, "agent-12-2293", 5, 510, &h('a')),
        offer(700, LANE, 5, 510, &h('a')),
    ]);
    assert_eq!(r["waiting"][0]["reason"], json!("help_offer_lost"), "{r}");
    assert_eq!(r["waiting"][0]["by"], json!("agent-12-2293"), "{r}");
    let r = with(vec![at(
        offer(603, "agent-29-59cb6a", 5, 510, &h('a')),
        "2026-11-14T10:05:00Z",
    )]);
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
}

/// The lane that failed the audit is the effective holder at the FAIL (`audit_claim::holder_of`
/// on the comments before it): the lowest reservation on the head, moved by the leader's
/// `moai-claim-handoff … by=leader` line posted before the FAIL. The handed-off lane keeps the
/// re-audit and is offered nothing; everyone else is offered help, with that lane named. A
/// handoff after the FAIL moves the hold, but the FAIL was still judged by the earlier holder.
#[test]
fn the_fails_auditor_is_the_holder_after_a_handoff_before_the_fail() {
    let now = "2026-11-14T13:00:00Z";
    let handed = |to: &str, when: &str| {
        vec![
            at(audit_claim(500, "agent-3", &h('a')), "2026-11-14T09:00:00Z"),
            handoff(505, to, "2026-11-14T09:30:00Z"),
            at(verdict(510, &h('a'), "FAIL"), when),
        ]
    };
    // Handed to me (or to my session spelling): the re-audit is mine, no offer.
    for to in [LANE, "agent-10-59cb6a"] {
        let r = help(now, handed(to, FAILED_AT));
        assert_eq!(r["next"], json!(null), "{to}: {r}");
        assert_eq!(
            r["pr_skipped"],
            json!([{"pr": 5, "reason": "audit_reserved", "by": to, "comment": 505}]),
            "{to}: {r}"
        );
    }
    // Handed to another lane: that lane is the auditor, and I am offered the repair help.
    let r = help(now, handed("agent-7", FAILED_AT));
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
    assert_eq!(r["next"]["evidence"]["failed_by"], json!("agent-7"), "{r}");
    assert_eq!(r["pr_skipped"], json!([]), "{r}");
    // A handoff after the FAIL does not change who failed the audit.
    let after = vec![
        at(audit_claim(500, "agent-3", &h('a')), "2026-11-14T09:00:00Z"),
        at(verdict(510, &h('a'), "FAIL"), FAILED_AT),
        handoff(515, "agent-7", "2026-11-14T10:05:00Z"),
    ];
    let r = help(now, after);
    assert_eq!(kind(&r), json!("offer_repair_help"), "{r}");
    assert_eq!(r["next"]["evidence"]["failed_by"], json!("agent-3"), "{r}");
    // The lowest reservation on the head holds it, not the newest.
    let two = vec![
        at(
            audit_claim(500, "agent-10", &h('a')),
            "2026-11-14T09:00:00Z",
        ),
        at(audit_claim(501, "agent-3", &h('a')), "2026-11-14T09:05:00Z"),
        at(verdict(510, &h('a'), "FAIL"), FAILED_AT),
    ];
    let r = help(now, two);
    assert_eq!(r["next"], json!(null), "{r}");
    assert_eq!(r["pr_skipped"][0]["by"], json!("agent-10"), "{r}");
}

// --- #2933 S6: a lane's wait record `waiting: <what> since <UTC>` ----------------------------

const WAIT_NOW: &str = "2026-11-14T10:20:00Z";
const WAIT_SINCE: &str = "2026-11-14T10:00:00Z";

/// A comment whose body is `body`.
fn wait_note(id: u64, body: &str) -> Value {
    json!({"id": id, "body": body})
}

/// A comment naming `LANE` on its first line with `line` below it.
fn lane_line(id: u64, line: &str) -> Value {
    wait_note(id, &format!("lane={LANE}\n{line}"))
}

/// A comment of the lane holding the wait record `waiting: <what> since <since>`.
fn rec(id: u64, what: &str, since: &str) -> Value {
    lane_line(id, &format!("waiting: {what} since {since}"))
}

/// The `lane_waiting` entries of a result, in order.
fn lane_waiting(r: &Value) -> Vec<Value> {
    r["waiting"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|w| w["reason"] == json!("lane_waiting"))
        .cloned()
        .collect()
}

/// Issue 2933 with the lane's live claim (comment 10), then `rest`, read at `WAIT_NOW`.
fn wait_issue(rest: Vec<Value>) -> Value {
    let mut comments = vec![claim(10, LANE)];
    comments.extend(rest);
    json!({"now": WAIT_NOW, "threads": [{"number": 2933, "title": "card", "comments": comments}]})
}

fn entry(number: u64, what: &str, comment: u64) -> Value {
    json!({"reason": "lane_waiting", "number": number, "what": what, "since": WAIT_SINCE,
        "age_secs": 1200, "comment": comment})
}

/// A wait record on the Issue the lane holds a live claim on is read, with its age on the
/// injected clock; the same on the lane's own pull request and on one whose audit the lane
/// reserved. A pull request of another lane that the lane did not reserve is not the lane's.
#[test]
fn a_wait_record_is_read_from_a_claimed_issue_an_own_pr_and_a_reserved_audit() {
    let line = format!("waiting: CI of PR #5 since {WAIT_SINCE}");
    // The record is the first line; the lane is named further down.
    let r = next(&wait_issue(vec![wait_note(
        11,
        &format!("{line}\nlane={LANE}"),
    )]));
    assert_eq!(
        lane_waiting(&r),
        vec![entry(2933, "CI of PR #5", 11)],
        "{r}"
    );
    // The record sits below the lane's name, after an empty line.
    let r = next(&wait_issue(vec![wait_note(
        11,
        &format!("lane={LANE}\n\n{line}\n"),
    )]));
    assert_eq!(
        lane_waiting(&r),
        vec![entry(2933, "CI of PR #5", 11)],
        "{r}"
    );
    // Own pull request, and a pull request of another lane whose audit this lane reserved.
    let note = lane_line(21, &line);
    let snap = json!({"now": WAIT_NOW, "prs": [
        pr(7, &h('a'), Some(LANE), json!([note.clone()])),
        pr(8, &h('b'), Some("agent-29"), json!([audit_claim(20, LANE, &h('b')), note.clone()])),
        pr(9, &h('c'), Some("agent-29"), json!([audit_claim(20, "agent-3", &h('c')), note.clone()])),
        pr(10, &h('d'), Some("agent-29"), json!([note.clone()])),
    ]});
    let r = next(&snap);
    assert_eq!(
        lane_waiting(&r),
        vec![entry(7, "CI of PR #5", 21), entry(8, "CI of PR #5", 21)],
        "{r}"
    );
    // The entry carries exactly these keys.
    let first = lane_waiting(&r)[0].clone();
    let mut keys: Vec<&str> = first
        .as_object()
        .unwrap()
        .keys()
        .map(String::as_str)
        .collect();
    keys.sort_unstable();
    assert_eq!(
        keys,
        ["age_secs", "comment", "number", "reason", "since", "what"],
        "{r}"
    );
}

/// An Issue is the lane's only with a live claim: another lane's claim, a released claim and a
/// thread with no claim at all give no entry.
#[test]
fn a_wait_record_on_an_issue_without_a_live_claim_of_the_lane_is_not_its_own() {
    let note = rec(11, "a reply", WAIT_SINCE);
    let other = json!({"now": WAIT_NOW, "threads": [{"number": 2933, "title": "card",
        "comments": [claim(10, "agent-99"), note.clone()]}]});
    assert_eq!(lane_waiting(&next(&other)), Vec::<Value>::new());
    let released = json!({"now": WAIT_NOW, "threads": [{"number": 2933, "title": "card",
        "comments": [claim(10, LANE),
            {"id": 12, "body": format!("moai-claim-release: comment=10 by={LANE}")}, note.clone()]}]});
    assert_eq!(lane_waiting(&next(&released)), Vec::<Value>::new());
    let bare = json!({"now": WAIT_NOW, "threads": [{"number": 2933, "title": "card",
        "comments": [note]}]});
    assert_eq!(lane_waiting(&next(&bare)), Vec::<Value>::new());
}

/// Every lane posts as one account, so a record counts as the lane's only when its comment names
/// the lane — as a whole word, under a session spelling too. Another lane's comment, `agent-100`
/// and `agent-1` are not `agent-10`, and another lane's `waiting: done` clears nothing.
#[test]
fn only_a_comment_that_names_the_lane_carries_its_wait_record() {
    let line = format!("waiting: a reply since {WAIT_SINCE}");
    let by = |who: &str| wait_note(11, &format!("{who}\n{line}"));
    for who in [
        "lane=agent-10",
        "agent-10",
        "agent-10-59cb6a",
        "session=agent-10-59cb6a",
        "by=agent-10-hBp17,",
    ] {
        let r = next(&wait_issue(vec![by(who)]));
        assert_eq!(
            lane_waiting(&r),
            vec![entry(2933, "a reply", 11)],
            "{who}: {r}"
        );
    }
    for who in [
        "lane=agent-3",
        "lane=agent-100",
        "agent-1",
        "agent-10x",
        "just a note",
    ] {
        let r = next(&wait_issue(vec![by(who)]));
        assert_eq!(lane_waiting(&r), Vec::<Value>::new(), "{who}: {r}");
    }
    // Another lane's `waiting: done` leaves mine standing.
    let r = next(&wait_issue(vec![
        rec(11, "a reply", WAIT_SINCE),
        wait_note(12, "lane=agent-3\nwaiting: done"),
    ]));
    assert_eq!(lane_waiting(&r), vec![entry(2933, "a reply", 11)], "{r}");
    // The same on a pull request the lane did not author or reserve.
    let theirs = json!({"now": WAIT_NOW, "prs": [
        pr(5, &h('a'), Some("agent-29"), json!([rec(11, "a reply", WAIT_SINCE)]))]});
    assert_eq!(lane_waiting(&next(&theirs)), Vec::<Value>::new());
}

/// The newest record of a number stands; a later comment that is no wait record clears nothing; a
/// later `waiting: done` or `waiting: -` clears it, and a record after that arms it again.
#[test]
fn the_newest_wait_record_stands_and_waiting_done_clears_it() {
    let chatter = lane_line(13, "just a note");
    // Newest wins (the list order is the age order).
    let r = next(&wait_issue(vec![
        rec(11, "CI", "2026-11-14T09:00:00Z"),
        rec(12, "an audit", WAIT_SINCE),
    ]));
    assert_eq!(lane_waiting(&r), vec![entry(2933, "an audit", 12)], "{r}");
    // A later comment of the lane that is no wait record does not clear it.
    let r = next(&wait_issue(vec![rec(12, "an audit", WAIT_SINCE), chatter]));
    assert_eq!(lane_waiting(&r), vec![entry(2933, "an audit", 12)], "{r}");
    // `done` and `-` clear it.
    for clear in [
        "waiting: done",
        "waiting: -",
        "waiting: done since 2026-11-14T10:10:00Z",
    ] {
        let r = next(&wait_issue(vec![
            rec(12, "an audit", WAIT_SINCE),
            lane_line(13, clear),
        ]));
        assert_eq!(lane_waiting(&r), Vec::<Value>::new(), "{clear}: {r}");
        // A record after the clearing line arms it again.
        let r = next(&wait_issue(vec![
            rec(12, "an audit", WAIT_SINCE),
            lane_line(13, clear),
            rec(14, "a shell", WAIT_SINCE),
        ]));
        assert_eq!(
            lane_waiting(&r),
            vec![entry(2933, "a shell", 14)],
            "{clear}: {r}"
        );
    }
    // Within one comment the last line wins.
    let r = next(&wait_issue(vec![lane_line(
        12,
        &format!("waiting: CI since {WAIT_SINCE}\nwaiting: done"),
    )]));
    assert_eq!(lane_waiting(&r), Vec::<Value>::new(), "{r}");
    // The number decides: a clearing line on one Issue does not clear another's record.
    let snap = json!({"now": WAIT_NOW, "threads": [
        {"number": 2933, "title": "a", "comments": [claim(10, LANE), rec(11, "CI", WAIT_SINCE)]},
        {"number": 2934, "title": "b", "comments": [claim(20, LANE), lane_line(21, "waiting: done")]}]});
    let r = next(&snap);
    assert_eq!(lane_waiting(&r), vec![entry(2933, "CI", 11)], "{r}");
}

/// A line that is not `waiting: <what> since <YYYY-MM-DDTHH:MM:SSZ>` at column 0 is no record: it
/// is ignored, and it neither clears nor replaces an older valid record.
#[test]
fn a_malformed_wait_line_is_ignored() {
    let valid = rec(11, "CI", WAIT_SINCE);
    let bad = [
        "waiting: CI",
        "waiting: CI since",
        "waiting: CI since yesterday",
        "waiting: CI since 2026-11-14T10:00:00",
        "waiting: CI since 2026-11-14 10:00:00Z",
        "waiting: CI since 2026-13-14T10:00:00Z",
        "waiting: CI since 2026-11-14T25:00:00Z",
        "waiting: CI since 2026-11-14T-1:00:00Z",
        "waiting: CI since 2026-11-14T10:00:00Z extra",
        "waiting: since 2026-11-14T10:00:00Z",
        "waiting:",
        "waiting: finished",
        " waiting: CI since 2026-11-14T10:00:00Z",
        "\twaiting: CI since 2026-11-14T10:00:00Z",
        "Waiting: CI since 2026-11-14T10:00:00Z",
        "not waiting: CI since 2026-11-14T10:00:00Z",
    ];
    for (i, line) in bad.iter().enumerate() {
        let id = 20 + i as u64;
        // Alone: nothing.
        let r = next(&wait_issue(vec![lane_line(id, line)]));
        assert_eq!(lane_waiting(&r), Vec::<Value>::new(), "{line:?}: {r}");
        // After a valid record: the valid one stands.
        let r = next(&wait_issue(vec![valid.clone(), lane_line(id, line)]));
        assert_eq!(
            lane_waiting(&r),
            vec![entry(2933, "CI", 11)],
            "{line:?}: {r}"
        );
    }
}

/// `<what>` is cut at 160 characters (N = 160 stays whole, N + 1 is cut), counted in characters;
/// control characters are stripped, never echoed.
#[test]
fn what_is_capped_at_160_characters_and_carries_no_control_character() {
    for (n, kept) in [(1, 1), (159, 159), (160, 160), (161, 160), (400, 160)] {
        let r = next(&wait_issue(vec![rec(11, &"x".repeat(n), WAIT_SINCE)]));
        let got = lane_waiting(&r);
        assert_eq!(got.len(), 1, "{n}: {r}");
        assert_eq!(got[0]["what"], json!("x".repeat(kept)), "{n}: {r}");
    }
    // Characters, not bytes: 160 two-byte characters are whole, 161 are cut.
    for (n, kept) in [(160, 160), (161, 160)] {
        let r = next(&wait_issue(vec![rec(11, &"é".repeat(n), WAIT_SINCE)]));
        assert_eq!(
            lane_waiting(&r)[0]["what"],
            json!("é".repeat(kept)),
            "{n}: {r}"
        );
    }
    // Escape, tab, bell and a bare carriage return are stripped; the text between stays.
    let r = next(&wait_issue(vec![rec(
        11,
        "\u{1b}[31mred\u{7}\ta\rb",
        WAIT_SINCE,
    )]));
    assert_eq!(lane_waiting(&r)[0]["what"], json!("[31mredab"), "{r}");
    // Stripped to nothing: no record.
    let r = next(&wait_issue(vec![rec(11, "\u{1b}\u{7}", WAIT_SINCE)]));
    assert_eq!(lane_waiting(&r), Vec::<Value>::new(), "{r}");
}

/// `age_secs` is `now` − `since` on the injected clock; with no (or an unreadable) `now` it is
/// `null` — never the real clock — and a `since` ahead of `now` is age 0.
#[test]
fn the_age_comes_from_the_injected_clock() {
    let note = rec(11, "CI", WAIT_SINCE);
    for (now, age) in [
        ("2026-11-14T10:00:00Z", 0),
        ("2026-11-14T10:00:01Z", 1),
        ("2026-11-14T10:20:00Z", 1200),
        ("2026-11-15T10:00:00Z", 86_400),
        ("2026-11-14T09:59:59Z", 0),
    ] {
        let mut s = wait_issue(vec![note.clone()]);
        s["now"] = json!(now);
        let r = next(&s);
        assert_eq!(lane_waiting(&r)[0]["age_secs"], json!(age), "{now}: {r}");
        assert_eq!(
            lane_waiting(&r)[0]["since"],
            json!(WAIT_SINCE),
            "{now}: {r}"
        );
    }
    let mut s = wait_issue(vec![note.clone()]);
    s.as_object_mut().unwrap().remove("now");
    let r = next(&s);
    assert_eq!(lane_waiting(&r).len(), 1, "{r}");
    assert_eq!(lane_waiting(&r)[0]["age_secs"], json!(null), "{r}");
    assert!(r.get("now").is_none(), "{r}");
    let mut s = wait_issue(vec![note]);
    s["now"] = json!("soon");
    let r = next(&s);
    assert_eq!(lane_waiting(&r)[0]["age_secs"], json!(null), "{r}");
}

/// A wait record is information only: the same snapshot with and without it gives the same task,
/// the same skipped lists and the same other `waiting` entries, in every branch of the list.
#[test]
fn a_wait_record_changes_no_task_and_no_order() {
    // Each builder puts `extra` comments on a thread the lane owns: its own pull request 7, a
    // pull request whose audit it reserved, or its claimed Issue 2933.
    type Build = Box<dyn Fn(Vec<Value>) -> Value>;
    let cases: Vec<(Build, Value)> = vec![
        // An own pull request with an audit FAIL, a free Issue behind it.
        (
            Box::new(|extra| {
                let mut c = vec![verdict(30, &h('a'), "FAIL")];
                c.extend(extra);
                json!({"now": WAIT_NOW, "prs": [pr(7, &h('a'), Some(LANE), json!(c))],
                    "open_issues": [{"number": 2600, "title": "f", "comments": []}]})
            }),
            json!("repair_audit_fail"),
        ),
        // An audit the lane reserved.
        (
            Box::new(|extra| {
                let mut c = vec![audit_claim(30, LANE, &h('a'))];
                c.extend(extra);
                json!({"now": WAIT_NOW, "prs": [pr(7, &h('a'), Some("agent-29"), json!(c))]})
            }),
            json!("audit"),
        ),
        // A claimed card to continue.
        (
            Box::new(|extra| {
                let mut c = vec![claim(10, LANE)];
                c.extend(extra);
                json!({"now": WAIT_NOW,
                    "threads": [{"number": 2933, "title": "card", "comments": c}]})
            }),
            json!("continue_card"),
        ),
        // A free Issue to pick after a held one.
        (
            Box::new(|extra| {
                json!({"now": WAIT_NOW,
                    "open_issues": [
                        {"number": 2601, "title": "held", "labels": ["hold"], "comments": []},
                        {"number": 2600, "title": "free", "comments": []}],
                    "prs": [pr(7, &h('a'), Some(LANE), json!(extra))]})
            }),
            json!("pick_issue"),
        ),
        // Nothing to do at all.
        (
            Box::new(
                |extra| json!({"now": WAIT_NOW, "prs": [pr(7, &h('a'), Some(LANE), json!(extra))]}),
            ),
            json!(null),
        ),
    ];
    let others = |r: &Value| -> Vec<Value> {
        r["waiting"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|w| w["reason"] != json!("lane_waiting"))
            .cloned()
            .collect()
    };
    for (i, (build, expected)) in cases.iter().enumerate() {
        let without = next(&build(vec![]));
        let with = next(&build(vec![rec(40, "CI", WAIT_SINCE)]));
        assert_eq!(without["next"]["kind"], *expected, "case {i}: {without}");
        assert_eq!(lane_waiting(&without), Vec::<Value>::new(), "case {i}");
        // The record is read ...
        assert_eq!(lane_waiting(&with).len(), 1, "case {i}: {with}");
        // ... and changes nothing else.
        for key in [
            "next",
            "skipped",
            "pr_skipped",
            "end_condition",
            "partial",
            "now",
        ] {
            assert_eq!(
                with[key], without[key],
                "case {i} {key}: {with} vs {without}"
            );
        }
        assert_eq!(others(&with), others(&without), "case {i}: {with}");
    }
}

/// #3005 P8: a watching lane stops by itself. `--max-empty N` counts reads with no task and
/// nothing to wait for, per lane and repository, in the clone's state folder.
mod idle_limit {
    use super::*;
    use crate::common::git::git_ok;

    struct Clone {
        _tmp: TempDir,
        repo: std::path::PathBuf,
    }

    fn clone() -> Clone {
        let tmp = TempDir::new().unwrap();
        let repo = std::fs::canonicalize(tmp.path()).unwrap().join("clone");
        std::fs::create_dir_all(&repo).unwrap();
        git_ok(&repo, &["init", "-q", "-b", "main"]);
        Clone { _tmp: tmp, repo }
    }

    fn read(c: &Clone, lane: &str, repo: &str, snap: &Value, max: &str) -> (Value, Value) {
        let file = c.repo.join("snap.json");
        std::fs::write(&file, snap.to_string()).unwrap();
        let out = Command::new(EXE)
            .args([
                "lane-next",
                "--lane",
                lane,
                "--repo",
                repo,
                "--max-empty",
                max,
            ])
            .arg("--snapshot")
            .arg(&file)
            .arg("--dir")
            .arg(&c.repo)
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
        (v["result"].clone(), v)
    }

    /// A full read with nothing to do: the card list was read (`cards`), so the lane-wide end
    /// is known (`met: true`) and `partial` is empty.
    fn nothing() -> Value {
        json!({"now": "2026-11-15T02:00:00Z", "cards": []})
    }

    fn waiting() -> Value {
        let blocked = json!({"id": 50, "body": "LANE-BLOCKED kind=decision owner=leader #2827 B1 — (가)/(나), safe default (나)"});
        json!({"cards": [], "threads": [{"number": 2812, "comments": [claim(40, LANE), blocked]}]})
    }

    /// The lane's own open pull request, no blocker now (CI pending): the lane is in flight.
    fn own_pr_in_flight() -> Value {
        json!({"cards": [], "prs": [{"number": 10, "head": h('a'), "author_lane": LANE,
            "comments": [], "ci": "pending"}]})
    }

    /// The lane's own draft pull request: `pr_skipped` names it `draft`, and `own_open_prs`
    /// lists it, so the lane is in flight (the leader's NN6 decision on #3020).
    fn own_draft_pr() -> Value {
        json!({"cards": [], "prs": [{"number": 10, "head": h('a'), "author_lane": LANE,
            "draft": true, "comments": []}]})
    }

    fn state_file(c: &Clone) -> std::path::PathBuf {
        c.repo
            .join(".git/.moai-cli/lane-idle")
            .join(format!("{LANE}__o__r.json"))
    }

    /// An own draft pull request is work in flight: four reads with `--max-empty 3` do not
    /// stop the lane and the count stays 0.
    #[test]
    fn an_own_draft_pull_request_is_not_an_empty_read() {
        let c = clone();
        for n in 1..=4 {
            let (r, _) = read(&c, LANE, "o/r", &own_draft_pr(), "3");
            assert_eq!(
                r["pr_skipped"],
                json!([{"pr": 10, "reason": "draft"}]),
                "the snapshot is the own-draft case: {r}"
            );
            assert_eq!(r["own_open_prs"], json!([10]), "{r}");
            assert!(r.get("status").is_none(), "read {n} stopped the lane: {r}");
            assert_eq!(r["idle"]["consecutive_empty"], json!(0), "read {n}: {r}");
        }
    }

    /// A draft of another lane is not the lane's own work: it is not in `own_open_prs`, and
    /// the empty reads count to the limit.
    #[test]
    fn a_foreign_draft_pull_request_does_not_keep_the_lane_busy() {
        let c = clone();
        let snap = json!({"cards": [], "prs": [{"number": 11, "head": h('b'),
            "author_lane": "agent-99", "draft": true, "comments": []}]});
        let mut last = Value::Null;
        for _ in 1..=3 {
            let (r, _) = read(&c, LANE, "o/r", &snap, "3");
            assert_eq!(r["own_open_prs"], json!([]), "{r}");
            last = r;
        }
        assert_eq!(last["status"], json!("idle_limit"), "{last}");
    }

    /// A lane with its own open pull request waiting on CI is not idle: three reads with
    /// `--max-empty 3` do not stop it, and the count stays 0 (the leader's repro of #3020 B1).
    #[test]
    fn an_own_pull_request_in_flight_is_not_an_empty_read() {
        let c = clone();
        for n in 1..=4 {
            let (r, _) = read(&c, LANE, "o/r", &own_pr_in_flight(), "3");
            assert_eq!(r["next"], json!(null), "read {n}: {r}");
            assert_eq!(
                r["pr_skipped"],
                json!([{"pr": 10, "reason": "own_no_blocker"}]),
                "the snapshot is the own-PR case: {r}"
            );
            assert!(r.get("status").is_none(), "read {n} stopped the lane: {r}");
            assert_eq!(r["idle"]["consecutive_empty"], json!(0), "read {n}: {r}");
        }
        // Empty reads before it do not survive it: 2 empty, the PR, then 2 empty is not 4.
        let c = clone();
        for snap in [
            nothing(),
            nothing(),
            own_pr_in_flight(),
            nothing(),
            nothing(),
        ] {
            read(&c, LANE, "o/r", &snap, "3");
        }
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "3");
        assert_eq!(r["idle"]["consecutive_empty"], json!(3), "{r}");
        assert_eq!(r["status"], json!("idle_limit"), "3 since the PR: {r}");
    }

    /// With nothing open and no `--tip` every read carries the note that no tip base is known,
    /// and a busy Issue always carries the comment-window note ("only the newest 100 comments
    /// read"); those notes alone are not a failed read, or a lane could never go idle.
    #[test]
    fn the_always_present_partial_notes_alone_do_not_stop_the_count() {
        let notes = [
            json!([moai_cli::lane_next::TIP_BASE_UNKNOWN]),
            json!(["#1975: only the newest 100 comments read"]),
        ];
        for note in notes {
            let c = clone();
            let snap = json!({"cards": [], "partial": note});
            for n in 1..=2u64 {
                let (r, _) = read(&c, LANE, "o/r", &snap, "3");
                assert_eq!(r["idle"]["consecutive_empty"], json!(n), "{r}");
            }
            let (r, _) = read(&c, LANE, "o/r", &snap, "3");
            assert_eq!(r["status"], json!("idle_limit"), "{r}");
        }
    }

    /// A read GitHub answered in part (`partial`), or whose lane-wide end is unknown
    /// (`met: null`), is not an empty read: three of them with `--max-empty 3` do not stop the
    /// lane, change the saved count or write the file (the leader's repro of #3020 B2).
    #[test]
    fn a_partial_read_is_not_counted() {
        let partial = json!({"cards": [], "partial": ["open pull requests: HTTP 502"]});
        let unknown = json!({"now": "2026-11-15T02:00:00Z"});
        for (name, snap) in [("partial", partial), ("unknown end", unknown)] {
            let c = clone();
            for n in 1..=4 {
                let (r, _) = read(&c, LANE, "o/r", &snap, "3");
                assert!(r.get("status").is_none(), "{name} read {n}: {r}");
                assert_eq!(r["idle"]["consecutive_empty"], json!(0), "{name}: {r}");
                assert_eq!(r["idle"]["limit_reached"], json!(false), "{name}: {r}");
                assert!(r["idle"]["not_counted"].is_string(), "{name}: {r}");
                assert_eq!(r["writes"], json!(false), "{name}: nothing written: {r}");
            }
            assert!(!state_file(&c).exists(), "{name} wrote the state file");
        }
        // Between full reads a partial one neither adds nor resets: 2 empty, partial, empty is 3.
        let c = clone();
        let part = json!({"cards": [], "partial": ["x"]});
        for snap in [nothing(), nothing(), part] {
            read(&c, LANE, "o/r", &snap, "3");
        }
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "3");
        assert_eq!(r["idle"]["consecutive_empty"], json!(3), "{r}");
        assert_eq!(r["status"], json!("idle_limit"), "{r}");
    }

    /// Reads 1 to 5 go on, read 6 is `idle_limit` with the `LANE-IDLE` line, and the count
    /// stays in a file under the clone's state folder.
    #[test]
    fn the_sixth_empty_read_is_the_limit_and_the_fifth_is_not() {
        let c = clone();
        for n in 1..=5u64 {
            let (r, _) = read(&c, LANE, "o/r", &nothing(), "6");
            assert_eq!(r["next"], json!(null), "{r}");
            assert_eq!(r["idle"]["consecutive_empty"], json!(n), "{r}");
            assert_eq!(r["idle"]["limit_reached"], json!(false), "read {n}: {r}");
            assert!(r.get("status").is_none(), "read {n}: {r}");
            assert_eq!(r["writes"], json!(true));
        }
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "6");
        assert_eq!(r["status"], json!("idle_limit"), "{r}");
        assert_eq!(r["idle"]["consecutive_empty"], json!(6));
        let line = r["idle"]["line"].as_str().unwrap();
        assert!(
            line.starts_with(&format!("LANE-IDLE lane={LANE} checks=6 last=20")),
            "{line}"
        );
        let state = c
            .repo
            .join(".git/.moai-cli/lane-idle")
            .join(format!("{LANE}__o__r.json"));
        assert!(state.is_file(), "{}", state.display());
    }

    /// A read with a task, or with something to wait for, sets the count back to 0.
    #[test]
    fn a_task_or_a_wait_resets_the_count() {
        let c = clone();
        for _ in 0..4 {
            read(&c, LANE, "o/r", &nothing(), "6");
        }
        let (r, _) = read(&c, LANE, "o/r", &waiting(), "6");
        assert_eq!(r["next"], json!(null), "{r}");
        assert_eq!(
            r["idle"]["consecutive_empty"],
            json!(0),
            "a wait is not empty: {r}"
        );
        for _ in 0..3 {
            read(&c, LANE, "o/r", &nothing(), "6");
        }
        let task = json!({"prs": [pr(5, &h('a'), Some("agent-3"), json!([]))]});
        let (r, _) = read(&c, LANE, "o/r", &task, "6");
        assert!(!r["next"].is_null(), "{r}");
        assert_eq!(r["idle"]["consecutive_empty"], json!(0), "{r}");
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "6");
        assert_eq!(
            r["idle"]["consecutive_empty"],
            json!(1),
            "counting starts over: {r}"
        );
    }

    /// Another lane, or another repository, has its own count.
    #[test]
    fn lanes_and_repositories_do_not_share_a_count() {
        let c = clone();
        for _ in 0..3 {
            read(&c, LANE, "o/r", &nothing(), "6");
        }
        let (r, _) = read(&c, "agent-11", "o/r", &nothing(), "6");
        assert_eq!(r["idle"]["consecutive_empty"], json!(1), "{r}");
        let (r, _) = read(&c, LANE, "o/s", &nothing(), "6");
        assert_eq!(r["idle"]["consecutive_empty"], json!(1), "{r}");
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "6");
        assert_eq!(r["idle"]["consecutive_empty"], json!(4), "{r}");
    }

    /// The two edges of the option: 0 counts nothing and writes nothing, 1 stops at once, and
    /// more than 1000 is refused.
    #[test]
    fn max_empty_zero_one_and_over_the_largest() {
        let c = clone();
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "0");
        assert!(
            r.get("idle").is_none() && r["writes"] == json!(false),
            "{r}"
        );
        assert!(!c.repo.join(".git/.moai-cli/lane-idle").exists());
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "1");
        assert_eq!(r["status"], json!("idle_limit"), "{r}");
        let (r, v) = read(&c, LANE, "o/r", &nothing(), "1001");
        assert!(r.is_null() && v["ok"] == json!(false), "{v}");
        let (r, _) = read(&c, LANE, "o/r", &nothing(), "1000");
        assert_eq!(r["idle"]["max_empty"], json!(1000), "{r}");
    }

    /// Without `--max-empty` a `--snapshot` read counts nothing and writes nothing (the
    /// measurement use of the tool stays read-only); outside a git clone the count is off and
    /// the lane keeps going.
    #[test]
    fn a_snapshot_read_without_the_option_and_a_folder_that_is_no_clone_count_nothing() {
        let c = clone();
        let file = c.repo.join("snap.json");
        std::fs::write(&file, nothing().to_string()).unwrap();
        let out = Command::new(EXE)
            .args(["lane-next", "--lane", LANE, "--repo", "o/r", "--snapshot"])
            .arg(&file)
            .arg("--dir")
            .arg(&c.repo)
            .output()
            .unwrap();
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert!(v["result"].get("idle").is_none(), "{v}");
        assert_eq!(v["result"]["writes"], json!(false));
        assert!(!c.repo.join(".git/.moai-cli/lane-idle").exists());

        let plain = TempDir::new().unwrap();
        let (r, v) = read(
            &Clone {
                repo: plain.path().to_path_buf(),
                _tmp: TempDir::new().unwrap(),
            },
            LANE,
            "o/r",
            &nothing(),
            "6",
        );
        assert_eq!(v["ok"], json!(true), "{v}");
        assert_eq!(r["idle"]["counting"], json!(false), "{r}");
        assert!(r.get("status").is_none(), "{r}");
    }
}
