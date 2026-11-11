//! PLAN → RUN → VERIFY → AUDIT end to end on one fixture SPEC (Issue #2004, AC-05D-8):
//! 05-A parses and approves, 05-B admits, 05-C runs, 05-D audits through the CLI.
use moai_cli::audit::{observe, AuditInput, Candidate, Reviewer};
use moai_cli::plan::{admit, AdmissionRequest, Approval, Task};
use moai_cli::run::{Method, PinnedRef, RunRequest, VerificationRequest};
use moai_cli::spec::{approve, parse_gears, Acceptance, SortedKeys, SpecDocument, Verification};
use serde_json::Value;
use std::{
    fs,
    io::Write,
    path::Path,
    process::{Command, Stdio},
};

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

/// 05-A: one GEARS requirement, one criterion, two verification commands; approved.
fn approved_spec() -> SpecDocument {
    let mut req = parse_gears(
        "REQ-E2E-001",
        "When an audit runs, the audit lane shall admit only fresh runner receipts.",
    )
    .unwrap();
    req.observable_result = Some("admission is admitted only with fresh runner receipts".into());
    req.acceptance_refs = vec!["AC-E2E-1".into()];
    let doc = SpecDocument {
        spec_id: "SPEC-E2E".into(),
        revision: 1,
        status: "proposed".into(),
        syntax: "GEARS".into(),
        source: None,
        requirements: vec![req],
        acceptance: vec![Acceptance {
            id: "AC-E2E-1".into(),
            observable_result: "both checks pass".into(),
            verification: Verification {
                kind: "command".into(),
                selector: "true".into(),
            },
        }],
        scope: vec!["README.md".into()],
        policy_revision: Some("rc1-r1".into()),
        verification_commands: vec!["true".into(), "true".into()],
        unresolved: vec![],
        snapshot: None,
    };
    let v = moai_cli::spec::validate(&doc);
    assert_eq!(serde_json::to_value(&v).unwrap()["structure"], "passed");
    approve(&doc, &SortedKeys)
}

fn audit_input(dir: &Path, refs: &[usize]) -> AuditInput {
    let seen = observe(dir).unwrap();
    let spec = approved_spec();
    let snap = spec.snapshot.clone().unwrap();
    // 05-B: the plan admits the single task at the current commit.
    let adm = AdmissionRequest {
        protocol_version: 1,
        spec,
        tasks: vec![Task {
            id: "T-E2E".into(),
            depends_on: vec![],
            write_paths: vec!["README.md".into()],
            verifier_refs: vec!["AC-E2E-1".into()],
            host: None,
        }],
        approval: Approval {
            id: "APR-E2E".into(),
            approver: "operator".into(),
            basis: "plan review".into(),
            snapshot_reference: snap.reference,
            snapshot_digest: snap.digest,
            allowed_actions: vec!["run".into()],
            revoked: false,
        },
        expected_revision: seen.commit.clone(),
        capabilities: vec![],
    };
    let admitted = admit(&adm, &seen.commit);
    assert_eq!(
        serde_json::to_value(&admitted).unwrap()["decision"],
        "ready"
    );
    let pinned: PinnedRef =
        serde_json::from_value(serde_json::to_value(&admitted.pinned[0]).unwrap()).unwrap();
    AuditInput {
        protocol_version: 1,
        request_id: "AUD-E2E".into(),
        candidate: Candidate {
            commit: seen.commit.clone(),
            tree_digest: seen.tree_digest.clone(),
            workspace_id: seen.workspace_id.clone().unwrap(),
        },
        // 05-C: the auditor re-runs these commands itself.
        run: RunRequest {
            protocol_version: 1,
            intent: "RUN".into(),
            request_id: "RUN-E2E".into(),
            task_id: "T-E2E".into(),
            attempt_id: "A-E2E".into(),
            owner: None,
            pinned,
            admission: adm,
            method: Method {
                change_strategy: Some("tdd".into()),
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
        },
        required_checks: vec![0, 1],
        self_reports: vec![],
        findings: vec![],
        opinion: None,
        reviewer: Reviewer {
            host: "claude".into(),
            identity: "e2e reviewer".into(),
            context_sharing: "fresh".into(),
            read_only: true,
        },
        waivers_allowed: false,
        waivers: vec![],
        pr: None,
        pr_lookup_failed: false,
    }
}

/// 05-D through the CLI: `moai-cli audit-reduce PROJECT`.
fn audit_reduce(dir: &Path, input: &AuditInput) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("audit-reduce")
        .arg(dir)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    match child
        .stdin
        .take()
        .unwrap()
        .write_all(&serde_json::to_vec(input).unwrap())
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
    serde_json::from_slice(&out.stdout).unwrap()
}

#[test]
fn plan_run_verify_audit_admits_once_and_rejects_a_not_run_check() {
    let dir = tempfile::TempDir::new().unwrap();
    fs::write(dir.path().join("README.md"), "e2e\n").unwrap();
    let ok = git(dir.path(), &["init", "-q"])
        && git(dir.path(), &["add", "README.md"])
        && git(dir.path(), &["commit", "-q", "-m", "e2e"]);
    if !ok {
        eprintln!("not_run: git is not available on this machine");
        return;
    }
    let admitted = audit_reduce(dir.path(), &audit_input(dir.path(), &[0, 1]));
    assert_eq!(admitted["ok"], true);
    assert_eq!(admitted["operation"], "audit.reduce");
    let v = &admitted["result"];
    assert_eq!(v["admission"], "admitted", "{admitted}");
    assert_eq!(v["status"], "passed");
    assert_eq!(v["receipt"]["producer"]["kind"], "runner-receipt");
    println!("admitted: {}", serde_json::to_string(&admitted).unwrap());

    // The same fixture with one required check left out.
    let rejected = audit_reduce(dir.path(), &audit_input(dir.path(), &[0]));
    let v = &rejected["result"];
    assert_eq!(v["admission"], "rejected", "{rejected}");
    assert_eq!(v["status"], "not_run");
    assert_eq!(v["must_pass"][1]["status"], "not_run");
    println!("rejected: {}", serde_json::to_string(&rejected).unwrap());
}
