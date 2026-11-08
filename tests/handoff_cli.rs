//! 06-B (#2006) stalled-claim re-inspection and explicit handoff. Library tests drive the
//! handoff/effects API against a journal in a temporary git repository with an injected
//! clock; remote probes are fake probe commands (never the network). CLI tests run the
//! real binary against a real repository and a linked worktree. Live GitHub round trips
//! are not_run (the built-in comment probe is exercised through RecordedTransport only).
use moai_cli::claim::{
    is_active, is_key, ClaimRecord, Journal, Owner, SpecRef, TaskRef, WorktreeId,
};
use moai_cli::contract;
use moai_cli::forge::policy::{format_utc, Clock};
use moai_cli::forge::transport::{Exchange, HttpRequest, HttpResponse, RecordedTransport};
use moai_cli::handoff::{claim_handoff, effect_confirm, effect_intend, effects_reconcile};
use serde_json::{json, Value};
use std::{
    cell::{Cell, RefCell},
    fs,
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    thread,
    time::Duration,
};

const CLOCK0: i64 = 1_794_000_000;

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
    // The built-in probe's rate-limit retries wait through the injected clock (never the
    // wall clock); recording the waits keeps the test instant.
    fn sleep(&self, secs: u64) {
        self.slept.borrow_mut().push(secs);
        self.now.set(self.now.get() + secs as i64);
    }
}

fn task_ref(issue: i64) -> TaskRef {
    TaskRef {
        provider: "github".into(),
        host: "github.com".into(),
        owner: "acme".into(),
        repo: "widgets".into(),
        repository_id: 101,
        issue_id: 7000 + issue,
        issue_number: issue,
    }
}

fn worktree_id(path: &Path, head: &str) -> WorktreeId {
    WorktreeId {
        common_dir: path
            .parent()
            .unwrap()
            .join(".git")
            .to_string_lossy()
            .into_owned(),
        git_dir: path.join(".git").to_string_lossy().into_owned(),
        path: path.to_string_lossy().into_owned(),
        head: head.to_string(),
    }
}

fn owner(kind: &str, id: &str) -> Owner {
    Owner {
        kind: kind.into(),
        id: id.into(),
        host_id: None,
        pid: None,
        process_start: None,
    }
}

fn seed_record(
    repo: &Path,
    issue: i64,
    epoch: u64,
    owner: Owner,
    effects: Vec<Value>,
) -> ClaimRecord {
    let task = task_ref(issue);
    let head = git(repo, &["rev-parse", "HEAD"]);
    ClaimRecord {
        schema_version: 1,
        task_ref: task.clone(),
        task_key: moai_cli::claim::task_key(&task),
        logical_key: "06-B".into(),
        owner,
        attempt: Some(1),
        epoch,
        base_revision: "a".repeat(40),
        worktree: worktree_id(repo, head.trim()),
        write_scope: vec!["src/".into()],
        spec_ref: SpecRef {
            id: "SPEC-RC1-06-B".into(),
            revision: "r1".into(),
            digest: None,
        },
        comment_id: None,
        claim_token: None,
        lease_until: None,
        state: "running".into(),
        effects,
        last_observed_at: format_utc(CLOCK0),
    }
}

fn git(dir: &Path, args: &[&str]) -> String {
    let out = Command::new("git")
        // #2092 (as #2082/#2088): background maintenance can leave temporary lock files
        // while a snapshot comparison runs — turn it off for every fixture git call.
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .env("LC_ALL", "C")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "git {args:?} failed: {}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8_lossy(&out.stdout).into_owned()
}

fn git_repo(dir: &Path) {
    fs::create_dir_all(dir).unwrap();
    git(dir, &["init", "-q"]);
    fs::write(dir.join("tracked.txt"), b"one\n").unwrap();
    git(
        dir,
        &[
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
            "add",
            "tracked.txt",
        ],
    );
    git(
        dir,
        &[
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
            "commit",
            "-q",
            "-m",
            "init",
        ],
    );
}

fn journal_dir(repo: &Path) -> PathBuf {
    repo.join(".git").join(".moai-cli").join("claims")
}

fn journal_of(repo: &Path) -> Journal {
    // Caller-side waiting policy (REQ-06A-05): macOS has been observed to answer
    // WouldBlock for one poll after a same-process drop under parallel test load.
    for _ in 0..200 {
        let j = Journal::open(&journal_dir(repo)).unwrap();
        if j.try_lock().unwrap() {
            return j;
        }
        drop(j);
        thread::sleep(Duration::from_millis(10));
    }
    panic!("journal lock not acquired within 2s");
}

fn write_probe(dir: &Path, name: &str, script: &str) -> PathBuf {
    let path = dir.join(name);
    fs::write(&path, format!("#!/bin/sh\n{script}\n")).unwrap();
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(&path, fs::Permissions::from_mode(0o755)).unwrap();
    }
    path
}

fn exchange(method: &str, url: &str, status: u16, body: Value) -> Exchange {
    Exchange {
        request: HttpRequest {
            method: method.into(),
            url: url.into(),
            headers: Default::default(),
            body: None,
        },
        response: HttpResponse {
            status,
            headers: Default::default(),
            body: Some(body),
        },
    }
}

fn comments_url(issue: i64) -> String {
    format!("https://api.github.com/repos/acme/widgets/issues/{issue}/comments?per_page=100")
}

fn publish_effect(operation_id: &str, marker: &str, state: &str) -> Value {
    json!({
        "kind": "comment_post",
        "operation_id": operation_id,
        "intent_digest": moai_cli::digest(b"payload"),
        "state": state,
        "epoch": 1,
        "marker": marker,
    })
}

// ---- liveness (REQ-06B-02) ----

#[test]
fn liveness_reads_pid_and_start_time() {
    let mut child = Command::new("/bin/sleep").arg("30").spawn().unwrap();
    let pid = child.id();
    thread::sleep(Duration::from_millis(120));
    let start = Command::new("ps")
        .args(["-p", &pid.to_string(), "-o", "lstart="])
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    let start = String::from_utf8_lossy(&start.stdout).trim().to_string();

    let (alive, _) = moai_cli::handoff::probe_liveness(Some(pid), Some(&start));
    assert_eq!(alive, "alive");
    // A different recorded start time means the pid was reused: dead.
    let (reused, ev) =
        moai_cli::handoff::probe_liveness(Some(pid), Some("Thu Jan  1 00:00:00 2000"));
    assert_eq!(reused, "dead");
    assert!(ev["reason"] == "pid reused");
    // No recorded start: the answer is honest indeterminate, never a guess.
    let (unknown, ev) = moai_cli::handoff::probe_liveness(Some(pid), None);
    assert_eq!(unknown, "indeterminate");
    assert!(ev["reason"] == "no recorded start time");
    // An absent pid is dead, not unknown.
    let (gone, _) = moai_cli::handoff::probe_liveness(Some(999_999_999), Some("x"));
    assert_eq!(gone, "dead");

    child.kill().unwrap();
    let _ = child.wait();
}

// ---- handoff gate and transition (REQ-06B-01..05) ----

fn live_sleep_owner() -> (Owner, std::process::Child, String) {
    let child = Command::new("/bin/sleep").arg("30").spawn().unwrap();
    let pid = child.id();
    thread::sleep(Duration::from_millis(120));
    let out = Command::new("ps")
        .args(["-p", &pid.to_string(), "-o", "lstart="])
        .env("LC_ALL", "C")
        .output()
        .unwrap();
    let start = String::from_utf8_lossy(&out.stdout).trim().to_string();
    (
        Owner {
            kind: "local-process".into(),
            id: "agent-9".into(),
            host_id: Some("mac-1".into()),
            pid: Some(pid),
            process_start: Some(start),
        },
        child,
        "mac-1".into(),
    )
}

#[test]
fn handoff_requires_explicit_when_alive() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let (owner, mut child, host) = live_sleep_owner();
    let record = seed_record(&repo, 21, 1, owner, vec![]);
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    let clock = FakeClock::at(CLOCK0);
    // REQ-06B-03: an alive owner is refused without the explicit flag, and the refusal is
    // marked handoff_pending (still active — it keeps blocking new acquires).
    let out = claim_handoff(
        &journal,
        &key,
        1,
        "leader handoff",
        false,
        Some(&host),
        &clock,
    )
    .unwrap();
    assert_eq!(out["outcome"], "handoff_pending");
    assert_eq!(out["verdict"], "blocked");
    assert_eq!(out["liveness"], "alive");
    let pending = journal.read_events().unwrap();
    let last = pending.last().unwrap();
    assert_eq!(last.record.state, "handoff_pending");
    assert!(is_active("handoff_pending"));
    assert_eq!(last.record.effects[0]["kind"], "liveness");

    // With the explicit flag the handoff proceeds and closes the record as handed_off.
    let out = claim_handoff(
        &journal,
        &key,
        1,
        "leader handoff",
        true,
        Some(&host),
        &clock,
    )
    .unwrap();
    assert_eq!(out["outcome"], "handed_off");
    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["to_epoch"], 2);
    let events = journal.read_events().unwrap();
    let closed = events.last().unwrap();
    assert_eq!(closed.record.state, "handed_off");
    assert!(!is_active("handed_off"));
    let kinds: Vec<&str> = closed
        .record
        .effects
        .iter()
        .filter_map(|e| e["kind"].as_str())
        .collect();
    assert!(kinds.contains(&"liveness"));
    assert!(kinds.contains(&"worktree_evidence"));
    assert!(kinds.contains(&"worktree_lock_intent"));
    assert!(kinds.contains(&"handoff"));
    // REQ-06B-01: a request naming another epoch than the record's is a stale_request
    // refusal, even against a closed record.
    let err = claim_handoff(&journal, &key, 2, "late", true, Some(&host), &clock).unwrap_err();
    assert!(err.contains("stale_request"), "{err}");
    assert_eq!(contract::classify(&err), contract::Kind::Conflict);
    // A late request naming the closed record's own epoch reads as not active (unclaimed).
    let late = claim_handoff(&journal, &key, 1, "late", true, Some(&host), &clock).unwrap();
    assert_eq!(late["outcome"], "unclaimed");
    // The handed-off record no longer blocks: the next acquire writes epoch 2 = to_epoch.
    assert_eq!(moai_cli::claim::task_key(&task_ref(21)), key);
    child.kill().unwrap();
    let _ = child.wait();
}

#[test]
fn handoff_remote_owner_no_pid_probe() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    // A remote-task owner with a bogus pid: if the probe ran, liveness would be dead.
    let mut owner = owner("remote-task", "cloud-runner");
    owner.host_id = Some("remote-1".into());
    owner.pid = Some(999_999_999);
    owner.process_start = Some("whatever".into());
    let record = seed_record(&repo, 22, 3, owner, vec![]);
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    let clock = FakeClock::at(CLOCK0);
    let out = claim_handoff(&journal, &key, 3, "handoff", false, Some("mac-1"), &clock).unwrap();
    assert_eq!(
        out["liveness"], "not_applicable",
        "no local probe for a remote owner"
    );
    assert_eq!(out["outcome"], "handed_off");
}

// ---- fencing (REQ-06B-08) and stale_request (REQ-06B-01) ----

#[test]
fn handoff_stale_epoch_fenced() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let record = seed_record(&repo, 23, 2, owner("session", "agent-x"), vec![]);
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();
    let clock = FakeClock::at(CLOCK0);

    // A publish intent below the current epoch is fenced: stale_epoch refusal, and the
    // request is kept only as a stale_result diagnostic.
    let err = effect_intend(
        &journal,
        &key,
        1,
        "comment_post",
        "acme",
        "op-1",
        1,
        "comment_post",
        b"payload-a",
        &clock,
    )
    .unwrap_err();
    assert!(err.contains("stale_epoch"), "{err}");
    assert_eq!(contract::classify(&err), contract::Kind::Conflict);
    let events = journal.read_events().unwrap();
    let diag = events.last().unwrap();
    assert_eq!(diag.event, "stale_result");
    assert!(diag
        .record
        .effects
        .iter()
        .any(|e| e["kind"] == "stale_result"));
    // The diagnostic carries the refused request; no publish effect was created.
    assert!(!diag
        .record
        .effects
        .iter()
        .any(|e| e["operation_id"].is_string()));

    // REQ-06B-01 on the handoff side: a from_epoch below the current is stale_request.
    let err = claim_handoff(&journal, &key, 1, "late", false, None, &clock).unwrap_err();
    assert!(err.contains("stale_request"), "{err}");

    // A current-epoch intent on the active record goes through.
    let out = effect_intend(
        &journal,
        &key,
        2,
        "comment_post",
        "acme",
        "op-1",
        1,
        "comment_post",
        b"payload-a",
        &clock,
    )
    .unwrap();
    assert_eq!(out["outcome"], "intended");
    assert_eq!(out["state"], "intended");
    let marker = out["marker"].as_str().unwrap().to_string();
    assert!(marker.starts_with("<!-- moai:op comment_post:acme:op-1:1 -->"));

    // The same key with the same payload is the same intent (draft-08 semantics,
    // borrowed locally); the same key with another payload is a conflict.
    let again = effect_intend(
        &journal,
        &key,
        2,
        "comment_post",
        "acme",
        "op-1",
        1,
        "comment_post",
        b"payload-a",
        &clock,
    )
    .unwrap();
    assert_eq!(
        again["notes"][0],
        "the same key and payload were already intended; nothing changed"
    );
    let err = effect_intend(
        &journal,
        &key,
        2,
        "comment_post",
        "acme",
        "op-1",
        1,
        "comment_post",
        b"payload-b",
        &clock,
    )
    .unwrap_err();
    assert!(err.contains("different payload"), "{err}");
}

// ---- effects confirm + reconcile (REQ-06B-06/07/09/10) ----

fn reconcile_probe_setup(repo: &Path, state: &str) -> (Journal, String, String) {
    let marker = "<!-- moai:op comment_post:acme:op-9:1 -->";
    let record = seed_record(
        repo,
        24,
        1,
        owner("session", "agent-x"),
        vec![publish_effect("comment_post-acme-op-9-1", marker, state)],
    );
    let key = record.task_key.clone();
    let journal = journal_of(repo);
    journal.commit("acquire", &record).unwrap();
    (journal, key, marker.to_string())
}

#[test]
fn reconcile_uncertain_no_duplicate() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let found = write_probe(
        tmp.path(),
        "probe-found",
        "printf '{\"remote_id\": \"c-123\"}'\nexit 0",
    );
    let absent = write_probe(tmp.path(), "probe-absent", "exit 4");
    let clock = FakeClock::at(CLOCK0);

    // Uncertain + the marker found on the remote: confirmed with the remote id. The
    // reconcile has no write path at all, so nothing is re-created (REQ-06B-07).
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    let out = effects_reconcile(
        &journal,
        &key,
        Some(found.to_str().unwrap()),
        &RecordedTransport::new(vec![]),
        &clock,
    )
    .unwrap();
    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["transitions"][0]["to"], "confirmed");
    assert_eq!(out["transitions"][0]["response_id"], "c-123");
    let events = journal.read_events().unwrap();
    assert_eq!(events.last().unwrap().event, "effects_reconcile");
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "comment_post-acme-op-9-1")
        .unwrap();
    assert_eq!(effect["state"], "confirmed");
    assert_eq!(effect["response_id"], "c-123");

    // Uncertain + the marker confirmed absent: not_performed — still no re-creation.
    drop(journal); // release this test's handle before the next setup takes the lock
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    let out = effects_reconcile(
        &journal,
        &key,
        Some(absent.to_str().unwrap()),
        &RecordedTransport::new(vec![]),
        &clock,
    )
    .unwrap();
    assert_eq!(out["transitions"][0]["to"], "not_performed");
    let events = journal.read_events().unwrap();
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "comment_post-acme-op-9-1")
        .unwrap();
    assert_eq!(effect["state"], "not_performed");
}

#[test]
fn reconcile_idempotent() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    // A counting probe: every invocation appends to a log file.
    let log = tmp.path().join("probe.log");
    let probe = write_probe(
        tmp.path(),
        "probe-count",
        &format!(
            "printf '%s\\n' \"$1\" >> {}\nprintf '{{\"remote_id\": \"c-1\"}}'\nexit 0",
            log.display()
        ),
    );
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    let log_before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    let first = effects_reconcile(
        &journal,
        &key,
        Some(probe.to_str().unwrap()),
        &RecordedTransport::new(vec![]),
        &clock,
    )
    .unwrap();
    assert_eq!(first["verdict"], "passed");
    let after_first = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    assert_ne!(
        log_before, after_first,
        "the first run transitions and writes once"
    );

    // The second identical run: nothing is uncertain any more, so nothing is probed and
    // the journal is byte-identical (AC-06B-5).
    let count_after_first = fs::read_to_string(&log).unwrap().lines().count();
    let second = effects_reconcile(
        &journal,
        &key,
        Some(probe.to_str().unwrap()),
        &RecordedTransport::new(vec![]),
        &clock,
    )
    .unwrap();
    assert_eq!(second["verdict"], "not_run");
    let after_second = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    assert_eq!(after_first, after_second, "the journal is untouched");
    let count_after_second = fs::read_to_string(&log).unwrap().lines().count();
    assert_eq!(count_after_first, count_after_second, "no probe ran again");
}

#[test]
fn reconcile_rate_limited_waits() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    // Exit 5 + seconds: rate limited; the state must not change (REQ-06B-09).
    let limited = write_probe(tmp.path(), "probe-429", "printf '30'\nexit 5");
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    let before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    let out = effects_reconcile(
        &journal,
        &key,
        Some(limited.to_str().unwrap()),
        &RecordedTransport::new(vec![]),
        &clock,
    )
    .unwrap();
    assert_eq!(out["verdict"], "blocked");
    assert!(out["notes"][0].as_str().unwrap().contains("retry-after"));
    let after = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    assert_eq!(before, after, "a rate-limited probe writes nothing");
    let events = journal.read_events().unwrap();
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "comment_post-acme-op-9-1")
        .unwrap();
    assert_eq!(
        effect["state"], "uncertain",
        "state unchanged until retry-after"
    );
}

/// AC-1/AC-2 (#2099): the built-in probe reads Issue comments only, so it cannot judge a
/// pr_create effect — even when a comment happens to carry the same operation's marker in
/// the `moai-op` spelling. A false "confirmed absent" is a duplicate-PR hazard, because
/// sync.md retries only what not_performed names.
#[test]
fn reconcile_builtin_never_judges_pr_create() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
    effect["kind"] = json!("pr_create");
    let record = seed_record(&repo, 28, 1, owner("session", "agent-x"), vec![effect]);
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    // The PR marker appears in a comment body in the moai-op spelling — a literal
    // contains() misses it, and judging a PR from comments would be wrong anyway.
    let page = json!([{"id": 5001, "body": "report\n\n<!-- moai-op pr_create:101:op-9:1 -->"}]);
    let t = RecordedTransport::new(vec![exchange("GET", &comments_url(28), 200, page)]);
    let before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();

    assert!(
        !out["transitions"]
            .as_array()
            .unwrap()
            .iter()
            .any(|x| x["to"] == "not_performed"),
        "a PR must never be judged absent from comments: {}",
        out
    );
    let events = journal.read_events().unwrap();
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "pr_create-101-op-9-1")
        .unwrap();
    assert_eq!(
        effect["state"], "uncertain",
        "comments cannot judge a pr_create; the state stays"
    );
    // Nothing was judged at all: not_run, with the probe requirement named.
    assert_eq!(out["verdict"], "not_run");
    assert!(out["transitions"][0]["why"]
        .as_str()
        .unwrap()
        .contains("--remote-probe"));
    assert_eq!(
        fs::read(journal_dir(&repo).join("journal.log")).unwrap(),
        before,
        "no judged transition writes nothing"
    );
}

#[test]
fn reconcile_builtin_gh_probe_reads_only() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op comment_post:acme:op-7:1 -->";
    let record = seed_record(
        &repo,
        25,
        1,
        owner("session", "agent-x"),
        vec![
            publish_effect("comment_post-acme-op-7-1", marker, "uncertain"),
            publish_effect(
                "comment_post-acme-op-8-1",
                "<!-- moai:op comment_post:acme:op-8:1 -->",
                "uncertain",
            ),
        ],
    );
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    // The built-in probe GETs the comments once (both effects, one page) and finds only
    // op-7's marker; op-8 is confirmed absent. GETs only — the sent log proves it.
    let page = json!([
        {"id": 4242, "body": format!("report\n\n{marker}")},
        {"id": 4243, "body": "unrelated"},
    ]);
    let t = RecordedTransport::new(vec![exchange("GET", &comments_url(25), 200, page)]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "passed");
    let sent = t.sent.borrow();
    assert!(
        sent.iter().all(|r| r.method == "GET"),
        "reconcile sends GET only"
    );
    let transitions = out["transitions"].as_array().unwrap();
    assert_eq!(transitions.len(), 2);
    let confirmed = transitions.iter().find(|x| x["to"] == "confirmed").unwrap();
    assert_eq!(confirmed["response_id"], "4242");
    assert!(transitions.iter().any(|x| x["to"] == "not_performed"));
    assert_eq!(t.remaining(), 0);

    // A rate-limited read path (403 + retry-after that outlives the budget) blocks the
    // whole reconcile without touching any state.
    drop(journal);
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    let before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    // The rate limit holds for the whole retry budget (initial + 3 retries).
    let limited = || {
        exchange(
            "GET",
            &comments_url(24),
            403,
            json!({"message": "API rate limit exceeded"}),
        )
    };
    let t = RecordedTransport::new(vec![limited(), limited(), limited(), limited()]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "blocked");
    assert_eq!(
        fs::read(journal_dir(&repo).join("journal.log")).unwrap(),
        before
    );
}

// ---- confirm path (REQ-06B-06) ----

#[test]
fn effect_confirm_marks_intended_only() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    // Confirming an effect that was never intended is an error.
    let err = effect_confirm(&journal, &key, "no-such-op", "c-1", "posted", &clock).unwrap_err();
    assert!(err.contains("unknown operation_id"), "{err}");

    // Intend (current epoch), then confirm; a second confirm is refused.
    effect_intend(
        &journal,
        &key,
        1,
        "comment_post",
        "acme",
        "op-c",
        1,
        "comment_post",
        b"p",
        &clock,
    )
    .unwrap();
    let out = effect_confirm(
        &journal,
        &key,
        "comment_post-acme-op-c-1",
        "c-77",
        "posted",
        &clock,
    )
    .unwrap();
    assert_eq!(out["outcome"], "confirmed");
    assert_eq!(out["response_id"], "c-77");
    let err = effect_confirm(
        &journal,
        &key,
        "comment_post-acme-op-c-1",
        "c-77",
        "posted",
        &clock,
    )
    .unwrap_err();
    assert!(err.contains("already"), "{err}");

    // The interrupted case: confirm names uncertain and the effect stays reconcilable.
    effect_intend(
        &journal,
        &key,
        1,
        "comment_post",
        "acme",
        "op-d",
        1,
        "comment_post",
        b"p",
        &clock,
    )
    .unwrap();
    let out = effect_confirm(
        &journal,
        &key,
        "comment_post-acme-op-d-1",
        "",
        "uncertain",
        &clock,
    )
    .unwrap();
    assert_eq!(out["outcome"], "uncertain");
    let events = journal.read_events().unwrap();
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "comment_post-acme-op-d-1")
        .unwrap();
    assert_eq!(effect["state"], "uncertain");
}

// ---- CLI level: real git, dirty worktree preservation (REQ-06B-04, AC-06B-4) ----

fn cli(repo: &Path, args: &[&str], stdin: Option<&str>) -> std::process::Output {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .current_dir(repo)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(if stdin.is_some() {
            Stdio::piped()
        } else {
            Stdio::null()
        })
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    if let (Some(text), Some(mut sin)) = (stdin, child.stdin.take()) {
        match sin.write_all(text.as_bytes()) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
    }
    child.wait_with_output().unwrap()
}

fn envelope(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn hash_tree(root: &Path) -> Vec<String> {
    fn walk(root: &Path, dir: &Path, out: &mut Vec<String>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
            if p.is_dir() {
                walk(root, &p, out);
            } else {
                let bytes = fs::read(&p).unwrap();
                out.push(format!("{rel} {}", moai_cli::digest(&bytes)));
            }
        }
    }
    let mut out = Vec::new();
    walk(root, root, &mut out);
    out.sort();
    out
}

#[test]
fn handoff_cli_preserves_dirty_worktree_and_records_lock_intent() {
    let tmp = tempfile::tempdir().unwrap();
    let main = tmp.path().join("main");
    git_repo(&main);
    let wt = tmp.path().join("wt");
    git(
        &main,
        &[
            "worktree",
            "add",
            "-q",
            "-b",
            "wt-branch",
            wt.to_str().unwrap(),
        ],
    );
    git(
        &wt,
        &[
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
            "checkout",
            "-q",
            "wt-branch",
        ],
    );

    // Dirty three ways: a modified tracked file, a new untracked file, a deleted one.
    fs::write(wt.join("tracked.txt"), b"changed\n").unwrap();
    fs::write(wt.join("untracked.txt"), b"precious\n").unwrap();
    fs::remove_file(wt.join("tracked.txt")).unwrap();
    // (modified then deleted: still exactly one deletion; add another modified file)
    fs::write(wt.join("untracked.txt"), b"precious v2\n").unwrap();
    let status_before = git(&wt, &["status", "--porcelain"]);
    let hashes_before = hash_tree(&wt);
    assert!(!status_before.trim().is_empty());

    // Seed an active record whose worktree is the linked worktree, then hand off through
    // the real CLI with an operator owner (liveness not_applicable → no flag needed).
    let record = seed_record(&wt, 26, 1, owner("operator", "leader"), vec![]);
    let key = record.task_key.clone();
    {
        let journal = journal_of(&main);
        journal.commit("acquire", &record).unwrap();
    }
    let out = cli(
        &main,
        &[
            "claim-handoff",
            ".",
            "--task",
            &key,
            "--from-epoch",
            "1",
            "--reason",
            "lane stalled after push; explicit leader handoff",
        ],
        None,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope(&out);
    assert_eq!(v["result"]["outcome"], "handed_off");
    assert_eq!(v["result"]["liveness"], "not_applicable");
    assert_eq!(v["result"]["verdict"], "passed");

    // AC-06B-4: the dirty worktree is byte-for-byte what it was; no git mutation ran.
    assert_eq!(git(&wt, &["status", "--porcelain"]), status_before);
    assert_eq!(hash_tree(&wt), hashes_before);
    assert!(
        git(&main, &["stash", "list"]).trim().is_empty(),
        "no stash was created"
    );

    // REQ-06B-05: the worktree lock was recorded as an intent, never executed —
    // `git worktree list --porcelain` shows no locked line for the linked worktree.
    let effect = fs::read_to_string(journal_dir(&main).join("journal.log")).unwrap();
    let last_line = effect.lines().last().unwrap();
    let closed: Value = serde_json::from_str(last_line).unwrap();
    assert_eq!(closed["event"], "handoff");
    assert_eq!(closed["record"]["state"], "handed_off");
    let effects = closed["record"]["effects"].as_array().unwrap();
    let lock_intent = effects
        .iter()
        .find(|e| e["kind"] == "worktree_lock_intent")
        .expect("lock intent recorded");
    assert_eq!(lock_intent["state"], "intended");
    assert!(lock_intent["command"]
        .as_str()
        .unwrap()
        .contains("git worktree lock --reason"));
    let porcelain = git(&main, &["worktree", "list", "--porcelain"]);
    assert!(
        !porcelain.contains("locked"),
        "the worktree lock was not executed: {porcelain}"
    );
    // The evidence entry carries the worktree snapshot.
    let evidence = effects
        .iter()
        .find(|e| e["kind"] == "worktree_evidence")
        .unwrap();
    assert_eq!(evidence["evidence"]["status"], "collected");
    assert!(evidence["evidence"]["status_v2_z"].as_str().is_some());

    // The effects pipeline through the CLI: intend (fenced at a stale epoch), confirm
    // uncertain, reconcile by probe — the full product path (AC-06B-1/2 at CLI level).
    let err_out = cli(
        &main,
        &[
            "effect-intend",
            ".",
            "--task",
            &key,
            "--epoch",
            "0",
            "--operation",
            "comment_post",
            "--project-id",
            "acme",
            "--local-id",
            "op-x",
            "--attempt",
            "1",
            "--kind",
            "comment_post",
        ],
        Some(r#"{"body":"x"}"#),
    );
    assert!(!err_out.status.success());
    let v = envelope(&err_out);
    assert_eq!(v["error"]["kind"], json!("conflict"));
    assert!(v["error"]["message"]
        .as_str()
        .unwrap()
        .contains("stale_epoch"));

    let probe = write_probe(
        tmp.path(),
        "cli-probe",
        "printf '{\"remote_id\": \"c-9\"}'\nexit 0",
    );
    // The new owner's acquire writes the to_epoch record (epoch 2); publish intents
    // attach to that active epoch, not to the closed one.
    let next = seed_record(&wt, 26, 2, owner("session", "agent-7"), vec![]);
    assert_eq!(next.task_key, key, "the same task, the fenced epoch");
    {
        let journal = journal_of(&main);
        journal.commit("acquire", &next).unwrap();
    }
    let intend = cli(
        &main,
        &[
            "effect-intend",
            ".",
            "--task",
            &key,
            "--epoch",
            "2",
            "--operation",
            "comment_post",
            "--project-id",
            "acme",
            "--local-id",
            "op-x",
            "--attempt",
            "1",
            "--kind",
            "comment_post",
        ],
        Some(r#"{"body":"report text"}"#),
    );
    assert!(
        intend.status.success(),
        "{}",
        String::from_utf8_lossy(&intend.stderr)
    );
    let v = envelope(&intend);
    let operation_id = v["result"]["operation_id"].as_str().unwrap().to_string();
    assert_eq!(v["result"]["state"], "intended");

    let confirm = cli(
        &main,
        &[
            "effect-confirm",
            ".",
            "--task",
            &key,
            "--operation-id",
            &operation_id,
            "--observed-state",
            "uncertain",
        ],
        None,
    );
    assert!(
        confirm.status.success(),
        "{}",
        String::from_utf8_lossy(&confirm.stderr)
    );
    assert_eq!(envelope(&confirm)["result"]["outcome"], "uncertain");

    let reconcile = cli(
        &main,
        &[
            "effects-reconcile",
            ".",
            "--task",
            &key,
            "--remote-probe",
            probe.to_str().unwrap(),
        ],
        None,
    );
    assert!(
        reconcile.status.success(),
        "{}",
        String::from_utf8_lossy(&reconcile.stderr)
    );
    let v = envelope(&reconcile);
    assert_eq!(v["result"]["verdict"], "passed");
    assert_eq!(v["result"]["transitions"][0]["to"], "confirmed");
    assert_eq!(v["result"]["transitions"][0]["response_id"], "c-9");
}

#[test]
fn handoff_cli_refuses_stale_request_and_unknown_key() {
    let tmp = tempfile::tempdir().unwrap();
    let main = tmp.path().join("main");
    git_repo(&main);
    let record = seed_record(&main, 27, 4, owner("session", "agent-x"), vec![]);
    let key = record.task_key.clone();
    {
        let journal = journal_of(&main);
        journal.commit("acquire", &record).unwrap();
    }
    // Stale from_epoch: refused with kind conflict and the message names stale_request.
    let out = cli(
        &main,
        &[
            "claim-handoff",
            ".",
            "--task",
            &key,
            "--from-epoch",
            "3",
            "--reason",
            "late request",
        ],
        None,
    );
    assert!(!out.status.success());
    let v = envelope(&out);
    assert_eq!(v["error"]["kind"], json!("conflict"));
    assert!(v["error"]["message"]
        .as_str()
        .unwrap()
        .contains("stale_request"));

    // Unknown key shape is invalid input, not a probe.
    let out = cli(
        &main,
        &[
            "claim-handoff",
            ".",
            "--task",
            "zz",
            "--from-epoch",
            "1",
            "--reason",
            "x",
        ],
        None,
    );
    assert!(!out.status.success());
    assert_eq!(envelope(&out)["error"]["kind"], json!("invalid_input"));
    assert!(is_key("0123456789abcdef") && !is_key("zz"));
}
