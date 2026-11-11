use moai_cli::{
    contract::Kind,
    digest,
    forge::{
        policy::Clock,
        transport::{HttpRequest, HttpResponse, Transport},
    },
    message::{self, AckRequest, Message, ReconcileRequest, RecordRequest},
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
            if self.lands.get() {
                self.comments.borrow_mut().push(json!({"id":id,"body":request.body.as_ref().unwrap()["body"],"user":{"id":self.actor.get()}}));
            }
            if self.lost.replace(false) {
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
fn message() -> Message {
    serde_json::from_value(json!({"schema_version":1,"report_id":"r1","project_id":"p1",
        "task":{"owner":"org","repo":"repo","repository_id":1,"issue_id":2,"issue_node_id":"I_2","issue_number":3,"task_id":"T1","revision":"rev1"},
        "sender":{"actor_id":11,"session_id":"sender"},"receiver":{"actor_id":22,"session_id":"receiver"},
        "owner_epoch":1,"leader_epoch":1,"event":"report","candidate_digest":digest(b"candidate"),"created_at":"2026-11-11T00:00:00Z",
        "approval_refs":["issue-comment:10"],"hold_refs":["issue-comment:9"],"payload":{"text":"hello"}})).unwrap()
}
fn record(root: &Path, m: &Message, remote: &Remote) -> Result<Value, moai_cli::contract::OpError> {
    message::record_cli(
        root,
        RecordRequest { message: m.clone() },
        "decision:1",
        remote,
        &Time,
    )
}
fn reconcile(
    root: &Path,
    m: &Message,
    remote: &Remote,
) -> Result<Value, moai_cli::contract::OpError> {
    message::reconcile_cli(root, ReconcileRequest { message: m.clone() }, remote, &Time)
}
fn ack(root: &Path, m: &Message, remote: &Remote) -> Result<Value, moai_cli::contract::OpError> {
    message::ack_cli(
        root,
        AckRequest {
            message: m.clone(),
            payload_digest: digest(br#"{"text":"hello"}"#),
        },
        "decision:1",
        remote,
        &Time,
    )
}
fn entry(root: &Path, inbox: bool) -> Value {
    serde_json::from_slice(
        &fs::read(root.join(if inbox {
            ".moai-cli/messages/inbox/r1.json"
        } else {
            ".moai-cli/messages/outbox/r1.json"
        }))
        .unwrap(),
    )
    .unwrap()
}
fn rewrite_wire(comment: &mut Value, f: impl FnOnce(&mut Value)) {
    let text = comment["body"].as_str().unwrap();
    let lines: Vec<_> = text.split('\n').collect();
    let mut wire: Value = serde_json::from_str(lines[1]).unwrap();
    f(&mut wire);
    comment["body"] = json!(format!("{}\n{}\n{}", lines[0], wire, lines[2]));
}

fn assert_pagination_rejected(
    links: &[(&str, &str)],
    kind: Kind,
    forbidden: Option<&str>,
    expected_gets: usize,
) {
    for explicit_ack in [true, false] {
        let root = temp_root();
        let receiver = temp_root();
        let remote = Remote::new();
        let m = message();
        record(root.path(), &m, &remote).unwrap();
        remote.actor.set(22);
        // Produce a genuine receiver ACK, available only on the second page.
        ack(receiver.path(), &m, &remote).unwrap();
        remote.page_size.set(1);
        *remote.links.borrow_mut() = links
            .iter()
            .map(|(url, link)| (url.to_string(), link.to_string()))
            .collect();
        remote.comments_gets.set(0);
        remote.comments_get_limit.set(Some(2));
        remote.sent.borrow_mut().clear();
        let outbox = root.path().join(".moai-cli/messages/outbox/r1.json");
        let before = fs::read(&outbox).unwrap();
        let comments = remote.comments.borrow().clone();
        let error = if explicit_ack {
            ack(root.path(), &m, &remote)
        } else {
            reconcile(root.path(), &m, &remote)
        }
        .unwrap_err();
        assert_eq!(error.kind, kind);
        assert_eq!(remote.comments_gets.get(), expected_gets);
        if let Some(target) = forbidden {
            assert_eq!(
                remote
                    .sent
                    .borrow()
                    .iter()
                    .filter(|request| request.method == "GET" && request.url == target)
                    .count(),
                0
            );
        }
        assert_eq!(remote.posts(), 0);
        assert_eq!(fs::read(&outbox).unwrap(), before);
        assert!(!root
            .path()
            .join(".moai-cli/messages/inbox/r1.json")
            .exists());
        assert_eq!(*remote.comments.borrow(), comments);
        if !explicit_ack {
            assert!(remote
                .sent
                .borrow()
                .iter()
                .all(|request| request.method == "GET"));
        }
    }
}

#[test]
fn ack_and_reconcile_refuse_next_links_outside_exact_issue_comments() {
    for target in [
        "https://api.github.com/repos/org/repo/issues/4/comments?per_page=100&page=2",
        "https://api.github.com/repos/org/other/issues/3/comments?per_page=100&page=2",
        "https://other.example/repos/org/repo/issues/3/comments?per_page=100&page=2",
    ] {
        let link = format!("<{target}>; rel=\"next\"");
        assert_pagination_rejected(
            &[(
                "https://api.github.com/repos/org/repo/issues/3/comments?per_page=100",
                &link,
            )],
            Kind::InvalidInput,
            Some(target),
            1,
        );
    }
}

#[test]
fn ack_and_reconcile_block_cyclic_pagination_before_repeating_get() {
    assert_pagination_rejected(
        &[
            (
                "https://api.github.com/repos/org/repo/issues/3/comments?per_page=100",
                "<https://api.github.com/repos/org/repo/issues/3/comments?per_page=100&page=2>; rel=\"next\"",
            ),
            (
                "https://api.github.com/repos/org/repo/issues/3/comments?per_page=100&page=2",
                "<https://api.github.com/repos/org/repo/issues/3/comments?per_page=100>; rel=\"next\"",
            ),
        ],
        Kind::Blocked,
        None,
        2,
    );
}

#[test]
fn ack_and_reconcile_reject_malformed_next_link() {
    assert_pagination_rejected(
        &[(
            "https://api.github.com/repos/org/repo/issues/3/comments?per_page=100",
            "not-a-link; rel=\"next\"",
        )],
        Kind::Transport,
        None,
        1,
    );
}

#[test]
fn durable_intent_survives_crash_before_any_network_and_restart_delivers_once() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    *remote.expected_intent.borrow_mut() =
        Some(root.path().join(".moai-cli/messages/outbox/r1.json"));
    remote.crash.set(true);
    assert!(
        std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| record(
            root.path(),
            &m,
            &remote
        )))
        .is_err()
    );
    assert_eq!(remote.posts(), 0);
    assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
    assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
    assert_eq!(remote.posts(), 1);
}
#[test]
fn lost_record_response_reconciles_landed_or_blocks_blind_retry_if_absent() {
    for landed in [true, false] {
        let root = temp_root();
        let remote = Remote::new();
        let m = message();
        remote.lost.set(true);
        remote.lands.set(landed);
        if landed {
            assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
            assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
        } else {
            assert_eq!(
                record(root.path(), &m, &remote).unwrap_err().kind,
                Kind::Transport
            );
            assert_eq!(
                record(root.path(), &m, &remote).unwrap_err().kind,
                Kind::Transport
            );
        }
        assert_eq!(entry(root.path(), false)["record"]["attempts"], 1);
        assert_eq!(
            reconcile(root.path(), &m, &remote).unwrap()["state"],
            if landed { "queued" } else { "publish_unknown" }
        );
        assert_eq!(remote.posts(), 1);
    }
}
#[test]
fn independent_receiver_ack_loss_resume_and_later_page_sender_reconciliation() {
    let sender = temp_root();
    let receiver = temp_root();
    let remote = Remote::new();
    let m = message();
    let queued = record(sender.path(), &m, &remote).unwrap();
    assert_eq!(queued["queued"], true);
    assert_eq!(queued["received_ack"], false);
    remote.actor.set(22);
    remote.lost.set(true);
    assert_eq!(
        ack(receiver.path(), &m, &remote).unwrap()["state"],
        "received_ack"
    );
    assert_eq!(
        ack(receiver.path(), &m, &remote).unwrap()["state"],
        "received_ack"
    );
    assert_eq!(entry(receiver.path(), true)["received"], true);
    remote.page_size.set(1);
    let before = remote.posts();
    let result = reconcile(sender.path(), &m, &remote).unwrap();
    assert_eq!(result["received_ack"], true);
    assert_eq!(result["business_acceptance"], false);
    assert_eq!(result["business_completion"], false);
    assert_eq!(before, remote.posts());
    assert!(remote
        .sent
        .borrow()
        .iter()
        .any(|r| r.url.ends_with("&page=2")));
}
#[test]
fn lost_unobserved_ack_never_reposts_even_after_restart() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    record(root.path(), &m, &remote).unwrap();
    remote.actor.set(22);
    remote.lost.set(true);
    remote.lands.set(false);
    assert_eq!(
        ack(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Transport
    );
    remote.lands.set(true);
    assert_eq!(
        ack(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Transport
    );
    assert_eq!(remote.posts(), 2);
    assert_eq!(
        reconcile(root.path(), &m, &remote).unwrap()["state"],
        "publish_unknown"
    );
    assert_eq!(entry(root.path(), true)["ack"]["attempts"], 1);
}
#[test]
fn duplicate_identical_comments_are_one_receipt_but_changed_payload_conflicts() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    record(root.path(), &m, &remote).unwrap();
    let mut duplicate = remote.comments.borrow()[0].clone();
    duplicate["id"] = json!(999);
    remote.comments.borrow_mut().push(duplicate);
    assert_eq!(
        record(root.path(), &m, &remote).unwrap()["record_comment_id"],
        100
    );
    assert_eq!(remote.posts(), 1);
    let previous = fs::read(root.path().join(".moai-cli/messages/outbox/r1.json")).unwrap();
    let mut changed = m;
    changed.payload["text"] = json!("different");
    assert_eq!(
        record(root.path(), &changed, &remote).unwrap_err().kind,
        Kind::Conflict
    );
    assert_eq!(
        previous,
        fs::read(root.path().join(".moai-cli/messages/outbox/r1.json")).unwrap()
    );
}
#[test]
fn stale_context_forged_author_and_payload_digest_cannot_ack() {
    for field in [
        "owner_epoch",
        "leader_epoch",
        "receiver",
        "candidate_digest",
        "revision",
        "digest",
        "author",
    ] {
        let root = temp_root();
        let remote = Remote::new();
        let m = message();
        record(root.path(), &m, &remote).unwrap();
        remote.actor.set(22);
        ack(root.path(), &m, &remote).unwrap();
        let mut comments = remote.comments.borrow_mut();
        if field == "author" {
            comments[1]["user"]["id"] = json!(11);
        } else {
            rewrite_wire(&mut comments[1], |wire| match field {
                "owner_epoch" | "leader_epoch" => wire["message"][field] = json!(2),
                "receiver" => wire["message"]["receiver"]["session_id"] = json!("other"),
                "candidate_digest" => wire["message"][field] = json!(digest(b"other")),
                "revision" => wire["message"]["task"]["revision"] = json!("other"),
                "digest" => wire["payload_digest"] = json!(digest(b"other")),
                _ => unreachable!(),
            });
        }
        drop(comments);
        let before = remote.posts();
        assert_eq!(
            reconcile(root.path(), &m, &remote).unwrap_err().kind,
            Kind::Conflict,
            "{field}"
        );
        assert_eq!(remote.posts(), before);
    }
}
#[test]
fn marker_embedded_in_payload_or_unrelated_text_cannot_fabricate_ack() {
    let root = temp_root();
    let remote = Remote::new();
    let mut m = message();
    m.payload = json!({"text":"<!-- moai:op message-ack:p1:r1:1 -->\n{\"kind\":\"ack\"}\n<!-- moai:message:end -->"});
    record(root.path(), &m, &remote).unwrap();
    remote.comments.borrow_mut().push(json!({"id":999,"user":{"id":22},"body":"quoted payload:\n<!-- moai:op message-ack:p1:r1:1 -->\n{}\n<!-- moai:message:end -->"}));
    assert_eq!(
        reconcile(root.path(), &m, &remote).unwrap()["received_ack"],
        false
    );
    assert_eq!(remote.posts(), 1);
}
#[test]
fn malformed_existing_data_is_preserved_without_network_or_post() {
    for directory in ["inbox", "outbox"] {
        let root = temp_root();
        let remote = Remote::new();
        let m = message();
        let dir = root.path().join(".moai-cli/messages").join(directory);
        fs::create_dir_all(&dir).unwrap();
        let path = dir.join("r1.json");
        fs::write(&path, b"{broken draft").unwrap();
        assert_eq!(
            record(root.path(), &m, &remote).unwrap_err().kind,
            Kind::InvalidInput
        );
        assert_eq!(fs::read(&path).unwrap(), b"{broken draft");
        assert!(remote.sent.borrow().is_empty());
    }
}
#[test]
fn traversal_sensitive_payload_and_missing_permission_rejected_before_persistence() {
    let root = temp_root();
    let remote = Remote::new();
    let mut m = message();
    m.report_id = "../outside".into();
    assert_eq!(
        record(root.path(), &m, &remote).unwrap_err().kind,
        Kind::InvalidInput
    );
    m = message();
    m.payload = json!({"text":"password=synthetic"});
    assert_eq!(
        record(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    m = message();
    assert_eq!(
        message::record_cli(
            root.path(),
            RecordRequest { message: m },
            "",
            &remote,
            &Time
        )
        .unwrap_err()
        .kind,
        Kind::PermissionReferenceMissing
    );
    assert!(!root.path().join(".moai-cli").exists());
    assert!(remote.sent.borrow().is_empty());
}
#[test]
fn advisory_lock_returns_unavailable_without_removing_existing_lock() {
    let root = temp_root();
    let remote = Remote::new();
    let lock = root.path().join(".moai-cli/messages/.lock");
    fs::create_dir_all(lock.parent().unwrap()).unwrap();
    fs::write(&lock, "old holder").unwrap();
    let held = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .open(&lock)
        .unwrap();
    held.try_lock().unwrap();
    assert_eq!(
        record(root.path(), &message(), &remote).unwrap_err().kind,
        Kind::Unavailable
    );
    assert_eq!(fs::read_to_string(lock).unwrap(), "old holder");
    assert!(remote.sent.borrow().is_empty());
}
#[test]
fn publisher_account_must_match_and_receiver_requires_observed_sender() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    remote.actor.set(33);
    assert_eq!(
        record(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(remote.posts(), 0);
    remote.actor.set(22);
    assert_eq!(
        ack(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(remote.posts(), 0);
    remote.actor.set(11);
    record(root.path(), &m, &remote).unwrap();
    assert_eq!(
        ack(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(remote.posts(), 1);
    assert!(!root
        .path()
        .join(".moai-cli/messages/inbox/r1.json")
        .exists());
}
#[test]
fn sorted_recursive_payload_digest_and_reconcile_do_not_imply_business_completion() {
    let sender = temp_root();
    let receiver = temp_root();
    let remote = Remote::new();
    let mut m = message();
    m.payload = serde_json::from_str(r#"{"z":{"b":2,"a":1},"a":[3,2,1]}"#).unwrap();
    let expected = digest(br#"{"a":[3,2,1],"z":{"a":1,"b":2}}"#);
    assert_eq!(
        record(sender.path(), &m, &remote).unwrap()["payload_digest"],
        expected
    );
    m.payload = serde_json::from_str(r#"{"a":[3,2,1],"z":{"a":1,"b":2}}"#).unwrap();
    assert_eq!(
        record(sender.path(), &m, &remote).unwrap()["state"],
        "queued"
    );
    let before = remote.posts();
    let received = reconcile(receiver.path(), &m, &remote).unwrap();
    assert_eq!(received["state"], "queued");
    assert_eq!(received["received"], false);
    assert_eq!(received["received_ack"], false);
    assert_eq!(remote.posts(), before);
}
#[test]
fn unrelated_drafts_hold_and_remote_comments_are_preserved() {
    let root = temp_root();
    let remote = Remote::new();
    let draft = root.path().join(".moai-cli/messages/draft.json");
    fs::create_dir_all(draft.parent().unwrap()).unwrap();
    fs::write(&draft, b"HOLD\nprivate draft").unwrap();
    let hold = root.path().join("HOLD");
    fs::write(&hold, "do not change").unwrap();
    let unrelated =
        json!({"id":8,"user":{"id":11},"body":"HOLD: awaiting approval; unrelated draft"});
    remote.comments.borrow_mut().push(unrelated.clone());
    record(root.path(), &message(), &remote).unwrap();
    reconcile(root.path(), &message(), &remote).unwrap();
    assert_eq!(fs::read(draft).unwrap(), b"HOLD\nprivate draft");
    assert_eq!(fs::read_to_string(hold).unwrap(), "do not change");
    assert_eq!(remote.comments.borrow()[0], unrelated);
    assert!(remote
        .sent
        .borrow()
        .iter()
        .all(|r| r.method == "GET" || (r.method == "POST" && r.url.ends_with("/comments"))));
}
#[cfg(unix)]
#[test]
fn known_symlink_components_and_entry_paths_are_refused_without_target_change() {
    use std::os::unix::fs::symlink;
    for entry_link in [false, true] {
        let root = temp_root();
        let outside = temp_root();
        let remote = Remote::new();
        let target = outside.path().join("keep.json");
        fs::write(&target, b"HOLD draft").unwrap();
        if entry_link {
            let dir = root.path().join(".moai-cli/messages/outbox");
            fs::create_dir_all(&dir).unwrap();
            symlink(&target, dir.join("r1.json")).unwrap();
        } else {
            symlink(outside.path(), root.path().join(".moai-cli")).unwrap();
        }
        assert_eq!(
            record(root.path(), &message(), &remote).unwrap_err().kind,
            Kind::InvalidInput
        );
        assert_eq!(fs::read(&target).unwrap(), b"HOLD draft");
        assert!(remote.sent.borrow().is_empty());
    }
}

#[test]
fn delayed_unknown_publication_is_observed_without_a_second_post() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    remote.lost.set(true);
    remote.lands.set(false);
    assert_eq!(
        record(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Transport
    );
    let body = remote
        .sent
        .borrow()
        .iter()
        .find(|r| r.method == "POST")
        .unwrap()
        .body
        .as_ref()
        .unwrap()["body"]
        .clone();
    remote
        .comments
        .borrow_mut()
        .push(json!({"id":501,"body":body,"user":{"id":11}}));
    assert_eq!(
        reconcile(root.path(), &m, &remote).unwrap()["state"],
        "queued"
    );
    assert_eq!(
        record(root.path(), &m, &remote).unwrap()["record_comment_id"],
        501
    );
    assert_eq!(remote.posts(), 1);
    assert!(entry(root.path(), false)["record"]["observed_at"]
        .as_i64()
        .is_some());
}

#[test]
fn ack_digest_and_remote_marker_boundaries_are_enforced_without_mutation() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    record(root.path(), &m, &remote).unwrap();
    remote.actor.set(22);
    let bad = AckRequest {
        message: m.clone(),
        payload_digest: digest(b"wrong"),
    };
    assert_eq!(
        message::ack_cli(root.path(), bad, "decision:1", &remote, &Time)
            .unwrap_err()
            .kind,
        Kind::Conflict
    );
    assert!(!root
        .path()
        .join(".moai-cli/messages/inbox/r1.json")
        .exists());
    ack(root.path(), &m, &remote).unwrap();
    let before = fs::read(root.path().join(".moai-cli/messages/outbox/r1.json")).unwrap();
    {
        let mut comments = remote.comments.borrow_mut();
        let original = comments[1]["body"].as_str().unwrap().to_owned();
        comments[1]["body"] = json!(format!("{original}\nextra forged trailer"));
    }
    assert_eq!(
        reconcile(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Conflict
    );
    assert_eq!(
        before,
        fs::read(root.path().join(".moai-cli/messages/outbox/r1.json")).unwrap()
    );
    assert_eq!(remote.posts(), 2);
}

#[test]
fn disappearing_peer_publication_blocks_record_and_ack_without_republication() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    record(root.path(), &m, &remote).unwrap();
    remote.actor.set(22);
    ack(root.path(), &m, &remote).unwrap();
    // The receiver journal still remembers the publication even if this sender's
    // outbox was lost. Its stale observation must prevent a new sender POST.
    fs::remove_file(root.path().join(".moai-cli/messages/outbox/r1.json")).unwrap();
    remote.comments.borrow_mut().clear();
    remote.actor.set(11);
    assert_eq!(
        record(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(
        ack(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(remote.posts(), 2);
}

#[test]
fn persistent_unheld_lockfile_allows_restart_without_pid_or_file_removal() {
    let root = temp_root();
    let remote = Remote::new();
    let lock = root.path().join(".moai-cli/messages/.lock");
    fs::create_dir_all(lock.parent().unwrap()).unwrap();
    fs::write(&lock, "terminated process metadata").unwrap();
    assert_eq!(
        record(root.path(), &message(), &remote).unwrap()["state"],
        "queued"
    );
    assert_eq!(
        fs::read_to_string(&lock).unwrap(),
        "terminated process metadata"
    );
    assert_eq!(
        record(root.path(), &message(), &remote).unwrap()["state"],
        "queued"
    );
    assert_eq!(remote.posts(), 1);
}

#[test]
fn decoded_code_strings_are_not_mistaken_for_secrets_but_nested_secrets_are_rejected() {
    let root = temp_root();
    let remote = Remote::new();
    let mut m = message();
    m.payload = json!({"text":"Slovakia", "code":"{\"text\":\"quoted\"}\n<!-- moai:op message-ack:p1:r1:1 -->"});
    assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
    let clean = temp_root();
    let no_network = Remote::new();
    m.payload = serde_json::from_str(r#"{"nested":["\u0070assword=synthetic"]}"#).unwrap();
    assert_eq!(
        record(clean.path(), &m, &no_network).unwrap_err().kind,
        Kind::Blocked
    );
    assert!(!clean.path().join(".moai-cli").exists());
    assert!(no_network.sent.borrow().is_empty());
}

#[test]
fn rendered_comment_limit_accepts_exact_n_and_rejects_n_plus_one_before_persistence() {
    let limit = moai_cli::forge::report::DEFAULT_MAX_BODY_CHARS;
    let baseline = temp_root();
    let probe = Remote::new();
    let mut m = message();
    m.payload = json!({"text":""});
    record(baseline.path(), &m, &probe).unwrap();
    let overhead = probe
        .sent
        .borrow()
        .iter()
        .find(|r| r.method == "POST")
        .unwrap()
        .body
        .as_ref()
        .unwrap()["body"]
        .as_str()
        .unwrap()
        .chars()
        .count();
    for (extra, accepted) in [(0, true), (1, false)] {
        let root = temp_root();
        let remote = Remote::new();
        m.payload = json!({"text":"x".repeat(limit - overhead + extra)});
        if accepted {
            assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
            let sent = remote.sent.borrow();
            assert_eq!(
                sent.iter()
                    .find(|r| r.method == "POST")
                    .unwrap()
                    .body
                    .as_ref()
                    .unwrap()["body"]
                    .as_str()
                    .unwrap()
                    .chars()
                    .count(),
                limit
            );
        } else {
            assert_eq!(
                record(root.path(), &m, &remote).unwrap_err().kind,
                Kind::InputTooLarge
            );
            assert!(!root.path().join(".moai-cli").exists());
            assert!(remote.sent.borrow().is_empty());
        }
    }
}

#[test]
fn reconcile_only_records_observations_without_creating_a_receipt_for_any_actor() {
    let sender = temp_root();
    let remote = Remote::new();
    let m = message();
    record(sender.path(), &m, &remote).unwrap();
    for actor in [11, 22, 33] {
        let observer = temp_root();
        remote.actor.set(actor);
        let result = reconcile(observer.path(), &m, &remote).unwrap();
        assert_eq!(result["state"], "queued");
        assert_eq!(result["received"], false);
        assert_eq!(entry(observer.path(), true)["received"], false);
        assert_eq!(entry(observer.path(), true)["ack"]["attempts"], 0);
    }
    assert_eq!(remote.posts(), 1);
}

#[test]
fn explicit_reinvocation_recovers_known_403_refusal_with_a_new_durable_attempt() {
    let root = temp_root();
    let remote = Remote::new();
    let m = message();
    remote.post_status.set(403);
    assert_eq!(
        record(root.path(), &m, &remote).unwrap_err().kind,
        Kind::Blocked
    );
    assert_eq!(remote.posts(), 1);
    assert_eq!(entry(root.path(), false)["record"]["attempts"], 1);
    assert_eq!(
        reconcile(root.path(), &m, &remote).unwrap()["state"],
        "blocked"
    );
    remote.post_status.set(201);
    assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
    assert_eq!(remote.posts(), 2);
    assert_eq!(entry(root.path(), false)["record"]["attempts"], 2);
    assert_eq!(entry(root.path(), false)["record"]["response_status"], 201);
    assert_eq!(record(root.path(), &m, &remote).unwrap()["state"], "queued");
    assert_eq!(remote.posts(), 2);
}

#[test]
fn timeout_and_server_error_remain_unknown_and_never_allow_a_blind_second_post() {
    for status in [408, 500] {
        let root = temp_root();
        let remote = Remote::new();
        let m = message();
        remote.post_status.set(status);
        assert_eq!(
            record(root.path(), &m, &remote).unwrap_err().kind,
            Kind::Transport
        );
        remote.post_status.set(201);
        assert_eq!(
            record(root.path(), &m, &remote).unwrap_err().kind,
            Kind::Transport
        );
        assert_eq!(
            reconcile(root.path(), &m, &remote).unwrap()["state"],
            "publish_unknown"
        );
        assert_eq!(remote.posts(), 1);
        assert_eq!(entry(root.path(), false)["record"]["attempts"], 1);
    }
}

#[test]
fn explicit_ack_reinvocation_recovers_a_known_refusal_but_not_a_timeout() {
    for status in [403, 408] {
        let root = temp_root();
        let remote = Remote::new();
        let m = message();
        record(root.path(), &m, &remote).unwrap();
        remote.actor.set(22);
        remote.post_status.set(status);
        assert_eq!(
            ack(root.path(), &m, &remote).unwrap_err().kind,
            if status == 403 {
                Kind::Blocked
            } else {
                Kind::Transport
            }
        );
        assert_eq!(entry(root.path(), true)["received"], true);
        remote.post_status.set(201);
        if status == 403 {
            assert_eq!(
                ack(root.path(), &m, &remote).unwrap()["state"],
                "received_ack"
            );
            assert_eq!(entry(root.path(), true)["ack"]["attempts"], 2);
            assert_eq!(remote.posts(), 3);
        } else {
            assert_eq!(
                ack(root.path(), &m, &remote).unwrap_err().kind,
                Kind::Transport
            );
            assert_eq!(entry(root.path(), true)["ack"]["attempts"], 1);
            assert_eq!(remote.posts(), 2);
        }
    }
}

#[test]
fn opaque_task_session_and_revision_ids_survive_real_record_ack_and_reconcile() {
    let sender = temp_root();
    let receiver = temp_root();
    let remote = Remote::new();
    let mut m = message();
    m.task.task_id = "audit:#2414@3661f864e7aa7f50c9d73a6fd2dc94c2a4942971".into();
    m.sender.session_id = "hB:pE".into();
    m.receiver.session_id = format!("hC:pQ:{}", "session".repeat(15));
    m.task.revision = format!(" refs/heads/audit:revision@{} ", "a".repeat(80));
    assert_eq!(
        record(sender.path(), &m, &remote).unwrap()["state"],
        "queued"
    );
    remote.actor.set(22);
    assert_eq!(
        ack(receiver.path(), &m, &remote).unwrap()["state"],
        "received_ack"
    );
    assert_eq!(
        reconcile(sender.path(), &m, &remote).unwrap()["received_ack"],
        true
    );
    for stored in [entry(sender.path(), false), entry(receiver.path(), true)] {
        assert_eq!(stored["message"]["task"]["task_id"], m.task.task_id);
        assert_eq!(stored["message"]["task"]["revision"], m.task.revision);
        assert_eq!(
            stored["message"]["sender"]["session_id"],
            m.sender.session_id
        );
        assert_eq!(
            stored["message"]["receiver"]["session_id"],
            m.receiver.session_id
        );
    }
    for comment in remote.comments.borrow().iter() {
        let wire: Value = serde_json::from_str(
            comment["body"]
                .as_str()
                .unwrap()
                .split('\n')
                .nth(1)
                .unwrap(),
        )
        .unwrap();
        assert_eq!(wire["message"]["task"]["task_id"], m.task.task_id);
        assert_eq!(wire["message"]["task"]["revision"], m.task.revision);
        assert_eq!(wire["message"]["sender"]["session_id"], m.sender.session_id);
        assert_eq!(
            wire["message"]["receiver"]["session_id"],
            m.receiver.session_id
        );
    }
    assert_eq!(remote.posts(), 2);
}

#[test]
fn opaque_identity_validation_rejects_only_empty_or_control_data_before_persistence() {
    for field in ["task", "sender", "receiver", "revision"] {
        for bad in ["", "literal\ncontrol"] {
            let root = temp_root();
            let remote = Remote::new();
            let mut m = message();
            match field {
                "task" => m.task.task_id = bad.into(),
                "sender" => m.sender.session_id = bad.into(),
                "receiver" => m.receiver.session_id = bad.into(),
                "revision" => m.task.revision = bad.into(),
                _ => unreachable!(),
            }
            assert_eq!(
                record(root.path(), &m, &remote).unwrap_err().kind,
                Kind::InvalidInput
            );
            assert!(!root.path().join(".moai-cli").exists());
            assert!(remote.sent.borrow().is_empty());
        }
    }
}

#[test]
fn marker_id_n_and_n_plus_one_preserve_the_shared_marker_contract() {
    for field in ["report_id", "project_id"] {
        for size in [64, 65] {
            let root = temp_root();
            let remote = Remote::new();
            let mut m = message();
            if field == "report_id" {
                m.report_id = "x".repeat(size);
            } else {
                m.project_id = "x".repeat(size);
            }
            let outcome = record(root.path(), &m, &remote);
            if size == 64 {
                assert_eq!(outcome.unwrap()["state"], "queued");
                assert_eq!(remote.posts(), 1);
            } else {
                assert_eq!(outcome.unwrap_err().kind, Kind::InvalidInput);
                assert!(remote.sent.borrow().is_empty());
                assert!(!root.path().join(".moai-cli").exists());
            }
        }
    }
}

#[test]
fn zero_actor_or_ownership_epoch_cannot_persist_or_publish() {
    for field in ["sender", "receiver", "owner_epoch", "leader_epoch"] {
        let root = temp_root();
        let remote = Remote::new();
        let mut m = message();
        match field {
            "sender" => m.sender.actor_id = 0,
            "receiver" => m.receiver.actor_id = 0,
            "owner_epoch" => m.owner_epoch = 0,
            "leader_epoch" => m.leader_epoch = 0,
            _ => unreachable!(),
        }
        assert_eq!(
            record(root.path(), &m, &remote).unwrap_err().kind,
            Kind::InvalidInput
        );
        assert!(remote.sent.borrow().is_empty());
        assert!(!root.path().join(".moai-cli").exists());
    }
}
