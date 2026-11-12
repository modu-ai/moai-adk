//! Issue #2392: reproduction fixtures for the run/audit/verify/scope boundaries that the
//! design's static review (revision 2930b8631f) flagged. Each `#[ignore]` test states the
//! behaviour the fix must establish and fails on the current code; run them with
//! `cargo test --test boundary_repro -- --ignored`. A fix removes the `ignore` of the
//! fixture it repairs, so the suite pins the repaired behaviour from then on.
use moai_cli::audit::{audit, observe, AuditInput, Candidate, Finding, Opinion, Reviewer};
use moai_cli::plan::{admit, AdmissionRequest, Approval, Decision, Task};
use moai_cli::run::{
    remaining_evidence, run, run_verify, BaselineRef, Method, PinnedRef, Producer, ProducerKind,
    Receipt, ResolvedMethod, RunRequest, Summary, VerificationRequest,
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
        baseline_ref: None,
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
    // Oracle edit (2026-11-10, before the fix, decision 6774899518 row 16 = B): the
    // aggregation stays as #2191 accepted it (the declared runner group is kept), and the
    // output says that the producer kind was declared by the caller, at the top and per row.
    let by = find(&body, "by_producer").cloned().unwrap_or(Value::Null);
    assert_eq!(by["runner-receipt"]["receipts"], json!(1), "{body}");
    assert_eq!(
        body["result"]["producer_declared_by"],
        json!("caller"),
        "cost-report did not say the producer kind is caller-declared: {body}"
    );
    assert_eq!(
        body["result"]["per_receipt"][0]["producer_declared_by"],
        json!("caller"),
        "{body}"
    );
}

/// F1b: under `preserve`, a run without a phase is the baseline step; it can never stand in
/// for a later step, so it cannot skip the trusted-baseline gate.
///
/// Oracle edit (2026-11-10, before the fix, decision 6774899518 row 13, F1b = B "a missing
/// phase is the baseline"): blocking the run is no longer the expectation; it runs, its
/// receipt is the `baseline` step, and that receipt settles `baseline` only.
#[test]
fn f1b_preserve_without_phase_does_not_skip_the_baseline_gate() {
    let Some(dir) = project() else { return };
    let rev = observe(dir.path()).unwrap().commit;
    let req = run_request(&["true"], &[0], &rev, "preserve");
    let out = run(dir.path(), &req, &rev).unwrap();
    assert_eq!(out.decision, Decision::Ready, "{:?}", out.reasons);
    let receipt = out.receipt.unwrap();
    assert_eq!(receipt.phase.as_deref(), Some("baseline"));
    let required: Vec<String> = vec!["baseline".into(), "after".into(), "intent-diff".into()];
    assert_eq!(
        remaining_evidence(&required, &[receipt]),
        ["after", "intent-diff"],
        "a phase-less preserve run settled more than the baseline step"
    );
}

/// A preserve request for one step, with an optional stored-baseline reference.
fn preserve_request(
    commands: &[&str],
    rev: &str,
    phase: &str,
    attempt: &str,
    baseline_ref: Option<BaselineRef>,
) -> RunRequest {
    let mut req = run_request(commands, &[0], rev, "preserve");
    req.phase = Some(phase.into());
    req.attempt_id = attempt.into();
    req.baseline_ref = baseline_ref;
    req
}

/// F1c / D1 (decision 6774899518 row 13 = A): `run-verify` stores the baseline receipt it
/// produced under the git common dir and reports its digest; a later step names it by attempt
/// id and digest, and the runner reads the file again, checks the digest and what the receipt
/// says (runner, preserve, baseline, passed, same task and spec). Nothing else unlocks it.
#[test]
fn f1c_stored_baseline_unlocks_after_and_nothing_else_does() {
    let Some(dir) = project() else { return };
    let p = dir.path();
    let seen = observe(p).unwrap();
    let rev = seen.commit.clone();
    let store = p.join(".git/.moai-cli/receipts/run");

    // The baseline run stores its receipt outside the tree: the candidate digest is unchanged.
    let base = run_verify(
        p,
        &preserve_request(&["true"], &rev, "baseline", "B1", None),
        &rev,
    )
    .unwrap();
    let stored = serde_json::to_value(&base).unwrap();
    assert_eq!(stored["baseline_store"]["stored"], json!(true), "{stored}");
    let digest = stored["baseline_store"]["digest"]
        .as_str()
        .unwrap()
        .to_string();
    let file = store.join("B1.json");
    assert_eq!(moai_cli::digest(&fs::read(&file).unwrap()), digest);
    assert_eq!(observe(p).unwrap().tree_digest, seen.tree_digest);

    let good = BaselineRef {
        attempt_id: "B1".into(),
        digest: digest.clone(),
    };
    let after = run(
        p,
        &preserve_request(&["true"], &rev, "after", "A2", Some(good.clone())),
        &rev,
    )
    .unwrap();
    assert_eq!(after.decision, Decision::Ready, "{:?}", after.reasons);

    // A wrong digest, a missing file and an unsafe attempt id are refused.
    for (bad, reason) in [
        (
            BaselineRef {
                attempt_id: "B1".into(),
                digest: format!("sha256:{}", "0".repeat(64)),
            },
            "baseline-digest-mismatch",
        ),
        (
            BaselineRef {
                attempt_id: "B9".into(),
                digest: digest.clone(),
            },
            "baseline-missing",
        ),
        (
            BaselineRef {
                attempt_id: "../B1".into(),
                digest: digest.clone(),
            },
            "baseline-untrusted",
        ),
    ] {
        let out = run(
            p,
            &preserve_request(&["true"], &rev, "after", "A3", Some(bad)),
            &rev,
        )
        .unwrap();
        assert_eq!(out.decision, Decision::Blocked);
        assert_eq!(out.reasons, [reason], "{:?}", out.reasons);
    }

    // A stored baseline that failed binds by digest but not by content (its spec differs too).
    let failed = serde_json::to_value(
        run_verify(
            p,
            &preserve_request(&["false"], &rev, "baseline", "B2", None),
            &rev,
        )
        .unwrap(),
    )
    .unwrap();
    let failed_ref = BaselineRef {
        attempt_id: "B2".into(),
        digest: failed["baseline_store"]["digest"].as_str().unwrap().into(),
    };
    let out = run(
        p,
        &preserve_request(&["true"], &rev, "after", "A4", Some(failed_ref)),
        &rev,
    )
    .unwrap();
    assert_eq!(
        out.reasons,
        ["baseline-untrusted"],
        "a failed baseline unlocked after"
    );

    // A file rewritten in place with a matching reference must still say what a runner
    // baseline says: a self-report with its digest is refused.
    // Each bound field on its own: the file is rewritten as `F<n>.json` with the attempt id
    // set to that name (so only the edited field differs) and referenced by its own digest.
    let original: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
    // `result` gets its own row: the failed run above also differs in its spec digest, so it
    // alone did not pin the passed requirement (mutation D1-passed survived).
    let edits: [(&str, Value); 8] = [
        ("result", json!("failed")),
        ("producer", json!({"kind": "self-report"})),
        ("task_id", json!("T2")),
        ("spec_digest", json!(format!("sha256:{}", "1".repeat(64)))),
        ("plan_digest", json!(format!("sha256:{}", "2".repeat(64)))),
        ("phase", json!("after")),
        ("attempt_id", json!("other")),
        (
            "method",
            json!({"change_strategy": "tdd", "domain_modeling": "not-required", "legacy_mode": null}),
        ),
    ];
    // Control: the same rewrite with no field edited unlocks, so each refusal below is the
    // edited field's doing.
    for (n, edit) in std::iter::once(None)
        .chain(edits.iter().map(Some))
        .enumerate()
    {
        let name = format!("F{n}");
        let mut forged = original.clone();
        forged["attempt_id"] = json!(name);
        if let Some((field, value)) = edit {
            forged[*field] = value.clone();
        }
        let bytes = serde_json::to_vec(&forged).unwrap();
        fs::write(store.join(format!("{name}.json")), &bytes).unwrap();
        let forged_ref = BaselineRef {
            attempt_id: name.clone(),
            digest: moai_cli::digest(&bytes),
        };
        let out = run(
            p,
            &preserve_request(&["true"], &rev, "after", "A5", Some(forged_ref)),
            &rev,
        )
        .unwrap();
        match edit {
            None => assert_eq!(out.decision, Decision::Ready, "control: {:?}", out.reasons),
            Some((field, _)) => assert_eq!(out.reasons, ["baseline-untrusted"], "{field}"),
        }
    }

    // A second baseline under the same attempt id does not replace the first.
    let again = serde_json::to_value(
        run_verify(
            p,
            &preserve_request(&["false"], &rev, "baseline", "B1", None),
            &rev,
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(again["baseline_store"]["stored"], json!(false), "{again}");
    assert_eq!(moai_cli::digest(&fs::read(&file).unwrap()), digest);

    // An unsafe attempt id is not written anywhere.
    let unsafe_id = serde_json::to_value(
        run_verify(
            p,
            &preserve_request(&["true"], &rev, "baseline", "../x", None),
            &rev,
        )
        .unwrap(),
    )
    .unwrap();
    assert_eq!(
        unsafe_id["baseline_store"]["stored"],
        json!(false),
        "{unsafe_id}"
    );
    assert!(!store.parent().unwrap().join("x").exists());

    // A step that is not the baseline, and a non-preserve run, store nothing.
    let none = serde_json::to_value(
        run_verify(
            p,
            &preserve_request(&["true"], &rev, "after", "A6", Some(good)),
            &rev,
        )
        .unwrap(),
    )
    .unwrap();
    assert!(none["baseline_store"].is_null(), "{none}");
    assert!(!store.join("A6.json").exists());
}

/// D1: the auditor's run writes nothing, even for a phase-less `preserve` request (which is a
/// baseline step after F1b).
#[test]
fn d1_audit_writes_no_baseline_receipt() {
    let Some(dir) = project() else { return };
    let mut input = audit_input(dir.path(), &["true"]);
    input.run.method.change_strategy = Some("preserve".into());
    let _ = audit(dir.path(), &input).unwrap();
    assert!(
        !dir.path().join(".git/.moai-cli").exists(),
        "audit-reduce wrote under the git common dir"
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
    // Oracle edit (2026-11-10, before the fix, decision 6774899518 row 17 = A): the result is
    // exactly `held` with reason `blocking-finding-open` for critical, high and any unknown
    // severity; medium, low and info do not block (case-insensitive).
    for blocking in ["critical", "high", "HIGH", "major", ""] {
        input.findings[0].severity = blocking.into();
        let out = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
        assert_eq!(out["admission"], "held", "{blocking:?}: {out}");
        assert_eq!(out["reasons"], json!(["blocking-finding-open"]), "{out}");
    }
    for open in ["medium", "Medium", "low", "info"] {
        input.findings[0].severity = open.into();
        let out = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
        assert_eq!(
            out["admission"], "admitted",
            "{open:?} must not block: {out}"
        );
    }
    // A failing required check still rejects first.
    let mut failing = audit_input(dir.path(), &["false"]);
    failing.findings = vec![Finding {
        rule_id: "X-1".into(),
        severity: "critical".into(),
        message: "unresolved".into(),
    }];
    let out = serde_json::to_value(audit(dir.path(), &failing).unwrap()).unwrap();
    assert_eq!(out["admission"], "rejected", "{out}");
    assert_eq!(out["reasons"], json!(["must-pass-not-met"]), "{out}");
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
    // Oracle edit (2026-11-10, before the fix, decision 6774899518 row 19 = A): zero is
    // exactly `not_run`, and only when the output carries test counts; a command with no test
    // output keeps its exit-code result, and ignored tests are not tests that ran.
    fs::write(
        dir.path().join("ignored.txt"),
        "test result: ok. 0 passed; 0 failed; 3 ignored\n",
    )
    .unwrap();
    fs::write(
        dir.path().join("split.txt"),
        "test result: ok. 0 passed; 0 failed; 0 ignored\n\
         test result: ok. 2 passed; 0 failed; 0 ignored\n",
    )
    .unwrap();
    let timeout = Duration::from_secs(10);
    let zero = run_check(dir.path(), 0, "cat zero.txt", timeout);
    assert_eq!(
        zero.result,
        CheckResult::NotRun,
        "zero tests must be not_run"
    );
    let ignored = run_check(dir.path(), 0, "cat ignored.txt", timeout);
    assert_eq!(
        ignored.result,
        CheckResult::NotRun,
        "only ignored tests must be not_run"
    );
    let one = run_check(dir.path(), 0, "cat one.txt", timeout);
    assert_eq!(
        one.result,
        CheckResult::Passed,
        "one test that ran must pass"
    );
    let split = run_check(dir.path(), 0, "cat split.txt", timeout);
    assert_eq!(
        split.result,
        CheckResult::Passed,
        "an empty target beside a target that ran tests must pass"
    );
    let plain = run_check(dir.path(), 0, "true", timeout);
    assert_eq!(
        plain.result,
        CheckResult::Passed,
        "a command without test output keeps its exit-code result"
    );
}

/// F5d (design table row 5, "다른 언어 오류"): a pytest assertion failure is a test failure,
/// so a Python RED can be confirmed; a Python import error is an environment failure, so it
/// is never counted as RED. The cargo cases are the control and already hold today.
#[test]
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
    // Decision 6774899518 row 20 = A: unittest too, and the environment still wins when both
    // signs appear in one run.
    let unittest_assert = "FAIL: test_a (test_x.T.test_a)\n\
                           Traceback (most recent call last):\n\
                           AssertionError: 1 != 2\n\
                           FAILED (failures=1)\n";
    let unittest_import = "ERROR: test_x (unittest.loader._FailedTest.test_x)\n\
                           ImportError: Failed to import test module: test_x\n\
                           ModuleNotFoundError: No module named 'missing'\n\
                           FAILED (errors=1)\n";
    assert_eq!(
        classify(None, false, false, unittest_assert),
        Some(FailureClass::TestFailure),
        "a unittest assertion failure was not classified as a test failure"
    );
    assert_eq!(
        classify(None, false, false, unittest_import),
        Some(FailureClass::EnvironmentFailure),
        "a unittest import failure was not classified as an environment failure"
    );
    let both = format!("{pytest_assert}{pytest_import}");
    assert_eq!(
        classify(None, false, false, &both),
        Some(FailureClass::EnvironmentFailure),
        "an import error beside an assertion failure must not count as RED"
    );
    // Each environment sign on its own (mutation M5d-1 survived while both appeared together).
    for alone in [
        "ImportError: cannot import name 'x' from 'y'\n",
        "ModuleNotFoundError: No module named 'missing'\n",
    ] {
        assert_eq!(
            classify(None, false, false, alone),
            Some(FailureClass::EnvironmentFailure),
            "{alone:?}"
        );
    }
    // Boundary: a Python failure with neither sign stays unclassified (never RED).
    assert_eq!(
        classify(
            None,
            false,
            false,
            "Traceback (most recent call last):\nKeyError: 'x'\n"
        ),
        Some(FailureClass::Unclassified)
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
/// The verdict is judged on the project as it is after the checks ran, so the change is the
/// same `candidate-changed` rejection as a change made before the audit; a check that leaves
/// the tree alone is still admitted.
#[test]
fn f8_tree_change_during_the_run_is_not_admitted() {
    let Some(dir) = project() else { return };
    let input = audit_input(dir.path(), &["touch mutated.txt"]);
    let verdict = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert!(dir.path().join("mutated.txt").exists(), "the check did run");
    assert_ne!(
        verdict["admission"], "admitted",
        "a check that changed the tree left the receipt digest unchanged: {verdict}"
    );
    assert_eq!(verdict["admission"], "rejected", "{verdict}");
    assert_eq!(
        verdict["reasons"],
        json!(["candidate-changed"]),
        "{verdict}"
    );

    // An edit to a tracked file during the run counts too, not only a new file.
    let Some(dir) = project() else { return };
    let input = audit_input(dir.path(), &["sh -c 'echo changed >> README.md'"]);
    let verdict = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_eq!(
        verdict["reasons"],
        json!(["candidate-changed"]),
        "{verdict}"
    );

    // Boundary: a check that changes nothing is admitted.
    let Some(dir) = project() else { return };
    let input = audit_input(dir.path(), &["true"]);
    let verdict = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_eq!(verdict["admission"], "admitted", "{verdict}");
}

/// D2 stage 1 (decision 6774899518 row 14 = B, staged): the verdict says on what basis the
/// reviewer's independence was judged. Today the reviewer fields are the caller's own words,
/// so the basis is `self-declared`; the acceptance table itself is unchanged. Stage 2 derives
/// author and reviewer from the commit's claim trailer and the host session id.
#[test]
fn d2_independence_basis_is_reported_as_self_declared() {
    let Some(dir) = project() else { return };
    let input = audit_input(dir.path(), &["true"]);
    let out = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_eq!(out["independence_basis"], json!("self-declared"), "{out}");
    assert_eq!(out["independence"]["accepted"], json!(true), "{out}");
    assert_eq!(out["admission"], "admitted", "{out}");
}

/// D3-B (decision 6774899518 row 15 = A, pin only): a review export the reviewer writes into
/// PROJECT — here a decision record under `.moai-cli/` — is a change to the candidate, so the
/// verdict is `rejected` with `candidate-changed`, never an admission of the unchanged commit.
#[test]
fn d3b_review_export_into_project_is_a_new_candidate() {
    let Some(dir) = project() else { return };
    let input = audit_input(dir.path(), &["true"]);
    let export = dir.path().join(".moai-cli/decisions");
    fs::create_dir_all(&export).unwrap();
    fs::write(export.join("review.jsonl"), "{\"verdict\":\"pass\"}\n").unwrap();
    let out = serde_json::to_value(audit(dir.path(), &input).unwrap()).unwrap();
    assert_eq!(out["admission"], "rejected", "{out}");
    assert_eq!(out["reasons"], json!(["candidate-changed"]), "{out}");
}

/// Pre-audit 6777054180 B1/B2: the stored-baseline file name rule at its edges, with literal
/// values. 128 characters are stored and 129 are not; the empty id and an id starting with a
/// dot (which could collide with the store's `.<id>.<pid>.tmp` files) are refused. A later
/// step naming such an id is `baseline-untrusted`.
#[test]
fn d1_attempt_id_file_name_edges() {
    let Some(dir) = project() else { return };
    let p = dir.path();
    let rev = observe(p).unwrap().commit;
    let store = |id: &str| {
        serde_json::to_value(
            run_verify(
                p,
                &preserve_request(&["true"], &rev, "baseline", id, None),
                &rev,
            )
            .unwrap(),
        )
        .unwrap()["baseline_store"]
            .clone()
    };
    let id128 = "a".repeat(128);
    let id129 = "b".repeat(129);
    assert_eq!(id128.len(), 128);
    assert_eq!(id129.len(), 129);
    let ok = store(&id128);
    assert_eq!(ok["stored"], json!(true), "{ok}");
    for refused in [id129.as_str(), "", ".x"] {
        let out = store(refused);
        assert_eq!(out["stored"], json!(false), "{refused:?}: {out}");
        assert_eq!(
            out["reason"],
            json!("attempt-id-not-a-safe-file-name"),
            "{refused:?}: {out}"
        );
    }
    // The 128-character baseline unlocks; refused names never reach the file system.
    let digest = ok["digest"].as_str().unwrap().to_string();
    let after = |id: &str| {
        run(
            p,
            &preserve_request(
                &["true"],
                &rev,
                "after",
                "A1",
                Some(BaselineRef {
                    attempt_id: id.into(),
                    digest: digest.clone(),
                }),
            ),
            &rev,
        )
        .unwrap()
    };
    assert_eq!(after(&id128).decision, Decision::Ready);
    for refused in [id129.as_str(), "", ".x"] {
        assert_eq!(
            after(refused).reasons,
            ["baseline-untrusted"],
            "{refused:?}"
        );
    }
}

/// Pre-audit 6777054180 N1: the pytest `E` detail line needs the space after `E`, and a
/// severity is compared after trimming.
#[test]
fn n1_pytest_e_line_and_trimmed_severity() {
    assert_eq!(
        classify(None, false, false, "E   assert x\n"),
        Some(FailureClass::TestFailure)
    );
    assert_eq!(
        classify(None, false, false, "eassert x\n"),
        Some(FailureClass::Unclassified)
    );
    assert!(!moai_cli::audit::blocking_severity(" medium "));
    assert!(moai_cli::audit::blocking_severity(" high "));
}
