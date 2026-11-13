//! AC-1c: a real temporary repository, remote ledger/PR fixtures, GET-only transport.
mod common;
use common::CleanEnv;
use moai_cli::forge::{
    policy::Clock,
    transport::{Exchange, RecordedTransport},
};
use moai_cli::queue_query::{parse_order, query_cli, QueueQueryRequest};
use serde_json::{json, Value};

const API: &str = "https://api.github.com/repos/acme/widgets";
struct Fixed;
impl Clock for Fixed {
    fn now(&self) -> i64 {
        1_794_000_000
    }
    fn sleep(&self, _: u64) {}
}
fn get(tail: &str, body: Value) -> Exchange {
    common::exchange(
        "GET",
        &format!("{API}/{tail}"),
        common::json_response(200, body),
    )
}
fn req() -> QueueQueryRequest {
    serde_json::from_value(json!({"schema_version":1,"owner":"acme","repo":"widgets","base_branch":"integration","ledger_issue":1,"ledger_comment_id":5})).unwrap()
}
fn ledger(order: &[i64]) -> Exchange {
    let mut body = "queue\n| 순서 | PR | head |\n|---|---|---|\n".to_string();
    for (i, n) in order.iter().enumerate() {
        body.push_str(&format!("| {} | #{n} (#7) | ignored |\n", i + 1));
    }
    get(
        "issues/comments/5",
        json!({"id":5,"issue_url":format!("{API}/issues/1"),"body":body,"updated_at":"2026-11-10T00:00:00Z"}),
    )
}
fn pr(n: i64, head: &str) -> Value {
    json!({"number":n,"state":"open","draft":false,"body":"Refs #7","head":{"sha":head,"ref":"issue/7-x"},"base":{"sha":"0".repeat(40),"ref":"integration"},"mergeable_state":"clean"})
}
fn audit(n: i64, head: &str, base: &str, verdict: &str) -> Value {
    json!({"id":200,"body":format!("AUDIT #{n} (k) head {head} base {base} — verdict: {verdict}")})
}
fn check(id: i64, name: &str, head: &str, state: &str) -> Value {
    json!({"id":id,"name":name,"head_sha":head,"app":{"slug":"github-actions"},"status":"completed","conclusion":state})
}
fn row(n: i64, detail: Value, comments: Vec<Value>, checks: Vec<Value>) -> Vec<Exchange> {
    let head = detail["head"]["sha"].as_str().unwrap();
    vec![
        get(&format!("pulls/{n}"), detail.clone()),
        get(
            &format!("issues/{n}/comments?per_page=100"),
            json!(comments),
        ),
        get("issues/7", json!({"state":"open","body":""})),
        get(
            "issues/7/comments?per_page=100",
            json!([{"id":100,"body":"moai-claim: session=agent-20 epoch=1 lease_until=2030-01-01T00:00:00Z"}]),
        ),
        get(
            &format!("commits/{head}/check-runs?per_page=100&filter=all"),
            json!({"check_runs":checks}),
        ),
        get(&format!("pulls/{n}"), detail.clone()),
    ]
}
fn run(ex: Vec<Exchange>) -> (Value, RecordedTransport) {
    let tmp = tempfile::tempdir().unwrap();
    common::git::git(tmp.path(), &["init"]);
    let before = common::git::git(tmp.path(), &["status", "--porcelain"]);
    let mut ex = ex;
    ex.push(ex[0].clone());
    let t = RecordedTransport::new(ex);
    let out = query_cli(tmp.path(), req(), &t, &Fixed).unwrap();
    assert_eq!(
        before,
        common::git::git(tmp.path(), &["status", "--porcelain"])
    );
    assert!(
        !tmp.path().join(".moai-cli").exists(),
        "read-only query created a store"
    );
    assert_eq!(t.remaining(), 0);
    (out, t)
}

#[test]
fn ledger_order_owner_exact_head_and_predecessor_in_temporary_repo() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    let mut ex = vec![ledger(&[20, 10])];
    for n in [20, 10] {
        ex.extend(row(
            n,
            pr(n, &h),
            vec![audit(n, &h, &base, "PASS")],
            vec![
                check(1, "rc-ci", &h, "success"),
                check(2, "windows-build", &h, "success"),
            ],
        ));
    }
    let (out, _) = run(ex);
    assert_eq!(out["rows"][0]["position"], 1);
    assert_eq!(out["rows"][0]["pr"], 20);
    assert_eq!(out["rows"][0]["session"], "agent-20");
    assert_eq!(out["rows"][0]["head"], h);
    assert_eq!(out["rows"][0]["verdict"]["verdict"], "PASS");
    assert_eq!(out["rows"][0]["blocked_reasons"], json!([]));
    assert_eq!(
        out["rows"][1]["blocked_reasons"],
        json!(["waiting_predecessor"])
    );
    assert_eq!(out["rows"][1]["preceding_prs"], json!([20]));
    assert_eq!(out["writes"], "none");
}

#[test]
fn stale_conditional_wrong_pr_or_wrong_base_pass_never_counts() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    for note in [
        audit(20, &"b".repeat(40), &base, "PASS"),
        audit(99, &h, &base, "PASS"),
        audit(20, &h, &"c".repeat(40), "PASS"),
        audit(20, &h, &base, "PASS if CI passes"),
        json!({"id":201,"body":format!("quoted example\nAUDIT #20 (k) head {h} base {base} — verdict: PASS")}),
    ] {
        let mut p = pr(20, &h);
        p["mergeable_state"] = json!("dirty");
        let mut ex = vec![ledger(&[20])];
        ex.extend(row(20, p, vec![note], vec![]));
        let (out, _) = run(ex);
        let r = &out["rows"][0];
        assert!(r["verdict"].is_null());
        assert_eq!(r["ci"]["ci"]["state"], "not_run");
        assert_eq!(
            r["blocked_reasons"],
            json!(["audit_missing_or_stale", "ci_not_success", "merge_conflict"])
        );
    }
}

/// #2554 (user decision U10): on a PR `windows-build` does not run (it runs after merge).
/// A missing check run or one its `if` skipped (conclusion `skipped`, pre-audit FAIL
/// 6827770462) does not block; one that ran on the head and is not `success` does.
#[test]
fn windows_build_blocks_only_when_it_ran_and_did_not_succeed() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    for (windows, blocked) in [
        (None, false),
        (Some("skipped"), false),
        (Some("success"), false),
        (Some("failure"), true),
        (Some("cancelled"), true),
    ] {
        let mut checks = vec![check(10, "rc-ci", &h, "success")];
        if let Some(conclusion) = windows {
            checks.push(check(11, "windows-build", &h, conclusion));
        }
        let mut ex = vec![ledger(&[20])];
        ex.extend(row(
            20,
            pr(20, &h),
            vec![audit(20, &h, &base, "PASS")],
            checks,
        ));
        let (out, _) = run(ex);
        let reasons = out["rows"][0]["blocked_reasons"]
            .as_array()
            .unwrap()
            .clone();
        assert_eq!(
            reasons.contains(&json!("windows_build_not_success")),
            blocked,
            "windows-build {windows:?}: {reasons:?}"
        );
        assert!(!reasons.contains(&json!("ci_not_success")), "{reasons:?}");
    }
}

#[test]
fn newest_ci_attempt_wins_even_when_old_success_appears_first() {
    let h = "a".repeat(40);
    let mut checks = vec![
        check(10, "rc-ci", &h, "success"),
        check(11, "rc-ci", &h, "failure"),
        check(90, "windows-build", &"b".repeat(40), "success"),
    ];
    checks[1]["status"] = json!("in_progress");
    checks[1]["conclusion"] = Value::Null;
    let mut ex = vec![ledger(&[20])];
    ex.extend(row(
        20,
        pr(20, &h),
        vec![audit(20, &h, &"0".repeat(40), "FAIL")],
        checks,
    ));
    let (out, _) = run(ex);
    assert_eq!(out["rows"][0]["ci"]["ci"]["state"], "in_progress");
    assert_eq!(out["rows"][0]["ci"]["windows-build"]["state"], "not_run");
    assert_eq!(out["rows"][0]["verdict"]["verdict"], "FAIL");
}

#[test]
fn empty_queue_is_valid_but_ambiguous_or_malformed_order_is_not() {
    let (out, _) = run(vec![ledger(&[])]);
    assert_eq!(out["rows"], json!([]));
    for body in [
        "no table",
        "| 순서 | PR |\n|---|---|\n|2|#4|",
        "| 순서 | PR |\n|---|---|\n|1|#4|\n|2|#4|",
        "| 순서 | PR |\n|---|---|\n|1|#4|\n|2|bad|",
        "| 순서 | PR |\n|---|---|\n\n| 순서 | PR |\n|---|---|",
    ] {
        assert!(parse_order(body).is_err(), "{body}");
    }
    let table = |n: usize| {
        let mut body = "| 순서 | PR |\n|---|---|\n".to_string();
        for i in 1..=n {
            body.push_str(&format!("|{i}|#{i}|\n"));
        }
        body
    };
    assert_eq!(parse_order(&table(50)).unwrap().len(), 50);
    assert!(parse_order(&table(51)).is_err());
    let body = "```\n| 순서 | PR |\n|---|---|\n|1|#99|\n```\n| 순서 | PR |\n|---|---|\n|1|#4|";
    assert_eq!(parse_order(body).unwrap(), vec![4]);
}

#[test]
fn malformed_request_and_wrong_ledger_issue_do_not_read_prs() {
    let tmp = tempfile::tempdir().unwrap();
    let t = RecordedTransport::new(vec![]);
    let mut r = req();
    r.owner = "../other".into();
    assert!(query_cli(tmp.path(), r, &t, &Fixed).is_err());
    let t = RecordedTransport::new(vec![get(
        "issues/comments/5",
        json!({"id":5,"issue_url":format!("{API}/issues/2"),"body":""}),
    )]);
    assert!(query_cli(tmp.path(), req(), &t, &Fixed)
        .unwrap_err()
        .contains("does not belong"));
    assert_eq!(t.remaining(), 0);
}

#[test]
fn moving_head_refuses_a_mixed_snapshot() {
    let h = "a".repeat(40);
    let mut ex = vec![ledger(&[20])];
    ex.extend(row(20, pr(20, &h), vec![], vec![]));
    ex.last_mut().unwrap().response.body = Some(pr(20, &"b".repeat(40)));
    let tmp = tempfile::tempdir().unwrap();
    common::git::git(tmp.path(), &["init"]);
    let t = RecordedTransport::new(ex);
    assert!(query_cli(tmp.path(), req(), &t, &Fixed)
        .unwrap_err()
        .contains("changed during query"));
}

#[test]
fn paginated_checks_use_newest_attempt_and_comment_order_is_by_id() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    let mut ex = vec![ledger(&[20])];
    let mut pass = audit(20, &h, &base, "PASS");
    pass["id"] = json!(202);
    ex.extend(row(
        20,
        pr(20, &h),
        vec![pass, audit(20, &h, &base, "FAIL")],
        vec![check(1, "rc-ci", &h, "success")],
    ));
    let next = format!("{API}/commits/{h}/check-runs?per_page=100&filter=all&page=2");
    ex[5]
        .response
        .headers
        .insert("link".into(), format!("<{next}>; rel=\"next\""));
    ex.insert(6,common::exchange("GET",&next,common::json_response(200,json!({"check_runs":[check(3,"rc-ci",&h,"failure"),check(4,"windows-build",&h,"success")]}))));
    let (out, _) = run(ex);
    assert_eq!(out["rows"][0]["verdict"]["verdict"], "PASS");
    assert_eq!(out["rows"][0]["ci"]["ci"]["state"], "failure");
}

#[test]
fn expired_claim_blocker_and_closed_predecessor_remain_distinct() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    let mut first = pr(20, &h);
    first["state"] = json!("closed");
    let checks = vec![
        check(1, "rc-ci", &h, "success"),
        check(2, "windows-build", &h, "success"),
    ];
    let mut ex = vec![ledger(&[20, 10])];
    ex.extend(row(
        20,
        first,
        vec![audit(20, &h, &base, "PASS")],
        checks.clone(),
    ));
    let mut second = row(10, pr(10, &h), vec![audit(10, &h, &base, "PASS")], checks);
    second[2].response.body = Some(json!({"state":"open","body":"Blocked-By: #99 #98"}));
    second[3].response.body = Some(
        json!([{"id":100,"body":"moai-claim: session=agent-20 epoch=1 lease_until=2020-01-01T00:00:00Z"}]),
    );
    second.insert(4, get("issues/98", json!({"state":"closed"})));
    second.insert(5, get("issues/99", json!({"state":"open"})));
    ex.extend(second);
    let (out, _) = run(ex);
    assert_eq!(out["rows"][1]["session"], "agent-20");
    assert_eq!(out["rows"][1]["open_blockers"], json!([99]));
    assert_eq!(out["rows"][1]["preceding_prs"], json!([]));
    assert_eq!(
        out["rows"][1]["blocked_reasons"],
        json!(["waiting_prerequisite", "claim_expired_or_handoff"])
    );
}

#[test]
fn ledger_edit_and_pagination_cycle_are_errors_not_partial_success() {
    let tmp = tempfile::tempdir().unwrap();
    common::git::git(tmp.path(), &["init"]);
    let mut ex = vec![ledger(&[])];
    let mut changed = ledger(&[]);
    changed.response.body.as_mut().unwrap()["updated_at"] = json!("new");
    ex.push(changed);
    assert!(
        query_cli(tmp.path(), req(), &RecordedTransport::new(ex), &Fixed)
            .unwrap_err()
            .contains("ledger changed")
    );
    let h = "a".repeat(40);
    let mut ex = vec![ledger(&[20])];
    ex.extend(row(20, pr(20, &h), vec![], vec![]));
    ex[2].response.headers.insert(
        "link".into(),
        format!("<{API}/issues/20/comments?per_page=100>; rel=\"next\""),
    );
    assert!(
        query_cli(tmp.path(), req(), &RecordedTransport::new(ex), &Fixed)
            .unwrap_err()
            .contains("pagination")
    );
}

#[cfg(unix)]
#[test]
fn binary_routes_typed_request_and_envelope_using_fake_gh() {
    use std::process::{Command, Stdio};
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    std::fs::create_dir(&project).unwrap();
    common::git::git(&project, &["init"]);
    let fake = tmp.path().join("gh");
    // A standalone script for this fixture only; no live gh or inherited credentials.
    let body = ledger(&[]).response.body.unwrap();
    let script=format!("#!/bin/sh\ncase \"$*\" in\n*' -X GET '*) ;;\n*) exit 77 ;;\nesac\nprintf '%s\\n\\n%s\\n' 'HTTP/2 200' '{}'\n",body.to_string().replace('\'',"'\\''"));
    common::write_executable(&fake, &script);
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(["queue-query", project.to_str().unwrap()])
        .clean_env()
        .env("HOME", tmp.path())
        .env("PATH", "/usr/bin:/bin")
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("MOAI_FORGE_GH_BIN", fake)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, &serde_json::to_vec(&req_json()).unwrap());
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["operation"], "queue.query");
    assert_eq!(v["result"]["rows"], json!([]));
    assert!(!project.join(".moai-cli").exists());
}

#[cfg(unix)]
fn req_json() -> Value {
    json!({"schema_version":1,"owner":"acme","repo":"widgets","base_branch":"integration","ledger_issue":1,"ledger_comment_id":5})
}

#[test]
fn final_verdict_token_beats_key_text() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    for (key, verdict, reasons) in [
        ("verdict:PASS", "FAIL", json!(["audit_failed"])),
        ("verdict:FAIL", "PASS", json!([])),
    ] {
        let note = json!({"id":201,"body":format!("AUDIT #20 ({key}) head {h} base {base} — verdict: {verdict}")});
        let mut ex = vec![ledger(&[20])];
        ex.extend(row(
            20,
            pr(20, &h),
            vec![note],
            vec![
                check(1, "rc-ci", &h, "success"),
                check(2, "windows-build", &h, "success"),
            ],
        ));
        let (out, _) = run(ex);
        assert_eq!(out["rows"][0]["verdict"]["verdict"], verdict);
        assert_eq!(out["rows"][0]["verdict"]["comment_id"], 201);
        assert_eq!(out["rows"][0]["blocked_reasons"], reasons);
    }
}

#[test]
fn repair_fence_delimiters_keep_example_tables_hidden() {
    let example = "| position | PR |\n|---|---|\n|1|#99|\n";
    let actual = "| position | PR |\n|---|---|\n|1|#20|\n";
    for (open, nested, short, close) in [
        ("~~~~markdown", "```", "~~~", "~~~~"),
        ("````markdown", "~~~", "```", "`````"),
    ] {
        let body =
            format!("{open}\n{nested}\n{example}{nested}\n{short}\n{example}{close}\n{actual}");
        assert_eq!(parse_order(&body).unwrap(), vec![20], "{body}");
    }
}

#[test]
fn repair_final_pr_identity_change_refuses_same_sha_snapshot() {
    let h = "a".repeat(40);
    let base = "0".repeat(40);
    for field in ["base_branch", "number", "head_branch"] {
        let mut ex = vec![ledger(&[20])];
        ex.extend(row(
            20,
            pr(20, &h),
            vec![audit(20, &h, &base, "PASS")],
            vec![
                check(1, "rc-ci", &h, "success"),
                check(2, "windows-build", &h, "success"),
            ],
        ));
        let after = ex.last_mut().unwrap().response.body.as_mut().unwrap();
        match field {
            "base_branch" => after["base"]["ref"] = json!("unrelated-branch"),
            "number" => after["number"] = json!(99),
            "head_branch" => after["head"]["ref"] = json!("issue/7-other"),
            _ => unreachable!(),
        }
        let tmp = tempfile::tempdir().unwrap();
        common::git::git(tmp.path(), &["init"]);
        ex.push(ledger(&[20]));
        let t = RecordedTransport::new(ex);
        let result = query_cli(tmp.path(), req(), &t, &Fixed);
        assert!(
            result
                .as_ref()
                .is_err_and(|e| e.contains("changed during query")),
            "{field}: {result:?}"
        );
        assert_eq!(t.remaining(), 1);
    }
}

/// #2563: the CI check renamed `rc-ci` → `ci`. The `ci` row reads runs under either
/// spelling and the newest run (largest id) decides, whichever name it carries; a matrix
/// job name never counts. A newer failure under the other name is not hidden by an older
/// success, so accepting both names does not weaken the success requirement.
#[test]
fn ci_row_reads_both_names_and_the_newest_run_decides() {
    let h = "a".repeat(40);
    for (checks, state) in [
        (vec![check(10, "ci", &h, "success")], "success"),
        (vec![check(10, "rc-ci", &h, "success")], "success"),
        (
            vec![
                check(10, "rc-ci", &h, "success"),
                check(11, "ci", &h, "failure"),
            ],
            "failure",
        ),
        (
            vec![
                check(10, "ci", &h, "success"),
                check(11, "rc-ci", &h, "failure"),
            ],
            "failure",
        ),
        (
            vec![
                check(11, "ci", &h, "success"),
                check(10, "rc-ci", &h, "failure"),
            ],
            "success",
        ),
        (
            vec![check(10, "ci / linux-x86_64", &h, "success")],
            "not_run",
        ),
    ] {
        let mut ex = vec![ledger(&[20])];
        ex.extend(row(
            20,
            pr(20, &h),
            vec![audit(20, &h, &"0".repeat(40), "PASS")],
            checks.clone(),
        ));
        let (out, _) = run(ex);
        let row = &out["rows"][0];
        assert_eq!(row["ci"]["ci"]["state"], state, "{checks:?}");
        let blocked = row["blocked_reasons"].as_array().unwrap();
        assert_eq!(
            blocked.iter().any(|b| b == "ci_not_success"),
            state != "success",
            "{checks:?} {blocked:?}"
        );
    }
}
