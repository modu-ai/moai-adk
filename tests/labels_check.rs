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

#[test]
fn a_failed_read_is_an_error_not_an_empty_report() {
    let t = RecordedTransport::new(vec![Exchange {
        request: HttpRequest::new("GET", format!("{API}/issues?state=open&per_page=100")),
        response: HttpResponse::new(404),
    }]);
    let e = check(&t, "o", "r", &axes()).unwrap_err();
    assert!(e.contains("HTTP 404"), "{e}");
}
