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
                self.comments.borrow_mut().push(json!({"id":id,"body":request.body.as_ref().unwrap()["body"],"user":{"id":self.actor.get()},"created_at":"2026-11-11T00:00:00Z"}));
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
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("dispatch-check")
        .env_clear()
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
        .push(json!({"id":9001,"body":"HOLD: preserve the user's draft","user":{"id":99}}));
    r.comments.borrow_mut().push(
        json!({"id":9002,"body":"<!-- moai-op other-operation:p1:d2:1 -->","user":{"id":99}}),
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
