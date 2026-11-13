//! RUN methodology and real VERIFY (Issue #2003, 05-C).
mod common;
use moai_cli::plan::{admit, AdmissionRequest, Approval, Task};
use moai_cli::run::{
    admissible_evidence, budget_exhausted, ingest_self_report, oracle_changes, red_confirmed,
    remaining_evidence, required_evidence, resolve_method, revalidate, run, stale_reasons, Budget,
    Method, PinnedRef, Receipt, RunRequest, VerificationRequest,
};
use moai_cli::spec::{approve, SortedKeys, SpecDocument};
use moai_cli::verify::{classify, parse_argv, run_check, CheckResult, FailureClass};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
    time::Duration,
};

const REV: &str = "1111111111111111111111111111111111111111";

fn root(rel: &str) -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)
}
/// The 05-A fixture with the given verification commands, approved through the 05-A API.
fn spec_with(commands: &[&str]) -> SpecDocument {
    let mut d: SpecDocument = serde_json::from_slice(
        &fs::read(root("tests/fixtures/spec/snapshot_canonical_a.json")).unwrap(),
    )
    .unwrap();
    d.verification_commands = commands.iter().map(|c| c.to_string()).collect();
    approve(&d, &SortedKeys)
}
fn admission(spec: SpecDocument, revision: &str) -> AdmissionRequest {
    let snap = spec.snapshot.clone().unwrap();
    AdmissionRequest {
        protocol_version: 1,
        spec,
        tasks: vec![Task {
            id: "T1".into(),
            depends_on: vec![],
            write_paths: vec!["src/x.rs".into()],
            verifier_refs: vec!["AC-S-1".into()],
            host: None,
        }],
        approval: Approval {
            id: "APR-1".into(),
            approver: "operator".into(),
            basis: "plan review".into(),
            snapshot_reference: snap.reference,
            snapshot_digest: snap.digest,
            allowed_actions: vec!["run".into()],
            revoked: false,
        },
        expected_revision: revision.into(),
        capabilities: vec![],
    }
}
fn method(strategy: &str, domain: &str) -> Method {
    Method {
        change_strategy: Some(strategy.into()),
        domain_modeling: Some(domain.into()),
        legacy_mode: None,
    }
}
/// A request whose pinned contract is the one 05-B admits right now.
fn request(commands: &[&str], refs: &[usize], revision: &str) -> RunRequest {
    let adm = admission(spec_with(commands), revision);
    let decision = admit(&adm, revision);
    let pinned: PinnedRef =
        serde_json::from_value(serde_json::to_value(&decision.pinned[0]).unwrap()).unwrap();
    RunRequest {
        protocol_version: 1,
        intent: "RUN".into(),
        request_id: "REQ-1".into(),
        task_id: "T1".into(),
        attempt_id: "A1".into(),
        owner: None,
        pinned,
        admission: adm,
        method: method("tdd", "not-required"),
        phase: None,
        verification: VerificationRequest {
            commands_ref: refs.to_vec(),
            timeout_ms: 30_000,
        },
        baseline: None,
        baseline_ref: None,
        diff_paths: vec![],
    }
}
fn project() -> tempfile::TempDir {
    let dir = tempfile::TempDir::new().unwrap();
    fs::write(dir.path().join("README.md"), "fixture\n").unwrap();
    dir
}
fn v(x: impl serde::Serialize) -> Value {
    serde_json::to_value(x).unwrap()
}

#[test]
fn contract_mismatch_no_exec() {
    let dir = project();
    let marker = dir.path().join("marker.txt");
    let cmds = ["touch marker.txt"];

    // The repository moved after admission: stale, nothing runs.
    let out = v(run(
        dir.path(),
        &request(&cmds, &[0], REV),
        "2222222222222222222222222222222222222222",
    )
    .unwrap());
    assert_eq!(out["decision"], "stale", "{out}");
    assert!(out["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("expected-revision-mismatch")));
    assert!(out["receipt"].is_null());
    assert!(!marker.exists());

    // The pinned contract the caller holds differs from what 05-B admits now.
    let mut req = request(&cmds, &[0], REV);
    req.pinned.plan_digest.pop();
    req.pinned.plan_digest.push('x');
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["decision"], "stale");
    assert!(out["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("pinned-contract-mismatch")));
    assert!(!marker.exists());

    // The recorded snapshot digest no longer matches the spec content.
    let mut req = request(&cmds, &[0], REV);
    let snap = req.admission.spec.snapshot.as_mut().unwrap();
    let last = snap.digest.pop().unwrap();
    snap.digest.push(if last == '0' { '1' } else { '0' });
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["decision"], "stale");
    assert!(!marker.exists());

    // A task the plan does not admit.
    let mut req = request(&cmds, &[0], REV);
    req.task_id = "T9".into();
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["decision"], "blocked");
    assert!(out["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("task-not-admitted")));
    assert!(!marker.exists());

    // Control: with everything matching, the command does run.
    let out = v(run(dir.path(), &request(&cmds, &[0], REV), REV).unwrap());
    assert_eq!(out["decision"], "ready", "{out}");
    assert!(marker.exists());
}

#[test]
fn red_classification_counts_only_assertion_failures() {
    use FailureClass::*;
    // Build failure, missing command, timeout and network errors are environment failures.
    let compile = "error[E0425]: cannot find value `x`\nerror: could not compile `demo`";
    assert_eq!(
        classify(None, false, false, compile),
        Some(EnvironmentFailure)
    );
    assert_eq!(
        classify(Some("not-found"), false, false, ""),
        Some(EnvironmentFailure)
    );
    assert_eq!(
        classify(None, true, false, "test result: FAILED"),
        Some(EnvironmentFailure)
    );
    assert_eq!(
        classify(
            None,
            false,
            false,
            "warning: spurious network error (2 tries remaining)"
        ),
        Some(EnvironmentFailure)
    );
    // Only a real assertion failure is a test failure.
    let assertion =
        "thread 'adds' panicked at src/lib.rs:9:9\ntest result: FAILED. 0 passed; 1 failed";
    assert_eq!(classify(None, false, false, assertion), Some(TestFailure));
    assert_eq!(
        classify(None, false, false, "exit 1, no output"),
        Some(Unclassified)
    );
    assert_eq!(classify(None, false, true, "test result: ok"), None);

    // Real spawns: a missing binary and a timeout.
    let dir = project();
    let missing = run_check(
        dir.path(),
        0,
        "moai-no-such-binary-2003",
        Duration::from_secs(5),
    );
    assert_eq!(missing.spawn_error.as_deref(), Some("not-found"));
    assert_eq!(missing.failure_class, Some(EnvironmentFailure));
    let slow = run_check(dir.path(), 0, "sleep 5", Duration::from_millis(200));
    assert!(slow.timed_out);
    assert_eq!(slow.failure_class, Some(EnvironmentFailure));

    // RED is confirmed only when every failure is an assertion failure.
    let tdd = resolve_method(&method("tdd", "not-required")).unwrap();
    let mut test_fail = slow.clone();
    test_fail.timed_out = false;
    test_fail.failure_class = Some(TestFailure);
    assert_eq!(
        red_confirmed(&tdd, Some("red"), &[test_fail.clone()]),
        Some(true)
    );
    assert_eq!(
        red_confirmed(&tdd, Some("red"), &[test_fail.clone(), slow.clone()]),
        Some(false)
    );
    assert_eq!(red_confirmed(&tdd, Some("red"), &[missing]), Some(false));
    assert_eq!(
        red_confirmed(&tdd, Some("green"), &[test_fail.clone()]),
        None
    );
    let docs = resolve_method(&method("docs", "not-required")).unwrap();
    assert_eq!(red_confirmed(&docs, Some("schema"), &[test_fail]), None);
}

#[test]
fn oracle_change_guard_keeps_a_pass_from_being_summarized() {
    let paths: Vec<String> = fs::read_to_string(root("tests/fixtures/run/diff_oracle_change.txt"))
        .unwrap()
        .lines()
        .map(String::from)
        .collect();
    assert_eq!(
        oracle_changes(&paths),
        [
            "crates/report/golden/summary.txt",
            "tests/fixtures/run/receipt_self_report.json",
            "tests/run_verify.rs",
            "tools/check_test.py"
        ]
    );
    let dir = project();
    let mut req = request(&["true"], &[0], REV);
    req.diff_paths = paths;
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["receipt"]["checks"][0]["result"], "passed");
    assert_eq!(out["receipt"]["result"], "unknown", "{out}");
    assert_eq!(
        out["receipt"]["oracle_changes"].as_array().unwrap().len(),
        4
    );
    // The same passing run without oracle changes is passed.
    let out = v(run(dir.path(), &request(&["true"], &[0], REV), REV).unwrap());
    assert_eq!(out["receipt"]["result"], "passed");
}

#[test]
fn strategy_matrix_gives_distinct_required_evidence() {
    let tdd = resolve_method(&method("tdd", "not-required")).unwrap();
    let preserve = resolve_method(&method("preserve", "not-required")).unwrap();
    let domain = resolve_method(&method("contract", "required")).unwrap();
    assert_eq!(required_evidence(&tdd), ["red", "green", "regression"]);
    assert_eq!(
        required_evidence(&preserve),
        ["baseline", "after", "intent-diff"]
    );
    assert_eq!(
        required_evidence(&domain),
        [
            "contract-tests",
            "migration",
            "rollback",
            "domain-invariants"
        ]
    );
    // Legacy single-mode input maps to the two fields and is recorded.
    let legacy = resolve_method(&Method {
        change_strategy: None,
        domain_modeling: None,
        legacy_mode: Some("ddd".into()),
    })
    .unwrap();
    assert_eq!(
        v(&legacy),
        json!({"change_strategy": "preserve", "domain_modeling": "required", "legacy_mode": "ddd"})
    );
    let legacy_tdd = resolve_method(&Method {
        change_strategy: None,
        domain_modeling: None,
        legacy_mode: Some("tdd".into()),
    })
    .unwrap();
    assert_eq!(legacy_tdd.change_strategy, "tdd");
    assert_eq!(legacy_tdd.domain_modeling, "not-required");
    // Unknown values and mixed inputs are refused.
    assert!(resolve_method(&method("yolo", "not-required")).is_err());
    assert!(resolve_method(&method("tdd", "maybe")).is_err());
    assert!(resolve_method(&Method {
        change_strategy: Some("tdd".into()),
        domain_modeling: Some("not-required".into()),
        legacy_mode: Some("ddd".into()),
    })
    .is_err());

    // preserve: the after step without a baseline receipt is blocked before anything runs.
    let dir = project();
    let marker = dir.path().join("marker.txt");
    let mut req = request(&["touch marker.txt"], &[0], REV);
    req.method = method("preserve", "not-required");
    req.phase = Some("after".into());
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["decision"], "blocked");
    assert!(out["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("baseline-missing")));
    assert!(!marker.exists());
    // The baseline step itself runs.
    req.phase = Some("baseline".into());
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["decision"], "ready");
    assert!(marker.exists());
    // A phase that is not a step of the strategy is an input error.
    req.phase = Some("red".into());
    assert!(run(dir.path(), &req, REV).is_err());
}

#[test]
fn stale_and_not_run_preserved() {
    let dir = project();
    let req = request(&["true", "true"], &[0], REV);
    let out = run(dir.path(), &req, REV).unwrap();
    let receipt = out.receipt.clone().unwrap();
    let j = v(&receipt);
    assert_eq!(j["checks"][0]["result"], "passed");
    assert_eq!(j["checks"][1]["result"], "not_run");
    assert_eq!(
        j["result"], "not_run",
        "a declared check that did not run is never summarized as passed"
    );

    let spec_digest = receipt.spec_digest.clone();
    let policy = receipt.policy_digest.clone();
    let tree = moai_cli::tree_digest(dir.path(), &[".git", "target"]).unwrap();
    assert_eq!(tree, receipt.tree_digest);
    assert!(stale_reasons(&receipt, &tree, &spec_digest, policy.as_deref()).is_empty());
    fs::write(dir.path().join("README.md"), "changed\n").unwrap();
    let tree2 = moai_cli::tree_digest(dir.path(), &[".git", "target"]).unwrap();
    assert_eq!(
        stale_reasons(&receipt, &tree2, &spec_digest, policy.as_deref()),
        ["tree-changed"]
    );
    assert_eq!(
        stale_reasons(&receipt, &tree, "sha256:other", Some("sha256:other")),
        ["spec-changed", "policy-changed"]
    );
    assert_eq!(
        revalidate(&[receipt], &tree2, &spec_digest, policy.as_deref()),
        ["T1"]
    );
}

#[test]
fn self_report_not_promoted() {
    let claimed: Value = serde_json::from_slice(
        &fs::read(root("tests/fixtures/run/receipt_self_report.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(claimed["producer"]["kind"], "runner-receipt");
    let ingested = ingest_self_report(claimed.clone()).unwrap();
    assert_eq!(v(&ingested)["producer"]["kind"], "self-report");
    assert!(!admissible_evidence(&ingested));

    // A runner receipt produced here is admissible.
    let dir = project();
    let out = run(dir.path(), &request(&["true"], &[0], REV), REV).unwrap();
    let receipt: Receipt = out.receipt.unwrap();
    assert!(admissible_evidence(&receipt));
    assert_eq!(v(&receipt)["producer"]["kind"], "runner-receipt");

    // A baseline handed in on stdin is a self-report and does not unlock the after step.
    let mut req = request(&["true"], &[0], REV);
    req.method = method("preserve", "not-required");
    req.phase = Some("after".into());
    req.baseline = Some(claimed);
    let out = v(run(dir.path(), &req, REV).unwrap());
    assert_eq!(out["decision"], "blocked");
    assert!(out["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("baseline-untrusted")));
}

#[test]
fn command_argv_rules_refuse_shell_shapes() {
    assert_eq!(
        parse_argv("cargo test --offline --locked").unwrap(),
        ["cargo", "test", "--offline", "--locked"]
    );
    for bad in [
        "",
        "   ",
        "CARGO_HOME=/tmp cargo test",
        "--offline cargo",
        "./run.sh",
        "bin/check",
        "scripts\\check.ps1",
    ] {
        let err = parse_argv(bad).unwrap_err();
        assert!(
            !err.contains("CARGO_HOME") && !err.contains("run.sh"),
            "{err}"
        );
    }
    // A command the snapshot cannot run blocks the request before anything starts.
    let dir = project();
    let marker = dir.path().join("marker.txt");
    let out = v(run(
        dir.path(),
        &request(&["touch marker.txt", "./run.sh"], &[0, 1], REV),
        REV,
    )
    .unwrap());
    assert_eq!(out["decision"], "blocked");
    assert!(out["reasons"]
        .as_array()
        .unwrap()
        .contains(&json!("command-not-runnable")));
    assert!(!marker.exists());
}

// ---- real runner on a temporary Cargo project (AC-05C-6) ---------------------------------

fn git(dir: &Path, args: &[&str]) -> bool {
    Command::new("git")
        .arg("-C")
        .arg(dir)
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(["-c", "commit.gpgsign=false"])
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .args(args)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .output()
        .is_ok_and(|o| o.status.success())
}

#[test]
fn real_runner_receipt() {
    let tools = ["git", "cargo"].iter().all(|t| {
        Command::new(t)
            .arg("--version")
            .output()
            .is_ok_and(|o| o.status.success())
    });
    if !tools {
        eprintln!("not_run: git or cargo is not available on this machine");
        return;
    }
    let dir = tempfile::TempDir::new().unwrap();
    fs::create_dir_all(dir.path().join("src")).unwrap();
    fs::copy(
        root("tests/fixtures/run/cargo_project/Cargo.toml.txt"),
        dir.path().join("Cargo.toml"),
    )
    .unwrap();
    fs::copy(
        root("tests/fixtures/run/cargo_project/lib.rs.txt"),
        dir.path().join("src/lib.rs"),
    )
    .unwrap();
    assert!(git(dir.path(), &["init", "-q"]));
    assert!(git(dir.path(), &["add", "."]));
    assert!(git(dir.path(), &["commit", "-q", "-m", "fixture"]));
    let head = moai_cli::plan::current_revision(dir.path()).unwrap();

    let req = request(&["cargo test --offline"], &[0], &head);
    // A separate target directory: the outer `cargo test` holds the lock on its own.
    let target = tempfile::TempDir::new().unwrap();
    let mut child = Command::new(crate::common::MOAI_BIN)
        .arg("run-verify")
        .arg(dir.path())
        .env("CARGO_TARGET_DIR", target.path())
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(&serde_json::to_vec(&req).unwrap())
    {
        Ok(()) => {}
        Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
        Err(e) => panic!("stdin write failed: {e}"),
    }
    let out = child.wait_with_output().unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["ok"], json!(true));
    assert_eq!(env["operation"], "run.verify");
    let r = &env["result"]["receipt"];
    assert_eq!(env["result"]["decision"], "ready", "{env}");
    assert_eq!(r["producer"]["kind"], "runner-receipt");
    assert_eq!(r["base_revision"], json!(head));
    assert_eq!(r["result"], "passed", "{r}");
    let check = &r["checks"][0];
    assert_eq!(check["argv"], json!(["cargo", "test", "--offline"]));
    assert_eq!(check["exit_code"], json!(0));
    assert_eq!(check["timed_out"], json!(false));
    assert_eq!(check["tests_passed"], json!(1));
    assert!(check["stdout_sha256"]
        .as_str()
        .unwrap()
        .starts_with("sha256:"));
    assert!(check["duration_ms"].as_u64().is_some());
    assert!(r["toolchain"]["rustc"]
        .as_str()
        .unwrap()
        .starts_with("rustc "));
    assert!(r["toolchain"]["cargo"]
        .as_str()
        .unwrap()
        .starts_with("cargo "));
    assert!(r["tree_digest"].as_str().unwrap().starts_with("sha256:"));
    // Print the receipt so the run log carries it as evidence.
    println!("{}", serde_json::to_string_pretty(&env).unwrap());
}

/// agent-2 FAIL on 5413e0221: a grandchild holding the pipes kept `run_check` waiting past the
/// timeout. The runner ends the whole process group and bounds the wait for output.
#[cfg(unix)]
#[test]
fn timeout_bounds_runtime_even_with_grandchildren() {
    let dir = project();
    // #2511: the bound proves only that the runner did not wait for the 120 s grandchildren;
    // it is far above the timeout so machine load cannot fail it, and far below their life.
    fs::write(dir.path().join("grand.sh"), "sleep 120 &\nsleep 120\n").unwrap();
    let started = std::time::Instant::now();
    let c = run_check(dir.path(), 0, "sh grand.sh", Duration::from_millis(500));
    let took = started.elapsed();
    assert!(took < Duration::from_secs(60), "took {took:?}");
    assert!(c.timed_out);
    assert_eq!(c.result, CheckResult::Failed);
    assert_eq!(c.failure_class, Some(FailureClass::EnvironmentFailure));
    // The group is gone, so its pipes closed and the output was read to the end.
    assert_eq!(c.output_complete, Some(true));

    // A command that exits but leaves a background child holding its output returns too.
    fs::write(dir.path().join("daemon.sh"), "sleep 120 &\necho started\n").unwrap();
    let started = std::time::Instant::now();
    let c = run_check(dir.path(), 0, "sh daemon.sh", Duration::from_secs(90));
    let took = started.elapsed();
    assert!(took < Duration::from_secs(60), "took {took:?}");
    assert!(!c.timed_out);
    assert_eq!(c.result, CheckResult::Passed);
}

#[test]
fn retry_budget_stops_on_repeated_failure() {
    let dir = project();
    let failing = |attempt: &str| -> Receipt {
        let mut req = request(&["false"], &[0], REV);
        req.attempt_id = attempt.into();
        run(dir.path(), &req, REV).unwrap().receipt.unwrap()
    };
    let attempts = [failing("A1"), failing("A2")];
    assert_eq!(attempts[0].checks[0].result, CheckResult::Failed);
    let budget = Budget { max_attempts: 2 };
    assert!(budget_exhausted(Some(&budget), &attempts));
    assert!(!budget_exhausted(
        Some(&Budget { max_attempts: 3 }),
        &attempts
    ));
    assert!(
        !budget_exhausted(None, &attempts),
        "no budget, no retry decision"
    );
    let required = attempts[0].required_evidence.clone();
    assert_eq!(
        remaining_evidence(&required, &attempts),
        ["red", "green", "regression"]
    );
}

/// Fixture receipt for the cost_report tests: two checks (120 ms passed, 80 ms
/// failed) whose span runs from 130 to 1120.
fn cost_fixture_receipt(
    task_id: &str,
    kind: moai_cli::run::ProducerKind,
) -> moai_cli::run::Receipt {
    let check =
        |command_ref: usize, duration_ms: Option<u64>, started: u64| moai_cli::verify::CheckRun {
            command_ref,
            command: format!("command-{command_ref}"),
            argv: vec!["echo".into(), command_ref.to_string()],
            result: if command_ref == 1 {
                CheckResult::Passed
            } else {
                CheckResult::Failed
            },
            exit_code: Some(0),
            timed_out: false,
            cancelled: false,
            spawn_error: None,
            started_unix_ms: Some(started),
            finished_unix_ms: Some(started + duration_ms.unwrap_or(0)),
            duration_ms,
            stdout_sha256: None,
            stderr_sha256: None,
            failure_class: None,
            test_filters: Vec::new(),
            tests_passed: None,
            tests_failed: None,
            output_complete: Some(true),
        };
    moai_cli::run::Receipt {
        producer: moai_cli::run::Producer { kind },
        task_id: task_id.into(),
        attempt_id: "attempt-1".into(),
        spec_ref: "SPEC-demo".into(),
        spec_digest: "sha256:aa".into(),
        plan_digest: "sha256:bb".into(),
        base_revision: "0".repeat(40),
        tree_digest: "sha256:cc".into(),
        policy_digest: None,
        toolchain: moai_cli::verify::Toolchain {
            rustc: None,
            cargo: None,
        },
        method: moai_cli::run::ResolvedMethod {
            change_strategy: "tdd".into(),
            domain_modeling: "not-required".into(),
            legacy_mode: None,
        },
        phase: None,
        required_evidence: Vec::new(),
        checks: vec![check(1, Some(120), 1_000), check(2, Some(80), 130)],
        oracle_changes: Vec::new(),
        red_confirmed: None,
        result: moai_cli::run::Summary::Passed,
    }
}

fn cost_report_of(receipts: Vec<moai_cli::run::Receipt>) -> serde_json::Value {
    moai_cli::run::cost_report(moai_cli::run::CostRequest {
        protocol_version: 1,
        receipts,
    })
    .unwrap()
}

/// #2191 AC-1/AC-2: cost_report aggregates runner receipts — check outcomes,
/// durations and wall spans carry the measured basis, external backend calls are
/// structurally zero, tokens are absent, and the estimates list is empty (nothing
/// in the receipts is estimated).
#[test]
fn cost_report_aggregates_fixture_receipts() {
    use moai_cli::run::ProducerKind::RunnerReceipt;
    let report = cost_report_of(vec![
        cost_fixture_receipt("task-a", RunnerReceipt),
        cost_fixture_receipt("task-b", RunnerReceipt),
    ]);
    assert_eq!(report["receipts"], json!(2));
    let runner = &report["by_producer"]["runner-receipt"];
    assert_eq!(runner["receipts"], json!(2));
    assert_eq!(
        runner["checks"],
        json!({"total": 4, "passed": 2, "failed": 2, "not_run": 0, "basis": "measured"})
    );
    // Measured duration: 120 + 80 per receipt, aggregated over both.
    assert_eq!(runner["duration_ms"]["total"], json!(400));
    assert_eq!(runner["duration_ms"]["samples_missing"], json!(0));
    assert_eq!(runner["duration_ms"]["basis"], json!("measured"));
    // Wall span per receipt: earliest start (130) to latest finish (1120) = 990;
    // both receipts share the fixture timestamps so p50 is 990.
    assert_eq!(runner["wall_span_ms"]["p50"], json!(990));
    assert_eq!(runner["wall_span_ms"]["basis"], json!("measured"));
    // External backend calls are structurally zero: the runner spawns local
    // commands only. Tokens have no source in a receipt and are absent, not
    // estimated; the estimates list is empty (REQ-D5-01).
    assert_eq!(
        runner["external_backend_calls"],
        json!({"count": 0, "basis": "structural"})
    );
    assert_eq!(report["by_producer"]["self-report"], json!(null));
    assert_eq!(report["tokens"]["basis"], json!("absent"));
    assert_eq!(report["estimates"], json!([]));
    let first = &report["per_receipt"][0];
    assert_eq!(first["task_id"], json!("task-a"));
    assert_eq!(first["duration_ms"], json!(200));
    assert_eq!(first["producer_kind"], json!("runner-receipt"));
    assert_eq!(first["basis"], json!("measured"));
}

/// #2191 audit repair (REQ-D5-01): a self-report's numbers are what the reporter
/// claims, not runner measurements — the same fixture with only the producer kind
/// switched is tagged self_reported, and its external backend call count is absent
/// rather than structurally zero. Nothing lands in the runner group.
#[test]
fn cost_report_tags_self_report_receipts_self_reported() {
    use moai_cli::run::ProducerKind::SelfReport;
    let report = cost_report_of(vec![cost_fixture_receipt("task-s", SelfReport)]);
    // No runner receipts: the runner group is null, not a zero-filled measured row.
    assert_eq!(report["by_producer"]["runner-receipt"], json!(null));
    let own = &report["by_producer"]["self-report"];
    assert_eq!(own["receipts"], json!(1));
    assert_eq!(
        own["checks"],
        json!({"total": 2, "passed": 1, "failed": 1, "not_run": 0, "basis": "self_reported"})
    );
    assert_eq!(own["duration_ms"]["total"], json!(200));
    assert_eq!(own["duration_ms"]["basis"], json!("self_reported"));
    assert_eq!(own["wall_span_ms"]["basis"], json!("self_reported"));
    assert_eq!(
        own["external_backend_calls"],
        json!({"count": null, "basis": "absent"})
    );
    let row = &report["per_receipt"][0];
    assert_eq!(row["producer_kind"], json!("self-report"));
    assert_eq!(row["basis"], json!("self_reported"));
    // No measured or structural basis anywhere for a self-report-only input.
    let text = report.to_string();
    assert!(!text.contains("\"measured\""), "{text}");
    assert!(!text.contains("\"structural\""), "{text}");
}

/// #2191 audit repair: mixed input keeps the two producers apart — runner totals
/// count only runner receipts, and per_receipt rows carry their own basis.
#[test]
fn cost_report_keeps_mixed_producers_apart() {
    use moai_cli::run::ProducerKind::{RunnerReceipt, SelfReport};
    let report = cost_report_of(vec![
        cost_fixture_receipt("task-r", RunnerReceipt),
        cost_fixture_receipt("task-s", SelfReport),
    ]);
    assert_eq!(report["receipts"], json!(2));
    let runner = &report["by_producer"]["runner-receipt"];
    let own = &report["by_producer"]["self-report"];
    assert_eq!(runner["receipts"], json!(1));
    assert_eq!(runner["checks"]["total"], json!(2));
    assert_eq!(runner["duration_ms"]["total"], json!(200));
    assert_eq!(runner["checks"]["basis"], json!("measured"));
    assert_eq!(own["receipts"], json!(1));
    assert_eq!(own["duration_ms"]["total"], json!(200));
    assert_eq!(own["checks"]["basis"], json!("self_reported"));
    assert_eq!(report["per_receipt"][0]["basis"], json!("measured"));
    assert_eq!(report["per_receipt"][1]["basis"], json!("self_reported"));
}
