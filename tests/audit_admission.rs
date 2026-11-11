//! Independent AUDIT and completion admission (Issue #2004, 05-D).
use moai_cli::audit::{
    audit, independence, observe, pr_binding, reduce, AuditInput, Candidate, Finding, Opinion,
    PrBinding, Reviewer, Waiver,
};
use moai_cli::plan::{admit, AdmissionRequest, Approval, Decision, Task};
use moai_cli::run::{Method, PinnedRef, RunOutcome, RunRequest, VerificationRequest};
use moai_cli::spec::{approve, SortedKeys, SpecDocument};
use serde_json::{json, Value};
use std::{fs, path::Path, process::Command};

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
fn run_request(commands: &[&str], refs: &[usize], revision: &str) -> RunRequest {
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
    }
}
fn fresh_reviewer() -> Reviewer {
    Reviewer {
        host: "claude".into(),
        identity: "agent-2 subagent".into(),
        context_sharing: "fresh".into(),
        read_only: true,
    }
}
/// An audit input whose candidate is what the project really is now.
fn input(dir: &Path, commands: &[&str], refs: &[usize], required: &[usize]) -> AuditInput {
    let seen = observe(dir).unwrap();
    AuditInput {
        protocol_version: 1,
        request_id: "AUD-1".into(),
        candidate: Candidate {
            commit: seen.commit.clone(),
            tree_digest: seen.tree_digest.clone(),
            workspace_id: seen.workspace_id.clone().unwrap(),
        },
        run: run_request(commands, refs, &seen.commit),
        required_checks: required.to_vec(),
        self_reports: vec![],
        findings: vec![],
        opinion: None,
        reviewer: fresh_reviewer(),
        waivers_allowed: false,
        waivers: vec![],
        pr: None,
        pr_lookup_failed: false,
    }
}
fn v(x: impl serde::Serialize) -> Value {
    serde_json::to_value(x).unwrap()
}
fn reasons(verdict: &Value) -> Vec<String> {
    verdict["reasons"]
        .as_array()
        .unwrap()
        .iter()
        .map(|r| r.as_str().unwrap().to_string())
        .collect()
}
const NOW: &str = "2026-11-08T12:00:00Z";

#[test]
fn rejects_synthetic_and_foreign_evidence() {
    let Some(dir) = project() else { return };
    let inp = input(dir.path(), &["true"], &[0], &[0]);
    let seen = observe(dir.path()).unwrap();
    let passing = moai_cli::run::run(dir.path(), &inp.run, &seen.commit).unwrap();
    let receipt_json = v(passing.receipt.as_ref().unwrap());

    // Only a self-report (whatever producer it claims), no runner receipt: rejected.
    let mut synthetic = inp.clone();
    synthetic.self_reports = vec![receipt_json];
    let none = RunOutcome {
        decision: Decision::Blocked,
        reasons: vec!["command-not-runnable".into()],
        receipt: None,
    };
    let out = v(reduce(&synthetic, &seen, &none, NOW));
    assert_eq!(out["admission"], "rejected", "{out}");
    assert_eq!(out["excluded_self_reports"], json!(["A1"]));

    // A receipt of another tree is stale, excluded, and leaves nothing to pass.
    let mut other_tree = passing.clone();
    other_tree.receipt.as_mut().unwrap().tree_digest = "sha256:other".into();
    let out = v(reduce(&inp, &seen, &other_tree, NOW));
    assert_eq!(out["admission"], "rejected");
    assert_eq!(out["stale_receipts"][0]["reasons"], json!(["tree-changed"]));
    assert_eq!(out["must_pass"][0]["status"], "stale");

    // Work from another worktree.
    let mut foreign = inp.clone();
    foreign.candidate.workspace_id = "sha256:another-worktree".into();
    let out = v(reduce(&foreign, &seen, &passing, NOW));
    assert_eq!(out["admission"], "rejected");
    assert!(reasons(&out).contains(&"foreign-workspace".to_string()));

    // Control: the same receipt, untouched, is admitted.
    assert_eq!(
        v(reduce(&inp, &seen, &passing, NOW))["admission"],
        "admitted"
    );
}

#[test]
fn must_pass_matrix() {
    let Some(dir) = project() else { return };
    let seen = observe(dir.path()).unwrap();

    let passed = v(audit(dir.path(), &input(dir.path(), &["true"], &[0], &[0])).unwrap());
    assert_eq!(passed["admission"], "admitted", "{passed}");
    assert_eq!(passed["status"], "passed");

    let failed = v(audit(dir.path(), &input(dir.path(), &["false"], &[0], &[0])).unwrap());
    assert_eq!(
        (failed["admission"].clone(), failed["status"].clone()),
        (json!("rejected"), json!("failed"))
    );

    let not_run = v(audit(
        dir.path(),
        &input(dir.path(), &["true", "true"], &[0], &[0, 1]),
    )
    .unwrap());
    assert_eq!(
        (not_run["admission"].clone(), not_run["status"].clone()),
        (json!("rejected"), json!("not_run"))
    );
    assert_eq!(not_run["must_pass"][1]["status"], "not_run");

    let inp = input(dir.path(), &["true"], &[0], &[0]);
    let good = moai_cli::run::run(dir.path(), &inp.run, &seen.commit).unwrap();

    let mut no_times = good.clone();
    no_times.receipt.as_mut().unwrap().checks[0].started_unix_ms = None;
    let unknown = v(reduce(&inp, &seen, &no_times, NOW));
    assert_eq!(
        (unknown["admission"].clone(), unknown["status"].clone()),
        (json!("rejected"), json!("unknown"))
    );

    let blocked_run = RunOutcome {
        decision: Decision::Blocked,
        reasons: vec!["baseline-missing".into()],
        receipt: None,
    };
    let blocked = v(reduce(&inp, &seen, &blocked_run, NOW));
    assert_eq!(
        (blocked["admission"].clone(), blocked["status"].clone()),
        (json!("rejected"), json!("blocked"))
    );

    let mut other_spec = good.clone();
    other_spec.receipt.as_mut().unwrap().spec_digest = "sha256:other".into();
    let stale = v(reduce(&inp, &seen, &other_spec, NOW));
    assert_eq!(
        (stale["admission"].clone(), stale["status"].clone()),
        (json!("rejected"), json!("stale"))
    );

    // No averaging: one failing required check rejects a run where the others pass.
    let mixed = v(audit(
        dir.path(),
        &input(
            dir.path(),
            &["true", "false", "true"],
            &[0, 1, 2],
            &[0, 1, 2],
        ),
    )
    .unwrap());
    assert_eq!(mixed["admission"], "rejected");
}

#[test]
fn independence_recorded_and_enforced() {
    let Some(dir) = project() else { return };
    let mut inp = input(dir.path(), &["true"], &[0], &[0]);
    inp.reviewer.context_sharing = "same-session".into();
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(out["admission"], "blocked");
    assert_eq!(out["status"], "blocked");
    assert!(reasons(&out).contains(&"independence-missing".to_string()));
    assert_eq!(out["reviewer"]["context_sharing"], "same-session");

    let out = v(audit(dir.path(), &input(dir.path(), &["true"], &[0], &[0])).unwrap());
    assert_eq!(out["admission"], "admitted");
    assert_eq!(
        out["reviewer"],
        json!({"host": "claude", "identity": "agent-2 subagent", "context_sharing": "fresh", "read_only": true})
    );
    assert_eq!(out["independence"]["accepted"], true);
    assert!(out["independence"]["limitation"].is_null());

    // The table: Codex subagent threads and forks proceed with a recorded limitation.
    let rule = |host: &str, sharing: &str| {
        independence(&Reviewer {
            host: host.into(),
            identity: "x".into(),
            context_sharing: sharing.into(),
            read_only: true,
        })
    };
    assert!(rule("codex", "fresh").accepted);
    assert!(rule("codex", "fresh")
        .limitation
        .unwrap()
        .contains("undocumented"));
    assert!(rule("claude", "fork").accepted && rule("claude", "fork").limitation.is_some());
    assert!(
        rule("codex", "separate-session").accepted
            && rule("codex", "separate-session").limitation.is_none()
    );
    assert!(!rule("claude", "same-session").accepted);
    assert!(!rule("unknown-host", "fresh").accepted);
    assert!(!rule("claude", "telepathy").accepted);
    let writer = independence(&Reviewer {
        host: "claude".into(),
        identity: "x".into(),
        context_sharing: "fresh".into(),
        read_only: false,
    });
    assert!(writer.accepted && writer.limitation.unwrap().contains("write"));
}

#[test]
fn evidence_over_opinion() {
    let Some(dir) = project() else { return };
    let mut inp = input(dir.path(), &["false"], &[0], &[0]);
    inp.opinion = Some(Opinion {
        verdict: "pass".into(),
        summary: "looks done".into(),
    });
    inp.findings = vec![Finding {
        rule_id: "R-1".into(),
        severity: "low".into(),
        message: "naming".into(),
    }];
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(out["admission"], "held", "{out}");
    let rules: Vec<&str> = out["findings"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["rule_id"].as_str().unwrap())
        .collect();
    assert_eq!(rules, ["R-1", "opinion-conflicts-with-evidence"]);

    // The opposite conflict is not admitted either, and its finding is kept.
    let mut inp = input(dir.path(), &["true"], &[0], &[0]);
    inp.opinion = Some(Opinion {
        verdict: "fail".into(),
        summary: "missing case".into(),
    });
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(out["admission"], "held");
    assert_eq!(out["status"], "unknown");
    assert!(out["findings"]
        .as_array()
        .unwrap()
        .iter()
        .any(|f| f["rule_id"] == "opinion-conflicts-with-evidence"));

    // Agreement changes nothing.
    let mut inp = input(dir.path(), &["true"], &[0], &[0]);
    inp.opinion = Some(Opinion {
        verdict: "pass".into(),
        summary: "ok".into(),
    });
    assert_eq!(v(audit(dir.path(), &inp).unwrap())["admission"], "admitted");
}

#[test]
fn waiver_preserves_original() {
    let Some(dir) = project() else { return };
    let waiver = |class: &str, expires: &str| Waiver {
        gate: "check:0".into(),
        gate_class: class.into(),
        reason: "flaky upstream mirror".into(),
        approver: "operator".into(),
        scope: "SPEC-T-SNAP@1".into(),
        expires_at: expires.into(),
    };
    let mut inp = input(dir.path(), &["false"], &[0], &[0]);
    inp.waivers_allowed = true;
    inp.waivers = vec![waiver("quality", "2099-01-01T00:00:00Z")];
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(out["admission"], "admitted", "{out}");
    assert_eq!(
        out["must_pass"][0]["status"], "failed",
        "the original result is kept"
    );
    assert_eq!(out["must_pass"][0]["waived"], true);
    assert_eq!(out["waivers"][0]["outcome"], "waived");
    assert_eq!(out["waivers"][0]["original"], "failed");
    assert_eq!(out["waivers"][0]["expires_at"], "2099-01-01T00:00:00Z");

    for (w, allowed, refusal) in [
        (
            waiver("security", "2099-01-01T00:00:00Z"),
            true,
            "gate-not-waivable",
        ),
        (
            waiver("secret", "2099-01-01T00:00:00Z"),
            true,
            "gate-not-waivable",
        ),
        (
            waiver("permission", "2099-01-01T00:00:00Z"),
            true,
            "gate-not-waivable",
        ),
        (waiver("quality", "2000-01-01T00:00:00Z"), true, "expired"),
        (waiver("quality", "next tuesday"), true, "invalid-expiry"),
        (
            waiver("quality", "2099-01-01T00:00:00Z"),
            false,
            "waivers-not-allowed",
        ),
    ] {
        let mut inp = input(dir.path(), &["false"], &[0], &[0]);
        inp.waivers_allowed = allowed;
        inp.waivers = vec![w];
        let out = v(audit(dir.path(), &inp).unwrap());
        assert_eq!(out["admission"], "rejected", "{refusal}");
        assert_eq!(out["waivers"][0]["outcome"], "refused");
        assert_eq!(out["waivers"][0]["refusal"], refusal);
        assert_eq!(out["must_pass"][0]["waived"], false);
    }
}

#[test]
fn reviewer_edit_new_candidate() {
    let Some(dir) = project() else { return };
    let inp = input(dir.path(), &["true"], &[0], &[0]);
    assert_eq!(v(audit(dir.path(), &inp).unwrap())["admission"], "admitted");
    // The reviewer edits the code: the tree is a new candidate and the old verdict is void.
    fs::write(dir.path().join("README.md"), "reviewer edit\n").unwrap();
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(out["admission"], "rejected");
    assert_eq!(out["status"], "stale");
    assert!(reasons(&out).contains(&"candidate-changed".to_string()));
}

#[test]
fn pr_head_binding() {
    let read = |name: &str| -> Value {
        serde_json::from_slice(
            &fs::read(root(&format!("tests/fixtures/audit/{name}.json"))).unwrap(),
        )
        .unwrap()
    };
    let commit = "5ad9cf3154dbd6f010d2bc136885276592724ce9";
    assert_eq!(
        pr_binding(commit, Some(&read("pr_head_match")), false),
        PrBinding::Ok
    );
    assert_eq!(
        pr_binding(commit, Some(&read("pr_head_mismatch")), false),
        PrBinding::Stale
    );
    assert_eq!(pr_binding(commit, None, true), PrBinding::Unknown);
    assert_eq!(
        pr_binding(commit, Some(&json!({"head": {}})), false),
        PrBinding::Unknown
    );
    assert_eq!(pr_binding(commit, None, false), PrBinding::NotApplicable);

    let Some(dir) = project() else { return };
    let mut inp = input(dir.path(), &["true"], &[0], &[0]);
    inp.pr = Some(read("pr_head_mismatch"));
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(
        (out["admission"].clone(), out["status"].clone()),
        (json!("rejected"), json!("stale"))
    );
    assert_eq!(out["pr_binding"], "stale");
    let mut inp = input(dir.path(), &["true"], &[0], &[0]);
    inp.pr_lookup_failed = true;
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(
        (out["admission"].clone(), out["status"].clone()),
        (json!("held"), json!("unknown"))
    );
    let mut inp = input(dir.path(), &["true"], &[0], &[0]);
    let head = inp.candidate.commit.clone();
    inp.pr = Some(json!({"head": {"sha": head}}));
    let out = v(audit(dir.path(), &inp).unwrap());
    assert_eq!(out["admission"], "admitted");
    assert_eq!(out["pr_binding"], "ok");
}
