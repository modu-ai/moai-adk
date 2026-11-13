//! rc.2 A11 (#2265) `lane-status` over recorded GitHub exchanges (`RecordedTransport`, injected
//! clock, no network) and a real temporary git repository for the worktree list.
//! Research record: docs/design-review/lane-status-a11-2026-11-10.md.
mod common;

use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, RecordedTransport};
use moai_cli::lane_status::{status_cli, IdleCause, LaneStatusRequest, Phase};
use serde_json::{json, Value};
use std::{cell::Cell, fs, path::Path};

const API: &str = "https://api.github.com/repos/acme/widgets";
const LIVE: &str = "2030-01-01T00:00:00Z";
const NOW: i64 = 1_794_000_000; // 2026-11-07, before LIVE

struct FakeClock(Cell<i64>);
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.0.get()
    }
    fn sleep(&self, secs: u64) {
        self.0.set(self.0.get() + secs as i64);
    }
}

fn sha(c: char) -> String {
    std::iter::repeat_n(c, 40).collect()
}

fn get(url: &str, body: Value) -> Exchange {
    common::exchange("GET", url, common::json_response(200, body))
}

fn claim(id: i64, session: &str, epoch: u64) -> Value {
    json!({"id": id, "created_at": format!("2026-11-10T00:00:{:02}Z", id % 60),
           "body": format!("moai-claim: session={session} epoch={epoch} lease_until={LIVE}")})
}

fn note(id: i64, body: &str) -> Value {
    json!({"id": id, "created_at": format!("2026-11-10T00:01:{:02}Z", id % 60), "body": body})
}

fn issue(n: i64, body: &str) -> Value {
    json!({"number": n, "state": "open", "body": body})
}

/// The recorded remote. Issues 11-19:
/// 11 agent-1 claimed · 12 agent-2 implementing (worktree) · 13 agent-3 PR #113 with
/// duplicate and out-of-order reports · 14 handed off to agent-5 · 15 agent-6 PR #115 PASS
/// (waiting) · 16 no claim (ready) · 17 agent-7 PR #117 whose head no report names ·
/// 18 agent-8 blocked by open #99 · 19 agent-9 (a pending user decision from the caller).
fn remote(reverse: bool) -> Vec<Exchange> {
    let (a, b) = (sha('a'), sha('b'));
    let (c, d) = (sha('c'), sha('d'));
    let (e, f) = (sha('e'), sha('f'));
    let base = sha('0');
    let mut prs = vec![
        json!({"number": 113, "head": {"sha": b, "ref": "issue/13-x"}, "body": "x\n\nRefs #13\n"}),
        json!({"number": 115, "head": {"sha": d, "ref": "issue/15-x"}, "body": "Refs #15"}),
        json!({"number": 117, "head": {"sha": f, "ref": "issue/17-x"}, "body": "Refs #17"}),
        json!({"number": 200, "head": {"sha": sha('9'), "ref": "other"}, "body": "Refs #999"}),
    ];
    let mut comments: Vec<(i64, Vec<Value>)> = vec![
        (11, vec![claim(1101, "agent-1", 1)]),
        (12, vec![claim(1201, "agent-2", 1)]),
        (
            13,
            vec![
                claim(1301, "agent-3", 1),
                note(1302, &format!("report: pushed head `{a}` (base {base})")),
                note(1303, &format!("report again: head {a}")),
                note(1304, &format!("report: head {b}")),
                note(1305, &format!("late report: head {a}")),
            ],
        ),
        (
            14,
            vec![
                claim(1401, "agent-4", 1),
                note(
                    1402,
                    "moai-claim-handoff: comment=1401 to=agent-5 by=leader",
                ),
            ],
        ),
        (
            15,
            vec![
                claim(1501, "agent-6", 1),
                note(
                    1502,
                    &format!("AUDIT #15 (k) head {d} base {base} — verdict: PASS"),
                ),
            ],
        ),
        (16, vec![note(1601, "design notes, no claim")]),
        (
            17,
            vec![
                claim(1701, "agent-7", 1),
                note(1702, &format!("report: head {e}")),
            ],
        ),
        (18, vec![claim(1801, "agent-8", 1)]),
        (19, vec![claim(1901, "agent-9", 1)]),
    ];
    // The commit lists stay in order: the API's order is the record of head order.
    let commits = [
        (113, vec![a.clone(), b.clone()]),
        (115, vec![c.clone(), d.clone()]),
        (117, vec![e.clone(), f.clone()]),
    ];
    if reverse {
        prs.reverse();
        for (_, list) in comments.iter_mut() {
            list.reverse();
        }
    }
    let mut out = vec![get(
        &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
        Value::Array(prs),
    )];
    for (n, list) in comments {
        let body = match n {
            18 => "상위: #10\nBlocked-By: #99",
            _ => "상위: #10\nBlocked-By: 없음",
        };
        out.push(get(&format!("{API}/issues/{n}"), issue(n, body)));
        out.push(get(
            &format!("{API}/issues/{n}/comments?per_page=100"),
            Value::Array(list),
        ));
    }
    out.push(get(&format!("{API}/issues/99"), issue(99, "")));
    for (pr, list) in commits.iter() {
        let body: Vec<Value> = list.iter().map(|s| json!({"sha": s})).collect();
        out.push(get(
            &format!("{API}/pulls/{pr}/commits?per_page=100"),
            Value::Array(body),
        ));
        // #2403 item 3: the PR's own comments (audit reservations and verdicts).
        out.push(get(
            &format!("{API}/issues/{pr}/comments?per_page=100"),
            json!([]),
        ));
    }
    out
}

fn request(lanes: Value, decisions: Value) -> LaneStatusRequest {
    serde_json::from_value(json!({
        "schema_version": 1,
        "owner": "acme",
        "repo": "widgets",
        "base_branch": "v2026.11.0",
        "issues": [19, 11, 12, 13, 14, 15, 16, 17, 18, 13],
        "lanes": lanes,
        "pending_decisions": decisions,
    }))
    .unwrap()
}

fn lanes() -> Value {
    let pane = |state: &str| json!({"state": state, "source": "hc get t6x", "observed_at": "2026-11-10T09:00:00Z"});
    json!([
        {"session": "agent-1", "name": "lane-1"},
        {"session": "agent-2", "pane": pane("working")},
        {"session": "agent-3"}, {"session": "agent-5"}, {"session": "agent-6"},
        {"session": "agent-7"}, {"session": "agent-8"}, {"session": "agent-9"},
        {"session": "agent-10"},
        {"session": "agent-11", "pane": pane("unreachable")},
        {"session": "agent-12", "pane": pane("permission_prompt")},
        {"session": "agent-13", "pane": pane("input_pending")},
    ])
}

fn decisions() -> Value {
    json!([{"issue": 19, "source": "decisions-2026-11-10.md HOLD row", "recorded_at": "2026-11-10T08:00:00Z"}])
}

fn git(dir: &Path, args: &[&str]) {
    let status = common::git::command()
        .args(args)
        .current_dir(dir)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_INDEX_FILE")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("GIT_AUTHOR_NAME", "t")
        .env("GIT_AUTHOR_EMAIL", "t@example.invalid")
        .env("GIT_COMMITTER_NAME", "t")
        .env("GIT_COMMITTER_EMAIL", "t@example.invalid")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?}");
}

/// A project repository with one lane worktree, `issue/12-x`.
fn project(root: &Path) -> std::path::PathBuf {
    let p = root.join("primary");
    fs::create_dir(&p).unwrap();
    git(&p, &["init", "-q", "-b", "main"]);
    fs::write(p.join("a.txt"), "a\n").unwrap();
    git(&p, &["add", "a.txt"]);
    git(&p, &["commit", "-q", "-m", "a"]);
    let wt = root.join("wt-12");
    git(
        &p,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "issue/12-x",
            wt.to_str().unwrap(),
        ],
    );
    p
}

fn run(project: &Path, req: LaneStatusRequest, exchanges: Vec<Exchange>) -> Value {
    let t = RecordedTransport::new(exchanges);
    let clock = FakeClock(Cell::new(NOW));
    let out = status_cli(project, req, &t, &clock).unwrap();
    assert_eq!(t.remaining(), 0, "every recorded exchange was used");
    assert!(
        t.sent.borrow().iter().all(|r| r.method == "GET"),
        "GET only"
    );
    out
}

/// The summary line recomputed from the arrays of one result (#2678).
fn summary_of(out: &Value) -> String {
    let named = out["idle"].as_array().unwrap();
    let idle = named.iter().filter(|l| l["idle"] == json!(true)).count();
    let rows = out["rows"].as_array().unwrap().len();
    let ready = out["ready"].as_array().unwrap().len();
    let counts = format!(
        "{} lane(s) named, {idle} idle; {ready} ready Issue(s)",
        named.len()
    );
    if rows == 0 {
        format!(
            "0 rows: no claim, pull request or worktree was found for the named Issues; {counts}"
        )
    } else {
        format!("{rows} row(s); {counts}")
    }
}

fn row<'a>(out: &'a Value, session: &str, issue: i64) -> &'a Value {
    out["rows"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["session"] == json!(session) && r["issue"] == json!(issue))
        .unwrap_or_else(|| panic!("no row ({session}, #{issue}) in {out}"))
}

fn idle<'a>(out: &'a Value, session: &str) -> &'a Value {
    out["idle"]
        .as_array()
        .unwrap()
        .iter()
        .find(|r| r["session"] == json!(session))
        .unwrap_or_else(|| panic!("no idle entry for {session} in {out}"))
}

/// AC-1: one table over nine cards and twelve lanes, every phase present, each with the
/// record it was read from, ownership from the claim comments.
#[test]
fn lane_status_covers_every_phase_with_its_evidence() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let out = run(&p, request(lanes(), decisions()), remote(false));
    // #2678: every count is said in words, and the idle count is the lanes judged idle, not
    // the lanes named (audit B1/B2 6850252311): 12 lanes named, 9 of them `idle: true`.
    assert_eq!(out["summary"], json!(summary_of(&out)));
    assert_eq!(
        out["summary"],
        json!("8 row(s); 12 lane(s) named, 9 idle; 1 ready Issue(s)")
    );

    let r = row(&out, "agent-1", 11);
    assert_eq!(r["phase"], json!("claimed"));
    assert_eq!(r["phase_evidence"]["claim_comment_id"], json!(1101));
    assert_eq!(r["lane"], json!("lane-1"));
    assert_eq!(r["ownership"]["claim_comment_id"], json!(1101));
    assert_eq!(r["ownership"]["live"], json!(true));

    let r = row(&out, "agent-2", 12);
    assert_eq!(r["phase"], json!("implementing"));
    assert_eq!(r["branch"], json!("issue/12-x"));
    assert_eq!(
        r["worktree"],
        json!(root.join("wt-12").display().to_string())
    );
    assert_eq!(r["phase_evidence"]["worktree"], r["worktree"]);

    let r = row(&out, "agent-3", 13);
    assert_eq!(r["phase"], json!("pr_open"));
    assert_eq!(
        (r["pr"].clone(), r["head"].clone()),
        (json!(113), json!(sha('b')))
    );
    assert_eq!(r["branch"], json!("issue/13-x"));

    let r = row(&out, "agent-5", 14);
    assert_eq!(r["phase"], json!("handoff"));
    assert_eq!(r["phase_evidence"]["handoff_comment_id"], json!(1402));
    assert_eq!(r["ownership"]["claim_comment_id"], Value::Null);

    let r = row(&out, "agent-6", 15);
    assert_eq!(r["phase"], json!("waiting"));
    assert_eq!(r["phase_evidence"]["verdict"], json!("PASS"));
    assert_eq!(r["phase_evidence"]["verdict_comment_id"], json!(1502));

    // Every phase variant appears in this fixture.
    let phases: Vec<Value> = out["rows"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r["phase"].clone())
        .collect();
    for p in Phase::ALL {
        assert!(
            phases.contains(&json!(p.as_str())),
            "{p:?} missing in {phases:?}"
        );
    }
    // The issue listed twice is one row; #16 (no claim) and the foreign PR #200 make none.
    assert_eq!(out["rows"].as_array().unwrap().len(), 8, "{out}");
    assert_eq!(out["writes"], json!("none"));
}

/// AC-2: every idle cause, each with its evidence; `unknown` when no record supports one.
#[test]
fn idle_causes_are_classified_from_records_only() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let out = run(&p, request(lanes(), decisions()), remote(false));

    let cause = |s: &str| idle(&out, s)["cause"].clone();
    assert_eq!(cause("agent-2"), Value::Null);
    assert_eq!(idle(&out, "agent-2")["idle"], json!(false));
    assert_eq!(cause("agent-3"), json!("waiting_audit"));
    assert_eq!(idle(&out, "agent-3")["evidence"]["pr"], json!(113));
    assert_eq!(cause("agent-6"), json!("waiting_integration"));
    assert_eq!(cause("agent-8"), json!("waiting_prerequisite"));
    assert_eq!(idle(&out, "agent-8")["evidence"]["blocked_by"], json!([99]));
    assert_eq!(cause("agent-9"), json!("user_decision"));
    assert_eq!(
        idle(&out, "agent-9")["evidence"]["source"],
        json!("decisions-2026-11-10.md HOLD row")
    );
    assert_eq!(cause("agent-10"), json!("ready_card_unassigned"));
    assert_eq!(idle(&out, "agent-10")["evidence"]["ready"], json!([16]));
    assert_eq!(cause("agent-11"), json!("unreachable"));
    assert_eq!(
        idle(&out, "agent-11")["evidence"]["source"],
        json!("hc get t6x")
    );
    assert_eq!(cause("agent-12"), json!("permission_prompt"));
    assert_eq!(cause("agent-13"), json!("input_pending"));
    // An active card and no pane observation: nothing says whether the lane is idle.
    assert_eq!(cause("agent-1"), json!("unknown"));
    assert_eq!(idle(&out, "agent-1")["idle"], Value::Null);
    assert_eq!(
        idle(&out, "agent-1")["evidence"]["active_issues"],
        json!([11])
    );
    assert_eq!(cause("agent-5"), json!("unknown"));

    // The draft pairs the lane without a card with the ready card; nothing is posted.
    assert_eq!(
        out["assignment_drafts"],
        json!([{"lane": "agent-10", "issue": 16, "posted": false}])
    );

    // `no_ready_card`: the same remote without the unclaimed #16 in the inspected set.
    let mut req = request(lanes(), decisions());
    req.issues.retain(|n| *n != 16);
    let mut ex = remote(false);
    ex.retain(|e| !e.request.url.contains("/issues/16"));
    let out2 = run(&p, req, ex);
    assert_eq!(idle(&out2, "agent-10")["cause"], json!("no_ready_card"));
    assert_eq!(out2["assignment_drafts"], json!([]));
    // #2678: the same counts with no ready Issue left.
    assert_eq!(out2["summary"], json!(summary_of(&out2)));
    assert!(
        out2["summary"]
            .as_str()
            .unwrap()
            .ends_with("; 0 ready Issue(s)"),
        "{}",
        out2["summary"]
    );

    // Every idle cause variant appears across the two runs.
    let mut seen: Vec<Value> = Vec::new();
    for o in [&out, &out2] {
        seen.extend(
            o["idle"]
                .as_array()
                .unwrap()
                .iter()
                .map(|i| i["cause"].clone()),
        );
    }
    for c in IdleCause::ALL {
        assert!(
            seen.contains(&json!(c.as_str())),
            "{c:?} missing in {seen:?}"
        );
    }
}

/// AC-3: duplicate, out-of-order and unreported heads are judged by the PR's commit order,
/// not by comment ids or arrival; reversing every list gives a byte-identical table.
#[test]
fn duplicate_out_of_order_and_unreported_heads() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let out = run(&p, request(lanes(), decisions()), remote(false));

    let ev = &row(&out, "agent-3", 13)["events"];
    assert_eq!(
        ev["duplicate"],
        json!([{"comment_id": 1303, "head": sha('a'), "first_reported_by": 1302}])
    );
    assert_eq!(
        ev["out_of_order"],
        json!([{"comment_id": 1305, "head": sha('a'), "newer_head": sha('b'), "newer_reported_by": 1304}])
    );
    assert_eq!(ev["unreported_head"], Value::Null);
    assert_eq!(
        ev["ignored_shas"],
        json!(1),
        "the base SHA is outside the PR's commits"
    );

    let ev = &row(&out, "agent-7", 17)["events"];
    assert_eq!(
        ev["unreported_head"],
        json!({"head": sha('f'), "last_reported_head": sha('e')})
    );
    // An AUDIT line naming the head counts as a report of it.
    assert_eq!(
        row(&out, "agent-6", 15)["events"]["unreported_head"],
        Value::Null
    );

    let reversed = run(&p, request(lanes(), decisions()), remote(true));
    assert_eq!(
        serde_json::to_vec(&reversed).unwrap(),
        serde_json::to_vec(&out).unwrap(),
        "input order changes nothing"
    );
}

/// The page limit: a list that links 50 pages is read whole; a 51st page is refused.
#[test]
fn page_limit_at_its_edge() {
    let tmp = tempfile::tempdir().unwrap();
    let list = format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100");
    let paged = |pages: usize| -> Vec<Exchange> {
        let mut out = Vec::new();
        for k in 1..=pages {
            let url = if k == 1 {
                list.clone()
            } else {
                format!("{list}&page={k}")
            };
            let mut resp = common::json_response(200, json!([]));
            if k < pages {
                resp = resp.with_header("link", format!("<{list}&page={}>; rel=\"next\"", k + 1));
            }
            out.push(common::exchange("GET", &url, resp));
        }
        out
    };
    let req = || -> LaneStatusRequest {
        serde_json::from_value(json!({
            "schema_version": 1, "owner": "acme", "repo": "widgets",
            "base_branch": "v2026.11.0", "issues": [1],
        }))
        .unwrap()
    };
    let mut ok = paged(50);
    ok.push(get(&format!("{API}/issues/1"), issue(1, "")));
    ok.push(get(
        &format!("{API}/issues/1/comments?per_page=100"),
        json!([]),
    ));
    let out = run(tmp.path(), req(), ok);
    assert_eq!(out["rows"], json!([]));

    let t = RecordedTransport::new(paged(51));
    let err = status_cli(tmp.path(), req(), &t, &FakeClock(Cell::new(NOW))).unwrap_err();
    assert_eq!(err, "lane status page limit exceeded (50)");
    assert_eq!(t.sent.borrow().len(), 50, "the 51st page is not requested");
}

/// Limits: 1-50 inspected Issues, at most 50 lanes; N passes, N+1 is refused before any read.
#[test]
fn request_limits_at_their_edges() {
    let tmp = tempfile::tempdir().unwrap();
    let p = tmp.path();
    let req = |issues: Vec<i64>, lanes: usize| -> LaneStatusRequest {
        let lanes: Vec<Value> = (0..lanes)
            .map(|i| json!({"session": format!("s{i}")}))
            .collect();
        serde_json::from_value(json!({
            "schema_version": 1, "owner": "acme", "repo": "widgets",
            "base_branch": "v2026.11.0", "issues": issues, "lanes": lanes,
        }))
        .unwrap()
    };
    let refused = |r: LaneStatusRequest| {
        let t = RecordedTransport::new(vec![]);
        let err = status_cli(p, r, &t, &FakeClock(Cell::new(NOW))).unwrap_err();
        assert!(t.sent.borrow().is_empty(), "refused before any read");
        err
    };
    assert_eq!(
        refused(req((1..=51).collect(), 0)),
        "lane status issues must be 1 to 50 Issue numbers"
    );
    assert_eq!(
        refused(req(vec![], 0)),
        "lane status issues must be 1 to 50 Issue numbers"
    );
    assert_eq!(
        refused(req(vec![1], 51)),
        "lane status lanes must be at most 50"
    );
    // Caller evidence without a source or a time is refused.
    for lanes in [
        json!([{"session": "s", "pane": {"state": "working", "source": " ", "observed_at": "t"}}]),
        json!([{"session": "s", "pane": {"state": "working", "source": "hc", "observed_at": ""}}]),
        json!([{"session": ""}]),
    ] {
        let r: LaneStatusRequest = serde_json::from_value(json!({
            "schema_version": 1, "owner": "acme", "repo": "widgets",
            "base_branch": "v2026.11.0", "issues": [1], "lanes": lanes,
        }))
        .unwrap();
        assert_eq!(
            refused(r),
            "lane status evidence must be given with a session, a source and a time"
        );
    }
    let r: LaneStatusRequest = serde_json::from_value(json!({
        "schema_version": 1, "owner": "acme", "repo": "widgets", "base_branch": "v2026.11.0",
        "issues": [1], "pending_decisions": [{"issue": 1, "source": "", "recorded_at": "t"}],
    }))
    .unwrap();
    assert_eq!(
        refused(r),
        "lane status evidence must be given with a session, a source and a time"
    );
    // N = 50 Issues and 50 lanes pass validation: the first GitHub read is attempted.
    let t = RecordedTransport::new(vec![]);
    let err = status_cli(
        p,
        req((1..=50).collect(), 50),
        &t,
        &FakeClock(Cell::new(NOW)),
    );
    assert!(err.is_err());
    assert_eq!(
        t.sent.borrow().len(),
        1,
        "validation passed; the PR list read was tried"
    );
}

/// #2403 item 3: one card handed from agent-a (epoch 1) to agent-b (epoch 2) whose PR #121
/// carries both claim tokens, audit reservations on a stale and on the current head, and a
/// verdict posted on the PR (not the Issue).
fn audit_remote(reverse: bool) -> Vec<Exchange> {
    let (h1, hm, h2) = (sha('1'), sha('5'), sha('2'));
    let mut issue_comments = vec![
        claim(2101, "agent-a", 1),
        note(
            2102,
            "moai-claim-handoff: comment=2101 to=agent-b by=leader",
        ),
        claim(2103, "agent-b", 2),
        // The Issue also carries a verdict on the same head; the PR's verdict wins.
        note(
            2104,
            &format!(
                "AUDIT #21 (k) head {} base {} — verdict: PASS",
                sha('2'),
                sha('0')
            ),
        ),
    ];
    let mut pr_comments = vec![
        // Stale: names the first head.
        note(3001, &format!("moai-audit-claim: lane=lane-x head={h1}")),
        // Current head: the lowest id (3002) wins over 3003.
        note(3002, &format!("moai-audit-claim: lane=lane-y head={h2}")),
        note(3003, &format!("moai-audit-claim: lane=lane-z head={h2}")),
        // Not reservations: 39 and 41 hex characters, and a reservation quoted mid-line.
        note(
            3004,
            &format!("moai-audit-claim: lane=lane-q head={}", &h2[..39]),
        ),
        note(3005, &format!("moai-audit-claim: lane=lane-r head={h2}0")),
        note(
            3006,
            &format!("see `moai-audit-claim: lane=lane-s head={h2}`"),
        ),
        // Quoted mid-line without a code span, and a reservation with a trailing word.
        note(
            3008,
            &format!("quoted: moai-audit-claim: lane=lane-t head={h2}"),
        ),
        note(
            3009,
            &format!("moai-audit-claim: lane=lane-u head={h2} extra"),
        ),
        // Third line of a comment: a reservation. Uppercase 40-hex head: not a reservation.
        note(
            3010,
            &format!("감사 시작합니다.\n\nmoai-audit-claim: lane=lane-v head={h2}"),
        ),
        note(
            3011,
            &format!("moai-audit-claim: lane=lane-w head={}", "A".repeat(40)),
        ),
        note(
            3007,
            &format!("AUDIT #121 (k) head {h2} base {} — verdict: FAIL", sha('0')),
        ),
    ];
    let commits = vec![
        json!({"sha": h1, "commit": {"message": "feat: x\n\nMoai-Claim: 21#2101\n"}}),
        json!({"sha": hm, "commit": {"message": "Merge branch 'v2026.11.0' into issue/21-x"}}),
        // 2101 repeats (listed once), 2102 is the handoff line (not a claim: unresolved),
        // 2199 names no comment (unresolved), 99#… is another Issue (ignored).
        json!({"sha": h2, "commit": {"message": "fix: y\n\nMoai-Claim: 21#2101\nMoai-Claim: 21#2103\nMoai-Claim: 21#2102\nMoai-Claim: 21#2199\nMoai-Claim: 99#9901\n"}}),
    ];
    if reverse {
        issue_comments.reverse();
        pr_comments.reverse();
    }
    vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([{"number": 121, "head": {"sha": h2, "ref": "issue/21-x"}, "body": "Refs #21"}]),
        ),
        get(&format!("{API}/issues/21"), issue(21, "Blocked-By: 없음")),
        get(
            &format!("{API}/issues/21/comments?per_page=100"),
            Value::Array(issue_comments),
        ),
        get(
            &format!("{API}/pulls/121/commits?per_page=100"),
            Value::Array(commits),
        ),
        get(
            &format!("{API}/issues/121/comments?per_page=100"),
            Value::Array(pr_comments),
        ),
    ]
}

fn audit_request() -> LaneStatusRequest {
    serde_json::from_value(json!({
        "schema_version": 1, "owner": "acme", "repo": "widgets", "base_branch": "v2026.11.0",
        "issues": [21], "lanes": [], "pending_decisions": [],
    }))
    .unwrap()
}

#[test]
fn audit_columns_list_reservations_verdict_and_author_sessions() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let (h1, h2) = (sha('1'), sha('2'));
    for reverse in [false, true] {
        let out = run(&p, audit_request(), audit_remote(reverse));
        let a = &row(&out, "agent-b", 21)["audit"];
        assert_eq!(a["head"], json!(h2), "{out}");
        assert_eq!(
            a["reservations"],
            json!([
                {"lane": "lane-x", "comment_id": 3001, "head": h1, "current": false, "state": "stale"},
                {"lane": "lane-y", "comment_id": 3002, "head": h2, "current": true, "state": "current"},
                {"lane": "lane-z", "comment_id": 3003, "head": h2, "current": true, "state": "current"},
                // A reservation on the third line of a comment still counts.
                {"lane": "lane-v", "comment_id": 3010, "head": h2, "current": true, "state": "current"},
            ]),
            "reverse={reverse}"
        );
        assert_eq!(
            a["winner_on_head"],
            json!({"lane": "lane-y", "comment_id": 3002})
        );
        assert_eq!(
            a["verdict_on_head"],
            json!({"verdict": "FAIL", "comment_id": 3007, "source": "pr"})
        );
        // Both epochs' sessions wrote this head; neither may audit it. A token for another
        // Issue is ignored; one whose comment is not a claim on this Issue is unresolved.
        assert_eq!(
            a["author_sessions"],
            json!([
                {"session": "agent-a", "claim_comment_id": 2101},
                {"session": "agent-b", "claim_comment_id": 2103},
            ])
        );
        assert_eq!(a["unresolved_claim_ids"], json!([2102, 2199]));
    }
}

#[test]
fn audit_columns_are_null_without_a_pr_and_empty_without_records() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let out = run(&p, request(lanes(), decisions()), remote(false));
    // #11 has no PR: no audit object.
    assert_eq!(row(&out, "agent-1", 11)["audit"], Value::Null);
    // #15's PR has no reservation and no claim token; its PASS sits on the Issue.
    let a = &row(&out, "agent-6", 15)["audit"];
    assert_eq!(a["reservations"], json!([]));
    assert_eq!(a["winner_on_head"], Value::Null);
    assert_eq!(a["author_sessions"], json!([]));
    assert_eq!(a["unresolved_claim_ids"], json!([]));
    assert_eq!(a["verdict_on_head"]["source"], json!("issue"));
    assert_eq!(a["verdict_on_head"]["verdict"], json!("PASS"));
}

// ---- #2499: winner_on_head follows the leader handoff like audit-claim; lane names
// normalise by the audit_claim rule ----

fn audit_remote_with(
    issue_comments: Vec<Value>,
    commit_message: &str,
    pr_comments: Vec<Value>,
) -> Vec<Exchange> {
    let h2 = sha('2');
    vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([{"number": 121, "head": {"sha": h2, "ref": "issue/21-x"}, "body": "Refs #21"}]),
        ),
        get(&format!("{API}/issues/21"), issue(21, "Blocked-By: 없음")),
        get(
            &format!("{API}/issues/21/comments?per_page=100"),
            Value::Array(issue_comments),
        ),
        get(
            &format!("{API}/pulls/121/commits?per_page=100"),
            json!([{"sha": h2, "commit": {"message": commit_message}}]),
        ),
        get(
            &format!("{API}/issues/121/comments?per_page=100"),
            Value::Array(pr_comments),
        ),
    ]
}

fn as_comments(list: &[Value]) -> Vec<moai_cli::lane_status::Comment> {
    list.iter()
        .map(|c| moai_cli::lane_status::Comment {
            id: c["id"].as_i64().unwrap(),
            created_at: c["created_at"].as_str().unwrap_or_default().to_string(),
            body: c["body"].as_str().unwrap().to_string(),
        })
        .collect()
}

#[test]
fn winner_on_head_follows_the_last_leader_handoff_like_audit_claim() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let h2 = sha('2');
    let pr_comments = vec![
        note(4001, &format!("moai-audit-claim: lane=agent-9 head={h2}")),
        note(4002, &format!("moai-audit-claim: lane=agent-8 head={h2}")),
        note(
            4003,
            "moai-claim-handoff: comment=4001 to=agent-27 by=leader",
        ),
        // Not by the leader: ignored.
        note(
            4004,
            "moai-claim-handoff: comment=4001 to=agent-30 by=agent-9",
        ),
        // Names the losing reservation: ignored.
        note(
            4005,
            "moai-claim-handoff: comment=4002 to=agent-31 by=leader",
        ),
        // The last leader line on the winning reservation counts.
        note(
            4006,
            "moai-claim-handoff: comment=4001 to=agent-26 by=leader",
        ),
    ];
    let expected = moai_cli::audit_claim::holder_of(&as_comments(&pr_comments), &h2).unwrap();
    assert_eq!(
        (expected.lane.as_str(), expected.comment_id),
        ("agent-26", 4006)
    );
    for reverse in [false, true] {
        let mut prc = pr_comments.clone();
        if reverse {
            prc.reverse();
        }
        let issue_comments = vec![claim(2101, "agent-b", 1)];
        let out = run(
            &p,
            audit_request(),
            audit_remote_with(issue_comments, "feat: x\n\nMoai-Claim: 21#2101\n", prc),
        );
        let a = &row(&out, "agent-b", 21)["audit"];
        assert_eq!(
            a["winner_on_head"],
            json!({"lane": expected.lane, "comment_id": expected.comment_id}),
            "reverse={reverse} {out}"
        );
        assert_eq!(a["winner_via"], json!("handoff"));
        assert_eq!(a["winner_reservation_id"], json!(4001));
    }
}

#[test]
fn winner_is_author_uses_the_lane_name_rule() {
    let tmp = tempfile::tempdir().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let p = project(&root);
    let h2 = sha('2');
    // (claim session, reservation lane, expected winner_is_author)
    for (session, lane, author) in [
        ("pr-1", "pr-1-hB", true),
        ("agent-22-2293", "agent-22", true),
        ("agent-2", "agent-22", false),
        // Outside the rule: byte-exact only.
        ("codex-t47", "codex-t47-hB", false),
        ("codex-t47", "codex-t47", true),
    ] {
        let out = run(
            &p,
            audit_request(),
            audit_remote_with(
                vec![claim(2101, session, 1)],
                "feat: x\n\nMoai-Claim: 21#2101\n",
                vec![note(
                    4001,
                    &format!("moai-audit-claim: lane={lane} head={h2}"),
                )],
            ),
        );
        let a = &row(&out, session, 21)["audit"];
        assert_eq!(a["winner_on_head"]["lane"], json!(lane), "{out}");
        assert_eq!(a["winner_via"], json!("reservation"));
        assert_eq!(a["winner_is_author"], json!(author), "{session} vs {lane}");
        assert_eq!(a["lane_name_rule"], json!(moai_cli::audit_claim::NAME_RULE));
    }
}

// S6 packets deliberately share an authenticated GitHub account: sessions, not
// account names, identify the two lanes.
fn s6_message() -> moai_cli::message::Message {
    serde_json::from_value(json!({
        "schema_version": 1, "report_id": "s6-report", "project_id": "p1",
        "task": {"owner": "acme", "repo": "widgets", "repository_id": 1,
            "issue_id": 2, "issue_node_id": "I_2", "issue_number": 21,
            "task_id": "T1", "revision": "1"},
        "sender": {"actor_id": 11, "session_id": "sender"},
        "receiver": {"actor_id": 11, "session_id": "receiver"},
        "owner_epoch": 1, "leader_epoch": 1, "event": "report",
        "candidate_digest": moai_cli::digest(b"candidate"),
        "created_at": "2099-01-01T00:00:00Z",
        "approval_refs": ["decision:1"], "hold_refs": [],
        "payload": {"private_text": "never expose this payload"}
    }))
    .unwrap()
}

fn s6_packet(id: u64, mut m: moai_cli::message::Message, ack: bool) -> Value {
    #[derive(serde::Serialize)]
    struct Wire<'a> {
        schema_version: u32,
        kind: &'a str,
        message: &'a moai_cli::message::Message,
        payload_digest: String,
    }
    m.payload.sort_all_objects();
    let wire = Wire {
        schema_version: 1,
        kind: if ack { "ack" } else { "record" },
        payload_digest: moai_cli::digest(&serde_json::to_vec(&m.payload).unwrap()),
        message: &m,
    };
    let marker = moai_cli::forge::outbox::op_marker(
        if ack { "message-ack" } else { "message-record" },
        &m.project_id,
        &m.report_id,
        1,
    );
    json!({"id": id, "user": {"id": 11}, "created_at": "2026-11-11T00:00:00Z",
        "body": format!("{marker}\n{}\n<!-- moai:message:end -->",
            serde_json::to_string(&wire).unwrap())})
}

fn s6_run(p: &Path, comments: Vec<Value>, offset: i64, packets: bool) -> Value {
    let mut exchanges = vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([]),
        ),
        get(
            &format!("{API}/issues/21"),
            json!({"id": 2, "node_id": "I_2", "number": 21, "state": "open", "body": ""}),
        ),
        get(
            &format!("{API}/issues/21/comments?per_page=100"),
            json!(comments),
        ),
    ];
    if packets {
        exchanges.push(get(API, json!({"id": 1})));
    }
    let t = RecordedTransport::new(exchanges);
    let mut req = audit_request();
    req.lanes = serde_json::from_value(json!([{"session": "named-only"}])).unwrap();
    let out = status_cli(p, req, &t, &FakeClock(Cell::new(1_794_355_200 + offset))).unwrap();
    assert!(
        out["communication"].is_object(),
        "missing S6 communication: {out}"
    );
    assert_eq!(
        t.remaining(),
        0,
        "reuse issue and comments; only repository identity is extra"
    );
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
    assert_eq!(out["writes"], "none");
    out
}

fn s6_lane<'a>(out: &'a Value, session: &str) -> &'a Value {
    out["communication"]["lanes"]
        .as_array()
        .unwrap()
        .iter()
        .find(|lane| lane["session"] == session)
        .unwrap()
}

#[test]
fn s6_server_record_age_has_literal_299_300_301_boundaries() {
    let tmp = tempfile::tempdir().unwrap();
    for (age, alert) in [
        (299, "waiting"),
        (300, "unacknowledged"),
        (301, "unacknowledged"),
    ] {
        let out = s6_run(
            tmp.path(),
            vec![s6_packet(101, s6_message(), false)],
            age,
            true,
        );
        assert_eq!(out["communication"]["available"], true);
        assert_eq!(out["communication"]["observed_at"], 1_794_355_200_i64 + age);
        let lane = s6_lane(&out, "sender");
        let report = &lane["sent_reports"][0];
        assert_eq!(
            report["elapsed_secs"], age,
            "message timestamp must not fake server age"
        );
        assert_eq!(report["ack_alert"], alert);
        assert_eq!(report["received_ack"], false);
        assert_eq!(report["record_comment_id"], 101);
        assert_eq!(report["ack_comment_id"], Value::Null);
        assert_eq!(
            lane["unacknowledged_report_ids"],
            if age >= 300 {
                json!(["s6-report"])
            } else {
                json!([])
            }
        );
    }
}

#[test]
fn s6_explicit_receipt_deduplicates_and_discovers_distinct_sessions() {
    let tmp = tempfile::tempdir().unwrap();
    let record = s6_packet(101, s6_message(), false);
    let ack = s6_packet(102, s6_message(), true);
    let out = s6_run(
        tmp.path(),
        vec![ack.clone(), record.clone(), ack, record],
        301,
        true,
    );
    assert_eq!(out["communication"]["available"], true);
    let sessions: Vec<_> = out["communication"]["lanes"]
        .as_array()
        .unwrap()
        .iter()
        .map(|l| l["session"].clone())
        .collect();
    assert_eq!(
        sessions,
        json!(["named-only", "receiver", "sender"])
            .as_array()
            .unwrap()
            .clone()
    );
    let sender = s6_lane(&out, "sender");
    assert_eq!(sender["sent_reports"].as_array().unwrap().len(), 1);
    assert_eq!(sender["sent_reports"][0]["received_ack"], true);
    assert_eq!(sender["sent_reports"][0]["ack_alert"], "received");
    assert_eq!(sender["sent_reports"][0]["ack_comment_id"], 102);
    assert_eq!(
        sender["received_acks"],
        json!([{"report_id": "s6-report", "issue": 21, "ack_comment_id": 102}])
    );
    assert_eq!(sender["unacknowledged_report_ids"], json!([]));
    assert_eq!(s6_lane(&out, "receiver")["sent_reports"], json!([]));
    assert!(!serde_json::to_string(&out["communication"])
        .unwrap()
        .contains("never expose"));
    assert!(!serde_json::to_string(&out["communication"])
        .unwrap()
        .contains("private_text"));
}

#[test]
fn s6_invalid_packets_fail_closed_without_changing_old_observations() {
    let tmp = tempfile::tempdir().unwrap();
    let record = s6_packet(101, s6_message(), false);
    let ack = s6_packet(102, s6_message(), true);
    let mut actor = ack.clone();
    actor["user"]["id"] = json!(99);
    let mut digest = ack.clone();
    digest["body"] = json!(digest["body"].as_str().unwrap().replace(
        &moai_cli::digest(&serde_json::to_vec(&s6_message().payload).unwrap()),
        &"0".repeat(64)
    ));
    let mut foreign = s6_message();
    foreign.task.issue_id = 999;
    let mut conflict = s6_message();
    conflict.receiver.session_id = "someone-else".into();
    let mut malformed = record.clone();
    malformed["body"] = json!(format!(
        "{}\nnot-json\n<!-- moai:message:end -->",
        record["body"].as_str().unwrap().lines().next().unwrap()
    ));
    for (case, comments) in [
        ("orphan", vec![ack]),
        ("actor", vec![record.clone(), actor]),
        ("digest", vec![record.clone(), digest]),
        ("context", vec![s6_packet(101, foreign, false)]),
        (
            "immutable-id",
            vec![record.clone(), s6_packet(103, conflict, false)],
        ),
        ("malformed", vec![record, malformed]),
    ] {
        let out = s6_run(tmp.path(), comments, 301, true);
        assert_eq!(out["communication"]["available"], false, "{case}: {out}");
        assert!(
            !out["communication"]["errors"]
                .as_array()
                .unwrap()
                .is_empty(),
            "{case}"
        );
        assert_eq!(out["rows"], json!([]), "{case}");
        assert_eq!(out["ready"], json!([21]), "{case}");
        assert!(!serde_json::to_string(&out["communication"]["errors"])
            .unwrap()
            .contains("private_text"));
    }
}

#[test]
fn s6_future_or_malformed_server_time_is_not_clamped_to_waiting() {
    let tmp = tempfile::tempdir().unwrap();
    for timestamp in [
        "2099-01-01T00:00:00Z",
        "not-a-time",
        "999999999999999999999",
    ] {
        let mut packet = s6_packet(101, s6_message(), false);
        packet["created_at"] = json!(timestamp);
        let out = s6_run(tmp.path(), vec![packet], 301, true);
        assert_eq!(out["communication"]["available"], false, "{timestamp}");
    }
}

#[test]
fn s6_no_packets_preserves_get_sequence_and_project_files() {
    let tmp = tempfile::tempdir().unwrap();
    let sentinel = tmp.path().join("user-owned");
    fs::write(&sentinel, b"unchanged").unwrap();
    let out = s6_run(tmp.path(), vec![], 300, false);
    assert_eq!(out["communication"]["available"], true);
    assert_eq!(out["communication"]["errors"], json!([]));
    assert_eq!(s6_lane(&out, "named-only")["sent_reports"], json!([]));
    assert_eq!(out["rows"], json!([]));
    assert_eq!(fs::read(&sentinel).unwrap(), b"unchanged");
    assert_eq!(fs::read_dir(tmp.path()).unwrap().count(), 1);
}

fn s6_dispatch() -> moai_cli::message::Message {
    let mut m = s6_message();
    m.event = "dispatch".into();
    m.payload = json!({"dispatch_id": "job1", "revision": 1, "previous": null,
        "scope": ["src/example.rs"], "completion": ["tests pass"], "audit": null});
    m
}

#[test]
fn s6_dispatch_receipt_is_not_business_acceptance() {
    let tmp = tempfile::tempdir().unwrap();
    let m = s6_dispatch();
    let record = s6_packet(101, m.clone(), false);
    let receipt = s6_packet(102, m.clone(), true);
    for received in [false, true] {
        let comments = if received {
            vec![record.clone(), receipt.clone()]
        } else {
            vec![record.clone()]
        };
        let out = s6_run(tmp.path(), comments, 301, true);
        assert_eq!(out["communication"]["available"], true);
        let dispatch = &s6_lane(&out, "sender")["dispatches"][0];
        assert_eq!(dispatch["received_ack"], received);
        assert_eq!(dispatch["disposition"], Value::Null);
        assert_eq!(dispatch["claim_authorized"], false);
        assert_eq!(dispatch["ownership_transferred"], false);
        assert_eq!(dispatch["issue"], 21);
        assert_eq!(dispatch["from"], "sender");
        assert_eq!(dispatch["to"], "receiver");
    }
    let mut decision = m.clone();
    decision.report_id = "s6-decision".into();
    decision.event = "dispatch-decision".into();
    std::mem::swap(&mut decision.sender, &mut decision.receiver);
    let mut dispatch_payload = m.payload.clone();
    dispatch_payload.sort_all_objects();
    decision.payload = json!({"dispatch_report_id": m.report_id,
        "dispatch_payload_digest": moai_cli::digest(&serde_json::to_vec(&dispatch_payload).unwrap()),
        "disposition": "accepted", "reason": "explicit receiver decision"});
    let out = s6_run(
        tmp.path(),
        vec![record, receipt, s6_packet(103, decision, false)],
        301,
        true,
    );
    assert_eq!(out["communication"]["available"], true);
    let dispatch = &s6_lane(&out, "sender")["dispatches"][0];
    assert_eq!(dispatch["received_ack"], true);
    assert_eq!(dispatch["disposition"], "accepted");
    assert_eq!(dispatch["ownership_transferred"], false);
    assert_eq!(
        s6_lane(&out, "receiver")["sent_reports"][0]["event"],
        "dispatch-decision"
    );
}

#[test]
fn s6_dispatch_new_revision_supersedes_old_receiver() {
    let tmp = tempfile::tempdir().unwrap();
    let m = s6_dispatch();
    let mut payload = m.payload.clone();
    payload.sort_all_objects();
    let mut next = m.clone();
    next.report_id = "s6-revision-two".into();
    next.task.revision = "2".into();
    next.receiver.session_id = "next-receiver".into();
    next.owner_epoch = 2;
    next.payload["revision"] = json!(2);
    next.payload["previous"] = json!({"report_id": m.report_id,
        "payload_digest": moai_cli::digest(&serde_json::to_vec(&payload).unwrap()),
        "comment_id": 101});
    let out = s6_run(
        tmp.path(),
        vec![
            s6_packet(103, next, false),
            s6_packet(102, m.clone(), true),
            s6_packet(101, m, false),
        ],
        301,
        true,
    );
    assert_eq!(out["communication"]["available"], true);
    let dispatches = s6_lane(&out, "sender")["dispatches"].as_array().unwrap();
    assert_eq!(dispatches.len(), 2);
    let old = dispatches.iter().find(|d| d["superseded"] == true).unwrap();
    assert_eq!(old["current_revision"], 2);
    assert_eq!(old["disposition"], Value::Null);
    assert_eq!(old["claim_authorized"], false);
    assert_eq!(old["ownership_transferred"], false);
}

#[test]
fn s6_review_malformed_marker_is_not_ordinary_prose() {
    let tmp = tempfile::tempdir().unwrap();
    for (from, to) in [(":1 -->", ":not-a-number -->"), ("<!-- ", "<!--  ")] {
        let mut packet = s6_packet(101, s6_message(), false);
        packet["body"] = json!(packet["body"].as_str().unwrap().replacen(from, to, 1));
        let out = s6_run(tmp.path(), vec![packet], 301, true);
        assert_eq!(out["communication"]["available"], false);
    }
}

#[test]
fn s6_review_decision_cannot_reference_a_foreign_task() {
    let tmp = tempfile::tempdir().unwrap();
    let mut dispatch = s6_dispatch();
    dispatch.payload.sort_all_objects();
    let mut decision = dispatch.clone();
    decision.report_id = "s6-decision".into();
    decision.event = "dispatch-decision".into();
    decision.task.task_id = "T2".into();
    std::mem::swap(&mut decision.sender, &mut decision.receiver);
    decision.payload = json!({"dispatch_report_id": dispatch.report_id,
        "dispatch_payload_digest": moai_cli::digest(&serde_json::to_vec(&dispatch.payload).unwrap()),
        "disposition": "accepted", "reason": "wrong task"});
    let out = s6_run(
        tmp.path(),
        vec![
            s6_packet(101, dispatch.clone(), false),
            s6_packet(102, dispatch, true),
            s6_packet(103, decision, false),
        ],
        301,
        true,
    );
    assert_eq!(out["communication"]["available"], false);
}

#[test]
fn s6_review_incomplete_pagination_cannot_become_a_complete_snapshot() {
    let tmp = tempfile::tempdir().unwrap();
    let comments = common::json_response(200, json!([s6_packet(101, s6_dispatch(), false)]))
        .with_header(
            "link",
            format!("{API}/issues/21/comments?page=2; rel=\"next\""),
        );
    let t = RecordedTransport::new(vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([]),
        ),
        get(
            &format!("{API}/issues/21"),
            json!({"id": 2, "node_id": "I_2", "number": 21, "state": "open", "body": ""}),
        ),
        common::exchange(
            "GET",
            &format!("{API}/issues/21/comments?per_page=100"),
            comments,
        ),
        get(API, json!({"id": 1})),
    ]);
    let out = status_cli(
        tmp.path(),
        audit_request(),
        &t,
        &FakeClock(Cell::new(1_794_355_501)),
    )
    .unwrap();
    assert_eq!(out["communication"]["available"], false);
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
}

#[test]
fn s6_review_identity_is_immutable_across_inspected_issues() {
    let tmp = tempfile::tempdir().unwrap();
    let mut other = s6_message();
    other.task.issue_id = 3;
    other.task.issue_node_id = "I_3".into();
    other.task.issue_number = 22;
    let t = RecordedTransport::new(vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([]),
        ),
        get(
            &format!("{API}/issues/21"),
            json!({"id": 2, "node_id": "I_2", "number": 21, "state": "open", "body": ""}),
        ),
        get(
            &format!("{API}/issues/21/comments?per_page=100"),
            json!([s6_packet(101, s6_message(), false)]),
        ),
        get(
            &format!("{API}/issues/22"),
            json!({"id": 3, "node_id": "I_3", "number": 22, "state": "open", "body": ""}),
        ),
        get(
            &format!("{API}/issues/22/comments?per_page=100"),
            json!([s6_packet(201, other, false)]),
        ),
        get(API, json!({"id": 1})),
    ]);
    let mut req = audit_request();
    req.issues.push(22);
    let out = status_cli(tmp.path(), req, &t, &FakeClock(Cell::new(1_794_355_501))).unwrap();
    assert_eq!(out["communication"]["available"], false);
    assert_eq!(t.remaining(), 0);
    assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
}

fn b1_numeric_comment_pages(
    project: &Path,
    first: Vec<Value>,
    second: Vec<Value>,
    source_repo: u64,
    expected_repo: u64,
) -> (Value, RecordedTransport) {
    let next = format!(
        "https://api.github.com/repositories/{source_repo}/issues/21/comments?per_page=100&page=2"
    );
    let first_response = common::json_response(200, json!(first))
        .with_header("link", format!("<{next}>; rel=\"next\""));
    let transport = RecordedTransport::new(vec![
        get(
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            json!([]),
        ),
        get(
            &format!("{API}/issues/21"),
            json!({"id": 2, "node_id": "I_2", "number": 21, "state": "open", "body": ""}),
        ),
        common::exchange(
            "GET",
            &format!("{API}/issues/21/comments?per_page=100"),
            first_response,
        ),
        get(&next, json!(second)),
        get(API, json!({"id": expected_repo})),
    ]);
    let out = status_cli(
        project,
        audit_request(),
        &transport,
        &FakeClock(Cell::new(1_794_355_501)),
    )
    .unwrap();
    (out, transport)
}

#[test]
fn b1_page_101_claim_is_not_silently_treated_as_unassigned() {
    let tmp = tempfile::tempdir().unwrap();
    let first = (1..=100).map(|id| note(id, "ordinary status")).collect();
    let (out, transport) = b1_numeric_comment_pages(
        tmp.path(),
        first,
        vec![claim(2001, "after-page", 1)],
        101,
        101,
    );
    assert_eq!(
        out["ready"],
        json!([]),
        "page-two ownership must prevent assignment: {out}"
    );
    let owner = row(&out, "after-page", 21);
    assert_eq!(owner["ownership"]["claim_comment_id"], 2001);
    assert_eq!(owner["phase"], "claimed");
    assert_eq!(out["communication"]["available"], true);
    assert_eq!(transport.remaining(), 0);
    assert!(transport
        .sent
        .borrow()
        .iter()
        .all(|request| request.method == "GET"));
}

#[test]
fn b1_public_pr_pagination_preserves_page_two_pr_and_audit() {
    let tmp = tempfile::tempdir().unwrap();
    let head = sha('b');
    let first: Vec<Value> = (1..=100)
        .map(|n| {
            json!({"number":1000+n,"head":{"sha":sha('c'),"ref":"unrelated"},
            "body":"Refs #999"})
        })
        .collect();
    let next = "https://api.github.com/repositories/101/pulls?per_page=100&page=2";
    let transport = RecordedTransport::new(vec![
        common::exchange(
            "GET",
            &format!("{API}/pulls?state=open&base=v2026.11.0&per_page=100"),
            common::json_response(200, json!(first))
                .with_header("link", format!("<{next}>; rel=\"next\"")),
        ),
        get(
            next,
            json!([{"number":121,"head":{"sha":head,"ref":"issue/21-message"},
            "body":"Refs #21"}]),
        ),
        get(
            &format!("{API}/issues/21"),
            json!({"id":2,"node_id":"I_2","number":21,"state":"open","body":""}),
        ),
        get(
            &format!("{API}/issues/21/comments?per_page=100"),
            json!([claim(2001, "owner", 1)]),
        ),
        get(
            &format!("{API}/pulls/121/commits?per_page=100"),
            json!([{"sha":head,"commit":{"message":"Moai-Claim: 21#2001"}}]),
        ),
        get(
            &format!("{API}/issues/121/comments?per_page=100"),
            json!([
                note(
                    3001,
                    &format!("moai-audit-claim: lane=reviewer head={head}")
                ),
                note(
                    3002,
                    &format!(
                        "AUDIT #121 (B1) head {head} base {} — verdict: PASS",
                        sha('a')
                    )
                ),
            ]),
        ),
    ]);
    let out = status_cli(
        tmp.path(),
        audit_request(),
        &transport,
        &FakeClock(Cell::new(1_794_355_501)),
    )
    .unwrap();
    let entry = row(&out, "owner", 21);
    assert_eq!(entry["pr"], 121);
    assert_eq!(entry["phase"], "pr_open");
    assert_eq!(entry["audit"]["winner_on_head"]["lane"], "reviewer");
    assert_eq!(entry["audit"]["verdict_on_head"]["verdict"], "PASS");
    assert_eq!(transport.remaining(), 0);
    assert!(transport
        .sent
        .borrow()
        .iter()
        .all(|request| request.method == "GET"));
}

#[test]
fn b1_numeric_comment_alias_binds_ack_to_the_verified_repository() {
    let tmp = tempfile::tempdir().unwrap();
    for source_repo in [1, 2] {
        let mut first: Vec<_> = (1..=99).map(|id| note(id, "ordinary status")).collect();
        first.push(s6_packet(101, s6_message(), false));
        let (out, transport) = b1_numeric_comment_pages(
            tmp.path(),
            first,
            vec![s6_packet(102, s6_message(), true)],
            source_repo,
            1,
        );
        if source_repo == 1 {
            assert_eq!(out["communication"]["available"], true, "{out}");
            let sender = s6_lane(&out, "sender");
            assert_eq!(sender["sent_reports"][0]["received_ack"], true);
            assert_eq!(
                sender["received_acks"],
                json!([{"report_id":"s6-report","issue":21,"ack_comment_id":102}])
            );
        } else {
            assert_eq!(out["communication"]["available"], false, "{out}");
            assert!(out["communication"]["lanes"]
                .as_array()
                .unwrap()
                .iter()
                .all(
                    |lane| lane["sent_reports"] == json!([]) && lane["received_acks"] == json!([])
                ));
        }
        assert_eq!(transport.remaining(), 0);
        assert!(transport
            .sent
            .borrow()
            .iter()
            .all(|request| request.method == "GET"));
    }
}
