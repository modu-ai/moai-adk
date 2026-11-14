//! Issue #2674 S4: `gate_loop::once` runs gate-merge on every open non-draft pull request and,
//! once per head, updates the branch on a code overlap (G6) or asks the claim holder (G2).
//! "Once" is the hidden marker in the pull request's comments: a second pass with the marker
//! present writes nothing. GitHub is replayed; git runs on a temporary repository.

mod common;

use common::git::{git, git_ok, git_raw};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::gate_loop::{once, LoopTarget};
use serde_json::{json, Value};
use std::fs;
use std::path::{Path, PathBuf};
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
    let a = commit(
        d,
        &[("src/a.rs", "1\n2\n3\n4\n5\n6\n7\n"), ("src/b.rs", "b\n")],
        "base",
    );
    git_ok(d, &["checkout", "-q", "-b", "pr"]);
    let b = commit(
        d,
        &[("src/a.rs", "one\n2\n3\n4\n5\n6\n7\n")],
        "pr\n\nMoai-Claim: 9#555",
    );
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
        // #2906 S5b: G5 reads the local check line for the head.
        json!({"id": 250, "body": format!("local-check: head={head} result=pass cmd=tools/local_check.sh at=2026-11-14T12:00:00Z\n")}),
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
            response: ok(json!({"head": {"sha": head}, "body": PASS_BODY})),
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
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues/comments/555")),
            response: ok(json!({"id": 555, "body": "moai-claim: session=agent-26 epoch=1\n"})),
        },
        // #2859: the open pull requests (only this one), read before the tip.
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/pulls?state=open&per_page=100")),
            response: ok(json!([{"number": 9}])),
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
        fetch: false,
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
fn a_passing_pull_request_is_merged() {
    let (dir, a, b) = repo(false);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(json!([
        {"number": 9, "draft": false, "head": {"sha": b}}
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
    // #2806: one static progress line per pull request.
    let seen: std::rc::Rc<std::cell::RefCell<Vec<String>>> = Default::default();
    let keep = seen.clone();
    moai_cli::progress::install(Box::new(move |l| keep.borrow_mut().push(l.to_string())));
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    );
    moai_cli::progress::uninstall();
    let out = out.unwrap();
    assert_eq!(
        out["pull_requests"],
        json!([{"pr": 9, "outcome": "merged", "merge_commit": "9999999999999999999999999999999999999999"}])
    );
    assert_eq!(t.remaining(), 0);
    assert_eq!(*seen.borrow(), ["checking pull request #9 (1/1)"]);
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("skipped"));
    assert!(writes(&t).is_empty());

    let t = RecordedTransport::new(vec![]);
    assert!(once(&target(&dir), "", false, false, &t, &t, &NoClock).is_err());
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
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
        ex[6] = tip_ci(&dir, json!([{"name": "ci", "conclusion": "failure"}]));
        let t = RecordedTransport::new(ex);
        let out = once(
            &target(&dir),
            "fixture-approved",
            false,
            false,
            &t,
            &t,
            &NoClock,
        )
        .unwrap();
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
        false,
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
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

/// #2860 rule 5: the oldest pull request goes first. GitHub lists newest first; three numbers
/// listed out of order are read 3, 7, 12 (each then fails its first read and is skipped).
#[test]
fn pull_requests_are_handled_oldest_number_first() {
    let (dir, _a, b) = repo(false);
    let t = RecordedTransport::new(vec![list(json!([
        {"number": 12, "draft": false, "head": {"sha": b}},
        {"number": 3, "draft": true, "head": {"sha": b}},
        {"number": 7, "draft": false, "head": {"sha": b}}
    ]))]);
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
    let order: Vec<u64> = out["pull_requests"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r["pr"].as_u64().unwrap())
        .collect();
    assert_eq!(order, [3, 7, 12], "{out}");
    let reads: Vec<String> = t
        .sent
        .borrow()
        .iter()
        .skip(1)
        .map(|r| r.url.trim_start_matches(API).to_string())
        .collect();
    assert_eq!(reads, ["/pulls/3", "/pulls/7", "/pulls/12"]);
}

/// The PR read gate-merge makes after the gate passes, for a draft (node id for GraphQL).
fn remote_draft(dir: &TempDir, head: &str) -> Vec<Exchange> {
    let mut ex = remote(dir, head);
    ex[0].response = ok(
        json!({"head": {"sha": head}, "base": {"ref": "main"}, "draft": true, "node_id": "PR_node9"}),
    );
    ex
}

fn ready_exchange(response: HttpResponse) -> Exchange {
    Exchange {
        request: HttpRequest::new("POST", "https://api.github.com/graphql"),
        response,
    }
}

/// #2860 rule 4 (leader decision 6870150891): a passing draft is gated like any other pull
/// request and marked ready after the status, right before the merge.
#[test]
fn a_passing_draft_is_marked_ready_then_merged() {
    let (dir, a, b) = repo(false);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": true, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    ex.extend(remote_draft(&dir, &b));
    ex.extend(status_exchanges(&b));
    ex.push(ready_exchange(ok(
        json!({"data": {"markPullRequestReadyForReview": {"pullRequest": {"isDraft": false}}}}),
    )));
    ex.extend([
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("merged"), "{out}");
    assert_eq!(t.remaining(), 0);
    {
        let sent = t.sent.borrow();
        let ready = sent
            .iter()
            .find(|r| r.url.ends_with("/graphql"))
            .expect("ready mutation sent");
        let body = ready.body.as_ref().unwrap();
        assert!(body["query"]
            .as_str()
            .unwrap()
            .contains("markPullRequestReadyForReview"));
        assert_eq!(body["variables"]["id"], json!("PR_node9"));
    }
    let order = writes(&t);
    let status = order
        .iter()
        .position(|w| w.starts_with("POST /statuses"))
        .unwrap();
    let merge = order
        .iter()
        .position(|w| w.starts_with("PUT /pulls/9/merge"))
        .unwrap();
    let ready_at = order.iter().position(|w| w.ends_with("/graphql")).unwrap();
    assert!(status < ready_at && ready_at < merge, "{order:?}");
}

/// A ready mutation that GitHub refuses (or answers with errors) merges nothing.
#[test]
fn a_draft_that_cannot_be_marked_ready_is_not_merged() {
    for response in [
        ok(json!({"errors": [{"message": "no"}]})),
        // An answer without errors whose pull request is still a draft (#2869 audit N1).
        ok(json!({"data": {"markPullRequestReadyForReview": {"pullRequest": {"isDraft": true}}}})),
        // GraphQL may answer partial data with errors: still not ready enough to merge.
        ok(json!({"errors": [{"message": "partial"}],
            "data": {"markPullRequestReadyForReview": {"pullRequest": {"isDraft": false}}}})),
        HttpResponse {
            status: 502,
            headers: Default::default(),
            body: None,
        },
    ] {
        let (dir, a, b) = repo(false);
        let cs = comments(&b, &a, "agent-10", None);
        let mut ex = vec![list(
            json!([{"number": 9, "draft": true, "head": {"sha": b}}]),
        )];
        ex.extend(reads(&dir, &b, &cs));
        ex.extend(remote_draft(&dir, &b));
        ex.extend(status_exchanges(&b));
        ex.push(ready_exchange(response));
        let t = RecordedTransport::new(ex);
        let out = once(
            &target(&dir),
            "fixture-approved",
            false,
            false,
            &t,
            &t,
            &NoClock,
        )
        .unwrap();
        let row = &out["pull_requests"][0];
        assert_eq!(row["outcome"], json!("stopped"), "{out}");
        assert_eq!(row["rule"], json!("READY"), "{out}");
        assert!(
            !writes(&t).iter().any(|w| w.contains("/merge")),
            "{:?}",
            writes(&t)
        );
    }
}

/// #2860 D2: a dry run reads, writes nothing, needs no permission reference, and names what the
/// pass would do — merge a passing pull request, update the branch on a code overlap.
#[test]
fn a_dry_run_writes_nothing_and_names_each_decision() {
    let (dir, a, b) = repo(false);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "", false, true, &t, &t, &NoClock).unwrap();
    assert_eq!(out["dry_run"], json!(true));
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("would_merge"),
        "{out}"
    );
    assert!(writes(&t).is_empty(), "{:?}", writes(&t));
    assert_eq!(t.remaining(), 0);

    let (dir, a, b) = repo(true);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    ex.push(marker_read(&cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "", false, true, &t, &t, &NoClock).unwrap();
    let row = &out["pull_requests"][0];
    assert_eq!(row["outcome"], json!("would_update_branch"), "{out}");
    assert_eq!(row["rule"], json!("G6"));
    assert!(writes(&t).is_empty(), "{:?}", writes(&t));
}

/// #2860 AC-3 (C2, #2833): the loop itself reaches the mechanical-delta acceptance. The PASS is
/// on head `b`; the open head `n` only merged the tip into `b` (an update-branch). One pass
/// records AUTO-DELTA first and merges `n` without a new verdict.
#[test]
fn the_loop_merges_a_tip_only_update_without_a_new_verdict() {
    let (dir, a, b) = repo(false);
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
    // The claim and the PASS name the earlier head `b`.
    // #2906 S5b: the local check line names the current head `n` (G5 reads it there).
    let lc = format!(
        "local-check: head={n} result=pass cmd=tools/local_check.sh at=2026-11-14T12:00:00Z\n"
    );
    let cs = comments(&b, &a, "agent-10", Some(&lc));
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": n}}]),
    )];
    ex.extend(reads(&dir, &n, &cs));
    ex.extend(remote(&dir, &n));
    ex.push(Exchange {
        request: HttpRequest::new("POST", format!("{API}/issues/9/comments")),
        response: created(),
    });
    ex.extend(status_exchanges(&n));
    ex.extend([
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
    let out = once(
        &target(&dir),
        "fixture-approved",
        false,
        false,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
    assert_eq!(out["pull_requests"][0]["outcome"], json!("merged"), "{out}");
    assert_eq!(t.remaining(), 0);
    let sent = t.sent.borrow();
    let first_note = sent
        .iter()
        .find(|r| r.method == "POST" && r.url.ends_with("/issues/9/comments"))
        .unwrap();
    let body = first_note.body.as_ref().unwrap()["body"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(
        body.lines().next().unwrap(),
        format!("AUTO-DELTA #9 {b}→{n}")
    );
    let merge_put = sent.iter().find(|r| r.method == "PUT").unwrap();
    assert_eq!(merge_put.body.as_ref().unwrap()["sha"], json!(n));
    drop(sent);

    // The dry run names the same decision without writing.
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": n}}]),
    )];
    ex.extend(reads(&dir, &n, &cs));
    let t = RecordedTransport::new(ex);
    let out = once(&target(&dir), "", false, true, &t, &t, &NoClock).unwrap();
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("would_merge_auto_delta"),
        "{out}"
    );
    assert!(writes(&t).is_empty());
}

/// #2869 audit B1: the leader runs the usual command with `--dry-run` added, so the permission
/// reference is present. The dry run must still stop at its own row: one row per pull request,
/// only GET requests sent, and the recorded reads used up exactly (a fall-through into
/// gate-merge would read the remote PR again and write).
#[test]
fn a_dry_run_with_a_permission_reference_still_writes_nothing() {
    for overlap in [false, true] {
        let (dir, a, b) = repo(overlap);
        let cs = comments(&b, &a, "agent-10", None);
        let mut ex = vec![list(
            json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
        )];
        ex.extend(reads(&dir, &b, &cs));
        if overlap {
            ex.push(marker_read(&cs));
        }
        let t = RecordedTransport::new(ex);
        let out = once(
            &target(&dir),
            "fixture-approved",
            false,
            true,
            &t,
            &t,
            &NoClock,
        )
        .unwrap();
        let rows = out["pull_requests"].as_array().unwrap();
        assert_eq!(rows.len(), 1, "{out}");
        let want = if overlap {
            "would_update_branch"
        } else {
            "would_merge"
        };
        assert_eq!(rows[0]["outcome"], json!(want), "{out}");
        assert!(
            t.sent.borrow().iter().all(|r| r.method == "GET"),
            "{:?}",
            writes(&t)
        );
        assert_eq!(t.remaining(), 0);
    }
}

/// #2906 S3: a bare `origin` whose base branch moved on to `c` and that carries
/// `refs/pull/9/head` = `b`, and a clone made before either (`--no-local`, so it holds only
/// `a`). `work` keeps `main` at `c`, so `reads(&work, …)` answers the tip read the gate makes
/// once `c` is fetched. Returns (work, root, clone, a, b, c).
fn origin_ahead() -> (TempDir, TempDir, PathBuf, String, String, String) {
    let (work, a, b) = repo(false);
    let c = git(work.path(), &["rev-parse", "main"]);
    let root = TempDir::new().unwrap();
    let origin = root.path().join("origin.git");
    let local = root.path().join("local");
    let o = origin.to_str().unwrap();
    git_ok(
        root.path(),
        &["clone", "-q", "--bare", work.path().to_str().unwrap(), o],
    );
    git_ok(&origin, &["update-ref", "-d", "refs/heads/pr"]);
    git_ok(&origin, &["update-ref", "refs/heads/main", &a]);
    git_ok(
        root.path(),
        &["clone", "-q", "--no-local", o, local.to_str().unwrap()],
    );
    git_ok(&origin, &["update-ref", "refs/heads/main", &c]);
    git_ok(&origin, &["update-ref", "refs/pull/9/head", &b]);
    (work, root, local, a, b, c)
}

fn clone_target(dir: &Path, fetch: bool) -> LoopTarget<'_> {
    LoopTarget {
        dir,
        owner: "o",
        repo: "r",
        base_branch: "main",
        tip: "refs/remotes/origin/main",
        review_confirmed_prs: vec![],
        fetch,
    }
}

/// #2906 S3 AC14: without `--fetch` a head the clone does not hold is skipped; with it the
/// pass first fetches the base branch and `refs/pull/<n>/head` into named remote refs (no
/// FETCH_HEAD) and judges the pull request on the fetched tip.
#[test]
fn fetch_brings_open_heads_before_the_pass() {
    let (work, _root, local, a, b, c) = origin_ahead();
    let cs = comments(&b, &a, "agent-10", None);
    let pr = json!([{"number": 9, "draft": false, "head": {"sha": b}}]);

    let mut ex = vec![list(pr.clone())];
    ex.extend(reads(&work, &b, &cs));
    let t = RecordedTransport::new(ex);
    let out = once(
        &clone_target(&local, false),
        "",
        false,
        true,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("skipped"),
        "{out}"
    );
    assert!(out.get("fetch").is_none(), "{out}");
    let probe = format!("{b}^{{commit}}");
    assert!(!git_raw(&local, &["cat-file", "-e", &probe]).success);
    assert_eq!(git(&local, &["rev-parse", "refs/remotes/origin/main"]), a);

    let mut ex = vec![list(pr)];
    ex.extend(reads(&work, &b, &cs));
    let t = RecordedTransport::new(ex);
    let out = once(
        &clone_target(&local, true),
        "",
        false,
        true,
        &t,
        &t,
        &NoClock,
    )
    .unwrap();
    assert_eq!(out["fetch"]["ok"], json!(true), "{out}");
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("would_merge"),
        "{out}"
    );
    assert_eq!(git(&local, &["rev-parse", "refs/remotes/origin/pull/9"]), b);
    assert_eq!(git(&local, &["rev-parse", "refs/remotes/origin/main"]), c);
    assert!(!local.join(".git").join("FETCH_HEAD").exists());
    assert!(writes(&t).is_empty(), "{:?}", writes(&t));
    assert_eq!(t.remaining(), 0);
}

/// #2906 S3: a fetch that fails is reported in the output, and the pass goes on with what is
/// local (here the repository has no `origin` at all, and the head is already local).
#[test]
fn a_failed_fetch_is_reported_and_the_pass_goes_on() {
    let (dir, a, b) = repo(false);
    let cs = comments(&b, &a, "agent-10", None);
    let mut ex = vec![list(
        json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
    )];
    ex.extend(reads(&dir, &b, &cs));
    let t = RecordedTransport::new(ex);
    let mut tg = target(&dir);
    tg.fetch = true;
    let out = once(&tg, "", false, true, &t, &t, &NoClock).unwrap();
    assert_eq!(out["fetch"]["ok"], json!(false), "{out}");
    assert!(
        out["fetch"]["error"]
            .as_str()
            .is_some_and(|e| !e.trim().is_empty()),
        "{out}"
    );
    assert_eq!(
        out["pull_requests"][0]["outcome"],
        json!("would_merge"),
        "{out}"
    );
    assert!(writes(&t).is_empty(), "{:?}", writes(&t));
}

/// #2906 S3 AC15: the shell loop's jq list (`auto_loop.sh`) never hands these to its gate —
/// the last `ci`/`rc-ci` conclusion is not SUCCESS, or the last verdict is not a PASS on the
/// current head. The dry run reaches the same end: `stopped`, never a `would_*` action.
/// #2906 S5b: CI no longer blocks, so a failed or running `ci` stops only without a
/// `local-check:` pass line for the head (G5); with one the PR would merge.
#[test]
fn shell_prefilter_cases_stop_in_a_dry_run() {
    let older = "efefefefefefefefefefefefefefefefefefefef";
    let (dir, a, b) = repo(false);
    let pass = comments(&b, &a, "agent-10", None);
    let checks = |c: Value| Exchange {
        request: HttpRequest::new("GET", format!("{API}/commits/{b}/check-runs?per_page=100")),
        response: ok(json!({"check_runs": [{"name": "ci", "conclusion": c}]})),
    };
    let fail_on_head = json!([
        {"id": 100, "body": format!("moai-audit-claim: lane=agent-10 head={b}\n")},
        {"id": 200, "body": format!("AUDIT #9 (x) head {b} base {a} — verdict: FAIL\n\n감사: agent-10.\n")},
    ]);
    let mut no_local_check = pass.clone();
    no_local_check
        .as_array_mut()
        .unwrap()
        .retain(|c| c["id"] != json!(250));
    let cases: Vec<(&str, Value, Option<Value>)> = vec![
        ("ci failure", no_local_check.clone(), Some(json!("failure"))),
        ("ci pending", no_local_check, Some(Value::Null)),
        (
            "PASS on an older head",
            comments(older, &a, "agent-10", None),
            None,
        ),
        ("FAIL on the current head", fail_on_head, None),
    ];
    for (name, cs, ci) in cases {
        let mut ex = vec![list(
            json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
        )];
        ex.extend(reads(&dir, &b, &cs));
        if let Some(c) = ci {
            ex[3] = checks(c);
        }
        let t = RecordedTransport::new(ex);
        let out = once(&target(&dir), "", false, true, &t, &t, &NoClock).unwrap();
        let row = &out["pull_requests"][0];
        assert_eq!(row["outcome"], json!("stopped"), "{name}: {out}");
        assert!(
            row["rule"].as_str().is_some_and(|r| !r.is_empty()),
            "{name}: {out}"
        );
        assert!(writes(&t).is_empty(), "{name}: {:?}", writes(&t));
    }
    for ci in [json!("failure"), Value::Null] {
        let mut ex = vec![list(
            json!([{"number": 9, "draft": false, "head": {"sha": b}}]),
        )];
        ex.extend(reads(&dir, &b, &pass));
        ex[3] = checks(ci.clone());
        let t = RecordedTransport::new(ex);
        let out = once(&target(&dir), "", false, true, &t, &t, &NoClock).unwrap();
        let row = &out["pull_requests"][0];
        assert_eq!(row["outcome"], json!("would_merge"), "ci={ci}: {out}");
        assert!(writes(&t).is_empty(), "ci={ci}: {:?}", writes(&t));
    }
}
