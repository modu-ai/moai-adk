//! The one journal lock-busy contract (Issue #2345, AC-1): with the claim journal lock
//! held, every journal-writing operation answers the same failure envelope — `ok: false`,
//! kind `unavailable`, `retryable: true`, exit 1, one shared message. Before #2345 the
//! ops answered per op: `claim-acquire`/`claim-release`/`queue-next` returned an `ok: true`
//! `busy` outcome, the 06-B/06-C ops a `conflict` error the classifier made non-retryable.
//! Each case runs the real binary as a subprocess against a real git repository while this
//! test process holds `<git common dir>/.moai-cli/claims/.lock`.

mod common;

use serde_json::{json, Value};
use std::{
    fs::{self, File},
    io::Write,
    path::{Path, PathBuf},
    process::{Command, Stdio},
    thread,
    time::Duration,
};

fn now_utc() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs() as i64
}

fn git(dir: &Path, args: &[&str]) {
    let status = Command::new("git")
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .current_dir(dir)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_SYSTEM", "/dev/null")
        .status()
        .unwrap();
    assert!(status.success(), "git {args:?} failed");
}

fn git_repo(dir: &Path) {
    fs::create_dir_all(dir).unwrap();
    git(dir, &["init", "-q"]);
    git(
        dir,
        &[
            "-c",
            "user.email=test@example.invalid",
            "-c",
            "user.name=test",
            "commit",
            "--allow-empty",
            "-q",
            "-m",
            "init",
        ],
    );
}

/// A stateful `gh` stub (read-your-writes) for the one setup acquire, from
/// `tests/claim_cli.rs`. The busy cases themselves must never reach `gh`: their stub
/// exits loudly instead, so a busy answer that got past the lock cannot pass.
fn claim_stub(dir: &Path) -> PathBuf {
    let script = "#!/bin/sh\n# claim test stub for gh api (stateful: read-your-writes)\ncase \"$*\" in\n  *per_page=100*)\n    if [ -f \"$CLAIM_STUB_DIR/posted\" ]; then printf 'HTTP/2.0 200 OK\\r\\n\\r\\n[{\"id\":%s,\"body\":\"%s\"}]' \"${CLAIM_STUB_ID:-4242}\" \"$CLAIM_STUB_BODY\";\n    else printf 'HTTP/2.0 200 OK\\r\\n\\r\\n[]'; fi;;\n  *)\n    : > \"$CLAIM_STUB_DIR/posted\"\n    printf 'HTTP/2.0 201 Created\\r\\n\\r\\n{\"id\":%s}' \"${CLAIM_STUB_ID:-4242}\";;\nesac\n";
    let path = dir.join("gh-stub");
    common::write_executable(&path, script);
    path
}

/// A `gh` stub that fails loudly: any journal-writing op that got past a held lock shows
/// up as a transport failure instead of a busy envelope.
fn loud_gh(dir: &Path) -> PathBuf {
    let path = dir.join("gh-loud");
    common::write_executable(&path, "#!/bin/sh\n# busy cases must never reach gh\necho \"gh reached while the lock was held\" >&2\nexit 64\n");
    path
}

fn acquire_request(lease_until: &str, session: &str) -> String {
    json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com",
            "owner": "modu-ai", "repo": "moai-adk",
            "repository_id": 9001, "issue_id": 123456789, "issue_number": 2005,
        },
        "logical_key": "busy-test",
        "owner": {"kind": "session", "id": session},
        "attempt": 1,
        "base_revision": "a".repeat(40),
        "write_scope": ["src/busy.rs"],
        "spec_ref": {"id": "SPEC-BUSY", "revision": "r1", "digest": null},
        "lease_until": lease_until,
        "projection_intents": [],
    })
    .to_string()
}

fn queue_request(lease_until: &str) -> String {
    json!({
        "schema_version": 1,
        "owner": "modu-ai",
        "repo": "moai-adk",
        "milestone": 3,
        "claim": {
            "owner": {"kind": "session", "id": "busy-test"},
            "base_revision": "a".repeat(40),
            "write_scope": ["src/busy.rs"],
            "spec_ref": {"id": "SPEC-BUSY", "revision": "r1", "digest": null},
            "lease_until": lease_until,
        },
    })
    .to_string()
}

fn run_op(
    repo: &Path,
    args: &[&str],
    stdin: Option<&str>,
    home: &Path,
    gh: &Path,
    stub_body: &str,
) -> std::process::Output {
    let mut child = Command::new(crate::common::MOAI_BIN)
        .args(args)
        .current_dir(repo)
        .env("HOME", home)
        .env("MOAI_FORGE_GH_BIN", gh)
        // The stateful stub keeps its posted marker next to itself (see claim_stub); the
        // value is unused by the loud stub the busy cases run with.
        .env("CLAIM_STUB_DIR", gh.parent().unwrap())
        .env("CLAIM_STUB_BODY", stub_body)
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

fn envelope(op: &str, out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "{op}: stdout is not JSON: {e}\n{}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn journal_dir(repo: &Path) -> PathBuf {
    repo.join(".git").join(".moai-cli").join("claims")
}

/// Holds the journal lock for the duration of the returned guard.
fn hold_journal_lock(repo: &Path) -> File {
    let dir = journal_dir(repo);
    fs::create_dir_all(&dir).unwrap();
    let lock = fs::OpenOptions::new()
        .read(true)
        .write(true)
        .create(true)
        .truncate(false)
        .open(dir.join(".lock"))
        .unwrap();
    lock.try_lock()
        .expect("the test must hold the journal lock");
    lock
}

#[test]
fn every_journal_writing_op_answers_one_busy_envelope() {
    let tmp = tempfile::tempdir().unwrap();
    let repo = tmp.path().join("repo");
    let home = tmp.path().join("home");
    fs::create_dir_all(&home).unwrap();
    git_repo(&repo);
    fs::create_dir_all(repo.join(".moai-cli")).unwrap();
    fs::write(
        repo.join(".moai-cli/project.yaml"),
        r#"{"schema_version":1,"project_id":"busy-fixture","purpose":"busy contract","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"]}"#,
    )
    .unwrap();
    let gh = claim_stub(tmp.path());
    let loud = loud_gh(tmp.path());
    let lease = moai_cli::forge::policy::format_utc(now_utc() + 3600);

    // Setup: one real claim, so the --task ops find a record and reach their lock path.
    let claim_body = format!("moai-claim: session=busy-test epoch=1 lease_until={lease}");
    let out = run_op(
        &repo,
        &["claim-acquire", "."],
        Some(&acquire_request(&lease, "busy-test")),
        &home,
        &gh,
        &claim_body,
    );
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v = envelope("claim-acquire", &out);
    assert_eq!(v["result"]["outcome"], json!("claimed"));
    let key = v["result"]["record"]["task_key"].as_str().unwrap();
    let head = {
        let o = Command::new("git")
            .args(["rev-parse", "HEAD"])
            .current_dir(&repo)
            .output()
            .unwrap();
        String::from_utf8(o.stdout).unwrap().trim().to_string()
    };
    let receipts = tmp.path().join("receipts.json");
    fs::write(&receipts, "[]").unwrap();

    // A second, overlapping active record, so scope-check takes its blocked write path
    // (a claim never conflicts with itself — scope.rs skips the same task_key).
    let other: moai_cli::claim::ClaimRecord = serde_json::from_value(json!({
        "schema_version": 1,
        "task_ref": {
            "provider": "github", "host": "github.com",
            "owner": "modu-ai", "repo": "moai-adk",
            "repository_id": 9001, "issue_id": 123456790, "issue_number": 2006,
        },
        "task_key": "ffffffffffffffff",
        "logical_key": "busy-other",
        "owner": {"kind": "session", "id": "busy-other"},
        "epoch": 1,
        "base_revision": "a".repeat(40),
        "worktree": {"common_dir": "", "git_dir": "", "path": "", "head": "a".repeat(40)},
        "write_scope": ["src/busy.rs"],
        "spec_ref": {"id": "SPEC-BUSY", "revision": "r1", "digest": null},
        "state": "running",
        "last_observed_at": "2026-11-10T00:00:00Z",
    }))
    .unwrap();
    moai_cli::claim::Journal::open(&journal_dir(&repo))
        .unwrap()
        .commit("scope_check", &other)
        .unwrap();

    let holder = hold_journal_lock(&repo);

    // (CLI args, stdin, expected operation name) — one row per journal-writing op.
    let cases: Vec<(Vec<String>, Option<String>, &str)> = vec![
        (
            vec![
                "layout-apply".into(),
                ".".into(),
                "--plan-digest".into(),
                "0".repeat(64),
                "--permission-reference".into(),
                "test:busy-layout".into(),
            ],
            None,
            "layout.apply",
        ),
        (
            vec!["claim-acquire".into(), ".".into()],
            Some(acquire_request(&lease, "busy-test-2")),
            "claim.acquire",
        ),
        (
            vec!["queue-next".into(), ".".into()],
            Some(queue_request(&lease)),
            "queue.next",
        ),
        (
            vec![
                "claim-release".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--epoch".into(),
                "1".into(),
                "--reason".into(),
                "busy-test".into(),
            ],
            None,
            "claim.release",
        ),
        (
            vec![
                "claim-handoff".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--from-epoch".into(),
                "1".into(),
                "--reason".into(),
                "busy-test".into(),
            ],
            None,
            "claim.handoff",
        ),
        (
            vec![
                "effect-intend".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--epoch".into(),
                "1".into(),
                "--operation".into(),
                "test_op".into(),
                "--project-id".into(),
                "p".into(),
                "--local-id".into(),
                "l".into(),
                "--attempt".into(),
                "1".into(),
                "--kind".into(),
                "comment_post".into(),
            ],
            Some("{}".into()),
            "effect.intend",
        ),
        (
            vec![
                "effect-confirm".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--operation-id".into(),
                "op1".into(),
                "--response-id".into(),
                "r1".into(),
                "--observed-state".into(),
                "done".into(),
            ],
            None,
            "effect.confirm",
        ),
        (
            vec![
                "effects-reconcile".into(),
                ".".into(),
                "--task".into(),
                key.into(),
            ],
            None,
            "effects.reconcile",
        ),
        (
            vec![
                "worktree-apply".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--branch-prefix".into(),
                "issue/".into(),
                "--plan-digest".into(),
                "0".repeat(64),
            ],
            None,
            "worktree.apply",
        ),
        (
            vec![
                "scope-check".into(),
                ".".into(),
                "--task".into(),
                key.into(),
            ],
            None,
            "scope.check",
        ),
        (
            vec![
                "pr-body-render".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--receipts".into(),
                receipts.display().to_string(),
                "--base-sha".into(),
                head.clone(),
                "--base-branch".into(),
                "v2026.11.0".into(),
                "--default-branch".into(),
                "main".into(),
            ],
            None,
            "pr.body.render",
        ),
        (
            vec![
                "pr-confirm".into(),
                ".".into(),
                "--task".into(),
                key.into(),
                "--operation-id".into(),
                "op1".into(),
                "--response-id".into(),
                "r1".into(),
                "--observed-state".into(),
                "opened".into(),
            ],
            None,
            "pr.confirm",
        ),
    ];
    assert_eq!(
        cases.len(),
        12,
        "all journal-writing operations are covered"
    );

    let mut shared_message: Option<String> = None;
    for (args, stdin, op) in &cases {
        let args: Vec<&str> = args.iter().map(String::as_str).collect();
        let out = run_op(&repo, &args, stdin.as_deref(), &home, &loud, "unused");
        assert_eq!(
            out.status.code(),
            Some(1),
            "{op}: a held journal lock must exit 1, got success: {}",
            String::from_utf8_lossy(&out.stdout)
        );
        let v = envelope(op, &out);
        assert_eq!(v["protocol_version"], json!(1), "{op}");
        assert_eq!(v["operation"], json!(op), "{op}");
        assert_eq!(v["ok"], json!(false), "{op}");
        let error = &v["error"];
        assert_eq!(
            error.as_object().map(|o| o.len()),
            Some(3),
            "{op}: error carries exactly kind, message, retryable: {error}"
        );
        assert_eq!(error["kind"], json!("unavailable"), "{op}");
        assert_eq!(error["retryable"], json!(true), "{op}");
        let message = error["message"].as_str().unwrap().to_string();
        assert_eq!(message, "journal busy: the claim journal lock is unavailable (another process holds it); the caller may retry the same command after a bounded wait, within its own cap (REQ-06A-05)", "{op}: literal busy contract");
        // The single busy contract: one message across every op, set at the raise site.
        match &shared_message {
            None => shared_message = Some(message.clone()),
            Some(first) => assert_eq!(&message, first, "{op}: busy messages diverged"),
        }
        // stderr carries the same one human line.
        assert_eq!(
            String::from_utf8_lossy(&out.stderr).trim(),
            message,
            "{op}: stderr line"
        );
    }

    // Receipt rule-out: the same command succeeds once the holder is gone, so the busy
    // answers above came from the held lock, not from broken fixtures.
    drop(holder);
    for _ in 0..200 {
        let out = run_op(
            &repo,
            &[
                "claim-release",
                ".",
                "--task",
                key,
                "--epoch",
                "1",
                "--reason",
                "busy released",
            ],
            None,
            &home,
            &gh,
            &claim_body,
        );
        let v = envelope("claim-release", &out);
        if out.status.success() && v["ok"] == json!(true) {
            return;
        }
        // macOS can answer a just-dropped lock late (see tests/claim_cli.rs); retry.
        thread::sleep(Duration::from_millis(10));
    }
    panic!("claim-release did not succeed within 2s of dropping the journal lock");
}
