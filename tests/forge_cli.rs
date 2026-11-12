//! 07-A (#2008) GitHub adapter. Every library test replays recorded, synthetic exchanges
//! through `RecordedTransport` with an injected clock: no network, no sleeping. The CLI tests
//! point `forge-read` at a stub `gh` written into a temporary directory.
use moai_cli::forge::{
    self,
    policy::Clock,
    transport::{Exchange, HttpRequest, RecordedTransport},
    write::{LastRead, WriteOutcome, Writer},
    Client, Config, IssueKey, ReadRequest,
};
mod common;

use serde_json::{json, Value};
use std::{
    cell::{Cell, RefCell},
    fs,
    process::{Command, Stdio},
};

struct FakeClock {
    now: Cell<i64>,
    slept: RefCell<Vec<u64>>,
}
impl FakeClock {
    fn at(now: i64) -> Self {
        FakeClock {
            now: Cell::new(now),
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

fn fixture(name: &str) -> Value {
    let path = format!("{}/tests/fixtures/forge/{name}", env!("CARGO_MANIFEST_DIR"));
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}
fn exchanges(v: &Value) -> Vec<Exchange> {
    serde_json::from_value(v.clone()).unwrap()
}
fn request(state: &str) -> ReadRequest {
    serde_json::from_value(json!({"schema_version": 1, "owner": "acme", "repo": "widgets", "state": state, "include_comments": false})).unwrap()
}

#[test]
fn forge_ids_are_stable_and_repo_scoped() {
    let issue = |owner: &str, repo: &str, repo_id: i64, id: i64, number: i64| {
        json!({"ref": {"provider": "github", "host": "github.com",
            "project": {"owner": owner, "repo": repo, "repository_id": repo_id},
            "issue": {"id": id, "node_id": format!("I_{id}"), "number": number}}})
    };
    let before = issue("acme", "widgets", 101, 7001, 7);
    let renamed = issue("acme-labs", "widgets-core", 101, 7001, 7);
    let other_repo_same_number = issue("acme", "gadgets", 202, 9007, 7);
    assert_eq!(forge::identity(&before), forge::identity(&renamed));
    assert_ne!(
        forge::identity(&before),
        forge::identity(&other_repo_same_number)
    );
    assert!(
        forge::identity(&json!({"ref": {"issue": {"number": 7}}})).is_none(),
        "number alone is not a key"
    );

    // Items produced by a real read carry the full reference.
    let t = RecordedTransport::new(exchanges(&fixture("conditional_304.json")["phases"][0]));
    let (result, _) =
        forge::read(&request("open"), None, &t, &FakeClock::at(1_794_000_000)).unwrap();
    let first = &result["items"][0]["ref"];
    assert_eq!(
        first["project"],
        json!({"owner": "acme", "repo": "widgets", "repository_id": 101})
    );
    assert_eq!(
        first["issue"],
        json!({"id": 7001, "node_id": "I_7001", "number": 7})
    );
    assert_eq!(first["etag"], json!("W/\"issues-1\""));
}

#[test]
fn forge_conditional_get_uses_etag_and_304() {
    let phases = fixture("conditional_304.json")["phases"].clone();
    let clock = FakeClock::at(1_794_000_000);
    let mut prior = None;
    let mut results = Vec::new();
    for phase in phases.as_array().unwrap() {
        // Each phase's recorded requests declare the If-None-Match header that must be sent.
        let t = RecordedTransport::new(exchanges(phase));
        let (result, state) = forge::read(&request("open"), prior.take(), &t, &clock).unwrap();
        assert_eq!(t.remaining(), 0);
        let urls: Vec<String> = t.sent.borrow().iter().map(|r| r.url.clone()).collect();
        let mut unique = urls.clone();
        unique.dedup();
        assert_eq!(urls, unique, "each URL is requested once per poll");
        assert!(t.sent.borrow().iter().all(|r| r.method == "GET"));
        results.push(result);
        prior = Some(state);
    }
    assert_eq!(results[0]["cache_hit"], json!(false));
    assert_eq!(results[1]["cache_hit"], json!(false));
    assert_eq!(results[2]["cache_hit"], json!(true));
    assert_eq!(results[2]["items"], results[1]["items"]);
    assert_eq!(results[2]["requests"].as_array().unwrap().len(), 2);
    assert!(results[2]["requests"]
        .as_array()
        .unwrap()
        .iter()
        .all(|r| r["status"] == 304));
    // The pull request is kept apart from Issue TODOs (REQ-FORGE-006).
    let kinds: Vec<&str> = results[0]["items"]
        .as_array()
        .unwrap()
        .iter()
        .map(|i| i["kind"].as_str().unwrap())
        .collect();
    assert_eq!(kinds, ["issue", "pull_request"]);
    // The cursor is the newest updated_at minus the 60 s overlap, not the wall clock.
    assert_eq!(
        prior.as_ref().unwrap().issues_since.as_deref(),
        Some("2026-10-31T23:59:00Z")
    );
    assert_eq!(results[0]["rate_limit"]["remaining"], json!(4998));
    assert_eq!(
        results[0]["rate_limit"]["reset"],
        json!("2026-11-08T10:44:28Z")
    );
}

#[test]
fn forge_backoff_follows_github_headers() {
    let script = exchanges(&fixture("ratelimit_sequence.json")["exchanges"]);
    let req = HttpRequest::new("GET", "https://api.github.com/repos/acme/widgets");

    let t = RecordedTransport::new(script.clone());
    let clock = FakeClock::at(1_794_000_000);
    let client = Client::new(&t, &clock, Config::default(), Default::default());
    let resp = client.send_with_backoff(&req).unwrap();
    assert_eq!(resp.status, 200);
    // retry-after 3 s; then until x-ratelimit-reset (1794000100, reached exactly: 3 + 97);
    // then a secondary limit on the third attempt waits 60 s << 2.
    assert_eq!(*clock.slept.borrow(), vec![3, 97, 240]);
    assert_eq!(1_794_000_000 + 3 + 97, 1_794_000_100);
    assert!(clock.slept.borrow()[2] >= 60);

    let t = RecordedTransport::new(script);
    let clock = FakeClock::at(1_794_000_000);
    let config = Config {
        max_retries: 2,
        ..Config::default()
    };
    let client = Client::new(&t, &clock, config, Default::default());
    let err = client.send_with_backoff(&req).unwrap_err();
    assert_eq!((err.kind, err.retryable), ("blocked", true));
    assert_eq!(
        t.sent.borrow().len(),
        3,
        "requests stop after the retry budget"
    );
}

#[test]
fn forge_follows_link_header_only() {
    let f = fixture("link_three_pages.json");
    let script = exchanges(&f["exchanges"]);
    let expected: Vec<String> = script.iter().map(|e| e.request.url.clone()).collect();
    let t = RecordedTransport::new(script);
    let (result, state) =
        forge::read(&request("all"), None, &t, &FakeClock::at(1_794_000_000)).unwrap();
    let sent: Vec<String> = t.sent.borrow().iter().map(|r| r.url.clone()).collect();
    assert_eq!(
        sent, expected,
        "every page URL is the Link value, byte for byte"
    );
    assert_eq!(result["items"].as_array().unwrap().len(), 3);
    assert_eq!(state.issues_since.as_deref(), Some("2026-11-02T23:59:00Z"));
}

#[test]
fn forge_write_helpers_never_patch_sets() {
    let f = fixture("labels_conflict.json");
    let key = IssueKey {
        owner: "acme".into(),
        repo: "widgets".into(),
        repository_id: 101,
        id: 7001,
        number: 7,
    };
    let last = LastRead {
        updated_at: "2026-11-01T00:00:00Z".into(),
        body_digest: moai_cli::digest(b"spec"),
    };

    // A person edited the issue after our read: stop, send nothing but the re-read.
    let t = RecordedTransport::new(exchanges(&f["conflict"]));
    let clock = FakeClock::at(1_794_000_000);
    let client = Client::new(&t, &clock, Config::default(), Default::default());
    let outcome = Writer::new(&client)
        .add_labels(&key, &last, &["host:codex"])
        .unwrap();
    assert_eq!(
        outcome,
        WriteOutcome::Conflict {
            updated_at: "2026-11-01T10:00:00Z".into()
        }
    );
    assert_eq!(
        t.sent
            .borrow()
            .iter()
            .map(|r| r.method.as_str())
            .collect::<Vec<_>>(),
        ["GET"]
    );

    // Unchanged issue: add, remove and assign through the dedicated endpoints only.
    let t = RecordedTransport::new(exchanges(&f["applied"]));
    let clock = FakeClock::at(1_794_000_000);
    let client = Client::new(&t, &clock, Config::default(), Default::default());
    let writer = Writer::new(&client);
    assert_eq!(
        writer.add_labels(&key, &last, &["host:codex"]).unwrap(),
        WriteOutcome::Applied
    );
    assert_eq!(
        writer.remove_label(&key, &last, "priority:P1 old").unwrap(),
        WriteOutcome::Applied
    );
    assert_eq!(
        writer.add_assignees(&key, &last, &["octocat"]).unwrap(),
        WriteOutcome::Applied
    );
    assert_eq!(t.remaining(), 0);
    let sent = t.sent.borrow();
    assert!(sent
        .iter()
        .all(|r| r.method != "PATCH" && r.method != "PUT"));
    assert_eq!(sent[1].body, Some(json!({"labels": ["host:codex"]})));
    assert_eq!(sent[5].body, Some(json!({"assignees": ["octocat"]})));
    // Writes are at least one second apart.
    assert_eq!(*clock.slept.borrow(), vec![1, 1]);
    assert!(writer.add_assignees(&key, &last, &["a"; 11]).is_err());
}

// ---- CLI: stub gh, poll-state location, token never surfaced ----

const SENTINEL: &str = "ghp_moaiTestSentinelNotAToken000";

fn stub_gh(dir: &std::path::Path, mode: &str) -> std::path::PathBuf {
    let repo = "HTTP/2.0 200 OK\r\nEtag: \"repo-1\"\r\nX-Ratelimit-Limit: 5000\r\nX-Ratelimit-Remaining: 4999\r\nX-Ratelimit-Reset: 1794134668\r\nX-Ratelimit-Resource: core\r\n\r\n{\"id\":101,\"full_name\":\"acme/widgets\"}";
    let issues = "HTTP/2.0 200 OK\r\nEtag: W/\"issues-1\"\r\n\r\n[{\"id\":7001,\"node_id\":\"I_7001\",\"number\":7,\"title\":\"t\",\"state\":\"open\",\"state_reason\":null,\"labels\":[],\"updated_at\":\"2026-11-01T00:00:00Z\"}]";
    let comments = "HTTP/2.0 200 OK\r\nEtag: W/\"c-1\"\r\n\r\n[{\"id\":5001,\"node_id\":\"IC_5001\",\"issue_url\":\"https://api.github.com/repos/acme/widgets/issues/7\",\"updated_at\":\"2026-11-01T01:00:00Z\",\"body\":\"secret-ish comment\",\"user\":{\"login\":\"someone\"}}]";
    let unauthorized = "HTTP/2.0 401 Unauthorized\r\n\r\n{\"message\":\"Bad credentials\"}";
    let script = format!(
        "#!/bin/sh\n# test stub for gh api\nfor a in \"$@\"; do last=\"$a\"; done\nif [ \"{mode}\" = 401 ]; then printf '%s' '{unauthorized}'; echo \"gh: HTTP 401 $GH_TOKEN\" >&2; exit 1; fi\ncase \"$last\" in\n  *issues/comments*) printf '%s' '{comments}';;\n  *issues*) printf '%s' '{issues}';;\n  *) printf '%s' '{repo}';;\nesac\n"
    );
    let path = dir.join("gh-stub");
    common::write_executable(&path, &script);
    path
}

fn forge_read(
    project: &std::path::Path,
    gh: &std::path::Path,
    input: &str,
) -> std::process::Output {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("forge-read")
        .arg(project)
        .env("MOAI_FORGE_GH_BIN", gh)
        .env("GH_TOKEN", SENTINEL)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    // The child may exit before reading stdin; only a broken pipe is tolerated (#2024).
    let mut stdin = child.stdin.take().unwrap();
    match std::io::Write::write_all(&mut stdin, input.as_bytes()) {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    drop(stdin);
    child.wait_with_output().unwrap()
}

#[cfg(unix)]
#[test]
fn forge_read_cli_writes_only_its_poll_state_and_never_the_token() {
    let tmp = tempfile::tempdir().unwrap();
    let project = tmp.path().join("project");
    fs::create_dir(&project).unwrap();
    let gh = stub_gh(tmp.path(), "ok");
    let input = r#"{"schema_version":1,"owner":"acme","repo":"widgets"}"#;

    let out = forge_read(&project, &gh, input);
    let stdout = String::from_utf8_lossy(&out.stdout);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(out.status.success(), "{stderr}");
    let v: Value = serde_json::from_str(&stdout).unwrap();
    assert_eq!(
        (
            v["protocol_version"].clone(),
            v["operation"].clone(),
            v["ok"].clone()
        ),
        (json!(1), json!("forge.read"), json!(true))
    );
    let result = &v["result"];
    for key in [
        "schema_version",
        "observed_at",
        "rate_limit",
        "last_success_at",
        "items",
    ] {
        assert!(result.get(key).is_some(), "{key}");
    }
    assert_eq!(
        result["comments"][0],
        json!({"kind": "comment", "id": 5001, "node_id": "IC_5001", "issue_number": 7, "updated_at": "2026-11-01T01:00:00Z"})
    );
    assert!(
        !stdout.contains("secret-ish") && !stdout.contains("someone"),
        "comment bodies and logins are not kept"
    );

    let state = project.join(".moai-cli/forge/poll-state.json");
    let state_text = fs::read_to_string(&state).unwrap();
    assert!(
        !project.join(".moai").exists(),
        "nothing under the Go .moai/"
    );
    let mut entries: Vec<String> = walk(&project);
    entries.sort();
    assert_eq!(
        entries,
        [
            ".moai-cli",
            ".moai-cli/forge",
            ".moai-cli/forge/poll-state.json"
        ]
    );
    for text in [&*stdout, &*stderr, &state_text] {
        assert!(!text.contains(SENTINEL), "token value surfaced");
    }

    // Unauthorized: blocked envelope, exit 1, the token is still nowhere, state untouched.
    let before = fs::read(&state).unwrap();
    let gh401 = stub_gh(&tmp.path().join(""), "401");
    let out = forge_read(&project, &gh401, input);
    assert!(!out.status.success());
    let stdout = String::from_utf8_lossy(&out.stdout);
    let v: Value = serde_json::from_str(&stdout).unwrap();
    assert_eq!(v["ok"], json!(false));
    assert_eq!(v["error"]["kind"], json!("blocked"));
    assert!(!stdout.contains(SENTINEL) && !String::from_utf8_lossy(&out.stderr).contains(SENTINEL));
    assert_eq!(fs::read(&state).unwrap(), before);
}

fn walk(root: &std::path::Path) -> Vec<String> {
    let mut out = Vec::new();
    let mut stack = vec![root.to_path_buf()];
    while let Some(dir) = stack.pop() {
        for e in fs::read_dir(&dir).unwrap() {
            let p = e.unwrap().path();
            out.push(p.strip_prefix(root).unwrap().to_string_lossy().into_owned());
            if p.is_dir() {
                stack.push(p);
            }
        }
    }
    out
}

#[test]
fn forge_read_rejects_bad_input_without_requests() {
    let tmp = tempfile::tempdir().unwrap();
    // A real stub so the gh seam is accepted; rejected input never reaches it.
    let gh = stub_gh(tmp.path(), "ok");
    for (input, kind) in [
        (
            r#"{"schema_version":1,"owner":"acme","repo":"widgets","token":"x"}"#,
            "invalid_input",
        ),
        (
            r#"{"schema_version":2,"owner":"acme","repo":"widgets"}"#,
            "unsupported_protocol_version",
        ),
        (
            r#"{"schema_version":1,"owner":"../etc","repo":"widgets"}"#,
            "invalid_input",
        ),
        (
            r#"{"schema_version":1,"owner":"acme","repo":"widgets","state":"merged"}"#,
            "invalid_input",
        ),
    ] {
        let out = forge_read(tmp.path(), &gh, input);
        assert!(!out.status.success(), "{input}");
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["operation"], json!("forge.read"));
        assert_eq!(v["error"]["kind"], json!(kind), "{input}");
        // stderr carries the same message as one human line.
        let err = String::from_utf8_lossy(&out.stderr);
        assert_eq!(err.lines().count(), 1, "{err}");
        assert_eq!(err.trim_end(), v["error"]["message"].as_str().unwrap());
    }
    // The test seam is accepted only as an absolute path to an existing file.
    for bad in ["gh", "/no/such/gh"] {
        let out = forge_read(
            tmp.path(),
            std::path::Path::new(bad),
            r#"{"schema_version":1,"owner":"acme","repo":"widgets"}"#,
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v["error"]["kind"], json!("invalid_input"), "{bad}");
    }
    let big = format!(
        r#"{{"schema_version":1,"owner":"acme","repo":"{}"}}"#,
        "x".repeat(70_000)
    );
    let out = forge_read(tmp.path(), &gh, &big);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["error"]["kind"], json!("input_too_large"));
    assert!(
        !tmp.path().join(".moai-cli").exists(),
        "no state for rejected input"
    );
}

// ---- #2048: a stalled gh must not block the caller ----

/// A `gh` that prints nothing and never exits on its own. The `sleep` runs as a grandchild of
/// the shell, so it keeps the output pipes open even if only the shell is ended.
#[cfg(unix)]
fn stalled_gh(dir: &std::path::Path, name: &str) -> std::path::PathBuf {
    // One script and pid file per call. The ETXTBSY fix itself (#2283) is in
    // `common::write_executable`: the script is written by a child process, so no
    // concurrently forked test child can inherit a write fd on it.
    let path = dir.join(name);
    let pid_file = dir.join(format!("{name}.pid"));
    let script = format!(
        "#!/bin/sh\n# test stub: gh that hangs\nsleep 120 &\necho $! > '{}'\nwait\n",
        pid_file.display()
    );
    common::write_executable(&path, &script);
    path
}

/// True once `pid` no longer runs, waiting at most `limit`. A killed process orphaned to init
/// stays visible to `kill -0` as a zombie until init reaps it (#2283: Linux CI showed state
/// `Z`, parent 1, right after the group kill), so a zombie counts as ended.
#[cfg(unix)]
fn process_ends_within(pid: &str, limit: std::time::Duration) -> bool {
    let started = std::time::Instant::now();
    loop {
        let signalable = Command::new("kill")
            .args(["-0", pid])
            .stderr(Stdio::null())
            .status()
            .unwrap()
            .success();
        let zombie = fs::read_to_string(format!("/proc/{pid}/stat"))
            .ok()
            .and_then(|s| s.rsplit_once(") ").map(|(_, rest)| rest.starts_with('Z')))
            .unwrap_or(false);
        if !signalable || zombie {
            return true;
        }
        if started.elapsed() >= limit {
            return false;
        }
        std::thread::sleep(std::time::Duration::from_millis(20));
    }
}

#[cfg(unix)]
#[test]
fn transport_times_out_on_stalled_gh() {
    use moai_cli::forge::transport::{GhTransport, Transport};
    use std::time::{Duration, Instant};
    let tmp = tempfile::tempdir().unwrap();
    let bin = stalled_gh(tmp.path(), "gh-stalled-timeout")
        .to_str()
        .unwrap()
        .to_string();
    let req = HttpRequest::new("GET", "https://api.github.com/repos/acme/widgets");
    let (tx, rx) = std::sync::mpsc::channel();
    let started = Instant::now();
    std::thread::spawn(move || {
        let t = GhTransport::new(bin).timeout(Duration::from_secs(2));
        let _ = tx.send(t.send(&req));
    });
    // The stub sleeps 120s; the caller must get an answer long before that. #2511: the bound
    // is far above the 2s timeout so machine load cannot fail it, and far below the stub's life.
    let answer = rx
        .recv_timeout(Duration::from_secs(60))
        .expect("GhTransport::send did not return within 60s on a stalled gh");
    let err = answer.expect_err("a stalled gh is not a response");
    assert!(started.elapsed() < Duration::from_secs(60));
    assert!(err.contains("timed out"), "{err}");

    // The grandchild that held the pipes was ended with the group.
    let pid = fs::read_to_string(tmp.path().join("gh-stalled-timeout.pid")).unwrap();
    assert!(
        // Returns as soon as the grandchild is gone; 30s only bounds a broken group kill.
        process_ends_within(pid.trim(), Duration::from_secs(30)),
        "sleep grandchild {} still running",
        pid.trim()
    );

    // Through the adapter the timeout is the retryable `transport` kind.
    let t = GhTransport::new(
        stalled_gh(tmp.path(), "gh-stalled-adapter")
            .to_str()
            .unwrap(),
    )
    .timeout(Duration::from_millis(500));
    let e = forge::read(&request("open"), None, &t, &FakeClock::at(1_794_000_000)).unwrap_err();
    assert_eq!(e.kind, "transport");
    assert!(
        e.message.starts_with("gh transport: gh did not answer"),
        "{}",
        e.message
    );
}
