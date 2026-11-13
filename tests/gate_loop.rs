//! Issue #2674 S4: `gate_loop::once` runs gate-merge on every open non-draft pull request and,
//! once per head, updates the branch on a code overlap (G6) or asks the claim holder (G2).
//! "Once" is the hidden marker in the pull request's comments: a second pass with the marker
//! present writes nothing. GitHub is replayed; git runs on a temporary repository.

mod common;

use common::git::{git, git_ok};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::gate_loop::{once, LoopTarget};
use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use tempfile::TempDir;

const API: &str = "https://api.github.com/repos/o/r";

struct NoClock;
impl Clock for NoClock {
    fn now(&self) -> i64 {
        1_000
    }
    fn sleep(&self, _secs: u64) {}
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

/// Base A; the PR changes the first line of src/a.rs; the tip changes either src/b.rs (no
/// overlap) or the last line of src/a.rs (a clean code overlap).
fn repo(overlap: bool) -> (TempDir, String, String) {
    let dir = TempDir::new().unwrap();
    let d = dir.path();
    git_ok(d, &["init", "-q", "-b", "main"]);
    let checker = "import sys\nsys.exit(0 if 'SELF-AUDIT-OK' in open(sys.argv[2]).read() else 1)\n";
    let a = commit(
        d,
        &[
            ("tools/check_pr_self_audit.py", checker),
            ("src/a.rs", "1\n2\n3\n4\n5\n6\n7\n"),
            ("src/b.rs", "b\n"),
        ],
        "base",
    );
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(d, &[("src/a.rs", "one\n2\n3\n4\n5\n6\n7\n")], "pr");
    git_ok(d, &["checkout", "-q", "main"]);
    if overlap {
        commit(d, &[("src/a.rs", "1\n2\n3\n4\n5\n6\nseven\n")], "tip");
    } else {
        commit(d, &[("src/b.rs", "b2\n")], "tip");
    }
    (dir, a, b)
}

fn ok(body: Value) -> HttpResponse {
    HttpResponse::new(200).with_body(body)
}

fn list(prs: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new(
            "GET",
            format!("{API}/pulls?state=open&base=main&per_page=100"),
        ),
        response: ok(prs),
    }
}

fn comments(head: &str, base: &str, auditor: &str, extra: Option<&str>) -> Value {
    let mut v = vec![
        json!({"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={head}\n")}),
        json!({"id": 200, "body": format!("AUDIT #9 (x) head {head} base {base} — verdict: PASS\n\n감사: {auditor}.\n")}),
    ];
    if let Some(e) = extra {
        v.push(json!({"id": 300, "body": e}));
    }
    Value::Array(v)
}

/// The tip's push CI read (every attempt) on the local `main`, answering `runs` (#2777).
fn tip_ci(dir: &TempDir, runs: Value) -> Exchange {
    let tip = git(dir.path(), &["rev-parse", "main"]).trim().to_string();
    Exchange {
        request: HttpRequest::new(
            "GET",
            format!("{API}/commits/{tip}/check-runs?filter=all&per_page=100"),
        ),
        response: ok(json!({ "check_runs": runs })),
    }
}

/// The four reads gate-merge's collect makes for PR 9, the tip's push CI green.
fn reads(dir: &TempDir, head: &str, cs: &Value) -> Vec<Exchange> {
    vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls/9")),
            response: ok(json!({"head": {"sha": head}, "body": "SELF-AUDIT-OK"})),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
            response: ok(cs.clone()),
        },
        Exchange {
            request: HttpRequest::new(
                "GET",
                format!("{API}/commits/{head}/check-runs?per_page=100"),
            ),
            response: ok(json!({"check_runs": [{"name": "ci", "conclusion": "success"}]})),
        },
        tip_ci(dir, json!([{"name": "ci", "conclusion": "success"}])),
    ]
}

/// The PR and its base branch on GitHub, read by gate-merge after the gate passes.
fn remote(dir: &TempDir, head: &str) -> Vec<Exchange> {
    let tip = git(dir.path(), &["rev-parse", "main"]).trim().to_string();
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

fn marker_read(cs: &Value) -> Exchange {
    Exchange {
        request: HttpRequest::new("GET", format!("{API}/issues/9/comments?per_page=100")),
        response: ok(cs.clone()),
    }
}

fn created() -> HttpResponse {
    HttpResponse::new(201).with_body(json!({"id": 1}))
}

fn target(dir: &TempDir) -> LoopTarget<'_> {
    LoopTarget {
        dir: dir.path(),
        owner: "o",
        repo: "r",
        base_branch: "main",
        tip: "main",
        review_confirmed_prs: vec![],
    }
}

fn writes(t: &RecordedTransport) -> Vec<String> {
    t.sent
        .borrow()
        .iter()
        .filter(|r| r.method != "GET")
        .map(|r| format!("{} {}", r.method, r.url.trim_start_matches(API)))
        .collect()
}

#[test]
fn a_passing_pull_request_is_merged_and_a_draft_is_skipped() {
    let (dir, a, b) = repo(false);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(json!([
        {"number": 9, "draft": false, "head": {"sha": b}},
        {"number": 10, "draft": true, "head": {"sha": b}}
    ]))];
    ex.extend(reads(&dir, &b, &cs));
    ex.extend(remote(&dir, &b));
    ex.extend([
        Exchange {
            request: HttpRequest::new("POST", format!("{API}/statuses/{b}")),
            response: created(),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/commits/{b}/statuses?per_page=100")),
            response: ok(json!([])),
        },
        Exchange {
            request: HttpRequest::new("PUT", format!("{API}/pulls/9/merge")),
            response: ok(json!({"sha": "9999999999999999999999999999999999999999"})),
        },
        Exchange {
            request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
            response: created(),
        },
    ]);
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(
        out["pull_requests"],
        json!([{"pr": 9, "outcome": "merged", "merge_commit": "9999999999999999999999999999999999999999"}])
    );
    assert_eq!(t.remaining(), 0, "the draft is never read");
}

#[test]
fn a_code_overlap_updates_the_branch_once_per_head() {
    let (dir, a, b) = repo(true);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    ex.push(marker_read(&cs));
    ex.push(Exchange {
        request: HttpRequest::new("PUT", format!("{API}/pulls/9/update-branch")),
        response: HttpResponse::new(202).with_body(json!({})),
    });
    ex.push(Exchange {
        request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
        response: created(),
    });
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("updated_branch"),
        "{out}"
    );
    assert_eq!(
        writes(&t),
        vec!["PUT /pulls/9/update-branch", "POST /issues/9/comments"]
    );
    let sent = t.sent.borrow();
    let note = sent.iter().rfind(|r| r.method == "POST").unwrap();
    let mark = format!("<!-- moai:gate-loop update-branch head={b} -->");
    assert!(note.body.as_ref().unwrap()["body"]
        .as_str()
        .unwrap()
        .contains(&mark));
    let put = sent.iter().find(|r| r.method == "PUT").unwrap();
    assert_eq!(put.body.as_ref().unwrap()["expected_head_sha"], json!(b));
    drop(sent);

    // The next pass sees the marker and writes nothing.
    let cs = comments(&b, &a, "agent-10", Some(&mark));
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    ex.push(marker_read(&cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("already_updated"));
    assert!(writes(&t).is_empty());
}

#[test]
fn a_verdict_by_a_lane_without_a_claim_asks_once() {
    let (dir, a, b) = repo(false);
    let cs = comments(&b, &a, "agent-27", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    ex.push(marker_read(&cs));
    ex.push(Exchange {
        request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
        response: created(),
    });
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("asked_auditor"));
    assert_eq!(writes(&t), vec!["POST /issues/9/comments"]);

    let mark = format!("<!-- moai:gate-loop auditor head={b} -->");
    let cs = comments(&b, &a, "agent-27", Some(&mark));
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    ex.push(marker_read(&cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("already_asked"));
    assert!(writes(&t).is_empty());
}

#[test]
fn a_head_not_fetched_is_skipped_and_no_permission_reads_nothing() {
    let (dir, a, _) = repo(false);
    let missing = "cdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd";
    let cs = comments(missing, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": missing}}]),
    )];
    ex.extend(reads(&dir, missing, &cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("skipped"));
    assert!(writes(&t).is_empty());

    let t = RecordedTransport::new(vec![]);
    assert!(once(&target(&dir), "", false, &t, &t, &NoClock).is_err());
    assert!(t.sent.borrow().is_empty());
}

/// A real conflict is a G6 stop too, but not the overlap the loop resolves: no update-branch,
/// no note (mutation L4).
#[test]
fn a_conflict_is_reported_not_updated() {
    let (dir, a, b) = repo(false);
    let d = dir.path();
    commit(
        d,
        &[("src/a.rs", "uno\n2\n3\n4\n5\n6\n7\n")],
        "tip changes the PR's line",
    );
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("stopped"),
        "{out}"
    );
    assert_eq!(out["pull_requests"][0]["rule"], json!("G6"));
    assert!(writes(&t).is_empty());
}

/// #2777 AC-2: gate-loop uses the same G8 judgment. On a red tip a passing PR is not merged
/// and a code-overlap PR gets no update-branch or note — the pass writes nothing.
#[test]
fn a_red_tip_stops_the_loop_without_any_write() {
    for overlap in [false, true] {
        let (dir, a, b) = repo(overlap);
        let cs = comments(&b, &a, "agent-10", None);
        let mut ex = vec![list(
            json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
        )];
        ex.extend(reads(&dir, &b, &cs));
        ex[4] = tip_ci(&dir, json!([{"name": "ci", "conclusion": "failure"}]));
        let t = RecordedTransport::new(ex);
        let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
        let row = &out["pull_requests"][0];
        assert_eq!(row["outcome"], json!("stopped"), "overlap={overlap} {out}");
        assert_eq!(row["rule"], json!("G8"), "overlap={overlap} {out}");
        assert!(writes(&t).is_empty(), "overlap={overlap}");
        assert_eq!(t.remaining(), 0);
    }
}

fn passing_until_status(dir: &TempDir, a: &str, b: &str) -> Vec<Exchange> {
    let cs = comments(b, a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(dir, b, &cs));
    ex.extend(remote(dir, b));
    ex
}

fn status_exchanges(b: &str) -> Vec<Exchange> {
    vec![
        Exchange {
            request: HttpRequest::new("POST", format!("{API}/statuses/{b}")),
            response: created(),
        },
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/commits/{b}/statuses?per_page=100")),
            response: ok(json!([])),
        },
    ]
}

#[test]
fn with_the_app_token_only_the_status_goes_through_the_app_transport() {
    let (dir, a, b) = repo(false);
    let mut ops = passing_until_status(&dir, &a, &b);
    ops.extend([
        Exchange {
            request: HttpRequest::new("PUT", format!("{API}/pulls/9/merge")),
            response: ok(json!({"sha": "9999999999999999999999999999999999999999"})),
        },
        Exchange {
            request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
            response: created(),
        },
    ]);
    let ops = RecordedTransport::new(ops);
    let app = RecordedTransport::new(status_exchanges(&b));
    let out = once(
        &target(&dir),
        "fixture-approved",
        true,
        &ops,
        &app,
        &NoClock,
    )
    .unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("merged"), "{out}");
    assert_eq!((ops.remaining(), app.remaining()), (0, 0));
    assert!(app
        .sent
        .borrow()
        .iter()
        .all(|r| r.url.contains("/statuses")));
    assert!(ops
        .sent
        .borrow()
        .iter()
        .all(|r| !r.url.contains("/statuses")));
}

#[test]
fn a_merge_that_cannot_be_read_back_is_reported_uncertain() {
    let (dir, a, b) = repo(false);
    let mut ex = passing_until_status(&dir, &a, &b);
    ex.extend(status_exchanges(&b));
    // The merge answer is lost (method mismatch) and nothing is left to read it back.
    ex.push(Exchange {
        request: HttpRequest::new("PATCH", "lost"),
        response: ok(json!({})),
    });
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "fixture-approved", false, &t, &t, &NoClock).unwrap();
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("uncertain"),
        "{out}"
    );
    assert_eq!(
        writes(&t),
        vec![
            format!("POST /statuses/{b}"),
            "PUT /pulls/9/merge".to_string()
        ]
    );
}
