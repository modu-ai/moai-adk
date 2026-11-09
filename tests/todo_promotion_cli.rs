//! 07-B (#2009) selected TODO promotion. Library tests drive `promote::preview`/`apply`
//! through a scripted transport with an injected clock: no network, no sleeping. The CLI tests
//! point `todo-promote-*` at a stub `gh`. Every text is synthetic; credential-shaped values
//! are assembled at run time from harmless pieces so no such literal sits in the repository.
use moai_cli::{
    contract,
    forge::{
        outbox::{self, OutboxState},
        policy::Clock,
        promote::{self, PromoteRequest},
        transport::{HttpRequest, HttpResponse, Transport},
        Client, Config, ForgeError,
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

const API: &str = "https://api.github.com/repos/acme/widgets";
const LIST: &str = "https://api.github.com/repos/acme/widgets/issues?state=all&sort=updated&direction=asc&per_page=100";
const CREATE: &str = "https://api.github.com/repos/acme/widgets/issues";

struct FakeClock {
    now: Cell<i64>,
    slept: RefCell<Vec<u64>>,
}
impl FakeClock {
    fn new() -> Self {
        FakeClock {
            now: Cell::new(1_794_000_000),
            slept: RefCell::new(Vec::new()),
        }
    }
}
impl Clock for FakeClock {
    fn now(&self) -> i64 {
        self.now.get()
    }
    fn sleep(&self, secs: u64) {
        self.slept.borrow_mut().push(secs);
        self.now.set(self.now.get() + secs as i64);
    }
}

enum Reply {
    Resp(HttpResponse),
    /// No HTTP response at all (timeout, connection loss).
    Fail(&'static str),
    /// Runs `before` (another writer, say), then answers.
    After(Box<dyn Fn()>, HttpResponse),
}

/// Expected (method, url) in order; records every request actually sent.
struct Script {
    steps: RefCell<VecDeque<(&'static str, String, Reply)>>,
    sent: RefCell<Vec<HttpRequest>>,
}
impl Script {
    fn new(steps: Vec<(&'static str, String, Reply)>) -> Self {
        Script {
            steps: RefCell::new(steps.into()),
            sent: RefCell::new(Vec::new()),
        }
    }
    fn remaining(&self) -> usize {
        self.steps.borrow().len()
    }
    fn methods(&self) -> Vec<String> {
        self.sent
            .borrow()
            .iter()
            .map(|r| r.method.clone())
            .collect()
    }
    fn posts(&self) -> Vec<HttpRequest> {
        self.sent
            .borrow()
            .iter()
            .filter(|r| r.method == "POST")
            .cloned()
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
            Reply::After(before, r) => {
                before();
                Ok(r)
            }
        }
    }
}

fn ok(body: Value) -> HttpResponse {
    HttpResponse::new(200)
        .with_header("etag", "\"e1\"")
        .with_body(body)
}
fn status(code: u16, body: Value) -> HttpResponse {
    HttpResponse::new(code).with_body(body)
}
fn repo() -> (&'static str, String, Reply) {
    (
        "GET",
        API.into(),
        Reply::Resp(ok(
            json!({"id": 101, "visibility": "private", "private": true}),
        )),
    )
}
fn list(issues: Value) -> (&'static str, String, Reply) {
    ("GET", LIST.into(), Reply::Resp(ok(issues)))
}
fn issue(number: i64, title: &str, body: &str) -> Value {
    json!({"id": 9000 + number, "node_id": format!("I_{number}"), "number": number, "title": title,
           "body": body, "html_url": format!("https://github.com/acme/widgets/issues/{number}"),
           "updated_at": "2026-11-08T00:00:00Z", "state": "open"})
}
fn issue_get(number: i64, v: Value) -> (&'static str, String, Reply) {
    ("GET", format!("{API}/issues/{number}"), Reply::Resp(ok(v)))
}
fn marker(local: &str, attempt: u32) -> String {
    outbox::op_marker("promote", "demo", local, attempt)
}

/// Design-22 shaped tasks file. T-3 is never selected; its text must never leave.
const TASKS: &str = include_str!("fixtures/forge/promote/tasks_three.md");

fn project_with(tasks: &str) -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("p");
    fs::create_dir(&project).unwrap();
    fs::write(project.join("tasks.md"), tasks).unwrap();
    (tmp, project)
}
fn request(selected: &[&str]) -> PromoteRequest {
    serde_json::from_value(json!({"schema_version": 1, "tasks_path": "tasks.md",
        "target": {"owner": "acme", "repo": "widgets"}, "selected": selected}))
    .unwrap()
}
fn client<'a>(t: &'a Script, c: &'a FakeClock) -> Client<'a, Script, FakeClock> {
    Client::new(t, c, Config::default(), Default::default())
}
fn promotions(project: &Path) -> Option<Value> {
    fs::read(project.join(".moai-cli/forge/promotions.json"))
        .ok()
        .map(|b| serde_json::from_slice(&b).unwrap())
}
fn outbox_entry(project: &Path, local: &str, attempt: u32) -> Option<outbox::OutboxEntry> {
    outbox::load(
        project,
        &outbox::op_id("promote", "demo", local, attempt).unwrap(),
    )
    .unwrap()
}
fn files_under(dir: &Path) -> Vec<PathBuf> {
    let mut out = Vec::new();
    for e in fs::read_dir(dir).unwrap() {
        let p = e.unwrap().path();
        if p.is_dir() {
            out.extend(files_under(&p));
        }
        out.push(p);
    }
    out.sort();
    out
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
fn preview_digest(
    project: &Path,
    req: &PromoteRequest,
    steps: Vec<(&'static str, String, Reply)>,
) -> (Value, String) {
    let t = Script::new(steps);
    let c = FakeClock::new();
    let pv = promote::preview(project, req, &client(&t, &c)).unwrap();
    assert_eq!(t.remaining(), 0);
    (pv.result, pv.digest)
}
/// A created Issue as GitHub would echo it, carrying the body that was sent.
fn created(number: i64, title: &str, local: &str, attempt: u32) -> Value {
    issue(
        number,
        title,
        &format!(
            "- acceptance: form renders\n- Next: write the view\n\n{}",
            marker(local, attempt)
        ),
    )
}

/// T-4 as GitHub would echo it after creation.
fn created_t4(number: i64) -> Value {
    issue(
        number,
        "Fix flaky upload test",
        &format!("- acceptance: green\n\n{}", marker("T-4", 1)),
    )
}

#[test]
fn promote_only_selected_after_preview() {
    // Three tasks, two selected (TASK-FORGE-MIGRATE).
    let (_tmp, project) = project_with(TASKS);
    let tasks_before = fs::read(project.join("tasks.md")).unwrap();
    let req = request(&["T-1", "T-4"]);
    let existing = json!([issue(5, "Unrelated report", "reported earlier")]);
    let (result, digest) = preview_digest(&project, &req, vec![repo(), list(existing.clone())]);
    let rows = result["rows"].as_array().unwrap();
    assert_eq!(
        rows.iter()
            .map(|r| r["local_id"].as_str().unwrap())
            .collect::<Vec<_>>(),
        ["T-1", "T-4"]
    );
    assert_eq!(rows[0]["state"], "preview_only");
    assert_eq!(rows[1]["state"], "preview_only");
    assert_eq!(
        rows[0]["target"],
        json!({"owner": "acme", "repo": "widgets", "visibility": "private"})
    );
    // The key set, not its order: serde_json keeps insertion order (preserve_order, #1992).
    let mut keys: Vec<&String> = rows[0].as_object().unwrap().keys().collect();
    keys.sort();
    assert_eq!(
        keys,
        [
            "local_id",
            "redacted_body_digest",
            "redacted_title",
            "state",
            "target"
        ]
    );
    let text = result.to_string();
    for unselected in ["T-3", "payroll", "private-note-unselected"] {
        assert!(
            !text.contains(unselected),
            "unselected task leaked: {unselected}"
        );
    }

    // Apply needs a permission reference and the same digest; neither failure sends a write.
    for (perm, d, kind) in [
        ("", digest.as_str(), "permission_reference_missing"),
        (
            "placeholder",
            digest.as_str(),
            "permission_reference_missing",
        ),
        ("ok-ref", "sha256:other", "stale_plan"),
    ] {
        let t = Script::new(vec![repo(), list(existing.clone())]);
        let c = FakeClock::new();
        let e = promote::apply(&project, &req, d, perm, &client(&t, &c)).unwrap_err();
        assert_eq!(kind_of(&e), kind);
        assert!(t.posts().is_empty());
    }

    let t = Script::new(vec![
        repo(),
        list(existing.clone()),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(201, created(12, "Add login page", "T-1", 1))),
        ),
        issue_get(12, created(12, "Add login page", "T-1", 1)),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(201, created_t4(13))),
        ),
        issue_get(13, created_t4(13)),
    ]);
    let c = FakeClock::new();
    let out = promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert_eq!(t.remaining(), 0);
    assert_eq!(
        t.methods(),
        ["GET", "GET", "POST", "GET", "POST", "GET"],
        "each create is confirmed by GET"
    );
    let posts = t.posts();
    assert_eq!(posts.len(), 2);
    let body = posts[0].body.as_ref().unwrap();
    assert_eq!(body["title"], "Add login page");
    let sent = body["body"].as_str().unwrap();
    assert!(sent.ends_with(&marker("T-1", 1)));
    assert!(
        !sent.contains("not_run") && !sent.contains("External issue"),
        "local-only lines stay local"
    );
    for r in t.sent.borrow().iter() {
        let all = serde_json::to_string(r).unwrap();
        assert!(!all.contains("T-3") && !all.contains("payroll"));
    }
    assert_eq!(
        out["outcomes"][0],
        json!({"local_id": "T-1", "outcome": "published", "number": 12})
    );
    assert_eq!(out["outcomes"][1]["outcome"], "published");
    let map = promotions(&project).unwrap();
    let records = map["records"].as_array().unwrap();
    assert_eq!(
        records.len(),
        2,
        "two mappings, none for the unselected task"
    );
    assert!(records.iter().all(|r| r["state"] == "published"));
    assert_eq!(
        records[0]["remote"]["issue"],
        json!({"id": 9012, "node_id": "I_12", "number": 12})
    );
    assert_eq!(records[1]["remote"]["issue"]["number"], 13);
    assert_eq!(records[0]["remote"]["project"]["repository_id"], 101);
    assert!(!map.to_string().contains("T-3"));
    // PROMO-PRESERVE: the user's tasks.md is never edited; only .moai-cli/forge changes.
    assert_eq!(fs::read(project.join("tasks.md")).unwrap(), tasks_before);
    let others: Vec<PathBuf> = files_under(&project)
        .into_iter()
        .filter(|p| !p.starts_with(project.join(".moai-cli")))
        .collect();
    assert_eq!(others, [project.join("tasks.md")]);

    // A similar title only marks a candidate (REQ-PROMO-009): never linked to the
    // existing Issue and, as implemented, not created either.
    let (_tmp2, project2) = project_with(TASKS);
    let similar = json!([issue(5, "Fix flaky upload test", "reported earlier")]);
    let req4 = request(&["T-4"]);
    let (result, digest) = preview_digest(&project2, &req4, vec![repo(), list(similar.clone())]);
    assert_eq!(result["rows"][0]["state"], "duplicate_candidate");
    let t = Script::new(vec![repo(), list(similar)]);
    let out = promote::apply(&project2, &req4, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert!(t.posts().is_empty());
    assert_eq!(
        out["outcomes"][0],
        json!({"local_id": "T-4", "outcome": "not_sent", "state": "duplicate_candidate"})
    );
    assert!(
        promotions(&project2).is_none(),
        "no link to the similar Issue"
    );
}

#[test]
fn secrets_never_leave_and_are_not_echoed() {
    // Credential shapes built from pieces: cloud key id, classic token, bearer header.
    let aws = format!("{}{}", "AK", "IAZZZZZZZZZZZZZZZZ");
    let ghp = format!("{}{}", "gh", "p_moaiTestDummyValue0000000000000000");
    let bearer = format!("{} {}", "Bearer", "moai.dummy.value");
    let tasks = format!(
        "Project: demo\n\
- [ ] T-1 · Rotate deploy key · state=ready\n  - Next: staging uses {aws}\n  - disclosure_review: approved\n\
- [ ] T-2 · Wire CI · state=ready\n  - Next: export {ghp}\n  - disclosure_review: approved\n\
- [ ] T-5 · Call the API · state=ready\n  - Next: send Authorization: {bearer}\n  - disclosure_review: approved\n\
- [ ] T-6 · Clean /Users/someone/work · state=ready\n  - disclosure_review: approved\n\
- [ ] T-7 · Plain but unreviewed · state=ready\n"
    );
    let (_tmp, project) = project_with(&tasks);
    let req = request(&["T-1", "T-2", "T-5", "T-6", "T-7"]);
    let (result, digest) = preview_digest(&project, &req, vec![repo(), list(json!([]))]);
    let text = result.to_string();
    for secret in [
        aws.as_str(),
        ghp.as_str(),
        "moai.dummy.value",
        "/Users/someone",
    ] {
        assert!(!text.contains(secret), "secret echoed in preview");
    }
    assert!(result["rows"]
        .as_array()
        .unwrap()
        .iter()
        .all(|r| r["state"] == "blocked_review"));
    assert_eq!(result["rows"][3]["redacted_title"], "[redacted]");
    assert_eq!(
        result["rows"][4]["redacted_title"], "Plain but unreviewed",
        "unreviewed is blocked, not redacted"
    );

    let t = Script::new(vec![repo(), list(json!([]))]);
    let c = FakeClock::new();
    let out = promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert_eq!(
        t.methods(),
        ["GET", "GET"],
        "blocked rows never reach the network"
    );
    assert!(!out.to_string().contains(&aws) && !out.to_string().contains(&ghp));
    assert!(!project.join(".moai-cli").exists(), "nothing to record");
}

#[test]
fn timeout_yields_uncertain_and_reconciles_before_retry() {
    let (_tmp, project) = project_with(TASKS);
    let req = request(&["T-1"]);
    let (_, digest) = preview_digest(&project, &req, vec![repo(), list(json!([]))]);
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        ("POST", CREATE.into(), Reply::Fail("operation timed out")),
    ]);
    let c = FakeClock::new();
    let e = promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap_err();
    assert_eq!(kind_of(&e), "partial");
    assert!(e.message.contains("uncertain at T-1"));
    assert_eq!(
        promotions(&project).unwrap()["records"][0]["state"],
        "uncertain"
    );
    let entry = outbox_entry(&project, "T-1", 1).unwrap();
    assert_eq!(entry.state, OutboxState::PublishUnknown);
    assert!(entry.intent.is_some() && entry.response.is_some() && entry.observed.is_none());

    // Not found is not "not created": the next run shows `uncertain` and sends nothing.
    let (result, digest) = preview_digest(&project, &req, vec![repo(), list(json!([]))]);
    assert_eq!(result["rows"][0]["state"], "uncertain");
    let t = Script::new(vec![repo(), list(json!([]))]);
    let out = promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert!(
        t.posts().is_empty(),
        "uncertain is never created again automatically"
    );
    assert_eq!(out["outcomes"][0]["outcome"], "not_sent");

    // The Issue did land: its marker is found, confirmed by GET and linked, not re-created.
    let landed = json!([created(31, "Add login page", "T-1", 1)]);
    let (result, digest) = preview_digest(&project, &req, vec![repo(), list(landed.clone())]);
    assert_eq!(result["rows"][0]["state"], "already_published");
    let t = Script::new(vec![
        repo(),
        list(landed),
        issue_get(31, created(31, "Add login page", "T-1", 1)),
    ]);
    let out = promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert!(t.posts().is_empty());
    assert_eq!(
        out["outcomes"][0],
        json!({"local_id": "T-1", "outcome": "linked", "number": 31})
    );
    let map = promotions(&project).unwrap();
    assert_eq!(map["records"][0]["state"], "published");
    assert_eq!(map["records"][0]["remote"]["issue"]["node_id"], "I_31");
    assert_eq!(
        outbox_entry(&project, "T-1", 1).unwrap().state,
        OutboxState::Confirmed
    );

    // A 5xx is unknown in the same way.
    let (_tmp2, project2) = project_with(TASKS);
    let (_, digest) = preview_digest(&project2, &req, vec![repo(), list(json!([]))]);
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        ("POST", CREATE.into(), Reply::Resp(status(502, json!({})))),
    ]);
    let e = promote::apply(&project2, &req, &digest, "ref-2009", &client(&t, &c)).unwrap_err();
    assert_eq!(kind_of(&e), "partial");
    assert_eq!(
        promotions(&project2).unwrap()["records"][0]["state"],
        "uncertain"
    );
    assert_eq!(
        outbox_entry(&project2, "T-1", 1).unwrap().state,
        OutboxState::PublishUnknown
    );
}

#[test]
fn mapping_conflict_stops_without_overwrite() {
    let (_tmp, project) = project_with(TASKS);
    let req = request(&["T-1"]);
    let (_, digest) = preview_digest(&project, &req, vec![repo(), list(json!([]))]);
    // Another writer records the mapping while this run is creating the Issue.
    let other = json!({"schema_version": 1, "project_id": "demo", "records": []});
    let path = project.join(".moai-cli/forge/promotions.json");
    let (p2, o2) = (path.clone(), other.clone());
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        (
            "POST",
            CREATE.into(),
            Reply::After(
                Box::new(move || {
                    fs::create_dir_all(p2.parent().unwrap()).unwrap();
                    fs::write(&p2, serde_json::to_vec(&o2).unwrap()).unwrap();
                }),
                status(201, created(12, "Add login page", "T-1", 1)),
            ),
        ),
        issue_get(12, created(12, "Add login page", "T-1", 1)),
    ]);
    let c = FakeClock::new();
    let e = promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap_err();
    assert_eq!(kind_of(&e), "conflict");
    let now: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    assert_eq!(now, other, "the other writer's mapping is kept");
    // The outbox still knows what was created, for a later reconcile.
    assert_eq!(
        outbox_entry(&project, "T-1", 1).unwrap().state,
        OutboxState::Confirmed
    );

    // The library guard alone: a stale read digest never overwrites.
    let (p, read) = promote::load_promotions(&project).unwrap();
    fs::write(
        &path,
        b"{\"schema_version\":1,\"project_id\":\"demo\",\"records\":[]}\n",
    )
    .unwrap();
    let e = promote::save_promotions(&project, &p, &read).unwrap_err();
    assert_eq!(kind_of(&e), "conflict");
}

#[test]
fn offline_keeps_issue_authority() {
    let (_tmp, project) = project_with(TASKS);
    let req = request(&["T-1"]);
    let dir = project.join(".moai-cli/forge");
    fs::create_dir_all(&dir).unwrap();
    let published = json!({"schema_version": 1, "project_id": "demo", "records": [{
        "local_id": "T-1", "marker": marker("T-1", 1), "state": "published", "attempt": 1,
        "remote": {"provider": "github", "host": "github.com",
                   "project": {"owner": "acme", "repo": "widgets", "repository_id": 101},
                   "issue": {"id": 9012, "node_id": "I_12", "number": 12}},
        "recorded_at": "2026-11-08T00:00:00Z"}]});
    let bytes = serde_json::to_vec_pretty(&published).unwrap();
    fs::write(dir.join("promotions.json"), &bytes).unwrap();
    let before = files_under(&project);

    let c = FakeClock::new();
    let t = Script::new(vec![(
        "GET",
        API.into(),
        Reply::Fail("network unreachable"),
    )]);
    let e = promote::preview(&project, &req, &client(&t, &c))
        .err()
        .unwrap();
    assert_eq!(kind_of(&e), "transport");
    let t = Script::new(vec![(
        "GET",
        API.into(),
        Reply::Fail("network unreachable"),
    )]);
    let e = promote::apply(&project, &req, "sha256:any", "ref-2009", &client(&t, &c)).unwrap_err();
    assert_eq!(kind_of(&e), "transport");
    assert_eq!(
        fs::read(dir.join("promotions.json")).unwrap(),
        bytes,
        "published record unchanged"
    );
    assert_eq!(files_under(&project), before);

    // Online again, the Issue stays the authority: published is shown, nothing is sent.
    let (result, digest) = preview_digest(&project, &req, vec![repo(), list(json!([]))]);
    assert_eq!(result["rows"][0]["state"], "already_published");
    let t = Script::new(vec![repo(), list(json!([]))]);
    promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert!(t.posts().is_empty());
    assert_eq!(fs::read(dir.join("promotions.json")).unwrap(), bytes);
}

#[test]
fn outbox_transitions_and_requery() {
    let (_tmp, project) = project_with(TASKS);
    let req = request(&["T-1"]);
    let (_, digest) = preview_digest(&project, &req, vec![repo(), list(json!([]))]);
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(201, created(12, "Add login page", "T-1", 1))),
        ),
        issue_get(12, created(12, "Add login page", "T-1", 1)),
    ]);
    let c = FakeClock::new();
    promote::apply(&project, &req, &digest, "ref-2009", &client(&t, &c)).unwrap();
    assert_eq!(
        t.methods(),
        ["GET", "GET", "POST", "GET"],
        "confirmation GET after the create"
    );
    let e = outbox_entry(&project, "T-1", 1).unwrap();
    assert_eq!(e.state, OutboxState::Confirmed);
    assert_eq!(e.marker, marker("T-1", 1));
    assert!(e.intent.is_some() && e.response.is_some() && e.observed.is_some());

    // The confirming GET disagrees (another Issue answers): conflict, recorded uncertain.
    let (_tmp2, project2) = project_with(TASKS);
    let (_, digest) = preview_digest(&project2, &req, vec![repo(), list(json!([]))]);
    let mut wrong = created(12, "Add login page", "T-1", 1);
    wrong["node_id"] = json!("I_other");
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(201, created(12, "Add login page", "T-1", 1))),
        ),
        issue_get(12, wrong),
    ]);
    let err = promote::apply(&project2, &req, &digest, "ref-2009", &client(&t, &c)).unwrap_err();
    assert_eq!(kind_of(&err), "partial");
    assert_eq!(
        outbox_entry(&project2, "T-1", 1).unwrap().state,
        OutboxState::Conflict
    );
    assert_eq!(
        promotions(&project2).unwrap()["records"][0]["state"],
        "uncertain"
    );

    // A refused create (4xx) is blocked and records no mapping.
    let (_tmp3, project3) = project_with(TASKS);
    let (_, digest) = preview_digest(&project3, &req, vec![repo(), list(json!([]))]);
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(422, json!({"message": "Validation Failed"}))),
        ),
    ]);
    let err = promote::apply(&project3, &req, &digest, "ref-2009", &client(&t, &c)).unwrap_err();
    assert_eq!(kind_of(&err), "partial");
    assert_eq!(
        outbox_entry(&project3, "T-1", 1).unwrap().state,
        OutboxState::Blocked
    );
    assert!(promotions(&project3).is_none());

    // Writes are spaced at least one second apart (secondary rate limits).
    let (_tmp4, project4) =
        project_with(&TASKS.replace("T-4 · Fix flaky upload test", "T-4 · Tidy changelog"));
    let req4 = request(&["T-1", "T-4"]);
    let (_, digest) = preview_digest(&project4, &req4, vec![repo(), list(json!([]))]);
    let mut second = issue(13, "Tidy changelog", &marker("T-4", 1));
    second["body"] = json!(format!("- acceptance: green\n\n{}", marker("T-4", 1)));
    let t = Script::new(vec![
        repo(),
        list(json!([])),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(201, created(12, "Add login page", "T-1", 1))),
        ),
        issue_get(12, created(12, "Add login page", "T-1", 1)),
        (
            "POST",
            CREATE.into(),
            Reply::Resp(status(201, second.clone())),
        ),
        issue_get(13, second),
    ]);
    let c4 = FakeClock::new();
    promote::apply(&project4, &req4, &digest, "ref-2009", &client(&t, &c4)).unwrap();
    assert_eq!(*c4.slept.borrow(), [1]);
}

// ---- CLI: stub gh, envelope, no files on preview ----

fn stub_gh(dir: &Path) -> PathBuf {
    let repo = "HTTP/2.0 200 OK\r\nEtag: \"repo-1\"\r\n\r\n{\"id\":101,\"visibility\":\"public\"}";
    let issues = "HTTP/2.0 200 OK\r\nEtag: W/\"issues-1\"\r\n\r\n[]";
    let script = format!(
        "#!/bin/sh\n# test stub for gh api (GET only)\nfor a in \"$@\"; do last=\"$a\"; done\nfor a in \"$@\"; do [ \"$a\" = POST ] && {{ echo 'stub refuses writes' >&2; exit 9; }}; done\ncase \"$last\" in\n  *issues*) printf '%s' '{issues}';;\n  *) printf '%s' '{repo}';;\nesac\n"
    );
    let path = dir.join("gh-stub");
    common::write_executable(&path, &script);
    path
}

fn cli(args: &[&str], gh: &Path, input: &str) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
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
    serde_json::from_slice(&out.stdout).unwrap()
}

#[test]
fn preview_creates_no_files() {
    let (tmp, project) = project_with(TASKS);
    let gh = stub_gh(tmp.path());
    let before = files_under(&project);
    let input = json!({"schema_version": 1, "tasks_path": "tasks.md",
        "target": {"owner": "acme", "repo": "widgets"}, "selected": ["T-1"]})
    .to_string();
    let p = project.to_str().unwrap();
    let out = cli(&["todo-promote-preview", p], &gh, &input);
    assert_eq!(out["operation"], "todo.promote.plan");
    assert_eq!(out["ok"], true, "{out}");
    assert_eq!(out["result"]["rows"][0]["state"], "preview_only");
    assert_eq!(
        files_under(&project),
        before,
        "preview writes nothing (REQ-PROMO-011)"
    );

    // A stale digest through the CLI: envelope kind and next action, still nothing written.
    let out = cli(
        &[
            "todo-promote-apply",
            p,
            "--preview-digest",
            "sha256:old",
            "--permission-reference",
            "ref-2009",
        ],
        &gh,
        &input,
    );
    assert_eq!(out["operation"], "todo.promote.apply");
    assert_eq!(out["error"]["kind"], "stale_plan");
    assert_eq!(out["error"]["next_actions"], json!(["todo.promote.plan"]));
    assert_eq!(files_under(&project), before);

    // The tasks path is required, bounded and not a symlink.
    let bad = json!({"schema_version": 1, "target": {"owner": "acme", "repo": "widgets"}, "selected": ["T-1"]}).to_string();
    assert_eq!(
        cli(&["todo-promote-preview", p], &gh, &bad)["error"]["kind"],
        "invalid_input"
    );
    #[cfg(unix)]
    {
        std::os::unix::fs::symlink(project.join("tasks.md"), project.join("link.md")).unwrap();
        let linked = input.replace("tasks.md", "link.md");
        assert_eq!(
            cli(&["todo-promote-preview", p], &gh, &linked)["error"]["kind"],
            "invalid_input"
        );
    }
}
