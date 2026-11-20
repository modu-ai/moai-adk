mod common;
#[cfg(unix)]
use common::CleanEnv;
use moai_cli::{
    digest,
    forge::{
        policy::Clock,
        transport::{HttpRequest, HttpResponse, Transport},
    },
    message::{self, Message},
};
use serde_json::{json, Value};
use std::{
    cell::{Cell, RefCell},
    fs,
    path::Path,
};
use tempfile::TempDir;

fn temp_root() -> TempDir {
    TempDir::new_in(fs::canonicalize(std::env::temp_dir()).unwrap()).unwrap()
}
struct Time;
impl Clock for Time {
    fn now(&self) -> i64 {
        1_794_355_200
    }
    fn sleep(&self, _: u64) {}
}
struct Remote {
    comments: RefCell<Vec<Value>>,
    sent: RefCell<Vec<HttpRequest>>,
    actor: Cell<u64>,
    lost: Cell<bool>,
    lands: Cell<bool>,
    page_size: Cell<usize>,
    crash: Cell<bool>,
    expected_intent: RefCell<Option<std::path::PathBuf>>,
    post_status: Cell<u16>,
    links: RefCell<Vec<(String, String)>>,
    lose_n: Cell<Option<usize>>,
    hide_n: Cell<Option<usize>>,
    comments_gets: Cell<usize>,
    comments_get_limit: Cell<Option<usize>>,
}
impl Remote {
    fn new() -> Self {
        Self {
            comments: RefCell::new(vec![]),
            sent: RefCell::new(vec![]),
            actor: Cell::new(11),
            lost: Cell::new(false),
            lands: Cell::new(true),
            page_size: Cell::new(100),
            crash: Cell::new(false),
            expected_intent: RefCell::new(None),
            post_status: Cell::new(201),
            links: RefCell::new(vec![]),
            lose_n: Cell::new(None),
            hide_n: Cell::new(None),
            comments_gets: Cell::new(0),
            comments_get_limit: Cell::new(None),
        }
    }
    fn posts(&self) -> usize {
        self.sent
            .borrow()
            .iter()
            .filter(|r| r.method == "POST")
            .count()
    }
}
impl Transport for Remote {
    fn send(&self, request: &HttpRequest) -> Result<HttpResponse, String> {
        self.sent.borrow_mut().push(request.clone());
        if self.crash.replace(false) {
            let path = self.expected_intent.borrow();
            let entry: Value =
                serde_json::from_slice(&fs::read(path.as_ref().unwrap()).unwrap()).unwrap();
            assert_eq!(entry["record"]["attempted"], false);
            assert_eq!(entry["message"]["report_id"], "r1");
            panic!("simulated crash after durable intent, before network");
        }
        if request.method == "POST" {
            if self.post_status.get() != 201 {
                return Ok(HttpResponse::new(self.post_status.get()));
            }
            let id = 100 + self.comments.borrow().len() as u64;
            if self.lands.get() && self.hide_n.get() != Some(self.posts()) {
                self.comments.borrow_mut().push(json!({"id":id,"author_association":"MEMBER","body":request.body.as_ref().unwrap()["body"],"user":{"id":self.actor.get()},"created_at":"2026-11-11T00:00:00Z"}));
            }
            if self.lost.replace(false) || self.lose_n.get() == Some(self.posts()) {
                return Err("response lost".into());
            }
            return Ok(HttpResponse::new(201).with_body(json!({"id":id})));
        }
        assert_eq!(request.method, "GET");
        let body = if request.url.ends_with("/user") {
            json!({"id":self.actor.get()})
        } else if request.url.contains("/comments?") {
            let gets = self.comments_gets.get() + 1;
            self.comments_gets.set(gets);
            if self
                .comments_get_limit
                .get()
                .is_some_and(|limit| gets > limit)
            {
                return Err("comments GET limit exceeded".into());
            }
            let page: usize = request
                .url
                .split("&page=")
                .nth(1)
                .unwrap_or("1")
                .parse()
                .unwrap();
            let size = self.page_size.get();
            let comments = self.comments.borrow();
            let start = (page - 1) * size;
            let body = json!(comments.iter().skip(start).take(size).collect::<Vec<_>>());
            let mut response = HttpResponse::new(200).with_body(body);
            if let Some((_, link)) = self
                .links
                .borrow()
                .iter()
                .find(|(url, _)| url == &request.url)
            {
                response = response.with_header("link", link.clone());
            } else if start + size < comments.len() {
                response = response.with_header("link", format!("<https://api.github.com/repos/org/repo/issues/3/comments?per_page=100&page={}>; rel=\"next\"", page + 1));
            }
            return Ok(response);
        } else if request.url.ends_with("/issues/3") {
            json!({"id":2,"node_id":"I_2","number":3,"body":"HOLD: preserve draft"})
        } else if request.url.ends_with("/repos/org/repo") {
            json!({"id":1})
        } else {
            panic!("unexpected URL {}", request.url)
        };
        Ok(HttpResponse::new(200).with_body(body))
    }
}

use moai_cli::dispatch::{
    self, AckRequest as DispatchAck, CheckRequest, RecordRequest as DispatchRecord,
};
fn packet() -> Message {
    serde_json::from_value(json!({"schema_version":1,"report_id":"d1","project_id":"p1",
 "task":{"owner":"org","repo":"repo","repository_id":1,"issue_id":2,"issue_node_id":"I_2","issue_number":3,"task_id":"T1","revision":"1"},
 "sender":{"actor_id":11,"session_id":"leader"},"receiver":{"actor_id":22,"session_id":"lane"},"owner_epoch":1,"leader_epoch":1,
 "event":"dispatch","candidate_digest":digest(b"candidate"),"created_at":"2026-11-11T00:00:00Z","approval_refs":["decision:1"],"hold_refs":[],
 "payload":{"dispatch_id":"job1","revision":1,"previous":null,"scope":["src/example.rs"],"completion":["tests pass"],"audit":null}})).unwrap()
}
fn send(root: &Path, m: &Message, r: &Remote) -> Result<Value, moai_cli::contract::OpError> {
    dispatch::record_cli(
        root,
        DispatchRecord { message: m.clone() },
        "decision:1",
        r,
        &Time,
    )
}
fn check(m: &Message, r: &Remote) -> Result<Value, moai_cli::contract::OpError> {
    dispatch::check_cli(
        CheckRequest {
            dispatch: m.clone(),
        },
        r,
        &Time,
    )
}
#[test]
fn queued_is_neither_received_nor_accepted() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    let v = send(root.path(), &m, &r).unwrap();
    assert_eq!(v["queued"], true);
    assert_eq!(v["received_ack"], false);
    assert_eq!(v["disposition"], Value::Null);
    assert_eq!(r.posts(), 1);
    assert_eq!(check(&m, &r).unwrap()["claim_authorized"], false);
}

fn payload_digest(m: &Message) -> String {
    let mut p = m.payload.clone();
    p.sort_all_objects();
    digest(&serde_json::to_vec(&p).unwrap())
}
fn answer(m: &Message, kind: &str) -> Message {
    let mut d = m.clone();
    d.report_id = format!("{}-decision", m.report_id);
    d.event = "dispatch-decision".into();
    std::mem::swap(&mut d.sender, &mut d.receiver);
    d.payload = json!({"dispatch_report_id":m.report_id,"dispatch_payload_digest":payload_digest(m),"disposition":kind,"reason":"explicit receiver decision"});
    d
}
fn ack(
    root: &Path,
    m: &Message,
    d: Message,
    r: &Remote,
) -> Result<Value, moai_cli::contract::OpError> {
    dispatch::ack_cli(
        root,
        DispatchAck {
            dispatch: m.clone(),
            decision: d,
        },
        "decision:1",
        r,
        &Time,
    )
}
fn next(m: &Message, v: &Value) -> Message {
    let mut n = m.clone();
    n.report_id = "d2".into();
    n.task.revision = "2".into();
    n.receiver.session_id = "next-lane".into();
    n.owner_epoch = 2;
    n.payload["revision"] = json!(2);
    n.payload["previous"] = json!({"report_id":m.report_id,"payload_digest":v["current_payload_digest"],"comment_id":v["record_comment_id"]});
    n
}
#[test]
fn explicit_decisions_and_restart_keep_receipt_and_acceptance_distinct() {
    for kind in ["accepted", "deferred", "rejected"] {
        let root = temp_root();
        let r = Remote::new();
        let m = packet();
        send(root.path(), &m, &r).unwrap();
        r.actor.set(22);
        let d = answer(&m, kind);
        let v = ack(root.path(), &m, d.clone(), &r).unwrap();
        assert_eq!(v["received_ack"], true);
        assert_eq!(v["disposition"], kind);
        assert_eq!(r.posts(), 3);
        let restarted = temp_root();
        assert_eq!(
            ack(restarted.path(), &m, d, &r).unwrap()["disposition"],
            kind
        );
        assert_eq!(r.posts(), 3);
        assert_eq!(v["claim_authorized"], false);
        assert_eq!(v["audit_gate_integrated"], false);
        assert_eq!(v["business_completion"], false);
    }
}
struct At(i64);
impl Clock for At {
    fn now(&self) -> i64 {
        self.0
    }
    fn sleep(&self, _: u64) {}
}
#[test]
fn ack_warning_uses_remote_time_at_299_and_300_without_takeover() {
    let root = temp_root();
    let r = Remote::new();
    let mut m = packet();
    m.created_at = "2099-01-01T00:00:00Z".into();
    send(root.path(), &m, &r).unwrap();
    for (offset, state) in [(299, "waiting"), (300, "unacknowledged"), (-1, "unknown")] {
        let v = dispatch::check_cli(
            CheckRequest {
                dispatch: m.clone(),
            },
            &r,
            &At(Time.now() + offset),
        )
        .unwrap();
        assert_eq!(v["ack_alert"], state);
        assert_eq!(v["ownership_transferred"], false);
        if offset < 0 {
            assert_eq!(v["observation_current"], false);
        }
    }
    r.actor.set(22);
    message::ack_cli(
        root.path(),
        message::AckRequest {
            message: m.clone(),
            payload_digest: payload_digest(&m),
        },
        "decision:1",
        &r,
        &Time,
    )
    .unwrap();
    let v = check(&m, &r).unwrap();
    assert_eq!(v["ack_alert"], "received");
    assert_eq!(v["disposition"], Value::Null);
}
#[test]
fn latest_revision_supersedes_old_receiver_without_granting_a_claim() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    let v = send(root.path(), &m, &r).unwrap();
    let n = next(&m, &v);
    send(root.path(), &n, &r).unwrap();
    r.actor.set(22);
    let v = check(&m, &r).unwrap();
    assert_eq!(v["superseded"], true);
    assert_eq!(v["current_revision"], 2);
    assert_eq!(v["disposition"], Value::Null);
    let posts = r.posts();
    assert!(ack(root.path(), &m, answer(&m, "accepted"), &r).is_err());
    assert_eq!(r.posts(), posts);
    let v = ack(root.path(), &n, answer(&n, "accepted"), &r).unwrap();
    assert_eq!(v["disposition"], "accepted");
    assert_eq!(v["claim_authorized"], false);
}
#[test]
fn decision_identity_revision_and_immutable_disposition_fail_before_post() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    for field in [
        "owner_epoch",
        "leader_epoch",
        "candidate_digest",
        "task",
        "sender",
        "payload",
    ] {
        let mut d = answer(&m, "accepted");
        match field {
            "owner_epoch" => d.owner_epoch = 2,
            "leader_epoch" => d.leader_epoch = 2,
            "candidate_digest" => d.candidate_digest = digest(b"different"),
            "task" => d.task.revision = "2".into(),
            "sender" => d.sender.session_id = "other".into(),
            _ => d.payload["dispatch_payload_digest"] = json!(digest(b"wrong")),
        }
        let before = r.posts();
        assert!(ack(root.path(), &m, d, &r).is_err(), "{field}");
        assert_eq!(r.posts(), before);
    }
    ack(root.path(), &m, answer(&m, "deferred"), &r).unwrap();
    let posts = r.posts();
    assert!(ack(root.path(), &m, answer(&m, "accepted"), &r).is_err());
    assert_eq!(r.posts(), posts);
}
#[test]
fn new_revision_requires_exact_predecessor_and_current_leader_epoch() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    let v = send(root.path(), &m, &r).unwrap();
    for field in ["gap", "digest", "comment", "leader", "owner"] {
        let mut n = next(&m, &v);
        match field {
            "gap" => {
                n.task.revision = "3".into();
                n.payload["revision"] = json!(3)
            }
            "digest" => n.payload["previous"]["payload_digest"] = json!(digest(b"wrong")),
            "comment" => n.payload["previous"]["comment_id"] = json!(999),
            "leader" => n.leader_epoch = 2,
            _ => n.owner_epoch = 0,
        }
        assert!(send(root.path(), &n, &r).is_err(), "{field}");
        assert_eq!(r.posts(), 1);
    }
}
// A repository member's comment (the fixture's `author_association` is `MEMBER`) whose user id is
// not the account the packet names, or whose timestamp is unknown, still closes the read. A
// comment from anyone who is not a member is not read at all since #3067 S3; see
// `packets_from_non_members_are_ignored_and_reported_by_event`.
#[test]
fn forged_author_and_unknown_timestamp_are_not_accepted_observations() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    let pristine = r.comments.borrow().clone();
    r.comments.borrow_mut()[0]["user"]["id"] = json!(99);
    assert!(check(&m, &r).is_err());
    *r.comments.borrow_mut() = pristine.clone();
    r.comments.borrow_mut()[0]["created_at"] = json!("invalid");
    assert!(check(&m, &r).is_err());
    *r.comments.borrow_mut() = pristine;
    assert_eq!(check(&m, &r).unwrap()["queued"], true);
}
#[test]
fn incomplete_or_foreign_pagination_prevents_decision_and_leaves_journals_unchanged() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    let file = root.path().join(".moai-cli/messages/outbox/d1.json");
    let bytes = fs::read(&file).unwrap();
    let posts = r.posts();
    for link in [
        "<https://api.github.com/repos/org/repo/issues/4/comments?page=2>; rel=\"next\"",
        "<https://api.github.com/repos/org/repo/issues/3/comments?per_page=100>; rel=\"next\"",
        "broken; rel=\"next\"",
    ] {
        *r.links.borrow_mut() = vec![(
            "https://api.github.com/repos/org/repo/issues/3/comments?per_page=100".into(),
            link.into(),
        )];
        assert!(ack(root.path(), &m, answer(&m, "accepted"), &r).is_err());
        assert_eq!(r.posts(), posts);
        assert_eq!(fs::read(&file).unwrap(), bytes);
    }
    assert!(!r.sent.borrow().iter().any(|q| q.url.contains("/issues/4/")));
}
#[test]
fn publish_unknown_is_not_permission_to_repeat_or_accept() {
    for lands in [true, false] {
        let root = temp_root();
        let r = Remote::new();
        let m = packet();
        r.lost.set(true);
        r.lands.set(lands);
        let result = send(root.path(), &m, &r);
        assert_eq!(result.is_ok(), lands);
        assert_eq!(r.posts(), 1);
        let result = send(root.path(), &m, &r);
        assert_eq!(result.is_ok(), lands);
        assert_eq!(r.posts(), 1);
    }
}
#[test]
fn audit_key_is_observed_but_never_changes_existing_reservations() {
    let root = temp_root();
    let r = Remote::new();
    let mut m = packet();
    let head = "a".repeat(40);
    m.task.task_id = format!("audit:#9@{head}");
    m.payload["audit"] = json!({"pr":9,"head":head});
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    let v = ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    assert_eq!(v["c3"], "observation_only_gate_integration_unfulfilled");
    assert_eq!(v["claim_authorized"], false);
    assert!(r
        .sent
        .borrow()
        .iter()
        .all(|q| !q.url.contains("/pulls/") && !q.url.contains("/statuses/")));
    m.payload["audit"]["head"] = json!("b".repeat(40));
    assert!(send(root.path(), &m, &r).is_err());
}

#[test]
fn old_accepted_and_duplicate_receipts_do_not_accept_a_new_revision() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    let v = send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    let duplicate = r.comments.borrow()[0].clone();
    r.comments.borrow_mut().push(duplicate);
    r.page_size.set(1);
    r.actor.set(11);
    let n = next(&m, &v);
    let v = send(root.path(), &n, &r).unwrap();
    assert_eq!(v["received_ack"], false);
    assert_eq!(v["disposition"], Value::Null);
    assert_eq!(check(&m, &r).unwrap()["superseded"], true);
}
#[test]
fn conflicting_same_revision_and_gap_are_unknown_not_latest_wins() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    let mut fork = m.clone();
    fork.report_id = "fork".into();
    fork.receiver.session_id = "other".into();
    message::record_cli(
        root.path(),
        message::RecordRequest { message: fork },
        "decision:1",
        &r,
        &Time,
    )
    .unwrap();
    assert!(check(&m, &r).is_err());
    r.actor.set(22);
    let before = r.posts();
    assert!(ack(root.path(), &m, answer(&m, "accepted"), &r).is_err());
    assert_eq!(r.posts(), before);
}
#[test]
fn lost_received_or_decision_response_resumes_only_after_observation() {
    for fail_post in [2, 3] {
        for landed in [false, true] {
            let root = temp_root();
            let r = Remote::new();
            let m = packet();
            send(root.path(), &m, &r).unwrap();
            r.actor.set(22);
            r.lose_n.set(Some(fail_post));
            if !landed {
                r.hide_n.set(Some(fail_post));
            }
            let result = ack(root.path(), &m, answer(&m, "accepted"), &r);
            assert_eq!(result.is_ok(), landed);
            let posts = r.posts();
            let result = ack(root.path(), &m, answer(&m, "accepted"), &r);
            assert_eq!(result.is_ok(), landed);
            assert_eq!(r.posts(), posts);
            if !landed {
                assert_eq!(check(&m, &r).unwrap()["disposition"], Value::Null);
            }
        }
    }
}
#[test]
fn remote_receipt_in_the_future_is_unknown_not_received_success() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    r.comments.borrow_mut()[1]["created_at"] = json!("2026-11-11T01:00:00Z");
    let v = check(&m, &r).unwrap();
    assert_eq!(v["ack_alert"], "unknown");
    assert_eq!(v["observation_current"], false);
}
#[test]
fn malformed_inputs_and_missing_permission_do_not_contact_provider_or_create_files() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    for field in [
        "empty_scope",
        "empty_completion",
        "zero_revision",
        "unknown_field",
        "secret",
    ] {
        let mut bad = m.clone();
        match field {
            "empty_scope" => bad.payload["scope"] = json!([]),
            "empty_completion" => bad.payload["completion"] = json!([" "]),
            "zero_revision" => bad.payload["revision"] = json!(0),
            "unknown_field" => bad.payload["ignored"] = json!(true),
            _ => bad.payload["scope"] = json!([format!("{}{}", "ghp_", "x".repeat(40))]),
        }
        assert!(send(root.path(), &bad, &r).is_err(), "{field}");
        assert!(r.sent.borrow().is_empty());
    }
    assert!(
        dispatch::record_cli(root.path(), DispatchRecord { message: m }, "", &r, &Time).is_err()
    );
    assert!(r.sent.borrow().is_empty());
    assert!(!root.path().join(".moai-cli").exists());
}
#[cfg(unix)]
#[test]
fn native_cli_check_envelope_uses_synthetic_gh_and_never_posts() {
    use std::{
        io::Write,
        os::unix::fs::PermissionsExt,
        process::{Command, Stdio},
    };
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    let bin = root.path().join("gh-fixture");
    let data = root.path().join("data.json");
    fs::write(&data, serde_json::to_vec(&*r.comments.borrow()).unwrap()).unwrap();
    let script = format!(
        r#"#!/usr/bin/env python3
import json,sys
from pathlib import Path
a=sys.argv
assert a[a.index('-X')+1]=='GET',a
u=next(x for x in a if x.startswith('https://'))
if '/comments?' in u:b=json.loads(Path({:?}).read_text())
elif u.endswith('/issues/3'):b={{'id':2,'node_id':'I_2','number':3}}
elif u.endswith('/repos/org/repo'):b={{'id':1}}
else:raise Exception('unexpected fixture URL')
print('HTTP/2 200'+chr(10)+chr(10)+json.dumps(b))
"#,
        data.to_str().unwrap()
    );
    fs::write(&bin, script).unwrap();
    fs::set_permissions(&bin, fs::Permissions::from_mode(0o700)).unwrap();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("dispatch-check")
        .clean_env()
        .env("PATH", "/usr/bin:/bin")
        .env("HOME", root.path())
        .env("MOAI_FORGE_GH_BIN", &bin)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(&serde_json::to_vec(&json!({"dispatch":m})).unwrap())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{} {}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["operation"], "dispatch.check");
    assert_eq!(v["result"]["claim_authorized"], false);
}

#[test]
fn audit_decision_report_id_collision_preserves_s1() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    let v = send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    let d1 = answer(&m, "accepted");
    ack(root.path(), &m, d1.clone(), &r).unwrap();
    r.actor.set(11);
    let m2 = next(&m, &v);
    send(root.path(), &m2, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m2, answer(&m2, "accepted"), &r).unwrap();
    {
        let mut comments = r.comments.borrow_mut();
        let c = comments.last_mut().unwrap();
        c["body"] = json!(c["body"]
            .as_str()
            .unwrap()
            .replace("d2-decision", "d1-decision"));
    }
    let before = journal_bytes(root.path());
    let posts = r.posts();
    let control = message::reconcile_cli(
        root.path(),
        message::ReconcileRequest { message: d1 },
        &r,
        &Time,
    )
    .unwrap_err();
    assert_eq!(
        control.kind,
        moai_cli::contract::Kind::Conflict,
        "S1 control must report identity conflict"
    );
    let observed = check(&m2, &r);
    assert!(
        observed.is_err(),
        "S2 must also reject reused immutable decision ID, observed={observed:?}"
    );
    assert_eq!(r.posts(), posts);
    assert_eq!(journal_bytes(root.path()), before);
}

fn accepted_pair() -> (TempDir, Remote, Message, Message) {
    let root = temp_root();
    let r = Remote::new();
    let one = packet();
    let v = send(root.path(), &one, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &one, answer(&one, "accepted"), &r).unwrap();
    r.actor.set(11);
    let two = next(&one, &v);
    send(root.path(), &two, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &two, answer(&two, "accepted"), &r).unwrap();
    (root, r, one, two)
}
fn rewrite_packet(comment: &mut Value, mut m: Message, ack: bool) {
    m.payload.sort_all_objects();
    let old = comment["body"].as_str().unwrap();
    let mut wire: Value = serde_json::from_str(old.lines().nth(1).unwrap()).unwrap();
    wire["message"] = serde_json::to_value(&m).unwrap();
    wire["payload_digest"] = json!(payload_digest(&m));
    wire["kind"] = json!(if ack { "ack" } else { "record" });
    let marker = moai_cli::forge::outbox::op_marker(
        if ack { "message-ack" } else { "message-record" },
        &m.project_id,
        &m.report_id,
        1,
    );
    comment["body"] = json!(format!(
        "{marker}\n{}\n<!-- moai:message:end -->",
        serde_json::to_string(&wire).unwrap()
    ));
    comment["user"]["id"] = json!(if ack {
        m.receiver.actor_id
    } else {
        m.sender.actor_id
    });
}
#[test]
fn every_skipped_message_kind_preserves_relevant_report_id_identity() {
    let mut missed = Vec::new();
    for kind in [
        "decision-record",
        "decision-ack",
        "generic-event",
        "other-task",
        "other-project",
    ] {
        let (root, r, one, two) = accepted_pair();
        let d1 = answer(&one, "accepted");
        let mut reused = answer(&two, "accepted");
        reused.report_id = d1.report_id.clone();
        match kind {
            "generic-event" => reused.event = "report".into(),
            "other-task" => reused.task.task_id = "another-task".into(),
            "other-project" => reused.project_id = "another-project".into(),
            _ => {}
        }
        {
            let mut comments = r.comments.borrow_mut();
            rewrite_packet(comments.last_mut().unwrap(), reused, kind == "decision-ack");
        }
        let posts = r.posts();
        let control = message::reconcile_cli(
            root.path(),
            message::ReconcileRequest { message: d1 },
            &r,
            &Time,
        )
        .unwrap_err();
        assert_eq!(
            control.kind,
            moai_cli::contract::Kind::Conflict,
            "S1: {kind}"
        );
        match check(&two, &r) {
            Err(e) => assert_eq!(e.kind, moai_cli::contract::Kind::Conflict, "S2: {kind}"),
            Ok(v) => missed.push((kind, v)),
        }
        assert_eq!(r.posts(), posts, "observation must not POST: {kind}");
    }
    assert!(
        missed.is_empty(),
        "identity checks skipped message kinds: {missed:?}"
    );
}
#[test]
fn unrelated_generic_packets_do_not_block_valid_dispatch_observation() {
    let (_root, r, _, two) = accepted_pair();
    let mut m = answer(&two, "accepted");
    m.report_id = "unrelated".into();
    m.event = "report".into();
    for other in ["one-unrelated-task", "another-unrelated-task"] {
        m.task.task_id = other.into();
        let mut c = r.comments.borrow().last().unwrap().clone();
        c["id"] = json!(1000 + r.comments.borrow().len());
        c["created_at"] = json!("2099-01-01T00:00:00Z");
        rewrite_packet(&mut c, m.clone(), false);
        r.comments.borrow_mut().push(c);
    }
    r.comments
        .borrow_mut()
        .push(json!({"id":9001,"author_association":"MEMBER","body":"HOLD: preserve the user's draft","user":{"id":99}}));
    r.comments.borrow_mut().push(
        json!({"id":9002,"author_association":"MEMBER","body":"<!-- moai-op other-operation:p1:d2:1 -->","user":{"id":99}}),
    );
    let v = check(&two, &r).unwrap();
    assert_eq!(v["disposition"], "accepted");
    assert_eq!(v["observation_current"], true);
}

fn journal_bytes(root: &Path) -> Vec<(std::path::PathBuf, Vec<u8>)> {
    let mut files = Vec::new();
    for dir in ["outbox", "inbox"] {
        for p in fs::read_dir(root.join(".moai-cli/messages").join(dir)).unwrap() {
            let p = p.unwrap().path();
            files.push((
                p.strip_prefix(root).unwrap().to_path_buf(),
                fs::read(p).unwrap(),
            ));
        }
    }
    files.sort_by(|a, b| a.0.cmp(&b.0));
    files
}
#[test]
fn earlier_page_foreign_identity_collision_is_not_lost_before_discovery() {
    let (root, r, one, two) = accepted_pair();
    let d1 = answer(&one, "accepted");
    let mut foreign = d1.clone();
    foreign.project_id = "foreign-project".into();
    foreign.event = "report".into();
    let mut c = r.comments.borrow()[0].clone();
    c["id"] = json!(1);
    rewrite_packet(&mut c, foreign, false);
    r.comments.borrow_mut().insert(0, c);
    r.page_size.set(1);
    assert_eq!(
        message::reconcile_cli(
            root.path(),
            message::ReconcileRequest { message: d1 },
            &r,
            &Time
        )
        .unwrap_err()
        .kind,
        moai_cli::contract::Kind::Conflict
    );
    assert_eq!(
        check(&two, &r).unwrap_err().kind,
        moai_cli::contract::Kind::Conflict
    );
}
#[test]
fn decision_receipts_need_their_record_and_never_create_business_acceptance() {
    for orphan in [false, true] {
        let (root, r, _, two) = accepted_pair();
        let mut d = answer(&two, "accepted");
        if orphan {
            d.report_id = "orphan-decision".into();
        }
        let mut c = r.comments.borrow().last().unwrap().clone();
        c["id"] = json!(1000);
        rewrite_packet(&mut c, d.clone(), true);
        r.comments.borrow_mut().push(c);
        let posts = r.posts();
        let bytes = journal_bytes(root.path());
        if orphan {
            assert_eq!(
                message::reconcile_cli(
                    root.path(),
                    message::ReconcileRequest { message: d },
                    &r,
                    &Time
                )
                .unwrap_err()
                .kind,
                moai_cli::contract::Kind::Conflict
            );
            assert_eq!(
                check(&two, &r).unwrap_err().kind,
                moai_cli::contract::Kind::Conflict
            );
        } else {
            let v = check(&two, &r).unwrap();
            assert_eq!(v["disposition"], "accepted");
            assert_eq!(v["observation_current"], true);
        }
        assert_eq!(r.posts(), posts);
        assert_eq!(journal_bytes(root.path()), bytes);
    }
}

// The fake remote stamps every comment 2026-11-11T00:00:00Z, which is `Time.now()`.
const BASE: i64 = 1_794_355_200;
fn at(secs: i64) -> Value {
    json!(format!("2026-11-11T00:{:02}:{:02}Z", secs / 60, secs % 60))
}

#[test]
fn dispatch_times_are_separate_values() {
    let root = temp_root();
    let r = Remote::new();
    let mut m = packet();
    // A time written inside a message is a claim by its author, never the observed time.
    m.created_at = "2099-01-01T00:00:00Z".into();
    send(root.path(), &m, &r).unwrap();
    let queued = check(&m, &r).unwrap();
    assert_eq!(queued["sent_at"], BASE);
    for key in ["received_at", "ack_comment_id", "decided_at"] {
        assert_eq!(queued[key], Value::Null, "{key}");
    }
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    {
        let mut comments = r.comments.borrow_mut();
        comments[1]["created_at"] = at(40);
        comments[2]["created_at"] = at(41);
    }
    let v = dispatch::check_cli(
        CheckRequest {
            dispatch: m.clone(),
        },
        &r,
        &At(BASE + 400),
    )
    .unwrap();
    assert_eq!(v["sent_at"], BASE);
    assert_eq!(v["received_at"], BASE + 40);
    assert_eq!(v["decided_at"], BASE + 41);
    assert_eq!(v["record_comment_id"], 100);
    assert_eq!(v["ack_comment_id"], 101);
    assert_eq!(v["decision_comment_id"], 102);
    assert_eq!(v["disposition"], "accepted");
    // The earlier keys keep their meaning.
    assert_eq!(v["received_ack"], true);
    assert_eq!(v["ack_alert"], "received");
    assert_eq!(v["claim_authorized"], false);
    assert_eq!(v["ownership_transferred"], false);
    assert_eq!(v["audit_gate_integrated"], false);

    // A duplicate ACK keeps the lowest comment id, even when a later comment is earlier in time.
    let mut dup = r.comments.borrow()[1].clone();
    dup["id"] = json!(50);
    dup["created_at"] = at(45);
    r.comments.borrow_mut().push(dup);
    let v = dispatch::check_cli(
        CheckRequest {
            dispatch: m.clone(),
        },
        &r,
        &At(BASE + 400),
    )
    .unwrap();
    assert_eq!(v["ack_comment_id"], 50);
    assert_eq!(v["received_at"], BASE + 45);

    // An ACK is not a decision.
    let root = temp_root();
    let r = Remote::new();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    message::ack_cli(
        root.path(),
        message::AckRequest {
            message: m.clone(),
            payload_digest: payload_digest(&m),
        },
        "decision:1",
        &r,
        &Time,
    )
    .unwrap();
    r.comments.borrow_mut()[1]["created_at"] = at(40);
    let v = dispatch::check_cli(
        CheckRequest {
            dispatch: m.clone(),
        },
        &r,
        &At(BASE + 400),
    )
    .unwrap();
    assert_eq!(v["received_at"], BASE + 40);
    assert_eq!(v["decided_at"], Value::Null);
    assert_eq!(v["decision_comment_id"], Value::Null);

    // An ACK observed before its own record is a conflict, not a negative duration.
    r.comments.borrow_mut()[0]["created_at"] = at(30);
    r.comments.borrow_mut()[1]["created_at"] = at(20);
    let e = dispatch::check_cli(CheckRequest { dispatch: m }, &r, &At(BASE + 400)).unwrap_err();
    assert_eq!(e.kind, moai_cli::contract::Kind::Conflict);
}

fn start_of(m: &Message) -> Message {
    let mut s = answer(m, "accepted");
    s.report_id = format!("{}-start", m.report_id);
    s.event = "dispatch-start".into();
    s.payload = json!({"dispatch_report_id":m.report_id,"dispatch_payload_digest":payload_digest(m),"note":"work begun"});
    s
}
// The writer command arrives in a later slice; the fold is read through the generic S1 record.
fn post(root: &Path, m: &Message, r: &Remote) {
    message::record_cli(
        root,
        message::RecordRequest { message: m.clone() },
        "decision:1",
        r,
        &Time,
    )
    .unwrap();
}
fn later(m: &Message, r: &Remote) -> Result<Value, moai_cli::contract::OpError> {
    dispatch::check_cli(
        CheckRequest {
            dispatch: m.clone(),
        },
        r,
        &At(BASE + 400),
    )
}
fn conflict_kind(result: Result<Value, moai_cli::contract::OpError>) -> moai_cli::contract::Kind {
    result.unwrap_err().kind
}

#[test]
fn dispatch_start_requires_accepted_current_revision() {
    use moai_cli::contract::Kind;
    // Accepted, current: the start is folded with its own server time and comment id.
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    let before = later(&m, &r).unwrap();
    assert_eq!(before["started"], false);
    assert_eq!(before["started_at"], Value::Null);
    assert_eq!(before["start_comment_id"], Value::Null);
    post(root.path(), &start_of(&m), &r);
    // A start in the same second as the decision is allowed.
    assert_eq!(later(&m, &r).unwrap()["started"], true);
    r.comments.borrow_mut()[3]["created_at"] = at(60);
    let v = later(&m, &r).unwrap();
    assert_eq!(v["started"], true);
    assert_eq!(v["started_at"], BASE + 60);
    assert_eq!(v["start_comment_id"], 103);
    assert_eq!(v["decided_at"], BASE);
    assert_eq!(v["claim_authorized"], false);
    assert_eq!(v["ownership_transferred"], false);
    assert_eq!(v["audit_gate_integrated"], false);
    // A start before the decision is a conflict, not a negative duration.
    r.comments.borrow_mut()[2]["created_at"] = at(61);
    assert_eq!(conflict_kind(later(&m, &r)), Kind::Conflict);

    // Deferred or rejected decisions cannot start.
    for kind in ["deferred", "rejected"] {
        let root = temp_root();
        let r = Remote::new();
        send(root.path(), &m, &r).unwrap();
        r.actor.set(22);
        ack(root.path(), &m, answer(&m, kind), &r).unwrap();
        post(root.path(), &start_of(&m), &r);
        assert_eq!(conflict_kind(later(&m, &r)), Kind::Conflict, "{kind}");
    }

    // No ACK at all, and an ACK without a decision, cannot start either.
    let root = temp_root();
    let r = Remote::new();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    post(root.path(), &start_of(&m), &r);
    assert_eq!(conflict_kind(later(&m, &r)), Kind::Conflict);
    let root = temp_root();
    let r = Remote::new();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    message::ack_cli(
        root.path(),
        message::AckRequest {
            message: m.clone(),
            payload_digest: payload_digest(&m),
        },
        "decision:1",
        &r,
        &Time,
    )
    .unwrap();
    post(root.path(), &start_of(&m), &r);
    assert_eq!(conflict_kind(later(&m, &r)), Kind::Conflict);

    // A start that was current when written stays readable after a newer revision; one written
    // after the newer record is a conflict.
    let root = temp_root();
    let r = Remote::new();
    let v = send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    post(root.path(), &start_of(&m), &r);
    r.actor.set(11);
    let n = next(&m, &v);
    send(root.path(), &n, &r).unwrap();
    let new = later(&n, &r).unwrap();
    assert_eq!(new["started"], false);
    let old = later(&m, &r).unwrap();
    assert_eq!(old["superseded"], true);
    assert_eq!(old["started"], false);
    r.comments.borrow_mut()[3]["created_at"] = at(60);
    assert_eq!(conflict_kind(later(&n, &r)), Kind::Conflict);
    assert_eq!(conflict_kind(later(&m, &r)), Kind::Conflict);

    // The payload schema is closed: an unknown key is invalid, a blank note is a conflict.
    // Only the receiver of the dispatch can start it, never its sender.
    for field in [
        "unknown_key",
        "blank_note",
        "wrong_digest",
        "sender_starts",
        "receiver_wrong",
    ] {
        let root = temp_root();
        let r = Remote::new();
        send(root.path(), &m, &r).unwrap();
        r.actor.set(22);
        ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
        let mut s = start_of(&m);
        match field {
            "unknown_key" => s.payload["extra"] = json!(true),
            "blank_note" => s.payload["note"] = json!(" "),
            "sender_starts" => {
                s.sender = m.sender.clone();
                r.actor.set(11);
            }
            "receiver_wrong" => s.receiver = m.receiver.clone(),
            _ => s.payload["dispatch_payload_digest"] = json!(digest(b"wrong")),
        }
        post(root.path(), &s, &r);
        let want = if field == "unknown_key" {
            Kind::InvalidInput
        } else {
            Kind::Conflict
        };
        assert_eq!(conflict_kind(later(&m, &r)), want, "{field}");
    }
}

#[test]
fn recording_the_same_dispatch_twice_leaves_one_row() {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    post(root.path(), &start_of(&m), &r);
    let once = later(&m, &r).unwrap();
    assert_eq!(once["start_comment_id"], 103);
    // The same packets observed again under other comment ids: one row, lowest id wins.
    for (index, id, secs) in [(3, 500, 90), (0, 600, 5), (1, 700, 6)] {
        let mut dup = r.comments.borrow()[index].clone();
        dup["id"] = json!(id);
        dup["created_at"] = at(secs);
        r.comments.borrow_mut().push(dup);
    }
    let v = later(&m, &r).unwrap();
    assert_eq!(v["start_comment_id"], 103);
    assert_eq!(v["started_at"], BASE);
    assert_eq!(v["record_comment_id"], 100);
    assert_eq!(v["ack_comment_id"], 101);
    assert_eq!(v["decision_comment_id"], 102);
    assert_eq!(v["sent_at"], BASE);
    // A lower id wins even when its server time is later.
    let mut dup = r.comments.borrow()[3].clone();
    dup["id"] = json!(50);
    dup["created_at"] = at(70);
    r.comments.borrow_mut().push(dup);
    let v = later(&m, &r).unwrap();
    assert_eq!(v["start_comment_id"], 50);
    assert_eq!(v["started_at"], BASE + 70);
    // A second, different start packet for the same dispatch is a conflict, not a second row.
    let mut other = start_of(&m);
    other.report_id = "d1-start-other".into();
    other.payload["note"] = json!("a different packet");
    post(root.path(), &other, &r);
    assert_eq!(
        conflict_kind(later(&m, &r)),
        moai_cli::contract::Kind::Conflict
    );
}

fn begin(
    root: &Path,
    m: &Message,
    s: Message,
    r: &Remote,
) -> Result<Value, moai_cli::contract::OpError> {
    dispatch::start_cli(
        root,
        dispatch::StartRequest {
            dispatch: m.clone(),
            start: s,
        },
        "decision:1",
        r,
        &Time,
    )
}
fn outbox_has(root: &Path, report_id: &str) -> bool {
    root.join(".moai-cli/messages/outbox")
        .join(format!("{report_id}.json"))
        .exists()
}
fn accepted_world() -> (TempDir, Remote, Message) {
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    r.actor.set(22);
    ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    (root, r, m)
}

#[test]
fn dispatch_start_write_is_idempotent() {
    let (root, r, m) = accepted_world();
    assert_eq!(r.posts(), 3);
    let s = start_of(&m);
    let v = begin(root.path(), &m, s.clone(), &r).unwrap();
    assert_eq!(v["started"], true);
    assert_eq!(v["start_comment_id"], 103);
    assert_eq!(v["started_at"], BASE);
    assert_eq!(v["disposition"], "accepted");
    assert_eq!(v["claim_authorized"], false);
    assert_eq!(v["ownership_transferred"], false);
    assert_eq!(v["audit_gate_integrated"], false);
    assert_eq!(r.posts(), 4);
    assert!(outbox_has(root.path(), "d1-start"));
    // The same request again, in this folder and in a fresh one: no POST, one outbox file,
    // one row, and the first start time.
    let restarted = temp_root();
    for folder in [root.path(), restarted.path()] {
        let again = begin(folder, &m, s.clone(), &r).unwrap();
        assert_eq!(again["started_at"], v["started_at"]);
        assert_eq!(again["start_comment_id"], 103);
        assert_eq!(r.posts(), 4);
        assert_eq!(r.comments.borrow().len(), 4);
        assert!(outbox_has(folder, "d1-start"));
    }
    let starts = fs::read_dir(root.path().join(".moai-cli/messages/outbox"))
        .unwrap()
        .filter(|e| {
            e.as_ref()
                .unwrap()
                .file_name()
                .to_string_lossy()
                .contains("start")
        })
        .count();
    assert_eq!(starts, 1);

    // A lost response is resumed only after observation, and never posts twice.
    for landed in [false, true] {
        let (root, r, m) = accepted_world();
        r.lose_n.set(Some(4));
        if !landed {
            r.hide_n.set(Some(4));
        }
        let first = begin(root.path(), &m, start_of(&m), &r);
        assert_eq!(first.is_ok(), landed);
        let posts = r.posts();
        let second = begin(root.path(), &m, start_of(&m), &r);
        assert_eq!(second.is_ok(), landed);
        assert_eq!(r.posts(), posts);
        assert_eq!(check(&m, &r).unwrap()["started"], landed, "landed={landed}");
    }
}

#[test]
fn dispatch_start_refuses_before_ack_or_decision() {
    use moai_cli::contract::Kind;
    for case in ["no_ack", "ack_only", "deferred", "rejected"] {
        let root = temp_root();
        let r = Remote::new();
        let m = packet();
        send(root.path(), &m, &r).unwrap();
        r.actor.set(22);
        match case {
            "no_ack" => {}
            "ack_only" => {
                message::ack_cli(
                    root.path(),
                    message::AckRequest {
                        message: m.clone(),
                        payload_digest: payload_digest(&m),
                    },
                    "decision:1",
                    &r,
                    &Time,
                )
                .unwrap();
            }
            kind => {
                ack(root.path(), &m, answer(&m, kind), &r).unwrap();
            }
        }
        let posts = r.posts();
        let e = begin(root.path(), &m, start_of(&m), &r).unwrap_err();
        assert_eq!(e.kind, Kind::Blocked, "{case}");
        assert_eq!(r.posts(), posts, "{case}");
        assert!(!outbox_has(root.path(), "d1-start"), "{case}");
    }

    // A superseded revision cannot be started; the new one has no ACK yet.
    let (root, r, m) = accepted_world();
    r.actor.set(11);
    let v = check(&m, &r).unwrap();
    let n = next(&m, &v);
    send(root.path(), &n, &r).unwrap();
    r.actor.set(22);
    let posts = r.posts();
    assert_eq!(
        begin(root.path(), &m, start_of(&m), &r).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(
        begin(root.path(), &n, start_of(&n), &r).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(r.posts(), posts);
    // Even when the newer revision is accepted, the old one stays blocked.
    ack(root.path(), &n, answer(&n, "accepted"), &r).unwrap();
    let posts = r.posts();
    assert_eq!(
        begin(root.path(), &m, start_of(&m), &r).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(r.posts(), posts);
    assert!(!outbox_has(root.path(), "d1-start"));
    assert_eq!(
        begin(root.path(), &n, start_of(&n), &r).unwrap()["started"],
        true
    );

    // A malformed or foreign start never reaches the provider.
    let (root, r, m) = accepted_world();
    for field in ["sender", "receiver", "digest", "event", "note"] {
        let mut s = start_of(&m);
        match field {
            "sender" => s.sender = m.sender.clone(),
            "receiver" => s.receiver = m.receiver.clone(),
            "digest" => s.payload["dispatch_payload_digest"] = json!(digest(b"wrong")),
            "event" => s.event = "dispatch-decision".into(),
            _ => s.payload["note"] = json!(" "),
        }
        let sent = r.sent.borrow().len();
        assert!(begin(root.path(), &m, s, &r).is_err(), "{field}");
        assert_eq!(r.sent.borrow().len(), sent, "{field}");
        assert!(!outbox_has(root.path(), "d1-start"), "{field}");
    }

    // A second, different start packet for the same dispatch is refused before any POST.
    begin(root.path(), &m, start_of(&m), &r).unwrap();
    let posts = r.posts();
    let mut other = start_of(&m);
    other.report_id = "d1-start-other".into();
    other.payload["note"] = json!("a different packet");
    assert_eq!(
        begin(root.path(), &m, other, &r).unwrap_err().kind,
        Kind::Conflict
    );
    assert_eq!(r.posts(), posts);
    assert!(!outbox_has(root.path(), "d1-start-other"));
}

#[test]
fn dispatch_start_requires_permission_reference() {
    let (root, r, m) = accepted_world();
    let sent = r.sent.borrow().len();
    let too_long = "x".repeat(moai_cli::contract::PERMISSION_REFERENCE_MAX_BYTES + 1);
    for permission in ["", " ", "null", "placeholder", "a\nb", too_long.as_str()] {
        let result = dispatch::start_cli(
            root.path(),
            dispatch::StartRequest {
                dispatch: m.clone(),
                start: start_of(&m),
            },
            permission,
            &r,
            &Time,
        );
        assert_eq!(
            result.unwrap_err().kind,
            moai_cli::contract::Kind::PermissionReferenceMissing,
            "{permission:?}"
        );
        assert_eq!(r.sent.borrow().len(), sent, "{permission:?}");
    }
    assert!(!outbox_has(root.path(), "d1-start"));
    // The binary refuses the missing flag before it reads stdin or contacts anything.
    #[cfg(unix)]
    {
        use std::process::{Command, Stdio};
        let out = Command::new(crate::common::MOAI_BIN)
            .args(["dispatch-start", "."])
            .clean_env()
            .env("HOME", root.path())
            .current_dir(root.path())
            .stdin(Stdio::null())
            .output()
            .unwrap();
        assert!(!out.status.success());
        assert!(String::from_utf8_lossy(&out.stderr).contains("--permission-reference"));
    }
}

#[test]
fn ack_alert_boundary_uses_the_shared_constant() {
    let n = dispatch::ACK_ALERT_SECS;
    assert_eq!(n, 300);
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    for (offset, state) in [
        (n - 1, "waiting"),
        (n, "unacknowledged"),
        (n + 1, "unacknowledged"),
        (-1, "unknown"),
    ] {
        let v = dispatch::check_cli(
            CheckRequest {
                dispatch: m.clone(),
            },
            &r,
            &At(BASE + offset),
        )
        .unwrap();
        assert_eq!(v["ack_alert"], state, "offset {offset}");
        assert_eq!(
            v["elapsed_secs"],
            if offset < 0 {
                Value::Null
            } else {
                json!(offset)
            }
        );
    }
}

// Issue #3067 S3: only comments a repository member wrote are read as dispatch packets. A forged
// packet from anyone else is left out and reported, so it can neither join the observation nor
// close the read for the lanes that depend on it.

/// A copy of one recorded comment as another account wrote it: a new id, a foreign user id and
/// the given `author_association` (`None` leaves the field out).
fn forged(c: &Value, id: u64, association: Option<&str>) -> Value {
    let mut f = c.clone();
    f["id"] = json!(id);
    f["user"]["id"] = json!(99);
    match association {
        Some(a) => f["author_association"] = json!(a),
        None => {
            f.as_object_mut().unwrap().remove("author_association");
        }
    }
    f
}
/// The dispatch view without the report of what the author filter left out.
fn without_ignored(v: &Value) -> Value {
    let mut v = v.clone();
    v.as_object_mut().unwrap().remove("ignored_control_lines");
    v
}
/// A world with the record, the ACK, the accepted decision and the start (comments 100..=103).
fn started_world() -> (TempDir, Remote, Message) {
    let (root, r, m) = accepted_world();
    begin(root.path(), &m, start_of(&m), &r).unwrap();
    assert_eq!(r.comments.borrow().len(), 4);
    (root, r, m)
}

#[test]
fn packets_from_non_members_are_ignored_and_reported_by_event() {
    for association in [
        Some("NONE"),
        Some("CONTRIBUTOR"),
        Some("FIRST_TIMER"),
        Some("FIRST_TIME_CONTRIBUTOR"),
        Some("MANNEQUIN"),
        Some("member"),
        Some(""),
        None,
    ] {
        let (root, r, m) = started_world();
        let clean = check(&m, &r).unwrap();
        assert_eq!(clean["started"], true);
        assert!(clean.get("ignored_control_lines").is_none(), "{clean}");
        // A forged copy of every packet (dispatch, ACK, decision, start), before and after.
        let real = r.comments.borrow().clone();
        let mut all = Vec::new();
        for (i, c) in real.iter().enumerate() {
            all.push(forged(c, 5000 + i as u64, association));
        }
        all.extend(real.iter().cloned());
        for (i, c) in real.iter().enumerate() {
            all.push(forged(c, 6000 + i as u64, association));
        }
        *r.comments.borrow_mut() = all;
        let posts = r.posts();
        // The read, a repeated start and a repeated decision all go through.
        let v = check(&m, &r).unwrap();
        assert_eq!(without_ignored(&v), clean, "{association:?}");
        let ignored = &v["ignored_control_lines"];
        assert_eq!(ignored["count"], 8, "{association:?}: {v}");
        assert_eq!(
            ignored["by_event"],
            json!({"dispatch":4,"dispatch-decision":2,"dispatch-start":2}),
            "{association:?}: {v}"
        );
        assert_eq!(ignored["reason"], "no_member_authority");
        assert_eq!(ignored["other_non_member_comments"], 0);
        assert_eq!(ignored["items"].as_array().unwrap().len(), 8);
        assert_eq!(
            ignored["items"][0]["author_association"],
            association.unwrap_or("")
        );
        let again = begin(root.path(), &m, start_of(&m), &r).unwrap();
        assert_eq!(without_ignored(&again), clean, "{association:?}");
        assert_eq!(again["ignored_control_lines"]["count"], 8);
        let decided = ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
        assert_eq!(without_ignored(&decided), clean, "{association:?}");
        assert_eq!(r.posts(), posts, "{association:?}: nothing was posted");
    }
}

#[test]
fn a_forged_packet_before_the_real_one_cannot_block_the_next_steps() {
    // The forged dispatch comes first on the first page; the real flow still runs to the end.
    let root = temp_root();
    let r = Remote::new();
    let m = packet();
    send(root.path(), &m, &r).unwrap();
    let real = r.comments.borrow()[0].clone();
    r.comments
        .borrow_mut()
        .insert(0, forged(&real, 4000, Some("NONE")));
    r.actor.set(22);
    let v = ack(root.path(), &m, answer(&m, "accepted"), &r).unwrap();
    assert_eq!(v["disposition"], "accepted");
    assert_eq!(v["ignored_control_lines"]["count"], 1, "{v}");
    let v = begin(root.path(), &m, start_of(&m), &r).unwrap();
    assert_eq!(v["started"], true);
    assert_eq!(
        v["ignored_control_lines"]["by_event"],
        json!({"dispatch":1})
    );
}

#[test]
fn members_stay_authoritative_and_a_member_with_a_foreign_user_still_closes_the_read() {
    for association in ["OWNER", "MEMBER", "COLLABORATOR"] {
        // The same account writing the same packet again, under a later id, is one row.
        let (_root, r, m) = started_world();
        let clean = check(&m, &r).unwrap();
        let mut same = r.comments.borrow()[3].clone();
        same["id"] = json!(5000);
        same["author_association"] = json!(association);
        r.comments.borrow_mut().push(same);
        let v = check(&m, &r).unwrap();
        assert_eq!(v, clean, "{association}");
        assert!(v.get("ignored_control_lines").is_none(), "{v}");
        // A member association does not make another user id acceptable: the packet's author
        // must still be the account the packet names, and the read stays closed.
        let (_root, r, m) = started_world();
        let other = forged(&r.comments.borrow()[3], 5000, Some(association));
        r.comments.borrow_mut().push(other);
        assert!(check(&m, &r).is_err(), "{association}");
    }
}

#[test]
fn comments_from_only_non_members_read_as_no_record() {
    use moai_cli::contract::Kind;
    let (root, r, m) = started_world();
    for c in r.comments.borrow_mut().iter_mut() {
        c["author_association"] = json!("NONE");
    }
    let posts = r.posts();
    let e = check(&m, &r).unwrap_err();
    assert_eq!(e.kind, Kind::Blocked, "{e:?}");
    assert!(begin(root.path(), &m, start_of(&m), &r).is_err());
    assert!(ack(root.path(), &m, answer(&m, "accepted"), &r).is_err());
    assert_eq!(r.posts(), posts, "nothing was posted");
}

#[test]
fn the_filter_covers_every_page_and_the_counts_add_up() {
    let (_root, r, m) = started_world();
    let clean = check(&m, &r).unwrap();
    let real = r.comments.borrow().clone();
    // Five forged comments among the four real ones, two per page: page 1 holds forged ones
    // only, a forged copy sits at every page edge.
    let mut all = vec![
        forged(&real[0], 5000, Some("NONE")),
        forged(&real[3], 5001, Some("CONTRIBUTOR")),
    ];
    all.push(real[0].clone());
    all.push(forged(&real[2], 5002, Some("NONE")));
    all.push(real[1].clone());
    all.push(real[2].clone());
    all.push(forged(&real[1], 5003, Some("FIRST_TIMER")));
    all.push(real[3].clone());
    all.push(forged(&real[3], 5004, None));
    *r.comments.borrow_mut() = all;
    r.page_size.set(2);
    let before = r.sent.borrow().len();
    let v = check(&m, &r).unwrap();
    assert_eq!(without_ignored(&v), clean);
    let ignored = &v["ignored_control_lines"];
    assert_eq!(ignored["count"], 5, "{v}");
    assert_eq!(
        ignored["by_event"],
        json!({"dispatch":2,"dispatch-decision":1,"dispatch-start":2}),
        "{v}"
    );
    let pages = r.sent.borrow()[before..]
        .iter()
        .filter(|q| q.url.contains("/comments?"))
        .count();
    assert!(pages >= 5, "every page was read: {pages}");
}

#[test]
fn a_non_member_packet_that_does_not_parse_is_reported_not_fatal() {
    use moai_cli::contract::Kind;
    let (_root, r, m) = started_world();
    let clean = check(&m, &r).unwrap();
    let junk = |id: u64, association: &str, body: &str| {
        json!({"id":id,"author_association":association,"body":body,"user":{"id":99},
            "created_at":"2026-11-11T00:00:00Z"})
    };
    r.comments.borrow_mut().push(junk(
        7000,
        "NONE",
        "<!-- moai:op message-record:p1:zz:1 -->\nnot json",
    ));
    r.comments
        .borrow_mut()
        .push(junk(7001, "CONTRIBUTOR", "just a remark"));
    let v = check(&m, &r).unwrap();
    assert_eq!(without_ignored(&v), clean);
    assert_eq!(v["ignored_control_lines"]["count"], 1, "{v}");
    assert_eq!(
        v["ignored_control_lines"]["by_event"],
        json!({"unreadable":1})
    );
    assert_eq!(v["ignored_control_lines"]["other_non_member_comments"], 1);
    // The same junk from a member is still a conflict: the read stays closed for members.
    r.comments.borrow_mut()[4]["author_association"] = json!("MEMBER");
    assert_eq!(conflict_kind(check(&m, &r)), Kind::Conflict);
}
