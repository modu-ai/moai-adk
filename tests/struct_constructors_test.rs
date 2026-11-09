//! D9 #2256: constructors for the shared request/record structs. A constructor takes the
//! required fields and fills every optional field with its default, so adding another
//! optional field later does not break call sites (E0063, design report §2-§3). Each
//! case pins known fields to independent literal expectations (or the serde form with
//! only the required keys), while leaving future optional fields out of the pattern.

use moai_cli::claim::{ClaimRequest, Owner, SpecRef, TaskRef, WorktreeId};
use moai_cli::forge::transport::{GhTransport, HttpRequest, HttpResponse};
use serde_json::json;
use std::time::Duration;

#[test]
fn http_request_new_has_no_headers_and_no_body() {
    let req = HttpRequest::new("GET", "https://api.github.com/repos/o/r");
    let HttpRequest {
        method,
        url,
        headers,
        body,
        ..
    } = &req;
    assert_eq!(method, "GET");
    assert_eq!(url, "https://api.github.com/repos/o/r");
    assert_eq!(headers.len(), 0);
    assert_eq!(body, &None);
    let req = req
        .with_header("If-None-Match", "\"etag\"")
        .with_body(json!({"a": 1}));
    assert_eq!(
        req.headers.get("If-None-Match").map(String::as_str),
        Some("\"etag\"")
    );
    assert_eq!(req.body, Some(json!({"a": 1})));
}

#[test]
fn http_response_new_has_no_headers_and_no_body() {
    let resp = HttpResponse::new(200);
    let HttpResponse {
        status,
        headers,
        body,
        ..
    } = &resp;
    assert_eq!(*status, 200);
    assert_eq!(headers.len(), 0);
    assert_eq!(body, &None);
    let resp = resp.with_header("etag", "\"x\"").with_body(json!([1, 2]));
    assert_eq!(resp.header("ETag"), Some("\"x\""));
    assert_eq!(resp.body, Some(json!([1, 2])));
}

#[test]
fn gh_transport_new_uses_the_default_timeout_and_no_token() {
    let t = GhTransport::new("gh");
    assert_eq!(t.bin, "gh");
    assert_eq!(t.timeout, Duration::from_secs(60));
    assert!(t.gh_token.is_none());
    let t = t
        .timeout(Duration::from_secs(5))
        .gh_token(Some("fixture-token".into()));
    assert_eq!(t.timeout, Duration::from_secs(5));
    assert_eq!(t.gh_token.as_deref(), Some("fixture-token"));
}

fn task_ref() -> TaskRef {
    TaskRef::new("github", "github.com", "acme", "widgets", 101, 7001, 7)
}

#[test]
fn task_ref_new_sets_every_field() {
    let TaskRef {
        provider,
        host,
        owner,
        repo,
        repository_id,
        issue_id,
        issue_number,
        ..
    } = task_ref();
    assert_eq!(provider, "github");
    assert_eq!(host, "github.com");
    assert_eq!(owner, "acme");
    assert_eq!(repo, "widgets");
    assert_eq!(repository_id, 101);
    assert_eq!(issue_id, 7001);
    assert_eq!(issue_number, 7);
}

#[test]
fn owner_new_leaves_the_optional_fields_empty() {
    let o = Owner::new("session", "s-1");
    let Owner {
        kind,
        id,
        host_id,
        pid,
        process_start,
        ..
    } = &o;
    assert_eq!(kind, "session");
    assert_eq!(id, "s-1");
    assert_eq!(host_id, &None);
    assert_eq!(pid, &None);
    assert_eq!(process_start, &None);
    let o = o
        .host_id("h")
        .pid(42)
        .process_start("Mon Nov  9 00:00:00 2026");
    assert_eq!(o.host_id.as_deref(), Some("h"));
    assert_eq!(o.pid, Some(42));
    assert_eq!(o.process_start.as_deref(), Some("Mon Nov  9 00:00:00 2026"));
}

#[test]
fn worktree_id_new_has_no_branch() {
    let w = WorktreeId::new("/r/.git", "/r/.git/worktrees/a", "/r/a", "a".repeat(40));
    let WorktreeId {
        common_dir,
        git_dir,
        path,
        head,
        branch,
        ..
    } = &w;
    assert_eq!(common_dir, "/r/.git");
    assert_eq!(git_dir, "/r/.git/worktrees/a");
    assert_eq!(path, "/r/a");
    assert_eq!(head, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa");
    assert_eq!(branch, &None);
    assert_eq!(w.branch("issue/1-x").branch.as_deref(), Some("issue/1-x"));
}

fn spec_ref() -> SpecRef {
    SpecRef {
        id: "SPEC-1".into(),
        revision: "r1".into(),
        digest: None,
    }
}

#[test]
fn claim_request_new_equals_the_serde_form_with_only_required_keys() {
    let built = ClaimRequest::new(
        1,
        task_ref(),
        "issue/7",
        Owner::new("session", "s-1"),
        "b".repeat(40),
        vec!["src/".into()],
        spec_ref(),
    );
    let parsed: ClaimRequest = serde_json::from_value(json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com", "owner": "acme", "repo": "widgets",
            "repository_id": 101, "issue_id": 7001, "issue_number": 7
        },
        "logical_key": "issue/7",
        "owner": {"kind": "session", "id": "s-1"},
        "base_revision": "b".repeat(40),
        "write_scope": ["src/"],
        "spec_ref": {"id": "SPEC-1", "revision": "r1"}
    }))
    .unwrap();
    assert_eq!(built, parsed);

    let full = built
        .assignee("octo")
        .provider_session("ps-1")
        .attempt(2)
        .lease_until("2026-11-09T16:00:00Z")
        .projection_intents(vec![json!({"kind": "comment"})])
        .expected_epoch(3);
    assert_eq!(full.assignee.as_deref(), Some("octo"));
    assert_eq!(full.provider_session.as_deref(), Some("ps-1"));
    assert_eq!(full.attempt, Some(2));
    assert_eq!(full.lease_until, "2026-11-09T16:00:00Z");
    assert_eq!(full.projection_intents, vec![json!({"kind": "comment"})]);
    assert_eq!(full.expected_epoch, Some(3));
}
