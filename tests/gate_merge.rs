//! Issue #2674 S2: `gate_merge::merge` writes only after the gate passes, in order — status,
//! merge with the head pinned, record — and stops at the first step that does not succeed.
//! GitHub is replayed through RecordedTransport; git runs on a temporary repository.

mod common;

use common::git::{git, git_ok};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::gate_collect::Target;
use moai_cli::gate_merge::{merge, MERGE_ATTEMPTS};
use serde_json::{json, Value};
use std::cell::RefCell;
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const API: &str = "https://api.github.com/repos/o/r";
/// A PR body that `pr-self-audit-check` passes (the four measured lines, #2279; #2708 made the
/// gate's body check the binary's own).
const PASS_BODY: &str = "Refs #2200\nMoai-Claim: 9#555\n\n## Self-Audit\n\n\
- mutation: claim-verify guard mutated (claim filter removed) -> check FAILED, measured\n\
- clone: check set from a fresh clone, not the /tmp symlink path\n\
- host: api.github.com exercised via fake gh executable, no network\n\
- receipt: receipt digest re-read and matched before PASS recorded\n\
- plan: PLAN comment 6872621175\n\
- red: gate_check failed first without the plan line\n";

struct FakeClock {
    slept: RefCell<Vec<u64>>,
}
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        1_000
    }
    fn sleep(&self, secs: u64) {
        self.slept.borrow_mut().push(secs);
    }
}
fn clock() -> FakeClock {
    FakeClock {
        slept: RefCell::new(Vec::new()),
    }
}

fn commit(dir: &Path, files: &[(&str, &str)], msg: &str) -> String {
    for (path, text) in files {
        let p = dir.join(path);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(p, text).unwrap();
    }
    git_ok(dir, &["add", "-A"]);
    git_ok(
        dir,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "commit",
            "-q",
            "-m",
            msg,
        ],
    );
    git(dir, &["rev-parse", "HEAD"]).trim().to_string()
}

/// Base A, PR head B (changes src/a.rs), tip C (changes src/b.rs); the body check is the
/// binary's own (#2708), so a passing body is [`PASS_BODY`].
fn repo() -> (TempDir, String, String) {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let a = commit(d, &[("src/a.rs", "a\n"), ("src/b.rs", "b\n")], "base");
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(d, &[("src/a.rs", "a2\n")], "pr\n\nMoai-Claim: 9#555");
    git_ok(d, &["checkout", "-q", "main"]);
    commit(d, &[("src/b.rs", "b2\n")], "tip");
    (dir, a, b)
}

fn ok(body: Value) -> HttpResponse {
    HttpResponse::new(200).with_body(body)
}

/// The four reads `collect` makes: the PR, its comments, its head's check runs and the tip's
/// push CI (green, #2777).
fn reads(head: &str, base: &str, tip: &str, body: &str) -> Vec<Exchange> {
    vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": head}, "body": body})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([
                {"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={head}\n")},
                {"id": 200, "body": format!("AUDIT #9 (x) head {head} base {base} — verdict: PASS\n\n감사: agent-10.\n")},
                // #2906 S5b: G5 reads the local check line for the head.
                local_check(head)
            ])),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{head}/check-runs?per_page=100"),
            ),
            response: ok(json!({"check_runs": [
                {"name": "ci", "conclusion": "success"},
                {"name": "ci / linux-x86_64", "conclusion": "success"}
            ]})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/comments/555")),
            response: ok(json!({"id": 555, "body": "moai-claim: session=agent-26 epoch=1\n"})),
        },
        // #2859: the open pull requests (only this one), read before the tip.
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls?state=open&per_page=100")),
            response: ok(json!([{"number": 9}])),
        },
        tip_ci(tip, json!([{"name": "ci", "conclusion": "success"}])),
    ]
}

/// The `local-check:` pass line for `head` (#2906 S5b, G5).
fn local_check(head: &str) -> Value {
    json!({"id": 250, "body": format!("local-check: head={head} result=pass cmd=tools/local_check.sh at=2026-11-14T12:00:00Z\n")})
}

/// The tip's push CI read (every attempt) answering `runs`.
fn tip_ci(tip: &str, runs: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new(
            "GET",
            format!("{API}/commits/{tip}/check-runs?filter=all&per_page=100"),
        ),
        response: ok(json!({ "check_runs": runs })),
    }
}

/// The PR and its base branch as GitHub has them, read after the gate passed (audit F1).
fn remote(head: &str, tip: &str) -> Vec<Exchange> {
    vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": head}, "base": {"ref": "main"}})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/git/ref/heads/main")),
            response: ok(json!({"object": {"sha": tip}})),
        },
    ]
}

fn tip_of(dir: &TempDir) -> String {
    git(dir.path(), &["rev-parse", "main"]).trim().to_string()
}

/// Every read up to the status: the gate's three and the remote two, all matching.
fn passing(dir: &TempDir, head: &str, base: &str) -> Vec<Exchange> {
    let mut ex = reads(head, base, &tip_of(dir), PASS_BODY);
    ex.extend(remote(head, &tip_of(dir)));
    ex
}

fn status(head: &str, code: u16) -> Vec<Exchange> {
    let mut v = vec![Exchange {
        request: HttpRequest::new("POST", format!("{API}/statuses/{head}")),
        response: HttpResponse::new(code).with_body(json!({})),
    }];
    if code == 201 {
        v.push(Exchange {
            request: HttpRequest::new("GET", format!("{API}/commits/{head}/statuses?per_page=100")),
            response: ok(
                json!([{"context": "moai/audit", "state": "success", "description": "x"}]),
            ),
        });
    }
    v
}

fn put(code: u16) -> Exchange {
    Exchange {
        request: HttpRequest::new("PUT", format!("{API}/pulls/9/merge")),
        response: HttpResponse::new(code).with_body(if code == 200 {
            json!({"sha": "9999999999999999999999999999999999999999", "merged": true})
        } else {
            json!({"message": "no"})
        }),
    }
}

fn record() -> Exchange {
    Exchange {
        request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
        response: HttpResponse::new(201).with_body(json!({"id": 1})),
    }
}

fn target(dir: &TempDir) -> Target<'_> {
    Target {
        dir: dir.path(),
        owner: "o",
        repo: "r",
        pr: 9,
        tip: "main",
        review_confirmed_prs: vec![],
    }
}

fn methods(t: &RecordedTransport) -> Vec<String> {
    t.sent.borrow().iter().map(|r| r.method.clone()).collect()
}

#[test]
fn a_passing_gate_posts_status_merges_with_the_head_and_records() {
    let (dir, a, b) = repo();
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.push(put(200));
    ex.push(record());
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(true), "{out}");
    assert_eq!(out["writes"], json!(["status", "merge", "record"]));
    assert_eq!(t.remaining(), 0);
    let sent = t.sent.borrow();
    let put_req = sent.iter().find(|r| r.method == "PUT").unwrap();
    assert_eq!(
        put_req.body.as_ref().unwrap()["sha"],
        json!(b),
        "the merge pins the head"
    );
    let rec = sent.iter().rfind(|r| r.method == "POST").unwrap();
    assert!(rec.body.as_ref().unwrap()["body"]
        .as_str()
        .unwrap()
        .contains("판정 200 PASS"));
}

#[test]
fn a_stopped_gate_writes_nothing() {
    let (dir, a, b) = repo();
    let t = RecordedTransport::new(reads(&b, &a, &tip_of(&dir), "no self-audit section"));
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false));
    assert_eq!(out["rule"], json!("G4"));
    assert_eq!(out["writes"], json!([]));
    assert_eq!(methods(&t), vec!["GET"; 6]);
}

/// #2777: on a red tip gate-merge stops at G8 before any write; the same PR merges once its
/// body names it the repair.
#[test]
fn a_red_tip_stops_the_merge_before_any_write() {
    let (dir, a, b) = repo();
    let tip = tip_of(&dir);
    let mut ex = reads(&b, &a, &tip, PASS_BODY);
    ex[5] = tip_ci(&tip, json!([{"name": "ci", "conclusion": "failure"}]));
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false), "{out}");
    assert_eq!(out["rule"], json!("G8"));
    assert_eq!(out["writes"], json!([]));
    assert_eq!(methods(&t), vec!["GET"; 6]);

    let mut ex = reads(
        &b,
        &a,
        &tip,
        &format!("{PASS_BODY}moai-tip-repair: tip run 1\n"),
    );
    ex[5] = tip_ci(&tip, json!([{"name": "ci", "conclusion": "failure"}]));
    ex.extend(remote(&b, &tip));
    ex.extend(status(&b, 201));
    ex.extend([put(200), record()]);
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(true), "{out}");
    assert_eq!(t.remaining(), 0);
}

#[test]
fn not_mergeable_is_retried_up_to_three_times_and_never_recorded() {
    let (dir, a, b) = repo();
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.extend([put(405), put(405), put(200), record()]);
    let t = RecordedTransport::new(ex);
    let c = clock();
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &c).unwrap();
    assert_eq!(out["merged"], json!(true));
    assert_eq!(*c.slept.borrow(), vec![5, 5]);

    assert_eq!(MERGE_ATTEMPTS, 3);
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.extend([put(405), put(405), put(405)]);
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false));
    assert_eq!(out["writes"], json!(["status"]));
    assert_eq!(t.remaining(), 0);
    assert_eq!(methods(&t).iter().filter(|m| *m == "PUT").count(), 3);
    assert!(
        !methods(&t)[9..].contains(&"POST".to_string()),
        "no record after a failed merge"
    );
}

#[test]
fn a_moved_head_is_not_retried() {
    let (dir, a, b) = repo();
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.push(put(409));
    let t = RecordedTransport::new(ex);
    let c = clock();
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &c).unwrap();
    assert_eq!(out["merged"], json!(false));
    assert!(c.slept.borrow().is_empty());
    assert_eq!(t.remaining(), 0);
}

#[test]
fn no_permission_or_a_failed_status_stops_before_the_merge() {
    let (dir, a, b) = repo();
    let t = RecordedTransport::new(vec![]);
    assert!(merge(&target(&dir), "", false, &t, &t, &clock()).is_err());
    assert!(
        t.sent.borrow().is_empty(),
        "nothing read without a permission reference"
    );

    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 500));
    let t = RecordedTransport::new(ex);
    assert!(merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).is_err());
    assert!(!methods(&t).contains(&"PUT".to_string()));
}

/// An exchange the transport fails on (method mismatch): stands for a timeout or a lost answer.
fn lost() -> Exchange {
    Exchange {
        request: HttpRequest::new("PATCH", "lost"),
        response: HttpResponse::new(200),
    }
}

fn pr_read_back(merged: bool) -> Exchange {
    Exchange {
        request: HttpRequest::new("GET", format!("{API}/pulls/9")),
        response: ok(json!({"merged": merged, "merge_commit_sha": if merged {
            json!("8888888888888888888888888888888888888888")
        } else {
            Value::Null
        }})),
    }
}

#[test]
fn a_remote_base_or_head_other_than_the_judged_one_writes_nothing() {
    let (dir, a, b) = repo();
    // GitHub's base moved past the local tip: the gate judged a stale tip.
    let mut ex = reads(&b, &a, &tip_of(&dir), PASS_BODY);
    ex.extend(remote(&b, &"1".repeat(40)));
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false), "{out}");
    assert_eq!(out["rule"], json!("G6"));
    assert_eq!(out["writes"], json!([]));
    assert_eq!(methods(&t), vec!["GET"; 8]);
    assert_eq!(t.remaining(), 0);

    // The head moved on GitHub after the snapshot.
    let mut ex = reads(&b, &a, &tip_of(&dir), PASS_BODY);
    ex.extend(remote(&"2".repeat(40), &tip_of(&dir)));
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["writes"], json!([]), "{out}");
    assert_eq!(methods(&t), vec!["GET"; 8]);
}

#[test]
fn a_lost_merge_answer_is_read_back() {
    let (dir, a, b) = repo();
    // The PUT merged but its answer was lost: the read-back finds it and the record follows.
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.extend([lost(), pr_read_back(true), record()]);
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(true), "{out}");
    assert_eq!(
        out["merge_commit"],
        json!("8888888888888888888888888888888888888888")
    );
    assert_eq!(out["writes"], json!(["status", "merge", "record"]));
    assert_eq!(t.remaining(), 0);

    // Not merged on read-back: no record.
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.extend([lost(), pr_read_back(false)]);
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false), "{out}");
    assert_eq!(out["writes"], json!(["status"]));
    assert_eq!(t.remaining(), 0);

    // The read-back fails too: the result says it does not know, never a bare error.
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.push(lost());
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!("uncertain"), "{out}");
    assert_eq!(out["writes"], json!(["status"]));
}

#[test]
fn a_record_that_fails_keeps_the_merge_result() {
    let (dir, a, b) = repo();
    // Refused record: merged, not recorded.
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.push(put(200));
    ex.push(Exchange {
        request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
        response: HttpResponse::new(403).with_body(json!({"message": "no"})),
    });
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(true), "{out}");
    assert_eq!(out["record_posted"], json!(false));
    assert_eq!(out["writes"], json!(["status", "merge"]));
    assert_eq!(
        out["merge_commit"],
        json!("9999999999999999999999999999999999999999")
    );

    // A lost record answer: maybe posted — said so, the merge result kept.
    let mut ex = passing(&dir, &b, &a);
    ex.extend(status(&b, 201));
    ex.extend([put(200), lost()]);
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(true), "{out}");
    assert_eq!(out["record_posted"], json!("uncertain"));
    assert_eq!(out["writes"], json!(["status", "merge"]));
}

#[test]
fn with_the_app_token_only_the_status_goes_through_the_app_transport() {
    let (dir, a, b) = repo();
    let mut ops = passing(&dir, &b, &a);
    ops.extend([put(200), record()]);
    let ops = RecordedTransport::new(ops);
    let app = RecordedTransport::new(status(&b, 201));
    let out = merge(
        &target(&dir),
        "fixture-approved",
        true,
        &ops,
        &app,
        &clock(),
    )
    .unwrap();
    assert_eq!(out["writes"], json!(["status", "merge", "record"]), "{out}");
    assert_eq!((ops.remaining(), app.remaining()), (0, 0));
    let app_urls: Vec<String> = app.sent.borrow().iter().map(|r| r.url.clone()).collect();
    assert!(
        app_urls.iter().all(|u| u.contains("/statuses")),
        "{app_urls:?}"
    );
    assert!(ops
        .sent
        .borrow()
        .iter()
        .all(|r| !r.url.contains("/statuses")));
}

// ---- #2833 D1: a PASS carried by a mechanical delta is recorded before any other write ----

/// `repo()` with the tip merged into the PR branch: the verdict head is the PR commit `h`, the
/// current head `n` is the update-branch merge.
fn delta_repo() -> (TempDir, String, String, String) {
    let (dir, a, h) = repo();
    let d = dir.path();
    git_ok(d, &["checkout", "-q", "pr"]);
    git_ok(
        d,
        &[
            "-c",
            "user.name=t",
            "-c",
            "user.email=t@example.invalid",
            "merge",
            "-q",
            "--no-ff",
            "-m",
            "Merge main",
            "main",
        ],
    );
    let n = git(d, &["rev-parse", "HEAD"]).trim().to_string();
    git_ok(d, &["checkout", "-q", "main"]);
    (dir, a, h, n)
}

/// The gate's reads for head `n` with the verdict on `h`, plus `extra` comments, then the two
/// remote reads.
fn delta_reads(dir: &TempDir, n: &str, h: &str, base: &str, extra: &[Value]) -> Vec<Exchange> {
    let tip = tip_of(dir);
    let mut comments = vec![
        json!({"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={h}\n")}),
        json!({"id": 200, "body": format!("AUDIT #9 (x) head {h} base {base} — verdict: PASS\n\n감사: agent-10.\n")}),
        local_check(n),
    ];
    comments.extend(extra.iter().cloned());
    let mut ex = reads(n, base, &tip, PASS_BODY);
    ex[1] = Exchange {
        request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
        response: ok(Value::Array(comments)),
    };
    ex.extend(remote(n, &tip));
    ex
}

#[test]
fn a_mechanical_delta_is_recorded_first_then_merged() {
    let (dir, a, h, n) = delta_repo();
    let mut ex = delta_reads(&dir, &n, &h, &a, &[]);
    ex.push(record()); // the AUTO-DELTA comment
    ex.extend(status(&n, 201));
    ex.push(put(200));
    ex.push(record()); // the merge record
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(true), "{out}");
    assert_eq!(
        out["writes"],
        json!(["auto_delta", "status", "merge", "record"])
    );
    assert_eq!(t.remaining(), 0);
    let sent = t.sent.borrow();
    let note = sent
        .iter()
        .find(|r| r.method == "POST" && r.url.ends_with("/issues/9/comments"))
        .unwrap();
    let body = note.body.as_ref().unwrap()["body"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(
        body.lines().next().unwrap(),
        format!("AUTO-DELTA #9 {h}→{n}"),
        "{body}"
    );
    assert!(body.contains("| `src/a.rs` |"), "{body}");
    // #2906 S5b: the record names the local check line G5 read, not GitHub CI.
    assert!(
        body.contains(&format!("local-check on {}: pass (G5).", &n[..10])),
        "{body}"
    );
    assert!(!body.contains("CI on"), "{body}");
    // The merge pins the current head, not the verdict head.
    let merge_put = sent.iter().find(|r| r.method == "PUT").unwrap();
    assert_eq!(merge_put.body.as_ref().unwrap()["sha"], json!(n));
}

#[test]
fn an_existing_auto_delta_line_is_not_written_again() {
    let (dir, a, h, n) = delta_repo();
    let line = json!({"id": 300, "body": format!("AUTO-DELTA #9 {h}→{n}\n\nearlier pass")});
    let mut ex = delta_reads(&dir, &n, &h, &a, &[line]);
    ex.extend(status(&n, 201));
    ex.push(put(200));
    ex.push(record());
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["writes"], json!(["status", "merge", "record"]), "{out}");
    assert_eq!(t.remaining(), 0);
}

#[test]
fn a_failed_auto_delta_record_writes_nothing_else() {
    let (dir, a, h, n) = delta_repo();
    let mut ex = delta_reads(&dir, &n, &h, &a, &[]);
    ex.push(Exchange {
        request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
        response: HttpResponse::new(403).with_body(json!({"message": "no"})),
    });
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false), "{out}");
    assert_eq!(out["writes"], json!([]));
    assert!(
        out["reason"].as_str().unwrap().contains("HTTP 403"),
        "{out}"
    );
    assert_eq!(t.remaining(), 0);
    assert!(t
        .sent
        .borrow()
        .iter()
        .all(|r| !r.url.contains("/statuses") && r.method != "PUT"));
}
