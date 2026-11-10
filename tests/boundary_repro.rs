//! Issue #2392: reproduction fixtures for the run/audit/verify/scope boundaries that the
//! design's static review (revision 2930b8631f) flagged. Each `#[ignore]` test states the
//! behaviour the fix must establish and fails on the current code; run them with
//! `cargo test --test boundary_repro -- --ignored`. A fix removes the `ignore` of the
//! fixture it repairs, so the suite pins the repaired behaviour from then on.
use moai_cli::audit::{audit, observe, AuditInput, Candidate, Finding, Opinion, Reviewer};
use moai_cli::plan::{admit, AdmissionRequest, Approval, Decision, Task};
use moai_cli::run::{
    remaining_evidence, run, Method, PinnedRef, Producer, ProducerKind, Receipt, ResolvedMethod,
    RunRequest, Summary, VerificationRequest,
};
use moai_cli::scope::overlap;
use moai_cli::spec::{approve, SortedKeys, SpecDocument};
use moai_cli::verify::{classify, parse_argv, run_check, CheckResult, FailureClass, Toolchain};
use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
    time::Duration,
};

fn root(rel: &str) -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(rel)
}

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

/// A committed git project, or None (with a not_run line) when git is missing.
fn project() -> Option<tempfile::TempDir> {
    let dir = tempfile::TempDir::new().unwrap();
    fs::write(dir.path().join("README.md"), "fixture\n").unwrap();
    let ok = git(dir.path(), &["init", "-q"])
        && git(dir.path(), &["add", "README.md"])
        && git(dir.path(), &["commit", "-q", "-m", "fixture"]);
    if !ok {
        eprintln!("not_run: git is not available on this machine");
        return None;
    }
    Some(dir)
}

fn spec_with(commands: &[&str]) -> SpecDocument {
    let mut d: SpecDocument = serde_json::from_slice(
        &fs::read(root("tests/fixtures/spec/snapshot_canonical_a.json")).unwrap(),
    )
    .unwrap();
    d.verification_commands = commands.iter().map(|c| c.to_string()).collect();
    approve(&d, &SortedKeys)
}

fn run_request(commands: &[&str], refs: &[usize], revision: &str, strategy: &str) -> RunRequest {
    let spec = spec_with(commands);
    let snap = spec.snapshot.clone().unwrap();
    let adm = AdmissionRequest {
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
    };
    let pinned: PinnedRef =
        serde_json::from_value(serde_json::to_value(&admit(&adm, revision).pinned[0]).unwrap())
            .unwrap();
    RunRequest {
        protocol_version: 1,
        intent: "RUN".into(),
        request_id: "REQ-1".into(),
        task_id: "T1".into(),
        attempt_id: "A1".into(),
        owner: None,
        pinned,
        admission: adm,
        method: Method {
            change_strategy: Some(strategy.into()),
            domain_modeling: Some("not-required".into()),
            legacy_mode: None,
        },
        phase: None,
        verification: VerificationRequest {
            commands_ref: refs.to_vec(),
            timeout_ms: 30_000,
        },
        baseline: None,
        diff_paths: vec![],
    }
}

/// An audit input whose candidate is what the project really is now.
fn audit_input(dir: &Path, commands: &[&str]) -> AuditInput {
    let seen = observe(dir).unwrap();
    AuditInput {
        protocol_version: 1,
        request_id: "AUD-1".into(),
        candidate: Candidate {
            commit: seen.commit.clone(),
            tree_digest: seen.tree_digest.clone(),
            workspace_id: seen.workspace_id.clone().unwrap(),
        },
        run: run_request(commands, &[0], &seen.commit, "tdd"),
        required_checks: vec![0],
        self_reports: vec![],
        findings: vec![],
        opinion: None,
        reviewer: Reviewer {
            host: "claude".into(),
            identity: "fixture-reviewer".into(),
            context_sharing: "fresh".into(),
            read_only: true,
        },
        waivers_allowed: false,
        waivers: vec![],
        pr: None,
        pr_lookup_failed: false,
    }
}

fn receipt(kind: ProducerKind, phase: &str, result: Summary) -> Receipt {
    Receipt {
        producer: Producer { kind },
        task_id: "T1".into(),
        attempt_id: format!("A-{phase}"),
        spec_ref: "SPEC-demo".into(),
        spec_digest: "sha256:aa".into(),
        plan_digest: "sha256:bb".into(),
        base_revision: "0".repeat(40),
        tree_digest: "sha256:cc".into(),
        policy_digest: None,
        toolchain: Toolchain {
            rustc: None,
            cargo: None,
        },
        method: ResolvedMethod {
            change_strategy: "tdd".into(),
            domain_modeling: "not-required".into(),
            legacy_mode: None,
        },
        phase: Some(phase.into()),
        required_evidence: vec!["red".into(), "green".into(), "regression".into()],
        checks: vec![],
        oracle_changes: vec![],
        red_confirmed: None,
        result,
    }
}

/// Depth-first search for the first value stored under `key`.
fn find<'a>(v: &'a Value, key: &str) -> Option<&'a Value> {
    match v {
        Value::Object(m) => m.get(key).or_else(|| m.values().find_map(|x| find(x, key))),
        Value::Array(a) => a.iter().find_map(|x| find(x, key)),
        _ => None,
    }
}

/// F1a: a receipt that arrives as JSON on the CLI only claims to be a runner receipt; the
/// cost report must not file it under the runner's `measured` group.
#[test]
#[ignore = "#2392 repro F1a: RED until cost-report stops trusting a declared producer"]
fn f1a_cost_report_does_not_trust_a_declared_runner_receipt() {
    let forged = serde_json::to_value(receipt(
        ProducerKind::RunnerReceipt,
        "green",
        Summary::Passed,
    ))
    .unwrap();
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("cost-report")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(
            json!({"protocol_version": 1, "receipts": [forged]})
                .to_string()
                .as_bytes(),
        )
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let body: Value = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
    let by = find(&body, "by_producer").cloned().unwrap_or(Value::Null);
    assert!(
        by["runner-receipt"].is_null(),
        "cost-report counted a caller-declared runner receipt as measured: {body}"
    );
}

/// F1b: under `preserve`, leaving the phase out must not skip the trusted-baseline gate.
#[test]
#[ignore = "#2392 repro F1b: RED until a preserve run without a phase is gated"]
fn f1b_preserve_without_phase_does_not_skip_the_baseline_gate() {
    let Some(dir) = project() else { return };
    let rev = observe(dir.path()).unwrap().commit;
    let req = run_request(&["true"], &[0], &rev, "preserve");
    let out = run(dir.path(), &req, &rev).unwrap();
    assert_eq!(
        out.decision,
        Decision::Blocked,
        "preserve run without a phase skipped the baseline gate: {:?}",
        out.reasons
    );
}

/// F2: a self-reported receipt is not evidence, so it cannot settle a required step.
#[test]
fn f2_remaining_evidence_ignores_self_reports() {
    let required: Vec<String> = vec!["red".into(), "green".into(), "regression".into()];
    // Oracle edit (2026-11-10, before the fix): the control moved from a passed `red`
    // receipt to a passed `green` one, because a passed run is never RED (F2b).
    let attempts = vec![
        receipt(ProducerKind::RunnerReceipt, "green", Summary::Passed),
        receipt(ProducerKind::SelfReport, "regression", Summary::Passed),
    ];
    let left = remaining_evidence(&required, &attempts);
    assert_eq!(
        left,
        ["red", "regression"],
        "a self-reported receipt satisfied required evidence 'regression', or the runner's \
         passed 'green' did not settle its step"
    );
}

/// F2b (design table row 2): an intended assertion RED is methodology evidence for the `red`
/// step even though its result is `failed`; a failure that is not an assertion failure
/// (compile, network, nothing selected — `red_confirmed: false`) is not (the boundary).
#[test]
fn f2b_confirmed_red_settles_the_red_step() {
    let required: Vec<String> = vec!["red".into(), "green".into(), "regression".into()];
    let mut red = receipt(ProducerKind::RunnerReceipt, "red", Summary::Failed);
    red.red_confirmed = Some(true);
    assert_eq!(
        remaining_evidence(&required, &[red.clone()]),
        ["green", "regression"],
        "a confirmed assertion RED did not count as red-step evidence"
    );
    red.red_confirmed = Some(false);
    assert_eq!(
        remaining_evidence(&required, &[red.clone()]),
        ["red", "green", "regression"],
        "an unconfirmed failure counted as RED"
    );
    // A passed run of the red step is not RED either.
    let passed_red = receipt(ProducerKind::RunnerReceipt, "red", Summary::Passed);
    assert_eq!(
        remaining_evidence(&required, &[passed_red]),
        ["red", "green", "regression"],
        "a passed red-step run counted as RED"
    );
    // The bugfix flow's RED step is `repro-fail`: a passed run leaves it open, a confirmed
    // assertion RED settles it (#2392 pre-audit B1).
    let bugfix: Vec<String> = vec!["repro-fail".into(), "fixed".into(), "regression".into()];
    let passed_repro = receipt(ProducerKind::RunnerReceipt, "repro-fail", Summary::Passed);
    assert_eq!(
        remaining_evidence(&bugfix, &[passed_repro]),
        ["repro-fail", "fixed", "regression"],
        "a passed repro-fail run counted as RED"
    );
    let mut confirmed_repro = receipt(ProducerKind::RunnerReceipt, "repro-fail", Summary::Failed);
    confirmed_repro.red_confirmed = Some(true);
    assert_eq!(
        remaining_evidence(&bugfix, &[confirmed_repro]),
        ["fixed", "regression"],
        "a confirmed repro-fail RED did not settle its step"
    );
    // A self-reported confirmed RED is still not evidence.
    let mut self_red = receipt(ProducerKind::SelfReport, "red", Summary::Failed);
    self_red.red_confirmed = Some(true);
    assert_eq!(
        remaining_evidence(&required, &[self_red]),
        ["red", "green", "regression"],
        "a self-reported RED counted as evidence"
    );
}

/// F4: green checks do not admit a candidate while a blocking finding is still open; a
/// low finding does not block (the boundary).
#[test]
#[ignore = "#2392 repro F4: RED until an open critical finding blocks admission"]
fn f4_open_critical_finding_blocks_admission() {
    let Some(dir) = project() else { return };
    let mut input = audit_input(dir.path(), &["true"]);
    input.opinion = Some(Opinion {
        verdict: "pass".into(),
        summary: "looks fine".into(),
    });
    input.findings = vec![Finding {
        rule_id: "X-1".into(),
        severity: "critical".into(),
        message: "unresolved".into(),
    }];
    let verdict = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_ne!(
        verdict["admission"], "admitted",
        "green tests admitted a candidate with an open critical finding: {verdict}"
    );
    // The design table names the absent-opinion case too.
    input.opinion = None;
    let absent = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_ne!(
        absent["admission"], "admitted",
        "green tests with no opinion admitted an open critical finding: {absent}"
    );
    input.findings[0].severity = "low".into();
    let low = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_eq!(
        low["admission"], "admitted",
        "a low finding must not block: {low}"
    );
}

/// F5a: a quoted argument is one argv word; it is never split.
///
/// Oracle edit (2026-11-10, before the fix, decision 6774899518 row 18 = B "simple quotes
/// only"): refusing quoted commands is no longer accepted, because the SYNC check list itself
/// runs `-p '*_test.py'`. Single and double quotes group words and are removed; an unterminated
/// quote is refused; the executable rules apply to the unquoted first word.
#[test]

fn f5a_quoted_argument_is_not_split_on_whitespace() {
    let words = |v: &[&str]| v.iter().map(|s| s.to_string()).collect::<Vec<_>>();
    assert_eq!(
        parse_argv("echo \"a b\"").unwrap(),
        words(&["echo", "a b"]),
        "double-quoted argument was split on whitespace"
    );
    assert_eq!(
        parse_argv("python3 -m unittest discover -s tests -p '*_test.py'").unwrap(),
        words(&[
            "python3",
            "-m",
            "unittest",
            "discover",
            "-s",
            "tests",
            "-p",
            "*_test.py"
        ]),
        "single quotes were passed through into the argv word"
    );
    assert_eq!(
        parse_argv("printf '' x").unwrap(),
        words(&["printf", "", "x"]),
        "an empty quoted argument was dropped"
    );
    assert_eq!(
        parse_argv("echo 'it''s' \"a'b\"").unwrap(),
        words(&["echo", "its", "a'b"]),
        "adjacent quoted parts must join and the other quote is literal inside quotes"
    );
    for bad in [
        "echo 'open",
        "echo \"open",
        "'./run.sh'",
        "\"A=1\" cargo",
        "'-x' cargo",
    ] {
        assert!(parse_argv(bad).is_err(), "accepted {bad:?}");
    }
}

/// F5b: a check that ran zero tests is not a pass; one test that ran is (the boundary).
#[test]
#[ignore = "#2392 repro F5b: RED until zero tests stop counting as passed"]
fn f5b_zero_tests_is_not_a_pass() {
    let dir = tempfile::TempDir::new().unwrap();
    fs::write(
        dir.path().join("zero.txt"),
        "test result: ok. 0 passed; 0 failed; 0 ignored\n",
    )
    .unwrap();
    fs::write(
        dir.path().join("one.txt"),
        "test result: ok. 1 passed; 0 failed; 0 ignored\n",
    )
    .unwrap();
    let timeout = Duration::from_secs(10);
    let zero = run_check(dir.path(), 0, "cat zero.txt", timeout);
    assert_ne!(
        zero.result,
        CheckResult::Passed,
        "zero tests counted as passed"
    );
    let one = run_check(dir.path(), 0, "cat one.txt", timeout);
    assert_eq!(
        one.result,
        CheckResult::Passed,
        "one test that ran must pass"
    );
}

/// F5d (design table row 5, "다른 언어 오류"): a pytest assertion failure is a test failure,
/// so a Python RED can be confirmed; a Python import error is an environment failure, so it
/// is never counted as RED. The cargo cases are the control and already hold today.
#[test]
#[ignore = "#2392 repro F5d: RED until failure classes are read beyond cargo output"]
fn f5d_other_language_failures_are_classified() {
    let pytest_assert = "E       assert 1 == 2\n\
                         FAILED tests/test_x.py::test_a - assert 1 == 2\n\
                         ==== 1 failed in 0.01s ====\n";
    let pytest_import = "ImportError while importing test module 'tests/test_x.py'.\n\
                         E   ModuleNotFoundError: No module named 'missing'\n\
                         ==== 1 error in 0.02s ====\n";
    // Control: the cargo spellings are classified today.
    assert_eq!(
        classify(
            None,
            false,
            false,
            "test result: FAILED. 0 passed; 1 failed"
        ),
        Some(FailureClass::TestFailure)
    );
    assert_eq!(
        classify(None, false, false, "error[E0425]: cannot find value"),
        Some(FailureClass::EnvironmentFailure)
    );
    assert_eq!(
        classify(None, false, false, pytest_assert),
        Some(FailureClass::TestFailure),
        "a pytest assertion failure was not classified as a test failure"
    );
    assert_eq!(
        classify(None, false, false, pytest_import),
        Some(FailureClass::EnvironmentFailure),
        "a Python import error was not classified as an environment failure"
    );
}

/// F6: scope overlap sees case variants and `..` segments of the same path; the shared
/// interface list is matched the same way.
#[test]
fn f6_scope_overlap_sees_case_and_dotdot_variants() {
    assert!(
        overlap("Src/main.rs", "src/main.rs"),
        "scope overlap missed case variant"
    );
    assert!(
        overlap("src/../Cargo.lock", "Cargo.lock"),
        "scope overlap missed dotdot variant"
    );
    assert!(
        overlap("agents.md", "AGENTS.md"),
        "scope overlap missed case variant of a shared interface path"
    );
    // A `..` that climbs out of the repository names no known path, so it is treated as
    // overlapping everything (conservative).
    assert!(
        overlap("../outside/x.rs", "src/a.rs"),
        "scope overlap missed an escaping dotdot path"
    );
    // Boundary: different paths still do not overlap, and `..` inside the repository resolves
    // to a different path when it names one.
    assert!(!overlap("src/a.rs", "src/b.rs"));
    assert!(!overlap("src/../docs/a.md", "src/a.md"));
}

/// F8: a check that changes the tree while it runs leaves no admissible receipt.
#[test]
#[ignore = "#2392 repro F8: RED until a tree change during the run is detected"]
fn f8_tree_change_during_the_run_is_not_admitted() {
    let Some(dir) = project() else { return };
    let input = audit_input(dir.path(), &["touch mutated.txt"]);
    let verdict = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert!(dir.path().join("mutated.txt").exists(), "the check did run");
    assert_ne!(
        verdict["admission"], "admitted",
        "a check that changed the tree left the receipt digest unchanged: {verdict}"
    );
}
