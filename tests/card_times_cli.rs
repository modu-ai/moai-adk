//! #3067 S6: `card-times` reads the server times of cards and their pull requests and writes
//! nothing. The collection and the answer are exercised in process through a fake GraphQL
//! transport; the binary runs against a stub `gh` (`MOAI_FORGE_GH_BIN`) in a temporary project and
//! a temporary HOME, so there is no network, no real GitHub and no real HOME.
#![cfg(unix)]
mod common;

use common::CleanEnv;
use moai_cli::card_times::{self, Request, ABSENT, OBSERVED, UNKNOWN};
use moai_cli::forge::policy::{parse_utc, Clock};
use moai_cli::forge::transport::{HttpRequest, HttpResponse, Transport};
use moai_cli::message::Message;
use serde_json::{json, Value};
use std::cell::RefCell;
use std::collections::{BTreeMap, BTreeSet};
use std::fs;
use std::process::{Command, Stdio};

const NOW: &str = "2026-11-16T12:00:00Z";
const OWNER: &str = "acme";
const REPO: &str = "widgets";
/// The card that holds a whole dispatch.
const DISPATCHED: u64 = 101;

struct At(i64);

impl Clock for At {
    fn now(&self) -> i64 {
        self.0
    }
    fn sleep(&self, _secs: u64) {}
}

fn now() -> i64 {
    parse_utc(NOW).unwrap()
}

fn h(c: char) -> String {
    c.to_string().repeat(40)
}

fn conn(nodes: Vec<Value>, total: usize) -> Value {
    json!({"totalCount": total, "nodes": nodes})
}

fn comment(id: u64, at: &str, body: &str) -> Value {
    json!({"databaseId": id, "authorAssociation": "MEMBER", "body": body, "createdAt": at,
           "author": {"databaseId": 11}})
}

fn stranger(id: u64, at: &str, body: &str) -> Value {
    let mut c = comment(id, at, body);
    c["authorAssociation"] = json!("NONE");
    c
}

fn pr(number: u64, branch: &str, merged_at: Option<&str>, comments: Vec<Value>) -> Value {
    let total = comments.len();
    json!({
        "number": number,
        "state": if merged_at.is_some() { "MERGED" } else { "OPEN" },
        "isDraft": false,
        "createdAt": "2026-11-10T08:00:00Z",
        "mergedAt": merged_at,
        "merged": merged_at.is_some(),
        "headRefOid": h('b'),
        "headRefName": branch,
        "body": "",
        "comments": conn(comments, total),
        "timelineItems": conn(vec![], 0),
        "allCommits": conn(vec![json!({"commit": {"oid": h('b'), "committedDate": "2026-11-10T07:50:00Z"}})], 1),
        "headCommit": {"nodes": [{"commit": {"oid": h('b'), "checkSuites": {"totalCount": 1, "nodes": [
            {"createdAt": "2026-11-10T08:05:00Z",
             "checkRuns": {"totalCount": 1, "nodes": [{"startedAt": "2026-11-10T08:06:00Z"}]}}]}}}]},
    })
}

fn issue(
    n: u64,
    state: &str,
    labels: &[&str],
    comments: Vec<Value>,
    total: usize,
    prs: Vec<Value>,
) -> Value {
    let refs = prs.len();
    json!({
        "number": n,
        "state": state,
        "databaseId": 5000 + n,
        "id": format!("I_{n}"),
        "labels": {"nodes": labels.iter().map(|l| json!({"name": l})).collect::<Vec<_>>()},
        "comments": conn(comments, total),
        "timelineItems": conn(prs.into_iter().map(|p| json!({"source": p})).collect(), refs),
    })
}

fn message(
    report: &str,
    event: &str,
    sender: (u64, &str),
    receiver: (u64, &str),
    payload: Value,
) -> Message {
    serde_json::from_value(json!({
        "schema_version": 1, "report_id": report, "project_id": "p1",
        "task": {"owner": OWNER, "repo": REPO, "repository_id": 77, "issue_id": 5000 + DISPATCHED,
                 "issue_node_id": format!("I_{DISPATCHED}"), "issue_number": DISPATCHED,
                 "task_id": "T1", "revision": "1"},
        "sender": {"actor_id": sender.0, "session_id": sender.1},
        "receiver": {"actor_id": receiver.0, "session_id": receiver.1},
        "owner_epoch": 1, "leader_epoch": 1, "event": event,
        "candidate_digest": moai_cli::digest(b"candidate"), "created_at": "2026-11-10T00:00:00Z",
        "approval_refs": ["decision:1"], "hold_refs": [], "payload": payload,
    }))
    .unwrap()
}

/// The comment body `message-record` / `message-ack` publishes for `m`.
fn wire_body(m: &Message, kind: &str) -> String {
    #[derive(serde::Serialize)]
    struct Projection<'a> {
        schema_version: u32,
        kind: &'a str,
        message: &'a Message,
        payload_digest: &'a str,
    }
    let mut m = m.clone();
    m.payload.sort_all_objects();
    let digest = moai_cli::digest(&serde_json::to_vec(&m.payload).unwrap());
    let marker = moai_cli::forge::outbox::op_marker(
        if kind == "ack" {
            "message-ack"
        } else {
            "message-record"
        },
        &m.project_id,
        &m.report_id,
        1,
    );
    let wire = Projection {
        schema_version: 1,
        kind,
        message: &m,
        payload_digest: &digest,
    };
    format!(
        "{marker}\n{}\n<!-- moai:message:end -->",
        serde_json::to_string(&wire).unwrap()
    )
}

fn by(mut c: Value, actor: u64) -> Value {
    c["author"] = json!({"databaseId": actor});
    c
}

/// The four dispatch comments of card 101: sent 06:00, received 06:03, decided 06:04, started 06:10.
fn dispatch_comments() -> Vec<Value> {
    let mut sent = message(
        "d1",
        "dispatch",
        (11, "leader"),
        (22, "lane"),
        json!({"dispatch_id": "job1", "revision": 1, "previous": null, "scope": ["src/x.rs"],
               "completion": ["tests pass"], "audit": null}),
    );
    sent.payload.sort_all_objects();
    let digest = moai_cli::digest(&serde_json::to_vec(&sent.payload).unwrap());
    let reply = |report: &str, event: &str, payload: Value| {
        let mut m = sent.clone();
        m.report_id = report.into();
        m.event = event.into();
        std::mem::swap(&mut m.sender, &mut m.receiver);
        m.payload = payload;
        m
    };
    let decision = reply(
        "d1-decision",
        "dispatch-decision",
        json!({"dispatch_report_id": "d1", "dispatch_payload_digest": digest,
               "disposition": "accepted", "reason": "ok"}),
    );
    let start = reply(
        "d1-start",
        "dispatch-start",
        json!({"dispatch_report_id": "d1", "dispatch_payload_digest": digest, "note": "started"}),
    );
    vec![
        by(
            comment(9001, "2026-11-10T06:00:00Z", &wire_body(&sent, "record")),
            11,
        ),
        by(
            comment(9002, "2026-11-10T06:03:00Z", &wire_body(&sent, "ack")),
            22,
        ),
        by(
            comment(
                9003,
                "2026-11-10T06:04:00Z",
                &wire_body(&decision, "record"),
            ),
            22,
        ),
        by(
            comment(9004, "2026-11-10T06:10:00Z", &wire_body(&start, "record")),
            22,
        ),
    ]
}

#[derive(Default)]
struct World {
    issues: BTreeMap<u64, Value>,
    missing: BTreeSet<u64>,
}

impl World {
    /// What GitHub answers to a query that asks for the Issues named in it.
    fn answer(&self, query: &str) -> Value {
        let asked = |n: u64| query.contains(&format!("i{n}:issue(number:{n})"));
        let mut repo = json!({"databaseId": 77});
        let mut errors = Vec::new();
        for (n, node) in &self.issues {
            if asked(*n) {
                repo[format!("i{n}")] = node.clone();
            }
        }
        for n in self.missing.iter().filter(|n| asked(**n)) {
            repo[format!("i{n}")] = Value::Null;
            errors.push(
                json!({"type": "NOT_FOUND", "path": ["repository", format!("i{n}")],
                "message": format!("Could not resolve to an Issue with the number of {n}.")}),
            );
        }
        let mut body = json!({"data": {"repository": repo}});
        if !errors.is_empty() {
            body["errors"] = json!(errors);
        }
        body
    }
}

struct Fake {
    world: World,
    sent: RefCell<Vec<HttpRequest>>,
}

impl Fake {
    fn new(world: World) -> Fake {
        Fake {
            world,
            sent: RefCell::new(Vec::new()),
        }
    }
}

impl Transport for Fake {
    fn send(&self, req: &HttpRequest) -> Result<HttpResponse, String> {
        self.sent.borrow_mut().push(req.clone());
        let query = req
            .body
            .as_ref()
            .and_then(|b| b["query"].as_str())
            .unwrap_or("");
        Ok(HttpResponse::new(200).with_body(self.world.answer(query)))
    }
}

fn request(issues: &[u64]) -> Request {
    serde_json::from_value(
        json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": issues}),
    )
    .unwrap()
}

/// Six kinds of card: a merged one with a whole dispatch, an unmerged one, one on hold, one not
/// started, one that does not exist and one with more comments than a request reads.
fn world() -> World {
    let mut w = World::default();
    let mut c101 = dispatch_comments();
    c101.push(comment(
        9005,
        "2026-11-10T07:00:00Z",
        "moai-claim: session=agent-1 epoch=1",
    ));
    c101.push(comment(
        9006,
        "2026-11-10T07:30:00Z",
        "pr-wait: 2026-11-10T01:00:00Z waits for the cap",
    ));
    let merged = pr(
        201,
        "issue/101-x",
        Some("2026-11-10T11:00:00Z"),
        vec![
            comment(
                9100,
                "2026-11-10T09:00:00Z",
                &format!("moai-audit-claim: lane=agent-2 head={}", h('b')),
            ),
            comment(
                9101,
                "2026-11-10T10:00:00Z",
                &format!("AUDIT #201 head {} — verdict: PASS", h('b')),
            ),
        ],
    );
    let stray = pr(299, "other-branch", None, vec![]);
    let total = c101.len();
    w.issues.insert(
        101,
        issue(101, "CLOSED", &[], c101, total, vec![merged, stray]),
    );
    let c102 = vec![
        comment(
            9300,
            "2026-11-10T07:00:00Z",
            "moai-claim: session=agent-3 epoch=1",
        ),
        stranger(
            9301,
            "2026-11-10T07:10:00Z",
            "pr-wait: 2026-11-10T00:00:00Z not a member",
        ),
    ];
    let open = pr(
        202,
        "issue/102-y",
        None,
        vec![stranger(
            9400,
            "2026-11-10T09:00:00Z",
            &format!("AUDIT #202 head {} — verdict: PASS", h('b')),
        )],
    );
    w.issues.insert(
        102,
        issue(102, "OPEN", &[], c102.clone(), c102.len(), vec![open]),
    );
    w.issues.insert(
        103,
        issue(103, "OPEN", &["needs:decision"], vec![], 0, vec![]),
    );
    w.issues
        .insert(104, issue(104, "OPEN", &[], vec![], 0, vec![]));
    w.missing.insert(105);
    let many: Vec<Value> = (0..100u64)
        .map(|i| comment(9500 + i, "2026-11-10T07:00:00Z", "chatter"))
        .collect();
    w.issues
        .insert(106, issue(106, "OPEN", &[], many, 150, vec![]));
    w
}

fn all_stages(card: &Value) -> Vec<(String, Value)> {
    let mut out = Vec::new();
    for key in ["sent_at", "received_at", "decided_at", "started_at"] {
        out.push((format!("dispatch.{key}"), card["dispatch"][key].clone()));
    }
    out.push(("claim_at".into(), card["claim_at"].clone()));
    out
}

fn states(v: &Value, keys: &[&str]) -> Vec<String> {
    keys.iter()
        .map(|k| v[*k]["state"].as_str().unwrap().to_string())
        .collect()
}

fn card(out: &Value, n: u64) -> &Value {
    out["cards"]
        .as_array()
        .unwrap()
        .iter()
        .find(|c| c["card"] == json!(n))
        .unwrap_or_else(|| panic!("card {n} missing in {out}"))
}

#[test]
fn card_times_covers_unmerged_held_and_unread_cards() {
    let fake = Fake::new(world());
    let out =
        card_times::run(&request(&[101, 102, 103, 104, 105, 106]), &fake, &At(now())).unwrap();
    assert_eq!(fake.sent.borrow().len(), 3, "two cards to a request");
    assert_eq!(out["api_calls"], 3);

    // The merged card: every stage a server time, spans from the two ends.
    let c = card(&out, 101);
    assert_eq!(c["state"], "closed");
    assert_eq!(c["dispatch"]["state"], OBSERVED, "{}", c["dispatch"]);
    assert_eq!(c["dispatch"]["sent_at"]["at"], "2026-11-10T06:00:00Z");
    assert_eq!(c["dispatch"]["received_at"]["at"], "2026-11-10T06:03:00Z");
    assert_eq!(c["dispatch"]["decided_at"]["at"], "2026-11-10T06:04:00Z");
    assert_eq!(c["dispatch"]["started_at"]["at"], "2026-11-10T06:10:00Z");
    assert_eq!(c["dispatch"]["started_at"]["comment_id"], 9004);
    assert_eq!(c["dispatch"]["spans"]["sent_to_received"]["secs"], 180);
    assert_eq!(c["dispatch"]["spans"]["sent_to_started"]["secs"], 600);
    assert_eq!(c["dispatch"]["disposition"], "accepted");
    assert_eq!(c["claim_at"]["comment_id"], 9005);
    assert_eq!(c["pr_wait_state"], OBSERVED);
    assert_eq!(c["pr_wait"][0]["at"], "2026-11-10T07:30:00Z");
    assert_eq!(c["pr_wait"][0]["stated_at"], "2026-11-10T01:00:00Z");
    let prs = c["prs"].as_array().unwrap();
    assert_eq!(
        prs.len(),
        1,
        "the stray cross-reference is not the card's PR: {prs:?}"
    );
    let p = &prs[0];
    assert_eq!(p["number"], 201);
    assert_eq!(
        states(
            p,
            &[
                "opened",
                "ready",
                "ci_queued",
                "ci_running",
                "audit_reserved",
                "verdict",
                "merged"
            ]
        ),
        vec![OBSERVED; 7]
    );
    assert_eq!(p["verdict"]["result"], "PASS");
    assert_eq!(p["spans"]["verdict_to_merged"]["secs"], 3600);
    assert_eq!(p["spans"]["opened_to_merged"]["secs"], 3 * 3600);

    // The unmerged card: a claim and an open PR; merged, verdict and reservation are absent
    // (the stranger's AUDIT line does not count), and the stranger is listed.
    let c = card(&out, 102);
    assert_eq!(c["claim_at"]["state"], OBSERVED);
    assert_eq!(
        c["pr_wait_state"], ABSENT,
        "a non-member pr-wait: line is not honoured"
    );
    let p = &c["prs"][0];
    assert_eq!(
        states(p, &["merged", "verdict", "audit_reserved"]),
        vec![ABSENT; 3]
    );
    assert_eq!(p["state"], "open");
    assert_eq!(c["dispatch"]["state"], ABSENT);

    // Held, and a card never started: absent, not unknown.
    let c = card(&out, 103);
    assert_eq!(
        (c["held"].clone(), c["held_by"].clone()),
        (json!(true), json!("needs:decision"))
    );
    for n in [103, 104] {
        let c = card(&out, n);
        assert_eq!(c["state"], "open");
        assert_eq!(c["claim_at"]["state"], ABSENT, "{c}");
        assert_eq!(c["dispatch"]["state"], ABSENT);
        assert_eq!(c["prs"], json!([]));
        assert_eq!(c["prs_complete"], true);
    }
    assert_eq!(card(&out, 104)["held"], false);

    // An Issue that does not exist: its own state, nothing guessed.
    let c = card(&out, 105);
    assert_eq!(c["state"], "not_found");
    for (name, stage) in all_stages(c) {
        assert_eq!(stage["state"], UNKNOWN, "{name}");
    }

    // More comments than one request reads: the comment-derived stages are unknown, never absent.
    let c = card(&out, 106);
    assert_eq!(c["comments_complete"], false);
    assert_eq!(c["claim_at"]["state"], UNKNOWN);
    assert_eq!(c["dispatch"]["state"], UNKNOWN);
    assert_eq!(c["pr_wait_state"], UNKNOWN);
    assert!(
        out["partial"]
            .as_array()
            .unwrap()
            .iter()
            .any(|p| p.as_str().unwrap().contains("#106")),
        "{}",
        out["partial"]
    );

    // The two non-member lines are reported, not honoured.
    let ignored = &out["ignored_control_lines"];
    assert_eq!(ignored["count"], 2, "{ignored}");
    let kinds: BTreeSet<&str> = ignored["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| i["kind"].as_str().unwrap())
        .collect();
    assert_eq!(kinds, BTreeSet::from(["pr_wait", "audit_verdict"]));
    assert_eq!(out["writes"], false);
    assert_eq!(out["claim_authorized"], false);
    assert_eq!(out["ownership_transferred"], false);
    assert_eq!(out["audit_gate_integrated"], false);
}

#[test]
fn only_graphql_queries_are_sent() {
    let fake = Fake::new(world());
    card_times::run(&request(&[101, 102, 103, 104, 105, 106]), &fake, &At(now())).unwrap();
    for req in fake.sent.borrow().iter() {
        assert_eq!(
            (req.method.as_str(), req.url.as_str()),
            ("POST", "https://api.github.com/graphql")
        );
        let query = req.body.as_ref().unwrap()["query"].as_str().unwrap();
        assert!(query.trim_start().starts_with("query"), "{query}");
        assert!(!query.to_lowercase().contains("mutation"), "{query}");
    }
}

#[test]
fn the_request_budget_turns_the_cards_it_cannot_read_unknown() {
    let issues = [101, 102, 103, 104, 105, 106];
    for (budget, read) in [(3u32, 6usize), (2, 4), (1, 2), (0, 0)] {
        let fake = Fake::new(world());
        let got = card_times::collect(&fake, &request(&issues), now(), budget);
        assert_eq!(got.api_calls, budget.min(3), "budget {budget}");
        assert_eq!(fake.sent.borrow().len(), budget.min(3) as usize);
        let unread: Vec<&Value> = got.cards.iter().skip(read).collect();
        assert_eq!(got.cards.len(), 6);
        for c in &unread {
            assert_eq!(c["state"], UNKNOWN, "budget {budget}: {c}");
            assert_eq!(c["claim_at"]["state"], UNKNOWN);
            assert_eq!(c["dispatch"]["state"], UNKNOWN);
            assert_eq!(c["prs_complete"], false);
        }
        assert_eq!(
            got.partial
                .iter()
                .filter(|p| p.contains("request budget"))
                .count(),
            (6 - read) / 2,
            "budget {budget}: {:?}",
            got.partial
        );
    }
    assert_eq!(card_times::API_BUDGET, 30, "the budget lane-next spends");
}

#[test]
fn a_time_in_the_future_or_a_failed_request_is_unknown_never_a_value() {
    let mut w = World::default();
    let future = pr(301, "issue/301-z", Some("2099-01-01T00:00:00Z"), vec![]);
    w.issues
        .insert(301, issue(301, "OPEN", &[], vec![], 0, vec![future]));
    let out = card_times::run(&request(&[301]), &Fake::new(w), &At(now())).unwrap();
    let p = &card(&out, 301)["prs"][0];
    assert_eq!(p["merged"]["state"], UNKNOWN, "{p}");
    assert_eq!(p["merged"]["at_epoch"], Value::Null);
    assert_eq!(p["spans"]["opened_to_merged"]["state"], UNKNOWN);
    assert!(p["spans"]["opened_to_merged"].get("secs").is_none());

    struct Down;
    impl Transport for Down {
        fn send(&self, _req: &HttpRequest) -> Result<HttpResponse, String> {
            Ok(HttpResponse::new(502))
        }
    }
    let out = card_times::run(&request(&[301, 302]), &Down, &At(now())).unwrap();
    for c in out["cards"].as_array().unwrap() {
        assert_eq!(c["state"], UNKNOWN);
        assert_eq!(c["claim_at"]["state"], UNKNOWN);
    }
    assert!(
        out["partial"][0].as_str().unwrap().contains("HTTP 502"),
        "{}",
        out["partial"]
    );
}

#[test]
fn the_card_count_has_both_edges() {
    let ok = |n: usize| request(&(1..=n as u64).collect::<Vec<_>>()).validate();
    assert!(ok(1).is_ok());
    assert!(ok(50).is_ok());
    assert!(ok(51).unwrap_err().contains("1 to 50"));
    let none: Request = serde_json::from_value(
        json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": []}),
    )
    .unwrap();
    assert!(none.validate().unwrap_err().contains("1 to 50"));
    assert!(request(&[5, 5]).validate().unwrap_err().contains("twice"));
    assert!(request(&[0]).validate().is_err());
    let bad_version: Request = serde_json::from_value(
        json!({"schema_version": 2, "owner": OWNER, "repo": REPO, "issues": [1]}),
    )
    .unwrap();
    assert!(bad_version.validate().is_err());
    let bad_owner: Request = serde_json::from_value(
        json!({"schema_version": 1, "owner": "a/b", "repo": REPO, "issues": [1]}),
    )
    .unwrap();
    assert!(bad_owner.validate().is_err());
    assert!(serde_json::from_value::<Request>(
        json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": [1], "extra": 1})
    )
    .is_err());
}

// --- the binary ---------------------------------------------------------------------------

const EXE: &str = common::MOAI_BIN;

struct Bin {
    _dir: tempfile::TempDir,
    project: std::path::PathBuf,
    home: std::path::PathBuf,
    stub: std::path::PathBuf,
    calls: std::path::PathBuf,
}

impl Bin {
    /// A project with a `.git` folder, an empty HOME, and a `gh` stub that answers the Nth call
    /// with `resp-N.json` (HTTP 502 when there is none) and keeps what it was sent.
    fn new(answers: &[Value]) -> Bin {
        let dir = tempfile::TempDir::new().unwrap();
        let base = fs::canonicalize(dir.path()).unwrap();
        let (project, home, calls) = (base.join("project"), base.join("home"), base.join("calls"));
        for d in [&project.join(".git"), &home, &calls] {
            fs::create_dir_all(d).unwrap();
        }
        fs::write(project.join(".git/HEAD"), "ref: refs/heads/main\n").unwrap();
        fs::write(project.join("README.md"), "project\n").unwrap();
        for (i, a) in answers.iter().enumerate() {
            fs::write(calls.join(format!("resp-{}.json", i + 1)), a.to_string()).unwrap();
        }
        let stub = base.join("gh-stub");
        common::write_executable(
            &stub,
            "#!/bin/sh\nn=1\nwhile [ -e \"$GH_DIR/req-$n.json\" ]; do n=$((n+1)); done\n/bin/cat > \"$GH_DIR/req-$n.json\"\nprintf '%s\\n' \"$*\" >> \"$GH_DIR/args.log\"\nif [ -e \"$GH_DIR/resp-$n.json\" ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n'; /bin/cat \"$GH_DIR/resp-$n.json\"; else printf 'HTTP/2.0 502 Bad Gateway\\r\\n\\r\\n{}'; fi\n",
        );
        Bin {
            _dir: dir,
            project,
            home,
            stub,
            calls,
        }
    }

    fn run(&self, stdin: &Value) -> Value {
        let mut child = Command::new(EXE)
            .arg("card-times")
            .current_dir(&self.project)
            .clean_env()
            .env("HOME", &self.home)
            .env("MOAI_FORGE_GH_BIN", &self.stub)
            .env("GH_DIR", &self.calls)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        common::feed_stdin(&mut child, stdin.to_string().as_bytes());
        let out = child.wait_with_output().unwrap();
        serde_json::from_slice(&out.stdout)
            .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
    }

    fn sent(&self) -> Vec<String> {
        let mut out = Vec::new();
        let mut n = 1;
        while let Ok(text) = fs::read_to_string(self.calls.join(format!("req-{n}.json"))) {
            out.push(text);
            n += 1;
        }
        out
    }
}

fn answers() -> Vec<Value> {
    let w = world();
    [[101u64, 102], [103, 104], [105, 106]]
        .iter()
        .map(|pair| {
            let query = pair
                .iter()
                .map(|n| format!("i{n}:issue(number:{n})"))
                .collect::<Vec<_>>()
                .join(" ");
            w.answer(&query)
        })
        .collect()
}

type TreeState = Vec<(String, Vec<u8>)>;

fn trees(b: &Bin) -> (TreeState, TreeState) {
    (
        common::cli::tree_state(&b.project),
        common::cli::tree_state(&b.home),
    )
}

#[test]
fn card_times_writes_nothing() {
    let b = Bin::new(&answers());
    let before = trees(&b);
    let v = b.run(&json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": [101, 102, 103, 104, 105, 106]}));
    assert_eq!(v["operation"], "card.times", "{v}");
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(
        card(&v["result"], 101)["dispatch"]["started_at"]["at"],
        "2026-11-10T06:10:00Z"
    );
    assert_eq!(card(&v["result"], 105)["state"], "not_found");
    assert_eq!(
        trees(&b),
        before,
        "no file in the project (.git included) or the home changed"
    );
    assert!(!b.project.join(".moai-cli").exists() && !b.home.join(".moai-cli").exists());
    let sent = b.sent();
    assert_eq!(sent.len(), 3, "{sent:?}");
    for body in &sent {
        let req: Value = serde_json::from_str(body).unwrap();
        let query = req["query"].as_str().unwrap();
        assert!(
            query.trim_start().starts_with("query") && !query.to_lowercase().contains("mutation"),
            "{query}"
        );
    }
    let args = fs::read_to_string(b.calls.join("args.log")).unwrap();
    for line in args.lines() {
        assert!(line.contains("graphql"), "{line}");
        for verb in ["PUT", "PATCH", "DELETE"] {
            assert!(!line.contains(verb), "{line}");
        }
    }

    // The same when the read fails: HTTP 502 for every call, still nothing written.
    let b = Bin::new(&[]);
    let before = trees(&b);
    let v =
        b.run(&json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": [101, 102]}));
    assert_eq!(v["ok"], true, "{v}");
    assert_eq!(card(&v["result"], 101)["state"], UNKNOWN);
    assert_eq!(trees(&b), before);
    assert!(!b.project.join(".moai-cli").exists() && !b.home.join(".moai-cli").exists());

    // And when the input is refused before any read.
    for bad in [
        json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": []}),
        json!({"schema_version": 1, "owner": OWNER, "repo": REPO, "issues": (1..=51).collect::<Vec<u64>>()}),
    ] {
        let b = Bin::new(&[]);
        let before = trees(&b);
        let v = b.run(&bad);
        assert_eq!(v["ok"], false, "{v}");
        assert!(b.sent().is_empty(), "a refused input reads nothing");
        assert_eq!(trees(&b), before);
    }
}

#[test]
fn the_binary_lists_card_times_in_its_help() {
    let out = Command::new(EXE)
        .args(["help", "--all"])
        .clean_env()
        .stdout(Stdio::piped())
        .output()
        .unwrap();
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("card-times"), "{text}");
}
