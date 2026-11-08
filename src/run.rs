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
use serde_json::Value;
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
    /// Paths the change touches, for the oracle-change check.
    #[serde(default)]
    pub diff_paths: Vec<String>,
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
pub fn remaining_evidence(required: &[String], attempts: &[Receipt]) -> Vec<String> {
    required
        .iter()
        .filter(|e| {
            !attempts
                .iter()
                .any(|r| r.phase.as_deref() == Some(e.as_str()) && r.result == Summary::Passed)
        })
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

    // REQ-05C-04: improving under `preserve` needs a trusted baseline first.
    if method.change_strategy == "preserve"
        && !matches!(req.phase.as_deref(), None | Some("baseline"))
    {
        let reason = match &req.baseline {
            None => "baseline-missing",
            Some(v) => match ingest_self_report(v.clone()) {
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
    let snapshot = req
        .admission
        .spec
        .snapshot
        .as_ref()
        .expect("admitted spec is approved");
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
        red_confirmed: red_confirmed(&method, req.phase.as_deref(), &checks),
        result: summarize(&checks, &oracle),
        method,
        phase: req.phase.clone(),
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
