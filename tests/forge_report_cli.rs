//! 07-C (#2010) concise Issue report. Library tests drive `report::publish` through a scripted
//! transport with an injected clock: no network, no sleeping. The CLI tests point
//! `forge-report` at a stub `gh` that refuses writes, so they cover only failures that must
//! stop before any request. Texts are synthetic; credential-shaped values are assembled at
//! run time so no such literal sits in the repository.
use moai_cli::{
    contract,
    forge::{
        outbox::{self, OutboxState},
        policy::Clock,
        report::{self, ReportPayload},
        transport::{HttpRequest, HttpResponse, Transport},
        Client, ForgeError,
    },
};
mod common;

use serde_json::{json, Value};
use std::{
    cell::{Cell, RefCell},
    collections::VecDeque,
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

const REPO: &str = "https://api.github.com/repos/acme/widgets";

struct FakeClock(Cell<i64>);
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.0.get()
    }
    fn sleep(&self, secs: u64) {
        self.0.set(self.0.get() + secs as i64);
    }
}
fn clock() -> FakeClock {
    FakeClock(Cell::new(1_794_000_000))
}

enum Reply {
    Resp(HttpResponse),
    /// No HTTP response (timeout, connection loss).
    Fail(&'static str),
    /// A created or edited comment echoing the body that was sent.
    Echo(u16, i64),
    /// A GET of the comment last echoed.
    Last,
}

struct Script {
    steps: RefCell<VecDeque<(&'static str, String, Reply)>>,
    sent: RefCell<Vec<HttpRequest>>,
    last: RefCell<Value>,
}
impl Script {
    fn new(steps: Vec<(&'static str, String, Reply)>) -> Self {
        Script {
            steps: RefCell::new(steps.into()),
            sent: RefCell::new(Vec::new()),
            last: RefCell::new(Value::Null),
        }
    }
    fn writes(&self) -> Vec<String> {
        self.sent
            .borrow()
            .iter()
            .filter(|r| r.method != "GET")
            .map(|r| r.method.clone())
            .collect()
    }
}
impl Transport for Script {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse, String> {
        self.sent.borrow_mut().push(req.clone());
        let (method, url, reply) = self
            .steps
            .borrow_mut()
            .pop_front()
            .ok_or_else(|| format!("unexpected request {} {}", req.method, req.url))?;
        if method != req.method || url != req.url {
            return Err(format!(
                "expected {method} {url} but got {} {}",
                req.method, req.url
            ));
        }
        match reply {
            Reply::Resp(r) => Ok(r),
            Reply::Fail(why) => Err(why.into()),
            Reply::Echo(status, id) => {
                let c = comment(id, req.body.as_ref().unwrap()["body"].as_str().unwrap());
                *self.last.borrow_mut() = c.clone();
                Ok(HttpResponse::new(status).with_body(c))
            }
            Reply::Last => Ok(ok(self.last.borrow().clone(), None)),
        }
    }
}

fn ok(body: Value, link: Option<&str>) -> HttpResponse {
    let resp = HttpResponse::new(200)
        .with_header("etag", "\"e\"")
        .with_body(body);
    match link {
        Some(l) => resp.with_header("link", l),
        None => resp,
    }
}
fn comment(id: i64, body: &str) -> Value {
    json!({"id": id, "node_id": format!("IC_{id}"), "body": body, "updated_at": "2026-11-08T10:00:00Z",
           "html_url": format!("https://github.com/acme/widgets/issues/7#issuecomment-{id}")})
}
fn fixture(name: &str) -> Value {
    let path = format!(
        "{}/tests/fixtures/forge/report/{name}",
        env!("CARGO_MANIFEST_DIR")
    );
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}
fn page_url(n: u32) -> String {
    match n {
        1 => format!("{REPO}/issues/7/comments?per_page=100"),
        _ => format!(
            "https://api.github.com/repositories/101/issues/7/comments?per_page=100&page={n}"
        ),
    }
}
/// The three fixture pages, with `extra` comments appended to the last page.
fn reads(extra: Vec<Value>) -> Vec<(&'static str, String, Reply)> {
    let link = |n: u32| {
        format!(
            "<{}>; rel=\"next\", <{}>; rel=\"last\"",
            page_url(n),
            page_url(3)
        )
    };
    let mut p3 = fixture("comments_p3.json")["comments"]
        .as_array()
        .unwrap()
        .clone();
    p3.extend(extra);
    vec![
        (
            "GET",
            REPO.into(),
            Reply::Resp(ok(json!({"id": 101}), None)),
        ),
        (
            "GET",
            format!("{REPO}/issues/7"),
            Reply::Resp(ok(
                json!({"id": 7001, "node_id": "I_7001", "number": 7}),
                None,
            )),
        ),
        (
            "GET",
            page_url(1),
            Reply::Resp(ok(
                fixture("comments_p1.json")["comments"].clone(),
                Some(&link(2)),
            )),
        ),
        (
            "GET",
            page_url(2),
            Reply::Resp(ok(
                fixture("comments_p2.json")["comments"].clone(),
                Some(&link(3)),
            )),
        ),
        ("GET", page_url(3), Reply::Resp(ok(Value::Array(p3), None))),
    ]
}
fn with(
    mut steps: Vec<(&'static str, String, Reply)>,
    more: Vec<(&'static str, String, Reply)>,
) -> Vec<(&'static str, String, Reply)> {
    steps.extend(more);
    steps
}
fn payload() -> ReportPayload {
    serde_json::from_value(fixture("payload_not_run.json")).unwrap()
}
fn new_project() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let p = tmp.path().join("p");
    fs::create_dir(&p).unwrap();
    (tmp, p)
}
fn run(project: &Path, p: &ReportPayload, t: &Script) -> Result<Value, ForgeError> {
    let c = clock();
    let client = Client::new(t, &c, Default::default(), Default::default());
    let out = report::publish(project, p, "ref-2010", &client);
    assert_eq!(t.steps.borrow().len(), 0, "every scripted request was made");
    out
}
fn entry(project: &Path, attempt: u32) -> Option<outbox::OutboxEntry> {
    outbox::load(
        project,
        &outbox::op_id("report", "acme.widgets", "7", attempt).unwrap(),
    )
    .unwrap()
}
/// The error message classifies to the kind the library reports (01-C envelope agreement).
fn kind_of(e: &ForgeError) -> &'static str {
    assert_eq!(
        contract::classify(&e.message).as_str(),
        e.kind,
        "{}",
        e.message
    );
    e.kind
}
const POST: &str = "https://api.github.com/repos/acme/widgets/issues/7/comments";
fn patch(id: i64) -> String {
    format!("{REPO}/issues/comments/{id}")
}

#[test]
fn report_updates_single_marker_comment() {
    let (_tmp, project) = new_project();
    let p = payload();
    let body = report::render(&p).unwrap();
    let mut lines = body.lines();
    assert_eq!(lines.next(), Some("<!-- moai:report run=rc1-07c-demo -->"));
    assert_eq!(lines.next(), Some("MoAI run: rc1-07c-demo"));
    for field in [
        "상태: review",
        "수행 내용: ",
        "시험 결과: ",
        "다음 작업: ",
        "근거: ",
        "상세 로그: ",
    ] {
        assert!(body.contains(field), "template field {field}");
    }
    assert!(!body.contains('\r'));

    // First run: no marker comment in three pages (one quotes another run's marker) → POST.
    let t = Script::new(with(
        reads(vec![]),
        vec![
            ("POST", POST.into(), Reply::Echo(201, 9001)),
            ("GET", patch(9001), Reply::Last),
        ],
    ));
    let out = run(&project, &p, &t).unwrap();
    assert_eq!(t.writes(), ["POST"]);
    assert_eq!(out["action"], "created");
    assert_eq!(out["comment"]["id"], 9001);
    assert_eq!(
        out["limits"],
        json!({"max_body_chars": 65535, "tail_lines": 20, "defaults": "unmeasured"})
    );
    let posted = t.last.borrow()["body"].as_str().unwrap().to_string();
    assert_eq!(posted, body);
    assert_eq!(entry(&project, 1).unwrap().state, OutboxState::Confirmed);

    // Second run, new content: the marker comment is found → one PATCH, no POST.
    let mut p2 = payload();
    p2.next = "Merge after audit".into();
    let t = Script::new(with(
        reads(vec![comment(9001, &posted)]),
        vec![
            ("PATCH", patch(9001), Reply::Echo(200, 9001)),
            ("GET", patch(9001), Reply::Last),
        ],
    ));
    let out = run(&project, &p2, &t).unwrap();
    assert_eq!(t.writes(), ["PATCH"]);
    assert_eq!(out["action"], "updated");
    assert!(t.last.borrow()["body"]
        .as_str()
        .unwrap()
        .contains("다음 작업: Merge after audit"));
    assert_eq!(entry(&project, 2).unwrap().state, OutboxState::Confirmed);

    // Third run, same content: nothing to write.
    let current = t.last.borrow()["body"].as_str().unwrap().to_string();
    let t = Script::new(reads(vec![comment(9001, &current)]));
    let out = run(&project, &p2, &t).unwrap();
    assert!(t.writes().is_empty());
    assert_eq!(out["action"], "unchanged");

    // Two marker comments: refuse to pick one.
    let t = Script::new(reads(vec![
        comment(9001, &current),
        comment(9002, &current),
    ]));
    let e = run(&project, &p2, &t).unwrap_err();
    assert_eq!(kind_of(&e), "conflict");
    assert!(t.writes().is_empty());

    // An ExternalRef that no longer matches the Issue stops before reading comments.
    let t = Script::new(vec![
        (
            "GET",
            REPO.into(),
            Reply::Resp(ok(json!({"id": 101}), None)),
        ),
        (
            "GET",
            format!("{REPO}/issues/7"),
            Reply::Resp(ok(
                json!({"id": 9999, "node_id": "I_other", "number": 7}),
                None,
            )),
        ),
    ]);
    let e = run(&project, &p2, &t).unwrap_err();
    assert_eq!(kind_of(&e), "conflict");
}

#[test]
fn timeout_then_reconcile_creates_no_duplicate() {
    let p = payload();
    let body = report::render(&p).unwrap();

    // The POST timed out but did land: the next run finds it and writes nothing new.
    let (_tmp, project) = new_project();
    let t = Script::new(with(
        reads(vec![]),
        vec![("POST", POST.into(), Reply::Fail("operation timed out"))],
    ));
    let e = run(&project, &p, &t).unwrap_err();
    assert_eq!(kind_of(&e), "transport");
    assert!(e.retryable);
    let first = entry(&project, 1).unwrap();
    assert_eq!(first.state, OutboxState::PublishUnknown);
    assert!(first.intent.is_some() && first.response.is_some() && first.observed.is_none());
    let t = Script::new(reads(vec![comment(9001, &body)]));
    let out = run(&project, &p, &t).unwrap();
    assert!(t.writes().is_empty(), "marker found: no second comment");
    assert_eq!(out["action"], "unchanged");
    assert_eq!(entry(&project, 1).unwrap().state, OutboxState::Confirmed);
    assert!(entry(&project, 2).is_none());

    // The POST timed out and did not land: the comments are read first, then one POST.
    let (_tmp2, project2) = new_project();
    let t = Script::new(with(
        reads(vec![]),
        vec![("POST", POST.into(), Reply::Resp(HttpResponse::new(502)))],
    ));
    assert_eq!(kind_of(&run(&project2, &p, &t).unwrap_err()), "transport");
    let t = Script::new(with(
        reads(vec![]),
        vec![
            ("POST", POST.into(), Reply::Echo(201, 9001)),
            ("GET", patch(9001), Reply::Last),
        ],
    ));
    let out = run(&project2, &p, &t).unwrap();
    let methods: Vec<String> = t.sent.borrow().iter().map(|r| r.method.clone()).collect();
    assert_eq!(
        methods,
        ["GET", "GET", "GET", "GET", "GET", "POST", "GET"],
        "re-read before the retry"
    );
    assert_eq!(out["action"], "created");
    assert_eq!(
        entry(&project2, 1).unwrap().state,
        OutboxState::PublishUnknown
    );
    assert_eq!(entry(&project2, 2).unwrap().state, OutboxState::Confirmed);

    // Afterwards exactly one marker comment exists: a third run sees it and stays put.
    let t = Script::new(reads(vec![comment(9001, &body)]));
    assert_eq!(run(&project2, &p, &t).unwrap()["action"], "unchanged");

    // PATCH timeouts (audit FAIL on c300973): run 1 confirms body D1, run 2's PATCH to D2
    // times out. The remote body decides what run 3 does.
    let mut p2 = payload();
    p2.next = "Merge after audit".into();
    let body2 = report::render(&p2).unwrap();
    let after_patch_timeout = || {
        let (tmp, project) = new_project();
        let t = Script::new(with(
            reads(vec![]),
            vec![
                ("POST", POST.into(), Reply::Echo(201, 9001)),
                ("GET", patch(9001), Reply::Last),
            ],
        ));
        run(&project, &p, &t).unwrap();
        let t = Script::new(with(
            reads(vec![comment(9001, &body)]),
            vec![("PATCH", patch(9001), Reply::Fail("operation timed out"))],
        ));
        assert_eq!(kind_of(&run(&project, &p2, &t).unwrap_err()), "transport");
        assert_eq!(
            entry(&project, 2).unwrap().state,
            OutboxState::PublishUnknown
        );
        (tmp, project)
    };

    // (a) Not landed: the remote still has D1, so the same comment is PATCHed again.
    let (_t3, project3) = after_patch_timeout();
    let t = Script::new(with(
        reads(vec![comment(9001, &body)]),
        vec![
            ("PATCH", patch(9001), Reply::Echo(200, 9001)),
            ("GET", patch(9001), Reply::Last),
        ],
    ));
    let out = run(&project3, &p2, &t).unwrap();
    assert_eq!(
        t.writes(),
        ["PATCH"],
        "retry after re-reading, never a new comment"
    );
    assert_eq!(out["action"], "updated");
    assert_eq!(t.last.borrow()["body"], json!(body2));
    let third = entry(&project3, 3).unwrap();
    assert_eq!(third.state, OutboxState::Confirmed);
    assert_eq!(
        third.intent.unwrap().detail["supersedes"],
        json!(outbox::op_id("report", "acme.widgets", "7", 2).unwrap())
    );

    // (b) Landed: the remote has D2, so the unknown write is confirmed and nothing is sent.
    let (_t4, project4) = after_patch_timeout();
    let t = Script::new(reads(vec![comment(9001, &body2)]));
    let out = run(&project4, &p2, &t).unwrap();
    assert!(t.writes().is_empty());
    assert_eq!(out["action"], "unchanged");
    assert_eq!(entry(&project4, 2).unwrap().state, OutboxState::Confirmed);
    assert!(entry(&project4, 3).is_none());

    // (c) Neither D1 nor D2: someone edited it in between; conflict, nothing written.
    let (_t5, project5) = after_patch_timeout();
    let t = Script::new(reads(vec![comment(9001, &format!("{body}\n\nedited"))]));
    assert_eq!(kind_of(&run(&project5, &p2, &t).unwrap_err()), "conflict");
    assert!(t.writes().is_empty());
}

#[test]
fn done_requires_all_required_checks_passed() {
    let mut p = payload();
    let review = report::render(&p).unwrap();
    assert!(
        review.contains("- live publish: not_run · exit none"),
        "{review}"
    );
    assert!(review.contains("- gap: live publish needs a user-designated test Issue"));

    p.status = report::Status::Done;
    let e = report::render(&p).unwrap_err();
    assert_eq!(kind_of(&e), "invalid_input");
    // `required` defaults to true; an optional not_run check does not block done.
    p.checks[1].required = false;
    let done = report::render(&p).unwrap();
    assert!(done.contains("상태: done") && done.contains("not_run · exit none · revision"));
    // A pass on another revision is not a pass on the candidate.
    p.checks[0].revision = "2222222222222222222222222222222222222222".into();
    assert_eq!(kind_of(&report::render(&p).unwrap_err()), "invalid_input");
    // No required check at all is not done either.
    p.checks.retain(|c| !c.required);
    assert_eq!(kind_of(&report::render(&p).unwrap_err()), "invalid_input");

    // Through the CLI the gate stops before gh is called; stdout is the envelope.
    let (tmp, project) = new_project();
    let gh = stub_gh(tmp.path());
    let mut v = fixture("payload_not_run.json");
    v["status"] = json!("done");
    let (out, _) = cli(&project, &gh, &v.to_string(), "ref-2010");
    assert_eq!(out["operation"], "forge.report");
    assert_eq!(out["error"]["kind"], "invalid_input");
    assert!(
        !tmp.path().join("gh-called").exists(),
        "no request before the gate"
    );
    assert!(!project.join(".moai-cli").exists());
    let (out, _) = cli(
        &project,
        &gh,
        &fixture("payload_not_run.json").to_string(),
        "",
    );
    assert_eq!(out["error"]["kind"], "permission_reference_missing");
}

#[test]
fn secrets_block_publication() {
    let token = format!("{}{}", "gh", "p_moaiTestDummyValue0000000000000000");
    let key = format!("{}{}", "AK", "IAZZZZZZZZZZZZZZZZ");
    let bearer = format!("{} {}", "Bearer", "moai.dummy.value");
    for place in 0..4 {
        let mut p = payload();
        match place {
            0 => p.work = format!("rotated {token}"),
            1 => p.checks[0].tail.push(format!("export AWS={key}")),
            2 => p.gaps.push(format!("Authorization: {bearer}")),
            _ => p.next = "clean /Users/someone/secret-dir".into(),
        }
        let e = report::render(&p).unwrap_err();
        assert_eq!(kind_of(&e), "blocked");
        assert!(e.message.contains("blocked_sensitive"));
        for v in [&token, &key, "moai.dummy.value", "/Users/someone"] {
            assert!(!e.message.contains(v), "value echoed");
        }
        let (_tmp, project) = new_project();
        let t = Script::new(vec![]);
        assert_eq!(kind_of(&run(&project, &p, &t).unwrap_err()), "blocked");
        assert!(t.sent.borrow().is_empty(), "nothing is sent");
    }
    // Control characters are removed, not rendered.
    let mut p = payload();
    p.work = "line one\nline\u{7} two".into();
    assert!(report::render(&p)
        .unwrap()
        .contains("수행 내용: line one line  two"));

    // Through the CLI: neither stdout nor stderr carries the value; gh is never started.
    let (tmp, project) = new_project();
    let gh = stub_gh(tmp.path());
    let mut v = fixture("payload_not_run.json");
    v["work"] = json!(format!("rotated {token}"));
    let (out, stderr) = cli(&project, &gh, &v.to_string(), "ref-2010");
    assert_eq!(out["error"]["kind"], "blocked");
    assert!(!out.to_string().contains(&token) && !stderr.contains(&token));
    assert!(!tmp.path().join("gh-called").exists());
}

#[test]
fn human_edited_comment_is_not_overwritten() {
    let (_tmp, project) = new_project();
    let p = payload();
    let t = Script::new(with(
        reads(vec![]),
        vec![
            ("POST", POST.into(), Reply::Echo(201, 9001)),
            ("GET", patch(9001), Reply::Last),
        ],
    ));
    run(&project, &p, &t).unwrap();
    let posted = t.last.borrow()["body"].as_str().unwrap().to_string();

    let edited = format!("{posted}\n\nEdited by a maintainer.");
    let mut p2 = payload();
    p2.next = "Merge after audit".into();
    let t = Script::new(reads(vec![comment(9001, &edited)]));
    let e = run(&project, &p2, &t).unwrap_err();
    assert_eq!(kind_of(&e), "conflict");
    assert!(t.writes().is_empty(), "no PATCH over a human edit");
    let proposal = project.join(".moai-cli/evidence/rc1-07c-demo/proposed-comment.md");
    assert_eq!(
        fs::read_to_string(&proposal).unwrap(),
        report::render(&p2).unwrap()
    );
    assert!(
        entry(&project, 2).is_none(),
        "no outbox intent for a write that did not happen"
    );

    // A marker comment this machine never published is not overwritten either.
    let (_tmp2, project2) = new_project();
    let t = Script::new(reads(vec![comment(9001, &posted)]));
    assert_eq!(kind_of(&run(&project2, &p2, &t).unwrap_err()), "conflict");
    assert!(t.writes().is_empty());

    // A deleted marker comment is recorded and created again once.
    let t = Script::new(with(
        reads(vec![]),
        vec![
            ("POST", POST.into(), Reply::Echo(201, 9002)),
            ("GET", patch(9002), Reply::Last),
        ],
    ));
    let out = run(&project, &p2, &t).unwrap();
    assert_eq!(out["observed_deleted"], 9001);
    let e2 = entry(&project, 2).unwrap();
    assert_eq!(e2.intent.unwrap().detail["observed_deleted"], 9001);
}

#[test]
fn artifact_refs_require_digest_and_match_local_files() {
    let (_tmp, project) = new_project();
    let log = b"line 1\nline 2\n".to_vec();
    fs::write(project.join("full.log"), &log).unwrap();
    let digest = moai_cli::digest(&log);
    let local = json!({"kind": "local", "digest": digest, "size_bytes": log.len(), "location": "full.log",
                       "retention": "unknown", "access": "private", "redacted": true, "source": "full.log"});
    let mut v = fixture("payload_not_run.json");

    // No digest, a malformed digest, or an http location: refused before anything is written.
    for bad in [
        json!({"kind": "external", "size_bytes": 1, "location": "https://example.com/a", "retention": "unknown", "access": "public", "redacted": false}),
        json!({"kind": "external", "digest": "sha256:xyz", "size_bytes": 1, "location": "https://example.com/a", "retention": "unknown", "access": "public", "redacted": false}),
        json!({"kind": "actions_artifact", "digest": digest, "size_bytes": 1, "location": "http://example.com/a", "retention": "until:2027-02-06T00:00:00Z", "access": "repo", "redacted": false}),
        json!({"kind": "local", "digest": digest, "size_bytes": log.len(), "location": "../escape", "retention": "unknown", "access": "private", "redacted": true}),
    ] {
        v["artifacts"] = json!([bad]);
        let p: ReportPayload = serde_json::from_value(v.clone()).unwrap();
        assert_eq!(kind_of(&report::render(&p).unwrap_err()), "invalid_input");
    }

    // A local file that does not hash to its digest is refused and not copied.
    let mut wrong = local.clone();
    wrong["digest"] = json!(moai_cli::digest(b"other"));
    v["artifacts"] = json!([wrong]);
    let p: ReportPayload = serde_json::from_value(v.clone()).unwrap();
    let t = Script::new(vec![]);
    assert_eq!(
        kind_of(&run(&project, &p, &t).unwrap_err()),
        "invalid_input"
    );
    assert!(!project.join(".moai-cli").exists());
    assert!(t.sent.borrow().is_empty());

    // A matching local file is preserved under .moai-cli/evidence/<run>/ and marked for export.
    v["artifacts"] = json!([local, {"kind": "release_asset", "digest": digest, "size_bytes": log.len(),
        "location": "https://github.com/acme/widgets/releases/download/v1/full.log",
        "retention": "until:2027-02-06T00:00:00Z", "access": "public", "redacted": true}]);
    let p: ReportPayload = serde_json::from_value(v.clone()).unwrap();
    let body = report::render(&p).unwrap();
    assert!(body.contains(&format!("- local · .moai-cli/evidence/rc1-07c-demo/full.log · {digest} · 14 bytes · retention unknown · access private · redacted true · export/pin 필요")), "{body}");
    assert!(body.contains(
        "- release_asset · https://github.com/acme/widgets/releases/download/v1/full.log"
    ));
    assert!(
        !body.contains("line 2"),
        "the log itself stays out of the body"
    );
    let t = Script::new(with(
        reads(vec![]),
        vec![
            ("POST", POST.into(), Reply::Echo(201, 9001)),
            ("GET", patch(9001), Reply::Last),
        ],
    ));
    let out = run(&project, &p, &t).unwrap();
    let kept = fs::read(project.join(".moai-cli/evidence/rc1-07c-demo/full.log")).unwrap();
    assert_eq!(moai_cli::digest(&kept), digest);
    assert_eq!(
        out["artifacts"][0]["location"],
        ".moai-cli/evidence/rc1-07c-demo/full.log"
    );

    // Without a source the file must already be in the evidence directory, unchanged.
    let mut in_place = v.clone();
    in_place["artifacts"][0]
        .as_object_mut()
        .unwrap()
        .remove("source");
    let p: ReportPayload = serde_json::from_value(in_place).unwrap();
    assert!(report::preserve_local(&project, &p).is_ok());
    fs::write(
        project.join(".moai-cli/evidence/rc1-07c-demo/full.log"),
        b"tampered",
    )
    .unwrap();
    assert_eq!(
        kind_of(&report::preserve_local(&project, &p).unwrap_err()),
        "invalid_input"
    );

    // The rendered body is bounded (REQ-REPORT-011, unmeasured default).
    assert_eq!(report::DEFAULT_MAX_BODY_CHARS, 65_535);
    let mut small: ReportPayload = serde_json::from_value(fixture("payload_not_run.json")).unwrap();
    small.config.max_body_chars = 100;
    let e = report::render(&small).unwrap_err();
    assert_eq!(kind_of(&e), "invalid_input");
    assert!(e.message.contains("too_large"));
    // Only the configured tail is rendered.
    let mut tail: ReportPayload = serde_json::from_value(fixture("payload_not_run.json")).unwrap();
    tail.checks[0].tail = (1..=30).map(|i| format!("out {i}")).collect();
    tail.config.tail_lines = 2;
    let body = report::render(&tail).unwrap();
    assert!(body.contains("  out 29\n  out 30") && !body.contains("out 28"));
}

// ---- CLI: stub gh that records any call and refuses it ----

fn stub_gh(dir: &Path) -> PathBuf {
    let marker = dir.join("gh-called");
    let script = format!("#!/bin/sh\n# test stub: records the call, refuses everything\ntouch '{}'\necho 'stub refuses' >&2\nexit 9\n", marker.display());
    let path = dir.join("gh-stub");
    common::write_executable(&path, &script);
    path
}

fn cli(project: &Path, gh: &Path, input: &str, permission: &str) -> (Value, String) {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args([
            "forge-report",
            project.to_str().unwrap(),
            "--permission-reference",
            permission,
        ])
        .env("MOAI_FORGE_GH_BIN", gh)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    let mut stdin = child.stdin.take().unwrap();
    match std::io::Write::write_all(&mut stdin, input.as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(stdin);
    let out = child.wait_with_output().unwrap();
    (
        serde_json::from_slice(&out.stdout).unwrap(),
        String::from_utf8_lossy(&out.stderr).into(),
    )
}

/// The stdin limit of `forge-report` stays where it was when the command left `src/main.rs`
/// (#2834 R2b): 64 KiB is read and judged as JSON, one byte more is refused as too large, and
/// neither reaches `gh`.
#[test]
fn cli_stdin_limit_is_64_kib_exactly() {
    let (tmp, project) = new_project();
    let gh = stub_gh(tmp.path());
    let (at, _) = cli(&project, &gh, &"a".repeat(65536), "ref-2834");
    assert_eq!(at["error"]["kind"], "invalid_input", "{at}");
    assert_eq!(at["error"]["message"], "invalid typed JSON input");
    let (over, stderr) = cli(&project, &gh, &"a".repeat(65537), "ref-2834");
    assert_eq!(over["error"]["kind"], "input_too_large", "{over}");
    assert_eq!(
        over["error"]["message"],
        "input exceeds 64 KiB; put large output in an artifact reference, not the payload"
    );
    assert!(stderr.contains("input exceeds 64 KiB"), "{stderr}");
    assert!(!tmp.path().join("gh-called").exists());
}
