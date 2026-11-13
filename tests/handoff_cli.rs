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
use moai_cli::forge::transport::{Exchange, RecordedTransport};
use moai_cli::handoff::{claim_handoff, effect_confirm, effect_intend, effects_reconcile};
mod common;

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
    TaskRef::new(
        "github",
        "github.com",
        "acme",
        "widgets",
        101,
        7000 + issue,
        issue,
    )
}

fn worktree_id(path: &Path, head: &str) -> WorktreeId {
    WorktreeId::new(
        path.parent()
            .unwrap()
            .join(".git")
            .to_string_lossy()
            .into_owned(),
        path.join(".git").to_string_lossy().into_owned(),
        path.to_string_lossy().into_owned(),
        head,
    )
}

fn owner(kind: &str, id: &str) -> Owner {
    Owner::new(kind, id)
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
        assignee: None,
        provider_session: None,
        integration_owner: None,
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

/// Seeds the journal for a test whose next step is the CLI binary, without taking the
/// journal lock. A lock held in this multithreaded test binary is an `flock` on an open
/// file description; any child another test forks in that window keeps a copy of the
/// descriptor until its `exec`, so the lock outlives our drop and the CLI's `try_lock`
/// answers `unavailable` (Issue #2572, rc-ci run 43253440881). Nothing else writes this
/// temporary journal, so the lock buys nothing here.
fn seed_journal(repo: &Path, event: &str, record: &ClaimRecord) {
    let journal = Journal::open(&journal_dir(repo)).unwrap();
    journal.commit(event, record).unwrap();
}

fn write_probe(dir: &Path, name: &str, script: &str) -> PathBuf {
    let path = dir.join(name);
    common::write_executable(&path, &format!("#!/bin/sh\n{script}\n"));
    path
}

fn exchange(method: &str, url: &str, status: u16, body: Value) -> Exchange {
    exchange_with_headers(method, url, status, body, &[])
}

fn exchange_with_headers(
    method: &str,
    url: &str,
    status: u16,
    body: Value,
    headers: &[(&str, &str)],
) -> Exchange {
    let mut response = common::json_response(status, body);
    // Response headers are stored lower-cased, the way the real gh parser does.
    for (k, v) in headers {
        response = response.with_header(k.to_ascii_lowercase(), *v);
    }
    common::exchange(method, url, response)
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
    // An absent pid is dead, not unknown. The pid belongs to a process that has exited
    // (#2349: an out-of-range pid makes `ps` print an error, which is indeterminate).
    let mut exited = Command::new("/usr/bin/true").spawn().unwrap();
    let exited_pid = exited.id();
    exited.wait().unwrap();
    let (gone, ev) = moai_cli::handoff::probe_liveness(Some(exited_pid), Some("x"));
    assert_eq!(gone, "dead", "{ev}");

    child.kill().unwrap();
    let _ = child.wait();
}

/// A directory holding a fake `ps` with the given script body.
fn fake_ps(body: &str) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let ps = dir.path().join("ps");
    common::write_executable(&ps, &format!("#!/bin/sh\n{body}\n"));
    (dir, ps)
}

/// #2349 (f): only `ps`'s own "no such process" answer (exit 1, no output, no error text)
/// means dead. Any other failure says nothing about the owner and is indeterminate.
#[test]
fn liveness_ps_failure_is_indeterminate_not_dead() {
    let (_d1, denied) = fake_ps("echo 'ps: permission denied' >&2\nexit 2");
    let (verdict, ev) = moai_cli::handoff::probe_liveness_with(&denied, Some(4242), Some("x"));
    assert_eq!(verdict, "indeterminate", "{ev}");
    assert_eq!(ev["exit_code"], 2);
    assert!(ev["reason"].as_str().unwrap().starts_with("ps failed"));

    let (_d2, err1) = fake_ps("echo 'ps: sysctl failed' >&2\nexit 1");
    let (verdict, ev) = moai_cli::handoff::probe_liveness_with(&err1, Some(4242), Some("x"));
    assert_eq!(verdict, "indeterminate", "exit 1 with error text: {ev}");

    let (_d3, other) = fake_ps("exit 3");
    let (verdict, ev) = moai_cli::handoff::probe_liveness_with(&other, Some(4242), Some("x"));
    assert_eq!(verdict, "indeterminate", "silent exit 3: {ev}");

    let (_d4, absent) = fake_ps("exit 1");
    let (verdict, ev) = moai_cli::handoff::probe_liveness_with(&absent, Some(4242), Some("x"));
    assert_eq!(verdict, "dead", "{ev}");
    assert_eq!(ev["reason"], "no such process");
}

/// #2349 (f) end to end: a failing `ps` must not let `claim-handoff` take over a live
/// owner without `--force-explicit`; the owner process keeps running.
#[test]
fn claim_handoff_with_failing_ps_stays_pending() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let (owner, mut child, host) = live_sleep_owner();
    let record = seed_record(&repo, 28, 1, owner, vec![]);
    let key = record.task_key.clone();
    seed_journal(&repo, "acquire", &record);
    let (bin, _ps) = fake_ps("echo 'ps: permission denied' >&2\nexit 2");
    let out = Command::new(crate::common::MOAI_BIN)
        .args([
            "claim-handoff",
            ".",
            "--task",
            &key,
            "--from-epoch",
            "1",
            "--reason",
            "probe with a failing ps",
            "--host-id",
            &host,
        ])
        .current_dir(&repo)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.path().display()))
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .stdin(Stdio::null())
        .output()
        .unwrap();
    let v = envelope(&out);
    assert_eq!(v["result"]["outcome"], "handoff_pending", "{v}");
    assert_eq!(v["result"]["liveness"], "indeterminate");
    assert_eq!(v["result"]["verdict"], "blocked");
    assert!(
        child.try_wait().unwrap().is_none(),
        "the owner process is still running"
    );
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
        Owner::new("local-process", "agent-9")
            .host_id("mac-1")
            .pid(pid)
            .process_start(start),
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

/// AC-1 (#2133): a branch name with query hazards (`+` `#` `&` `=` — all legal git
/// branch characters) reaches GitHub percent-encoded, so the recorded URL matches the
/// encoded form byte for byte, `state=all` survives the query intact, and the hit
/// confirms. The fixture branch is check-ref-format legal (no `~ ^ : ? * [`, no `..`).
#[test]
fn reconcile_head_query_encodes_special_branch_names() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
    effect["kind"] = json!("pr_create");
    let mut record = seed_record(&repo, 38, 1, owner("session", "agent-x"), vec![effect]);
    record.worktree.branch = Some("hotfix/r+2026#v2&w=1".into());
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    // Raw `#` would truncate the query before `state=all`; raw `&` would inject a
    // parameter; `+` would decode as a space. Encoded, none of that can happen. The
    // slash stays raw: RFC 3986 query allows it and GitHub accepts it (#2126 precedent).
    let head_url = "https://api.github.com/repos/acme/widgets/pulls?head=acme:hotfix/r%2B2026%23v2%26w%3D1&state=all&per_page=100";
    let full_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let t = RecordedTransport::new(vec![exchange(
        "GET",
        head_url,
        200,
        json!([{"number": 77, "body": format!("Implements it\n\n<!-- moai-op pr_create:101:op-9:1 -->")}]),
    )]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();

    assert_eq!(out["verdict"], "passed", "{}", out);
    assert_eq!(out["transitions"][0]["to"], "confirmed");
    let sent = t.sent.borrow();
    assert_eq!(sent.len(), 1, "only the encoded head query ran");
    assert_eq!(
        sent[0].url, head_url,
        "the URL matches the encoded form byte for byte"
    );
    assert!(!sent.iter().any(|r| r.url == full_url));
}

/// AC-1 (#2126): when the record knows its branch, the pr_create effect is judged from
/// the narrow `pulls?head=<owner>:<branch>` query — a hit confirms with ZERO full-list
/// GETs (asserted from the sent log). The `moai-op` spelling matches through the
/// identity comparison.
#[test]
fn reconcile_pr_create_head_query_confirms_without_full_scan() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
    effect["kind"] = json!("pr_create");
    let mut record = seed_record(&repo, 35, 1, owner("session", "agent-x"), vec![effect]);
    record.worktree.branch = Some("issue/2007-06-C-e2".into());
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    let head_url = "https://api.github.com/repos/acme/widgets/pulls?head=acme:issue/2007-06-C-e2&state=all&per_page=100";
    let full_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let page = json!([
        {"number": 77, "state": "open", "body": format!("Implements it\n\n<!-- moai-op pr_create:101:op-9:1 -->")},
    ]);
    let t = RecordedTransport::new(vec![exchange("GET", head_url, 200, page)]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();

    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["transitions"][0]["to"], "confirmed");
    assert_eq!(out["transitions"][0]["response_id"], "77");
    let sent = t.sent.borrow();
    assert_eq!(sent.len(), 1, "only the head query ran");
    assert_eq!(sent[0].url, head_url);
    assert!(
        !sent.iter().any(|r| r.url == full_url),
        "AC-1: a head-query hit performs zero full-list GETs"
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
    assert_eq!(effect["state"], "confirmed");
}

/// AC-1 fallback (#2126): the head query answering empty falls back to the full scan —
/// both URLs run, and the full scan's marker confirms.
#[test]
fn reconcile_pr_create_head_empty_falls_back_to_full_scan() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
    effect["kind"] = json!("pr_create");
    let mut record = seed_record(&repo, 36, 1, owner("session", "agent-x"), vec![effect]);
    record.worktree.branch = Some("renamed-branch".into());
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    // The branch query answers empty (the PR was made from a different branch), and the
    // full scan finds the marker.
    let head_url = "https://api.github.com/repos/acme/widgets/pulls?head=acme:renamed-branch&state=all&per_page=100";
    let full_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let t = RecordedTransport::new(vec![
        exchange("GET", head_url, 200, json!([])),
        exchange(
            "GET",
            full_url,
            200,
            json!([{"number": 90, "body": format!("x\n\n<!-- moai-op pr_create:101:op-9:1 -->")}]),
        ),
    ]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["transitions"][0]["to"], "confirmed");
    assert_eq!(out["transitions"][0]["response_id"], "90");
    let sent = t.sent.borrow();
    assert_eq!(sent.len(), 2, "head query, then the full scan");
    assert_eq!(sent[0].url, head_url);
    assert_eq!(sent[1].url, full_url);
}

/// AC-1 ambiguity (#2126): two PRs on the branch carrying the same operation identity —
/// the winner is unclear, the state stays uncertain, and the full scan is not consulted.
#[test]
fn reconcile_pr_create_head_ambiguous_stays_uncertain() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
    effect["kind"] = json!("pr_create");
    let mut record = seed_record(&repo, 37, 1, owner("session", "agent-x"), vec![effect]);
    record.worktree.branch = Some("issue/2007-06-C-e2".into());
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    let head_url = "https://api.github.com/repos/acme/widgets/pulls?head=acme:issue/2007-06-C-e2&state=all&per_page=100";
    let full_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let t = RecordedTransport::new(vec![exchange(
        "GET",
        head_url,
        200,
        json!([
            {"number": 77, "body": format!("one\n\n<!-- moai-op pr_create:101:op-9:1 -->")},
            {"number": 78, "body": format!("two\n\n<!-- moai:op pr_create:101:op-9:1 -->")},
        ]),
    )]);
    let before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "uncertain");
    assert!(out["transitions"][0]["why"]
        .as_str()
        .unwrap()
        .contains("ambiguous"));
    assert!(
        !t.sent.borrow().iter().any(|r| r.url == full_url),
        "an ambiguous head answer is not retried against the full list"
    );
    assert_eq!(
        fs::read(journal_dir(&repo).join("journal.log")).unwrap(),
        before
    );
}

/// #2329: two PRs of one task and epoch — the first with the original marker, the second
/// with the `--pr-seq 2` marker (`<task_key>-2`) — are told apart by the full scan: the
/// second PR's uncertain effect confirms against the second PR only, and the first PR's
/// old-form marker is still read as its own operation.
#[test]
fn reconcile_tells_two_prs_of_one_epoch_apart() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let first_marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let second_marker = "<!-- moai:op pr_create:101:op-9-2:1 -->";
    let mut first = publish_effect("pr_create-101-op-9-1", first_marker, "uncertain");
    first["kind"] = json!("pr_create");
    let mut second = publish_effect("pr_create-101-op-9-2-1", second_marker, "uncertain");
    second["kind"] = json!("pr_create");
    let record = seed_record(
        &repo,
        38,
        1,
        owner("session", "agent-x"),
        vec![first, second],
    );
    let key = record.task_key.clone();
    let journal = journal_of(&repo);
    journal.commit("acquire", &record).unwrap();

    let full_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let t = RecordedTransport::new(vec![exchange(
        "GET",
        full_url,
        200,
        json!([
            {"number": 2328, "body": "second\n\n<!-- moai-op pr_create:101:op-9-2:1 -->"},
            {"number": 2324, "body": "first\n\n<!-- moai-op pr_create:101:op-9:1 -->"},
        ]),
    )]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "passed", "{out}");
    let by_op: std::collections::BTreeMap<String, String> = out["transitions"]
        .as_array()
        .unwrap()
        .iter()
        .map(|t| {
            (
                t["operation_id"].as_str().unwrap().to_string(),
                t["response_id"].as_str().unwrap_or_default().to_string(),
            )
        })
        .collect();
    assert_eq!(by_op["pr_create-101-op-9-1"], "2324");
    assert_eq!(by_op["pr_create-101-op-9-2-1"], "2328");
}

/// AC-1/AC-2 (#2099), updated by #2107: a pr_create effect routes to the pull-list
/// lookup, never to the comments — a comment carrying the same operation's marker in the
/// `moai-op` spelling must NOT confirm a PR. When the pull lookup cannot answer (the
/// recorded script here has no exchange for it), the effect stays uncertain.
#[test]
fn reconcile_pr_create_routes_to_pulls_not_comments() {
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

    // Only a comments exchange is recorded; if the probe correctly routes pr_create to
    // the pull list, this page is never fetched and never confirms anything.
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
    assert!(
        t.sent.borrow().iter().all(|r| !r.url.contains("/issues/")),
        "pr_create is judged from the pull list, not the comments: {:?}",
        t.sent
            .borrow()
            .iter()
            .map(|r| r.url.clone())
            .collect::<Vec<_>>()
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
        "an unanswered pull lookup leaves the state"
    );
    assert_eq!(
        out["verdict"], "uncertain",
        "the lookup failed, so the run is uncertain"
    );
    assert_eq!(
        fs::read(journal_dir(&repo).join("journal.log")).unwrap(),
        before,
        "no judged transition writes nothing"
    );
}

/// AC-1 (#2107): the built-in pull lookup judges a pr_create effect from the PR body
/// marker — identity matching makes both spellings one marker. Exactly one PR carrying
/// it confirms; none is a confirmed absence; more than one is ambiguous and stays
/// uncertain. GETs only.
#[test]
fn reconcile_builtin_judges_pr_create_from_pulls() {
    let pulls_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let clock = FakeClock::at(CLOCK0);
    let effect_for = |issue: i64, marker: &str| {
        let tmp = tempfile::tempdir().unwrap();
        let repo = tmp.path().join("repo");
        git_repo(&repo);
        let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
        effect["kind"] = json!("pr_create");
        let record = seed_record(&repo, issue, 1, owner("session", "agent-x"), vec![effect]);
        let key = record.task_key.clone();
        let journal = journal_of(&repo);
        journal.commit("acquire", &record).unwrap();
        (tmp, repo, key, journal)
    };
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";

    // Found: one open PR body carries the marker in the moai-op spelling.
    let (_t1, _repo, key, journal) = effect_for(29, marker);
    let page = json!([
        {"number": 77, "body": format!("Implements it\n\n<!-- moai-op pr_create:101:op-9:1 -->")},
        {"number": 78, "body": "unrelated"},
    ]);
    let t = RecordedTransport::new(vec![exchange("GET", pulls_url, 200, page)]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["transitions"][0]["to"], "confirmed");
    assert_eq!(out["transitions"][0]["response_id"], "77");
    let sent = t.sent.borrow();
    assert!(sent.iter().all(|r| r.method == "GET"));
    assert!(
        sent.iter().all(|r| r.url.contains("/pulls")),
        "the probe reads the pull list"
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
    assert_eq!(effect["state"], "confirmed");

    // Absent: no open PR carries it — a confirmed absence, judged from the PR list.
    let (_t2, _repo, key, journal) = effect_for(30, marker);
    let t = RecordedTransport::new(vec![exchange(
        "GET",
        pulls_url,
        200,
        json!([{"number": 78, "body": "unrelated"}]),
    )]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["transitions"][0]["to"], "not_performed");

    // Ambiguous: two PRs carry the same operation identity — the winner is unclear, so
    // the state stays and the run is uncertain (a duplicate would be hiding here).
    let (_t3, repo, key, journal) = effect_for(31, marker);
    let page = json!([
        {"number": 77, "body": format!("one\n\n<!-- moai-op pr_create:101:op-9:1 -->")},
        {"number": 78, "body": format!("two\n\n<!-- moai:op pr_create:101:op-9:1 -->")},
    ]);
    let t = RecordedTransport::new(vec![exchange("GET", pulls_url, 200, page)]);
    let before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "uncertain");
    assert!(out["transitions"][0]["why"]
        .as_str()
        .unwrap()
        .contains("ambiguous"));
    let events = journal.read_events().unwrap();
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "pr_create-101-op-9-1")
        .unwrap();
    assert_eq!(effect["state"], "uncertain");
    assert_eq!(
        fs::read(journal_dir(&repo).join("journal.log")).unwrap(),
        before
    );
}

/// #2116 repair: a not_performed verdict is only ever a FULL-scan verdict. A merged or
/// closed PR carrying the marker still confirms (the #2101 no-false-absence rule —
/// scanning open PRs only would have judged it absent and triggered a duplicate PR); a
/// scan that hit the page cap never saw everything, so the effect stays uncertain; only
/// a scan that saw every page with zero hits may say not_performed.
#[test]
fn reconcile_pr_create_absence_requires_a_full_scan() {
    let pulls_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100";
    let clock = FakeClock::at(CLOCK0);
    let marker = "<!-- moai:op pr_create:101:op-9:1 -->";
    let setup = |issue: i64| {
        let tmp = tempfile::tempdir().unwrap();
        let repo = tmp.path().join("repo");
        git_repo(&repo);
        let mut effect = publish_effect("pr_create-101-op-9-1", marker, "uncertain");
        effect["kind"] = json!("pr_create");
        let record = seed_record(&repo, issue, 1, owner("session", "agent-x"), vec![effect]);
        let key = record.task_key.clone();
        let journal = journal_of(&repo);
        journal.commit("acquire", &record).unwrap();
        // `tmp` stays alive with the caller: dropping it deletes the journal.
        (tmp, repo, key, journal)
    };

    // ① The marker lives in a CLOSED PR's body: confirmed, never a false absence.
    let (_t, _repo, key, journal) = setup(32);
    let page = json!([
        {"number": 77, "state": "closed", "body": format!("merged\n\n<!-- moai-op pr_create:101:op-9:1 -->")},
        {"number": 78, "state": "merged", "body": "unrelated"},
    ]);
    let t = RecordedTransport::new(vec![exchange("GET", pulls_url, 200, page)]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["transitions"][0]["to"], "confirmed");
    assert_eq!(out["transitions"][0]["response_id"], "77");
    let events = journal.read_events().unwrap();
    let effect = events
        .last()
        .unwrap()
        .record
        .effects
        .iter()
        .find(|e| e["operation_id"] == "pr_create-101-op-9-1")
        .unwrap();
    assert_eq!(effect["state"], "confirmed");

    // ② The scan hits the 50-page cap before seeing everything: the effect stays
    // uncertain — a capped scan is never a confirmed absence — and nothing is written.
    // The recorded request URLs follow the Link chain exactly (page 1, 2, ...); the cap
    // fires on the 51st iteration before any request is sent.
    let (_t, repo, key, journal) = setup(33);
    let page_url = |n: usize| -> String {
        if n == 0 {
            pulls_url.to_string()
        } else {
            format!("{pulls_url}&page={n}")
        }
    };
    let mut script = Vec::new();
    for i in 0..50 {
        let body = json!([{"number": 1000 + i, "state": "closed", "body": "unrelated"}]);
        let link = format!("<{}>; rel=\"next\"", page_url(i + 1));
        script.push(exchange_with_headers(
            "GET",
            &page_url(i),
            200,
            body,
            &[("Link", link.as_str())],
        ));
    }
    let before = fs::read(journal_dir(&repo).join("journal.log")).unwrap();
    let t = RecordedTransport::new(script);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "uncertain", "{}", out);
    assert!(out["transitions"][0]["why"]
        .as_str()
        .unwrap()
        .contains("page limit"));
    assert_eq!(
        fs::read(journal_dir(&repo).join("journal.log")).unwrap(),
        before,
        "a capped scan writes nothing"
    );

    // ③ A scan that saw every page (two here) with zero hits is a confirmed absence.
    let (_t, _repo, key, journal) = setup(34);
    let page2_url = "https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100&page=2";
    let t = RecordedTransport::new(vec![
        exchange_with_headers(
            "GET",
            pulls_url,
            200,
            json!([{"number": 78, "state": "open", "body": "unrelated"}]),
            &[(
                "Link",
                "<https://api.github.com/repos/acme/widgets/pulls?state=all&per_page=100&page=2>; rel=\"next\"",
            )],
        ),
        exchange("GET", page2_url, 200, json!([{"number": 79, "body": "also unrelated"}])),
    ]);
    let out = effects_reconcile(&journal, &key, None, &t, &clock).unwrap();
    assert_eq!(out["verdict"], "passed");
    assert_eq!(out["transitions"][0]["to"], "not_performed");
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
    let mut child = Command::new(crate::common::MOAI_BIN)
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
    seed_journal(&main, "acquire", &record);
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
    seed_journal(&main, "acquire", &next);
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
    seed_journal(&main, "acquire", &record);
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

/// #2392 D3-C (decision 6774899518 row 15 = A, test only): an effect whose result was lost
/// (`uncertain`) must be reconciled before it is sent again. Re-intending it with the same key
/// and payload today answers `outcome: intended, verdict: passed`, so a caller that sends on a
/// passed intent sends a second comment. The repair answers `outcome: needs_reconcile,
/// verdict: blocked` and points to `effects-reconcile`; the journal stays unchanged.
#[test]
fn d3c_reintending_an_uncertain_effect_is_not_a_fresh_intent() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    git_repo(&repo);
    let clock = FakeClock::at(CLOCK0);
    let (journal, key, _) = reconcile_probe_setup(&repo, "uncertain");
    let before = journal.read_events().unwrap().len();
    let out = effect_intend(
        &journal,
        &key,
        1,
        "comment_post",
        "acme",
        "op-9",
        1,
        "comment_post",
        b"payload",
        &clock,
    )
    .unwrap();
    assert_eq!(out["state"], "uncertain", "{out}");
    assert_eq!(journal.read_events().unwrap().len(), before, "{out}");
    assert_ne!(
        out["verdict"], "passed",
        "an uncertain effect was answered as a passed intent, inviting a duplicate send: {out}"
    );
    assert_eq!(out["outcome"], "needs_reconcile", "{out}");
    assert_eq!(out["verdict"], "blocked", "{out}");
    assert!(
        out["notes"].to_string().contains("effects-reconcile"),
        "{out}"
    );

    // Boundary: a still `intended` effect re-intended with the same payload stays a passed
    // intent (the existing idempotent answer), so only the lost result is blocked.
    let repo = tmp.path().join("repo-intended");
    git_repo(&repo);
    let (journal, key, _) = reconcile_probe_setup(&repo, "intended");
    let out = effect_intend(
        &journal,
        &key,
        1,
        "comment_post",
        "acme",
        "op-9",
        1,
        "comment_post",
        b"payload",
        &clock,
    )
    .unwrap();
    assert_eq!(out["state"], "intended", "{out}");
    assert_eq!(out["verdict"], "passed", "{out}");
}

/// Issue #2572 made deterministic: a child is forked in another thread and held between
/// its fork and its `exec` (the window a parallel test's `Command::spawn` opens by chance
/// under load). The parent learns the fork happened from a byte the child writes, runs
/// `seed`, and lets the child exec only after the CLI answered. Event-driven, no sleeps.
#[cfg(unix)]
fn handoff_while_a_child_sits_between_fork_and_exec(
    seed: impl FnOnce(&Path, &ClaimRecord) -> Option<Journal>,
) -> Value {
    use std::io::Read;
    use std::os::unix::net::UnixStream;
    use std::os::unix::process::CommandExt;

    let tmp = tempfile::tempdir().unwrap();
    let main = tmp.path().join("main");
    git_repo(&main);
    let record = seed_record(&main, 33, 4, owner("session", "agent-x"), vec![]);
    let key = record.task_key.clone();
    // Seeded (and, for the old helper, still locked) before the fork, as in the flake.
    let held = seed(&main, &record);

    let (mut forked, child_side) = UnixStream::pair().unwrap();
    let (release, mut child_wait) = UnixStream::pair().unwrap();
    let pause = thread::spawn(move || {
        let mut cmd = Command::new("/usr/bin/true");
        // SAFETY: the closure only makes write/read system calls on descriptors created
        // before the fork; it allocates nothing and takes no lock.
        unsafe {
            cmd.pre_exec(move || {
                (&child_side).write_all(&[1])?;
                let mut b = [0u8; 1];
                let _ = child_wait.read(&mut b);
                Ok(())
            });
        }
        cmd.status().unwrap()
    });
    let mut b = [0u8; 1];
    forked.read_exact(&mut b).unwrap();
    drop(held);

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
    (&release).write_all(&[1]).unwrap();
    assert!(pause.join().unwrap().success());
    envelope(&out)
}

/// The cause of #2572: a journal lock taken in this process and dropped while a forked
/// child has not exec'd yet is still held by that child, so the CLI answers busy.
#[cfg(unix)]
#[test]
fn a_lock_held_at_a_fork_outlives_its_drop_until_the_child_execs() {
    let v = handoff_while_a_child_sits_between_fork_and_exec(|repo, record| {
        let journal = journal_of(repo);
        journal.commit("acquire", record).unwrap();
        Some(journal)
    });
    assert_eq!(v["error"]["kind"], json!("unavailable"), "{v}");
}

/// The repair: `seed_journal` never takes the lock, so a child forked at any moment holds
/// no lock and the CLI reaches its own answer (here the stale-request conflict). Seeding
/// while another handle holds the lock proves it takes none: a lock taken and dropped
/// inside the helper would leave the same window a fork can fall into.
#[cfg(unix)]
#[test]
fn seed_journal_leaves_no_lock_for_a_forked_child() {
    let v = handoff_while_a_child_sits_between_fork_and_exec(|repo, record| {
        let other = journal_of(repo);
        seed_journal(repo, "acquire", record);
        drop(other);
        None
    });
    assert_eq!(v["error"]["kind"], json!("conflict"), "{v}");
    assert!(v["error"]["message"]
        .as_str()
        .unwrap()
        .contains("stale_request"));
}
