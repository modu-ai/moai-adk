//! RUN methodology and real VERIFY (Issue #2003, 05-C).
//!
//! A RUN request names a task admitted by 05-B, a change strategy and a domain-modelling
//! choice, and which of the approved snapshot's verification commands to run. `run` re-reads
//! the pinned contract first and runs nothing on a mismatch; otherwise it runs the commands
//! through the trusted runner (`crate::verify`) and returns a receipt. Receipts that arrive
//! as JSON from outside are self-reports and are never treated as runner receipts.

use crate::digest;
use crate::plan::{admit, AdmissionRequest, Decision};
use crate::verify::{self, CheckResult, CheckRun, FailureClass, Toolchain};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::path::Path;
use std::time::Duration;

pub type Result<T> = std::result::Result<T, String>;

/// Paths a tree digest leaves out: git metadata and build output (#2003 leader decision H1).
pub const TREE_EXCLUDES: [&str; 2] = [".git", "target"];

/// Required evidence per change strategy; the table, not the request, decides it.
const STRATEGIES: [(&str, &[&str]); 6] = [
    ("tdd", &["red", "green", "regression"]),
    ("bugfix", &["repro-fail", "fixed", "regression"]),
    ("preserve", &["baseline", "after", "intent-diff"]),
    ("contract", &["contract-tests", "migration", "rollback"]),
    ("eval", &["trials"]),
    ("docs", &["schema", "links", "smoke"]),
];
const DOMAIN_EVIDENCE: &str = "domain-invariants";

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Method {
    pub change_strategy: Option<String>,
    pub domain_modeling: Option<String>,
    /// The older single choice, `tdd` or `ddd`; mapped to the two fields above.
    pub legacy_mode: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct ResolvedMethod {
    pub change_strategy: String,
    /// `not-required` or `required`.
    pub domain_modeling: String,
    /// Set when the request used the legacy single choice.
    pub legacy_mode: Option<String>,
}

/// Mirror of 05-B `plan::PinnedContract`, which has no `Deserialize`; compared field by field
/// with what `plan::admit` returns now.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct PinnedRef {
    pub spec_ref: String,
    pub plan_digest: String,
    pub task_id: String,
    pub expected_revision: String,
    pub scope: Vec<String>,
    pub approval_ref: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct VerificationRequest {
    /// Indexes into the approved snapshot's `verification_commands`.
    pub commands_ref: Vec<usize>,
    /// Per command; required, no default (P11).
    pub timeout_ms: u64,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RunRequest {
    pub protocol_version: u32,
    /// Always `RUN`.
    pub intent: String,
    pub request_id: String,
    pub task_id: String,
    pub attempt_id: String,
    /// Opaque; owner and epoch belong to 06-A.
    pub owner: Option<String>,
    pub pinned: PinnedRef,
    pub admission: AdmissionRequest,
    pub method: Method,
    /// A step of the change strategy (`red`, `baseline`, …), when the run records one.
    pub phase: Option<String>,
    pub verification: VerificationRequest,
    /// A baseline receipt for `preserve`; arrives as JSON, so it is a self-report.
    pub baseline: Option<Value>,
    /// A baseline receipt `run-verify` stored, named by attempt id and digest (#2392 D1).
    #[serde(default)]
    pub baseline_ref: Option<BaselineRef>,
    /// Paths the change touches, for the oracle-change check.
    #[serde(default)]
    pub diff_paths: Vec<String>,
}

/// Names a baseline receipt `run-verify` stored under the git common dir: the attempt id it
/// was stored under and the sha256 `run-verify` reported for the file.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct BaselineRef {
    pub attempt_id: String,
    pub digest: String,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "kebab-case")]
pub enum ProducerKind {
    RunnerReceipt,
    SelfReport,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Producer {
    pub kind: ProducerKind,
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Summary {
    Passed,
    Failed,
    NotRun,
    Blocked,
    Unknown,
}

/// Shaped after an in-toto statement (subject: the tree digest; the rest: what ran), unsigned:
/// it claims no tamper resistance. `producer.kind` says how far it can be trusted.
#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Receipt {
    pub producer: Producer,
    pub task_id: String,
    pub attempt_id: String,
    pub spec_ref: String,
    pub spec_digest: String,
    pub plan_digest: String,
    pub base_revision: String,
    pub tree_digest: String,
    pub policy_digest: Option<String>,
    pub toolchain: Toolchain,
    pub method: ResolvedMethod,
    pub phase: Option<String>,
    pub required_evidence: Vec<String>,
    /// Every declared check, run or `not_run`.
    pub checks: Vec<CheckRun>,
    pub oracle_changes: Vec<String>,
    /// For the RED step of `tdd` / `bugfix` only.
    pub red_confirmed: Option<bool>,
    pub result: Summary,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct RunOutcome {
    /// `ready` when the commands ran; `stale` or `blocked` when nothing ran.
    pub decision: Decision,
    pub reasons: Vec<String>,
    pub receipt: Option<Receipt>,
}

pub fn resolve_method(m: &Method) -> Result<ResolvedMethod> {
    let explicit = m.change_strategy.is_some() || m.domain_modeling.is_some();
    let (strategy, domain, legacy) = match (&m.legacy_mode, explicit) {
        (Some(_), true) => {
            return Err(
                "invalid run request: give change_strategy and domain_modeling, or legacy_mode"
                    .into(),
            )
        }
        (Some(l), false) => match l.as_str() {
            "tdd" => ("tdd", "not-required", Some(l.clone())),
            "ddd" => ("preserve", "required", Some(l.clone())),
            _ => return Err("invalid run request: legacy_mode must be tdd or ddd".into()),
        },
        (None, _) => (
            m.change_strategy.as_deref().unwrap_or_default(),
            m.domain_modeling.as_deref().unwrap_or_default(),
            None,
        ),
    };
    if !STRATEGIES.iter().any(|(s, _)| *s == strategy) {
        return Err("invalid run request: unknown change_strategy".into());
    }
    if !matches!(domain, "not-required" | "required") {
        return Err("invalid run request: domain_modeling must be not-required or required".into());
    }
    Ok(ResolvedMethod {
        change_strategy: strategy.into(),
        domain_modeling: domain.into(),
        legacy_mode: legacy,
    })
}

pub fn required_evidence(m: &ResolvedMethod) -> Vec<String> {
    let mut out: Vec<String> = STRATEGIES
        .iter()
        .find(|(s, _)| *s == m.change_strategy)
        .map(|(_, e)| e.iter().map(|x| x.to_string()).collect())
        .unwrap_or_default();
    if m.domain_modeling == "required" {
        out.push(DOMAIN_EVIDENCE.into());
    }
    out
}

/// Test, fixture and golden paths among `paths`, sorted: changes to what judges the code.
pub fn oracle_changes(paths: &[String]) -> Vec<String> {
    let mut out: Vec<String> = paths
        .iter()
        .map(|p| p.trim().trim_start_matches("./").to_string())
        .filter(|p| {
            let parts: Vec<&str> = p.split('/').collect();
            parts.first() == Some(&"tests")
                || parts.iter().any(|s| matches!(*s, "fixtures" | "golden"))
                || p.ends_with("_test.py")
        })
        .collect();
    out.sort();
    out.dedup();
    out
}

/// RED is confirmed for the `red` / `repro-fail` step of `tdd` / `bugfix` when something failed
/// and every failure is an assertion failure.
pub fn red_confirmed(m: &ResolvedMethod, phase: Option<&str>, checks: &[CheckRun]) -> Option<bool> {
    let red_step = matches!(
        (m.change_strategy.as_str(), phase),
        ("tdd", Some("red")) | ("bugfix", Some("repro-fail"))
    );
    if !red_step {
        return None;
    }
    let failed: Vec<&CheckRun> = checks
        .iter()
        .filter(|c| c.result == CheckResult::Failed)
        .collect();
    Some(
        !failed.is_empty()
            && failed
                .iter()
                .all(|c| c.failure_class == Some(FailureClass::TestFailure)),
    )
}

/// Accepts agent-written receipt JSON as a self-report, whatever producer it claims.
pub fn ingest_self_report(value: Value) -> Result<Receipt> {
    let mut r: Receipt =
        serde_json::from_value(value).map_err(|_| "invalid typed JSON input".to_string())?;
    r.producer.kind = ProducerKind::SelfReport;
    Ok(r)
}

/// Only a receipt this runner produced counts as evidence for admission.
pub fn admissible_evidence(r: &Receipt) -> bool {
    r.producer.kind == ProducerKind::RunnerReceipt
}

/// Why a receipt no longer describes the current state; empty when it still does.
pub fn stale_reasons(
    r: &Receipt,
    tree_digest: &str,
    spec_digest: &str,
    policy_digest: Option<&str>,
) -> Vec<&'static str> {
    let mut out = Vec::new();
    if r.tree_digest != tree_digest {
        out.push("tree-changed");
    }
    if r.spec_digest != spec_digest {
        out.push("spec-changed");
    }
    if r.policy_digest.as_deref() != policy_digest {
        out.push("policy-changed");
    }
    out
}

/// Tasks whose receipts are stale and need verifying again, sorted.
pub fn revalidate(
    receipts: &[Receipt],
    tree_digest: &str,
    spec_digest: &str,
    policy_digest: Option<&str>,
) -> Vec<String> {
    let mut out: Vec<String> = receipts
        .iter()
        .filter(|r| !stale_reasons(r, tree_digest, spec_digest, policy_digest).is_empty())
        .map(|r| r.task_id.clone())
        .collect();
    out.sort();
    out.dedup();
    out
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Budget {
    /// Required when retries are wanted; there is no default (P11).
    pub max_attempts: u32,
}

/// REQ-05C-11: with a budget, stop once `max_attempts` attempts have run and the last ones all
/// failed on the same checks. Without a budget there is no retry decision (one result only).
/// The caller holds the attempt history: receipts are not stored anywhere yet (P09).
pub fn budget_exhausted(budget: Option<&Budget>, attempts: &[Receipt]) -> bool {
    let Some(b) = budget else {
        return false;
    };
    let n = b.max_attempts as usize;
    if n == 0 || attempts.len() < n {
        return false;
    }
    let failing = |r: &Receipt| -> Vec<usize> {
        r.checks
            .iter()
            .filter(|c| c.result == CheckResult::Failed)
            .map(|c| c.command_ref)
            .collect()
    };
    let last = &attempts[attempts.len() - n..];
    let first = failing(&last[0]);
    !first.is_empty() && last.iter().all(|r| failing(r) == first)
}

/// Required evidence the attempts have not produced yet: what is left to report at a stop.
///
/// Only a receipt this runner produced counts (a self-report is never evidence). The RED step
/// of `tdd` / `bugfix` (`red`, `repro-fail`) is settled by an intended assertion failure —
/// that step's receipt with `red_confirmed: Some(true)` — and not by a pass; a compile,
/// network or no-selection failure leaves it open (#2392, design boundary table row 2). Every
/// other step is settled by a passed receipt of that step.
pub fn remaining_evidence(required: &[String], attempts: &[Receipt]) -> Vec<String> {
    let settles = |r: &Receipt, step: &str| {
        r.producer.kind == ProducerKind::RunnerReceipt
            && r.phase.as_deref() == Some(step)
            && if matches!(step, "red" | "repro-fail") {
                r.red_confirmed == Some(true)
            } else {
                r.result == Summary::Passed
            }
    };
    required
        .iter()
        .filter(|e| !attempts.iter().any(|r| settles(r, e.as_str())))
        .cloned()
        .collect()
}

fn summarize(checks: &[CheckRun], oracle: &[String]) -> Summary {
    if checks.iter().any(|c| c.result == CheckResult::Failed) {
        Summary::Failed
    } else if checks.iter().any(|c| c.result == CheckResult::NotRun) {
        Summary::NotRun
    } else if !oracle.is_empty() {
        // Passing while the oracle itself changed needs a reader, not a summary.
        Summary::Unknown
    } else {
        Summary::Passed
    }
}

fn halt(decision: Decision, reasons: Vec<String>) -> RunOutcome {
    RunOutcome {
        decision,
        reasons,
        receipt: None,
    }
}

/// Re-reads the pinned contract, then runs the requested commands in `project`. Input errors
/// return `Err`; a mismatch returns `stale` or `blocked` without running anything.
pub fn run(project: &Path, req: &RunRequest, current_revision: &str) -> Result<RunOutcome> {
    if req.intent != "RUN" {
        return Err("invalid run request: intent must be RUN".into());
    }
    if req.verification.timeout_ms == 0 {
        return Err("invalid run request: timeout_ms must be positive".into());
    }
    let method = resolve_method(&req.method)?;
    let evidence = required_evidence(&method);
    // Under `preserve`, a run without a phase is the baseline step (#2392 decision 6774899518
    // row 13, F1b = B), so it can never stand in for a later step.
    let phase: Option<String> = match (method.change_strategy.as_str(), &req.phase) {
        ("preserve", None) => Some("baseline".into()),
        (_, p) => p.clone(),
    };
    if let Some(p) = &req.phase {
        if !evidence.iter().any(|e| e == p) || p == DOMAIN_EVIDENCE {
            return Err("invalid run request: phase is not a step of the change strategy".into());
        }
    }
    let commands = &req.admission.spec.verification_commands;
    if req
        .verification
        .commands_ref
        .iter()
        .any(|i| *i >= commands.len())
    {
        return Err("invalid run request: commands_ref out of range".into());
    }

    // REQ-05C-01: the pinned contract against what 05-B admits now.
    let admitted = admit(&req.admission, current_revision);
    if admitted.decision != Decision::Ready {
        return Ok(halt(
            admitted.decision,
            admitted.reasons.iter().map(|r| r.to_string()).collect(),
        ));
    }
    let Some(now) = admitted.pinned.iter().find(|p| p.task_id == req.task_id) else {
        return Ok(halt(Decision::Blocked, vec!["task-not-admitted".into()]));
    };
    let now = PinnedRef {
        spec_ref: now.spec_ref.clone(),
        plan_digest: now.plan_digest.clone(),
        task_id: now.task_id.clone(),
        expected_revision: now.expected_revision.clone(),
        scope: now.scope.clone(),
        approval_ref: now.approval_ref.clone(),
    };
    if now != req.pinned {
        return Ok(halt(
            Decision::Stale,
            vec!["pinned-contract-mismatch".into()],
        ));
    }

    let snapshot = req
        .admission
        .spec
        .snapshot
        .as_ref()
        .expect("admitted spec is approved");

    // REQ-05C-04: improving under `preserve` needs a trusted baseline first: one this runner
    // stored (#2392 D1). A baseline handed in as JSON is a self-report and never unlocks it.
    if method.change_strategy == "preserve" && phase.as_deref() != Some("baseline") {
        let reason = match (&req.baseline_ref, &req.baseline) {
            (Some(r), _) => stored_baseline_reason(
                project,
                r,
                &req.task_id,
                &snapshot.digest,
                &admitted.plan.plan_digest,
            ),
            (None, None) => "baseline-missing",
            (None, Some(v)) => match ingest_self_report(v.clone()) {
                Ok(r) if admissible_evidence(&r) => "",
                _ => "baseline-untrusted",
            },
        };
        if !reason.is_empty() {
            return Ok(halt(Decision::Blocked, vec![reason.into()]));
        }
    }

    // Every requested command must be runnable before any of them starts.
    let mut wanted: Vec<usize> = req.verification.commands_ref.clone();
    wanted.sort_unstable();
    wanted.dedup();
    if wanted
        .iter()
        .any(|i| verify::parse_argv(&commands[*i]).is_err())
    {
        return Ok(halt(Decision::Blocked, vec!["command-not-runnable".into()]));
    }

    let tree_digest = crate::tree_digest(project, &TREE_EXCLUDES)?;
    let toolchain = verify::toolchain(project);
    let timeout = Duration::from_millis(req.verification.timeout_ms);
    let checks: Vec<CheckRun> = commands
        .iter()
        .enumerate()
        .map(|(i, c)| {
            if wanted.contains(&i) {
                verify::run_check(project, i, c, timeout)
            } else {
                verify::not_run(i, c)
            }
        })
        .collect();
    let oracle = oracle_changes(&req.diff_paths);
    let receipt = Receipt {
        producer: Producer {
            kind: ProducerKind::RunnerReceipt,
        },
        task_id: req.task_id.clone(),
        attempt_id: req.attempt_id.clone(),
        spec_ref: snapshot.reference.clone(),
        spec_digest: snapshot.digest.clone(),
        plan_digest: admitted.plan.plan_digest.clone(),
        base_revision: current_revision.into(),
        tree_digest,
        policy_digest: req
            .admission
            .spec
            .policy_revision
            .as_ref()
            .map(|p| digest(p.as_bytes())),
        toolchain,
        red_confirmed: red_confirmed(&method, phase.as_deref(), &checks),
        result: summarize(&checks, &oracle),
        method,
        phase,
        required_evidence: evidence,
        checks,
        oracle_changes: oracle,
    };
    Ok(RunOutcome {
        decision: Decision::Ready,
        reasons: Vec::new(),
        receipt: Some(receipt),
    })
}

/// What a stored baseline proves, said with every store and in the docs: the file is unsigned,
/// so a writer on this host can replace it together with its reference.
pub const BASELINE_TRUST: &str =
    "local file under the git common dir, checked by sha256; unsigned, so same-host tampering is not prevented";

/// An attempt id that is safe as a file name: 1 to 128 of `A-Z a-z 0-9 . _ -`, not starting
/// with a dot.
fn safe_attempt_id(id: &str) -> bool {
    (1..=128).contains(&id.len())
        && !id.starts_with('.')
        && id
            .bytes()
            .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'.' | b'_' | b'-'))
}

/// `<git common dir>/.moai-cli/receipts/run`, beside the claim journal and outside every
/// tree digest; None outside a git repository.
fn baseline_store_dir(project: &Path) -> Option<std::path::PathBuf> {
    let id = crate::identity::identity(project).ok()?;
    let common = id["clone_anchor"].as_str()?;
    Some(
        crate::clone_root::path(Path::new(common))
            .join("receipts")
            .join("run"),
    )
}

/// Why a referenced stored baseline does not unlock a later `preserve` step; empty when it
/// does. The file must hash to the reference and say what a passed runner baseline of this
/// task, spec and plan says.
fn stored_baseline_reason(
    project: &Path,
    r: &BaselineRef,
    task_id: &str,
    spec_digest: &str,
    plan_digest: &str,
) -> &'static str {
    if !safe_attempt_id(&r.attempt_id) {
        return "baseline-untrusted";
    }
    let Some(dir) = baseline_store_dir(project) else {
        return "baseline-missing";
    };
    let bytes = match std::fs::read(dir.join(format!("{}.json", r.attempt_id))) {
        Ok(b) => b,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return "baseline-missing",
        Err(_) => return "baseline-untrusted",
    };
    if digest(&bytes) != r.digest {
        return "baseline-digest-mismatch";
    }
    let Ok(b) = serde_json::from_slice::<Receipt>(&bytes) else {
        return "baseline-untrusted";
    };
    let binds = b.producer.kind == ProducerKind::RunnerReceipt
        && b.method.change_strategy == "preserve"
        && b.phase.as_deref() == Some("baseline")
        && b.result == Summary::Passed
        && b.attempt_id == r.attempt_id
        && b.task_id == task_id
        && b.spec_digest == spec_digest
        && b.plan_digest == plan_digest;
    if binds {
        ""
    } else {
        "baseline-untrusted"
    }
}

/// Stores a baseline receipt under its attempt id, never replacing one: written to a temporary
/// file in the store and published by a hard link, which fails when the name exists.
fn store_baseline(project: &Path, receipt: &Receipt) -> Value {
    let refuse = |reason: &str| json!({"stored": false, "reason": reason});
    if !safe_attempt_id(&receipt.attempt_id) {
        return refuse("attempt-id-not-a-safe-file-name");
    }
    let Some(dir) = baseline_store_dir(project) else {
        return refuse("not-a-git-repository");
    };
    let Ok(bytes) = serde_json::to_vec_pretty(receipt) else {
        return refuse("encode-failed");
    };
    if std::fs::create_dir_all(&dir).is_err() {
        return refuse("store-unwritable");
    }
    let target = dir.join(format!("{}.json", receipt.attempt_id));
    let tmp = dir.join(format!(
        ".{}.{}.tmp",
        receipt.attempt_id,
        std::process::id()
    ));
    if std::fs::write(&tmp, &bytes).is_err() {
        let _ = std::fs::remove_file(&tmp);
        return refuse("store-unwritable");
    }
    let linked = std::fs::hard_link(&tmp, &target);
    let _ = std::fs::remove_file(&tmp);
    match linked {
        Ok(()) => json!({
            "stored": true,
            "attempt_id": receipt.attempt_id,
            "path": target.to_string_lossy(),
            "digest": digest(&bytes),
            "trust": BASELINE_TRUST,
        }),
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => refuse("attempt-id-exists"),
        Err(_) => refuse("store-unwritable"),
    }
}

/// `run-verify`'s answer: the run, plus the stored baseline when this run was one.
#[derive(Clone, Debug, PartialEq, Eq, Serialize)]
pub struct VerifyOutcome {
    #[serde(flatten)]
    pub outcome: RunOutcome,
    /// Set for a `preserve` baseline run that produced a receipt; null otherwise.
    pub baseline_store: Option<Value>,
}

/// `run-verify`: [`run`], then the baseline receipt is stored when the run was the baseline
/// step of `preserve` (#2392 decision 6774899518 row 13 = A). The auditor's run calls [`run`]
/// and stores nothing.
pub fn run_verify(
    project: &Path,
    req: &RunRequest,
    current_revision: &str,
) -> Result<VerifyOutcome> {
    let outcome = run(project, req, current_revision)?;
    let baseline_store = outcome
        .receipt
        .as_ref()
        .filter(|r| {
            r.method.change_strategy == "preserve" && r.phase.as_deref() == Some("baseline")
        })
        .map(|r| store_baseline(project, r));
    Ok(VerifyOutcome {
        outcome,
        baseline_store,
    })
}

/// Cost-request input (#2191, REQ-D5-01): an array of run/verify receipts to
/// aggregate. Receipts arrive as JSON in the exact `Receipt` schema — self-reports
/// are aggregated as what they claim, and the report's producer rows keep that
/// distinction visible.
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct CostRequest {
    pub protocol_version: u32,
    pub receipts: Vec<Receipt>,
}

/// `cost-report` (#2191): aggregates cost items across run/verify receipts —
/// check outcomes, per-command and total durations, wall spans, external backend
/// calls and tokens — separately per producer kind (REQ-D5-01: estimates and
/// measurements are distinguished). Only runner receipts carry numbers the runner
/// measured, so their group is `measured`; a self-report group carries what the
/// reporter claims and is tagged `self_reported`. External backend calls are
/// structurally zero only for runner receipts (the runner spawns local commands and
/// has no network path); for self-reports nothing establishes the count, so it is
/// `absent`. Token accounting has no source in any receipt and is `absent` rather
/// than estimated; the report contains no estimate rows.
pub fn cost_report(req: CostRequest) -> Result<Value> {
    if req.protocol_version != 1 {
        return Err("unsupported protocol_version; supported: 1".into());
    }
    let runner: Vec<&Receipt> = req
        .receipts
        .iter()
        .filter(|r| r.producer.kind == ProducerKind::RunnerReceipt)
        .collect();
    let self_report: Vec<&Receipt> = req
        .receipts
        .iter()
        .filter(|r| r.producer.kind == ProducerKind::SelfReport)
        .collect();
    let per_receipt: Vec<Value> = req.receipts.iter().map(receipt_row).collect();
    Ok(json!({
        "protocol_version": 1,
        "receipts": req.receipts.len(),
        "by_producer": {
            "runner-receipt": producer_group(&runner, ProducerKind::RunnerReceipt),
            "self-report": producer_group(&self_report, ProducerKind::SelfReport),
        },
        "per_receipt": per_receipt,
        // Every receipt here arrived as caller JSON, so its `producer.kind` is what the caller
        // declared; the groups keep #2191's aggregation (#2392 decision 6774899518 row 16 = B).
        "producer_declared_by": PRODUCER_DECLARED_BY,
        "tokens": {
            "basis": "absent",
            "note": "receipts carry no token source; nothing is estimated"
        },
        "estimates": [],
        "network": "none",
        "writes": "none"
    }))
}

/// Who stated a cost-report receipt's `producer.kind`: the caller, never re-verified here.
const PRODUCER_DECLARED_BY: &str = "caller";

/// Evidence basis of a receipt's own numbers: measured by the runner, or claimed
/// by a self-report.
fn producer_basis(kind: ProducerKind) -> &'static str {
    match kind {
        ProducerKind::RunnerReceipt => "measured",
        ProducerKind::SelfReport => "self_reported",
    }
}

/// Sum of a receipt's check durations, with the count of checks that carry none.
fn receipt_duration(r: &Receipt) -> (u128, usize) {
    r.checks
        .iter()
        .fold((0, 0), |(sum, missing), c| match c.duration_ms {
            Some(ms) => (sum + ms as u128, missing),
            None => (sum, missing + 1),
        })
}

/// Earliest check start to latest check finish, when both exist and are ordered.
fn receipt_wall_span(r: &Receipt) -> Option<u64> {
    let start = r.checks.iter().filter_map(|c| c.started_unix_ms).min();
    let finish = r.checks.iter().filter_map(|c| c.finished_unix_ms).max();
    match (start, finish) {
        (Some(s), Some(f)) if f >= s => Some(f - s),
        _ => None,
    }
}

fn receipt_row(r: &Receipt) -> Value {
    let (duration, missing) = receipt_duration(r);
    json!({
        "task_id": r.task_id,
        "attempt_id": r.attempt_id,
        "producer_kind": r.producer.kind,
        "producer_declared_by": PRODUCER_DECLARED_BY,
        "basis": producer_basis(r.producer.kind),
        "checks_total": r.checks.len(),
        "duration_ms": if missing == 0 { json!(duration as u64) } else { json!(null) },
        "duration_missing_samples": missing,
        "wall_span_ms": receipt_wall_span(r),
        "result": r.result,
    })
}

/// Aggregate of one producer kind's receipts; every row carries that kind's basis.
/// A kind with no receipts is `null`: an empty group asserts no basis at all.
fn producer_group(receipts: &[&Receipt], kind: ProducerKind) -> Value {
    if receipts.is_empty() {
        return Value::Null;
    }
    let basis = producer_basis(kind);
    let (mut total, mut passed, mut failed, mut not_run) = (0usize, 0usize, 0usize, 0usize);
    let (mut duration_total, mut duration_missing) = (0u128, 0usize);
    let mut wall_spans = Vec::new();
    for r in receipts {
        for c in &r.checks {
            total += 1;
            match c.result {
                CheckResult::Passed => passed += 1,
                CheckResult::Failed => failed += 1,
                CheckResult::NotRun => not_run += 1,
            }
        }
        let (d, m) = receipt_duration(r);
        duration_total += d;
        duration_missing += m;
        wall_spans.extend(receipt_wall_span(r));
    }
    wall_spans.sort_unstable();
    let p50 = wall_spans.get((wall_spans.len().max(1) - 1) / 2).copied();
    let p95 = wall_spans
        .get(((wall_spans.len() as f64 * 0.95).ceil() as usize).max(1) - 1)
        .copied();
    let external = match kind {
        ProducerKind::RunnerReceipt => json!({"count": 0, "basis": "structural"}),
        ProducerKind::SelfReport => json!({"count": null, "basis": "absent"}),
    };
    json!({
        "receipts": receipts.len(),
        "checks": {
            "total": total,
            "passed": passed,
            "failed": failed,
            "not_run": not_run,
            "basis": basis
        },
        "duration_ms": {
            "total": duration_total as u64,
            "samples_missing": duration_missing,
            "basis": basis
        },
        "wall_span_ms": {
            "p50": p50,
            "p95": p95,
            "samples": wall_spans.len(),
            "basis": basis
        },
        "external_backend_calls": external,
    })
}
