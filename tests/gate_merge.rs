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

/// Base A, PR head B (changes src/a.rs), tip C (changes src/b.rs); the checker stub passes a
/// body holding `SELF-AUDIT-OK`.
fn repo() -> (TempDir, String, String) {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let checker = "import sys\nsys.exit(0 if 'SELF-AUDIT-OK' in open(sys.argv[2]).read() else 1)\n";
    let a = commit(
        d,
        &[
            ("tools/check_pr_self_audit.py", checker),
            ("src/a.rs", "a\n"),
            ("src/b.rs", "b\n"),
        ],
        "base",
    );
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(d, &[("src/a.rs", "a2\n")], "pr");
    git_ok(d, &["checkout", "-q", "main"]);
    commit(d, &[("src/b.rs", "b2\n")], "tip");
    (dir, a, b)
}

fn ok(body: Value) -> HttpResponse {
    HttpResponse::new(200).with_body(body)
}

/// The three reads `collect` makes.
fn reads(head: &str, base: &str, body: &str) -> Vec<Exchange> {
    vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": head}, "body": body})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(json!([
                {"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={head}\n")},
                {"id": 200, "body": format!("AUDIT #9 (x) head {head} base {base} — verdict: PASS\n\n감사: agent-10.\n")}
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
    ]
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
    let mut ex = reads(head, base, "SELF-AUDIT-OK");
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
    let t = RecordedTransport::new(reads(&b, &a, "no self-audit section"));
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false));
    assert_eq!(out["rule"], json!("G4"));
    assert_eq!(out["writes"], json!([]));
    assert_eq!(methods(&t), vec!["GET", "GET", "GET"]);
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
        !methods(&t)[7..].contains(&"POST".to_string()),
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
    let mut ex = reads(&b, &a, "SELF-AUDIT-OK");
    ex.extend(remote(&b, &"1".repeat(40)));
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["merged"], json!(false), "{out}");
    assert_eq!(out["rule"], json!("G6"));
    assert_eq!(out["writes"], json!([]));
    assert_eq!(methods(&t), vec!["GET"; 5]);
    assert_eq!(t.remaining(), 0);

    // The head moved on GitHub after the snapshot.
    let mut ex = reads(&b, &a, "SELF-AUDIT-OK");
    ex.extend(remote(&"2".repeat(40), &tip_of(&dir)));
    let t = RecordedTransport::new(ex);
    let out = merge(&target(&dir), "fixture-approved", false, &t, &t, &clock()).unwrap();
    assert_eq!(out["writes"], json!([]), "{out}");
    assert_eq!(methods(&t), vec!["GET"; 5]);
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
