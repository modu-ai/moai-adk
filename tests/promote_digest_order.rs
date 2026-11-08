//! Issue #2074: the 07-B `preview_digest` must not depend on the key order of JSON objects
//! it reads. serde_json is built with `preserve_order` (#1992), so a `Value` keeps whatever
//! order its source had. Each case feeds the same meaning in two key orders and compares the
//! digests. No network: responses are recorded exchanges.
use moai_cli::forge::{
    policy::Clock,
    promote::{self, PromoteRequest},
    transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport},
    Client, Config,
};
use serde_json::Value;
use std::{fs, path::Path};

const API: &str = "https://api.github.com/repos/acme/widgets";
const LIST: &str = "https://api.github.com/repos/acme/widgets/issues?state=all&sort=updated&direction=asc&per_page=100";
const TASKS: &str = include_str!("fixtures/forge/promote/tasks_three.md");

struct FixedClock;
impl Clock for FixedClock {
    fn now(&self) -> i64 {
        1_794_000_000
    }
    fn sleep(&self, _: u64) {}
}

fn get(url: &str, body: &str) -> Exchange {
    Exchange {
        request: HttpRequest {
            method: "GET".into(),
            url: url.into(),
            headers: Default::default(),
            body: None,
        },
        response: HttpResponse {
            status: 200,
            headers: [("etag".to_string(), "\"e1\"".to_string())].into(),
            body: Some(serde_json::from_str(body).unwrap()),
        },
    }
}

fn digest(project: &Path, request: &str, repo: &str, issues: &str) -> String {
    let req: PromoteRequest = serde_json::from_str(request).unwrap();
    let t = RecordedTransport::new(vec![get(API, repo), get(LIST, issues)]);
    let client = Client::new(&t, &FixedClock, Config::default(), Default::default());
    let pv = promote::preview(project, &req, &client).unwrap();
    assert_eq!(t.remaining(), 0);
    pv.digest
}

fn project() -> tempfile::TempDir {
    let tmp = tempfile::tempdir().unwrap();
    fs::write(tmp.path().join("tasks.md"), TASKS).unwrap();
    tmp
}

const REPO_A: &str = r#"{"id": 101, "visibility": "private", "private": true}"#;
const REPO_B: &str = r#"{"private": true, "visibility": "private", "id": 101}"#;
const ISSUES_A: &str = r#"[{"id": 9007, "node_id": "I_7", "number": 7, "title": "Something else",
    "body": "x", "html_url": "https://github.com/acme/widgets/issues/7",
    "updated_at": "2026-11-08T00:00:00Z", "state": "open"}]"#;
const ISSUES_B: &str = r#"[{"state": "open", "updated_at": "2026-11-08T00:00:00Z",
    "html_url": "https://github.com/acme/widgets/issues/7", "body": "x",
    "title": "Something else", "number": 7, "node_id": "I_7", "id": 9007}]"#;
const REQ_A: &str = r#"{"schema_version": 1, "tasks_path": "tasks.md",
    "target": {"owner": "acme", "repo": "widgets"}, "selected": ["T-1", "T-4"]}"#;
const REQ_B: &str = r#"{"selected": ["T-1", "T-4"], "target": {"repo": "widgets", "owner": "acme"},
    "tasks_path": "tasks.md", "schema_version": 1}"#;

#[test]
fn preview_digest_ignores_request_key_order() {
    let p = project();
    let a = digest(p.path(), REQ_A, REPO_A, ISSUES_A);
    let b = digest(p.path(), REQ_B, REPO_A, ISSUES_A);
    assert_eq!(a, b);
}

#[test]
fn preview_digest_ignores_response_key_order() {
    let p = project();
    let a = digest(p.path(), REQ_A, REPO_A, ISSUES_A);
    let b = digest(p.path(), REQ_A, REPO_B, ISSUES_B);
    assert_eq!(a, b);
}

#[test]
fn preview_digest_is_repeatable() {
    let p = project();
    let a = digest(p.path(), REQ_A, REPO_A, ISSUES_A);
    let b = digest(p.path(), REQ_A, REPO_A, ISSUES_A);
    assert_eq!(a, b);
    // Sanity: the digest does see a change in meaning.
    let other: Value = serde_json::from_str(ISSUES_A).unwrap();
    let mut other = other;
    other[0]["title"] = Value::from("Add login page");
    let c = digest(p.path(), REQ_A, REPO_A, &other.to_string());
    assert_ne!(a, c, "a duplicate title must change the preview");
}
