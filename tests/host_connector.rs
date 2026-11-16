//! #3071 H1: the host connector. Card code asks a host four things (open issues, a summary
//! posted once, a PR's state, a PR's verdicts) through `HostConnector`; the GitHub implementation
//! runs on a recorded transport here, the card commands' tests run on `FakeConnector`. Nothing in
//! this file touches the network, a real HOME or a token.
mod common;

use common::{exchange, json_response};
use moai_cli::forge::connector::{
    check_marker, Call, ConnectorError, FakeConnector, GithubConnector, HostComment, HostConnector,
    HostIssue, HostKind, PrPhase, PrState, ProjectRef, SummaryOutcome, SummaryTarget, Verdicts,
};
use moai_cli::forge::policy::Clock;
use moai_cli::forge::transport::{Exchange, HttpResponse, RecordedTransport};
use moai_cli::forge::Config;
use serde_json::{json, Value};

const BASE: &str = "https://api.github.com/repos/modu-ai/moai-adk";

struct NoSleep;

impl Clock for NoSleep {
    fn now(&self) -> i64 {
        0
    }
    fn sleep(&self, _secs: u64) {}
}

fn project() -> ProjectRef {
    ProjectRef::parse(HostKind::Github, "modu-ai/moai-adk").unwrap()
}

fn issues_url() -> String {
    format!("{BASE}/issues?state=open&sort=created&direction=asc&per_page=100")
}

fn comments_url(n: u64) -> String {
    format!("{BASE}/issues/{n}/comments?per_page=100")
}

fn get(url: &str, response: HttpResponse) -> Exchange {
    exchange("GET", url, response)
}

fn issue(n: u64) -> Value {
    json!({
        "number": n, "title": format!("title {n}"), "body": format!("body {n}"),
        "labels": [{"name": "kind:task"}, {"name": "rc.2"}],
        "user": {"login": "someone"}, "author_association": "MEMBER",
        "created_at": "2026-11-15T00:00:00Z", "updated_at": "2026-11-15T01:00:00Z",
    })
}

fn comment(id: u64, association: &str, body: &str) -> Value {
    json!({"id": id, "user": {"login": "writer"}, "author_association": association,
           "body": body, "updated_at": "2026-11-15T02:00:00Z"})
}

fn next_page(url: &str) -> String {
    format!("<{url}&page=2>; rel=\"next\"")
}

/// Runs `body` against a connector that replays `exchanges`; returns the transport for counting.
fn with_connector<R>(
    exchanges: Vec<Exchange>,
    config: Config,
    body: impl FnOnce(&GithubConnector<'_, RecordedTransport, NoSleep>) -> R,
) -> (R, RecordedTransport) {
    let transport = RecordedTransport::new(exchanges);
    let result = {
        let connector = GithubConnector::new(&transport, &NoSleep, config);
        body(&connector)
    };
    (result, transport)
}

fn marker() -> &'static str {
    "<!-- moai-card-export card=C-7 sha256=abc -->"
}

#[test]
fn a_project_is_written_one_way_and_gitlab_is_refused_before_any_request() {
    let p = ProjectRef::parse(HostKind::Github, "Modu-AI/MoAI-ADK").unwrap();
    assert_eq!(p.source_project(), "modu-ai/moai-adk");
    assert_eq!(p.host().as_str(), "github");
    for bad in ["", "a", "a/b/c", "/b", "a/", "a b/c", "../x", "a/.."] {
        assert!(
            matches!(
                ProjectRef::parse(HostKind::Github, bad),
                Err(ConnectorError::Invalid(_))
            ),
            "{bad:?}"
        );
    }
    // GitLab: the host is read, then refused with what a connector needs; no connector, no
    // transport and no token exist at that point.
    let host = HostKind::parse("gitlab").unwrap();
    assert_eq!(host, HostKind::Gitlab);
    let refusal = host.require_implemented().unwrap_err();
    let message = refusal.to_string();
    assert!(matches!(refusal, ConnectorError::NotImplemented { .. }));
    assert!(message.contains("not implemented"), "{message}");
    assert!(message.contains("test project"), "{message}");
    assert!(message.contains("access token"), "{message}");
    assert_eq!(
        ProjectRef::parse(HostKind::Gitlab, "group/project").unwrap_err(),
        refusal
    );
    assert!(matches!(
        HostKind::parse("bitbucket"),
        Err(ConnectorError::Invalid(_))
    ));
}

#[test]
fn open_issues_carry_their_text_labels_and_comments_with_who_wrote_them() {
    let mut pr = issue(9);
    pr["pull_request"] = json!({"url": "x"});
    let (listed, transport) = with_connector(
        vec![
            // The host lists them out of order; the answer is ascending by number.
            get(
                &issues_url(),
                json_response(200, json!([issue(5), pr, issue(3)])),
            ),
            get(&comments_url(5), json_response(200, json!([]))),
            get(
                &comments_url(3),
                json_response(
                    200,
                    json!([comment(11, "OWNER", "claim"), comment(12, "NONE", "claim")]),
                ),
            ),
        ],
        Config::default(),
        |c| c.open_issues(&project()),
    );
    let listed = listed.unwrap();
    assert_eq!(transport.remaining(), 0);
    assert_eq!(listed.pull_requests_skipped, 1);
    let numbers: Vec<u64> = listed.issues.iter().map(|i| i.number).collect();
    assert_eq!(numbers, [3, 5]);
    let first = &listed.issues[0];
    assert_eq!(first.title, "title 3");
    assert_eq!(first.body, "body 3");
    assert_eq!(first.labels, ["kind:task", "rc.2"]);
    assert!(first.by_member());
    let by_member: Vec<(u64, bool)> = first
        .comments
        .iter()
        .map(|c| (c.id, c.by_member()))
        .collect();
    assert_eq!(by_member, [(11, true), (12, false)]);
}

#[test]
fn a_comment_without_an_association_is_not_a_member_comment() {
    let mut bare = comment(1, "", "x");
    bare.as_object_mut().unwrap().remove("author_association");
    for (value, expected) in [
        (bare, false),
        (comment(2, "member", "x"), false),
        (comment(3, "COLLABORATOR", "x"), true),
    ] {
        let (listed, _) = with_connector(
            vec![
                get(&issues_url(), json_response(200, json!([issue(1)]))),
                get(&comments_url(1), json_response(200, json!([value]))),
            ],
            Config::default(),
            |c| c.open_issues(&project()),
        );
        let listed = listed.unwrap();
        assert_eq!(listed.issues[0].comments[0].by_member(), expected);
    }
}

/// N issues over one page boundary: 100 fit on a page, the 101st needs the second.
fn listing(count: u64) -> Vec<Exchange> {
    let first: Vec<Value> = (1..=count.min(100)).map(issue).collect();
    let mut exchanges = vec![];
    let mut page_one = json_response(200, json!(first));
    if count > 100 {
        page_one = page_one.with_header("link", next_page(&issues_url()));
    }
    exchanges.push(get(&issues_url(), page_one));
    if count > 100 {
        let rest: Vec<Value> = (101..=count).map(issue).collect();
        exchanges.push(get(
            &format!("{}&page=2", issues_url()),
            json_response(200, json!(rest)),
        ));
    }
    for n in 1..=count {
        exchanges.push(get(&comments_url(n), json_response(200, json!([]))));
    }
    exchanges
}

#[test]
fn every_page_is_read_at_one_hundred_and_at_one_hundred_and_one_issues() {
    for (count, requests) in [(100u64, 101usize), (101, 103)] {
        let (listed, transport) = with_connector(listing(count), Config::default(), |c| {
            c.open_issues(&project())
        });
        let listed = listed.unwrap();
        assert_eq!(listed.issues.len() as u64, count, "{count} issues");
        assert_eq!(transport.sent.borrow().len(), requests, "{count} issues");
        assert_eq!(transport.remaining(), 0);
        assert!(transport.sent.borrow().iter().all(|r| r.method == "GET"));
    }
}

#[test]
fn a_listing_cut_short_by_the_page_limit_is_incomplete_not_shorter() {
    let config = Config {
        max_pages: 1,
        ..Config::default()
    };
    let (listed, transport) = with_connector(listing(101), config, |c| c.open_issues(&project()));
    assert!(matches!(
        listed,
        Err(ConnectorError::Incomplete { ref reason }) if reason.contains("page limit")
    ));
    // The second page was never asked for.
    assert_eq!(transport.sent.borrow().len(), 1);

    // The comments of one issue count the same way.
    let config = Config {
        max_pages: 1,
        ..Config::default()
    };
    let (listed, _) = with_connector(
        vec![
            get(&issues_url(), json_response(200, json!([issue(1)]))),
            get(
                &comments_url(1),
                json_response(200, json!([comment(1, "OWNER", "x")]))
                    .with_header("link", next_page(&comments_url(1))),
            ),
        ],
        config,
        |c| c.open_issues(&project()),
    );
    assert!(matches!(listed, Err(ConnectorError::Incomplete { .. })));
}

#[test]
fn a_summary_is_posted_once_with_its_marker_and_no_token() {
    let target = SummaryTarget {
        project: project(),
        number: 42,
    };
    let (outcome, transport) = with_connector(
        vec![
            get(&comments_url(42), json_response(200, json!([]))),
            exchange(
                "POST",
                &format!("{BASE}/issues/42/comments"),
                json_response(201, json!({"id": 77})),
            ),
        ],
        Config::default(),
        |c| {
            c.post_summary(
                &target,
                "status ready\ntoken ghp_abcdefghijklmnopqrstuvwx\n",
                marker(),
            )
        },
    );
    assert_eq!(outcome.unwrap(), SummaryOutcome::Posted { comment_id: 77 });
    let sent = transport.sent.borrow();
    let body = sent[1].body.as_ref().unwrap()["body"].as_str().unwrap();
    assert!(body.ends_with(&format!("\n\n{}", marker())), "{body}");
    assert!(!body.contains("ghp_"), "{body}");
    assert!(body.contains("[redacted]"), "{body}");
}

#[test]
fn a_second_summary_finds_the_members_marker_and_posts_nothing() {
    let target = SummaryTarget {
        project: project(),
        number: 42,
    };
    let (outcome, transport) = with_connector(
        vec![get(
            &comments_url(42),
            json_response(
                200,
                json!([comment(5, "OWNER", &format!("summary\n\n{}", marker()))]),
            ),
        )],
        Config::default(),
        |c| c.post_summary(&target, "summary", marker()),
    );
    assert_eq!(
        outcome.unwrap(),
        SummaryOutcome::AlreadyPosted { comment_id: 5 }
    );
    assert_eq!(transport.sent.borrow().len(), 1, "only the read, no POST");
}

#[test]
fn a_marker_written_by_a_non_member_does_not_stop_the_summary() {
    let target = SummaryTarget {
        project: project(),
        number: 42,
    };
    // NONE, a lower-case `member` and a missing association all fail closed.
    for association in ["NONE", "member", ""] {
        let (outcome, transport) = with_connector(
            vec![
                get(
                    &comments_url(42),
                    json_response(200, json!([comment(5, association, marker())])),
                ),
                exchange(
                    "POST",
                    &format!("{BASE}/issues/42/comments"),
                    json_response(201, json!({"id": 6})),
                ),
            ],
            Config::default(),
            |c| c.post_summary(&target, "summary", marker()),
        );
        assert_eq!(
            outcome.unwrap(),
            SummaryOutcome::Posted { comment_id: 6 },
            "{association:?}"
        );
        assert_eq!(transport.sent.borrow().len(), 2);
    }
}

#[test]
fn a_bad_marker_or_a_refused_post_is_an_error_and_never_a_posted_summary() {
    let target = SummaryTarget {
        project: project(),
        number: 42,
    };
    let too_long = marker_of_len(201);
    let mut bad_markers: Vec<&str> = vec![
        "",
        "no marker",
        "<!-- other -->",
        " <!-- moai-x -->",
        "<!-- moai-x",
        "<!-- moai-x -->\n",
        "<!-- moai-x -->\nsecond line",
        "<!-- moai-a\nb -->",
        "<!-- moai-a\tb -->",
        "<!-- moai-a\rb -->",
        "<!-- moai-x --> text <!-- moai-y -->",
    ];
    bad_markers.push(&too_long);
    for bad in bad_markers {
        let (outcome, transport) = with_connector(vec![], Config::default(), |c| {
            c.post_summary(&target, "summary", bad)
        });
        assert!(
            matches!(outcome, Err(ConnectorError::Invalid(_))),
            "{bad:?}"
        );
        assert!(transport.sent.borrow().is_empty(), "{bad:?}: no request");
    }
    let (outcome, _) = with_connector(
        vec![
            get(&comments_url(42), json_response(200, json!([]))),
            exchange(
                "POST",
                &format!("{BASE}/issues/42/comments"),
                json_response(422, json!({"message": "Validation Failed"})),
            ),
        ],
        Config::default(),
        |c| c.post_summary(&target, "summary", marker()),
    );
    assert!(matches!(outcome, Err(ConnectorError::Host(_))));
    // Only 201 is a created comment: an answer that carries an id with another status is not.
    let (outcome, _) = with_connector(
        vec![
            get(&comments_url(42), json_response(200, json!([]))),
            exchange(
                "POST",
                &format!("{BASE}/issues/42/comments"),
                json_response(200, json!({"id": 9})),
            ),
        ],
        Config::default(),
        |c| c.post_summary(&target, "summary", marker()),
    );
    assert!(
        matches!(outcome, Err(ConnectorError::Host(_))),
        "{outcome:?}"
    );
}

/// A marker of `len` bytes: `<!-- moai-` (10), filler, ` -->` (4).
fn marker_of_len(len: usize) -> String {
    format!("<!-- moai-{} -->", "x".repeat(len - 14))
}

/// The marker rules one at a time, each with a case that only that rule rejects.
#[test]
fn each_marker_rule_rejects_what_only_it_catches() {
    // Length: 200 bytes pass, 201 do not (the numbers are written out, not read from the code).
    assert_eq!(marker_of_len(200).len(), 200);
    assert!(check_marker(&marker_of_len(200)).is_ok());
    assert_eq!(marker_of_len(201).len(), 201);
    assert!(matches!(
        check_marker(&marker_of_len(201)),
        Err(ConnectorError::Invalid(_))
    ));
    // The shortest marker passes.
    assert!(check_marker("<!-- moai-a -->").is_ok());
    // Only the end rule: right start, no control character, short, no inner `-->`.
    assert!(matches!(
        check_marker("<!-- moai-x"),
        Err(ConnectorError::Invalid(_))
    ));
    // Only the control rule: it ends ` -->`, so a missing end check cannot be what rejects it.
    for with_control in [
        "<!-- moai-a\nb -->",
        "<!-- moai-a\tb -->",
        "<!-- moai-a\rb -->",
        "<!-- moai-a\u{7}b -->",
    ] {
        assert!(
            matches!(check_marker(with_control), Err(ConnectorError::Invalid(_))),
            "{with_control:?}"
        );
    }
    // Only the inner rule: a comment closed early.
    assert!(matches!(
        check_marker("<!-- moai-x --> text <!-- moai-y -->"),
        Err(ConnectorError::Invalid(_))
    ));
    // Only the start rule.
    assert!(matches!(
        check_marker("<!-- other-x -->"),
        Err(ConnectorError::Invalid(_))
    ));
}

#[test]
fn a_pr_is_open_merged_or_closed() {
    let cases = [
        (
            json!({"state": "open", "merged": false}),
            PrPhase::Open,
            None,
        ),
        (
            json!({"state": "closed", "merged": true, "merge_commit_sha": "m".repeat(40)}),
            PrPhase::Merged,
            Some("m".repeat(40)),
        ),
        (
            json!({"state": "closed", "merged": false, "merge_commit_sha": "n".repeat(40)}),
            PrPhase::Closed,
            None,
        ),
        (
            json!({"state": "closed", "merged_at": "2026-11-15T03:00:00Z", "merge_commit_sha": "o".repeat(40)}),
            PrPhase::Merged,
            Some("o".repeat(40)),
        ),
    ];
    for (mut pull, phase, merge_commit) in cases {
        pull["head"] = json!({"sha": "h".repeat(40)});
        let (state, _) = with_connector(
            vec![get(&format!("{BASE}/pulls/8"), json_response(200, pull))],
            Config::default(),
            |c| c.pr_state(&project(), 8),
        );
        let state = state.unwrap();
        assert_eq!(state.phase, phase);
        assert_eq!(state.merge_commit, merge_commit);
        assert_eq!(state.head, "h".repeat(40));
    }
    // A response that names no state is not guessed at; a missing PR is the host's 404.
    let (state, _) = with_connector(
        vec![get(
            &format!("{BASE}/pulls/8"),
            json_response(200, json!({"head": {"sha": "h".repeat(40)}})),
        )],
        Config::default(),
        |c| c.pr_state(&project(), 8),
    );
    assert!(matches!(state, Err(ConnectorError::Host(_))));
    let (state, _) = with_connector(
        vec![get(
            &format!("{BASE}/pulls/9"),
            json_response(404, json!({"message": "Not Found"})),
        )],
        Config::default(),
        |c| c.pr_state(&project(), 9),
    );
    match state {
        Err(ConnectorError::Host(e)) => assert_eq!(e.kind, "not_found"),
        other => panic!("{other:?}"),
    }
}

#[test]
fn verdicts_are_the_checks_and_the_newest_audit_status_of_the_current_head() {
    let head = "a".repeat(40);
    let runs = format!("{BASE}/commits/{head}/check-runs?per_page=100");
    let statuses = format!("{BASE}/commits/{head}/statuses?per_page=100");
    let pull = json!({"state": "open", "merged": false, "head": {"sha": head}});
    let (verdicts, transport) = with_connector(
        vec![
            get(&format!("{BASE}/pulls/8"), json_response(200, pull.clone())),
            get(
                &runs,
                json_response(
                    200,
                    json!({"check_runs": [{"name": "rc-ci", "status": "completed", "conclusion": "success"}]}),
                )
                .with_header("link", next_page(&runs)),
            ),
            get(
                &format!("{runs}&page=2"),
                json_response(
                    200,
                    json!({"check_runs": [{"name": "lint", "status": "in_progress", "conclusion": null}]}),
                ),
            ),
            get(
                &statuses,
                json_response(
                    200,
                    json!([
                        {"context": "other", "state": "success", "description": "x"},
                        {"context": "moai/audit", "state": "failure", "description": "newest"},
                        {"context": "moai/audit", "state": "success", "description": "older"},
                    ]),
                ),
            ),
        ],
        Config::default(),
        |c| c.verdicts(&project(), 8),
    );
    let verdicts: Verdicts = verdicts.unwrap();
    assert_eq!(transport.remaining(), 0);
    assert_eq!(verdicts.head, head);
    let checks: Vec<(&str, Option<&str>)> = verdicts
        .checks
        .iter()
        .map(|c| (c.name.as_str(), c.conclusion.as_deref()))
        .collect();
    assert_eq!(checks, [("rc-ci", Some("success")), ("lint", None)]);
    let audit = verdicts.audit.unwrap();
    assert_eq!(
        (audit.state.as_str(), audit.description.as_str()),
        ("failure", "newest")
    );

    // No audit status on the head: none, not a guess.
    let (verdicts, _) = with_connector(
        vec![
            get(&format!("{BASE}/pulls/8"), json_response(200, pull)),
            get(&runs, json_response(200, json!({"check_runs": []}))),
            get(&statuses, json_response(200, json!([]))),
        ],
        Config::default(),
        |c| c.verdicts(&project(), 8),
    );
    let verdicts = verdicts.unwrap();
    assert!(verdicts.checks.is_empty());
    assert!(verdicts.audit.is_none());
}

fn issue_for_fake(n: u64) -> HostIssue {
    HostIssue {
        number: n,
        title: format!("t{n}"),
        body: String::new(),
        labels: vec![],
        author: "a".into(),
        association: "OWNER".into(),
        created_at: String::new(),
        updated_at: String::new(),
        comments: vec![HostComment {
            id: n,
            author: "b".into(),
            association: "NONE".into(),
            body: String::new(),
            updated_at: String::new(),
        }],
    }
}

#[test]
fn the_fake_connector_records_calls_and_posts_a_marker_once() {
    let fake = FakeConnector::new(HostKind::Github)
        .with_issues(vec![issue_for_fake(1), issue_for_fake(2)])
        .with_pr(PrState {
            number: 8,
            phase: PrPhase::Merged,
            head: "h".repeat(40),
            merge_commit: Some("m".repeat(40)),
        });
    let p = project();
    assert_eq!(fake.open_issues(&p).unwrap().issues.len(), 2);
    let target = SummaryTarget {
        project: p.clone(),
        number: 3,
    };
    assert!(matches!(
        fake.post_summary(&target, "one", marker()).unwrap(),
        SummaryOutcome::Posted { .. }
    ));
    assert!(matches!(
        fake.post_summary(&target, "two", marker()).unwrap(),
        SummaryOutcome::AlreadyPosted { .. }
    ));
    assert_eq!(fake.pr_state(&p, 8).unwrap().phase, PrPhase::Merged);
    assert!(fake.pr_state(&p, 9).is_err());
    assert_eq!(fake.posted().len(), 1, "the marker went out once");
    assert_eq!(
        fake.calls(),
        vec![
            Call::OpenIssues {
                project: "modu-ai/moai-adk".into()
            },
            Call::PostSummary {
                project: "modu-ai/moai-adk".into(),
                number: 3,
                marker: marker().into()
            },
            Call::PostSummary {
                project: "modu-ai/moai-adk".into(),
                number: 3,
                marker: marker().into()
            },
            Call::PrState {
                project: "modu-ai/moai-adk".into(),
                number: 8
            },
            Call::PrState {
                project: "modu-ai/moai-adk".into(),
                number: 9
            },
        ]
    );

    let cut = FakeConnector::new(HostKind::Github).with_incomplete("page limit exceeded");
    assert!(matches!(
        cut.open_issues(&p),
        Err(ConnectorError::Incomplete { .. })
    ));
}

/// AC10 for this slice: the connector takes its transport and clock from the caller, so it has
/// no way to start a process, read the environment or reach the network by itself.
#[test]
fn the_connector_source_starts_no_process_and_reads_no_environment() {
    let source = include_str!("../src/forge/connector.rs");
    let transport = format!("{}{}", "Gh", "Transport");
    for needle in [
        transport.as_str(),
        "std::process",
        "Command::new",
        "std::env",
        "env::var",
        "home_dir",
    ] {
        assert!(!source.contains(needle), "connector.rs mentions {needle}");
    }
}
