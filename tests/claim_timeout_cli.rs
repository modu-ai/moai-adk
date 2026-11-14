//! #2906 S1: `claim-timeout-check` reads GitHub with GET only and turns any unreadable
//! activity into `uncertain`. The library runs over recorded exchanges; the binary runs over a
//! stub `gh` in a temporary HOME. No live GitHub call is made.
mod common;

use common::CleanEnv;
use moai_cli::claim_timeout::{check, Target};
use moai_cli::forge::policy::{format_utc, parse_utc, Clock};
use moai_cli::forge::transport::{Exchange, HttpResponse, RecordedTransport};
use serde_json::{json, Value};
use std::process::Command;

/// A fixed past time: the binary refuses a `--now` more than 300 s after the system clock.
const NOW: &str = "2026-11-01T12:00:00Z";
const ISSUE_URL: &str = "https://api.github.com/repos/acme/widgets/issues/7/comments?per_page=100";
const LINE: &str = "moai-claim-handoff: comment=100 to=agent-25 by=timeout";

struct FixedClock(i64);
impl Clock for FixedClock {
    fn now(&self) -> i64 {
        self.0
    }
    fn sleep(&self, _: u64) {}
}

fn now() -> i64 {
    parse_utc(NOW).unwrap()
}

fn at(offset: i64) -> String {
    format_utc(now() + offset)
}

fn repo_url() -> String {
    format!(
        "https://api.github.com/repos/acme/widgets/issues/comments?since={}&per_page=100",
        at(-7200)
    )
}

fn commits_url() -> String {
    format!(
        "https://api.github.com/repos/acme/widgets/commits?sha=issue/7-x&since={}&per_page=100",
        at(-3 * 3600)
    )
}

fn raw(id: i64, body: &str, created: i64) -> Value {
    json!({"id": id, "user": {"login": "GoosLab"}, "body": body,
           "created_at": at(created), "updated_at": at(created)})
}

/// The Issue: the holder's claim (lease expired 10 s ago, written three hours ago) and the
/// request naming it an hour ago.
fn issue_body() -> Value {
    json!([
        raw(
            100,
            &format!(
                "moai-claim: session=agent-7-abc epoch=1 lease_until={} lane=agent-7 issue=7 branch=issue/7-x",
                at(-10)
            ),
            -3 * 3600
        ),
        raw(200, "agent-25: claim 100 shows no activity; timeout takeover after 1800 s", -3600),
    ])
}

/// A commit carrying the claim token, older than the heartbeat window.
fn commits_body() -> Value {
    json!([{"sha": "a1", "commit": {"message": "fix: x\n\nMoai-Claim: 7#100",
                                    "committer": {"date": at(-7300)}}}])
}

fn get(url: &str, status: u16, body: Value) -> Exchange {
    common::exchange("GET", url, common::json_response(status, body))
}

fn target() -> Target<'static> {
    Target {
        owner: "acme",
        repo: "widgets",
        issue: 7,
        claim: 100,
        request: 200,
        me: "agent-25",
    }
}

fn run(exchanges: Vec<Exchange>) -> (Value, RecordedTransport) {
    let t = RecordedTransport::new(exchanges);
    let v = check(&target(), &t, &FixedClock(now()), now());
    (v, t)
}

/// AC12: a failed read or a page cap overflow is `uncertain` with no line; a branch that no
/// longer exists (404) is zero commits, not a failure (decision Q4).
#[test]
fn unreadable_activity_is_uncertain() {
    let ok_issue = || get(ISSUE_URL, 200, issue_body());
    for (case, exchanges) in [
        (
            "repository comments",
            vec![
                ok_issue(),
                get(&repo_url(), 500, json!({"message": "boom"})),
            ],
        ),
        (
            "branch commits",
            vec![
                ok_issue(),
                get(&repo_url(), 200, json!([])),
                get(&commits_url(), 500, json!({"message": "boom"})),
            ],
        ),
    ] {
        let (v, _) = run(exchanges);
        assert_eq!(v["state"], json!("uncertain"), "{case}: {v}");
        assert_eq!(v["met"], json!(false), "{case}: {v}");
        assert!(v["line"].is_null(), "{case}: {v}");
        assert!(
            v["reason"].as_str().unwrap_or("").contains(case),
            "{case}: {v}"
        );
    }

    // Ten pages of Issue comments that still name a next page: the cap stops the read.
    let pages: Vec<Exchange> = (1..=10)
        .map(|n| {
            let url = if n == 1 {
                ISSUE_URL.to_string()
            } else {
                format!("{ISSUE_URL}&page={n}")
            };
            Exchange {
                request: moai_cli::forge::transport::HttpRequest::new("GET", &url),
                response: HttpResponse::new(200).with_body(json!([])).with_header(
                    "link",
                    format!("<{ISSUE_URL}&page={}>; rel=\"next\"", n + 1),
                ),
            }
        })
        .collect();
    let (v, t) = run(pages);
    assert_eq!(v["state"], json!("uncertain"), "{v}");
    assert!(v["reason"].as_str().unwrap().contains("page limit"), "{v}");
    assert_eq!(t.sent.borrow().len(), 10, "{v}");

    // A deleted claim branch: zero commits, the judgement still runs.
    let (v, _) = run(vec![
        ok_issue(),
        get(&repo_url(), 200, json!([])),
        get(&commits_url(), 404, json!({"message": "Not Found"})),
    ]);
    assert_eq!(v["state"], json!("met"), "{v}");
    assert_eq!(v["activity_read"]["branch_found"], json!(false), "{v}");
}

/// AC12: the whole check sends three GETs and nothing else.
#[test]
fn only_get_requests_are_sent() {
    let (v, t) = run(vec![
        get(ISSUE_URL, 200, issue_body()),
        get(
            &repo_url(),
            200,
            json!([raw(150, "agent-9: unrelated", -60)]),
        ),
        get(&commits_url(), 200, commits_body()),
    ]);
    assert_eq!(v["met"], json!(true), "{v}");
    assert_eq!(v["line"], json!(LINE));
    assert_eq!(v["activity_read"]["token_commits"], json!(1), "{v}");
    let sent = t.sent.borrow();
    assert_eq!(sent.len(), 3);
    assert!(sent.iter().all(|r| r.method == "GET"), "{sent:?}");
    assert_eq!(t.remaining(), 0);
}

/// AC13 end to end: the binary over a stub `gh` prints the envelope with `met` and the line,
/// exits 0, and a judgement that is not met is also exit 0 with `met: false`.
#[cfg(unix)]
#[test]
fn the_binary_prints_the_timeout_line_when_met() {
    let home = tempfile::TempDir::new().unwrap();
    let gh = home.path().join("gh-stub");
    common::write_executable(
        &gh,
        "#!/bin/sh\nfor a; do url=$a; done\ncase \"$url\" in\n\
         */issues/7/comments*) body=$ISSUE_JSON ;;\n\
         */issues/comments*) body='[]' ;;\n\
         */commits*) printf 'HTTP/2.0 404 Not Found\\r\\n\\r\\n{\"message\":\"Not Found\"}'; exit 1 ;;\n\
         *) exit 9 ;;\nesac\nprintf 'HTTP/2.0 200 OK\\r\\n\\r\\n%s' \"$body\"\n",
    );
    let run = |now: &str| {
        Command::new(common::MOAI_BIN)
            .args([
                "claim-timeout-check",
                "--issue",
                "7",
                "--claim",
                "100",
                "--request",
                "200",
                "--lane",
                "agent-25",
                "--repo",
                "acme/widgets",
                "--now",
                now,
            ])
            .clean_env()
            .env("HOME", home.path())
            .env("PATH", "/usr/bin:/bin")
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_SYSTEM", "/dev/null")
            .env("MOAI_FORGE_GH_BIN", &gh)
            .env("ISSUE_JSON", issue_body().to_string())
            .current_dir(home.path())
            .output()
            .unwrap()
    };
    let out = run(NOW);
    let envelope: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(out.status.code(), Some(0), "{envelope}");
    assert_eq!(envelope["operation"], json!("claim.timeout_check"));
    assert_eq!(envelope["result"]["met"], json!(true), "{envelope}");
    assert_eq!(envelope["result"]["line"], json!(LINE));

    // 20 minutes earlier the lease (expired 10 s before NOW) is still live.
    let out = run(&at(-1200));
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(out.status.code(), Some(0), "{envelope}");
    assert_eq!(envelope["result"]["met"], json!(false), "{envelope}");
    assert!(envelope["result"]["line"].is_null(), "{envelope}");

    // N3: a `--now` later than the system clock by more than 300 s is a usage error, before
    // any read.
    let out = run("2999-01-01T00:00:00Z");
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert_ne!(out.status.code(), Some(0), "{text}");
    assert!(text.contains("--now"), "{text}");
    assert!(text.contains("system clock"), "{text}");
    assert!(!text.contains("\"met\""), "{text}");
}
