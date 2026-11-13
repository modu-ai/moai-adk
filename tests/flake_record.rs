//! Issue #2666 F2: `flake_record::record` keeps one open `[flake] <test>` Issue per flaky test
//! and one entry per run and job — created, commented, or left alone when already recorded.
//! GitHub is replayed through RecordedTransport.

use moai_cli::flake_record::{record, run_marker, test_marker};
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use serde_json::{json, Value};

const API: &str = "https://api.github.com/repos/o/r";
const T1: &str = "moai-cli::apply_receipt apply_recover_restores_only_unchanged_targets";
const MAC: &str = "ci / macos-arm64";
const LINUX: &str = "ci / linux-x86_64";
const RUN: u64 = 43386397109;

fn classified(rows: &[(&str, &str)]) -> Value {
    let flaky: Vec<Value> = rows
        .iter()
        .map(|(job, test)| json!({"job": job, "test": test, "message": "panicked at tests/x.rs:1:2 boom"}))
        .collect();
    json!({"flaky": flaky, "still_failing": [{"job": MAC, "test": "moai-cli::real fails"}], "not_in_later": [],
        "first_heads": {MAC: "1111111111111111111111111111111111111111"}})
}

fn get(url: String, body: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new("GET", url),
        response: HttpResponse::new(200).with_body(body),
    }
}

fn issues(list: Value) -> Exchange {
    get(format!("{API}/issues?state=open&per_page=100"), list)
}

fn comments(n: u64, list: Value) -> Exchange {
    get(format!("{API}/issues/{n}/comments?per_page=100"), list)
}

fn post(url: String, code: u16, body: Value) -> Exchange {
    Exchange {
        request: HttpRequest::new("POST", url),
        response: HttpResponse::new(code).with_body(body),
    }
}

fn create(number: u64, test: &str, job: &str) -> Exchange {
    post(
        format!("{API}/issues"),
        201,
        json!({"number": number, "body": format!("{}\n\n{}", test_marker(test), run_marker(RUN, job))}),
    )
}

/// An exchange the transport fails on (method mismatch): a lost answer.
fn lost() -> Exchange {
    Exchange {
        request: HttpRequest::new("PATCH", "lost"),
        response: HttpResponse::new(200),
    }
}

fn tracked(number: u64, test: &str) -> Value {
    json!({"number": number, "body": format!("{}\nfirst entry", test_marker(test))})
}

fn run(t: &RecordedTransport, c: &Value) -> Value {
    record("o", "r", RUN, c, "fixture-approved", t).unwrap()
}

fn sent_posts(t: &RecordedTransport) -> Vec<HttpRequest> {
    t.sent
        .borrow()
        .iter()
        .filter(|r| r.method == "POST")
        .cloned()
        .collect()
}

#[test]
fn a_test_without_an_open_issue_gets_one() {
    let t = RecordedTransport::new(vec![
        // A pull request carrying the marker and an Issue naming it on a later line do not count.
        issues(json!([
            {"number": 5, "pull_request": {}, "body": test_marker(T1)},
            {"number": 6, "body": format!("about\n{}", test_marker(T1))}
        ])),
        create(70, T1, MAC),
    ]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(
        out["created"],
        json!([{"test": T1, "job": MAC, "issue": 70}]),
        "{out}"
    );
    assert_eq!(out["writes"], json!(1));
    assert_eq!(t.remaining(), 0);
    let posts = sent_posts(&t);
    let body = posts[0].body.as_ref().unwrap();
    assert_eq!(body["title"], json!(format!("[flake] {T1}")));
    let text = body["body"].as_str().unwrap();
    assert_eq!(text.lines().next(), Some(test_marker(T1).as_str()));
    assert!(text.lines().any(|l| l == run_marker(RUN, MAC)));
    assert!(text.contains("panicked at tests/x.rs:1:2 boom"));
    assert!(text.contains("- head: `1111111111111111111111111111111111111111`"));
    assert!(text.contains(&format!("https://github.com/o/r/actions/runs/{RUN}")));
    // Only flaky rows are recorded.
    assert!(!text.contains("moai-cli::real"));
}

#[test]
fn an_open_issue_gets_a_comment_once_per_run_and_job() {
    let t = RecordedTransport::new(vec![
        issues(json!([tracked(40, T1)])),
        comments(40, json!([{"body": run_marker(1, MAC)}])), // another run
        post(format!("{API}/issues/40/comments"), 201, json!({"id": 1})),
    ]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(
        out["commented"],
        json!([{"test": T1, "job": MAC, "issue": 40}]),
        "{out}"
    );
    let posts = sent_posts(&t);
    let text = posts[0].body.as_ref().unwrap()["body"]
        .as_str()
        .unwrap()
        .to_string();
    assert_eq!(text.lines().next(), Some(run_marker(RUN, MAC).as_str()));

    // The same run and job again: nothing written.
    let t = RecordedTransport::new(vec![
        issues(json!([tracked(40, T1)])),
        comments(
            40,
            json!([{"body": format!("{}\nx", run_marker(RUN, MAC))}]),
        ),
    ]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(out["writes"], json!(0), "{out}");
    assert_eq!(
        out["already_recorded"],
        json!([{"test": T1, "job": MAC, "issue": 40}])
    );
    assert!(sent_posts(&t).is_empty());
    assert_eq!(t.remaining(), 0);

    // Recorded in the Issue body itself (the run that created it): no comment read needed.
    let t = RecordedTransport::new(vec![issues(json!([{"number": 40,
        "body": format!("{}\n\n{}", test_marker(T1), run_marker(RUN, MAC))}]))]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(out["writes"], json!(0), "{out}");
}

#[test]
fn two_jobs_of_one_test_share_the_new_issue() {
    let t = RecordedTransport::new(vec![
        issues(json!([])),
        create(71, T1, LINUX),
        comments(71, json!([])),
        post(format!("{API}/issues/71/comments"), 201, json!({"id": 2})),
    ]);
    let out = run(&t, &classified(&[(LINUX, T1), (MAC, T1)]));
    assert_eq!(
        out["created"],
        json!([{"test": T1, "job": LINUX, "issue": 71}]),
        "{out}"
    );
    assert_eq!(
        out["commented"],
        json!([{"test": T1, "job": MAC, "issue": 71}])
    );
    assert_eq!(t.remaining(), 0);
}

#[test]
fn two_open_issues_for_one_test_stop_before_writing() {
    let t = RecordedTransport::new(vec![issues(json!([tracked(40, T1), tracked(41, T1)]))]);
    let err = record(
        "o",
        "r",
        RUN,
        &classified(&[(MAC, T1)]),
        "fixture-approved",
        &t,
    )
    .unwrap_err();
    assert!(err.contains("more than one open Issue"), "{err}");
    assert!(sent_posts(&t).is_empty());
}

#[test]
fn listings_follow_the_next_link() {
    let next = format!("{API}/issues?state=open&per_page=100&page=2");
    let t = RecordedTransport::new(vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues?state=open&per_page=100")),
            response: HttpResponse::new(200)
                .with_header("link", format!("<{next}>; rel=\"next\""))
                .with_body(json!([{"number": 1, "body": "other"}])),
        },
        get(next, json!([tracked(40, T1)])),
        comments(40, json!([])),
        post(format!("{API}/issues/40/comments"), 201, json!({"id": 1})),
    ]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(
        out["commented"][0]["issue"],
        json!(40),
        "the Issue on page 2: {out}"
    );
}

#[test]
fn no_permission_reference_reads_and_writes_nothing() {
    let t = RecordedTransport::new(vec![]);
    assert!(record("o", "r", RUN, &classified(&[(MAC, T1)]), "", &t).is_err());
    assert!(t.sent.borrow().is_empty());
}

#[test]
fn a_lost_answer_is_read_back_before_anything_is_claimed() {
    // Lost create answer, found on the second listing: created, no second Issue.
    let t = RecordedTransport::new(vec![
        issues(json!([])),
        lost(),
        issues(json!([{"number": 72, "body": test_marker(T1)}])),
    ]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(
        out["created"],
        json!([{"test": T1, "job": MAC, "issue": 72}]),
        "{out}"
    );
    assert_eq!(t.remaining(), 0);

    // Lost create answer, not there on the second listing: an error, not a success.
    let t = RecordedTransport::new(vec![issues(json!([])), lost(), issues(json!([]))]);
    assert!(record(
        "o",
        "r",
        RUN,
        &classified(&[(MAC, T1)]),
        "fixture-approved",
        &t
    )
    .is_err());

    // Lost create answer and the listing fails too: uncertain, not created.
    let t = RecordedTransport::new(vec![issues(json!([])), lost()]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(out["created"], json!([]), "{out}");
    assert_eq!(out["uncertain"][0]["test"], json!(T1));
    assert_eq!(out["writes"], json!(0));

    // Lost comment answer, found on read-back: commented.
    let t = RecordedTransport::new(vec![
        issues(json!([tracked(40, T1)])),
        comments(40, json!([])),
        lost(),
        comments(40, json!([{"body": run_marker(RUN, MAC)}])),
    ]);
    let out = run(&t, &classified(&[(MAC, T1)]));
    assert_eq!(out["commented"][0]["issue"], json!(40), "{out}");
}

#[test]
fn a_refused_write_is_an_error() {
    let t = RecordedTransport::new(vec![
        issues(json!([])),
        post(format!("{API}/issues"), 403, json!({"message": "no"})),
    ]);
    let err = record(
        "o",
        "r",
        RUN,
        &classified(&[(MAC, T1)]),
        "fixture-approved",
        &t,
    )
    .unwrap_err();
    assert!(err.contains("HTTP 403"), "{err}");
}

#[test]
fn a_lost_comment_answer_not_found_on_read_back_is_an_error() {
    let t = RecordedTransport::new(vec![
        issues(json!([tracked(40, T1)])),
        comments(40, json!([])),
        lost(),
        comments(40, json!([{"body": run_marker(1, MAC)}])), // another run only
    ]);
    let err = record(
        "o",
        "r",
        RUN,
        &classified(&[(MAC, T1)]),
        "fixture-approved",
        &t,
    )
    .unwrap_err();
    assert!(err.contains("comment on #40 failed"), "{err}");
}
