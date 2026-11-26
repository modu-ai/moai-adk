//! #2740: `labels_check::check` reads open Issues through a recorded transport — GET only,
//! every page, pull requests skipped — and reports each Issue that breaks the axis rule of
//! this repository's `.github/moai/labels.yml`. Nothing is written.

use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::labels_check::{check, parse_axes};
use serde_json::{json, Value};

const API: &str = "https://api.github.com/repos/o/r";

fn issue(number: u64, labels: &[&str]) -> Value {
    json!({"number": number, "labels": labels.iter().map(|l| json!({"name": l})).collect::<Vec<_>>()})
}

fn axes() -> moai_cli::labels_check::Axes {
    let file = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(".github/moai/labels.yml");
    parse_axes(&std::fs::read_to_string(file).unwrap()).unwrap()
}

#[test]
fn every_page_is_read_with_get_only_and_pull_requests_are_skipped() {
    let page2 = format!("{API}/issues?state=open&per_page=100&page=2");
    let t = RecordedTransport::new(vec![
        Exchange {
            request: HttpRequest::new("GET", format!("{API}/issues?state=open&per_page=100")),
            response: HttpResponse::new(200)
                .with_header("link", format!("<{page2}>; rel=\"next\""))
                .with_body(json!([
                    issue(10, &["type:fix", "priority:P1", "area:cli"]),
                    issue(11, &["type:feature", "area:ci"]),
                    // A pull request carries no axis rule here.
                    {"number": 12, "labels": [], "pull_request": {"url": "x"}},
                ])),
        },
        Exchange {
            request: HttpRequest::new("GET", page2),
            response: HttpResponse::new(200).with_body(json!([issue(
                13,
                &[
                    "type:security",
                    "priority:P2",
                    "area:cli",
                    "area:ci",
                    "area:hooks"
                ]
            ),])),
        },
    ]);
    let r = check(&t, "o", "r", &axes()).unwrap();
    assert_eq!(t.remaining(), 0, "both pages read");
    let sent = t.sent.borrow();
    assert!(
        sent.iter().all(|q| q.method == "GET"),
        "write calls: {sent:?}"
    );
    assert_eq!(r["open_issues_read"], 3);
    assert_eq!(r["violating"], 2);
    let numbers: Vec<u64> = r["issues"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| i["number"].as_u64().unwrap())
        .collect();
    assert_eq!(numbers, [13, 11]);
    assert_eq!(
        r["issues"][1]["violations"],
        json!([{"kind": "priority_count", "count": 0}])
    );
    assert_eq!(
        r["issues"][0]["violations"],
        json!([
            {"kind": "type_count", "count": 0},
            {"kind": "area_too_many", "count": 3},
            {"kind": "not_used", "label": "type:security"}
        ])
    );
    assert_eq!(
        r["by_kind"],
        json!({"area_too_many": 1, "not_used": 1, "priority_count": 1, "type_count": 1})
    );
}

/// `n` pages; page `k` points to page `k + 1` while `k < n`, and the last one points to
/// `last_next` when given.
fn pages(n: usize, last_next: Option<String>) -> RecordedTransport {
    let url = |k: usize| {
        if k == 1 {
            format!("{API}/issues?state=open&per_page=100")
        } else {
            format!("{API}/issues?state=open&per_page=100&page={k}")
        }
    };
    let ex = (1..=n)
        .map(|k| {
            let next = if k < n {
                Some(url(k + 1))
            } else {
                last_next.clone()
            };
            let mut resp = HttpResponse::new(200).with_body(json!([issue(
                k as u64,
                &["type:fix", "priority:P1", "area:cli"]
            )]));
            if let Some(next) = next {
                resp = resp.with_header("link", format!("<{next}>; rel=\"next\""));
            }
            Exchange {
                request: HttpRequest::new("GET", url(k)),
                response: resp,
            }
        })
        .collect();
    RecordedTransport::new(ex)
}

/// #2743 audit B1: the page cap is the only stop for an endless `next` chain. Thirty pages are
/// read to the end; a thirty-first is refused before it is requested, and a page that points
/// to itself ends the same way.
#[test]
fn thirty_pages_are_read_and_a_thirty_first_is_refused_before_it_is_sent() {
    let t = pages(30, None);
    let r = check(&t, "o", "r", &axes()).unwrap();
    assert_eq!(r["open_issues_read"], 30);
    assert_eq!(t.sent.borrow().len(), 30);

    let t = pages(
        30,
        Some(format!("{API}/issues?state=open&per_page=100&page=31")),
    );
    let e = check(&t, "o", "r", &axes()).unwrap_err();
    assert!(e.contains("more than 30 issue pages"), "{e}");
    assert_eq!(
        t.sent.borrow().len(),
        30,
        "the 31st page is never requested"
    );

    // A page whose `next` is itself, answered again and again.
    let first = format!("{API}/issues?state=open&per_page=100");
    let ex = (0..31)
        .map(|_| Exchange {
            request: HttpRequest::new("GET", first.clone()),
            response: HttpResponse::new(200)
                .with_header("link", format!("<{first}>; rel=\"next\""))
                .with_body(json!([])),
        })
        .collect();
    let t = RecordedTransport::new(ex);
    let e = check(&t, "o", "r", &axes()).unwrap_err();
    assert!(e.contains("more than 30 issue pages"), "{e}");
    assert_eq!(t.sent.borrow().len(), 30);
}

/// #2743 audit N1: only plain GitHub names reach the URL path.
#[test]
fn owner_and_repo_must_be_plain_names() {
    for (owner, repo) in [
        ("a/b/../../user", "r"),
        ("o", "r/../x"),
        ("", "r"),
        ("o", "r?x=1"),
        ("..", "r"),
    ] {
        let t = RecordedTransport::new(vec![]);
        let e = check(&t, owner, repo, &axes()).unwrap_err();
        assert!(e.contains("GitHub names"), "{owner}/{repo}: {e}");
        assert!(t.sent.borrow().is_empty());
    }
}

#[test]
fn a_failed_read_is_an_error_not_an_empty_report() {
    let t = RecordedTransport::new(vec![Exchange {
        request: HttpRequest::new("GET", format!("{API}/issues?state=open&per_page=100")),
        response: HttpResponse::new(404),
    }]);
    let e = check(&t, "o", "r", &axes()).unwrap_err();
    assert!(e.contains("HTTP 404"), "{e}");
}
